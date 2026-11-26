package data_test

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

const ssbCutoff = 200

func eachNorwegianRow(t *testing.T, file string, check func(row map[string]string)) {
	t.Helper()
	for _, row := range fejktest.TSVRows(t, "nb_NO/"+file+".tsv") {
		check(row)
	}
}

func eachNorwegianGeoRow(t *testing.T, table string, check func(r []string)) {
	t.Helper()
	for _, r := range geoRows(t, "NO", table) {
		check(r)
	}
}

func TestNorwegianNamesAreBorneBy200OrMore(t *testing.T) {
	for _, file := range []string{"first-name", "last-name"} {
		t.Run(file, func(t *testing.T) {
			eachNorwegianRow(t, file, func(row map[string]string) {
				if n, err := strconv.Atoi(row["count"]); err != nil || n < ssbCutoff {
					t.Errorf("%s row %v is borne by fewer than %d persons, which SSB does not publish", file, row, ssbCutoff)
				}
			})
		})
	}
}

func norwegianLastNamesMostCommonFirst(t *testing.T) []string {
	t.Helper()
	counts := map[string]int{}
	var names []string
	eachNorwegianRow(t, "last-name", func(row map[string]string) {
		n, err := strconv.Atoi(row["count"])
		if err != nil {
			t.Fatalf("last-name.tsv row %v has no count", row)
		}
		counts[row["name"]] = n
		names = append(names, row["name"])
	})
	sort.Slice(names, func(i, j int) bool { return counts[names[i]] > counts[names[j]] })
	return names
}

func TestNorwegianLastNamesStartWithHansenJohansenOlsen(t *testing.T) {
	names := norwegianLastNamesMostCommonFirst(t)
	if want := []string{"Hansen", "Johansen", "Olsen"}; !slices.Equal(names[:3], want) {
		t.Errorf("the three most common last names are %v, want %v", names[:3], want)
	}
}

var norwegianMonthDays = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func TestNorwegianBirthDayIsADayOfAMonthPlus80Once(t *testing.T) {
	seen := map[string]bool{}
	eachNorwegianRow(t, "birth-day", func(row map[string]string) {
		ddmm := row["ddmm"]
		dd, errDay := strconv.Atoi(ddmm[:2])
		mm, errMonth := strconv.Atoi(ddmm[2:])
		if len(ddmm) != 4 || errDay != nil || errMonth != nil || mm < 81 || mm > 92 || dd < 1 || dd > norwegianMonthDays[mm-81] || seen[ddmm] {
			t.Errorf("birth-day.tsv row %q is not a day of a month plus 80, once", ddmm)
		}
		seen[ddmm] = true
	})
}

func TestNorwegianBirthDaysAreTheDaysOfAYearWithout29February(t *testing.T) {
	if rows := fejktest.TSVRows(t, "nb_NO/birth-day.tsv"); len(rows) != 365 {
		t.Errorf("birth-day.tsv holds %d days, want 365, a year without 29 February", len(rows))
	}
}

type norwegianFnr struct{ day, mon, yy, individ, year int }

// readNorwegianFnr reads a number as Skatteetaten defines it: six digits of date with 80 added
// to the month, three individ digits whose range gives the century, and two check digits.
// Its year is 0 where the individ and the year digits are a pair fejkdata does not draw.
func readNorwegianFnr(got string) norwegianFnr {
	dd, _ := strconv.Atoi(got[0:2])
	mm, _ := strconv.Atoi(got[2:4])
	yy, _ := strconv.Atoi(got[4:6])
	individ, _ := strconv.Atoi(got[6:9])
	year := 0
	switch {
	case individ <= 499:
		year = 1900 + yy
	case yy < 40:
		year = 2000 + yy
	}
	return norwegianFnr{dd, mm, yy, individ, year}
}

func eachNorwegianFodselsnummer(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(5000, func() string { return fejktest.Fake(t, f, "nb_NO.fodselsnummer") }, check)
}

func TestNorwegianFodselsnummerHasValidCheckDigits(t *testing.T) {
	eachNorwegianFodselsnummer(t, func(got string) {
		if !fejktest.ValidFnr(got) {
			t.Fatalf("%q is not a fødselsnummer with valid check digits", got)
		}
	})
}

