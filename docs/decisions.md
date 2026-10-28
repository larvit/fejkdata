# Decisions

## GitHub is canonical, and the module path names it

2026-09-20, Lilleman auf Larv.

Goal 1 wants the usage the tool earns, and code health expects extenders who did not write it;
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
backticks literally. Below the clause it reaches code health's reader — who opens the file —
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

2026-10-07, Lilleman auf Larv. Serves goals 6, 6.1 and 6.2.

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
against code health.

The CLI takes no argument but its flags: one is misuse naming the spelling that reads what
it likely means, `fejkdata < FILE` for a file's name, dropping it for `-`, and the template
piped in otherwise; so is a terminal with nothing piped in, which would otherwise wait for
typed input. `--list` prints paths, each read as `{/path}`. Valid while the CLI reads one
template per run.

## The CLI reads only the library's public API

2026-10-07, Lilleman auf Larv; revised 2026-10-11 to name package `cli`. Serves goal 8.6: a
CLI carrying your own modules is a few lines of Go over `cli.Main` and works just like
`fejkdata`. `cli`'s non-test code imports nothing under `internal/`, so whatever the CLI
does, a program of your own can do through the public API. Until `todo.md` item 129 moves
the lone-reference check into the library, the CLI copies how the grammar tells a lone reference, in `isPath`,
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
format string, and `echo 42` prints its newline as `echo x` does (goal 4.1).

## `FakeTemplate` and `NewTemplate` both stay

2026-09-03, Lilleman auf Larv.

They reach the same value but not at the same cost: `NewTemplate` pays the compile and
validation once and renders many times, `FakeTemplate` is the one-shot call, and
`--repeat` is exactly the case that needs the first. The pair is `regexp.MustCompile`
and `regexp.Match`, not two spellings of one result.

## The shipped data is embedded, not discovered

2026-09-02, Lilleman auf Larv. Serves goal 2: nothing is fetched at run time.

A directory a machine happens to have would make `--seed 42` machine-dependent;
`--data-path` layers over the embedded packages.

## The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0

2026-10-04, Lilleman auf Larv; revised 2026-10-09 to put the packages in the core's Go module. Serves goals 2.2, 5.2, 6.3, 8 and 9.1. Carried out by `todo.md` items 181, 186 and 196, all in 0.3.0. The CLI binary may grow to hundreds of MB.

- One Go package per locale, one per country's `geo` tree, and one for `misc`, all in the core's Go module. A package registers nothing when imported.
- A bare `New()` loads no data and fails, naming the option to add. `WithoutShippedData` goes.
- A data package's `FS` is an `fs.FS` passed to `WithDataFS`, as anyone's data is, and `New` takes several. It is an exported `var FS embed.FS`, the shape `//go:embed` gives; a function returning it would only stop a program reassigning the FS variable of a package it imports. A module carrying functions loads through an option of its own, so a data module that starts carrying functions breaks its users. Such a function reaches the randomness through an interface the root declares; `internal/drawstate`'s type stays internal.
- A module names, in a manifest, every module it reads by default, directly or through another, so the first error names the whole set to import; `todo.md` item 196 carries this out. A manifest is optional: a `--data-path` folder without one is a module. Nothing loads a default on its own.
- A manifest names no fejkdata version, and `New` checks none: the shipped packages ship in the core's Go module, and a module from anyone else keeps loading across a minor or a patch, as the README's Versioning promises. The maintainer chose so on 2026-10-09.
- In the CLI, a flag of its own says a `--data-path` replaces what it clashes with, so flags still go anywhere on the line. Data authors and hand fixture authors bench it.
- The README asks for functions with no side effects of their own.
- The CLI is a library call. One exported list names every shipped package, and `fejkdata`'s own `main` passes it to that call. The list lives in a package of its own, so a library importing the core links no data it did not import. A custom CLI is the same few lines with other modules added. No API and no flag exists only for custom CLIs; `Main` tells one apart only by its build info and file name, so its help and errors carry its name and its `--version` its own module's version beside fejkdata's. The maintainer chose so on 2026-10-11.
- The help's examples and its hint for an empty stdin read no data, so each renders in any custom CLI, and both point to `--list` for the paths. "Shipped data", in `--no-shipped-data` and in `-d`'s help, names the modules passed to `Main`, a custom CLI's own included. The audience bench's custom-CLI builder seat ruled so on 2026-10-11, before 0.1.0.

Valid while the shipped data stays under 100 MiB, which `TestShippedDataStaysUnder100MiB` checks: `go get` fetches the whole Go module, and the `go` command refuses a Go module zip over 500 MiB (`MaxZipFile` in `golang.org/x/mod/zip`). Past it, `todo.md` item 193 moves the largest packages into nested Go modules that keep their import paths, each requiring the core release that drops the package, so no build sees it twice.

## A shipped package's folder spells its namespace, its Go name is that namespace in lower case without separators, and its manifest mounts it with `namespace`

2026-10-11, Lilleman auf Larv, from two benches of a Go test author, a data author and a
custom-CLI builder. Serves goal 1 and code health. `sv_SE` ships in `data/sv_SE` as package
`svse`, `en_US` in `data/en_US` as `enus`, `geo.SE` in `data/geo/SE` as `geose`, `geo.US` in
`data/geo/US` as `geous`, and `misc` in `data/misc`: Go linters flag a capital or an
underscore in a package name, and `SE.FS` read as a type, while a folder spelling its
namespace takes a reader from a dot path in an error or a template to its file. The second
bench chose that over flat folders such as `data/svse`, 2 to 1, at the cost of a package
name that differs from its folder. A locale is one folder, never a language folder holding
regions: `en_US` and `en_GB` are different languages that share a name. Each package
holds its files at its root, and its manifest's `namespace` puts them under the namespace,
so no path repeats it. Valid until a locale's lowered name clashes with another's.

## A source whose manifest carries an index loads each category on first reach; any other loads in `New`, with what it reads

2026-10-03, a restructure the maintainer approved; widened to any source on 2026-10-09.

