package fejkdata

import (
	"encoding/json"
	"go/ast"
	"reflect"
	"slices"
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
			var frames frameStack
			if got := len(expand(f.drawState, tmpl, renderEnv{frames: &frames})); got < tmpl.compiled.grow {
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

// nodeSwitchSkips is the kinds a switch is never handed: its callers step past them first.
var nodeSwitchSkips = map[string][]string{
	"columnItems": {"folder", "table", "tableRow"},
	"prove":       {"folder"},
	"render":      {"folder"},
	"stepInto":    {"choice", "table", "tableRow"},
}

func TestNodeSwitchesHandleEveryKind(t *testing.T) {
	f := newGenerator(t, writeFiles(t, map[string]string{
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code"}`,
		"r.tsv":  "code\tname\n01\t{digits(2)}\n02\tB\n",
	}))
	tbl := f.root.children["r"].(*table)
	samples := map[string]node{}
	for _, n := range []node{
		compiled(t, `["a","b"]`),
		compiled(t, `{"format":"{x}","x":"1"}`),
		&f.root,
		&nullItem{},
		tbl,
		tbl.formatTemplate.fields["name"],
		tbl.rowNode,
	} {
		samples[reflect.TypeOf(n).Elem().Name()] = n
	}
	var kinds []string
	_, files := sourceFiles(t)
	for _, file := range files {
		for _, decl := range file.Decls {
			if d, isFunc := decl.(*ast.FuncDecl); isFunc && d.Name.Name == "isNode" {
				kinds = append(kinds, receiverType(d.Recv))
			}
		}
	}
	slices.Sort(kinds)
	sampled := make([]string, 0, len(samples))
	for kind := range samples {
		sampled = append(sampled, kind)
	}
	slices.Sort(sampled)
	if !slices.Equal(kinds, sampled) {
		t.Fatalf("node kinds %v, samples %v; give every kind a sample here", kinds, sampled)
	}
	for name, call := range map[string]func(node){
		"columnItems": func(n node) { columnItems(n) },
		"contained":   func(n node) { contained(n) },
		"paths":       func(n node) { paths(n, false) },
		"prove":       func(n node) { (&valueProof{}).prove(n) },
		"render": func(n node) {
			var frames frameStack
			render(engine(1).drawState, n, renderEnv{frames: &frames, row: renderedRow{tbl, 0}})
		},
		"renderEdges": func(n node) { renderEdges(n) },
		"stepInto":    func(n node) { _, _ = stepInto(n, "x") },
	} {
		mustPanic(t, name+" on an unhandled node", func() { call(&unhandledNode{}) })
		for _, kind := range kinds {
			if slices.Contains(nodeSwitchSkips[name], kind) {
				mustPanic(t, name+" on a skipped "+kind, func() { call(samples[kind]) })
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s on %s: %v", name, kind, r)
					}
				}()
				call(samples[kind])
			}()
		}
	}
	for kind, n := range samples {
		recurses := kind != "nullItem" && !slices.Contains(nodeSwitchSkips["render"], kind)
		if edges := renderEdges(n); recurses != (len(edges) > 0) {
			t.Errorf("%s: render recurses into it %v, renderEdges lists %d edges; a kind render recurses into has edges, and no other", kind, recurses, len(edges))
		}
	}
}
