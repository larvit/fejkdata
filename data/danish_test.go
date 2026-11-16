package data_test

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
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

type danishNameSet struct {
	what string
	got  map[string]int
	top  map[string]int
}

func danishNameSets(t *testing.T) []danishNameSet {
	t.Helper()
	first, last := nameCounts(t, "da_DK", "first-name"), nameCounts(t, "da_DK", "last-name")[""]
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

func TestDanishNamesFollowTheirCounts(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	seen := fejktest.TallyDraws(20000, func() string { return fejktest.FakeTemplate(t, f, "{/da_DK.last-name}") })
	if seen["Nielsen"] < 3*seen["Olsen"] || seen["Jensen"] < 3*seen["Olsen"] || seen["Nielsen"] == 0 {
		t.Errorf("Nielsen, Jensen and Olsen drawn %d, %d and %d times in 20000, want Nielsen and Jensen about five times Olsen", seen["Nielsen"], seen["Jensen"], seen["Olsen"])
	}
}

func TestShippedDanishAddressIsAStreetANumberAPostnummerAndAPostdistrikt(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[A-Z]?\n\d{4} \pL.*$`)
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, f, "da_DK.address") }, func(a string) {
		if !address.MatchString(a) {
			t.Fatalf("da_DK.address = %q, want a street and a number, then a postnummer and a postdistrikt", a)
		}
	})
}

func TestShippedDanishPlacesAreNamedAsDARWritesThem(t *testing.T) {
	districts := geoNameSet(t, "DK", "locality")
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
	f := shipped(t, fejkdata.WithSeed(1))
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
			if got := fejktest.Fake(t, f, path); got != want {
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

const danishHouseNumberDraws = 40000

func danishHouseNumbers(t *testing.T) houseNumberShares {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(11))
	return houseNumberSharesOf(danishHouseNumberDraws, func() string { return fejktest.Fake(t, f, "da_DK.address.street-number") })
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
	if rows := fejktest.TSVRows(t, "da_DK/individ.tsv"); len(rows) != 2 {
		t.Errorf("individ.tsv holds %d rows, want one per sex", len(rows))
	}
}

func TestDanishIndividHoldsOneDigitOfEachSexsParity(t *testing.T) {
	for _, row := range fejktest.TSVRows(t, "da_DK/individ.tsv") {
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
	year := fejktest.CprYear(yy, int(got[7]-'0'))
	born := time.Date(year, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
	age := fejktest.TestDay.Year() - year
	if born.AddDate(age, 0, 0).After(fejktest.TestDay) {
		age--
	}
	return danishCprBirth{born, dd, mm, year, age}
}

const danishCprDraws = 20000

func eachDanishCpr(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(danishCprDraws, func() string { return fejktest.Fake(t, f, "da_DK.cpr") }, check)
}

func TestDanishCprIsDdmmyyAndFourDigitsWithAValidCheckDigit(t *testing.T) {
	eachDanishCpr(t, func(got string) {
		if len(got) != 11 || got[6] != '-' || !fejktest.ValidCpr(got) {
			t.Fatalf("%q is not ddmmyy-ssss with a valid check digit", got)
		}
	})
}

func TestDanishCprIsBornOnADayOfTheCalendarNotAfterToday(t *testing.T) {
	eachDanishCpr(t, func(got string) {
		b := birthOfDanishCpr(got)
		if b.born.Day() != b.day || int(b.born.Month()) != b.mon || b.born.After(fejktest.TestDay) {
			t.Fatalf("%q is born on %s, want a day of the calendar not after %s", got, b.born.Format("2006-01-02"), fejktest.TestDay.Format("2006-01-02"))
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
	f := shipped(t)
	fejktest.EachDraw(500, func() string {
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
		if odd := (number[10]-'0')%2 == 1; !fejktest.ValidCpr(number) || odd != (sex == "mand") {
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
	for _, row := range fejktest.TSVRows(t, "da_DK/phone-"+kind+".tsv") {
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
	f := shipped(t, fejkdata.WithSeed(2))
	for path, want := range map[string]string{"da_DK.phone-mobile[34 2]": "34 2", "da_DK.phone-mobile[21 1]": "21 1", "da_DK.phone-fixed[34 3]": "34 3", "da_DK.phone-fixed[75 1]": "75 1"} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

func TestShippedDanishPhoneSeriesIsNoRowOfAnotherKind(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(2))
	for _, path := range []string{"da_DK.phone-fixed[34 2]", "da_DK.phone-mobile[34 3]", "da_DK.phone-mobile[20 0]", "da_DK.phone-fixed[37 1]", "da_DK.phone-fixed[67 1]", "da_DK.phone-mobile[90 1]", "da_DK.phone-fixed[80 1]", "da_DK.phone-fixed[30 0]"} {
		t.Run(path, func(t *testing.T) {
			if got, err := f.Fake(path); err == nil {
				t.Errorf("%s = %q, but the plan allocates no such series to that kind", path, got)
			}
		})
	}
}

var danishNationalPhone = regexp.MustCompile(`^(\d\d) (\d)\d( \d\d){2}$`)

func danishPhoneSeries(n string) string {
	m := danishNationalPhone.FindStringSubmatch(n)
	if m == nil {
		return ""
	}
	return m[1] + " " + m[2]
}

func danishFixedPhone(t *testing.T) func(n string) bool {
	t.Helper()
	fixed := danishSeries(t, false)
	return func(n string) bool { return fixed[danishPhoneSeries(n)] }
}

func TestShippedDanishPhoneIsEightDigitsGroupedInTwos(t *testing.T) {
	eachPhone(t, "da_DK", func(n string) {
		if !danishNationalPhone.MatchString(n) {
			t.Fatalf("phone %q is not 8 digits grouped in twos", n)
		}
	})
}

func TestShippedDanishPhoneLiesInASeriesTheNumberingPlanAllocates(t *testing.T) {
	mobile, fixed := danishSeries(t, true), danishSeries(t, false)
	eachPhone(t, "da_DK", func(n string) {
		if series := danishPhoneSeries(n); !mobile[series] && !fixed[series] {
			t.Fatalf("phone %q lies in series %q, which the plan allocates to neither mobile nor fixed numbers", n, series)
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

func danishSoleProprietorshipShapes(t *testing.T, words []string, draws int) map[string]int {
	t.Helper()
	places, firsts, lasts := nameColumnSet(t, "geo/DK/locality.tsv"), nameColumnSet(t, "da_DK/first-name.tsv"), nameColumnSet(t, "da_DK/last-name.tsv")
	f := shipped(t, fejkdata.WithSeed(8))
	seen := map[string]int{}
	fejktest.EachDraw(draws, func() string { return fejktest.Fake(t, f, "da_DK.company.enk") }, func(name string) {
		for _, shape := range danishCompanyShape(name, words, places, firsts, lasts) {
			seen[shape]++
		}
	})
	return seen
}

func TestShippedDanishCompanyNamesAreBuiltOfNamesPlacesAndTradeWords(t *testing.T) {
	t.Parallel()
	words := readJSON[[]string](t, "da_DK/company-word.json")
	if len(words) < 20 {
		t.Fatalf("company-word.json is %d words, want an array of 20 or more", len(words))
	}
	const draws = 4000
	seen := danishSoleProprietorshipShapes(t, words, draws)
	for _, shape := range []string{"place and trade word", "last name and trade word", "first and last name"} {
		t.Run(shape, func(t *testing.T) {
			if seen[shape] < draws/10 {
				t.Errorf("%d sole proprietorships drew a name of the shape %s %d times, want %d or more", draws, shape, seen[shape], draws/10)
			}
		})
	}
}
