package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
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

// The first two digits under which Nkom's E.164 register has assigned numbers to providers.
// Every other series of 8-digit numbers is blocked, free or kept for machines or special numbers.
var (
	nkomAssigned = map[string][]string{
		"mobile": strings.Fields("40 41 45 46 47 48 90 91 92 93 94 95 96 97 98 99"),
		"fixed":  strings.Fields("21 22 23 24 31 32 33 35 37 38 51 52 53 55 56 57 61 62 63 64 66 67 68 69 70 71 72 73 74 75 76 77 78 79"),
	}
	norwegianMobile = regexp.MustCompile(`^[49]\d( \d{2}){3}$`)
	norwegianFixed  = regexp.MustCompile(`^[23567]\d( \d{2}){3}$`)
)

func TestShippedNorwegianPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	f := shipped(t, WithSeed(11))
	var fixed int
	for i := 0; i < 30000; i++ {
		n := fake(t, f, "nb_NO.phone")
		switch {
		case norwegianMobile.MatchString(n):
		case norwegianFixed.MatchString(n):
			fixed++
		default:
			t.Fatalf("phone %q is neither a mobile nor a fixed number grouped in twos", n)
		}
	}
	if fixed < 2580 || fixed > 2880 {
		t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", fixed)
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

func TestShippedNorwegianPhoneTablesHoldFourDigitPrefixesOfSeriesNkomHasAssigned(t *testing.T) {
	for kind, series := range map[string]*regexp.Regexp{
		"mobile": regexp.MustCompile(`^[49]\d \d{2}$`),
		"fixed":  regexp.MustCompile(`^[23567]\d \d{2}$`),
	} {
		seen := map[string]bool{}
		for _, row := range tsvRows(t, "data/nb_NO/phone-"+kind+".tsv") {
			p := row["prefix"]
			if !series.MatchString(p) || !slices.Contains(nkomAssigned[kind], p[:2]) || seen[p] {
				t.Errorf("phone-%s row %q is not four digits in a series Nkom has assigned to providers, or repeats one", kind, p)
			}
			seen[p] = true
		}
	}
}

// Nkom assigns 56 12, 40 00 and 90 00 whole. It leaves 56 20 to 56 29 and 47 55 free, blocks 45 36,
// gives 99 90 to global title addresses, and assigns only some thousands of 56 97 and 56 98.
func TestShippedNorwegianPhonePrefixIsARowOnlyWhereNkomHasAssignedItWhole(t *testing.T) {
	f := shipped(t, WithSeed(2))
	for path, want := range map[string]string{"nb_NO.phone-fixed[56 12]": "56 12", "nb_NO.phone-mobile[40 00]": "40 00", "nb_NO.phone-mobile[90 00]": "90 00"} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	for _, path := range []string{"nb_NO.phone-fixed[56 20]", "nb_NO.phone-fixed[56 97]", "nb_NO.phone-fixed[56 98]", "nb_NO.phone-mobile[45 36]", "nb_NO.phone-mobile[47 55]", "nb_NO.phone-mobile[99 90]"} {
		if got, err := f.Fake(path); err == nil {
			t.Errorf("%s = %q, but Nkom has not assigned all of it", path, got)
		}
	}
}

func TestShippedNorwegianDateIsDayMonthYearWithDotsFrom1970To2029(t *testing.T) {
	f := shipped(t, WithSeed(3))
	first, last := 9999, 0
	for i := 0; i < 500; i++ {
		got := fake(t, f, "nb_NO.date")
		d, err := time.Parse("02.01.2006", got)
		if err != nil || d.Year() < 1970 || d.Year() > 2029 {
			t.Fatalf("date %q is not dd.mm.yyyy in 1970 to 2029: %v", got, err)
		}
		first, last = min(first, d.Year()), max(last, d.Year())
	}
	if first != 1970 || last != 2029 {
		t.Errorf("500 dates span %d to %d, want 1970 to 2029", first, last)
	}
}

func TestShippedNorwegianEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	f := shipped(t, WithSeed(4))
	email := regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*@([a-z0-9-]+\.)*example\.(com|org|net)$`)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got := fake(t, f, "nb_NO.email")
		if !email.MatchString(got) {
			t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
		}
		seen[got] = true
	}
	if len(seen) < 400 {
		t.Errorf("500 emails drew %d distinct, want 400 or more", len(seen))
	}
}

// The active units of each form in Enhetsregisteret on 2026-10-22, leaving out the bankrupt, those being
// wound up and those forcibly dissolved.
func TestShippedNorwegianCompanyEndsInALegalFormDrawnByEnhetsregisteret(t *testing.T) {
	t.Parallel()
	f := shipped(t, WithSeed(6))
	registered := map[string]int{"ENK": 465166, "AS": 423370, "DA": 9105, "SA": 6761, "ANS": 6214, "ASA": 205, "KS": 105}
	const draws = shareDraws
	var total int
	for _, units := range registered {
		total += units
	}
	name := regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'-]*[\p{L}\p{N}]$`)
	abbreviation := regexp.MustCompile(` (AS|DA|ANS|ASA|KS|SA)$`)
	form := map[string]int{}
	for i := 0; i < draws; i++ {
		got := fake(t, f, "nb_NO.company")
		if !name.MatchString(got) {
			t.Fatalf("company %q is not a name that may end in a legal form", got)
		}
		if strings.HasSuffix(got, " ENK") {
			t.Fatalf("company %q ends in ENK, which carries no abbreviation", got)
		}
		if m := abbreviation.FindStringSubmatch(got); m != nil {
			form[m[1]]++
		} else {
			form["ENK"]++
		}
	}
	for code, units := range registered {
		assertShare(t, code+" of the companies", form[code], draws, float64(units)/float64(total))
	}
}

