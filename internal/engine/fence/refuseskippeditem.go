package fence

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/tree"
)

// refuseSkippedItem refuses a read through a name bound to a field of its template, where it steps into an
// item of a choice and what it renders reads a name that item binds. The read opens no frame for the
// item, so the name has no pick there, unless the template lies in the item and its frame is already open.
func refuseSkippedItem(nr nameRead) error {
	read := nr.read
	if !read.Named().BindsField() {
		return nil
	}
	for _, r := range nr.routes {
		x := r.a
		if x.Kind() != tree.NamedRead {
			continue
		}
		scope := x.Named().Scope()
		if scope.Kind() != tree.ScopeChoiceItem || nr.in.NameScope().Within(scope) {
			continue
		}
		for _, leaf := range read.Leaves() {
			if _, steppedInto := nr.names.pathTo(scope.Owner(), leaf); steppedInto {
				return fmt.Errorf("{%s} steps into an item of a choice without rendering it, so name %q, bound in that item, has no pick there; bind %q outside the choice, or read {%s} whole",
					read.Spelling(), x.Named().Name(), x.Named().Name(), read.Named().Name())
			}
		}
	}
	return nil
}
