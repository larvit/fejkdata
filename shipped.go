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

// shippedCategory is a shipped category as shippedindex.go lists it: the table beside it
// that it links to, "" for none, and the paths List advertises below it.
type shippedCategory struct {
	parent string
	paths  []string
}

// shelve builds the folders of the shipped index, with every category in them unloaded.
func shelve() folder {
	root := folder{children: map[string]node{}}
	for p, c := range shippedIndex {
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
			g.unloaded = map[string]shippedCategory{}
		}
		g.unloaded[segs[len(segs)-1]] = c
	}
	return root
}

// pendingLoad is an unloaded shipped category: the folder holding it, that folder's path,
// and its name there.
type pendingLoad struct {
	dir  []string
	in   *folder
	name string
}

// unloadedAt is the unloaded shipped category a path from root names, or descends into.
func unloadedAt(root *folder, segs []string) (pendingLoad, bool) {
	g := root
	for i, seg := range segs {
		switch n := g.children[seg].(type) {
		case *folder:
			g = n
			continue
		case nil:
			if _, unloaded := g.unloaded[seg]; unloaded {
				return pendingLoad{dir: segs[:i:i], in: g, name: seg}, true
			}
		}
		return pendingLoad{}, false
	}
	return pendingLoad{}, false
}

// loadShippedAt loads the shipped category a caller's path names or descends into, unless
// it is loaded already.
func (f *Generator) loadShippedAt(segs []string) error {
	if p, unloaded := unloadedAt(&f.root, segs); unloaded {
		return loadShipped(&f.root, []pendingLoad{p})
	}
	return nil
}

// shippedReads is every unloaded shipped category the templates of scope reference, each
// reference read from the folder dir.
func shippedReads(root *folder, dir []string, scope nodeScope) []pendingLoad {
	var out []pendingLoad
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
			if p, unloaded := unloadedAt(root, segs); unloaded {
				out = append(out, p)
			}
		}
		return nil
	})
	return out
}

// loadShipped loads the shipped categories wanted, with every category they read and the
// whole table family of each, then binds what it loaded, as New binds a whole load. A failed
// bind leaves every category it loaded unloaded again.
func loadShipped(root *folder, wanted []pendingLoad) (err error) {
	var loaded []pendingLoad
	var entries []shippedCategory
	defer func() {
		if err == nil {
			return
		}
		for i, p := range loaded {
			delete(p.in.children, p.name)
			p.in.unloaded[p.name] = entries[i]
		}
	}()
	var sites []categorySite
	for queue := wanted; len(queue) > 0; queue = queue[1:] {
		p := queue[0]
		c, unloaded := p.in.unloaded[p.name]
		if !unloaded {
			continue
		}
		site, err := p.load()
		if err != nil {
			return err
		}
		loaded, entries = append(loaded, p), append(entries, c)
		sites = append(sites, site)
		queue = append(queue, shippedReads(root, p.dir, sitesScope([]categorySite{site}))...)
		queue = append(queue, p.family(c)...)
	}
	sort.Slice(sites, func(i, j int) bool { return sites[i].path < sites[j].path })
	return categoryBinding(sites, root.children).bind()
}

// load parses and compiles the category, and moves it from its folder's unloaded
// categories to its children.
func (p pendingLoad) load() (categorySite, error) {
	dir := path.Join(append([]string{shippedSource.baseDir}, p.dir...)...)
	entries, err := fs.ReadDir(shippedSource.fsys, dir)
	if err != nil {
		return categorySite{}, fmt.Errorf("%s: %w", dir, err)
	}
	file := p.name + ".json"
	if err := loadFile(shippedSource, p.in, path.Join(dir, file), file, newCategoryFiles(shippedSource, dir, entries)); err != nil {
		return categorySite{}, err
	}
	delete(p.in.unloaded, p.name)
	return categorySite{dir: p.dir, path: join(strings.Join(p.dir, "."), p.name), in: p.in, n: p.in.children[p.name]}, nil
}

// family is the tables beside p that a table family links it to: its parent, and every
// table whose parent it is.
func (p pendingLoad) family(c shippedCategory) []pendingLoad {
	var out []pendingLoad
	if c.parent != "" {
		out = append(out, pendingLoad{dir: p.dir, in: p.in, name: c.parent})
	}
	for name, sibling := range p.in.unloaded {
		if sibling.parent == p.name {
			out = append(out, pendingLoad{dir: p.dir, in: p.in, name: name})
		}
	}
	return out
}
