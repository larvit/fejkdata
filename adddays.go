package fejkdata

import (
	"fmt"
	"math"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

func addDaysLayout(d arm, l *links) (string, error) {
	layout, why := dateLayout(d.leaves, l, map[node]bool{})
	switch {
	case why != "":
		return "", fmt.Errorf("%s is not proven a date: %s; read a field whose format is one date(from,to,'layout') or born(min,max,'layout')", d.spelling, why)
	case !builtinfunc.ReadsBack(layout):
		return "", fmt.Errorf("%s prints in '%s', which does not read back as the day it printed; give it a layout with a four-digit year, a month and a day, such as '2006-01-02'", d.spelling, layout)
	}
	return layout, nil
}

// proveAddDays proves the day count of each adddays of t that reads one is a whole number within
// the cap.
func proveAddDays(label string, t *template, l *links) error {
	for _, o := range l.of(t).compiled.ops {
		if o.Kind != grammar.BuiltinCall || o.Fn != builtinfunc.AddDaysName || len(o.operands) == 1 {
			continue
		}
		if err := proveDayCount(o.operands[1], l); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, o.Body, err)
		}
	}
	return nil
}

func proveDayCount(n arm, l *links) error {
	v := (&valueProof{links: l}).proveArm(n)
	switch {
	case v.NotOperand != "":
		return fmt.Errorf("%s is not always a whole number of days: %s; write an integer or read one {int(min,max)}", n.spelling, v.NotOperand)
	case !v.Integral:
		return fmt.Errorf("%s is not always a whole number of days: it may print a fraction; read an {int(min,max)}", n.spelling)
	case math.Max(math.Abs(v.Lo), math.Abs(v.Hi)) > builtinfunc.MaxDays:
		return fmt.Errorf("%s may reach %.0f, which exceeds the maximum of %d days either way; narrow it", n.spelling, math.Max(math.Abs(v.Lo), math.Abs(v.Hi)), builtinfunc.MaxDays)
	}
	return nil
}

// dateLayout is the one layout every render of nodes prints a date in, or why no proof holds.
func dateLayout(nodes []node, l *links, seen map[node]bool) (layout, why string) {
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

func nodeDateLayout(n node, l *links, seen map[node]bool) (string, string) {
	if seen[n] {
		return "", "it reads itself"
	}
	seen[n] = true
	defer delete(seen, n)
	switch n := n.(type) {
	case *choice:
		return dateLayout(n.items, l, seen)
	case *template:
		ops := l.of(n).compiled.ops
		if n.repeat != 1 || len(n.tokens) != 1 || len(ops) != 1 {
			return "", fmt.Sprintf("it renders %q", n.format)
		}
		o := ops[0]
		if layout, isDate := builtinfunc.DateLayout(o.Fn, o.Args); isDate {
			return layout, ""
		}
		switch {
		case o.Fn == builtinfunc.AddDaysName:
			return "", fmt.Sprintf("it renders {%s}; add both offsets in one adddays", o.Body)
		case o.Kind == grammar.PathRead:
			var leaves []node
			for _, a := range o.arms {
				leaves = append(leaves, a.leaves...)
			}
			return dateLayout(leaves, l, seen)
		}
		return "", fmt.Sprintf("it renders %q", n.format)
	}
	return "", "it renders a table cell or a null"
}
