package fejkdata

import (
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

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

func TestAClauseDrawsWholeRowsOfATableReadingSeveralColumns(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"car.json": `{"format":"{make} {model}","rows":"car.tsv"}`,
		"car.tsv":  "make\tmodel\nVolvo\tV70\nSaab\t900\nKia\tRio\n",
	}))
	for src, want := range map[string][]string{
		`"{/car where length 7}"`:               {"Kia Rio"},
		`"{/car where not containing 'a'}"`:     {"Volvo V70"},
		`"{/car distinct 2 as c}{c[1]}/{c[2]}"`: {"Volvo V70/Saab 900", "Volvo V70/Kia Rio", "Saab 900/Volvo V70", "Saab 900/Kia Rio", "Kia Rio/Volvo V70", "Kia Rio/Saab 900"},
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
		"boat.json":   `{"format":"{make} {model}","rows":"boat.tsv"}`,
		"boat.tsv":    "make\tmodel\nX\t{/other}\n",
	}))
	if _, err := New(WithSeed(1), onTestDay, WithDataPath(writeFiles(t, map[string]string{
		"pet.json": `{"format":"{name where length 3}","rows":"pet.tsv"}`,
		"pet.tsv":  "name\ncat\n",
	}))); err == nil {
		t.Error("a table's format filtering its own column loaded, want it refused: the row is drawn before the format renders")
	}
	for src, why := range map[string]string{
		`{"format":"{w where length 9}","w":["cat","horse"]}`:                                  "no item passes",
		`{"format":"{n as k}{w where length k}","n":"{int(3,6)}","w":["cat","bird","horse"]}`:  "no item of length 6",
		`{"format":"{n as k}{w where length k}","n":["3","x"],"w":["cat"]}`:                    "a length that can be no whole number",
		`{"format":"{w where length 5}","w":["{lower(5)}","horse"]}`:                           "a generated value",
		`{"format":"{w where length 1}","w":["a",null]}`:                                       "a null",
		`{"format":"{w as x}{x where length 5}","w":["horse"]}`:                                "a name, already drawn",
		`{"format":"{w where length 5 as x}{x.y}","w":["horse"]}`:                              "a field of a filtered draw",
		`{"format":"{a|b where length 1}","a":"x","b":"y"}`:                                    "two arms",
		`{"format":"{w where length k}","w":["cat"]}`:                                          "a length naming no name",
		`{"format":"{w where length k}","k":"3","w":["cat"]}`:                                  "a length reading a field, not a name",
		`{"format":"{n as k}{w where length k}","n":"50","w":["cat"]}`:                         "a length past every item",
		`{"format":"{w where length 3}","w":[{"format":"{a}{b}","a":"x","b":"yz"}]}`:           "an item composed of two reads",
		`{"format":"{w where length 2}","w":{"format":"ab","repeat":2}}`:                       "a repeat",
		`{"format":"{v as n}{w where length 1}","v":"a","w":["{n}"]}`:                          "an item reading a name bound outside it",
		`{"format":"{w where length 1}","w":[{"format":"{v where length 2}","v":["a","bb"]}]}`: "an item drawing through a clause of its own",
		`{"format":"{w where length 3 as  x}{ x}","w":["cat"]}`:                                "a name padded with a space",
		`"{/animal.name where length 3}"`:                                                      "a table cell carrying a token",
		`"{/boat where length 7}"`:                                                             "a row reading a cell that carries a token",
	} {
		if _, err := f.NewTemplate(src); err == nil {
			t.Errorf("NewTemplate(%s) = nil error, want it refused for %s", src, why)
		}
	}
}

func TestAClauseWordRightAfterAsIsTheNameItBinds(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{x as where}{where}","x":"a"}`:                                "a",
		`{"format":"{x as distinct}{distinct}","x":"a"}`:                          "a",
		`{"format":"{w where length 1 as where}{where}","w":["a","bb"]}`:          "a",
		`{"format":"{w distinct 1 as distinct}{distinct[1]}","w":["a"]}`:          "a",
		`{"format":"{w where length 1 distinct 1 as where}{where[1]}","w":["a"]}`: "a",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render(%s) = %q, want %q", src, got, want)
		}
	}
}

// A range is proved at its narrowest, the highest its lower end may print and the lowest its upper
// end may, and the error names the values that narrow it.
func TestWhereNamesTheLengthsThatLeaveNoValue(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}))
	for src, want := range map[string]string{
		`{"format":"{n as k}{w where length k to 6}","n":"{int(3,8)}","w":["cat","bird","horse"]}`:                       "where k is 8, length 8 to 6 holds no length; keep k at or below 6",
		`{"format":"{n as k}{w where length 3 to k}","n":"{int(1,5)}","w":["cat","bird","horse"]}`:                       "where k is 1, length 3 to 1 holds no length; keep k at 3 or above",
		`{"format":"{n as k}{m as j}{w where length k to j}","n":"{int(3,5)}","m":"{int(4,6)}","w":["cat","bird"]}`:      "where k is 5 and j is 4, length 5 to 4 holds no length; keep k at or below j",
		`{"format":"{n as k}{w where length k to 9}","n":"{int(3,6)}","w":["cat","bird","horse"]}`:                       "where k is 6, no value passes the where clause",
		`{"format":"{n as k}{w where length k to 9 distinct 2 joined ','}","n":"{int(3,5)}","w":["cat","bird","horse"]}`: "where k is 5, distinct 2 draws 2 values, and only 1 passes",
	} {
		if _, err := f.NewTemplate(src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("NewTemplate(%s) = %v, want an error holding %q", src, err, want)
		}
	}
}

