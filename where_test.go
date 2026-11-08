package fejkdata

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// drawsExactly renders render 400 times and fails unless every value is one of want and
// each of want shows up.
func drawsExactly(t *testing.T, label string, render func() string, want ...string) {
	t.Helper()
	seen := map[string]int{}
	for i := 0; i < 400; i++ {
		got := render()
		if !slices.Contains(want, got) {
			t.Fatalf("%s = %q, want one of %q", label, got, want)
		}
		seen[got]++
	}
	for _, w := range want {
		if seen[w] == 0 {
			t.Errorf("%s never drew %q in 400 renders: %v", label, w, seen)
		}
	}
}

func TestWhereDrawsOnlyTheItemsThatPass(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`{"format":"{w where length 5}","w":["cat","horse","mouse","elephant"]}`:                            {"horse", "mouse"},
		`{"format":"{w where length 4 to 6}","w":["cat","bird","horse","turtle","giraffes"]}`:               {"bird", "horse", "turtle"},
		`{"format":"{w where not 'cat', 'dog'}","w":["cat","dog","catalog"]}`:                               {"catalog"},
		`{"format":"{w where not containing 'cat'}","w":["cat","concatenate","dog"]}`:                       {"dog"},
		`{"format":"{w where length 3 and not 'cat'}","w":["cat","dog","horse"]}`:                           {"dog"},
		`{"format":"{w where length 3}","w":["åsa","anna"]}`:                                                {"åsa"},
		`{"format":"<{w where length 2}>","w":[{"format":"x{v}","v":["a","bb"]},"c"]}`:                      {"<xa>"},
		`{"format":"{w where length 1}","w":[{"format":"{v as n}{n}","v":["a","bb"]},"cc"]}`:                {"a"},
		`{"format":"{w.v where length 2}","w":[{"format":"{v}","v":["a","bb"]},{"format":"{v}","v":"cc"}]}`: {"bb", "cc"},
	} {
		drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
	}
}

func TestWhereKeepsTheWeightsOfTheItemsThatPass(t *testing.T) {
	f := engine(3)
	src := `{"format":"{w where length 2}","w":[{"format":"a1","weight":3},{"format":"b1","weight":1},{"format":"ccc","weight":100}]}`
	a := 0
	for i := 0; i < 4000; i++ {
		switch got := mustRender(t, f, src); got {
		case "a1":
			a++
		case "b1":
		default:
			t.Fatalf("render = %q, want a1 or b1", got)
		}
	}
	if a < 2800 || a > 3200 {
		t.Errorf("a1 drew %d of 4000, want about 3000: weight 3 against b1's 1", a)
	}
}

func TestWhereBindsTheFilteredDraw(t *testing.T) {
	f := engine(1)
	src := `{"format":"{w where length 5 as x}{x} {uppercase(x)}","w":["cat","horse","mouse"]}`
	drawsExactly(t, src, func() string { return mustRender(t, f, src) }, "horse HORSE", "mouse MOUSE")
}

func TestWhereReadsALengthFromAName(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}))
	for src, want := range map[string][]string{
		`{"format":"{n as k}{k}={w where length k}","n":"{int(3,4)}","w":["cat","bird","horse"]}`:            {"3=cat", "4=bird"},
		`{"format":"{n as k}{k}={w where length 3 to k}","n":["4","5"],"w":["cat","bird","horse","turtle"]}`: {"4=cat", "4=bird", "5=cat", "5=bird", "5=horse"},
		`{"format":"{n as k}{w where length k as v}{k}={v}","n":"{int(3,4)}","w":["cat","bird"]}`:            {"3=cat", "4=bird"},
	} {
		tp, err := f.NewTemplate(src)
		if err != nil {
			t.Fatalf("NewTemplate(%s): %v", src, err)
		}
		drawsExactly(t, src, tp.Fake, want...)
	}
}

func TestWhereFiltersATable(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"animal.json": `{"format":"{name}","rows":"animal.tsv","key":"code"}`,
		"animal.tsv":  "code\tname\nC\tcat\nH\thorse\nM\tmouse\n",
	}))
	for src, want := range map[string][]string{
		`"{/animal.name where length 5}"`:     {"horse", "mouse"},
		`"{/animal where length 3}"`:          {"cat"},
		`"{/animal.code where not 'C', 'H'}"`: {"M"},
	} {
		tp, err := f.NewTemplate(src)
		if err != nil {
			t.Fatalf("NewTemplate(%s): %v", src, err)
		}
		drawsExactly(t, src, tp.Fake, want...)
	}
}

func TestWhereRefusesAtLoad(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"animal.json": `{"format":"{name}","rows":"animal.tsv"}`,
		"animal.tsv":  "name\ncat\n{/other}\n",
		"other.json":  `"horse"`,
	}))
	if _, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, map[string]string{
		"pet.json": `{"format":"{name where length 3}","rows":"pet.tsv"}`,
		"pet.tsv":  "name\ncat\n",
	}))); err == nil {
		t.Error("a table's format filtering its own column loaded, want it refused: the row is drawn before the format renders")
	}
	for src, why := range map[string]string{
		`{"format":"{w where length 9}","w":["cat","horse"]}`:                                 "no item passes",
		`{"format":"{n as k}{w where length k}","n":"{int(3,6)}","w":["cat","bird","horse"]}`: "no item of length 6",
		`{"format":"{n as k}{w where length k}","n":["3","x"],"w":["cat"]}`:                   "a length that can be no whole number",
		`{"format":"{w where length 5}","w":["{lower(5)}","horse"]}`:                          "a generated value",
		`{"format":"{w where length 1}","w":["a",null]}`:                                      "a null",
		`{"format":"{w as x}{x where length 5}","w":["horse"]}`:                               "a name, already drawn",
		`{"format":"{w where length 5 as x}{x.y}","w":["horse"]}`:                             "a field of a filtered draw",
		`{"format":"{a|b where length 1}","a":"x","b":"y"}`:                                   "two arms",
		`{"format":"{w where length k}","w":["cat"]}`:                                         "a length naming no name",
		`{"format":"{w where length k}","k":"3","w":["cat"]}`:                                 "a length reading a field, not a name",
		`{"format":"{n as k}{w where length k}","n":"50","w":["cat"]}`:                        "a length past every item",
		`{"format":"{w where length 3}","w":[{"format":"{a}{b}","a":"x","b":"yz"}]}`:          "an item composed of two reads",
		`{"format":"{w where length 2}","w":{"format":"ab","repeat":2}}`:                      "a repeat",
		`{"format":"{v as n}{w where length 1}","v":"a","w":["{n}"]}`:                         "an item reading a name bound outside it",
		`"{/animal.name where length 3}"`:                                                     "a table cell carrying a token",
	} {
		if _, err := f.NewTemplate(src); err == nil {
			t.Errorf("NewTemplate(%s) = nil error, want it refused for %s", src, why)
		}
	}
}

func TestWhereFiltersTheShippedWords(t *testing.T) {
	t.Parallel()
	f := shipped(t)
	for _, src := range []string{"{/en_US.word where length 5}", "{/en_US.word.verb where length 5}", "{/en_US.word where length 5 as w}{uppercase(w)}"} {
		for i := 0; i < 50; i++ {
			got := tmpl(t, f, src)
			if utf8.RuneCountInString(got) != 5 || strings.ToLower(got) != got && strings.ToUpper(got) != got {
				t.Fatalf("FakeTemplate(%q) = %q, want a word of five letters", src, got)
			}
		}
	}
}
