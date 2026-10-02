# Decisions

## GitHub is canonical, and the module path names it

2026-09-20, Lilleman auf Larv.

Goal 1 wants the usage the tool earns and goal 3 expects extenders who did not write it;
both need a stranger to file an issue and open a pull request. Gitea has no anonymous
issue, and no cross-host pull request at all, so a contributor would need an account and
a fork on a personal instance. Valid while the project wants contribution from outside:
the workflow decided nothing, being portable already — `github.api_url` and
`secrets.GITHUB_TOKEN` resolve on either host, so only the names moved.


## The Gitea copy stays, as a pull mirror of GitHub

2026-10-03, decided by the maintainer. Technical principle: one owner per value
(`~/.claude/principles/technical-principles.md`). GitHub owns the history, and
`gitea.larvit.se/larvit/fejkdata` stays as a read-only pull mirror of it, so it is never
a second owner. Until `todo.md` item 68 converts it, the Gitea copy goes stale. Valid
while the maintainer keeps a Gitea copy.

## The vocabulary sits below `doc.go`'s package clause, not in the package doc

2026-09-21, Lilleman auf Larv.

Goal 1 is what a developer choosing a library reads first, and the package doc is that
page: the glossary would fill two thirds of it with the render path's units, mostly
unexported, and Go doc comments have no code markup, so every name would render its
backticks literally. Below the clause it reaches goal 3's reader — who opens the file —
and nothing else. Valid while the glossary's subject is the render path; one written to
teach the exported API belongs above the clause.

## Options and fields share one namespace

2026-09-02, Lilleman auf Larv.

`format`, `weight`, `repeat`, `separator`, `datatype` and `drawGroup` are reserved;
every other key is a field. Nesting fields under a key, or prefixing options, would tax
every template to guard against a misspelt option.

`todo.md` item 4 deletes `drawGroup`, which this reserves.

## `{a|b}` stays beside nested choices

2026-09-02, Lilleman auf Larv.

`[[…], […]]` picks the same way, but its arms are anonymous; `{female|male}` keeps
`person.female` addressable.

## Flags follow getopt_long

2026-09-02, Lilleman auf Larv.

`--name value` and `--name=value` both work; a short flag's value attaches or follows
(`-s42`, `-s 42`) and short flags bundle (`-hn 3`), as every shell user expects. A
single-dash long flag is rejected naming the double-dash spelling, and `-s=42` is
rejected naming both short spellings: `=` belongs to the long form, and reading `=42` as
the value would make `-d=./x` a directory named `=./x`.

## An argument is a template by its shape, not by a flag

2026-09-03, Lilleman auf Larv.

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

2026-09-03, Lilleman auf Larv.

`New` proves the loaded tree acyclic, an inline node is a finite tree of its own, and
nothing in the tree can reference it, so no render of it reaches itself.

## An inline template that does not compile is misuse (exit 2), including a reference that resolves to nothing

2026-09-24, Lilleman auf Larv.

The whole argument is the spelling under test, and `NewTemplate` compiles, links and
validates as one step. An unknown *path* stays a runtime error (exit 1): there the
argument is well-formed and only the data is absent.

## A padded JSON argument is rejected, not trimmed

2026-09-03, Lilleman auf Larv.

Padding is the one place the two readings disagree — a format string renders it, JSON
drops it — so the spelling that renders is named rather than silently chosen.

## `FakeTemplate` and `NewTemplate` both stay

2026-09-03, Lilleman auf Larv.

They reach the same value but not at the same cost: `NewTemplate` pays the compile and
validation once and renders many times, `FakeTemplate` is the one-shot call, and
`--repeat` is exactly the case that needs the first. The pair is `regexp.MustCompile`
and `regexp.Match`, not two spellings of one result.

## The shipped data is embedded, not discovered

2026-09-02, Lilleman auf Larv.

A directory a machine happens to have would make `--seed 42` machine-dependent. Data
still lives in `data/` as JSON; `--data-path` layers over it.

## A bare reference draws each time; a reference path is held

2026-09-02, Lilleman auf Larv.

Goal 5: `{/p} {/p}` is two draws, as `{word} {word}` is, while every `{/p.first}` in
one render and draw group reads one draw, and a bare `{/p}` beside them is a load error,
as `{p}` beside `{p.first}` is: a bare token is by contract an independent draw, a path
pins its level, and a fresh draw of a pinned level could show another row. A builtin's
operand holds what it reads for its expansion, references included, so
`{uppercase(/p)} {/p}` is one draw, and `{/p}` written twice beside it is a load error.

