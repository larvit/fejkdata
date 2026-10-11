// Package enus is fejkdata's en_US locale module, whose categories are en_US.*; its address reads
// geo.US, which package geous holds.
package enus

import (
	"embed"
	"io/fs"

	geous "github.com/larvit/fejkdata/data/geo/US"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.US.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geous.Modules()...) }
