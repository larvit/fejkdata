package fejkdata

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestWalkPathStopsAtAMissingSegment(t *testing.T) {
	n := compiled(t, `{"format":"{a}","a":{"format":"{b}","b":"leaf"}}`)
	var seen []string
	walk := pathWalk{
		mode:  walkEvery,
		pins:  &pinSet{},
		level: func(tm *template, rest []string) error { seen = append(seen, "level:"+rest[0]); return nil },
		leaf:  func(n node) error { seen = append(seen, "leaf"); return nil },
	}
	if _, err := walkPath(n, []string{"a", "b"}, walk); err != nil {
		t.Fatalf("walkPath(a.b) = %v", err)
	}
	if want := []string{"level:a", "level:b", "leaf"}; !slices.Equal(seen, want) {
		t.Errorf("walk visited %v, want %v", seen, want)
	}
	seen = nil
	_, err := walkPath(n, []string{"a", "nope", "deeper"}, walk)
	if err == nil || !strings.Contains(err.Error(), `no field "nope"`) {
		t.Errorf("walkPath(a.nope.deeper) = %v, want the missing segment named", err)
	}
	if slices.Contains(seen, "leaf") {
		t.Errorf("walk reached a leaf past a missing segment: %v", seen)
	}
}

func TestWalkPathChoiceConsumesNoSegment(t *testing.T) {
	n := compiled(t, `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2"}]`)
	var leaves []node
	_, err := walkPath(n, []string{"f"}, pathWalk{
		mode: walkEvery,
		pins: &pinSet{},
		leaf: func(n node) error { leaves = append(leaves, n); return nil },
	})
	if err != nil || len(leaves) != 2 {
		t.Fatalf("walkPath through a choice = %v, %d leaves, want both variants' f", err, len(leaves))
	}
}

func TestWalkCoverStopsAtAChoice(t *testing.T) {
	n := compiled(t, `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2"}]`)
	var leaves []node
	_, err := walkPath(n, []string{"f"}, pathWalk{
		mode: walkCover,
		leaf: func(n node) error { leaves = append(leaves, n); return nil },
	})
	if err != nil || len(leaves) != 1 || leaves[0] != n {
		t.Fatalf("cover through a choice = %v, leaves %v, want the choice alone", err, leaves)
	}
}

func TestDrawPathPanicsOnAnUnprovedPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), `plain.f: no field "f"`) {
			t.Errorf("drawPath(plain, f) recovered %v, want a panic naming the path and the missing field", r)
		}
	}()
	drawPath(compiled(t, `"plain"`), []string{"f"}, "plain", &pinSet{}, &pathDraws{s: engine(1).rand})
}

func TestDeepDottedPath(t *testing.T) {
	// A 5-segment path descends through alternating object/array nodes; choices
	// on the path are single-variant, so it resolves deterministically.
	f := engine(1)
	f.categories = map[string]node{
		"deep": compiled(t, `{"format":"{a}","a":{"format":"{b}","b":{"format":"{c}","c":{"format":"{d}","d":"leaf"}}}}`),
	}
	if got, err := f.Fake("deep.a.b.c.d"); err != nil || got != "leaf" {
		t.Fatalf("Fake(deep.a.b.c.d) = %q, %v, want leaf", got, err)
	}
	// Rendering the whole tree resolves the same chain.
	if got, err := f.Fake("deep"); err != nil || got != "leaf" {
		t.Fatalf("Fake(deep) = %q, %v, want leaf", got, err)
	}
}

func TestDescendIntoStringErrors(t *testing.T) {
	f := engine(1)
	f.categories = map[string]node{"greeting": compiled(t, `"hej"`)}
	if _, err := f.Fake("greeting.extra"); err == nil || !strings.Contains(err.Error(), `no field "extra"`) {
		t.Fatalf("Fake(greeting.extra) = %v, want a no-field error", err)
	}
}

// TestPathThroughChoice pins the rule that keeps a dotted path from rendering on
// one call and failing on the next: every variant must carry the rest of the path.
func TestPathThroughChoice(t *testing.T) {
	dir := writeData(t, map[string]string{
		"every":  `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2","g":"x"}]`,
		"notall": `[{"format":"{f}","f":"1"},"plain"]`,
		"some":   `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2","extra":"x"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 200; i++ {
		if got := fake(t, f, "every.f"); got != "1" && got != "2" {
			t.Fatalf("every.f = %q, want 1 or 2", got)
		}
	}
	// A path only some variants carry is reported against what all of them carry.
	if _, err := f.Fake("some.extra"); err == nil || !strings.Contains(err.Error(), "all carry [f]") {
		t.Errorf("Fake(some.extra) = %v, want it to name what every variant carries", err)
	}
	var first string
	for i := 0; i < 200; i++ {
		_, err := f.Fake("notall.f")
		if err == nil {
			t.Fatal("notall.f = nil error, want the same failure every call")
		}
		if i == 0 {
			first = err.Error()
		} else if err.Error() != first {
			t.Fatalf("notall.f error varies between calls:\n  %s\n  %s", first, err.Error())
		}
	}
}

// TestPathKeyIsUnambiguous pins that the one shape which could collide cannot be
// written: a field literally named "a.b" and a field "a" holding "b" would both
// spell "a.b", so a dotted field name is rejected at New and a dot means a path
// wherever it appears.
func TestPathKeyIsUnambiguous(t *testing.T) {
	dir := writeData(t, map[string]string{
		"cat": `[{"format":"{a.b}","a.b":"1"},{"format":"{a}","a":{"format":"{b}","b":"2"}}]`,
	})
	_, err := New(WithoutShippedData(), WithDataPath(dir))
	if err == nil || !strings.Contains(err.Error(), `field "a.b" contains "."`) {
		t.Fatalf("New = %v, want the dotted field name rejected", err)
	}
	// The same data without the dotted key is fine, and the path resolves.
	f := newGenerator(t, writeData(t, map[string]string{
		"cat": `{"format":"{a.b}","a":{"format":"{b}","b":"2"}}`,
	}), WithSeed(1))
	if !slices.Contains(f.List(), "cat.a.b") {
		t.Error("List() omits cat.a.b, which the data carries")
	}
	if got := fake(t, f, "cat"); got != "2" {
		t.Fatalf("cat = %q, want 2", got)
	}
}

// TestMissingFieldNamesItself keeps the precise diagnosis for the ordinary typo: a
// single-variant choice always picks the same item, so it needs no every-variant
// guard and the error can name the field that is missing.
func TestMissingFieldNamesItself(t *testing.T) {
	f := newGenerator(t, "data", WithSeed(1))
	_, err := f.Fake("sv_SE.person.typo")
	if err == nil || !strings.Contains(err.Error(), `no field "typo"`) {
		t.Errorf("Fake(person.typo) = %v, want it to name the missing field", err)
	}
}

func TestBindingKeyIsNotAPathSegment(t *testing.T) {
	dir := writeData(t, map[string]string{
		"color": `["red","blue"]`,
		"name":  `{"format":"{/color} {w}","w":"x"}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if slices.Contains(f.List(), "name./color") {
		t.Error("List() advertises a binding key")
	}
	if _, err := f.Fake("name./color"); err == nil || !strings.Contains(err.Error(), `no field "/color"`) {
		t.Errorf("Fake(name./color) = %v, want a no-field error", err)
	}
}