func TestNorwegianFodselsnummerIsBornOnADayOfAMonthPlus80(t *testing.T) {
	eachNorwegianFodselsnummer(t, func(got string) {
		if n := readNorwegianFnr(got); n.mon < 81 || n.mon > 92 || n.day < 1 || n.day > norwegianMonthDays[n.mon-81] {
			t.Fatalf("%q has day %d of month %d, want a day of a month plus 80", got, n.day, n.mon)
		}
	})
}

func TestNorwegianFodselsnummerPairsItsIndividWithAYearOfItsCentury(t *testing.T) {
	eachNorwegianFodselsnummer(t, func(got string) {
		if n := readNorwegianFnr(got); n.year == 0 {
			t.Fatalf("%q has individ %d with year %d, a pair fejkdata does not draw", got, n.individ, n.yy)
		}
	})
}

func TestNorwegianFodselsnummerIsBornFrom1940To2025(t *testing.T) {
	eachNorwegianFodselsnummer(t, func(got string) {
		if year := readNorwegianFnr(got).year; year < 1940 || year > 2025 {
			t.Fatalf("%q is born in %d, want 1940 to 2025", got, year)
		}
	})
}

func TestNorwegianFodselsnummerDrawsBirthsBeforeAndFrom2000(t *testing.T) {
	var before2000, from2000 int
	eachNorwegianFodselsnummer(t, func(got string) {
		if readNorwegianFnr(got).year < 2000 {
			before2000++
		} else {
			from2000++
		}
	})
	if before2000 == 0 || from2000 == 0 {
		t.Errorf("5000 numbers held %d born before 2000 and %d from 2000, want both", before2000, from2000)
	}
}

func eachNorwegianFodselsnummerWithASex(t *testing.T, check func(got, sex, number string)) {
	t.Helper()
	f := shipped(t)
	fejktest.EachDraw(500, func() string {
		got, err := f.FakeTemplate("{/nb_NO.sex as s}{s.name}|{/nb_NO.fodselsnummer.birth}{fnr(s.individ.digit)}")
		if err != nil {
			t.Fatal(err)
		}
		return got
	}, func(got string) {
		sex, number, _ := strings.Cut(got, "|")
		check(got, sex, number)
	})
}

func TestNorwegianFodselsnummerAgreesWithABoundSex(t *testing.T) {
	eachNorwegianFodselsnummerWithASex(t, func(got, sex, number string) {
		if odd := (number[8]-'0')%2 == 1; !fejktest.ValidFnr(number) || odd != (sex == "mann") {
			t.Fatalf("%q: want a valid number, odd for a mann and even for a kvinne", got)
		}
	})
}

func TestNorwegianFodselsnummerDrawsBothSexesWhenItsSexIsBound(t *testing.T) {
	seen := map[string]bool{}
	eachNorwegianFodselsnummerWithASex(t, func(_, sex, _ string) { seen[sex] = true })
	if len(seen) != 2 {
		t.Errorf("500 draws gave the sexes %v, want both", seen)
	}
}

