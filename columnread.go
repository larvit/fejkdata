package fejkdata

import (
	"slices"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/proven"
)

// columnRead is a record's column that a format reads, where the format only reads that one
// reference or name. The format then counts as the column: it takes the column's datatype and
// null. category and field name the column; category is "" for a column of the reading
// template's own record, which checkColumns reaches on its own. spread is set where the read goes
// through the items of a choice: column is then a choice of the column each item holds, which
// checkColumns checks for agreement, since nothing else reaches those items.
type columnRead struct {
	a               arm
	category, field string
	column          node
	spread          bool
}

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name. It is nil where the format reads anything else, where t's repeat is not 1, or where t
// declares a string, which reads the column as text.
func columnReadOf(t *template, l *loadFinds) *columnRead {
	a, one := readsOneArm(t, l)
	if !one {
		return nil
	}
	start, category, tail := columnSource(t, a, l)
	target, isTemplate := start.(*template)
	if !isTemplate || !target.isRecord || len(tail) == 0 {
		return nil
	}
	columns := columnsAt(target, tail)
	switch len(columns) {
	case 0:
		return nil
	case 1:
		return &columnRead{a: a, category: category, field: tail[len(tail)-1], column: columns[0]}
	}
	return &columnRead{a: a, category: category, field: tail[len(tail)-1], column: &choice{items: columns}, spread: true}
}

// columnsAt is the columns n holds along tail: the field of a template, and the column of each item
// of a choice. It is nil where a step has no field to follow, where a template on the way repeats,
// or where an item of a choice is no template. It runs before checkNameReads, so a read that fence
// will refuse comes back nil.
func columnsAt(n node, tail []string) (columns []node) {
	if len(tail) == 0 {
		return []node{n}
	}
	switch n := n.(type) {
	case *template:
		field, ok := n.fields[tail[0]]
		if !ok || n.repeat != 1 {
			return nil
		}
		return columnsAt(field, tail[1:])
	case *choice:
		for _, it := range n.items {
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
func readsOneArm(t *template, l *loadFinds) (arm, bool) {
	if t.repeat != 1 || (t.datatype != nil && *t.datatype == DataTypeString) {
		return arm{}, false
	}
	o, one := printedOp(l.of(t).compiled.ops)
	if !one || o.Kind != grammar.PathRead || len(o.arms) != 1 || o.Where != nil || o.Distinct != nil || readsJoined(o.arms[0]) {
		return arm{}, false
	}
	return o.arms[0], true
}

// readsJoined is whether a reads the members of a distinct draw joined, which is text whatever
// they are.
func readsJoined(a arm) bool {
	return a.kind == namedRead && a.named.distinct != nil && a.member == 0
}

// columnSource is the node a's read starts from (nil where a reads no node), the category holding
// that node (empty where it is t's own), and the path below the node.
func columnSource(t *template, a arm, l *loadFinds) (start node, category string, tail []string) {
	if b := a.named; a.kind == namedRead {
		target := l.binding(b).target
		if b.bindsField() {
			// {f as n}{n}: column f of the record binding n, in t's own category.
			return b.binder, "", slices.Concat([]string{splitRead(b.ref, nil).head}, target.tail, a.tail)
		}
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		return target.start, categoryOf(l.of(b.binder).refs.bySpelling[b.ref].head), slices.Concat(target.tail, a.tail)
	}
	if grammar.IsRef(a.head) {
		// {/c.x}: column x of category c.
		return t.startOf(a.head, l.of(t).refs), categoryOf(a.head), a.tail
	}
	return nil, "", nil
}

// checkReadColumns runs check on the column of another category that each of items reads, and on
// the choice of columns that an item of the column name of label reads through a choice.
func checkReadColumns(label, name string, items []*template, check func(label, name string, column node) error) error {
	for _, it := range items {
		r := it.readsColumn
		switch {
		case r == nil:
		case r.category != "":
			if err := check(r.category, r.field, r.column); err != nil {
				return err
			}
		case r.spread:
			if err := check(label, name, r.column); err != nil {
				return err
			}
		}
	}
	return nil
}

// itemDatatype is the datatype a column item declares, else that of the column it reads.
func itemDatatype(t *template, l *loadFinds) DataType {
	r := l.of(t).readsColumn
	switch {
	case t.datatype != nil:
		return *t.datatype
	case r != nil:
		return columnDatatype(r.column, l)
	}
	return DataTypeString
}

// proveColumnItem proves what a column item renders or, where the item is a column read, the column it reads.
func (p *valueProof) proveColumnItem(t *template) proven.Facts {
	r := t.readsColumn
	if r == nil {
		return p.prove(t)
	}
	return p.proveColumn(r.column)
}

// readColumnItem reads the column t reads, where t is a column read. Rendering t's format
// instead would print a null the column drew as "". t's own name scope opens first, since the read
// may be of a name t binds, such as {/c as n}{n.x}.
func readColumnItem(s *drawstate.State, t *template, env renderEnv) (r readValue, reads bool) {
	if t.readsColumn == nil {
		return readValue{}, false
	}
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	return readField(s, t, env, t.readsColumn.a), true
}
