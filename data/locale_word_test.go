package data_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func localeWordForms(t *testing.T, locale, part string) map[string]string {
	t.Helper()
	forms := map[string]string{}
	for _, row := range fejktest.TSVRows(t, locale+"/word-"+part+".tsv") {
		for column, form := range row {
			if column != "word" {
				forms[row["word"]] = form
			}
		}
		if _, ok := forms[row["word"]]; !ok {
			forms[row["word"]] = ""
		}
	}
	return forms
}

func localeWordSets(t *testing.T, locale, part string) (words, forms map[string]bool) {
	t.Helper()
	words, forms = map[string]bool{}, map[string]bool{}
	for word, form := range localeWordForms(t, locale, part) {
		words[word] = true
		forms[form] = true
	}
	return words, forms
}

func wordTableColumns(t *testing.T, locale, part string) []string {
	t.Helper()
	raw, err := os.ReadFile(locale + "/word-" + part + ".tsv")
	if err != nil {
		t.Fatal(err)
	}
	header, _, _ := strings.Cut(string(raw), "\n")
	return strings.Split(header, "\t")
}

func hasWords(l shippedLocale) bool { return l.wordForms != nil }

func TestLocaleWordTablesAndWordJSONCarryTheFormsTheLocaleStates(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		fields := readJSON(t, l.name+"/word.json")
		for part, column := range l.wordForms {
			t.Run(part, func(t *testing.T) {
				want := []string{"word"}
				if column != "" {
					want = append(want, column)
				}
				if got := wordTableColumns(t, l.name, part); !slices.Equal(got, want) {
					t.Errorf("word-%s.tsv holds the columns %q, want %q", part, got, want)
				}
				field, _ := fields[part].(map[string]any)
				if _, reads := field[column]; reads != (column != "") || column == "" && field != nil {
					t.Errorf("word.json's %s is %v, want a field reading the form %q and no other", part, fields[part], column)
				}
			})
		}
	})
}

func TestLocaleWordTablesHoldTheirSourcesWordsAndForms(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ locale, part, word, form string }{
		{"nb_NO", "noun", "hus", "huset"},
		{"nb_NO", "noun", "jente", "jenta"},
		{"nb_NO", "verb", "løpe", "løper"},
		{"nb_NO", "adj", "stor", "store"},
		{"nb_NO", "adv", "ofte", ""},
		{"nn_NO", "noun", "hus", "huset"},
		{"nn_NO", "noun", "jente", "jenta"},
		{"nn_NO", "verb", "kaste", "kastar"},
		{"nn_NO", "adj", "stor", "store"},
		{"nn_NO", "adv", "ofte", ""},
		{"da_DK", "noun", "hus", "huset"},
		{"da_DK", "noun", "pige", "pigen"},
		{"da_DK", "verb", "løbe", "løber"},
		{"da_DK", "adj", "stor", "store"},
		{"da_DK", "adv", "ofte", ""},
		{"fi_FI", "noun", "talo", ""},
		{"fi_FI", "noun", "kissa", ""},
		{"fi_FI", "verb", "juosta", ""},
		{"fi_FI", "adj", "suuri", ""},
		{"fi_FI", "adv", "usein", ""},
		{"de_DE", "noun", "Haus", "Häuser"},
		{"de_DE", "noun", "Frau", "Frauen"},
		{"de_DE", "verb", "laufen", ""},
		{"de_DE", "adj", "groß", ""},
		{"de_DE", "adv", "oft", ""},
		{"nb_NO", "verb", "avslå", "avslår"},
		{"da_DK", "noun", "abessiner", "abessineren"},
		{"sv_SE", "noun", "hus", "huset"},
		{"sv_SE", "noun", "flicka", "flickan"},
		{"sv_SE", "verb", "springa", "springer"},
		{"sv_SE", "adj", "stor", "stora"},
		{"sv_SE", "adv", "ofta", ""},
		{"en_US", "noun", "house", ""},
		{"en_US", "noun", "river", ""},
		{"en_US", "verb", "run", "runs"},
		{"en_US", "verb", "carry", "carries"},
		{"en_US", "verb", "go", "goes"},
		{"en_US", "verb", "pass", "passes"},
		{"en_US", "verb", "have", "has"},
		{"en_US", "verb", "be", "is"},
		{"en_US", "verb", "woo", "woos"},
		{"en_US", "verb", "tattoo", "tattoos"},
		{"en_US", "verb", "radio", "radios"},
		{"en_US", "verb", "solo", "solos"},
		{"en_US", "verb", "echo", "echoes"},
		{"en_US", "verb", "veto", "vetoes"},
		{"en_US", "verb", "stomach", "stomachs"},
		{"en_US", "adj", "big", ""},
		{"en_US", "adv", "often", ""},
	} {
		t.Run(c.locale+" "+c.part+" "+c.word, func(t *testing.T) {
			if got, ok := localeWordForms(t, c.locale, c.part)[c.word]; !ok || got != c.form {
				t.Errorf("%s word-%s holds %q with form %q (held: %v), want %q", c.locale, c.part, c.word, got, ok, c.form)
			}
		})
	}
}

