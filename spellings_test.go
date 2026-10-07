package fejkdata

import (
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
	if _, err := resolved(t, `[{"format":"a","weight":2}]`); err == nil || !strings.Contains(err.Error(), "no effect here") {
		t.Errorf("a weight on a one-item choice's item = %v, want it refused as doing nothing", err)
	}
}

func TestARepeatedItemCountsAsWritten(t *testing.T) {
	for _, src := range []string{`["a", "a", "b"]`, `[{"format":"{x}","x":"1"},{"format":"{x}","x":"1"}]`, `{"format":"{w}","w":["", "", "x"]}`, `{"format":"","w":[null,null,"a"]}`, `{"format":"{x|x|y}","x":"1","y":"2"}`} {
		if _, err := resolved(t, src); err != nil {
			t.Errorf("compile(%s) = %v, want it loaded", src, err)
		}
	}
	f, n, as := engine(1), compiled(t, `["a", "a", "b"]`), 0
	for i := 0; i < 600; i++ {
		if renderOnce(f.drawState, n) == "a" {
			as++
		}
	}
	if as < 340 || as > 460 {
		t.Errorf(`["a","a","b"] drew a %d times in 600, want about 400`, as)
	}
}

func TestASeparatorWithoutARepeatIsRejected(t *testing.T) {
	for src, want := range map[string]string{
		`{"format":"{x}","x":"v","separator":","}`: "separator",
	} {
		if _, err := resolved(t, src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %s", src, err, want)
		}
	}
}

func TestADefaultWrittenOutIsTheDefault(t *testing.T) {
	for long, short := range map[string]string{
		`{"format":"Malmö"}`:                                                   `"Malmö"`,
		`{"format":"{digits(3)}"}`:                                             `"{digits(3)}"`,
		`[{"format":"a","weight":1},"b"]`:                                      `["a","b"]`,
		`{"format":"{x}","x":["a","b"],"repeat":1}`:                            `{"format":"{x}","x":["a","b"]}`,
		`{"format":"{x}","x":["a","b"],"repeat":2,"separator":""}`:             `{"format":"{x}","x":["a","b"],"repeat":2}`,
		`{"format":"","n":{"format":"{x}","x":["1","2"],"datatype":"string"}}`: `{"format":"","n":{"format":"{x}","x":["1","2"]}}`,
		`null`: `""`,
		`{"format":"{p}","p":{"format":"{x}","x":[null,"a"]}}`: `{"format":"{p}","p":{"format":"{x}","x":["","a"]}}`,
	} {
		sameRenders(t, long, short)
	}
	if r := compiled(t, `{"format":"","n":{"format":"1","datatype":"string"}}`).(*template); !r.isRecord || r.fields["n"].(*template).datatype != DataTypeString {
		t.Error("a column of datatype string is no string column of a record")
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
