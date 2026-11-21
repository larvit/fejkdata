// Package data lists every module fejkdata ships.
package data

import (
	"io/fs"

	dadk "github.com/larvit/fejkdata/data/da_DK"
	dede "github.com/larvit/fejkdata/data/de_DE"
	enus "github.com/larvit/fejkdata/data/en_US"
	etee "github.com/larvit/fejkdata/data/et_EE"
	fifi "github.com/larvit/fejkdata/data/fi_FI"
	frfr "github.com/larvit/fejkdata/data/fr_FR"
	geode "github.com/larvit/fejkdata/data/geo/DE"
	geodk "github.com/larvit/fejkdata/data/geo/DK"
	geoee "github.com/larvit/fejkdata/data/geo/EE"
	geofi "github.com/larvit/fejkdata/data/geo/FI"
	geofr "github.com/larvit/fejkdata/data/geo/FR"
	geolt "github.com/larvit/fejkdata/data/geo/LT"
	geolv "github.com/larvit/fejkdata/data/geo/LV"
	geono "github.com/larvit/fejkdata/data/geo/NO"
	geose "github.com/larvit/fejkdata/data/geo/SE"
	geous "github.com/larvit/fejkdata/data/geo/US"
	ltlt "github.com/larvit/fejkdata/data/lt_LT"
	lvlv "github.com/larvit/fejkdata/data/lv_LV"
	"github.com/larvit/fejkdata/data/misc"
	nbno "github.com/larvit/fejkdata/data/nb_NO"
	nnno "github.com/larvit/fejkdata/data/nn_NO"
	svse "github.com/larvit/fejkdata/data/sv_SE"
)

//go:generate go run ../internal/shipped/generate -root ..

// Modules returns every module fejkdata ships, to pass to fejkdata.WithDataFS.
func Modules() []fs.FS {
	return []fs.FS{dadk.FS, dede.FS, enus.FS, etee.FS, fifi.FS, frfr.FS, geode.FS, geodk.FS, geoee.FS, geofi.FS, geofr.FS, geolt.FS, geolv.FS, geono.FS, geose.FS, geous.FS, ltlt.FS, lvlv.FS, misc.FS, nbno.FS, nnno.FS, svse.FS}
}
