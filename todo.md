# todo

## Scoring

`Score = -R - S/4 + 2*A + 2*G*W`

`Bar = 9`

`Next ID = 160`

| Goal | W |
|---|---|
| 1 | 1.00 |
| 1.1 | 1.00 |
| 2 | 0.95 |
| 2.1 | 0.95 |
| 2.2 | 0.95 |
| 2.3 | 0.95 |
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
| 5.9 | 0.80 |
| 6 | 0.75 |
| 6.1 | 0.75 |
| 6.2 | 0.75 |
| 7 | 0.70 |
| 7.1 | 0.70 |
| 7.2 | 0.70 |
| 7.3 | 0.70 |
| 8 | 0.65 |
| 9 | 0.60 |
| 9.1 | 0.60 |
| 9.2 | 0.60 |
| 9.3 | 0.60 |
| 9.4 | 0.60 |
| 9.5 | 0.60 |
| 9.6 | 0.60 |
| 10 | 0.55 |
| 10.1 | 0.55 |
| 11 | 0.50 |
| 12 | 0.45 |
| 12.1 | 0.45 |
| 12.2 | 0.45 |
| 12.3 | 0.45 |
| 13 | 0.40 |
| 13.1 | 0.40 |
| 14 | 0.35 |
| 15 | 0.30 |
| 16 | 0.25 |

## Items

