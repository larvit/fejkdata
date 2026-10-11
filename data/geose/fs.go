// Package geose is fejkdata's Swedish places module, whose categories are geo.SE.*.
package geose

import "embed"

// FS holds the module's categories and its manifest; pass it to fejkdata.WithDataFS.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS
