package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRunSQLUnknownCategoryFails(t *testing.T) {
	code, _, errb := runOut("{/nope}", "--data-path", svSE, "--format", "sql")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if !strings.Contains(errb, "nope") {
		t.Errorf("stderr %q should name the unknown category", errb)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestRunSQLWriteFailureFails(t *testing.T) {
	var errb bytes.Buffer
	code := run("fejkdata", shipped, []string{"--seed", "1", "--format", "sql", "--data-path", recordDir(t)}, system{1, testDay}, strings.NewReader("{/users}"), failingWriter{}, &errb)
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "disk full") {
		t.Errorf("stderr %q should name the write error", errb.String())
	}
}

func TestRunFormatSQLNamesASelectedRowsTableWithoutItsSelector(t *testing.T) {
	dir := regionDir(t)
	if code, out, _ := runOut("{/region[12]}", "--no-shipped-data", "-d", dir, "--format", "sql"); code != 0 || out != `INSERT INTO "region" ("code", "name") VALUES ('12', 'Skåne län');`+"\n" {
		t.Fatalf("--format sql {/region[12]} = %d, %q, want the table named without its selector", code, out)
	}
}

func TestRunRepeatOfZeroWritesNoRecord(t *testing.T) {
	dir := recordDir(t)
	for _, c := range []struct {
		args []string
		want string
	}{
		{nil, ""},
		{[]string{"--format", "json"}, "[]\n"},
		{[]string{"--format", "ndjson"}, ""},
		{[]string{"--format", "csv"}, "first,last\n"},
		{[]string{"--format", "pgcopy"}, ""},
		{[]string{"--format", "sql"}, ""},
		{[]string{"--format", "tsv"}, "first\tlast\n"},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			if code, out, errb := runOut("{/users}", append(c.args, "-n", "0", "-d", dir)...); code != 0 || out != c.want {
				t.Errorf("run(-n 0 %v) = %d, %q, %q, want %q", c.args, code, out, errb, c.want)
			}
		})
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

func eachNDJSONRecord(t *testing.T, out string, check func(record map[string]string)) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		var record map[string]string
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("ndjson line %q is not one object: %v", line, err)
		}
		check(record)
	}
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
	eachNDJSONRecord(t, out, func(record map[string]string) {
		if len(record) != 2 {
			t.Fatalf("ndjson record %v does not hold two columns", record)
		}
	})
}

