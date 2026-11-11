package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

func fallbackAt(f tree.Fallback, s *drawstate.State, env renderEnv) string {
	if f.Read() == nil {
		return f.Text()
	}
	return readName(s, env, *f.Read()).text
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
func clauseRead(clause []tree.Arm, spelling string) *tree.Arm {
	return &clause[slices.IndexFunc(clause, func(a tree.Arm) bool { return a.Spelling() == spelling })]
}

func clauseReads(clause []tree.Arm, of []string) []*tree.Arm {
	var reads []*tree.Arm
	for _, n := range of {
		reads = append(reads, clauseRead(clause, n))
	}
	return reads
}

// fallbackOf is what a draw under w prints where no value passes, deflt where w names no fallback.
func fallbackOf(w *grammar.Where, clause []tree.Arm, deflt string) tree.Fallback {
	switch {
	case w == nil || w.Else == nil:
		return tree.NewFallbackText(deflt)
	case w.Else.Name != "":
		return tree.NewFallbackRead(clauseRead(clause, w.Else.Name))
	}
	return tree.NewFallbackText(w.Else.Text)
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
func compileClauseNames(t *tree.Template, w *grammar.Where, l *tree.LoadFinds) ([]tree.Arm, error) {
	var arms []tree.Arm
	for _, c := range clauseSpellings(w) {
		read := splitRead(c.spelling, nil)
		if !isName(t, read.head, l) {
			return nil, c.fieldRead()
		}
		if b := l.Of(t).NameScope().Lookup(read.head); b.Distinct() != nil && len(read.tail) == 0 && c.listsStrings() {
			return nil, c.membersRead(b.Distinct().N)
		}
		a, err := compileArm(t, c.spelling, l)
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
			if r.a.Named().Values() != nil && len(r.a.Tail()) > 0 {
				return fmt.Errorf("%s: token {%s}: name %q holds a filtered draw, which is read whole: {%s}", s.label, r.o.Token().Body, r.a.Head(), r.a.Head())
			}
		}
	}
	return nil
}

// refuseSelfReads refuses a token whose where clause reads the name the token binds, directly or
// through what the names it reads draw: drawing the name would read itself.
func refuseSelfReads(ts []templateSite) error {
	for _, s := range ts {
		for _, o := range s.t.Compiled().Ops() {
			if o.Token().Kind != grammar.NameBind || o.Draw() == nil {
				continue
			}
			b := s.t.NameScope().Bindings()[o.Token().Bound]
			r := selfRead{b: b, names: map[*tree.NameBinding]bool{}, nodes: map[tree.Node]bool{}}
			for _, a := range o.Clause() {
				if !r.reaches(a) {
					continue
				}
				via := ""
				if a.Named() != b {
					via = fmt.Sprintf(", and drawing %q reads %q", a.Head(), b.Name())
				}
				return fmt.Errorf("%s: token {%s}: its where clause reads %q%s, the name it binds, so the draw would read itself; bind what the clause reads to another name", s.label, o.Token().Body, a.Head(), via)
			}
		}
	}
	return nil
}

// selfRead walks what reading an arm draws, looking for a read of b.
type selfRead struct {
	b     *tree.NameBinding
	names map[*tree.NameBinding]bool
	nodes map[tree.Node]bool
}

// reaches says whether reading a reads b: a name holding a filtered draw reads only its clause's
// names, since its values are fixed at load; any other read renders its leaves.
func (r selfRead) reaches(a tree.Arm) bool {
	if a.Kind() == tree.NamedRead {
		if a.Named() == r.b {
			return true
		}
		if r.names[a.Named()] {
			return false
		}
		r.names[a.Named()] = true
		if v := a.Named().Values(); v != nil {
			return slices.ContainsFunc(v.Clause(), r.reaches)
		}
	}
	return slices.ContainsFunc(a.Leaves(), r.reachesNode)
}

func (r selfRead) reachesNode(n tree.Node) bool {
	if r.nodes[n] {
		return false
	}
	r.nodes[n] = true
	t, ok := n.(*tree.Template)
	if !ok {
		return slices.ContainsFunc(renderEdges(n, nil), func(e renderEdge) bool { return r.reachesNode(e.to) })
	}
	for _, o := range t.Compiled().Ops() {
		if slices.ContainsFunc(o.Reads(), func(a *tree.Arm) bool { return r.reaches(*a) }) {
			return true
		}
	}
	return false
}

// valueDrawOf builds o's draw over the values o's path renders.
func valueDrawOf(t *tree.Template, o *tree.Op, l *tree.LoadFinds) (*tree.ValueDraw, error) {
	clause := "a where clause filters"
	if o.Token().Where == nil {
		clause = "a distinct clause draws"
	}
	var start tree.Node
	var tail []string
	if o.Token().Kind == grammar.PathRead {
		a := o.Arms()[0]
		if a.Kind() == tree.NamedRead {
			return nil, fmt.Errorf("%q is a name, drawn once; write the clause on the draw it binds: {%s … as %s}", a.Head(), a.Named().Ref(), a.Head())
		}
		start, tail = t.StartOf(a.Head(), l.Of(t).Refs()), a.Tail()
	} else {
		target := l.Binding(l.Of(t).NameScope().Bindings()[o.Token().Bound]).Target()
		start, tail = target.Start(), target.Tail()
	}
	w := valueWalk{finds: l, seen: map[tree.Node]bool{}}
	if err := w.walk(start, tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("%s values fixed at load, and %w", clause, err)
	}
	return newValueDraw(w.out, o.Token().Where, o.Token().Distinct, o.Clause()), nil
}

// newValueDraw keeps the values that pass where's fixed conditions, shortest first, for draws of
// distinct's count, or of one value where distinct is nil.
func newValueDraw(values []weightedText, where *grammar.Where, distinct *grammar.Distinct, clause []tree.Arm) *tree.ValueDraw {
	lo, hi := lengthEnds(where, clause)
	count, sep := 1, ""
	if distinct != nil {
		count, sep = distinct.N, distinct.Sep
	}
	fallback := fallbackOf(where, clause, "")
	if where == nil {
		return tree.NewValueDraw(newLengthIndex(values), lo, hi, clause, false, count, sep, tree.ReadStrings{}, fallback)
	}
	var kept []weightedText
	containing := nonEmpty(where.NotContaining)
	for _, value := range values {
		if !slices.Contains(where.Not, value.text) && !holdsAny(value.text, containing) {
			kept = append(kept, value)
		}
	}
	strings := tree.NewReadStrings(clauseReads(clause, where.NotNames), clauseReads(clause, where.NotContainingNames))
	return tree.NewValueDraw(newLengthIndex(kept), lo, hi, clause, true, count, sep, strings, fallback)
}

// lengthEnds are the ends of w's length, 0 to every length where w has none.
func lengthEnds(w *grammar.Where, clause []tree.Arm) (lo, hi tree.LengthEnd) {
	if w == nil || w.Length == nil {
		return tree.NewLengthEndN(0), tree.NewLengthEndN(math.MaxInt)
	}
	return endOf(w.Length.Lo, clause), endOf(w.Length.Hi, clause)
}

func endOf(l grammar.Length, clause []tree.Arm) tree.LengthEnd {
	if l.Name == "" {
		return tree.NewLengthEndN(l.N)
	}
	return tree.NewLengthEndRead(clauseRead(clause, l.Name))
}

// lengthProof is what proves a draw's length: the lengths of its values, shortest first, the
// clause's ends, and how many values one draw takes.
type lengthProof struct {
	lengths []int
	lo, hi  tree.LengthEnd
	count   int
	filters bool // a where clause filters the values
}

// proveDrawLengths refuses a draw in ts that some render leaves with too few values.
func proveDrawLengths(ts []templateSite) error {
	for _, s := range ts {
		for _, o := range s.t.Compiled().Ops() {
			p, draws := drawLengthProof(o)
			if !draws {
				continue
			}
			if err := p.proveEveryDrawPasses(); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, o.Token().Body, err)
			}
		}
	}
	return nil
}

