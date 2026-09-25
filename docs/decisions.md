# Decisions

## GitHub is canonical, and the module path names it

Goal 1 wants the usage the tool earns and goal 2 expects extenders who did not write it;
both need a stranger to file an issue and open a pull request. Gitea has no anonymous
issue, and no cross-host pull request at all, so a contributor would need an account and
a fork on a personal instance. Valid while the project wants contribution from outside:
the workflow decided nothing, being portable already — `github.api_url` and
`secrets.GITHUB_TOKEN` resolve on either host, so only the names moved.

## The vocabulary sits below `doc.go`'s package clause, not in the package doc

Goal 1 is what a developer choosing a library reads first, and the package doc is that
page: the glossary would fill two thirds of it with the render path's units, mostly
unexported, and Go doc comments have no code markup, so every name would render its
backticks literally. Below the clause it reaches goal 2's reader — who opens the file —
and nothing else. Valid while the glossary's subject is the render path; one written to
teach the exported API belongs above the clause.

## Options and fields share one namespace

`format`, `weight`, `repeat`, `separator`, `datatype` and `drawGroup` are reserved;
every other key is a field. Nesting fields under a key, or prefixing options, would tax
every template to guard against a misspelt option.

## `{a|b}` stays beside nested choices

`[[…], […]]` picks the same way, but its arms are anonymous; `{female|male}` keeps
`person.female` addressable.

## Flags follow getopt_long

`--name value` and `--name=value` both work; a short flag's value attaches or follows
(`-s42`, `-s 42`) and short flags bundle (`-hn 3`), as every shell user expects. A
single-dash long flag is rejected naming the double-dash spelling, and `-s=42` is
rejected naming both short spellings: `=` belongs to the long form, and reading `=42` as
the value would make `-d=./x` a directory named `=./x`.

## An argument is a template by its shape, not by a flag

A JSON object, array or string, or a string carrying a `{` token, is an inline template;
anything else is a path. A name may not contain a brace, a bracket or a quote, so a path
can never collide with any of those spellings, and the leading `[` or `"` is gated on
valid JSON so a stray copied bracket never swallows an argument — it names nothing, and
says so. No `--template` flag is needed. Reserving the characters whole — though only a
leading one could collide — keeps one simple name rule instead of a leading-position
special case. The JSON string is what makes the library's own advice reachable: the
error for an object holding only a format names `"…"`, and that spelling has to work
where it is printed. An argument or struct tag of one reference alone, `{/users}`, is
refused naming the path `users`: both render the same text, and only the path names a
record. A folder-relative `{.name}` or `{..name}` is refused naming `{/name}`, since an
inline template sits in no folder. `IsTemplate` exports the rule, so the CLI, struct
tags and any other caller read one.

## An inline template skips the cycle fence

`New` proves the loaded tree acyclic, an inline node is a finite tree of its own, and
nothing in the tree can reference it, so no render of it reaches itself. Every other
fence runs over both, from one `checkScope`, except that struct tags leave column
agreement to their Go types.

## An inline template that does not compile is misuse (exit 2), including a reference that resolves to nothing

The whole argument is the spelling under test, and `NewTemplate` compiles, links and
validates as one step. An unknown *path* stays a runtime error (exit 1): there the
argument is well-formed and only the data is absent.

## A padded JSON argument is rejected, not trimmed

Padding is the one place the two readings disagree — a format string renders it, JSON
drops it — so the spelling that renders is named rather than silently chosen.

## `FakeTemplate` and `NewTemplate` both stay

They reach the same value but not at the same cost: `NewTemplate` pays the compile and
validation once and renders many times, `FakeTemplate` is the one-shot call, and
`--repeat` is exactly the case that needs the first. The pair is `regexp.MustCompile`
and `regexp.Match`, not two spellings of one result.

## The shipped data is embedded, not discovered

A directory a machine happens to have would make `--seed 42` machine-dependent. Data
still lives in `data/` as JSON; `--data-path` layers over it.

## A bare reference draws each time; a reference path is held

`{/p} {/p}` is two draws, as `{word} {word}` is, while every `{/p.first}` in one render
reads one draw, and a bare `{/p}` beside them is a load error: a bare token is by
contract an independent draw, a path pins its level, and a fresh draw of a pinned level
could show another row. A builtin's operand holds what it reads for its expansion,
references included, so `{uppercase(/p)} {/p}` is one draw — the rule every operand
follows.

