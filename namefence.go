package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// checkNameReads refuses a read of a name that would disagree with another read of it, or that
// steps over a frame a name needs: a read inside the field bound to the name, a read {n} beside
// {n.w} where n's pick renders w twice, one where n's category reads w through a name of its own or
// through a reference, and one that steps into a choice's item without rendering it.
func checkNameReads(label string, t *template, names resolvedNames) error {
	for _, r := range namedReads(t) {
		if err := checkNameRead(t, r.a, names); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, r.o.Body, err)
		}
	}
	return nil
}

func checkNameRead(t *template, read *arm, names resolvedNames) error {
	b := read.named
	if b.bindsField() && rendersInside(compilePath(names.targets[b].start, names.targets[b].tail).leaves, t) {
		return fmt.Errorf("name %q is read inside %q, the field bound to it; read the name outside that field", b.name, b.ref)
	}
	for _, leaf := range read.leaves {
		if err := refuseTwiceDrawn(names.addressed[b], read.spelling, leaf, read.key()); err != nil {
			return err
		}
	}
	below := armsRead(read.leaves)
	if err := refuseDrawnApart(read, below, names); err != nil {
		return err
	}
	if err := refuseSkippedItem(read, below, names); err != nil {
		return err
	}
	return refuseFreshDraw(read, below, names)
}

func rendersInside(nodes []node, t *template) bool {
	seen := map[node]bool{}
	var inside func(n node) bool
	inside = func(n node) bool {
		if n == node(t) {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		for _, c := range contained(n) {
			if inside(c.node) {
				return true
			}
		}
		return false
	}
	for _, n := range nodes {
		if inside(n) {
			return true
		}
	}
	return false
}

// refuseTwiceDrawn refuses read, a read of a name that lands on n, where n renders a field twice and
// another read of the name addresses a draw those two renders share. The pick keeps one draw of the
// field: where n binds a category whose format is {w}-{w}, a {n.w} beside {n} cannot say which draw
// it reads. key is n's level under the pick; addressed maps each level a read of the name addresses
// to that read.
func refuseTwiceDrawn(addressed map[pickKey]string, read string, n node, key pickKey) error {
	switch n := n.(type) {
	case *choice:
		for _, item := range n.items {
			if err := refuseTwiceDrawn(addressed, read, item, key); err != nil {
				return err
			}
		}
	case *template:
		return refuseTwiceDrawnIn(addressed, read, n, key)
	}
	return nil
}

// landing is a read of t the pick keeps, by the read of the name that addresses it, in the op at
// index op.
type landing struct {
	a  arm
	by string
	op int
}

// refuseTwiceDrawnIn is refuseTwiceDrawn over the reads of t that land in the pick. It recurses into
// each of them, as render does: beside {n.w.a}, it refuses {n} where n's format reads {w} and w's
// reads {a}-{a}, since "w" and "w.a" are addressed. One op draws once, whichever arm it takes, so
// only reads in two ops can draw apart.
func refuseTwiceDrawnIn(addressed map[pickKey]string, read string, t *template, key pickKey) error {
	var into []landing
	for i, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if by, lands := landsInPick(addressed, key, a); lands {
				into = append(into, landing{a, by, i})
			}
		}
	}
	for i, x := range into {
		for _, y := range into[:i] {
			if x.op != y.op && x.a.head == y.a.head && readTwiceApart(t, x.a, y.a) {
				return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, x.a.head, x.by, read, x.by)
			}
		}
	}
	for _, l := range into {
		for _, leaf := range l.a.leaves {
			if err := refuseTwiceDrawn(addressed, read, leaf, key.under(l.a.key())); err != nil {
				return err
			}
		}
	}
	return nil
}

// readTwiceApart reports whether the pick keeping x and y would make them read one draw that drawn
// apart they would read two: they pass a choice or a table together, or they read one leaf that is
// no constant. Two fields of one plain template share nothing that draws.
func readTwiceApart(t *template, x, y arm) bool {
	common := 0
	for common < min(len(x.tail), len(y.tail)) && x.tail[common] == y.tail[common] {
		common++
	}
	return drawsAlong(t.startOf(x.head), x.tail[:common], x.key() == y.key())
}

// drawsAlong reports whether the walk of path from start meets a draw: a choice or a table at one of
// its levels, or, with leaf, a last level that is no constant.
func drawsAlong(start node, path []string, leaf bool) bool {
	n := start
	for i := 0; ; i++ {
		switch {
		case isDraw(n):
			return true
		case i == len(path):
			return leaf && !isConstant(n)
		}
		var err error
		if n, err = childNamed(n, path[i]); err != nil {
			return true
		}
	}
}

// isDraw reports whether n draws a variant or a row, which a pick keeps and a fresh read does not.
func isDraw(n node) bool {
	switch n.(type) {
	case *choice, *table:
		return true
	}
	return false
}

// isConstant reports whether every render of n is the same text.
func isConstant(n node) bool {
	t, isTemplate := n.(*template)
	if !isTemplate {
		return false
	}
	_, fixed := t.fixedText()
	return fixed
}

