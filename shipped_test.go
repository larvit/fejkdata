package fejkdata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/larvit/fejkdata/data"
	"github.com/larvit/fejkdata/data/en_US"
	"github.com/larvit/fejkdata/data/geo/SE"
	"github.com/larvit/fejkdata/data/geo/US"
	"github.com/larvit/fejkdata/data/misc"
	"github.com/larvit/fejkdata/data/sv_SE"
	"github.com/larvit/fejkdata/internal/datafiles"
)

// withoutManifest is fsys with its manifest hidden, so it loads whole in New.
type withoutManifest struct{ fs.FS }

func (w withoutManifest) Open(name string) (fs.File, error) {
	if name == datafiles.ManifestFile {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return w.FS.Open(name)
}

// shippedModule is a module data.Modules lists, and its folder under data/, slash-separated.
type shippedModule struct {
	fsys fs.FS
	dir  string
}

// importPath is the Go package the module's FS is exported from.
func (m shippedModule) importPath() string { return "github.com/larvit/fejkdata/data/" + m.dir }

// holds reports whether the category at path is the module's: its tree spells its folder.
func (m shippedModule) holds(path string) bool {
	return strings.HasPrefix(path, strings.ReplaceAll(m.dir, "/", ".")+".")
}

// shippedModules is every module data.Modules lists, each beside its folder under data/.
func shippedModules(t testing.TB) []shippedModule {
	t.Helper()
	out := []shippedModule{{en_US.FS, "en_US"}, {SE.FS, "geo/SE"}, {US.FS, "geo/US"}, {misc.FS, "misc"}, {sv_SE.FS, "sv_SE"}}
	var listed []fs.FS
	for _, m := range out {
		listed = append(listed, m.fsys)
	}
	if !slices.Equal(listed, data.Modules()) {
		t.Fatal("data.Modules() lists other modules than shippedModules does")
	}
	return out
}

// newShippedWhole loads every shipped category in New, as a source without an index does, so
// a test reads the whole shipped tree.
func newShippedWhole(t testing.TB, opts ...Option) *Generator {
	t.Helper()
	var modules []fs.FS
	for _, m := range shippedModules(t) {
		modules = append(modules, withoutManifest{m.fsys})
	}
	f, err := New(append([]Option{WithDataFS(modules...)}, opts...)...)
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

// REGENERATE=1 rewrites the manifests.
func TestShippedManifestsAreCurrent(t *testing.T) {
	modules := shippedModules(t)
	manifests, err := shippedManifests(newShippedWhole(t), modules)
	if err != nil {
		t.Fatal(err)
	}
	for i, m := range modules {
		file := filepath.Join("data", filepath.FromSlash(m.dir), datafiles.ManifestFile)
		if os.Getenv("REGENERATE") == "1" {
			if err := os.WriteFile(file, manifests[i], 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if string(manifests[i]) != string(want) {
			t.Errorf("%s is stale; regenerate it: docker compose run --rm --user \"$(id -u):$(id -g)\" generate", file)
		}
	}
}

// shippedManifests is each module's manifest, read from whole's tree: its index, one category a
// line, and every other module it reads, directly or through another.
func shippedManifests(whole *Generator, modules []shippedModule) ([][]byte, error) {
	owner := func(path string) int {
		return slices.IndexFunc(modules, func(m shippedModule) bool { return m.holds(path) })
	}
	index := make([][]string, len(modules))
	direct := make([][]int, len(modules))
	for _, s := range categorySites(&whole.root) {
		i := owner(s.path)
		if i < 0 {
			return nil, fmt.Errorf("no shipped module holds %s", s.path)
		}
		line, err := json.Marshal(indexEntry(s))
		if err != nil {
			return nil, err
		}
		key, _ := json.Marshal(s.path)
		index[i] = append(index[i], "\n\t\t"+string(key)+": "+string(line))
		for _, c := range needs(&whole.root, s) {
			if j := owner(categoryPath(c.dir, c.name)); j != i && !slices.Contains(direct[i], j) {
				direct[i] = append(direct[i], j)
			}
		}
	}
	out := make([][]byte, len(modules))
	for i := range modules {
		var b bytes.Buffer
		b.WriteString("{\n\t\"index\": {" + strings.Join(index[i], ",") + "\n\t}")
		if reads := readsOf(i, direct, modules); len(reads) > 0 {
			line, _ := json.Marshal(reads)
			b.WriteString(",\n\t\"reads\": " + string(line))
		}
		b.WriteString("\n}\n")
		out[i] = b.Bytes()
	}
	return out, nil
}

// readsOf is every module module i reads, directly or through another, by import path.
func readsOf(i int, direct [][]int, modules []shippedModule) []string {
	seen := map[int]bool{i: true}
	queue := slices.Clone(direct[i])
	var out []string
	for ; len(queue) > 0; queue = queue[1:] {
		j := queue[0]
		if seen[j] {
			continue
		}
		seen[j] = true
		out = append(out, modules[j].importPath())
		queue = append(queue, direct[j]...)
	}
	sort.Strings(out)
	return out
}

func shippedIndex(t *testing.T) map[string]datafiles.IndexEntry {
	t.Helper()
	index := map[string]datafiles.IndexEntry{}
	for _, m := range shippedModules(t) {
		manifest, err := datafiles.FS(m.fsys).Manifest()
		if err != nil {
			t.Fatal(err)
		}
		maps.Copy(index, manifest.Index)
	}
	return index
}

func TestEachShippedModuleNamesWhatItReads(t *testing.T) {
	for _, m := range shippedModules(t) {
		manifest, err := datafiles.FS(m.fsys).Manifest()
		if err != nil {
			t.Fatal(err)
		}
		if m.dir == "sv_SE" && !slices.Contains(manifest.Reads, "github.com/larvit/fejkdata/data/geo/SE") {
			t.Errorf("sv_SE reads %v, want geo/SE among them", manifest.Reads)
		}
		if slices.Contains(manifest.Reads, m.importPath()) {
			t.Errorf("%s reads %v, want itself left out", m.dir, manifest.Reads)
		}
	}
}

func TestEachShippedModuleRendersBesideTheModulesItReads(t *testing.T) {
	modules := shippedModules(t)
	for _, m := range modules {
		manifest, err := datafiles.FS(m.fsys).Manifest()
		if err != nil {
			t.Fatal(err)
		}
		set := []fs.FS{m.fsys}
		for _, read := range manifest.Reads {
			i := slices.IndexFunc(modules, func(r shippedModule) bool { return r.importPath() == read })
			if i < 0 {
				t.Fatalf("%s reads %s, which no shipped module is", m.dir, read)
			}
			set = append(set, modules[i].fsys)
		}
		f, err := New(WithDataFS(set...))
		if err != nil {
			t.Fatalf("New(%s with what it reads) = %v", m.dir, err)
		}
		for _, p := range sortedNames(manifest.Index) {
			if _, err := f.Fake(p); err != nil {
				t.Errorf("%s with what it reads: Fake(%q) = %v", m.dir, p, err)
			}
		}
	}
}

func TestAReadNoModuleProvidesNamesTheReadersModules(t *testing.T) {
	f, err := New(WithDataFS(sv_SE.FS))
	if err != nil {
		t.Fatalf("New(WithDataFS(sv_SE.FS)) = %v", err)
	}
	if _, err := f.Fake("sv_SE.address"); err == nil || !strings.Contains(err.Error(), "the module holding sv_SE.address reads by default github.com/larvit/fejkdata/data/geo/SE") {
		t.Errorf(`Fake("sv_SE.address") = %v, want the module it reads named`, err)
	}
	for name, opt := range map[string]Option{
		"no manifest":     WithDataPath(writeData(t, map[string]string{"x": `"{/y}"`})),
		"no reads in one": WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(`{"index": {"x": {"paths": [""]}}}`)}, "x.json": {Data: []byte(`"{/y}"`)}}),
	} {
		f, err := New(opt)
		if err == nil {
			_, err = f.Fake("x")
		}
		if err == nil || !strings.Contains(err.Error(), "the module holding x names no module it reads by default") {
			t.Errorf("%s: x = %v, want it said that x's module names none", name, err)
		}
	}
}

func TestADataPathManifestNamesWhatItReads(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".fejkdata.json": `{"reads": ["example.com/places"]}`,
		"x.json":         `"{/place}"`,
	})
	if _, err := New(WithDataPath(dir)); err == nil || !strings.Contains(err.Error(), "the module holding x reads by default example.com/places") {
		t.Errorf("New = %v, want the module x reads named", err)
	}
}

