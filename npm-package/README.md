# fejkdata

Fake data for tests. Related values match, like a postcode and its city. National ids pass
validation where possible.

```sh
npm install @larvit/fejkdata
```

It runs in Node 22.3 or later, in Bun, in Deno and in a browser. It is an ES module with
TypeScript types.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
console.log(fejkdata.render("{/sv_SE.person}", { seed: 1, today: "2026-01-15" }));
```

```text
Felix Östberg
```

`render` returns what the [fejkdata CLI](https://github.com/larvit/fejkdata#cli) prints
for the same template, seed and day. The package is the CLI built to WebAssembly, with all
its data. It doesn't fetch anything else when it runs, and it doesn't depend on any other
package.

## Size and speed

An install downloads 18 MB and unpacks to 69 MB, nearly all of it the data.

`load` reads and compiles the WebAssembly, which takes about 0.3 seconds in Node. Call it
as often as you like: it loads once. In a browser it fetches `fejkdata.wasm` from beside
`index.js`, so serve the two together.

Each `render` starts from nothing and takes about 0.1 seconds before its first row. Ask for
many rows with `repeat`, not with many calls.

## Arguments

`render(template, { seed, today, repeat, format })` returns a string.

- `template` is a string. It holds a format string or JSON text. Pass an object through
  `JSON.stringify` first.
- `seed` is a whole number from 0 up. The same seed and day give the same text. Write a
  seed above 2^53 - 1 as a bigint, such as `9007199254740993n`.
- `today` is a string like `"2026-01-15"`, or a `Date`, whose UTC day is read. Ages and
  dates count from this day.
- `repeat` is how many times to render the template. It defaults to 1.
- `format` is `"text"` unless you ask for records. Records come as `"csv"`, `"json"`,
  `"ndjson"`, `"pgcopy"`, `"sql"` or `"tsv"`.

You must pass `seed` and `today`. A render never reads the clock or the system's randomness.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
const person = { format: "{/sv_SE.person as p}", first: "{p.first}", last: "{p.last}" };
const rows = fejkdata.render(JSON.stringify(person), { seed: 1, today: "2026-01-15", repeat: 1000, format: "csv" });
console.log(rows.split("\n").slice(0, 3).join(" | "));
```

```text
first,last | Vinston,Larsson | Felix,Lind
```

## Errors

A render fejkdata refuses throws a `FejkdataError`. Its message is the CLI's error text. Its
`exitCode` is the CLI's exit code: 2 for a mistake in the template or the arguments, 1 for
anything else.

```js
import { FejkdataError, load } from "@larvit/fejkdata";

const fejkdata = await load();
try {
	fejkdata.render("{/sv_SE.persn}", { seed: 1, today: "2026-01-15" });
} catch (error) {
	if (!(error instanceof FejkdataError)) {
		throw error;
	}
	console.log(error.exitCode, error.message);
}
```

```text
2 fejkdata: template: reference {/sv_SE.persn}: no entry "persn"
```

## Globals

The package loads through `wasm_exec.js`, the loader Go ships for its WebAssembly. It sets
`globalThis.Go`. Where the runtime has no `fs`, `process` or `path` global, it sets a stub
for each.

The [README](https://github.com/larvit/fejkdata#readme) documents templates and the
shipped data.
