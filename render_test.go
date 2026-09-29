package fejkdata

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRecursionHasNoDepthLimit(t *testing.T) {
	// Build {format:{a}, a:[{format:{a}, a:[ ... "deep" ]]}} 50 levels deep.
	tmpl := `"deep"`
	for i := 0; i < 50; i++ {
		tmpl = `{"format":"{a}","a":` + tmpl + `}`
	}
	if got := mustRender(t, engine(1), tmpl); got != "deep" {
		t.Fatalf("deep recursion = %q, want deep", got)
	}
}

// TestGrowIsALowerBound pins the property that makes the pre-sized render buffer
// safe: grow must never exceed what expand emits. render multiplies it by repeat,
// so an over-estimate would amplify up to a million-fold.
func TestGrowIsALowerBound(t *testing.T) {
	f := engine(3)
	for _, format := range []string{
		"",
		"plain literal",
		"{digits(2)}-{int(1,9)}{int(1,9)}-{upper(2)}-{lower(2)}",
		"01Aa# literal",
		"{{}} {{{x}}}",
		"Ö dag åäö 日本語",
		"{x}{x}{x}",
		"{hex(8)}-{int(10,99)}-{nanoid(5)}",
		"9{d}{luhn()}",
		"{a|b} and {a|b}",
	} {
		src := `{"format":` + quote(format) + `,"x":"1","a":"A","b":"B","d":"012345678901234"}`
		tmpl, ok := compiled(t, src).(*template)
		if !ok {
			t.Fatalf("format %q did not compile to a template", format)
		}
		for i := 0; i < 50; i++ {
			var set holdSet
			if got := len(expand(f.rand, tmpl, renderScope{set: &set})); got < tmpl.compiled.grow {
				t.Errorf("format %q: expand emitted %d bytes, below grow %d", format, got, tmpl.compiled.grow)
			}
		}
	}
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type unhandledNode struct{}

func (*unhandledNode) isNode() {}

func TestNodeSwitchesPanicOnAnUnhandledNode(t *testing.T) {
	for name, call := range map[string]func(){
		"contained":   func() { contained(&unhandledNode{}) },
		"paths":       func() { paths(&unhandledNode{}) },
		"render":      func() { render(nil, &unhandledNode{}, renderScope{}) },
		"renderEdges": func() { renderEdges(&unhandledNode{}) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil || !strings.HasPrefix(fmt.Sprint(r), "fejkdata: internal error: ") || !strings.Contains(fmt.Sprint(r), "*fejkdata.unhandledNode") {
					t.Errorf("recovered %v, want an internal error naming the node kind", r)
				}
			}()
			call()
		})
	}
}
