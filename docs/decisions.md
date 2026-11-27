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
a second owner. It syncs every hour, and its description and website link name GitHub. The
copy from before the move is kept private and archived as `larvit/fejkdata-pre-github`,
for its old pull requests. Valid while the maintainer keeps a Gitea copy.

## The vocabulary sits in `internal/engine`'s package doc

2026-09-21, Lilleman auf Larv; revised 2026-11-11. A developer choosing a library, the reader
goal 1 serves, reads the root's package doc first. The glossary would fill two thirds of that
page with the engine's units, mostly unexported. The glossary names the load and render path,
which `internal/engine` and its stages hold, and an internal package's doc reaches only code
health's reader. It therefore sits above `internal/engine/doc.go`'s package clause, where `go
doc` shows it. Go doc comments have no code markup, so its names keep their backticks; that
reader takes them as code. Valid while the glossary's subject is the engine; one written to
teach the exported API belongs in the root's package doc.

## Options and fields share one namespace

2026-09-02, Lilleman auf Larv.

`format`, `weight`, `repeat`, `separator`, `datatype` and `hidden` are reserved, and so is `let`, the old
name of `hidden`; every other key is a field. Nesting fields under a key, or prefixing options, would tax every template to
guard against a misspelt option. `hidden` is the one option that holds fields, the ones that write
no column ("A `hidden` field writes no column and is left out of the listing").

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
the lone-reference check into the library, the CLI cuts the reference's `{/` and `}` itself and asks
the library's `IsLonePath` whether what is left is one path.

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

2026-10-04, Lilleman auf Larv; revised 2026-10-09 to put the packages in the core's Go module. Serves goals 2.2, 5.2, 6.3, 8 and 9.1. Carried out by `todo.md` items 181, 186 and 196, all in 0.9.0. The CLI binary may grow to hundreds of MB.

- One Go package per locale, one per country's `geo` tree, and one for `misc`, all in the core's Go module. A package registers nothing when imported.
- A bare `New()` loads no data and fails, naming the option to add. `WithoutShippedData` goes.
- A data package's `FS` is an `fs.FS` passed to `WithDataFS`, as anyone's data is, and `New` takes several. It is an exported `var FS embed.FS`, the shape `//go:embed` gives; a function returning it would only stop a program reassigning the FS variable of a package it imports. A module carrying functions loads through an option of its own, so a data module that starts carrying functions breaks its users. Such a function reaches the randomness through an interface the root declares; `internal/drawstate`'s type stays internal.
- A module names, in a manifest, every module it reads by default, directly or through another, so the first error names the whole set to import; `todo.md` item 196 carries this out. A manifest is optional: a `--data-path` folder without one is a module. Nothing loads a default on its own.
- A manifest names no fejkdata version, and `New` checks none: the shipped packages ship in the core's Go module, and a module from anyone else keeps loading across a minor or a patch, as the README's Versioning promises. The maintainer chose so on 2026-10-09.
- In the CLI, a flag of its own says a `--data-path` replaces what it clashes with, so flags still go anywhere on the line. Data authors and hand fixture authors bench it.
- The README asks for functions with no side effects of their own.
- The CLI is a library call. One exported list names every shipped package, and `fejkdata`'s own `main` passes it to that call. The list lives in a package of its own, so a library importing the core links no data it did not import. A custom CLI is the same few lines with other modules added. No API and no flag exists only for custom CLIs; `Main` tells one apart only by its build info and file name, so its help and errors carry its name and its `--version` its own module's version beside fejkdata's. The maintainer chose so on 2026-10-11.
- The help's examples and its hint for an empty stdin read no data, so each renders in any custom CLI, and both point to `--list` for the paths. Only the opening that `fejkdata`'s own `main` prints, since its data holds `sv_SE.person`, reads shipped data; a CLI another module builds on `Main` prints no opening, whatever its name. "Shipped data", in `--no-shipped-data` and in `-d`'s help, names the modules passed to `Main`, a custom CLI's own included. The audience bench's custom-CLI builder seat ruled so on 2026-10-11, before 0.1.0.

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
parent. Proving the rest of an index is its author's,
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

A heading is spelled `## Unreleased` or `## 0.5.0 - 2026-11-20`. Keep a Changelog puts
brackets around the name to link it to a comparison URL, and this changelog has no such
URLs. The tooling still reads the bracketed headings of the releases up to 0.4.0, since
packaging an old release reads its tag's changelog. The changelog check refuses a heading it
can't read, so a pull request catches one that `publish_release.py` would reject on `main`.

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

`1/0` and a read whose every value is the fixed text `0` are decidable, so they join the
operand that is never a number, fixed text such as `abc`, `NaN` or `Inf`, as a load error;
the fold stops where an operand
varies, so `a/(b*c)` with `b` fixed at `0` and `c` varying loads and prints `Inf` every
draw — catching it needs zero-absorbing algebra for a shape nobody writes. A [typed column](../README.md#datatype) bounds its operands
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

It reads through a field as well: `{p.id}` where `p` binds a record whose `id` is an integer is
an integer column, and where `p` binds a choice of templates each with an `id`, every item's `id`
must agree on datatype, or the load fails as it does for a column's own items; one may be null.

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
`Generator`, `Template` and `RecordTemplate`. Within a package, a function never shares a
method's name, since a search returns both and a call site shows no receiver. Review judges the
whole rule, and no test scans the source for it. Serves code health: names that tell the truth,
and reaching the unit behind a symptom without asking a person. Decided 2026-09-25 by Lilleman
and revised 2026-11-04; valid while code health counts reading cost.

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
- A step down after a `..` draws afresh inside the row stepped up to (`drawAcrossTables`). Under a
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

2026-09-17, Lilleman auf Larv. Serves goal 3.1: a value drawn below another belongs to it.

A descendant is drawn inside the nearest pinned ancestor, so every ancestor row must
lead to a row at every level below it, or a render could find nothing to draw. A geo
import keeps that true by dropping a locality with no street or no postal code, and
logging each one. Drawing outside the ancestor instead would break the consistency the
link exists for without saying so.

What a municipality or region left with no locality means differs by country.
`geo-dk.py`, `geo-fi.py`, `geo-no.py` and `geo-se.py` ship the whole country, so such a
row would be missing from the shipped data: they exit and name it. `geo-us.py` ships a
sample of the country's places, so a county with none of them has no place to keep: it
drops the county and counts it, and drops a state with no county and names it.

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

## A locale package imports the data packages it reads, and its `Modules()` returns them all

2026-09-24, Lilleman auf Larv; revised 2026-10-11, when the shipped data became Go packages,
from a bench of a Go test author, a data author and a custom-CLI builder, 3 to 0. Serves
goals 5.2 and 8.2. A locale's `address` reads its country's `geo` tree, so `svse` imports
`geose` and `svse.Modules()` returns both: the Go import graph states the dependency, and a
program passing `Modules()` cannot leave the places out. `svse.FS` alone loads for a stand-in
providing `geo.SE`'s paths, and without one `sv_SE.address` fails at first reach with
`no entry "geo"`; until `todo.md` item 196 names the missing module in that error, the bare
`FS` works against goal 5.2. Every package exports `Modules()`, so one spelling loads any of
them; a data package imports only the packages it reads. A locale that reads another locale of its
own country, as `nn_NO` reads `nb_NO` (revised 2026-11-14, Lilleman auf Larv), imports that
one instead, and reaches the country's `geo` package through the other's `Modules()`:
`nnno` imports `nbno` and not `geono`.

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
`misc.territory` 3 ms, then within goal 12; its 12.2 now asks for under 100 ms. `--place-share`, `--street-share` and
`--min-streets` on the import scripts build another set. Valid while the first reach of
a country stays unnoticed beside a test run; it works against goal 12.2 until `todo.md`
item 465 lands.

## A locale's `address` binds its country's record once and prints it whole

2026-09-18, Lilleman auf Larv; revised 2026-10-05. Goal 3.1.

A record cannot read another whole and keep its columns, so `sv_SE.address` names the
same four columns as `geo.SE.address`. Its format binds one name,
`{/geo.SE.address as a}`, and prints it whole as `{a}`. Each column reads one part of it,
`{a.street}`, so the country's format is written once and every column describes the
address printed. Valid while a record cannot read another whole and keep its columns.

## A postort's kommun comes from its name, its tätort, the SCB småort holding its centre or its codes, never from distance

2026-09-24, Lilleman auf Larv.

GeoNames leaves a fifth of Sweden's codes without a kommun and carries stale spellings;
the nearest code across a border named the wrong kommun half the time it was tried, so a
postort none of the four rules place is dropped, as is one not cased like a place name.

A postort that is no tätort and is not named as a kommun takes the kommun of the SCB småort holding
its centre, ahead of its codes, since SCB is the source of truth for which kommun a place lies in (the
maintainer, 2026-11-05, goal 3.2). The test is whether the småort's polygon contains the centre, so it
measures no distance. Valid while SCB publishes its småorter with their polygons.

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

## A builtin reads the date the caller gives, never the clock, so `born` bounds a date by an age

2026-09-18, Lilleman auf Larv; revised 2026-11-04.

Goal 9.1 has the caller supply everything that can change the output, so in the library a seeded fixture
changes with the day only where the caller gives it a new day. The earlier decision bounded a date by days,
and a fixed earliest day made the oldest `personnummer` wrong as years passed, against goal 3.
`born` counts an age by the year of the date, whether or not the birthday has passed, since
folkbokföringslagen 18 § changes the separator "det år en person fyller 100 år". Valid while no
builtin needs a part of the date finer than the year to count an age. `{date(from,to,…)}` keeps its
absolute days for what no age bounds.

## `New` requires `WithToday`, and `--today` defaults to the system's date

2026-11-04, Lilleman auf Larv.

Goal 9.1 has a library caller supply the date as it supplies the seed, and goal 6.3 has the CLI fill
it in. A panel of five readers seated as the README's audience chose the spelling: `--today` beside
`WithToday(time.Time)` won four to one over `--as-of` and `WithAsOf`, and all five wanted `--date` least,
since it collides with `{date()}` and `sv_SE.date`. `born` won three to two over `birthdate`, and `age`
and `turns` were wanted least. A date is required even where no template reads it, since a load cannot
promise that the data never will. Valid while goal 9.1 stands.

## `sv_SE.personnummer` and `sv_SE.personnummer-12` draw four age bands, up to age 113; `personnummer` is a table under `sex` with a column per band, and `personnummer-12` has no `ten`

2026-11-04, Lilleman auf Larv.

Goal 3 asks that a valid value pass its real check, and goal 3.2 asks for a published source. The bands
and weights are SCB's, in `docs/research/research-age-bands.md`. The law gives no oldest age, so the top
band stops at 113, the oldest validated age in Sweden: Astrid Zachrison, 113 years
(LongeviQuest; Wikipedia's list of Swedish supercentenarians). Bands relative to a date cut across
centuries, so `personnummer-12` cannot write `19{ten}` and `20{ten}`: it prints the century in `born`'s
layout, and `{luhn(9)}` keeps it out of the check digit. A panel of three readers chose `{luhn(9)}` over
`{luhn(-9)}`, `{luhnlast(9)}` and `{luhntail(9)}`. No template conditional exists to swap 238 and 239 for
980 and 981, so a date before 2013, which can draw a birth before 1900, falls outside the test series,
and so does a birth in a year Skatteverket hasn't published. Valid while the callers' dates lie from 2013
up to the years Skatteverket has published, and while no Swede passes 113.

2026-11-14: goal 3.1 has a person's sex and number come from one pick, and a category can't read a sex
its caller picked. So `sv_SE.personnummer` is a table under `sex`, one row per sex, so `sex[f].personnummer` and a bound
`{s.personnummer}` carry that sex's birth number. Each band is a column, `child`, `adult`, `senior` and
`centenarian`, and `number`, the table's format, starts with `sv_SE.personnummer-date`, the four bands
by their weights. A panel of three readers, seated as a validator-facing author, a hand fixture author
and a Go test author, chose `personnummer.child` over `personnummer[child]`, `personnummer-child` and
`personnummer.0-17`, and `{s.personnummer}` over `{s.birth-number.personnummer}` and
`personnummer-f`. A row's cells draw apart, so its CSV columns hold the sex and five numbers of that sex that share no date.

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
locales. `sv_SE.person`'s `personnummer` and `samordningsnummer` read the number through the
person's pick of a sex, so each agrees with that sex, and `sv_SE.personnummer`, a table under
`sex`, reads it in each row's cells. A
samordningsnummer's day is the birthday plus 60: `sv_SE.samordningsnummer` draws it from 61 to
88, valid in every month, and `sv_SE.person` computes it from the birth date it binds.

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

## The Danish id draws valid CPR numbers over the last hundred years

2026-11-08, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) and goal 3.1 (values that belong
together share one pick) are served. Goal 3.3 (a form that reaches nothing real) is not:
the CPR office publishes no test series and no unassigned range, and says every number it issues
is real and never reused. Digitaliseringsstyrelsen's test data is handed out on application, not
published. So `da_DK.cpr` is valid and may be a living person's, and its page under `docs/locales` says so. The
office's own "fiktive" form, 6 added to the first day digit, is used for real payroll cases and
fails every date check, so it isn't a series that reaches nothing real. A form that reaches
nothing real isn't shipped. Item 23 in `todo.md` ends the departure from goal 3.3. The maintainer
accepted shipping `da_DK.cpr` until then, over holding it back or drawing the "fiktive" form.
Valid while the CPR office publishes no test series or unassigned range.

The rules are the office's "Personnummeret i CPR-systemet" (2008): the ten digits weighted
4 3 2 7 6 5 4 3 2 1 add up to a multiple of 11, the last digit is even for a woman and odd for a
man, and the year digits with the seventh digit name the century. cpr.dk adds that numbers
failing the check have been issued since 2007. Those aren't drawn: a checker may require the
check, and none requires that it fail.

A template can't redraw a check digit of 10, and the sex's parity falls on the last digit, the
check. So `{cpr(x)}` picks the seventh digit, an eighth and a ninth digit, and the check digit,
the ninth by trial, and `x` supplies only the parity, from `da_DK.individ` under `sex`, one digit a
sex. Whatever the first eight digits, at least four of the ten ninth digits pass, so a completion
always exists.

The century can't be written in data: the seventh digit depends on the full year, and a template
prints only the year digits. So `{cpr(x)}` takes the year as the latest not after `WithToday`
that bears them, and draws the seventh digit evenly among those the office's table of century
digits gives that year, since the office publishes no frequency. The date comes from
`born(0,99,…)`, which draws from 100 calendar years, so each pair of year digits names one year, and someone of 100 or older, 0.02% of
the population, isn't drawn. The three age bands carry Statistics Denmark's FOLK1A shares for
1 January 2026, the 100 or older folded into the last band. Years outside 1858 to 2057, which the
table can't write, render nothing, a visibly short number, as in `fnr`.

`cpr` is the register's own name, and the function takes `fnr`'s shape: a derivation after the
date, with the sex's digit as its operand. The number is written `ddmmyy-ssss`, as the office
writes it.

Fresh-context readers seated as the README's hand fixture, Go test, validator-facing and bulk
fixture authors picked the spellings. All four chose `cpr` for the path and `{cpr(x)}` for the
function, since Danes write it, and all four chose `individ` over `parity` for the table, the
name `nb_NO` uses for the same job.

## The French id draws valid NIRs over the last hundred years

2026-11-19, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) is served, and goal 3.3 (a form
that reaches nothing real) is not: INSEE publishes no test series and no unassigned range for
the NIR, and the Sécurité sociale's own test data is not published. The "fictif" NIRs in the
SNDS documentation are a pattern seen in its data, not a series INSEE keeps. So `fr_FR.nir` is
valid and may be a living person's, and its page under `docs/locales` says so, as it does for `da_DK.cpr`. Its
place of birth is a department of INSEE's 101 drawn evenly and a commune code in its range
that need not exist, and Corsica is always `2A` or `2B`. Item 23 ends the departure from goal
3.3. Valid while INSEE publishes no test series or unassigned range.

