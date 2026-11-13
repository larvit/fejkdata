# todo

## Scoring

`Score = 100 * V / (2**S + 8)`

`Bar = 0`

`Next ID = 457`

## Items

| ID | Release | Blocked by | Exempt | Item | L | E | I | R | S | Moves | Score |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 455 | 0.5.0 |  |  | **Run mypy at full strictness over data-import, release-tooling and vulnerability-filing in the gate, and fix every error it reports.** | 3 | 1 | 1 | 3 | 7 | code health by 3 | 0.9 |
| 456 | 0.5.0 |  |  | **Gate cyclomatic complexity at 10 in both gocyclo lines, with every function over it split or listed, with its reason, as a table-shaped dispatch.** | 3 | 1 | 1 | 3 | 8 | code health by 5 | 0.7 |
| 205 | 0.6.0 |  |  | **Add `is_IS` with its given names by sex, last names and titles, and `is_IS.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 327 | 0.6.0 | 205 |  | **Add `is_IS`'s national personal id, the kennitala, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 453 | 0.6.0 |  | owner | **Reach the comprehension floor: lift Locality so a name rule asks a route what it keeps, draws and addresses, and no rule slices `levels`, a tail or `bound` itself.** | 4 | 2 | 2 | 6 | 6 | code health by 6 | 3.3 |
| 326 | 0.6.0 | 205 |  | **Add `is_IS.phone`, drawn from Iceland's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 328 | 0.6.0 | 205 |  | **Add `is_IS.company`, `is_IS.email` and `is_IS.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 451 | 0.6.0 |  |  | **Add a `fi_FI` henkilötunnus variant over DVV's issued individual numbers 002 to 899, beside `fi_FI.henkilotunnus`, which draws 900 to 999.** | 2 | 1 | 2 | 3 | 3 | 1 by 1, 3 by 3 | 2.1 |
| 324 | 0.6.0 | 205 |  | **Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 325 | 0.6.0 | 324 |  | **Give `geo/IS`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.5 |
| 334 | 0.7.0 |  |  | **Add `nn_NO`, written in Nynorsk, reading `nb_NO`'s names and id and `geo/NO`, with month and weekday names from CLDR and titles of its own.** | 2 | 2 | 1 | 3 | 2 | 1 by 1, 2.1 by 2 | 2.2 |
| 169 | 0.8.0 |  | question | **Reword the decision "A parent row with no child row is a load error" to say what each geo import does with such a row: drop it, or exit where `geo-se.py` would lose a kommun or a län.** | 1 | 2 | 1 | 2 | 1 | code health by 2 | 7.8 |
| 351 | 0.8.0 |  | question | **Name in the decision "A known vulnerability in the Go version or a required module blocks a release, even in code fejkdata never calls, and blocks no merge" each line its escape hatch edits, so dropping the scan from the release job releases instead of skipping the job.** | 1 | 2 | 2 | 3 | 1 | 1.1 by 2 | 1.8 |
| 452 | 0.8.0 | 453 | defect | **Refuse at `New` a read of a name whose render reaches a table cell that draws afresh a path another read of the name reads from the pick; today the name fence walks no cell.** | 2 | 2 | 2 | 4 | 3 | 3.1 by 2, 4.6 by 2 | 12.5 |
| 343 | 0.8.0 |  | defect | **Split the README's Data section, about 225 lines under one heading, into sub-headings, one per feature, so each feature has its own heading (goal 11.1).** | 2 | 2 | 1 | 3 | 3 | 11.1 by 3 | 9.4 |
| 354 | 0.8.0 |  |  | **Make `test_main_hands_the_cut_tag_to_the_workflow` pass across UTC midnight, by comparing the changelog's date with the one `patch_release.main` read.** | 1 | 1 | 1 | 1 | 1 | code health by 2 | 7.8 |
| 167 | 0.8.0 |  |  | **Lower the reading cost of the files a chunk reads most before its first commit.** | 1 | 2 | 1 | 2 | 3 | code health by 3 | 7.3 |
| 195 | 0.8.0 |  | principle | **Parse each builtin argument once, in `checkArgs`, and hand `prep` and `proveNumber` the parsed value.** | 2 | 1 | 2 | 3 | 3 | code health by 3 | 7.3 |
| 40 | 0.8.0 |  |  | **Parse `defaultTable`'s path with the library's grammar.** | 1 | 1 | 2 | 2 | 2 | code health by 2 | 6.5 |
| 95 | 0.8.0 |  | principle | **Hold a struct's columns and the field each fills in one slice.** | 1 | 1 | 2 | 2 | 2 | code health by 2 | 6.5 |
| 116 | 0.8.0 |  |  | **Turn a row index into its rows file's line in one function.** | 1 | 1 | 1 | 1 | 2 | code health by 2 | 6.5 |
| 350 | 0.8.0 |  | decision | **Bring the release path's links down until paired rulings against f90b4d6 and 5aca8a1 no longer find the project reads worse.** | 3 | 2 | 3 | 6 | 5 | code health by 5 | 4.9 |
| 34 | 0.8.0 |  |  | **Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 3 by 4 | 1.8 |
| 155 | 0.8.0 | 34 |  | **Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2 by 2, 3 by 3 | 4.3 |
| 27 | 0.8.0 |  |  | **Merge `en_US.ip` and `sv_SE.ip` into one `misc.ip`, and `en_US.version` and `sv_SE.version` into one `misc.version`, each pair today byte-identical.** | 2 | 1 | 2 | 3 | 2 | 2 by 1 | 0.7 |
| 15 | 0.8.0 | 27 |  | **Add `misc.ip`'s documentation, private and CIDR forms, `misc.ipv6` in `2001:db8::/32`, and a locally administered `misc.mac`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 3.3 by 4, 2 by 1 | 4.1 |
| 22 | 0.8.0 |  |  | **Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 3.3 by 6 | 4.1 |
| 252 | 0.8.0 |  |  | **Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.** | 1 | 3 | 1 | 3 | 1 | 1.1 by 4 | 3.7 |
| 256 | 0.8.0 |  |  | **Add `misc.bic`, built by its structure rules.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2, 3 by 2 | 3.6 |
| 254 | 0.8.0 |  |  | **Say in the docs which of three kinds each category's values are: real, valid or invented.** | 2 | 2 | 1 | 3 | 4 | 1 by 2, 3 by 8 | 3.5 |
| 220 | 0.8.0 |  |  | **Suggest the nearest name where a path segment names nothing: `no entry "adress"; did you mean "address"?`.** | 1 | 1 | 1 | 1 | 3 | 1 by 2, 5 by 5 | 3.4 |
| 24 | 0.8.0 |  | defect | **Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 3 by 3 | 3.4 |
| 30 | 0.8.0 |  |  | **Draw `email.local` from `username`, so the two share one handle list.** | 1 | 1 | 1 | 1 | 2 | code health by 1 | 3.3 |
| 255 | 0.8.0 |  |  | **Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.** | 2 | 1 | 2 | 3 | 3 | 1 by 1, 3 by 3, 2 by 2 | 3.2 |
| 117 | 0.8.0 |  |  | **List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 4 | 3.2 |
| 168 | 0.8.0 |  | principle | **Test in the merge gate the shared steps of `data-import/`, starting with `geo.fence`, the småort steps of `geo-se.py`, and the name-shape rules of `company-no.py` and `company_names.py`.** | 3 | 1 | 1 | 3 | 5 | code health by 3 | 2.9 |
| 257 | 0.8.0 |  |  | **Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 2 by 3, 3 by 3 | 2.9 |
| 389 | 0.8.0 |  |  | **Add `misc.rrule`, recurrence rules a test checks against RFC 5545's grammar.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2 by 1, 3 by 2 | 2.9 |
| 113 | 0.8.0 |  | defect | **Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 3 by 6 | 2.8 |
| 406 | 0.8.0 |  |  | **Add a transform that writes a value escaped for a JSON string, so a text template spelling JSON stays valid whatever a drawn value holds.** | 2 | 1 | 1 | 2 | 2 | 1 by 2, 4 by 2 | 2.8 |
| 260 | 0.8.0 |  |  | **Add `misc.element`, `misc.unit` and `misc.planet`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 2 | 2.7 |
| 261 | 0.8.0 |  |  | **Add `misc.animal` by class, and `misc.plant`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 2 | 2.7 |
| 164 | 0.8.0 | 453 |  | **Let a table's format or a cell bind a name, which lives in that format or cell and keeps one pick through the draw.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 4 by 3 | 2.7 |
| 262 | 0.8.0 |  |  | **Add `misc.bloodtype`, weighted by its global distribution.** | 1 | 2 | 1 | 2 | 1 | 1 by 1, 2 by 1, 3.2 by 1 | 2.6 |
| 50 | 0.8.0 |  |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 52 | 0.8.0 |  |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 92 | 0.8.0 |  |  | **Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 219 | 0.8.0 |  |  | **Name the path and the Go type where a struct field's type cannot hold what its tag renders: `sv_SE.person renders text, which int cannot hold`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 221 | 0.8.0 |  |  | **Refuse two references under a record `--format` by saying a run renders one record: today `{/users}{/orders}` reports that an inline record has no fields.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 224 | 0.8.0 |  |  | **Name the module in every error about its manifest or its files: today an `fs.FS` module's errors name only the file inside it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 228 | 0.8.0 |  |  | **Offer `{/path as p}` where a token starts with a name nothing binds and no loaded path starts with, in a record's format too: today `{p.first}` with `p` unbound suggests `{/p.first}`.** | 2 | 1 | 1 | 2 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 229 | 0.8.0 |  |  | **Name the line and column where JSON parsing stopped, when a template that opens like a JSON object then fails as a format string.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.6 |
| 38 | 0.8.0 |  | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 2 | 2 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 44 | 0.8.0 |  | decision | **Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 45 | 0.8.0 |  |  | **Name the node a selector follows in `childNamed`'s refusal.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 49 | 0.8.0 |  |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 110 | 0.8.0 |  |  | **Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 225 | 0.8.0 |  |  | **Drop "a field cannot be named weight" from the refusal of a `weight` that is no number, and name the column a table's `weight` reads.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 2.4 |
| 355 | 0.8.0 |  |  | **Run `docker build .` and the Packages dry run on a release commit.** | 2 | 2 | 2 | 4 | 2 | 1.1 by 3 | 2.3 |
| 181 | 0.8.0 |  | decision | **Fail a load where two sources define one category, unless the library's option or the CLI's flag says the later replaces the earlier, and then make every read reach the replacement.** | 4 | 1 | 3 | 6 | 4 | 1 by 2, 5 by 5 | 2.3 |
| 81 | 0.8.0 |  | decision | **Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.** | 3 | 1 | 3 | 5 | 4 | 1 by 1, 5 by 4, 3 by 2 | 2.3 |
| 35 | 0.8.0 |  |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 2.2 |
| 265 | 0.8.0 |  |  | **Add `misc.commit` messages, branch names and `sha`s.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 2.2 |
| 258 | 0.8.0 |  |  | **Add `misc.product` names, adjectives, materials and departments, and `misc.garmentsize`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 3 | 2.2 |
| 266 | 0.8.0 |  |  | **Add `misc.password`, `hash`, `jwt`, `slug`, `hostname`, `avatar` and `placeholderimage`.** | 1 | 1 | 1 | 1 | 3 | 1 by 1, 2 by 3 | 2.2 |
| 196 | 0.8.0 |  | decision | **Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 5 by 6 | 2.2 |
| 263 | 0.8.0 |  |  | **Add `misc.color` names with their hex, rgb, hsl and cmyk forms.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 1 | 2.2 |
| 28 | 0.8.0 |  | decision | **Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 2.1 by 3, 3 by 2 | 2.1 |
| 405 | 0.8.0 |  |  | **Let `repeat` take a range, such as 5 to 15, drawing the count on each render.** | 2 | 1 | 1 | 2 | 3 | 1 by 2, 4 by 2 | 2.1 |
| 13 | 0.8.0 |  | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 3.2 by 2 | 2.1 |
| 165 | 0.8.0 |  |  | **Let a table carry `name` with neither `key` nor `parent`, and list the rows a name matches by their line where it matches several.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 2 | 2.0 |
| 166 | 0.8.0 |  |  | **Let a key or name cell hold a bracket, a brace, a double quote or a pipe, so the table loads, its row draws at random, and only a selector cannot reach it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 2 | 2.0 |
| 39 | 0.8.0 |  |  | **Stop the `fence.FieldKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 2.0 |
| 42 | 0.8.0 |  |  | **Refuse a struct tag's `datatype` other than its Go type's before proving its values, as the README's Library section promises.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 2.0 |
| 46 | 0.8.0 |  |  | **Report the same error every load for a table with two bad options.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 2.0 |
| 51 | 0.8.0 |  |  | **Spell a table one way across the errors that name it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 2.0 |
| 115 | 0.8.0 |  |  | **Spell a rows file's line one way in every error, naming the file.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 2.0 |
| 408 | 0.8.0 |  |  | **List every shipped path in the README's Data section with an example value and whether it is real, plausible or valid, and where it comes from.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 3 by 4, 11 by 4 | 2.0 |
| 259 | 0.8.0 |  |  | **Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 2 by 4, 3.2 by 3 | 1.9 |
| 376 | 0.8.0 |  |  | **Let a template ask for a `sv_SE.personnummer` of a child, an adult, someone of 65 or more or a centenarian, and for a mix of them.** | 2 | 2 | 1 | 3 | 4 | 1 by 3, 3 by 2 | 1.9 |
| 267 | 0.8.0 |  |  | **Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.** | 1 | 2 | 1 | 2 | 4 | 1 by 1, 2 by 2, 3.2 by 2 | 1.8 |
| 32 | 0.8.0 |  |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 2 | 1 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.7 |
| 264 | 0.8.0 |  |  | **Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.7 |
| 268 | 0.8.0 |  |  | **Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.7 |
| 387 | 0.8.0 |  |  | **Let a template draw a number from a bounded normal distribution, so durations and start times cluster without weighted ranges.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 4 by 1 | 1.7 |
| 454 | 0.8.0 |  |  | **Give `geo.FI`'s postitoimipaikat and kunnat their Swedish names, such as Helsingfors and Åbo.** | 2 | 2 | 2 | 4 | 3 | 1 by 1, 2.1 by 2 | 1.6 |
| 91 | 0.8.0 |  |  | **Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 4 by 4 | 1.6 |
| 109 | 0.8.0 |  | principle | **Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.** | 1 | 2 | 1 | 2 | 4 | code health by 1 | 1.6 |
| 23 | 0.8.0 |  |  | **Give every other shipped category that could reach something real a path that never does.** | 2 | 2 | 2 | 4 | 5 | 1 by 2, 3.3 by 6 | 1.6 |
| 194 | 0.8.0 |  | decision | **Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 5 by 4 | 1.6 |
| 158 | 0.8.0 |  |  | **Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.** | 1 | 1 | 1 | 1 | 2 | 1 by 2 | 1.6 |
| 163 | 0.8.0 |  |  | **Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.** | 2 | 1 | 2 | 3 | 2 | 1 by 2 | 1.6 |
| 269 | 0.8.0 |  |  | **Add `misc.direction`, `misc.continent` and `misc.ulid`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 1 | 1.5 |
| 161 | 0.8.0 |  |  | **Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 1 | 1.4 |
| 18 | 0.8.0 |  |  | **Give the address records one column set across countries: `region` and `municipality` as columns on `geo.SE.address` too.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 4 by 3 | 1.3 |
| 29 | 0.8.0 |  |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 4 by 3 | 1.3 |
| 36 | 0.8.0 |  |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 3 | 1 | 2 | 4 | 6 | 1 by 3, 5 by 7 | 1.1 |
| 56 | 0.8.0 |  | defect | **Build on a manual run of `test.yml` on `main`.** | 1 | 2 | 1 | 2 | 1 | 1.1 by 1 | 0.9 |
| 53 | 0.8.0 |  |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 3 | 1 | 2 | 4 | 5 | 1 by 2, 4 by 2 | 0.8 |
| 404 | 0.8.0 |  |  | **Let a template in a `repeat` or a choice item bind a name to a path through an enclosing name, `{l.street as s}`, so each render draws a fresh row inside the enclosing pick.** | 3 | 2 | 2 | 5 | 6 | 1 by 3, 4 by 4 | 0.8 |
| 75 | 0.8.0 |  |  | **Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.** | 2 | 1 | 1 | 2 | 5 | 1 by 1, 4 by 3 | 0.8 |
| 55 | 0.8.0 |  |  | **Document `ErrNoColumns` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 1 | 2 | 2 | 1 by 1 | 0.8 |
| 17 | 0.8.0 |  |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 3 | 1 | 3 | 5 | 6 | 1 by 1, 4 by 3, 2 by 2 | 0.7 |
| 14 | 0.8.0 |  |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`.** | 1 | 1 | 1 | 1 | 2 | 4 by 1 | 0.6 |
| 33 | 0.8.0 |  |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 3 | 1 | 3 | 5 | 5 | 1 by 1, 4 by 2 | 0.6 |
| 193 | 0.8.0 |  | decision | **Move the largest shipped packages into nested Go modules that keep their import paths, once `TestShippedDataStaysUnder100MiB` fails.** | 4 | 2 | 3 | 7 | 6 | 1 by 1, 2 by 3 | 0.5 |
| 393 | 0.8.0 |  |  | **Write several linked templates in one run with a repeated `--emit <template>=<count>`, parents first, each template's rows contiguous, and `{/run.<template> as c}` reading another template's rows.** | 4 | 2 | 2 | 6 | 7 | 1 by 6 | 0.4 |
| 407 | 0.8.0 |  |  | **Let a choice deal its items in exact counts across a run's renders, so 200 issues hold exactly 4 in progress, where weights only approach it.** | 3 | 2 | 2 | 5 | 6 | 1 by 3 | 0.4 |
| 218 | 0.8.0 |  | decision | **Give every country shipped without streets its streets, once the maintainer accepts ODbL data or the country opens a register.** | 3 | 2 | 2 | 5 | 7 | 1 by 1, 2.1 by 4 | 0.3 |
| 129 | 0.8.0 |  | decision | **Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.** | 4 | 1 | 3 | 6 | 8 | 1 by 3, 4 by 7 | 0.3 |
| 41 | 0.8.0 |  | decision | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 4 | 1 | 3 | 6 | 7 | 1 by 1, 4 by 4 | 0.3 |
| 186 | 0.8.0 |  | decision | **Load a module carrying functions through an option of its own, its functions reaching the generator's randomness and date through an interface the root declares.** | 4 | 1 | 3 | 6 | 7 | 1 by 3 | 0.2 |
| 66 | 0.9.0 |  |  | **Cut the README's Layout block to the lines that say what a file name cannot.** | 1 | 2 | 1 | 2 | 1 | code health by 2 | 7.8 |
| 67 | 0.9.0 |  |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 2 | 2 | code health by 2 | 6.5 |
| 64 | 0.9.0 |  |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 2 | 1 | 2 | 4 | code health by 2 | 3.3 |
| 62 | 0.9.0 |  |  | **Add `{btc()}` and `{eth()}`.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 2 by 2, 3 by 2 | 1.8 |
| 59 | 0.9.0 |  |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 2 | 1 | 2 | 1 | 1 by 1 | 1.0 |
| 63 | 0.9.0 |  |  | **Pair a street with its exact postnummer.** | 3 | 2 | 3 | 6 | 6 | 1 by 2, 3 by 6 | 0.9 |
| 68 | 0.9.0 |  | decision | **Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.** | 1 | 3 | 2 | 4 | 1 | 1.1 by 1 | 0.9 |
| 65 | 0.9.0 |  |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 2 | 1 | 3 | 3 | 1 by 1 | 0.6 |
| 348 | 1.0.0 |  | question | **Decide how a Go minor that drops a system the binaries run on, such as an older macOS, ships from 1.0.0 on, where dropping a system is a major, and record it in the decision "CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag".** | 1 | 2 | 1 | 2 | 1 | 1.1 by 3 | 2.8 |
| 69 | 1.0.0 |  |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 2 | 1 | 2 | 3 | 1 by 8 | 4.8 |
| 70 | 1.0.0 |  |  | **Publish a homepage with an in-browser generator, the library compiled to WebAssembly.** | 3 | 2 | 2 | 5 | 6 | 1 by 6 | 0.8 |
| 396 | 1.1.0 |  | question | **Decide whether `fejkdata` writes Go struct types, and TypeScript types for the npm package, from a record's columns.** | 1 | 1 | 1 | 1 | 1 | 1 by 2 | 1.9 |
| 190 | 1.1.0 |  | question | **Decide whether fejkdata pseudonymises real datasets, and if so add the goal, the audience persona and the items below to the README and the plan.** | 4 | 2 | 3 | 7 | 9 | 1 by 4, 3 by 3 | 0.1 |
| 357 | 1.2.0 |  | question | **Create the `@larvit` organisation on npm, and make this repository's Packages workflow the trusted publisher of `@larvit/fejkdata`.** | 1 | 2 | 1 | 2 | 1 | 1.1 by 1 | 0.9 |
| 358 | 1.2.0 |  |  | **Record in `docs/decisions.md` the JavaScript API that a panel of JavaScript and TypeScript test authors picks for `@larvit/fejkdata`.** | 1 | 2 | 1 | 2 | 3 | 1.2 by 3 | 1.7 |
| 359 | 1.2.0 | 358 |  | **Build the library to WebAssembly behind a bridge that reads only its public API, and export it from `@larvit/fejkdata` as item 358's decision spells it.** | 3 | 2 | 2 | 5 | 6 | 1.2 by 8 | 1.0 |
| 360 | 1.2.0 | 359 |  | **Ship TypeScript declarations with `@larvit/fejkdata`, and fail the gate where a typed usage test stops compiling under `tsc --strict`.** | 2 | 1 | 1 | 2 | 3 | 1.2 by 5 | 2.8 |
| 362 | 1.2.0 | 357, 359 |  | **Publish `@larvit/fejkdata` to npm, with provenance, for every release CI cuts, patches included, and pack it in the Packages dry run.** | 3 | 3 | 3 | 7 | 5 | 1.1 by 4, 1.2 by 6 | 2.3 |
| 361 | 1.2.0 | 359 |  | **Test `@larvit/fejkdata` in the gate on Node, Bun, Deno and a headless browser, and pin that its seeded renders match the Go library's byte for byte.** | 3 | 1 | 2 | 4 | 5 | 1.2 by 4 | 0.9 |
| 363 | 1.2.0 | 360, 361 |  | **Document `@larvit/fejkdata` in the README: its install line, a JavaScript section beside Library, the runtimes it is tested on and its Versioning row, with every JavaScript example run as a test.** | 2 | 2 | 1 | 3 | 4 | 1.1 by 2, 1.2 by 4 | 2.3 |
| 345 | 1.3.0 |  | question | **Decide how a locale with names in a script such as Han, Devanagari, Arabic or Bengali gets valid emails, and record it in the `slugify` decision.** | 1 | 2 | 1 | 2 | 1 | 2.1 by 2, 3.3 by 3 | 4.0 |
| 86 | 1.3.0 |  |  | **Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.** | 1 | 1 | 1 | 1 | 2 | code health by 2 | 6.5 |
| 154 | 1.3.0 | 34 |  | **Add `en_US.routing`, an ABA routing number over the Fed's prefix ranges with its check digit drawn by `{aba()}`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2.1 by 3, 3 by 3 | 4.9 |
| 243 | 1.3.0 |  |  | **Add `en_US.ein`, drawing only the prefixes the IRS assigns.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 2.1 by 3, 3 by 3 | 4.9 |
| 210 | 1.3.0 |  |  | **Add `fr_FR` with its given names by sex, last names and titles, and `fr_FR.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 273 | 1.3.0 | 210 |  | **Add `fr_FR`'s national personal id, the NIR, the numéro de sécurité sociale, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 206 | 1.3.0 | 345 |  | **Add `zh_CN` with its given names by sex, last names and titles, and `zh_CN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 277 | 1.3.0 | 206 |  | **Add `zh_CN`'s national personal id, the 18-digit resident identity card number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 208 | 1.3.0 |  |  | **Add `es_MX` with its given names by sex, last names and titles, and `es_MX.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 282 | 1.3.0 | 208 |  | **Add `es_MX`'s national personal id, the CURP, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 212 | 1.3.0 |  |  | **Add `pt_BR` with its given names by sex, last names and titles, and `pt_BR.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 287 | 1.3.0 | 212 |  | **Add `pt_BR`'s national personal id, the CPF, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 207 | 1.3.0 | 345 |  | **Add `hi_IN` with its given names by sex, last names and titles, and `hi_IN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 306 | 1.3.0 | 207 |  | **Add `hi_IN`'s national personal id, the Aadhaar number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 209 | 1.3.0 | 345 | decision | **Add `ar_EG` with its given names by sex, last names and titles, and `ar_EG.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 310 | 1.3.0 | 209 |  | **Add `ar_EG`'s national personal id, the national ID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 211 | 1.3.0 | 345 |  | **Add `bn_BD` with its given names by sex, last names and titles, and `bn_BD.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 314 | 1.3.0 | 211 |  | **Add `bn_BD`'s national personal id, the NID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 213 | 1.3.0 |  |  | **Add `id_ID` with its given names by sex, last names and titles, and `id_ID.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 318 | 1.3.0 | 213 |  | **Add `id_ID`'s national personal id, the NIK, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 214 | 1.3.0 |  | decision | **Add `ur_PK` with its given names by sex, last names and titles, and `ur_PK.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 322 | 1.3.0 | 214 |  | **Add `ur_PK`'s national personal id, the CNIC number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 217 | 1.3.0 |  |  | **Add `de_DE` with its given names by sex, last names and titles, and `de_DE.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 3.3 |
| 332 | 1.3.0 | 217 |  | **Add `de_DE`'s national personal id, the Steuer-Identifikationsnummer, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 4.3 |
| 245 | 1.3.0 |  |  | **Draw `en_US.phone` from every area code NANPA has in service.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2, 3.2 by 3 | 4.2 |
| 250 | 1.3.0 |  |  | **Add `en_US.month` and `en_US.weekday` names from CLDR, and `en_US.holiday`, the federal holidays.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 4, 3.2 by 2 | 3.7 |
| 244 | 1.3.0 |  |  | **Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.** | 2 | 2 | 2 | 4 | 4 | 1 by 2, 2.1 by 4, 3.2 by 3 | 3.2 |
| 198 | 1.3.0 |  |  | **Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.** | 2 | 2 | 2 | 4 | 3 | 1 by 1, 2.1 by 2, 3.3 by 3 | 3.1 |
| 272 | 1.3.0 | 210 |  | **Add `fr_FR.phone`, drawn from France's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 276 | 1.3.0 | 206 |  | **Add `zh_CN.phone`, drawn from China's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 281 | 1.3.0 | 208 |  | **Add `es_MX.phone`, drawn from Mexico's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 286 | 1.3.0 | 212 |  | **Add `pt_BR.phone`, drawn from Brazil's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 305 | 1.3.0 | 207 |  | **Add `hi_IN.phone`, drawn from India's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 309 | 1.3.0 | 209 |  | **Add `ar_EG.phone`, drawn from Egypt's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 313 | 1.3.0 | 211 |  | **Add `bn_BD.phone`, drawn from Bangladesh's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 317 | 1.3.0 | 213 |  | **Add `id_ID.phone`, drawn from Indonesia's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 321 | 1.3.0 | 214 |  | **Add `ur_PK.phone`, drawn from Pakistan's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 331 | 1.3.0 | 217 |  | **Add `de_DE.phone`, drawn from Germany's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.9 |
| 248 | 1.3.0 |  |  | **Add `en_US.car`, a make and model from NHTSA vPIC.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 2, 3.2 by 2 | 2.6 |
| 246 | 1.3.0 |  |  | **Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2 | 2.2 |
| 274 | 1.3.0 | 210 |  | **Add `fr_FR.company`, `fr_FR.email` and `fr_FR.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 278 | 1.3.0 | 206, 345 |  | **Add `zh_CN.company`, `zh_CN.email` and `zh_CN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 283 | 1.3.0 | 208 |  | **Add `es_MX.company`, `es_MX.email` and `es_MX.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 288 | 1.3.0 | 212 |  | **Add `pt_BR.company`, `pt_BR.email` and `pt_BR.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 307 | 1.3.0 | 207, 345 |  | **Add `hi_IN.company`, `hi_IN.email` and `hi_IN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 311 | 1.3.0 | 209, 345 |  | **Add `ar_EG.company`, `ar_EG.email` and `ar_EG.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 315 | 1.3.0 | 211, 345 |  | **Add `bn_BD.company`, `bn_BD.email` and `bn_BD.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 319 | 1.3.0 | 213 |  | **Add `id_ID.company`, `id_ID.email` and `id_ID.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 323 | 1.3.0 | 214 |  | **Add `ur_PK.company`, `ur_PK.email` and `ur_PK.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 333 | 1.3.0 | 217 |  | **Add `de_DE.company`, `de_DE.email` and `de_DE.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 2.2 |
| 247 | 1.3.0 |  |  | **Add `en_US.licenseplate`, drawn per state from that state's pattern.** | 2 | 2 | 2 | 4 | 4 | 1 by 1, 2.1 by 3, 3 by 2 | 2.1 |
| 270 | 1.3.0 | 210 |  | **Add `geo/FR` with France's places and postal codes, and `fr_FR.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 275 | 1.3.0 | 206 |  | **Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 279 | 1.3.0 | 208 |  | **Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 284 | 1.3.0 | 212 |  | **Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 304 | 1.3.0 | 207 |  | **Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 308 | 1.3.0 | 209 |  | **Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 312 | 1.3.0 | 211 |  | **Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 316 | 1.3.0 | 213 |  | **Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 320 | 1.3.0 | 214 |  | **Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 329 | 1.3.0 | 217 |  | **Add `geo/DE` with Germany's places and postal codes, and `de_DE.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.5 |
| 271 | 1.3.0 | 270 |  | **Give `geo/FR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.5 |
| 280 | 1.3.0 | 279 |  | **Give `geo/MX`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.5 |
| 285 | 1.3.0 | 284 |  | **Give `geo/BR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.5 |
| 330 | 1.3.0 | 329 |  | **Give `geo/DE`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.5 |
| 432 | 2.0.0 |  | question | **Reword goals 2.2, 7, 8.1, 8.5, 8.6, 9.2 and 10, and the README's Versioning table, for a Rust core that a native Go library and the npm package both serve.** | 1 | 2 | 1 | 2 | 2 | 1 by 1 | 0.8 |
| 433 | 2.0.0 |  |  | **Measure a seeded render of the README's person.json, one row and 100,000 rows, from Rust compiled to WebAssembly under wazero against the native Go library, and record whether wazero comes within a few percent.** | 2 | 1 | 1 | 2 | 2 | 1 by 1 | 0.8 |
| 434 | 2.0.0 | 432, 433 |  | **Add a Cargo workspace with a compose service, and a gate leg running rustfmt, clippy with a complexity limit of 14, the tests, and cargo-mutants for the stability pass.** | 3 | 1 | 1 | 3 | 4 | 1 by 1 | 0.4 |
| 435 | 2.0.0 | 434 |  | **Port internal/drawstate to Rust so a seed draws the same numbers as Go, and fail the gate where the Go and Rust CLIs print different bytes for a corpus of templates at a fixed seed and date.** | 3 | 1 | 2 | 4 | 4 | 9 by 5 | 10.4 |
| 436 | 2.0.0 | 434 |  | **Port internal/datafiles, internal/rows and internal/jsonvalue to Rust, loading on first reach a source with an index, and embed the shipped data per locale behind a Cargo feature.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.3 |
| 437 | 2.0.0 | 434 |  | **Port internal/grammar to Rust: templates, paths, selectors and where clauses.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.3 |
| 438 | 2.0.0 | 436, 437 |  | **Port internal/engine/tree and internal/engine/load to Rust: compile, bind, link and resolve.** | 4 | 2 | 3 | 7 | 9 | 1 by 3 | 0.1 |
| 440 | 2.0.0 | 435, 438 |  | **Port internal/engine/render and internal/avoid to Rust: drawing, named picks and distinct draws.** | 4 | 1 | 3 | 6 | 8 | 1 by 3 | 0.1 |
| 444 | 2.0.0 | 440 |  | **Fail the Rust gate where a render allocates more than its pinned count, as the Go performance gate asserts allocations.** | 2 | 1 | 1 | 2 | 3 | 12 by 5 | 15.6 |
| 441 | 2.0.0 | 435, 437 |  | **Port internal/builtinfunc to Rust: every builtin, Go's reference-time date layouts such as '060102', the checksums, calc with its Inf, and reword.** | 3 | 1 | 3 | 5 | 8 | 1 by 3 | 0.1 |
| 442 | 2.0.0 | 440, 441 |  | **Port records and their serializers to Rust: JSON, NDJSON, CSV and SQL, typed columns, linked templates through --emit, streamed.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.3 |
| 443 | 2.0.0 | 442 |  | **Port the CLI to Rust, with its getopt_long flags, exit codes, --list and --data-path.** | 3 | 1 | 3 | 5 | 6 | 1 by 3 | 0.4 |
| 446 | 2.0.0 | 433, 442 |  | **Serve the Go library from the Rust core through cgo, under the module path github.com/larvit/fejkdata/v2, keeping New, its options, Fake, the templates, records and FakeStruct.** | 4 | 3 | 3 | 8 | 8 | 1 by 3 | 0.1 |
| 445 | 2.0.0 | 443 |  | **Build and publish the Rust CLI on every channel the Go one reaches: the binaries for six targets, deb, rpm and apk, Homebrew, Scoop and the image, with cargo-audit in place of govulncheck.** | 4 | 3 | 3 | 8 | 8 | 1.1 by 3 | 0.1 |
| 447 | 2.0.0 | 443, 445, 446 |  | **Delete the Go engine, keep the Go tests that check the Go library's API, and revise each decision in docs/decisions.md that names Go.** | 3 | 2 | 3 | 6 | 6 | code health by 3 | 1.6 |
| 448 | 2.0.0 | 359, 442 |  | **Build @larvit/fejkdata's WebAssembly from the Rust core with wasm-bindgen, in place of the Go build item 359 makes.** | 3 | 2 | 2 | 5 | 5 | 1.2 by 6 | 1.3 |
| 450 | 2.0.0 | 447 |  | **Write MIGRATION.md, listing each change 2.0.0 makes to what a user relies on and what to do about it.** | 1 | 2 | 1 | 2 | 3 | 1 by 2 | 1.2 |
| 449 | 2.0.0 | 447, 448 |  | **Rewrite the README's Install, Library, Layout and Development sections, and AGENTS.md's rules, for a Rust core with a Go library.** | 2 | 2 | 1 | 3 | 4 | 11 by 2 | 0.6 |
| 439 | 2.0.0 | 438 |  | **Port internal/engine/fence to Rust: every proof and fence a load runs, with errors that name the same path, spelling and fix as Go's.** | 4 | 2 | 3 | 7 | 7 | 1 by 3 | 0.2 |

