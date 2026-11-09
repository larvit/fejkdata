# Changelog

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). A release's
`Breaking` section comes first and names each rejected spelling with its
replacement, and each removed path, column or flag.

## Unreleased

### Breaking

- `let` is an option, so a field can no longer be named `let`; give it another name.
- A `null` or a `datatype` in the fields of a template inside a column, as in
  `"ride": {"format":"{n}","n":null}`, now counts: `null` was text and `datatype` did nothing there.
  A bare number or boolean there stays text.
- A column that reads one field through a name, as `"len":"{p.len}"` does, now has that field's
  datatype and null, so an integer `len` is written `3` where it was written `"3"`. Write
  the column as `{"format":"{p.len}","datatype":"string"}` to keep the text, with a null written `""`. A choice whose
  items hold different datatypes of that field fails to load; give them one datatype.
- `en_US.word` and `en_US.sentence` draw from WordNet 3.0, about 70,000 words where 271 were hand-written, so what a
  seed draws changes and a word may be long or rare. `en_US.word`'s format is now `{any}`, a pick between the four
  parts of speech, and both categories read the new tables `en_US.word-noun`, `word-verb`, `word-adj` and `word-adv`.
  A program embedding the `en_US` module carries WordNet's licence, `WORDNET-LICENSE.txt`, as the release archives and
  packages do.
- `sv_SE.person` gains the columns `personnummer` and `samordningsnummer`, which spell the person's `birthdate` and
  sex: `{/sv_SE.person as p}{p.birthdate} {p.personnummer}` prints e.g. `1932-04-16 320416-2394`. They sort by name
  among the other columns, so `--format csv` and `--format sql` write them: add them to the table, or read by header.
  `birthdate` draws `sv_SE.personnummer`'s four age bands, counted from `WithToday` or `--today`, where it drew 1930 to
  2025 evenly, so what a seed draws changes, and changes again with the date. The format binds the date, `{birth as b}`,
  and the new paths are `sv_SE.person.birth`, `birth.date`, `birth.sign`, `personnummer`, `samordningsnummer` and
  `samordningsnummer.birthday`, the birth date's day of the month.
- A field, column, category, folder or bound name can no longer contain ` where ` or ` distinct `, or end in
  ` where` or ` distinct`, since ` where ` now opens a filter and ` distinct ` a draw of distinct values; rename it.
- `geo.SE.locality` places four postorter in the kommun of the SCB småort holding their centre, where their postal
  codes named another: Lönneberga in 0860 Hultsfred, not 0884 Vimmerby; Axmar in 2180 Gävle, not 2182 Söderhamn; Rökå
  in 2418 Malå, not 2481 Lycksele; and Torsö in 1083 Sölvesborg, not 1493 Mariestad. Each weighs that småort's people,
  57 to 129, where it weighed 200, so `sv_SE.address` draws them less often, and what a seed draws changes.

### Added

- A `where` clause draws only among the values that pass it, keeping their weights:
  `{/en_US.word where length 5 and not containing 'x'}`. The conditions are `length 5`, `length 4 to 6`,
  `not 'cat', 'dog'` and `not containing 'cat'`. A length may read a name, `{n as k}{/en_US.word where length k}`,
  and `as` after the clause binds the filtered draw. A path whose values are not fixed text, or a clause no
  value can pass, fails to load.
- A `where` clause also filters `{digits(n)}`, `{lower(n)}`, `{upper(n)}`, `{hex(n)}`, `{base64(n)}`, `{nanoid(n)}`,
  `{uuid()}` and `{ulid()}`, drawing a value that holds none of its strings: `{hex(4) where not containing '7958'}`.
  `not` and `not containing` may read a name, as in `{code as c}{hex(4) where not containing c}`. Where the strings a
  name holds leave no value, `else 'none'` names what prints; without `else`, a path prints an empty string and a
  call a value of its own shape, such as the Nil UUID for `{uuid()}`.
- `distinct N` draws N different values in one pick, each by its weight among those not drawn yet, and each value
  of a nested choice counts on its own. `{day distinct 2 joined ','}` prints them joined, as `TU,MO`, and
  `{/en_US.word distinct 3 as w}` binds them, read as `{w[1]}` to `{w[3]}`, also inside `calc`, `adddays` and
  `where length`. It comes after a `where` clause and before `as`. Asking for more values than the path can draw
  fails to load.
- `{reword(text, /en_US.word)}` replaces each word of `text` with a drawn word of its length, or the nearest
  length the words hold, in the case of the word it replaces. Digits and punctuation stay, and a word reads as one
  fake throughout a render, in every call reading the same words.
