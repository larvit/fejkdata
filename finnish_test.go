package fejkdata

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
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
		t.Run(c.what, func(t *testing.T) {
			for name, want := range c.want {
				t.Run(name, func(t *testing.T) {
					if c.got[name] != want {
						t.Errorf("%s: %q is borne by %d, want DVV's %d", c.what, name, c.got[name], want)
					}
				})
			}
		})
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
	bySex := finnishCounts(t, "first-name")
	first := map[string]map[string]int{"nainen": bySex["f"], "mies": bySex["m"]}
	last := finnishCounts(t, "last-name")[""]
	titles := map[string]int{}
	eachDraw(2000, func() string {
		got, err := f.FakeTemplate("{/fi_FI.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, func(got string) {
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
	})
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
		t.Run(tmpl, func(t *testing.T) {
			got, err := f.FakeTemplate(tmpl)
			if err != nil || got != want {
				t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
			}
		})
	}
}

func finnishSexCounts(t *testing.T) map[string]int {
	t.Helper()
	counts := map[string]int{}
	for _, row := range tsvRows(t, "data/fi_FI/sex.tsv") {
		n, err := strconv.Atoi(row["count"])
		if err != nil {
			t.Fatalf("sex.tsv row %v has no count", row)
		}
		counts[row["code"]+" "+row["name"]] = n
	}
	return counts
}

func TestFinnishSexesAreWeighedByStatisticsFinlandsPopulationOf2025(t *testing.T) {
	counts := finnishSexCounts(t)
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

// Traficom's Regulation 32 V/2025 M, Annex 1: the telecommunications areas' area codes and the mobile
// codes whose row says "No" under Available ranges.
var (
	finnishFixedAreas = map[string]string{
		"02": "Turku and Pori", "03": "Häme", "05": "Kymi", "06": "Vaasa", "08": "Oulu", "09": "Uusimaa I",
		"013": "North Karelia", "014": "Central Finland", "015": "Mikkeli", "016": "Lapland", "017": "Kuopio",
		"018": "Åland", "019": "Uusimaa II",
	}
	finnishMobileAreas = []string{
		"040", "041", "042", "044", "046", "050",
		"0435", "0436", "0437", "0438", "0439", "0450", "0451", "0452", "0453", "0456", "0458",
	}
)

func finnishFixedAreaCodes() []string {
	codes := make([]string, 0, len(finnishFixedAreas))
	for code := range finnishFixedAreas {
		codes = append(codes, code)
	}
	return codes
}

func finnishAreaSet(codes ...string) map[string]bool {
	out := map[string]bool{}
	for _, c := range codes {
		out[c] = true
	}
	return out
}

func finnishPhoneAreas(t *testing.T, kind string) map[string]bool {
	t.Helper()
	got := map[string]bool{}
	for _, row := range tsvRows(t, "data/fi_FI/phone-"+kind+".tsv") {
		if got[row["area"]] {
			t.Errorf("phone-%s repeats %q", kind, row["area"])
		}
		got[row["area"]] = true
	}
	return got
}

func TestShippedFinnishPhoneTablesHoldExactlyTheCodesAnnex1Lists(t *testing.T) {
	fixed := finnishFixedAreaCodes()
	for kind, want := range map[string][]string{"fixed": fixed, "mobile": finnishMobileAreas} {
		t.Run(kind, func(t *testing.T) {
			got, wantSet := finnishPhoneAreas(t, kind), finnishAreaSet(want...)
			if len(got) != len(wantSet) {
				t.Errorf("phone-%s holds %d area codes, Annex 1 lists %d", kind, len(got), len(wantSet))
			}
			for code := range wantSet {
				t.Run(code, func(t *testing.T) {
					if !got[code] {
						t.Errorf("phone-%s lacks %q", kind, code)
					}
				})
			}
		})
	}
}

func TestShippedFinnishPhoneFixedAreaNamesItsTelecommunicationsArea(t *testing.T) {
	for _, row := range tsvRows(t, "data/fi_FI/phone-fixed.tsv") {
		t.Run(row["area"], func(t *testing.T) {
			if want := finnishFixedAreas[row["area"]]; row["name"] != want {
				t.Errorf("area %q is named %q, the plan names it %q", row["area"], row["name"], want)
			}
		})
	}
}

func TestShippedFinnishPhoneAreaNdcIsItsCodeWithoutTheTrunkZero(t *testing.T) {
	for _, kind := range []string{"fixed", "mobile"} {
		t.Run(kind, func(t *testing.T) {
			for _, row := range tsvRows(t, "data/fi_FI/phone-"+kind+".tsv") {
				t.Run(row["area"], func(t *testing.T) {
					if row["ndc"] != strings.TrimPrefix(row["area"], "0") {
						t.Errorf("phone-%s area %q has ndc %q, want the code without the trunk 0", kind, row["area"], row["ndc"])
					}
				})
			}
		})
	}
}

var finnishNationalPhone = regexp.MustCompile(`^(0\d{1,3}) (\d+(?: \d+)?)$`)

func eachShippedFinnishPhone(t *testing.T, check func(n string)) {
	t.Helper()
	f := shipped(t, WithSeed(11))
	eachDraw(30000, func() string { return fake(t, f, "fi_FI.phone") }, check)
}

func TestShippedFinnishPhoneIsAnAllocatedAreaCodeAndASubscriberNumberOfTheLengthTheRegulationGives(t *testing.T) {
	fixed, mobile := finnishAreaSet(finnishFixedAreaCodes()...), finnishAreaSet(finnishMobileAreas...)
	eachShippedFinnishPhone(t, func(n string) {
		m := finnishNationalPhone.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q is not an area code and a subscriber number in groups", n)
		}
		length := len(m[1]) - 1 + len(strings.ReplaceAll(m[2], " ", ""))
		switch {
		case fixed[m[1]] && length >= 7 && length <= 9:
		case mobile[m[1]] && length >= 8 && length <= 10:
		default:
			t.Fatalf("phone %q has area code %q and %d digits after the 0, which the plan allows for neither a fixed nor a mobile number", n, m[1], length)
		}
	})
}

