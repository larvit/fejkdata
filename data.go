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
// label prefixes file names in errors; onDisk marks path as a directory that must
// exist.
type dataSource struct {
	fsys   fs.FS
	label  string
	onDisk bool
	path   string
	dir    string
}

func (s dataSource) labelled(p string) string {
	if s.label == "" {
		return p
	}
	return path.Join(s.label, p)
}

// loadData loads every source into one namespace tree and returns its root
// children. A directory becomes a folder; each *.json file in it compiles to a node
// keyed by its base name (address.json -> "address"); each subdirectory becomes a
// nested folder, so folders turn into dot-path segments. Sources merge left to right:
// matching folders merge by their children, and any other clash is won by the last
// source loaded. Once merged, linkRefs binds every reference against the
// final tree.
func loadData(sources []dataSource) (map[string]node, error) {
	root := map[string]node{}
	for _, src := range sources {
		if src.onDisk {
			if src.path == "" {
				return nil, fmt.Errorf("a data path is empty")
			}
			info, err := os.Stat(src.path)
			if err != nil {
				return nil, fmt.Errorf("data path %s: %w", src.path, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", src.path)
			}
		}
		dir := src.dir
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
	setTablePaths(root)
	if err := linkTables(root); err != nil {
		return nil, err
	}
	if err := linkRefs(root); err != nil {
		return nil, err
	}
	if err := checkNoCycles(root); err != nil {
		return nil, err
	}
	if err := checkScope(treeScope(root)); err != nil {
		return nil, err
	}
	return root, nil
}

// loadDir compiles one directory into a folder. fs.ReadDir yields entries sorted
// by name, so the tree is built deterministically. Empty subdirectories (no JSON
// anywhere under them) are skipped rather than added as empty namespaces.
func loadDir(src dataSource, dir string) (*folder, error) {
	entries, err := fs.ReadDir(src.fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src.labelled(dir), err)
	}
	g := &folder{children: map[string]node{}}
	files := &categoryFiles{src: src, dir: dir, tsv: map[string]bool{}}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tsv") && !e.IsDir() {
			files.tsv[e.Name()] = false
		}
	}
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

// readRows reads the rows file a category names beside it.
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

// nodeScope is the set of nodes one validation pass covers: a whole loaded tree,
// or a single inline node.
type nodeScope func(fn func(path string, n node) error) error

func treeScope(root map[string]node) nodeScope {
	return func(fn func(path string, n node) error) error { return walkNodes(root, fn) }
}

func inlineScope(n node, label string) nodeScope {
	return func(fn func(path string, m node) error) error { return eachNode(n, label, fn) }
}

func checkScope(s nodeScope) error {
	if err := checkColumns(s); err != nil {
		return err
	}
	return checkRenders(s)
}

// checkRenders runs the per-node fences over a scope, each over the whole scope
// before the next, so which of several broken nodes is reported does not depend on
// the walk. Its walks terminate only where nothing renders itself, so run it over a
// loaded tree after checkNoCycles.
func checkRenders(s nodeScope) error {
	mem := renderCounts{}
	if err := s(func(path string, n node) error { return repeatCheck(path, n, mem) }); err != nil {
		return err
	}
	if err := s(heldCheck); err != nil {
		return err
	}
	fence := &drawCheck{}
	refs := false
	if err := s(func(path string, n node) error {
		if t, ok := n.(*template); ok && len(t.refs) > 0 {
			refs = true
		}
		return fence.checkDrawGroup(path, n)
	}); err != nil {
		return err
	}
	if refs {
		if err := s(fence.checkDraws); err != nil {
			return err
		}
		if err := s(fence.checkRecordDraws); err != nil {
			return err
		}
	}
	return s((&valueProof{}).checkDatatype)
}
