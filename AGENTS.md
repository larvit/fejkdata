# Rules

- Go runs only through `docker compose run --rm <test|vet|fmt|build|cyclo>`; the merge gate is `docker build .` plus CI's changelog check.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output or of `testdata/shipped_shape.txt` is its own commit.
- A change under `data/` or to the shape pin adds its `CHANGELOG.md` entry under `Unreleased` in the same PR, and so does a change to a flag, an exit code, an exported name, a fence, a builtin or the lowest Go; what is major is the README's Versioning table.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- GitHub is where the project lives, so `gh` is the tool: merges are fast-forward
  only, `gh pr merge --rebase` where the branch is already on top of main.
- README goal 2 is gated at 7.0 on the `comprehension-panel` skill, and below it
  nothing else merges — no feature, no category, no data — bar a security fix, a
  dependency bump and the infrastructure the round itself runs on. `todo.md` carries the round: the nine-seat findings run at depth 1,
  a PR per item it names, then the run again, until the score passes. At or above 7.0
  every pull request scores with the four-seat run and answers it in one run. No merge
  lowers the last score.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- One spelling per result: reject the other at `New`, and let the error name the spelling to use.
- Prose naming a source states what that source states: read the register's own field
  before paraphrasing it, and name the register you read before claiming none
  publishes a spelling.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
- Cyclomatic complexity is gated at 14: the table-shaped dispatches (`linkParent`, `compileTemplate`, `renderEdges`) sit at it and stay whole; a function that would pass it is decomposed.

# Decisions

In [docs/decisions.md](docs/decisions.md):

- GitHub is canonical, and the module path names it
- The vocabulary sits below `doc.go`'s package clause, not in the package doc
- Options and fields share one namespace
- `{a|b}` stays beside nested choices
- Flags follow getopt_long
- An argument is a template by its shape, not by a flag
- An inline template skips the cycle fence
- An inline template that does not compile is misuse (exit 2), including a reference that resolves to nothing
- A padded JSON argument is rejected, not trimmed
- `FakeTemplate` and `NewTemplate` both stay
- The shipped data is embedded, not discovered
- A bare reference draws each time; a reference path is held
- Reference sigils follow the filesystem
- A change to what exists is a major; a minor only adds
- Seeded output is promised within one version
- An error is a contract by what it names, not its bytes
- A format holding a reference is fenced at link, and its errors name its path
- Raising the lowest supported Go is a major
- The format check runs on the latest Go only
- The changelog heading is the one spelling of a release; CI cuts the tag
- A `--data-path` override rebinds every reference to the category it replaces
- The repeat cap bounds renders, not bytes
- 64-bit targets only
- A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`
- In data, a default written out and a constant spelled as a sample are load errors
- Samples say what they emit, transforms what they do
- A record is a template seen as columns; a Go struct is the one second schema
- A struct's records follow Go's field access, and compile on first use
- A render shares one reference draw per category, per group
- A draw group name is local to its category
- The expansion hold and the render's draws are two fences
- A record makes its draw maps up front, a `Fake` on its first read
- A category never references itself, and a record's fences run at load
- A record's column set is fixed before the first draw
- Null is a `null` item, not a rate
- A typed column holds one value, not composed text
- A column of one reference alone is the column it reads
- A typed column's calc is refused unless proven
- `Column` carries text, not a Go value
- `hold` names what a draw is kept in, `draw` the draw itself
- One name, one meaning
- The package stays flat
- The performance gate asserts allocations, not wall-clock time
- Rows live in a TSV, the shape in JSON
- A selector is bracketed, and a dot inside it is literal
- `parent` names the link column and the table alike
- After a row, a path names a column or a linked table
- A table read into is pinned; a table read whole draws afresh
- Every reference path into one family selects the same rows, per render and group
- A parent row with no child row is a load error
- The choice-of-rows fence guards a data file's root, and requires string fields
- A table is a record of string columns
- The key index is built at load, the rest on first draw
- Two categories may name one TSV
- The rows of a table are alternatives
- A table never reaches its own family, by any route
- Tables carrying token cells stay small
- A path is walked once without drawing before it is walked for real
- A country's postal codes and streets are siblings under its locality
- A locale's `address` reads its country's `geo` tree, so the shipped set loads whole
- The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more
- A locale's `address` restates its country record's format
- A postort's kommun comes from its name, its tätort or its codes, never from distance
- A highway designation is not a street, and a US postal code belongs to the place holding most of its land inside places
- `--list` stays a plain list of paths
- A layout is always quoted
- A title is a table under `sex`
- A table owns the spelling of a selector on it
- No builtin reads the clock, so a date is bounded by days, never by an age
- A name column without a key resolves inside its parent
- `misc` is what every locale shares
- A register's canonical spelling loses to the one its domain writes
- `misc.territory` is the spine, and a `misc` table naming a territory links to it
- A table whose register publishes no frequency draws evenly
- `misc.timezone` weighs a zone by the people living in it
- `misc.port` selects by number, and carries no `name`
- `misc.car` is one flat table, not a make linked to its models
- `misc.territory` names its sovereign in a column, and there is no `misc.country` table
- `misc.loglevel` is a table keyed by the code, rendering POSIX's keyword
- `misc.tld` keys carry the leading dot, where other tables key on a bare code
- `misc.tld` is a table of its own, and `misc.territory.tld` stays a column
- `misc.territory` carries a currency code, it does not link to `misc.currency`
- An extension may name two media types
- The Swedish ids draw Skatteverket's test series
- The US given names come from a mirror of the SSA file
- `List` advertises direct descents only
- Each entry point to `drawCheck.reads` says what its caller gets