## Details

### 455. Run mypy at full strictness over data-import, release-tooling and vulnerability-filing in the gate, and fix every error it reports.

`mypy --strict` reports 1,240 errors today: 725 in data-import, 417 in release-tooling and 98 in vulnerability-filing. Add mypy beside ruff in the Dockerfile's `python-lint` stage, and annotate each folder until it passes.

### 456. Gate cyclomatic complexity at 10 in both gocyclo lines, with every function over it split or listed, with its reason, as a table-shaped dispatch.

`AGENTS.md` gates cyclomatic complexity at 14, twice the ceiling of seven that technical-principles.md ("About seven things per level") sets, and nothing records why. Functions of 8 to 14 branches are where the comprehension panel scores Locality about 5.6, under the 7.0 floor. `gocyclo -over 7 -ignore _test .` lists 83 functions today, and `-over 10` lists 25.

Chosen: lower the gate in steps, to 10 first, because 25 functions fit one chunk and 83 do not. The gate is applied by both gocyclo lines, the Dockerfile's and compose.yaml's `cyclo`. Split each of the 25, or list it in a short allow-list beside those lines with its reason, where it is a table-shaped dispatch: a switch with one case per kind of its input. gocyclo cannot exempt one function, so the allow-list filters its output. Change `AGENTS.md`'s rule to 10 and the allow-list. When this item ships, file the next step, to 7, as item 453 and later chunks split the rest.

### 327. Add `is_IS`'s national personal id, the kennitala, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 453. Reach the comprehension floor: lift Locality so a name rule asks a route what it keeps, draws and addresses, and no rule slices `levels`, a tail or `bound` itself.

The two scoring runs on PR #372 at 5a00a5c left Locality at 5.63 both times, Shape at 6.50 and 6.38, Self-sufficiency at 6.63 and 6.13 and the overall at 6.38 and 6.13, under the floor of 7.0. Item 421 gave each name rule a file; what the seats still found hard is what the rules share. Every rule reads the routes `walkRoutes` emits, and decides by index arithmetic over a route's `levels`, a read's tail and `pickReads.bound`, whose layout `placedAt` and `tree.PickKey` fix elsewhere. A reader holds that layout in every rule file.

Give the route a type that answers the rules' questions by name: what it keeps, where it draws afresh, which level it addresses, where two routes part. The index conventions then live in that one type, beside the invariant that `levels` is one longer than the tail, and a rule reads without `namefence.go` and `tree/pick.go` open beside it. Give the walk's `(at, inPick)` state one named value too, so the three node cases of `walkRoutes` say what each does to it. No refusal, error text or rule order changes.

The places the seats named, which this item restructures:

- `internal/engine/fence/namefence.go:112` `walkRoutes`: seven of the eight seats of the two runs; a recursive closure whose node cases treat `(at, inPick)` differently, with the meaning decided in `placedAt` and `below`.
- `internal/engine/fence/refusekeptandfreshfield.go:76` `readsFieldApart`: five seats; level slicing against `tail[:j]`.
- `internal/engine/fence/refusefreshdraw.go:21` `refuseFreshDraw` with `:78` `readAddressing` and `:90` `rendersField`: three seats; `r.Levels()[p.bound:]` explained only at `pickReads`.
- `internal/engine/fence/refusedrawnapart.go:39` `sharedDraw`: three seats; order-dependent cases with a landing check outside the switch.
- `internal/engine/fence/refusetwicedrawn.go:32` `readTwiceApart`: indexes `levels[common]`, safe only by an invariant no line states.
- `internal/engine/tree/pick.go:83` `LevelsUnder` with `:97` `LandsInPick`: all eight seats; the one place load and render agree on a pick's levels, in another package than its four fence readers.

