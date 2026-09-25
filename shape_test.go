package fejkdata

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const shapePin = "testdata/shipped_shape.txt"

// REPIN=1 rewrites the pin.
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

func shippedShape(f *Generator) string {
	facts := map[string]string{}
	var walk func(prefix string, n node)
	walk = func(prefix string, n node) {
		switch n := n.(type) {
		case *folder:
			for _, name := range sortedNames(n.children) {
				walk(join(prefix, name), n.children[name])
			}
		case *choice:
			facts[prefix] = readsFact(n)
		case *table:
			facts[prefix] = "\tformat " + strconv.Quote(n.formatTemplate.format) + tableFacts(n) + readsFact(n)
			for _, name := range n.header {
				facts[join(prefix, name)] = "\tstring"
			}
		case *template:
			facts[prefix] = "\tformat " + strconv.Quote(n.format) + readsFact(n)
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

// tableFacts names the columns a table's options read.
func tableFacts(t *table) string {
	var b strings.Builder
	for _, o := range []struct {
		name string
		col  int
	}{{"key", t.keyIndex}, {"name", t.nameIndex}, {"weight", t.weightIndex}, {"parent", t.parentIndex}} {
		if o.col >= 0 {
			b.WriteString("\t" + o.name + " " + t.header[o.col])
		}
	}
	return b.String()
}

// readsFact names the categories any template under n references, sorted.
func readsFact(n node) string {
	set := map[string]bool{}
	var collect func(node)
	collect = func(n node) {
		switch n := n.(type) {
		case *choice:
			for _, it := range n.items {
				collect(it)
			}
		case *table:
			collect(n.formatTemplate)
			for _, cell := range n.cellTemplates {
				collect(cell)
			}
		case *template:
			for _, b := range n.refs {
				set[strings.TrimPrefix(b.key, "/")] = true
			}
			for _, field := range n.fields {
				collect(field)
			}
		}
	}
	collect(n)
	if len(set) == 0 {
		return ""
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return "\treads " + strings.Join(keys, " ")
}

func TestShippedShapeNamesReads(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{
		"a":     `{"format":"{x}","x":["{/b}",{"format":"{/c.v}","weight":2}]}`,
		"b":     `"y"`,
		"c":     `{"format":"{v} {n}","n":[null,{"format":"{int(1,9)}","datatype":"integer"}],"v":["z","w"]}`,
		"d/pos": `["{.q}","{/b}"]`,
		"d/q":   `"r"`,
		"e":     `"{/a}"`,
	}))
	want := "a\tformat \"{x}\"\treads b c\na.x\tstring\nb\tformat \"y\"\nc\tformat \"{v} {n}\"\nc.n\tinteger null\nc.v\tstring\nd.pos\treads b d.q\nd.q\tformat \"r\"\ne\tformat \"{/a}\"\treads a\n"
	if got := shippedShape(f); got != want {
		t.Fatalf("shippedShape =\n%s\nwant\n%s", got, want)
	}
}

func TestShippedShapeNamesTables(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"w.json": `"x"`,
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code","name":"name","weight":"w"}`,
		"r.tsv":  "code\tname\tw\n1\ta\t2\n2\tb\t3\n",
		"m.json": `{"format":"{name} {/w}","rows":"m.tsv","key":"code","parent":"r"}`,
		"m.tsv":  "code\tname\tr\n10\tc\t1\n20\td\t2\n",
	}))
	want := "m\tformat \"{name} {/w}\"\tkey code\tparent r\treads w\nm.code\tstring\nm.name\tstring\nm.r\tstring\nr\tformat \"{name}\"\tkey code\tname name\tweight w\nr.code\tstring\nr.m\nr.m.code\nr.m.name\nr.m.r\nr.name\tstring\nr.w\tstring\nw\tformat \"x\"\n"
	if got := shippedShape(f); got != want {
		t.Fatalf("shippedShape =\n%s\nwant\n%s", got, want)
	}
}
