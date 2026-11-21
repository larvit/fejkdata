// Package shipped generates the files the repository keeps about the shipped data: each
// module's manifest, the listing of every shipped path as JSON and as Markdown, and the pins of
// the shipped shape and of the listing's own shape. A function here returns a file's bytes and
// writes none: the generate command writes them, and the root's tests compare them with the
// files.
package shipped

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// File is a generated file: its slash-separated name from the repository's root, and its body.
type File struct {
	Name string
	Body []byte
}

// Data is the shipped modules loaded whole: one tree, and the namespace of each module in it.
// Every category of the tree lies under one of the namespaces, since each module mounts under
// its own.
type Data struct {
	root       tree.Folder
	namespaces []string
}

// Whole returns a module's files under a manifest naming only its namespace, so engine.Load loads
// it whole, and returns that namespace too.
func Whole(module fs.FS) (fs.FS, string, error) {
	manifest, err := datafiles.FS(module).Manifest()
	if err != nil {
		return nil, "", err
	}
	bare, err := json.Marshal(struct {
		Namespace string `json:"namespace"`
	}{manifest.Namespace})
	if err != nil {
		return nil, "", err
	}
	out := fstest.MapFS{datafiles.ManifestFile: {Data: bare}}
	err = fs.WalkDir(module, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || p == datafiles.ManifestFile {
			return err
		}
		b, err := fs.ReadFile(module, p)
		out[p] = &fstest.MapFile{Data: b}
		return err
	})
	return out, manifest.Namespace, err
}

// Load loads every category of modules whole. A module whose manifest names no namespace is
// refused: its categories would lie under no module.
func Load(modules []fs.FS) (*Data, error) {
	d := &Data{}
	var sources []datafiles.Source
	for i, m := range modules {
		whole, namespace, err := Whole(m)
		if err != nil {
			return nil, err
		}
		if namespace == "" {
			return nil, fmt.Errorf("module %d names no namespace in its %s, so no shipped module holds its categories", i, datafiles.ManifestFile)
		}
		sources = append(sources, datafiles.FS(whole))
		d.namespaces = append(d.namespaces, namespace)
	}
	if err := d.refuseNested(); err != nil {
		return nil, err
	}
	root, err := engine.Load(sources)
	d.root = root
	return d, err
}

// refuseNested refuses a module mounted inside another, since a category under the inner
// namespace would lie under both.
func (d *Data) refuseNested() error {
	for _, outer := range d.namespaces {
		for _, inner := range d.namespaces {
			if strings.HasPrefix(inner, outer+".") {
				return fmt.Errorf("the module %q is mounted inside the module %q, so a category under it has two holders", inner, outer)
			}
		}
	}
	return nil
}

// holder is the index in d.namespaces of the module holding the category or path.
func (d *Data) holder(path string) int {
	i := slices.IndexFunc(d.namespaces, func(n string) bool { return strings.HasPrefix(path, n+".") })
	if i < 0 {
		panic(invariant.Broken("no shipped module holds %s", path))
	}
	return i
}
