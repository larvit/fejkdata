# Release checklist

## Before v0.1.0

### Data

Plan for the major data update. Research notes: `~/code/claude-handoffs/fejkdata/2026-09-16/`.

#### Structure (proposal, pending review)

`New` costs ~23 ms and 58k allocations for today's 200 KB, one allocation per
node, so real-world tables cannot ship as nested JSON objects.

- Add a table node: a category JSON naming a `rows` TSV beside it, whose header
  names the columns and whose rows are the draw. Options `key` (the column a path
  selects a row by), `weight` (the column that skews the draw), `parent` (the
  table a column links to). `{name}` tokens read columns; a cell is a string node.
  Index rows on first draw; at load prove only the header, the options and any
  cell token.
- Reject a choice of same-shaped templates naming the table spelling, and convert
  `misc.country`, `currency`, `language`, `httpstatus`, `mimetype`, `timezone`, `car`.
- Let a path select a row by key, `geo.SE.municipality[0180]`, and descend to a
  linked table by name, `geo.SE.region[01].street`; a key or a linked table name
  may not equal a column name (load error).
- Draw linked tables consistently within one render and draw group: the first
  table drawn pins its ancestors, and a descendant is drawn inside them.
- Keep country data under `geo/<alpha2>/`; a locale's `address` reads its
  country's tree by reference, and a locale's `country` table carries the
  localised names keyed by alpha2.
- Ship full registers as packs, each a Go module with its own `embed.FS` and a
  zip for `--data-path`; the default embed stays under ~1 MB per country and
  `New` under ~50 ms.
- Add `DATA-LICENSES.md` listing every dataset, its licence and its attribution.
- Make every shipped category a record with its building blocks as columns
  (`sv_SE.person` → `femalefirst`, `malefirst`; `misc.uuid` → `variant`).
- Share the handle lists between `email.local` and `username` only if that is a
  clean win — a reference between shipped categories is a major once tagged.

#### Geo tree (proposal, pending review)

Five linked tables per country, generic names so a template ports across
countries; the README maps each to the native term.

| Table | SE | US | Weight |
|---|---|---|---|
| `region` | län (21) | state (56) | population |
| `municipality` | kommun (290) | county (3,234) | population |
| `locality` | postort (~1,780) | place (~19,500) | population |
| `postal-code` | postnummer (~10,500 deliverable) | ZCTA (33,791) | address count or 1 |
| `street` | gatunamn, top N per locality | street name, top N per place | address count |

- `geo.SE.address` is a record over one consistent draw: street, number, postal
  code, locality; `geo.SE.locality[stockholm].address` stays inside Stockholm.
- Countries, in order: SE, US, then NO, DK, FI, NL, FR, AU, CA, ES, GB, DE.
- Sweden's street ↔ postnummer pairing is not open: Lantmäteriet's addresses need
  an approved application under personal-data terms and PostNord's register is
  commercial. Decide: apply as larvit, compose NVDB streets with GeoNames
  postnummer per postort (approximate pairing), or both.

#### Sources (`research-geo-se.md`, `research-geo-world.md`)

| Data | Source | Licence |
|---|---|---|
| SE län, kommun, population, tätorter | SCB | CC0 |
| SE street names per kommun | Trafikverket NVDB Gatunamn | CC0 |
| SE postnummer → postort | GeoNames SE.zip (stale, retired codes included) | CC BY 4.0 |
| SE full addresses | Lantmäteriet Belägenhetsadress | CC BY 4.0 + personal-data terms, application |
| SE Göteborg, Malmö full addresses | OpenAddresses | CC0 |
| US states, counties, places, population, ZCTA | Census (Gazetteer, SUB-EST2025, ZCTA relationships) | public domain |
| US addresses, streets | DOT National Address Database | public domain |
| NO addresses | Kartverket Matrikkelen; Bring postnummer | CC BY 4.0; NLOD 2.0 |
| DK addresses | DAWA / DAR | CC BY 4.0 |
| FI addresses, postal codes | DVV/SYKE; Posti PCF/BAF | CC BY 4.0; Posti terms |
| NL addresses | BAG | CC0 |
| FR addresses | BAN, La Poste hexasmal | Licence Ouverte 2.0 |
| AU addresses | G-NAF | CC BY 4.0 EULA |
| CA addresses | StatCan National Address Register | StatCan Open Licence |
| ES addresses | Catastro INSPIRE AD (no Basque, Navarre) | Catastro licence, verify |
| GB streets, postcodes | OS Open Names, Code-Point Open (no house numbers) | OGL v3 |
| DE Gemeinden, addresses | Destatis GV100, 15 Länder Hauskoordinaten (no Bavaria) | dl-de/by-2-0 |
| Cities with population, any country | GeoNames cities500 | CC BY 4.0 |
| Names by frequency | SCB, SSA, Census | CC0, public domain |
| Never | OpenStreetMap, OpenPLZ, countries-states-cities-database | ODbL share-alike |

