package data_test

import "testing"

func TestLithuanianFirstNamesWeighByThePublishedCount(t *testing.T) {
	first := nameCounts(t, "lt_LT", "first-name")
	for _, c := range []struct {
		sex, name string
		want      int
	}{{"m", "Jonas", 127263}, {"m", "Antanas", 96104}, {"m", "Matas", 13487}, {"f", "Ona", 102245}, {"f", "Marija", 67263}, {"f", "Vida", 15980}} {
		if got := first[c.sex][c.name]; got != c.want {
			t.Errorf("%q of sex %q weighs %d, want %d", c.name, c.sex, got, c.want)
		}
	}
}

func TestLithuanianRecalledFirstNamesWeighByRankBelowTheLastPublishedCount(t *testing.T) {
	first := nameCounts(t, "lt_LT", "first-name")
	for _, c := range []struct {
		sex, name string
		want      int
	}{{"m", "Rokas", 13222}, {"m", "Edmundas", 6743}, {"f", "Gintarė", 15666}, {"f", "Jūratė", 7990}} {
		if got := first[c.sex][c.name]; got != c.want {
			t.Errorf("%q of sex %q weighs %d, want %d", c.name, c.sex, got, c.want)
		}
	}
}

func TestLithuanianNamesKeepTheirLetters(t *testing.T) {
	first, last := nameCounts(t, "lt_LT", "first-name"), nameCounts(t, "lt_LT", "last-name")
	for _, name := range []string{"Stankevičius", "Žukauskas", "Balčiūnas", "Čepulis", "Mažeika", "Šimkus"} {
		if last["m"][name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex m", name)
		}
	}
	for sex, names := range map[string][]string{"f": {"Danutė", "Nijolė", "Aušra"}, "m": {"Kęstutis", "Arūnas", "Žydrūnas"}} {
		for _, name := range names {
			if first[sex][name] == 0 {
				t.Errorf("first-name.tsv has no %q of sex %q", name, sex)
			}
		}
	}
}

func TestLithuanianLastNamesWeighByRankAndTheRarerOnesShareTheLast(t *testing.T) {
	last := nameCounts(t, "lt_LT", "last-name")
	for sex, want := range map[string]map[string]int{
		"m": {"Kazlauskas": 100000, "Petrauskas": 70711, "Jankauskas": 57735, "Stankevičius": 50000, "Vasiliauskas": 44721},
		"f": {"Kazlauskaitė": 100000, "Kazlauskienė": 100000, "Petrauskaitė": 70711, "Stankevičiūtė": 50000, "Stankevičienė": 50000},
	} {
		for name, n := range want {
			if got := last[sex][name]; got != n {
				t.Errorf("%q of sex %q weighs %d, want %d", name, sex, got, n)
			}
		}
	}
}

func TestLithuanianWomenBearAFeminineFormOfALastName(t *testing.T) {
	last := nameCounts(t, "lt_LT", "last-name")
	for _, c := range []struct{ sex, name string }{
		{"f", "Kazlauskaitė"}, {"f", "Kazlauskienė"}, {"m", "Kazlauskas"}, {"f", "Butkutė"}, {"f", "Butkienė"}, {"f", "Kairytė"}, {"f", "Jonaitytė"},
		{"f", "Mažeikaitė"}, {"f", "Ivanova"}, {"m", "Ivanov"},
	} {
		if last[c.sex][c.name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex %q", c.name, c.sex)
		}
	}
	for _, c := range []struct{ sex, name string }{{"f", "Kazlauskas"}, {"m", "Kazlauskaitė"}, {"m", "Kazlauskienė"}, {"f", "Ivanov"}} {
		if last[c.sex][c.name] != 0 {
			t.Errorf("last-name.tsv has %q of sex %q", c.name, c.sex)
		}
	}
}

func TestLithuanianPersonBearsALastNameOfItsSex(t *testing.T) {
	last := nameCounts(t, "lt_LT", "last-name")
	bySex := map[string]string{"moteris": "f", "vyras": "m"}
	eachPerson(t, "lt_LT", func(p drawnPerson) {
		if last[bySex[p.sex]][p.last] == 0 {
			t.Fatalf("%q: %q is not a last name shipped for the sex %q", p.raw, p.last, p.sex)
		}
	})
}
