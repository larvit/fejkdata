// Package fence proves a finished node tree at load. It refuses a reference cycle, a repeat past
// the cap, a template that could render a value its datatype or its calc forbids, and one that two
// readers could read differently.
package fence

import (
	"github.com/larvit/fejkdata/internal/engine/fence/names"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// NodeSet is the set of nodes one fence covers, walked by calling it with fn.
type NodeSet = func(fn func(label string, n tree.Node) error) error

// Templates runs the template fences over the finished templates of ts.
func Templates(ts []load.TemplateSite) error {
	for _, fence := range []func([]load.TemplateSite) error{
		proveDrawLengths, refuseSelfReads, refuseFilteredTails, names.RefuseReads, proveCalcsAndAddDays,
	} {
		if err := fence(ts); err != nil {
			return err
		}
	}
	return nil
}

func proveCalcsAndAddDays(ts []load.TemplateSite) error {
	proof := &valueProof{}
	for _, s := range ts {
		if err := proof.checkCalcs(s.Label, s.Template); err != nil {
			return err
		}
		if err := proveAddDays(s.Label, s.Template); err != nil {
			return err
		}
	}
	return nil
}

// CheckNodes runs the per-node fences over s, each over every node before the next,
// so which of several broken nodes is reported does not depend on the walk.
func CheckNodes(s NodeSet) error {
	mem := RenderCounts{}
	if err := s(func(label string, n tree.Node) error { return RepeatCheck(label, n, mem) }); err != nil {
		return err
	}
	return s((&valueProof{}).checkDatatype)
}
