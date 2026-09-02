package fejkdata

import (
	"errors"
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
	dir := writeData(t, map[string]string{"greeting": `"Hej {..sv_SE.person}!"`})
	f, err := New(WithDataPath(dir), WithSeed(1))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	if got := fake(t, f, "greeting"); !strings.HasPrefix(got, "Hej ") || !strings.HasSuffix(got, "!") {
		t.Errorf("greeting = %q", got)
	}
}

func TestWithoutShippedDataNeedsASource(t *testing.T) {
	_, err := New(WithoutShippedData())
	if err == nil || !strings.Contains(err.Error(), "WithDataPath") || !errors.Is(err, ErrNoData) {
		t.Fatalf("New(WithoutShippedData()) = %v, want ErrNoData naming WithDataPath", err)
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
			}
		}()
	}
	wg.Wait()
}
