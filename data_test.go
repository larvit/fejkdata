package fejkdata

import (
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
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
		t.Run(c.path, func(t *testing.T) {
			t.Run("en_US", func(t *testing.T) {
				fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "en_US."+c.path) }, func(v string) {
					if !c.en.MatchString(v) {
						t.Fatalf("en_US %s = %q, want %s", c.path, v, c.en)
					}
				})
			})
			t.Run("sv_SE", func(t *testing.T) {
				fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "sv_SE."+c.path) }, func(v string) {
					if !c.sv.MatchString(v) {
						t.Fatalf("sv_SE %s = %q, want %s", c.path, v, c.sv)
					}
				})
			})
		})
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
	t.Run("datetime", func(t *testing.T) {
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "datetime") }, func(v string) {
			if !datetime.MatchString(v) {
				t.Fatalf("misc datetime %q is not an RFC 3339 UTC instant", v)
			}
		})
	})
	t.Run("objectid", func(t *testing.T) {
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "objectid") }, func(v string) {
			if !objectid.MatchString(v) {
				t.Fatalf("misc objectid %q is not 24 hex chars", v)
			}
		})
	})
	t.Run("uuid", func(t *testing.T) {
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "uuid") }, func(v string) {
			if !v4.MatchString(v) {
				t.Fatalf("misc uuid %q is not a valid v4", v)
			}
		})
	})
	t.Run("mac", func(t *testing.T) {
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "mac") }, func(v string) {
			if !mac.MatchString(v) {
				t.Fatalf("misc mac %q is not a MAC address", v)
			}
		})
	})
	t.Run("creditcard", func(t *testing.T) {
		fejktest.EachDraw(200, func() string { return fejktest.Fake(t, f, "creditcard") }, func(v string) {
			if len(v) < 15 || len(v) > 16 || !fejktest.LuhnValid(v) {
				t.Fatalf("creditcard %q is not a 15-16 digit Luhn-valid number", v)
			}
		})
	})
}

func eachMiscTestNumberOf300Draws(t *testing.T, f *Generator, check func(path, v string)) {
	t.Helper()
	for i := 0; i < 300; i++ {
		for _, path := range []string{"creditcard-test", "iban-test"} {
			check(path, fejktest.Fake(t, f, path))
		}
	}
}

func valuesOfMiscTestNumbers(tables ...map[string]string) map[string]bool {
	published := map[string]bool{}
	for _, table := range tables {
		for _, number := range table {
			published[number] = true
		}
	}
	return published
}

func TestShippedMiscTestNumbers(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	cards := map[string]string{
		"visa": "4242424242424242", "mastercard": "5555555555554444", "amex": "378282246310005",
		"discover": "6011111111111117", "diners": "3056930009020004", "jcb": "3566002020360505",
		"unionpay": "6200000000000005",
	}
	ibans := map[string]string{
		"BE": "BE62510007547061", "DE": "DE89370400440532013000", "DK": "DK5000400440116243",
		"ES": "ES0700120345030000067890", "FI": "FI2112345600000785", "NO": "NO9386011117947",
		"SE": "SE3550000000054910000003",
	}
	t.Run("creditcard-test", func(t *testing.T) {
		for brand, number := range cards {
			t.Run(brand, func(t *testing.T) {
				if got := fejktest.Fake(t, f, "creditcard-test["+brand+"]"); got != number {
					t.Fatalf("creditcard-test[%s] = %q, want %q", brand, got, number)
				}
				if !fejktest.LuhnValid(number) {
					t.Fatalf("test card %q is not Luhn-valid", number)
				}
			})
		}
	})
	t.Run("iban-test", func(t *testing.T) {
		for cc, number := range ibans {
			t.Run(cc, func(t *testing.T) {
				if got := fejktest.Fake(t, f, "iban-test["+cc+"]"); got != number {
					t.Fatalf("iban-test[%s] = %q, want %q", cc, got, number)
				}
				if !ibanValid(number) {
					t.Fatalf("test IBAN %q fails the mod-97 check", number)
				}
			})
		}
	})
	t.Run("a draw reaches every published number and no other", func(t *testing.T) {
		published := valuesOfMiscTestNumbers(cards, ibans)
		seen := map[string]bool{}
		eachMiscTestNumberOf300Draws(t, f, func(path, v string) {
			if !published[v] {
				t.Fatalf("%s drew %q, which is no published test number", path, v)
			}
			seen[v] = true
		})
		if len(seen) != len(published) {
			t.Fatalf("300 draws reached %d of %d test numbers", len(seen), len(published))
		}
	})
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
		"territory.sovereign":   regexp.MustCompile(`^[A-Z]{2}$`),
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
	t.Run("shape", func(t *testing.T) {
		for p, rx := range re {
			t.Run(p, func(t *testing.T) {
				fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, p) }, func(v string) {
					if !rx.MatchString(v) {
						t.Fatalf("misc %s = %q, want %s", p, v, rx)
					}
				})
			})
		}
	})
	t.Run("emoji", func(t *testing.T) {
		fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, "emoji") }, func(e string) {
			if e == "" || []rune(e)[0] < 128 {
				t.Fatalf("emoji %q is not a non-ASCII glyph", e)
			}
		})
	})
	t.Run("coordinate", func(t *testing.T) {
		fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, "coordinate") }, func(c string) {
			parts := strings.SplitN(c, ", ", 2)
			lat, err1 := strconv.ParseFloat(parts[0], 64)
			lon, err2 := strconv.ParseFloat(parts[len(parts)-1], 64)
			if len(parts) != 2 || err1 != nil || err2 != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
				t.Fatalf("coordinate %q is not a 'lat, lon' point within bounds", c)
			}
		})
	})
}

