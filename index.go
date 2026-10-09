package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/grammar"
)

// indexed is a category an index names and no call has reached yet, and the source holding it.
type indexed struct {
	src   datafiles.Source
	entry datafiles.IndexEntry
}

// placeIndex sets every category index names unloaded in root, in place of what root holds
// at its path.
func placeIndex(root *folder, src datafiles.Source, index map[string]datafiles.IndexEntry) {
	for _, p := range sortedNames(index) {
		segs := strings.Split(p, ".")
		g := madeFolder(root, segs[:len(segs)-1])
		name := segs[len(segs)-1]
		delete(g.children, name)
		if g.unloaded == nil {
			g.unloaded = map[string]indexed{}
		}
		g.unloaded[name] = indexed{src: src, entry: index[p]}
	}
}

// categoryAt is where a category sits: the folder holding it, that folder's path, and its name.
type categoryAt struct {
	dir  []string
	in   *folder
	name string
}

func (c categoryAt) path() string { return categoryPath(c.dir, c.name) }

// categoryUnder is the category, loaded or not, that a path from root names or descends into.
func categoryUnder(root *folder, segs []string) (categoryAt, bool) {
	g, i := folderAt(root, segs)
	if i == len(segs) {
		return categoryAt{}, false
	}
	if _, unloaded := g.unloaded[segs[i]]; !unloaded && g.children[segs[i]] == nil {
		return categoryAt{}, false
	}
	return categoryAt{dir: segs[:i:i], in: g, name: segs[i]}, true
}

// loadCallerPath strips the / a caller's path may start with, splits the path, and loads the
// unloaded category it names or descends into. The caller walks it next, and a walk loads nothing.
func (f *Generator) loadCallerPath(path string) (string, []string, error) {
	path, err := grammar.CallerPath(path)
	if err != nil {
		return "", nil, fmt.Errorf("fejkdata: %w", err)
	}
	segs, err := grammar.SplitPath(path)
	if err != nil {
		return "", nil, fmt.Errorf("fejkdata: %w", err)
	}
	if c, ok := categoryUnder(&f.root, segs); ok {
		if err := loadReached(&f.root, nil, []categoryAt{c}); err != nil {
			return "", nil, fmt.Errorf("fejkdata: %w", err)
		}
	}
	return path, segs, nil
}

// referenced is every category the templates of nodes reference, each reference read from
// the folder dir.
func referenced(root *folder, dir []string, nodes nodeSet) []categoryAt {
	var out []categoryAt
	_ = nodes(func(_ string, n node) error {
		t, isTemplate := n.(*template)
		if !isTemplate {
			return nil
		}
		for _, name := range refTokens(t.tokens) {
			segs, err := refSegments(name, dir)
			if err != nil {
				continue // resolveRefs reports it
			}
			if c, ok := categoryUnder(root, segs); ok {
				out = append(out, c)
			}
		}
		return nil
	})
	return out
}

// linkedTables is the table beside s that its parent column names, and every unloaded
// table beside s whose parent s is.
func linkedTables(s categorySite) []categoryAt {
	var out []categoryAt
	if t, isTable := s.n.(*table); isTable && t.rows.Options().Parent != "" {
		out = append(out, categoryAt{dir: s.dir, in: s.in, name: t.rows.Options().Parent})
	}
	for _, name := range sortedNames(s.in.unloaded) {
		if s.in.unloaded[name].entry.Parent == s.name {
			out = append(out, categoryAt{dir: s.dir, in: s.in, name: name})
		}
	}
	return out
}

// loadReached loads the unloaded categories among wanted, and every unloaded category they
// or sites reference or link to, then runs sites and all it loaded through a whole load's
// pipeline. On an error it unloads again what it loaded, so the next call fails the same way.
func loadReached(root *folder, sites []categorySite, wanted []categoryAt) error {
	queue := wanted
	for _, s := range sites {
		queue = append(queue, reachedFrom(root, s)...)
	}
	var loaded []categorySite
	var was []indexed
	unload := func() {
		for i, s := range loaded {
			delete(s.in.children, s.name)
			s.in.unloaded[s.name] = was[i]
		}
	}
	for ; len(queue) > 0; queue = queue[1:] {
		c := queue[0]
		x, unloaded := c.in.unloaded[c.name]
		if !unloaded {
			continue
		}
		if err := x.src.Load(c.dir, c.name, compileInto(func([]string) *folder { return c.in })); err != nil {
			unload()
			return err
		}
		delete(c.in.unloaded, c.name)
		s := siteIn(c.dir, c.in, c.name)
		loaded, was = append(loaded, s), append(was, x)
		queue = append(queue, reachedFrom(root, s)...)
	}
	sites = append(sites[:len(sites):len(sites)], loaded...)
	if len(sites) == 0 {
		return nil
	}
	if err := categoryPipeline(sites, root.children).run(); err != nil {
		unload()
		return err
	}
	return nil
}

func reachedFrom(root *folder, s categorySite) []categoryAt {
	return append(referenced(root, s.dir, siteNodes([]categorySite{s})), linkedTables(s)...)
}

// reaching is every unloaded category under root whose index entry leads to a path in
// targets: through its reads, its parent or a child table, and through other unloaded
// categories.
func reaching(root *folder, targets map[string]bool) []categoryAt {
	if len(targets) == 0 {
		return nil
	}
	unloaded := map[string]categoryAt{}
	unloadedUnder(root, nil, unloaded)
	from := map[string][]string{} // a path, and the unloaded categories leading to it
	for _, p := range sortedNames(unloaded) {
		c := unloaded[p]
		e := c.in.unloaded[c.name].entry
		for _, r := range e.Reads {
			from[r] = append(from[r], p)
		}
		if e.Parent != "" {
			parent := categoryPath(c.dir, e.Parent)
			from[parent] = append(from[parent], p)
			from[p] = append(from[p], parent)
		}
	}
	var out []categoryAt
	seen := map[string]bool{}
	for queue := sortedNames(targets); len(queue) > 0; queue = queue[1:] {
		for _, p := range from[queue[0]] {
			c, isUnloaded := unloaded[p]
			if !isUnloaded || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, c)
			queue = append(queue, p)
		}
	}
	return out
}

func unloadedUnder(g *folder, dir []string, out map[string]categoryAt) {
	for name := range g.unloaded {
		c := categoryAt{dir: dir, in: g, name: name}
		out[c.path()] = c
	}
	for name, n := range g.children {
		if sub, isFolder := n.(*folder); isFolder {
			unloadedUnder(sub, append(dir[:len(dir):len(dir)], name), out)
		}
	}
}
