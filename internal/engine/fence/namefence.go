package fence

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// checkNameReads refuses a read of a name that could disagree with another read of it, or that reads
// a name of a choice's item it never renders.
func checkNameReads(label string, t *tree.Template, names resolvedNames) error {
	for _, r := range load.NamedReads(t, tree.AfterFinish) {
		if err := checkNameRead(t, r.Read, names); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, r.Op.Token().Body, err)
		}
	}
	return nil
}

// nameRead is one read of a name, with what each refusal rule reads of it.
type nameRead struct {
	in     *tree.Template // the template holding read
	read   *tree.Read
	pick   pickReads     // every read of the name, read included
	routes []route       // every read that rendering read makes
	names  resolvedNames // every name of the load
}

// nameRules is the refusal rules of a read of a name, each in the file of its name. A name keeps one
// pick (goal 4.3), and a template that can't keep that fails to load (goal 4.6). The README lists
// what the rules refuse, under Names, "Refused at `New`".
//
// Four refuse a read that reaches one level by a route another read of the name reaches by a
// second route, where the two can draw apart:
//   - refuseTwiceDrawn: the pick renders a field twice
//   - refuseDrawnApart: a nested name has a pick of its own
//   - refuseFreshDraw: a reference draws afresh
//   - refuseKeptAndFreshField: a field read the pick doesn't keep draws afresh
//
// One refuses a read that has no pick to read:
//   - refuseSkippedItem: the read steps into a choice's item without opening the item's frame
//
// A sixth, RefuseReadsInsideFields, refuses a name read inside the field bound to it. The load
// pipeline calls it before the cycle check, so such a read never reaches these.
//
// Each rule decides alone, so the order changes no verdict. It picks the error of a read that
// breaks two rules, and errors are a contract by what they name: keep the order.
// docs/decisions.md#an-error-is-a-contract-by-what-it-names-not-its-bytes
var nameRules = []func(nameRead) error{
	refuseTwiceDrawn,
	refuseDrawnApart,
	refuseSkippedItem,
	refuseFreshDraw,
	refuseKeptAndFreshField,
}

// checkNameRead refuses read, a read in t of a name, by the first of nameRules it breaks.
func checkNameRead(t *tree.Template, read *tree.Read, names resolvedNames) error {
	pick := names.pick(read.Named())
	nr := nameRead{in: t, read: read, pick: pick, routes: routesUnder(read.Leaves(), read.Key(), pick.addressed), names: names}
	for _, refuse := range nameRules {
		if err := refuse(nr); err != nil {
			return err
		}
	}
	return nil
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
	a      *tree.Read
	in     *tree.Template
	op     int
	kind   routeKind
	at     tree.PickKey
	levels []tree.PickKey
	by     string
}

// routesUnder is every read that rendering nodes makes, where nodes sit at key start in a name's pick
// and addressed holds the keys the name's reads address.
func routesUnder(nodes []tree.Node, start tree.PickKey, addressed map[tree.PickKey]string) []route {
	return walkRoutes(nodes, start, addressed, true)
}

// routesApart is every read that rendering nodes makes outside any pick, each a routeApart.
func routesApart(nodes []tree.Node) []route {
	return walkRoutes(nodes, "", nil, false)
}

// walkRoutes lists reads in render order. It visits a node once per key, and steps into its children
// by tree.RenderEdges, the list a render steps by:
//   - a template's edges are the leaves of its ops' reads, so it walks the reads and lists each as a
//     route. A read that stays in the pick's category passes its place in the pick down.
//   - a choice's items take the choice's place in the pick, since the pick keeps the item drawn.
//   - a table's format is apart: the pick pins the row, so its columns read one row whichever read
//     renders them.
//   - a column's cells are not walked, so no rule sees a read in a cell (item 452).
func walkRoutes(nodes []tree.Node, start tree.PickKey, addressed map[tree.PickKey]string, inPick bool) []route {
	type site struct {
		n      tree.Node
		at     tree.PickKey
		inPick bool
	}
	var found []route
	seen := map[site]bool{}
	var walk func(n tree.Node, at tree.PickKey, inPick bool)
	walk = func(n tree.Node, at tree.PickKey, inPick bool) {
		if seen[site{n, at, inPick}] {
			return
		}
		seen[site{n, at, inPick}] = true
		switch n := n.(type) {
		case *tree.Template:
			for i, o := range n.Compiled().Ops() {
				for _, a := range o.AllReads() {
					r := route{a: a, in: n, op: i}
					if inPick {
						r = r.placedAt(at, addressed)
					}
					found = append(found, r)
					if !staysInCategory(a) {
						continue
					}
					below, kept := r.below()
					for _, leaf := range a.Leaves() {
						walk(leaf, below, kept)
					}
				}
			}
		case *tree.Choice:
			for _, e := range tree.RenderEdges(n, tree.AfterFinish) {
				walk(e.To, at, inPick)
			}
		case *tree.TableColumn:
		default:
			for _, e := range tree.RenderEdges(n, tree.AfterFinish) {
				walk(e.To, "", false)
			}
		}
	}
	for _, n := range nodes {
		walk(n, start, inPick)
	}
	return found
}

// placedAt is r, met in what the pick renders at key at, with its place in the pick.
func (r route) placedAt(at tree.PickKey, addressed map[tree.PickKey]string) route {
	if r.a.Kind() == tree.NamedRead || grammar.IsRef(r.a.Head()) {
		return r
	}
	r.kind, r.at = routeFresh, at
	if by, kept := tree.LandsInPick(addressed, at, *r.a); kept {
		r.kind, r.by = routeKept, by
	}
	r.levels = make([]tree.PickKey, len(r.a.Levels()))
	for i, l := range r.a.Levels() {
		r.levels[i] = at.Under(l)
	}
	return r
}

// below is the key in the pick of what r renders, and whether that is part of the pick.
func (r route) below() (tree.PickKey, bool) {
	if r.kind != routeKept {
		return "", false
	}
	return r.levels[len(r.levels)-1], true
}

// staysInCategory reports whether what a renders is in a's category. A reference's and a
// reference-bound name's are not, and neither shares the pick.
func staysInCategory(a *tree.Read) bool {
	return a.Kind() == tree.NamedRead && a.Named().BindsField() || a.Kind() == tree.FreshRead && !grammar.IsRef(a.Head())
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

// pickReads is the reads of one name: the reads through it, and the keys they address, each with the
// spelling of the first read reaching it. bound is how many segments of a read's path the binding holds.
type pickReads struct {
	reads     []*tree.Read
	addressed map[tree.PickKey]string
	bound     int
}

func (names resolvedNames) pick(b *tree.NameBinding) pickReads {
	return pickReads{reads: names.reads[b], addressed: b.Addressed(), bound: len(b.Target().Tail())}
}

// landingOn is the spelling of a read that lands on key.
func (p pickReads) landingOn(key tree.PickKey) (string, bool) {
	for _, r := range p.reads {
		if r.Key() == key {
			return r.Spelling(), true
		}
	}
	return "", false
}

// secondDraw is how to spell by, a read of b, so it draws afresh.
func secondDraw(b *tree.NameBinding, by string) string {
	return b.Ref() + strings.TrimPrefix(by, b.Name())
}

// keyOf is the key in a pick of the level the segments of parts, one after another, lead to.
func keyOf(parts ...[]string) tree.PickKey {
	return tree.PickKey(strings.Join(slices.Concat(parts...), "."))
}