// Section 11 of the regulation reserves the number ranges beginning with 2 to 8 for subscriber
// numbering in a telecommunications area, and assigns no new range beginning with 12 to 19.
func eachShippedFinnishFixedSubscriberNumber(t *testing.T, check func(n, subscriber string)) {
	t.Helper()
	eachShippedFinnishPhone(t, func(n string) {
		m := finnishNationalPhone.FindStringSubmatch(n)
		if _, fixed := finnishFixedAreas[m[1]]; fixed {
			check(n, m[2])
		}
	})
}

func TestShippedFinnishPhoneFixedSubscriberNumberBeginsWith2To8(t *testing.T) {
	eachShippedFinnishFixedSubscriberNumber(t, func(n, subscriber string) {
		if subscriber[0] < '2' || subscriber[0] > '8' {
			t.Fatalf("phone %q has a fixed subscriber number beginning with %c, not 2 to 8", n, subscriber[0])
		}
	})
}

func TestShippedFinnishPhoneDrawsAFixedSubscriberNumberBeginningWithEachDigitFrom2To8(t *testing.T) {
	first := map[byte]bool{}
	eachShippedFinnishFixedSubscriberNumber(t, func(_, subscriber string) { first[subscriber[0]] = true })
	for _, d := range []byte("2345678") {
		t.Run(string(d), func(t *testing.T) {
			if !first[d] {
				t.Errorf("30000 phones drew no fixed subscriber number beginning with %c", d)
			}
		})
	}
}

