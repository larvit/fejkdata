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

// frenchPlaces maps each commune's code in geo/FR to its row, and each postal code to the codes
// of the communes the Base Adresse Nationale has an address in under it.
func frenchPlaces(t *testing.T) (communes map[string]map[string]string, byPostalCode map[string][]string) {
	t.Helper()
	communes, byPostalCode = map[string]map[string]string{}, map[string][]string{}
	for _, r := range fejktest.TSVRows(t, "geo/FR/locality.tsv") {
		communes[r["code"]] = r
	}
	for _, r := range fejktest.TSVRows(t, "geo/FR/postal-code.tsv") {
		byPostalCode[r["code"]] = append(byPostalCode[r["code"]], r["locality"])
	}
	return communes, byPostalCode
}

// INSEE's legal populations cover every région but Mayotte, so 17 of the 18 ship.
func TestGeoFRHoldsEachRegionWithALegalPopulation(t *testing.T) {
	t.Parallel()
	rows := map[string]map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/FR/region.tsv") {
		rows[r["code"]] = r
	}
	if _, held := rows["06"]; held || len(rows) != 17 {
		t.Errorf("geo/FR/region.tsv holds %d régions, want the 17 other than Mayotte, 06", len(rows))
	}
	for code, want := range map[string][2]string{
		"11": {"Île-de-France", "Europe/Paris"}, "84": {"Auvergne-Rhône-Alpes", "Europe/Paris"}, "94": {"Corse", "Europe/Paris"},
		"01": {"Guadeloupe", "America/Guadeloupe"}, "02": {"Martinique", "America/Martinique"}, "03": {"Guyane", "America/Cayenne"},
		"04": {"La Réunion", "Indian/Reunion"},
	} {
		if got := [2]string{rows[code]["name"], rows[code]["timezone"]}; got != want {
			t.Errorf("région %s is %v, want %v", code, got, want)
		}
	}
}

func TestGeoFRDepartementHoldsItsRegionAndName(t *testing.T) {
	t.Parallel()
	got := map[string][2]string{}
	for _, r := range fejktest.TSVRows(t, "geo/FR/municipality.tsv") {
		got[r["code"]] = [2]string{r["name"], r["region"]}
	}
	if _, held := got["976"]; held || len(got) != 100 {
		t.Errorf("geo/FR/municipality.tsv holds %d départements, want the 100 other than Mayotte, 976", len(got))
	}
	for code, want := range map[string][2]string{"01": {"Ain", "84"}, "2A": {"Corse-du-Sud", "94"}, "75": {"Paris", "11"}, "974": {"La Réunion", "04"}} {
		if got[code] != want {
			t.Errorf("département %s is %v, want %v", code, got[code], want)
		}
	}
}

// The Base Adresse Nationale places every address of L'Abergement-Clémenciat, commune 01001 of
// the Ain, under 01400, as it does addresses of Châtillon-sur-Chalaronne, commune 01093.
func TestGeoFRPostalCodeHoldsEachCommuneWithAnAddressUnderIt(t *testing.T) {
	t.Parallel()
	communes, byPostalCode := frenchPlaces(t)
	if c := communes["01001"]; c["name"] != "L'Abergement-Clémenciat" || c["municipality"] != "01" {
		t.Errorf("commune 01001 is %v, want L'Abergement-Clémenciat in département 01", c)
	}
	for _, want := range []string{"01001", "01093"} {
		if !slices.Contains(byPostalCode["01400"], want) {
			t.Errorf("postal code 01400 holds %v, want %s among them", byPostalCode["01400"], want)
		}
	}
}

func TestGeoFRParisLyonAndMarseilleAreOnePlaceEach(t *testing.T) {
	t.Parallel()
	communes, byPostalCode := frenchPlaces(t)
	for code, want := range map[string]string{"75056": "Paris", "69123": "Lyon", "13055": "Marseille"} {
		if communes[code]["name"] != want {
			t.Errorf("commune %s is %q, want %q: a letter addresses \"75011 Paris\"", code, communes[code]["name"], want)
		}
	}
	for code, c := range communes {
		if strings.Contains(c["name"], "Arrondissement") {
			t.Errorf("commune %s is %q, want no arrondissement as a place", code, c["name"])
		}
	}
	for postalCode, want := range map[string]string{"75001": "75056", "75020": "75056", "75116": "75056", "69009": "69123", "13016": "13055"} {
		if !slices.Equal(byPostalCode[postalCode], []string{want}) {
			t.Errorf("postal code %s holds %v, want %s alone", postalCode, byPostalCode[postalCode], want)
		}
	}
}

