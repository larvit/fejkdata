// Command fejkdata prints fake values from the shipped data and any directories
// layered over it.
//
//	fejkdata sv_SE.person                        # a full person
//	fejkdata sv_SE.person.last                   # just the surname
//	fejkdata --data-path ./mydata sv_SE.person   # layer custom data; the last dir wins
//	fejkdata --seed 42 sv_SE.address
package main

import (
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

  <path>              a category, or a dotted path into one (person, person.last)

  -d, --data-path D      a data directory to layer over the shipped data (repeatable; last wins on a clash)
  -h, --help             print this help, then exit
      --list             list the paths the data offers, then exit
      --no-shipped-data  load only the --data-path directories
  -n, --repeat N         render the path N times (default 1)
  -s, --seed N           seed for reproducible output
      --separator S      string between repeated values (default newline)
      --version          print the version, then exit

Flags may come before or after <path>; -- ends the flags.
`

type invocation struct {
	dirs      []string
	help      bool
	list      bool
	noShipped bool
	paths     []string
	repeat    int
	seed      uint64
	seeded    bool
	separator string
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
	{"help", "h", false, func(in *invocation, _ string) error { in.help = true; return nil }},
	{"list", "", false, func(in *invocation, _ string) error { in.list = true; return nil }},
	{"no-shipped-data", "", false, func(in *invocation, _ string) error { in.noShipped = true; return nil }},
	{"repeat", "n", true, func(in *invocation, v string) error {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return fmt.Errorf("--repeat needs a positive integer, got %q", v)
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

func parseArgs(argv []string) (invocation, error) {
	in := invocation{repeat: 1, separator: "\n"}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "--":
			in.paths = append(in.paths, argv[i+1:]...)
			return in, nil
		case strings.HasPrefix(arg, "--"):
			name, value, hasValue := strings.Cut(arg[2:], "=")
			def := flagByLong(name)
			if def == nil {
				return in, fmt.Errorf("unknown flag --%s", name)
			}
			if !def.value {
				if hasValue {
					return in, fmt.Errorf("--%s takes no value", name)
				}
				_ = def.set(&in, "")
				continue
			}
			if !hasValue {
				if i++; i >= len(argv) {
					return in, fmt.Errorf("--%s needs a value", name)
				}
				value = argv[i]
			}
			if err := def.set(&in, value); err != nil {
				return in, err
			}
		case len(arg) > 1 && arg[0] == '-':
			name, _, _ := strings.Cut(arg[1:], "=")
			def := flagByShort(name)
			if def == nil {
				if flagByLong(name) != nil {
					return in, fmt.Errorf("unknown flag %s; use --%s", arg, name)
				}
				return in, fmt.Errorf("unknown flag %s", arg)
			}
			if !def.value {
				_ = def.set(&in, "")
				continue
			}
			if i++; i >= len(argv) {
				return in, fmt.Errorf("--%s needs a value", def.long)
			}
			if err := def.set(&in, argv[i]); err != nil {
				return in, err
			}
		default:
			in.paths = append(in.paths, arg)
		}
	}
	return in, nil
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
	if in.noShipped && len(in.dirs) == 0 {
		return misuse(stderr, errors.New("--no-shipped-data needs at least one --data-path"))
	}
	if in.list && len(in.paths) > 0 {
		return misuse(stderr, errors.New("--list takes no path"))
	}
	if !in.list && len(in.paths) != 1 {
		return misuse(stderr, fmt.Errorf("expected one path, got %d", len(in.paths)))
	}

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
	f, err := fejkdata.New(opts...)
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
	vals := make([]string, in.repeat)
	for i := range vals {
		if vals[i], err = f.Fake(in.paths[0]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	fmt.Fprintln(stdout, strings.Join(vals, in.separator))
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