// Each kind and area code length has its own subscriber lengths, so each is counted apart.
func TestShippedFinnishPhoneDrawsEveryLengthTheRegulationGivesForEachAreaCodeLength(t *testing.T) {
	seen := map[string]map[int]bool{}
	eachShippedFinnishPhone(t, func(n string) {
		m := finnishNationalPhone.FindStringSubmatch(n)
		kind := "mobile"
		if _, fixed := finnishFixedAreas[m[1]]; fixed {
			kind = "fixed"
		}
		key := kind + " " + strconv.Itoa(len(m[1])) + "-character area code"
		if seen[key] == nil {
			seen[key] = map[int]bool{}
		}
		seen[key][len(m[1])-1+len(strings.ReplaceAll(m[2], " ", ""))] = true
	})
	for key, want := range map[string][]int{
		"fixed 2-character area code": {7, 8, 9}, "fixed 3-character area code": {7, 8, 9},
		"mobile 3-character area code": {8, 9, 10}, "mobile 4-character area code": {8, 9, 10},
	} {
		t.Run(key, func(t *testing.T) {
			for _, length := range want {
				t.Run(strconv.Itoa(length), func(t *testing.T) {
					if !seen[key][length] {
						t.Errorf("30000 phones drew no %s number of %d digits after the 0", key, length)
					}
				})
			}
		})
	}
}

func TestShippedFinnishPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	var nFixed int
	eachShippedFinnishPhone(t, func(n string) {
		if _, fixed := finnishFixedAreas[strings.Fields(n)[0]]; fixed {
			nFixed++
		}
	})
	if nFixed < 2580 || nFixed > 2880 {
		t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", nFixed)
	}
}

func TestShippedFinnishPhoneInternationalDropsTheTrunkZeroBehindTheCountryCode(t *testing.T) {
	f := shipped(t, WithSeed(5))
	eachDraw(300, func() string { return fakeTemplate(t, f, "{/fi_FI.phone as p}{p}|{p.national}|{p.international}") }, func(got string) {
		parts := strings.Split(got, "|")
		if len(parts) != 3 || parts[0] != parts[1] || parts[2] != "+358 "+strings.TrimPrefix(parts[0], "0") {
			t.Fatalf("%q: want the national form twice, then +358 and the national form without its 0", got)
		}
	})
}

func TestShippedFinnishPhoneAreaSelectsOneRowByItsCode(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for path, want := range map[string]string{
		"fi_FI.phone-fixed[09].name": "Uusimaa I", "fi_FI.phone-fixed[013].name": "North Karelia",
		"fi_FI.phone-mobile[040].ndc": "40", "fi_FI.phone-mobile[0451].ndc": "451",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fake(t, f, path); got != want {
				t.Errorf("printed %q, want %q", got, want)
			}
		})
	}
}

func eachShippedFinnishDate(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, WithSeed(3))
	eachDraw(500, func() string { return fake(t, f, "fi_FI.date") }, check)
}

func finnishDateYear(got string) int {
	d, _ := time.Parse("2.1.2006", got)
	return d.Year()
}

func TestShippedFinnishDateIsDayMonthYearWithDotsAndNoLeadingZeroFrom1970To2029(t *testing.T) {
	date := regexp.MustCompile(`^([1-9]|[12]\d|3[01])\.([1-9]|1[0-2])\.\d{4}$`)
	eachShippedFinnishDate(t, func(got string) {
		d, err := time.Parse("2.1.2006", got)
		if err != nil || !date.MatchString(got) || d.Year() < 1970 || d.Year() > 2029 {
			t.Fatalf("date %q is not d.m.yyyy without leading zeros in 1970 to 2029: %v", got, err)
		}
	})
}

func TestShippedFinnishDateSpans1970To2029(t *testing.T) {
	years := map[int]bool{}
	eachShippedFinnishDate(t, func(got string) { years[finnishDateYear(got)] = true })
	if !years[1970] || !years[2029] {
		t.Errorf("500 dates drew the years %v, want 1970 and 2029 among them", years)
	}
}

func TestShippedFinnishDatesIncludeADayOfOneDigit(t *testing.T) {
	short := false
	eachShippedFinnishDate(t, func(got string) { short = short || got[1] == '.' })
	if !short {
		t.Error("500 dates held no day of one digit, want one")
	}
}

func TestShippedFinnishEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	f := shipped(t, WithSeed(4))
	eachDraw(500, func() string { return fake(t, f, "fi_FI.email") }, func(got string) {
		if !exampleEmail.MatchString(got) {
			t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
		}
	})
}

