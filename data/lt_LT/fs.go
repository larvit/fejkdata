// Package ltlt is fejkdata's lt_LT locale module, whose categories are lt_LT.*.
package ltlt

import (
	"embed"
	"io/fs"
)

// FS holds the locale's categories and its manifest.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS, which reads no other module, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return []fs.FS{FS} }