type absentWord struct{ locale, part, word string }

func assertWordsAbsent(t *testing.T, absent []absentWord, why string) {
	t.Helper()
	for _, c := range absent {
		t.Run(c.locale+" "+c.part+" "+c.word, func(t *testing.T) {
			if _, ok := localeWordForms(t, c.locale, c.part)[c.word]; ok {
				t.Errorf("%s word-%s holds %q, %s", c.locale, c.part, c.word, why)
			}
		})
	}
}

func TestLocaleWordTablesDropTheRudeWordsListed(t *testing.T) {
	t.Parallel()
	assertWordsAbsent(t, []absentWord{
		{"nb_NO", "noun", "helvete"},
		{"nn_NO", "noun", "helvete"},
		{"da_DK", "noun", "lort"},
		{"fi_FI", "noun", "kusi"},
		{"de_DE", "noun", "Arsch"},
		{"de_DE", "noun", "Titte"},
		{"da_DK", "verb", "lude"},
	}, "a word on LDNOOBW's list, or whose form is")
}

func TestLocaleWordTablesHoldOnlyTheirSourcesHeadwords(t *testing.T) {
	t.Parallel()
	assertWordsAbsent(t, []absentWord{
		{"nb_NO", "adj", "dumma"},
		{"da_DK", "adv", "afsindig"},
		{"da_DK", "noun", "aalborgenser"},
	}, "a form or variant its source doesn't list as a headword of that part")
}

func TestLocaleNounTablesHoldTensOfThousandsOfNouns(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		if n := len(localeWordForms(t, l.name, "noun")); n < 20000 {
			t.Errorf("%s word-noun holds %d nouns, want its source's, tens of thousands", l.name, n)
		}
	})
}

func TestLocaleWordTablesHoldWordsOfLettersAlone(t *testing.T) {
	t.Parallel()
	for _, l := range shippedLocales {
		for part, column := range l.wordForms {
			t.Run(l.name+" "+part, func(t *testing.T) {
				spelled := l.word
				if part == "noun" {
					spelled = l.noun
				}
				for word, form := range localeWordForms(t, l.name, part) {
					if !spelled.MatchString(word) || column != "" && !spelled.MatchString(form) {
						t.Fatalf("%s word-%s holds %q with form %q, want each spelled %s", l.name, part, word, form, spelled)
					}
				}
			})
		}
	}
}

func TestLocaleWordPathsDrawTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(5))
	for _, l := range shippedLocales {
		for part, column := range l.wordForms {
			words, forms := localeWordSets(t, l.name, part)
			paths := map[string]map[string]bool{l.name + ".word." + part: words}
			if column != "" {
				paths[l.name+".word."+part+"."+column] = forms
			}
			for path, held := range paths {
				t.Run(path, func(t *testing.T) {
					fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
						if !held[got] {
							t.Fatalf("%s = %q, which its table does not hold", path, got)
						}
					})
				})
			}
		}
	}
}

func TestLocaleBoundWordReadsItsOwnForm(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(6))
	for _, l := range shippedLocales {
		for part, column := range l.wordForms {
			if column == "" {
				continue
			}
			path := l.name + ".word." + part
			t.Run(path, func(t *testing.T) {
				forms := localeWordForms(t, l.name, part)
				fejktest.EachDraw(200, func() string { return fejktest.FakeTemplate(t, f, "{/"+path+" as w}{w}|{w."+column+"}") }, func(got string) {
					word, form, _ := strings.Cut(got, "|")
					if forms[word] != form {
						t.Fatalf("%s %q printed %s %q, want its own, %q", path, word, column, form, forms[word])
					}
				})
			})
		}
	}
}