func TestNorwegianFodselsnummerDrawsEveryThirdIndividDigitOfABoundSexsParity(t *testing.T) {
	thirds := map[string]map[byte]bool{"kvinne": {}, "mann": {}}
	eachNorwegianFodselsnummerWithASex(t, func(_, sex, number string) { thirds[sex][number[8]] = true })
	for sex, digits := range thirds {
		t.Run(sex, func(t *testing.T) {
			if len(digits) != 5 {
				t.Errorf("500 draws gave a %s the third individ digits %v, want all five of its parity", sex, digits)
			}
		})
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

func norwegianFixedPhone(*testing.T) func(n string) bool { return norwegianFixed.MatchString }

func TestShippedNorwegianPhoneIsAMobileOrAFixedNumberGroupedInTwos(t *testing.T) {
	eachPhone(t, "nb_NO", func(n string) {
		if !norwegianMobile.MatchString(n) && !norwegianFixed.MatchString(n) {
			t.Fatalf("phone %q is neither a mobile nor a fixed number grouped in twos", n)
		}
	})
}

func TestShippedNorwegianPhoneTablesHoldFourDigitPrefixesOfSeriesNkomHasAssigned(t *testing.T) {
	for kind, series := range map[string]*regexp.Regexp{
		"mobile": regexp.MustCompile(`^[49]\d \d{2}$`),
		"fixed":  regexp.MustCompile(`^[23567]\d \d{2}$`),
	} {
		t.Run(kind, func(t *testing.T) {
			seen := map[string]bool{}
			eachNorwegianRow(t, "phone-"+kind+"-prefix", func(row map[string]string) {
				p := row["prefix"]
				if !series.MatchString(p) || !slices.Contains(nkomAssigned[kind], p[:2]) || seen[p] {
					t.Errorf("phone-%s-prefix row %q is not four digits in a series Nkom has assigned to providers, or repeats one", kind, p)
				}
				seen[p] = true
			})
		})
	}
}

// Nkom assigns 56 12, 40 00 and 90 00 whole. It leaves 56 20 to 56 29 and 47 55 free, blocks 45 36,
// gives 99 90 to global title addresses, and assigns only some thousands of 56 97 and 56 98.
func TestShippedNorwegianPhonePrefixIsARowWhereNkomHasAssignedItWhole(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(2))
	for path, want := range map[string]string{"nb_NO.phone-fixed-prefix[56 12]": "56 12", "nb_NO.phone-mobile-prefix[40 00]": "40 00", "nb_NO.phone-mobile-prefix[90 00]": "90 00"} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

func TestShippedNorwegianPhonePrefixIsNoRowWhereNkomHasNotAssignedItWhole(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(2))
	for _, path := range []string{"nb_NO.phone-fixed-prefix[56 20]", "nb_NO.phone-fixed-prefix[56 97]", "nb_NO.phone-fixed-prefix[56 98]", "nb_NO.phone-mobile-prefix[45 36]", "nb_NO.phone-mobile-prefix[47 55]", "nb_NO.phone-mobile-prefix[99 90]"} {
		t.Run(path, func(t *testing.T) {
			if got, err := f.Fake(path); err == nil {
				t.Errorf("%s = %q, but Nkom has not assigned all of it", path, got)
			}
		})
	}
}

// oneWord is, per form, the weight of the shipped shapes that are one word and the weight of all its
// shipped shapes, as company-no.py wrote them; the test holds a render to them.
func TestShippedNorwegianCompanyNamesFollowTheShapesOfEnhetsregisteret(t *testing.T) {
	t.Parallel()
	oneWord := map[string][2]int{"AS": {101525, 337759}, "ENK": {9115, 350561}, "SA": {874, 5505}, "DA": {2067, 6829}, "ANS": {1165, 3980}, "ASA": {88, 186}, "KS": {16, 83}}
	const draws = fejktest.ShareDraws / 10
	for form, c := range oneWord {
		t.Run(form, func(t *testing.T) {
			f := shipped(t, fejkdata.WithSeed(8))
			var single int
			fejktest.EachDraw(draws, func() string { return fejktest.Fake(t, f, "nb_NO.company."+strings.ToLower(form)) }, func(got string) {
				if name := strings.TrimSuffix(got, " "+form); !strings.Contains(name, " ") {
					single++
				}
			})
			assertShare(t, form+" names of one word", single, draws, float64(c[0])/float64(c[1]))
		})
	}
}

func TestNorwegianCompanyWordsAreRepeatedInEnhetsregisteret(t *testing.T) {
	word := regexp.MustCompile(`^\p{Lu}\p{L}{2,}$`)
	eachNorwegianRow(t, "company-word", func(row map[string]string) {
		if n, err := strconv.Atoi(row["count"]); err != nil || n < 100 || !word.MatchString(row["word"]) {
			t.Errorf("row %v, want a word of three or more letters, capitalised, that 100 or more register names carry", row)
		}
	})
}

func TestNorwegianCompanyWordsHoldWordsThatManyRegisterNamesCarry(t *testing.T) {
	have := map[string]bool{}
	eachNorwegianRow(t, "company-word", func(row map[string]string) { have[row["word"]] = true })
	for _, w := range []string{"Holding", "Invest", "Eiendom", "Consulting", "Bygg"} {
		t.Run(w, func(t *testing.T) {
			if !have[w] {
				t.Errorf("company-word.tsv omits %q, a word that 100 or more register names carry", w)
			}
		})
	}
}

func norwegianLastNameSlugs(t *testing.T) []string {
	t.Helper()
	var slugs []string
	fold := strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
	ascii := regexp.MustCompile(`^[a-z]{6,}$`)
	eachNorwegianRow(t, "last-name", func(row map[string]string) {
		if slug := fold.Replace(strings.ToLower(row["name"])); ascii.MatchString(slug) {
			slugs = append(slugs, slug)
		}
	})
	return slugs
}

func slugsHeldByAnEmail(slugs []string, emails map[string]int) map[string]bool {
	seen := map[string]bool{}
	for email := range emails {
		for _, slug := range slugs {
			if strings.Contains(email, slug) {
				seen[slug] = true
			}
		}
	}
	return seen
}

func TestShippedNorwegianEmailDrawsItsNamesFromSSB(t *testing.T) {
	slugs := norwegianLastNameSlugs(t)
	f := shipped(t, fejkdata.WithSeed(9))
	seen := slugsHeldByAnEmail(slugs, fejktest.TallyDraws(5000, func() string { return fejktest.Fake(t, f, "nb_NO.email") }))
	if len(seen) < 100 {
		t.Errorf("5000 emails carry %d distinct last names of SSB's, want 100 or more", len(seen))
	}
}

func TestShippedNorwegianCompanyNamesVary(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(7))
	seen := fejktest.TallyDraws(1000, func() string { return fejktest.Fake(t, f, "nb_NO.company") })
	if len(seen) < 700 {
		t.Errorf("1000 companies drew %d distinct, want 700 or more", len(seen))
	}
}

