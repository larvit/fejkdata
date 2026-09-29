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
			if r := it.link.readsColumn; r != nil {
				if err := check(r.a.key[1:], r.a.tail[0], r.column); err != nil {
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
				return fmt.Errorf("%s: field %q: %w", path, name, disagreement(items[0], first, it, d))
			}
		}
		return nil
	}
	return s(func(path string, n node) error {
		t, ok := n.(*template)
		if !ok || !t.isRecord {
			return nil
		}
		for _, name := range sortedNames(t.fields) {
			if err := check(path, name, t.fields[name]); err != nil {
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
	if t.datatype == DataTypeString && t.link.readsColumn != nil {
		return columnDatatype(t.link.readsColumn.column)
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
			panic(internalError("columnItems has no case for node %T", n))
		}
	}
	collect(n)
	return items, nullable
}

// itemKind is how a column item comes by its datatype: none, read from a typed column, or declared.
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

// itemPair is two items of one column holding different datatypes: fix is the one to change,
// and want the datatype of the other.
type itemPair struct {
	first, second DataType
	fix, other    *template
	want          DataType
}

// disagreementKey is the two items' kinds and whether fix's values prove to hold want.
type disagreementKey struct {
	fix, other itemKind
	proven     bool
}

var disagreements = map[disagreementKey]func(d itemPair) error{
	{kindDeclares, kindDeclares, false}: holdOne,
	{kindDeclares, kindDeclares, true}:  holdOne,
	{kindReads, kindDeclares, false}:    readAsText,
	{kindReads, kindDeclares, true}:     readTyped,
	{kindReads, kindReads, false}:       readAsText,
	{kindReads, kindReads, true}:        readTyped,
	{kindText, kindDeclares, false}:     declare,
	{kindText, kindDeclares, true}:      declare,
	{kindText, kindReads, false}:        otherAsText,
	{kindText, kindReads, true}:         declare,
}

// disagreement names the fix for items a and b of one column holding different datatypes, fixing
// the one that declares least, a when both declare alike.
func disagreement(a *template, da DataType, b *template, db DataType) error {
	d := itemPair{first: da, second: db, fix: a, other: b, want: db}
	if kindOf(b) < kindOf(a) {
		d.fix, d.other, d.want = b, a, da
	}
	proven := (&valueProof{}).proveColumnItem(d.fix).not[d.want] == ""
	return disagreementFix(disagreementKey{kindOf(d.fix), kindOf(d.other), proven})(d)
}

func disagreementFix(k disagreementKey) func(d itemPair) error {
	fix := disagreements[k]
	if fix == nil {
		panic(internalError("no disagreement names the fix for %+v", k))
	}
	return fix
}

func holdOne(d itemPair) error {
	return fmt.Errorf("its items hold %s and %s; a column holds one datatype", d.first, d.second)
}

func readTyped(d itemPair) error {
	return fmt.Errorf("its items hold %s and %s; a column holds one datatype, so %s", d.first, d.second, typedAs(d.fix, d.want))
}

func readAsText(d itemPair) error {
	return fmt.Errorf("its items hold %s and %s; a column holds one datatype, so to read %q as text, %s", d.first, d.second, d.fix.format, asText(d.fix))
}

func otherAsText(d itemPair) error {
	return fmt.Errorf(`item %q is not %s, the datatype item %q takes from the column it reads; to read that column as text, %s`, d.fix.format, dataTypeNouns[d.want], d.other.format, asText(d.other))
}

func declare(d itemPair) error {
	if d.fix.fromString { // an object may carry a weight, which this spelling would drop
		return fmt.Errorf(`item %q declares no datatype, and a column holds one; write it as {"format":%q,"datatype":%q}`, d.fix.format, d.fix.format, d.want)
	}
	return fmt.Errorf(`item %q declares no datatype beside one holding %s; a column holds one, so give it "datatype": %q`, d.fix.format, d.want, d.want)
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
