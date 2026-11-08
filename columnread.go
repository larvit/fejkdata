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
// declares a string, which reads the column as text. targets is an argument since the bindings
// take theirs only once every check has passed.
func columnReadOf(t *template, targets map[*nameBinding]nameTarget) *columnRead {
	if !readsOneArm(t) {
		return nil
	}
	a := t.compiled.ops[0].arms[0]
	start, category, tail := columnSource(t, a, targets)
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

func readsOneArm(t *template) bool {
	ops := t.compiled.ops
	return (t.datatype == nil || *t.datatype != DataTypeString) && t.repeat == 1 && len(ops) == 1 && ops[0].Kind == grammar.PathRead && len(ops[0].arms) == 1 && ops[0].Where == nil
}

// columnSource is the node a's read starts from (nil where a reads no node), the category holding
// that node (empty where it is t's own), and the path below the node.
func columnSource(t *template, a arm, targets map[*nameBinding]nameTarget) (start node, category string, tail []string) {
	switch b := a.named; {
	case a.kind == namedRead && b.bindsField():
		// {f as n}{n}: column f of the record binding n, in t's own category.
		return b.binder, "", slices.Concat([]string{splitArm(b.ref, nil).head}, targets[b].tail, a.tail)
	case a.kind == namedRead:
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		return targets[b].start, categoryOf(b.binder.refs.byName[b.ref].head), slices.Concat(targets[b].tail, a.tail)
	case grammar.IsRef(a.head):
		// {/c.x}: column x of category c.
		return t.startOf(a.head), categoryOf(a.head), a.tail
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
func itemDatatype(t *template) DataType {
	switch {
	case t.datatype != nil:
		return *t.datatype
	case t.readsColumn != nil:
		return columnDatatype(t.readsColumn.column)
	}
	return DataTypeString
}

// proveColumnItem proves what a column item renders or, where the item is a column read, the column it reads.
func (p *valueProof) proveColumnItem(t *template) proven.Facts {
	if t.readsColumn == nil {
		return p.prove(t)
	}
	return p.proveColumn(t.readsColumn.column)
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
