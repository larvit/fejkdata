# Decisions

## GitHub is canonical, and the module path names it

2026-09-20, Lilleman auf Larv.

Goal 1 wants the usage the tool earns and goal 3 expects extenders who did not write it;
both need a stranger to file an issue and open a pull request. Gitea has no anonymous
issue, and no cross-host pull request at all, so a contributor would need an account and
a fork on a personal instance. Valid while the project wants contribution from outside.

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

`format`, `weight`, `repeat`, `separator` and `datatype` are reserved; every other key
is a field. Nesting fields under a key, or prefixing options, would tax every template to
guard against a misspelt option.

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

## A struct tag is a template by its shape

2026-09-03, Lilleman auf Larv.

A JSON object, array or string, or a string carrying a `{` token, is an inline template;
anything else is a path. A name may not contain a brace, a bracket or a quote, so a path
can never collide with any of those spellings, and the leading `[` or `"` is gated on
valid JSON so a stray copied bracket never swallows a tag — it names nothing, and says
so. Reserving the characters whole — though only a leading one
could collide — keeps one simple name rule instead of a leading-position special case.
A JSON string stays a template even where its bare text renders alike, `"x {/a}"` beside
`x {/a}`: inside the quotes, `\t` and `\n` let a one-line template write a tab or a newline. A
folder-relative `{.name}` or `{..name}` is refused naming
`{/name}`, since an inline template sits in no folder. `IsTemplate` exports the rule, so
struct tags and any other caller read one.

## The CLI renders exactly the template on its stdin, and adds nothing

2026-10-07, Lilleman auf Larv. Serves goals 7, 7.1 and 7.2.

The template comes from stdin, from `echo`, a file or a heredoc, so one command line gets
a value and a quoted heredoc needs no escape. A format string is stdin's every byte, and
fejkdata adds no newline of its own under `--format text`: `echo` ends each render with its newline,
`printf '%s'` sends none, and `-n` joins renders with `--separator`, empty by default. A
bench of the README's personas picked this on 2026-10-07 over dropping a newline or
adding one: a rule a user states in one sentence, at the cost of `printf` or a
`--separator` now and then.

Stdin holding only whitespace is a format string, and renders as sent.

A lone reference, `{/users}` or `"{/users}"`, names the record `users` under `--format`,
one newline after it allowed, with `users` the default `--table`; one to nothing there
exits 1. `todo.md` item 129 moves that reading into the library; until then it works
against goal 3.2.

The CLI takes no argument but its flags: one is misuse naming the spelling that reads what
it likely means, `fejkdata < FILE` for a file's name, dropping it for `-`, and the template
piped in otherwise; so is a terminal with nothing piped in, which would otherwise wait for
typed input. `--list` prints paths, each read as `{/path}`. Valid while the CLI reads one
template per run.

## The CLI reads only the library's public API

2026-10-07, Lilleman auf Larv. Serves goal 9.6: a CLI carrying your own modules is a few
lines of Go and works just like `fejkdata`, so `fejkdata`'s own CLI is such a CLI and
its non-test code imports nothing under `internal/`. Until `todo.md` item 129 moves it into
the library, the CLI copies how the grammar tells a lone reference, in `isPath`,
`hasEmptySegment` and `withoutSelectors`, and a test pins the copy to the grammar.

## A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129

2026-09-24, Lilleman auf Larv.

The whole template is the spelling under test, and `NewTemplate` compiles, links and
validates as one step. Under the CLI's `--format`, a lone reference to nothing exits 1
until item 129: it reads a path, and only the data is absent.

## Whitespace around a JSON object, array, string or null is dropped, and a bare number or boolean is a format string

2026-10-07, Lilleman auf Larv. Applies KISS.

Padding is where the two readings disagree: a format string renders it, JSON drops it. A
template that parses as a JSON object, array, string or `null` is its JSON, so its padding
is dropped, as `encoding/json` drops it. `null` stays JSON, since it is how a template
renders nothing. A bare number or boolean gains nothing from the JSON reading, so it is a
format string, and `echo 42` prints its newline as `echo x` does (goal 5.1).

