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
fejkdata 'misc.territory[SE].capital'          # Stockholm — a table's row, selected by key or name
fejkdata 'geo.SE.locality[Lund].street'        # Fjelievägen — a linked table, drawn inside the row
fejkdata --data-path ./mydata sv_SE.word       # layer a directory over the shipped data
fejkdata --no-shipped-data -d ./mydata --list  # only your data
fejkdata 'name: {/sv_SE.person.last}'          # name: <a surname> — an inline template
fejkdata "{date(1990-01-01,2010-12-31,'2006-01-02')}"  # 2003-11-27 — the argument in "…", the layout in '…'
fejkdata '{"format":"name: {x}","x":["bosse","lina"]}'  # name: bosse or name: lina
```

A path names a category, or a field inside one: each dot segment descends one
level — folders, then the category (a JSON file), then fields — and `[SE]` after a
[table](#table) selects its row. An argument that is
a JSON object, array or string, or that carries a `{` token, is instead an
**inline template**: a format string or a JSON value compiled and rendered on the
spot. Its tokens reach the data by reference from the root —
`{/sv_SE.person.last}`, so shipped and `--data-path` categories are alike
available. A path never contains a brace or a quote, and a bracket only as a selector
after a name, so the two spellings cannot collide; which spellings an inline template
rejects, and what each names instead, is under [Decisions](#decisions).

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

Columns are written in name order, whatever order the fields appear in. A category
whose fields are the parts of one value — `sv_SE.price`, `sv_SE.version`, `misc.uuid` —
projects those parts rather than the value, so for one column holding what `Fake`
renders, write `{"format":"","price":"{/sv_SE.price}"}` as a fieldless category asks
for.

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
The columns are the point, and their facts stay together: a record is one render, so
columns that read a path into one category — `{/currency.code}` and
`{/currency.symbol}` — read one draw of it ([References](#references)), and a
[draw group](#draw-group) draws a column apart. That one draw is also why the columns of one
record may not overlap — `{/cat.a}`, or a bare `{/cat}`, beside
`{/cat.a.b}` is refused, naming the fields to write instead, as
[One draw, one spelling](#one-draw-one-spelling) refuses that pair inside a single
format. Both fences run at load, so a category that loads renders as either shape. A
category never references itself — `{/users.first}` or a bare `{/users}` inside `users`
describes a draw other than the fields beside it — so read a sibling as a path, and put
a value two fields share in its own category and reference that. A field hold, transform or
operand ties fields together within one column as always (see
[Correlated fields](#correlated-fields) and [Decisions](#decisions)).

## Data

The shipped set under [`data/`](data) — one folder per locale (`en_US`, `sv_SE`)
plus a locale-neutral `misc` folder — is embedded, so the CLI and the library
work with no data on disk. A directory is a namespace: each JSON file is a
category named after the file, each subdirectory a dot-path segment, so
`mydata/sv_SE/person.json` is `sv_SE.person` and replaces the shipped one.
Sources merge in order; matching folders combine, any other clash is won by the
last loaded. Names may not use `.`, `|`, `(`, `{`, `}`, `[`, `]`, `"` or `/`, nor be
`-`, which a struct tag reserves; dot-prefixed entries are skipped, so a data directory can also be a checkout.

Each locale carries `address`, `color`, `company`, `date`, `email`, `first-name`,
`ip`, `last-name`, `person`, `phone`, `price`, `sentence`, `sex`, `time`, `url`,
`username`, `version` and `word`, formatted per locale; `sv_SE` adds
`personnummer` and `samordningsnummer`, `en_US` adds `ssn` and `itin`. `misc`
carries `car`, `coordinate`, `creditcard` (Luhn-valid), `currency` (ISO 4217),
`datetime` (RFC 3339), `emoji`, `httpmethod`, `httpstatus`, `language` (ISO 639-1,
with its 639-2/T code), `mac`, `mimetype`, `objectid`, `port`, `protocol`,
`territory` (ISO 3166-1), `timezone` (IANA), `tld` (IANA root zone), `useragent` and
`uuid` (v4). Many carry sub-fields — `misc.currency.symbol`, `misc.territory.alpha2`,
`misc.httpstatus.code` — which `--list` shows. `car`, `currency`, `httpmethod`,
`httpstatus`, `language`, `mimetype`, `port`, `protocol`, `territory`, `timezone`,
`tld` and `useragent` are [tables](#table), so `misc.territory[SE].capital` and
`misc.currency[Euro].symbol` select a row; `car` and `useragent` carry no key or name,
so they are drawn from rather than selected in.

`misc.httpmethod`, `misc.protocol` and `misc.port` are IANA's registries.
`misc.httpmethod` is the eight methods RFC 9110 defines and PATCH; the register's
other entries, WebDAV and DeltaV among them, do not ship. Its `safe` and
`idempotent` are `true` or `false`, so a Go `bool` reads them, though `--format json`
and `sql` write them as text, a table column carrying no datatype. `misc.protocol`
carries a `number` and a `name`, the keyword itself where the register spells none out,
and selects by either spelling.
`misc.port` is the TCP assignments, rendering the number a port field holds with the
IANA `service` beside it, and selecting by number alone. A draw spans the whole
register, so pin `misc.port[443]` where a fixture needs a port a reader recognises;
it stops at 49150, IANA assigning nothing above, so an ephemeral source port is
`{int(49152,65535)}`.
`misc.tld` is the root zone's delegated TLDs, keyed with the leading dot
`misc.territory.tld` already carries, `.se`. Its `type` is the register's —
`country-code`, `generic`, `generic-restricted`, `infrastructure` or `sponsored` — and
its `unicode` the form the register displays, `.рф` for `.xn--p1ai` and the key itself
for an ASCII one, which selects the row too, so `misc.tld[.рф]` renders `.xn--p1ai`. A
draw spans the whole zone, `.arpa` and `.zuerich` alike, so pin `misc.tld[.com]` where
a fixture needs one a reader recognises.
[`DATA-LICENSES.md`](DATA-LICENSES.md) names each table's source and licence.

`misc.timezone` is every zone tzdb gives a shipped territory, from one apiece for most
of them to dozens for the largest, weighted by the population GeoNames records in each,
so `misc.territory[US].timezone` draws `America/New_York` far more often than
`America/Nome`. It links to `misc.territory`, so `misc.territory[SE].timezone` is
`Europe/Stockholm` and a drawn territory and zone agree. Its `offset` is the zone's
*standard* offset, so do not pair it with a drawn `misc.datetime`: in a zone that
observes DST it is the wrong one half the year, which is what storing a zone name
avoids. A zone tzdb named recently — `Europe/Kyiv`, `America/Ciudad_Juarez` — is
rejected outright by a consumer resolving it against older tzdata, so where the
consumer validates the zone, pin one rather than draw it. There is no `UTC` row, tzdb
giving that name no territory — spell it as the text `"UTC"`. `misc.useragent` carries
`browser`, `device` and `os` beside the string, and `misc.car` a `make` and a `model`,
until an international source replaces them.

ISO 3166-1 codes territories, not sovereign states, so that is what the table is
called: Greenland and Åland have codes of their own, and `misc.territory.country`
names the state each belongs to — `DK` for Greenland, `FI` for Åland, and its own
code for a sovereign one, or for one the register names no state for.

`sex`, `first-name` and `last-name` are tables weighted by bearers, from SCB, the
SSA and the Census Bureau. `first-name` links to `sex`, so `sv_SE.sex[f].first-name`
draws a woman's name, and a name both sexes carry is a row under each, so
`en_US.sex[m].first-name[Taylor]` names the one a `first-name[Taylor]` alone cannot.
`person` reads one draw of the three, so its `first` and `sex` columns agree, and so
does a `personnummer` in the same render: its birth number, `sv_SE.birth-number`
under `sex`, is Skatteverket's test series, 238 for a woman and 239 for a man, which no
real person is ever given. `en_US.title` links to `sex` too, so a person's prefix
never contradicts it.

A person of a chosen sex is assembled from the tables — `sex[f].first-name` beside
`last-name` — while a shipped `personnummer` agrees with the sex its own render
*drew*, not with one a path selects. That test series is also small: a personnummer
is one of about 70,000 values, a day in 1930–2025 against the two birth numbers, so a
fixture past a few hundred rows repeats one and a `UNIQUE` column needs a category of
your own. `sv_SE.date` and `en_US.date` are uniform over 1970-01-01 to 2029-12-31,
`misc.datetime` over 2000-01-01 to 2029-12-31. What the two locales do not share:
`en_US.address` carries a `region` column the Swedish one has no use for, and
`sv_SE.title` has no `parent`, so it is selected as `sv_SE.title[dr]` rather than
inside a sex.

A `geo` folder holds one tree per country under its alpha-2 code: five
[linked tables](#linked-tables) named alike, and an `address` record over one
consistent draw of them, which the locale's `address` reads.

| Table | `geo.SE` | `geo.US` | Weight |
|-------|----------|----------|--------|
| `region` | län, by code or name | state, by USPS abbreviation or name; `code` is the FIPS code | population |
| `municipality` | kommun, by code or name | county, by FIPS code or name | population |
| `locality` | postort, by name | incorporated place of 25,000 people or more with a postal code of its own, by GEOID or name; Hawaii has none | tätort population, the kommun's where the postort names it, else 200; place population |
| `postal-code` | postnummer with street delivery, by code | ZCTA, by code | one; address ranges |
| `street` | gatunamn, the ten with most road segments per postort | street name, the ten with most address ranges per place | segments; address ranges |

`geo.SE.region[Skåne län].municipality` draws a kommun in Skåne,
`geo.SE.locality[Lund].street` a street in Lund, and
`geo.US.region[IL].locality[Springfield]` settles which Springfield. A region row
carries its `timezone`, the state's predominant zone, and a locality its `lat` and
`lon`. What ports across countries is the five table names, the `name` column,
selection by name, and the `address` record's columns `street`, `street-number`,
`postal-code` and `locality`; every other column is the country's own, `code` on a
Swedish region but `abbr` on a US one.

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
record of its own, so its references draw apart from its parent's. `fake:"-"` leaves a
struct field, embedded or named, or a pointer to one, unfilled. Untagged fields keep
their values, and so does a pointer back to a struct already being filled; a type
whose fields reach more than 1024 structs is refused, naming `fake:"-"` to cut it. The
first call for a type compiles its tags and reports what they get wrong, with the same
error on every later call; a `datatype` in a tag names the Go type that already sets it.

A `*Generator` is safe for concurrent use; a seeded sequence is reproducible only
when drawn from one goroutine. Changing how a value is composed shifts the seeded
stream for that value and everything drawn after it.

## Data format

Every value is a **node**, nestable without limit:

| Node | JSON | Renders |
|------|------|---------|
| string | `"Malmö"` | its text, with any `{…}` tokens expanded |
| choice | `["a", "b", …]` | one item, picked at random |
| template | `{"format": "…", …}` | its format, with `{name}` tokens rendering the named fields |
| table | `{"format": "…", "rows": "x.tsv", …}` | its format over one row of the TSV beside it ([Table](#table)) |

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

A column of one reference alone to another record's column — `"score": "{/src.score}"`
— is that column: it takes the column's datatype and is null where the column is, and
a struct field tagged `src.score` is nil there. A `datatype` of its own types the
column's values where they prove it, and one restating the datatype it takes is refused.
Any other read renders the column's text, a null as `""`.

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
load: `null` anywhere but a column, naming `""`, and a column whose items hold
different datatypes.

### Table

A table is a category whose rows come from a TSV beside its JSON file: the header
names the columns, each line below it is one row, and the format renders the row
drawn. Save `mydata/country.tsv` and `mydata/country.json`:

```tsv
alpha2	name	population
DK	Denmark	5900000
NO	Norway	5500000
SE	Sweden	10500000
```

```json
{ "format": "{name} ({alpha2})", "rows": "country.tsv", "key": "alpha2", "name": "name", "weight": "population" }
```

```sh
fejkdata -d ./mydata country                    # Sweden (SE), about half the time
fejkdata -d ./mydata country.alpha2             # NO
fejkdata -d ./mydata 'country[SE]'              # Sweden (SE)
fejkdata -d ./mydata 'country[Norway].alpha2'   # NO
fejkdata -d ./mydata --format csv country       # alpha2,name,population  →  SE,Sweden,10500000
```

`rows` names the TSV beside the category file; `key` names the column a path
selects a row by, `name` a column it also selects by, `weight` a column of positive
numbers that skews the draw, and `parent` the table a column links to
([Linked tables](#linked-tables)). The format's `{tokens}` read the columns, and the
columns are the [record](#records)'s columns, so `--format csv` writes the rows and
`--list` shows `country.alpha2`. A cell is a string node: `1{digits(2)} {digits(2)}`
in a cell draws digits and `{/misc.uuid}` reads a reference, while `{name}` in a cell
is refused, since a cell has no sibling. Each cell may select its own row of another
table, `{/misc.currency[SEK].symbol}` on one row and `{/misc.currency[EUR].symbol}`
on the next, since only one row renders. `New` proves the header, the options and every
cell token, and refuses a TSV no category names, a key that is empty or repeats, a
weight that is not a positive number, and a key or name holding `[`, `]`, `{`, `}`,
`"` or `|`, which a selector cannot spell; the rows are indexed on the first draw that
selects one. A `name` needs a `key`, since a name naming several rows is reported by
their keys, or a `parent`, inside whose row a name names one row, so `first-name[Kim]`
is settled by the `sex` selected before it and a name repeating inside one parent row
is refused; a name spelling another row's key is refused, since the key would
select first and the name never. The table's options are its own — `rows`, `key`,
`name`, `weight` and `parent` — so a column may be named `name`, as one usually is.

A choice of templates sharing one format and one set of string fields is a table
written by hand, and `New` refuses it in a data file naming the TSV to write; an
inline template has no file beside it, so there it stays a choice.

### Row selection

`[key]` or `[name]` after a table's name selects one row: `misc.territory[SE]` and
`misc.territory[Sweden]` name one row, and `misc.territory[SE].capital` reads its column.
A name naming several rows is an error listing their keys, unless a row selected
before it settles which ([Linked tables](#linked-tables)). A selector is part of the path, so it works
wherever a path does: `Fake`, `FakeRecord`, a `{/misc.territory[SE].capital}` reference
and a struct tag. A dot inside the brackets belongs to the key or name, so
`city[St. Louis]` selects it. A path starts with a name, and `[` still opens a JSON
array at the start of a CLI argument, so `'[SE]'` alone names nothing.

### Linked tables

A table's `parent` names a column and, by the same name, the table beside it that the
column links to by key. Save `mydata/city.tsv` and `mydata/city.json` beside the
`country` table above:

```tsv
name	country	population
Copenhagen	DK	660000
Göteborg	SE	600000
Oslo	NO	710000
Stockholm	SE	990000
```

```json
{ "format": "{name}", "rows": "city.tsv", "key": "name", "parent": "country", "weight": "population" }
```

```sh
fejkdata -d ./mydata 'country[SE].city'      # Stockholm or Göteborg
fejkdata -d ./mydata country.city.name       # a country drawn, then a city inside it
fejkdata -d ./mydata 'city[Oslo].country'    # NO — the link column's cell
fejkdata -d ./mydata '{/city.name}, {/country.name}'   # Oslo, Norway — one consistent draw
```

A path descends from a row to a linked table by name, at any depth, and `--list`
advertises each direct step. Within one render and [draw group](#draw-group), linked
tables agree: the first table a reference path reads pins its ancestors, and a
descendant read after it is drawn inside them, so `{/city.name}` and
`{/country.name}` are a city and its country whichever is read first. A selected row
pins the render the same way, so every reference path into one family of linked
tables in one render and group selects the same rows: one that selects none beside
one that does is refused naming the spelling that does, `{/country[SE].city.name}`
beside `{/country[SE].name}`, and two selecting different rows are refused naming a
`drawGroup` to draw them apart in. A bare `{/city}` beside a path into its family is
refused too, since a bare reference draws each time, and so is a path that draws a
table another path in the group selects a row of, `{/city.name}` beside
`{/country[SE].name}`, which the first token rendered would otherwise decide. `New`
also refuses a link cell that is no key of the parent, a parent row no child links
to, a chain of parents that closes, a table named like a column of any table above
it, and a cell or format of a table that references a table of its own family, through
any template, a `repeat` or a `drawGroup` included, since a row rendered whole would
draw the family apart from itself: read the family from a template beside it, or add
the value as a column.

### Options and fields

`format`, `weight`, `repeat`, `separator`, `datatype` and `drawGroup` are the only options;
**any other key is a field** (see [Decisions](#decisions)), and `rows` makes a category
a [table](#table), so no template carries a field of that name. An object that does nothing a
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
| `{date(from,to,'layout')}` | sample | a second between two `YYYY-MM-DD` days, both included, in a quoted Go layout: `'2006-01-02'`, `'January 2, 2006'`, `'060102'`, `'2006-01-02T15:04:05Z'` |
| `{time('layout')}` | sample | a second within a day: `'15:04'`, `'3:04 PM'` |
| `{seq()}`, `{seq(name)}` | counter | next integer from 1 in this generator; `name` selects an independent counter |
| `{calc(expr)}`, `{calc(expr,dp)}` | computation | an arithmetic expression over sibling fields ([Computation](#computation)) |
| `{lowercase(x)}`, `{uppercase(x)}`, `{ascii(x)}` | transform | a field's value rewritten ([Transforms](#transforms)) |

A derivation reads what is to its left, so place it after its payload; the
buffer is per expansion, so a nested template keeps fixed parts out of the sum. A
Swedish personnummer is a Luhn checksum over the nine digits before it, six of
them a birthdate:

```json
{ "format": "{date(1930-01-01,2010-12-31,'060102')}-{birth}{luhn()}", "birth": ["238", "239"] }
```

Renders e.g. `811218-2389`. A layout is Go's: the reference time `Mon Jan 2
15:04:05 MST 2006` spelled as the output should look, quoted, since a layout may
carry the comma that separates arguments, with English names. Every second
between the two days is reachable, so a layout with a clock draws the time too, and
`from` may equal `to`, which is that one day. The instant is UTC, so a zone in the
layout prints `UTC` or `Z`.
The quotes delimit a layout outside a selector only, so `[O'Fallon]` in an
argument stays a name. Rejected at `New`: a bound that is no calendar date, or not
before the other; an unquoted layout, naming the single-quoted one; a layout naming no field, which is
text, as is one day in a layout with no clock; for `date` a layout naming no date
field, naming `time`; and for `time` a layout naming a date field, naming `date`. `{seq()}` spans `Fake`
calls and `repeat`, resets with a new generator, and is the natural primary key for
the SQL example above.

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

Renders e.g. `Hej, Pat Smith!`. A reference path into a category is held like a
[correlated](#correlated-fields) path, but for the whole render — one `Fake`, or one
record — rather than one format: `{.person.first} {.person.last}` name one
person, as do the same two references in sibling fields or a nested template, and
`{lowercase(.person.first)}` reads that same draw. Each `repeat` iteration is
a render of its own, in no group, so it draws anew, and a [draw group](#draw-group) holds a
draw apart. A bare reference names no field and makes its own picks each time —
`{/misc.uuid} {/misc.uuid}` is two draws — while the reference paths inside what it
renders still read the render's draws. Rejected at `New`: a path that is
unknown, names a folder, has no folder above, reads a field not every variant
of a choice carries, or names the category the reference sits in, and a reference
that leads back to its own value, directly, mutually or through a chain.

### Draw group

A template may carry `drawGroup` to hold its reference draws apart: every reference path
it renders, however deep short of a `repeat` or a nested `drawGroup`, reads the draw of
that group, and the templates of one category naming one group in one render read one
draw. A name is local to its category, so a category another one references never joins
its groups by name; the unnamed group spans them all.

```json
{ "format": "{payer} pays {payee}; signed {signature}",
  "payer": { "format": "{/sv_SE.person.first} {/sv_SE.person.last}", "drawGroup": "payer" },
  "payee": "{/sv_SE.person.first} {/sv_SE.person.last}",
  "signature": { "format": "{/sv_SE.person.last}", "drawGroup": "payer" } }
```

Renders e.g. `Sara Eriksson pays Ebba Lind; signed Eriksson`: the signature reads the
payer's draw, while the payee is drawn apart. Rejected at load, each naming nothing: a
`drawGroup` of `""` (the default); one naming the draw group its template already draws
in; one on a template that renders no reference path — a bare reference to a
[table](#table) counts, since the group answers for the family it draws in — short of a
`repeat` or a nested `drawGroup`, on a `repeat` itself — each iteration renders in no
draw group — or on an inline template's root, which nothing references. So is a path
reading into a level that carries a `drawGroup`.

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
`{uppercase(w)}`) is drawn **once per expansion**, a reference path (`{/cat.p.first}`)
**once per render** in its [draw group](#draw-group), and every other route to either — a bare
`{p}`, a second bare `{w}`, `{/cat.net}`, a nested template rendering `{/cat.p.last}`
beside `{p.first}`, a bare `{/cat}` beside `{/cat.p.first}`, at any
depth — is a load error naming the spelling to use. A name nothing reads that way is
drawn each time: `{word} {word}` differs. An expansion is one render of one format, so
each nested template draws its own names again; a render is one `Fake` or one record,
and each `repeat` iteration is an expansion and a render of its own.

```text
token {p} renders a level that {p.first} reads a path into; name the fields you want instead
token {w} is repeated, and uppercase operand "w" holds "w" to one draw per expansion; write {w} once
```

### Performance

Each file is parsed, validated and weight-indexed once, in `New`. Proving the draw
fences adds one pass over the loaded tree, and walks what a render reads only where
data binds a reference, so a set that binds none pays for the pass alone. A `Fake` call
then costs about what its output costs: an unweighted pick is O(1) whatever the
list's length, a weighted one O(log n), and long formats, deep nesting and many
tokens add cost in proportion to the output.

## Versioning

Semver tags on `main`, `v0.1.0` first; one version covers the shipped data, the
library and the CLI, and [`CHANGELOG.md`](CHANGELOG.md) names what each release
changed. A consumer's data, code and scripts keep working across a minor or a patch:
a minor only adds, and a major is the only release that changes what exists.

| Surface | Major | Minor |
|---------|-------|-------|
| Shipped data | remove or rename a path; change a category's format; remove a value, or change a weight or a repeat; add a reference from one shipped category into another; change a table's key, name, weight or parent column, or remove a row | a path outside a record's columns, a locale, a value in a list, a row |
| Records | remove, rename, retype or add a column; let a column be null | a record, as a new category |
| Data format | a fence: a spelling `New` rejects that it accepted; a template option, since it reserves a field name | a builtin |
| CLI | remove or rename a flag, or change its default; change what an exit code means; change the framing a `--format` writes (header, quoting, statement shape), the `--list` layout, or what an error names | a flag, a format |
| Library | change or remove an exported name; raise the lowest supported Go | an exported name, a `With…` option |

A patch changes no row of this table: performance, docs, or a fix inside a promised
behaviour that changes no value, path, format or spelling.

Seeded output is a promise within one version: same seed, same version, same
data, same output. Any release may shift a stream, since a value added to a list
moves every draw after it, so pin fixtures per version. An error's wording may
improve in a minor; the path, rejected spelling and replacement it names may not.

Before `v1.0.0` a minor is the breaking unit: `0.(x+1).0` may carry a major's
changes, each named in the changelog, and a `0.x.y` patch may not. `v1.0.0` is
cut once the shipped data is in its record shape and one full minor has shipped
with no breaking change. From `v2` the module path carries `/vN`, so fences ship
batched into as few majors as possible.

[`testdata/shipped_shape.txt`](testdata/shipped_shape.txt) pins every path, each
template category's format, the categories each category reads, each column's
datatype and nullability, and each table's key, name, weight and parent columns; a pull request that
changes it or `data/` adds its `CHANGELOG.md` entry, which CI checks. A removed,
renamed or retyped line is a major.

## Audience

App developers writing tests and fixtures, in Go and at a shell:

- a **bulk fixture author**, thousands of rows into CSV or SQL
- a **Go test author**, filling a struct with `FakeStruct`
- a **hand fixture author**, one value at a shell
- a **validator-facing author**, who needs a value a real checker accepts

and a **contributor**, who reads [`todo.md`](todo.md), [`AGENTS.md`](AGENTS.md) and
the Development section below, and who ships a register the four above then draw from.

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
10. **Data is sourced, or on its way there** — a shipped fact, a name, place,
    code, id or classification, is read from a register or open dataset by a
    [`data-import/`](data-import) script wherever one exists to read; where none
    does yet a small hand-written set ships and [`todo.md`](todo.md) carries the
    step that replaces it. Only non-factual copy stays authored. A sourced table
    holds the rows its source holds: none is added by hand, and one is dropped
    only by a rule the script states.
11. **Breadth follows what most systems store** — a category is added in proportion
    to how many real schemas hold it: names, addresses, phones, ids, money and
    timestamps before anything domain-specific, and a catalogue serving one niche
    waits behind everything serving many.
12. **Realism is the default, inertness is selectable** — where a value could reach
    something real, a domain anyone may register or an account a bank could issue,
    the realistic breadth ships *and* so does the subset that provably reaches
    nothing, each on its own path. A fixture that looks nothing like production
    tests nothing; the caller who needs a value that can touch nothing asks for it
    by name.

## Decisions

- **Options and fields share one namespace.** `format`, `weight`, `repeat`,
  `separator`, `datatype` and `drawGroup` are reserved; every other key is a field. Nesting fields under a
  key, or prefixing options, would tax every template to guard against a
  misspelt option.
- **`{a|b}` stays beside nested choices.** `[[…], […]]` picks the same way, but
  its arms are anonymous; `{female|male}` keeps `person.female` addressable.
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
  names `"…"`, and that spelling has to work where it is printed. An argument or struct
  tag of one reference alone, `{/users}`, is refused naming the path `users`: both
  render the same text, and only the path names a record. A folder-relative `{.name}`
  or `{..name}` is refused naming `{/name}`, since an inline template sits in no
  folder. `IsTemplate` exports the
  rule, so the CLI, struct tags and any other caller read one.
- **An inline template skips the cycle fence.** `New` proves the loaded tree
  acyclic, an inline node is a finite tree of its own, and nothing in the tree can
  reference it, so no render of it reaches itself. Every other fence runs over both,
  from one `checkScope`, except that struct tags leave column agreement to their Go
  types.
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
  is two draws, as `{word} {word}` is, while every `{/p.first}` in one render reads
  one draw, and a bare `{/p}` beside them is a load error: a bare token
  is by contract an independent draw, a path pins its level, and a fresh draw of a
  pinned level could show another row. A builtin's operand holds what it reads for
  its expansion, references included, so `{uppercase(/p)} {/p}` is one draw — the
  rule every operand follows.
- **Reference sigils follow the filesystem.** `/` is the root, `.` this file's
  folder, `..` the folder above — what those spellings already mean to anyone who
  has typed a path. A locale's files reach each other without naming the locale,
  so a folder renames and copies without editing its references.
- **A change to what exists is a major; a minor only adds.** Data files, the CLI
  and the Go API are the public API, and a consumer must be able to take a minor
  without an edit — so an added column is a major, since it changes the CSV header
  and the `INSERT` column list, as is a removed value, which changes what a fixture
  holds, and a new option, which reserves a field name. One spelling per result
  grows by tightening, so every fence invalidates some file. Each such release
  names the rejected spelling and its replacement in the changelog and in the load
  error, and that is the whole migration: a fence rejects one spelling with one
  replacement, so the fix is local to each site. A fence that would need a
  non-local rewrite ships a converter with its release instead. Before `v1.0.0` a
  minor carries what a major would.
- **Seeded output is promised within one version.** Any edit to a category shifts
  its stream and everything drawn after it, so a promise across versions would
  freeze every shipped list; a fixture is re-pinned on a bump, as this repo's own are.
- **An error is a contract by what it names, not its bytes.** A script branches on
  the exit code and reads the named path or spelling, so those hold; wording improves
  in a minor.
- **Raising the lowest supported Go is a major.** A consumer building on it breaks,
  which is the one test every rule above applies; Go's convention of a minor is not
  followed.
- **The changelog heading is the one spelling of a release; CI cuts the tag.** A
  tag pushed by hand is served by `go get` at once, so a tag whose commit lacks its
  heading is burnt, not fixed. The heading on a gate-passed `main` commit is the
  trigger instead: the tag can land only there, and the Gitea release the same job
  publishes keeps one text as its body and is where prebuilt binaries will attach.
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
  type. The Go type is a struct column's one datatype, so its items need not agree on
  one among themselves; each must only hold that type.
- **A struct's records follow Go's field access, and compile on first use.** The
  fields an embedded struct promotes are the struct's own — `e.First`, as
  `encoding/json` and SQL mappers read them — so they are columns of its record and
  share its draws; a tagged field that another field hides is refused, not dropped. A
  named struct field is another entity and a record of its own. `fake:"-"` leaves a
  struct field, embedded or named, unfilled, so no name may be `-`; a pointer back to
  a struct already being filled is left alone, since filling it would never end. `New` cannot
  see a caller's types, so the first `FakeStruct` for a type compiles its tags and the
  answer, error included, is kept per type: a test's first call is its load, and no
  `NewStruct` handle is needed, as the cache already compiles once.
- **A render shares one reference draw per category, per group.** Every reference
  path into a category in one `Fake`, or one record, reads one draw of it, so a
  value's facts agree across its fields, nested templates and columns alike —
  `{/currency.code}` in one field and `{/currency.symbol}` in another name one
  currency, whichever view renders them. A `repeat` iteration is a render of its
  own, since repeating asks for another entity, and a [draw group](#draw-group) names further
  entities within one render, so a payer and a payee are two groups over one
  `person` rather than two copies of it. Only references share: a sibling field is
  local to its own expansion, so a `first` column does not silently bind to a
  `first` in the column next to it.
- **A draw group name is local to its category.** A category's groups are its own
  entities, so a caller naming a group the same way never joins them by accident,
  and renaming a group inside one file changes no render elsewhere. The unnamed
  group still spans categories, since facts that belong together across categories
  must agree.
- **The expansion hold and the render's draws are two fences.** One proves a sibling
  path or an operand is reached only by its readers within an expansion, the other
  does the same for reference paths across a render and its draw groups. They pin
  different things — an operand pins the value its own render produced and stops at a
  reference, a path pins every level it passes through — so one walk would carry both
  rules and both scopes anyway, and tell them apart at every step.
- **A record makes its draw maps up front, a `Fake` on its first read.** A record's
  columns always read through the render's draws, so making the maps where the set is
  declared keeps them on that frame's stack. A `Fake` often reads no reference path at
  all, and making them anyway cost about a fifth of the cheapest render, so it makes
  them on the first read instead, at two heap allocations for a render that does share
  a draw. The allocation gate over a repeat of a reference path and over a named draw
  group prices that, and pins the two measures that keep a draw set off the heap.
- **A category never references itself, and a record's fences run at load.** A category
  is one unit: a reference back into it — `{/users.first}` inside `users` — describes a
  draw other than the fields beside it, so `New` refuses it and the sibling path stays
  the one spelling for a field of one's own. A value two fields share goes in its own
  category, which both reference. That settled, a record's column fences run at `New`
  too, so a category that loads renders as whichever shape is asked for, and a reference
  reaching back into a category through another one is refused there as the overlap it
  is. Which reads those fences weigh differs on purpose: a record-only template's inert
  format renders nothing, so a `drawGroup` on it can never matter and is refused, while
  one on a rendering format can matter to a caller that bare-references it.
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
- **A column of one reference alone is the column it reads.** `{/src.score}` renders
  exactly what `src.score` draws, so it takes that column's datatype and null rather
  than restating them, and a `datatype` restating the one it takes is a second
  spelling. Any other `datatype` still types the values — the one way to type a column
  someone else wrote.
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
- **Rows live in a TSV, the shape in JSON.** `New` allocates once per node, so a
  register of thirty thousand rows written as JSON objects would cost it a second;
  a TSV is one allocation whose cells are substrings, and the JSON says only how a
  row is composed. The TSV sits beside its category file, named by `rows`, so a
  data directory stays a directory of categories, and one nothing names is a
  load error rather than a file silently ignored.
- **A selector is bracketed, and a dot inside it is literal.** `municipality[0180]`
  reads as selection to anyone who has indexed an array, and `[St. Louis]` keeps a
  name whole where a colon or a dot-separated spelling could not; zsh needs the
  brackets quoted, which the README's examples show. A name that spells another
  row's key is refused at load rather than shadowed: a key names one row by
  contract, so the name could never select its own, and the check is one lookup per
  row against the key index that already exists.
- **`parent` names the link column and the table alike.** One word says both, so a
  child table sits in its parent's folder and links on a column of the parent's
  name; the geo plan wants exactly that, and a table needing another folder or
  another column name would be asking for a second spelling.
- **After a row, a path names a column or a linked table.** `locality[Lund].address`,
  a template beside the family under a selected row, is refused today; admitting it
  later is additive, since a refused spelling gains a meaning and no accepted one
  changes, so the door stays open for the address records the plan describes.
- **A table read into is pinned; a table rendered whole draws afresh.** A path into a
  table pins its row for the render and group, as a reference path pins its level,
  and a bare `{/city}` draws each time, as a bare reference does; so a bare table
  beside a path into its family is refused like a bare reference beside a path into
  it. A bare table reference still counts as a read for a `drawGroup`, since the group
  is what draws it apart from the family's pins.
- **Every reference path into one family selects the same rows, per render and
  group.** A selector pins rows for the render, and a read that draws freely before it
  could pin a row the selector contradicts, so accepting both would make the result
  depend on which token rendered first. Requiring one selection per family per group
  is checkable at load with no data lookup beyond the selectors themselves, and the
  error names the rewrite. Two selectors naming one row by key and by name compare
  equal, since the load check resolves them.
- **A parent row with no child row is a load error.** A descendant is drawn inside the
  nearest pinned ancestor, so every ancestor row must lead to a row at every level
  below it, or a render could find nothing to draw. The import script drops or fills
  such rows; the alternative, falling back to a free draw, would break the consistency
  the link exists for without saying so.
- **The choice-of-rows fence guards a data file's root, and requires string fields.**
  A table is a category with a TSV beside its file, so only a root choice has the
  spelling the fence names; a nested choice of same-shaped templates and an inline
  one keep loading. Fields must all be strings because a cell is a string node: a
  choice whose items carry a nested choice is not one table but two linked ones.
- **A table is a record of string columns.** Its columns are the CSV header and the
  `INSERT` column list, fixed by the TSV header, so a table is a record by
  construction; every column is a string until a typed column option earns its place.
- **The key index is built at load, the rest on first draw.** A link is proved
  against the parent's keys and a key's uniqueness is a data mistake, so both are
  load-time; the name index and the per-parent child lists serve only a draw or a
  selection, so they wait for the first one, keeping `New` linear in the bytes read.
- **Two categories may name one TSV.** Each is a view of the file with its own
  format and options, at the cost of holding the rows twice, which is what a
  category over a register with two natural formats asks for; a TSV nothing names
  stays a load error, since that one is a file forgotten rather than shared.
- **The rows of a table are alternatives.** Only one row renders, so a cell in one
  row and a cell in another never meet, and each may select its own row of another
  table; the cells of one row, and whatever they reach, do meet, and so does the
  format beside them. The rule the fence applies is that two reads meet unless
  they sit in two rows of one table, or in a row outside a selected ancestor's. A
  choice's items get no such treatment yet: two items selecting different rows
  are still refused.
- **A table never reaches its own family, by any route.** A `repeat` iteration and a
  `drawGroup` each draw apart on purpose, but a row that lists three localities from
  other regions is the output the family exists to prevent, so the own-family fence
  walks through both rather than stopping where the draw fences do.
- **Tables carrying token cells stay small.** The family fence compares the reads of
  every pair of rows that can render together, so a table whose every row's cell
  selects a row of another table loads in time quadratic in its rows: about a
  second at four thousand rows. No shipped table carries such cells, and a register
  is a column set rather than a set of references, so the fence is left as it is
  until a real data set needs the indexed form.
- **A path is walked once without drawing before it is walked for real.** A path
  that fails below its first level then moves no seeded stream, at the cost of one
  draw-free walk per call, which allocates nothing.
- **A country's postal codes and streets are siblings under its locality.** No open
  source pairs a Swedish street with its postnummer, and pairing the US through its
  ZIPs would shape the two trees differently, so both draw inside the pinned
  locality and an address agrees at that level. A street's own code is the exact
  pairing to add when a source carries it.
- **A locale's `address` reads its country's `geo` tree, so the shipped set loads
  whole.** `data/sv_SE` alone no longer loads: a test loads `data` and prefixes
  the locale, and `--no-shipped-data -d` takes the whole `data` folder or a set of
  one's own.
- **The default embed holds every Swedish postort the import can place and give a
  street-delivery code and a street, and the US places of 25,000 or more.** Sweden
  fits whole in 700 KB; every US place of 10,000 would pass a
  megabyte and fetch 1,200 counties of TIGER files, so the threshold sits where the
  two countries match in size, and `--min-population` and
  `--streets-per-locality` on the import scripts build a fuller set. The two trees
  add about 20 ms to `New`, which loads the shipped set in about 45 ms.
- **A locale's `address` restates its country record's format.** A record cannot
  read another whole and keep its columns, so `sv_SE.address` names the same four
  columns as `geo.SE.address`, each a reference into it, and the format appears
  twice; a column is spelled the same in both, `street-number`, so the two never
  disagree on a name.
- **A postort's kommun comes from its name, its tätort or its codes, never from
  distance.** GeoNames leaves a fifth of Sweden's codes without a kommun and
  carries stale spellings; the nearest code across a border named the wrong kommun
  half the time it was tried, so a postort none of the three rules place is
  dropped, as is one not cased like a place name.
- **A highway designation is not a street, and a US postal code belongs to the place
  holding most of its land inside places.** `I- 55 Bus` and `US Hwy 1` carry the
  most address ranges in many places and would head every address, so the import
  drops names spelled as a route. A ZCTA goes to the place its largest in-place part
  lies in, census-designated places left out since they never ship, and ships only
  when that place does, so a few dozen places whose every code lies mostly in a
  bigger neighbour ship no address; counting the land outside every place too would
  drop a quarter of the places, whose codes straddle unincorporated land, for a
  postal city the USPS mostly names the same way.
- **`--list` stays a plain list of paths.** It is what a script reads, so every line
  has to be a path that `Fake` takes; a marker for the tables a `[selector]` follows,
  or a legend above them, would make the output something to parse before use.
  `--help` names the selector spelling instead, and the Table section teaches it.
- **A layout is always quoted.** A layout may carry the comma that separates
  arguments, `'January 2, 2006'`, and one spelling for every layout beats a rule
  about which ones need the quotes, so the bare spelling is refused naming the
  quoted one. The layout is Go's reference time because the library renders with
  it and a Go caller already knows it; its names are English, and a locale's own
  month and weekday names are data.
- **A title is a table under `sex`.** A prefix drawn apart would put `Mr` on a record
  whose `sex` column says `female`, which is the disagreement the record exists to
  prevent; the tables this set already has are what a title needs, so `en_US.title`
  links to `sex` as `first-name` does. Swedish has no everyday sexed honorific, so
  `sv_SE.title` is a table as well but carries no `parent`. Its weight column is
  `share`, not the `count` a name table carries, because the values are a curated
  proportion rather than bearers anyone counted.
- **A table owns the spelling of a selector on it.** A reference reaches a table by a
  path that carries no selector — `sv_SE.person.first` reads `first-name` through
  `sex` — so the walk that resolved a name cannot say where a reader would type one.
  The table's own location can, which is why it keeps its path, and why an ambiguity
  error names `sv_SE.sex[f].first-name[Kim]` rather than the table's own name.
- **No builtin reads the clock, so a date is bounded by days, never by an age.**
  An `age(min,max)` would make a seeded fixture change with the day it runs on,
  which is what a seed exists to prevent; a birthdate for someone 20 to 60 is
  `date(1966-01-01,2006-12-31,…)`, re-pinned as any fixture is.
- **A name column without a key resolves inside its parent.** A given name both
  sexes carry is a row under each, so `name` cannot be the key; the parent's row
  tells the two apart, `sex[f].first-name[Kim]`, the ambiguity error spells each
  row inside its parent, and a name repeating inside one parent row is refused at
  load, since nothing could then select it.
- **`misc` is what every locale shares.** A category whose facts differ by country
  belongs in that country's locale, read from the register that country's own
  records use; `misc` takes only sources that are international. NHTSA vPIC and
  Mobility Sweden's registrations are national, so they build `en_US.car` and
  `sv_SE.car`; `misc.car` waits for an international source rather than take one
  of theirs.
- **A register's canonical spelling loses to the one its domain writes.** Where a
  source offers several spellings of one fact, the shipped one is what records in
  that domain carry. `misc.timezone` reads `zone.tab` and not the `zone1970.tab`
  that supersedes it, because the latter keeps one zone per set of countries that
  have agreed since 1970: it spells Sweden `Europe/Berlin`, and no Swedish system
  writes that. For the same reason `misc.language` takes the ISO 639-2 register's
  first synonym over CLDR, which says "Chinese, Mandarin" for `zh`.
- **`misc.territory` is the spine, and a `misc` table naming a territory links to
  it.** Where every territory row has a child the column is a `parent`, and the
  import drops the child rows whose territory the set does not ship — 17 of
  `misc.timezone`'s, Antarctica's ten among them. Agreement across a record is
  worth more than the last rows of a table. Where no such link can hold the fact
  stays a column. Layer your own `misc.territory` over the shipped one and you
  must layer `misc.timezone` too, or the link fails at load naming the row.
- **A table whose register publishes no frequency draws evenly.** `misc.httpmethod`,
  `misc.port`, `misc.httpstatus`, `misc.mimetype` and `misc.tld` weigh every row alike,
  so GET is a ninth of the methods drawn. Goal 10 keeps an authored fact out of a
  sourced table, and no register publishes how often a method, a port or a TLD is used,
  so a weight here would be invented. Where one exists it is read, as `misc.timezone`
  reads GeoNames populations and `sv_SE.first-name` SCB bearers.
- **`misc.port` selects by number, and carries no `name`.** 29 of its services sit on
  more than one port, `http-alt` on three, so `misc.port[http-alt]` could name no one
  row. The number is what a port field holds anyway, and `misc.port[443].service`
  reads the other way.
- **`misc.car` is one flat table, not a make linked to its models.** A row is a make
  and a model drawn together, so no render pairs a Volvo with a RAV4. Two linked
  tables would reach the same pairs and add a selector nothing asks for; a make alone
  is `misc.car.make`.
- **`misc.territory` names its sovereign in a column, and there is no `misc.country`
  table.** ISO 3166-1 codes territories, so `territory` is the honest name, and
  `is_independent` in the register gives each one's state. A second table of the 195
  sovereigns would hold a `DK` row beside the territory `DK` row, both carrying
  Denmark's capital, currency, flag and TLD — two owners for one fact, drifting at
  the next import. Splitting the columns to avoid that is worse: put `capital` on
  the territory alone and a country row can no longer name Copenhagen. The column
  cannot be a `parent`, since a table never reaches its own family; a test proves
  every value names a row instead. A territory the register records no state for
  stands alone, which is `EH` alone, and naming one for it would be a claim
  fejkdata has no business making.
- **`misc.tld` is a table of its own, and `misc.territory.tld` stays a column.** A
  `parent` demands a child for every parent row, so linking them would drop every root
  zone row naming no territory, which is most of them, and goal 10 holds a sourced
  table whole. The loader refuses the link outright anyway: `tld` is a column of
  `misc.territory`, and a table may not be named like a column of its ancestor.
- **`misc.territory` carries a currency code, it does not link to `misc.currency`.**
  A `parent` demands a child for every parent row, and ISO 4217 registers codes no
  country's row can name: the funds codes (Mvdol, WIR Euro, US Dollar (Next day)),
  and VED beside VES, both Venezuela's, of which a country row names one. Linking
  would trade the register for the link, and `misc.territory.currency` already pairs a
  country with its currency in one draw.
- **An extension may name two media types.** `.xml`, `.rtf`, `.sub`, `.mpp` and `.ac`
  each name two rows of `misc.mimetype`. Separating them would mean dropping a
  registered type, or naming one by an extension that is not its own — `.mpt` is
  Project's template, not its document. Selecting such a name is an error listing
  both keys; select by type instead.
- **The Swedish ids draw Skatteverket's test series.** A Luhn-valid personnummer
  over a random birth number may be a living person's; 238 and 239 after any date
  are blocked from assignment, so the shipped `personnummer` and
  `samordningsnummer` use those. They sit in a `birth-number` table under `sex`
  rather than as a column of it: the render's shared draw of the family is what
  makes the number and the name agree on sex, and `sex` stays one shape across
  locales instead of collecting every sex-keyed id fact. A samordningsnummer's
  day, the birthday plus 60, is drawn from 61 to 88, valid in every month, rather
  than computed from the date drawn.
- **The US given names come from a mirror of the SSA file.** ssa.gov refuses a
  client outside the US, so `names-us.py` reads a GitHub copy that ends at 2020,
  which a count over the births since 1930 barely feels; `--names` takes the
  official zip. The SSA's placeholder rows are top-1000 entries that name nobody, so
  the import drops them by name rather than by a rank a regeneration would move.
- **`List` advertises direct descents only.** `region.municipality.locality` is
  listed, and `region.locality` resolves too but is not: the set of every descent
  through a chain of five tables is every subsequence of it, and the direct chain is
  the one a reader can predict from the tables' parents.

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

A `--user` command answering `permission denied` on `/cache` has met root-owned
files in the `gocache` volume: Docker creates the volume root-owned, and every
command here that omits `--user` writes into it as root. Hand it back, and again
whenever it recurs:

```sh
docker compose run --rm --user root --entrypoint chown test -R "$(id -u):$(id -g)" /cache
```

Every pull request runs `docker build .` against both the latest and the lowest
supported Go, and must pass before it can be merged — unless it changes none of
the files the build and its tests read, nor the workflow itself, in which case
it's skipped (see [Decisions](#decisions)). That build is the whole gate but the
changelog check, which CI runs against the PR base — vet, complexity, format check
and tests — so run it locally before pushing:

```sh
docker build .                                  # latest
docker build --build-arg GO_VERSION=1.22.12 .   # lowest supported
GO_VERSION=1.22.12 docker compose run --rm test # the same tests, without the image build
```

A change to the shipped data re-pins [`testdata/shipped_shape.txt`](testdata/shipped_shape.txt)
in its own commit:

```sh
REPIN=1 docker compose run --rm --user "$(id -u):$(id -g)" test
```

A shipped table built from a source is rebuilt by its script under
[`data-import/`](data-import), one command per dataset, fetching the source named in
[`DATA-LICENSES.md`](DATA-LICENSES.md). Downloads are cached under
`data-import/cache/`, which is ignored by version control and grows past a
gigabyte, so pass `--cache DIR` to reuse a copy you already have and delete it to
fetch afresh; `geo-us.py` fetches two
TIGER/Line files per county it ships, a few hundred megabytes, `geo-se.py` needs
a Trafikverket API key, free at [data.trafikverket.se](https://data.trafikverket.se/),
in `TRAFIKVERKET_API_KEY` or a `--key-file`, and the Census host behind `geo-us.py`
and `names-us.py` rejects a client for a while after a burst, so `--surnames` takes
a copy of the surname file:

```sh
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/currency.py
docker compose run --rm --user "$(id -u):$(id -g)" -e TRAFIKVERKET_API_KEY data-import data-import/geo-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-us.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/httpmethod.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/httpstatus.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/language.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/mimetype.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-us.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/port.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/protocol.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/territory.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/timezone.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/tld.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/useragent.py
```

To release, head `CHANGELOG.md` with the version's section in place of `Unreleased`
and merge: once `main` passes the gate, CI tags that commit `vX.Y.Z` and publishes
the Gitea release with the section as its body. A top heading of `[Unreleased]`
publishes nothing.

## Layout

```
fejkdata.go     Generator, New, options, the embedded data set, List
node.go         the node model and JSON -> node compilation
table.go        tables: the rows TSV, its options and links, row selection and draws
path.go         the dotted-path walk with its selectors, and proving a path resolves
render.go       Fake and the recursive renderer (choices, format strings, expansions)
record.go       records: Record, the JSON/CSV/SQL serializers, and their entry points
struct.go       structs: FakeStruct, fake tags, and a field's Go type as its column's datatype
inline.go       inline templates: Template, NewTemplate, FakeTemplate, IsTemplate, and their compile and link
template.go     the {token} grammar: scanning, tokens, operands, validation, compiling a format
hold.go         the hold: one draw per expansion for paths and operands, and its fences
draw.go         one reference draw per render and group: draw sets, the group option, and its fence
family.go       a family of linked tables: the rows a render pins, and the fence over paths into one family
reference.go    reference sigils, and binding references across the tree
graph.go        the render graph: edges, cycles, the repeat bound, tree walks
builtins.go     the {name()} function registry and its implementations
layout.go       date and time layouts: the instants one is proved against, and the two samples
checksum.go     the check characters a derivation appends, and the IBAN they sit inside
transform.go    the builtins that rewrite an operand's value, and the ASCII folding
calc.go         the {calc()} arithmetic evaluator: parser, eval, validation
datatype.go     column datatypes: DataType, where datatype and null may sit, a column's datatype
value.go        the value proof: what a typed column or calc operand holds, checked at load
data.go         data loading: fs.FS folders/files -> namespace tree, multi-source merge
cmd/fejkdata/   the fejkdata CLI
data/           shipped data (JSON, and a TSV per table), embedded at build: locale folders, geo, misc
data-import/    the scripts that rebuild each sourced table (see DATA-LICENSES.md)
release-tooling/ the release CI publishes from the changelog heading
testdata/       the pinned shipped shape (see Versioning)
```

## License

MIT — see [LICENSE](LICENSE). Forked from [github.com/Timewave-AB/fakes](https://github.com/Timewave-AB/fakes).
