package load

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

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

// clauseRead returns the read in clause, an op's compiled where clause, that reads spelling; spelling
// must be one the clause reads.
func clauseRead(clause []tree.Read, spelling string) *tree.Read {
	return &clause[slices.IndexFunc(clause, func(a tree.Read) bool { return a.Spelling() == spelling })]
}

func clauseReads(clause []tree.Read, of []string) []*tree.Read {
	var reads []*tree.Read
	for _, n := range of {
		reads = append(reads, clauseRead(clause, n))
	}
	return reads
}

// fallbackOf is what a draw under w prints where no value passes, deflt where w names no fallback.
func fallbackOf(w *grammar.Where, clause []tree.Read, deflt string) tree.Fallback {
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
func compileClauseNames(t *tree.Template, w *grammar.Where, l *tree.LoadFinds) ([]tree.Read, error) {
	var reads []tree.Read
	for _, c := range clauseSpellings(w) {
		read := splitRead(c.spelling, nil)
		if !isName(t, read.head, l) {
			return nil, c.fieldRead()
		}
		if b := l.Of(t).NameScope().Lookup(read.head); b.Distinct() != nil && len(read.tail) == 0 && c.listsStrings() {
			return nil, c.membersRead(b.Distinct().N)
		}
		a, err := compileRead(t, c.spelling, l)
		if err != nil {
			return nil, err
		}
		reads = append(reads, a)
	}
	return reads, nil
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
		a := o.Alternatives()[0]
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
func newValueDraw(values []weightedText, where *grammar.Where, distinct *grammar.Distinct, clause []tree.Read) *tree.ValueDraw {
	lo, hi := LengthEnds(where, clause)
	spec := tree.ValueDrawSpec{Lo: lo, Hi: hi, Clause: clause, Count: 1, Fallback: fallbackOf(where, clause, "")}
	if distinct != nil {
		spec.Count, spec.Sep = distinct.N, distinct.Sep
	}
	if where == nil {
		spec.Index = newLengthIndex(values)
		return tree.NewValueDraw(spec)
	}
	var kept []weightedText
	containing := nonEmpty(where.NotContaining)
	for _, value := range values {
		if !slices.Contains(where.Not, value.text) && !avoid.HoldsAny(value.text, containing) {
			kept = append(kept, value)
		}
	}
	spec.Index, spec.Filters = newLengthIndex(kept), true
	spec.Strings = tree.NewReadStrings(clauseReads(clause, where.NotNames), clauseReads(clause, where.NotContainingNames))
	return tree.NewValueDraw(spec)
}

// LengthEnds are the ends of w's length, 0 to every length where w has none.
func LengthEnds(w *grammar.Where, clause []tree.Read) (lo, hi tree.LengthEnd) {
	if w == nil || w.Length == nil {
		return tree.NewLengthEndN(0), tree.NewLengthEndN(math.MaxInt)
	}
	return endOf(w.Length.Lo, clause), endOf(w.Length.Hi, clause)
}

func endOf(l grammar.Length, clause []tree.Read) tree.LengthEnd {
	if l.Name == "" {
		return tree.NewLengthEndN(l.N)
	}
	return tree.NewLengthEndRead(clauseRead(clause, l.Name))
}

// nonEmpty is texts without the empty string, which every value holds.
func nonEmpty(texts []string) []string {
	return slices.DeleteFunc(slices.Clone(texts), func(t string) bool { return t == "" })
}
