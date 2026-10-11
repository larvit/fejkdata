// Package svse is fejkdata's sv_SE locale module, whose categories are sv_SE.*. Its address reads geose's places.
package svse

import "embed"

// FS holds the module's categories and its manifest; pass it to fejkdata.WithDataFS.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS
