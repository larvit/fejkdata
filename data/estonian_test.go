package data_test

import "testing"

func TestEstonianNamesWeighByRankAndTheRarerOnesShareTheLast(t *testing.T) {
	first, last := nameCounts(t, "et_EE", "first-name"), nameCounts(t, "et_EE", "last-name")
	for _, c := range []struct {
		what  string
		names map[string]int
		want  map[string]int
	}{
		{"girls", first["f"], map[string]int{"Olga": 100000, "Tatjana": 70711, "Svetlana": 57735, "Maria": 50000, "Ljudmila": 33333, "Marika": 31623, "Agnes": 10000, "Zoja": 10000}},
		{"boys", first["m"], map[string]int{"Aleksandr": 100000, "Sergei": 70711, "Vladimir": 57735, "Andrei": 50000, "Ahti": 10483, "Viljar": 10483}},
		{"last names of women", last["f"], map[string]int{"Tamm": 100000, "Saar": 70711, "Sepp": 57735, "Mägi": 50000, "Ivanova": 28868, "Petrova": 27735, "Aavik": 10000}},
		{"last names of men", last["m"], map[string]int{"Tamm": 100000, "Mägi": 50000, "Ivanov": 28868, "Petrov": 27735, "Aavik": 10000}},
	} {
		t.Run(c.what, func(t *testing.T) {
			for name, want := range c.want {
				if got := c.names[name]; got != want {
					t.Errorf("%q weighs %d, want %d", name, got, want)
				}
			}
		})
	}
}

func TestEstonianNamesKeepTheirLetters(t *testing.T) {
	first, last := nameCounts(t, "et_EE", "first-name"), lastNameCounts(t, "et_EE")
	for _, name := range []string{"Põder", "Lõhmus", "Jõgi", "Kütt", "Mägi", "Härm"} {
		if last[name] == 0 {
			t.Errorf("last-name.tsv has no %q", name)
		}
	}
	for sex, names := range map[string][]string{"f": {"Õie", "Kärt", "Ülle"}, "m": {"Jüri", "Tõnu", "Märt"}} {
		for _, name := range names {
			if first[sex][name] == 0 {
				t.Errorf("first-name.tsv has no %q of sex %q", name, sex)
			}
		}
	}
}

func TestEstonianPersonBearsALastNameOfItsSex(t *testing.T) {
	last := nameCounts(t, "et_EE", "last-name")
	bySex := map[string]string{"naine": "f", "mees": "m"}
	eachPerson(t, "et_EE", func(p drawnPerson) {
		if last[bySex[p.sex]][p.last] == 0 {
			t.Fatalf("%q: %q is not a last name shipped for the sex %q", p.raw, p.last, p.sex)
		}
	})
}

func TestEstonianWomenBearTheFeminineFormOfARussianLastName(t *testing.T) {
	last := nameCounts(t, "et_EE", "last-name")
	for _, c := range []struct{ sex, name string }{{"f", "Ivanova"}, {"m", "Ivanov"}, {"f", "Tamm"}, {"m", "Tamm"}} {
		if last[c.sex][c.name] == 0 {
			t.Errorf("last-name.tsv has no %q of sex %q", c.name, c.sex)
		}
	}
	for _, c := range []struct{ sex, name string }{{"f", "Ivanov"}, {"m", "Ivanova"}, {"f", "Aleva"}} {
		if last[c.sex][c.name] != 0 {
			t.Errorf("last-name.tsv has %q of sex %q", c.name, c.sex)
		}
	}
}
