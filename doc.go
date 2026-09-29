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
//   - draw — one pick from a choice, or one row taken from a table: `pick`,
//     `resolveChoice`, `table.drawRow`.
//   - expansion — one render of one format: `expand`.
//   - render — what owns one `renderDraws`, and so what one reference draw spans:
//     `Generator.Fake`, `FakeRecord`, `Template.Fake` and `RecordTemplate.Fake`
//     start one each, `FakeStruct` one per record, and `expandAnew` one per
//     iteration of a `template.repeat`.
//   - hold — keeping one draw of a name, so every route to it reads that value:
//     an expansion holds a sibling path's head and an operand's field, its draw
//     group a reference path's: `formatOps.held` and `heldCheck` an expansion's,
//     `readReference` a draw group's.
//   - memo — what a hold or a draw group keeps its draws in, the variant drawn at
//     each level and the value each path read: `drawMemo`, `readMemo`.
//   - draw group — reference draws kept apart inside one render:
//     `template.drawGroup` as the data spells it, `templateLink.drawGroupKey` as a
//     render reads it, `renderDraws`, `groupDraws`, `renderScope`.
//   - pin — fixing which row of a table the render uses, which the draw fences replay:
//     `pinSet`, `pinSet.pin`, `table.drawIn`.
//   - family — a table and every table reaching it through a chain of
//     `table.parentT`, named by the root that chain ends at: `table.familyRoot`.
//   - whole — a read of a table with no selector and no descent, {/city}, landing on
//     the `table` itself, so it draws a row apart from every pin: `surveyAt.wholePins`.
//   - fence — a load-time check, so rendering a compiled tree cannot fail: `New`
//     runs them over the data set, `NewTemplate` and `FakeStruct` over what those
//     compile. The draw fences are `heldCheck`, `drawFence` and `checkFamilies`.
