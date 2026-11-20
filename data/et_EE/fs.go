// Package etee is fejkdata's et_EE locale module, whose categories are et_EE.*; its address reads
// geo.EE, which package geoee holds.
package etee

import (
	"embed"
	"io/fs"

	geoee "github.com/larvit/fejkdata/data/geo/EE"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.EE.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geoee.Modules()...) }
