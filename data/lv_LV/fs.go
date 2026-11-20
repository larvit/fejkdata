// Package lvlv is fejkdata's lv_LV locale module, whose categories are lv_LV.*; its address reads
// geo.LV, which package geolv holds.
package lvlv

import (
	"embed"
	"io/fs"

	geolv "github.com/larvit/fejkdata/data/geo/LV"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.LV.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geolv.Modules()...) }
