package engine

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/drawstate"
)

func TestAChildTableThatFailsItsLazyLoadIsNotLinkedToItsLoadedParent(t *testing.T) {
	root, err := Load([]datafiles.Source{datafiles.FS(fstest.MapFS{
		"country.json": {Data: []byte(`{"format":"{name}","rows":"country.tsv","key":"code"}`)},
		"country.tsv":  {Data: []byte("code\tname\nSE\tSweden\n")},
	})})
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	placeIndex(&root, datafiles.FS(fstest.MapFS{
		"city.json": {Data: []byte(`{"format":"{name}{/nope}","rows":"city.tsv","key":"name","parent":"country"}`)},
		"city.tsv":  {Data: []byte("name\tcountry\nStockholm\tSE\n")},
	}), map[string]datafiles.IndexEntry{"city": {Parent: "country", Paths: []string{"", "country", "name"}}})
	s := drawstate.New(1, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
	if _, err := Fake(s, &root, "city"); err == nil || !strings.Contains(err.Error(), "{/nope}") {
		t.Fatalf(`Fake("city") = %v, want an error naming {/nope}`, err)
	}
	if got, err := Fake(s, &root, "country.city.name"); err == nil {
		t.Fatalf(`Fake("country.city.name") = %q, want an error: city failed to load`, got)
	}
	if got := List(&root); slices.Contains(got, "country.city.name") {
		t.Fatalf("List() = %v, want no country.city.name", got)
	}
}
