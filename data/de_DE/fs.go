// Package dede is fejkdata's de_DE locale module, whose categories are de_DE.*.
package dede

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
