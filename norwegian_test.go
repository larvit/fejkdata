package fejkdata

import (
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const ssbCutoff = 200

func TestNorwegianNamesAreBorneBy200OrMore(t *testing.T) {
	for _, file := range []string{"first-name", "last-name"} {
		for _, row := range tsvRows(t, "data/nb_NO/"+file+".tsv") {
			if n, err := strconv.Atoi(row["count"]); err != nil || n < ssbCutoff {
				t.Errorf("%s row %v is borne by fewer than %d persons, which SSB does not publish", file, row, ssbCutoff)
			}
		}
	}
}

func TestNorwegianFirstNamesAreRowsUnderTheSexesThatBearThem(t *testing.T) {
	bySex := map[string]map[string]bool{"f": {}, "m": {}}
	for _, row := range tsvRows(t, "data/nb_NO/first-name.tsv") {
		if bySex[row["sex"]] == nil {
			t.Fatalf("first-name.tsv row %v has a sex other than f and m", row)
		}
		bySex[row["sex"]][row["name"]] = true
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
}

func TestNorwegianLastNamesStartWithHansenJohansenOlsen(t *testing.T) {
	counts := map[string]int{}
	var names []string
	for _, row := range tsvRows(t, "data/nb_NO/last-name.tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil {
			t.Fatalf("last-name.tsv row %v has no count", row)
		}
		counts[row["name"]] = n
		names = append(names, row["name"])
	}
	sort.Slice(names, func(i, j int) bool { return counts[names[i]] > counts[names[j]] })
	if want := []string{"Hansen", "Johansen", "Olsen"}; !slices.Equal(names[:3], want) {
		t.Errorf("the three most common last names are %v, want %v", names[:3], want)
	}
}

func TestNorwegianPersonIsATitleOrNoneAFirstNameOfItsSexAndALastName(t *testing.T) {
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
		got, err := f.FakeTemplate("{/nb_NO.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		sex, prefix, name, surname, whole := cells[0], cells[1], cells[2], cells[3], cells[4]
		if !first[sex][name] {
			t.Fatalf("%q: %q is not a first name shipped for the sex %q, which is kvinne or mann", got, name, sex)
		}
		if !last[surname] {
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

func TestNorwegianBirthDaysAreEveryDayOfAYearInSkatteetatensSyntheticMonths(t *testing.T) {
	days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	seen := map[string]bool{}
	rows := tsvRows(t, "data/nb_NO/birth-day.tsv")
	for _, row := range rows {
		ddmm := row["ddmm"]
		dd, errDay := strconv.Atoi(ddmm[:2])
		mm, errMonth := strconv.Atoi(ddmm[2:])
		if len(ddmm) != 4 || errDay != nil || errMonth != nil || mm < 81 || mm > 92 || dd < 1 || dd > days[mm-81] || seen[ddmm] {
			t.Errorf("birth-day.tsv row %q is not a day of a month plus 80, once", ddmm)
		}
		seen[ddmm] = true
	}
	if len(rows) != 365 {
		t.Errorf("birth-day.tsv holds %d days, want 365, a year without 29 February", len(rows))
	}
}

// TestNorwegianFodselsnummerIsSkatteetatensSyntheticSeries reads each number as Skatteetaten
// defines it: six digits of date with 80 added to the month, three individ digits whose
// range gives the century, and two check digits.
func TestNorwegianFodselsnummerIsSkatteetatensSyntheticSeries(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	var before2000, from2000 int
	for i := 0; i < 5000; i++ {
		got := fake(t, f, "nb_NO.fodselsnummer")
		if !validFnr(got) {
			t.Fatalf("%q is not a fødselsnummer with valid check digits", got)
		}
		dd, _ := strconv.Atoi(got[0:2])
		mm, _ := strconv.Atoi(got[2:4])
		yy, _ := strconv.Atoi(got[4:6])
		individ, _ := strconv.Atoi(got[6:9])
		if mm < 81 || mm > 92 || dd < 1 || dd > days[mm-81] {
			t.Fatalf("%q has day %d of month %d, want a day of a month plus 80", got, dd, mm)
		}
		year := 0
		switch {
		case individ <= 499:
			year = 1900 + yy
		case yy < 40:
			year = 2000 + yy
		default:
			t.Fatalf("%q has individ %d with year %d, a pair fejkdata does not draw", got, individ, yy)
		}
		if year < 1940 || year > 2025 {
			t.Fatalf("%q is born in %d, want 1940 to 2025", got, year)
		}
		if year < 2000 {
			before2000++
		} else {
			from2000++
		}
	}
	if before2000 == 0 || from2000 == 0 {
		t.Errorf("5000 numbers held %d born before 2000 and %d from 2000, want both", before2000, from2000)
	}
}

func TestNorwegianFodselsnummerAgreesWithABoundSex(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	thirds := map[string]map[byte]bool{"kvinne": {}, "mann": {}}
	for i := 0; i < 500; i++ {
		got, err := f.FakeTemplate("{/nb_NO.sex as s}{s.name}|{/nb_NO.fodselsnummer.birth}{fnr(s.individ.digit)}")
		if err != nil {
			t.Fatal(err)
		}
		sex, number, _ := strings.Cut(got, "|")
		if odd := (number[8]-'0')%2 == 1; !validFnr(number) || odd != (sex == "mann") {
			t.Fatalf("%q: want a valid number, odd for a mann and even for a kvinne", got)
		}
		seen[sex] = true
		thirds[sex][number[8]] = true
	}
	if len(seen) != 2 {
		t.Errorf("500 draws gave the sexes %v, want both", seen)
	}
	for sex, digits := range thirds {
		if len(digits) != 5 {
			t.Errorf("500 draws gave a %s the third individ digits %v, want all five of its parity", sex, digits)
		}
	}
}

// Nkom's E.164 register and Nummerforskriften § 16 give 8-digit mobile numbers
// the series 40 to 49 and 90 to 99, and 8-digit fixed numbers 20 to 39, 50 to 57 and 60 to 79.
// Språkrådet writes both in groups of two since 22 May 2019. The register blocks 20, 30, 50
// and 60 whole and gives 58 and 59 to machine-to-machine traffic.
var (
	norwegianMobile = regexp.MustCompile(`^[49]\d( \d{2}){3}$`)
	norwegianFixed  = regexp.MustCompile(`^[23567]\d( \d{2}){3}$`)
	nkomUnassigned  = []string{"20", "30", "50", "58", "59", "60"}
)

func TestShippedNorwegianPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	f := shipped(t, WithSeed(11))
	var mobile, fixed int
	for i := 0; i < 3000; i++ {
		n := fake(t, f, "nb_NO.phone")
		switch {
		case norwegianMobile.MatchString(n):
			mobile++
		case norwegianFixed.MatchString(n):
			fixed++
		default:
			t.Fatalf("phone %q is neither a mobile nor a fixed number grouped in twos", n)
		}
		if slices.Contains(nkomUnassigned, n[:2]) {
			t.Fatalf("phone %q is in a series Nkom blocks or gives to machines", n)
		}
	}
	if fixed == 0 || mobile < 7*fixed || mobile > 15*fixed {
		t.Fatalf("3000 phones drew %d mobile and %d fixed, want about ten to one", mobile, fixed)
	}
}

func TestShippedNorwegianPhoneInternational(t *testing.T) {
	f := shipped(t, WithSeed(5))
	for i := 0; i < 300; i++ {
		got := fakeTemplate(t, f, "{/nb_NO.phone as p}{p}|{p.national}|{p.international}")
		parts := strings.Split(got, "|")
		if len(parts) != 3 || parts[0] != parts[1] || parts[2] != "+47 "+parts[0] {
			t.Fatalf("%q: want the national form twice, then +47 and the national form", got)
		}
	}
}

func TestShippedNorwegianPhoneTablesHoldFourDigitPrefixesOfAssignedSeries(t *testing.T) {
	for file, series := range map[string]*regexp.Regexp{
		"phone-mobile": regexp.MustCompile(`^[49]\d \d{2}$`),
		"phone-fixed":  regexp.MustCompile(`^[23567]\d \d{2}$`),
	} {
		seen := map[string]bool{}
		for _, row := range tsvRows(t, "data/nb_NO/"+file+".tsv") {
			p := row["prefix"]
			if !series.MatchString(p) || slices.Contains(nkomUnassigned, p[:2]) || seen[p] {
				t.Errorf("%s row %q is not four digits in an assigned series of its kind, once", file, p)
			}
			seen[p] = true
		}
	}
}

// Nkom assigns 56 12 whole, 40 00 and 90 00 whole, 56 20 to 56 29 to nobody, and of 56 97 and
// 56 98 only some thousands, so a prefix is a row only when all ten thousand numbers are assigned.
func TestShippedNorwegianPhonePrefixIsARowOnlyWhereNkomHasAssignedItWhole(t *testing.T) {
	f := shipped(t, WithSeed(2))
	for path, want := range map[string]string{"nb_NO.phone-fixed[56 12]": "56 12", "nb_NO.phone-mobile[40 00]": "40 00", "nb_NO.phone-mobile[90 00]": "90 00"} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	for _, prefix := range []string{"56 20", "56 97", "56 98"} {
		if got, err := f.Fake("nb_NO.phone-fixed[" + prefix + "]"); err == nil {
			t.Errorf("phone-fixed[%s] = %q, but Nkom has not assigned all of it", prefix, got)
		}
	}
}
