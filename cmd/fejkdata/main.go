// Command fejkdata prints fake values from the shipped data and any directories
// layered over it.
//
//	fejkdata sv_SE.person                        # a full person
//	fejkdata sv_SE.person.last                   # just the surname
//	fejkdata --data-path ./mydata sv_SE.person   # layer custom data; the last dir wins
//	fejkdata --seed 42 sv_SE.address
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"gitea.larvit.se/larvit/fejkdata"
)

const usage = `Usage: fejkdata [flags] <path|template>

  <path>                 a category, or a dotted path into one (person, person.last);
                         a table's row by key or name: 'misc.country[SE]', 'misc.country[Sweden].tld'
  <template>             a format string or JSON value to render inline, e.g.
                         'name: {/sv_SE.person.last}' or '{"format":"{x}","x":["bosse","lina"]}'

A layout inside a template is single-quoted, so quote the whole argument with " to
keep it: "{date(1990-01-01,2010-12-31,'2006-01-02')}".

An argument containing a { token, or a JSON object, array or string, is a
template; any other argument is a path (a path never contains a brace or a quote,
and a bracket only as a [selector] after a table's name). Templates reach the
data by reference from the root — {/sv_SE.person.last} — whether the data is
shipped or layered with --data-path. An argument carrying a closing brace or a
quote but no valid JSON names neither.

With --format json, ndjson, csv or sql the argument must name a record — a
template whose fields are its columns — and the rows are written as one JSON
array, one JSON object per line, one CSV row (after a header), or one INSERT.

  -d, --data-path D      a data directory to layer over the shipped data (repeatable; last wins on a clash)
      --format F         output form: text (default), json, ndjson, csv or sql
  -h, --help             print this help, then exit
      --list             list the paths the data offers, then exit
      --no-shipped-data  load only the --data-path directories
  -n, --repeat N         render the value N times, 1..1048576 (default 1)
  -s, --seed N           seed for reproducible output
      --separator S      string between repeated values (default newline)
      --table T          the INSERT target for --format sql (default: the path's last segment, or records for an inline template)
      --version          print the version, then exit

Flags may come before or after <path|template>; -- ends the flags. A short flag's
value attaches or follows (-n3, -n 3); short flags bundle (-hn 3).
`

type invocation struct {
	dirs         []string
	format       string
	formatSet    bool
	help         bool
	list         bool
	noShipped    bool
	paths        []string
	repeat       int
	repeatSet    bool
	seed         uint64
	seeded       bool
	separator    string
	separatorSet bool
	table        string
	tableSet     bool
	version      bool
}

type flagDef struct {
	long  string
	short string
	value bool
	set   func(*invocation, string) error
}

var flagDefs = []flagDef{
	{"data-path", "d", true, func(in *invocation, v string) error { in.dirs = append(in.dirs, v); return nil }},
	{"format", "", true, func(in *invocation, v string) error { in.format, in.formatSet = v, true; return nil }},
	{"help", "h", false, func(in *invocation, _ string) error { in.help = true; return nil }},
	{"list", "", false, func(in *invocation, _ string) error { in.list = true; return nil }},
	{"no-shipped-data", "", false, func(in *invocation, _ string) error { in.noShipped = true; return nil }},
	{"repeat", "n", true, func(in *invocation, v string) error {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > fejkdata.MaxRepeat {
			return fmt.Errorf("--repeat needs an integer in 1..%d, got %q", fejkdata.MaxRepeat, v)
		}
		in.repeat, in.repeatSet = n, true
		return nil
	}},
	{"seed", "s", true, func(in *invocation, v string) error {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("--seed needs an unsigned integer, got %q", v)
		}
		in.seed, in.seeded = n, true
		return nil
	}},
	{"separator", "", true, func(in *invocation, v string) error { in.separator, in.separatorSet = v, true; return nil }},
	{"table", "", true, func(in *invocation, v string) error { in.table, in.tableSet = v, true; return nil }},
	{"version", "", false, func(in *invocation, _ string) error { in.version = true; return nil }},
}

func flagByLong(name string) *flagDef {
	for i := range flagDefs {
		if flagDefs[i].long == name {
			return &flagDefs[i]
		}
	}
	return nil
}

func flagByShort(name string) *flagDef {
	for i := range flagDefs {
		if flagDefs[i].short == name {
			return &flagDefs[i]
		}
	}
	return nil
}

// flagArg is one flag as written: its definition and the value attached to it,
// if any.
type flagArg struct {
	def    *flagDef
	value  string
	inline bool
}

