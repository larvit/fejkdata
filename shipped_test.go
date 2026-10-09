package fejkdata

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/larvit/fejkdata/internal/datafiles"
)

var shippedManifestFile = path.Join("data", datafiles.ManifestFile)

// withoutManifest is fsys with its manifest hidden, so it loads whole in New.
type withoutManifest struct{ fs.FS }

func (w withoutManifest) Open(name string) (fs.File, error) {
	if name == datafiles.ManifestFile {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return w.FS.Open(name)
}

// newShippedWhole loads every shipped category in New, as a source without an index does, so
// a test reads the whole shipped tree.
func newShippedWhole(t testing.TB, opts ...Option) *Generator {
	t.Helper()
	data, err := fs.Sub(shippedFS, "data")
	if err != nil {
		t.Fatal(err)
	}
	f, err := New(append([]Option{WithoutShippedData(), WithDataFS(withoutManifest{data})}, opts...)...)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	return f
}

// loadedCategories is the dot path of every category f has parsed and resolved.
func loadedCategories(f *Generator) []string {
	var out []string
	for _, s := range categorySites(&f.root) {
		out = append(out, s.path)
	}
	return out
}

// REGENERATE=1 rewrites the manifest.
func TestShippedManifestIsCurrent(t *testing.T) {
	got, err := shippedManifest(newShippedWhole(t))
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("REGENERATE") == "1" {
		if err := os.WriteFile(shippedManifestFile, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(shippedManifestFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s is stale; regenerate it: docker compose run --rm --user \"$(id -u):$(id -g)\" generate", shippedManifestFile)
	}
}

// shippedManifest is the manifest of whole's tree, one category a line.
func shippedManifest(whole *Generator) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{\n\t\"index\": {")
	for i, s := range categorySites(&whole.root) {
		e := datafiles.IndexEntry{Paths: paths(s.n, false), Reads: categoryReads(&whole.root, s)}
		if t, isTable := s.n.(*table); isTable {
			e.Parent = t.rows.Options().Parent
		}
		sort.Strings(e.Paths)
		line, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			b.WriteString(",")
		}
		key, _ := json.Marshal(s.path)
		b.WriteString("\n\t\t" + string(key) + ": " + string(line))
	}
	b.WriteString("\n\t}\n}\n")
	return b.Bytes(), nil
}

func shippedIndex(t *testing.T) map[string]datafiles.IndexEntry {
	t.Helper()
	b, err := os.ReadFile(shippedManifestFile)
	if err != nil {
		t.Fatal(err)
	}
	var m datafiles.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m.Index
}

func TestNewLoadsNoShippedCategory(t *testing.T) {
	f, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if got := loadedCategories(f); len(got) > 0 {
		t.Fatalf("New() loaded %v", got)
	}
}

func TestReachingACategoryLoadsWhatItReads(t *testing.T) {
	for _, tc := range []struct {
		reach        func(f *Generator) error
		want, absent []string
	}{
		{
			reach: func(f *Generator) error { _, err := f.Fake("sv_SE.person.first"); return err },
			want:  []string{"sv_SE.first-name", "sv_SE.last-name", "sv_SE.person", "sv_SE.sex", "sv_SE.title"},
		},
		{
			reach:  func(f *Generator) error { _, err := f.Fake("geo.SE.region"); return err },
			want:   []string{"geo.SE.locality", "geo.SE.municipality", "geo.SE.postal-code", "geo.SE.region", "geo.SE.street"},
			absent: []string{"geo.SE.address", "geo.US.region"},
		},
		{
			reach:  func(f *Generator) error { _, err := f.FakeTemplate("{/misc.territory[SE].name}"); return err },
			want:   []string{"misc.territory", "misc.timezone"},
			absent: []string{"misc.tld"},
		},
		{
			reach:  func(f *Generator) error { _, err := f.FakeRecord("misc.uuid"); return err },
			want:   []string{"misc.uuid"},
			absent: []string{"misc.objectid"},
		},
		{
			reach: func(f *Generator) error {
				var v struct {
					Name string `fake:"sv_SE.first-name"`
				}
				return f.FakeStruct(&v)
			},
			want:   []string{"sv_SE.first-name"},
			absent: []string{"sv_SE.person"},
		},
	} {
		f, err := New()
		if err != nil {
			t.Fatalf("New() = %v", err)
		}
		if err := tc.reach(f); err != nil {
			t.Fatal(err)
		}
		got := loadedCategories(f)
		for _, w := range tc.want {
			if !slices.Contains(got, w) {
				t.Errorf("loaded %v, want %s among them", got, w)
			}
		}
		for _, a := range tc.absent {
			if slices.Contains(got, a) {
				t.Errorf("loaded %v, want %s left unloaded", got, a)
			}
		}
	}
}

func TestADataPathLoadsWhatItReads(t *testing.T) {
	f, err := New(WithDataPath(writeData(t, map[string]string{"greeting": `"Hej {/sv_SE.person}!"`})))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	got := loadedCategories(f)
	for _, w := range []string{"greeting", "sv_SE.first-name", "sv_SE.person"} {
		if !slices.Contains(got, w) {
			t.Errorf("New loaded %v, want %s among them", got, w)
		}
	}
	for _, a := range []string{"misc.uuid", "sv_SE.address", "geo.SE.region"} {
		if slices.Contains(got, a) {
			t.Errorf("New loaded %v, want %s left unloaded", got, a)
		}
	}
}

