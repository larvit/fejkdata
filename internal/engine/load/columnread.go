package load

import (
	"slices"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name. It is nil where the format reads anything else, where t's repeat is not 1, or where t
// declares a string, which reads the column as text.
func columnReadOf(t *tree.Template, l *tree.LoadFinds) *tree.ColumnRead {
	a, one := onlyRead(t, l)
	if !one {
		return nil
	}
	start, category, tail := columnSource(t, a, l)
	target, isTemplate := start.(*tree.Template)
	if !isTemplate || !target.IsRecord() || len(tail) == 0 {
		return nil
	}
	columns := columnsAt(target, tail)
	switch len(columns) {
	case 0:
		return nil
	case 1:
		return tree.NewColumnRead(tree.ColumnReadSpec{Read: a, Category: category, Field: tail[len(tail)-1], Column: columns[0]})
	}
	return tree.NewColumnRead(tree.ColumnReadSpec{Read: a, Category: category, Field: tail[len(tail)-1], Column: tree.NewChoice(columns, nil, nil), Spread: true})
}

// columnsAt is the columns n holds along tail: the field of a template, and the column of each item
// of a choice. It is nil where a step has no field to follow, where a template on the way repeats,
// or where an item of a choice is no template. It runs before checkNameReads, so a read that fence
// will refuse comes back nil.
func columnsAt(n tree.Node, tail []string) (columns []tree.Node) {
	if len(tail) == 0 {
		return []tree.Node{n}
	}
	switch n := n.(type) {
	case *tree.Template:
		field, ok := n.Fields()[tail[0]]
		if !ok || n.Repeat() != 1 {
			return nil
		}
		return columnsAt(field, tail[1:])
	case *tree.Choice:
		for _, it := range n.Items() {
			below := columnsAt(it, tail)
			if below == nil {
				return nil
			}
			columns = append(columns, below...)
		}
	}
	return columns
}

// onlyRead is the one read t prints, where t's format prints one plain read and t reads it as its
// datatype.
func onlyRead(t *tree.Template, l *tree.LoadFinds) (tree.Read, bool) {
	if t.Repeat() != 1 || (t.DataType() != nil && *t.DataType() == datatype.String) {
		return tree.Read{}, false
	}
	o, one := PrintedOp(l.Of(t).Compiled().Ops())
	if !one {
		return tree.Read{}, false
	}
	tok, reads := o.Token(), o.Alternatives()
	if tok.Kind != grammar.PathRead || len(reads) != 1 || tok.Where != nil || tok.Distinct != nil || reads[0].ReadsJoined() {
		return tree.Read{}, false
	}
	return reads[0], true
}

// columnSource is the node a starts from (nil where a reads no node), the category holding
// that node (empty where it is t's own), and the path below the node.
func columnSource(t *tree.Template, a tree.Read, l *tree.LoadFinds) (start tree.Node, category string, tail []string) {
	if b := a.Named(); a.Kind() == tree.NamedRead {
		target := l.Binding(b).Target()
		if b.BindsField() {
			// {f as n}{n}: column f of the record binding n, in t's own category.
			return b.Binder(), "", slices.Concat([]string{splitRead(b.Ref(), nil).head}, target.Tail(), a.Tail())
		}
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		ref := l.Of(b.Binder()).Refs().BySpelling()[b.Ref()]
		return target.Start(), categoryOf(ref.Head()), slices.Concat(target.Tail(), a.Tail())
	}
	if grammar.IsRef(a.Head()) {
		// {/c.x}: column x of category c.
		return t.StartOf(a.Head(), l.Of(t).Refs()), categoryOf(a.Head()), a.Tail()
	}
	return nil, "", nil
}
