package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	enus "github.com/larvit/fejkdata/data/en_US"
)

func englishWordForms(t *testing.T, table string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("data/en_US/" + table + ".tsv")
	if err != nil {
		t.Fatal(err)
	}
	forms := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n")[1:] {
		word, form, _ := strings.Cut(line, "\t")
		forms[word] = form
	}
	return forms
}

func TestEnglishWordHoldsWordNetWordsPerPartOfSpeech(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ table, word, form string }{
		{"word-noun", "house", ""},
		{"word-noun", "river", ""},
		{"word-verb", "run", "runs"},
		{"word-verb", "carry", "carries"},
		{"word-verb", "go", "goes"},
		{"word-verb", "pass", "passes"},
		{"word-verb", "have", "has"},
		{"word-verb", "be", "is"},
		{"word-adj", "big", ""},
		{"word-adv", "often", ""},
	} {
		if got, ok := englishWordForms(t, c.table)[c.word]; !ok || got != c.form {
			t.Errorf("%s holds %q with form %q (held: %v), want %q", c.table, c.word, got, ok, c.form)
		}
	}
	for table, least := range map[string]int{"word-noun": 40000, "word-verb": 5000, "word-adj": 10000, "word-adv": 2500} {
		if n := len(englishWordForms(t, table)); n < least {
			t.Errorf("%s holds %d words, want WordNet's, %d or more", table, n, least)
		}
	}
}

func TestEnglishWordTablesHoldLowerCaseLettersAlone(t *testing.T) {
	t.Parallel()
	letters := regexp.MustCompile(`^[a-z]+$`)
	for _, table := range []string{"word-noun", "word-verb", "word-adj", "word-adv"} {
		for word, form := range englishWordForms(t, table) {
			if !letters.MatchString(word) || (form != "" && !letters.MatchString(form)) {
				t.Fatalf("%s holds %q with form %q, want lower-case letters alone", table, word, form)
			}
		}
	}
}

func TestEnglishWordPathsDrawTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	en := shipped(t, WithSeed(5))
	nouns, verbs, adjs, advs := englishWordForms(t, "word-noun"), englishWordForms(t, "word-verb"), englishWordForms(t, "word-adj"), englishWordForms(t, "word-adv")
	present := map[string]bool{}
	for _, form := range verbs {
		present[form] = true
	}
	seen := map[string]int{}
	for i := 0; i < 2000; i++ {
		for path, held := range map[string]func(string) bool{
			"en_US.word.noun":         func(w string) bool { _, ok := nouns[w]; return ok },
			"en_US.word.verb":         func(w string) bool { _, ok := verbs[w]; return ok },
			"en_US.word.verb.present": func(w string) bool { return present[w] },
			"en_US.word.adj":          func(w string) bool { _, ok := adjs[w]; return ok },
			"en_US.word.adv":          func(w string) bool { _, ok := advs[w]; return ok },
		} {
			if got := fake(t, en, path); !held(got) {
				t.Fatalf("%s = %q, which its table does not hold", path, got)
			}
		}
		w := fake(t, en, "en_US.word")
		switch _, n := nouns[w]; {
		case n:
			seen["noun"]++
		default:
			seen["other"]++
		}
	}
	if seen["noun"] == 0 || seen["other"] == 0 {
		t.Errorf("en_US.word drew %v, want nouns and other parts of speech", seen)
	}
}

func TestEnglishBoundVerbReadsItsOwnPresent(t *testing.T) {
	t.Parallel()
	en := shipped(t, WithSeed(6))
	forms := englishWordForms(t, "word-verb")
	for i := 0; i < 200; i++ {
		word, form, _ := strings.Cut(fakeTemplate(t, en, "{/en_US.word.verb as w}{w}|{w.present}"), "|")
		if forms[word] != form {
			t.Fatalf("en_US.word.verb %q printed present %q, want its own, %q", word, form, forms[word])
		}
	}
}

func TestEnglishWordWeighsAPartOfSpeechByItsRows(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("data/en_US/word.json")
	if err != nil {
		t.Fatal(err)
	}
	var word struct {
		Any []struct {
			Format string `json:"format"`
			Weight int    `json:"weight"`
		} `json:"any"`
	}
	if err := json.Unmarshal(raw, &word); err != nil {
		t.Fatal(err)
	}
	if len(word.Any) != 4 {
		t.Fatalf("any holds %d parts of speech, want 4", len(word.Any))
	}
	for _, part := range word.Any {
		table := strings.Trim(part.Format, "{.}")
		if rows := len(englishWordForms(t, table)); part.Weight != rows {
			t.Errorf("%s weighs %d, want its %d rows", table, part.Weight, rows)
		}
	}
}

func TestEnglishWordDrawsFarMoreDistinctWordsThanItsHandWrittenList(t *testing.T) {
	t.Parallel()
	en := shipped(t, WithSeed(8))
	distinct := map[string]bool{}
	for i := 0; i < 5000; i++ {
		distinct[fake(t, en, "en_US.word")] = true
	}
	if len(distinct) < 3000 {
		t.Errorf("5000 draws of en_US.word gave %d distinct words, want a dictionary's spread, not 271", len(distinct))
	}
}

func TestEnglishSentenceDrawsTheWordsItsPatternNeeds(t *testing.T) {
	t.Parallel()
	en := shipped(t, WithSeed(7))
	set := func(table string) map[string]bool {
		held := map[string]bool{}
		for word := range englishWordForms(t, table) {
			held[word] = true
		}
		return held
	}
	adj, noun, adv, verb := set("word-adj"), set("word-noun"), set("word-adv"), set("word-verb")
	present := map[string]bool{}
	for _, form := range englishWordForms(t, "word-verb") {
		present[form] = true
	}
	for i := 0; i < 300; i++ {
		sentence := fake(t, en, "en_US.sentence")
		w := strings.Fields(strings.ToLower(strings.TrimRight(sentence, ".!?")))
		last := len(w) - 1
		ok := noun[w[last]] && adj[w[last-1]]
		if strings.HasSuffix(sentence, "?") {
			ok = ok && verb[w[last-3]] && noun[w[last-4]] && adj[w[last-5]]
		} else {
			ok = ok && len(w) == 7 && present[w[3]] && ((adj[w[1]] && noun[w[2]]) || (noun[w[1]] && adv[w[2]]))
		}
		if !ok {
			t.Errorf("sentence %q holds a word its pattern doesn't read from the tables", sentence)
		}
	}
}

func TestEnglishParagraphIsThreeToSixSentences(t *testing.T) {
	t.Parallel()
	en := shipped(t, WithSeed(9))
	paragraph := regexp.MustCompile(`^[A-Z][^.!?]*[.!?]( [A-Z][^.!?]*[.!?]){2,5}$`)
	terminators := regexp.MustCompile(`[.!?]`)
	counts := map[int]bool{}
	for i := 0; i < 300; i++ {
		p := fake(t, en, "en_US.paragraph")
		if !paragraph.MatchString(p) {
			t.Fatalf("en_US.paragraph = %q, want three to six sentences", p)
		}
		counts[len(terminators.FindAllString(p, -1))] = true
	}
	for n := 3; n <= 6; n++ {
		if !counts[n] {
			t.Errorf("300 paragraphs never held %d sentences (held %v)", n, counts)
		}
	}
}

func TestEnglishWordCarriesWordNetsNotice(t *testing.T) {
	t.Parallel()
	notice, err := enus.FS.ReadFile("WORDNET-LICENSE.txt")
	if err != nil {
		t.Fatalf("the en_US module doesn't carry WordNet's licence: %v", err)
	}
	for _, want := range []string{
		"WordNet Release 3.0",
		"WordNet 3.0 Copyright 2006 by Princeton University. All rights reserved.",
		`PROVIDED "AS IS"`,
	} {
		if !strings.Contains(string(notice), want) {
			t.Errorf("WordNet's licence in the en_US module lacks %q", want)
		}
	}
}
