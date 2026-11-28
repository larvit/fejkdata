# todo

## Scoring

`Score = 100 * V / (2**S + 8)`

`Bar = 0`

`Next ID = 613`

## Renamed

- 0.8.1 renamed to 0.9.0

## Items

| ID | Release | Blocked by | Exempt | Item | L | E | I | R | S | Moves | Score |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 611 | 0.9.0 |  |  | **Lower the reading cost of the files a chunk reads most before its first commit.** | 2 | 2 | 1 | 3 | 3 | code health by 4 | 7.8 |
| 612 | 0.9.0 |  | owner | **Run the `npm-package-lowest` stage on Node 22.3.0, the lowest `engines` names, and keep Renovate from bumping it.** | 1 | 2 | 2 | 3 | 1 | 1.2 by 3 | 2.1 |
| 363 | 0.9.0 |  |  | **Document `@larvit/fejkdata` in the README: its install line, a JavaScript section beside Library, the runtimes it is tested on and its Versioning row, with every JavaScript example run as a test.** | 2 | 2 | 1 | 3 | 4 | 1.1 by 2, 1.2 by 4 | 1.8 |
| 603 | 0.10.0 |  |  | **Add a mutation tool for JavaScript.** | 3 | 1 | 1 | 3 | 3 | code health by 4 | 7.8 |
| 606 | 0.10.0 |  |  | **Compile and run each `go` block of the README in a test, so an example that fails to load fails the gate.** | 2 | 2 | 2 | 4 | 3 | 5 by 3, code health by 2 | 4.9 |
| 567 | 0.10.0 |  |  | **Ship a home for odd values, such as `misc.edge.name`, whose rows each say which assumption they break, with invalid values only under a path that says `invalid`.** | 2 | 1 | 2 | 3 | 4 | 1 by 5, 3 by 4, 4 by 2 | 3.1 |
| 569 | 0.10.0 | 567 |  | **Ship odd names in Han, Arabic, Cyrillic, Greek, Hebrew, Devanagari and Thai, without waiting for their locales.** | 1 | 2 | 1 | 2 | 3 | 1 by 4, 2.1 by 3 | 3.1 |
| 252 | 0.10.0 |  |  | **Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.** | 1 | 3 | 1 | 3 | 1 | 1.1 by 4 | 2.9 |
| 568 | 0.10.0 | 567 |  | **Ship odd Latin-script names: one name, one letter, over 100 characters, particles, apostrophes, two surnames, display orders, and words such as `Null`.** | 1 | 2 | 1 | 2 | 3 | 1 by 4, 2.1 by 2 | 2.7 |
| 220 | 0.10.0 |  |  | **Suggest the nearest name where a path segment names nothing: `no entry "adress"; did you mean "address"?`.** | 1 | 1 | 1 | 1 | 3 | 1 by 2, 5 by 5 | 2.7 |
| 30 | 0.10.0 |  |  | **Draw `email.local` from `username`, so the two share one handle list.** | 1 | 1 | 1 | 1 | 2 | code health by 1 | 2.6 |
| 534 | 0.10.0 |  | owner | **Reach the comprehension floor: lift Locality so the rule that a name keeps one pick has one home that render and the name fence both call, and render reaches each child only through the edges the fences walk.** | 4 | 2 | 2 | 6 | 6 | code health by 6 | 2.6 |
| 255 | 0.10.0 |  |  | **Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.** | 2 | 1 | 2 | 3 | 3 | 1 by 1, 3 by 3, 2 by 2 | 2.5 |
| 117 | 0.10.0 |  |  | **List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 4 | 2.5 |
| 604 | 0.10.0 |  | decision | **Carry a parent's null, tab and line break whole to the children of a run, and accept a one-column parent that draws an empty or null value.** | 3 | 2 | 3 | 6 | 3 | 1 by 3, 5 by 3 | 2.5 |
| 596 | 0.10.0 |  | owner | **Publish an article on why Nordic test data needs national ids and addresses that match, and post fejkdata where Go and testing developers ask for tools.** | 1 | 2 | 1 | 2 | 2 | 1.1 by 4 | 2.4 |
| 168 | 0.10.0 |  | principle | **Test in the merge gate the shared steps of `data-import/`, starting with `geo.fence`, the småort steps of `geo-se.py`, and the name-shape rules of `company-no.py` and `company_names.py`.** | 3 | 1 | 1 | 3 | 5 | code health by 3 | 2.3 |
| 457 | 0.10.0 |  |  | **Rename each script under `data-import/` to a name `import` can spell, so a test imports it like any module and the mutation tests can run on the folder.** | 3 | 1 | 1 | 3 | 5 | code health by 3 | 2.3 |
| 257 | 0.10.0 |  |  | **Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 2 by 3, 3 by 3 | 2.2 |
| 389 | 0.10.0 |  |  | **Add `misc.rrule`, recurrence rules a test checks against RFC 5545's grammar.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2 by 1, 3 by 2 | 2.2 |
| 530 | 0.10.0 |  |  | **Draw `en_US.ssn` from the series the SSA never issues, area `000`, `666` or `900` to `999`, and say in the README which checker refuses it.** | 2 | 2 | 2 | 4 | 3 | 3.3 by 6 | 2.2 |
| 406 | 0.10.0 |  |  | **Add a transform that writes a value escaped for a JSON string, so a text template spelling JSON stays valid whatever a drawn value holds.** | 2 | 1 | 1 | 2 | 2 | 1 by 2, 4 by 2 | 2.2 |
| 570 | 0.10.0 | 567 |  | **Ship odd valid emails, and invalid ones under an `invalid` path.** | 1 | 1 | 1 | 1 | 3 | 1 by 3, 3 by 2 | 2.2 |
| 260 | 0.10.0 |  |  | **Add `misc.element`, `misc.unit` and `misc.planet`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 2 | 2.1 |
| 261 | 0.10.0 |  |  | **Add `misc.animal` by class, and `misc.plant`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 2 | 2.1 |
| 164 | 0.10.0 |  |  | **Let a table's format or a cell bind a name, which lives in that format or cell and keeps one pick through the draw.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 4 by 3 | 2.1 |
| 262 | 0.10.0 |  |  | **Add `misc.bloodtype`, weighted by its global distribution.** | 1 | 2 | 1 | 2 | 1 | 1 by 1, 2 by 1, 3.2 by 1 | 2.0 |
| 50 | 0.10.0 |  |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 52 | 0.10.0 |  |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 92 | 0.10.0 |  |  | **Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 219 | 0.10.0 |  |  | **Name the path and the Go type where a struct field's type cannot hold what its tag renders: `sv_SE.person renders text, which int cannot hold`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 221 | 0.10.0 |  |  | **Refuse two references under a record `--format` by saying a run renders one record: today `{/users}{/orders}` reports that an inline record has no fields.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 224 | 0.10.0 |  |  | **Name the module in every error about its manifest or its files: today an `fs.FS` module's errors name only the file inside it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 228 | 0.10.0 |  |  | **Offer `{/path as p}` where a token starts with a name nothing binds and no loaded path starts with, in a record's format too: today `{p.first}` with `p` unbound suggests `{/p.first}`.** | 2 | 1 | 1 | 2 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 229 | 0.10.0 |  |  | **Name the line and column where JSON parsing stopped, when a template that opens like a JSON object then fails as a format string.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 3 | 2.0 |
| 602 | 0.10.0 |  |  | **Render a second template from the npm package in under 10 ms, today about 100 ms, by keeping the loaded data between renders.** | 2 | 1 | 2 | 3 | 3 | 1.2 by 3, 12 by 2 | 2.0 |
| 38 | 0.10.0 |  | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 2 | 2 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 44 | 0.10.0 |  | decision | **Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 45 | 0.10.0 |  |  | **Name the node a selector follows in `childNamed`'s refusal.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 49 | 0.10.0 |  |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 110 | 0.10.0 |  |  | **Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 225 | 0.10.0 |  |  | **Drop "a field cannot be named weight" from the refusal of a `weight` that is no number, and name the column a table's `weight` reads.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 5 by 2 | 1.9 |
| 355 | 0.10.0 |  |  | **Run `docker build .` and the Packages dry run on a release commit.** | 2 | 2 | 2 | 4 | 2 | 1.1 by 3 | 1.8 |
| 181 | 0.10.0 |  | decision | **Fail a load where two sources define one category, unless the library's option or the CLI's flag says the later replaces the earlier, and then make every read reach the replacement.** | 4 | 1 | 3 | 6 | 4 | 1 by 2, 5 by 5 | 1.8 |
| 81 | 0.10.0 |  | decision | **Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.** | 3 | 1 | 3 | 5 | 4 | 1 by 1, 5 by 4, 3 by 2 | 1.8 |
| 35 | 0.10.0 |  |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 1.7 |
| 265 | 0.10.0 |  |  | **Add `misc.commit` messages, branch names and `sha`s.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 1.7 |
| 258 | 0.10.0 |  |  | **Add `misc.product` names, adjectives, materials and departments, and `misc.garmentsize`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 3 | 1.7 |
| 266 | 0.10.0 |  |  | **Add `misc.password`, `hash`, `jwt`, `slug`, `hostname`, `avatar` and `placeholderimage`.** | 1 | 1 | 1 | 1 | 3 | 1 by 1, 2 by 3 | 1.7 |
| 605 | 0.10.0 |  |  | **Run linked templates from the Python and npm packages, with a function that takes each template and its count.** | 2 | 2 | 3 | 5 | 3 | 1.3 by 4 | 1.7 |
| 196 | 0.10.0 |  | decision | **Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 5 by 6 | 1.7 |
| 263 | 0.10.0 |  |  | **Add `misc.color` names with their hex, rgb, hsl and cmyk forms.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 3.2 by 1 | 1.7 |
| 405 | 0.10.0 |  |  | **Let `repeat` take a range, such as 5 to 15, drawing the count on each render.** | 2 | 1 | 1 | 2 | 3 | 1 by 2, 4 by 2 | 1.7 |
| 13 | 0.10.0 |  | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 3.2 by 2 | 1.6 |
| 165 | 0.10.0 |  |  | **Let a table carry `name` with neither `key` nor `parent`, and list the rows a name matches by their line where it matches several.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 2 | 1.6 |
| 166 | 0.10.0 |  |  | **Let a key or name cell hold a bracket, a brace, a double quote or a pipe, so the table loads, its row draws at random, and only a selector cannot reach it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 2 | 1.6 |
| 39 | 0.10.0 |  |  | **Stop the `fence.FieldKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 1.6 |
| 42 | 0.10.0 |  |  | **Refuse a struct tag's `datatype` other than its Go type's before proving its values, as the README's Library section promises.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 1.6 |
| 46 | 0.10.0 |  |  | **Report the same error every load for a table with two bad options.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 1.6 |
| 51 | 0.10.0 |  |  | **Spell a table one way across the errors that name it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 1.6 |
| 115 | 0.10.0 |  |  | **Spell a rows file's line one way in every error, naming the file.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 1.6 |
| 522 | 0.10.0 |  | decision | **Rebuild `et_EE`'s first and last names from Statistics Estonia's CC BY-SA 4.0 counts by sex, replacing the language model's recall.** | 2 | 2 | 2 | 4 | 4 | 3.2 by 6 | 1.5 |
| 572 | 0.10.0 | 567 |  | **Ship the Big List of Naughty Strings by kind, with Unicode edge cases, complex emoji and text that looks like another type.** | 2 | 1 | 1 | 2 | 4 | 1 by 4, 3.2 by 1 | 1.5 |
| 259 | 0.10.0 |  |  | **Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 2 by 4, 3.2 by 3 | 1.5 |
| 610 | 0.10.0 |  | owner | **Set `@larvit/fejkdata`'s publishing access to "Require two-factor authentication and disallow tokens", and check its live page, once 0.9.0 is published.** | 1 | 1 | 2 | 2 | 1 | 1.1 by 2 | 1.5 |
| 571 | 0.10.0 | 567 |  | **Ship odd dates, prices and numbers at their boundaries.** | 1 | 1 | 1 | 1 | 3 | 1 by 3 | 1.4 |
| 267 | 0.10.0 |  |  | **Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.** | 1 | 2 | 1 | 2 | 4 | 1 by 1, 2 by 2, 3.2 by 2 | 1.4 |
| 32 | 0.10.0 |  |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 2 | 1 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.3 |
| 264 | 0.10.0 |  |  | **Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.3 |
| 268 | 0.10.0 |  |  | **Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 1.3 |
| 387 | 0.10.0 |  |  | **Let a template draw a number from a bounded normal distribution, so durations and start times cluster without weighted ranges.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 4 by 1 | 1.3 |
| 109 | 0.10.0 |  | principle | **Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.** | 1 | 2 | 1 | 2 | 4 | code health by 1 | 1.3 |
| 91 | 0.10.0 |  |  | **Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 4 by 4 | 1.3 |
| 23 | 0.10.0 |  |  | **Give every other shipped category that could reach something real a path that never does.** | 2 | 2 | 2 | 4 | 5 | 1 by 2, 3.3 by 6 | 1.3 |
| 158 | 0.10.0 |  |  | **Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.** | 1 | 1 | 1 | 1 | 2 | 1 by 2 | 1.3 |
| 163 | 0.10.0 |  |  | **Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.** | 2 | 1 | 2 | 3 | 2 | 1 by 2 | 1.3 |
| 194 | 0.10.0 |  | decision | **Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 5 by 4 | 1.2 |
| 496 | 0.10.0 |  |  | **Name each `geo.DE` Kreis in one style, as Destatis's Gemeindeverzeichnis names it.** | 2 | 2 | 1 | 3 | 3 | 1 by 1, 3.2 by 2 | 1.2 |
| 573 | 0.10.0 | 567, 568, 569, 570, 571, 572 |  | **Map each falsehood about names, addresses, emails, dates and text to the path that tests it in `docs/edge-cases.md`, and show in the README how to mix odd values in.** | 1 | 2 | 1 | 2 | 2 | 1.1 by 2 | 1.2 |
| 269 | 0.10.0 |  |  | **Add `misc.direction`, `misc.continent` and `misc.ulid`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 1 | 1.2 |
| 361 | 0.10.0 |  |  | **Test `@larvit/fejkdata` in the gate on Node, Bun, Deno and a headless browser, and pin that its seeded renders match the Go library's byte for byte.** | 3 | 1 | 2 | 4 | 4 | 1.2 by 4 | 1.2 |
| 161 | 0.10.0 |  |  | **Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 1 | 1.1 |
| 497 | 0.10.0 |  | owner | **Draw every word of each locale's `word` tables from its dictionary, rude words included.** | 3 | 2 | 2 | 5 | 6 | 3.2 by 2, code health by 2 | 1.0 |
| 29 | 0.10.0 |  |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 4 by 3 | 1.0 |
| 544 | 0.10.0 |  |  | **Find out whether a very high repeat limit, nested repeats included, can render in bounded memory, and write down what it takes.** | 3 | 3 | 2 | 6 | 3 | 12 by 3 | 1.0 |
| 36 | 0.10.0 |  |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 3 | 1 | 2 | 4 | 6 | 1 by 3, 5 by 7 | 0.9 |
| 495 | 0.10.0 |  |  | **Weigh `geo.DE`'s places and Länder by population, not by the postal codes they hold.** | 2 | 2 | 2 | 4 | 5 | 1 by 2, 3.2 by 3 | 0.8 |
| 56 | 0.10.0 |  | defect | **Build on a manual run of `test.yml` on `main`.** | 1 | 2 | 1 | 2 | 1 | 1.1 by 1 | 0.7 |
| 53 | 0.10.0 |  |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 3 | 1 | 2 | 4 | 5 | 1 by 2, 4 by 2 | 0.7 |
| 404 | 0.10.0 |  |  | **Let a template in a `repeat` or a choice item bind a name to a path through an enclosing name, `{l.street as s}`, so each render draws a fresh row inside the enclosing pick.** | 3 | 2 | 2 | 5 | 6 | 1 by 3, 4 by 4 | 0.6 |
| 55 | 0.10.0 |  |  | **Document `ErrNoColumns` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 1 | 2 | 2 | 1 by 1 | 0.6 |
| 75 | 0.10.0 |  |  | **Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.** | 2 | 1 | 1 | 2 | 5 | 1 by 1, 4 by 3 | 0.6 |
| 17 | 0.10.0 |  |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 3 | 1 | 3 | 5 | 6 | 1 by 1, 4 by 3, 2 by 2 | 0.5 |
| 14 | 0.10.0 |  |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`.** | 1 | 1 | 1 | 1 | 2 | 4 by 1 | 0.5 |
| 33 | 0.10.0 |  |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 3 | 1 | 3 | 5 | 5 | 1 by 1, 4 by 2 | 0.5 |
| 193 | 0.10.0 |  | decision | **Move the largest shipped packages into nested Go modules that keep their import paths, once `TestShippedDataStaysUnder100MiB` fails.** | 4 | 2 | 3 | 7 | 6 | 1 by 1, 2 by 3 | 0.4 |
| 407 | 0.10.0 |  |  | **Let a choice deal its items in exact counts across a run's renders, so 200 issues hold exactly 4 in progress, where weights only approach it.** | 3 | 2 | 2 | 5 | 6 | 1 by 3 | 0.3 |
| 218 | 0.10.0 |  | decision | **Replace each country's invented streets with its real streets, from OpenStreetMap under ODbL or from the country's register.** | 3 | 2 | 2 | 5 | 7 | 1 by 1, 2.1 by 4 | 0.2 |
| 129 | 0.10.0 |  | decision | **Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.** | 4 | 1 | 3 | 6 | 8 | 1 by 3, 4 by 7 | 0.2 |
| 41 | 0.10.0 |  | decision | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 4 | 1 | 3 | 6 | 7 | 1 by 1, 4 by 4 | 0.2 |
| 186 | 0.10.0 |  | decision | **Load a module carrying functions through an option of its own, its functions reaching the generator's randomness and date through an interface the root declares.** | 4 | 1 | 3 | 6 | 7 | 1 by 3 | 0.2 |
| 67 | 0.11.0 |  |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 2 | 2 | code health by 2 | 5.2 |
| 64 | 0.11.0 |  |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 2 | 1 | 2 | 4 | code health by 2 | 2.6 |
| 578 | 0.11.0 | 567 |  | **Draw number ranges, fractions, `bis`, `ter` and `quater`, and house names as odd house numbers.** | 2 | 1 | 2 | 3 | 3 | 2.1 by 2, 1 by 3 | 2.2 |
| 576 | 0.11.0 |  |  | **Offer a gender value that can be nonbinary, undisclosed or unknown, beside the two-valued `sex` that names and ids read.** | 2 | 1 | 2 | 3 | 3 | 1 by 3, 2.1 by 1 | 1.8 |
| 577 | 0.11.0 |  |  | **Draw apartments, floors, staircases, c/o and company lines, PO boxes and poste restante as optional address parts in each country.** | 3 | 2 | 2 | 5 | 6 | 2.1 by 5, 1 by 4 | 0.9 |
| 580 | 0.11.0 | 577 |  | **Offer addresses of one, three, four and five lines, with and without a country line.** | 2 | 1 | 2 | 3 | 3 | 2.1 by 2, 1 by 2 | 1.8 |
| 583 | 0.11.0 | 567 |  | **Draw `en_US.phone.international`, and ship odd phone numbers: extensions, vanity letters, `00` prefixes, odd spacing, non-ASCII digits, and invalid ones.** | 2 | 2 | 2 | 4 | 4 | 2.1 by 2, 1 by 3 | 1.5 |
| 581 | 0.11.0 | 567 |  | **Ship odd addresses from countries outside the shipped ones.** | 1 | 2 | 1 | 2 | 3 | 1 by 3 | 1.4 |
| 584 | 0.11.0 |  |  | **Draw a datetime in any zone with its true offset that day, and the local times that a clock change skips or repeats.** | 3 | 1 | 2 | 4 | 5 | 1 by 5, 3 by 3 | 1.4 |
| 62 | 0.11.0 |  |  | **Add `{btc()}` and `{eth()}`.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 2 by 2, 3 by 2 | 1.4 |
| 582 | 0.11.0 | 567 |  | **Draw the longest, the shortest and the non-ASCII places, and the postal codes that many places share, by path.** | 2 | 1 | 1 | 2 | 3 | 1 by 2, 3 by 1 | 1.3 |
| 574 | 0.11.0 |  |  | **Draw a middle name, an initial, a suffix and a title after the name in each locale's `person` where people there use them.** | 3 | 1 | 2 | 4 | 5 | 2.1 by 4, 1 by 3 | 1.2 |
| 575 | 0.11.0 |  |  | **Offer a person's previous, preferred and legal name as values that agree with the person.** | 2 | 1 | 2 | 3 | 4 | 2.1 by 2, 1 by 2 | 1.2 |
| 579 | 0.11.0 |  |  | **Ship localities with no street, with an address that names the locality and a number.** | 3 | 2 | 2 | 5 | 5 | 2.1 by 3, 3.2 by 2 | 0.8 |
| 59 | 0.11.0 |  |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 2 | 1 | 2 | 1 | 1 by 1 | 0.8 |
| 63 | 0.11.0 |  |  | **Pair a street with its exact postnummer.** | 3 | 2 | 3 | 6 | 6 | 1 by 2, 3 by 6 | 0.7 |
| 65 | 0.11.0 |  |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 2 | 1 | 3 | 3 | 1 by 1 | 0.5 |
| 587 | 0.12.0 | 567 | question | **Decide whether an `--edge 5%` flag mixes odd values into any template, once the odd-value tables have been in use.** | 1 | 2 | 1 | 2 | 1 | 1 by 3 | 2.3 |
| 589 | 0.12.0 | 572 |  | **Load every odd value into PostgreSQL and MySQL from each format that can carry it in the merge gate, and compare what comes back.** | 2 | 1 | 1 | 2 | 4 | 3 by 2, code health by 2 | 3.1 |
| 593 | 0.12.0 | 567 |  | **Ship national ids, IBANs and cards with a bad checksum or wrong length under an `invalid` path, and French NIRs of people born abroad.** | 2 | 1 | 1 | 2 | 4 | 1 by 4, 3 by 2 | 1.8 |
| 585 | 0.12.0 |  |  | **Add `nfc` and `nfd` transforms.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 4 by 1 | 1.3 |
| 591 | 0.12.0 | 567 |  | **Ship skipped days and ambiguous date spellings.** | 1 | 1 | 1 | 1 | 2 | 1 by 2 | 1.3 |
| 592 | 0.12.0 | 567 |  | **Ship odd company names, such as `! LTD`, script-like names, emoji, digits only and very long ones.** | 1 | 1 | 1 | 1 | 2 | 1 by 2 | 1.3 |
| 590 | 0.12.0 | 567 |  | **Ship odd URLs, file names, version numbers and IP spellings.** | 1 | 1 | 1 | 1 | 3 | 1 by 2 | 0.9 |
| 595 | 0.12.0 |  |  | **Let a CLI flag make CSV cells that start a spreadsheet formula harmless.** | 2 | 1 | 2 | 3 | 3 | 1 by 2 | 0.9 |
| 588 | 0.12.0 | 567 |  | **Draw each row of a table exactly once, so a test run covers every odd value rather than some by chance.** | 3 | 1 | 2 | 4 | 5 | 1 by 4, 4 by 1 | 0.9 |
| 594 | 0.12.0 |  |  | **Give `misc.territory` every currency and calling code a territory uses, and add Kosovo and Antarctica.** | 3 | 2 | 2 | 5 | 5 | 3.2 by 3, 1 by 2 | 0.8 |
| 586 | 0.12.0 |  |  | **Add transforms that make a value messy: stray spaces, doubled spaces, a case flip, a typo and broken encoding.** | 2 | 1 | 2 | 3 | 5 | 1 by 4 | 0.8 |
| 472 | 0.13.0 |  | question | **Decide how the shipped data is split into files and modules before it reaches the first size wall.** | 1 | 2 | 1 | 2 | 1 | 2 by 3 | 2.0 |
| 466 | 0.13.0 |  | decision | **Number a pick's memo slots at load and index child rows by parent row, so a render allocates no per-pick map.** | 3 | 1 | 2 | 4 | 5 | 12.3 by 4, 12.1 by 2 | 3.0 |
| 471 | 0.13.0 |  | decision | **Prove each shipped module in its own test package, so the test cache skips the data that didn't change.** | 2 | 2 | 1 | 3 | 4 | code health by 2, 1 by 1 | 2.9 |
| 467 | 0.13.0 |  | decision | **Render a record into one reused buffer, and write `CSVLine` into one buffer, so a row costs one allocation for its text.** | 2 | 1 | 2 | 3 | 3 | 12.3 by 3 | 0.9 |
| 469 | 0.13.0 | 466, 467 | decision | **Render `--repeat` rows on every CPU, each row drawn from a seed derived from the run's seed and its row number, so the bytes don't depend on the number of CPUs.** | 4 | 2 | 3 | 7 | 8 | 12.3 by 8 | 0.1 |
| 473 | 0.13.0 | 469 | decision | **Publish a bulk-throughput benchmark beside gofakeit, faker-js, Mimesis and Faker, with the command that reproduces each row.** | 3 | 2 | 1 | 4 | 4 | 1 by 3, 12.3 by 3 | 1.6 |
| 468 | 0.13.0 |  | decision | **Set the CLI's GC target to 400 unless `GOGC` is set.** | 1 | 2 | 2 | 3 | 1 | 12.3 by 3 | 1.5 |
| 465 | 0.13.0 |  | decision | **Load a table's rows under one parent on that parent's first reach, from an index `go generate` writes beside the manifest, for every large table.** | 4 | 2 | 3 | 7 | 8 | 12.2 by 8 | 0.2 |
| 470 | 0.13.0 | 465 | decision | **Hold a parsed block of a table as its text and its row offsets, not one string per cell.** | 3 | 1 | 3 | 5 | 6 | 1 by 2 | 0.2 |
| 205 | 0.14.0 |  |  | **Add `is_IS` with its given names by sex, last names and titles, and `is_IS.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 327 | 0.14.0 | 205 |  | **Add `is_IS`'s national personal id, the kennitala, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 326 | 0.14.0 | 205 |  | **Add `is_IS.phone`, drawn from Iceland's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 328 | 0.14.0 | 205 |  | **Add `is_IS.company`, `is_IS.email` and `is_IS.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 324 | 0.14.0 | 205 |  | **Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 325 | 0.14.0 | 324 |  | **Give `geo/IS`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.1 |
| 348 | 1.0.0 |  | question | **Decide how a Go minor that drops a system the binaries run on, such as an older macOS, ships from 1.0.0 on, where dropping a system is a major, and record it in the decision "CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag".** | 1 | 2 | 1 | 2 | 1 | 1.1 by 3 | 2.2 |
| 69 | 1.0.0 |  |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 2 | 1 | 2 | 3 | 1 by 8 | 3.8 |
| 396 | 1.1.0 |  | question | **Decide whether `fejkdata` writes Go struct types, and TypeScript types for the npm package, from a record's columns.** | 1 | 1 | 1 | 1 | 1 | 1 by 2 | 1.5 |
| 190 | 1.1.0 |  | question | **Decide whether fejkdata pseudonymises real datasets, and if so add the goal, the audience persona and the items below to the README and the plan.** | 4 | 2 | 3 | 7 | 9 | 1 by 4, 3 by 3 | 0.1 |
| 86 | 1.3.0 |  |  | **Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.** | 1 | 1 | 1 | 1 | 2 | code health by 2 | 5.2 |
| 154 | 1.3.0 |  |  | **Add `en_US.routing`, an ABA routing number over the Fed's prefix ranges with its check digit drawn by `{aba()}`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2.1 by 3, 3 by 3 | 3.8 |
| 243 | 1.3.0 |  |  | **Add `en_US.ein`, drawing only the prefixes the IRS assigns.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 2.1 by 3, 3 by 3 | 3.8 |
| 206 | 1.3.0 |  |  | **Add `zh_CN` with its given names by sex, last names and titles, and `zh_CN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 277 | 1.3.0 | 206 |  | **Add `zh_CN`'s national personal id, the 18-digit resident identity card number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 208 | 1.3.0 |  |  | **Add `es_MX` with its given names by sex, last names and titles, and `es_MX.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 282 | 1.3.0 | 208 |  | **Add `es_MX`'s national personal id, the CURP, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 212 | 1.3.0 |  |  | **Add `pt_BR` with its given names by sex, last names and titles, and `pt_BR.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 287 | 1.3.0 | 212 |  | **Add `pt_BR`'s national personal id, the CPF, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 207 | 1.3.0 |  |  | **Add `hi_IN` with its given names by sex, last names and titles, and `hi_IN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 306 | 1.3.0 | 207 |  | **Add `hi_IN`'s national personal id, the Aadhaar number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 209 | 1.3.0 |  | decision | **Add `ar_EG` with its given names by sex, last names and titles, and `ar_EG.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 310 | 1.3.0 | 209 |  | **Add `ar_EG`'s national personal id, the national ID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 211 | 1.3.0 |  |  | **Add `bn_BD` with its given names by sex, last names and titles, and `bn_BD.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 314 | 1.3.0 | 211 |  | **Add `bn_BD`'s national personal id, the NID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 213 | 1.3.0 |  |  | **Add `id_ID` with its given names by sex, last names and titles, and `id_ID.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 318 | 1.3.0 | 213 |  | **Add `id_ID`'s national personal id, the NIK, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 214 | 1.3.0 |  | decision | **Add `ur_PK` with its given names by sex, last names and titles, and `ur_PK.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 2.6 |
| 322 | 1.3.0 | 214 |  | **Add `ur_PK`'s national personal id, the CNIC number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 3 by 3 | 3.3 |
| 245 | 1.3.0 |  |  | **Draw `en_US.phone` from every area code NANPA has in service.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2, 3.2 by 3 | 3.2 |
| 250 | 1.3.0 |  |  | **Add `en_US.holiday`, the federal holidays.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 4, 3.2 by 2 | 2.9 |
| 244 | 1.3.0 |  |  | **Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.** | 2 | 2 | 2 | 4 | 4 | 1 by 2, 2.1 by 4, 3.2 by 3 | 2.5 |
| 198 | 1.3.0 |  |  | **Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.** | 2 | 2 | 2 | 4 | 3 | 1 by 1, 2.1 by 2, 3.3 by 3 | 2.4 |
| 276 | 1.3.0 | 206 |  | **Add `zh_CN.phone`, drawn from China's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 281 | 1.3.0 | 208 |  | **Add `es_MX.phone`, drawn from Mexico's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 286 | 1.3.0 | 212 |  | **Add `pt_BR.phone`, drawn from Brazil's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 305 | 1.3.0 | 207 |  | **Add `hi_IN.phone`, drawn from India's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 309 | 1.3.0 | 209 |  | **Add `ar_EG.phone`, drawn from Egypt's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 313 | 1.3.0 | 211 |  | **Add `bn_BD.phone`, drawn from Bangladesh's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 317 | 1.3.0 | 213 |  | **Add `id_ID.phone`, drawn from Indonesia's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 321 | 1.3.0 | 214 |  | **Add `ur_PK.phone`, drawn from Pakistan's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 2.2 |
| 248 | 1.3.0 |  |  | **Add `en_US.car`, a make and model from NHTSA vPIC.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 2, 3.2 by 2 | 2.0 |
| 246 | 1.3.0 |  |  | **Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2 | 1.7 |
| 278 | 1.3.0 | 206 |  | **Add `zh_CN.company`, `zh_CN.email` and `zh_CN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 283 | 1.3.0 | 208 |  | **Add `es_MX.company`, `es_MX.email` and `es_MX.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 288 | 1.3.0 | 212 |  | **Add `pt_BR.company`, `pt_BR.email` and `pt_BR.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 307 | 1.3.0 | 207 |  | **Add `hi_IN.company`, `hi_IN.email` and `hi_IN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 311 | 1.3.0 | 209 |  | **Add `ar_EG.company`, `ar_EG.email` and `ar_EG.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 315 | 1.3.0 | 211 |  | **Add `bn_BD.company`, `bn_BD.email` and `bn_BD.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 319 | 1.3.0 | 213 |  | **Add `id_ID.company`, `id_ID.email` and `id_ID.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 323 | 1.3.0 | 214 |  | **Add `ur_PK.company`, `ur_PK.email` and `ur_PK.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 1.7 |
| 247 | 1.3.0 |  |  | **Add `en_US.licenseplate`, drawn per state from that state's pattern.** | 2 | 2 | 2 | 4 | 4 | 1 by 1, 2.1 by 3, 3 by 2 | 1.6 |
| 275 | 1.3.0 | 206 |  | **Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 279 | 1.3.0 | 208 |  | **Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 284 | 1.3.0 | 212 |  | **Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 304 | 1.3.0 | 207 |  | **Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 308 | 1.3.0 | 209 |  | **Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 312 | 1.3.0 | 211 |  | **Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 316 | 1.3.0 | 213 |  | **Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 320 | 1.3.0 | 214 |  | **Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 1.2 |
| 271 | 1.3.0 |  |  | **Give `geo/FR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.1 |
| 280 | 1.3.0 | 279 |  | **Give `geo/MX`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.1 |
| 285 | 1.3.0 | 284 |  | **Give `geo/BR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 3.3 by 2 | 1.1 |
| 432 | 2.0.0 |  | question | **Reword goals 2.2, 7, 8.1, 8.5, 8.6, 9.2 and 10, and the README's Versioning table, for a Rust core that a native Go library and the npm package both serve.** | 1 | 2 | 1 | 2 | 2 | 1 by 1 | 0.6 |
| 433 | 2.0.0 |  |  | **Measure a seeded render of the README's person.json, one row and 100,000 rows, from Rust compiled to WebAssembly under wazero against the native Go library, and record whether wazero comes within a few percent.** | 2 | 1 | 1 | 2 | 2 | 1 by 1 | 0.6 |
| 434 | 2.0.0 | 432, 433 |  | **Add a Cargo workspace with a compose service, and a gate leg running rustfmt, clippy at the Go gate's complexity limit, the tests, and cargo-mutants for the stability pass.** | 3 | 1 | 1 | 3 | 4 | 1 by 1 | 0.3 |
| 435 | 2.0.0 | 434 |  | **Port internal/drawstate to Rust so a seed draws the same numbers as Go, and fail the gate where the Go and Rust CLIs print different bytes for a corpus of templates at a fixed seed and date.** | 3 | 1 | 2 | 4 | 4 | 9 by 5 | 10.4 |
| 436 | 2.0.0 | 434 |  | **Port internal/datafiles, internal/rows and internal/jsonvalue to Rust, loading on first reach a source with an index, and embed the shipped data per locale behind a Cargo feature.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.2 |
| 437 | 2.0.0 | 434 |  | **Port internal/grammar to Rust: templates, paths, selectors and where clauses.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.2 |
| 438 | 2.0.0 | 436, 437 |  | **Port internal/engine/tree and internal/engine/load to Rust: compile, bind, link and resolve.** | 4 | 2 | 3 | 7 | 9 | 1 by 3 | 0.0 |
| 440 | 2.0.0 | 435, 438 |  | **Port internal/engine/render and internal/avoid to Rust: drawing, named picks and distinct draws.** | 4 | 1 | 3 | 6 | 8 | 1 by 3 | 0.1 |
| 444 | 2.0.0 | 440 |  | **Fail the Rust gate where a render allocates more than its pinned count, as the Go performance gate asserts allocations.** | 2 | 1 | 1 | 2 | 3 | 12.1 by 5 | 15.6 |
| 441 | 2.0.0 | 435, 437 |  | **Port internal/builtinfunc to Rust: every builtin, Go's reference-time date layouts such as '060102', the checksums, calc with its Inf, and reword.** | 3 | 1 | 3 | 5 | 8 | 1 by 3 | 0.1 |
| 442 | 2.0.0 | 440, 441 |  | **Port records and their serializers to Rust: JSON, NDJSON, CSV and SQL, typed columns, linked templates through --emit, streamed.** | 3 | 1 | 2 | 4 | 6 | 1 by 2 | 0.2 |
| 443 | 2.0.0 | 442 |  | **Port the CLI to Rust, with its getopt_long flags, exit codes, --list and --data-path.** | 3 | 1 | 3 | 5 | 6 | 1 by 3 | 0.3 |
| 446 | 2.0.0 | 433, 442 |  | **Serve the Go library from the Rust core through cgo, under the module path github.com/larvit/fejkdata/v2, keeping New, its options, Fake, the templates, records and FakeStruct.** | 4 | 3 | 3 | 8 | 8 | 1 by 3 | 0.1 |
| 445 | 2.0.0 | 443 |  | **Build and publish the Rust CLI on every channel the Go one reaches: the binaries for six targets, deb, rpm and apk, Homebrew, Scoop and the image, with cargo-audit in place of govulncheck.** | 4 | 3 | 3 | 8 | 8 | 1.1 by 3 | 0.1 |
| 447 | 2.0.0 | 443, 445, 446 |  | **Delete the Go engine, keep the Go tests that check the Go library's API, and revise each decision in docs/decisions.md that names Go.** | 3 | 2 | 3 | 6 | 6 | code health by 3 | 1.3 |
| 448 | 2.0.0 | 442 |  | **Build @larvit/fejkdata's WebAssembly from the Rust core with wasm-bindgen, in place of the Go build item 359 makes.** | 3 | 2 | 2 | 5 | 5 | 1.2 by 6 | 1.1 |
| 450 | 2.0.0 | 447 |  | **Write MIGRATION.md, listing each change 2.0.0 makes to what a user relies on and what to do about it.** | 1 | 2 | 1 | 2 | 3 | 1 by 2 | 0.9 |
| 449 | 2.0.0 | 447, 448 |  | **Rewrite the README's Install, Library, Layout and Development sections, and AGENTS.md's rules, for a Rust core with a Go library.** | 2 | 2 | 1 | 3 | 4 | 11 by 2 | 0.4 |
| 439 | 2.0.0 | 438 |  | **Port internal/engine/fence to Rust: every proof and fence a load runs, with errors that name the same path, spelling and fix as Go's.** | 4 | 2 | 3 | 7 | 7 | 1 by 3 | 0.2 |

## Details

### 611. Lower the reading cost of the files a chunk reads most before its first commit.

The median of new input tokens a chunk reads before its first commit is rising: 102,969 over the last 5 chunks, against 72,954 over the 5 before them, as `reading-cost` measured on 2026-11-27 in item 359's chunk.

The files opened and searched most over the last 5 chunks: `README.md` 18, `.github/workflows/packages.yml` 5, `.github/workflows/test.yml` 5, `cli/records_test.go` 4, `internal/engine/load/reference.go` 4, `DATA-LICENSES.md` 3, `cli/cli.go` 3, `cli/flags.go` 3, `data-import/names-us.py` 3 and `fejkdata.go` 3. Make each chunk's way into them shorter, such as a README that opens on where each task starts, and measure the median again with `reading-cost --reset` in the chunk that ships this item.

Touches:

- `README.md`

### 612. Run the `npm-package-lowest` stage on Node 22.3.0, the lowest `engines` names, and keep Renovate from bumping it.

The Dockerfile's `npm-package-lowest` stage says it runs "the lowest Node that package.json's engines names", and `engines` names `>=22.3.0`. The stage's image is `node:26.13.1-slim`, which Renovate bumped in 901829cc, so no test runs the package on Node 22 and the `engines` range is unchecked. Found by item 359's stability review.

