// Package fejkdata generates fake data from recursive JSON templates.
//
// Data lives in JSON, not in Go. A generator starts from the shipped data set and
// layers any directories you add; folders and files become a dot-path namespace,
// then generate values by path:
//
//	f, _ := fejkdata.New(fejkdata.WithSeed(42))
//	f.Fake("sv_SE.address")          // "Räfsalsvägen 43\n442 42 Kungälv"
//	f.Fake("sv_SE.address.locality") // "Umeå": a second call draws afresh
//
// Several sources merge in order, the last winning a name clash, so custom data
// layers over the built-ins. The JSON template format is documented in the [README].
//
// [README]: https://github.com/larvit/fejkdata#readme
package fejkdata

// Vocabulary
//
//   - draw — one item taken from a choice, or one row from a table: `drawItem`,
//     `drawThroughChoices`, `renderEnv.drawRowOf`. Every {…} draws afresh, bar a read of a name.
//   - expansion — one render of one format: `expand`.
//   - render — one walk of the tree from an entry point, over one `frameStack`:
//     `Generator.Fake`, `FakeRecord`, `Template.Fake` and `RecordTemplate.Fake`
//     start one each, and `FakeStruct` one per record.
//   - env — what a render reads names and rows through: `renderEnv`.
//   - scope — where a name is bound: a category, or one repeat iteration: `nameScope`.
//   - named pick — the draw a token such as {/person as p} binds to a name, drawn
//     on its first read and kept while its scope renders: `nameBinding`,
//     `namedPick`, `pickFrame`.
//   - memo — what a named pick keeps its draws in, the variant drawn at each level
//     and the value each path read: `drawMemo`.
//   - level — `pickKey`.
//   - pick key — `pickKey`.
//   - addressed key — `pickKey`.
//   - pin — fixing which row of a table one path, or one named pick, uses:
//     `pinSet`, `namedPick.pins`, `table.followStep`.
//   - link — a table's tie to its parent table, each row to its parent row:
//     `table.linkParent`.
//   - family — a table and every table reaching it through a chain of parent
//     links.
//   - resolve — tying a reference or a name to what it names in the assembled tree:
//     `resolveTemplates`, `resolveRefs`.
//   - head — the string a read starts from: a field or a name as written, or a
//     reference's category path with its sigil, /sv_SE.address: `arm.head`. The node it names is
//     the read's start: `template.startOf`, `nameTarget.start`.
//   - fence — a load-time check, run over each category, template or struct as it
//     loads or compiles, so rendering a compiled tree cannot fail.
