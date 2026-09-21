// Package fejkdata generates fake data from recursive JSON templates.
//
// Data lives in JSON, not in Go. A generator starts from the shipped data set and
// layers any directories you add; folders and files become a dot-path namespace,
// then generate values by path:
//
//	f, _ := fejkdata.New(fejkdata.WithSeed(42))
//	f.Fake("sv_SE.address")          // "Järvedsvägen 43\n891 77 Järved"
//	f.Fake("sv_SE.address.locality") // "Sundbyberg": a second call draws afresh
//
// Several sources merge in order, the last winning a name clash, so custom data
// layers over the built-ins. The JSON template format is documented in the README.
//
// # Vocabulary
//
// The words the rest of the package uses, and the unit that owns each:
//
//   - draw — one pick from a choice, or one row taken from a table, and the value
//     the pick produced: `pick` and `drawn` settle a choice, `table.draw` a row, and
//     `draw` carries what a read produced.
//   - expansion — one render of one format, by `expand`. A nested template is an
//     expansion of its own.
//   - render — what owns one `drawSet`, and so what one reference draw spans:
//     `Generator.Fake`, `FakeRecord`, `FakeStruct`, `Template.Fake` and
//     `RecordTemplate.Fake` each start one, as does `expandAnew` per iteration of a
//     `template.repeat` — an iteration is an expansion and a render of its own.
//   - hold — keeping one draw of a name, so every route to it reads that value.
//     `template.held` lists what an expansion holds, `draws` keeps them and
//     `readField` reads through it; a reference path is held for the whole render.
//   - draw group — `template.drawGroup`, which gives the reference paths under it
//     draws of their own inside the render. `drawSet` holds the unnamed group's and
//     each named one's, and `drawScope` says which one a template renders in.
//   - pin — fixing which row of a table the render uses. `draws.pin` pins it and the
//     rows of its ancestors, and `draws.rowOf` draws one where none is pinned.
//   - family — a table and every table reaching it through a chain of
//     `table.parentT`, named by the root that chain ends at: `table.family`.
//   - whole — a read that lands on a `row`, so the row renders through its table's
//     format: `tableRead.whole` says a read did, `table.whole` is the node.
//   - fence — a load-time check refusing data that two routes to one draw would
//     disagree in, which is why a render cannot fail: `heldCheck` over an
//     expansion's held names, `drawCheck` over a render's reference paths, and
//     `checkFamilies` over the rows they pin. `New` runs them over the data set,
//     `NewTemplate` and `FakeStruct` over what those compile.
package fejkdata