func TestADataPathLoadsTheShippedCategoriesReachingWhatItReplaces(t *testing.T) {
	f, err := New(WithDataPath(writeFiles(t, map[string]string{
		"sv_SE/last-name.json": `{"format":"{name}","rows":"last-name.tsv","key":"name"}`,
		"sv_SE/last-name.tsv":  "name\nSvensson\n",
	})))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	got := loadedCategories(f)
	for _, w := range []string{"sv_SE.last-name", "sv_SE.person"} {
		if !slices.Contains(got, w) {
			t.Errorf("New loaded %v, want %s among them", got, w)
		}
	}
	if slices.Contains(got, "misc.uuid") {
		t.Errorf("New loaded %v, want misc.uuid left unloaded", got)
	}
	if _, err := New(WithDataPath(writeData(t, map[string]string{"sv_SE/last-name": `"x"`}))); err == nil || !strings.Contains(err.Error(), "sv_SE.person") {
		t.Fatalf("New(a last-name the shipped person cannot read) = %v, want an error naming sv_SE.person", err)
	}
}

func TestAnIndexedSourceStandsLikeTheShippedOne(t *testing.T) {
	fsys := fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"a": {"paths": [""], "reads": ["b"]}, "b": {"paths": [""]}, "sub.c": {"paths": ["", "d"]}}}`)},
		"a.json":         {Data: []byte(`"{/b}!"`)},
		"b.json":         {Data: []byte(`"x"`)},
		"sub/c.json":     {Data: []byte(`{"format":"{d}","d":"y"}`)},
	}
	f, err := New(WithoutShippedData(), WithDataFS(fsys), WithSeed(1))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if got := loadedCategories(f); len(got) > 0 {
		t.Fatalf("New() loaded %v", got)
	}
	if got, want := f.List(), []string{"a", "b", "sub.c", "sub.c.d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
	if got := fake(t, f, "a"); got != "x!" {
		t.Fatalf(`Fake("a") = %q, want "x!"`, got)
	}
	if got, want := loadedCategories(f), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded %v, want %v", got, want)
	}
}

func TestAnIndexedCategoryFailsAtFirstReachAndEveryReachAfter(t *testing.T) {
	f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"bad": {"paths": [""]}, "good": {"paths": [""]}}}`)},
		"bad.json":       {Data: []byte(`"{/nope}"`)},
		"good.json":      {Data: []byte(`"fine"`)},
	}))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	_, first := f.Fake("bad")
	if first == nil || !strings.Contains(first.Error(), "{/nope}") {
		t.Fatalf(`Fake("bad") = %v, want an error naming {/nope}`, first)
	}
	if _, again := f.Fake("bad"); again == nil || again.Error() != first.Error() {
		t.Fatalf(`Fake("bad") again = %v, want %v`, again, first)
	}
	if got := fake(t, f, "good"); got != "fine" {
		t.Fatalf(`Fake("good") = %q`, got)
	}
}

func TestABrokenManifestFailsNew(t *testing.T) {
	for manifest, want := range map[string]string{
		`{`:                                    datafiles.ManifestFile,
		`{"indx": {}}`:                         "indx",
		`{"index": {"a..b": {"paths": [""]}}}`: "a..b",
	} {
		_, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(manifest)}}))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("New(manifest %s) = %v, want an error naming %s", manifest, err, want)
		}
	}
}

func TestAnIndexedSourceReplacingAShippedCategoryLoadsItsReadersInNew(t *testing.T) {
	_, err := New(WithDataFS(fstest.MapFS{
		".fejkdata.json":       {Data: []byte(`{"index": {"sv_SE.last-name": {"paths": [""]}}}`)},
		"sv_SE/last-name.json": {Data: []byte(`"x"`)},
	}))
	if err == nil || !strings.Contains(err.Error(), "sv_SE.person") {
		t.Fatalf("New() = %v, want an error naming sv_SE.person", err)
	}
}

func TestOnDemandRendersAsTheWholeLoad(t *testing.T) {
	whole := newShippedWhole(t)
	paths := whole.List()
	lazy, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if got := lazy.List(); !reflect.DeepEqual(got, paths) {
		t.Fatalf("List() on demand = %v, want %v", got, paths)
	}
	a, err := New(WithSeed(7))
	if err != nil {
		t.Fatal(err)
	}
	b := newShippedWhole(t, WithSeed(7))
	for _, p := range paths {
		for range 3 {
			got, gotErr := a.Fake(p)
			want, wantErr := b.Fake(p)
			if got != want || (gotErr == nil) != (wantErr == nil) {
				t.Fatalf("Fake(%q) on demand = %q, %v; whole = %q, %v", p, got, gotErr, want, wantErr)
			}
		}
	}
}

func TestEveryShippedCategoryLoadsAlone(t *testing.T) {
	for _, p := range sortedNames(shippedIndex(t)) {
		f, err := New()
		if err != nil {
			t.Fatalf("New() = %v", err)
		}
		if _, err := f.Fake(p); err != nil {
			t.Errorf("Fake(%q) on a fresh New() = %v", p, err)
		}
	}
}

func TestListRunsBesideAnOnDemandLoad(t *testing.T) {
	f, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	var wg sync.WaitGroup
	for _, p := range []string{"sv_SE.person", "geo.US.address", "misc.territory"} {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := f.Fake(p); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			f.List()
		}()
	}
	wg.Wait()
}

// categoryReads is the sorted path of every other category s's templates reference.
func categoryReads(root *folder, s categorySite) []string {
	seen := map[string]bool{s.path: true}
	var out []string
	for _, c := range referenced(root, s.dir, siteNodes([]categorySite{s})) {
		if p := c.path(); !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
