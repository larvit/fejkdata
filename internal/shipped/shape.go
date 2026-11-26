package shipped

import (
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// ShapeFile is the pin of the shipped shape.
const ShapeFile = "testdata/shipped_shape.txt"

// Shape is the pin of the tree's shape: a line for every listed path, with what a caller relies
// on there, such as a category's format and the categories it reads, a table's options, and a
// column's datatype and whether it may be null.
func (d *Data) Shape() File {
	root := &d.root
	facts := shapeFacts{}
	for _, name := range tree.SortedNames(root.Children()) {
		facts.walk(name, root.Children()[name])
	}
	var b strings.Builder
	for _, p := range engine.List(root) {
		b.WriteString(p + facts[p] + "\n")
	}
	return File{ShapeFile, []byte(b.String())}
}

// shapeFacts is what a shape line says after its path, by path.
type shapeFacts map[string]string

func (facts shapeFacts) walk(prefix string, n tree.Node) {
	switch n := n.(type) {
	case *tree.Folder:
		for _, name := range tree.SortedNames(n.Children()) {
			facts.walk(prefix+"."+name, n.Children()[name])
		}
	case *tree.Choice:
		facts[prefix] = readsFact(n)
	case *tree.Table:
		facts[prefix] = "\tformat " + strconv.Quote(n.FormatTemplate().Format()) + tableFacts(n) + readsFact(n) + "\tcolumns " + strings.Join(n.Rows().Header(), " ")
		for _, name := range n.Rows().Header() {
			facts[prefix+"."+name] = "\tstring"
		}
	case *tree.Template:
		facts.template(prefix, n)
	}
}

func (facts shapeFacts) template(prefix string, t *tree.Template) {
	facts[prefix] = "\tformat " + strconv.Quote(t.Format()) + readsFact(t)
	var order []string
	for _, c := range t.Columns() {
		order = append(order, c.Name())
		fact := "\t" + c.DataType().String()
		if _, nullable := load.ColumnItems(c.Field()); nullable {
			fact += " null"
		}
		facts[prefix+"."+c.Name()] = fact
	}
	if len(order) > 0 {
		facts[prefix] += "\tcolumns " + strings.Join(order, " ")
	}
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
	collectReads(set, n)
	if len(set) == 0 {
		return ""
	}
	return "\treads " + strings.Join(tree.SortedNames(set), " ")
}

func collectReads(set map[string]bool, n tree.Node) {
	switch n := n.(type) {
	case *tree.Choice:
		for _, it := range n.Items() {
			collectReads(set, it)
		}
	case *tree.Table:
		collectReads(set, n.FormatTemplate())
		for _, cell := range n.CellTemplates() {
			collectReads(set, cell)
		}
	case *tree.Template:
		collectTemplateReads(set, n)
	}
}

func collectTemplateReads(set map[string]bool, t *tree.Template) {
	for _, b := range t.Refs().BySpelling() {
		set[strings.TrimPrefix(b.Head(), "/")] = true
	}
	for _, field := range t.Fields() {
		collectReads(set, field)
	}
}
