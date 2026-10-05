package fejkdata

import (
	"fmt"
	"slices"
	"strings"
)

// checkNameReads refuses each binding of t that checkUses refuses, and a read such as {n} whose
// pick renders twice a field another read, {n.w}, reads once: the pick keeps one draw of it.
func checkNameReads(path string, t *template) error {
	for _, tok := range t.tokens {
		if tok.kind != nameBind {
			continue
		}
		if err := t.nameScope.bindings[tok.bound].checkUses(); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), tok.body, err)
		}
	}
	return namedReads(t, func(o *op, a *arm) error {
		if a.named.bindsField() && rendersInside(compilePath(a.named.head, a.named.tail).leaves, t) {
			return fmt.Errorf("%s: token {%s}: name %q is read inside %q, the field it binds; read the name outside that field", t.site.label(path), o.body, a.named.name, a.named.ref)
		}
		for _, leaf := range a.leaves {
			if err := a.named.checkOnce(a.spelling, leaf, a.path); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
			}
		}
		return nil
	})
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

// checkUses refuses a binding read once at a spot the bound spelling can stand: that spelling
// draws the same way without the name. A field stands where the reading template reaches the
// binder through fields. Where the spelling would be a CLI argument or tag of one reference alone,
// that entry point's own refusal then names the bare path.
func (b *nameBinding) checkUses() error {
	r := b.uses[0]
	if len(b.uses) > 1 || r.nested {
		return nil
	}
	spelling := b.ref
	if b.bindsField() {
		down, reaches := fieldPathTo(r.in, b.binder)
		if !reaches {
			return nil
		}
		spelling = joinSegments(append(down, b.ref))
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
	n, err := parseCalc(spelling)
	v, isVar := n.(calcVar)
	return err == nil && isVar && string(v) == spelling
}

// checkOnce walks n, rendering at key under b's pick, into each level a read of b addresses.
func (b *nameBinding) checkOnce(read string, n node, key string) error {
	switch n := n.(type) {
	case *choice:
		for _, item := range n.items {
			if err := b.checkOnce(read, item, key); err != nil {
				return err
			}
		}
	case *template:
		return b.checkTemplateOnce(read, n, key)
	}
	return nil
}

func (b *nameBinding) checkTemplateOnce(read string, t *template, key string) error {
	reads := map[string]int{}
	var into []arm
	for _, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if isRef(a.head) || a.kind == namedRead {
				continue
			}
			if _, kept := b.addressed[join(key, a.path)]; kept {
				into = append(into, a)
			}
			reads[a.head]++
		}
	}
	for _, head := range sortedNames(reads) {
		if by, kept := b.addressed[join(key, head)]; kept && reads[head] > 1 {
			return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, head, by, read, by)
		}
	}
	for _, a := range into {
		for _, leaf := range a.leaves {
			if err := b.checkOnce(read, leaf, join(key, a.path)); err != nil {
				return err
			}
		}
	}
	return nil
}
