package fejkdata

import (
	"strings"
	"testing"
)

// nameTables is a region table linked to by a municipality table.
func nameTables() map[string]string {
	return map[string]string{
		"region.json":       `{"format":"{name}","rows":"region.tsv","key":"code"}`,
		"region.tsv":        "code\tname\n01\tStockholms län\n12\tSkåne län\n",
		"municipality.json": `{"format":"{name}","rows":"municipality.tsv","key":"code","parent":"region"}`,
		"municipality.tsv":  "code\tname\tregion\n0180\tStockholm\t01\n0184\tSolna\t01\n1280\tMalmö\t12\n1281\tLund\t12\n",
	}
}

var municipalityCode = map[string]string{"Stockholm": "0180", "Solna": "0184", "Malmö": "1280", "Lund": "1281"}

var municipalityRegion = map[string]string{"Stockholm": "Stockholms län", "Solna": "Stockholms län", "Malmö": "Skåne län", "Lund": "Skåne län"}

func TestANameIsOnePickOfEverythingUnderIt(t *testing.T) {
	files := nameTables()
	files["card.json"] = `"{/region as r}{r.name}|{r.municipality.name}|{r.municipality.code}|{r}"`
	f := newGenerator(t, writeFiles(t, files), WithSeed(3))
	for i := 0; i < 200; i++ {
		got := strings.Split(fake(t, f, "card"), "|")
		region, municipality, code, whole := got[0], got[1], got[2], got[3]
		if municipalityCode[municipality] != code || municipalityRegion[municipality] != region || whole != region {
			t.Fatalf("card = %q, want one region, and one municipality inside it", got)
		}
	}
}

func TestANameReadWholeAgreesWithItsPaths(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy"],"last":["Byron","Ek","Lind"]}`,
		"card":   `"{/person as p}{p}|{p.last}|{p.first}"`,
	})
	f := newGenerator(t, dir, WithSeed(5))
	for i := 0; i < 200; i++ {
		got := strings.Split(fake(t, f, "card"), "|")
		if got[0] != got[2]+" "+got[1] {
			t.Fatalf("card = %q, want the whole person to match its parts", got)
		}
	}
}

func TestANameKeepsWhatItsPathsAddress(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `{"format":"{w}-{w}","w":["a","b","c","d","e","f","g","h"]}`,
		"card": `"{/word as n}{n}|{n}"`,
	})
	f := newGenerator(t, dir, WithSeed(7))
	differ := false
	for i := 0; i < 100; i++ {
		got := strings.Split(fake(t, f, "card"), "|")
		if got[0] != got[1] {
			t.Fatalf("card = %q, want one pick of n", got)
		}
		differ = differ || got[0][0] != got[0][2]
	}
	if !differ {
		t.Fatal("{w}-{w} under a name drew w once in 100 renders, want each {w} a draw of its own")
	}
}

func TestABindingPrintsNothing(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `["a","b"]`,
		"card": `"<{/word as w}>{w}{w}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "card"); got != "<>aa" && got != "<>bb" {
		t.Fatalf("card = %q, want the binding to print nothing", got)
	}
}

func TestSiblingFieldsAndColumnsReadOneName(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy"],"last":["Byron","Ek","Lind"]}`,
		"row":    `{"format":"{first} {last}","first":"{/person as p}{p.first}","last":"{p.last}","whole":"{p}"}`,
	})
	f := newGenerator(t, dir, WithSeed(9))
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()
		if c[2].Value != c[0].Value+" "+c[1].Value {
			t.Fatalf("columns = %v, want one person across them", c)
		}
		if got := strings.Fields(fake(t, f, "row")); len(got) != 2 {
			t.Fatalf("row = %q, want a first and a last name", got)
		}
	}
}

