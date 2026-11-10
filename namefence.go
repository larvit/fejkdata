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
	for _, r := range namedReads(t, names.links) {
		if err := checkNameRead(t, r.a, names); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, r.o.Body, err)
		}
	}
	return nil
}

// checkNameRead refuses read, a read in t of a name, where the name can't keep its promise. A name
// keeps one pick (goal 4.3), and a template that can't keep that fails to load (goal 4.6). There are
// two ways to break it.
//
// The read has no pick to read:
//   - the name is read inside the field bound to it: refuseReadInsideField
//   - the read steps into a choice's item without opening the item's frame: refuseSkippedItem
//
// The read and another read of the name reach one level by routes that can draw apart:
//   - the pick renders a field twice: refuseTwiceDrawn
//   - a nested name has a pick of its own: refuseDrawnApart
//   - a reference draws afresh: refuseFreshDraw
//   - a field read the pick doesn't keep draws afresh: refuseKeptAndFreshField
func checkNameRead(t *template, read *arm, names resolvedNames) error {
	if err := refuseReadInsideField(t, read.named, names.links); err != nil {
		return err
	}
	pick := names.pick(read.named)
	routes := routesUnder(read.leaves, read.key(), pick.addressed, names.links)
	if err := refuseTwiceDrawn(pick, read.spelling, routes, names.links); err != nil {
		return err
	}
	if err := refuseDrawnApart(pick, read, routes, names); err != nil {
		return err
	}
	if err := refuseSkippedItem(t, read, routes, names); err != nil {
		return err
	}
	if err := refuseFreshDraw(pick, read, routes, names); err != nil {
		return err
	}
	return refuseKeptAndFreshField(read, routes, names)
}

