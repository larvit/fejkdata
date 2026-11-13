package fejkdata

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
)

func TestDeepDottedPath(t *testing.T) {
	// A 5-segment path descends through alternating object/array nodes; choices
	// on the path are single-variant, so it resolves deterministically.
	f := emptyGenerator(1)
	f.root = *tree.NewFolderOf(map[string]tree.Node{
		"deep": compiled(t, `{"format":"{a}","a":{"format":"{b}","b":{"format":"{c}","c":{"format":"{d}","d":"leaf"}}}}`),
	})
	if got, err := f.Fake("deep.a.b.c.d"); err != nil || got != "leaf" {
		t.Fatalf("Fake(deep.a.b.c.d) = %q, %v, want leaf", got, err)
	}
	// Rendering the whole tree resolves the same chain.
	if got, err := f.Fake("deep"); err != nil || got != "leaf" {
		t.Fatalf("Fake(deep) = %q, %v, want leaf", got, err)
	}
}

func TestDescendIntoStringErrors(t *testing.T) {
	f := emptyGenerator(1)
	f.root = *tree.NewFolderOf(map[string]tree.Node{"greeting": compiled(t, `"hej"`)})
	if _, err := f.Fake("greeting.extra"); err == nil || !strings.Contains(err.Error(), `no field "extra"`) {
		t.Fatalf("Fake(greeting.extra) = %v, want a no-field error", err)
	}
}

func generatorOfLevelsCarryingARepeat(t *testing.T) *Generator {
	t.Helper()
	dir := writeData(t, map[string]string{
		"outer": `{"format":"{inner}","inner":{"format":"{a}","a":"v","repeat":3,"separator":","}}`,
		"x":     `{"format":"{a}","a":"v","repeat":3,"separator":","}`,
		"y":     `{"format":"{p}","p":[{"format":"{a}","a":"v","repeat":3,"separator":","},{"format":"{a}","a":"w"}]}`,
	})
	return newGenerator(t, dir, WithSeed(1))
}

func TestPathRefusesALevelCarryingARepeat(t *testing.T) {
	f := generatorOfLevelsCarryingARepeat(t)
	for path, level := range map[string]string{"x.a": "x", "outer.inner.a": "outer.inner"} {
		t.Run(path, func(t *testing.T) {
			want := fmt.Sprintf("the level %q carries a repeat", level)
			if got, err := f.Fake(path); err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("Fake(%s) = %q, %v, want %s", path, got, err, want)
			}
			if _, err := f.NewTemplate("{/" + path + "}"); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("the level %q carries a repeat", "/"+level)) {
				t.Errorf("NewTemplate({/%s}) = %v, want the level /%s named", path, err, level)
			}
		})
	}
}

func TestListAdvertisesNoPathThroughALevelCarryingARepeat(t *testing.T) {
	f := generatorOfLevelsCarryingARepeat(t)
	for _, p := range f.List() {
		t.Run(p, func(t *testing.T) {
			if _, err := f.Fake(p); err != nil {
				t.Errorf("Fake(%q) = %v, but List() advertises it", p, err)
			}
		})
	}
}

func TestPathThroughANameRefusesALevelCarryingARepeat(t *testing.T) {
	f := generatorOfLevelsCarryingARepeat(t)
	for tmpl, level := range map[string]string{"{/outer.inner as p}{p.a}": "p", "{/outer as o}{o.inner.a}": "o.inner"} {
		t.Run(tmpl, func(t *testing.T) {
			if _, err := f.NewTemplate(tmpl); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("the level %q carries a repeat", level)) {
				t.Errorf("NewTemplate(%s) = %v, want the level %s named", tmpl, err, level)
			}
		})
	}
}

func generatorOfChoicesUnderAPath(t *testing.T) *Generator {
	t.Helper()
	dir := writeData(t, map[string]string{
		"every":  `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2","g":"x"}]`,
		"notall": `[{"format":"{f}","f":"1"},"plain"]`,
		"some":   `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2","extra":"x"}]`,
	})
	return newGenerator(t, dir, WithSeed(1))
}

