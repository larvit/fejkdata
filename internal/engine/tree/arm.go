package tree

import (
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

// Arm is one alternative of a {a|b} token or one operand, split into the head,
// whose node `Template.StartOf` finds, and the tail of a dotted path into it.
type Arm struct {
	spelling string // as written, for messages
	head     string
	tail     []string
	levels   []PickKey // the key of each level the steps pass, from the head they start at to the leaf; a step's At indexes it
	steps    []PathStep
	leaves   []Node // every node the path may land on, one per variant it passes
	kind     ArmKind
	named    *NameBinding // NamedRead: the binding of the name it reads through
	member   int          // NamedRead of a distinct draw: the member it reads, counted from 1; 0 reads them joined
}

func (a Arm) Head() string { return a.head }

func (a Arm) Kind() ArmKind { return a.kind }

func (a Arm) Leaves() []Node { return a.leaves }

func (a Arm) Levels() []PickKey { return a.levels }

func (a Arm) Member() int { return a.member }

func (a Arm) Named() *NameBinding { return a.named }

func (a Arm) Spelling() string { return a.spelling }

func (a Arm) Steps() []PathStep { return a.steps }

func (a Arm) Tail() []string { return a.tail }

// ArmKind is how a render reads an arm, fixed at compile.
type ArmKind uint8

const (
	FreshRead ArmKind = iota // drawn afresh at every read
	NamedRead                // read through a name, kept in its pick
)

func NewArm(spelling, head string, tail []string, levels []PickKey, steps []PathStep, leaves []Node) Arm {
	return Arm{spelling: spelling, head: head, tail: tail, levels: levels, steps: steps, leaves: leaves}
}

// ThroughName is a read through b's pick, of b's member where b draws distinct members.
func (a Arm) ThroughName(b *NameBinding, member int) Arm {
	a.kind, a.named, a.member = NamedRead, b, member
	return a
}

// Key is the key of the leaf a lands on, the one key every way of writing this read shares.
func (a Arm) Key() PickKey { return a.levels[len(a.levels)-1] }

// ReadsJoined is whether a reads the members of a distinct draw joined, which is text whatever
// they are.
func (a Arm) ReadsJoined() bool {
	return a.kind == NamedRead && a.named.Distinct() != nil && a.member == 0
}

// Op is one compiled unit of a format string: a literal run, a read,
// or a builtin already prepared with its args.
type Op struct {
	token grammar.Token
	arms  []Arm // grammar.PathRead: the '|' alternatives, split into head and tail once
	call  builtinfunc.Call
	// operands are the fields the builtin reads, in the order its operands func fixed; a render
	// reads them before a call that has no draw. nil for a token that reads none.
	operands []Arm
	// clause is the reads of the token's where clause, each once: its length's first, then not's,
	// not containing's and else's
	clause []Arm
	// draw is what a token with a where or a distinct clause, or a {reword()} call, prints in
	// place of a plain read or call; nil for any other token
	draw TokenDraw
}

// NewOp is the op of tok before its call and draw: the arms it reads, its builtin's
// operands and its where clause's reads.
func NewOp(tok grammar.Token, arms, operands, clause []Arm) Op {
	return Op{token: tok, arms: arms, operands: operands, clause: clause}
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

// Reads is every arm a render of o reads: its arms, its builtin's operands and its clause's reads.
// Each points into o's slices, so the same arm always yields the same pointer.
func (o *Op) Reads() []*Arm {
	found := make([]*Arm, 0, len(o.arms)+len(o.operands)+len(o.clause))
	for _, arms := range [][]Arm{o.arms, o.operands, o.clause} {
		for i := range arms {
			found = append(found, &arms[i])
		}
	}
	return found
}

func (o *Op) Arms() []Arm { return o.arms }

func (o *Op) Call() builtinfunc.Call { return o.call }

func (o *Op) Clause() []Arm { return o.clause }

func (o *Op) Draw() TokenDraw { return o.draw }

func (o *Op) Operands() []Arm { return o.operands }

func (o *Op) Token() grammar.Token { return o.token }

// TokenDraw is a token's draw, built at load once its paths resolve: a *ValueDraw, an
// *AlphabetDraw or a *RewordDraw, and no other.
type TokenDraw interface {
	isTokenDraw()
	// MayFallBack says whether a draw may print the where clause's fallback, or a path's empty
	// default, where the strings it reads leave no value.
	MayFallBack() bool
}
