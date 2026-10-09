package fejkdata

import (
	"errors"
	"fmt"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// DataType is what a record column holds, which decides how a record writes its value. It
// prints as data spells it: string, integer, number or boolean.
type DataType = datatype.DataType

// The datatypes a column declares with "datatype"; a column without one is a string.
const (
	DataTypeString  DataType = datatype.String
	DataTypeInteger DataType = datatype.Integer
	DataTypeNumber  DataType = datatype.Number
	DataTypeBoolean DataType = datatype.Boolean
)

// datatypeOf reads a template's "datatype", nil where it carries none or sits outside a column.
func datatypeOf(m map[string]any, pos position) (*DataType, error) {
	v, ok := m["datatype"]
	if !ok {
		return nil, nil
	}
	name, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("datatype must be a string, not %s", jsonKind(v))
	}
	for d := DataTypeString; d < datatype.Count; d++ {
		if name != d.String() {
			continue
		}
		if !pos.column() {
			return nil, nil
		}
		return &d, nil
	}
	return nil, fmt.Errorf(`datatype takes "string", "integer", "number" or "boolean", got %q`, name)
}

// checkColumns rejects a record column whose items hold different datatypes, checking a column
// after the columns its items read, so a column read is named before its readers.
func checkColumns(s nodeSet) error {
	checked := map[node]bool{}
	var check func(label, name string, column node) error
	check = func(label, name string, column node) error {
		if checked[column] {
			return nil
		}
		checked[column] = true
		items, _ := columnItems(column)
		for _, it := range items {
			if r := it.readsColumn; r != nil && r.category != "" {
				if err := check(r.category, r.field, r.column); err != nil {
					return err
				}
			}
		}
		if len(items) == 0 {
			return nil
		}
		first := columnDatatype(column)
		for _, it := range items[1:] {
			if d := itemDatatype(it); d != first {
				return fmt.Errorf("%s: field %q: %w", label, name, disagreement(items[0], first, it, d))
			}
		}
		return nil
	}
	return s(func(label string, n node) error {
		t, ok := n.(*template)
		if !ok || !t.isRecord {
			return nil
		}
		for _, name := range sortedNames(t.fields) {
			if err := check(label, name, t.fields[name]); err != nil {
				return err
			}
		}
		return nil
	})
}

// columnDatatype is the datatype a column holds, which its first item decides: checkColumns refuses
// a column whose items disagree. A table's column, like one only ever null, has no item and is a string.
func columnDatatype(n node) DataType {
	items, _ := columnItems(n)
	if len(items) == 0 {
		return DataTypeString
	}
	return itemDatatype(items[0])
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

// columnItems is a column's template items, its choices unwrapped, and whether one is null.
func columnItems(n node) (items []*template, nullable bool) {
	var collect func(node)
	collect = func(n node) {
		switch n := n.(type) {
		case *choice:
			for _, it := range n.items {
				collect(it)
			}
		case *template:
			items = append(items, n)
		case *nullItem:
			nullable = true
		case *tableColumn:
		default:
			panic(invariant.Broken("columnItems has no case for node %T", n))
		}
	}
	collect(n)
	return items, nullable
}

// itemKind is how a column item comes by its datatype, ranked in this order: none, read from a typed column, or declared.
type itemKind int

const (
	kindText itemKind = iota
	kindReads
	kindDeclares
)

func kindOf(t *template) itemKind {
	switch {
	case t.datatype != nil:
		return kindDeclares
	case itemDatatype(t) != DataTypeString:
		return kindReads
	}
	return kindText
}

// disagreement names the fix for items a and b of one column holding different datatypes, fixing
// the one that declares least, a when both declare alike.
func disagreement(a *template, da DataType, b *template, db DataType) error {
	c := clash{fix: a, other: b, want: db, held: fmt.Sprintf("its items hold %s and %s; a column holds one datatype", da, db)}
	if kindOf(b) < kindOf(a) {
		c.fix, c.other, c.want = b, a, da
	}
	message, ok := disagreements[[2]itemKind{kindOf(c.fix), kindOf(c.other)}]
	if !ok {
		panic(invariant.Broken("disagreement has no message for item kinds %d and %d; two text items hold one datatype, so they never disagree", kindOf(c.fix), kindOf(c.other)))
	}
	return message(c)
}

// clash is two items of one column holding different datatypes: other holds want; held says
// what each holds.
type clash struct {
	fix, other *template
	want       DataType
	held       string
}

// disagreements is a clash's message by the kinds of its items, fix's first.
var disagreements = map[[2]itemKind]func(clash) error{
	{kindDeclares, kindDeclares}: clash.bothDeclare,
	{kindReads, kindReads}:       clash.retypeRead,
	{kindReads, kindDeclares}:    clash.retypeRead,
	{kindText, kindReads}:        clash.textBesideRead,
	{kindText, kindDeclares}:     clash.declareText,
}

func (c clash) bothDeclare() error { return errors.New(c.held) }

// fits reports whether fix may hold want.
func (c clash) fits() bool { return (&valueProof{}).proveColumnItem(c.fix).Not[c.want] == "" }

func (c clash) retypeRead() error {
	if c.fits() {
		return fmt.Errorf("%s, so %s", c.held, typedAs(c.fix, c.want))
	}
	return fmt.Errorf("%s, so to read %q as text, %s", c.held, c.fix.format, typedAs(c.fix, DataTypeString))
}

func (c clash) textBesideRead() error {
	if !c.fits() {
		return fmt.Errorf(`item %q is not %s, the datatype item %q takes from the column it reads; to read that column as text, %s`, c.fix.format, datatype.Noun(c.want), c.other.format, typedAs(c.other, DataTypeString))
	}
	return c.declareText()
}

func (c clash) declareText() error {
	if c.fix.fromString { // an object may carry a weight, which this spelling would drop
		return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, c.fix.format, c.fix.format, c.want)
	}
	return fmt.Errorf(`item %q declares no datatype beside one holding %s; a column holds one, so give it "datatype": %q`, c.fix.format, c.want, c.want)
}

// typedAs names the spelling giving a column-read item datatype d, keeping the other keys an object
// item carries.
func typedAs(t *template, d DataType) string {
	if t.fromString {
		return fmt.Sprintf(`write %q as {"format":%q,"datatype":%q}`, t.format, t.format, d)
	}
	return fmt.Sprintf(`give %q "datatype": %q`, t.format, d)
}

// columnRead is a record's column read by a format that only reads one reference or name,
// which is the column: it takes the column's datatype and null. category and field name it;
// category is "" for a column of the reading template's own record, which checkColumns reaches
// on its own.
type columnRead struct {
	a               arm
	category, field string
	column          node
}

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name, and nil where it does not or t declares a string, which reads the column as text.
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
		start, tail = b.binder, append(append([]string{splitArm(b.ref, nil).head}, targets[b].tail...), a.tail...)
	case a.kind == namedRead:
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		start, category = targets[b].start, categoryOf(b.binder.refs.byName[b.ref].head)
		tail = append(targets[b].tail[:len(targets[b].tail):len(targets[b].tail)], a.tail...)
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
