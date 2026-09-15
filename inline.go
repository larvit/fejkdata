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
	return renderOnce(t.g.rand, t.n)
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
	if err := bindInline(n, "template", f.categories, checkScope); err != nil {
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

// IsTemplate reports whether arg is an inline template rather than a path, by its shape: a {
// token, or a JSON object, array or string, is a template, and anything else is a path. A
// name never holds a bracket, a brace or a quote, so an arg holding one that is not valid
// JSON names neither, and errors; so do a template of one reference alone, which is a path
// written as a template, and a path written with a leading /.
func IsTemplate(arg string) (bool, error) {
	inline, err := isTemplate(arg)
	if err != nil {
		return false, fmt.Errorf("fejkdata: %w", err)
	}
	return inline, nil
}

func isTemplate(arg string) (bool, error) {
	if strings.ContainsRune(arg, '{') || (isJSONStart(strings.TrimSpace(arg)) && json.Valid([]byte(arg))) {
		if path, lone := loneReference(arg); lone {
			return false, pathAdvice(path, fmt.Sprintf("%s is the path %s written as a template", arg, path))
		}
		return true, nil
	}
	if i := strings.IndexAny(arg, `[]}"`); i >= 0 {
		return false, fmt.Errorf("%q holds a %q, which no path may, and it is not valid JSON, so it names no template either", arg, arg[i:i+1])
	}
	if path := strings.TrimLeft(arg, "/"); path != arg && path != "" {
		return false, pathAdvice(path, fmt.Sprintf("path %s starts with /, and every path starts at the root already", arg))
	}
	return false, nil
}

func isJSONStart(arg string) bool {
	return strings.HasPrefix(arg, "[") || strings.HasPrefix(arg, `"`)
}

// pathAdvice refuses a spelling of path, naming path to write, or why no name can spell it.
func pathAdvice(path, refusal string) error {
	if err := checkPathNames(path); err != nil {
		return err
	}
	return fmt.Errorf("%s; write %s", refusal, path)
}

// loneReference is the path a template spells when the value it holds — a format string, a
// JSON string, or an object holding only a format — is one reference token and nothing else.
func loneReference(arg string) (string, bool) {
	var raw any
	if json.Unmarshal([]byte(arg), &raw) != nil {
		raw = arg
	}
	if m, isObject := raw.(map[string]any); isObject && len(m) == 1 {
		raw = m["format"]
	}
	format, isString := raw.(string)
	body, lone := loneRef(format)
	if !isString || !lone {
		return "", false
	}
	if strings.HasPrefix(body, "/") {
		body = "/" + strings.TrimLeft(body, "/")
	}
	_, path, err := refShape(body)
	return path, err == nil
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

// bindInline links an inline node's references against root and runs check over it, naming its
// nodes from label.
func bindInline(n node, label string, root map[string]node, check func(nodeScope) error) error {
	scope := inlineScope(n, label)
	if err := linkNodeRefs(scope, root); err != nil {
		return err
	}
	return check(scope)
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
