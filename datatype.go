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

var dataTypeNames = [...]string{"string", "integer", "number", "boolean"}

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
	atTop                    // a category or an inline template, whose fields are the columns
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

// columnDatatype is the datatype a column's items declare. They must agree, since a
// column holds one; a column only ever null is a string.
func columnDatatype(n node) (DataType, error) {
	var declared []DataType
	var collect func(node)
	collect = func(n node) {
		switch n := n.(type) {
		case *choice:
			for _, it := range n.items {
				collect(it)
			}
		case *template:
			declared = append(declared, n.datatype)
		}
	}
	collect(n)
	if len(declared) == 0 {
		return DataTypeString, nil
	}
	for _, d := range declared {
		if d != declared[0] {
			return declared[0], fmt.Errorf("its items declare %s and %s; a column holds one datatype, so give every item the same", declared[0], d)
		}
	}
	return declared[0], nil
}

// datatypeSpec is what a datatype's text must satisfy: a grammar, the states a render
// may end in, and how an error names the datatype.
type datatypeSpec struct {
	grammar *grammar
	accept  uint32
	noun    string
}

var datatypeSpecs = map[DataType]datatypeSpec{
	DataTypeInteger: {numberGrammar, integerAccept, "an integer"},
	DataTypeNumber:  {numberGrammar, numberAccept, "a number"},
	DataTypeBoolean: {booleanGrammar, booleanAccept, "a boolean"},
}

// datatypeCheck proves every render of a typed column is text its datatype takes. One
// check covers a scope, so a node several columns reach is read once per grammar.
type datatypeCheck struct {
	languages map[*grammar]*textLanguage
	proof     *calcProof
}

func (c *datatypeCheck) check(path string, n node) error {
	t, ok := n.(*template)
	if !ok || t.datatype == DataTypeString {
		return nil
	}
	spec := datatypeSpecs[t.datatype]
	w, escapes := c.language(spec.grammar).node(t, nil).escape(spec.accept)
	if !escapes {
		return nil
	}
	msg := fmt.Sprintf("%s: datatype %s, but it can render %s, which is not %s", path, t.datatype, w, spec.noun)
	if w.why != "" {
		msg += ": " + w.why
	}
	return errors.New(msg)
}

func (c *datatypeCheck) language(g *grammar) *textLanguage {
	if c.proof == nil {
		c.proof = newCalcProof()
		c.languages = map[*grammar]*textLanguage{}
	}
	l, made := c.languages[g]
	if !made {
		l = newTextLanguage(g, c.proof)
		c.languages[g] = l
	}
	return l
}
