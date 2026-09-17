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
  template may carry a field of that name. Refused at `New`: a `name` without a
  `key`, a name spelling another row's key, a table named like a column of any
  table above it, a table whose format or cell references its own family, and,
  within one render and draw group, a path drawing a table another path selects a
  row of, or two paths pinning different rows of one table.
- `New` refuses a root choice of templates sharing one format and one set of string
  fields, naming the rows TSV to write instead.
- `misc.country`, `misc.currency`, `misc.language`, `misc.httpstatus` and
  `misc.mimetype` are tables. `misc.country` is every ISO 3166 country that has a
  capital, a currency and a TLD, with the columns `calling-code`, `capital`,
  `currency`, `flag`, `languages`, `numeric` and `tld` added; `misc.currency` the
  current ISO 4217 currencies with a minor unit, with `decimals` and `numeric`
  added, and its symbols from CLDR. `DATA-LICENSES.md` lists each source.
- `geo.SE` and `geo.US`: five linked tables per country, `region`, `municipality`,
  `locality`, `postal-code` and `street`, weighted by population and address counts
  and built from SCB, GeoNames, Trafikverket NVDB and the US Census Bureau, and an
  `address` record over one consistent draw of them. `sv_SE.address` and
  `en_US.address` read those records, so `en_US.address.street` no longer carries
  `name` and `suffix`, and a locale folder loads only beside `geo`.
