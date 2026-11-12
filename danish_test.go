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

type danishNameSet struct {
	what string
	got  map[string]int
	top  map[string]int
}

func danishNameSets(t *testing.T) []danishNameSet {
	t.Helper()
	first, last := danishNames(t, "first-name"), danishNames(t, "last-name")[""]
	return []danishNameSet{{"girls", first["f"], dstGirls}, {"boys", first["m"], dstBoys}, {"last names", last, dstLast}}
}

func TestDanishTopNamesAreStatisticsDenmarksCounts(t *testing.T) {
	for _, c := range danishNameSets(t) {
		t.Run(c.what, func(t *testing.T) {
			for name, want := range c.top {
				t.Run(name, func(t *testing.T) {
					if c.got[name] != want {
						t.Errorf("%s: %q is borne by %d, want Statistics Denmark's %d", c.what, name, c.got[name], want)
					}
				})
			}
		})
	}
}

func TestDanishNameTablesHoldTwoHundredNamesOrMore(t *testing.T) {
	for _, c := range danishNameSets(t) {
		t.Run(c.what, func(t *testing.T) {
			if len(c.got) < 200 {
				t.Errorf("%s: %d names, want 200 or more", c.what, len(c.got))
			}
		})
	}
}

func noOtherNameIsBorneByAsManyAsTheTwentieth(t *testing.T, c danishNameSet) {
	t.Helper()
	twentieth := 1 << 30
	for _, n := range c.top {
		twentieth = min(twentieth, n)
	}
	for name, n := range c.got {
		if _, top := c.top[name]; !top && n >= twentieth {
			t.Errorf("%s: %q is borne by %d, but Statistics Denmark's twentieth is borne by %d, so %q would be in its top 20", c.what, name, n, twentieth, name)
		}
	}
}

func TestDanishOtherNamesAreRarerThanTheTwentieth(t *testing.T) {
	for _, c := range danishNameSets(t) {
		t.Run(c.what, func(t *testing.T) { noOtherNameIsBorneByAsManyAsTheTwentieth(t, c) })
	}
}

func eachRowIsALetteredName(t *testing.T, file string) {
	t.Helper()
	name := regexp.MustCompile(`^[\pL]{2,}([ '-][\pL]{2,})*$`)
	for _, row := range tsvRows(t, "data/da_DK/"+file+".tsv") {
		if !name.MatchString(row["name"]) {
			t.Errorf("%s row %v is not a name in letters", file, row)
		}
	}
}

func TestDanishNamesAreLetters(t *testing.T) {
	for _, file := range []string{"first-name", "last-name"} {
		t.Run(file, func(t *testing.T) { eachRowIsALetteredName(t, file) })
	}
}

func TestDanishFirstNamesHoldTheSexesFAndM(t *testing.T) {
	first := danishNames(t, "first-name")
	if len(first) != 2 || first["f"] == nil || first["m"] == nil {
		t.Errorf("first-name.tsv holds sexes %v, want f and m", first)
	}
}

func TestDanishAGirlsNameIsNotABoysByAccident(t *testing.T) {
	first := danishNames(t, "first-name")
	for name, sex := range map[string]string{"Mette": "f", "Freja": "f", "Søren": "m", "Mads": "m"} {
		t.Run(name, func(t *testing.T) {
			other := map[string]string{"f": "m", "m": "f"}[sex]
			if first[sex][name] == 0 || first[other][name] != 0 {
				t.Errorf("%s is not a name of sex %q alone", name, sex)
			}
		})
	}
}

func TestDanishKimIsBorneByGirlsAndBoys(t *testing.T) {
	first := danishNames(t, "first-name")
	if first["f"]["Kim"] == 0 || first["m"]["Kim"] == 0 {
		t.Error("Kim is borne by girls and boys, want a row under each sex")
	}
}

type danishPerson struct{ raw, sex, prefix, first, last, whole string }

func eachDanishPerson(t *testing.T, check func(p danishPerson)) {
	t.Helper()
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	eachDraw(2000, func() string {
		got, err := f.FakeTemplate("{/da_DK.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 5 {
			t.Fatalf("%q does not hold five cells", got)
		}
		check(danishPerson{got, cells[0], cells[1], cells[2], cells[3], cells[4]})
	})
}