func TestRunRecordTSVFeedsALaterRun(t *testing.T) {
	writeFile := func(dir, name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	first := t.TempDir()
	writeFile(first, "series.json", `{"format":"","id":"{seq()}","customer":["Ada","{{x}}"]}`)
	code, tsv, errb := runOut("{/series}", "--seed", "1", "--format", "tsv", "-n", "30", "-d", first)
	if code != 0 {
		t.Fatalf("tsv run = %d, stderr=%q", code, errb)
	}
	if !strings.HasPrefix(tsv, "customer\tid\n") || strings.Count(tsv, "\n") != 31 {
		t.Fatalf("tsv output = %q, want a header and 30 rows", tsv)
	}
	_, truth, _ := runOut("{/series}", "--seed", "1", "--format", "ndjson", "-n", "30", "-d", first)
	valid := map[string]bool{}
	eachNDJSONRecord(t, truth, func(m map[string]string) { valid[m["id"]+" "+m["customer"]] = true })

	second := t.TempDir()
	writeFile(second, "series.tsv", tsv)
	writeFile(second, "series.json", `{"format":"{id}","rows":"series.tsv"}`)
	writeFile(second, "orders.json", `{"format":"","series_id":"{/series as s}{s.id}","customer":"{s.customer}"}`)
	code, out, errb := runOut("{/orders}", "--seed", "2", "--format", "ndjson", "-n", "100", "-d", second)
	if code != 0 {
		t.Fatalf("second run = %d, stderr=%q", code, errb)
	}
	seen := map[string]bool{}
	eachNDJSONRecord(t, out, func(m map[string]string) {
		key := m["series_id"] + " " + m["customer"]
		if !valid[key] {
			t.Fatalf("order %q names no series the first run wrote", key)
		}
		seen[m["customer"]] = true
	})
	if !seen["{x}"] || !seen["Ada"] {
		t.Fatalf("customers drawn = %v, want both Ada and the literal {x} the first run wrote as {{x}}", seen)
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
	if !regexp.MustCompile(`^INSERT INTO "people" \("first", "last"\) VALUES \('[^']*', '[^']*'\),\('[^']*', '[^']*'\);\n$`).MatchString(out) {
		t.Fatalf("sql output = %q, want one INSERT INTO people of two rows", out)
	}
}

func TestRunRecordSQLStartsANewStatementEveryThousandRows(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "sql", "--repeat", "2001", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	statements := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(statements) != 3 {
		t.Fatalf("%d statements, want 3", len(statements))
	}
	for i, want := range []int{1000, 1000, 1} {
		t.Run(fmt.Sprintf("statement %d", i), func(t *testing.T) {
			if !strings.HasPrefix(statements[i], `INSERT INTO "users" ("first", "last") VALUES (`) || !strings.HasSuffix(statements[i], ");") {
				t.Errorf("statement %d = %q, want a whole INSERT", i, statements[i])
			}
			if got := strings.Count(statements[i], "),(") + 1; got != want {
				t.Errorf("statement %d holds %d rows, want %d", i, got, want)
			}
		})
	}
}

func TestRunRecordPGCopyIsOneCopyBlockFedFromStdin(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "pgcopy", "--repeat", "2", "--table", "people", "--data-path", recordDir(t))
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if !regexp.MustCompile(`^COPY "people" \("first", "last"\) FROM STDIN;\n[^\t\n]+\t[^\t\n]+\n[^\t\n]+\t[^\t\n]+\n\\\.\n$`).MatchString(out) {
		t.Fatalf("pgcopy output = %q, want a COPY line, two tab-separated rows and the \\. terminator", out)
	}
}

func TestRunRecordPGCopyDefaultsItsTableToTheReferencesLastSegment(t *testing.T) {
	code, out, errb := runOut("{/users}", "--seed", "1", "--format", "pgcopy", "--data-path", recordDir(t))
	if code != 0 || !strings.HasPrefix(out, `COPY "users" (`) {
		t.Fatalf("default table = %d, %q, want the reference's last segment; stderr %q", code, out, errb)
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
}

func TestRunRecordRefusesWhitespaceAroundALoneReference(t *testing.T) {
	code, out, errb := runOut("{/users} ", "--format", "csv", "--data-path", recordDir(t))
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
		{[]string{"--format", "yaml"}, "--format takes csv, json, ndjson, pgcopy, sql, text or tsv"},
		{[]string{"--format", "sql", "--table", ""}, "--table names the INSERT target"},
		{[]string{"--format", "pgcopy", "--table", ""}, "--table names the COPY target"},
		{[]string{"--format", "pgcopy", "--separator", ","}, "--separator joins text renders; --format pgcopy always separates fields with a tab"},
		{[]string{"--format", "csv", "--separator", ";"}, `--separator joins text renders; --format csv always separates fields with ","`},
		{[]string{"--format", "tsv", "--separator", ","}, "--separator joins text renders; --format tsv always separates fields with a tab"},
		{[]string{"--format", "json", "--separator", ","}, "--separator joins text renders"},
		{[]string{"--format", "ndjson", "--separator", ","}, "--separator joins text renders"},
		{[]string{"--format", "sql", "--separator", ","}, "--separator joins text renders"},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			code, out, errb := runOut("{/users}", c.args...)
			if code != 2 || out != "" || !strings.Contains(errb, c.want) {
				t.Errorf("run(%v) = %d, %q, %q; want misuse naming %q", c.args, code, out, errb, c.want)
			}
		})
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

// TestRunRecordOfAPathWithNoColumns pins the way out of asking a choice or a bare format
// for columns: the record to write, and the quoting a shell needs to take it. The
// template is well-formed and only the data cannot serve it, so the code is 1.
func TestRunRecordOfAPathWithNoColumns(t *testing.T) {
	for category, wrapper := range map[string]string{
		"sv_SE.personnummer": `{"format":"","personnummer":"{/sv_SE.personnummer}"}`,
		"sv_SE.time":         `{"format":"","time":"{/sv_SE.time}"}`,
	} {
		t.Run(category, func(t *testing.T) {
			code, out, errb := runOut("{/"+category+"}", "--format", "csv")
			if code != 1 {
				t.Fatalf("run(--format csv) of {/%s} = %d, want 1", category, code)
			}
			if !strings.Contains(errb, wrapper) {
				t.Errorf("{/%s}: stderr = %q, want it to name %s", category, errb, wrapper)
			}
			if !strings.Contains(errb, "single quotes") {
				t.Errorf("{/%s}: stderr = %q, want it to name %s", category, errb, "single quotes")
			}
			if out != "" {
				t.Errorf("{/%s}: stdout = %q, want nothing", category, out)
			}
		})
	}
}
