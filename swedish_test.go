package fejkdata

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestSwedishPersonNamesHaveNoTripleLetter pins an orthographic rule the shape
// regexes miss: no generated name repeats a character three times over.
func TestSwedishPersonNamesHaveNoTripleLetter(t *testing.T) {
	f := shipped(t, WithSeed(11))
	for _, path := range []string{"sv_SE.person", "sv_SE.person.last"} {
		for i := 0; i < 20000; i++ {
			name := fake(t, f, path)
			r := []rune(name)
			for j := 0; j+2 < len(r); j++ {
				if r[j] == r[j+1] && r[j] == r[j+2] {
					t.Fatalf("%s = %q, which triples %q", path, name, string(r[j]))
				}
			}
		}
	}
}

// Bolagsverket's group digits are 5 for aktiebolag, 9 for handelsbolag and kommanditbolag and 7 and 8
// for föreningar; digits three and four of 20 or more tell an organisationsnummer from a personnummer.
func TestSwedishOrganisationsnummer(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	re := regexp.MustCompile(`^[5789]\d[2-9]\d{3}-\d{4}$`)
	group := map[byte]int{}
	for i := 0; i < 5000; i++ {
		v := fake(t, sv, "sv_SE.organisationsnummer")
		if !re.MatchString(v) || !luhnValid(digitsOnly(v)) {
			t.Fatalf("organisationsnummer %q, want a group digit 5, 7, 8 or 9, a digit 2 to 9 third, and a Luhn check digit", v)
		}
		group[v[0]]++
	}
	if group['5'] < 4690 || group['5'] > 4840 || group['9'] <= group['7'] || group['9'] <= group['8'] || group['7'] == 0 || group['8'] == 0 {
		t.Fatalf("group digits drawn %v, want 5 for 95.3%%, 9 for 3.2%% and 7 and 8 for 0.7%% each", group)
	}
}

// Bolagsverket's new registrations of 2025: 50,859 aktiebolag, 8,468 enskilda näringsidkare, 1,557
// handelsbolag, 398 bostadsrättsföreningar, 319 ekonomiska föreningar and 176 kommanditbolag. Each
// expected share comes from those counts, and each bound is five standard deviations of a draw.
func TestSwedishCompanyLegalFormsFollowBolagsverketsRegistrations(t *testing.T) {
	t.Parallel()
	sv := shipped(t, WithSeed(1))
	const draws = shareDraws
	form := map[string]int{}
	for i := 0; i < draws; i++ {
		v := strings.ToLower(fake(t, sv, "sv_SE.company"))
		switch {
		case strings.HasSuffix(v, " ab"), strings.HasSuffix(v, " aktiebolag"):
			form["AB"]++
		case strings.HasPrefix(v, "bostadsrättsföreningen "):
			form["BRF"]++
		case strings.HasSuffix(v, " ekonomisk förening"), strings.HasSuffix(v, " ek. för."):
			form["EK"]++
		case strings.HasSuffix(v, " handelsbolag"), strings.HasPrefix(v, "handelsbolaget "):
			form["HB"]++
		case strings.HasSuffix(v, " kommanditbolag"), strings.HasPrefix(v, "kommanditbolaget "):
			form["KB"]++
		default:
			form["E"]++
		}
	}
	for f, registered := range map[string]int{"AB": 50859, "E": 8468, "HB": 1557, "BRF": 398, "EK": 319, "KB": 176} {
		assertShare(t, f+" of the companies", form[f], draws, float64(registered)/61777)
	}
}

// The wordings of a legal form and the one-word names are drawn as often as the register's active companies
// carry them. The counts are those data-import/company-se.py wrote into company.json from the register: a
// wording's over the names not in capitals, and oneWord's the names of one coined word against all of a form's.
func TestSwedishCompanyWordingsAndShapesFollowTheRegister(t *testing.T) {
	t.Parallel()
	type wording struct {
		form, text string
		before     bool
		count      int
	}
	wordings := []wording{
		{"AB", " AB", false, 718311}, {"AB", " Aktiebolag", false, 60401},
		{"HB", " Handelsbolag", false, 20268}, {"HB", "Handelsbolaget ", true, 1526}, {"HB", " handelsbolag", false, 1232},
		{"KB", " Kommanditbolag", false, 7885}, {"KB", "Kommanditbolaget ", true, 1160}, {"KB", " kommanditbolag", false, 387},
		{"EK", " Ekonomisk förening", false, 3730}, {"EK", " ekonomisk förening", false, 3166}, {"EK", " ek. för.", false, 1022}, {"EK", " Ek. för.", false, 587},
		{"BRF", "Bostadsrättsföreningen ", true, 26480},
	}
	oneWord := map[string][2]int{"AB": {144879, 552375}, "E": {34479, 134451}, "HB": {5046, 15696}, "BRF": {8409, 22408}, "EK": {1251, 6563}, "KB": {1754, 6611}}
	sv := shipped(t, WithSeed(2))
	const draws = shareDraws
	written, form, single := map[string]int{}, map[string]int{}, map[string]int{}
	for i := 0; i < draws; i++ {
		v := fake(t, sv, "sv_SE.company")
		f, name := "E", v
		for _, w := range wordings {
			if w.before && strings.HasPrefix(v, w.text) {
				f, name = w.form, strings.TrimPrefix(v, w.text)
				written[w.text]++
			} else if !w.before && strings.HasSuffix(v, w.text) {
				f, name = w.form, strings.TrimSuffix(v, w.text)
				written[w.text]++
			}
		}
		form[f]++
		if !strings.Contains(name, " ") {
			single[f]++
		}
	}
	for _, w := range wordings {
		var all int
		for _, o := range wordings {
			if o.form == w.form {
				all += o.count
			}
		}
		assertShare(t, strings.TrimSpace(w.text)+" among the "+w.form, written[w.text], form[w.form], float64(w.count)/float64(all))
	}
	for f, c := range oneWord {
		assertShare(t, f+" names of one word", single[f], form[f], float64(c[0])/float64(c[1]))
	}
}

func assertShare(t *testing.T, what string, got, of int, want float64) {
	t.Helper()
	sd := math.Sqrt(want * (1 - want) / float64(of))
	if g := float64(got) / float64(of); math.Abs(g-want) > 5*sd {
		t.Errorf("%s: %d of %d (%.3f%%), want %.3f%% within %.3f%%", what, got, of, g*100, want*100, 5*sd*100)
	}
}

// company-word.tsv holds a word only where the register repeats it, so no row names one company.
func TestSwedishCompanyWordsAreRepeatedInTheRegister(t *testing.T) {
	b, err := os.ReadFile("data/sv_SE/company-word.tsv")
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(b)), "\n")
	if rows[0] != "word\tcount" {
		t.Fatalf("header %q, want word and count", rows[0])
	}
	word := regexp.MustCompile(`^\p{Lu}\p{L}{2,}$`)
	for _, row := range rows[1:] {
		w, c, _ := strings.Cut(row, "\t")
		if n, err := strconv.Atoi(c); err != nil || n < 100 || !word.MatchString(w) {
			t.Errorf("row %q, want a word of three or more letters, capitalised, that 100 or more register names carry", row)
		}
	}
	for _, w := range []string{"Bygg", "Konsult", "Fastigheter", "VVS"} {
		if !slices.ContainsFunc(rows, func(r string) bool { return strings.HasPrefix(r, w+"\t") }) {
			t.Errorf("company-word.tsv omits %q, a word that 100 or more register names carry", w)
		}
	}
}

func TestSwedishVat(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	vat := regexp.MustCompile(`^SE([5789]\d[2-9]\d{7})01$`)
	for i := 0; i < 2000; i++ {
		n := fake(t, sv, "sv_SE.vat")
		if m := vat.FindStringSubmatch(n); m == nil || !luhnValid(m[1]) {
			t.Fatalf("sv_SE.vat %q, want SE + a Luhn-valid organisationsnummer + 01", n)
		}
		pair := fakeTemplate(t, sv, "{/sv_SE.organisationsnummer as o}{o}|{o.vat}")
		nr, v, _ := strings.Cut(pair, "|")
		if v != "SE"+strings.ReplaceAll(nr, "-", "")+"01" {
			t.Fatalf("a bound organisationsnummer %q has the VAT number %q, want SE + its digits + 01", nr, v)
		}
	}
}

