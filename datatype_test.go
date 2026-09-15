package fejkdata

import (
	"encoding/json"
	"regexp"
	"slices"
	"strconv"
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
		`{"format":"{p}","p":{"format":"{x}","x":[null,"a"]}}`: `so write ""`,
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error containing %q", src, err, want)
		}
	}
}

func TestDatatypeRejectsAValueItsTypeRejects(t *testing.T) {
	tree := map[string]string{
		"cat": `[{"format":"{code}","code":"200"},{"format":"{code}","code":"2x"}]`,
		"src": `{"format":"","code":[null,"200","2x"],"score":[null,{"format":"{int(1,9)}","datatype":"integer"}]}`,
	}
	for _, c := range []struct{ name, column, want string }{
		{"an item beside a typed one", `[{"format":"1","datatype":"integer"},"x"]`, `write it as {"format":"x","datatype":"integer"}`},
		{"a weighted item beside a typed one", `[{"format":"1","datatype":"integer"},{"format":"2","weight":3}]`, `give it "datatype": "integer"`},
		{"items of two datatypes", `[{"format":"1","datatype":"integer"},{"format":"true","datatype":"boolean"}]`, "a column holds one datatype"},
		{"an item beside a typed column it reads", `["{/src.score}","x"]`, `write it as {"format":"x","datatype":"integer"}`},
		{"a typed item beside a string column it reads", `["{/src.code}",{"format":"1","datatype":"integer"}]`, `item "{/src.code}" declares no datatype beside one declaring integer`},
		{"a datatype over a typed column", `{"format":"{/src.score}","datatype":"integer"}`, `{/src.score} takes datatype integer from the column it reads; drop "datatype"`},
		{"another datatype over a typed column", `{"format":"{/src.score}","datatype":"number"}`, `{/src.score} takes datatype integer from the column it reads; drop "datatype"`},
		{"a value of the column it reads", `{"format":"{/src.code}","datatype":"integer"}`, `"2x" is not an integer`},
		{"a null read into text", `{"format":"{x}","x":"{/src.score}","datatype":"integer"}`, "reads a null"},
		{"a sample with leading zeros", `{"format":"{digits(3)}","datatype":"integer"}`, "{digits(3)} prints text, not an integer"},
		{"a fraction", `{"format":"{v}","v":["1","1.5"],"datatype":"integer"}`, `"1.5" is not an integer`},
		{"past int64", `{"format":"9223372036854775808","datatype":"integer"}`, "past the int64 range"},
		{"a float sample", `{"format":"{float(0,1,2)}","datatype":"integer"}`, "{float(0,1,2)} prints a number, not an integer"},
		{"composed digits", `{"format":"1{digits(2)}","datatype":"integer"}`, "is not one value"},
		{"a sign before a sample", `{"format":"-{int(1,9)}","datatype":"integer"}`, "is not one value"},
		{"a repeat", `{"format":"{int(1,9)}","repeat":2,"separator":",","datatype":"integer"}`, "carries a repeat"},
		{"through a reference", `{"format":"{/cat.code}","datatype":"integer"}`, `"2x" is not an integer`},
		{"a bare dot", `{"format":".5","datatype":"number"}`, `".5" is not a number`},
		{"a plus sign", `{"format":"+1","datatype":"number"}`, `"+1" is not a number`},
		{"a text sample", `{"format":"{hex(4)}","datatype":"number"}`, "{hex(4)} prints text, not a number"},
		{"a capital", `{"format":"{b}","b":["true","True"],"datatype":"boolean"}`, `"True" is not a boolean`},
		{"a transform", `{"format":"{lowercase(b)}","b":["TRUE","FALSE"],"datatype":"boolean"}`, "{lowercase(b)} rewrites text"},
		{"a number as a boolean", `{"format":"{int(0,1)}","datatype":"boolean"}`, "{int(0,1)} prints an integer, not a boolean"},
		{"an operand that is not always a number", `{"format":"{calc(a * 2)}","a":["1","x"],"datatype":"number"}`, `operand "a": "x" is not a number`},
		{"a divisor that can be zero", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":"{int(0,9)}","datatype":"number"}`, "divides by b, which is not proven nonzero"},
		{"an overflow", `{"format":"{calc(a * a)}","a":"{digits(200)}","datatype":"number"}`, "is not proven within 1e300"},
		{"a division in an integer column", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":"{int(1,9)}","datatype":"integer"}`, "prints a number, not an integer"},
		{"a composed operand", `{"format":"{calc(n * 2)}","n":"{int(1,99)}.{digits(2)}","datatype":"number"}`, "{seq()}, {digits()} or {calc()}"},
		{"a divisor that prints as zero", `{"format":"{calc(1 / b, 2)}","b":"{float(4.9999999999999994e-79,5e-79,78)}","datatype":"number"}`, "divides by b, which is not proven nonzero"},
		{"a divisor that rounds to zero", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":"{float(0.1,1,0)}","datatype":"number"}`, "divides by b, which is not proven nonzero"},
		{"a zero among a divisor's literals", `{"format":"{calc(a / b)}","a":"{int(1,9)}","b":["0","5"],"datatype":"number"}`, "divides by b, which is not proven nonzero"},
		{"a negated divisor crossing zero", `{"format":"{calc(a / (-b + 10))}","a":"{int(1,9)}","b":"{int(1,20)}","datatype":"number"}`, "which is not proven nonzero"},
		{"a subtracted divisor crossing zero", `{"format":"{calc(a / (10 - b))}","a":"{int(1,9)}","b":"{int(1,20)}","datatype":"number"}`, "which is not proven nonzero"},
		{"a quotient past the limit", `{"format":"{calc(a / b / b)}","a":"{digits(300)}","b":"{float(0.000001,1,6)}","datatype":"number"}`, "is not proven within 1e300"},
		{"an operand past the limit", `{"format":"{calc(a)}","a":"{digits(400)}","datatype":"number"}`, "is not proven within 1e300"},
		{"a whole calc past int64", `{"format":"{calc(a * 2)}","a":"{seq()}","datatype":"integer"}`, "{calc(a * 2)} is not proven within int64"},
		{"a whole float past int64", `{"format":"{float(0,1e19,0)}","datatype":"integer"}`, "{float(0,1e19,0)} is not proven within int64"},
		{"a signed zero integer", `{"format":"-0","datatype":"integer"}`, `"-0" is zero written with a sign; write "0"`},
		{"a signed zero number", `{"format":"-0.00","datatype":"number"}`, `"-0.00" is zero written with a sign; write "0.00"`},
	} {
		row := `{"format":"","col":` + c.column + `}`
		files := map[string]string{"row": row}
		for name, body := range tree {
			files[name] = body
		}
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, files)))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want an error containing %q", c.name, err, c.want)
		}
		f := newGenerator(t, writeData(t, tree))
		if _, err := f.NewTemplate(row); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: NewTemplate = %v, want the inline template refused the same way", c.name, err)
		}
	}
}

