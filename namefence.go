package fejkdata

import (
	"fmt"
	"slices"
	"sort"
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
		for _, leaf := range a.leaves {
			if err := a.named.checkOnce(a.spelling, leaf, a.path); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
			}
		}
		return nil
	})
}

// checkUses refuses a binding of a category read once whole, and reads of b in two draw groups,
// or inside a repeat, where what b names reads a reference path.
func (b *nameBinding) checkUses() error {
	if r := b.uses[0]; len(b.uses) == 1 && r.tail == "" && !r.nested && len(b.tail) == 0 {
		spelling := b.ref
		if !r.operand {
			spelling = "{" + spelling + "}"
		}
		return fmt.Errorf("name %q is read once, whole, which the bare reference draws the same way; write %s where it is read, and drop the token", b.name, spelling)
	}
	if !readsHeld(b.head, map[node]bool{}) {
		return nil
	}
	for _, u := range b.uses {
		if u.nested {
			return fmt.Errorf("name %q is read inside a repeat, and what it names reads a reference path, which each iteration draws apart; bind the name inside the repeat, or bind a second name there", b.name)
		}
		if u.group != b.uses[0].group {
			return fmt.Errorf("name %q is read in two draw groups, %s and %s, and what it names reads a reference path, which each draw group draws apart; read the name in one draw group, or bind a name in each", b.name, groupSpelling(b.uses[0].group), groupSpelling(u.group))
		}
	}
	return nil
}

func groupSpelling(group string) string {
	if group == "" {
		return "the unnamed one"
	}
	return fmt.Sprintf("%q", group)
}

// readsHeld reports whether rendering n, or any field under it, can read a reference path.
func readsHeld(n node, seen map[node]bool) bool {
	if seen[n] {
		return false
	}
	seen[n] = true
	if t, isTemplate := n.(*template); isTemplate {
		for _, o := range t.compiled.ops {
			for _, a := range slices.Concat(o.arms, o.operands) {
				if a.kind == refPathRead {
					return true
				}
			}
		}
	}
	for _, c := range contained(n) {
		if readsHeld(c.node, seen) {
			return true
		}
	}
	for _, e := range renderEdges(n) {
		if readsHeld(e.to, seen) {
			return true
		}
	}
	return false
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
	fresh := map[string]int{}
	var into []arm
	for _, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if isRef(a.head) || a.kind == namedRead {
				continue
			}
			if _, kept := b.addressed[join(key, a.path)]; kept {
				into = append(into, a)
			}
			if a.kind == freshRead {
				fresh[a.head]++
			}
		}
	}
	heads := make([]string, 0, len(fresh))
	for head := range fresh {
		heads = append(heads, head)
	}
	sort.Strings(heads)
	for _, head := range heads {
		if by, kept := b.addressed[join(key, head)]; kept && fresh[head] > 1 {
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
