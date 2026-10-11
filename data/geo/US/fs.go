// Package geous is fejkdata's US places module, whose categories are geo.US.*.
package geous

import "embed"

// FS holds the module's categories and its manifest; pass it to fejkdata.WithDataFS.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS
