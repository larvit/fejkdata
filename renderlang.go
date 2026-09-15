package fejkdata

import (
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// grammar is a deterministic automaton over a scalar's text: state 0 is dead, 1 the
// start, and each state lists the runes that leave it and where they lead.
type grammar [][]arc

type arc struct {
	on string
	to int
}

func (g *grammar) run(q int, s string) int {
	for _, r := range s {
		if q = g.step(q, r); q == 0 {
			return 0
		}
	}
	return q
}

func (g *grammar) step(q int, r rune) int {
	for _, a := range (*g)[q] {
		if strings.ContainsRune(a.on, r) {
			return a.to
		}
	}
	return 0
}

const (
	decimalDigits = "0123456789"
	nonZeroDigits = "123456789"
)

// numberGrammar reads a JSON number. States: 2 "-", 3 "0", 4 more integer digits, 5 ".",
// 6 fraction digits, 7 "e", 8 its sign, 9 exponent digits.
var numberGrammar = &grammar{
	nil,
	{{"-", 2}, {"0", 3}, {nonZeroDigits, 4}},
	{{"0", 3}, {nonZeroDigits, 4}},
	{{".", 5}, {"eE", 7}},
	{{decimalDigits, 4}, {".", 5}, {"eE", 7}},
	{{decimalDigits, 6}},
	{{decimalDigits, 6}, {"eE", 7}},
	{{"+-", 8}, {decimalDigits, 9}},
	{{decimalDigits, 9}},
	{{decimalDigits, 9}},
}

const (
	integerAccept uint32 = 1<<3 | 1<<4
	numberAccept         = integerAccept | 1<<6 | 1<<9
)

var booleanGrammar = &grammar{
	nil,
	{{"t", 2}, {"f", 6}},
	{{"r", 3}}, {{"u", 4}}, {{"e", 5}}, nil,
	{{"a", 7}}, {{"l", 8}}, {{"s", 9}}, {{"e", 10}}, nil,
}

const booleanAccept uint32 = 1<<5 | 1<<10

// decimalGrammar reads what a calc operand must render to be proven finite: a sign,
// digits and at most one dot. Past the sign, states 4–9 are positive and 10–15 their
// negatives: 4 zero digits, 5 a nonzero integer, 6 a leading dot, 7 zero with a dot,
// 8 a nonzero integer with a zero fraction, 9 a nonzero fraction.
var decimalGrammar = &grammar{
	nil,
	{{"+", 2}, {"-", 3}, {"0", 4}, {nonZeroDigits, 5}, {".", 6}},
	{{"0", 4}, {nonZeroDigits, 5}, {".", 6}},
	{{"0", 10}, {nonZeroDigits, 11}, {".", 12}},
	{{"0", 4}, {nonZeroDigits, 5}, {".", 7}},
	{{decimalDigits, 5}, {".", 8}},
	{{"0", 7}, {nonZeroDigits, 9}},
	{{"0", 7}, {nonZeroDigits, 9}},
	{{"0", 8}, {nonZeroDigits, 9}},
	{{decimalDigits, 9}},
	{{"0", 10}, {nonZeroDigits, 11}, {".", 13}},
	{{decimalDigits, 11}, {".", 14}},
	{{"0", 13}, {nonZeroDigits, 15}},
	{{"0", 13}, {nonZeroDigits, 15}},
	{{"0", 14}, {nonZeroDigits, 15}},
	{{decimalDigits, 15}},
}

const (
	decimalAccept     uint32 = 1<<4 | 1<<5 | 1<<7 | 1<<8 | 1<<9 | 1<<10 | 1<<11 | 1<<13 | 1<<14 | 1<<15
	decimalNegative   uint32 = 0xfc00
	decimalZero       uint32 = 1<<4 | 1<<7 | 1<<10 | 1<<13
	decimalFractional uint32 = 1<<9 | 1<<15
)

// relation is what a node's renders do to a grammar: from each state, the states a
// render can end in, and one render reaching each.
type relation struct {
	g  *grammar
	to []uint32
	w  []witness // w[from*len(to)+to]
}

// witness is one render, cut past witnessCap bytes, and why it can occur when the text
// alone does not say.
type witness struct {
	text string
	cut  bool
	why  string
}

const witnessCap = 60

func (w witness) then(next witness) witness {
	if w.why == "" {
		w.why = next.why
	}
	if w.cut {
		return w
	}
	w.text += next.text
	w.cut = next.cut
	if len(w.text) > witnessCap {
		end := witnessCap
		for !utf8.RuneStart(w.text[end]) {
			end--
		}
		w.text, w.cut = w.text[:end], true
	}
	return w
}

func (w witness) String() string {
	if w.cut {
		return strconv.Quote(w.text + "…")
	}
	return strconv.Quote(w.text)
}

func newRelation(g *grammar) *relation {
	n := len(*g)
	return &relation{g: g, to: make([]uint32, n), w: make([]witness, n*n)}
}

func (r *relation) add(from, to int, w witness) {
	if r.to[from]&(1<<to) == 0 {
		r.to[from] |= 1 << to
		r.w[from*len(r.to)+to] = w
	}
}

// textRelation is the relation of a render that is always s.
func textRelation(g *grammar, s, why string) *relation {
	r := newRelation(g)
	w := witness{why: why}.then(witness{text: s})
	for q := range r.to {
		r.add(q, g.run(q, s), w)
	}
	return r
}

// union is the renders of either relation; a nil relation has none.
func union(a, b *relation) *relation {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	u := newRelation(a.g)
	for _, r := range []*relation{a, b} {
		for from, ends := range r.to {
			for to := range r.to {
				if ends&(1<<to) != 0 {
					u.add(from, to, r.w[from*len(r.to)+to])
				}
			}
		}
	}
	return u
}

// then is a render of r followed by a render of next.
func (r *relation) then(next *relation) *relation {
	c := newRelation(r.g)
	n := len(r.to)
	for from, mids := range r.to {
		for mid := 0; mid < n; mid++ {
			if mids&(1<<mid) == 0 {
				continue
			}
			for to := 0; to < n; to++ {
				if next.to[mid]&(1<<to) != 0 && c.to[from]&(1<<to) == 0 {
					c.add(from, to, r.w[from*n+mid].then(next.w[mid*n+to]))
				}
			}
		}
	}
	return c
}

// power is k renders of r in a row, k at least 1, composed by squaring.
func (r *relation) power(k int) *relation {
	var out *relation
	for base := r; ; base = base.then(base) {
		if k&1 == 1 {
			if out == nil {
				out = base
			} else {
				out = out.then(base)
			}
		}
		if k >>= 1; k == 0 {
			return out
		}
	}
}

// closure is any number of renders of r in a row, where r includes the empty render.
func (r *relation) closure() *relation {
	for {
		next := r.then(r)
		if slices.Equal(next.to, r.to) {
			return r
		}
		r = next
	}
}

// escape finds a render from the start that ends outside accept, preferring one that
// carries a reason.
func (r *relation) escape(accept uint32) (witness, bool) {
	var found witness
	escapes := false
	for to := range r.to {
		if (r.to[1]&^accept)&(1<<to) == 0 {
			continue
		}
		if w := r.w[len(r.to)+to]; !escapes || found.why == "" && w.why != "" {
			found, escapes = w, true
		}
	}
	return found, escapes
}

// textShape is the text a builtin can emit: alternatives, each a sequence of runs.
type textShape [][]charRun

// charRun is between min and max characters, each one of chars; max -1 is unbounded.
// chars is ASCII, so a run of k characters is k bytes.
type charRun struct {
	chars    string
	min, max int
}

// textLanguage is what one grammar makes of the renders a check reads, worked out once
// per node and fold.
type textLanguage struct {
	g     *grammar
	proof *calcProof
	memo  map[languageKey]*relation
	empty *relation
}

type languageKey struct {
	n    node
	fold string
}

// fold is the transforms a render passes through before the grammar reads it, innermost
// first. Each rewrites rune by rune, so folding a render is folding each of its pieces.
type fold []string

func (f fold) apply(s string) string {
	for _, name := range f {
		s = transforms[name](s)
	}
	return s
}

func newTextLanguage(g *grammar, proof *calcProof) *textLanguage {
	return &textLanguage{g: g, proof: proof, memo: map[languageKey]*relation{}, empty: textRelation(g, "", "")}
}

func (l *textLanguage) node(n node, f fold) *relation {
	key := languageKey{n, strings.Join(f, ",")}
	if r, done := l.memo[key]; done {
		return r
	}
	r := l.empty // a null renders ""
	switch n := n.(type) {
	case *choice:
		r = nil
		for _, it := range n.items {
			r = union(r, l.node(it, f))
		}
	case *template:
		r = l.format(n, f)
		if n.repeat > 1 {
			r = r.then(l.text(f.apply(n.separator)).then(r).power(n.repeat - 1))
		}
	}
	l.memo[key] = r
	return r
}

func (l *textLanguage) text(s string) *relation { return textRelation(l.g, s, "") }

// format reads a template's format the way expand renders it: literal runs and tokens
// in turn.
func (l *textLanguage) format(t *template, f fold) *relation {
	r := l.empty
	var lit strings.Builder
	_ = eachToken(t.format, func(tok ftoken) error {
		if tok.kind == 'l' {
			lit.WriteRune(tok.r)
			return nil
		}
		r = r.then(l.text(f.apply(lit.String()))).then(l.token(t, tok.body, f))
		lit.Reset()
		return nil
	})
	return r.then(l.text(f.apply(lit.String())))
}

// token reads one {…} token: a field read, a transform over one, a calc, or what a
// builtin emits.
func (l *textLanguage) token(t *template, body string, f fold) *relation {
	name, args, isFunc := funcCall(body)
	if !isFunc {
		var r *relation
		for _, a := range splitArms(body, t.refs) {
			r = union(r, l.read(t, a, f))
		}
		return r
	}
	if _, isTransform := transforms[name]; isTransform {
		leaf, chain, _ := unwrapTransform(args[0])
		inner := slices.Clone(chain)
		slices.Reverse(inner)
		return l.read(t, splitArm(leaf, t.refs), append(append(inner, name), f...))
	}
	if name == "calc" {
		return l.calc(t, args, f)
	}
	return l.shape(builtins[name].emits(args), f)
}

// read is one arm of a token: every node its path can land on.
func (l *textLanguage) read(t *template, a arm, f fold) *relation {
	var r *relation
	for _, leaf := range pathLeaves(t.fields[a.key], a.tail) {
		r = union(r, l.node(leaf, f))
	}
	return r
}

func (l *textLanguage) calc(t *template, args []string, f fold) *relation {
	b, d := l.proof.call(t, args)
	if d != nil {
		return textRelation(l.g, f.apply(d.render), d.why)
	}
	return l.shape(printedFloat(b.lo, b.hi, calcDecimals(args), b.integral), f)
}

func (l *textLanguage) shape(s textShape, f fold) *relation {
	var r *relation
	for _, alt := range s {
		seq := l.empty
		for _, run := range alt {
			seq = seq.then(l.run(run, f))
		}
		r = union(r, seq)
	}
	return r
}

// run reads a charRun: min characters, then up to max-min more.
func (l *textLanguage) run(c charRun, f fold) *relation {
	one := newRelation(l.g)
	for from := range one.to {
		for _, ch := range c.chars {
			s := f.apply(string(ch))
			one.add(from, l.g.run(from, s), witness{text: s})
		}
	}
	more := l.empty
	switch optional := union(one, l.empty); {
	case c.max < 0:
		more = optional.closure()
	case c.max > c.min:
		more = optional.power(c.max - c.min)
	}
	if c.min == 0 {
		return more
	}
	return one.power(c.min).then(more)
}
