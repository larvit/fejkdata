package fejkdata

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
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
	r, err := f.FakeRecord("users")
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
		x, _ := a.FakeRecord("users")
		y, _ := b.FakeRecord("users")
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
	r, err := f.FakeRecord("user")
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
		{"pick", `names a choice, not a template; a record is a template whose fields are its columns; render it as a column of one: {"format":"","pick":"{/pick}"}`},
		{"group", "names a folder"},
		{"nope", "no entry"},
	} {
		if _, err := f.FakeRecord(c.path); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeRecord(%q) = %v, want an error containing %q", c.path, err, c.want)
		}
	}
}

func TestRecordJSON(t *testing.T) {
	f := recordCat(t)
	r, err := f.FakeRecord("users")
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
	seen := map[string]bool{}
	for i := 0; i < 200 && len(seen) < 4; i++ {
		r, err := f.FakeRecord("note")
		if err != nil {
			t.Fatal(err)
		}
		if got := r.CSVHeader(); got != "word" {
			t.Fatalf("CSVHeader() = %q, want word", got)
		}
		rec, err := csv.NewReader(strings.NewReader(r.CSVLine() + "\n")).Read()
		if err != nil {
			t.Fatalf("CSVLine() is not valid CSV: %v\n%q", err, r.CSVLine())
		}
		want := r.Columns()[0].Value
		if len(rec) != 1 || rec[0] != want {
			t.Fatalf("CSVLine() = %v, want the field value %q round-tripped", rec, want)
		}
		seen[want] = true
	}
	if len(seen) != 4 {
		t.Fatalf("round-tripped %d of the 4 values; the comma, quote and newline shapes must each survive", len(seen))
	}
	for value, want := range map[string]string{`\.`: `"\."`, " x": `" x"`, " x": "\" x\"", "a\rb": "\"a\rb\""} {
		body, _ := json.Marshal(map[string]string{"format": "", "v": value})
		r, err := newGenerator(t, writeData(t, map[string]string{"q": string(body)})).FakeRecord("q")
		if err != nil {
			t.Fatal(err)
		}
		if got := r.CSVLine(); got != want {
			t.Errorf("CSVLine() of %q = %q, want %q, quoted where encoding/csv quotes", value, got, want)
		}
	}
}

func TestRecordCSVEmptyValueStaysARow(t *testing.T) {
	dir := writeData(t, map[string]string{"blank": `{"format": "", "note": ""}`})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.FakeRecord("blank")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(r.CSVHeader() + "\n" + r.CSVLine() + "\n")).ReadAll()
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	if len(rows) != 2 || len(rows[1]) != 1 || rows[1][0] != "" {
		t.Fatalf("one empty column parsed to %v, want a header and one row of one empty field", rows)
	}
}

func TestRecordTSV(t *testing.T) {
	dir := writeData(t, map[string]string{
		"row": `{"format":"","age":{"format":"42","datatype":"integer"},"gone":null,"id":"{{x}}","name":"a\tb\r\nc","nick":""}`,
		"odd": `{"format":"","x\ny":"1","z\tw":"2"}`,
	})
	r, err := newGenerator(t, dir, WithSeed(1)).FakeRecord("row")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.TSVHeader(), "age\tgone\tid\tname\tnick"; got != want {
		t.Errorf("TSVHeader() = %q, want %q", got, want)
	}
	if got, want := r.TSVLine(), "42\t\t{{x}}\ta b  c\t"; got != want {
		t.Errorf("TSVLine() = %q, want %q: a null column and an empty one an empty cell, braces doubled, a tab, CR or LF a space", got, want)
	}
	odd, err := newGenerator(t, dir, WithSeed(1)).FakeRecord("odd")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := odd.TSVHeader(), "x y\tz w"; got != want {
		t.Errorf("TSVHeader() of names holding a line break and a tab = %q, want %q, one line of two cells", got, want)
	}
}

