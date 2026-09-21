# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). A release's
`Breaking` section comes first and names each rejected spelling with its
replacement, and each removed path, column or flag.

## [Unreleased]

### Added

- First release: the CLI, the library and the shipped data set.
- Table categories: a category JSON naming a `rows` TSV beside it, with the options
  `key`, `name`, `weight` and `parent`; a path selects a row by key or name,
  `misc.territory[SE]`, and descends to a linked table by name; linked tables draw
  consistently within one render and draw group. `rows` is an option, so no
  template may carry a field of that name. A `name` without a `key` resolves
  inside the table's `parent`. Refused at `New`: a `name` without a `key` or a
  `parent`, a name repeating inside one parent row, a name spelling another row's key, a table named like a column of any
  table above it, a table whose format or cell references its own family, and,
  within one render and draw group, a path drawing a table another path selects a
  row of, or two paths pinning different rows of one table. Two cells are weighed
  against each other only where they can render together: not two rows of one table, and
  not rows under different rows of a table a path pinned.
- `New` refuses a root choice of templates sharing one format and one set of string
  fields, naming the rows TSV to write instead.
- `misc.territory`, `misc.currency`, `misc.language`, `misc.httpstatus` and
  `misc.mimetype` are tables. `misc.territory` is every ISO 3166-1 territory that has
  a capital, a currency and a TLD, with the columns `calling-code`, `capital`,
  `country`, `currency`, `flag`, `languages`, `numeric` and `tld` added, `country`
  holding the alpha-2 code of the sovereign state it belongs to and its own where it
  is one; `misc.currency` the
  current ISO 4217 currencies with a minor unit, with `decimals` and `numeric`
  added, and its symbols from CLDR; `misc.language` every ISO 639-2 entry carrying a
  639-1 code, with a `code3` column holding its 639-2/T code; `misc.httpstatus` each
  code the IANA registry lists with a plain reason phrase; and `misc.mimetype` each
  type mime-db records as IANA-registered and gives a filename extension.
  `DATA-LICENSES.md` lists each source.
- `misc.timezone`, `misc.car` and `misc.useragent` are tables too. `misc.timezone`
  is every zone tzdb `zone.tab` gives a shipped territory, with `offset` holding the
  zone's standard UTC offset and `territory` linking to `misc.territory`, so
  `misc.territory[SE].timezone` draws `Europe/Stockholm`; zones of a territory
  `misc.territory` does not ship, Antarctica's among them, and the constant `UTC`
  are gone; spell that one as the text `"UTC"`. A `weight` column carries the
  population GeoNames records in each zone, so a draw favours the zones people live
  in. `misc.useragent` is the top-user-agents desktop and mobile lists with `browser`,
  `device` and `os` columns. `misc.car` keeps its makes and models in a `make` and a
  `model` column.
- `misc.httpmethod`, `misc.protocol` and `misc.port` are IANA's registries as tables.
  `misc.httpmethod` is the eight methods RFC 9110 defines and PATCH, carrying
  `safe` and `idempotent` as `true` or `false`; `misc.protocol` the keyword, its
  `number` and its `name`, selectable by either; `misc.port` renders a TCP port number,
  with the IANA `service` name beside it, keeping the first service the registry
  describes for a port so a number selects one row.
- `misc.tld` is the IANA root zone as a table: every delegated TLD keyed with its
  leading dot, `.se` and `.xn--p1ai`, the register's `type` beside it, and a `unicode`
  column holding the form the register displays, so `misc.tld[.рф]` selects the row
  `.xn--p1ai` renders. A delegation the register marks "Not assigned" does not ship.
- `misc.loglevel` is the eight syslog severities: the `keyword` a configuration writes,
  `err` and `info`, keyed by the numerical `code` a PRI encodes and selectable by
  either, with the `severity` RFC 5424 spells beside it.
- `geo.SE` and `geo.US`: five linked tables per country, `region`, `municipality`,
  `locality`, `postal-code` and `street`, weighted by population and address counts
  and built from SCB, GeoNames, Trafikverket NVDB and the US Census Bureau, and an
  `address` record over one consistent draw of them. `sv_SE.address` and
  `en_US.address` read those records, so `en_US.address.street` no longer carries
  `name` and `suffix`, and a locale folder loads only beside `geo`.
- `{date(from,to,'layout')}` and `{time('layout')}`: a second between two days, or
  within one, in a single-quoted Go layout, drawn in UTC; `from` may equal `to`.
  `sv_SE.date`, `en_US.date`, `sv_SE.time` and `en_US.time` render through them, so
  `date.year`, `date.month`, `date.day`, `time.hour`, `time.minute`, `time.minute.t`,
  `time.sec` and `time.ampm` are no longer paths — a part of a date is now its own
  `{date(…,'2006')}`. `misc.datetime` is an RFC 3339 instant.
- `sex`, `first-name` and `last-name` tables in `sv_SE` and `en_US`, weighted by
  bearers from SCB, the SSA and the Census Bureau; `first-name` links to `sex`, and a
  name both sexes carry is a row under each. `person` reads them, so its columns are
  `first`, `last`, `prefix` and `sex`; `person.femalefirst` and `person.malefirst` are
  no longer paths — draw `sex[f].first-name` and `sex[m].first-name` instead.
  `en_US.title` links to `sex` as well, so `en_US.person.prefix` draws `Mr` or `Ms`
  without contradicting the record's `sex`; `sv_SE.title` is an unsexed table of the
  same shape.
- `sv_SE.personnummer` and `sv_SE.samordningsnummer`, Skatteverket's test series
  from a `sv_SE.birth-number` table under `sex`, in place of `sv_SE.ssn`, whose
  `ssn.mmdd`, `ssn.mmdd.m` and `ssn.mmdd.d` go with it. `en_US.ssn` now draws the
  ranges the SSA assigns and carries the columns `area`, `group` and `serial`, so
  `--format csv en_US.ssn` writes a header where it used to fail; `en_US.itin` is new
  and carries the same three. `person.prefix` is null where a person has no title,
  where it used to be an empty string, so `--format sql` writes `NULL` and a `string`
  struct field reading it becomes `*string`.
- `ErrNoColumns` is exported, so a caller can tell the one record fence a path can
  answer from the rest.
- An error names a spelling that runs: a layout is named single-quoted and free of
  its own quotes, a row of a table with no key is named as the path that selects it,
  `sv_SE.sex[f].first-name[Kim]`, and a category with no columns names the record
  that gives it one.