Out of scope: `render/pick.go` `readName` and its memo, which item 423 took, and `tree/loadfinds.go`'s two phases.

The architecture pass of 0.5.0's release review added two parts. Make `sharedDraw` return a named case, so its cases no longer depend on their order. Move the name rules, `drawsalong.go`, `namefence.go`, `readinsidefield.go` and the five `refuse*.go` files, into a package `internal/engine/fence/names`, and revise the decision that places every fence in `internal/engine/fence`. Items 452 and 164 wait for this one, since each would change every rule again.

### 326. Add `is_IS.phone`, drawn from Iceland's numbering plan.

Find the numbering plan and its licence first.

### 451. Add a `fi_FI` henkilötunnus variant over DVV's issued individual numbers 002 to 899, beside `fi_FI.henkilotunnus`, which draws 900 to 999.

`fi_FI.henkilotunnus` draws numbers DVV issues to no one, as the decision "The Finnish id draws DVV's unissued 900 to 999" records, which names this variant as a later item. A checker that treats endings from 900 as synthetic ("keinohenkilötunnus", per the Vero API FAQ on testing) may refuse them, so a hand fixture for such a checker needs numbers of the issued range. Goal 3 asks the best category to offer more than one kind. A drawn number may be a living person's, so the README says so beside the variant, as it does for the Danish `cpr`, and the default stays 900 to 999 for goal 3.3.

`{hetu(x)}` draws 900 to 999 only, so the item decides how the variant takes the range, records it in `docs/decisions.md` and revises the decision above. Recommended: a second argument to `{hetu(x)}` naming the range, beside a path such as `fi_FI.henkilotunnus-issued`, since the control character and century sign are the same and a second builtin adds a name for one meaning. Bench the path's and the argument's spelling with the README's audience before it ships. The individual number stays of the parity `x` reads, and 001 stays undrawn, as DVV issues from 002.

### 324. Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.

Read them from the register the streets item names, Staðfangaskrá, where it holds them; else find an open source first.

### 325. Give `geo/IS`'s places their streets.

Staðfangaskrá, Iceland's address register, holds every street with its póstnúmer.

### 334. Add `nn_NO`, written in Nynorsk, reading `nb_NO`'s names and id and `geo/NO`, with month and weekday names from CLDR and titles of its own.

Nynorsk is Norway's second written standard, the official form in some municipalities. Names, places, streets, ids and phone numbers read the same in both standards, so only the words that differ need a source of their own.

### 169. Reword the decision "A parent row with no child row is a load error" to say what each geo import does with such a row: drop it, or exit where `geo-se.py` would lose a kommun or a län.

Its entry says "The import script drops or fills such rows". There are two geo imports, and `geo-se.py` exits when `geo.fence` drops a kommun or a län, while `geo-us.py` drops counties and states. Rewording a decision is the maintainer's call.

### 351. Name in the decision "A known vulnerability in the Go version or a required module blocks a release, even in code fejkdata never calls, and blocks no merge" each line its escape hatch edits, so dropping the scan from the release job releases instead of skipping the job.

The decision says that where no Go version fixes a vulnerability yet, the maintainer drops the vulnerabilities job from the release job's `needs`. The release job's `if` in `.github/workflows/test.yml` also requires `needs.vulnerabilities.result == 'success'`, and a job missing from `needs` is no success, so the release job would be skipped, not run. A skipped job isn't red, so nothing would ship and nothing would say why. Rewording the decision is the maintainer's call. The principles review of item 336 found it.

### 452. Refuse at `New` a read of a name whose render reaches a table cell that draws afresh a path another read of the name reads from the pick; today the name fence walks no cell.

`fence/namefence.go` `walkRoutes` steps into a node's children by `tree.RenderEdges`, and skips a `*tree.TableColumn` on purpose, to keep what item 421 found. So no refusal rule sees a read inside a table cell.

Found by probe on 2026-11-12: with `place.json` `{"format":"{zip}/{tag}","rows":"place.tsv","key":"name"}` and `place.tsv` holding the row `Stockholm`, `1{digits(4)}`, `{/place.zip}`, the template `{/place as n}{n.tag} {n.zip}` loads and prints two different postal codes, such as `19166 15556`. The same shape in a template's fields, `"zip":"1{digits(4)}","tag":"{/place.zip}"`, is refused by `refuseFreshDraw`: "`{n.tag}` reaches `{/place.zip}`, which draws afresh a path `{n.zip}` reads from the pick".

Drop the `*tree.TableColumn` case in `walkRoutes` so a cell's reads become routes, apart as a table's format is, and check each of the five rules in `nameRules` against a cell. Test first with the probe above; every shipped category must still load. A template that loaded is then refused, which the changelog states.

### 343. Split the README's Data section, about 225 lines under one heading, into sub-headings, one per feature, so each feature has its own heading (goal 11.1).

`## Data` in `README.md`, from its heading to the next `## ` heading, holds the locale categories, the linked tables, the `sv_SE` personnummer recipe, the `nb_NO` fodselsnummer paragraphs and the `geo` namespace with no sub-heading, and every `[Data](#data)` link, and each `[index](#data)` link, points at the whole section. The prose reviewer of PR #249 found that the `nb_NO` fodselsnummer text uses "bound sex" and "as a personnummer does", which lines further down explain. The split keeps every README example `readme_test.go` runs passing, and every in-README link resolving, each to the sub-heading it means.

### 354. Make `test_main_hands_the_cut_tag_to_the_workflow` pass across UTC midnight, by comparing the changelog's date with the one `patch_release.main` read.

`release-tooling/patch_release_test.py` `test_main_hands_the_cut_tag_to_the_workflow` reads the wall clock after `patch_release.main` read it at `patch_release.py:270`, and asserts the changelog's date equals its own. A gate run that crosses UTC midnight between the two reads fails. Applies the technical principles' test bullet: a test is deterministic. Found by the principles pass of PR #267.

### 167. Lower the reading cost of the files a chunk reads most before its first commit.

`reading-cost` on 2026-10-11: the last 5 chunks' median is 91452 new input tokens before their first commit, the 5 before them 69756. Read most over the last 5 (opens and searches): `README.md` 37, `todo.md` 29, `docs/decisions.md` 15, `CHANGELOG.md` 12, `data_test.go` 11, `shipped_test.go` 8, `cmd/fejkdata/main.go` 7, `data-import/geo-us.py` 7, `DATA-LICENSES.md` 5, `bench_test.go` 5. The chunk shipping this item runs `reading-cost --reset`.

`reading-cost` after the chunk of item 292, on 2026-10-22: "Rising: the last 5 chunks' median is 113334 new input tokens before their first commit, the 5 before them 78241". It listed no file read most over the last 5.

`reading-cost` after the chunk of item 347, on 2026-10-28: "Rising: the last 5 chunks' median is 110298 new input tokens before their first commit, the 5 before them 97116". It listed no file read most over the last 5.

`reading-cost` on 2026-11-03: "Rising: the last 5 chunks' median is 102919 new input tokens before their first commit, the 5 before them 70758". Read most over the last 5 (opens and searches): `todo.md` 4, `README.md` 3, `docs/decisions.md` 3, `.github/workflows/packages.yml` 2, `data_test.go` 2, `norwegian_test.go` 2, `release-tooling/goreleaser.yaml` 2, `vulnerability-filing/file_vulnerabilities.py` 2, `vulnerability-filing/file_vulnerabilities_test.py` 2, `.dockerignore` 1.

`reading-cost` on 2026-11-04: "Rising: the last 5 chunks' median is 117951 new input tokens before their first commit, the 5 before them 98791". Read most over the last 5 (opens and searches): `todo.md` 31, `README.md` 10, `data_test.go` 8, `docs/decisions.md` 8, `table_test.go` 5, `node.go` 4, `path.go` 4, `AGENTS.md` 3, `env.go` 3, `internal/datafiles/datafiles_test.go` 3.

`reading-cost` after the chunk of item 365, on 2026-11-04: "Rising: the last 5 chunks' median is 118390 new input tokens before their first commit, the 5 before them 98791". Read most over the last 5 (opens and searches): `README.md` 34, `todo.md` 30, `data_test.go` 15, `DATA-LICENSES.md` 8, `CHANGELOG.md` 7, `compose.yaml` 5, `testdata/shipped_shape.txt` 4, `docs/decisions.md` 3, `data-import/loglevel.py` 2, `data/misc/car.tsv` 2.

`reading-cost` at the release review of 0.3.0, on 2026-11-04: "Rising: the last 5 chunks' median is 103168 new input tokens before their first commit, the 5 before them 89405". Read most over the last 5 (opens and searches): `README.md` 29, `todo.md` 17, `cli/cli.go` 13, `docs/decisions.md` 12, `CHANGELOG.md` 7, `name_test.go` 6, `namefence_test.go` 6, `data_test.go` 5, `shape_test.go` 4, `docs/research/research-sources-se.md` 3.

`reading-cost` after the chunk of item 377, on 2026-11-04: "Rising: the last 5 chunks' median is 103168 new input tokens before their first commit, the 5 before them 89405". Read most over the last 5 (opens and searches): `README.md` 29, `todo.md` 17, `cli/cli.go` 13, `docs/decisions.md` 12, `CHANGELOG.md` 7, `name_test.go` 6, `namefence_test.go` 6, `data_test.go` 5, `shape_test.go` 4, `docs/research/research-sources-se.md` 3.

`reading-cost` before the chunk of item 379, on 2026-11-08: "Rising: the last 5 chunks' median is 104363 new input tokens before their first commit, the 5 before them 72993". Read most over the last 5 (opens and searches): `README.md` 27, `todo.md` 24, `readme_test.go` 11, `builtins_test.go` 7, `cli/cli_test.go` 7, `record_test.go` 7, `compose.yaml` 6, `danish_test.go` 6, `norwegian_test.go` 5, `cli/cli.go` 4.

`reading-cost` at the release review of 0.4.0, on 2026-11-09: "Rising: the last 5 chunks' median is 101184 new input tokens before their first commit, the 5 before them 86259". Read most over the last 5 (opens and searches): `todo.md` 37, `README.md` 21, `swedish_test.go` 11, `calc_test.go` 6, `compose.yaml` 6, `adddays.go` 5, `valueproof.go` 4, `anonymize_test.go` 3, `docs/decisions.md` 3, `internal/builtinfunc/builtinfunc.go` 3.

`reading-cost` on 2026-11-08: "Rising: the last 5 chunks' median is 112711 new input tokens before their first commit, the 5 before them 85771". Read most over the last 5 (opens and searches): `README.md` 38, `todo.md` 22, `swedish_test.go` 14, `docs/decisions.md` 11, `builtins_test.go` 7, `valueproof.go` 7, `CHANGELOG.md` 6, `DATA-LICENSES.md` 6, `calc_test.go` 6, `adddays.go` 5.

`reading-cost` before the chunk of item 380, on 2026-11-09: "Rising: the last 5 chunks' median is 112711 new input tokens before their first commit, the 5 before them 85771". Read most over the last 5 (opens and searches): `todo.md` 36, `README.md` 25, `swedish_test.go` 11, `compose.yaml` 8, `valueproof.go` 7, `builtins_test.go` 6, `calc_test.go` 6, `internal/builtinfunc/builtinfunc.go` 6, `adddays.go` 5, `docs/decisions.md` 4.

`reading-cost` before the chunk of items 413 and 411, on 2026-11-10: "Rising: the last 5 chunks' median is 71284 new input tokens before their first commit, the 5 before them 55774". Read most over the last 5 (opens and searches): `todo.md` 45, `format.go` 12, `docs/decisions.md` 10, `README.md` 9, `resolve.go` 8, `arm.go` 7, `index.go` 5, `reference.go` 5, `table_test.go` 5, `inline.go` 4.

### 195. Parse each builtin argument once, in `checkArgs`, and hand `prep` and `proveNumber` the parsed value.

