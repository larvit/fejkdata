package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// drawSet is one render's reference draws: the unnamed draw group's, and each named one's.
type drawSet struct {
	unnamed draws
	named   map[string]*draws
}

// drawScope is where a render reads its reference paths: a draw set, in the draw group of the
// template rendering.
type drawScope struct {
	set   *drawSet
	group string
}

// newDrawSet makes the unnamed group's maps where the set is declared, keeping them on that frame's
// stack for a render that reads through them; a zero drawSet makes them on its first read instead.
func newDrawSet() drawSet {
	return drawSet{unnamed: draws{variant: map[string]node{}, value: map[string]draw{}}}
}

// renderOnce renders n as one render, over draws of its own.
func renderOnce(s *session, n node) string {
	var set drawSet
	return render(s, n, drawScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc drawScope) in(t *template) drawScope {
	if t.drawGroupKey != "" {
		sc.group = t.drawGroupKey
	}
	return sc
}

// draws is the set's draws for sc's draw group.
func (sc drawScope) draws() *draws {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*draws{}
		}
		d = &draws{variant: map[string]node{}, value: map[string]draw{}}
		sc.set.named[strings.Clone(sc.group)] = d // a key from sc would leak sc, and with it every render's draw set, to the heap
	}
	return d
}

// drawGroupOf reads a template's "drawGroup" (default ""), which a repeat cannot carry: each
// iteration renders in no draw group.
func drawGroupOf(m map[string]any, repeat int) (string, error) {
	v, ok := m["drawGroup"]
	if !ok {
		return "", nil
	}
	name, ok := v.(string)
	switch {
	case !ok:
		return "", fmt.Errorf("drawGroup must be a string, got %T", v)
	case name == "":
		return "", fmt.Errorf(`drawGroup "" is the default, so it has no effect; drop it`)
	case repeat > 1:
		return "", fmt.Errorf("drawGroup %q on a repeat names nothing, since each iteration is a render of its own; drop it", name)
	}
	return name, nil
}

// keyDrawGroup keys t's draw group by the category t sits in, "" for an inline template, so a name
// is local to its category.
func (t *template) keyDrawGroup(category string) {
	if t.drawGroup != "" {
		t.drawGroupKey = category + "/" + t.drawGroup
	}
}

