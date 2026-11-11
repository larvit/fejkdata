package fence

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/proven"
)

// valueProof proves what the nodes of a finished tree render, each node once per proof.
type valueProof struct {
	memo       map[tree.Node]proven.Facts
	columnMemo map[tree.Node]proven.Facts
}

// checkDatatype rejects a typed column item some render of which is not text of its datatype.
func (p *valueProof) checkDatatype(label string, n tree.Node) error {
	t, ok := n.(*tree.Template)
	if !ok || t.DataType() == nil || *t.DataType() == datatype.String {
		return nil
	}
	if reason := p.proveColumnItem(t).Not[*t.DataType()]; reason != "" {
		return fmt.Errorf("%s: datatype %s: %s", label, *t.DataType(), reason)
	}
	return nil
}

// proveColumn proves a column over what its items draw, a null item marking it null rather than
// rendering "".
func (p *valueProof) proveColumn(n tree.Node) proven.Facts {
	if v, done := p.columnMemo[n]; done {
		return v
	}
	if p.columnMemo == nil {
		p.columnMemo = map[tree.Node]proven.Facts{}
	}
	items, nullable := ColumnItems(n)
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

// Prove is what every render of n can print.
func Prove(n tree.Node) proven.Facts { return (&valueProof{}).prove(n) }

func (p *valueProof) prove(n tree.Node) proven.Facts {
	if v, done := p.memo[n]; done {
		return v
	}
	if p.memo == nil {
		p.memo = map[tree.Node]proven.Facts{}
	}
	var v proven.Facts
	switch n := n.(type) {
	case *tree.Choice:
		v = p.proveUnion(n.Items())
	case *tree.Template:
		v = p.proveTemplate(n)
	case *tree.TableColumn:
		v = p.proveCells(n)
	case *tree.Table:
		v = p.prove(n.RowNode())
	case *tree.TableRow:
		v = proven.Unproven(fmt.Sprintf("%q renders a row of %s, which is composed text", n.Table().FormatTemplate().Format(), n.Table().Rows().Segment()))
	case *tree.NullItem:
		v = proven.NeverNumber(`it reads a null, which renders "" outside its own column`)
	default:
		panic(invariant.Broken("prove has no case for node %T", n))
	}
	p.memo[n] = v
	return v
}

// proveCells proves a table column over every cell it may render.
func (p *valueProof) proveCells(c *tree.TableColumn) proven.Facts {
	var v proven.Facts
	for r := 0; r < c.Table().Rows().Len(); r++ {
		var w proven.Facts
		if cell := c.Table().CellTemplate(r, c.Index()); cell != nil {
			w = p.prove(cell)
		} else {
			w = proven.Literal(c.Table().Rows().Cell(r, c.Index()))
		}
		if r == 0 {
			v = w
		} else {
			v = v.Or(w)
		}
	}
	return v
}

func (p *valueProof) proveUnion(nodes []tree.Node) proven.Facts {
	v := p.prove(nodes[0])
	for _, n := range nodes[1:] {
		v = v.Or(p.prove(n))
	}
	return v
}

// repeated proves a template whose repeat is not 1: a repeat of 0 renders no text, any other composes it.
func repeated(t *tree.Template) proven.Facts {
	if t.Repeat() == 0 {
		return proven.NeverNumber(fmt.Sprintf("%q repeats 0 times, so it renders nothing", t.Format()))
	}
	return proven.Unproven(fmt.Sprintf("%q carries a repeat, which composes text rather than one value", t.Format()))
}

// proveTemplate proves a template that renders one value: fixed text, or a format whose one token
// that prints stands alone.
func (p *valueProof) proveTemplate(t *tree.Template) proven.Facts {
	lit, fixed := t.FixedText()
	o, one := load.PrintedOp(t.Compiled().Ops())
	switch {
	case t.Repeat() != 1:
		return repeated(t)
	case fixed:
		return proven.Literal(lit)
	case !one:
		v := proven.Unproven(notOneValue(t.Format(), typedCalls))
		v.NotOperand = notOneValue(t.Format(), operandCalls)
		return v
	}
	tok := o.Token()
	body, name, args := tok.Body, tok.Fn, tok.Args
	switch {
	case tok.Kind == grammar.LiteralRun:
		return proven.Literal(tok.Lit)
	case tok.Kind == grammar.PathRead && (tok.Distinct != nil || slices.ContainsFunc(o.Alternatives(), tree.Read.ReadsJoined)):
		return joinedText(body)
	case p.mayFallBack(o):
		return fallsBack(body)
	case tok.Kind == grammar.PathRead:
		var leaves []tree.Node
		for _, a := range o.Alternatives() {
			leaves = append(leaves, a.Leaves()...)
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
	return proven.Printing(body, datatype.String, proven.Facts{NotOperand: fmt.Sprintf("{%s} prints text, not a number", body)})
}

func (p *valueProof) proveCalc(o tree.Op) proven.Facts {
	body, args := o.Token().Body, o.Token().Args
	v, doubt := proven.OfCalc(builtinfunc.ParsedCalc(args[0]).Expr, p.operand(o))
	if doubt != "" {
		return proven.Unproven(fmt.Sprintf("{%s}: %s", body, doubt))
	}
	return proven.PrintedNumber(body, v, builtinfunc.CalcDecimals(args))
}

// operand proves what each operand of the calc o renders, by its spelling.
func (p *valueProof) operand(o tree.Op) func(name string) proven.Facts {
	return func(name string) proven.Facts {
		a := o.Operands()[slices.IndexFunc(o.Operands(), func(a tree.Read) bool { return a.Spelling() == name })]
		if len(a.Leaves()) == 0 {
			panic(invariant.Broken("calc operand %q was read before it was compiled", name))
		}
		return p.proveRead(a)
	}
}

// checkCalcs refuses each calc of t with an operand that fixed text proves never a number, or a
// divisor that fixed values prove always zero.
func (p *valueProof) checkCalcs(label string, t *tree.Template) error {
	for _, o := range t.Compiled().Ops() {
		tok := o.Token()
		if tok.Fn != builtinfunc.CalcName {
			continue
		}
		c, operand := builtinfunc.ParsedCalc(tok.Args[0]), p.operand(o)
		err := p.checkOperands(tok.Args[0], c, operand)
		if err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, tok.Body, err)
		}
	}
	return nil
}

func (p *valueProof) checkOperands(text string, c grammar.Calc, operand func(name string) proven.Facts) error {
	for _, name := range c.Operands {
		if v := operand(name); v.Never {
			return fmt.Errorf("calc(%q): operand %q is never a number: %s", text, name, v.NotOperand)
		}
	}
	if divisor, zero := zeroDivisor(c.Expr, operand); zero {
		return fmt.Errorf("calc(%q) divides by %s, which is always zero", text, divisor)
	}
	return nil
}

// zeroDivisor finds a division whose right side folds to zero from operands that are each one value.
func zeroDivisor(n grammar.CalcNode, operand func(name string) proven.Facts) (string, bool) {
	switch n := n.(type) {
	case grammar.CalcNeg:
		return zeroDivisor(n.X, operand)
	case grammar.CalcBin:
		varies := false
		v, doubt := proven.OfCalc(n.R, func(name string) proven.Facts {
			f := operand(name)
			varies = varies || f.Lo != f.Hi
			return f
		})
		if n.Operator == '/' && doubt == "" && !varies && v.Lo == 0 && v.Hi == 0 {
			return grammar.CalcText(n.R), true
		}
		if d, zero := zeroDivisor(n.L, operand); zero {
			return d, true
		}
		return zeroDivisor(n.R, operand)
	}
	return "", false
}

// proveRead proves what a prints: one of its leaves, or text where it reads a distinct draw's members
// joined.
func (p *valueProof) proveRead(a tree.Read) proven.Facts {
	switch {
	case a.ReadsJoined():
		return joinedText(a.Spelling())
	case p.readsFallback(a):
		return fallsBack(a.Spelling())
	}
	return p.proveUnion(a.Leaves())
}

// mayFallBack says whether o may print its where clause's fallback, or a path's empty default.
func (p *valueProof) mayFallBack(o tree.Op) bool {
	if o.Draw() != nil {
		return o.Draw().MayFallBack()
	}
	return slices.ContainsFunc(o.Alternatives(), p.readsFallback)
}

func (p *valueProof) readsFallback(a tree.Read) bool {
	if a.Kind() != tree.NamedRead {
		return false
	}
	v := a.Named().Values()
	return v != nil && v.MayFallBack()
}

func fallsBack(token string) proven.Facts {
	return proven.Unproven(fmt.Sprintf("{%s} leaves out a string read from a name, so it may print its fallback or an empty string, which are text; leave out quoted strings instead, as in not '1'", token))
}

func joinedText(token string) proven.Facts {
	return proven.Printing(token, datatype.String, proven.Facts{NotOperand: "{" + token + "} joins several values as text", Never: true})
}

var typedCalls, operandCalls = builtinfunc.TypedCalls(), builtinfunc.OperandCalls()

func notOneValue(format, calls string) string {
	return fmt.Sprintf("%q is not one value; write one literal or one %s, or read one", format, calls)
}
