package data_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// The areas of the fixed-line table of the Lithuanian numbering plan that the Communications Regulatory
// Authority (RRT) adopted on 1 September 2023 and published through the ITU on 6 March 2025, each with the
// digits that start every number in it, the area code and any digits the plan fixes after it.
var lithuanianFixedAreas = []string{
	"310", "313", "315", "318", "319", "340", "342", "343", "345", "346", "347", "349",
	"37", "380", "381", "382", "383", "385", "386", "387", "389",
	"41", "421", "422", "425", "426", "427", "428",
	"440", "441", "443", "444", "445", "446", "447", "448", "449", "450", "451",
	"454", "455", "458", "459", "460", "462", "463", "464", "469",
	"520", "521", "522", "523", "524", "525", "526", "527", "528",
}

var lithuanianNational = regexp.MustCompile(`^0 (5 \d{3} \d{4}|(37|41|45|46) \d{2} \d{4}|\d{3} \d{5})$`)

func lithuanianFixedPhone(*testing.T) func(n string) bool {
	return func(n string) bool { return n[2] != '6' }
}

func TestShippedLithuanianPhoneIsGroupedByItsAreaCodeLength(t *testing.T) {
	eachPhone(t, "lt_LT", func(n string) {
		if !lithuanianNational.MatchString(n) {
			t.Fatalf("phone %q is not a 0 and 8 digits grouped by the length of their area code, as Lithuania writes them", n)
		}
	})
}

func TestShippedLithuanianFixedPhoneLiesInAnAreaThePlanAllocates(t *testing.T) {
	eachPhone(t, "lt_LT", func(n string) {
		digits := strings.ReplaceAll(n[2:], " ", "")
		if digits[0] == '6' {
			return
		}
		for _, area := range lithuanianFixedAreas {
			if strings.HasPrefix(digits, area) && len(digits) == 8 {
				return
			}
		}
		t.Fatalf("phone %q starts in no area the plan allocates", n)
	})
}

func TestShippedLithuanianPhoneTableHoldsExactlyTheAllocatedAreas(t *testing.T) {
	var got []string
	for _, row := range fejktest.TSVRows(t, "lt_LT/phone-fixed.tsv") {
		if row["name"] == "" {
			t.Errorf("area %q has no name", row["area"])
		}
		got = append(got, row["area"])
	}
	sort.Strings(got)
	want := append([]string{}, lithuanianFixedAreas...)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("phone-fixed holds %q, want the plan's %q", got, want)
	}
}
