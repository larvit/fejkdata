package fejkdata

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestDatatypeAndNullSitOnlyInAColumn(t *testing.T) {
	for _, src := range []string{
		`{"format":"","age":{"format":"{int(18,99)}","datatype":"integer"}}`,
		`{"format":"","n":{"format":"42","datatype":"integer"}}`,
		`{"format":"","gone":null}`,
		`{"format":"","middle":[null,"Ann","Eva"]}`,
		`{"format":"","age":[null,{"format":"{int(18,99)}","datatype":"integer","weight":9}]}`,
		`{"format":"","pick":[[null,"a"],"b"]}`,
	} {
		if _, err := compile(parse(t, src)); err != nil {
			t.Errorf("compile(%s) = %v, want a column to take a datatype and null", src, err)
		}
	}
	for src, want := range map[string]string{
		`{"format":"","n":{"format":"1","datatype":"int"}}`:                             `datatype takes "integer", "number" or "boolean", got "int"`,
		`{"format":"","n":{"format":"1","datatype":1}}`:                                 "datatype must be a string",
		`{"format":"{int(1,9)}","datatype":"integer"}`:                                  "datatype only types a record column",
		`[{"format":"1","datatype":"integer"},"x"]`:                                     "datatype only types a record column",
		`{"format":"{p}","p":{"format":"{n}","n":{"format":"1","datatype":"integer"}}}`: "datatype only types a record column",
		`{"format":"{n}","repeat":2,"n":{"format":"1","datatype":"integer"}}`:           "datatype only types a record column",
		`null`: `so write ""`,
		`{"format":"{p}","p":{"format":"{x}","x":[null,"a"]}}`:        `so write ""`,
		`{"format":"","c":[{"format":"1","datatype":"integer"},"x"]}`: "a column holds one datatype",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error containing %q", src, err, want)
		}
	}
}

func TestDatatypeRejectsARenderItsTypeRejects(t *testing.T) {
	cat := `[{"format":"{code}","code":"200"},{"format":"{code}","code":"2x"}]`
	for _, c := range []struct{ name, column, want string }{
		{"a leading zero", `{"format":"{digits(3)}","datatype":"integer"}`, "which is not an integer"},
		{"a fraction", `{"format":"{v}","v":["1","1.5"],"datatype":"integer"}`, `can render "1.5", which is not an integer`},
		{"a signed sample before digits", `{"format":"{int(-5,5)}{digits(2)}","datatype":"integer"}`, "which is not an integer"},
		{"a separator", `{"format":"{int(1,9)}","repeat":2,"separator":",","datatype":"integer"}`, "which is not an integer"},
		{"through a reference", `{"format":"{/cat.code}","datatype":"integer"}`, `can render "2x"`},
		{"a bare dot", `{"format":".5","datatype":"number"}`, `can render ".5", which is not a number`},
		{"a trailing dot", `{"format":"{int(1,9)}.","datatype":"number"}`, "which is not a number"},
		{"a plus sign", `{"format":"+1","datatype":"number"}`, `can render "+1"`},
		{"a capital", `{"format":"{b}","b":["true","True"],"datatype":"boolean"}`, `can render "True", which is not a boolean`},
		{"an upper-casing transform", `{"format":"{uppercase(b)}","b":["true","false"],"datatype":"boolean"}`, "which is not a boolean"},
		{"an operand that is not always a number", `{"format":"{calc(a * 2)}","a":["1","x"],"datatype":"number"}`, `operand "a" can render "x"`},
		{"a divisor that can be zero", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":"{int(0,9)}","datatype":"number"}`, "divides by b, which can be zero"},
		{"an overflow", `{"format":"{calc(a * a)}","a":"{digits(200)}","datatype":"number"}`, "can overflow"},
		{"a division in an integer column", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":"{int(1,9)}","datatype":"integer"}`, "which is not an integer"},
	} {
		row := `{"format":"","col":` + c.column + `}`
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": cat, "row": row})))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want an error containing %q", c.name, err, c.want)
		}
		f := newGenerator(t, writeData(t, map[string]string{"cat": cat}))
		if _, err := f.NewTemplate(row); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: NewTemplate = %v, want the inline template refused the same way", c.name, err)
		}
	}
}

var integerText = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

