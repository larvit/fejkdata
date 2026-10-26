# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). A release's
`Breaking` section comes first and names each rejected spelling with its
replacement, and each removed path, column or flag.

## [Unreleased]

### Breaking

- `{ascii(x)}` and `{slugify(x)}` fold 129 more Latin letters to base letters, most of
  Latin Extended-A, part of Latin Extended-B and `ẞ`, where they dropped them: `Wałęsa`
  becomes `Walesa`, not `Wasa`, and slugs to `walesa`.
- `sv_SE.company` loses its fields `base` and `suffix`. No field in this release gives a
  bound company its name or its legal form. `registered` draws among the legal forms,
  and each form (`ab`, `e`, `hb`, `brf`, `ek`, `kb`) holds a `name` and, but for `e`, a
  `written` company. A form's `name` draws a fresh name; it doesn't read the one inside
  a company already drawn. The category reads `sv_SE.first-name`, `sv_SE.last-name` and
  `geo.SE.locality`, which `sv_SE` already carried.
- `sv_SE.company` draws its legal form by Bolagsverket's new registrations of 2025, and
  builds its name in a shape, a pattern of words, that the register's active companies'
  names have. What a seed draws changes.
- `sv_SE.company` spells a legal form as the register's active companies do, in the
  spellings that 1% or more of a form's names use. An aktiebolag ends `AB` or
  `Aktiebolag`. A handelsbolag ends `Handelsbolag` or `handelsbolag`, or starts
  `Handelsbolaget`. A kommanditbolag ends `Kommanditbolag` or `kommanditbolag`, or starts
  `Kommanditbolaget`. An ekonomisk förening ends `Ekonomisk förening`,
  `ekonomisk förening`, `ek. för.` or `Ek. för.`. A bostadsrättsförening starts
  `Bostadsrättsföreningen`. An enskild näringsidkare has no legal-form word. `HB`, `KB`
  and `Brf` are under 1% of the names, so they don't ship.
- `sv_SE.phone` loses its fields `a`, `b`, `c` and `prefix`, and no field replaces them.
  `sv_SE.phone-area` now draws an area code and `sv_SE.phone-subscriber` subscriber digits,
  and `sv_SE.phone` reads both. A fixed number's area code is one of the 264 geographic area
  codes in PTS's numbering plan of 2024-01-08, not one of 20. Its subscriber number has any
  length the plan allows that code, five to eight digits, grouped as Språkrådet writes them,
  `08-16 01 01` or `031-668 01 50`, and starts with no `0`. What a seed draws changes.

### Added

