// Package nbno is fejkdata's nb_NO locale module, whose categories are nb_NO.*.
package nbno

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
