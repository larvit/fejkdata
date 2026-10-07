package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	svSE = "../../data/sv_SE"
	enUS = "../../data/en_US"
	misc = "../../data/misc"
)

func runOut(stdin string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

// runTerminal runs as from a shell with nothing piped in.
func runTerminal(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, nil, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunOutputsValue(t *testing.T) {
	code, out, errb := runOut("{/person}", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if out == "" || strings.Contains(out, "\n") {
		t.Fatalf("output = %q, stderr=%q; want one value and nothing added", out, errb)
	}
}

func TestRunDotPath(t *testing.T) {
	code, full, _ := runOut("{/person}", "--seed", "7", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("person run = %d", code)
	}
	code, last, errb := runOut("{/person.last}", "--seed", "7", "--data-path", svSE)
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
	_, want, _ := runOut("{/address}", "--seed", "42", "--data-path", svSE)
	for _, args := range [][]string{
		{"--seed=42", "--data-path", svSE},
		{"-s", "42", "-d", svSE},
		{"--data-path", svSE, "--seed", "42"},
	} {
		code, got, errb := runOut("{/address}", args...)
		if code != 0 {
			t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
		}
		if got != want {
			t.Errorf("run(%v) = %q, want %q", args, got, want)
		}
	}
}

func TestRunShortFlagValues(t *testing.T) {
	_, want, _ := runOut("{/address}", "--seed", "42", "--data-path", svSE)
	for _, args := range [][]string{
		{"-s42", "-d", svSE},
		{"-s", "42", "-d" + svSE},
		{"-d", svSE, "-s42"},
	} {
		code, got, errb := runOut("{/address}", args...)
		if code != 0 {
			t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
		}
		if got != want {
			t.Errorf("run(%v) = %q, want %q", args, got, want)
		}
	}
	_, three, _ := runOut("{/word}\n", "-s", "1", "-n3", "-d", svSE)
	if lines := strings.Split(strings.TrimRight(three, "\n"), "\n"); len(lines) != 3 {
		t.Errorf("-n3 gave %d lines: %q", len(lines), three)
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-s=42", "-d", svSE}, "-s42"},
		{[]string{"-s=42", "-d", svSE}, "--seed=42"},
		{[]string{"-d=" + svSE}, "--data-path="},
		{[]string{"-nd", "3", "-d", svSE}, `--repeat needs an integer in 1..1048576, got "d"`},
		{[]string{"--seed=", "-d", svSE}, `--seed needs an unsigned integer, got ""`},
	} {
		code, out, errb := runOut("{/word}", c.args...)
		if code != 2 || out != "" {
			t.Errorf("run(%v) = %d, stdout %q, want misuse", c.args, code, out)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
		}
	}
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
		code, _, errb := runOut("{/person}", c.args...)
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
		{[]string{"--seed", "-d", svSE}, `--seed needs an unsigned integer, got "-d"`},
		{[]string{"-d", svSE, "--seed"}, "--seed needs a value"},
		{[]string{"--seed", "x", "-d", svSE}, "--seed"},
		{[]string{"--list=1", "-d", svSE}, "--list takes no value"},
		{[]string{"--repeat", "0", "-d", svSE}, "--repeat"},
		{[]string{"-n", "-1", "-d", svSE}, "--repeat"},
	} {
		code, _, errb := runOut("{/person}", c.args...)
		if code != 2 {
			t.Errorf("run(%v) = %d, want 2", c.args, code)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
		}
	}
}

func TestRunRepeat(t *testing.T) {
	code, out, errb := runOut("{/word}\n", "--seed", "1", "--repeat", "3", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("repeat=3 gave %d lines: %q", len(lines), out)
	}
	_, short, _ := runOut("{/word}\n", "--seed", "1", "-n", "3", "-d", svSE)
	if short != out {
		t.Errorf("-n 3 = %q, want the same as --repeat 3 %q", short, out)
	}
}

func TestRunSeparator(t *testing.T) {
	code, out, errb := runOut("{/word}", "--repeat", "3", "--separator", ",", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if n := strings.Count(out, "\n"); n != 0 {
		t.Errorf("want no newline, got %d: %q", n, out)
	}
	if !strings.Contains(out, ",") {
		t.Errorf("values should be comma-joined: %q", out)
	}
}

func TestRunRepeatAdvancesRNG(t *testing.T) {
	code, out, errb := runOut("{/person}\n", "--seed", "1", "--repeat", "5", "--data-path", svSE)
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
	for _, c := range []struct {
		stdin string
		args  []string
	}{
		{"", nil},
		{"", []string{"--data-path", svSE}},
		{"{/person}", []string{"-d", svSE, "word"}},
		{"", []string{"-d", svSE, "--list", "person"}},
		{"{/sv_SE.person}", []string{"--no-shipped-data"}},
	} {
		code, out, errb := runOut(c.stdin, c.args...)
		if code != 2 {
			t.Errorf("run(%q, %v) = %d, want 2", c.stdin, c.args, code)
		}
		if !strings.Contains(errb, "try 'fejkdata --help'") {
			t.Errorf("run(%q, %v) stderr = %q, want a pointer to --help", c.stdin, c.args, errb)
		}
		if out != "" {
			t.Errorf("run(%q, %v) stdout = %q, want nothing", c.stdin, c.args, out)
		}
	}
}

func TestRunTakesNoOperand(t *testing.T) {
	file := filepath.Join(t.TempDir(), "users.tmpl")
	if err := os.WriteFile(file, []byte("{/sv_SE.person}"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"sv_SE.person"}, "echo '{/sv_SE.person}' | fejkdata"},
		{[]string{"/en_US.address"}, "echo '{/en_US.address}' | fejkdata"},
		{[]string{"{/en_US.person}"}, "echo '{/en_US.person}' | fejkdata"},
		{[]string{"name: {/sv_SE.person.last}"}, "echo 'name: {/sv_SE.person.last}' | fejkdata"},
		{[]string{"geo.US.locality[O'Fallon].name"}, `echo '{/geo.US.locality[O'\''Fallon].name}' | fejkdata`},
		{[]string{"en_US.address"}, "echo '{/en_US.address}' | fejkdata"},
		{[]string{`C:\temp {/sv_SE.person}`}, `printf '%s' 'C:\temp {/sv_SE.person}' | fejkdata`},
		{[]string{"-"}, "drop it"},
		{[]string{file}, "fejkdata < " + file},
	} {
		code, out, errb := runOut("", c.args...)
		if code != 2 || out != "" || !strings.Contains(errb, "unexpected argument") || !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) = %d, %q, %q; want misuse naming %q", c.args, code, out, errb, c.want)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("bad file descriptor") }

func TestRunUnreadableStdinIsMisuse(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run(nil, failingReader{}, &out, &errb); code != 2 || !strings.Contains(errb.String(), "cannot be read") {
		t.Errorf("run with an unreadable stdin = %d, %q; want misuse naming it", code, errb.String())
	}
}

func TestRunTerminalStdinIsMisuse(t *testing.T) {
	code, out, errb := runTerminal("--seed", "1")
	if code != 2 || out != "" || !strings.Contains(errb, "| fejkdata") {
		t.Errorf("run with nothing piped = %d, %q, %q; want misuse showing a template piped to fejkdata", code, out, errb)
	}
	for _, args := range [][]string{{"--list"}, {"--help"}, {"--version"}} {
		if code, out, errb := runTerminal(args...); code != 0 || out == "" {
			t.Errorf("run(%v) with nothing piped = %d, %q, %q; want it to run, reading no stdin", args, code, out, errb)
		}
	}
}

func TestRunRendersExactlyWhatIsSent(t *testing.T) {
	for stdin, want := range map[string]string{
		"hihi":                           "hihi",
		"hihi\n":                         "hihi\n",
		"hihi\r\n":                       "hihi\r\n",
		"hihi\n\n":                       "hihi\n\n",
		"{{x}}\n":                        "{x}\n",
		"\"a\"\n":                        "a",
		" \t\"a\"\r\n\n":                 "a",
		"{/misc.territory[SE].alpha2}":   "SE",
		"{/misc.territory[SE].alpha2}\n": "SE\n",
	} {
		code, out, errb := runOut(stdin, "--seed", "1")
		if code != 0 || out != want {
			t.Errorf("run(%q) = %d, %q, %q; want %q", stdin, code, out, errb, want)
		}
	}
	if code, out, errb := runOut("{\"format\":\"{x}\\n\",\"x\":[\"a\",\"b\"]}\n"); code != 0 || (out != "a\n" && out != "b\n") {
		t.Errorf("a JSON template's format = %d, %q, %q; want its own newline only", code, out, errb)
	}
	if code, out, errb := runOut("{/sv_SE.word}\n", "--seed", "1", "-n", "3", "--separator", ","); code != 0 || strings.Count(out, "\n,") != 2 || !strings.HasSuffix(out, "\n") {
		t.Errorf("echo's newline and a separator = %d, %q, %q; want each render's newline, then the separator", code, out, errb)
	}
}

func TestRunRefusesBlankStdin(t *testing.T) {
	for _, stdin := range []string{"", "\n", " \r\n\t"} {
		if code, out, errb := runOut(stdin); code != 2 || out != "" || !strings.Contains(errb, "holds no template") {
			t.Errorf("run(%q) = %d, %q, %q; want misuse naming the empty stdin", stdin, code, out, errb)
		}
	}
}

func TestRunTextPrintsAsWritten(t *testing.T) {
	for _, stdin := range []string{"sv_SE.person", "[Skåne län]", "x[1]y", `say "hi"`} {
		code, out, errb := runOut(stdin)
		if code != 0 || out != stdin {
			t.Errorf("run(%q) = %d, %q, %q; want the text as written", stdin, code, out, errb)
		}
	}
}

func TestRunMultipleDataPaths(t *testing.T) {
	code, out, errb := runOut("{/person}", "-d", enUS, "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run(multi-dir) = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty output for multi-dir run")
	}
}

func TestRunUnknownCategoryFails(t *testing.T) {
	code, _, errb := runOut("{/nope}", "--data-path", svSE, "--format", "csv")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if !strings.Contains(errb, "nope") {
		t.Errorf("stderr %q should name the unknown category", errb)
	}
}

func TestRunList(t *testing.T) {
	code, out, errb := runOut("", "--data-path", svSE, "--list")
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
		code, out, errb := runOut("", h)
		if code != 0 {
			t.Errorf("%s = %d, want 0", h, code)
		}
		if !strings.Contains(out, "Usage") || errb != "" {
			t.Errorf("%s: stdout = %q, stderr = %q, want usage on stdout only", h, out, errb)
		}
	}
}

func TestRunVersion(t *testing.T) {
	code, out, errb := runOut("", "--version")
	if code != 0 {
		t.Fatalf("--version = %d, stderr=%q", code, errb)
	}
	version, ok := strings.CutPrefix(out, "fejkdata ")
	if !ok || strings.TrimSpace(version) == "" {
		t.Errorf("--version = %q, want the command name and a version on stdout", out)
	}
}

func TestRunMissingDirFails(t *testing.T) {
	code, _, errb := runOut("{/person}", "--data-path", "../../data/nope")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if errb == "" {
		t.Error("want an error message on stderr")
	}
}

func TestRunShippedDataByDefault(t *testing.T) {
	code, out, errb := runOut("{/sv_SE.person}", "--seed", "1")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Fatalf("run = %d, out=%q, stderr=%q", code, out, errb)
	}
	code, list, _ := runOut("", "--list")
	if code != 0 || !strings.Contains(list, "en_US.person\n") || !strings.Contains(list, "misc.uuid\n") {
		t.Errorf("--list without --data-path = %d, %q", code, list)
	}
	code, _, errb = runOut("{/person}")
	if code != 2 || !strings.Contains(errb, "person") {
		t.Errorf("a category outside the shipped tree: code %d, stderr %q", code, errb)
	}
}

