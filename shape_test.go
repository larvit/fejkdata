package fejkdata

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

const shapePin = "testdata/shipped_shape.txt"

// REPIN=1 rewrites the pin.
func TestShippedShapeIsPinned(t *testing.T) {
	got := shippedShape(newShippedWhole(t, WithSeed(1), onTestDay))
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
	var walk func(prefix string, n tree.Node)
	walk = func(prefix string, n tree.Node) {
		switch n := n.(type) {
		case *tree.Folder:
			for _, name := range tree.SortedNames(n.Children()) {
				walk(prefix+"."+name, n.Children()[name])
			}
		case *tree.Choice:
			facts[prefix] = readsFact(n)
		case *tree.Table:
			facts[prefix] = "\tformat " + strconv.Quote(n.FormatTemplate().Format()) + tableFacts(n) + readsFact(n)
			for _, name := range n.Rows().Header() {
				facts[prefix+"."+name] = "\tstring"
			}
		case *tree.Template:
			facts[prefix] = "\tformat " + strconv.Quote(n.Format()) + readsFact(n)
			if record, err := recordOf(n); err == nil {
				for _, c := range record.Columns() {
					fact := "\t" + c.DataType().String()
					if _, nullable := fence.ColumnItems(c.Field()); nullable {
						fact += " null"
					}
					facts[prefix+"."+c.Name()] = fact
				}
			}
		}
	}
	for _, name := range tree.SortedNames(f.root.Children()) {
		walk(name, f.root.Children()[name])
	}
	var b strings.Builder
	for _, p := range f.List() {
		b.WriteString(p + facts[p] + "\n")
	}
	return b.String()
}

// tableFacts names the columns a table's options read.
func tableFacts(t *tree.Table) string {
	var b strings.Builder
	c := t.Rows().Options()
	for _, o := range []struct{ name, column string }{{"key", c.Key}, {"name", c.Name}, {"weight", c.Weight}, {"parent", c.Parent}} {
		if o.column != "" {
			b.WriteString("\t" + o.name + " " + o.column)
		}
	}
	return b.String()
}

// readsFact names the categories any template under n references, sorted.
func readsFact(n tree.Node) string {
	set := map[string]bool{}
	var collect func(tree.Node)
	collect = func(n tree.Node) {
		switch n := n.(type) {
		case *tree.Choice:
			for _, it := range n.Items() {
				collect(it)
			}
		case *tree.Table:
			collect(n.FormatTemplate())
			for _, cell := range n.CellTemplates() {
				collect(cell)
			}
		case *tree.Template:
			for _, b := range n.Refs().BySpelling() {
				set[strings.TrimPrefix(b.Head(), "/")] = true
			}
			for _, field := range n.Fields() {
				collect(field)
			}
		}
	}
	collect(n)
	if len(set) == 0 {
		return ""
	}
	return "\treads " + strings.Join(tree.SortedNames(set), " ")
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