func TestShippedFinnishEmailsAreMostlyDistinct(t *testing.T) {
	f := shipped(t, WithSeed(4))
	seen := tallyDraws(500, func() string { return fake(t, f, "fi_FI.email") })
	if len(seen) < 400 {
		t.Errorf("500 emails drew %d distinct, want 400 or more", len(seen))
	}
}

func finnishNameSlugs(t *testing.T, file string) map[string]bool {
	t.Helper()
	fold := strings.NewReplacer("ä", "a", "ö", "o", "å", "a")
	out := map[string]bool{}
	for _, row := range tsvRows(t, file) {
		out[fold.Replace(strings.ToLower(row["name"]))] = true
	}
	return out
}

func TestShippedFinnishEmailHoldsAFirstAndALastNameInHalfItsAddresses(t *testing.T) {
	firsts, lasts := finnishNameSlugs(t, "data/fi_FI/first-name.tsv"), finnishNameSlugs(t, "data/fi_FI/last-name.tsv")
	trailingDigits := regexp.MustCompile(`\d+$`)
	f := shipped(t, WithSeed(9))
	seen := tallyDraws(5000, func() string {
		local, _, _ := strings.Cut(fake(t, f, "fi_FI.email"), "@")
		return lastNameInEmail(trailingDigits.ReplaceAllString(local, ""), firsts, lasts)
	})
	named := 5000 - seen[""]
	delete(seen, "")
	assertShare(t, "addresses holding a first and a last name", named, 5000, 0.5)
	if len(seen) < 40 {
		t.Errorf("5000 emails held %d distinct last names, want 40 or more", len(seen))
	}
}

var finnishLegalFormSuffix = regexp.MustCompile(` (Oy|Ky|Ay|osk)$`)

func finnishLegalForm(company string) string {
	if m := finnishLegalFormSuffix.FindStringSubmatch(company); m != nil {
		return m[1]
	}
	return "Tmi"
}

func eachShippedFinnishCompany(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, WithSeed(6))
	eachDraw(shareDraws, func() string { return fake(t, f, "fi_FI.company") }, check)
}

func TestShippedFinnishCompanyIsANameThatMayEndInALegalForm(t *testing.T) {
	t.Parallel()
	name := regexp.MustCompile(`^\p{L}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (Oy|Ky|Ay|osk))?$`)
	eachShippedFinnishCompany(t, func(got string) {
		if !name.MatchString(got) || strings.HasSuffix(got, " Tmi") {
			t.Fatalf("company %q is not a name that may end in a legal form", got)
		}
	})
}

func TestShippedFinnishCompanyEndsInALegalFormDrawnByStatisticsFinlandsEnterprises(t *testing.T) {
	t.Parallel()
	registered := map[string]int{"Tmi": 295203, "Oy": 240524, "Ky": 13356, "Ay": 4726, "osk": 2811}
	total := registered["Tmi"] + registered["Oy"] + registered["Ky"] + registered["Ay"] + registered["osk"]
	form := map[string]int{}
	eachShippedFinnishCompany(t, func(got string) { form[finnishLegalForm(got)]++ })
	for code, n := range registered {
		t.Run(code, func(t *testing.T) {
			assertShare(t, code+" of the companies", form[code], shareDraws, float64(n)/float64(total))
		})
	}
}

func TestShippedFinnishCompanyPathsDrawOneFormEach(t *testing.T) {
	f := shipped(t, WithSeed(7))
	for path, suffix := range map[string]string{"tmi": "", "oy": " Oy", "ky": " Ky", "ay": " Ay", "osk": " osk"} {
		t.Run(path, func(t *testing.T) {
			eachDraw(100, func() string { return fake(t, f, "fi_FI.company."+path) }, func(got string) {
				if suffix == "" && finnishLegalFormSuffix.MatchString(got) || !strings.HasSuffix(got, suffix) {
					t.Fatalf("fi_FI.company.%s = %q, want a name ending in %q, or in no legal form where that is empty", path, got, suffix)
				}
			})
		})
	}
}

