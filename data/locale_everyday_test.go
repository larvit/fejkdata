package data_test

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func hasCalendar(l shippedLocale) bool { return l.months != nil && l.weekdays != nil }

func hasEveryday(l shippedLocale) bool { return l.everyday != nil }

func (l shippedLocale) calendars() map[string][]string {
	return map[string][]string{"month": l.months, "weekday": l.weekdays}
}

func TestShippedCalendarsNameEveryMonthAndWeekdayByNumber(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	eachLocale(t, hasCalendar, func(t *testing.T, l shippedLocale) {
		for table, names := range l.calendars() {
			for i, want := range names {
				path := l.name + "." + table + "[" + strconv.Itoa(i+1) + "]"
				t.Run(path, func(t *testing.T) {
					if got := fejktest.Fake(t, f, path); got != want {
						t.Errorf("Fake(%q) = %q, want %q", path, got, want)
					}
				})
			}
		}
	})
}

func TestShippedCalendarsDrawEveryMonthAndWeekdayAlone(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	eachLocale(t, hasCalendar, func(t *testing.T, l shippedLocale) {
		for table, names := range l.calendars() {
			path := l.name + "." + table
			t.Run(path, func(t *testing.T) {
				got := sortedNamesDrawn(fejktest.TallyDraws(2000, func() string { return fejktest.Fake(t, f, path) }))
				want := slices.Clone(names)
				slices.Sort(want)
				if !slices.Equal(got, want) {
					t.Errorf("%s drew %v in 2000 draws, want exactly %v", path, got, want)
				}
			})
		}
	})
}

func TestShippedEverydayCategoriesDrawTheirLocalShape(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	eachLocale(t, hasEveryday, func(t *testing.T, l shippedLocale) {
		e := l.everyday
		for category, pattern := range map[string]string{"price": e.price, "time": e.time, "url": e.url, "username": e.username} {
			path, re := l.name+"."+category, regexp.MustCompile(pattern)
			t.Run(path, func(t *testing.T) {
				fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
					if !re.MatchString(got) {
						t.Fatalf("%s drew %q, want a match for %s", path, got, pattern)
					}
				})
			})
		}
	})
}

func TestShippedUrlsDrawTheLocalesOwnHostAndPathWords(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	eachLocale(t, hasEveryday, func(t *testing.T, l shippedLocale) {
		tld := strings.ToLower(l.name[len(l.name)-2:])
		path := l.name + ".url"
		var local, deep int
		fejktest.EachDraw(600, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
			if strings.Contains(got, "example."+tld) {
				local++
			}
			if strings.Count(got, "/") == 4 {
				deep++
			}
		})
		if local == 0 || deep == 0 {
			t.Errorf("%s drew %d hosts under .%s and %d two-segment paths in 600 draws, want some of each", path, local, tld, deep)
		}
	})
}

func TestShippedColorsDrawHexAndTheLocalesOwnNames(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	hex := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	eachLocale(t, hasEveryday, func(t *testing.T, l shippedLocale) {
		path := l.name + ".color"
		seen := fejktest.TallyDraws(6000, func() string { return fejktest.Fake(t, f, path) })
		var hexes int
		for got, n := range seen {
			if hex.MatchString(got) {
				hexes += n
			}
		}
		if hexes < 2000 || hexes > 5000 {
			t.Errorf("%s drew %d hex colours in 6000 draws, want about two thirds", path, hexes)
		}
		for _, name := range l.everyday.colors {
			if seen[name] == 0 {
				t.Errorf("%s drew no %q in 6000 draws; drew %v", path, name, sortedNamesDrawn(seen))
			}
		}
	})
}

func TestShippedUsernamesAreDrawnInMoreThanOneShape(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	eachLocale(t, hasEveryday, func(t *testing.T, l shippedLocale) {
		path := l.name + ".username"
		seen := fejktest.TallyDraws(1000, func() string { return fejktest.Fake(t, f, path) })
		var dotted, digits int
		for got := range seen {
			if strings.ContainsAny(got, "._") {
				dotted++
			}
			if strings.ContainsAny(got, "0123456789") {
				digits++
			}
		}
		if len(seen) < 300 || dotted == 0 || digits == 0 {
			t.Errorf("%s drew %d usernames, %d with a separator and %d with a digit in 1000 draws, want many of each", path, len(seen), dotted, digits)
		}
	})
}

type usernameGenders struct {
	women, men         []string
	womanEnds, manEnds *regexp.Regexp
}

var usernameGenderLocales = map[string]usernameGenders{
	"lt_LT": {
		[]string{"aiste", "dovile", "egle", "gabija", "gintare", "ieva", "indre", "justina", "laura", "monika", "rasa", "ruta", "simona"},
		[]string{"algirdas", "andrius", "darius", "giedrius", "jonas", "kestutis", "lukas", "mantas", "marius", "paulius", "rokas", "tomas", "vytautas", "zilvinas"},
		regexp.MustCompile(`(aite|iene|yte|ute)$`), regexp.MustCompile(`(as|is|us|ys)$`),
	},
	"lv_LV": {
		[]string{"agnese", "anna", "dace", "elina", "evija", "ilze", "inese", "laura", "liga", "madara", "olga", "sandra", "santa", "zane"},
		[]string{"aleksandrs", "arturs", "edgars", "gatis", "janis", "juris", "kristaps", "martins", "maris", "peteris", "rihards", "toms", "valdis"},
		regexp.MustCompile(`[ae]$`), regexp.MustCompile(`s$`),
	},
}

func TestLithuanianAndLatvianUsernamesBearALastNameOfTheFirstNamesGender(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for locale, g := range usernameGenderLocales {
		t.Run(locale, func(t *testing.T) {
			var women, men int
			fejktest.EachDraw(4000, func() string { return fejktest.Fake(t, f, locale+".username") }, func(got string) {
				name := strings.TrimRight(got, "0123456789")
				for _, w := range g.women {
					if rest, ok := strings.CutPrefix(name, w); ok && strings.TrimLeft(rest, "._") != "" {
						women++
						if !g.womanEnds.MatchString(rest) {
							t.Fatalf("%s drew %q, a woman's first name with no feminine last name", locale, got)
						}
					}
				}
				for _, m := range g.men {
					if rest, ok := strings.CutPrefix(name, m); ok && strings.TrimLeft(rest, "._") != "" {
						men++
						if !g.manEnds.MatchString(rest) {
							t.Fatalf("%s drew %q, a man's first name with no masculine last name", locale, got)
						}
					}
				}
			})
			if women == 0 || men == 0 {
				t.Errorf("%s drew %d usernames of a woman's name and %d of a man's in 4000 draws, want some of each", locale, women, men)
			}
		})
	}
}
