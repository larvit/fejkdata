package datafiles

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
)

// Files is the tree of the files held in memory, each keyed by its slash path from the root.
func Files(files map[string]string) Source { return Source{fsys: heldFiles(files)} }

// heldFiles serves what a Source reads: a whole file, and a folder's entries.
type heldFiles map[string]string

func (f heldFiles) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: errors.ErrUnsupported}
}

func (f heldFiles) ReadFile(name string) ([]byte, error) {
	content, held := f[name]
	if !held {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	return []byte(content), nil
}

// ReadDir lists the files in the folder name and the folders that hold a file below it, by name.
func (f heldFiles) ReadDir(name string) ([]fs.DirEntry, error) {
	prefix := name + "/"
	if name == "." {
		prefix = ""
	}
	isDir := map[string]bool{}
	for path := range f {
		if rest, inside := strings.CutPrefix(path, prefix); inside {
			entry, _, below := strings.Cut(rest, "/")
			isDir[entry] = below
		}
	}
	if len(isDir) == 0 && name != "." {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	entries := make([]fs.DirEntry, 0, len(isDir))
	for entry, dir := range isDir {
		entries = append(entries, heldEntry{entry, dir})
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, nil
}

type heldEntry struct {
	name string
	dir  bool
}

func (e heldEntry) Name() string { return e.name }
func (e heldEntry) IsDir() bool  { return e.dir }

func (e heldEntry) Type() fs.FileMode {
	if e.dir {
		return fs.ModeDir
	}
	return 0
}

func (e heldEntry) Info() (fs.FileInfo, error) { return nil, errors.ErrUnsupported }
