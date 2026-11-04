package fejkdata

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestShippedUSTaxIds pins the SSA and IRS ranges: an SSN's area is 001-899 but
// 666, its group 01-99 and its serial 0001-9999; an ITIN is 9XX-GG-XXXX with GG
// in 50-65, 70-88, 90-92 or 94-99.
func TestShippedUSTaxIds(t *testing.T) {
	f := shipped(t, WithSeed(5))
	ssn := regexp.MustCompile(`^(\d{3})-(\d{2})-(\d{4})$`)
	itin := regexp.MustCompile(`^9\d{2}-(5\d|6[0-5]|7\d|8[0-8]|9[0-24-9])-\d{4}$`)
	for i := 0; i < 2000; i++ {
		v := fake(t, f, "en_US.ssn")
		m := ssn.FindStringSubmatch(v)
		if m == nil {
			t.Fatalf("ssn %q, want AAA-GG-SSSS", v)
		}
		area, _ := strconv.Atoi(m[1])
		group, _ := strconv.Atoi(m[2])
		serial, _ := strconv.Atoi(m[3])
		if area < 1 || area > 899 || area == 666 || group < 1 || serial < 1 {
			t.Fatalf("ssn %q is in a range the SSA never assigns", v)
		}
		if v := fake(t, f, "en_US.itin"); !itin.MatchString(v) {
			t.Fatalf("itin %q, want %s", v, itin)
		}
	}
}

// NANPA's central office code records mark N11, 555, 950, 958, 959 and 976 unassignable.
func TestShippedUSPhone(t *testing.T) {
	f := shipped(t, WithSeed(11))
	re := regexp.MustCompile(`^(?:\(\d{3}\) |\d{3}-)([2-9]\d{2})-\d{4}$`)
	unassignable := map[string]bool{"555": true, "950": true, "958": true, "959": true, "976": true}
	for i := 0; i < 2000; i++ {
		n := fake(t, f, "en_US.phone")
		m := re.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q does not match %s", n, re)
		}
		if exchange := m[1]; exchange[1:] == "11" || unassignable[exchange] {
			t.Fatalf("phone %q has the unassignable exchange %s", n, exchange)
		}
	}
}

// TestShippedUSTitleAgreesWithSex pins that one person's title and sex agree, and
// that the everyday titles are reachable.
func TestShippedUSTitleAgreesWithSex(t *testing.T) {
	f := shipped(t, WithSeed(4))
	female := map[string]bool{"Miss": true, "Mrs": true, "Ms": true}
	seen := map[string]bool{}
	for i := 0; i < 3000; i++ {
		got := fakeTemplate(t, f, `{/en_US.person as p}{p.prefix}|{p.sex}`)
		title, sex, _ := strings.Cut(got, "|")
		title = strings.TrimSpace(title)
		seen[title] = true
		if title == "Mr" && sex != "male" || female[title] && sex != "female" {
			t.Fatalf("%q: the title contradicts the sex", got)
		}
	}
	for _, want := range []string{"", "Mr", "Ms"} {
		if !seen[want] {
			t.Errorf("en_US.person.prefix never drew %q in 3000 draws", want)
		}
	}
	swedish := map[string]bool{"dr": true, "prof": true}
	for i := 0; i < 50; i++ {
		if got := fake(t, f, "sv_SE.title"); !swedish[got] {
			t.Fatalf("sv_SE.title = %q, want an unsexed Swedish honorific", got)
		}
	}
}