Goal 12: `New` pays nothing for a category a run never reaches. A manifest,
`.fejkdata.json` at a source's root, may carry an index: each category's parent table and
the paths `List` advertises below it. A category of an indexed source loads with every
category it references and its whole table family, and binds as a whole load does, so
it renders the same as after a whole load. `go generate` writes each shipped
module's manifest; `TestShippedManifestsAreCurrent` fails while one is stale. CI loads
the whole shipped set, and each shipped category alone into a fresh `New`, so none fails
at first reach.

Goal 5: a source without an index loads whole in `New`, with every indexed category it
reads, so each mistake in it is `New`'s error. Goal 5.3: an indexed source from anyone
else may fail at first reach; the call then returns the load's error, matching
`ErrLoad`, the same every call. An indexed category reading one that another source
replaced fails the same way, at its first reach, when the replacement breaks it. The
maintainer chose that on 2026-10-10 over loading in `New` every indexed category
depending on a replaced one: three rounds of the comprehension panel ruled that rule
harder to follow (code health), and it needed every entry to list what its templates read.

A first reach checks only a table's `parent` against its entry, so a table loads with its
parent and a failed load puts back whole. Proving the rest of an index is its author's,
with the tool `todo.md` item 194 plans.

Valid while CI proves the shipped set whole and each category alone.

## Every `{…}` draws afresh, and only a name keeps a pick

2026-10-05, larv-review on the change the maintainer planned on 2026-10-03.
Serves goals 4.2 and 3.1: each `{…}` makes a new pick, and values that belong together come
from the one pick a name keeps.

- `{p.a} {p.b}`, `{w} {uppercase(w)}`, `{/p.first} {/p.last}` and `{/p} {/p}` are two draws
  each, and `{p as q}{q.a} {q.b}` is one. Nothing is kept per format or per render, so the
  only load checks on draws are names, paths through names, cycles and the repeat cap.
- `{x as n}` binds a reference, or a field of the template binding it or a path into one,
  `{place as p}`: an inline template, and a choice of rows inside one category, have no reference to
  bind, and goal 3.1 holds there too.

Valid while goal 4.2 has each `{…}` make a new pick.

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

Goal 5. A script branches on the exit code and reads the named path or spelling, so
those hold; wording improves in a minor.

## A load reports every mistake at once

2026-09-30, Lilleman auf Larv. Goal 5. An author learns everything wrong with the data
in one load, not one mistake per run: `New` and `NewTemplate` report every mistake in one
error, ordered by the path each names. A mistake following only from another, such as a
reference into a category that did not compile, is left out, so each one names something
to fix. Valid while a load checks the whole set before its first render.

## Formats are checked after all data is loaded, so their errors name the category

2026-09-28, larv-review; approved 2026-09-28 by lilleman. Serves code health. Loading reads every
file first, then connects references across them. A format holding a reference can only
be checked once it is connected, so every format is checked then, and all checks run at
one moment. An error from those checks names the category path you would type (`sub.x`),
or a table cell's category and line, not the file. Valid while checking a format needs
its references connected.

## Raising the lowest supported Go is a major

2026-09-16, Lilleman auf Larv.

A consumer building on it breaks, which is the one test every rule above applies; Go's
convention of a minor is not followed. The major that raises it edits each place that
spells it: `go.mod`'s `go` line, the Dockerfile's `LOWEST_GO_VERSION`, the
`matchCurrentValue` of `renovate.json`'s rule that keeps Renovate off that ARG, and the
README's `go get` and Development lines. The `lowest` stage fails while its Go's minor is
not `go.mod`'s `go` line, so a Renovate update that a stale `matchCurrentValue` lets
through fails the gate. Serves the technical principles' semver, which counts a dropped
runtime version as a change to what exists. Valid while fejkdata follows semver for the Go
versions a consumer builds it with.

## The format check runs on the latest Go only

2026-09-21, Lilleman auf Larv.

`gofmt`'s output is the toolchain's, not the code's: 1.27 stopped padding a map
literal's values out to a lone long key, so no source satisfies both it and 1.22's. A
consumer on the lowest supported Go depends on the code compiling and its tests passing
there, which is the `lowest` stage, and never runs `gofmt` over this source, so the
latest toolchain alone defines the one canonical form a reader meets.
Valid while the lowest supported Go is not the latest.

## The changelog heading is the one spelling of a release; CI cuts the tag

2026-09-16, Lilleman auf Larv.

A tag pushed by hand is served by `go get` at once, so a tag whose commit lacks its
heading is burnt, not fixed. The heading on a gate-passed `main` commit is the trigger
instead: the tag can land only there, save a patch CI cuts, whose tag lands on a commit
that passed its own tree's `docker build .` (decision "CI releases each update to what a
release carries at once, as a fejkdata patch cut from the last release's tag"). The GitHub
release the same job publishes keeps one text as its body and is where prebuilt
binaries attach.

## A `--data-path` override rebinds every reference to the category it replaces

2026-09-24, Lilleman auf Larv.

References bind against the merged tree, so once shipped data uses `{.person}`, a
consumer's `sv_SE/person.json` is what every shipped reference into `person` reads, and
`New` fails on shipped data the consumer never wrote when that file lacks a field those
references read. Accepted: overriding is the point of layering, the error names the
reference and the field, and the fix is the consumer's file carrying the fields the
shipped tree reads.

A `--data-path` category replaces a shipped one without saying so, against goal 8.3, until `todo.md` item 181 ends it.

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

2026-09-24, Lilleman auf Larv. Serves goal 5.1: a value that is always `NaN` or `Inf` is
no valid value.