// oneWord is, per form, the weight of the shipped shapes that are one word and the weight of all its
// shipped shapes, as company-no.py wrote them; the test holds a render to them.
func TestShippedNorwegianCompanyNamesFollowTheShapesOfEnhetsregisteret(t *testing.T) {
	t.Parallel()
	oneWord := map[string][2]int{"AS": {101525, 337759}, "ENK": {9115, 350561}, "SA": {874, 5505}, "DA": {2067, 6829}, "ANS": {1165, 3980}, "ASA": {88, 186}, "KS": {16, 83}}
	const draws = 40000
	for form, c := range oneWord {
		t.Run(form, func(t *testing.T) {
			f := shipped(t, WithSeed(8))
			var single int
			for i := 0; i < draws; i++ {
				name := strings.TrimSuffix(fake(t, f, "nb_NO.company."+strings.ToLower(form)), " "+form)
				if !strings.Contains(name, " ") {
					single++
				}
			}
			assertShare(t, form+" names of one word", single, draws, float64(c[0])/float64(c[1]))
		})
	}
}

func TestNorwegianCompanyWordsAreRepeatedInEnhetsregisteret(t *testing.T) {
	rows := tsvRows(t, "data/nb_NO/company-word.tsv")
	word := regexp.MustCompile(`^\p{Lu}\p{L}{2,}$`)
	have := map[string]bool{}
	for _, row := range rows {
		have[row["word"]] = true
		if n, err := strconv.Atoi(row["count"]); err != nil || n < 100 || !word.MatchString(row["word"]) {
			t.Errorf("row %v, want a word of three or more letters, capitalised, that 100 or more register names carry", row)
		}
	}
	for _, w := range []string{"Holding", "Invest", "Eiendom", "Consulting", "Bygg"} {
		if !have[w] {
			t.Errorf("company-word.tsv omits %q, a word that 100 or more register names carry", w)
		}
	}
}

func TestShippedNorwegianEmailDrawsItsNamesFromSSB(t *testing.T) {
	var slugs []string
	fold := strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
	ascii := regexp.MustCompile(`^[a-z]{6,}$`)
	for _, row := range tsvRows(t, "data/nb_NO/last-name.tsv") {
		if slug := fold.Replace(strings.ToLower(row["name"])); ascii.MatchString(slug) {
			slugs = append(slugs, slug)
		}
	}
	f := shipped(t, WithSeed(9))
	seen := map[string]bool{}
	for i := 0; i < 5000; i++ {
		email := fake(t, f, "nb_NO.email")
		for _, slug := range slugs {
			if strings.Contains(email, slug) {
				seen[slug] = true
			}
		}
	}
	if len(seen) < 100 {
		t.Errorf("5000 emails carry %d distinct last names of SSB's, want 100 or more", len(seen))
	}
}

func TestShippedNorwegianCompanyNamesVary(t *testing.T) {
	f := shipped(t, WithSeed(7))
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		seen[fake(t, f, "nb_NO.company")] = true
	}
	if len(seen) < 700 {
		t.Errorf("1000 companies drew %d distinct, want 700 or more", len(seen))
	}
}