func refuseReadInsideField(t *template, b *nameBinding, l *links) error {
	if !b.bindsField() {
		return nil
	}
	target := l.binding(b).target
	if rendersInside(compilePath(target.start, target.tail).leaves, t) {
		return fmt.Errorf("name %q is read inside %q, the field bound to it; read the name outside that field", b.name, b.ref)
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

// routeKind says what a name's pick does with a read that rendering the name reaches.
type routeKind uint8

const (
	routeApart routeKind = iota // a reference, a read through a name, or any read below a read that isn't kept
	routeFresh                  // a field read in what the pick renders, not kept, so it draws afresh
	routeKept                   // the pick keeps its value, and what it renders is part of the pick
)

// route is one read met while nodes render. Field a is the read, in the template holding it and op the
// index of the op holding it. Unless kind is routeApart, at is the template's key in the pick, and
// levels is a's levels moved under that key. by is the read of the name that addresses a kept route.
type route struct {
	a      *arm
	in     *template
	op     int
	kind   routeKind
	at     pickKey
	levels []pickKey
	by     string
}

// routesUnder is every read that rendering nodes makes, where nodes sit at key start in a name's pick
// and addressed holds the keys the name's reads address.
func routesUnder(nodes []node, start pickKey, addressed map[pickKey]string, l *links) []route {
	return walkRoutes(nodes, start, addressed, true, l)
}

// routesApart is every read that rendering nodes makes outside any pick, each a routeApart.
func routesApart(nodes []node, l *links) []route {
	return walkRoutes(nodes, "", nil, false, l)
}

// walkRoutes lists reads in render order. It goes down every variant of a choice, a table's format and
// every field a read renders, and visits a template once per key. A table's format is apart: the pick
// pins the row, so its columns read one row whichever read renders them.
func walkRoutes(nodes []node, start pickKey, addressed map[pickKey]string, inPick bool, l *links) []route {
	type site struct {
		n      node
		at     pickKey
		inPick bool
	}
	var found []route
	seen := map[site]bool{}
	var walk func(n node, at pickKey, inPick bool)
	walk = func(n node, at pickKey, inPick bool) {
		if seen[site{n, at, inPick}] {
			return
		}
		seen[site{n, at, inPick}] = true
		switch n := n.(type) {
		case *choice:
			for _, item := range n.items {
				walk(item, at, inPick)
			}
		case *table, *tableRow:
			for _, e := range renderEdges(n, l) {
				walk(e.to, "", false)
			}
		case *template:
			for i, o := range l.of(n).compiled.ops {
				for _, a := range o.reads() {
					r := route{a: a, in: n, op: i}
					if inPick {
						r = r.placedAt(at, addressed)
					}
					found = append(found, r)
					if !staysInCategory(a) {
						continue
					}
					below, kept := r.below()
					for _, leaf := range a.leaves {
						walk(leaf, below, kept)
					}
				}
			}
		}
	}
	for _, n := range nodes {
		walk(n, start, inPick)
	}
	return found
}

// placedAt is r, met in what the pick renders at key at, with its place in the pick.
func (r route) placedAt(at pickKey, addressed map[pickKey]string) route {
	if r.a.kind == namedRead || grammar.IsRef(r.a.head) {
		return r
	}
	r.kind, r.at = routeFresh, at
	if by, kept := landsInPick(addressed, at, *r.a); kept {
		r.kind, r.by = routeKept, by
	}
	r.levels = make([]pickKey, len(r.a.levels))
	for i, l := range r.a.levels {
		r.levels[i] = at.under(l)
	}
	return r
}

// below is the key in the pick of what r renders, and whether that is part of the pick.
func (r route) below() (pickKey, bool) {
	if r.kind != routeKept {
		return "", false
	}
	return r.levels[len(r.levels)-1], true
}

// staysInCategory reports whether what a renders is in a's category. A reference's and a
// reference-bound name's are not, and neither shares the pick.
func staysInCategory(a *arm) bool {
	return a.kind == namedRead && a.named.bindsField() || a.kind == freshRead && !grammar.IsRef(a.head)
}

// keptBySite is the routes the pick keeps, one group per template and key, each group in the order
// the walk first met it.
func keptBySite(routes []route) [][]route {
	type site struct {
		in *template
		at pickKey
	}
	var groups [][]route
	index := map[site]int{}
	for _, r := range routes {
		if r.kind != routeKept {
			continue
		}
		i, met := index[site{r.in, r.at}]
		if !met {
			i = len(groups)
			index[site{r.in, r.at}] = i
			groups = append(groups, nil)
		}
		groups[i] = append(groups[i], r)
	}
	return groups
}

// pickReads is the reads of one name: the arms that read it, and the keys they address, each with the
// spelling of the first read reaching it. bound is how many segments of a read's path the binding holds.
type pickReads struct {
	reads     []*arm
	addressed map[pickKey]string
	bound     int
}

func (names resolvedNames) pick(b *nameBinding) pickReads {
	bl := names.links.binding(b)
	return pickReads{reads: names.reads[b], addressed: bl.addressed, bound: len(bl.target.tail)}
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

// refuseTwiceDrawn refuses read, a read of a name, where what it renders reads a field twice and
// another read of the name addresses a draw the two reads share, so one level has two routes. The
// pick keeps one draw of the field: where n binds a category whose format is {w}-{w}, a {n.w} beside
// {n} cannot say which draw it reads. It looks at every level the pick renders: beside {n.w.a}, it
// refuses {n} where n's format reads {w} and w's reads {a}-{a}, since "w" and "w.a" are addressed. One
// op draws once, whichever arm it takes, so only reads in two ops can draw apart.
func refuseTwiceDrawn(pick pickReads, read string, routes []route, l *links) error {
	for _, kept := range keptBySite(routes) {
		for i, x := range kept {
			for _, y := range kept[:i] {
				if x.op != y.op && x.a.head == y.a.head && readTwiceApart(pick, x, y, l) {
					return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, x.a.head, x.by, read, x.by)
				}
			}
		}
	}
	return nil
}

// readTwiceApart reports whether x and y, two reads of one template that the pick keeps, could differ
// if drawn apart. They could where they:
//   - pass one choice or table
//   - read one leaf that isn't a constant
//   - share a level a read of the name lands on, since reading that whole level draws what the two keep
//
// Two fields of a template that isn't a choice do none of these.
func readTwiceApart(pick pickReads, x, y route, l *links) bool {
	common := 0
	for common < min(len(x.a.tail), len(y.a.tail)) && x.a.tail[common] == y.a.tail[common] {
		common++
	}
	_, landed := pick.landingOn(x.levels[common])
	endsOnShared := x.a.key() == y.a.key() || landed
	return drawsAlong(x.in.startOf(x.a.head, l.of(x.in).refs), x.a.tail[:common], endsOnShared)
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
func refuseDrawnApart(pick pickReads, read *arm, routes []route, names resolvedNames) error {
	outer := read.named
	for _, r := range routes {
		inner := r.a.named
		if r.a.kind != namedRead || !inner.bindsField() || inner == outer { // a name's own read inside its field is refused where it sits
			continue
		}
		target := names.links.binding(inner).target
		above, inside := names.pathTo(names.links.binding(outer).target.start, target.start)
		if !inside {
			continue
		}
		if by, shares := sharedDraw(pick, target, above); shares {
			return fmt.Errorf("{%s} reads field %q through name %q, a pick of its own, so {%s} would read another draw; drop {%s} to keep one pick, or write {%s} for a second draw",
				read.spelling, inner.ref, inner.name, by, by, secondDraw(outer, by))
		}
	}
	return nil
}

// sharedDraw reports whether a read of one name shares, with a second name bound to target, a draw
// that could differ, and returns that read. Parameter above is target's path under the first name's
// pick. It walks target's path level by level, and at each level, in this order:
//   - no read of the first name addresses the level: they share none
//   - the level is a choice or a table: they share it
//   - the level is target's leaf: they share it unless it is a constant
//   - a read lands on the level, and a draw lies below it on target's path: they share that draw
//   - target's path goes on by a step no field spells, such as "..": they count as sharing one
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
func refuseSkippedItem(t *template, read *arm, routes []route, names resolvedNames) error {
	if !read.named.bindsField() {
		return nil
	}
	for _, r := range routes {
		x := r.a
		if x.kind != namedRead {
			continue
		}
		scope := x.named.scope
		if scope.kind() != scopeChoiceItem || names.links.of(t).nameScope.within(scope) {
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
func refuseFreshDraw(pick pickReads, read *arm, routes []route, names resolvedNames) error {
	b := read.named
	target := names.links.binding(b).target
	for _, reached := range refsReached(routes, names.links) {
		path, within := names.pathTo(target.start, reached.leaf)
		if !within || len(path) == 0 || !drawsAlong(target.start, path, true) {
			continue
		}
		key := pickKey(strings.Join(underSelector(target, path), "."))
		if by, other := pick.readAddressing(read, key, reached, names); other {
			return fmt.Errorf("{%s} reaches {%s}, which draws afresh a path {%s} reads from the pick, so {%s} and {%s} could differ; drop {%s}, or write {%s} for a second draw",
				read.spelling, reached.ref.spelling, by, read.spelling, by, by, secondDraw(b, by))
		}
	}
	return nil
}

// refReach is a node that a reference lands on.
type refReach struct {
	ref  *arm
	leaf node
}

// refsReached is every node the references among routes land on, and every node those render in turn,
// across categories. Each node is listed once, with the first reference to reach it.
func refsReached(routes []route, l *links) []refReach {
	var found []refReach
	seen := map[node]bool{}
	for queue := slices.Clone(routes); len(queue) > 0; queue = queue[1:] {
		x := queue[0].a
		if !grammar.IsRef(x.head) {
			continue
		}
		for _, leaf := range x.leaves {
			if !seen[leaf] {
				seen[leaf] = true
				found = append(found, refReach{x, leaf})
				queue = append(queue, routesApart([]node{leaf}, l)...)
			}
		}
	}
	return found
}

// underSelector is path, from a binding's start, as a key spells it. A key holds the selector of a
// binding such as {/region[01] as r}, and a walk of the tree passes no selector.
func underSelector(target nameTarget, path []string) []string {
	if len(target.tail) > 0 && grammar.IsSelector(target.tail[0]) {
		return slices.Concat(target.tail[:1], path)
	}
	return path
}

// fieldRead is a field read that draws afresh. Field of is the read of the name whose render makes it.
type fieldRead struct {
	route
	of *arm
}

// freshFieldReads is, for one name, the reads of a field that rendering a read of it makes afresh,
// listed under each level they read. Reads of one key render the same, so it walks each key once.
func freshFieldReads(reads []*arm, addressed map[pickKey]string, l *links) map[pickKey][]fieldRead {
	byLevel := map[pickKey][]fieldRead{}
	walked := map[pickKey]bool{}
	for _, read := range reads {
		if walked[read.key()] {
			continue
		}
		walked[read.key()] = true
		for _, r := range routesUnder(read.leaves, read.key(), addressed, l) {
			if r.kind != routeFresh {
				continue
			}
			for _, level := range r.levels {
				byLevel[level] = append(byLevel[level], fieldRead{r, read})
			}
		}
	}
	return byLevel
}

// refuseKeptAndFreshField refuses read, a read of a name, where it reads a field from the pick that
// another read of the name draws afresh, and the two could differ: with {n} and {n.w}, where n's
// format is {w.a} and w's is {a}, {n} reads the pick's w.a and {n.w} draws w's a afresh.
func refuseKeptAndFreshField(read *arm, routes []route, names resolvedNames) error {
	b := read.named
	byLevel := names.fresh[b]
	for _, x := range routes {
		if x.kind != routeKept {
			continue
		}
		for _, level := range x.levels {
			for _, u := range byLevel[level] {
				if shared, apart := readsFieldApart(x, u, names.links); apart {
					keep := spellingThrough(b, names.links.binding(b).target, shared)
					second := secondDraw(b, u.of.spelling)
					return fmt.Errorf("{%s} reads {%s} from the pick, but {%s} renders {%s}, the same field, afresh, so {%s} and {%s} could differ; read {%s} in place of {%s} to keep one draw, or write {%s} for a second draw",
						read.spelling, x.a.spelling, u.of.spelling, u.a.spelling, read.spelling, u.of.spelling, keep, u.of.spelling, second)
				}
			}
		}
	}
	return nil
}

// readsFieldApart reports whether x, which the pick keeps, and u, which draws afresh, read a field
// they could draw apart: they share a level, and the walk to the deepest one they share meets a
// choice or table, or one ends there on something that isn't a constant. It returns that level.
// u.levels[0] is u's head, so u.levels[j] is where the walk of u's tail[:j] ends.
func readsFieldApart(x route, u fieldRead, l *links) (shared pickKey, apart bool) {
	for j := len(u.levels) - 1; j >= 0; j-- {
		if slices.Contains(x.levels, u.levels[j]) {
			oneEndsHere := x.levels[len(x.levels)-1] == u.levels[j] || j == len(u.levels)-1
			return u.levels[j], drawsAlong(u.in.startOf(u.a.head, l.of(u.in).refs), u.a.tail[:j], oneEndsHere)
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
// a level whose render reads reached.leaf, the node at key, and doesn't render reached.ref.
func (p pickReads) readAddressing(self *arm, key pickKey, reached refReach, names resolvedNames) (string, bool) {
	for _, r := range p.reads {
		if r != self && (slices.Contains(r.levels[p.bound:], key) || rendersField(names, r, reached)) {
			return r.spelling, true
		}
	}
	return "", false
}

// rendersField reports whether rendering what read lands on, down every field it renders, reads
// reached.leaf or a node under it through a field read, and doesn't render reached.ref. That reference
// is drawn once with a level that holds it, so a read of that level and a read of the whole agree.
func rendersField(names resolvedNames, read *arm, reached refReach) bool {
	routes := routesApart(read.leaves, names.links)
	if slices.ContainsFunc(routes, func(r route) bool { return r.a == reached.ref }) {
		return false
	}
	return slices.ContainsFunc(routes, func(r route) bool {
		if r.a.kind != freshRead || grammar.IsRef(r.a.head) {
			return false
		}
		return slices.ContainsFunc(r.a.leaves, func(l node) bool {
			_, under := names.pathTo(reached.leaf, l)
			return under
		})
	})
}
