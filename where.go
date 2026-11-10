package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// valueDraw draws among the values a path renders, fixed at load with the odds each has in a draw of
// the path: one of the values that pass its where clause, or count distinct ones where a distinct
// clause asks. A name the clause reads is read at each draw.
type valueDraw struct {
	lengthIndex
	lo, hi  lengthEnd
	clause  []arm  // the op's clause: the names the where clause reads
	filters bool   // a where clause filters the values
	count   int    // how many distinct values one draw takes
	sep     string // what prints between them
	strings readStrings
	// fallback prints in place of each value a draw can't take, once a string read at render left
	// too few
	fallback fallback
}

// lengthEnd is one end of a where clause's length: n, or, where read is not nil, what that read of
// a name prints.
type lengthEnd struct {
	n    int
	read *arm
}

// readStrings are the reads of the strings a value must not equal, and of those it must not contain.
type readStrings struct{ not, notContaining []*arm }

// any says whether r reads a string, which alone can leave a draw too few values.
func (r readStrings) any() bool { return len(r.not)+len(r.notContaining) > 0 }

// fallback is what a draw prints where no value passes: text, or, where read is not nil, what that
// read of a name prints.
type fallback struct {
	text string
	read *arm
}

func (f fallback) at(s *drawstate.State, env renderEnv) string {
	if f.read == nil {
		return f.text
	}
	return readName(s, env, *f.read).text
}

// clauseSpelling is a spelling a where clause reads, and the word of the clause reading it.
type clauseSpelling struct{ word, spelling string }

