// Package frfr is fejkdata's fr_FR locale module, whose categories are fr_FR.*; its address reads
// geo.FR, which package geofr holds.
package frfr

import (
	"embed"
	"io/fs"

	geofr "github.com/larvit/fejkdata/data/geo/FR"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.FR.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geofr.Modules()...) }
