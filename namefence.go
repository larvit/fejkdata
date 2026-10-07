package fejkdata

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/grammar"
)

// checkNameReads refuses a read of a name inside the field bound to it, and a read {n} beside {n.w}
// where n's pick renders w twice.
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
	}
	return nil
}

// rendersInside reports whether t sits in what one of nodes contains.
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

// refuseTwiceDrawnIn is refuseTwiceDrawn over t's own reads of its fields; a reference or a read
// through a name draws apart from the pick. It recurses into each read whose key under key is
// addressed: beside {n.w.a}, it refuses {n} where n's format reads {w} and w's reads {a}-{a},
// since "w" and "w.a" are addressed.
func refuseTwiceDrawnIn(addressed map[pickKey]string, read string, t *template, key pickKey) error {
	reads := map[string]int{}
	var into []arm
	for _, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if grammar.IsRef(a.head) || a.kind == namedRead {
				continue
			}
			if _, kept := addressed[key.under(a.key())]; kept {
				into = append(into, a)
			}
			reads[a.head]++
		}
	}
	for _, head := range sortedNames(reads) {
		if by, kept := addressed[key.under(pickKey(head))]; kept && reads[head] > 1 {
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
