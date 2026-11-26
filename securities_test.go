package fejkdata

import (
	"strings"
	"testing"
	"unicode"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestSecurityCheckersAcceptPublishedIdentifiers(t *testing.T) {
	for _, s := range []string{"US0378331005", "US5949181045", "GB0002634946", "DE000BAY0017"} {
		if !fejktest.ValidIsin(s) {
			t.Errorf("ValidIsin(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"US0378331004", "US037833100", "us0378331005"} {
		if fejktest.ValidIsin(s) {
			t.Errorf("ValidIsin(%q) = true, want false", s)
		}
	}
	for _, s := range []string{"037833100", "594918104", "931142103", "02079K305", "Z9A0*@#92", "*@#*@#*@7", "ZZZZZZZZ0", "A#A#A#A#4"} {
		if !fejktest.ValidCusip(s) {
			t.Errorf("ValidCusip(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"037833101", "03783310", "03783310A", "Z9A0*@#93", "*@#*@#*@8"} {
		if fejktest.ValidCusip(s) {
			t.Errorf("ValidCusip(%q) = true, want false", s)
		}
	}
}

func TestShippedMiscIsin(t *testing.T) {
	f := newBareModule(t, "misc", WithSeed(1))
	countries, letters, digits := map[string]bool{}, false, false
	fejktest.EachDraw(500, func() string { return fejktest.Fake(t, f, "isin") }, func(v string) {
		if !fejktest.ValidIsin(v) {
			t.Fatalf("misc isin %q is not a valid ISIN", v)
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
