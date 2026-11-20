// Package ltlt is fejkdata's lt_LT locale module, whose categories are lt_LT.*; its address reads
// geo.LT, which package geolt holds.
package ltlt

import (
	"embed"
	"io/fs"

	geolt "github.com/larvit/fejkdata/data/geo/LT"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.LT.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geolt.Modules()...) }
