package fejkdata

import (
	"strconv"
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
	f, err := New(append([]Option{withShipped(), WithSeed(1), onTestDay}, opts...)...)
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
	seen := tallyDraws(50, func() string { return tmpl(t, f, `{"format":"name: {x}","x":["bosse","lina"]}`) })
	if seen["name: bosse"] == 0 || seen["name: lina"] == 0 || len(seen) != 2 {
		t.Fatalf("JSON template produced %v, want both names", seen)
	}
}

func TestFakeTemplateJSONArray(t *testing.T) {
	f := shipped(t)
	seen := tallyDraws(50, func() string { return tmpl(t, f, `["foo","bar","baz"]`) })
	if len(seen) != 3 {
		t.Fatalf("JSON array choice produced %v, want three items", seen)
	}
}

func TestFakeTemplateNamedReference(t *testing.T) {
	dir := writeData(t, map[string]string{
		"person": `[{"format":"{first} {last}","first":"Ada","last":"Lovelace"},{"format":"{first} {last}","first":"Bo","last":"Ek","born":"1990"}]`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	eachDraw(100, func() string { return tmpl(t, f, "{/person as p}{p.first} {p.last}") }, func(got string) {
		if got != "Ada Lovelace" && got != "Bo Ek" {
			t.Fatalf("correlated references = %q, want one person's first and last", got)
		}
	})
}

func TestFakeTemplateDeterministic(t *testing.T) {
	a, b := shipped(t), shipped(t)
	in := "row: {/misc.uuid} {digits(3)}"
	eachDraw(20, func() string { return tmpl(t, a, in) }, func(x string) {
		if y := tmpl(t, b, in); x != y {
			t.Fatalf("same seed diverged: %q != %q", x, y)
		}
	})
}

func TestFieldlessTokenHint(t *testing.T) {
	f := shipped(t)
	_, err := f.FakeTemplate(`{sv_SE.person.last}`)
	if err == nil || !strings.Contains(err.Error(), "write {/sv_SE.person.last}") {
		t.Fatalf("FakeTemplate(bare token) = %v, want a hint naming {/sv_SE.person.last}", err)
	}
}

func TestFieldlessObjectHintDoesNotCallItABareString(t *testing.T) {
	_, err := shipped(t).FakeTemplate(`{"format":"{x}","repeat":2}`)
	if err == nil || !strings.Contains(err.Error(), `this template has none — write {/x}`) {
		t.Errorf("FakeTemplate(fieldless object) = %v, want the hint without calling it a bare string", err)
	}
}

func TestFieldlessTokenHintIsOfferedOnlyWhereTheNameIsTheWholeToken(t *testing.T) {
	f := shipped(t)
	// A hint is only a drop-in where the name is the whole token: {/x} inside a
	// transform or an alternation renders a different value, so none is offered.
	for _, input := range []string{`{ /sv_SE.person.last }`, "{lowercase(x)}", "{x|y}"} {
		t.Run(input, func(t *testing.T) {
			_, err := f.FakeTemplate(input)
			if err == nil || strings.Contains(err.Error(), "write {") {
				t.Errorf("FakeTemplate(%q) = %v, want no hint naming a spelling that means something else", input, err)
			}
		})
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
		{`"{digits(-1)}"`, "must not be negative"},
		{`name: {/no.such.path}`, "no entry"},
		{`name: {..nope}`, "write {/nope}"},
	} {
		t.Run(c.input, func(t *testing.T) {
			_, err := f.FakeTemplate(c.input)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("FakeTemplate(%q) = %v, want an error containing %q", c.input, err, c.want)
			}
		})
	}
}

func TestFakeTemplateJSONString(t *testing.T) {
	f := shipped(t)
	got := tmpl(t, f, `"name: {/sv_SE.person.last}"`)
	if !strings.HasPrefix(got, "name: ") || strings.Contains(got, "{") {
		t.Fatalf("FakeTemplate(JSON string) = %q, want a rendered last name after the prefix", got)
	}
}

func TestPaddedJSONIsTheJSON(t *testing.T) {
	f := shipped(t)
	in := `{"format":"{x}","x":["a","b"]}`
	padded, err := f.NewTemplate(" \t" + in + "\n ")
	if err != nil {
		t.Fatalf("NewTemplate(padded JSON) = %v, want the JSON it pads", err)
	}
	if got := padded.Fake(); got != "a" && got != "b" {
		t.Fatalf("padded JSON rendered %q, want a or b with no padding", got)
	}
}

func TestNewTemplateReusable(t *testing.T) {
	f := shipped(t)
	reusable, err := f.NewTemplate(`{digits(2)}`)
	if err != nil {
		t.Fatalf("NewTemplate: %v", err)
	}
	seen := tallyDraws(50, reusable.Fake)
	if len(seen) < 2 {
		t.Fatalf("Template.Fake() repeated %v, want varied draws from one compile", seen)
	}
}

func TestIsTemplate(t *testing.T) {
	for arg, want := range map[string]bool{
		"sv_SE.person":                           false,
		"person.last":                            false,
		"name: {x}":                              true,
		`{"format":"x"}`:                         true,
		`["a","b"]`:                              true,
		`[1, 2]`:                                 true,
		` ["a","b"]`:                             true, // padding is the template's own error, not a shape verdict
		`"hello"`:                                true,
		"{/a}{/b}":                               true,
		"{/a|/b}":                                true,
		"{uppercase(/a)}":                        true,
		"{{/a}}":                                 true,
		`"{/a} x"`:                               true,
		"x[1]":                                   false, // a row selected by key or name
		"x[St. Louis].y":                         false,
		"{/sv_SE.person.last}":                   true,
		`"{/sv_SE.person}"`:                      true,
		"{.person.last}":                         true,
		` "{/sv_SE.person}"`:                     true,
		`{"format":"{/sv_SE.person}"}`:           true,
		"{//sv_SE.person}":                       true,
		"{/misc.currency[US Dollar (Next day)]}": true,
		"/sv_SE.person":                          false,
	} {
		t.Run(arg, func(t *testing.T) {
			if got, err := IsTemplate(arg); err != nil || got != want {
				t.Errorf("IsTemplate(%q) = %v, %v; want %v", arg, got, err, want)
			}
		})
	}
}

func TestIsTemplateRejectsAMalformedArgument(t *testing.T) {
	for arg, want := range map[string]string{
		"[abc]":          `starts with "["`,
		"[abc].field":    `starts with "["`,
		"x[1]y":          `"]"`,
		"x[[1]]":         `"["`,
		"x[]":            "empty",
		"//sv_SE.person": "write /sv_SE.person",
		"a]b":            `holds a "]"`,
		"a}b":            `holds a "}"`,
		`"abc`:           `holds a "\""`,
		`"a]b`:           `holds a "\""`, // the opener the reader typed, not the bracket behind it
	} {
		t.Run(arg, func(t *testing.T) {
			if _, err := IsTemplate(arg); err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("IsTemplate(%q) = %v; want it rejected naming %s", arg, err, want)
			}
		})
	}
}

