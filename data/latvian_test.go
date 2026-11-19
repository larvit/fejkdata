package data_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestLatvianFirstNamesWeighByThePublishedCount(t *testing.T) {
	first := nameCounts(t, "lv_LV", "first-name")
	for _, c := range []struct {
		sex, name string
		want      int
	}{{"m", "Jānis", 44740}, {"m", "Aleksandrs", 31812}, {"f", "Anna", 21772}, {"f", "Jeļena", 20246}} {
		if got := first[c.sex][c.name]; got != c.want {
			t.Errorf("%q of sex %q weighs %d, want %d", c.name, c.sex, got, c.want)
		}
	}
}

func TestLatvianNamesKeepTheirLetters(t *testing.T) {
	first, last := nameCounts(t, "lv_LV", "first-name"), lastNameCounts(t, "lv_LV")
	for _, name := range []string{"Bērziņš", "Krūmiņš", "Kļaviņš", "Eglītis", "Dombrovskis", "Vītols"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
	for sex, names := range map[string][]string{"f": {"Līga", "Ņina", "Jeļena"}, "m": {"Ģirts", "Ēriks", "Mārtiņš"}} {
		for _, name := range names {
			if first[sex][name] == 0 {
				t.Errorf("first-name.tsv has no %q of sex %q", name, sex)
			}
		}
	}
}

func TestLatvianLastNamesWeighByRankAndTheRarerOnesShareTheLast(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	for sex, want := range map[string]map[string]int{
		"m": {"Bērziņš": 100000, "Kalniņš": 70711, "Ozoliņš": 57735, "Jansons": 50000, "Ivanovs": 21822, "Zvirbulis": 9129},
		"f": {"Bērziņa": 100000, "Kalniņa": 70711, "Ozoliņa": 57735, "Jansone": 50000, "Ivanova": 21822, "Zvirbule": 9129},
	} {
		for name, n := range want {
			if got := last[sex][name]; got != n {
				t.Errorf("%q of sex %q weighs %d, want %d", name, sex, got, n)
			}
		}
	}
}

func TestLatvianWomenBearTheFeminineFormOfALastName(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	for _, c := range []struct{ sex, name string }{
		{"f", "Bērziņa"}, {"m", "Bērziņš"}, {"f", "Balode"}, {"m", "Balodis"}, {"f", "Dombrovska"}, {"f", "Ozola"}, {"f", "Millere"},
		{"f", "Zvejniece"}, {"f", "Kalēja"}, {"f", "Liepa"}, {"m", "Liepa"},
	} {
		if last[c.sex][c.name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex %q", c.name, c.sex)
		}
	}
	for _, c := range []struct{ sex, name string }{{"f", "Bērziņš"}, {"m", "Bērziņa"}, {"f", "Balodis"}, {"m", "Balode"}, {"f", "Ivanovs"}} {
		if last[c.sex][c.name] != 0 {
			t.Errorf("last-name.tsv has %q of sex %q", c.name, c.sex)
		}
	}
}

func TestLatvianPersonBearsALastNameOfItsSex(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	bySex := map[string]string{"sieviete": "f", "vīrietis": "m"}
	eachPerson(t, "lv_LV", func(p drawnPerson) {
		if last[bySex[p.sex]][p.last] == 0 {
			t.Fatalf("%q: %q is not a last name shipped for the sex %q", p.raw, p.last, p.sex)
		}
	})
}

const latvianPersonasKodsDraws = 20000

func eachLatvianPersonasKods(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(latvianPersonasKodsDraws, func() string { return fejktest.Fake(t, f, "lv_LV.personas-kods") }, check)
}

func TestLatvianPersonasKodsIsElevenDigitsWithAValidDateAndCheckDigit(t *testing.T) {
	eachLatvianPersonasKods(t, func(got string) {
		if !fejktest.ValidPersonasKods(got) || got[:2] == "32" {
			t.Fatalf("%q is not a valid personas kods with a date", got)
		}
	})
}

func latvianPersonasKodsBirth(got string) time.Time {
	year := 1800 + 100*int(got[6]-'0')
	yy, _ := strconv.Atoi(got[4:6])
	mm, _ := strconv.Atoi(got[2:4])
	dd, _ := strconv.Atoi(got[:2])
	return time.Date(year+yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
}

func latvianPersonasKodsAge(got string) int {
	born := latvianPersonasKodsBirth(got)
	age := fejktest.TestDay.Year() - born.Year()
	if born.AddDate(age, 0, 0).After(fejktest.TestDay) {
		age--
	}
	return age
}

// The register gives a code without a date to everyone it registered from July 2017, so the
// category draws no one younger than 9.
func TestLatvianPersonasKodsIsOfSomeoneAged9To99(t *testing.T) {
	eachLatvianPersonasKods(t, func(got string) {
		if age := latvianPersonasKodsAge(got); age < 9 || age > 99 {
			t.Fatalf("%q is of someone aged %d, want 9 to 99 on %s", got, age, fejktest.TestDay.Format("2006-01-02"))
		}
	})
}

// Eurostat's table demo_pjan, the population on 1 January 2025: 1,700,722 people aged 9 to 99,
// of whom 183,940 were 9 to 17, 1,110,183 were 18 to 64 and 406,599 were 65 to 99.
func TestLatvianPersonasKodsFollowsTheAgeBandsOfEurostatsPopulation(t *testing.T) {
	bands := map[string]int{}
	eachLatvianPersonasKods(t, func(got string) {
		switch age := latvianPersonasKodsAge(got); {
		case age < 18:
			bands["9 to 17"]++
		case age < 65:
			bands["18 to 64"]++
		default:
			bands["65 to 99"]++
		}
	})
	for band, want := range map[string]float64{"9 to 17": 10.82, "18 to 64": 65.28, "65 to 99": 23.91} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / latvianPersonasKodsDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the codes, want %.2f%% by Eurostat's population", band, share, want)
			}
		})
	}
}
