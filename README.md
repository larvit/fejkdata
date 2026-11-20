# fejkdata

[![Go Reference](https://pkg.go.dev/badge/github.com/larvit/fejkdata.svg)](https://pkg.go.dev/github.com/larvit/fejkdata)
[![Tests](https://github.com/larvit/fejkdata/actions/workflows/test.yml/badge.svg)](https://github.com/larvit/fejkdata/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/larvit/fejkdata)](CHANGELOG.md)

Locale-aware fake data for tests and fixtures, from JSON templates: a Go library and a CLI
with the data built in and no dependencies.

```sh
go install github.com/larvit/fejkdata/cmd/fejkdata@latest
echo '{/sv_SE.person}' | fejkdata   # Sara Eriksson
```

- **Data built in:** the `sv_SE` and `en_US` locales, `da_DK`'s names, persons, addresses, CPR numbers, phone numbers, companies, emails and dates, `de_DE`'s names, persons, addresses, Steuer-IDs, phone numbers, companies, emails and dates, `et_EE`'s names, persons, isikukoodid, phone numbers, companies, emails, dates and addresses, `lt_LT`'s names, persons, asmens kodai, phone numbers, companies, emails, dates and addresses, `lv_LV`'s names, persons, personas kodi, phone numbers, companies, emails, dates and addresses, `nb_NO`'s names, persons,
  fødselsnummer, phone numbers, companies, emails, dates and addresses, `nn_NO`'s persons, fødselsnummer, phone numbers, companies, emails, dates, addresses, months and weekdays, `fi_FI`'s names, persons, henkilötunnus, phone numbers, companies, emails, dates and addresses, and `fr_FR`'s names, persons, NIRs, phone numbers, companies, emails, dates and addresses, with names weighted by how common they are. Every Swedish postal town ships, and so
  does every Danish and Norwegian one holding a street address, and every mainland Finnish one with a street and residents, beside a sample of US places, every German
  place with a postal code, every French commune with people and an address, and every Estonian, Latvian and Lithuanian place with a street. Each address takes its postal code from one town, and its street too, save in [France](docs/locales/geo.md#french-places) and [Germany](docs/locales/geo.md#german-places), whose streets are invented.
- **Valid ids:** personnummer and samordningsnummer from Skatteverket's test series, and
  fødselsnummer from Skatteetaten's synthetic series, and henkilötunnus with the individual numbers
  900 to 999, which DVV issues to no one, none of which a real person holds, though a checker that
  requires a month of 1 to 12 refuses the fødselsnummer, and one that applies the extra checks of the
  synthetic henkilötunnus refuses those; Steuer-IDs from ELSTER's test series, which a checker that refuses a leading zero refuses; Danish CPR numbers that pass the
  office's modulus 11 check, and French NIRs that pass INSEE's key check, and Estonian isikukoodid, Latvian personas kodi and Lithuanian asmens kodai that pass the check digit, so a drawn one may be a living person's; Swedish
  organisationsnummer and VAT numbers that pass the format and the checksum, not a VIES lookup, so a drawn one may be a real organisation's; SSNs
  and ITINs in the ranges the SSA and the IRS assign, so a drawn one may be a real
  person's; Luhn-valid card numbers; and IBANs for seven European countries.
- **Registers:** ISO 3166-1 territories, ISO 4217 currencies, ISO 639 languages, IANA
  time zones, TLDs, ports, HTTP status codes, media types and more, as tables you
  select a row from.
- **Bulk output:** records as JSON, NDJSON, CSV, PostgreSQL `COPY` text or multi-row SQL `INSERT`s, streamed.
- **Go structs:** `FakeStruct` fills a struct from its `fake:"…"` tags.
- **Your own data:** each JSON file in a folder you pass adds a path, such as
  `{/acme.customer}`, and a mistake in it fails at load, never mid-render.
- **Reproducible:** the same seed, date, data and versions give the same output, when one goroutine
  draws it.

[Install](#install) · [CLI](#cli) · [Library](#library) · [Data](#data) · [Data format (templates)](#data-format) ·
[Functions](#functions) · [Versioning](#versioning) · [Audience](#audience) ·
[Goals](#goals) · [Development](#development) · [License](#license)

## Install

```sh
brew install larvit/tap/fejkdata                                  # macOS, Linux
scoop bucket add larvit https://github.com/larvit/scoop-bucket && scoop install fejkdata  # Windows
docker run --rm -i larvit/fejkdata < template.json                # any OS with Docker
go install github.com/larvit/fejkdata/cmd/fejkdata@latest         # any OS with Go
```

On shared CI runners, pull `ghcr.io/larvit/fejkdata` instead: Docker Hub limits anonymous
pulls per IP address.

Debian, Ubuntu, Fedora, RHEL and Alpine install from fejkdata's signed repositories, which
hold the latest release:

```sh
# Debian, Ubuntu
sudo curl -fsSLo /etc/apt/keyrings/fejkdata.asc https://larvit.github.io/fejkdata/fejkdata.asc
echo 'deb [signed-by=/etc/apt/keyrings/fejkdata.asc] https://larvit.github.io/fejkdata/deb stable main' | sudo tee /etc/apt/sources.list.d/fejkdata.list
sudo apt update && sudo apt install fejkdata

# Fedora, RHEL
sudo curl -fsSLo /etc/yum.repos.d/fejkdata.repo https://larvit.github.io/fejkdata/rpm/fejkdata.repo
sudo dnf install fejkdata

# Alpine, as root
wget -qO /etc/apk/keys/fejkdata.rsa.pub https://larvit.github.io/fejkdata/apk/fejkdata.rsa.pub
echo https://larvit.github.io/fejkdata/apk >> /etc/apk/repositories
apk add fejkdata
```

Each [release](https://github.com/larvit/fejkdata/releases) also carries the binaries for
Linux, macOS and Windows on amd64 and arm64, and the `.deb`, `.rpm` and `.apk` packages.
The binaries are unsigned. macOS refuses to run one a browser downloaded until you run
`xattr -d com.apple.quarantine fejkdata`.

No npm package ships yet. From JavaScript or TypeScript, run the CLI as a child process,
with `--seed` and `--today` to repeat a run.

## Quick start

### Bulk fixtures

Save this record as `person.json`. Its `format` draws one sex, one last name and one address, and each column reads those same draws, so a row's name, sex, personnummer, email and address agree. [Records](#records) explains the record.

```json
{
  "format": "{/sv_SE.sex as s}{/sv_SE.last-name as l}{/sv_SE.address as a}",
  "sex": "{s.name}",
  "first": "{s.first-name}",
  "last": "{l}",
  "personnummer": "{s.personnummer}",
  "email": "{slugify(s.first-name)}.{slugify(l)}@example.com",
  "street": "{a.street} {a.street-number}",
  "postal-code": "{a.postal-code}",
  "locality": "{a.locality}"
}
```

```sh
fejkdata --seed 1 --today 2026-01-15 --format csv < person.json    # a header, then one person, the same every day
fejkdata --format csv --repeat 1000000 < person.json > people.csv  # a million rows
fejkdata --format sql --table people --repeat 1000000 < person.json > people.sql
```

The first command prints:

```text
email,first,last,locality,personnummer,postal-code,sex,street
viktoria.larsson@example.com,Viktoria,Larsson,Östersund,501120-2388,831 73,kvinna,Kärrvägen 1
```

A fixture this size repeats personnummer and email values. [Personnummer](docs/locales/sv_SE.md#personnummer-and-samordningsnummer) explains why a personnummer repeats. On MySQL or MariaDB, [Records](#records) says what to set before the SQL loads.

### Go tests

Declare the struct, and `FakeStruct` fills it from its tags. A tag binds a draw with `as`, such as `{/sv_SE.sex as s}`, and the other tags read that draw in any field order, so the fields agree. [Names](#names) explains `as`; [Library](#library) covers the rest of the API.

```go
package signup

import (
	"testing"
	"time"

	"github.com/larvit/fejkdata"
	svse "github.com/larvit/fejkdata/data/sv_SE"
)

type Person struct {
	Sex          string `fake:"{/sv_SE.sex as s}{s.name}"`
	First        string `fake:"{s.first-name}"`
	Last         string `fake:"{/sv_SE.last-name as l}{l}"`
	Personnummer string `fake:"{s.personnummer}"`
	Email        string `fake:"{slugify(s.first-name)}.{slugify(l)}@example.com"`
	Street       string `fake:"{/sv_SE.address as a}{a.street} {a.street-number}"`
	PostalCode   string `fake:"{a.postal-code}"`
	Locality     string `fake:"{a.locality}"`
}

func TestSignup(t *testing.T) {
	day := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	f, err := fejkdata.New(fejkdata.WithDataFS(svse.Modules()...), fejkdata.WithSeed(1), fejkdata.WithToday(day))
	if err != nil {
		t.Fatal(err)
	}
	var p Person
	if err := f.FakeStruct(&p); err != nil {
		t.Fatal(err)
	}
	// Sex is kvinna or man.
	// With seed 1, p is {Sex:kvinna First:Viktoria Last:Larsson Personnummer:501120-2388 Email:viktoria.larsson@example.com Street:Kärrvägen 1 PostalCode:831 73 Locality:Östersund}
}
```

`WithSeed(1)` and `WithToday(day)` give the same person on every run of one version; another seed gives another.

### Replacing real data with fakes

To replace the real values in existing text, such as a document holding real names, ids and URLs,
your program finds each real value and puts a fake of the same shape in its place.
fejkdata renders the fakes. Each part of the job:

- The same fake wherever one original recurs in a render: bind a pick to a name,
  `{/en_US.person as p}`, and read `{p}` each time ([Names](#names)).
- Free text, word by word: `{reword(text, /en_US.word)}` keeps its punctuation, digits,
  case and word lengths, where the text writes each `{` and `}` as `{{` and `}}`
  ([Transforms](#transforms)).
- A value of a given length: `{lower(n)}`, `{upper(n)}`, `{digits(n)}` or `{hex(n)}`
  ([Functions](#functions)).
- A v4 UUID: `{/misc.uuid}`, since `{uuid()}` draws a v7.
- An id of a fixed shape: literal text around a draw, such as `557058:{/misc.uuid}`
  ([Format string](#format-string)).
- A URL of a given form: `{/en_US.url}` draws a host and maybe a path, and the template
  adds the rest, such as `{/en_US.url}?{lower(4)}={digits(3)}`.
- A fake that doesn't hold its original: a `where` clause on its draw, such as
  `{hex(4) where not containing '7958'}`, which may read the original from a name. A clause
  filters the alphabet calls, such as `{hex(n)}` and `{uuid()}`, and a path whose values are fixed
  text ([Filters](#filters)).
- A longer word list than a locale's `word`: a category of your own ([Your own data](#your-own-data)).

```json
"{/en_US.person as p}{p} <{lower(6)}@example.com> signed as {p}, id 557058:{/misc.uuid}, {/en_US.url}?{lower(4)}={digits(3)}"
```

Renders e.g. `David Greer <tdsagk@example.com> signed as David Greer, id
557058:3bcf0fc3-f7b0-4d58-bf7f-88c29c39f316, https://blog.example.com?ojat=642`.

Your program keeps the rest of the job: giving one original the same fake across renders, and
keeping a fake from equalling or holding its original where no clause can filter it, as for
`{/en_US.person}`, `{reword()}` and `{/misc.uuid}`.

## CLI

```sh
echo '{/sv_SE.person}' | fejkdata                        # Sara Eriksson
echo '{/sv_SE.person.last}' | fejkdata                   # Eriksson
echo '{/sv_SE.person} hihi' | fejkdata                   # Sara Eriksson hihi
echo '{/sv_SE.address}' | fejkdata --seed 42             # the same address every run
echo '{/sv_SE.personnummer}' | fejkdata --seed 42 --today 2026-11-04   # the same number on any day
echo '{/sv_SE.word}' | fejkdata -n 3                     # nät, barn and sol, a line each
printf '%s' '{/sv_SE.word}' | fejkdata -n 3 --separator ', '   # nät, barn, sol
fejkdata --list                                          # every path the data offers
echo '{/misc.territory[SE].capital}' | fejkdata          # Stockholm — a table's row, selected by key or name
echo '{/geo.SE.locality[Lund].street}' | fejkdata        # Fjelievägen — a linked table, drawn inside the row
echo '{/sv_SE.word}' | fejkdata --data-path ./mydata     # layer a directory over the shipped data
fejkdata --no-shipped-data -d ./mydata --list            # only your data
echo '{"format":"name: {x}","x":["bosse","lina"]}' | fejkdata   # name: bosse or name: lina, with no trailing newline; see What stdin renders
fejkdata < template.txt                                  # a template kept in a file
fejkdata <<'EOF'                                         # born 2003-11-27 — a quoted heredoc passes ', $ and \ as written
born {date(1990-01-01,2010-12-31,'2006-01-02')}
EOF
```

fejkdata renders the template on its stdin: a format string, or a JSON object, array, string or `null`. Its `{…}` tokens reach the data by reference from the root,
`{/sv_SE.person.last}`, so shipped and `--data-path` categories are alike available,
and everything outside them prints as written. A reference names a category, or a field
inside one: each dot segment descends one level — folders, then the category (a JSON
file), then fields — and `[SE]` after a [table](#table) selects its row. `--list` prints
these paths; a template reads one as `{/path}`.

| Flag | |
|------|--|
| `-d`, `--data-path D` | a directory to layer over the shipped data; repeatable, the last wins a name clash |
| `--no-shipped-data` | load only the `--data-path` directories |
| `-s`, `--seed N` | same seed, date, data and versions: identical output (default: a random seed) |
| `--today DAY` | the date to treat as today, as `YYYY-MM-DD` from year 1000 on; ages in [`{born(…)}`](#functions), and so in `sv_SE.personnummer` and `sv_SE.person`, count from it, as does `today` in a [`{date(…)}`](#dates-around-today) bound (default: the system's date, so those draws change as days pass); give it beside `--seed` to repeat a run's output on a later day |
| `-n`, `--repeat N` | render the template N times (0 to 1048576), each an independent draw, streamed; `0` prints nothing, `[]` under `--format json`, or the header under `--format csv` or `tsv` |
| `--separator S` | between repeated renders (default empty); refused with a record `--format`, which has its own separators |
| `--format F` | `text` (default), `json`, `ndjson`, `csv`, `tsv`, `pgcopy` or `sql` — a record's columns, one record per row (json frames them as an array) |
| `--table T` | the INSERT or COPY target for `--format sql` or `pgcopy`, ignored under any other (default: a lone reference's final name, selectors left out, such as `users` for `{/users}`, else `records`) |
| `--list` | print every path, then exit; `--repeat`, `--separator`, `--format` and `--table` do nothing here |
| `--version`, `-h`, `--help` | print, then exit |

`--name value` and `--name=value` both work, a short flag's value attaches or
follows (`-n3`, `-n 3`) and short flags bundle (`-hn 3`) — see
[Decisions](docs/decisions.md#flags-follow-getopt_long); flags go anywhere, `--` ends them. Exit codes: `0` success, `1` runtime error (data
that fails to load, a missing `-d` folder included, a lone reference to nothing under `--format`), `2` misuse — a bad flag, an argument
other than a flag, an unreadable stdin, nothing piped in, or a
template that does not compile. From a checkout:
`go run ./cmd/fejkdata …`.

### What stdin renders

Under `--format text`, the default, what you send is what renders, and fejkdata adds nothing. A format string keeps every
byte, so `echo`, a file or a heredoc ends each render with its newline, and `printf '%s'`
sends none; `-n` joins renders with `--separator`, empty by default. A JSON object, array
or string drops the whitespace around it, and `null` alone renders nothing; `42` or `true`
alone is a format string, so it prints as written, newline included. An empty or blank stdin is a format string like
any other, and renders as sent. A JSON template that should end
each render in a newline writes `\n` in its format, from a quoted heredoc or a file. A
lone reference, one `{/…}` and nothing else, is the record it names under `--format`, one
newline after it allowed. Some shells' `echo` reads a backslash as an escape, so pipe a
template holding one from a quoted heredoc or a file. The rest of stdin's rules are under
[Decisions](docs/decisions.md#the-cli-renders-exactly-the-template-on-its-stdin-and-adds-nothing).

### PowerShell and cmd.exe

In cmd.exe, `echo` sends the space before `|` and any single quotes, so write neither: around `{/sv_SE.person}` they are text, and `--format` stops naming the record.

```bat
echo {/sv_SE.person}| fejkdata
```

In PowerShell 7, single quotes keep the braces:

```powershell
'{/sv_SE.person}' | fejkdata
```

Windows PowerShell 5.1 pipes a byte-order mark ahead of the template. That makes a lone reference under `--format` text and breaks a JSON template. Its `>` also writes UTF-16, so a redirected file is not UTF-8. Read the template from a file and redirect in cmd.exe instead:

```powershell
cmd /c "fejkdata --format csv --repeat 1000000 < person.json > people.csv"
```

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
echo '{/sql}' | fejkdata --seed 1 --data-path ./mydata
# INSERT INTO users VALUES('zoom'),('wahoo'),('blip');
echo '{/sql}' | fejkdata --repeat 100 --data-path ./mydata > seed.sql
```

That `format` string is the free-form spelling — you hand-write the whole row.
For structured output a record writes the row for you.

### Records

A record is a template seen as columns: its fields are the columns, its `format`
the whole. `--format json|ndjson|csv|tsv|pgcopy|sql` writes the records; the library's
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
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format json               # [{"first":"Bo","last":"Lovelace"}]
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format ndjson             # {"first":"Bo","last":"Lovelace"}
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format csv                # first,last  →  Bo,Lovelace
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format sql                # INSERT INTO "users" ("first", "last") VALUES ('Bo', 'Lovelace');
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format sql --table people # INSERT into another table
```

`--repeat` streams that many records — `json` frames them as one array document,
`ndjson` writes one object per line; `csv` and `tsv` a header, then a row per record; `pgcopy` a
`COPY … FROM STDIN;` line, a tab-separated row per record and a closing `\.`; `sql` one multi-row
INSERT per line, each of up to 1000 rows. Only a category-level template is a record, or a one-item choice holding one; a
field, a choice of more items or a folder errors, and so does a `repeat` on
the template itself, which composes the format into one string rather than
projecting columns — ask for more records with `--repeat`. A `repeat` on a column
is fine.

Columns are written in name order, whatever order the fields appear in. A category
whose fields are the parts of one value — `misc.uuid` — projects those parts rather than
the value, so for one column holding what `Fake` renders, write
`{"format":"","uuid":"{/misc.uuid}"}` as a fieldless category asks for. A category whose
every field is in [`hidden`](#hidden), such as `sv_SE.price`, has no column and fails with `ErrNoColumns`,
which names the same one-column record to write instead.

A field only other fields read goes under [`hidden`](#hidden) and writes no column.

A column carrying a newline keeps it inside the quoted CSV field or the SQL string
literal, so a row can span physical lines: read the stream with a CSV or SQL
parser rather than splitting it on newlines.

A statement fails whole: one row a constraint refuses loses its whole statement, up to 1000 rows.
A shipped category may repeat a value. `{seq()}` counts 1, 2, 3 through a run. A column that ends
in a `.`, the count and fixed text, as the email below ends in `.{seq(email)}@example.com`, never
repeats, whatever is drawn before the `.`, so a `UNIQUE` column loads. A named counter, `seq(email)`,
counts its column from 1. Save `ids.json`:

```json
{
  "format": "",
  "callsign": "T-{seq(callsign)}",
  "email": "{/sv_SE.username}.{seq(email)}@example.com"
}
```

`fejkdata --seed 1 --format csv --repeat 3 < ids.json` prints:

```text
callsign,email
T-1,fizz14.1@example.com
T-2,modig.drake2024.2@example.com
T-3,majahed7.3@example.com
```

The SQL is ANSI — identifiers in double quotes, a literal quote doubled (`''`),
backslashes passed through — so a hyphenated field like `postal-code` stays a
valid identifier. PostgreSQL and SQLite take it as written; MySQL and MariaDB need
`ANSI_QUOTES` and `NO_BACKSLASH_ESCAPES` set first, or they read `"users"` as a
string and a backslash as an escape.

`--format tsv` writes records as the TSV file of a [table](#table), so a later run can draw from them.
A `series` record is run once, and its output kept as `mydata/series.tsv`:

```json
{ "format": "", "id": "{seq()}", "customer": ["Ada", "Bo", "Cy"] }
```

```sh
echo '{/series}' | fejkdata --seed 1 -d ./gen --format tsv -n 3 > mydata/series.tsv
```

```tsv
customer	id
Bo	1
Ada	2
Cy	3
```

Beside it, `mydata/series.json` makes the file a table. A second run then draws orders, each
naming a series that exists and the customer on that row:

```json
{ "format": "{id}", "rows": "series.tsv" }
```

```sh
echo '{"format":"","series_id":"{/series as s}{s.id}","customer":"{s.customer}"}' |
  fejkdata --seed 1 -d ./mydata --format csv -n 4
```

```text
customer,series_id
Bo,1
Cy,3
Ada,2
Ada,2
```

Each order draws a row afresh, so several orders may name one series, and the table's
[`weight` option](#table) skews that. To load the first run's records into a database, run it
again with the same seed and `--format csv` or `sql`, as the file is written for a table:
a brace is doubled, `{{x}}` for `{x}`, and a tab or line break in a value is a space. A null
column and an empty string are both an empty cell. A record of one column whose value is
empty writes a blank line, which a table refuses, so give such a record a second column.

A record written only to emit columns still needs a `format` — the grammar's one
required key — so `"format": ""` carries the fields with an inert format: `Fake` renders
it as nothing.

A record is one render, so its columns read one pick of each [name](#names) its category
binds: `"code": "{/currency as c}{c.code}"` beside `"symbol": "{c.symbol}"` is one
currency, while `{/currency.code}` and `{/currency.symbol}` in two columns may name two. A
column the format binds to a name, as `{place as p}` binds `place`, holds that same pick, so it
agrees with the columns reading `p`; a column bound only in part, `{place.town as t}`, draws
afresh, so bind it whole and read `{p.town}`.

A reference into a category's own fields, `{/users.first}` inside `users`, is a draw of its
own, apart from the fields beside it; to share one draw, read the sibling as a field, or bind
it to a name and read the name
([Decisions](docs/decisions.md#a-category-may-reference-itself-and-a-records-fences-run-at-load)).

### Load into a database

fejkdata writes to its standard output and nothing else, so a pipe carries a run into a
database or an HTTP endpoint ([Decisions](docs/decisions.md#no-database-or-http-sink-is-built-in)). Save
`mydata/users.json`:

```json
{ "format": "", "id": { "format": "{seq()}", "datatype": "integer" }, "name": ["Ada", "Bo", "Cy"] }
```

Each line sends a thousand records to `users`: a table for `psql`, `sqlite3` and `mysql`,
which must already have the columns `id` and `name`, a collection for `mongoimport` and an
endpoint for `curl`. `csv` writes the columns in name order, so the `\copy` lists them:

```sh
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format sql --repeat 1000 | psql mydb
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format pgcopy --repeat 1000 | psql -v ON_ERROR_STOP=1 mydb
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format sql --repeat 1000 | sqlite3 app.db
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format sql --repeat 1000 | mysql --init-command="SET SESSION sql_mode = CONCAT(@@sql_mode, ',ANSI_QUOTES,NO_BACKSLASH_ESCAPES')" mydb
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format csv --repeat 1000 | psql -c '\copy users (id, name) FROM pstdin WITH (FORMAT csv, HEADER)' mydb
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format ndjson --repeat 1000 | mongoimport --db mydb --collection users
echo '{/users}' | fejkdata --seed 1 -d ./mydata --format ndjson --repeat 1000 | curl --data-binary @- -H 'Content-Type: application/x-ndjson' https://example.test/users
```

The `mysql` line sets the two modes the [SQL note](#records) names.

Loading a million rows from a file into a table of `id bigint primary key, name text` on PostgreSQL
17 through `psql` took 1.4 s from `pgcopy` and 6.9 s from `sql`. Generating either file takes about 1.3 s.
`-v ON_ERROR_STOP=1` makes `psql` stop at a missing table or column, rather than read every row as SQL.

## Library

```sh
go get github.com/larvit/fejkdata   # Go 1.22+, 64-bit targets
```

```go
import (
	"time"

	"github.com/larvit/fejkdata"
	svse "github.com/larvit/fejkdata/data/sv_SE"
)

f, err := fejkdata.New(fejkdata.WithDataFS(svse.Modules()...), fejkdata.WithSeed(42), fejkdata.WithToday(time.Now()))
if err != nil {
	log.Fatal(err)
}
v, err := f.Fake("sv_SE.address") // "Triogatan 43\n442 42 Kungälv"
paths := f.List()                  // the paths Fake renders, sorted; direct descents only
v, err = f.FakeTemplate("name: {/sv_SE.person.last}")      // compile + render in one call
t, err := f.NewTemplate(`{"format":"name: {x}","x":["bosse","lina"]}`) // compile once
v = t.Fake()                                              // render many times, no re-parse
r, err := f.FakeRecord("users")           // one record: each field a column
s := r.JSON()                             // {"first":"Ada","last":"Lovelace"}
r, err = f.FakeRecordTemplate(`{"format":"{x}","x":["a","b"]}`) // compile + render inline
rt, err := f.NewRecordTemplate(`{"format":"","id":{"format":"{seq(users)}","datatype":"integer"}}`) // compile once
r = rt.Fake()                             // render a row; loop this for bulk output
err = f.FakeStruct(&user)                 // fill a struct's fake:"…" tagged fields
ok, err := fejkdata.IsTemplate(arg)       // an inline template by its shape, else a path
ok = fejkdata.IsLonePath("users")         // true: {/users} is one reference and nothing else
name := fejkdata.PathName("misc.currency[US Dollar]") // "currency": the final name, selectors left out
calls := fejkdata.Builtins()              // each way to call a function, such as hex(n), sorted
```

| Option | |
|--------|--|
| `WithSeed(n)` | required; the same seed, date, data and versions give the same sequence. For a new seed each run, set `n := rand.Uint64()` (`math/rand/v2`), and log `n` to replay the run |
| `WithToday(day)` | required; the date to treat as today. Ages in [`{born(…)}`](#functions), and so in `sv_SE.personnummer` and `sv_SE.person`, count from it, as does `today` in a [`{date(…)}`](#dates-around-today) bound. Give `time.Now()` for the current date, or a fixed day to repeat a run's output on a later day. It uses the calendar day that `day` shows in its own location and ignores the time of day. `New` refuses a year outside 1000 to 9999 |
| `WithDataPath(dir)` | layer a directory; repeat to layer several, the last wins a clash |
| `WithDataFS(modules...)` | layer modules in order, each an `fs.FS`: a shipped package's `Modules()...`, or your own `embed.FS`; `WithDataFS()` loads none |

A shipped package's `Modules()` returns its own `FS` and every module it reads, such as a
locale's country's places ([`geo`](docs/locales/geo.md)); `misc.*` comes from `misc.Modules()`, in
`github.com/larvit/fejkdata/data/misc`. Your program links only the packages it imports.
`data.Modules()`, from `github.com/larvit/fejkdata/data`, is every shipped module, and
links all of them. Two locales, with `enus` from `github.com/larvit/fejkdata/data/en_US`,
are `WithDataFS(append(svse.Modules(), enus.Modules()...)...)`. To stand in your
own places, pass the locale's `FS` alone beside a tree providing `geo.SE.address`, a record
with a `format` and the columns `sv_SE.address` reads, `street`, `street-number`,
`postal-code`, `locality`, `municipality` and `region`: `WithDataFS(svse.FS, myPlaces)`. `New` fails where no
`WithDataFS` or `WithDataPath` names the data to load, and where `WithSeed` or `WithToday` is missing.

`New` refuses data that breaks the grammar, could mean two things or cannot render a valid
value, and `NewTemplate` refuses such a template. A category of an [indexed](#indexed-sources) source is
checked on the first call reaching it instead. Data that fails to load, in `New` or on
that first call, gives an error matching `ErrLoad`. On a loaded generator:

- `Fake` fails only for a path that names nothing, could name two things or reads one
  draw of a level carrying a `repeat`, or for data that fails to load (`ErrLoad`), with
  the same error every call.
- `FakeStruct` fails only for a non-struct argument, a type its tags do not describe, or
  data that fails to load (`ErrLoad`), with the same error every call.
- `Template.Fake` cannot fail.

A path, wherever a call takes one, may start with one `/`, as a reference does.

A `*Record` carries its columns via `Columns()` — each a `Column` of `Name`,
`DataType`, rendered `Value` and `Null` — and serializes them with `JSON()`
(one object), `CSVHeader()`/`CSVLine()`, `TSVHeader()`/`TSVLine()`, `SQLInsert(table)` (one statement),
`SQLInsertHeader(table)`/`SQLValues()` or `PGCopyHeader(table)`/`PGCopyLine()` — the shapes the CLI's `--format` writes, each without a
trailing newline. To write a multi-row INSERT, write the header, then each record's `SQLValues()` with a comma
between them, then `;`. To write a COPY, write the header, each record's `PGCopyLine()` on a line of its own, then a line `\.`. `FakeRecord` and
`FakeRecordTemplate` take a record; a path or template that is not one — a bare string, a
choice of more than one item, or a folder — errors.

```go
type User struct {
	ID   int64   `fake:"{seq()}"`
	Last string  `fake:"sv_SE.person.last"`
	Age  uint8   `fake:"{int(18,99)}"`
	Nick *string `fake:"[null, \"{/sv_SE.username}\"]"`
	Home Address // filled from Address's own tags
}

type Address struct {
	Street     string `fake:"{/sv_SE.address as a}{a.street} {a.street-number}"`
	PostalCode string `fake:"{a.postal-code}"` // the same address's code
	Locality   string `fake:"{a.locality}"`
}
```

`FakeStruct` fills a struct through a pointer: each exported field tagged `fake:"…"` is
a column of one record, its tag a path or an inline template — told apart by
`IsTemplate` — and its Go type the column's
[datatype](#datatype): a string, bool, integer or float kind, or a pointer to one. A
column that can be null needs a pointer, which a [`null`](#null) item leaves nil. An
integer stays within int64 whatever its kind, and a value the kind cannot hold, such as
`{int(0,300)}` in a `uint8`, is refused naming a kind that holds it. The tags are a record's columns, so they share its
[names](#names): `First` tagged `{/sv_SE.person as p}{p.first}` and `Sex` tagged `{p.sex}`
describe one person, while two path tags into `sv_SE.person` are two draws. The fields an
embedded struct promotes are columns of the same record; a named struct field, or a
pointer to one, fills from its own tags as a record of its own, so its names and picks are
its own. `fake:"-"` leaves a field unfilled.
Untagged fields keep their values, and so does a pointer back to a struct already being filled; a type
whose fields reach more than 1024 structs is refused, naming `fake:"-"` to cut it. The
first call for a type compiles its tags and reports what they get wrong, with the same
error on every later call; a `datatype` in a tag other than the one its Go type sets is
refused, naming that type.

A `*Generator` is safe for concurrent use; a sequence is reproducible only
when drawn from one goroutine. Editing a template shifts what it draws, and
everything drawn after it.

### Your own CLI

Package `cli` is the `fejkdata` command, and its `Main` takes the modules the command
loads, so a CLI carrying your own data is a `main` package of your own:

```go
package main

import (
	"embed"
	"os"

	"github.com/larvit/fejkdata/cli"
	"github.com/larvit/fejkdata/data"
)

//go:embed *.json
var mydata embed.FS

func main() { os.Exit(cli.Main(append(data.Modules(), mydata)...)) }
```

The embed's folders are the paths: `//go:embed acme` serves `acme/customer.json` as
`{/acme.customer}`. Add `*.tsv` to the embed where your data holds a table. The CLI loads
its data only when it runs. To catch broken data at `go test`, call
`fejkdata.New(fejkdata.WithDataFS(...), fejkdata.WithSeed(1), fejkdata.WithToday(time.Now()))` over the same modules, and
render each path once: a module with an [index](#indexed-sources) checks a category only when a call first
reaches it. The CLI takes `fejkdata`'s flags and exit codes, and names itself by the file
it runs from in its help and errors. Its `--help` lists the [functions](#functions) a
template calls. `--no-shipped-data` leaves out the modules passed to
`Main`, and `--version` prints your module's version beside the fejkdata version it is built with.

## Data
### Shipped modules

The shipped data comes in modules, each a Go package in the folder under [`data/`](data)
its namespace spells, a dot a slash: one per locale (`data/da_DK`, `data/de_DE`, `data/en_US`, `data/et_EE`, `data/fi_FI`, `data/fr_FR`, `data/lt_LT`, `data/lv_LV`, `data/nb_NO`,
`data/nn_NO`, `data/sv_SE`), one per country's tree of places (`data/geo/DE`, `data/geo/DK`, `data/geo/EE`, `data/geo/FI`, `data/geo/FR`, `data/geo/LT`, `data/geo/LV`, `data/geo/NO`, `data/geo/SE`, `data/geo/US`), and the locale-neutral
`data/misc`. A package's Go name is its namespace in lower case without separators:
`dadk`, `dede`, `enus`, `etee`, `fifi`, `frfr`, `ltlt`, `lvlv`, `nbno`, `nnno`, `svse`, `geode`, `geodk`, `geoee`, `geofi`, `geofr`, `geolt`, `geolv`, `geono`, `geose`, `geous`, `misc`. Each embeds its files, so nothing is
read from disk. The CLI carries every module, and a library loads those it passes to `WithDataFS`
([Library](#library)). A directory is a namespace: each JSON file is a category named
after the file, each subdirectory a dot-path segment, so `mydata/sv_SE/person.json` is
`sv_SE.person` and replaces the shipped one.
Sources merge in order; matching folders combine, any other clash is won by the
last loaded. Names may not use `.`, `|`, `(`, `{`, `}`, `[`, `]`, `"` or `/`, nor be `-`, which a
struct tag reserves. Nor may a name contain ` as `, which binds a [name](#names), ` where `, which
opens a [filter](#filters), or ` distinct `, which opens a [distinct draw](#distinct-draws), or end in
` where` or ` distinct`. A dot-prefixed
entry is never a category or a folder, so a data directory can also be a checkout, though
a category may name a hidden rows file beside it.

#### Indexed sources

A source whose root holds `.fejkdata.json` with an `index` loads each category on the
first call reaching it, so `New` does not check them: load each category once in a test,
as CI does for the shipped set. Any other source loads in `New`, with the indexed
categories it reads. An indexed category reading one that another source replaced loads
on its own first reach, and fails there if the replacement breaks it. The index names
every category of its source, so a category file it leaves out never loads.

The index maps each category's dot path to an entry: `parent`, the table it links to,
which needs its own entry in the same folder; and `paths`, every path `List` advertises
below the category, `""` for the category itself. Each shipped module's manifest, such as
[`data/sv_SE/.fejkdata.json`](data/sv_SE/.fejkdata.json), carries one, so in a copy you
edit, delete its `index` and keep its `namespace`, or keep the index current.
A table whose parent column differs from its entry's `parent` fails at first reach, and a
table linking under another source's table needs a source without an index. `//go:embed`
of a directory leaves the manifest out, and the source then loads whole in `New` and at
the root, with no `namespace`; name the manifest in the pattern too.

#### Manifest namespace

A manifest's `namespace` is the dot path its source's root sits under: a folder holding
`person.json` with `"namespace": "sv_SE"` provides `sv_SE.person`, its subfolders adding
segments below that. It needs no index. Each shipped module carries one, so
`data/sv_SE/person.json` is `sv_SE.person`.

#### Data sources

Where a script under [`data-import/`](data-import) reads a source, the facts come from
that source, and the script drops a row only by a rule it states; elsewhere a small hand-written
set ships, and [`todo.md`](todo.md) carries the step that replaces it (goal 3.2).
The `sentence` and `paragraph` patterns state no fact and stay hand-written. Where a category
has no path that draws only values reaching nothing real (goal 3.3), write a fixed value such
as `example.com` into the template; `todo.md` plans the paths. [`url` and `email`](#locale-categories) have such a path, `reserved`.
[`DATA-LICENSES.md`](DATA-LICENSES.md) names each table's source and licence.

### Locale categories

`sv_SE` and `en_US` each carry `address`, `color`, `company`, `date`, `email`, `first-name`,
`last-name`, `month`, `paragraph`, `person`, `phone`, `price`, `sentence`, `sex`, `time`, `title`, `url`, `username`,
`weekday` and `word`, formatted per locale; `sv_SE` adds `birth-number`, `personnummer`,
`samordningsnummer`, `personnummer-12`, `personnummer-date`, `samordningsnummer-12`, `organisationsnummer`, `vat`,
`licenseplate`, `licenseplate-letters`, `bankgiro`, `plusgiro`, `bankaccount`, `bloodtype`, `car`, `company-coined`,
`company-word`, `holiday`, `phone-area`, `phone-subscriber`, `sni`, `ssyk`, `territory`, `word-adj`, `word-adv`, `word-noun` and `word-verb`;
`en_US` adds `ssn`, `itin`,
`phone-exchange` (the three digits after the area code), `word-adj`, `word-adv`, `word-noun` and `word-verb`.

What `sv_SE` and `en_US` do not share: `sv_SE.title` has no `parent`, so it is selected as
`sv_SE.title[dr]` rather than inside a sex.

`da_DK`, `de_DE`, `et_EE`, `fi_FI`, `fr_FR`, `lt_LT`, `lv_LV`, `nb_NO` and `nn_NO` carry `color`, `month`, `paragraph`, `price`, `sentence`,
`time`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb` too. `price` and `time` are written as the
country writes them:

| Locale | `price` | `time` |
|---|---|---|
| `da_DK` | `1.234,50 kr.` | `15.04` |
| `de_DE` | `1.234,50 €` | `15:04` |
| `et_EE`, `fr_FR`, `lt_LT`, `lv_LV` | `1 234,50 €` | `15:04` |
| `fi_FI` | `1 234,50 €` | `15.04` |
| `nb_NO`, `nn_NO` | `1 234,50 kr` | `15:04` |

A `url` and a `username` are made of the locale's own words, in ASCII, and a `color` is a colour name in the locale's
language or, in two draws of three, a hex code such as `#1a2b3c`. `month` and `weekday` are Unicode CLDR's stand-alone
names, so `{/fi_FI.month[1]}` is `tammikuu`, not the declined `tammikuuta`. As for `sv_SE`, `month` is selected by 1 to 12
and `weekday` by 1 (Monday) to 7 (Sunday), as ISO 8601 numbers them: `en_US.weekday[1]` is `Monday`, not `Sunday`. `lv_LV.weekday` is capitalised, `Pirmdiena`, as CLDR gives it.

The `url` of every locale that has one carries a path `reserved`, and so does the `email` of `da_DK`, `de_DE`, `en_US`,
`et_EE`, `fi_FI`, `fr_FR`, `lt_LT`, `lv_LV`, `nb_NO`, `nn_NO` and `sv_SE`. Each `reserved` draws only domains nobody can register:
`example.com`, `example.org` and `example.net`, or a subdomain of one, which [RFC 2606](https://www.rfc-editor.org/rfc/rfc2606)
reserves, so `{/sv_SE.url.reserved}` never reaches a real site and `{/sv_SE.email.reserved}` never a real mailbox.
The default of each of those `url`s also draws domains anyone may register, such as `example.se` and `example.io`, and so
do the default `sv_SE.email` and `en_US.email`; the other locales' default `email` draws only reserved domains.

The dates of `da_DK.date`, `de_DE.date`, `en_US.date`, `et_EE.date`, `fi_FI.date`, `fr_FR.date`, `lt_LT.date`, `lv_LV.date`, `nb_NO.date`, `nn_NO.date` and `sv_SE.date` are uniform over 1970-01-01 to 2029-12-31, `misc.datetime` over 2000-01-01 to 2029-12-31.

#### Names and sex

`first-name` and `last-name` are tables weighted by bearers, and `sex` by the people of each sex, from
SCB, Statistics Norway, DVV, Statistics Finland, INSEE, the SSA and the Census Bureau, for Germany from the city of Munich's newborn names and a
language model's recall, for Denmark from Statistics Denmark's top 20 and a
language model's recall, for Latvia's first names from the Office of Citizenship and Migration Affairs' register statistics, for Lithuania's first names from the Centre of Registers' top 50 and a language model's recall, and for Estonia's names and the last names of Latvia and Lithuania from a language model's recall alone (the weights of the recalled names are estimates). `sex` counts the people of each
sex in its `count` column: Sweden's population at the end of 2024, Finland's at the end of 2025, Norway's and Denmark's on 1 January 2026, Germany's, Estonia's, Lithuania's, Latvia's and France's on 1 January 2025 (Eurostat) and
the United States' on 1 July 2024. `first-name` links to `sex`, so `sv_SE.sex[f].first-name`
draws a woman's name, and a name both sexes carry is a row under each, so
`en_US.sex[m].first-name[Taylor]` names the one a `first-name[Taylor]` alone cannot.
`person` binds one first name and reads its sex through it, `{.first-name as n}` and
`{n..sex.name}`, so its `first` and `sex` columns agree. `en_US.title` links to `sex`

too, so a person's prefix never contradicts it.

A middle name is a second draw under the first name's sex:

```json
"{/sv_SE.first-name as n}{n} {n..sex.first-name} {/sv_SE.last-name}"
```

Renders e.g. `Louice Kajsa Grönberg`. About one draw in 220 repeats the first name.

### Locale pages

Each folder of `data/` has a page under [`docs/locales`](docs/locales) with its notes: where its data comes from, which
checker accepts its ids, and how its companies, phone numbers and words are drawn. The `geo` page holds the places of
every country that ships them; its [Swedish places](docs/locales/geo.md#swedish-places) says why a street may sit under a
neighbouring postort.

| Folder | Page | Sections |
|---|---|---|
| `da_DK` | [Danish](docs/locales/da_DK.md) | CPR number; CPR number matched to a sex; Danish companies; Danish dates and emails; Danish phone numbers; Danish words |
| `de_DE` | [German](docs/locales/de_DE.md) | Steuer-ID; German companies; German dates and emails; German phone numbers; German words |
| `en_US` | [US English](docs/locales/en_US.md) | English words; Social Security numbers and ITINs; US phone exchanges |
| `et_EE` | [Estonian](docs/locales/et_EE.md) | Estonian companies; Estonian dates and emails; Estonian phone numbers; Estonian words |
| `fi_FI` | [Finnish](docs/locales/fi_FI.md) | Finnish companies; Finnish dates and emails; Finnish phone numbers; Finnish words; Henkilötunnus; Henkilötunnus matched to a sex; Henkilötunnus of the issued range |
| `fr_FR` | [French](docs/locales/fr_FR.md) | NIR; French companies; French dates and emails; French phone numbers; French words |
| `geo` | [Places](docs/locales/geo.md) | Across countries; Danish places; Estonian places; Finnish places; French places; German places; Latvian places; Lithuanian places; Norwegian places; Swedish places; US places |
| `lt_LT` | [Lithuanian](docs/locales/lt_LT.md) | Lithuanian companies; Lithuanian dates and emails; Lithuanian phone numbers; Lithuanian words |
| `lv_LV` | [Latvian](docs/locales/lv_LV.md) | Latvian companies; Latvian dates and emails; Latvian phone numbers; Latvian words |
| `misc` | [Shared by every locale](docs/locales/misc.md) | Reserved addresses; Credit cards and IBANs; Registry tables; Territories; Time zones |
| `nb_NO` | [Norwegian Bokmål](docs/locales/nb_NO.md) | Fødselsnummer; Fødselsnummer matched to a sex; Norwegian companies; Norwegian dates and emails; Norwegian phone numbers; Norwegian words |
| `nn_NO` | [Norwegian Nynorsk](docs/locales/nn_NO.md) | Persons, ids, contact details and companies; Nynorsk months, weekdays and titles; Nynorsk words |
| `sv_SE` | [Swedish](docs/locales/sv_SE.md) | Bankgiro, plusgiro and bank accounts; Blood types; Cars; Industries and occupations; Licence plates; Months, weekdays, holidays and territories; Organisationsnummer and VAT numbers; Person; Personnummer and samordningsnummer; Personnummer and samordningsnummer, twelve digits; Personnummer of a sex or an age band; Swedish companies; Swedish phone numbers; Swedish words |

## Data format

Every value is a **node**, nestable without limit:

| Node | JSON | Renders |
|------|------|---------|
| string | `"Malmö"` | its text, with any `{…}` tokens expanded |
| choice | `["a", "b", …]` | one item, picked at random |
| template | `{"format": "…", …}` | its format, with `{name}` tokens rendering the named fields |
| table | `{"format": "…", "rows": "x.tsv", …}` | its format over one row of the TSV beside it ([Table](#table)) |
| number | `5`, `1.50` | its text as written |
| boolean | `true` | its text as written |

In a record column a bare number or boolean is refused, since it could be the text `"5"`
or the typed value `{"format": "5", "datatype": "integer"}`. In a struct tag, the
field's Go type decides.

### Format string

Every character is literal except a `{…}` token:

| Token | Renders |
|-------|---------|
| `{name}` | the sibling field `name`, or the pick bound to the name ([Names](#names)) |
| `{x as name}` | nothing: binds one pick of `x`, a reference, a sibling field or a path into one, to `name` ([Names](#names)) |
| `{name.field}` | `field` of a fresh draw of the sibling `name`, or of the pick bound to it ([Correlated fields](#correlated-fields)) |
| `{a\|b}` | one of the named fields, even odds |
| `{fn(args)}` | a builtin ([Functions](#functions)) |
| `{/path}`, `{.path}`, `{..path}` | a node reached from the data root, this file's folder, or the folder above ([References](#references)) |
| `{{`, `}}` | a literal `{` or `}` |

Each `{…}` draws afresh, so `{w} {w}`, `{p.a} {p.b}` and `{/person.first} {/person.last}`
are two draws each. Only a [name](#names) keeps one pick for every read of it.

A token's sibling fields are the fields of the template whose format holds it, and a token reads
no other field. In `{"format":"{a}","a":"x{b}","b":"y"}`, `a` has no fields, so its `{b}` fails to
load. Give `a` a `b` of its own, `"a":{"format":"x{b}","b":"y"}`, or bind `b` in the outer format,
beside `a`, and read the name: `{"format":"{b as n}{a}","a":"x{n}","b":"y"}`. Inside a choice's
item, give the field its own.

```json
"100 Main St, Apt {int(1,9)}{upper(1)} — tel {digits(3)}-{digits(4)}"
```

Renders e.g. `100 Main St, Apt 4B — tel 555-0199`. Random characters come from
the sample functions, so a phone pattern is `070-{digits(3)} {digits(2)} {digits(2)}`.
A lone `}` is a load error naming `}}`. A repeated arm counts as written, so
`{a|a|b}` draws `a` two times in three.

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
writing it as `{ "format": "AB", "weight": 3 }`. A repeated item counts as written,
so `["a", "a", "b"]` draws `a` two times in three, and an item of weight `0` is never
drawn. A weight outside a choice does nothing, and so does one above `0` on the item of
a one-item choice. Rejected at load: a weight that is negative, not a number, or too
close to `0` to tell from it, such as `1e-400`; and a choice whose every weight is `0`, a one-item choice of weight `0` among them.

### Repeat

A template may carry `repeat` (an integer of 0 or more, default `1`; `0` renders nothing) to render its format
that many times — each an independent draw — joined by `separator` (default `""`):

```json
{ "format": "{word}", "repeat": 3, "separator": " ", "word": ["foo", "bar", "baz"] }
```

Renders e.g. `bar foo baz`. A `separator` without a `repeat` above `1` does nothing.
Rejected at load: a `repeat` that multiplies to more than 1 048 576 renders along any
path of nested repeats.

### Datatype

A record column may declare `datatype` — `integer`, `number` or `boolean` — so `json`
writes `42` rather than `"42"` and `sql` a bare literal; a column without one, or with
`string`, is a string:

```json
{ "format": "",
  "id": { "format": "{seq()}", "datatype": "integer" },
  "paid": { "format": "{p}", "p": ["true", "false"], "datatype": "boolean" },
  "total": { "format": "{calc(net * qty, 2)}", "net": ["19.99", "5.00"], "qty": ["3", "7"], "datatype": "number" } }
```

Writes e.g. `{"id":1,"paid":true,"total":59.97}`. A column is a field of the top-level
template, or an item of a choice standing in for one. A field of a template that stands as a
column counts too, for a name that reads it, as below. `datatype` anywhere else does
nothing. A typed column holds one value, alone in its format: a
literal, one `{int()}`, `{float()}`, `{seq()}` or `{calc()}` call, or a read that lands only on such
values. `integer` is an int64 written `-?(0|[1-9][0-9]*)` — `{float()}` prints one at
`0` decimals within int64 — `number` a JSON number, `boolean` `true` or `false`. A value its
datatype cannot hold is a load error naming it:

```text
order.id: datatype integer: {digits(3)} prints text, not an integer
order.id: datatype integer: "1{digits(2)}" is not one value; write one literal or one {float()}, {int()}, {seq()} or {calc()}, or read one
```

A column whose format is one read of a record's column and nothing else —
`"score": "{/src.score}"`, `"score": "{s.score}"` with `s` bound to `/src`, `"a": "{p}"`
beside `{place as p}` in its own record, or `"len": "{p.len}"` beside `{pattern as p}` where
each item of the field `pattern` has a `len` — is that column: it takes the column's datatype and
is null where the column is, and a struct field tagged `src.score` is nil there. The `len`
of every item `p` can bring must hold one datatype, or the load fails; one may be null.
`sched.json` shows it:

```json
{
  "format": "{pattern as p}",
  "pattern": [
    { "format": "daily", "len": null },
    { "format": "weekly", "len": { "format": "{int(1,9)}", "datatype": "integer" } }
  ],
  "len": "{p.len}"
}
```

`fejkdata --seed 1 --format ndjson --repeat 3 < sched.json` prints:

```text
{"len":1,"pattern":"weekly"}
{"len":null,"pattern":"daily"}
{"len":null,"pattern":"daily"}
```

`len` is an integer column, null for `daily`.

A `datatype` of its own types the column's values where they prove it, and `"string"`
reads them as text. Any other read renders the column's text, a null as `""`.

A typed column's `{calc()}` must be proven to print a number: each operand a number
literal, an `{int()}`, `{float()}`, `{seq()}` or `{digits()}` call, a calc, or a read of
such values, whose bounds keep every divisor from zero and the result within `1e300`.
What the bounds cannot show is refused — `{calc(a / b)}: divides by b, which is not
proven nonzero`. The calc fills an `integer` column at `0` decimals, or over whole
operands with no `/`, while its bounds stay within int64.

### Null

A `null` item draws a record column as null: `json` writes `null`, `sql` `NULL`, `pgcopy` `\N`, and
`csv` an empty field, with an empty string written `""` — the convention PostgreSQL's
`COPY … CSV` reads. A record of one null column is a blank line, which `COPY` reads as
null in CSV mode but most CSV readers skip, so write such a record as `json`, `sql` or `pgcopy`. `Fake`
renders a null as `""`. The other items' weights skew its odds:

```json
{ "format": "", "deleted_at": null, "middle": [null, { "format": "{n}", "n": ["Ann", "Eva"], "weight": 3 }] }
```

`deleted_at` is null every draw, `middle` a name three draws in four. Anywhere but a
column, `null` renders `""`. Rejected at load: a column whose items hold different
datatypes. A read through a name keeps the null too ([Datatype](#datatype)).

### Hidden

`hidden` is an option holding fields that write no column. `--list` and the `.fejkdata.json` index
leave them out, at any depth and with every path below them, but a `{/path}` to one still draws
it, so a field only other fields read goes there. Save `ride.json`:

```json
{
  "format": "{day as d}",
  "hidden": { "day": ["Mon", "Tue", "Wed", "Thu"] },
  "name": "run {d}",
  "rule": "every {d}"
}
```

`fejkdata --seed 2 --format csv --repeat 3 < ride.json` prints:

```text
name,rule
run Tue,every Tue
run Mon,every Mon
run Tue,every Tue
```

A field is in `hidden` or beside it, never both, and an option name can't be a field in
`hidden`. A record holding every field in `hidden` has no column and fails with `ErrNoColumns`.

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
echo '{/country}' | fejkdata -d ./mydata                  # Sweden (SE), about half the time
echo '{/country.alpha2}' | fejkdata -d ./mydata           # NO
echo '{/country[SE]}' | fejkdata -d ./mydata              # Sweden (SE)
echo '{/country[Norway].alpha2}' | fejkdata -d ./mydata   # NO
echo '{/country}' | fejkdata -d ./mydata --format csv     # alpha2,name,population  →  SE,Sweden,10500000
```

`rows` names the TSV beside the category file; `key` names the column a path
selects a row by, `name` a column it also selects by, `weight` a column of numbers
of 0 or more that skews the draw, and `parent` the table a column links to
([Linked tables](#linked-tables)). The format's `{tokens}` read the columns, and the
columns are the [record](#records)'s columns, so `--format csv` writes the rows and
`--list` shows `country.alpha2`. A cell is a string node: `1{digits(2)} {digits(2)}`
in a cell draws digits and `{/misc.uuid}` reads a reference, while `{name}` in a cell
is refused, since a cell has no sibling. Each cell may select its own row of another
table, `{/misc.currency[SEK].symbol}` on one row and `{/misc.currency[EUR].symbol}`
on the next, since only one row renders. `New` proves the header, the options and every
cell token, and refuses:

- a JSON member other than `format`, `rows`, `key`, `name`, `weight` and `parent`, which
  could be meant as a field or as an option a table does not take;
- a key cell that is empty or repeats;
- a weight that is negative, not a number, or too close to `0` to tell from it;
- every weight `0`, across the table or inside one parent row;
- a key or name holding `[`, `]`, `{`, `}`, `"` or `|`, which a selector cannot spell.

`New` maps the keys, and a name or parent column is mapped on the first draw that selects
by name or descends through the table. A `name` needs a `key`, since a name naming several rows is reported by
their keys, or a `parent`, inside whose row a name names one row, so `first-name[Kim]`
is settled by the `sex` selected before it and a name repeating inside one parent row
is refused; a name spelling another row's key is refused, since the key would
select first and the name never. A table's options are its own, so a column may be named
`name`, as one usually is.

### Row selection

`[key]` or `[name]` after a table's name selects one row: `misc.territory[SE]` and
`misc.territory[Sweden]` name one row, and `misc.territory[SE].capital` reads its column.
A name naming several rows is an error listing their keys, unless a row selected
before it settles which ([Linked tables](#linked-tables)). A selector is part of the path, so it works
wherever a path does: `Fake`, `FakeRecord`, a `{/misc.territory[SE].capital}` reference
and a struct tag. A dot inside the brackets belongs to the key or name, so
`city[St. Louis]` selects it.

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
echo '{/country[SE].city}' | fejkdata -d ./mydata    # Stockholm or Göteborg
echo '{/country.city.name}' | fejkdata -d ./mydata   # a country drawn, then a city inside it
echo '{/city[Oslo].country}' | fejkdata -d ./mydata  # NO — the link column's cell
echo '{/city as c}{c.name}, {c..country.name}' | fejkdata -d ./mydata   # Oslo, Norway — one city and its country
```

A path descends from a row to a linked table by name, at any depth, and `--list`
advertises each direct step. One path reads one chain of rows: `country.city.name` draws a
country by its weights, then a city inside it by theirs. Two paths are two draws, so
`{/city.name}, {/country.name}` may name a city and another country. The reads of one
[name](#names) agree instead, and the table it binds decides whose weights govern:
`{/city as c}` draws a city by city weight and `{c..country.name}` is its country, while
`{/country as k}` draws a country by country weight and `{k.city.name}` a city inside it.
A name keeps every row its reads draw, so `{k.city.name}` and `{k.city.population}` describe
one city, and `{k.name}` its country.

`New` refuses a link cell that is no key of the parent, a parent row no child links to, a
chain of parents that closes, and a table named like a column of any table above it.

### Step up

`..` after a row steps up to the row its link column names. Name the parent table right
after the `..`:

```sh
echo '{/city[Oslo]..country.name}' | fejkdata -d ./mydata       # Norway
echo '{/city..country}' | fejkdata -d ./mydata                  # a city drawn, then its country: Sweden (SE)
echo '{/city[Göteborg]..country.city}' | fejkdata -d ./mydata   # Stockholm or Göteborg, drawn afresh
echo '{/city as c}{c}, {c..country.name}' | fejkdata -d ./mydata  # Oslo, Norway
```

It steps up from a selected row, a drawn row or a [name](#names) bound to a row, one
level per `..`: `geo.SE.locality..municipality..region.name`. `.country` after a row
still reads the link column's cell, `NO`.

After a `..`, `.` steps down again and draws afresh inside the row stepped up to, so
`city[Göteborg]..country.city` may draw Göteborg again. Under a name, the paths stepping
down from one `..` read one draw: `{c..country.city.name}` and
`{c..country.city.population}` describe one city.

A `..` that names anything but the parent table is an error naming the parent. A selector
after a `..` is an error naming the table to select from, since it could name a row outside
the one stepped up to. `--list` advertises no path with a `..`. At the start of a
reference, `..` is the folder above ([References](#references)).

### Options and fields

`format`, `weight`, `repeat`, `separator`, `datatype` and `hidden` are the only options, and
`let`, the old name of `hidden`, is refused; **any other key is a field** (see [Decisions](docs/decisions.md#options-and-fields-share-one-namespace)), and `rows` makes a category
a [table](#table), so no template carries a field of that name. `hidden` holds fields that write
no [column](#hidden). A [name](#names) shares the namespace too.

### Functions

A `{name(args)}` token calls a builtin. Arguments are checked at `New`: a bad
count, range, country or expression fails fast; an integer may carry a sign or leading
zeros (`+5` and `05` are `5`); a count of `0` renders nothing, except in `luhn(n)`, which refuses it; bounds are finite; and a length, count or decimal place beyond a sane maximum is rejected, so a
fat-fingered `hex(2000000000)` never tries to allocate gigabytes. Every builtin draws only from the seed, and a
time-based id takes its timestamp from the rng, not the clock. A builtin also receives the output its template has rendered so far, which `{fnr(x)}` uses to complete a number from the digits before it. `born` and a `date` bound relative to `today` also read the date, as their rows below say. The same seed and date repeat the output.

| Function | Kind | Renders |
|----------|------|---------|
| `{luhn()}`, `{luhn(n)}` | derivation | Luhn check digit over the digits emitted so far in this expansion, or over the last `n` digits of them |
| `{mod11()}` | derivation | weighted mod-11 check char (weights 2–7 from the right); `X` for 10 |
| `{ean()}` | derivation | EAN-13 / UPC-A / ISBN-13 / GTIN check digit |
| `{fnr(x)}` | derivation | completes a Norwegian fødselsnummer after the last seven digits emitted, its date and first individ digit, with a second individ digit it picks, the third that `x` reads, and the two check digits; fewer than seven digits read as if zeros came first, and `x` that is not one digit renders nothing |
| `{cpr(x)}` | derivation | completes a Danish CPR number after the last six digits emitted, its date, with a seventh digit that the CPR office's table of century digits allows for the year, an eighth and a ninth digit it picks, and the check digit, whose parity is that of the digit `x` reads; the year is the latest one not after today that bears the year digits; fewer than six digits read as if zeros came first, and `x` that is not one digit, or a year outside 1858 to 2057, renders nothing |
| `{nir()}` | derivation | the key of a French NIR: two digits, 97 less the remainder, divided by 97, of the last 13 letters and digits emitted, where the sixth and seventh may be `2A` or `2B`, counted as 19 or 18; nothing unless they are 13 digits once `2A` and `2B` are so counted |
| `{asmenskodas(x,d)}` | derivation | a Lithuanian asmens kodas, 11 digits: the digit for the sex that `x` reads, `1` for a man or `2` for a woman, raised by 2 for each century after the 1800s; the date that `d` reads, written `YYYYMMDD`, as `YYMMDD`; a serial number from 001 to 999 it picks; and the check digit. Prints no code for another `x`, or for a `d` that isn't a day from 1800 to 2099 ([Lithuanian personal codes](docs/locales/lt_LT.md)) |
| `{isikukood(x,d)}` | derivation | an Estonian isikukood, 11 digits: the digit for the sex that `x` reads, `1` for a man or `2` for a woman, raised by 2 for each century after the 1800s; the date that `d` reads, written `YYYYMMDD`, as `YYMMDD`; a serial number from 001 to 999 it picks; and the check digit. Prints no code for another `x`, or for a `d` that is no day from 1800 to 2099 ([Estonian personal codes](docs/locales/et_EE.md)) |
| `{personaskods(d)}` | derivation | a Latvian personas kods, 11 digits: for a `d` before 1 July 2017, written `YYYYMMDD`, the date as `DDMMYY`, `0`, `1` or `2` for the 1800s, 1900s or 2000s, a serial number from 000 to 999 it picks, and the check digit; for a `d` of 1 July 2017 or later, `32`, eight digits it picks and the check digit. Prints no code where `d` isn't a date from 1800 to 2099 ([Latvian personal codes](docs/locales/lv_LV.md)) |
| `{hetu(x)}`, `{hetu(x,'issued')}` | derivation | completes a Finnish henkilötunnus after the last six digits emitted, its date, with the century sign DVV gives the year, an individual number from 900 to 999 of the parity of the digit `x` reads, and the control character; `{hetu(x,'issued')}` takes the individual number from 002 to 899, which DVV issues; the year is the latest one not after today that bears the year digits; fewer than six digits read as if zeros came first, and `x` that is not one digit, or a year outside 1800 to 2099, renders nothing |
| `{bgmod11(x)}` | derivation | a random digit, then the check digit, that complete the digits emitted so far in this expansion to a number Bankgirot's 11-modul accepts: its digits, weighted 1 to 10 from the right and then 1 again, sum to a multiple of 11; the literal digits `x` count as coming first and are not printed (`0` for none) |
| `{digits(n)}` | sample | `n` digits 0–9 |
| `{upper(n)}`, `{lower(n)}` | sample | `n` letters A–Z, a–z |
| `{int(min,max)}` | sample | uniform integer in `[min, max]` |
| `{float(min,max,dp)}` | sample | number in `[min, max]` with `dp` decimals |
| `{hex(n)}` | sample | `n` lowercase hex digits |
| `{base64(n)}` | sample | `n` random bytes, base64 |
| `{steuerid()}` | sample | a German Steueridentifikationsnummer of ELSTER's test series, with a leading zero (see [Steuer-ID](docs/locales/de_DE.md#steuer-id)) |
| `{uuid()}` | sample | UUID v7 (`misc.uuid` ships v4 as data) |
| `{ulid()}` | sample | ULID, 26 Crockford base32 chars |
| `{nanoid(n)}` | sample | URL-safe Nano ID, `n` chars |
| `{iban(CC)}` | sample | mod-97-valid IBAN for BE, DE, DK, ES, FI, NO or SE, which a real account may hold, with the account check digits of BE, ES, FI and NO and a listed bank's IBAN ID in SE (see [Credit cards and IBANs](docs/locales/misc.md#credit-cards-and-ibans)); `misc.iban-test` draws published test IBANs |
| `{date(from,to,'layout')}` | sample | a second between two days, both included, each `YYYY-MM-DD`, or both `today` with a day count added or taken away, as `today-30` or `today+14`, from `WithToday` or `--today` ([Dates around today](#dates-around-today)), in a quoted Go layout: `'2006-01-02'`, `'January 2, 2006'`, `'060102'`, `'2006-01-02T15:04:05Z'` |
| `{born(min,max,'layout')}` | sample | a second in the calendar years in which someone turns `min` to `max`, both included, counted from the year of `WithToday` or `--today`, and never after that date; the layout as in `date` |
| `{adddays(d,n)}`, `{adddays(d,n,'layout')}` | computation | the date `d` reads plus `n` days, an integer, field, path or name, negative for earlier; the time of day stays, and the layout is `d`'s unless given ([Date offsets](#date-offsets)) |
| `{tick(start,step,jitter,'layout')}` | counter | a time that steps with each render: `start` plus `step` seconds per render and 0 to `jitter` seconds, below `step`, so renders ascend ([Time steps](#time-steps)) |
| `{time('layout')}` | sample | a second within a day: `'15:04'`, `'3:04 PM'` |
| `{seq()}`, `{seq(name)}` | counter | next integer from 1 in this generator; `name` selects an independent counter |
| `{calc(expr)}`, `{calc(expr,dp)}` | computation | an arithmetic expression over sibling fields and names ([Computation](#computation)) |
| `{lowercase(x)}`, `{uppercase(x)}`, `{ascii(x)}`, `{slugify(x)}` | transform | a field's value rewritten ([Transforms](#transforms)) |
| `{reword(text,words)}` | transform | each word of `text` replaced by one drawn from `words` ([Transforms](#transforms)) |

A derivation reads what is to its left, so place it after its payload; the
buffer is per expansion, so a nested template keeps fixed parts out of the sum. A
Swedish personnummer is a Luhn checksum over the nine digits before it, six of
them a birthdate:

```json
{ "format": "{born(0,99,'060102')}-{birth}{luhn()}", "birth": ["238", "239"] }
```

Renders e.g. `811218-2384`. `{luhn(9)}` leaves out what comes before the last nine digits, as a twelve-digit
personnummer's century stays out of its check digit: `{born(0,99,'20060102')}{birth}{luhn(9)}`. It is refused at
`New` where `n` is below 1 or above 1048576.

A layout is Go's: the reference time `Mon Jan 2
15:04:05 MST 2006` spelled as the output should look, quoted, since a layout may
carry the comma that separates arguments, with English names. In `date`, every second
between the two days is reachable, so a layout with a clock draws the time too, and
`from` may equal `to`, which is that one day. The instant is UTC, so a zone in the
layout prints `UTC` or `Z`.
The quotes delimit a layout outside a selector only, so `[O'Fallon]` in an
argument stays a name. A layout naming no field prints itself, `date` with a clock-only
layout prints the drawn instant's clock, and `date` over one day with no clock field prints
that day.
Rejected at `New`: a `date` bound that is no calendar date or `today` form, or after the other; an unquoted
or double-quoted layout, naming the single-quoted one; and `time` with a layout naming a
date field, naming `date`, since `time` draws no date.

`born` counts an age by the year of the date given to `WithToday` or `--today`, whether or not the
birthday has passed: with today at 2026-11-04, `born(18,64,'2006-01-02')` draws from 1962-01-01 to 2008-12-31,
and `born(0,17,'2006-01-02')` from 2009-01-01 to 2026-11-04. It is refused at `New` where an age is
negative, `min` is above `max` or `max` is above 999.

`{seq()}` spans `Fake`
calls and `repeat`, resets with a new generator, and is the natural primary key for
the SQL example above; give each record its own counter, `{seq(orders)}`, where one
generator fills several.

### Date offsets

`{adddays(d,n)}` prints the date `d` plus `n` days, so a fixture's end falls a drawn number of
days after its start. Read `d` through a [name](#names): the name keeps one draw, so the start
printed and the end computed from it share one date:

```json
{ "format": "{start as s}{len as n}{s},{n},{adddays(s,n)}", "start": "{date(2026-01-01,2026-12-31,'2006-01-02')}", "len": "{int(1,30)}" }
```

Renders e.g. `2026-03-28,5,2026-04-02`. A third argument prints the result in another layout:
`{adddays(s,-1,'January 2, 2006')}` prints the day before `s` as `March 27, 2026`.

Refused at `New`: a `d` that may render anything but a `date` or `born` call, such as
`"on {date(…)}"` or another `adddays`; a `d` whose calls print in more than one layout; a layout that
does not read back as the day it printed, such as `'01-02'` with no year or `'060102'` with a
two-digit one; an `n` that is not always a whole number; and an offset beyond a million days
either way.

### Dates around today

A `date` bound is a `YYYY-MM-DD` day, or `today` with a day count added or taken away, as `today-30`
and `today+14`. `today` is the date the caller gives, `WithToday` or `--today`, and both bounds take
one form. A due date between a month ago and a fortnight ahead stays current as the days pass:

```json
{ "format": "{date(today-30,today+14,'2006-01-02')}" }
```

Renders e.g. `2026-11-03` on 2026-11-12. A day and a relative bound in one call, a `from` after `to`
and an offset beyond a million days either way are refused at `New`. A bound stops at year 0000
or 9999.

### Time steps

`{tick(start,step,jitter,'layout')}` prints a time that steps with each render of the call, so one
per record in a `--repeat` run. Render n is `start` plus n−1 times `step` seconds, plus 0 to `jitter`
seconds drawn from the seed. `start` is a `YYYY-MM-DD` day, read as its midnight in UTC, or a UTC timestamp such as
`2026-01-01T00:00:00Z`. Renders ascend in time, `step` minus `jitter` to `step` plus `jitter` seconds apart:

```json
{ "format": "{seq()} {tick(2026-01-01T00:00:00Z,30,5,'2006-01-02T15:04:05Z')}" }
```

Renders e.g. `1 2026-01-01T00:00:03Z`, `2 2026-01-01T00:00:31Z` and `3 2026-01-01T00:01:02Z`.

Calls with the same four arguments step one counter, so the second of two such calls in a record
prints the next render's time; calls with other arguments count apart. To print one instant twice,
bind it from a field:

```json
{ "format": "{t as when}{when} {when}", "t": "{tick(2026-01-01,60,0,'15:04')}" }
```

A `jitter` of `step` or more could swap two renders, so it is refused at `New`, as is a `step`
outside 1 to 86400 seconds.

### Computation

`{calc(expr)}` evaluates `+ - * /`, parentheses and unary minus over number
literals, sibling fields and [names](#names), each rendered then read as a number;
a second argument rounds to that many decimals. A hyphen is always subtraction, so a
hyphenated field or name can't be an operand.

A calc computes from the values its format shows where both read one [name](#names):

```json
{ "format": "{net as n}{qty as q}{n} x {q} = {calc(n * q, 2)}", "net": ["19.99", "5.00"], "qty": ["3", "7"] }
```

Renders e.g. `19.99 x 3 = 59.97`.

A result that rounds to zero prints unsigned — `0`, `0.00` — as `{float()}`'s does. An
operand that is never a number is rejected at load: `"abc"`, `"NaN"`, `"Inf"`, or a read
whose every value is such fixed text. So is a divisor that is always zero, as in `1/0` or
a read whose every value is the fixed text `0`; a divisor that varies, as in `a/(b*c)` with
`c` varying, loads. An operand that is sometimes not a number yields `NaN`, and a divisor
that is sometimes zero `Inf`: both print rather than fail, except in a
[typed column](#datatype), which must prove neither happens.

### Transforms

`{lowercase(x)}`, `{uppercase(x)}`, `{ascii(x)}` and `{slugify(x)}` rewrite the value of `x` — a
field, a path, a `..path` or a [name](#names) — and nest. `ascii` folds the Latin letters of the default table of
ruby-i18n, Rails' translation library: Latin-1's and most of Latin Extended-A's (`Åsa Öberg` → `Asa Oberg`, `Wałęsa` → `Walesa`), and
drops any other non-ASCII rune, such as the `ễ` of `Nguyễn`. `slugify` folds as `ascii` does, except that it changes a non-ASCII rune `ascii` would drop
to `-`. It then lowercases, changes each run of characters outside `a` to `z` and `0` to `9`
to one `-`, and trims `-` from both ends: `Anna Lena` → `anna-lena`, `Åsa Öberg` →
`asa-oberg`, `O'Brien` → `o-brien`, `日本` → the empty string. Two non-empty slugs joined by
`.` make a valid email local part, the part before the `@`, even from a compound name such as
`von Walden`. An email matches the name beside it where both read one name:

```json
{ "format": "{person as p}{p.first} {p.last} <{slugify(p.first)}.{slugify(p.last)}@example.com>",
  "person": [
    { "format": "{first} {last}", "first": "Åsa", "last": "Öberg" },
    { "format": "{first} {last}", "first": "Bo", "last": "Ek" }
  ] }
```

Renders `Åsa Öberg <asa.oberg@example.com>` or `Bo Ek <bo.ek@example.com>`, never
a mix.

A word is a run of letters. `{reword(text,words)}` replaces each word of `text` with one drawn
from the path `words`, such as `/en_US.word`: a word of the same length, or else of the nearest
length `words` holds, the shorter on a tie. The fake takes the case of the replaced word's letter at
each position, or of its last letter beyond it. Digits, punctuation and every other character stay,
so `PROJ-123` keeps its digits:

```json
{ "format": "{reword(text, /en_US.word)}", "text": "Ping Anna: PROJ-123 is done, ANNA." }
```

Renders e.g. `Wolf Dirt: KNOT-123 at lazy, DIRT.` The fakes are random words, not a paraphrase.
Digits stay, so replace each phone number or personnummer in `text` yourself first, as
[Replacing real data with fakes](#replacing-real-data-with-fakes) shows. A word reads as one fake,
whatever its case, in every call of a render that reads the same list as `words`; a record renders
each row anew. A URL's scheme and host are words too, so to keep them, pass only the rest as
`text`. The field holding the original is a [format string](#format-string), so write each `{` and `}` in it as `{{` and `}}`.
`words` is a path, never a name, and must render only fixed text, as a [filter](#filters)'s path
must. It draws by weight among its values made of letters alone, and fails to load where it has
none. Neither argument may be a call, and no transform takes one, so `reword` doesn't nest.

### References

A reference renders a node from elsewhere in the data — the path `Fake` takes,
across every loaded source — so one category borrows another. The sigil says
where the path starts, as in a filesystem: `{/en_US.person}` from the data root,
`{.username}` from the folder this file sits in, `{..username}` from the folder
above. `sv_SE.email` can therefore read its own locale's `username`
without naming `sv_SE`:

```json
"Hej, {/en_US.person}!"
```

Renders e.g. `Hej, Pat Smith!`. Each reference draws afresh ([Format string](#format-string)),
so `{.person.first} {.person.last}` may name two people. Bind the person to a [name](#names) to read one: `{.person as p}{p.first}
{p.last}`. Rejected at `New`: a path that is
unknown, names a folder, has no folder above, reads a field not every variant
of a choice carries, and a reference that leads back to its own value, directly,
mutually or through a chain.

### Names

`{x as n}` binds one pick of `x`, a reference, a sibling field or a path into one, to the name
`n` and prints nothing. `{n}` then
prints that pick, and `{n.path}` reads a path through it, so every read of `n` describes one
row:

```json
"{/misc.territory as t}{t.flag} {t}: capital {t.capital}, calling code +{t.calling-code}"
```

Renders e.g. `🇳🇴 Norway: capital Oslo, calling code +47`. A name keeps what each `{n.path}`
lands on and every level it passes, so `{n}` and `{n.path}` agree, and so does a
[transform](#transforms) of `n.path`. Everything else under the name draws as it would
anywhere: if `n`'s category has the format `{w}-{w}`, it draws `w` twice, and `{n.w}` beside
`{n}` may then be refused: see the list under "Refused at `New`". A sibling field is bound the same way, so
`{place as p}{p.postal-code} {p.locality}` reads one place ([Correlated fields](#correlated-fields)).

A name is a pick of its own: `{t.name}` and `{/misc.territory.name}` beside it are two draws,
and so are two names bound to one category.

A name is drawn on its first read, and lives in the category binding it: any field of the
category may read it, and a record's columns read one pick. Each render of the category picks
anew, and so does each bare reference to it. A name a `repeat` binds picks again on every
iteration, and one bound outside the repeat keeps its pick on every line. A name a choice's
item binds lives in that item, and picks anew each time the item is drawn. A category a
reference enters does not see the name; bind it in the category that reads it.

Refused at `New`, each error naming what to write instead:

- a name bound twice outside any `repeat`, twice in one `repeat`, or both inside a `repeat`
  and outside it;
- a name that is a field or an option;
- a binding of anything but a reference, or a field of the template binding it or a path into
  one;
- a read of a name inside the field it binds: `{x as n}` with `"x": "{n}"` makes the pick read
  itself, so read `n` outside `x`. A binding and a read of it in one field or struct tag is fine:
  `{/sv_SE.sex as s}{s.name}`;
- reading a name outside the choice item that binds it, since the other items leave it
  unbound;
- a path through a name that selects a row: `{r.municipality[0180]}` could mean municipality
  0180 whatever region `r` picked, or 0180 only where `r` picked the region holding it;
- `{n}` beside `{n.w}` where `n`'s category reads `w` twice and the two reads could see different
  draws: `{w}-{w}` where `w` isn't a constant, or `{w.a} {w.b}` where `w` is a choice or `{n.w}`
  reads it whole. `{w.a} {w.b}` beside `{n.w.z}` loads where `w` is a template, an object with a
  `format`, and so does `{w}-{w}` beside `{n.w}` where `w` is a constant;
- `{n.w}` beside a read of `n` that reads `w` through a name of `n`'s category, where the two share
  a draw: `sv_SE.company`'s format is `{registered as r}{r}`, so
  `{/sv_SE.company as c}{c} {c.registered}` could print two companies. Drop `{c.registered}` to
  keep one pick, or write `{/sv_SE.company.registered}` for a second draw;
- `{n.path}` beside a read of `n` that renders a reference to `path`, or to a path above it, even
  through other references: a reference draws afresh, so `{c}` and `{c.ab.name}` can differ in
  `{/sv_SE.company as c}{c} {c.ab.name}`, where `sv_SE.company`'s items read `{.company.ab.name}`.
  Drop `{c.ab.name}`, or write `{/sv_SE.company.ab.name}` for a second draw. The same holds for a
  `path` that runs into a table linked to the one `n` binds: `{/region as r}{r} {r.municipality.name}`,
  where the table `region`'s format is `{/region.municipality.name}`. The same holds for a reference
  in a cell of the table `n` binds: `{/place as n}{n.tag} {n.zip}`, where a `tag` cell reads
  `{/place.zip}`. The same holds for a `path` that lands on a level above the reference's target, where that level's format reads the target:
  `{/region as r}{r} {r.municipality}`, where `municipality`'s format is `{name}`. A level whose
  format doesn't read the target is fine, and so is a reference to a constant, with no choice on
  the way;
- two reads of `n` where one reads a field from the pick and the other draws that field afresh:
  where `n`'s format is `{w.a}` and `w`'s is `{a}`, `{n}` reads `w.a` from the pick, but `{n.w}` draws
  its own `a`, so an `a` that isn't a constant could print two values. Read `{n.w.a}` in place of
  `{n.w}` to keep one draw, or write `{/cat.w}` for a second draw, where `n` binds `/cat`. A constant
  `a` is fine;
- a path through a name bound to a field holding a choice, where the path steps into an item and
  the field it lands on reads a name that item binds: `{r.legal}` where each item binds `w` and
  its `legal` reads `{w.legal}`. Bind `w` outside the choice, or read `{r}` whole;
- a binding in a table's format or cell.

### Filters

A `where` clause after a path, or after `{base64(n)}`, `{digits(n)}`, `{hex(n)}`, `{lower(n)}`,
`{nanoid(n)}`, `{ulid()}`, `{upper(n)}` or `{uuid()}`, draws only among the values that pass it:

```json
"{/en_US.word where length 5 and not containing 'x'}"
```

Renders e.g. `crane`. A clause holds one or more conditions, joined by `and`:

- `length 5`, or `length 4 to 6` with both ends included, counts a value's Unicode code points.
  A length may read a [name](#names), as in `{n as k}{/en_US.word where length k}`.
- `not 'cat', 'dog'` drops those values, each compared whole.
- `not containing 'cat'` drops every value holding `cat`, `concatenate` included.

Matching is case-sensitive. A quoted value can't hold an apostrophe. A path's values that pass keep
their [weights](#weight). A call's draw never retries: it draws each character evenly among those
that still lead to a value passing the clause, so the values that pass aren't equally likely. `as`
after a path's clause binds the filtered draw: `{/en_US.word where length 5 as w}{w} {uppercase(w)}`.

`not` and `not containing` may also read a name, unquoted, and take the string it holds in that
render:

```json
{ "format": "{code as c}{c} → {hex(4) where not containing c}", "code": ["7958", "beef"] }
```

Renders e.g. `7958 → 3fa0`: the hex never holds the code printed before it. To leave out each
member of a name bound by `distinct`, list the members: `not containing m[1], m[2]`. `not
containing` skips an empty string, which every value holds.

Quoted strings are known at load, so only a string read from a name can leave no value to draw.
`else`, last in the clause, names the fallback that prints then: a quoted text, as in
`{w where not t else 'none'}`, or a name. Without `else`, a path prints an empty string, and a call
its default, a value of its own shape:

| Call | Prints |
| --- | --- |
| `{base64(n)}` | n zero bytes, encoded, as `AA==` for `{base64(1)}` |
| `{digits(n)}`, `{hex(n)}` and `{nanoid(n)}` | n zeros |
| `{lower(n)}` | n `a`s |
| `{ulid()}` | 26 zeros |
| `{upper(n)}` | n `A`s |
| `{uuid()}` | the Nil UUID, `00000000-0000-0000-0000-000000000000` |

The fallback and the default print unchecked, so a render never fails: the Nil UUID holds `0000`.
In a distinct draw, the fallback prints in place of each member left without a value. A fallback,
or a path's empty string, is text, even where `else` names a number. So a [typed](#datatype)
column refuses a clause that may print one, whether the column reads the clause directly or
through `calc`, and so does a `length` reading a name that binds the clause, or a field holding it.

Refused at `New`:

- a clause no value passes, even at one length a name it reads may print;
- a path whose values are not fixed text, such as `/misc.uuid`, whose format calls `{hex(n)}`, a
  null, a repeat, or a format reading two fields, save a table's format reading only its own
  columns, as `misc.car`'s does;
- a path through a selector, a linked table or `..`;
- a clause on any other call, on `{a|b}`, or in a table's format or cell;
- `distinct` on a call;
- `as` after a call's clause: put the call and its clause in a field, and bind the field;
- a clause on a name, which is drawn already: write it on the draw the name binds;
- a path after the name a clause binds, such as `{w.x}`: read `{w}` whole;
- a name bound by `distinct`, read whole in `not` or `not containing`: list its members, as
  `not containing m[1], m[2]`;
- a field read unquoted in a clause: for one of its values, bind it first, as `{t as n}`, and
  read `n`; for every value, quote each, as `not containing '7958', 'beef'`;
- a clause reading the name its own token binds, directly or through the clauses of the names it
  reads, as `{w where not u as u}`: drawing the name would read itself.

### Distinct draws

`distinct N` after a path draws N different values in one pick. `joined ','` prints them as one
value, with `,` between each two:

```json
{ "format": "BYDAY={day distinct 2 joined ','}", "day": ["MO", "TU", "WE", "TH", "FR"] }
```

Renders e.g. `BYDAY=TU,MO`, never `MO,MO`. `as` binds the pick to a name, and `{n[1]}` to `{n[N]}`
read its members, each as often as you like:

```json
"{/en_US.word where length 5 distinct 3 as w}{w[1]}, {w[2]} and {w[3]}"
```

Renders e.g. `crane, shore and plume`. With `joined` before `as`, `{n}` reads the members joined.
`distinct` comes after a [`where`](#filters) clause and before `as`, and draws among the values that
pass it.

Different means different values, whatever they print: two rows that both print `cat` may both be
drawn. Each value of a nested choice is a member of its own, so `/en_US.word`, a choice of four parts
of speech, draws different words. Each member is drawn by its [weight](#weight) among the values not
drawn yet, so a value of weight `0` never is.

Joined members are text: a column reading them is a string column, and `calc`, `adddays` and
`where length` refuse them. A single member, such as `{w[1]}`, is a value of the column it comes
from and keeps its type, so `calc(w[1]*2)` and `where length w[1]` read it.

Refused at `New`:

- more members than the path renders values of weight above `0`, or than pass its `where` clause,
  even at one length a name it reads may print;
- `distinct N` with neither `joined` nor `as`;
- `{n}` on a name bound without `joined`, a member outside `[1]` to `[N]`, and a path after the
  name, such as `{n[1].x}`;
- `distinct` on a name, which is drawn already: write it on the draw the name binds;
- everything a [filter](#filters) refuses.

### Correlated fields

`{name.field}` reads a path into a sibling. Several tokens read one draw of it through a
[name](#names): bind the sibling, and every read of
the name reads that one row — a locality and the postal code that really covers it:

```json
{ "format": "{place as p}{street} {int(1,99)}\n{p.postal-code} {p.locality}",
  "street": ["Kungsgatan", "Storgatan"],
  "place": [
    { "format": "{locality}", "locality": "Stockholm", "postal-code": "1{digits(2)} {digits(2)}", "weight": 975 },
    { "format": "{locality}", "locality": "Tranås",    "postal-code": "573 {digits(2)}", "weight": 18 }
  ] }
```

Renders e.g. `Kungsgatan 35` / `176 99 Stockholm`, never a Stockholm code beside
Tranås; each row's `weight` says how often it appears. `{p.geo.town.name}
{p.geo.town.zip}` share the town too, as a name keeps every level its reads pass. Every variant of a choice
on the path must carry the rest of it, so a row missing a field is named at load:

```text
token {p.postal-code}: not every variant of this 2-way choice carries "postal-code"; all carry [locality]
```

The sub-fields stay addressable — `Fake("address.place.locality")` renders, and
`List` advertises it. A path may not read into a level carrying a `repeat`, and `List`
advertises no such path; read that level whole.

### Ratios and rules

A ratio is a [weight](#weight). Eighty percent `active` and twenty percent `inactive`:

```json
[
  { "format": "active", "weight": 80 },
  { "format": "inactive", "weight": 20 }
]
```

Weights are relative, not percentages: `4` and `1` give the same odds as `80` and `20`.

There is no `if`: a rule such as "if the country is `SE`, the currency is `SEK`" is data. Write the country
and its currency in one row, and bind the row to a [name](#names), as in
[Correlated fields](#correlated-fields):

```json
{ "format": "{row as r}{r.country} pays in {r.currency}",
  "row": [
    { "format": "{country}", "country": "SE", "currency": "SEK", "weight": 3 },
    { "format": "{country}", "country": "NO", "currency": "NOK", "weight": 2 },
    { "format": "{country}", "country": "DK", "currency": "DKK" }
  ] }
```

Renders `SE pays in SEK` three times in six, and never `SE pays in NOK`.

### Performance

After a category's first load, the target is about a microsecond per value (goal 12.3).

A source with an [index](#indexed-sources), as the shipped set has, loads nothing in `New`: a
category loads on the first `Fake`, `FakeRecord`, `FakeStruct` or template that reads
it, with every category it reads and its whole table family. Each file is parsed,
validated and weight-indexed once, bar a table's name and parent columns, which are
mapped on the first `Fake` that selects by name or descends through it. A `Fake` call then
costs about what its output costs: an unweighted pick is O(1) whatever the list's length,
a weighted one O(log n), and long formats, deep nesting and many tokens add cost in
proportion to the output.

## Versioning

Semver tags, `v0.1.0` first, on `main` save a patch CI cuts for an update to Go, a module
or GoReleaser, which sits on a branch cut from the release before it; one version covers
the shipped data, the library and the CLI, and [`CHANGELOG.md`](CHANGELOG.md) names what
each release changed. A consumer's data, code and scripts keep working across a minor or a
patch. A minor only adds, and a major is the only release that changes what exists. Two
exceptions: any release, a patch too, may change what a seed draws; and before 1.0.0 a
patch built with a new Go minor may drop a system the binaries, packages and image run on,
such as an older macOS, and its changelog section names each one under Breaking. Where you
compare seeded output, install one exact version
(`go install github.com/larvit/fejkdata/cmd/fejkdata@vX.Y.Z`, the image
`larvit/fejkdata:X.Y.Z` or a release's binary; `go.mod` already pins one),
give a fixed `--today` or `WithToday` day, and regenerate seeded fixtures and expected values when you raise that version.

| Surface | Major | Minor |
|---------|-------|-------|
| Shipped data | remove or rename a path; change a category's format; remove a value, or change a weight or a repeat; add a reference from one shipped category into another; change a table's key, name, weight or parent column, or remove a row | a path outside a record's columns, a locale, a value in a list, a row |
| Records (a table is one) | remove, rename, retype or add a column; let a column be null | a record, as a new category |
| Data format | a fence: a spelling a load rejects that it accepted, in a category or a manifest; a template option, since it reserves a field name; change what a builtin returns for an input | a builtin |
| CLI | remove or rename a flag, or change its default; change what an exit code means; change the framing a `--format` writes (header, quoting, statement shape), the `--list` layout, or what an error names | a flag, a format |
| Library | change or remove an exported name; raise the lowest supported Go | an exported name, a `With…` option |
| Data packages | rename or remove a package; move a category to another package; make a package read another it did not | a package |
| Supported systems | from 1.0.0, drop a system the binaries, packages or image run on | a system |

A patch changes no row of this table: performance, docs, an update to Go, a module or
GoReleaser, or a fix inside a promised behaviour
that adds or removes no value and changes no path, format or spelling.

An error's wording may improve in a minor; the path, rejected spelling and replacement
it names may not.

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

App developers and their coding agents, writing tests and fixtures, in Go, in JavaScript
or TypeScript, and at a shell:

- a **bulk fixture author**, millions of rows into CSV or SQL
- a **Go test author**, filling a struct with `FakeStruct`
- a **JavaScript or TypeScript test author**, rendering templates from the npm package
  in Node, Bun, Deno or a browser
- a **hand fixture author**, a few values at a shell
- a **validator-facing author**, who needs a value a real checker accepts
- a **data author**, who writes categories under `--data-path` and meets every load
  error the fences raise
- a **custom-CLI builder**, who ships a CLI like `fejkdata` with data and functions of
  their own
- a **coding agent**, which runs the CLI for a developer: it finds a path with `--list`,
  and fixes a template from the error that exit 2 prints, which names the fix

and a **contributor**, who reads [`todo.md`](todo.md), [`AGENTS.md`](AGENTS.md) and
the Development section below, and who ships a register the eight above then draw from.

## Goals

Where the project is heading; the sections before Audience document what ships today. A
change that breaks a maintain goal is redesigned or dropped; the achieve goals are weighed
against each other in the order listed.

1. **Achieve:** **The best fake-data tool there is, and the one developers reach for.**
   - Signal: stars, image pulls and package installs rise.
   - Counter-signal: issues reporting a broken or wrong value rise.
   - Why: users won by breadth leave when the values let them down.
   1. **Achieve:** You can install it the way you install your other tools, and
      developers, and the coding agents working for them, hear about it where they look.
      - Signal: the channels carrying the current release rise.
      - Counter-signal: a channel serves an old release.
      - Why: a channel that lags the release is worse than none.
   2. **Achieve:** JavaScript and TypeScript code renders the same templates and data as
      Go, from an npm package that runs wherever JavaScript does.
      - Signal: the package's downloads rise.
      - Counter-signal: a library feature reaches JavaScript a release after Go.
      - Why: JavaScript users stay only while the package keeps pace with Go.
2. **Achieve:** **Enough data ships built in to fill a real test database: the CLI and the
   JavaScript package carry all of it, a Go program imports what it needs, and nothing is
   fetched at run time.**
   - Signal: the locales covering every common category rise.
   - Counter-signal: categories ship as stubs.
   - Why: a thin category is a gap the user hits later.
   1. **Achieve:** Each locale covers the common categories: names, addresses, phone
      numbers, ids, companies and dates. Locales come in this order: the Nordic
      countries, then the rest of the EU, then the rest of Europe, then the rest of the
      world.
      - Signal: the filled cells of the locale-by-category grid rise.
      - Counter-signal: a cell ships with a handful of values.
      - Why: coverage counts only where it fills a database.
   2. **Maintain:** A Go program carries only the data it imports.
3. **Achieve:** **Every value is one of three kinds, and the docs say which: real, a thing
   that exists, such as the territory `SE`; valid, made up but passing the check the real
   system applies, such as a Swedish personal number; or invented, such as a town that
   does not exist.** Real and valid values come first, and the best category offers more
   than one kind, so you choose.
   - Signal: the categories offering a real or a valid value rise.
   - Counter-signal: a value the docs call real or valid fails its real check.
   - Why: a kind promised but not delivered is worse than one plainly invented.
   1. **Maintain:** Values that belong together come from the same random pick. You
      choose which ones by giving that pick a name.
   2. **Achieve:** Shipped data comes from good published sources where one exists, and
      an open licence is enough, share-alike ones such as CC BY-SA 4.0 and ODbL included.
      - Signal: the tables naming the register they come from rise.
      - Counter-signal: categories users ask for stay unshipped for want of a source.
      - Why: a good source is preferred, not required.
   3. **Achieve:** Values look real by default. You can also ask for values that can
      never reach anything real, such as a domain nobody can register.
      - Signal: the categories offering a form that reaches nothing real rise.
      - Counter-signal: a default value reaches a real person or system.
      - Why: realism helps only while it stays harmless.
4. **Achieve:** **Templates are easy to write and read.** A template is the text you
   write to describe a value, such as `{first} {last}`.
   - Signal: readers seated as the audience predict what the README's examples print
     more often.
   - Counter-signal: the Data format section grows faster than the categories using it.
   - Why: ease bought with more syntax is harder to read.
   1. **Maintain:** Everything outside `{…}` prints exactly as written.
   2. **Maintain:** Each `{…}` makes a new random pick.
   3. **Maintain:** A name keeps one pick. `{/person as p}` picks a person and prints
      nothing. `{p}` then prints that person, and `{p.last}` their last name.
   4. **Maintain:** A path works like a folder path: `.` goes down a level, and `..` goes
      up one.
   5. **Maintain:** `[SE]` selects one row of a table, and means that row wherever you
      write it.
   6. **Maintain:** If two people could read a template differently, loading it fails,
      and the error shows how to write each meaning.
   7. **Maintain:** Whatever is well-formed and means one thing loads, even where part of
      it does nothing. fejkdata does not guess at mistakes.
   8. **Maintain:** When a clearer name turns up, it replaces the old one, even if that
      needs a new major version.
   9. **Maintain:** Everything fejkdata renders is written as a template: on the command
      line, in a struct tag and in the library.
5. **Achieve:** **Errors say plainly what is wrong and how to fix it.**
   - Signal: the errors that name the fix rise.
   - Counter-signal: an error grows past the one thing that is wrong.
   - Why: a long error hides its cause.
   1. **Maintain:** A template that could render a value its own rules forbid, such as
      `NaN` or `Inf`, fails to load instead.
   2. **Maintain:** Data that reads something nobody provided fails to load, and the
      error names the module that provides it by default, or says that none does.
   3. **Maintain:** A module from fejkdata is proven whole before it ships; a module from
      anyone else may load on first reach, and its author proves it.
   4. **Maintain:** Errors surface when the data loads, never mid-render.
6. **Maintain:** **One command line gets you a value.**
   1. **Maintain:** Flags work like in other command-line tools, and can go anywhere on
      the line.
   2. **Maintain:** Your first template of your own needs no escaping and no options.
   3. **Maintain:** The CLI fills in what you leave out with a sane default: all the
      shipped data, the system's randomness and today's date.
7. **Maintain:** **Source data is JSON or TSV, readable and editable by hand; only what
   data can't express, such as a check digit, is Go code.**
8. **Maintain:** **Flexible and extensible.**
   1. **Maintain:** Data and functions come in modules, a Go package or a data folder,
      from fejkdata or anyone else, and fejkdata's load as anyone else's do.
   2. **Maintain:** A module names the modules it reads by default, and anything
      providing the same paths can stand in for one.
   3. **Maintain:** Loading two things under one name fails, in the library and the CLI
      alike, unless you say the second replaces the first, and then every read of that
      name reads the second.
   4. **Maintain:** The builtins ship with fejkdata itself, so a template means the same
      in every program.
   5. **Maintain:** A function loaded from a module can ask for the input it needs, and
      gets the randomness and the date the generator uses.
   6. **Maintain:** A CLI carrying your own modules is a few lines of Go, and works just
      like `fejkdata`.
9. **Maintain:** **The same seed, date, data and versions always give the same output.**
   1. **Maintain:** In the library, in Go and JavaScript alike, the caller supplies
      everything that can change the output: data, seed and date.
   2. **Maintain:** The JavaScript package renders what the Go library renders, byte for
      byte.
10. **Maintain:** **Every runtime dependency is one the standard library can't replace,
    and the JavaScript package doesn't depend on any other npm package.**
11. **Achieve:** **The docs take a reader from their first value to the feature their job needs.**
    - Signal: readers seated as each persona in Audience find, from the README alone, the
      path and the feature a task of theirs needs.
    - Counter-signal: the README grows faster than what ships, or a rule is stated twice.
    - Why: docs bought by adding text make every answer harder to find.
    1. **Maintain:** Every feature has its own heading.
    2. **Maintain:** Every example runs as a test and shows its output.
    3. **Maintain:** Each rule is stated in one place.
12. **Achieve:** **It is fast enough that speed never rules it out: the first value
    appears at once, and bulk output stays within reach of the fastest general
    fake-data libraries, without giving up goal 3's values.**
    - Signal: the README's benchmark shows the rows a second beside other tools, and
      the gap narrows.
    - Counter-signal: a faster draw changes a value's kind or breaks the links
      between values.
    - Why: people rule a tool out by its slowest run, and speed that costs correct
      values loses them anyway.
    1. **Maintain:** A render's cost grows only with the data it reaches.
    2. **Achieve:** The CLI prints the first value of any category in under 100 ms,
       however much data ships.
       - Signal: the first value of each category stays under 100 ms as data is added.
       - Counter-signal: adding one country's data slows another category's first
         value.
       - Why: 100 ms is the limit under which an answer feels instant.
    3. **Achieve:** A million rows of an eight-column person record take under 10
       seconds on one CPU, and take less with each CPU you add.
       - Signal: the rows a second on that record rise.
       - Counter-signal: a faster path changes what a seed draws without the
         changelog saying so.
       - Why: people fill a test database with millions of rows, and a slow tool is
         replaced by a faster one.

### Documentation

How the README serves goal 11:

- The top shows the CLI and the library side by side, each a short quick start that
  renders the same value.
- A table of the shipped data follows: every path, an example value, its kind and its
  source.
- Recipes show common tasks end to end, such as filling a database or a Go struct.
- Edge cases and a category's caveats sit in the reference sections, below what a first
  reader needs.

## Layout

```
fejkdata.go     Generator, New, options, List
datatype.go     DataType, what a record column holds
render.go       Fake
record.go       records: Record, the JSON/CSV/TSV/SQL serializers, and their entry points
struct.go       structs: FakeStruct, and fake tags as one record's columns
inline.go       inline templates: Template, NewTemplate, FakeTemplate, IsTemplate, IsLonePath, PathName
cli/            the fejkdata command, as the library call Main
cmd/fejkdata/   the fejkdata binary: Main over every shipped module
internal/avoid/ the draw of a value of fixed runs of characters that holds none of a set of strings, planned over their Aho–Corasick automaton
internal/builtinfunc/ the {name()} functions: their checks and draws, checksums, date and time layouts, date offsets, transforms, reword's arguments and calc arithmetic, and, for each sample a where clause filters, the characters it draws and its default
internal/datafiles/ the walk of a data tree: its folders, its category files and the rows files beside them, and its manifest
internal/datatype/ the datatype a record column holds; the root's DataType is an alias of it
internal/drawstate/ the seeded randomness, the {seq()} counters and the date a generator draws through
internal/engine/ the load pipeline every entry point runs, a source's index and the lazy load of an indexed category on the first call reaching it, the draw of a caller's path and record, and the vocabulary the engine is written in
internal/engine/fence/ the proofs and fences a finished tree passes at load: no reference cycle, no repeat past the cap, where clauses that read no name they bind, filter no draw read through a path and have enough values to draw, {adddays()} day counts, record columns whose items agree on a datatype, and what a typed column, calc operand or struct field can hold
internal/engine/fence/names/ the rules that refuse a read of a name that could not keep one pick, and the walk they share of the reads a pick renders
internal/engine/load/ compiling JSON into nodes, its formats into ops, binding names, resolving references, linking tables, and the paths List advertises
internal/engine/render/ drawing a compiled tree: choices, format strings, expansions, paths and table steps, named picks and the frames they render in, where and reword draws, and a record's columns
internal/engine/tree/ the node types of the template tree, what a load found for them, and the methods every other package reads or sets a node through
internal/fejktest/ what the root's tests and data/'s tests share: the test day, Fake helpers, draw loops, the ids' checksums, a TSV's rows and a country's geo tables
internal/grammar/ how the template language is written: format tokens, where and distinct clauses, paths and selectors, reference sigils, identifiers, calc syntax, and whether a string is a template or a path
internal/invariant/ the one phrase every package panics with when an invariant breaks
internal/jsonvalue/ a decoded JSON value, its numbers kept as written, the one input a template compiles from, and the kind an error names it by
internal/proven/ what a proof knows of a value, and the bounds a calc takes from its operands
internal/rows/  a table's rows: the TSV, the options proved over it, the links between tables, row selection and draws, and the pin set one path or one named pick fixes
data/           the shipped modules, each a Go package embedding its categories (JSON, and a TSV per table) and its manifest; Modules, which lists them; and the tests of the shipped data
data-import/    the scripts that rebuild each sourced table (see DATA-LICENSES.md), each run by a command under Development, and the shared modules they import
dockerhub/      the Docker Hub page's short description and overview, which CI syncs from `main`
docs/           the decision log, the locale pages (`docs/locales`, one per folder of `data/`), the register research behind the shipped data, the survey of other libraries, panels' scores and rulings on how readable the code is, and what each chunk cost an agent to read
release-tooling/ what CI releases with: the release from the changelog heading; the patch of the last release for main's newer Go, modules or GoReleaser; a release's binaries, packages, image and package repositories; and the filer of known vulnerabilities as `todo.md` items
testdata/       the pinned shipped shape (see Versioning)
```

## Development

Everything runs in Docker — **no local tooling beyond Docker is needed**.
Source is bind-mounted; build caches persist in the `gocache` volume.

```sh
docker compose run --rm test    # go test -race
docker compose run --rm bench   # benchmarks
docker compose run --rm build   # compile the library
docker compose run --rm vet     # go vet
docker compose run --rm cyclo   # cyclomatic complexity over 7 (test files excluded)
docker compose run --rm dev     # interactive shell
```

`mutate` runs mutation tests with [Gremlins](https://gremlins.dev). It isn't one of the
checks a pull request must pass. Gremlins takes each Go line the branch changed since
`origin/main`, test files excluded, and swaps one operator in it at a time, such as `<`
for `<=`. A swap that no test fails on prints as `LIVED`: the tests run the line and
don't check what it does. A swap on a line no test runs prints as `NOT COVERED`. A run
over a branch that changed 61 lines in 7 Go files took a minute and a half. Nearly all
of that was the one run of the test suite Gremlins starts with.

```sh
docker compose run --rm mutate
```

In a git worktree, git keeps its data in the main checkout, outside the folder the
container mounts, and the command above stops with `an error occured while calling git
diff`. Pass the main checkout's git folder in `GIT_COMMON_DIR`:

```sh
GIT_COMMON_DIR="$(git rev-parse --path-format=absolute --git-common-dir)" docker compose run --rm mutate
```

Commands that rewrite source keep your file ownership when run with `--user`:

```sh
docker compose run --rm --user "$(id -u):$(id -g)" fmt       # gofmt -w .
docker compose run --rm --user "$(id -u):$(id -g)" generate  # go generate: each data/ module's .fejkdata.json
docker compose run --rm --user "$(id -u):$(id -g)" tidy      # go mod tidy
```

A `--user` command answering `permission denied` on `/cache` has met root-owned
files in the `gocache` volume: Docker creates the volume root-owned, and every
command here that omits `--user` writes into it as root. Hand it back, and again
whenever it recurs:

```sh
docker compose run --rm --user root --entrypoint chown test -R "$(id -u):$(id -g)" /cache
```

Every pull request must pass `docker build .` before it can be merged. CI skips the build
for a pull request that changes none of the files the build and its tests read, nor the
workflow itself. Otherwise CI runs the build's two halves, the targets `latest` and
`lowest-and-tooling`, on two runners. `latest` runs the format check, vet, the complexity
check and the Go tests under the race detector, on the latest supported Go.
`lowest-and-tooling` runs vet and the Go tests without the race detector, on the lowest
Go, which the Dockerfile's `LOWEST_GO_VERSION` names. That target also runs the release
tooling's tests, the geo imports' tests, the check of GoReleaser's config, and ruff and
`mypy --strict` over the Python in `data-import/` and `release-tooling/`. The rest of the gate is the Packages dry run below and the changelog check, which
CI runs against the PR base and which refuses a `## ` heading in `CHANGELOG.md` other than
`Unreleased` or a version with its date. The check `gate` passes once each of these has
passed or been skipped, and it is the only check that branch protection on `main`
requires. Run `docker build .` locally before pushing. It keeps Go's build and module
caches between runs, so a rerun tests only the packages a change reaches:

```sh
docker build .                                   # the whole build, on both Go versions
GO_VERSION=1.22.12 docker compose run --rm test  # the tests on the lowest Go, without the image build
```

The decision "A known vulnerability in the Go version or a required module blocks a
release, even in code fejkdata never calls, and blocks no merge", in
[docs/decisions.md](docs/decisions.md), says when CI checks for a known vulnerability and
how it files what it finds in `todo.md`. The pull request it opens from `security-plan` merges
itself, which needs "Allow auto-merge" on in the repository's settings. When the step
"Fetch govulncheck's report from vuln.go.dev" fails, vuln.go.dev did not answer: rerun
it. When GitHub pauses the daily run after 60 days with no activity, re-enable it under
Actions. The same check:

```sh
docker compose run --rm vulncheck
```

A pull request that changes a workflow or what a release packages also runs the Packages
workflow as a dry run. The `packaging` line of the `changes` step in
`.github/workflows/test.yml` lists those files. The dry run runs every Action a release
runs save `actions/deploy-pages`, with a token that can only read. It builds the binaries
for every platform, the packages and the images, then the apt, dnf and apk repositories,
signed with throwaway keys, and installs fejkdata from each repository. It doesn't publish
any of them or read a repository secret, so it runs on a pull request from a fork too.
The repository checks, locally:

```sh
docker build -f release-tooling/package-repos.Dockerfile --target throwaway-keys --output /tmp/fejkdata-keys .
PACKAGES_GPG_KEY="$(cat /tmp/fejkdata-keys/gpg.asc)" PACKAGES_RSA_KEY="$(cat /tmp/fejkdata-keys/rsa.pem)" \
  docker build -f release-tooling/package-repos.Dockerfile --target checks \
  --secret id=gpg,env=PACKAGES_GPG_KEY --secret id=rsa,env=PACKAGES_RSA_KEY .
```

A panel of simulated readers scores how easy the code is to understand, and
[`AGENTS.md`](AGENTS.md) states the 7.0 target, which blocks no merge. Goal 4 is checked
against simulated template writers from the Audience.

A change to the shipped data runs `generate` above first and commits each module's
`.fejkdata.json` with the change, then re-pins
[`testdata/shipped_shape.txt`](testdata/shipped_shape.txt) in its own commit:

```sh
REPIN=1 docker compose run --rm --user "$(id -u):$(id -g)" test
```

A shipped table built from a source is rebuilt by its script under
[`data-import/`](data-import), one command per dataset, from the source named in
[`DATA-LICENSES.md`](DATA-LICENSES.md). Run the block below in order: some scripts
read a table another one writes. `cldr.py` and `timezone.py` read
`misc/territory.tsv`. `company-se.py` reads the sv_SE names and `geo/SE`,
`company-no.py` the nb_NO names and `geo/NO`, `company-fr.py` the fr_FR names,
`company-ee.py` the et_EE names, `company-lt.py` the lt_LT names and `company-lv.py` the lv_LV names.

Downloads are cached under `data-import/cache/`, which version control ignores and
which grows past a gigabyte. Delete that folder to fetch afresh, or pass
`--cache DIR` to a script to reuse a copy of it you already have. The container sees
only the checkout, so a folder or file you pass to a script, such as `DIR`,
`--key-file` or `--surnames`, must sit inside it.

- `geo-se.py` needs a Trafikverket API key, free at
  [data.trafikverket.se](https://data.trafikverket.se/), in `TRAFIKVERKET_API_KEY`
  or a `--key-file`.
- The Census host behind `geo-us.py` and `names-us.py` rejects a client for a while
  after a burst, so give `names-us.py` a copy of the Census `names.zip` with `--surnames`.
- `geo-se.py` fetches SCB's småort polygons, about 120 MB, and its tätort polygons.
- `geo-us.py` fetches two TIGER/Line files per county it ships, about a gigabyte in all.
- `company-fr.py` fetches Sirene's StockUniteLegale, about a gigabyte.
- `geo-fr.py` fetches the Base Adresse Nationale's address file, about 940 MB.
- `geo-ee.py` fetches the address data's postal codes and the settlement units, about 45 MB, and Statistics Estonia's table RL21003.
- `geo-lt.py` fetches the Address Register's addresses and premises, about 105 MB, and the State Data Agency's table M3010202.
- `geo-lv.py` fetches the State Address Register's buildings and apartments, about 290 MB, and the Central Statistical Bureau's table RIG010.
- `company-lt.py` fetches the Register of Legal Entities' open data, 41 MB.
- `company-lv.py` fetches the Register of Enterprises' open data, 129 MB.
- `word-de.py` fetches Wikidata's lexeme dump, about 670 MB.
- `word-et.py` reads the same lexeme dump, so it downloads the 670 MB only where `word-de.py` hasn't.
- `word-fr.py` fetches kaikki.org's French dictionary, about 590 MB.
- `word-lt.py` fetches kaikki.org's Lithuanian dictionary, 84 MB.
- `word-lv.py` fetches Tēzaurs.lv's wordforms, 178 MB zipped and 14 GB unzipped, and reads them as a stream.
- `word-se.py` fetches SALDO's morphology, 254 MB.

```sh
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/territory.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-dk.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-ee.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-fi.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-fr.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-no.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/names-us.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-de.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-dk.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-ee.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-fi.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-fr.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-lt.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-lv.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-no.py
docker compose run --rm --user "$(id -u):$(id -g)" -e TRAFIKVERKET_API_KEY data-import data-import/geo-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/geo-us.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/bankaccount-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/car-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/cldr.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-ee.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-fr.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-lt.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-lv.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-no.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/company-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/currency.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/httpmethod.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/httpstatus.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/language.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/licenseplate-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/loglevel.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/mimetype.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/phone-area-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/phone-de.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/phone-dk.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/phone-lt.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/phone-no.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/port.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/protocol.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/sex-population.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/sni-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/ssyk-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/timezone.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/tld.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/useragent.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-de.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-dk.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-et.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-fi.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-fr.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-lt.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-lv.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-no.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-se.py
docker compose run --rm --user "$(id -u):$(id -g)" data-import data-import/word-us.py
```

To release, head `CHANGELOG.md` with the version's section in place of `Unreleased`
and merge: once `main` passes the gate, CI tags that commit `vX.Y.Z` and publishes
the GitHub release with the section as its body. A top heading of `Unreleased`
publishes nothing. CI then runs the Packages workflow on the new tag, which attaches the
binaries and packages, pushes the image to GHCR and copies it to Docker Hub, updates the
Homebrew tap and the Scoop bucket, and deploys the apt, dnf and apk repositories to GitHub
Pages. The workflow reads these repository secrets:

- `DOCKERHUB_TOKEN`: a personal access token of the Docker Hub account `larvit`, with Read &
  Write and no Delete. The repository variable `DOCKERHUB_USERNAME` names the account.
- `PACKAGES_GPG_KEY`, `PACKAGES_RSA_KEY`: the private keys that sign the repositories.
- `TAP_GITHUB_TOKEN`: a token that can push to `larvit/homebrew-tap` and
  `larvit/scoop-bucket`.

The image's Docker Hub repository and GHCR package must be public, and Pages must deploy
from GitHub Actions. To package a release again, run the workflow by hand with its tag; it
replaces what it uploaded before. For a tag older than the latest release it leaves the
tap, the bucket, the image's `latest` tags and the repositories alone, since each holds one
version.

Renovate merges an update itself once `gate` passes. These wait for a human:

- an update to `actions/deploy-pages` or `peter-evans/dockerhub-description`, since both
  publish, so no pull request runs them;
- one to `actions/upload-pages-artifact`, since only `actions/deploy-pages` reads its
  artifact;
- one to `actions/cache`, `reproducible-containers/buildkit-cache-dance` or the `busybox`
  image it copies with, since only a push to `main` runs their save step, and a broken one
  fails `gate` there and holds every release;
- a runner update, since jobs no pull request runs, such as the release job, run on the
  runner;
- a GoReleaser major, since it may rename the release assets, which the gate doesn't
  check.

An update to a tool, such as Python or `govulncheck`, isn't released. CI releases an update
to what a release carries as a patch of the last release. The decision "CI releases each
update to what a release carries at once, as a fejkdata patch cut from the last release's
tag", in [docs/decisions.md](docs/decisions.md), says what a release carries, how a patch
is built, and when one waits for you. Where the release job could not read a Go minor's
release notes, write what they require under Breaking yourself before you cut a release.

Renovate leaves the Dockerfile's `LOWEST_GO_VERSION` alone: raising it is a major.

The Docker Hub page's short description and overview are the files in `dockerhub/`. A push
to `main` that changes one syncs both to Docker Hub. The sync reads the secret
`DOCKERHUB_DESCRIPTION_TOKEN`: a personal access token of `larvit` with Read, Write &
Delete, the scope Docker Hub's API requires to edit a description. It lives in the
environment `dockerhub-description`, which only `main` may use, so the Packages workflow
cannot read it and the release holds no Delete scope.

### Add a locale

A pull request that adds a locale ships, in this order:

1. Its entry in `shippedLocales` ([`data/locales_test.go`](data/locales_test.go)), in name order,
   in a commit of its own: the tests every locale shares run once per entry. The entry sets a field,
   or names the category in `testedElsewhere`, for each shared category the locale ships, and a test
   fails until it does, or while a locale in `data/modules.go` has no entry.
2. The locale's package, added to [`data/modules.go`](data/modules.go).
3. Its `.fejkdata.json`, from `generate`, committed with the data.
4. The re-pin of `testdata/shipped_shape.txt` in a commit of its own, both as
   [the paragraph on a change to the shipped data](#development) says.
5. A raised `base` in `TestNoNewAllocRegression` (`perf_test.go`), where the new manifest paths
   take `New` past its 10% slack.
6. Its page, `docs/locales/<name>.md`, with a `#` heading naming the folder and a `##` heading per feature; its row in the "Locale pages" table, listing those headings; and its name in the "Data built in" bullet.
7. A `CHANGELOG.md` entry under `Unreleased`.
8. A [`DATA-LICENSES.md`](DATA-LICENSES.md) entry for each dataset the locale adds.

## License

MIT — see [LICENSE](LICENSE). Forked from [github.com/Timewave-AB/fakes](https://github.com/Timewave-AB/fakes).
