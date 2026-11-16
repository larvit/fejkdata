package data_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

var germanLegalFormSuffix = regexp.MustCompile(` (KG|OHG|GmbH|UG \(haftungsbeschränkt\)|AG)$`)

func germanLegalForm(company string) string {
	if m := germanLegalFormSuffix.FindStringSubmatch(company); m != nil {
		return m[1]
	}
	return "Einzelunternehmen"
}

func eachShippedGermanCompany(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(6))
	fejktest.EachDraw(fejktest.ShareDraws, func() string { return fejktest.Fake(t, f, "de_DE.company") }, check)
}

func TestShippedGermanCompanyIsANameThatMayEndInALegalForm(t *testing.T) {
	t.Parallel()
	name := regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (KG|OHG|GmbH|UG \(haftungsbeschränkt\)|AG))?$`)
	eachShippedGermanCompany(t, func(got string) {
		if !name.MatchString(got) || strings.HasSuffix(got, " Einzelunternehmen") {
			t.Fatalf("company %q is not a name that may end in a legal form", got)
		}
	})
}

// Destatis' Unternehmensregister, 2024: 2,062,752 sole proprietors, 440,951 partnerships and 861,720
// corporations, each group split evenly over the forms drawn.
func TestShippedGermanCompanyEndsInALegalFormDrawnByDestatisLegalUnits(t *testing.T) {
	t.Parallel()
	registered := map[string]int{"Einzelunternehmen": 2062752, "KG": 220476, "OHG": 220475, "GmbH": 287240, "UG (haftungsbeschränkt)": 287240, "AG": 287240}
	total := sumOfCounts(registered)
	form := map[string]int{}
	eachShippedGermanCompany(t, func(got string) { form[germanLegalForm(got)]++ })
	for code, n := range registered {
		t.Run(code, func(t *testing.T) {
			assertShare(t, code+" of the companies", form[code], fejktest.ShareDraws, float64(n)/float64(total))
		})
	}
}

func TestShippedGermanCompanyPathsDrawOneFormEach(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(7))
	for path, suffix := range map[string]string{"einzel": "", "kg": " KG", "ohg": " OHG", "gmbh": " GmbH", "ug": " UG (haftungsbeschränkt)", "ag": " AG"} {
		t.Run(path, func(t *testing.T) {
			fejktest.EachDraw(100, func() string { return fejktest.Fake(t, f, "de_DE.company."+path) }, func(got string) {
				if suffix == "" && germanLegalFormSuffix.MatchString(got) || !strings.HasSuffix(got, suffix) {
					t.Fatalf("de_DE.company.%s = %q, want a name ending in %q", path, got, suffix)
				}
			})
		})
	}
}

func TestShippedGermanCompanyLegalIsItsAbbreviation(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(7))
	fejktest.EachDraw(200, func() string { return fejktest.FakeTemplate(t, f, "{/de_DE.company as c}{c}|{c.legal}") }, func(got string) {
		company, legal, _ := strings.Cut(got, "|")
		if legal != "" && !strings.HasSuffix(company, " "+legal) || legal == "" && germanLegalFormSuffix.MatchString(company) {
			t.Fatalf("company %q has legal %q", company, legal)
		}
	})
}

func TestShippedGermanCompaniesAreMostlyDistinct(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(6))
	seen := fejktest.TallyDraws(500, func() string { return fejktest.Fake(t, f, "de_DE.company") })
	if len(seen) < 400 {
		t.Errorf("500 companies drew %d distinct, want 400 or more", len(seen))
	}
}

func eachShippedGermanDate(t *testing.T, check func(got string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, "de_DE.date") }, check)
}

// DIN 5008 writes a date as TT.MM.JJJJ.
func TestShippedGermanDateIsDayMonthYearWithDotsAndLeadingZerosFrom1970To2029(t *testing.T) {
	date := regexp.MustCompile(`^(0[1-9]|[12]\d|3[01])\.(0[1-9]|1[0-2])\.\d{4}$`)
	years := map[int]bool{}
	eachShippedGermanDate(t, func(got string) {
		d, err := time.Parse("02.01.2006", got)
		if err != nil || !date.MatchString(got) || d.Year() < 1970 || d.Year() > 2029 {
			t.Fatalf("date %q is not dd.mm.yyyy in 1970 to 2029: %v", got, err)
		}
		years[d.Year()] = true
	})
	if !years[1970] || !years[2029] {
		t.Errorf("500 dates drew the years %v, want 1970 and 2029 among them", years)
	}
}

func TestShippedGermanEmailIsAsciiAtADomainNobodyCanRegister(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, "de_DE.email") }, func(got string) {
		if !exampleEmail.MatchString(got) {
			t.Fatalf("email %q is not lower-case ASCII at example.com, example.org or example.net", got)
		}
	})
}

func TestShippedGermanEmailsAreMostlyDistinctAndFoldUmlauts(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	seen := fejktest.TallyDraws(5000, func() string { return fejktest.Fake(t, f, "de_DE.email") })
	if len(seen) < 4000 {
		t.Errorf("5000 emails drew %d distinct, want 4000 or more", len(seen))
	}
	folded := false
	for got := range seen {
		folded = folded || strings.Contains(got, "muller") || strings.Contains(got, "schroder") || strings.Contains(got, "konig")
	}
	if !folded {
		t.Error("5000 emails held no Müller, Schröder or König folded to ASCII, want one")
	}
}
