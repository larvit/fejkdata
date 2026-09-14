# Release checklist

## Before the first release

The record API lands first, so the data update can use it.

### Record API

- Rename the record entry points after the string ones: `Record(path)` becomes
  `FakeRecord(path)`, and `FakeRecord(inline)` becomes `FakeRecordTemplate(inline)`.
- Typed columns — a column declares its type, so `json` writes `42` rather than
  `"42"` and `sql` an unquoted literal: string, integer, number, boolean, and a
  way to write null. A template that can render a value its type rejects is a
  load error. The option key is reserved from then on, so a common column name
  like `type` is a poor pick.
- Struct-filling — fill a Go struct from `fake:"…"` tags holding a path or an
  inline template, for parity with gofakeit and go-faker. The field's Go type is
  the column type, through the same conversion and load checks as typed columns,
  and a nested struct is its own draw group. Revise the Decision "A record is a
  template seen as columns, not a second schema format" with that reason.
- Draw groups — references into one category share one draw per render (one
  record, or one `Fake`) in both views; each `repeat` iteration draws anew, and a
  bare reference draws each time. An option naming a draw group splits a render
  into several entities. Replaces the Decision "A record shares one reference draw
  per category". Each expectation becomes a test:
  - `first`, `last` and `email` reading `person` → one person
  - `from_first`/`from_last` grouped `from`, `to_first`/`to_last` grouped `to` → two people
  - `host`, plus `guests` repeated 3 times → four people
  - `code` and `symbol` sibling fields reading `currency`, as `{code} {symbol}` → a matching pair
  - `{a} & {b}`, each reading `person` → one person, or two when `a` and `b` name different groups
  - two bare `{/sv_SE.word}` → two words

### Data

- Major data update. Shipped categories render as records with their building
  blocks as columns (`sv_SE.person` → `femalefirst`, `malefirst`; `misc.uuid` →
  `variant`), and `sv_SE.address` draws its postal code apart from its locality.
- Decide what a version promises about shipped data: its paths, its record
  columns, and whether a seed renders the same output across versions.
- `email.local` and `username` share their handle lists, while their name
  variants differ on purpose. Share the lists only if that is a clean win.

### Release

- Versioning — semver tags, `vX.Y.Z`. From v2 the module path carries `/vN`
  (`go.mod`, imports, the README's install lines). Pick the first tag, and reword
  the Decision "After the first tag, a new fence is a major version" to match.
- Changelog — a `## Changelog` section in the README, added with the first tag;
  the fence Decision's "changelog" points there.
- CLI without Go — investigate prebuilt binaries: GoReleaser publishing to Gitea
  releases, a container image, Homebrew and Scoop.
- Homepage — a simple page for fejkdata.
