// Package misc is fejkdata's locale-neutral module, whose categories are misc.*.
package misc

import "embed"

// FS holds the module's categories and its manifest; pass it to fejkdata.WithDataFS.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS
