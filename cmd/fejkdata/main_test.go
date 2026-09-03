package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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

func TestRunShortFlagValues(t *testing.T) {
	_, want, _ := runOut("--seed", "42", "--data-path", svSE, "address")
	for _, args := range [][]string{
		{"-s42", "-d", svSE, "address"},
		{"-s", "42", "-d" + svSE, "address"},
		{"-d", svSE, "address", "-s42"},
	} {
		code, got, errb := runOut(args...)
		if code != 0 {
			t.Fatalf("run(%v) = %d, stderr=%q", args, code, errb)
		}
		if got != want {
			t.Errorf("run(%v) = %q, want %q", args, got, want)
		}
	}
	_, three, _ := runOut("-s", "1", "-n3", "-d", svSE, "word")
	if lines := strings.Split(strings.TrimRight(three, "\n"), "\n"); len(lines) != 3 {
		t.Errorf("-n3 gave %d lines: %q", len(lines), three)
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-s=42", "-d", svSE, "address"}, "-s42"},
		{[]string{"-s=42", "-d", svSE, "address"}, "--seed=42"},
		{[]string{"-d=" + svSE, "address"}, "--data-path="},
		{[]string{"-nd", "3", "-d", svSE, "word"}, `--repeat needs an integer in 1..1048576, got "d"`},
		{[]string{"--seed=", "-d", svSE, "word"}, `--seed needs an unsigned integer, got ""`},
	} {
		code, out, errb := runOut(c.args...)
		if code != 2 || out != "" {
			t.Errorf("run(%v) = %d, stdout %q, want misuse", c.args, code, out)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("run(%v) stderr = %q, want %q", c.args, errb, c.want)
		}
	}
	code, out, _ := runOut("-hd", svSE)
	if code != 0 || !strings.Contains(out, "Usage") {
		t.Errorf("-hd (bundled help) = %d, %q, want usage", code, out)
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
	for _, args := range [][]string{{}, {"--data-path", svSE}, {"-d", svSE, "person", "word"}, {"-d", svSE, "--list", "person"}, {"--no-shipped-data", "sv_SE.person"}} {
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

func TestRunShippedDataByDefault(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "sv_SE.person")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Fatalf("run = %d, out=%q, stderr=%q", code, out, errb)
	}
	code, list, _ := runOut("--list")
	if code != 0 || !strings.Contains(list, "en_US.person\n") || !strings.Contains(list, "misc.uuid\n") {
		t.Errorf("--list without --data-path = %d, %q", code, list)
	}
	code, _, errb = runOut("person")
	if code != 1 || !strings.Contains(errb, "person") {
		t.Errorf("a category outside the shipped tree: code %d, stderr %q", code, errb)
	}
}

func TestClassify(t *testing.T) {
	for arg, want := range map[string]argKind{
		"sv_SE.person":   argPath,
		"person.last":    argPath,
		"name: {x}":      argTemplate, // a { token: a path can never carry a brace
		`{"format":"x"}`: argTemplate,
		`["a","b"]`:      argTemplate, // a JSON array carries no brace
		`[1, 2]`:         argTemplate,
		` ["a","b"]`:     argTemplate, // padding is the template's own error, not a shape verdict
		`"hello"`:        argTemplate, // a JSON string, the spelling a format-only object names
	} {
		got, err := classify(arg)
		if err != nil || got != want {
			t.Errorf("classify(%q) = %v, %v; want %v", arg, got, err, want)
		}
	}
	for arg, want := range map[string]string{
		"[abc]":       `holds a "["`,
		"[abc].field": `holds a "["`,
		"x[1]":        `holds a "["`,
		"a]b":         `holds a "]"`,
		"a}b":         `holds a "}"`,
		`"abc`:        `holds a "\""`,
		`"a]b`:        `holds a "\""`, // the opener the reader typed, not the bracket behind it
	} {
		_, err := classify(arg)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("classify(%q) = %v; want it rejected naming %s", arg, err, want)
		}
	}
}

