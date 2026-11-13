// Package cli is the fejkdata command: it renders the template on its stdin with fake
// values from the modules it is given and any directories layered over them. A program
// passing its own modules to Main is a fejkdata CLI carrying them, for example
// os.Exit(cli.Main(append(data.Modules(), mydata)...)).
package cli

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/larvit/fejkdata"
)

const usage = `Usage: <template> | COMMAND [flags]
       COMMAND --list | --help | --version

COMMAND reads one template from stdin and renders it:

  echo 'id: {int(1,999)}' | COMMAND
  echo '{"format":"{x}","x":["bosse","lina"]}' | COMMAND
  COMMAND < template.txt

COMMAND <<'EOF'
born {date(1990-01-01,2010-12-31,'2006-01-02')}
EOF

A template is a format string, or a JSON object, array, string or null. Its {…} tokens
reach the data by reference from the root: {/path}, for any path --list prints, shipped
or layered with --data-path. A reference names a category or a dotted path into one, a
table's row by key or name ({/table[KEY]}), and .. steps up from a row to the row it
links to.

Each {…} draws afresh, unless it reads a name. {/path as p} draws once, binds the pick
to p and prints nothing, and every {p} and {p.field} reads that one pick, so values that
belong together agree:
  DOCS#names

{name(args)} calls a builtin, such as {lower(8)} for eight letters a-z. The builtins:
BUILTINS
What each takes and renders:
  DOCS#functions
The whole template language:
  DOCS#data-format
Writing a --data-path directory:
  DOCS#your-own-data

Under --format text, the default, what you send is what renders, and COMMAND adds
nothing: a format string keeps every byte, so echo's newline ends each render, and
printf '%s' sends none. -n joins renders with --separator, empty by default. A JSON
object, array or string drops the whitespace around it, and null alone renders nothing;
42 or true alone is a format string, printed as written. An empty or blank stdin is a
format string like any other, and renders as sent. A quoted heredoc, <<'EOF', passes $,
a backslash or a quote as written.

With --format json, ndjson, csv, tsv, pgcopy or sql the template must be a record — a JSON
template whose fields are its columns, or a lone reference such as {/path}:
one reference and nothing else, with one newline after it allowed — and the rows are
written as one JSON array, one JSON object per line, one CSV or tab-separated row (after a
header), a COPY block, or INSERTs of up to 1000 rows, one per line. A tsv file is what a table reads,
so a later run can draw from this run's records: it writes a brace doubled, and a tab
or line break in a value as a space.

  -d, --data-path D      a data directory to layer over the shipped data (repeatable; last wins on a clash)
      --format F         output form: text (default), json, ndjson, csv, tsv, pgcopy or sql
  -h, --help             print this help, then exit
      --list             list the paths the data offers, each written after {/ in a template, then exit;
                         --repeat, --separator, --format and --table do nothing here
      --no-shipped-data  load only the --data-path directories
  -n, --repeat N         render the template N times, 0..1048576 (default 1); 0 prints nothing,
                         [] under --format json, or the header under --format csv or tsv
  -s, --seed N           same seed, date, data and versions: identical output (default: a random seed)
      --separator S      string between repeated renders (default empty); refused with a record
                         --format, which has its own separators
      --table T          the INSERT or COPY target for --format sql or pgcopy, ignored under any other
                         (default: a lone reference's last segment, else records)
      --today DAY        the date to treat as today, as YYYY-MM-DD (default: the system's date);
                         give it beside --seed to repeat a run's output on a later day
      --version          print the version, then exit

A short flag's value attaches or follows (-n3, -n 3); short flags bundle (-hn 3);
-- ends the flags.

Exit codes: 0 success, 1 runtime error, 2 misuse.
`

type invocation struct {
	args         []string
	dirs         []string
	format       string
	help         bool
	list         bool
	noShipped    bool
	repeat       int
	seed         uint64
	seeded       bool
	separator    string
	separatorSet bool
	table        string
	tableSet     bool
	today        time.Time
	version      bool
}

type flagDef struct {
	long  string
	short string
	value bool
	set   func(*invocation, string) error
}

func (in invocation) options(modules []fs.FS, sys system) []fejkdata.Option {
	if in.noShipped {
		modules = nil
	}
	opts := []fejkdata.Option{fejkdata.WithDataFS(modules...)}
	for _, dir := range in.dirs {
		opts = append(opts, fejkdata.WithDataPath(dir))
	}
	seed := in.seed
	if !in.seeded {
		seed = sys.seed
	}
	today := in.today
	if today.IsZero() {
		today = sys.today
	}
	return append(opts, fejkdata.WithSeed(seed), fejkdata.WithToday(today))
}

// write streams the input's renders to w, repeat of them joined by the separator; a write
// failure surfaces from Flush, bufio keeping the first one.
func (in invocation) write(f *fejkdata.Generator, src input, w io.Writer) error {
	if in.format != "text" {
		return in.writeRecords(f, src, w)
	}
	t, err := f.NewTemplate(src.template)
	if err != nil {
		return templateError(err)
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

// Main runs the command on the process's arguments and standard streams over modules, each
// passed to [fejkdata.WithDataFS], and returns the exit code: 0 ok, 1 runtime error,
// 2 misuse. --no-shipped-data leaves modules out.
func Main(modules ...fs.FS) int {
	var stdin io.Reader = os.Stdin
	if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		stdin = nil
	}
	return run(commandName(os.Args[0]), modules, os.Args[1:], system{rand.Uint64(), time.Now()}, stdin, os.Stdout, os.Stderr)
}

// commandName is the name the command runs under: the file arg0 names, less a Windows
// .exe, or fejkdata where arg0 names none.
func commandName(arg0 string) string {
	name := filepath.Base(arg0)
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".exe") {
		name = strings.TrimSuffix(name, ext)
	}
	if name == "." || name == "" {
		return "fejkdata"
	}
	return name
}

// run is Main over the given streams, run as the command name, with sys's seed and date where no
// flag gives them. A nil stdin is a terminal, which holds no template.
func run(name string, modules []fs.FS, args []string, sys system, stdin io.Reader, stdout, stderr io.Writer) int {
	in, err := parseArgs(args)
	if err != nil {
		return misuse(stderr, name, err)
	}
	if in.help {
		info, _ := debug.ReadBuildInfo()
		io.WriteString(stdout, help(name, docsURL(info)))
		return 0
	}
	if in.version {
		info, _ := debug.ReadBuildInfo()
		fmt.Fprintln(stdout, versionLine(name, info))
		return 0
	}
	if err := in.check(name); err != nil {
		return misuse(stderr, name, err)
	}
	var src input
	if !in.list {
		if src, err = readStdin(stdin, name); err != nil {
			return fail(stderr, name, err)
		}
	}
	f, err := fejkdata.New(in.options(modules, sys)...)
	if err != nil {
		return fail(stderr, name, err)
	}
	if in.list {
		for _, p := range f.List() {
			fmt.Fprintln(stdout, p)
		}
		return 0
	}
	if err := in.write(f, src, stdout); err != nil {
		return fail(stderr, name, err)
	}
	return 0
}

func help(name, docs string) string {
	var lines []string
	line := " "
	for _, b := range fejkdata.Builtins() {
		if len(line)+1+len(b) > 88 {
			lines = append(lines, line)
			line = " "
		}
		line += " " + b
	}
	lines = append(lines, line)
	return strings.NewReplacer("COMMAND", name, "DOCS", docs, "BUILTINS", strings.Join(lines, "\n")).Replace(usage)
}