// checkNestedDrawGroup refuses a template beneath one drawing in group that names group again,
// short of a repeat or another draw group.
func checkNestedDrawGroup(fields map[string]node, group string) error {
	if group == "" {
		return nil
	}
	var walk func(path string, n node) error
	walk = func(path string, n node) error {
		t, isTemplate := n.(*template)
		switch {
		case isTemplate && t.drawGroup == group:
			return fmt.Errorf("%q names drawGroup %q, the draw group this template draws in already; drop it", path, group)
		case isTemplate && (t.drawGroup != "" || t.repeat > 1):
			return nil
		}
		for _, c := range contained(n) {
			if err := walk(join(path, c.name), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	for _, name := range sortedNames(fields) {
		if err := walk(name, fields[name]); err != nil {
			return err
		}
	}
	return nil
}

// drawCheck fences each template of a scope as a render of its own, remembering which nodes read a
// reference path.
type drawCheck struct {
	memo map[readsMemo]bool
}

// readsMemo is one answer reads has given: for a node, and for each place it stops.
type readsMemo struct {
	n           node
	stopAtGroup bool
}

// checkDrawGroup refuses a draw group that splits nothing: one whose render reads every reference
// path inside a repeat or a nested draw group, which draw apart from it whatever it names.
func (c *drawCheck) checkDrawGroup(path string, n node) error {
	if t, ok := n.(*template); ok && t.drawGroup != "" && !c.splitsDraws(t) {
		return fmt.Errorf("%s: drawGroup %q splits nothing, since nothing it renders reads a reference path outside a repeat or a nested drawGroup; drop it", path, t.drawGroup)
	}
	return nil
}

func (c *drawCheck) checkDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !c.readsPath(t) {
		return nil
	}
	w := newDrawWalk()
	for _, e := range renderEdges(t) {
		w.edge(t, e, drawAt{group: t.drawGroupKey, route: drawRoute{e.reached(), e.label}})
	}
	if err := w.check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// checkRecordDraws fences the columns of a record — a template compiled at the top without a
// repeat — so a load proves the record view of it as well as the string view.
func (c *drawCheck) checkRecordDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !t.record {
		return nil
	}
	columns := recordColumns(t)
	if len(columns) == 0 {
		return nil
	}
	if err := checkColumnDraws(t, columns); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// readsPath reports whether rendering n reads a reference path, short of a repeat, which renders
// over draws of its own.
func (c *drawCheck) readsPath(n node) bool { return c.reads(n, false) }

// splitsDraws reports whether rendering n reads a reference path that n's own draw group answers
// for: one outside a repeat and outside a nested draw group, which hold their own draws.
func (c *drawCheck) splitsDraws(n node) bool { return c.reads(n, true) }

// reads walks what rendering n renders for a reference path, stopping at a repeat — and at a nested
// draw group when stopAtGroup — since each holds draws of its own.
func (c *drawCheck) reads(n node, stopAtGroup bool) bool {
	k := readsMemo{n, stopAtGroup}
	if r, done := c.memo[k]; done {
		return r
	}
	r := false
	for _, e := range renderEdges(n) {
		if a, _, isRef := refRead(n, e.label); (isRef && len(a.tail) > 0) || (!repeats(e.to) && !(stopAtGroup && grouped(e.to)) && c.reads(e.to, stopAtGroup)) {
			r = true
			break
		}
	}
	if c.memo == nil {
		c.memo = map[readsMemo]bool{}
	}
	c.memo[k] = r
	return r
}

func repeats(n node) bool {
	t, isTemplate := n.(*template)
	return isTemplate && t.repeat > 1
}

func grouped(n node) bool {
	t, isTemplate := n.(*template)
	return isTemplate && t.drawGroup != ""
}

// refRead is the reference an edge of n reads, and the node it is bound to; false when the edge
// reads none.
func refRead(n node, label string) (arm, node, bool) {
	t, isTemplate := n.(*template)
	if !isTemplate {
		return arm{}, nil, false
	}
	a := splitArm(label, t.refs)
	return a, t.fields[a.key], isRef(a.key)
}

// checkColumnDraws fences a record's columns as one render.
func checkColumnDraws(t *template, columns []string) error {
	w := newDrawWalk()
	for _, name := range columns {
		w.walk(t.fields[name], drawAt{group: t.drawGroupKey, route: drawRoute{spelling: fmt.Sprintf("column %q", name)}})
	}
	return w.check()
}

// drawWalk gathers what one render reads by reference and what it draws afresh, each by draw group,
// for check to compare.
type drawWalk struct {
	reads []pathRead
	read  map[drawKey]bool
	fresh []freshDraw
	seen  map[drawVisit]bool
}

// drawAt is where a walk stands: the draw group it draws in, how the render's root reached it, the
// reference it last crossed, and whether it renders inside a reference path's draw.
type drawAt struct {
	group string
	route drawRoute
	via   string
	held  bool
}

// drawRoute is how a render reaches a draw: as its author spells it, and the root edge's label.
type drawRoute struct{ spelling, label string }

// pathRead is one reference a render reads: a path, or a bare reference with no tail.
type pathRead struct {
	at     drawAt
	a      arm
	target node
}

type freshDraw struct {
	at drawAt
	n  node
}

type drawVisit struct {
	n     node
	group string
	held  bool
}

type drawKey struct {
	group string
	key   any
}

func newDrawWalk() *drawWalk {
	return &drawWalk{read: map[drawKey]bool{}, seen: map[drawVisit]bool{}}
}

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there.
func (w *drawWalk) walk(n node, at drawAt) {
	v := drawVisit{n, at.group, at.held}
	if w.seen[v] {
		return
	}
	w.seen[v] = true
	if !at.held {
		w.fresh = append(w.fresh, freshDraw{at, n})
	}
	if repeats(n) {
		return
	}
	if t, isTemplate := n.(*template); isTemplate && t.drawGroupKey != "" {
		at.group = t.drawGroupKey
	}
	for _, e := range renderEdges(n) {
		w.edge(n, e, at)
	}
}

func (w *drawWalk) edge(from node, e renderEdge, at drawAt) {
	a, target, reads := refRead(from, e.label)
	if !reads {
		w.walk(e.to, at)
		return
	}
	if k := (drawKey{at.group, a.path}); !w.read[k] {
		w.read[k] = true
		w.reads = append(w.reads, pathRead{at, a, target})
	}
	at.via, at.held = a.name, len(a.tail) > 0
	w.walk(e.to, at)
}

// check refuses what one draw per reference path cannot answer for, reads compared in path order so
// which pair is reported does not vary.
func (w *drawWalk) check() error {
	sort.SliceStable(w.reads, func(i, j int) bool {
		if w.reads[i].at.group != w.reads[j].at.group {
			return w.reads[i].at.group < w.reads[j].at.group
		}
		return w.reads[i].a.path < w.reads[j].a.path
	})
	if err := w.checkOverlaps(); err != nil {
		return err
	}
	return w.checkFreshDraws()
}

// checkOverlaps refuses a read of a level beside a path another read takes into it.
func (w *drawWalk) checkOverlaps() error {
	for i, level := range w.reads {
		for _, into := range w.reads[i+1:] {
			if into.at.group == level.at.group && strings.HasPrefix(into.a.path, level.a.path+".") {
				return overlapError(level.at.route, level.a.name, into)
			}
		}
	}
	return nil
}

// checkFreshDraws refuses a node drawn afresh where a reference path's draw holds it.
func (w *drawWalk) checkFreshDraws() error {
	pins := map[drawKey]pathRead{}
	for _, r := range w.reads {
		if len(r.a.tail) == 0 {
			continue
		}
		held := map[node]bool{}
		coverPath(r.target, r.a.tail, held)
		for n := range held {
			if _, pinned := pins[drawKey{r.at.group, n}]; !pinned {
				pins[drawKey{r.at.group, n}] = r
			}
		}
	}
	for _, f := range w.fresh {
		if r, pinned := pins[drawKey{f.at.group, f.n}]; pinned {
			return overlapError(f.at.route, f.at.via, r)
		}
	}
	return nil
}

func overlapError(route drawRoute, ref string, into pathRead) error {
	return fmt.Errorf("%s renders a level that %s reads a path into; name the fields you want instead, or draw them apart with a drawGroup", route.spelled(ref), into.at.route.spelled(into.a.name))
}

// spelled names the route, and the reference it reaches a draw by where its root edge is not that
// reference.
func (r drawRoute) spelled(ref string) string {
	if ref == "" || ref == r.label {
		return r.spelling
	}
	return fmt.Sprintf("%s with {%s}", r.spelling, ref)
}