func TestUsageReferencesResolve(t *testing.T) {
	for _, token := range regexp.MustCompile(`\{/[^}]+\}`).FindAllString(usage, -1) {
		code, out, errb := runOut("--seed", "1", token)
		if code != 0 || strings.TrimSpace(out) == "" {
			t.Errorf("usage advertises %s: run = %d, %q, stderr %q", token, code, out, errb)
		}
	}
}

func TestRunShapeMisuseBeforeLoad(t *testing.T) {
	code, _, errb := runOut("--no-shipped-data", "[abc]")
	if code != 2 || !strings.Contains(errb, "[abc]") || strings.Contains(errb, "--data-path") {
		t.Fatalf("shape misuse with no data = %d, %q; want the shape error before any load", code, errb)
	}
}

func TestRunInlineTemplate(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "name: {/sv_SE.person.last}")
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline format string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("--seed", "1", `{"format":"name: {x}","x":["bosse","lina"]}`)
	if code != 0 || (out != "name: bosse\n" && out != "name: lina\n") {
		t.Fatalf("inline JSON template = %d, %q, want one name, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("--seed", "1", `"name: {/sv_SE.person.last}"`)
	if code != 0 || !strings.HasPrefix(out, "name: ") || strings.Contains(out, "{") {
		t.Fatalf("inline JSON string = %d, %q, stderr %q", code, out, errb)
	}
	code, out, errb = runOut("--seed", "1", "-n", "2", `{digits(1)}`)
	if code != 0 || len(strings.Split(strings.TrimRight(out, "\n"), "\n")) != 2 {
		t.Fatalf("inline template with --repeat = %d, %q, stderr %q", code, out, errb)
	}
}

func TestRunTemplateMisuse(t *testing.T) {
	for arg, want := range map[string]string{
		"{bad":            "unterminated",
		"[red,green]":     "names no template either",
		`{"format":"x"}`:  "is a string",
		"{/no.such.path}": "no entry",
		"x[1]":            "names no template either",
		` ["a","b"] `:     "may not be padded",
	} {
		code, out, errb := runOut("--seed", "1", arg)
		if code != 2 || out != "" || !strings.Contains(errb, "try 'fejkdata --help'") || !strings.Contains(errb, want) {
			t.Errorf("run(%q) = %d, %q, %q; want misuse naming %q and --help", arg, code, out, errb, want)
		}
		if strings.Contains(errb, "fejkdata: fejkdata:") {
			t.Errorf("run(%q) doubled the program prefix: %q", arg, errb)
		}
	}
}

func TestRunNoShippedData(t *testing.T) {
	code, list, errb := runOut("--no-shipped-data", "-d", svSE, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if strings.Contains(list, "en_US") || !strings.Contains(list, "person\n") {
		t.Errorf("--no-shipped-data --list = %q, want only the given dir", list)
	}
	code, out, _ := runOut("--no-shipped-data", "-d", svSE, "-s", "3", "person")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Errorf("run = %d, out=%q", code, out)
	}
	code, _, errb = runOut("--no-shipped-data", "sv_SE.person")
	if code != 2 || !strings.Contains(errb, "--no-shipped-data needs at least one --data-path") {
		t.Errorf("--no-shipped-data alone = %d, %q, want misuse naming --data-path", code, errb)
	}
}

func TestRunRepeatIsBounded(t *testing.T) {
	for _, args := range [][]string{
		{"--repeat", "9223372036854775807", "sv_SE.word"},
		{"--repeat", "1048577", "sv_SE.word"},
		{"-n", "99999999999", "sv_SE.word"},
	} {
		code, out, errb := runOut(args...)
		if code != 2 || out != "" || !strings.Contains(errb, "1..1048576") {
			t.Errorf("run(%v) = %d, %q, %q, want misuse naming the range", args, code, out, errb)
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(`"x"`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errb := runOut("--no-shipped-data", "-d", dir, "--repeat", "1048576", "--separator", "", "x")
	if code != 0 || len(out) != 1048576+1 {
		t.Errorf("repeat at the cap = %d, %d bytes, stderr %q; want every render streamed", code, len(out), errb)
	}
}

func TestRunListTakesNoRepeatOrSeparator(t *testing.T) {
	for _, args := range [][]string{{"--list", "-n", "3"}, {"--list", "--separator", ","}} {
		code, _, errb := runOut(args...)
		if code != 2 || !strings.Contains(errb, "--list takes no") {
			t.Errorf("run(%v) = %d, %q, want misuse", args, code, errb)
		}
	}
}

func TestRunUnknownFlagIsNamedByRune(t *testing.T) {
	code, _, errb := runOut("-ä", "sv_SE.word")
	if code != 2 || !strings.Contains(errb, "unknown flag -ä") {
		t.Errorf("run(-ä) = %d, %q, want the flag named whole", code, errb)
	}
}

func TestRunEmptyDataPathIsNamed(t *testing.T) {
	code, _, errb := runOut("--data-path=", "sv_SE.word")
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
	code, out, errb := runOut("--seed", "1", "--format", "json", "--data-path", recordDir(t), "users")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("json output is not one valid JSON object: %v\n%q", err, out)
	}
	if len(m) != 2 || m["first"] == "" || m["last"] == "" {
		t.Fatalf("json output = %q, want first and last columns", out)
	}
}

func TestRunRecordCSVRoundTrips(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "--format", "csv", "--repeat", "3", "--data-path", recordDir(t), "users")
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
	code, out, errb := runOut("--seed", "1", "--format", "sql", "--repeat", "2", "--table", "people", "--data-path", recordDir(t), "users")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], `INSERT INTO "people" ("first", "last") VALUES (`) {
		t.Fatalf("sql output = %q, want two INSERT INTO people statements", out)
	}
}

