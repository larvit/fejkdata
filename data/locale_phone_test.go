package data_test

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func hasPhone(l shippedLocale) bool { return l.phone != nil }

func hasPhoneAreaTables(l shippedLocale) bool { return hasPhone(l) && l.phone.areaTables != nil }

func eachPhone(t *testing.T, locale string, check func(n string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(11))
	fejktest.EachDraw(30000, func() string { return fejktest.Fake(t, f, locale+".phone") }, check)
}

func TestLocalePhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	eachLocale(t, hasPhone, func(t *testing.T, l shippedLocale) {
		isFixed := l.phone.fixed(t)
		var fixed int
		eachPhone(t, l.name, func(n string) {
			if isFixed(n) {
				fixed++
			}
		})
		if fixed < 2580 || fixed > 2880 {
			t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", fixed)
		}
	})
}

func TestLocalePhoneAreaNdcIsItsCodeWithoutTheTrunkZero(t *testing.T) {
	eachLocale(t, hasPhoneAreaTables, func(t *testing.T, l shippedLocale) {
		for _, table := range l.phone.areaTables {
			t.Run(table, func(t *testing.T) {
				for _, row := range fejktest.TSVRows(t, l.name+"/"+table+".tsv") {
					if !strings.HasPrefix(row["area"], "0") || row["ndc"] != row["area"][1:] {
						t.Errorf("%s area %q has ndc %q, want the code without the trunk 0", table, row["area"], row["ndc"])
					}
				}
			})
		}
	})
}

func TestLocalePhoneInternationalIsItsCallingCodeAndItsNationalForm(t *testing.T) {
	eachLocale(t, hasPhone, func(t *testing.T, l shippedLocale) {
		for _, path := range []string{"phone", "mobile", "landline"} {
			t.Run(path, func(t *testing.T) {
				f := shipped(t, fejkdata.WithSeed(5))
				fejktest.EachDraw(300, func() string {
					return fejktest.FakeTemplate(t, f, "{/"+l.name+"."+path+" as p}{p}|{p.national}|{p.international}")
				}, func(got string) {
					parts := strings.Split(got, "|")
					if len(parts) != 3 || parts[0] != parts[1] {
						t.Fatalf("%q: want the national form twice, then the international one", got)
					}
					if parts[2] != l.phone.callingCode+" "+strings.TrimPrefix(parts[0], l.phone.trunkPrefix) {
						t.Fatalf("%q: want %s and the national form behind it, without its trunk prefix %q", got, l.phone.callingCode, l.phone.trunkPrefix)
					}
				})
			})
		}
	})
}

func TestLocaleMobileDrawsOnlyMobileNumbersAndLandlineOnlyFixedOnes(t *testing.T) {
	eachLocale(t, hasPhone, func(t *testing.T, l shippedLocale) {
		isFixed := l.phone.fixed(t)
		shapes := map[string]bool{}
		fp := shipped(t, fejkdata.WithSeed(12))
		fejktest.EachDraw(150000, func() string { return fejktest.Fake(t, fp, l.name+".phone") }, func(n string) { shapes[digitShape(n)] = true })
		f := shipped(t, fejkdata.WithSeed(11))
		for path, wantFixed := range map[string]bool{"mobile": false, "landline": true} {
			fejktest.EachDraw(3000, func() string { return fejktest.Fake(t, f, l.name+"."+path) }, func(n string) {
				if isFixed(n) != wantFixed {
					t.Fatalf("%s.%s drew %q, which is fixed: %v, want %v", l.name, path, n, isFixed(n), wantFixed)
				}
				if !shapes[digitShape(n)] {
					t.Fatalf("%s.%s drew %q, whose grouping and length %q %s.phone never draws", l.name, path, n, digitShape(n), l.name)
				}
			})
		}
	})
}

func digitShape(number string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '9'
		}
		return r
	}, number)
}

func TestNynorskMobileAndLandlineAreBokmalsOwn(t *testing.T) {
	isFixed := norwegianFixedPhone(t)
	f := shipped(t, fejkdata.WithSeed(11))
	for path, wantFixed := range map[string]bool{"mobile": false, "landline": true} {
		fejktest.EachDraw(500, func() string {
			return fejktest.FakeTemplate(t, f, "{/nn_NO."+path+" as p}{p}|{p.international}")
		}, func(got string) {
			national, international, _ := strings.Cut(got, "|")
			if isFixed(national) != wantFixed || international != "+47 "+national {
				t.Fatalf("nn_NO.%s drew %q, want a Norwegian number that is fixed: %v and its +47 form", path, got, wantFixed)
			}
		})
	}
}