func TestAFieldReadsTheNameItsCategoryBinds(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":"Ada","last":"Byron"}`,
		"row":    `{"format":"{first}","first":"{/person as p}{p.first}","last":"{p.last}"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "row.last"); got != "Byron" {
		t.Fatalf("row.last = %q, want Byron", got)
	}
}

func TestANameInsideARepeatPicksEachIteration(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `["a","b","c","d","e","f","g","h"]`,
		"line": `{"format":"{each}","each":{"format":"{/word as w}{w}{w}","repeat":40,"separator":" "}}`,
	})
	f := newGenerator(t, dir, WithSeed(11))
	seen := map[string]bool{}
	for _, pair := range strings.Fields(fake(t, f, "line")) {
		if pair[0] != pair[1] {
			t.Fatalf("pair %q, want one pick per iteration", pair)
		}
		seen[pair] = true
	}
	if len(seen) < 2 {
		t.Fatalf("pairs %v, want each iteration to pick anew", seen)
	}
}

func TestANameOutsideARepeatKeepsItsPick(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `["a","b","c","d","e","f","g","h"]`,
		"line": `{"format":"{/word as w}{each}","each":{"format":"{w}","repeat":40,"separator":" "}}`,
	})
	f := newGenerator(t, dir, WithSeed(13))
	got := strings.Fields(fake(t, f, "line"))
	for _, w := range got {
		if w != got[0] {
			t.Fatalf("line = %q, want one pick on every line", got)
		}
	}
}

func TestTwoRendersPickTwice(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `["a","b","c","d","e","f","g","h"]`,
		"one":  `"{/word as w}{w}{w}"`,
		"two":  `"{/one}{/one}{/one}{/one}{/one}{/one}"`,
	})
	f := newGenerator(t, dir, WithSeed(17))
	got := fake(t, f, "two")
	if strings.Count(got, got[:1]) == len(got) || got[0] != got[1] {
		t.Fatalf("two = %q, want each bare reference to its own pick", got)
	}
}

func TestANameInAnInlineTemplate(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy"],"last":["Byron","Ek","Lind"]}`,
	})
	f := newGenerator(t, dir, WithSeed(19))
	for i := 0; i < 50; i++ {
		got := strings.Split(fakeTemplate(t, f, "{/person as p}{p}|{p.first} {p.last}"), "|")
		if got[0] != got[1] {
			t.Fatalf("template = %q, want one person", got)
		}
	}
}

func TestAChoiceItemKeepsTheNamesItBinds(t *testing.T) {
	f, n, apart := engine(1), compiled(t, `{"format":"{x}|{x}","x":[{"format":"{v as w}{w}{w}","v":["a","b"]},"c"]}`), false
	for i := 0; i < 100; i++ {
		got := renderOnce(f.drawState, n)
		halves := strings.Split(got, "|")
		for _, h := range halves {
			if h != "aa" && h != "bb" && h != "c" {
				t.Fatalf("render %q holds %q, want each draw of the item to read one pick twice", got, h)
			}
		}
		apart = apart || halves[0] != "c" && halves[1] != "c" && halves[0] != halves[1]
	}
	if !apart {
		t.Error("100 renders of {x}|{x} never drew two different halves, want a pick per draw of the item")
	}
}

