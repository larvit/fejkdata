package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/larvit/fejkdata"
)

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
	{"separator", "", true, func(in *invocation, v string) error { in.separator, in.separatorSet = v, true; return nil }},
	{"table", "", true, func(in *invocation, v string) error { in.table, in.tableSet = v, true; return nil }},
	{"today", "", true, func(in *invocation, v string) error {
		day, err := time.Parse(time.DateOnly, v)
		if err != nil || day.Year() < 1000 {
			return fmt.Errorf("--today needs a date as YYYY-MM-DD, from year 1000 on, got %q", v)
		}
		in.today = day
		return nil
	}},
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

// checkFlags rejects a flag value or combination that cannot run.
func (in invocation) checkFlags() error {
	if _, ok := recordFormats[in.format]; !ok && in.format != "text" {
		return fmt.Errorf("--format takes %s, got %q", formatNames(), in.format)
	}
	if target := recordFormats[in.format].target; in.tableSet && in.table == "" && target != "" {
		return fmt.Errorf("--table names the %s target, so it cannot be empty", target)
	}
	if f, records := recordFormats[in.format]; records && in.separatorSet && !in.list {
		return fmt.Errorf("--separator joins text renders; --format %s %s", in.format, f.writes)
	}
	return nil
}

func pipeHint(name string) string { return "echo 'id: {int(1,999)}' | " + name }

// check rejects a flag combination or an operand that cannot run, before any data is
// read.
func (in invocation) check(name string) error {
	if err := in.checkFlags(); err != nil {
		return err
	}
	if len(in.args) > 0 {
		return operandError(in.args[0], name)
	}
	return nil
}