// drawLengthProof is what proves the length of o's draw, where o draws through a where or a
// distinct clause.
func drawLengthProof(o tree.Op) (lengthProof, bool) {
	switch d := o.Draw().(type) {
	case *tree.ValueDraw:
		return lengthProof{d.Index().Lengths(), d.Lo(), d.Hi(), d.Count(), d.Filters()}, true
	case *tree.AlphabetDraw:
		p := lengthProof{lengths: []int{d.Runs().Len()}, count: 1, filters: true}
		p.lo, p.hi = lengthEnds(o.Token().Where, o.Clause())
		return p, true
	}
	return lengthProof{}, false
}

// proveEveryDrawPasses refuses a draw that some render leaves with too few values: one where too few
// values pass, or one whose length reads a name that may print a length too few values have.
func (v lengthProof) proveEveryDrawPasses() error {
	bounds := map[*tree.Arm][2]float64{}
	for _, e := range []tree.LengthEnd{v.lo, v.hi} {
		if e.Read() == nil {
			continue
		}
		p := (&valueProof{}).proveArm(*e.Read())
		switch {
		case p.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", e.Read().Spelling(), p.NotOperand)
		case !p.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", e.Read().Spelling())
		}
		bounds[e.Read()] = [2]float64{p.Lo, p.Hi}
	}
	if v.lo.Read() != nil && v.lo.Read() == v.hi.Read() {
		for n := bounds[v.lo.Read()][0]; n <= bounds[v.lo.Read()][1]; n++ {
			if passing := v.passing(n, n); passing < v.count {
				return fmt.Errorf("where %s is %.0f, %s", v.lo.Read().Spelling(), n, v.shortfall(passing))
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
func narrowest(lo, hi tree.LengthEnd, bounds map[*tree.Arm][2]float64) (float64, float64) {
	from, to := float64(lo.N()), float64(hi.N())
	if lo.Read() != nil {
		from = bounds[lo.Read()][1]
	}
	if hi.Read() != nil {
		to = bounds[hi.Read()][0]
	}
	return from, to
}

// rangeShortfall says that the length lo to hi leaves passing values, too few, naming the length's
// names at the values that narrow it to lo to hi.
func (v lengthProof) rangeShortfall(lo, hi float64, passing int) error {
	var at []string
	if v.lo.Read() != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.lo.Read().Spelling(), lo))
	}
	if v.hi.Read() != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.hi.Read().Spelling(), hi))
	}
	switch {
	case len(at) == 0:
		return errors.New(v.shortfall(passing))
	case lo <= hi:
		return fmt.Errorf("where %s, %s", strings.Join(at, " and "), v.shortfall(passing))
	case v.hi.Read() == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %d", at[0], lo, hi, v.lo.Read().Spelling(), v.hi.N())
	case v.lo.Read() == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at %d or above", at[0], lo, hi, v.hi.Read().Spelling(), v.lo.N())
	}
	return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %s", strings.Join(at, " and "), lo, hi, v.lo.Read().Spelling(), v.hi.Read().Spelling())
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
	from, to := tree.SpanOf(v.lengths, int(min(max(lo, -1), past)), int(max(min(hi, past), -1)))
	return max(to-from, 0)
}

