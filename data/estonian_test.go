package data_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestEstonianNamesWeighByRankAndTheRarerOnesShareTheLast(t *testing.T) {
	first, last := nameCounts(t, "et_EE", "first-name"), nameCounts(t, "et_EE", "last-name")
	for _, c := range []struct {
		what  string
		names map[string]int
		want  map[string]int
	}{
		{"girls", first["f"], map[string]int{"Olga": 100000, "Tatjana": 70711, "Svetlana": 57735, "Maria": 50000, "Ljudmila": 33333, "Marika": 31623, "Agnes": 10000, "Zoja": 10000}},
		{"boys", first["m"], map[string]int{"Aleksandr": 100000, "Sergei": 70711, "Vladimir": 57735, "Andrei": 50000, "Ahti": 10483, "Viljar": 10483}},
		{"last names of women", last["f"], map[string]int{"Tamm": 100000, "Saar": 70711, "Sepp": 57735, "Mägi": 50000, "Ivanova": 28868, "Petrova": 27735, "Aavik": 10000}},
		{"last names of men", last["m"], map[string]int{"Tamm": 100000, "Mägi": 50000, "Ivanov": 28868, "Petrov": 27735, "Aavik": 10000}},
	} {
		t.Run(c.what, func(t *testing.T) {
			for name, want := range c.want {
				if got := c.names[name]; got != want {
					t.Errorf("%q weighs %d, want %d", name, got, want)
				}
			}
		})
	}
}

func TestEstonianNamesKeepTheirLetters(t *testing.T) {
	first, last := nameCounts(t, "et_EE", "first-name"), lastNameCounts(t, "et_EE")
	for _, name := range []string{"Põder", "Lõhmus", "Jõgi", "Kütt", "Mägi", "Härm"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
	for sex, names := range map[string][]string{"f": {"Õie", "Kärt", "Ülle"}, "m": {"Jüri", "Tõnu", "Märt"}} {
		for _, name := range names {
			if first[sex][name] == 0 {
				t.Errorf("first-name.tsv has no %q of sex %q", name, sex)
			}
		}
	}
}

func TestEstonianPersonBearsALastNameOfItsSex(t *testing.T) {
	last := nameCounts(t, "et_EE", "last-name")
	bySex := map[string]string{"naine": "f", "mees": "m"}
	eachPerson(t, "et_EE", func(p drawnPerson) {
		if last[bySex[p.sex]][p.last] == 0 {
			t.Fatalf("%q: %q is not a last name shipped for the sex %q", p.raw, p.last, p.sex)
		}
	})
}

func TestEstonianWomenBearTheFeminineFormOfARussianLastName(t *testing.T) {
	last := nameCounts(t, "et_EE", "last-name")
	for _, c := range []struct{ sex, name string }{{"f", "Ivanova"}, {"m", "Ivanov"}, {"f", "Tamm"}, {"m", "Tamm"}} {
		if last[c.sex][c.name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex %q", c.name, c.sex)
		}
	}
	for _, c := range []struct{ sex, name string }{{"f", "Ivanov"}, {"m", "Ivanova"}, {"f", "Aleva"}} {
		if last[c.sex][c.name] != 0 {
			t.Errorf("last-name.tsv has %q of sex %q", c.name, c.sex)
		}
	}
}

const estonianIsikukoodDraws = 20000

func eachEstonianIsikukood(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(estonianIsikukoodDraws, func() string { return fejktest.Fake(t, f, "et_EE.isikukood") }, check)
}

func TestEstonianIsikukoodIsElevenDigitsWithAValidDateAndCheckDigit(t *testing.T) {
	eachEstonianIsikukood(t, func(got string) {
		if !fejktest.ValidIsikukood(got) {
			t.Fatalf("%q is not a valid isikukood", got)
		}
	})
}

func estonianIsikukoodBirth(got string) time.Time {
	year := 1800 + 100*(int(got[0]-'1')/2)
	yy, _ := strconv.Atoi(got[1:3])
	mm, _ := strconv.Atoi(got[3:5])
	dd, _ := strconv.Atoi(got[5:7])
	return time.Date(year+yy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
}

func TestEstonianIsikukoodIsOfSomeoneBornInTheLastHundredYearsNotAfterToday(t *testing.T) {
	eachEstonianIsikukood(t, func(got string) {
		born := estonianIsikukoodBirth(got)
		if born.After(fejktest.TestDay) || born.Before(fejktest.TestDay.AddDate(-100, 0, 0)) {
			t.Fatalf("%q is of someone born %s, want a day in the 100 years up to %s", got, born.Format("2006-01-02"), fejktest.TestDay.Format("2006-01-02"))
		}
	})
}

// Statistics Estonia's table RV0240, the population on 1 January 2026: 1,360,490 people aged 0 to 99,
// of whom 254,482 were under 18, 815,319 were 18 to 64 and 290,689 were 65 to 99.
func TestEstonianIsikukoodFollowsTheAgeBandsOfStatisticsEstoniasPopulation(t *testing.T) {
	bands := map[string]int{}
	eachEstonianIsikukood(t, func(got string) {
		age := fejktest.TestDay.Year() - estonianIsikukoodBirth(got).Year()
		if estonianIsikukoodBirth(got).AddDate(age, 0, 0).After(fejktest.TestDay) {
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
	for band, want := range map[string]float64{"under 18": 18.71, "18 to 64": 59.93, "65 to 99": 21.37} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / estonianIsikukoodDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the codes, want %.2f%% by Statistics Estonia's population", band, share, want)
			}
		})
	}
}

func TestEstonianIsikukoodAgreesWithABoundSex(t *testing.T) {
	f := shipped(t)
	seen := map[string]bool{}
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/et_EE.sex as s}{s.code}|{/et_EE.isikukood.birth as b}{isikukood(s.sex-digit.digit,b)}")
	}, func(got string) {
		sex, code, _ := strings.Cut(got, "|")
		seen[sex] = true
		if want := map[string]int{"m": 1, "f": 0}[sex]; len(code) != 11 || int(code[0]-'0')%2 != want || !fejktest.ValidIsikukood(code) {
			t.Fatalf("%q: want a valid isikukood starting with an odd digit for a man and an even one for a woman", got)
		}
	})
	if !seen["m"] || !seen["f"] {
		t.Errorf("500 codes drew the sexes %v, want both", seen)
	}
}

