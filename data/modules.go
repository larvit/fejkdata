// Package data lists every module fejkdata ships. A module's tree spells its namespace:
// data/geo/SE embeds geo/SE/…, whose categories are geo.SE.*.
package data

import (
	"io/fs"

	"github.com/larvit/fejkdata/data/en_US"
	"github.com/larvit/fejkdata/data/geo/SE"
	"github.com/larvit/fejkdata/data/geo/US"
	"github.com/larvit/fejkdata/data/misc"
	"github.com/larvit/fejkdata/data/sv_SE"
)

//go:generate env REGENERATE=1 go test -run ^TestShippedManifestsAreCurrent$ ..

// Modules returns every module fejkdata ships, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS {
	return []fs.FS{en_US.FS, SE.FS, US.FS, misc.FS, sv_SE.FS}
}
