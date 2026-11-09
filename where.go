package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// maxDrawValues caps the values one where or distinct clause weighs at load.
const maxDrawValues = 1 << 20

// valueDraw draws among the values a path renders, fixed at load with the odds each has in a draw of
// the path: one of the values that pass its where clause, or count distinct ones where a distinct
// clause asks. A name the clause reads is read at each draw.
type valueDraw struct {
	lengthIndex
	lo, hi  lengthEnd
	reads   []arm  // the op's operands: the names the clause reads, which lo, hi, strings and fallback index
	filters bool   // a where clause filters the values
	count   int    // how many distinct values one draw takes
	sep     string // what prints between them
	strings readStrings
	// fallback prints in place of each value a draw can't take, once a string read at render left
	// too few
	fallback fallback
}

// lengthIndex holds values by length, each with its odds.
type lengthIndex struct {
	texts   []string  // shortest first
	lengths []int     // each text's length in code points
	cum     []float64 // running sum of the texts' odds
}

// lengthEnd is one end of a where clause's length: n, or the name reads[read] where read is 0 or more.
type lengthEnd struct {
	n    int
	read int
}

// readStrings are the reads, by index into a draw's reads, of the strings a value must not equal,
// and of those it must not contain.
type readStrings struct{ not, notContaining []int }

// any says whether r reads a string, which alone can leave a draw too few values.
func (r readStrings) any() bool { return len(r.not)+len(r.notContaining) > 0 }

// fallback is what a draw prints where no value passes: text, or the name reads[read] where read is
// 0 or more.
type fallback struct {
	text string
	read int
}

func (f fallback) at(s *drawstate.State, env renderEnv, reads []arm) string {
	if f.read < 0 {
		return f.text
	}
	return readName(s, env, reads[f.read]).text
}

// clauseName is a name a where clause reads, and the word of the clause reading it.
type clauseName struct{ word, name string }

// clauseNames is every name w reads, each once: its length's first, then not's, not containing's and
// else's.
func clauseNames(w *grammar.Where) []clauseName {
	if w == nil {
		return nil
	}
	var names []clauseName
	add := func(word string, ns ...string) {
		for _, n := range ns {
			if n != "" && readOf(names, n) < 0 {
				names = append(names, clauseName{word, n})
			}
		}
	}
	if w.Length != nil {
		add("length", w.Length.Lo.Name, w.Length.Hi.Name)
	}
	add("not", w.NotNames...)
	add("not containing", w.NotContainingNames...)
	if w.Else != nil {
		add("else", w.Else.Name)
	}
	return names
}

// readOf is the index of name among names, which is the index of its read among the op's operands.
func readOf(names []clauseName, name string) int {
	return slices.IndexFunc(names, func(c clauseName) bool { return c.name == name })
}

func clauseReads(names []clauseName, of []string) []int {
	var reads []int
	for _, n := range of {
		reads = append(reads, readOf(names, n))
	}
	return reads
}

// fallbackOf is what a draw under w prints where no value passes, deflt where w names no fallback.
func fallbackOf(w *grammar.Where, names []clauseName, deflt string) fallback {
	switch {
	case w == nil || w.Else == nil:
		return fallback{text: deflt, read: -1}
	case w.Else.Name != "":
		return fallback{read: readOf(names, w.Else.Name)}
	}
	return fallback{text: w.Else.Text, read: -1}
}

// unbound is the refusal of c where no field and no name holds it.
func (c clauseName) unbound() error {
	head := splitArm(c.name, nil).head
	if c.word == "length" {
		return fmt.Errorf("length %s reads a name, and no name %q is bound; bind one first, as {… as %s}", c.name, head, head)
	}
	return fmt.Errorf("%s %s reads a name, and no name %q is bound; bind one first, as {… as %s}, or quote the text, as '%s'", c.word, c.name, head, head, c.name)
}

func (c clauseName) listsStrings() bool { return c.word == "not" || c.word == "not containing" }

// fieldRead is the refusal of c reading a field, which a clause reads only through a name.
func (c clauseName) fieldRead() error {
	bind := fmt.Sprintf("bind it first, as {%s as n}, and write %s n", c.name, c.word)
	if c.listsStrings() {
		return fmt.Errorf("%s %s reads a name, and %q is a field; for one of its values, %s; for every value, quote each, as %s '…', '…'", c.word, c.name, c.name, bind, c.word)
	}
	return fmt.Errorf("%s %s reads a name, and %q is a field; %s", c.word, c.name, c.name, bind)
}

// membersRead is the refusal of c reading whole a name holding n distinct members, which a list of
// strings names one by one.
func (c clauseName) membersRead(n int) error {
	var members []string
	for k := 1; k <= n; k++ {
		switch {
		case k <= 2 || k == n:
			members = append(members, fmt.Sprintf("%s[%d]", c.name, k))
		case k == 3:
			members = append(members, "…")
		}
	}
	return fmt.Errorf("%s %s reads name %q, which holds %d distinct members; list them, as %s %s", c.word, c.name, c.name, n, c.word, strings.Join(members, ", "))
}

// compileClauseNames compiles the names w reads, each bound in a template enclosing t.
func (t *template) compileClauseNames(w *grammar.Where, targets map[*nameBinding]nameTarget) ([]arm, error) {
	var arms []arm
	for _, c := range clauseNames(w) {
		read := splitArm(c.name, nil)
		if !t.isName(read.head) {
			return nil, c.fieldRead()
		}
		if b := t.nameScope.lookup(read.head); b.distinct != nil && len(read.tail) == 0 && c.listsStrings() {
			return nil, c.membersRead(b.distinct.N)
		}
		a, err := t.compileArm(c.name, targets)
		if err != nil {
			return nil, err
		}
		arms = append(arms, a)
	}
	return arms, nil
}

// valueDraws builds the draw of each token in ts with a where or a distinct clause, and gives each
// name bound to such a token that token's draw.
func valueDraws(ts []templateSite, targets map[*nameBinding]nameTarget) error {
	for _, s := range ts {
		for i := range s.t.compiled.ops {
			o := &s.t.compiled.ops[i]
			if o.Where == nil && o.Distinct == nil {
				continue
			}
			if o.Kind == grammar.BuiltinCall {
				var err error
				if o.sample, err = newSampleDraw(o); err != nil {
					return fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
				}
				continue
			}
			v, err := s.t.valueDrawOf(o, targets)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
			}
			o.values = v
			if o.Kind == grammar.NameBind {
				s.t.nameScope.bindings[o.Bound].values = v
			}
		}
	}
	if err := refuseSelfReads(ts); err != nil {
		return err
	}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			if r.a.named.values != nil && len(r.a.tail) > 0 {
				return fmt.Errorf("%s: token {%s}: name %q holds a filtered draw, which is read whole: {%s}", s.label, r.o.Body, r.a.head, r.a.head)
			}
		}
	}
	return nil
}

// refuseSelfReads refuses a token whose where clause reads the name the token binds, directly or
// through what the names it reads draw: drawing the name would read itself.
func refuseSelfReads(ts []templateSite) error {
	for _, s := range ts {
		for _, o := range s.t.compiled.ops {
			if o.Kind != grammar.NameBind || o.values == nil {
				continue
			}
			b := s.t.nameScope.bindings[o.Bound]
			r := selfRead{b: b, names: map[*nameBinding]bool{}, nodes: map[node]bool{}}
			for _, a := range o.values.reads {
				if !r.reaches(a) {
					continue
				}
				via := ""
				if a.named != b {
					via = fmt.Sprintf(", and drawing %q reads %q", a.head, b.name)
				}
				return fmt.Errorf("%s: token {%s}: its where clause reads %q%s, the name it binds, so the draw would read itself; read the clause's strings from another name", s.label, o.Body, a.head, via)
			}
		}
	}
	return nil
}

// selfRead walks what reading an arm draws, looking for a read of b.
type selfRead struct {
	b     *nameBinding
	names map[*nameBinding]bool
	nodes map[node]bool
}

// reaches says whether reading a reads b: a name holding a filtered draw reads only its clause's
// names, since its values are fixed at load; any other read renders its leaves.
func (r selfRead) reaches(a arm) bool {
	if a.kind == namedRead {
		if a.named == r.b {
			return true
		}
		if r.names[a.named] {
			return false
		}
		r.names[a.named] = true
		if a.named.values != nil {
			return slices.ContainsFunc(a.named.values.reads, r.reaches)
		}
	}
	return slices.ContainsFunc(a.leaves, r.reachesNode)
}

func (r selfRead) reachesNode(n node) bool {
	if r.nodes[n] {
		return false
	}
	r.nodes[n] = true
	t, ok := n.(*template)
	if !ok {
		return slices.ContainsFunc(renderEdges(n), func(e renderEdge) bool { return r.reachesNode(e.to) })
	}
	for _, o := range t.compiled.ops {
		if slices.ContainsFunc(slices.Concat(o.operands, o.arms), r.reaches) {
			return true
		}
	}
	return false
}

// valueDrawOf builds o's draw over the values o's path renders, and proves every draw leaves
// enough to draw.
func (t *template) valueDrawOf(o *op, targets map[*nameBinding]nameTarget) (*valueDraw, error) {
	clause := "a where clause filters"
	if o.Where == nil {
		clause = "a distinct clause draws"
	}
	var start node
	var tail []string
	if o.Kind == grammar.PathRead {
		a := o.arms[0]
		if a.kind == namedRead {
			return nil, fmt.Errorf("%q is a name, drawn once; write the clause on the draw it binds: {%s … as %s}", a.head, a.named.ref, a.head)
		}
		start, tail = t.startOf(a.head), a.tail
	} else {
		target := targets[t.nameScope.bindings[o.Bound]]
		start, tail = target.start, target.tail
	}
	w := valueWalk{targets: targets, seen: map[node]bool{}}
	if err := w.walk(start, tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("%s values fixed at load, and %w", clause, err)
	}
	v := newValueDraw(w.out, o.Where, o.Distinct, o.operands)
	return v, v.lengthProof().proveEveryDrawPasses()
}

type weightedText struct {
	text   string
	weight float64
}

// valueWalk lists every value a path renders with its odds, refusing a path whose values are not
// fixed at load.
type valueWalk struct {
	targets map[*nameBinding]nameTarget
	seen    map[node]bool // the nodes the walk is inside
	out     []weightedText
}

// walk lists the values of the path tail into n, each with weight times its odds, written
// between pre and suf.
func (w *valueWalk) walk(n node, tail []string, weight float64, pre, suf string) error {
	if w.seen[n] {
		return errors.New("the path reads itself")
	}
	w.seen[n] = true
	defer delete(w.seen, n)
	switch n := n.(type) {
	case *choice:
		return w.walkChoice(n, tail, weight, pre, suf)
	case *template:
		if len(tail) > 0 {
			return w.walk(n.fields[tail[0]], tail[1:], weight, pre, suf)
		}
		return w.walkTemplate(n, weight, pre, suf)
	case *table:
		return w.walkTable(n, tail, weight, pre, suf)
	case *tableColumn:
		return w.walkRows(n.t, []rowPart{{column: n}}, weight, pre, suf)
	case *nullItem:
		return errors.New("the path may draw a null")
	}
	panic(invariant.Broken("a value walk reaches %T", n))
}

func (w *valueWalk) walkChoice(c *choice, tail []string, weight float64, pre, suf string) error {
	total, prev := float64(len(c.items)), 0.0
	if c.cum != nil {
		total = c.cum[len(c.cum)-1]
	}
	for i, item := range c.items {
		odds := 1.0
		if c.cum != nil {
			odds, prev = c.cum[i]-prev, c.cum[i]
		}
		if err := w.walk(item, tail, weight*odds/total, pre, suf); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) walkTemplate(t *template, weight float64, pre, suf string) error {
	if t.repeat != 1 {
		return fmt.Errorf("%q repeats", t.format)
	}
	if lit, fixed := t.fixedText(); fixed {
		return w.emit(pre+lit+suf, weight)
	}
	before, read, after, err := oneRead(t)
	if err != nil {
		return err
	}
	if read == nil {
		return w.emit(pre+before+after+suf, weight)
	}
	for _, a := range read.arms {
		if err := w.walkArm(t, a, weight/float64(len(read.arms)), pre+before, after+suf); err != nil {
			return err
		}
	}
	return nil
}

// oneRead splits t's format around its one read, refusing a format that draws anything else.
func oneRead(t *template) (before string, read *op, after string, err error) {
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		switch {
		case o.Kind == grammar.LiteralRun && read == nil:
			before += o.Lit
		case o.Kind == grammar.LiteralRun:
			after += o.Lit
		case o.Kind == grammar.BuiltinCall:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws anew each time", t.format, o.Body)
		case o.Where != nil || o.Distinct != nil:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws through a clause of its own", t.format, o.Body)
		case read != nil:
			return "", nil, "", fmt.Errorf("%q reads both {%s} and {%s}", t.format, read.Body, o.Body)
		default:
			read = o
		}
	}
	return before, read, after, nil
}

// walkArm walks one arm of t's read: a path from a field or a reference, or through a name t
// binds itself; that name's one read draws as a fresh read would.
func (w *valueWalk) walkArm(t *template, a arm, weight float64, pre, suf string) error {
	if a.kind != namedRead {
		return w.walk(t.startOf(a.head), a.tail, weight, pre, suf)
	}
	if a.named.binder != t {
		return fmt.Errorf("%q reads name %q, bound outside it, which keeps one draw for a whole render", t.format, a.head)
	}
	target, resolved := w.targets[a.named]
	if !resolved {
		target = a.named.target
	}
	return w.walk(target.start, slices.Concat(target.tail, a.tail), weight, pre, suf)
}

func (w *valueWalk) walkTable(t *table, tail []string, weight float64, pre, suf string) error {
	if len(tail) == 0 {
		if parts, own := ownColumnReads(t); own {
			return w.walkRows(t, parts, weight, pre, suf)
		}
		return w.walk(t.formatTemplate, nil, weight, pre, suf)
	}
	if c, isColumn := t.formatTemplate.fields[tail[0]].(*tableColumn); isColumn && len(tail) == 1 {
		return w.walk(c, nil, weight, pre, suf)
	}
	return fmt.Errorf("the path runs through a selector, a linked table or \"..\": %s of %s", grammar.JoinSegments(tail), t.rows.Segment())
}

// rowPart is one piece of a row's text: fixed text, or the row's cell in column.
type rowPart struct {
	lit    string
	column *tableColumn
}

// ownColumnReads is t's format as fixed text and plain reads of t's own columns; own is false where
// the format reads anything else.
func ownColumnReads(t *table) (parts []rowPart, own bool) {
	for _, o := range t.formatTemplate.compiled.ops {
		if o.Kind == grammar.LiteralRun {
			parts = append(parts, rowPart{lit: o.Lit})
			continue
		}
		if o.Kind != grammar.PathRead || o.Where != nil || o.Distinct != nil || len(o.arms) != 1 {
			return nil, false
		}
		a := o.arms[0]
		c, isColumn := t.formatTemplate.fields[a.head].(*tableColumn)
		if !isColumn || a.kind != freshRead || len(a.tail) > 0 {
			return nil, false
		}
		parts = append(parts, rowPart{column: c})
	}
	return parts, true
}

