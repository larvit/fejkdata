package fejkdata

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Statistics Denmark's top 20 of the names that people living in Denmark bore on 1 January 2026,
// its "Names in the entire population" page.
var (
	dstGirls = map[string]int{
		"Anne": 43021, "Mette": 37932, "Kirsten": 34588, "Hanne": 34482, "Anna": 33441,
		"Helle": 32720, "Maria": 30400, "Susanne": 29668, "Lene": 29290, "Marianne": 25598,
		"Camilla": 24494, "Lone": 23918, "Louise": 23824, "Charlotte": 23455, "Pia": 23359,
		"Tina": 22855, "Emma": 22590, "Ida": 22571, "Gitte": 22382, "Julie": 21912,
	}
	dstBoys = map[string]int{
		"Peter": 45777, "Michael": 43976, "Lars": 42596, "Thomas": 41800, "Jens": 40558,
		"Henrik": 40416, "Søren": 38087, "Christian": 37349, "Martin": 36799, "Jan": 35210,
		"Morten": 33384, "Jesper": 33139, "Anders": 32980, "Mads": 32239, "Niels": 30836,
		"Rasmus": 30222, "Mikkel": 29280, "Kim": 27770, "Per": 27640, "Ole": 26195,
	}
	dstLast = map[string]int{
		"Nielsen": 222355, "Jensen": 218886, "Hansen": 186105, "Andersen": 143880, "Pedersen": 140838,
		"Christensen": 106860, "Larsen": 102287, "Sørensen": 97587, "Rasmussen": 84147, "Jørgensen": 78057,
		"Petersen": 68658, "Madsen": 58255, "Kristensen": 55775, "Olsen": 42709, "Thomsen": 37486,
		"Christiansen": 33781, "Poulsen": 29488, "Johansen": 28983, "Møller": 28916, "Mortensen": 27141,
	}
)

// danishNames reads a names table as the counts of its rows, per sex for first names.
func danishNames(t *testing.T, file string) map[string]map[string]int {
	t.Helper()
	out := map[string]map[string]int{}
	for _, row := range tsvRows(t, "data/da_DK/"+file+".tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil || n < 1 {
			t.Fatalf("%s row %v has no count", file, row)
		}
		sex := row["sex"]
		if out[sex] == nil {
			out[sex] = map[string]int{}
		}
		if _, dup := out[sex][row["name"]]; dup {
			t.Fatalf("%s holds %q twice for sex %q", file, row["name"], sex)
		}
		out[sex][row["name"]] = n
	}
	return out
}

func TestDanishTopNamesAreStatisticsDenmarksCounts(t *testing.T) {
	first, last := danishNames(t, "first-name"), danishNames(t, "last-name")[""]
	for _, c := range []struct {
		what string
		got  map[string]int
		want map[string]int
	}{{"girls", first["f"], dstGirls}, {"boys", first["m"], dstBoys}, {"last names", last, dstLast}} {
		for name, want := range c.want {
			if c.got[name] != want {
				t.Errorf("%s: %q is borne by %d, want Statistics Denmark's %d", c.what, name, c.got[name], want)
			}
		}
	}
}

func TestDanishOtherNamesAreRarerThanTheTwentieth(t *testing.T) {
	first, last := danishNames(t, "first-name"), danishNames(t, "last-name")[""]
	for _, c := range []struct {
		what string
		got  map[string]int
		top  map[string]int
		min  int
	}{{"girls", first["f"], dstGirls, 200}, {"boys", first["m"], dstBoys, 200}, {"last names", last, dstLast, 200}} {
		twentieth := 1 << 30
		for _, n := range c.top {
			twentieth = min(twentieth, n)
		}
		if len(c.got) < c.min {
			t.Errorf("%s: %d names, want %d or more", c.what, len(c.got), c.min)
		}
		for name, n := range c.got {
			if _, top := c.top[name]; !top && n >= twentieth {
				t.Errorf("%s: %q is borne by %d, but Statistics Denmark's twentieth is borne by %d, so %q would be in its top 20", c.what, name, n, twentieth, name)
			}
		}
	}
}

func TestDanishNamesAreLettersAndAGirlsNameIsNotABoysByAccident(t *testing.T) {
	name := regexp.MustCompile(`^[\pL]{2,}([ '-][\pL]{2,})*$`)
	for _, file := range []string{"first-name", "last-name"} {
		for _, row := range tsvRows(t, "data/da_DK/"+file+".tsv") {
			if !name.MatchString(row["name"]) {
				t.Errorf("%s row %v is not a name in letters", file, row)
			}
		}
	}
	first := danishNames(t, "first-name")
	if len(first) != 2 || first["f"] == nil || first["m"] == nil {
		t.Fatalf("first-name.tsv holds sexes %v, want f and m", first)
	}
	for name, sex := range map[string]string{"Mette": "f", "Freja": "f", "Søren": "m", "Mads": "m"} {
		other := map[string]string{"f": "m", "m": "f"}[sex]
		if first[sex][name] == 0 || first[other][name] != 0 {
			t.Errorf("%s is not a name of sex %q alone", name, sex)
		}
	}
	if first["f"]["Kim"] == 0 || first["m"]["Kim"] == 0 {
		t.Error("Kim is borne by girls and boys, want a row under each sex")
	}
}

func TestDanishPersonIsATitleOrNoneAFirstNameOfItsSexAndALastName(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	firsts := danishNames(t, "first-name")
	first := map[string]map[string]int{"kvinde": firsts["f"], "mand": firsts["m"]}
	last := danishNames(t, "last-name")[""]
	titles := map[string]int{}
	for i := 0; i < 2000; i++ {
		got, err := f.FakeTemplate("{/da_DK.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		sex, prefix, name, surname, whole := cells[0], cells[1], cells[2], cells[3], cells[4]
		if first[sex][name] == 0 {
			t.Fatalf("%q: %q is not a first name shipped for the sex %q, which is kvinde or mand", got, name, sex)
		}
		if last[surname] == 0 {
			t.Fatalf("%q: %q is not a shipped last name", got, surname)
		}
		if whole != prefix+name+" "+surname {
			t.Fatalf("%q: the person is not its prefix, first name and last name", got)
		}
		titles[prefix]++
	}
	if titles["dr. "] == 0 || titles["professor "] == 0 || titles[""] < 1800 || len(titles) != 3 {
		t.Errorf("prefixes over 2000 persons = %q, want a few %q and %q beside many with none", titles, "dr. ", "professor ")
	}
}

func TestDanishNamesFollowTheirCounts(t *testing.T) {
	f, err := New(withShipped(), WithSeed(5), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for i := 0; i < 20000; i++ {
		seen[fakeTemplate(t, f, "{/da_DK.last-name}")]++
	}
	if seen["Nielsen"] < 3*seen["Olsen"] || seen["Jensen"] < 3*seen["Olsen"] || seen["Nielsen"] == 0 {
		t.Errorf("Nielsen, Jensen and Olsen drawn %d, %d and %d times in 20000, want Nielsen and Jensen about five times Olsen", seen["Nielsen"], seen["Jensen"], seen["Olsen"])
	}
}

func TestDanishSexAndTitleAreSelectable(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	for tmpl, want := range map[string]string{
		"{/da_DK.sex[f]}":           "kvinde",
		"{/da_DK.sex[m]}":           "mand",
		"{/da_DK.title[dr.]}":       "dr.",
		"{/da_DK.title[professor]}": "professor",
	} {
		got, err := f.FakeTemplate(tmpl)
		if err != nil || got != want {
			t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
		}
	}
}
