package tree

// PathStep is one step of a compiled path, taken at the node the steps before it
// reached, a choice there drawn first; at indexes the tail where it is taken, and
// only a table's step sits past the tail's end. A step holds no node, so a
// caller's steps stay on its stack.
type PathStep struct {
	kind StepKind
	at   int
	name string // StepField: the field; StepSelect: the selector; StepColumn, StepChild: the column or table stepped to
	row  int    // StepSelect: the row the selector names
}

func NewPathStep(kind StepKind, at int, name string, row int) PathStep {
	return PathStep{kind: kind, at: at, name: name, row: row}
}

type StepKind uint8

const (
	StepField  StepKind = iota + 1 // into a folder's entry or a template's field
	StepSelect                     // pin the selected row of the table
	StepDraw                       // draw a row of the table inside the pins
	StepRow                        // land on the table's row node
	StepColumn
	StepChild
	StepParent // up to the table's parent, at the row its own row links to
)

func (r PathStep) At() int { return r.at }

func (r PathStep) Kind() StepKind { return r.kind }

func (r PathStep) Name() string { return r.name }

func (r PathStep) Row() int { return r.row }