// The 99 combinations that Transportstyrelsen's page listed as blocked when it was updated on 2026-03-31.
const blockedPlates = "APA ARG ASS BAJ BSS CUC CUK CUM DUM ETA ETT FAG FAN FEG FEL FEM FES FET FNL FUC FUK FUL GAM GAY GEJ GEY GHB GUD GYN HAT HBT HKH HOR HOT KGB KKK KUC KUF KUG KUK KYK LAM LAT LEM LOJ LSD LUS MAD MAO MEN MES MLB MUS NAZ NRP NSF NYP OND OOO ORM PAJ PKK PLO PMS PUB RAP RAS ROM RPS RUS SEG SEX SJU SOS SPY SUG SUP SUR TBC TOA TOK TRE TYP UFO USA WAM WAR WWW XTC XTZ XUK XXL XXX ZEX ZOG ZPY ZUG ZUP ZOO"

func TestSwedishLicensePlateLettersAreEveryCombinationButTheBlocked(t *testing.T) {
	b, err := os.ReadFile("data/sv_SE/licenseplate-letters.tsv")
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(b)), "\n")
	if rows[0] != "letters" {
		t.Fatalf("header %q, want letters", rows[0])
	}
	blocked := strings.Fields(blockedPlates)
	if len(blocked) != 99 {
		t.Fatalf("the test holds %d blocked combinations, want 99", len(blocked))
	}
	const alphabet = "ABCDEFGHJKLMNOPRSTUWXYZ"
	want := len(alphabet)*len(alphabet)*len(alphabet) - len(blocked)
	if len(rows)-1 != want {
		t.Fatalf("%d rows, want %d: 23 letters cubed less the 99 blocked", len(rows)-1, want)
	}
	seen := map[string]bool{}
	for _, r := range rows[1:] {
		if len(r) != 3 || strings.Trim(r, alphabet) != "" {
			t.Fatalf("row %q, want three of %s", r, alphabet)
		}
		if seen[r] {
			t.Fatalf("row %q repeats", r)
		}
		seen[r] = true
	}
	for _, c := range blocked {
		if seen[c] {
			t.Errorf("row %q is blocked by Transportstyrelsen", c)
		}
	}
}

func TestSwedishLicensePlate(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	plate := regexp.MustCompile(`^([A-HJ-PR-UW-Z]{3}) (?:(\d{3})|(\d{2})([A-HJ-NPR-UW-Z]))$`)
	blocked := strings.Fields(blockedPlates)
	var old, current int
	for i := 0; i < 5000; i++ {
		v := fake(t, sv, "sv_SE.licenseplate")
		m := plate.FindStringSubmatch(v)
		if m == nil || slices.Contains(blocked, m[1]) || m[2] == "000" {
			t.Fatalf("sv_SE.licenseplate %q, want ABC 123 or ABC 12A: no I, Q or V, no O in the last place, no blocked combination and no 000", v)
		}
		if m[2] != "" {
			old++
		} else {
			current++
		}
	}
	if old < 2000 || current < 2000 {
		t.Errorf("%d plates of the form ABC 123 and %d of ABC 12A in 5000, want about half each", old, current)
	}
}

// ABC 123 runs from 001 to 999 and ABC 12A from 00A to 99Z, each number as often as another, and
// the last character is any of the 22 letters but O.
func TestSwedishLicensePlateNumbersAndLastLetters(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	const draws = 30000
	old, current, last := map[string]int{}, map[string]int{}, map[string]int{}
	for i := 0; i < draws; i++ {
		old[fake(t, sv, "sv_SE.licenseplate.old-number")]++
		current[fake(t, sv, "sv_SE.licenseplate.current-number")]++
		last[fake(t, sv, "sv_SE.licenseplate.last")]++
	}
	var below100, below10 int
	for n := 1; n <= 999; n++ {
		c := old[fmt.Sprintf("%03d", n)]
		if c == 0 {
			t.Errorf("sv_SE.licenseplate.old-number never drew %03d in %d draws", n, draws)
		}
		if n < 100 {
			below100 += c
		}
		if n < 10 {
			below10 += c
		}
	}
	if len(old) != 999 {
		t.Errorf("old-number drew %d distinct numbers, want the 999 from 001 to 999", len(old))
	}
	assertShare(t, "old numbers below 100", below100, draws, 99.0/999)
	assertShare(t, "old numbers below 10", below10, draws, 9.0/999)
	for n := 0; n <= 99; n++ {
		if current[fmt.Sprintf("%02d", n)] == 0 {
			t.Errorf("sv_SE.licenseplate.current-number never drew %02d in %d draws", n, draws)
		}
	}
	if len(current) != 100 {
		t.Errorf("current-number drew %d distinct numbers, want the 100 from 00 to 99", len(current))
	}
	var got []string
	for l := range last {
		got = append(got, l)
	}
	slices.Sort(got)
	if strings.Join(got, "") != "ABCDEFGHJKLMNPRSTUWXYZ" {
		t.Errorf("last drew %v, want each of A-Z but I, O, Q and V", got)
	}
}

// TestSwedishPersonnummer checks what the shape regex can't: the date is a real
// calendar date, the birth number is Skatteverket's test series (238 female, 239
// male), the trailing digit is a Luhn checksum over the other nine, and a
// samordningsnummer is the same with 60 added to the day.
func TestSwedishPersonnummer(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	re := regexp.MustCompile(`^\d{6}-23[89]\d$`)
	reShort := regexp.MustCompile(`^\d{6}[-+]23[89]\d$`)
	sawLongMonthEnd, birth := false, map[string]bool{}
	for i := 0; i < 2000; i++ {
		v := fake(t, sv, "sv_SE.personnummer")
		d := digitsOnly(v)
		if !reShort.MatchString(v) {
			t.Fatalf("personnummer %q, want YYMMDD-238C or YYMMDD-239C, or + for -", v)
		}
		if _, err := time.Parse("060102", d[:6]); err != nil {
			t.Fatalf("personnummer %q is not a valid calendar date: %v", v, err)
		}
		if !luhnValid(d) {
			t.Fatalf("personnummer %q fails the Luhn check", v)
		}
		sawLongMonthEnd = sawLongMonthEnd || d[4:6] == "31"
		birth[d[6:9]] = true
		s := fake(t, sv, "sv_SE.samordningsnummer")
		sd := digitsOnly(s)
		day, err := strconv.Atoi(sd[4:6])
		if !re.MatchString(s) || err != nil || day < 61 || day > 88 || !luhnValid(sd) {
			t.Fatalf("samordningsnummer %q, want YYMM(61-88)-23[89]C, Luhn-valid", s)
		}
		if _, err := time.Parse("0601", sd[:4]); err != nil {
			t.Fatalf("samordningsnummer %q is not a valid year and month: %v", s, err)
		}
	}
	if !sawLongMonthEnd {
		t.Fatal("never generated a 31st")
	}
	if !birth["238"] || !birth["239"] {
		t.Fatalf("birth numbers drawn %v, want both 238 and 239", birth)
	}
}

// PTS's numbering plan of 2024-01-08 gives each area code an N(S)N, the digits after the
// trunk 0, of at most 9. The shortest is 8 for a three-digit code and for 18, 19, 31 and
// 54, and 7 for every other one- or two-digit code.
func swedishPhoneShortest(ndc string) int {
	if len(ndc) == 3 || slices.Contains([]string{"18", "19", "31", "54"}, ndc) {
		return 8
	}
	return 7
}

// Språkrådet groups the subscriber digits, the ones after the area code, in twos and
// threes from the left; "08-575 686 43" and "08-57 56 86 43" both group eight digits.
var swedishPhoneGrouped = map[int]*regexp.Regexp{
	5: regexp.MustCompile(`^[1-9]\d{2} \d{2}$`),
	6: regexp.MustCompile(`^[1-9]\d \d{2} \d{2}$`),
	7: regexp.MustCompile(`^[1-9]\d{2} \d{2} \d{2}$`),
	8: regexp.MustCompile(`^([1-9]\d{2} \d{3} \d{2}|[1-9]\d \d{2} \d{2} \d{2})$`),
}

func TestShippedSwedishPhoneMobile(t *testing.T) {
	f := shipped(t, WithSeed(11))
	re := regexp.MustCompile(`^0(70|72|73|76|79)-\d{3} \d{2} \d{2}$`)
	seen := map[string]bool{}
	for i := 0; i < 3000; i++ {
		if m := re.FindStringSubmatch(fake(t, f, "sv_SE.phone")); m != nil {
			seen[m[1]] = true
		}
	}
	if len(seen) != 5 {
		t.Fatalf("3000 phones drew the mobile codes %v, want 70, 72, 73, 76 and 79", seen)
	}
}