- `{adddays(d,n)}` prints the date `d` plus `n` days, keeping its time of day: `{start as s}{len as n}{s},{adddays(s,n)}`
  prints an end `n` days after the start it shows. A third argument, `{adddays(d,n,'layout')}`, prints it in another layout.
- The option `let` holds fields that write no column: `{"format":"{a}","let":{"a":"{/x}"},"b":"{a}"}`
  has the one column `b`. A record holding every field in `let` fails with `ErrNoColumns`.
- `misc.creditcard-test` and `misc.iban-test` draw only numbers published for testing:
  Stripe's test card for each of seven brands, and Stripe's test IBAN for each of BE, DE, ES,
  FI, NO and SE, with Danske Bank's fictive example for DK. Both are tables, so `misc.creditcard-test[visa]`
  and `misc.iban-test[SE]` select one row.
- `en_US.word.noun`, `.verb`, `.adj` and `.adv` draw one part of speech, and `en_US.word.verb.present` draws a verb and
  prints its third person, such as `carries`. `en_US.word.any` prints what `en_US.word` prints. They read the tables
  `en_US.word-noun`, `word-verb`, `word-adj` and `word-adv`.
- `en_US.paragraph`, three to six sentences of `en_US.sentence`.
- `da_DK.date`, a date from 1970 to 2029 such as `2.4.2019`, with no zero before a one-digit day or month.
- `da_DK.email`, a lower-case ASCII address at `example.com`, `example.org` or `example.net`, or a
  subdomain of one. Half of its local parts are a first and a last name drawn from `da_DK.first-name`
  and `da_DK.last-name`.
- `da_DK.company`, an invented name and a legal form, ENK (a sole proprietorship, whose name has no
  abbreviation), ApS, A/S, I/S or K/S, such as `Nordholm Holding ApS`. Each form is drawn as often as
  Statistics Denmark counted enterprises of it in 2024, except that I/S and K/S, which it counts together,
  are drawn evenly. `da_DK.company.enk`, `.aps`, `.as`, `.is` and `.ks` each draw one form, and
  `{/da_DK.company as c}{c.legal}` prints the abbreviation of the same pick. The name reads
  `da_DK.company-coined`, `da_DK.company-word`, `da_DK`'s names and `geo.DK.locality`. Statistics Denmark's
  counts are CC BY 4.0, so a program showing them credits "Kilde: Danmarks Statistik".
- The README's [Load into a database](https://github.com/larvit/fejkdata#load-into-a-database)
  section pipes records into PostgreSQL, SQLite, MySQL and MongoDB.
- `da_DK`, Danish, with `da_DK.person` and the tables it reads: `da_DK.first-name`, `da_DK.last-name`,
  `da_DK.sex` (`kvinde` and `mand`) and `da_DK.title` (`dr.` and `professor`). A first name that girls
  and boys both bear has a row under each sex, and a person's first name and sex agree. The top 20
  girls' names, boys' names and last names carry Statistics Denmark's counts of 1 January 2026, and
  `sex` carries its population on that day. Every other name is a language model's recall of a name
  common in Denmark, with an estimated count, as no good online source of Danish name counts was
  found. Statistics Denmark's tables are free to reuse with a source credit, so a program showing
  them credits "Kilde: Danmarks Statistik". The Go package is
  `github.com/larvit/fejkdata/data/da_DK`, and the CLI carries it.
- `geo.DK`, Denmark's places: its regions and kommuner from Statistics Denmark, and every postdistrikt
  holding a street address in Danmarks Adresseregister, each with its postnumre and its streets. A
  kommune's `code` is four digits as the address register writes it, `0101`, and a region's is
  `1084`, as DAGI, the register of Denmark's divisions, writes it. `da_DK.address` prints `geo.DK.address`, a street, a house number, a postnummer and a
  postdistrikt, such as `Vestergade 12\n8000 Aarhus C`. A program showing the addresses credits
  Klimadatastyrelsen. The Go package is `github.com/larvit/fejkdata/data/geo/DK`, which
  `data/da_DK`'s `Modules()` also returns.
- `--format tsv`, and `Record.TSVHeader` and `Record.TSVLine` in the library, write records as the
  TSV file of a table, so a later run draws from an earlier run's records. A brace in a value is
  written doubled, a tab or a line break as a space, and a null column as an empty cell.
