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
	"strconv"
	"strings"

	"gitea.larvit.se/larvit/fejkdata"
)

const usage = `Usage: fejkdata [flags] <path>

  <path>                 a category, or a dotted path into one (person, person.last)

  -d, --data-path D      a data directory to layer over the shipped data (repeatable; last wins on a clash)
  -h, --help             print this help, then exit
      --list             list the paths the data offers, then exit
      --no-shipped-data  load only the --data-path directories
  -n, --repeat N         render the path N times, 1..1048576 (default 1)
  -s, --seed N           seed for reproducible output
      --separator S      string between repeated values (default newline)
      --version          print the version, then exit

Flags may come before or after <path>; -- ends the flags. A short flag's value
attaches or follows (-n3, -n 3); short flags bundle (-hn 3).
`

type invocation struct {
	dirs         []string
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
	in := invocation{repeat: 1, separator: "\n"}
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

// check rejects a flag combination that cannot run.
func (in invocation) check() error {
	if in.list && len(in.paths) > 0 {
		return errors.New("--list takes no path")
	}
	if in.list && (in.repeatSet || in.separatorSet) {
		return errors.New("--list takes no --repeat or --separator")
	}
	if !in.list && len(in.paths) != 1 {
		return fmt.Errorf("expected one path, got %d", len(in.paths))
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

// write streams the path's renders to w, repeat of them joined by the separator
// and ended by a newline. A path that renders once renders every time, so a Fake
// failure comes before anything is written; a write failure surfaces from Flush,
// bufio keeping the first one.
func (in invocation) write(f *fejkdata.Generator, w io.Writer) error {
	out := bufio.NewWriter(w)
	for i := 0; i < in.repeat; i++ {
		v, err := f.Fake(in.paths[0])
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
	if err := in.check(); err != nil {
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
	if err := in.write(f, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func misuse(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "fejkdata: %v\ntry 'fejkdata --help'\n", err)
	return 2
}

// buildVersion is the module version go install stamps into the binary, or "devel".
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "devel"
}
