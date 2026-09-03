# fejkdata

Locale-aware fake data for tests and fixtures, generated from JSON templates. Use
it from Go or the CLI — no data on disk, no dependencies, and a seed makes output
reproducible.

```sh
go install gitea.larvit.se/larvit/fejkdata/cmd/fejkdata@latest
fejkdata sv_SE.person          # Sara Eriksson
```

## CLI

```sh
fejkdata sv_SE.person                          # Sara Eriksson
fejkdata sv_SE.person.last                     # Eriksson
fejkdata --seed 42 sv_SE.address               # the same address every run
fejkdata -n 3 --separator ', ' sv_SE.word      # nät, barn, sol
fejkdata --list                                # every path the data offers
fejkdata --data-path ./mydata sv_SE.word       # layer a directory over the shipped data
fejkdata --no-shipped-data -d ./mydata --list  # only your data
fejkdata 'name: {/sv_SE.person.last}'          # name: <a surname> — an inline template
fejkdata '{"format":"name: {x}","x":["bosse","lina"]}'  # name: bosse or name: lina
```

A path names a category, or a field inside one: each dot segment descends one
level — folders, then the category (a JSON file), then fields. An argument that is
a JSON object or array, or that carries a `{` token, is instead an **inline
template**: a format string or a JSON value compiled and rendered on the spot. Its
tokens reach the data by reference from the root — `{/sv_SE.person.last}` (or
`{.name}`, which means the same here), so shipped and `--data-path` categories are
available — and `{..name}` is rejected, an inline template having no folder to step
up from. A path never contains a brace, so the two cannot collide (see
[Decisions](#decisions)).

| Flag | |
|------|--|
| `-d`, `--data-path D` | a directory to layer over the shipped data; repeatable, the last wins a name clash |
| `--no-shipped-data` | load only the `--data-path` directories |
| `-s`, `--seed N` | reproducible output |
| `-n`, `--repeat N` | render the value N times (up to 1048576), each an independent draw, streamed |
| `--separator S` | between repeated values (default a newline) |
| `--list` | print every path, then exit |
| `--version`, `-h`, `--help` | print, then exit |

`--name value` and `--name=value` both work, a short flag's value attaches or
follows (`-n3`, `-n 3`) and short flags bundle (`-hn 3`) — see
[Decisions](#decisions); flags go anywhere, `--` ends them. Exit codes: `0` success, `1` runtime error (missing
dir, unknown path), `2` misuse. From a checkout: `go run ./cmd/fejkdata …`.

### Your own data

A category is a JSON file in a directory. Save this as `mydata/sql.json`:

```json
{
  "format": "INSERT INTO users VALUES({sql-username});",
  "sql-username": {
    "format": "'{username}'",
    "repeat": 3,
    "separator": "),(",
    "username": ["pixelfox", "snork", "turbohund", "blip", "zoom", "wahoo"]
  }
}
```

```sh
fejkdata --seed 1 --data-path ./mydata sql
# INSERT INTO users VALUES('zoom'),('wahoo'),('blip');
fejkdata --repeat 100 --data-path ./mydata sql > seed.sql
```

## Library

```sh
go get gitea.larvit.se/larvit/fejkdata   # Go 1.22+
```

```go
f, err := fejkdata.New(fejkdata.WithSeed(42))
if err != nil {
	log.Fatal(err)
}
v, err := f.Fake("sv_SE.address") // "Kungsvägen 68\n379 17 Stockholm"
paths := f.List()                  // every path Fake accepts, sorted
v, err = f.FakeTemplate("name: {/sv_SE.person.last}")      // compile + render in one call
t, err := f.NewTemplate(`{"format":"name: {x}","x":["bosse","lina"]}`) // compile once
v = t.Fake()                                              // render many times, no re-parse
```

| Option | |
|--------|--|
| `WithSeed(n)` | same seed, same data: identical sequence |
| `WithDataPath(dir)` | layer a directory; repeat to layer several, the last wins a clash |
| `WithDataFS(fsys)` | layer an `fs.FS`, such as your own `embed.FS` |
| `WithoutShippedData()` | load only what you give |

A `*Generator` is safe for concurrent use; a seeded sequence is reproducible only
when drawn from one goroutine. Changing how a value is composed shifts the seeded
stream for that value and everything drawn after it.

## Data

The shipped set under [`data/`](data) — one folder per locale (`en_US`, `sv_SE`)
plus a locale-neutral `misc` folder — is embedded, so the CLI and the library
work with no data on disk. A directory is a namespace: each JSON file is a
category named after the file, each subdirectory a dot-path segment, so
`mydata/sv_SE/person.json` is `sv_SE.person` and replaces the shipped one.
Sources merge in order; matching folders combine, any other clash is won by the
last loaded. Names may not use `.`, `|`, `(`, `{`, `}`, `[`, `]` or `/`; dot-prefixed
entries are skipped, so a data directory can also be a checkout.

Each locale carries `address`, `color`, `company`, `date`, `email`, `ip`,
`person`, `phone`, `price`, `sentence`, `ssn`, `time`, `url`, `username`,
`version` and `word`, formatted per locale. `misc` carries `car`, `coordinate`,
`country` (ISO 3166), `creditcard` (Luhn-valid), `currency` (ISO 4217), `emoji`,
`httpstatus`, `language` (ISO 639), `mac`, `mimetype`, `objectid`, `timezone`
(IANA), `useragent` and `uuid` (v4). Many carry sub-fields — `misc.currency.symbol`,
`misc.country.alpha2`, `misc.httpstatus.code` — which `--list` shows.

## Data format

Every value is a **node**, nestable without limit:

| Node | JSON | Renders |
|------|------|---------|
| string | `"Malmö"` | its text, with any `{…}` tokens expanded |
| choice | `["a", "b", …]` | one item, picked at random |
| template | `{"format": "…", …}` | its format, with `{name}` tokens rendering the named fields |

### Format string

Every character is literal except a `{…}` token:

| Token | Renders |
|-------|---------|
| `{name}` | the sibling field `name` |
| `{name.field}` | `field` of one draw of `name` ([Correlated fields](#correlated-fields)) |
| `{a\|b}` | one of the named fields, even odds |
| `{fn(args)}` | a builtin ([Functions](#functions)) |
| `{/path}`, `{.path}`, `{..path}` | a node reached from the data root, this file's folder, or the folder above ([References](#references)) |
| `{{`, `}}` | a literal `{` or `}` |

```json
"100 Main St, Apt {int(1,9)}{upper(1)} — tel {digits(3)}-{digits(4)}"
```

Renders e.g. `100 Main St, Apt 4B — tel 555-0199`. Random characters come from
the sample functions, so a phone pattern is `070-{digits(3)} {digits(2)} {digits(2)}`.
A lone `}` is a load error naming `}}`; the arms of `{a|b}` must differ, a
repeated arm being a second spelling of [weight](#weight).

### Weight

An item of a choice may carry a `weight` (default `1`) to skew its odds:

```json
[
  { "format": "070-{digits(3)} {digits(2)} {digits(2)}", "weight": 10 },
  "08-{digits(3)} {digits(2)} {digits(2)}",
  { "format": "010-{digits(3)} {digits(2)} {digits(2)}", "weight": 2 }
]
```

Renders `070-412 38 91` ten times as often as `08-…`. A string item is weighted by
writing it as `{ "format": "AB", "weight": 3 }`. Rejected at load: a weight that
is negative, non-numeric, `0` (never drawn — remove the item), `1` (the default)
or outside a choice, and a repeated item — `["a", "a", "b"]` is `[{ "format": "a", "weight": 2 }, "b"]`,
and the error says so.

### Repeat

A template may carry `repeat` (an integer above `1`) to render its format that
many times — each an independent draw — joined by `separator` (default `""`):

```json
{ "format": "{word}", "repeat": 3, "separator": " ", "word": ["foo", "bar", "baz"] }
```

Renders e.g. `bar foo baz`. Rejected at load: a `separator` without a `repeat`,
a `separator` of `""` (the default), and a `repeat` that multiplies to more than
1 048 576 renders along any path of nested repeats.

### Options and fields

`format`, `weight`, `repeat` and `separator` are the only options; **any other
key is a field** (see [Decisions](#decisions)). An object that does nothing a
string can't — only a `format` — is rejected naming the string, as is a one-item
choice naming its item.

### Functions

A `{name(args)}` token calls a builtin. Arguments are checked at `New`: a bad
count, range, country or expression fails fast; an integer is written plain
(`5`, not `+5` or `05`); bounds are finite; a sample that could only ever emit one
value (`int(5,5)`, `float(1,1,2)`) is rejected naming the text to write instead;
and a length, count or decimal place beyond a sane maximum is rejected, so a
fat-fingered `hex(2000000000)` never tries to allocate gigabytes. Every builtin draws only from the seed — a
time-based id takes its timestamp from the rng, not the clock — so seeded output
stays reproducible.

| Function | Kind | Renders |
|----------|------|---------|
| `{luhn()}` | derivation | Luhn check digit over the digits emitted so far in this expansion |
| `{mod11()}` | derivation | weighted mod-11 check char (weights 2–7 from the right); `X` for 10 |
| `{ean()}` | derivation | EAN-13 / UPC-A / ISBN-13 / GTIN check digit |
| `{digits(n)}` | sample | `n` digits 0–9 |
| `{upper(n)}`, `{lower(n)}` | sample | `n` letters A–Z, a–z |
| `{int(min,max)}` | sample | uniform integer in `[min, max]` |
| `{float(min,max,dp)}` | sample | number in `[min, max]` with `dp` decimals |
| `{hex(n)}` | sample | `n` lowercase hex digits |
| `{base64(n)}` | sample | `n` random bytes, base64 |
| `{uuid()}` | sample | UUID v7 (`misc.uuid` ships v4 as data) |
| `{ulid()}` | sample | ULID, 26 Crockford base32 chars |
| `{nanoid(n)}` | sample | URL-safe Nano ID, `n` chars |
| `{iban(CC)}` | sample | length- and mod-97-valid IBAN for BE, DE, DK, ES, FI, NO or SE |
| `{seq()}`, `{seq(name)}` | counter | next integer from 1 in this generator; `name` selects an independent counter |
| `{calc(expr)}`, `{calc(expr,dp)}` | computation | an arithmetic expression over sibling fields ([Computation](#computation)) |
| `{lowercase(x)}`, `{uppercase(x)}`, `{ascii(x)}` | transform | a field's value rewritten ([Transforms](#transforms)) |

A derivation reads what is to its left, so place it after its payload; the
buffer is per expansion, so a nested template keeps fixed parts out of the sum. A
Swedish personnummer is a Luhn checksum over the nine digits before it:

```json
{ "format": "{century}{core}", "century": ["19", "20"],
  "core": { "format": "{digits(2)}{mmdd}-{digits(3)}{luhn()}", "mmdd": ["0115", "0704", "1218"] } }
```

Renders e.g. `19811218-9876`. `{seq()}` spans `Fake` calls and `repeat`, resets
with a new generator, and is the natural primary key for the SQL example above.

### Computation

`{calc(expr)}` evaluates `+ - * /`, parentheses and unary minus over number
literals and sibling field names, each rendered then read as a number; a second
argument rounds to that many decimals. A hyphen is always subtraction, so a
hyphenated field can't be an operand.

```json
{ "format": "{net} x {qty} = {calc(net * qty, 2)}", "net": ["19.99", "5.00"], "qty": ["3", "7"] }
```

Renders e.g. `19.99 x 3 = 59.97`. An operand that can never be a number (`"abc"`,
or a choice of such) is rejected at load, as is a division by a constant zero
(`1/0`, or a fixed `"0"` field); an operand that sometimes is not a number yields
`NaN`, and a division by one that is not constant `Inf` — both print rather than
fail.

### Transforms

`{lowercase(x)}`, `{uppercase(x)}` and `{ascii(x)}` rewrite the value of `x` — a
field, a path or a `..path` — and nest. `ascii` folds Latin letters (`Åsa Öberg`
→ `Asa Oberg`) and drops any other non-ASCII rune. `x` is held
([One draw, one spelling](#one-draw-one-spelling)), so an email built from a name
matches the name beside it:

```json
{ "format": "{p.first} {p.last} <{lowercase(ascii(p.first))}.{lowercase(ascii(p.last))}@example.com>",
  "p": [
    { "format": "{first} {last}", "first": "Åsa", "last": "Öberg" },
    { "format": "{first} {last}", "first": "Bo", "last": "Ek" }
  ] }
```

Renders `Åsa Öberg <asa.oberg@example.com>` or `Bo Ek <bo.ek@example.com>`, never
a mix.

### References

A reference renders a node from elsewhere in the data — the path `Fake` takes,
across every loaded source — so one category borrows another. The sigil says
where the path starts, as in a filesystem: `{/en_US.person}` from the data root,
`{.username}` from the folder this file sits in, `{..username}` from the folder
above. `data/sv_SE/email.json` can therefore read its own locale's `username`
without naming `sv_SE`:

```json
"Hej, {/en_US.person}!"
```

Renders e.g. `Hej, Pat Smith!`. A reference into a category is held like a
[correlated](#correlated-fields) path — `{.person.first} {.person.last}` name one
person, `{lowercase(.person.first)}` reads that same draw, and `{.person.first}`
beside `{/sv_SE.person.last}` in `sv_SE` is one person too — while a bare
`{/misc.uuid} {/misc.uuid}` is two draws. Rejected at `New`: a path that is
unknown, names a folder, has no folder above, or reads a field not every variant
of a choice carries, and a reference that leads back to its own value, directly,
mutually or through a chain.

### Correlated fields

`{name.field}` reads a path into a sibling, which holds the sibling to one draw
([One draw, one spelling](#one-draw-one-spelling)), so several tokens read one
row — a locality and the postal code that really covers it:

```json
{ "format": "{street} {int(1,99)}\n{place.postal-code} {place.locality}",
  "street": ["Kungsgatan", "Storgatan"],
  "place": [
    { "format": "{locality}", "locality": "Stockholm", "postal-code": "1{digits(2)} {digits(2)}", "weight": 975 },
    { "format": "{locality}", "locality": "Tranås",    "postal-code": "573 {digits(2)}", "weight": 18 }
  ] }
```

Renders e.g. `Kungsgatan 35` / `176 99 Stockholm`, never a Stockholm code beside
Tranås; each row's `weight` says how often it appears. A path is held at every
level it passes through: `{p.geo.town.name} {p.geo.town.zip}` share the town.
Every variant of a choice on the path must carry the rest of it, so a row missing
a field is named at load:

```text
token {place.postal-code}: field "place": not every variant of this 2-way choice carries "postal-code"; all carry [locality]
```

The sub-fields stay addressable — `Fake("address.place.locality")` renders, and
`List` advertises it. A path may not read into a level carrying a `repeat`.

### One draw, one spelling

A name any token reads as a path (`{p.first}`) or as an operand (`{calc(net * 2)}`,
`{uppercase(w)}`) is drawn **once per expansion**, and every other route to it —
a bare `{p}`, a second bare `{w}`, `{/cat.net}`, a nested template rendering
`{/cat.p.last}`, at any depth — is a load error naming the spelling to use. A
name nothing reads that way is drawn each time: `{word} {word}` differs. An
expansion is one render of one format, so each `repeat` iteration and each nested
template draws again.

```text
token {p} renders a level that {p.first} reads a path into; name the fields you want instead
token {w} is repeated, and uppercase operand "w" holds "w" to one draw per expansion; write {w} once
```

### Performance

Each file is parsed, validated and weight-indexed once, in `New`. A `Fake` call
then costs about what its output costs: an unweighted pick is O(1) whatever the
list's length, a weighted one O(log n), and long formats, deep nesting and many
tokens add cost in proportion to the output.

## Goals

1. **Valid by construction** — every value passes the check its real consumer
   applies; facts that belong together come from one draw, within a value and
   across categories.
2. **Text means what it says** — a format renders as written; only `{…}` varies,
   random characters included (`{digits(3)}`). One spelling per result; the wrong
   one is a load error naming the right one.
3. **Every mistake is a load error** — `New` rejects; `Fake` on a loaded generator
   fails only for an unknown path.
4. **Zero to a value in one command** — `go install`, then `fejkdata sv_SE.person`:
   no checkout, no flag. Flags are GNU-form (`--seed 42`, `-n 3`) in any position;
   the first custom template needs no escape and no option.
5. **Data lives in JSON** — a builtin only for what data can't express.
6. **Reproducible** — seed in, same stream out; no builtin reads a clock.
7. **Zero dependencies** — standard library only.
8. **Docs index the grammar** — every syntax feature is a heading; every example
   runs under test and shows its output; a rule is stated once.

## Decisions

- **Options and fields share one namespace.** `format`, `weight`, `repeat` and
  `separator` are reserved; every other key is a field. Nesting fields under a
  key, or prefixing options, would tax every template to guard against a
  misspelt option.
- **`{a|b}` stays beside nested choices.** `[[…], […]]` picks the same way, but
  its arms are anonymous; `{femalefirst|malefirst}` keeps `person.femalefirst`
  addressable.
- **Flags follow getopt_long.** `--name value` and `--name=value` both work; a
  short flag's value attaches or follows (`-s42`, `-s 42`) and short flags bundle
  (`-hn 3`), as every shell user expects. A single-dash long flag is rejected
  naming the double-dash spelling, and `-s=42` is rejected naming both short
  spellings: `=` belongs to the long form, and reading `=42` as the value would
  make `-d=./x` a directory named `=./x`.
- **An argument is a template by its shape, not by a flag.** A JSON object or
  array, or a string carrying a `{` token, is an inline template; anything else is
  a path. A name may not contain a brace or a bracket, so a path can never collide
  with either spelling, and the `[` of a JSON array is gated on valid JSON so a
  stray copied bracket never swallows an argument. No `--template` flag is needed.
  Reserving both brackets — though only a leading `[` could collide — keeps one
  simple name rule instead of a leading-position special case.
- **The shipped data is embedded, not discovered.** A directory a machine happens
  to have would make `--seed 42` machine-dependent. Data still lives in `data/`
  as JSON; `--data-path` layers over it.
- **A bare reference draws each time; a reference path is held.** `{/p} {/p}`
  is two draws, as `{word} {word}` is, while `{/p.first}` beside a nested template
  rendering `{/p.first}` is a load error: a bare token is by contract an
  independent draw, a path pins its level, and any route into a pinned level from
  another expansion could show another row.
- **Reference sigils follow the filesystem.** `/` is the root, `.` this file's
  folder, `..` the folder above — what those spellings already mean to anyone who
  has typed a path. A locale's files reach each other without naming the locale,
  so a folder renames and copies without editing its references.
- **After the first tag, a new fence is a major version.** Data files are the
  public API, and one spelling per result grows by tightening, so every fence
  invalidates some file. Each such release names the rejected spelling and its
  replacement in the changelog and in the load error, and that is the whole
  migration: a fence rejects one spelling with one replacement, so the fix is
  local to each site. A fence that would need a non-local rewrite ships a
  converter with its release instead. Before the first tag there is no
  compatibility promise.
- **A `--data-path` override rebinds every reference to the category it
  replaces.** References bind against the merged tree, so once shipped data uses
  `{.person}`, a consumer's `sv_SE/person.json` is what every shipped reference
  into `person` reads, and `New` fails on shipped data the consumer never wrote
  when that file lacks a field those references read. Accepted: overriding is the
  point of layering, the error names the reference and the field, and the fix is
  the consumer's file carrying the fields the shipped tree reads.
- **The repeat cap bounds renders, not bytes.** A repeat, alone or nested, may
  ask for at most 1 048 576 renders; how large each render is stays what the data
  asked for, so `{hex(1048576)}` repeated to the cap is a terabyte, loaded without
  complaint. A byte estimate would need every builtin to declare a width to fence
  a shape no data comes near, and the harm lands on the author who wrote it.
- **64-bit targets only.** The gate builds amd64, and the buffer sizing a render
  pre-computes (renders × bytes) assumes a 64-bit int; on a 32-bit target it could
  overflow and panic.
- **A constant zero divisor is a load error; a divisor that is not constant prints
  `Inf`.** `1/0` and a fixed `"0"` field are decidable, so they join the
  never-numeric operand as a load error; the fold stops where an operand varies,
  so `a/(b*c)` with `b` fixed at `0` and `c` varying loads and prints `Inf` every
  draw — catching it needs zero-absorbing algebra for a shape nobody writes.
- **In data, a default written out and a constant spelled as a sample are load
  errors.** `weight: 1`, `repeat: 1`, `separator: ""`, `int(5,5)`, `float(1,1,2)`,
  `+5` and `05` each spell what a shorter form already spells, so each is rejected
  naming that form. The CLI's numbers follow the shell instead: `--seed 007` and
  `--repeat +3` are 7 and 3, as every command line reads them.
- **Samples say what they emit, transforms what they do.** `{upper(2)}` is two
  letters, `{uppercase(x)}` is `x` upper-cased; one name for both would turn on
  whether the argument looks like a number.

## Development

Everything runs in Docker — **no local tooling beyond Docker is needed**.
Source is bind-mounted; build caches persist in the `gocache` volume.

```sh
docker compose run --rm test    # go test -race
docker compose run --rm cover   # tests with coverage
docker compose run --rm bench   # benchmarks
docker compose run --rm build   # compile the library
docker compose run --rm vet     # go vet
docker compose run --rm cyclo   # cyclomatic complexity over 14 (test files excluded)
docker compose run --rm dev     # interactive shell
```

Commands that rewrite source keep your file ownership when run with `--user`:

```sh
docker compose run --rm --user "$(id -u):$(id -g)" fmt   # gofmt -w .
docker compose run --rm --user "$(id -u):$(id -g)" tidy  # go mod tidy
```

Every pull request runs `docker build .` against both the latest and the lowest
supported Go, and must pass before it can be merged. That build is the whole
gate — vet, complexity, format check and tests — so run it locally before pushing:

```sh
docker build .                                  # latest
docker build --build-arg GO_VERSION=1.22.12 .   # lowest supported
GO_VERSION=1.22.12 docker compose run --rm test # the same tests, without the image build
```

## Layout

```
fejkdata.go     Generator, New, options, the embedded data set, List
node.go         the node model and JSON -> node compilation
path.go         the dotted-path walk, and proving a path resolves
render.go       Fake and the recursive renderer (choices, format strings, expansions)
template.go     the {token} grammar: scanning, tokens, operands, validation, compiling a format
hold.go         the hold: one draw per expansion for paths and operands, and its fences
reference.go    reference sigils, and binding references across the tree
graph.go        the render graph: edges, cycles, the repeat bound, tree walks
builtins.go     the {name()} function registry and its implementations
calc.go         the {calc()} arithmetic evaluator: parser, eval, validation
data.go         data loading: fs.FS folders/files -> namespace tree, multi-source merge
cmd/fejkdata/   the fejkdata CLI
data/           shipped data (JSON), embedded at build: locale folders + a misc folder
```

## License

MIT — see [LICENSE](LICENSE). Forked from [github.com/Timewave-AB/fakes](https://github.com/Timewave-AB/fakes).
