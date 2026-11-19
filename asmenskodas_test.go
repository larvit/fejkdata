package fejkdata

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// Per python-stdnum's lt.asmens, 33309240064 is a valid code, whose first ten digits have a first
// sum of 10 and so take the second weights, and the same ten digits with another check digit are
// not.
func TestValidAsmensKodasAppliesTheTwoWeights(t *testing.T) {
	for code, want := range map[string]bool{
		"33309240064": true, "33309240063": false, "33309240060": false, "3330924006": false, "333092400640": false,
		"93309240064": false, "03309240064": false, "73309240064": false, "33313240064": false, "33309320064": false,
	} {
		if got := fejktest.ValidAsmensKodas(code); got != want {
			t.Errorf("ValidAsmensKodas(%q) = %v, want %v", code, got, want)
		}
	}
}

const asmensKodasTemplate = `{"format":"{asmenskodas(sex,d)}","d":"%s","sex":"%s"}`

func asmensKodasOf(t *testing.T, f *Generator, date, sex string) string {
	t.Helper()
	return mustRender(t, f, fmt.Sprintf(asmensKodasTemplate, date, sex))
}

// The first digit is 1 for a man and 2 for a woman born in the 1800s, 3 and 4 in the 1900s and 5
// and 6 in the 2000s, as LST 1185-91 lists them according to the Lithuanian Wikipedia; the date
// follows as year, month and day of two digits each.
func TestBuiltinAsmenskodasStartsWithTheDigitOfItsSexAndCentury(t *testing.T) {
	f := emptyGenerator(21)
	for _, c := range []struct{ date, sex, want string }{
		{"18990101", "1", "1990101"}, {"18990101", "2", "2990101"},
		{"19330924", "1", "3330924"}, {"19330924", "2", "4330924"},
		{"20240229", "1", "5240229"}, {"20240229", "2", "6240229"},
		{"20991231", "2", "6991231"},
	} {
		t.Run(c.date+" "+c.sex, func(t *testing.T) {
			got := asmensKodasOf(t, f, c.date, c.sex)
			if len(got) != 11 || got[:7] != c.want || !fejktest.ValidAsmensKodas(got) {
				t.Errorf("asmenskodas for %s and %s = %q, want a valid code starting with %s", c.date, c.sex, got, c.want)
			}
		})
	}
}

func TestBuiltinAsmenskodasRendersNothingForAnOperandItDoesNotRead(t *testing.T) {
	f := emptyGenerator(22)
	for _, c := range []struct{ date, sex string }{
		{"19330924", ""}, {"19330924", "0"}, {"19330924", "3"}, {"19330924", "m"},
		{"", "1"}, {"1933092a", "1"}, {"193309", "1"}, {"1933-09-24", "1"}, {"17991231", "1"},
		{"21000101", "1"}, {"19331324", "1"}, {"19330932", "1"}, {"20230229", "1"},
	} {
		t.Run(c.date+" "+c.sex, func(t *testing.T) {
			if got := asmensKodasOf(t, f, c.date, c.sex); got != "" {
				t.Errorf("asmenskodas for %q and %q = %q, want nothing", c.date, c.sex, got)
			}
		})
	}
}

// The serial number, the eighth to tenth digit, is three digits; the sources state no range, so 000
// is left out.
func TestBuiltinAsmenskodasDrawsASerialNumberFrom001To999AndItsCheckDigit(t *testing.T) {
	seen := map[int]bool{}
	for seed := range uint64(4000) {
		got := asmensKodasOf(t, emptyGenerator(seed), "19330924", "1")
		serial, err := strconv.Atoi(got[7:10])
		if err != nil || serial < 1 || !fejktest.ValidAsmensKodas(got) {
			t.Fatalf("seed %d: %q, want a valid code with a serial number from 001 to 999", seed, got)
		}
		seen[serial/100] = true
	}
	if len(seen) != 10 {
		t.Errorf("4000 codes drew serial numbers in the hundreds %v, want every hundred from 0 to 9", seen)
	}
}

func TestBuiltinAsmenskodasTakesTheCenturyFromTheDateAndNotFromToday(t *testing.T) {
	for _, today := range []time.Time{time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2150, 6, 1, 0, 0, 0, 0, time.UTC)} {
		f := engineOn(23, today)
		if got := asmensKodasOf(t, f, "19330924", "2"); got[:7] != "4330924" {
			t.Errorf("asmenskodas on %s = %q, want one starting 4330924", today.Format("2006-01-02"), got)
		}
	}
}