// walkRows lists each row of t as parts spell it, weighed by the row, as a draw over the whole table
// does.
func (w *valueWalk) walkRows(t *table, parts []rowPart, weight float64, pre, suf string) error {
	total := 0.0
	for r := 0; r < t.rows.Len(); r++ {
		total += t.rows.Weight(r)
	}
	for r := 0; r < t.rows.Len(); r++ {
		text := pre
		for _, p := range parts {
			if p.column == nil {
				text += p.lit
				continue
			}
			if cell := t.cellTemplate(r, p.column.i); cell != nil {
				return fmt.Errorf("column %s of %s holds %q, which is not fixed text", p.column.name(), t.rows.Segment(), cell.format)
			}
			text += t.rows.Cell(r, p.column.i)
		}
		if err := w.emit(text+suf, weight*t.rows.Weight(r)/total); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) emit(text string, weight float64) error {
	if weight == 0 {
		return nil
	}
	if len(w.out) == maxDrawValues {
		return fmt.Errorf("the path renders more than %d values", maxDrawValues)
	}
	w.out = append(w.out, weightedText{text, weight})
	return nil
}

// newValueDraw keeps the values that pass where's fixed conditions, shortest first, for draws of
// distinct's count, or of one value where distinct is nil.
func newValueDraw(values []weightedText, where *grammar.Where, distinct *grammar.Distinct, reads []arm) *valueDraw {
	names := clauseNames(where)
	v := &valueDraw{reads: reads, count: 1, fallback: fallbackOf(where, names, "")}
	v.lo, v.hi = lengthEnds(where, names)
	if distinct != nil {
		v.count, v.sep = distinct.N, distinct.Sep
	}
	if where == nil {
		v.lengthIndex = newLengthIndex(values)
		return v
	}
	var kept []weightedText
	containing := nonEmpty(where.NotContaining)
	for _, value := range values {
		if !slices.Contains(where.Not, value.text) && !holdsAny(value.text, containing) {
			kept = append(kept, value)
		}
	}
	v.lengthIndex, v.filters = newLengthIndex(kept), true
	v.strings = readStrings{clauseReads(names, where.NotNames), clauseReads(names, where.NotContainingNames)}
	return v
}

// lengthEnds are the ends of w's length, 0 to every length where w has none.
func lengthEnds(w *grammar.Where, names []clauseName) (lo, hi lengthEnd) {
	if w == nil || w.Length == nil {
		return lengthEnd{n: 0, read: -1}, lengthEnd{n: math.MaxInt, read: -1}
	}
	return endOf(w.Length.Lo, names), endOf(w.Length.Hi, names)
}

func endOf(l grammar.Length, names []clauseName) lengthEnd {
	if l.Name == "" {
		return lengthEnd{n: l.N, read: -1}
	}
	return lengthEnd{read: readOf(names, l.Name)}
}

func newLengthIndex(values []weightedText) lengthIndex {
	type sized struct {
		weightedText
		length int
	}
	vs := make([]sized, len(values))
	for i, v := range values {
		vs[i] = sized{v, utf8.RuneCountInString(v.text)}
	}
	sort.SliceStable(vs, func(i, j int) bool { return vs[i].length < vs[j].length })
	x := lengthIndex{texts: make([]string, len(vs)), lengths: make([]int, len(vs)), cum: make([]float64, len(vs))}
	total := 0.0
	for i, v := range vs {
		total += v.weight
		x.texts[i], x.lengths[i], x.cum[i] = v.text, v.length, total
	}
	return x
}

// span bounds the texts in x.texts[from:to], those whose length is lo to hi.
func (x *lengthIndex) span(lo, hi int) (from, to int) { return spanOf(x.lengths, lo, hi) }

func spanOf(lengths []int, lo, hi int) (from, to int) {
	from = sort.Search(len(lengths), func(i int) bool { return lengths[i] >= lo })
	to = sort.Search(len(lengths), func(i int) bool { return lengths[i] > hi })
	return from, to
}

// drawIn draws one of the texts in x.texts[from:to], by their odds.
func (x *lengthIndex) drawIn(s *drawstate.State, from, to int) string {
	return x.texts[x.drawIndex(s, from, to, nil)]
}

// drawIndex draws the index of one of the texts in x.texts[from:to] by their odds, leaving out
// the ascending indexes drawn. It costs drawn's length plus a binary search over the span: it draws
// among the odds left, and steps past each drawn text the draw lands at or beyond.
func (x *lengthIndex) drawIndex(s *drawstate.State, from, to int, drawn []int) int {
	base, left := x.before(from), x.cum[to-1]-x.before(from)
	for _, d := range drawn {
		left -= x.cum[d] - x.before(d)
	}
	v := base + s.Float64()*left
	for _, d := range drawn {
		if x.before(d) > v {
			break
		}
		v += x.cum[d] - x.before(d)
	}
	return undrawnNear(from+sort.Search(to-from, func(i int) bool { return x.cum[from+i] > v }), from, to, drawn)
}

// before is the running sum of the odds of the texts before index i.
func (x *lengthIndex) before(i int) float64 {
	if i == 0 {
		return 0
	}
	return x.cum[i-1]
}

// undrawnNear is i, or where rounding landed i on a drawn index or on to, the first undrawn index
// after it, else the last one before it, within from:to.
func undrawnNear(i, from, to int, drawn []int) int {
	for j := i; j < to; j++ {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	for j := i - 1; j >= from; j-- {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	panic(invariant.Broken("a draw ran out of the values the load proved enough"))
}

// lengthProof is what proves a draw's length: the lengths of its values, shortest first, the
// clause's ends, the reads the ends index, and how many values one draw takes.
type lengthProof struct {
	lengths []int
	lo, hi  lengthEnd
	reads   []arm
	count   int
	filters bool // a where clause filters the values
}

func (v *valueDraw) lengthProof() lengthProof {
	return lengthProof{v.lengths, v.lo, v.hi, v.reads, v.count, v.filters}
}

// proveEveryDrawPasses refuses a draw that some render leaves with too few values: one where too few
// values pass, or one whose length reads a name that may print a length too few values have.
func (v lengthProof) proveEveryDrawPasses() error {
	bounds := make([][2]float64, len(v.reads))
	for _, e := range []lengthEnd{v.lo, v.hi} {
		if e.read < 0 {
			continue
		}
		a := v.reads[e.read]
		p := (&valueProof{}).proveArm(a)
		switch {
		case p.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", a.spelling, p.NotOperand)
		case !p.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", a.spelling)
		}
		bounds[e.read] = [2]float64{p.Lo, p.Hi}
	}
	if v.lo.read >= 0 && v.lo.read == v.hi.read {
		for n := bounds[v.lo.read][0]; n <= bounds[v.lo.read][1]; n++ {
			if passing := v.passing(n, n); passing < v.count {
				return fmt.Errorf("where %s is %.0f, %s", v.reads[v.lo.read].spelling, n, v.shortfall(passing))
			}
		}
		return nil
	}
	lo, hi := float64(v.lo.n), float64(v.hi.n)
	if v.lo.read >= 0 {
		lo = bounds[v.lo.read][1]
	}
	if v.hi.read >= 0 {
		hi = bounds[v.hi.read][0]
	}
	if passing := v.passing(lo, hi); passing < v.count {
		return v.rangeShortfall(lo, hi, passing)
	}
	return nil
}

// rangeShortfall says that the length lo to hi leaves passing values, too few, naming the length's
// names at the values that narrow it to lo to hi.
func (v lengthProof) rangeShortfall(lo, hi float64, passing int) error {
	var at []string
	if v.lo.read >= 0 {
		at = append(at, fmt.Sprintf("%s is %.0f", v.reads[v.lo.read].spelling, lo))
	}
	if v.hi.read >= 0 {
		at = append(at, fmt.Sprintf("%s is %.0f", v.reads[v.hi.read].spelling, hi))
	}
	switch {
	case len(at) == 0:
		return errors.New(v.shortfall(passing))
	case lo <= hi:
		return fmt.Errorf("where %s, %s", strings.Join(at, " and "), v.shortfall(passing))
	case v.hi.read < 0:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %d", at[0], lo, hi, v.reads[v.lo.read].spelling, v.hi.n)
	case v.lo.read < 0:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at %d or above", at[0], lo, hi, v.reads[v.hi.read].spelling, v.lo.n)
	}
	return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %s", strings.Join(at, " and "), lo, hi, v.reads[v.lo.read].spelling, v.reads[v.hi.read].spelling)
}

// shortfall says that passing values are fewer than a draw takes.
func (v lengthProof) shortfall(passing int) string {
	n := v.count
	switch {
	case !v.filters:
		return fmt.Sprintf("distinct %d draws %d values, and the path renders %d of weight above 0", n, n, passing)
	case n == 1:
		return "no value passes the where clause"
	case passing == 0:
		return fmt.Sprintf("distinct %d draws %d values, and none pass the where clause", n, n)
	case passing == 1:
		return fmt.Sprintf("distinct %d draws %d values, and only 1 passes the where clause", n, n)
	}
	return fmt.Sprintf("distinct %d draws %d values, and only %d pass the where clause", n, n, passing)
}

// passing is how many values are lo to hi long.
func (v lengthProof) passing(lo, hi float64) int {
	past := 1.0
	if len(v.lengths) > 0 {
		past += float64(v.lengths[len(v.lengths)-1])
	}
	from, to := spanOf(v.lengths, int(min(max(lo, -1), past)), int(max(min(hi, past), -1)))
	return max(to-from, 0)
}

// draw draws count distinct passing values, reading the clause's names in env. Where the strings it
// reads leave fewer values than count, the fallback prints in place of each value missing.
func (v *valueDraw) draw(s *drawstate.State, env renderEnv) []string {
	from, to := v.span(v.lo.at(s, env, v.reads), v.hi.at(s, env, v.reads))
	if to-from < v.count {
		panic(invariant.Broken("a clause the load proved leaves %d values leaves %d", v.count, to-from))
	}
	drawn := v.holding(s, env, from, to) // ascending
	out := make([]string, v.count)
	for k := range out {
		if len(drawn) == to-from {
			out[k] = v.fallback.at(s, env, v.reads)
			continue
		}
		i := v.drawIndex(s, from, to, drawn)
		at, _ := slices.BinarySearch(drawn, i)
		drawn = slices.Insert(drawn, at, i)
		out[k] = v.texts[i]
	}
	return out
}

// holding is the indexes, ascending, of the texts in v.texts[from:to] that equal a not string, or
// hold a not containing string, that v reads in env.
func (v *valueDraw) holding(s *drawstate.State, env renderEnv, from, to int) []int {
	not, containing := v.strings.read(s, env, v.reads)
	if len(not) == 0 && len(containing) == 0 {
		return nil
	}
	var out []int
	for i := from; i < to; i++ {
		if slices.Contains(not, v.texts[i]) || holdsAny(v.texts[i], containing) {
			out = append(out, i)
		}
	}
	return out
}

// read reads the strings in env: those a value must not equal, and those a value must not hold,
// leaving out an empty one, which every value holds.
func (r readStrings) read(s *drawstate.State, env renderEnv, reads []arm) (not, containing []string) {
	for _, i := range r.not {
		not = append(not, readName(s, env, reads[i]).text)
	}
	for _, i := range r.notContaining {
		if text := readName(s, env, reads[i]).text; text != "" {
			containing = append(containing, text)
		}
	}
	return not, containing
}

// nonEmpty is texts without the empty string, which every value holds.
func nonEmpty(texts []string) []string {
	return slices.DeleteFunc(slices.Clone(texts), func(t string) bool { return t == "" })
}

func holdsAny(text string, of []string) bool {
	return slices.ContainsFunc(of, func(s string) bool { return strings.Contains(text, s) })
}

func (e lengthEnd) at(s *drawstate.State, env renderEnv, reads []arm) int {
	if e.read < 0 {
		return e.n
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(readName(s, env, reads[e.read]).text), 64)
	return int(v)
}
