// Package misc is fejkdata's locale-neutral module, whose categories are misc.*.
package misc

import (
	"embed"
	"io/fs"
)

// FS holds the module's categories and its manifest.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS, which reads no other module, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return []fs.FS{FS} }
