package fejkdata

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestPathCheckStopsAtAMissingSegment(t *testing.T) {
	n := compiled(t, `{"format":"{a}","a":{"format":"{b}","b":"leaf"}}`)
	if leaves := compilePath(n, []string{"a", "b"}).leaves; len(leaves) != 1 {
		t.Fatalf("compilePath(a.b).leaves = %v, want the one leaf", leaves)
	}
	w := &pathCheck{tail: []string{"a", "nope", "deeper"}}
	_, err := w.run(n)
	if err == nil || !strings.Contains(err.Error(), `no field "nope"`) {
		t.Errorf("walk(a.nope.deeper) = %v, want the missing segment named", err)
	}
	if len(w.leaves) > 0 {
		t.Errorf("walk reached a leaf past a missing segment: %v", w.leaves)
	}
}

func TestPathCheckChoiceConsumesNoSegment(t *testing.T) {
	n := compiled(t, `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2"}]`)
	if leaves := compilePath(n, []string{"f"}).leaves; len(leaves) != 2 {
		t.Fatalf("compilePath through a choice = %v, want both variants' f", leaves)
	}
}

func TestDrawStepsPanicsOnAStepItsNodeLacks(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), `no field "f"`) {
			t.Errorf("drawSteps(plain, f) recovered %v, want a panic naming the missing field", r)
		}
	}()
	drawSteps(engine(1).draws, compiled(t, `"plain"`), []pathStep{{kind: stepField, name: "f"}}, &pinSet{}, nil, nil)
}

func TestDeepDottedPath(t *testing.T) {
	// A 5-segment path descends through alternating object/array nodes; choices
	// on the path are single-variant, so it resolves deterministically.
	f := engine(1)
	f.root.children = map[string]node{
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
	f.root.children = map[string]node{"greeting": compiled(t, `"hej"`)}
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

func TestFakePathNavigation(t *testing.T) {
	f := engine(1)
	f.root.children = map[string]node{
		"addr": compiled(t, `{"format":"{street}","street":"Main"}`),
	}
	if got, err := f.Fake("addr"); err != nil || !strings.Contains(got, "Main") {
		t.Fatalf("Fake(addr) = %q, %v", got, err)
	}
	if got, err := f.Fake("addr.street"); err != nil || got != "Main" {
		t.Fatalf("Fake(addr.street) = %q, %v, want Main", got, err)
	}
	if _, err := f.Fake("addr.nope"); err == nil {
		t.Error("Fake(addr.nope) = nil error, want missing-field error")
	}
	if _, err := f.Fake("missing"); err == nil {
		t.Error("Fake(missing) = nil error, want unknown-category error")
	}
}
