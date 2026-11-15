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
	whole := wholeFieldNames(t, l)
	var columns []tree.RecordColumn
	for _, name := range tree.SortedNames(t.Fields()) {
		if !t.Lets()[name] {
			field := t.Fields()[name]
			columns = append(columns, tree.NewRecordColumn(name, ColumnDatatype(field, l.ReadsColumn), field, whole[field]))
		}
	}
	return columns
}

// wholeFieldNames maps each field of t that a name of t binds whole to the first such name.
func wholeFieldNames(t *tree.Template, l *tree.LoadFinds) map[tree.Node]*tree.NameBinding {
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
	return whole
}

// ReadsColumn is the column a template reads, if it only reads one: LoadFinds.ReadsColumn while
// a load runs, Template.ReadsColumn once it has finished.
type ReadsColumn func(*tree.Template) *tree.ColumnRead

// ColumnDatatype is the datatype a column holds, which its first item decides: fence.Check
// refuses a column whose items disagree. A column with no item, such as a table's column or one only
// ever null, is a string. So is a column that reads itself back, which fence.Check later
// refuses.
func ColumnDatatype(n tree.Node, readsColumn ReadsColumn) datatype.DataType {
	return columnDatatype(n, readsColumn, map[tree.Node]bool{})
}

func columnDatatype(n tree.Node, readsColumn ReadsColumn, seen map[tree.Node]bool) datatype.DataType {
	items, _ := ColumnItems(n)
	if len(items) == 0 || seen[n] {
		return datatype.String
	}
	seen[n] = true
	return itemDatatype(items[0], readsColumn, seen)
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
func ItemDatatype(t *tree.Template, readsColumn ReadsColumn) datatype.DataType {
	return itemDatatype(t, readsColumn, map[tree.Node]bool{})
}

func itemDatatype(t *tree.Template, readsColumn ReadsColumn, seen map[tree.Node]bool) datatype.DataType {
	r := readsColumn(t)
	switch {
	case t.DataType() != nil:
		return *t.DataType()
	case r != nil:
		return columnDatatype(r.Column(), readsColumn, seen)
	}
	return datatype.String
}
