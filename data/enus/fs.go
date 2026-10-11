// Package enus is fejkdata's en_US locale module, whose categories are en_US.*. Its address reads geous's places.
package enus

import "embed"

// FS holds the module's categories and its manifest; pass it to fejkdata.WithDataFS.
//
//go:embed *.json *.tsv .fejkdata.json
var FS embed.FS