// clauseSpellings is every spelling w reads, each once: its length's first, then not's, not containing's and
// else's.
func clauseSpellings(w *grammar.Where) []clauseSpelling {
	if w == nil {
		return nil
	}
	var names []clauseSpelling
	add := func(word string, ns ...string) {
		for _, n := range ns {
			if n != "" && !slices.ContainsFunc(names, func(c clauseSpelling) bool { return c.spelling == n }) {
				names = append(names, clauseSpelling{word, n})
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

// clauseRead returns the arm in clause, an op's compiled where clause, that reads spelling; spelling
// must be one the clause reads.
func clauseRead(clause []arm, spelling string) *arm {
	return &clause[slices.IndexFunc(clause, func(a arm) bool { return a.spelling == spelling })]
}

func clauseReads(clause []arm, of []string) []*arm {
	var reads []*arm
	for _, n := range of {
		reads = append(reads, clauseRead(clause, n))
	}
	return reads
}

// fallbackOf is what a draw under w prints where no value passes, deflt where w names no fallback.
func fallbackOf(w *grammar.Where, clause []arm, deflt string) fallback {
	switch {
	case w == nil || w.Else == nil:
		return fallback{text: deflt}
	case w.Else.Name != "":
		return fallback{read: clauseRead(clause, w.Else.Name)}
	}
	return fallback{text: w.Else.Text}
}

// unbound is the refusal of c where no field and no name holds it.
func (c clauseSpelling) unbound() error {
	head := splitRead(c.spelling, nil).head
	if c.word == "length" {
		return fmt.Errorf("length %s reads a name, and no name %q is bound; bind one first, as {… as %s}", c.spelling, head, head)
	}
	return fmt.Errorf("%s %s reads a name, and no name %q is bound; bind one first, as {… as %s}, or quote the text, as '%s'", c.word, c.spelling, head, head, c.spelling)
}

func (c clauseSpelling) listsStrings() bool { return c.word == "not" || c.word == "not containing" }

// fieldRead is the refusal of c reading a field, which a clause reads only through a name.
func (c clauseSpelling) fieldRead() error {
	bind := fmt.Sprintf("bind it first, as {%s as n}, and write %s n", c.spelling, c.word)
	if c.listsStrings() {
		return fmt.Errorf("%s %s reads a name, and %q is a field; for one of its values, %s; for every value, quote each, as %s '…', '…'", c.word, c.spelling, c.spelling, bind, c.word)
	}
	return fmt.Errorf("%s %s reads a name, and %q is a field; %s", c.word, c.spelling, c.spelling, bind)
}

// membersRead is the refusal of c reading whole a name holding n distinct members, which a list of
// strings names one by one.
func (c clauseSpelling) membersRead(n int) error {
	var members []string
	for k := 1; k <= n; k++ {
		switch {
		case k <= 2 || k == n:
			members = append(members, fmt.Sprintf("%s[%d]", c.spelling, k))
		case k == 3:
			members = append(members, "…")
		}
	}
	return fmt.Errorf("%s %s reads name %q, which holds %d distinct members; list them, as %s %s", c.word, c.spelling, c.spelling, n, c.word, strings.Join(members, ", "))
}

// compileClauseNames compiles the names w reads, each bound in a template enclosing t.
func (t *template) compileClauseNames(w *grammar.Where, l *loadFinds) ([]arm, error) {
	var arms []arm
	for _, c := range clauseSpellings(w) {
		read := splitRead(c.spelling, nil)
		if !t.isName(read.head, l) {
			return nil, c.fieldRead()
		}
		if b := l.of(t).nameScope.lookup(read.head); b.distinct != nil && len(read.tail) == 0 && c.listsStrings() {
			return nil, c.membersRead(b.distinct.N)
		}
		a, err := t.compileArm(c.spelling, l)
		if err != nil {
			return nil, err
		}
		arms = append(arms, a)
	}
	return arms, nil
}

// refuseFilteredTails refuses a path through a name holding a filtered draw, which is read whole.
func refuseFilteredTails(ts []templateSite) error {
	for _, s := range ts {
		for _, r := range namedReads(s.t, nil) {
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
			if o.Kind != grammar.NameBind || o.draw == nil {
				continue
			}
			b := s.t.nameScope.bindings[o.Bound]
			r := selfRead{b: b, names: map[*nameBinding]bool{}, nodes: map[node]bool{}}
			for _, a := range o.clause {
				if !r.reaches(a) {
					continue
				}
				via := ""
				if a.named != b {
					via = fmt.Sprintf(", and drawing %q reads %q", a.head, b.name)
				}
				return fmt.Errorf("%s: token {%s}: its where clause reads %q%s, the name it binds, so the draw would read itself; bind what the clause reads to another name", s.label, o.Body, a.head, via)
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
		if v := a.named.values; v != nil {
			return slices.ContainsFunc(v.clause, r.reaches)
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
		return slices.ContainsFunc(renderEdges(n, nil), func(e renderEdge) bool { return r.reachesNode(e.to) })
	}
	for _, o := range t.compiled.ops {
		if slices.ContainsFunc(o.reads(), func(a *arm) bool { return r.reaches(*a) }) {
			return true
		}
	}
	return false
}

// valueDrawOf builds o's draw over the values o's path renders.
func (t *template) valueDrawOf(o *op, l *loadFinds) (*valueDraw, error) {
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
		start, tail = t.startOf(a.head, l.of(t).refs), a.tail
	} else {
		target := l.binding(l.of(t).nameScope.bindings[o.Bound]).target
		start, tail = target.start, target.tail
	}
	w := valueWalk{finds: l, seen: map[node]bool{}}
	if err := w.walk(start, tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("%s values fixed at load, and %w", clause, err)
	}
	return newValueDraw(w.out, o.Where, o.Distinct, o.clause), nil
}

// newValueDraw keeps the values that pass where's fixed conditions, shortest first, for draws of
// distinct's count, or of one value where distinct is nil.
func newValueDraw(values []weightedText, where *grammar.Where, distinct *grammar.Distinct, clause []arm) *valueDraw {
	v := &valueDraw{clause: clause, count: 1, fallback: fallbackOf(where, clause, "")}
	v.lo, v.hi = lengthEnds(where, clause)
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
	v.strings = readStrings{clauseReads(clause, where.NotNames), clauseReads(clause, where.NotContainingNames)}
	return v
}

// lengthEnds are the ends of w's length, 0 to every length where w has none.
func lengthEnds(w *grammar.Where, clause []arm) (lo, hi lengthEnd) {
	if w == nil || w.Length == nil {
		return lengthEnd{n: 0}, lengthEnd{n: math.MaxInt}
	}
	return endOf(w.Length.Lo, clause), endOf(w.Length.Hi, clause)
}

func endOf(l grammar.Length, clause []arm) lengthEnd {
	if l.Name == "" {
		return lengthEnd{n: l.N}
	}
	return lengthEnd{read: clauseRead(clause, l.Name)}
}

// lengthProof is what proves a draw's length: the lengths of its values, shortest first, the
// clause's ends, and how many values one draw takes.
type lengthProof struct {
	lengths []int
	lo, hi  lengthEnd
	count   int
	filters bool // a where clause filters the values
}

// proveDrawLengths refuses a draw in ts that some render leaves with too few values.
func proveDrawLengths(ts []templateSite) error {
	for _, s := range ts {
		for _, o := range s.t.compiled.ops {
			p, draws := drawLengthProof(o)
			if !draws {
				continue
			}
			if err := p.proveEveryDrawPasses(); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
			}
		}
	}
	return nil
}

// drawLengthProof is what proves the length of o's draw, where o draws through a where or a
// distinct clause.
func drawLengthProof(o op) (lengthProof, bool) {
	switch d := o.draw.(type) {
	case *valueDraw:
		return lengthProof{d.lengths, d.lo, d.hi, d.count, d.filters}, true
	case *alphabetDraw:
		p := lengthProof{lengths: []int{d.runs.Len()}, count: 1, filters: true}
		p.lo, p.hi = lengthEnds(o.Where, o.clause)
		return p, true
	}
	return lengthProof{}, false
}

// proveEveryDrawPasses refuses a draw that some render leaves with too few values: one where too few
// values pass, or one whose length reads a name that may print a length too few values have.
func (v lengthProof) proveEveryDrawPasses() error {
	bounds := map[*arm][2]float64{}
	for _, e := range []lengthEnd{v.lo, v.hi} {
		if e.read == nil {
			continue
		}
		p := (&valueProof{}).proveArm(*e.read)
		switch {
		case p.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", e.read.spelling, p.NotOperand)
		case !p.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", e.read.spelling)
		}
		bounds[e.read] = [2]float64{p.Lo, p.Hi}
	}
	if v.lo.read != nil && v.lo.read == v.hi.read {
		for n := bounds[v.lo.read][0]; n <= bounds[v.lo.read][1]; n++ {
			if passing := v.passing(n, n); passing < v.count {
				return fmt.Errorf("where %s is %.0f, %s", v.lo.read.spelling, n, v.shortfall(passing))
			}
		}
		return nil
	}
	lo, hi := narrowest(v.lo, v.hi, bounds)
	if passing := v.passing(lo, hi); passing < v.count {
		return v.rangeShortfall(lo, hi, passing)
	}
	return nil
}

// narrowest is the narrowest length lo to hi a render can read, given the bounds of what each
// name read prints. Every render must pass, so the low end takes its name's highest value and the
// high end its lowest, which may cross.
func narrowest(lo, hi lengthEnd, bounds map[*arm][2]float64) (float64, float64) {
	from, to := float64(lo.n), float64(hi.n)
	if lo.read != nil {
		from = bounds[lo.read][1]
	}
	if hi.read != nil {
		to = bounds[hi.read][0]
	}
	return from, to
}

// rangeShortfall says that the length lo to hi leaves passing values, too few, naming the length's
// names at the values that narrow it to lo to hi.
func (v lengthProof) rangeShortfall(lo, hi float64, passing int) error {
	var at []string
	if v.lo.read != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.lo.read.spelling, lo))
	}
	if v.hi.read != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.hi.read.spelling, hi))
	}
	switch {
	case len(at) == 0:
		return errors.New(v.shortfall(passing))
	case lo <= hi:
		return fmt.Errorf("where %s, %s", strings.Join(at, " and "), v.shortfall(passing))
	case v.hi.read == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %d", at[0], lo, hi, v.lo.read.spelling, v.hi.n)
	case v.lo.read == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at %d or above", at[0], lo, hi, v.hi.read.spelling, v.lo.n)
	}
	return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %s", strings.Join(at, " and "), lo, hi, v.lo.read.spelling, v.hi.read.spelling)
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

