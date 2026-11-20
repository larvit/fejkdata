package data_test

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// codedPlaces maps each locality's code in a country's tree to its row, and each postal code to
// the codes of the localities with an address under it.
func codedPlaces(t *testing.T, country string) (localities map[string]map[string]string, byPostalCode map[string][]string) {
	t.Helper()
	localities, byPostalCode = map[string]map[string]string{}, map[string][]string{}
	for _, r := range fejktest.TSVRows(t, "geo/"+country+"/locality.tsv") {
		localities[r["code"]] = r
	}
	for _, r := range fejktest.TSVRows(t, "geo/"+country+"/postal-code.tsv") {
		byPostalCode[r["code"]] = append(byPostalCode[r["code"]], r["locality"])
	}
	return localities, byPostalCode
}

// codedRows maps each code of a country's table to its row.
func codedRows(t *testing.T, country, table string) map[string]map[string]string {
	t.Helper()
	rows := map[string]map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/"+country+"/"+table+".tsv") {
		rows[r["code"]] = r
	}
	return rows
}

func TestGeoEEHoldsTheFifteenMaakonnad(t *testing.T) {
	t.Parallel()
	rows := codedRows(t, "EE", "region")
	if len(rows) != 15 {
		t.Errorf("geo/EE/region.tsv holds %d maakonnad, want 15", len(rows))
	}
	for code, want := range map[string]string{"0037": "Harju maakond", "0074": "Saare maakond", "0079": "Tartu maakond"} {
		if got := rows[code]; got["name"] != want || got["timezone"] != "Europe/Tallinn" {
			t.Errorf("maakond %s is %v, want %s in Europe/Tallinn", code, got, want)
		}
	}
}

// Ruhnu's addresses are farms' names, such as Antsu, so Ruhnu vald, 0689, holds no street.
func TestGeoEEMunicipalityHoldsItsMaakondAndShipsOnlyWithAStreet(t *testing.T) {
	t.Parallel()
	rows := codedRows(t, "EE", "municipality")
	for code, want := range map[string][2]string{"0784": {"Tallinn", "0037"}, "0198": {"Harku vald", "0037"}, "0296": {"Keila linn", "0037"}, "0793": {"Tartu linn", "0079"}} {
		if got := [2]string{rows[code]["name"], rows[code]["region"]}; got != want {
			t.Errorf("municipality %s is %v, want %v", code, got, want)
		}
	}
	if _, held := rows["0689"]; held {
		t.Error("geo/EE/municipality.tsv holds Ruhnu vald, 0689, want it dropped: no address there names a street")
	}
}

// The address data places Sarapuu tn 70 in Tabasalu alevik, Harku vald, under 76901, and Pargi
// tn 2b in Keila linn, a linn with no asustusüksus, under 76607.
func TestGeoEEPostalCodeHoldsEachLocalityWithAnAddressUnderIt(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "EE")
	for postalCode, want := range map[string][2]string{"76901": {"Tabasalu alevik", "0198"}, "76607": {"Keila linn", "0296"}} {
		held := false
		for _, l := range byPostalCode[postalCode] {
			held = held || [2]string{localities[l]["name"], localities[l]["municipality"]} == want
		}
		if !held {
			t.Errorf("postal code %s holds the localities %v, want %v among them", postalCode, byPostalCode[postalCode], want)
		}
	}
}

func TestGeoEETallinnIsOnePlace(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "EE")
	if got := localities["0784"]; got["name"] != "Tallinn" || got["municipality"] != "0784" {
		t.Errorf("locality 0784 is %v, want Tallinn in the municipality 0784", got)
	}
	for code, l := range localities {
		if strings.HasSuffix(l["name"], "linnaosa") {
			t.Errorf("locality %s is %q, want no linnaosa as a place", code, l["name"])
		}
	}
	for _, postalCode := range []string{"10111", "11212", "13516"} {
		if !slices.Equal(byPostalCode[postalCode], []string{"0784"}) {
			t.Errorf("postal code %s holds %v, want Tallinn, 0784, alone", postalCode, byPostalCode[postalCode])
		}
	}
}

