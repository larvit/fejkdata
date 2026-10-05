// Package drawstate holds a generator's seeded randomness and its {seq()} counters.
package drawstate

import (
	"math/rand/v2"
	"sort"
)

// State is one generator's draw state, kept across its renders.
type State struct {
	rand     *rand.Rand
	counters map[string]uint64
}

// New builds the state a seed reproduces.
func New(seed uint64) *State {
	return &State{rand: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)), counters: map[string]uint64{}}
}

func (s *State) IntN(n int) int { return s.rand.IntN(n) }

func (s *State) Float64() float64 { return s.rand.Float64() }

// Weighted draws an index into cum, a running sum of positive weights, with odds by weight.
func (s *State) Weighted(cum []float64) int {
	x := s.Float64() * cum[len(cum)-1]
	return min(sort.Search(len(cum), func(i int) bool { return cum[i] > x }), len(cum)-1) // x can round up to the total
}

// Seq advances the counter {seq(key)} reads, or {seq()} for an empty key, and returns it.
func (s *State) Seq(key string) uint64 {
	s.counters[key]++
	return s.counters[key]
}
