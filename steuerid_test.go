package fejkdata

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestValidSteuerIDAcceptsElstersExamples(t *testing.T) {
	for _, id := range []string{"02476291358", "86095742719", "47036892816", "65929970489", "57549285017", "25768131411"} {
		if !fejktest.ValidSteuerID(id) {
			t.Errorf("%s is an example in ELSTER's Prüfung der Steuer- und Steueridentifikationsnummer, want it valid", id)
		}
	}
}

func TestValidSteuerIDRefusesWhatTheRulesForbid(t *testing.T) {
	for name, id := range map[string]string{
		"a wrong check digit":         "86095742710",
		"ten digits":                  "8609574271",
		"twelve digits":               "860957427199",
		"a letter":                    "8609574271x",
		"no digit twice":              "01234567896",
		"three equal digits in a row": "11134567891",
		"four equal digits":           "11211345674",
		"two digits twice":            "11223456785",
	} {
		t.Run(name, func(t *testing.T) {
			if fejktest.ValidSteuerID(id) {
				t.Errorf("%s breaks the rules for a Steueridentifikationsnummer, want it refused", id)
			}
		})
	}
}

func steuerIDs(t *testing.T, seed uint64, n int) []string {
	t.Helper()
	f := emptyGenerator(seed)
	return fejktest.DrawsInTurn(n, func() string { return mustRender(t, f, `{"format":"{steuerid()}"}`) })
}

func TestBuiltinSteuerIDDrawsAValidTestNumberWithALeadingZero(t *testing.T) {
	for _, got := range steuerIDs(t, 1, 5000) {
		if !fejktest.ValidSteuerID(got) || got[0] != '0' {
			t.Fatalf("steuerid() = %q, want a valid number that starts with the 0 ELSTER keeps for test numbers", got)
		}
	}
}

func mostEqualDigits(id string) int {
	most := 0
	for _, d := range id[:10] {
		most = max(most, strings.Count(id[:10], string(d)))
	}
	return most
}

func TestBuiltinSteuerIDDrawsADigitTwiceAndThreeTimes(t *testing.T) {
	twice, thrice := 0, 0
	for _, got := range steuerIDs(t, 2, 2000) {
		if mostEqualDigits(got) == 3 {
			thrice++
		} else {
			twice++
		}
	}
	if twice < 500 || thrice < 500 {
		t.Errorf("steuerid() drew a digit twice %d times and three times %d times in 2000, want both common", twice, thrice)
	}
}

func TestBuiltinSteuerIDVariesEveryPosition(t *testing.T) {
	var seen [11]map[byte]bool
	for i := range seen {
		seen[i] = map[byte]bool{}
	}
	for _, got := range steuerIDs(t, 3, 2000) {
		for i := range got {
			seen[i][got[i]] = true
		}
	}
	for i := 1; i < 11; i++ {
		if len(seen[i]) != 10 {
			t.Errorf("steuerid() drew %d different digits at position %d in 2000, want all ten", len(seen[i]), i+1)
		}
	}
}

func TestBuiltinSteuerIDLeavesOutEveryDigitInTurn(t *testing.T) {
	var leftOut [10]bool
	for _, got := range steuerIDs(t, 5, 2000) {
		for d := range 10 {
			if !strings.Contains(got[:10], string(rune('0'+d))) {
				leftOut[d] = true
			}
		}
	}
	for d := 1; d < 10; d++ {
		if !leftOut[d] {
			t.Errorf("steuerid() never left out the digit %d from the first ten in 2000, want any digit but the 0 left out", d)
		}
	}
}

func TestBuiltinSteuerIDRepeatsWithItsSeed(t *testing.T) {
	a, b := steuerIDs(t, 4, 50), steuerIDs(t, 4, 50)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("steuerid() drew %q then %q from one seed, want the same number", a[i], b[i])
		}
	}
}
