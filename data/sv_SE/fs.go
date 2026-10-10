// Package sv_SE is fejkdata's sv_SE locale module.
package sv_SE

import "embed"

// FS holds the sv_SE categories and the module's manifest; pass it to fejkdata.WithDataFS.
//
//go:embed sv_SE .fejkdata.json
var FS embed.FS
