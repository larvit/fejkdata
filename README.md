# fejkdata

Locale-aware fake data for tests and fixtures, generated from JSON templates. Use
it as a Go library or the CLI — no data on disk, no dependencies, and a seed makes output
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
a JSON object, array or string, or that carries a `{` token, is instead an
**inline template**: a format string or a JSON value compiled and rendered on the
spot. Its tokens reach the data by reference from the root —
`{/sv_SE.person.last}`, so shipped and `--data-path` categories are alike
available. An inline template sits in no folder, so the folder-relative `{.name}`
and `{..name}` are rejected naming the root spelling, and one reference alone —
`{/sv_SE.person}` — is the path written as a template, rejected naming the path. A
path never contains a brace, a bracket or a quote, so the two cannot collide (see
[Decisions](#decisions)).

| Flag | |
|------|--|
| `-d`, `--data-path D` | a directory to layer over the shipped data; repeatable, the last wins a name clash |
| `--no-shipped-data` | load only the `--data-path` directories |
| `-s`, `--seed N` | reproducible output |
| `-n`, `--repeat N` | render the value N times (up to 1048576), each an independent draw, streamed |
| `--separator S` | between repeated values (default a newline) |
| `--format F` | `text` (default), `json`, `ndjson`, `csv` or `sql` — a record's columns, one record per row (json frames them as an array) |
| `--table T` | the INSERT target for `--format sql` (default: the path's last segment, or `records` for an inline template) |
| `--list` | print every path, then exit |
| `--version`, `-h`, `--help` | print, then exit |

`--name value` and `--name=value` both work, a short flag's value attaches or
follows (`-n3`, `-n 3`) and short flags bundle (`-hn 3`) — see
[Decisions](#decisions); flags go anywhere, `--` ends them. Exit codes: `0` success, `1` runtime error (missing
dir, unknown path), `2` misuse — a bad flag, an argument that names neither a
template nor a path, or an inline template that does not compile. From a checkout:
`go run ./cmd/fejkdata …`.

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

That `format` string is the free-form spelling — you hand-write the whole row.
For structured output a record writes the row for you.

### Records

A record is a template seen as columns: its fields are the columns, its `format`
the whole. `--format json|ndjson|csv|sql` writes the records; the library's
`FakeRecord` (below) hands back the columns. A column is a string unless it declares a
[datatype](#datatype), and a [`null`](#null) item draws it as null. Save
`mydata/users.json`:

```json
{
  "format": "{first} {last}",
  "first": ["Ada", "Bo"],
  "last": ["Lovelace", "Ek"]
}
```

```sh
fejkdata --seed 1 --data-path ./mydata --format json users            # [{"first":"Bo","last":"Lovelace"}]
fejkdata --seed 1 --data-path ./mydata --format ndjson users          # {"first":"Bo","last":"Lovelace"}
fejkdata --seed 1 --data-path ./mydata --format csv  users            # first,last  →  Bo,Lovelace
fejkdata --seed 1 --data-path ./mydata --format sql  users            # INSERT INTO "users" ("first", "last") VALUES ('Bo', 'Lovelace');
fejkdata --seed 1 --data-path ./mydata --format sql --table people users # INSERT into another table
```

`--repeat` streams that many records — `json` frames them as one array document,
`ndjson` writes one object per line, `csv` a row after a header, `sql` one INSERT
per line. Only a category-level template is a record; a field, choice or folder
errors, and so does a `repeat` on
the template itself, which composes the format into one string rather than
projecting columns — ask for more records with `--repeat`. A `repeat` on a column
is fine.

A column carrying a newline keeps it inside the quoted CSV field or the SQL string
literal, so a row can span physical lines: read the stream with a CSV or SQL
parser rather than splitting it on newlines.

The SQL is ANSI — identifiers in double quotes, a literal quote doubled (`''`),
backslashes passed through — so a hyphenated field like `postal-code` stays a
valid identifier. PostgreSQL and SQLite take it as written; MySQL and MariaDB need
`ANSI_QUOTES` and `NO_BACKSLASH_ESCAPES` set first, or they read `"users"` as a
string and a backslash as an escape.

A record written only to emit columns still needs a `format` — the grammar's one
required key — so `"format": ""` carries the fields with an inert format: it
renders nothing by `Fake`, and is compiled only so the tree's fences still run.
The columns are the point, and their facts stay together: two columns that read a
path into one category — `{/currency.code}` and `{/currency.symbol}` — share one
draw of it, so the record is internally consistent. A bare `{/currency}` names no
field, so it keeps drawing on its own. That one draw is also why two
columns may not read overlapping reference *paths* — `{/cat.a}` beside
`{/cat.a.b}` is refused, naming the fields to write instead, as
[One draw, one spelling](#one-draw-one-spelling) refuses that pair inside a single
format. A column may not reference the record it belongs to by any spelling: `{/users.first}`
or a bare `{/users}` inside `users` describes a draw other than the columns beside
it, so put a value two columns share in its own category and reference that. A field hold, transform or
operand ties fields together within one column as always (see
[Correlated fields](#correlated-fields) and [Decisions](#decisions)).

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
r, err := f.FakeRecord("users")           // one record: each field a column
s := r.JSON()                             // {"first":"Ada","last":"Lovelace"}
r, err = f.FakeRecordTemplate(`{"format":"{x}","x":["a","b"]}`) // compile + render inline
err = f.FakeStruct(&user)                 // fill a struct's fake:"…" tagged fields
ok, err := fejkdata.IsTemplate(arg)       // an inline template by its shape, else a path
```

| Option | |
|--------|--|
| `WithSeed(n)` | same seed, same data: identical sequence |
| `WithDataPath(dir)` | layer a directory; repeat to layer several, the last wins a clash |
| `WithDataFS(fsys)` | layer an `fs.FS`, such as your own `embed.FS` |
| `WithoutShippedData()` | load only what you give |

A `*Record` carries its columns via `Columns()` — each a `Column` of `Name`,
`DataType`, rendered `Value` and `Null` — and serializes them with `JSON()`
(one object), `CSVHeader()`/`CSVLine()`, or `SQLInsert(table)` — the shapes the
CLI's `--format` writes. `FakeRecord` and `FakeRecordTemplate` take a record; a
path or template that is not one — a bare string, a choice, or a folder — errors.

```go
type User struct {
	ID   int64   `fake:"{seq()}"`
	Last string  `fake:"sv_SE.person.last"`
	Age  uint8   `fake:"{int(18,99)}"`
	Nick *string `fake:"[null, \"{/sv_SE.username}\"]"`
	Home Address // filled from Address's own tags
}
```

`FakeStruct` fills a struct through a pointer: each exported field tagged `fake:"…"` is
a column of one record, its tag a path or an inline template — told apart by
`IsTemplate`, as the CLI tells an argument — and its Go type the column's
[datatype](#datatype): a string, bool, integer or float kind, or a pointer to one,
which a [`null`](#null) item leaves nil. An integer stays within int64 whatever its
kind, and a value the kind cannot hold, such as `{int(0,300)}` in a `uint8`, is refused
naming a kind that holds it. The fields an embedded struct promotes are columns of the
same record; a named struct field, or a pointer to one, fills from its own tags as a
record of its own, so its references draw apart from its parent's, and `fake:"-"`
leaves it unfilled. Untagged fields keep their values, and so does a pointer back to a
struct already being filled. The first call for a type compiles its tags and reports
what they get wrong, with the same error on every later call; a `datatype` in a tag
names the Go type that already sets it.

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
last loaded. Names may not use `.`, `|`, `(`, `{`, `}`, `[`, `]`, `"` or `/`;
dot-prefixed entries are skipped, so a data directory can also be a checkout.

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

### Datatype

A record column may declare `datatype` — `integer`, `number` or `boolean` — so `json`
writes `42` rather than `"42"` and `sql` a bare literal; a column without one is a
string:

```json
{ "format": "",
  "id": { "format": "{seq()}", "datatype": "integer" },
  "paid": { "format": "{p}", "p": ["true", "false"], "datatype": "boolean" },
  "total": { "format": "{calc(net * qty, 2)}", "net": ["19.99", "5.00"], "qty": ["3", "7"], "datatype": "number" } }
```

Writes e.g. `{"id":1,"paid":true,"total":59.97}`. A column is a field of the top-level
template, or an item of a choice standing in for one; `datatype` anywhere else is a
load error. A typed column holds one value, alone in its format: a literal, one
`{int()}`, `{float()}`, `{seq()}` or `{calc()}` call, or a read that lands only on such
values. `integer` is an int64 written `0|-?[1-9][0-9]*` — `{float()}` prints one at
`0` decimals within int64 — `number` a JSON number, `boolean` `true` or `false`. A value its
datatype cannot hold is a load error naming it:

```text
order.id: datatype integer: {digits(3)} prints text, not an integer
order.id: datatype integer: "1{digits(2)}" is not one value; write one literal or one {int()}, {float()}, {seq()} or {calc()}, or read one
```

A typed column's `{calc()}` must be proven to print a number: each operand a number
literal, an `{int()}`, `{float()}`, `{seq()}` or `{digits()}` call, a calc, or a read of
such values, whose bounds keep every divisor from zero and the result within `1e300`.
What the bounds cannot show is refused — `{calc(a / b)}: divides by b, which is not
proven nonzero`. The calc fills an `integer` column at `0` decimals, or over whole
operands with no `/`, while its bounds stay within int64.

### Null

A `null` item draws a record column as null: `json` writes `null`, `sql` `NULL`, and
`csv` an empty field, with an empty string written `""` — the convention PostgreSQL's
`COPY … CSV` reads. A record of one null column is a blank line, which `COPY` reads as
null but most CSV readers skip, so write such a record as `json` or `sql`. `Fake`
renders a null as `""`. The other items' weights skew its odds:

```json
{ "format": "", "deleted_at": null, "middle": [null, { "format": "{n}", "n": ["Ann", "Eva"], "weight": 3 }] }
```

`deleted_at` is null every draw, `middle` a name three draws in four. Rejected at
load: `null` anywhere but a column, naming `""`, and a column whose items declare
different datatypes.

### Options and fields

`format`, `weight`, `repeat`, `separator` and `datatype` are the only options; **any
other key is a field** (see [Decisions](#decisions)). An object that does nothing a
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

Renders e.g. `19.99 x 3 = 59.97`. A result that rounds to zero prints unsigned — `0`,
`0.00` — as `{float()}`'s does. An operand that can never be a number (`"abc"`,
or a choice of such) is rejected at load, as is a division by a constant zero
(`1/0`, or a fixed `"0"` field); an operand that sometimes is not a number yields
`NaN`, and a division by one that is not constant `Inf` — both print rather than
fail, except in a [typed column](#datatype), which must prove neither happens.

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
[correlated](#correlated-fields) path — `{.person.femalefirst} {.person.last}` name
one person, `{lowercase(.person.femalefirst)}` reads that same draw, and
`{.person.femalefirst}` beside `{/sv_SE.person.last}` in `sv_SE` is one person too — while a bare
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
3. **Every mistake is a load error** — `New` rejects the data and `NewTemplate`
   the inline template; on a loaded generator `Fake` fails only for an unknown
   path, `FakeStruct` only for a non-struct argument or a type its tags do not
   describe, with the same error every call, and `Template.Fake` cannot fail at all.
4. **Zero to a value in one command** — `go install`, then `fejkdata sv_SE.person`:
   no checkout, no flag. Flags are GNU-form (`--seed 42`, `-n 3`) in any position;
   the first custom template needs no escape and no option.
5. **Data lives in JSON** — a builtin only for what data can't express.
6. **Reproducible** — seed in, same stream out; no builtin reads a clock.
7. **Zero dependencies** — standard library only.
8. **Docs index the grammar** — every syntax feature is a heading; every example
   runs under test and shows its output; a rule is stated once.
9. **Fast enough to be free** — a value renders in about a microsecond and `New`
   parses and validates the whole set once upfront, so generating fixtures stays
   noise against a test's own runtime.

## Decisions

- **Options and fields share one namespace.** `format`, `weight`, `repeat`,
  `separator` and `datatype` are reserved; every other key is a field. Nesting fields under a
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
- **An argument is a template by its shape, not by a flag.** A JSON object, array
  or string, or a string carrying a `{` token, is an inline template; anything else
  is a path. A name may not contain a brace, a bracket or a quote, so a path can
  never collide with any of those spellings, and the leading `[` or `"` is gated on
  valid JSON so a stray copied bracket never swallows an argument — it names
  nothing, and says so. No `--template` flag is needed. Reserving the characters
  whole — though only a leading one could collide — keeps one simple name rule
  instead of a leading-position special case. The JSON string is what makes the
  library's own advice reachable: the error for an object holding only a format
  names `"…"`, and that spelling has to work where it is printed. One reference alone,
  `{/users}`, is refused naming the path `users`: both render the same text, and only
  the path names a record. `IsTemplate` exports the rule, so the CLI, struct tags and
  any other caller read one.
- **An inline template skips the cycle fence, and only that one.** `New` proves the
  loaded tree acyclic, an inline node is a finite tree of its own, and nothing in
  the tree can reference it, so no render of it reaches itself. Every other fence
  runs over both, from one `checkScope`.
- **An inline template that does not compile is misuse (exit 2), including a
  reference that resolves to nothing** — the whole argument is the spelling under
  test, and `NewTemplate` compiles, links and validates as one step. An unknown
  *path* stays a runtime error (exit 1): there the argument is well-formed and only
  the data is absent.
- **A padded JSON argument is rejected, not trimmed.** Padding is the one place the
  two readings disagree — a format string renders it, JSON drops it — so the
  spelling that renders is named rather than silently chosen.
- **`FakeTemplate` and `NewTemplate` both stay.** They reach the same value but not
  at the same cost: `NewTemplate` pays the compile and validation once and renders
  many times, `FakeTemplate` is the one-shot call, and `--repeat` is exactly the
  case that needs the first. The pair is `regexp.MustCompile` and `regexp.Match`,
  not two spellings of one result.
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
- **A constant zero divisor is a load error; in a string column a divisor that is not
  constant prints `Inf`.** `1/0` and a fixed `"0"` field are decidable, so they join
  the never-numeric operand as a load error; the fold stops where an operand varies,
  so `a/(b*c)` with `b` fixed at `0` and `c` varying loads and prints `Inf` every
  draw — catching it needs zero-absorbing algebra for a shape nobody writes. A
  [typed column](#datatype) bounds its operands instead and refuses a divisor it
  cannot keep from zero.
- **In data, a default written out and a constant spelled as a sample are load
  errors.** `weight: 1`, `repeat: 1`, `separator: ""`, `datatype: "string"`,
  `int(5,5)`, `float(1,1,2)`,
  `+5` and `05` each spell what a shorter form already spells, so each is rejected
  naming that form. The CLI's numbers follow the shell instead: `--seed 007` and
  `--repeat +3` are 7 and 3, as every command line reads them.
- **Samples say what they emit, transforms what they do.** `{upper(2)}` is two
  letters, `{uppercase(x)}` is `x` upper-cased; one name for both would turn on
  whether the argument looks like a number.
- **A record is a template seen as columns; a Go struct is the one second schema.** A
  template's `format` composes its fields into one string; `FakeRecord` and
  `--format` project the same fields as columns. Two views of one dataset, so a
  record author writes the same JSON they already know, and a column is the same
  field `Fake` renders by dotted path. The `format` is inert to a record — a
  record-only template writes `"format": ""` — but it is compiled and fenced, so
  a template that loads renders as whichever shape is asked for. `FakeStruct` takes
  its columns from a struct instead, because a Go caller has already written that
  schema: the fields name the columns and their types are the datatypes, so a tag
  says only what to draw, and a `datatype` in it would be a second spelling of the
  type.
- **A struct's records follow Go's field access, and compile on first use.** The
  fields an embedded struct promotes are the struct's own — `e.First`, as
  `encoding/json` and SQL mappers read them — so they are columns of its record and
  share its draws; a tagged field that another field hides is refused, not dropped. A
  named struct field is another entity and a record of its own; `fake:"-"` leaves it
  unfilled, whatever a category named `-` holds, and a pointer back to a struct
  already being filled is left alone, since filling it would never end. `New` cannot
  see a caller's types, so the first `FakeStruct` for a type compiles its tags and the
  answer, error included, is kept per type: a test's first call is its load, and no
  `NewStruct` handle is needed, as the cache already compiles once.
- **A record shares one reference draw per category.** Two columns that reference
  one category — `{/currency.code}` beside `{/currency.symbol}` — read one draw of
  it, so a record's facts agree the way a template's [correlated
  fields](#correlated-fields) do. The draw is one per record, so it spans a
  column's `repeat` and nested templates too (one record is one coherent unit);
  a bare reference — `{/currency}`, no field — stays an independent draw every
  time, the rule a format string already follows. Only references share: a sibling
  field is local to its own column, so a `first` column does not silently bind to
  a `first` in the column next to it.

  The string view of that same template does not share. `Fake` renders each
  sibling field as its own expansion, so a `{/currency.code}` field beside a
  `{/currency.symbol}` field is two draws and may render `EUR $`; writing both
  references in one `format` holds them together, as
  [One draw, one spelling](#one-draw-one-spelling) says. The scope is what makes a
  row coherent when the columns *are* the output, and there the caller cannot fall
  back on one format string. Widening it to every render would change what `Fake`
  has emitted since the start, for a correlation a single format already reaches.
- **A record's column set is fixed before the first draw.** Only a category-level
  template is a record: a path descending into a field, or naming a folder or a
  choice, errors. A tail may pass through a choice whose variants carry different
  fields, so the columns — and with them the CSV header written once ahead of every
  row — would vary per draw. A fixed column set is what the CSV and `INSERT`
  contracts rest on, so the restriction holds even where a particular choice would
  happen to agree.
- **Null is a `null` item, not a rate.** A null is one more outcome of a column's
  draw, so a choice's weights skew it like any other; a null-rate option would be a
  second way to state odds.
- **A typed column holds one value, not composed text.** Its bounds come from a
  literal or a call's arguments, so a load error names a real value, a range check is
  one comparison, and `1{digits(2)}` is a second spelling of `{int(100,199)}`.
- **A typed column's calc is refused unless proven.** Operand bounds must keep each
  divisor from zero and the result finite; what they cannot show is refused rather
  than trusted, since a bare `NaN` breaks the JSON and SQL it lands in.
- **`Column` carries text, not a Go value.** `Value` is the rendered string beside
  `DataType` and `Null`, which each serializer writes as the load check proved it; a
  `Value any` would hand every caller a type switch.
- **The package stays flat.** Go ties a package to one directory, so folders would
  split the API into packages.
- **The performance gate asserts allocations, not wall-clock time.** `AllocsPerRun`
  is deterministic across machines, so a ±10% ceiling does not flake under CI load,
  while time varies with the machine and its neighbours. A rendering slowdown
  almost always costs an allocation too (a lost pre-size, a per-item map, an extra
  copy). The benchmark suite (see Development) reports time for a human, not as a
  pass/fail gate.

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
supported Go, and must pass before it can be merged — unless it changes none of
the files the build and its tests read, nor the workflow itself, in which case
it's skipped (see [Decisions](#decisions)). That build is the whole gate — vet,
complexity, format check and tests — so run it locally before pushing:

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
record.go       records: Record, the JSON/CSV/SQL serializers, and their entry points
struct.go       structs: FakeStruct, fake tags, and a field's Go type as its column's datatype
inline.go       inline templates: Template, NewTemplate, FakeTemplate, IsTemplate, and their compile and link
template.go     the {token} grammar: scanning, tokens, operands, validation, compiling a format
hold.go         the hold: one draw per expansion for paths and operands, and its fences
reference.go    reference sigils, and binding references across the tree
graph.go        the render graph: edges, cycles, the repeat bound, tree walks
builtins.go     the {name()} function registry and its implementations
calc.go         the {calc()} arithmetic evaluator: parser, eval, validation
datatype.go     column datatypes: DataType, where datatype and null may sit, a column's datatype
value.go        the value proof: what a typed column or calc operand holds, checked at load
data.go         data loading: fs.FS folders/files -> namespace tree, multi-source merge
cmd/fejkdata/   the fejkdata CLI
data/           shipped data (JSON), embedded at build: locale folders + a misc folder
```

## License

MIT — see [LICENSE](LICENSE). Forked from [github.com/Timewave-AB/fakes](https://github.com/Timewave-AB/fakes).
