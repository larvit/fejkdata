package names

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// pathWalk is a walk down path from a node, a level at a time. Field n is the node it stands on,
// and steps counts the segments of path it walked to get there.
type pathWalk struct {
	n     tree.Node
	path  []string
	steps int
}

func walkFrom(start tree.Node, path []string) *pathWalk {
	return &pathWalk{n: start, path: path}
}

// levelKind is what a walk stands on. The node and whether the path ends on it tell the kinds
// apart, so a level is of exactly one.
type levelKind uint8

const (
	drawLevel    levelKind = iota // a choice or a table, which draws wherever on the path it stands
	innerLevel                    // any other node the path goes on below
	constantLeaf                  // the path ends on a template whose every render is the same text
	varyingLeaf                   // the path ends on any other node but a choice or a table
)

func (w *pathWalk) kind() levelKind {
	switch {
	case isDraw(w.n):
		return drawLevel
	case w.steps < len(w.path):
		return innerLevel
	case isConstant(w.n):
		return constantLeaf
	}
	return varyingLeaf
}

// walked is the segments leading to the level the walk stands on.
func (w *pathWalk) walked() []string { return w.path[:w.steps] }

// drawsBelow reports whether the rest of the walk meets a choice or a table, here or below, or
// ends on a level that isn't a constant.
func (w *pathWalk) drawsBelow() bool { return drawsAlong(w.n, w.path[w.steps:], true) }

// step moves the walk down one level. It reports false, and stays, where no field spells the
// next segment, such as "..".
func (w *pathWalk) step() bool {
	n, err := load.ChildNamed(w.n, w.path[w.steps])
	if err != nil {
		return false
	}
	w.n, w.steps = n, w.steps+1
	return true
}

// drawsAlong reports whether the walk of path from start meets a choice or a table at one of its
// levels, or, with leaf, ends on a level that isn't a constant. A walk it can't follow counts as
// meeting one.
func drawsAlong(start tree.Node, path []string, leaf bool) bool {
	for w := walkFrom(start, path); ; {
		switch w.kind() {
		case drawLevel:
			return true
		case constantLeaf:
			return false
		case varyingLeaf:
			return leaf
		}
		if !w.step() {
			return true
		}
	}
}

// isDraw reports whether n draws a variant or a row, which a pick keeps and a fresh read does not.
func isDraw(n tree.Node) bool {
	switch n.(type) {
	case *tree.Choice, *tree.Table:
		return true
	}
	return false
}

// isConstant reports whether every render of n is the same text.
func isConstant(n tree.Node) bool {
	t, isTemplate := n.(*tree.Template)
	if !isTemplate {
		return false
	}
	_, fixed := t.FixedText()
	return fixed
}
