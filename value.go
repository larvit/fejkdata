package fejkdata

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// proven is what a proof knows of every render of a node: bounds on the number each
// reads as, and per datatype why some render's text is not one ("" when none).
type proven struct {
	lo, hi     float64
	nonZero    float64 // every value is at least this far from zero; 0 when one can be zero
	integral   bool
	notOperand string // why some render reads as no finite number, the way calc reads it
	not        [len(dataTypeNames)]string
}

// valueProof proves what typed columns and their calc operands hold, each node once per scope.
type valueProof struct {
	memo map[node]proven
}

// checkDatatype rejects a typed column some render of which is not text of its datatype.
func (p *valueProof) checkDatatype(path string, n node) error {
	t, ok := n.(*template)
	if !ok || t.datatype == DataTypeString {
		return nil
	}
	if err := p.prove(t, t.datatype); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// prove reports why some render of n is not text of datatype d.
func (p *valueProof) prove(n node, d DataType) error {
	if reason := p.of(n).not[d]; reason != "" {
		return fmt.Errorf("datatype %s: %s", d, reason)
	}
	return nil
}

func (p *valueProof) of(n node) proven {
	if v, done := p.memo[n]; done {
		return v
	}
	if p.memo == nil {
		p.memo = map[node]proven{}
	}
	var v proven
	switch n := n.(type) {
	case *choice:
		v = p.unite(n.items)
	case *template:
		v = p.template(n)
	default:
		v = unproven(`it reads a null, which renders "" outside its own column`)
	}
	p.memo[n] = v
	return v
}

func (p *valueProof) unite(nodes []node) proven {
	v := p.of(nodes[0])
	for _, n := range nodes[1:] {
		w := p.of(n)
		v.lo, v.hi, v.nonZero = min(v.lo, w.lo), max(v.hi, w.hi), min(v.nonZero, w.nonZero)
		v.integral = v.integral && w.integral
		if v.notOperand == "" {
			v.notOperand = w.notOperand
		}
		for d := range v.not {
			if v.not[d] == "" {
				v.not[d] = w.not[d]
			}
		}
	}
	return v
}

// template proves a template that renders one value: fixed text, or a format that is
// one token alone.
func (p *valueProof) template(t *template) proven {
	switch {
	case t.repeat != 1:
		return unproven(fmt.Sprintf("%q carries a repeat, which composes text rather than one value", t.format))
	case t.fixed:
		return literalValue(t.lit)
	case len(t.ops) != 1:
		v := unproven(notOneValue(t.format, "{int()}, {float()}, {seq()} or {calc()}"))
		v.notOperand = notOneValue(t.format, "{int()}, {float()}, {seq()}, {digits()} or {calc()}")
		return v
	}
	body := t.format[1 : len(t.format)-1]
	name, args, isFunc := funcCall(body)
	switch _, isTransform := transforms[name]; {
	case !isFunc:
		var leaves []node
		for _, a := range splitArms(body, t.refs) {
			leaves = append(leaves, pathLeaves(t.fields[a.key], a.tail)...)
		}
		return p.unite(leaves)
	case name == "calc":
		return p.calc(t, body, args)
	case builtins[name].number != nil:
		return builtins[name].number(body, args)
	case isTransform:
		return unproven(fmt.Sprintf("{%s} rewrites text rather than printing a value; write the values it would print", body))
	}
	return printing(body, DataTypeString, proven{notOperand: fmt.Sprintf("{%s} prints text, not a number", body)})
}

func (p *valueProof) calc(t *template, body string, args []string) proven {
	expr, err := parseCalc(args[0])
	if err != nil {
		panic(fmt.Sprintf("fejkdata: calc(%q) reached a proof unparsed: %v", args[0], err))
	}
	v, doubt := p.expr(expr, t.fields)
	if doubt == "" && !(magnitude(v) <= calcLimit) {
		doubt = calcText(expr) + " is not proven within 1e300"
	}
	if doubt != "" {
		return unproven(fmt.Sprintf("{%s}: %s", body, doubt))
	}
	return printedNumber(body, v, calcDecimals(args))
}

// calcLimit is the largest magnitude a proof accepts as finite, far enough below
// math.MaxFloat64 that rounding in the bounds cannot hide an overflow.
const calcLimit = 1e300

// expr bounds a calc expression from its operands, or says why it cannot.
func (p *valueProof) expr(n calcNode, fields map[string]node) (proven, string) {
	switch n := n.(type) {
	case calcNum:
		v := float64(n)
		return bounded(v, v, v == math.Trunc(v)), ""
	case calcVar:
		v := p.of(fields[string(n)])
		if v.notOperand != "" {
			return proven{}, fmt.Sprintf("operand %q: %s", string(n), v.notOperand)
		}
		return proven{lo: v.lo, hi: v.hi, nonZero: v.nonZero, integral: v.integral}, ""
	case calcNeg:
		v, doubt := p.expr(n.x, fields)
		v.lo, v.hi = -v.hi, -v.lo
		return v, doubt
	case calcBin:
		l, doubt := p.expr(n.l, fields)
		if doubt != "" {
			return l, doubt
		}
		r, doubt := p.expr(n.r, fields)
		if doubt != "" {
			return r, doubt
		}
		return combine(n, l, r)
	}
	panic(fmt.Sprintf("fejkdata: calc node %T has no bound", n))
}

// combine bounds one operation from the bounds of its sides.
func combine(n calcBin, l, r proven) (proven, string) {
	var v proven
	integral := l.integral && r.integral
	switch n.op {
	case '+':
		v = bounded(l.lo+r.lo, l.hi+r.hi, integral)
	case '-':
		v = bounded(l.lo-r.hi, l.hi-r.lo, integral)
	case '*':
		v = bounded(min(l.lo*r.lo, l.lo*r.hi, l.hi*r.lo, l.hi*r.hi), max(l.lo*r.lo, l.lo*r.hi, l.hi*r.lo, l.hi*r.hi), integral)
		v.nonZero = max(v.nonZero, l.nonZero*r.nonZero)
	default:
		if r.nonZero == 0 {
			return v, fmt.Sprintf("divides by %s, which is not proven nonzero", calcText(n.r))
		}
		m := magnitude(l) / r.nonZero
		v = proven{lo: -m, hi: m, nonZero: l.nonZero / magnitude(r)}
	}
	if !(magnitude(v) <= calcLimit) {
		return v, calcText(n) + " is not proven within 1e300"
	}
	return v, ""
}

// bounded is a number in [lo, hi], its distance from zero read off the bounds.
func bounded(lo, hi float64, integral bool) proven {
	v := proven{lo: lo, hi: hi, integral: integral}
	switch {
	case lo > 0:
		v.nonZero = lo
	case hi < 0:
		v.nonZero = -hi
	}
	return v
}

func magnitude(v proven) float64 { return math.Max(math.Abs(v.lo), math.Abs(v.hi)) }

// printedNumber is what a token printing v to dp decimals holds: an integer when whole
// and within int64, else a number.
func printedNumber(token string, v proven, dp int) proven {
	if dp >= 0 {
		half, _ := strconv.ParseFloat("5e-"+strconv.Itoa(dp+1), 64) // math.Pow(10, -dp) can land below the tie and let a printed zero through
		v = proven{lo: v.lo - half, hi: v.hi + half, nonZero: math.Max(0, v.nonZero-half), integral: v.integral || dp == 0}
	}
	if dp != 0 && !(dp < 0 && v.integral) {
		return printing(token, DataTypeNumber, v)
	}
	v = printing(token, DataTypeInteger, v)
	if !(magnitude(v) < math.MaxInt64) {
		v.not[DataTypeInteger] = fmt.Sprintf("{%s} is not proven within int64", token)
	}
	return v
}

// printing is v for a token whose every render is text of datatype prints, with a reason
// against each datatype that text is not.
func printing(token string, prints DataType, v proven) proven {
	for d := DataTypeInteger; d <= DataTypeBoolean; d++ {
		if prints != d && !(prints == DataTypeInteger && d == DataTypeNumber) {
			v.not[d] = fmt.Sprintf("{%s} prints %s, not %s", token, dataTypeNouns[prints], dataTypeNouns[d])
		}
	}
	return v
}

func notOneValue(format, calls string) string {
	return fmt.Sprintf("%q is not one value; write one literal or one %s, or read one", format, calls)
}

// unproven is a render no datatype and no calc can take, for why.
func unproven(why string) proven {
	v := proven{notOperand: why}
	for d := DataTypeInteger; d <= DataTypeBoolean; d++ {
		v.not[d] = why
	}
	return v
}

var (
	integerText = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	numberText  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

// literalValue proves fixed text: the number calc reads it as, and each datatype it is.
func literalValue(text string) proven {
	var v proven
	if f, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		v.notOperand = fmt.Sprintf("%q is not a number", text)
	} else {
		v = bounded(f, f, f == math.Trunc(f))
	}
	if _, err := strconv.ParseInt(text, 10, 64); !integerText.MatchString(text) {
		v.not[DataTypeInteger] = fmt.Sprintf("%q is not an integer", text)
	} else if err != nil {
		v.not[DataTypeInteger] = fmt.Sprintf("%q is past the int64 range", text)
	}
	if v.notOperand != "" || !numberText.MatchString(text) {
		v.not[DataTypeNumber] = fmt.Sprintf("%q is not a number", text)
	}
	if text != "true" && text != "false" {
		v.not[DataTypeBoolean] = fmt.Sprintf("%q is not a boolean", text)
	}
	return signedZero(text, v)
}

// signedZero refuses a zero written with a sign as a typed value, naming it unsigned.
func signedZero(text string, v proven) proven {
	if !strings.HasPrefix(text, "-") || v.notOperand != "" || v.lo != 0 {
		return v
	}
	for _, d := range []DataType{DataTypeInteger, DataTypeNumber} {
		if v.not[d] == "" {
			v.not[d] = fmt.Sprintf("%q is zero written with a sign; write %q", text, text[1:])
		}
	}
	return v
}
