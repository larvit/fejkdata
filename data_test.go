package fejkdata

import (
	"regexp"
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
			regexp.MustCompile(`^.+ (AB|HB|KB)$`)},
		{"color", color, color},
		{"url", url, url},
		{"username", uname, uname},
		{"price",
			regexp.MustCompile(`^\$\d{1,3}(,\d{3})?\.\d{2}$`),
			regexp.MustCompile(`^\d{1,3}( \d{3})?(,\d{2})? kr$`)},
	}

	f := newGenerator(t, "data", WithSeed(1))
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

// TestShippedMiscCategories covers the locale-neutral data/misc folder: a proper
// v4 UUID (version/variant nibbles fixed), a MAC address, and credit-card numbers
// whose trailing {luhn()} check passes.
func TestShippedMiscCategories(t *testing.T) {
	f := newGenerator(t, "data/misc", WithSeed(1))
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
	f := newGenerator(t, "data/misc", WithSeed(1))
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
	f := newGenerator(t, "data", WithSeed(11))
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

// TestSwedishPersonnummer checks what the shape regex can't: the date is a real
// calendar date, the birth number is Skatteverket's test series (238 female, 239
// male), the trailing digit is a Luhn checksum over the other nine, and a
// samordningsnummer is the same with 60 added to the day.
func TestSwedishPersonnummer(t *testing.T) {
	sv := newGenerator(t, "data", WithSeed(1))
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
	for i := 0; i < 200; i++ {
		got := fakeTemplate(t, sv, `{/sv_SE.person.sex} {/sv_SE.personnummer}`)
		sex, id, _ := strings.Cut(got, " ")
		if want := map[string]string{"kvinna": "238", "man": "239"}[sex]; want == "" || !strings.Contains(id, "-"+want) {
			t.Fatalf("%q: a person and a personnummer in one render disagree on sex", got)
		}
	}
}

// TestShippedUSTaxIds pins the SSA and IRS ranges: an SSN's area is 001-899 but
// 666, its group 01-99 and its serial 0001-9999; an ITIN is 9XX-GG-XXXX with GG
// in 50-65, 70-88, 90-92 or 94-99.
func TestShippedUSTaxIds(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(5))
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
	f := newGenerator(t, "data", WithSeed(9))
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
	for _, locale := range []string{"sv_SE", "en_US"} {
		seen := map[string]bool{}
		for i := 0; i < 2000; i++ {
			seen[fake(t, f, locale+".sex[f].first-name")] = true
			first, sex := fake(t, f, locale+".person.first"), fake(t, f, locale+".person.sex")
			if first == "" || sex == "" {
				t.Fatalf("%s.person lacks a first name or a sex", locale)
			}
		}
		if len(seen) < 300 || !seen["Anna"] && locale == "sv_SE" || !seen["Mary"] && locale == "en_US" {
			t.Fatalf("%s female first names: %d distinct in 2000, want a weighted register", locale, len(seen))
		}
		got := fakeTemplate(t, f, `{/`+locale+`.sex.code}|{/`+locale+`.person.first}`)
		code, first, _ := strings.Cut(got, "|")
		if v := fake(t, f, locale+".sex["+code+"].first-name["+first+"]"); v != first {
			t.Fatalf("%s: %q drawn as a %s name is not one", locale, first, code)
		}
	}
}

func TestShippedSwedishPhone(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(11))
	re := regexp.MustCompile(`^0\d{1,2}-\d{3} \d{2} \d{2}$`)
	for i := 0; i < 50; i++ {
		if n := fake(t, f, "sv_SE.phone"); !re.MatchString(n) {
			t.Fatalf("phone %q does not match %s", n, re)
		}
	}
}

func TestShippedSwedishAddress(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(3))
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

func TestShippedPersonHasParts(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(7))
	for _, path := range []string{"sv_SE.person", "en_US.person"} {
		for i := 0; i < 30; i++ {
			if name := fake(t, f, path); len(name) < 3 || !regexp.MustCompile(`\S \S`).MatchString(name) {
				t.Fatalf("%s %q lacks first and last name", path, name)
			}
		}
	}
}

func TestShippedUSPhone(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(11))
	re := regexp.MustCompile(`^(\(\d{3}\) \d{3}-\d{4}|\d{3}-\d{3}-\d{4})$`)
	for i := 0; i < 50; i++ {
		if n := fake(t, f, "en_US.phone"); !re.MatchString(n) {
			t.Fatalf("phone %q does not match %s", n, re)
		}
	}
}

// TestShippedNamespacedTree loads the whole data/ tree (not a single locale) and
// reaches each locale through its folder segment: data/sv_SE/person -> sv_SE.person.
func TestShippedNamespacedTree(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(1))
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
	f := newGenerator(t, "data", WithSeed(5))
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
	f := newGenerator(t, "data", WithSeed(6))
	for i := 0; i < 300; i++ {
		if s := fake(t, f, "sv_SE.person.last"); !swedishName.MatchString(s) {
			t.Fatalf("last name %q is not a Swedish surname", s)
		}
	}
}

func TestShippedStreetNumberFormats(t *testing.T) {
	// Reachable via a hyphenated path; covers all five weighted number variants.
	f := newGenerator(t, "data", WithSeed(8))
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
	f := newGenerator(t, "data", WithSeed(4))
	female := map[string]bool{"Miss": true, "Mrs": true, "Ms": true}
	seen := map[string]bool{}
	for i := 0; i < 3000; i++ {
		got := fakeTemplate(t, f, `{/en_US.person.prefix}|{/en_US.person.sex}`)
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