func TestNameErrors(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"bound twice", map[string]string{"word": `["a","b"]`, "card": `{"format":"{x}{y}","x":"{/word as w}{w}","y":"{/word as w}"}`},
			`name "w" is bound twice outside any repeat`},
		{"bound inside a repeat written first", map[string]string{"word": `["a","b"]`, "card": `{"format":"{a}{z}","a":{"format":"{/word as w}{w}","repeat":2},"z":"{/word as w}{w}"}`},
			`name "w" is bound outside this repeat too`},
		{"bound twice in one repeat", map[string]string{"word": `["a","b"]`, "card": `{"format":"{x}","x":{"format":"{/word as n}{/word as n}{n}","repeat":2}}`},
			`name "n" is bound twice in one repeat`},
		{"bound in a choice's item", map[string]string{"word": `["a","b"]`, "card": `{"format":"{x}{w}","x":["{/word as w}","b"]}`},
			`name "w" is bound at field "x", inside an item of a choice, which a read outside that item cannot see`},
		{"bound outside the repeat too", map[string]string{"word": `["a","b"]`, "card": `{"format":"{/word as w}{x}","x":{"format":"{/word as w}{w}","repeat":2}}`},
			`name "w" is bound outside this repeat too`},
		{"a field too", map[string]string{"word": `["a","b"]`, "card": `{"format":"{/word as w}{w}","w":"x"}`},
			`name "w" is a field of the root template too`},
		{"an option", map[string]string{"word": `["a","b"]`, "card": `"{/word as format}{format}"`},
			`"format" is an option and can never be a name`},
		{"no reference or field", map[string]string{"card": `{"format":"{word as w}{w}{w}","words":["a","b"]}`},
			`no field "word"; a binding names a field, or a reference`},
		{"read once through a path that does not resolve", map[string]string{"word": `{"format":"{w}","w":["a","b"]}`, "card": `"{/word as a}{a.zz}"`},
			`no field "zz"`},
		{"read once through a field path that does not resolve", map[string]string{"card": `{"format":"{place as p}{p.zz}","place":{"format":"{x}","x":["a","b"]}}`},
			`no field "zz"`},
		{"one arm read twice in an alternation", map[string]string{"cat": `{"format":"{w|w}","w":["a","b"]}`, "card": `"{/cat as n}{n} {n.w}"`},
			""},
		{"a field and a path into it in one alternation", map[string]string{"cat": `{"format":"{w|w.x}","w":{"format":"{x}","x":["a","b"]}}`, "card": `"{/cat as n}{n} {n.w}"`},
			""},
		{"one field read in two alternations", map[string]string{"cat": `{"format":"{w|v}{w|v}","w":["a","b"],"v":["c","d"]}`, "card": `"{/cat as n}{n} {n.w}"`},
			`renders field "w" twice`},
		{"read in a repeat", map[string]string{"nm.json": `{"format":"{first}","rows":"nm.tsv"}`, "nm.tsv": "first\tlast\nAda\tByron\nBo\tEk\n", "person": `{"format":"{first} {last}","first":"{/nm.first}","last":"{/nm.last}"}`, "card": `{"format":"{/person as p}{each}","each":{"format":"{p.first}","repeat":2}}`},
			""},
		{"a misspelt name", map[string]string{"word": `["a","b"]`, "card": `"{/word as p}{p}{p}{pp}"`},
			`no field or name "pp"; the names bound here are "p"`},
		{"read outside a repeat in the root choice", map[string]string{"word": `["a","b"]`, "card": `[{"format":"{/word as p}{p}{p}","repeat":2},"{p}"]`},
			`name "p" is bound at an item of the root choice, inside a repeat`},
		{"an empty name", map[string]string{"word": `["a","b"]`, "card": `"{/word as }"`},
			"a binding names nothing"},
		{"rendered twice where a path reads it", map[string]string{"word": `{"format":"{w}-{w}","w":["a","b"]}`, "card": `"{/word as n}{n}|{n.w}"`},
			`{n} renders field "w" twice`},
		{"a binding in a cell", map[string]string{"word": `["a","b"]`, "t.json": `{"format":"{k}","rows":"t.tsv","key":"k"}`, "t.tsv": "k\tc\nx\t{/word as w}{w}{w}\ny\tz\n"},
			"a table binds no name"},
		{"in a category's name", map[string]string{"a as b": `"x"`},
			`contains " as "`},
		{"in a folder's name", map[string]string{"a as b/c": `"x"`},
			`contains " as "`},
		{"in a column's name", map[string]string{"t.json": `{"format":"{k}","rows":"t.tsv","key":"k"}`, "t.tsv": "k\ta as b\nx\t1\ny\t2\n"},
			`contains " as "`},
		{"read outside the repeat binding it", map[string]string{"word": `["a","b"]`, "card": `{"format":"{x}{w}","x":{"format":"{/word as w}{w}","repeat":2}}`},
			`no field "w"; name "w" is bound at field "x", inside a repeat`},
		{"padded", map[string]string{"word": `["a","b"]`, "card": `"{/word as  w}{w}"`},
			`write {/word as w}`},
		{"selector after a name", map[string]string{"region.json": nameTables()["region.json"], "region.tsv": nameTables()["region.tsv"], "municipality.json": nameTables()["municipality.json"], "municipality.tsv": nameTables()["municipality.tsv"], "card": `"{/region as r}{r.name} {r.municipality[0180].name}"`},
			`read the row without the name, {/region.municipality[0180].name}`},
		{"in a table's format", map[string]string{"word": `["a","b"]`, "t.json": `{"format":"{/word as w}{w}","rows":"t.tsv","key":"k"}`, "t.tsv": "k\nx\ny\n"},
			"a table binds no name"},
		{"in a field's name", map[string]string{"card": `{"format":"{a as b}","a as b":"x"}`},
			`contains " as "`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := map[string]string{}
			for name, body := range c.files {
				if !strings.Contains(name, ".") {
					name += ".json"
				}
				files[name] = body
			}
			_, err := New(WithDataPath(writeFiles(t, files)))
			if c.want == "" {
				if err != nil {
					t.Fatalf("New() = %v, want it to load", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("New() = %v, want an error holding %q", err, c.want)
			}
		})
	}
}

