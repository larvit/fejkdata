package fejkdata

import (
	"encoding/csv"
	"encoding/json"
	"reflect"
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
		{"pick", "names a choice"},
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

func TestRecordRejectsOverlappingReferenceColumns(t *testing.T) {
	cat := `{"format":"{a}","a":[{"format":"A={b}","b":"1"},{"format":"A={b}","b":"2"}]}`
	for _, c := range []struct{ name, row string }{
		{"through a column repeat, which draws anew", `{"format":"","whole":"{/cat.a}","inner":{"format":"{/cat.a.b}","repeat":2,"separator":"-"}}`},
		{"in a group of its own", `{"format":"","whole":{"format":"{/cat.a}","drawGroup":"g"},"inner":"{/cat.a.b}"}`},
	} {
		f := newGenerator(t, writeData(t, map[string]string{"cat": cat, "row": c.row}), WithSeed(1))
		if _, err := f.FakeRecord("row"); err != nil {
			t.Errorf("%s: FakeRecord = %v, want it accepted", c.name, err)
		}
	}
	for _, c := range []struct{ name, row string }{
		{"sibling columns", `{"format":"","whole":"{/cat.a}","inner":"{/cat.a.b}"}`},
		{"through a nested template", `{"format":"","whole":"{/cat.a}","inner":{"format":"{/cat.a.b} {x}","x":"1"}}`},
		{"through a choice variant", `{"format":"","whole":"{/cat.a}","inner":[{"format":"{/cat.a.b} {x}","x":"1"},{"format":"{/cat.a.b}! {x}","x":"2"}]}`},
		{"as a builtin operand", `{"format":"","whole":"{uppercase(/cat.a)}","inner":"{/cat.a.b}"}`},
		{"a bare reference beside a path", `{"format":"","whole":"{/cat}","inner":"{/cat.a.b}"}`},
	} {
		f := newGenerator(t, writeData(t, map[string]string{"cat": cat, "row": c.row}), WithSeed(1))
		_, err := f.FakeRecord("row")
		if err == nil || !strings.Contains(err.Error(), "reads a path into") {
			t.Errorf("%s: FakeRecord = %v, want the overlap rejected the way one format is", c.name, err)
			continue
		}
		if !strings.Contains(err.Error(), `"whole"`) || !strings.Contains(err.Error(), `"inner"`) {
			t.Errorf("%s: error %q names neither column; it must name both", c.name, err)
		}
		if _, err := f.Fake("row"); err != nil {
			t.Errorf("%s: Fake(row) = %v, want the string view untouched", c.name, err)
		}
	}
}

func TestInlineRecordRejectsOverlappingColumns(t *testing.T) {
	dir := writeData(t, map[string]string{
		"cat": `{"format":"","a":[{"format":"A={b}","b":"1"},{"format":"A={b}","b":"2"}]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	_, err := f.FakeRecordTemplate(`{"format":"","whole":"{/cat.a}","inner":"{/cat.a.b}"}`)
	if err == nil || !strings.Contains(err.Error(), "reads a path into") {
		t.Fatalf("inline record over an overlapping pair = %v, want the inline entry point to refuse it too", err)
	}
}

func TestRecordRejectsAColumnReadingItsOwnRecord(t *testing.T) {
	for _, c := range []struct{ name, column string }{
		{"a path into itself", `"full":"{/person.first} {/person.last}"`},
		{"the record read whole", `"whole":"{/person}"`},
		{"the record as an operand", `"up":"{uppercase(/person)}"`},
	} {
		person := `{"format":"{first} {last}","first":["Ada","Bo"],"last":["Lovelace","Ek"],` + c.column + `}`
		f := newGenerator(t, writeData(t, map[string]string{"person": person}), WithSeed(1))
		if _, err := f.FakeRecord("person"); err == nil || !strings.Contains(err.Error(), "points back at this record") {
			t.Errorf("%s: FakeRecord = %v, want it refused; the column would contradict the columns beside it", c.name, err)
		}
	}
}

func TestRecordBareReferenceStaysIndependentAsAnOperand(t *testing.T) {
	dir := writeData(t, map[string]string{
		"cur": `[{"format":"{code}","code":"aud"},{"format":"{code}","code":"eur"}]`,
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
	for i := 0; i < 50; i++ {
		v, err := f.FakeTemplate("{/cur}|{uppercase(/cur)}")
		if err != nil {
			t.Fatal(err)
		}
		if parts := strings.Split(v, "|"); !strings.EqualFold(parts[0], parts[1]) {
			t.Fatalf("one format rendered %q; within an expansion a bare reference is still one draw", v)
		}
	}
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
		"row": `[{"format":"{x}","x":"1"},{"format":"{x}","x":"2"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for _, path := range []string{"cat.sub", "row.x"} {
		if _, err := f.FakeRecord(path); err == nil || !strings.Contains(err.Error(), "field") {
			t.Errorf("FakeRecord(%q) = %v, want a 'descends into a field' error", path, err)
		}
	}
}

func TestRecordSharesAReferenceAcrossColumns(t *testing.T) {
	dir := writeData(t, map[string]string{
		"currency": `[{"format":"{code}","code":"AUD","symbol":"$"},{"format":"{code}","code":"EUR","symbol":"€"}]`,
		"price":    `{"format":"{code} {symbol}","code":"{/currency.code}","symbol":"{/currency.symbol}"}`,
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

func TestRepeatIterationsDrawReferencesAnew(t *testing.T) {
	dir := writeData(t, map[string]string{
		"party":  `{"format":"{host}: {guests}","guests":{"format":"{/person.first} {/person.last}","repeat":3,"separator":", "},"host":"{/person.first} {/person.last}"}`,
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

func TestRecordGroupsDrawApart(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person":   drawPeople,
		"transfer": `{"format":"{from_first} {from_last} to {to_first} {to_last}","from_first":{"format":"{/person.first}","drawGroup":"from"},"from_last":{"format":"{/person.last}","drawGroup":"from"},"to_first":{"format":"{/person.first}","drawGroup":"to"},"to_last":{"format":"{/person.last}","drawGroup":"to"}}`,
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
				t.Fatalf("%s drew %q and %q, want each group one person", view, pair[0], pair[1])
			}
			apart[view] = apart[view] || pair[0] != pair[1]
		}
	}
	for _, view := range []string{"FakeRecord", "Fake"} {
		if !apart[view] {
			t.Errorf("%s: groups from and to drew one person in 100 renders, want a draw each", view)
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
		agree := func(name string, want Column) bool {
			return cols[name].Null == want.Null && cols[name].Value == want.Value
		}
		switch {
		case score.DataType != DataTypeInteger || cols["chain"].DataType != DataTypeInteger || code.DataType != DataTypeInteger || cols["pick"].DataType != DataTypeInteger || cols["label"].DataType != DataTypeString || cols["mixed"].DataType != DataTypeString || cols["text"].DataType != DataTypeString:
			t.Fatalf("%s: want score, chain, code and pick integer columns, label, mixed and text string ones", r.JSON())
		case score.Null == (len(score.Value) == 1 && score.Value >= "1" && score.Value <= "9"):
			t.Fatalf("score = %+v, want null or a digit from src.score", score)
		case code.Null == (code.Value == "200" || code.Value == "404"):
			t.Fatalf("code = %+v, want null or a code from src.code", code)
		case !agree("same", score) || !agree("chain", score) || !agree("dot", score) || !agree("text", code):
			t.Fatalf("%s: want every read of one src column one draw, through mid too", r.JSON())
		case cols["pick"].Value != "7" && !agree("pick", score), cols["mixed"].Null || cols["mixed"].Value != "n/a" && cols["mixed"].Value != score.Value:
			t.Fatalf("pick = %+v, mixed = %+v beside score %+v: want pick 7 or the score's draw, mixed n/a or the score's text", cols["pick"], cols["mixed"], score)
		case cols["label"].Null || cols["label"].Value != "n="+score.Value:
			t.Fatalf("label = %+v beside score %+v, want the text of the read, a null as \"\"", cols["label"], score)
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