func drawValue(v *tree.ValueDraw, s *drawstate.State, env renderEnv) string {
	return strings.Join(drawValues(v, s, env), v.Sep())
}

// drawValues draws count distinct passing values, reading the clause's names in env. Where the
// strings it reads leave fewer values than count, the fallback prints in place of each value missing.
func drawValues(v *tree.ValueDraw, s *drawstate.State, env renderEnv) []string {
	from, to := v.Index().Span(lengthEndAt(v.Lo(), s, env), lengthEndAt(v.Hi(), s, env))
	if to-from < v.Count() {
		panic(invariant.Broken("a clause the load proved leaves %d values leaves %d", v.Count(), to-from))
	}
	drawn := holding(v, s, env, from, to) // ascending
	out := make([]string, v.Count())
	for k := range out {
		if len(drawn) == to-from {
			out[k] = fallbackAt(v.Fallback(), s, env)
			continue
		}
		i := drawIndex(v.Index(), s, from, to, drawn)
		at, _ := slices.BinarySearch(drawn, i)
		drawn = slices.Insert(drawn, at, i)
		out[k] = v.Index().Texts()[i]
	}
	return out
}

// holding is the indexes, ascending, of the texts in v.texts[from:to] that equal a not string, or
// hold a not containing string, that v reads in env.
func holding(v *tree.ValueDraw, s *drawstate.State, env renderEnv, from, to int) []int {
	not, containing := readClauseStrings(v.Strings(), s, env)
	if len(not) == 0 && len(containing) == 0 {
		return nil
	}
	var out []int
	for i := from; i < to; i++ {
		if slices.Contains(not, v.Index().Texts()[i]) || holdsAny(v.Index().Texts()[i], containing) {
			out = append(out, i)
		}
	}
	return out
}

// readClauseStrings reads r's strings in env: those a value must not equal, and those a value
// must not hold, leaving out an empty one, which every value holds.
func readClauseStrings(r tree.ReadStrings, s *drawstate.State, env renderEnv) (not, containing []string) {
	for _, a := range r.Not() {
		not = append(not, readName(s, env, *a).text)
	}
	for _, a := range r.NotContaining() {
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

func lengthEndAt(e tree.LengthEnd, s *drawstate.State, env renderEnv) int {
	if e.Read() == nil {
		return e.N()
	}
	text := readName(s, env, *e.Read()).text
	v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		panic(invariant.Broken("a length bound read %q, which a load proved a number", text))
	}
	return int(v)
}
