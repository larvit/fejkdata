package engine

import (
	"testing"
	"testing/fstest"

	"github.com/larvit/fejkdata/internal/datafiles"
)

func TestDrawnFromIsTheLastTableAPathStepsThroughElseTheCategoryItEnters(t *testing.T) {
	root, err := Load([]datafiles.Source{datafiles.FS(fstest.MapFS{
		"geo/country.json": {Data: []byte(`{"format":"{name}","rows":"country.tsv","key":"code"}`)},
		"geo/country.tsv":  {Data: []byte("code\tname\nSE\tSweden\n")},
		"geo/city.json":    {Data: []byte(`{"format":"{name}","rows":"city.tsv","key":"name","parent":"country"}`)},
		"geo/city.tsv":     {Data: []byte("name\tcountry\nStockholm\tSE\n")},
		"geo/street.json":  {Data: []byte(`{"format":"{name}","rows":"street.tsv","key":"name","parent":"city"}`)},
		"geo/street.tsv":   {Data: []byte("name\tcity\nStorgatan\tStockholm\n")},
		"city.json":        {Data: []byte(`"a city in no table"`)},
		"person.json":      {Data: []byte(`{"format":"{first} from {home}","first":["Sara","Lina"],"home":"{/geo.country.city}"}`)},
	})})
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	for path, want := range map[string]string{
		"geo.country":                  "geo.country",
		"geo.country.code":             "geo.country",
		"geo.country.city":             "geo.city",
		"geo.country.city.name":        "geo.city",
		"geo.country.city.street":      "geo.street",
		"geo.country.city.street.name": "geo.street",
		"geo.country[SE].city":         "geo.city",
		"geo.city":                     "geo.city",
		"city":                         "city",
		"person":                       "person",
		"person.first":                 "person",
		"person.home":                  "person",
		"/person.home":                 "person",
	} {
		t.Run(path, func(t *testing.T) {
			if got, err := DrawnFrom(&root, path); err != nil || got != want {
				t.Errorf("DrawnFrom(%q) = %q, %v, want %q", path, got, err, want)
			}
		})
	}
	for _, path := range []string{"geo", "geo.country.nope", "nope"} {
		t.Run(path, func(t *testing.T) {
			if got, err := DrawnFrom(&root, path); err == nil {
				t.Errorf("DrawnFrom(%q) = %q, want an error: it names no value", path, got)
			}
		})
	}
}

func TestDrawnFromFindsATableThatIsTheFirstCategory(t *testing.T) {
	root, err := Load([]datafiles.Source{datafiles.FS(fstest.MapFS{
		"geo/city.json":    {Data: []byte(`{"format":"{name}","rows":"city.tsv","key":"name","parent":"country"}`)},
		"geo/city.tsv":     {Data: []byte("name\tcountry\nStockholm\tSE\n")},
		"geo/country.json": {Data: []byte(`{"format":"{name}","rows":"country.tsv","key":"code"}`)},
		"geo/country.tsv":  {Data: []byte("code\tname\nSE\tSweden\n")},
	})})
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if got, err := DrawnFrom(&root, "geo.country.city"); err != nil || got != "geo.city" {
		t.Errorf("DrawnFrom(%q) = %q, %v, want %q", "geo.country.city", got, err, "geo.city")
	}
}
