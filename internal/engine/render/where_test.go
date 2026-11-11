package render

import (
	"testing"
)

// Rounding in drawIndex may land on a drawn index or one past the span; undrawnNear then takes the
// first undrawn index after it, else the last one before it, inside the span.
func TestUndrawnNearStaysInsideTheSpan(t *testing.T) {
	for _, c := range []struct {
		name              string
		i, from, to, want int
		drawn             []int
	}{
		{name: "drawn last index steps back", i: 2, from: 0, to: 3, drawn: []int{2}, want: 1},
		{name: "drawn last index of a short span", i: 1, from: 0, to: 2, drawn: []int{1}, want: 0},
		{name: "one past the span", i: 3, from: 0, to: 3, want: 2},
		{name: "drawn index takes the next", i: 1, from: 0, to: 3, drawn: []int{1}, want: 2},
		{name: "one past the span beside a drawn last index", i: 3, from: 0, to: 3, drawn: []int{2}, want: 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := undrawnNear(c.i, c.from, c.to, c.drawn); got != c.want {
				t.Errorf("undrawnNear(%d, %d, %d, %v) = %d, want %d", c.i, c.from, c.to, c.drawn, got, c.want)
			}
		})
	}
}