## The Estonian id draws valid isikukoodid over the last hundred years

2026-11-19, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) is served, and goal 3.3 (a form
that reaches nothing real) is not: the Interior Minister's regulation "Isikukoodide moodustamise
ja andmise kord" (RT I, 08.01.2019, 11), which names EVS 585 for the structure, publishes no test
series and no unassigned range, and neither does the Population Register. So `et_EE.isikukood` is
valid and may be a living person's, and its page under `docs/locales` says so, as it does for `fr_FR.nir`. Item 23
ends the departure from goal 3.3. The regulation lists the first digits 1 to 6, so the id stops at
the 2000s, and `{isikukood(x,d)}` prints no code for a later year. EVS 585 is not public, so the
check digit follows python-stdnum's `ee.ik`, which weights 1 to 9 and 1 and then 3 to 9, 1, 2 and 3.
No source states which serial numbers the register gives, so the serial number is drawn evenly from
001 to 999, never 000. The first digit names the century of a date that follows it, so the builtin
takes the date as an operand and prints the whole code, where `hetu` and `cpr` complete one after
its date; its first operand is `et_EE.sex-digit`'s digit, 1 for a man and 2 for a woman, as
`fr_FR.sex-digit` does. The age bands weigh Statistics Estonia's population of 1 January 2026.
Valid while the regulation and the register publish no test series or unassigned range.

## The Latvian id draws valid personas kodi over the last hundred years

2026-11-19, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) is served, and goal 3.3 (a form that
reaches nothing real) is not: the Office of Citizenship and Migration Affairs, which gives the
codes, and the Population Register Act publish no test series and no unassigned range. So
`lv_LV.personas-kods` is valid and may be a living person's, and its page under `docs/locales` says so, as it does for
`et_EE.isikukood`. Item 23 ends the departure from goal 3.3. The Act gives a person registered from
1 July 2017 a code that starts with 32 and then random digits, with no date, so `{personaskods(d)}`
prints one for a birth date from 1 July 2017 and a dated code for an earlier one; the shipped bands
draw both, and some 9% of the codes drawn in 2026 are undated. The Act gives a `32` code to everyone
registered from that day, an immigrant born earlier included; the builtin decides by the birth date
alone, since it holds no registration date, so it never gives such a person one. The Office states
no check digit. python-stdnum's `lv.pvn` computes it and says no independent source confirms it,
and the State Revenue Service's formula as Wikipedia gives it, which the tests' checker applies,
yields the same digit. It is 1 plus each of the first ten digits times its weight, 10, 5, 8, 4, 2,
1, 6, 3, 7 and 9, added up, modulo 11, with 10 written as 0. The undated code takes the same digit,
since the Act says only that its digits are random. The code has no hyphen, which forms often put
after the sixth digit, since python-stdnum's `compact` strips it. No source states which serial numbers the register gives, so the serial number
is drawn evenly from 000 to 999. The builtin takes the date as an operand and prints the whole
code, as `{isikukood(x,d)}` does, since the century digit and the 32 depend on it. There is no sex
digit, since the code holds none. The age bands weigh Eurostat's population of 1 January 2025.
Valid while the Office and the Act publish no test series or unassigned range.

## The Lithuanian id draws valid asmens kodai over the last hundred years

2026-11-19, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) is served, and goal 3.3 (a form that
reaches nothing real) is not: the Population Register publishes no test series and no unassigned
range, and no page of the Centre of Registers states one. So `lt_LT.asmens-kodas` is valid and may
be a living person's, and its page under `docs/locales` says so, as it does for `et_EE.isikukood`. Item 23 ends the
departure from goal 3.3.

The standard LST 1185-91, which the Lithuanian Wikipedia names for the code's structure, is not
public. So the check digit follows python-stdnum's `lt.asmens`, which takes it from its `ee.ik`:
the weights 1 to 9 and 1 and then 3 to 9, 1, 2 and 3, with the first digits 1 to 6 for the 1800s
to the 2000s. Two forms of the code are left out. The Wikipedia says a person who doesn't know the
birth month or day has a 0 for it, and a code starting with 9 has no date and no check digit, but
python-stdnum checks the digit on such a code all the same, so no source says what a valid one is.
No source states which serial numbers the register gives, and the Wikipedia calls the serial number
the order number of those born that day, so it is drawn evenly from 001 to 999, never 000. The age
bands weigh Eurostat's population of 1 January 2025.

The first digit names the century of a date that follows it, so `{asmenskodas(x,d)}` takes the date
as an operand and prints the whole code; its first operand is `lt_LT.sex-digit`'s digit, 1 for a man
and 2 for a woman. It shares its draw with `{isikukood(x,d)}`, since the two schemes agree today,
and keeps its own name, since each name says whose code it prints. Valid while the Register
publishes no test series or unassigned range.

## The Finnish id draws DVV's unissued 900 to 999, and `henkilotunnus-issued` the issued 002 to 899 through `{hetu(x,'issued')}`

2026-11-12, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) and goal 3.3 (a form that reaches
nothing real) are both served. DVV's page "Personal identity code" says the individual numbers
issued are 002 to 899, and the test table linked from "Reform of personal identity code" holds
hypothetical codes with 902 to 904. So `fi_FI.henkilotunnus` draws 900 to 999, 50 of each
parity, and no drawn code is a living person's. The Vero API FAQ on testing (vero.fi, "Vero API: testaus, onko henkilötunnuksen muodolle jotain tiettyjä vaatimuksia") calls codes whose ending
starts with 9 synthetic ("keinohenkilötunnuksia") that may take extra checks, so a checker may
refuse them; its page under `docs/locales` says so, as it does for the fødselsnummer. This follows the Swedish and
Norwegian ids, which draw an issuer's test series, and not the Danish, whose office publishes
none.

Valid while DVV issues no number from 900.

2026-11-14: a hand fixture for a checker that refuses the synthetic endings needs the issued range, and goal 3
asks the best category to offer more than one kind. `fi_FI.henkilotunnus-issued` draws 002 to 899
through `{hetu(x,'issued')}`: the sign and the control character are the same, so a second builtin
would add a name for one meaning. A panel of three readers, seated as a validator-facing author, a
hand fixture author and a Go test author, chose `henkilotunnus-issued` over `-real`, `-valid` and
`.issued`, and `'issued'` over a number range, a bare word and `{hetu-issued(x)}`. A drawn number
may be a living person's, which the `fi_FI` page says beside it; the default stays 900 to 999 for goal
3.3.

The control character, the century signs and the parity of the individual number are DVV's. The
sign is the latest year not after today that bears the year digits, as `{cpr(x)}` takes its
century, and a year outside 1800 to 2099 renders nothing. The date is two bands, born 0 to 64 and
65 to 99, weighted by Statistics Finland's population aged 65 or over at the end of 2025, drawn
evenly within a band.

Fresh-context readers seated as the README's hand fixture, Go test, validator-facing and bulk
fixture authors picked the spellings. `henkilotunnus` won three of four for the path, and `hetu`
all four for the function, since Finns write it. The validator-facing author would have named
the series in the path, and the bulk fixture author argued against; Skatteverket's test series
sets the precedent of a plain name with the series documented.

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
Code health and goal 12: `drawAcrossTables` draws every path that may cross a table, from its
steps, compiled when templates resolve, by
`compiledPath` or, for a caller's path, by `callerPathSteps` into a stack buffer. Both take each
step through `compileStep`, and differ only at a choice: `compiledPath` walks every variant,
`callerPathSteps` the first. On 2026-09-29 one loop over all three walks, switching on a mode field,
leaked `compiledPath`'s leaves and errors with the draw's pins, arm and memo, since Go tracks a
struct's fields as one, and a repeat of a reference path rose from 66 to 106 allocations.
`drawAcrossTables` and `compileStep` take their state as parameters and a step holds no node, so no
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
its first read, so a record's columns, drawn in name order, read one pick whichever binds
it.

The scope is lexical. A read entering a category, by a reference, through a name or from
`Fake`, sees no frame (one render of a scope's picks) opened above it, so a category renders
the same whoever references it. Such a read opens the frames around where it lands: a read through a
name from its pick, so `{n}` and `{n.path}` read one pick of each name inside, and a
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

Shipped by the maintainer's decision of 2026-10-11 with its ratchet waived: item 182, the shipped data as Go packages, at a632a92. Its last paired ruling against bf1becb was same overall, Mid A better and Junior A worse, with Shape fallen; the scoring run that would rule it worse under the floor was not run.

Shipped under the floor before that by the maintainer's decisions of 2026-10-07, each with the scoring run and its ratchet waived: the CLI reading its template from stdin; and goal 4.7 loading whatever is well-formed and means one thing, dropping every refusal that guessed at a mistake.

The scoring runs are in [the comprehension history](comprehension-history.md). The nine-seat panel at depth 1, run beside the scoring run of 2026-10-09 on commit b9d1af6: Navigation 7.33, Locality 6.17, Shape 6.67, Self-sufficiency 6.83, overall 6.67. Against the panel on ad59967, its overall rose from 6.56, its Navigation from 7.17, its Locality from 6.06 and its Self-sufficiency from 6.56, and its Shape held at 6.67. Locality is lowest again, as at every panel since 2026-09-22. All 13 seats of the two runs found the named-pick keep rule hardest to follow: `keptInPick` and `readUnder`, where a key `pickKey.under` builds at render must equal one `addressedKeys` built at load, with `readField` and `readName` around them. What the two runs found is filed in `todo.md`, and the run below is the next.

The nine-seat panel at depth 1 and the four-seat scoring run of 2026-11-11 read commit 3f909f1. `larv-review needs-scoring-run` called no run on that commit, so the history doesn't hold them. The four seats: Navigation 7.13, Locality 5.63, Shape 6.50, Self-sufficiency 6.88, overall 6.38. The nine: Navigation 7.11, Locality 5.78, Shape 6.56, Self-sufficiency 6.72, overall 6.44. Against the nine-seat panel on b9d1af6 every mean fell, Navigation from 7.33, Locality from 6.17, Shape from 6.67, Self-sufficiency from 6.83 and the overall from 6.67, and Locality is lowest again. Nearly every seat found the named picks hardest: `checkNameRead`, `sharedDraw` and `refuseFreshDraw` in `fence/namefence.go`, `readName` and `enter` in `render`, `LandsInPick`, `drawSteps`, and the `LoadFinds` phase around `Finish`. Four also named `render`, `load.RenderEdges` and `fence.walkRoutes` as three lists of one node's children, kept in step by a comment. Items 421 to 430 hold what the seats named, and item 421 starts lifting Locality to the floor. Valid while the comprehension floor's target is 7.0.

## The node types sit in `internal/engine/tree`, compiling, binding, linking and resolving in `internal/engine/load`, every proof and fence of the finished tree in `internal/engine/fence`, with the name rules in its package `names`, drawing in `internal/engine/render`, and the pipeline and lazy loading in `internal/engine`, with the root the public API and its options; compile builds each node with its syntax, and `LoadFinds.Finish` writes into it what its load found, save a folder's children, which its load puts, and a table's `siblings`, which `load.LinkTable` sets; what reads no engine type, a table's rows included, sits in `internal/`; `internal/shipped` generates the files about the shipped data from a loaded tree

2026-10-05, Lilleman auf Larv; revised 2026-11-11, three times, and 2026-11-14. Serves code health. On 2026-10-09 the
maintainer chose to regroup the engine into packages, untangled first, over naming the root
files by area; this is the plan for that choice, made with the systems architect.

Why the engine stayed whole. Two architect reviews found its types, `template`, `table`,
`arm`, `op` and `nameBinding`, bound in one cycle, because several load stages each write
part of the same structs. Moving them while that holds would spread the knot across
packages, so the untangling comes first. The draw state, the grammar, `DataType`, what a
proof knows of a value, the builtins, the reading of data files and a table's rows read no
engine type, so each sits in its own package under `internal/`, the leaf libraries, behind
the import list `imports_test.go` checks. That list is a static check of the source, like the
cyclomatic one, not a test of behaviour, so the technical principle "Assert the outcome",
which bans a test that reads the source, doesn't cover it.

How load finishes a value. The node graph is cyclic by design, since a category may
reference itself, so a node must exist before its references resolve. Compile builds each
node with its syntax only. Every later load stage collects what it finds in one
`tree.LoadFinds` value, and its `Finish` writes it into the load's nodes, once; a second call
is an `invariant` panic. A node is one struct, not a load-time type beside a render-time one:
two types would double every node type for a guarantee that `tree` being the only package
to write a node's fields already gives. So a read of a template's or a name binding's load
finds before `Finish` doesn't fail to compile. It is an `invariant` panic that names the phase.
A load stage reads finds through its load's `LoadFinds`, and a stage after the load reads
the finished nodes: where a `load` helper serves both, it takes the read as a function, such as `LoadFinds.Compiled` or `Template.Compiled`. A load reads the categories finished before it and doesn't write them. A child table is therefore linked without writing its parent: the
parent's children come from its folder, so a lazy load that fails leaves every finished
table as it was. Two writers besides `Finish` remain: a load puts each folder's children as
it reaches them, and `load.LinkTable` gives each table its `siblings`, through a `tree` method.
`load.LinkTable` also links the table's rows to its parent's rows, which `internal/rows` writes
in its own type.

Packages, by the stage that writes or reads a type. `tree` holds the node types, the folder
among them, their constructors, setters, `Finish` and accessors, so the whole cycle sits in one
package; it imports only leaf libraries, as `imports_test.go` lists. `load` compiles, binds,
links and resolves; `fence` runs every proof and fence of the finished tree, the cycle and
repeat fences among them, since they check the finished graph; `render` draws. `load` and
`render` each import `tree` and leaf libraries, and neither imports the other. `fence` imports
`load` as well as `tree`, since the fences walk the graph `load` built, read its names and
compile a path as it does; an import one way keeps the stages acyclic, and `render` imports
neither. The rules that refuse a read of a name sit in `fence/names`. They share one walk of
the reads a pick renders, and no other fence reads that walk or the types it returns. `fence`
calls the package's `RefuseReads`, and the pipeline its `RefuseReadsInsideFields`.
`internal/engine` holds the pipeline and lazy loading, and the entry points' walks of a
caller's path and record, importing `tree`, `load`, `fence`, `fence/names`, `render` and the
leaf libraries.
The root keeps only the public API and its options: it imports `internal/engine`, `tree`, whose
folder its generator holds, and leaf libraries, never a stage, and builds `Record` from
the column values the engine draws. Names and paths get no package across the stages: each stage keeps
its part of them, as `fence` keeps its name rules in `fence/names`, and `tree` holds the step type every stage walks. The three functions
walking a path, one to load every branch, one to draw one and one to find the caller's
folders, stay in their stages.

`internal/shipped` sits beside the root, above the engine: it generates the repository's files
about the shipped data, the manifests, the listing and the two pins, from a tree
`internal/engine` loaded, and no package of the library imports it. Its
command, `internal/shipped/generate`, is the one writer of those files, and the root's tests
compare each file with the bytes the package returns.

The allocation gate holds because the hot path stays inside `render`, where `renderEnv`'s
rule of passing by value applies unchanged. `render` reads `tree` through concrete pointer
methods, never an interface, and `tree.TokenDraw` is a closed sum that `render`
switches on. `perf_test.go` stays in the root and uses only the public API; a cut that raises
an allocation count is fixed, not re-pinned.

Each cut merged only where its paired ruling read no worse. `internal/` holds more
directories than the principle's about seven per level, since a reader finds the leaf libraries
by name. Valid while the root holds no stage.

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

## Locales ship by region: the Nordic countries, then the rest of the EU, then the rest of Europe, then the rest of the world

2026-10-11, Lilleman auf Larv; revised 2026-10-31 to order locales by region, in place of
Ethnologue's ten most-spoken languages. Serves goal 2.1.

The Nordic countries are `sv_SE`, `nb_NO` with `nn_NO`, `da_DK`, `fi_FI` and `is_IS`. The
rest of the EU filed so far is `de_DE` and `fr_FR`. The rest of the world filed so far comes
from the old ranking, [the Ethnologue
200](https://www.ethnologue.com/insights/ethnologue200/) for 2026, each language in the
country with the most speakers of it: `en_US`, `zh_CN`, `hi_IN`, `es_MX`, `ar_EG`, `bn_BD`,
`pt_BR`, `id_ID` and `ur_PK`. Egypt and Pakistan are not yet confirmed: `todo.md` item 209
and `todo.md` item 214 check each against Ethnologue's country pages before it ships, and
record the result here. A region's locales ship before the next region's; within a region,
the plan's scores order them. Each locale carries goal 2.1's categories, and its `address`
reads its country's `geo/` tree by reference.

## A country with no open address register covering it whole draws invented streets

2026-11-14, decided by the maintainer. Serves goal 2.1. Where no open register covers a
country's streets, the only source is OpenStreetMap. ODbL binds a database derived from it
to share-alike and attribution, and every program linking that country's geo package would
inherit both terms. So that country's `address` draws an invented street: names from a
language model's recall, each weighing the same, as the project principle for `da_DK.company`
states. A half address helps no one, and an invented street is better than none.
The `geo` page's section for each such country says that its streets are invented.
`data-import/geo.py`'s `invented_streets` builds the fallback, and `geo.DE` is its first user. Goal 3 allows invented values but ranks them after
real ones, so a country's streets offer one kind until `todo.md` item 218 gives it its real
streets. Valid until `todo.md` item 218 ships: the decision "The project takes open licences, share-alike ones included" accepts ODbL data.

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
reaches. It runs on every push to `main` and daily, as the first steps of the release job,
which branch protection does not require. Since 2026-11-19 a pull request doesn't run the
scan: the scan blocked no merge, and `main`'s next run files what it would have found. So a vulnerability published in Go, or an outage
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
`release-tooling/file_vulnerabilities.py` writes the item in `todo.md`'s shape with a
Score of 0.0, since CI cannot reach `rank_todo.py`, which scores it on the next commit to
the plan. No test reads the repository's own `todo.md`, so an edit to the plan can't fail
the gate: where the plan's layout drifts, the filer exits 2 naming what it misses, or
`rank_todo.py` refuses the row it filed.
`release-tooling/changelog.py` words the fixed item's line, and `patch_release.py plan`
records it in `CHANGELOG.md`. A pull
request from `security-plan`, rebuilt from `main` on every run, carries the plan and the
changelog to `main` and
merges itself once the tests pass. Where no Go version fixes a vulnerability yet,
govulncheck has no ignore list: the maintainer drops the release job's step "Stop the
release while the report names a known vulnerability" until one does, and records why
here. The paired ruling against f90b4d6 found the project reads worse for it: the gate
read `todo.md`, the filer spliced `todo.md` as text, and a release waits on vuln.go.dev.
The maintainer merged it anyway, as the cost of filing into the plan. Since 2026-11-19 the gate doesn't read `todo.md`, and the filer reads the
plan's rows only inside the item table. The filer still splices its item in as text, since
CI cannot run the maintainer's plan tools, and a release still waits on vuln.go.dev, which
this decision asks. The paired ruling of that change against f90b4d6 found the project
reads better, by the engine's and the CLI's files; no seat named a place on the release
path. The ruling against the commit before the change named the filer's call, its layout
list and the one release job as reading better. Valid while the binaries are built from the gate's Go version.

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
govulncheck before and after. Where GoReleaser moved, it also sets the tree's GoReleaser
images and takes `main`'s `goreleaser.yaml`, or the tree's own gate fails. The section names what changed, the vulnerabilities the second
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

The release job's last step plans known vulnerabilities, and runs after a failed check or
patch step too, so fixed items still move. It writes
under Unreleased's Breaking each requirement that a Go minor `main` carried since the last
release sets on a built-for system, so the release that first ships the minor names it,
whether CI cuts it as a patch or the maintainer cuts it. A release section copied into
`main` removes from Unreleased each Breaking line and each fix that the section already
holds. Where the step can't read the notes or `goreleaser.yaml`'s systems, it records no
such line and still moves fixed items, since the patch step fails on the same input. A
release job whose plan step fails after it published still hands its tag to the Packages
workflow.

A patch built with a new Go minor may drop a system the binaries ran on, such as an older
macOS. That departs from the technical principles' semver, read strictly, which counts a
dropped runtime version as a change to what exists. The maintainer, 2026-10-26, accepted
it until 1.0.0, for goal 1.1, and the README's Versioning says so. From 1.0.0 dropping a
system is a major; `todo.md` item 348 decides how a Go minor that drops one ships then.

The job compares with the last release, not with the push, and runs one at a time, so a
rerun cuts nothing twice and a missed run is caught up daily; publishing `main`'s heading
stays on a push, which the gate passed. `patch_release.py` takes the forge's latest release
for the last release, and the Packages workflow reads the same with `gh`. Each run copies into
`main`'s changelog any section the last release has and `main` lacks, so a lost run loses
no section, and moves rows planned under a patch CI took to the next patch. The Packages
workflow builds a tag with the Go its own Dockerfile names. The paired ruling against
5aca8a1 found the project reads worse for the Go patches: two linked scripts, four linked
jobs and several writers of `CHANGELOG.md`. The maintainer merged it anyway, as the cost
of releasing Go patches without waiting. Since 2026-11-19 the filer is a module `patch_release.py` imports, and one release job scans,
releases and plans, with the Packages workflow after it. `changelog.py` writes every
section and line CI puts in `CHANGELOG.md`; `go_release_notes.py` words the Breaking lines.
The paired ruling of that change against 5aca8a1 found the project reads better, by the
engine's and the CLI's files; no seat named a place on the release path. The ruling against
the commit before the change named the imported filer and the one release job as reading
better.
Valid while a release carries only its Go, the modules `go.mod` requires and the GoReleaser
that packages it.

## Goal 1.1's "where they look" means LLMs first: the ones developers ask, and the coding agents that pick a tool

2026-11-21, Lilleman auf Larv. Serves goal 1.1.

Developers find a tool by asking an LLM, by searching the web and by asking other people.
LLMs come first, in two ways. A developer asks one which tool to use, and a coding agent
picks one by itself while it builds something. A search for "fejkdata" by its own name
found nothing on 2026-11-21, and an LLM asked about fake-data tools named Faker, Mimesis
and Mockaroo instead. So an item that helps an LLM find fejkdata, recommend it for the
right task or use it correctly names goal 1.1. Such an item weighs more than one aimed only
at search engines or word of mouth. Valid while developers choose most of their tools
through LLMs.

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
On 2026-11-25 the maintainer confirmed this for scripts such as Han, Devanagari, Arabic and
Bengali: the locale ships a romanized column beside each name, such as pinyin or ISO 15919,
and its emails read that column.

Valid while every non-ASCII letter of a shipped first or last name has an entry in `ascii`'s
table, which `TestAsciiFoldsEveryShippedNameLetter` checks, and RFC 5322 keeps letters, digits and
`-` in atext.

## `ascii` folds the Latin letters of ruby-i18n's default table and the Vietnamese ones, except `ŉ`, so names such as `Wałęsa` and `Nguyễn` keep their letters

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

Revised 2026-11-25: ruby-i18n's table lacks `Ơ`, `Ư` and the vowels U+1EA0 to U+1EF9, so
`Nguyễn` became `Nguyn`, `Đặng` became `Dng`, and `slugify` gave `nguy-n`. The table now holds
them, each mapped to the first letter of its full canonical decomposition in Unicode 17.0's
`UnicodeData.txt`, so `Ấ` gives `A`. `ascii` already folds a decomposed `ễ`, the letter then its
marks, as it drops the marks; `slugify` turns each mark into `-`.

Valid while every entry of `ascii`'s table maps a letter to letters and comes from ruby-i18n's
default table or from Unicode's decomposition of a Vietnamese letter.

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
`platformAutomerge: false` would make a merge wait for every check.

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
`lowest-and-tooling` on two runners, each with caches of its own, since each runs its own
Go version and only `latest` runs the race detector. The build runs on the builder
`setup-buildx-action` made, since the cache mount lives in that builder and not in
Docker's default one. Go's caches key every entry on its inputs, so a restored cache
replays only work whose inputs are unchanged.

GoReleaser runs on the runner, outside the cache mount, so the Packages workflow keeps its
Go caches through `actions/setup-go`'s own cache. setup-go keys that cache by the runner's
image, the Go version and `go.mod`, and saves it at the end of a successful job, unless
the job restored that key. So a pull request's dry run starts from the build cache of the
first release with its key, and rebuilds whatever changed since. A failed save in setup-go
only warns, so an update that broke the save can't fail a release, and Renovate may merge
setup-go's updates.

A pull request never runs the save and the extraction, so Renovate leaves an update to
`actions/cache`, `buildkit-cache-dance` or its `busybox` image to the maintainer: one
that broke them would fail `gate` on `main`, and hold every release.

With the caches warm, three runs of a pull request's Tests workflow measured 1m47s to
1m53s on 2026-10-28, and one with no cache to restore measured 2m31s.

Valid while GitHub's cache serves a pull request the caches of its base branch.

## A Swedish street's postal codes keep delivery digit 9

2026-11-04, the maintainer. Goal 3.2.

sv.wikipedia's tables read xx9 xx in a two-position place as "Tävlingspost", temporary
codes for a large expected volume of mail, and xxx 9x in a three-position place as rural delivery.
Dropping 9 after a two-digit prefix would remove real street codes: directories list residents at
169 41 and 169 34 Solna and companies at 589 41 Linköping. Dropping them would leave Solna 3 codes.
GeoNames doesn't mark a temporary code, and PostNord publishes no open list, so the digit can't tell
a temporary code from a street code. A rare temporary code on a fake address is the smaller error.
Valid until an open register marks temporary codes.

## A Swedish postort with no published population of its own weighs 200

2026-11-04, the maintainer. Goal 3.2.

SCB publishes populations for a tätort, a kommun and an unnamed småort, none for a part
of a tätort, so 373 postorter such as Bromma and Farsta stay at the tätort floor. A suburb's people
are counted in the tätort that its neighbouring postort of the same name weighs by, so Stockholm's
share of draws stays near its real one and a suburb is drawn too rarely. The one published table by
postal area, `Totalbef_postnr.xls`, holds figures for 2012 and survives only in the Web Archive with
no licence stated, so it stays unused (the maintainer, 2026-11-05). Valid until SCB publishes a current table by postal area under a stated licence.

## A Swedish road segment inside a tätort leans towards the postort the tätort is named after, by distance over the fourth root of its codes

2026-11-04, the maintainer. Goal 2.

A segment went to the postort owning the nearest GeoNames point, which left Stockholm 363 street
names: most of its postal codes sit on a few points, close to its suburbs' points. So a segment
inside an SCB tätort (CC0) named after a postort of the same kommun can go to that postort instead.
To compare, the import divides each postort's distance by the fourth root of its number of codes, and
the shorter result wins. The exponent was tested on 46 well-known streets of the Stockholm area,
picked by hand. Plain nearest placed 21 right, a fourth root 30, an exponent of 0.35 placed 32, and a
square root 25. The fourth root won because 0.35 placed only two more and left Solna 61 names, and a
square root 12, against 234 before; it leaves Solna 210 and gives Stockholm 1580. Valid while no open
source maps a street to its postort; Lantmäteriet's address product needs personal-data terms, and
OpenStreetMap is ODbL.

## `sv_SE.word` holds SALDO's words, under one short path per part of speech: `noun`, `verb`, `adj` and `adv`

2026-11-04, the maintainer. Goal 3.2.

A panel of three readers, seated as bulk, Go and data authors, all chose `word.adj` over
`word.adjective` and over `sv_SE.adj`, since `en_US.sentence.adj` already shortens it. A word's
form sits below its path and reads one bound pick, as a company's `legal` does. `word` weighs a part
of speech by its row count, per "A table whose register publishes no frequency draws evenly".

SALDO marks no word as rude, so the import drops exact matches of LDNOOBW's Swedish list (CC BY 4.0,
widely used to filter generated text). It leaves compounds and any rude word the list lacks.
Valid while no open Swedish source tags register. Item 497 ends this filter: the project principle on a
locale's `word` keeps rude words.

## `en_US.word` holds WordNet's words, under one short path per part of speech, and its notice ships with them

2026-11-08, the maintainer. Goals 3.2 and 2.

`en_US.word` holds WordNet 3.0's nouns, verbs, adjectives and adverbs under the paths `noun`, `verb`, `adj` and
`adv`, as `sv_SE.word` does, and `verb.present` prints a verb's third person. WordNet lists lemmas only, so a verb is
always its base form; Moby Part-of-Speech II, the public-domain alternative, tags a verb only as "usually a
participle" and mixes `stirred` and `hiving` with base forms.

WordNet's licence (SPDX `WordNet`) asks for its copyright notice and disclaimer on every copy of the database.
`data/en_US/WORDNET-LICENSE.txt` holds them, `en_US` embeds the file, and the release archives and packages carry
it. `DATA-LICENSES.md` names the licence.

The import keeps a word that WordNet's data files spell in lower-case letters alone, since its index files
lower-case every lemma and would let `paris` and `abc` through. A verb gets its third person by the spelling
rule, with `be`, `have`, `stomach`, `quiz`, `whiz` and the `-o` verbs that take a plain `s` as exceptions, since
WordNet has no forms. The import drops a word whose every sense is in WordNet's usage domains of disparagement, ethnic
slurs or obscenity, which tag slurs such as `dago` that LDNOOBW's English list lacks, and the exact matches
of that list. A word with one tagged sense among others, such as `boy`, `queen` or `tool`, stays, so `gook` stays
for its other sense. Item 497 ends both filters: the project principle on a locale's `word` keeps rude words.

WordNet's sense counts cover a quarter of its lemmas, so `word` weighs a part of speech by its row count and
draws each word evenly, which keeps the spread this data exists to widen.
Valid while no open English source gives a frequency for every word.

## `de_DE.word.noun` carries `plural`, where the nouns of `nb_NO`, `nn_NO`, `da_DK` and `sv_SE` carry `definite`

2026-11-16, the maintainer. Goal 3.2.

Each locale's `word` holds an open dictionary's nouns, verbs, adjectives and adverbs under `noun`, `verb`, `adj`
and `adv`, drawn evenly and weighed by row count, per "A table whose register publishes no frequency draws
evenly". Norsk ordbank (Språkbanken, CC BY 4.0) gives `nb_NO` Bokmål 2005 and `nn_NO` Nynorsk 2012. COR₁
(Dansk Sprognævn, CC0) gives `da_DK` the words of Retskrivningsordbogen. Kotus's Nykysuomen sanalista
(CC BY 4.0) gives `fi_FI` the headwords of Kielitoimiston sanakirja. Wikidata's lexemes (CC0) give `de_DE`
its words and their forms. Each import drops
a word or form on LDNOOBW's list for its language, as the `sv_SE` import does; the German import compares
in lower case, since the list writes its nouns in lower case too. Item 497 ends this filter: the project
principle on a locale's `word` keeps rude words.

The Norwegian and Danish tables carry the forms `sv_SE` does: `noun.definite`, `verb.present` and
`adj.definite`. The Finnish list holds no inflected forms, so `fi_FI.word` carries none, and its sentences
use base forms alone. A German definite form needs the noun's gender, so `de_DE.word.noun` carries
`plural`, whose article is always `die`, and `de_DE.sentence` builds on plurals. A panel of three readers,
seated as bulk, Go and data authors, chose `plural` over `pl` and `many`, 3 to 0.
Valid while none of these languages has an open word list with a frequency for every word.

## `fr_FR.word` holds the English Wiktionary's French words, and its nouns carry `plural`

2026-11-20, the maintainer. Goal 3.2.

The tables hold the nouns, verbs, adjectives and adverbs of the English Wiktionary's French entries, as kaikki.org
extracts them (CC BY-SA 4.0), in lower-case letters alone. A word is a headword under its French head template, so
a verb is an infinitive: kaikki.org gives a conjugated form the generic head template alone. Wikidata's French
lexemes (CC0) hold 12,800 nouns, and Lexique 3.83 (CC BY-SA 4.0) lists a plural only for a noun its corpus held in
the plural, 16,000 of its 28,000, so neither reaches the tens of thousands of nouns the other locales hold. A noun
carries `plural`, as `de_DE`'s does: a French article needs the noun's gender and its elision, and `des` before a
plural needs neither. An invariable noun is its own plural, since kaikki.org lists it no form.
`fr_FR.sentence` builds on plural nouns, infinitives and adverbs, which take no agreement.
Valid while neither Wikidata nor Lexique reaches 20,000 nouns with a plural.

## `et_EE.word` holds Wikidata's Estonian words, and its nouns carry `plural`

2026-11-20, Lilleman auf Larv. Goal 3.2.

The tables hold the nouns, verbs, adjectives and adverbs of Wikidata's Estonian lexemes (CC0), as `de_DE`'s do, each
with a lemma of lower-case letters alone: about 59,000 nouns, 5,900 verbs, 9,100 adjectives and 4,200 adverbs. Wikidata
lemmatises a verb as its `-ma` infinitive. A noun carries its nominative plural, such as `majad`, and one with none is
left out. The English Wiktionary's Estonian entries (CC BY-SA 4.0, as kaikki.org extracts them) hold 3,900 such nouns,
so they don't reach the tens of thousands the other locales hold. The EKI Combined Dictionary and Estonian WordNet can
be had only by asking their maintainers, Ekilex (CC BY 4.0) needs an account's API key, and Eesthetic (CC BY 4.0) holds
5,000 nouns and 5,000 verbs, and neither adjectives nor adverbs. `et_EE.sentence` builds on plural nouns, singular nouns
and adjectives beside `on`, `-ma` infinitives after `hakkama`, and adverbs, none of which takes a case ending the tables
lack.
Valid while no open source of Estonian words holds as many nouns with a plural as Wikidata's lexemes.

## `lt_LT.word` holds the English Wiktionary's Lithuanian words, and its nouns carry `plural`

2026-11-20, Lilleman auf Larv. Goal 3.2.

The tables hold the nouns, verbs, adjectives and adverbs of the English Wiktionary's Lithuanian entries, as kaikki.org
extracts them (CC BY-SA 4.0). Each is a headword of lower-case letters alone, not a form or variant of another entry:
about 3,200 nouns, 900 verbs, 700 adjectives and 250 adverbs. A noun carries its plural, spelled as the page the plural
links to, which is without stress marks. Both the entry's declension table and its head template give one, and each is
wrong for some nouns, such as `lankytoja` for `lankytojas` in the head template. So a noun takes the one that links, and
is left out where both link and differ, or neither does. Wikidata's Lithuanian lexemes (CC0) held 14 nouns when counted
on 20 November 2026, and CLARIN-LT's
Lithuanian Hunspell dictionary lists no part of speech and is under an end-user licence, so no open source reaches the
tens of thousands the other locales hold.
`lt_LT.sentence` builds on nominative nouns, infinitives and adverbs, which need no gender, and quotes an adjective
as a word, since Lithuanian puts a neuter adjective after `viskas` and the table holds no neuter form.
Valid while no open source of Lithuanian words holds more nouns with a plural than the English Wiktionary.

## `lv_LV.word` holds Tēzaurs.lv's Latvian words in use, and its nouns carry `plural`

2026-11-20, Lilleman auf Larv. Goal 3.2.

The tables hold the nouns, verbs, adjectives and adverbs of Tēzaurs.lv's wordforms (Spring 2026, CC BY-SA 4.0), each with
a lemma of lower-case letters alone: about 51,000 nouns, 20,500 verbs, 8,700 adjectives and 7,500 adverbs. A lemma
that carries a usage label calling it regional, a dialectism, obsolete, historical, outdated or an old rarely used word is
left out, since Tēzaurs.lv holds many such words and a fake sentence should read as today's Latvian. A word labelled
colloquial, slang or rude stays, as the project principle on a locale's `word` keeps rude words. A verb is an infinitive, ending in `-t` or
`-ties`, since Tēzaurs.lv files some participles as verbs. A noun carries its nominative plural, such as `mājas`, and
one with none is left out. Wikidata's Latvian lexemes (CC0) hold 300 nouns, and the English Wiktionary's Latvian
entries, as kaikki.org extracts them (CC BY-SA 4.0), hold 5,200 nouns with a plural: neither reaches the tens of
thousands the other locales hold. `lv_LV.sentence` builds on plural nouns, infinitives and adverbs, which need no
gender, and puts an adjective after `viss ir`: the adjective's lemma is masculine singular, which `viss` takes whatever the noun
beside it, so the adjective needs no agreement with the drawn noun.
Valid while no open source of Latvian words holds as many nouns with a plural as Tēzaurs.lv.

## `{n.w}` is refused beside a read of `n` that reads `w` through a name of `n`'s category

2026-11-04, the maintainer. Goals 4.3 and 4.6.

`sv_SE.company` prints its wording through `{registered as r}{r}`, so its `legal` field reads the
word from that pick. The pick is `r`'s, so `{c}` prints `r`'s draw and `{c.registered}` draws its own.
A reader of `{c} {c.registered}` expects the one pick goal 4.3 promises, and each name keeping its
own pick says two, so by goal 4.6 the template fails to load, like `{n.w}` beside the format `{w}-{w}`.

Landing the read on `r`'s pick would make two names share draws. Moving the data's choice a level
down only moves the disagreement to the new level, and leaves users' categories with it.
Valid while each name keeps its own pick.

## No column is declared unique; `{seq()}` makes a column unique

2026-11-08, the maintainer. Goals 5.4 and 12.1.

A bulk fixture author asked for a column that redraws on a repeat, so a `UNIQUE` column loads.
Its values can run out, and in the library no run length is known at load, so running out would
be an error mid-render, which goal 5.4 rules out. Remembering every value drawn makes a render's
cost grow with the run, not with the data it reaches, against goal 12.1. Printing a repeat would
break the promise. Faker's `unique` raises after 1000 tries, which is the error mid-render ruled out
here. A column carrying `{seq()}`, such as `T-{seq(c)}`, is unique by construction.
Valid while goals 5.4 and 12.1 hold.

## `da_DK`'s names are Statistics Denmark's top 20 and a language model's recall

2026-11-08, the maintainer. Goal 3.2. Applies the project principle on name tables.

Statistics Denmark publishes the top 20 girls' names, boys' names and last names of the people
living in Denmark, and a count one name at a time; it offers no bulk table. Wikidata's Danish
names miss most of the common ones, such as Peter and Jensen. No good source of Danish name
counts was found online, and goal 3.2 prefers a good source without requiring one. The maintainer
chose the model's own knowledge over Wikidata.

So the top 20 of each list keep Statistics Denmark's counts of 1 January 2026, and every other
name is a language model's recall of a name common in Denmark. The commoner ones come first, from
the most to the least common, with an estimated count of the twentieth name's count times 20 over
its rank; the rarer ones follow in no order and share the count of the last rank. A recalled name
never outweighs the twentieth. The sex counts are Statistics Denmark's table FOLK1A.
Valid while no good source of Danish name counts is found; a bulk file replaces the recall.

## `de_DE`'s first names are Munich's newborns

2026-11-15, in the chunk that added `de_DE`. Goal 3.2. `de_DE`'s last names apply the project principle on name tables.

Germany's statistics office publishes no name counts. Cities publish their newborns' first names with
counts by sex, among them Munich, Bonn and Moers, and none covers Germany. Munich's thirteen yearly
files, 2013 to 2025, share one layout and one licence, the Datenlizenz Deutschland, Namensnennung 2.0, so
they alone are pooled: that gives about a thousand names, and a name weighs the sum of the numbers
published for it. Adding another city's files is a later widening. A file replaces a count below 5 (2025:
below 6) with a phrase, so a name published only so has no row. The names are those of children
born since 2013 to people living in Munich, not Germany's whole population.

