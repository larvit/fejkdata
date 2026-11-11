package fejkdata

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestFinnishFirstNamesAreRowsUnderTheSexesThatBearThem(t *testing.T) {
	first := finnishCounts(t, "first-name")
	if len(first) != 2 || first["f"] == nil || first["m"] == nil {
		t.Fatalf("first-name.tsv holds sexes %v, want f and m", first)
	}
	if first["f"]["Kim"] != 146 || first["m"]["Kim"] != 4872 {
		t.Errorf("Kim is borne by %d women and %d men, want DVV's 146 and 4872", first["f"]["Kim"], first["m"]["Kim"])
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

// DVV's "Personal identity code" page: ddmmyy, a century sign, an individual number and a control
// character, which is the remainder of the nine digits divided by 31 read in its table.
const dvvControlTable = "0123456789ABCDEFHJKLMNPRSTUVWXY"

var hetuShape = regexp.MustCompile(`^(\d{2})(\d{2})(\d{2})([+\-A])(\d{3})([0-9A-HJ-NPR-Y])$`)

type finnishBirth struct {
	born       time.Time
	age        int
	individual int
}

func parseHenkilotunnus(t *testing.T, got string) finnishBirth {
	t.Helper()
	m := hetuShape.FindStringSubmatch(got)
	if m == nil {
		t.Fatalf("%q is not ddmmyy, a sign, three digits and a control character", got)
	}
	n := func(s string) int { v, _ := strconv.Atoi(s); return v }
	digits := n(got[:6] + got[7:10])
	if want := dvvControlTable[digits%31]; got[10] != want {
		t.Fatalf("%q has the control character %c, want %c", got, got[10], want)
	}
	year := n(m[3]) + map[string]int{"+": 1800, "-": 1900, "A": 2000}[m[4]]
	born := time.Date(year, time.Month(n(m[2])), n(m[1]), 0, 0, 0, 0, time.UTC)
	if born.Day() != n(m[1]) {
		t.Fatalf("%q is not a day of the calendar", got)
	}
	age := testDay.Year() - year
	if born.AddDate(age, 0, 0).After(testDay) {
		age--
	}
	return finnishBirth{born, age, n(m[5])}
}

const finnishHenkilotunnusDraws = 20000

func eachFinnishHenkilotunnus(t *testing.T, check func(got string, b finnishBirth)) {
	t.Helper()
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	eachDraw(finnishHenkilotunnusDraws, func() string { return fake(t, f, "fi_FI.henkilotunnus") }, func(got string) {
		check(got, parseHenkilotunnus(t, got))
	})
}

func TestFinnishHenkilotunnusHasDVVsShapeAndControlCharacterOnADayNotAfterToday(t *testing.T) {
	eachFinnishHenkilotunnus(t, func(got string, b finnishBirth) {
		if b.born.After(testDay) || b.age > 99 {
			t.Fatalf("%q was born %s, want someone aged 0 to 99 on %s", got, b.born.Format("2006-01-02"), testDay.Format("2006-01-02"))
		}
		if b.individual < 900 {
			t.Fatalf("%q has the individual number %03d, want 900 to 999, which DVV's 002 to 899 issued range leaves out", got, b.individual)
		}
	})
}

func TestFinnishHenkilotunnusDrawsBothTheNineteenHundredsAndTheTwoThousands(t *testing.T) {
	signs := map[byte]bool{}
	eachFinnishHenkilotunnus(t, func(got string, _ finnishBirth) { signs[got[6]] = true })
	if len(signs) != 2 || !signs['-'] || !signs['A'] {
		t.Errorf("%d numbers held the signs %v, want - and A", finnishHenkilotunnusDraws, signs)
	}
}

// Statistics Finland's "Finland in Figures 2026", 11 June 2026: the population at the end of 2025 was
// 5,652,881 persons, 1,347,494 of them aged 65 or over, 23.84%.
func TestFinnishHenkilotunnusFollowsStatisticsFinlandsShareAged65OrOver(t *testing.T) {
	old := 0
	eachFinnishHenkilotunnus(t, func(_ string, b finnishBirth) {
		if b.age >= 65 {
			old++
		}
	})
	if share := 100 * float64(old) / finnishHenkilotunnusDraws; share < 22.3 || share > 25.3 {
		t.Errorf("%.1f%% of the numbers were of people aged 65 or over, want 23.84%% by Statistics Finland's population", share)
	}
}

func TestFinnishHenkilotunnusAgreesWithABoundSexAndDrawsBoth(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	sexes := map[string]bool{}
	eachDraw(500, func() string {
		got, err := f.FakeTemplate("{/fi_FI.sex as s}{s.name}|{/fi_FI.henkilotunnus.birth}{hetu(s.individ.digit)}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, func(got string) {
		sex, number, _ := strings.Cut(got, "|")
		if odd := parseHenkilotunnus(t, number).individual%2 == 1; odd != (sex == "mies") {
			t.Fatalf("%q: want an odd individual number for a mies and an even one for a nainen", got)
		}
		sexes[sex] = true
	})
	if len(sexes) != 2 {
		t.Errorf("500 numbers held the sexes %v, want both", sexes)
	}
}
