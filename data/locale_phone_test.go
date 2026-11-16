package data_test

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func hasPhone(l shippedLocale) bool { return l.phone != nil }

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
			behind := parts[0]
			if l.phone.dropsTrunkPrefix {
				behind = strings.TrimPrefix(behind, "0")
			}
			if parts[2] != l.phone.callingCode+" "+behind {
				t.Fatalf("%q: want %s and the national form behind it, without its 0: %t", got, l.phone.callingCode, l.phone.dropsTrunkPrefix)
			}
		})
	})
}