var jsonInteger = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)

func TestDatatypeAcceptsAColumnThatAlwaysParses(t *testing.T) {
	cat := `[{"format":"{code}","code":"200"},{"format":"{code}","code":"404"}]`
	for _, column := range []string{
		`{"format":"{int(1,99)}","datatype":"integer"}`,
		`{"format":"{int(-9,-1)}","datatype":"integer"}`,
		`{"format":"{seq()}","datatype":"integer"}`,
		`{"format":"{/cat.code}","datatype":"integer"}`,
		`{"format":"{a|b}","a":"1","b":"{int(5,9)}","datatype":"integer"}`,
		`{"format":"{float(-1.5,9.5,0)}","datatype":"integer"}`,
		`{"format":"{float(-1,1,0)}","datatype":"integer"}`,
		`{"format":"{calc(a * b)}","a":"{int(-9,-1)}","b":"{int(0,9)}","datatype":"integer"}`,
		`{"format":"{calc(x + 1)}","x":"{float(0,9,0)}","datatype":"integer"}`,
		`{"format":"{float(-1,1,2)}","datatype":"number"}`,
		`{"format":"{v}","v":["1","2.5","6.022e23"],"datatype":"number"}`,
		`{"format":"{v}","v":["-1e-400","0.5"],"datatype":"number"}`,
		`{"format":"{b}","b":["true","false"],"datatype":"boolean"}`,
		`{"format":"{calc(net * qty, 2)}","net":["19.99","5.00"],"qty":["3","7"],"datatype":"number"}`,
		`{"format":"{calc(a + b)}","a":"{int(1,9)}","b":"{int(-9,9)}","datatype":"integer"}`,
		`{"format":"{calc(x / (b - c), 2)}","x":"{int(1,9)}","b":"{int(10,20)}","c":"{int(1,5)}","datatype":"number"}`,
		`{"format":"{calc(x * x * x * x, 2)}","x":"{float(0,1,80)}","datatype":"number"}`,
		`{"format":"{calc(a / (b + 1), 2)}","a":"{int(1,9)}","b":"{digits(2)}","datatype":"number"}`,
		`{"format":"{calc(a / b, 0)}","a":"{int(1,9)}","b":"{int(1,9)}","datatype":"integer"}`,
		`{"format":"{calc(sub * 1.25, 2)}","sub":{"format":"{calc(a * b)}","a":"{int(1,9)}","b":"{float(0,5,2)}"},"datatype":"number"}`,
		`{"format":"{/src.code}","datatype":"integer"}`,
	} {
		row := `{"format":"","col":` + column + `}`
		src := `{"format":"","code":["200","404"]}`
		f, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": cat, "row": row, "src": src})), WithSeed(1))
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
			_, int64Err := strconv.ParseInt(c.Value, 10, 64)
			if c.DataType == DataTypeBoolean && !isBool || c.DataType != DataTypeBoolean && !isNumber || c.DataType == DataTypeInteger && (!jsonInteger.MatchString(c.Value) || int64Err != nil) {
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
