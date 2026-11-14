package names

import "fmt"

// refuseTwiceDrawn refuses a read of a name where what it renders reads a field twice and
// another read of the name addresses a draw the two reads share, so one level has two routes. The
// pick keeps one draw of the field: where n binds a category whose format is {w}-{w}, a {n.w} beside
// {n} cannot say which draw it reads. It looks at every level the pick renders: beside {n.w.a}, it
// refuses {n} where n's format reads {w} and w's reads {a}-{a}, since "w" and "w.a" are addressed. One
// op draws once, whichever alternative it takes, so only reads in two ops can draw apart.
func refuseTwiceDrawn(nr nameRead) error {
	read := nr.read.Spelling()
	for _, kept := range keptBySite(nr.routes) {
		for i, x := range kept {
			for _, y := range kept[:i] {
				if x.op != y.op && x.a.Head() == y.a.Head() && readTwiceApart(nr.pick, x, y) {
					return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, x.a.Head(), x.by, read, x.by)
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
func readTwiceApart(pick pickReads, x, y route) bool {
	shared, share := x.deepestShared(y)
	if !share {
		return false
	}
	_, landed := pick.landingOn(shared)
	endsOnShared := x.landsOn(shared) && y.landsOn(shared) || landed
	return x.drawsDownTo(shared, endsOnShared)
}