func TestRunSelectsATableRow(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"region.json": `{"format":"{name}","rows":"region.tsv","key":"code","name":"name"}`,
		"region.tsv":  "code\tname\n01\tStockholms län\n12\tSkåne län\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if code, out, errb := runOut("{/region[12]}", "--no-shipped-data", "-d", dir); code != 0 || out != "Skåne län" {
		t.Fatalf("{/region[12]} = %d, %q, stderr=%q", code, out, errb)
	}
	if code, out, _ := runOut("{/region[Skåne län].code}", "--no-shipped-data", "-d", dir); code != 0 || out != "12" {
		t.Fatalf("{/region[Skåne län].code} = %d, %q", code, out)
	}
	if code, out, _ := runOut("{/region[12]}", "--no-shipped-data", "-d", dir, "--format", "sql"); code != 0 || out != `INSERT INTO "region" ("code", "name") VALUES ('12', 'Skåne län');`+"\n" {
		t.Fatalf("--format sql {/region[12]} = %d, %q, want the table named without its selector", code, out)
	}
	if code, _, errb := runOut("{/region[99]}", "--no-shipped-data", "-d", dir); code != 2 || !strings.Contains(errb, `"99"`) {
		t.Fatalf("{/region[99]} = %d, stderr=%q, want misuse naming the row", code, errb)
	}
	if code, out, _ := runOut("{/misc.territory[SE]}", "--seed", "1", "--format", "csv"); code != 0 || !strings.HasPrefix(out, "alpha2,") || !strings.Contains(out, "\nSE,SWE,") {
		t.Fatalf("--format csv {/misc.territory[SE]} = %d, %q", code, out)
	}
}