| ID | Release | Exempt | Item | R | S | A | G | Goals | Score |
|---|---|---|---|---|---|---|---|---|---|
| 77 | 0.1.0 | decision | **Split the shipped data into Go modules per locale and country, which a library imports by choice, and embed every one in the CLI.** | 5 | 10 | 10 | 10 | 2, 1.1, 9 | 32.5 |
| 78 | 0.1.0 | decision | **Ship 10% of every place's streets, at least 10, and 10% of all US places drawn by population, each place weighted by population.** | 6 | 7 | 10 | 10 | 2, 14 | 31.2 |
| 61 | 0.1.0 |  | **Add a locale, with its `geo/` tree, for each of the ten most-spoken languages and each Nordic country.** | 4 | 9 | 9 | 10 | 2.1 | 30.8 |
| 16 | 0.1.0 |  | **Add locale categories: company, phone, finance, vehicle, words.** | 3 | 8 | 8 | 9 | 2.2, 15 | 28.1 |
| 15 | 0.1.0 |  | **Add `misc` tables and templates, one row of its detail's table per chunk.** | 3 | 9 | 8 | 8 | 2, 15, 14 | 25.9 |
| 129 | 0.1.0 | decision | **Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.** | 5 | 8 | 8 | 10 | 5.9, 7 | 25.0 |
| 57 | 0.1.0 |  | **Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.** | 1 | 2 | 6 | 6 | 1.1 | 22.5 |
| 96 | 0.1.0 |  | **Refuse two reads into one sibling, category or struct category that no name joins, `{place.postal-code} {place.locality}`, showing the spelling for one pick and the one for two.** | 4 | 4 | 7 | 7 | 5.6, 4 | 20.9 |
| 153 | 0.1.0 | decision | **Run the nine-seat comprehension panel and the scoring run again once items 145 to 152 and item 10 have merged, and file what the panel names.** | 1 | 3 | 3 | 9 | 3 | 20.4 |
| 157 | 0.1.0 |  | **Show in the README's CLI section how PowerShell and cmd.exe pipe a template, one line per shell family.** | 1 | 1 | 5 | 6 | 7.2 | 17.1 |
| 89 | 0.1.0 |  | **Fail `New` without `WithSeed`, naming the option, and seed the CLI from the system.** | 3 | 3 | 6 | 7 | 10.1 | 16.0 |
| 145 | 0.1.0 |  | **Move `doc.go`'s definitions of a level, a pick key and an addressed key, with their worked example, to `pickKey` in `pick.go`, and gather `levelKeys`, `addressedKeys`, `renderEnv.keeps` and `readField`'s kept read beside `readUnder`, so a reader follows a named read in one file.** | 2 | 3 | 1 | 9 | 3.4, 3.2 | 15.4 |
| 127 | 0.1.0 |  | **Give a Swedish street only postal codes whose delivery digit means street or rural delivery, dropping codes whose delivery digit is 9, or record why they stay.** | 3 | 3 | 4 | 6 | 4 | 14.4 |
| 110 | 0.1.0 |  | **Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 10 | 0.1.0 |  | **Test that every node kind reaches each switch over node kinds.** | 1 | 3 | 2 | 6 | 3.2 | 13.1 |
| 148 | 0.1.0 |  | **State in `geo.py` the locality row each country's script hands it, and meet the fence that every parent row has a child row in one shared step.** | 2 | 3 | 2 | 6 | 2.1, 3.2 | 12.6 |
| 146 | 0.1.0 |  | **Move `sharedPaths` and `subPaths` from `list.go` beside `carriedByAll` in `path.go`, and `join` beside `sortedNames` in `sorted.go`.** | 1 | 2 | 1 | 6 | 3.1, 3.2 | 11.3 |
| 147 | 0.1.0 |  | **Rename `namedNode`, `compiledPath.level` and `checkCells`, so "named" means bound to a name, "level" means only a prefix of a read's path, and `checkCells`' new name says it compiles.** | 1 | 2 | 1 | 6 | 3.3 | 11.3 |
| 151 | 0.1.0 |  | **Load a caller's path and walk it in one `Generator` method, so no entry point walks the root before the shipped category the path names has loaded.** | 2 | 2 | 1 | 6 | 3.2 | 10.3 |
| 149 | 0.1.0 |  | **Say in one line what `config.load` does and what the 64-bit assertion asserts, beside their decision links.** | 1 | 1 | 1 | 5 | 3.4 | 9.8 |
| 150 | 0.1.0 |  | **Move `columnRead` and `columnReadOf` from `resolve.go` into `datatype.go`, and state at `template.readsColumn` that `renderLeaf`, `proveColumnItem` and `itemDatatype` change with it.** | 1 | 2 | 1 | 5 | 3.1, 3.2 | 9.5 |
| 152 | 0.1.0 |  | **Delete the doc lines that restate their code.** | 1 | 2 | 1 | 5 | 3.4 | 9.5 |
| 58 | 0.2.0 |  | **Ship prebuilt binaries, a container image, and packages for Homebrew, Scoop and the biggest Linux package managers, so the CLI needs no Go.** | 3 | 7 | 8 | 9 | 1.1, 7 | 29.2 |
| 113 | 0.2.0 | defect | **Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.** | 3 | 5 | 7 | 9 | 4 | 25.0 |
| 154 | 0.2.0 |  | **Add `en_US.routing`, an ABA routing number over the Fed's prefix ranges with its check digit drawn by `{aba()}`.** | 2 | 2 | 6 | 7 | 2.2, 15 | 22.8 |
| 155 | 0.2.0 |  | **Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.** | 2 | 2 | 6 | 7 | 2, 15 | 22.8 |
| 26 | 0.2.0 | defect | **Spell `sv_SE.personnummer` with the `+` Skatteverket uses from the year the holder turns 100, judged by a date the caller gives.** | 4 | 6 | 7 | 8 | 4 | 22.1 |
| 34 | 0.2.0 |  | **Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.** | 2 | 4 | 6 | 7 | 4 | 20.9 |
| 81 | 0.2.0 | decision | **Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.** | 4 | 4 | 6 | 8 | 6.1, 4 | 20.6 |
| 91 | 0.2.0 |  | **Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.** | 3 | 4 | 6 | 7 | 4.1, 5.3 | 19.9 |
| 36 | 0.2.0 |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 5 | 6 | 7 | 8 | 6 | 19.5 |
| 88 | 0.2.0 |  | **Add a 12-digit `YYYYMMDDNNNC` form of `sv_SE.personnummer` and `sv_SE.samordningsnummer`, the form Skatteverket's testpersonnummer series and many systems store.** | 2 | 3 | 6 | 6 | 4 | 19.4 |
| 41 | 0.2.0 | decision | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 6 | 7 | 7 | 8 | 5.5 | 19.1 |
| 52 | 0.2.0 |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 2 | 5 | 7 | 6 | 19.0 |
| 117 | 0.2.0 |  | **List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.** | 1 | 2 | 5 | 7 | 6 | 19.0 |
| 28 | 0.2.0 | decision | **Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.** | 4 | 4 | 6 | 7 | 4.1 | 18.9 |
| 24 | 0.2.0 | defect | **Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.** | 1 | 1 | 4 | 7 | 4 | 18.6 |
| 92 | 0.2.0 |  | **Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.** | 1 | 2 | 5 | 6 | 6 | 17.5 |
| 53 | 0.2.0 |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 4 | 5 | 6 | 6 | 4 | 16.9 |
| 55 | 0.2.0 |  | **Document `NewRecordTemplate`, `ErrNoColumns`, `ErrNoData` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 6 | 7 | 12 | 16.8 |
| 18 | 0.2.0 |  | **Give the address records one column set across countries: `region` and `municipality` as columns on `geo.SE.address` too.** | 4 | 4 | 6 | 6 | 5 | 16.6 |
| 44 | 0.2.0 | decision | **Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.** | 1 | 1 | 4 | 6 | 6 | 15.8 |
| 54 | 0.2.0 | defect | **Read the tag through `git/ref/tags/{tag}` in `publish_release.py`.** | 2 | 1 | 3 | 6 | 1.1 | 15.8 |
| 31 | 0.2.0 |  | **Rename `misc.territory.country` to `sovereign`.** | 3 | 2 | 4 | 7 | 5.8 | 15.7 |
| 20 | 0.2.0 | defect | **Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.** | 2 | 2 | 4 | 6 | 4 | 15.7 |
| 50 | 0.2.0 |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 2 | 4 | 6 | 6 | 15.5 |
| 17 | 0.2.0 |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 5 | 6 | 6 | 6 | 5 | 15.1 |
| 29 | 0.2.0 |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 4 | 4 | 5 | 6 | 5 | 14.6 |
| 42 | 0.2.0 |  | **Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.** | 2 | 2 | 4 | 6 | 6 | 14.5 |
| 32 | 0.2.0 |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 3 | 3 | 4 | 6 | 4.1 | 14.4 |
| 159 | 0.2.0 |  | **Name the quoted spelling where a template that is a bare JSON number or boolean is refused, padded or not: `42` names `"42"`.** | 1 | 2 | 3 | 6 | 6, 5.1 | 14.1 |
| 22 | 0.2.0 |  | **Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.** | 3 | 4 | 7 | 8 | 16 | 14.0 |
| 46 | 0.2.0 |  | **Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.** | 2 | 2 | 3 | 7 | 6, 10 | 14.0 |
| 158 | 0.2.0 |  | **Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.** | 2 | 2 | 4 | 5 | 7.2, 5.1 | 13.5 |
| 79 | 0.2.0 | decision | **Beside a `--data-path`, load in `New` the user's categories, the shipped ones they override or read, and the shipped ones whose reads reach an overridden one; load the rest on first reach.** | 6 | 6 | 7 | 8 | 13 | 12.9 |
| 27 | 0.2.0 |  | **Merge `en_US.ip` and `sv_SE.ip`, today byte-identical, into one `misc.ip`.** | 3 | 2 | 4 | 6 | 8 | 12.3 |
| 107 | 0.2.0 | decision | **Move the template engine into `internal/`, leaving the root the public API and its options.** | 6 | 9 | 3 | 8 | 3.2, 3.4 | 12.2 |
| 112 | 0.2.0 |  | **Check a calc's field operands where its template resolves, beside its name operands, once item 36 reports every mistake at once.** | 2 | 2 | 1 | 6 | 3.2 | 10.3 |
| 108 | 0.2.0 | defect | **Name item 26 in the decision "No builtin reads the clock, so a date is bounded by days, never by an age", as the item that ends its conflict with goal 4.** | 1 | 1 | 1 | 4 | 3.4 | 8.0 |
| 38 | 0.3.0 | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 45 | 0.3.0 |  | **Name the node a selector follows in `childNamed`'s refusal.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 49 | 0.3.0 |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 3 | 6 | 6 | 13.8 |
| 13 | 0.3.0 | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 2 | 2 | 3 | 6 | 4 | 13.7 |
| 39 | 0.3.0 |  | **Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 2 | 3 | 6 | 6 | 13.5 |
| 51 | 0.3.0 |  | **Spell a table one way across the errors that name it.** | 1 | 2 | 3 | 6 | 6 | 13.5 |
| 115 | 0.3.0 |  | **Spell a rows file's line one way in every error, naming the file.** | 1 | 2 | 3 | 6 | 6 | 13.5 |
| 14 | 0.3.0 |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`.** | 2 | 2 | 4 | 6 | 8 | 13.3 |
| 11 | 0.3.0 |  | **Test that every `geo` tree holds the five table names and the `address` columns that port across countries.** | 1 | 2 | 3 | 5 | 4 | 13.0 |
| 35 | 0.3.0 |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 2 | 4 | 5 | 8 | 13.0 |
| 40 | 0.3.0 |  | **Parse `defaultTable`'s path with the library's grammar.** | 2 | 2 | 3 | 5 | 3.2 | 12.5 |
| 21 | 0.3.0 |  | **List `title` in both locales, and `sv_SE`'s `birth-number`, in the README's Data list of what each locale carries.** | 1 | 1 | 4 | 6 | 12 | 12.2 |
| 86 | 0.3.0 |  | **Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.** | 1 | 2 | 2 | 5 | 3.2 | 11.5 |
| 23 | 0.3.0 |  | **Give every other shipped category that could reach something real a path that never does.** | 3 | 6 | 6 | 7 | 16 | 11.0 |
| 75 | 0.3.0 |  | **Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.** | 2 | 6 | 4 | 7 | 12.2 | 10.8 |
| 33 | 0.3.0 |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 4 | 5 | 4 | 5 | 5 | 10.8 |
| 47 | 0.3.0 | defect | **Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.** | 1 | 1 | 2 | 5 | 6 | 10.2 |
| 72 | 0.3.0 | principle | **Fail the merge gate when test coverage falls below the last recorded figure.** | 2 | 3 | 2 | 5 | 3 | 10.2 |
| 30 | 0.3.0 |  | **Draw `email.local` from `username`, so the two share one handle list.** | 2 | 2 | 3 | 5 | 8 | 10.0 |
| 95 | 0.3.0 | principle | **Hold a struct's columns and the field each fills in one slice.** | 1 | 2 | 1 | 5 | 3.2 | 9.5 |
| 116 | 0.3.0 |  | **Turn a row index into its rows file's line in one function.** | 1 | 2 | 1 | 5 | 3.2 | 9.5 |
| 19 | 0.3.0 |  | **Fill the 398 Swedish localities weighted 200 from SCB småorter.** | 2 | 4 | 4 | 6 | 14 | 9.2 |
| 56 | 0.3.0 | defect | **Build on a manual run of `test.yml`.** | 1 | 1 | 1 | 4 | 1 | 8.8 |
| 109 | 0.3.0 | principle | **Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.** | 2 | 4 | 1 | 4 | 3 | 6.2 |
| 63 | 0.4.0 |  | **Pair a street with its exact postnummer.** | 3 | 6 | 5 | 8 | 4.1 | 19.1 |
| 62 | 0.4.0 |  | **Add `{btc()}` and `{eth()}`.** | 2 | 4 | 4 | 6 | 4 | 15.2 |
| 59 | 0.4.0 |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 1 | 5 | 5 | 12 | 13.2 |
| 65 | 0.4.0 |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 3 | 4 | 5 | 5.8 | 13.2 |
| 66 | 0.4.0 |  | **Cut the README's Layout block to the lines that say what a file name cannot.** | 1 | 1 | 2 | 5 | 3.4 | 11.8 |
| 67 | 0.4.0 |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 6 | 3.4 | 11.3 |
| 64 | 0.4.0 |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 4 | 1 | 6 | 3.4 | 10.8 |
| 68 | 0.4.0 | decision | **Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.** | 2 | 1 | 1 | 3 | 1.1 | 5.8 |
| 69 | 1.0.0 |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 3 | 7 | 9 | 1.1 | 30.2 |
| 70 | 1.0.0 |  | **Publish a homepage with an in-browser generator, the library compiled to WebAssembly.** | 3 | 6 | 6 | 6 | 1.1 | 19.5 |

## Details

### 77. Split the shipped data into Go modules per locale and country, which a library imports by choice, and embed every one in the CLI.

The CLI binary may grow to hundreds of MB. The maintainer decided on 2026-10-04, under goals 2.3, 6.2, 7.3, 9 and 10.1:
- One module per locale, one per country's `geo/` tree, and one for `misc`. A module registers nothing when imported.
- A bare `New()` loads no data and fails, naming the option to add. `WithoutShippedData` goes.
- A data module is an `fs.FS` passed to `WithDataFS`, as anyone's data is, and `New` takes several. A module carrying functions loads through an option of its own, so a data module that starts carrying functions breaks its users. Such a function reaches the randomness through an interface the root declares; `internal/drawstate`'s type stays internal.
- A module names, in a manifest, every module it reads by default, directly or through another, so the first error names the whole set to import. A manifest is optional: a `--data-path` folder without one is a module. A read nothing provides fails, naming the module that provides that path by default or saying none does. Nothing loads a default on its own, and anything providing the same paths stands in.
- A module's manifest names the fejkdata version it was built for, and a mismatch fails in `New`, naming the `go get` line that aligns them. Each shipped module requires the core at its own version.
- Every shipped module releases in lockstep, under one version number, so the README's one version still covers the library, the CLI and the data. CI cuts one tag per module per release, which revises the decision "The changelog heading is the one spelling of a release; CI cuts the tag".
- Two sources defining one name fail to load, unless an option says the second replaces the first; every read of that name, a shipped module's included, then reads the second. In the CLI, a flag standing on its own, benched with the data and hand fixture authors, does the same for a `--data-path`, so flags still go anywhere on the line. This revises the decisions "A `--data-path` override rebinds every reference to the category it replaces" and "With only the shipped set, a category loads on the first call reaching it; beside a `--data-path`, every category loads in `New`".
- The builtins stay in fejkdata itself. A module's function asks for the inputs it needs through one generic mechanism, and gets the randomness and the date the generator uses. The README asks for functions with no side effects of their own.
- The CLI is a library call. One exported list names every shipped module, and `fejkdata`'s own `main` passes it to that call. The list lives where no library user's module graph pulls in every shipped module, and the architect places it. A custom CLI is the same few lines with other modules added, and no code exists only for custom CLIs. A README section shows it with example code, and the README's Library quick start shows the whole import block for one locale. The README's Versioning section gives each module's tag spelling, and says to pin every fejkdata module at one version.

Split it into items before starting. Before the module API ships, a panel of the README's audience personas tries pulling in data with it.

### 78. Ship 10% of every place's streets, at least 10, and 10% of all US places drawn by population, each place weighted by population.

Needs item 77. Ten streets per place crowds 500 Stockholm customers onto ten streets. `geo-se.py` and `geo-us.py` keep 10% of each place's streets, at least 10. `geo-us.py` draws 10% of every Census place, down to the smallest village, picking each with odds by population, and every table weights a place by its population, so a draw is not mostly tiny villages. Replaces the decision "The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more", which the maintainer withdrew on 2026-10-03. Re-pin seeded output and the shape in their own commits.

### 61. Add a locale, with its `geo/` tree, for each of the ten most-spoken languages and each Nordic country.

Pick the ten from a published ranking of languages by total speakers, such as Ethnologue's, each in the country where it has the most speakers, and record the ranking in the decision that names them. The Nordic ones are sv_SE (shipped), nb_NO, da_DK, fi_FI and is_IS. A second country reading GeoNames' postal codes moves `geonames()` from `geo-se.py`, and the nearest-centroid street assignment in its `streets()`, into `geo.py`. Each locale carries goal 2.2's categories, and its `address` reads its country's `geo/` tree by reference. `rows.Pins` holds eight pins inline, sized for a five-deep geo tree; a deeper tree needs more to keep a render off the heap.

### 16. Add locale categories: company, phone, finance, vehicle, words.

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
| `phone` NANP with valid NPA (the NXX ships as `en_US.phone-exchange`) | t | NANPA rules | facts |
| `bankaccount` | t | — | — |
| `licenseplate` per state | T | hand-authored patterns | facts |
| `car` make, model | T | NHTSA vPIC | public domain |
| `word`, `sentence`, `paragraph` | T+t | Moby POS or WordNet | public domain / WordNet |
| month and weekday names, `holiday` | t+T | CLDR en | Unicode |

### 15. Add `misc` tables and templates, one row of its detail's table per chunk.

Shape: T = table, t = template, c = choice.

| Category | Shape | Source | Licence |
|---|---|---|---|
| `ip` v4 documentation ranges, private, CIDR; `ipv6` `2001:db8::/32`; `mac` locally administered | t | RFC 5737, 1918, 9637 | facts |
| `creditcard` per network: IIN, length, CVV, expiry | T+t | network rules, `{luhn()}` | facts |
| `bic` | t | structure rules | facts |
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

### 129. Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.

Goal 5.9. Today a path is a second spelling of the template that reads it; afterwards only the template remains.

- A lone reference, a template of one reference token and nothing else, is what it references, as a record column reading one reference already is: `FakeRecord("{/users}")` returns the columns of `users`. The CLI's `loneReference` and the `isPath` check it copies from the token grammar then go, and `--format sql` takes its default `--table` from the library's parse of the reference.
- `Fake` and `FakeRecord` take a template. `FakeTemplate` folds into `Fake`, `FakeRecordTemplate` into `FakeRecord`, and `IsTemplate` goes, with `grammar.IsTemplate`.
- `Fake` of a lone reference costs what `Fake` of its path costs today, and the allocation gate holds a case for it.
- `List` and `--list` give each entry as the template that renders it, `{/sv_SE.person}`.
- Text that is not a JSON value and holds no `{…}` prints as written (goal 5.1), in a string field's struct tag and `Fake` alike, as stdin's text does in the CLI: `Fake("sv_SE.person")` returns `sv_SE.person`. `CHANGELOG.md` says that a path with no `{…}` prints as its text. A struct tag's `-` still skips its field.
- Where text with no `{…}` is refused, as a record or in a typed struct field, and it spells a path the data offers, the error names `{/that.path}`.
- A lone reference to nothing exits 2 under `--format`, where it exits 1 today.
- A lone reference written as an object holding only a format, `{"format":"{/users}"}`, or as a one-item choice, `["{/users}"]`, names the record as `{/users}` does.
- `grammar.CallerPath`, which strips a path's leading `/`, goes with the path spelling, and so does the README's Library line allowing it.
- Decide whether a lone reference under `--format` draws as its path draws today or as a template draws: under one seed, drawing `users` as a path and as the template `{/users}` picks different rows.
- This revises the decisions "A struct tag is a template by its shape", "`FakeTemplate` and `NewTemplate` both stay", "`--list` stays a plain list of paths", "The CLI renders exactly the template on its stdin, and adds nothing", and the exit-1 rule for a lone reference to nothing in "A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129". The README's examples and exit codes, the godoc examples in `doc.go` and `CHANGELOG.md`'s entries move to the template spelling. Delete items 40 and 110, which this closes.

Split it into items for the CLI, struct tags and the library before starting.

### 57. Rewrite `CHANGELOG.md`'s `[Unreleased]` as what v0.1.0 holds.

Nothing has shipped, so "no longer paths", "where it used to fail" and "where it used to be an empty string" describe versions no reader can have installed.

### 96. Refuse two reads into one sibling, category or struct category that no name joins, `{place.postal-code} {place.locality}`, showing the spelling for one pick and the one for two.

Asked of the maintainer by the product-owner pass of the chunk that made every `{…}` draw afresh. Since every `{…}` draws afresh, `{net} x {qty} = {calc(net * qty, 2)}`, `{place.postal-code} {place.locality}` and struct tags `sv_SE.address.street` beside `sv_SE.address.postal-code` load and may render a pair that disagrees, and nothing tells the author to bind a name (goal 4.1). Goal 5.6 asks such a template to fail, showing `{place as p}{p.postal-code} {p.locality}` for one place and two names for two. The decision "Every `{…}` draws afresh, and only a name keeps a pick" lists the draw checks as names, paths through names, cycles and the repeat cap, so this fence revises it. `{w} {w}`, one spelling twice, stays two draws. The maintainer answered yes on 2026-10-05, for reads inside one category or one struct; reads across categories, `{/city.name}, {/country.name}`, keep loading.

### 153. Run the nine-seat comprehension panel and the scoring run again once items 145 to 152 and item 10 have merged, and file what the panel names.

The decision "comprehension floor: every dimension and the overall at 7.0 or above" holds the scores this run is measured against.

### 157. Show in the README's CLI section how PowerShell and cmd.exe pipe a template, one line per shell family.

Split from item 129. cmd.exe's `echo` keeps its quotes and a trailing space, so `echo {/users} | fejkdata` sends `{/users} ` and stops naming the record.

### 89. Fail `New` without `WithSeed`, naming the option, and seed the CLI from the system.

Goal 10.1 has the caller supply the seed, and today `New` without `WithSeed` seeds itself. This item ships in 0.1.0 beside item 77, so `New` breaks once before anyone depends on it. Mark `WithSeed` required in the Library section's options table. Its tests then stop swapping `randomBytes`, as `fejkdata_test.go` does today, which fails every `New` in the package while that test runs.

### 145. Move `doc.go`'s definitions of a level, a pick key and an addressed key, with their worked example, to `pickKey` in `pick.go`, and gather `levelKeys`, `addressedKeys`, `renderEnv.keeps` and `readField`'s kept read beside `readUnder`, so a reader follows a named read in one file.

All 13 seats of the 2026-10-07 panel and scoring run on ad59967 again named the named-pick read path hardest, and twelve kept `doc.go`'s vocabulary open beside `pick.go`, `env.go` and `namefence.go` to follow it. Load builds a pick key in `levelKeys` (`arm.go`) and `addressedKeys` (`name.go`), and render builds the same key in `keeps` (`env.go`) and `readUnder` (`pick.go`); five seats named the two sides agreeing by hand as the unit they least want to change. `readUnder`'s loop stops one level short and its next line builds that level's key the same way, which one seat searched for a difference. Update the README's Layout block, and leave `doc.go` a one-line entry per term naming `pickKey`.

### 127. Give a Swedish street only postal codes whose delivery digit means street or rural delivery, dropping codes whose delivery digit is 9, or record why they stay.

`docs/research/research-geo-se.md` reads delivery digit 0 or 1 as boxes, 2–4 and 6–7 as street, 5 as rural, 8 as reply mail and 9 as competitions or temporary, with exceptions in big cities. `street_delivery` drops 0, 1 and 8 and keeps 9. The panel of 2026-10-05 found it; whether a shipped street carries a 9 code is unchecked.

### 110. Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.

Today `Fake("sv_SE.person.")` reports `no field ""`, while `{/sv_SE.person.}` reports `path has an empty segment`. `grammar.CheckSegments` is the check a reference, a sibling path and a step up `..` run.

### 10. Test that every node kind reaches each switch over node kinds.

Today nine, from `render` to `columnItems`, must agree, and only `render` and `renderEdges` say so. Three architect seats of the 2026-10-07 panel and scoring run on ad59967 named the switches kept in step by hand.

### 148. State in `geo.py` the locality row each country's script hands it, and meet the fence that every parent row has a child row in one shared step.

Needed by item 61. The extender seat of the 2026-10-07 panel on ad59967, reading for item 61, rebuilt the row's keys, `name`, `municipality`, `population`, `lat`, `lon` and `codes`, from `geo-se.py` and `geo-us.py` side by side. It found the fence met two ways: `geo-se.py`'s `main` exits on a kommun with no locality, while `geo-us.py`'s `main` prunes counties and regions left with none.

### 146. Move `sharedPaths` and `subPaths` from `list.go` beside `carriedByAll` in `path.go`, and `join` beside `sortedNames` in `sorted.go`.

Three architect seats of the 2026-10-07 panel and scoring run on ad59967 found that `sharedPaths` fills `choice.shared` in `node.go`, which `carriedByAll` reads to accept or refuse every path through a choice, so `List`'s file decides which paths compile; `join` serves `list.go`, `data.go`, `graph.go` and `path.go`.

### 147. Rename `namedNode`, `compiledPath.level` and `checkCells`, so "named" means bound to a name, "level" means only a prefix of a read's path, and `checkCells`' new name says it compiles.

`namedNode` in `graph.go` is a child and its segment, beside `namedRead` and `namedPick`, which go through a bound name; `compiledPath.level` is an error's label, where a level is a prefix of a read's path; `checkCells` compiles each cell and stores it in `cellTemplates`. Three seats of the 2026-10-07 panel and scoring run on ad59967 named the first two, and four named `checkCells`.

### 151. Load a caller's path and walk it in one `Generator` method, so no entry point walks the root before the shipped category the path names has loaded.

`Fake` and `FakeRecord` each call `loadShippedAt` before they walk, and its doc asks the next entry point to do the same. The greenfield architect of the 2026-10-07 panel on ad59967 named it the one corner a new senior would break without warning. A test passing `WithDataPath` would not catch it, because then `New` loads every category.

### 149. Say in one line what `config.load` does and what the 64-bit assertion asserts, beside their decision links.

Five seats of the 2026-10-07 panel and scoring run on ad59967 met a bare decision link where a definition belonged: `config.load` in `fejkdata.go` has only the link as its doc, and `var _ [^uint(0)>>63 - 1]struct{}` only a link beside it.

### 150. Move `columnRead` and `columnReadOf` from `resolve.go` into `datatype.go`, and state at `template.readsColumn` that `renderLeaf`, `proveColumnItem` and `itemDatatype` change with it.

Two seats of the 2026-10-07 scoring run on ad59967 found `readsColumn` changing what `renderLeaf`, `proveColumnItem` and `itemDatatype` do, and found that by grep; `columnReadOf`, which sets it, sits in `resolve.go`, apart from all three.

### 152. Delete the doc lines that restate their code.

Named by seats of the 2026-10-07 panel and scoring run on ad59967: `Children` in `internal/rows/rows.go`, `spelled` in `name.go`, the line in `graph.go` saying a folder has no edges, the Dockerfile's "Module layer cached separately from source." and `compose.yaml`'s "Interactive shell for ad-hoc work".

### 58. Ship prebuilt binaries, a container image, and packages for Homebrew, Scoop and the biggest Linux package managers, so the CLI needs no Go.

GoReleaser attaches the binaries to the release the tag workflow publishes, and builds deb and rpm packages, an Alpine APK and an AUR package beside Homebrew and Scoop. A checkout build prints `devel` for `--version`; the binaries carry the stamped tag. v0.1.0 ships first, so it can be tested before it is packaged.

### 113. Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.

`iban` in `internal/builtinfunc/checksum.go` draws a random numeric BBAN and sets only the mod-97 digits. BE's last two digits are the first ten mod 97, NO carries a mod-11 account check digit, FI a Luhn digit, ES two control digits, and SE's three-digit bank id must be a real bank's (`docs/research/research-sources-se.md`). Goal 4 asks for the check the real system applies; the README's `{iban(CC)}` row promises only length and mod 97, and changes with it.

### 154. Add `en_US.routing`, an ABA routing number over the Fed's prefix ranges with its check digit drawn by `{aba()}`.

Needs item 34.

### 155. Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.

Needs item 34: `{luhn()}` skips letters, so it draws a wrong ISIN check digit, and CUSIP's check digit is not Luhn.

### 26. Spell `sv_SE.personnummer` with the `+` Skatteverket uses from the year the holder turns 100, judged by a date the caller gives.

The format hard-codes `-`, and its earliest birthdate, 1930-01-01, makes the oldest draws invalid from 2030. A system holding Swedish personnummer often breaks on the `+`, so a fixture carrying one is a valuable test. Whether a number takes `+` depends on the date it is read, and the decision "No builtin reads the clock, so a date is bounded by days, never by an age" forbids a builtin from reading that date off the clock. This item revises it: once the caller gives the date, an age is a bound, and the CLI's default date makes output change with the day it runs unless the flag is given. Do the first step before the rest:
- Explore how a caller passes input data, such as the date a number is read on, to a render: in the template, as a plain option, or both. Make the library require the date, as item 89 makes it require the seed, and let the CLI default it to the system's when its flag is absent. Bench the spelling; `--now` reads like the clock. Shape the input so it can later narrow the draws to adults, children, 65 and over, or a mix, and so let a test ask for a centenarian; narrowing itself waits for its own item.
- Draw birthdates back past 100 years before the given date, so a `+` can appear. Read in 2026, the oldest shipped birthdate is 96 years old, so no draw carries a `+` until 2030. Skatteverket's test series reaches back to 1890, with birth numbers 980/981 before 1900. Draw no birthdate after the given date.
- Weight the ages by the territory's population, so the share of `+` numbers follows its demography: [research-age-bands.md](docs/research/research-age-bands.md).
- Let a data author's own category spell the separator too, through a builtin or a function a data module provides (item 77), and update the README's personnummer recipe and its "a day in 1930–2025" count.
- Rewrite the README's `--seed` and `WithSeed` rows to goal 10's wording, and name the new required option in the changelog: `New` breaks a second time here, after item 89, as the maintainer accepted.
- Research: `sv_SE.samordningsnummer` also hard-codes `-` and draws from 1930. Check whether Lag (2022:1697) om samordningsnummer gives it the same `+`; folkbokföringslagen 18 § states the rule for personnummer only.

### 34. Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.

`isin`: Luhn over letters expanded to digits. `aba`: 3-7-1 weights. `vin`: position 9 over the whole; a sample taking the WMI, since the check sits mid-string.

### 81. Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.

The decision "A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`" lets `{calc(a/(b*c))}` with `b` fixed at `"0"` write `Inf` into every row. An operand that is sometimes not a number prints `NaN`. Both work against goals 6.1 and 4. `valueproof.go` already proves a typed column's divisor nonzero. Once every operand is proven, `neverNumeric` and `constantValue` in `calc.go` go, and with them their reading of `NaN` and `Inf` as numbers, which `proven.Literal` refuses: calc reads text as a number one way.

