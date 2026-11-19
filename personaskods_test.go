package fejkdata

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// 16117519997 is a valid old-format code and 32867300679 a valid new-format one, as python-stdnum's
// lv.pvn documents them; 32867300677 is the same new-format code with the wrong check digit.
// 16117510010 is a code whose sum is 10, so its check digit is 0, and 29020020009 is 29 February of
// 2000, which exists where 29 February 1900 does not.
func TestValidPersonasKodsChecksTheDateTheCenturyAndTheCheckDigit(t *testing.T) {
	for code, want := range map[string]bool{
		"16117519997": true, "32867300679": true, "16117510010": true, "29020020009": true,
		"01019901234": true, "01019921235": true, "31129911238": true,
		"16117519990": false, "32867300677": false, "16117510011": false, "29020010003": false,
		"29020000008": false, "31117510007": false, "16117530002": false, "1611751999": false,
		"161175199970": false, "1611751999a": false,
	} {
		if got := fejktest.ValidPersonasKods(code); got != want {
			t.Errorf("ValidPersonasKods(%q) = %v, want %v", code, got, want)
		}
	}
}

const personasKodsTemplate = `{"format":"{personaskods(d)}","d":"%s"}`

func personasKodsOf(t *testing.T, f *Generator, date string) string {
	t.Helper()
	return mustRender(t, f, fmt.Sprintf(personasKodsTemplate, date))
}

// A code starts with the date as day, month and two digits of the year, then 0 for the 1800s, 1 for
// the 1900s and 2 for the 2000s.
func TestBuiltinPersonasKodsStartsWithTheDateAndTheDigitOfItsCentury(t *testing.T) {
	f := emptyGenerator(21)
	for _, c := range []struct{ date, want string }{
		{"18990101", "0101990"}, {"18000101", "0101000"}, {"19000101", "0101001"},
		{"19680528", "2805681"}, {"20240229", "2902242"}, {"20991231", "3112992"},
	} {
		t.Run(c.date, func(t *testing.T) {
			got := personasKodsOf(t, f, c.date)
			if len(got) != 11 || got[:7] != c.want || !fejktest.ValidPersonasKods(got) {
				t.Errorf("personaskods for %s = %q, want a valid code starting with %s", c.date, got, c.want)
			}
		})
	}
}

func TestBuiltinPersonasKodsRendersNothingForADateItDoesNotRead(t *testing.T) {
	f := emptyGenerator(22)
	for _, date := range []string{
		"", "1968052a", "1968052", "680528", "196805280", "1968-05-28", "17991231", "21000101",
		"19681328", "19680532", "20230229",
	} {
		t.Run(date, func(t *testing.T) {
			if got := personasKodsOf(t, f, date); got != "" {
				t.Errorf("personaskods for %q = %q, want nothing", date, got)
			}
		})
	}
}

// The serial number, the eighth to tenth digit, is three digits; no source states which the register
// issues, so all of 000 to 999 are drawn.
func TestBuiltinPersonasKodsDrawsASerialNumberFrom000To999AndItsCheckDigit(t *testing.T) {
	hundreds, zero := map[int]bool{}, false
	for seed := range uint64(4000) {
		got := personasKodsOf(t, emptyGenerator(seed), "19680528")
		serial, err := strconv.Atoi(got[7:10])
		if err != nil || !fejktest.ValidPersonasKods(got) {
			t.Fatalf("seed %d: %q, want a valid code with a serial number from 000 to 999", seed, got)
		}
		hundreds[serial/100] = true
		zero = zero || got[7:10] == "000"
	}
	if len(hundreds) != 10 || !zero {
		t.Errorf("4000 codes drew serial numbers in the hundreds %v and 000 drawn: %v, want every hundred and 000", hundreds, zero)
	}
}

func TestBuiltinPersonasKodsTakesTheCenturyFromTheDateAndNotFromToday(t *testing.T) {
	for _, today := range []time.Time{time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2150, 6, 1, 0, 0, 0, 0, time.UTC)} {
		f := engineOn(23, today)
		if got := personasKodsOf(t, f, "19680528"); got[:7] != "2805681" {
			t.Errorf("personaskods on %s = %q, want one starting 2805681", today.Format("2006-01-02"), got)
		}
	}
}

func TestBuiltinPersonasKodsReadsTheDateOfABornCall(t *testing.T) {
	f := emptyGenerator(24)
	for range 200 {
		got := mustRender(t, f, `{"format":"{personaskods(d)}","d":"{born(0,99,'20060102')}"}`)
		if !fejktest.ValidPersonasKods(got) || got[6] != '1' && got[6] != '2' {
			t.Fatalf("personaskods over born(0,99) = %q, want a valid code of someone born in the 1900s or 2000s", got)
		}
	}
}