func TestUsageReferencesResolve(t *testing.T) {
	for _, token := range regexp.MustCompile(`\{/[^}]+\}`).FindAllString(usage, -1) {
		code, out, errb := runOut(token, "--seed", "1")
		if code != 0 || strings.TrimSpace(out) == "" {
			t.Errorf("usage advertises %s: run = %d, %q, stderr %q", token, code, out, errb)
		}
	}
}

func TestRunOperandMisuseBeforeLoad(t *testing.T) {
	code, _, errb := runOut("", "--no-shipped-data", "x")
	if code != 2 || !strings.Contains(errb, "unexpected argument") || strings.Contains(errb, "--data-path") {
		t.Fatalf("an operand with no data = %d, %q; want the operand refused before any load", code, errb)
	}
}

func TestRunInlineTemplate(t *testing.T) {
	code, out, errb := runOut("name: {/sv_SE.person.last}", "--seed", "1")
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline format string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut(`{"format":"name: {x}","x":["bosse","lina"]}`, "--seed", "1")
	if code != 0 || (out != "name: bosse" && out != "name: lina") {
		t.Fatalf("inline JSON template = %d, %q, want one name, stderr %q", code, out, errb)
	}
	code, out, errb = runOut(`"name: {/sv_SE.person.last}"`, "--seed", "1")
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline JSON string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("{digits(1)}\n", "--seed", "1", "-n", "2")
	if code != 0 || len(strings.Split(strings.TrimRight(out, "\n"), "\n")) != 2 {
		t.Fatalf("inline template with --repeat = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("{/sv_SE.person} hihi", "--seed", "1")
	if code != 0 || !strings.HasSuffix(out, " hihi") || strings.Contains(out, "{") {
		t.Fatalf("a reference and text = %d, %q, stderr %q", code, out, errb)
	}
}

func TestRunFormatTakesOnlyALoneReferenceAsARecord(t *testing.T) {
	for _, stdin := range []string{"{//sv_SE.person}", "{/sv_SE.person|/misc.uuid}", "{/sv_SE.person as p}", "{/}", "{/.sv_SE.person}"} {
		if code, out, errb := runOut(stdin, "--format", "json"); code != 2 || out != "" {
			t.Errorf("run(%q, --format json) = %d, %q, %q; want misuse", stdin, code, out, errb)
		}
	}
	if code, out, errb := runOut(`{"format":"x"}`); code != 0 || out != "x" {
		t.Errorf(`run({"format":"x"}) = %d, %q, %q; want x`, code, out, errb)
	}
}

func TestRunTemplateMisuse(t *testing.T) {
	for stdin, want := range map[string]string{
		"{bad":                  "unterminated",
		"name: {/no.such.path}": "no entry",
		"a } b":                 "}}",
		"42":                    "number",
		"42\n":                  "number",
		"true\n":                "boolean",
		"null\n":                "only null",
		" null ":                "only null",
		"{//sv_SE.person}":      "write {/sv_SE.person}",
	} {
		code, out, errb := runOut(stdin, "--seed", "1")
		if code != 2 || out != "" || !strings.Contains(errb, "try 'fejkdata --help'") || !strings.Contains(errb, want) {
			t.Errorf("run(%q) = %d, %q, %q; want misuse naming %q and --help", stdin, code, out, errb, want)
		}
		if strings.Contains(errb, "fejkdata: fejkdata:") {
			t.Errorf("run(%q) doubled the program prefix: %q", stdin, errb)
		}
	}
}

func TestRunNoShippedData(t *testing.T) {
	code, list, errb := runOut("", "--no-shipped-data", "-d", misc, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if strings.Contains(list, "sv_SE") || !strings.Contains(list, "uuid\n") {
		t.Errorf("--no-shipped-data --list = %q, want only the given dir", list)
	}
	code, out, _ := runOut("{/uuid}", "--no-shipped-data", "-d", misc, "-s", "3")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Errorf("run = %d, out=%q", code, out)
	}
	code, _, errb = runOut("{/sv_SE.person}", "--no-shipped-data")
	if code != 2 || !strings.Contains(errb, "--no-shipped-data needs at least one --data-path") {
		t.Errorf("--no-shipped-data alone = %d, %q, want misuse naming --data-path", code, errb)
	}
}

func TestRunRepeatIsBounded(t *testing.T) {
	for _, args := range [][]string{
		{"--repeat", "9223372036854775807"},
		{"--repeat", "1048577"},
		{"-n", "99999999999"},
	} {
		code, out, errb := runOut("{/sv_SE.word}", args...)
		if code != 2 || out != "" || !strings.Contains(errb, "1..1048576") {
			t.Errorf("run(%v) = %d, %q, %q, want misuse naming the range", args, code, out, errb)
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(`"x"`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errb := runOut("{/x}", "--no-shipped-data", "-d", dir, "--repeat", "1048576", "--separator", "")
	if code != 0 || len(out) != 1048576 {
		t.Errorf("repeat at the cap = %d, %d bytes, stderr %q; want every render streamed", code, len(out), errb)
	}
}

func TestRunListTakesNoRepeatOrSeparator(t *testing.T) {
	for _, args := range [][]string{{"--list", "-n", "3"}, {"--list", "--separator", ","}} {
		code, _, errb := runOut("", args...)
		if code != 2 || !strings.Contains(errb, "--list takes no") {
			t.Errorf("run(%v) = %d, %q, want misuse", args, code, errb)
		}
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

func recordDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	content := `{"format":"{first} {last}","first":["Ada","Bo"],"last":["Lovelace","Ek"]}`
	if err := os.WriteFile(filepath.Join(dir, "users.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRunRecordJSON(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "json", "--repeat", "2", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	var arr []map[string]string
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		t.Fatalf("json output is not one JSON array: %v\n%q", err, out)
	}
	if len(arr) != 2 || len(arr[0]) != 2 || arr[0]["first"] == "" || arr[0]["last"] == "" {
		t.Fatalf("json output = %q, want an array of two records with first and last columns", out)
	}
}

func TestRunRecordNDJSON(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "ndjson", "--repeat", "2", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("ndjson output = %d lines %q, want one object per line", len(lines), out)
	}
	for _, line := range lines {
		var m map[string]string
		if err := json.Unmarshal([]byte(line), &m); err != nil || len(m) != 2 {
			t.Fatalf("ndjson line %q is not one object: %v", line, err)
		}
	}
}

func TestRunRecordCSVRoundTrips(t *testing.T) {
	code, out, errb := runOut("{/users}\r\n", "--seed", "1", "--format", "csv", "--repeat", "3", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("csv output = %d lines %q, want a header and 3 rows", len(lines), out)
	}
	if lines[0] != "first,last" {
		t.Fatalf("csv header = %q, want first,last", lines[0])
	}
	r, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("csv output is not valid CSV: %v", err)
	}
	if len(r) != 4 || len(r[0]) != 2 {
		t.Fatalf("csv parsed to %v, want 4 records of 2 fields", r)
	}
}

func TestRunRecordSQL(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "sql", "--repeat", "2", "--table", "people", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], `INSERT INTO "people" ("first", "last") VALUES (`) {
		t.Fatalf("sql output = %q, want two INSERT INTO people statements", out)
	}
}

func TestRunRecordDefaultTable(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "sql", "--data-path", recordDir(t))
	if code != 0 || !strings.HasPrefix(out, `INSERT INTO "users" (`) {
		t.Fatalf("default table = %d, %q, want the reference's last segment; stderr %q", code, out, errb)
	}
}

func TestRunRecordOfAQuotedLoneReference(t *testing.T) {
	code, out, errb := runOut(`"{/users}"`, "--seed", "1", "--format", "sql", "--data-path", recordDir(t))
	if code != 0 || !strings.HasPrefix(out, `INSERT INTO "users" (`) {
		t.Fatalf("a JSON string holding {/users} = %d, %q, %q; want the record users", code, out, errb)
	}
	code, out, errb = runOut("{/users} ", "--format", "csv", "--data-path", recordDir(t))
	if code != 2 || out != "" || !strings.Contains(errb, "whitespace around {/users}") {
		t.Errorf("{/users} and a space = %d, %q, %q; want misuse naming the whitespace", code, out, errb)
	}
}

func TestRunRecordInlineTemplate(t *testing.T) {
	code, out, errb := runOut(`{"format":"{x}","x":["a","b"]}`, "--seed", "1", "--format", "json")
	if code != 0 {
		t.Fatalf("inline record = %d, stderr=%q", code, errb)
	}
	var arr []map[string]string
	if err := json.Unmarshal([]byte(out), &arr); err != nil || len(arr) != 1 || (arr[0]["x"] != "a" && arr[0]["x"] != "b") {
		t.Fatalf("inline record json = %q, want one record with a column x (err %v)", out, err)
	}
}

func TestRunRecordMisuse(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--format", "yaml"}, "--format takes csv, json, ndjson, sql or text"},
		{[]string{"--format", "json", "--separator", ","}, "--separator joins text values"},
		{[]string{"--table", "t"}, "--table names the INSERT target"},
		{[]string{"--format", "json", "--table", "t"}, "--table names the INSERT target"},
		{[]string{"--format", "sql", "--table", ""}, "--table names the INSERT target"},
	} {
		code, out, errb := runOut("{/users}", c.args...)
		if code != 2 || out != "" || !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) = %d, %q, %q; want misuse naming %q", c.args, code, out, errb, c.want)
		}
	}
}

