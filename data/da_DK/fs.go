// Package dadk is fejkdata's da_DK locale module, whose categories are da_DK.*.
package dadk

import (
	"embed"
	"io/fs"
)

// FS holds the locale, its categories and its manifest.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return []fs.FS{FS} }
