package tree

import (
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

// Read is one read of a path or a name: one alternative of a {a|b} token, one operand of a builtin
// or one name of its where clause. It is split into the head, whose node `Template.StartOf` finds,
// and the tail of a dotted path into it.
type Read struct {
	spelling string // as written, for messages
	head     string
	tail     []string
	levels   []PickKey // the key of each level the steps pass, from the head they start at to the leaf; a step's At indexes it
	steps    []PathStep
	leaves   []Node // every node the path may land on, one per variant it passes
	kind     ReadKind
	named    *NameBinding // NamedRead: the binding of the name it reads through
	member   int          // NamedRead of a distinct draw: the member it reads, counted from 1; 0 reads them joined
}

func (a Read) Head() string { return a.head }

func (a Read) Kind() ReadKind { return a.kind }

func (a Read) Leaves() []Node { return a.leaves }

func (a Read) Levels() []PickKey { return a.levels }

func (a Read) Member() int { return a.member }

func (a Read) Named() *NameBinding { return a.named }

func (a Read) Spelling() string { return a.spelling }

func (a Read) Steps() []PathStep { return a.steps }

func (a Read) Tail() []string { return a.tail }

// ReadKind is how a render takes the value of a read, fixed at compile.
type ReadKind uint8

const (
	FreshRead ReadKind = iota // drawn afresh, unless a name's pick holds its level
	NamedRead                 // read through a name, kept in its pick
)

func NewRead(spelling, head string, tail []string, levels []PickKey, steps []PathStep, leaves []Node) Read {
	return Read{spelling: spelling, head: head, tail: tail, levels: levels, steps: steps, leaves: leaves}
}

// ThroughName is a read through b's pick, of b's member where b draws distinct members.
func (a Read) ThroughName(b *NameBinding, member int) Read {
	a.kind, a.named, a.member = NamedRead, b, member
	return a
}

// Key is the key of the leaf a lands on, the one key every way of writing this read shares.
func (a Read) Key() PickKey { return a.levels[len(a.levels)-1] }

// ReadsJoined is whether a reads the members of a distinct draw joined, which is text whatever
// they are.
func (a Read) ReadsJoined() bool {
	return a.kind == NamedRead && a.named.Distinct() != nil && a.member == 0
}

// Op is one compiled unit of a format string: a literal run, a path read with its alternatives,
// or a builtin already prepared with its args.
type Op struct {
	token        grammar.Token
	alternatives []Read // grammar.PathRead: the '|' alternatives, split into head and tail once
	call         builtinfunc.Call
	// operands are the fields the builtin reads, in the order its operands func fixed; a render
	// reads them before a call that has no draw. nil for a token that reads none.
	operands []Read
	// clause is the reads of the token's where clause, each once: its length's first, then not's,
	// not containing's and else's
	clause []Read
	// draw is what a token with a where or a distinct clause, or a {reword()} call, prints in
	// place of a plain read or call; nil for any other token
	draw TokenDraw
}

// NewOp is the op of tok before its call and draw: the reads of its alternatives, its builtin's
// operands and its where clause's reads.
func NewOp(tok grammar.Token, alternatives, operands, clause []Read) Op {
	return Op{token: tok, alternatives: alternatives, operands: operands, clause: clause}
}

// WithCall is o with its builtin prepared.
func (o Op) WithCall(call builtinfunc.Call) Op {
	o.call = call
	return o
}

// WithDraw is o with the draw it prints in place of a plain read or call.
func (o Op) WithDraw(d TokenDraw) Op {
	o.draw = d
	return o
}

// AllReads is every read o holds: those of its alternatives, its builtin's operands and its
// clause's. Each points into o's slices, so the same read always yields the same pointer.
func (o *Op) AllReads() []*Read {
	found := make([]*Read, 0, len(o.alternatives)+len(o.operands)+len(o.clause))
	for _, group := range [][]Read{o.alternatives, o.operands, o.clause} {
		for i := range group {
			found = append(found, &group[i])
		}
	}
	return found
}

func (o *Op) Alternatives() []Read { return o.alternatives }

// AlternativeLeaves is every leaf of o's alternatives, in order.
func (o *Op) AlternativeLeaves() []Node {
	var leaves []Node
	for _, a := range o.alternatives {
		leaves = append(leaves, a.Leaves()...)
	}
	return leaves
}

func (o *Op) Call() builtinfunc.Call { return o.call }

func (o *Op) Clause() []Read { return o.clause }

func (o *Op) Draw() TokenDraw { return o.draw }

func (o *Op) Operands() []Read { return o.operands }

func (o *Op) Token() grammar.Token { return o.token }

// TokenDraw is a token's draw, built at load once its paths resolve: a *ValueDraw, an
// *AlphabetDraw or a *RewordDraw, and no other.
type TokenDraw interface {
	isTokenDraw()
	// MayFallBack says whether a draw may print the where clause's fallback, or a path's empty
	// default, where the strings it reads leave no value.
	MayFallBack() bool
}
