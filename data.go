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
	if err := categoryBinding(categorySites(&folder{children: root}), root).bind(); err != nil {
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

// nodeScope is the set of nodes one validation pass covers: the categories one load
// binds, or a single inline node.
type nodeScope func(fn func(path string, n node) error) error

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
	return categorySite{dir: dir, path: join(strings.Join(dir, "."), name), in: in, n: in.children[name]}
}

func sitesScope(sites []categorySite) nodeScope {
	return func(fn func(path string, n node) error) error {
		for _, s := range sites {
			if err := eachNode(s.n, s.path, fn); err != nil {
				return err
			}
		}
		return nil
	}
}

func inlineScope(n node, label string) nodeScope {
	return func(fn func(path string, m node) error) error { return eachNode(n, label, fn) }
}

// binding is one scope's way through bind, the one load pipeline; its fields are where
// scopes differ.
type binding struct {
	scope nodeScope
	link  func() error
	// Set by a caller that proves the columns against their Go types itself.
	typedByGo bool
}

// categoryBinding binds the categories of one load, their references resolving
// against root.
func categoryBinding(sites []categorySite, root map[string]node) binding {
	return binding{
		scope: sitesScope(sites),
		link: func() error {
			if err := linkTables(sites); err != nil {
				return err
			}
			return linkRefs(sites, root)
		},
	}
}

func (b binding) bind() error {
	if err := b.link(); err != nil {
		return err
	}
	if err := checkNoCycles(b.scope); err != nil {
		return err
	}
	if !b.typedByGo {
		if err := checkColumns(b.scope); err != nil {
			return err
		}
	}
	return checkNodeFences(b.scope)
}

// checkNodeFences runs the per-node fences every binding needs over a scope, each
// over the whole scope before the next, so which of several broken nodes is reported
// does not depend on the walk. Its walks recurse unguarded, so bind refuses the scope's cycles
// first.
func checkNodeFences(s nodeScope) error {
	mem := renderCounts{}
	if err := s(func(path string, n node) error { return repeatCheck(path, n, mem) }); err != nil {
		return err
	}
	return s((&valueProof{}).checkDatatype)
}
