package fejkdata

import (
	"slices"
	"strconv"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

// Bankinfrastruktur i Sverige's clearing number table, column "IBAN ID".
var swedishIBANIDs = []string{
	"120", "211", "212", "213", "230", "300", "500", "600", "800", "902", "904", "907", "910", "915",
	"917", "918", "919", "923", "925", "927", "928", "930", "939", "940", "946", "947", "950", "955",
	"957", "958", "959", "960", "963", "964", "965", "966", "967", "968", "970", "971", "975", "978",
}

func weightedSum(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	return sum
}

// spanishControl is the control digit of the Banco de España: weights 1 2 4 8 5 10 9 7 3 6.
func spanishControl(ten string) byte {
	switch dc := 11 - weightedSum(ten, []int{1, 2, 4, 8, 5, 10, 9, 7, 3, 6})%11; dc {
	case 10:
		return '1'
	case 11:
		return '0'
	default:
		return byte('0' + dc)
	}
}

func luhnValid(s string) bool {
	sum := 0
	for i := range len(s) {
		d := int(s[len(s)-1-i] - '0')
		if i%2 == 1 {
			if d *= 2; d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// nationalCheckHolds applies the check a country's banks give the BBAN inside its IBAN.
func nationalCheckHolds(iban string) bool {
	bban := iban[4:]
	switch iban[:2] {
	case "BE":
		n, _ := strconv.Atoi(bban[:10])
		want := n % 97
		if want == 0 {
			want = 97
		}
		got, _ := strconv.Atoi(bban[10:])
		return got == want
	case "ES":
		return bban[8] == spanishControl("00"+bban[:8]) && bban[9] == spanishControl(bban[10:])
	case "FI":
		return luhnValid(bban)
	case "NO":
		return weightedSum(bban, []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2, 1})%11 == 0
	case "SE":
		return slices.Contains(swedishIBANIDs, bban[:3])
	}
	return true
}

func TestNationalIBANChecksAcceptStripesTestIBANs(t *testing.T) {
	for _, iban := range []string{
		"BE62510007547061", "ES0700120345030000067890", "FI2112345600000785", "NO9386011117947",
		"SE3550000000054910000003",
	} {
		if !nationalCheckHolds(iban) {
			t.Errorf("the national check refuses %s, a published test IBAN", iban)
		}
	}
}

func TestIBANDrawsPassTheirCountrysNationalCheck(t *testing.T) {
	f := emptyGenerator(1)
	for _, cc := range []string{"BE", "ES", "FI", "NO", "SE"} {
		t.Run(cc, func(t *testing.T) {
			fejktest.EachDraw(500, func() string { return mustRender(t, f, `"{iban(`+cc+`)}"`) }, func(got string) {
				if !ibanValid(got) || !nationalCheckHolds(got) {
					t.Fatalf("iban(%s) = %q fails the mod-97 or the national check", cc, got)
				}
			})
		})
	}
}

func TestSwedishIBANDrawsEveryListedBank(t *testing.T) {
	f := emptyGenerator(1)
	seen := map[string]bool{}
	for range 2000 {
		seen[mustRender(t, f, `"{iban(SE)}"`)[4:7]] = true
	}
	for _, id := range swedishIBANIDs {
		if !seen[id] {
			t.Errorf("iban(SE) drew no IBAN ID %s in 2000 draws", id)
		}
	}
	if len(seen) != len(swedishIBANIDs) {
		t.Errorf("iban(SE) drew %d distinct IBAN IDs, want the %d listed", len(seen), len(swedishIBANIDs))
	}
}