func TestRunRecordDefaultTable(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "--format", "sql", "--data-path", recordDir(t), "users")
	if code != 0 || !strings.HasPrefix(out, `INSERT INTO "users" (`) {
		t.Fatalf("default table = %d, %q, want the path's last segment; stderr %q", code, out, errb)
	}
}

func TestRunRecordInlineTemplate(t *testing.T) {
	code, out, errb := runOut("--seed", "1", "--format", "json", `{"format":"{x}","x":["a","b"]}`)
	if code != 0 {
		t.Fatalf("inline record = %d, stderr=%q", code, errb)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil || m["x"] != "a" && m["x"] != "b" {
		t.Fatalf("inline record json = %q, want a column x (err %v)", out, err)
	}
}

func TestRunRecordMisuse(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--format", "yaml", "users"}, "--format takes text, json, csv or sql"},
		{[]string{"--format", "json", "--separator", ",", "users"}, "--separator joins text values"},
		{[]string{"--table", "t", "users"}, "--table names the INSERT target"},
		{[]string{"--format", "json", "--table", "t", "users"}, "--table names the INSERT target"},
	} {
		code, out, errb := runOut(c.args...)
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
	code, _, errb := runOut("--format", "json", "--data-path", dir, "word")
	if code != 1 {
		t.Fatalf("record on a choice = %d, want exit 1 (runtime error), stderr %q", code, errb)
	}
}

func TestRunNumbersFollowTheShell(t *testing.T) {
	_, want, _ := runOut("--seed", "7", "--repeat", "3", "sv_SE.word")
	code, got, errb := runOut("--seed", "007", "--repeat", "+3", "sv_SE.word")
	if code != 0 || got != want {
		t.Errorf("run(--seed 007 --repeat +3) = %d, %q, stderr %q; want the same as --seed 7 --repeat 3 %q", code, got, errb, want)
	}
}