// TestShippedPersonNames pins the name tables: a sex pins its first names, a name
// both sexes carry appears under both, and the record's sex agrees with its name.
func TestShippedPersonNames(t *testing.T) {
	f := shipped(t, WithSeed(9))
	t.Run("a selector reads its row", func(t *testing.T) {
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
			"en_US.last-name[Smith]":              "Smith",
		} {
			t.Run(path, func(t *testing.T) {
				if got := fejktest.Fake(t, f, path); got != want {
					t.Errorf("Fake(%q) = %q, want %q", path, got, want)
				}
			})
		}
	})
	t.Run("a last name carries its count", func(t *testing.T) {
		path := "sv_SE.last-name[Andersson].count"
		if got := fejktest.Fake(t, f, path); !regexp.MustCompile(`^[1-9]\d*$`).MatchString(got) {
			t.Errorf("Fake(%q) = %q, want a count", path, got)
		}
	})
	t.Run("a name both sexes carry is refused without its sex", func(t *testing.T) {
		if _, err := f.Fake("en_US.first-name[Taylor]"); err == nil || !strings.Contains(err.Error(), "sex[f].first-name[Taylor]") {
			t.Fatalf("Fake(en_US.first-name[Taylor]) = %v, want both sexes' rows listed", err)
		}
	})
	t.Run("female first names are a weighted register", func(t *testing.T) {
		for _, locale := range []string{"sv_SE", "en_US", "nb_NO", "da_DK"} {
			t.Run(locale, func(t *testing.T) {
				seen := femaleFirstNamesDrawnBesidePeopleOfTheirSex(t, f, locale)
				if len(seen) < 300 || !seen["Anna"] && locale == "sv_SE" || !seen["Mary"] && locale == "en_US" || !seen["Kari"] && locale == "nb_NO" || !seen["Mette"] && locale == "da_DK" {
					t.Fatalf("%s female first names: %d distinct in 2000, want a weighted register", locale, len(seen))
				}
			})
		}
	})
}

func femaleFirstNamesDrawnBesidePeopleOfTheirSex(t *testing.T, f *Generator, locale string) map[string]bool {
	t.Helper()
	seen := map[string]bool{}
	for i := 0; i < 2000; i++ {
		seen[fejktest.Fake(t, f, locale+".sex[f].first-name")] = true
		got := fejktest.FakeTemplate(t, f, `{/`+locale+`.person as p}{p.first}|{p.sex}`)
		first, sex, _ := strings.Cut(got, "|")
		fejktest.Fake(t, f, locale+".sex["+sex+"].first-name["+first+"]")
	}
	return seen
}

