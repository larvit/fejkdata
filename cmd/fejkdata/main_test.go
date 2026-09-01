package main

import (
	"bytes"
	"strings"
	"testing"
)

const (
	svSE = "../../data/sv_SE"
	enUS = "../../data/en_US"
)

func runOut(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunOutputsValue(t *testing.T) {
	code, out, errb := runOut("--data-path", svSE, "person")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("empty output, stderr=%q", errb)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output should end with newline, got %q", out)
	}
}

func TestRunDotPath(t *testing.T) {
	code, full, _ := runOut("--seed", "7", "--data-path", svSE, "person")
	if code != 0 {
		t.Fatalf("person run = %d", code)
	}
	code, last, errb := runOut("--seed", "7", "--data-path", svSE, "person.last")
	if code != 0 {
		t.Fatalf("person.last run = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(last) == "" {
		t.Fatal("empty person.last output")
	}
	if last == full {
		t.Errorf("person.last %q should differ from person %q", last, full)
	}
}

func TestRunSeedSpellings(t *testing.T) {
	_, want, _ := runOut("--seed", "42", "--data-path", svSE, "address")
	for _, args := range [][]string{
		{"--seed=42", "--data-path", svSE, "address"},
		{"-s", "42", "-d", svSE, "address"},
		{"--data-path", svSE, "address", "--seed", "42"},
	} {
		code, got, errb := runOut(args...)
		if code != 0 {
			t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
		}
		if got != want {
			t.Errorf("run(%v) = %q, want %q", args, got, want)
		}
	}
}

func TestRunDoubleDashEndsFlags(t *testing.T) {
	code, out, errb := runOut("--data-path", svSE, "--", "person")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Fatalf("run = %d, out=%q, stderr=%q", code, out, errb)
	}
	code, _, errb = runOut("--data-path", svSE, "--", "--list")
	if code != 1 || !strings.Contains(errb, "--list") {
		t.Errorf("after --, --list should be a path: code %d, stderr %q", code, errb)
	}
}

func TestRunSingleDashLongIsRejected(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-seed", "42", "-d", svSE, "person"}, "use --seed"},
		{[]string{"-seed=42", "-d", svSE, "person"}, "use --seed"},
		{[]string{"-data-path", svSE, "person"}, "use --data-path"},
		{[]string{"-list", "-d", svSE}, "use --list"},
		{[]string{"--nope", "-d", svSE, "person"}, "unknown flag --nope"},
		{[]string{"-x", "-d", svSE, "person"}, "unknown flag -x"},
	} {
		code, _, errb := runOut(c.args...)
		if code != 2 {
			t.Errorf("run(%v) = %d, want 2", c.args, code)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
		}
	}
}

func TestRunFlagValues(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--seed", "-d", svSE, "person"}, `--seed needs an unsigned integer, got "-d"`},
		{[]string{"-d", svSE, "person", "--seed"}, "--seed needs a value"},
		{[]string{"--seed", "x", "-d", svSE, "person"}, "--seed"},
		{[]string{"--list=1", "-d", svSE}, "--list takes no value"},
		{[]string{"--repeat", "0", "-d", svSE, "person"}, "--repeat"},
		{[]string{"-n", "-1", "-d", svSE, "person"}, "--repeat"},
	} {
		code, _, errb := runOut(c.args...)
		if code != 2 {
			t.Errorf("run(%v) = %d, want 2", c.args, code)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
		}
	}
}

func TestRunRepeat(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "--repeat", "3", "--data-path", svSE, "word")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("repeat=3 gave %d lines: %q", len(lines), out)
	}
	_, short, _ := runOut("--seed", "1", "-n", "3", "-d", svSE, "word")
	if short != out {
		t.Errorf("-n 3 = %q, want the same as --repeat 3 %q", short, out)
	}
}

func TestRunSeparator(t *testing.T) {
	code, out, errb := runOut("--repeat", "3", "--separator", ",", "--data-path", svSE, "word")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if n := strings.Count(out, "\n"); n != 1 {
		t.Errorf("want one trailing newline, got %d: %q", n, out)
	}
	if !strings.Contains(out, ",") {
		t.Errorf("values should be comma-joined: %q", out)
	}
}

func TestRunRepeatAdvancesRNG(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "--repeat", "5", "--data-path", svSE, "person")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	uniq := map[string]bool{}
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		uniq[l] = true
	}
	if len(uniq) < 2 {
		t.Errorf("repeat should vary output, all identical: %q", out)
	}
}

func TestRunMisuse(t *testing.T) {
	for _, args := range [][]string{{}, {"--data-path", svSE}, {"person"}, {"-d", svSE, "person", "word"}, {"-d", svSE, "--list", "person"}} {
		code, out, errb := runOut(args...)
		if code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
		if !strings.Contains(errb, "try 'fejkdata --help'") {
			t.Errorf("run(%v) stderr = %q, want a pointer to --help", args, errb)
		}
		if out != "" {
			t.Errorf("run(%v) stdout = %q, want nothing", args, out)
		}
	}
}

func TestRunMultipleDataPaths(t *testing.T) {
	code, out, errb := runOut("-d", enUS, "--data-path", svSE, "person")
	if code != 0 {
		t.Fatalf("run(multi-dir) = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty output for multi-dir run")
	}
}

func TestRunUnknownCategoryFails(t *testing.T) {
	code, _, errb := runOut("--data-path", svSE, "nope")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if !strings.Contains(errb, "nope") {
		t.Errorf("stderr %q should name the unknown category", errb)
	}
}

func TestRunList(t *testing.T) {
	code, out, errb := runOut("--data-path", svSE, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	for _, want := range []string{"person", "person.last", "address"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q:\n%s", want, out)
		}
	}
}

func TestRunHelpOnStdout(t *testing.T) {
	for _, h := range []string{"-h", "--help"} {
		code, out, errb := runOut(h)
		if code != 0 {
			t.Errorf("%s = %d, want 0", h, code)
		}
		if !strings.Contains(out, "Usage") || errb != "" {
			t.Errorf("%s: stdout = %q, stderr = %q, want usage on stdout only", h, out, errb)
		}
	}
}

func TestRunVersion(t *testing.T) {
	code, out, errb := runOut("--version")
	if code != 0 {
		t.Fatalf("--version = %d, stderr=%q", code, errb)
	}
	version, ok := strings.CutPrefix(out, "fejkdata ")
	if !ok || strings.TrimSpace(version) == "" {
		t.Errorf("--version = %q, want the command name and a version on stdout", out)
	}
}

func TestRunMissingDirFails(t *testing.T) {
	code, _, errb := runOut("--data-path", "../../data/nope", "person")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if errb == "" {
		t.Error("want an error message on stderr")
	}
}
