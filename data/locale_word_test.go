package data_test

import (
	"cmp"
	"maps"
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

var partsOfSpeech = []string{"noun", "verb", "adj", "adv"}

func hasWords(l shippedLocale) bool { return l.words != nil }

// wordPart is one part of speech of a locale's word, with what its table should hold.
type wordPart struct {
	locale, part string
	wordTable
}

func (p wordPart) path() string { return p.locale + ".word." + p.part }

func everyWordTable(wordTable) bool { return true }
func hasForm(w wordTable) bool      { return w.form != "" }
func hasNoForm(w wordTable) bool    { return w.form == "" }

func eachWordPart(t *testing.T, has func(w wordTable) bool, test func(t *testing.T, p wordPart)) {
	t.Helper()
	for _, l := range shippedLocales {
		for part, table := range l.words {
			if has(table) {
				t.Run(l.name+".word."+part, func(t *testing.T) { test(t, wordPart{l.name, part, table}) })
			}
		}
	}
}

func TestLocaleWordTableWithAFormHoldsItsColumnAndWordJSONReadsIt(t *testing.T) {
	t.Parallel()
	eachWordPart(t, hasForm, func(t *testing.T, p wordPart) {
		if got, want := wordTableColumns(t, p.locale, p.part), []string{"word", p.form}; !slices.Equal(got, want) {
			t.Errorf("word-%s.tsv holds the columns %q, want %q", p.part, got, want)
		}
		field := readJSON[map[string]any](t, p.locale+"/word.json")[p.part]
		if fields, _ := field.(map[string]any); fields[p.form] == nil {
			t.Errorf("word.json's %s is %v, want a field reading the form %q", p.part, field, p.form)
		}
	})
}

func TestLocaleWordTableWithoutAFormHoldsItsWordAloneAndWordJSONReadsNoColumn(t *testing.T) {
	t.Parallel()
	eachWordPart(t, hasNoForm, func(t *testing.T, p wordPart) {
		if got, want := wordTableColumns(t, p.locale, p.part), []string{"word"}; !slices.Equal(got, want) {
			t.Errorf("word-%s.tsv holds the columns %q, want %q", p.part, got, want)
		}
		field := readJSON[map[string]any](t, p.locale+"/word.json")[p.part]
		if _, plain := field.(string); !plain {
			t.Errorf("word.json's %s is %v, want a plain format that reads no column", p.part, field)
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
		{"fr_FR", "noun", "maison", "maisons"},
		{"fr_FR", "noun", "cheval", "chevaux"},
		{"fr_FR", "noun", "pays", "pays"},
		{"fr_FR", "noun", "souris", "souris"},
		{"fr_FR", "verb", "courir", ""},
		{"fr_FR", "adj", "grand", ""},
		{"fr_FR", "adv", "souvent", ""},
		{"fr_FR", "adv", "bien", ""},
		{"fr_FR", "adv", "peu", ""},
		{"fr_FR", "verb", "décourager", ""},
		{"fr_FR", "noun", "merde", "merdes"},
		{"fr_FR", "verb", "branler", ""},
		{"et_EE", "noun", "maja", "majad"},
		{"et_EE", "noun", "mees", "mehed"},
		{"et_EE", "verb", "oskama", ""},
		{"et_EE", "adj", "hell", ""},
		{"et_EE", "adv", "alati", ""},
		{"et_EE", "noun", "kurat", "kuradid"},
		{"et_EE", "noun", "sitt", "sitad"},
		{"lv_LV", "noun", "māja", "mājas"},
		{"lv_LV", "noun", "vīrs", "vīri"},
		{"lv_LV", "verb", "skriet", ""},
		{"lv_LV", "adj", "liels", ""},
		{"lv_LV", "adv", "bieži", ""},
		{"lt_LT", "noun", "mama", "mamos"},
		{"lt_LT", "noun", "vyras", "vyrai"},
		{"lt_LT", "verb", "bėgti", ""},
		{"lt_LT", "adj", "didelis", ""},
		{"lt_LT", "adv", "dažnai", ""},
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
		{"fr_FR", "verb", "aimez"},
		{"fr_FR", "verb", "abolisses"},
		{"fr_FR", "verb", "expliquent"},
		{"lv_LV", "adj", "abejāds"},
		{"lv_LV", "verb", "pamācošs"},
		{"lv_LV", "verb", "rakstāms"},
		{"lv_LV", "noun", "abanācija"},
		{"lt_LT", "adj", "didelio"},
		{"lt_LT", "verb", "bėgdamas"},
		{"lt_LT", "noun", "namų"},
	}, "a form or variant its source doesn't list as a headword of that part")
}

func TestLocaleNounTablesHoldTheirSourcesNouns(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		want := cmp.Or(l.fewestNouns, 20000)
		if n := len(localeWordForms(t, l.name, "noun")); n < want {
			t.Errorf("%s word-noun holds %d nouns, want its source's, at least %d", l.name, n, want)
		}
	})
}