func TestFakeTemplateRepeatBound(t *testing.T) {
	f := shipped(t)
	_, err := f.FakeTemplate(`{"format":"{x}","repeat":200,"x":{"format":"{y}","repeat":200,"y":{"format":"z","repeat":200}}}`)
	if err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Errorf("nested repeat over the cap = %v, want it rejected naming the maximum", err)
	}
}

func TestIsLonePath(t *testing.T) {
	for body, want := range map[string]bool{
		"x":                                   true,
		"a.b":                                 true,
		"sv_SE.person":                        true,
		"a..":                                 true,
		"a..b":                                true,
		"misc.currency[US Dollar (Next day)]": true,
		"t[a as b].c":                         true,
		"t[f(x)]":                             true,
		"":                                    false,
		"/x":                                  false,
		".x":                                  false,
		"a.":                                  false,
		"a...b":                               false,
		"a|b":                                 false,
		"x(1)":                                false,
		"w where length 5":                    false,
		"x as p":                              false,
		"x}}":                                 false,
		"x} {/y":                              false,
	} {
		t.Run(strconv.Quote(body), func(t *testing.T) {
			if got := IsLonePath(body); got != want {
				t.Errorf("IsLonePath(%q) = %v, want %v", body, got, want)
			}
		})
	}
}
