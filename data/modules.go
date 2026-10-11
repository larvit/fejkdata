// Package data lists every module fejkdata ships.
package data

import (
	"io/fs"

	"github.com/larvit/fejkdata/data/enus"
	"github.com/larvit/fejkdata/data/geose"
	"github.com/larvit/fejkdata/data/geous"
	"github.com/larvit/fejkdata/data/misc"
	"github.com/larvit/fejkdata/data/svse"
)

//go:generate env REGENERATE=1 go test -run ^TestShippedManifestsAreCurrent$ ..

// Modules returns every module fejkdata ships, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS {
	return []fs.FS{enus.FS, geose.FS, geous.FS, misc.FS, svse.FS}
}