Pin the stage to a Node 22.3.0 image at its full patch version, and add a Renovate rule that leaves it there, as the project already holds other pins. A test asserts the stage's Node version, so the next bump fails the gate.

The human chose to keep this item in 0.9.0, though it delays the cut (asked and answered on 2026-11-28).

Touches:

- `Dockerfile`: `npm-package-lowest` (small)
- `renovate.json`

### 363. Document `@larvit/fejkdata` in the README: its install line, a JavaScript section beside Library, the runtimes it is tested on and its Versioning row, with every JavaScript example run as a test.

The README's opening says a Go library and a CLI; it names the package once it ships. The README already holds the package's install paragraph and a JavaScript quick start; this item adds what is left. Goal 11.2 asks every example to run as a test, so the JavaScript section's examples run in the gate, as `readme_test.go` runs the `json` blocks. The Versioning table gains a row for the package: removing or renaming an export, changing a declared type or raising the lowest supported runtime is a major, and an export is a minor.

The section lists the runtimes the gate tests the package on when this item ships: Node, Bun and Deno from item 549. Item 361 adds a browser, a bundler, Jest and Vitest, and extends the list when it ships.

Item 359 rewrites the package's calls in the same release, so this item reads the README as item 359 leaves it.

Touches:

- `README.md`: `## Install`, `### JavaScript and TypeScript tests`, `## Versioning`
- `npm-package/render.test.js`: lines 208-210 (small)

