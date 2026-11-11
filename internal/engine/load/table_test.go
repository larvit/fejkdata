package load

import (
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
)

type unhandledNode struct{ tree.Node }

func TestRouteStepsRoutesOnlyOntoATablesNodes(t *testing.T) {
	m := map[string]any{"format": "{name}", "rows": "r.tsv", "key": "code"}
	tbl, err := compileTable(m, nil, "r", func(string) (string, error) { return "code\tname\n01\tA\n", nil })
	if err != nil {
		t.Fatalf("compileTable = %v", err)
	}
	route := func(n tree.Node) (panicked any) {
		defer func() { panicked = recover() }()
		_, _ = routeSteps(nil, nil, tbl, tableRoute{then: pathPos{n: n}}, 0)
		return nil
	}
	for kind, n := range map[string]tree.Node{"table": tbl, "row": tbl.RowNode(), "column": tbl.FormatTemplate().Fields()["name"]} {
		if r := route(n); r != nil {
			t.Errorf("routeSteps onto a %s panicked: %v", kind, r)
		}
	}
	for kind, n := range map[string]tree.Node{
		"choice":    compiled(t, `["a","b"]`),
		"folder":    tree.NewFolder(),
		"null":      &tree.NullItem{},
		"template":  compiled(t, `{"format":"{x}","x":"1"}`),
		"unhandled": &unhandledNode{},
	} {
		if route(n) == nil {
			t.Errorf("routeSteps onto a %s returned, want a panic", kind)
		}
	}
}
