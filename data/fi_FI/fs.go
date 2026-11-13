// Package fifi is fejkdata's fi_FI locale module, whose categories are fi_FI.*; its address reads
// geo.FI, which package geofi holds.
package fifi

import (
	"embed"
	"io/fs"

	geofi "github.com/larvit/fejkdata/data/geo/FI"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.FI.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geofi.Modules()...) }