func TestDanishPersonHasAFirstNameOfItsSex(t *testing.T) {
	firsts := danishNames(t, "first-name")
	first := map[string]map[string]int{"kvinde": firsts["f"], "mand": firsts["m"]}
	eachDanishPerson(t, func(p danishPerson) {
		if first[p.sex][p.first] == 0 {
			t.Fatalf("%q: %q is not a first name shipped for the sex %q", p.raw, p.first, p.sex)
		}
	})
}

func TestDanishPersonHasAShippedLastName(t *testing.T) {
	last := danishNames(t, "last-name")[""]
	eachDanishPerson(t, func(p danishPerson) {
		if last[p.last] == 0 {
			t.Fatalf("%q: %q is not a shipped last name", p.raw, p.last)
		}
	})
}

func TestDanishPersonIsItsPrefixFirstNameAndLastName(t *testing.T) {
	eachDanishPerson(t, func(p danishPerson) {
		if p.whole != p.prefix+p.first+" "+p.last {
			t.Fatalf("%q: the person is not its prefix, first name and last name", p.raw)
		}
	})
}

func TestDanishPersonMostlyBearsNoTitle(t *testing.T) {
	titles := map[string]int{}
	eachDanishPerson(t, func(p danishPerson) { titles[p.prefix]++ })
	if titles["dr. "] == 0 || titles["professor "] == 0 || titles[""] < 1800 || len(titles) != 3 {
		t.Errorf("prefixes over 2000 persons = %q, want a few %q and %q beside many with none", titles, "dr. ", "professor ")
	}
}

func TestDanishNamesFollowTheirCounts(t *testing.T) {
	f, err := New(withShipped(), WithSeed(5), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	seen := tallyDraws(20000, func() string { return fakeTemplate(t, f, "{/da_DK.last-name}") })
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
		t.Run(tmpl, func(t *testing.T) {
			got, err := f.FakeTemplate(tmpl)
			if err != nil || got != want {
				t.Errorf("FakeTemplate(%q) = %q, %v, want %q", tmpl, got, err, want)
			}
		})
	}
}

