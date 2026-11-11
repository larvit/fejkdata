package render

import (
	"slices"
	"sort"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// drawIn draws one of the texts in x.Texts()[from:to], by their odds.
func drawIn(x *tree.LengthIndex, s *drawstate.State, from, to int) string {
	return x.Texts()[drawIndex(x, s, from, to, nil)]
}

// drawIndex draws the index of one of the texts in x.Texts()[from:to] by their odds, leaving out
// the ascending indexes drawn. It costs drawn's length plus a binary search over the span: it draws
// among the odds left, and steps past each drawn text the draw lands at or beyond.
func drawIndex(x *tree.LengthIndex, s *drawstate.State, from, to int, drawn []int) int {
	cum := x.Cum()
	base, left := x.Before(from), cum[to-1]-x.Before(from)
	for _, d := range drawn {
		left -= x.Odds(d)
	}
	v := base + s.Float64()*left
	for _, d := range drawn {
		if x.Before(d) > v {
			break
		}
		v += x.Odds(d)
	}
	return undrawnNear(from+sort.Search(to-from, func(i int) bool { return cum[from+i] > v }), from, to, drawn)
}

// undrawnNear is i, or where rounding landed i on a drawn index or on to, the first undrawn index
// after it, else the last one before it, within from:to.
func undrawnNear(i, from, to int, drawn []int) int {
	for j := i; j < to; j++ {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	for j := i - 1; j >= from; j-- {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	panic(invariant.Broken("a draw ran out of the values the load proved enough"))
}