No open register publishes German surnames, and the lists on the web are Wikipedia's (CC BY-SA) and a
genealogy wiki's. So the last names are a language model's recall of names common in Germany. The
commoner ones come ranked, from the most to the least common, and the rarer ones share the last rank's
weight. A ranked name weighs 700000 over the square root of its rank: the weights fall with the rank, as
surname counts do, but slowly enough that the 220th still draws, at a fifteenth of the first's
weight. The figure is an estimate.

The sex counts are Eurostat's table demo_pjan for Germany. Valid while no open register of German
names with counts for the whole population is found; one replaces the file or the recall.

## The project takes open licences, share-alike ones included

2026-11-19, the maintainer: "we want to use open licenses in general". Serves goal 3.2. A source under an open
licence is taken, CC BY-SA 4.0 and ODbL included. `DATA-LICENSES.md` names its licence and attribution, and the
README says what share-alike asks of a user. This ends the refusal of share-alike data that the decisions on
streets and on `et_EE`'s names record.

## `et_EE`'s names are a language model's recall

2026-11-19, in the chunk that added `et_EE`. Goal 3.2, goal 3.3. Applies the project principle on name tables.

Statistics Estonia publishes name counts by sex and year, and states that its open data is CC BY-SA 4.0
(stat.ee, "Principles of dissemination of official statistics"). When this table was built the project took no share-alike
data, so the file was no open bulk file of name counts under a licence the project takes. The first and last names are a language model's recall of names
common in Estonia, with none of the office's counts or ranks. The commoner names of each list are ranked, and
the rarer ones share the last rank's weight. A ranked name weighs 100000 over the square root of its rank, the
shape `de_DE`'s last names have at 700000. The figures are estimates.

The lists include the Russian-language names many Estonians bear. A woman bearing one takes its feminine
form, `Ivanova` for `Ivanov`, as Estonia's Russian-speaking women do, so `last-name` links to `sex` and has a
row under each: the Estonian names and the others the same under both, and a Russian name in `ov` or `ev`
with an `a` added under `f`. `person` draws the last name under its first name's sex.

The sex counts are Eurostat's table demo_pjan for Estonia. `title` holds `dr` and `prof`, as Sõnaveeb and the
Institute of the Estonian Language write them. Valid until `todo.md` item 522 replaces the recall with the
office's counts, which "The project takes open licences, share-alike ones included" allows.

## `lt_LT`'s top 50 first names are the Centre of Registers', and every other name a language model's recall

2026-11-19, in the chunk that added `lt_LT`. Goal 3.2. `lt_LT`'s first and last names apply the project principle on name tables.

The State Enterprise Centre of Registers publishes on data.gov.lt, under CC BY 4.0, the 50 commonest men's names and
women's names in the Population Register, with the count of people born since 1900 bearing each. It publishes no other
first name, a top 20 of the newborns of each quarter aside, whose counts are of births and not of the living, so the
decision leaves them out. Statistics Lithuania and the Centre publish no last names. So the top 50 of each sex keep their
published counts, and every other first name and every last name is a language model's recall of a name common in Lithuania.

A recalled first name doesn't outrank any published one: the commoner ones, 50 per sex, are ranked on from 51, and one of rank `r`
weighs the smallest published count times 50 over `r`, so the first of them weighs a little
less than the fiftieth published name. The rarer ones share the weight of the last rank. A recalled last name
weighs 100000 over the square root of its rank, as `et_EE`'s and `lv_LV`'s do, and the rarer ones share the last rank's weight. The
figures are estimates. The register holds the names of foreigners in Latin transliteration too, such as `Ivan`, and
they weigh as published.

A Lithuanian last name has the ending of its bearer's sex, and a name ships two feminine forms: the unmarried, `Kazlauskaitė`
for `Kazlauskas`, and the married, `Kazlauskienė`. Both are shipped under `f`, sharing the masculine name's weight, so a
woman's last names weigh as a man's and a woman is drawn either form. The feminine form follows from the masculine name's last
letters (`-ius`, `-us`, `-ys`, `-is`, `-as`, `-a`), so a name that Lithuanian spells as an exception takes the form the rule
gives. A Russian or Polish name ending in `ov` or `ski` takes `a`, as `Ivanova` and `Kowalska`; any other is the same under both. So `last-name` links to `sex`, and `person`
draws the last name under its first name's sex. `data-import/names-lt.py` holds the endings, and its test the pairs.

The sex counts are Eurostat's table demo_pjan for Lithuania. `title` holds `dr.` and `prof.`, curated. Valid while no open
register of Lithuanian last names or of names below the top 50 exists; one replaces the recall.

## `lv_LV`'s first names are the register's first given names, and its last names a language model's recall

2026-11-19, in the chunk that added `lv_LV`. Goal 3.2, goal 3.3. `lv_LV`'s last names apply the project principle on name tables.