func TestShippedNorwegianAddressIsAStreetANumberAPostnummerAndAPoststed(t *testing.T) {
	f := shipped(t, WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[A-Z]?\n\d{4} \pL.*$`)
	for i := 0; i < 300; i++ {
		if a := fake(t, f, "nb_NO.address"); !address.MatchString(a) {
			t.Fatalf("nb_NO.address = %q, want a street and a number, then a postnummer and a poststed", a)
		}
	}
}

// Kartverket's address register spells a poststed in capitals, MO I RANA; a place name is written Mo i Rana.
func TestShippedNorwegianPoststedIsCasedAsAPlaceName(t *testing.T) {
	names := map[string]bool{}
	capitals := regexp.MustCompile(`\p{Lu}{2}`)
	for _, r := range geoRows(t, "NO", "locality") {
		names[r[0]] = true
		if capitals.MatchString(r[0]) {
			t.Errorf("geo.NO: poststed %q holds a word in capitals", r[0])
		}
	}
	for _, name := range []string{"Oslo", "Bergen", "Mo i Rana", "Kristiansund N", "Nes på Hedmarken", "Ålesund"} {
		if !names[name] {
			t.Errorf("geo.NO: no poststed %q", name)
		}
	}
}

// A poststed that is also a name, such as Berg, counts as the name. A seeded render of the AS shapes
// company-no.py wrote held a poststed that is no name in 2.15% of AS names; the test holds a render to that share.
func TestShippedNorwegianCompanyNamesHoldAPoststed(t *testing.T) {
	named := map[string]bool{}
	for _, file := range []string{"first-name", "last-name"} {
		for _, row := range tsvRows(t, "data/nb_NO/"+file+".tsv") {
			named[strings.ToLower(row["name"])] = true
		}
	}
	places := map[string]bool{}
	for _, r := range geoRows(t, "NO", "locality") {
		if !named[strings.ToLower(r[0])] {
			places[r[0]] = true
		}
	}
	f := shipped(t, WithSeed(10))
	const draws = 20000
	holding := 0
	for i := 0; i < draws; i++ {
		if holdsPlace(strings.Fields(strings.TrimSuffix(fake(t, f, "nb_NO.company.as"), " AS")), places) {
			holding++
		}
	}
	assertShare(t, "AS names holding a poststed that is no name", holding, draws, 0.0215)
}

// holdsPlace reports whether a run of the words, joined by spaces, is one of places.
func holdsPlace(words []string, places map[string]bool) bool {
	for start := range words {
		for end := start + 1; end <= len(words); end++ {
			if places[strings.Join(words[start:end], " ")] {
				return true
			}
		}
	}
	return false
}

func TestShippedNorwegianStreetsAreStreetNames(t *testing.T) {
	refused := map[string]bool{
		"E6 Vinne": true, "Einvik E-6": true, "Europaveg-39": true, "Gate 2": true, "HÅKON DEN GODES GATE": true,
		"Midlertidig Storehaugen 1500": true, "Pundsvik-FV819": true, "Veg 1152 i Geiranger": true,
		"Veg 7205": true, "Veg nr 1306": true, "Vegkode 81682": true, "Vei07": true, "Veinr. 8113": true,
	}
	for _, r := range geoRows(t, "NO", "street") {
		if refused[r[0]] {
			t.Errorf("geo.NO: street %q in %s, want no numbered placeholder, road designation or name in capitals", r[0], r[1])
		}
	}
}

// SSB names a kommune or fylke in each of its official languages, Trondheim - Tråante; Kartverket's
// address register writes the Norwegian one.
func TestShippedNorwegianKommuneAndFylkeSelectByTheirNorwegianNames(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for path, want := range map[string]string{
		"geo.NO.municipality[Trondheim].code":  "5001",
		"geo.NO.municipality[Kautokeino].code": "5612",
		"geo.NO.municipality[Oslo].code":       "0301",
		"geo.NO.region[Troms].code":            "55",
		"geo.NO.region[Trøndelag].code":        "50",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
}

// Of Kartverket's 2.57 million street addresses, 18.3% carry a letter after the number, 22.2% a
// number of one digit and 3.5% one of four.
func TestShippedNorwegianHouseNumbersFollowKartverketsRegister(t *testing.T) {
	f := shipped(t, WithSeed(11))
	const draws = 40000
	var lettered, oneDigit, fourDigits int
	for i := 0; i < draws; i++ {
		n := fake(t, f, "nb_NO.address.street-number")
		digits := strings.TrimRight(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		if digits != n {
			lettered++
		}
		switch len(digits) {
		case 1:
			oneDigit++
		case 4:
			fourDigits++
		}
	}
	assertShare(t, "house numbers with a letter", lettered, draws, 0.183)
	assertShare(t, "house numbers of one digit", oneDigit, draws, 0.222)
	assertShare(t, "house numbers of four digits", fourDigits, draws, 0.035)
}

func TestNorwegianCoinedRootsHoldNoWordOfAPlace(t *testing.T) {
	words := map[string]bool{}
	for _, table := range []string{"locality", "municipality"} {
		for _, row := range tsvRows(t, "data/geo/NO/"+table+".tsv") {
			for _, w := range strings.Fields(strings.ToLower(row["name"])) {
				words[w] = true
			}
		}
	}
	b, err := os.ReadFile("data/nb_NO/company-coined.json")
	if err != nil {
		t.Fatal(err)
	}
	var coined struct{ A []string }
	if err := json.Unmarshal(b, &coined); err != nil {
		t.Fatal(err)
	}
	for _, root := range coined.A {
		if words[strings.ToLower(root)] {
			t.Errorf("company-coined root %q is a word of a geo/NO place name", root)
		}
	}
}
