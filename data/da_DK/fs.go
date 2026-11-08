// Package dadk is fejkdata's da_DK locale module, whose categories are da_DK.*; its address reads
// geo.DK, which package geodk holds.
package dadk

import (
	"embed"
	"io/fs"

	geodk "github.com/larvit/fejkdata/data/geo/DK"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.DK.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geodk.Modules()...) }
