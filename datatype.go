package fejkdata

import (
	"errors"
	"fmt"

	"github.com/larvit/fejkdata/internal/datatype"
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

// position is where a JSON value sits, which decides whether it may carry a datatype or
// be null.
type position int

const (
	inFormat position = iota // rendered by a format, so neither
	atTop                    // a category or an inline template, whose fields may be columns
	inColumn                 // a column, or a choice item standing in for one
)

// datatypeOf reads a template's "datatype" (default DataTypeString).
func datatypeOf(m map[string]any, pos position) (DataType, error) {
	v, ok := m["datatype"]
	if !ok {
		return DataTypeString, nil
	}
	name, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("datatype must be a string, got %T", v)
	}
	if name == DataTypeString.String() {
		return 0, fmt.Errorf("datatype %q is the default, so it has no effect; drop it", name)
	}
	for d := DataTypeInteger; d < datatype.Count; d++ {
		if name != d.String() {
			continue
		}
		if pos != inColumn {
			return 0, errors.New("datatype only types a record column — a field of the top-level template — so it has no effect here")
		}
		return d, nil
	}
	return 0, fmt.Errorf(`datatype takes "integer", "number" or "boolean", got %q`, name)
}

// checkColumns rejects a record column whose items hold different datatypes, checking a column
// after the columns its items read, so a column read is named before its readers.
func checkColumns(s nodeScope) error {
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
	if t.datatype == DataTypeString && t.readsColumn != nil {
		return columnDatatype(t.readsColumn.column)
	}
	return t.datatype
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
	case t.datatype != DataTypeString:
		return kindDeclares
	case itemDatatype(t) != DataTypeString:
		return kindReads
	}
	return kindText
}

// disagreement names the fix for items a and b of one column holding different datatypes, fixing
// the one that declares least, a when both declare alike.
func disagreement(a *template, da DataType, b *template, db DataType) error {
	fix, other, want := a, b, db
	if kindOf(b) < kindOf(a) {
		fix, other, want = b, a, da
	}
	held := fmt.Sprintf("its items hold %s and %s; a column holds one datatype", da, db)
	fits := (&valueProof{}).proveColumnItem(fix).Not[want] == ""
	switch kindOf(fix) {
	case kindDeclares:
		return errors.New(held)
	case kindReads:
		if fits {
			return fmt.Errorf("%s, so %s", held, typedAs(fix, want))
		}
		return fmt.Errorf("%s, so to read %q as text, %s", held, fix.format, asText(fix))
	case kindText:
		switch kindOf(other) {
		case kindText:
			panic(invariant.Broken("two text items hold one datatype, so they never disagree"))
		case kindReads:
			if !fits {
				return fmt.Errorf(`item %q is not %s, the datatype item %q takes from the column it reads; to read that column as text, %s`, fix.format, datatype.Noun(want), other.format, asText(other))
			}
		}
		if fix.fromString { // an object may carry a weight, which this spelling would drop
			return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, fix.format, fix.format, want)
		}
		return fmt.Errorf(`item %q declares no datatype beside one holding %s; a column holds one, so give it "datatype": %q`, fix.format, want, want)
	}
	panic(invariant.Broken("disagreement has no case for item kind %d", kindOf(fix)))
}

// typedAs names the spelling giving a column-read item datatype d, keeping the other keys an object
// item carries.
func typedAs(t *template, d DataType) string {
	if t.fromString {
		return fmt.Sprintf(`write %q as {"format":%q,"datatype":%q}`, t.format, t.format, d)
	}
	return fmt.Sprintf(`give %q "datatype": %q`, t.format, d)
}

// asText names the spelling that reads the column a column-read item reads as text, keeping the
// other keys an object item carries.
func asText(t *template) string {
	if t.fromString {
		return fmt.Sprintf(`write {"format":"{text}","text":%q}`, t.format)
	}
	return fmt.Sprintf(`set its "format" to "{text}" and add "text": %q`, t.format)
}
