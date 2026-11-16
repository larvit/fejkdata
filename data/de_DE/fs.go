// Package dede is fejkdata's de_DE locale module, whose categories are de_DE.*; its address reads
// geo.DE, which package geode holds.
package dede

import (
	"embed"
	"io/fs"

	geode "github.com/larvit/fejkdata/data/geo/DE"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.DE.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geode.Modules()...) }