An LLM is a source only for non-factual lists (product adjectives, catchphrases,
hacker phrases), never for names, places, codes or ids.

#### Categories to add

Union of @faker-js/faker, Python Faker, gofakeit, Datafaker, Bogus, Mimesis, Chance
and Ruby faker (`research-libraries.md`). What sets fejkdata apart, and what every
issue tracker asks for: real values, frequency weights, and facts that agree across
a record. No pop-culture catalogues.

- Person: real first and last names weighted by official frequency tables (SCB, SSA,
  Census), middle name, generic (non-gendered) first names, gender, title, suffix,
  birthdate and age, job title/area/level, blood type weighted by prevalence, marital
  status, nationality and demonym, valid national ids (personnummer, samordningsnummer,
  SSN, EIN, and one per new locale), passport number, bio.
- Address: the geo tree, secondary address (`lgh 1201`, `Apt 4B`), PO box, building
  number, latitude and longitude inside the locality, timezone by country and region,
  cardinal direction, continent, full ISO 3166 country row (name per locale, alpha2,
  alpha3, numeric, calling code, TLD, capital, currency, flag emoji, languages), ISO
  3166-2 subdivisions.
- Company: name patterns per locale, legal suffix, valid organisationsnummer and VAT
  number (SE), EIN (US), DUNS, industry (SNI/NAICS), department, profession,
  catchphrase, slogan, buzzwords, domain from the name.
- Internet: ipv4 public/private/CIDR, ipv6, hostname, TLD, port, protocol, http method
  and version, log level, slug, password, JWT shape, hash (md5/sha1/sha256), avatar
  and placeholder image URLs, user agent per browser, safe email domains, full IANA
  mime list with extensions correlated.
- Finance: BIC, BBAN, ABA routing with check digit, credit card per network with
  valid prefix, length, CVV and expiry, bankgiro and plusgiro (SE), clearing number,
  ISIN and CUSIP with check digit, crypto addresses, stock ticker, transaction type
  and description, amount by currency.
- Commerce: product name, adjective, material, description, department, brand, SKU,
  promo code, EAN/UPC/ISBN/GTIN/ASIN, IMEI, garment size, price per locale.
- Date and time: `{date(from,to,layout)}` and `{time()}` builtins, birthdate by age
  range, weekday and month names per locale, ISO datetime, unix time, duration, cron.
- Phone: national, international and E.164 per locale, area code, extension.
- Text: per-locale word classes (noun, verb, adjective, adverb, preposition), sentence,
  paragraph, lorem ipsum, question, quote, hacker and hipster phrases.
- Vehicle: real make and model pairs, type, fuel, transmission, color, year, VIN with
  check digit, license plate per country, bicycle. Airline: airline, airport (IATA,
  ICAO, tied to the geo tree), aircraft, flight number, seat, record locator.
- Science and nature: element (symbol, number), unit, scientist, planet, animal by
  class, plant.
- Food and drink, book (title, author, genre, publisher), music (genre, instrument),
  movie genre, sport, programming language, OS, database (column, type, engine),
  version control (branch, commit message, sha), semver, file (name, extension, path).
- Color: named per locale, hex, rgb, hsl, cmyk, css function.
- Locales: nb_NO, da_DK, fi_FI, de_DE, en_GB, nl_NL, fr_FR, es_ES with person, address,
  phone, national id, company and date names each.

### Release

- CLI without Go — investigate prebuilt binaries: GoReleaser attaching them to
  the Gitea release the tag workflow publishes, a container image, Homebrew and
  Scoop. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag.
- Homepage — a simple page for fejkdata with an in-browser generator: the library
  compiled to WebAssembly, so visitors generate as much data as they like in their
  own browser.
