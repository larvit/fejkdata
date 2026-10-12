# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). A release's
`Breaking` section comes first and names each rejected spelling with its
replacement, and each removed path, column or flag.

## [Unreleased]

The first release: a Go library and a CLI that render locale-aware fake data from JSON
templates, and the data they ship with. The
[README](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md) documents each part.
Install the CLI with `go install github.com/larvit/fejkdata/cmd/fejkdata@v0.1.0`, or the
library with `go get github.com/larvit/fejkdata@v0.1.0`. Before 1.0.0 a minor may break,
and any release may change what a seed draws, so pin one exact version
([Versioning](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md#versioning)).

### Added

- The CLI, `fejkdata`, renders exactly the template on its stdin and adds nothing:
  `echo '{/sv_SE.person} hihi' | fejkdata` prints e.g. `Sara Eriksson hihi`. `--seed`
  repeats a run's output while the version and data stay the same, `-n` renders the
  template that many times, and `--list` prints the paths the data offers.
  `--data-path` layers a directory of your own data over the shipped set, and
  `--no-shipped-data` loads that directory alone. `--format json|ndjson|csv|sql` writes
  a record's columns, one record per row. The CLI exits `0` on success, `1` on a runtime
  error such as data that fails to load, and `2` on misuse.
- The library, `github.com/larvit/fejkdata`, for Go 1.22 or later on a 64-bit target.
  `New` loads the modules `WithDataFS` names and the folders `WithDataPath` names, and
  fails, naming `WithDataFS`, where neither is given. `WithSeed` makes its draws
  repeat while the version and data stay the same. `Fake` renders a path, and
  `FakeTemplate` and `NewTemplate` render a template. `FakeRecord`, `FakeRecordTemplate`
  and `NewRecordTemplate` render a record, whose `JSON`, `CSVHeader`, `CSVLine` and
  `SQLInsert` methods write its columns as the CLI does. `FakeStruct` fills a struct's
  `fake:"…"` tagged fields. `List` gives the paths `Fake` renders, direct descents only,
  and `IsTemplate` tells an inline template from a path. Data that fails to load gives
  an error matching `ErrLoad`, and a record with no fields gives one matching
  `ErrNoColumns`.
- Package `github.com/larvit/fejkdata/cli` is the `fejkdata` command:
  `cli.Main(modules...)` runs it over the modules given, so
  `cli.Main(append(data.Modules(), mydata)...)` builds a CLI carrying your own data.
  That CLI names itself by the file it runs from, and its `--version` prints its own
  module's version before fejkdata's.
- The template language
  ([Data format](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md#data-format)).
  Every value in a data file is a string, a choice, a template, a table, a number or a
  boolean. A `{…}` token reads a sibling field, a name, or a reference from the data
  root, the file's folder or the folder above (`{/sv_SE.person.last}`, `{.x}`, `{..x}`),
  or it calls a builtin. Every `{…}` draws afresh, and a name is the one way to keep a
  pick: `{/sv_SE.person as p}{p.first} {p.last}` is one person. The builtins draw check
  digits (`luhn`, `mod11`, `ean`), samples (`digits`, `int`, `float`, `uuid`, `ulid`,
  `iban`, `date`, `time` and more), a counter (`seq`), arithmetic (`calc`) and
  transforms (`lowercase`, `uppercase`, `ascii`). `New` refuses data that breaks the
  grammar, could mean two things or cannot render a valid value.
- Records ([Records](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md#records)):
  a template seen as columns. A column is a string unless it declares a `datatype`,
  `integer`, `number` or `boolean`, and a `null` item draws it as null.
- Tables ([Table](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md#table)): a
  category naming a `rows` TSV beside it, with the options `key`, `name`, `weight` and
  `parent`. A path selects a row by key or name, `misc.territory[SE]`, descends to a
  linked table, `geo.SE.locality[Lund].street`, and steps up to the parent row,
  `{/misc.timezone[Europe/Oslo]..territory.name}`.
- Data sources ([Data](https://github.com/larvit/fejkdata/blob/v0.1.0/README.md#data)):
  a directory is a namespace, and each JSON file in it a category. A `.fejkdata.json`
  manifest at a source's root may carry a `namespace`, the dot path the source mounts
  under, and an `index`, with which each category loads on the first call reaching it.
- The shipped data, in Go packages that each export `FS` and `Modules()`:
  `github.com/larvit/fejkdata/data/en_US` (package `enus`), `…/sv_SE` (`svse`),
  `…/misc`, `…/geo/SE` (`geose`) and `…/geo/US` (`geous`). `data.Modules()`, from
  `github.com/larvit/fejkdata/data`, lists every one, and the CLI carries them all.
  [`DATA-LICENSES.md`](https://github.com/larvit/fejkdata/blob/v0.1.0/DATA-LICENSES.md)
  names each source.
  - `sv_SE` and `en_US`: an `address`, a `person` and more per locale. `sex`,
    `first-name` and `last-name` are tables weighted by bearers, from SCB, the SSA and
    the Census Bureau. A person's first name and sex agree, and so does an `en_US`
    person's title; `person.prefix` is null where a person has no title.
    `sv_SE.personnummer` and `sv_SE.samordningsnummer` draw Skatteverket's test series.
    `en_US.ssn` and `en_US.itin` draw the ranges the SSA and the IRS assign, with the
    columns `area`, `group` and `serial`, and `en_US.phone` draws an exchange NANPA may
    assign.
  - `misc`: IANA's registries of HTTP methods, protocols, ports and root-zone TLDs, the
    HTTP status codes, ISO 3166-1 territories, ISO 4217 currencies with their CLDR
    symbols, ISO 639 languages, the syslog severities, mime-db's media types, tzdb's
    zones weighted by the people living in each, user agents and cars, all as tables;
    and coordinates, credit card numbers, RFC 3339 instants, emoji, MAC addresses,
    ObjectIds and UUIDs.
  - `geo.SE` and `geo.US`: five linked tables per country, `region`, `municipality`,
    `locality`, `postal-code` and `street`, weighted by population and address counts
    and built from SCB, GeoNames, Trafikverket NVDB, the US Census Bureau and Wikidata,
    and an `address` record that draws a street and a postal code inside one
    locality; `sv_SE.address` and `en_US.address` print it. Every Swedish postort
    (postal town) the import can place ships with all its streets. A US place ships
    only if it is incorporated and holds most of some postal code's land. Of those
    places, every state capital but Honolulu ships, as does every place of 100,000
    people or more and a tenth of all of them drawn by population. Each US place
    carries a quarter of its streets, at least ten.
