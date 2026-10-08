// Command fejkdata renders the template on its stdin with fake values from the shipped
// data and any directories layered over it.
//
//	echo '{/sv_SE.person}' | fejkdata                        # a full person
//	echo 'name: {/sv_SE.person.last}' | fejkdata             # text around a surname
//	echo '{/sv_SE.person}' | fejkdata --data-path ./mydata   # layer custom data; the last dir wins
//	echo '{/sv_SE.address}' | fejkdata --seed 42
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata"
)

const usage = `Usage: <template> | fejkdata [flags]
       fejkdata --list | --help | --version

fejkdata reads one template from stdin and renders it:

  echo 'name: {/sv_SE.person.last}' | fejkdata
  echo '{"format":"{x}","x":["bosse","lina"]}' | fejkdata
  fejkdata < template.txt

fejkdata <<'EOF'
born {date(1990-01-01,2010-12-31,'2006-01-02')}
EOF

A template is a format string, or a JSON object, array or string. Its {…} tokens
reach the data by reference from the root, {/sv_SE.person.last}, whether the data
is shipped or layered with --data-path. A reference names a category or a dotted
path into one ({/sv_SE.person.last}), a table's row by key or name
({/misc.territory[SE]}), and .. steps up from a row to the row it links to
({/geo.SE.locality..municipality.name}).

Under --format text, the default, what you send is what renders, and fejkdata adds
nothing: a format string keeps every byte, so echo's newline ends each render, and
printf '%s' sends none. -n joins renders with --separator, empty by default. A JSON
template is any JSON value, and the whitespace around it is dropped; 42 or true alone
renders its text, and null nothing. An empty or blank stdin renders as it is. A
quoted heredoc, <<'EOF', passes $, a backslash or a quote as written.

With --format json, ndjson, csv or sql the template must be a record — a JSON
template whose fields are its columns, or a lone reference such as {/sv_SE.person}:
one reference and nothing else, with one newline after it allowed — and the rows are
written as one JSON array, one JSON object per line, one CSV row (after a header), or
one INSERT.

  -d, --data-path D      a data directory to layer over the shipped data (repeatable; last wins on a clash)
      --format F         output form: text (default), json, ndjson, csv or sql
  -h, --help             print this help, then exit
      --list             list the paths the data offers, each written after {/ in a template, then exit;
                         --repeat, --separator, --format and --table do nothing here
      --no-shipped-data  load only the --data-path directories
  -n, --repeat N         render the template N times, 0..1048576 (default 1)
  -s, --seed N           same seed, version and data: identical output
      --separator S      string between repeated renders (default empty); a record --format ignores it
      --table T          the INSERT target for --format sql, ignored under any other
                         (default: a lone reference's last segment, else records)
      --version          print the version, then exit

A short flag's value attaches or follows (-n3, -n 3); short flags bundle (-hn 3);
-- ends the flags.
`

type invocation struct {
	args      []string
	dirs      []string
	format    string
	help      bool
	list      bool
	noShipped bool
	repeat    int
	seed      uint64
	seeded    bool
	separator string
	table     string
	tableSet  bool
	version   bool
}

type flagDef struct {
	long  string
	short string
	value bool
	set   func(*invocation, string) error
}

