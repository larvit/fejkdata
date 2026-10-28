package fejkdata

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestShippedDataCategories asserts every shipped locale emits only well-formed
// values for each data category. en and sv patterns differ where the format is
// locale-specific (date, time, company, price, ...); the rest are shared.
func TestShippedDataCategories(t *testing.T) {
	letters := regexp.MustCompile(`^[\pL'-]+$`)
	semver := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-(alpha|beta|rc)\.\d+)?$`)
	ip := regexp.MustCompile(`^((\d{1,3}\.){3}\d{1,3}|([0-9a-f]{4}:){7}[0-9a-f]{4})$`)
	color := regexp.MustCompile(`^(#[0-9a-f]{6}|[\pL ]+)$`)
	url := regexp.MustCompile(`^https?://([a-z0-9-]+\.)+[a-z]{2,}(/[a-z0-9./-]*)?$`)
	email := regexp.MustCompile(`^[a-z0-9._-]+@([a-z0-9-]+\.)+[a-z]{2,}$`)
	uname := regexp.MustCompile(`^[a-z][a-z0-9._]*$`)

	cases := []struct {
		path   string
		en, sv *regexp.Regexp
	}{
		{"word", letters, letters},
		{"sentence",
			regexp.MustCompile(`^[A-Z].* .*[.!?]$`),
			regexp.MustCompile(`^[A-ZÅÄÖ].* .*[.!?]$`)},
		{"date",
			regexp.MustCompile(`^(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])/\d{4}$`),
			regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)},
		{"time",
			regexp.MustCompile(`^([1-9]|1[0-2]):[0-5]\d (AM|PM)$`),
			regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)},
		{"version", semver, semver},
		{"email", email, email},
		{"ip", ip, ip},
		{"company",
			regexp.MustCompile(`^.+ (Inc\.|LLC|Corp\.|Co\.|Group|Ltd\.)$`),
			regexp.MustCompile(`^\pL.*[\pL\d.]$`)},
		{"color", color, color},
		{"url", url, url},
		{"username", uname, uname},
		{"price",
			regexp.MustCompile(`^\$\d{1,3}(,\d{3})?\.\d{2}$`),
			regexp.MustCompile(`^\d{1,3}( \d{3})?(,\d{2})? kr$`)},
	}

	f := shipped(t, WithSeed(1))
	for _, c := range cases {
		for i := 0; i < 200; i++ {
			if v := fake(t, f, "en_US."+c.path); !c.en.MatchString(v) {
				t.Fatalf("en_US %s = %q, want %s", c.path, v, c.en)
			}
			if v := fake(t, f, "sv_SE."+c.path); !c.sv.MatchString(v) {
				t.Fatalf("sv_SE %s = %q, want %s", c.path, v, c.sv)
			}
		}
	}
}

// TestShippedMiscCategories covers the locale-neutral misc module: a proper
// v4 UUID (version/variant nibbles fixed), a MAC address, and credit-card numbers
// whose trailing {luhn()} check passes.
func TestShippedMiscCategories(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	v4 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	mac := regexp.MustCompile(`^([0-9a-f]{2}:){5}[0-9a-f]{2}$`)
	objectid := regexp.MustCompile(`^[0-9a-f]{24}$`)
	datetime := regexp.MustCompile(`^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$`)
	for i := 0; i < 200; i++ {
		if v := fake(t, f, "datetime"); !datetime.MatchString(v) {
			t.Fatalf("misc datetime %q is not an RFC 3339 UTC instant", v)
		}
		if v := fake(t, f, "objectid"); !objectid.MatchString(v) {
			t.Fatalf("misc objectid %q is not 24 hex chars", v)
		}
		if v := fake(t, f, "uuid"); !v4.MatchString(v) {
			t.Fatalf("misc uuid %q is not a valid v4", v)
		}
		if v := fake(t, f, "mac"); !mac.MatchString(v) {
			t.Fatalf("misc mac %q is not a MAC address", v)
		}
		if v := fake(t, f, "creditcard"); len(v) < 15 || len(v) > 16 || !luhnValid(v) {
			t.Fatalf("creditcard %q is not a 15-16 digit Luhn-valid number", v)
		}
	}
}

