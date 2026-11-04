package fejkdata

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestShippedAddressIsOneNamedPick pins that an address's street, postal code,
// locality and region are one place in the geo tables, read through a name in a template,
// and that a reference beside the name is another address.
func TestShippedAddressIsOneNamedPick(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ locale, country, template, largest string }{
		{"sv_SE", "SE", "{a.street}|{a.postal-code}|{a.locality}|", "Stockholm"},
		{"en_US", "US", "{a.street}|{a.postal-code}|{a.locality}|{a.region}", "New York"},
		{"nb_NO", "NO", "{a.street}|{a.postal-code}|{a.locality}|", "Oslo"},
	} {
		t.Run(c.locale, func(t *testing.T) {
			t.Parallel()
			f := shipped(t, WithSeed(3))
			places := geoPlaces(t, c.country)
			apart, localities := false, map[string]int{}
			for i := 0; i < 2000; i++ {
				got := fakeTemplate(t, f, `{/`+c.locale+`.address as a}`+c.template+`|{/`+c.locale+`.address.street}`)
				parts := strings.Split(got, "|")
				if !places.has(parts[0], parts[1], parts[2], parts[3]) {
					t.Fatalf("%s.address: %q is no street, postal code, locality and region of one place", c.locale, got)
				}
				apart = apart || parts[0] != parts[4]
				localities[parts[2]]++
			}
			if !apart {
				t.Errorf("%s: a reference beside a name drew the name's street in 2000 renders, want two addresses", c.locale)
			}
			for l, n := range localities {
				if n > localities[c.largest] {
					t.Errorf("%s.address drew %s %d times and %s %d, want the locality weighted by population", c.locale, l, n, c.largest, localities[c.largest])
				}
			}
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
	for i := 0; i < 200; i++ {
		if err := f.FakeStruct(&row); err != nil {
			t.Fatal(err)
		}
		if !places.has(row.Street, row.PostalCode, row.Locality, "") {
			t.Fatalf("sv_SE.address: struct tags reading one name, %q, %q and %q, are not the street, postal code and locality of one place", row.Street, row.PostalCode, row.Locality)
		}
	}
}

// 0644000 is Los Angeles.
func TestShippedPlacesSpreadAddresses(t *testing.T) {
	for country, city := range map[string]struct {
		key   string
		least int
	}{"NO": {"Oslo", 2000}, "SE": {"Göteborg", 1500}, "US": {"0644000", 2500}} {
		streets := 0
		for _, r := range geoRows(t, country, "street") {
			if r[1] == city.key {
				streets++
			}
		}
		if streets < city.least {
			t.Errorf("geo.%s: %s ships %d streets, want at least %d", country, city.key, streets, city.least)
		}
	}
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
		if n < 2500 {
			return
		}
	}
	t.Error("geo.US: every locality holds 2500 people or more, want places drawn down to villages")
}

// TestShippedCapitalsAndBigPlaces pins state capitals and places of 100,000 or more that a
// draw by population leaves out.
func TestShippedCapitalsAndBigPlaces(t *testing.T) {
	state := map[string]string{}
	for _, r := range geoRows(t, "US", "municipality") {
		state[r[0]] = r[2]
	}
	shipped := map[string]bool{}
	for _, r := range geoRows(t, "US", "locality") {
		shipped[r[1]+", "+state[r[2]]] = true
	}
	for _, place := range []string{
		"Albany, NY", "Augusta, ME", "Baton Rouge, LA", "Bismarck, ND", "Boise, ID", "Charleston, WV",
		"Frankfort, KY", "Harrisburg, PA", "Helena, MT", "Juneau, AK", "Kansas City, KS", "Lafayette, LA",
		"Lexington, KY", "Montpelier, VT", "New Orleans, LA", "Pierre, SD", "St. Paul, MN", "Stockton, CA",
		"Ventura, CA",
	} {
		if !shipped[place] {
			t.Errorf("geo.US: no locality %s, want every state capital and place of 100,000 or more", place)
		}
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
// country's geo tables; pass an empty region for NO and SE, whose addresses carry none.
type geoPlaceSet struct {
	name, region map[string]string
	locality     map[string]string
	streets      map[[2]string]bool
}

func geoPlaces(t *testing.T, country string) geoPlaceSet {
	rows := func(table string) [][]string { return geoRows(t, country, table) }
	s := geoPlaceSet{name: map[string]string{}, region: map[string]string{}, locality: map[string]string{}, streets: map[[2]string]bool{}}
	if country == "US" {
		county := map[string]string{}
		for _, r := range rows("municipality") {
			county[r[0]] = r[2]
		}
		for _, r := range rows("locality") {
			s.name[r[0]], s.region[r[0]] = r[1], county[r[2]]
		}
	} else {
		for _, r := range rows("locality") {
			s.name[r[0]] = r[0]
		}
	}
	for _, r := range rows("postal-code") {
		s.locality[r[0]] = r[1]
	}
	for _, r := range rows("street") {
		s.streets[[2]string{r[0], r[1]}] = true
	}
	return s
}

func (s geoPlaceSet) has(street, postalCode, locality, region string) bool {
	l, ok := s.locality[postalCode]
	return ok && s.name[l] == locality && s.region[l] == region && s.streets[[2]string{street, l}]
}

func TestGeoPlacesHoldOnlyOneLocalitysStreetPostalCodeNameAndRegion(t *testing.T) {
	s := geoPlaceSet{
		name:     map[string]string{"a": "Alby", "b": "Borg"},
		region:   map[string]string{"a": "Norr", "b": "Söder"},
		locality: map[string]string{"111": "a", "222": "b"},
		streets:  map[[2]string]bool{{"Storgatan", "a"}: true, {"Kyrkvägen", "b"}: true},
	}
	for _, c := range []struct {
		street, postalCode, locality, region string
		want                                 bool
	}{
		{"Storgatan", "111", "Alby", "Norr", true},
		{"Storgatan", "222", "Borg", "Söder", false},
		{"Kyrkvägen", "111", "Alby", "Norr", false},
		{"Storgatan", "111", "Borg", "Norr", false},
		{"Storgatan", "111", "Alby", "Söder", false},
	} {
		if got := s.has(c.street, c.postalCode, c.locality, c.region); got != c.want {
			t.Errorf("has(%q, %q, %q, %q) = %v, want %v", c.street, c.postalCode, c.locality, c.region, got, c.want)
		}
	}
}

// TestEveryGeoTreeHoldsWhatPortsAcrossCountries reads each country under data/geo. A country may ship
// without streets, as the decision on countries with no open address register says, and then its
// address holds neither street column.
func TestEveryGeoTreeHoldsWhatPortsAcrossCountries(t *testing.T) {
	entries, err := os.ReadDir("data/geo")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		dir := "data/geo/" + e.Name()
		if !e.IsDir() {
			continue
		}
		for _, table := range []string{"region", "municipality", "locality", "postal-code"} {
			if _, err := os.Stat(dir + "/" + table + ".tsv"); err != nil {
				t.Errorf("%s holds no table %s: %v", dir, table, err)
			}
		}
		b, err := os.ReadFile(dir + "/address.json")
		if err != nil {
			t.Fatal(err)
		}
		var address map[string]json.RawMessage
		if err := json.Unmarshal(b, &address); err != nil {
			t.Fatalf("%s/address.json: %v", dir, err)
		}
		columns := []string{"postal-code", "locality"}
		if _, err := os.Stat(dir + "/street.tsv"); err == nil {
			columns = append(columns, "street", "street-number")
		}
		for _, c := range columns {
			if _, held := address[c]; !held {
				t.Errorf("%s/address.json holds no column %s", dir, c)
			}
		}
	}
}
