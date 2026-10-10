// Package en_US is fejkdata's en_US locale module.
package en_US

import "embed"

// FS holds the en_US categories and the module's manifest; pass it to fejkdata.WithDataFS.
//
//go:embed en_US .fejkdata.json
var FS embed.FS
