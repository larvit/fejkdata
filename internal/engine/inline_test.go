package engine

import (
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

func compileInput(t *testing.T, input string) (tree.Node, *tree.LoadFinds) {
	t.Helper()
	l := tree.NewLoadFinds()
	n, err := load.Compile(InputValue(input), l)
	if err != nil {
		t.Fatal(err)
	}
	return n, l
}

func wroteTemplateOrAField(root *tree.Template) bool {
	written := root.Finished()
	for _, field := range root.Fields() {
		switch field := field.(type) {
		case *tree.Template:
			written = written || field.Finished()
		case *tree.Choice:
			written = written || field.NameScope() != nil
		}
	}
	return written
}

func TestALoadFailingBeforeFinishWritesNoNode(t *testing.T) {
	for src, want := range map[string]string{
		`{"format":"{w as e}{reword(e,e)}","w":["ab","cd"]}`: "is a name, which holds one pick",
	} {
		t.Run(src, func(t *testing.T) {
			n, l := compileInput(t, src)
			if err := loadInline(tree.NewFolder(), n, l, "t", false); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("loadInline() = %v, want an error holding %q", err, want)
			}
			if wroteTemplateOrAField(n.(*tree.Template)) {
				t.Error("a load that failed before finish wrote its template or a field")
			}
		})
	}
}

func TestFinishWritesEachNodeOnce(t *testing.T) {
	n, l := compileInput(t, `{"format":"{/w as n}{n} {c}","c":["x","y"]}`)
	w, err := LoadTemplate(tree.NewFolder(), `["p","q"]`, "w")
	if err != nil {
		t.Fatal(err)
	}
	if err := loadInline(tree.NewFolderOf(map[string]tree.Node{"w": w.n}), n, l, "t", false); err != nil {
		t.Fatal(err)
	}
	root := n.(*tree.Template)
	if !root.Finished() || len(root.Compiled().Ops()) == 0 || root.Fields()["c"].(*tree.Choice).NameScope() == nil {
		t.Fatal("a load that passed left its template, its ops or its choice unwritten")
	}
	if b := root.NameScope().Bindings()["n"]; !b.Finished() || b.Target().Start() != w.n {
		t.Fatal("a load that passed left its binding unwritten")
	}
	defer func() {
		if recover() == nil {
			t.Error("a second finish wrote the nodes again, want a panic")
		}
	}()
	l.Finish()
}
