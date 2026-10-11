package datafiles

import (
	"io"
	"io/fs"
	"path"
	"strings"
	"time"
)

// mounted is fsys with its root moved down to the folder at, so a file f of fsys is at/f;
// the folders above at hold only the next folder towards it.
type mounted struct {
	fsys fs.FS
	at   string
}

func (m mounted) Open(name string) (fs.File, error) {
	inner, ok := m.inner(name)
	if ok && inner != "." {
		return m.fsys.Open(inner)
	}
	if ok {
		f, err := m.fsys.Open(".")
		return mountRoot{File: f, name: path.Base(name)}, err
	}
	if next, ok := m.towards(name); ok {
		return &mountDir{name: path.Base(name), next: next}, nil
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (m mounted) ReadDir(name string) ([]fs.DirEntry, error) {
	if inner, ok := m.inner(name); ok {
		return fs.ReadDir(m.fsys, inner)
	}
	if next, ok := m.towards(name); ok {
		return []fs.DirEntry{mountEntry(next)}, nil
	}
	return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
}

// inner is name as fsys spells it, where name lies at or below at.
func (m mounted) inner(name string) (string, bool) {
	if name == m.at {
		return ".", true
	}
	rest, below := strings.CutPrefix(name, m.at+"/")
	return rest, below && fs.ValidPath(name)
}

// towards is the folder below name on the way to at, where name lies above at.
func (m mounted) towards(name string) (string, bool) {
	rest := m.at
	if name != "." {
		var above bool
		if rest, above = strings.CutPrefix(m.at, name+"/"); !above {
			return "", false
		}
	}
	next, _, _ := strings.Cut(rest, "/")
	return next, true
}

// mountRoot is fsys's root, named as the folder it is mounted at.
type mountRoot struct {
	fs.File
	name string
}

func (r mountRoot) Stat() (fs.FileInfo, error) {
	info, err := r.File.Stat()
	return renamed{FileInfo: info, name: r.name}, err
}

func (r mountRoot) ReadDir(n int) ([]fs.DirEntry, error) {
	if d, ok := r.File.(fs.ReadDirFile); ok {
		return d.ReadDir(n)
	}
	return nil, &fs.PathError{Op: "readdir", Path: r.name, Err: fs.ErrInvalid}
}

type renamed struct {
	fs.FileInfo
	name string
}

func (r renamed) Name() string { return r.name }

// mountDir is a folder above the mount, holding only the folder next towards it.
type mountDir struct {
	name, next string
	read       bool
}

func (d *mountDir) Stat() (fs.FileInfo, error) { return mountEntry(d.name), nil }
func (d *mountDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}
func (d *mountDir) Close() error { return nil }

func (d *mountDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.read {
		if n > 0 {
			return nil, io.EOF
		}
		return nil, nil
	}
	d.read = true
	return []fs.DirEntry{mountEntry(d.next)}, nil
}

// mountEntry is a folder above the mount, as both its DirEntry and its FileInfo.
type mountEntry string

func (e mountEntry) Name() string               { return string(e) }
func (e mountEntry) IsDir() bool                { return true }
func (e mountEntry) Type() fs.FileMode          { return fs.ModeDir }
func (e mountEntry) Info() (fs.FileInfo, error) { return e, nil }
func (e mountEntry) Size() int64                { return 0 }
func (e mountEntry) Mode() fs.FileMode          { return fs.ModeDir | 0o555 }
func (e mountEntry) ModTime() time.Time         { return time.Time{} }
func (e mountEntry) Sys() any                   { return nil }