func TestLocaleWordTablesHoldWordsOfLettersAlone(t *testing.T) {
	t.Parallel()
	eachWordPart(t, everyWordTable, func(t *testing.T, p wordPart) {
		for word := range localeWordForms(t, p.locale, p.part) {
			if !p.spelled.MatchString(word) {
				t.Fatalf("word-%s holds %q, want it spelled %s", p.part, word, p.spelled)
			}
		}
	})
}

func TestLocaleWordTablesHoldFormsOfLettersAlone(t *testing.T) {
	t.Parallel()
	eachWordPart(t, hasForm, func(t *testing.T, p wordPart) {
		for word, form := range localeWordForms(t, p.locale, p.part) {
			if !p.spelled.MatchString(form) {
				t.Fatalf("word-%s holds %q with the form %q, want it spelled %s", p.part, word, form, p.spelled)
			}
		}
	})
}

func TestLocaleWordPathsDrawTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(5))
	eachWordPart(t, everyWordTable, func(t *testing.T, p wordPart) {
		words, _ := localeWordSets(t, p.locale, p.part)
		fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, p.path()) }, func(got string) {
			if !words[got] {
				t.Fatalf("%s = %q, which its table does not hold", p.path(), got)
			}
		})
	})
}

func TestLocaleWordFormPathsDrawAFormOfTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(5))
	eachWordPart(t, hasForm, func(t *testing.T, p wordPart) {
		_, forms := localeWordSets(t, p.locale, p.part)
		path := p.path() + "." + p.form
		fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
			if !forms[got] {
				t.Fatalf("%s = %q, which its table does not hold", path, got)
			}
		})
	})
}

func TestLocaleBoundWordReadsItsOwnForm(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(6))
	eachWordPart(t, hasForm, func(t *testing.T, p wordPart) {
		forms := localeWordForms(t, p.locale, p.part)
		fejktest.EachDraw(200, func() string { return fejktest.FakeTemplate(t, f, "{/"+p.path()+" as w}{w}|{w."+p.form+"}") }, func(got string) {
			word, form, _ := strings.Cut(got, "|")
			if forms[word] != form {
				t.Fatalf("%s %q printed %s %q, want its own, %q", p.path(), word, p.form, form, forms[word])
			}
		})
	})
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
		word := readJSON[struct {
			Any []struct {
				Format string `json:"format"`
				Weight int    `json:"weight"`
			} `json:"any"`
		}](t, l.name+"/word.json")
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
	patterns := readJSON[any](t, locale+"/sentence.json")
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

func sentenceSeparator(r rune) bool { return strings.ContainsRune(" .,!?„“", r) }

// localeWordsAndForms is every word and form the locale's tables of parts hold.
func localeWordsAndForms(t *testing.T, locale string, parts ...string) map[string]bool {
	t.Helper()
	held := map[string]bool{}
	for _, part := range parts {
		words, forms := localeWordSets(t, locale, part)
		maps.Copy(held, words)
		maps.Copy(held, forms)
	}
	return held
}

func TestLocaleSentenceIsMadeOfItsWordsAndItsPatternsOwn(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(7))
	sentence := regexp.MustCompile(`^\p{Lu}[^.!?]*[.!?]$`)
	eachLocale(t, hasWords, func(t *testing.T, l shippedLocale) {
		held, nouns := localeWordsAndForms(t, l.name, partsOfSpeech...), localeWordsAndForms(t, l.name, "noun")
		maps.Copy(held, sentencePatternWords(t, l.name))
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
