package data_test

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func hasPerson(l shippedLocale) bool { return l.person != nil }

type drawnPerson struct{ raw, sex, prefix, first, last, whole string }

func eachPerson(t *testing.T, locale string, check func(p drawnPerson)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(2000, func() string {
		return fejktest.FakeTemplate(t, f, "{/"+locale+".person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		check(drawnPerson{got, cells[0], cells[1], cells[2], cells[3], cells[4]})
	})
}

func TestLocalePersonHasAFirstNameOfItsSex(t *testing.T) {
	eachLocale(t, hasPerson, func(t *testing.T, l shippedLocale) {
		bySex := nameCounts(t, l.name, "first-name")
		first := map[string]map[string]int{}
		for sex, word := range l.person.sexWords {
			first[word] = bySex[sex]
		}
		eachPerson(t, l.name, func(p drawnPerson) {
			if first[p.sex][p.first] == 0 {
				t.Fatalf("%q: %q is not a first name shipped for the sex %q", p.raw, p.first, p.sex)
			}
		})
	})
}

func TestLocalePersonHasAShippedLastName(t *testing.T) {
	eachLocale(t, hasPerson, func(t *testing.T, l shippedLocale) {
		last := nameCounts(t, l.name, "last-name")[""]
		eachPerson(t, l.name, func(p drawnPerson) {
			if last[p.last] == 0 {
				t.Fatalf("%q: %q is not a shipped last name", p.raw, p.last)
			}
		})
	})
}

func TestLocalePersonIsItsPrefixFirstNameAndLastName(t *testing.T) {
	eachLocale(t, hasPerson, func(t *testing.T, l shippedLocale) {
		eachPerson(t, l.name, func(p drawnPerson) {
			if p.whole != p.prefix+p.first+" "+p.last {
				t.Fatalf("%q: the person is not its prefix, first name and last name", p.raw)
			}
		})
	})
}

func TestLocalePersonMostlyBearsNoTitle(t *testing.T) {
	eachLocale(t, hasPerson, func(t *testing.T, l shippedLocale) {
		prefixes := map[string]int{}
		eachPerson(t, l.name, func(p drawnPerson) { prefixes[p.prefix]++ })
		for _, title := range l.person.titles {
			if prefixes[title+" "] == 0 {
				t.Errorf("no person of 2000 bears the title %q", title)
			}
		}
		if prefixes[""] < 1800 || len(prefixes) != len(l.person.titles)+1 {
			t.Errorf("prefixes over 2000 persons = %q, want a few of %q beside many with none", prefixes, l.person.titles)
		}
	})
}

func TestLocaleSexAndTitleAreSelectable(t *testing.T) {
	f := shipped(t)
	eachLocale(t, hasPerson, func(t *testing.T, l shippedLocale) {
		want := map[string]string{}
		for sex, word := range l.person.sexWords {
			want["{/"+l.name+".sex["+sex+"]}"] = word
		}
		for _, title := range l.person.titles {
			want["{/"+l.name+".title["+title+"]}"] = title
		}
		for tmpl, want := range want {
			t.Run(tmpl, func(t *testing.T) {
				if got, err := f.FakeTemplate(tmpl); err != nil || got != want {
					t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
				}
			})
		}
	})
}
