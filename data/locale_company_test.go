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
	for _, form := range c.forms {
		if form.suffix != "" {
			suffixes = append(suffixes, regexp.QuoteMeta(form.suffix))
		}
	}
	return regexp.MustCompile(` (` + strings.Join(suffixes, "|") + `)$`)
}

func (c *company) bareForm() legalForm {
	for _, form := range c.forms {
		if form.suffix == "" {
			return form
		}
	}
	panic("no legal form without a suffix")
}

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
		bare := " " + l.company.bareForm().code
		eachCompany(t, l.name, func(got string) {
			if !l.company.name.MatchString(got) || strings.HasSuffix(got, bare) {
				t.Fatalf("company %q is not a name that may end in a legal form, which%s is not", got, bare)
			}
		})
	})
}

func TestLocaleCompanyEndsInALegalFormAsOftenAsItsSourceCountsTheForm(t *testing.T) {
	t.Parallel()
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		t.Parallel()
		suffix, bare := l.company.legalFormSuffix(), l.company.bareForm().code
		registered, drawn := map[string]int{}, map[string]int{}
		for _, form := range l.company.forms {
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

func TestLocaleCompanyPathsDrawOneFormEach(t *testing.T) {
	eachLocale(t, hasCompany, func(t *testing.T, l shippedLocale) {
		f := shipped(t, fejkdata.WithSeed(7))
		suffix := l.company.legalFormSuffix()
		for _, form := range l.company.forms {
			path := l.name + ".company." + form.path
			t.Run(path, func(t *testing.T) {
				fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, path) }, func(got string) {
					if form.suffix == "" && suffix.MatchString(got) || form.suffix != "" && !strings.HasSuffix(got, " "+form.suffix) {
						t.Fatalf("%s = %q, want a name ending in %q, or in no legal form where that is empty", path, got, form.suffix)
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
