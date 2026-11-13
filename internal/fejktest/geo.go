package fejktest

import (
	"cmp"
	"testing"
)

// GeoPlaceSet says whether a street, postal code, locality and region belong to one locality of a
// country's geo tables; pass an empty region where the address prints none.
type GeoPlaceSet struct {
	name, region map[string]string
	locality     map[string]string
	streets      map[[2]string]bool
}

// GeoPlaces keys a locality by its code where its table has one, else by its name.
func GeoPlaces(t *testing.T, dir string) GeoPlaceSet {
	rows := func(table string) []map[string]string { return TSVRows(t, dir+"/"+table+".tsv") }
	s := GeoPlaceSet{name: map[string]string{}, region: map[string]string{}, locality: map[string]string{}, streets: map[[2]string]bool{}}
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

func (s GeoPlaceSet) Has(street, postalCode, locality, region string) bool {
	l, ok := s.locality[postalCode]
	return ok && s.name[l] == locality && (region == "" || s.region[l] == region) && s.streets[[2]string{street, l}]
}
