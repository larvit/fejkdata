package names

import (
	"slices"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// routeKind says what a name's pick does with a read that rendering the name reaches.
type routeKind uint8

const (
	routeApart routeKind = iota // a reference, a read through a name, or any read below a read that isn't kept
	routeFresh                  // a field read in what the pick renders, not kept, so it draws afresh
	routeKept                   // the pick keeps its value, and what it renders is part of the pick
)

// route is one read met while a name's pick renders: the read a, in the template in, in the op at
// index op of that template.
//
// A kept or a fresh route has a place in the pick. Field at is the key of its template there, and
// levels holds one key for each node the read passes: its head, then the end of each longer prefix
// of its tail. So levels is one longer than the tail, and levels[i] is the key of the node where
// the walk of a.Tail()[:i] from the head ends. A level's key is its path from the start of the
// name's target (tree.PickKey), so two routes pass one level exactly where they hold one key.
// Only this file indexes levels or a tail: a rule asks a route by key.
//
// Field by is the spelling of the read of the name that addresses a kept route. A routeApart has
// neither at, levels nor by.
type route struct {
	a      *tree.Read
	in     *tree.Template
	op     int
	kind   routeKind
	at     tree.PickKey
	levels []tree.PickKey
	by     string
}

// placedAt is r, met in what the pick renders at key at, with its place in the pick. The pick
// keeps r where a read of the name addresses r's first level (tree.LandsInPick).
func (r route) placedAt(at tree.PickKey, addressed map[tree.PickKey]string) route {
	if !readsField(r.a) {
		return r
	}
	r.kind, r.at = routeFresh, at
	if by, kept := tree.LandsInPick(addressed, at, *r.a); kept {
		r.kind, r.by = routeKept, by
	}
	r.levels = tree.LevelsUnder(at, *r.a, tree.KeyUnder(at, *r.a))
	return r
}

// kept reports whether the pick keeps r's value.
func (r route) kept() bool { return r.kind == routeKept }

// drawsAfresh reports whether r is a field read the pick renders and doesn't keep.
func (r route) drawsAfresh() bool { return r.kind == routeFresh }

// landing is the key of the level r lands on, its last.
func (r route) landing() tree.PickKey { return r.levels[len(r.levels)-1] }

// landsOn reports whether level is the one r lands on, and not one it passes on the way.
func (r route) landsOn(level tree.PickKey) bool { return r.landing() == level }

// below is where what r renders stands: at r's landing where the pick keeps r, else apart.
func (r route) below() place {
	if !r.kept() {
		return apart
	}
	return place{at: r.landing(), inPick: true}
}

// deepestShared is the deepest level that r and other both pass or land on. Below it the two
// routes part.
func (r route) deepestShared(other route) (tree.PickKey, bool) {
	for i := len(r.levels) - 1; i >= 0; i-- {
		if slices.Contains(other.levels, r.levels[i]) {
			return r.levels[i], true
		}
	}
	return "", false
}

// drawsDownTo reports whether r's walk from its head down to level, one of r's levels, meets a
// choice or a table, or, with leaf, ends on a level that isn't a constant.
func (r route) drawsDownTo(level tree.PickKey, leaf bool) bool {
	steps := slices.Index(r.levels, level)
	return drawsAlong(r.in.StartOf(r.a.Head(), r.in.Refs()), r.a.Tail()[:steps], leaf)
}

// readsField reports whether a reads a field of the template it sits in: neither a reference nor
// a read through a name.
func readsField(a *tree.Read) bool {
	return a.Kind() == tree.FreshRead && !grammar.IsRef(a.Head())
}

// staysInCategory reports whether what a renders is in a's category. A reference's and a
// reference-bound name's are not, and neither shares the pick.
func staysInCategory(a *tree.Read) bool {
	return a.Kind() == tree.NamedRead && a.Named().BindsField() || readsField(a)
}

// keptBySite is the routes the pick keeps, one group per template and key, each group in the order
// the walk first met it.
func keptBySite(routes []route) [][]route {
	type site struct {
		in *tree.Template
		at tree.PickKey
	}
	var groups [][]route
	index := map[site]int{}
	for _, r := range routes {
		if !r.kept() {
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
