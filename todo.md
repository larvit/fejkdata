# todo

## Scoring

`Score = -R - S/4 + 2*A + 2*G*W`

`Bar = 9`

`Next ID = 76`

| Goal | W |
|---|---|
| 1 | 1.00 |
| 1.1 | 1.00 |
| 2 | 0.95 |
| 2.1 | 0.95 |
| 2.2 | 0.95 |
| 3 | 0.90 |
| 3.1 | 0.90 |
| 3.2 | 0.90 |
| 3.3 | 0.90 |
| 3.4 | 0.90 |
| 4 | 0.85 |
| 4.1 | 0.85 |
| 5 | 0.80 |
| 5.1 | 0.80 |
| 5.2 | 0.80 |
| 5.3 | 0.80 |
| 5.4 | 0.80 |
| 5.5 | 0.80 |
| 5.6 | 0.80 |
| 5.7 | 0.80 |
| 5.8 | 0.80 |
| 6 | 0.75 |
| 7 | 0.70 |
| 7.1 | 0.70 |
| 7.2 | 0.70 |
| 8 | 0.65 |
| 9 | 0.60 |
| 10 | 0.55 |
| 11 | 0.50 |
| 11.1 | 0.50 |
| 11.2 | 0.50 |
| 11.3 | 0.50 |
| 12 | 0.45 |
| 13 | 0.40 |
| 13.1 | 0.40 |
| 14 | 0.35 |
| 15 | 0.30 |

## Items