### 603. Add a mutation tool for JavaScript.

The `mutate` service runs Gremlins, and Gremlins mutates Go only. So no tool mutates `npm-package/index.js`, the package's code. Its tests are `npm-package/render.test.js` and `npm-package/load.deno.js`.

Use StrykerJS (`@stryker-mutator/core`), pinned to 10.1.1. That version needs Node 22 or later, so it runs on Node 26. Its command runner runs Node's built-in test runner, so the tests need no other framework.

The tests need the built `fejkdata.wasm` and the Go binary at `$FEJKDATA_CLI`. So the tool runs as a Dockerfile stage or a compose service beside the `npm-package` stage, and the README's Development section lists it. The package must still have no npm dependency. StrykerJS is installed in that stage at build time, never in `npm-package/package.json`'s `dependencies`.

Item 549's review mutated `index.js` by hand. Of 15 mutants, three lived: removing the global `__fejkdataRun`, retrying a failed load, and a program stopping while it loads. Tests now cover all three. A tool would find the next ones without a reviewer.

### 606. Compile and run each `go` block of the README in a test, so an example that fails to load fails the gate.

Filed from an outside reviewer's finding on item 393's pull request: the `FakeRun` example read `/sv_SE.person` and loaded no shipped module, so it returned a load error. `readme_test.go` runs every `json` block and the Python and npm tests run theirs, and nothing runs a `go` block. Wrap each block in a `main` with the imports it names, build it against the working tree, and run it over the data files the README tells the reader to save.

### 567. Ship a home for odd values, such as `misc.edge.name`, whose rows each say which assumption they break, with invalid values only under a path that says `invalid`.

From a user's feedback, citing `awesome-falsehoods-programmers-believe`: people assume too much about names and addresses. The engine carries any string, but the shipped data holds few odd ones. Default draws keep their real weights; a user mixes odd values in at a weight of their own.

A bench of the audience picks the spelling, and the column that names the broken assumption (working name `breaks`). Weigh a locale's own odd values, `sv_SE.edge`, against `misc`. Goal 3 has a fourth kind, invalid, so `internal/shipped` and `testdata/listing_shape.txt` need that kind word. Ship one small table with it, names, so the shape is proven. An odd value is never a real person's and reaches nothing real.

### 569. Ship odd names in Han, Arabic, Cyrillic, Greek, Hebrew, Devanagari and Thai, without waiting for their locales.

A handful of common real names per script, from an open source such as Wikidata, each with its script as a column. Each row carries a romanized column, as the `slugify` decision says, so an email can be built from it.

### 252. Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.

The AUR closed new account registration on 2026-10-14, so 0.1.0's packaging shipped without it. Once an account exists, restore what that chunk dropped: the `aurs` block in `release-tooling/goreleaser.yaml` (`git show 285ed85:release-tooling/goreleaser.yaml`), `aur` in both `--skip` lists, the `AUR_KEY` secret in both workflows, and the README's Install line and secret bullet.

### 568. Ship odd Latin-script names: one name, one letter, over 100 characters, particles, apostrophes, two surnames, display orders, and words such as `Null`.

Cover: a single name, one letter, `O'Brien` with both apostrophes, `St. John`, `de la Cruz`, hyphenated and spaced double surnames, Spanish two-part surnames, `Null`, `True`, `Test`, digits, inner capitals, equal first and last names, a suffix, ALL CAPS, lower case, `Last, First`, and family name first. Draw from W3C's "Personal names around the world" and the names gist, and cite them.

### 534. Reach the comprehension floor: lift Locality so the rule that a name keeps one pick has one home that render and the name fence both call, and render reaches each child only through the edges the fences walk.

