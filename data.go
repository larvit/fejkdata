package fejkdata

import (
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
)

// loadSources merges sources in order, the last winning a clash. A source whose manifest
// carries an index leaves its categories unloaded; every other source's categories load
// here, with what they reach and every unloaded category leading to one of them or to one a
// later source replaced.
func loadSources(sources []datafiles.Source) (folder, error) {
	root := folder{children: map[string]node{}}
	targets := map[string]bool{}
	for _, src := range sources {
		before := categoriesUnder(&root)
		provides, err := place(&root, src)
		if err != nil {
			return folder{}, err
		}
		if len(before) == 0 {
			continue
		}
		after := categoriesUnder(&root)
		for p := range before {
			if provides(p) || !after[p] {
				targets[p] = true
			}
		}
	}
	sites := categorySites(&root)
	for _, s := range sites {
		targets[s.path] = true
	}
	return root, loadReached(&root, sites, reaching(&root, targets))
}

// place adds the categories of src to root, and returns whether src provides a path.
func place(root *folder, src datafiles.Source) (func(string) bool, error) {
	m, err := src.Manifest()
	if err != nil {
		return nil, err
	}
	if m.Index != nil {
		placeIndex(root, src, m.Index)
		return func(p string) bool { _, ok := m.Index[p]; return ok }, nil
	}
	g := &folder{children: map[string]node{}}
	if err := src.Walk(compileInto(func(dir []string) *folder { return madeFolder(g, dir) })); err != nil {
		return nil, err
	}
	mergeFolder(root, g)
	provided := categoriesUnder(g)
	return func(p string) bool { return provided[p] }, nil
}

// categoriesUnder is the path of every category under root, loaded or not.
func categoriesUnder(root *folder) map[string]bool {
	out := map[string]bool{}
	for _, s := range categorySites(root) {
		out[s.path] = true
	}
	unloaded := map[string]categoryAt{}
	unloadedUnder(root, nil, unloaded)
	for p := range unloaded {
		out[p] = true
	}
	return out
}

func compileInto(place func(dir []string) *folder) func(datafiles.Category) error {
	return func(c datafiles.Category) error {
		n, err := compileCategory(c)
		if err != nil {
			return err
		}
		place(c.Folders).children[c.Name] = n
		return nil
	}
}

// madeFolder is the folder that dir names under root, made where it is missing or a category.
func madeFolder(root *folder, dir []string) *folder {
	g := root
	for _, seg := range dir {
		sub, isFolder := g.children[seg].(*folder)
		if !isFolder {
			sub = &folder{children: map[string]node{}}
			g.children[seg] = sub
			delete(g.unloaded, seg)
		}
		g = sub
	}
	return g
}

func mergeFolder(dst, src *folder) {
	for k, v := range src.children {
		if dg, ok := dst.children[k].(*folder); ok {
			if sg, ok := v.(*folder); ok {
				mergeFolder(dg, sg)
				continue
			}
		}
		dst.children[k] = v
		delete(dst.unloaded, k)
	}
}

// nodeSet is the set of nodes one validation pass covers: one load's categories, or a
// single inline node.
type nodeSet func(fn func(label string, n node) error) error

// categorySite is a loaded category and where it sits: the folder holding it, that
// folder's path, its name and its own dot path.
type categorySite struct {
	dir  []string
	name string
	path string
	in   *folder
	n    node
}

// categorySites lists every category under root, folders and names in sorted order.
func categorySites(root *folder) []categorySite {
	var out []categorySite
	var walk func(dir []string, g *folder)
	walk = func(dir []string, g *folder) {
		for _, name := range sortedNames(g.children) {
			if sub, isFolder := g.children[name].(*folder); isFolder {
				walk(append(dir[:len(dir):len(dir)], name), sub)
				continue
			}
			out = append(out, siteIn(dir, g, name))
		}
	}
	walk(nil, root)
	return out
}

func siteIn(dir []string, in *folder, name string) categorySite {
	return categorySite{dir: dir, name: name, path: categoryPath(dir, name), in: in, n: in.children[name]}
}

func categoryPath(dir []string, name string) string { return join(strings.Join(dir, "."), name) }

func siteNodes(sites []categorySite) nodeSet {
	return func(fn func(label string, n node) error) error {
		for _, s := range sites {
			if err := eachNode(s.n, s.path, fn); err != nil {
				return err
			}
		}
		return nil
	}
}

func inlineNodes(n node, label string) nodeSet {
	return func(fn func(label string, m node) error) error { return eachNode(n, label, fn) }
}

// pipeline is the one load pipeline, run over one nodeSet; its fields are where
// entry points differ.
type pipeline struct {
	nodes   nodeSet
	resolve func() error
	// Set by a caller that proves the columns against their Go types itself.
	typedByGo bool
}

// categoryPipeline is the pipeline over the categories of one load, their references
// resolving against root.
func categoryPipeline(sites []categorySite, root map[string]node) pipeline {
	return pipeline{
		nodes: siteNodes(sites),
		resolve: func() error {
			if err := linkTables(sites); err != nil {
				return err
			}
			return resolveCategoryTemplates(sites, root)
		},
	}
}

func (p pipeline) run() error {
	if err := p.resolve(); err != nil {
		return err
	}
	if err := checkNoCycles(p.nodes); err != nil {
		return err
	}
	settleRecords(p.nodes)
	if !p.typedByGo {
		if err := checkColumns(p.nodes); err != nil {
			return err
		}
	}
	return checkNodeFences(p.nodes)
}

// checkNodeFences runs the per-node fences over s, each over every node before the next,
// so which of several broken nodes is reported does not depend on the walk. Its walks
// recurse unguarded, so pipeline.run refuses cycles first.
func checkNodeFences(s nodeSet) error {
	mem := renderCounts{}
	if err := s(func(label string, n node) error { return repeatCheck(label, n, mem) }); err != nil {
		return err
	}
	return s((&valueProof{}).checkDatatype)
}
