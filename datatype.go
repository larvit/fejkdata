package fejkdata

import (
	"errors"
	"fmt"
)

// DataType is what a record column holds, which decides how a record writes its value.
type DataType int

// The datatypes a column declares with "datatype"; a column without one is a string.
const (
	DataTypeString DataType = iota
	DataTypeInteger
	DataTypeNumber
	DataTypeBoolean
)

var (
	dataTypeNames = [...]string{"string", "integer", "number", "boolean"}
	dataTypeNouns = [...]string{"text", "an integer", "a number", "a boolean"}
)

// String is the datatype as data spells it.
func (d DataType) String() string {
	if d < 0 || int(d) >= len(dataTypeNames) {
		return fmt.Sprintf("DataType(%d)", int(d))
	}
	return dataTypeNames[d]
}

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
	for d := DataTypeInteger; d <= DataTypeBoolean; d++ {
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
	var check func(path, name string, column node) error
	check = func(path, name string, column node) error {
		if checked[column] {
			return nil
		}
		checked[column] = true
		items, _ := columnItems(column)
		for _, it := range items {
			if r := it.readsColumn; r != nil {
				if err := check(r.a.key[1:], r.a.tail[0], r.column); err != nil {
					return err
				}
			}
		}
		if _, err := columnDatatype(column); err != nil {
			return fmt.Errorf("%s: field %q: %w", path, name, err)
		}
		return nil
	}
	return s(func(path string, n node) error {
		t, ok := n.(*template)
		if !ok || !t.record {
			return nil
		}
		for _, name := range recordColumns(t) {
			if err := check(path, name, t.fields[name]); err != nil {
				return err
			}
		}
		return nil
	})
}

// columnDatatype is the datatype a column's items hold. They must agree, since a
// column holds one; a column only ever null is a string.
func columnDatatype(n node) (DataType, error) {
	items, _ := columnItems(n)
	if len(items) == 0 {
		return DataTypeString, nil
	}
	first := itemDatatype(items[0])
	for _, t := range items[1:] {
		if d := itemDatatype(t); d != first {
			return first, disagreement(items[0], first, t, d)
		}
	}
	return first, nil
}

// itemDatatype is the datatype a column item declares, else that of the column it is.
func itemDatatype(t *template) DataType {
	if t.datatype != DataTypeString {
		return t.datatype
	}
	return readDatatype(t)
}

// readDatatype is the datatype of the column t is, reading it by one reference alone; a string
// when t reads none.
func readDatatype(t *template) DataType {
	if t.readsColumn == nil {
		return DataTypeString
	}
	items, _ := columnItems(t.readsColumn.column)
	if len(items) == 0 {
		return DataTypeString
	}
	return itemDatatype(items[0]) // checkColumns refuses that column where its items disagree
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
		case *null:
			nullable = true
		}
	}
	collect(n)
	return items, nullable
}

// disagreement names the fix for two items of one column holding different datatypes.
func disagreement(a *template, da DataType, b *template, db DataType) error {
	bare, typed, want := b, a, da
	if da == DataTypeString {
		bare, typed, want = a, b, db
	}
	switch {
	case da != DataTypeString && db != DataTypeString:
		return bothTyped(a, da, b, db)
	case typed.datatype == DataTypeString && (&valueProof{}).columnItem(bare).not[want] != "":
		return fmt.Errorf(`item %q is not %s, the datatype item %q takes from the column it reads; to read that column as text, %s`, bare.format, dataTypeNouns[want], typed.format, asText(typed))
	case bare.fromString: // an object may carry a weight, which this spelling would drop
		return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, bare.format, bare.format, want)
	}
	return fmt.Errorf(`item %q declares no datatype beside one declaring %s; a column holds one, so give it "datatype": %q`, bare.format, want, want)
}

// bothTyped names the fix for two items holding different datatypes: reading one that takes its
// datatype from the column it reads as text, since only a declared datatype can be edited away.
func bothTyped(a *template, da DataType, b *template, db DataType) error {
	read := a
	if a.datatype != DataTypeString {
		read = b
	}
	if read.datatype != DataTypeString {
		return fmt.Errorf("its items hold %s and %s; a column holds one datatype", da, db)
	}
	return fmt.Errorf("its items hold %s and %s; a column holds one datatype, so to read %q as text, %s", da, db, read.format, asText(read))
}

// asText names the spelling that reads the column a column-read item reads as text, keeping the
// other keys an object item carries.
func asText(t *template) string {
	if t.fromString {
		return fmt.Sprintf(`write {"format":"{text}","text":%q}`, t.format)
	}
	return fmt.Sprintf(`set its "format" to "{text}" and add "text": %q`, t.format)
}
