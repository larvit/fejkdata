package load

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// compiled compiles a JSON fragment and resolves it against no data.
func compiled(t *testing.T, s string) tree.Node {
	t.Helper()
	v, err := jsonvalue.Decode([]byte(s))
	if err != nil {
		t.Fatalf("decode %q: %v", s, err)
	}
	l := tree.NewLoadFinds()
	n, err := Compile(v, l)
	if err != nil {
		t.Fatalf("compile %q: %v", s, err)
	}
	if _, err := ResolveInlineTemplates(n, "template", nil, l); err != nil {
		t.Fatalf("resolve %q: %v", s, err)
	}
	l.Finish()
	return n
}

func TestCompiledPathStopsAtAMissingSegment(t *testing.T) {
	n := compiled(t, `{"format":"{a}","a":{"format":"{b}","b":"leaf"}}`)
	if leaves := CompilePath(n, []string{"a", "b"}).Leaves; len(leaves) != 1 {
		t.Fatalf("compilePath(a.b).leaves = %v, want the one leaf", leaves)
	}
	w := &compiledPath{tail: []string{"a", "nope", "deeper"}}
	_, err := w.run(n)
	if err == nil || !strings.Contains(err.Error(), `no field "nope"`) {
		t.Errorf("walk(a.nope.deeper) = %v, want the missing segment named", err)
	}
	if len(w.Leaves) > 0 {
		t.Errorf("walk reached a leaf past a missing segment: %v", w.Leaves)
	}
}

func TestCompiledPathChoiceConsumesNoSegment(t *testing.T) {
	n := compiled(t, `[{"format":"{f}","f":"1"},{"format":"{f}","f":"2"}]`)
	if leaves := CompilePath(n, []string{"f"}).Leaves; len(leaves) != 2 {
		t.Fatalf("compilePath through a choice = %v, want both variants' f", leaves)
	}
}
