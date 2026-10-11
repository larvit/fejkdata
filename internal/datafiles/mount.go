package datafiles

import (
	"io/fs"
	"strings"
)

// mounted is fsys with its root moved down to the folder at, so a file f of fsys is at/f. It
// serves what a walk asks of a tree, a folder's entries and a file's bytes: the folders above
// at list only the next folder towards it, and opening one fails.
type mounted struct {
	fsys fs.FS
	at   string
}

func (m mounted) Open(name string) (fs.File, error) {
	if inner, below := m.inner(name); below {
		return m.fsys.Open(inner)
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (m mounted) ReadFile(name string) ([]byte, error) {
	if inner, below := m.inner(name); below {
		return fs.ReadFile(m.fsys, inner)
	}
	return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
}

func (m mounted) ReadDir(name string) ([]fs.DirEntry, error) {
	if inner, below := m.inner(name); below {
		return fs.ReadDir(m.fsys, inner)
	}
	rest, above := strings.CutPrefix(m.at, name+"/")
	if name == "." {
		rest, above = m.at, true
	}
	if !above {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	next, _, _ := strings.Cut(rest, "/")
	return []fs.DirEntry{folderEntry(next)}, nil
}

// inner is name as fsys spells it, where name lies at or below at.
func (m mounted) inner(name string) (string, bool) {
	inner, below := underMount(m.at, name)
	return inner, below && fs.ValidPath(name)
}

// underMount is name as the tree mounted at at spells it, where name lies at or below at.
func underMount(at, name string) (string, bool) {
	switch name {
	case at:
		return ".", true
	case "":
		return name, false
	}
	if at == "" {
		return name, true
	}
	return strings.CutPrefix(name, at+"/")
}

// folderEntry is a folder above the mount, as a folder's entries list it.
type folderEntry string

func (e folderEntry) Name() string      { return string(e) }
func (e folderEntry) IsDir() bool       { return true }
func (e folderEntry) Type() fs.FileMode { return fs.ModeDir }
func (e folderEntry) Info() (fs.FileInfo, error) {
	return nil, &fs.PathError{Op: "stat", Path: string(e), Err: fs.ErrNotExist}
}
