package fejkdata

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// newGenerator creates a generator over a single data directory, failing on
// error. Most tests load one dir; newGeneratorN loads several (last-loaded wins
// on conflicts).
func newGenerator(t *testing.T, dir string, opts ...Option) *Generator {
	t.Helper()
	return newGeneratorN(t, []string{dir}, opts...)
}

func newGeneratorN(t *testing.T, dirs []string, opts ...Option) *Generator {
	t.Helper()
	all := []Option{WithoutShippedData()}
	for _, dir := range dirs {
		all = append(all, WithDataPath(dir))
	}
	f, err := New(append(all, opts...)...)
	if err != nil {
		t.Fatalf("New(%q): %v", dirs, err)
	}
	return f
}

// fake generates a value, failing the test on error.
func fake(t *testing.T, f *Generator, path string) string {
	t.Helper()
	s, err := f.Fake(path)
	if err != nil {
		t.Fatalf("Fake(%q): %v", path, err)
	}
	return s
}

func TestNewMissingDirectory(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath("data/de_DE"))
	if err == nil || !strings.Contains(err.Error(), "de_DE") || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("New(missing) error = %v, want the real error, naming the path", err)
	}
}

func TestNewRejectsAnEmptyDataPath(t *testing.T) {
	_, err := New(WithoutShippedData(), WithDataPath(""))
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("New(WithDataPath(\"\")) = %v, want the empty path named", err)
	}
}

func TestNewReportsAnEntropyFailure(t *testing.T) {
	saved := randomBytes
	randomBytes = func([]byte) (int, error) { return 0, errors.New("no entropy") }
	defer func() { randomBytes = saved }()
	_, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{"w.json": {Data: []byte(`"x"`)}}))
	if err == nil || !strings.Contains(err.Error(), "no entropy") {
		t.Fatalf("New() without entropy = %v, want the failure reported", err)
	}
}

func TestWithSeedIsDeterministic(t *testing.T) {
	a, b := newGenerator(t, "data", WithSeed(42)), newGenerator(t, "data", WithSeed(42))
	for i := 0; i < 50; i++ {
		if x, y := fake(t, a, "sv_SE.person"), fake(t, b, "sv_SE.person"); x != y {
			t.Fatalf("same seed diverged at %d: %q != %q", i, x, y)
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	a, b := newGenerator(t, "data", WithSeed(1)), newGenerator(t, "data", WithSeed(2))
	for i := 0; i < 50; i++ {
		if fake(t, a, "en_US.person") != fake(t, b, "en_US.person") {
			return
		}
	}
	t.Fatal("seeds 1 and 2 produced identical sequences")
}

// fakeTemplate renders an inline template against a loaded generator, so its
// references resolve.
func fakeTemplate(t *testing.T, f *Generator, s string) string {
	t.Helper()
	got, err := f.FakeTemplate(s)
	if err != nil {
		t.Fatalf("FakeTemplate(%s) = %v", s, err)
	}
	return got
}
