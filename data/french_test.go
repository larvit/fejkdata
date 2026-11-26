package data_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The sums of the counts that INSEE's "Fichier des prénoms" publishes for the births of 1944 to 2023,
// each year's rows added up for the name and sex.
var inseeFirstNames = map[string]map[string]int{
	"f": {"Marie": 751837, "Nathalie": 379515, "Isabelle": 361448, "Léa": 179442, "Hélène": 144812, "Camille": 186445, "Dominique": 164957},
	"m": {"Jean": 771461, "Michel": 519386, "Frédéric": 312551, "Jean-Pierre": 189570, "Camille": 34693, "Dominique": 226341},
}

// The sums of the counts that INSEE's "Fichier des noms" publishes for the births of 1891 to 2000.
var inseeLastNames = map[string]int{"Martin": 250013, "Bernard": 131330, "Lefevre": 64107, "da Silva": 28730, "dos Santos": 17661, "Le Gall": 27857, "de Oliveira": 8253}

func TestFrenchFirstNamesAreInseesCounts(t *testing.T) {
	first := nameCounts(t, "fr_FR", "first-name")
	for sex, names := range inseeFirstNames {
		for name, want := range names {
			t.Run(sex+"/"+name, func(t *testing.T) {
				if got := first[sex][name]; got != want {
					t.Errorf("%q of sex %q is borne by %d, want INSEE's %d", name, sex, got, want)
				}
			})
		}
	}
}

func TestFrenchLastNamesAreInseesCounts(t *testing.T) {
	last := nameCounts(t, "fr_FR", "last-name")[""]
	for name, want := range inseeLastNames {
		t.Run(name, func(t *testing.T) {
			if got := last[name]; got != want {
				t.Errorf("%q is borne by %d, want INSEE's %d", name, got, want)
			}
		})
	}
}

func TestFrenchFirstNamesFollowTheirCounts(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	seen := fejktest.TallyDraws(20000, func() string { return fejktest.FakeTemplate(t, f, "{/fr_FR.sex[m].first-name}") })
	if seen["Jean"] < 2*seen["Frédéric"] || seen["Frédéric"] == 0 {
		t.Errorf("Jean and Frédéric drawn %d and %d times in 20000, want Jean about 2.5 times Frédéric, as 771461 is to 312551", seen["Jean"], seen["Frédéric"])
	}
}

func TestFrenchFirstNameUnderItsSexReadsItsCount(t *testing.T) {
	f := shipped(t)
	if got := fejktest.FakeTemplate(t, f, "{/fr_FR.sex[m].first-name[Jean].count}"); got != "771461" {
		t.Errorf("Jean under the sex m has the count %q, want INSEE's 771461", got)
	}
}

const frenchNirDraws = 20000

func eachFrenchNir(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(frenchNirDraws, func() string { return fejktest.Fake(t, f, "fr_FR.nir") }, check)
}

// frenchNirDepartment is the department of a NIR: three characters where it is overseas, else two.
func frenchNirDepartment(got string) string {
	if got[5:7] == "97" {
		return got[5:8]
	}
	return got[5:7]
}

func TestFrenchNirIsFifteenCharactersWithAValidKey(t *testing.T) {
	eachFrenchNir(t, func(got string) {
		if !fejktest.ValidNir(got) {
			t.Fatalf("%q is not a NIR of 15 characters with a valid key", got)
		}
	})
}

func TestFrenchNirNamesAHexagonalOrOverseasDepartment(t *testing.T) {
	seen := map[string]bool{}
	eachFrenchNir(t, func(got string) {
		dep := frenchNirDepartment(got)
		seen[dep] = true
		if n, err := strconv.Atoi(dep); err == nil && (n == 0 || n == 20 || (n > 95 && (n < 971 || n > 976 || n == 975))) {
			t.Fatalf("%q names the department %s, want one of INSEE's", got, dep)
		}
	})
	for _, want := range []string{"01", "2A", "2B", "75", "95", "971", "974", "976"} {
		if !seen[want] {
			t.Errorf("%d numbers never named the department %s, want every department drawn", frenchNirDraws, want)
		}
	}
}

