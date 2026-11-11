package tree

// PathStep is one step of a compiled path, taken at the node the steps before it
// reached, a choice there drawn first; At indexes the tail where it is taken, and
// only a table's step sits past the tail's end. A step holds no node, so a
// caller's steps stay on its stack.
type PathStep struct {
	Kind StepKind
	At   int
	Name string // StepField: the field; StepSelect: the selector; StepColumn, StepChild: the column or table stepped to
	Row  int    // StepSelect: the row the selector names
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

// Child is what seg names under n: a folder's entry or a template's field.
func Child(n Node, seg string) (Node, bool) {
	var child Node
	ok := false
	switch n := n.(type) {
	case *Folder:
		child, ok = n.children[seg]
	case *Template:
		child, ok = n.syntax.Fields[seg]
	}
	return child, ok
}
