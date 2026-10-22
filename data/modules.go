// Package data lists every module fejkdata ships.
package data

import (
	"io/fs"

	enus "github.com/larvit/fejkdata/data/en_US"
	geono "github.com/larvit/fejkdata/data/geo/NO"
	geose "github.com/larvit/fejkdata/data/geo/SE"
	geous "github.com/larvit/fejkdata/data/geo/US"
	"github.com/larvit/fejkdata/data/misc"
	nbno "github.com/larvit/fejkdata/data/nb_NO"
	svse "github.com/larvit/fejkdata/data/sv_SE"
)

//go:generate env REGENERATE=1 go test -run ^TestShippedManifestsAreCurrent$ ..

// Modules returns every module fejkdata ships, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS {
	return []fs.FS{enus.FS, geono.FS, geose.FS, geous.FS, misc.FS, nbno.FS, svse.FS}
}
