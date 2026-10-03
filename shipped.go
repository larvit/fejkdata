package fejkdata

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed data
var shippedFS embed.FS

var shippedSource = dataSource{fsys: shippedFS, baseDir: "data"}

// shippedEntry is a shipped category's entry in shippedindex.go: the table beside it that
// it links to, "" for none, and the paths List advertises below it.
type shippedEntry struct {
	parent string
	paths  []string
}

// unloadedTree builds the folders of the shipped index, with every category in them unloaded.
func unloadedTree() folder {
	root := folder{children: map[string]node{}}
	for p, e := range shippedIndex {
		segs := strings.Split(p, ".")
		g := &root
		for _, seg := range segs[:len(segs)-1] {
			sub, isFolder := g.children[seg].(*folder)
			if !isFolder {
				sub = &folder{children: map[string]node{}}
				g.children[seg] = sub
			}
			g = sub
		}
		if g.unloaded == nil {
			g.unloaded = map[string]shippedEntry{}
		}
		g.unloaded[segs[len(segs)-1]] = e
	}
	return root
}

// unloadedCategory is a shipped category not loaded yet: the folder holding it, that
// folder's path, and its name there.
type unloadedCategory struct {
	dir  []string
	in   *folder
	name string
}

// unloadedAt is the unloaded shipped category a path from root names, or descends into.
func unloadedAt(root *folder, segs []string) (unloadedCategory, bool) {
	g := root
	for i, seg := range segs {
		switch n := g.children[seg].(type) {
		case *folder:
			g = n
			continue
		case nil:
			if _, unloaded := g.unloaded[seg]; unloaded {
				return unloadedCategory{dir: segs[:i:i], in: g, name: seg}, true
			}
		}
		return unloadedCategory{}, false
	}
	return unloadedCategory{}, false
}

// loadShippedAt loads the shipped category a caller's path names or descends into, unless
// it is loaded already.
func (f *Generator) loadShippedAt(segs []string) {
	if u, unloaded := unloadedAt(&f.root, segs); unloaded {
		loadShipped(&f.root, []unloadedCategory{u})
	}
}

// unloadedReads is every unloaded shipped category the templates of scope reference, each
// reference read from the folder dir.
func unloadedReads(root *folder, dir []string, scope nodeScope) []unloadedCategory {
	var out []unloadedCategory
	_ = scope(func(_ string, n node) error {
		t, isTemplate := n.(*template)
		if !isTemplate {
			return nil
		}
		for _, name := range refTokens(t.tokens) {
			segs, err := refSegments(name, dir)
			if err != nil {
				continue // the link reports it
			}
			if u, unloaded := unloadedAt(root, segs); unloaded {
				out = append(out, u)
			}
		}
		return nil
	})
	return out
}

// loadShipped loads the shipped categories wanted, with every category they read and the
// whole table family of each, then binds what it loaded, as New binds a whole load.
// TestEveryShippedCategoryLoadsAlone proves each closure, so a failure is the index's.
func loadShipped(root *folder, wanted []unloadedCategory) {
	var sites []categorySite
	for queue := wanted; len(queue) > 0; queue = queue[1:] {
		u := queue[0]
		e, unloaded := u.in.unloaded[u.name]
		if !unloaded {
			continue
		}
		site, err := u.load()
		if err != nil {
			panic(internalError("shipped %s: %v", u.name, err))
		}
		sites = append(sites, site)
		queue = append(queue, unloadedReads(root, u.dir, sitesScope([]categorySite{site}))...)
		queue = append(queue, u.family(e)...)
	}
	if len(sites) == 0 {
		return
	}
	sort.Slice(sites, func(i, j int) bool { return sites[i].path < sites[j].path })
	if err := categoryBinding(sites, root.children).bind(); err != nil {
		panic(internalError("binding shipped categories: %v", err))
	}
}

// load parses and compiles the category, and moves it from its folder's unloaded
// categories to its children.
func (u unloadedCategory) load() (categorySite, error) {
	dir := path.Join(append([]string{shippedSource.baseDir}, u.dir...)...)
	entries, err := fs.ReadDir(shippedSource.fsys, dir)
	if err != nil {
		return categorySite{}, fmt.Errorf("%s: %w", dir, err)
	}
	file := u.name + ".json"
	if err := loadFile(shippedSource, u.in, path.Join(dir, file), file, newCategoryFiles(shippedSource, dir, entries)); err != nil {
		return categorySite{}, err
	}
	delete(u.in.unloaded, u.name)
	return categorySite{dir: u.dir, path: join(strings.Join(u.dir, "."), u.name), in: u.in, n: u.in.children[u.name]}, nil
}

// family is the tables beside u that a table family links it to: its parent, and every
// table whose parent it is.
func (u unloadedCategory) family(e shippedEntry) []unloadedCategory {
	var out []unloadedCategory
	if e.parent != "" {
		out = append(out, unloadedCategory{dir: u.dir, in: u.in, name: e.parent})
	}
	for name, sibling := range u.in.unloaded {
		if sibling.parent == u.name {
			out = append(out, unloadedCategory{dir: u.dir, in: u.in, name: name})
		}
	}
	return out
}
