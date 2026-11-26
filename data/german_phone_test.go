package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The 81 blocks of "Zugeteilte Rufnummernblöcke (Mobile Dienste)", the Bundesnetzagentur's list
// RNB_csv.csv, without the trunk 0.
var germanMobileBlocks = strings.Fields(`
	15019 15020 1511 1512 1514 1515 1516 1517 15180 15181 15182 15183 15184 15185 15186 15187 15188
	15189 15310 15333 1520 1521 1522 1523 1525 1526 1529 15410 15510 15511 15550 15551 15552 15553
	15554 15555 15556 15557 15558 15559 15560 15561 15562 15563 15564 15565 15566 15567 15568 15569 15591
	15630 15636 15678 15679 15700 15701 15702 15703 15704 15706 1573 1575 1577 1578 1579 15888 1590
	160 162 163 170 171 172 173 174 175 176 177 178 179`)

// The Ortsnetzkennzahlen of the Bundesnetzagentur's Vorwahlverzeichnis of 2022-07-27 that it marks
// active, by the length of the code without its 0: Berlin, Hamburg, Frankfurt and München have two digits.
var germanAreaCodesByLength = map[int]int{2: 4, 3: 85, 4: 3882, 5: 1229}

func germanPhoneRows(t *testing.T, kind string) []map[string]string {
	t.Helper()
	return fejktest.TSVRows(t, "de_DE/phone-"+kind+"-prefix.tsv")
}

func TestShippedGermanPhoneMobilePrefixHoldsExactlyTheBlocksTheBundesnetzagenturListsAssigned(t *testing.T) {
	got := map[string]bool{}
	for _, row := range germanPhoneRows(t, "mobile") {
		if got[row["area"]] {
			t.Errorf("phone-mobile-prefix repeats %q", row["area"])
		}
		got[row["area"]] = true
	}
	for _, block := range germanMobileBlocks {
		if !got["0"+block] {
			t.Errorf("phone-mobile-prefix lacks 0%s", block)
		}
	}
	if len(got) != len(germanMobileBlocks) {
		t.Errorf("phone-mobile-prefix holds %d blocks, the list assigns %d", len(got), len(germanMobileBlocks))
	}
}

func TestShippedGermanPhoneFixedPrefixHoldsEveryActiveOrtsnetzkennzahl(t *testing.T) {
	byLength, seen := map[int]int{}, map[string]bool{}
	for _, row := range germanPhoneRows(t, "fixed") {
		if seen[row["area"]] {
			t.Errorf("phone-fixed-prefix repeats %q", row["area"])
		}
		seen[row["area"]] = true
		byLength[len(row["area"])-1]++
	}
	for length, want := range germanAreaCodesByLength {
		if byLength[length] != want {
			t.Errorf("phone-fixed-prefix holds %d codes of %d digits, the directory lists %d active", byLength[length], length, want)
		}
	}
	for _, inactive := range []string{"033052", "033927"} {
		if seen[inactive] {
			t.Errorf("phone-fixed-prefix holds %s, which the directory marks inactive", inactive)
		}
	}
}

func TestShippedGermanPhoneAreaSelectsOneRowByItsCode(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(1))
	for path, want := range map[string]string{
		"de_DE.phone-fixed-prefix[030].name": "Berlin", "de_DE.phone-fixed-prefix[069].name": "Frankfurt am Main",
		"de_DE.phone-fixed-prefix[0201].name": "Essen", "de_DE.phone-fixed-prefix[033051].name": "Nassenheide",
		"de_DE.phone-mobile-prefix[01511].ndc": "1511", "de_DE.phone-mobile-prefix[015550].ndc": "15550",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("printed %q, want %q", got, want)
			}
		})
	}
}

var germanNationalPhone = regexp.MustCompile(`^(0\d{2,5}) ([2-9]\d{5,7})$`)

func germanAreaCodes(t *testing.T, kind string) map[string]bool {
	t.Helper()
	got := map[string]bool{}
	for _, row := range germanPhoneRows(t, kind) {
		got[row["area"]] = true
	}
	return got
}

func germanFixedPhone(t *testing.T) func(n string) bool {
	t.Helper()
	fixed := germanAreaCodes(t, "fixed")
	return func(n string) bool { return fixed[strings.Fields(n)[0]] }
}

func TestShippedGermanPhoneIsAnAreaCodeAndASubscriberNumberWithoutAZeroOrAOneFirst(t *testing.T) {
	eachPhone(t, "de_DE", func(n string) {
		if !germanNationalPhone.MatchString(n) {
			t.Fatalf("phone %q is not a code, a space and a subscriber number of 6 to 8 digits not starting with 0 or 1", n)
		}
	})
}

// Verfügung 25/2006, section 2.3: a number allocated now is ten digits long in the four
// two-digit areas and eleven elsewhere. Nummernplan Mobile Dienste, sections 2.3 and 2.4: eleven
// for a block of four or five digits, ten or eleven for one of three.
func TestShippedGermanPhoneIsAsLongAsThePlansGive(t *testing.T) {
	fixed, mobile := germanAreaCodes(t, "fixed"), germanAreaCodes(t, "mobile")
	lengths := map[string]map[int]bool{}
	eachPhone(t, "de_DE", func(n string) {
		m := germanNationalPhone.FindStringSubmatch(n)
		if m == nil {
			return
		}
		area, total := m[1], len(m[1])-1+len(m[2])
		want := map[int]bool{11: true}
		switch {
		case !fixed[area] && !mobile[area]:
			t.Fatalf("phone %q has area code %s, which no table holds", n, area)
		case fixed[area] && len(area) == 3:
			want = map[int]bool{10: true}
		case mobile[area] && len(area) == 4:
			want = map[int]bool{10: true, 11: true}
		}
		if !want[total] {
			t.Fatalf("phone %q has %d digits, the plans give %v", n, total, want)
		}
		if lengths[area] == nil {
			lengths[area] = map[int]bool{}
		}
		lengths[area][total] = true
	})
	for _, area := range []string{"0170", "0176", "0160"} {
		if !lengths[area][10] || !lengths[area][11] {
			t.Errorf("block %s drew lengths %v, want both 10 and 11", area, lengths[area])
		}
	}
}