func TestGeoFRLocalityIsSelectedByItsINSEECode(t *testing.T) {
	t.Parallel()
	shape := regexp.MustCompile(`^(\d{5}|2[AB]\d{3})$`)
	communes, _ := frenchPlaces(t)
	for code := range communes {
		if !shape.MatchString(code) {
			t.Fatalf("a commune has the code %q, want its five-character INSEE code", code)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	if got := fejktest.FakeTemplate(t, f, "{/geo.FR.locality[75056]}, {/geo.FR.locality[2A004].name}"); got != "Paris, Ajaccio" {
		t.Errorf("localities 75056 and 2A004 printed %q, want \"Paris, Ajaccio\"", got)
	}
}

func TestGeoFRDrawsALocalityByItsPopulation(t *testing.T) {
	t.Parallel()
	if got := readJSON[map[string]any](t, "geo/FR/locality.json")["weight"]; got != "population" {
		t.Errorf("geo/FR/locality.json weighs by %v, want the commune's population", got)
	}
	communes, _ := frenchPlaces(t)
	for code, c := range communes {
		if c["population"] == "0" {
			t.Errorf("commune %s ships with no people, want it dropped", code)
		}
	}
	f := shipped(t, fejkdata.WithSeed(3))
	count := map[string]int{}
	fejktest.EachDraw(20000, func() string {
		return fejktest.FakeTemplate(t, f, "{/geo.FR.locality}")
	}, func(got string) { count[got]++ })
	for name, n := range count {
		if n > count["Paris"] {
			t.Errorf("drew %s %d times and Paris %d, want the commune with most people drawn most", name, n, count["Paris"])
		}
	}
}

var frenchAddress = regexp.MustCompile(`^(\d{1,4}(?: bis)?) (.+)\n(\d{5}) (.+)$`)

func TestFrenchAddressIsANumberAndAStreetThenAPostalCodeAndItsCommune(t *testing.T) {
	t.Parallel()
	communes, byPostalCode := frenchPlaces(t)
	streets := map[string]bool{}
	for _, r := range fejktest.TSVRows(t, "geo/FR/street.tsv") {
		streets[r["name"]] = true
	}
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/fr_FR.address}") }, func(got string) {
		m := frenchAddress.FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("fr_FR.address printed %q, want \"<number> <street>\\n<postal code> <commune>\"", got)
		}
		if !streets[m[2]] {
			t.Fatalf("fr_FR.address printed the street %q, which geo/FR/street.tsv does not hold", m[2])
		}
		for _, l := range byPostalCode[m[3]] {
			if communes[l]["name"] == m[4] {
				return
			}
		}
		t.Fatalf("fr_FR.address printed %q, want a postal code and commune that the Base Adresse Nationale pairs", got)
	})
}

func TestFrenchAddressNumberSometimesEndsInBis(t *testing.T) {
	t.Parallel()
	f := shipped(t, fejkdata.WithSeed(3))
	bis := 0
	fejktest.EachDraw(2000, func() string { return fejktest.FakeTemplate(t, f, "{/fr_FR.address.street-number}") }, func(got string) {
		if strings.HasSuffix(got, " bis") {
			bis++
		}
	})
	if bis == 0 || bis > 400 {
		t.Errorf("%d of 2000 street numbers end in bis, want some and under a fifth", bis)
	}
}

func TestFrenchAddressColumnsAreOnePlace(t *testing.T) {
	t.Parallel()
	communes, byPostalCode := frenchPlaces(t)
	departments, regions := map[string]map[string]string{}, map[string]string{}
	for _, r := range fejktest.TSVRows(t, "geo/FR/municipality.tsv") {
		departments[r["name"]] = r
	}
	for _, r := range fejktest.TSVRows(t, "geo/FR/region.tsv") {
		regions[r["code"]] = r["name"]
	}
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string {
		return fejktest.FakeTemplate(t, f, "{/fr_FR.address as a}{a.postal-code}|{a.locality}|{a.municipality}|{a.region}")
	}, func(got string) {
		parts := strings.Split(got, "|")
		department := departments[parts[2]]
		for _, l := range byPostalCode[parts[0]] {
			if c := communes[l]; c["name"] == parts[1] && c["municipality"] == department["code"] && regions[department["region"]] == parts[3] {
				return
			}
		}
		t.Fatalf("fr_FR.address's columns are %q, want a postal code, its commune, the commune's département and that one's région", got)
	})
}

// Item 271 gives geo.FR its real streets. Until then its address draws an invented one, as the
// decision "A country with no open address register covering it whole draws invented streets" has
// it: each name weighs the same, and a street belongs to no place.
func TestGeoFRStreetsAreInventedNamesDrawnEvenly(t *testing.T) {
	t.Parallel()
	rows := fejktest.TSVRows(t, "geo/FR/street.tsv")
	if len(rows) < 200 {
		t.Errorf("geo/FR/street.tsv holds %d names, want at least 200", len(rows))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if len(r) != 1 {
			t.Fatalf("a street row holds %v, want its name alone", r)
		}
		if seen[r["name"]] {
			t.Errorf("geo/FR/street.tsv holds %q twice", r["name"])
		}
		seen[r["name"]] = true
	}
	street := readJSON[map[string]any](t, "geo/FR/street.json")
	for _, key := range []string{"weight", "parent"} {
		if _, held := street[key]; held {
			t.Errorf("geo/FR/street.json sets %q, want an even draw of a street no place owns", key)
		}
	}
}

func TestGeoPageSaysFrenchStreetsAreInvented(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("../docs/locales/geo.md")
	if err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(string(b), "## French places")
	if !found {
		t.Fatal("docs/locales/geo.md has no section \"French places\"")
	}
	section, _, _ := strings.Cut(after, "\n#")
	if !strings.Contains(section, "invented") {
		t.Errorf("docs/locales/geo.md's French places section does not say the streets are invented:\n%s", section)
	}
}
