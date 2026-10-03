package fejkdata

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:generate env REGENERATE=1 go test -run ^TestShippedIndexIsCurrent$ .

//go:embed data
var shippedFS embed.FS

var shippedSource = dataSource{fsys: shippedFS, baseDir: "data"}

// shippedEntry is a shipped category's entry in shippedindex.go: its parent table, "" for
// none, and the paths List advertises below it. After changing it, empty
// shippedIndex's literal by hand so the package builds, then run generate.
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

// unloadedCategory is a shipped category not loaded yet.
type unloadedCategory struct {
	dir  []string
	in   *folder
	name string
}

// unloadedAt is the unloaded shipped category a path from root names, or descends into.
func unloadedAt(root *folder, segs []string) (unloadedCategory, bool) {
	g, i := folderAt(root, segs)
	if i == len(segs) {
		return unloadedCategory{}, false
	}
	if _, unloaded := g.unloaded[segs[i]]; !unloaded {
		return unloadedCategory{}, false
	}
	return unloadedCategory{dir: segs[:i:i], in: g, name: segs[i]}, true
}

// loadShippedAt loads the shipped category a caller's path names or descends into. An entry
// point calls it before walking root: a walk is a query and loads nothing.
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

// loadShipped loads the shipped categories wanted, every category they read, and the table
// family of each one it loads, then binds them as a whole load binds.
// TestEveryShippedCategoryLoadsAlone reaches each category, so a failure here means a stale
// index.
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
			panic(internalError("shipped %s: %v; after a change under data/, regenerate shippedindex.go", join(strings.Join(u.dir, "."), u.name), err))
		}
		sites = append(sites, site)
		queue = append(queue, unloadedReads(root, u.dir, sitesScope([]categorySite{site}))...)
		queue = append(queue, u.linkedTables(e)...)
	}
	if len(sites) == 0 {
		return
	}
	if err := categoryBinding(sites, root.children).bind(); err != nil {
		panic(internalError("binding shipped categories: %v; after a change under data/, regenerate shippedindex.go", err))
	}
}

// load parses and compiles the category, and moves it from its folder's unloaded map to
// the folder's children.
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
	return siteIn(u.dir, u.in, u.name), nil
}

// linkedTables is u's parent beside it, and every table beside it whose parent u is.
func (u unloadedCategory) linkedTables(e shippedEntry) []unloadedCategory {
	var out []unloadedCategory
	if e.parent != "" {
		out = append(out, unloadedCategory{dir: u.dir, in: u.in, name: e.parent})
	}
	for _, name := range sortedNames(u.in.unloaded) {
		if u.in.unloaded[name].parent == u.name {
			out = append(out, unloadedCategory{dir: u.dir, in: u.in, name: name})
		}
	}
	return out
}