The Office of Citizenship and Migration Affairs (PMLP) publishes the names in the Natural Persons Register on
data.gov.lv under CC0 1.0, with the sex of their bearers and the count of persons, in a file with each
name or combination of names. The import counts each person once, under the first of their given names, so a
name weighs the persons whose first given name it is, as `fi_FI`'s does. The 2,000 commonest of each sex are
kept. The register holds the names of foreigners in Latin transliteration too, and they weigh
as published; the table is the register's, not a list of Latvian names. A person of no stated sex is dropped.

The office's open data holds first names only, and no other open register of Latvian last names was found, so they are a
language model's recall of names common in Latvia, ranked and weighted as `et_EE`'s are: a ranked name weighs
100000 over the square root of its rank, and the rarer ones share the last rank's weight. The figures are
estimates. A woman bears the feminine form of a man's last name, which its ending gives: `Bērziņa` for
`Bērziņš`, `Balode` for `Balodis`. A name that does not end as a masculine form does, such as `Liepa`, is the
same under both. The endings are a rule, so a name that Latvian spells as an exception takes the rule's form. So
`last-name` links to `sex` and has a row under each, and `person` draws the last name under its first name's sex.
`data-import/names-lv.py` holds the endings, and its test the pairs.

The sex counts are Eurostat's table demo_pjan for Latvia. `title` holds `Dr.` and `prof.`, abbreviations
Cabinet Regulation No. 114 of 2 March 2004 lists for an academic degree and title. Valid while no open register of Latvian last names
exists; one replaces the recall.

## `da_DK.company` weighs its legal forms by Statistics Denmark's enterprises and draws its names from a language model's recall

2026-11-08, the maintainer. Goal 3.2. Applies the project principle on company categories.

Statistics Denmark's table GF15, "General enterprise statistics", counts the enterprises of 2024 by
form of organisation: 132,403 sole proprietorships (ENK), 218,208 private limited companies (`ApS`),
30,750 public limited companies (`A/S`) and 11,679 partnerships and limited partnerships. The
company's forms weigh by those counts. The table doesn't split the partnerships between `I/S` and
`K/S`, so they draw evenly. Co-operative societies, funds, public authorities and the table's other
category are not drawn, nor is `P/S`, which no category of the table names.

No open bulk file of Danish company names was found. So the name shapes and the words of
`company-coined` and `company-word` are a language model's recall of how Danish companies are named,
each shape and word weighing the same, as for a table whose register publishes no frequency. An open
bulk file of Danish company names replaces them, as Enhetsregisteret's names shaped the Norwegian company.

## `fi_FI.company` weighs its legal forms by Statistics Finland's enterprises and draws its names from a language model's recall

2026-11-12, Lilleman auf Larv. Goal 3.2. Applies the project principle on company categories, approved 2026-11-12.

Statistics Finland's table 13w6, "Yritykset toimialoittain ja oikeudellisen muodon mukaan", counts the
legal units of 2024 by legal form: 295,203 natural persons (sole traders, `tmi`), 240,524 limited
companies (`Oy`), 13,356 limited partnerships (`Ky`), 4,726 general partnerships (`Ay`), 2,811
co-operatives (`osk`) and 51,401 of other forms. The company's forms weigh by the first five counts. The
table doesn't split its other forms, so none is drawn. A sole trader's name has no abbreviation, as the
Danish ENK's has none.

No open bulk file of Finnish company names was found. So the name shapes and the words of
`company-coined` and `company-word` are a language model's recall of how Finnish companies are named,
each shape and word weighing the same. An open bulk file of Finnish company names replaces them. The names
don't read a place. A later change may add a place from `geo/FI` and a trade word as a name's shape, as
`da_DK.company` has.

## `de_DE.company` weighs its legal forms by Destatis' legal units and draws its names from a language model's recall

2026-11-16, Lilleman auf Larv. Goal 3.2. Applies the project principle on company categories; the choices below are the
implementer's recommendation, not yet the maintainer's.

Destatis' Statistisches Unternehmensregister table "Rechtliche Einheiten nach zusammengefassten Rechtsformen" counts the
legal units of 2024: 2,062,752 sole proprietors, 440,951 partnerships and 861,720 corporations, and 178,442 of other
forms. The company draws the first three groups by those counts. The table doesn't split a group, so a partnership is
a `KG` or an `OHG` and a corporation a `GmbH`, a `UG (haftungsbeschränkt)` or an `AG`, the forms of a group drawn
evenly, as the Danish partnerships are. That draws an `AG` as often as a `GmbH`, which the real counts don't, and a
finer Destatis table replaces the split. The other forms, such as associations and foundations, are not drawn,
as the Finnish table's other forms are not. A sole proprietor's name has no abbreviation, as the Danish ENK's has none.

No open bulk file of German company names was found. So the name shapes and the words of `company-coined` and
`company-word` are a language model's recall of how German companies are named, each shape and word weighing the same.
The names read `de_DE`'s names and `geo.DE.locality`. An open bulk file of German company names replaces them.

`de_DE.email` writes `Müller` as `muller`, not `mueller`, since `slugify` folds `ü` to `u` by ruby-i18n's table, as for
every locale. Its own words hold no `ae`, `oe` or `ue` spelling, so a local part never mixes the two.

## A `hidden` field writes no column and is left out of the listing

2026-11-08, Lilleman auf Larv, as `let`; widened 2026-11-19, and renamed `hidden` the same day.
Serves four goals:

- 4.7, read as: a template that means one thing loads.
- 4.3: a name keeps one pick.
- 4.8: a clearer name replaces the old one.
- 4: a reader tells a path meant to be read from a scaffold.

A record wrote one column per field, so a field there only to be named by another, as
`{"ride":…}` is by `{ride.fare}`, showed up as a column nobody asked for. `hidden` holds such
fields: they bind and resolve as any field does and write no column. It costs two reserved
keys: a field can no longer be named `hidden`, nor `let`, its old name. A key in `hidden` and beside it fails to load, so a
field has one place, and an option name in `hidden` fails too. A record whose fields are all in
`hidden` has no columns and is `ErrNoColumns`.

The same field was also a listed path outside a record: `--list` printed it, the `.fejkdata.json`
index wrote it and `testdata/shipped_shape.txt`, which pins the shipped paths, held it, so a reader
couldn't tell a path meant to be read from a scaffold such as `misc.ipv6.g`. So a `hidden` field, at
any depth, is left out of `--list` and the index with every path below it, and a read of its path
still draws it. This narrows the listing further than "`List` advertises direct descents only".
Only the listing leaves it out: a choice's shared paths, which resolve a read through the choice,
keep it. A prefix such as `_` would have made a key's spelling carry meaning and changed what a
user's own `_x` field means.

A shipped field is moved into `hidden` when only its own category's fields read it, no README text,
test or other category reads its path, and it is a piece only the template needs: a syllable
half, a digit class, a price's amount or separator-bearing part, a number's head or tail. A part
the README names or a domain gives a name, such as `phone.block`, an SSN's `area` or an email's
`local`, stays a field and a column. The pieces inside the variants of a `username` or an
`email.local` choice, such as `first`, `last`, `sep` and `n`, stay listed too: the audit moved none
that a choice's variants share.

`let` is the option's old name: it shipped from 0.4.0 through 0.6.0, chosen on 2026-11-08 by a writer
panel over `hidden`, `private` and `_`-prefixed names, as the spelling readers took for a local
binding. Goal 4.8 has the clearer name replace the old one, so `let` is gone as an option. It stays
reserved: a template written with `let` fails to load and the error says the option is now named
`hidden`, because a bare `let` taken for a field would write a column and change what loads. A
later major may free the word.

The panel that chose `hidden` was three fresh readers, seated as the README's data author, hand
fixture author and contributor, each shown `hidden`, `let`, `local` and `scratch` in another
order. All three ranked `hidden` first and `let` second. A reader takes `hidden` for a field that is
drawn and read by its path yet left out of columns and listings, as a dotfile is. A reader takes
`let` for a binding scoped to its template, which says nothing about being left out of a listing.
`local` reads as scope, which would refuse a read from another category.

Valid while a data author reads `hidden` as "a field drawn and read, never listed".

## No database or HTTP sink is built in

2026-11-08, Lilleman auf Larv. Serves goal 6, one command line gets you a value, and goal 10,
every runtime dependency is one the standard library can't replace.

`fejkdata` writes JSON, CSV, TSV and SQL `INSERT`s to stdout, and the README shows four common
databases, PostgreSQL, SQLite, MySQL and MongoDB, and an HTTP endpoint reading that output on stdin. A built-in sink would add a driver per database, or an HTTP client with its
authentication and retry rules, each a dependency the standard library can't replace, for a step
`psql`, `sqlite3`, `mysql`, `mongoimport` and `curl` already do. Valid while those tools read
what `fejkdata` writes.

## `misc.creditcard-test` and `misc.iban-test` are tables of published test numbers

2026-11-08, Lilleman auf Larv. Serves goal 3.3: a drawn `misc.creditcard` number or `{iban(CC)}` can reach a real card or account, so a form made of numbers published for testing sits beside it.

A panel of three readers, each seated as one or two of the README's authors, chose the spelling:
`creditcard-test` and `iban-test` won two to one over `testcard` and `testiban`, because `--list` puts
`misc.creditcard-test` beside `misc.creditcard` and `-test` follows `personnummer-12` and `phone-exchange`.

Each is a table keyed by brand or by country code, one row per key. The cards are the brands' rows of
Stripe's testing page (docs.stripe.com/testing, "Simulate a payment by card brand"), read on 2026-11-08.
The IBANs are the first row, the success case, of its SEPA Direct Debit test IBANs for BE, DE, ES, FI,
NO and SE. Stripe lists no Danish IBAN, so `DK` is the example on Danske Bank's page
(danskeci.com, "IBAN"), which calls its example accounts fictive. Neither source says that no real card
or account holds a number, so the `misc` page claims only that they are published for testing. Valid while
those pages state them.

## A date offset is `adddays(d,n)`, and its `d` reads `date` or `born` calls in one layout that reads back

2026-11-08, Lilleman auf Larv.

Goal 4 asks that a template be easy to write and read, and goals 5.1 and 5.4 that a value its rules
forbid fail at load. A panel of four readers, seated as the bulk fixture author, the Go test author, the
JavaScript test author and the hand fixture author, chose a day-offset function four to none over a
`date()` whose bounds read a name and over `calc` on a date: its unit is in its name, and neither `date`
nor `calc` changes meaning. A second panel of three chose the name `adddays` three to none over `days`, `dateadd` and
`plusdays`; `days` read as a count. The same panel chose, two to one, an optional third argument for
the layout, which defaults to `d`'s.

`adddays` parses the text `d` printed, so it reads only what proves one layout at load, and only a
layout that parses back as the day it printed, checked on every day of a leap year and on days in
years 1, 1969 and 9999. It does not read
another `adddays`, whose result could pass year 9999. Valid while goal 4 stands and no builtin but
`date` and `born` draws a date.

## A where clause follows its path: `length 5`, `length 4 to 6`, `not 'cat'` and `not containing 'cat'`, joined by `and`

2026-11-08, Lilleman auf Larv.

Goal 4 asks that a template be easy to write and read. A panel of four readers, seated as the bulk
fixture author, the data author, the JavaScript test author and the hand fixture author, all ranked
first a clause after the path, where `as` goes: `{/en_US.word where length 5}`. Each read
a function form such as `{length(/en_US.word, 5)}` as measuring a length. Each read an excluded `'cat'`
as that word alone, never `concatenate`, and asked for a word of its own, such as `containing`,
before any substring match. The maintainer chose the spelling from that: `length 4 to 6` includes both ends, and
takes `to` since `..` already steps up a level (goal 4.4). A length counts Unicode code points, which Go
and JavaScript count alike. Counting what a reader sees means segmenting graphemes: Go's standard
library doesn't, and JavaScript's `Intl.Segmenter` follows its runtime's Unicode tables, so the two
could count one value differently, which goal 9.2 rules out. Valid while goal 4 stands.

## A distinct draw is `distinct N` after any `where` and before `as`; its members read `[1]` to `[N]`, and `joined 'sep'` prints them as one

2026-11-08, Lilleman auf Larv.

Goal 4.2 keeps `{d},{d}` as two draws, so several different values of one path in one pick take a spelling of
their own, and goal 4 asks that it be easy to write and read. A panel of four readers, seated as the bulk
fixture author, the data author, the JavaScript test author and the hand fixture author, all chose
`distinct N`: they read `pick` as allowing repeats and `take` as the first rows. All four counted
members from 1; two chose `{t[1]}`, one `{t#1}` and one `{t.1}`. For one value, two chose
`joined ','`, one `sep ','` and one `{join(d, ',')}` over a bound name.

The two readers who chose otherwise read `[1]` as a selector's key. On a distinct name `[k]` is always a member,
since such a name refuses every other tail. Valid while goal 4 stands.

Members are different values, not different items: each value of a nested choice counts on its own.
`/en_US.word` is a choice of four parts of speech, so counting items would draw different parts of
speech, never three different words.

## `{reword(text,words)}` replaces each word of `text` with one drawn from the path `words`

2026-11-08, Lilleman auf Larv; renamed 2026-11-09.

Goal 4 asks that a template be easy to write and read, and a name that promises more than the call
does misleads its reader. The builtin's output is not anonymised (GDPR recital 26: the person is no
longer identifiable) and not pseudonymised (Article 4(5): the data can no longer be attributed to a
person without additional information, kept separately). Its digits stay, so a personnummer in the
text passes through whole. The name says what the call does to the text.

A panel of four readers, seated as the hand fixture author, the bulk fixture author, the JavaScript
test author and the Go test author, first ranked `anonymize` ahead of `reword`, `fakewords`,
`rewrite`, `scramble` and `mask`; all four read the text first and the word list second, hence
`(text,words)`. Asked again for a name that claims neither term, three put `reword` first and one
`fakewords`. All four put `rewrite`, `scramble` and `mask` last: those names read as rewriting the
digits too, shuffling letters or redacting. `reword` can read as a paraphrase, so the README's
Transforms section says the fakes are random words.

`words` is a path, so the template picks the language of the fakes. Valid while goal 4 stands.

## A where clause also filters a call drawing from an alphabet, reads a name unquoted, and may end in `else` and a fallback

2026-11-09, Lilleman auf Larv.

A program replacing real values needs a fake that holds no part of its original, such as a uuid
holding none of a document's own tokens. The maintainer chose that a template may read those strings
at render, from a name, and may name a fallback. Strings read at render can leave no value to draw,
and goal 5.4 rules out a render error, so a draw then prints the fallback.

A panel of four readers, seated as the Go test author, the JavaScript test author, the bulk fixture
author and the data author, picked the spelling. All four chose `else` over `otherwise`, `fallback`,
`default`, `or` and `??`, last in the clause and before `as`. All four read an unquoted word in
`not` as a name, and a name as its one pick. On a field holding a list, all four leaned to reading
every string, but only 55 to 65 percent sure, so, as goal 4.6 asks, `New` refuses a field there,
and its error shows both meanings: bind the field to a name for one of its values, or quote each
string for every one. A joined distinct name read whole is refused for the same reason, and its
error lists the members to read. All four wanted the fallback to take a name or a path; `else` takes
a name, so a path is bound to one first. Valid while goal 4 stands.

