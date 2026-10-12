package fejkdata

import (
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
	enus "github.com/larvit/fejkdata/data/en_US"
	geose "github.com/larvit/fejkdata/data/geo/SE"
	geous "github.com/larvit/fejkdata/data/geo/US"
	"github.com/larvit/fejkdata/data/misc"
	svse "github.com/larvit/fejkdata/data/sv_SE"
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

// shippedModule is a module data.Modules lists, its folder under data/, and the namespace its
// manifest mounts it under.
type shippedModule struct {
	fsys      fs.FS
	modules   func() []fs.FS
	dir       string
	namespace string
}

// holds reports whether the category at path is the module's.
func (m shippedModule) holds(path string) bool { return strings.HasPrefix(path, m.namespace+".") }

// whole is the module's files with a manifest naming only its namespace, so it loads whole in New.
func (m shippedModule) whole(t testing.TB) fs.FS {
	t.Helper()
	out := fstest.MapFS{datafiles.ManifestFile: {Data: []byte(`{"namespace": "` + m.namespace + `"}`)}}
	err := fs.WalkDir(m.fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || p == datafiles.ManifestFile {
			return err
		}
		b, err := fs.ReadFile(m.fsys, p)
		out[p] = &fstest.MapFile{Data: b}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// shippedModules is every module data.Modules lists, each beside its folder under data/: its
// namespace, a dot a slash.
func shippedModules(t testing.TB) []shippedModule {
	t.Helper()
	out := []shippedModule{
		{fsys: enus.FS, modules: enus.Modules, namespace: "en_US"},
		{fsys: geose.FS, modules: geose.Modules, namespace: "geo.SE"},
		{fsys: geous.FS, modules: geous.Modules, namespace: "geo.US"},
		{fsys: misc.FS, modules: misc.Modules, namespace: "misc"},
		{fsys: svse.FS, modules: svse.Modules, namespace: "sv_SE"},
	}
	for i := range out {
		out[i].dir = strings.ReplaceAll(out[i].namespace, ".", "/")
	}
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
		modules = append(modules, m.whole(t))
	}
	f, err := New(append([]Option{WithDataFS(modules...), WithSeed(1)}, opts...)...)
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
	manifests, _, err := shippedManifests(newShippedWhole(t), modules)
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
// line, and its namespace; and the namespace of every other module each reads, directly or
// through another.
func shippedManifests(whole *Generator, modules []shippedModule) ([][]byte, [][]string, error) {
	owner := func(path string) int {
		return slices.IndexFunc(modules, func(m shippedModule) bool { return m.holds(path) })
	}
	index := make([][]string, len(modules))
	direct := make([][]int, len(modules))
	for _, s := range categorySites(&whole.root) {
		i := owner(s.path)
		if i < 0 {
			return nil, nil, fmt.Errorf("no shipped module holds %s", s.path)
		}
		line, err := json.Marshal(indexEntry(s))
		if err != nil {
			return nil, nil, err
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
	reads := make([][]string, len(modules))
	for i := range modules {
		namespace, _ := json.Marshal(modules[i].namespace)
		out[i] = []byte("{\n\t\"index\": {" + strings.Join(index[i], ",") + "\n\t},\n\t\"namespace\": " + string(namespace) + "\n}\n")
		reads[i] = readsOf(i, direct, modules)
	}
	return out, reads, nil
}

// readsOf is the namespace of every module module i reads, directly or through another.
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
		out = append(out, modules[j].namespace)
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

func TestEachShippedPackagesModulesHoldWhatItReadsAndRender(t *testing.T) {
	modules := shippedModules(t)
	_, reads, err := shippedManifests(newShippedWhole(t), modules)
	if err != nil {
		t.Fatal(err)
	}
	for i, m := range modules {
		set := m.modules()
		for _, read := range reads[i] {
			j := slices.IndexFunc(modules, func(r shippedModule) bool { return r.namespace == read })
			if j < 0 || !slices.Contains(set, modules[j].fsys) {
				t.Errorf("%s reads %s, which its Modules leaves out", m.namespace, read)
			}
		}
		manifest, err := datafiles.FS(m.fsys).Manifest()
		if err != nil {
			t.Fatal(err)
		}
		f, err := New(WithSeed(1), WithDataFS(set...))
		if err != nil {
			t.Fatalf("New(%s's Modules) = %v", m.namespace, err)
		}
		for _, p := range sortedNames(manifest.Index) {
			if _, err := f.Fake(p); err != nil {
				t.Errorf("%s's Modules: Fake(%q) = %v", m.namespace, p, err)
			}
		}
	}
}

func TestALocaleAloneFailsAtItsAddressNamingTheMissingGeo(t *testing.T) {
	f, err := New(WithSeed(1), WithDataFS(svse.FS))
	if err != nil {
		t.Fatalf("New(WithDataFS(svse.FS)) = %v", err)
	}
	if _, err := f.Fake("sv_SE.address"); !errors.Is(err, ErrLoad) || !strings.Contains(err.Error(), `reference {/geo.SE.address}: no entry "geo"`) {
		t.Errorf(`Fake("sv_SE.address") = %v, want a load error naming the missing geo`, err)
	}
	if got := fake(t, f, "sv_SE.person"); got == "" {
		t.Error("sv_SE.person rendered empty, want the locale's other categories to load alone")
	}
}

func TestAMountedModuleNamesItsRootFolderByItsNamespace(t *testing.T) {
	_, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"namespace": "my.data"}`)},
		"t.json":         {Data: []byte(`{"format":"{name}","rows":"t.tsv","key":"name"}`)},
	}))
	if err == nil || !strings.Contains(err.Error(), "not beside it in my/data") {
		t.Errorf("New = %v, want the root folder named by its namespace's folders", err)
	}
}

func TestAMountedModuleNamesItsFilesByItsNamespace(t *testing.T) {
	_, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"namespace": "my.data"}`)},
		"word.json":      {Data: []byte(`{`)},
	}))
	if err == nil || !strings.Contains(err.Error(), "my/data/word.json") {
		t.Errorf("New = %v, want the broken file named under its namespace", err)
	}
}

