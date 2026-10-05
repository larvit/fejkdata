package fejkdata

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
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

// Fake renders the template once.
func (t *Template) Fake() string {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderOnce(t.g.drawState, t.n)
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
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := bindInline(&f.root, n, "template", false); err != nil {
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

// IsTemplate reports whether arg is an inline template rather than a path, by its shape: an
// arg holding a {, or a valid JSON array or string, is a template, and anything else is a
// path. A path errors where it holds a } or a quote, starts with [, or is no valid path; so
// do a template of one reference alone, which is a path written as a template, and a path
// written with a leading /.
func IsTemplate(arg string) (bool, error) {
	inline, err := grammar.IsTemplate(arg)
	if err != nil {
		return false, fmt.Errorf("fejkdata: %w", err)
	}
	return inline, nil
}

func compileInput(input string) (node, error) {
	v, err := inputValue(input)
	if err != nil {
		return nil, err
	}
	return compile(v)
}

// inputValue reads an inline template as the value compile takes: the JSON value it holds, or
// the input itself as a format string when it is not JSON.
func inputValue(input string) (any, error) {
	var raw any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return input, nil
	}
	if trimmed := strings.TrimSpace(input); trimmed != input {
		return nil, fmt.Errorf("a JSON template may not be padded with spaces, which a format string would render; write %s", trimmed)
	}
	return raw, nil
}

// bindInline loads the shipped categories n reads, then binds n against root.
func bindInline(root *folder, n node, label string, typedByGo bool) error {
	scope := inlineScope(n, label)
	if err := refuseFolderRefs(scope); err != nil {
		return err
	}
	loadShipped(root, unloadedReads(root, nil, scope))
	return binding{
		scope:     scope,
		link:      func() error { return linkNodeRefs(scope, root.children) },
		typedByGo: typedByGo,
	}.bind()
}

// refuseFolderRefs refuses each reference in scope not written from the root, {/x}: an
// inline node sits in no folder.
func refuseFolderRefs(scope nodeScope) error {
	return scope(func(label string, m node) error {
		t, ok := m.(*template)
		if !ok {
			return nil
		}
		for _, name := range refTokens(t.tokens) {
			sigil, rest, err := grammar.RefShape(name)
			if err != nil {
				return fmt.Errorf("%s: reference {%s}: %w", label, name, err)
			}
			if sigil != "/" {
				return fmt.Errorf("%s: reference {%s}: an inline template has no folder; write {/%s}", label, name, rest)
			}
		}
		return nil
	})
}

// linkNodeRefs binds the references in an inline node's templates against the
// loaded tree.
func linkNodeRefs(scope nodeScope, root map[string]node) error {
	if err := scope(func(label string, m node) error {
		if t, ok := m.(*template); ok {
			return linkTemplate(nil, label, "", t, root)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, pass := range linkPasses {
		if err := scope(func(label string, m node) error {
			if t, ok := m.(*template); ok {
				return pass(label, t)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