The scoring runs at b0ebfb80 (PR #474), at 1909b699 (PR #489) and at 347eee72 (PR #550) left Locality at 5.63, 6.38, 6.00, 6.25, 5.63 and 5.88, Shape 6.63, 6.88, 6.88, 6.88, 6.25 and 6.63, Self-sufficiency 6.13, 6.63, 6.63, 6.63, 6.50 and 6.50 and Overall 6.25, 6.88, 6.63, 6.75, 6.25 and 6.50, under the floor of 7.0. Item 462 gave a pick its own frames, and the seats moved on to how the one-pick rule is split between `tree/pick.go`, `render/pick.go` and `fence/names`, joined only by prose:

- Locality: `internal/engine/tree/pick.go:86` `PlacedRead`: dotted string keys and `first[:atLen]` shaped to keep the render env off the heap, not by meaning
- Locality: `internal/engine/tree/pick.go:123` `LandsInPick`: one rule, a name keeps one pick, enforced across render, tree and fence/names that must agree
- Locality: `internal/engine/render/pick.go:157` `readPath`: frame swaps, framesFrom's memo and steppedDown pins; unclear that sibling fields reach one memo entry
- Locality: `internal/engine/render/render.go:32` `render`: the type switch must stay in step with tree.RenderEdges by convention only
- Locality: `internal/engine/tree/renderedge.go:34` `RenderEdges`: a render path added without its Rendered method leaves a hole in three fences
- Shape: `internal/engine/load/read.go:121` `compileNamedRead`: `Read.tail` is relative to the name while levels and steps span the path, and no type says which
- Self-sufficiency: `internal/engine/fence/names/pickreads.go:73` `keyOf`: agrees with grammar's JoinSegments only for the inputs it sees, which nothing states
- Self-sufficiency: `internal/engine/fence/names/refusedrawnapart.go:39` `sharedDraw`: early return justified by an invariant whose source wasn't found
- Self-sufficiency: `internal/engine/fence/names/refusefreshdraw.go:19` `refuseFreshDraw`: rendersField adds a fourth, negative condition the doc doesn't list
- Shape: `internal/engine/tree/loadfinds.go:12` `LoadFinds`: tree holds cross-phase working state, not only node types
- Shape: `internal/engine/fence/names/refusetwicedrawn.go:32` `readTwiceApart`: nearly the same boolean over route ends as readsFieldApart, and nothing says why they stay separate
- Self-sufficiency: `internal/engine/tree/read.go:11` `Read`: says levels run from the head, wrong for named reads whose levels start at ""
- Locality: `internal/engine/render/path.go:64` `drawAcrossTables`: a nil keptDraws means fresh, invisible at the call sites
- Shape: `internal/engine/render/pick.go:38` `namedPick`: render/pick.go would read truer as namedpick.go, with keptDraws beside it
- Locality: `internal/engine/fence/names/refusefreshdraw.go:104` `rendersField`: a negative exception that holds only if you know how render draws a fresh read beneath a pick
- Locality: `internal/engine/load/table.go:227` `routeThrough`: fromRow, drawsRow, past and stepUpRoute interact when a selector is followed by `..`
- Locality: `internal/engine/lazy.go:48` `callerPath`: loads as a side effect, and `batch.load` at :133 builds a fresh LoadFinds per load without saying why that is safe
- Locality: `internal/engine/load/path.go:132` `walkEvery`: a full-slice expression keeps each variant's append from overwriting
- Shape: `internal/grammar/where.go:59` `cutWord`: a hand-written clause scanner with a special case for `{x as where}`
- Shape: `internal/engine/tree/name.go:136` `NameScope.Kind`: the same rule as `opensScope` at `internal/engine/load/name.go:54`, in another package with no comment linking them, and its check order is load-bearing
- Shape: `internal/engine/fence/names/route.go:87` `drawsDownTo`: indexes the levels, though `route`'s doc says no rule does
- Self-sufficiency: `internal/engine/render/render.go:150` `readSibling`: clearing the pick is what makes every `{…}` draw afresh, and nothing nearby says so
- Locality: `internal/engine/tree/renderedge.go:9` file header: rendering a child outside the `Rendered*` methods skips the cycle check, the repeat cap and the name fence, and `render/render.go:32` doesn't repeat it (PR #550's run)
- Shape: `run.go:149` `readNeeds`: a stringly typed map of template names and `/run.x` paths, told apart by a leading `/` (PR #550's run)
- Shape: `run.go:83` `planRun`: a full run planner at the root, where the API files are otherwise thin; it stubs categories and slices a TSV header by its first newline at `run.go:325` (PR #550's second run)
- Shape: `internal/engine/tree/table.go:29` `RowsTable`: the pin set and the table each have two names, `rows.Pins` and `tree.PinSet` (PR #550's second run)

Give the pick-key model one home that holds its own invariants: the keys, where a read lands and which levels a binding addresses, in one place render and the name fence both call, with a key built by one function, so `keyOf` and `JoinSegments` can't disagree. Make render reach a node's children only through what `tree.RenderEdges` lists, or check the two agree in a test that fails when a new node type skips one, so the comment at `renderedge.go:9` is enforced. Let a type say whether a read's segments count from the name or from the path's head. Have each name rule's doc state every condition its code checks, and state the invariant `sharedDraw`'s early return rests on where it is established. Say why `readTwiceApart` and `readsFieldApart` differ, or merge them, and correct `Read.levels`'s doc. Give `NameScope.Kind` and `opensScope` one rule, say why a lazy load may use its own LoadFinds, and comment `readSibling`'s clearing of the pick. No render output or seed changes, and no allocation count rises; the performance gate holds.

Touches:

- `internal/engine/tree/pick.go`
- `internal/engine/tree/renderedge.go`: `RenderEdges`, `templateEdges`, `choiceEdges`, `cellEdges`
- `internal/engine/tree/read.go`
- `internal/engine/fence/names/refusetwicedrawn.go`: `readTwiceApart`
- `internal/engine/load/read.go`: `compileNamedRead`
- `internal/engine/load/resolve.go`: `setAddressed` (small)
- `internal/engine/render/render.go`: `render`
- `internal/engine/render/pick.go`: `readUnder`, `namedPick.readPath`, `namedPick.framesFrom`
- `internal/engine/render/path.go`: `drawFresh` (small)
- `internal/engine/fence/names/pickreads.go`: `keyOf`, `pickReads.keyAt`, `pickReads.addresses`
- `internal/engine/fence/names/route.go`: `route.placedAt`
- `internal/engine/fence/names/routewalk.go`: `routesUnder` (small), `walkRoutes` (small)
- `internal/engine/fence/names/refusefreshdraw.go`: `refuseFreshDraw`, `rendersField`
- `internal/engine/fence/names/refusedrawnapart.go`: `sharedDraw`
- `internal/engine/fence/names/refusekeptandfreshfield.go`: `drawnAfresh` (small), `freshFieldReads` (small), `readsFieldApart` (small)
- `internal/engine/tree/name.go`: `NameScope.Kind`
- `internal/engine/load/name.go`: `opensScope`
- `internal/engine/lazy.go`: `callerPath`, `batch.load` (small)
- `internal/engine/doc.go`: `lines 10-26` (small)

### 255. Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.

A table and its template, from each network's rules, which are facts.

### 117. List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.

Today the refusal names only the missing field, so a data author fixing a typo opens the category to find the spelling. Item 92 covers the hyphenated `calc` operand only.

### 604. Carry a parent's null, tab and line break whole to the children of a run, and accept a one-column parent that draws an empty or null value.

Filed from item 393's stability review. Today a run hands a parent's rows to its children as a TSV. So a child reads a null as an empty text, and a tab or line break as a space. A typed copy of a nullable parent column, such as an integer `{c.manager_id}`, fails at load where the parent drew a null. A child's copy of a multi-line value such as `sv_SE.address` differs from the parent's. A parent with one column that draws an empty or null value has no row a TSV can hold, so the run refuses it and asks for a second column. This item lifts each of these limits.

Chosen: carry the values whole, so a null stays a null and text keeps its tabs and line breaks. A nullable foreign key is common in the schemas bulk fixture authors load. Rejected: refusing, before the first row, a parent column that may draw a null where a child reads it. That would take linked templates away from those schemas.

Carrying the values whole means a run hands its rows over without a TSV. The chunk that ships this picks how: a table that holds a null, or a second kind of row source beside the table. Either way it revises the decision "A table is a record of string columns", to say what a run's rows are. It also drops the limits this item lifts from the decision on linked templates and from the README's Values bullet.

### 596. Publish an article on why Nordic test data needs national ids and addresses that match, and post fejkdata where Go and testing developers ask for tools.

The maintainer does this item, not a chunk: each step speaks for the project in public. The loop skips it.

- The article: the checks behind the Swedish personnummer, the Norwegian fødselsnummer and the Finnish henkilötunnus, the test series their issuers keep, and why a postal code must match its town.
- Posts: Go and testing communities and newsletters, and lists of database-seeding and fake-data tools.

Awesome Go is done from our side: the pull request there is approved and waits to be merged. Item 548's monthly table shows whether each step moved what LLMs answer.

### 168. Test in the merge gate the shared steps of `data-import/`, starting with `geo.fence`, the småort steps of `geo-se.py`, and the name-shape rules of `company-no.py` and `company_names.py`.

Against "Most, if not all, code should have automated tests" in `technical-principles.md`: no Python under `data-import/` has a test, and the gate, `docker build .`, runs Go alone. Today `geo.fence` is checked only by regenerating both countries' TSVs from the full downloads and diffing them. Each locale's places item adds a geo script for its country, each calling `fence`.

The tests also cover the rules of `company-no.py`: `shape_of` reads an unlisted word in an ENK name as the owner's last name, `roots` takes a word carried in at least two forms, and `register`, `without_wording`, `kind`, `count_shapes` and `endings` hold the rest. They cover `word_usage` and `trade_words` in `company_names.py`, which `company-se.py` and `company-no.py` share; its `Vocabulary.acronym_max` makes a word in capitals an acronym only up to that many letters or where it has no vowel.

They cover `scb_smaorter`, `inside` and `population_of` in `geo-se.py` too, where no gate test kills a mutant today, such as a swapped latitude and longitude or a dropped kommun filter.

### 457. Rename each script under `data-import/` to a name `import` can spell, so a test imports it like any module and the mutation tests can run on the folder.

A script under `data-import/` is named with a hyphen, as `geo-se.py`, and a module with an underscore, as `company_names.py`. A test reaches a script only through `geo_test.script`, which loads it by file path. `mutate-changes` (mutmut 3) cannot map such a file to its mutants: on the chunk that shipped mypy it stopped with "tests recorded trampoline hits but none match any mutant key", so it tested no mutant in any Python folder. The comprehension panel names the same split as a place: "hyphen against underscore tells script from module, stated nowhere", which `geo.py`'s docstring now states. Rename the scripts, update each command the README lists and the Dockerfile's stages, drop `geo_test.script`, and check that `mutate-changes` then runs over `data-import/` and `release-tooling/`. While renaming, move the six shared modules, `cldr.py`, `company_names.py`, `geo.py`, `source.py`, `tsv.py` and `xlsx.py`, with their tests, into one folder under `data-import/` named for what it holds, and leave the scripts at the top: the 0.6.0 release review found that a reader must open each of the 44 files to tell a module from a script. Related: item 168 adds the tests the mutants need.

### 257. Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.

Templates; the check-digit rules are facts.

### 389. Add `misc.rrule`, recurrence rules a test checks against RFC 5545's grammar.

From a real-world use case, a scheduler's fixtures, whose author called it niche, and it would have replaced most of one fixture's template. No register publishes how often each rule is used, so it draws evenly, as the decision "A table whose register publishes no frequency draws evenly" states, rather than the weighted mix the author asked for.

### 530. Draw `en_US.ssn` from the series the SSA never issues, area `000`, `666` or `900` to `999`, and say in the README which checker refuses it.

Found while shipping the project principle for ids with no test series (PR #469): `data/en_US/ssn.json` draws areas 001 to 665 and 667 to 899, so every number it prints may be a living person's. `docs/locales/en_US.md` says so, but it is the one national id left under neither national-id principle in `AGENTS.md`.

The SSA publishes that it assigns no area 000, 666 or 900 to 999, no group 00 and no serial 0000 (the randomization pages that page links). That is a series the issuer never issues, which the principle "A national id draws a series its issuer keeps for tests or never issues, where one is published" asks for, under goal 3.3. Many validators refuse such numbers, so the README says which one does; the draw may then fail a real check, a trade goal 3 weighs, so the first step is to check each of the SSA's published rules against the common checkers.

Recommendation: ship the never-issued series as the default path, and keep today's issued draw as a second path (as `fi_FI.henkilotunnus-issued` does for the Finnish id), so a user chooses. Area 900 to 999 overlaps `en_US.itin`, whose groups differ, so the two stay apart by group. Record the choice in `docs/decisions.md`, and add the category to the principle's "Applied by" list. Its draw changes what exists, so it follows the README's Versioning table.

### 406. Add a transform that writes a value escaped for a JSON string, so a text template spelling JSON stays valid whatever a drawn value holds.

A record's columns are flat, so a parent with an array of children renders as a text template whose format spells JSON, such as `{{"summary":"{s}"}}`. A drawn value holding `"`, `\` or a newline breaks that JSON, as a company name or a sentence may. The transform writes the value with JSON's string escapes, beside `lowercase`, `uppercase`, `ascii` and `slugify`. A writer panel picks its name.

### 570. Ship odd valid emails, and invalid ones under an `invalid` path.

Valid by RFC 5321 and RFC 6531: `+tag`, capitals, apostrophe, quoted local part, 64-character local part, 254 characters in all, an internationalised domain, an IP literal. Invalid: two `@`, a trailing dot, two dots in a row, a 65-character local part. Every domain is a reserved one, and an IP literal is from a documentation range.

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

### 602. Render a second template from the npm package in under 10 ms, today about 100 ms, by keeping the loaded data between renders.

The npm package's `render` runs the CLI once per call, so each call builds a generator anew: `fejkdata.New` over every shipped module. Item 549 measured about 0.1 seconds a call in Node 26, Bun and Deno, whatever the template. A test suite that renders a few values in each of several hundred tests pays that each time. The Python package pays a process start per call in the same way.

Keep what a load builds between calls, and give each call its own seed and day, so a render still keeps no state a caller can see: the same seed and day must print the same bytes whatever was rendered before. `npm-package/render.test.js` holds that test, and its comparison with the Go binary must still pass.

Touches:

- `npm-package/wasm/main.go`
- `cli/cli.go`: `Run` (small)
- `npm-package/render.test.js`
- `npm-package/README.md`: `## Size and speed` (small)

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

### 605. Run linked templates from the Python and npm packages, with a function that takes each template and its count.

Filed from item 393's product-owner review. The Python package's `render` always passes `--repeat` and a template on stdin, both of which `--emit` refuses, so a Python caller can't write linked templates; the npm package's `render` is shaped the same way. Add a function per package, such as `render_run({"customers": 100, "orders": 500}, format="sql", seed=1, today=...)`, that passes `--emit` and returns each template's output. Bench its spelling with the audience before it ships.

### 196. Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0", and ends the locale-address decision's conflict with goal 5.2. Carved out of item 182 on 2026-10-11, after six rounds of the comprehension panel ruled each attempt worse. Tried and ruled worse: a module list carried per category from each source's manifest (two routes, `walked` and `indexed`); the union of every manifest's list on the `Generator`, noted at four entry points; and the same noted inside `loadReached`, its list threaded through `loadInline` and `structCompile`. Each named every module read, even loaded ones, until the list held namespaces and the note named only those nothing in the tree provides, which the product-owner and goals reviewers accepted. A shipped locale's `Modules()` already carries its places, so this serves a bare `FS` beside a stand-in, a `--data-path` folder and anyone else's module; the error names the module and points to `Modules()`, as the 2026-10-11 bench asked. Bench the manifest key and the note's wording with the data author, since entries changed from import paths to namespaces after the 2026-10-11 bench.

### 263. Add `misc.color` names with their hex, rgb, hsl and cmyk forms.

A table and its templates, from CSS Color 4 (W3C) and the XKCD colour survey (CC0).

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

### 522. Rebuild `et_EE`'s first and last names from Statistics Estonia's CC BY-SA 4.0 counts by sex, replacing the language model's recall.

Ends the decision "`et_EE`'s names are a language model's recall".

Change `data-import/names-ee.py` to read the office's published counts by sex and year, as `names-de.py` reads Munich's, and keep the feminine form of a Russian surname under `f`. Name the licence, its share-alike terms and the attribution in `DATA-LICENSES.md`, and say in the README's License section what share-alike data asks of a user: the table needs its attribution, and a changed copy stays CC BY-SA 4.0. Remove `et_EE.first-name` and `et_EE.last-name` from the name-table principle's list in `AGENTS.md`.

### 572. Ship the Big List of Naughty Strings by kind, with Unicode edge cases, complex emoji and text that looks like another type.

Check the list's licence and add it to `DATA-LICENSES.md`. Unicode: decomposed accents, zero-width, right-to-left, astral, `İ` and `ı`, `ß`, lookalikes, odd spaces. Emoji: skin tones, flags, joined families, keycaps. Lookalikes: `null`, `true`, `=1+1`, `1/2`, `MARCH1`. A tab, line break or NUL needs an escape in a TSV cell, or a JSON category. Rows with brackets or quotes need item 166 where the cell is a key. A test renders every row through each format that can carry it and reads it back, and checks the refusal or the folding in the others.

### 259. Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.

Tables and templates, from OurAirports (public domain) and Wikidata (CC0).

### 610. Set `@larvit/fejkdata`'s publishing access to "Require two-factor authentication and disallow tokens", and check its live page, once 0.9.0 is published.

Only an owner of the npm organization `larvit` can set the access. Do it after the `npm` job has published 0.9.0, since npmjs.com takes the setting on a package that exists and the job publishes through the trusted publisher, not a token. Then open the package's page and check the README, the provenance badge and the files.

The trusted publisher (repository `larvit/fejkdata`, workflow `test.yml`, environment `npm`, `npm publish` allowed) lapses unless it publishes within two days of being set, and it was set on 2026-11-27. If it has lapsed when 0.9.0 is ready, set it again at most two days before the cut. npm 11.15.0 or later sets it with `npm trust github @larvit/fejkdata --file test.yml --repo larvit/fejkdata --env npm --allow-publish`. The agent that finds the release ready reminds the maintainer.

### 571. Ship odd dates, prices and numbers at their boundaries.

Dates: leap day, years 1 and 9999, ISO week 53, epoch 0 and -1, 2038, a leap second as text. Prices: 0, negative, huge, 3 and 4 decimals, accounting `(12.00)`, `1'234`, and `1 234` with a no-break space. Numbers: int32 and int64 limits, 2^53+1, `-0`, `1e400`, `NaN` as text, `007`, `1E5`.

### 267. Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.

Tables, from Project Gutenberg's catalog, MusicBrainz and USDA, each CC0.

### 264. Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.

Tables or choices, from GitHub Linguist (MIT) and curated lists. Item 32 adds `misc.browser`.

### 268. Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.

Tables or choices, from lorem ipsum and lists an LLM writes.

### 387. Let a template draw a number from a bounded normal distribution, so durations and start times cluster without weighted ranges.

From a real-world use case, a scheduler's fixtures: durations cluster around 2 hours and starts around 08:00. A weighted choice of `{int()}` ranges fills a typed column today, coarsely. Bounds are required, so a typed column's proof and `calc`'s still hold. A writer panel picks the spelling, between a third argument to `{int()}` and a function of its own, and what a draw outside the bounds does.

### 109. Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.

Technical principle: indent with hard tabs. All 19 scripts indent with four spaces, so an edit there either keeps the wrong style or rewrites the whole file.

### 91. Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.

A calc operand parses as a bare identifier, so `{calc(c.lat * 60)}` fails with `unexpected ".lat * 60"`, and a second binding of `/misc.coordinate.lat` would draw another coordinate. Resolve the path through `compileRead`'s leaves, as `operandNodes` does for a bare name. A data author who bound a row, `{.product as p}`, then bound its column, `{p.price as pr}`, got `no field "p"; a binding names a field, or a reference…` though `p` is a bound name.

### 23. Give every other shipped category that could reach something real a path that never does.

`phone` draws live PTS, NANP and Nkom ranges — PTS's five fiction series are the Swedish inert set, sourced in [research-sources-se.md](docs/research/research-sources-se.md), and the NANP and Norwegian ones still want a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.

### 158. Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.

A file a Windows editor saves with a BOM prints an invisible U+FEFF first, and a JSON template with one is read as a format string. PowerShell 5.1 encodes what it pipes to a program by `$OutputEncoding`, ASCII by default. Measured on a GitHub `windows-latest` runner (PowerShell 5.1.26100) while writing the README's PowerShell section: `$OutputEncoding` read US-ASCII, yet a pipe to a program still began with the bytes `EF BB BF`, so a JSON template exited 2 (`'{' inside a token in "\ufeff{…`) and a lone reference under `--format` exited 2 (`an inline record has no fields`). PowerShell 7.6.7 sent no BOM.

### 163. Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.

The README's Audience names a bulk fixture author writing millions of rows, and `-n 5000000` is misuse today, so they loop. The decision "The repeat cap bounds renders, not bytes" covers nested repeats in data, not `-n`. The README's Functions section justifies the builtin caps by a fat-fingered `hex(2000000000)`, a guess at a mistake under goal 4.7; the reason that holds is that one render may not allocate gigabytes. The goals audit of the narrowed goal 4.7 found it.

### 194. Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.

Goal 5.3 has an author prove their module, and goal 8.1 has fejkdata's data load as anyone's; today only fejkdata's own test, `TestShippedManifestsAreCurrent` over `indexEntry`, writes and checks an index. A category file the index leaves out never loads, silently.

### 496. Name each `geo.DE` Kreis in one style, as Destatis's Gemeindeverzeichnis names it.

GeoNames spells a Kreis sometimes bare ("Heidekreis") and sometimes with its kind ("Landkreis Oberspreewald-Lausitz", "Stadt Salzgitter"), and `data-import/geo-de.py` keeps that mix, with one override for Nürnberg. Take the names, with their kind, from the Gemeindeverzeichnis by Kreis code, and drop the override.

### 573. Map each falsehood about names, addresses, emails, dates and text to the path that tests it in `docs/edge-cases.md`, and show in the README how to mix odd values in.

The recipe works today: `"last": ["{/sv_SE.last-name}", {"format": "O'Brien", "weight": 0.05}]`. The page is the one agents and searchers land on, so it names each assumption in words. Each later item that ships an odd value adds its row.

### 269. Add `misc.direction`, `misc.continent` and `misc.ulid`.

Choices and a template.

### 361. Test `@larvit/fejkdata` in the gate on Node, Bun, Deno and a headless browser, and pin that its seeded renders match the Go library's byte for byte.

Goal 9.2 holds only while a test checks it: render a fixed set of seeded templates, records included, in Go and through the package on each runtime, and fail on any byte that differs. Each runtime runs in a container pinned to its full patch version, as a target of `docker build .` or a compose service the README's Development section lists.

Item 549 already does this on Node 26.13.1, Node 22.3.0, Bun and Deno, byte for byte against the Go build. What remains:

- A headless browser. The test loads the package as a page would, without a bundler.
- A bundler. A test builds a small app that imports the package with Vite or webpack, and checks that the WASM file is found and renders.
- Jest and Vitest. The package is an ES module only, so a test loads it from a Jest test and from a Vitest test, as a user's test suite would.

Item 549's review left the last two open.

### 161. Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.

`grammar/identifier.go` allows a field named ` x`, and `{ x}` reads it, but `checkBind` in `format.go` refuses every padded binding, so no spelling binds that field to a name (goal 3.1), and the refusal names one meaning only (goal 4.6). The goals audit of the narrowed goal 4.7 found it.

### 497. Draw every word of each locale's `word` tables from its dictionary, rude words included.

The human chose this when asked whether a locale's `word` should be a project principle: "Why without rude words? Skip that rule, just random instead and then make it a principle."

Today six import scripts, `word-se.py`, `word-us.py`, `word-no.py` (for both nb_NO and nn_NO), `word-dk.py`, `word-fi.py` and `word-de.py`, leave out each word on LDNOOBW's rude-word list for their language, through `wordtables.rude`. `word-us.py` also leaves out WordNet's usage domains of disparagement, ethnic slurs and obscenity. Remove these filters, `wordtables.rude` and each script's `--rude` flag, rerun the six scripts, and re-pin what moves. Delete `TestLocaleWordTablesDropTheRudeWordsListed` from `data/locale_word_test.go`. Replace the README's "drops the words on LDNOOBW's … list" sentences with one saying the tables hold rude words, so read the output before showing it to anyone. Drop LDNOOBW from `DATA-LICENSES.md`. Revise the three decisions that cite this item, and the principle's last sentence in `AGENTS.md`. Add the CHANGELOG entry.

### 29. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 544. Find out whether a very high repeat limit, nested repeats included, can render in bounded memory, and write down what it takes.

The maintainer, 2026-11-21: the CLI streams a top-level `--repeat`, yet it stops at 1,048,576, and a repeat inside a template builds its whole result in memory. Find out whether streaming can reach deep, nested structures, so the limit can be very high without memory trouble.

Questions to answer with a measurement, not a guess:

- Does a 100-million-row `--repeat` stream in flat memory today if the CLI's own limit is lifted, and does seeded output stay the same?
- Can a repeat inside a template, and nested ones, write to the output as they go, so the library and the CLI need no whole result string?
- What would the load check in `internal/engine/fence/graph.go` bound instead of a count, given goal 5.4's rule that errors surface at load?
- Once item 393 builds `--emit`: how does `{/run.<template>}` hold a parent's rows while its children render, and what bounds that? The other questions don't wait for it.

The answer is a decision in `docs/decisions.md`, replacing "The repeat cap bounds renders, not bytes" where it changes, and a plan item for the change. Goal 12 may need new wording, which is the maintainer's.

If the cap changes, change AGENTS.md's rule that calls it "a million rows", and the public texts that say so, in the same chunk.

### 36. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 495. Weigh `geo.DE`'s places and Länder by population, not by the postal codes they hold.

Today a locality weighs its postal-code count, which favours small places over the people living in them, as the README says. Destatis's Gemeindeverzeichnis publishes the population of each Gemeinde; match it to GeoNames' places by Kreis and name, and weigh a place by its people where the match holds. Decision: "A table whose register publishes no frequency draws evenly" does not apply, since a register does publish one.

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

### 407. Let a choice deal its items in exact counts across a run's renders, so 200 issues hold exactly 4 in progress, where weights only approach it.

A fixture often needs a fixed mix: of 200 issues, 4 in progress, 30 done and the rest to do. A weighted choice gives about 4, and the count shifts with the seed, so a test asserting on it needs a loader to fix the counts. A dealt choice is a deck, drawn without replacement. Decide what a render past the last card does: refuse at `New` where the repeat is known to exceed the deck, start a fresh deck, or fall back to the weights; and which renders share one deck: the run, as `{seq()}` shares one counter, or each `repeat`. The deck is per-generator state beside `seq()`'s counters, so the same seed and date still give the same output (goal 9). A writer panel picks the spelling.

### 218. Replace each country's invented streets with its real streets, from OpenStreetMap under ODbL or from the country's register.

Ends the decision "A country with no open address register covering it whole draws invented streets". The decision "The project takes open licences, share-alike ones included" accepts ODbL data, so only the build waits.

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

### 67. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens by restating the decision "GitHub is canonical, and the module path names it", a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 64. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 578. Draw number ranges, fractions, `bis`, `ter` and `quater`, and house names as odd house numbers.

`4-6`, `12 1/2`, `0`, and a house name in place of a number. Letters and French `bis` ship. Put the shares that a register supports into the default draw, and the rest in the odd-value tables.

### 576. Offer a gender value that can be nonbinary, undisclosed or unknown, beside the two-valued `sex` that names and ids read.

`sex` stays two-valued, since the national ids need it. Find a published share of each answer where a statistics office gives one; otherwise call the weights estimates. Today only the `en_US` title `Mx` leaves the two.

### 577. Draw apartments, floors, staircases, c/o and company lines, PO boxes and poste restante as optional address parts in each country.

Every address today is two lines with a street and a number. Cover `lgh 1102`, `2. th`, `Apt 4B`, `bt. 12`, `as 5` and each country's own form, after the postal operator's published addressing rules. Choose shares from a register where one counts dwellings per building. A PO box has its own postal code in several countries. Split the item per country when it starts.

### 580. Offer addresses of one, three, four and five lines, with and without a country line.

The lines come from the same draw as the two-line form.

### 583. Draw `en_US.phone.international`, and ship odd phone numbers: extensions, vanity letters, `00` prefixes, odd spacing, non-ASCII digits, and invalid ones.

Follow libphonenumber's falsehoods list and E.123. Emergency and short codes appear only as text to reject, and no number reaches a subscriber. Item 23 gives the ranges that reach nobody.

### 581. Ship odd addresses from countries outside the shipped ones.

No postcode, Irish Eircode, Japanese blocks, `BFPO`, `APO AP`, Queens `37-12`, Mannheim `Q7, 24`. Source each form from its postal operator or the UPU. The numbers are invented.

### 584. Draw a datetime in any zone with its true offset that day, and the local times that a clock change skips or repeats.

`misc.timezone.offset` is standard time only, and `misc.datetime` is always UTC. Compute from the IANA database embedded in the binary. `time/tzdata` adds about 450 kB, so weigh it against goal 2.2. Seeded output must not depend on the host's zone files.

### 62. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 582. Draw the longest, the shortest and the non-ASCII places, and the postal codes that many places share, by path.

The data is there: `Ii`, a 45-character French commune, FR 51300 over 46 localities. `where` clauses may already say some of it, so start from what a template can spell today.

### 574. Draw a middle name, an initial, a suffix and a title after the name in each locale's `person` where people there use them.

Suffixes such as `Jr` and `III` in `en_US`, and degrees after the name where the locale writes them. Weigh each share by a published figure where one exists, or call it an estimate. A new column in a record that exists is breaking, so list it or keep it out of the columns by default.

### 575. Offer a person's previous, preferred and legal name as values that agree with the person.

Nothing changes for someone who doesn't ask for them. The previous last name is drawn from the same locale's last names.

### 579. Ship localities with no street, with an address that names the locality and a number.

The project principle "A geo module that takes its localities from an address register leaves out a locality with no street" drops them in `geo.EE`, `geo.LT` and `geo.LV`. Revise it in `AGENTS.md`, and change the address shape so a place without a street is valid.

### 59. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 63. Pair a street with its exact postnummer.

Sweden needs an application to Lantmäteriet; today a Swedish street goes to the nearest postal code centroid. Norway needs none: Kartverket's address register, which `data-import/geo-no.py` reads, pairs each address with its postnummer, though `geo.NO` too draws a street and a postnummer apart, each inside one poststed. Finland needs none either: Posti's basic address file, which `data-import/geo-fi.py` reads, lists each street under its postinumero, though `geo.FI` draws the two apart, each inside one postitoimipaikka. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 65. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go. Write the wider check in `release-tooling/`'s tested Python, and move the workflows' other branching bash there with it: which paths count as a code change and the gate's condition in `test.yml`, the latest-tag test and the version checks in `packages.yml`. The 0.6.0 release review found that no test runs any of it, and a wrong pathspec skips the test job without a failure.

### 587. Decide whether an `--edge 5%` flag mixes odd values into any template, once the odd-value tables have been in use.

It has the widest reach for the least effort, but every shipped category would need a type that says which odd table fits it. The other way is a README recipe per category, which already works.

### 589. Load every odd value into PostgreSQL and MySQL from each format that can carry it in the merge gate, and compare what comes back.

The round trip today is through Python's CSV and JSON readers, run once by hand: 516 strings, no mismatch. Pin both images to a full patch version.

Compare the value only where the format can carry it. Test the rest on their own: `sql` and `pgcopy` refuse a NUL, or follow whatever item 566 ships, and `tsv` turns tabs and line breaks into spaces.

### 593. Ship national ids, IBANs and cards with a bad checksum or wrong length under an `invalid` path, and French NIRs of people born abroad.

Each invalid row says what is wrong with it. A NIR with department `99` is valid and belongs in `fr_FR.nir`, whose docs must say so.

### 585. Add `nfc` and `nfd` transforms.

`José` decomposed is the classic failed lookup. A dependency such as `golang.org/x/text` needs the owner's view on adding a module.

### 591. Ship skipped days and ambiguous date spellings.

Samoa's 2011-12-30, the 1582 Gregorian gap, `03/04/05`, two-digit years, 12 AM.

### 592. Ship odd company names, such as `! LTD`, script-like names, emoji, digits only and very long ones.

Take the forms from a register that really holds them, such as Companies House, and change the name where it would point at a living company.

### 590. Ship odd URLs, file names, version numbers and IP spellings.

URLs: IDN, punycode, port, IPv6 host, long query, double encoding, on reserved domains. File names: `CON`, trailing dot, 255 bytes, leading dash, two extensions. Versions: four parts, date-based, `1.10` against `1.9`. IPs: compressed IPv6, octal and integer IPv4.

### 595. Let a CLI flag make CSV cells that start a spreadsheet formula harmless.

Off by default, since it changes the value. Follow OWASP's CSV injection advice. A bench picks the flag's name.

### 588. Draw each row of a table exactly once, so a test run covers every odd value rather than some by chance.

A distinct draw of N covers part of a table. A bench picks the spelling. The same seed may give the same order.

### 594. Give `misc.territory` every currency and calling code a territory uses, and add Kosovo and Antarctica.

One currency and one calling code ship per row, so PA has only PAB. The 242 rows ask for a capital, a currency and a TLD, which Antarctica lacks in part. Kosovo has no ISO code but the user-assigned `XK`. Check the table before adding a row. A changed column is breaking.

### 586. Add transforms that make a value messy: stray spaces, doubled spaces, a case flip, a typo and broken encoding.

A bench of the audience picks the names. Every random choice comes from the seed, so output stays reproducible. Broken encoding is UTF-8 read as Windows-1252, such as `Ã¶`.

### 472. Decide how the shipped data is split into files and modules before it reaches the first size wall.

The walls, with today's distance to each:

- Module zip: Go limits it to 500 MiB (`golang.org/x/mod/zip`). Today's tracked files take 36.9 MB, about 13 times below.
- One file: GitHub blocks a file over 100 MiB and warns at 50 MiB (GitHub's docs). `geo/US/street.tsv` is 10.8 MB, about 9 times below.
- Pages: a GitHub Pages site is limited to 1 GB (GitHub's docs), and it holds the apt, dnf and apk repositories; about 10 times below, an estimate.
- CLI binary: 40.1 MB, 85% of it data, so about 380 MB at 10 times the data, an estimate.
- Build: a full rebuild takes 4.7 s, so the build is not a wall.

The options, for the maintainer to choose:

1. Split a table's file by its parent, and a flat table into fixed-size files. The study recommends this one.
2. Move the data into modules per region.
3. Build a CLI per region.
4. Deflate a block per parent. A spike shrank `geo/US/street.tsv` from 10.8 to 3.05 MB, but a parent's first reach took 40 to 65 µs against 13 to 24 µs, and a draw across every parent 250 to 320 ms per million against 146 to 156 ms. It needs a second loader, and the repository and the module zip carry the data twice. A binary form `go generate` derives from the TSV keeps goal 7.

### 466. Number a pick's memo slots at load and index child rows by parent row, so a render allocates no per-pick map.

- Profile of the README's eight-column person record: garbage collection marks about 19% of the time, `mallocgc` 16%, and string-keyed maps about 20%: the per-pick memo in `renderAt` and `variantOf`, `rows.drawUnder` and `parentRow`. Of a row's allocations, `strings.Builder` growth makes 31%, the memo maps 21% and pick frames 28%.
- A spike that turned the memo maps into small inline arrays and indexed child rows by parent row, with item 467's changes, took the record from 11.8 to 13.6 µs and 53.8 allocations a row to 8.8 to 10.3 µs and 28.8. `sv_SE.address` ran 22% faster, `sv_SE.person` 19% and a ten-column US record 14%, byte for byte the same on six templates of 200,000 rows each.
- A plain swap of a map for a slice gained only 3 to 9%: the slots must be numbered at load.
- Files: `internal/engine/render/pick.go` and `internal/rows/rows.go`.
- Readability: the comprehension seats found `drawMemo` and `namedPick.readPath` in `internal/engine/render/pick.go` hard to read, `drawMemo` for its one map per syntax feature; where this changes either, make it easier to read.
- Success: no output byte changes, and the performance gate asserts the allocations a row takes, as the decision "The performance gate asserts allocations, not wall-clock time" says. That assertion is goal 12.1's move.

### 471. Prove each shipped module in its own test package, so the test cache skips the data that didn't change.

- `go test ./data/...` takes 57 to 61 s and 1.0 GB without `-race` today, and would take about 10 minutes at 10 times the data, an estimate.
- Item 465's index moves the whole-table proofs to generate and test time (goal 5.3), so each module's own package must run them as a test.

### 467. Render a record into one reused buffer, and write `CSVLine` into one buffer, so a row costs one allocation for its text.

- `strings.Builder` growth makes 31% of a row's allocations, and `CSVLine` takes 5.5% of its time. Rendering a record into a reused buffer may save a further 15 to 25%, an estimate no spike measured.
- The bytes stay the same.
- Not blocked by 466: this changes the builders in `internal/engine/render/render.go` and the root's `record.go`, and 466 the memo in `internal/engine/render/pick.go` and `internal/rows`. Should the two meet in one function, land 466 first.

### 469. Render `--repeat` rows on every CPU, each row drawn from a seed derived from the run's seed and its row number, so the bytes don't depend on the number of CPUs.

- A spike on 16 CPUs with the README's record ran at 68 to 82 thousand rows a second on 1 worker, 137 thousand on 2, 296 to 303 thousand on 4, 572 to 648 thousand on 8 and 677 to 853 thousand on 16. With items 466 to 468 that may reach about 0.9 million rows a second on 8 workers, an estimate never measured with all of them together.
- It changes seeded output once, which the README's Versioning section allows in any release, a patch too. `{seq()}` and `--emit` keep one serial path, which seeds each row from the run's seed and its row number as the parallel path does, so adding `{seq()}` changes no other column's values.
- It needs a public bulk API in the library, lazy loading that is safe across goroutines, and a test that the bytes are equal at 1, 2 and 8 workers.
- It ships alone: its R is above 5.

### 473. Publish a bulk-throughput benchmark beside gofakeit, faker-js, Mimesis and Faker, with the command that reproduces each row.

- Measured on 16 CPUs with Go 1.27.4, for a comparable ten-column US CSV row: gofakeit v7.18.4 ran at 426 to 465 thousand rows a second (6 allocations a row), fejkdata at 79 to 85 thousand, faker-js 10.6.0 at 75 to 81 thousand, Mimesis 22.2.0 at 72 to 75 thousand and Faker 40.54.0 at 3.1 thousand. gofakeit's own `BENCHMARKS.md` gives 1,084 ns for an address and 3,466 ns for a person, on an Apple M1 Max.
- gofakeit's values don't hold together (a Texas city under the state WY, the street "South Avenuehaven"); fejkdata's do. Say so beside the numbers.
- Run it after items 466 to 469, and state what it measures then. Pin every tool to one CPU and show fejkdata's all-CPU row beside it, labelled, so the table never sets one tool's CPUs against another's.
- Prose in the README or `docs/` follows the project's prose rules.

### 468. Set the CLI's GC target to 400 unless `GOGC` is set.

- On the README's record, a million rows took 10.6 to 11.2 s where 12.6 to 13.5 s was, and `sv_SE.person` 7.7 to 9.3 s where 9.9 to 11.1 s was. Peak memory rose from 64 MB to 134 to 149 MB. Output bytes don't change.
- `GOGC=off` took 5.3 GB, and `GOMEMLIMIT` alone gained nothing.
- The library sets no GC target: that is its caller's.
- Where the README's CLI part states the memory the CLI uses, state the new figure.

### 465. Load a table's rows under one parent on that parent's first reach, from an index `go generate` writes beside the manifest, for every large table.

- Scope: every table larger than one block. A block is one parent's rows in a table with a parent column, one index line per parent; a flat table, with none, gets fixed-size blocks. This item sets the block size and the table size above which an index is written.
- Index: a spike's index for `geo.US`'s streets holds one line per parent — key, first byte, last byte, row count and total weight — 1,716 lines and 55 KB. Reading it took 0.3 ms and parsing one locality's streets 19 µs, against 24 to 33 ms to parse all 475,271 rows.
- Growth: a later spike took the first reach of one locality to 13 to 30 µs at 1, 5 and 10 times today's US street table, with the index at 67 KB at 1 times and 669 KB at 10 times. Today's loader takes 645 to 688 ms and 276 MB at 5 times, and 1.32 to 1.50 s and 552 MB at 10 times.
- Proof: the index states what the whole file's proofs prove (every parent key is a key of the parent table, every parent weighs above 0). `go generate` proves them and a test keeps the index current, as `TestShippedManifestsAreCurrent` does for the manifests. A stale index fails at load: it holds the table's byte length, checked when the table loads, and each block's parse checks its row count and parent key against it, so a hand-edited TSV beside an old index is a load error naming the table, never another parent's rows. Any module may carry an index, so a shipped module still loads as anyone else's does (goal 8.1); a table without one loads whole.
- Risk: `Len`, `Cell(row, col)` and `Draw` in `internal/rows` assume the whole table. A draw over the whole street table with no locality (`{/geo.US.street}` alone) would draw a parent by its total weight, then a row under it. Storing each parent's starting cumulative weight in the index, and summing a block's weights in file order, may keep seeded output; that is a design claim no spike checked. Check the seeded pins, and re-pin a changed one as its own commit.
- Readability: the comprehension seats found the worklist in `loadReached` (`internal/engine/lazy.go`) hard to read, and it grows through `batch.load`; where this changes either, make it easier to read.
- Success: the CLI's first `en_US.address` under 20 ms and its peak under 40 MB, from 110 ms, as re-measured after the one-pass parent read, and 80 MB, and no first reach whose cost grows with the shipped data.
- It ships alone: its R is above 5.

### 470. Hold a parsed block of a table as its text and its row offsets, not one string per cell.

- Today a street row keeps about 110 bytes: a `[]string` of 16-byte cells, plus lookups. Text and row offsets would keep about 27 bytes, an estimate no spike measured.
- After item 465, a run that reaches every parent still keeps 40 MB at today's US streets, and 275 to 407 MB at 10 times them.

### 327. Add `is_IS`'s national personal id, the kennitala, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 326. Add `is_IS.phone`, drawn from Iceland's numbering plan.

Find the numbering plan and its licence first.

### 324. Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.

Read them from the register the streets item names, Staðfangaskrá, where it holds them; else find an open source first.

### 325. Give `geo/IS`'s places their streets.

Staðfangaskrá, Iceland's address register, holds every street with its póstnúmer.

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

### 86. Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.

Both phone formats carry the same 21-code `area` list, so an edit to one copy can miss the other.

### 243. Add `en_US.ein`, drawing only the prefixes the IRS assigns.

The IRS's prefix list is facts. [Source research](docs/research/research-sources-world.md).

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

### 245. Draw `en_US.phone` from every area code NANPA has in service.

Today it draws 21 area codes; the exchange already ships as `en_US.phone-exchange`. No research file covers NANPA yet: find its list of area codes in service and that list's licence first.

### 250. Add `en_US.holiday`, the federal holidays.

[Source research](docs/research/research-sources-world.md). The federal holidays are listed in 5 U.S.C. 6103, which no research file covers yet.

### 244. Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.

NAICS 2022 is public domain; SEC's `company_tickers.json` is free to reuse. [Source research](docs/research/research-sources-world.md).

### 198. Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.

`streets()` in `geo-us.py` fetches only each drawn place's main county, the one holding most of its people. A ZCTA in a place's other counties then has no address range and is dropped: New York, under Kings County, ships Brooklyn's codes, one Manhattan code and none on Staten Island, and its streets are mostly Brooklyn's. Whether a neighbouring county is fetched depends on which other places the draw picks. The fix reads every county in `localities`' `county_part`.

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

### 248. Add `en_US.car`, a make and model from NHTSA vPIC.

NHTSA vPIC is public domain. The decision "`misc` is what every locale shares" puts a national register's cars in its locale. [Source research](docs/research/part2-internet-vehicle-airline.md).

### 246. Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.

No national register defines the account number, and no research file covers Nacha yet. Item 154 adds the routing number beside it.

### 247. Add `en_US.licenseplate`, drawn per state from that state's pattern.

No open list of state patterns exists: write each state's pattern by hand and check it against that state's DMV. Wikipedia's list (CC BY-SA) and openalpr's (AGPL) may be read, never copied. [Source research](docs/research/research-sources-world.md).

### 275. Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.

`rows.Pins` holds eight pins inline, enough for a five-deep geo tree. If CN's tree runs deeper, raise that number to keep a render off the heap. Find an open source of its places and postal codes, and that source's licence, first. It draws invented streets (decision: no open address register covering it whole).

### 279. Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.

SEPOMEX publishes every código postal with its asentamiento.

### 284. Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.

IBGE's CNEFE lists every address with its CEP.

### 304. Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.

India's PIN code directory is open data on data.gov.in. It draws invented streets (decision: no open address register covering it whole).

### 308. Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It draws invented streets (decision: no open address register covering it whole).

### 312. Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It draws invented streets (decision: no open address register covering it whole).

### 316. Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It draws invented streets (decision: no open address register covering it whole).

### 320. Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It draws invented streets (decision: no open address register covering it whole).

### 271. Give `geo/FR`'s places their streets.

The Base Adresse Nationale (Licence Ouverte 2.0) holds every street with its code postal.

### 280. Give `geo/MX`'s places their streets.

INEGI's Marco Geoestadístico names the streets.

### 285. Give `geo/BR`'s places their streets.

IBGE's CNEFE lists every address with its CEP.

### 432. Reword goals 2.2, 7, 8.1, 8.5, 8.6, 9.2 and 10, and the README's Versioning table, for a Rust core that a native Go library and the npm package both serve.

The owner decided on 2026-11-11 that fejkdata moves to Rust as a whole and keeps Go as a native library. Today those goals name Go as the one language: Go packages as modules, a CLI of your own in a few lines of Go, functions in Go, the JavaScript package matching the Go library, and the lowest supported Go in the Versioning table. The wording is the owner's; every other item of 2.0.0 waits for it.

While rewording, weigh two points the architecture pass of 0.5.0's release review raised. Item 446 serves the Go library through cgo, so `go install` fails without a C toolchain, which works against goals 8.6 and 10 as they read today. Item 359, the Go library built to WebAssembly, would serve goal 1.2 with one engine, without the Rust core.

### 433. Measure a seeded render of the README's person.json, one row and 100,000 rows, from Rust compiled to WebAssembly under wazero against the native Go library, and record whether wazero comes within a few percent.

The owner accepts a Go library that runs the Rust core as WebAssembly only where it costs no more than a few percent against native Go. Expect 1.5 to 3 times slower before the cost of copying strings across the boundary. Where it does come within a few percent, item 446 binds through wazero instead of cgo, and Go users keep `CGO_ENABLED=0`.

### 434. Add a Cargo workspace with a compose service, and a gate leg running rustfmt, clippy at the Go gate's complexity limit, the tests, and cargo-mutants for the stability pass.

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
