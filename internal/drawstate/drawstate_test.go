package drawstate

import "testing"

func TestOneSeedDrawsOneSequence(t *testing.T) {
	a, b := New(42), New(42)
	for i := 0; i < 50; i++ {
		if x, y := a.Uint64(), b.Uint64(); x != y {
			t.Fatalf("seed 42 diverged at draw %d: %d != %d", i, x, y)
		}
	}
}

func TestTwoSeedsDrawTwoSequences(t *testing.T) {
	a, b := New(1), New(2)
	for i := 0; i < 50; i++ {
		if a.Uint64() != b.Uint64() {
			return
		}
	}
	t.Fatal("seeds 1 and 2 drew identical sequences")
}

func TestCountersCountFromOneEach(t *testing.T) {
	s := New(1)
	for want := uint64(1); want <= 3; want++ {
		if got := s.Next(""); got != want {
			t.Fatalf("Next(\"\") = %d, want %d", got, want)
		}
	}
	if got := s.Next("orders"); got != 1 {
		t.Fatalf("Next(orders) = %d, want its own count from 1", got)
	}
	if got := s.Next(""); got != 4 {
		t.Fatalf("Next(\"\") after Next(orders) = %d, want 4", got)
	}
	if got := New(1).Next(""); got != 1 {
		t.Fatalf("a new state's Next = %d, want 1", got)
	}
}
