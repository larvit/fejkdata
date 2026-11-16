package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The sums of the counts that the city of Munich's open data "Vornamen von Neugeborenen" publishes
// for the years 2013 to 2025, each file's rows added up for the name and sex.
var munichCounts = map[string]map[string]int{
	"f": {"Emilia": 1443, "Emma": 1290, "Anna": 1261, "Mia": 967, "Maria": 581, "Noa": 35},
	"m": {"Maximilian": 1751, "Felix": 1447, "Noah": 1166, "Alexander": 921, "Luca": 710, "Noa": 51},
}

func TestGermanFirstNamesAreMunichsCounts(t *testing.T) {
	first := nameCounts(t, "de_DE", "first-name")
	if len(first) != 2 {
		t.Fatalf("first-name.tsv holds sexes %v, want f and m", first)
	}
	for sex, names := range munichCounts {
		for name, want := range names {
			t.Run(sex+"/"+name, func(t *testing.T) {
				if got := first[sex][name]; got != want {
					t.Errorf("%q of sex %q is borne by %d, want Munich's %d", name, sex, got, want)
				}
			})
		}
	}
}

func TestGermanAGirlsNameIsNotABoysByAccident(t *testing.T) {
	first := nameCounts(t, "de_DE", "first-name")
	for name, sex := range map[string]string{"Emilia": "f", "Marie": "f", "Maximilian": "m", "Felix": "m"} {
		t.Run(name, func(t *testing.T) {
			other := map[string]string{"f": "m", "m": "f"}[sex]
			if first[sex][name] == 0 || first[other][name] != 0 {
				t.Errorf("%s is not a name of sex %q alone", name, sex)
			}
		})
	}
}

func TestGermanNameTablesHoldTwoHundredNamesOrMore(t *testing.T) {
	first, last := nameCounts(t, "de_DE", "first-name"), nameCounts(t, "de_DE", "last-name")[""]
	for what, got := range map[string]map[string]int{"girls": first["f"], "boys": first["m"], "last names": last} {
		if len(got) < 200 {
			t.Errorf("%s: %d names, want 200 or more", what, len(got))
		}
	}
}

func TestGermanNamesAreLetters(t *testing.T) {
	name := regexp.MustCompile(`^[\pL]{2,}([ '-][\pL]{2,})*$`)
	for _, file := range []string{"first-name", "last-name"} {
		for _, row := range fejktest.TSVRows(t, "de_DE/"+file+".tsv") {
			if !name.MatchString(row["name"]) {
				t.Errorf("%s row %v is not a name in letters", file, row)
			}
		}
	}
}

func TestGermanLastNamesKeepTheirUmlauts(t *testing.T) {
	last := nameCounts(t, "de_DE", "last-name")[""]
	for _, name := range []string{"Müller", "Schröder", "Köhler", "Weiß"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
}

func TestGermanLastNamesWeighSevenHundredThousandOverTheRootOfTheirRank(t *testing.T) {
	last := nameCounts(t, "de_DE", "last-name")[""]
	for name, want := range map[string]int{"Müller": 700000, "Schmidt": 494975, "Schneider": 404145, "Schlegel": 47194, "Abel": 47194} {
		t.Run(name, func(t *testing.T) {
			if got := last[name]; got != want {
				t.Errorf("%q weighs %d, want %d: 700000 over the square root of rank 1, 2, 3 and 220, and the rarer names the 220th's", name, got, want)
			}
		})
	}
}

func TestGermanFirstNamesFollowTheirCounts(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	seen := fejktest.TallyDraws(20000, func() string { return fejktest.FakeTemplate(t, f, "{/de_DE.sex[m].first-name}") })
	if 2*seen["Maximilian"] < 3*seen["Luca"] || seen["Luca"] == 0 {
		t.Errorf("Maximilian and Luca drawn %d and %d times in 20000, want Maximilian about 2.5 times Luca, as 1751 is to 710", seen["Maximilian"], seen["Luca"])
	}
}

func TestGermanSexAndTitleAreSelectable(t *testing.T) {
	f := shipped(t)
	for tmpl, want := range map[string]string{
		"{/de_DE.sex[f]}":                              "Frau",
		"{/de_DE.sex[m]}":                              "Mann",
		"{/de_DE.title[Dr.]}":                          "Dr.",
		"{/de_DE.title[Prof.]}":                        "Prof.",
		"{/de_DE.title[Prof. Dr.]}":                    "Prof. Dr.",
		"{/de_DE.sex[m].first-name[Maximilian].count}": "1751",
	} {
		t.Run(tmpl, func(t *testing.T) {
			got, err := f.FakeTemplate(tmpl)
			if err != nil || got != want {
				t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
			}
		})
	}
}

func TestGermanSteuerIDIsAValidTestNumber(t *testing.T) {
	f := shipped(t)
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "de_DE.steuer-id") }, func(got string) {
		if !fejktest.ValidSteuerID(got) || got[0] != '0' {
			t.Fatalf("de_DE.steuer-id = %q, want a valid number that starts with the 0 ELSTER keeps for test numbers", got)
		}
	})
}

func TestShippedGermanCompaniesAreMostlyDistinct(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(6))
	seen := fejktest.TallyDraws(500, func() string { return fejktest.Fake(t, f, "de_DE.company") })
	if len(seen) < 400 {
		t.Errorf("500 companies drew %d distinct, want 400 or more", len(seen))
	}
}

func TestShippedGermanEmailsFoldUmlauts(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	seen := fejktest.TallyDraws(5000, func() string { return fejktest.Fake(t, f, "de_DE.email") })
	folded := false
	for got := range seen {
		folded = folded || strings.Contains(got, "muller") || strings.Contains(got, "schroder") || strings.Contains(got, "konig")
	}
	if !folded {
		t.Error("5000 emails held no Müller, Schröder or König folded to ASCII, want one")
	}
}
