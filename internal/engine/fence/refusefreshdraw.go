package fence

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// refuseFreshDraw refuses read, a read of a name, where what it renders holds a reference that draws
// afresh a level another read of the name reads from the pick. For example, {n} and {n.ab.name}
// could differ where n's format reads {/cat.ab.name}, and so could {n} and {n.ab} where ab's format
// reads {name}. It takes each node the references reach, through other references too
// (refsReached), and refuses read where a node meets all three of these, checked in this order:
//   - the node lies below the name's start, at a path from it (pathTo)
//   - the walk of that path meets a choice or a table, or ends on a level that isn't a constant,
//     since a reference to a constant can't differ (drawsAlong)
//   - another read of the name lands on or passes that level, or renders the node through a field
//     read without rendering the reference (readAddressing)
func refuseFreshDraw(nr nameRead) error {
	read, b := nr.read, nr.read.Named()
	target := b.Target()
	for _, reached := range refsReached(nr.routes) {
		path, below := nr.names.pathTo(target.Start(), reached.leaf)
		if !below || len(path) == 0 {
			continue
		}
		if !drawsAlong(target.Start(), path, true) {
			continue
		}
		if by, other := nr.pick.readAddressing(read, keyOf(underSelector(target, path)), reached, nr.names); other {
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
func refsReached(routes []route) []refReach {
	var found []refReach
	seen := map[tree.Node]bool{}
	for queue := slices.Clone(routes); len(queue) > 0; queue = queue[1:] {
		x := queue[0].a
		if !grammar.IsRef(x.Head()) {
			continue
		}
		for _, leaf := range x.Leaves() {
			if !seen[leaf] {
				seen[leaf] = true
				found = append(found, refReach{x, leaf})
				queue = append(queue, routesApart([]tree.Node{leaf})...)
			}
		}
	}
	return found
}

// underSelector is path, from a binding's start, as a key spells it. A key holds the selector of a
// binding such as {/region[01] as r}, and a walk of the tree passes no selector.
func underSelector(target tree.NameTarget, path []string) []string {
	if len(target.Tail()) > 0 && grammar.IsSelector(target.Tail()[0]) {
		return slices.Concat(target.Tail()[:1], path)
	}
	return path
}

// readAddressing is the spelling of a read other than self that lands on or passes key, or lands on
// a level whose render reads reached.leaf, the node at key, and doesn't render reached.ref.
func (p pickReads) readAddressing(self *tree.Read, key tree.PickKey, reached refReach, names resolvedNames) (string, bool) {
	for _, r := range p.reads {
		if r != self && (slices.Contains(r.Levels()[p.bound:], key) || rendersField(names, r, reached)) {
			return r.Spelling(), true
		}
	}
	return "", false
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
		if r.a.Kind() != tree.FreshRead || grammar.IsRef(r.a.Head()) {
			return false
		}
		return slices.ContainsFunc(r.a.Leaves(), func(l tree.Node) bool {
			_, under := names.pathTo(reached.leaf, l)
			return under
		})
	})
}
