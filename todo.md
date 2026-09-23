# Plan

Ordered as the releases that ship it.

## v0.1.0

### Comprehension, until the panel scores 7.0

Nothing else ships while the score is under 7.0 — no feature, no category, no data — bar
what `AGENTS.md` excepts, per README goal 2. This section is one
round, replayed: the `comprehension-panel` skill's nine-seat findings run at depth 1,
a PR per item it names, then the run again, until the score passes. The run is the
nine-seat one: the round is meant to buy architectural change, and the four-seat
scoring run names too little to steer one. A fresh run replaces the items below.

Scores, newest last: 5.9 on 2026-09-20 (Navigation 7.0, Locality 5.2, Shape 5.9,
Self-sufficiency 5.7); 5.8 on 2026-09-22 (Navigation 6.7, Locality 5.0, Shape 5.9,
Self-sufficiency 5.6), every seat capped by Locality or Self-sufficiency. Eight of nine
named `drawWalk.walk` and `rowSet` or `checkFamilies` hardest and least wanted to modify:
a load-time copy of the pinning `hold.pin` and `hold.rowOf` do at render, tied to them
by nothing but tests, whose premise lives only in the README's Decisions.

This round, in order:

- Move the README's Decisions section to `docs/decisions.md`, leaving a one-line index
  of the titles in `AGENTS.md`, and drop the `AGENTS.md` line pointing decisions at the
  README. Give each fence a bare link to the entry it enforces: every seat needed
  "The rows of a table are alternatives", "The expansion hold and the render's draws
  are two fences" or "A table read into is pinned" and found it by grepping 470 lines.
  Carry the timezone weight's premise in with it: weighting by GeoNames city
  population moved 300 seeded draws of `misc.territory[US].timezone` from 44 landing on
  the four zones most Americans live in to 278, and dropped `America/Indiana/Petersburg`
  (pop. 2,400) from 17 draws to 0. Carry in the decision `linkTemplateRefs`'s doc holds
  today, which no entry records: a category is a unit, so a reference back into it
  describes a draw other than the fields beside it, which is what goal 3 asks for.
- Keep reference bindings out of `template.fields`, so no walk over a template's fields
  needs the `isRef` filter `named`, `paths`, `recordColumns` and `template.field` carry.
- Compile a template's format once: `compileFormat` runs again from `linkTemplateRefs`,
  so `ops`, `bound` and `held` are provisional until linking and its fences run twice.
- Parse a format once and have the fences read `t.ops`: `eachToken` rescans the format
  in `fieldTokens`, `operandTokens`, `boundReaders`, `renderEdges`, `loneRef`,
  `operandReader`, `checkNoRepeatedRead` and `compileOps`, and function-versus-field is
  decided by `indexOutside(body, '(')` in two of them and by `funcCall` in the rest.
- Name `pathWalk`'s modes: render, probe, check and cover are chosen today by which
  callbacks are nil and whether its session is, and `walkChoice` returns `nil, nil` where
  none is set. The mode may live in `pathWalk`; the session may not, since beside `pins` it
  moves the hold maps to the heap.
- Register a transform once: `transforms` restates the `lowercase`, `uppercase` and
  `ascii` entries in `builtins`, and `valueProof.template` classifies by the first while
  render calls the second. Derive the numeric builtins `valueProof.template` lists in
  its error text from `builtin.number` while there.
- Split `drawCheck.reads`'s six-term condition into named predicates.
- Rename what names two things: `template.go` holds the token grammar while `template`
  is `node.go`'s type and `Template` `inline.go`'s; `kind` `'b'` is a brace body in
  `ftoken` and a builtin in `op`; `hold`'s receiver is `d`; `drawAt.alt` holds the rows
  a walk entered.
- Run the nine-seat panel again and file what it names here.

### Data

Shape: T = table, t = template, c = choice. Read every factual list from a register
by a `data-import/` script, as README goal 12 asks.

- Place xlsx cells by their `r` reference in `data-import/xlsx.py`: a sheet omitting an
  empty cell shifts every later column left.
- Spell `misc.creditcard`'s digit runs `{digits(n)}`, and refuse a repeat over a lone
  `{digits(1)}` at `New`, naming that spelling: both render the same run.
