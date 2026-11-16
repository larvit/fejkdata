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
