package data_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The Address Register holds ten apskritys and sixty savivaldybės.
func TestGeoLTHoldsEachApskritisAsARegionAndEachSavivaldybeAsAMunicipality(t *testing.T) {
	t.Parallel()
	regions, municipalities := codedRows(t, "LT", "region"), codedRows(t, "LT", "municipality")
	if len(regions) != 10 || len(municipalities) != 60 {
		t.Errorf("geo/LT holds %d regions and %d municipalities, want the 10 apskritys and the 60 savivaldybės", len(regions), len(municipalities))
	}
	for code, want := range map[string]string{"10": "Vilniaus apskritis", "2": "Kauno apskritis", "9": "Utenos apskritis"} {
		if got := regions[code]; got["name"] != want || got["timezone"] != "Europe/Vilnius" {
			t.Errorf("region %s is %v, want %s in Europe/Vilnius", code, got, want)
		}
	}
	for code, want := range map[string][2]string{"13": {"Vilniaus miesto savivaldybė", "10"}, "62": {"Molėtų rajono savivaldybė", "9"}, "72": {"Raseinių rajono savivaldybė", "2"}} {
		if got := [2]string{municipalities[code]["name"], municipalities[code]["region"]}; got != want {
			t.Errorf("municipality %s is %v, want %v", code, got, want)
		}
	}
}

// The Universal Postal Union's guide addresses a letter to Laisvės pr. in Vilnius under a code of five digits after
// LT-, and one to Plento g. in Ariogala under LT-60249.
func TestGeoLTPostalCodeHoldsEachLocalityWithAnAddressUnderIt(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LT")
	streets := streetsByLocality(t, "LT")
	for postalCode, want := range map[string][3]string{"LT-04339": {"Vilnius", "13", "Laisvės pr."}, "LT-60249": {"Ariogala", "72", "Plento g."}} {
		held := false
		for _, l := range byPostalCode[postalCode] {
			held = held || [2]string{localities[l]["name"], localities[l]["municipality"]} == [2]string{want[0], want[1]} && streets[l+"|"+want[2]]
		}
		if !held {
			t.Errorf("postal code %s holds the localities %v, want %v among them", postalCode, byPostalCode[postalCode], want)
		}
	}
	if !slices.Equal(byPostalCode["LT-04339"], []string{"31003"}) {
		t.Errorf("postal code LT-04339 holds %v, want Vilnius, 31003, alone", byPostalCode["LT-04339"])
	}
}

func TestGeoLTLocalityIsSelectedByItsCodeInTheRegister(t *testing.T) {
	t.Parallel()
	shape := regexp.MustCompile(`^\d{5}$`)
	localities, _ := codedPlaces(t, "LT")
	for code := range localities {
		if !shape.MatchString(code) {
			t.Fatalf("a locality has the code %q, want its five-digit code in the register", code)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	if got := fejktest.FakeTemplate(t, f, "{/geo.LT.locality[31003]}, {/geo.LT.locality[17507].name}, {/geo.LT.municipality[62]}, {/geo.LT.region[10]}"); got != "Vilnius, Kaunas, Molėtų rajono savivaldybė, Vilniaus apskritis" {
		t.Errorf("localities 31003 and 17507, municipality 62 and region 10 printed %q, want \"Vilnius, Kaunas, Molėtų rajono savivaldybė, Vilniaus apskritis\"", got)
	}
}

func TestGeoLTDrawsALocalityByItsPopulation(t *testing.T) {
	t.Parallel()
	for table, want := range map[string]string{"region": "population", "municipality": "population", "locality": "population", "postal-code": "addresses", "street": "addresses"} {
		if got := readJSON[map[string]any](t, "geo/LT/"+table+".json")["weight"]; got != want {
			t.Errorf("geo/LT/%s.json weighs by %v, want its %s", table, got, want)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	count := map[string]int{}
	fejktest.EachDraw(20000, func() string {
		return fejktest.FakeTemplate(t, f, "{/geo.LT.locality}")
	}, func(got string) { count[got]++ })
	for name, n := range count {
		if n > count["Vilnius"] {
			t.Errorf("drew %s %d times and Vilnius %d, want the locality with most people drawn most", name, n, count["Vilnius"])
		}
	}
	if share := float64(count["Vilnius"]) / 20000; share < 0.12 || share > 0.3 {
		t.Errorf("drew Vilnius in %.2f of the draws, want about a fifth: its blocks of flats count each apartment", share)
	}
}

var lithuanianAddress = regexp.MustCompile(`^(.+) (\d{1,4}\p{Lu}?)\n(LT-\d{5}) (.+)$`)

func TestLithuanianAddressIsAStreetAndANumberThenItsPostalCodeAndLocality(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LT")
	streets := streetsByLocality(t, "LT")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/lt_LT.address}") }, func(got string) {
		m := lithuanianAddress.FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("lt_LT.address printed %q, want \"<street> <number>\\n<postal code> <locality>\"", got)
		}
		for _, l := range byPostalCode[m[3]] {
			if localities[l]["name"] == m[4] && streets[l+"|"+m[1]] {
				return
			}
		}
		t.Fatalf("lt_LT.address printed %q, want a street and a postal code of one locality named %q", got, m[4])
	})
}

func TestLithuanianAddressNumberSometimesEndsInACapitalLetter(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(3))
	lettered := 0
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/lt_LT.address.street-number}") }, func(got string) {
		if regexp.MustCompile(`^\d+[A-Z]$`).MatchString(got) {
			lettered++
		}
	})
	if lettered == 0 || lettered > 600 {
		t.Errorf("%d of 2000 street numbers end in a capital letter, want some and fewer than 600: about one address in seven of the register's does", lettered)
	}
}

func TestLithuanianAddressColumnsAreOnePlace(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LT")
	municipalities, regions := codedRows(t, "LT", "municipality"), codedRows(t, "LT", "region")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/lt_LT.address as a}{a.postal-code}|{a.locality}|{a.municipality}|{a.region}")
	}, func(got string) {
		parts := strings.Split(got, "|")
		for _, l := range byPostalCode[parts[0]] {
			m := municipalities[localities[l]["municipality"]]
			if localities[l]["name"] == parts[1] && m["name"] == parts[2] && regions[m["region"]]["name"] == parts[3] {
				return
			}
		}
		t.Fatalf("lt_LT.address's columns printed %q, want the postal code, locality, municipality and region of one place", got)
	})
}

func TestGeoPageSaysALithuanianPlaceWithoutAStreetIsLeftOut(t *testing.T) {
	t.Parallel()
	if section := geoPageSection(t, "Lithuanian places"); !strings.Contains(section, "no street") {
		t.Errorf("docs/locales/geo.md's Lithuanian places section does not say a place with no street is left out:\n%s", section)
	}
}
