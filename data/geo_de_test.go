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

// GeoNames' DE.txt places the postal code 01945 in Tettau, Guteborn and Grünewald, among other
// places, all of Landkreis Oberspreewald-Lausitz (12066) in Brandenburg.
func TestGeoDEPostalCodeHoldsEveryPlaceGeoNamesNamesForIt(t *testing.T) {
	t.Parallel()
	names, byPostalCode := germanPlaces(t)
	var got []string
	for _, l := range byPostalCode["01945"] {
		got = append(got, names[l])
	}
	for _, want := range []string{"Grünewald", "Guteborn", "Tettau"} {
		if !slices.Contains(got, want) {
			t.Errorf("postal code 01945 holds %v, want %s among them", got, want)
		}
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
	if got := readJSON[map[string]any](t, "geo/DE/locality.json")["weight"]; got != "postal-codes" {
		t.Errorf("geo/DE/locality.json weighs by %v, want the locality's postal-codes", got)
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

func TestGeoDEKreiseAreNamedAsAGermanWritesThem(t *testing.T) {
	t.Parallel()
	names := map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/DE/municipality.tsv") {
		names[r["code"]] = r["name"]
		if strings.Contains(r["name"], ", ") {
			t.Errorf("Kreis %s is %q, want GeoNames' \"Salzgitter, Stadt\" written \"Stadt Salzgitter\"", r["code"], r["name"])
		}
	}
	for code, want := range map[string]string{"03102": "Stadt Salzgitter", "09564": "Kreisfreie Stadt Nürnberg", "02000": "Freie und Hansestadt Hamburg"} {
		if names[code] != want {
			t.Errorf("Kreis %s is %q, want %q", code, names[code], want)
		}
	}
}

func TestGeoDEHamburgAndBerlinAreOnePlaceEach(t *testing.T) {
	t.Parallel()
	got := map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/DE/locality.tsv") {
		if r["municipality"] == "02000" || r["municipality"] == "11000" {
			got[r["code"]] = r["name"]
		}
	}
	want := map[string]string{"02000-hamburg": "Hamburg", "11000-berlin": "Berlin"}
	if len(got) != 2 || got["02000-hamburg"] != "Hamburg" || got["11000-berlin"] != "Berlin" {
		t.Errorf("the city states' places are %v, want %v: a letter addresses \"22767 Hamburg\"", got, want)
	}
}

// A GeoNames refresh must not move a code to another place, so a locality's code is its Kreis and its name.
func TestGeoDELocalityCodeIsItsKreisAndItsNameInASCII(t *testing.T) {
	t.Parallel()
	shape := regexp.MustCompile(`^\d{5}-[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, r := range fejktest.TSVRows(t, "geo/DE/locality.tsv") {
		if !shape.MatchString(r["code"]) || !strings.HasPrefix(r["code"], r["municipality"]+"-") {
			t.Fatalf("locality %q has the code %q, want its Kreis's code and its name in ASCII", r["name"], r["code"])
		}
	}
	names, _ := germanPlaces(t)
	if names["12066-tettau"] != "Tettau" || names["05315-koeln"] == "" {
		t.Errorf("12066-tettau is %q and 05315-koeln %q, want Tettau and a place", names["12066-tettau"], names["05315-koeln"])
	}
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
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/de_DE.address}") }, func(got string) {
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
	street := readJSON[map[string]any](t, "geo/DE/street.json")
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
