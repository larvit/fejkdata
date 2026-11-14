// Package names refuses, at load, a read of a name that could not keep one pick.
package names

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// RefuseReads refuses, in the templates of ts, a read of a name that could disagree with another
// read of it, or that reads a name bound in a choice's item the read never renders.
func RefuseReads(ts []load.TemplateSite) error {
	names := namesOf(ts)
	for _, s := range ts {
		for _, r := range load.NamedReads(s.Template, tree.AfterFinish) {
			if err := checkNameRead(s.Template, r.Read, names); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.Label, r.Op.Token().Body, err)
			}
		}
	}
	return nil
}

// nameRead is one read of a name, with what each refusal rule reads of it.
type nameRead struct {
	in     *tree.Template // the template holding read
	read   *tree.Read
	pick   pickReads     // every read of the name, read included
	routes []route       // every read made while read renders
	names  resolvedNames // every name of the load
}

// nameRules is the rules that refuse a read of a name, in the order checkNameRead applies them. A
// name keeps one pick (goal 4.3), and a template that can't keep that fails to load (goal 4.6).
// The README lists what the rules refuse, under Names, "Refused at `New`".
//
// Four of them refuse a read that could disagree with another read of the name: both reach one
// level, each by its own route, and the two routes can draw apart. One, refuseSkippedItem,
// refuses a read that has no pick to read.
//   - refuseTwiceDrawn: the pick renders a field twice
//   - refuseDrawnApart: a nested name has a pick of its own
//   - refuseSkippedItem: the read steps into a choice's item without opening the item's frame
//   - refuseFreshDraw: a reference draws afresh
//   - refuseKeptAndFreshField: a field read the pick doesn't keep draws afresh
//
// A sixth rule, RefuseReadsInsideFields, refuses a name read inside the field bound to it. The
// load pipeline calls it before the cycle check, so such a read never reaches these.
//
// Each rule decides alone, so the order changes no verdict. A read that breaks two rules gets the
// error of the first, and an error is a contract by what it names: keep the order.
// docs/decisions.md#an-error-is-a-contract-by-what-it-names-not-its-bytes
var nameRules = []func(nameRead) error{
	refuseTwiceDrawn,
	refuseDrawnApart,
	refuseSkippedItem,
	refuseFreshDraw,
	refuseKeptAndFreshField,
}

// checkNameRead refuses read, a read in t of a name, by the first of nameRules it breaks.
func checkNameRead(t *tree.Template, read *tree.Read, names resolvedNames) error {
	pick := names.pick(read.Named())
	nr := nameRead{in: t, read: read, pick: pick, routes: routesUnder(read.Leaves(), read.Key(), pick.addressed), names: names}
	for _, refuse := range nameRules {
		if err := refuse(nr); err != nil {
			return err
		}
	}
	return nil
}

// secondDraw is how to spell by, a read of b, so it draws afresh.
func secondDraw(b *tree.NameBinding, by string) string {
	return b.Ref() + strings.TrimPrefix(by, b.Name())
}