func TestDatatypeAcceptsAColumnThatAlwaysParses(t *testing.T) {
	cat := `[{"format":"{code}","code":"200"},{"format":"{code}","code":"404"}]`
	for _, column := range []string{
		`{"format":"{int(1,99)}","datatype":"integer"}`,
		`{"format":"-{int(1,9)}","datatype":"integer"}`,
		`{"format":"1{digits(2)}","datatype":"integer"}`,
		`{"format":"{seq()}","datatype":"integer"}`,
		`{"format":"{/cat.code}","datatype":"integer"}`,
		`{"format":"{float(-1,1,2)}","datatype":"number"}`,
		`{"format":"{int(1,9)}e{int(1,9)}","datatype":"number"}`,
		`{"format":"6.022e23","datatype":"number"}`,
		`{"format":"{lowercase(b)}","b":["TRUE","False"],"datatype":"boolean"}`,
		`{"format":"{calc(net * qty, 2)}","net":["19.99","5.00"],"qty":["3","7"],"datatype":"number"}`,
		`{"format":"{calc(a + b)}","a":"{int(1,9)}","b":"{int(-9,9)}","datatype":"integer"}`,
		`{"format":"{calc(a / (b + 1), 2)}","a":"{int(1,9)}","b":"{digits(2)}","datatype":"number"}`,
		`{"format":"{calc(a / b, 0)}","a":"{int(1,9)}","b":"{int(1,9)}","datatype":"integer"}`,
		`{"format":"{calc(sub * 1.25, 2)}","sub":{"format":"{calc(a * b)}","a":"{int(1,9)}","b":"{float(0,5,2)}"},"datatype":"number"}`,
	} {
		row := `{"format":"","col":` + column + `}`
		f, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": cat, "row": row})), WithSeed(1))
		if err != nil {
			t.Errorf("%s: New = %v, want it loaded", column, err)
			continue
		}
		for i := 0; i < 200; i++ {
			r, err := f.FakeRecord("row")
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			if err := json.Unmarshal([]byte(r.JSON()), &m); err != nil {
				t.Errorf("%s: JSON() = %s is not JSON: %v", column, r.JSON(), err)
				break
			}
			c := r.Columns()[0]
			_, isBool := m["col"].(bool)
			_, isNumber := m["col"].(float64)
			if c.DataType == DataTypeBoolean && !isBool || c.DataType != DataTypeBoolean && !isNumber || c.DataType == DataTypeInteger && !integerText.MatchString(c.Value) {
				t.Errorf("%s: column %+v written as %s, want its datatype", column, c, r.JSON())
				break
			}
		}
	}
}

func TestNullColumn(t *testing.T) {
	dir := writeData(t, map[string]string{
		"row": `{"format":"[{middle}]","gone":null,"middle":[null,"Ann"],"score":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	drew := map[bool]bool{}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		gone, middle, score := r.Columns()[0], r.Columns()[1], r.Columns()[2]
		if !gone.Null || gone.Value != "" || gone.DataType != DataTypeString {
			t.Fatalf("gone = %+v, want a null string column every draw", gone)
		}
		if middle.Null == (middle.Value == "Ann") {
			t.Fatalf("middle = %+v, want null or Ann", middle)
		}
		if score.DataType != DataTypeInteger {
			t.Fatalf("score = %+v, want the integer its non-null item declares, null or not", score)
		}
		drew[middle.Null] = true
	}
	if len(drew) != 2 {
		t.Errorf("middle drew only null=%v in 100 records, want both", drew)
	}
	if v := fake(t, f, "row"); v != "[]" && v != "[Ann]" {
		t.Errorf("Fake(row) = %q, want a null to render as \"\"", v)
	}
	if v := fake(t, f, "row.gone"); v != "" {
		t.Errorf("Fake(row.gone) = %q, want \"\"", v)
	}
	if !slices.Contains(f.List(), "row.gone") {
		t.Errorf("List() = %v, want the null column row.gone, which Fake accepts", f.List())
	}
}

func TestEveryBuiltinSaysWhatItEmits(t *testing.T) {
	for name, b := range builtins {
		if _, isTransform := transforms[name]; b.emits == nil && name != "calc" && !isTransform {
			t.Errorf("builtin %s declares no emits, so a typed column calling it cannot be checked", name)
		}
	}
}
