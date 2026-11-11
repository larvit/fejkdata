package tree

// ValueDraw draws among the values a path renders, fixed at load with the odds each has in a draw of
// the path: one of the values that pass its where clause, or count distinct ones where a distinct
// clause asks. A name the clause reads is read at each draw.
type ValueDraw struct{ spec ValueDrawSpec }

// ValueDrawSpec is what NewValueDraw builds a ValueDraw of.
type ValueDrawSpec struct {
	Index   LengthIndex
	Lo, Hi  LengthEnd
	Clause  []Arm  // the op's clause: the names the where clause reads
	Filters bool   // a where clause filters the values
	Count   int    // how many distinct values one draw takes
	Sep     string // what prints between them
	Strings ReadStrings
	// Fallback prints in place of each value a draw can't take, once a string read at render left
	// too few
	Fallback Fallback
}

func NewValueDraw(s ValueDrawSpec) *ValueDraw { return &ValueDraw{s} }

func (*ValueDraw) isTokenDraw() {}

func (v *ValueDraw) MayFallBack() bool { return v.spec.Strings.Any() }

func (v *ValueDraw) Clause() []Arm { return v.spec.Clause }

func (v *ValueDraw) Count() int { return v.spec.Count }

func (v *ValueDraw) Fallback() Fallback { return v.spec.Fallback }

func (v *ValueDraw) Filters() bool { return v.spec.Filters }

func (v *ValueDraw) Hi() LengthEnd { return v.spec.Hi }

func (v *ValueDraw) Index() *LengthIndex { return &v.spec.Index }

func (v *ValueDraw) Lo() LengthEnd { return v.spec.Lo }

func (v *ValueDraw) Sep() string { return v.spec.Sep }

func (v *ValueDraw) Strings() ReadStrings { return v.spec.Strings }

// LengthEnd is one end of a where clause's length: n, or, where read is not nil, what that read of
// a name prints.
type LengthEnd struct {
	n    int
	read *Arm
}

func NewLengthEndN(n int) LengthEnd { return LengthEnd{n: n} }

func NewLengthEndRead(read *Arm) LengthEnd { return LengthEnd{read: read} }

func (e LengthEnd) N() int { return e.n }

func (e LengthEnd) Read() *Arm { return e.read }

// ReadStrings are the reads of the strings a value must not equal, and of those it must not contain.
type ReadStrings struct{ not, notContaining []*Arm }

func NewReadStrings(not, notContaining []*Arm) ReadStrings { return ReadStrings{not, notContaining} }

// Any says whether r reads a string, which alone can leave a draw too few values.
func (r ReadStrings) Any() bool { return len(r.not)+len(r.notContaining) > 0 }

func (r ReadStrings) Not() []*Arm { return r.not }

func (r ReadStrings) NotContaining() []*Arm { return r.notContaining }

// Fallback is what a draw prints where no value passes: text, or, where read is not nil, what that
// read of a name prints.
type Fallback struct {
	text string
	read *Arm
}

func NewFallbackText(text string) Fallback { return Fallback{text: text} }

func NewFallbackRead(read *Arm) Fallback { return Fallback{read: read} }

func (f Fallback) Read() *Arm { return f.read }

func (f Fallback) Text() string { return f.text }
