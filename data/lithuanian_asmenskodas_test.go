package data_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
)

const lithuanianAsmensKodasDraws = 20000

func eachLithuanianAsmensKodas(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(lithuanianAsmensKodasDraws, func() string { return fejktest.Fake(t, f, "lt_LT.asmens-kodas") }, check)
}

func TestLithuanianAsmensKodasIsElevenDigitsWithAValidDateAndCheckDigit(t *testing.T) {
	eachLithuanianAsmensKodas(t, func(got string) {
		if !fejktest.ValidAsmensKodas(got) {
			t.Fatalf("%q is not a valid asmens kodas", got)
		}
	})
}

func lithuanianAsmensKodasBirth(got string) time.Time {
	year := 1800 + 100*(int(got[0]-'1')/2)
	yy, _ := strconv.Atoi(got[1:3])
	mm, _ := strconv.Atoi(got[3:5])
	dd, _ := strconv.Atoi(got[5:7])
	return time.Date(year+yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
}

func TestLithuanianAsmensKodasIsOfSomeoneBornInTheLastHundredYearsNotAfterToday(t *testing.T) {
	eachLithuanianAsmensKodas(t, func(got string) {
		born := lithuanianAsmensKodasBirth(got)
		if born.After(fejktest.TestDay) || born.Before(fejktest.TestDay.AddDate(-100, 0, 0)) {
			t.Fatalf("%q is of someone born %s, want a day in the 100 years up to %s", got, born.Format("2006-01-02"), fejktest.TestDay.Format("2006-01-02"))
		}
	})
}

// Eurostat's table demo_pjan, the population on 1 January 2025: 2,890,284 people aged 0 to 99, of
// whom 497,128 were under 18, 1,790,026 were 18 to 64 and 603,130 were 65 to 99.
func TestLithuanianAsmensKodasFollowsTheAgeBandsOfEurostatsPopulation(t *testing.T) {
	bands := map[string]int{}
	eachLithuanianAsmensKodas(t, func(got string) {
		born := lithuanianAsmensKodasBirth(got)
		age := fejktest.TestDay.Year() - born.Year()
		if born.AddDate(age, 0, 0).After(fejktest.TestDay) {
			age--
		}
		switch {
		case age < 18:
			bands["under 18"]++
		case age < 65:
			bands["18 to 64"]++
		default:
			bands["65 to 99"]++
		}
	})
	for band, want := range map[string]float64{"under 18": 17.20, "18 to 64": 61.93, "65 to 99": 20.87} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / lithuanianAsmensKodasDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the codes, want %.2f%% by Eurostat's population", band, share, want)
			}
		})
	}
}

func TestLithuanianAsmensKodasAgreesWithABoundSex(t *testing.T) {
	f := shipped(t)
	seen := map[string]bool{}
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/lt_LT.sex as s}{s.code}|{/lt_LT.asmens-kodas.birth as b}{asmenskodas(s.sex-digit.digit,b)}")
	}, func(got string) {
		sex, code, _ := strings.Cut(got, "|")
		seen[sex] = true
		if want := map[string]int{"m": 1, "f": 0}[sex]; len(code) != 11 || int(code[0]-'0')%2 != want || !fejktest.ValidAsmensKodas(code) {
			t.Fatalf("%q: want a valid asmens kodas starting with an odd digit for a man and an even one for a woman", got)
		}
	})
	if !seen["m"] || !seen["f"] {
		t.Errorf("500 codes drew the sexes %v, want both", seen)
	}
}