## Reference sigils follow the filesystem

`/` is the root, `.` this file's folder, `..` the folder above — what those spellings
already mean to anyone who has typed a path. A locale's files reach each other without
naming the locale, so a folder renames and copies without editing its references.

## A change to what exists is a major; a minor only adds

Data files, the CLI and the Go API are the public API, and a consumer must be able to
take a minor without an edit — so an added column is a major, since it changes the CSV
header and the `INSERT` column list, as is a removed value, which changes what a fixture
holds, and a new option, which reserves a field name. One spelling per result grows by
tightening, so every fence invalidates some file. Each such release names the rejected
spelling and its replacement in the changelog and in the load error, and that is the
whole migration: a fence rejects one spelling with one replacement, so the fix is local
to each site. A fence that would need a non-local rewrite ships a converter with its
release instead. Before `v1.0.0` a minor carries what a major would.

## Seeded output is promised within one version

Any edit to a category shifts its stream and everything drawn after it, so a promise
across versions would freeze every shipped list; a fixture is re-pinned on a bump, as
this repo's own are.

## An error is a contract by what it names, not its bytes

A script branches on the exit code and reads the named path or spelling, so those hold;
wording improves in a minor.

## A format holding a reference is fenced at link, and its errors name its path

2026-09-24, larv-review; approved 2026-09-25 by lilleman. Goal 2: a format compiles once, and one holding a reference
can compile only once bound, so its fences fire at link and name the category path as
every link error does, a cell's with its line; one without a reference still names its
file at compile. Valid while compilation needs the bound references.

## Raising the lowest supported Go is a major

A consumer building on it breaks, which is the one test every rule above applies; Go's
convention of a minor is not followed.

## The format check runs on the latest Go only

`gofmt`'s output is the toolchain's, not the code's: 1.27 stopped padding a map
literal's values out to a lone long key, so no source satisfies both it and 1.22's. A
consumer on the lowest supported Go depends on the code compiling and its tests passing
there, which is the `portable` stage, and never runs `gofmt` over this source, so the
latest toolchain alone defines the one canonical form goal 2 asks a reader to meet.
Valid while the lowest supported Go is not the latest.

## The changelog heading is the one spelling of a release; CI cuts the tag

A tag pushed by hand is served by `go get` at once, so a tag whose commit lacks its
heading is burnt, not fixed. The heading on a gate-passed `main` commit is the trigger
instead: the tag can land only there, and the GitHub release the same job publishes
keeps one text as its body and is where prebuilt binaries will attach.

## A `--data-path` override rebinds every reference to the category it replaces

References bind against the merged tree, so once shipped data uses `{.person}`, a
consumer's `sv_SE/person.json` is what every shipped reference into `person` reads, and
`New` fails on shipped data the consumer never wrote when that file lacks a field those
references read. Accepted: overriding is the point of layering, the error names the
reference and the field, and the fix is the consumer's file carrying the fields the
shipped tree reads.

## The repeat cap bounds renders, not bytes

A repeat, alone or nested, may ask for at most 1 048 576 renders; how large each render
is stays what the data asked for, so `{hex(1048576)}` repeated to the cap is a terabyte,
loaded without complaint. A byte estimate would need every builtin to declare a width to
fence a shape no data comes near, and the harm lands on the author who wrote it.

## 64-bit targets only

The gate builds amd64, and the buffer sizing a render pre-computes (renders × bytes)
assumes a 64-bit int; on a 32-bit target it could overflow and panic.

## A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`

