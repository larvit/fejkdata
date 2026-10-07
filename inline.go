package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/datafiles"
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

// Fake renders the template once; each call draws anew.
func (t *Template) Fake() string {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return renderOnce(t.g.drawState, t.n)
}

// NewTemplate compiles an inline template — a format string or a JSON value — and
// resolves its references against the loaded tree, so repeated renders pay the
// compile and validation once. It shares [New]'s guarantees: a bad template errors
// here, and rendering cannot fail.
func (f *Generator) NewTemplate(input string) (*Template, error) {
	n, err := compileInput(input)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := loadInline(&f.root, n, "template", false); err != nil {
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
// path. A path errors where it holds a } or a quote, starts with [, or is no valid path. A path
// may start with /, as a reference does, and every call taking a path drops it.
func IsTemplate(arg string) (bool, error) {
	inline, err := grammar.IsTemplate(arg)
	if err != nil {
		return false, fmt.Errorf("fejkdata: %w", err)
	}
	return inline, nil
}

func compileInput(input string) (node, error) {
	return compile(inputValue(input))
}

// inputValue reads an inline template as the value compile takes: the JSON value it holds, the
// whitespace around it dropped, or the input itself as a format string when it is not JSON.
func inputValue(input string) any {
	raw, err := datafiles.DecodeJSON([]byte(input))
	if err != nil {
		return input
	}
	return raw
}

// loadInline loads the shipped categories n reads, then runs n through the pipeline against root.
func loadInline(root *folder, n node, label string, typedByGo bool) error {
	nodes := inlineNodes(n, label)
	if err := refuseFolderRefs(nodes); err != nil {
		return err
	}
	loadShipped(root, unloadedReads(root, nil, nodes))
	return pipeline{
		nodes:     nodes,
		resolve:   func() error { return resolveInlineTemplates(nodes, root.children) },
		typedByGo: typedByGo,
	}.run()
}

// refuseFolderRefs refuses each reference in nodes not written from the root, {/x}: an
// inline node sits in no folder.
func refuseFolderRefs(nodes nodeSet) error {
	return nodes(func(label string, m node) error {
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