Applies "Validate data and build DTOs as early as possible" in `~/.claude/principles/technical-principles.md`. Today `checkArgs` (`internal/builtinfunc/builtins.go`) checks an argument, then `prep` and `proveNumber` parse it again through `atoi`/`atof`, which panic on bad input; `layout.go`, `transform.go` and `calc.go` call them too; two parsers can drift, and a drift panics at compile instead of failing the load.

### 40. Parse `defaultTable`'s path with the library's grammar.

The CLI counts brackets of its own, so a change to the selector grammar desyncs the default `--table` silently.

### 95. Hold a struct's columns and the field each fills in one slice.

`structShape.fieldIndexes` pairs with its record template's `columns` by index alone, against the technical principle "one owner per value".

### 116. Turn a row index into its rows file's line in one function.

Today `containedNode`'s cell line in `graph.go`, `compileCells` in `table.go` and each site in `internal/rows` add the header line and the one-based count.

### 350. Bring the release path's links down until paired rulings against f90b4d6 and 5aca8a1 no longer find the project reads worse.

The decisions "A known vulnerability in the Go version or a required module blocks a release, even in code fejkdata never calls, and blocks no merge" and "CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag" each record a paired ruling that found the project reads worse, and this item ends both costs. f90b4d6 is from before CI filed vulnerabilities into the plan, and 5aca8a1 from before it cut patches. The means: less of `todo.md`'s layout in the filer and the merge gate, one writer of `CHANGELOG.md`'s release sections, and fewer linked jobs between the gate, the release and the plan. The ruling against f90b4d6, on filing vulnerabilities into the plan, named the merge gate reading `todo.md`, the filer splicing `todo.md` as text, and a release waiting on vuln.go.dev. The ruling against 5aca8a1, on CI cutting patches, named two linked scripts, four linked jobs and several writers of `CHANGELOG.md`. A ruling against 5aca8a1 alone can't see the first loss, since 5aca8a1 already holds it. The seats' places from the 2026-10-14 and 2026-10-15 rulings name what to make easier to read: `vulnerability-filing/file_vulnerabilities.py` `ROW`, `file_items`, `plan` and `release_for`, `release-tooling/patch_release.py` `FILER`, `.github/workflows/test.yml` `release` and `plan` with their combined conditions, `.github/workflows/packages.yml` "Whether the tag is the latest release", which reads the latest release from a script named for patches, `Dockerfile`'s `vulnerability-filing` stage, which copies `todo.md` into the merge gate, and `CHANGELOG.md`'s three writers.

### 34. Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.

`isin`: Luhn over letters expanded to digits. `aba`: 3-7-1 weights. `vin`: position 9 over the whole; a sample taking the WMI, since the check sits mid-string.

### 155. Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.

`{luhn()}` skips letters, so it draws a wrong ISIN check digit, and CUSIP's check digit is not Luhn.

### 15. Add `misc.ip`'s documentation, private and CIDR forms, `misc.ipv6` in `2001:db8::/32`, and a locally administered `misc.mac`.

Templates, from RFC 5737, RFC 1918 and RFC 9637, which are facts.

### 22. Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.

17 of the 40 distinct ones shipped today sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost` provably reach nothing.

### 252. Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.

The AUR closed new account registration on 2026-10-14, so 0.1.0's packaging shipped without it. Once an account exists, restore what that chunk dropped: the `aurs` block in `release-tooling/goreleaser.yaml` (`git show 285ed85:release-tooling/goreleaser.yaml`), `aur` in both `--skip` lists, the `AUR_KEY` secret in both workflows, and the README's Install line and secret bullet.

### 256. Add `misc.bic`, built by its structure rules.

A template; the structure rules are facts.

### 254. Say in the docs which of three kinds each category's values are: real, valid or invented.

Goal 3 names the kinds. A real value is a thing that exists, such as the territory `SE`. A valid value is made up but passes the check the real system applies, such as a Swedish personal number. An invented value is made up entirely, such as a town that does not exist. The maintainer, 2026-10-18: all three have their place, and it is good to tell them apart. Before the three words ship, a panel of readers seated as the README's audience picks the spelling.

### 24. Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.

ISO 3166 reserves it for Western Sahara and the root zone has never been delegated it, so no resolver answers for it.

### 255. Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.

A table and its template, from each network's rules, which are facts.

### 117. List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.

Today the refusal names only the missing field, so a data author fixing a typo opens the category to find the spelling. Item 92 covers the hyphenated `calc` operand only.

### 168. Test in the merge gate the shared steps of `data-import/`, starting with `geo.fence`, the småort steps of `geo-se.py`, and the name-shape rules of `company-no.py` and `company_names.py`.

Against "Most, if not all, code should have automated tests" in `technical-principles.md`: no Python under `data-import/` has a test, and the gate, `docker build .`, runs Go alone. Today `geo.fence` is checked only by regenerating both countries' TSVs from the full downloads and diffing them. Each locale's places item adds a geo script for its country, each calling `fence`.

The tests also cover the rules of `company-no.py`: `shape_of` reads an unlisted word in an ENK name as the owner's last name, `roots` takes a word carried in at least two forms, and `register`, `without_wording`, `kind`, `count_shapes` and `endings` hold the rest. They cover `word_usage` and `trade_words` in `company_names.py`, which `company-se.py` and `company-no.py` share; its `Vocabulary.acronym_max` makes a word in capitals an acronym only up to that many letters or where it has no vowel.

They cover `scb_smaorter`, `inside` and `population_of` in `geo-se.py` too, where no gate test kills a mutant today, such as a swapped latitude and longitude or a dropped kommun filter.

### 257. Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.

Templates; the check-digit rules are facts.

### 389. Add `misc.rrule`, recurrence rules a test checks against RFC 5545's grammar.

From a real-world use case, a scheduler's fixtures, whose author called it niche, and it would have replaced most of one fixture's template. No register publishes how often each rule is used, so it draws evenly, as the decision "A table whose register publishes no frequency draws evenly" states, rather than the weighted mix the author asked for.

### 113. Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.

`iban` in `internal/builtinfunc/checksum.go` draws a random numeric BBAN and sets only the mod-97 digits. BE's last two digits are the first ten mod 97, NO carries a mod-11 account check digit, FI a Luhn digit, ES two control digits, and SE's three-digit bank id must be a real bank's (`docs/research/research-sources-se.md`). Goal 3 asks for the check the real system applies; the README's `{iban(CC)}` row promises only length and mod 97, and changes with it.

### 406. Add a transform that writes a value escaped for a JSON string, so a text template spelling JSON stays valid whatever a drawn value holds.

A record's columns are flat, so a parent with an array of children renders as a text template whose format spells JSON, such as `{{"summary":"{s}"}}`. A drawn value holding `"`, `\` or a newline breaks that JSON, as a company name or a sentence may. The transform writes the value with JSON's string escapes, beside `lowercase`, `uppercase`, `ascii` and `slugify`. A writer panel picks its name.

### 260. Add `misc.element`, `misc.unit` and `misc.planet`.

Tables, from PubChem, BIPM and NASA, under public domain and CC BY 4.0 terms.

### 261. Add `misc.animal` by class, and `misc.plant`.

Tables, from Wikidata (CC0).

### 164. Let a table's format or a cell bind a name, which lives in that format or cell and keeps one pick through the draw.

The maintainer ruled so on 2026-10-09: a name is a lexical closure, as the decision "A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read" now states. Today `table.go` refuses `{/sv_SE.person as p}{p.first} {p.last}` in a cell with `a table binds no name`. A cell reads no name its table's format binds. Revise that decision's title and the README's Names list with it.

### 262. Add `misc.bloodtype`, weighted by its global distribution.

A choice, from AABB's figures, which are facts. Item 242 adds a Swedish one.

### 50. Reword `prove`'s refusal of a typed column reading a row, `{/region}`.

Say the format is the table's own (`region's format "{name}"`), name the read to write, `{/region.<column>}` with its columns, and stop calling a one-column format, `{code}`, composed text.

### 52. Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.

With several `--data-path` layers the author has to work out which directory won, while a parse error on the same file names it.

### 92. Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.

A hyphen is always subtraction in a calc, and the shipped data has hyphenated fields (`postal-code` and `street-number` in `geo.SE.address`). The refusal names the first fragment, not the field the author meant.

### 228. Offer `{/path as p}` where a token starts with a name nothing binds and no loaded path starts with, in a record's format too: today `{p.first}` with `p` unbound suggests `{/p.first}`.

From the 2026-10-11 audience bench, a Go test author tagging `{p.first}` before binding `p`. The hint is built in `checkFieldRead` before the root is known, and only where the template has no fields; offering the binding there also told `{sv_SE.person.last}` to bind a name `sv_SE`. Decide between the reference and the binding once the root is loaded.

### 38. Compare `calcParser.binary`'s operator as a rune.

`byte(p.rs[p.pos])` reads U+012B `ī` as `+`, so `{calc(a ī b)}` compiles as `a + b`.

### 44. Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.

The decision "`misc.tld` keys carry the leading dot, where other tables key on a bare code" leaves `misc.tld[se]` missing today.

### 45. Name the node a selector follows in `childNamed`'s refusal.

`sv_SE.person[1].first` answers `1 is not a table`.

### 49. Report a CRLF rows file holding only its header as having no rows.

`parseRows` answers "line 2 is empty".

### 110. Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.

Today `Fake("sv_SE.person.")` reports `no field ""`, while `{/sv_SE.person.}` reports `path has an empty segment`. `grammar.CheckSegments` is the check a reference, a sibling path and a step up `..` run.

### 355. Run `docker build .` and the Packages dry run on a release commit.

The 0.2.0 commit changed only `CHANGELOG.md`, so the `changes` job skipped the `docker build` jobs and the dry run, and the release job tagged a commit neither had built.

Make the `changes` job count a change to `CHANGELOG.md`'s top heading from `Unreleased` to a version as a code and packaging change, so the release commit builds through `docker build .` and runs the dry run before it is tagged. Adding `CHANGELOG.md` to the `code` and `packaging` lists would do the same, but it would also run the full build on every changelog fix-up and the dry run on every data pull request, which slows every such pull request.

### 181. Fail a load where two sources define one category, unless the library's option or the CLI's flag says the later replaces the earlier, and then make every read reach the replacement.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Proposed spellings: `WithReplace()` and a boolean `--replace` covering every `-d`, so it goes anywhere on the line. Bench the option, the flag and the clash error with the data and hand fixture authors before merging; the error names the option or the flag. A clash is found from the categories each source's index names and the files each source without a manifest holds, without parsing them. Revises the decision "A `--data-path` override rebinds every reference to the category it replaces", ending its conflict with goal 8.3. The changelog names the flag, the option and the new refusal. The same `embed.FS` passed twice, as `svse.Modules()` beside `data.Modules()` passes `geose.FS`, loads once and is no clash. Until then, a `-d` file replacing `sv_SE/first-name.json` breaks `{/sv_SE.person}` with `token {n.name}: no variant of this 2-way choice carries "name"`, naming no replacement; the clash error ends that.

### 81. Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.

The decision "A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`" lets `{calc(a/(b*c))}` with `b` fixed at `"0"` write `Inf` into every row. An operand that is sometimes not a number prints `NaN`. Both work against goals 5.1 and 3. `valueproof.go` already proves a typed column's divisor nonzero. Item 414 made calc read text as a number one way, through `proven.Literal`, which refuses `NaN` and `Inf`.

### 265. Add `misc.commit` messages, branch names and `sha`s.

Templates, from lists an LLM writes and `{hex(40)}`.

### 258. Add `misc.product` names, adjectives, materials and departments, and `misc.garmentsize`.

Tables or choices, from lists an LLM writes.

### 266. Add `misc.password`, `hash`, `jwt`, `slug`, `hostname`, `avatar` and `placeholderimage`.

Templates, by each one's shape. A JWT's parts take `{base64url(n)}`, which item 35 adds.

### 196. Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0", and ends the locale-address decision's conflict with goal 5.2. Carved out of item 182 on 2026-10-11, after six rounds of the comprehension panel ruled each attempt worse. Tried and ruled worse: a module list carried per category from each source's manifest (two routes, `walked` and `indexed`); the union of every manifest's list on the `Generator`, noted at four entry points; and the same noted inside `loadReached`, its list threaded through `loadInline` and `structCompile`. Each named every module read, even loaded ones, until the list held namespaces and the note named only those nothing in the tree provides, which the product-owner and goals reviewers accepted. A shipped locale's `Modules()` already carries its places, so this serves a bare `FS` beside a stand-in, a `--data-path` folder and anyone else's module; the error names the module and points to `Modules()`, as the 2026-10-11 bench asked. Bench the manifest key and the note's wording with the data author, since entries changed from import paths to namespaces after the 2026-10-11 bench.

### 263. Add `misc.color` names with their hex, rgb, hsl and cmyk forms.

A table and its templates, from CSS Color 4 (W3C) and the XKCD colour survey (CC0).