- `da_DK.cpr`, a Danish CPR number such as `070761-4285`: a birth date of someone under 100, drawn
  by Statistics Denmark's shares of under 18, 18 to 64 and 65 and over, a seventh digit that the
  CPR office's table of century digits allows for the year, and a last digit that is even for a
  woman and odd for a man, such that the ten digits pass the office's modulus 11 check. The office
  publishes no test series and no unassigned range, so a number may be a living person's. Its
  `birth` field is the date, and `da_DK.individ` holds a digit of each `da_DK.sex`'s parity.
- The builtin `{cpr(x)}` completes a CPR number. It reads the last six digits emitted, the date,
  and adds a seventh digit that the year of that date allows, an eighth and a ninth it picks, and a
  check digit with the parity of the digit `x` reads. The year is the latest one not after the date
  of `WithToday` or `--today` that bears the year digits. It adds nothing where `x` is not one digit
  or that year lies outside 1858 to 2057, the years the office's table writes.
  `{/da_DK.cpr.birth}-{cpr(s.individ.digit)}` agrees with a bound sex `s`.
- `da_DK.phone`, a Danish phone number such as `41 23 45 67`, mobile ten times as often as fixed,
  grouped in pairs as the Danish Language Council (Dansk Sprognævn) writes it. Its `national` and
  `international` (`+45 41 23 45 67`) fields are the two spellings of one number. It reads
  `da_DK.phone-mobile` and `da_DK.phone-fixed`, tables of the three-digit prefixes that Annex 1 of
  Executive Order No. 1883 of 7 December 2020 on the Overall Danish Numbering Plan allocates to
  mainly mobile or mainly fixed numbers, each keyed by its prefix as written, `phone-mobile[34 2]`.
  A number's `block` is that prefix and its `subscriber` the last five digits. Neither the order nor
  the authority's guide to it lists the numbers assigned to providers, so a number may be a real
  subscriber's.
- `fejkdata.Builtins()` returns each way a template calls a builtin function, sorted, with its
  arguments named, such as `hex(n)`, `iban(CC)`, `seq()` and `seq(name)`. `--help` lists them and
  explains names: `{/path as p}` draws once and prints nothing, and every `{p}` and `{p.field}`
  reads that one pick. It also links the README's Names, Functions, Data format and Your own data
  sections. The links point at the README of the fejkdata release the binary was built with, such
  as `.../blob/v0.4.0/README.md#names`. A build from anything else, such as a development build, a
  pseudo-version or a replaced module, links the latest README.
- The README's [Ratios and rules](https://github.com/larvit/fejkdata#ratios-and-rules) section shows
  80% `active` as a weighted choice and a currency that follows its country as a named pick. The
  README now also says a drawn `misc.creditcard` number or `{iban(CC)}` may be a real card's or
  account's.

### Changed

- `geo.US`'s Utqiaġvik is written in NFC, as every other shipped value is, so its `ġ` is one code
  point and the name is nine code points long.
- A token reading a field beside the field that holds it, as `{b}` does in
  `{"format":"{a}","a":"x{b}","b":"y"}`, still fails to load, and the error now states the rule and
  names the fix.

## [0.3.0] - 2026-11-05

### Breaking

- `New` requires `WithToday(day)`, the date to treat as today, beside `WithSeed`: add
  `fejkdata.WithToday(time.Now())` for the current date, or a fixed day to repeat a run's output on a
  later day. Without it `New` fails and names the option.
- `sv_SE.personnummer` draws four age bands as often as SCB counted them in
  Sweden at the end of 2024, up to age 113, and counts ages from `WithToday` or `--today`, where it drew
  evenly from 1930 to 2025. What a seed draws changes, and changes again with the date: give `WithToday` or
  `--today` a fixed day to repeat a run on a later day. `sv_SE.personnummer` spells `+` from the year the
  holder turns 100, so a system reading it must accept the `+`. It is now a choice of four age bands, where
  a format string stood.
- `misc.territory.country` is now `misc.territory.sovereign`. It still holds the alpha-2
  code of the state a territory belongs to: `{/misc.territory[GL].sovereign}` is `DK`.
  The column sorts by name, so `--format csv` writes it after `numeric`, where `country`
  came after `capital`: read by header, not by position.
- `sv_SE.person` gains a column, `birthdate`, a date from 1930-01-01 to 2025-12-31, drawn apart
  from any `personnummer`. It sorts first, so `--format csv` writes it before `first` and
  `--format sql` inserts it: add the column to the table, or read by header.
- `sv_SE.sex`, `en_US.sex` and `nb_NO.sex` draw a woman or a man as often as each sex occurs in the
  population, where they drew them evenly. Each gains a `count` column, the people of that sex: in Sweden
  at the end of 2024, in Norway on 1 January 2026 and in the United States on 1 July 2024. What a seed
  draws changes, and `--format csv` and `--format sql` write the new column for a lone `{/sv_SE.sex}`:
  add it to the table, or read by header.
