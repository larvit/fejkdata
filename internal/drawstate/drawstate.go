// Package drawstate holds a generator's seeded randomness, its {seq()} counters and the date it
// treats as today.
package drawstate

import (
	"math/rand/v2"
	"sort"
	"time"
)

// State is one generator's draw state, kept across its renders.
type State struct {
	rand     *rand.Rand
	counters map[string]uint64
	today    time.Time
}

// New builds the state a seed and a date reproduce. It keeps the calendar day the date shows in its
// own location, stored as midnight UTC.
func New(seed uint64, today time.Time) *State {
	year, month, day := today.Date()
	return &State{
		rand:     rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		counters: map[string]uint64{},
		today:    time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
}

func (s *State) IntN(n int) int { return s.rand.IntN(n) }

func (s *State) Float64() float64 { return s.rand.Float64() }

// Today is the date to treat as today, at midnight UTC.
func (s *State) Today() time.Time { return s.today }

// Weighted draws an index into cum, a running sum of weights of 0 or more with a positive total,
// with odds by weight, so an index of weight 0 is never drawn.
func (s *State) Weighted(cum []float64) int {
	total := cum[len(cum)-1]
	x := s.Float64() * total
	// x stays below total: a product of a float64 under 1 and a positive float64 rounds below it.
	return sort.Search(len(cum), func(i int) bool { return cum[i] > x })
}

// Seq advances the counter {seq(key)} reads, or {seq()} for an empty key, and returns it.
func (s *State) Seq(key string) uint64 {
	s.counters[key]++
	return s.counters[key]
}
