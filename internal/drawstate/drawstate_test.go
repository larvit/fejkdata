package drawstate

import "testing"

func TestOneSeedDrawsOneSequence(t *testing.T) {
	a, b := New(42), New(42)
	for i := 0; i < 50; i++ {
		if x, y := a.IntN(1<<30), b.IntN(1<<30); x != y {
			t.Fatalf("seed 42 diverged at draw %d: %d != %d", i, x, y)
		}
	}
}

func TestTwoSeedsDrawTwoSequences(t *testing.T) {
	a, b := New(1), New(2)
	for i := 0; i < 50; i++ {
		if a.IntN(1<<30) != b.IntN(1<<30) {
			return
		}
	}
	t.Fatal("seeds 1 and 2 drew identical sequences")
}

func TestCountersCountFromOneEach(t *testing.T) {
	s := New(1)
	for want := uint64(1); want <= 3; want++ {
		if got := s.Seq(""); got != want {
			t.Fatalf("Seq(\"\") = %d, want %d", got, want)
		}
	}
	if got := s.Seq("orders"); got != 1 {
		t.Fatalf("Seq(orders) = %d, want its own count from 1", got)
	}
	if got := s.Seq(""); got != 4 {
		t.Fatalf("Seq(\"\") after Seq(orders) = %d, want 4", got)
	}
	if got := New(1).Seq(""); got != 1 {
		t.Fatalf("a new state's Seq = %d, want 1", got)
	}
}
