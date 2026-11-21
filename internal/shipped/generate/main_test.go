package main

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/shipped"
)

const repositoryRoot = "../../.."

func names(files []shipped.File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

func TestGeneratedWithRepinIsTheTwoPins(t *testing.T) {
	files, err := generated(repositoryRoot, true)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := names(files), []string{"testdata/shipped_shape.txt", "testdata/listing_shape.txt"}; !slices.Equal(got, want) {
		t.Errorf("generated with repin names %v, want %v", got, want)
	}
}

func TestGeneratedIsEachManifestThenTheListing(t *testing.T) {
	var manifests []string
	err := filepath.WalkDir(filepath.Join(repositoryRoot, "data"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Name() == ".fejkdata.json" {
			rel, _ := filepath.Rel(repositoryRoot, p)
			manifests = append(manifests, filepath.ToSlash(rel))
		}
		return err
	})
	if err != nil || len(manifests) == 0 {
		t.Fatalf("found manifests %v under data/: %v", manifests, err)
	}
	files, err := generated(repositoryRoot, false)
	if err != nil {
		t.Fatal(err)
	}
	got := names(files)
	if len(got) != len(manifests)+2 {
		t.Fatalf("generated names %v, want the %d manifests and the two listing files", got, len(manifests))
	}
	generatedManifests := slices.Clone(got[:len(manifests)])
	slices.Sort(generatedManifests)
	slices.Sort(manifests)
	if !slices.Equal(generatedManifests, manifests) {
		t.Errorf("generated names the manifests %v, want %v", generatedManifests, manifests)
	}
	if listing, want := got[len(manifests):], []string{"data-listing/paths.json", "docs/shipped-paths.md"}; !slices.Equal(listing, want) {
		t.Errorf("generated ends with %v, want %v", listing, want)
	}
}

func TestGeneratedNamesAMissingKindsFile(t *testing.T) {
	files, err := generated(t.TempDir(), false)
	if err == nil || !strings.Contains(err.Error(), "kinds-and-sources.tsv") {
		t.Errorf("generated in a folder with no kinds file = %v, %v, want an error naming kinds-and-sources.tsv", names(files), err)
	}
}