func TestShippedNorwegianAddressIsAStreetANumberAPostnummerAndAPoststed(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	address := regexp.MustCompile(`^[\pL\d].* [1-9]\d{0,3}[A-Z]?\n\d{4} \pL.*$`)
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, f, "nb_NO.address") }, func(a string) {
		if !address.MatchString(a) {
			t.Fatalf("nb_NO.address = %q, want a street and a number, then a postnummer and a poststed", a)
		}
	})
}

// Kartverket's address register spells a poststed in capitals, MO I RANA; a place name is written Mo i Rana.
func TestShippedNorwegianPoststedIsCasedAsAPlaceName(t *testing.T) {
	capitals := regexp.MustCompile(`\p{Lu}{2}`)
	eachNorwegianGeoRow(t, "locality", func(r []string) {
		if capitals.MatchString(r[0]) {
			t.Errorf("geo.NO: poststed %q holds a word in capitals", r[0])
		}
	})
}

func TestShippedNorwegianPlacesHoldPoststedsUnderTheirPlaceNames(t *testing.T) {
	names := geoNameSet(t, "NO", "locality")
	for _, name := range []string{"Oslo", "Bergen", "Mo i Rana", "Kristiansund N", "Nes på Hedmarken", "Ålesund"} {
		t.Run(name, func(t *testing.T) {
			if !names[name] {
				t.Errorf("geo.NO: no poststed %q", name)
			}
		})
	}
}

func norwegianPoststedsThatAreNoName(t *testing.T) map[string]bool {
	t.Helper()
	named := map[string]bool{}
	for _, file := range []string{"first-name", "last-name"} {
		eachNorwegianRow(t, file, func(row map[string]string) { named[strings.ToLower(row["name"])] = true })
	}
	places := map[string]bool{}
	eachNorwegianGeoRow(t, "locality", func(r []string) {
		if !named[strings.ToLower(r[0])] {
			places[r[0]] = true
		}
	})
	return places
}