- `geo.SE.locality` weights 24 postorter by the population of the SCB småort 2023 each lies in,
  50 to 199 people, where the table weighted each 200. `sv_SE.address` draws those postorter less
  often, and what a seed draws changes.
- `sv_SE.company` and `nb_NO.company` gain a field and record column, `legal`, the legal-form word of
  the company: `{/sv_SE.company as c}{c} ({c.legal})` reads both from one pick. `legal` is empty for an
  enskild näringsidkare and an ENK. `--format csv` and `--format sql` write the column for a lone
  `{/sv_SE.company}` or `{/nb_NO.company}`: add it to the table, or read by header. An item of
  `sv_SE.company.registered` is now one wording of one form, weighted by the form's registrations and
  the wording's share, so what a seed draws changes. `{c.registered}` beside `{c}` or `{c.legal}` no
  longer loads: read `{c}` for the bound company, or write `{/sv_SE.company.registered}` for a second one.
- A load refuses `{n.w}` beside a read of `n` that reads `w` through a name of `n`'s category, such as
  the format `{w as r}{r}`, since the two could print different draws. A pair that shares no draw, such
  as a constant `w`, still loads. The error says how to write each meaning.
- A load refuses `{n.path}` beside a read of `n` that renders a reference to `path`, or to a path above
  it, also where `path` runs into a table linked to the one `n` binds: a reference draws afresh, so the
  two could print different draws. `{/sv_SE.company as c}{c} {c.ab.name}` no longer loads: drop
  `{c.ab.name}` to keep one pick, or write `{/sv_SE.company.ab.name}` for a second draw. Nor does
  `{/region as r}{r} {r.municipality.name}` where the table `region`'s format is
  `{/region.municipality.name}`. Nor does `{/region as r}{r} {r.municipality}`, where `municipality`'s format is `{name}`: `path` lands
  on a level above the reference's target, and that level's format reads the target. A reference to a
  constant, with no choice on the way, still loads, and so does a `path` whose level's format doesn't
  read the target.
- A load refuses a read such as `{r.legal}`, where `r` is bound to a field holding a choice and the
  `legal` of an item reads a name that item binds. Such a template loaded, then failed when drawn. Bind
  the name outside the choice, or read `{r}` whole.
- A load refuses two reads of `n` where one reads a field from the pick and the other draws that field
  afresh: `{n}` beside `{n.w}`, where `n`'s format is `{w.a}` and `w`'s is `{a}`, and `a` isn't a
  constant. The two could print different values. Read `{n.w.a}` in place of `{n.w}` to keep one draw,
  or write `{/cat.w}` for a second draw, where `n` binds `/cat`.
- `geo.SE`: a road segment inside a tätort leans towards the postort the tätort is named after, so Stockholm
  holds 1580 street names where it held 363, and a postort beside it, such as Solna, holds fewer. Five postorter
  join `locality`: Hjorted, Läckeby, Löderup, Mjöhult and Sövde. Seeded Swedish addresses change.
- `sv_SE.word` and `sv_SE.sentence` draw from SALDO, Språkbanken's Swedish dictionary, about 108,000 words where 233 were
  hand-written, so what a seed draws changes and a word may be long or rare.
- `--format sql` writes up to 1000 rows per INSERT, one statement per line, where it wrote one INSERT
  per row. A script counting lines or splitting statements on newlines sees fewer of them. A statement fails
  whole, so one row a constraint refuses, such as a repeated value in a `UNIQUE` column, loses up to 1000
  rows where it lost one: drop the constraint while loading, or load through a category of your own whose values are unique.

### Added

- `--today DAY` gives the CLI the date to treat as today, as `YYYY-MM-DD`; it defaults to the
  system's date.
- `{born(min,max,'layout')}` draws a birthdate by age, counted from `WithToday` or `--today`.
- `{luhn(n)}` computes the check digit over the last `n` digits, as a twelve-digit personnummer needs.
- A load accepts `{n}` beside `{n.w.z}` where `n`'s format is `{w.a} {w.b}` and `w` is a template that
  isn't a choice, and `{n}` beside `{n.w}` where `n`'s format is `{w}-{w}` and `w` is a constant. It
  refused both, though the two reads share no draw that can differ.
- `sv_SE.word.noun`, `.verb`, `.adj` and `.adv` draw one part of speech; `sv_SE.word.noun.definite`,
  `sv_SE.word.verb.present` and `sv_SE.word.adj.definite` each draw a word and print one form of it. `sv_SE.word.any` prints what
  `sv_SE.word` prints. They read the tables `sv_SE.word-noun`, `word-verb`, `word-adj` and `word-adv`.
