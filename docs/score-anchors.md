# Score anchors

Shipped items that anchor each level of L, E and I in `todo.md` (`~/.claude/rules/todo-items.md`). A level no shipped item shows has no row. This repository merges fast-forward, so `score-outcomes` finds no shipped item; each L anchor is chosen by its pull request's lines changed and the components it touched.

| Level | Items |
|---|---|
| L 1 | 119, 151 |
| L 2 | 71, 148 |
| L 3 | 2, 58, 97 |
| L 4 | 4, 102 |
| E 1 | 71, 151 |
| E 2 | 119, 148 |
| E 3 | 58 |
| I 1 | 119, 153 |
| I 2 | 2, 71 |
| I 3 | 4, 58 |

Each anchor's pull request, and why it sits at its level:

- L 1, item 119 (#175, +24/−8 in 4 files): a comment stating how a name's reads stay on one pick, and one test.
- L 1, item 151 (#194, +37/−24 in 5 files): one method that loads the shipped category a caller's path reaches.
- L 2, item 71 (#160, +254/−90 in 16 files): a name as a `calc` operand, inside the calc component.
- L 2, item 148 (#193, +125/−36 in 7 files): the parent-has-a-child fence as one shared step of the geo imports.
- L 3, item 2 (#158, +351/−96 in 17 files): the `..` step, new syntax across the parser, paths and render.
- L 3, item 58 (#220, +588/−37 in 15 files): a new tool, GoReleaser, and live publishing config.
- L 3, item 97 (#164, +280/−118 in 24 files): a new package, `internal/drawstate`.
- L 4, item 4 (#162, +1339/−3686 in 47 files): a rewrite of how every `{…}` draws, across the engine.
- L 4, item 102 (#172, +1078/−790 in 32 files): a new package, `internal/rows`, and a rewrite of every path's walk.
- E 1, items 71 and 151: Go the merge gate tests, so a defect fails a test.
- E 2, item 119: a comment's claim, which only a reader checks.
- E 2, item 148: Python under `data-import/`, which no gate test runs; only regenerating the TSVs and diffing them checks it.
- E 3, item 58: the Packages workflow runs only on a release, so a defect in it meets an installer first; until its first run after merge, the README's install commands failed.
- I 1, items 119 and 153: a comment, and a record of the panel's scores; a revert fixes either in minutes.
- I 2, items 2 and 71: one template feature each, `..` and a named calc operand.
- I 3, item 4: every render's draws, for every consumer.
- I 3, item 58: every install channel at once.
