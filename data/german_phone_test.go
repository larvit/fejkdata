package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The 80 blocks of "Zugeteilte Rufnummernblöcke (Mobile Dienste)", the Bundesnetzagentur's list
// RNB_csv.csv, without the trunk 0. The list writes 15591 with a space where it has a semicolon.
var germanMobileBlocks = strings.Fields(`
	15019 15020 1511 1512 1514 1515 1516 1517 15180 15181 15182 15183 15184 15185 15186 15187 15188
	15189 15310 15333 1520 1521 1522 1523 1525 1526 1529 15410 15510 15511 15550 15551 15552 15553
	15554 15555 15556 15557 15558 15559 15560 15561 15562 15563 15564 15565 15566 15567 15568 15569
	15630 15636 15678 15679 15700 15701 15702 15703 15704 15706 1573 1575 1577 1578 1579 15888 1590
	15591 160 162 163 170 171 172 173 174 175 176 177 178 179`)

// The Ortsnetzkennzahlen of the Bundesnetzagentur's Vorwahlverzeichnis of 2022-07-27 that it marks
// active, by length: its Berlin, Hamburg, Frankfurt and München; then three, four and five digits.
var germanAreaCodesByLength = map[int]int{2: 4, 3: 85, 4: 3882, 5: 1229}

func germanPhoneRows(t *testing.T, kind string) []map[string]string {
	t.Helper()
	return fejktest.TSVRows(t, "de_DE/phone-"+kind+".tsv")
}

func TestShippedGermanPhoneMobileHoldsExactlyTheBlocksTheBundesnetzagenturListsAssigned(t *testing.T) {
	got := map[string]bool{}
	for _, row := range germanPhoneRows(t, "mobile") {
		if got[row["area"]] {
			t.Errorf("phone-mobile repeats %q", row["area"])
		}
		got[row["area"]] = true
	}
	for _, block := range germanMobileBlocks {
		if !got["0"+block] {
			t.Errorf("phone-mobile lacks 0%s", block)
		}
	}
	if len(got) != len(germanMobileBlocks) {
		t.Errorf("phone-mobile holds %d blocks, the list assigns %d", len(got), len(germanMobileBlocks))
	}
}

func TestShippedGermanPhoneFixedHoldsEveryActiveOrtsnetzkennzahl(t *testing.T) {
	byLength, seen := map[int]int{}, map[string]bool{}
	for _, row := range germanPhoneRows(t, "fixed") {
		if seen[row["area"]] {
			t.Errorf("phone-fixed repeats %q", row["area"])
		}
		seen[row["area"]] = true
		byLength[len(row["area"])-1]++
	}
	for length, want := range germanAreaCodesByLength {
		if byLength[length] != want {
			t.Errorf("phone-fixed holds %d codes of %d digits, the directory lists %d active", byLength[length], length, want)
		}
	}
	for _, inactive := range []string{"033052", "033927"} {
		if seen[inactive] {
			t.Errorf("phone-fixed holds %s, which the directory marks inactive", inactive)
		}
	}
}

func TestShippedGermanPhoneAreaNdcIsItsCodeWithoutTheTrunkZero(t *testing.T) {
	for _, kind := range []string{"fixed", "mobile"} {
		t.Run(kind, func(t *testing.T) {
			for _, row := range germanPhoneRows(t, kind) {
				if row["ndc"] != strings.TrimPrefix(row["area"], "0") || row["area"][0] != '0' {
					t.Errorf("phone-%s area %q has ndc %q, want the code without the trunk 0", kind, row["area"], row["ndc"])
				}
			}
		})
	}
}

func TestShippedGermanPhoneAreaSelectsOneRowByItsCode(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(1))
	for path, want := range map[string]string{
		"de_DE.phone-fixed[030].name": "Berlin", "de_DE.phone-fixed[069].name": "Frankfurt am Main",
		"de_DE.phone-fixed[0201].name": "Essen", "de_DE.phone-fixed[033051].name": "Nassenheide",
		"de_DE.phone-mobile[0151].ndc": "151", "de_DE.phone-mobile[015550].ndc": "15550",
	} {
		t.Run(path, func(t *testing.T) {
			if got := fejktest.Fake(t, f, path); got != want {
				t.Errorf("printed %q, want %q", got, want)
			}
		})
	}
}

var germanNationalPhone = regexp.MustCompile(`^(0\d{2,5}) ([1-9]\d{5,7})$`)

func eachShippedGermanPhone(t *testing.T, check func(n string)) {
	t.Helper()
	f := shipped(t, fejkdata.WithSeed(11))
	fejktest.EachDraw(30000, func() string { return fejktest.Fake(t, f, "de_DE.phone") }, check)
}

func germanAreaCodes(t *testing.T, kind string) map[string]bool {
	t.Helper()
	got := map[string]bool{}
	for _, row := range germanPhoneRows(t, kind) {
		got[row["area"]] = true
	}
	return got
}

func TestShippedGermanPhoneIsAnAreaCodeAndASubscriberNumberWithoutAZeroFirst(t *testing.T) {
	eachShippedGermanPhone(t, func(n string) {
		if !germanNationalPhone.MatchString(n) {
			t.Fatalf("phone %q is not a code, a space and a subscriber number of 6 to 8 digits not starting with 0", n)
		}
	})
}

// Verfügung 25/2006, section 2.3: a number allocated now is ten digits long in the four
// two-digit areas and eleven elsewhere. Nummernplan Mobile Dienste, sections 2.3 and 2.4: eleven
// for a block of four or five digits, ten or eleven for one of three.
func TestShippedGermanPhoneIsAsLongAsThePlansGive(t *testing.T) {
	fixed, mobile := germanAreaCodes(t, "fixed"), germanAreaCodes(t, "mobile")
	lengths := map[string]map[int]bool{}
	eachShippedGermanPhone(t, func(n string) {
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

func TestShippedGermanPhoneIsMobileTenTimesAsOftenAsFixed(t *testing.T) {
	fixed := germanAreaCodes(t, "fixed")
	var nFixed int
	eachShippedGermanPhone(t, func(n string) {
		if fixed[strings.Fields(n)[0]] {
			nFixed++
		}
	})
	if nFixed < 2580 || nFixed > 2880 {
		t.Fatalf("30000 phones drew %d fixed, want about an eleventh, 2727 give or take 150", nFixed)
	}
}

func TestShippedGermanPhoneInternationalDropsTheTrunkZeroBehindTheCountryCode(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	fejktest.EachDraw(300, func() string {
		return fejktest.FakeTemplate(t, f, "{/de_DE.phone as p}{p}|{p.national}|{p.international}")
	}, func(got string) {
		parts := strings.Split(got, "|")
		if len(parts) != 3 || parts[0] != parts[1] || parts[2] != "+49 "+strings.TrimPrefix(parts[0], "0") {
			t.Fatalf("%q: want the national form twice, then +49 and the national form without its 0", got)
		}
	})
}
