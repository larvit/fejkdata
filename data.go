package fejkdata

import (
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
)

// loadSources merges sources in order, the last winning a clash. A source whose manifest
// carries an index leaves its categories unloaded; every other source loads here, with
// the unloaded categories it needs.
// docs/decisions.md#a-source-whose-manifest-carries-an-index-loads-each-category-on-first-reach-any-other-loads-in-new-with-what-it-reads
func loadSources(sources []datafiles.Source) (folder, error) {
	root := folder{children: map[string]node{}}
	walked := map[string][]string{}
	for _, src := range sources {
		if err := addSource(&root, src, walked); err != nil {
			return folder{}, err
		}
	}
	sites := categorySites(&root)
	for i := range sites {
		sites[i].defaultModules = walked[sites[i].path]
	}
	return root, loadReached(&root, sites, nil)
}

// addSource places src's categories in root. For each category it loads now, it records in
// walked the modules src reads by default.
func addSource(root *folder, src datafiles.Source, walked map[string][]string) error {
	m, err := src.Manifest()
	if err != nil {
		return err
	}
	if m.Index != nil {
		placeIndex(root, src, m)
		return nil
	}
	g := &folder{children: map[string]node{}}
	if err := src.Walk(compileInto(func(dir []string) *folder { return madeFolder(g, dir) })); err != nil {
		return err
	}
	for _, s := range categorySites(g) {
		walked[s.path] = m.Reads
	}
	mergeFolder(root, g)
	return nil
}

func compileInto(place func(dir []string) *folder) func(datafiles.Category) error {
	return func(c datafiles.Category) error {
		n, err := compileCategory(c)
		if err != nil {
			return err
		}
		place(c.Folders).put(c.Name, n)
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
			g.put(seg, sub)
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
		dst.put(k, v)
	}
}

// nodeSet is the set of nodes one validation pass covers: one load's categories, or a
// single inline node.
type nodeSet func(fn func(label string, n node) error) error

// categorySite is a loaded category, where it sits, its own dot path, and the modules its
// source reads by default.
type categorySite struct {
	categoryAt
	n              node
	path           string
	defaultModules []string
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
	return categorySite{categoryAt: categoryAt{dir: dir, in: in, name: name}, n: in.children[name], path: categoryPath(dir, name)}
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
