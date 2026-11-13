package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTodaySpellings(t *testing.T) {
	for _, args := range [][]string{{"--today", "2030-06-15"}, {"--today=2030-06-15"}} {
		t.Run(argsName(args), func(t *testing.T) {
			code, out, errb := runOut("{born(0,0,'2006')}", args...)
			if code != 0 || out != "2030" {
				t.Errorf("run(%v) = %d, %q, %q; want 2030", args, code, out, errb)
			}
		})
	}
}

func TestRunRefusesAMalformedToday(t *testing.T) {
	for _, v := range []string{"tomorrow", "", "2030-6-15", "2030-02-30", "20300615", "2030-06-15T12:00:00Z", "0000-06-15", "0001-01-01", "0999-12-31", "10000-01-01"} {
		t.Run(fmt.Sprintf("%q", v), func(t *testing.T) {
			code, out, errb := runOut("x", "--today", v)
			if code != 2 || out != "" || !strings.Contains(errb, "--today") || !strings.Contains(errb, "YYYY-MM-DD") {
				t.Errorf("run(--today %q) = %d, %q, %q; want misuse naming --today and YYYY-MM-DD", v, code, out, errb)
			}
		})
	}
}

func TestRunSeedSpellings(t *testing.T) {
	_, want, _ := runOut("{/address}", "--seed", "42", "--data-path", svSE)
	for _, args := range [][]string{
		{"--seed=42", "--data-path", svSE},
		{"-s", "42", "-d", svSE},
		{"--data-path", svSE, "--seed", "42"},
	} {
		t.Run(argsName(args), func(t *testing.T) {
			code, got, errb := runOut("{/address}", args...)
			if code != 0 {
				t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
			}
			if got != want {
				t.Errorf("run(%v) = %q, want %q", args, got, want)
			}
		})
	}
}

func TestRunShortFlagValues(t *testing.T) {
	_, want, _ := runOut("{/address}", "--seed", "42", "--data-path", svSE)
	for _, args := range [][]string{
		{"-s42", "-d", svSE},
		{"-s", "42", "-d" + svSE},
		{"-d", svSE, "-s42"},
	} {
		t.Run(argsName(args), func(t *testing.T) {
			code, got, errb := runOut("{/address}", args...)
			if code != 0 {
				t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
			}
			if got != want {
				t.Errorf("run(%v) = %q, want %q", args, got, want)
			}
		})
	}
}

func TestRunShortRepeatTakesItsValueAttached(t *testing.T) {
	_, three, _ := runOut("{/word}\n", "-s", "1", "-n3", "-d", svSE)
	if lines := strings.Split(strings.TrimRight(three, "\n"), "\n"); len(lines) != 3 {
		t.Errorf("-n3 gave %d lines: %q", len(lines), three)
	}
}

func TestRunMalformedFlagValueIsMisuse(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-s=42", "-d", svSE}, "-s42"},
		{[]string{"-s=42", "-d", svSE}, "--seed=42"},
		{[]string{"-d=" + svSE}, "--data-path="},
		{[]string{"-nd", "3", "-d", svSE}, `--repeat needs an integer in 0..1048576, got "d"`},
		{[]string{"--seed=", "-d", svSE}, `--seed needs an unsigned integer, got ""`},
	} {
		t.Run(argsName(c.args)+": "+c.want, func(t *testing.T) {
			code, out, errb := runOut("{/word}", c.args...)
			if code != 2 || out != "" {
				t.Errorf("run(%v) = %d, stdout %q, want misuse", c.args, code, out)
			}
			if !strings.Contains(errb, c.want) {
				t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
			}
		})
	}
}

func TestRunBundledShortHelpPrintsUsage(t *testing.T) {
	code, out, _ := runOut("", "-hd", svSE)
	if code != 0 || !strings.Contains(out, "Usage") {
		t.Errorf("-hd (bundled help) = %d, %q, want usage", code, out)
	}
}

func TestRunDoubleDashEndsFlags(t *testing.T) {
	code, out, errb := runOut("{/person}", "--data-path", svSE, "--")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Fatalf("run = %d, out=%q, stderr=%q", code, out, errb)
	}
	code, _, errb = runOut("{/person}", "--data-path", svSE, "--", "--list")
	if code != 2 || !strings.Contains(errb, "stdin") {
		t.Errorf("after --, --list should be an argument: code %d, stderr %q", code, errb)
	}
}