## `FakeTemplate` and `NewTemplate` both stay

2026-09-03, Lilleman auf Larv.

They reach the same value but not at the same cost: `NewTemplate` pays the compile and
validation once and renders many times, `FakeTemplate` is the one-shot call, and
`--repeat` is exactly the case that needs the first. The pair is `regexp.MustCompile`
and `regexp.Match`, not two spellings of one result.

## The shipped data is embedded, not discovered

2026-09-02, Lilleman auf Larv.

A directory a machine happens to have would make `--seed 42` machine-dependent. Data
still lives in `data/`; `--data-path` layers over it.

`todo.md` item 182 revises it: a library imports the shipped data as Go packages.

## The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0

2026-10-04, Lilleman auf Larv; revised 2026-10-09 to put the packages in the core's Go module. Serves goals 2.3, 6.2, 7.3, 9 and 10.1. Carried out by `todo.md` items 180, 181, 182, 183, 184, 185 and 186; until then the shipped data is embedded whole, as the decision "The shipped data is embedded, not discovered" states. The CLI binary may grow to hundreds of MB.

- One Go package per locale, one per country's `geo/` tree, and one for `misc`, all in the core's Go module. A package registers nothing when imported.
- A bare `New()` loads no data and fails, naming the option to add. `WithoutShippedData` goes.
- A data package's `FS` is an `fs.FS` passed to `WithDataFS`, as anyone's data is, and `New` takes several. A module carrying functions loads through an option of its own, so a data module that starts carrying functions breaks its users. Such a function reaches the randomness through an interface the root declares; `internal/drawstate`'s type stays internal.
- A module names, in a manifest, every module it reads by default, directly or through another, so the first error names the whole set to import. A manifest is optional: a `--data-path` folder without one is a module. Nothing loads a default on its own.
- A module from anyone else names, in its manifest, the fejkdata version it was built for, and any other version than the core's fails in `New`, naming the `go get` line that aligns them. The maintainer chose exact equality on 2026-10-09, so such a module needs a new release on every fejkdata release, even when its data did not change. A shipped package's manifest names no version: it ships in the core's Go module.
- In the CLI, a flag of its own says a `--data-path` replaces what it clashes with, so flags still go anywhere on the line. Data authors and hand fixture authors bench it.
- The README asks for functions with no side effects of their own.
- The CLI is a library call. One exported list names every shipped package, and `fejkdata`'s own `main` passes it to that call. The list lives in a package of its own, so a library importing the core links no data it did not import. A custom CLI is the same few lines with other modules added, and no code exists only for custom CLIs. A README section shows it with example code, and the README's Library quick start shows the whole import block for one locale.
- Before the module API ships, a panel of the README's audience personas tries pulling in data with it.

Valid while downloading every shipped package costs a library user little, since `go get` fetches the whole Go module. Past that, the largest packages move into nested Go modules that keep their import paths, each requiring the core release that drops the package, so no build sees it twice; the `go` command refuses a Go module zip over 500 MiB (`MaxZipFile` in `golang.org/x/mod/zip`).

## With only the shipped set, a category loads on the first call reaching it; beside a `--data-path`, every category loads in `New`

2026-10-03, a restructure the maintainer approved.

Goal 13: `New` pays nothing for a category a run never reaches, which matters once the
shipped set grows to hundreds of MB. A category loads with every category it references
and its whole table family, and binds as a whole load does, so it renders the same as
after a whole load. `List` reads `shippedindex.go`, which `go generate` writes from
`data/`; `TestShippedIndexIsCurrent` fails while the index is stale. CI loads the whole
shipped set, and each shipped category alone into a fresh `New`, so none fails at first
reach.

Goal 6: beside a `--data-path`, `WithDataPath` or `WithDataFS`, every category loads in
`New`, so each mistake in the user's data is `New`'s error. That parses the whole shipped
set, against goal 13, until `todo.md` item 180 ends it.

Valid while CI proves the shipped set whole and each category alone.

