package data_test

import (
	"regexp"
	"strconv"
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

func germanNames(t *testing.T, file string) map[string]map[string]int {
	t.Helper()
	out := map[string]map[string]int{}
	for _, row := range fejktest.TSVRows(t, "de_DE/"+file+".tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil || n < 1 {
			t.Fatalf("%s row %v has no count", file, row)
		}
		if out[row["sex"]] == nil {
			out[row["sex"]] = map[string]int{}
		}
		if _, dup := out[row["sex"]][row["name"]]; dup {
			t.Fatalf("%s holds %q twice for sex %q", file, row["name"], row["sex"])
		}
		out[row["sex"]][row["name"]] = n
	}
	return out
}

func TestGermanFirstNamesAreMunichsCounts(t *testing.T) {
	first := germanNames(t, "first-name")
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
	first := germanNames(t, "first-name")
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
	first, last := germanNames(t, "first-name"), germanNames(t, "last-name")[""]
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
	last := germanNames(t, "last-name")[""]
	for _, name := range []string{"Müller", "Schröder", "Köhler", "Weiß"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
}

func TestGermanLastNamesWeighSevenHundredThousandOverTheRootOfTheirRank(t *testing.T) {
	last := germanNames(t, "last-name")[""]
	for name, want := range map[string]int{"Müller": 700000, "Schmidt": 494975, "Schneider": 404145, "Schlegel": 47194, "Abel": 47194} {
		t.Run(name, func(t *testing.T) {
			if got := last[name]; got != want {
				t.Errorf("%q weighs %d, want %d: 700000 over the square root of rank 1, 2, 3 and 220, and the rarer names the 220th's", name, got, want)
			}
		})
	}
}

type germanPerson struct{ raw, sex, prefix, first, last, whole string }

func eachGermanPerson(t *testing.T, check func(p germanPerson)) {
	t.Helper()
	f, err := fejkdata.New(withShipped(), fejkdata.WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	fejktest.EachDraw(2000, func() string {
		got, err := f.FakeTemplate("{/de_DE.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		check(germanPerson{got, cells[0], cells[1], cells[2], cells[3], cells[4]})
	})
}

func TestGermanPersonHasAFirstNameOfItsSex(t *testing.T) {
	firsts := germanNames(t, "first-name")
	first := map[string]map[string]int{"Frau": firsts["f"], "Mann": firsts["m"]}
	eachGermanPerson(t, func(p germanPerson) {
		if first[p.sex][p.first] == 0 {
			t.Fatalf("%q: %q is not a first name shipped for the sex %q", p.raw, p.first, p.sex)
		}
	})
}

func TestGermanPersonHasAShippedLastName(t *testing.T) {
	last := germanNames(t, "last-name")[""]
	eachGermanPerson(t, func(p germanPerson) {
		if last[p.last] == 0 {
			t.Fatalf("%q: %q is not a shipped last name", p.raw, p.last)
		}
	})
}

func TestGermanPersonIsItsPrefixFirstNameAndLastName(t *testing.T) {
	eachGermanPerson(t, func(p germanPerson) {
		if p.whole != p.prefix+p.first+" "+p.last {
			t.Fatalf("%q: the person is not its prefix, first name and last name", p.raw)
		}
	})
}

func TestGermanPersonMostlyBearsNoTitle(t *testing.T) {
	titles := map[string]int{}
	eachGermanPerson(t, func(p germanPerson) { titles[p.prefix]++ })
	if titles["Dr. "] == 0 || titles["Prof. "] == 0 || titles["Prof. Dr. "] == 0 || titles[""] < 1800 || len(titles) != 4 {
		t.Errorf("prefixes over 2000 persons = %q, want a few of Dr., Prof. and Prof. Dr. beside many with none", titles)
	}
}

func TestGermanFirstNamesFollowTheirCounts(t *testing.T) {
	f, err := fejkdata.New(withShipped(), fejkdata.WithSeed(5), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	seen := fejktest.TallyDraws(20000, func() string { return fejktest.FakeTemplate(t, f, "{/de_DE.sex[m].first-name}") })
	if 2*seen["Maximilian"] < 3*seen["Luca"] || seen["Luca"] == 0 {
		t.Errorf("Maximilian and Luca drawn %d and %d times in 20000, want Maximilian about 2.5 times Luca, as 1751 is to 710", seen["Maximilian"], seen["Luca"])
	}
}

func TestGermanSexAndTitleAreSelectable(t *testing.T) {
	f, err := fejkdata.New(withShipped(), fejkdata.WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
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