Goal 5.2 overrides this; `todo.md` item 4 replaces it.

## Reference sigils follow the filesystem

2026-09-02, Lilleman auf Larv.

`/` is the root, `.` this file's folder, `..` the folder above — what those spellings
already mean to anyone who has typed a path. A locale's files reach each other without
naming the locale, so a folder renames and copies without editing its references.

## A change to what exists is a major; a minor only adds

2026-09-16, Lilleman auf Larv.

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

2026-09-16, Lilleman auf Larv.

Any edit to a category shifts its stream and everything drawn after it, so a promise
across versions would freeze every shipped list; a fixture is re-pinned on a bump, as
this repo's own are.

## An error is a contract by what it names, not its bytes

2026-09-16, Lilleman auf Larv.

Goal 6. A script branches on the exit code and reads the named path or spelling, so
those hold; wording improves in a minor.

## A load reports every mistake at once

2026-09-30, Lilleman auf Larv. Goal 6. An author learns everything wrong with the data
in one load, not one mistake per run: `New` and `NewTemplate` report every mistake in one
error, ordered by the path each names. A mistake following only from another, such as a
reference into a category that did not compile, is left out, so each one names something
to fix. Valid while a load checks the whole set before its first render.

## Formats are checked after all data is loaded, so their errors name the category

2026-09-28, larv-review; approved 2026-09-28 by lilleman. Goal 3. Loading reads every
file first, then connects references across them. A format holding a reference can only
be checked once it is connected, so every format is checked then, and all checks run at
one moment. An error from those checks names the category path you would type (`sub.x`),
or a table cell's category and line, not the file. Valid while checking a format needs
its references connected.

## Raising the lowest supported Go is a major

2026-09-16, Lilleman auf Larv.

A consumer building on it breaks, which is the one test every rule above applies; Go's
convention of a minor is not followed.

## The format check runs on the latest Go only

2026-09-21, Lilleman auf Larv.

`gofmt`'s output is the toolchain's, not the code's: 1.27 stopped padding a map
literal's values out to a lone long key, so no source satisfies both it and 1.22's. A
consumer on the lowest supported Go depends on the code compiling and its tests passing
there, which is the `portable` stage, and never runs `gofmt` over this source, so the
latest toolchain alone defines the one canonical form goal 3 asks a reader to meet.
Valid while the lowest supported Go is not the latest.

## The changelog heading is the one spelling of a release; CI cuts the tag

2026-09-16, Lilleman auf Larv.

A tag pushed by hand is served by `go get` at once, so a tag whose commit lacks its
heading is burnt, not fixed. The heading on a gate-passed `main` commit is the trigger
instead: the tag can land only there, and the GitHub release the same job publishes
keeps one text as its body and is where prebuilt binaries will attach.

## A `--data-path` override rebinds every reference to the category it replaces

2026-09-24, Lilleman auf Larv.

References bind against the merged tree, so once shipped data uses `{.person}`, a
consumer's `sv_SE/person.json` is what every shipped reference into `person` reads, and
`New` fails on shipped data the consumer never wrote when that file lacks a field those
references read. Accepted: overriding is the point of layering, the error names the
reference and the field, and the fix is the consumer's file carrying the fields the
shipped tree reads.

## The repeat cap bounds renders, not bytes

2026-09-02, Lilleman auf Larv.

A repeat, alone or nested, may ask for at most 1 048 576 renders; how large each render
is stays what the data asked for, so `{hex(1048576)}` repeated to the cap is a terabyte,
loaded without complaint. A byte estimate would need every builtin to declare a width to
fence a shape no data comes near, and the harm lands on the author who wrote it.

## 64-bit targets only

2026-09-02, Lilleman auf Larv.

The gate builds amd64, and the buffer sizing a render pre-computes (renders × bytes)
assumes a 64-bit int; on a 32-bit target it could overflow and panic.

## A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`

2026-09-24, Lilleman auf Larv.

