package fejkdata

import (
	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// loadSources merges sources in order, the last winning a clash. A source whose manifest
// carries an index leaves its categories unloaded; every other source loads here, with
// the unloaded categories it needs.
// docs/decisions.md#a-source-whose-manifest-carries-an-index-loads-each-category-on-first-reach-any-other-loads-in-new-with-what-it-reads
func loadSources(sources []datafiles.Source) (tree.Folder, error) {
	root := tree.NewFolder()
	l := tree.NewLoadFinds()
	for _, src := range sources {
		m, err := src.Manifest()
		if err != nil {
			return tree.Folder{}, err
		}
		if err := addSource(root, src, m, l); err != nil {
			return tree.Folder{}, err
		}
	}
	err := loadReached(root, categorySites(root), l, nil)
	return *root, err
}

func addSource(root *tree.Folder, src datafiles.Source, m datafiles.Manifest, l *tree.LoadFinds) error {
	src = src.Mounted(m.Namespace)
	if m.Index != nil {
		placeIndex(root, src, m.Index)
		return nil
	}
	g := tree.NewFolder()
	if err := src.Walk(compileInto(func(dir []string) *tree.Folder { return madeFolder(g, dir) }, l)); err != nil {
		return err
	}
	mergeFolder(root, g)
	return nil
}

func compileInto(place func(dir []string) *tree.Folder, l *tree.LoadFinds) func(datafiles.Category) error {
	return func(c datafiles.Category) error {
		n, err := load.CompileCategory(c, l)
		if err != nil {
			return err
		}
		place(c.Folders).Put(c.Name, n)
		return nil
	}
}

func madeFolder(root *tree.Folder, dir []string) *tree.Folder {
	g := root
	for _, seg := range dir {
		sub, isFolder := g.Children()[seg].(*tree.Folder)
		if !isFolder {
			sub = tree.NewFolder()
			g.Put(seg, sub)
		}
		g = sub
	}
	return g
}

func mergeFolder(dst, src *tree.Folder) {
	for k, v := range src.Children() {
		if dg, ok := dst.Children()[k].(*tree.Folder); ok {
			if sg, ok := v.(*tree.Folder); ok {
				mergeFolder(dg, sg)
				continue
			}
		}
		dst.Put(k, v)
	}
}

// nodeSet is the set of nodes one validation pass covers: one load's categories, or a
// single inline node.
type nodeSet func(fn func(label string, n tree.Node) error) error

// categorySite is a loaded category, where it sits and its own dot path.
type categorySite struct {
	categoryAt
	n    tree.Node
	path string
}

func categorySites(root *tree.Folder) []categorySite {
	var out []categorySite
	var walk func(dir []string, g *tree.Folder)
	walk = func(dir []string, g *tree.Folder) {
		for _, name := range tree.SortedNames(g.Children()) {
			if sub, isFolder := g.Children()[name].(*tree.Folder); isFolder {
				walk(append(dir[:len(dir):len(dir)], name), sub)
				continue
			}
			out = append(out, siteIn(dir, g, name))
		}
	}
	walk(nil, root)
	return out
}

func siteIn(dir []string, in *tree.Folder, name string) categorySite {
	return categorySite{categoryAt: categoryAt{dir: dir, in: in, name: name}, n: in.Children()[name], path: load.CategoryPath(dir, name)}
}

func siteNodes(sites []categorySite) nodeSet {
	return func(fn func(label string, n tree.Node) error) error {
		for _, s := range sites {
			if err := load.EachNode(s.n, s.path, fn); err != nil {
				return err
			}
		}
		return nil
	}
}

func inlineNodes(n tree.Node, label string) nodeSet {
	return func(fn func(label string, m tree.Node) error) error { return load.EachNode(n, label, fn) }
}

// pipeline is the one load pipeline, run over one nodeSet and the finds its compile began;
// its fields are where entry points differ.
type pipeline struct {
	nodes   nodeSet
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
