package fence

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// drawsAlong reports whether the walk of path from start meets a choice or a table at one of its
// levels, or, with leaf, ends on a level that isn't a constant.
func drawsAlong(start tree.Node, path []string, leaf bool) bool {
	n := start
	for i := 0; ; i++ {
		switch {
		case isDraw(n):
			return true
		case i == len(path):
			return leaf && !isConstant(n)
		}
		var err error
		if n, err = load.ChildNamed(n, path[i]); err != nil {
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