func TestRecordWritesTypedAndNullColumns(t *testing.T) {
	dir := writeData(t, map[string]string{
		"row": `{"format":"","age":{"format":"42","datatype":"integer"},"gone":null,"name":"O'Brien","nick":"","paid":{"format":"true","datatype":"boolean"},"price":{"format":"19.99","datatype":"number"}}`,
	})
	r, err := newGenerator(t, dir, WithSeed(1)).FakeRecord("row")
	if err != nil {
		t.Fatal(err)
	}
	want := []Column{
		{Name: "age", DataType: DataTypeInteger, Value: "42"},
		{Name: "gone", Null: true},
		{Name: "name", Value: "O'Brien"},
		{Name: "nick"},
		{Name: "paid", DataType: DataTypeBoolean, Value: "true"},
		{Name: "price", DataType: DataTypeNumber, Value: "19.99"},
	}
	if got := r.Columns(); !reflect.DeepEqual(got, want) {
		t.Errorf("Columns() = %+v, want %+v", got, want)
	}
	if got, want := r.JSON(), `{"age":42,"gone":null,"name":"O'Brien","nick":"","paid":true,"price":19.99}`; got != want {
		t.Errorf("JSON() = %s, want %s", got, want)
	}
	if got, want := r.SQLInsert("t"), `INSERT INTO "t" ("age", "gone", "name", "nick", "paid", "price") VALUES (42, NULL, 'O''Brien', '', true, 19.99);`; got != want {
		t.Errorf("SQLInsert() = %s, want %s", got, want)
	}
	if got, want := r.CSVLine(), `42,,O'Brien,"",true,19.99`; got != want {
		t.Errorf("CSVLine() = %s, want %s: null an unquoted empty field, an empty string quoted", got, want)
	}
	if got := DataTypeNumber.String(); got != "number" {
		t.Errorf("DataTypeNumber.String() = %q, want the data's spelling", got)
	}
	lone, err := newGenerator(t, writeData(t, map[string]string{"gone": `{"format":"","note":null}`})).FakeRecord("gone")
	if err != nil {
		t.Fatal(err)
	}
	if got := lone.CSVLine(); got != "" {
		t.Errorf("a lone null column wrote CSVLine() = %q, want the blank line PostgreSQL's COPY reads as null", got)
	}
}

func TestRecordColumnsReadingOneReferenceLoad(t *testing.T) {
	cat := `{"format":"{a}","a":[{"format":"A={b}","b":"1"},{"format":"A={b}","b":"2"}]}`
	mid := `"{/cat.a}"`
	for _, row := range []string{
		`{"format":"","whole":"{/cat.a}","inner":{"format":"{/cat.a.b}","repeat":2,"separator":"-"}}`,
		`{"format":"","whole":"{/cat.a}","inner":"{/cat.a.b}"}`,
		`{"format":"","whole":"{/cat.a}","inner":{"format":"{/cat.a.b} {x}","x":"1"}}`,
		`{"format":"","whole":"{/cat.a}","inner":[{"format":"{/cat.a.b} {x}","x":"1"},{"format":"{/cat.a.b}! {x}","x":"2"}]}`,
		`{"format":"","whole":"{uppercase(/cat.a)}","inner":"{/cat.a.b}"}`,
		`{"format":"","whole":"{/cat}","inner":"{/cat.a.b}"}`,
		`{"format":"","whole":"{/mid}","inner":"{/cat.a.b}"}`,
	} {
		f := newGenerator(t, writeData(t, map[string]string{"cat": cat, "mid": mid, "row": row}), WithSeed(1))
		if _, err := f.FakeRecord("row"); err != nil {
			t.Errorf("%s: FakeRecord = %v, want it accepted", row, err)
		}
	}
}

func TestRecordColumnReadingItsOwnRecordIsAFreshDraw(t *testing.T) {
	person := `{"format":"{first} {last}","first":["Ada","Bo"],"last":["Lovelace","Ek"],"other":"{/person.first}"}`
	f := newGenerator(t, writeData(t, map[string]string{"person": person}), WithSeed(1))
	apart := false
	for i := 0; i < 50 && !apart; i++ {
		r, err := f.FakeRecord("person")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()
		apart = c[0].Value != c[2].Value
	}
	if !apart {
		t.Error("other matched first in 50 records, want a draw of its own")
	}
}