### 28. Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.

Today `sv_SE.personnummer` draws its own sex, so a chosen sex and a number agree only through the README Data section's template, which reads both through one bound `sex`. `sv_SE.person` reads its own sex and birth date in its `personnummer` field in `data/sv_SE/person.json`, so this item may follow that shape.

### 405. Let `repeat` take a range, such as 5 to 15, drawing the count on each render.

An epic holds 5 to 15 stories, an order 1 to 8 lines. Today `repeat` is one integer, so a varied count takes a weighted choice of templates that differ only in their `repeat`. The repeat cap counts a range's upper bound. A writer panel picks the spelling, such as `"repeat": [5, 15]`; the README's Repeat section and the decision "The repeat cap bounds renders, not bytes" name it. Pairs with item 404, for a parent with a varying number of children.

### 13. Place xlsx cells by their `r` reference in `data-import/xlsx.py`.

A sheet omitting an empty cell shifts every later column left.

### 165. Let a table carry `name` with neither `key` nor `parent`, and list the rows a name matches by their line where it matches several.

The maintainer ruled so on 2026-10-09. Today `bindOptions` in `internal/rows` refuses such a table, since `ambiguous`, which reports a name matching several rows, lists those rows by key. Afterwards `ambiguous` lists them by line, the header being line 1, and ends with "make the names unique, or add a key column".

### 166. Let a key or name cell hold a bracket, a brace, a double quote or a pipe, so the table loads, its row draws at random, and only a selector cannot reach it.

The maintainer ruled so on 2026-10-09. Today `checkSelectorCells` in `internal/rows` refuses the whole table. Afterwards a path naming such a key fails with an error naming the character a selector cannot spell, and the README says a row whose name holds such a character is still selected by its key.

### 39. Stop the `fence.FieldKinds` suggestion naming a field's own kind or a narrower one.

An `int64` or `float64` field unproven in range is told to become itself, and a `uint64` one to become `int64`.

### 42. Refuse a struct tag's `datatype` other than its Go type's before proving its values, as the README's Library section promises.

`datatype: boolean` over `{int(1,2)}` on an `int` field answers `prints an integer, not a boolean`.

### 46. Report the same error every load for a table with two bad options.

`readTableOptions` in `table.go` returns on the first in Go's map order.

### 51. Spell a table one way across the errors that name it.

In `internal/rows`, `Pins.PinRow` names `t.path` and `Pins.MustRow` panics with `t.segment`, so one table is `territory` and `misc.territory`, and the short spelling names no file where two folders hold that name.

### 115. Spell a rows file's line one way in every error, naming the file.

Today a cell reads `g.t.b, line 3` (`containedNode.labelIn`), a cell check `t.tsv: line 3, b` (`compileCells`) and a link `t.tsv line 3` (`rows.Table.Link`). A reference cycle found from a table's format through a cell names no line at all: `reference cycle: g.t -> format -> b -> b -> /h -> …`, where the column-to-cell step in `renderEdges` repeats the column name.

### 408. List every shipped path in the README's Data section with an example value and whether it is real, plausible or valid, and where it comes from.

Goal 3 says the docs say which kind each value is, but today that sits in prose spread over the Data section, and a reader can't scan for whether a path exists or what it prints. With a Nordic locale added each release, a table is how a reader finds `fi_FI.phone` and what it looks like. Generate or test the table against what loads, so a new path can't ship without a row, and keep the example values seeded. Fits item 343, which splits the Data section by feature.

### 259. Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.

Tables and templates, from OurAirports (public domain) and Wikidata (CC0).

### 376. Let a template ask for a `sv_SE.personnummer` of a child, an adult, someone of 65 or more or a centenarian, and for a mix of them.

The category weighs four age bands in one choice, and a template can reach only the whole. A test of a system's handling of the `+` has to draw until one comes, at 0.028%. Bench the spelling with the README's audience: a table selector, a path per band or a builtin. The decision "`sv_SE.personnummer` and `sv_SE.personnummer-12` are each a choice of four age bands, up to age 113, and `personnummer-12` has no `ten`" sets the bands and weights.

### 267. Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.

Tables, from Project Gutenberg's catalog, MusicBrainz and USDA, each CC0.

### 264. Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.

Tables or choices, from GitHub Linguist (MIT) and curated lists. Item 32 adds `misc.browser`.

### 268. Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.

Tables or choices, from lorem ipsum and lists an LLM writes.

### 387. Let a template draw a number from a bounded normal distribution, so durations and start times cluster without weighted ranges.

From a real-world use case, a scheduler's fixtures: durations cluster around 2 hours and starts around 08:00. A weighted choice of `{int()}` ranges fills a typed column today, coarsely. Bounds are required, so a typed column's proof and `calc`'s still hold. A writer panel picks the spelling, between a third argument to `{int()}` and a function of its own, and what a draw outside the bounds does.

### 454. Give `geo.FI`'s postitoimipaikat and kunnat their Swedish names, such as Helsingfors and Åbo.

Posti's postcode file names each postitoimipaikka in Swedish too (field 49, 30 characters), and Statistics Finland's classification names each kunta in Swedish. Add a `swedish-name` column to the locality and municipality tables, as the street table has, falling back to the Finnish name where no Swedish one exists, and recase Posti's Swedish names as `recased` does the Finnish ones.

### 91. Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.

A calc operand parses as a bare identifier, so `{calc(c.lat * 60)}` fails with `unexpected ".lat * 60"`, and a second binding of `/misc.coordinate.lat` would draw another coordinate. Resolve the path through `compileRead`'s leaves, as `operandNodes` does for a bare name. A data author who bound a row, `{.product as p}`, then bound its column, `{p.price as pr}`, got `no field "p"; a binding names a field, or a reference…` though `p` is a bound name.

### 109. Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.

Technical principle: indent with hard tabs. All 19 scripts indent with four spaces, so an edit there either keeps the wrong style or rewrites the whole file.

### 23. Give every other shipped category that could reach something real a path that never does.

`phone` draws live PTS, NANP and Nkom ranges — PTS's five fiction series are the Swedish inert set, sourced in [research-sources-se.md](docs/research/research-sources-se.md), and the NANP and Norwegian ones still want a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.

### 194. Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.

Goal 5.3 has an author prove their module, and goal 8.1 has fejkdata's data load as anyone's; today only fejkdata's own test, `TestShippedManifestsAreCurrent` over `indexEntry`, writes and checks an index. A category file the index leaves out never loads, silently.

### 158. Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.

