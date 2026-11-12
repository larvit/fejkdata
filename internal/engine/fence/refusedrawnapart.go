package fence

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// refuseDrawnApart refuses a read of a name where what it renders reads a field through a name
// bound to that field and another read of the name shares a draw with it. Each name keeps its own
// pick: where n binds a category whose format is {w as r}{r}, {n} prints r's draw of w and {n.w}
// n's; a field that only reads fields of a template that isn't a choice, or a constant, shares none.
func refuseDrawnApart(nr nameRead) error {
	outer := nr.read.Named()
	for _, r := range nr.routes {
		inner := r.a.Named()
		if r.a.Kind() != tree.NamedRead || !inner.BindsField() || inner == outer { // RefuseReadsInsideFields refused a name's own read inside its field
			continue
		}
		target := inner.Target()
		above, inside := nr.names.pathTo(outer.Target().Start(), target.Start())
		if !inside {
			continue
		}
		if by, shares := sharedDraw(nr.pick, target, above); shares {
			return fmt.Errorf("{%s} reads field %q through name %q, a pick of its own, so {%s} would read another draw; drop {%s} to keep one pick, or write {%s} for a second draw",
				nr.read.Spelling(), inner.Ref(), inner.Name(), by, by, secondDraw(outer, by))
		}
	}
	return nil
}

// sharedDraw reports whether a read of one name shares, with a second name bound to target, a draw
// that could differ, and returns that read. Parameter above is target's path under the first name's
// pick. It walks target's path from its start, a level at a time, and the first name's pick holds
// a draw of a level only while a read of that name addresses the level. Each level is settled by
// the first case that fits, and each case rests on the ones before it not fitting.
func sharedDraw(pick pickReads, target tree.NameTarget, above []string) (by string, shares bool) {
	n := target.Start()
	for i := 0; ; i++ {
		key := keyOf(above, target.Tail()[:i])
		by, here := pick.addressed[key]
		switch {
		case !here:
			// First, since by names the read every later case reports. A read that addresses a
			// level addresses every level above it, so no deeper level is addressed either.
			return "", false
		case isDraw(n):
			// Before the leaf case: a choice or a table is a draw wherever on the path it sits.
			return by, true
		case i == len(target.Tail()):
			// The leaf is no choice or table here, so only a constant renders the same twice.
			return by, !isConstant(n)
		}
		// Before the step down: a read landing here renders all below it, whether or not a read
		// addresses the next level, where the first case would otherwise say they share none.
		if lander, lands := pick.landingOn(key); lands && drawsAlong(n, target.Tail()[i:], true) {
			return lander, true
		}
		var err error
		if n, err = load.ChildNamed(n, target.Tail()[i]); err != nil {
			// A step no field spells, such as "..", can't be followed: they count as sharing a draw.
			return by, true
		}
	}
}
