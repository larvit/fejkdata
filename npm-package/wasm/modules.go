//go:build js && wasm

package main

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"syscall/js"
	"testing/fstest"

	"github.com/larvit/fejkdata/data"
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

// shipped maps each shipped module's namespace to its Modules.
var shipped = map[string]func() []fs.FS{
	"da_DK": dadk.Modules, "de_DE": dede.Modules, "en_US": enus.Modules, "et_EE": etee.Modules,
	"fi_FI": fifi.Modules, "fr_FR": frfr.Modules, "lt_LT": ltlt.Modules, "lv_LV": lvlv.Modules,
	"nb_NO": nbno.Modules, "nn_NO": nnno.Modules, "sv_SE": svse.Modules, "misc": misc.Modules,
	"geo.DE": geode.Modules, "geo.DK": geodk.Modules, "geo.EE": geoee.Modules, "geo.FI": geofi.Modules,
	"geo.FR": geofr.Modules, "geo.LT": geolt.Modules, "geo.LV": geolv.Modules, "geo.NO": geono.Modules,
	"geo.SE": geose.Modules, "geo.US": geous.Modules,
}

const leftOut = "; load's data leaves out the module this reads"

func newBridge(names, files js.Value) (bridge, error) {
	layered, err := filesFS(files)
	if err != nil {
		return bridge{}, err
	}
	if names.IsNull() {
		return bridge{modules: append(data.Modules(), layered...)}, nil
	}
	chosen, err := chosenModules(names)
	if err != nil {
		return bridge{}, err
	}
	return bridge{modules: append(chosen, layered...), every: append(data.Modules(), layered...)}, nil
}

// chosenModules is the modules the array names names, each with the modules it reads, once.
func chosenModules(names js.Value) ([]fs.FS, error) {
	var chosen []fs.FS
	var unknown []string
	for i := range names.Length() {
		modules, ok := shipped[names.Index(i).String()]
		if !ok {
			unknown = append(unknown, fmt.Sprintf("%q", names.Index(i).String()))
			continue
		}
		for _, m := range modules() {
			if !slices.Contains(chosen, m) {
				chosen = append(chosen, m)
			}
		}
	}
	if unknown != nil {
		return nil, fmt.Errorf("fejkdata: load: data names %s, which no shipped module is; the shipped modules are %s", strings.Join(unknown, ", "), shippedNames())
	}
	return chosen, nil
}

func shippedNames() string {
	names := make([]string, 0, len(shipped))
	for name := range shipped {
		names = append(names, name)
	}
	slices.Sort(names)
	return strings.Join(names, ", ")
}

// filesFS is the array of [path, bytes] pairs files as one module, or none where it is empty.
func filesFS(files js.Value) ([]fs.FS, error) {
	if files.Length() == 0 {
		return nil, nil
	}
	module := fstest.MapFS{}
	for i := range files.Length() {
		path, bytes := files.Index(i).Index(0).String(), files.Index(i).Index(1)
		if !fs.ValidPath(path) || path == "." {
			return nil, fmt.Errorf("fejkdata: load: files holds %q, which is no path; write one such as \"mydata/users.json\", with no leading / and no ..", path)
		}
		body := make([]byte, bytes.Length())
		js.CopyBytesToGo(body, bytes)
		module[path] = &fstest.MapFile{Data: body}
	}
	return []fs.FS{module}, nil
}
