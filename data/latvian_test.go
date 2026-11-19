package data_test

import "testing"

func TestLatvianFirstNamesWeighByThePublishedCount(t *testing.T) {
	first := nameCounts(t, "lv_LV", "first-name")
	for _, c := range []struct {
		sex, name string
		want      int
	}{{"m", "Jānis", 44740}, {"m", "Aleksandrs", 31812}, {"f", "Anna", 21772}, {"f", "Jeļena", 20246}} {
		if got := first[c.sex][c.name]; got != c.want {
			t.Errorf("%q of sex %q weighs %d, want %d", c.name, c.sex, got, c.want)
		}
	}
}

func TestLatvianNamesKeepTheirLetters(t *testing.T) {
	first, last := nameCounts(t, "lv_LV", "first-name"), lastNameCounts(t, "lv_LV")
	for _, name := range []string{"Bērziņš", "Krūmiņš", "Kļaviņš", "Eglītis", "Dombrovskis", "Vītols"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
	for sex, names := range map[string][]string{"f": {"Līga", "Ņina", "Jeļena"}, "m": {"Ģirts", "Ēriks", "Mārtiņš"}} {
		for _, name := range names {
			if first[sex][name] == 0 {
				t.Errorf("first-name.tsv has no %q of sex %q", name, sex)
			}
		}
	}
}

func TestLatvianLastNamesWeighByRankAndTheRarerOnesShareTheLast(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	for sex, want := range map[string]map[string]int{
		"m": {"Bērziņš": 100000, "Kalniņš": 70711, "Ozoliņš": 57735, "Jansons": 50000, "Ivanovs": 21822, "Zvirbulis": 9129},
		"f": {"Bērziņa": 100000, "Kalniņa": 70711, "Ozoliņa": 57735, "Jansone": 50000, "Ivanova": 21822, "Zvirbule": 9129},
	} {
		for name, n := range want {
			if got := last[sex][name]; got != n {
				t.Errorf("%q of sex %q weighs %d, want %d", name, sex, got, n)
			}
		}
	}
}

func TestLatvianWomenBearTheFeminineFormOfALastName(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	for _, c := range []struct{ sex, name string }{
		{"f", "Bērziņa"}, {"m", "Bērziņš"}, {"f", "Balode"}, {"m", "Balodis"}, {"f", "Dombrovska"}, {"f", "Ozola"}, {"f", "Millere"},
		{"f", "Zvejniece"}, {"f", "Kalēja"}, {"f", "Liepa"}, {"m", "Liepa"},
	} {
		if last[c.sex][c.name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex %q", c.name, c.sex)
		}
	}
	for _, c := range []struct{ sex, name string }{{"f", "Bērziņš"}, {"m", "Bērziņa"}, {"f", "Balodis"}, {"m", "Balode"}, {"f", "Ivanovs"}} {
		if last[c.sex][c.name] != 0 {
			t.Errorf("last-name.tsv has %q of sex %q", c.name, c.sex)
		}
	}
}

func TestLatvianPersonBearsALastNameOfItsSex(t *testing.T) {
	last := nameCounts(t, "lv_LV", "last-name")
	bySex := map[string]string{"sieviete": "f", "vīrietis": "m"}
	eachPerson(t, "lv_LV", func(p drawnPerson) {
		if last[bySex[p.sex]][p.last] == 0 {
			t.Fatalf("%q: %q is not a last name shipped for the sex %q", p.raw, p.last, p.sex)
		}
	})
}
