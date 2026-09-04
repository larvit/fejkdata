package fejkdata

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Template is an inline template compiled, referenced and validated against a
// generator's loaded data once, ready to render many times with [Template.Fake].
// It is safe for concurrent use: Fake serializes on its generator's lock, so a
// seeded sequence is reproducible only when a generator — and its templates — are
// drawn from one goroutine.
type Template struct {
	g *Generator
	n node
}

// Fake renders the template with one draw.
func (t *Template) Fake() string {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return render(t.g.rand, t.n, nil)
}

// NewTemplate compiles an inline template — a format string or a JSON value — and
// binds its references against the loaded tree, so repeated renders pay the
// compile and validation once. It shares [New]'s guarantees: a bad template errors
// here, and rendering cannot fail.
func (f *Generator) NewTemplate(input string) (*Template, error) {
	n, err := compileInput(input)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	scope := inlineScope(n)
	if err := linkNodeRefs(scope, f.categories); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	if err := checkScope(scope); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Template{g: f, n: n}, nil
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

// compileInput compiles an inline template: a JSON value, or a bare format string
// when the input is not JSON.
func compileInput(input string) (node, error) {
	var raw any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return compile(input)
	}
	if trimmed := strings.TrimSpace(input); trimmed != input {
		return nil, fmt.Errorf("a JSON template may not be padded with spaces, which a format string would render; write %s", trimmed)
	}
	return compile(raw)
}

// linkNodeRefs binds the references in an inline node's templates against the
// loaded tree.
func linkNodeRefs(scope nodeScope, root map[string]node) error {
	return scope(func(path string, m node) error {
		t, ok := m.(*template)
		if !ok {
			return nil
		}
		for _, name := range refTokens(t.format) {
			sigil, rest, err := refShape(name)
			if err != nil {
				return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
			}
			if sigil != "/" {
				return fmt.Errorf("%s: reference {%s}: an inline template has no folder; write {/%s}", path, name, rest)
			}
		}
		return linkTemplateRefs(nil, path, t, root)
	})
}