func TestGeoEELocalityIsSelectedByItsEHAKCode(t *testing.T) {
	t.Parallel()
	shape := regexp.MustCompile(`^\d{4}$`)
	localities, _ := codedPlaces(t, "EE")
	for code := range localities {
		if !shape.MatchString(code) {
			t.Fatalf("a locality has the code %q, want its four-digit EHAK code", code)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	if got := fejktest.FakeTemplate(t, f, "{/geo.EE.locality[0784]}, {/geo.EE.locality[0296].name}, {/geo.EE.locality[8151]}, {/geo.EE.region[0037]}"); got != "Tallinn, Keila linn, Tartu linn, Harju maakond" {
		t.Errorf("localities 0784, 0296 and 8151 and region 0037 printed %q, want \"Tallinn, Keila linn, Tartu linn, Harju maakond\"", got)
	}
}

func TestGeoEEDrawsALocalityByItsPopulation(t *testing.T) {
	t.Parallel()
	for table, want := range map[string]string{"region": "population", "municipality": "population", "locality": "population", "postal-code": "addresses", "street": "addresses"} {
		if got := readJSON[map[string]any](t, "geo/EE/"+table+".json")["weight"]; got != want {
			t.Errorf("geo/EE/%s.json weighs by %v, want its %s", table, got, want)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	count := map[string]int{}
	fejktest.EachDraw(20000, func() string {
		return fejktest.FakeTemplate(t, f, "{/geo.EE.locality}")
	}, func(got string) { count[got]++ })
	for name, n := range count {
		if n > count["Tallinn"] {
			t.Errorf("drew %s %d times and Tallinn %d, want the locality with most people drawn most", name, n, count["Tallinn"])
		}
	}
}

var estonianAddress = regexp.MustCompile(`^(.+ (?:tn|mnt|pst|tee|allee|põik|väljak|plats)) (\d{1,4}[a-zõäöü]?)\n(\d{5}) (.+)$`)

// streetsByLocality holds "locality code|street name" for every street of a country's tree.
func streetsByLocality(t *testing.T, country string) map[string]bool {
	t.Helper()
	streets := map[string]bool{}
	for _, r := range fejktest.TSVRows(t, "geo/"+country+"/street.tsv") {
		streets[r["locality"]+"|"+r["name"]] = true
	}
	return streets
}

func TestEstonianAddressIsAStreetAndANumberThenAPostalCodeAndItsLocality(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "EE")
	streets := streetsByLocality(t, "EE")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/et_EE.address}") }, func(got string) {
		m := estonianAddress.FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("et_EE.address printed %q, want \"<street> <number>\\n<postal code> <locality>\"", got)
		}
		for _, l := range byPostalCode[m[3]] {
			if localities[l]["name"] == m[4] && streets[l+"|"+m[1]] {
				return
			}
		}
		t.Fatalf("et_EE.address printed %q, want a street and a postal code of one locality named %q", got, m[4])
	})
}

func TestEstonianAddressNumberSometimesEndsInALetter(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(3))
	lettered := 0
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/et_EE.address.street-number}") }, func(got string) {
		if regexp.MustCompile(`^\d+[a-z]$`).MatchString(got) {
			lettered++
		}
	})
	if lettered == 0 || lettered > 600 {
		t.Errorf("%d of 2000 street numbers end in a letter, want some and fewer than 600: about one address in nine of the source's does", lettered)
	}
}

func TestEstonianAddressColumnsAreOnePlace(t *testing.T) {
	t.Parallel()
	localities, byPostalCode := codedPlaces(t, "EE")
	municipalities, regions := codedRows(t, "EE", "municipality"), codedRows(t, "EE", "region")
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(500, func() string {
		return fejktest.FakeTemplate(t, f, "{/et_EE.address as a}{a.postal-code}|{a.locality}|{a.municipality}|{a.region}")
	}, func(got string) {
		parts := strings.Split(got, "|")
		for _, l := range byPostalCode[parts[0]] {
			m := municipalities[localities[l]["municipality"]]
			if localities[l]["name"] == parts[1] && m["name"] == parts[2] && regions[m["region"]]["name"] == parts[3] {
				return
			}
		}
		t.Fatalf("et_EE.address's columns printed %q, want the postal code, locality, municipality and region of one place", got)
	})
}

// geoPageSection returns the section of docs/locales/geo.md under a heading.
func geoPageSection(t *testing.T, heading string) string {
	t.Helper()
	b, err := os.ReadFile("../docs/locales/geo.md")
	if err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(string(b), "## "+heading)
	if !found {
		t.Fatalf("docs/locales/geo.md has no section %q", heading)
	}
	section, _, _ := strings.Cut(after, "\n#")
	return section
}

func TestGeoPageSaysAnEstonianPlaceWithoutAStreetIsLeftOut(t *testing.T) {
	t.Parallel()
	if section := geoPageSection(t, "Estonian places"); !strings.Contains(section, "farm") {
		t.Errorf("docs/locales/geo.md's Estonian places section does not say a place whose addresses are farms' names is left out:\n%s", section)
	}
}