- Add the remaining `misc` tables and templates, one row of the table below per chunk.
- Read `misc.emoji` from the Unicode `emoji-test.txt` register, and `misc.car` from an
  international make and model source; vPIC and Mobility Sweden are national, so they
  build `en_US.car` and `sv_SE.car`.
- Add the remaining locale categories: company, phone, finance, vehicle, words.
- Give every shipped category building-block columns, `misc.uuid` → `variant` among
  them, and one record shape across the national ids; re-pin the shape.
- Give the address records one column set across countries: `region` and
  `municipality` as columns on `geo.SE.address` too.
- Fill the 398 Swedish localities weighted 200 from SCB småorter.
- Give `url` and `email` a path that draws only domains nobody can register, keeping
  the wide set as the default, per goal 14: 17 of the 40 distinct ones shipped today
  sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's
  `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost`
  provably reach nothing.
- Audit the rest of the shipped set against goal 14 and give each a never-reaching
  path where it lacks one: `phone` draws live PTS and NANP ranges — PTS's five
  fiction series are the Swedish inert set, sourced in
  [research-sources-se.md](docs/research/research-sources-se.md), and the NANP one
  still wants a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.
- Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does
  not hold: ISO 3166 reserves it for Western Sahara and the root zone has never been
  delegated it, so no resolver answers for it and goal 3 promises otherwise.
- Draw `en_US.phone`'s `exch` as a NANP central office code: `{int(100,999)}` renders
  a leading 1 in about an eighth of draws, which libphonenumber rejects, and
  `TestShippedUSPhone` proves the shape rather than the rule. Assert the rule with it.
- Give a `sv_SE.personnummer` over 100 the `+` separator Skatteverket spells, or stop
  drawing birthdates that reach 100: the format hard-codes `-`, and the 1930 floor
  makes the oldest draws invalid from 2031.
- Settle `en_US.ip` and `sv_SE.ip`, today byte-identical: either one `misc.ip` as the
  decision "`misc` is what every locale shares" asks, or a decision saying why an
  address that carries no locale stays per-locale.
- Accept a middle name, and draw a shipped `personnummer` inside a *selected* sex;
  both want a draw group sharing its family's pins. Stop the conflict error naming a
  rewrite that returns a different value where the read it conflicts with sits inside
  another category, and report a struct column's draw conflict with the path spelling
  a tag takes.
- Let `geo.SE.locality[Lund].address` descend from a selected row into the template
  beside the family: require the path step to reach a sibling category, and seed every
  draw group of the render from the outer selector's pins.
- Read `email.local`'s and `username`'s handles from the shipped name tables, which
  goal 12 asks for and the hand-written list they use today does not meet. Keep the
  handle shape: `username` shortens a surname to `ahl` or `sjo`, and the US one draws
  unisex given names, neither of which falls out of the tables unaided.
- Decide whether `email.local` and `username` share one list: a reference between
  shipped categories breaks a consumer once tagged, so it rides a 0.(x+1).0 after that.

`misc` (locale-neutral)

| Category | Shape | Source | Licence |
|---|---|---|---|
| `ip` v4 documentation ranges, private, CIDR; `ipv6` `2001:db8::/32`; `mac` locally administered | t | RFC 5737, 1918, 9637 | facts |
| `creditcard` per network: IIN, length, CVV, expiry | T+t | network rules, `{luhn()}` | facts |
| `bic`, `isin`, `cusip` | t | structure rules | facts |
| `ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin`, `sku` | t | check-digit rules, GS1 demo prefix 952 | facts |
| `product` name, adjective, material, department, `garmentsize` | T/c | LLM-written lists | — |
| `airport` IATA, ICAO, territory, municipality; `airline`; `aircraft`; `flight`, `seat`, `pnr` | T+t | OurAirports, Wikidata | public domain, CC0 |
| `element`, `unit`, `planet` | T | PubChem, BIPM, NASA | public domain, CC BY 4.0 |
| `animal` by class, `plant` | T | Wikidata | CC0 |
| `bloodtype` weighted (global) | c | AABB | facts |
| `color` named, hex, rgb, hsl, cmyk | T+t | CSS Color 4, XKCD | W3C, CC0 |
| `programminglanguage`, `os`, `browser`, `database` column/type/engine, `fileext` | T/c | Linguist, curated | MIT |
| `commit` message, branch, `sha` | t | LLM-written, `{hex(40)}` | — |
| `password`, `hash`, `jwt`, `slug`, `hostname`, `avatar`, `placeholderimage` | t | shapes | — |
| `book` title, author; `genre`; `instrument`; `dish`, `ingredient` | T | Gutenberg catalog, MusicBrainz, USDA | CC0 |
| `lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword`, `quote` | T/c | lorem ipsum, LLM-written | — |
| `direction`, `continent`, `ulid` | c/t | — | — |

