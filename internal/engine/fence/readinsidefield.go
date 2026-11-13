package fence

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// RefuseReadsInsideFields refuses, in the templates of ts, a read of a name inside the field bound
// to it.
func RefuseReadsInsideFields(ts []load.TemplateSite) error {
	for _, s := range ts {
		for _, r := range load.NamedReads(s.Template, tree.AfterFinish) {
			b := r.Read.Named()
			if !b.BindsField() {
				continue
			}
			target := b.Target()
			if rendersInside(load.CompilePath(target.Start(), target.Tail()).Leaves, s.Template) {
				return fmt.Errorf("%s: token {%s}: name %q is read inside %q, the field bound to it; read the name outside that field", s.Label, r.Op.Token().Body, b.Name(), b.Ref())
			}
		}
	}
	return nil
}

func rendersInside(nodes []tree.Node, t *tree.Template) bool {
	seen := map[tree.Node]bool{}
	var inside func(n tree.Node) bool
	inside = func(n tree.Node) bool {
		if n == tree.Node(t) {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		for _, c := range load.Contained(n) {
			if inside(c.Node) {
				return true
			}
		}
		return false
	}
	for _, n := range nodes {
		if inside(n) {
			return true
		}
	}
	return false
}
