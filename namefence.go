package fejkdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkNameReads refuses a read of a name that could disagree with another read of it, or that reads
// a name of a choice's item it never renders.
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
	pick := pickReads{reads: names.reads[b], addressed: names.addressed[b]}
	for _, leaf := range read.leaves {
		if err := refuseTwiceDrawn(pick, read.spelling, leaf, read.key()); err != nil {
			return err
		}
	}
	below := armsRead(read.leaves)
	if err := refuseDrawnApart(read, below, names); err != nil {
		return err
	}
	if err := refuseSkippedItem(t, read, below, names); err != nil {
		return err
	}
	if err := refuseFreshDraw(read, below, names); err != nil {
		return err
	}
	return refuseKeptAndFreshField(read, names)
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

// pickReads is the reads of one name: the arms that read it, and the keys they address, each with the
// spelling of the first read reaching it.
type pickReads struct {
	reads     []*arm
	addressed map[pickKey]string
}

// landingOn is the spelling of a read that lands on key.
func (p pickReads) landingOn(key pickKey) (string, bool) {
	for _, r := range p.reads {
		if r.key() == key {
			return r.spelling, true
		}
	}
	return "", false
}

// refuseTwiceDrawn refuses read, a read of a name that lands on n, where n renders a field twice and
// another read of the name addresses a draw those two renders share. The pick keeps one draw of the
// field: where n binds a category whose format is {w}-{w}, a {n.w} beside {n} cannot say which draw
// it reads. key is n's level under the pick.
func refuseTwiceDrawn(pick pickReads, read string, n node, key pickKey) error {
	switch n := n.(type) {
	case *choice:
		for _, item := range n.items {
			if err := refuseTwiceDrawn(pick, read, item, key); err != nil {
				return err
			}
		}
	case *template:
		return refuseTwiceDrawnIn(pick, read, n, key)
	}
	return nil
}

// landing is an arm of a template that lands in the pick: by is the read of the name addressing it,
// and op the index of the op holding it.
type landing struct {
	a  arm
	by string
	op int
}

// refuseTwiceDrawnIn is refuseTwiceDrawn over the reads of t that land in the pick. It recurses into
// each of them, as render does: beside {n.w.a}, it refuses {n} where n's format reads {w} and w's
// reads {a}-{a}, since "w" and "w.a" are addressed. One op draws once, whichever arm it takes, so
// only reads in two ops can draw apart.
func refuseTwiceDrawnIn(pick pickReads, read string, t *template, key pickKey) error {
	var into []landing
	for i, o := range t.compiled.ops {
		for _, a := range slices.Concat(o.arms, o.operands) {
			if by, lands := landsInPick(pick.addressed, key, a); lands {
				into = append(into, landing{a, by, i})
			}
		}
	}
	for i, x := range into {
		for _, y := range into[:i] {
			if x.op != y.op && x.a.head == y.a.head && readTwiceApart(pick, key, t, x.a, y.a) {
				return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, x.a.head, x.by, read, x.by)
			}
		}
	}
	for _, l := range into {
		for _, leaf := range l.a.leaves {
			if err := refuseTwiceDrawn(pick, read, leaf, key.under(l.a.key())); err != nil {
				return err
			}
		}
	}
	return nil
}

// readTwiceApart reports whether x and y, two reads of t that the pick keeps, pass one choice or table, or
// read one leaf that isn't a constant, or share a level a read of the name lands on, where a read of
// the whole level draws what the two keep. Drawn apart, such reads could differ. Two fields of a
// template that isn't a choice share none of these. key is t's level under the pick.
func readTwiceApart(pick pickReads, key pickKey, t *template, x, y arm) bool {
	common := 0
	for common < min(len(x.tail), len(y.tail)) && x.tail[common] == y.tail[common] {
		common++
	}
	_, landed := pick.landingOn(key.under(x.levels[common]))
	return drawsAlong(t.startOf(x.head), x.tail[:common], x.key() == y.key() || landed)
}

// drawsAlong reports whether the walk of path from start meets a choice or a table at one of its
// levels, or, with leaf, ends on a level that isn't a constant.
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
// n's; a field that only reads fields of a template that isn't a choice, or a constant, shares none.
func refuseDrawnApart(read *arm, below []arm, names resolvedNames) error {
	outer := read.named
	for _, x := range below {
		inner := x.named
		if x.kind != namedRead || !inner.bindsField() || inner == outer { // a name's own read inside its field is refused where it sits
			continue
		}
		target := names.targetOf(inner)
		above, inside := names.pathTo(names.targets[outer].start, target.start)
		if !inside {
			continue
		}
		if by, shares := sharedDraw(pickReads{reads: names.reads[outer], addressed: names.addressed[outer]}, target, above); shares {
			return fmt.Errorf("{%s} reads field %q through name %q, a pick of its own, so {%s} would read another draw; drop {%s} to keep one pick, or write {%s} for a second draw",
				read.spelling, inner.ref, inner.name, by, by, secondDraw(outer, by))
		}
	}
	return nil
}