`1/0` and a fixed `"0"` field are decidable, so they join the never-numeric operand as a
load error; the fold stops where an operand varies, so `a/(b*c)` with `b` fixed at `0`
and `c` varying loads and prints `Inf` every draw — catching it needs zero-absorbing
algebra for a shape nobody writes. A [typed column](../README.md#datatype) bounds its operands
instead and refuses a divisor it cannot keep from zero.

## In data, a default written out and a constant spelled as a sample are load errors

2026-09-24, Lilleman auf Larv.

`weight: 1`, `repeat: 1`, `separator: ""`, `datatype: "string"`, `int(5,5)`,
`float(1,1,2)`, `+5` and `05` each spell what a shorter form already spells, so each is
rejected naming that form. The CLI's numbers follow the shell instead: `--seed 007` and
`--repeat +3` are 7 and 3, as every command line reads them.

## Samples say what they emit, transforms what they do

2026-09-02, Lilleman auf Larv.

`{upper(2)}` is two letters, `{uppercase(x)}` is `x` upper-cased; one name for both
would turn on whether the argument looks like a number.

## A record is a template seen as columns; a Go struct is the one second schema

2026-09-15, Lilleman auf Larv.

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

2026-09-15, Lilleman auf Larv.

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

2026-09-15, Lilleman auf Larv.

Every reference path into a category that selects no row, in one `Fake` or one record,
reads one draw of it, so a value's facts agree across its fields, nested templates and
columns alike — `{/currency.code}` in one field and `{/currency.symbol}` in another name
one currency, whichever view renders them. A `repeat` iteration is a render of its own, since
repeating asks for another entity, and a [draw group](../README.md#draw-group) names further
entities within one render, so a payer and a payee are two groups over one `person`
rather than two copies of it. Only references share: a sibling field is local to its own
expansion, so a `first` column does not silently bind to a `first` in the column next to
it.

Goal 5.2 overrides this; `todo.md` item 4 replaces it.

## A draw group name is local to its category

2026-09-16, Lilleman auf Larv.

A category's groups are its own entities, so a caller naming a group the same way never
joins them by accident, and renaming a group inside one file changes no render
elsewhere. The unnamed group still spans categories, since facts that belong together
across categories must agree.

`todo.md` item 4 deletes what this decides.

## The expansion hold and the render's draws are two fences

2026-09-16, Lilleman auf Larv.

Goal 3: one proves a sibling path or an operand is reached only by its readers within
an expansion, the other that reference paths, whole-table reads and a table family's
rows agree across a render and its draw groups. Their scopes differ and only the second
tracks rows and pins, so one walk would carry both scopes and tell them apart at every
step.

`todo.md` item 4 deletes what this decides.

## A render's draws make their maps on the first read

2026-09-29, Lilleman auf Larv.

Every render starts from empty draws, its draw groups included, since goal 3 prices
a reader meeting two spellings of one empty set. Making the maps where the draws are declared kept a
record's on its frame's stack, but a `Fake`'s draws live on the `Generator`, so there it
cost the cheapest render two heap allocations and 240→390 ns, measured 2026-09-28.
Lazily, a render reading a reference path pays those two allocations, a record or struct
included, and one reading none pays nothing; goal 12 holds either way.

`todo.md` item 4 deletes what this decides.

## A category never references itself, and a record's fences run at load

2026-09-16, Lilleman auf Larv.

A category is one unit: a reference back into it — `{/users.first}` inside `users` —
describes a draw other than the fields beside it, where goal 4 asks facts that belong
together to come from one draw, so `New` refuses it and the sibling path stays the one
spelling for a field of one's own. A value two fields share goes in its own category,
which both reference. That settled, a record's column fences run at `New` too, so a
category that loads renders as whichever shape is asked for, and a reference reaching
back into a category through another one is refused there as the overlap it is. Which
reads those fences weigh differs on purpose: a record-only template's inert format
renders nothing, so a `drawGroup` on it can never matter and is refused, while one on a
rendering format can matter to a caller that bare-references it.

`todo.md` item 4 deletes the `drawGroup` refusal this names.

## A record's column set is fixed before the first draw

2026-09-04, Lilleman auf Larv.

Only a category-level template is a record: a path descending into a field, or naming a
folder or a choice, errors. A tail may pass through a choice whose variants carry
different fields, so the columns — and with them the CSV header written once ahead of
every row — would vary per draw. A fixed column set is what the CSV and `INSERT`
contracts rest on, so the restriction holds even where a particular choice would happen
to agree.

## Null is a `null` item, not a rate

2026-09-15, Lilleman auf Larv.

A null is one more outcome of a column's draw, so a choice's weights skew it like any
other; a null-rate option would be a second way to state odds.

## A typed column holds one value, not composed text

2026-09-15, Lilleman auf Larv.

Its bounds come from a literal or a call's arguments, so a load error names a real
value, a range check is one comparison, and `1{digits(2)}` is a second spelling of
`{int(100,199)}`.

## A column of one reference alone is the column it reads

2026-09-15, Lilleman auf Larv.

`{/src.score}` renders exactly what `src.score` draws, so it takes that column's
datatype and null rather than restating them, and a `datatype` restating the one it
takes is a second spelling. Any other `datatype` still types the values — the one way to
type a column someone else wrote.

## A typed column's calc is refused unless proven

2026-09-15, Lilleman auf Larv.

Operand bounds must keep each divisor from zero and the result finite; what they cannot
show is refused rather than trusted, since a bare `NaN` breaks the JSON and SQL it lands
in.

## `Column` carries text, not a Go value

2026-09-15, Lilleman auf Larv.

`Value` is the rendered string beside `DataType` and `Null`, which each serializer
writes as the load check proved it; a `Value any` would hand every caller a type switch.

## A unit takes the stem of what it is, and a file the stem of the units it holds

2026-09-29, Lilleman auf Larv.

Goal 3 wants a name to reach one unit: `drawMemo` is what a draw is kept in, `readValue` what
a read drew.

## One name, one meaning

Two declarations of any kind — function, method, field, type — share a name only where one
definition covers both: the same kind of value, counted or addressed the same way, so a
reader landing on the wrong one concludes nothing false, as `Fake` renders one value on
`Generator`, `Template` and `RecordTemplate`. `TestNoFunctionSpellsAMethod` keeps a bare function off a method's
name, since a search returns both and a call site shows no receiver; review judges the
rest. Serves goal 3: names that tell the truth, and reaching the unit behind a symptom
without asking a person. Decided 2026-09-25 by Lilleman; valid while goal 3 counts reading
cost.

## The package stays flat

2026-09-15, Lilleman auf Larv.

Go ties a package to one directory, so folders would split the API into packages.

## The performance gate asserts allocations, not wall-clock time

2026-09-03, Lilleman auf Larv.

`AllocsPerRun` is deterministic across machines, so a 10% ceiling does not flake under
CI load, while time varies with the machine and its neighbours. A rendering slowdown
almost always costs an allocation too (a lost pre-size, a per-item map, an extra copy).
The benchmark suite (the README's Development) reports time for a human, not as a pass/fail gate.

## Rows live in a TSV, the shape in JSON

2026-09-17, Lilleman auf Larv.

`New` allocates once per node, so a register of thirty thousand rows written as JSON
objects would cost it a second; a TSV is one allocation whose cells are substrings, and
the JSON says only how a row is composed. The TSV sits beside its category file, named
by `rows`, so a data directory stays a directory of categories, and one nothing names is
a load error rather than a file silently ignored.

## A selector is bracketed, and a dot inside it is literal

2026-09-17, Lilleman auf Larv.

`municipality[0180]` reads as selection to anyone who has indexed an array, and `[St.
Louis]` keeps a name whole where a colon or a dot-separated spelling could not; zsh
needs the brackets quoted, which the README's examples show.

## `parent` names the link column and the table alike

2026-09-17, Lilleman auf Larv.

One word says both, so a child table sits in its parent's folder and links on a column
of the parent's name; the geo plan wants exactly that, and a table needing another
folder or another column name would be asking for a second spelling.

## After a row, a path names a column or a linked table

2026-09-17, Lilleman auf Larv.

`locality[Lund].address`, a template beside the family under a selected row, is refused
today; admitting it later is additive, since a refused spelling gains a meaning and no
accepted one changes, so the door stays open for the address records the plan describes.

## A selected row is fixed, not drawn, and an unnamed read beside it is refused

Decided 2026-09-27 by the maintainer, for goals 4 and 5; valid while a selector names
exactly one row.

- Selected rows never conflict: `{/misc.territory[SE].capital} / {/misc.territory[FI].capital}`
  renders `Stockholm / Helsinki`, and two items of one choice may select different rows.
- What a path draws beneath a selected row is one draw per selection, per render and
  group, shared by every path through it: `{/geo.SE.municipality[1281].locality.name}`
  twice names one locality.
- A read whose path selects no row is refused beside a selection in its family, in the
  same render and group, naming the rewrite through each selection and a `drawGroup`:
  beside `{/misc.territory[SE].capital}`, a writer reading `{/misc.territory.currency}`
  expects `SEK`, so the error names `{/misc.territory[SE].currency}`, since
  `Stockholm pays in EUR` would pass review. A second selection does not make the
  unnamed read load.

Checked against three simulated template writers from the Audience, who agreed on each
case but the last, where two expected a render and all three accepted the refusal.

Goal 5.2 overrides its shared draw beneath a selection and its refusal of an unnamed read; `todo.md` item 4 replaces them.

## A link column after a row steps up to the parent row, and a path may not end on it

Decided 2026-09-27 by the maintainer, for goal 5; valid while a link column holds the
parent's key. It extends [After a row, a path names a column or a linked
table](#after-a-row-a-path-names-a-column-or-a-linked-table) from child tables to the
parent. A row is selected or drawn alike.

- Where a segment follows, the link column's name reads the parent row, one level per
  step, chained: `city[Oslo].country.name` renders `Norway`, `locality.municipality.name`
  the drawn locality's municipality, and
  `geo.US.locality[1714000].municipality.region.name` Chicago's state.
- A path ending on it is refused at `New`, `NewTemplate` and `Fake` alike, naming the
  parent's key column and the parent's format spelled as paths: with a `country` format
  of `{name} ({alpha2})`, `city[Oslo].country` names `city[Oslo].country.alpha2` and
  `{/city[Oslo].country.name} ({/city[Oslo].country.alpha2})`. `List` leaves such a path
  out, and a record's link column still holds the key, since it is a column, not a path.
- A path may not step back down into a child table after a step up:
  `city[Oslo].country.city.name` could be Oslo or a city of Norway drawn afresh, so it is
  refused, naming the path to the row it stepped up from, `city[Oslo].name`, and the path
  down from the parent, `country[NO].city.name`; after a drawn row, `city.country.city.name`
  names `city.name` and `country.city.name`.
- A route through another row is another path to the same data, not a second spelling:
  `city[Oslo].country.name` and `country[NO].name` both load.

Checked against three simulated template writers from the Audience: all three read the
parent row where a segment followed, and they split between the code and the name where
the path ended.

Goal 5.4 overrides this; `todo.md` item 2 replaces it.

## A table selects by one key column, a code, never a free-form name

Decided 2026-09-27 by the maintainer, for goal 5's one spelling per result; valid while
the key is the code a user writes, as a database keys a table. `country[NO]` loads, and
`country[Norway]` is refused naming `[NO]`, found by the row whose cell spells the
selector. The key is the friendliest code the register holds, `misc.territory[SE]` and
`misc.currency[EUR]`. There is no `name` option, and a table nothing selects into, a name
table, `street`, `car` or `useragent`, carries no key and is only drawn.

- `geo.US.locality` keys on the Census GEOID, `[1714000]` for Chicago.
- `geo.SE.locality` keys on an integer the import assigns, since a postort has no code of
  its own. The first import numbers the postorter in name order; later imports keep each
  id by the postort's name, unique across Sweden as the key it replaces, give a new
  postort the next unused integer, and retire a removed or renamed one's id for good.

## A parent row with no child row is a load error

2026-09-17, Lilleman auf Larv.

A descendant is drawn inside the nearest pinned ancestor, so every ancestor row must
lead to a row at every level below it, or a render could find nothing to draw. The
import script drops or fills such rows; the alternative, falling back to a free draw,
would break the consistency the link exists for without saying so.

## The choice-of-rows fence guards a data file's root, and requires string fields

2026-09-17, Lilleman auf Larv.

A table is a category with a TSV beside its file, so only a root choice has the spelling
the fence names; a nested choice of same-shaped templates and an inline one keep
loading. Fields must all be strings because a cell is a string node: a choice whose
items carry a nested choice is not one table but two linked ones.

## A table is a record of string columns

2026-09-17, Lilleman auf Larv.

Its columns are the CSV header and the `INSERT` column list, fixed by the TSV header, so
a table is a record by construction; every column is a string until a typed column
option earns its place.

## The key map is built at load, the rest on first draw

2026-09-28, Lilleman auf Larv.

A link is proved against the parent's keys and a key's uniqueness is a data mistake, so
both are load-time; the row lookup, by name and by parent, serves only a draw, so it waits
for the first one, keeping `New` linear in the bytes read.

## Two categories may name one TSV

2026-09-17, Lilleman auf Larv.

Each is a view of the file with its own format and options, at the cost of holding the
rows twice, which is what a category over a register with two natural formats asks for;
a TSV nothing names stays a load error, since that one is a file forgotten rather than
shared.

## A table never reaches its own family, by any route

2026-09-17, Lilleman auf Larv.

A `repeat` iteration and a `drawGroup` each draw apart on purpose, but a row that lists
three localities from other regions is the output the family exists to prevent, so the
own-family fence walks through both rather than stopping where the draw fences do.

`todo.md` item 4 deletes what this decides.

## A cell may read a reference, and each row its own

2026-09-29, Lilleman auf Larv.

Goal 4: facts that belong together come from one draw, so a shop row's `phone` cell reads
`{/sv_SE.phone}` on a Swedish shop and `{/en_US.phone}` on a US one, and the table's
format, the same for every row, cannot say that. A restructure for goal 3 keeps the fences
that prove such cells safe. Valid while a table's rows need generators of their own.

`todo.md` item 4 deletes the fences this keeps; a cell may still read a reference.

## Tables carrying token cells stay small

2026-09-17, Lilleman auf Larv.

The family fence compares the reads of every pair of rows that can render together, so a
table whose every row's cell selects a row of another table loads in time quadratic in
its rows: about a second at four thousand rows. No shipped table carries such cells, and
a register is a column set rather than a set of references, so the fence is left as it
is until a real data set needs the indexed form.

`todo.md` item 4 deletes what this decides.

## A path is walked once without drawing before it is walked for real

2026-09-17, Lilleman auf Larv.

Goals 9 and 12: a path that fails moves no seeded stream, at the cost of one draw-free
walk per call, compiling its steps into a stack buffer of 16. Valid while that walk
costs little against the draw it guards.

## A country's postal codes and streets are siblings under its locality

2026-09-18, Lilleman auf Larv.

No open source pairs a Swedish street with its postnummer, and pairing the US through
its ZIPs would shape the two trees differently, so both draw inside the pinned locality
and an address agrees at that level. A street's own code is the exact pairing to add
when a source carries it.

## A locale's `address` reads its country's `geo` tree, so the shipped set loads whole

2026-09-24, Lilleman auf Larv.

`data/sv_SE` alone no longer loads: a test loads `data` and prefixes the locale, and
`--no-shipped-data -d` takes the whole `data` folder or a set of one's own.

## The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more

2026-09-24, Lilleman auf Larv.

Sweden fits whole in 700 KB; every US place of 10,000 would pass a megabyte and fetch
1,200 counties of TIGER files, so the threshold sits where the two countries match in
size, and `--min-population` and `--streets-per-locality` on the import scripts build a
fuller set. The two trees add about 20 ms to `New`, which loads the shipped set in about
45 ms.

## A locale's `address` restates its country record's format

2026-09-18, Lilleman auf Larv.

A record cannot read another whole and keep its columns, so `sv_SE.address` names the
same four columns as `geo.SE.address`, each a reference into it, and the format appears
twice; a column is spelled the same in both, `street-number`, so the two never disagree
on a name.

## A postort's kommun comes from its name, its tätort or its codes, never from distance

2026-09-24, Lilleman auf Larv.

GeoNames leaves a fifth of Sweden's codes without a kommun and carries stale spellings;
the nearest code across a border named the wrong kommun half the time it was tried, so a
postort none of the three rules place is dropped, as is one not cased like a place name.

## A highway designation is not a street, and a US postal code belongs to the place holding most of its land inside places

2026-09-24, Lilleman auf Larv.

`I- 55 Bus` and `US Hwy 1` carry the most address ranges in many places and would head
every address, so the import drops names spelled as a route. A ZCTA goes to the place
its largest in-place part lies in, census-designated places left out since they never
ship, and ships only when that place does, so a few dozen places whose every code lies
mostly in a bigger neighbour ship no address; counting the land outside every place too
would drop a quarter of the places, whose codes straddle unincorporated land, for a
postal city the USPS mostly names the same way.

## `--list` stays a plain list of paths

2026-09-18, Lilleman auf Larv.

It is what a script reads, so every line has to be a path that `Fake` takes; a marker
for the tables a `[selector]` follows, or a legend above them, would make the output
something to parse before use. `--help` names the selector spelling instead, and the
Table section teaches it.

## A layout is always quoted

2026-09-18, Lilleman auf Larv.

A layout may carry the comma that separates arguments, `'January 2, 2006'`, and one
spelling for every layout beats a rule about which ones need the quotes, so the bare
spelling is refused naming the quoted one. The layout is Go's reference time because the
library renders with it and a Go caller already knows it; its names are English, and a
locale's own month and weekday names are data.

## A title is a table under `sex`

2026-09-18, Lilleman auf Larv.

A prefix drawn apart would put `Mr` on a record whose `sex` column says `female`, which
is the disagreement the record exists to prevent; the tables this set already has are
what a title needs, so `en_US.title` links to `sex` as `first-name` does. Swedish has no
everyday sexed honorific, so `sv_SE.title` is a table as well but carries no `parent`.
Its weight column is `share`, not the `count` a name table carries, because the values
are a curated proportion rather than bearers anyone counted.

## A table owns the spelling of a selector on it

2026-09-18, Lilleman auf Larv.

A reference reaches a table by a path that carries no selector — `sv_SE.address` reads
`geo.SE.locality` through its template — so the walk that reached it cannot say where
a reader would type a selector. The table's own location can, which is why it keeps its
path, and why an error names `geo.SE.municipality[1281]` rather than `municipality[1281]`.

## No builtin reads the clock, so a date is bounded by days, never by an age

2026-09-18, Lilleman auf Larv.

An `age(min,max)` would make a seeded fixture change with the day it runs on, which is
what a seed exists to prevent; a birthdate for someone 20 to 60 is
`date(1966-01-01,2006-12-31,…)`, re-pinned as any fixture is.

## `misc` is what every locale shares

2026-09-18, Lilleman auf Larv.

A category whose facts differ by country belongs in that country's locale, read from the
register that country's own records use; `misc` takes only sources that are
international. NHTSA vPIC and Mobility Sweden's registrations are national, so they are
for `en_US.car` and `sv_SE.car`, and `misc.car` waits for an international source.

## A register's canonical spelling loses to the one its domain writes

2026-09-18, Lilleman auf Larv.

Where a source offers several spellings of one fact, the shipped one is what records in
that domain carry. `misc.timezone` reads `zone.tab` and not the `zone1970.tab` that
supersedes it, because the latter keeps one zone per set of countries that have agreed
since 1970: it spells Sweden `Europe/Berlin`, and no Swedish system writes that. For the
same reason `misc.language` takes the ISO 639-2 register's first synonym over CLDR,
which says "Chinese, Mandarin" for `zh`.

## `misc.territory` is the spine, and a `misc` table naming a territory links to it

2026-09-24, Lilleman auf Larv.

Where every territory row has a child the column is a `parent`, and the import drops the
child rows whose territory the set does not ship — 17 of `misc.timezone`'s, Antarctica's
ten among them. Agreement across a record is worth more than the last rows of a table.
Where no such link can hold the fact stays a column. Layer your own `misc.territory`
over the shipped one and you must layer `misc.timezone` too, or the link fails at load
naming the row.

## A table whose register publishes no frequency draws evenly

2026-09-20, Lilleman auf Larv.

`misc.httpmethod`, `misc.port`, `misc.httpstatus`, `misc.mimetype` and `misc.tld` weigh
every row alike, so GET is a ninth of the methods drawn. Goal 13 keeps an authored fact
out of a sourced table, and no register publishes how often a method, a port or a TLD is
used, so a weight here would be invented. Where one exists it is read, as
`misc.timezone` reads GeoNames populations and `sv_SE.first-name` SCB bearers.

## `misc.timezone` weighs a zone by the people living in it

2026-09-24, Lilleman auf Larv.

The weight is the population GeoNames records in the zone's cities of 15,000 or more,
floored at 15,000, which goal 15 asks for: over 300 seeded draws of
`misc.territory[US].timezone`, the four zones most Americans live in took 278 where an
even weight gave them 44, and `America/Indiana/Petersburg`, a town of 2,400, fell from 17
to 0. Valid while a draw weighted this way lands where people live.

## `misc.car` is one flat table, not a make linked to its models

2026-09-18, Lilleman auf Larv.

A row is a make and a model drawn together, so no render pairs a Volvo with a RAV4. Two
linked tables would reach the same pairs and add a selector nothing asks for; a make
alone is `misc.car.make`.

## `misc.territory` names its sovereign in a column, and there is no `misc.country` table

2026-09-24, Lilleman auf Larv.

ISO 3166-1 codes territories, so `territory` is the honest name, and `is_independent` in
the register gives each one's state. A second table of the 195 sovereigns would hold a
`DK` row beside the territory `DK` row, both carrying Denmark's capital, currency, flag
and TLD — two owners for one fact, drifting at the next import. Splitting the columns to
avoid that is worse: put `capital` on the territory alone and a country row can no
longer name Copenhagen. The column cannot be a `parent`, since a table never reaches its
own family; a test proves every value names a row instead. A territory the register
records no state for stands alone, which is `EH` alone, and naming one for it would be a
claim fejkdata has no business making.

## `misc.loglevel` is a table keyed by POSIX's keyword, carrying the code

2026-09-27, Lilleman auf Larv.

A flat list of names carries neither the code a PRI encodes nor a selector reaching it,
and goal 4 draws the two as one fact. The canonical spelling losing to the one its
domain writes, above, settles the rest: a configuration writes `info`, so RFC 5424's
`Informational` stays the `severity` column.

## `misc.tld` keys carry the leading dot, where other tables key on a bare code

2026-09-20, Lilleman auf Larv.

The register spells a TLD `.se` and `misc.territory.tld` already ships it so, which a
bare key would make two spellings of one fact; `{/misc.tld}` also composes onto a host
with no separator. `misc.tld[se]` misses for it, which `todo.md` item 44 ends.

## `misc.tld` is a table of its own, and `misc.territory.tld` stays a column

2026-09-20, Lilleman auf Larv.

A `parent` demands a child for every parent row, so linking them would drop every root
zone row naming no territory, which is most of them, and goal 13 holds a sourced table
whole. The loader refuses the link outright anyway: `tld` is a column of
`misc.territory`, and a table may not be named like a column of its ancestor.

## `misc.territory` carries a currency code, it does not link to `misc.currency`

2026-09-18, Lilleman auf Larv.

A `parent` demands a child for every parent row, and ISO 4217 registers codes no
country's row can name: the funds codes (Mvdol, WIR Euro, US Dollar (Next day)), and VED
beside VES, both Venezuela's, of which a country row names one. Linking would trade the
register for the link, and `misc.territory.currency` already pairs a country with its
currency in one draw.

## An extension may name two media types

2026-09-18, Lilleman auf Larv.

`.xml`, `.rtf`, `.sub`, `.mpp` and `.ac` each name two rows of `misc.mimetype`.
Separating them would mean dropping a registered type, or naming one by an extension
that is not its own — `.mpt` is Project's template, not its document. Selecting such a
name is an error listing both keys; select by type instead.

## The Swedish ids draw Skatteverket's test series

2026-09-18, Lilleman auf Larv.

A Luhn-valid personnummer over a random birth number may be a living person's; 238 and
239 after any date are blocked from assignment, so the shipped `personnummer` and
`samordningsnummer` use those. They sit in a `birth-number` table under `sex` rather
than as a column of it: the render's shared draw of the family is what makes the number
and the name agree on sex, and `sex` stays one shape across locales instead of
collecting every sex-keyed id fact. A samordningsnummer's day, the birthday plus 60, is
drawn from 61 to 88, valid in every month, rather than computed from the date drawn.

## The US given names come from a mirror of the SSA file

2026-09-18, Lilleman auf Larv.

ssa.gov refuses a client outside the US, so `names-us.py` reads a GitHub copy that ends
at 2020, which a count over the births since 1930 barely feels; `--names` takes the
official zip. The SSA's placeholder rows are top-1000 entries that name nobody, so the
import drops them by name rather than by a rank a regeneration would move.

## `List` advertises direct descents only

2026-09-17, Lilleman auf Larv.

`region.municipality.locality` is listed, and `region.locality` resolves too but is not:
the set of every descent through a chain of five tables is every subsequence of it, and
the direct chain is the one a reader can predict from the tables' parents.

## A path draws through its compiled steps

2026-09-30, larv-review on the comprehension round the maintainer approved on 2026-09-29;
replaces "The path walks are separate loops".
Goals 3 and 12: `drawSteps` draws every path from its steps, compiled at link by
`pathCheck` or, for a caller's path, by `probePath` into a stack buffer. The check and
the probe stay two loops: on 2026-09-29 one loop over all three walks, switching on a
mode field, leaked the check's leaves and errors with the draw's pins, arm and memo,
since Go tracks a struct's fields as one, and a repeat of a reference path rose from 66
to 106 allocations. `drawSteps` takes its state as parameters and a step holds no node,
so no benchmark gained an allocation. Valid while Go's escape analysis tracks
a struct's fields as one.

## A name lives in the category binding it, or in the repeat binding it, and is drawn on its first read

2026-10-03, larv-review. Goals 5.3 and 5.6.
`bindNames` scopes a name at compile, and a render keeps one `pickFrame` per scope. One
field binds a name where another reads it, `"first": "{/person as p}{p.first}"` beside
`"last": "{p.last}"`, so the scope cannot be the template holding the binding; the category is the smallest unit holding every field and record column. The
scope is lexical: a category referencing another never sees its names, so what a category
renders never depends on who references it. A pick is drawn on its first read, so a record's
columns, rendered in name order, read one pick whichever binds it. A read entering a category
past its root, `Fake("cat.field")` or `{/cat.field}`, keeps the category's frame in the memo of
that read: per render and group, as the read's own draw is. Valid while a name is read only
inside the category binding it.
