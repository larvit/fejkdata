// Package misc is fejkdata's misc module.
package misc

import "embed"

// FS holds the misc categories and the module's manifest; pass it to fejkdata.WithDataFS.
//
//go:embed misc .fejkdata.json
var FS embed.FS
