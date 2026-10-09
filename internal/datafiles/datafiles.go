// Package datafiles walks a data tree and reads its category files and their rows.
package datafiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// Source is one tree to load: an fs.FS and the directory in it to start from. label
// prefixes file names in errors; onDisk marks label as a directory that must exist.
type Source struct {
	fsys   fs.FS
	label  string
	onDisk bool
	base   string
}

// Dir is the tree in the directory path, its errors labelled with it.
func Dir(path string) Source {
	return Source{fsys: os.DirFS(path), label: path, onDisk: true}
}

// FS is the tree in fsys below base, "" for its root.
func FS(fsys fs.FS, base string) Source { return Source{fsys: fsys, base: base} }

// Category is one category file: the folders above it from the source's base, its name,
// and its parsed JSON.
type Category struct {
	Folders []string
	Name    string
	JSON    jsonvalue.Value
	rows    *rowsFiles
}

// ReadRows reads a rows file beside the category.
func (c Category) ReadRows(name string) (string, error) { return c.rows.read(name) }

func (s Source) labelled(p string) string {
	if s.label == "" {
		return p
	}
	return path.Join(s.label, p)
}

// dirPath is the folder that dir names, as fsys spells it.
func (s Source) dirPath(dir []string) string {
	if p := path.Join(append([]string{s.base}, dir...)...); p != "" {
		return p
	}
	return "."
}

// ManifestFile is the file at a source's root that describes the source.
const ManifestFile = ".fejkdata.json"

// Manifest is what a source's manifest says of it.
type Manifest struct {
	// Index is every category of the source by dot path; nil where the manifest has none.
	Index map[string]IndexEntry `json:"index"`
}

// IndexEntry is what an index says of a category: its parent table, "" for none, the paths
// List advertises below it, and the categories its templates reference.
type IndexEntry struct {
	Parent string   `json:"parent,omitempty"`
	Paths  []string `json:"paths"`
	Reads  []string `json:"reads,omitempty"`
}

// Manifest reads the manifest at the source's root, the zero Manifest where there is none.
func (s Source) Manifest() (Manifest, error) {
	if err := s.check(); err != nil {
		return Manifest{}, err
	}
	b, err := fs.ReadFile(s.fsys, path.Join(s.base, ManifestFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Manifest{}, nil
	}
	if err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", s.ManifestPath(), err)
	}
	m, err := decodeManifest(b)
	if err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", s.ManifestPath(), err)
	}
	return m, nil
}

// ManifestPath is the manifest's path as errors name it.
func (s Source) ManifestPath() string { return s.labelled(path.Join(s.base, ManifestFile)) }

// decodeManifest decodes a manifest, refusing any key or shape it does not define, and
// reports every mistake in the order of the entries naming them.
func decodeManifest(b []byte) (Manifest, error) {
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return Manifest{}, err
	}
	top, isObject := raw.(map[string]any)
	if !isObject {
		return Manifest{}, fmt.Errorf("a manifest is a JSON object")
	}
	var m Manifest
	var errs []error
	for _, k := range sortedKeys(top) {
		if k != "index" {
			errs = append(errs, fmt.Errorf("unknown key %q; a manifest holds index", k))
			continue
		}
		index, isObject := top[k].(map[string]any)
		if !isObject {
			errs = append(errs, fmt.Errorf("index is a JSON object of category paths"))
			continue
		}
		m.Index = map[string]IndexEntry{}
		for _, p := range sortedKeys(index) {
			e, err := decodeEntry(p, index[p])
			if err != nil {
				errs = append(errs, fmt.Errorf("index entry %q: %w", p, err))
			}
			m.Index[p] = e
		}
	}
	for _, p := range sortedKeys(m.Index) {
		if err := checkParent(m.Index, p); err != nil {
			errs = append(errs, fmt.Errorf("index entry %q: %w", p, err))
		}
	}
	if len(errs) > 0 {
		return Manifest{}, errors.Join(errs...)
	}
	return m, nil
}

func decodeEntry(p string, v any) (IndexEntry, error) {
	var e IndexEntry
	errs := []error{checkCategoryPath(p)}
	fields, isObject := v.(map[string]any)
	if !isObject {
		return e, fmt.Errorf("an entry is a JSON object of parent, paths and reads")
	}
	if _, has := fields["paths"]; !has {
		errs = append(errs, fmt.Errorf("paths is missing"))
	}
	var err error
	for _, k := range sortedKeys(fields) {
		switch k {
		case "parent":
			var isString bool
			if e.Parent, isString = fields[k].(string); !isString {
				err = fmt.Errorf("parent is a string")
			} else if err = grammar.CheckIdentifier(e.Parent); err != nil {
				err = fmt.Errorf("parent %w", err)
			}
		case "paths":
			e.Paths, err = stringList(k, fields[k], func(q string) error {
				if q == "" {
					return nil
				}
				return checkCategoryPath(q)
			})
		case "reads":
			e.Reads, err = stringList(k, fields[k], checkCategoryPath)
		default:
			err = fmt.Errorf("unknown key %q; an entry holds parent, paths and reads", k)
		}
		errs = append(errs, err)
	}
	return e, errors.Join(errs...)
}