func TestShippedSwedishPhoneFixedIsGroupedAndLongEnough(t *testing.T) {
	f := shipped(t, WithSeed(11))
	mobile := regexp.MustCompile(`^0(70|72|73|76|79)-`)
	fixed := regexp.MustCompile(`^0(\d{1,3})-(\d[\d ]+)$`)
	for i := 0; i < 6000; i++ {
		n := fake(t, f, "sv_SE.phone")
		if mobile.MatchString(n) {
			continue
		}
		m := fixed.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q is neither a mobile nor a fixed number", n)
		}
		ndc, subscriber := m[1], m[2]
		digits := len(strings.ReplaceAll(subscriber, " ", ""))
		if re := swedishPhoneGrouped[digits]; re == nil || !re.MatchString(subscriber) {
			t.Fatalf("phone %q: subscriber %q is not grouped as Språkrådet writes it, or starts with the trunk 0", n, subscriber)
		}
		if total := len(ndc) + digits; total < swedishPhoneShortest(ndc) || total > 9 {
			t.Fatalf("phone %q has %d digits after the 0, PTS gives area code %s %d to 9", n, total, ndc, swedishPhoneShortest(ndc))
		}
	}
}

func TestShippedSwedishPhoneAreaDrawsEachLengthItAllows(t *testing.T) {
	f := shipped(t, WithSeed(11))
	for _, c := range []struct {
		area    string
		lengths []int
	}{{"08", []int{6, 7, 8}}, {"040", []int{5, 6, 7}}, {"031", []int{6, 7}}, {"0480", []int{5, 6}}} {
		t.Run(c.area, func(t *testing.T) {
			got := map[int]bool{}
			eight := map[bool]bool{}
			for i := 0; i < 400; i++ {
				subscriber := fake(t, f, "sv_SE.phone-area["+c.area+"].subscriber")
				digits := len(strings.ReplaceAll(subscriber, " ", ""))
				if re := swedishPhoneGrouped[digits]; re == nil || !re.MatchString(subscriber) {
					t.Fatalf("subscriber %q is not grouped as Språkrådet writes it, or starts with the trunk 0", subscriber)
				}
				got[digits] = true
				if digits == 8 {
					eight[len(strings.Fields(subscriber)) == 3] = true
				}
			}
			for digits := 5; digits <= 8; digits++ {
				if want := slices.Contains(c.lengths, digits); got[digits] != want {
					t.Errorf("subscribers of %d digits drawn is %v, want %v", digits, got[digits], want)
				}
			}
			if got[8] && len(eight) != 2 {
				t.Errorf("eight digits were written one way only, want both of Språkrådet's")
			}
		})
	}
}

func TestShippedSwedishPhoneInternational(t *testing.T) {
	f := shipped(t, WithSeed(5))
	for i := 0; i < 300; i++ {
		got := fakeTemplate(t, f, "{/sv_SE.phone as p}{p}|{p.international}")
		national, international, _ := strings.Cut(got, "|")
		code, subscriber, _ := strings.Cut(strings.TrimPrefix(national, "0"), "-")
		if want := "+46 " + code + " " + subscriber; international != want {
			t.Fatalf("%q: international form %q, want %q", national, international, want)
		}
	}
	for i := 0; i < 20; i++ {
		if got := fake(t, f, "sv_SE.phone.international"); !regexp.MustCompile(`^\+46 \d{1,3} \d`).MatchString(got) {
			t.Fatalf("sv_SE.phone.international = %q", got)
		}
	}
}

func TestShippedSwedishPhoneAreaIsPTSsWholePlan(t *testing.T) {
	b, err := os.ReadFile("data/sv_SE/phone-area.tsv")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")[1:] {
		cells := strings.Split(line, "\t")
		area, ndc, subscriber := cells[0], cells[1], cells[3]
		seen[area] = true
		if area != "0"+ndc {
			t.Errorf("area %s is not the trunk 0 and %s", area, ndc)
		}
		want := fmt.Sprintf("digits-%d-%d", swedishPhoneShortest(ndc)-len(ndc), 9-len(ndc))
		if subscriber != "{.phone-subscriber."+want+"}" {
			t.Errorf("area %s draws %s, PTS allows %s", area, subscriber, want)
		}
	}
	if len(seen) != 264 {
		t.Fatalf("phone-area.tsv has %d area codes, PTS's plan gives 264", len(seen))
	}
	f := shipped(t, WithSeed(2))
	drawn := map[string]bool{}
	for i := 0; i < 6000; i++ {
		drawn[fake(t, f, "sv_SE.phone-area")] = true
	}
	if len(drawn) != 264 {
		t.Fatalf("6000 draws reached %d of the 264 area codes", len(drawn))
	}
	for area, name := range map[string]string{"08": "Stockholm", "0480": "Kalmar"} {
		if got := fake(t, f, "sv_SE.phone-area["+area+"].name"); got != name {
			t.Errorf("area code %s is %q, want %q", area, got, name)
		}
	}
}

func TestShippedSwedishAddress(t *testing.T) {
	f := shipped(t, WithSeed(3))
	digit := regexp.MustCompile(`\d`)
	for i := 0; i < 30; i++ {
		a := fake(t, f, "sv_SE.address")
		if !regexp.MustCompile(`\n`).MatchString(a) || !digit.MatchString(a) {
			t.Fatalf("address %q is not a multi-line address with a number", a)
		}
	}

	locality := regexp.MustCompile(`^\p{L}+([ -]\p{L}+)*$`)
	for i := 0; i < 30; i++ {
		if c := fake(t, f, "sv_SE.address.locality"); !locality.MatchString(c) {
			t.Fatalf("locality %q is not a Swedish place name", c)
		}
	}
}

// TestShippedSwedishSmallPlacesKeepTheirPopulation pins postorter that lie inside an SCB småort
// of 2023, which holds 50 to 199 people.
func TestShippedSwedishSmallPlacesKeepTheirPopulation(t *testing.T) {
	t.Parallel()
	want := map[string]string{"Ambjörby": "127", "Storlien": "89", "Utö": "54", "Målerås": "196", "Latikberg": "51"}
	for _, r := range geoRows(t, "SE", "locality") {
		if people, ok := want[r[0]]; ok {
			if r[2] != people {
				t.Errorf("geo.SE: %s holds %s people, want %s from SCB's småorter 2023", r[0], r[2], people)
			}
			delete(want, r[0])
		}
	}
	for name := range want {
		t.Errorf("geo.SE: no locality %s", name)
	}
}

func TestShippedSwedishPostortTakesTheKommunOfTheSmaortHoldingItsCentre(t *testing.T) {
	t.Parallel()
	want := map[string][2]string{"Lönneberga": {"0860", "129"}, "Axmar": {"2180", "103"}, "Rökå": {"2418", "60"}, "Torsö": {"1083", "57"}}
	for _, r := range geoRows(t, "SE", "locality") {
		if w, ok := want[r[0]]; ok {
			if r[1] != w[0] || r[2] != w[1] {
				t.Errorf("geo.SE: %s lies in kommun %s with %s people, want %s with %s from the SCB småort holding its centre", r[0], r[1], r[2], w[0], w[1])
			}
			delete(want, r[0])
		}
	}
	for name := range want {
		t.Errorf("geo.SE: no locality %s", name)
	}
}

func TestShippedSwedishStreetsAreStreetNames(t *testing.T) {
	refused := map[string]bool{
		"Färjeled Henriksdalsbryggan - Barnängsbryggan": true, "HÄLLSÄTTERSLEDEN": true, "Kärrs VSF": true,
		"Länsväg 694": true, "NN": true, "Risövägen,Dejstolvägen,Runvägen": true, "tornslingan": true,
		"Öresundsvägen, Bondevägen": true,
	}
	for _, r := range geoRows(t, "SE", "street") {
		if refused[r[0]] {
			t.Errorf("geo.SE: street %q in %s, want no road number, ferry route, label, placeholder, two names or miscased name", r[0], r[1])
		}
	}
}

// swedishName matches one or more letter-words, optionally space/hyphen joined
// ("Storgatan", "Norra Promenaden", "von Flemming"). Used by the composition
// tests so shipped name lists can grow without re-enumerating them here.
var swedishName = regexp.MustCompile(`^\p{L}+([ -]\p{L}+)*$`)

func TestShippedStreetIsARegisteredName(t *testing.T) {
	f := shipped(t, WithSeed(5))
	street := regexp.MustCompile(`^\p{L}[\p{L}\d:.-]*([ -][\p{L}\d:.-]+)*$`)
	for i := 0; i < 300; i++ {
		if s := fake(t, f, "sv_SE.address.street"); !street.MatchString(s) {
			t.Fatalf("street %q is not a Swedish street name", s)
		}
	}
}

func TestShippedLastNameComposition(t *testing.T) {
	// last is a choice of patronymic {first}sson templates, compound
	// {first}{last} templates and literal surnames.
	f := shipped(t, WithSeed(6))
	for i := 0; i < 300; i++ {
		if s := fake(t, f, "sv_SE.person.last"); !swedishName.MatchString(s) {
			t.Fatalf("last name %q is not a Swedish surname", s)
		}
	}
}

// The share of each age band is SCB's for Sweden at 2024-12-31, per
// docs/research/research-age-bands.md: 0-17 20.36%, 18-64 58.80%, 65-99 20.81%, 100 or more 0.028%.
// Read on 2026-11-04, the bands are the birth years 2009-2026, 1962-2008, 1927-1961 and 1913-1926.
func TestSwedishBirthDatesWeighAgesByPopulation(t *testing.T) {
	for path, layout := range map[string]string{"sv_SE.personnummer-12": "20060102", "sv_SE.person.birthdate": "2006-01-02"} {
		t.Run(path, func(t *testing.T) {
			const draws = 200000
			sv := shipped(t, WithToday(dayOf(2026, 11, 4)))
			bands := [4]int{}
			oldest, youngest := [4]int{9999, 9999, 9999, 9999}, [4]int{}
			for range draws {
				v := fake(t, sv, path)
				born, err := time.Parse(layout, v[:min(len(v), len(layout))])
				if err != nil {
					t.Fatalf("%s %q holds no calendar date: %v", path, v, err)
				}
				band := 3
				switch age := 2026 - born.Year(); {
				case age < 18:
					band = 0
				case age < 65:
					band = 1
				case age < 100:
					band = 2
				}
				bands[band]++
				oldest[band], youngest[band] = min(oldest[band], born.Year()), max(youngest[band], born.Year())
			}
			for i, want := range []float64{20.36, 58.80, 20.81} {
				if got := 100 * float64(bands[i]) / draws; got < want-0.3 || got > want+0.3 {
					t.Errorf("age band %d drew %.2f%%, want about %.2f%%", i, got, want)
				}
			}
			if bands[3] < 33 || bands[3] > 80 {
				t.Errorf("age 100 or more drew %d in %d, want about 56, 0.028%%", bands[3], draws)
			}
			if want := [4]int{2009, 1962, 1927, 1913}; oldest != want {
				t.Errorf("the bands' oldest birth years = %v, want %v", oldest, want)
			}
			if want := [4]int{2026, 2008, 1961, 1926}; youngest != want {
				t.Errorf("the bands' youngest birth years = %v, want %v", youngest, want)
			}
		})
	}
}

// The law sets the + from the year the holder turns 100. Read on 2026-11-04, that is
// everyone born in 1926 or before, back to 1913, age 113 being the oldest validated Swede's.
func TestSwedishPersonnummerSpellsThePlusFromTheYearTheHolderTurns100(t *testing.T) {
	today := dayOf(2026, 11, 4)
	sv := shipped(t, WithToday(today))
	pluses := 0
	for range 200000 {
		v := fake(t, sv, "sv_SE.personnummer")
		day, err := time.Parse("060102", v[:6])
		if err != nil {
			t.Fatalf("personnummer %q holds no calendar date: %v", v, err)
		}
		yy, century := day.Year()%100, 1900
		if v[6] == '-' {
			if yy <= 26 {
				century = 2000
			}
			if holder := dayOf(century+yy, day.Month(), day.Day()); holder.After(today) || holder.Year() < 1927 {
				t.Fatalf("personnummer %q read on %s, want - for a birth in 1927 or later, and none after the date", v, today.Format(time.DateOnly))
			}
			continue
		}
		pluses++
		if v[6] != '+' || yy < 13 || yy > 26 {
			t.Fatalf("personnummer %q read on %s, want + for a birth from 1913 to 1926", v, today.Format(time.DateOnly))
		}
	}
	if pluses == 0 {
		t.Error("never drew a +")
	}
}

// Both forms draw one date from one set of bands, so a seed gives one person in both spellings.
func TestSwedishPersonnummerFormsSpellOneNumber(t *testing.T) {
	short, long := shipped(t, WithToday(dayOf(2026, 11, 4))), shipped(t, WithToday(dayOf(2026, 11, 4)))
	saw := map[byte]bool{}
	for range 100000 {
		a, b := fake(t, short, "sv_SE.personnummer"), fake(t, long, "sv_SE.personnummer-12")
		if a[:6]+a[7:] != b[2:] {
			t.Fatalf("personnummer %q and personnummer-12 %q, want the short form's digits after the century", a, b)
		}
		if year, _ := strconv.Atoi(b[:4]); (a[6] == '+') != (year <= 1926) {
			t.Fatalf("personnummer %q and personnummer-12 %q, want + exactly for a birth in 1926 or before", a, b)
		}
		saw[a[6]] = true
	}
	if !saw['+'] || !saw['-'] {
		t.Errorf("separators drawn %v, want both", saw)
	}
}

// The 12-digit form is the century and the ten digits of the short form without its separator, and
// the Luhn check stays over those ten digits, since weighing the century too would change it.
func TestSwedishPersonnummerTwelveDigits(t *testing.T) {
	today := dayOf(2026, 11, 4)
	sv := shipped(t, WithToday(today))
	re := regexp.MustCompile(`^(19|20)\d{6}23[89]\d$`)
	first, last := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC), time.Time{}
	for i := 0; i < 20000; i++ {
		v := fake(t, sv, "sv_SE.personnummer-12")
		if !re.MatchString(v) {
			t.Fatalf("personnummer-12 %q, want YYYYMMDD23[89]C", v)
		}
		born, err := time.Parse("20060102", v[:8])
		if err != nil || born.After(today) {
			t.Fatalf("personnummer-12 %q holds no calendar date up to %s: %v", v, today.Format(time.DateOnly), err)
		}
		if !luhnValid(v[2:]) {
			t.Fatalf("personnummer-12 %q fails the Luhn check over its last ten digits", v)
		}
		if born.Before(first) {
			first = born
		}
		if born.After(last) {
			last = born
		}
	}
	if first.Year() > 1927 || last.Year() != 2026 {
		t.Errorf("personnummer-12 drew %s to %s in 20000, want it to reach the date's year and back towards 1926", first.Format(time.DateOnly), last.Format(time.DateOnly))
	}
}

func TestSwedishSamordningsnummerTwelveDigits(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	re := regexp.MustCompile(`^(19|20)\d{6}23[89]\d$`)
	centuries := map[string]int{}
	for i := 0; i < 3000; i++ {
		v := fake(t, sv, "sv_SE.samordningsnummer-12")
		if !re.MatchString(v) {
			t.Fatalf("samordningsnummer-12 %q, want YYYYMM(61-88)23[89]C", v)
		}
		day, err := strconv.Atoi(v[6:8])
		if err != nil || day < 61 || day > 88 {
			t.Fatalf("samordningsnummer-12 %q, want a day of 61 to 88", v)
		}
		if _, err := time.Parse("200601", v[:6]); err != nil {
			t.Fatalf("samordningsnummer-12 %q is not a valid year and month: %v", v, err)
		}
		if !luhnValid(v[2:]) {
			t.Fatalf("samordningsnummer-12 %q fails the Luhn check over its last ten digits", v)
		}
		if year, _ := strconv.Atoi(v[:4]); year < 1930 || year > 2025 {
			t.Fatalf("samordningsnummer-12 %q, want a year from 1930 to 2025, whichever century its two digits sit in", v)
		}
		centuries[v[:2]]++
	}
	if centuries["19"] < 1500 || centuries["20"] < 400 {
		t.Errorf("centuries drawn %v in 3000, want 1930 to 1999 and 2000 to 2025 in proportion to their days", centuries)
	}
}

