package fejkdata

import (
	"strconv"
	"strings"
	"testing"
)

// DVV's "Etunimitilasto" and "Sukunimitilasto" of 4 August 2026: the first given names of the
// Finnish citizens in its population register, and their last names.
var (
	dvvWomen = map[string]int{"Anne": 29746, "Päivi": 29329, "Tuula": 28486, "Anna": 27794, "Minna": 27424}
	dvvMen   = map[string]int{"Juha": 42928, "Timo": 41616, "Mikko": 37134, "Matti": 36668, "Jari": 36444}
	dvvLast  = map[string]int{"Korhonen": 21144, "Virtanen": 20295, "Mäkinen": 18833, "Nieminen": 18669, "Mäkelä": 18117}
)

func finnishCounts(t *testing.T, file string) map[string]map[string]int {
	t.Helper()
	out := map[string]map[string]int{}
	for _, row := range tsvRows(t, "data/fi_FI/"+file+".tsv") {
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

func TestFinnishTopNamesAreDVVsCounts(t *testing.T) {
	first, last := finnishCounts(t, "first-name"), finnishCounts(t, "last-name")[""]
	for _, c := range []struct {
		what      string
		got, want map[string]int
	}{{"women", first["f"], dvvWomen}, {"men", first["m"], dvvMen}, {"last names", last, dvvLast}} {
		for name, want := range c.want {
			if c.got[name] != want {
				t.Errorf("%s: %q is borne by %d, want DVV's %d", c.what, name, c.got[name], want)
			}
		}
	}
}

func TestFinnishNamesAreBorneByTheFewestDVVPublishes(t *testing.T) {
	for file, fewest := range map[string]int{"first-name": 5, "last-name": 20} {
		counts := finnishCounts(t, file)
		total := 0
		for _, byName := range counts {
			total += len(byName)
			for name, n := range byName {
				if n < fewest {
					t.Errorf("%s: %q is borne by %d, which DVV does not publish below %d", file, name, n, fewest)
				}
			}
		}
		if total < 2000 {
			t.Errorf("%s holds %d names, want 2000 or more", file, total)
		}
	}
}

func TestFinnishFirstNamesAreRowsUnderTheSexesThatBearThem(t *testing.T) {
	first := finnishCounts(t, "first-name")
	if len(first) != 2 || first["f"] == nil || first["m"] == nil {
		t.Fatalf("first-name.tsv holds sexes %v, want f and m", first)
	}
	if first["f"]["Kim"] != 146 || first["m"]["Kim"] != 4872 {
		t.Errorf("Kim is borne by %d girls and %d boys, want DVV's 146 and 4872", first["f"]["Kim"], first["m"]["Kim"])
	}
	if first["m"]["Anne"] != 0 || first["f"]["Juha"] != 0 {
		t.Error("Anne is a boy's name or Juha a girl's, want each under the sex DVV counts it for")
	}
}

func TestFinnishPersonIsATitleOrNoneAFirstNameOfItsSexAndALastName(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	first := map[string]map[string]int{}
	for sex, name := range map[string]string{"f": "nainen", "m": "mies"} {
		first[name] = finnishCounts(t, "first-name")[sex]
	}
	last := finnishCounts(t, "last-name")[""]
	titles := map[string]int{}
	for i := 0; i < 2000; i++ {
		got, err := f.FakeTemplate("{/fi_FI.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		sex, prefix, name, surname, whole := cells[0], cells[1], cells[2], cells[3], cells[4]
		if first[sex][name] == 0 {
			t.Fatalf("%q: %q is not a first name shipped for the sex %q, which is nainen or mies", got, name, sex)
		}
		if last[surname] == 0 {
			t.Fatalf("%q: %q is not a shipped last name", got, surname)
		}
		if whole != prefix+name+" "+surname {
			t.Fatalf("%q: the person is not its prefix, first name and last name", got)
		}
		titles[prefix]++
	}
	if titles["tohtori "] == 0 || titles["professori "] == 0 || titles[""] < 1800 || len(titles) != 3 {
		t.Errorf("prefixes over 2000 persons = %q, want a few %q and %q beside many with none", titles, "tohtori ", "professori ")
	}
}

func TestFinnishSexAndTitleAreSelectable(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	for tmpl, want := range map[string]string{
		"{/fi_FI.sex[f]}":            "nainen",
		"{/fi_FI.sex[m]}":            "mies",
		"{/fi_FI.title[tohtori]}":    "tohtori",
		"{/fi_FI.title[professori]}": "professori",
	} {
		got, err := f.FakeTemplate(tmpl)
		if err != nil || got != want {
			t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
		}
	}
}

func TestFinnishSexesAreWeighedByStatisticsFinlandsPopulationOf2025(t *testing.T) {
	counts := map[string]int{}
	for _, row := range tsvRows(t, "data/fi_FI/sex.tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil {
			t.Fatalf("sex.tsv row %v has no count", row)
		}
		counts[row["code"]+" "+row["name"]] = n
	}
	want := map[string]int{"f nainen": 2853669, "m mies": 2799212}
	if len(counts) != 2 || counts["f nainen"] != want["f nainen"] || counts["m mies"] != want["m mies"] {
		t.Errorf("sex.tsv counts %v, want Statistics Finland's population of 31 December 2025, %v", counts, want)
	}
}