func TestPathThroughAChoiceDrawsWhatEveryVariantCarries(t *testing.T) {
	f := generatorOfChoicesUnderAPath(t)
	eachDraw(200, func() string { return fake(t, f, "every.f") }, func(got string) {
		if got != "1" && got != "2" {
			t.Fatalf("every.f = %q, want 1 or 2", got)
		}
	})
}

func TestPathOnlySomeVariantsCarryNamesWhatAllCarry(t *testing.T) {
	f := generatorOfChoicesUnderAPath(t)
	if _, err := f.Fake("some.extra"); err == nil || !strings.Contains(err.Error(), "all carry [f]") {
		t.Errorf("Fake(some.extra) = %v, want it to name what every variant carries", err)
	}
}

// TestPathNotEveryVariantCarriesFailsTheSameOnEveryCall pins the rule that keeps a
// dotted path from rendering on one call and failing on the next: every variant
// must carry the rest of the path.
func TestPathNotEveryVariantCarriesFailsTheSameOnEveryCall(t *testing.T) {
	f := generatorOfChoicesUnderAPath(t)
	failures := tallyDraws(200, func() string {
		_, err := f.Fake("notall.f")
		if err == nil {
			t.Fatal("notall.f = nil error, want the same failure every call")
		}
		return err.Error()
	})
	if len(failures) != 1 {
		t.Fatalf("notall.f error varies between calls: %v", failures)
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
	_, err := New(WithSeed(1), onTestDay, WithDataPath(dir))
	if err == nil || !strings.Contains(err.Error(), `field "a.b" contains "."`) {
		t.Fatalf("New = %v, want the dotted field name rejected", err)
	}
	// The same data without the dotted key is fine, and the path reaches a node.
	f := newGenerator(t, writeData(t, map[string]string{
		"cat": `{"format":"{a.b}","a":{"format":"{b}","b":"2"}}`,
	}), WithSeed(1), onTestDay)
	if !slices.Contains(f.List(), "cat.a.b") {
		t.Error("List() omits cat.a.b, which the data carries")
	}
	if got := fake(t, f, "cat"); got != "2" {
		t.Fatalf("cat = %q, want 2", got)
	}
}

// TestMissingFieldNamesItself keeps the precise diagnosis for the ordinary typo: a
// single-variant choice always draws the same item, so it needs no every-variant
// guard and the error can name the field that is missing.
func TestMissingFieldNamesItself(t *testing.T) {
	f := shipped(t, WithSeed(1))
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
	f := emptyGenerator(1)
	f.root = *tree.NewFolderOf(map[string]tree.Node{
		"addr": compiled(t, `{"format":"{street}","street":"Main"}`),
	})
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

func TestALeadingSlashIsThePath(t *testing.T) {
	f := newGenerator(t, writeData(t, map[string]string{"person": `{"format":"{first}","first":"Ada"}`}), WithSeed(1))
	if v := fake(t, f, "/person"); v != "Ada" {
		t.Errorf("Fake(/person) = %q, want Ada", v)
	}
	if _, err := f.Fake("//person"); err == nil || !strings.Contains(err.Error(), "write /person") {
		t.Errorf("Fake(//person) = %v, want it refused naming /person", err)
	}
	r, err := f.FakeRecord("/person")
	if err != nil || len(r.Columns()) != 1 || r.Columns()[0].Value != "Ada" {
		t.Errorf("FakeRecord(/person) = %+v, %v, want the record person", r, err)
	}
	if _, err := f.FakeRecord("//person"); err == nil || !strings.Contains(err.Error(), "write /person") {
		t.Errorf("FakeRecord(//person) = %v, want it refused naming /person", err)
	}
	if _, err := f.FakeRecord("person["); err == nil || !strings.Contains(err.Error(), "never closes") {
		t.Errorf("FakeRecord(person[) = %v, want the open selector named", err)
	}
	var v struct {
		A string `fake:"//person"`
	}
	if err := f.FakeStruct(&v); err == nil || !strings.Contains(err.Error(), "write /person") {
		t.Errorf("FakeStruct(//person tag) = %v, want it refused naming /person", err)
	}
}
