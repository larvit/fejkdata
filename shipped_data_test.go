package fejkdata

import (
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

func TestNewDefaultsToShippedData(t *testing.T) {
	f, err := New(WithSeed(1))
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	paths := f.List()
	for _, p := range []string{"sv_SE.person", "en_US.address", "misc.uuid"} {
		if !slices.Contains(paths, p) {
			t.Errorf("List() omits shipped %q", p)
		}
	}
	if got := fake(t, f, "sv_SE.person"); got == "" {
		t.Error("sv_SE.person rendered empty")
	}
}

func TestWithDataPathLayersOverShipped(t *testing.T) {
	dir := writeData(t, map[string]string{"sv_SE/word": `"only-mine"`, "greeting": `"hej"`})
	f, err := New(WithDataPath(dir), WithSeed(1))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	if got := fake(t, f, "sv_SE.word"); got != "only-mine" {
		t.Errorf("sv_SE.word = %q, want the layered file to win", got)
	}
	if got := fake(t, f, "sv_SE.person"); got == "" {
		t.Error("sv_SE.person should still come from the shipped data")
	}
	if got := fake(t, f, "greeting"); got != "hej" {
		t.Errorf("greeting = %q, want hej", got)
	}
}

func TestUserDataMayReferenceShipped(t *testing.T) {
	dir := writeData(t, map[string]string{"greeting": `"Hej {/sv_SE.person}!"`})
	f, err := New(WithDataPath(dir), WithSeed(1))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	if got := fake(t, f, "greeting"); !strings.HasPrefix(got, "Hej ") || !strings.HasSuffix(got, "!") {
		t.Errorf("greeting = %q", got)
	}
}

func TestAGeneratorWithDataSaysNothingOfNoData(t *testing.T) {
	if _, err := shipped(t).Fake("nope"); err == nil || strings.Contains(err.Error(), "no data is loaded") {
		t.Errorf("Fake(nope) = %v, want no entry without the no-data note", err)
	}
}

func TestAGeneratorWithNoDataRendersWhatReadsNone(t *testing.T) {
	for name, opts := range map[string][]Option{
		"no source":    {WithoutShippedData()},
		"empty folder": {WithoutShippedData(), WithDataPath(t.TempDir())},
	} {
		f, err := New(opts...)
		if err != nil {
			t.Fatalf("%s: New = %v, want a generator holding no data", name, err)
		}
		if v, err := f.FakeTemplate("{digits(3)}"); err != nil || len(v) != 3 {
			t.Errorf("%s: FakeTemplate = %q, %v, want three digits", name, v, err)
		}
		if _, err := f.Fake("x"); err == nil || !strings.Contains(err.Error(), "no data is loaded") {
			t.Errorf("%s: Fake(x) = %v, want a reference to nothing refused, saying no data is loaded", name, err)
		}
		if _, err := f.NewTemplate("{/x}"); err == nil || !strings.Contains(err.Error(), "no data is loaded") {
			t.Errorf("%s: NewTemplate({/x}) = %v, want a reference to nothing refused, saying no data is loaded", name, err)
		}
		if _, err := f.FakeRecord("x"); err == nil || !strings.Contains(err.Error(), "no data is loaded") {
			t.Errorf("%s: FakeRecord(x) = %v, want a reference to nothing refused, saying no data is loaded", name, err)
		}
	}
}

func TestWithoutShippedDataListsOnlyOwn(t *testing.T) {
	dir := writeData(t, map[string]string{"greeting": `"hej"`})
	f := newGenerator(t, dir, WithSeed(1))
	if got := f.List(); !reflect.DeepEqual(got, []string{"greeting"}) {
		t.Errorf("List() = %v, want only the loaded dir", got)
	}
}

func TestWithDataFS(t *testing.T) {
	fsys := fstest.MapFS{
		"greeting.json":  {Data: []byte(`"hej"`)},
		"nested/x.json":  {Data: []byte(`{"format":"{y}","y":"z"}`)},
		".hidden.json":   {Data: []byte(`"ignored"`)},
		"broken/no.json": {Data: []byte(`"x"`)},
	}
	f, err := New(WithoutShippedData(), WithDataFS(fsys), WithSeed(1))
	if err != nil {
		t.Fatalf("New(WithDataFS) = %v", err)
	}
	if got := fake(t, f, "greeting"); got != "hej" {
		t.Errorf("greeting = %q, want hej", got)
	}
	if got := fake(t, f, "nested.x.y"); got != "z" {
		t.Errorf("nested.x.y = %q, want z", got)
	}
	bad := fstest.MapFS{"broken.json": {Data: []byte(`{ not json`)}}
	if _, err := New(WithoutShippedData(), WithDataFS(bad)); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Errorf("New(bad fs) = %v, want an error naming the file", err)
	}
}

func TestFakeIsSafeForConcurrentUse(t *testing.T) {
	f, err := New(WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := f.Fake("sv_SE.person"); err != nil {
					t.Error(err)
					return
				}
				f.List()
				if _, err := f.FakeRecord("sv_SE.person"); err != nil {
					t.Error(err)
					return
				}
				tmpl, err := f.NewTemplate("{/sv_SE.person.last}")
				if err != nil {
					t.Error(err)
					return
				}
				tmpl.Fake()
				var u struct {
					Last string `fake:"sv_SE.person.last"`
				}
				if err := f.FakeStruct(&u); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
