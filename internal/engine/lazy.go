package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/fence"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// placeIndex sets every category the index names as unloaded in root, in place of what
// root holds at its path.
func placeIndex(root *tree.Folder, src datafiles.Source, index map[string]datafiles.IndexEntry) {
	for _, p := range tree.SortedNames(index) {
		segs := strings.Split(p, ".")
		madeFolder(root, segs[:len(segs)-1]).PutUnloaded(segs[len(segs)-1], tree.NewIndexed(index[p], src))
	}
}

// categoryAt is where a category, loaded or not, sits: the folder holding it, that
// folder's path, and its name.
type categoryAt struct {
	dir  []string
	in   *tree.Folder
	name string
}

func (c categoryAt) unloaded() bool { return c.in.Unloaded()[c.name] != nil }

// categoryOn is the category, loaded or not, that a path from root names or passes.
func categoryOn(root *tree.Folder, segs []string) (categoryAt, bool) {
	g, i := load.FolderAt(root, segs)
	if i == len(segs) {
		return categoryAt{}, false
	}
	c := categoryAt{dir: segs[:i:i], in: g, name: segs[i]}
	return c, c.unloaded() || g.Children()[c.name] != nil
}

// callerPath strips the / a caller's path may start with, splits the path, and loads the
// unloaded category it names or passes. The caller walks it next, and a walk loads nothing.
func callerPath(root *tree.Folder, path string) (string, []string, error) {
	path, err := grammar.CallerPath(path)
	if err != nil {
		return "", nil, err
	}
	segs, err := grammar.SplitPath(path)
	if err != nil {
		return "", nil, err
	}
	if c, ok := categoryOn(root, segs); ok && c.unloaded() {
		if err := loadReached(root, nil, tree.NewLoadFinds(), []categoryAt{c}); err != nil {
			return "", nil, loadError{err}
		}
	}
	return path, segs, nil
}

// referenced is every category, loaded or not, that the templates of nodes reference, each
// reference read from the folder dir.
func referenced(root *tree.Folder, dir []string, nodes fence.NodeSet) []categoryAt {
	var out []categoryAt
	_ = nodes(func(_ string, n tree.Node) error {
		t, isTemplate := n.(*tree.Template)
		if !isTemplate {
			return nil
		}
		for _, spelling := range load.RefSpellings(t.Tokens()) {
			segs, err := load.RefSegments(spelling, dir)
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
func needs(root *tree.Folder, s categorySite) []categoryAt {
	out := referenced(root, s.dir, siteNodes([]categorySite{s}))
	if parent := parentOf(s.n); parent != "" {
		out = append(out, categoryAt{dir: s.dir, in: s.in, name: parent})
	}
	for _, name := range tree.SortedNames(s.in.Unloaded()) {
		if s.in.Unloaded()[name].Entry().Parent == s.name {
			out = append(out, categoryAt{dir: s.dir, in: s.in, name: name})
		}
	}
	return out
}

// loadReached loads the unloaded categories among wanted, and every unloaded category they
// or sites need, compiling each into l, the tree.LoadFinds sites were compiled into. It then runs
// sites and all it loaded through a whole load's pipeline. On an error it puts back what it
// loaded, so the next call fails the same way.
func loadReached(root *tree.Folder, sites []categorySite, l *tree.LoadFinds, wanted []categoryAt) error {
	queue := wanted
	for _, s := range sites {
		queue = append(queue, needs(root, s)...)
	}
	var b batch
	err := b.load(root, queue, l)
	if err == nil {
		err = b.run(root, sites, l)
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
	was  *tree.Indexed
}

// load loads each unloaded category in queue and what it needs. It stops at a table whose
// parent column disagrees with its entry. Otherwise a table could load without its parent.
func (b *batch) load(root *tree.Folder, queue []categoryAt, l *tree.LoadFinds) error {
	for ; len(queue) > 0; queue = queue[1:] {
		c := queue[0]
		x := c.in.Unloaded()[c.name]
		if x == nil {
			continue
		}
		if err := x.Src().Load(c.dir, c.name, compileInto(func([]string) *tree.Folder { return c.in }, l)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				err = fmt.Errorf("%s indexes %s: %w", x.Src().ManifestPath(), load.CategoryPath(c.dir, c.name), err)
			}
			return err
		}
		s := siteIn(c.dir, c.in, c.name)
		*b = append(*b, loaded{site: s, was: x})
		if parent := parentOf(s.n); parent != x.Entry().Parent {
			return fmt.Errorf("%s: the index entry for %s %s, and its table %s", x.Src().ManifestPath(), s.path, namesParent(x.Entry().Parent), namesParent(parent))
		}
		queue = append(queue, needs(root, s)...)
	}
	return nil
}

func (b batch) run(root *tree.Folder, sites []categorySite, l *tree.LoadFinds) error {
	sites = slices.Clip(sites)
	for _, x := range b {
		sites = append(sites, x.site)
	}
	if len(sites) == 0 {
		return nil
	}
	return categoryPipeline(sites, root.Children(), l).run()
}

func (b batch) putBack() {
	for _, l := range b {
		l.site.in.PutUnloaded(l.site.name, l.was)
	}
}

func namesParent(parent string) string {
	if parent == "" {
		return "names no parent"
	}
	return fmt.Sprintf("names parent %q", parent)
}

func parentOf(n tree.Node) string {
	if t, isTable := n.(*tree.Table); isTable {
		return t.Rows().Options().Parent
	}
	return ""
}
