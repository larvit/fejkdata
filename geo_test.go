package fejkdata

import (
	"cmp"
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var shippedAddressLocales = []struct{ locale, country, template, largest string }{
	{"sv_SE", "SE", "{a.street}|{a.postal-code}|{a.locality}|", "Stockholm"},
	{"en_US", "US", "{a.street}|{a.postal-code}|{a.locality}|{a.region}", "New York"},
	{"nb_NO", "NO", "{a.street}|{a.postal-code}|{a.locality}|", "Oslo"},
	{"da_DK", "DK", "{a.street}|{a.postal-code}|{a.locality}|", "København S"},
	{"fi_FI", "FI", "{a.street}|{a.postal-code}|{a.locality}|", "Helsinki"},
}

func eachNamedAddressBesideAStreet(t *testing.T, locale, template string, check func(got string, parts []string)) {
	t.Helper()
	f := shipped(t, WithSeed(3))
	eachDraw(2000, func() string {
		return fakeTemplate(t, f, `{/`+locale+`.address as a}`+template+`|{/`+locale+`.address.street}`)
	}, func(got string) { check(got, strings.Split(got, "|")) })
}

// TestShippedAddressIsOneNamedPick pins that an address's street, postal code,
// locality and region are one place in the geo tables, read through a name in a template.
func TestShippedAddressIsOneNamedPick(t *testing.T) {
	t.Parallel()
	for _, c := range shippedAddressLocales {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			places := geoPlaces(t, c.country)
			eachNamedAddressBesideAStreet(t, c.locale, c.template, func(got string, parts []string) {
				if !places.has(parts[0], parts[1], parts[2], parts[3]) {
					t.Fatalf("%s.address: %q is no street, postal code, locality and region of one place", c.locale, got)
				}
			})
		})
	}
}

func TestShippedAddressReferenceBesideANameIsAnotherAddress(t *testing.T) {
	t.Parallel()
	for _, c := range shippedAddressLocales {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			apart := false
			eachNamedAddressBesideAStreet(t, c.locale, c.template, func(_ string, parts []string) {
				apart = apart || parts[0] != parts[4]
			})
			if !apart {
				t.Errorf("%s: a reference beside a name drew the name's street in 2000 renders, want two addresses", c.locale)
			}
		})
	}
}

func errorOnALocalityDrawnMoreThan(t *testing.T, locale, largest string, localities map[string]int) {
	t.Helper()
	for l, n := range localities {
		if n > localities[largest] {
			t.Errorf("%s.address drew %s %d times and %s %d, want the locality weighted by population", locale, l, n, largest, localities[largest])
		}
	}
}

func TestShippedAddressDrawsItsLargestLocalityMost(t *testing.T) {
	t.Parallel()
	for _, c := range shippedAddressLocales {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			localities := map[string]int{}
			eachNamedAddressBesideAStreet(t, c.locale, c.template, func(_ string, parts []string) {
				localities[parts[2]]++
			})
			errorOnALocalityDrawnMoreThan(t, c.locale, c.largest, localities)
		})
	}
}

func TestShippedAddressStructTagsReadOneName(t *testing.T) {
	t.Parallel()
	f := shipped(t, WithSeed(3))
	places := geoPlaces(t, "SE")
	var row struct {
		Street     string `fake:"{/sv_SE.address as a}{a.street}"`
		PostalCode string `fake:"{a.postal-code}"`
		Locality   string `fake:"{a.locality}"`
	}
	eachDraw(200, func() string {
		if err := f.FakeStruct(&row); err != nil {
			t.Fatal(err)
		}
		return row.Street
	}, func(string) {
		if !places.has(row.Street, row.PostalCode, row.Locality, "") {
			t.Fatalf("sv_SE.address: struct tags reading one name, %q, %q and %q, are not the street, postal code and locality of one place", row.Street, row.PostalCode, row.Locality)
		}
	})
}