func (v *valueDraw) draw(s *drawstate.State, _ *template, env renderEnv, _ string) string {
	return strings.Join(v.drawValues(s, env), v.sep)
}

func (v *valueDraw) mayFallBack() bool { return v.strings.any() }

// drawValues draws count distinct passing values, reading the clause's names in env. Where the
// strings it reads leave fewer values than count, the fallback prints in place of each value missing.
func (v *valueDraw) drawValues(s *drawstate.State, env renderEnv) []string {
	from, to := v.span(v.lo.at(s, env), v.hi.at(s, env))
	if to-from < v.count {
		panic(invariant.Broken("a clause the load proved leaves %d values leaves %d", v.count, to-from))
	}
	drawn := v.holding(s, env, from, to) // ascending
	out := make([]string, v.count)
	for k := range out {
		if len(drawn) == to-from {
			out[k] = v.fallback.at(s, env)
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
	not, containing := v.strings.read(s, env)
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
func (r readStrings) read(s *drawstate.State, env renderEnv) (not, containing []string) {
	for _, a := range r.not {
		not = append(not, readName(s, env, *a).text)
	}
	for _, a := range r.notContaining {
		if text := readName(s, env, *a).text; text != "" {
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

func (e lengthEnd) at(s *drawstate.State, env renderEnv) int {
	if e.read == nil {
		return e.n
	}
	text := readName(s, env, *e.read).text
	v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		panic(invariant.Broken("a length bound read %q, which a load proved a number", text))
	}
	return int(v)
}