- `Record.SQLInsertHeader(table)` and `Record.SQLValues()`, the two parts a program joins into a
  multi-row INSERT. `SQLInsert(table)` still writes one row's statement.
- `sv_SE.licenseplate`, a Swedish registration number, `ABC 123` or `ABC 12A` at even odds.
  It is valid, not real: a drawn plate may be on a real car. Its three letters never form one
  of the 99 combinations that Transportstyrelsen's page listed as blocked in March 2026, and
  never include I, Q or V. The last character of `ABC 12A` is never O. Its fields are `old`
  and `current`, one form each, and the parts they join: `old-number`, `001` to `999`,
  `current-number`, `00` to `99`, and `last`, the final letter. `sv_SE.licenseplate-letters`
  holds the three-letter combinations, in its field `letters`.
- `sv_SE.bankgiro`, a bankgiro number of seven or eight digits, `991-2346` or `5555-5551`, and
  `sv_SE.plusgiro`, a plusgiro number of two to eight digits, `12 34 56-6`. The last digit of each
  is a Luhn check digit. Wikipedia and convention give the plusgiro length and spelling.
- `sv_SE.bankaccount`, a clearing number and an account number that passes the check its bank
  demands, `1522-8923497`, from Bankinfrastruktur i Sverige AB's list of clearing numbers and
  Bankgirot's "Bankernas kontonummer" of 2024-02-22. `{/sv_SE.bankaccount[5000]}` selects a
  clearing number, and a row's fields are `clearing`, `bank` and `number`. The 15 clearing numbers
  `9565` to `9569` and `9600` to `9609` are not shipped, since that document doesn't publish their
  check. The drawn numbers are valid, not real: one may belong to a real account.
- `sv_SE.personnummer-12` and `sv_SE.samordningsnummer-12`, the twelve-digit form of each, century
  first and no hyphen, `199001012385`. The Luhn check digit covers the last ten digits, which
  `sv_SE.samordningsnummer-12`'s field `ten` holds.
- `sv_SE.sni`, SCB's SNI 2025 industry codes, the 835 five-digit ones, selected by a code such as
  `62.100` and drawn as the code and its Swedish label.
- `sv_SE.ssyk`, SCB's SSYK 2012 occupation codes, the 429 four-digit ones, selected by a code such as
  `2512` and drawn as the code and its Swedish label.
- `sv_SE.car`, a make and a model such as `Volvo XC60`, weighted by Mobility Sweden's registrations of
  new passenger cars in 2025. It holds the 119 models with 500 or more registrations.
- `sv_SE.month`, CLDR's Swedish month names, selected by 1 to 12: `{/sv_SE.month[1]}` is `januari`.
- `sv_SE.weekday`, CLDR's Swedish weekday names, selected by 1 (Monday) to 7: `{/sv_SE.weekday[7]}` is `söndag`.
- `sv_SE.holiday`, the thirteen public holidays of lag (1989:253), selected by a code such as
  `nationaldagen`. It holds no dates and leaves out the law's entry for every Sunday.
- `sv_SE.territory`, the Swedish name of each `misc.territory` row, selected by its alpha-2 code.
- `sv_SE.bloodtype`, a blood type at Sweden's shares, `A+` 37 percent down to `AB-` 1 percent, from
  sv.wikipedia's Blodgrupp article.
- `{bgmod11(x)}`, a builtin that prints a random digit and the check digit that complete the digits
  emitted so far, after the literal digits `x`, to a number Bankgirot's 11-modul accepts.

### Changed

- Installing with Homebrew needs a newer Homebrew than before. If Homebrew can't read the cask,
  run `brew update` and install again.

### Fixed

- A record naming a choice, such as `{/sv_SE.bankgiro}` under `--format csv`, gives an error matching
  `ErrNoColumns`, and the CLI adds the same hint on quoting as for a category with no fields.
- A choice item whose format reads only a name the item binds, such as `{/ab as w}{w.legal}`, draws
  without a panic. It loaded, then panicked when drawn.

## [0.2.1] - 2026-11-04

### Changed

- The binaries, packages and image are built with GoReleaser v2.20.0.

## [0.2.0] - 2026-10-28

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

- A patch of the latest release follows an update to Go, a Go minor included, to a module
  fejkdata requires, or to GoReleaser, once the tests pass with it. Until 1.0.0, a Go minor
  may stop supporting an older system, such as a macOS version; the changelog section of
  the release that first ships the Go minor names each one under Breaking.

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
