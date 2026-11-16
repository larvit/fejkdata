package data_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

var (
	paddedDate   = regexp.MustCompile(`^(0[1-9]|[12]\d|3[01])\.(0[1-9]|1[0-2])\.\d{4}$`)
	unpaddedDate = regexp.MustCompile(`^([1-9]|[12]\d|3[01])\.([1-9]|1[0-2])\.\d{4}$`)
)

func hasDate(l shippedLocale) bool { return l.dateLayout != "" }

func padsItsDate(l shippedLocale) bool { return l.dateLayout == "02.01.2006" }

func eachDate(t *testing.T, locale string, check func(got string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, locale+".date") }, check)
}

func TestLocaleDateIsDayMonthYearWithDotsFrom1970To2029(t *testing.T) {
	eachLocale(t, hasDate, func(t *testing.T, l shippedLocale) {
		shape := unpaddedDate
		if padsItsDate(l) {
			shape = paddedDate
		}
		eachDate(t, l.name, func(got string) {
			d, err := time.Parse(l.dateLayout, got)
			if err != nil || !shape.MatchString(got) || d.Year() < 1970 || d.Year() > 2029 {
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
	eachLocale(t, func(l shippedLocale) bool { return hasDate(l) && !padsItsDate(l) }, func(t *testing.T, l shippedLocale) {
		short := false
		eachDate(t, l.name, func(got string) { short = short || got[1] == '.' })
		if !short {
			t.Error("500 dates held no day of one digit, want one")
		}
	})
}