Without `else`, a path prints an empty string, so no category needs a default. A call prints a
value of its own shape, such as the Nil UUID, which RFC 9562 defines, so a column holding a uuid
still holds one. The fallback or the default may hold one of the strings, as the Nil UUID holds
`0000`; only a render error could rule that out.

Skipping each character that would complete a string can get stuck: `{digits(2)}` avoiding `'00'`
to `'09'` has no second digit after a first `0`. So a call's draw first plans which characters each
position may take, over an Aho–Corasick automaton of the strings, and never retries. The plan costs
about the value's length times the strings' total length times the alphabet's size, which goal 12.1
allows. Each position draws evenly among the characters the plan leaves it, so the values that pass
aren't equally likely. An even draw over them would weigh each character by the count of passing
values it leads to, and a uuid's counts already exceed 2^64.

## One run writes several linked templates through a repeated `--emit <template>=<count>`, and a template reads another's rows with `{/run.<template>}`

2026-11-11, Lilleman auf Larv, from a bench of five bulk fixture authors, who chose the flag 4 to 1 over a manifest file and a folder of templates, and `{/run.…}` 5 to 0. Serves goal 1, and goal 6.1: a repeated flag works like other command-line tools'.

A template here is a category of the data folder, named by its path, such as `customers`. A child is a template that reads another's rows, and that other is its parent. What holds:

- A template reads another's rows only through `{/run.<template>}`.
- The run fails at load (exit 2) in three cases: it doesn't name a template that another reads; it names one with count 0 while a child's count is above 0; or its templates read each other in a cycle, which the error names. Both counts 0 is fine.
- A child may outnumber its parents, and a parent may have no child. The linked tables' load error for "a parent row no child links to" doesn't carry over.
- Each template's rows come out together, parents first, so a SQL load never sees a dangling key. Templates that read nothing of each other come in the order of their `--emit` flags.
- Under `--format sql` each template's name is its INSERT target. With csv, tsv or ndjson, more than one template needs `--out-dir <folder>`, which writes `<template>.<extension>` for each, and without it the run exits 2.
- `--emit` is refused beside `--table`, `-n` and a template on stdin.
- Adding a template to a run doesn't change the rows another template draws under the same `--seed`.

Each reader answered before seeing the candidates. Two wrote a flag of pairs, `--emit customers=1000` or `-t customers:1000`; three wrote bare arguments, one with `--out ./out`. A flag wins over bare arguments because an argument that isn't a flag is misuse today. Four then chose the flag over a manifest file, which drifts from the command that sets the counts, and over a folder of templates, which mixes how many rows with what a row looks like. The fifth chose the flag for work by hand and a manifest for a scheduled job, and wanted both to take the same fields.

All five chose `{/run.customers as c}` over a flag that binds one template's column to another's. That flag makes a template mean a real customer in one run and an invented one when run alone. They also chose it over a new keyword, since `as` already binds a name.

All five read `--run` as something that executes. Three read `--record` as one row, a session or an output format, and all five ruled out `--table customers=1000`, since `--table` names the INSERT target. `--emit` is the one spelling no reader read as another thing. Three readers feared a category named `run` could clash with `{/run.…}`; item 393 reserves the word or refuses the clash, and says which here.

Item 393 builds it. Valid while the README's Audience names bulk fixture authors.

## `geo.FI`'s streets come from Posti's basic address file, which holds none on Åland

2026-11-13, in the chunk that added `geo.FI`. Serves goal 3.2.

Posti's basic address file names every street under its postinumero, in Finnish and, where Posti has one, in Swedish. Digiroad names the streets too, but as road links without a postinumero, so each link would have to be placed in one. Both of Posti's files come under one set of terms, so `geo.FI` takes its postinumerot and its streets from Posti. The terms let anyone download the files and pass them on, provided the recipient is "aware of the terms of use for the service as well as the download date". DATA-LICENSES.md links the terms and names both days, and the archives, the packages and the image all carry it.

The file holds no street on Åland, so Åland's maakunta and kunnat ship no place. A postinumero weighs its residents in Statistics Finland's Paavo, and a street its ranges in Posti's file, since neither source counts addresses. Valid while Posti's files are free to download and hold no street on Åland.

## A `geo.FI` `swedish-name` holds the Finnish name where its source gives no Swedish one

2026-11-13, in the chunk that added `geo.FI`, from a bench of two readers. Serves goal 2.1. Widened to the postitoimipaikka and the kunta in the chunk that gave them the column.

A street's `swedish-name` is the Swedish name Posti gives it most often in its postitoimipaikka, the first in the alphabet winning a tie. Where Posti gives none, it is the Finnish name: a TSV cell can't be empty, and a Swedish address writes the Finnish name of a street that has no Swedish one. Both readers chose the spelling `swedish-name` over `name-sv`.

A postitoimipaikka's and a kunta's `swedish-name` follow the same rule. A postitoimipaikka's is the Swedish name Posti gives its postinumerot most often. A kunta's is Statistics Finland's.

Valid while a `geo.FI` table names one language per column.

## A table's parent column is read in one pass over its runs of equal cells, and a table larger than one block loads by block from a block index `go generate` writes

2026-11-14, in the chunk that measured the cold start. Serves goals 12.1, that a render's cost grows only with the data it reaches, and 12.2, that the first value of any category appears in under 100 ms. Revised the same day, after a study of bulk speed and data growth: loading by block, first built only if one pass left the CLI's first `en_US.address` at 100 ms or more, is now unconditional and covers every table larger than one block. Carried out by `todo.md` items 465, 470, 471 and 472.

The first reach of a country's streets is the cost; `New` and every later draw are not. Measured at 12c96997 with Go 1.27.4 in a container: `New` took 0.3 ms with `en_US`'s modules and 1 to 2 ms with every shipped module, the library's first `en_US.address` took 126 to 142 ms and peaked near 80 MB, and its second took 20 µs. In the CLI, a name or a uuid took 5 to 7 ms; `en_US.address` took 150 ms and peaked near 80 MB, `geo.FI.address` 87 ms and 53 MB, `sv_SE.address` 75 ms and 37 MB, and `geo.NO.address` 48 ms. A CPU profile of 40 first draws put 72% in loading the tables a category reads and 13% in building the row lookup on the first draw. For `geo.US`'s 475,271 street rows, 10.8 MB under 1,716 localities, the phases were:

| Phase | Time |
|---|---|
| Split the cells | 24 to 33 ms |
| Parse the weights | about 25 ms |
| Prove each parent weighs above 0 | about 20 ms |
| Link each row to its parent | about 18 ms |
| Build the row lookup | about 19 ms |

The strategies, with what a spike measured:

- **One pass over the runs of equal parent cells** does the last three phases in 3.7 ms, about 53 ms saved. A run is a stretch of adjacent rows with the same parent cell. That puts the library's first `en_US.address` at 75 to 90 ms and the CLI's at about 97 ms. The five street tables, 110,503 to 475,271 rows each, hold their rows grouped by parent; 18 of the 32 shipped tables with a parent column are scattered, the largest of them 13,724 rows, and they must cost no more than today. Chosen first: it is the smallest change and moves no format or output.
- **Load a table by block, from a block index `go generate` writes beside the manifest.** A block is one parent's rows, or a fixed number of rows in a table with no parent column; item 465 sets the block size, and a table larger than one block gets an index. The block index is not the manifest's index, which lets a source load a category on first reach: it lets a category load part of one table on first reach. Chosen second, unconditionally.
  - **First spike**: an index of 1,716 lines and 55 KB for `geo.US`'s streets was read in 0.3 ms, and one parent's streets parsed in 19 µs. The item's 208 KB trimmed copy of the US data took about 6 ms and peaked near 32 MB.
  - **The study's spike**, at 1, 5 and 10 times today's US street table, took 13 to 30 µs for the first reach of one locality at every size, with an index of 67 KB at 1 times and 669 KB at 10 times. Today's loader takes 645 to 688 ms and 276 MB at 5 times, and 1.32 to 1.50 s and 552 MB at 10 times, so each country added later makes its first reach slower, however one pass reads its table.
  - **What it touches**: `rows`' whole-table indexes, and an index to keep current. A draw over the whole street table with no locality draws a parent by its total weight, which may change what a seed draws; storing each parent's starting cumulative weight and summing a block in file order may keep it, and item 465 settles that.
  - **Proofs**: the index states what the whole-table proofs prove. `go generate` and a test run them, and any module may carry an index, so a shipped module still loads as anyone else's does (goal 8.1). A stale index fails: it holds the table's byte length, checked at load, and each block's parse checks its row count and parent key against it, so a hand-edited TSV with an old index is a load error naming the table, never another parent's rows.
  - **A run that reaches every parent** still holds the table. Item 470 holds a parsed block as its text and row offsets, about 27 bytes a row where today's cells take about 110 (estimated).
- **Parse the cells and weights in two goroutines** took 20 ms where one fused pass over both took 32 ms. It saves at most 12 ms and leaves the proofs. Rejected.
- **Skip the proofs for shipped data** saves about 38 ms, but loads a shipped module differently from anyone else's, which goal 8.1 forbids. Rejected.
- **Ship fewer streets** would parse a tenth of the file in 2.5 ms, but ships less data, against goal 2 and the default-embed decision of 2026-10-11. Rejected.
- **A deflated block per parent** was spiked: `geo.US`'s street TSV shrinks from 10.8 to 3.05 MB, but a parent's first reach takes 40 to 65 µs against 13 to 24 µs uncompressed in the same spike, and a draw across every parent takes 250 to 320 ms per million against 146 to 156 ms. It needs a second loader, and the repository and the module zip carry the data twice. A binary form `go generate` derives from the TSV would keep goal 7, since the TSV stays the hand-edited source; it is rejected for now and revisited where one of the limits below is near.
- **Splitting the data differently** is not needed yet. The limits, with today's data at 36.9 MB tracked: GitHub blocks a file over 100 MiB (cited), about 9 times `geo.US`'s street file; the Go module zip is limited to 500 MiB (cited), about 13 times; and the GitHub Pages site holding the apt, dnf and apk repositories is limited to 1 GB (cited), near at about 10 times (estimated). The CLI binary, 40 MB of which 85% is data, and `go test ./data/...`, about a minute, are costs, not limits. Item 471 gives each module its own test package, so the test cache skips data that didn't change, and item 472 asks the maintainer how to split files and modules before the first limit.

Valid while a shipped table is a TSV, and the shipped data may grow by several times.

## Bulk speed comes from fewer allocations, a higher GC target and rows rendered on every CPU, and a weighted draw stays a binary search over cumulative weights

2026-11-14, from the study of bulk speed and data growth behind the decision above. Serves goal 1, the best fake-data tool, and goal 12.3, a million rows in under 10 seconds on one CPU, scaling with the CPUs. Carried out by `todo.md` items 466, 467, 468, 469 and 473.

Measured with Go 1.27.4 in a container on 16 CPUs, three runs each. A million rows is 1,048,576, written as CSV to a file: `en_US.person` took 2.95 to 3.25 s, `en_US.address` 6.3 to 7.0 s, `sv_SE.person` 9.4 s, and the README's eight-column `person.json` 13.3 to 13.7 s, 76 to 79 thousand rows a second. Every tool wrote the same ten-column US record: fejkdata ran at 79 to 85 thousand rows a second, gofakeit v7.18.4 at 426 to 465 thousand, faker-js 10.6.0 at 75 to 81 thousand, Mimesis 22.2.0 at 72 to 75 thousand and Faker 40.54.0 at 3.1 thousand. gofakeit's rows don't hold together (a city of Texas under the state WY); fejkdata's do. In a CPU profile of the README's record, garbage collection took about 19%, `mallocgc` 16%, string-keyed maps about 20% and the weighted draw 3.2%.

The path, in order, each step with what its spike measured:

- **Fewer allocations**: four changes together took the README's record from 11.8–13.6 µs and 53.8 allocations a row to 8.8–10.3 µs and 28.8, `sv_SE.address` 22% faster, with the same bytes on six templates of 200,000 rows. They are memo slots numbered at load and child rows indexed by parent row (item 466), and no builder for a one-read template and `CSVLine` written into one buffer (item 467). Swapping a map for a slice without numbering the slots gained only 3 to 9%. Rendering a whole record into one reused buffer would gain a further 15 to 25% (estimated, not spiked; item 467).
- **A GC target of 400 in the CLI**, unless `GOGC` is set, ran the README's record in 10.6 to 11.2 s where the same comparison's default took 12.6 to 13.5 s, for 134 to 149 MB where 64 MB was. `GOGC=off` took 5.3 GB, and `GOMEMLIMIT` alone made no difference (item 468).
- **Rows rendered on every CPU** ran the README's record at 68 to 82 thousand rows a second on one worker, 296 to 303 thousand on four and 572 to 648 thousand on eight. Every row is seeded from the run's seed and its row number, on the serial path too, so the bytes don't depend on the number of CPUs (goal 9), and `{seq()}` and `--emit`, which need rows in order, render serially without changing another column's values. That changes seeded output once, which the README's Versioning section allows in any release, a patch too (item 469).
- **Publishing the comparison**: item 473 measures fejkdata again after the steps above, and publishes it with the command that reproduces each row.
- **The alias method** for weighted draws gains at most 3% and changes seeded output. Rejected; a weighted draw stays a binary search over cumulative weights.

Valid while a row's cost is its allocations, not its weighted draw.

## The German id draws ELSTER's test series

2026-11-16, Lilleman auf Larv.

Goal 3 (a valid id passes the check the real system applies) and goal 3.3 (a form that reaches
nothing real) are both served. The Federal Central Tax Office's (BZSt) "Erzeugung von
Identifikationsnummern" says an Identifikationsnummer has no leading zero, and "Prüfung der Steuer- und
Steueridentifikationsnummer" of ELSTER, the tax authorities' online portal, allows one only for a test
number. So `de_DE.steuer-id` draws a leading zero, and no drawn number is a person's. The check digit
and the other rules (one digit twice or three times among the first ten, never three in a row) are the
same for a test number, so `{steuerid()}` computes the eleventh digit by ELSTER's algorithm. This follows
the Swedish, Norwegian and Finnish ids, which draw an issuer's test or unissued series.

Valid while the BZSt doesn't issue any number with a leading zero.

## `fr_FR.company` draws five forms, names an `EI` by its owner, and takes its words from Sirene

2026-11-19, Lilleman auf Larv.

Goal 3.2 (a table follows the register that publishes it) is served: Sirene's StockUniteLegale is
an open bulk file of French company names under the Licence Ouverte 2.0, so `fr_FR.company` does not
draw from a language model's recall as the project principle for countries without one does. The
forms are the five a French company name usually ends in: `EI`, `SARL`, `SAS`, `SNC` and `SA`. The SCI
of category 6540 holds 1.9 million units but names the property it owns, and the associations of
9220 are not companies, so neither is drawn. An `EI`'s name is its owner's first and last name, as the
register holds it in `nom` and `prenom`, with no `EI` after it, although a sole proprietor must add
`EI` to their name on documents since 2022: the register's names carry none, and the shared company
test refuses a bare form's name that ends in its code. Sirene files an `EURL` under 5499 and a `SASU`
under 5710, as no unit holds 5498 or 5720, so they count as `SARL` and `SAS`. Sirene
writes a name in capitals without accents, so the 120 commonest trade words have their accents from
a list written by hand, which also spells the roots that take an accent, and any other root keeps Sirene's spelling. A longer list would need
a source of French spelling that the project doesn't take yet.

