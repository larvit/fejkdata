# todo

## Scoring

`Score = -R - S/4 + 2*A + 2*G*W`

`Bar = 9`

`Next ID = 64`

| Goal | W |
|---|---|
| 1 | 1.00 |
| 1.1 | 0.97 |
| 2 | 0.94 |
| 2.1 | 0.91 |
| 2.2 | 0.88 |
| 2.3 | 0.85 |
| 2.4 | 0.82 |
| 3 | 0.79 |
| 3.1 | 0.76 |
| 4 | 0.74 |
| 4.1 | 0.71 |
| 4.2 | 0.68 |
| 4.3 | 0.65 |
| 4.4 | 0.62 |
| 4.5 | 0.59 |
| 4.6 | 0.56 |
| 4.7 | 0.53 |
| 4.8 | 0.50 |
| 5 | 0.47 |
| 6 | 0.44 |
| 6.1 | 0.41 |
| 6.2 | 0.38 |
| 7 | 0.35 |
| 8 | 0.32 |
| 9 | 0.29 |
| 10 | 0.26 |
| 10.1 | 0.24 |
| 10.2 | 0.21 |
| 10.3 | 0.18 |
| 11 | 0.15 |
| 12 | 0.12 |
| 12.1 | 0.09 |
| 13 | 0.06 |
| 14 | 0.03 |

## Items

