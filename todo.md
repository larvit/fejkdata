# Release checklist

## Before v0.1.0

### Data

Plan for the major data update. Research notes: `~/code/claude-handoffs/fejkdata/2026-09-16/`.

#### Structure (proposal, pending review)

`New` costs ~23 ms and 58k allocations for today's 200 KB, one allocation per
node, so real-world tables cannot ship as nested JSON objects. JSON says how a
value is composed; TSV says which values exist.

- Add a table node: a category JSON naming a `rows` TSV beside it, whose header
  names the columns and whose rows are the draw. Options `key` (the column a path
  selects a row by), `name` (the column a path also selects by), `weight` (the
  column that skews the draw), `parent` (the table a column links to). `{name}`
  tokens read columns; a cell is a string node, so it may carry tokens. Index
  rows on first draw; at load prove only the header, the options and any cell
  token. A TSV nothing names is a load error.
- In a data file, refuse a choice whose items are templates with one identical
  field set, naming the TSV to write; an inline template is exempt, having no
  file. A choice of strings stays a choice.
- Convert `misc.country`, `currency`, `language`, `httpstatus`, `mimetype`,
  `timezone`, `car` and `useragent` to tables.
- Let a path select a row by key or by name, `geo.SE.municipality[0180]` and
  `geo.SE.municipality[Stockholm]`, and descend to a linked table by name,
  `geo.SE.region[Skåne län].street`; a key or a linked table name may not equal
  a column name (load error). A name that matches several rows outside a pinned
  parent is an error listing the keys; inside one, `region[IL].locality[Springfield]`
  resolves. Brackets need quoting in zsh; `:` would not.
- Draw linked tables consistently within one render and draw group: the first
  table drawn pins its ancestors, and a descendant is drawn inside them.
- Keep country data under `geo/<alpha2>/`; a locale's `address` reads its
  country's tree by reference, and a locale's `country` table carries the
  localised names keyed by alpha2.
- Ship full registers as packs, each a Go module with its own `embed.FS` and a
  zip for `--data-path`; the default embed stays under ~1 MB per country and
  `New` under ~50 ms.
- Add `DATA-LICENSES.md` listing every dataset, its licence and its attribution,
  and a `data-import/` directory of Python scripts that rebuild each TSV from its
  source, so a refresh is one command per dataset.
- Make every shipped category a record with its building blocks as columns
  (`sv_SE.person` → `first`, `last`, `sex`, done in step 3; `misc.uuid` → `variant`).
- Share the handle lists between `email.local` and `username` only if that is a
  clean win — a reference between shipped categories is a major once tagged.

#### Geo tree

Five linked tables per country, generic names so a template ports across
countries; the README maps each to the native term.

| Table | SE | US | Weight |
|---|---|---|---|
| `region` | län (21) | state and DC (50; Hawaii has no incorporated place) | population |
| `municipality` | kommun (290) | county with a shipped place | population |
| `locality` | postort, tätort population | place of 25,000+ | population |
| `postal-code` | postnummer with street delivery | ZCTA of a shipped place | one; address ranges |
| `street` | gatunamn, top 10 per postort | street name, top 10 per place | segments; address ranges |

- Shipped in step 2, README Data. `geo.SE.address` is a record over one consistent
  draw. Each region row carries its timezone, each locality its centroid.
- Let `geo.SE.locality[Lund].address` descend from a selected row into the
  template beside the family: the path step must reach a sibling category and the
  outer selector's pins seed every draw group of the render.
- Ship the fuller sets, every US place of 10,000 and more streets per locality, as
  packs; `--min-population` and `--streets-per-locality` on the scripts build them.
- Fill the 398 Swedish localities weighted 200 from SCB småorter before v0.1.0.
- Give the address records one column set across countries: `region` and
  `municipality` as building-block columns on `geo.SE.address` too, in step 6.
- v0.1.0 ships SE and US; then NO, DK, FI, NL, FR, AU, CA, ES, GB, DE.
- Revisit an application to Lantmäteriet for the exact street to postnummer
  pairing after v0.1.0; today a street goes to the nearest postal code centroid.

