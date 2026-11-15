package names

import (
	"fmt"

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
// pick. It walks target's path from its start, a level at a time, and ends at the first level that
// is no read's address, or that isn't an innerLevel: there it shares a draw unless the level is a
// constantLeaf. An innerLevel that a read lands on, with a draw somewhere below, ends it too.
func sharedDraw(pick pickReads, target tree.NameTarget, above []string) (by string, shares bool) {
	for w := walkFrom(target.Start(), target.Tail()); ; w.step() {
		key := keyOf(above, w.walked())
		by, addressed := pick.addresser(key)
		if !addressed {
			// The first name's pick holds a draw of a level only where a read of that name
			// addresses the level. Such a read addresses every level above it too, so no
			// deeper level is addressed either.
			return "", false
		}
		if kind := w.kind(); kind != innerLevel {
			return by, kind != constantLeaf
		}
		// A read landing here renders every level below, addressed or not.
		if lander, lands := pick.landingOn(key); lands && w.drawsBelow() {
			return lander, true
		}
	}
}
