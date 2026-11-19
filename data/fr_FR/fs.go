// Package frfr is fejkdata's fr_FR locale module, whose categories are fr_FR.*.
package frfr

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