func TestANamespaceMountsAModuleUnderIt(t *testing.T) {
	for name, opt := range map[string]Option{
		"a data path": WithDataPath(writeFiles(t, map[string]string{
			".fejkdata.json": `{"namespace": "my.data"}`,
			"word.json":      `"hej"`,
			"sub/word.json":  `"hej"`,
		})),
		"an indexed source": WithDataFS(fstest.MapFS{
			".fejkdata.json": {Data: []byte(`{"index": {"my.data.word": {"paths": [""]}, "my.data.sub.word": {"paths": [""]}}, "namespace": "my.data"}`)},
			"word.json":      {Data: []byte(`"hej"`)},
			"sub/word.json":  {Data: []byte(`"hej"`)},
		}),
	} {
		f, err := New(opt, WithSeed(1))
		if err != nil {
			t.Fatalf("New(%s) = %v", name, err)
		}
		for _, p := range []string{"my.data.word", "my.data.sub.word"} {
			if got := fake(t, f, p); got != "hej" {
				t.Errorf("%s: %s = %q, want hej", name, p, got)
			}
		}
	}
}

func TestNewLoadsNoShippedCategory(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
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
		f, err := New(withShipped(), WithSeed(1))
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
	f, err := New(withShipped(), WithSeed(1), WithDataPath(writeData(t, map[string]string{"greeting": `"Hej {/sv_SE.person}!"`})))
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
		f, err := New(withShipped(), WithSeed(1), opt)
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
	f, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
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
		`{"reads": ["b"]}`:      {`unknown key "reads"; a manifest holds index and namespace`},
		`{"namespace": 5}`:      {"namespace must be a string, not a number"},
		`{"namespace": "a..b"}`: {`namespace "a..b"`},
		`{"index": {"b.x": {"paths": [""]}}, "namespace": "a"}`: {`index entry "b.x" lies outside the namespace "a"`},
	} {
		_, err := New(WithSeed(1), WithDataFS(fstest.MapFS{".fejkdata.json": {Data: []byte(manifest)}}))
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
		if _, err := New(withShipped(), WithSeed(1), opt); !errors.Is(err, ErrLoad) {
			t.Errorf("New(%s) = %v, want it to match ErrLoad", name, err)
		}
	}
}

func TestOnDemandRendersAsTheWholeLoad(t *testing.T) {
	whole := newShippedWhole(t)
	paths := whole.List()
	lazy, err := New(withShipped(), WithSeed(1))
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
		f, err := New(withShipped(), WithSeed(1))
		if err != nil {
			t.Fatalf("New() = %v", err)
		}
		if _, err := f.Fake(p); err != nil {
			t.Errorf("Fake(%q) on a fresh New() = %v", p, err)
		}
	}
}

func TestListRunsBesideAnOnDemandLoad(t *testing.T) {
	f, err := New(withShipped(), WithSeed(1))
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
	f, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
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

func TestAMountedIndexedCategoryWithNoFileNamesTheManifestBesideItsFiles(t *testing.T) {
	f, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"my.gone": {"paths": [""]}}, "namespace": "my"}`)},
	}))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if _, err := f.Fake("my.gone"); err == nil || !strings.Contains(err.Error(), "my/.fejkdata.json indexes my.gone") {
		t.Fatalf(`Fake("my.gone") = %v, want the manifest named as my/.fejkdata.json`, err)
	}
}

func TestAnIndexedCategoryWithNoFileNamesTheManifest(t *testing.T) {
	f, err := New(WithSeed(1), WithDataFS(fstest.MapFS{
		".fejkdata.json": {Data: []byte(`{"index": {"gone": {"paths": [""]}}}`)},
	}))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if _, err := f.Fake("gone"); err == nil || !strings.Contains(err.Error(), ".fejkdata.json indexes gone") {
		t.Fatalf(`Fake("gone") = %v, want the manifest naming gone`, err)
	}
}
