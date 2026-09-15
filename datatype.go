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

// checkColumns rejects a record column whose items hold different datatypes.
func checkColumns(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !t.record {
		return nil
	}
	for _, name := range recordColumns(t) {
		if _, err := columnDatatype(t.fields[name]); err != nil {
			return fmt.Errorf("%s: field %q: %w", path, name, err)
		}
	}
	return nil
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

// itemDatatype is the datatype a column item declares, else that of the column it reads whole.
func itemDatatype(t *template) DataType {
	if t.datatype != DataTypeString {
		return t.datatype
	}
	return readDatatype(t)
}

// readDatatype is the datatype of the column t reads whole; a string when it reads none.
func readDatatype(t *template) DataType {
	if t.inherits == nil {
		return DataTypeString
	}
	d, _ := columnDatatype(t.inherits) // checkColumns refuses that column where it sits
	return d
}

// columnItems is a column's template items, its choices unwrapped, and whether one is null
// or reads whole a column that can be.
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
			if n.inherits != nil {
				_, inherited := columnItems(n.inherits)
				nullable = nullable || inherited
			}
		case *null:
			nullable = true
		}
	}
	collect(n)
	return items, nullable
}

// disagreement names the fix for two items of one column holding different datatypes.
func disagreement(a *template, da DataType, b *template, db DataType) error {
	bare, want := b, da
	if da == DataTypeString {
		bare, want = a, db
	}
	switch {
	case da != DataTypeString && db != DataTypeString:
		return fmt.Errorf("its items declare %s and %s; a column holds one datatype", da, db)
	case bare.fields == nil: // a JSON string; an object, which may carry a weight, has a fields map
		return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, bare.format, bare.format, want)
	}
	return fmt.Errorf(`item %q declares no datatype beside one declaring %s; a column holds one, so give it "datatype": %q`, bare.format, want, want)
}
