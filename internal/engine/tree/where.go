package tree

// ValueDraw draws among the values a path renders, fixed at load with the odds each has in a draw of
// the path: one of the values that pass its where clause, or count distinct ones where a distinct
// clause asks. A name the clause reads is read at each draw.
type ValueDraw struct {
	index   LengthIndex
	lo, hi  LengthEnd
	clause  []Arm  // the op's clause: the names the where clause reads
	filters bool   // a where clause filters the values
	count   int    // how many distinct values one draw takes
	sep     string // what prints between them
	strings ReadStrings
	// fallback prints in place of each value a draw can't take, once a string read at render left
	// too few
	fallback Fallback
}

func NewValueDraw(index LengthIndex, lo, hi LengthEnd, clause []Arm, filters bool, count int, sep string, strings ReadStrings, fallback Fallback) *ValueDraw {
	return &ValueDraw{index, lo, hi, clause, filters, count, sep, strings, fallback}
}

// LengthEnd is one end of a where clause's length: n, or, where read is not nil, what that read of
// a name prints.
type LengthEnd struct {
	n    int
	read *Arm
}

func NewLengthEndN(n int) LengthEnd { return LengthEnd{n: n} }

func NewLengthEndRead(read *Arm) LengthEnd { return LengthEnd{read: read} }

// ReadStrings are the reads of the strings a value must not equal, and of those it must not contain.
type ReadStrings struct{ not, notContaining []*Arm }

func NewReadStrings(not, notContaining []*Arm) ReadStrings { return ReadStrings{not, notContaining} }

// Any says whether r reads a string, which alone can leave a draw too few values.
func (r ReadStrings) Any() bool { return len(r.not)+len(r.notContaining) > 0 }

// Fallback is what a draw prints where no value passes: text, or, where read is not nil, what that
// read of a name prints.
type Fallback struct {
	text string
	read *Arm
}

func NewFallbackText(text string) Fallback { return Fallback{text: text} }

func NewFallbackRead(read *Arm) Fallback { return Fallback{read: read} }

func (*ValueDraw) isTokenDraw() {}

func (v *ValueDraw) MayFallBack() bool { return v.strings.Any() }

func (r Fallback) Read() *Arm { return r.read }

func (r Fallback) Text() string { return r.text }

func (r LengthEnd) N() int { return r.n }

func (r LengthEnd) Read() *Arm { return r.read }

func (r ReadStrings) Not() []*Arm { return r.not }

func (r ReadStrings) NotContaining() []*Arm { return r.notContaining }

func (r *ValueDraw) Clause() []Arm { return r.clause }

func (r *ValueDraw) Count() int { return r.count }

func (r *ValueDraw) Fallback() Fallback { return r.fallback }

func (r *ValueDraw) Filters() bool { return r.filters }

func (r *ValueDraw) Hi() LengthEnd { return r.hi }

func (r *ValueDraw) Index() *LengthIndex { return &r.index }

func (r *ValueDraw) Lo() LengthEnd { return r.lo }

func (r *ValueDraw) Sep() string { return r.sep }

func (r *ValueDraw) Strings() ReadStrings { return r.strings }