Works against goal 2.3 until `todo.md` item 182 ends it.

## Every `{…}` draws afresh, and only a name keeps a pick

2026-10-05, larv-review on the change the maintainer planned on 2026-10-03.
Serves goals 5.2 and 4.1: each `{…}` makes a new pick, and values that belong together come
from the one pick a name keeps.

- `{p.a} {p.b}`, `{w} {uppercase(w)}`, `{/p.first} {/p.last}` and `{/p} {/p}` are two draws
  each, and `{p as q}{q.a} {q.b}` is one. Nothing is kept per format or per render, so the
  only load checks on draws are names, paths through names, cycles and the repeat cap.
- `{x as n}` binds a reference, or a field of the template binding it or a path into one,
  `{place as p}`: an inline template, and a choice of rows inside one category, have no reference to
  bind, and goal 4.1 holds there too.

Valid while goal 5.2 has each `{…}` make a new pick.

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
holds, and a new option, which reserves a field name. A new refusal tightens what
loads, so every new fence invalidates some file. Each such release names the rejected
spelling and its replacement in the changelog and in the load error, and that is the
whole migration: a fence rejects one spelling with one replacement, so the fix is local
to each site. A fence that would need a non-local rewrite ships a converter with its
release instead. Before `v1.0.0` a minor carries what a major would.

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

A `--data-path` category replaces a shipped one without saying so, against goal 9.3, until `todo.md` item 181 ends it.

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

2026-09-24, Lilleman auf Larv. Serves goal 6.1: a value that is always `NaN` or `Inf` is
no valid value.