func TestRunRecordOnABareValueIsRuntimeError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "word.json"), []byte(`["a","b"]`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errb := runOut("{/word}", "--format", "json", "--data-path", dir)
	if code != 1 {
		t.Fatalf("record on a choice = %d, want exit 1 (runtime error), stderr %q", code, errb)
	}
}

func TestRunRecordRejectsATopLevelRepeat(t *testing.T) {
	dir := t.TempDir()
	content := `{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`
	if err := os.WriteFile(filepath.Join(dir, "rep.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errb := runOut("{/rep}", "--format", "json", "--data-path", dir)
	if code != 1 || !strings.Contains(errb, "carries repeat 3") {
		t.Errorf("record on a repeating path = %d, %q; want exit 1 naming the repeat", code, errb)
	}
	code, _, errb = runOut(content, "--format", "json")
	if code != 2 || !strings.Contains(errb, "carries repeat 3") {
		t.Errorf("record on a repeating inline template = %d, %q; want exit 2 naming the repeat", code, errb)
	}
	if code, out, _ := runOut("{/rep}", "--seed", "1", "--data-path", dir); code != 0 || strings.TrimSpace(out) != "x-|x-|x-" {
		t.Errorf("text view = %d, %q; want the repeat still composed", code, out)
	}
}

func TestRunNumbersFollowTheShell(t *testing.T) {
	_, want, _ := runOut("{/sv_SE.word}", "--seed", "7", "--repeat", "3")
	code, got, errb := runOut("{/sv_SE.word}", "--seed", "007", "--repeat", "+3")
	if code != 0 || got != want {
		t.Errorf("run(--seed 007 --repeat +3) = %d, %q, stderr %q; want the same as --seed 7 --repeat 3 %q", code, got, errb, want)
	}
}

// TestRunRecordOfAFieldlessPath pins the way out of asking a category of one value
// for columns: the record to write, and the quoting a shell needs to take it. The
// template is well-formed and only the data cannot serve it, so the code is 1.
func TestRunRecordOfAFieldlessPath(t *testing.T) {
	code, out, errb := runOut("{/sv_SE.personnummer}", "--format", "csv")
	if code != 1 {
		t.Fatalf("run(--format csv) of {/sv_SE.personnummer} = %d, want 1", code)
	}
	for _, want := range []string{`{"format":"","personnummer":"{/sv_SE.personnummer}"}`, "single quotes"} {
		if !strings.Contains(errb, want) {
			t.Errorf("stderr = %q, want it to name %s", errb, want)
		}
	}
	if out != "" {
		t.Errorf("stdout = %q, want nothing", out)
	}
}
