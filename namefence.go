package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkNameReads refuses each binding of t that refuseSingleRead refuses, a read of a name inside
// the field bound to it, and a read {n} beside {n.w} where n's pick renders w twice.
func checkNameReads(label string, t *template, names resolvedNames) error {
	for _, tok := range t.tokens {
		if tok.Kind != grammar.NameBind {
			continue
		}
		b := t.nameScope.bindings[tok.Bound]
		if err := b.refuseSingleRead(names.uses[b]); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, tok.Body, err)
		}
	}
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

// refuseSingleRead refuses b when its one read sits in b's own scope, not in a repeat nested inside
// it, at a spot where the bound spelling could stand. That spelling draws the same without the
// name: write {/word} for {/word as w}{w}. A bound field can stand in only where the reading
// template reaches the binder through fields. A calc operand cannot read a reference or a path.
// Where the spelling would be a whole CLI argument or a whole struct tag, that entry point refuses
// it and names the bare path.
func (b *nameBinding) refuseSingleRead(uses []nameUse) error {
	if len(uses) == 0 {
		panic(invariant.Broken("name %q has no read at resolve, though its compile found one", b.name))
	}
	r := uses[0]
	if len(uses) > 1 || r.nested {
		return nil
	}
	spelling := b.ref
	if b.bindsField() {
		down, reaches := fieldPathTo(r.in, b.binder)
		if !reaches {
			return nil
		}
		spelling = grammar.JoinSegments(append(down, b.ref))
	}
	switch {
	case strings.HasPrefix(r.tail, ".."):
		spelling += r.tail
	case r.tail != "":
		spelling += "." + r.tail
	}
	if r.noRef && !calcReads(spelling) {
		return nil
	}
	if !r.operand {
		spelling = "{" + spelling + "}"
	}
	return fmt.Errorf("name %q is read once, so it keeps no pick for another read; write %s where it is read, and drop the token", b.name, spelling)
}

// fieldPathTo is the fields leading from t down to the template target, if t reaches it so.
func fieldPathTo(t, target *template) ([]string, bool) {
	if t == target {
		return nil, true
	}
	for _, name := range sortedNames(t.fields) {
		if sub, isTemplate := t.fields[name].(*template); isTemplate {
			if rest, reaches := fieldPathTo(sub, target); reaches {
				return append([]string{name}, rest...), true
			}
		}
	}
	return nil, false
}

// calcReads reports whether a calc reads spelling as one operand.
func calcReads(spelling string) bool {
	c, err := grammar.ParseCalc(spelling)
	v, isVar := c.Expr.(grammar.CalcVar)
	return err == nil && isVar && v.Name == spelling
}

// refuseTwiceDrawn refuses read, a read of a name landing on n at key under the name's pick, where n
// renders a field twice and another read of the name addresses that field: the pick keeps one draw
// of the field, so beside {n} for a category {w}-{w}, {n.w} cannot say which draw it reads.
// addressed maps each level a read of the name addresses to that read.
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
// through a name draws apart from the pick.
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
