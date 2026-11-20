package names

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// refuseFreshDraw refuses read, a read of a name, where what it renders holds a reference that draws
// afresh a level another read of the name reads from the pick. For example, {n} and {n.ab.name}
// could differ where n's format reads {/cat.ab.name}, and so could {n} and {n.ab} where ab's format
// reads {name}. It refuses read where a node a reference reaches (refsReached) meets all three of
// these, each checked where it is named:
//   - the node is a level of the pick, below the name's start (levelOf)
//   - two draws of that level can differ (canDiffer)
//   - another read of the name reads that level from the pick (readingFromPick)
func refuseFreshDraw(nr nameRead) error {
	read, b := nr.read, nr.read.Named()
	for _, reached := range refsReached(nr.routes) {
		level, inPick := nr.levelOf(reached.leaf)
		if !inPick || !level.canDiffer() {
			continue
		}
		if by, reads := nr.readingFromPick(level, reached); reads {
			return fmt.Errorf("{%s} reaches {%s}, which draws afresh a path {%s} reads from the pick, so {%s} and {%s} could differ; drop {%s}, or write {%s} for a second draw",
				read.Spelling(), reached.ref.Spelling(), by, read.Spelling(), by, by, secondDraw(b, by))
		}
	}
	return nil
}

// refReach is a node that a reference lands on.
type refReach struct {
	ref  *tree.Read
	leaf tree.Node
}

// refsReached is every node the references among routes land on, and every node those render in turn,
// across categories. Each node is listed once, with the first reference to reach it.
//
// pending holds the reads still to follow, routes first. Following a reference adds every read
// its leaves make, in whatever category they sit, so the loop ends when no reference reaches a
// node it has not seen.
func refsReached(routes []route) []refReach {
	var found []refReach
	seen := map[tree.Node]bool{}
	for pending := slices.Clone(routes); len(pending) > 0; pending = pending[1:] {
		ref := pending[0].a
		if !grammar.IsRef(ref.Head()) {
			continue
		}
		for _, leaf := range ref.Leaves() {
			if seen[leaf] {
				continue
			}
			seen[leaf] = true
			found = append(found, refReach{ref, leaf})
			pending = append(pending, routesApart([]tree.Node{leaf})...)
		}
	}
	return found
}

// pickLevel is a node below the start of a name's target, as the name's pick sees it: path leads
// down to it from start, and the pick keeps its draw under key.
type pickLevel struct {
	start tree.Node
	path  []string
	key   tree.PickKey
}

// levelOf is n as a level of the pick of nr's name, where n lies below the start of the name's target.
func (nr nameRead) levelOf(n tree.Node) (pickLevel, bool) {
	start := nr.read.Named().Target().Start()
	path, below := nr.names.pathTo(start, n)
	if !below || len(path) == 0 {
		return pickLevel{}, false
	}
	return pickLevel{start: start, path: path, key: nr.pick.keyAt(path)}, true
}

// canDiffer reports whether two draws of l can differ: the walk down to it meets a choice or a
// table, or it ends on a level that isn't a constant. A reference to a constant can't differ.
func (l pickLevel) canDiffer() bool { return drawsAlong(l.start, l.path, true) }

// readingFromPick is the spelling of the first read of the name, other than nr's own, that reads
// level from the pick while reached.ref draws it afresh.
func (nr nameRead) readingFromPick(level pickLevel, reached refReach) (string, bool) {
	for _, r := range nr.pick.reads {
		if r != nr.read && nr.readsFromPick(r, level, reached) {
			return r.Spelling(), true
		}
	}
	return "", false
}

// readsFromPick reports whether r, a read of the name, reads level from the pick. It does where it
// lands on or passes the level's key, or where it lands on a level whose render reads
// reached.leaf, the node at level, through a field read and doesn't render reached.ref.
func (nr nameRead) readsFromPick(r *tree.Read, level pickLevel, reached refReach) bool {
	return nr.pick.addresses(r, level.key) || rendersField(nr.names, r, reached)
}

// rendersField reports whether rendering what read lands on, down every field it renders, reads
// reached.leaf or a node under it through a field read, and doesn't render reached.ref. That reference
// is drawn once with a level that holds it, so a read of that level and a read of the whole agree.
func rendersField(names resolvedNames, read *tree.Read, reached refReach) bool {
	routes := routesApart(read.Leaves())
	if slices.ContainsFunc(routes, func(r route) bool { return r.a == reached.ref }) {
		return false
	}
	return slices.ContainsFunc(routes, func(r route) bool {
		if !readsField(r.a) {
			return false
		}
		return slices.ContainsFunc(r.a.Leaves(), func(l tree.Node) bool {
			_, under := names.pathTo(reached.leaf, l)
			return under
		})
	})
}
