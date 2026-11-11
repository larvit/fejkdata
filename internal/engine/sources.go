package engine

import (
	"errors"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// ErrLoad marks data that fails to load: in Load, or on the first call reaching a category
// of an indexed source.
var ErrLoad = errors.New("data fails to load")

// loadError is a load's error, unchanged in its text, matching ErrLoad.
type loadError struct{ error }

func (e loadError) Unwrap() error { return e.error }

func (loadError) Is(target error) bool { return target == ErrLoad }

// Load merges sources in order, the last winning a clash. A source whose manifest
// carries an index leaves its categories unloaded; every other source loads here, with
// the unloaded categories it needs. Its error matches ErrLoad.
// docs/decisions.md#a-source-whose-manifest-carries-an-index-loads-each-category-on-first-reach-any-other-loads-in-new-with-what-it-reads
func Load(sources []datafiles.Source) (tree.Folder, error) {
	root := tree.NewFolder()
	l := tree.NewLoadFinds()
	for _, src := range sources {
		m, err := src.Manifest()
		if err != nil {
			return tree.Folder{}, loadError{err}
		}
		if err := addSource(root, src, m, l); err != nil {
			return tree.Folder{}, loadError{err}
		}
	}
	if err := loadReached(root, categorySites(root), l, nil); err != nil {
		return *root, loadError{err}
	}
	return *root, nil
}

func addSource(root *tree.Folder, src datafiles.Source, m datafiles.Manifest, l *tree.LoadFinds) error {
	src = src.Mounted(m.Namespace)
	if m.Index != nil {
		placeIndex(root, src, m.Index)
		return nil
	}
	g := tree.NewFolder()
	if err := src.Walk(compileInto(func(dir []string) *tree.Folder { return madeFolder(g, dir) }, l)); err != nil {
		return err
	}
	mergeFolder(root, g)
	return nil
}

func compileInto(place func(dir []string) *tree.Folder, l *tree.LoadFinds) func(datafiles.Category) error {
	return func(c datafiles.Category) error {
		n, err := load.CompileCategory(c, l)
		if err != nil {
			return err
		}
		place(c.Folders).Put(c.Name, n)
		return nil
	}
}

func madeFolder(root *tree.Folder, dir []string) *tree.Folder {
	g := root
	for _, seg := range dir {
		sub, isFolder := g.Children()[seg].(*tree.Folder)
		if !isFolder {
			sub = tree.NewFolder()
			g.Put(seg, sub)
		}
		g = sub
	}
	return g
}

func mergeFolder(dst, src *tree.Folder) {
	for k, v := range src.Children() {
		if dg, ok := dst.Children()[k].(*tree.Folder); ok {
			if sg, ok := v.(*tree.Folder); ok {
				mergeFolder(dg, sg)
				continue
			}
		}
		dst.Put(k, v)
	}
}

// categorySite is a loaded category, where it sits and its own dot path.
type categorySite struct {
	categoryAt
	n    tree.Node
	path string
}

func categorySites(root *tree.Folder) []categorySite {
	var out []categorySite
	var walk func(dir []string, g *tree.Folder)
	walk = func(dir []string, g *tree.Folder) {
		for _, name := range tree.SortedNames(g.Children()) {
			if sub, isFolder := g.Children()[name].(*tree.Folder); isFolder {
				walk(append(dir[:len(dir):len(dir)], name), sub)
				continue
			}
			out = append(out, siteIn(dir, g, name))
		}
	}
	walk(nil, root)
	return out
}

func siteIn(dir []string, in *tree.Folder, name string) categorySite {
	return categorySite{categoryAt: categoryAt{dir: dir, in: in, name: name}, n: in.Children()[name], path: load.CategoryPath(dir, name)}
}
