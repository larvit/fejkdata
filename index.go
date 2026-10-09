package fejkdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/grammar"
)

// ErrLoad marks the error of a category that fails to load on the first call reaching it,
// which only a source whose manifest carries an index defers to then.
var ErrLoad = errors.New("data fails to load")

// indexed is a category an index names and no call has reached yet, and the source holding it.
type indexed struct {
	entry datafiles.IndexEntry
	src   datafiles.Source
}

// placeIndex sets every category index names unloaded in root, in place of what root holds
// at its path.
func placeIndex(root *folder, src datafiles.Source, index map[string]datafiles.IndexEntry) {
	for _, p := range sortedNames(index) {
		segs := strings.Split(p, ".")
		madeFolder(root, segs[:len(segs)-1]).putUnloaded(segs[len(segs)-1], &indexed{entry: index[p], src: src})
	}
}

// categoryAt is where a category, loaded or not, sits: the folder holding it, that
// folder's path, and its name.
type categoryAt struct {
	dir  []string
	in   *folder
	name string
}

func (c categoryAt) path() string { return categoryPath(c.dir, c.name) }

// categoryOn is the category, loaded or not, that a path from root names or passes.
func categoryOn(root *folder, segs []string) (categoryAt, bool) {
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
// unloaded category it names or passes. The caller walks it next, and a walk loads nothing.
func (f *Generator) loadCallerPath(path string) (string, []string, error) {
	path, err := grammar.CallerPath(path)
	if err != nil {
		return "", nil, fmt.Errorf("fejkdata: %w", err)
	}
	segs, err := grammar.SplitPath(path)
	if err != nil {
		return "", nil, fmt.Errorf("fejkdata: %w", err)
	}
	if c, ok := categoryOn(&f.root, segs); ok {
		if err := loadReached(&f.root, nil, []categoryAt{c}); err != nil {
			return "", nil, fmt.Errorf("fejkdata: %w: %w", ErrLoad, err)
		}
	}
	return path, segs, nil
}

// referenced is every category, loaded or not, that the templates of nodes reference, each
// reference read from the folder dir.
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
			if c, ok := categoryOn(root, segs); ok {
				out = append(out, c)
			}
		}
		return nil
	})
	return out
}

// dependencies is every category, loaded or not, that s references, the table beside s
// that its parent column names, and every unloaded table beside s whose parent s is.
func dependencies(root *folder, s categorySite) []categoryAt {
	out := referenced(root, s.dir, siteNodes([]categorySite{s}))
	if parent := parentOf(s.n); parent != "" {
		out = append(out, categoryAt{dir: s.dir, in: s.in, name: parent})
	}
	for _, name := range sortedNames(s.in.unloaded) {
		if s.in.unloaded[name].entry.Parent == s.name {
			out = append(out, categoryAt{dir: s.dir, in: s.in, name: name})
		}
	}
	return out
}

// dependents is every unloaded category under root whose index entry leads to one of
// paths: through its reads, its parent or a child table, and through other unloaded
// categories.
func dependents(root *folder, paths []string) []categoryAt {
	if len(paths) == 0 {
		return nil
	}
	unloaded := unloadedUnder(root)
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
	queue := slices.Clone(paths)
	sort.Strings(queue)
	for ; len(queue) > 0; queue = queue[1:] {
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

func unloadedUnder(root *folder) map[string]categoryAt {
	out := map[string]categoryAt{}
	var walk func(g *folder, dir []string)
	walk = func(g *folder, dir []string) {
		for name := range g.unloaded {
			c := categoryAt{dir: dir, in: g, name: name}
			out[c.path()] = c
		}
		for name, n := range g.children {
			if sub, isFolder := n.(*folder); isFolder {
				walk(sub, append(dir[:len(dir):len(dir)], name))
			}
		}
	}
	walk(root, nil)
	return out
}

// loadReached loads the unloaded categories among wanted, and every unloaded category they
// or sites depend on, then runs sites and all it loaded through a whole load's pipeline.
// On an error it puts back what it loaded, so the next call fails the same way. Checking
// each parent column against its entry before linking keeps that whole: a table then
// loads with its parent, so no table loaded earlier links to one put back.
func loadReached(root *folder, sites []categorySite, wanted []categoryAt) error {
	queue := wanted
	for _, s := range sites {
		queue = append(queue, dependencies(root, s)...)
	}
	type load struct {
		site categorySite
		was  *indexed
	}
	var loads []load
	putBack := func() {
		for _, l := range loads {
			l.site.in.putUnloaded(l.site.name, l.was)
		}
	}
	for ; len(queue) > 0; queue = queue[1:] {
		c := queue[0]
		x, unloaded := c.in.unloaded[c.name]
		if !unloaded {
			continue
		}
		err := x.src.Load(c.dir, c.name, compileInto(func([]string) *folder { return c.in }))
		if err == nil {
			loads = append(loads, load{site: siteIn(c.dir, c.in, c.name), was: x})
			if parentOf(c.in.children[c.name]) != x.entry.Parent {
				err = staleEntry(root, loads[len(loads)-1].site, x)
			}
		}
		if err != nil {
			putBack()
			return err
		}
		queue = append(queue, dependencies(root, loads[len(loads)-1].site)...)
	}
	for _, l := range loads {
		sites = append(sites[:len(sites):len(sites)], l.site)
	}
	if len(sites) == 0 {
		return nil
	}
	err := categoryPipeline(sites, root.children).run()
	for _, l := range loads {
		if err == nil && !sameEntry(indexEntry(root, l.site), l.was.entry) {
			err = staleEntry(root, l.site, l.was)
		}
	}
	if err != nil {
		putBack()
	}
	return err
}

// indexEntry is what an index says of the loaded category s.
func indexEntry(root *folder, s categorySite) datafiles.IndexEntry {
	e := datafiles.IndexEntry{Parent: parentOf(s.n), Paths: paths(s.n, false)}
	sort.Strings(e.Paths)
	seen := map[string]bool{s.path: true}
	for _, c := range referenced(root, s.dir, siteNodes([]categorySite{s})) {
		if p := c.path(); !seen[p] {
			seen[p] = true
			e.Reads = append(e.Reads, p)
		}
	}
	sort.Strings(e.Reads)
	return e
}

func sameEntry(a, b datafiles.IndexEntry) bool {
	return a.Parent == b.Parent && slices.Equal(a.Paths, b.Paths) && slices.Equal(a.Reads, b.Reads)
}

// staleEntry names the entry x holds for s, and the entry s calls for. Before linking,
// the paths and reads it names may still lack what s's family adds.
func staleEntry(root *folder, s categorySite, x *indexed) error {
	want, _ := json.Marshal(indexEntry(root, s))
	return fmt.Errorf("%s: the index entry for %s is stale; the category calls for %s", x.src.ManifestPath(), s.path, want)
}

func parentOf(n node) string {
	if t, isTable := n.(*table); isTable {
		return t.rows.Options().Parent
	}
	return ""
}