### 91. Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.

A calc operand parses as a bare identifier, so `{calc(c.lat * 60)}` fails with `unexpected ".lat * 60"`, and a second binding of `/misc.coordinate.lat` would draw another coordinate. Resolve the path through `compileArm`'s leaves, as `operandNodes` does for a bare name.

### 36. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 88. Add a 12-digit `YYYYMMDDNNNC` form of `sv_SE.personnummer` and `sv_SE.samordningsnummer`, the form Skatteverket's testpersonnummer series and many systems store.

A system that stores the 12-digit form never meets the `+`, so it needs a fixture in that form.

### 41. Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.

`geo.SE.municipality[Lund]` and `misc.territory[Sweden]` select a row today. Steps:
- Have a selector that misses name the key of every row whose cell spells it, matched without case, else say the table selects by its key: `misc.territory[Sweden]` and `misc.territory[se]` name `[SE]`, `misc.protocol[tcp]` its key, and `misc.loglevel[Error]` `[err]`.
- Delete the `name` option and the load checks it carries, and have `New` refuse a table still carrying `name`, naming `key`.
- Key `misc.loglevel` on `keyword`, and `geo.SE.locality` on the integer `data-import/geo-se.py` assigns as the decision states, linking `street` and `postal-code` by it.
- Drop `key` from `last-name` and `sv_SE.title`, and re-pin `testdata/shipped_shape.txt`.
- Leave no selection by name in README.md, CHANGELOG.md, docs/ or `--help`.

