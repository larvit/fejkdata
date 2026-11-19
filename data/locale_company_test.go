package data_test

import (
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

// legalForm returns the abbreviation of the legal form a company name ends in, or opens with, or "".
func (c *company) legalForm(name string) string {
	for _, form := range c.marked {
		if strings.HasSuffix(name, " "+form.abbreviation) || strings.HasPrefix(name, form.abbreviation+" ") {
			return form.abbreviation
		}
	}
	return ""
}

func (c *company) forms() []legalForm { return append([]legalForm{c.bare}, c.marked...) }

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

func TestLocaleCompanyIsANameThatMayCarryALegalForm(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		eachCompany(t, l.name, func(got string) {
			if !l.company.name.MatchString(got) || strings.HasSuffix(got, " "+l.company.bare.code) {
				t.Fatalf("company %q is not a name, optionally carrying a legal form, that doesn't end in %q", got, l.company.bare.code)
			}
		})
	})
}

func TestLocaleCompanyCarriesALegalFormAsOftenAsItsSourceCountsTheForm(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		bare := l.company.bare.code
		registered, drawn := map[string]int{}, map[string]int{}
		for _, form := range l.company.forms() {
			registered[form.code] = form.units
		}
		eachCompany(t, l.name, func(got string) {
			if form := l.company.legalForm(got); form != "" {
				drawn[form]++
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

func TestLocaleBareCompanyPathDrawsANameCarryingNoLegalForm(t *testing.T) {
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(7))
		path := l.name + ".company." + l.company.bare.path
		fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
			if form := l.company.legalForm(got); form != "" {
				t.Fatalf("%s = %q, want a name carrying no legal form", path, got)
			}
		})
	})
}

func TestLocaleMarkedCompanyPathsDrawANameCarryingTheirLegalForm(t *testing.T) {
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(7))
		for _, form := range l.company.marked {
			path := l.name + ".company." + form.path
			t.Run(path, func(t *testing.T) {
				fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
					if l.company.legalForm(got) != form.abbreviation {
						t.Fatalf("%s = %q, want a name carrying %q", path, got, form.abbreviation)
					}
				})
			})
		}
	})
}

func TestLocaleCompanyLegalIsTheFormItsNameCarries(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		eachCompanyWithItsLegal(t, l.name, func(company, legal string) {
			if want := l.company.legalForm(company); legal != want {
				t.Fatalf("company %q has legal %q, want %q, the form its name carries", company, legal, want)
			}
		})
	})
}
