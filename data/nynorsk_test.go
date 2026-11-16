package data_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestShippedNynorskTitleIsSelectable(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(3))
	for _, title := range []string{"dr.", "professor"} {
		t.Run(title, func(t *testing.T) {
			if got := fejktest.Fake(t, f, "nn_NO.title["+title+"]"); got != title {
				t.Errorf("nn_NO.title[%s] = %q, want %q", title, got, title)
			}
		})
	}
}

func TestShippedNynorskReadsTheSameAsBokmal(t *testing.T) {
	for _, c := range []struct{ nn, nb string }{
		{"{/nn_NO.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}", "{/nb_NO.person as p}{p.sex}|{p.prefix}|{p.first}|{p.last}|{p}"},
		{"{/nn_NO.address as a}{a.street}|{a.street-number}|{a.postal-code}|{a.locality}|{a}", "{/nb_NO.address as a}{a.street}|{a.street-number}|{a.postal-code}|{a.locality}|{a}"},
		{"{/nn_NO.date}", "{/nb_NO.date}"},
	} {
		t.Run(c.nn, func(t *testing.T) {
			nn, nb := shipped(t, fejkdata.WithSeed(5)), shipped(t, fejkdata.WithSeed(5))
			for i := range 300 {
				want := fejktest.FakeTemplate(t, nb, c.nb)
				if got := fejktest.FakeTemplate(t, nn, c.nn); got != want {
					t.Fatalf("draw %d of %q = %q, want %q, what %q draws", i, c.nn, got, want, c.nb)
				}
			}
		})
	}
}

var nynorskNationalPhone = regexp.MustCompile(`^\d\d \d\d \d\d \d\d$`)

func TestShippedNynorskFodselsnummerHasValidCheckDigits(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	fejktest.EachDraw(300, func() string { return fejktest.Fake(t, f, "nn_NO.fodselsnummer") }, func(got string) {
		if !fejktest.ValidFnr(got) {
			t.Errorf("nn_NO.fodselsnummer = %q, want a fødselsnummer with valid check digits", got)
		}
	})
}

func TestShippedNynorskPhoneIsOneNumberInItsTwoSpellings(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	fejktest.EachDraw(300, func() string {
		return fejktest.FakeTemplate(t, f, "{/nn_NO.phone as p}{p}|{p.national}|{p.international}")
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 3 || !nynorskNationalPhone.MatchString(cells[1]) || cells[0] != cells[1] || cells[2] != "+47 "+cells[1] {
			t.Errorf("nn_NO.phone, its national and its international = %q, want one number and +47 and the same", got)
		}
	})
}

func TestShippedNynorskEmailIsItsLocalPartAtItsDomain(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	fejktest.EachDraw(300, func() string {
		return fejktest.FakeTemplate(t, f, "{/nn_NO.email as e}{e}|{e.local}|{e.domain}")
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 3 || cells[1] == "" || cells[0] != cells[1]+"@"+cells[2] {
			t.Errorf("nn_NO.email, its local and its domain = %q, want local@domain", got)
		}
	})
}

func TestShippedNynorskCompanyEndsInItsLegalFormAsOftenAsBokmals(t *testing.T) {
	f := shipped(t, fejkdata.WithSeed(5))
	legal := 0
	fejktest.EachDraw(1000, func() string {
		return fejktest.FakeTemplate(t, f, "{/nn_NO.company as c}{c}|{c.legal}")
	}, func(got string) {
		cells := strings.Split(got, "|")
		if len(cells) != 2 || cells[0] == "" || cells[1] != "" && !strings.HasSuffix(cells[0], " "+cells[1]) {
			t.Errorf("nn_NO.company and its legal = %q, want a name ending in its legal form where it has one", got)
		}
		if cells[1] != "" {
			legal++
		}
	})
	if legal < 100 || legal > 900 {
		t.Errorf("nn_NO.company drew a legal form %d times in 1000, want a few hundred, as nb_NO.company does", legal)
	}
}
