package fejkdata

import (
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
