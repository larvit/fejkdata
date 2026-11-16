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

type calendar struct {
	locale   string
	months   []string
	weekdays []string
}

var calendars = []calendar{
	{
		"en_US",
		[]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		[]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
	},
	{
		"nb_NO",
		[]string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		[]string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
	},
	{
		"da_DK",
		[]string{"januar", "februar", "marts", "april", "maj", "juni", "juli", "august", "september", "oktober", "november", "december"},
		[]string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
	},
	{
		"fi_FI",
		[]string{"tammikuu", "helmikuu", "maaliskuu", "huhtikuu", "toukokuu", "kesäkuu", "heinäkuu", "elokuu", "syyskuu", "lokakuu", "marraskuu", "joulukuu"},
		[]string{"maanantai", "tiistai", "keskiviikko", "torstai", "perjantai", "lauantai", "sunnuntai"},
	},
	{
		"de_DE",
		[]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		[]string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"},
	},
}

func TestShippedCalendarsNameEveryMonthAndWeekdayByNumber(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, c := range calendars {
		for table, names := range map[string][]string{"month": c.months, "weekday": c.weekdays} {
			for i, want := range names {
				path := c.locale + "." + table + "[" + strconv.Itoa(i+1) + "]"
				t.Run(path, func(t *testing.T) {
					if got := fejktest.Fake(t, f, path); got != want {
						t.Errorf("Fake(%q) = %q, want %q", path, got, want)
					}
				})
			}
		}
	}
}

func TestShippedCalendarsDrawEveryMonthAndWeekdayAlone(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, c := range calendars {
		for table, names := range map[string][]string{"month": c.months, "weekday": c.weekdays} {
			path := c.locale + "." + table
			t.Run(path, func(t *testing.T) {
				got := sortedNamesDrawn(fejktest.TallyDraws(2000, func() string { return fejktest.Fake(t, f, path) }))
				want := slices.Clone(names)
				slices.Sort(want)
				if !slices.Equal(got, want) {
					t.Errorf("%s drew %v in 2000 draws, want exactly %v", path, got, want)
				}
			})
		}
	}
}

type everyday struct {
	locale   string
	price    string
	time     string
	url      string
	username string
	colors   []string
}

const (
	slug    = `[a-z0-9]+(-[a-z0-9]+)*`
	urlPath = `(/` + slug + `){0,2}`
)

var everydays = []everyday{
	{"nb_NO", `^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grønn", "Rød"}},
	{"nn_NO", `^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Raud"}},
	{"da_DK", `^\d{1,3}(\.\d{3})?(,\d{2})? kr\.$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|dk)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Rød"}},
	{"fi_FI", `^\d{1,3}( \d{3})?,\d{2} €$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|fi)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Sininen", "Vihreä", "Punainen"}},
	{"de_DE", `^\d{1,3}(\.\d{3})?,\d{2} €$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|de)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blau", "Grün", "Rot"}},
}

func TestShippedEverydayCategoriesDrawTheirLocalShape(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, e := range everydays {
		for category, pattern := range map[string]string{"price": e.price, "time": e.time, "url": e.url, "username": e.username} {
			path, re := e.locale+"."+category, regexp.MustCompile(pattern)
			t.Run(path, func(t *testing.T) {
				fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
					if !re.MatchString(got) {
						t.Fatalf("%s drew %q, want a match for %s", path, got, pattern)
					}
				})
			})
		}
	}
}

func TestShippedUrlsDrawTheLocalesOwnHostAndPathWords(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, e := range everydays {
		tld := e.locale[len(e.locale)-2:]
		path := e.locale + ".url"
		t.Run(path, func(t *testing.T) {
			var local, deep int
			fejktest.EachDraw(600, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
				if strings.Contains(got, "example."+strings.ToLower(tld)) {
					local++
				}
				if strings.Count(got, "/") == 4 {
					deep++
				}
			})
			if local == 0 || deep == 0 {
				t.Errorf("%s drew %d hosts under .%s and %d two-segment paths in 600 draws, want some of each", path, local, strings.ToLower(tld), deep)
			}
		})
	}
}

func TestShippedColorsDrawHexAndTheLocalesOwnNames(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	hex := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	for _, e := range everydays {
		path := e.locale + ".color"
		t.Run(path, func(t *testing.T) {
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
			for _, name := range e.colors {
				if seen[name] == 0 {
					t.Errorf("%s drew no %q in 6000 draws; drew %v", path, name, sortedNamesDrawn(seen))
				}
			}
		})
	}
}

func TestShippedUsernamesAreDrawnInMoreThanOneShape(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, e := range everydays {
		path := e.locale + ".username"
		t.Run(path, func(t *testing.T) {
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
}
