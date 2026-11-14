package names

import "github.com/larvit/fejkdata/internal/engine/tree"

// place is where a node stands as a name's pick renders: in the pick, at the key the pick keeps it
// under, or apart from the pick, where no read is kept.
type place struct {
	at     tree.PickKey
	inPick bool
}

// apart is the place of every node the pick doesn't hold.
var apart = place{}

// routesUnder is every read that rendering nodes makes, where nodes sit at key start in a name's pick
// and addressed holds the keys the name's reads address.
func routesUnder(nodes []tree.Node, start tree.PickKey, addressed map[tree.PickKey]string) []route {
	return walkRoutes(nodes, place{at: start, inPick: true}, addressed)
}

// routesApart is every read that rendering nodes makes outside any pick, each a routeApart.
func routesApart(nodes []tree.Node) []route {
	return walkRoutes(nodes, apart, nil)
}

// walkRoutes lists the reads that rendering nodes makes, in render order, where nodes stand at
// start. It visits a node once per place.
func walkRoutes(nodes []tree.Node, start place, addressed map[tree.PickKey]string) []route {
	w := routeWalk{addressed: addressed, seen: map[routeSite]bool{}}
	for _, n := range nodes {
		w.walk(n, start)
	}
	return w.found
}

// routeWalk is one walk of walkRoutes: the routes found so far, and the sites visited.
type routeWalk struct {
	addressed map[tree.PickKey]string
	found     []route
	seen      map[routeSite]bool
}

// routeSite is a node met at a place.
type routeSite struct {
	n  tree.Node
	at place
}

// walk lists the reads that rendering n makes, where n stands at p. Each case says where the nodes
// n renders stand; those nodes are n's edges in tree.RenderEdges.
func (w *routeWalk) walk(n tree.Node, p place) {
	if w.seen[routeSite{n, p}] {
		return
	}
	w.seen[routeSite{n, p}] = true
	switch n := n.(type) {
	case *tree.Template:
		// A read's route says where the read's leaves stand.
		w.walkTemplate(n, p)
	case *tree.Choice:
		// An item stands where its choice stands, since the pick keeps the item drawn.
		w.walkEdges(n, p)
	case *tree.TableColumn:
		// A cell is not walked: no rule sees what a cell reads.
	default:
		// Anything else renders apart. A table's format does: the pick pins the row, so its
		// columns read one row whichever read renders them.
		w.walkEdges(n, apart)
	}
}

func (w *routeWalk) walkEdges(n tree.Node, p place) {
	for _, e := range tree.RenderEdges(n) {
		w.walk(e.To, p)
	}
}

// walkTemplate lists each read of t's ops as a route, where t stands at p, and walks the leaves of
// each read that stays in t's category. Those leaves stand below the route.
func (w *routeWalk) walkTemplate(t *tree.Template, p place) {
	for i, o := range t.Compiled().Ops() {
		for _, a := range o.AllReads() {
			r := route{a: a, in: t, op: i}
			if p.inPick {
				r = r.placedAt(p.at, w.addressed)
			}
			w.found = append(w.found, r)
			if !staysInCategory(a) {
				continue
			}
			for _, leaf := range a.Leaves() {
				w.walk(leaf, r.below())
			}
		}
	}
}