`1/0` and a fixed `"0"` field are decidable, so they join the never-numeric operand as a
load error; the fold stops where an operand varies, so `a/(b*c)` with `b` fixed at `0`
and `c` varying loads and prints `Inf` every draw — catching it needs zero-absorbing
algebra for a shape nobody writes. A [typed column](../README.md#datatype) bounds its operands
instead and refuses a divisor it cannot keep from zero.

## In data, a default written out and a constant spelled as a sample are load errors

`weight: 1`, `repeat: 1`, `separator: ""`, `datatype: "string"`, `int(5,5)`,
`float(1,1,2)`, `+5` and `05` each spell what a shorter form already spells, so each is
rejected naming that form. The CLI's numbers follow the shell instead: `--seed 007` and
`--repeat +3` are 7 and 3, as every command line reads them.

## Samples say what they emit, transforms what they do

`{upper(2)}` is two letters, `{uppercase(x)}` is `x` upper-cased; one name for both
would turn on whether the argument looks like a number.

## A record is a template seen as columns; a Go struct is the one second schema

A template's `format` composes its fields into one string; `FakeRecord` and `--format`
project the same fields as columns. Two views of one dataset, so a record author writes
the same JSON they already know, and a column is the same field `Fake` renders by dotted
path. The `format` is inert to a record — a record-only template writes `"format": ""` —
but it is compiled and fenced, so a template that loads renders as whichever shape is
asked for. `FakeStruct` takes its columns from a struct instead, because a Go caller has
already written that schema: the fields name the columns and their types are the
datatypes, so a tag says only what to draw, and a `datatype` in it would be a second
spelling of the type. The Go type is a struct column's one datatype, so its items need
not agree on one among themselves; each must only hold that type.

## A struct's records follow Go's field access, and compile on first use

The fields an embedded struct promotes are the struct's own — `e.First`, as
`encoding/json` and SQL mappers read them — so they are columns of its record and share
its draws; a tagged field that another field hides is refused, not dropped. A named
struct field is another entity and a record of its own. `fake:"-"` leaves a struct
field, embedded or named, unfilled, so no name may be `-`; a pointer back to a struct
already being filled is left alone, since filling it would never end. `New` cannot see a
caller's types, so the first `FakeStruct` for a type compiles its tags and the answer,
error included, is kept per type: a test's first call is its load, and no `NewStruct`
handle is needed, as the cache already compiles once.

## A render shares one reference draw per category, per group

Every reference path into a category in one `Fake`, or one record, reads one draw of it,
so a value's facts agree across its fields, nested templates and columns alike —
`{/currency.code}` in one field and `{/currency.symbol}` in another name one currency,
whichever view renders them. A `repeat` iteration is a render of its own, since
repeating asks for another entity, and a [draw group](../README.md#draw-group) names further
entities within one render, so a payer and a payee are two groups over one `person`
rather than two copies of it. Only references share: a sibling field is local to its own
expansion, so a `first` column does not silently bind to a `first` in the column next to
it.

## A draw group name is local to its category

A category's groups are its own entities, so a caller naming a group the same way never
joins them by accident, and renaming a group inside one file changes no render
elsewhere. The unnamed group still spans categories, since facts that belong together
across categories must agree.

## The expansion hold and the render's draws are two fences

One proves a sibling path or an operand is reached only by its readers within an
expansion, the other does the same for reference paths across a render and its draw
groups. They pin different things — an operand pins the value its own render produced
and stops at a reference, a path pins every level it passes through — so one walk would
carry both rules and both scopes anyway, and tell them apart at every step.

## A record makes its draw maps up front, a `Fake` on its first read

A record's columns always read through the render's draws, so making the maps where the
set is declared keeps them on that frame's stack. A `Fake` often reads no reference path
at all, and making them anyway cost about a fifth of the cheapest render, so it makes
them on the first read instead, at two heap allocations for a render that does share a
draw. The allocation gate over a repeat of a reference path and over a named draw group
prices that, and pins the two measures that keep a hold set off the heap.

## A category never references itself, and a record's fences run at load

A category is one unit: a reference back into it — `{/users.first}` inside `users` —
describes a draw other than the fields beside it, where goal 3 asks facts that belong
together to come from one draw, so `New` refuses it and the sibling path stays the one
spelling for a field of one's own. A value two fields share goes in its own category,
which both reference. That settled, a record's column fences run at `New` too, so a
category that loads renders as whichever shape is asked for, and a reference reaching
back into a category through another one is refused there as the overlap it is. Which
reads those fences weigh differs on purpose: a record-only template's inert format
renders nothing, so a `drawGroup` on it can never matter and is refused, while one on a
rendering format can matter to a caller that bare-references it.

## A record's column set is fixed before the first draw

Only a category-level template is a record: a path descending into a field, or naming a
folder or a choice, errors. A tail may pass through a choice whose variants carry
different fields, so the columns — and with them the CSV header written once ahead of
every row — would vary per draw. A fixed column set is what the CSV and `INSERT`
contracts rest on, so the restriction holds even where a particular choice would happen
to agree.

## Null is a `null` item, not a rate

A null is one more outcome of a column's draw, so a choice's weights skew it like any
other; a null-rate option would be a second way to state odds.

## A typed column holds one value, not composed text

Its bounds come from a literal or a call's arguments, so a load error names a real
value, a range check is one comparison, and `1{digits(2)}` is a second spelling of
`{int(100,199)}`.

## A column of one reference alone is the column it reads

`{/src.score}` renders exactly what `src.score` draws, so it takes that column's
datatype and null rather than restating them, and a `datatype` restating the one it
takes is a second spelling. Any other `datatype` still types the values — the one way to
type a column someone else wrote.

## A typed column's calc is refused unless proven

Operand bounds must keep each divisor from zero and the result finite; what they cannot
show is refused rather than trusted, since a bare `NaN` breaks the JSON and SQL it lands
in.

## `Column` carries text, not a Go value

`Value` is the rendered string beside `DataType` and `Null`, which each serializer
writes as the load check proved it; a `Value any` would hand every caller a type switch.

## `hold` names what a draw is kept in, `draw` the draw itself

Goal 2 wants a name to reach one unit, so a unit takes the stem of what it is, and a
file the stem of the units it holds.

## A function may not spell a method; two types may

A call writes a method with its receiver and a function bare, so one name on both greps
as one unit, which is the reaching cost goal 2 counts, while the receiver before a
shared method name says which type answers. `TestNoFunctionSpellsAMethod` holds it. A
type is outside the rule: `renderScope.hold` answers with a `*hold`, so the two spell
one unit, which no test can judge.

## The package stays flat

Go ties a package to one directory, so folders would split the API into packages.

## The performance gate asserts allocations, not wall-clock time

`AllocsPerRun` is deterministic across machines, so a ±10% ceiling does not flake under
CI load, while time varies with the machine and its neighbours. A rendering slowdown
almost always costs an allocation too (a lost pre-size, a per-item map, an extra copy).
The benchmark suite (the README's Development) reports time for a human, not as a pass/fail gate.

## Rows live in a TSV, the shape in JSON

`New` allocates once per node, so a register of thirty thousand rows written as JSON
objects would cost it a second; a TSV is one allocation whose cells are substrings, and
the JSON says only how a row is composed. The TSV sits beside its category file, named
by `rows`, so a data directory stays a directory of categories, and one nothing names is
a load error rather than a file silently ignored.

## A selector is bracketed, and a dot inside it is literal

`municipality[0180]` reads as selection to anyone who has indexed an array, and `[St.
Louis]` keeps a name whole where a colon or a dot-separated spelling could not; zsh
needs the brackets quoted, which the README's examples show. A name that spells another
row's key is refused at load rather than shadowed: a key names one row by contract, so
the name could never select its own, and the check is one lookup per row against the key
index that already exists.

## `parent` names the link column and the table alike

One word says both, so a child table sits in its parent's folder and links on a column
of the parent's name; the geo plan wants exactly that, and a table needing another
folder or another column name would be asking for a second spelling.

## After a row, a path names a column or a linked table

`locality[Lund].address`, a template beside the family under a selected row, is refused
today; admitting it later is additive, since a refused spelling gains a meaning and no
accepted one changes, so the door stays open for the address records the plan describes.

## A table read into is pinned; a table read whole draws afresh

A path into a table pins its row for the render and group, as a reference path pins its
level, and a bare `{/city}` draws a row each time and pins no row of its family, as a
bare reference does; so a bare table beside a path into its family is refused like a
bare reference beside a path into it. A bare table reference still counts as a read for
a `drawGroup`, since the group is what draws it apart from the family's pins.

## Every reference path into one family selects the same rows, per render and group

A selector pins rows for the render, and a read that draws freely before it could pin a
row the selector contradicts, so accepting both would make the result depend on which
token rendered first. Requiring one selection per family per group is checkable at load
with no data lookup beyond the selectors themselves, and the error names the rewrite.
Two selectors naming one row by key and by name compare equal, since the load check
resolves them.

## A parent row with no child row is a load error

A descendant is drawn inside the nearest pinned ancestor, so every ancestor row must
lead to a row at every level below it, or a render could find nothing to draw. The
import script drops or fills such rows; the alternative, falling back to a free draw,
would break the consistency the link exists for without saying so.

## The choice-of-rows fence guards a data file's root, and requires string fields

A table is a category with a TSV beside its file, so only a root choice has the spelling
the fence names; a nested choice of same-shaped templates and an inline one keep
loading. Fields must all be strings because a cell is a string node: a choice whose
items carry a nested choice is not one table but two linked ones.

## A table is a record of string columns

Its columns are the CSV header and the `INSERT` column list, fixed by the TSV header, so
a table is a record by construction; every column is a string until a typed column
option earns its place.

## The key index is built at load, the rest on first draw

A link is proved against the parent's keys and a key's uniqueness is a data mistake, so
both are load-time; the name index and the per-parent child lists serve only a draw or a
selection, so they wait for the first one, keeping `New` linear in the bytes read.

## Two categories may name one TSV

Each is a view of the file with its own format and options, at the cost of holding the
rows twice, which is what a category over a register with two natural formats asks for;
a TSV nothing names stays a load error, since that one is a file forgotten rather than
shared.

## The rows of a table are alternatives

Only one row renders per read, so within one read a cell in one row and a cell in
another never meet, and each may select its own row of another table; the cells of one
row, and whatever they reach, do meet, and so does the format beside them. The rule the
fence applies is that two reads meet unless they sit in rows that never render together:
rows falling under different rows of a table a path pinned, two rows of that table
itself included. A choice's items get no such treatment yet: two items selecting
different rows are still refused.

## A table never reaches its own family, by any route

A `repeat` iteration and a `drawGroup` each draw apart on purpose, but a row that lists
three localities from other regions is the output the family exists to prevent, so the
own-family fence walks through both rather than stopping where the draw fences do.

## Tables carrying token cells stay small

The family fence compares the reads of every pair of rows that can render together, so a
table whose every row's cell selects a row of another table loads in time quadratic in
its rows: about a second at four thousand rows. No shipped table carries such cells, and
a register is a column set rather than a set of references, so the fence is left as it
is until a real data set needs the indexed form.

## A path is walked once without drawing before it is walked for real

A path that fails below its first level then moves no seeded stream, at the cost of one
draw-free walk per call, which allocates nothing.

## A country's postal codes and streets are siblings under its locality

No open source pairs a Swedish street with its postnummer, and pairing the US through
its ZIPs would shape the two trees differently, so both draw inside the pinned locality
and an address agrees at that level. A street's own code is the exact pairing to add
when a source carries it.

## A locale's `address` reads its country's `geo` tree, so the shipped set loads whole

`data/sv_SE` alone no longer loads: a test loads `data` and prefixes the locale, and
`--no-shipped-data -d` takes the whole `data` folder or a set of one's own.

## The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more

Sweden fits whole in 700 KB; every US place of 10,000 would pass a megabyte and fetch
1,200 counties of TIGER files, so the threshold sits where the two countries match in
size, and `--min-population` and `--streets-per-locality` on the import scripts build a
fuller set. The two trees add about 20 ms to `New`, which loads the shipped set in about
45 ms.

## A locale's `address` restates its country record's format

A record cannot read another whole and keep its columns, so `sv_SE.address` names the
same four columns as `geo.SE.address`, each a reference into it, and the format appears
twice; a column is spelled the same in both, `street-number`, so the two never disagree
on a name.

## A postort's kommun comes from its name, its tätort or its codes, never from distance

GeoNames leaves a fifth of Sweden's codes without a kommun and carries stale spellings;
the nearest code across a border named the wrong kommun half the time it was tried, so a
postort none of the three rules place is dropped, as is one not cased like a place name.

## A highway designation is not a street, and a US postal code belongs to the place holding most of its land inside places

`I- 55 Bus` and `US Hwy 1` carry the most address ranges in many places and would head
every address, so the import drops names spelled as a route. A ZCTA goes to the place
its largest in-place part lies in, census-designated places left out since they never
ship, and ships only when that place does, so a few dozen places whose every code lies
mostly in a bigger neighbour ship no address; counting the land outside every place too
would drop a quarter of the places, whose codes straddle unincorporated land, for a
postal city the USPS mostly names the same way.

## `--list` stays a plain list of paths

It is what a script reads, so every line has to be a path that `Fake` takes; a marker
for the tables a `[selector]` follows, or a legend above them, would make the output
something to parse before use. `--help` names the selector spelling instead, and the
Table section teaches it.

## A layout is always quoted

A layout may carry the comma that separates arguments, `'January 2, 2006'`, and one
spelling for every layout beats a rule about which ones need the quotes, so the bare
spelling is refused naming the quoted one. The layout is Go's reference time because the
library renders with it and a Go caller already knows it; its names are English, and a
locale's own month and weekday names are data.

## A title is a table under `sex`

A prefix drawn apart would put `Mr` on a record whose `sex` column says `female`, which
is the disagreement the record exists to prevent; the tables this set already has are
what a title needs, so `en_US.title` links to `sex` as `first-name` does. Swedish has no
everyday sexed honorific, so `sv_SE.title` is a table as well but carries no `parent`.
Its weight column is `share`, not the `count` a name table carries, because the values
are a curated proportion rather than bearers anyone counted.

## A table owns the spelling of a selector on it

A reference reaches a table by a path that carries no selector — `sv_SE.person.first`
reads `first-name` through `sex` — so the walk that resolved a name cannot say where a
reader would type one. The table's own location can, which is why it keeps its path, and
why an ambiguity error names `sv_SE.sex[f].first-name[Kim]` rather than the table's own
name.

## No builtin reads the clock, so a date is bounded by days, never by an age

An `age(min,max)` would make a seeded fixture change with the day it runs on, which is
what a seed exists to prevent; a birthdate for someone 20 to 60 is
`date(1966-01-01,2006-12-31,…)`, re-pinned as any fixture is.

## A name column without a key resolves inside its parent

A given name both sexes carry is a row under each, so `name` cannot be the key; the
parent's row tells the two apart, `sex[f].first-name[Kim]`, the ambiguity error spells
each row inside its parent, and a name repeating inside one parent row is refused at
load, since nothing could then select it.

## `misc` is what every locale shares

A category whose facts differ by country belongs in that country's locale, read from the
register that country's own records use; `misc` takes only sources that are
international. NHTSA vPIC and Mobility Sweden's registrations are national, so they are
for `en_US.car` and `sv_SE.car`, and `misc.car` waits for an international source.

## A register's canonical spelling loses to the one its domain writes

Where a source offers several spellings of one fact, the shipped one is what records in
that domain carry. `misc.timezone` reads `zone.tab` and not the `zone1970.tab` that
supersedes it, because the latter keeps one zone per set of countries that have agreed
since 1970: it spells Sweden `Europe/Berlin`, and no Swedish system writes that. For the
same reason `misc.language` takes the ISO 639-2 register's first synonym over CLDR,
which says "Chinese, Mandarin" for `zh`.

## `misc.territory` is the spine, and a `misc` table naming a territory links to it

Where every territory row has a child the column is a `parent`, and the import drops the
child rows whose territory the set does not ship — 17 of `misc.timezone`'s, Antarctica's
ten among them. Agreement across a record is worth more than the last rows of a table.
Where no such link can hold the fact stays a column. Layer your own `misc.territory`
over the shipped one and you must layer `misc.timezone` too, or the link fails at load
naming the row.

## A table whose register publishes no frequency draws evenly

`misc.httpmethod`, `misc.port`, `misc.httpstatus`, `misc.mimetype` and `misc.tld` weigh
every row alike, so GET is a ninth of the methods drawn. Goal 12 keeps an authored fact
out of a sourced table, and no register publishes how often a method, a port or a TLD is
used, so a weight here would be invented. Where one exists it is read, as
`misc.timezone` reads GeoNames populations and `sv_SE.first-name` SCB bearers.

## `misc.timezone` weighs a zone by the people living in it

The weight is the population GeoNames records in the zone's cities of 15,000 or more,
floored at 15,000, which goal 14 asks for: over 300 seeded draws of
`misc.territory[US].timezone`, the four zones most Americans live in took 278 where an
even weight gave them 44, and `America/Indiana/Petersburg`, a town of 2,400, fell from 17
to 0. Valid while a draw weighted this way lands where people live.

## `misc.port` selects by number, and carries no `name`

29 of its services sit on more than one port, `http-alt` on three, so
`misc.port[http-alt]` could name no one row. The number is what a port field holds
anyway, and `misc.port[443].service` reads the other way.

## `misc.car` is one flat table, not a make linked to its models

A row is a make and a model drawn together, so no render pairs a Volvo with a RAV4. Two
linked tables would reach the same pairs and add a selector nothing asks for; a make
alone is `misc.car.make`.

## `misc.territory` names its sovereign in a column, and there is no `misc.country` table

ISO 3166-1 codes territories, so `territory` is the honest name, and `is_independent` in
the register gives each one's state. A second table of the 195 sovereigns would hold a
`DK` row beside the territory `DK` row, both carrying Denmark's capital, currency, flag
and TLD — two owners for one fact, drifting at the next import. Splitting the columns to
avoid that is worse: put `capital` on the territory alone and a country row can no
longer name Copenhagen. The column cannot be a `parent`, since a table never reaches its
own family; a test proves every value names a row instead. A territory the register
records no state for stands alone, which is `EH` alone, and naming one for it would be a
claim fejkdata has no business making.

## `misc.loglevel` is a table keyed by the code, rendering POSIX's keyword

A flat list of names carries neither the code a PRI encodes nor a selector reaching it,
and goal 3 draws the two as one fact. The canonical spelling losing to the one its
domain writes, above, settles the rest: a configuration writes `info`, so RFC 5424's
`Informational` stays the `severity` column, and the code and the keyword hold the two
selector slots.

## `misc.tld` keys carry the leading dot, where other tables key on a bare code

The register spells a TLD `.se` and `misc.territory.tld` already ships it so, which a
bare key would make two spellings of one fact; `{/misc.tld}` also composes onto a host
with no separator. `misc.tld[se]` misses for it, which `todo.md` carries.

## `misc.tld` is a table of its own, and `misc.territory.tld` stays a column

A `parent` demands a child for every parent row, so linking them would drop every root
zone row naming no territory, which is most of them, and goal 12 holds a sourced table
whole. The loader refuses the link outright anyway: `tld` is a column of
`misc.territory`, and a table may not be named like a column of its ancestor.

## `misc.territory` carries a currency code, it does not link to `misc.currency`

A `parent` demands a child for every parent row, and ISO 4217 registers codes no
country's row can name: the funds codes (Mvdol, WIR Euro, US Dollar (Next day)), and VED
beside VES, both Venezuela's, of which a country row names one. Linking would trade the
register for the link, and `misc.territory.currency` already pairs a country with its
currency in one draw.

## An extension may name two media types

`.xml`, `.rtf`, `.sub`, `.mpp` and `.ac` each name two rows of `misc.mimetype`.
Separating them would mean dropping a registered type, or naming one by an extension
that is not its own — `.mpt` is Project's template, not its document. Selecting such a
name is an error listing both keys; select by type instead.

## The Swedish ids draw Skatteverket's test series

A Luhn-valid personnummer over a random birth number may be a living person's; 238 and
239 after any date are blocked from assignment, so the shipped `personnummer` and
`samordningsnummer` use those. They sit in a `birth-number` table under `sex` rather
than as a column of it: the render's shared draw of the family is what makes the number
and the name agree on sex, and `sex` stays one shape across locales instead of
collecting every sex-keyed id fact. A samordningsnummer's day, the birthday plus 60, is
drawn from 61 to 88, valid in every month, rather than computed from the date drawn.

## The US given names come from a mirror of the SSA file

ssa.gov refuses a client outside the US, so `names-us.py` reads a GitHub copy that ends
at 2020, which a count over the births since 1930 barely feels; `--names` takes the
official zip. The SSA's placeholder rows are top-1000 entries that name nobody, so the
import drops them by name rather than by a rank a regeneration would move.

## `List` advertises direct descents only

`region.municipality.locality` is listed, and `region.locality` resolves too but is not:
the set of every descent through a chain of five tables is every subsequence of it, and
the direct chain is the one a reader can predict from the tables' parents.
