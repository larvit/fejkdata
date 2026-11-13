package data_test

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data/en_US"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func englishWordForms(t *testing.T, table string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("en_US/" + table + ".tsv")
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

func englishWordSet(t *testing.T, table string) map[string]bool {
	t.Helper()
	held := map[string]bool{}
	for word := range englishWordForms(t, table) {
		held[word] = true
	}
	return held
}

func englishPresentForms(t *testing.T) map[string]bool {
	t.Helper()
	present := map[string]bool{}
	for _, form := range englishWordForms(t, "word-verb") {
		present[form] = true
	}
	return present
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
		{"word-verb", "woo", "woos"},
		{"word-verb", "tattoo", "tattoos"},
		{"word-verb", "radio", "radios"},
		{"word-verb", "solo", "solos"},
		{"word-verb", "echo", "echoes"},
		{"word-verb", "veto", "vetoes"},
		{"word-verb", "stomach", "stomachs"},
		{"word-adj", "big", ""},
		{"word-adv", "often", ""},
	} {
		t.Run(c.table+" "+c.word, func(t *testing.T) {
			if got, ok := englishWordForms(t, c.table)[c.word]; !ok || got != c.form {
				t.Errorf("%s holds %q with form %q (held: %v), want %q", c.table, c.word, got, ok, c.form)
			}
		})
	}
}

func TestEnglishWordTablesHoldAsManyWordsAsWordNet(t *testing.T) {
	t.Parallel()
	for table, least := range map[string]int{"word-noun": 40000, "word-verb": 5000, "word-adj": 10000, "word-adv": 2500} {
		t.Run(table, func(t *testing.T) {
			if n := len(englishWordForms(t, table)); n < least {
				t.Errorf("%s holds %d words, want WordNet's, %d or more", table, n, least)
			}
		})
	}
}

func TestEnglishWordTablesDropProperNounsAndSlurs(t *testing.T) {
	t.Parallel()
	nouns := englishWordForms(t, "word-noun")
	for _, word := range []string{"paris", "texas", "einstein", "monday", "london", "abc", "dago", "wop", "yid"} {
		t.Run(word, func(t *testing.T) {
			if _, ok := nouns[word]; ok {
				t.Errorf("word-noun holds %q, a proper noun, an acronym or a slur", word)
			}
		})
	}
}

func TestEnglishNounTableKeepsTheCommonNouns(t *testing.T) {
	t.Parallel()
	nouns := englishWordForms(t, "word-noun")
	for _, word := range []string{"table", "city", "teacher", "boy", "queen", "tool"} {
		t.Run(word, func(t *testing.T) {
			if _, ok := nouns[word]; !ok {
				t.Errorf("word-noun lacks %q", word)
			}
		})
	}
}

func holdsLowerCaseLettersAlone(t *testing.T, table string) {
	t.Helper()
	letters := regexp.MustCompile(`^[a-z]+$`)
	for word, form := range englishWordForms(t, table) {
		if !letters.MatchString(word) || (form != "" && !letters.MatchString(form)) {
			t.Fatalf("%s holds %q with form %q, want lower-case letters alone", table, word, form)
		}
	}
}

func TestEnglishWordTablesHoldLowerCaseLettersAlone(t *testing.T) {
	t.Parallel()
	for _, table := range []string{"word-noun", "word-verb", "word-adj", "word-adv"} {
		t.Run(table, func(t *testing.T) { holdsLowerCaseLettersAlone(t, table) })
	}
}

func TestEnglishWordPathsDrawTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(5))
	nouns, verbs, adjs, advs, present := englishWordSet(t, "word-noun"), englishWordSet(t, "word-verb"), englishWordSet(t, "word-adj"), englishWordSet(t, "word-adv"), englishPresentForms(t)
	for _, c := range []struct {
		path string
		held map[string]bool
	}{
		{"en_US.word.noun", nouns},
		{"en_US.word.verb", verbs},
		{"en_US.word.verb.present", present},
		{"en_US.word.adj", adjs},
		{"en_US.word.adv", advs},
	} {
		t.Run(c.path, func(t *testing.T) {
			fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, en, c.path) }, func(got string) {
				if !c.held[got] {
					t.Fatalf("%s = %q, which its table does not hold", c.path, got)
				}
			})
		})
	}
}

func TestEnglishWordDrawsNounsAndOtherPartsOfSpeech(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(5))
	nouns := englishWordSet(t, "word-noun")
	seen := fejktest.TallyDraws(2000, func() string {
		if nouns[fejktest.Fake(t, en, "en_US.word")] {
			return "noun"
		}
		return "other"
	})
	if seen["noun"] == 0 || seen["other"] == 0 {
		t.Errorf("en_US.word drew %v, want nouns and other parts of speech", seen)
	}
}

func TestEnglishBoundVerbReadsItsOwnPresent(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(6))
	forms := englishWordForms(t, "word-verb")
	fejktest.EachDraw(200, func() string { return fejktest.FakeTemplate(t, en, "{/en_US.word.verb as w}{w}|{w.present}") }, func(got string) {
		word, form, _ := strings.Cut(got, "|")
		if forms[word] != form {
			t.Fatalf("en_US.word.verb %q printed present %q, want its own, %q", word, form, forms[word])
		}
	})
}

func TestEnglishWordWeighsAPartOfSpeechByItsRows(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("en_US/word.json")
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
		t.Run(table, func(t *testing.T) {
			if rows := len(englishWordForms(t, table)); part.Weight != rows {
				t.Errorf("%s weighs %d, want its %d rows", table, part.Weight, rows)
			}
		})
	}
}

func TestEnglishWordDrawsAWideSpreadOfWords(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(8))
	distinct := fejktest.TallyDraws(5000, func() string { return fejktest.Fake(t, en, "en_US.word") })
	if len(distinct) < 3000 {
		t.Errorf("5000 draws of en_US.word gave %d distinct words, want a dictionary's spread", len(distinct))
	}
}

func TestEnglishSentenceDrawsTheWordsItsPatternNeeds(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(7))
	adj, noun, adv, verb, present := englishWordSet(t, "word-adj"), englishWordSet(t, "word-noun"), englishWordSet(t, "word-adv"), englishWordSet(t, "word-verb"), englishPresentForms(t)
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, en, "en_US.sentence") }, func(sentence string) {
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
	})
}

func TestEnglishParagraphIsThreeToSixSentences(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(9))
	paragraph := regexp.MustCompile(`^[A-Z][^.!?]*[.!?]( [A-Z][^.!?]*[.!?]){2,5}$`)
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, en, "en_US.paragraph") }, func(p string) {
		if !paragraph.MatchString(p) {
			t.Fatalf("en_US.paragraph = %q, want three to six sentences", p)
		}
	})
}

func TestEnglishParagraphDrawsEachSentenceCountFromThreeToSix(t *testing.T) {
	t.Parallel()
	en := shipped(t, fejkdata.WithSeed(9))
	terminators := regexp.MustCompile(`[.!?]`)
	counts := fejktest.TallyDraws(300, func() string {
		return strconv.Itoa(len(terminators.FindAllString(fejktest.Fake(t, en, "en_US.paragraph"), -1)))
	})
	for _, n := range []int{3, 4, 5, 6} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			if counts[strconv.Itoa(n)] == 0 {
				t.Errorf("300 paragraphs never held %d sentences (held %v)", n, counts)
			}
		})
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
		t.Run(want, func(t *testing.T) {
			if !strings.Contains(string(notice), want) {
				t.Errorf("WordNet's licence in the en_US module lacks %q", want)
			}
		})
	}
}
