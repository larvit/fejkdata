package fejkdata

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestValidIsin(t *testing.T) {
	for s, want := range map[string]bool{
		"US0378331005": true,
		"US5949181045": true,
		"GB0002634946": true,
		"DE000BAY0017": true,
		"US0378331004": false,
		"US037833100":  false,
		"us0378331005": false,
	} {
		t.Run(s, func(t *testing.T) {
			if got := fejktest.ValidIsin(s); got != want {
				t.Errorf("ValidIsin(%q) = %v, want %v", s, got, want)
			}
		})
	}
}

func TestValidCusip(t *testing.T) {
	for s, want := range map[string]bool{
		"037833100": true,
		"594918104": true,
		"931142103": true,
		"02079K305": true,
		"Z9A0*@#92": true,
		"*@#*@#*@7": true,
		"ZZZZZZZZ0": true,
		"A#A#A#A#4": true,
		"037833101": false,
		"03783310":  false,
		"03783310A": false,
		"Z9A0*@#93": false,
		"*@#*@#*@8": false,
	} {
		t.Run(s, func(t *testing.T) {
			if got := fejktest.ValidCusip(s); got != want {
				t.Errorf("ValidCusip(%q) = %v, want %v", s, got, want)
			}
		})
	}
}

func TestShippedMiscIsin(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	territories := map[string]bool{}
	for _, row := range fejktest.TSVRows(t, "data/misc/territory.tsv") {
		territories[row["alpha2"]] = true
	}
	countries, letters, digits := map[string]bool{}, false, false
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, "isin") }, func(v string) {
		if !fejktest.ValidIsin(v) {
			t.Fatalf("misc isin %q is not a valid ISIN", v)
		}
		if !territories[v[:2]] {
			t.Fatalf("misc isin %q starts with %q, which misc.territory has no row for", v, v[:2])
		}
		countries[v[:2]] = true
		for _, c := range v[2:11] {
			letters = letters || unicode.IsLetter(c)
			digits = digits || unicode.IsDigit(c)
		}
	})
	if len(countries) < 20 || !letters || !digits {
		t.Fatalf("500 ISINs held %d country codes, letters %v, digits %v", len(countries), letters, digits)
	}
}

func TestShippedMiscCusip(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	letters, digits := false, false
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, "cusip") }, func(v string) {
		if !fejktest.ValidCusip(v) {
			t.Fatalf("misc cusip %q is not a valid CUSIP", v)
		}
		letters = letters || strings.ContainsAny(v[:8], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		digits = digits || strings.ContainsAny(v[:8], "0123456789")
	})
	if !letters || !digits {
		t.Fatalf("500 CUSIPs held letters %v, digits %v", letters, digits)
	}
}

func TestValidBic(t *testing.T) {
	for s, want := range map[string]bool{
		"DEUTDEFF":     true,
		"DEUTDEFF500":  true,
		"NDEAFIHH":     true,
		"CHASUS33XXX":  true,
		"DEUTDEF0":     false,
		"DEUTDEF1":     false,
		"DEUTDEF2":     false,
		"DEUTDEFF50":   false,
		"deutdeff":     false,
		"DEU1DEFF":     false,
		"DEUTD1FF":     false,
		"DEUTDEFF5000": false,
	} {
		t.Run(s, func(t *testing.T) {
			if got := fejktest.ValidBic(s); got != want {
				t.Errorf("ValidBic(%q) = %v, want %v", s, got, want)
			}
		})
	}
}

func TestShippedMiscBic(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	territories := map[string]bool{}
	for _, row := range fejktest.TSVRows(t, "data/misc/territory.tsv") {
		territories[row["alpha2"]] = true
	}
	countries, lengths := map[string]bool{}, map[int]bool{}
	kinds := map[string]bool{}
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "bic") }, func(v string) {
		if !fejktest.ValidBic(v) {
			t.Fatalf("misc bic %q is not a valid BIC", v)
		}
		if !territories[v[4:6]] {
			t.Fatalf("misc bic %q holds %q, which misc.territory has no row for", v, v[4:6])
		}
		countries[v[4:6]] = true
		lengths[len(v)] = true
		for _, i := range []int{6, 7} {
			kinds[fmt.Sprintf("location %d %s", i-5, charKind(v[i]))] = true
		}
		if len(v) == 11 && v[8:] == "XXX" {
			kinds["branch XXX"] = true
		} else if len(v) == 11 {
			for _, i := range []int{8, 9, 10} {
				kinds[fmt.Sprintf("branch %d %s", i-7, charKind(v[i]))] = true
			}
		}
	})
	for _, want := range []string{
		"location 1 digit", "location 1 letter", "location 2 digit", "location 2 letter",
		"branch 1 digit", "branch 1 letter", "branch 2 digit", "branch 2 letter",
		"branch 3 digit", "branch 3 letter", "branch XXX",
	} {
		if !kinds[want] {
			t.Errorf("2000 BICs held no %q", want)
		}
	}
	if len(countries) < 20 || !lengths[8] || !lengths[11] {
		t.Fatalf("2000 BICs held %d country codes and lengths %v", len(countries), lengths)
	}
}

func charKind(c byte) string {
	if c >= '0' && c <= '9' {
		return "digit"
	}
	return "letter"
}
