package builtinfunc

import (
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
)

func derived(name, emitted string) string {
	return Prep(name, nil)(drawstate.New(1, time.Time{}), emitted, nil)
}

func TestIsinCompletesPublishedIdentifiers(t *testing.T) {
	for payload, want := range map[string]string{
		"US037833100": "5",
		"US594918104": "5",
		"GB000263494": "6",
		"DE000BAY001": "7",
		"ZA999999999": "6",
		"AZ000000009": "8",
	} {
		if got := derived("isin", payload); got != want {
			t.Errorf("isin() after %s = %q, want %q", payload, got, want)
		}
	}
}

func TestCusipCompletesPublishedIdentifiers(t *testing.T) {
	for payload, want := range map[string]string{
		"03783310": "0",
		"59491810": "4",
		"93114210": "3",
		"02079K30": "5",
		"Z9A0*@#9": "2",
		"*@#*@#*@": "7",
		"ZZZZZZZZ": "0",
		"A#A#A#A#": "4",
	} {
		if got := derived("cusip", payload); got != want {
			t.Errorf("cusip() after %s = %q, want %q", payload, got, want)
		}
	}
}

func TestAbaCompletesPublishedRoutingNumbers(t *testing.T) {
	for payload, want := range map[string]string{
		"01100001": "5",
		"02100002": "1",
		"99999999": "2",
		"00000000": "0",
	} {
		if got := derived("aba", payload); got != want {
			t.Errorf("aba() after %s = %q, want %q", payload, got, want)
		}
	}
}

func TestCheckDigitsReadOnlyTheTailTheyNeed(t *testing.T) {
	for _, c := range []struct{ name, emitted, want string }{
		{"isin", "ISIN US037833100", "5"},
		{"cusip", "CUSIP 02079K30", "5"},
		{"aba", "ABA 01100001", "5"},
		{"aba", "9 01100001", "5"},
	} {
		if got := derived(c.name, c.emitted); got != c.want {
			t.Errorf("%s() after %q = %q, want %q", c.name, c.emitted, got, c.want)
		}
	}
}

func TestCheckDigitsTakeNoArgument(t *testing.T) {
	for _, name := range []string{"isin", "cusip", "aba"} {
		if err := Check(name, []string{"1"}); err == nil || !strings.Contains(err.Error(), "takes 0 arguments") {
			t.Errorf("Check(%s, 1) = %v, want it to refuse an argument", name, err)
		}
	}
}

const vinAlphabet = "0123456789ABCDEFGHJKLMNPRSTUVWXYZ"

func vinValid(vin string) bool {
	values := map[byte]int{'A': 1, 'B': 2, 'C': 3, 'D': 4, 'E': 5, 'F': 6, 'G': 7, 'H': 8, 'J': 1, 'K': 2, 'L': 3, 'M': 4, 'N': 5, 'P': 7, 'R': 9, 'S': 2, 'T': 3, 'U': 4, 'V': 5, 'W': 6, 'X': 7, 'Y': 8, 'Z': 9}
	if len(vin) != 17 {
		return false
	}
	sum := 0
	for i, w := range []int{8, 7, 6, 5, 4, 3, 2, 10, 0, 9, 8, 7, 6, 5, 4, 3, 2} {
		v, ok := values[vin[i]]
		if !ok {
			v = int(vin[i] - '0')
		}
		sum += v * w
	}
	check := byte('0' + sum%11)
	if sum%11 == 10 {
		check = 'X'
	}
	return vin[8] == check
}

func TestTheVinOracleAcceptsPublishedVins(t *testing.T) {
	for _, vin := range []string{"1M8GDM9AXKP042788", "11111111111111111"} {
		if !vinValid(vin) {
			t.Errorf("vinValid(%s) = false, want a published VIN accepted", vin)
		}
	}
}

func TestVinDrawsAWholeVinOfTheWmiThatPassesItsCheck(t *testing.T) {
	call := Prep("vin", []string{"1M8"})
	s := drawstate.New(1, time.Time{})
	seen := map[string]bool{}
	for range 500 {
		vin := call(s, "", nil)
		if !strings.HasPrefix(vin, "1M8") || !vinValid(vin) {
			t.Fatalf("vin(1M8) = %q, want 17 characters from 1M8 whose position 9 passes", vin)
		}
		if i := strings.IndexFunc(vin, func(r rune) bool { return !strings.ContainsRune(vinAlphabet, r) }); i >= 0 {
			t.Fatalf("vin(1M8) = %q holds %q, which no VIN may hold", vin, vin[i])
		}
		if !strings.ContainsRune("ABCDEFGHJKLMNPRSTVWXY123456789", rune(vin[9])) {
			t.Fatalf("vin(1M8) = %q, want a model year code at position 10", vin)
		}
		seen[vin] = true
	}
	if len(seen) < 400 {
		t.Errorf("%d different VINs in 500 draws, want them to vary", len(seen))
	}
}

func TestVinDrawsAnXWhereTheCheckIsTen(t *testing.T) {
	call := Prep("vin", []string{"1M8"})
	s := drawstate.New(1, time.Time{})
	for range 2000 {
		if vin := call(s, "", nil); vin[8] == 'X' {
			return
		}
	}
	t.Error("no check digit X in 2000 draws, want one where the sum leaves 10")
}

func TestVinRefusesAWmiThatIsNotThreeVinCharacters(t *testing.T) {
	for _, wmi := range []string{"", "1M", "1M8G", "1m8", "1I8", "1O8", "1Q8", "1-8"} {
		if err := Check("vin", []string{wmi}); err == nil || !strings.Contains(err.Error(), "vin(") {
			t.Errorf("Check(vin, %q) = %v, want an error naming vin and what a WMI is", wmi, err)
		}
	}
	if err := Check("vin", []string{"1M8"}); err != nil {
		t.Errorf("Check(vin, 1M8) = %v, want it accepted", err)
	}
}
