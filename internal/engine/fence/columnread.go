package fejkdata

import (
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/proven"
)

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