func TestShippedFinnishCompanyLegalIsItsAbbreviation(t *testing.T) {
	f := shipped(t, WithSeed(7))
	eachDraw(200, func() string { return fakeTemplate(t, f, "{/fi_FI.company as c}{c}|{c.legal}") }, func(got string) {
		company, legal, _ := strings.Cut(got, "|")
		want := finnishLegalForm(company)
		if want == "Tmi" {
			want = ""
		}
		if legal != want {
			t.Fatalf("company %q has legal %q", company, legal)
		}
	})
}

func finnishCompanyShape(name string, words []string, firsts, lasts map[string]bool) []string {
	var shapes []string
	for _, w := range words {
		if head, ok := strings.CutSuffix(name, " "+w); ok && lasts[head] {
			shapes = append(shapes, "last name and trade word")
		}
	}
	if first, last, ok := strings.Cut(name, " "); ok && firsts[first] && lasts[last] {
		shapes = append(shapes, "first and last name")
	}
	return shapes
}

func finnishNames(t *testing.T, file string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, row := range tsvRows(t, file) {
		out[row["name"]] = true
	}
	return out
}

func tallyFinnishSoleTraderShapes(t *testing.T, f *Generator, draws int, words []string, firsts, lasts map[string]bool) map[string]int {
	t.Helper()
	seen := map[string]int{}
	eachDraw(draws, func() string { return fake(t, f, "fi_FI.company.tmi") }, func(name string) {
		for _, shape := range finnishCompanyShape(name, words, firsts, lasts) {
			seen[shape]++
		}
	})
	return seen
}

func TestShippedFinnishCompanyNamesAreBuiltOfNamesAndTradeWords(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("data/fi_FI/company-word.json")
	if err != nil {
		t.Fatal(err)
	}
	var words []string
	if err := json.Unmarshal(src, &words); err != nil || len(words) < 20 {
		t.Fatalf("company-word.json is %d words (%v), want an array of 20 or more", len(words), err)
	}
	firsts, lasts := finnishNames(t, "data/fi_FI/first-name.tsv"), finnishNames(t, "data/fi_FI/last-name.tsv")
	const draws = 4000
	seen := tallyFinnishSoleTraderShapes(t, shipped(t, WithSeed(8)), draws, words, firsts, lasts)
	for _, shape := range []string{"last name and trade word", "first and last name"} {
		t.Run(shape, func(t *testing.T) {
			if seen[shape] < draws/10 {
				t.Errorf("%d sole traders drew a name of the shape %s %d times, want %d or more", draws, shape, seen[shape], draws/10)
			}
		})
	}
}

