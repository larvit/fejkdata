package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// drawSet is one render's reference draws: the unnamed group's, and each named group's.
type drawSet struct {
	unnamed draws
	named   map[string]*draws
}

// drawScope is where a render reads its reference paths: a draw set, in the group of the
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

// in is the scope t renders in: the group it names, else its caller's.
func (sc drawScope) in(t *template) drawScope {
	if t.drawGroup != "" {
		sc.group = t.drawGroup
	}
	return sc
}

// draws is the set's draws for sc's group.
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

// groupOf reads a template's "group" (default "").
func groupOf(m map[string]any) (string, error) {
	v, ok := m["group"]
	if !ok {
		return "", nil
	}
	name, ok := v.(string)
	switch {
	case !ok:
		return "", fmt.Errorf("group must be a string, got %T", v)
	case name == "":
		return "", fmt.Errorf(`group "" is the default, so it has no effect; drop it`)
	}
	return name, nil
}

// drawCheck fences each template of a scope as a render of its own, remembering which nodes
// read a reference path.
type drawCheck struct {
	reads map[node]bool
}

func (c *drawCheck) check(path string, n node) error {
	t, ok := n.(*template)
	switch {
	case !ok:
		return nil
	case !c.readsPath(t) && t.drawGroup != "":
		return fmt.Errorf("%s: group %q splits nothing, since nothing it renders reads a reference path; drop it", path, t.drawGroup)
	case !c.readsPath(t):
		return nil
	}
	w := newDrawWalk(nil)
	for _, e := range renderEdges(t) {
		w.edge(t, e, t.drawGroup, drawRoute{e.reached(), e.label}, "", false)
	}
	if err := w.check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// readsPath reports whether rendering n reads a reference path, however deep.
func (c *drawCheck) readsPath(n node) bool {
	if r, done := c.reads[n]; done {
		return r
	}
	r := false
	for _, e := range renderEdges(n) {
		if a, _, isRef := refRead(n, e.label); (isRef && len(a.tail) > 0) || c.readsPath(e.to) {
			r = true
			break
		}
	}
	if c.reads == nil {
		c.reads = map[node]bool{}
	}
	c.reads[n] = r
	return r
}

// refRead is the reference an edge of n reads, and the node it is bound to; false when the
// edge reads none.
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
	w := newDrawWalk(t)
	for _, name := range columns {
		w.walk(t.fields[name], t.drawGroup, drawRoute{spelling: fmt.Sprintf("column %q", name)}, "", false)
	}
	return w.check()
}

// drawWalk gathers what one render reads through its draws and what it draws afresh, each by
// group, for check to compare. record is set for a record's columns, which may not read the
// record back.
type drawWalk struct {
	record *template
	reads  []pathRead
	read   map[drawKey]bool
	fresh  []freshDraw
	seen   map[drawVisit]bool
	err    error
}

// drawRoute is how a render reaches a draw: as its author spells it, and the root edge's label.
type drawRoute struct{ spelling, label string }

type pathRead struct {
	group  string
	route  drawRoute
	a      arm
	target node
}

type freshDraw struct {
	group string
	route drawRoute
	via   string
	n     node
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

func newDrawWalk(record *template) *drawWalk {
	return &drawWalk{record: record, read: map[drawKey]bool{}, seen: map[drawVisit]bool{}}
}

// walk follows what rendering n renders. held says n renders inside a reference path's draw, so
// what it draws belongs to that draw; a repeat renders over draws of its own, so the walk stops there.
func (w *drawWalk) walk(n node, group string, route drawRoute, via string, held bool) {
	v := drawVisit{n, group, held}
	if w.seen[v] || w.err != nil {
		return
	}
	w.seen[v] = true
	if !held {
		w.fresh = append(w.fresh, freshDraw{group, route, via, n})
	}
	t, isTemplate := n.(*template)
	if isTemplate && t.repeat > 1 {
		return
	}
	if isTemplate && t.drawGroup != "" {
		group = t.drawGroup
	}
	for _, e := range renderEdges(n) {
		w.edge(n, e, group, route, via, held)
	}
}

func (w *drawWalk) edge(from node, e renderEdge, group string, route drawRoute, via string, held bool) {
	a, target, reads := refRead(from, e.label)
	switch {
	case !reads:
		w.walk(e.to, group, route, via, held)
	case w.record != nil && target == node(w.record):
		w.err = fmt.Errorf("%s reads {%s}, which points back at this record; a column cannot read another column — move the shared value into its own category and reference that", route.spelling, a.name)
	case len(a.tail) > 0:
		if k := (drawKey{group, a.path}); !w.read[k] {
			w.read[k] = true
			w.reads = append(w.reads, pathRead{group, route, a, target})
		}
		w.walk(e.to, group, route, a.name, true)
	default:
		w.walk(e.to, group, route, a.name, false)
	}
}

// check refuses what one draw per reference path cannot answer for: a path read into a level
// another read renders, and a node drawn afresh beside a path whose draw holds it.
func (w *drawWalk) check() error {
	if w.err != nil {
		return w.err
	}
	sort.SliceStable(w.reads, func(i, j int) bool {
		if w.reads[i].group != w.reads[j].group {
			return w.reads[i].group < w.reads[j].group
		}
		return w.reads[i].a.path < w.reads[j].a.path
	})
	pins := map[drawKey]pathRead{}
	for i, level := range w.reads {
		for _, into := range w.reads[i+1:] {
			if into.group == level.group && strings.HasPrefix(into.a.path, level.a.path+".") {
				return overlap(level.route, level.a.name, into)
			}
		}
		held := map[node]bool{}
		coverPath(level.target, level.a.tail, held)
		for n := range held {
			if _, pinned := pins[drawKey{level.group, n}]; !pinned {
				pins[drawKey{level.group, n}] = level
			}
		}
	}
	for _, f := range w.fresh {
		if r, pinned := pins[drawKey{f.group, f.n}]; pinned {
			return overlap(f.route, f.via, r)
		}
	}
	return nil
}

func overlap(route drawRoute, ref string, into pathRead) error {
	return fmt.Errorf("%s renders a level that %s reads a path into; name the fields you want instead, or draw them apart with a group", route.spelled(ref), into.route.spelled(into.a.name))
}

// spelled names the route, and the reference it reaches a draw by where its root edge is not
// that reference.
func (r drawRoute) spelled(ref string) string {
	if ref == "" || ref == r.label {
		return r.spelling
	}
	return fmt.Sprintf("%s with {%s}", r.spelling, ref)
}