| ID | Release | Exempt | Item | R | S | A | G | Goals | Score |
|---|---|---|---|---|---|---|---|---|---|
| 53 | 0.1.0 |  | **Ship prebuilt binaries, a container image, Homebrew and Scoop, so the CLI needs no Go.** | 3 | 6 | 8 | 9 | 1.1, 6 | 29.0 |
| 1 | 0.1.0 |  | **Bind a pick to a name with `{x as n}`, read it with `{n}` and `{n.path}`, and refuse a name bound twice.** | 4 | 6 | 8 | 9 | 4.3, 3.1 | 24.2 |
| 3 | 0.1.0 |  | **Rewrite the shipped categories so facts that belong together come from one named pick.** | 5 | 5 | 8 | 9 | 3.1 | 23.4 |
| 4 | 0.1.0 | decision | **Draw every `{…}` afresh, keep a pick only by name, and delete `drawGroup` with the fences that held picks implicitly.** | 7 | 8 | 8 | 10 | 4.2, 3.1 | 22.2 |
| 23 | 0.1.0 | defect | **Draw `en_US.phone`'s `exch` as a NANP central office code, and assert the rule in `TestShippedUSPhone`.** | 2 | 2 | 6 | 8 | 3 | 22.1 |
| 52 | 0.1.0 |  | **Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.** | 1 | 2 | 6 | 6 | 1.1 | 22.1 |
| 5 | 0.1.0 |  | **Run the nine-seat comprehension panel after the draw restructure, and file what it names.** | 1 | 3 | 3 | 9 | 2 | 21.2 |
| 24 | 0.1.0 | defect | **Give a `sv_SE.personnummer` over 100 the `+` separator Skatteverket spells, or stop drawing birthdates that reach 100.** | 2 | 2 | 5 | 8 | 3 | 20.1 |
| 31 | 0.1.0 |  | **Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.** | 2 | 4 | 6 | 7 | 3 | 20.1 |
| 12 | 0.1.0 | defect | **Strip the whitespace `data-import/territory.py` copies from its register.** | 1 | 1 | 5 | 7 | 3 | 19.8 |
| 6 | 0.1.0 | question | **Decide how the comprehension gate treats panel noise: identical code scored 5.7 and 5.4.** | 1 | 1 | 3 | 8 | 2 | 19.8 |
| 2 | 0.1.0 | decision | **Step up from a row to the row it links to with `..`: `{/city[Oslo]..country.name}` renders `Norway`, and `city[Oslo].country` stays the cell `NO`.** | 4 | 5 | 7 | 8 | 4.4 | 18.7 |
| 22 | 0.1.0 | defect | **Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.** | 1 | 1 | 4 | 7 | 3 | 17.8 |
| 26 | 0.1.0 |  | **Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.** | 4 | 4 | 6 | 7 | 3.1 | 17.6 |
| 48 | 0.1.0 |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 4 | 5 | 6 | 6 | 3 | 16.2 |
| 18 | 0.1.0 |  | **Give the address records one column set across countries: `region` and `municipality` as columns on `geo.SE.address` too.** | 4 | 4 | 6 | 6 | 4 | 15.9 |
| 49 | 0.1.0 | defect | **Read the tag through `git/ref/tags/{tag}` in `publish_release.py`.** | 2 | 1 | 3 | 6 | 1.1 | 15.4 |
| 47 | 0.1.0 |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 2 | 5 | 7 | 5 | 15.1 |
| 33 | 0.1.0 |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 5 | 6 | 7 | 8 | 5 | 15.0 |
| 19 | 0.1.0 | defect | **Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.** | 2 | 2 | 4 | 6 | 3 | 15.0 |
| 37 | 0.1.0 |  | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 6 | 7 | 7 | 8 | 4.7 | 14.7 |
| 17 | 0.1.0 |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 5 | 6 | 6 | 6 | 4 | 14.4 |
| 43 | 0.1.0 | defect | **Refuse a `Fake` path into a level carrying a `repeat`, as the README's Correlated fields section states.** | 2 | 2 | 4 | 6 | 4 | 14.4 |
| 50 | 0.1.0 |  | **Document `NewRecordTemplate`, `ErrNoColumns`, `ErrNoData` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 6 | 7 | 10 | 14.1 |
| 27 | 0.1.0 |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 4 | 4 | 5 | 6 | 4 | 13.9 |
| 29 | 0.1.0 |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 3 | 3 | 4 | 6 | 3.1 | 13.4 |
| 13 | 0.1.0 | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 2 | 2 | 3 | 6 | 3 | 13.0 |
| 10 | 0.1.0 |  | **Test that every node kind reaches each switch over node kinds.** | 1 | 3 | 2 | 6 | 2.2 | 12.8 |
| 15 | 0.1.0 |  | **Add the remaining `misc` tables and templates, one row of its detail's table per chunk.** | 3 | 9 | 8 | 7 | 13, 12 | 12.4 |
| 11 | 0.1.0 |  | **Test that every `geo` tree holds the five table names and the `address` columns that port across countries.** | 1 | 2 | 3 | 5 | 3 | 12.4 |
| 45 | 0.1.0 |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 2 | 4 | 6 | 5 | 12.1 |
| 16 | 0.1.0 |  | **Add the remaining locale categories: company, phone, finance, vehicle, words.** | 3 | 8 | 8 | 7 | 13 | 11.8 |
| 8 | 0.1.0 |  | **Move `binding`, `bind` and `checkNodeFences` out of `data.go` into a file of their own.** | 1 | 2 | 2 | 5 | 2.1 | 11.6 |
| 9 | 0.1.0 |  | **Move `table.route`, `selector`, `step` and `drawStep` from `path.go` to `table.go`.** | 1 | 2 | 2 | 5 | 2.1 | 11.6 |
| 28 | 0.1.0 |  | **Rename `misc.territory.country` to `sovereign`.** | 3 | 2 | 4 | 7 | 4.8 | 11.5 |
| 38 | 0.1.0 |  | **Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.** | 2 | 2 | 4 | 6 | 5 | 11.1 |
| 25 | 0.1.0 |  | **Merge `en_US.ip` and `sv_SE.ip`, today byte-identical, into one `misc.ip`.** | 3 | 2 | 4 | 6 | 4.7 | 10.9 |
| 14 | 0.1.0 |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`, and refuse a repeat over a lone `{digits(1)}` at `New`, naming that spelling.** | 3 | 3 | 4 | 6 | 4.7 | 10.6 |
| 21 | 0.1.0 |  | **Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.** | 3 | 4 | 7 | 8 | 14 | 10.5 |
| 35 | 0.1.0 | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 3 | 6 | 5 | 10.4 |
| 40 | 0.1.0 |  | **Name the node a selector follows in `stepInto`'s refusal.** | 1 | 1 | 3 | 6 | 5 | 10.4 |
| 44 | 0.1.0 |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 3 | 6 | 5 | 10.4 |
| 30 | 0.1.0 |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 4 | 5 | 4 | 5 | 4 | 10.2 |
| 36 | 0.1.0 |  | **Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 2 | 3 | 6 | 5 | 10.1 |
| 46 | 0.1.0 |  | **Spell a table one way across the errors that name it.** | 1 | 2 | 3 | 6 | 5 | 10.1 |
| 41 | 0.1.0 |  | **Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.** | 2 | 2 | 3 | 7 | 5, 8 | 10.1 |
| 32 | 0.1.0 |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 2 | 4 | 5 | 7 | 10.0 |
| 20 | 0.1.0 |  | **List `title` in both locales, and `sv_SE`'s `birth-number`, in the README's Data list of what each locale carries.** | 1 | 1 | 4 | 6 | 10 | 9.9 |
| 7 | 0.1.0 |  | **Fill `arm` in one place.** | 4 | 4 | 2 | 6 | 2.2 | 9.6 |
| 34 | 0.1.0 |  | **Rename `newRand` to `newGeneratorState` and `Generator.rand` to `state`.** | 1 | 1 | 1 | 5 | 2.3 | 9.2 |
| 39 | 0.1.0 |  | **Call `templateError` a compile failure in its doc.** | 1 | 1 | 1 | 5 | 2.3 | 9.2 |
| 51 | 0.1.0 | defect | **Build on a manual run of `test.yml`.** | 1 | 1 | 1 | 4 | 1 | 8.8 |
| 42 | 0.1.0 | defect | **Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.** | 1 | 1 | 2 | 5 | 5 | 7.4 |
| 57 | 0.2.0 |  | **Pair a street with its exact postnummer.** | 3 | 6 | 5 | 8 | 3.1 | 17.7 |
| 56 | 0.2.0 |  | **Add `{btc()}` and `{eth()}`.** | 2 | 4 | 4 | 6 | 3 | 14.5 |
| 54 | 0.2.0 |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 1 | 5 | 5 | 10 | 11.3 |
| 60 | 0.2.0 |  | **Cut the README's Layout block to the lines that say what a file name cannot.** | 1 | 1 | 2 | 5 | 2.4 | 10.9 |
| 55 | 0.2.0 |  | **Add the locales nb_NO, da_DK, fi_FI, de_DE, en_GB, nl_NL, fr_FR and es_ES, and their `geo/` trees.** | 4 | 9 | 8 | 7 | 13 | 10.6 |
| 61 | 0.2.0 |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 6 | 2.4 | 10.3 |
| 59 | 0.2.0 |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 3 | 4 | 5 | 4.8 | 10.2 |
| 58 | 0.2.0 |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 4 | 1 | 6 | 2.4 | 9.8 |
| 62 | 1.0.0 |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 3 | 7 | 9 | 1.1 | 29.7 |
| 63 | 1.0.0 |  | **Publish a homepage with an in-browser generator, the library compiled to WebAssembly.** | 3 | 6 | 6 | 6 | 1.1 | 19.1 |

## Details

### 53. Ship prebuilt binaries, a container image, Homebrew and Scoop, so the CLI needs no Go.

GoReleaser attaches them to the release the tag workflow publishes. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag.

### 1. Bind a pick to a name with `{x as n}`, read it with `{n}` and `{n.path}`, and refuse a name bound twice.

- A name is one pick of everything under it, drawn once: `{/country as k} {k.city.name} ({k.city.population})` is one city of `k`'s country.
- A name is visible in the template that binds it and everything that template renders, its sibling fields and a record's columns included, so one field may bind and another read.
- A name bound inside a `repeat` level picks again on every iteration; one bound outside it keeps its pick on every line.
- Refused at load: a name bound twice in one scope, and a name that is also a field or an option of the template, since options and fields share one namespace.

### 3. Rewrite the shipped categories so facts that belong together come from one named pick.

- `sv_SE.person` and `en_US.person` bind the first name and read sex and title through it: `"first": "{.first-name as n}{n.name}"`, `"sex": "{n..sex.name}"`, a title from `{n..sex.title.name}`.
- `geo.SE.address` and `geo.US.address` bind the street and read locality, postal code and region through it; the locale `address` categories bind `/geo.XX.address` once.
- Needs item 1 and 2. Every shipped value must still pass its consumer's check before and after item 4.
- Re-pin seeded output in its own commit.

### 4. Draw every `{…}` afresh, keep a pick only by name, and delete `drawGroup` with the fences that held picks implicitly.

- Needs items 1–3. After it, `{/person} ({/person.first})` renders two people, `{w} {uppercase(w)}` renders `b G` where `w` is a choice of letters, and `{/city.name}, {/country[SE].name}` and `{/country[SE].name} / {/country[NO].name}` both load.
- Deletes the token-order rules (`drawsApart`, `readFold`, `branches`, the pairwise replay), the expansion hold fence, `drawGroup`, the own-family fence and the cell-reference fences, and revises the decisions that point here. The remaining draw checks are names, paths through names, cycles and the repeat cap.
- Fixes on the way: `{/sel}|{/sel}` panics out of `Fake` today where two draws land on rows whose cells select different rows of another table; it must render.
- Rewrites the README's References, Draw group, Correlated fields, Linked tables and "One draw, one spelling" sections; the Linked tables section says which pick's weights govern a family, since the bound row decides.
- Re-pin seeded output and `testdata/shipped_shape.txt` in their own commits; the CHANGELOG names the grammar change.

### 23. Draw `en_US.phone`'s `exch` as a NANP central office code, and assert the rule in `TestShippedUSPhone`.

`{int(100,999)}` renders a leading 1 in about an eighth of draws, which libphonenumber rejects, and the test proves the shape rather than the rule.

### 52. Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.

Nothing has shipped, so "no longer paths", "where it used to fail" and "where it used to be an empty string" describe versions no reader can have installed.

### 5. Run the nine-seat comprehension panel after the draw restructure, and file what it names.

Needs item 4, and runs as item 6 decides.

### 24. Give a `sv_SE.personnummer` over 100 the `+` separator Skatteverket spells, or stop drawing birthdates that reach 100.

The format hard-codes `-`, and the 1930 floor makes the oldest draws invalid from 2031.

### 31. Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.

`isin`: Luhn over letters expanded to digits. `aba`: 3-7-1 weights. `vin`: position 9 over the whole; a sample taking the WMI, since the check sits mid-string.

### 12. Strip the whitespace `data-import/territory.py` copies from its register.

`misc.territory[CW].capital` renders ` Willemstad`, with a leading space.

### 6. Decide how the comprehension gate treats panel noise: identical code scored 5.7 and 5.4.

Two nine-seat runs on `5cb919e`, same briefs, same model, 2026-09-30:

| Seat | Run 1 | Run 2 |
|---|---|---|
| Junior A | 5 | 5 |
| Junior B | 4 | 5 |
| Mid A | 6 | 5 |
| Mid B | 6 | 6 |
| Senior domain | 6 | 5 |
| Senior maintainability | 6 | 6 |
| Senior extender | 6 | 5 |
| Architect greenfield | 6 | 6 |
| Architect inherited | 6 | 6 |
| Mean (Navigation / Locality / Shape / Self-sufficiency) | 5.7 (7.0 / 5.1 / 5.9 / 5.7) | 5.4 (6.3 / 4.8 / 5.8 / 5.3) |

Both runs named the same hardest units, `drawsApart` and `readFold.reads`. Scores since 2026-09-20 range 5.7–6.1, wider than this 0.3 spread by only 0.1, so `AGENTS.md`'s "no merge lowers the last score" can block or pass a chunk by chance. The last recorded nine-seat score is 5.7, on 2026-09-30; the ratchet reads it from here until this item is decided.

### 2. Step up from a row to the row it links to with `..`: `{/city[Oslo]..country.name}` renders `Norway`, and `city[Oslo].country` stays the cell `NO`.

- `..` goes up one level, to the parent table, and the name after it is that parent's: `{c..country.name}`, and from a locality `{l..municipality..region.name}`. Naming any other table is a load error naming the parent.
- It steps up from a selected row, a drawn row or a name bound to a row.
- After `..`, `.` steps down again: `{c..country.city.name}` is a city of `c`'s country, drawn afresh.
- `.country` after a row keeps reading the link column's cell, so the README's `city[Oslo].country` example stays true.
- A `Fake` path and `--list` take `..` too; `List` keeps advertising direct descents only.

### 22. Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.

ISO 3166 reserves it for Western Sahara and the root zone has never been delegated it, so no resolver answers for it.

### 26. Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.

Needs item 4.

### 48. Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.

Today only a JSON field takes one, so `misc.httpmethod`'s booleans are typed in Go and text everywhere else. Typing a shipped column changes what `json` and `sql` write, so the capability is a minor and applying it to `misc.httpmethod` is a major.

### 49. Read the tag through `git/ref/tags/{tag}` in `publish_release.py`.

GitHub answers `tags/{tag}` with 404 whatever exists, so the burnt-version guard never fires and a release lands on a tag already pointing at another commit.

### 47. Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.

With several `--data-path` layers the author has to work out which directory won, while a parse error on the same file names it.

### 33. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 19. Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.

Neither names a `weight`, so a bare `sex` draws evenly.

### 37. Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.

`geo.SE.municipality[Lund]` and `misc.territory[Sweden]` select a row today. Steps:
- Have a selector that misses name the key of every row whose cell spells it, matched without case, else say the table selects by its key: `misc.territory[Sweden]` and `misc.territory[se]` name `[SE]`, `misc.protocol[tcp]` its key, and `misc.loglevel[Error]` `[err]`.
- Delete the `name` option and the load checks it carries, and have `New` refuse a table still carrying `name`, naming `key`.
- Key `misc.loglevel` on `keyword`, and `geo.SE.locality` on the integer `data-import/geo-se.py` assigns as the decision states, linking `street` and `postal-code` by it.
- Drop `key` from `last-name` and `sv_SE.title`, and re-pin `testdata/shipped_shape.txt`.
- Leave no selection by name in README.md, CHANGELOG.md, docs/ or `--help`.

### 17. Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.

`misc.uuid` gains `variant`, for one.

### 43. Refuse a `Fake` path into a level carrying a `repeat`, as the README's Correlated fields section states.

`probePath` skips `pathCheck.enter`, so `Fake("x.a")` under `"repeat":3` renders one draw.

### 27. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 13. Place xlsx cells by their `r` reference in `data-import/xlsx.py`.

A sheet omitting an empty cell shifts every later column left.

### 10. Test that every node kind reaches each switch over node kinds.

After item 4; recount the switches then. Today ten, from `render` to `columnItems`, must agree, and only `render` and `renderEdges` say so.

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

### 11. Test that every `geo` tree holds the five table names and the `address` columns that port across countries.

Nothing checks it, and a new country breaks it silently.

### 45. Reword `prove`'s refusal of a typed column reading a row, `{/region}`.

Say the format is the table's own (`region's format "{name}"`), name the read to write, `{/region.<column>}` with its columns, and stop calling a one-column format, `{code}`, composed text.

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