`sv_SE` ([source research](docs/research/research-sources-se.md))

| Category | Shape | Source | Licence |
|---|---|---|---|
| `person` title, birthdate, age, blood type weighted | t | geblod.nu distribution | facts |
| `organisationsnummer` by form, `vat` | t | Bolagsverket group digits, Luhn, `SE…01` | facts |
| `company` name patterns, legal form weighted | t | Bolagsverket registrations 2025 | CC BY 2.5 SE |
| `sni` industry codes | T | SCB SNI 2025 | "Källa: SCB" |
| `ssyk` occupations | T | SCB SSYK 2012 | "Källa: SCB" |
| `phone` mobile, fixed per area code, E.164 | T+t | PTS numbering plan 2024 | facts |
| `bankgiro`, `plusgiro`, `bankaccount` clearing + Luhn/mod-11 | T+t | Bankinfrastruktur CSV, Bankgirot rules | facts |
| `licenseplate` `ABC 123`, `ABC 12A`, blocked combinations | t | Transportstyrelsen | facts |
| `car` make, model weighted by registrations | T | Mobility Sweden 2025 | cite |
| `word` noun, verb, adjective, adverb; `sentence` | T+t | SALDO | CC BY 4.0 |
| month and weekday names, `holiday` | t+T | CLDR sv, lag 1989:253 | Unicode |
| `territory` localised names keyed by alpha2 | T | CLDR sv territory names | Unicode |

`en_US` ([source research](docs/research/research-sources-world.md), and `part1`–`part4`)

| Category | Shape | Source | Licence |
|---|---|---|---|
| `ein` valid ranges | t | IRS | facts |
| `company` names, suffix, `naics` | t+T | SEC tickers for patterns, NAICS 2022 | public domain |
| `phone` NANP with valid NPA/NXX | t | NANPA rules | facts |
| `routing` ABA with check, `bankaccount` | t | Fed prefix ranges | facts |
| `licenseplate` per state | T | hand-authored patterns | facts |
| `car` make, model | T | NHTSA vPIC | public domain |
| `word`, `sentence`, `paragraph` | T+t | Moby POS or WordNet | public domain / WordNet |
| month and weekday names, `holiday` | t+T | CLDR en | Unicode |

### Builtins the data cannot express

- Add `{isin()}` (Luhn over letters expanded to digits), `{cusip()}`, `{aba()}`
  (3-7-1 weights) and `{vin()}` (position 9 over the whole; a sample taking the WMI,
  since the check sits mid-string).
- Add `{base64url(n)}` for JWT shapes.
- Add a `unix` layout to `date()` once something needs it.

### Library and CLI

- Report the same error every load for a table with two bad options:
  `readTableOptions` returns on the first in Go's map order.
- Refuse two whole reads of one table family in one render, which goal 5 promises is a
  load error: `{/sel}|{/sel}` panics out of `Fake` with "a fence should have refused this
  at New" where the two draws land on rows whose cells select different rows of another
  table, and `{/geo.SE.locality}|{/geo.SE.municipality}` renders a locality outside the
  municipality beside it. Each read draws its own row, so cells of two rows render
  together, where the fence counts them alternatives and compares neither. Have the error
  name the path spelling that holds, `{/geo.SE.locality.name}|{/geo.SE.municipality.name}`,
  as the README's rule on a rejected spelling asks.
- Name the row a refused read sits in: where both reads come from TSV cells the overlap
  error names the category and the two columns but no row, so the author greps a
  register-sized file for the cell — and "name the fields you want instead" rewrites a
  cell that is not the one to fix.