// splitFlags reads one argv element as flags: --name or --name=value, or -abc
// where each letter is a short flag and the first that takes a value takes the
// rest of the element.
func splitFlags(arg string) ([]flagArg, error) {
	if strings.HasPrefix(arg, "--") {
		name, value, inline := strings.Cut(arg[2:], "=")
		def := flagByLong(name)
		if def == nil {
			return nil, fmt.Errorf("unknown flag --%s", name)
		}
		return []flagArg{{def, value, inline}}, nil
	}
	if name, _, _ := strings.Cut(arg[1:], "="); len(name) > 1 && flagByLong(name) != nil {
		return nil, fmt.Errorf("unknown flag %s; use --%s", arg, name)
	}
	var flags []flagArg
	letters := arg[1:]
	for i, r := range letters {
		letter := string(r)
		def := flagByShort(letter)
		if def == nil {
			return nil, fmt.Errorf("unknown flag -%s", letter)
		}
		if !def.value {
			flags = append(flags, flagArg{def: def})
			continue
		}
		rest := letters[i+len(letter):]
		if strings.HasPrefix(rest, "=") {
			return nil, fmt.Errorf("-%s takes its value attached (-%s%s) or next (-%s %s); = belongs to --%s=%s",
				letter, letter, rest[1:], letter, rest[1:], def.long, rest[1:])
		}
		return append(flags, flagArg{def, rest, rest != ""}), nil
	}
	return flags, nil
}

func parseArgs(argv []string) (invocation, error) {
	in := invocation{repeat: 1, separator: "\n", format: "text"}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			in.paths = append(in.paths, argv[i+1:]...)
			return in, nil
		}
		if len(arg) < 2 || arg[0] != '-' {
			in.paths = append(in.paths, arg)
			continue
		}
		flags, err := splitFlags(arg)
		if err != nil {
			return in, err
		}
		for _, fl := range flags {
			if !fl.def.value {
				if fl.inline {
					return in, fmt.Errorf("--%s takes no value", fl.def.long)
				}
				_ = fl.def.set(&in, "")
				continue
			}
			value := fl.value
			if !fl.inline {
				if i++; i >= len(argv) {
					return in, fmt.Errorf("--%s needs a value", fl.def.long)
				}
				value = argv[i]
			}
			if err := fl.def.set(&in, value); err != nil {
				return in, err
			}
		}
	}
	return in, nil
}

// recordFormat is one way to write a record out: the line each record renders,
// the header that precedes the first one, and the open/close frame plus the
// between-record separator a document form needs.
type recordFormat struct {
	header func(*fejkdata.Record) string
	line   func(r *fejkdata.Record, table string) string
	open   string
	close  string
	sep    string
}

// recordFormats is every --format that writes records. json frames the records
// as one array document; ndjson is the same column, one object per line.
var recordFormats = map[string]recordFormat{
	"csv":    {header: (*fejkdata.Record).CSVHeader, line: func(r *fejkdata.Record, _ string) string { return r.CSVLine() }, sep: "\n"},
	"json":   {line: jsonLine, open: "[", close: "]", sep: ",\n"},
	"ndjson": {line: jsonLine, sep: "\n"},
	"sql":    {line: func(r *fejkdata.Record, table string) string { return r.SQLInsert(table) }, sep: "\n"},
}

func jsonLine(r *fejkdata.Record, _ string) string { return r.JSON() }

// writesRecords reports whether the format writes records rather than plain text.
func (in invocation) writesRecords() bool {
	return in.format != "text"
}