A file a Windows editor saves with a BOM prints an invisible U+FEFF first, and a JSON template with one is read as a format string. PowerShell 5.1 encodes what it pipes to a program by `$OutputEncoding`, ASCII by default. Measured on a GitHub `windows-latest` runner (PowerShell 5.1.26100) while writing the README's PowerShell section: `$OutputEncoding` read US-ASCII, yet a pipe to a program still began with the bytes `EF BB BF`, so a JSON template exited 2 (`'{' inside a token in "\ufeff{…`) and a lone reference under `--format` exited 2 (`an inline record has no fields`). PowerShell 7.6.7 sent no BOM.

### 163. Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.

The README's Audience names a bulk fixture author writing millions of rows, and `-n 5000000` is misuse today, so they loop. The decision "The repeat cap bounds renders, not bytes" covers nested repeats in data, not `-n`. The README's Functions section justifies the builtin caps by a fat-fingered `hex(2000000000)`, a guess at a mistake under goal 4.7; the reason that holds is that one render may not allocate gigabytes. The goals audit of the narrowed goal 4.7 found it.

### 269. Add `misc.direction`, `misc.continent` and `misc.ulid`.

Choices and a template.

### 161. Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.

`grammar/identifier.go` allows a field named ` x`, and `{ x}` reads it, but `checkBind` in `format.go` refuses every padded binding, so no spelling binds that field to a name (goal 3.1), and the refusal names one meaning only (goal 4.6). The goals audit of the narrowed goal 4.7 found it.

### 29. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 36. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 56. Build on a manual run of `test.yml` on `main`.

A manual run diffs against `origin/main`, so on `main` itself it compares `HEAD` with itself and skips `docker build`.

### 53. Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.

Today only a JSON field takes one, so `misc.httpmethod`'s booleans are typed in Go and text everywhere else. Typing a shipped column changes what `json` and `sql` write, so the capability is a minor and applying it to `misc.httpmethod` is a major.

### 404. Let a template in a `repeat` or a choice item bind a name to a path through an enclosing name, `{l.street as s}`, so each render draws a fresh row inside the enclosing pick.

A parent with its children, a locality and its streets or an epic and its stories, binds the parent once and wants each child to name a different row under it. With `{/geo.DK.locality as l}` in the format and `{l.street.name}` in a field repeating 3 times, the CLI prints `Vissenbjerg: Søndergade, Søndergade, Søndergade`, since a name keeps every level its reads pass; `{l.street.name as s}` in the field is refused with `no field "l"; … a binding names a field, or a reference`. Today the only way round is one run per parent. Revise the decision "Every `{…}` draws afresh, and only a name keeps a pick", whose list of bindings leaves out a path through a name, and state how the new pick sits beside the enclosing name's. A writer panel picks the spelling. `readName` is a place the comprehension seats found hard; leave it easier to read.

### 75. Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.

`readme_test.go` loads and renders each `json` block, but runs no `sh` example (CLI, Records, Table, Linked tables), no `go` example (Library) and no `text` block of error text, and compares no "Renders e.g." output, so goal 11.2 is not met.

### 17. Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.

`misc.uuid` gains `variant`, for one.

### 14. Spell `misc.creditcard`'s digit runs `{digits(n)}`.

A `"repeat": 14` over `{digits(1)}` draws every 14-digit run with the odds `{digits(14)}` does. Both load; make the shipped data write the shorter one. Re-pin seeded output in its own commit.

### 193. Move the largest shipped packages into nested Go modules that keep their import paths, once `TestShippedDataStaysUnder100MiB` fails.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Each nested module requires the core release that drops its package, so no build sees the package twice. It brings per-module tags and a workspace, so it moves into the release of the chunk whose data trips the gate.

### 393. Write several linked templates in one run with a repeated `--emit <template>=<count>`, parents first, each template's rows contiguous, and `{/run.<template> as c}` reading another template's rows.

Builds the decision "One run writes several linked templates through a repeated `--emit <template>=<count>`, and a template reads another's rows with `{/run.<template>}`", whose "What holds" list is the spec: the load errors (exit 2), the order of the rows, `--out-dir`, the flags `--emit` is refused beside, and that adding a template doesn't change another's rows under the same `--seed`. Take each test's expected values from that list.

Three of the bench's readers feared a category named `run` clashing with `{/run.…}`. Either reserve `run` as a root name, or refuse a category named `run` in a run that reads `{/run.…}`; record which in the decision's paragraph that names this item.

`--format tsv` shows the pieces exist: a table over an earlier run's records draws an id that exists, with the customer of the same row. Keep only the key columns of a written parent in memory, and measure 120,000 parents against goal 12. `{seq(name)}` counters stay per template. Ship the first chunk with two templates, then widen. `--emit` is a new flag, so it needs a `CHANGELOG.md` entry under `Unreleased` (`AGENTS.md`) and its own README heading.

### 407. Let a choice deal its items in exact counts across a run's renders, so 200 issues hold exactly 4 in progress, where weights only approach it.

A fixture often needs a fixed mix: of 200 issues, 4 in progress, 30 done and the rest to do. A weighted choice gives about 4, and the count shifts with the seed, so a test asserting on it needs a loader to fix the counts. A dealt choice is a deck, drawn without replacement. Decide what a render past the last card does: refuse at `New` where the repeat is known to exceed the deck, start a fresh deck, or fall back to the weights; and which renders share one deck: the run, as `{seq()}` shares one counter, or each `repeat`. The deck is per-generator state beside `seq()`'s counters, so the same seed and date still give the same output (goal 9). A writer panel picks the spelling.

### 218. Give every country shipped without streets its streets, once the maintainer accepts ODbL data or the country opens a register.

Ends the decision "A country with no open address register covering it whole ships its places and postal codes without streets".

### 129. Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.

Goal 4.9. Today a path is a second spelling of the template that reads it; afterwards only the template remains.

- A lone reference, a template of one reference token and nothing else, is what it references, as a record column reading one reference already is: `FakeRecord("{/users}")` returns the columns of `users`. The CLI's `loneReference` and the exported `IsLonePath` it calls then go, and `--format sql` takes its default `--table` from the library's parse of the reference.
- `Fake` and `FakeRecord` take a template. `FakeTemplate` folds into `Fake`, `FakeRecordTemplate` into `FakeRecord`, and `IsTemplate` goes, with `grammar.IsTemplate`.
- `Fake` of a lone reference costs what `Fake` of its path costs today, and the allocation gate holds a case for it.
- `List` and `--list` give each entry as the template that renders it, `{/sv_SE.person}`.
- Text that is not a JSON value and holds no `{…}` prints as written (goal 4.1), in a string field's struct tag and `Fake` alike, as stdin's text does in the CLI: `Fake("sv_SE.person")` returns `sv_SE.person`. `CHANGELOG.md` says that a path with no `{…}` prints as its text. A struct tag's `-` still skips its field.
- Where text with no `{…}` is refused, as a record or in a typed struct field, and it spells a path the data offers, the error names `{/that.path}`.
- A lone reference to nothing exits 2 under `--format`, where it exits 1 today.
- A lone reference written as an object holding only a format, `{"format":"{/users}"}`, or as a one-item choice, `["{/users}"]`, names the record as `{/users}` does.
- A struct tag of a bare JSON number or boolean, `fake:"42"`, renders its text, as stdin and `NewTemplate` already do.
- `grammar.CallerPath`, which strips a path's leading `/`, goes with the path spelling, and so does the README's Library line allowing it.
- Decide whether a lone reference under `--format` draws as its path draws today or as a template draws: under one seed, drawing `users` as a path and as the template `{/users}` picks different rows.
- This revises the decisions "A struct tag is a template by its shape", "`FakeTemplate` and `NewTemplate` both stay", "`--list` stays a plain list of paths", "The CLI renders exactly the template on its stdin, and adds nothing", and the exit-1 rule for a lone reference to nothing in "A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129". The README's examples and exit codes, the godoc examples in `doc.go` and `CHANGELOG.md`'s entries move to the template spelling. Delete items 40 and 110, which this closes.

Split it into items for the CLI, struct tags and the library before starting.

### 41. Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.

`geo.SE.municipality[Lund]` and `misc.territory[Sweden]` select a row today. Steps:
- Have a selector that misses name the key of every row whose cell spells it, matched without case, else say the table selects by its key: `misc.territory[Sweden]` and `misc.territory[se]` name `[SE]`, `misc.protocol[tcp]` its key, and `misc.loglevel[Error]` `[err]`.
- Delete the `name` option and the load checks it carries, and have `New` refuse a table still carrying `name`, naming `key`.
- Key `misc.loglevel` on `keyword`, `geo.SE.locality` on the integer `data-import/geo-se.py` assigns as the decision states, and `geo.NO.locality` likewise from `data-import/geo-no.py`, linking each one's `street` and `postal-code` by it.
- Drop `key` from each locale's `last-name`, from `sv_SE.title` and from `nb_NO.title`, and re-pin `testdata/shipped_shape.txt`.
- Leave no selection by name in README.md, CHANGELOG.md, docs/ or `--help`.

### 186. Load a module carrying functions through an option of its own, its functions reaching the generator's randomness and date through an interface the root declares.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Its functions get the date `WithToday` gives. `cli.Main`'s signature grows here, which a minor before 1.0 allows.

### 66. Cut the README's Layout block to the lines that say what a file name cannot.

`doc.go` restates its name.

### 67. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens by restating the decision "GitHub is canonical, and the module path names it", a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 64. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 62. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 59. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 63. Pair a street with its exact postnummer.

Sweden needs an application to Lantmäteriet; today a Swedish street goes to the nearest postal code centroid. Norway needs none: Kartverket's address register, which `data-import/geo-no.py` reads, pairs each address with its postnummer, though `geo.NO` too draws a street and a postnummer apart, each inside one poststed. Finland needs none either: Posti's basic address file, which `data-import/geo-fi.py` reads, lists each street under its postinumero, though `geo.FI` draws the two apart, each inside one postitoimipaikka. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 68. Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.

Gitea converts a repository to a mirror only by re-creating it, so until that runs the copy there is a second owner of one history.

### 65. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go.

### 348. Decide how a Go minor that drops a system the binaries run on, such as an older macOS, ships from 1.0.0 on, where dropping a system is a major, and record it in the decision "CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag".

The maintainer, 2026-10-26: until 1.0.0, a CI patch built with a new Go minor may drop a system the binaries ran on, and its changelog section names each one. The README's Versioning table makes dropping a system a major from 1.0.0. Roughly every second Go minor of 1.21 to 1.27 raised a macOS, Windows or Linux floor. Two choices: hold such a minor for a major, or let fejkdata's supported systems follow Go after 1.0.0 too, recorded as a lasting departure from the technical principles' semver.

### 69. Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.

The human cuts v1.0.0 once the shipped data is in its record shape and one full minor has shipped with no breaking change, per the README's Versioning table; pairing a street with its exact postnummer rewrites shipped rows, so a minor with no breaking change follows it first.

### 396. Decide whether `fejkdata` writes Go struct types, and TypeScript types for the npm package, from a record's columns.

From the same review, which asked for `fejkdata generate`. The decision "A record is a template seen as columns; a Go struct is the one second schema" runs the other way: `FakeStruct` takes columns from a struct a caller already wrote, so a generator would make a template a second source of that struct. Item 360 already types the npm package's API. Ask the maintainer whether a Go test author or a TypeScript author has asked for per-template types; if none has, drop the idea. If yes, add the goal or sub-goal, an audience line, and the items.

### 190. Decide whether fejkdata pseudonymises real datasets, and if so add the goal, the audience persona and the items below to the README and the plan.

Pseudonymising a copy of production is work a hosted model cannot do well: sending the data out is the leak, a model cannot keep one person's fake identical across tables and runs, and millions of rows cost real money. A model can still name which columns hold personal data and of what kind; fejkdata then replaces them offline. Greenmask, PostgreSQL Anonymizer and Presidio cover parts of this; none draws Swedish values a real validator accepts.

Proposed goal, its place among the others for the maintainer to choose:

> **It pseudonymises a real dataset on your own machine: each personal value becomes a valid fake one.**
> 1. Columns you do not name pass through byte for byte.
> 2. The same value under the same key becomes the same fake, in every table and every run; without the key, nobody can tell which real value a fake came from.
> 3. Values that belong to one person come from one pick, so a row's name, email and personnummer still agree.
> 4. Two values never become one fake in a column that must stay unique.

Proposed audience persona: a **pseudonymising author**, who replaces the personal data in a copy of production before it reaches a test or staging environment, and relies on one real value giving one fake across tables and runs.

Items to file once the goal is in:

1. Read CSV and NDJSON records on stdin, replace the columns a mapping names with a template's render, and stream every other column through unchanged. The flag or file spelling goes to a bench with the persona first.
2. Seed each replaced value's draw from HMAC-SHA256 under a key read from a file or the environment, never a flag. An unkeyed hash fails sub-goal 2: every personnummer can be hashed and the mapping reversed.
3. Key a whole row's pick by one identity column, such as a customer id, so goal 3.1's named pick holds for a person across rows and tables.
4. Redraw on a collision in a column marked unique, deterministically from the key, within a memory bound the README states.
5. Keep, on request, a personnummer's birth date or sex and a postal code's region, the docs naming what each kept field gives away.
6. Read a PostgreSQL `COPY` or MySQL `INSERT` dump.
7. Name the output pseudonymised in the README, never anonymised: under GDPR, data is personal while the key exists, and a rare mix of kept fields can still identify someone.

### 357. Create the `@larvit` organisation on npm, and make this repository's Packages workflow the trusted publisher of `@larvit/fejkdata`.

The npm account and its credentials are the maintainer's. Trusted publishing lets the workflow publish through GitHub's OIDC token, so no npm token is stored in the repository. If npm takes a trusted publisher only for a package that already exists, the first version goes out by hand from the tarball item 362's dry run packs, and every later one from CI. Item 362 needs this answered.

### 358. Record in `docs/decisions.md` the JavaScript API that a panel of JavaScript and TypeScript test authors picks for `@larvit/fejkdata`.

Seat the panel as the README's JavaScript or TypeScript test author, and bench the spelling of:

- creating a generator: an async factory, since a WebAssembly module instantiates asynchronously, and every call after it synchronous;
- the seed, which is a uint64 in Go: a `bigint`, a `number`, or either;
- the date, which goal 9.1 makes the caller's;
- choosing the shipped data a generator loads, such as `sv_SE` and `misc`, with every shipped module embedded;
- passing data of one's own, as files by path, which load as `WithDataFS` over an in-memory filesystem;
- the methods mirroring `Fake`, `FakeTemplate`, `NewTemplate`, `FakeRecord`, `FakeRecordTemplate`, `NewRecordTemplate`, `List` and `IsTemplate`;
- a record: its columns, its `json`, `csv` and `sql` serialisations, and a plain object of typed values, where an `integer` column beyond 2^53 cannot be a `number`;
- errors: what is thrown, and how a caller tells a load error, `ErrLoad` in Go, from misuse;
- module format: ESM only, or CommonJS as well.

`FakeStruct` has no JavaScript counterpart, since a JavaScript object carries no tags; a record template is the nearest. Functions of one's own stay Go only (goal 8.5).

### 359. Build the library to WebAssembly behind a bridge that reads only its public API, and export it from `@larvit/fejkdata` as item 358's decision spells it.

Measured on `main` at b8014f2, 2026-10-30, with `-ldflags="-s -w"`: `cmd/fejkdata` builds unchanged for `GOOS=js` and `GOOS=wasip1`, and `go vet` passes under `GOOS=js`. The wasm is 25.6 MB, 6.9 MB gzipped; native is 22.9 MB, 6.7 MB. Under Node 24's `node:wasi`, the CLI rendered a seeded `sv_SE.address` in about 450 ms, against 100 ms native, and 100,000 CSV rows of the README's `person.json` in 4.3 s, against 1.4 s.

The systems architect picks between `GOOS=js` with `syscall/js` and Go's `wasm_exec.js`, which runs in browsers without a WASI host, and `GOOS=wasip1` with `//go:wasmexport`, which needs one. The bridge reads only the library's public API, as the decision "The CLI reads only the library's public API" holds for the CLI, and lives in a folder named for what it holds, such as `javascript-package/`. The wasm builds on the latest supported Go only; the lowest supported Go stays the Go library's. Every shipped module is embedded, which goal 2.2 now allows. Measure the time from import to a generator's first render, for goal 12, and file an item if a test author would notice it.

### 360. Ship TypeScript declarations with `@larvit/fejkdata`, and fail the gate where a typed usage test stops compiling under `tsc --strict`.

The declarations are written by hand beside the bridge. The usage test calls every export the decision of item 358 names, and asserts the types of what each returns. `tsc` runs in a container in the gate, as a development dependency only, so goal 10 holds.

### 362. Publish `@larvit/fejkdata` to npm, with provenance, for every release CI cuts, patches included, and pack it in the Packages dry run.

Goal 1.1's counter-signal is a channel serving an old release, so the patch releases CI cuts for a new Go, a module or GoReleaser publish too. The package's version is the release tag's. A version npm has published cannot be published again, so the dry run on a pull request packs the tarball and checks its contents without publishing.

### 361. Test `@larvit/fejkdata` in the gate on Node, Bun, Deno and a headless browser, and pin that its seeded renders match the Go library's byte for byte.

Goal 9.2 holds only while a test checks it: render a fixed set of seeded templates, records included, in Go and through the package on each runtime, and fail on any byte that differs. Each runtime runs in a container pinned to its full patch version, as a target of `docker build .` or a compose service the README's Development section lists. The browser test loads the package as a page would, without a bundler.

### 363. Document `@larvit/fejkdata` in the README: its install line, a JavaScript section beside Library, the runtimes it is tested on and its Versioning row, with every JavaScript example run as a test.

The README's opening says a Go library and a CLI; it names the package once it ships. Goal 11.2 asks every example to run as a test, so the JavaScript section's examples run in the gate, as `readme_test.go` runs the `json` blocks. The Versioning table gains a row for the package: removing or renaming an export, changing a declared type or raising the lowest supported runtime is a major, and an export is a minor.

### 345. Decide how a locale with names in a script such as Han, Devanagari, Arabic or Bengali gets valid emails, and record it in the `slugify` decision.

The decision "A `slugify` transform makes emails from compound names valid, and the shipped data keeps the compounds" says: "A locale whose names need such letters romanizes them in its data, since `slugify` uses `ascii`'s table and no table of its own." This may conflict with goal 3.3, values look real. It may also conflict with items 206 (`zh_CN`), 207 (`hi_IN`), 209 (`ar_EG`) and 211 (`bn_BD`), whose names are in Han, Devanagari, Arabic or Bengali script. Such names slug to the empty string, so `{slugify(first)}.{slugify(last)}@example.com` becomes `.@example.com`.

The options to put to the maintainer:

1. A locale ships a romanized column beside each name for its email, such as pinyin or ISO 15919.
2. `slugify` gets a transliteration table beyond `ascii`'s.
3. A locale's email draws from a username-like category that holds no names.

The answer goes into `docs/decisions.md`. Option 1 also changes the locales' data, and option 2 the code of `slugify`.

### 86. Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.

Both phone formats carry the same 21-code `area` list, so an edit to one copy can miss the other.

### 243. Add `en_US.ein`, drawing only the prefixes the IRS assigns.

The IRS's prefix list is facts. [Source research](docs/research/research-sources-world.md).

### 210. Add `fr_FR` with its given names by sex, last names and titles, and `fr_FR.person`.

France holds more French speakers than DR Congo: 64M to 49M on Wikipedia's "Geographical distribution of French speakers", which credits a 2022 OIF report but flags its table as not taken from that report.

### 273. Add `fr_FR`'s national personal id, the NIR, the numéro de sécurité sociale, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 206. Add `zh_CN` with its given names by sex, last names and titles, and `zh_CN.person`.

Written in simplified characters.

### 277. Add `zh_CN`'s national personal id, the 18-digit resident identity card number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 208. Add `es_MX` with its given names by sex, last names and titles, and `es_MX.person`.

Mexico holds the most Spanish speakers.

### 282. Add `es_MX`'s national personal id, the CURP, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 212. Add `pt_BR` with its given names by sex, last names and titles, and `pt_BR.person`.

Brazil holds the most Portuguese speakers.

### 287. Add `pt_BR`'s national personal id, the CPF, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 207. Add `hi_IN` with its given names by sex, last names and titles, and `hi_IN.person`.

Hindi is written in Devanagari.

### 306. Add `hi_IN`'s national personal id, the Aadhaar number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 209. Add `ar_EG` with its given names by sex, last names and titles, and `ar_EG.person`.

Standard Arabic has no native speakers. Confirm from Ethnologue's country pages that Egypt holds the most second-language speakers, and record the answer in the decision.

### 310. Add `ar_EG`'s national personal id, the national ID number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 211. Add `bn_BD` with its given names by sex, last names and titles, and `bn_BD.person`.

Bangladesh holds the most Bengali speakers.

### 314. Add `bn_BD`'s national personal id, the NID number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 213. Add `id_ID` with its given names by sex, last names and titles, and `id_ID.person`.

Indonesian is spoken almost only in Indonesia.

### 318. Add `id_ID`'s national personal id, the NIK, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 214. Add `ur_PK` with its given names by sex, last names and titles, and `ur_PK.person`.

Urdu is Pakistan's national language, yet India has more native speakers: 50.8M by its 2011 census against Pakistan's 22.3M by its 2023 one. Confirm from Ethnologue's country pages that Pakistan holds the most speakers in all, else the locale is ur_IN, and record the answer in the decision.

### 322. Add `ur_PK`'s national personal id, the CNIC number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 332. Add `de_DE`'s national personal id, the Steuer-Identifikationsnummer, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 245. Draw `en_US.phone` from every area code NANPA has in service.

Today it draws 21 area codes; the exchange already ships as `en_US.phone-exchange`. No research file covers NANPA yet: find its list of area codes in service and that list's licence first.

### 250. Add `en_US.month` and `en_US.weekday` names from CLDR, and `en_US.holiday`, the federal holidays.

CLDR is under the Unicode licence. [Source research](docs/research/research-sources-world.md). The federal holidays are listed in 5 U.S.C. 6103, which no research file covers yet.

### 244. Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.

NAICS 2022 is public domain; SEC's `company_tickers.json` is free to reuse. [Source research](docs/research/research-sources-world.md).

### 198. Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.

`streets()` in `geo-us.py` fetches only each drawn place's main county, the one holding most of its people. A ZCTA in a place's other counties then has no address range and is dropped: New York, under Kings County, ships Brooklyn's codes, one Manhattan code and none on Staten Island, and its streets are mostly Brooklyn's. Whether a neighbouring county is fetched depends on which other places the draw picks. The fix reads every county in `localities`' `county_part`.

### 272. Add `fr_FR.phone`, drawn from France's numbering plan.

Find the numbering plan and its licence first.

### 276. Add `zh_CN.phone`, drawn from China's numbering plan.

Find the numbering plan and its licence first.

### 281. Add `es_MX.phone`, drawn from Mexico's numbering plan.

Find the numbering plan and its licence first.

### 286. Add `pt_BR.phone`, drawn from Brazil's numbering plan.

Find the numbering plan and its licence first.

### 305. Add `hi_IN.phone`, drawn from India's numbering plan.

Find the numbering plan and its licence first.

### 309. Add `ar_EG.phone`, drawn from Egypt's numbering plan.

Find the numbering plan and its licence first.

### 313. Add `bn_BD.phone`, drawn from Bangladesh's numbering plan.

Find the numbering plan and its licence first.

### 317. Add `id_ID.phone`, drawn from Indonesia's numbering plan.

Find the numbering plan and its licence first.

### 321. Add `ur_PK.phone`, drawn from Pakistan's numbering plan.

Find the numbering plan and its licence first.

### 331. Add `de_DE.phone`, drawn from Germany's numbering plan.

Find the numbering plan and its licence first.

### 248. Add `en_US.car`, a make and model from NHTSA vPIC.

NHTSA vPIC is public domain. The decision "`misc` is what every locale shares" puts a national register's cars in its locale. [Source research](docs/research/part2-internet-vehicle-airline.md).

### 246. Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.

No national register defines the account number, and no research file covers Nacha yet. Item 154 adds the routing number beside it.

### 247. Add `en_US.licenseplate`, drawn per state from that state's pattern.

No open list of state patterns exists: write each state's pattern by hand and check it against that state's DMV. Wikipedia's list (CC BY-SA) and openalpr's (AGPL) may be read, never copied. [Source research](docs/research/research-sources-world.md).

### 270. Add `geo/FR` with France's places and postal codes, and `fr_FR.address` printing them.

The Base Adresse Nationale (Licence Ouverte 2.0) names each address's commune and code postal.

### 275. Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.

`rows.Pins` holds eight pins inline, enough for a five-deep geo tree. If CN's tree runs deeper, raise that number to keep a render off the heap. Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 279. Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.

SEPOMEX publishes every código postal with its asentamiento.

### 284. Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.

IBGE's CNEFE lists every address with its CEP.

### 304. Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.

India's PIN code directory is open data on data.gov.in. It ships without streets (decision: no open address register covering it whole).

### 308. Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 312. Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 316. Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 320. Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 329. Add `geo/DE` with Germany's places and postal codes, and `de_DE.address` printing them.

GeoNames publishes Germany's postal codes.

### 271. Give `geo/FR`'s places their streets.

The Base Adresse Nationale (Licence Ouverte 2.0) holds every street with its code postal.

### 280. Give `geo/MX`'s places their streets.

INEGI's Marco Geoestadístico names the streets.

### 285. Give `geo/BR`'s places their streets.

IBGE's CNEFE lists every address with its CEP.

### 330. Give `geo/DE`'s places their streets.

Each Land publishes its own address data. Where open registers cover every Land, they are its street source; else it ships without streets (decision: no open address register covering it whole), and this item goes.

### 432. Reword goals 2.2, 7, 8.1, 8.5, 8.6, 9.2 and 10, and the README's Versioning table, for a Rust core that a native Go library and the npm package both serve.

The owner decided on 2026-11-11 that fejkdata moves to Rust as a whole and keeps Go as a native library. Today those goals name Go as the one language: Go packages as modules, a CLI of your own in a few lines of Go, functions in Go, the JavaScript package matching the Go library, and the lowest supported Go in the Versioning table. The wording is the owner's; every other item of 2.0.0 waits for it.

While rewording, weigh two points the architecture pass of 0.5.0's release review raised. Item 446 serves the Go library through cgo, so `go install` fails without a C toolchain, which works against goals 8.6 and 10 as they read today. Item 359, the Go library built to WebAssembly, would serve goal 1.2 with one engine, without the Rust core.

### 433. Measure a seeded render of the README's person.json, one row and 100,000 rows, from Rust compiled to WebAssembly under wazero against the native Go library, and record whether wazero comes within a few percent.

The owner accepts a Go library that runs the Rust core as WebAssembly only where it costs no more than a few percent against native Go. Expect 1.5 to 3 times slower before the cost of copying strings across the boundary. Where it does come within a few percent, item 446 binds through wazero instead of cgo, and Go users keep `CGO_ENABLED=0`.

### 434. Add a Cargo workspace with a compose service, and a gate leg running rustfmt, clippy with a complexity limit of 14, the tests, and cargo-mutants for the stability pass.

Rust runs through `docker compose run --rm`, as Go does. Clippy's cognitive-complexity lint stands in for the Go gate's cyclomatic limit. cargo-mutants on the changed lines will be slower than Go's mutation run, since Rust compiles more slowly; measure one run and say so in the Development section.

### 435. Port internal/drawstate to Rust so a seed draws the same numbers as Go, and fail the gate where the Go and Rust CLIs print different bytes for a corpus of templates at a fixed seed and date.

Go's `math/rand/v2` PCG is a 128-bit LCG step with a DXSM output. `IntN` is Lemire's multiply method, and `Float64` takes 53 bits of one draw. That comes to about 60 lines of Rust over `u128`. Every draw goes through `drawstate`'s `IntN`, `Float64` and `Weighted`. Seeds then stay as they are across 2.0.0, and the Go build serves as the port's test oracle, provided the Rust engine draws in the same order. The corpus takes the README's examples, `testdata/`, and the templates the Go tests render. A difference in `calc`'s float formatting may be accepted as a break and listed in item 450.

### 436. Port internal/datafiles, internal/rows and internal/jsonvalue to Rust, loading on first reach a source with an index, and embed the shipped data per locale behind a Cargo feature.

The data under `data/` stays as it is, JSON and TSV. A program that enables one locale's feature carries only that locale's data and its country's geo data, as goal 2.2 asks of Go today.

### 438. Port internal/engine/tree and internal/engine/load to Rust: compile, bind, link and resolve.

This is a milestone. Split it into two or three items before starting. It waited for items 421 and 423, now shipped, because the places the comprehension seats find hardest sit in the code this item and items 439 and 440 port: `namefence.go`, `pick.go:87` `landsInPick`, `path.go` and `table.go:250` `route`. Lifting Locality, the lowest dimension at about 5.5, in Go first, where the Go tests check the change, let this item translate the cleaner design rather than redesign and translate at once.

### 444. Fail the Rust gate where a render allocates more than its pinned count, as the Go performance gate asserts allocations.

A counting global allocator in the test binary stands in for Go's `testing.AllocsPerRun`.

### 441. Port internal/builtinfunc to Rust: every builtin, Go's reference-time date layouts such as '060102', the checksums, calc with its Inf, and reword.

The date layouts are template syntax users write, so they keep Go's spelling.

### 446. Serve the Go library from the Rust core through cgo, under the module path github.com/larvit/fejkdata/v2, keeping New, its options, Fake, the templates, records and FakeStruct.

The owner wants Go to stay a native library. cgo's per-call cost, tens of nanoseconds, disappears inside a render. The module ships a prebuilt static library for each of the six targets. Go users then need `CGO_ENABLED=1` and a C linker, a break item 450 lists. `FakeStruct` stays Go reflection that builds a record template. A data package passes its embedded files into the core, so goal 2.2 holds. A function of one's own runs through a callback from Rust into Go, and its signature may change. Where item 433 finds wazero within a few percent, bind through wazero instead.

### 445. Build and publish the Rust CLI on every channel the Go one reaches: the binaries for six targets, deb, rpm and apk, Homebrew, Scoop and the image, with cargo-audit in place of govulncheck.

GoReleaser gives way to cargo builds for the six targets. `patch_release.py` cuts a patch on a new Rust toolchain or crate where it does on a new Go or module today. The decision on known vulnerabilities blocking a release carries over to RustSec advisories.

### 448. Build @larvit/fejkdata's WebAssembly from the Rust core with wasm-bindgen, in place of the Go build item 359 makes.

The owner's reasons for the move include the WebAssembly's speed, memory and size. The Go build item 359 measured runs about 4.5 times slower than native under Node's WASI. Items 360 to 363 carry over to the Rust build unchanged.

### 450. Write MIGRATION.md, listing each change 2.0.0 makes to what a user relies on and what to do about it.

The known breaks so far: the Go library needs cgo (`CGO_ENABLED=1` and a C linker) on the six targets only; the module path takes `/v2`; `go install …/cmd/fejkdata` stops installing the CLI; a Go function of one's own may change signature; error wording may change; and possibly `calc`'s float formatting. Each item of 2.0.0 that breaks something adds it here.

### 439. Port internal/engine/fence to Rust: every proof and fence a load runs, with errors that name the same path, spelling and fix as Go's.

An error is a contract by what it names, not its bytes, so the wording may change in 2.0.0.