func TestRecordBareReferenceStaysIndependentAsAnOperand(t *testing.T) {
	dir := writeData(t, map[string]string{
		"cur": `[{"format":"{code}","code":"aud"},{"format":"{code}","code":"eur","symbol":"€"}]`,
		"row": `{"format":"","up":"{uppercase(/cur)}","low":"{lowercase(/cur)}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	sawMismatch := false
	for i := 0; i < 200 && !sawMismatch; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range r.Columns() {
			m[c.Name] = c.Value
		}
		sawMismatch = !strings.EqualFold(m["up"], m["low"])
	}
	if !sawMismatch {
		t.Error("two bare-reference operand columns never disagreed; a bare reference draws on its own, as the plain spelling does")
	}
	for i := 0; i < 200; i++ {
		v, err := f.FakeTemplate("{/cur}|{uppercase(/cur)}")
		if err != nil {
			t.Fatal(err)
		}
		if parts := strings.Split(v, "|"); !strings.EqualFold(parts[0], parts[1]) {
			return
		}
	}
	t.Error("{/cur}|{uppercase(/cur)} never disagreed in 200 renders; each reference draws on its own")
}

func TestRecordSQLInsert(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format": "{last}", "last": "O'Brien"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.FakeRecord("person")
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
		"row": `[{"format":"{x}","x":"1"},{"format":"{x}","x":"2","y":"b"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for _, path := range []string{"cat.sub", "row.x"} {
		if _, err := f.FakeRecord(path); err == nil || !strings.Contains(err.Error(), "field") {
			t.Errorf("FakeRecord(%q) = %v, want a 'descends into a field' error", path, err)
		}
	}
}

func TestRecordColumnsReadOneName(t *testing.T) {
	dir := writeData(t, map[string]string{
		"currency": `[{"format":"{code}","code":"AUD","symbol":"$"},{"format":"{code}","code":"EUR","symbol":"€","name":"Euro"}]`,
		"price":    `{"format":"{code} {symbol}","code":"{/currency as c}{c.code}","symbol":"{c.symbol}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	symbols := map[string]string{"AUD": "$", "EUR": "€"}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("price")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range r.Columns() {
			m[c.Name] = c.Value
		}
		if symbol, known := symbols[m["code"]]; !known || m["symbol"] != symbol {
			t.Fatalf("record %s, want one currency draw across columns", r.JSON())
		}
		v := fake(t, f, "price")
		if code, symbol, _ := strings.Cut(v, " "); symbol == "" || symbols[code] != symbol {
			t.Fatalf("Fake(price) = %q, want its fields one currency draw, as the record's columns are", v)
		}
	}
}

func TestRepeatIterationsPickNamesAnew(t *testing.T) {
	dir := writeData(t, map[string]string{
		"party":  `{"format":"{host}: {guests}","guests":{"format":"{/person as p}{p.first} {p.last}","repeat":3,"separator":", "},"host":"{/person as h}{h.first} {h.last}"}`,
		"person": drawPeople,
	})
	f := newGenerator(t, dir, WithSeed(1))
	differed := map[[2]int]bool{}
	check := func(view string, names []string) {
		t.Helper()
		if len(names) != 4 {
			t.Fatalf("%s rendered %q, want a host and three guests", view, names)
		}
		for i, name := range names {
			if !onePerson(name) {
				t.Fatalf("%s rendered %q, want each name one person", view, names)
			}
			for j := range names[:i] {
				differed[[2]int{j, i}] = differed[[2]int{j, i}] || names[j] != name
			}
		}
	}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("party")
		if err != nil {
			t.Fatal(err)
		}
		guests, host := r.Columns()[0].Value, r.Columns()[1].Value
		check("FakeRecord", append([]string{host}, strings.Split(guests, ", ")...))
		host, guests, _ = strings.Cut(fake(t, f, "party"), ": ")
		check("Fake", append([]string{host}, strings.Split(guests, ", ")...))
	}
	for pair, ok := range differed {
		if !ok {
			t.Errorf("names %v never differed in 200 renders; the host and each repeat iteration are a draw of their own, so four people", pair)
		}
	}
}

func TestRecordNamesDrawApart(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person":   drawPeople,
		"transfer": `{"format":"{from_first} {from_last} to {to_first} {to_last}","from_first":"{/person as from}{from.first}","from_last":"{from.last}","to_first":"{/person as to}{to.first}","to_last":"{to.last}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	apart := map[string]bool{}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("transfer")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, c := range r.Columns() {
			m[c.Name] = c.Value
		}
		from, to, _ := strings.Cut(fake(t, f, "transfer"), " to ")
		for view, pair := range map[string][2]string{"FakeRecord": {m["from_first"] + " " + m["from_last"], m["to_first"] + " " + m["to_last"]}, "Fake": {from, to}} {
			if !onePerson(pair[0]) || !onePerson(pair[1]) {
				t.Fatalf("%s drew %q and %q, want each name one person", view, pair[0], pair[1])
			}
			apart[view] = apart[view] || pair[0] != pair[1]
		}
	}
	for _, view := range []string{"FakeRecord", "Fake"} {
		if !apart[view] {
			t.Errorf("%s: names from and to drew one person in 100 renders, want a pick each", view)
		}
	}
}

func TestRecordColumnOfOneReferenceIsTheColumnItReads(t *testing.T) {
	dir := writeData(t, map[string]string{
		"mid": `{"format":"","score":"{/src.score}"}`,
		"row": `{"format":"","chain":"{/mid.score}","code":{"format":"{/src.code}","datatype":"integer"},"dot":"{.src.score}","label":"n={/src.score}","mixed":[{"format":"{text}","text":"{/src.score}"},"n/a"],"pick":["{/src.score}",{"format":"7","datatype":"integer"}],"same":"{/src.score}","score":"{/src.score}","text":"{/src.code}"}`,
		"src": `{"format":"","code":[null,"200","404"],"score":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	inline, err := f.NewRecordTemplate(`{"format":"","score":"{/src.score}"}`)
	if err != nil {
		t.Fatal(err)
	}
	nulls := map[string]int{}
	for i := 0; i < 200; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		cols := map[string]Column{}
		for _, c := range r.Columns() {
			cols[c.Name] = c
		}
		score, code := cols["score"], cols["code"]
		digit := func(c Column) bool { return len(c.Value) == 1 && c.Value >= "1" && c.Value <= "9" }
		isScore := func(name string) bool { return cols[name].Null != digit(cols[name]) }
		switch {
		case score.DataType != DataTypeInteger || cols["chain"].DataType != DataTypeInteger || code.DataType != DataTypeInteger || cols["pick"].DataType != DataTypeInteger || cols["label"].DataType != DataTypeString || cols["mixed"].DataType != DataTypeString || cols["text"].DataType != DataTypeString:
			t.Fatalf("%s: want score, chain, code and pick integer columns, label, mixed and text string ones", r.JSON())
		case !isScore("score") || !isScore("same") || !isScore("chain") || !isScore("dot"):
			t.Fatalf("%s: want score, same, chain and dot each null or a digit from src.score", r.JSON())
		case code.Null == (code.Value == "200" || code.Value == "404"):
			t.Fatalf("code = %+v, want null or a code from src.code", code)
		case cols["pick"].Value != "7" && !isScore("pick"), cols["mixed"].Null || cols["mixed"].Value != "n/a" && cols["mixed"].Value != "" && !digit(cols["mixed"]):
			t.Fatalf("pick = %+v, mixed = %+v: want pick 7 or a score, mixed n/a or a score's text", cols["pick"], cols["mixed"])
		case cols["label"].Null || !strings.HasPrefix(cols["label"].Value, "n="):
			t.Fatalf("label = %+v, want the text of the read, a null as \"\"", cols["label"])
		case strings.Contains(r.JSON(), `"score":null`) != score.Null:
			t.Fatalf("JSON() = %s, want a null score written null", r.JSON())
		}
		ir := inline.Fake()
		ic, sqlValue := ir.Columns()[0], ir.Columns()[0].Value
		if ic.Null {
			sqlValue = "NULL"
		}
		if ic.DataType != DataTypeInteger || ic.Null == (len(ic.Value) == 1) || ir.SQLInsert("t") != `INSERT INTO "t" ("score") VALUES (`+sqlValue+`);` || ir.CSVLine() != ic.Value {
			t.Fatalf("inline score = %+v written %s and %q, want an integer column, NULL and an empty field or a bare digit", ic, ir.SQLInsert("t"), ir.CSVLine())
		}
		for name, c := range map[string]Column{"code": code, "inline": ic, "score": score} {
			if c.Null {
				nulls[name]++
			}
		}
	}
	for _, name := range []string{"code", "inline", "score"} {
		if n := nulls[name]; n == 0 || n == 200 {
			t.Errorf("%s drew null %d times in 200 records, want both outcomes", name, n)
		}
	}
}

func TestRecordSQLQuotesIdentifiers(t *testing.T) {
	dir := writeData(t, map[string]string{
		"row": `{"format": "", "postal-code": "1", "street-number": "2"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.FakeRecord("row")
	if err != nil {
		t.Fatal(err)
	}
	got := r.SQLInsert("my-table")
	if got != `INSERT INTO "my-table" ("postal-code", "street-number") VALUES ('1', '2');` {
		t.Fatalf("SQLInsert() = %q, want hyphenated identifiers and table quoted", got)
	}
}

func TestFakeRecordTemplateAndNewRecordTemplate(t *testing.T) {
	f := recordCat(t)
	in := `{"format":"{x} {y}","x":["1","2"],"y":["3","4"]}`
	want, err := f.FakeRecordTemplate(in)
	if err != nil {
		t.Fatalf("FakeRecordTemplate: %v", err)
	}
	if len(want.Columns()) != 2 {
		t.Fatalf("FakeRecordTemplate columns = %v, want two columns", want.Columns())
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
		{`["a","b"]`, "a record is a template whose fields are its columns"},
		{`{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`, "carries repeat 3"},
	} {
		if _, err := f.FakeRecordTemplate(c.input); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeRecordTemplate(%q) = %v, want an error naming %q", c.input, err, c.want)
		}
	}
}

func TestRecordRejectsATopLevelRepeat(t *testing.T) {
	dir := writeData(t, map[string]string{
		"rep": `{"format":"{a}-","repeat":3,"separator":"|","a":["x","y"]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if _, err := f.FakeRecord("rep"); err == nil || !strings.Contains(err.Error(), "carries repeat 3") {
		t.Errorf("FakeRecord on a repeating template = %v, want an error naming the repeat", err)
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

// TestRecordOfAChoiceNamesTheWrapper pins the way out of the first command a fixture
// author types: a choice has no columns, and the error names the record that gives it one.
func TestRecordOfAChoiceNamesTheWrapper(t *testing.T) {
	f := shipped(t, WithSeed(1))
	want := `{"format":"","personnummer":"{/sv_SE.personnummer}"}`
	_, err := f.FakeRecord("sv_SE.personnummer")
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("FakeRecord(sv_SE.personnummer) = %v, want it to name %s", err, want)
	}
	if !errors.Is(err, ErrNoColumns) {
		t.Errorf("FakeRecord(sv_SE.personnummer) = %v, want it to match ErrNoColumns", err)
	}
	if _, err := f.FakeRecordTemplate(`{"format":"","personnummer":"{/sv_SE.personnummer}"}`); err != nil {
		t.Fatalf("the named wrapper does not render: %v", err)
	}
}

func TestEveryRecordCarriesItsShapeOnceLoaded(t *testing.T) {
	f := newShippedWhole(t, WithSeed(1), onTestDay)
	records := 0
	if err := load.EachNode(&f.root, "", func(label string, n tree.Node) error {
		tm, isTemplate := n.(*tree.Template)
		if !isTemplate || !tm.IsRecord() || len(tm.Fields()) == 0 {
			return nil
		}
		records++
		if want := len(tm.Fields()) - len(tm.Lets()); len(tm.Columns()) != want {
			t.Errorf("%s: %d record columns, want one per field outside let, %d", label, len(tm.Columns()), want)
		}
		for _, c := range tm.Columns() {
			if c.Field() != tm.Fields()[c.Name()] || c.DataType() != fence.ColumnDatatype(c.Field(), nil) {
				t.Errorf("%s: column %q holds field %v of datatype %s, want field %q's", label, c.Name(), c.Field(), c.DataType(), c.Name())
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if records == 0 {
		t.Fatal("no shipped record walked")
	}
}

func TestRecordSQLHeaderAndValuesJoinIntoAMultiRowInsert(t *testing.T) {
	t.Parallel()
	dir := writeData(t, map[string]string{"t": `{"format":"","age":{"format":"42","datatype":"integer"},"name":"O'Brien"}`})
	f, err := New(WithDataPath(dir), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.FakeRecord("t")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.SQLInsertHeader("my table"), `INSERT INTO "my table" ("age", "name") VALUES`; got != want {
		t.Errorf("SQLInsertHeader = %q, want %q", got, want)
	}
	if got, want := r.SQLValues(), `(42, 'O''Brien')`; got != want {
		t.Errorf("SQLValues = %q, want %q", got, want)
	}
}

func TestRecordLetFieldsWriteNoColumn(t *testing.T) {
	t.Parallel()
	dir := writeData(t, map[string]string{
		"sched": `{"format":"{weekday as d}","let":{"weekday":["MO","TU"],"helper":"x"},"day":"{d}","rrule":"BYDAY={d}"}`,
		"copy":  `{"format":"","rrule":"{/sched.rrule}","weekday":"{/sched.weekday}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	inline, err := f.NewRecordTemplate(`{"format":"{weekday as d}","let":{"weekday":["MO","TU"]},"day":"{d}"}`)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		r, err := f.FakeRecord("sched")
		if err != nil {
			t.Fatal(err)
		}
		cols := r.Columns()
		if len(cols) != 2 || cols[0].Name != "day" || cols[1].Name != "rrule" || cols[1].Value != "BYDAY="+cols[0].Value || cols[0].Value != "MO" && cols[0].Value != "TU" {
			t.Fatalf("Columns = %+v, want day and rrule only, rrule reading the day's pick", cols)
		}
		if got := r.CSVHeader(); got != "day,rrule" {
			t.Fatalf("CSVHeader = %q, want day,rrule", got)
		}
		seen[cols[0].Value] = true
		if c := inline.Fake().Columns(); len(c) != 1 || c[0].Name != "day" {
			t.Fatalf("inline Columns = %+v, want day only", c)
		}
	}
	if len(seen) != 2 {
		t.Errorf("drew days %v in 50 records, want both", seen)
	}
	cp, err := f.FakeRecord("copy")
	if err != nil {
		t.Fatal(err)
	}
	if got := cp.Columns()[1].Name; got != "weekday" || cp.Columns()[1].Value == "" {
		t.Errorf("copy's columns = %+v, want a let field still readable from another category as weekday", cp.Columns())
	}
}

func TestRecordLetErrors(t *testing.T) {
	t.Parallel()
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}), WithSeed(1))
	for _, c := range []struct{ input, want string }{
		{`{"format":"","let":"a","b":"y"}`, "let is an option and takes an object of fields, not a string"},
		{`{"format":"","let":{"a":"x"},"a":"y"}`, `field "a" is in let and beside it`},
		{`{"format":"","let":{"a.b":"x"},"c":"y"}`, "field"},
		{`{"format":"","let":{"weight":"w"},"b":"y"}`, `"weight" is an option`},
	} {
		if _, err := f.NewRecordTemplate(c.input); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("NewRecordTemplate(%s) = %v, want an error containing %q", c.input, err, c.want)
		}
	}
	_, err := f.NewRecordTemplate(`{"format":"","let":{"a":"x"}}`)
	if !errors.Is(err, ErrNoColumns) || !strings.Contains(err.Error(), `move one out of "let"`) {
		t.Errorf("NewRecordTemplate(only let fields) = %v, want ErrNoColumns naming let", err)
	}
	if _, err := f.NewTemplate(`{"format":"{let}"}`); err == nil || !strings.Contains(err.Error(), "option") {
		t.Errorf("NewTemplate reading let = %v, want let refused as an option, never a field", err)
	}
}

func TestRecordColumnReadingAFieldThroughANameKeepsItsDatatypeAndNull(t *testing.T) {
	t.Parallel()
	dir := writeData(t, map[string]string{
		"sched": `{"format":"{pattern as p}{shape as s}","pattern":[{"format":"daily","len":null},{"format":"weekly","len":{"format":"{int(1,9)}","datatype":"integer"}}],"shape":{"format":"s","size":[null,{"format":"{int(1,9)}","datatype":"integer"}]},"len":"{p.len}","size":"{s.size}","kind":"{p}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	nulls := map[string]int{}
	for i := 0; i < 200; i++ {
		r, err := f.FakeRecord("sched")
		if err != nil {
			t.Fatal(err)
		}
		cols := map[string]Column{}
		for _, c := range r.Columns() {
			cols[c.Name] = c
		}
		l, s, kind := cols["len"], cols["size"], cols["kind"].Value
		digit := func(c Column) bool { return !c.Null && len(c.Value) == 1 && c.Value >= "1" && c.Value <= "9" }
		switch {
		case l.DataType != DataTypeInteger || s.DataType != DataTypeInteger:
			t.Fatalf("%s: want len and size integer columns", r.JSON())
		case kind == "daily" && !l.Null, kind == "weekly" && !digit(l), kind != "daily" && kind != "weekly":
			t.Fatalf("%s: want len null for daily and a digit for weekly", r.JSON())
		case s.Null == digit(s):
			t.Fatalf("size = %+v, want null or a digit", s)
		case strings.Contains(r.JSON(), `"len":null`) != l.Null || strings.Contains(r.JSON(), `"len":"`) || strings.Contains(r.JSON(), `"size":"`):
			t.Fatalf("JSON() = %s, want a null written null and a digit written bare", r.JSON())
		}
		for name, c := range map[string]Column{"len": l, "size": s} {
			if c.Null {
				nulls[name]++
			}
		}
	}
	for _, name := range []string{"len", "size"} {
		if n := nulls[name]; n == 0 || n == 200 {
			t.Errorf("%s drew null %d times in 200 records, want both outcomes", name, n)
		}
	}
}

func TestRecordColumnReadingThroughItemsOfDifferentDatatypesIsRefused(t *testing.T) {
	t.Parallel()
	dir := writeData(t, map[string]string{
		"c": `{"format":"{pattern as p}","pattern":[{"format":"a","len":{"format":"1","datatype":"integer"}},{"format":"b","len":"x"}],"length":"{p.len}"}`,
	})
	if _, err := New(WithDataPath(dir), WithSeed(1), onTestDay); err == nil || !strings.Contains(err.Error(), `c: field "length"`) || !strings.Contains(err.Error(), "integer") {
		t.Errorf("New = %v, want a column whose items hold integer and text refused", err)
	}
}

func TestRecordColumnReadingTwoFieldsDeepKeepsItsDatatypeAndNull(t *testing.T) {
	t.Parallel()
	dir := writeData(t, map[string]string{
		"deep": `{"format":"{shape as p}","shape":{"format":"s","d":{"format":"x","e":[null,{"format":"{int(1,9)}","datatype":"integer"}]}},"e":"{p.d.e}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	seen := map[bool]bool{}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("deep")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()[0]
		if c.DataType != DataTypeInteger || c.Null != (c.Value == "") {
			t.Fatalf("column = %+v, want an integer column, null or a digit", c)
		}
		seen[c.Null] = true
	}
	if !seen[true] || !seen[false] {
		t.Errorf("100 records drew null and digit as %v, want both", seen)
	}
}

func TestColumnIsTheEngineColumn(t *testing.T) {
	if reflect.TypeOf(Column{}) != reflect.TypeOf(engine.Column{}) {
		t.Errorf("fejkdata.Column is %T, want the engine's %T, as DataType is", Column{}, engine.Column{})
	}
}
