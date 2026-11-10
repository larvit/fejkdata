package fejkdata

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/proven"
)

// valueProof proves what typed columns and their calc operands hold, each node once per proof.
// links is the load's, where the proof runs before finish, and nil after it.
type valueProof struct {
	memo       map[node]proven.Facts
	columnMemo map[node]proven.Facts
	links      *links
}

// checkDatatype rejects a typed column item some render of which is not text of its datatype.
func (p *valueProof) checkDatatype(label string, n node) error {
	t, ok := n.(*template)
	if !ok || t.datatype == nil || *t.datatype == DataTypeString {
		return nil
	}
	if reason := p.proveColumnItem(t).Not[*t.datatype]; reason != "" {
		return fmt.Errorf("%s: datatype %s: %s", label, *t.datatype, reason)
	}
	return nil
}

// checkField rejects a column a field of Go type ft cannot fill: a datatype other than the one the
// Go type sets, a null outside a pointer, or a value its kind's datatype or range refuses.
func (p *valueProof) checkField(label string, ft reflect.Type, column node) error {
	elem := ft
	if ft.Kind() == reflect.Pointer {
		elem = ft.Elem()
	}
	kind := columnKinds[elem.Kind()]
	items, _ := columnItems(column)
	for _, it := range items {
		if it.datatype != nil && *it.datatype != kind.datatype {
			return fmt.Errorf("%s: its Go type %s sets the datatype %s; drop \"datatype\"", label, ft, kind.datatype)
		}
	}
	if ft.Kind() != reflect.Pointer && p.proveColumn(column).Nullable {
		return fmt.Errorf("%s: its tag can draw null, which %s cannot hold; make it *%s", label, ft, ft)
	}
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

// proveColumn proves a column over what its items draw, a null item marking it null rather than
// rendering "".
func (p *valueProof) proveColumn(n node) proven.Facts {
	if v, done := p.columnMemo[n]; done {
		return v
	}
	if p.columnMemo == nil {
		p.columnMemo = map[node]proven.Facts{}
	}
	items, nullable := columnItems(n)
	var v proven.Facts
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

func (p *valueProof) prove(n node) proven.Facts {
	if v, done := p.memo[n]; done {
		return v
	}
	if p.memo == nil {
		p.memo = map[node]proven.Facts{}
	}
	var v proven.Facts
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
		v = proven.Unproven(fmt.Sprintf("%q renders a row of %s, which is composed text", n.t.formatTemplate.format, n.t.rows.Segment()))
	case *nullItem:
		v = proven.Unproven(`it reads a null, which renders "" outside its own column`)
	default:
		panic(invariant.Broken("prove has no case for node %T", n))
	}
	p.memo[n] = v
	return v
}

// proveCells proves a table column over every cell it may render.
func (p *valueProof) proveCells(c *tableColumn) proven.Facts {
	var v proven.Facts
	for r := 0; r < c.t.rows.Len(); r++ {
		var w proven.Facts
		if cell := c.t.cellTemplate(r, c.i); cell != nil {
			w = p.prove(cell)
		} else {
			w = proven.Literal(c.t.rows.Cell(r, c.i))
		}
		if r == 0 {
			v = w
		} else {
			v = v.Or(w)
		}
	}
	return v
}

func (p *valueProof) proveUnion(nodes []node) proven.Facts {
	v := p.prove(nodes[0])
	for _, n := range nodes[1:] {
		v = v.Or(p.prove(n))
	}
	return v
}

// proveTemplate proves a template that renders one value: fixed text, or a format whose one token
// that prints stands alone.
func (p *valueProof) proveTemplate(t *template) proven.Facts {
	lit, fixed := t.fixedText()
	o, one := t.printedOp(p.links)
	switch {
	case t.repeat != 1:
		return proven.Unproven(fmt.Sprintf("%q carries a repeat, which composes text rather than one value", t.format))
	case fixed:
		return proven.Literal(lit)
	case !one:
		v := proven.Unproven(notOneValue(t.format, typedCalls))
		v.NotOperand = notOneValue(t.format, operandCalls)
		return v
	}
	body, name, args := o.Body, o.Fn, o.Args
	switch {
	case o.Kind == grammar.LiteralRun:
		return proven.Literal(o.Lit)
	case o.Kind == grammar.PathRead && (o.Distinct != nil || slices.ContainsFunc(o.arms, readsJoined)):
		return joinedText(body)
	case p.mayFallBack(o):
		return fallsBack(body)
	case o.Kind == grammar.PathRead:
		var leaves []node
		for _, a := range o.arms {
			leaves = append(leaves, a.leaves...)
		}
		return p.proveUnion(leaves)
	case name == builtinfunc.CalcName:
		return p.proveCalc(o)
	case builtinfunc.IsTransform(name):
		return proven.Unproven(fmt.Sprintf("{%s} rewrites text rather than printing a value; write the values it would print", body))
	}
	if v, numeric := builtinfunc.ProveNumber(name, body, args); numeric {
		return v
	}
	return proven.Printing(body, DataTypeString, proven.Facts{NotOperand: fmt.Sprintf("{%s} prints text, not a number", body)})
}

func (p *valueProof) proveCalc(o op) proven.Facts {
	body, args := o.Body, o.Args
	nodes := operandNodes(o)
	v, doubt := proven.OfCalc(builtinfunc.ParsedCalc(args[0]).Expr, func(name string) proven.Facts {
		if slices.ContainsFunc(o.operands, func(a arm) bool { return a.spelling == name && p.readsFallback(a) }) {
			return fallsBack(name)
		}
		return p.proveUnion(nodes(name))
	})
	if doubt != "" {
		return proven.Unproven(fmt.Sprintf("{%s}: %s", body, doubt))
	}
	return proven.PrintedNumber(body, v, builtinfunc.CalcDecimals(args))
}

// proveArm proves what a prints: one of its leaves, or text where it reads a distinct draw's members
// joined.
func (p *valueProof) proveArm(a arm) proven.Facts {
	switch {
	case readsJoined(a):
		return joinedText(a.spelling)
	case p.readsFallback(a):
		return fallsBack(a.spelling)
	}
	return p.proveUnion(a.leaves)
}

// mayFallBack says whether o may print its where clause's fallback, or a path's empty default.
func (p *valueProof) mayFallBack(o op) bool {
	if o.draw != nil {
		return o.draw.mayFallBack()
	}
	return slices.ContainsFunc(o.arms, p.readsFallback)
}

func (p *valueProof) readsFallback(a arm) bool {
	if a.kind != namedRead {
		return false
	}
	v := p.links.binding(a.named).values
	return v != nil && v.mayFallBack()
}

func fallsBack(token string) proven.Facts {
	return proven.Unproven(fmt.Sprintf("{%s} may print its fallback or default, where the strings it reads leave no value", token))
}

func joinedText(token string) proven.Facts {
	return proven.Printing(token, DataTypeString, proven.Facts{NotOperand: "{" + token + "} joins several values as text"})
}

var typedCalls, operandCalls = builtinfunc.TypedCalls(), builtinfunc.OperandCalls()

func notOneValue(format, calls string) string {
	return fmt.Sprintf("%q is not one value; write one literal or one %s, or read one", format, calls)
}
