// Package nbno is fejkdata's nb_NO locale module, whose categories are nb_NO.*; its address and
// company read geo.NO, which package geono holds.
package nbno

import (
	"embed"
	"io/fs"

	geono "github.com/larvit/fejkdata/data/geo/NO"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for geo.NO.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, geono.Modules()...) }