### 52. Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.

With several `--data-path` layers the author has to work out which directory won, while a parse error on the same file names it.

### 117. List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.

Today the refusal names only the missing field, so a data author fixing a typo opens the category to find the spelling. Item 92 covers the hyphenated `calc` operand only.

### 28. Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.

Today `personnummer` draws its own sex, so a name and a number agree only through the README Data section's template, which reads both through one bound `sex`.

### 24. Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.

ISO 3166 reserves it for Western Sahara and the root zone has never been delegated it, so no resolver answers for it.

### 92. Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.

A hyphen is always subtraction in a calc, and the shipped data has hyphenated fields (`postal-code` and `street-number` in `geo/SE/address`). The refusal names the first fragment, not the field the author meant.

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

### 29. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 42. Refuse a struct tag's `datatype` naming the Go type that sets it before proving its values, as the README's Library section promises.

`datatype: boolean` over `{int(1,2)}` on an `int` field answers `prints an integer, not a boolean`.

### 159. Name the quoted spelling where a template that is a bare JSON number or boolean is refused, padded or not: `42` names `"42"`.

On stdin `42` is refused as "a number", which does not name the text the author meant. The goals audit of the stdin change found it.

### 22. Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.

17 of the 40 distinct ones shipped today sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost` provably reach nothing.

### 46. Report the same error every load for a table with two bad options, and for a folder with two unnamed rows files.

`readTableOptions` in `table.go` and `walkDir` in `internal/datafiles` return on the first in Go's map order.

### 158. Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.

A file a Windows editor saves with a BOM prints an invisible U+FEFF first, and a JSON template with one is read as a format string. PowerShell 5.1 encodes what it pipes to a program by `$OutputEncoding`, ASCII by default. The maintainer chose 0.2.0 on 2026-10-07.

### 79. Beside a `--data-path`, load in `New` the user's categories, the shipped ones they override or read, and the shipped ones whose reads reach an overridden one; load the rest on first reach.

Needs item 77, which makes an override fail unless it says it replaces.

The decision "With only the shipped set, a category loads on the first call reaching it; beside a `--data-path`, every category loads in `New`" parses the whole shipped set on every `--data-path` run, about 45 ms today and seconds once items 61 and 78 land. Goal 6 needs only the user's categories, the shipped ones they override, the shipped ones they read, and the shipped ones whose reads reach an overridden category, so the index would carry each category's reads.

### 107. Move the template engine into `internal/`, leaving the root the public API and its options.

The engine returns column values and the root builds `Record`. Revises the decision "The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`", and, if the vocabulary moves, "The vocabulary sits below `doc.go`'s package clause, not in the package doc".

