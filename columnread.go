package fejkdata

import (
	"slices"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/proven"
)

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name. It is nil where the format reads anything else, where t's repeat is not 1, or where t
// declares a string, which reads the column as text.
func columnReadOf(t *tree.Template, l *tree.LoadFinds) *tree.ColumnRead {
	a, one := readsOneArm(t, l)
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
		return tree.NewColumnRead(tree.ColumnReadSpec{Arm: a, Category: category, Field: tail[len(tail)-1], Column: columns[0]})
	}
	return tree.NewColumnRead(tree.ColumnReadSpec{Arm: a, Category: category, Field: tail[len(tail)-1], Column: tree.NewChoice(columns, nil, nil), Spread: true})
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

// readsOneArm is the one arm t prints, where t's format prints one plain read and t reads it as its
// datatype.
func readsOneArm(t *tree.Template, l *tree.LoadFinds) (tree.Arm, bool) {
	if t.Repeat() != 1 || (t.DataType() != nil && *t.DataType() == DataTypeString) {
		return tree.Arm{}, false
	}
	o, one := printedOp(l.Of(t).Compiled().Ops())
	if !one {
		return tree.Arm{}, false
	}
	tok, arms := o.Token(), o.Arms()
	if tok.Kind != grammar.PathRead || len(arms) != 1 || tok.Where != nil || tok.Distinct != nil || arms[0].ReadsJoined() {
		return tree.Arm{}, false
	}
	return arms[0], true
}

// columnSource is the node a's read starts from (nil where a reads no node), the category holding
// that node (empty where it is t's own), and the path below the node.
func columnSource(t *tree.Template, a tree.Arm, l *tree.LoadFinds) (start tree.Node, category string, tail []string) {
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

// checkReadColumns runs check on the column of another category that each of items reads, and on
// the choice of columns that an item of the column name of label reads through a choice.
func checkReadColumns(label, name string, items []*tree.Template, check func(label, name string, column tree.Node) error) error {
	for _, it := range items {
		r := it.ReadsColumn()
		switch {
		case r == nil:
		case r.Category() != "":
			if err := check(r.Category(), r.Field(), r.Column()); err != nil {
				return err
			}
		case r.Spread():
			if err := check(label, name, r.Column()); err != nil {
				return err
			}
		}
	}
	return nil
}

// itemDatatype is the datatype a column item declares, else that of the column it reads.
func itemDatatype(t *tree.Template, l *tree.LoadFinds) DataType {
	r := l.Of(t).ReadsColumn()
	switch {
	case t.DataType() != nil:
		return *t.DataType()
	case r != nil:
		return columnDatatype(r.Column(), l)
	}
	return DataTypeString
}

// proveColumnItem proves what a column item renders or, where the item is a column read, the column it reads.
func (p *valueProof) proveColumnItem(t *tree.Template) proven.Facts {
	r := t.ReadsColumn()
	if r == nil {
		return p.prove(t)
	}
	return p.proveColumn(r.Column())
}

// readColumnItem reads the column t reads, where t is a column read. Rendering t's format
// instead would print a null the column drew as "". t's own name scope opens first, since the read
// may be of a name t binds, such as {/c as n}{n.x}.
func readColumnItem(s *drawstate.State, t *tree.Template, env renderEnv) (r readValue, reads bool) {
	if t.ReadsColumn() == nil {
		return readValue{}, false
	}
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	return readField(s, t, env, t.ReadsColumn().Arm()), true
}