// A poststed that is also a name, such as Berg, counts as the name. A seeded render of the AS shapes
// company-no.py wrote held a poststed that is no name in 2.15% of AS names; the test holds a render to that share.
func TestShippedNorwegianCompanyNamesHoldAPoststed(t *testing.T) {
	places := norwegianPoststedsThatAreNoName(t)
	f := shipped(t, fejkdata.WithSeed(10))
	const draws = 20000
	holding := 0
	fejktest.EachDraw(draws, func() string { return fejktest.Fake(t, f, "nb_NO.company.as") }, func(got string) {
		if holdsPlace(strings.Fields(strings.TrimSuffix(got, " AS")), places) {
			holding++
		}
	})
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
	eachNorwegianGeoRow(t, "street", func(r []string) {
		if refused[r[0]] {
			t.Errorf("geo.NO: street %q in %s, want no numbered placeholder, road designation or name in capitals", r[0], r[1])
		}
	})
}

// SSB names a kommune or fylke in each of its official languages, Trondheim - Tråante; Kartverket's
// address register writes the Norwegian one.
func TestShippedNorwegianKommuneAndFylkeSelectByTheirNorwegianNames(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(1))
	for path, want := range map[string]string{
		"geo.NO.municipality[Trondheim].code":  "5001",
		"geo.NO.municipality[Kautokeino].code": "5612",
		"geo.NO.municipality[Oslo].code":       "0301",
		"geo.NO.region[Troms].code":            "55",
		"geo.NO.region[Trøndelag].code":        "50",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("%s = %q, want %q", path, got, want)
			}
		})
	}
}

const norwegianHouseNumberDraws = 40000

func norwegianHouseNumbers(t *testing.T) houseNumberShares {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(11))
	return houseNumberSharesOf(norwegianHouseNumberDraws, func() string { return fejktest.Fake(t, f, "nb_NO.address.street-number") })
}

// Of Kartverket's 2.57 million street addresses, 18.3% carry a letter after the number, 22.2% a
// number of one digit and 3.5% one of four.
func TestShippedNorwegianHouseNumbersCarryALetterAsOftenAsKartverketsRegister(t *testing.T) {
	assertShare(t, "house numbers with a letter", norwegianHouseNumbers(t).lettered, norwegianHouseNumberDraws, 0.183)
}

func TestShippedNorwegianHouseNumbersHaveOneDigitAsOftenAsKartverketsRegister(t *testing.T) {
	assertShare(t, "house numbers of one digit", norwegianHouseNumbers(t).oneDigit, norwegianHouseNumberDraws, 0.222)
}

func TestShippedNorwegianHouseNumbersHaveFourDigitsAsOftenAsKartverketsRegister(t *testing.T) {
	assertShare(t, "house numbers of four digits", norwegianHouseNumbers(t).fourDigits, norwegianHouseNumberDraws, 0.035)
}

func norwegianPlaceWords(t *testing.T) map[string]bool {
	t.Helper()
	words := map[string]bool{}
	for _, table := range []string{"locality", "municipality"} {
		for _, row := range fejktest.TSVRows(t, "geo/NO/"+table+".tsv") {
			for _, w := range strings.Fields(strings.ToLower(row["name"])) {
				words[w] = true
			}
		}
	}
	return words
}

func norwegianCoinedRoots(t *testing.T) []string {
	t.Helper()
	return readJSON[struct{ A []string }](t, "nb_NO/company-coined.json").A
}

func TestNorwegianCoinedRootsHoldNoWordOfAPlace(t *testing.T) {
	words := norwegianPlaceWords(t)
	for _, root := range norwegianCoinedRoots(t) {
		t.Run(root, func(t *testing.T) {
			if words[strings.ToLower(root)] {
				t.Errorf("company-coined root %q is a word of a geo/NO place name", root)
			}
		})
	}
}

func TestNorwegianCompanyLegalIsEachOfTheCommonForms(t *testing.T) {
	t.Parallel()
	seen := map[string]int{}
	eachCompanyWithItsLegal(t, "nb_NO", func(_, legal string) { seen[legal]++ })
	for _, form := range []string{"", "AS", "DA", "SA", "ANS"} {
		t.Run(fmt.Sprintf("legal %q", form), func(t *testing.T) {
			if seen[form] == 0 {
				t.Errorf("no company of %d has the legal %q", companyLegalDraws, form)
			}
		})
	}
}