// stringList is the JSON array of strings v, each passing check.
func stringList(key string, v any, check func(string) error) ([]string, error) {
	items, isArray := v.([]any)
	if !isArray {
		return nil, fmt.Errorf("%s is an array of strings", key)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, isString := item.(string)
		if !isString {
			return nil, fmt.Errorf("%s is an array of strings", key)
		}
		if err := check(s); err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		out = append(out, s)
	}
	return out, nil
}

// checkParent refuses a parent naming no entry beside p: a table loads with its parent,
// from one source.
func checkParent(index map[string]IndexEntry, p string) error {
	parent := index[p].Parent
	if parent == "" {
		return nil
	}
	sibling := parent
	if dot := strings.LastIndexByte(p, '.'); dot >= 0 {
		sibling = p[:dot+1] + parent
	}
	if _, ok := index[sibling]; !ok {
		return fmt.Errorf("parent %q names no entry beside it in this index", parent)
	}
	return nil
}

func checkCategoryPath(p string) error {
	for _, seg := range strings.Split(p, ".") {
		if err := grammar.CheckIdentifier(seg); err != nil {
			return fmt.Errorf("%q: %w", p, err)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Walk hands compile every category of the tree, folders and files in name order. A
// folder holding no category anywhere below it is skipped; a hidden file or folder is no
// category, though a category may name a hidden rows file beside it.
func (s Source) Walk(compile func(Category) error) error {
	if err := s.check(); err != nil {
		return err
	}
	_, err := s.walkDir(nil, compile)
	return err
}

// check refuses a data path that is empty or names no directory.
func (s Source) check() error {
	if !s.onDisk {
		return nil
	}
	if s.label == "" {
		return fmt.Errorf("a data path is empty")
	}
	info, err := os.Stat(s.label)
	if err != nil {
		return fmt.Errorf("data path %s: %w", s.label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", s.label)
	}
	return nil
}

// walkDir walks the folder that dir names, and reports whether it handed over any category.
func (s Source) walkDir(dir []string, compile func(Category) error) (bool, error) {
	full := s.dirPath(dir)
	entries, err := fs.ReadDir(s.fsys, full)
	if err != nil {
		return false, fmt.Errorf("%s: %w", s.labelled(full), err)
	}
	rows := newRowsFiles(s, full, entries)
	handed := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		var holds bool
		if e.IsDir() {
			holds, err = s.walkFolder(append(dir[:len(dir):len(dir)], e.Name()), compile)
		} else {
			holds, err = s.compileFile(dir, e.Name(), rows, compile)
		}
		if err != nil {
			return false, err
		}
		handed = handed || holds
	}
	return handed, nil
}

// walkFolder walks a subfolder, and refuses its name if it holds a category.
func (s Source) walkFolder(dir []string, compile func(Category) error) (bool, error) {
	holds, err := s.walkDir(dir, compile)
	if err != nil || !holds {
		return false, err
	}
	if err := grammar.CheckIdentifier(dir[len(dir)-1]); err != nil {
		return false, fmt.Errorf("%s: folder %w", s.labelled(s.dirPath(dir)), err)
	}
	return true, nil
}

// Load hands compile the category file name.json in the folder dir.
func (s Source) Load(dir []string, name string, compile func(Category) error) error {
	full := s.dirPath(dir)
	entries, err := fs.ReadDir(s.fsys, full)
	if err != nil {
		return fmt.Errorf("%s: %w", s.labelled(full), err)
	}
	_, err = s.compileFile(dir, name+".json", newRowsFiles(s, full, entries), compile)
	return err
}

// compileFile hands compile a *.json file as a category named after it, and reports
// whether it was one; any other file is skipped.
func (s Source) compileFile(dir []string, file string, rows *rowsFiles, compile func(Category) error) (bool, error) {
	if !strings.HasSuffix(file, ".json") {
		return false, nil
	}
	full := path.Join(rows.dir, file)
	name := strings.TrimSuffix(file, ".json")
	if err := grammar.CheckIdentifier(name); err != nil {
		return false, fmt.Errorf("%s: category %w", s.labelled(full), err)
	}
	b, err := fs.ReadFile(s.fsys, full)
	if err != nil {
		return false, fmt.Errorf("%s: %w", s.labelled(full), err)
	}
	raw, err := jsonvalue.Decode(b)
	if err != nil {
		return false, fmt.Errorf("%s: %w", s.labelled(full), err)
	}
	if err := compile(Category{Folders: dir, Name: name, JSON: raw, rows: rows}); err != nil {
		return false, fmt.Errorf("%s: %w", s.labelled(full), err)
	}
	return true, nil
}

// rowsFiles is what a category may name beside itself: the rows files of its folder.
type rowsFiles struct {
	src Source
	dir string
	tsv map[string]bool
}

func newRowsFiles(src Source, dir string, entries []fs.DirEntry) *rowsFiles {
	files := &rowsFiles{src: src, dir: dir, tsv: map[string]bool{}}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tsv") && !e.IsDir() {
			files.tsv[e.Name()] = true
		}
	}
	return files
}

func (r *rowsFiles) read(name string) (string, error) {
	if !r.tsv[name] {
		return "", fmt.Errorf("rows names %s, which is not beside it in %s", name, r.src.labelled(r.dir))
	}
	b, err := fs.ReadFile(r.src.fsys, path.Join(r.dir, name))
	if err != nil {
		return "", fmt.Errorf("%s: %w", r.src.labelled(path.Join(r.dir, name)), err)
	}
	return string(b), nil
}