| ID | Release | Exempt | Item | R | S | A | G | Goals | Score |
|---|---|---|---|---|---|---|---|---|---|
| 61 | 0.1.0 |  | **Add a locale, with its `geo/` tree, for each of the ten most-spoken languages and each Nordic country.** | 4 | 9 | 9 | 10 | 2.1 | 30.8 |
| 16 | 0.1.0 |  | **Add the remaining locale categories: company, phone, finance, vehicle, words.** | 3 | 8 | 8 | 9 | 2.2, 14 | 28.1 |
| 15 | 0.1.0 |  | **Add the remaining `misc` tables and templates, one row of its detail's table per chunk.** | 3 | 9 | 8 | 8 | 2, 14, 13 | 25.9 |
| 3 | 0.1.0 | decision | **Rewrite the shipped categories so facts that belong together come from one named pick.** | 5 | 5 | 8 | 9 | 4.1 | 25.0 |
| 4 | 0.1.0 | decision | **Draw every `{…}` afresh, keep a pick only by name, and delete `drawGroup` with the fences that held picks implicitly.** | 7 | 8 | 8 | 10 | 5.2, 4.1 | 24.0 |
| 25 | 0.1.0 | defect | **Draw `en_US.phone`'s `exch` as a NANP central office code, and assert the rule in `TestShippedUSPhone`.** | 2 | 2 | 6 | 8 | 4 | 23.1 |
| 57 | 0.1.0 |  | **Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.** | 1 | 2 | 6 | 6 | 1.1 | 22.5 |
| 2 | 0.1.0 | decision | **Step up from a row to the row it links to with `..`: `{/city[Oslo]..country.name}` renders `Norway`, and `city[Oslo].country` stays the cell `NO`.** | 4 | 5 | 7 | 8 | 5.4 | 21.6 |
| 26 | 0.1.0 | defect | **Give a `sv_SE.personnummer` over 100 the `+` separator Skatteverket spells, or stop drawing birthdates that reach 100.** | 2 | 2 | 5 | 8 | 4 | 21.1 |
| 34 | 0.1.0 |  | **Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.** | 2 | 4 | 6 | 7 | 4 | 20.9 |
| 12 | 0.1.0 | defect | **Strip the whitespace `data-import/territory.py` copies from its register.** | 1 | 1 | 5 | 7 | 4 | 20.6 |
| 5 | 0.1.0 |  | **Run the nine-seat comprehension panel after the draw restructure, and file what it names.** | 1 | 3 | 3 | 9 | 3 | 20.4 |
| 71 | 0.1.0 |  | **Accept a name as a `calc` operand, so `{calc(n * 2)}` computes from the pick `{n}` prints.** | 3 | 4 | 5 | 7 | 5.3 | 17.2 |
| 58 | 0.2.0 |  | **Ship prebuilt binaries, a container image, and packages for Homebrew, Scoop and the biggest Linux package managers, so the CLI needs no Go.** | 3 | 7 | 8 | 9 | 1.1, 7 | 29.2 |
| 36 | 0.2.0 |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 5 | 6 | 7 | 8 | 6 | 19.5 |
| 41 | 0.2.0 |  | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 6 | 7 | 7 | 8 | 5.7 | 19.1 |
| 52 | 0.2.0 |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 2 | 5 | 7 | 6 | 19.0 |
| 28 | 0.2.0 |  | **Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.** | 4 | 4 | 6 | 7 | 4.1 | 18.9 |
| 24 | 0.2.0 | defect | **Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.** | 1 | 1 | 4 | 7 | 4 | 18.6 |
| 55 | 0.2.0 |  | **Document `NewRecordTemplate`, `ErrNoColumns`, `ErrNoData` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 6 | 7 | 11 | 17.5 |
| 53 | 0.2.0 |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 4 | 5 | 6 | 6 | 4 | 16.9 |
| 18 | 0.2.0 |  | **Give the address records one column set across countries: `region` and `municipality` as columns on `geo.SE.address` too.** | 4 | 4 | 6 | 6 | 5 | 16.6 |
| 44 | 0.2.0 | decision | **Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.** | 1 | 1 | 4 | 6 | 6 | 15.8 |
| 54 | 0.2.0 | defect | **Read the tag through `git/ref/tags/{tag}` in `publish_release.py`.** | 2 | 1 | 3 | 6 | 1.1 | 15.8 |
| 31 | 0.2.0 |  | **Rename `misc.territory.country` to `sovereign`.** | 3 | 2 | 4 | 7 | 5.8 | 15.7 |
| 20 | 0.2.0 | defect | **Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.** | 2 | 2 | 4 | 6 | 4 | 15.7 |
| 50 | 0.2.0 |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 2 | 4 | 6 | 6 | 15.5 |
| 17 | 0.2.0 |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 5 | 6 | 6 | 6 | 5 | 15.1 |
| 48 | 0.2.0 | defect | **Refuse a `Fake` path into a level carrying a `repeat`, as the README's Correlated fields section states.** | 2 | 2 | 4 | 6 | 5 | 15.1 |
| 22 | 0.2.0 |  | **Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.** | 3 | 4 | 7 | 8 | 15 | 14.8 |
| 29 | 0.2.0 |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 4 | 4 | 5 | 6 | 5 | 14.6 |
| 42 | 0.2.0 |  | **Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.** | 2 | 2 | 4 | 6 | 6 | 14.5 |
| 32 | 0.2.0 |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 3 | 3 | 4 | 6 | 4.1 | 14.4 |
| 27 | 0.2.0 |  | **Merge `en_US.ip` and `sv_SE.ip`, today byte-identical, into one `misc.ip`.** | 3 | 2 | 4 | 6 | 5.7 | 14.1 |
| 46 | 0.2.0 |  | **Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.** | 2 | 2 | 3 | 7 | 6, 9 | 14.0 |
| 73 | 0.2.0 | question | **Decide whether the default embed's partial Swedish and US place tables stand against goal 13.1, or have the goals say which goal wins.** | 1 | 1 | 3 | 6 | 13.1 | 9.6 |
| 74 | 0.2.0 | question | **Decide whether `misc.timezone` may drop the 17 zones of territories `misc.territory` leaves out, against goal 13.1, or have the goals say which goal wins.** | 1 | 1 | 3 | 6 | 13.1 | 9.6 |
| 14 | 0.3.0 |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`, and refuse a repeat over a lone `{digits(1)}` at `New`, naming that spelling.** | 3 | 3 | 4 | 6 | 5.7 | 13.9 |
| 38 | 0.3.0 | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 45 | 0.3.0 |  | **Name the node a selector follows in `stepInto`'s refusal.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 49 | 0.3.0 |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 13 | 0.3.0 | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 2 | 2 | 3 | 6 | 4 | 13.7 |
| 39 | 0.3.0 |  | **Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 2 | 3 | 6 | 6 | 13.5 |
| 51 | 0.3.0 |  | **Spell a table one way across the errors that name it.** | 1 | 2 | 3 | 6 | 6 | 13.5 |
| 10 | 0.3.0 |  | **Test that every node kind reaches each switch over node kinds.** | 1 | 3 | 2 | 6 | 3.2 | 13.1 |
| 11 | 0.3.0 |  | **Test that every `geo` tree holds the five table names and the `address` columns that port across countries.** | 1 | 2 | 3 | 5 | 4 | 13.0 |
| 35 | 0.3.0 |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 2 | 4 | 5 | 8 | 13.0 |
| 21 | 0.3.0 |  | **List `title` in both locales, and `sv_SE`'s `birth-number`, in the README's Data list of what each locale carries.** | 1 | 1 | 4 | 6 | 11 | 12.8 |
| 75 | 0.3.0 |  | **Run every README `sh` example as a test, and compare each "Renders" line to a seeded render.** | 2 | 5 | 4 | 7 | 11.2 | 11.8 |
| 23 | 0.3.0 |  | **Give every other shipped category that could reach something real a path that never does.** | 3 | 6 | 6 | 7 | 15 | 11.7 |
| 8 | 0.3.0 |  | **Move `binding`, `bind` and `checkNodeFences` out of `data.go` into a file of their own.** | 1 | 2 | 2 | 5 | 3.1 | 11.5 |
| 9 | 0.3.0 |  | **Move `table.route`, `selector`, `step` and `drawStep` from `path.go` to `table.go`.** | 1 | 2 | 2 | 5 | 3.1 | 11.5 |
| 30 | 0.3.0 |  | **Draw `email.local` from `username`, so the two share one handle list.** | 2 | 2 | 3 | 5 | 5.7 | 11.5 |
| 40 | 0.3.0 |  | **Parse `defaultTable`'s path with the library's grammar.** | 2 | 2 | 3 | 5 | 5.7 | 11.5 |
| 33 | 0.3.0 |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 4 | 5 | 4 | 5 | 5 | 10.8 |
| 47 | 0.3.0 | defect | **Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.** | 1 | 1 | 2 | 5 | 6 | 10.2 |
| 72 | 0.3.0 | principle | **Fail the merge gate when test coverage falls below the last recorded figure.** | 2 | 3 | 2 | 5 | 3 | 10.2 |
| 7 | 0.3.0 |  | **Fill `arm` in one place.** | 4 | 4 | 2 | 6 | 3.2 | 9.8 |
| 19 | 0.3.0 |  | **Fill the 398 Swedish localities weighted 200 from SCB småorter.** | 2 | 4 | 4 | 6 | 13.1 | 9.8 |
| 37 | 0.3.0 |  | **Rename `newRand` to `newGeneratorState` and `Generator.rand` to `state`.** | 1 | 1 | 1 | 5 | 3.3 | 9.8 |
| 43 | 0.3.0 |  | **Call `templateError` a compile failure in its doc.** | 1 | 1 | 1 | 5 | 3.3 | 9.8 |
| 56 | 0.3.0 | defect | **Build on a manual run of `test.yml`.** | 1 | 1 | 1 | 4 | 1 | 8.8 |
| 63 | 0.4.0 |  | **Pair a street with its exact postnummer.** | 3 | 6 | 5 | 8 | 4.1 | 19.1 |
| 62 | 0.4.0 |  | **Add `{btc()}` and `{eth()}`.** | 2 | 4 | 4 | 6 | 4 | 15.2 |
| 59 | 0.4.0 |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 1 | 5 | 5 | 11 | 13.8 |
| 65 | 0.4.0 |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 3 | 4 | 5 | 5.8 | 13.2 |
| 60 | 0.4.0 |  | **Ship the full registers as packs, each a Go module with its own `embed.FS` and a zip for `--data-path`.** | 5 | 8 | 7 | 6 | 14, 12 | 12.4 |
| 66 | 0.4.0 |  | **Cut the README's Layout block to the lines that say what a file name cannot.** | 1 | 1 | 2 | 5 | 3.4 | 11.8 |
| 67 | 0.4.0 |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 6 | 3.4 | 11.3 |
| 64 | 0.4.0 |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 4 | 1 | 6 | 3.4 | 10.8 |
| 68 | 0.4.0 | decision | **Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.** | 2 | 1 | 1 | 3 | 1.1 | 5.8 |
| 69 | 1.0.0 |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 3 | 7 | 9 | 1.1 | 30.2 |
| 70 | 1.0.0 |  | **Publish a homepage with an in-browser generator, the library compiled to WebAssembly.** | 3 | 6 | 6 | 6 | 1.1 | 19.5 |

## Details

### 61. Add a locale, with its `geo/` tree, for each of the ten most-spoken languages and each Nordic country.

Pick the ten from a published ranking of languages by total speakers, such as Ethnologue's, each in the country where it has the most speakers, and record the ranking in the decision that names them. The Nordic ones are sv_SE (shipped), nb_NO, da_DK, fi_FI and is_IS. Each locale carries goal 2.2's categories, and its `address` reads its country's `geo/` tree by reference.

### 16. Add the remaining locale categories: company, phone, finance, vehicle, words.

Shape: T = table, t = template, c = choice.

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

### 15. Add the remaining `misc` tables and templates, one row of its detail's table per chunk.

Shape: T = table, t = template, c = choice.

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

### 3. Rewrite the shipped categories so facts that belong together come from one named pick.

- `sv_SE.person` and `en_US.person` bind the first name and read sex and title through it: `"first": "{.first-name as n}{n.name}"`, `"sex": "{n..sex.name}"`, a title from `{n..sex.title.name}`.
- `geo.SE.address` and `geo.US.address` bind the street and read locality, postal code and region through it; the locale `address` categories bind `/geo.XX.address` once. They print it whole as `{a}`, so the country record's format appears once; revise the decision "A locale's `address` restates its country record's format".
- Needs item 2. Every shipped value must still pass its consumer's check before and after item 4.
- Re-pin seeded output in its own commit.

### 4. Draw every `{…}` afresh, keep a pick only by name, and delete `drawGroup` with the fences that held picks implicitly.

- Needs items 2, 3 and 71. After it, `{/person} ({/person.first})` renders two people, as does `{/sv_SE.person as a}{/sv_SE.person as b}{a} & {b}`. `{w} {uppercase(w)}` renders `b G` where `w` is a choice of letters, and `{/city.name}, {/country[SE].name}` and `{/country[SE].name} / {/country[NO].name}` both load.
- Deletes the token-order rules (`drawsApart`, `readFold`, `branches`, the pairwise replay), the expansion hold fence, `drawGroup`, the own-family fence and the cell-reference fences, and revises the decisions that point here. The remaining draw checks are names, paths through names, cycles and the repeat cap.
- Fixes on the way: `{/sel}|{/sel}` panics out of `Fake` today where two draws land on rows whose cells select different rows of another table; it must render.
- Rewrites the README's References, Draw group, Correlated fields, Linked tables, Names and "One draw, one spelling" sections; the Linked tables section says which pick's weights govern a family, since the bound row decides.
- Widens the refusal of a name read once to every binding, a path into it or a reference path bound included, since `{n.x}` read once is then `{/ref.x}`, and deletes the names fence over draw groups and nested repeats (`checkUses`, `readsHeld`).
- Re-pin seeded output and `testdata/shipped_shape.txt` in their own commits; the CHANGELOG names the grammar change.

### 25. Draw `en_US.phone`'s `exch` as a NANP central office code, and assert the rule in `TestShippedUSPhone`.

`{int(100,999)}` renders a leading 1 in about an eighth of draws, which libphonenumber rejects, and the test proves the shape rather than the rule.

### 57. Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.

Nothing has shipped, so "no longer paths", "where it used to fail" and "where it used to be an empty string" describe versions no reader can have installed.

### 2. Step up from a row to the row it links to with `..`: `{/city[Oslo]..country.name}` renders `Norway`, and `city[Oslo].country` stays the cell `NO`.

- `..` goes up one level, to the parent table, and the name after it is that parent's: `{c..country.name}`, and from a locality `{l..municipality..region.name}`. Naming any other table is a load error naming the parent.
- It steps up from a selected row, a drawn row or a name bound to a row.
- After `..`, `.` steps down again: `{c..country.city.name}` is a city of `c`'s country, drawn afresh.
- `.country` after a row keeps reading the link column's cell, so the README's `city[Oslo].country` example stays true.
- A `Fake` path and `--list` take `..` too; `List` keeps advertising direct descents only.

### 26. Give a `sv_SE.personnummer` over 100 the `+` separator Skatteverket spells, or stop drawing birthdates that reach 100.

The format hard-codes `-`, and the 1930 floor makes the oldest draws invalid from 2031.

### 34. Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.

`isin`: Luhn over letters expanded to digits. `aba`: 3-7-1 weights. `vin`: position 9 over the whole; a sample taking the WMI, since the check sits mid-string.

### 12. Strip the whitespace `data-import/territory.py` copies from its register.

`misc.territory[CW].capital` renders ` Willemstad`, with a leading space.

### 5. Run the nine-seat comprehension panel after the draw restructure, and file what it names.

Needs item 4.

### 71. Accept a name as a `calc` operand, so `{calc(n * 2)}` computes from the pick `{n}` prints.

`checkCalc` and the value proof read an operand as a sibling field, so `{calc(n * 2)}` is refused with `no field "n"` today. Item 4 stops an operand holding its field for the expansion, after which a name is the one way to show the operand a calc computes from, as the README's Computation example does.

### 58. Ship prebuilt binaries, a container image, and packages for Homebrew, Scoop and the biggest Linux package managers, so the CLI needs no Go.

GoReleaser attaches the binaries to the release the tag workflow publishes, and builds deb and rpm packages, an Alpine APK and an AUR package beside Homebrew and Scoop. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag. v0.1.0 ships first, so it can be tested before it is packaged.

### 36. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 41. Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.

`geo.SE.municipality[Lund]` and `misc.territory[Sweden]` select a row today. Steps:
- Have a selector that misses name the key of every row whose cell spells it, matched without case, else say the table selects by its key: `misc.territory[Sweden]` and `misc.territory[se]` name `[SE]`, `misc.protocol[tcp]` its key, and `misc.loglevel[Error]` `[err]`.
- Delete the `name` option and the load checks it carries, and have `New` refuse a table still carrying `name`, naming `key`.
- Key `misc.loglevel` on `keyword`, and `geo.SE.locality` on the integer `data-import/geo-se.py` assigns as the decision states, linking `street` and `postal-code` by it.
- Drop `key` from `last-name` and `sv_SE.title`, and re-pin `testdata/shipped_shape.txt`.
- Leave no selection by name in README.md, CHANGELOG.md, docs/ or `--help`.

### 52. Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.

With several `--data-path` layers the author has to work out which directory won, while a parse error on the same file names it.

### 28. Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.

Needs item 4.

### 24. Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.

ISO 3166 reserves it for Western Sahara and the root zone has never been delegated it, so no resolver answers for it.

### 53. Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.

Today only a JSON field takes one, so `misc.httpmethod`'s booleans are typed in Go and text everywhere else. Typing a shipped column changes what `json` and `sql` write, so the capability is a minor and applying it to `misc.httpmethod` is a major.

### 44. Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.

The decision "`misc.tld` keys carry the leading dot, where other tables key on a bare code" leaves `misc.tld[se]` missing today.

### 54. Read the tag through `git/ref/tags/{tag}` in `publish_release.py`.

GitHub answers `tags/{tag}` with 404 whatever exists, so the burnt-version guard never fires and a release lands on a tag already pointing at another commit.

### 31. Rename `misc.territory.country` to `sovereign`.

It holds a code, and `country` collides with the `misc.country` path it replaced.

### 20. Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.

Neither names a `weight`, so a bare `sex` draws evenly.

### 50. Reword `prove`'s refusal of a typed column reading a row, `{/region}`.

Say the format is the table's own (`region's format "{name}"`), name the read to write, `{/region.<column>}` with its columns, and stop calling a one-column format, `{code}`, composed text.

