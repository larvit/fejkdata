package fejkdata

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/grammar"
)

// indexed is a category an index names and no call has reached yet, and the source holding it.
type indexed struct {
	entry datafiles.IndexEntry
	src   datafiles.Source
}

// placeIndex sets every category the index names as unloaded in root, in place of what
// root holds at its path.
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

func (c categoryAt) unloaded() bool { return c.in.unloaded[c.name] != nil }

// categoryOn is the category, loaded or not, that a path from root names or passes.
func categoryOn(root *folder, segs []string) (categoryAt, bool) {
	g, i := folderAt(root, segs)
	if i == len(segs) {
		return categoryAt{}, false
	}
	c := categoryAt{dir: segs[:i:i], in: g, name: segs[i]}
	return c, c.unloaded() || g.children[c.name] != nil
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
	if c, ok := categoryOn(&f.root, segs); ok && c.unloaded() {
		if err := loadReached(&f.root, nil, []categoryAt{c}); err != nil {
			return "", nil, fmt.Errorf("fejkdata: %w", loadError{err})
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

// needs is every category, loaded or not, that s references, the table beside s that its
// parent column names, and every unloaded table beside s whose parent s is.
func needs(root *folder, s categorySite) []categoryAt {
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

// loadReached loads the unloaded categories among wanted and every unloaded category they
// or sites need, then runs sites and all it loaded through a whole load's pipeline. On an
// error it puts back what it loaded, so the next call fails the same way.
func loadReached(root *folder, sites []categorySite, wanted []categoryAt) error {
	queue := wanted
	for _, s := range sites {
		queue = append(queue, needs(root, s)...)
	}
	var b batch
	err := b.load(root, queue)
	if err == nil {
		err = b.run(root, sites)
	}
	if err != nil {
		b.putBack()
	}
	return err
}

// batch is the categories one loadReached loaded, each beside the entry it was loaded from.
type batch []loaded

type loaded struct {
	site categorySite
	was  *indexed
}

// load loads each unloaded category in queue and what it needs. It stops at a table whose
// parent column disagrees with its entry. Otherwise a table could load without its parent,
// or an earlier table could link to one that putBack then removes.
func (b *batch) load(root *folder, queue []categoryAt) error {
	for ; len(queue) > 0; queue = queue[1:] {
		c := queue[0]
		x := c.in.unloaded[c.name]
		if x == nil {
			continue
		}
		if err := x.src.Load(c.dir, c.name, compileInto(func([]string) *folder { return c.in })); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				err = fmt.Errorf("%s indexes %s: %w", x.src.ManifestPath(), categoryPath(c.dir, c.name), err)
			}
			return err
		}
		s := siteIn(c.dir, c.in, c.name)
		*b = append(*b, loaded{site: s, was: x})
		if parent := parentOf(s.n); parent != x.entry.Parent {
			return fmt.Errorf("%s: the index entry for %s %s, and its table %s", x.src.ManifestPath(), s.path, namesParent(x.entry.Parent), namesParent(parent))
		}
		queue = append(queue, needs(root, s)...)
	}
	return nil
}

func (b batch) run(root *folder, sites []categorySite) error {
	sites = slices.Clip(sites)
	for _, l := range b {
		sites = append(sites, l.site)
	}
	if len(sites) == 0 {
		return nil
	}
	return categoryPipeline(sites, root.children).run()
}

func (b batch) putBack() {
	for _, l := range b {
		l.site.in.putUnloaded(l.site.name, l.was)
	}
}

func namesParent(parent string) string {
	if parent == "" {
		return "names no parent"
	}
	return fmt.Sprintf("names parent %q", parent)
}

func parentOf(n node) string {
	if t, isTable := n.(*table); isTable {
		return t.rows.Options().Parent
	}
	return ""
}