`1/0` and a fixed `"0"` field are decidable, so they join the never-numeric operand as a
load error; the fold stops where an operand varies, so `a/(b*c)` with `b` fixed at `0`
and `c` varying loads and prints `Inf` every draw — catching it needs zero-absorbing
algebra for a shape nobody writes. A [typed column](../README.md#datatype) bounds its operands
instead and refuses a divisor it cannot keep from zero. In a string column, an operand that
is sometimes not a number loads, and prints `NaN` on each draw where it is not one. `Inf`
and `NaN` in a string column work against goals 6.1 and 4 until `todo.md` item 81
refuses them there too.

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
datatypes, so a tag says only what to draw: a `datatype` in it loads where it names the
one the Go type sets, and is refused where it names another. The Go type is a struct column's one datatype, so its items need
not agree on one among themselves; each must only hold that type.

## A struct's records follow Go's field access, and compile on first use

2026-09-15, Lilleman auf Larv.

The fields an embedded struct promotes are the struct's own — `e.First`, as
`encoding/json` and SQL mappers read them — so they are columns of its record and share
its names: one tag may bind `{/person as p}{p.first}` and another read `{p.sex}` (goal 4.1),
while a path tag draws afresh as `{/path}` does (goal 5.2).

A tagged field that
another field hides is refused, not dropped. A named
struct field is another entity and a record of its own. `fake:"-"` leaves any struct
field unfilled, so no name may be `-`; a pointer back to a struct
already being filled is left alone, since filling it would never end. `New` cannot see a
caller's types, so the first `FakeStruct` for a type compiles its tags and the answer,
error included, is kept per type: a test's first call is its load, and no `NewStruct`
handle is needed, as the cache already compiles once.

## A category may reference itself, and a record's fences run at load

2026-09-16, Lilleman auf Larv; revised 2026-10-07. Serves goals 5.2 and 5.7.

A reference back into a category, `{/users.first}` inside `users`, is a fresh draw, as
every `{…}` is; only a cycle is refused. A record's column fences run at `New` too, so a category that loads renders as a value and as a record alike.

## A record's column set is fixed before the first draw

2026-09-04, Lilleman auf Larv.

Only a category-level template is a record: a path descending into a field, or naming a
folder or a choice of more than one item, errors. A tail may pass through a choice whose variants carry
different fields, so the columns — and with them the CSV header written once ahead of
every row — would vary per draw. A fixed column set is what the CSV and `INSERT`
contracts rest on, so the restriction holds even where a particular choice would happen
to agree.

## Null is a `null` item, not a rate

2026-09-15, Lilleman auf Larv.

A null is one more outcome of a column's draw, so a choice's weights skew it like any
other; a null-rate option would be a second way to state odds. Outside a record column a
null renders `""`.

## A typed column holds one value, not composed text

2026-09-15, Lilleman auf Larv. Serves goal 6.1: a typed column's values are proven valid at
load, and composed text is not proven.

Its bounds come from a literal or a call's arguments, so a load error names a real
value and a range check is one comparison: `1{digits(2)}` is refused in a typed column,
where `{int(100,199)}` states its range.

## A column that only reads one reference or name is the column it reads

2026-09-15, Lilleman auf Larv.

`{/src.score}` renders exactly what `src.score` draws, and so do `{s.score}` where `s` is
bound to `/src` and `{p}` where `p` is bound to the column `place` of its own record, so it
takes that column's datatype and null. A `datatype` of its own, `string` included, wins
over the one it reads: the one way to type a column someone else wrote.

## A typed column's calc is refused unless proven

2026-09-15, Lilleman auf Larv. Serves goal 6.1, read as: a template fails to load where
the loader cannot prove it renders no forbidden value.

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

## The performance gate asserts allocations, not wall-clock time

2026-09-03, Lilleman auf Larv. Goal 13.

`AllocsPerRun` is deterministic across machines, so a 10% ceiling does not flake under
CI load, while time varies with the machine and its neighbours. A rendering slowdown
almost always costs an allocation too (a lost pre-size, a per-item map, an extra copy).
The benchmark suite (the README's Development) reports time for a human, not as a pass/fail gate.

A case guarding against one allocation per render, such as a render's scope reaching the
heap through `pickKey.under`, holds at its baseline with no margin: from a baseline of
10, a 10% margin lets that allocation through. Lilleman auf Larv decided it on 2026-10-06.

## Rows live in a TSV, the shape in JSON

2026-09-17, Lilleman auf Larv. Goals 8 and 13.

`New` allocates once per node, so a register of thirty thousand rows written as JSON
objects would cost it a second; a TSV is one allocation whose cells are substrings, and
the JSON says only how a row is composed. The TSV sits beside its category file, named
by `rows`, so a data directory stays a directory of categories, and one nothing names is
ignored.

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

## A selected row is fixed, not drawn

Decided 2026-09-27 by the maintainer, for goals 4 and 5; valid while a selector names
exactly one row.

Each selector reads the row it names: `{/misc.territory[SE].capital} / {/misc.territory[FI].capital}`
renders `Stockholm / Helsinki`, and two items of one choice may select different rows.

## `..` steps up to the parent row, and a step down after it draws afresh, or once per name's pick

2026-10-05, larv-review in PR #158. Serves goal 5.4, which has `..` go up one level, and goal
5.2, which has each `{…}` make a new pick.

- `..` names the parent table next, `city..country`, as a folder path names the folder it
  enters. `.country` after a row stays the link column's cell.
- A step down after a `..` draws afresh inside the row stepped up to (`drawSteps`). Under a
  name, the paths stepping down from one `..` share that draw through the pick's memo
  (`drawMemo.stepDownPins`), as every level a name's reads pass is shared, so two reads of
  one step down describe one row.
- A route through another row is another path to the same data, not a second spelling:
  `city[Oslo]..country.name` and `country[NO].name` both load.

## A table selects by one key column, a code, never a free-form name

Decided 2026-09-27 by the maintainer, for goal 5.5: a name may repeat across rows or change
with its register, while a code selects one row for good. Valid while
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

The code still takes a `name` option and selects by it; `todo.md` item 41 brings it to this decision.

## A parent row with no child row is a load error

2026-09-17, Lilleman auf Larv.

A descendant is drawn inside the nearest pinned ancestor, so every ancestor row must
lead to a row at every level below it, or a render could find nothing to draw. The
import script drops or fills such rows; the alternative, falling back to a free draw,
would break the consistency the link exists for without saying so.

## A table is a record of string columns

2026-09-17, Lilleman auf Larv.

Its columns are the CSV header and the `INSERT` column list, fixed by the TSV header, so
a table is a record by construction; every column is a string until a typed column
option earns its place.

## The key map is built at load, the rest on first draw

2026-09-28, Lilleman auf Larv.

A link is proved against the parent's keys and a key's uniqueness is a data mistake, so
both are load-time; the row lookup, by name and by parent, serves only a draw, so it waits
for the first one, keeping `New` linear in the bytes read. On 2026-10-05, building every
shipped lookup at load added 13 ms and 3.2 MiB to a 60 ms `New`. That cost grows with every
locale and with item 78, and most tables are never drawn by name or parent. The maintainer
kept building the lookup on first draw, once per table.

## Two categories may name one TSV

2026-09-17, Lilleman auf Larv.

Each is a view of the file with its own format and options, at the cost of holding the
rows twice, which is what a category over a register with two natural formats asks for.

## A cell may read a reference, and each row its own

2026-09-29, Lilleman auf Larv.

Goal 4: facts that belong together come from one draw, so a shop row's `phone` cell reads
`{/sv_SE.phone}` on a Swedish shop and `{/en_US.phone}` on a US one, and the table's
format, the same for every row, cannot say that. Valid while a table's rows need
generators of their own.

## A path is walked once without drawing before it is walked for real

2026-09-17, Lilleman auf Larv.

Goals 10 and 13: a path that fails moves no seeded stream, at the cost of one draw-free
walk per call, compiling its steps into a stack buffer of 16. Valid while that walk
costs little against the draw it guards.

## A country's postal codes and streets are siblings under its locality

2026-09-18, Lilleman auf Larv.

No open source pairs a Swedish street with its postnummer, and pairing the US through
its ZIPs would shape the two trees differently, so both draw inside the pinned locality
and an address agrees at that level. A street's own code is the exact pairing to add
when a source carries it.

## A locale's `address` reads its country's `geo` tree, so a locale folder is no data set on its own

2026-09-24, Lilleman auf Larv.

`data/sv_SE` alone no longer loads: a test loads `data` and prefixes the locale, and
`--no-shipped-data -d` takes the whole `data` folder or a set of one's own.

## The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more

2026-09-24, Lilleman auf Larv.

Sweden fits whole in 700 KB; every US place of 10,000 would pass a megabyte and fetch
1,200 counties of TIGER files, so the threshold sits where the two countries match in
size, and `--min-population` and `--streets-per-locality` on the import scripts build a
fuller set. Loading the whole shipped set took about 45 ms, about 20 ms of it the two
trees.

Withdrawn by the maintainer on 2026-10-03; `todo.md` item 78 replaces it.

## A locale's `address` binds its country's record once and prints it whole

2026-09-18, Lilleman auf Larv; revised 2026-10-05. Goal 4.1.

A record cannot read another whole and keep its columns, so `sv_SE.address` names the
same four columns as `geo.SE.address`. Its format binds one name,
`{/geo.SE.address as a}`, and prints it whole as `{a}`. Each column reads one part of it,
`{a.street}`, so the country's format is written once and every column describes the
address printed. Valid while a record cannot read another whole and keep its columns.

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

2026-09-18, Lilleman auf Larv. Serves goal 5.6.

A layout may carry the comma that separates arguments,
`'January 2, 2006'`. A bare `2006-01-02` reads like a third date, and a double-quoted
layout could be quotes printed in the output, so both are refused, naming the
single-quoted spelling. The layout is Go's reference time because the
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
ten among them. Agreement across a record is worth more than the last rows of a table
(goal 4.1), as the maintainer confirmed on 2026-10-03.
Where no such link can hold the fact stays a column. Layer your own `misc.territory`
over the shipped one and you must layer `misc.timezone` too, or the link fails at load
naming the row.

## A table whose register publishes no frequency draws evenly

2026-09-20, Lilleman auf Larv.

`misc.httpmethod`, `misc.port`, `misc.httpstatus`, `misc.mimetype` and `misc.tld` weigh
every row alike, so GET is a ninth of the methods drawn. Where a register publishes a
frequency it is read, as `misc.timezone` reads GeoNames populations and
`sv_SE.first-name` SCB bearers.

## `misc.timezone` weighs a zone by the people living in it

2026-09-24, Lilleman auf Larv.

The weight is the population GeoNames records in the zone's cities of 15,000 or more,
floored at 15,000, which goal 16 asks for: over 300 seeded draws of
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
longer name Copenhagen. The column cannot be a `parent`, since a table is never its own
parent; a test proves every value names a row instead. A territory the register
records no state for stands alone, which is `EH` alone, and naming one for it would be a
claim fejkdata has no business making.

## `misc.loglevel` is a table keyed by POSIX's keyword, carrying the code

2026-09-27, Lilleman auf Larv.

A flat list of names carries neither the code a PRI encodes nor a selector reaching it,
and goal 4 draws the two as one fact. The canonical spelling losing to the one its
domain writes, above, settles the rest: a configuration writes `info`, so RFC 5424's
`Informational` stays the `severity` column.

The data still keys it on the code; `todo.md` item 41 brings it to this decision.

## `misc.tld` keys carry the leading dot, where other tables key on a bare code

2026-09-20, Lilleman auf Larv.

The register spells a TLD `.se` and `misc.territory.tld` already ships it so, which a
bare key would make two spellings of one fact; `{/misc.tld}` also composes onto a host
with no separator. `misc.tld[se]` misses for it, which `todo.md` item 44 ends.

## `misc.tld` is a table of its own, and `misc.territory.tld` stays a column

2026-09-20, Lilleman auf Larv.

A `parent` demands a child for every parent row, so linking them would drop every root
zone row naming no territory, which is most of them, and goal 2 wants them shipped. The
loader refuses the link outright anyway: `tld` is a column of `misc.territory`, and a table
may not be named like a column of its ancestor.

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
`samordningsnummer` use those. They sit in a `birth-number` table under `sex`, not in a
column of it, so a pick of a sex can read the number, and `sex` keeps one shape across
locales. `todo.md` item 28 has `personnummer` read it that way, so the number agrees
with the person's sex. A samordningsnummer's day, the birthday plus 60, is
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
Goals 3 and 13: `drawSteps` draws every path from its steps, compiled when templates resolve, by
`compiledPath` or, for a caller's path, by `callerPathSteps` into a stack buffer. Both take each
step through `compileStep`, and differ only at a choice: `compiledPath` walks every variant,
`callerPathSteps` the first. On 2026-09-29 one loop over all three walks, switching on a mode field,
leaked `compiledPath`'s leaves and errors with the draw's pins, arm and memo, since Go tracks a
struct's fields as one, and a repeat of a reference path rose from 66 to 106 allocations.
`drawSteps` and `compileStep` take their state as parameters and a step holds no node, so no
benchmark gained an allocation. Valid while Go's escape analysis tracks a struct's fields
as one.

## A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read

2026-10-03, larv-review in PR #145; approved by Lilleman auf Larv. Goals 5.3 and 5.6.
`bindNames` scopes a name at compile, and a render keeps one `pickFrame` per scope. One
field binds a name where another reads it, `"first": "{/person as p}{p.first}"` beside
`"last": "{p.last}"`, so the scope cannot be the template holding the binding; the category
is the smallest unit holding every field and record column. A choice's item binds a name
for its own reads, picked anew each time the item is drawn; a read outside the item is
refused, since it would see a binding another item leaves out. A pick is drawn on
its first read, so a record's columns, rendered in name order, read one pick whichever binds
it.

The scope is lexical. A read entering a category, by a reference, through a name or from
`Fake`, sees no frame (one render of a scope's picks) opened above it, so a category renders
the same whoever references it. Such a read opens the frames around where it lands: a read through a
name from its pick's memo, so `{n}` and `{n.path}` read one pick of each name inside, and a
reference or `Fake` fresh ones.

A name is a closure. It reaches every field, choice item and repeat nested inside the
scope binding it, bar the field bound to it, and keeps one pick through the draw. A name bound inside a repeat picks
anew on each iteration. A name an enclosing scope binds may not be bound again, so no name
shadows another.
The maintainer ruled so on 2026-10-09.

A pick keeps only the levels a read of the name addresses, so a template draws the same way
under a name as anywhere else (goal 5.2). A read of a name bound to a field stays in the
category's frames, since the field sits inside it.

Valid while names are read only inside the category binding them.

## comprehension floor: every dimension and the overall at 7.0 or above, from 0.2.0

2026-10-04, Lilleman auf Larv. Serves goal 3 and applies KISS.

By the maintainer's decision of 2026-10-09, every 0.1.0 chunk ships under the floor, including the chunk that holds the release's scoring run, so that 0.1.0 can ship features. Items 170 to 175 and 177 move to 0.2.0. Item 178 plans first how item 107 regroups the engine. A paired ruling that rules worse still blocks a merge. Valid until 0.1.0 is cut.

Shipped under the floor before that by the maintainer's decisions of 2026-10-07, each with the scoring run and its ratchet waived: the CLI reading its template from stdin; and goal 5.7 loading whatever is well-formed and means one thing, dropping every refusal that guessed at a mistake.

The scoring runs are in [the comprehension history](comprehension-history.md). The nine-seat panel at depth 1, run beside the scoring run of 2026-10-09 on commit b9d1af6: Navigation 7.33, Locality 6.17, Shape 6.67, Self-sufficiency 6.83, overall 6.67. Against the panel on ad59967, its overall rose from 6.56, its Navigation from 7.17, its Locality from 6.06 and its Self-sufficiency from 6.56, and its Shape held at 6.67. Locality is lowest again, as at every panel since 2026-09-22. All 13 seats of the two runs found the named-pick keep rule hardest to follow: `keptInPick` and `readUnder`, where a key `pickKey.under` builds at render must equal one `addressedKeys` built at load, with `readField` and `readName` around them. What the two runs found is filed in `todo.md`, and `todo.md` item 177 runs both again. Valid while `AGENTS.md` gates goal 3 at 7.0.

## The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`

2026-10-05, Lilleman auf Larv. Serves goals 3.2 and 3.4. Two architect reviews found the
engine's types, `template`, `table`, `arm`, `op` and `nameBinding`, bound in one cycle
because several passes each write part of the same structs; the template language's
recursion reads few of those types. Moving the engine while several resolve passes still
wrote each value would have spread that knot across packages, behind forwarding calls from
every public method. The draw state, the grammar, `DataType`, what a proof knows of a
value, the builtins, the reading of data files and a table's rows read no engine type, so
each sits in a package under `internal/`, behind the import list `imports_test.go` checks.
A rows table holds the engine table owning it as a value of a type parameter, so
`internal/rows` imports no engine type. Valid until `todo.md` item 107 moves the engine.

## Goal 5.1's `{…}` covers `{{`, `}}` and a lone `}`, so the goal names no escape

2026-10-05, Lilleman auf Larv. Serves goal 5.1: everything outside `{…}` prints exactly as
written. Yet `{{` prints `{`, `}}` prints `}`, and a lone `}` is a load error naming `}}`, as
the README's Format string section states. The maintainer ruled the escape implicit in the
goal: doubled braces and a lone `}` belong to the `{…}` syntax, so the goal names no escape and
a lone `}` stays an error. Valid while a template escapes a brace by doubling it.

## Goal 7.2 holds while a first template can be written with no escape; a literal brace still takes `{{` or `}}`

2026-10-05, Lilleman auf Larv. Serves goal 7.2. Goals are aims: it is enough that a first
template can be written with no escape and no options, as `echo 'name: {/sv_SE.person.last}' | fejkdata`
is. A template that prints a literal brace writes `{{` or `}}`, and the goal allows that. Valid
while `{{` and `}}` are the template syntax's only escape.
