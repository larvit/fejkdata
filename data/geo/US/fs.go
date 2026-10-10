// Package US is fejkdata's US geography module.
package US

import "embed"

// FS holds the geo.US categories and the module's manifest; pass it to fejkdata.WithDataFS.
//
//go:embed geo .fejkdata.json
var FS embed.FS
