package fejktest

import "testing"

func TestGeoPlacesHoldOnlyOneLocalitysStreetPostalCodeNameAndRegion(t *testing.T) {
	s := GeoPlaceSet{
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
			if got := s.Has(c.street, c.postalCode, c.locality, c.region); got != c.want {
				t.Errorf("Has(%q, %q, %q, %q) = %v, want %v", c.street, c.postalCode, c.locality, c.region, got, c.want)
			}
		})
	}
}