func geoStreetsOfLocality(t *testing.T, country, locality string) int {
	t.Helper()
	streets := 0
	for _, r := range geoRows(t, country, "street") {
		if r[1] == locality {
			streets++
		}
	}
	return streets
}

// 0644000 is Los Angeles.
func TestShippedBigCitiesShipManyStreets(t *testing.T) {
	for country, city := range map[string]struct {
		key   string
		least int
	}{"DK": {"København S", 500}, "FI": {"Helsinki", 2000}, "NO": {"Oslo", 2000}, "SE": {"Göteborg", 1500}, "US": {"0644000", 2500}} {
		t.Run(country, func(t *testing.T) {
			if streets := geoStreetsOfLocality(t, country, city.key); streets < city.least {
				t.Errorf("geo.%s: %s ships %d streets, want at least %d", country, city.key, streets, city.least)
			}
		})
	}
}

func geoUSHoldsALocalityOfFewerThan(t *testing.T, people int) bool {
	t.Helper()
	header, _, _ := strings.Cut(geoTSV(t, "US", "locality"), "\n")
	population := slices.Index(strings.Split(header, "\t"), "population")
	if population < 0 {
		t.Fatal("geo.US locality.tsv has no population column")
	}
	for _, r := range geoRows(t, "US", "locality") {
		n, err := strconv.Atoi(r[population])
		if err != nil {
			t.Fatal(err)
		}
		if n < people {
			return true
		}
	}
	return false
}

func TestShippedUSPlacesAreDrawnDownToVillages(t *testing.T) {
	if !geoUSHoldsALocalityOfFewerThan(t, 2500) {
		t.Error("geo.US: every locality holds 2500 people or more, want places drawn down to villages")
	}
}

func geoUSLocalitiesWithTheirState(t *testing.T) map[string]bool {
	t.Helper()
	state := map[string]string{}
	for _, r := range geoRows(t, "US", "municipality") {
		state[r[0]] = r[2]
	}
	shipped := map[string]bool{}
	for _, r := range geoRows(t, "US", "locality") {
		shipped[r[1]+", "+state[r[2]]] = true
	}
	return shipped
}

// TestShippedCapitalsAndBigPlaces pins state capitals and places of 100,000 or more that a
// draw by population leaves out.
func TestShippedCapitalsAndBigPlaces(t *testing.T) {
	shipped := geoUSLocalitiesWithTheirState(t)
	for _, place := range []string{
		"Albany, NY", "Augusta, ME", "Baton Rouge, LA", "Bismarck, ND", "Boise, ID", "Charleston, WV",
		"Frankfort, KY", "Harrisburg, PA", "Helena, MT", "Juneau, AK", "Kansas City, KS", "Lafayette, LA",
		"Lexington, KY", "Montpelier, VT", "New Orleans, LA", "Pierre, SD", "St. Paul, MN", "Stockton, CA",
		"Ventura, CA",
	} {
		t.Run(place, func(t *testing.T) {
			if !shipped[place] {
				t.Errorf("geo.US: no locality %s, want every state capital and place of 100,000 or more", place)
			}
		})
	}
}

