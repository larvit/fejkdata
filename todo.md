# Release checklist

## Before v0.1.0

### Data

- Major data update. Shipped categories render as records with their building
  blocks as columns (`sv_SE.person` → `femalefirst`, `malefirst`; `misc.uuid` →
  `variant`), and `sv_SE.address` draws its postal code apart from its locality.
- `email.local` and `username` share their handle lists, while their name
  variants differ on purpose. Share the lists only if that is a clean win — a
  reference between shipped categories is a major once tagged.

### Release

- CLI without Go — investigate prebuilt binaries: GoReleaser attaching them to
  the Gitea release the tag workflow publishes, a container image, Homebrew and
  Scoop. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag.
- Homepage — a simple page for fejkdata with an in-browser generator: the library
  compiled to WebAssembly, so visitors generate as much data as they like in their
  own browser.
