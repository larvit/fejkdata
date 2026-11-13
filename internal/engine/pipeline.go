package engine

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
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

// run loads, then fences. resolve binds what each template's tokens read, and refuses a name or
// reference that can't resolve. load.SettleRecords fixes each record's columns, and Finish writes
// what the load found into the nodes, so every fence reads finished nodes. fence.RefuseReadsInsideFields
// comes before the cycle check: a name read inside the field bound to it is also a reference cycle,
// and its error says how to fix the read. The cycle check comes before fence.Templates and
// fence.CheckNodes, whose walks recurse unguarded.
func (p pipeline) run() error {
	ts, err := p.resolve()
	if err != nil {
		return err
	}
	load.SettleRecords(p.nodes, p.finds)
	p.finds.Finish()
	if err := fence.RefuseReadsInsideFields(ts); err != nil {
		return err
	}
	if err := fence.CheckNoCycles(p.nodes); err != nil {
		return err
	}
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
