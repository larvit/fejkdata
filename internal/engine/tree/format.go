package tree

// UnboundRead is a token reading a head no field holds, which only a name bound around the
// template can answer; err is the refusal where none does, and body is the token's.
type UnboundRead struct {
	head, body string
	err        error
}

func NewUnboundRead(head, body string, err error) UnboundRead { return UnboundRead{head, body, err} }

// FormatOps is a compiled format: its ops, and the size of its literal text.
type FormatOps struct {
	ops  []Op
	grow int
}

func NewFormatOps(ops []Op, grow int) FormatOps { return FormatOps{ops, grow} }

func (r FormatOps) Grow() int { return r.grow }

func (r FormatOps) Ops() []Op { return r.ops }

func (r UnboundRead) Body() string { return r.body }

func (r UnboundRead) Err() error { return r.err }

func (r UnboundRead) Head() string { return r.head }
