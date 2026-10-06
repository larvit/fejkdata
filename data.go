package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
)

func loadData(sources []datafiles.Source) (map[string]node, error) {
	root := map[string]node{}
	for _, src := range sources {
		g := &folder{children: map[string]node{}}
		if err := src.Walk(compileInto(func(dir []string) *folder { return madeFolder(g, dir) })); err != nil {
			return nil, err
		}
		mergeChildren(root, g.children)
	}
	if len(root) == 0 {
		return nil, fmt.Errorf("no .json data found")
	}
	if err := categoryPipeline(categorySites(&folder{children: root}), root).run(); err != nil {
		return nil, err
	}
	return root, nil
}

// compileInto compiles each category it is handed into the folder place returns for its Folders.
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
		}
		g = sub
	}
	return g
}

// mergeChildren overlays src onto dst. Two folders under the same key merge
// recursively (so locales/categories from several paths combine); every other
// key is replaced, making the last-loaded directory win on a conflict.
func mergeChildren(dst, src map[string]node) {
	for k, v := range src {
		if dg, ok := dst[k].(*folder); ok {
			if sg, ok := v.(*folder); ok {
				mergeChildren(dg.children, sg.children)
				continue
			}
		}
		dst[k] = v
	}
}

// nodeSet is the set of nodes one validation pass covers: the categories one load
// reads, or a single inline node.
type nodeSet func(fn func(label string, n node) error) error

// categorySite is a loaded category and where it sits: the folder holding it, that
// folder's path, and its own dot path.
type categorySite struct {
	dir  []string
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
	return categorySite{dir: dir, path: categoryPath(dir, name), in: in, n: in.children[name]}
}

// categoryPath is the dot path of the category name in the folder dir.
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
