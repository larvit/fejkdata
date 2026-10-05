package fejkdata

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// swedishPlaces is a two-variant sibling whose variants pair a locality with the
// postal-code prefix that really belongs to it.
const swedishPlaces = `{"format":"%s","place":[{"format":"{locality}","locality":"Stockholm","postal-code":"1{digits(2)} {digits(2)}"},{"format":"{locality}","locality":"Tranås","postal-code":"573 {digits(2)}"}]}`

// agree reports whether a rendered "postcode locality" pair is a real pairing.
var agree = regexp.MustCompile(`^(1[0-9]{2} [0-9]{2} Stockholm|573 [0-9]{2} Tranås)$`)

func places(format string) string {
	return strings.Replace(swedishPlaces, "%s", format, 1)
}

// varies renders src until two renders of it satisfy differ, failing after 400 tries.
func varies(t *testing.T, f *Generator, src string, differ func(got string) bool) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if differ(mustRender(t, f, src)) {
			return
		}
	}
	t.Fatalf("400 renders of %s never drew apart, want each {…} a pick of its own", src)
}

func TestEveryTokenDrawsAfresh(t *testing.T) {
	f := engine(3)
	pair := func(got string) bool { p := strings.Split(got, "|"); return p[0] != p[1] }
	varies(t, f, `{"format":"{w}|{w}","w":["a","b","c","d","e"]}`, pair)
	varies(t, f, `{"format":"{p.first}|{p.first}","p":{"format":"{first}","first":["Anna","Astrid","Elin","Karin"]}}`, pair)
	varies(t, f, `{"format":"{w}|{uppercase(w)}","w":["a","b","c","d","e"]}`, func(got string) bool {
		p := strings.Split(got, "|")
		return strings.ToUpper(p[0]) != p[1]
	})
	varies(t, f, `{"format":"{n}|{calc(n * 1)}","n":["1","2","3","4"]}`, pair)
	varies(t, f, places("{place.postal-code} {place.locality}"), func(got string) bool { return !agree.MatchString(got) })
}

func TestEveryReferenceDrawsAfresh(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"person": `{"format":"{first} {last}","first":["Ada","Bo","Cy","Di"],"last":["Byron","Ek","Lind","Ros"]}`,
	}), WithSeed(4))
	for _, src := range []string{
		`"{/person.first}|{/person.first}"`,
		`"{/person}|{/person.first} {/person.last}"`,
		`{"format":"{a}|{b}","a":"{/person.last}","b":"{/person.last}"}`,
	} {
		differ := false
		for i := 0; i < 400 && !differ; i++ {
			p := strings.Split(fakeTemplate(t, f, src), "|")
			differ = p[0] != p[1]
		}
		if !differ {
			t.Errorf("400 renders of %s never drew apart, want each reference a pick of its own", src)
		}
	}
}

func TestANamedFieldCorrelatesItsReads(t *testing.T) {
	f := engine(1)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got := mustRender(t, f, places("{place as p}{p.postal-code} {p.locality}"))
		if !agree.MatchString(got) {
			t.Fatalf("draw %d = %q, want a postcode and locality that pair", i, got)
		}
		seen[got[len(got)-4:]] = true
	}
	if len(seen) < 2 {
		t.Fatalf("500 draws produced only %v, want both localities", seen)
	}
}

func TestANamedFieldReadsItsVariantsWhole(t *testing.T) {
	// One variant holds addr as a plain template where another holds a choice; the name
	// keeps the variant drawn at every level its reads pass.
	f := engine(13)
	tmpl := `{"format":"{p as q}{q.addr.city}/{q.addr.zip}","p":[
		{"format":"{addr}","addr":{"format":"{city}","city":"Kiruna","zip":"98100"}},
		{"format":"{addr}","addr":[
			{"format":"{city}","city":"Malmö","zip":"21100"},
			{"format":"{city}","city":"Lund","zip":"22100"}]}]}`
	for i := 0; i < 400; i++ {
		got := mustRender(t, f, tmpl)
		if got != "Kiruna/98100" && got != "Malmö/21100" && got != "Lund/22100" {
			t.Fatalf("draw %d = %q, want a city with its own zip", i, got)
		}
	}
}

