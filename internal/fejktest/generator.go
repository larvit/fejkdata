// Package fejktest holds what the tests of fejkdata and of its shipped data share. It imports
// no fejkdata package, so the root's own tests may import it.
package fejktest

import (
	"testing"
	"time"
)

// TestDay is the date every test generator takes as today.
var TestDay = DayOf(2026, 1, 15)

// Faker is what a fejkdata Generator offers a test.
type Faker interface {
	Fake(path string) (string, error)
	FakeTemplate(src string) (string, error)
}

// Fake generates a value, failing the test on error.
func Fake(t *testing.T, f Faker, path string) string {
	t.Helper()
	s, err := f.Fake(path)
	if err != nil {
		t.Fatalf("Fake(%q): %v", path, err)
	}
	return s
}

// FakeTemplate renders an inline template against a loaded generator, so its
// references resolve.
func FakeTemplate(t *testing.T, f Faker, s string) string {
	t.Helper()
	got, err := f.FakeTemplate(s)
	if err != nil {
		t.Fatalf("FakeTemplate(%s) = %v", s, err)
	}
	return got
}

func DayOf(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
