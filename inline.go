package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/grammar"
)

// Template is an inline template compiled, referenced and validated against a
// generator's loaded data once, ready to render many times with [Template.Fake].
// It is safe for concurrent use: Fake serializes on its generator's lock, so a
// sequence is reproducible only when one goroutine makes every draw from a generator
// and its templates.
type Template struct {
	g        *Generator
	template engine.Template
}

// Fake renders the template once; each call draws anew.
func (t *Template) Fake() string {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return t.template.Draw(t.g.drawState)
}

// NewTemplate compiles an inline template — a format string or a JSON value — and
// resolves its references against the loaded tree, so repeated renders pay the
// compile and validation once. It shares [New]'s guarantees: a bad template errors
// here, and rendering cannot fail.
func (f *Generator) NewTemplate(input string) (*Template, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	template, err := engine.LoadTemplate(&f.root, input, "template")
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Template{g: f, template: template}, nil
}

// FakeTemplate compiles and renders an inline template in one call. It is
// [NewTemplate] then [Template.Fake]; to render the same template many times, hold
// the *Template and call its Fake.
func (f *Generator) FakeTemplate(input string) (string, error) {
	t, err := f.NewTemplate(input)
	if err != nil {
		return "", err
	}
	return t.Fake(), nil
}

// IsTemplate reports whether arg is an inline template rather than a path, by its shape: an
// arg holding a {, or a valid JSON array or string, is a template, and anything else is a
// path. A path errors where it holds a } or a quote, starts with [, or is no valid path. A path
// may start with /, as a reference does, and every call taking a path drops it.
func IsTemplate(arg string) (bool, error) {
	inline, err := grammar.IsTemplate(arg)
	if err != nil {
		return false, fmt.Errorf("fejkdata: %w", err)
	}
	return inline, nil
}
