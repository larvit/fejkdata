package data_test

import (
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
var inseeLastNames = map[string]int{"Martin": 250013, "Bernard": 131330, "Lefevre": 64107, "da Silva": 28730, "Le Gall": 27857, "de Oliveira": 8253}

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

func TestFrenchFirstNamesHoldTheMostBorneTwoThousandOfEachSex(t *testing.T) {
	for sex, names := range nameCounts(t, "fr_FR", "first-name") {
		if len(names) != 2000 {
			t.Errorf("sex %q holds %d first names, want the 2000 most borne", sex, len(names))
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

func TestFrenchLastNamesHoldTheMostBorneFiveThousand(t *testing.T) {
	if got := len(nameCounts(t, "fr_FR", "last-name")[""]); got != 5000 {
		t.Errorf("last-name.tsv holds %d names, want the 5000 most borne", got)
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
