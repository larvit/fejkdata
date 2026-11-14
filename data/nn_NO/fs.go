// Package nnno is fejkdata's nn_NO locale module, whose categories are nn_NO.*; they read
// nb_NO's names, ids, phone numbers, emails and companies, and geo.NO's places.
package nnno

import (
	"embed"
	"io/fs"

	nbno "github.com/larvit/fejkdata/data/nb_NO"
)

// FS holds the locale alone, its categories and its manifest; pass Modules unless your own
// data stands in for nb_NO or geo.NO.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS

// Modules returns FS and every module it reads, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS { return append([]fs.FS{FS}, nbno.Modules()...) }