Valid while Sirene holds no accents and no French word list is shipped.

## `geo.EE` reads its places and addresses from one file, sums a linnaosa into its linn, and a place with no street ships no place

2026-11-20, in the chunk that added `geo.EE`. Serves goal 3.2. Applies the project principle on a locality with no street.

The Land and Spatial Development Board's file of postal codes lists every address of its address data system with its sihtnumber, its full address and a point, under CC0 1.0. It names each address's maakond, vald or linn and asustusüksus, so one file gives every table its rows and a count of addresses for each. A postal code and a street weigh by that count, and a place by its people, as the decision on the three Baltic modules' weights says.

The file names places without their codes, and külad share names, as three named Uusküla do, so the board's settlement units give each place its EHAK code, which selects it. Their page calls them open data and asks for a credit with the data's date, which DATA-LICENSES.md carries. A linnaosa is a part of its linn, and a letter to one names the linn, so its addresses are summed into the linn: Tallinn is one place.

An address in most külad is a farm's name, such as `Antsu`, with no street and no house number. Every other country's address is a street and a number, and `geo.EE.address` keeps that one layout, so a locality with no street is dropped. 1,131 localities remain, holding 78% of the addresses; the municipalities Kihnu, Muhu, Ruhnu and Vormsi hold none of them and are dropped too. A street is a near address ending in a house number after one of eight kinds: `tn`, `mnt` and `pst`, the three the board's handbook of address data abbreviates in section 27.14, and `tee`, `allee`, `põik`, `väljak` and `plats`, the kinds it names as written in full.

Valid while the address data is CC0 1.0 and a locale's `address` prints one layout.

## `geo.LV` takes a novads or a state city as its region, counts each apartment as an address, and writes the postal code after the locality

2026-11-20, in the chunk that added `geo.LV`. Serves goal 3.2. Applies the project principle on a locality with no street.

The State Land Service publishes the State Address Register as open data under CC BY 4.0: each novads, pilsēta, pagasts, ciems and iela, each building's address with its postal code and a point, and each apartment with its building. `geo.LV` reads only the rows in force, and counts a building as one address and each apartment in it as one more.

The register holds no level above the novads: the planning regions are no part of an address. So a region is a novads or one of the seven state cities, a municipality a pagasts or a pilsēta, and a locality a pilsēta or a ciems. A pilsēta is then a municipality and a locality of one name, and a state city a region too, which keeps every table's parent link whole. A region and a municipality keep the ATVK code the register carries for them. A ciems has no ATVK code, so a locality keeps its code in the register.

A building alone would weigh a block of flats as one farmhouse, so a building is one address and each apartment in it one more. Rīga then holds 30% of the addresses. House numbers are counted per building, since an apartment's number is no part of the record.

An address straight under a ciems or a pagasts is a house's name, such as `Riņņi`. `geo.LV.address` keeps the one layout of a street and a number, so a locality with no numbered address on a street is dropped, as in `geo.EE`. 819 localities remain, holding 83% of the addresses; 141 pagasti hold none of them and are dropped too. A korpuss, as in `195 k-1`, and a second number, as in `11/13`, are dropped from a number before it is counted.

The Universal Postal Union's guide to Latvian addresses, of September 2023, puts the postal code to the right of the locality's name, preceded by `LV` and a dash, and separated from the name by a comma. So the layout is `{street} {street-number}\n{locality}, {postal-code}`, and a postal code is stored with its prefix, `LV-1050`, as the register writes it.

Valid while the register's open data is CC BY 4.0 and a locale's `address` prints one layout.

## `geo.LT` names an apskritis and a savivaldybė with its kind, counts each of the premises as an address, and ships no point

2026-11-20, in the chunk that added `geo.LT`. Serves goal 3.2. Applies the project principle on a locality with no street.

The Centre of Registers publishes the Address Register as open data under CC BY 4.0, in files of its own site: each apskritis, savivaldybė, gyvenamoji vietovė and street, each address with its number and postal code, and each of the premises given an address. `geo.LT` reads them there rather than from the national open data portal, which serves the same tables and refuses requests under load.

An apskritis is a region, a savivaldybė a municipality and a gyvenamoji vietovė a locality, each under the code the register gives it. A seniūnija is no part of an address and is left out. The register names an apskritis and a savivaldybė only in the genitive, `Vilniaus` and `Vilniaus miesto`. A genitive alone doesn't name a place, so each is named with its kind, `Vilniaus apskritis` and `Vilniaus miesto savivaldybė`. A locality keeps its nominative name, and a street its genitive name and the short form of its kind, `Laisvės pr.`, as an address writes them. Kaimai share names, so a locality is selected by its code.

An address alone would weigh a block of flats as one house: Vilnius holds 7% of the addresses and 18% once each of the premises counts, so an address is one and each of the premises at it one more. House numbers are counted per address. A korpusas is no part of the record's number and is dropped.

An address in a small kaimas is a number with no street. `geo.LT.address` keeps the one layout of a street and a number, so a locality with no numbered address on a street is dropped, as in `geo.EE`. 7,066 localities remain, holding 97% of the addresses, and every municipality keeps one.

The Universal Postal Union's guide to Lithuanian addresses, of June 2018, puts five digits left of the locality's name, preceded by `LT-`. So the layout is the one most countries share, `{street} {street-number}\n{postal-code} {locality}`, and a postal code is stored with its prefix, `LT-04339`, as the register writes it. The guide writes a locality that is not its municipality's seat on a line of its own and the municipality after the postal code; `geo.LT.address` prints the one layout and holds the municipality in a column.

The register's CSV files hold no point. Its boundary files do, as polygons, and reading them is left until a point is asked for: a Lithuanian locality has no `lat` or `lon`.

Valid while the register's open data is CC BY 4.0 and a locale's `address` prints one layout.

## `geo.EE`, `geo.LT` and `geo.LV` weigh a place by its people and a postal code or a street by its addresses, and the localities a source doesn't count share, by their addresses, their municipality's people that it counts under no locality

2026-11-20, in the chunk that moved the three modules' places from address weights to people. Serves goal 3.2. Applies the project principle on a guessed population, for four Estonian localities.

Each country's statistics office publishes its people by place as open data, so a place weighs its people rather than its addresses, which count a summer house as much as a home. The column is `population` in `region`, `municipality` and `locality`, as in every other `geo` module. A postal code and a street keep `addresses`: no office counts people by postal code or street, and a draw then lands where the letters go, as `geo.DK`'s postal codes and streets weigh.

`geo.EE` reads Statistics Estonia's table RL21003, the census of 31 December 2021, under CC BY-SA 4.0. The table keys a settlement by its EHAK code. A settlement whose code changed since is found by its name within its maakond, and Kiltsi, renamed too, by a table in the import. The municipalities' codes changed with the reform of 2017 and later mergers, so a municipality and a region sum the census counts of their settlements, those with no street included. The census withholds a count under three, which weighs 2. Keskküla, Kohtla-Uueküla, Koplimetsa küla and Raadi alev were formed after the census, which counts their people in the settlements they were cut from, so no area's count is left to share out: each weighs a language model's guess, which the page and DATA-LICENSES.md call an estimate. A locality with a street and no count stops the import.

`geo.LT` reads the State Data Agency's table M3010202, the residents of each savivaldybė at the start of 2026, keyed by the Address Register's codes. Lithuania's open data portal publishes it under CC BY 4.0, in dataset 2689. The table counts no locality, so every locality shares its municipality's residents by its addresses, as the project principle on a guessed population asks before any guess. The agency's table M3010320 counts each locality, but neither its response nor the agency's portal names a licence for it: the portal's page on its web service gives no terms, its footer reserves all rights, and the open data portal's dataset leaves the table out. So the project doesn't read it.

`geo.LV` reads the Central Statistical Bureau's experimental table RIG010, the people at the start of 2026, under CC BY 4.0, keyed by ATVK code and, for a ciems, by its code in the register. It writes a count of one to ten as 5. The bureau's table IRD081 counts the same places by another method, and the two are never mixed. Carnikava became a pilsēta in August 2026; the table still counts it as a ciems of its pagasts, so the import moves both counts to the codes in force.

In both, a municipality weighs its own count and a region the sum of its municipalities. A Latvian locality the table doesn't count, 26 of them, shares, by its addresses, the people of its municipality that the table counts under no locality. A Latvian pagasts also houses people outside every ciems, so its houses there take a share too. A locality weighs at least one.

Valid while each office publishes its table under an open licence. Revisit `geo.LT` once the Lithuanian agency publishes M3010320 under one.

## `geo.EE` takes its EHAK codes from settlement units whose licence was not found

2026-11-20, Lilleman auf Larv. Serves goal 3.2. An exception to "The project takes open licences".

The Land and Spatial Development Board's page for its settlement units calls them open data and asks for a credit with the data's date, but names no licence. The board's terms of use allow using and publishing its data free of charge, commercially or not. The maintainer chose to ship the EHAK codes and names read from the settlement units anyway, and `DATA-LICENSES.md` marks the licence as not found. Only `geo.EE`'s codes and names come from them; its addresses are CC0 1.0.

The maintainer confirmed this on 2026-11-20 before 0.7.0 was cut. The board is asked to name a licence only if a user needs one.

Valid while the board names no licence for the settlement units and its terms of use allow publishing them.

## The table of every shipped path is a page in `docs/` that the README links

2026-11-21, the maintainer, in the chunk that shipped the table of every shipped path. Serves goal 11, whose counter-signal is that the README grows faster than what ships.

The README stays short, so the table sits in `docs/shipped-paths.md`. All 2,019 paths in the README would take it from 179 KiB to about 390 KiB, and GitHub stops rendering a README at 500 KiB. The page is about 240 KiB, and each country added to the data would add about 20 KiB. `TestDataListingIsCurrent` fails if the page passes 500 KiB.

The maintainer asked that the README link the page clearly and sum up the locales and their most important paths, so a reader finds a path from the README, as goal 11's signal asks. The README's Every shipped path sums them up.

Valid while the full table would crowd the README. The test's message says what to do at the limit: split the page per module.

## A first or last name a language model recalled is listed `real`, as a name in use; only its weight is an estimate

2026-11-21, the agent running the release review of 0.8.0. The principles pass left the choice to the maintainer, and the agent took it without asking, since the project principle on name tables already points this way. The maintainer may reverse it. Serves goal 3, and applies the project principle on name tables with no open bulk file of counts.

Goal 3 calls a value real when it is a thing that exists. A recalled name is one the model gave as a first or last name in use in the country. The principle on such tables already treats the names as names and calls only their weights estimates, so `data-listing/kinds-and-sources.tsv` lists these tables as `real` and its source cell says no register checked them. Listing them `invented` would call names such as `Müller` made up, which misleads more than it warns.

Valid while the recalled tables hold names in use. A name a user reports as not in use is removed. A table with many such reports is listed `invented` until a register replaces it.

## `--null-text` spells a null in `csv` alone, and `csv` quotes a value equal to it

2026-11-25, the agent shipping item 556. Goal 1: a bulk fixture author loads the CSV into MySQL with no extra step.

MySQL's `LOAD DATA`, with its fields enclosed in quotes, reads an unquoted `NULL` as null and a quoted `"NULL"` as text. So `csv` writes a null as the flag's text and quotes any other value equal to that text, a number's too. A text `encoding/csv` would quote is refused, since a quoted field loads as text.

`tsv` is the file a table reads, and a table cell has no null, so the flag is refused there, as under every format but `csv`. A panel of three readers (a bulk fixture author, a coding agent and a hand fixture author) ranked six names. `--null-text` and `--null-as` tied on points; `--null-text` was never ranked below fourth and never read as a switch.

Valid while `csv` is the one record format with a free choice of null spelling.

## `pgcopy` escapes a NUL, and `csv`, `sql` and `tsv` write it as it is

2026-11-26, the agent shipping item 566. Goal 5.4: errors surface when the data loads, never mid-render.

PostgreSQL's text can't hold a NUL. `pgcopy` writes one as `\000`, which COPY decodes and refuses with an error naming the line; a raw NUL made `psql` cut the row short and load the next line glued to it. `sql` has no escape for a NUL, and a refusal there needs an error from a serializer that returns a string. The CLI could check each record, but that copies its columns and fails a run partway, which goal 5.4 rules out. A load can't refuse a NUL either: `json` carries one as `\u0000`, and one draw is written in every format. So `csv`, `sql` and `tsv` write a NUL as it is, and the README says what `psql` does with one.

Valid while a serializer returns a string and no shipped path draws a NUL.

## A locale's full numbers are `phone-mobile` and `phone-fixed`, and its prefix tables `phone-mobile-prefix` and `phone-fixed-prefix`

2026-11-26. Goal 1: a user fills a customer table with numbers of one kind by path, in any locale.

`phone-mobile` and `phone-fixed` named prefix tables in `da_DK`, `de_DE`, `fi_FI`, `fr_FR`, `lt_LT` and `nb_NO`, so item 558 named the full numbers `mobile` and `landline`. Its panel of five readers (a hand fixture author, a coding agent, a validator-facing author, a Python test author and a bulk fixture author) put `mobile` first (three), `mobile-number` (one) or `phone-mobile-number` (one), and each proposed renaming the tables.

A second panel of the same five readers chose between three outcomes: keep `mobile` and `landline`, so that `phone-mobile` still names a prefix table; draw the full numbers from both spellings; or draw them from `phone-mobile` and `phone-fixed` alone. The panel was told that a missing path's error names nearby paths. It doesn't: it reads `no entry "phone-mobile"`. Redone with the real text, four readers chose `phone-mobile` and `phone-fixed` alone, since the name a reader guesses should draw what it says and `mobile` and `landline` had not shipped. One reader chose both spellings. The tables are `phone-mobile-prefix` and `phone-fixed-prefix`, the one break this decision costs, and `CHANGELOG.md` says how to upgrade.

`en_US` can't tell the kinds apart, since mobile and fixed numbers share area codes, so `en_US.phone-fixed` reads `en_US.phone-mobile`. Each locale's `phone` keeps its columns and its shares, and gains no column for the new paths.

Valid while an error for a missing path suggests no nearby path. Item 220 ends it: with a suggestion, a guessed `phone-mobile` could fail usefully instead of drawing a number.

## The Swedish, Finnish, French and US name tables keep every name above a floor of bearers, set so a table holds at most 60,000 rows

2026-11-26, the agent shipping item 561. Goal 2 asks that enough data ships, goal 12.2 that the CLI prints the first value of any category in under 100 ms however much ships, and goal 3.3 that no value points at a person.

`sv_SE`, `fi_FI`, `fr_FR` and `en_US` kept the commonest 2,000 first names of each sex and 5,000 last names, so a counted name below that rank could never be drawn. Each import now keeps every name borne by at least a floor of people (`--first-min` and `--last-min`).

