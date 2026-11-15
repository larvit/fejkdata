package tree

// UnboundRead is a token reading a head no field holds, which only a name bound around the
// template can answer; refusal is the error where none does, and body is the token's.
type UnboundRead struct {
	head, body string
	refusal    error
}

func NewUnboundRead(head, body string, refusal error) UnboundRead {
	return UnboundRead{head, body, refusal}
}

func (u UnboundRead) Body() string { return u.body }

func (u UnboundRead) Refusal() error { return u.refusal }

func (u UnboundRead) Head() string { return u.head }

// FormatOps is a compiled format: its ops, and the size of its literal text.
type FormatOps struct {
	ops  []Op
	grow int
}

func NewFormatOps(ops []Op, grow int) FormatOps { return FormatOps{ops, grow} }

func (f FormatOps) Grow() int { return f.grow }

func (f FormatOps) Ops() []Op { return f.ops }
