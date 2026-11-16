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

func TestShippedCalendarsNameEveryMonthAndWeekdayByNumber(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, c := range shippedLocales {
		for table, names := range map[string][]string{"month": c.months, "weekday": c.weekdays} {
			for i, want := range names {
				path := c.name + "." + table + "[" + strconv.Itoa(i+1) + "]"
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
	for _, c := range shippedLocales {
		for table, names := range map[string][]string{"month": c.months, "weekday": c.weekdays} {
			path := c.name + "." + table
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

func TestShippedEverydayCategoriesDrawTheirLocalShape(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, l := range shippedLocales {
		e := l.everyday
		if e == nil {
			continue
		}
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
	}
}

func TestShippedUrlsDrawTheLocalesOwnHostAndPathWords(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, l := range shippedLocales {
		e := l.everyday
		if e == nil {
			continue
		}
		tld := l.name[len(l.name)-2:]
		path := l.name + ".url"
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
	for _, l := range shippedLocales {
		e := l.everyday
		if e == nil {
			continue
		}
		path := l.name + ".color"
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
	for _, l := range shippedLocales {
		e := l.everyday
		if e == nil {
			continue
		}
		path := l.name + ".username"
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
