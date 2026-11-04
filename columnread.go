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
// template's own record, which checkColumns reaches on its own.
type columnRead struct {
	a               arm
	category, field string
	column          node
}

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name. It is nil where the format reads anything else, where t's repeat is not 1, or where t
// declares a string, which reads the column as text.
func columnReadOf(t *template, targets map[*nameBinding]nameTarget) *columnRead {
	ops := t.compiled.ops
	if t.datatype != nil && *t.datatype == DataTypeString || t.repeat != 1 || len(ops) != 1 || ops[0].Kind != grammar.PathRead || len(ops[0].arms) != 1 {
		return nil
	}
	a := ops[0].arms[0]
	var start node
	var category string
	var tail []string
	switch b := a.named; {
	case a.kind == namedRead && b.bindsField():
		// {f as n}{n}: column f of the record binding n, in t's own category.
		start, tail = b.binder, slices.Concat([]string{splitArm(b.ref, nil).head}, targets[b].tail, a.tail)
	case a.kind == namedRead:
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		start, category = targets[b].start, categoryOf(b.binder.refs.byName[b.ref].head)
		tail = slices.Concat(targets[b].tail, a.tail)
	case grammar.IsRef(a.head):
		// {/c.x}: column x of category c.
		start, category, tail = t.startOf(a.head), categoryOf(a.head), a.tail
	default:
		return nil
	}
	if target, isTemplate := start.(*template); isTemplate && target.isRecord && len(tail) == 1 {
		return &columnRead{a: a, category: category, field: tail[0], column: target.fields[tail[0]]}
	}
	return nil
}

// checkReadColumns runs check on the column of another category that each of items reads.
func checkReadColumns(items []*template, check func(label, name string, column node) error) error {
	for _, it := range items {
		if r := it.readsColumn; r != nil && r.category != "" {
			if err := check(r.category, r.field, r.column); err != nil {
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
// instead would print a null the column drew as "". t's own name scope opens first, as render
// opens it, since the read may be of a name t binds, such as {/c as n}{n.x}.
func readColumnItem(s *drawstate.State, t *template, env renderEnv) (r readValue, reads bool) {
	if t.readsColumn == nil {
		return readValue{}, false
	}
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	return readField(s, t, env, t.readsColumn.a), true
}