### 112. Check a calc's field operands where its template resolves, beside its name operands, once item 36 reports every mistake at once.

Needs item 36. `checkCalcFields` runs at compile and `checkCalcNames` in `resolveTemplates`, two passes over one rule. Moving the field check there before item 36 would change which of two mistakes in two category files a load reports, since every category compiles before any resolves.

### 108. Name item 26 in the decision "No builtin reads the clock, so a date is bounded by days, never by an age", as the item that ends its conflict with goal 4.

Item 26 revises the decision, but the decision does not point back to it.

### 38. Compare `calcParser.binary`'s operator as a rune.

`byte(p.rs[p.pos])` reads U+012B `ī` as `+`, so `{calc(a ī b)}` compiles as `a + b`.

### 45. Name the node a selector follows in `childNamed`'s refusal.

`sv_SE.person[1].first` answers `1 is not a table`.

### 49. Report a CRLF rows file holding only its header as having no rows.

`parseRows` answers "line 2 is empty".

### 13. Place xlsx cells by their `r` reference in `data-import/xlsx.py`.

A sheet omitting an empty cell shifts every later column left.

### 39. Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.

An `int64` or `float64` field unproven in range is told to become itself, and a `uint64` one to become `int64`.

### 51. Spell a table one way across the errors that name it.