- `{slugify(x)}`, a transform that makes a name safe in an email or a URL: `Anna Lena`
  becomes `anna-lena`. [Transforms](https://github.com/larvit/fejkdata#transforms) states
  the rules.
- `sv_SE.phone.national` and `sv_SE.phone.international`, the two spellings of one
  number: `070-123 45 67` and `+46 70 123 45 67`. Bind a number to read both,
  `{/sv_SE.phone as p}{p.national} {p.international}`.
- `sv_SE.phone-area`, a table of PTS's 264 geographic area codes, keyed by the code as
  Swedes write it, such as `08` or `0480`, with its `ndc` without the `0`, the `name` of its
  place and a `subscriber` number that fits it. `sv_SE.phone-subscriber` holds
  subscriber numbers by length, in the fields `digits-5-6`, `digits-5-7`, `digits-6-7` and
  `digits-6-8`.
- Prebuilt binaries for Linux, macOS and Windows on amd64 and arm64, packages for apt,
  dnf, apk, Homebrew and Scoop, and the image `larvit/fejkdata` on Docker Hub, also at `ghcr.io/larvit/fejkdata`, so
  the CLI installs without Go. 0.1.0 ships them too; the README's
  [Install](https://github.com/larvit/fejkdata#install) lists each command.
- The README's [quick starts](https://github.com/larvit/fejkdata#quick-start) for bulk fixtures and Go tests each render one person whose name, sex,
  personnummer, email and address agree.
- `sv_SE.organisationsnummer`, a Luhn-valid Swedish organisation number such as
  `556047-3521`. Its first digit is the group digit Bolagsverket gives aktiebolag (5),
  handelsbolag and kommanditbolag (9) or föreningar (7, 8). A drawn number may belong to
  a real organisation. Its `vat` field is that organisation's VAT number.
- `sv_SE.vat`, `SE`, an organisation number's ten digits and `01`. It passes the format
  and the checksum, not a VIES lookup.
- `nb_NO`, Norwegian in the Bokmål standard, with `nb_NO.person` and the tables it reads:
  `nb_NO.first-name` and `nb_NO.last-name`, holding the names that 200 or more persons bear
  in Statistics Norway's count of 2025, weighted by those persons; `nb_NO.sex`, `kvinne` and
  `mann`; and `nb_NO.title`, `dr.` and `professor`. A first name that women and men both
  bear has a row under each sex, and a person's first name and sex agree. Statistics
  Norway's tables are CC BY 4.0, so a program showing the names credits "Kilde:
  Statistisk sentralbyrå".
- `nb_NO.fodselsnummer`, a Norwegian national identity number from Skatteetaten's synthetic
  series: a birth date from 1940 to 2025 with 80 added to the month, three individ digits and
  the two check digits Skatteetaten defines. A checker that requires a month of 1 to 12
  refuses it. Its `birth` field is the date and the first individ digit. It reads
  `nb_NO.birth-day`, the 365 days of a year as `ddmm`, and `nb_NO.individ`, the five digits
  that each `nb_NO.sex` allows its last individ digit.
- The builtin `{fnr(x)}` completes a fødselsnummer. It reads the last seven digits emitted,
  the date and the first individ digit, and adds a second individ digit, the third that `x`
  reads, and the two check digits. It adds nothing where `x` is not one digit.
  `{/nb_NO.fodselsnummer.birth}{fnr(s.individ.digit)}` agrees with a bound sex `s`.
- `nb_NO.phone`, a Norwegian phone number such as `46 12 34 56`, mobile ten times as often as
  fixed, grouped in pairs as the Language Council of Norway (Språkrådet) writes it. Its `national` and `international`
  (`+47 46 12 34 56`) fields are the two spellings of one number. It reads `nb_NO.phone-mobile`
  and `nb_NO.phone-fixed`, tables of the four-digit prefixes behind which Nkom's E.164 register has
  assigned every number to providers, keyed by the prefix as written, `phone-fixed[56 12]`. A number's `block`
  is that prefix and its `subscriber` the last four digits. Nkom's data is NLOD 2.0, so a
  program showing a number credits Nkom as DATA-LICENSES.md asks.
- `nb_NO.date`, a date from 1970 to 2029 such as `02.04.2019`.
- `nb_NO.email`, a lower-case ASCII address at `example.com`, `example.org` or `example.net`, or
  a subdomain of one. Half of its local parts are a first and a last name drawn from `nb_NO.first-name` and `nb_NO.last-name`.
- `nb_NO.company`, an invented name and a legal form, ENK (a sole proprietorship, whose name has no
  abbreviation), AS, DA, SA, ANS, ASA or KS, such as `Nordicvik Holding AS`. Each form is drawn as often as
  Enhetsregisteret held active units of it, and `nb_NO.company.as`, `.enk`, `.da`, `.sa`, `.ans`, `.asa` and
  `.ks` each draw one form. The name has a shape the register's names of
  the form have, and reads `nb_NO.company-coined`, `nb_NO.company-word`, `nb_NO`'s names and
  `geo.NO.locality`, so a name may hold a poststed, such as `Bergen Nordicvik AS`.
  Enhetsregisteret's data is NLOD 2.0, so a program showing a company credits Brønnøysundregistrene, and
  Kartverket where the name holds a poststed, as DATA-LICENSES.md asks.
- `geo.NO`, Norway's places, in the package `github.com/larvit/fejkdata/data/geo/NO` (`geono`):
  `region` (fylke), `municipality` (kommune), `locality` (poststed), `postal-code` (postnummer)
  and `street`, from every street address in Kartverket's address register, with populations
  from Statistics Norway, and an `address` record binding one poststed. Both publish under
  CC BY 4.0, so a program showing an address credits Kartverket and Statistics Norway as
  DATA-LICENSES.md asks.
- `nb_NO.address`, such as `Storgata 12B\n0155 Oslo`, printing `geo.NO.address`, with the columns
  `street`, `street-number`, `postal-code` and `locality`. `nbno.Modules()` returns `geono`'s
  module too.
- `sv_SE.company-word`, the trade words that 100 or more of Bolagsverket's active companies'
  names carry, and `sv_SE.company-coined`, made-up one-word names.

### Changed

- A patch of the latest release follows each update to Go, a Go minor included, to a module
  fejkdata requires, or to GoReleaser, once the tests pass with it. Until 1.0.0, a Go minor
  may stop supporting an older system, such as a macOS version; the patch's changelog
  section names each one under Breaking.

## [0.1.2] - 2026-10-20

### Security

- The binaries and image are built with Go 1.27.3.

## [0.1.1] - 2026-10-17

### Security

- The binaries and image are built with Go 1.27.2.
- The binaries no longer carry Go 1.27.1's known vulnerabilities: GO-2026-6599 (CVE-2026-94448), GO-2026-6600 (CVE-2026-97030), GO-2026-6603 (CVE-2026-78659), GO-2026-6604 (CVE-2026-56857), GO-2026-6605 (CVE-2026-56866), GO-2026-6607 (CVE-2026-97031), GO-2026-6608 (CVE-2026-94440), GO-2026-6609 (CVE-2026-78667), GO-2026-6610 (CVE-2026-78660), GO-2026-6611 (CVE-2026-78669), GO-2026-6612 (CVE-2026-78663), GO-2026-6613 (CVE-2026-94439), GO-2026-6617 (CVE-2026-97032).

## [0.1.0] - 2026-10-13

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
  repeats a run's output while the version and data stay the same, and without it each
  run takes a new seed; `-n` renders the template that many times, and `--list`
  prints the paths the data offers.
  `--data-path` layers a directory of your own data over the shipped set, and
  `--no-shipped-data` loads that directory alone. `--format json|ndjson|csv|sql` writes
  a record's columns, one record per row. The CLI exits `0` on success, `1` on a runtime
  error such as data that fails to load, and `2` on misuse.
- The library, `github.com/larvit/fejkdata`, for Go 1.22 or later on a 64-bit target.
  `New` loads the modules `WithDataFS` names and the folders `WithDataPath` names, drawing
  from the seed `WithSeed` gives. Without a seed, `New` fails and names `WithSeed`;
  without data, it fails and names `WithDataFS`. The same seed repeats the draws
  while the version and data stay the same. `Fake` renders a path, and `FakeTemplate`
  and `NewTemplate` render a template. `FakeRecord`, `FakeRecordTemplate` and
  `NewRecordTemplate` render a record, whose `JSON`, `CSVHeader`, `CSVLine` and
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
