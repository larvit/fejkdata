// Package datafiles walks a data tree and reads its category files and their rows.
package datafiles

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

// Source is one tree to load, from the root of an fs.FS. label prefixes file names in
// errors; onDisk marks label as a directory that must exist. mount is the folders the
// tree's root sits under, in the namespace its categories are reached by.
type Source struct {
	fsys   fs.FS
	label  string
	onDisk bool
	mount  []string
}

// Dir is the tree in the directory path, its errors labelled with it.
func Dir(path string) Source {
	return Source{fsys: os.DirFS(path), label: path, onDisk: true}
}

// FS is the tree in fsys.
func FS(fsys fs.FS) Source { return Source{fsys: fsys} }

// Mounted is s with its root under the dot path namespace, "" for the namespace's root.
func (s Source) Mounted(namespace string) Source {
	if namespace != "" {
		s.mount = strings.Split(namespace, ".")
	}
	return s
}

// Category is one category file: the folders above it in the namespace, its name, and its
// parsed JSON.
type Category struct {
	Folders []string
	Name    string
	JSON    jsonvalue.Value
	rows    *rowsFiles
}

// ReadRows reads a rows file beside the category.
func (c Category) ReadRows(name string) (string, error) { return c.rows.read(name) }

// labelled is the file p of fsys as errors name it: under the source's label, or, for a
// mounted source with none, beside its namespace, "sv_SE (address.json)".
func (s Source) labelled(p string) string {
	switch {
	case s.label != "":
		return path.Join(s.label, p)
	case s.mount != nil:
		return fmt.Sprintf("%s (%s)", strings.Join(s.mount, "."), p)
	}
	return p
}

// dirPath is the folder that the namespace folders dir name, as fsys spells it.
func (s Source) dirPath(dir []string) string {
	if len(dir) < len(s.mount) || !slices.Equal(dir[:len(s.mount)], s.mount) {
		panic(invariant.Broken("folder %s lies outside the mount %s", strings.Join(dir, "."), strings.Join(s.mount, ".")))
	}
	if p := path.Join(dir[len(s.mount):]...); p != "" {
		return p
	}
	return "."
}

// Walk hands compile every category of the tree, folders and files in name order. A
// folder holding no category anywhere below it is skipped; a hidden file or folder is no
// category, though a category may name a hidden rows file beside it.
func (s Source) Walk(compile func(Category) error) error {
	if err := s.check(); err != nil {
		return err
	}
	_, err := s.walkDir(slices.Clip(s.mount), compile)
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

// Load hands compile the category file name.json in the namespace folder dir.
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
