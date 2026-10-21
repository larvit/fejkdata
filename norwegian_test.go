package fejkdata

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const ssbCutoff = 200

func TestNorwegianNamesAreSSBsTables(t *testing.T) {
	bySex := map[string]map[string]bool{"f": {}, "m": {}}
	for _, row := range tsvRows(t, "data/nb_NO/first-name.tsv") {
		if bySex[row["sex"]] == nil {
			t.Fatalf("first-name.tsv row %v has a sex other than f and m", row)
		}
		bySex[row["sex"]][row["name"]] = true
		if n, err := strconv.Atoi(row["count"]); err != nil || n < ssbCutoff {
			t.Errorf("first name %v is borne by fewer than %d persons, which SSB does not publish", row, ssbCutoff)
		}
	}
	for name, sex := range map[string]string{"Kari": "f", "Nora": "f", "Ola": "m", "Jakob": "m"} {
		other := map[string]string{"f": "m", "m": "f"}[sex]
		if !bySex[sex][name] || bySex[other][name] {
			t.Errorf("%s is not a name of sex %q alone", name, sex)
		}
	}
	if !bySex["f"]["Kim"] || !bySex["m"]["Kim"] {
		t.Error("Kim is borne by girls and boys, want a row under each sex")
	}

	last := tsvRows(t, "data/nb_NO/last-name.tsv")
	counts := map[string]int{}
	for _, row := range last {
		n, err := strconv.Atoi(row["count"])
		if err != nil || n < ssbCutoff {
			t.Errorf("last name %v is borne by fewer than %d persons, which SSB does not publish", row, ssbCutoff)
		}
		counts[row["name"]] = n
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return counts[names[i]] > counts[names[j]] })
	if want := []string{"Hansen", "Johansen", "Olsen"}; !slices.Equal(names[:3], want) {
		t.Errorf("the three most common last names are %v, want %v", names[:3], want)
	}
}

func TestNorwegianPersonNamesOneSexAndWearsATitleSometimes(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	first := map[string]map[string]bool{"kvinne": {}, "mann": {}}
	for _, row := range tsvRows(t, "data/nb_NO/first-name.tsv") {
		first[map[string]string{"f": "kvinne", "m": "mann"}[row["sex"]]][row["name"]] = true
	}
	last := map[string]bool{}
	for _, row := range tsvRows(t, "data/nb_NO/last-name.tsv") {
		last[row["name"]] = true
	}
	titles := map[string]int{}
	for i := 0; i < 2000; i++ {
		got, err := f.FakeTemplate("{/nb_NO.person as p}{p.sex}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		cells := strings.Split(got, "|")
		if len(cells) != 4 {
			t.Fatalf("%q does not hold four cells", got)
		}
		sex, name, surname, whole := cells[0], cells[1], cells[2], cells[3]
		if !first[sex][name] {
			t.Fatalf("%q: %q is not a first name shipped for %q", got, name, sex)
		}
		if !last[surname] {
			t.Fatalf("%q: %q is not a shipped last name", got, surname)
		}
		title, ok := strings.CutSuffix(whole, name+" "+surname)
		if !ok {
			t.Fatalf("%q: the person does not end with its first and last name", got)
		}
		titles[strings.TrimSpace(title)]++
	}
	if titles["dr."] == 0 || titles["professor"] == 0 || titles[""] < 1800 || len(titles) != 3 {
		t.Errorf("titles over 2000 persons = %v, want a few dr. and professor beside many with none", titles)
	}
}

func TestNorwegianSexAndTitleAreSelectable(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	for tmpl, want := range map[string]string{
		"{/nb_NO.sex[f]}":           "kvinne",
		"{/nb_NO.sex[m]}":           "mann",
		"{/nb_NO.title[dr.]}":       "dr.",
		"{/nb_NO.title[professor]}": "professor",
	} {
		got, err := f.FakeTemplate(tmpl)
		if err != nil || got != want {
			t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
		}
	}
}
