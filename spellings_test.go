package fejkdata

import (
	"regexp"
	"strings"
	"testing"
)

// sameRenders holds long to render as short does, draw for draw under one seed.
func sameRenders(t *testing.T, long, short string) {
	t.Helper()
	l, s := compiled(t, long), compiled(t, short)
	fl, fs := engine(1), engine(1)
	for i := 0; i < 20; i++ {
		if a, b := renderOnce(fl.drawState, l), renderOnce(fs.drawState, s); a != b {
			t.Fatalf("%s rendered %q where %s rendered %q", long, a, short, b)
		}
	}
}

func TestOneItemChoiceIsItsItem(t *testing.T) {
	for long, short := range map[string]string{
		`["x"]`:                            `"x"`,
		`[{"format":"{d}","d":["x","y"]}]`: `{"format":"{d}","d":["x","y"]}`,
		`{"format":"{w}","w":["only"]}`:    `{"format":"{w}","w":"only"}`,
		`{"format":"{w}","w":[["a","b"]]}`: `{"format":"{w}","w":["a","b"]}`,
	} {
		sameRenders(t, long, short)
	}
	if r, isTemplate := compiled(t, `[{"format":"","d":"x"}]`).(*template); !isTemplate || !r.isRecord {
		t.Error("a one-item choice of a record is no record, want the record it holds")
	}
	if _, err := resolved(t, `[{"format":"a","weight":0}]`); err == nil || !strings.Contains(err.Error(), "every weight is 0") {
		t.Errorf("a one-item choice of weight 0 = %v, want it refused as having nothing to draw", err)
	}
}

func TestAPartWithNoEffectRendersAsWithoutIt(t *testing.T) {
	for long, short := range map[string]string{
		`{"format":"Malmö"}`:                                                   `"Malmö"`,
		`{"format":"{digits(3)}"}`:                                             `"{digits(3)}"`,
		`[{"format":"a","weight":1},"b"]`:                                      `["a","b"]`,
		`[{"format":"a","weight":0},"b"]`:                                      `"b"`,
		`{"format":"a","weight":2}`:                                            `"a"`,
		`{"format":"a","weight":0}`:                                            `"a"`,
		`{"format":"{x}","x":{"format":"a","weight":2}}`:                       `{"format":"{x}","x":"a"}`,
		`[{"format":"a","weight":2}]`:                                          `"a"`,
		`{"format":"{x}","x":["a","b"],"repeat":1}`:                            `{"format":"{x}","x":["a","b"]}`,
		`{"format":"{x}","x":["a","b"],"separator":","}`:                       `{"format":"{x}","x":["a","b"]}`,
		`{"format":"{x}","x":["a","b"],"repeat":2,"separator":""}`:             `{"format":"{x}","x":["a","b"],"repeat":2}`,
		`{"format":"","n":{"format":"{x}","x":["1","2"],"datatype":"string"}}`: `{"format":"","n":{"format":"{x}","x":["1","2"]}}`,
		`{"format":"{n}","repeat":2,"n":{"format":"1","datatype":"integer"}}`:  `{"format":"{n}","repeat":2,"n":"1"}`,
		`[{"format":"1","datatype":"integer"},"x"]`:                            `["1","x"]`,
		`{"format":"{word as w}x","word":["a","b"]}`:                           `"x"`,
		`null`: `""`,
		`{"format":"{p}","p":{"format":"{x}","x":[null,"a"]}}`: `{"format":"{p}","p":{"format":"{x}","x":["","a"]}}`,
		`"{int(5,5)}"`: `"5"`,
		`{"format":"{x}","x":["a","b"],"repeat":0}`:                       `""`,
		`"{digits(0)}{upper(0)}{lower(0)}{hex(0)}{base64(0)}{nanoid(0)}"`: `""`,
		`"{float(1,1,2)}"`: `"1.00"`,
		`"{date(1990-01-01,1990-01-01,'2006-01-02')}"`: `"1990-01-01"`,
		`"{date(1990-01-01,1990-12-31,'x')}"`:          `"x"`,
		`"{time('x')}"`:                                `"x"`,
		`"{time('2006-01-02')}"`:                       `"1970-01-01"`,
	} {
		sameRenders(t, long, short)
	}
	if r := compiled(t, `{"format":"","n":{"format":"1","datatype":"string"}}`).(*template); !r.isRecord || *r.fields["n"].(*template).datatype != DataTypeString {
		t.Error("a column of datatype string is no string column of a record")
	}
	if got := mustRender(t, engine(1), `"{date(1990-01-01,1990-12-31,'15:04')}"`); !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(got) {
		t.Errorf("date with a clock-only layout = %q, want a clock", got)
	}
}