// TestShippedMiscReferenceData covers the locale-neutral reference categories:
// codes follow their standard shape, dotted sub-paths resolve, emoji is a real
// glyph, and a coordinate parses to a point within geographic bounds.
func TestShippedMiscReferenceData(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	re := map[string]*regexp.Regexp{
		"currency":              regexp.MustCompile(`^[A-Z]{3}$`),
		"currency.name":         regexp.MustCompile(`\p{L}`),
		"currency.symbol":       regexp.MustCompile(`^\S+$`),
		"territory":             regexp.MustCompile(`\p{L}`),
		"territory.alpha2":      regexp.MustCompile(`^[A-Z]{2}$`),
		"territory.alpha3":      regexp.MustCompile(`^[A-Z]{3}$`),
		"territory.country":     regexp.MustCompile(`^[A-Z]{2}$`),
		"language":              regexp.MustCompile(`\p{L}`),
		"language.code":         regexp.MustCompile(`^[a-z]{2}$`),
		"language.code3":        regexp.MustCompile(`^[a-z]{3}$`),
		"timezone":              regexp.MustCompile(`^[A-Za-z]+(/[A-Za-z0-9_+-]+)+$`),
		"timezone.offset":       regexp.MustCompile(`^[+-](0\d|1[0-4]):[0-5]\d$`),
		"timezone.territory":    regexp.MustCompile(`^[A-Z]{2}$`),
		"timezone.weight":       regexp.MustCompile(`^[1-9]\d*$`),
		"mimetype":              regexp.MustCompile(`^[a-z]+/[a-z0-9.+-]+$`),
		"mimetype.ext":          regexp.MustCompile(`^\.[a-z0-9_-]+$`),
		"httpmethod":            regexp.MustCompile(`^(CONNECT|DELETE|GET|HEAD|OPTIONS|PATCH|POST|PUT|TRACE)$`),
		"httpmethod.safe":       regexp.MustCompile(`^(true|false)$`),
		"httpmethod.idempotent": regexp.MustCompile(`^(true|false)$`),
		"port":                  regexp.MustCompile(`^[1-9]\d{0,4}$`),
		"port.service":          regexp.MustCompile(`^\S+$`),
		"protocol":              regexp.MustCompile(`^\S+( \S+)*$`),
		"protocol.name":         regexp.MustCompile(`\S`),
		"protocol.number":       regexp.MustCompile(`^\d{1,3}$`),
		"loglevel":              regexp.MustCompile(`^(alert|crit|debug|emerg|err|info|notice|warning)$`),
		"loglevel.code":         regexp.MustCompile(`^[0-7]$`),
		"loglevel.severity":     regexp.MustCompile(`^(Alert|Critical|Debug|Emergency|Error|Informational|Notice|Warning)$`),
		"tld":                   regexp.MustCompile(`^\.[a-z0-9]([a-z0-9-]*[a-z0-9])?$`),
		"tld.type":              regexp.MustCompile(`^(country-code|generic|generic-restricted|infrastructure|sponsored)$`),
		"tld.unicode":           regexp.MustCompile(`^\.\S+$`),
		"httpstatus":            regexp.MustCompile(`^[1-5]\d{2} \S.*$`),
		"httpstatus.code":       regexp.MustCompile(`^[1-5]\d{2}$`),
		"useragent":             regexp.MustCompile(`^Mozilla/5\.0 .+`),
		"useragent.browser":     regexp.MustCompile(`^(Chrome|Edge|Firefox|Opera|Safari|Samsung Internet)$`),
		"useragent.device":      regexp.MustCompile(`^(desktop|mobile)$`),
		"useragent.os":          regexp.MustCompile(`^(Android|ChromeOS|Linux|Windows|iOS|macOS)$`),
		"car":                   regexp.MustCompile(`^\S.* \S`),
		"car.make":              regexp.MustCompile(`^\S`),
	}
	for i := 0; i < 100; i++ {
		for p, rx := range re {
			if v := fake(t, f, p); !rx.MatchString(v) {
				t.Fatalf("misc %s = %q, want %s", p, v, rx)
			}
		}
		if e := fake(t, f, "emoji"); e == "" || []rune(e)[0] < 128 {
			t.Fatalf("emoji %q is not a non-ASCII glyph", e)
		}
		c := fake(t, f, "coordinate")
		parts := strings.SplitN(c, ", ", 2)
		lat, err1 := strconv.ParseFloat(parts[0], 64)
		lon, err2 := strconv.ParseFloat(parts[len(parts)-1], 64)
		if len(parts) != 2 || err1 != nil || err2 != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			t.Fatalf("coordinate %q is not a 'lat, lon' point within bounds", c)
		}
	}
}

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
	const draws = 400000
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
	const draws = 400000
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