func TestShippedDanishAddressIsAStreetANumberAPostnummerAndAPostdistrikt(t *testing.T) {
	f := shipped(t, WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[A-Z]?\n\d{4} \pL.*$`)
	eachDraw(300, func() string { return fake(t, f, "da_DK.address") }, func(a string) {
		if !address.MatchString(a) {
			t.Fatalf("da_DK.address = %q, want a street and a number, then a postnummer and a postdistrikt", a)
		}
	})
}

func TestShippedDanishPlacesAreNamedAsDARWritesThem(t *testing.T) {
	districts := map[string]bool{}
	for _, r := range geoRows(t, "DK", "locality") {
		districts[r[0]] = true
	}
	for _, name := range []string{"København K", "Frederiksberg C", "Aarhus C", "Odense C", "Aalborg", "Esbjerg", "Rønne"} {
		t.Run(name, func(t *testing.T) {
			if !districts[name] {
				t.Errorf("geo.DK: no postdistrikt %q", name)
			}
		})
	}
}

// Statistics Denmark's kommune 101 is numbered 0101, as DAR writes it, and its region 084 is
// numbered 1084, as DAGI writes it.
func TestShippedDanishPlacesAreNumberedAsTheRegistersWriteThem(t *testing.T) {
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
		t.Run(path, func(t *testing.T) {
			if got := fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

func TestShippedDanishPlacesHoldDenmarksFiveRegions(t *testing.T) {
	if n := len(geoRows(t, "DK", "region")); n != 5 {
		t.Errorf("geo.DK holds %d regions, want Denmark's 5", n)
	}
}

type danishHouseNumberShares struct{ lettered, oneDigit, fourDigits int }

const danishHouseNumberDraws = 40000

func danishHouseNumbers(t *testing.T) danishHouseNumberShares {
	t.Helper()
	f := shipped(t, WithSeed(11))
	var shares danishHouseNumberShares
	eachDraw(danishHouseNumberDraws, func() string { return fake(t, f, "da_DK.address.street-number") }, func(n string) {
		digits := strings.TrimRight(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		if digits != n {
			shares.lettered++
		}
		switch len(digits) {
		case 1:
			shares.oneDigit++
		case 4:
			shares.fourDigits++
		}
	})
	return shares
}

// Of DAR's 2.62 million house numbers in force, 14.4% carry a letter after the number, 29.3% have
// one digit and none has four.
func TestShippedDanishHouseNumbersCarryALetterAsOftenAsDAR(t *testing.T) {
	assertShare(t, "house numbers with a letter", danishHouseNumbers(t).lettered, danishHouseNumberDraws, 0.144)
}

func TestShippedDanishHouseNumbersHaveOneDigitAsOftenAsDAR(t *testing.T) {
	assertShare(t, "house numbers of one digit", danishHouseNumbers(t).oneDigit, danishHouseNumberDraws, 0.293)
}

func TestShippedDanishHouseNumbersNeverHaveFourDigits(t *testing.T) {
	if n := danishHouseNumbers(t).fourDigits; n > 0 {
		t.Errorf("da_DK.address drew %d house numbers of four digits in %d, want none, as DAR holds none", n, danishHouseNumberDraws)
	}
}

func TestDanishIndividHoldsOneRowPerSex(t *testing.T) {
	if rows := tsvRows(t, "data/da_DK/individ.tsv"); len(rows) != 2 {
		t.Errorf("individ.tsv holds %d rows, want one per sex", len(rows))
	}
}

func TestDanishIndividHoldsOneDigitOfEachSexsParity(t *testing.T) {
	for _, row := range tsvRows(t, "data/da_DK/individ.tsv") {
		t.Run(row["sex"], func(t *testing.T) {
			digit, err := strconv.Atoi(row["digit"])
			if err != nil || digit < 0 || digit > 9 || (digit%2 == 1) != (row["sex"] == "m") {
				t.Errorf("individ.tsv row %v: want one digit, odd for m and even for f", row)
			}
		})
	}
}

type danishCprBirth struct {
	born      time.Time
	day, mon  int
	year, age int
}

func birthOfDanishCpr(got string) danishCprBirth {
	dd, _ := strconv.Atoi(got[0:2])
	mm, _ := strconv.Atoi(got[2:4])
	yy, _ := strconv.Atoi(got[4:6])
	year := cprYear(yy, int(got[7]-'0'))
	born := time.Date(year, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
	age := testDay.Year() - year
	if born.AddDate(age, 0, 0).After(testDay) {
		age--
	}
	return danishCprBirth{born, dd, mm, year, age}
}

const danishCprDraws = 20000

func eachDanishCpr(t *testing.T, check func(got string)) {
	t.Helper()
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	eachDraw(danishCprDraws, func() string { return fake(t, f, "da_DK.cpr") }, check)
}

func TestDanishCprIsDdmmyyAndFourDigitsWithAValidCheckDigit(t *testing.T) {
	eachDanishCpr(t, func(got string) {
		if len(got) != 11 || got[6] != '-' || !validCpr(got) {
			t.Fatalf("%q is not ddmmyy-ssss with a valid check digit", got)
		}
	})
}

func TestDanishCprIsBornOnADayOfTheCalendarNotAfterToday(t *testing.T) {
	eachDanishCpr(t, func(got string) {
		b := birthOfDanishCpr(got)
		if b.born.Day() != b.day || int(b.born.Month()) != b.mon || b.born.After(testDay) {
			t.Fatalf("%q is born on %s, want a day of the calendar not after %s", got, b.born.Format("2006-01-02"), testDay.Format("2006-01-02"))
		}
	})
}

func TestDanishCprIsOfAPersonAged0To99(t *testing.T) {
	eachDanishCpr(t, func(got string) {
		if b := birthOfDanishCpr(got); b.age < 0 || b.age > 99 {
			t.Fatalf("%q is born in %d, aged %d, want 0 to 99", got, b.year, b.age)
		}
	})
}

func TestDanishCprDrawsBirthsInTheNineteenHundredsAndTheTwoThousands(t *testing.T) {
	centuries := map[int]bool{}
	eachDanishCpr(t, func(got string) { centuries[birthOfDanishCpr(got).year/100] = true })
	if !centuries[19] || !centuries[20] {
		t.Errorf("%d numbers were born in the centuries %v, want the 1900s and the 2000s", danishCprDraws, centuries)
	}
}

func TestDanishCprVariesItsSeventhDigit(t *testing.T) {
	sevenths := map[byte]bool{}
	eachDanishCpr(t, func(got string) { sevenths[got[7]] = true })
	if len(sevenths) < 6 {
		t.Errorf("%d numbers held the seventh digits %v, want them to vary", danishCprDraws, sevenths)
	}
}

func danishAgeBand(got string) string {
	switch age := birthOfDanishCpr(got).age; {
	case age < 18:
		return "under 18"
	case age < 65:
		return "18 to 64"
	default:
		return "65 to 99"
	}
}

// Statistics Denmark's FOLK1A, the population on 1 January 2026 by age: 6,025,603 people, of
// whom 1,135,912 were under 18, 3,618,817 were 18 to 64, and the 1,269,665 of 65 to 99 and the
// 1,209 of 100 or more make up the rest.
func TestDanishCprFollowsTheAgeBandsOfStatisticsDenmarksPopulation(t *testing.T) {
	bands := map[string]int{}
	eachDanishCpr(t, func(got string) { bands[danishAgeBand(got)]++ })
	for band, want := range map[string]float64{"under 18": 18.85, "18 to 64": 60.06, "65 to 99": 21.09} {
		t.Run(band, func(t *testing.T) {
			if share := 100 * float64(bands[band]) / danishCprDraws; share < want-1.5 || share > want+1.5 {
				t.Errorf("%s were %.1f%% of the numbers, want %.2f%% by Statistics Denmark's population", band, share, want)
			}
		})
	}
}

func eachDanishCprWithASex(t *testing.T, check func(got string)) {
	t.Helper()
	f, err := New(withShipped(), WithSeed(1), onTestDay)
	if err != nil {
		t.Fatal(err)
	}
	eachDraw(500, func() string {
		got, err := f.FakeTemplate("{/da_DK.sex as s}{s.name}|{/da_DK.cpr.birth}-{cpr(s.individ.digit)}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, check)
}

func TestDanishCprAgreesWithABoundSex(t *testing.T) {
	eachDanishCprWithASex(t, func(got string) {
		sex, number, _ := strings.Cut(got, "|")
		if odd := (number[10]-'0')%2 == 1; !validCpr(number) || odd != (sex == "mand") {
			t.Fatalf("%q: want a valid number, odd for a mand and even for a kvinde", got)
		}
	})
}

func TestDanishCprDrawsBothSexesWhenItsSexIsBound(t *testing.T) {
	seen := map[string]bool{}
	eachDanishCprWithASex(t, func(got string) {
		sex, _, _ := strings.Cut(got, "|")
		seen[sex] = true
	})
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

func phoneSeriesOf(t *testing.T, kind string) (got map[string]bool, repeated []string) {
	t.Helper()
	got = map[string]bool{}
	for _, row := range tsvRows(t, "data/da_DK/phone-"+kind+".tsv") {
		if got[row["prefix"]] {
			repeated = append(repeated, row["prefix"])
		}
		got[row["prefix"]] = true
	}
	return got, repeated
}

func absentFrom(want, got map[string]bool) []string {
	var absent []string
	for p := range want {
		if !got[p] {
			absent = append(absent, p)
		}
	}
	slices.Sort(absent)
	return absent
}

var danishPhoneKinds = map[string]bool{"mobile": true, "fixed": false}

func TestShippedDanishPhoneTablesRepeatNoSeries(t *testing.T) {
	for kind := range danishPhoneKinds {
		t.Run(kind, func(t *testing.T) {
			if _, repeated := phoneSeriesOf(t, kind); len(repeated) > 0 {
				t.Errorf("phone-%s repeats %q", kind, repeated)
			}
		})
	}
}

func TestShippedDanishPhoneTablesHoldEverySeriesTheNumberingPlanAllocates(t *testing.T) {
	for kind, mobile := range danishPhoneKinds {
		t.Run(kind, func(t *testing.T) {
			got, _ := phoneSeriesOf(t, kind)
			if lacking := absentFrom(danishSeries(t, mobile), got); len(lacking) > 0 {
				t.Errorf("phone-%s lacks %q, series the plan allocates to %s", kind, lacking, kind)
			}
		})
	}
}

func TestShippedDanishPhoneTablesHoldNoSeriesTheNumberingPlanDoesNotAllocate(t *testing.T) {
	for kind, mobile := range danishPhoneKinds {
		t.Run(kind, func(t *testing.T) {
			got, _ := phoneSeriesOf(t, kind)
			if extra := absentFrom(got, danishSeries(t, mobile)); len(extra) > 0 {
				t.Errorf("phone-%s holds %q, which the plan does not allocate to %s", kind, extra, kind)
			}
		})
	}
}

func TestShippedDanishPhoneSeriesIsARowOfItsKind(t *testing.T) {
	f := shipped(t, WithSeed(2))
	for path, want := range map[string]string{"da_DK.phone-mobile[34 2]": "34 2", "da_DK.phone-mobile[21 1]": "21 1", "da_DK.phone-fixed[34 3]": "34 3", "da_DK.phone-fixed[75 1]": "75 1"} {
		t.Run(path, func(t *testing.T) {
			if got := fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

func TestShippedDanishPhoneSeriesIsNoRowOfAnotherKind(t *testing.T) {
	f := shipped(t, WithSeed(2))
	for _, path := range []string{"da_DK.phone-fixed[34 2]", "da_DK.phone-mobile[34 3]", "da_DK.phone-mobile[20 0]", "da_DK.phone-fixed[37 1]", "da_DK.phone-fixed[67 1]", "da_DK.phone-mobile[90 1]", "da_DK.phone-fixed[80 1]", "da_DK.phone-fixed[30 0]"} {
		t.Run(path, func(t *testing.T) {
			if got, err := f.Fake(path); err == nil {
				t.Errorf("%s = %q, but the plan allocates no such series to that kind", path, got)
			}
		})
	}
}

var danishNationalPhone = regexp.MustCompile(`^(\d\d) (\d)\d( \d\d){2}$`)

func eachShippedDanishPhone(t *testing.T, check func(n string)) {
	t.Helper()
	f := shipped(t, WithSeed(11))
	eachDraw(30000, func() string { return fake(t, f, "da_DK.phone") }, check)
}

func danishPhoneSeries(n string) string {
	m := danishNationalPhone.FindStringSubmatch(n)
	if m == nil {
		return ""
	}
	return m[1] + " " + m[2]
}

func TestShippedDanishPhoneIsEightDigitsGroupedInTwos(t *testing.T) {
	eachShippedDanishPhone(t, func(n string) {
		if !danishNationalPhone.MatchString(n) {
			t.Fatalf("phone %q is not 8 digits grouped in twos", n)
		}
	})
}

func TestShippedDanishPhoneLiesInASeriesTheNumberingPlanAllocates(t *testing.T) {
	mobile, fixed := danishSeries(t, true), danishSeries(t, false)
	eachShippedDanishPhone(t, func(n string) {
		if series := danishPhoneSeries(n); !mobile[series] && !fixed[series] {
			t.Fatalf("phone %q lies in series %q, which the plan allocates to neither mobile nor fixed numbers", n, series)
		}
	})
}

func TestShippedDanishPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	fixed := danishSeries(t, false)
	var nFixed int
	eachShippedDanishPhone(t, func(n string) {
		if fixed[danishPhoneSeries(n)] {
			nFixed++
		}
	})
	if nFixed < 2580 || nFixed > 2880 {
		t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", nFixed)
	}
}

func TestShippedDanishPhoneInternational(t *testing.T) {
	f := shipped(t, WithSeed(5))
	eachDraw(300, func() string { return fakeTemplate(t, f, "{/da_DK.phone as p}{p}|{p.national}|{p.international}") }, func(got string) {
		parts := strings.Split(got, "|")
		if len(parts) != 3 || parts[0] != parts[1] || parts[2] != "+45 "+parts[0] {
			t.Fatalf("%q: want the national form twice, then +45 and the national form", got)
		}
	})
}

func eachShippedDanishDate(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, WithSeed(3))
	eachDraw(500, func() string { return fake(t, f, "da_DK.date") }, check)
}

func danishDateYear(got string) int {
	d, _ := time.Parse("2.1.2006", got)
	return d.Year()
}

// Dansk Sprognævn's answer SV00000046 gives 2.10.2003 and advises against 02.10.2003.
func TestShippedDanishDateIsDayMonthYearWithDotsAndNoLeadingZeroFrom1970To2029(t *testing.T) {
	date := regexp.MustCompile(`^([1-9]|[12]\d|3[01])\.([1-9]|1[0-2])\.\d{4}$`)
	eachShippedDanishDate(t, func(got string) {
		d, err := time.Parse("2.1.2006", got)
		if err != nil || !date.MatchString(got) || d.Year() < 1970 || d.Year() > 2029 {
			t.Fatalf("date %q is not d.m.yyyy without leading zeros in 1970 to 2029: %v", got, err)
		}
	})
}

func TestShippedDanishDateSpans1970To2029(t *testing.T) {
	years := map[int]bool{}
	eachShippedDanishDate(t, func(got string) { years[danishDateYear(got)] = true })
	if !years[1970] || !years[2029] {
		t.Errorf("500 dates drew the years %v, want 1970 and 2029 among them", years)
	}
}

func TestShippedDanishDatesIncludeADayOfOneDigit(t *testing.T) {
	short := false
	eachShippedDanishDate(t, func(got string) { short = short || got[1] == '.' })
	if !short {
		t.Error("500 dates held no day of one digit, want one")
	}
}

var exampleEmail = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*@([a-z0-9-]+\.)*example\.(com|org|net)$`)

func TestShippedDanishEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	f := shipped(t, WithSeed(4))
	eachDraw(500, func() string { return fake(t, f, "da_DK.email") }, func(got string) {
		if !exampleEmail.MatchString(got) {
			t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
		}
	})
}