func TestAJSONNumberOrBooleanInARecordColumnIsRefused(t *testing.T) {
	for src, want := range map[string][]string{
		`{"format":"","n":5}`:                    {`"5"`, `{"format":"5","datatype":"integer"}`},
		`{"format":"","n":1.50}`:                 {`"1.50"`, `{"format":"1.50","datatype":"number"}`},
		`{"format":"","b":true}`:                 {`"true"`, `{"format":"true","datatype":"boolean"}`},
		`{"format":"","n":12345678901234567890}`: {`"12345678901234567890"`, `{"format":"12345678901234567890","datatype":"number"}`},
		`{"format":"","n":[5,"x"]}`:              {`"5"`, `{"format":"5","datatype":"integer"}`},
	} {
		_, err := resolved(t, src)
		for _, w := range want {
			if err == nil || !strings.Contains(err.Error(), w) {
				t.Errorf("compile(%s) = %v, want both spellings, %s among them", src, err, w)
			}
		}
	}
}

func TestAJSONNumberNoDatatypeHoldsIsOfferedAsTextAlone(t *testing.T) {
	_, err := resolved(t, `{"format":"","n":1e400}`)
	if err == nil || !strings.Contains(err.Error(), `write "1e400" for text`) || strings.Contains(err.Error(), `"datatype"`) {
		t.Errorf("compile(1e400 in a column) = %v, want text offered alone", err)
	}
}

func TestAJSONNumberOrBooleanIsItsText(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"n": `{"format":"{x}","x":{"format":"{a} {b} {c} {d}","a":5,"b":1.50,"c":12345678901234567890,"d":true}}`,
		"l": `[1e3]`,
	}), WithSeed(1))
	if v := fake(t, f, "n"); v != "5 1.50 12345678901234567890 true" {
		t.Errorf("n = %q, want each number and boolean as written", v)
	}
	if v := fake(t, f, "l"); v != "1e3" {
		t.Errorf("l = %q, want 1e3 as written", v)
	}
	for in, want := range map[string]string{"42": "42", " false ": "false", "[7]": "7"} {
		if v, err := f.FakeTemplate(in); err != nil || v != want {
			t.Errorf("FakeTemplate(%q) = %q, %v, want %q", in, v, err, want)
		}
	}
}

func TestARepeatedItemCountsAsWritten(t *testing.T) {
	for _, src := range []string{`["a", "a", "b"]`, `{"format":"{x|x|y}","x":"a","y":"b"}`} {
		f, n, as := engine(1), compiled(t, src), 0
		for i := 0; i < 600; i++ {
			if renderOnce(f.drawState, n) == "a" {
				as++
			}
		}
		if as < 340 || as > 460 {
			t.Errorf("%s drew a %d times in 600, want about 400", src, as)
		}
	}
}

func TestInlineFolderSigilsAreRejected(t *testing.T) {
	f := shipped(t)
	for _, input := range []string{"{.sv_SE.person.last}", "{..sv_SE.person.last}"} {
		_, err := f.NewTemplate(input)
		if err == nil || !strings.Contains(err.Error(), "write {/sv_SE.person.last}") {
			t.Errorf("NewTemplate(%q) = %v, want an error naming the root spelling", input, err)
		}
	}
}
