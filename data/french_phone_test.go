package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// The roots of "Numéros pour œuvres audiovisuelles" in section 2.5 of the Arcep's plan national de
// numérotation, version of 1 January 2026, which may be used as telephone numbers in fiction. The
// plan allots no other number to anyone, and the roots are in the same grouping as a French number.
var (
	frenchFictionFixedRoots  = []string{"01 99 00", "02 61 91", "03 53 01", "04 65 71", "05 36 49"}
	frenchFictionMobileRoots = []string{"06 39 98"}
	frenchNational           = regexp.MustCompile(`^0[1-6]( \d{2}){4}$`)
)

func frenchFixedPhone(*testing.T) func(n string) bool {
	return func(n string) bool { return n[1] != '6' }
}

func TestShippedFrenchPhoneIsTenDigitsGroupedInTwos(t *testing.T) {
	eachPhone(t, "fr_FR", func(n string) {
		if !frenchNational.MatchString(n) {
			t.Fatalf("phone %q is not a 0, a digit from 1 to 6 and 8 more digits, grouped in twos", n)
		}
	})
}

func TestShippedFrenchPhoneLiesInARootTheArcepKeepsForFiction(t *testing.T) {
	roots := map[string]bool{}
	for _, r := range append(append([]string{}, frenchFictionFixedRoots...), frenchFictionMobileRoots...) {
		roots[r] = true
	}
	eachPhone(t, "fr_FR", func(n string) {
		if !roots[n[:8]] {
			t.Fatalf("phone %q lies outside the roots the plan keeps for fiction", n)
		}
	})
}

func TestShippedFrenchPhoneTablesHoldExactlyTheFictionRoots(t *testing.T) {
	for kind, want := range map[string][]string{"fixed": frenchFictionFixedRoots, "mobile": frenchFictionMobileRoots} {
		t.Run(kind, func(t *testing.T) {
			var got []string
			for _, row := range fejktest.TSVRows(t, "fr_FR/phone-"+kind+".tsv") {
				got = append(got, row["area"])
			}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("phone-%s holds %q, want the plan's %q", kind, got, want)
			}
		})
	}
}