var flagDefs = []flagDef{
	{"data-path", "d", true, func(in *invocation, v string) error { in.dirs = append(in.dirs, v); return nil }},
	{"format", "", true, func(in *invocation, v string) error { in.format = v; return nil }},
	{"help", "h", false, func(in *invocation, _ string) error { in.help = true; return nil }},
	{"list", "", false, func(in *invocation, _ string) error { in.list = true; return nil }},
	{"no-shipped-data", "", false, func(in *invocation, _ string) error { in.noShipped = true; return nil }},
	{"repeat", "n", true, func(in *invocation, v string) error {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > fejkdata.MaxRepeat {
			return fmt.Errorf("--repeat needs an integer in 0..%d, got %q", fejkdata.MaxRepeat, v)
		}
		in.repeat = n
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
	{"separator", "", true, func(in *invocation, v string) error { in.separator = v; return nil }},
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
	in := invocation{repeat: 1, format: "text"}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			in.args = append(in.args, argv[i+1:]...)
			return in, nil
		}
		if len(arg) < 2 || arg[0] != '-' {
			in.args = append(in.args, arg)
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
// as one array document; ndjson writes one object per line.
var recordFormats = map[string]recordFormat{
	"csv":    {header: (*fejkdata.Record).CSVHeader, line: func(r *fejkdata.Record, _ string) string { return r.CSVLine() }, sep: "\n"},
	"json":   {line: jsonLine, open: "[", close: "]", sep: ",\n"},
	"ndjson": {line: jsonLine, sep: "\n"},
	"sql":    {line: func(r *fejkdata.Record, table string) string { return r.SQLInsert(table) }, sep: "\n"},
}

func jsonLine(r *fejkdata.Record, _ string) string { return r.JSON() }

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
	if in.tableSet && in.table == "" && in.format == "sql" {
		return errors.New("--table names the INSERT target, so it cannot be empty")
	}
	return nil
}

const pipeHint = "echo '{/sv_SE.person}' | fejkdata"

// operandError refuses an argument that is no flag, naming the spelling that reads what
// it most likely means: dropping it for -, a file for a file's name, else a template
// piped in.
func operandError(arg string) error {
	unexpected := "unexpected argument '" + arg + "'"
	if arg == "-" {
		return fmt.Errorf("%s; stdin is read already, so drop it", unexpected)
	}
	if fi, err := os.Stat(arg); err == nil && fi.Mode().IsRegular() {
		return fmt.Errorf("%s; read a file from stdin: fejkdata < %s", unexpected, shellQuoted(arg))
	}
	ref := "{/" + strings.TrimPrefix(arg, "/") + "}"
	_, argLone := loneReference(arg)
	inline, err := fejkdata.IsTemplate(arg)
	_, refLone := loneReference(ref)
	example := pipeHint
	switch {
	case argLone || inline:
		example = piped(arg)
	case err == nil && refLone:
		example = piped(ref)
	}
	return fmt.Errorf("%s; the template is read from stdin: %s", unexpected, example)
}

// check rejects a flag combination or an operand that cannot run, before any data is
// read.
func (in invocation) check() error {
	if err := in.checkFlags(); err != nil {
		return err
	}
	if len(in.args) > 0 {
		return operandError(in.args[0])
	}
	return nil
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

// write streams the input's renders to w, repeat of them joined by the separator; a write
// failure surfaces from Flush, bufio keeping the first one.
func (in invocation) write(f *fejkdata.Generator, src input, w io.Writer) error {
	if in.writesRecords() {
		return in.writeRecords(f, src, w)
	}
	t, err := f.NewTemplate(src.template)
	if err != nil {
		return inputError{err}
	}
	out := bufio.NewWriter(w)
	for i := 0; i < in.repeat; i++ {
		if i > 0 {
			out.WriteString(in.separator)
		}
		out.WriteString(t.Fake())
	}
	return out.Flush()
}

// recordStream builds the record drawer for the input, plus the INSERT table
// a sql format names.
func (in invocation) recordStream(f *fejkdata.Generator, src input) (func() (*fejkdata.Record, error), string, error) {
	record := func() (*fejkdata.Record, error) { return f.FakeRecord(src.record) }
	if src.record == "" {
		t, err := f.NewRecordTemplate(src.template)
		if err != nil {
			return nil, "", inputError{spacedReference(src.template, err)}
		}
		record = func() (*fejkdata.Record, error) { return t.Fake(), nil }
	}
	table := in.table
	if table == "" {
		table = defaultTable(src.record)
	}
	return record, table, nil
}

// writeRecords streams a record per line in the chosen format, framing a document
// form with its open/close brackets and a header preceding the first record.
func (in invocation) writeRecords(f *fejkdata.Generator, src input, w io.Writer) error {
	record, table, err := in.recordStream(f, src)
	if err != nil {
		return err
	}
	format := recordFormats[in.format]
	if in.repeat == 0 {
		return writeNoRecord(record, format, w)
	}
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

// writeNoRecord writes zero records: an empty document where the format frames one, else nothing. It
// draws one record and drops it, so a lone reference to nothing fails as it would with records.
func writeNoRecord(record func() (*fejkdata.Record, error), format recordFormat, w io.Writer) error {
	if _, err := record(); err != nil {
		return err
	}
	if format.open == "" {
		return nil
	}
	_, err := io.WriteString(w, format.open+format.close+"\n")
	return err
}

// defaultTable names the INSERT target when --table is absent: the path's last
// segment, or "records" when stdin is not a lone reference.
func defaultTable(path string) string {
	if path == "" {
		return "records"
	}
	segments := strings.Split(withoutSelectors(path), ".")
	return segments[len(segments)-1]
}

// withoutSelectors is path with its [selectors] cut out.
func withoutSelectors(path string) string {
	var names strings.Builder
	for depth, i := 0, 0; i < len(path); i++ {
		switch {
		case path[i] == '[':
			depth++
		case path[i] == ']':
			depth--
		case depth == 0:
			names.WriteByte(path[i])
		}
	}
	return names.String()
}

// inputError marks a failure that is the input's own fault — nothing piped in, a template
// that does not compile, or a stdin that cannot be read. run reports it as misuse (exit 2, with a
// pointer to --help), unlike a lone reference to nothing under --format, which is a runtime
// error (exit 1).
type inputError struct{ error }

func (e inputError) Unwrap() error { return e.error }

// jsonSpace is the whitespace JSON allows around a value.
const jsonSpace = " \t\r\n"

// input is what stdin holds: the template, and the path a lone reference names as a
// record.
type input struct {
	template string
	record   string
}

// parseInput reads stdin as the template, and as a record where it is a lone reference: JSON
// with the whitespace around it dropped, else a format string with one newline ending it
// allowed.
func parseInput(raw string) input {
	trimmed := strings.Trim(raw, jsonSpace)
	text := raw
	if json.Valid([]byte(trimmed)) {
		text = trimmed
	}
	text, found := strings.CutSuffix(text, "\n")
	if found {
		text = strings.TrimSuffix(text, "\r")
	}
	record, _ := loneReference(text)
	return input{template: raw, record: record}
}

// loneReference returns the path that text reads when it is one reference from the root
// and nothing else, bare or as a JSON string.
func loneReference(text string) (string, bool) {
	var unquoted string
	if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) && json.Unmarshal([]byte(text), &unquoted) == nil {
		text = unquoted
	}
	path, opens := strings.CutPrefix(text, "{/")
	path, closes := strings.CutSuffix(path, "}")
	if !opens || !closes || !isPath(path) {
		return "", false
	}
	return path, true
}

// isPath reports whether a reference's body after its / is a path alone: no second /, no
// folder sigil, no empty segment, and none of the |, ( or " as " a token reads outside a selector.
// It copies the grammar, since the CLI reads only the library's public API.
func isPath(body string) bool {
	if body == "" || strings.ContainsAny(body[:1], "/.") {
		return false
	}
	names := withoutSelectors(body)
	if strings.ContainsAny(names, "|(") || strings.Contains(names, " as ") || hasEmptySegment(names) {
		return false
	}
	inline, err := fejkdata.IsTemplate(body)
	return !inline && err == nil
}

// hasEmptySegment reports whether a dotted path holds a segment with no name: two dots are a
// step up, and only a step up may end the path.
func hasEmptySegment(path string) bool {
	segments := strings.Split(path, ".")
	stepUp := false
	for i, seg := range segments[1:] {
		last := i == len(segments)-2
		switch {
		case seg != "":
			stepUp = false
		case stepUp != last:
			return true
		default:
			stepUp = !stepUp
		}
	}
	return false
}

// spacedReference adds to a record refusal that whitespace around a lone reference made
// the template text.
func spacedReference(template string, err error) error {
	if trimmed := strings.TrimSpace(template); trimmed != template && errors.Is(err, fejkdata.ErrNoColumns) {
		if path, lone := loneReference(trimmed); lone {
			return fmt.Errorf("%w; the whitespace around {/%s} makes it text, so remove it to write the record %s", err, path, path)
		}
	}
	return err
}

// piped is the command piping template to fejkdata; some shells' echo reads a backslash
// as an escape, so a template holding one goes through printf.
func piped(template string) string {
	if strings.ContainsRune(template, '\\') {
		return "printf '%s' " + shellQuoted(template) + " | fejkdata"
	}
	return "echo " + shellQuoted(template) + " | fejkdata"
}

// shellQuoted is s as a shell reads it back: bare where it holds nothing a shell expands.
func shellQuoted(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\r\n'\"$`\\*?[]{}()<>|&;#~!") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func main() {
	var stdin io.Reader = os.Stdin
	if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		stdin = nil
	}
	os.Exit(run(os.Args[1:], stdin, os.Stdout, os.Stderr))
}

// run returns the exit code: 0 ok, 1 runtime error, 2 misuse. A nil stdin is a
// terminal, which holds no template.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	in, err := parseArgs(args)
	if err != nil {
		return misuse(stderr, err)
	}
	if in.help {
		io.WriteString(stdout, usage)
		return 0
	}
	if in.version {
		fmt.Fprintln(stdout, "fejkdata "+buildVersion())
		return 0
	}
	if err := in.check(); err != nil {
		return misuse(stderr, err)
	}
	var src input
	if !in.list {
		if src, err = readStdin(stdin); err != nil {
			return fail(stderr, err)
		}
	}
	f, err := fejkdata.New(in.options()...)
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
	if err := in.write(f, src, stdout); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func readStdin(stdin io.Reader) (input, error) {
	if stdin == nil {
		return input{}, inputError{fmt.Errorf("the template comes from stdin, and nothing is piped in: %s", pipeHint)}
	}
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return input{}, inputError{fmt.Errorf("stdin cannot be read: %w", err)}
	}
	return parseInput(string(raw)), nil
}

func fail(stderr io.Writer, err error) int {
	var te inputError
	if errors.As(err, &te) {
		return misuse(stderr, te.error)
	}
	fmt.Fprintln(stderr, err)
	if errors.Is(err, fejkdata.ErrNoColumns) {
		fmt.Fprintln(stderr, "wrap that JSON in single quotes, which keep a shell from expanding its braces")
	}
	return 1
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
