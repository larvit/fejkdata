package engine

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

func siteNodes(sites []categorySite) fence.NodeSet {
	return func(fn func(label string, n tree.Node) error) error {
		for _, s := range sites {
			if err := load.EachNode(s.n, s.path, fn); err != nil {
				return err
			}
		}
		return nil
	}
}

func inlineNodes(n tree.Node, label string) fence.NodeSet {
	return func(fn func(label string, m tree.Node) error) error { return load.EachNode(n, label, fn) }
}

// pipeline is the one load pipeline, run over a set of nodes and the finds the compile of those nodes began;
// its fields are where entry points differ.
type pipeline struct {
	nodes   fence.NodeSet
	finds   *tree.LoadFinds
	resolve func() ([]load.TemplateSite, error)
	// Set by a caller that proves the columns against their Go types itself.
	typedByGo bool
}

// categoryPipeline is the pipeline over the categories of one load, their references
// resolving against root.
func categoryPipeline(sites []categorySite, root map[string]tree.Node, l *tree.LoadFinds) pipeline {
	return pipeline{
		nodes: siteNodes(sites),
		finds: l,
		resolve: func() ([]load.TemplateSite, error) {
			if err := linkTables(sites); err != nil {
				return nil, err
			}
			return resolveCategoryTemplates(sites, root, l)
		},
	}
}

// run needs each stage to follow the one before. resolve binds what each template's tokens read, and
// refuses a name or reference that can't resolve. fence.RefuseReadsInsideFields comes before the cycle
// check: a name read inside the field bound to it is also a reference cycle, and its error says how
// to fix the read. The cycle check comes before settleRecords, fence.Templates and fence.CheckNodes,
// whose walks recurse unguarded. settleRecords fixes each record's columns, whose datatypes
// fence.CheckColumns proves agree. Until finish, each stage reads from p.finds what the stages
// before it found; from finish on, each reads the nodes.
func (p pipeline) run() error {
	ts, err := p.resolve()
	if err != nil {
		return err
	}
	if err := fence.RefuseReadsInsideFields(ts, p.finds); err != nil {
		return err
	}
	if err := fence.CheckNoCycles(p.nodes, p.finds); err != nil {
		return err
	}
	settleRecords(p.nodes, p.finds)
	p.finds.Finish()
	if err := fence.Templates(ts); err != nil {
		return err
	}
	if !p.typedByGo {
		if err := fence.CheckColumns(p.nodes); err != nil {
			return err
		}
	}
	return fence.CheckNodes(p.nodes)
}

func linkTables(sites []categorySite) error {
	for _, s := range sites {
		if t, isTable := s.n.(*tree.Table); isTable {
			if err := load.LinkTable(t, s.in.Children()); err != nil {
				return fmt.Errorf("%s: %w", s.path, err)
			}
		}
	}
	return nil
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]tree.Node, l *tree.LoadFinds) ([]load.TemplateSite, error) {
	var ts []load.TemplateSite
	for _, s := range sites {
		if err := load.EachNode(s.n, s.path, func(label string, n tree.Node) error {
			if t, isTemplate := n.(*tree.Template); isTemplate {
				ts = append(ts, load.TemplateSite{Template: t, Folder: s.dir, Label: label})
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if err := load.CompileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}

// settleRecords fixes the columns of every record in nodes. The pipeline runs it after fence.CheckNoCycles: a
// column's datatype walks the column, and a cycle would walk forever.
func settleRecords(nodes fence.NodeSet, l *tree.LoadFinds) {
	_ = nodes(func(_ string, n tree.Node) error {
		if t, isTemplate := n.(*tree.Template); isTemplate {
			l.SetColumns(t, recordColumns(t, l))
		}
		return nil
	})
}

// recordColumns is t's columns in name order where t is a record with fields, else nil.
func recordColumns(t *tree.Template, l *tree.LoadFinds) []tree.RecordColumn {
	if !t.IsRecord() || len(t.Fields()) == 0 {
		return nil
	}
	whole := map[tree.Node]*tree.NameBinding{}
	for _, tok := range t.Tokens() {
		if tok.Kind != grammar.NameBind {
			continue
		}
		b := l.Of(t).NameScope().Bindings()[tok.Bound]
		if target := l.Binding(b).Target(); b.BindsField() && len(target.Tail()) == 0 && whole[target.Start()] == nil {
			whole[target.Start()] = b
		}
	}
	var columns []tree.RecordColumn
	for _, name := range tree.SortedNames(t.Fields()) {
		if !t.Lets()[name] {
			field := t.Fields()[name]
			columns = append(columns, tree.NewRecordColumn(name, fence.ColumnDatatype(field, l), field, whole[field]))
		}
	}
	return columns
}