var (
	estonianFixed  = regexp.MustCompile(`^(3[23589]|4[3-8]|6[0-8]|7[1-9]|88)\d \d{4}$`)
	estonianMobile = regexp.MustCompile(`^5\d{3} \d{4}$`)
)

func estonianFixedPhone(*testing.T) func(n string) bool { return estonianFixed.MatchString }

// The Annex to the Estonian numbering plan, regulation no. 25 of 21 May 2018.
func TestShippedEstonianPhoneLiesInARangeTheNumberingPlanAllocates(t *testing.T) {
	eachPhone(t, "et_EE", func(n string) {
		if !estonianFixed.MatchString(n) && !estonianMobile.MatchString(n) {
			t.Fatalf("phone %q is neither a fixed number in 32 33 35 38 39 43-48 60-68 71-79 88 nor a mobile 5xxx xxxx", n)
		}
	})
}

func TestShippedEstonianFixedPhoneDrawsEachOfItsThirtyStartsEvenly(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(17))
	starts := map[string]int{}
	fejktest.EachDraw(200000, func() string { return fejktest.Fake(t, f, "et_EE.phone") }, func(n string) {
		if estonianFixed.MatchString(n) {
			starts[n[:2]]++
		}
	})
	if len(starts) != 30 {
		t.Fatalf("drew %d fixed starts, want the plan's 30: %v", len(starts), starts)
	}
	for start, n := range starts {
		if n < 480 || n > 730 {
			t.Errorf("start %s drew %d of about 18182 fixed numbers, want about 606 give or take 125", start, n)
		}
	}
}
