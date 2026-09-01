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
// label prefixes file names in errors; path, when set, is a directory on disk that
// must exist.
type dataSource struct {
	fsys  fs.FS
	label string
	path  string
	root  string
}

func (s dataSource) name(p string) string {
	if s.label == "" {
		return p
	}
	return path.Join(s.label, p)
}

// loadData loads every source into one namespace tree and returns its root
// children. A directory becomes a group; each *.json file in it compiles to a node
// keyed by its base name (address.json -> "address"); each subdirectory becomes a
// nested group, so folders turn into dot-path segments. Sources merge left to right:
// matching groups merge by their children, and any other clash is won by the last
// source loaded. Once merged, linkRefs binds every {..path} reference against the
// final tree.
func loadData(sources []dataSource) (map[string]node, error) {
	root := map[string]node{}
	for _, src := range sources {
		if src.path != "" {
			info, err := os.Stat(src.path)
			if err != nil {
				return nil, fmt.Errorf("%s: no such directory", src.path)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", src.path)
			}
		}
		dir := src.root
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
	if err := linkRefs(root); err != nil {
		return nil, err
	}
	if err := checkNoCycles(root); err != nil {
		return nil, err
	}
	if err := checkBoundLevelsHeld(root); err != nil {
		return nil, err
	}
	return root, nil
}

// loadDir compiles one directory into a group. fs.ReadDir yields entries sorted
// by name, so the tree is built deterministically. Empty subdirectories (no JSON
// anywhere under them) are skipped rather than added as empty namespaces.
func loadDir(src dataSource, dir string) (*group, error) {
	entries, err := fs.ReadDir(src.fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src.name(dir), err)
	}
	g := &group{children: map[string]node{}}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") { // hidden: a checkout or an editor's file, never data
			continue
		}
		full := path.Join(dir, e.Name())
		if e.IsDir() {
			child, err := loadDir(src, full)
			if err != nil {
				return nil, err
			}
			if len(child.children) == 0 {
				continue
			}
			if err := checkName(e.Name()); err != nil {
				return nil, fmt.Errorf("%s: folder %w", src.name(full), err)
			}
			g.children[e.Name()] = child
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		if err := checkName(name); err != nil {
			return nil, fmt.Errorf("%s: category %w", src.name(full), err)
		}
		b, err := fs.ReadFile(src.fsys, full)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.name(full), err)
		}
		var raw any
		if err := json.Unmarshal(b, &raw); err != nil {
			return nil, fmt.Errorf("%s: %w", src.name(full), err)
		}
		n, err := compile(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.name(full), err)
		}
		g.children[name] = n
	}
	return g, nil
}

// mergeChildren overlays src onto dst. Two groups under the same key merge
// recursively (so locales/categories from several paths combine); every other
// key is replaced, making the last-loaded directory win on a conflict.
func mergeChildren(dst, src map[string]node) {
	for k, v := range src {
		if dg, ok := dst[k].(*group); ok {
			if sg, ok := v.(*group); ok {
				mergeChildren(dg.children, sg.children)
				continue
			}
		}
		dst[k] = v
	}
}
