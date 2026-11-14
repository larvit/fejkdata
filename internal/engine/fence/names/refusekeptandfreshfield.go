package names

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/tree"
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
			if !r.drawsAfresh() {
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
		if !x.kept() {
			continue
		}
		if u, shared, apart := drawnAfresh(x, nr.names.fresh[b]); apart {
			read, other := nr.read.Spelling(), u.of.Spelling()
			keep := nr.pick.spellingOf(shared)
			return fmt.Errorf("{%s} reads {%s} from the pick, but {%s} renders {%s}, the same field, afresh, so {%s} and {%s} could differ; read {%s} in place of {%s} to keep one draw, or write {%s} for a second draw",
				read, x.a.Spelling(), other, u.a.Spelling(), read, other, keep, other, secondDraw(b, other))
		}
	}
	return nil
}

// drawnAfresh finds a read in byLevel, the name's fresh field reads, that could draw apart from x.
// The pick keeps x. It returns the first such read and the level the two share.
func drawnAfresh(x route, byLevel map[tree.PickKey][]fieldRead) (fieldRead, tree.PickKey, bool) {
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
func readsFieldApart(x route, u fieldRead) (shared tree.PickKey, apart bool) {
	shared, share := u.deepestShared(x)
	if !share {
		return "", false
	}
	oneEndsHere := x.landsOn(shared) || u.landsOn(shared)
	return shared, u.drawsDownTo(shared, oneEndsHere)
}