func TestRunSingleDashLongIsRejected(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-seed", "42", "-d", svSE}, "use --seed"},
		{[]string{"-seed=42", "-d", svSE}, "use --seed"},
		{[]string{"-data-path", svSE}, "use --data-path"},
		{[]string{"-list", "-d", svSE}, "use --list"},
		{[]string{"--nope", "-d", svSE}, "unknown flag --nope"},
		{[]string{"-x", "-d", svSE}, "unknown flag -x"},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			code, _, errb := runOut("{/person}", c.args...)
			if code != 2 {
				t.Errorf("run(%v) = %d, want 2", c.args, code)
			}
			if !strings.Contains(errb, c.want) {
				t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
			}
		})
	}
}

func TestRunFlagValues(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--seed", "-d", svSE}, `--seed needs an unsigned integer, got "-d"`},
		{[]string{"-d", svSE, "--seed"}, "--seed needs a value"},
		{[]string{"--seed", "x", "-d", svSE}, "--seed"},
		{[]string{"--list=1", "-d", svSE}, "--list takes no value"},
		{[]string{"--repeat", "-1", "-d", svSE}, "--repeat"},
		{[]string{"-n", "-1", "-d", svSE}, "--repeat"},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			code, _, errb := runOut("{/person}", c.args...)
			if code != 2 {
				t.Errorf("run(%v) = %d, want 2", c.args, code)
			}
			if !strings.Contains(errb, c.want) {
				t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
			}
		})
	}
}

func TestRunIgnoresAFlagWithNoEffect(t *testing.T) {
	dir := recordDir(t)
	for _, c := range []struct {
		stdin         string
		base, ignored []string
	}{
		{"", []string{"--list"}, []string{"-n", "3"}},
		{"", []string{"--list"}, []string{"--separator", ","}},
		{"", []string{"--list"}, []string{"--format", "json"}},
		{"", []string{"--list"}, []string{"--table", "t"}},
		{"{/users}", nil, []string{"--table", "t"}},
		{"{/users}", []string{"--format", "json"}, []string{"--table", "t"}},
	} {
		t.Run(argsName(c.base)+" + "+argsName(c.ignored), func(t *testing.T) {
			args := append([]string{"--seed", "1", "--data-path", dir}, c.base...)
			_, want, _ := runOut(c.stdin, args...)
			code, got, errb := runOut(c.stdin, append(args, c.ignored...)...)
			if code != 0 || want == "" || got != want {
				t.Errorf("run(%v + %v) = %d, %q, %q; want what %v alone prints, %q", c.base, c.ignored, code, got, errb, c.base, want)
			}
		})
	}
}

func TestRunUnknownFlagIsNamedByRune(t *testing.T) {
	code, _, errb := runOut("{/sv_SE.word}", "-ä")
	if code != 2 || !strings.Contains(errb, "unknown flag -ä") {
		t.Errorf("run(-ä) = %d, %q, want the flag named whole", code, errb)
	}
}

func TestRunEmptyDataPathIsNamed(t *testing.T) {
	code, _, errb := runOut("{/sv_SE.word}", "--data-path=")
	if code != 1 || !strings.Contains(errb, "empty") {
		t.Errorf("run(--data-path=) = %d, %q, want the empty path named", code, errb)
	}
}

func flagAsHelpAndReadmeListIt(f flagDef) (inHelp, inReadme string) {
	if f.short != "" {
		return "  -" + f.short + ", --" + f.long, "`-" + f.short + "`, `--" + f.long
	}
	return "      --" + f.long, "`--" + f.long
}

func TestHelpAndReadmeNameEveryFlag(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range flagDefs {
		t.Run(f.long, func(t *testing.T) {
			inHelp, inReadme := flagAsHelpAndReadmeListIt(f)
			if !strings.Contains(usage, inHelp) {
				t.Errorf("--help does not list %q", inHelp)
			}
			if !strings.Contains(string(readme), inReadme) {
				t.Errorf("README's flag table does not list %q", inReadme)
			}
		})
	}
}