func TestFrenchNirCommuneAndOrderNumberAreNeverZero(t *testing.T) {
	eachFrenchNir(t, func(got string) {
		commune := got[7:10]
		if frenchNirDepartment(got) != got[5:7] {
			commune = got[8:10]
		}
		if strings.Trim(commune, "0") == "" || got[10:13] == "000" {
			t.Fatalf("%q has the commune %s and the order number %s, want INSEE's 001 to 999", got, commune, got[10:13])
		}
	})
}

func monthsSinceFrenchNirBirth(got string) int {
	yy, _ := strconv.Atoi(got[1:3])
	mm, _ := strconv.Atoi(got[3:5])
	year := fejktest.TestDay.Year() - ((fejktest.TestDay.Year()-yy)%100+100)%100
	return (fejktest.TestDay.Year()-year)*12 + int(fejktest.TestDay.Month()) - mm
}

func TestFrenchNirIsBornInAMonthOfTheCalendarNotAfterToday(t *testing.T) {
	eachFrenchNir(t, func(got string) {
		if mm, _ := strconv.Atoi(got[3:5]); mm < 1 || mm > 12 || monthsSinceFrenchNirBirth(got) < 0 {
			t.Fatalf("%q has the month %s, want a month of the calendar not after %s", got, got[3:5], fejktest.TestDay.Format("2006-01"))
		}
	})
}

func frenchNirAgeBand(got string) string {
	switch age := monthsSinceFrenchNirBirth(got) / 12; {
	case age < 18:
		return "under 18"
	case age < 65:
		return "18 to 64"
	default:
		return "65 to 99"
	}
}

// INSEE's table "Population au 1er janvier par sexe et âge détaillé" of the Bilan démographique 2025, the
// population of France on 1 January 2026: 69,081,996 people, of whom 13,852,666 were under 18, 39,890,078
// were 18 to 64 and 15,339,252 were 65 or more.
func TestFrenchNirFollowsTheAgeBandsOfInseesPopulation(t *testing.T) {
	bands := map[string]int{}
	eachFrenchNir(t, func(got string) { bands[frenchNirAgeBand(got)]++ })
	for band, want := range map[string]float64{"under 18": 20.05, "18 to 64": 57.74, "65 to 99": 22.20} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / frenchNirDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the numbers, want %.2f%% by INSEE's population", band, share, want)
			}
		})
	}
}

func TestFrenchNirStartsWithItsSexAndAgreesWithABoundOne(t *testing.T) {
	f := shipped(t)
	seen := map[string]bool{}
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/fr_FR.sex as s}{s.code}|{s.sex-digit.digit}{/fr_FR.nir.body}{nir()}")
	}, func(got string) {
		sex, number, _ := strings.Cut(got, "|")
		seen[sex] = true
		if want := map[string]byte{"m": '1', "f": '2'}[sex]; len(number) != 15 || number[0] != want || !fejktest.ValidNir(number) {
			t.Fatalf("%q: want a valid NIR starting with 1 for a man and 2 for a woman", got)
		}
	})
	if !seen["m"] || !seen["f"] {
		t.Errorf("500 numbers drew the sexes %v, want both", seen)
	}
}

func TestFrenchSentenceSpacesAQuestionOrExclamationMarkFromItsWord(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(7))
	tight := regexp.MustCompile(`[^ ][!?]$`)
	seen := map[string]bool{}
	fejktest.EachDraw(400, func() string { return fejktest.Fake(t, f, "fr_FR.sentence") }, func(s string) {
		if tight.MatchString(s) {
			t.Fatalf("fr_FR.sentence = %q, want a space before its closing ! or ?, as French sets it", s)
		}
		seen[s[len(s)-1:]] = true
	})
	for _, mark := range []string{"!", "?"} {
		if !seen[mark] {
			t.Errorf("400 sentences never ended in %q", mark)
		}
	}
}
