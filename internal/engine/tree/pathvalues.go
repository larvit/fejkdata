package tree

import (
	"sort"
)

// LengthIndex holds values by length, each with its odds.
type LengthIndex struct {
	texts   []string  // shortest first
	lengths []int     // each text's length in code points
	cum     []float64 // running sum of the texts' odds
}

func NewLengthIndex(texts []string, lengths []int, cum []float64) LengthIndex {
	return LengthIndex{texts, lengths, cum}
}

// Span bounds the texts in x.texts[from:to], those whose length is lo to hi.
func (x *LengthIndex) Span(lo, hi int) (from, to int) { return SpanOf(x.lengths, lo, hi) }

func (x *LengthIndex) Cum() []float64 { return x.cum }

func (x *LengthIndex) Lengths() []int { return x.lengths }

func (x *LengthIndex) Texts() []string { return x.texts }

func SpanOf(lengths []int, lo, hi int) (from, to int) {
	from = sort.Search(len(lengths), func(i int) bool { return lengths[i] >= lo })
	to = sort.Search(len(lengths), func(i int) bool { return lengths[i] > hi })
	return from, to
}

// Odds is the odds of the text at index i.
func (x *LengthIndex) Odds(i int) float64 { return x.cum[i] - x.Before(i) }

// Before is the running sum of the odds of the texts before index i.
func (x *LengthIndex) Before(i int) float64 {
	if i == 0 {
		return 0
	}
	return x.cum[i-1]
}

// Nearest is the span of the texts whose length is n or, where none is, nearest n, the shorter on
// a tie.
func (x *LengthIndex) Nearest(n int) (from, to int) {
	from, to = x.Span(n, n)
	if from < to {
		return from, to
	}
	l := x.lengths[len(x.lengths)-1]
	switch {
	case from == len(x.lengths):
	case from == 0 || x.lengths[from]-n < n-x.lengths[from-1]:
		l = x.lengths[from]
	default:
		l = x.lengths[from-1]
	}
	return x.Span(l, l)
}