// The 12-digit form holds the short form's range of birth years in its own file, so each
// draws the same first and last year.
func TestSwedishSamordningsnummerFormsSpanOneRangeOfYears(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	span := func(path string, year func(string) int) [2]int {
		first, last := 9999, 0
		for i := 0; i < 20000; i++ {
			y := year(fake(t, sv, path))
			first, last = min(first, y), max(last, y)
		}
		return [2]int{first, last}
	}
	short := func(v string) int {
		yy, _ := strconv.Atoi(v[:2])
		if yy >= 30 {
			return 1900 + yy
		}
		return 2000 + yy
	}
	long := func(v string) int { y, _ := strconv.Atoi(v[:4]); return y }
	if a, b := span("sv_SE.samordningsnummer", short), span("sv_SE.samordningsnummer-12", long); a != b || a[0] > 1931 || a[1] < 2024 {
		t.Errorf("samordningsnummer draws years %v, samordningsnummer-12 %v, want the same range, 1930 to 2025 at its ends", a, b)
	}
}

func TestShippedSwedishCalendar(t *testing.T) {
	f := shipped(t, WithSeed(3))
	for path, want := range map[string]string{
		"sv_SE.month[1]":                        "januari",
		"sv_SE.month[2]":                        "februari",
		"sv_SE.month[3]":                        "mars",
		"sv_SE.month[4]":                        "april",
		"sv_SE.month[5]":                        "maj",
		"sv_SE.month[6]":                        "juni",
		"sv_SE.month[7]":                        "juli",
		"sv_SE.month[8]":                        "augusti",
		"sv_SE.month[9]":                        "september",
		"sv_SE.month[10]":                       "oktober",
		"sv_SE.month[11]":                       "november",
		"sv_SE.month[12]":                       "december",
		"sv_SE.weekday[1]":                      "måndag",
		"sv_SE.weekday[2]":                      "tisdag",
		"sv_SE.weekday[3]":                      "onsdag",
		"sv_SE.weekday[4]":                      "torsdag",
		"sv_SE.weekday[5]":                      "fredag",
		"sv_SE.weekday[6]":                      "lördag",
		"sv_SE.weekday[7]":                      "söndag",
		"sv_SE.holiday[nyarsdagen]":             "nyårsdagen",
		"sv_SE.holiday[trettondedag-jul]":       "trettondedag jul",
		"sv_SE.holiday[langfredagen]":           "långfredagen",
		"sv_SE.holiday[paskdagen]":              "påskdagen",
		"sv_SE.holiday[annandag-pask]":          "annandag påsk",
		"sv_SE.holiday[forsta-maj]":             "första maj",
		"sv_SE.holiday[kristi-himmelsfardsdag]": "Kristi himmelsfärdsdag",
		"sv_SE.holiday[pingstdagen]":            "pingstdagen",
		"sv_SE.holiday[nationaldagen]":          "nationaldagen",
		"sv_SE.holiday[midsommardagen]":         "midsommardagen",
		"sv_SE.holiday[alla-helgons-dag]":       "alla helgons dag",
		"sv_SE.holiday[juldagen]":               "juldagen",
		"sv_SE.holiday[annandag-jul]":           "annandag jul",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	for path, want := range map[string][]string{
		"sv_SE.month":   {"januari", "februari", "mars", "april", "maj", "juni", "juli", "augusti", "september", "oktober", "november", "december"},
		"sv_SE.weekday": {"måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag", "söndag"},
		"sv_SE.holiday": {"nyårsdagen", "trettondedag jul", "långfredagen", "påskdagen", "annandag påsk", "första maj", "Kristi himmelsfärdsdag", "pingstdagen", "nationaldagen", "midsommardagen", "alla helgons dag", "juldagen", "annandag jul"},
	} {
		seen := map[string]bool{}
		for i := 0; i < 2000; i++ {
			seen[fake(t, f, path)] = true
		}
		got := make([]string, 0, len(seen))
		for name := range seen {
			got = append(got, name)
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%s drew %v in 2000 draws, want exactly %v", path, got, want)
		}
	}
}

func TestShippedSwedishTerritoryNamesEveryMiscTerritory(t *testing.T) {
	keys := func(file string) []string {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[1:] {
			key, _, _ := strings.Cut(line, "\t")
			out = append(out, key)
		}
		slices.Sort(out)
		return out
	}
	if got, want := keys("data/sv_SE/territory.tsv"), keys("data/misc/territory.tsv"); !slices.Equal(got, want) {
		t.Errorf("sv_SE/territory.tsv has %d rows, misc/territory.tsv %d: want the same alpha-2 codes", len(got), len(want))
	}
	f := shipped(t, WithSeed(3))
	for path, want := range map[string]string{
		"sv_SE.territory[SE]": "Sverige",
		"sv_SE.territory[US]": "USA",
		"sv_SE.territory[DE]": "Tyskland",
		"sv_SE.territory[GB]": "Storbritannien",
		"sv_SE.territory[GL]": "Grönland",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestShippedSwedishBloodTypeFollowsItsShares(t *testing.T) {
	f := shipped(t, WithSeed(11))
	const draws = 40000
	counts := map[string]int{}
	for i := 0; i < draws; i++ {
		counts[fake(t, f, "sv_SE.bloodtype")]++
	}
	want := map[string]float64{"A+": 37, "0+": 32, "B+": 10, "A-": 7, "0-": 6, "AB+": 5, "B-": 2, "AB-": 1}
	if len(counts) != len(want) {
		t.Errorf("drew %v, want exactly the types %v", counts, want)
	}
	for name, share := range want {
		if got := float64(counts[name]) * 100 / draws; math.Abs(got-share) > 0.5 {
			t.Errorf("%s drew %.1f%%, want %v%% within half a point", name, got, share)
		}
	}
}

// A personnummer spells + from the year the holder turns 100: with today at 2026-11-04, a birth in 1926 or before.
func TestSwedishPersonNumbersSpellItsBirthdateAndSex(t *testing.T) {
	sv := shipped(t, WithToday(dayOf(2026, 11, 4)))
	person, err := sv.NewTemplate(`{/sv_SE.person as p}{p.birthdate}|{p.personnummer}|{p.samordningsnummer}|{p.sex}`)
	if err != nil {
		t.Fatal(err)
	}
	birthNumber := map[string]string{"kvinna": "238", "man": "239"}
	pluses, days := 0, map[int]bool{}
	for range 200000 {
		v := person.Fake()
		f := strings.Split(v, "|")
		born, err := time.Parse(time.DateOnly, f[0])
		if err != nil || born.Year() < 1913 || born.After(dayOf(2026, 11, 4)) {
			t.Fatalf("person %q, want a birthdate from 1913-01-01 to 2026-11-04", v)
		}
		sign := "-"
		if born.Year() <= 1926 {
			sign, pluses = "+", pluses+1
		}
		pnr := born.Format("060102") + sign + birthNumber[f[3]]
		if len(f[1]) != 11 || f[1][:10] != pnr || !luhnValid(digitsOnly(f[1])) {
			t.Fatalf("person %q, want a Luhn-valid personnummer starting %s", v, pnr)
		}
		sam := born.Format("0601") + strconv.Itoa(born.Day()+60) + "-" + birthNumber[f[3]]
		if len(f[2]) != 11 || f[2][:10] != sam || !luhnValid(digitsOnly(f[2])) {
			t.Fatalf("person %q, want a Luhn-valid samordningsnummer starting %s", v, sam)
		}
		days[born.Day()+60] = true
	}
	if pluses == 0 {
		t.Error("never drew a +")
	}
	for day := 61; day <= 91; day++ {
		if !days[day] {
			t.Errorf("samordningsnummer never drew the day %d", day)
		}
	}
}

// sv_SE.sni is SCB's SNI 2025 detaljgrupper: 835 five-digit codes, each selected by its code and
// drawn as the code and its Swedish label.
func TestShippedSwedishSNI(t *testing.T) {
	f := shipped(t, WithSeed(3))
	for path, want := range map[string]string{
		"sv_SE.sni[01.110]":      "01.110 Odling av spannmål (utom ris), baljväxter och oljeväxter",
		"sv_SE.sni[62.100].name": "Dataprogrammering",
		"sv_SE.sni[17.110].name": "Massatillverkning",
		"sv_SE.sni[99.000].code": "99.000",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	rows := tsvRows(t, "data/sv_SE/sni.tsv")
	if len(rows) != 835 {
		t.Fatalf("sni.tsv has %d rows, want 835", len(rows))
	}
	code := regexp.MustCompile(`^\d\d\.\d{3}$`)
	for _, row := range rows {
		if !code.MatchString(row["code"]) || row["name"] != strings.TrimSpace(row["name"]) || row["name"] == "" {
			t.Errorf("sni.tsv row %q is not an NN.NNN code with a trimmed label", row)
		}
	}
}

// sv_SE.ssyk is SCB's SSYK 2012 at four digits: 429 occupation groups, each selected by its code.
func TestShippedSwedishSSYK(t *testing.T) {
	f := shipped(t, WithSeed(3))
	for path, want := range map[string]string{
		"sv_SE.ssyk[2512]":      "2512 Mjukvaru- och systemutvecklare m.fl.",
		"sv_SE.ssyk[1111].name": "Politiker",
		"sv_SE.ssyk[2221].name": "Grundutbildade sjuksköterskor",
		"sv_SE.ssyk[5223].name": "Butikssäljare, fackhandel",
		"sv_SE.ssyk[5120].code": "5120",
	} {
		if got := fake(t, f, path); got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	rows := tsvRows(t, "data/sv_SE/ssyk.tsv")
	if len(rows) != 429 {
		t.Fatalf("ssyk.tsv has %d rows, want 429", len(rows))
	}
	code := regexp.MustCompile(`^\d{4}$`)
	for _, row := range rows {
		if !code.MatchString(row["code"]) || row["name"] != strings.TrimSpace(row["name"]) || row["name"] == "" {
			t.Errorf("ssyk.tsv row %q is not a four-digit code with a trimmed label", row)
		}
	}
}

// sv_SE.car holds each model that Mobility Sweden's December 2025 report counts 500 or more new
// passenger cars of in 2025, less its "Övriga" rows, weighted by those registrations.
func TestShippedSwedishCar(t *testing.T) {
	rows := tsvRows(t, "data/sv_SE/car.tsv")
	if len(rows) != 119 {
		t.Fatalf("car.tsv has %d rows, want 119", len(rows))
	}
	total, registrations := 0, map[string]int{}
	for _, row := range rows {
		n, err := strconv.Atoi(row["registrations"])
		if err != nil || n < 500 {
			t.Errorf("car.tsv row %q has no registrations of 500 or more", row)
		}
		if strings.HasPrefix(strings.ToUpper(row["model"]), strings.ToUpper(row["make"])+" ") || strings.Contains(strings.ToUpper(row["model"]), "ÖVRIGA") {
			t.Errorf("car.tsv row %q repeats its make in the model or is an Övriga row", row)
		}
		total += n
		registrations[row["make"]+" "+row["model"]] = n
	}
	for name, want := range map[string]int{
		"Volvo XC60":                  17933,
		"Volkswagen ID.7/ID.7 Tourer": 8134,
		"Tesla Model Y":               5822,
		"Mercedes-Benz E-Klass":       2596,
		"Lynk & Co 01":                757,
		"Renault 5":                   1371,
		"Toyota RAV4":                 5091,
		"Porsche 911":                 575,
		"Toyota bZ4X":                 746,
		"BMW iX1":                     1795,
		"Audi Q4 e-tron":              2042,
		"Hyundai i10":                 578,
		"Volkswagen T-Roc":            3669,
		"Volvo XC90":                  2078,
		"Škoda Octavia":               2219,
		"Citroën C4":                  536,
	} {
		if registrations[name] != want {
			t.Errorf("car.tsv gives %s %d registrations, want %d", name, registrations[name], want)
		}
	}
	if total != 242981 {
		t.Errorf("car.tsv registrations sum to %d, want 242981", total)
	}
	f := shipped(t, WithSeed(7))
	const draws = 60000
	xc60 := 0
	for i := 0; i < draws; i++ {
		if fake(t, f, "sv_SE.car") == "Volvo XC60" {
			xc60++
		}
	}
	want := 17933 * 100 / float64(total)
	if got := float64(xc60) * 100 / draws; math.Abs(got-want) > 0.5 {
		t.Errorf("Volvo XC60 drew %.2f%%, want %.2f%% within half a point", got, want)
	}
}

// A company's format reads its wording through a name, so a read of the wording beside it would be a
// second company, and the load says how to write each meaning.
func TestACompanysRegisteredBesideTheCompanyIsRefused(t *testing.T) {
	t.Parallel()
	f := shipped(t, WithSeed(3))
	for _, locale := range []string{"sv_SE", "nb_NO"} {
		company := "/" + locale + ".company"
		_, err := f.NewTemplate("{" + company + " as c}{c} / {c.registered}")
		if err == nil {
			t.Fatalf("%s: {c.registered} beside {c} loaded, want it refused", locale)
		}
		for _, want := range []string{"{c.registered}", "{" + company + ".registered}"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error %q, want it to name %s", locale, err, want)
			}
		}
		for _, loads := range []string{"{" + company + " as c}{c}|{c.legal}", "{" + company + " as c}{c.registered}", "{" + company + ".registered}"} {
			if _, err := f.NewTemplate(loads); err != nil {
				t.Errorf("NewTemplate(%s) = %v, want it to load", loads, err)
			}
		}
	}
}

// The wordings are those TestSwedishCompanyWordingsAndShapesFollowTheRegister holds a company to, and an
// enskild näringsidkare carries none.
func TestSwedishCompanyLegalIsTheWordOfTheBoundCompany(t *testing.T) {
	t.Parallel()
	sv := shipped(t, WithSeed(3))
	after := []string{"AB", "Aktiebolag", "Handelsbolag", "handelsbolag", "Kommanditbolag", "kommanditbolag", "Ekonomisk förening", "ekonomisk förening", "ek. för.", "Ek. för."}
	before := []string{"Handelsbolaget", "Kommanditbolaget", "Bostadsrättsföreningen"}
	seen := map[string]int{}
	const draws = shareDraws / 10
	for i := 0; i < draws; i++ {
		company, legal, _ := strings.Cut(fakeTemplate(t, sv, "{/sv_SE.company as c}{c}|{c.legal}"), "|")
		seen[legal]++
		switch {
		case legal == "":
			for _, w := range after {
				if strings.HasSuffix(company, " "+w) {
					t.Fatalf("company %q has an empty legal, yet ends in %q", company, w)
				}
			}
			for _, w := range before {
				if strings.HasPrefix(company, w+" ") {
					t.Fatalf("company %q has an empty legal, yet starts with %q", company, w)
				}
			}
		case slices.Contains(before, legal):
			if !strings.HasPrefix(company, legal+" ") {
				t.Fatalf("company %q does not start with its legal %q", company, legal)
			}
		case slices.Contains(after, legal):
			if !strings.HasSuffix(company, " "+legal) {
				t.Fatalf("company %q does not end in its legal %q", company, legal)
			}
		default:
			t.Fatalf("legal %q of company %q is no wording of a legal form", legal, company)
		}
	}
	assertShare(t, "companies with no legal word", seen[""], draws, 8468.0/61777)
	raw, err := os.ReadFile("data/sv_SE/company.json")
	if err != nil {
		t.Fatal(err)
	}
	var company struct {
		Registered []struct{ Format, Legal string } `json:"registered"`
	}
	if err := json.Unmarshal(raw, &company); err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, item := range company.Registered {
		listed[item.Legal] = true
		if item.Legal != "" && !strings.HasPrefix(item.Format, item.Legal+" ") && !strings.HasSuffix(item.Format, " "+item.Legal) {
			t.Errorf("%q is listed with legal %q, which it does not start or end with", item.Format, item.Legal)
		}
	}
	for _, w := range append(after, before...) {
		if !listed[w] {
			t.Errorf("no registered company form carries the legal word %q", w)
		}
	}
}

func swedishWordForms(t *testing.T, table string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("data/sv_SE/" + table + ".tsv")
	if err != nil {
		t.Fatal(err)
	}
	forms := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n")[1:] {
		word, form, _ := strings.Cut(line, "\t")
		forms[word] = form
	}
	return forms
}

func TestSwedishWordHoldsSaldoFormsPerPartOfSpeech(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ table, word, form string }{
		{"word-noun", "hus", "huset"},
		{"word-noun", "flicka", "flickan"},
		{"word-verb", "springa", "springer"},
		{"word-adj", "stor", "stora"},
		{"word-adv", "ofta", ""},
	} {
		if got, ok := swedishWordForms(t, c.table)[c.word]; !ok || got != c.form {
			t.Errorf("%s holds %q with form %q (held: %v), want %q", c.table, c.word, got, ok, c.form)
		}
	}
	if n := len(swedishWordForms(t, "word-noun")); n < 50000 {
		t.Errorf("word-noun holds %d nouns, want SALDO's, tens of thousands", n)
	}
}

func TestSwedishWordPathsDrawTheirPartOfSpeech(t *testing.T) {
	t.Parallel()
	sv := shipped(t, WithSeed(5))
	nouns, verbs, adjs, advs := swedishWordForms(t, "word-noun"), swedishWordForms(t, "word-verb"), swedishWordForms(t, "word-adj"), swedishWordForms(t, "word-adv")
	definite := map[string]bool{}
	for _, form := range nouns {
		definite[form] = true
	}
	present := map[string]bool{}
	for _, form := range verbs {
		present[form] = true
	}
	seen := map[string]int{}
	for i := 0; i < 2000; i++ {
		for path, held := range map[string]func(string) bool{
			"sv_SE.word.noun":          func(w string) bool { _, ok := nouns[w]; return ok },
			"sv_SE.word.noun.definite": func(w string) bool { return definite[w] },
			"sv_SE.word.verb":          func(w string) bool { _, ok := verbs[w]; return ok },
			"sv_SE.word.verb.present":  func(w string) bool { return present[w] },
			"sv_SE.word.adj":           func(w string) bool { _, ok := adjs[w]; return ok },
			"sv_SE.word.adv":           func(w string) bool { _, ok := advs[w]; return ok },
		} {
			if got := fake(t, sv, path); !held(got) {
				t.Fatalf("%s = %q, which its table does not hold", path, got)
			}
		}
		w := fake(t, sv, "sv_SE.word")
		switch _, n := nouns[w]; {
		case n:
			seen["noun"]++
		default:
			seen["other"]++
		}
	}
	if seen["noun"] == 0 || seen["other"] == 0 {
		t.Errorf("sv_SE.word drew %v, want nouns and other parts of speech", seen)
	}
}

func TestSwedishBoundWordReadsItsOwnForm(t *testing.T) {
	t.Parallel()
	sv := shipped(t, WithSeed(6))
	for _, c := range []struct{ path, form, table string }{
		{"sv_SE.word.noun", "definite", "word-noun"},
		{"sv_SE.word.verb", "present", "word-verb"},
		{"sv_SE.word.adj", "definite", "word-adj"},
	} {
		forms := swedishWordForms(t, c.table)
		for i := 0; i < 200; i++ {
			word, form, _ := strings.Cut(fakeTemplate(t, sv, "{/"+c.path+" as w}{w}|{w."+c.form+"}"), "|")
			if forms[word] != form {
				t.Fatalf("%s %q printed %s %q, want its own, %q", c.path, word, c.form, form, forms[word])
			}
		}
	}
}

func TestSwedishWordWeighsAPartOfSpeechByItsRows(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("data/sv_SE/word.json")
	if err != nil {
		t.Fatal(err)
	}
	var word struct {
		Any []struct {
			Format string `json:"format"`
			Weight int    `json:"weight"`
		} `json:"any"`
	}
	if err := json.Unmarshal(raw, &word); err != nil {
		t.Fatal(err)
	}
	if len(word.Any) != 4 {
		t.Fatalf("any holds %d parts of speech, want 4", len(word.Any))
	}
	for _, part := range word.Any {
		table := strings.Trim(part.Format, "{.}")
		if rows := len(swedishWordForms(t, table)); part.Weight != rows {
			t.Errorf("%s weighs %d, want its %d rows", table, part.Weight, rows)
		}
	}
}

func TestSwedishSentenceDrawsTheFormsItsPatternNeeds(t *testing.T) {
	t.Parallel()
	sv := shipped(t, WithSeed(7))
	forms := func(table string) map[string]bool {
		held := map[string]bool{}
		for _, form := range swedishWordForms(t, table) {
			held[form] = true
		}
		return held
	}
	adj, noun, present := forms("word-adj"), forms("word-noun"), forms("word-verb")
	lemma := swedishWordForms(t, "word-verb")
	for i := 0; i < 200; i++ {
		sentence := fake(t, sv, "sv_SE.sentence")
		w := strings.Fields(strings.TrimRight(sentence, ".!?"))
		last := len(w) - 1
		var adjs, nouns []string
		var verbHeld bool
		if strings.HasSuffix(sentence, "?") {
			adjs, nouns = []string{w[last-4]}, []string{w[last-3], w[last]}
			_, verbHeld = lemma[w[last-2]]
		} else {
			adjs, nouns = []string{w[1], w[5]}, []string{w[2], w[6]}
			verbHeld = present[w[3]]
		}
		if !verbHeld {
			t.Errorf("sentence %q holds a verb form its pattern doesn't read", sentence)
		}
		for _, n := range nouns {
			if !noun[n] {
				t.Errorf("sentence %q prints %q, no definite noun", sentence, n)
			}
		}
		for _, a := range adjs {
			if !adj[a] {
				t.Errorf("sentence %q prints %q, no definite adjective", sentence, a)
			}
		}
	}
}

func TestSwedishPostortKeepsItsDeliveryDigitNineCodes(t *testing.T) {
	t.Parallel()
	shipped := map[string]string{}
	for _, r := range geoRows(t, "SE", "postal-code") {
		shipped[r[0]] = r[1]
	}
	for code, locality := range map[string]string{"169 41": "Solna", "169 34": "Solna", "589 41": "Linköping"} {
		if shipped[code] != locality {
			t.Errorf("%s belongs to %q, want %s, where people and companies have street addresses", code, shipped[code], locality)
		}
	}
}

func TestSwedishStreetSitsInThePostortWhoseAreaHoldsIt(t *testing.T) {
	t.Parallel()
	streets := map[string]map[string]bool{}
	for _, r := range geoRows(t, "SE", "street") {
		if streets[r[1]] == nil {
			streets[r[1]] = map[string]bool{}
		}
		streets[r[1]][r[0]] = true
	}
	for locality, names := range map[string][]string{
		"Stockholm": {"Sveavägen", "Birger Jarlsgatan", "Hornsgatan", "Götgatan", "Odengatan", "Fleminggatan"},
		"Bromma":    {"Bergslagsvägen"},
	} {
		for _, name := range names {
			if !streets[locality][name] {
				t.Errorf("%s has no %s", locality, name)
			}
		}
	}
	if got := len(streets["Stockholm"]); got < 1000 {
		t.Errorf("Stockholm has %d streets, want the 1000 or more of a capital's own postort", got)
	}
}

func bankgirotMod11Valid(digits string) bool {
	sum := 0
	for i := range digits {
		sum += int(digits[i]-'0') * ((len(digits)-1-i)%10 + 1)
	}
	return sum%11 == 0
}

// TestSwedishBankgiro draws 991-2346 and 5555-5551's two forms: seven digits with the hyphen before
// the last four, eight with it in the middle, the final digit Luhn's check digit as Bankgirot's
// 10-modul sheet has it.
func TestSwedishBankgiro(t *testing.T) {
	if !luhnValid("9912346") || !luhnValid("55555551") {
		t.Fatal("Bankgirot's two examples, 991-2346 and 5555-5551, must pass the Luhn oracle")
	}
	f := shipped(t, WithSeed(1))
	form := regexp.MustCompile(`^([1-9]\d{2}-\d{4}|[1-9]\d{3}-\d{4})$`)
	seven, eight := 0, 0
	for i := 0; i < 2000; i++ {
		v := fake(t, f, "sv_SE.bankgiro")
		if !form.MatchString(v) || !luhnValid(digitsOnly(v)) {
			t.Fatalf("sv_SE.bankgiro %q, want 991-2346 or 5555-5551's form with a Luhn check digit", v)
		}
		if len(v) == 8 {
			seven++
		} else {
			eight++
		}
	}
	if seven < 800 || eight < 800 {
		t.Errorf("%d bankgiro of seven digits and %d of eight in 2000, want about half each", seven, eight)
	}
}

// TestSwedishPlusgiro draws two to eight digits, the hyphen before the check digit and the rest in
// pairs from the right, with a Luhn check digit.
func TestSwedishPlusgiro(t *testing.T) {
	f := shipped(t, WithSeed(1))
	form := regexp.MustCompile(`^[1-9]\d?( \d{2})*-\d$`)
	lengths := map[int]int{}
	for i := 0; i < 3500; i++ {
		v := fake(t, f, "sv_SE.plusgiro")
		d := digitsOnly(v)
		if !form.MatchString(v) || !luhnValid(d) || len(d) < 2 || len(d) > 8 {
			t.Fatalf("sv_SE.plusgiro %q, want 2 to 8 digits, the hyphen before the last, a Luhn check digit", v)
		}
		if strings.Index(v, "-") != len(v)-2 {
			t.Fatalf("sv_SE.plusgiro %q has its hyphen elsewhere than before the last digit", v)
		}
		lengths[len(d)]++
	}
	for n := 2; n <= 8; n++ {
		if lengths[n] < 300 {
			t.Errorf("%d plusgiro of %d digits in 3500, want about 500", lengths[n], n)
		}
	}
}

// clearingRanges are Bankinfrastruktur i Sverige AB's clearing ranges, each with the bank's name,
// the account type and the comment of Bankgirot's "Bankernas kontonummer" of 2024-02-22, which the
// list's column Konto-typ gives as type:comment. The ranges 9565-9569 and 9600-9609 are left out:
// that document does not publish their comment 4.
var clearingRanges = []struct {
	first, last int
	bank        string
	typ, rule   int
}{
	{1100, 1199, "Nordea", 1, 1},
	{1200, 1399, "Danske Bank", 1, 1},
	{1400, 2099, "Nordea", 1, 1},
	{2110, 2119, "Juni Technology AB", 1, 1},
	{2120, 2129, "Steven AB", 1, 1},
	{2130, 2139, "Zimpler AB", 1, 1},
	{2300, 2399, "Ålandsbanken", 1, 2},
	{2400, 2499, "Danske Bank", 1, 1},
	{3000, 3299, "Nordea", 1, 1},
	{3300, 3300, "Nordea", 2, 1},
	{3301, 3399, "Nordea", 1, 1},
	{3400, 3409, "Länsförsäkringar Bank", 1, 1},
	{3410, 3781, "Nordea", 1, 1},
	{3782, 3782, "Nordea", 2, 1},
	{3783, 3999, "Nordea", 1, 1},
	{4000, 4999, "Nordea", 1, 2},
	{5000, 5999, "SEB", 1, 1},
	{6000, 6999, "Handelsbanken", 2, 2},
	{7000, 7999, "Swedbank", 1, 1},
	{8000, 8999, "Swedbank", 2, 3},
	{9020, 9029, "Länsförsäkringar Bank", 1, 2},
	{9040, 9049, "Citibank (filial)", 1, 2},
	{9060, 9069, "Länsförsäkringar Bank", 1, 1},
	{9070, 9079, "Multitude Bank plc.", 1, 1},
	{9100, 9109, "Nordnet Bank", 1, 2},
	{9120, 9124, "SEB", 1, 1},
	{9130, 9149, "SEB", 1, 1},
	{9150, 9169, "Skandiabanken", 1, 2},
	{9170, 9179, "Ikanobanken", 1, 1},
	{9180, 9189, "Danske Bank", 2, 1},
	{9190, 9199, "DNB Bank ASA, filial Sverige", 1, 2},
	{9230, 9239, "Marginalen Bank", 1, 1},
	{9250, 9259, "SBAB Bank", 1, 1},
	{9270, 9279, "ICA Banken", 1, 1},
	{9280, 9289, "Resurs Bank AB", 1, 1},
	{9300, 9349, "Swedbank", 2, 1},
	{9390, 9399, "Landshypotek", 1, 2},
	{9400, 9449, "Forex Bank", 1, 1},
	{9460, 9469, "Santander Consumer Bank", 1, 1},
	{9470, 9479, "BNP Paribas", 1, 2},
	{9500, 9549, "Nordea (Plusgirot)", 2, 3},
	{9550, 9564, "Avanza Bank", 1, 2},
	{9570, 9579, "Sparbanken Syd", 2, 1},
	{9580, 9589, "Aion Bank", 1, 1},
	{9590, 9599, "EP Bank", 1, 2},
	{9630, 9639, "Lån & Spar Bank A/S, filial", 1, 1},
	{9640, 9649, "NOBA Bank Group", 1, 2},
	{9650, 9659, "MedMera Bank AB", 1, 1},
	{9660, 9669, "Svea Bank", 1, 2},
	{9670, 9679, "JAK Medlemsbank", 1, 2},
	{9680, 9689, "Enity Bank Group", 1, 1},
	{9700, 9709, "Ekobanken", 1, 2},
	{9710, 9719, "Lunar Bank", 1, 2},
	{9750, 9759, "Northmill AB", 1, 2},
	{9780, 9789, "Klarna Bank", 1, 2},
	{9960, 9969, "Nordea (Plusgirot)", 2, 3},
}

func bankAccountParts(t *testing.T, v string) (clearing int, account string) {
	t.Helper()
	m := regexp.MustCompile(`^(\d{4})-(\d+)$`).FindStringSubmatch(v)
	if m == nil {
		t.Fatalf("sv_SE.bankaccount %q, want a four-digit clearing number, a hyphen and an account number", v)
	}
	clearing, _ = strconv.Atoi(m[1])
	return clearing, m[2]
}

// checkBankAccount applies the check Bankgirot gives the clearing number's range.
func checkBankAccount(t *testing.T, v string) {
	t.Helper()
	clearing, account := bankAccountParts(t, v)
	for _, r := range clearingRanges {
		if clearing < r.first || clearing > r.last {
			continue
		}
		c := fmt.Sprintf("%04d", clearing)
		ok := false
		switch {
		case r.typ == 1 && r.rule == 1:
			ok = len(account) == 7 && bankgirotMod11Valid(c[1:]+account)
		case r.typ == 1 && r.rule == 2:
			ok = len(account) == 7 && bankgirotMod11Valid(c+account)
		case r.typ == 2 && r.rule == 2:
			ok = len(account) == 9 && bankgirotMod11Valid(account)
		default:
			ok = len(account) == 10 && luhnValid(account)
		}
		if !ok {
			t.Errorf("sv_SE.bankaccount %q is not valid under %s's type %d, comment %d", v, r.bank, r.typ, r.rule)
		}
		return
	}
	t.Errorf("sv_SE.bankaccount %q has a clearing number that no listed range holds", v)
}

// Bankgirot's own example of a number correct under the 11-modul is 1912763608957.
func TestSwedishBankAccountOracleAcceptsBankgirotsExample(t *testing.T) {
	if !bankgirotMod11Valid("1912763608957") || bankgirotMod11Valid("1912763608956") {
		t.Fatal("the 11-modul oracle must accept Bankgirot's 1912763608957 and refuse a changed check digit")
	}
}

func TestSwedishBankAccountHasOneRowPerClearingNumber(t *testing.T) {
	b, err := os.ReadFile("data/sv_SE/bankaccount.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if lines[0] != "clearing\tbank\tnumber" {
		t.Fatalf("header %q, want clearing, bank and number", lines[0])
	}
	var want []string
	for _, r := range clearingRanges {
		for c := r.first; c <= r.last; c++ {
			want = append(want, fmt.Sprintf("%d\t%s", c, r.bank))
		}
	}
	if len(lines)-1 != len(want) {
		t.Fatalf("%d rows, want %d: one for each clearing number of the listed ranges", len(lines)-1, len(want))
	}
	for i, l := range lines[1:] {
		if !strings.HasPrefix(l, want[i]+"\t") {
			t.Fatalf("row %d is %q, want it to open with %q", i+1, l, want[i])
		}
	}
}

func TestSwedishBankAccountOfEveryClearingNumberChecks(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for _, r := range clearingRanges {
		for c := r.first; c <= r.last; c++ {
			v := fakeTemplate(t, f, fmt.Sprintf("{/sv_SE.bankaccount[%d]}", c))
			if !strings.HasPrefix(v, strconv.Itoa(c)+"-") {
				t.Fatalf("sv_SE.bankaccount[%d] = %q, want it to open with its clearing number", c, v)
			}
			checkBankAccount(t, v)
			if bank := fakeTemplate(t, f, fmt.Sprintf("{/sv_SE.bankaccount[%d].bank}", c)); bank != r.bank {
				t.Fatalf("sv_SE.bankaccount[%d].bank = %q, want %q", c, bank, r.bank)
			}
		}
	}
}

func TestSwedishBankAccount(t *testing.T) {
	f := shipped(t, WithSeed(2))
	banks := map[string]bool{}
	for i := 0; i < 3000; i++ {
		bank, v, _ := strings.Cut(fakeTemplate(t, f, "{/sv_SE.bankaccount as a}{a.bank}={a}"), "=")
		checkBankAccount(t, v)
		banks[bank] = true
	}
	for _, b := range []string{"Nordea", "SEB", "Swedbank", "Handelsbanken", "Danske Bank"} {
		if !banks[b] {
			t.Errorf("3000 sv_SE.bankaccount never drew a clearing number of %s", b)
		}
	}
}
