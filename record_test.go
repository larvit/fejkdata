package fejkdata

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func recordCat(t *testing.T) *Generator {
	t.Helper()
	dir := writeData(t, map[string]string{
		"users": `{
			"format": "{first} {last}",
			"first": ["Ada", "Bo"],
			"last": ["Lovelace", "Ek"]
		}`,
	})
	return newGenerator(t, dir, WithSeed(1))
}

func TestRecordProjectsFieldsAsColumns(t *testing.T) {
	f := recordCat(t)
	r, err := f.Record("users")
	if err != nil {
		t.Fatal(err)
	}
	got := r.Fields()
	if len(got) != 2 || got[0].Name != "first" || got[1].Name != "last" {
		t.Fatalf("Record.Fields() = %v, want columns first, last in name order", got)
	}
	if (got[0].Value != "Ada" && got[0].Value != "Bo") || (got[1].Value != "Lovelace" && got[1].Value != "Ek") {
		t.Fatalf("columns = %v, want the field values", got)
	}
}

func TestRecordIsDeterministic(t *testing.T) {
	a, b := recordCat(t), recordCat(t)
	for i := 0; i < 20; i++ {
		x, _ := a.Record("users")
		y, _ := b.Record("users")
		if x.JSON() != y.JSON() {
			t.Fatalf("same seed diverged: %s != %s", x.JSON(), y.JSON())
		}
	}
}

func TestRecordSkipsReferenceBindings(t *testing.T) {
	dir := writeData(t, map[string]string{
		"name": `{"format": "{first} {last}", "first": "Ada", "last": "Lovelace"}`,
		"user": `{"format": "they are {/name}", "id": "1"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.Record("user")
	if err != nil {
		t.Fatal(err)
	}
	got := r.Fields()
	if len(got) != 1 || got[0].Name != "id" {
		t.Fatalf("Record.Fields() = %v, want only the id column (a {/path} is not a column)", got)
	}
}

func TestRecordErrors(t *testing.T) {
	dir := writeData(t, map[string]string{
		"bare":      `"hello"`,
		"pick":      `["a", "b"]`,
		"group/cat": `"x"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for _, c := range []struct {
		path string
		want string
	}{
		{"bare", "has no fields, so no columns"},
		{"pick", "names a choice"},
		{"group", "names a folder"},
		{"nope", "no entry"},
	} {
		if _, err := f.Record(c.path); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Record(%q) = %v, want an error containing %q", c.path, err, c.want)
		}
	}
}

func TestRecordJSON(t *testing.T) {
	f := recordCat(t)
	r, err := f.Record("users")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(r.JSON()), &m); err != nil {
		t.Fatalf("Record.JSON() is not valid JSON: %v\n%s", err, r.JSON())
	}
	if len(m) != 2 || (m["first"] != "Ada" && m["first"] != "Bo") {
		t.Fatalf("Record.JSON() = %s, want two addressable columns", r.JSON())
	}
	for _, f := range r.Fields() {
		if m[f.Name] != f.Value {
			t.Fatalf("JSON column %q = %q, want %q", f.Name, m[f.Name], f.Value)
		}
	}
}

func TestRecordCSV(t *testing.T) {
	dir := writeData(t, map[string]string{
		"note": `{"format": "{word}, {word}", "word": ["a", "b,c", "d\"e", "f\n"]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.Record("note")
	if err != nil {
		t.Fatal(err)
	}
	if got := r.CSVHeader(); got != "word" {
		t.Errorf("CSVHeader() = %q, want word", got)
	}
	rec, err := csv.NewReader(strings.NewReader(r.CSVLine() + "\n")).Read()
	if err != nil {
		t.Fatalf("CSVLine() is not valid CSV: %v\n%q", err, r.CSVLine())
	}
	if len(rec) != 1 || rec[0] != r.Fields()[0].Value {
		t.Fatalf("CSVLine() = %v, want the field value round-tripped", rec)
	}
}

func TestRecordSQLInsert(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format": "{last}", "last": "O'Brien"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.Record("person")
	if err != nil {
		t.Fatal(err)
	}
	got := r.SQLInsert("people")
	if got != `INSERT INTO "people" ("last") VALUES ('O''Brien');` {
		t.Fatalf("SQLInsert() = %q, want quoted identifiers and the single quote doubled", got)
	}
}

func TestRecordSharesAReferenceAcrossColumns(t *testing.T) {
	dir := writeData(t, map[string]string{
		"currency": `[{"format":"{code}","code":"AUD","symbol":"$"},{"format":"{code}","code":"EUR","symbol":"€"}]`,
		"price":    `{"format":"","code":"{/currency.code}","symbol":"{/currency.symbol}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 100; i++ {
		r, err := f.Record("price")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range r.Fields() {
			m[c.Name] = c.Value
		}
		switch m["code"] {
		case "AUD":
			if m["symbol"] != "$" {
				t.Fatalf("record %q: code AUD but symbol %q, want one currency draw across columns", r.JSON(), m["symbol"])
			}
		case "EUR":
			if m["symbol"] != "€" {
				t.Fatalf("record %q: code EUR but symbol %q, want one currency draw across columns", r.JSON(), m["symbol"])
			}
		default:
			t.Fatalf("record %q has unexpected code %q", r.JSON(), m["code"])
		}
	}
}

func TestRecordSQLQuotesIdentifiers(t *testing.T) {
	dir := writeData(t, map[string]string{
		"row": `{"format": "", "postal-code": "1", "street-number": "2"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.Record("row")
	if err != nil {
		t.Fatal(err)
	}
	got := r.SQLInsert("my-table")
	if got != `INSERT INTO "my-table" ("postal-code", "street-number") VALUES ('1', '2');` {
		t.Fatalf("SQLInsert() = %q, want hyphenated identifiers and table quoted", got)
	}
}

func TestFakeRecordAndTemplate(t *testing.T) {
	f := recordCat(t)
	in := `{"format":"{x} {y}","x":["1","2"],"y":["3","4"]}`
	want, err := f.FakeRecord(in)
	if err != nil {
		t.Fatalf("FakeRecord: %v", err)
	}
	if len(want.Fields()) != 2 {
		t.Fatalf("FakeRecord fields = %v, want two columns", want.Fields())
	}
	reusable, err := f.NewRecordTemplate(in)
	if err != nil {
		t.Fatalf("NewRecordTemplate: %v", err)
	}
	for i := 0; i < 20; i++ {
		if got := reusable.Fake(); len(got.Fields()) != 2 {
			t.Fatalf("RecordTemplate.Fake() = %v, want two columns", got.Fields())
		}
	}
}

func TestInlineRecordErrors(t *testing.T) {
	f := recordCat(t)
	for _, c := range []struct {
		input string
		want  string
	}{
		{`"hello"`, "needs at least one field"},
		{`["a","b"]`, "a choice"},
	} {
		if _, err := f.FakeRecord(c.input); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeRecord(%q) = %v, want an error naming %q", c.input, err, c.want)
		}
	}
}