func TestWhereSaysAQuotedValueHoldsNoApostrophe(t *testing.T) {
	src := `{"format":"{w where not 'O'Brien'}","w":["Ann"]}`
	if _, err := engine(1).NewTemplate(src); err == nil || !strings.Contains(err.Error(), "a quoted value can't hold an apostrophe") {
		t.Errorf("NewTemplate(%s) = %v, want an error saying a quoted value can't hold an apostrophe", src, err)
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

// Rounding in drawIndex may land on a drawn index or one past the span; undrawnNear then takes the
// first undrawn index after it, else the last one before it, inside the span.
func TestUndrawnNearStaysInsideTheSpan(t *testing.T) {
	for _, c := range []struct {
		i, from, to, want int
		drawn             []int
	}{
		{i: 2, from: 0, to: 3, drawn: []int{2}, want: 1},
		{i: 1, from: 0, to: 2, drawn: []int{1}, want: 0},
		{i: 3, from: 0, to: 3, want: 2},
		{i: 1, from: 0, to: 3, drawn: []int{1}, want: 2},
		{i: 3, from: 0, to: 3, drawn: []int{2}, want: 1},
	} {
		if got := undrawnNear(c.i, c.from, c.to, c.drawn); got != c.want {
			t.Errorf("undrawnNear(%d, %d, %d, %v) = %d, want %d", c.i, c.from, c.to, c.drawn, got, c.want)
		}
	}
}

// quotedAllBut lists each character of alphabet but keep, quoted and joined by commas.
func quotedAllBut(alphabet, keep string) string {
	var quoted []string
	for _, c := range alphabet {
		if string(c) != keep {
			quoted = append(quoted, "'"+string(c)+"'")
		}
	}
	return strings.Join(quoted, ",")
}

func TestWhereOnACallDrawsOnlyTheValuesThatPass(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`"{digits(1) where not containing ` + quotedAllBut("0123456789", "9") + `}"`:                                                       {"9"},
		`"{digits(1) where not '3', '33', '' and length 1}"`:                                                                               {"0", "1", "2", "4", "5", "6", "7", "8", "9"},
		`"{hex(1) where not containing 'a','b','c','d','e','f'}"`:                                                                          {"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		`"{lower(1) where not containing ` + quotedAllBut("abcdefghijklmnopqrstuvwxyz", "q") + `}"`:                                        {"q"},
		`"{upper(1) where not containing ` + quotedAllBut("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "Q") + `}"`:                                        {"Q"},
		`"{nanoid(1) where not containing ` + quotedAllBut("_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", "_") + `}"`: {"_"},
	} {
		drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
	}
}

func TestWhereOnACallDrawsEachValueOfItsKindHoldingNoneOfTheStrings(t *testing.T) {
	f := engine(1)
	for src, pattern := range map[string]string{
		`"{digits(2) where not containing '10','11','12','13','14','15','16','17','18','19'}"`: `^[02-9][0-9]$`,
		`"{uuid() where not containing '0', 'f'}"`:                                             `^[1-9a-e]{8}-[1-9a-e]{4}-7[1-9a-e]{3}-[89ab][1-9a-e]{3}-[1-9a-e]{12}$`,
		`"{ulid() where not containing '0', 'Z'}"`:                                             `^[1-7][1-9A-HJKMNP-TV-Y]{25}$`,
		`"{base64(2) where not containing 'A', '/'}"`:                                          `^[B-Za-z0-9+]{2}[EIMQUYcgkosw048]=$`,
		`"{nanoid(4) where not containing '_', '-'}"`:                                          `^[0-9a-zA-Z]{4}$`,
	} {
		re := regexp.MustCompile(pattern)
		for i := 0; i < 400; i++ {
			if got := mustRender(t, f, src); !re.MatchString(got) {
				t.Fatalf("render(%s) = %q, want it to match %s", src, got, pattern)
			}
		}
	}
}

func TestWhereReadsTheStringsAName(t *testing.T) {
	f := engine(1)
	for src, want := range map[string][]string{
		`{"format":"{v as t}{t}:{w where not t}","v":["cat","dog"],"w":["cat","dog","cow"]}`:                                {"cat:dog", "cat:cow", "dog:cat", "dog:cow"},
		`{"format":"{v as t}{t}:{w where not containing t}","v":["a","o"],"w":["cat","dog","bee"]}`:                         {"a:dog", "a:bee", "o:cat", "o:bee"},
		`{"format":"{v as t}{t}:{digits(1) where not t}","v":["3","7"]}`:                                                    {"3:0", "3:1", "3:2", "3:4", "3:5", "3:6", "3:7", "3:8", "3:9", "7:0", "7:1", "7:2", "7:3", "7:4", "7:5", "7:6", "7:8", "7:9"},
		`{"format":"{v as t}{w where not containing t}","v":"","w":["a","b"]}`:                                              {"a", "b"},
		`{"format":"{v as t}{digits(1) where not containing t and not containing '1','2','3','4','5','6','7','8'}","v":""}`: {"0", "9"},
	} {
		drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
	}
	var want []string
	for _, m := range []string{"12", "21"} {
		for _, d := range "03456789" {
			want = append(want, m+":"+string(d))
		}
	}
	src := `{"format":"{v distinct 2 as m}{m[1]}{m[2]}:{digits(1) where not containing m[1], m[2]}","v":["1","2"]}`
	drawsExactly(t, src, func() string { return mustRender(t, f, src) }, want...)
}

func TestWherePrintsTheFallbackWhereNoValuePasses(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{v as t}{w where not t else 'none'}","v":"cat","w":["cat"]}`:                                "none",
		`{"format":"{v as t}{u as f}{w where not containing t else f}","v":"a","u":"fb","w":["cat","bar"]}`:     "fb",
		`{"format":"{v as t}{uuid() where not containing t else 'none'}","v":"-"}`:                              "none",
		`{"format":"{v as t}{uuid() where not containing t else '-'}","v":"-"}`:                                 "-",
		`{"format":"{v as t}{w where not t else '-' distinct 2 joined ','}","v":"cat","w":["cat","dog"]}`:       "dog,-",
		`{"format":"{v as t}{w where not t else '-' distinct 2 as m}{m[2]}{m[1]}","v":"cat","w":["cat","dog"]}`: "-dog",
	} {
		for i := 0; i < 20; i++ {
			if got := mustRender(t, f, src); got != want {
				t.Fatalf("render(%s) = %q, want %q", src, got, want)
			}
		}
	}
}

func TestWherePrintsADefaultWhereNoValuePassesAndNoFallbackIsNamed(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"<{v as t}{w where not t}>","v":"cat","w":["cat"]}`:                                                                                                  "<>",
		`{"format":"{v as t}{w where not t distinct 2 joined ','}","v":"cat","w":["cat","dog"]}`:                                                                        "dog,",
		`{"format":"{v as t}{uuid() where not containing t}","v":"-"}`:                                                                                                  "00000000-0000-0000-0000-000000000000",
		`{"format":"{v as t}{ulid() where not containing t,` + quotedAllBut("01234567", "0") + `}","v":"0"}`:                                                            "00000000000000000000000000",
		`{"format":"{v as t}{digits(3) where not containing t,` + quotedAllBut("0123456789", "0") + `}","v":"0"}`:                                                       "000",
		`{"format":"{v as t}{hex(2) where not containing t,` + quotedAllBut("0123456789abcdef", "0") + `}","v":"0"}`:                                                    "00",
		`{"format":"{v as t}{nanoid(2) where not containing t,` + quotedAllBut("_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", "0") + `}","v":"0"}`: "00",
		`{"format":"{v as t}{lower(2) where not containing t,` + quotedAllBut("abcdefghijklmnopqrstuvwxyz", "a") + `}","v":"a"}`:                                        "aa",
		`{"format":"{v as t}{upper(2) where not containing t,` + quotedAllBut("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "A") + `}","v":"A"}`:                                        "AA",
		`{"format":"{v as t}{base64(1) where not containing t}","v":"="}`:                                                                                               "AA==",
		`{"format":"{v as t}{base64(2) where not containing t}","v":"="}`:                                                                                               "AAA=",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render(%s) = %q, want %q", src, got, want)
		}
	}
}

func TestWhereOnACallRefusesAtLoad(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"x": `"x"`}))
	for src, want := range map[string]string{
		`"{int(1,9) where not '3'}"`: "{base64()}, {digits()}, {hex()}, {lower()}, {nanoid()}, {ulid()}, {upper()} or {uuid()}",
		`"{digits(1) where not containing ` + quotedAllBut("0123456789", "") + `}"`: "no value passes the where clause",
		`"{uuid() where not containing '-'}"`:                                       "no value passes the where clause",
		`"{digits(2) where length 3}"`:                                              "no value passes the where clause",
		`{"format":"{n as k}{digits(2) where length k}","n":"{int(2,3)}"}`:          "where k is 3, no value passes the where clause",
		`{"format":"{digits(1) where not t}","t":"3"}`:                              "{t as n}",
		`{"format":"{w where not containing t}","t":"3","w":["a"]}`:                 "{t as n}",
		`{"format":"{w where not 'a' else f}","f":"x","w":["a","b"]}`:               "{f as n}",
		`{"format":"{w where not t}","w":["a"]}`:                                    "or quote the text, as 't'",
	} {
		if _, err := f.NewTemplate(src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("NewTemplate(%s) = %v, want an error holding %q", src, err, want)
		}
	}
}
