package fejkdata

import (
	"strings"
	"testing"
)

// newFejkdata creates a faker over a single data directory, failing on error. Most
// tests load one dir; newFejkdataN loads several (last-loaded wins on conflicts).
func newFejkdata(t *testing.T, dir string, opts ...Option) *Fejkdata {
	t.Helper()
	return newFejkdataN(t, []string{dir}, opts...)
}

func newFejkdataN(t *testing.T, dirs []string, opts ...Option) *Fejkdata {
	t.Helper()
	f, err := New(dirs, opts...)
	if err != nil {
		t.Fatalf("New(%q): %v", dirs, err)
	}
	return f
}

// fake generates a value, failing the test on error.
func fake(t *testing.T, f *Fejkdata, path string) string {
	t.Helper()
	s, err := f.Fake(path)
	if err != nil {
		t.Fatalf("Fake(%q): %v", path, err)
	}
	return s
}

func TestNewMissingDirectory(t *testing.T) {
	_, err := New([]string{"data/de_DE"})
	if err == nil || !strings.Contains(err.Error(), "de_DE") {
		t.Fatalf("New(missing) error = %v, want it to name the path", err)
	}
}

func TestWithSeedIsDeterministic(t *testing.T) {
	a, b := newFejkdata(t, "data/sv_SE", WithSeed(42)), newFejkdata(t, "data/sv_SE", WithSeed(42))
	for i := 0; i < 50; i++ {
		if x, y := fake(t, a, "person"), fake(t, b, "person"); x != y {
			t.Fatalf("same seed diverged at %d: %q != %q", i, x, y)
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	a, b := newFejkdata(t, "data/en_US", WithSeed(1)), newFejkdata(t, "data/en_US", WithSeed(2))
	for i := 0; i < 50; i++ {
		if fake(t, a, "person") != fake(t, b, "person") {
			return
		}
	}
	t.Fatal("seeds 1 and 2 produced identical sequences")
}
