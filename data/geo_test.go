package data_test

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata"
	"github.com/larvit/fejkdata/internal/fejktest"
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
	f := shipped(t, fejkdata.WithSeed(3))
	fejktest.EachDraw(2000, func() string {
		return fejktest.FakeTemplate(t, f, `{/`+locale+`.address as a}`+template+`|{/`+locale+`.address.street}`)
	}, func(got string) { check(got, strings.Split(got, "|")) })
}

// TestShippedAddressIsOneNamedPick pins that an address's street, postal code,
// locality and region are one place in the geo tables, read through a name in a template.
func TestShippedAddressIsOneNamedPick(t *testing.T) {
	t.Parallel()
	for _, c := range shippedAddressLocales {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			places := fejktest.GeoPlaces(t, "geo/"+c.country)
			eachNamedAddressBesideAStreet(t, c.locale, c.template, func(got string, parts []string) {
				if !places.Has(parts[0], parts[1], parts[2], parts[3]) {
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
	f := shipped(t, fejkdata.WithSeed(3))
	places := fejktest.GeoPlaces(t, "geo/SE")
	var row struct {
		Street     string `fake:"{/sv_SE.address as a}{a.street}"`
		PostalCode string `fake:"{a.postal-code}"`
		Locality   string `fake:"{a.locality}"`
	}
	fejktest.EachDraw(200, func() string {
		if err := f.FakeStruct(&row); err != nil {
			t.Fatal(err)
		}
		return row.Street
	}, func(string) {
		if !places.Has(row.Street, row.PostalCode, row.Locality, "") {
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
	b, err := os.ReadFile("geo/" + country + "/" + table + ".tsv")
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

func geoCountryFolders(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("geo")
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, "geo/"+e.Name())
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