### 8. Move `binding`, `bind` and `checkNodeFences` out of `data.go` into a file of their own.

The greenfield architect found the fence pipeline last. After item 4, which shrinks the pipeline.

### 9. Move `table.route`, `selector`, `step` and `drawStep` from `path.go` to `table.go`.

The inherited architect's 3am trace crossed four files for one draw.

### 28. Rename `misc.territory.country` to `sovereign`.

It holds a code, and `country` collides with the `misc.country` path it replaced.

### 38. Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.

`datatype: boolean` over `{int(1,2)}` on an `int` field answers `prints an integer, not a boolean`.

### 14. Spell `misc.creditcard`'s digit runs `{digits(n)}`, and refuse a repeat over a lone `{digits(1)}` at `New`, naming that spelling.

Both render the same run.

### 21. Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.

17 of the 40 distinct ones shipped today sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost` provably reach nothing.

### 35. Compare `calcParser.binary`'s operator as a rune.

`byte(p.rs[p.pos])` reads U+012B `ī` as `+`, so `{calc(a ī b)}` compiles as `a + b`.

### 40. Name the node a selector follows in `stepInto`'s refusal.

`sv_SE.person[1].first` answers `1 is not a table`.

### 44. Report a CRLF rows file holding only its header as having no rows.

`parseRows` answers "line 2 is empty".

### 36. Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.

An `int64` or `float64` field unproven in range is told to become itself, and a `uint64` one to become `int64`.

### 46. Spell a table one way across the errors that name it.

`pinRow` names `t.path` and `mustRow` panics with `t.category`, so one table is `territory` and `misc.territory`, and the short spelling names no file where two folders hold that name.

### 41. Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.

`readTableOptions` and `loadDir` return on the first in Go's map order.

### 7. Fill `arm` in one place.

`splitArm` runs before the link and again after it, and `compileArm` finishes it from `reference.go`, so five seats traced three phases. After item 4.

### 34. Rename `newRand` to `newGeneratorState` and `Generator.rand` to `state`.

Both name half of a `generatorState`, hiding its `{seq()}` counters, and `New`'s local `rng` shadows the `rng` interface.

### 39. Call `templateError` a compile failure in its doc.

Its doc says render failure, and it wraps what `NewTemplate` and `NewRecordTemplate` refuse.

### 51. Build on a manual run of `test.yml`.

`workflow_dispatch` leaves `github.event.before` empty, so the diff compares `HEAD` with itself and skips `docker build`.

### 42. Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.

`loadDir` indexes hidden TSVs and skips them in two places, so one loads where its comment says a hidden file is never data.

### 57. Pair a street with its exact postnummer.

Needs an application to Lantmäteriet; today a street goes to the nearest postal code centroid. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 56. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 54. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 60. Cut the README's Layout block to the lines that say what a file name cannot.

`calc.go`, `doc.go` and `cmd/fejkdata/` restate their names.

### 55. Add the locales nb_NO, da_DK, fi_FI, de_DE, en_GB, nl_NL, fr_FR and es_ES, and their `geo/` trees.

Each locale carries person, address, phone, national id, company and date names. Their `address` reads the `geo/` trees NO, DK, FI, NL, FR, AU, CA, ES, GB and DE by reference; AU and CA get a tree with no locale.

### 61. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens with the README's GitHub decision, so it is a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 59. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go.

### 58. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 62. Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.

The human cuts v1.0.0 once the shipped data is in its record shape and one full minor has shipped with no breaking change, per the README's Versioning table; v0.2.0 rewrites shipped rows, so a v0.3.0 with no breaking change comes first.