// formatNames lists the --format values, text included, for the misuse error.
func formatNames() string {
	names := []string{"text"}
	for name := range recordFormats {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// checkFlags rejects a flag value or combination that cannot run.
func (in invocation) checkFlags() error {
	if _, ok := recordFormats[in.format]; !ok && in.format != "text" {
		return fmt.Errorf("--format takes %s, got %q", formatNames(), in.format)
	}
	if in.tableSet && in.table == "" {
		return errors.New("--table names the INSERT target, so it cannot be empty")
	}
	if in.tableSet && in.format != "sql" {
		return errors.New("--table names the INSERT target, so it needs --format sql")
	}
	if in.writesRecords() && in.separatorSet {
		return errors.New("--separator joins text values, so it has no effect with --format " + in.format)
	}
	return nil
}

// check rejects a flag combination or an argument that cannot run, and reports
// what the argument names, so its shape is settled before any data is read.
func (in invocation) check() (argKind, error) {
	if err := in.checkFlags(); err != nil {
		return argPath, err
	}
	if in.list && len(in.paths) > 0 {
		return argPath, errors.New("--list takes no path")
	}
	if in.list && (in.repeatSet || in.separatorSet || in.formatSet || in.tableSet) {
		return argPath, errors.New("--list takes no --repeat, --separator, --format or --table")
	}
	if in.list {
		return argPath, nil
	}
	if len(in.paths) != 1 {
		return argPath, fmt.Errorf("expected one path or template, got %d", len(in.paths))
	}
	return classify(in.paths[0])
}

func (in invocation) options() []fejkdata.Option {
	var opts []fejkdata.Option
	if in.noShipped {
		opts = append(opts, fejkdata.WithoutShippedData())
	}
	for _, dir := range in.dirs {
		opts = append(opts, fejkdata.WithDataPath(dir))
	}
	if in.seeded {
		opts = append(opts, fejkdata.WithSeed(in.seed))
	}
	return opts
}

// write streams the argument's renders to w, repeat of them joined by the
// separator and ended by a newline. A value that renders once renders every time,
// so a render failure comes before anything is written; a write failure surfaces
// from Flush, bufio keeping the first one.
func (in invocation) write(f *fejkdata.Generator, kind argKind, w io.Writer) error {
	if in.writesRecords() {
		return in.writeRecords(f, kind, w)
	}
	draw, err := in.textDraw(f, kind, in.paths[0])
	if err != nil {
		return err
	}
	out := bufio.NewWriter(w)
	for i := 0; i < in.repeat; i++ {
		v, err := draw()
		if err != nil {
			return err
		}
		if i > 0 {
			out.WriteString(in.separator)
		}
		out.WriteString(v)
	}
	out.WriteString("\n")
	return out.Flush()
}

// textDraw builds what one text render yields: the value, from a path or an
// inline template.
func (in invocation) textDraw(f *fejkdata.Generator, kind argKind, arg string) (func() (string, error), error) {
	if kind != argTemplate {
		return func() (string, error) { return f.Fake(arg) }, nil
	}
	t, err := f.NewTemplate(arg)
	if err != nil {
		return nil, templateError{err}
	}
	return func() (string, error) { return t.Fake(), nil }, nil
}

// recordStream builds the record drawer for the argument, plus the INSERT table
// a sql format names.
func (in invocation) recordStream(f *fejkdata.Generator, kind argKind, arg string) (func() (*fejkdata.Record, error), string, error) {
	record := func() (*fejkdata.Record, error) { return f.FakeRecord(arg) }
	if kind == argTemplate {
		t, err := f.NewRecordTemplate(arg)
		if err != nil {
			return nil, "", templateError{err}
		}
		record = func() (*fejkdata.Record, error) { return t.Fake(), nil }
	}
	table := in.table
	if table == "" {
		table = defaultTable(arg, kind)
	}
	return record, table, nil
}

// writeRecords streams a record per line in the chosen format, framing a document
// form with its open/close brackets and a header preceding the first record.
func (in invocation) writeRecords(f *fejkdata.Generator, kind argKind, w io.Writer) error {
	record, table, err := in.recordStream(f, kind, in.paths[0])
	if err != nil {
		return err
	}
	format := recordFormats[in.format]
	out := bufio.NewWriter(w)
	if format.open != "" {
		out.WriteString(format.open)
		out.WriteByte('\n')
	}
	for i := 0; i < in.repeat; i++ {
		r, err := record()
		if err != nil {
			return err
		}
		if i == 0 && format.header != nil {
			out.WriteString(format.header(r))
			out.WriteByte('\n')
		}
		if i > 0 {
			out.WriteString(format.sep)
		}
		out.WriteString(format.line(r, table))
	}
	if format.close != "" {
		out.WriteByte('\n')
		out.WriteString(format.close)
	}
	out.WriteByte('\n')
	return out.Flush()
}

// defaultTable names the INSERT target when --table is absent: the path's last
// segment, or "records" for an inline template that sits in no folder.
func defaultTable(arg string, kind argKind) string {
	if kind == argTemplate {
		return "records"
	}
	var names strings.Builder // the path with its [selectors] cut out
	for depth, i := 0, 0; i < len(arg); i++ {
		switch {
		case arg[i] == '[':
			depth++
		case arg[i] == ']':
			depth--
		case depth == 0:
			names.WriteByte(arg[i])
		}
	}
	segments := strings.Split(names.String(), ".")
	return segments[len(segments)-1]
}

// templateError marks a render failure that is the argument's own fault — an
// inline template that does not compile. run reports it as misuse (exit 2, with a
// pointer to --help), unlike an unknown path, which is a runtime error (exit 1).
type templateError struct{ error }

func (e templateError) Unwrap() error { return e.error }

type argKind int

const (
	argPath argKind = iota
	argTemplate
)

// classify reads what a positional argument names by its shape (see fejkdata.IsTemplate).
func classify(arg string) (argKind, error) {
	inline, err := fejkdata.IsTemplate(arg)
	if inline {
		return argTemplate, nil
	}
	return argPath, err
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

// run returns the exit code: 0 ok, 1 runtime error, 2 misuse.
func run(args []string, stdout, stderr io.Writer) int {
	in, err := parseArgs(args)
	if err != nil {
		return misuse(stderr, err)
	}
	if in.help {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if in.version {
		fmt.Fprintln(stdout, "fejkdata "+buildVersion())
		return 0
	}
	kind, err := in.check()
	if err != nil {
		return misuse(stderr, err)
	}
	f, err := fejkdata.New(in.options()...)
	if errors.Is(err, fejkdata.ErrNoData) {
		return misuse(stderr, errors.New("--no-shipped-data needs at least one --data-path"))
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if in.list {
		for _, p := range f.List() {
			fmt.Fprintln(stdout, p)
		}
		return 0
	}
	if err := in.write(f, kind, stdout); err != nil {
		var te templateError
		if errors.As(err, &te) {
			return misuse(stderr, te.error)
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func misuse(stderr io.Writer, err error) int {
	// A library error already names the program, so the prefix is not doubled.
	fmt.Fprintf(stderr, "fejkdata: %s\ntry 'fejkdata --help'\n", strings.TrimPrefix(err.Error(), "fejkdata: "))
	return 2
}

// buildVersion is the module version go install stamps into the binary, or "devel".
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "devel"
}
