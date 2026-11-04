package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// checkNameReads refuses a read of a name inside the field bound to it, a read {n} beside {n.w}
// where n's pick renders w twice, and one where n's category reads w through a name of its own.
func checkNameReads(label string, t *template, names resolvedNames) error {
	for _, r := range namedReads(t) {
		b, target := r.a.named, names.targets[r.a.named]
		if b.bindsField() && rendersInside(compilePath(target.start, target.tail).leaves, t) {
			return fmt.Errorf("%s: token {%s}: name %q is read inside %q, the field bound to it; read the name outside that field", label, r.o.Body, b.name, b.ref)
		}
		for _, leaf := range r.a.leaves {
			if err := refuseTwiceDrawn(names.addressed[b], r.a.spelling, leaf, r.a.key()); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", label, r.o.Body, err)
			}
		}
		if err := refuseDrawnApart(r.a, names); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, r.o.Body, err)
		}
	}
	return nil
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
// another read of the name addresses that field. The pick keeps one draw of the field: where n binds
// a category whose format is {w}-{w}, a {n.w} beside {n} cannot say which draw it reads. key is n's
// level under the pick; addressed maps each level a read of the name addresses to that read.
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

// refuseTwiceDrawnIn is refuseTwiceDrawn over the reads of t that land in the pick. It recurses into
// each of them, as render does: beside {n.w.a}, it refuses {n} where n's format reads {w} and w's
// reads {a}-{a}, since "w" and "w.a" are addressed.
func refuseTwiceDrawnIn(addressed map[pickKey]string, read string, t *template, key pickKey) error {
	reads := map[string]int{}
	addressedBy := map[string]string{}
	var into []arm
	for _, o := range t.compiled.ops {
		heads := map[string]bool{} // one op draws once, whichever arm it takes
		for _, a := range slices.Concat(o.arms, o.operands) {
			if by, lands := landsInPick(addressed, key, a); lands {
				into = append(into, a)
				heads[a.head] = true
				addressedBy[a.head] = by
			}
		}
		for head := range heads {
			reads[head]++
		}
	}
	for _, head := range sortedNames(reads) {
		if reads[head] > 1 {
			by := addressedBy[head]
			return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, head, by, read, by)
		}
	}
	for _, a := range into {
		for _, leaf := range a.leaves {
			if err := refuseTwiceDrawn(addressed, read, leaf, key.under(a.key())); err != nil {
				return err
			}
		}
	}
	return nil
}

// refuseDrawnApart refuses read, a read of a name, where what it renders reads a field through a name
// bound to that field and another read of the name passes the field. Each name keeps its own pick:
// where n binds a category whose format is {w as r}{r}, {n} prints r's draw of w and {n.w} n's.
func refuseDrawnApart(read *arm, names resolvedNames) error {
	outer := read.named
	for _, inner := range fieldNamesRead(read.leaves) {
		if inner == outer {
			continue // a read inside the field bound to it, refused where it sits
		}
		level, within := levelOf(names.targetOf(inner).start, names.targets[outer].start)
		if by, passed := names.addressed[outer][level]; within && passed {
			return fmt.Errorf("{%s} reads field %q through name %q, a pick of its own, so {%s} would read another draw; drop {%s} to keep one pick, or write {%s} for a second draw",
				read.spelling, inner.ref, inner.name, by, by, outer.ref+strings.TrimPrefix(by, outer.name))
		}
	}
	return nil
}

// fieldNamesRead is every name bound to a field that rendering nodes reads. It stays in their
// category: a path through a name can't follow a reference, or a name bound to one, out of it.
func fieldNamesRead(nodes []node) []*nameBinding {
	var found []*nameBinding
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
					switch {
					case a.kind == namedRead && a.named.bindsField():
						found = append(found, a.named)
					case a.kind == namedRead || grammar.IsRef(a.head):
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

// levelOf is want's key in a pick of from: the fields leading down to it, if it lies under from.
func levelOf(want, from node) (pickKey, bool) {
	if from == want {
		return "", true
	}
	for _, c := range contained(from) {
		if key, within := levelOf(want, c.node); within {
			return pickKey(join(c.name, string(key))), true
		}
	}
	return "", false
}
