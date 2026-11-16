package data_test

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func assertShare(t *testing.T, what string, got, of int, want float64) {
	t.Helper()
	sd := math.Sqrt(want * (1 - want) / float64(of))
	if g := float64(got) / float64(of); math.Abs(g-want) > 5*sd {
		t.Errorf("%s: %d of %d (%.3f%%), want %.3f%% within %.3f%%", what, got, of, g*100, want*100, 5*sd*100)
	}
}

func sortedNamesDrawn(seen map[string]int) []string {
	got := make([]string, 0, len(seen))
	for name := range seen {
		got = append(got, name)
	}
	slices.Sort(got)
	return got
}

type houseNumberShares struct{ lettered, oneDigit, fourDigits int }

func houseNumberSharesOf(n int, draw func() string) houseNumberShares {
	var shares houseNumberShares
	fejktest.EachDraw(n, draw, func(n string) {
		digits := strings.TrimRight(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		if digits != n {
			shares.lettered++
		}
		switch len(digits) {
		case 1:
			shares.oneDigit++
		case 4:
			shares.fourDigits++
		}
	})
	return shares
}
