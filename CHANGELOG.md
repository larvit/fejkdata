# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). A release's
`Breaking` section comes first and names each rejected spelling with its
replacement, and each removed path, column or flag.

## [Unreleased]

### Added

- First release: the CLI, the library and the shipped data set.
- Table categories: a category JSON naming a `rows` TSV beside it, with the options
  `key`, `name`, `weight` and `parent`; a path selects a row by key or name,
  `misc.country[SE]`, and descends to a linked table by name; linked tables draw
  consistently within one render and draw group. `rows` is an option, so no
  template may carry a field of that name.
- `New` refuses a root choice of templates sharing one format and one set of string
  fields, naming the rows TSV to write instead.
- `misc.country`, `misc.currency`, `misc.language`, `misc.httpstatus` and
  `misc.mimetype` are tables. `misc.country` is the full ISO 3166 register with the
  columns `calling-code`, `capital`, `currency`, `flag`, `languages`, `numeric` and
  `tld` added; `misc.currency` the current ISO 4217 list with `decimals` and
  `numeric` added, and its symbols from CLDR. `DATA-LICENSES.md` lists each source.