- Have the family fence offer only a rewrite that works: where the drawn table's rows select
  more than one row of the family, every `write {/terr[NO].name}` it offers fails the same
  way, so only the `drawGroup` it also names is left.
- Spell a table one way across the family errors: `checkFamilyPair` names
  `drawn.category`, `pinRow` names `t.path` and `mustRow` panics with `t.category`, so one
  table is `territory` and `misc.territory`, and the short spelling names no file where two
  folders hold that name.
- Name a spelling that works when a row selector misses: `misc.protocol[tcp]`,
  `misc.httpmethod[get]` and `misc.territory[se]` all answer "no row … has key or name"
  and stop there, where a case-insensitive match could name the row that exists, and a
  table with no `name` column could say it selects by key alone. `misc.tld[se]` misses
  on the leading dot its keys carry rather than on case, and `misc.loglevel[Error]`
  misses although a column carries `Error`, so the near miss is worth naming whatever
  shape it takes.
- Let a table column carry a `datatype`, so `--format json` writes
  `"safe": true` and `--format sql` a boolean rather than the text `'true'`. Today only
  a JSON field takes one, so `misc.httpmethod`'s booleans are typed in Go and text
  everywhere else. Typing a shipped column changes what `json` and `sql` write, so the
  capability is a minor and applying it to `misc.httpmethod` is a major.

### Open questions to settle

- Decide whether a goal 5 violation joins the comprehension gate's exception list: the
  `{/sel}|{/sel}` panic waits behind the whole round at 5.9, and
  `{/geo.SE.locality}|{/geo.SE.municipality}` mispairs the shipped data meanwhile, though a
  repair restoring behaviour the repo already claims may already pass the gate as written.
- Decide whether the README's Audience names the author who writes categories under
  `--data-path`: the four personas listed consume the shipped set and the contributor
  ships a register upstream, while "Your own data" is a README section, `-d` is a
  first-class flag, and every load error the draw fences raise lands on that author.
- Decide whether `data/misc`'s 33 flat files gain a level before v1.0.0: a folder is a
  path segment, so `misc/net/tld` is a rename a consumer pays for, and goal 13 promises
  the directory keeps growing.
- Decide whether `misc.browser` becomes a parent of `misc.useragent`, so
  `misc.browser[Chrome].useragent` resolves. Adding a `parent` after v0.1.0 breaks a
  consumer, so it rides a 0.(x+1).0.
- Decide whether `misc.timezone` keeps every zone tzdb names: one added recently,
  `Europe/Kyiv` among them, is rejected by a consumer resolving it against older
  tzdata, and the population weight makes that zone likelier, not rarer. Today the
  README names the action a reader takes instead; narrowing what ships is the
  alternative.
- Decide whether `misc.territory.country` is renamed `sovereign`: it holds a code, and
  `country` collides with the `misc.country` path it replaced. A rename after v0.1.0
  breaks a consumer, so it rides a 0.(x+1).0.
- Decide whether `parent: territory` stays, given a `--data-path` override of
  `misc.territory` now fails `New` unless `misc.timezone` is overridden with it.

### Release

- Read the tag through `git/ref/tags/{tag}` in `publish_release.py`: GitHub answers
  `tags/{tag}` with 404 whatever exists, so the burnt-version guard never fires and a
  release lands on a tag already pointing at another commit.
- Document `NewRecordTemplate` in the README's Library section, the one exported name
  it does not name.
- Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds: nothing has shipped, so
  "no longer paths", "where it used to fail" and "where it used to be an empty string"
  describe versions no reader can have installed.
- Ship prebuilt binaries so the CLI needs no Go: GoReleaser attaching them to the Gitea
  release the tag workflow publishes, a container image, Homebrew and Scoop. A checkout
  build prints `devel` for `--version`; the binaries carry the stamped tag.

## v0.2.0

- Ship the full registers as packs, each a Go module with its own `embed.FS` and a zip
  for `--data-path`: every US place of 10,000 and more streets per locality, built by
  `--min-population` and `--streets-per-locality`. The default embed stays under ~1 MB
  per country and `New` under ~50 ms.
- Add the locales nb_NO, da_DK, fi_FI, de_DE, en_GB, nl_NL, fr_FR and es_ES, with
  person, address, phone, national id, company and date names each, and the `geo/`
  trees NO, DK, FI, NL, FR, AU, CA, ES, GB and DE their `address` reads by reference;
  AU and CA get a tree with no locale of their own.
