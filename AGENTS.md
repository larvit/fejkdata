# Rules

- Go runs only through `docker compose run --rm <test|vet|fmt|build|cyclo|generate>`; the merge gate is `docker build .` plus CI's changelog check.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output or of `testdata/shipped_shape.txt` is its own commit.
- A change under `data/` or to the shape pin adds its `CHANGELOG.md` entry under `Unreleased` in the same PR, and so does a change to a flag, an exit code, an exported name, a fence, a builtin or the lowest Go; what is major is the README's Versioning table.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- GitHub is where the project lives, so `gh` is the tool: merges are fast-forward
  only, `gh pr merge --rebase` where the branch is already on top of main.
- CodeRabbit skips every pull request until the repository has 10 GitHub stars: check
  its status once when a PR is marked ready, answer a review if one is there, and never
  wait for one.
- From 0.2.0, README goal 3 is gated at 7.0 on the `comprehension-panel` skill, and below it
  nothing else merges — no feature, no category, no data — bar a security fix, a
  dependency bump, a repair of behaviour the README documents, the infrastructure
  the round itself runs on, and a chunk the comprehension-floor decision lists as
  shipped under the floor. `todo.md` carries the round: a PR per item it plans, the nine-seat
  findings at depth 1 or a restructure the maintainer approved, then the nine-seat run
  again, until the score passes. At or above 7.0
  every pull request scores with the four-seat run and answers it in one run. No merge
  lowers the last score its own panel recorded: the nine-seat round and the four-seat
  run are two ratchets, never compared with each other. A lower score blocks the merge
  even where it may be panel noise: fix what the seats name until the score rises. A PR
  that only records a run's scores and files what its seats name merges even when a
  score fell. Through 0.1.0, a merge is blocked only by a paired ruling that rules worse.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- Prose naming a source states what that source states: read the register's own field
  before paraphrasing it, and name the register you read before claiming none
  publishes a spelling.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
- Cyclomatic complexity is gated at 14: split a function that would go over it. A table-shaped dispatch, a switch with one case per kind of its input, may sit at 14 and stays whole.

# Decisions

In [docs/decisions.md](docs/decisions.md):

- GitHub is canonical, and the module path names it
- The Gitea copy stays, as a pull mirror of GitHub
- The vocabulary sits below `doc.go`'s package clause, not in the package doc
- Options and fields share one namespace
- `{a|b}` stays beside nested choices
- Flags follow getopt_long
- A struct tag is a template by its shape
- The CLI renders exactly the template on its stdin, and adds nothing
- The CLI reads only the library's public API
- A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129
- Whitespace around a JSON object, array, string or null is dropped, and a bare number or boolean is a format string
- `FakeTemplate` and `NewTemplate` both stay
- The shipped data is embedded, not discovered
- The shipped data is Go modules a library imports by choice, and the CLI carries every one, from 0.1.0
- With only the shipped set, a category loads on the first call reaching it; beside a `--data-path`, every category loads in `New`
- Every `{…}` draws afresh, and only a name keeps a pick
- Reference sigils follow the filesystem
- A change to what exists is a major; a minor only adds
- An error is a contract by what it names, not its bytes
- A load reports every mistake at once
- Formats are checked after all data is loaded, so their errors name the category
- Raising the lowest supported Go is a major
- The format check runs on the latest Go only
- The changelog heading is the one spelling of a release; CI cuts one tag per module
- The modules develop in one committed `go.work`, and no published `go.mod` carries a `replace`
- A `--data-path` override rebinds every reference to the category it replaces
- The repeat cap bounds renders, not bytes
- 64-bit targets only
- A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`
- Samples say what they emit, transforms what they do
- A record is a template seen as columns; a Go struct is the one second schema
- A struct's records follow Go's field access, and compile on first use
- A category may reference itself, and a record's fences run at load
- A record's column set is fixed before the first draw
- Null is a `null` item, not a rate
- A typed column holds one value, not composed text
- A column that only reads one reference or name is the column it reads
- A typed column's calc is refused unless proven
- `Column` carries text, not a Go value
- A unit takes the stem of what it is, and a file the stem of the units it holds
- One name, one meaning
- The performance gate asserts allocations, not wall-clock time
- Rows live in a TSV, the shape in JSON
- A selector is bracketed, and a dot inside it is literal
- `parent` names the link column and the table alike
- After a row, a path names a column or a linked table
- A selected row is fixed, not drawn
- `..` steps up to the parent row, and a step down after it draws afresh, or once per name's pick
- A table selects by one key column, a code, never a free-form name
- A parent row with no child row is a load error
- A table is a record of string columns
- The key map is built at load, the rest on first draw
- Two categories may name one TSV
- A cell may read a reference, and each row its own
- A path is walked once without drawing before it is walked for real
- A country's postal codes and streets are siblings under its locality
- A locale's `address` reads its country's `geo` tree, so a locale folder is no data set on its own
- The default embed holds every Swedish postort the import can place and give a street-delivery code and a street, and the US places of 25,000 or more
- A locale's `address` binds its country's record once and prints it whole
- A postort's kommun comes from its name, its tätort or its codes, never from distance
- A highway designation is not a street, and a US postal code belongs to the place holding most of its land inside places
- `--list` stays a plain list of paths
- A layout is always quoted
- A title is a table under `sex`
- A table owns the spelling of a selector on it
- No builtin reads the clock, so a date is bounded by days, never by an age
- `misc` is what every locale shares
- A register's canonical spelling loses to the one its domain writes
- `misc.territory` is the spine, and a `misc` table naming a territory links to it
- A table whose register publishes no frequency draws evenly
- `misc.timezone` weighs a zone by the people living in it
- `misc.car` is one flat table, not a make linked to its models
- `misc.territory` names its sovereign in a column, and there is no `misc.country` table
- `misc.loglevel` is a table keyed by POSIX's keyword, carrying the code
- `misc.tld` keys carry the leading dot, where other tables key on a bare code
- `misc.tld` is a table of its own, and `misc.territory.tld` stays a column
- `misc.territory` carries a currency code, it does not link to `misc.currency`
- An extension may name two media types
- The Swedish ids draw Skatteverket's test series
- The US given names come from a mirror of the SSA file
- `List` advertises direct descents only
- A path draws through its compiled steps
- A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read
- comprehension floor: every dimension and the overall at 7.0 or above, from 0.2.0
- The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`
- Goal 5.1's `{…}` covers `{{`, `}}` and a lone `}`, so the goal names no escape
- Goal 7.2 holds while a first template can be written with no escape; a literal brace still takes `{{` or `}}`

# Scoring run

The project values for `~/.claude/skills/comprehension-panel/scoring-run.md`:

- `{language}`: Go
- `{kind}`: a Go library and CLI that renders fake data from JSON templates
- `{domain}`: fake test-data generation and template languages
- `{domain docs}`: the documentation of any fake-data library, such as Faker
- `{3am question}`: a seeded render of `sv_SE.address` prints the postal code of one postort beside a street of another
