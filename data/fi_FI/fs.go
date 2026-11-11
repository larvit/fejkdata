// Package fifi is fejkdata's fi_FI locale module, whose categories are fi_FI.*.
package fifi

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