func TestLocaleWordDrawsNounsAndOtherPartsOfSpeech(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(5))
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		nouns, _ := localeWordSets(t, l.name, "noun")
		seen := fejktest.TallyDraws(2000, func() string {
			if nouns[fejktest.Fake(t, f, l.name+".word")] {
				return "noun"
			}
			return "other"
		})
		if seen["noun"] == 0 || seen["other"] == 0 {
			t.Errorf("%s.word drew %v, want nouns and other parts of speech", l.name, seen)
		}
	})
}

func TestLocaleWordWeighsAPartOfSpeechByItsRows(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		raw, err := os.ReadFile(l.name + "/word.json")
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
			if rows := len(localeWordForms(t, l.name, strings.TrimPrefix(table, "word-"))); part.Weight != rows {
				t.Errorf("%s weighs %d, want its %d rows", table, part.Weight, rows)
			}
		}
	})
}

// The words a sentence pattern writes itself: those outside its references, in every string it holds.
func sentencePatternWords(t *testing.T, locale string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(locale + "/sentence.json")
	if err != nil {
		t.Fatal(err)
	}
	var patterns any
	if err := json.Unmarshal(raw, &patterns); err != nil {
		t.Fatal(err)
	}
	words := map[string]bool{}
	references := regexp.MustCompile(`\{[^}]*\}`)
	var walk func(any)
	walk = func(v any) {
		switch v := v.(type) {
		case string:
			for _, w := range strings.FieldsFunc(references.ReplaceAllString(v, " "), sentenceSeparator) {
				words[w] = true
			}
		case []any:
			for _, item := range v {
				walk(item)
			}
		case map[string]any:
			for _, item := range v {
				walk(item)
			}
		}
	}
	walk(patterns)
	return words
}

func sentenceSeparator(r rune) bool { return strings.ContainsRune(" .,!?", r) }

func TestLocaleSentenceIsMadeOfItsWordsAndItsPatternsOwn(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(7))
	sentence := regexp.MustCompile(`^\p{Lu}[^.!?]*[.!?]$`)
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		held := sentencePatternWords(t, l.name)
		nouns := map[string]bool{}
		for part := range l.wordForms {
			words, forms := localeWordSets(t, l.name, part)
			for w := range words {
				held[w] = true
			}
			for w := range forms {
				held[w] = true
			}
			if part == "noun" {
				nouns = words
				for w := range forms {
					nouns[w] = true
				}
			}
		}
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, l.name+".sentence") }, func(s string) {
			if !sentence.MatchString(s) {
				t.Fatalf("%s.sentence = %q, want one sentence, capitalised", l.name, s)
			}
			hasNoun := false
			for _, w := range strings.FieldsFunc(s, sentenceSeparator) {
				if !held[w] {
					t.Fatalf("%s.sentence = %q holds %q, a word neither its tables nor its pattern hold", l.name, s, w)
				}
				hasNoun = hasNoun || nouns[w]
			}
			if !hasNoun {
				t.Fatalf("%s.sentence = %q holds no noun of its table", l.name, s)
			}
		})
	})
}

func TestLocaleParagraphDrawsThreeToSixSentences(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(9))
	paragraph := regexp.MustCompile(`^\p{Lu}[^.!?]*[.!?]( \p{Lu}[^.!?]*[.!?]){2,5}$`)
	terminators := regexp.MustCompile(`[.!?]`)
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		counts := fejktest.TallyDraws(300, func() string {
			p := fejktest.Fake(t, f, l.name+".paragraph")
			if !paragraph.MatchString(p) {
				t.Fatalf("%s.paragraph = %q, want three to six sentences", l.name, p)
			}
			return strconv.Itoa(len(terminators.FindAllString(p, -1)))
		})
		for _, n := range []string{"3", "4", "5", "6"} {
			if counts[n] == 0 {
				t.Errorf("300 paragraphs never held %s sentences (held %v)", n, counts)
			}
		}
	})
}
