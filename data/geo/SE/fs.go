// Package SE is fejkdata's Swedish geography module.
package SE

import "embed"

// FS holds the geo.SE categories and the module's manifest; pass it to fejkdata.WithDataFS.
//
//go:embed geo .fejkdata.json
var FS embed.FS
