package data_test

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

var (
	latvianFixed  = regexp.MustCompile(`^6\d \d{3} \d{3}$`)
	latvianMobile = regexp.MustCompile(`^2\d \d{3} \d{3}$`)
)

func latvianFixedPhone(*testing.T) func(n string) bool { return latvianFixed.MatchString }

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

func TestLatvianPersonasKodsIsElevenDigitsWithAValidCheckDigit(t *testing.T) {
	eachLatvianPersonasKods(t, func(got string) {
		if !fejktest.ValidPersonasKods(got) {
			t.Fatalf("%q is not a valid personas kods", got)
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

// A code with a date is of someone born before the register's change on 1 July 2017 and not more
// than 100 years ago.
func TestLatvianPersonasKodsWithADateIsOfSomeoneBornBeforeJuly2017Within100Years(t *testing.T) {
	dated := 0
	eachLatvianPersonasKods(t, func(got string) {
		if got[:2] == "32" {
			return
		}
		dated++
		born := latvianPersonasKodsBirth(got)
		if !born.Before(time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)) || born.Before(fejktest.TestDay.AddDate(-100, 0, 0)) {
			t.Fatalf("%q is of someone born %s, want a day from %s to 30 June 2017", got, born.Format("2006-01-02"), fejktest.TestDay.AddDate(-100, 0, 0).Format("2006-01-02"))
		}
	})
	if dated == 0 {
		t.Error("no code had a date, want most")
	}
}

// Eurostat's table demo_pjan, the population on 1 January 2025: 1,860,337 people aged 0 to 99,
// of whom 343,555 were under 18, 1,110,183 were 18 to 64 and 406,599 were 65 to 99. A code with no
// date is of someone under 18.
func TestLatvianPersonasKodsFollowsTheAgeBandsOfEurostatsPopulation(t *testing.T) {
	bands := map[string]int{}
	eachLatvianPersonasKods(t, func(got string) {
		switch {
		case got[:2] == "32", latvianPersonasKodsAge(got) < 18:
			bands["under 18"]++
		case latvianPersonasKodsAge(got) < 65:
			bands["18 to 64"]++
		default:
			bands["65 to 99"]++
		}
	})
	for band, want := range map[string]float64{"under 18": 18.47, "18 to 64": 59.68, "65 to 99": 21.86} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / latvianPersonasKodsDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the codes, want %.2f%% by Eurostat's population", band, share, want)
			}
		})
	}
}

func TestLatvianPersonasKodsHasNoDateForSomeoneBornFromJuly2017(t *testing.T) {
	undated := 0
	eachLatvianPersonasKods(t, func(got string) {
		if got[:2] == "32" {
			undated++
		}
	})
	if share := 100 * float64(undated) / latvianPersonasKodsDraws; share < 7 || share > 11.5 {
		t.Errorf("%.1f%% of the codes had no date, want about 9%%", share)
	}
}

// Table 1 of the National Numbering Plan, decision no. 1/33 of 27 September 2022 of the Public Utilities Commission.
func TestShippedLatvianPhoneLiesInARangeTheNumberingPlanAllocates(t *testing.T) {
	eachPhone(t, "lv_LV", func(n string) {
		if !latvianFixed.MatchString(n) && !latvianMobile.MatchString(n) {
			t.Fatalf("phone %q is neither a fixed number 6x xxx xxx nor a mobile 2x xxx xxx", n)
		}
	})
}

func TestShippedLatvianPhoneDrawsEachSecondDigitOfBothRanges(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(17))
	starts := map[string]int{}
	fejktest.EachDraw(20000, func() string { return fejktest.Fake(t, f, "lv_LV.phone") }, func(n string) {
		starts[n[:2]]++
	})
	for _, lead := range []string{"2", "6"} {
		for d := 0; d <= 9; d++ {
			if start := lead + strconv.Itoa(d); starts[start] == 0 {
				t.Errorf("no phone starts with %s in 20000 draws: %v", start, starts)
			}
		}
	}
}
