// Package svse is fejkdata's sv_SE locale module, whose categories are sv_SE.*; its address reads
// geo.SE, which package geose holds.
package svse

import (
	"embed"
	"io/fs"

	geose "github.com/larvit/fejkdata/data/geo/SE"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.SE.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geose.Modules()...) }
