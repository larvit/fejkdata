package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkNameReads refuses each binding of t that checkUses refuses, a read of a name inside the
// field bound to it, and a read {n} beside {n.w} where n's pick renders w twice: the pick keeps one
// draw of w.
func checkNameReads(label string, t *template, names linkedNames) error {
	for _, tok := range t.tokens {
		if tok.Kind != grammar.NameBind {
			continue
		}
		b := t.nameScope.bindings[tok.Bound]
		if err := b.checkUses(names.uses[b]); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, tok.Body, err)
		}
	}
	for _, r := range namedReads(t) {
		b, target := r.a.named, names.targets[r.a.named]
		if b.bindsField() && rendersInside(compilePath(target.head, target.tail).leaves, t) {
			return fmt.Errorf("%s: token {%s}: name %q is read inside %q, the field bound to it; read the name outside that field", label, r.o.Body, b.name, b.ref)
		}
		for _, leaf := range r.a.leaves {
			if err := checkOnce(names.addressed[b], r.a.spelling, leaf, r.a.path); err != nil {
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

// checkUses refuses a binding whose uses read it once at a spot the bound spelling can stand:
// that spelling draws the same way without the name. A bound field can stand in only where the
// reading template reaches the binder through fields. Where the spelling would be a CLI argument
// or tag of one reference alone, that entry point's own refusal then names the bare path.
func (b *nameBinding) checkUses(uses []nameUse) error {
	if len(uses) == 0 {
		panic(invariant.Broken("name %q has no read at link, though its compile found one", b.name))
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

// checkOnce walks n, rendering at key under a pick, into each level a read of the name addresses:
// the keys of addressed.
func checkOnce(addressed map[string]string, read string, n node, key string) error {
	switch n := n.(type) {
	case *choice:
		for _, item := range n.items {
			if err := checkOnce(addressed, read, item, key); err != nil {
				return err
			}
		}
	case *template:
		return checkTemplateOnce(addressed, read, n, key)
	}
	return nil
}

func checkTemplateOnce(addressed map[string]string, read string, t *template, key string) error {
	reads := map[string]int{}
	var into []arm
	for _, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if grammar.IsRef(a.head) || a.kind == namedRead {
				continue
			}
			if _, kept := addressed[join(key, a.path)]; kept {
				into = append(into, a)
			}
			reads[a.head]++
		}
	}
	for _, head := range sortedNames(reads) {
		if by, kept := addressed[join(key, head)]; kept && reads[head] > 1 {
			return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, head, by, read, by)
		}
	}
	for _, a := range into {
		for _, leaf := range a.leaves {
			if err := checkOnce(addressed, read, leaf, join(key, a.path)); err != nil {
				return err
			}
		}
	}
	return nil
}
