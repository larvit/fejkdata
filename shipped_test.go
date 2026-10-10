package fejkdata

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
		line, err := json.Marshal(indexEntry(s))
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
	m, err := shippedSource.Manifest()
	if err != nil {
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

func TestAReplacementBreakingAShippedReaderFailsItsFirstReach(t *testing.T) {
	for name, opt := range map[string]Option{
		"a data path": WithDataPath(writeData(t, map[string]string{"sv_SE/last-name": `"x"`})),
		"an indexed source": WithDataFS(fstest.MapFS{
			".fejkdata.json":       {Data: []byte(`{"index": {"sv_SE.last-name": {"paths": [""]}}}`)},
			"sv_SE/last-name.json": {Data: []byte(`"x"`)},
		}),
	} {
		f, err := New(opt)
		if err != nil {
			t.Fatalf("New(%s) = %v", name, err)
		}
		if got := loadedCategories(f); slices.Contains(got, "sv_SE.person") {
			t.Errorf("New(%s) loaded %v, want sv_SE.person left for its first reach", name, got)
		}
		if _, err := f.Fake("sv_SE.person"); !errors.Is(err, ErrLoad) || !strings.Contains(err.Error(), "sv_SE.person") {
			t.Errorf(`%s: Fake("sv_SE.person") = %v, want ErrLoad naming sv_SE.person`, name, err)
		}
	}
}

func TestAnIndexedSourceStandsLikeTheShippedOne(t *testing.T) {
	fsys := fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"a": {"paths": [""]}, "b": {"paths": [""]}, "sub.c": {"paths": ["", "d"]}}}`)},
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
	if !errors.Is(first, ErrLoad) {
		t.Fatalf(`Fake("bad") = %v, want it to wrap ErrLoad`, first)
	}
	if _, again := f.Fake("bad"); again == nil || again.Error() != first.Error() {
		t.Fatalf(`Fake("bad") again = %v, want %v`, again, first)
	}
	if got := fake(t, f, "good"); got != "fine" {
		t.Fatalf(`Fake("good") = %q`, got)
	}
}

func TestABrokenManifestFailsNewNamingEveryMistake(t *testing.T) {
	for manifest, want := range map[string][]string{
		`{`:                                                 {datafiles.ManifestFile},
		`{"index": {}} {}`:                                  {"after top-level value"},
		`{"indx": {}, "INDEX": {}}`:                         {`unknown key "INDEX"`, `unknown key "indx"`},
		`{"index": []}`:                                     {"index must be an object, not an array"},
		`{"index": {"a..b": 5}}`:                            {`"a..b"`, "must be an object, not a number"},
		`{"index": {"a": {"Paths": [""]}}}`:                 {`unknown key "Paths"`, "paths is missing"},
		`{"index": {"a": {}, "b": {}}}`:                     {`"a": paths is missing`, `"b": paths is missing`},
		`{"index": {"a": {"paths": ["x..y", 3]}}}`:          {`"x..y"`, "paths item 2 must be a string, not a number"},
		`{"index": {"a": {"paths": [""], "parent": 5}}}`:    {"parent must be a string, not a number"},
		`{"index": {"a": {"paths": [""], "parent": "b"}}}`:  {`parent "b" names no entry`},
		`{"index": {"a": {"paths": [""], "reads": ["b"]}}}`: {`unknown key "reads"`},
	} {
		_, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(manifest)}}))
		for _, w := range want {
			if err == nil || !strings.Contains(err.Error(), w) {
				t.Errorf("New(manifest %s) = %v, want an error naming %s", manifest, err, w)
			}
		}
		if err != nil && strings.Count(err.Error(), "\n") >= len(want) {
			t.Errorf("New(manifest %s) = %v, want %d mistakes, each once", manifest, err, len(want))
		}
	}
}

func TestNewMatchesErrLoadOnDataThatFailsToLoad(t *testing.T) {
	for name, opt := range map[string]Option{
		"a missing data path": WithDataPath(filepath.Join(t.TempDir(), "nope")),
		"a broken category":   WithDataPath(writeData(t, map[string]string{"x": `{ not json`})),
		"a broken manifest":   WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(`{`)}}),
	} {
		if _, err := New(opt); !errors.Is(err, ErrLoad) {
			t.Errorf("New(%s) = %v, want it to match ErrLoad", name, err)
		}
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

func TestAParentColumnItsIndexEntryOmitsFailsTheFirstReachNamingTheManifest(t *testing.T) {
	f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"city": {"paths": ["", "country", "name"]}, "country": {"paths": ["", "code", "name"]}}}`)},
		"country.json":   {Data: []byte(`{"format":"{name}","rows":"country.tsv","key":"code"}`)},
		"country.tsv":    {Data: []byte("code\tname\nSE\tSweden\n")},
		"city.json":      {Data: []byte(`{"format":"{name}","rows":"city.tsv","key":"name","parent":"country"}`)},
		"city.tsv":       {Data: []byte("name\tcountry\nOslo\tSE\n")},
	}))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if _, err := f.Fake("country"); err != nil {
		t.Fatalf(`Fake("country") = %v`, err)
	}
	_, first := f.Fake("city")
	if first == nil || !strings.Contains(first.Error(), ".fejkdata.json") || !strings.Contains(first.Error(), `parent ""`) {
		t.Fatalf(`Fake("city") = %v, want the entry's parent named`, first)
	}
	if _, again := f.Fake("city"); again == nil || again.Error() != first.Error() {
		t.Fatalf(`Fake("city") again = %v, want %v`, again, first)
	}
}

func TestADataPathTableUnderAShippedTableLoads(t *testing.T) {
	f, err := New(WithSeed(1), WithDataPath(writeFiles(t, map[string]string{
		"sv_SE/nickname.json": `{"format":"{name}","rows":"nickname.tsv","key":"name","parent":"sex"}`,
		"sv_SE/nickname.tsv":  "name\tsex\nKalle\tm\nLotta\tf\n",
	})))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if got := fake(t, f, "sv_SE.sex[m].nickname"); got != "Kalle" {
		t.Fatalf("sv_SE.sex[m].nickname = %q, want Kalle", got)
	}
	if got := fake(t, f, "sv_SE.person"); got == "" {
		t.Fatal("sv_SE.person rendered empty")
	}
}

// indexEntry is what an index says of the loaded category s.
func indexEntry(s categorySite) datafiles.IndexEntry {
	e := datafiles.IndexEntry{Parent: parentOf(s.n), Paths: paths(s.n, false)}
	sort.Strings(e.Paths)
	return e
}
