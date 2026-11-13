package drawstate

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

func firstDrawThatDiffers(a, b *State, draws int) (at, x, y int) {
	for i := 0; i < draws; i++ {
		if x, y := a.IntN(1<<30), b.IntN(1<<30); x != y {
			return i, x, y
		}
	}
	return -1, 0, 0
}

func TestOneSeedDrawsOneSequence(t *testing.T) {
	a, b := New(42, time.Time{}), New(42, time.Time{})
	if i, x, y := firstDrawThatDiffers(a, b, 50); i >= 0 {
		t.Fatalf("seed 42 diverged at draw %d: %d != %d", i, x, y)
	}
}

func TestTwoSeedsDrawTwoSequences(t *testing.T) {
	a, b := New(1, time.Time{}), New(2, time.Time{})
	if i, _, _ := firstDrawThatDiffers(a, b, 50); i < 0 {
		t.Fatal("seeds 1 and 2 drew identical sequences")
	}
}

func nextSeqs(s *State, name string, n int) []uint64 {
	var seqs []uint64
	for range n {
		seqs = append(seqs, s.Seq(name))
	}
	return seqs
}

func TestCountersCountFromOneEach(t *testing.T) {
	s := New(1, time.Time{})
	if got := nextSeqs(s, "", 3); !slices.Equal(got, []uint64{1, 2, 3}) {
		t.Fatalf("Seq(\"\") = %d, want 1, 2 and 3", got)
	}
	if got := s.Seq("orders"); got != 1 {
		t.Fatalf("Seq(orders) = %d, want its own count from 1", got)
	}
	if got := s.Seq(""); got != 4 {
		t.Fatalf("Seq(\"\") after Seq(orders) = %d, want 4", got)
	}
	if got := New(1, time.Time{}).Seq(""); got != 1 {
		t.Fatalf("a new state's Seq = %d, want 1", got)
	}
}

func weightedDraws(s *State, weights []float64, draws int) []int {
	counts := make([]int, len(weights))
	for i := 0; i < draws; i++ {
		counts[s.Weighted(weights)]++
	}
	return counts
}

func TestWeightedDrawsByWeight(t *testing.T) {
	counts := weightedDraws(New(1, time.Time{}), []float64{1, 2, 6}, 30000)
	for i, want := range []float64{1.0 / 6, 1.0 / 6, 4.0 / 6} {
		t.Run(fmt.Sprintf("index %d", i), func(t *testing.T) {
			if share := float64(counts[i]) / 30000; share < want-0.02 || share > want+0.02 {
				t.Errorf("index %d drew a share of %.3f, want about %.3f", i, share, want)
			}
		})
	}
}
