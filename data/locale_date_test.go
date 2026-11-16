package data_test

import (
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func hasDate(l shippedLocale) bool { return l.dateLayout != "" }

func padsItsDay(l shippedLocale) bool {
	return strings.Contains(time.Date(1999, time.November, 5, 0, 0, 0, 0, time.UTC).Format(l.dateLayout), "05")
}

func writesItsDayUnpadded(l shippedLocale) bool { return hasDate(l) && !padsItsDay(l) }

func eachDate(t *testing.T, locale string, check func(got string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, locale+".date") }, check)
}

func TestLocaleDateIsWrittenInItsLayoutFrom1970To2029(t *testing.T) {
	eachLocale(t, hasDate, func(t *testing.T, l shippedLocale) {
		eachDate(t, l.name, func(got string) {
			d, err := time.Parse(l.dateLayout, got)
			if err != nil || d.Format(l.dateLayout) != got || d.Year() < 1970 || d.Year() > 2029 {
				t.Fatalf("date %q is not %s in 1970 to 2029: %v", got, l.dateLayout, err)
			}
		})
	})
}

func TestLocaleDateSpans1970To2029(t *testing.T) {
	eachLocale(t, hasDate, func(t *testing.T, l shippedLocale) {
		years := map[int]bool{}
		eachDate(t, l.name, func(got string) {
			d, _ := time.Parse(l.dateLayout, got)
			years[d.Year()] = true
		})
		if !years[1970] || !years[2029] {
			t.Errorf("500 dates drew the years %v, want 1970 and 2029 among them", years)
		}
	})
}

func TestLocaleDatesWithoutPaddingIncludeADayOfOneDigit(t *testing.T) {
	eachLocale(t, writesItsDayUnpadded, func(t *testing.T, l shippedLocale) {
		short := false
		eachDate(t, l.name, func(got string) {
			d, err := time.Parse(l.dateLayout, got)
			short = short || err == nil && d.Day() < 10 && d.Format(l.dateLayout) == got
		})
		if !short {
			t.Error("500 dates held no day of one digit, want one")
		}
	})
}
