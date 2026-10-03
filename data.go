package fejkdata

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

// dataSource is one tree to load: an fs.FS and the directory in it to start from.
// label prefixes file names in errors; onDisk marks diskPath as a directory that must
// exist.
type dataSource struct {
	fsys     fs.FS
	label    string
	onDisk   bool
	diskPath string
	baseDir  string
}

func (s dataSource) labelled(p string) string {
	if s.label == "" {
		return p
	}
	return path.Join(s.label, p)
}

func loadData(sources []dataSource) (map[string]node, error) {
	root := map[string]node{}
	for _, src := range sources {
		if src.onDisk {
			if src.diskPath == "" {
				return nil, fmt.Errorf("a data path is empty")
			}
			info, err := os.Stat(src.diskPath)
			if err != nil {
				return nil, fmt.Errorf("data path %s: %w", src.diskPath, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", src.diskPath)
			}
		}
		dir := src.baseDir
		if dir == "" {
			dir = "."
		}
		g, err := loadDir(src, dir)
		if err != nil {
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

// loadDir compiles one directory into a folder. Empty subdirectories (no JSON
// anywhere under them) are skipped rather than added as empty namespaces.
func loadDir(src dataSource, dir string) (*folder, error) {
	entries, err := fs.ReadDir(src.fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src.labelled(dir), err)
	}
	g := &folder{children: map[string]node{}}
	files := newCategoryFiles(src, dir, entries)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") { // hidden: a checkout or an editor's file, never data
			continue
		}
		full := path.Join(dir, e.Name())
		if e.IsDir() {
			if err := loadFolder(src, g, full, e.Name()); err != nil {
				return nil, err
			}
			continue
		}
		if err := loadFile(src, g, full, e.Name(), files); err != nil {
			return nil, err
		}
	}
	for name, named := range files.tsv {
		if !named && !strings.HasPrefix(name, ".") {
			return nil, fmt.Errorf("%s: no category names it in its rows; a table's rows file sits beside a category file naming it", src.labelled(path.Join(dir, name)))
		}
	}
	return g, nil
}

// categoryFiles is what a category may name beside itself: the rows files of its
// directory, each marked once a category names it.
type categoryFiles struct {
	src dataSource
	dir string
	tsv map[string]bool
}

func newCategoryFiles(src dataSource, dir string, entries []fs.DirEntry) *categoryFiles {
	files := &categoryFiles{src: src, dir: dir, tsv: map[string]bool{}}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tsv") && !e.IsDir() {
			files.tsv[e.Name()] = false
		}
	}
	return files
}

func (c *categoryFiles) readRows(name string) (string, error) {
	if _, present := c.tsv[name]; !present {
		return "", fmt.Errorf("rows names %s, which is not beside it in %s", name, c.src.labelled(c.dir))
	}
	c.tsv[name] = true
	b, err := fs.ReadFile(c.src.fsys, path.Join(c.dir, name))
	if err != nil {
		return "", fmt.Errorf("%s: %w", c.src.labelled(path.Join(c.dir, name)), err)
	}
	return string(b), nil
}

// loadFolder adds a subdirectory as a nested folder, unless nothing under it is data.
func loadFolder(src dataSource, g *folder, full, name string) error {
	child, err := loadDir(src, full)
	if err != nil {
		return err
	}
	if len(child.children) == 0 {
		return nil
	}
	if err := checkName(name); err != nil {
		return fmt.Errorf("%s: folder %w", src.labelled(full), err)
	}
	g.children[name] = child
	return nil
}

// loadFile compiles a *.json file into a category named after it; any other file
// is skipped.
func loadFile(src dataSource, g *folder, full, file string, files *categoryFiles) error {
	if !strings.HasSuffix(file, ".json") {
		return nil
	}
	name := strings.TrimSuffix(file, ".json")
	if err := checkName(name); err != nil {
		return fmt.Errorf("%s: category %w", src.labelled(full), err)
	}
	b, err := fs.ReadFile(src.fsys, full)
	if err != nil {
		return fmt.Errorf("%s: %w", src.labelled(full), err)
	}
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("%s: %w", src.labelled(full), err)
	}
	n, err := compileCategory(raw, name, files)
	if err != nil {
		return fmt.Errorf("%s: %w", src.labelled(full), err)
	}
	g.children[name] = n
	return nil
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
	scope      nodeScope
	link       func() error
	scopeFence func() error
	// Set by a caller that proves the columns against their Go types itself.
	typedByGo bool
}

// categoryBinding binds the categories of one load, their references resolving
// against root.
func categoryBinding(sites []categorySite, root map[string]node) binding {
	return binding{
		scope: sitesScope(sites),
		link: func() error {
			setTablePaths(sites)
			if err := linkTables(sites); err != nil {
				return err
			}
			return linkRefs(sites, root)
		},
		scopeFence: func() error { return checkNoCycles(sites) },
	}
}

func (b binding) bind() error {
	if err := b.link(); err != nil {
		return err
	}
	if err := b.scopeFence(); err != nil {
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
// does not depend on the walk. Its walks recurse unguarded, so the scope's cycles are
// refused first: categories by checkNoCycles, and an inline node closes none, since nothing
// references it.
func checkNodeFences(s nodeScope) error {
	mem := renderCounts{}
	if err := s(func(path string, n node) error { return repeatCheck(path, n, mem) }); err != nil {
		return err
	}
	if err := s(heldCheck); err != nil {
		return err
	}
	fence := &drawFence{fold: newReadFold()}
	refs := false
	if err := s(func(path string, n node) error {
		if t, ok := n.(*template); ok && len(t.link.refs) > 0 {
			refs = true
		}
		return fence.checkDrawGroup(path, n)
	}); err != nil {
		return err
	}
	if refs {
		if err := s(checkOwnFamily); err != nil {
			return err
		}
		if err := s(fence.checkDraws); err != nil {
			return err
		}
		if err := s(fence.checkRecordDraws); err != nil {
			return err
		}
	}
	return s((&valueProof{}).checkDatatype)
}