func TestAnIndexedModuleNamesWhatItReads(t *testing.T) {
	f, err := New(WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"x": {"paths": [""]}}, "reads": ["example.com/places"]}`)},
		"x.json":         {Data: []byte(`"{/place}"`)},
	}))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	if _, err := f.Fake("x"); err == nil || !strings.Contains(err.Error(), "the module holding x reads by default example.com/places") {
		t.Errorf(`Fake("x") = %v, want the module x reads named`, err)
	}
}

func TestNewLoadsNoShippedCategory(t *testing.T) {
	f, err := New(withShipped())
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
		f, err := New(withShipped())
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
	f, err := New(withShipped(), WithDataPath(writeData(t, map[string]string{"greeting": `"Hej {/sv_SE.person}!"`})))
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
		f, err := New(withShipped(), opt)
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

func TestAnIndexedSourceLoadsOnFirstReachLikeTheShippedOne(t *testing.T) {
	fsys := fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"a": {"paths": [""]}, "b": {"paths": [""]}, "sub.c": {"paths": ["", "d"]}}}`)},
		"a.json":         {Data: []byte(`"{/b}!"`)},
		"b.json":         {Data: []byte(`"x"`)},
		"sub/c.json":     {Data: []byte(`{"format":"{d}","d":"y"}`)},
	}
	f, err := New(WithDataFS(fsys), WithSeed(1))
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
	f, err := New(WithDataFS(fstest.MapFS{
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
		`{`:                                              {datafiles.ManifestFile},
		`{"index": {}} {}`:                               {"after top-level value"},
		`{"indx": {}, "INDEX": {}}`:                      {`unknown key "INDEX"`, `unknown key "indx"`},
		`{"index": []}`:                                  {"index must be an object, not a list"},
		`{"index": {"a..b": 5}}`:                         {`"a..b"`, "must be an object, not a number"},
		`{"index": {"a": {"Paths": [""]}}}`:              {`unknown key "Paths"`, "paths is missing"},
		`{"index": {"a": {}, "b": {}}}`:                  {`"a": paths is missing`, `"b": paths is missing`},
		`{"index": {"a": {"paths": ["x..y", 3]}}}`:       {`"x..y"`, "paths item 2 must be a string, not a number"},
		`{"index": {"a": {"paths": [""], "parent": 5}}}`: {"parent must be a string, not a number"},
		`{"index": {"a": {"paths": [""]}, "a.b": {"paths": [""]}}}`: {`index entry "a.b": "a" is a category, so it holds no other`},
		`{"index": {"a": {"paths": ["", "x", "x"]}}}`:               {`paths item 3 repeats "x"`},
		`{"index": {"a": {"paths": [""], "parent": "b"}}}`:          {`parent "b" names no entry`},
		`{"index": {"a": {"paths": [""], "reads": ["b"]}}}`:         {`unknown key "reads"`},
		`{"reads": "b"}`:     {"reads must be a list, not a string"},
		`{"reads": ["", 3]}`: {"reads item 1 is empty", "reads item 2 must be a string, not a number"},
	} {
		_, err := New(WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(manifest)}}))
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
		if _, err := New(withShipped(), opt); !errors.Is(err, ErrLoad) {
			t.Errorf("New(%s) = %v, want it to match ErrLoad", name, err)
		}
	}
}