func TestANameBindsAReferenceFromItsFolder(t *testing.T) {
	dir := writeData(t, map[string]string{
		"sv/word": `["a","b","c","d","e","f","g","h"]`,
		"sv/card": `"{.word as w}{w}{w}"`,
	})
	f := newGenerator(t, dir, WithSeed(23))
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "sv.card"); got[0] != got[1] {
			t.Fatalf("sv.card = %q, want one pick", got)
		}
	}
}

func TestANameBindsAReferencePath(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first}","first":["Ada","Bo","Cy","Di"]}`,
		"card":   `"{/person.first as f}{f}|{f}"`,
	})
	f := newGenerator(t, dir, WithSeed(29))
	for i := 0; i < 50; i++ {
		got := strings.Split(fake(t, f, "card"), "|")
		if got[0] != got[1] {
			t.Fatalf("card = %q, want one pick", got)
		}
	}
}

type namedPerson struct {
	First string `fake:"{/person as p}{p.first}"`
	Whole string `fake:"{p}"`
}

func TestStructTagsReadOneName(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy"],"last":["Byron","Ek","Lind"]}`,
	})
	f := newGenerator(t, dir, WithSeed(31))
	for i := 0; i < 50; i++ {
		var p namedPerson
		if err := f.FakeStruct(&p); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(p.Whole, p.First+" ") {
			t.Fatalf("struct = %+v, want one person", p)
		}
	}
}

func TestANameEnteredPastItsCategoryAgreesWithItsWhole(t *testing.T) {
	dir := writeData(t, map[string]string{
		"given":  `["Ada","Bo","Cy","Di","Ed","Flo"]`,
		"person": `{"format":"{/given as g}{g}","first":"{g}"}`,
		"card":   `"{/person as p}{p}|{p.first}"`,
	})
	f := newGenerator(t, dir, WithSeed(37))
	for i := 0; i < 100; i++ {
		got := strings.Split(fake(t, f, "card"), "|")
		if got[0] != got[1] {
			t.Fatalf("card = %q, want {p} and {p.first} to read one pick of g", got)
		}
	}
}

func TestANameKeepsItsPickInARepeatReachedByPath(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word": `["a","b","c","d","e","f","g","h"]`,
		"line": `{"format":"{/word as w}{each}","each":{"format":"{w}","repeat":40,"separator":" "}}`,
		"far":  `"{/line.each}"`,
	})
	f := newGenerator(t, dir, WithSeed(41))
	for _, path := range []string{"line.each", "far"} {
		got := strings.Fields(fake(t, f, path))
		for _, w := range got {
			if w != got[0] {
				t.Fatalf("%s = %q, want one pick on every line", path, got)
			}
		}
	}
}

func TestAHeldPathReadsTheNameItsFieldReads(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"person.json": `{"format":"{first} {last}","rows":"person.tsv"}`,
		"person.tsv":  "first\tlast\nAda\tByron\nBo\tEk\nCy\tLind\n",
		"cat.json":    `{"format":"{/person as n}{field}","field":{"format":"{n.first} {x.y}","x":{"format":"{y}","y":"{n.last}"}}}`,
	})
	f := newGenerator(t, dir, WithSeed(43))
	people := map[string]bool{"Ada Byron": true, "Bo Ek": true, "Cy Lind": true}
	for i := 0; i < 100; i++ {
		if got := fake(t, f, "cat.field"); !people[got] {
			t.Fatalf("cat.field = %q, want one person read through n", got)
		}
	}
}

func TestANameIsATransformOperand(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first}","first":["Åsa","Bo","Cy"]}`,
		"card":   `"{/person as p}{p.first} <{lowercase(ascii(p.first))}>"`,
	})
	f := newGenerator(t, dir, WithSeed(47))
	want := map[string]bool{"Åsa <asa>": true, "Bo <bo>": true, "Cy <cy>": true}
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "card"); !want[got] {
			t.Fatalf("card = %q, want the operand to read the pick", got)
		}
	}
}

