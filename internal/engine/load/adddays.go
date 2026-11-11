package load

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

func addDaysLayout(d tree.Read, l *tree.LoadFinds) (string, error) {
	layout, why := dateLayout(d.Leaves(), l, map[tree.Node]bool{})
	switch {
	case why != "":
		return "", fmt.Errorf("%s is not proven a date: %s; read a field whose format is one date(from,to,'layout') or born(min,max,'layout')", d.Spelling(), why)
	case !builtinfunc.ReadsBack(layout):
		return "", fmt.Errorf("%s prints in '%s', which does not read back as the day it printed; give it a layout with a four-digit year, a month and a day, such as '2006-01-02'", d.Spelling(), layout)
	}
	return layout, nil
}

// dateLayout is the one layout every render of nodes prints a date in, or why no proof holds.
func dateLayout(nodes []tree.Node, l *tree.LoadFinds, seen map[tree.Node]bool) (layout, why string) {
	for _, n := range nodes {
		got, w := nodeDateLayout(n, l, seen)
		switch {
		case w != "":
			return "", w
		case layout != "" && got != layout:
			return "", fmt.Sprintf("it prints in two layouts, '%s' and '%s'; give both one layout", layout, got)
		}
		layout = got
	}
	return layout, ""
}

func nodeDateLayout(n tree.Node, l *tree.LoadFinds, seen map[tree.Node]bool) (string, string) {
	if seen[n] {
		return "", "it reads itself"
	}
	seen[n] = true
	defer delete(seen, n)
	switch n := n.(type) {
	case *tree.Choice:
		return dateLayout(n.Items(), l, seen)
	case *tree.Template:
		ops := l.Of(n).Compiled().Ops()
		if n.Repeat() != 1 || len(n.Tokens()) != 1 || len(ops) != 1 {
			return "", fmt.Sprintf("it renders %q", n.Format())
		}
		o := ops[0]
		if layout, isDate := builtinfunc.DateLayout(o.Token().Fn, o.Token().Args); isDate {
			return layout, ""
		}
		switch {
		case o.Token().Fn == builtinfunc.AddDaysName:
			return "", fmt.Sprintf("it renders {%s}; add both offsets in one adddays", o.Token().Body)
		case o.Token().Kind == grammar.PathRead:
			var leaves []tree.Node
			for _, a := range o.Reads() {
				leaves = append(leaves, a.Leaves()...)
			}
			return dateLayout(leaves, l, seen)
		}
		return "", fmt.Sprintf("it renders %q", n.Format())
	}
	return "", "it renders a table cell or a null"
}