In `internal/rows`, `Pins.PinRow` names `t.path` and `Pins.MustRow` panics with `t.segment`, so one table is `territory` and `misc.territory`, and the short spelling names no file where two folders hold that name.

### 115. Spell a rows file's line one way in every error, naming the file.

Today a cell reads `g.t.b, line 3` (`namedNode.labelIn`), a cell check `t.tsv: line 3, b` (`checkCells`) and a link `t.tsv line 3` (`rows.Table.Link`). A reference cycle found from a table's format through a cell names no line at all: `reference cycle: g.t -> format -> b -> b -> /h -> …`, where the column-to-cell step in `renderEdges` repeats the column name.

### 14. Spell `misc.creditcard`'s digit runs `{digits(n)}`.

A `"repeat": 14` over `{digits(1)}` draws every 14-digit run with the odds `{digits(14)}` does. Both load, since neither has a part that changes nothing; the shipped data writes the obvious one. Re-pin seeded output in its own commit.

### 11. Test that every `geo` tree holds the five table names and the `address` columns that port across countries.

Nothing checks it, and a new country breaks it silently.

### 40. Parse `defaultTable`'s path with the library's grammar.

The CLI counts brackets of its own, so a change to the selector grammar desyncs the default `--table` silently.

### 86. Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.

Both phone formats carry the same 21-code `area` list, so an edit to one copy can miss the other.