func TestANamedFieldInARepeatPicksEachLine(t *testing.T) {
	f := engine(5)
	tmpl := `{"format":"{lines}","lines":` + strings.Replace(places("{place as p}{p.postal-code} {p.locality}"), `"format":`, `"repeat":6,"separator":"\n","format":`, 1) + `}`
	varied := false
	for i := 0; i < 40; i++ {
		lines := strings.Split(mustRender(t, f, tmpl), "\n")
		if len(lines) != 6 {
			t.Fatalf("repeat 6 produced %d lines", len(lines))
		}
		for _, line := range lines {
			if !agree.MatchString(line) {
				t.Fatalf("line %q does not pair", line)
			}
		}
		varied = varied || lines[0] != lines[1] || lines[1] != lines[2]
	}
	if !varied {
		t.Fatal("40 renders of repeat 6 never varied within a render, want a pick per line")
	}
}

func TestAFieldBindingIsRefusedWhereNoFieldIs(t *testing.T) {
	for src, want := range map[string]string{
		`{"format":"{nope as p}{p}{p}","place":"x"}`:                            `no field "nope"`,
		`{"format":"{place as p}{p.x}","place":{"format":"{x}","x":["a","b"]}}`: `write {place.x} where it is read`,
		`{"format":"{place as p}{p}","place":["x","y"]}`:                        `write {place} where it is read`,
		`{"format":"{n}|{x}","x":{"format":"{a as n}{a}","a":["1","2"]}}`:       `write {x.a} where it is read`,
		`{"format":"{place as p}{calc(p * 2)}","place":["1","2"]}`:              `write place where it is read`,
		`{"format":"{place as place}{place}{place}","place":["x","y"]}`:         `name "place" is a field of the root template too`,
	} {
		if _, err := linked(t, src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s = %v, want it refused naming %s", src, err, want)
		}
	}
}

func TestCycleReachedOnlyByAPathTokenIsRejected(t *testing.T) {
	// A path token renders what it lands on, not the level it started from, so the
	// cycle walk has to follow it there.
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"a": `{"format":"{p.x}","p":{"format":"static","x":"{/b}"}}`,
		"b": `"{/a}"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "reference cycle") {
		t.Fatalf("New = %v, want the cycle through {p.x} rejected", err)
	}
}

func TestACycleInALaterVariantIsRejected(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"cat": `{"format":"{p.x}","p":[{"format":"h","x":"safe"},{"format":"h","x":"{/hop}"}]}`,
		"hop": `"{/cat}"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "reference cycle") {
		t.Fatalf("New = %v, want the cycle in the later variant rejected", err)
	}
}

func TestCycleThroughAPathTokenIsRejected(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"a": `{"format":"{p.x}","p":{"format":"{x}","x":"{/b}"}}`,
		"b": `"{/a}"`,
	})))
	if err == nil || !strings.Contains(err.Error(), "reference cycle") {
		t.Fatalf("New = %v, want the cycle through {p.x} rejected", err)
	}
}

func TestADeepDiamondChainLoads(t *testing.T) {
	// A diamond chain is 2^n routes through n nodes, so every load walk must visit each
	// node once, not once per route.
	files := map[string]string{"l0": `"x"`}
	for i := 1; i <= 30; i++ {
		files[fmt.Sprintf("l%d", i)] = fmt.Sprintf(
			`{"format":"{a}{b}","a":"{/l%d}","b":"{/l%d}"}`, i-1, i-1)
	}
	files["thing"] = `{"format":"{p.first} {/l30}","p":{"format":"x","first":["A","B"]}}`
	if _, err := New(WithoutShippedData(), WithDataPath(writeData(t, files))); err != nil {
		t.Fatalf("New = %v, want a deep diamond chain to load", err)
	}
}

func TestNestedChoiceDrawsOneVariant(t *testing.T) {
	// A choice item may itself be a choice, so a path unwraps until it reaches a value.
	f := engine(21)
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[mustRender(t, f, `{"format":"{p.x}","p":[{"format":"{x}","x":"1"},{"format":"{x}","x":"2"}]}`)] = true
	}
	if !seen["1"] || !seen["2"] || len(seen) != 2 {
		t.Fatalf("200 draws produced %v, want 1 and 2", seen)
	}
}

func TestRepeatingLevelIsNamedInTheError(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"cat": `{"format":"[{p.a.b}]","p":{"format":"{a}","a":{"format":"{b}","repeat":3,"separator":",","b":"z"}}}`,
	})))
	if err == nil || !strings.Contains(err.Error(), `"p.a"`) {
		t.Fatalf("New = %v, want it to name the level p.a that carries the repeat", err)
	}
}