// refuseDrawnApart refuses read, a read of a name, where what it renders reads a field through a name
// bound to that field and another read of the name shares a draw with it. Each name keeps its own
// pick: where n binds a category whose format is {w as r}{r}, {n} prints r's draw of w and {n.w}
// n's; a field that only reads fields of a plain template, or a constant, shares none.
func refuseDrawnApart(read *arm, below []arm, names resolvedNames) error {
	outer := read.named
	for _, x := range below {
		inner := x.named
		if x.kind != namedRead || !inner.bindsField() || inner == outer {
			continue // a read inside the field bound to it is refused where it sits
		}
		target := names.targetOf(inner)
		above, inside := names.pathTo(names.targets[outer].start, target.start)
		if !inside {
			continue
		}
		if by, shares := sharedDraw(names.reads[outer], names.addressed[outer], target, above); shares {
			return fmt.Errorf("{%s} reads field %q through name %q, a pick of its own, so {%s} would read another draw; drop {%s} to keep one pick, or write {%s} for a second draw",
				read.spelling, inner.ref, inner.name, by, by, secondDraw(outer, by))
		}
	}
	return nil
}

// sharedDraw reports whether a name's pick, with reads and the keys addressed, and the pick of a name
// bound to target, which lies at above under the first, share a draw that could differ, and the read
// that shares it. They share one where a read of the first passes a choice or a table the second
// reads through, lands on a level above the second's leaf with a draw below it, or addresses a
// leaf that is no constant.
func sharedDraw(reads []*arm, addressed map[pickKey]string, target nameTarget, above []string) (by string, shares bool) {
	n := target.start
	for i := 0; ; i++ {
		key := pickKey(strings.Join(slices.Concat(above, target.tail[:i]), "."))
		by, here := addressed[key]
		switch {
		case !here:
			return "", false
		case isDraw(n):
			return by, true
		case i == len(target.tail):
			return by, !isConstant(n)
		}
		if lander, lands := landingOn(reads, key); lands && drawsAlong(n, target.tail[i:], true) {
			return lander, true
		}
		var err error
		if n, err = childNamed(n, target.tail[i]); err != nil {
			return by, true
		}
	}
}

// landingOn is the spelling of a read that lands on key.
func landingOn(reads []*arm, key pickKey) (string, bool) {
	for _, r := range reads {
		if r.key() == key {
			return r.spelling, true
		}
	}
	return "", false
}

// secondDraw is how to spell by, a read of b, so it draws afresh.
func secondDraw(b *nameBinding, by string) string { return b.ref + strings.TrimPrefix(by, b.name) }

// refuseSkippedItem refuses read, a read through a name bound to a field, where it steps into an item
// of a choice and what it renders reads a name that item binds. The read opens no frame for the
// item, so the name has no pick there.
func refuseSkippedItem(read *arm, below []arm, names resolvedNames) error {
	if !read.named.bindsField() {
		return nil
	}
	for _, x := range below {
		if x.kind != namedRead || x.named.scope.kind() != scopeChoiceItem {
			continue
		}
		for _, leaf := range read.leaves {
			if _, steppedInto := names.pathTo(x.named.scope.owner, leaf); steppedInto && leaf != x.named.scope.owner {
				return fmt.Errorf("{%s} steps into an item of a choice without rendering it, so name %q, bound in that item, has no pick there; bind %q outside the choice, or read the item whole",
					read.spelling, x.named.name, x.named.name)
			}
		}
	}
	return nil
}

// refuseFreshDraw refuses read, a read of a name, where what it renders reads through a reference a
// path another read of the name reads. A reference draws afresh, so {n} and {n.ab.name} would differ
// where n's format reads {/cat.ab.name}; a constant draws nothing to differ.
func refuseFreshDraw(read *arm, below []arm, names resolvedNames) error {
	b := read.named
	start, tail := names.targets[b].start, len(names.targets[b].tail)
	for _, x := range below {
		if !grammar.IsRef(x.head) {
			continue
		}
		for _, leaf := range x.leaves {
			path, within := names.pathTo(start, leaf)
			if !within || len(path) == 0 || isConstant(leaf) {
				continue
			}
			if by, other := readAddressing(names.reads[b], read, tail, pickKey(strings.Join(path, "."))); other {
				return fmt.Errorf("{%s} renders {%s}, which draws afresh a path {%s} reads from the pick, so the two would differ; drop {%s}, or write {%s} for a second draw",
					read.spelling, x.spelling, by, by, secondDraw(b, by))
			}
		}
	}
	return nil
}

// readAddressing is the spelling of a read other than self that lands on or passes key, where tail
// is how many segments of its path the name's binding holds.
func readAddressing(reads []*arm, self *arm, tail int, key pickKey) (string, bool) {
	for _, r := range reads {
		if r != self && slices.Contains(r.levels[tail:], key) {
			return r.spelling, true
		}
	}
	return "", false
}

// armsRead is every arm that rendering nodes read, down every variant of a choice and every field
// a read renders. It stays in their category: a reference's or a reference-bound name's arm is
// counted, but what it renders is not followed, since it draws afresh.
func armsRead(nodes []node) []arm {
	var found []arm
	seen := map[node]bool{}
	var read func(n node)
	read = func(n node) {
		if seen[n] {
			return
		}
		seen[n] = true
		switch n := n.(type) {
		case *choice:
			for _, item := range n.items {
				read(item)
			}
		case *template:
			for _, o := range n.compiled.ops {
				for _, a := range slices.Concat(o.arms, o.operands) {
					found = append(found, a)
					if a.kind == namedRead && !a.named.bindsField() || a.kind == freshRead && grammar.IsRef(a.head) {
						continue
					}
					for _, leaf := range a.leaves {
						read(leaf)
					}
				}
			}
		}
	}
	for _, n := range nodes {
		read(n)
	}
	return found
}
