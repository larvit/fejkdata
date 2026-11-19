package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// The numbers of the fixed-line table of the Lithuanian numbering plan that the Communications Regulatory
// Authority (RRT) adopted on 1 September 2023 and the ITU published on 6 March 2025: the area codes the
// table doesn't mark "Reserved", each with its city or district.
var lithuanianAreas = []struct{ code, name string }{
	{"5", "Vilnius"}, {"37", "Kaunas"}, {"41", "Šiauliai"}, {"45", "Panevėžys"}, {"46", "Klaipėda"},
	{"310", "Varėna"}, {"313", "Druskininkai"}, {"315", "Alytus"}, {"318", "Lazdijai"},
	{"319", "Prienai and Birštonas"}, {"340", "Ukmergė"}, {"342", "Vilkaviškis"}, {"343", "Marijampolė"},
	{"345", "Šakiai"}, {"346", "Kaišiadorys"}, {"347", "Kėdainiai"}, {"349", "Jonava"},
	{"380", "Šalčininkai"}, {"381", "Anykščiai"}, {"382", "Širvintos"}, {"383", "Molėtai"},
	{"385", "Zarasai"}, {"386", "Ignalina and Visaginas"}, {"387", "Švenčionys"}, {"389", "Utena"},
	{"421", "Pakruojis"}, {"422", "Radviliškis"}, {"425", "Akmenė"}, {"426", "Joniškis"},
	{"427", "Kelmė"}, {"428", "Raseiniai"}, {"440", "Skuodas"}, {"441", "Šilutė"}, {"443", "Mažeikiai"},
	{"444", "Telšiai"}, {"445", "Kretinga"}, {"446", "Tauragė"}, {"447", "Jurbarkas"}, {"448", "Plungė"},
	{"449", "Šilalė"}, {"450", "Biržai"}, {"451", "Pasvalys"}, {"458", "Rokiškis"}, {"459", "Kupiškis"},
	{"460", "Palanga"}, {"469", "Neringa"}, {"528", "Trakai"},
}

// The digits that open every fixed number: an area code, and where the table fixes them, the first digits
// of its subscriber numbers.
var lithuanianFixedPrefixes = []string{
	"310", "313", "315", "318", "319", "340", "342", "343", "345", "346", "347", "349",
	"37", "380", "381", "382", "383", "385", "386", "387", "389",
	"41", "421", "422", "425", "426", "427", "428",
	"440", "441", "443", "444", "445", "446", "447", "448", "449", "450", "451",
	"454", "455", "458", "459", "460", "462", "463", "464", "469",
	"520", "521", "522", "523", "524", "525", "526", "527", "528",
}

// A national number is the trunk 0, then the area code and the subscriber number: Vilnius's 5 and 7 digits
// grouped 3 and 4, a two-digit code and 6 digits grouped 2 and 4, a three-digit code and 5 digits, and
// the mobile 6 with 7 digits grouped 3 and 5 whole.
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
		for _, prefix := range lithuanianFixedPrefixes {
			if strings.HasPrefix(digits, prefix) && len(digits) == 8 {
				return
			}
		}
		t.Fatalf("phone %q doesn't start in any area the plan allocates", n)
	})
}

func TestShippedLithuanianFixedTableHoldsExactlyTheAllocatedAreas(t *testing.T) {
	want := map[string]string{}
	for _, a := range lithuanianAreas {
		want["0"+a.code] = a.name
	}
	got := map[string]string{}
	for _, row := range fejktest.TSVRows(t, "lt_LT/phone-fixed.tsv") {
		got[row["area"]] = row["name"]
	}
	if len(got) != len(want) {
		t.Errorf("phone-fixed holds %d areas, want the plan's %d", len(got), len(want))
	}
	for area, name := range want {
		if got[area] != name {
			t.Errorf("phone-fixed[%s] is %q, want %q", area, got[area], name)
		}
	}
}

func TestShippedLithuanianMobileTableHoldsTheOneMobileRange(t *testing.T) {
	rows := fejktest.TSVRows(t, "lt_LT/phone-mobile.tsv")
	if len(rows) != 1 || rows[0]["area"] != "06" {
		t.Errorf("phone-mobile holds %v, want the one row 06", rows)
	}
}
