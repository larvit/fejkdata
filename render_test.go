package fejkdata

import (
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/render"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestRecursionHasNoDepthLimit(t *testing.T) {
	tmpl := strings.Repeat(`{"format":"{a}","a":`, 50) + `"deep"` + strings.Repeat("}", 50)
	if got := mustRender(t, emptyGenerator(1), tmpl); got != "deep" {
		t.Fatalf("deep recursion = %q, want deep", got)
	}
}

// TestGrowIsALowerBound pins the property that makes the pre-sized render buffer
// safe: grow must never exceed what one render emits. A repeat multiplies it,
// so an over-estimate would amplify up to a million-fold.
func TestGrowIsALowerBound(t *testing.T) {
	f := emptyGenerator(3)
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
		t.Run(strconv.Quote(format), func(t *testing.T) {
			src := `{"format":` + quote(format) + `,"x":"1","a":"A","b":"B","d":"012345678901234"}`
			tmpl, ok := compiled(t, src).(*tree.Template)
			if !ok {
				t.Fatalf("format %q did not compile to a template", format)
			}
			fejktest.EachDraw(50, func() string { return render.Once(f.drawState, tmpl) }, func(rendered string) {
				if got := len(rendered); got < tmpl.Compiled().Grow() {
					t.Errorf("format %q: a render emitted %d bytes, below grow %d", format, got, tmpl.Compiled().Grow())
				}
			})
		})
	}
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type unhandledNode struct{ tree.Node }

// nodeSwitchSkips names, per switch, the kinds it is never handed.
var nodeSwitchSkips = map[string][]string{
	"ChildNamed":  {"Choice", "Table", "TableRow"},
	"ColumnItems": {"Folder", "Table", "TableRow"},
	"render":      {"Folder"},
	"Prove":       {"Folder"},
}

// nodeSamples holds one node of each kind, by the kind's name, and the table they came from.
func nodeSamples(t *testing.T) (map[string]tree.Node, *tree.Table) {
	t.Helper()
	f := newGenerator(t, writeFiles(t, map[string]string{
		"r.json": `{"format":"{name}","rows":"r.tsv","key":"code"}`,
		"r.tsv":  "code\tname\n01\t{digits(2)}\n02\tB\n",
	}))
	tbl := f.root.Children()["r"].(*tree.Table)
	samples := map[string]tree.Node{}
	for _, n := range []tree.Node{
		compiled(t, `["a","b"]`),
		compiled(t, `{"format":"{x}","x":"1"}`),
		&f.root,
		&tree.NullItem{},
		tbl,
		tbl.FormatTemplate().Fields()["name"],
		tbl.RowNode(),
	} {
		samples[reflect.TypeOf(n).Elem().Name()] = n
	}
	return samples, tbl
}

// nodeSwitches calls each function holding a type switch over node kinds.
func nodeSwitches(tbl *tree.Table) map[string]func(tree.Node) {
	return map[string]func(tree.Node){
		"ChildNamed":  func(n tree.Node) { _, _ = load.ChildNamed(n, "x") },
		"ColumnItems": func(n tree.Node) { load.ColumnItems(n) },
		"Contained":   func(n tree.Node) { load.Contained(n) },
		"Paths":       func(n tree.Node) { load.Paths(n, false) },
		"render": func(n tree.Node) {
			var pins tree.PinSet
			s := emptyGenerator(1).drawState
			tbl.Rows().DrawIn(s, &pins)
			render.Pinned(s, n, &pins)
		},
		"RenderEdges": func(n tree.Node) { tree.RenderEdges(n) },
		"Prove":       func(n tree.Node) { fence.Prove(n) },
	}
}

func TestNodeSwitchesPanicOnAnUnhandledNode(t *testing.T) {
	_, tbl := nodeSamples(t)
	for name, call := range nodeSwitches(tbl) {
		t.Run(name, func(t *testing.T) {
			mustPanic(t, name+" on an unhandled node", func() { call(&unhandledNode{}) })
		})
	}
}

type nodeSwitchCall struct {
	name, kind string
	call       func()
}

func nodeSwitchCalls(t *testing.T, skipped bool) []nodeSwitchCall {
	t.Helper()
	samples, tbl := nodeSamples(t)
	var calls []nodeSwitchCall
	for name, call := range nodeSwitches(tbl) {
		for kind, n := range samples {
			if slices.Contains(nodeSwitchSkips[name], kind) == skipped {
				calls = append(calls, nodeSwitchCall{name, kind, func() { call(n) }})
			}
		}
	}
	return calls
}

func TestNodeSwitchesHandleEveryKind(t *testing.T) {
	for _, c := range nodeSwitchCalls(t, false) {
		t.Run(c.name+" on "+c.kind, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s on %s: %v", c.name, c.kind, r)
				}
			}()
			c.call()
		})
	}
}

func TestNodeSwitchesPanicOnAKindTheyAreNeverHanded(t *testing.T) {
	for _, c := range nodeSwitchCalls(t, true) {
		t.Run(c.name+" on "+c.kind, func(t *testing.T) {
			mustPanic(t, c.name+" on a skipped "+c.kind, c.call)
		})
	}
}

func TestRenderEdgesListWhatRenderRecursesInto(t *testing.T) {
	samples, _ := nodeSamples(t)
	for kind, n := range samples {
		t.Run(kind, func(t *testing.T) {
			recurses := kind != "NullItem" && !slices.Contains(nodeSwitchSkips["render"], kind)
			if edges := tree.RenderEdges(n); recurses != (len(edges) > 0) {
				t.Errorf("%s: render recurses into it %v, RenderEdges lists %d edges; a kind render recurses into has edges, and no other", kind, recurses, len(edges))
			}
		})
	}
}
