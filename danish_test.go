package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
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
	}{{"girls", first["f"], dstGirls}, {"boys", first["m"], dstBoys}, {"last names", last, dstLast}} {
		twentieth := 1 << 30
		for _, n := range c.top {
			twentieth = min(twentieth, n)
		}
		if len(c.got) < 200 {
			t.Errorf("%s: %d names, want 200 or more", c.what, len(c.got))
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
			t.Fatalf("%q: %q is not a first name shipped for the sex %q", got, name, sex)
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

func TestShippedDanishAddressIsAStreetANumberAPostnummerAndAPostdistrikt(t *testing.T) {
	f := shipped(t, WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[A-Z]?\n\d{4} \pL.*$`)
	for i := 0; i < 300; i++ {
		if a := fake(t, f, "da_DK.address"); !address.MatchString(a) {
			t.Fatalf("da_DK.address = %q, want a street and a number, then a postnummer and a postdistrikt", a)
		}
	}
}

// A postdistrikt is named as DAR writes it. Statistics Denmark's kommune 101 is numbered 0101, as DAR
// writes it, and its region 084 is numbered 1084, as DAGI writes it.
func TestShippedDanishPlacesAreNamedAndNumberedAsTheRegistersWriteThem(t *testing.T) {
	districts := map[string]bool{}
	for _, r := range geoRows(t, "DK", "locality") {
		districts[r[0]] = true
	}
	for _, name := range []string{"København K", "Frederiksberg C", "Aarhus C", "Odense C", "Aalborg", "Esbjerg", "Rønne"} {
		if !districts[name] {
			t.Errorf("geo.DK: no postdistrikt %q", name)
		}
	}
	f := shipped(t, WithSeed(1))
	for path, want := range map[string]string{
		"geo.DK.municipality[København].code":    "0101",
		"geo.DK.municipality[Odense].code":       "0461",
		"geo.DK.municipality[Aarhus].code":       "0751",
		"geo.DK.municipality[København].region":  "1084",
		"geo.DK.region[Region Nordjylland].code": "1081",
		"geo.DK.postal-code[8000].locality":      "Aarhus C",
		"geo.DK.postal-code[5000].locality":      "Odense C",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	if n := len(geoRows(t, "DK", "region")); n != 5 {
		t.Errorf("geo.DK holds %d regions, want Denmark's 5", n)
	}
}

// Of DAR's 2.62 million house numbers in force, 14.4% carry a letter after the number, 29.3% have
// one digit and none has four.
func TestShippedDanishHouseNumbersFollowDAR(t *testing.T) {
	f := shipped(t, WithSeed(11))
	const draws = 40000
	var lettered, oneDigit, fourDigits int
	for i := 0; i < draws; i++ {
		n := fake(t, f, "da_DK.address.street-number")
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
	assertShare(t, "house numbers with a letter", lettered, draws, 0.144)
	assertShare(t, "house numbers of one digit", oneDigit, draws, 0.293)
	if fourDigits > 0 {
		t.Errorf("da_DK.address drew %d house numbers of four digits in %d, want none, as DAR holds none", fourDigits, draws)
	}
}

func TestDanishIndividHoldsOneDigitOfEachSexsParity(t *testing.T) {
	rows := tsvRows(t, "data/da_DK/individ.tsv")
	if len(rows) != 2 {
		t.Fatalf("individ.tsv holds %d rows, want one per sex", len(rows))
	}
	for _, row := range rows {
		digit, err := strconv.Atoi(row["digit"])
		if err != nil || digit < 0 || digit > 9 || (digit%2 == 1) != (row["sex"] == "m") {
			t.Errorf("individ.tsv row %v: want one digit, odd for m and even for f", row)
		}
	}
}

// Statistics Denmark's FOLK1A, the population on 1 January 2026 by age: 6,025,603 people, of
// whom 1,135,912 were under 18, 3,618,817 were 18 to 64, and the 1,269,665 of 65 to 99 and the
// 1,209 of 100 or more make up the rest.
func TestDanishCprIsCprKontoretsNumberOfAPersonUnder100(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	const draws = 20000
	var under18, from18, from65 int
	centuries := map[int]bool{}
	sevenths := map[byte]bool{}
	for i := 0; i < draws; i++ {
		got := fake(t, f, "da_DK.cpr")
		if len(got) != 11 || got[6] != '-' || !validCpr(got) {
			t.Fatalf("%q is not ddmmyy-ssss with a valid check digit", got)
		}
		dd, _ := strconv.Atoi(got[0:2])
		mm, _ := strconv.Atoi(got[2:4])
		yy, _ := strconv.Atoi(got[4:6])
		year := cprYear(yy, int(got[7]-'0'))
		born := time.Date(year, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
		if born.Day() != dd || int(born.Month()) != mm || born.After(testDay) {
			t.Fatalf("%q is born on %s, want a day of the calendar not after %s", got, born.Format("2006-01-02"), testDay.Format("2006-01-02"))
		}
		age := testDay.Year() - year
		if born.AddDate(age, 0, 0).After(testDay) {
			age--
		}
		switch {
		case age < 0 || age > 99:
			t.Fatalf("%q is born in %d, aged %d, want 0 to 99", got, year, age)
		case age < 18:
			under18++
		case age < 65:
			from18++
		default:
			from65++
		}
		centuries[year/100] = true
		sevenths[got[7]] = true
	}
	if !centuries[19] || !centuries[20] {
		t.Errorf("%d numbers were born in the centuries %v, want the 1900s and the 2000s", draws, centuries)
	}
	if len(sevenths) < 6 {
		t.Errorf("%d numbers held the seventh digits %v, want them to vary", draws, sevenths)
	}
	for band, c := range map[string]struct{ got, want float64 }{
		"under 18": {float64(under18), 18.85},
		"18 to 64": {float64(from18), 60.06},
		"65 to 99": {float64(from65), 21.09},
	} {
		if share := 100 * c.got / draws; share < c.want-1.5 || share > c.want+1.5 {
			t.Errorf("%s were %.1f%% of the numbers, want %.2f%% by Statistics Denmark's population", band, share, c.want)
		}
	}
}

func TestDanishCprAgreesWithABoundSex(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got, err := f.FakeTemplate("{/da_DK.sex as s}{s.name}|{/da_DK.cpr.birth}-{cpr(s.individ.digit)}")
		if err != nil {
			t.Fatal(err)
		}
		sex, number, _ := strings.Cut(got, "|")
		if odd := (number[10]-'0')%2 == 1; !validCpr(number) || odd != (sex == "mand") {
			t.Fatalf("%q: want a valid number, odd for a mand and even for a kvinde", got)
		}
		seen[sex] = true
	}
	if len(seen) != 2 {
		t.Errorf("500 draws gave the sexes %v, want both", seen)
	}
}

// Annex 1 of Executive Order No. 1883 of 7 December 2020 on the Overall Danish Numbering Plan: the
// first two digits of an 8-digit number say mainly mobile or mainly fixed, and the prefixes of its
// note (!) are mobile inside fixed ranges. A prefix with third digit 0 is kept for expansion.
const (
	planMobile     = "20-29 30-31 40-42 50-53 60-61 71 81 91-93"
	planFixed      = "32-36 38-39 43-49 54-59 62-66 69 70 72-79 82 86-89 96-99"
	planMobileNote = "342 344-349 356-357 359 362 365-366 389 398 431 441 462 466 468 472 474 476 478 485-486 488-489 " +
		"493-496 498-499 542-543 545 551-552 556 571-574 577 579 584 586-587 589 597-598 627 629 641 649 658 " +
		"662-665 667 692-694 697 771-772 782-783 785-786 788-789 826-827 829"
)

func planSet(t *testing.T, spec string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, tok := range strings.Fields(spec) {
		lo, hi, ranged := strings.Cut(tok, "-")
		if !ranged {
			hi = lo
		}
		from, err1 := strconv.Atoi(lo)
		to, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil {
			t.Fatalf("plan token %q", tok)
		}
		for n := from; n <= to; n++ {
			out[strconv.Itoa(n)] = true
		}
	}
	return out
}

// danishSeries returns the three-digit series of one kind as the shipped tables write them, "34 2".
func danishSeries(t *testing.T, mobile bool) map[string]bool {
	t.Helper()
	note := planSet(t, planMobileNote)
	twos := planSet(t, planFixed)
	if mobile {
		twos = planSet(t, planMobile)
	}
	out := map[string]bool{}
	for two := range twos {
		for c := 1; c <= 9; c++ {
			series := two + strconv.Itoa(c)
			if mobile || !note[series] {
				out[two+" "+strconv.Itoa(c)] = true
			}
		}
	}
	if !mobile {
		return out
	}
	for series := range note {
		out[series[:2]+" "+series[2:]] = true
	}
	return out
}

func TestShippedDanishPhoneTablesHoldTheSeriesOfTheNumberingPlan(t *testing.T) {
	for kind, mobile := range map[string]bool{"mobile": true, "fixed": false} {
		want := danishSeries(t, mobile)
		got := map[string]bool{}
		for _, row := range tsvRows(t, "data/da_DK/phone-"+kind+".tsv") {
			if got[row["prefix"]] {
				t.Errorf("phone-%s repeats %q", kind, row["prefix"])
			}
			got[row["prefix"]] = true
		}
		for p := range want {
			if !got[p] {
				t.Errorf("phone-%s lacks %q, a series the plan allocates to %s", kind, p, kind)
			}
		}
		for p := range got {
			if !want[p] {
				t.Errorf("phone-%s holds %q, which the plan does not allocate to %s", kind, p, kind)
			}
		}
	}
}

func TestShippedDanishPhoneSeriesIsARowOnlyOfItsKind(t *testing.T) {
	f := shipped(t, WithSeed(2))
	for path, want := range map[string]string{"da_DK.phone-mobile[34 2]": "34 2", "da_DK.phone-mobile[21 1]": "21 1", "da_DK.phone-fixed[34 3]": "34 3", "da_DK.phone-fixed[75 1]": "75 1"} {
		if got := fake(t, f, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	for _, path := range []string{"da_DK.phone-fixed[34 2]", "da_DK.phone-mobile[34 3]", "da_DK.phone-mobile[20 0]", "da_DK.phone-fixed[37 1]", "da_DK.phone-fixed[67 1]", "da_DK.phone-mobile[90 1]", "da_DK.phone-fixed[80 1]", "da_DK.phone-fixed[30 0]"} {
		if got, err := f.Fake(path); err == nil {
			t.Errorf("%s = %q, but the plan allocates no such series to that kind", path, got)
		}
	}
}

func TestShippedDanishPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	mobile, fixed := danishSeries(t, true), danishSeries(t, false)
	national := regexp.MustCompile(`^(\d\d) (\d)\d( \d\d){2}$`)
	f := shipped(t, WithSeed(11))
	var nFixed int
	for i := 0; i < 30000; i++ {
		n := fake(t, f, "da_DK.phone")
		m := national.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q is not 8 digits grouped in twos", n)
		}
		switch series := m[1] + " " + m[2]; {
		case mobile[series]:
		case fixed[series]:
			nFixed++
		default:
			t.Fatalf("phone %q lies in series %q, which the plan allocates to neither mobile nor fixed numbers", n, series)
		}
	}
	if nFixed < 2580 || nFixed > 2880 {
		t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", nFixed)
	}
}

func TestShippedDanishPhoneInternational(t *testing.T) {
	f := shipped(t, WithSeed(5))
	for i := 0; i < 300; i++ {
		got := fakeTemplate(t, f, "{/da_DK.phone as p}{p}|{p.national}|{p.international}")
		parts := strings.Split(got, "|")
		if len(parts) != 3 || parts[0] != parts[1] || parts[2] != "+45 "+parts[0] {
			t.Fatalf("%q: want the national form twice, then +45 and the national form", got)
		}
	}
}

// Dansk Sprognævn's answer SV00000046 gives 2.10.2003 and advises against 02.10.2003.
func TestShippedDanishDateIsDayMonthYearWithDotsAndNoLeadingZeroFrom1970To2029(t *testing.T) {
	f := shipped(t, WithSeed(3))
	date := regexp.MustCompile(`^([1-9]|[12]\d|3[01])\.([1-9]|1[0-2])\.\d{4}$`)
	first, last, short := 9999, 0, false
	for i := 0; i < 500; i++ {
		got := fake(t, f, "da_DK.date")
		d, err := time.Parse("2.1.2006", got)
		if err != nil || !date.MatchString(got) || d.Year() < 1970 || d.Year() > 2029 {
			t.Fatalf("date %q is not d.m.yyyy without leading zeros in 1970 to 2029: %v", got, err)
		}
		first, last, short = min(first, d.Year()), max(last, d.Year()), short || d.Day() < 10
	}
	if first != 1970 || last != 2029 || !short {
		t.Errorf("500 dates span %d to %d, with a one-digit day: %v; want 1970 to 2029 and true", first, last, short)
	}
}

func TestShippedDanishEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	f := shipped(t, WithSeed(4))
	email := regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*@([a-z0-9-]+\.)*example\.(com|org|net)$`)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		got := fake(t, f, "da_DK.email")
		if !email.MatchString(got) {
			t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
		}
		seen[got] = true
	}
	if len(seen) < 400 {
		t.Errorf("500 emails drew %d distinct, want 400 or more", len(seen))
	}
}

func TestShippedDanishEmailHoldsAFirstAndALastNameInHalfItsAddresses(t *testing.T) {
	fold := strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
	slugs := func(file string) map[string]bool {
		out := map[string]bool{}
		for _, row := range tsvRows(t, file) {
			out[fold.Replace(strings.ToLower(row["name"]))] = true
		}
		return out
	}
	firsts, lasts := slugs("data/da_DK/first-name.tsv"), slugs("data/da_DK/last-name.tsv")
	trailingDigits := regexp.MustCompile(`\d+$`)
	f := shipped(t, WithSeed(9))
	var named int
	seen := map[string]bool{}
	for i := 0; i < 5000; i++ {
		local, _, _ := strings.Cut(fake(t, f, "da_DK.email"), "@")
		local = trailingDigits.ReplaceAllString(local, "")
		for cut := 1; cut < len(local); cut++ {
			right := strings.TrimLeft(local[cut:], "._-")
			if firsts[strings.TrimRight(local[:cut], "._-")] && lasts[right] {
				named++
				seen[right] = true
				break
			}
		}
	}
	if named < 2000 || len(seen) < 40 {
		t.Errorf("5000 emails held a first and a last name in %d, of %d distinct last names; want 2000 or more of 40 or more", named, len(seen))
	}
}

// Statistics Denmark's table GF15, 2024: 132,403 sole proprietorships, 218,208 private limited companies,
// 30,750 public limited companies and 11,679 partnerships and limited partnerships, split evenly.
func TestShippedDanishCompanyEndsInALegalFormDrawnByStatisticsDenmarksEnterprises(t *testing.T) {
	t.Parallel()
	f := shipped(t, WithSeed(6))
	registered := map[string]int{"ENK": 132403, "ApS": 218208, "A/S": 30750, "I/S": 5840, "K/S": 5839}
	var total int
	for _, n := range registered {
		total += n
	}
	name := regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (ApS|A/S|I/S|K/S))?$`)
	abbreviation := regexp.MustCompile(` (ApS|A/S|I/S|K/S)$`)
	form := map[string]int{}
	for i := 0; i < shareDraws; i++ {
		got := fake(t, f, "da_DK.company")
		if !name.MatchString(got) || strings.HasSuffix(got, " ENK") {
			t.Fatalf("company %q is not a name that may end in a legal form", got)
		}
		if m := abbreviation.FindStringSubmatch(got); m != nil {
			form[m[1]]++
		} else {
			form["ENK"]++
		}
	}
	for code, n := range registered {
		assertShare(t, code+" of the companies", form[code], shareDraws, float64(n)/float64(total))
	}
}

func TestShippedDanishCompanyPathsDrawOneFormEachAndLegalIsItsAbbreviation(t *testing.T) {
	f := shipped(t, WithSeed(7))
	for path, suffix := range map[string]string{"enk": "", "aps": " ApS", "as": " A/S", "is": " I/S", "ks": " K/S"} {
		for i := 0; i < 100; i++ {
			got := fake(t, f, "da_DK.company."+path)
			if suffix == "" && regexp.MustCompile(` (ApS|A/S|I/S|K/S)$`).MatchString(got) || !strings.HasSuffix(got, suffix) {
				t.Fatalf("da_DK.company.%s = %q, want a name ending in %q", path, got, suffix)
			}
		}
	}
	for i := 0; i < 200; i++ {
		company, legal, _ := strings.Cut(fakeTemplate(t, f, "{/da_DK.company as c}{c}|{c.legal}"), "|")
		if legal != "" && !strings.HasSuffix(company, " "+legal) || legal == "" && strings.Contains(company, "/") {
			t.Fatalf("company %q has legal %q", company, legal)
		}
	}
}

func TestShippedDanishCompanyNamesAreBuiltOfNamesPlacesAndTradeWords(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("data/da_DK/company-word.json")
	if err != nil {
		t.Fatal(err)
	}
	var words []string
	if err := json.Unmarshal(src, &words); err != nil || len(words) < 20 {
		t.Fatalf("company-word.json is %d words (%v), want an array of 20 or more", len(words), err)
	}
	set := func(file string) map[string]bool {
		out := map[string]bool{}
		for _, row := range tsvRows(t, file) {
			out[row["name"]] = true
		}
		return out
	}
	places, firsts, lasts := set("data/geo/DK/locality.tsv"), set("data/da_DK/first-name.tsv"), set("data/da_DK/last-name.tsv")
	f := shipped(t, WithSeed(8))
	seen := map[string]int{}
	const draws = 4000
	for i := 0; i < draws; i++ {
		name := fake(t, f, "da_DK.company.enk")
		for _, w := range words {
			if head, ok := strings.CutSuffix(name, " "+w); ok {
				switch {
				case places[head]:
					seen["place and trade word"]++
				case lasts[head]:
					seen["last name and trade word"]++
				}
			}
		}
		if first, last, ok := strings.Cut(name, " "); ok && firsts[first] && lasts[last] {
			seen["first and last name"]++
		}
	}
	for _, shape := range []string{"place and trade word", "last name and trade word", "first and last name"} {
		if seen[shape] < draws/10 {
			t.Errorf("%d sole proprietorships drew a name of the shape %s %d times, want %d or more", draws, shape, seen[shape], draws/10)
		}
	}
}
