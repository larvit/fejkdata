# fejkdata

Fake data for tests. Related values match, like a postcode and its city. National ids pass
validation where possible.

```sh
pip install fejkdata
```

It needs Python 3.11 or later. It installs on Linux, macOS and Windows, on amd64 and arm64.
There is no source build, so pip finds nothing to install on another platform. The package
doesn't put a `fejkdata` command on your path.

```python
import fejkdata

print(fejkdata.render("{/sv_SE.person}", seed=1, today="2026-01-15"))
```

```text
Felix Östberg
```

`render` returns what the [fejkdata CLI](https://github.com/larvit/fejkdata#cli) prints
for the same template, seed and day. The package carries the CLI and all its data. It
doesn't fetch anything when it runs, and it doesn't depend on any other package.

## Arguments

`render(template, *, seed, today, repeat=1, format="text", table=None, data_path=None,
no_shipped_data=False, separator=None, null_text=None)` returns a `str`.

- `template` is a string. It holds a format string or JSON text. Pass a dict through
  `json.dumps` first.
- `seed` is a whole number from 0 up. You must pass it by name. The same seed and day give
  the same text.
- `today` is a `datetime.date` or a string like `"2026-01-15"`. You must pass it by name
  too. Ages and dates count from this day.
- `repeat` is how many times to render the template. It defaults to 1.
- `format` is `"text"` unless you ask for records. Records come as `"csv"`, `"json"`,
  `"ndjson"`, `"pgcopy"`, `"sql"` or `"tsv"`.
- `table` is the table `"sql"` and `"pgcopy"` write to. Without it, `{/sv_SE.person}` writes
  to the table `person`, its last segment without any `[…]` selector, and any other record
  to `records`. Any other format ignores `table`.
- `data_path` is a folder of your own categories, as a string or a `pathlib.Path`, or a list
  of them. The folders layer over the shipped data, and the last one wins on a clash.
- `no_shipped_data=True` loads only the `data_path` folders.
- `separator` is the text between repeated renders. It is empty by default, and the CLI
  refuses it with a record format.
- `null_text` is what `"csv"` writes for a null column, instead of an empty field. The CLI
  refuses it under any other format. Write a text such as `\N` as `r"\N"`, since `"\N"` is
  a syntax error in Python.

Each of the last five passes the flag of its name, such as `--null-text`, and the
[CLI's flags](https://github.com/larvit/fejkdata#cli) say more.

```python
import json

person = {"format": "{/sv_SE.person as p}", "first": "{p.first}", "last": "{p.last}"}
rows = fejkdata.render(json.dumps(person), seed=1, today="2026-01-15", repeat=1000, format="csv")
print(rows.splitlines()[:3])
```

```text
['first,last', 'Vinston,Larsson', 'Felix,Lind']
```

```python
sql = fejkdata.render(json.dumps(person), seed=1, today="2026-01-15", repeat=2, format="sql", table="people")
print(sql)
```

```text
INSERT INTO "people" ("first", "last") VALUES ('Vinston', 'Larsson'),('Felix', 'Lind');

```

## Errors

A render the CLI refuses raises `fejkdata.FejkdataError`. Its message is the CLI's error
text, and its `returncode` is the CLI's exit code.

The [README](https://github.com/larvit/fejkdata#readme) documents templates and the
shipped data.
