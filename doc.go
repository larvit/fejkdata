// Package fejkdata generates locale-aware fake data for tests and fixtures from JSON templates.
//
// Data lives in JSON, not in Go, and comes in modules: a Go package exporting an fs.FS,
// or a folder. A generator loads the modules you name, such as data.Modules(), every
// module fejkdata ships, from github.com/larvit/fejkdata/data. Their folders and files
// become a dot-path namespace, and you generate values by path:
//
//	f, _ := fejkdata.New(fejkdata.WithDataFS(data.Modules()...), fejkdata.WithSeed(42), fejkdata.WithToday(time.Now()))
//	f.Fake("sv_SE.address")          // "Triogatan 43\n442 42 Kungälv"
//	f.Fake("sv_SE.address.locality") // "Umeå": a second call draws afresh
//
// Several sources merge in order, the last winning a name clash, so custom data
// layers over the built-ins. The JSON template format is documented in the [README].
//
// [README]: https://github.com/larvit/fejkdata#readme
package fejkdata
