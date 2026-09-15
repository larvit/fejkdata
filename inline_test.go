package fejkdata

import (
	"strings"
	"testing"
)

func tmpl(t *testing.T, f *Generator, input string) string {
	t.Helper()
	s, err := f.FakeTemplate(input)
	if err != nil {
		t.Fatalf("FakeTemplate(%q): %v", input, err)
	}
	return s
}

func shipped(t *testing.T, opts ...Option) *Generator {
	t.Helper()
	f, err := New(append([]Option{WithSeed(1)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFakeTemplateFormatString(t *testing.T) {
	f := shipped(t)
	got := tmpl(t, f, "name: {/sv_SE.person.last}")
	if !strings.HasPrefix(got, "name: ") || strings.HasSuffix(got, " ") || strings.Contains(got, "{") {
		t.Fatalf("FakeTemplate = %q, want a rendered last name after the prefix", got)
	}
}

func TestFakeTemplateJSONObject(t *testing.T) {
	f := shipped(t)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[tmpl(t, f, `{"format":"name: {x}","x":["bosse","lina"]}`)] = true
	}
	if !seen["name: bosse"] || !seen["name: lina"] || len(seen) != 2 {
		t.Fatalf("JSON template produced %v, want both names", seen)
	}
}

func TestFakeTemplateJSONArray(t *testing.T) {
	f := shipped(t)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[tmpl(t, f, `["foo","bar","baz"]`)] = true
	}
	if len(seen) != 3 {
		t.Fatalf("JSON array choice produced %v, want three items", seen)
	}
}

func TestFakeTemplateCorrelatedReferences(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 100; i++ {
		got := tmpl(t, f, "{/person.first} {/person.last}")
		if got != "Ada Lovelace" && got != "Bo Ek" {
			t.Fatalf("correlated references = %q, want one person's first and last", got)
		}
	}
}

func TestFakeTemplateDeterministic(t *testing.T) {
	a, b := shipped(t), shipped(t)
	for i := 0; i < 20; i++ {
		in := "row: {/misc.uuid} {digits(3)}"
		if x, y := tmpl(t, a, in), tmpl(t, b, in); x != y {
			t.Fatalf("same seed diverged: %q != %q", x, y)
		}
	}
}

func TestFieldlessTokenHint(t *testing.T) {
	f := shipped(t)
	_, err := f.FakeTemplate(`{sv_SE.person.last}`)
	if err == nil || !strings.Contains(err.Error(), "write {/sv_SE.person.last}") {
		t.Fatalf("FakeTemplate(bare token) = %v, want a hint naming {/sv_SE.person.last}", err)
	}
	_, err = f.FakeTemplate(`{"format":"{x}","repeat":2}`)
	if err == nil || !strings.Contains(err.Error(), `this template has none — write {/x}`) {
		t.Errorf("FakeTemplate(fieldless object) = %v, want the hint without calling it a bare string", err)
	}
	// A hint is only a drop-in where the name is the whole token: {/x} inside a
	// transform or an alternation renders a different value, so none is offered.
	for _, input := range []string{`{ /sv_SE.person.last }`, "{lowercase(x)}", "{x|y}"} {
		_, err := f.FakeTemplate(input)
		if err == nil || strings.Contains(err.Error(), "write {") {
			t.Errorf("FakeTemplate(%q) = %v, want no hint naming a spelling that means something else", input, err)
		}
	}
}

func TestFakeTemplateErrors(t *testing.T) {
	f := shipped(t)
	for _, c := range []struct {
		input string
		want  string
	}{
		{`{"x":"Q"}`, "missing string \"format\""},
		{`"{x}"`, `no field "x"`},
		{`"{digits(0)}"`, "must be positive"},
		{`name: {/no.such.path}`, "no entry"},
		{`name: {..nope}`, "write {/nope}"},
		{`{"format":"x"}`, "is a string"},
		{`{/misc.country} {/misc.country.alpha2}`, "renders a level"},
		{`{"format":"{/misc.country.alpha2} {x}","x":"{/misc.country}"}`, "reads a path into"},
	} {
		_, err := f.FakeTemplate(c.input)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("FakeTemplate(%q) = %v, want an error containing %q", c.input, err, c.want)
		}
	}
}

func TestFakeTemplateJSONString(t *testing.T) {
	f := shipped(t)
	got := tmpl(t, f, `"name: {/sv_SE.person.last}"`)
	if !strings.HasPrefix(got, "name: ") || strings.Contains(got, "{") {
		t.Fatalf("FakeTemplate(JSON string) = %q, want a rendered last name after the prefix", got)
	}
}

func TestPaddedJSONIsRejected(t *testing.T) {
	f := shipped(t)
	in := `{"format":"{x}","x":["a","b"]}`
	_, err := f.NewTemplate("  " + in + "  ")
	if err == nil || !strings.Contains(err.Error(), "write "+in) {
		t.Fatalf("NewTemplate(padded JSON) = %v, want an error naming the unpadded spelling", err)
	}
}

func TestNewTemplateReusable(t *testing.T) {
	f := shipped(t)
	reusable, err := f.NewTemplate(`{digits(2)}`)
	if err != nil {
		t.Fatalf("NewTemplate: %v", err)
	}
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[reusable.Fake()] = true
	}
	if len(seen) < 2 {
		t.Fatalf("Template.Fake() repeated %v, want varied draws from one compile", seen)
	}
}

func TestIsTemplate(t *testing.T) {
	for arg, want := range map[string]bool{
		"sv_SE.person":    false,
		"person.last":     false,
		"name: {x}":       true,
		`{"format":"x"}`:  true,
		`["a","b"]`:       true,
		`[1, 2]`:          true,
		` ["a","b"]`:      true, // padding is the template's own error, not a shape verdict
		`"hello"`:         true,
		"{/a}{/b}":        true,
		"{/a|/b}":         true,
		"{uppercase(/a)}": true,
		"{{/a}}":          true,
		`"{/a} x"`:        true,
	} {
		if got, err := IsTemplate(arg); err != nil || got != want {
			t.Errorf("IsTemplate(%q) = %v, %v; want %v", arg, got, err, want)
		}
	}
	for arg, want := range map[string]string{
		"[abc]":                `holds a "["`,
		"[abc].field":          `holds a "["`,
		"x[1]":                 `holds a "["`,
		"a]b":                  `holds a "]"`,
		"a}b":                  `holds a "}"`,
		`"abc`:                 `holds a "\""`,
		`"a]b`:                 `holds a "\""`, // the opener the reader typed, not the bracket behind it
		"{/sv_SE.person.last}": "{/sv_SE.person.last} is the path sv_SE.person.last written as a template; write sv_SE.person.last",
		`"{/sv_SE.person}"`:    "write sv_SE.person",
		"{.person.last}":       "write person.last",
		"/sv_SE.person":        "write sv_SE.person",
	} {
		if _, err := IsTemplate(arg); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("IsTemplate(%q) = %v; want it rejected naming %s", arg, err, want)
		}
	}
}

func TestFakeTemplateRepeatBound(t *testing.T) {
	f := shipped(t)
	_, err := f.FakeTemplate(`{"format":"{x}","repeat":200,"x":{"format":"{y}","repeat":200,"y":{"format":"z","repeat":200}}}`)
	if err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Errorf("nested repeat over the cap = %v, want it rejected naming the maximum", err)
	}
}