func TestOneRenderReadsOneNameAcrossNestedTemplates(t *testing.T) {
	dir := writeData(t, map[string]string{
		"word":  `["a","b","c","d","e","f","g","h"]`,
		"other": `{"format":"{x}","x":["1","2"]}`,
		"cat":   `{"format":"{/word as w}{a}","a":{"format":"{w}|{b}","b":"{w}{/other.x}"}}`,
	})
	f := newGenerator(t, dir, WithSeed(59))
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "cat.a"); got[0] != got[2] {
			t.Fatalf("cat.a = %q, want one pick of w in both templates", got)
		}
	}
}

func TestAFieldBindingReadInsideItsFieldIsRefused(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{x as n}{n}{n}","x":"{n}"}`:                          `name "n" is read inside "x", the field bound to it`,
		`{"format":"{x as n}{n.a}{n.a}","x":{"format":"{a}","a":"{n}"}}`: `name "n" is read inside "x", the field bound to it`,
		`{"format":"{x as n}{y as m}{n}{n}{m}{m}","x":"{m}","y":"{n}"}`:  "cycle",
	} {
		if _, err := f.NewTemplate(src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("NewTemplate(%s) = %v, want it refused naming %s", src, err, want)
		}
	}
}

func TestANameReadAloneIsTheRecordColumnItReads(t *testing.T) {
	dir := writeData(t, map[string]string{
		"src": `{"format":"","score":[null,{"format":"{int(1,9)}","datatype":"integer"}],"code":["1","2"]}`,
		"row": `{"format":"","a":"{/src.score as s}{s}","b":"{s}","c":"{/src as r}{r.score}","d":"{r.code}"}`,
	})
	f := newGenerator(t, dir, WithSeed(3))
	nulls := 0
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()
		if c[0].DataType != DataTypeInteger || c[2].DataType != DataTypeInteger || c[0].Value != c[1].Value || c[0].Null != c[1].Null {
			t.Fatalf("%s: want a, b and c integer columns of src.score, a and b one pick", r.JSON())
		}
		if c[0].Null {
			nulls++
		}
	}
	if nulls == 0 || nulls == 100 {
		t.Errorf("a was null %d times in 100 records, want both outcomes", nulls)
	}
	if _, err := f.NewRecordTemplate(`{"format":"","a":"{/src.score as s}{s}","b":"{s}"}`); err != nil {
		t.Errorf("NewRecordTemplate = %v, want a name-read column accepted inline", err)
	}
}

func TestAFieldBindingReadOutsideWhatItBindsLoads(t *testing.T) {
	if _, err := engine(1).NewTemplate(`{"format":"{x.y as n}{n}{n}{x}","x":{"format":"{z}","y":["a","b"],"z":"{n}"}}`); err != nil {
		t.Errorf("NewTemplate = %v, want a read of n from x.z, outside x.y, accepted", err)
	}
}

func TestANameReadBeforeItsBinderKeepsOnePick(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy"],"last":["Byron","Ek","Lind"]}`,
		"row":    `{"format":"{a}|{b}","a":"{p}","b":"{/person as p}{p.first} {p.last}"}`,
	})
	f := newGenerator(t, dir, WithSeed(2))
	for i := 0; i < 100; i++ {
		if a, b, _ := strings.Cut(fake(t, f, "row"), "|"); a != b {
			t.Fatalf("a = %q, b = %q, want one pick of p", a, b)
		}
	}
}

func TestARowThroughANameHintsAPathThatLoads(t *testing.T) {
	f := newGenerator(t, writeFiles(t, nameTables()), WithSeed(1))
	for tmpl, hint := range map[string]string{"{/region as r}{r[01]}": "{/region[01]}", "{/region as r}{r[01].municipality}": "{/region[01].municipality}"} {
		if _, err := f.NewTemplate(tmpl); err == nil || !strings.Contains(err.Error(), "read the row without the name, "+hint+",") {
			t.Errorf("NewTemplate(%s) = %v, want the hint %s", tmpl, err, hint)
		}
		if _, err := f.NewTemplate(hint); err != nil {
			t.Errorf("NewTemplate(%s), the hint, = %v", hint, err)
		}
	}
	files := nameTables()
	files["card.json"] = `{"format":"{x as r}{r[01]}","x":"{/region}"}`
	want := `a path through name "r" may not select a row; bind the row to a name of its own`
	if _, err := New(WithDataPath(writeFiles(t, files))); err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("New(a field-bound name selecting a row) = %v, want %s", err, want)
	}
}
