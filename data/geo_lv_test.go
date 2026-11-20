package data_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// The State Address Register holds 35 novadi and 7 state cities in force, and no level above them.
func TestGeoLVHoldsEachNovadsAndStateCityAsARegion(t *testing.T) {
	t.Parallel()
	rows := codedRows(t, "LV", "region")
	if len(rows) != 42 {
		t.Errorf("geo/LV/region.tsv holds %d regions, want the 35 novadi and the 7 state cities", len(rows))
	}
	for code, want := range map[string]string{"0001000": "Rīga", "0035000": "Limbažu nov.", "0051000": "Talsu nov."} {
		if got := rows[code]; got["name"] != want || got["timezone"] != "Europe/Riga" {
			t.Errorf("region %s is %v, want %s in Europe/Riga", code, got, want)
		}
	}
}

func TestGeoLVMunicipalityIsAPagastsOrATownUnderItsRegion(t *testing.T) {
	t.Parallel()
	rows := codedRows(t, "LV", "municipality")
	for code, want := range map[string][2]string{"0001000": {"Rīga", "0001000"}, "0035200": {"Ainaži", "0035000"}, "0048460": {"Siguldas pag.", "0048000"}} {
		if got := [2]string{rows[code]["name"], rows[code]["region"]}; got != want {
			t.Errorf("municipality %s is %v, want %v", code, got, want)
		}
	}
}

// The Universal Postal Union's guide addresses a letter to Valdemāra iela in Ainaži under LV-4035,
// and one to Krišjāņa Barona iela in Rīga under LV-1050.
func TestGeoLVPostalCodeHoldsEachLocalityWithAnAddressUnderIt(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LV")
	streets := streetsByLocality(t, "LV")
	for postalCode, want := range map[string][3]string{"LV-4035": {"Ainaži", "0035200", "Valdemāra iela"}, "LV-1050": {"Rīga", "0001000", "Krišjāņa Barona iela"}} {
		held := false
		for _, l := range byPostalCode[postalCode] {
			held = held || [2]string{localities[l]["name"], localities[l]["municipality"]} == [2]string{want[0], want[1]} && streets[l+"|"+want[2]]
		}
		if !held {
			t.Errorf("postal code %s holds the localities %v, want %v among them", postalCode, byPostalCode[postalCode], want)
		}
	}
	if !slices.Equal(byPostalCode["LV-1050"], []string{"100003003"}) {
		t.Errorf("postal code LV-1050 holds %v, want Rīga, 100003003, alone", byPostalCode["LV-1050"])
	}
}

func TestGeoLVLocalityIsSelectedByItsCodeInTheRegister(t *testing.T) {
	t.Parallel()
	shape := regexp.MustCompile(`^\d{9}$`)
	localities, _ := codedPlaces(t, "LV")
	for code := range localities {
		if !shape.MatchString(code) {
			t.Fatalf("a locality has the code %q, want its nine-digit code in the register", code)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	if got := fejktest.FakeTemplate(t, f, "{/geo.LV.locality[100003003]}, {/geo.LV.locality[100003366].name}, {/geo.LV.municipality[0048460]}, {/geo.LV.region[0051000]}"); got != "Rīga, Ainaži, Siguldas pag., Talsu nov." {
		t.Errorf("localities 100003003 and 100003366, municipality 0048460 and region 0051000 printed %q, want \"Rīga, Ainaži, Siguldas pag., Talsu nov.\"", got)
	}
}

func TestGeoLVDrawsALocalityByItsPopulation(t *testing.T) {
	t.Parallel()
	for table, want := range map[string]string{"region": "population", "municipality": "population", "locality": "population", "postal-code": "addresses", "street": "addresses"} {
		if got := readJSON[map[string]any](t, "geo/LV/"+table+".json")["weight"]; got != want {
			t.Errorf("geo/LV/%s.json weighs by %v, want its %s", table, got, want)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	count := map[string]int{}
	fejktest.EachDraw(20000, func() string {
		return fejktest.FakeTemplate(t, f, "{/geo.LV.locality}")
	}, func(got string) { count[got]++ })
	for name, n := range count {
		if n > count["Rīga"] {
			t.Errorf("drew %s %d times and Rīga %d, want the locality with most people drawn most", name, n, count["Rīga"])
		}
	}
	if share := float64(count["Rīga"]) / 20000; share < 0.2 || share > 0.5 {
		t.Errorf("drew Rīga in %.2f of the draws, want about a third: its blocks of flats count each apartment", share)
	}
}

var latvianAddress = regexp.MustCompile(`^(.+) (\d{1,4}\p{Lu}?)\n(.+), (LV-\d{4})$`)

func TestLatvianAddressIsAStreetAndANumberThenItsLocalityAndPostalCode(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LV")
	streets := streetsByLocality(t, "LV")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/lv_LV.address}") }, func(got string) {
		m := latvianAddress.FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("lv_LV.address printed %q, want \"<street> <number>\\n<locality>, <postal code>\"", got)
		}
		for _, l := range byPostalCode[m[4]] {
			if localities[l]["name"] == m[3] && streets[l+"|"+m[1]] {
				return
			}
		}
		t.Fatalf("lv_LV.address printed %q, want a street and a postal code of one locality named %q", got, m[3])
	})
}

func TestLatvianAddressNumberSometimesEndsInACapitalLetter(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(3))
	lettered := 0
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/lv_LV.address.street-number}") }, func(got string) {
		if regexp.MustCompile(`^\d+[A-Z]$`).MatchString(got) {
			lettered++
		}
	})
	if lettered == 0 || lettered > 800 {
		t.Errorf("%d of 2000 street numbers end in a capital letter, want some and fewer than 800: about one building in five of the register's does", lettered)
	}
}

func TestLatvianAddressColumnsAreOnePlace(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "LV")
	municipalities, regions := codedRows(t, "LV", "municipality"), codedRows(t, "LV", "region")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/lv_LV.address as a}{a.postal-code}|{a.locality}|{a.municipality}|{a.region}")
	}, func(got string) {
		parts := strings.Split(got, "|")
		for _, l := range byPostalCode[parts[0]] {
			m := municipalities[localities[l]["municipality"]]
			if localities[l]["name"] == parts[1] && m["name"] == parts[2] && regions[m["region"]]["name"] == parts[3] {
				return
			}
		}
		t.Fatalf("lv_LV.address's columns printed %q, want the postal code, locality, municipality and region of one place", got)
	})
}

func TestGeoPageSaysALatvianPlaceWithoutAStreetIsLeftOut(t *testing.T) {
	t.Parallel()
	if section := geoPageSection(t, "Latvian places"); !strings.Contains(section, "house's name") {
		t.Errorf("docs/locales/geo.md's Latvian places section does not say a place whose addresses are houses' names is left out:\n%s", section)
	}
}