`1/0` and a fixed `"0"` field are decidable, so they join the never-numeric operand as a
load error; the fold stops where an operand varies, so `a/(b*c)` with `b` fixed at `0`
and `c` varying loads and prints `Inf` every draw — catching it needs zero-absorbing
algebra for a shape nobody writes. A [typed column](../README.md#datatype) bounds its operands
instead and refuses a divisor it cannot keep from zero. In a string column, an operand that
is sometimes not a number loads, and prints `NaN` on each draw where it is not one. `Inf`
and `NaN` in a string column work against goals 5.1 and 3 until `todo.md` item 81
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
its names: one tag may bind `{/person as p}{p.first}` and another read `{p.sex}` (goal 3.1),
while a path tag draws afresh as `{/path}` does (goal 4.2).

A tagged field that
another field hides is refused, not dropped. A named
struct field is another entity and a record of its own. `fake:"-"` leaves any struct
field unfilled, so no name may be `-`; a pointer back to a struct
already being filled is left alone, since filling it would never end. `New` cannot see a
caller's types, so the first `FakeStruct` for a type compiles its tags and the answer,
error included, is kept per type: a test's first call is its load, and no `NewStruct`
handle is needed, as the cache already compiles once.

## A category may reference itself, and a record's fences run at load

2026-09-16, Lilleman auf Larv; revised 2026-10-07. Serves goals 4.2 and 4.7.

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

2026-09-15, Lilleman auf Larv. Serves goal 5.1: a typed column's values are proven valid at
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

2026-09-15, Lilleman auf Larv. Serves goal 5.1, read as: a template fails to load where
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

Code health wants a name to reach one unit: `drawMemo` is what a draw is kept in, `readValue` what
a read drew.

## One name, one meaning

Two declarations of any kind — function, method, field, type — share a name only where one
definition covers both: the same kind of value, counted or addressed the same way, so a
reader landing on the wrong one concludes nothing false, as `Fake` renders one value on
`Generator`, `Template` and `RecordTemplate`. `TestNoFunctionSpellsAMethod` keeps a bare function off a method's
name, since a search returns both and a call site shows no receiver; review judges the
rest. Serves code health: names that tell the truth, and reaching the unit behind a symptom
without asking a person. Decided 2026-09-25 by Lilleman; valid while code health counts reading
cost.

## The performance gate asserts allocations, not wall-clock time

2026-09-03, Lilleman auf Larv. Goal 12.

`AllocsPerRun` is deterministic across machines, so a 10% ceiling does not flake under
CI load, while time varies with the machine and its neighbours. A rendering slowdown
almost always costs an allocation too (a lost pre-size, a per-item map, an extra copy).
The benchmark suite (the README's Development) reports time for a human, not as a pass/fail gate.

A case guarding against one allocation per render, such as a render's scope reaching the
heap through `pickKey.under`, holds at its baseline with no margin: from a baseline of
10, a 10% margin lets that allocation through. Lilleman auf Larv decided it on 2026-10-06.

## Rows live in a TSV, the shape in JSON

2026-09-17, Lilleman auf Larv. Goals 7 and 12.

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

Decided 2026-09-27 by the maintainer, for goals 3 and 4; valid while a selector names
exactly one row.

Each selector reads the row it names: `{/misc.territory[SE].capital} / {/misc.territory[FI].capital}`
renders `Stockholm / Helsinki`, and two items of one choice may select different rows.

## `..` steps up to the parent row, and a step down after it draws afresh, or once per name's pick

2026-10-05, larv-review in PR #158. Serves goal 4.4, which has `..` go up one level, and goal
4.2, which has each `{…}` make a new pick.

- `..` names the parent table next, `city..country`, as a folder path names the folder it
  enters. `.country` after a row stays the link column's cell.
- A step down after a `..` draws afresh inside the row stepped up to (`drawSteps`). Under a
  name, the paths stepping down from one `..` share that draw through the pick's memo
  (`drawMemo.stepDownPins`), as every level a name's reads pass is shared, so two reads of
  one step down describe one row.
- A route through another row is another path to the same data, not a second spelling:
  `city[Oslo]..country.name` and `country[NO].name` both load.

## A table selects by one key column, a code, never a free-form name

Decided 2026-09-27 by the maintainer, for goal 4.5: a name may repeat across rows or change
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
locale and every place the geo tables ship, and most tables are never drawn by name or parent. The maintainer
kept building the lookup on first draw, once per table.

## Two categories may name one TSV

2026-09-17, Lilleman auf Larv.

Each is a view of the file with its own format and options, at the cost of holding the
rows twice, which is what a category over a register with two natural formats asks for.

## A cell may read a reference, and each row its own

2026-09-29, Lilleman auf Larv.

Goal 3: facts that belong together come from one draw, so a shop row's `phone` cell reads
`{/sv_SE.phone}` on a Swedish shop and `{/en_US.phone}` on a US one, and the table's
format, the same for every row, cannot say that. Valid while a table's rows need
generators of their own.

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

## A locale package imports its country's geo package, and its `Modules()` returns both

2026-09-24, Lilleman auf Larv; revised 2026-10-11, when the shipped data became Go packages,
from a bench of a Go test author, a data author and a custom-CLI builder, 3 to 0. Serves
goals 5.2 and 8.2. A locale's `address` reads its country's `geo` tree, so `svse` imports
`geose` and `svse.Modules()` returns both: the Go import graph states the dependency, and a
program passing `Modules()` cannot leave the places out. `svse.FS` alone loads for a stand-in
providing `geo.SE`'s paths, and without one `sv_SE.address` fails at first reach with
`no entry "geo"`; until `todo.md` item 196 names the missing module in that error, the bare
`FS` works against goal 5.2. Every package exports `Modules()`, so one spelling loads any of
them; a data package imports only the packages it reads.

## The default embed holds every Swedish postort the import can place, with all its streets, and, of the US places that hold most of a postal code's land, every state capital, every place of 100,000 or more and a population-weighted tenth, each with a quarter of its streets, at least ten

2026-10-11, Lilleman auf Larv; revised the same day, after a tenth of the streets gave
Göteborg 163 and Stockholm 37. Goals 2 and 3.2: ten streets per place put 500 Stockholm
customers on ten streets, and US places of 25,000 or more left out every town and
village. `geo-us.py` draws with Efraimidis and Spirakis's weighted sampling. It draws
only among places holding most of some ZCTA's land, since no other place can carry a
postal code. The draw alone kept 310 of the 343 places of 100,000 or more and left out
13 state capitals, so a test author pinning a city they know hit `no row of … has key
or name`; every capital and every place of 100,000 or more ships beside the draw. The
draw's smallest place holds 93 people. The capitals are each state's `capital` (P36) in
[Wikidata](https://www.wikidata.org/), matched to a Census place by its GNIS id and
else by a name no other place in its state carries. No census-designated place ships,
so Honolulu is the one capital missing. Every Swedish street is 174k rows; a quarter
of the US streets 475k. On 2026-10-11 the embed grew from 5.2 MB to 16 MB, and a CLI
run printing the first `en_US.address` took 145 ms, `sv_SE.address` 95 ms and
`misc.territory` 3 ms, within goal 12. `--place-share`, `--street-share` and
`--min-streets` on the import scripts build another set. Valid while the first reach of
a country stays unnoticed beside a test run.

## A locale's `address` binds its country's record once and prints it whole

2026-09-18, Lilleman auf Larv; revised 2026-10-05. Goal 3.1.

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
ship, and ships only when that place does, so a place whose every code lies mostly in a
bigger neighbour is never drawn; counting the land outside every place too
would drop a quarter of the places, whose codes straddle unincorporated land, for a
postal city the USPS mostly names the same way.

## `--list` stays a plain list of paths

2026-09-18, Lilleman auf Larv.

It is what a script reads, so every line has to be a path that `Fake` takes; a marker
for the tables a `[selector]` follows, or a legend above them, would make the output
something to parse before use. `--help` names the selector spelling instead, and the
Table section teaches it.

## A layout is always quoted

2026-09-18, Lilleman auf Larv. Serves goal 4.6.

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
what a title needs, so `en_US.title` links to `sex` as `first-name` does. Swedish and Norwegian have no
everyday sexed honorific, so `sv_SE.title` and `nb_NO.title` are tables as well but carry no `parent`.
A title table's weight column is `share`, not the `count` a name table carries, because the values
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
(goal 3.1), as the maintainer confirmed on 2026-10-03.
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
floored at 15,000, which goal 3.3 asks for: over 300 seeded draws of
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
and goal 3 draws the two as one fact. The canonical spelling losing to the one its
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

## The Norwegian id draws Skatteetaten's synthetic series

2026-10-22, Lilleman auf Larv.

A valid fødselsnummer over a random birth date may be a living person's, and Skatteetaten
publishes no unassigned range of the real series. Its synthetic register adds 80 to the month
(folkeregisteret-api-dokumentasjon, "Ny versjon av syntetisk folkeregister"), so
`nb_NO.fodselsnummer` draws those, and a checker that requires a month of 1 to 12 refuses it.
The check digits, the individ ranges per century and the parity of the third individ digit are
Skatteetaten's, from its validation page, Folkeregisterhandboken § 2-2 and
folkeregisterforskriften.

A pair of individ digits whose check digit would be 10 is no number, and a template can't redraw
it. So `{fnr(x)}` picks the second individ digit, and data draws the third: `nb_NO.individ` holds
each sex's five digits as rows under `sex`, so a bound sex decides the parity (goal 7). Each check
fails for at most one of the ten second digits, so any third digit completes any date. A second
digit drawn by data couldn't promise that.

An `x` that is not one digit renders nothing, a visibly short number. Goal 4.7 loads a template
whose part does nothing, and a typo such as `{fnr(s.name)}` shows at the first render.

The date is a row of `nb_NO.birth-day`: 365 days with the month already plus 80, since no layout
adds 80 to a month. 29 February is left out, since the year is drawn apart. A D-number, whose day
is the birthday plus 40, isn't shipped.

Fresh-context readers seated as the README's hand fixture, Go test, validator-facing and bulk
fixture authors picked the spellings. The first three were asked for the path, and two chose the
ASCII `fodselsnummer`, which python-stdnum spells too; the ø stays in prose. Three of the four
chose `fnr` for the function, since Norwegians write it, and the bulk fixture author chose `birth`
for the field. The validator-facing author would have named the synthetic series in the path;
Skatteverket's test series sets the precedent of a plain name with the series documented.

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
Code health and goal 12: `drawSteps` draws every path from its steps, compiled when templates resolve, by
`compiledPath` or, for a caller's path, by `callerPathSteps` into a stack buffer. Both take each
step through `compileStep`, and differ only at a choice: `compiledPath` walks every variant,
`callerPathSteps` the first. On 2026-09-29 one loop over all three walks, switching on a mode field,
leaked `compiledPath`'s leaves and errors with the draw's pins, arm and memo, since Go tracks a
struct's fields as one, and a repeat of a reference path rose from 66 to 106 allocations.
`drawSteps` and `compileStep` take their state as parameters and a step holds no node, so no
benchmark gained an allocation. Valid while Go's escape analysis tracks a struct's fields
as one.

## A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read

2026-10-03, larv-review in PR #145; approved by Lilleman auf Larv. Goals 4.3 and 4.6.
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
under a name as anywhere else (goal 4.2). A read of a name bound to a field stays in the
category's frames, since the field sits inside it.

Valid while names are read only inside the category binding them.

## comprehension floor: every dimension and the overall at 7.0 or above is the target, and blocks no merge

2026-10-04, Lilleman auf Larv. Serves code health and applies KISS.

2026-10-18, Lilleman auf Larv: the floor blocks no merge, matching `~/.claude`'s decision of 2026-10-13. In `todo.md`, code health weighs the work towards the floor against the other goals, most while the panel's overall score is under 7.0. The paragraphs below record how 0.1.0 shipped under the gate this replaces.

By the maintainer's decision of 2026-10-09, every 0.1.0 chunk ships under the floor, including the chunk that holds the release's scoring run, so that 0.1.0 can ship features. Items 170 to 175 and 177 move to 0.2.0. Item 178 plans first how item 107 regroups the engine. A paired ruling that rules worse still blocks a merge. Valid until 0.1.0 is cut.

Shipped under that decision: item 89, the last of 0.1.0, at 52e0304. Its paired ruling against 90ffcff was better. The release's scoring run read Navigation 7.00, Locality 5.75, Shape 6.38, Self-sufficiency 6.50 and overall 6.25; its seats named places items 170 to 172 and 178 hold, and `isPath`, filed as item 251.

Shipped under the floor as 0.1.0's distribution, each by the maintainer's decision: item 58, prebuilt binaries and packages, which the plan had under 0.2.0, on 2026-10-13; and the image on Docker Hub, on 2026-10-14. Neither changes Go code.

Shipped by the maintainer's decision of 2026-10-11 with its ratchet waived: item 182, the shipped data as Go packages, at a632a92. Its last paired ruling against bf1becb was same overall, Mid A better and Junior A worse, with Shape fallen, three seats against the namespace mount in `internal/datafiles/mount.go` and the `at` field of `datafiles.Source`; the scoring run that would rule it worse under the floor was not run. `todo.md` item 197 replaces that mount.

Shipped under the floor before that by the maintainer's decisions of 2026-10-07, each with the scoring run and its ratchet waived: the CLI reading its template from stdin; and goal 4.7 loading whatever is well-formed and means one thing, dropping every refusal that guessed at a mistake.

The scoring runs are in [the comprehension history](comprehension-history.md). The nine-seat panel at depth 1, run beside the scoring run of 2026-10-09 on commit b9d1af6: Navigation 7.33, Locality 6.17, Shape 6.67, Self-sufficiency 6.83, overall 6.67. Against the panel on ad59967, its overall rose from 6.56, its Navigation from 7.17, its Locality from 6.06 and its Self-sufficiency from 6.56, and its Shape held at 6.67. Locality is lowest again, as at every panel since 2026-09-22. All 13 seats of the two runs found the named-pick keep rule hardest to follow: `keptInPick` and `readUnder`, where a key `pickKey.under` builds at render must equal one `addressedKeys` built at load, with `readField` and `readName` around them. What the two runs found is filed in `todo.md`, and `todo.md` item 177 runs both again. Valid while the comprehension floor's target is 7.0.

## The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`

2026-10-05, Lilleman auf Larv. Serves code health. Two architect reviews found the
engine's types, `template`, `table`, `arm`, `op` and `nameBinding`, bound in one cycle
because several passes each write part of the same structs; the template language's
recursion reads few of those types. Moving the engine while several resolve passes still
wrote each value would have spread that knot across packages, behind forwarding calls from
every public method. The draw state, the grammar, `DataType`, what a proof knows of a
value, the builtins, the reading of data files and a table's rows read no engine type, so
each sits in a package under `internal/`, behind the import list `imports_test.go` checks.
A rows table holds the engine table owning it as a value of a type parameter, so
`internal/rows` imports no engine type. Valid until `todo.md` item 107 moves the engine.

## Goal 4.1's `{…}` covers `{{`, `}}` and a lone `}`, so the goal names no escape

2026-10-05, Lilleman auf Larv. Serves goal 4.1: everything outside `{…}` prints exactly as
written. Yet `{{` prints `{`, `}}` prints `}`, and a lone `}` is a load error naming `}}`, as
the README's Format string section states. The maintainer ruled the escape implicit in the
goal: doubled braces and a lone `}` belong to the `{…}` syntax, so the goal names no escape and
a lone `}` stays an error. Valid while a template escapes a brace by doubling it.

## Goal 6.2 holds while a first template can be written with no escape; a literal brace still takes `{{` or `}}`

2026-10-05, Lilleman auf Larv. Serves goal 6.2. Goals are aims: it is enough that a first
template can be written with no escape and no options, as `echo 'name: {/sv_SE.person.last}' | fejkdata`
is. A template that prints a literal brace writes `{{` or `}}`, and the goal allows that. Valid
while `{{` and `}}` are the template syntax's only escape.

## The locales are Ethnologue 2026's ten most-spoken languages, each in the country with the most speakers of it, the Nordic countries and Germany

2026-10-11, Lilleman auf Larv. Serves goal 2. [The Ethnologue
200](https://www.ethnologue.com/insights/ethnologue200/) for 2026 ranks languages by first-
and second-language speakers together: English, Mandarin Chinese, Hindi, Spanish, Standard
Arabic, French, Bengali, Portuguese, Indonesian and Urdu. Each ships in the country with the
most speakers of it: `en_US`, `zh_CN`, `hi_IN`, `es_MX`, `ar_EG`, `fr_FR`, `bn_BD`, `pt_BR`,
`id_ID` and `ur_PK`. The Nordic countries are `sv_SE`, `nb_NO`, `da_DK`, `fi_FI` and
`is_IS`, and the maintainer adds Germany, `de_DE`. Each locale carries goal 2.1's
categories, and its `address` reads its country's `geo/` tree by reference. Egypt and
Pakistan are not yet confirmed: `todo.md` item 209 and `todo.md` item 214 check each against
Ethnologue's country pages before it ships, and record the result here. Valid until
Ethnologue's ranking changes its ten.

## A country with no open address register covering it whole ships its places and postal codes without streets

2026-10-11, Lilleman auf Larv. Serves goal 2.2. Where no open register covers a country's
streets, the only source is OpenStreetMap. ODbL binds a database derived from it to
share-alike and attribution, and every program linking that country's geo package would
inherit both terms. So that country's `address` prints no street, and the README's Data
section says why beside each such country. This works against goal 2.1 until
`todo.md` item 218 gives those countries streets. Valid until the maintainer accepts ODbL
data, or the country opens a register.

## The apt, dnf and apk repositories are static files on GitHub Pages, holding the latest release only

2026-10-13, decided by the maintainer. Serves goal 1.1; technical principle: prefer the
vendor-neutral option (`~/.claude/principles/technical-principles.md`). The repositories
are built here, as static files any host can serve: `release-tooling/package-repos.Dockerfile`
signs them with a GPG key (apt, dnf) and an RSA key (apk). A hosted package service would
only save that tooling. GitHub already hosts the code (see "GitHub is canonical, and the
module path names it"), so Pages adds no new vendor. A Pages site holds at
most 1 GB and a release's packages take about 35 MB, so the repositories hold the latest
release only; each GitHub release still carries its own packages, and the image keeps a tag
per version. Leaving Pages costs
each Linux user an edit: their apt list, `.repo` file or apk line names `larvit.github.io`.
Valid while GitHub hosts the code.

## The image is pushed to GHCR and copied to Docker Hub, whose name the README shows

2026-10-14, decided by the maintainer. Serves goal 1.1; technical principle: prefer the
vendor-neutral option (`~/.claude/principles/technical-principles.md`). Docker Hub is
Docker's default registry, so `docker run larvit/fejkdata` works with no host in the name;
that is a capability only Docker Hub has. GHCR stays the image's source, and adds no
vendor while GitHub hosts the code; being OCI, the image moves to another registry by a
change of path. GoReleaser pushes there, and a job of its own copies the image to Docker
Hub, so a Docker Hub failure holds up no other channel. GHCR also serves CI on shared
runners, which Docker Hub's per-IP limit on anonymous pulls blocks. Leaving Docker Hub
costs each user an edit of every `docker run` line and pinned script that names the bare
`larvit/fejkdata`.

## A known vulnerability in the Go version or a required module blocks a release, even in code fejkdata never calls, and blocks no merge

2026-10-14, decided by the maintainer after Docker Hub flagged v0.1.0's image. Serves goal
1.1. `govulncheck -scan module` fails on any known vulnerability in the Go version the
gate builds with or in a module fejkdata requires, even one fejkdata never calls: the
release binaries carry that Go version, and image scanners flag it whatever the code
reaches. It runs on every pull request and push and daily, as a job the release job needs
but branch protection does not require. So a vulnerability published in Go, or an outage
of vuln.go.dev, holds the release and leaves unrelated pull requests mergeable. On `main`,
each module's new vulnerabilities are filed as one `security` item in `todo.md`, under the
patch after the last release, even where only a newer Go minor fixes them: CI releases a Go
minor as a patch, so the fix ships in the next patch CI cuts, unless that patch waits for the
maintainer for a reason the decision "CI releases each update to what a release carries at
once, as a fejkdata patch cut from the last release's tag" lists. The maintainer chose so on
2026-10-28. An item whose vulnerabilities are gone moves to `CHANGELOG.md` once the report
is clean, so the changelog says what was fixed. When the report comes back clean before CI
cuts the patch that fixes those vulnerabilities, as after a failed release job, the item's
line leaves Unreleased once that patch's section reaches `main`'s changelog.
`vulnerability-filing/file_vulnerabilities.py` writes the item in `todo.md`'s shape with a
Score of 0.0, since CI cannot reach `rank_todo.py`, which scores it on the next commit to
the plan; a test files one into the repository's own `todo.md`.
`release-tooling/patch_release.py` records the fixed item's line in `CHANGELOG.md`. A pull
request from `security-plan`, rebuilt from `main` on every run, carries both to `main` and
merges itself once the tests pass. Where no Go version fixes a vulnerability yet,
govulncheck has no ignore list: the maintainer drops the job from the release's `needs`
until one does, and records why here. The paired ruling against f90b4d6 found the project
reads worse for it: the gate now reads `todo.md`, and a release waits on vuln.go.dev. The
maintainer merged it anyway, as the cost of filing into the plan, and `todo.md` item 350
ends that cost. Valid while the binaries are built from the gate's Go version.

## CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag

2026-10-14, decided by the maintainer; widened 2026-10-18 from Go patches to Go minors,
modules and GoReleaser, since no patch should wait on a human. Serves goal 1.1.

What a release carries is its Go, the modules `go.mod` requires and the GoReleaser
`packages.yml` names; the Packages workflow builds every release with its own commit's
GoReleaser, which for a patch is `main`'s. A bump that passes the gate unchanged changes
no row of the README's Versioning table, so it is a patch. Renovate merges one to `main`
once `gate` passes, save a GoReleaser major, which waits for the maintainer, who
chose so on 2026-10-28: it may rename the release assets a consumer's script downloads,
which the gate doesn't check. A major that moves the built binary's path fails the
Packages dry run. When `main` then carries a newer Go, newer modules or another GoReleaser
than the last release, the release job cuts a patch through
`release-tooling/patch_release.py cut`. It puts `main`'s Go and modules into the last
release's tree with `main`'s `.github/` (the forge refuses the workflow token a branch
whose workflows differ from `main`'s) and a changelog section, scanning that tree with
govulncheck before and after. The section names what changed, the vulnerabilities the second
scan no longer finds, and each requirement a new Go minor's release notes set on a system
`goreleaser.yaml` builds for. The notes are Go's own, from the golang/website repository.
The job builds the tree through its own `docker build .`, which builds on the latest
supported Go, and on the lowest where the tree's Dockerfile has the `lowest` stage, which a
release before 0.2.0 lacks, then commits it to a branch named for the new version, cut
from the last release's tag. It publishes the commit with
`release-tooling/publish_release.py`, whose burnt-tag check refuses a version already
tagged elsewhere, such as a planned patch number CI has taken. So a patch ships only the
bump, never the work merged since, save the packaging that `main`'s `goreleaser.yaml`
gives every release it packages. The Packages dry run doesn't run on the patch: `main`'s
gate ran it with the GoReleaser and config that package it.

A patch waits for a human, cutting nothing, and the job exits 2 until the maintainer cuts
a release or fixes the cause, while:

- the patched tree still carries a known vulnerability or fails `docker build .`, such as
  after `main`'s code was fixed for a Go minor;
- `main`'s `goreleaser.yaml`, which packages the patch, builds for other targets than the
  release's tree names, a change of supported systems (a tree before 0.2.0 names none);
- a Go minor's release notes hold a sentence about a built-for system in a shape the
  script can't read, so the changelog would miss it;
- `main`'s `go.mod` changes more than module versions, which holds back its modules: a
  module required directly that the release doesn't, which is work merged since, or
  another line, such as a `go` line raising the lowest supported Go, which is a major.
  While nothing else is due, the job scans the last release's tree and exits 2 only where
  it carries a known vulnerability, since a fix on `main` doesn't reach its users.

An advisory published between the release job's scan of `main` and its scan of the tree
fails one release job, and the next run's scan of `main` files it.

The `plan` job runs after a failed release job too, so fixed items still move. It writes
under Unreleased's Breaking each requirement that a Go minor `main` carried since the last
release sets on a built-for system, so the release that first ships the minor names it,
whether CI cuts it as a patch or the maintainer cuts it. A release section copied into
`main` removes from Unreleased each Breaking line and each fix that the section already
holds. Where the job can't read the notes or `goreleaser.yaml`'s systems, it records no
such line and still moves fixed items, since the release job fails on the same input.

A patch built with a new Go minor may drop a system the binaries ran on, such as an older
macOS. That departs from the technical principles' semver, read strictly, which counts a
dropped runtime version as a change to what exists. The maintainer, 2026-10-26, accepted
it until 1.0.0, for goal 1.1, and the README's Versioning says so. From 1.0.0 dropping a
system is a major; `todo.md` item 348 decides how a Go minor that drops one ships then.

The job compares with the last release, not with the push, and runs one at a time, so a
rerun cuts nothing twice and a missed run is caught up daily; publishing `main`'s heading
stays on a push, which the gate passed. `patch_release.py` names the last release for
every job, as the forge's latest release, and owns `CHANGELOG.md`: each run copies into
`main`'s changelog any section the last release has and `main` lacks, so a lost run loses
no section, and moves rows planned under a patch CI took to the next patch. The Packages
workflow builds a tag with the Go its own Dockerfile names. The paired ruling against
5aca8a1 found the project reads worse for the Go patches: two linked scripts, four linked
jobs and several writers of `CHANGELOG.md`. The maintainer merged it anyway, as the cost
of releasing Go patches without waiting; `todo.md` item 350 brings those links back down.
Valid while a release carries only its Go, the modules `go.mod` requires and the GoReleaser
that packages it.

## A `todo.md` item names goal 1 in its Moves when it makes fejkdata better for the people using it, not only when it wins more of them

2026-10-18, Lilleman auf Larv. Serves goal 1.

Goal 1 asks for the best fake-data tool there is, and the one developers reach for. Before
this, only items about adoption named it, such as the quick starts and the install
channels. Goal 1 weighs the most, so the data, the errors, the template features and the
docs ranked as if they did nothing for it. Now an item names goal 1 by 1 to 3, beside its
more specific goals, when its user sees the change: a category or a locale, a clearer
error, a template feature or a doc. A refactor, a test or a process change moves code health or
another goal, and names goal 1 only if its user sees a change. The maintainer noted that
goal 1 may then be too broad. Valid while goal 1 asks for the best fake-data tool there is.

## A `slugify` transform makes emails from compound names valid, and the shipped data keeps the compounds

2026-10-21, Lilleman auf Larv. Serves goal 3.

A compound name such as `Anna Lena` or `von Walden` keeps its space in the email that the
quick starts build, which makes the address invalid. A transform fixes the address where it is
built.

`{slugify(x)}` applies four steps, in order:

1. Fold each letter by the table `ascii` uses: a letter with a diacritic becomes its base
   letter (`ö` to `o`, `å` to `a`, `à` to `a`), and a ligature its letters (`ß` to `ss`, `æ`
   to `ae`). Where `ascii` drops a non-ASCII rune with no entry, this step changes it to `-`.
2. Lowercase what is left.
3. Keep `a` to `z` and `0` to `9`; change every other character to `-`.
4. Collapse each run of `-` into one, and trim `-` from both ends.

So `Anna Lena` slugs to `anna-lena`, `Åsa Öberg` to `asa-oberg`, `O'Brien` to `o-brien`,
`Jean–Luc` (with an en dash) to `jean-luc`, and `日本` to the empty string. A slug keeps no
`.`, so the only `.` in an address is the one joining two slugs.

python-slugify replaces each run of characters outside `a` to `z`, `0` to `9` and `-` with a
hyphen, lowercases the text, and strips hyphens from both ends
([python-slugify](https://github.com/un33k/python-slugify/blob/master/slugify/slugify.py)).
Rails' `parameterize` replaces each character outside those with its separator and downcases
([Rails](https://api.rubyonrails.org/classes/ActiveSupport/Inflector.html)). Django's `slugify`
removes punctuation instead and turns whitespace into a dash
([Django](https://docs.djangoproject.com/en/stable/ref/utils/#django.utils.text.slugify)).
cocur/slugify spells `ä` as `ae` by default
([cocur/slugify](https://github.com/cocur/slugify)); this decision takes the base letter, as
`ascii` does. All 193 entries of `ascii`'s table match the default transliteration table of
Rails' I18n gem
([ruby-i18n](https://github.com/ruby-i18n/i18n/blob/master/lib/i18n/backend/transliterator.rb)).

Every character a slug holds is unreserved under
[RFC 3986 §2.3](https://www.rfc-editor.org/rfc/rfc3986#section-2.3), so it is safe in a URL,
and atext under [RFC 5322 §3.2.3](https://www.rfc-editor.org/rfc/rfc5322#section-3.2.3), so it
is allowed in an email's local part. A local part may be a dot-atom
([RFC 5322 §3.4.1](https://www.rfc-editor.org/rfc/rfc5322#section-3.4.1)). Then two non-empty
slugs joined by `.` form a valid local part.

A letter outside the table splits the name at that letter; a name whose letters are all
outside the table slugs to the empty string. A locale whose names need such letters
romanizes them in its data, since `slugify` uses `ascii`'s table and no table of its own.
`todo.md` item 345 asks whether the scripts of the planned locales change that.

Valid while every non-ASCII letter of a shipped first or last name has an entry in `ascii`'s
table, which `TestAsciiFoldsEveryShippedNameLetter` checks, and RFC 5322 keeps letters, digits and
`-` in atext.

## `ascii` folds the Latin letters of ruby-i18n's default table except `ŉ`, so names such as `Wałęsa` keep their letters

2026-10-21, Lilleman auf Larv. Serves goal 3.3.

`ascii` dropped a letter it had no entry for: `Wałęsa` became `Wasa`. A name that loses letters
does not look real, and goal 3.3 asks that values look real. The table took 129 more entries,
so `ascii` and `slugify` keep the letters they dropped. The entries are the letters beyond
Latin-1 in ruby-i18n's default transliteration table
([ruby-i18n](https://github.com/ruby-i18n/i18n/blob/master/lib/i18n/backend/transliterator.rb)),
from Latin Extended-A, part of Latin Extended-B, and `ẞ`, the capital sharp s.

An entry maps a letter to letters only, so `ŉ` stays out: the table maps it to `'n`, which
carries an apostrophe.

The change alters what `ascii` returns for those letters, so it changes what exists. Before
`v1.0.0` a minor carries that, as
[A change to what exists is a major; a minor only adds](#a-change-to-what-exists-is-a-major-a-minor-only-adds)
says. It ships in `0.2.0` and is named in the changelog
([Versioning](https://github.com/larvit/fejkdata/blob/main/README.md#versioning)).

Valid while every entry of `ascii`'s table maps a letter to letters and comes from ruby-i18n's
default table.

## A pull request that changes a workflow or what a release packages runs the Packages workflow as a dry run, and `gate` is `main`'s one required check

2026-10-28, decided by the maintainer. Serves goal 1.1 and code health.

Renovate merges a GitHub Action update itself once the gate passes, so each Action a
release runs must run on the pull request that updates it, or wait for the maintainer.
Otherwise a broken Action major first fails in a release run, after the tag is published,
and a channel the run didn't reach serves the old release. The Packages workflow therefore
runs without a tag as a dry run. That run builds everything a release builds, with
GoReleaser's `--snapshot` and throwaway signing keys, and doesn't publish anything. It
runs every Action a release runs save `actions/deploy-pages`. It doesn't read a
repository secret, so a pull request from a fork runs it too.

`packages.yml` runs the dry run in a job of its own, `packages-dry-run`, whose token can
only read, so an Action that a pull request updates can't write to the repository or to
GHCR. Its steps are those of `packages.yml`'s `packages` job, through a YAML anchor. The
job in `test.yml` that calls `packages.yml`, also named `packages-dry-run`, grants every
permission a release takes. GitHub refuses a call that grants less than any called job
requests, even a job the call skips. Each job's token gets only what that job requests.

An update the dry run can't check waits for the maintainer: an Action that publishes, one
whose output only such an Action reads, and a runner. `renovate.json` names each, and says
why.

The dry run also replaces the gate's former package-repository step, so the repository
check has one CI implementation. A pull request changing what that step watched, Go code
included, now runs the slower multi-platform dry run. The maintainer accepts that cost.

Renovate merges through GitHub's auto-merge, which waits only for required checks. A
reusable workflow's checks take its jobs' names when it runs, and the calling job's name
when it is skipped. So no one dry-run check appears on every pull request, and none can be
required. The `gate` job in `test.yml` runs on every pull request and every push to
`main`. It passes once the `changes` job, which holds the changelog check, has passed,
and the tests and the dry run have each passed or been skipped. Branch protection on
`main` requires it alone, and a push to `main` releases only once it passes. Renovate's
`platformAutomerge: false` would make a merge wait for every check, the vulnerabilities
job included, which the decision on known vulnerabilities forbids.

Valid while Renovate merges through GitHub's auto-merge, GitHub names a reusable
workflow's checks by its jobs, and every job in `packages.yml` that requests a write
permission runs only with a tag.

## CI keeps Go's build and module caches between `main`'s runs, and a pull request's build starts from them

2026-10-28, the maintainer, for code health: a gate that finishes sooner is one a
contributor waits on less before changing the code again.

The Dockerfile keeps Go's build and module caches in one cache mount, at `/cache/build`
and `/cache/mod`, the paths `compose.yaml` uses. In the `test` job of `test.yml`,
`actions/cache` and `buildkit-cache-dance` carry them between runs: a push to `main` saves
them, and a pull request restores `main`'s latest and saves nothing, so a pull request
doesn't pay for the upload. The job builds the Dockerfile's targets `latest` and
`lowest-and-tooling` on two runners, each with caches of its own, since each builds
different packages. The build runs on the builder `setup-buildx-action` made, since the
cache mount lives in that builder and not in Docker's default one. Go's caches key every
entry on its inputs, so a restored cache replays only work whose inputs are unchanged.

GoReleaser runs on the runner, outside the cache mount, so the Packages workflow keeps its
Go caches through `actions/setup-go`'s own cache. setup-go keys that cache by the runner's
image, the Go version and `go.mod`, and saves it at the end of a successful job, unless
the job restored that key. A release saves it, and a pull request's dry run restores it.
The build cache is therefore that of the first release with the key, until the Go version
or `go.mod` changes.

A pull request never runs the save and the extraction, so Renovate leaves an update to
`actions/cache`, `buildkit-cache-dance` or its `busybox` image to the maintainer: one
that broke them would fail `gate` on `main`, and hold every release.

With the caches warm, a pull request's run measured 2m28s on 2026-10-28, against 2m51s
without them. With the race tests on a runner of their own, fewer draws under the race
detector, and the dry run's caches kept by `actions/cache`, three runs measured 1m38s to
1m53s the same day.

Valid while GitHub's cache serves a pull request the caches of its base branch.