func TestShippedDanishEmailsAreMostlyDistinct(t *testing.T) {
	f := shipped(t, WithSeed(4))
	seen := tallyDraws(500, func() string { return fake(t, f, "da_DK.email") })
	if len(seen) < 400 {
		t.Errorf("500 emails drew %d distinct, want 400 or more", len(seen))
	}
}

func danishNameSlugs(t *testing.T, file string) map[string]bool {
	t.Helper()
	fold := strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
	out := map[string]bool{}
	for _, row := range tsvRows(t, file) {
		out[fold.Replace(strings.ToLower(row["name"]))] = true
	}
	return out
}

// lastNameInEmail returns the last name an address's local part spells after a first name, or "".
func lastNameInEmail(local string, firsts, lasts map[string]bool) string {
	for cut := 1; cut < len(local); cut++ {
		right := strings.TrimLeft(local[cut:], "._-")
		if firsts[strings.TrimRight(local[:cut], "._-")] && lasts[right] {
			return right
		}
	}
	return ""
}

func TestShippedDanishEmailHoldsAFirstAndALastNameInHalfItsAddresses(t *testing.T) {
	firsts, lasts := danishNameSlugs(t, "data/da_DK/first-name.tsv"), danishNameSlugs(t, "data/da_DK/last-name.tsv")
	trailingDigits := regexp.MustCompile(`\d+$`)
	f := shipped(t, WithSeed(9))
	seen := tallyDraws(5000, func() string {
		local, _, _ := strings.Cut(fake(t, f, "da_DK.email"), "@")
		return lastNameInEmail(trailingDigits.ReplaceAllString(local, ""), firsts, lasts)
	})
	named := 5000 - seen[""]
	delete(seen, "")
	if named < 2000 || len(seen) < 40 {
		t.Errorf("5000 emails held a first and a last name in %d, of %d distinct last names; want 2000 or more of 40 or more", named, len(seen))
	}
}

