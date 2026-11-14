// Package nnno is fejkdata's nn_NO locale module, whose categories are nn_NO.*; its person,
// fodselsnummer, phone, email and company read nb_NO, which package nbno holds, and its address
// reads geo.NO, which nbno's Modules returns.
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
