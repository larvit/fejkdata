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
// layers over the built-ins. The JSON template format is documented in the [README].
//
// [README]: https://github.com/larvit/fejkdata#readme
package fejkdata

// Vocabulary
//
//   - draw — one pick from a choice, or one row taken from a table, and the value a
//     read produced: `pick`, `resolveChoice`, `table.drawRow`, `draw`.
//   - expansion — one render of one format: `expand`.
//   - render — what owns one `holdSet`, and so what one reference draw spans:
//     `Generator.Fake`, `FakeRecord`, `Template.Fake` and `RecordTemplate.Fake`
//     start one each, `FakeStruct` one per record, and `expandAnew` one per
//     iteration of a `template.repeat`.
//   - hold — keeping one draw of a name, so every route to it reads that value:
//     `template.held`, `hold`, `readField`.
//   - draw group — reference draws held apart inside one render:
//     `template.drawGroup` as the data spells it, `template.drawGroupKey` as a
//     render reads it, `holdSet`, `renderScope`.
//   - pin — fixing which row of a table the render uses, which the draw fences replay:
//     `pinSet`, `pinSet.pin`, `table.drawIn`.
//   - family — a table and every table reaching it through a chain of
//     `table.parentT`, named by the root that chain ends at: `table.family`.
//   - whole — a read that lands on a `tableRow`, so the row renders through its table's
//     format: `tableRead.landsWhole` says a read did, `table.wholeRow` is the node.
//   - fence — a load-time check, so rendering a compiled tree cannot fail: `New`
//     runs them over the data set, `NewTemplate` and `FakeStruct` over what those
//     compile. The draw fences are `heldCheck`, `drawCheck` and `checkFamilies`.
