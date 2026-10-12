package fejkdata

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/data"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

func withShipped() Option { return WithDataFS(data.Modules()...) }

// newGenerator creates a generator over a single data directory, failing on
// error. Most tests load one dir; newGeneratorN loads several (last-loaded wins
// on conflicts).
func newGenerator(t *testing.T, dir string, opts ...Option) *Generator {
	t.Helper()
	return newGeneratorN(t, []string{dir}, opts...)
}

func newGeneratorN(t *testing.T, dirs []string, opts ...Option) *Generator {
	t.Helper()
	var all []Option
	for _, dir := range dirs {
		all = append(all, WithDataPath(dir))
	}
	f, err := New(append(append(all, WithSeed(1)), opts...)...)
	if err != nil {
		t.Fatalf("New(%q): %v", dirs, err)
	}
	return f
}

// newBareModule loads the shipped module in data/<module> without its manifest, so its
// categories sit at the root, failing on error.
func newBareModule(t *testing.T, module string, opts ...Option) *Generator {
	t.Helper()
	f, err := New(append([]Option{WithDataFS(withoutManifest{os.DirFS(filepath.Join("data", module))}), WithSeed(1)}, opts...)...)
	if err != nil {
		t.Fatalf("New(%s without its manifest): %v", module, err)
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
	_, err := New(WithSeed(1), WithDataPath("data/de_DE"))
	if err == nil || !strings.Contains(err.Error(), "de_DE") || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("New(missing) error = %v, want the real error, naming the path", err)
	}
}

func TestNewRejectsAnEmptyDataPath(t *testing.T) {
	_, err := New(WithSeed(1), WithDataPath(""))
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("New(WithDataPath(\"\")) = %v, want the empty path named", err)
	}
}

func TestNewWithoutASeedFailsNamingWithSeed(t *testing.T) {
	for name, opts := range map[string][]Option{"no option": nil, "only data": {WithDataFS()}} {
		if _, err := New(opts...); err == nil || !strings.Contains(err.Error(), "WithSeed(") {
			t.Errorf("New(%s) = %v, want it refused, naming WithSeed", name, err)
		}
	}
	if _, err := New(WithDataFS(nil)); err == nil || !strings.Contains(err.Error(), "WithSeed(") || !strings.Contains(err.Error(), "module 1 is nil") {
		t.Errorf("New(WithDataFS(nil)) = %v, want both the missing seed and the nil module named", err)
	}
}

func TestWithSeedIsDeterministic(t *testing.T) {
	a, b := shipped(t, WithSeed(42)), shipped(t, WithSeed(42))
	for i := 0; i < 50; i++ {
		if x, y := fake(t, a, "sv_SE.person"), fake(t, b, "sv_SE.person"); x != y {
			t.Fatalf("same seed diverged at %d: %q != %q", i, x, y)
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	a, b := shipped(t, WithSeed(1)), shipped(t, WithSeed(2))
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

// engine builds a seeded generator with no loaded categories, for rendering tests.
func engine(seed uint64) *Generator {
	return &Generator{drawState: drawstate.New(seed)}
}

// parse unmarshals a JSON template fragment into its dynamic form.
func parse(t *testing.T, s string) jsonvalue.Value {
	t.Helper()
	v, err := jsonvalue.Decode([]byte(s))
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

// resolved compiles a JSON fragment and resolves it against no data.
func resolved(t *testing.T, s string) (node, error) {
	t.Helper()
	n, err := compile(parse(t, s))
	if err != nil {
		return nil, err
	}
	return n, resolveInlineTemplates(inlineNodes(n, "template"), nil)
}

// compiled parses, compiles and resolves a JSON fragment into a node.
func compiled(t *testing.T, s string) node {
	t.Helper()
	n, err := resolved(t, s)
	if err != nil {
		t.Fatalf("compile %q: %v", s, err)
	}
	return n
}

func mustRender(t *testing.T, f *Generator, s string) string {
	t.Helper()
	return renderOnce(f.drawState, compiled(t, s))
}