func geoTSV(t *testing.T, country, table string) string {
	b, err := os.ReadFile("data/geo/" + country + "/" + table + ".tsv")
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func geoRows(t *testing.T, country, table string) [][]string {
	var out [][]string
	for _, line := range strings.Split(geoTSV(t, country, table), "\n")[1:] {
		out = append(out, strings.Split(line, "\t"))
	}
	return out
}

// geoPlaceSet says whether a street, postal code, locality and region belong to one locality of a
// country's geo tables; pass an empty region where the address prints none.
type geoPlaceSet struct {
	name, region map[string]string
	locality     map[string]string
	streets      map[[2]string]bool
}

// geoPlaces keys a locality by its code where its table has one, else by its name.
func geoPlaces(t *testing.T, country string) geoPlaceSet {
	rows := func(table string) []map[string]string { return tsvRows(t, "data/geo/"+country+"/"+table+".tsv") }
	s := geoPlaceSet{name: map[string]string{}, region: map[string]string{}, locality: map[string]string{}, streets: map[[2]string]bool{}}
	region := map[string]string{}
	for _, r := range rows("municipality") {
		region[r["code"]] = r["region"]
	}
	for _, r := range rows("locality") {
		key := cmp.Or(r["code"], r["name"])
		s.name[key], s.region[key] = r["name"], region[r["municipality"]]
	}
	for _, r := range rows("postal-code") {
		s.locality[r["code"]] = r["locality"]
	}
	for _, r := range rows("street") {
		s.streets[[2]string{r["name"], r["locality"]}] = true
	}
	return s
}

func (s geoPlaceSet) has(street, postalCode, locality, region string) bool {
	l, ok := s.locality[postalCode]
	return ok && s.name[l] == locality && (region == "" || s.region[l] == region) && s.streets[[2]string{street, l}]
}

func TestGeoPlacesHoldOnlyOneLocalitysStreetPostalCodeNameAndRegion(t *testing.T) {
	s := geoPlaceSet{
		name:     map[string]string{"a": "Alby", "b": "Borg"},
		region:   map[string]string{"a": "Norr", "b": "Söder"},
		locality: map[string]string{"111": "a", "222": "b"},
		streets:  map[[2]string]bool{{"Storgatan", "a"}: true, {"Kyrkvägen", "b"}: true},
	}
	for _, c := range []struct {
		name                                 string
		street, postalCode, locality, region string
		want                                 bool
	}{
		{"one locality", "Storgatan", "111", "Alby", "Norr", true},
		{"the postal code, name and region of another locality", "Storgatan", "222", "Borg", "Söder", false},
		{"the street of another locality", "Kyrkvägen", "111", "Alby", "Norr", false},
		{"the name of another locality", "Storgatan", "111", "Borg", "Norr", false},
		{"the region of another locality", "Storgatan", "111", "Alby", "Söder", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := s.has(c.street, c.postalCode, c.locality, c.region); got != c.want {
				t.Errorf("has(%q, %q, %q, %q) = %v, want %v", c.street, c.postalCode, c.locality, c.region, got, c.want)
			}
		})
	}
}

func geoCountryFolders(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("data/geo")
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, "data/geo/"+e.Name())
		}
	}
	return dirs
}

func TestEveryGeoTreeHoldsWhatPortsAcrossCountries(t *testing.T) {
	for _, dir := range geoCountryFolders(t) {
		for _, table := range []string{"region", "municipality", "locality", "postal-code"} {
			t.Run(dir+"/"+table, func(t *testing.T) {
				if _, err := os.Stat(dir + "/" + table + ".tsv"); err != nil {
					t.Errorf("%s holds no table %s: %v", dir, table, err)
				}
			})
		}
	}
}

func geoAddressColumnsOf(dir string) []string {
	columns := []string{"postal-code", "locality"}
	if _, err := os.Stat(dir + "/street.tsv"); err == nil {
		columns = append(columns, "street", "street-number")
	}
	return columns
}

func geoAddress(t *testing.T, dir string) map[string]json.RawMessage {
	t.Helper()
	b, err := os.ReadFile(dir + "/address.json")
	if err != nil {
		t.Fatal(err)
	}
	var address map[string]json.RawMessage
	if err := json.Unmarshal(b, &address); err != nil {
		t.Fatalf("%s/address.json: %v", dir, err)
	}
	return address
}

// A country may ship without streets, as the decision on countries with no open address register
// says, and then its address holds neither street column.
func TestEveryGeoAddressHoldsTheColumnsItsTreeHasTablesFor(t *testing.T) {
	for _, dir := range geoCountryFolders(t) {
		for _, c := range geoAddressColumnsOf(dir) {
			t.Run(dir+"/"+c, func(t *testing.T) {
				if _, held := geoAddress(t, dir)[c]; !held {
					t.Errorf("%s/address.json holds no column %s", dir, c)
				}
			})
		}
	}
}