### 17. Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.

`misc.uuid` gains `variant`, for one.

### 48. Refuse a `Fake` path into a level carrying a `repeat`, as the README's Correlated fields section states.

`probePath` skips `pathCheck.enter`, so `Fake("x.a")` under `"repeat":3` renders one draw.

### 22. Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.

17 of the 40 distinct ones shipped today sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost` provably reach nothing.

### 29. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 42. Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.

`datatype: boolean` over `{int(1,2)}` on an `int` field answers `prints an integer, not a boolean`.

### 46. Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.

`readTableOptions` and `loadDir` return on the first in Go's map order.

### 73. Decide whether the default embed's partial Swedish and US place tables stand against goal 13.1, or have the goals say which goal wins.

The decision "The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more" serves goal 2's built-in data with no download, ships less than its registers hold, and names no item that ends it. Item 60 ships the full registers as packs, beside the partial default.

### 74. Decide whether `misc.timezone` may drop the 17 zones of territories `misc.territory` leaves out, against goal 13.1, or have the goals say which goal wins.

The decision "`misc.territory` is the spine, and a `misc` table naming a territory links to it" serves goal 4.1 and drops those tzdb rows, and names no item that ends it.

### 14. Spell `misc.creditcard`'s digit runs `{digits(n)}`, and refuse a repeat over a lone `{digits(1)}` at `New`, naming that spelling.

Both render the same run.

### 38. Compare `calcParser.binary`'s operator as a rune.

`byte(p.rs[p.pos])` reads U+012B `ī` as `+`, so `{calc(a ī b)}` compiles as `a + b`.

### 45. Name the node a selector follows in `stepInto`'s refusal.

`sv_SE.person[1].first` answers `1 is not a table`.

### 49. Report a CRLF rows file holding only its header as having no rows.

`parseRows` answers "line 2 is empty".

### 13. Place xlsx cells by their `r` reference in `data-import/xlsx.py`.

A sheet omitting an empty cell shifts every later column left.

### 39. Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.

An `int64` or `float64` field unproven in range is told to become itself, and a `uint64` one to become `int64`.

### 51. Spell a table one way across the errors that name it.

`pinRow` names `t.path` and `mustRow` panics with `t.category`, so one table is `territory` and `misc.territory`, and the short spelling names no file where two folders hold that name.

### 10. Test that every node kind reaches each switch over node kinds.

Needs item 4; recount the switches then. Today ten, from `render` to `columnItems`, must agree, and only `render` and `renderEdges` say so.

### 11. Test that every `geo` tree holds the five table names and the `address` columns that port across countries.

Nothing checks it, and a new country breaks it silently.

### 75. Run every README `sh` example as a test, and compare each "Renders" line to a seeded render.

`readme_test.go` loads and renders each `json` block, but runs no `sh` example (CLI, Records, Table, Linked tables) and compares no "Renders e.g." output, so goal 11.2 is not met.

### 23. Give every other shipped category that could reach something real a path that never does.

`phone` draws live PTS and NANP ranges — PTS's five fiction series are the Swedish inert set, sourced in [research-sources-se.md](docs/research/research-sources-se.md), and the NANP one still wants a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.

### 8. Move `binding`, `bind` and `checkNodeFences` out of `data.go` into a file of their own.

The greenfield architect found the fence pipeline last. Needs item 4, which shrinks the pipeline.

### 9. Move `table.route`, `selector`, `step` and `drawStep` from `path.go` to `table.go`.

The inherited architect's 3am trace crossed four files for one draw.

### 40. Parse `defaultTable`'s path with the library's grammar.

The CLI counts brackets of its own, so a change to the selector grammar desyncs the default `--table` silently.

### 47. Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.

`loadDir` indexes hidden TSVs and skips them in two places, so one loads where its comment says a hidden file is never data.

### 72. Fail the merge gate when test coverage falls below the last recorded figure.

`technical-principles.md` holds that test coverage should not decline. `docker compose run --rm cover` prints the figure, and neither the `Dockerfile` gate nor `.github/workflows/test.yml` checks it.

### 7. Fill `arm` in one place.

`splitArm` runs before the link and again after it, and `compileArm` finishes it from `reference.go`, so five seats traced three phases. Needs item 4.

### 37. Rename `newRand` to `newGeneratorState` and `Generator.rand` to `state`.

Both name half of a `generatorState`, hiding its `{seq()}` counters, and `New`'s local `rng` shadows the `rng` interface.

### 43. Call `templateError` a compile failure in its doc.

Its doc says render failure, and it wraps what `NewTemplate` and `NewRecordTemplate` refuse.

### 56. Build on a manual run of `test.yml`.

`workflow_dispatch` leaves `github.event.before` empty, so the diff compares `HEAD` with itself and skips `docker build`.

### 63. Pair a street with its exact postnummer.

Needs an application to Lantmäteriet; today a street goes to the nearest postal code centroid. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 62. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 59. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 65. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go.

### 60. Ship the full registers as packs, each a Go module with its own `embed.FS` and a zip for `--data-path`.

Every US place of 10,000 and more streets per locality, built by `--min-population` and `--streets-per-locality`. The default embed stays under ~1 MB per country and `New` under ~50 ms.

### 66. Cut the README's Layout block to the lines that say what a file name cannot.

`calc.go`, `doc.go` and `cmd/fejkdata/` restate their names.

### 67. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens by restating the decision "GitHub is canonical, and the module path names it", a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 64. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 68. Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.

Gitea converts a repository to a mirror only by re-creating it, so until that runs the copy there is a second owner of one history.

### 69. Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.

The human cuts v1.0.0 once the shipped data is in its record shape and one full minor has shipped with no breaking change, per the README's Versioning table; pairing a street with its exact postnummer rewrites shipped rows, so a minor with no breaking change follows it first.
