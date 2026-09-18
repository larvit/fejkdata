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
	misc = "../../data/misc"
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
	// A template a shell split names both quotings, since which one cures it turns
	// on whether the argument carries a double-quoted JSON.
	_, _, errb := runOut("{date(1990-01-01,2010-12-31,", "January", "2,", "2006)}")
	for _, want := range []string{`wrap the whole argument in "…"`, `or in '…'`} {
		if !strings.Contains(errb, want) {
			t.Errorf("a split template stderr = %q, want it to name %s", errb, want)
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
	for arg, want := range map[string]argKind{"sv_SE.person": argPath, "name: {x}": argTemplate} {
		if got, err := classify(arg); err != nil || got != want {
			t.Errorf("classify(%q) = %v, %v; want %v", arg, got, err, want)
		}
	}
	if _, err := classify("[abc]"); err == nil || !strings.Contains(err.Error(), `"["`) || !strings.Contains(err.Error(), "JSON") {
		t.Errorf("classify([abc]) = %v; want it rejected naming the leading bracket", err)
	}
	if got, err := classify("geo.SE.municipality[St. Louis].name"); err != nil || got != argPath {
		t.Errorf("classify(a selecting path) = %v, %v; want a path", got, err)
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
	if code, out, errb := runOut("--no-shipped-data", "-d", dir, "region[12]"); code != 0 || out != "Skåne län\n" {
		t.Fatalf("region[12] = %d, %q, stderr=%q", code, out, errb)
	}
	if code, out, _ := runOut("--no-shipped-data", "-d", dir, "region[Skåne län].code"); code != 0 || out != "12\n" {
		t.Fatalf("region[Skåne län].code = %d, %q", code, out)
	}
	if code, out, _ := runOut("--no-shipped-data", "-d", dir, "--format", "sql", "region[12]"); code != 0 || out != `INSERT INTO "region" ("code", "name") VALUES ('12', 'Skåne län');`+"\n" {
		t.Fatalf("--format sql region[12] = %d, %q, want the table named without its selector", code, out)
	}
	if code, _, errb := runOut("--no-shipped-data", "-d", dir, "region[99]"); code != 1 || !strings.Contains(errb, `"99"`) {
		t.Fatalf("region[99] = %d, stderr=%q, want a runtime error naming the row", code, errb)
	}
	if code, _, errb := runOut("--no-shipped-data", "-d", dir, "[Skåne län]"); code != 2 || !strings.Contains(errb, "JSON") {
		t.Fatalf("[Skåne län] = %d, stderr=%q, want misuse: a leading bracket that is no JSON names nothing", code, errb)
	}
	if code, out, _ := runOut("--seed", "1", "--format", "csv", "misc.country[SE]"); code != 0 || !strings.HasPrefix(out, "alpha2,") || !strings.Contains(out, "\nSE,SWE,") {
		t.Fatalf("--format csv misc.country[SE] = %d, %q", code, out)
	}
}

func TestUsageReferencesResolve(t *testing.T) {
	for _, token := range regexp.MustCompile(`\{/[^}]+\}`).FindAllString(usage, -1) {
		path := token[2 : len(token)-1]
		code, out, errb := runOut("--seed", "1", path)
		if code != 0 || strings.TrimSpace(out) == "" {
			t.Errorf("usage advertises %s: run %s = %d, %q, stderr %q", token, path, code, out, errb)
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
		"{bad":                  "unterminated",
		"[red,green]":           `starts with "["`,
		`{"format":"x"}`:        "is a string",
		"name: {/no.such.path}": "no entry",
		"{/sv_SE.person}":       "write sv_SE.person",
		"x[1]y":                 `"]"`,
		` ["a","b"] `:           "may not be padded",
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
	code, list, errb := runOut("--no-shipped-data", "-d", misc, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if strings.Contains(list, "sv_SE") || !strings.Contains(list, "uuid\n") {
		t.Errorf("--no-shipped-data --list = %q, want only the given dir", list)
	}
	code, out, _ := runOut("--no-shipped-data", "-d", misc, "-s", "3", "uuid")
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
	code, out, errb := runOut("--seed", "1", "--format", "json", "--repeat", "2", "--data-path", recordDir(t), "users")
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
	code, out, errb := runOut("--seed", "1", "--format", "ndjson", "--repeat", "2", "--data-path", recordDir(t), "users")
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
		{[]string{"--format", "yaml", "users"}, "--format takes csv, json, ndjson, sql or text"},
		{[]string{"--format", "json", "--separator", ",", "users"}, "--separator joins text values"},
		{[]string{"--table", "t", "users"}, "--table names the INSERT target"},
		{[]string{"--format", "json", "--table", "t", "users"}, "--table names the INSERT target"},
		{[]string{"--format", "sql", "--table", "", "users"}, "--table names the INSERT target"},
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

func TestRunRecordRejectsATopLevelRepeat(t *testing.T) {
	dir := t.TempDir()
	content := `{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`
	if err := os.WriteFile(filepath.Join(dir, "rep.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errb := runOut("--format", "json", "--data-path", dir, "rep")
	if code != 1 || !strings.Contains(errb, "carries repeat 3") {
		t.Errorf("record on a repeating path = %d, %q; want exit 1 naming the repeat", code, errb)
	}
	code, _, errb = runOut("--format", "json", content)
	if code != 2 || !strings.Contains(errb, "carries repeat 3") {
		t.Errorf("record on a repeating inline template = %d, %q; want exit 2 naming the repeat", code, errb)
	}
	if code, out, _ := runOut("--seed", "1", "--data-path", dir, "rep"); code != 0 || strings.TrimSpace(out) != "x-|x-|x-" {
		t.Errorf("text view = %d, %q; want the repeat still composed", code, out)
	}
}

func TestRunNumbersFollowTheShell(t *testing.T) {
	_, want, _ := runOut("--seed", "7", "--repeat", "3", "sv_SE.word")
	code, got, errb := runOut("--seed", "007", "--repeat", "+3", "sv_SE.word")
	if code != 0 || got != want {
		t.Errorf("run(--seed 007 --repeat +3) = %d, %q, stderr %q; want the same as --seed 7 --repeat 3 %q", code, got, errb, want)
	}
}

// TestRunRecordOfAFieldlessPath pins the way out of asking a category of one value
// for columns: the record to write, and the quoting a shell needs to take it. The
// argument is well-formed and only the data cannot serve it, so the code is 1.
func TestRunRecordOfAFieldlessPath(t *testing.T) {
	code, out, errb := runOut("--format", "csv", "sv_SE.personnummer")
	if code != 1 {
		t.Fatalf("run(--format csv sv_SE.personnummer) = %d, want 1", code)
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