var danishLegalFormSuffix = regexp.MustCompile(` (ApS|A/S|I/S|K/S)$`)

func danishLegalForm(company string) string {
	if m := danishLegalFormSuffix.FindStringSubmatch(company); m != nil {
		return m[1]
	}
	return "ENK"
}

func eachShippedDanishCompany(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, WithSeed(6))
	eachDraw(shareDraws, func() string { return fake(t, f, "da_DK.company") }, check)
}

func TestShippedDanishCompanyIsANameThatMayEndInALegalForm(t *testing.T) {
	t.Parallel()
	name := regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (ApS|A/S|I/S|K/S))?$`)
	eachShippedDanishCompany(t, func(got string) {
		if !name.MatchString(got) || strings.HasSuffix(got, " ENK") {
			t.Fatalf("company %q is not a name that may end in a legal form", got)
		}
	})
}

// Statistics Denmark's table GF15, 2024: 132,403 sole proprietorships, 218,208 private limited companies,
// 30,750 public limited companies and 11,679 partnerships and limited partnerships, split evenly.
func TestShippedDanishCompanyEndsInALegalFormDrawnByStatisticsDenmarksEnterprises(t *testing.T) {
	t.Parallel()
	registered := map[string]int{"ENK": 132403, "ApS": 218208, "A/S": 30750, "I/S": 5840, "K/S": 5839}
	var total int
	for _, n := range registered {
		total += n
	}
	form := map[string]int{}
	eachShippedDanishCompany(t, func(got string) { form[danishLegalForm(got)]++ })
	for code, n := range registered {
		t.Run(code, func(t *testing.T) {
			assertShare(t, code+" of the companies", form[code], shareDraws, float64(n)/float64(total))
		})
	}
}

func TestShippedDanishCompanyPathsDrawOneFormEach(t *testing.T) {
	f := shipped(t, WithSeed(7))
	for path, suffix := range map[string]string{"enk": "", "aps": " ApS", "as": " A/S", "is": " I/S", "ks": " K/S"} {
		t.Run(path, func(t *testing.T) {
			eachDraw(100, func() string { return fake(t, f, "da_DK.company."+path) }, func(got string) {
				if suffix == "" && danishLegalFormSuffix.MatchString(got) || !strings.HasSuffix(got, suffix) {
					t.Fatalf("da_DK.company.%s = %q, want a name ending in %q", path, got, suffix)
				}
			})
		})
	}
}

func TestShippedDanishCompanyLegalIsItsAbbreviation(t *testing.T) {
	f := shipped(t, WithSeed(7))
	eachDraw(200, func() string { return fakeTemplate(t, f, "{/da_DK.company as c}{c}|{c.legal}") }, func(got string) {
		company, legal, _ := strings.Cut(got, "|")
		if legal != "" && !strings.HasSuffix(company, " "+legal) || legal == "" && strings.Contains(company, "/") {
			t.Fatalf("company %q has legal %q", company, legal)
		}
	})
}

func danishCompanyShape(name string, words []string, places, firsts, lasts map[string]bool) []string {
	var shapes []string
	for _, w := range words {
		if head, ok := strings.CutSuffix(name, " "+w); ok {
			switch {
			case places[head]:
				shapes = append(shapes, "place and trade word")
			case lasts[head]:
				shapes = append(shapes, "last name and trade word")
			}
		}
	}
	if first, last, ok := strings.Cut(name, " "); ok && firsts[first] && lasts[last] {
		shapes = append(shapes, "first and last name")
	}
	return shapes
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
	eachDraw(draws, func() string { return fake(t, f, "da_DK.company.enk") }, func(name string) {
		for _, shape := range danishCompanyShape(name, words, places, firsts, lasts) {
			seen[shape]++
		}
	})
	for _, shape := range []string{"place and trade word", "last name and trade word", "first and last name"} {
		t.Run(shape, func(t *testing.T) {
			if seen[shape] < draws/10 {
				t.Errorf("%d sole proprietorships drew a name of the shape %s %d times, want %d or more", draws, shape, seen[shape], draws/10)
			}
		})
	}
}
