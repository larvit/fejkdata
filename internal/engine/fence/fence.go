// Package fence proves a finished node tree at load. It refuses a reference cycle, a repeat past
// the cap, a template that could render a value its datatype or its calc forbids, and one that two
// readers could read differently.
package fence

import (
	"reflect"

	"github.com/larvit/fejkdata/internal/engine/fence/names"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// NodeSet is the set of nodes one fence covers, walked by calling it with fn.
type NodeSet = func(fn func(label string, n tree.Node) error) error

// GoStruct is the Go struct a record was loaded for. Each field of Type sets the datatype of
// its column of Record, which a datatype option sets in a record loaded from data. Label names
// the struct in an error.
type GoStruct struct {
	Label  string
	Type   reflect.Type
	Record *tree.Template
}

// Check runs every fence over a finished load: nodes and their templates ts. Columns are typed
// by the data, or by typedBy where it is set.
// names.RefuseReadsInsideFields comes before the cycle check: a name read inside the field bound
// to it is also a reference cycle, and its error says how to fix the read. The cycle check comes
// before Templates and checkNodes, whose walks recurse unguarded. A struct's fields are checked
// last, once checkNodes has refused what no column may hold.
func Check(nodes NodeSet, ts []load.TemplateSite, typedBy *GoStruct) error {
	if err := names.RefuseReadsInsideFields(ts); err != nil {
		return err
	}
	if err := checkNoCycles(nodes); err != nil {
		return err
	}
	if err := Templates(ts); err != nil {
		return err
	}
	if typedBy == nil {
		if err := checkColumns(nodes); err != nil {
			return err
		}
		return checkNodes(nodes)
	}
	if err := checkNodes(nodes); err != nil {
		return err
	}
	return typedBy.checkFields()
}

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

// checkNodes runs the per-node fences over s, each over every node before the next,
// so which of several broken nodes is reported does not depend on the walk.
func checkNodes(s NodeSet) error {
	mem := RenderCounts{}
	if err := s(func(label string, n tree.Node) error { return RepeatCheck(label, n, mem) }); err != nil {
		return err
	}
	return s((&valueProof{}).checkDatatype)
}
