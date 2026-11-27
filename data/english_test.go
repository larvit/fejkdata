package data_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/data/en_US"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// TestShippedUSTaxIds pins the SSA and IRS ranges: an SSN's area is 001-899 but
// 666, its group 01-99 and its serial 0001-9999; an ITIN is 9XX-GG-XXXX with GG
// in 50-65, 70-88, 90-92 or 94-99.
func TestShippedUSTaxIds(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	ssn := regexp.MustCompile(`^(\d{3})-(\d{2})-(\d{4})$`)
	itin := regexp.MustCompile(`^9\d{2}-(5\d|6[0-5]|7\d|8[0-8]|9[0-24-9])-\d{4}$`)
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "en_US.ssn") }, func(v string) {
		m := ssn.FindStringSubmatch(v)
		if m == nil {
			t.Fatalf("ssn %q, want AAA-GG-SSSS", v)
		}
		area, _ := strconv.Atoi(m[1])
		group, _ := strconv.Atoi(m[2])
		serial, _ := strconv.Atoi(m[3])
		if area < 1 || area > 899 || area == 666 || group < 1 || serial < 1 {
			t.Fatalf("ssn %q is in a range the SSA never assigns", v)
		}
		if v := fejktest.Fake(t, f, "en_US.itin"); !itin.MatchString(v) {
			t.Fatalf("itin %q, want %s", v, itin)
		}
	})
}

// The Census 2010 counts. A name takes the form Wikidata's people of United States citizenship with that family name
// mostly have, such as `O'Brien`. Otherwise it stays in letters alone, such as `Delacruz`.
func TestEnglishLastNamesAreCensusCounts(t *testing.T) {
	last := nameCounts(t, "en_US", "last-name")[""]
	for name, want := range map[string]int{"Smith": 2376206, "McDonald": 176094, "O'Brien": 115385, "MacDonald": 45782, "LeBlanc": 40923, "D'Angelo": 11587, "St. John": 15663, "Delacruz": 49158} {
		if got := last[name]; got != want {
			t.Errorf("%q is borne by %d, want the Census's %d", name, got, want)
		}
	}
	for _, name := range []string{"Obrien", "Macdonald", "Leblanc", "Dangelo", "Stjohn"} {
		if _, ok := last[name]; ok {
			t.Errorf("%q is a row, but Wikidata's people with that family name mostly have another form of it", name)
		}
	}
}

// NANPA's central office code records mark N11, 555, 950, 958, 959 and 976 unassignable.
func TestShippedUSPhone(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(11))
	re := regexp.MustCompile(`^(?:\(\d{3}\) |\d{3}-)([2-9]\d{2})-\d{4}$`)
	unassignable := map[string]bool{"555": true, "950": true, "958": true, "959": true, "976": true}
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "en_US.phone") }, func(n string) {
		m := re.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q does not match %s", n, re)
		}
		if exchange := m[1]; exchange[1:] == "11" || unassignable[exchange] {
			t.Fatalf("phone %q has the unassignable exchange %s", n, exchange)
		}
	})
}

func TestShippedUSPhoneMobileAndFixedDrawOneNationalPool(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(11))
	national := regexp.MustCompile(`^\((\d{3})\) ([2-9]\d{2})-(\d{4})$`)
	phoneAreas := map[string]bool{}
	fejktest.EachDraw(3000, func() string { return fejktest.Fake(t, f, "en_US.phone.area") }, func(area string) { phoneAreas[area] = true })
	for _, path := range []string{"phone-mobile", "phone-fixed"} {
		fejktest.EachDraw(500, func() string {
			return fejktest.FakeTemplate(t, f, "{/en_US."+path+" as p}{p}|{p.national}|{p.international}")
		}, func(got string) {
			parts := strings.Split(got, "|")
			m := national.FindStringSubmatch(parts[0])
			if len(parts) != 3 || m == nil || parts[1] != parts[0] || parts[2] != "+1 "+m[1]+" "+m[2]+" "+m[3] {
				t.Fatalf("en_US.%s drew %q, want (AAA) EEE-LLLL twice and +1 AAA EEE LLLL", path, got)
			}
			if !phoneAreas[m[1]] {
				t.Fatalf("en_US.%s drew the area code %s, which en_US.phone never draws", path, m[1])
			}
		})
	}
}

func eachTitleAndSex(drawn []string, check func(got, title, sex string)) {
	for _, got := range drawn {
		title, sex, _ := strings.Cut(got, "|")
		check(got, strings.TrimSpace(title), sex)
	}
}

// TestShippedUSTitleAgreesWithSex pins that one person's title and sex agree, and
// that the everyday titles are reachable.
func TestShippedUSTitleAgreesWithSex(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	drawn := fejktest.DrawsInTurn(3000, func() string { return fejktest.FakeTemplate(t, f, `{/en_US.person as p}{p.prefix}|{p.sex}`) })
	seen := map[string]bool{}
	eachTitleAndSex(drawn, func(_, title, _ string) { seen[title] = true })
	t.Run("a title agrees with the sex", func(t *testing.T) {
		female := map[string]bool{"Miss": true, "Mrs": true, "Ms": true}
		eachTitleAndSex(drawn, func(got, title, sex string) {
			if title == "Mr" && sex != "male" || female[title] && sex != "female" {
				t.Fatalf("%q: the title contradicts the sex", got)
			}
		})
	})
	for _, want := range []string{"", "Mr", "Ms"} {
		t.Run("draws the title "+strconv.Quote(want), func(t *testing.T) {
			if !seen[want] {
				t.Errorf("en_US.person.prefix never drew %q in 3000 draws", want)
			}
		})
	}
}

func TestEnglishWordTablesHoldAsManyWordsAsWordNet(t *testing.T) {
	t.Parallel()
	for part, least := range map[string]int{"noun": 40000, "verb": 5000, "adj": 10000, "adv": 2500} {
		t.Run(part, func(t *testing.T) {
			if n := len(localeWordForms(t, "en_US", part)); n < least {
				t.Errorf("word-%s holds %d words, want WordNet's, %d or more", part, n, least)
			}
		})
	}
}

func TestEnglishWordTablesDropProperNounsAndSlurs(t *testing.T) {
	t.Parallel()
	nouns := localeWordForms(t, "en_US", "noun")
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
	nouns := localeWordForms(t, "en_US", "noun")
	for _, word := range []string{"table", "city", "teacher", "boy", "queen", "tool"} {
		t.Run(word, func(t *testing.T) {
			if _, ok := nouns[word]; !ok {
				t.Errorf("word-noun lacks %q", word)
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
	adj, _ := localeWordSets(t, "en_US", "adj")
	noun, _ := localeWordSets(t, "en_US", "noun")
	adv, _ := localeWordSets(t, "en_US", "adv")
	verb, present := localeWordSets(t, "en_US", "verb")
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
