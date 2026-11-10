package fejkdata

import (
	"encoding/json"
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

// nodeSwitchSkips names, per switch, the kinds it is never handed.
var nodeSwitchSkips = map[string][]string{
	"childNamed":       {"choice", "table", "tableRow"},
	"columnItems":      {"folder", "table", "tableRow"},
	"render":           {"folder"},
	"routeSteps":       {"choice", "folder", "nullItem", "template"},
	"valueProof.prove": {"folder"},
}

// nodeSamples holds one node of each kind, by the kind's name, and the table they came from.
func nodeSamples(t *testing.T) (map[string]node, *table) {
	t.Helper()
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
	return samples, tbl
}

// nodeSwitches calls each function holding a type switch over node kinds.
func nodeSwitches(tbl *table) map[string]func(node) {
	return map[string]func(node){
		"childNamed":  func(n node) { _, _ = childNamed(n, "x") },
		"columnItems": func(n node) { columnItems(n) },
		"contained":   func(n node) { contained(n) },
		"paths":       func(n node) { paths(n, false) },
		"render": func(n node) {
			var frames frameStack
			render(engine(1).drawState, n, renderEnv{frames: &frames, row: renderedRow{tbl, 0}})
		},
		"renderEdges":      func(n node) { renderEdges(n, nil) },
		"routeSteps":       func(n node) { _, _ = routeSteps(nil, nil, tbl, tableRoute{then: pathPos{n: n}}, 0) },
		"valueProof.prove": func(n node) { (&valueProof{}).prove(n) },
	}
}

func TestNodeSwitchesPanicOnAnUnhandledNode(t *testing.T) {
	_, tbl := nodeSamples(t)
	for name, call := range nodeSwitches(tbl) {
		mustPanic(t, name+" on an unhandled node", func() { call(&unhandledNode{}) })
	}
}

func TestNodeSwitchesHandleEveryKind(t *testing.T) {
	samples, tbl := nodeSamples(t)
	for name, call := range nodeSwitches(tbl) {
		for kind, n := range samples {
			if slices.Contains(nodeSwitchSkips[name], kind) {
				mustPanic(t, name+" on a skipped "+kind, func() { call(n) })
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s on %s: %v", name, kind, r)
					}
				}()
				call(n)
			}()
		}
	}
}

func TestRenderEdgesListWhatRenderRecursesInto(t *testing.T) {
	samples, _ := nodeSamples(t)
	for kind, n := range samples {
		recurses := kind != "nullItem" && !slices.Contains(nodeSwitchSkips["render"], kind)
		if edges := renderEdges(n, nil); recurses != (len(edges) > 0) {
			t.Errorf("%s: render recurses into it %v, renderEdges lists %d edges; a kind render recurses into has edges, and no other", kind, recurses, len(edges))
		}
	}
}