### 23. Give every other shipped category that could reach something real a path that never does.

`phone` draws live PTS and NANP ranges — PTS's five fiction series are the Swedish inert set, sourced in [research-sources-se.md](docs/research/research-sources-se.md), and the NANP one still wants a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.

### 75. Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.

`readme_test.go` loads and renders each `json` block, but runs no `sh` example (CLI, Records, Table, Linked tables), no `go` example (Library) and no `text` block of error text, and compares no "Renders e.g." output, so goal 12.2 is not met.

### 47. Refuse a category naming a hidden rows file, `"rows": ".x.tsv"`, as a rows file that is not there.

`internal/datafiles` indexes hidden TSVs in `newRowsFiles` and skips them in `walkDir`, so one loads where its comment says a hidden file is never data.

### 72. Fail the merge gate when test coverage falls below the last recorded figure.

`technical-principles.md` holds that test coverage should not decline. `docker compose run --rm cover` prints the figure, and neither the `Dockerfile` gate nor `.github/workflows/test.yml` checks it.

### 95. Hold a struct's columns and the field each fills in one slice.

`structShape.fieldIndexes` pairs with its record template's `columns` by index alone, against the technical principle "one owner per value".

### 116. Turn a row index into its rows file's line in one function.

Today `namedNode`'s cell line in `graph.go`, `table.checkCells` and each site in `internal/rows` add the header line and the one-based count.

### 56. Build on a manual run of `test.yml`.

`workflow_dispatch` leaves `github.event.before` empty, so the diff compares `HEAD` with itself and skips `docker build`.

### 109. Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.

Technical principle: indent with hard tabs. All 19 scripts indent with four spaces, so an edit there either keeps the wrong style or rewrites the whole file.

### 63. Pair a street with its exact postnummer.

Needs an application to Lantmäteriet; today a street goes to the nearest postal code centroid. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 62. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 59. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 65. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go.

### 66. Cut the README's Layout block to the lines that say what a file name cannot.

`doc.go` and `cmd/fejkdata/` restate their names.

### 67. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens by restating the decision "GitHub is canonical, and the module path names it", a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 64. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 68. Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.

Gitea converts a repository to a mirror only by re-creating it, so until that runs the copy there is a second owner of one history.

### 69. Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.

The human cuts v1.0.0 once the shipped data is in its record shape and one full minor has shipped with no breaking change, per the README's Versioning table; pairing a street with its exact postnummer rewrites shipped rows, so a minor with no breaking change follows it first.