#### Builtins the data cannot express

- `{date(from,to,'layout')}` and `{time('layout')}` shipped in step 3; `age()` is
  rejected, README Decisions. Add a `unix` layout once something needs it.
- Derivations: `{isin()}` (Luhn over letters expanded to digits), `{cusip()}`,
  `{aba()}` (3-7-1 weights), `{vin()}` (position 9 over the whole; a sample taking
  the WMI, since the check sits mid-string).
- Samples: `{base64url(n)}` for JWT shapes; `{btc()}`, `{eth()}` after v0.1.0
  (Base58Check and EIP-55 need sha256 and keccak).

#### Categories

Union of @faker-js/faker, Python Faker, gofakeit, Datafaker, Bogus, Mimesis, Chance
and Ruby faker (`research-libraries.md`). What sets fejkdata apart, and what every
issue tracker asks for: real values, frequency weights, and facts that agree across
a record. No pop-culture catalogues. An LLM writes only non-factual lists
(catchphrases, hacker phrases, product adjectives), never names, places, codes or
ids. Shape: T = table, t = template, c = choice.

`sv_SE` (`research-sources-se.md`)

| Category | Shape | Source | Licence |
|---|---|---|---|
| `person` first (female, male, generic), middle, last, weighted | T | SCB 2022 whole-population xlsx, Skatteverket 2026 surnames | CC0, "Källa: SCB" |
| `person` title, gender, birthdate, age, blood type weighted | t | geblod.nu distribution | facts |
| `personnummer`, `samordningsnummer` | t | Skatteverket test series: date + 238/239, Luhn | CC0 |
| `organisationsnummer` by form, `vat` | t | Bolagsverket group digits, Luhn, `SE…01` | facts |
| `company` name patterns, legal form weighted | t | Bolagsverket registrations 2025 | CC BY 2.5 SE |
| `sni` industry codes | T | SCB SNI 2025 | "Källa: SCB" |
| `ssyk` occupations | T | SCB SSYK 2012 | "Källa: SCB" |
| `phone` mobile, fixed per area code, E.164 | T+t | PTS numbering plan 2024 | facts |
| `bankgiro`, `plusgiro`, `bankaccount` clearing + Luhn/mod-11 | T+t | Bankinfrastruktur CSV, Bankgirot rules | facts |
| `licenseplate` `ABC 123`, `ABC 12A`, blocked combinations | t | Transportstyrelsen | facts |
| `car` make, model weighted by registrations | T | Mobility Sweden 2025 | cite |
| `word` noun, verb, adjective, adverb; `sentence` | T+t | SALDO | CC BY 4.0 |
| `date`, `time`, month and weekday names, `holiday` | t+T | CLDR sv, lag 1989:253 | Unicode |
| `price` `1 234,56 kr`, `email` domains, `url` `.se`/`.nu` | t | curated | — |
| `country` localised names keyed by alpha2 | T | CLDR sv territory names | Unicode |
| `address` | t | `{/geo.SE…}` | — |

`en_US` (`research-sources-world.md`, part1–4)

| Category | Shape | Source | Licence |
|---|---|---|---|
| `person` first by sex weighted, last weighted, middle, generic | T | SSA baby names, Census 2010 surnames | public domain |
| `ssn`, `itin`, `ein` valid ranges | t | SSA POMS, IRS | facts |
| `company` names, suffix, `naics` | t+T | SEC tickers for patterns, NAICS 2022 | public domain |
| `phone` NANP with valid NPA/NXX | t | NANPA rules | facts |
| `routing` ABA with check, `bankaccount` | t | Fed prefix ranges | facts |
| `licenseplate` per state | T | hand-authored patterns | facts |
| `car` make, model | T | NHTSA vPIC | public domain |
| `word`, `sentence`, `paragraph` | T+t | Moby POS or WordNet | public domain / WordNet |
| `date`, `time`, names, `holiday` | t+T | CLDR en | Unicode |
| `price` `$1,234.56`, `email`, `url`, `address` | t | curated, `{/geo.US…}` | — |