// TestSwedishPersonnummer checks what the shape regex can't: the date is a real
// calendar date, the birth number is Skatteverket's test series (238 female, 239
// male), the trailing digit is a Luhn checksum over the other nine, and a
// samordningsnummer is the same with 60 added to the day.
func TestSwedishPersonnummer(t *testing.T) {
	sv := shipped(t, WithSeed(1))
	re := regexp.MustCompile(`^\d{6}-23[89]\d$`)
	sawLongMonthEnd, birth := false, map[string]bool{}
	for i := 0; i < 2000; i++ {
		v := fake(t, sv, "sv_SE.personnummer")
		d := digitsOnly(v)
		if !re.MatchString(v) {
			t.Fatalf("personnummer %q, want YYMMDD-238C or YYMMDD-239C", v)
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

// TestShippedUSTaxIds pins the SSA and IRS ranges: an SSN's area is 001-899 but
// 666, its group 01-99 and its serial 0001-9999; an ITIN is 9XX-GG-XXXX with GG
// in 50-65, 70-88, 90-92 or 94-99.
func TestShippedUSTaxIds(t *testing.T) {
	f := shipped(t, WithSeed(5))
	ssn := regexp.MustCompile(`^(\d{3})-(\d{2})-(\d{4})$`)
	itin := regexp.MustCompile(`^9\d{2}-(5\d|6[0-5]|7\d|8[0-8]|9[0-24-9])-\d{4}$`)
	for i := 0; i < 2000; i++ {
		v := fake(t, f, "en_US.ssn")
		m := ssn.FindStringSubmatch(v)
		if m == nil {
			t.Fatalf("ssn %q, want AAA-GG-SSSS", v)
		}
		area, _ := strconv.Atoi(m[1])
		group, _ := strconv.Atoi(m[2])
		serial, _ := strconv.Atoi(m[3])
		if area < 1 || area > 899 || area == 666 || group < 1 || serial < 1 {
			t.Fatalf("ssn %q is in a range the SSA never assigns", v)
		}
		if v := fake(t, f, "en_US.itin"); !itin.MatchString(v) {
			t.Fatalf("itin %q, want %s", v, itin)
		}
	}
}

// TestShippedPersonNames pins the name tables: a sex pins its first names, a name
// both sexes carry appears under both, and the record's sex agrees with its name.
func TestShippedPersonNames(t *testing.T) {
	f := shipped(t, WithSeed(9))
	for path, want := range map[string]string{
		"sv_SE.sex[f]":                        "kvinna",
		"sv_SE.sex[man].code":                 "m",
		"en_US.sex[f]":                        "female",
		"sv_SE.sex[f].first-name[Anna]":       "Anna",
		"sv_SE.sex[m].first-name[Erik].sex":   "m",
		"en_US.sex[m].first-name[James]":      "James",
		"en_US.sex[f].first-name[Taylor].sex": "f",
		"en_US.sex[m].first-name[Taylor].sex": "m",
		"sv_SE.last-name[Andersson]":          "Andersson",
		"sv_SE.last-name[Andersson].count":    "",
		"en_US.last-name[Smith]":              "Smith",
	} {
		got := fake(t, f, path)
		if want == "" && !regexp.MustCompile(`^[1-9]\d*$`).MatchString(got) {
			t.Errorf("Fake(%q) = %q, want a count", path, got)
		} else if want != "" && got != want {
			t.Errorf("Fake(%q) = %q, want %q", path, got, want)
		}
	}
	if _, err := f.Fake("en_US.first-name[Taylor]"); err == nil || !strings.Contains(err.Error(), "sex[f].first-name[Taylor]") {
		t.Fatalf("Fake(en_US.first-name[Taylor]) = %v, want both sexes' rows listed", err)
	}
	for _, locale := range []string{"sv_SE", "en_US", "nb_NO"} {
		seen := map[string]bool{}
		for i := 0; i < 2000; i++ {
			seen[fake(t, f, locale+".sex[f].first-name")] = true
			got := fakeTemplate(t, f, `{/`+locale+`.person as p}{p.first}|{p.sex}`)
			first, sex, _ := strings.Cut(got, "|")
			fake(t, f, locale+".sex["+sex+"].first-name["+first+"]")
		}
		if len(seen) < 300 || !seen["Anna"] && locale == "sv_SE" || !seen["Mary"] && locale == "en_US" || !seen["Kari"] && locale == "nb_NO" {
			t.Fatalf("%s female first names: %d distinct in 2000, want a weighted register", locale, len(seen))
		}
	}
}

// TestShippedPersonIsOneNamedPick pins that a person's first name, sex and title come
// from one pick, and that two names bound to one person category are two people.
func TestShippedPersonIsOneNamedPick(t *testing.T) {
	f := shipped(t, WithSeed(9))
	for _, locale := range []string{"sv_SE", "en_US", "nb_NO"} {
		apart := false
		for i := 0; i < 300; i++ {
			got := fakeTemplate(t, f, `{/`+locale+`.person as a}{/`+locale+`.person as b}{a.first}|{a.sex}|{a.prefix}|{b.first}|{b.sex}`)
			parts := strings.Split(got, "|")
			for _, p := range [][2]string{{parts[0], parts[1]}, {parts[3], parts[4]}} {
				fake(t, f, locale+".sex["+p[1]+"].first-name["+p[0]+"]")
			}
			if title := strings.TrimSpace(parts[2]); title != "" && locale == "en_US" {
				fake(t, f, "en_US.sex["+parts[1]+"].title["+title+"]")
			}
			apart = apart || parts[0] != parts[3]
		}
		if !apart {
			t.Errorf("%s: two names bound to person drew one first name in 300 renders, want two people", locale)
		}
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

// TestShippedAddressIsOneNamedPick pins that an address's street, postal code,
// locality and region are one place in the geo tables, read through a name in a template or
// across a struct's tags, and that a reference beside the name is another address.
func TestShippedAddressIsOneNamedPick(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ locale, country, template, largest string }{
		{"sv_SE", "SE", "{a.street}|{a.postal-code}|{a.locality}|", "Stockholm"},
		{"en_US", "US", "{a.street}|{a.postal-code}|{a.locality}|{a.region}", "New York"},
		{"nb_NO", "NO", "{a.street}|{a.postal-code}|{a.locality}|", "Oslo"},
	} {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			f := shipped(t, WithSeed(3))
			places := geoPlaces(t, c.country)
			apart, localities := false, map[string]int{}
			for i := 0; i < 2000; i++ {
				got := fakeTemplate(t, f, `{/`+c.locale+`.address as a}`+c.template+`|{/`+c.locale+`.address.street}`)
				parts := strings.Split(got, "|")
				if !places[strings.Join(parts[:4], "|")] {
					t.Fatalf("%s.address: %q is no street, postal code, locality and region of one place", c.locale, got)
				}
				apart = apart || parts[0] != parts[4]
				localities[parts[2]]++
			}
			if !apart {
				t.Errorf("%s: a reference beside a name drew the name's street in 2000 renders, want two addresses", c.locale)
			}
			if c.country == "SE" {
				var row struct {
					Street     string `fake:"{/sv_SE.address as a}{a.street}"`
					PostalCode string `fake:"{a.postal-code}"`
					Locality   string `fake:"{a.locality}"`
				}
				for i := 0; i < 200; i++ {
					if err := f.FakeStruct(&row); err != nil {
						t.Fatal(err)
					}
					if got := row.Street + "|" + row.PostalCode + "|" + row.Locality + "|"; !places[got] {
						t.Fatalf("sv_SE.address: struct tags reading one name, %q, are no street, postal code and locality of one place", got)
					}
				}
			}
			for l, n := range localities {
				if n > localities[c.largest] {
					t.Errorf("%s.address drew %s %d times and %s %d, want the locality weighted by population", c.locale, l, n, c.largest, localities[c.largest])
				}
			}
		})
	}
}

// 0644000 is Los Angeles.
func TestShippedPlacesSpreadAddresses(t *testing.T) {
	for country, city := range map[string]struct {
		key   string
		least int
	}{"NO": {"Oslo", 2000}, "SE": {"Göteborg", 1500}, "US": {"0644000", 2500}} {
		streets := 0
		for _, r := range geoRows(t, country, "street") {
			if r[1] == city.key {
				streets++
			}
		}
		if streets < city.least {
			t.Errorf("geo.%s: %s ships %d streets, want at least %d", country, city.key, streets, city.least)
		}
	}
	header, _, _ := strings.Cut(geoTSV(t, "US", "locality"), "\n")
	population := slices.Index(strings.Split(header, "\t"), "population")
	if population < 0 {
		t.Fatal("geo.US locality.tsv has no population column")
	}
	for _, r := range geoRows(t, "US", "locality") {
		n, err := strconv.Atoi(r[population])
		if err != nil {
			t.Fatal(err)
		}
		if n < 2500 {
			return
		}
	}
	t.Error("geo.US: every locality holds 2500 people or more, want places drawn down to villages")
}

// TestShippedCapitalsAndBigPlaces pins state capitals and places of 100,000 or more that a
// draw by population leaves out.
func TestShippedCapitalsAndBigPlaces(t *testing.T) {
	state := map[string]string{}
	for _, r := range geoRows(t, "US", "municipality") {
		state[r[0]] = r[2]
	}
	shipped := map[string]bool{}
	for _, r := range geoRows(t, "US", "locality") {
		shipped[r[1]+", "+state[r[2]]] = true
	}
	for _, place := range []string{
		"Albany, NY", "Augusta, ME", "Baton Rouge, LA", "Bismarck, ND", "Boise, ID", "Charleston, WV",
		"Frankfort, KY", "Harrisburg, PA", "Helena, MT", "Juneau, AK", "Kansas City, KS", "Lafayette, LA",
		"Lexington, KY", "Montpelier, VT", "New Orleans, LA", "Pierre, SD", "St. Paul, MN", "Stockton, CA",
		"Ventura, CA",
	} {
		if !shipped[place] {
			t.Errorf("geo.US: no locality %s, want every state capital and place of 100,000 or more", place)
		}
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

func geoTSV(t *testing.T, country, table string) string {
	b, err := os.ReadFile("data/geo/" + country + "/" + table + ".tsv")
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func geoRows(t *testing.T, country, table string) [][]string {
	var out [][]string
	for _, line := range strings.Split(geoTSV(t, country, table), "\n")[1:] {
		out = append(out, strings.Split(line, "\t"))
	}
	return out
}

// geoPlaces reads a country's geo tables into every "street|postal-code|locality|region"
// one locality holds, the region empty for NO and SE, whose addresses carry none.
func geoPlaces(t *testing.T, country string) map[string]bool {
	rows := func(table string) [][]string { return geoRows(t, country, table) }
	name, region := map[string]string{}, map[string]string{}
	if country == "US" {
		county := map[string]string{}
		for _, r := range rows("municipality") {
			county[r[0]] = r[2]
		}
		for _, r := range rows("locality") {
			name[r[0]], region[r[0]] = r[1], county[r[2]]
		}
	} else {
		for _, r := range rows("locality") {
			name[r[0]] = r[0]
		}
	}
	codes := map[string][]string{}
	for _, r := range rows("postal-code") {
		codes[r[1]] = append(codes[r[1]], r[0])
	}
	places := map[string]bool{}
	for _, r := range rows("street") {
		for _, code := range codes[r[1]] {
			places[strings.Join([]string{r[0], code, name[r[1]], region[r[1]]}, "|")] = true
		}
	}
	return places
}

func TestShippedPersonHasParts(t *testing.T) {
	f := shipped(t, WithSeed(7))
	for _, path := range []string{"sv_SE.person", "en_US.person"} {
		for i := 0; i < 30; i++ {
			if name := fake(t, f, path); len(name) < 3 || !regexp.MustCompile(`\S \S`).MatchString(name) {
				t.Fatalf("%s %q lacks first and last name", path, name)
			}
		}
	}
}

// NANPA's central office code records mark N11, 555, 950, 958, 959 and 976 unassignable.
func TestShippedUSPhone(t *testing.T) {
	f := shipped(t, WithSeed(11))
	re := regexp.MustCompile(`^(?:\(\d{3}\) |\d{3}-)([2-9]\d{2})-\d{4}$`)
	unassignable := map[string]bool{"555": true, "950": true, "958": true, "959": true, "976": true}
	for i := 0; i < 2000; i++ {
		n := fake(t, f, "en_US.phone")
		m := re.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q does not match %s", n, re)
		}
		if exchange := m[1]; exchange[1:] == "11" || unassignable[exchange] {
			t.Fatalf("phone %q has the unassignable exchange %s", n, exchange)
		}
	}
}

func TestShippedTableCellsAreTrimmed(t *testing.T) {
	err := fs.WalkDir(os.DirFS("data"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || !strings.HasSuffix(path, ".tsv") {
			return err
		}
		b, err := fs.ReadFile(os.DirFS("data"), path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			for _, cell := range strings.Split(line, "\t") {
				if cell != strings.TrimSpace(cell) {
					t.Errorf("data/%s:%d: cell %q is padded", path, i+1, cell)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestShippedNamespacedTree loads every shipped module and reaches each locale through its
// namespace: data/sv_SE/person.json, mounted under sv_SE, is sv_SE.person.
func TestShippedNamespacedTree(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for _, path := range []string{"sv_SE.person", "en_US.person", "sv_SE.address.locality"} {
		if got := fake(t, f, path); got == "" {
			t.Fatalf("Fake(%q) returned empty", path)
		}
	}
	if _, err := f.Fake("person"); err == nil {
		t.Fatal("Fake(person) = nil error; categories should be namespaced under the locale folder")
	}
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// luhnValid verifies a full number (payload + trailing check digit). It doubles
// from the second-from-right, independent of the builtin's own Luhn code.
func luhnValid(s string) bool {
	sum, double := 0, false
	for i := len(s) - 1; i >= 0; i-- {
		n := int(s[i] - '0')
		if double {
			if n *= 2; n > 9 {
				n -= 9
			}
		}
		double = !double
		sum += n
	}
	return sum%10 == 0
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

func TestShippedStreetNumberFormats(t *testing.T) {
	// Reachable via a hyphenated path; covers all five weighted number variants.
	f := shipped(t, WithSeed(8))
	re := regexp.MustCompile(`^[1-9]\d{0,2}[A-Z]?$`)
	for i := 0; i < 300; i++ {
		if n := fake(t, f, "sv_SE.address.street-number"); !re.MatchString(n) {
			t.Fatalf("street-number %q does not match %s", n, re)
		}
	}
}

// TestShippedUSTitleAgreesWithSex pins that one person's title and sex agree, and
// that the everyday titles are reachable.
func TestShippedUSTitleAgreesWithSex(t *testing.T) {
	f := shipped(t, WithSeed(4))
	female := map[string]bool{"Miss": true, "Mrs": true, "Ms": true}
	seen := map[string]bool{}
	for i := 0; i < 3000; i++ {
		got := fakeTemplate(t, f, `{/en_US.person as p}{p.prefix}|{p.sex}`)
		title, sex, _ := strings.Cut(got, "|")
		title = strings.TrimSpace(title)
		seen[title] = true
		if title == "Mr" && sex != "male" || female[title] && sex != "female" {
			t.Fatalf("%q: the title contradicts the sex", got)
		}
	}
	for _, want := range []string{"", "Mr", "Ms"} {
		if !seen[want] {
			t.Errorf("en_US.person.prefix never drew %q in 3000 draws", want)
		}
	}
	swedish := map[string]bool{"dr": true, "prof": true}
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "sv_SE.title"); !swedish[got] {
			t.Fatalf("sv_SE.title = %q, want an unsexed Swedish honorific", got)
		}
	}
}