func TestPathIntoARepeatingLevelIsRejected(t *testing.T) {
	// A path reads one level's draw, so it can never apply that level's repeat, directly or
	// behind a choice.
	for _, cat := range []string{
		`{"format":"[{p.a}]","p":{"format":"{a}","repeat":3,"separator":",","a":"z"}}`,
		`{"format":"[{p.a}]","p":[{"format":"{a}","repeat":3,"separator":",","a":"x"},{"format":"{a}","a":"y"}]}`,
	} {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": cat})))
		if err == nil || !strings.Contains(err.Error(), "repeat") {
			t.Errorf("New(%s) = %v, want a path into a repeating level rejected", cat, err)
		}
	}
}

func TestPathIntoAPlainTemplateNamesTheMissingField(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{
		"cat": `{"format":"{a.nope}","a":{"format":"x","b":"1"}}`,
	})))
	if err == nil || !strings.Contains(err.Error(), `no field "nope"`) {
		t.Fatalf("New = %v, want it to name the missing field", err)
	}
}

func TestADrawIsSeedStable(t *testing.T) {
	tmpl := places("{place as p}{p.postal-code} {p.locality}")
	if a, b := mustRender(t, engine(42), tmpl), mustRender(t, engine(42), tmpl); a != b {
		t.Fatalf("same seed gave %q and %q, want an identical draw", a, b)
	}
}

func TestDottedTokenErrors(t *testing.T) {
	rejected := map[string]struct {
		file string
		want string
	}{
		"unknown head": {
			`{"format":"{nope.x}","place":{"format":"{v}","v":"A"}}`,
			`no field "nope"`,
		},
		"unknown tail": {
			`{"format":"{place.nope}","place":[{"format":"{v}","v":"A"},{"format":"{v}","v":"B"}]}`,
			`"nope"`,
		},
		"tail only some variants carry": {
			`{"format":"{place.locality}","place":[{"format":"{locality}","locality":"A"},"x"]}`,
			`locality`,
		},
		"path into a literal": {
			`{"format":"{x.y}","x":"plain"}`,
			`"y"`,
		},
		"dotted field key": {
			`{"format":"[{a.b}]","a.b":"V"}`,
			`contains "."`,
		},
	}
	for name, c := range rejected {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": c.file})))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: New = %v, want it to mention %q", name, err, c.want)
		}
	}
}

func TestEmptyPathSegmentIsRejected(t *testing.T) {
	for name, file := range map[string]string{
		"trailing dot": `{"format":"[{a.}]","a":"x"}`,
		"triple dot":   `{"format":"[{a...b}]","a":"x"}`,
	} {
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, map[string]string{"cat": file})))
		if err == nil || !strings.Contains(err.Error(), "empty segment") {
			t.Errorf("%s: New = %v, want it to name the empty segment", name, err)
		}
	}
}

func TestASubFieldIsReachableByFake(t *testing.T) {
	dir := writeData(t, map[string]string{"address": places("{place as p}{p.postal-code} {p.locality}")})
	f, err := New(WithoutShippedData(), WithDataPath(dir), WithSeed(9))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	for _, path := range []string{"address", "address.place", "address.place.locality", "address.place.postal-code"} {
		if _, err := f.Fake(path); err != nil {
			t.Errorf("Fake(%q) = %v, want it to render", path, err)
		}
	}
}

func TestAFieldBoundWholeIsItsPickAsAColumn(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"row": `{"format":"{place as p}{p.locality}","place":[{"format":"{locality}","locality":"Stockholm","zip":"1"},{"format":"{locality}","locality":"Tranås","zip":"5"}],"zip":"{p.zip}"}`,
	}), WithSeed(1))
	zips := map[string]string{"Stockholm": "1", "Tranås": "5"}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		r, err := f.FakeRecord("row")
		if err != nil {
			t.Fatal(err)
		}
		c := r.Columns()
		if zips[c[0].Value] != c[1].Value {
			t.Fatalf("%s: want the place column and the zip read through p one place", r.JSON())
		}
		seen[c[0].Value] = true
	}
	if len(seen) != 2 {
		t.Errorf("place drew only %v in 100 records, want both", seen)
	}
}
