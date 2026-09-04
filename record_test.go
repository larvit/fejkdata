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
	got := r.Columns()
	if len(got) != 2 || got[0].Name != "first" || got[1].Name != "last" {
		t.Fatalf("Record.Columns() = %v, want columns first, last in name order", got)
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
	got := r.Columns()
	if len(got) != 1 || got[0].Name != "id" {
		t.Fatalf("Record.Columns() = %v, want only the id column (a {/path} is not a column)", got)
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
	for _, f := range r.Columns() {
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
	if len(rec) != 1 || rec[0] != r.Columns()[0].Value {
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

func TestRecordRejectsFieldDescent(t *testing.T) {
	dir := writeData(t, map[string]string{
		"cat": `{"format":"{sub}","sub":{"format":"{x}","x":"1"}}`,
		"row": `[{"format":"{x}","x":"1"},{"format":"{x}","x":"2"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for _, path := range []string{"cat.sub", "row.x"} {
		if _, err := f.Record(path); err == nil || !strings.Contains(err.Error(), "field") {
			t.Errorf("Record(%q) = %v, want a 'descends into a field' error", path, err)
		}
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
		for _, c := range r.Columns() {
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

func TestRecordSharesAReferenceIntoAColumnRepeat(t *testing.T) {
	dir := writeData(t, map[string]string{
		"currency": `[{"format":"{code}","code":"AUD"},{"format":"{code}","code":"EUR"}]`,
		"order":    `{"format":"","codes":{"format":"{/currency.code}","repeat":3,"separator":"-"}}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 50; i++ {
		r, err := f.Record("order")
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(r.Columns()[0].Value, "-")
		if len(parts) != 3 || parts[0] != parts[1] || parts[1] != parts[2] {
			t.Fatalf("codes column = %q, want one shared draw across its repeat", r.Columns()[0].Value)
		}
	}
}

func TestRecordBareReferenceStaysIndependent(t *testing.T) {
	dir := writeData(t, map[string]string{
		"currency": `[{"format":"{code}","code":"AUD"},{"format":"{code}","code":"EUR"}]`,
		"order":    `{"format":"","whole":"{/currency}","code":"{/currency.code}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	sawMismatch := false
	for i := 0; i < 100; i++ {
		r, err := f.Record("order")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range r.Columns() {
			m[c.Name] = c.Value
		}
		if m["whole"] != m["code"] {
			sawMismatch = true
			break
		}
	}
	if !sawMismatch {
		t.Fatal("a bare {/currency} column never disagreed with a tailed {/currency.code} column; a bare reference should draw independently")
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
	if len(want.Columns()) != 2 {
		t.Fatalf("FakeRecord columns = %v, want two columns", want.Columns())
	}
	reusable, err := f.NewRecordTemplate(in)
	if err != nil {
		t.Fatalf("NewRecordTemplate: %v", err)
	}
	for i := 0; i < 20; i++ {
		if got := reusable.Fake(); len(got.Columns()) != 2 {
			t.Fatalf("RecordTemplate.Fake() = %v, want two columns", got.Columns())
		}
	}
}

func TestInlineRecordErrors(t *testing.T) {
	f := recordCat(t)
	for _, c := range []struct {
		input string
		want  string
	}{
		{`"hello"`, "has no fields, so no columns"},
		{`["a","b"]`, "names a choice, not a template"},
		{`{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`, "carries repeat 3"},
	} {
		if _, err := f.FakeRecord(c.input); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeRecord(%q) = %v, want an error naming %q", c.input, err, c.want)
		}
	}
}

func TestRecordRejectsATopLevelRepeat(t *testing.T) {
	dir := writeData(t, map[string]string{
		"rep": `{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if _, err := f.Record("rep"); err == nil || !strings.Contains(err.Error(), "carries repeat 3") {
		t.Errorf("Record on a repeating template = %v, want an error naming the repeat", err)
	}
	if v, err := f.Fake("rep"); err != nil || v != "x-|x-|x-" {
		t.Errorf("Fake(rep) = %q, %v, want the repeat still composed for the string view", v, err)
	}
}

func TestRecordTemplateRejectsATopLevelRepeat(t *testing.T) {
	f := recordCat(t)
	in := `{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`
	if _, err := f.NewRecordTemplate(in); err == nil || !strings.Contains(err.Error(), "carries repeat 3") {
		t.Errorf("NewRecordTemplate on a repeating template = %v, want an error naming the repeat", err)
	}
	if _, err := f.NewTemplate(in); err != nil {
		t.Errorf("NewTemplate on the same input = %v, want the string view to still compile", err)
	}
}
