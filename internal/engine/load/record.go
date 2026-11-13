package load

import (
	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// SettleRecords fixes the columns of every record in nodes.
func SettleRecords(nodes func(fn func(label string, n tree.Node) error) error, l *tree.LoadFinds) {
	_ = nodes(func(_ string, n tree.Node) error {
		if t, isTemplate := n.(*tree.Template); isTemplate {
			l.SetColumns(t, recordColumns(t, l))
		}
		return nil
	})
}

// recordColumns is t's columns in name order where t is a record with fields, else nil.
func recordColumns(t *tree.Template, l *tree.LoadFinds) []tree.RecordColumn {
	if !t.IsRecord() || len(t.Fields()) == 0 {
		return nil
	}
	whole := map[tree.Node]*tree.NameBinding{}
	for _, tok := range t.Tokens() {
		if tok.Kind != grammar.NameBind {
			continue
		}
		b := l.Of(t).NameScope().Bindings()[tok.Bound]
		if target := l.Binding(b).Target(); b.BindsField() && len(target.Tail()) == 0 && whole[target.Start()] == nil {
			whole[target.Start()] = b
		}
	}
	var columns []tree.RecordColumn
	for _, name := range tree.SortedNames(t.Fields()) {
		if !t.Lets()[name] {
			field := t.Fields()[name]
			columns = append(columns, tree.NewRecordColumn(name, ColumnDatatype(field, l), field, whole[field]))
		}
	}
	return columns
}

// ColumnDatatype is the datatype a column holds, which its first item decides: fence.CheckColumns
// refuses a column whose items disagree. A table's column, like one only ever null, has no item and
// is a string, and so is a column that reads itself back, which fence.CheckNoCycles refuses.
func ColumnDatatype(n tree.Node, l *tree.LoadFinds) datatype.DataType {
	return columnDatatype(n, l, map[tree.Node]bool{})
}

func columnDatatype(n tree.Node, l *tree.LoadFinds, seen map[tree.Node]bool) datatype.DataType {
	items, _ := ColumnItems(n)
	if len(items) == 0 || seen[n] {
		return datatype.String
	}
	seen[n] = true
	return itemDatatype(items[0], l, seen)
}

// ColumnItems is a column's template items, its choices unwrapped, and whether one is null.
func ColumnItems(n tree.Node) (items []*tree.Template, nullable bool) {
	var collect func(tree.Node)
	collect = func(n tree.Node) {
		switch n := n.(type) {
		case *tree.Choice:
			for _, it := range n.Items() {
				collect(it)
			}
		case *tree.Template:
			items = append(items, n)
		case *tree.NullItem:
			nullable = true
		case *tree.TableColumn:
		default:
			panic(invariant.Broken("ColumnItems has no case for node %T", n))
		}
	}
	collect(n)
	return items, nullable
}

// ItemDatatype is the datatype a column item declares, else that of the column it reads.
func ItemDatatype(t *tree.Template, l *tree.LoadFinds) datatype.DataType {
	return itemDatatype(t, l, map[tree.Node]bool{})
}

func itemDatatype(t *tree.Template, l *tree.LoadFinds, seen map[tree.Node]bool) datatype.DataType {
	r := l.Of(t).ReadsColumn()
	switch {
	case t.DataType() != nil:
		return *t.DataType()
	case r != nil:
		return columnDatatype(r.Column(), l, seen)
	}
	return datatype.String
}
