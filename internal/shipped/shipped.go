// Package shipped generates the files the repository keeps about the shipped data: each
// module's manifest, the listing of every shipped path as JSON and as Markdown, and the pins of
// the shipped shape and of the listing's own shape. A function here returns a file's bytes and
// writes none: the generate command writes them, and the root's tests compare them with the
// files.
package shipped

import (
	"io/fs"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// File is a generated file: its slash-separated name from the repository's root, and its body.
type File struct {
	Name string
	Body []byte
}

// Whole is a module's files under a manifest naming only its namespace, so it loads whole in
// engine.Load, and that namespace.
func Whole(module fs.FS) (fs.FS, string, error) {
	manifest, err := datafiles.FS(module).Manifest()
	if err != nil {
		return nil, "", err
	}
	out := fstest.MapFS{datafiles.ManifestFile: {Data: []byte(`{"namespace": "` + manifest.Namespace + `"}`)}}
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

// Load loads every category of modules whole, and returns the tree with each module's namespace.
func Load(modules []fs.FS) (tree.Folder, []string, error) {
	var sources []datafiles.Source
	var namespaces []string
	for _, m := range modules {
		whole, namespace, err := Whole(m)
		if err != nil {
			return tree.Folder{}, nil, err
		}
		sources = append(sources, datafiles.FS(whole))
		namespaces = append(namespaces, namespace)
	}
	root, err := engine.Load(sources)
	return root, namespaces, err
}

// holder is the index in namespaces of the module holding the category or path, or -1.
func holder(namespaces []string, path string) int {
	return slices.IndexFunc(namespaces, func(n string) bool { return strings.HasPrefix(path, n+".") })
}
