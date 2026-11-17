package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

const companyLegalDraws = fejktest.ShareDraws / 10

func hasCompany(l shippedLocale) bool { return l.company != nil }

func sumOfCounts(counts map[string]int) int {
	var total int
	for _, n := range counts {
		total += n
	}
	return total
}

// legalFormSuffix matches the legal form a company name ends in, and captures it.
func (c *company) legalFormSuffix() *regexp.Regexp {
	var suffixes []string
	for _, form := range c.suffixed {
		suffixes = append(suffixes, regexp.QuoteMeta(form.suffix))
	}
	return regexp.MustCompile(` (` + strings.Join(suffixes, "|") + `)$`)
}

func (c *company) forms() []legalForm { return append([]legalForm{c.bare}, c.suffixed...) }

func eachCompany(t *testing.T, locale string, check func(got string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(6))
	fejktest.EachDraw(fejktest.ShareDraws, func() string { return fejktest.Fake(t, f, locale+".company") }, check)
}

func eachCompanyWithItsLegal(t *testing.T, locale string, check func(company, legal string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(4))
	fejktest.EachDraw(companyLegalDraws, func() string { return fejktest.FakeTemplate(t, f, "{/"+locale+".company as c}{c}|{c.legal}") }, func(got string) {
		company, legal, _ := strings.Cut(got, "|")
		check(company, legal)
	})
}

func TestLocaleCompanyIsANameThatMayEndInALegalForm(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		eachCompany(t, l.name, func(got string) {
			if !l.company.name.MatchString(got) || strings.HasSuffix(got, " "+l.company.bare.code) {
				t.Fatalf("company %q is not a name, optionally ending in a legal form, that doesn't end in %q", got, l.company.bare.code)
			}
		})
	})
}

func TestLocaleCompanyEndsInALegalFormAsOftenAsItsSourceCountsTheForm(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		suffix, bare := l.company.legalFormSuffix(), l.company.bare.code
		registered, drawn := map[string]int{}, map[string]int{}
		for _, form := range l.company.forms() {
			registered[form.code] = form.units
		}
		eachCompany(t, l.name, func(got string) {
			if m := suffix.FindStringSubmatch(got); m != nil {
				drawn[m[1]]++
			} else {
				drawn[bare]++
			}
		})
		total := sumOfCounts(registered)
		for code, units := range registered {
			t.Run(code, func(t *testing.T) {
				assertShare(t, code+" of the companies", drawn[code], fejktest.ShareDraws, float64(units)/float64(total))
			})
		}
	})
}

func TestLocaleBareCompanyPathDrawsANameEndingInNoLegalForm(t *testing.T) {
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(7))
		suffix, path := l.company.legalFormSuffix(), l.name+".company."+l.company.bare.path
		fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
			if suffix.MatchString(got) {
				t.Fatalf("%s = %q, want a name ending in no legal form", path, got)
			}
		})
	})
}

func TestLocaleSuffixedCompanyPathsDrawANameEndingInTheirLegalForm(t *testing.T) {
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(7))
		for _, form := range l.company.suffixed {
			path := l.name + ".company." + form.path
			t.Run(path, func(t *testing.T) {
				fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
					if !strings.HasSuffix(got, " "+form.suffix) {
						t.Fatalf("%s = %q, want a name ending in %q", path, got, form.suffix)
					}
				})
			})
		}
	})
}

func TestLocaleCompanyLegalIsTheFormItsNameEndsIn(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		suffix := l.company.legalFormSuffix()
		eachCompanyWithItsLegal(t, l.name, func(company, legal string) {
			want := ""
			if m := suffix.FindStringSubmatch(company); m != nil {
				want = m[1]
			}
			if legal != want {
				t.Fatalf("company %q has legal %q, want %q, the form its name ends in", company, legal, want)
			}
		})
	})
}
