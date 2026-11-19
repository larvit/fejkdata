// Package etee is fejkdata's et_EE locale module, whose categories are et_EE.*.
package etee

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
