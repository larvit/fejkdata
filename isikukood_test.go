package fejkdata

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// EVS 585, as the Population Register applies it and python-stdnum documents it: 36805280109 is a
// valid code, and the same ten digits with another check digit are not. 36805280113 is a code whose
// first sum is 10, so its second weights give the check digit, and 36805280600 one whose second
// sum is 10 too, so its check digit is 0.
func TestValidIsikukoodAppliesTheTwoWeightsAndTheZero(t *testing.T) {
	for code, want := range map[string]bool{
		"36805280109": true, "36805280108": false, "36805280113": true, "36805280112": false,
		"36805280600": true, "36805280609": false, "3680528010": false, "368052801090": false,
		"96805280109": false, "06805280109": false, "36813280109": false, "36805320109": false,
	} {
		if got := fejktest.ValidIsikukood(code); got != want {
			t.Errorf("ValidIsikukood(%q) = %v, want %v", code, got, want)
		}
	}
}

const isikukoodTemplate = `{"format":"{/d as b}{isikukood(sex,b)}","d":"%s","sex":"%s"}`

func isikukoodOf(t *testing.T, f *Generator, date, sex string) string {
	t.Helper()
	return mustRender(t, f, fmt.Sprintf(isikukoodTemplate, date, sex))
}

// EVS 585: the first digit is 1 for a man and 2 for a woman born in the 1800s, 3 and 4 in the
// 1900s, 5 and 6 in the 2000s, and 7 and 8 in the 2100s; the date follows as year, month and day
// of two digits each.
func TestBuiltinIsikukoodStartsWithTheDigitOfItsSexAndCentury(t *testing.T) {
	f := emptyGenerator(11)
	for _, c := range []struct{ date, sex, want string }{
		{"18990101", "1", "1990101"}, {"18990101", "2", "2990101"},
		{"19680528", "1", "3680528"}, {"19680528", "2", "4680528"},
		{"20240229", "1", "5240229"}, {"20240229", "2", "6240229"},
		{"21000101", "1", "7000101"}, {"21991231", "2", "8991231"},
		{"18000101", "1", "1000101"},
	} {
		t.Run(c.date+" "+c.sex, func(t *testing.T) {
			got := isikukoodOf(t, f, c.date, c.sex)
			if len(got) != 11 || got[:7] != c.want || !fejktest.ValidIsikukood(got) {
				t.Errorf("isikukood for %s and %s = %q, want a valid code starting with %s", c.date, c.sex, got, c.want)
			}
		})
	}
}

func TestBuiltinIsikukoodRendersNothingForAnOperandItDoesNotRead(t *testing.T) {
	f := emptyGenerator(12)
	for _, c := range []struct{ date, sex string }{
		{"19680528", ""}, {"19680528", "0"}, {"19680528", "3"}, {"19680528", "12"}, {"19680528", "m"},
		{"", "1"}, {"680528", "1"}, {"196805280", "1"}, {"1968-05-28", "1"}, {"17991231", "1"},
		{"22000101", "1"}, {"19681328", "1"}, {"19680532", "1"}, {"20230229", "1"},
	} {
		t.Run(c.date+" "+c.sex, func(t *testing.T) {
			if got := isikukoodOf(t, f, c.date, c.sex); got != "" {
				t.Errorf("isikukood for %q and %q = %q, want nothing", c.date, c.sex, got)
			}
		})
	}
}

// EVS 585 names the serial number, the eighth to tenth digit, as a number of the people born on one
// day, so it is three digits; the Population Register starts it at 001.
func TestBuiltinIsikukoodDrawsASerialNumberFrom001To999AndItsCheckDigit(t *testing.T) {
	seen := map[int]bool{}
	for seed := range uint64(4000) {
		got := isikukoodOf(t, emptyGenerator(seed), "19680528", "1")
		serial, err := strconv.Atoi(got[7:10])
		if err != nil || serial < 1 || !fejktest.ValidIsikukood(got) {
			t.Fatalf("seed %d: %q, want a valid code with a serial number from 001 to 999", seed, got)
		}
		seen[serial/100] = true
	}
	if len(seen) != 10 {
		t.Errorf("4000 codes drew serial numbers in the hundreds %v, want every hundred from 0 to 9", seen)
	}
}

func TestBuiltinIsikukoodTakesTheCenturyFromTheDateAndNotFromToday(t *testing.T) {
	for _, today := range []time.Time{time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2150, 6, 1, 0, 0, 0, 0, time.UTC)} {
		f := engineOn(13, today)
		if got := isikukoodOf(t, f, "19680528", "2"); got[:7] != "4680528" {
			t.Errorf("isikukood on %s = %q, want one starting 4680528", today.Format("2006-01-02"), got)
		}
	}
}

func TestBuiltinIsikukoodReadsTheDateOfABornCall(t *testing.T) {
	f := emptyGenerator(14)
	for range 200 {
		got := mustRender(t, f, `{"format":"{/d as b}{isikukood(sex,b)}","d":"{born(0,99,'20060102')}","sex":"2"}`)
		if !fejktest.ValidIsikukood(got) || got[0] != '4' && got[0] != '6' {
			t.Fatalf("isikukood over born(0,99) = %q, want a valid code of a woman born in the 1900s or 2000s", got)
		}
	}
}
