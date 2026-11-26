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

`render` returns what the [fejkdata CLI](https://github.com/larvit/fejkdata#cli) prints
for the same template, seed and day. The package carries the CLI and all its data. It
doesn't fetch anything when it runs, and it doesn't depend on any other package.

`repeat` renders the template many times, and `format` writes records as `csv`, `tsv`,
`json`, `ndjson`, `sql` or `pgcopy`:

```python
person = '{"format": "{/sv_SE.person as p}", "first": "{p.first}", "last": "{p.last}"}'
rows = fejkdata.render(person, seed=1, today="2026-01-15", repeat=1000, format="csv")
```

A template the CLI refuses raises `fejkdata.FejkdataError`. Its message is the CLI's error
text, and its `returncode` is the CLI's exit code.

The [README](https://github.com/larvit/fejkdata#readme) documents templates and the
shipped data.