A name table has no block index, so its whole file loads on first reach, at about 0.5 µs a row (CPU time, on a machine with a load average of 16 to 25 on 16 CPUs, so a quiet one is faster). A `person` or `company` loads a first-name and a last-name table. With a floor of 5 on every table, `sv_SE.last-name` held 190,014 rows and `fr_FR.last-name` 218,571, and the first value of `sv_SE.person` took 88 to 143 ms, from about 10, as the stability review measured.

So each table takes the lowest floor of 5, 10, 20, 50, 100, 200 or 500 that leaves it at most 60,000 rows, which `TestNameTablesOfCountedBearersHoldUnder60000RowsEach` holds. `sv_SE.last-name` takes 50 instead of 20: `sv_SE.company` is the slowest first value of any locale, 65 ms before and 110 with 20, and 70 to 80 with 50. The last column is the share of the bearers a floor of 5 would keep; the 77.2% and the 81.5% are the price of goal 12.2, and the tables with 100% lose no bearer.

| Table | Floor | Rows | Bytes | Bearers kept |
|---|---|---|---|---|
| `sv_SE.first-name` | 5 | 47,889 | 576,148 | 100% |
| `sv_SE.last-name` | 50 | 16,929 | 197,819 | 77.2% |
| `fi_FI.first-name` | 5 | 17,128 | 214,143 | 100% |
| `fi_FI.last-name` | 5 | 23,264 | 283,349 | 100% |
| `fr_FR.first-name` | 5 | 33,372 | 430,988 | 100% |
| `fr_FR.last-name` | 200 | 54,262 | 660,637 | 81.5% |
| `en_US.first-name` | 50 | 52,170 | 695,237 | 99.7% |
| `en_US.last-name` | 500 | 41,202 | 504,514 | 90.4% |

On the same machine the first value of `sv_SE.person`, `fr_FR.person` and `en_US.person` takes 50 to 60 ms, from 10 to 20. No floor is below 5, since a name with fewer bearers can point at a household; SCB lists names from 2 bearers, and its 2 to 4 stay out.

A name borne under both sexes has a row for each, so a selector without a sex fails for the names that had one row before; the changelog lists this under Breaking. A row under the minority sex weighs its real count, so a man named Anna is as rare as the register says, and no share of a name's bearers is a floor of its own.

Valid while a name table loads whole. Item 465 gives a flat table fixed-size blocks, and then the floors are chosen again.

## A record writes its columns in the order its template lists its fields, a table writes them in the order of its header, and both draw them in name order

2026-11-26, the maintainer for the order, and the agent shipping it for a table, a key written twice and a struct. Goal 4: a reader predicts the columns from the template, and a `LOAD DATA` column list follows it.

Columns came in name order, because a JSON object decodes into a Go map, which keeps no order. `internal/jsonvalue` now decodes an object with its keys in the order written, and a record's columns follow them, in every format and in `Columns()`. No flag switches back to name order: the written order is the one a reader expects, and a second order is one more thing to learn.

- A record still draws its columns in name order. A seed then draws the same values wherever the template lists a field, and moving a field moves only its column.
- A table has no fields in its JSON, so its columns follow the header of its TSV, which is the order its author wrote.
- A key written twice in an object keeps its first place and its last value. The last value is what `encoding/json` kept before, and the first place is where JavaScript's `JSON.parse` puts it.
- A struct that `FakeStruct` fills keeps name order inside the engine. Nobody sees that order: each value lands in its Go field by name.
- `testdata/shipped_shape.txt` lists the columns of each shipped record, a table's included, in the order it writes them, so a change that moves one shows in the pin, and the README's Versioning table calls it a major.

Valid while a template is a JSON object and a record's columns are its fields.

## The Python package is a wheel per platform carrying the CLI binary, built by `release-tooling/python_wheels.py`, and `render` requires `seed` and `today`

2026-11-26, the agent shipping item 551; the maintainer chose, in the item, a wheel that carries the binary. Goal 1.3, with goals 9.1, 9.2 and 10.

The package is six wheels, one per platform GoReleaser builds, and pip installs the one for its platform. Each carries that platform's CLI binary, and `fejkdata.render` runs it through `subprocess` with the template on its stdin. Every template feature and all the shipped data then reach Python in the release that ships them, and the bytes are the CLI's own. `render` passes nine of the CLI's flags, so a flag reaches Python only once `render` takes it.

- `release-tooling/python_wheels.py` writes the wheels with the standard library, from GoReleaser's `dist/`. A wheel is a zip with three metadata files, and a build backend would be a tool to pin and update for no gain. The repository has no `pyproject.toml`, and there is no source distribution, since one could not carry a binary.
- A wheel's platform tag says where pip may install it. The Linux binaries are static, so they take the `manylinux_2_17`, `manylinux2014` and `musllinux_1_1` tags, as the `ziglang` package's wheels do. The macOS tag is `macosx_13_0`, the oldest macOS Go 1.27 runs on by its release notes (https://go.dev/doc/go1.27, Ports). A Go minor that raises that floor raises the tag with it.
- `Requires-Python` is 3.11, the oldest version https://devguide.python.org/versions/ listed as supported on 2026-11-26.
- `seed` and `today` are required keywords. Goal 9.1 has the caller supply everything that changes the output, as `New` requires `WithToday`.
- `template` is a string, as on the CLI's stdin. A dict would need a rule for which strings are JSON and which are format strings, and goal 4.7 has fejkdata not guess.
- A render the CLI refuses raises `FejkdataError`, with the CLI's stderr as its message and the exit code as `returncode`. Three readers, seated as a Python test author, a coding agent and a bulk fixture author, each picked `FejkdataError` over `RenderError`, `Error` and `TemplateError`, since `TemplateError` misreads exit 1. Each picked `returncode` over `exit_code` and `code`, since `subprocess` calls it that.
- `render` takes `table`, `data_path`, `no_shipped_data`, `separator` and `null_text` as keywords, each passing the CLI flag of its name; a string value goes as `--flag=value`, so one that starts with `-` stays a value. Three readers, seated as a Python test author, a data author and a coding agent, each picked `table`, `data_path`, `separator` and `null_text` over `into`, `data_paths`, `sep` and `null`, and each picked `data_path` taking one path or a list of them, strings or `pathlib.Path`, over a list only. Two of the three picked `no_shipped_data`, since it is the flag with underscores; the third wanted `shipped_data=False` to avoid a double negative, and the flag's own name won. A keyword left as `None`, or `no_shipped_data` left `False`, passes no flag, so the CLI's default stands.
- `render` doesn't re-implement any refusal of the CLI. The CLI ignores `table` outside `sql` and `pgcopy` and refuses `separator` and `null_text` outside their formats (`cli/flags.go`, `cli/records.go`), and `render` does the same, so each rule has one home. Two of the three readers asked for `table` to raise too, which needs a change in the CLI first.
- The binary sits in the package's `bin/` folder and is not put on `PATH`. The item asks for `render` alone, and a `fejkdata` command from pip beside one from Homebrew would be two versions under one name.

Valid while GoReleaser builds the CLI for the six platforms and the CLI reads its template on stdin.

## The npm package is the CLI built to WebAssembly and run through `cli.Run`, loaded once by `await load()`, with a synchronous `render` that requires `seed` and `today`

2026-11-27, the agent shipping item 549; the maintainer ruled on 2026-11-25, in the item, that one package, `@larvit/fejkdata`, carries all the data. Goal 1.2, with goals 2, 9.1, 9.2 and 10.

`npm-package/wasm` builds to one WebAssembly file whose single function runs the CLI once: `cli.Run` over the arguments, the template as stdin, and every shipped module. `index.js` passes `--seed`, `--today`, `--repeat` and `--format`, returns stdout, and throws a `FejkdataError` carrying stderr and the exit code. The header of a CSV and the framing of SQL are the CLI's, so calling the CLI is what makes the bytes the same, and every template feature reaches JavaScript in the release that ships it. `render` passes four of the CLI's flags, so a flag reaches JavaScript only once `render` takes it, as in the Python package.

- `cli.Run` is a new exported name. The CLI's entry over given streams was unexported, and a WebAssembly program has no process arguments or standard streams that `cli.Main` could read.
- A panel of three readers, seated as a Node test author, a browser author and a coding agent, picked the spellings on 2026-11-27. Two of three chose `const fejkdata = await load()` and a synchronous `fejkdata.render` over a `render` that returns a promise and over an import that waits for the WebAssembly. That import breaks `require` and stalls a browser page on the download. All three chose `today` as a string or a `Date` read by its UTC day, since a local day makes a laptop and CI disagree. All three chose `seed` as `number | bigint`, a number refused unless it is a safe integer. All three chose `exitCode` on the error, since Node's `error.code` is a string.
- A render keeps no state: each call builds a generator anew, which takes about 0.1 seconds in Node before the first row. A seed and a day then mean the same in every call, as goal 9.1 asks.
- The built file is 69 MB and 18.7 MB gzipped, measured on this commit. `load` took about 0.3 seconds in Node 26.13.1.
- One render returns at most 2^29 - 24 bytes, which the README calls 512 MB. That is V8's string limit in UTF-16 units, the shortest among the engines, and a text holds at least as many UTF-8 bytes as UTF-16 units. Past it the program's stdout refuses the write, so the CLI reports it and `render` throws with exit code 1. Without the cap, V8 threw while taking the result, the Go buffer stayed allocated, and a second such render ended the program.
- `index.js` drops the last line of a CLI error where it helps only at a shell: `try 'fejkdata --help'`, and the advice to wrap JSON in single quotes.
- `wasm_exec.js`, the loader Go ships, is copied from the Go that compiled the file, since the two must match. It sets `globalThis.Go`, and stubs `fs`, `process` and `path` where the runtime has none. The Go program can reach only globals, so it hands its function to `index.js` through `globalThis.__fejkdataRun`, which `index.js` deletes at once.
- Node's `fetch` refuses a `file:` URL, so Node and Bun read the file through `process.getBuiltinModule`, which no bundler follows. Node has it from 22.3.0 and 20.16.0 (https://nodejs.org/api/process.html#processgetbuiltinmoduleid), and `engines` names 22.3.0, since Node 20 is past its end of life.
- The gate runs `render.test.js` in the latest Node, in Node 22.3.0 and in Bun, each comparing renders with the Go binary built from the same commit. Deno imports the package over HTTP, so `load` takes the path a browser takes. No browser runs in the gate.
- `package.json` holds version `0.0.0`. Item 550 publishes the package, sets its version, and adds `LICENSE` and `DATA-LICENSES.md`, under the licence expression the Python wheels carry.

Valid while Go's WebAssembly needs `wasm_exec.js` and the built file loads in well under a second.

## The Python wheels go to PyPI from the `pypi` job of `test.yml` through trusted publishing, wheels only, under `MIT AND LicenseRef-fejkdata-data-licenses`

2026-11-26, the agent shipping item 552; the maintainer registered the project on PyPI and approves its public texts. Goals 1.1 and 1.3, with goal 10.

Each release's six wheels are uploaded by `pypa/gh-action-pypi-publish` in the `pypi` job of `.github/workflows/test.yml`, which runs in the GitHub environment `pypi` once the Packages workflow has passed.

- The job holds no token. PyPI trusts the repository, the workflow file `test.yml` and the environment `pypi`, and takes GitHub's short-lived token (https://docs.pypi.org/trusted-publishers/using-a-publisher/). PyPI "cannot currently" take a reusable workflow as the publisher (https://docs.pypi.org/trusted-publishers/troubleshooting/), and `packages.yml` is one, so the upload is a job of `test.yml`. A Packages run started by hand uploads no wheel.
- The action uploads an attestation with each wheel, its default under trusted publishing. PyPI requires two-factor login of every account. Only the publisher and the project's owner can upload.
- The job skips a file PyPI already holds, so a rerun after a half-finished upload passes. PyPI never takes a file name twice, so a bad version is yanked, not replaced (https://docs.pypi.org/project-management/yanking/).
- `License-Expression` is `MIT AND LicenseRef-fejkdata-data-licenses`. The code is MIT, and the binary holds data under some twenty licences, several with no SPDX id, such as a statistics office's terms of use. A list of SPDX ids would be false the day a table with a new licence ships. The core metadata specification allows a `LicenseRef-` id of the project's own, and `DATA-LICENSES.md`, a `License-File` of every wheel, says what it stands for. No `License ::` classifier is set, since Metadata 2.4 deprecates them.
- The classifiers name the status `4 - Beta` while the version is below 1.0, the three operating systems, and Python 3.11 to 3.15. A new Python adds its classifier by hand. `Requires-Python` has no upper bound, so pip installs the package on a newer Python before then.
- The project links use the labels https://packaging.python.org/en/latest/specifications/well-known-project-urls/ lists: Homepage, Documentation, Source, Changelog and Issues.
- The long description is `python-package/README.md`, with Python first. Its links are absolute, since pypi.org doesn't resolve a relative one.
- No source distribution ships, the platform tags and `Requires-Python` stay as the decision "The Python package is a wheel per platform carrying the CLI binary" records them, and pip still puts no `fejkdata` command on the path. The dry run installs a wheel and renders from it on glibc Linux, musl Linux, macOS and Windows, each on amd64 and arm64, with Python 3.11 on all but glibc amd64. No runner has macOS 13, so the `macosx_13_0` tag rests on Go's release notes.
- The dry run runs `twine check --strict` and lists each wheel's size and files. It can't upload, so the first release shows whether PyPI takes the wheels.
- The Packages workflow's output `wheels` says whether a release built wheels, which a tag cut before the Python package existed doesn't. The `pypi` job runs only where it is `true`, and fails where the wheels are then missing.
- PyPI allows 100 MB a file and 10 GB a project (https://docs.pypi.org/project-management/storage-limits/). Each release, a patch release too, adds six wheels of about 111 MB together, as the dry run of pull request #526 listed them, so the project fills after some 90 releases. Then ask PyPI for more room or delete the oldest versions.

Valid while PyPI's trusted publishing doesn't take a reusable workflow and the project stays under 10 GB.

## A US or French last name takes the spelling most of Wikidata's people with that family name write, where at least 5 do

2026-11-27, the agent shipping item 562. Goal 3.2: a good published source is preferred, and Wikidata's family-name items (CC0) are one for how a surname is written. The Census writes surnames in letters alone and INSEE without accents, so `Obrien` and `Lefevre` shipped though most of their bearers write `O'Brien` and `Lefèvre`.

`data-import/wikidata_names.py` takes each name's candidate spellings (a particle with its apostrophe, space or capital for the US, one accent for France), counts the distinct people of that country's citizenship whose family-name item carries each spelling as its label, in the language or `mul`, and under any kind of family name. A spelling replaces the name when at least 5 people have it and they are more than half of the people counted under all the forms tried. Otherwise the name stays as its register writes it, and the locale page says which names stay.

- A floor of 3 lets two or three people decide a spelling, so it is 5. Both importers share the floor, so a table never holds a name the importer would not give.
- A name with a period, `St. John`, ships: the rule asks how bearers write the name, and the Go check that shipped names are letters accepts a leading `St. `. The importers' `tsv.NAME`, which filters the registers' own names before this choice, still refuses a period.
- A count of Wikidata's people is a count of people with an article or an item, not of bearers. Only the order between spellings of one name is used, so a table's weights stay the register's counts.

Valid while Wikidata's family-name items carry the spellings and the registers keep writing names without them.
