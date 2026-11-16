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
		f := shipped(t, fejkdata.WithSeed(5))
		fejktest.EachDraw(300, func() string {
			return fejktest.FakeTemplate(t, f, "{/"+l.name+".phone as p}{p}|{p.national}|{p.international}")
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
