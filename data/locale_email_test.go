package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

var exampleEmail = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*@([a-z0-9-]+\.)*example\.(com|org|net)$`)

func hasEmail(l shippedLocale) bool { return l.emailFold != nil }

func TestLocaleEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	eachLocale(t, hasEmail, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(4))
		fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, l.name+".email") }, func(got string) {
			if !exampleEmail.MatchString(got) {
				t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
			}
		})
	})
}

func TestLocaleEmailsAreMostlyDistinct(t *testing.T) {
	eachLocale(t, hasEmail, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(4))
		seen := fejktest.TallyDraws(5000, func() string { return fejktest.Fake(t, f, l.name+".email") })
		if len(seen) < 4000 {
			t.Errorf("5000 emails drew %d distinct, want 4000 or more", len(seen))
		}
	})
}

// nameSlugs reads a names table as its names in lower case, folded as the locale's email folds them.
func nameSlugs(t *testing.T, l shippedLocale, file string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for name := range nameColumnSet(t, l.name+"/"+file+".tsv") {
		out[l.emailFold.Replace(strings.ToLower(name))] = true
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

func TestLocaleEmailHoldsAFirstAndALastNameInHalfItsAddresses(t *testing.T) {
	trailingDigits := regexp.MustCompile(`\d+$`)
	eachLocale(t, hasEmail, func(t *testing.T, l shippedLocale) {
		firsts, lasts := nameSlugs(t, l, "first-name"), nameSlugs(t, l, "last-name")
		f := shipped(t, fejkdata.WithSeed(9))
		seen := fejktest.TallyDraws(5000, func() string {
			local, _, _ := strings.Cut(fejktest.Fake(t, f, l.name+".email"), "@")
			return lastNameInEmail(trailingDigits.ReplaceAllString(local, ""), firsts, lasts)
		})
		named := 5000 - seen[""]
		delete(seen, "")
		assertShare(t, "addresses holding a first and a last name", named, 5000, 0.5)
		if len(seen) < 40 {
			t.Errorf("5000 emails held %d distinct last names, want 40 or more", len(seen))
		}
	})
}

var urlLocales = []string{"da_DK", "de_DE", "en_US", "fi_FI", "nb_NO", "nn_NO", "sv_SE"}

var reservedHost = regexp.MustCompile(`^https://([a-z]+\.)?example\.(com|org|net)` + urlPath + `$`)

func TestEmailHasAReservedPathThatDrawsOnlyDomainsNobodyCanRegister(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	for _, name := range append([]string{"et_EE", "fr_FR"}, urlLocales...) {
		t.Run(name, func(t *testing.T) {
			fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, name+".email.reserved") }, func(got string) {
				if !exampleEmail.MatchString(got) {
					t.Fatalf("%s.email.reserved drew %q, want lower-case ASCII at example.com, example.org or example.net", name, got)
				}
			})
		})
	}
}

func TestUrlHasAReservedPathThatDrawsOnlyDomainsNobodyCanRegister(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	for _, name := range urlLocales {
		t.Run(name, func(t *testing.T) {
			fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, name+".url.reserved") }, func(got string) {
				if !reservedHost.MatchString(got) {
					t.Fatalf("%s.url.reserved drew %q, want a host at example.com, example.org or example.net", name, got)
				}
			})
		})
	}
}

// The share of the default draw that is already reserved: the reserved hosts of the old flat list over all its hosts.
func TestUrlAndEmailKeepTheirMixOfReservedAndRegistrableDomainsByDefault(t *testing.T) {
	const draws = 6000
	f := shipped(t, fejkdata.WithSeed(4))
	urlShare := map[string]float64{
		"da_DK": 7.0 / 12, "de_DE": 7.0 / 12, "en_US": 13.0 / 16, "fi_FI": 7.0 / 12, "nb_NO": 7.0 / 12, "nn_NO": 7.0 / 12, "sv_SE": 7.0 / 12,
	}
	for name, want := range urlShare {
		t.Run(name+".url", func(t *testing.T) {
			seen := fejktest.TallyDraws(draws, func() string { return fejktest.Fake(t, f, name+".url") })
			var reserved int
			for got, n := range seen {
				if reservedHost.MatchString(got) {
					reserved += n
				}
			}
			assertShare(t, name+".url at example.com, example.org or example.net", reserved, draws, want)
		})
	}
	for name, want := range map[string]float64{"en_US": 10.0 / 15, "sv_SE": 4.0 / 13} {
		t.Run(name+".email", func(t *testing.T) {
			seen := fejktest.TallyDraws(draws, func() string { return fejktest.Fake(t, f, name+".email") })
			var reserved int
			for got, n := range seen {
				if exampleEmail.MatchString(got) {
					reserved += n
				}
			}
			assertShare(t, name+".email at example.com, example.org or example.net", reserved, draws, want)
		})
	}
}
