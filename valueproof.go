package fejkdata

import (
	"fmt"
	"reflect"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/proven"
)

// valueProof proves what typed columns and their calc operands hold, each node once per scope.
type valueProof struct {
	memo       map[node]proven.Value
	columnMemo map[node]proven.Value
}

// checkDatatype rejects a typed column item some render of which is not text of its datatype,
// and one restating the datatype of the column it is.
// docs/decisions.md#a-column-that-only-reads-one-reference-or-name-is-the-column-it-reads
func (p *valueProof) checkDatatype(path string, n node) error {
	t, ok := n.(*template)
	if !ok || t.datatype == DataTypeString {
		return nil
	}
	if r := t.link.readsColumn; r != nil && columnDatatype(r.column) == t.datatype {
		return fmt.Errorf(`%s: %s takes datatype %s from the column it reads; drop "datatype"`, path, t.format, t.datatype)
	}
	if reason := p.proveColumnItem(t).Not[t.datatype]; reason != "" {
		return fmt.Errorf("%s: datatype %s: %s", path, t.datatype, reason)
	}
	return nil
}

// checkField rejects a column a field of Go type ft cannot fill: a datatype, which the Go type
// sets, a null outside a pointer, or a value its kind's datatype or range refuses.
func (p *valueProof) checkField(label string, ft reflect.Type, column node) error {
	items, _ := columnItems(column)
	for _, it := range items {
		if it.datatype != DataTypeString {
			return fmt.Errorf("%s: its Go type %s sets the datatype; drop \"datatype\"", label, ft)
		}
	}
	elem := ft
	if ft.Kind() == reflect.Pointer {
		elem = ft.Elem()
	} else if p.proveColumn(column).Nullable {
		return fmt.Errorf("%s: its tag can draw null, which %s cannot hold; make it *%s", label, ft, ft)
	}
	kind := columnKinds[elem.Kind()]
	if kind.datatype == DataTypeString {
		return nil
	}
	for _, it := range items {
		v := p.proveColumnItem(it)
		if reason := v.Not[kind.datatype]; reason != "" {
			return fmt.Errorf("%s (%s): %s", label, ft, reason)
		}
		if !kind.holds(v) {
			return fmt.Errorf("%s (%s): %q is not proven within %s; narrow it to that range, or make the field %s", label, ft, it.format, elem.Kind(), kind.wider)
		}
	}
	return nil
}

// proveColumnItem proves a column item: what it renders, or, when it is a column it reads, that column.
func (p *valueProof) proveColumnItem(t *template) proven.Value {
	if t.link.readsColumn == nil {
		return p.prove(t)
	}
	return p.proveColumn(t.link.readsColumn.column)
}

// proveColumn proves a column over what its items draw, a null item marking it null rather than
// rendering "".
func (p *valueProof) proveColumn(n node) proven.Value {
	if v, done := p.columnMemo[n]; done {
		return v
	}
	if p.columnMemo == nil {
		p.columnMemo = map[node]proven.Value{}
	}
	items, nullable := columnItems(n)
	var v proven.Value
	for i, it := range items {
		if w := p.proveColumnItem(it); i == 0 {
			v = w
		} else {
			v = v.Or(w)
		}
	}
	v.Nullable = v.Nullable || nullable
	p.columnMemo[n] = v
	return v
}

func (p *valueProof) prove(n node) proven.Value {
	if v, done := p.memo[n]; done {
		return v
	}
	if p.memo == nil {
		p.memo = map[node]proven.Value{}
	}
	var v proven.Value
	switch n := n.(type) {
	case *choice:
		v = p.proveUnion(n.items)
	case *template:
		v = p.proveTemplate(n)
	case *tableColumn:
		v = p.proveCells(n)
	case *table:
		v = p.prove(n.rowNode)
	case *tableRow:
		v = proven.Unproven(fmt.Sprintf("%q renders a row of %s, which is composed text", n.t.formatTemplate.format, n.t.segment))
	case *nullItem:
		v = proven.Unproven(`it reads a null, which renders "" outside its own column`)
	default:
		panic(internalError("prove has no case for node %T", n))
	}
	p.memo[n] = v
	return v
}

// proveCells proves a table column over every cell it may render.
func (p *valueProof) proveCells(c *tableColumn) proven.Value {
	var v proven.Value
	for r := 0; r < c.t.rowCount(); r++ {
		var w proven.Value
		if cell := c.t.cellTemplate(r, c.i); cell != nil {
			w = p.prove(cell)
		} else {
			w = proven.Literal(c.t.cell(r, c.i))
		}
		if r == 0 {
			v = w
		} else {
			v = v.Or(w)
		}
	}
	return v
}

func (p *valueProof) proveUnion(nodes []node) proven.Value {
	v := p.prove(nodes[0])
	for _, n := range nodes[1:] {
		v = v.Or(p.prove(n))
	}
	return v
}

// proveTemplate proves a template that renders one value: fixed text, or a format that is
// one token alone.
func (p *valueProof) proveTemplate(t *template) proven.Value {
	lit, fixed := t.fixedText()
	switch {
	case t.repeat != 1:
		return proven.Unproven(fmt.Sprintf("%q carries a repeat, which composes text rather than one value", t.format))
	case fixed:
		return proven.Literal(lit)
	case len(t.compiled.ops) != 1:
		v := proven.Unproven(notOneValue(t.format, typedCalls))
		v.NotOperand = notOneValue(t.format, operandCalls)
		return v
	}
	o := t.compiled.ops[0]
	body, name, args := o.Body, o.Fn, o.Args
	switch {
	case o.Kind == grammar.NameRead:
		var leaves []node
		for _, a := range o.arms {
			leaves = append(leaves, a.leaves...)
		}
		return p.proveUnion(leaves)
	case name == "calc":
		return p.proveCalc(o)
	case builtinfunc.IsTransform(name):
		return proven.Unproven(fmt.Sprintf("{%s} rewrites text rather than printing a value; write the values it would print", body))
	}
	if v, numeric := builtinfunc.ProveNumber(name, body, args); numeric {
		return v
	}
	return proven.Printing(body, DataTypeString, proven.Value{NotOperand: fmt.Sprintf("{%s} prints text, not a number", body)})
}

func (p *valueProof) proveCalc(o op) proven.Value {
	body, args := o.Body, o.Args
	nodes := operandNodes(o)
	v, doubt := proven.Calc(builtinfunc.ParsedCalc(args[0]).Expr, func(name string) proven.Value { return p.proveUnion(nodes(name)) })
	if doubt != "" {
		return proven.Unproven(fmt.Sprintf("{%s}: %s", body, doubt))
	}
	return proven.PrintedNumber(body, v, builtinfunc.CalcDecimals(args))
}

var typedCalls, operandCalls = builtinfunc.NumberCalls(false), builtinfunc.NumberCalls(true)

func notOneValue(format, calls string) string {
	return fmt.Sprintf("%q is not one value; write one literal or one %s, or read one", format, calls)
}