// sharedDraw reports whether the reads of one name share a draw that could differ with a name bound
// to target, which sits at the path above under the first name's pick, and returns the read that
// shares it. They share one where a read of the first name passes a choice or table on target's path, lands
// above target's leaf with a choice below it, or lands on a leaf that isn't a constant.
func sharedDraw(pick pickReads, target nameTarget, above []string) (by string, shares bool) {
	n := target.start
	for i := 0; ; i++ {
		key := pickKey(strings.Join(slices.Concat(above, target.tail[:i]), "."))
		by, here := pick.addressed[key]
		switch {
		case !here:
			return "", false
		case isDraw(n):
			return by, true
		case i == len(target.tail):
			return by, !isConstant(n)
		}
		if lander, lands := pick.landingOn(key); lands && drawsAlong(n, target.tail[i:], true) {
			return lander, true
		}
		var err error
		if n, err = childNamed(n, target.tail[i]); err != nil {
			return by, true
		}
	}
}

// secondDraw is how to spell by, a read of b, so it draws afresh.
func secondDraw(b *nameBinding, by string) string { return b.ref + strings.TrimPrefix(by, b.name) }

// refuseSkippedItem refuses read, a read through a name bound to a field of t, where it steps into an
// item of a choice and what it renders reads a name that item binds. The read opens no frame for the
// item, so the name has no pick there, unless t lies in the item and its frame is already open.
func refuseSkippedItem(t *template, read *arm, below []arm, names resolvedNames) error {
	if !read.named.bindsField() {
		return nil
	}
	for _, x := range below {
		if x.kind != namedRead {
			continue
		}
		scope := x.named.scope
		if scope.kind() != scopeChoiceItem || t.nameScope.within(scope) {
			continue
		}
		for _, leaf := range read.leaves {
			if _, steppedInto := names.pathTo(scope.owner, leaf); steppedInto {
				return fmt.Errorf("{%s} steps into an item of a choice without rendering it, so name %q, bound in that item, has no pick there; bind %q outside the choice, or read {%s} whole",
					read.spelling, x.named.name, x.named.name, read.named.name)
			}
		}
	}
	return nil
}

// refuseFreshDraw refuses read, a read of a name, where what it renders holds a reference that
// reaches a path another read of the name lands on, passes or renders, through other references too.
// A reference draws afresh, so {n} and {n.ab.name} could differ where n's format reads
// {/cat.ab.name}, and so could {n} and {n.ab} where ab's format reads {name}; a reference to a
// constant can't differ.
func refuseFreshDraw(read *arm, below []arm, names resolvedNames) error {
	b := read.named
	target := names.targets[b]
	start, tail := target.start, len(target.tail)
	seen := map[node]bool{}
	for queue := slices.Clone(below); len(queue) > 0; queue = queue[1:] {
		x := queue[0]
		if !grammar.IsRef(x.head) {
			continue
		}
		for _, leaf := range x.leaves {
			if seen[leaf] {
				continue
			}
			seen[leaf] = true
			queue = append(queue, armsRead([]node{leaf})...)
			path, within := names.pathTo(start, leaf)
			if !within || len(path) == 0 || !drawsAlong(start, path, true) {
				continue
			}
			if tail > 0 && grammar.IsSelector(target.tail[0]) {
				path = slices.Concat(target.tail[:1], path)
			}
			if by, other := readAddressing(names, names.reads[b], read, tail, pickKey(strings.Join(path, ".")), leaf); other {
				return fmt.Errorf("{%s} reaches {%s}, which draws afresh a path {%s} reads from the pick, so {%s} and {%s} could differ; drop {%s}, or write {%s} for a second draw",
					read.spelling, x.spelling, by, read.spelling, by, by, secondDraw(b, by))
			}
		}
	}
	return nil
}

// fieldRead is a read of a field met while of, a read of a name, renders. a is the read and in the
// template holding it; levels is a's levels moved under in's key in the pick. kept says whether the
// pick keeps a's value, as landsInPick reports; where it doesn't, a draws afresh.
type fieldRead struct {
	a      arm
	in     *template
	levels []pickKey
	kept   bool
	of     *arm
}

// fieldReadsUnder is every field read made while read renders, down every variant of a choice and
// every field a kept read renders. A reference draws afresh whatever its pick, and a read through a
// name has a pick of its own, so neither is listed.
func fieldReadsUnder(read *arm, addressed map[pickKey]string) []fieldRead {
	type site struct {
		n   node
		key pickKey
	}
	var found []fieldRead
	seen := map[site]bool{}
	var walk func(n node, key pickKey)
	walk = func(n node, key pickKey) {
		if seen[site{n, key}] {
			return
		}
		seen[site{n, key}] = true
		switch n := n.(type) {
		case *choice:
			for _, item := range n.items {
				walk(item, key)
			}
		case *template:
			for _, o := range n.compiled.ops {
				for _, a := range slices.Concat(o.arms, o.operands) {
					if a.kind == namedRead || grammar.IsRef(a.head) {
						continue
					}
					_, kept := landsInPick(addressed, key, a)
					levels := make([]pickKey, len(a.levels))
					for i, l := range a.levels {
						levels[i] = key.under(l)
					}
					found = append(found, fieldRead{a: a, in: n, levels: levels, kept: kept, of: read})
					if kept {
						for _, leaf := range a.leaves {
							walk(leaf, key.under(a.key()))
						}
					}
				}
			}
		}
	}
	for _, leaf := range read.leaves {
		walk(leaf, read.key())
	}
	return found
}

