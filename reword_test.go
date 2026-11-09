package fejkdata

import (
	"regexp"
	"strings"
	"testing"
)

func TestRewordReplacesEachWordKeepingPunctuationDigitsAndCase(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{reword(t, w)}","t":"Cat, 42 HORSE-Horse sat!","w":["abc","defgh"]}`: "Abc, 42 DEFGH-Defgh abc!",
		`{"format":"{reword(t,w)}","t":"McDonald","w":["abcdefgh"]}`:                     "AbCdefgh",
		`{"format":"{reword(t,w)}","t":"Åsa","w":["abc","defg"]}`:                        "Abc",
		`{"format":"{reword(t,w)}","t":"PROJ-123 a1b2","w":["abcd","x"]}`:                "ABCD-123 x1x2",
		`{"format":"{reword(t,w)}","t":"","w":["abc"]}`:                                  "",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render of %s = %q, want %q", src, got, want)
		}
	}
}

func TestRewordDrawsTheNearestLengthTheWordsHold(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{reword(t,w)}","t":"HI hi Hi","w":["abc","abcdefgh"]}`:         "ABC abc Abc",
		`{"format":"{reword(t,w)}","t":"Hello","w":["ab","abcdef"]}`:               "Abcdef",
		`{"format":"{reword(t,w)}","t":"abcdefghijkl","w":["ab","abcdefgh"]}`:      "abcdefgh",
		`{"format":"{reword(t,w)}","t":"abc","w":["ab","a-b-c","abcdef","12345"]}`: "ab",
		`{"format":"{reword(t,w)}","t":"abc","w":["ab","abcd"]}`:                   "ab",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render of %s = %q, want %q", src, got, want)
		}
	}
}

func TestRewordKeepsOneFakePerWordWithinARender(t *testing.T) {
	f := engine(1)
	src := `{"format":"{reword(t,w)}|{reword(t,w)}","t":"cat Cat CAT dog","w":["abc","xyz"]}`
	seen := map[string]bool{}
	for i := 0; i < 400; i++ {
		got := mustRender(t, f, src)
		halves := strings.Split(got, "|")
		words := strings.Fields(halves[0])
		if halves[0] != halves[1] || len(words) != 4 || strings.ToLower(words[1]) != words[0] || strings.ToUpper(words[2]) != words[2] || strings.ToLower(words[2]) != words[0] {
			t.Fatalf("render = %q, want cat, Cat and CAT one fake in their own case, the same in both calls", got)
		}
		seen[words[0]+" "+words[3]] = true
	}
	if len(seen) != 4 {
		t.Errorf("drew %v over 400 renders, want cat and dog each drawn afresh per render: all 4 pairs", seen)
	}
}

func TestRewordKeepsOneFakePerWordAcrossATemplateAndTheCategoryItReads(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"w.json":      `["cat","dog","pig","owl","ant","bee","elk","yak"]`,
		"ticket.json": `{"format":"{reword(u,/w)}","u":"Ann"}`,
	}))
	tp, err := f.NewTemplate(`{"format":"{reword(t,/w)}|{/ticket}","t":"Ann"}`)
	if err != nil {
		t.Fatalf("NewTemplate: %v", err)
	}
	for i := 0; i < 20; i++ {
		got := tp.Fake()
		if halves := strings.Split(got, "|"); halves[0] != halves[1] {
			t.Fatalf("render = %q, want Ann one fake in the template and in the category it reads", got)
		}
	}
}

func TestRewordDrawsWithTheWordsWeights(t *testing.T) {
	f := engine(3)
	src := `{"format":"{reword(t,w)}","t":"cat","w":[{"format":"abc","weight":3},{"format":"xyz","weight":1}]}`
	a := 0
	for i := 0; i < 4000; i++ {
		if mustRender(t, f, src) == "abc" {
			a++
		}
	}
	if a < 2800 || a > 3200 {
		t.Errorf("abc drew %d of 4000, want about 3000: weight 3 against xyz's 1", a)
	}
}

func TestRewordRefusesAtLoad(t *testing.T) {
	f := engine(1)
	for src, why := range map[string]string{
		`{"format":"{reword(t)}","t":"cat"}`:                          "one argument",
		`{"format":"{reword(t,w)}","t":"cat","w":["{uuid()}"]}`:       "words that are not fixed text",
		`{"format":"{reword(t,w)}","t":"cat","w":["a-b","12",""]}`:    "no word of letters only",
		`{"format":"{w as n}{reword(t,n)}","t":"cat","w":["abc"]}`:    "words that are a name's one pick",
		`{"format":"{reword(t,w)}","t":"cat","w":["abc",null]}`:       "a null among the words",
		`{"format":"{reword(t,lower(3))}","t":"cat"}`:                 "words from a call",
		`{"format":"{reword(uuid(),w)}","w":["abc"]}`:                 "text from a call",
		`{"format":"{reword(t,w)}","t":"cat"}`:                        "words naming no field",
		`{"format":"{uppercase(reword(t,w))}","t":"cat","w":["abc"]}`: "reword inside a transform",
	} {
		if _, err := f.NewTemplate(src); err == nil {
			t.Errorf("NewTemplate(%s) = nil error, want it refused for %s", src, why)
		}
	}
}

func TestRewordDrawsTheShippedWords(t *testing.T) {
	t.Parallel()
	f := shipped(t)
	tp, err := f.NewTemplate(`{"format":"{reword(t, /en_US.word)}","t":"Hello, World 2024!"}`)
	if err != nil {
		t.Fatal(err)
	}
	shape := regexp.MustCompile(`^\p{Lu}\p{Ll}{4}, \p{Lu}\p{Ll}{4} 2024!$`)
	for i := 0; i < 50; i++ {
		if got := tp.Fake(); !shape.MatchString(got) {
			t.Fatalf("Fake() = %q, want two capitalized five-letter words around the kept punctuation and digits", got)
		}
	}
}
