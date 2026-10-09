package fejkdata

import (
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
)

// loadSources merges sources in order, the last winning a clash. A source whose manifest
// carries an index leaves its categories unloaded; every other source loads here. So do
// the unloaded categories that depend on one it loaded or on one a later source replaced,
// so that a reader such a change breaks fails New, not a first reach.
func loadSources(sources []datafiles.Source) (folder, error) {
	root := folder{children: map[string]node{}}
	var changed []string
	for _, src := range sources {
		replaced, err := addSource(&root, src)
		if err != nil {
			return folder{}, err
		}
		changed = append(changed, replaced...)
	}
	sites := categorySites(&root)
	for _, s := range sites {
		changed = append(changed, s.path)
	}
	return root, loadReached(&root, sites, dependents(&root, changed))
}

// addSource adds the categories of src to root, and returns the path of each category
// already there that src replaced or put out of reach.
func addSource(root *folder, src datafiles.Source) ([]string, error) {
	m, err := src.Manifest()
	if err != nil {
		return nil, err
	}
	before := standing(root)
	if m.Index != nil {
		placeIndex(root, src, m.Index)
	} else {
		g := &folder{children: map[string]node{}}
		if err := src.Walk(compileInto(func(dir []string) *folder { return madeFolder(g, dir) })); err != nil {
			return nil, err
		}
		mergeFolder(root, g)
	}
	if len(before) == 0 {
		return nil, nil
	}
	after := standing(root)
	var replaced []string
	for p, was := range before {
		if after[p] != was {
			replaced = append(replaced, p)
		}
	}
	return replaced, nil
}

// standing is what stands at the path of every category under root: its node, or its
// unloaded entry.
func standing(root *folder) map[string]any {
	out := map[string]any{}
	for _, s := range categorySites(root) {
		out[s.path] = s.n
	}
	for p, c := range unloadedUnder(root) {
		out[p] = c.in.unloaded[c.name]
	}
	return out
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

// categorySite is a loaded category and where it sits: the folder holding it, that
// folder's path, its name and its own dot path.
type categorySite struct {
	dir  []string
	in   *folder
	n    node
	name string
	path string
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
	return categorySite{dir: dir, in: in, n: in.children[name], name: name, path: categoryPath(dir, name)}
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
