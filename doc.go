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

// Vocabulary
//
//   - draw — one item taken from a choice, or one row from a table: `render.drawItem`,
//     `render.drawThroughChoices`, `render.renderEnv.drawRowOf`. Every {…} draws afresh, bar a read of a name.
//   - expansion — one render of one format: `render.expand`.
//   - render — one walk of the tree from an entry point, over one `render.frameStack`:
//     `Generator.Fake`, `FakeRecord`, `Template.Fake` and `RecordTemplate.Fake`
//     start one each, and `FakeStruct` one per record.
//   - env — what a render reads names and rows through: `render.renderEnv`.
//   - spelling — what a token reads, as written: a name, a path into a field or a reference;
//     its `tree.Arm` holds it split into head and tail.
//   - scope — where a name is bound: a category, one repeat iteration or one choice item's render: `tree.NameScope`.
//   - named pick — the draw a token such as {/person as p} binds to a name, drawn
//     on its first read and kept while its scope renders: `tree.NameBinding`,
//     `render.namedPick`, `render.pickFrame`.
//   - level, pick key, addressed key — `tree.PickKey`.
//   - memo — what a named pick keeps its draws in, the variant drawn at each level
//     and the value each path read: `render.drawMemo`.
//   - pin — fixing which row of a table one path, or one named pick, uses:
//     `tree.PinSet`, `render.namedPick.pins`, `render.followStep`.
//   - link — a table's tie to its parent table, each row to its parent row:
//     `load.linkParent`.
//   - family — a table and every table reaching it through a chain of parent
//     links.
//   - resolve — tying a reference or a name to what it names in the assembled tree:
//     `load.CompileTemplates`, `load.resolveRefs`.
//   - head — the string a read starts from: a field or a name as written, or a
//     reference's category path with its sigil, /sv_SE.address: `tree.Arm.Head`. The node it names is
//     the read's start: `tree.Template.StartOf`, `tree.NameTarget.Start`.
//   - fence — a load-time check, run over each category, template or struct as it
//     loads or compiles, so rendering a compiled tree cannot fail.
