package fence

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// fieldRead is a field read that draws afresh. Field of is the read of the name whose render makes it.
type fieldRead struct {
	route
	of *tree.Read
}

// freshFieldReads is, for one name, the reads of a field that rendering a read of it makes afresh,
// listed under each level they read. Reads of one key render the same, so it walks each key once.
func freshFieldReads(reads []*tree.Read, addressed map[tree.PickKey]string) map[tree.PickKey][]fieldRead {
	byLevel := map[tree.PickKey][]fieldRead{}
	walked := map[tree.PickKey]bool{}
	for _, read := range reads {
		if walked[read.Key()] {
			continue
		}
		walked[read.Key()] = true
		for _, r := range routesUnder(read.Leaves(), read.Key(), addressed) {
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

// refuseKeptAndFreshField refuses a read of a name where it reads a field from the pick that
// another read of the name draws afresh, and the two could differ: with {n} and {n.w}, where n's
// format is {w.a} and w's is {a}, {n} reads the pick's w.a and {n.w} draws w's a afresh.
func refuseKeptAndFreshField(nr nameRead) error {
	b := nr.read.Named()
	for _, x := range nr.routes {
		if x.kind != routeKept {
			continue
		}
		if u, shared, apart := drawnAfresh(x, nr.names.fresh[b]); apart {
			read, other := nr.read.Spelling(), u.of.Spelling()
			keep := spellingThrough(b, b.Target(), shared)
			return fmt.Errorf("{%s} reads {%s} from the pick, but {%s} renders {%s}, the same field, afresh, so {%s} and {%s} could differ; read {%s} in place of {%s} to keep one draw, or write {%s} for a second draw",
				read, x.a.Spelling(), other, u.a.Spelling(), read, other, keep, other, secondDraw(b, other))
		}
	}
	return nil
}

// drawnAfresh is the first of the name's fresh field reads, byLevel, that x, which the pick keeps,
// could draw apart from, with the level the two share.
func drawnAfresh(x route, byLevel map[tree.PickKey][]fieldRead) (u fieldRead, shared tree.PickKey, apart bool) {
	for _, level := range x.levels {
		for _, u := range byLevel[level] {
			if shared, apart := readsFieldApart(x, u); apart {
				return u, shared, true
			}
		}
	}
	return fieldRead{}, "", false
}

// readsFieldApart reports whether x, which the pick keeps, and u, which draws afresh, read a field
// they could draw apart: they share a level, and the walk to the deepest one they share meets a
// choice or table, or one ends there on something that isn't a constant. It returns that level.
// u.levels[0] is u's head, so u.levels[j] is where the walk of u's tail[:j] ends.
func readsFieldApart(x route, u fieldRead) (shared tree.PickKey, apart bool) {
	for j := len(u.levels) - 1; j >= 0; j-- {
		if slices.Contains(x.levels, u.levels[j]) {
			oneEndsHere := x.levels[len(x.levels)-1] == u.levels[j] || j == len(u.levels)-1
			return u.levels[j], drawsAlong(u.in.StartOf(u.a.Head(), u.in.Refs()), u.a.Tail()[:j], oneEndsHere)
		}
	}
	return "", false
}

// spellingThrough is how to read level, a key in the pick of b, through b's name.
func spellingThrough(b *tree.NameBinding, target tree.NameTarget, level tree.PickKey) string {
	segs, err := grammar.SplitPath(string(level))
	if err != nil || len(segs) < len(target.Tail()) {
		panic(invariant.Broken("level %q of name %q is not a path under its binding", level, b.Name()))
	}
	return grammar.JoinSegments(append([]string{b.Name()}, segs[len(target.Tail()):]...))
}
