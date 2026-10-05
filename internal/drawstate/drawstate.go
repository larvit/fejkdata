// Package drawstate holds what a generator draws through: the seeded randomness and the
// {seq()} counters.
package drawstate

import "math/rand/v2"

// State is one generator's draw state, kept across its renders.
type State struct {
	*rand.Rand
	counters map[string]uint64
}

// New builds the state a seed reproduces.
func New(seed uint64) *State {
	return &State{Rand: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)), counters: map[string]uint64{}}
}

// Next advances the counter key names, the one {seq()} or {seq(key)} reads, and returns it.
func (s *State) Next(key string) uint64 {
	s.counters[key]++
	return s.counters[key]
}
