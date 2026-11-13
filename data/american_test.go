package data_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// TestShippedUSTaxIds pins the SSA and IRS ranges: an SSN's area is 001-899 but
// 666, its group 01-99 and its serial 0001-9999; an ITIN is 9XX-GG-XXXX with GG
// in 50-65, 70-88, 90-92 or 94-99.
func TestShippedUSTaxIds(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	ssn := regexp.MustCompile(`^(\d{3})-(\d{2})-(\d{4})$`)
	itin := regexp.MustCompile(`^9\d{2}-(5\d|6[0-5]|7\d|8[0-8]|9[0-24-9])-\d{4}$`)
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "en_US.ssn") }, func(v string) {
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
		if v := fejktest.Fake(t, f, "en_US.itin"); !itin.MatchString(v) {
			t.Fatalf("itin %q, want %s", v, itin)
		}
	})
}

// NANPA's central office code records mark N11, 555, 950, 958, 959 and 976 unassignable.
func TestShippedUSPhone(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(11))
	re := regexp.MustCompile(`^(?:\(\d{3}\) |\d{3}-)([2-9]\d{2})-\d{4}$`)
	unassignable := map[string]bool{"555": true, "950": true, "958": true, "959": true, "976": true}
	fejktest.EachDraw(2000, func() string { return fejktest.Fake(t, f, "en_US.phone") }, func(n string) {
		m := re.FindStringSubmatch(n)
		if m == nil {
			t.Fatalf("phone %q does not match %s", n, re)
		}
		if exchange := m[1]; exchange[1:] == "11" || unassignable[exchange] {
			t.Fatalf("phone %q has the unassignable exchange %s", n, exchange)
		}
	})
}

func eachTitleAndSex(drawn []string, check func(got, title, sex string)) {
	for _, got := range drawn {
		title, sex, _ := strings.Cut(got, "|")
		check(got, strings.TrimSpace(title), sex)
	}
}

// TestShippedUSTitleAgreesWithSex pins that one person's title and sex agree, and
// that the everyday titles are reachable.
func TestShippedUSTitleAgreesWithSex(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(4))
	drawn := fejktest.DrawsInTurn(3000, func() string { return fejktest.FakeTemplate(t, f, `{/en_US.person as p}{p.prefix}|{p.sex}`) })
	seen := map[string]bool{}
	eachTitleAndSex(drawn, func(_, title, _ string) { seen[title] = true })
	t.Run("a title agrees with the sex", func(t *testing.T) {
		female := map[string]bool{"Miss": true, "Mrs": true, "Ms": true}
		eachTitleAndSex(drawn, func(got, title, sex string) {
			if title == "Mr" && sex != "male" || female[title] && sex != "female" {
				t.Fatalf("%q: the title contradicts the sex", got)
			}
		})
	})
	for _, want := range []string{"", "Mr", "Ms"} {
		t.Run("draws the title "+strconv.Quote(want), func(t *testing.T) {
			if !seen[want] {
				t.Errorf("en_US.person.prefix never drew %q in 3000 draws", want)
			}
		})
	}
	t.Run("sv_SE.title is an unsexed honorific", func(t *testing.T) {
		swedish := map[string]bool{"dr": true, "prof": true}
		fejktest.EachDraw(50, func() string { return fejktest.Fake(t, f, "sv_SE.title") }, func(got string) {
			if !swedish[got] {
				t.Fatalf("sv_SE.title = %q, want an unsexed Swedish honorific", got)
			}
		})
	})
}