`misc` (locale-neutral)

| Category | Shape | Source | Licence |
|---|---|---|---|
| `country` name, alpha2, alpha3, numeric, calling code, TLD, capital, currency, flag, languages | T | datasets/country-codes | PDDL |
| `currency` code, numeric, name, symbol, decimals | T | ISO 4217 list-one, CLDR symbols | free, Unicode |
| `language` ISO 639-1/2 | T | LoC | public domain |
| `timezone` zone, country, offset | T | tzdb zone1970.tab | public domain |
| `mimetype` type, extensions | T | IANA + mime-db | CC0, MIT |
| `httpstatus`, `httpmethod`, `port`, `tld`, `protocol`, `loglevel` | T/c | IANA | CC0 |
| `useragent` per browser | T | top-user-agents | MIT |
| `ip` v4 documentation ranges, private, CIDR; `ipv6` `2001:db8::/32`; `mac` locally administered | t | RFC 5737, 1918, 9637 | facts |
| `creditcard` per network: IIN, length, CVV, expiry | T+t | network rules, `{luhn()}` | facts |
| `bic`, `iban` (exists), `isin`, `cusip` | t | structure rules | facts |
| `ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin`, `sku` | t | check-digit rules, GS1 demo prefix 952 | facts |
| `product` name, adjective, material, department, `garmentsize` | T/c | LLM-written lists | — |
| `airport` IATA, ICAO, country, municipality; `airline`; `aircraft`; `flight`, `seat`, `pnr` | T+t | OurAirports, Wikidata | public domain, CC0 |
| `element`, `unit`, `planet` | T | PubChem, BIPM, NASA | public domain, CC BY 4.0 |
| `animal` by class, `plant` | T | Wikidata | CC0 |
| `bloodtype` weighted (global) | c | AABB | facts |
| `color` named, hex, rgb, hsl, cmyk | T+t | CSS Color 4, XKCD | W3C, CC0 |
| `programminglanguage`, `os`, `database` column/type/engine, `fileext`, `semver` (exists) | T/c | Linguist, curated | MIT |
| `commit` message, branch, `sha` | t | LLM-written, `{hex(40)}` | — |
| `password`, `hash`, `jwt`, `slug`, `hostname`, `avatar`, `placeholderimage` | t | shapes | — |
| `book` title, author from Gutenberg; `genre`; `instrument`; `dish`, `ingredient` | T | Gutenberg catalog, MusicBrainz, USDA | CC0 |
| `lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword`, `quote` | T/c | lorem ipsum, LLM-written | — |
| `coordinate` (exists), `direction`, `continent`, `emoji` (exists), `uuid`, `ulid`, `objectid` | c/t | — | — |

Later locales: nb_NO, da_DK, fi_FI, de_DE, en_GB, nl_NL, fr_FR, es_ES with person,
address, phone, national id, company and date names each.

#### Order of work

1. Table node, key and name selection, parent links, consistent draws, the
   choice-of-rows fence, `DATA-LICENSES.md`, `data-import/` — done.
2. `geo/SE` and `geo/US`, and `address` in both locales on top of them — done.
3. Weighted person names and valid ids in both locales; `date()` — done. Middle
   names wait for a draw group that shares its family's pins, so a second name
   is drawn under the same sex.
4. `misc` conversions and the new `misc` tables.
5. The remaining locale categories: company, phone, finance, vehicle, words.
6. Records with building-block columns across the shipped set; shape re-pin.

### Release

- CLI without Go — investigate prebuilt binaries: GoReleaser attaching them to
  the Gitea release the tag workflow publishes, a container image, Homebrew and
  Scoop. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag.
- Homepage — a simple page for fejkdata with an in-browser generator: the library
  compiled to WebAssembly, so visitors generate as much data as they like in their
  own browser.
