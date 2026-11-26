# fejkdata

Fake data for tests. Related values match, like a postcode and its city. National ids pass
validation where possible.

```sh
pip install fejkdata
```

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

`render(template, *, seed, today, repeat=1, format="text")` returns a `str`.

- `template` is a string. It holds a format string or JSON text. Pass a dict through
  `json.dumps` first.
- `seed` is a whole number from 0 up. You must pass it by name. The same seed and day give
  the same text.
- `today` is a `datetime.date` or a string like `"2026-01-15"`. You must pass it by name
  too. Ages and dates count from this day.
- `repeat` is how many times to render the template. It defaults to 1.
- `format` is `"text"` unless you ask for records. Records come as `"csv"`, `"json"`,
  `"ndjson"`, `"pgcopy"`, `"sql"` or `"tsv"`.

```python
import json

person = {"format": "{/sv_SE.person as p}", "first": "{p.first}", "last": "{p.last}"}
rows = fejkdata.render(json.dumps(person), seed=1, today="2026-01-15", repeat=1000, format="csv")
print(rows.splitlines()[:3])
```

```text
['first,last', 'Vinston,Larsson', 'Felix,Lind']
```

## Errors

A render the CLI refuses raises `fejkdata.FejkdataError`. Its message is the CLI's error
text, and its `returncode` is the CLI's exit code.

The [README](https://github.com/larvit/fejkdata#readme) documents templates and the
shipped data.