// TestShippedPersonIsOneNamedPick pins that a person's first name, sex and title come
// from one pick, and that two names bound to one person category are two people.
func TestShippedPersonIsOneNamedPick(t *testing.T) {
	f := shipped(t, WithSeed(9))
	for _, locale := range []string{"sv_SE", "en_US", "nb_NO", "da_DK"} {
		t.Run(locale, func(t *testing.T) {
			if !twoPeopleOf300RendersDrawApart(t, f, locale) {
				t.Errorf("%s: two names bound to person drew one first name in 300 renders, want two people", locale)
			}
		})
	}
}

func twoPeopleOf300RendersDrawApart(t *testing.T, f *Generator, locale string) bool {
	t.Helper()
	apart := false
	for i := 0; i < 300; i++ {
		got := fejktest.FakeTemplate(t, f, `{/`+locale+`.person as a}{/`+locale+`.person as b}{a.first}|{a.sex}|{a.prefix}|{b.first}|{b.sex}`)
		parts := strings.Split(got, "|")
		for _, p := range [][2]string{{parts[0], parts[1]}, {parts[3], parts[4]}} {
			fejktest.Fake(t, f, locale+".sex["+p[1]+"].first-name["+p[0]+"]")
		}
		if title := strings.TrimSpace(parts[2]); title != "" && locale == "en_US" {
			fejktest.Fake(t, f, "en_US.sex["+parts[1]+"].title["+title+"]")
		}
		apart = apart || parts[0] != parts[3]
	}
	return apart
}

func TestShippedPersonHasParts(t *testing.T) {
	f := shipped(t, WithSeed(7))
	for _, path := range []string{"sv_SE.person", "en_US.person"} {
		t.Run(path, func(t *testing.T) {
			fejktest.EachDraw(30, func() string { return fejktest.Fake(t, f, path) }, func(name string) {
				if len(name) < 3 || !regexp.MustCompile(`\S \S`).MatchString(name) {
					t.Fatalf("%s %q lacks first and last name", path, name)
				}
			})
		})
	}
}

func tsvCellsAreTrimmed(t *testing.T, path string, b []byte) {
	t.Helper()
	for i, line := range strings.Split(string(b), "\n") {
		for _, cell := range strings.Split(line, "\t") {
			if cell != strings.TrimSpace(cell) {
				t.Errorf("data/%s:%d: cell %q is padded", path, i+1, cell)
			}
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
		tsvCellsAreTrimmed(t, path, b)
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
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got == "" {
				t.Fatalf("Fake(%q) returned empty", path)
			}
		})
	}
}

func TestShippedCategoryIsNotReachedWithoutItsNamespace(t *testing.T) {
	f := shipped(t, WithSeed(1))
	if _, err := f.Fake("person"); err == nil {
		t.Fatal("Fake(person) = nil error; categories should be namespaced under the locale folder")
	}
}

func TestShippedStreetNumberFormats(t *testing.T) {
	// Reachable via a hyphenated path; covers all five weighted number variants.
	f := shipped(t, WithSeed(8))
	re := regexp.MustCompile(`^[1-9]\d{0,2}[A-Z]?$`)
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, f, "sv_SE.address.street-number") }, func(n string) {
		if !re.MatchString(n) {
			t.Fatalf("street-number %q does not match %s", n, re)
		}
	})
}

// The sex tables hold the people of each sex: SCB's population at the end of 2024, Statistics
// Norway's and Denmark's on 1 January 2026 and the Census Bureau's estimate for 1 July 2024.
func TestShippedSexIsWeightedByPopulation(t *testing.T) {
	f := shipped(t, WithSeed(3))
	for path, want := range map[string]string{
		"sv_SE.sex[f].count": "5262925",
		"sv_SE.sex[m].count": "5324785",
		"nb_NO.sex[f].count": "2790483",
		"nb_NO.sex[m].count": "2836917",
		"da_DK.sex[f].count": "3029291",
		"da_DK.sex[m].count": "2996312",
		"en_US.sex[f].count": "171773106",
		"en_US.sex[m].count": "168337882",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("Fake(%q) = %q, want %q", path, got, want)
			}
		})
	}
}
