package data_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
)

// germanPlaces maps each locality's code in geo/DE to its name, and each postal code to the codes
// of the localities GeoNames places under it.
func germanPlaces(t *testing.T) (names map[string]string, byPostalCode map[string][]string) {
	t.Helper()
	names, byPostalCode = map[string]string{}, map[string][]string{}
	for _, r := range fejktest.TSVRows(t, "geo/DE/locality.tsv") {
		names[r["code"]] = r["name"]
	}
	for _, r := range fejktest.TSVRows(t, "geo/DE/postal-code.tsv") {
		byPostalCode[r["code"]] = append(byPostalCode[r["code"]], r["locality"])
	}
	return names, byPostalCode
}

func TestGeoDEHoldsTheSixteenLandsInGerman(t *testing.T) {
	t.Parallel()
	names := map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/DE/region.tsv") {
		names[r["code"]] = r["name"]
	}
	if len(names) != 16 {
		t.Errorf("geo/DE/region.tsv holds %d Länder, want 16", len(names))
	}
	for code, want := range map[string]string{"BY": "Bayern", "NI": "Niedersachsen", "NW": "Nordrhein-Westfalen", "MV": "Mecklenburg-Vorpommern", "ST": "Sachsen-Anhalt", "BE": "Berlin"} {
		if names[code] != want {
			t.Errorf("Land %s is %q, want %q", code, names[code], want)
		}
	}
}

// GeoNames' DE.txt places the postal code 01945 in Tettau, Guteborn and Grünewald, all of
// Landkreis Oberspreewald-Lausitz (12066) in Brandenburg.
func TestGeoDEPostalCodeHoldsEveryPlaceGeoNamesNamesForIt(t *testing.T) {
	t.Parallel()
	names, byPostalCode := germanPlaces(t)
	var got []string
	for _, l := range byPostalCode["01945"] {
		got = append(got, names[l])
	}
	slices.Sort(got)
	if want := []string{"Grünewald", "Guteborn", "Tettau"}; !slices.Equal(got, want) {
		t.Errorf("postal code 01945 holds %v, want %v", got, want)
	}
	for _, r := range fejktest.TSVRows(t, "geo/DE/locality.tsv") {
		if r["name"] == "Tettau" && r["municipality"] == "12066" {
			return
		}
	}
	t.Error("Tettau is no locality of Kreis 12066")
}

func TestGeoDEKreisHoldsItsLandAndName(t *testing.T) {
	t.Parallel()
	for _, r := range fejktest.TSVRows(t, "geo/DE/municipality.tsv") {
		if r["code"] == "12066" {
			if r["region"] != "BB" || r["name"] != "Landkreis Oberspreewald-Lausitz" {
				t.Errorf("Kreis 12066 is %v, want Landkreis Oberspreewald-Lausitz in BB", r)
			}
			return
		}
	}
	t.Error("geo/DE/municipality.tsv holds no Kreis 12066")
}

// GeoNames places 87491, Jungholz, in no Land, since it is Austrian.
func TestGeoDEDropsAPostalCodeInNoLand(t *testing.T) {
	t.Parallel()
	_, byPostalCode := germanPlaces(t)
	if len(byPostalCode["87491"]) != 0 {
		t.Errorf("postal code 87491 holds %v, want it dropped as no Land's", byPostalCode["87491"])
	}
}

func TestGeoDEDrawsALocalityByItsPostalCodes(t *testing.T) {
	t.Parallel()
	if got := readJSON(t, "geo/DE/locality.json")["weight"]; got != "postal-codes" {
		t.Errorf("geo/DE/locality.json weights by %v, want the locality's postal-codes", got)
	}
	f := shipped(t, fejkdata.WithSeed(3))
	count := map[string]int{}
	fejktest.EachDraw(20000, func() string {
		return fejktest.FakeTemplate(t, f, "{/geo.DE.locality}")
	}, func(got string) { count[got]++ })
	if count["Berlin"] <= count["Tettau"] {
		t.Errorf("drew Berlin %d times and Tettau %d, want the place with more postal codes drawn more", count["Berlin"], count["Tettau"])
	}
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return out
}

var germanAddress = regexp.MustCompile(`^(.+) (\d{1,3}[a-z]?)\n(\d{5}) (.+)$`)

func TestGermanAddressIsAStreetAndANumberThenAPostalCodeAndItsPlace(t *testing.T) {
	t.Parallel()
	names, byPostalCode := germanPlaces(t)
	streets := map[string]bool{}
	for _, r := range fejktest.TSVRows(t, "geo/DE/street.tsv") {
		streets[r["name"]] = true
	}
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{de_DE.address}") }, func(got string) {
		m := germanAddress.FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("de_DE.address printed %q, want \"<street> <number>\\n<postal code> <place>\"", got)
		}
		if !streets[m[1]] {
			t.Fatalf("de_DE.address printed the street %q, which geo/DE/street.tsv does not hold", m[1])
		}
		for _, l := range byPostalCode[m[3]] {
			if names[l] == m[4] {
				return
			}
		}
		t.Fatalf("de_DE.address printed %q, want a postal code and place that GeoNames pairs", got)
	})
}

// The decision "A country with no open address register covering it whole draws invented
// streets" says each name weighs the same, and a street belongs to no place.
func TestGeoDEStreetsAreInventedNamesDrawnEvenly(t *testing.T) {
	t.Parallel()
	rows := fejktest.TSVRows(t, "geo/DE/street.tsv")
	if len(rows) < 200 {
		t.Errorf("geo/DE/street.tsv holds %d names, want at least 200", len(rows))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if len(r) != 1 {
			t.Fatalf("a street row holds %v, want its name alone", r)
		}
		if seen[r["name"]] {
			t.Errorf("geo/DE/street.tsv holds %q twice", r["name"])
		}
		seen[r["name"]] = true
	}
	street := readJSON(t, "geo/DE/street.json")
	for _, key := range []string{"weight", "parent"} {
		if _, held := street[key]; held {
			t.Errorf("geo/DE/street.json sets %q, want an even draw of a street no place owns", key)
		}
	}
}

func TestReadmeSaysGermanStreetsAreInvented(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(string(b), "#### German places")
	if !found {
		t.Fatal("README.md has no section \"German places\"")
	}
	section, _, _ := strings.Cut(after, "\n#")
	if !strings.Contains(section, "invented") {
		t.Errorf("README.md's German places section does not say the streets are invented:\n%s", section)
	}
}
