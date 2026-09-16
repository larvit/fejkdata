package fejkdata

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

const shapePin = "testdata/shipped_shape.txt"

// TestShippedShapeIsPinned pins what a version promises about the shipped data (see
// the README's Versioning): every path, each category's format, and each record
// column's datatype and nullability. REPIN=1 rewrites the pin.
func TestShippedShapeIsPinned(t *testing.T) {
	f, err := New(WithSeed(1))
	if err != nil {
		t.Fatal(err)
	}
	got := shippedShape(f)
	if os.Getenv("REPIN") == "1" {
		if err := os.WriteFile(shapePin, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(shapePin)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("the shipped shape differs from %s: a removed, renamed or retyped line is a breaking change; add the CHANGELOG.md entry, then repin with REPIN=1", shapePin)
	}
}

// shippedShape lists every path Fake accepts, one per line. A category-level template
// carries its format, and each of its columns its datatype and whether it may be null.
func shippedShape(f *Generator) string {
	facts := map[string]string{}
	var walk func(prefix string, n node)
	walk = func(prefix string, n node) {
		switch n := n.(type) {
		case *folder:
			for _, name := range sortedNames(n.children) {
				walk(join(prefix, name), n.children[name])
			}
		case *template:
			facts[prefix] = "\tformat " + strconv.Quote(n.format)
			if _, columns, err := recordOf(n); err == nil {
				for _, c := range columns {
					fact := "\t" + c.DataType.String()
					if _, nullable := columnItems(n.fields[c.Name]); nullable {
						fact += " null"
					}
					facts[join(prefix, c.Name)] = fact
				}
			}
		}
	}
	for _, name := range sortedNames(f.categories) {
		walk(name, f.categories[name])
	}
	var b strings.Builder
	for _, p := range f.List() {
		b.WriteString(p + facts[p] + "\n")
	}
	return b.String()
}
