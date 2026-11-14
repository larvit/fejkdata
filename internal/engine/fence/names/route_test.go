package names

import (
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
)

func TestRouteLandsOnItsLastLevel(t *testing.T) {
	r := route{levels: []tree.PickKey{"w", "w.a"}}
	if got := r.landing(); got != "w.a" {
		t.Errorf("landing() = %q, want %q", got, "w.a")
	}
	if !r.landsOn("w.a") {
		t.Error(`landsOn("w.a") = false, want true`)
	}
	if r.landsOn("w") {
		t.Error(`landsOn("w") = true, want false: the route passes "w"`)
	}
}

func TestRoutesPartBelowTheDeepestLevelTheyShare(t *testing.T) {
	x := route{levels: []tree.PickKey{"w", "w.a"}}
	u := route{levels: []tree.PickKey{"w", "w.a", "w.a.b"}}
	for _, pair := range [][2]route{{x, u}, {u, x}} {
		shared, share := pair[0].deepestShared(pair[1])
		if !share || shared != "w.a" {
			t.Errorf("deepestShared = %q, %v, want %q, true", shared, share, "w.a")
		}
	}
	other := route{levels: []tree.PickKey{"v", "v.a"}}
	if shared, share := x.deepestShared(other); share {
		t.Errorf("deepestShared = %q, true, want none: the routes share no level", shared)
	}
}