- Add `{btc()}` and `{eth()}`, which need sha256 and keccak for Base58Check and EIP-55.
- Pair a street with its exact postnummer, which needs an application to Lantmäteriet;
  today a street goes to the nearest postal code centroid. It rewrites shipped rows, so
  it is a major once 1.0 is cut and a minor before that.

## v0.3.0

Names no items yet, so planning it is a chunk of its own: it takes the additive work
that lands after v0.2.0, and it has to add something, since a minor only adds. Its job
is to be the release with no breaking change that v1.0.0 waits on — v0.2.0 cannot be,
because pairing a street with its exact postnummer rewrites shipped rows.

## v1.0.0

- Cut it once the shipped data is in its record shape and one full minor has shipped
  with no breaking change, per the README's Versioning table: v0.1.0 settles the
  record shape and v0.3.0 is the quiet minor.
- Announce it where a developer choosing a fake-data tool already reads, which is
  goal 1's second half and waits for the tag: complete, stable, installable without
  Go. Write the README's first screen for someone deciding in a minute.

## Not release-bound

- Hold every change the charter says owes a `CHANGELOG.md` entry to one in CI: the
  check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a
  flag, an exit code, an exported name, a fence, a builtin and the lowest Go.
- Measure the allocation gate over the shipped `geo` trees too: `pinSet.pins [8]` is sized
  for them, and `perf_test.go` checks a synthetic five-deep tree only.
- Make a named draw group's hold lazy, `&hold{}` in `renderScope.hold`: it escapes
  to the heap anyway, so its two eager maps buy no stack and cost two allocations per
  group that may never read through them. Check it against the alloc tests.
- Name goal 2's four criteria with the axes the panel scores, so a recorded score maps
  back to the clause it came from: the goal spells them out in prose while `todo.md`
  records `Navigation`, `Locality`, `Shape` and `Self-sufficiency`.
- Put the round's nine-seat score and the four-seat score every later pull request
  takes on one scale, or say in `AGENTS.md` that they are two ratchets: it gates the
  round on the first and every merge after 7.0 on the second, then holds one "last
  score" across both, where the `comprehension-panel` skill has two independent panels
  drift unless a calibration seat reads the anchor project without its score.
- Delete `doc.go`'s "The draw fences are …" sentence: it names a function, a type and
  a helper reached only through `drawWalk.check`, and misses `checkDrawGroup`,
  `checkOwnFamily` and `checkColumnDraws`; the README's Layout indexes the fences by file.
- Cut `AGENTS.md` to the rules only it states: the round procedure stands word for word
  in `todo.md` and the `gh` rule opens with the README's GitHub decision, so each is a
  second copy that drifts. Split the comprehension rule while there, which packs the
  gate, its exceptions, the round and the at-or-above-7.0 regime into one line.
- Scope `AGENTS.md`'s `Hard tabs.` to the Go source, or drop it: `gofmt` already gates
  Go at `Dockerfile:21`, and the shipped JSON under `data/` is two-space, so the rule as
  written is one no Go file can break and every data file does.
- Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub or retire it:
  Gitea converts a repository to a mirror only by re-creating it, so until that runs
  the copy there is a second owner of one history and goes stale from this commit.
- Stop `/cache` collecting root-owned files, which today every command documented
  without `--user` leaves for the next `--user` one to trip over, and which the README
  answers with a chown a reader has to repeat. Put `--user` on all of them, or give the
  `x-go` anchor a `user:` — that one needs a mechanism, since Compose interpolates from
  the process environment and neither `UID` nor `GID` is exported there, so
  `user: "${UID}:${GID}"` resolves to `":"`.
- Spell the latest supported Go once: `Dockerfile`'s `ARG GO_VERSION` and
  `compose.yaml`'s `golang:${GO_VERSION:-…}` pin one version twice, and a bump that
  moved only the first left `docker compose run --rm fmt` formatting to a `gofmt` the
  gate rejects. A `renovate.json` custom manager reads the second today; drop it with
  the duplication.
- Publish a homepage with an in-browser generator: the library compiled to WebAssembly,
  so visitors generate as much data as they like in their own browser.