func TestOnDemandRendersAsTheWholeLoad(t *testing.T) {
	whole := newShippedWhole(t)
	paths := whole.List()
	lazy, err := New(withShipped())
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if got := lazy.List(); !reflect.DeepEqual(got, paths) {
		t.Fatalf("List() on demand = %v, want %v", got, paths)
	}
	a, err := New(withShipped(), WithSeed(7))
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
		f, err := New(withShipped())
		if err != nil {
			t.Fatalf("New() = %v", err)
		}
		if _, err := f.Fake(p); err != nil {
			t.Errorf("Fake(%q) on a fresh New() = %v", p, err)
		}
	}
}

func TestListRunsBesideAnOnDemandLoad(t *testing.T) {
	f, err := New(withShipped())
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

func TestATableWhoseEntryOmitsItsParentFailsAtFirstReach(t *testing.T) {
	f, err := New(WithDataFS(fstest.MapFS{
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
	if first == nil || !strings.Contains(first.Error(), ".fejkdata.json") || !strings.Contains(first.Error(), "names no parent") {
		t.Fatalf(`Fake("city") = %v, want the entry's parent named`, first)
	}
	if _, again := f.Fake("city"); again == nil || again.Error() != first.Error() {
		t.Fatalf(`Fake("city") again = %v, want %v`, again, first)
	}
}

func TestADataPathTableUnderAShippedTableLoads(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1), WithDataPath(writeFiles(t, map[string]string{
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

func TestAnIndexedCategoryWithNoFileNamesTheManifest(t *testing.T) {
	f, err := New(WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"gone": {"paths": [""]}}}`)},
	}))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if _, err := f.Fake("gone"); err == nil || !strings.Contains(err.Error(), ".fejkdata.json indexes gone") {
		t.Fatalf(`Fake("gone") = %v, want the manifest naming gone`, err)
	}
}