func TestShippedFinnishAddressIsAStreetANumberAPostinumeroAndAPostitoimipaikka(t *testing.T) {
	f := shipped(t, WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[a-z]?\n\d{5} \pL.*$`)
	eachDraw(300, func() string { return fake(t, f, "fi_FI.address") }, func(a string) {
		if !address.MatchString(a) {
			t.Fatalf("fi_FI.address = %q, want a street and a number, then a postinumero and a postitoimipaikka", a)
		}
	})
}

func TestShippedFinnishHouseNumbersWriteTheirDeliveryLetterInLowerCase(t *testing.T) {
	f := shipped(t, WithSeed(5))
	lettered := 0
	eachDraw(20000, func() string { return fake(t, f, "fi_FI.address.street-number") }, func(n string) {
		if strings.TrimRight(n, "abcdefghijklmnopqrstuvwxyzåäö") != n {
			lettered++
		}
	})
	if lettered == 0 {
		t.Error("fi_FI.address.street-number drew no house number with a lower-case delivery letter in 20000, want some, as Posti's file holds them")
	}
}

func finnishPostitoimipaikat(t *testing.T) map[string]bool {
	t.Helper()
	places := map[string]bool{}
	for _, r := range geoRows(t, "FI", "locality") {
		places[r[0]] = true
	}
	return places
}

func TestShippedFinnishPlacesAreWrittenAsPlaceNames(t *testing.T) {
	places := finnishPostitoimipaikat(t)
	for _, name := range []string{"Helsinki", "Espoo", "Tampere", "Turku", "Oulu", "Jyväskylä", "Heinola kk", "Alavus as", "Inhan tehtaat", "Kuoppa-aho", "Yli-Ii"} {
		t.Run(name, func(t *testing.T) {
			if !places[name] {
				t.Errorf("geo.FI: no postitoimipaikka %q", name)
			}
		})
	}
}

func TestShippedFinnishPlacesAreNumberedAsStatisticsFinlandWritesThem(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for path, want := range map[string]string{
		"geo.FI.municipality[Helsinki].code":   "091",
		"geo.FI.municipality[Turku].code":      "853",
		"geo.FI.municipality[Helsinki].region": "01",
		"geo.FI.region[Uusimaa].code":          "01",
		"geo.FI.region[Lappi].code":            "19",
		"geo.FI.postal-code[00100].locality":   "Helsinki",
		"geo.FI.postal-code[20720].locality":   "Turku",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

// Posti's files hold no street on Åland, so its maakunta ships no place.
func TestShippedFinnishPlacesHoldTheEighteenMainlandMaakunnat(t *testing.T) {
	if n := len(geoRows(t, "FI", "region")); n != 18 {
		t.Errorf("geo.FI holds %d regions, want mainland Finland's 18 maakunnat", n)
	}
}

func TestShippedFinnishStreetCarriesPostisSwedishName(t *testing.T) {
	if !slices.ContainsFunc(tsvRows(t, "data/geo/FI/street.tsv"), func(r map[string]string) bool {
		return r["name"] == "Uudenmaantie" && r["locality"] == "Turku" && r["swedish-name"] == "Nylandsvägen"
	}) {
		t.Error("geo.FI: no street Uudenmaantie in Turku with the swedish-name Nylandsvägen")
	}
}

func TestShippedFinnishStreetCarriesTheSwedishNamePostiGivesOnOnlySomeOfItsRanges(t *testing.T) {
	if !slices.ContainsFunc(tsvRows(t, "data/geo/FI/street.tsv"), func(r map[string]string) bool {
		return r["name"] == "Laurilantie" && r["locality"] == "Järvenpää" && r["swedish-name"] == "Laurilavägen"
	}) {
		t.Error("geo.FI: no street Laurilantie in Järvenpää with the swedish-name Laurilavägen")
	}
}

func eachFinnishStreetRow(t *testing.T, check func(i int, r map[string]string)) {
	t.Helper()
	for i, r := range tsvRows(t, "data/geo/FI/street.tsv") {
		check(i, r)
	}
}

func TestShippedFinnishStreetNamesOpenWithNoLowerCaseLetter(t *testing.T) {
	for _, column := range []string{"name", "swedish-name"} {
		t.Run(column, func(t *testing.T) {
			eachFinnishStreetRow(t, func(i int, r map[string]string) {
				if first, _ := utf8.DecodeRuneInString(r[column]); unicode.IsLower(first) {
					t.Errorf("geo.FI street row %d has the %s %q, want its first letter in upper case", i+1, column, r[column])
				}
			})
		})
	}
}

func TestShippedFinnishStreetWithoutASwedishNameCarriesItsFinnishOne(t *testing.T) {
	if i := slices.IndexFunc(tsvRows(t, "data/geo/FI/street.tsv"), func(r map[string]string) bool { return r["swedish-name"] == "" }); i >= 0 {
		t.Errorf("geo.FI street row %d has no swedish-name, want its Finnish name where Posti gives no Swedish one", i+1)
	}
}

// Posti lists highways such as `Valtatie 23`, postinumerot, and islands or lakes such as
// `Mustalampi, Lammassaari` or `Makkarasaari/Lapinsaari` beside the streets.
func TestShippedFinnishStreetNamesHoldNoDigitCommaBracketOrSlash(t *testing.T) {
	if i := slices.IndexFunc(tsvRows(t, "data/geo/FI/street.tsv"), func(r map[string]string) bool { return strings.ContainsAny(r["name"], "0123456789,()/") }); i >= 0 {
		t.Errorf("geo.FI street row %d holds a digit, comma, bracket or slash, want a street name", i+1)
	}
}
