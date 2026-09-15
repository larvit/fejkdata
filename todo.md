# Release checklist

## Before v0.1.0

### Data

- Major data update. Shipped categories render as records with their building
  blocks as columns (`sv_SE.person` → `femalefirst`, `malefirst`; `misc.uuid` →
  `variant`), and `sv_SE.address` draws its postal code apart from its locality.
  It also settles what a version promises about shipped data: its paths, its
  record columns with their datatypes and nulls — a column of one reference alone
  takes both, so typing a column or adding a null breaks its readers — and whether
  a seed renders the same output across versions.
- `email.local` and `username` share their handle lists, while their name
  variants differ on purpose. Share the lists only if that is a clean win.

### Release

- Versioning — semver tags, starting at `v0.1.0`; `v1.0.0` once the grammar
  settles. From v2 the module path carries `/vN` (`go.mod`, imports, the README's
  install lines), so fences ship batched into as few majors as possible. Reword
  the Decision "After the first tag, a new fence is a major version" to match:
  before `v1.0.0` a fence ships in a minor.
- Changelog — `CHANGELOG.md`, started with `v0.1.0`; the fence Decision's
  "changelog" links there.
- CLI without Go — investigate prebuilt binaries: GoReleaser publishing to Gitea
  releases, a container image, Homebrew and Scoop.
- Homepage — a simple page for fejkdata with an in-browser generator: the library
  compiled to WebAssembly, so visitors generate as much data as they like in their
  own browser.