// freshFieldReads is the reads of a field that rendering a read of b makes afresh, listed under each
// level they read. Reads of one key render the same, so it walks each key once.
func (names resolvedNames) freshFieldReads(b *nameBinding) map[pickKey][]fieldRead {
	if byLevel, built := names.fresh[b]; built {
		return byLevel
	}
	byLevel := map[pickKey][]fieldRead{}
	walked := map[pickKey]bool{}
	for _, read := range names.reads[b] {
		if walked[read.key()] {
			continue
		}
		walked[read.key()] = true
		for _, u := range fieldReadsUnder(read, names.addressed[b]) {
			if u.kept {
				continue
			}
			for _, level := range u.levels {
				byLevel[level] = append(byLevel[level], u)
			}
		}
	}
	names.fresh[b] = byLevel
	return byLevel
}

// refuseKeptAndFreshField refuses read, a read of a name, where it reads a field from the pick that
// another read of the name draws afresh, and the two could differ: with {n} and {n.w}, where n's
// format is {w.a} and w's is {a}, {n} reads the pick's w.a and {n.w} draws w's a afresh.
func refuseKeptAndFreshField(read *arm, names resolvedNames) error {
	b := read.named
	byLevel := names.freshFieldReads(b)
	for _, x := range fieldReadsUnder(read, names.addressed[b]) {
		if !x.kept {
			continue
		}
		for _, level := range x.levels {
			for _, u := range byLevel[level] {
				if level, apart := readsFieldApart(x, u); apart {
					return fmt.Errorf("{%s} reads {%s} from the pick, but {%s} renders {%s}, the same field, afresh, so {%s} and {%s} could differ; read {%s} in place of {%s} to keep one draw, or write {%s} for a second draw",
						read.spelling, x.a.spelling, u.of.spelling, u.a.spelling, read.spelling, u.of.spelling, spellingThrough(b, names.targetOf(b), level), u.of.spelling, secondDraw(b, u.of.spelling))
				}
			}
		}
	}
	return nil
}

// readsFieldApart reports whether x, which the pick keeps, and u, which draws afresh, read a field
// they could draw apart: they share a level, and the walk to the deepest one they share meets a
// choice or table, or one ends there on something that isn't a constant. It returns that level.
func readsFieldApart(x, u fieldRead) (level pickKey, apart bool) {
	for j := len(u.levels) - 1; j >= 0; j-- {
		if slices.Contains(x.levels, u.levels[j]) {
			both := x.levels[len(x.levels)-1] == u.levels[j] || j == len(u.levels)-1
			return u.levels[j], drawsAlong(u.in.startOf(u.a.head), u.a.tail[:j], both)
		}
	}
	return "", false
}

// spellingThrough is how to read level, a key in the pick of b, through b's name.
func spellingThrough(b *nameBinding, target nameTarget, level pickKey) string {
	segs, err := grammar.SplitPath(string(level))
	if err != nil || len(segs) < len(target.tail) {
		panic(invariant.Broken("level %q of name %q is not a path under its binding", level, b.name))
	}
	return grammar.JoinSegments(append([]string{b.name}, segs[len(target.tail):]...))
}

// readAddressing is the spelling of a read other than self that lands on or passes key, or lands on
// a level whose render reads leaf, the node at key; tail is how many segments of its path the name's
// binding holds.
func readAddressing(names resolvedNames, reads []*arm, self *arm, tail int, key pickKey, leaf node) (string, bool) {
	for _, r := range reads {
		if r != self && (slices.Contains(r.levels[tail:], key) || rendersField(names, r, leaf)) {
			return r.spelling, true
		}
	}
	return "", false
}

// rendersField reports whether rendering what read lands on, down every field it renders, reads
// leaf or a node under it through a field read, and no reference to such a node. A reference there is
// drawn once with the level, so a read of the level and one of the whole agree.
func rendersField(names resolvedNames, read *arm, leaf node) bool {
	reads := false
	for _, a := range armsRead(read.leaves) {
		if a.kind != freshRead {
			continue
		}
		for _, l := range a.leaves {
			if _, under := names.pathTo(leaf, l); under {
				if grammar.IsRef(a.head) {
					return false
				}
				reads = true
			}
		}
	}
	return reads
}

// armsRead is every arm that rendering nodes read, down every variant of a choice, a table's format
// and every field a read renders. It stays in their category: a reference's or a reference-bound
// name's arm is counted, but what it renders is not followed, since neither shares the pick.
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
		case *choice, *table, *tableRow:
			for _, e := range renderEdges(n) {
				read(e.to)
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
