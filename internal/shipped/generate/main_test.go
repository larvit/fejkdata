package main

import (
	"bytes"
	"io/fs"
	"os"
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
	if len(got) != len(manifests)+3 {
		t.Fatalf("generated names %v, want the %d manifests and the three listing files", got, len(manifests))
	}
	generatedManifests := slices.Clone(got[:len(manifests)])
	slices.Sort(generatedManifests)
	slices.Sort(manifests)
	if !slices.Equal(generatedManifests, manifests) {
		t.Errorf("generated names the manifests %v, want %v", generatedManifests, manifests)
	}
	if listing, want := got[len(manifests):], []string{"data-listing/paths.json", "docs/shipped-paths.md", "npm-package/shipped-paths.d.ts"}; !slices.Equal(listing, want) {
		t.Errorf("generated ends with %v, want %v", listing, want)
	}
}

func TestGeneratedNamesAMissingKindsFile(t *testing.T) {
	files, err := generated(t.TempDir(), false)
	if err == nil || !strings.Contains(err.Error(), "kinds-and-sources.tsv") {
		t.Errorf("generated in a folder with no kinds file = %v, %v, want an error naming kinds-and-sources.tsv", names(files), err)
	}
}

func rootWithKinds(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	kinds, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(shipped.KindsFile)))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, filepath.FromSlash(shipped.KindsFile))), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(shipped.KindsFile)), kinds, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func assertWritten(t *testing.T, root string, files []shipped.File) {
	t.Helper()
	for _, f := range files {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Name)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, f.Body) {
			t.Errorf("run wrote %s differing from what generated returned", f.Name)
		}
	}
}

func TestRunWritesEveryGeneratedFileUnderTheRoot(t *testing.T) {
	root := rootWithKinds(t)
	files, err := generated(root, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, filepath.FromSlash(f.Name))), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{"-root", root}); err != nil {
		t.Fatal(err)
	}
	assertWritten(t, root, files)
}

func TestRunWithRepinWritesOnlyThePins(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "testdata"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-root", root, "-repin"}); err != nil {
		t.Fatal(err)
	}
	var written []string
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		written = append(written, e.Name())
	}
	if want := []string{"testdata"}; !slices.Equal(written, want) {
		t.Fatalf("run -repin left %v under the root, want %v", written, want)
	}
	pins, err := generated(repositoryRoot, true)
	if err != nil {
		t.Fatal(err)
	}
	assertWritten(t, root, pins)
}

func TestRunNamesAFileItCannotWrite(t *testing.T) {
	err := run([]string{"-root", rootWithKinds(t)})
	if err == nil || !strings.Contains(err.Error(), ".fejkdata.json") {
		t.Errorf("run into a root with no data folders = %v, want an error naming a manifest", err)
	}
}

func TestRunPassesOnAGenerationError(t *testing.T) {
	err := run([]string{"-root", t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "kinds-and-sources.tsv") {
		t.Errorf("run in a folder with no kinds file = %v, want an error naming kinds-and-sources.tsv", err)
	}
}

func TestRunRefusesAnUnknownFlag(t *testing.T) {
	err := run([]string{"-nothing"})
	if err == nil || !strings.Contains(err.Error(), "nothing") {
		t.Errorf("run with an unknown flag = %v, want an error naming it", err)
	}
}

func TestGeneratedNamesAnIncompleteListing(t *testing.T) {
	root := rootWithKinds(t)
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(shipped.KindsFile)), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := generated(root, false)
	if err == nil || !strings.Contains(err.Error(), "the listing is not whole") {
		t.Errorf("generated with an empty kinds file = %v, want the listing is not whole", err)
	}
}

func TestRunAnswersHelpWithoutWritingAnything(t *testing.T) {
	root := t.TempDir()
	if err := run([]string{"-h", "-root", root}); err != nil {
		t.Errorf("run -h = %v, want nil", err)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Errorf("run -h wrote %d entries under the root", len(entries))
	}
}
