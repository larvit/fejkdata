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
for the same template, seed and day. The package is fejkdata built to WebAssembly, with all
its data. It doesn't fetch anything else when it runs, and it doesn't depend on any other
package. The [README](https://github.com/larvit/fejkdata#readme) documents templates and the
shipped data.

## Size and speed

An install downloads 19 MB and unpacks to 69 MB, nearly all of it the data. Install the
package as a dev dependency, `npm install --save-dev @larvit/fejkdata`, and keep it out of a
page's bundle.

`load` reads and compiles the WebAssembly, which takes about 0.3 seconds in Node. Every call
of `load` shares that one compile. In a browser it fetches `fejkdata.wasm` from beside `index.js`.
Serve `index.js`, `wasm_exec.js` and `fejkdata.wasm` from one folder, the last as
`application/wasm`.

A test runner that starts a worker per test file loads the 69 MB WebAssembly again in each
worker. That takes about 0.3 seconds a worker.

Each call of `render`, `fake`, `records` or `paths` loads the data it reads again, which takes
about 0.1 seconds before its first row. With `files`, each call also loads every file you
gave, read or not: about 0.65 seconds for a 0.47 MB file. Keep the files small.
Ask for many rows with `repeat`, not with many calls. One call returns at most 512 MB of
text. `records` counts 32 bytes for each cell on top of its text.

## Text

`load()` gives an object with four functions: `render`, `fake`, `records` and `paths`. None
of them keeps any state between calls.

`render(template, { seed, today, repeat, format })` returns a string.

- `template` is a string. It holds a format string or JSON text. Pass an object through
  `JSON.stringify` first.
- `seed` is a whole number from 0 to 2^64 - 1. The same seed and day give the same text. Write a
  seed above 2^53 - 1 as a bigint, such as `9007199254740993n`.
- `today` is a string like `"2026-01-15"`, or a `Date`, whose UTC day is read. Ages and
  dates count from this day.
- `repeat` is how many times to render the template. It defaults to 1.
- `format` is `"text"` unless you ask for records. A record is one render of a JSON
  template, with a column for each of its fields. Records come as `"csv"`, `"json"`,
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

## One value

`fake(path, { seed, today })` draws one value of a path. It refuses a template, so
`sv_SE.person as p` and `sv_SE.first-name where length 3` are errors.

Some paths are not where you might guess. A postcode and a city are parts of an address:

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
const day = { seed: 1, today: "2026-01-15" };
for (const path of ["sv_SE.address.postal-code", "sv_SE.address.locality", "en_US.address.postal-code", "en_US.address.region", "sv_SE.person.first"]) {
	console.log(path, fejkdata.fake(path, day));
}
console.log(fejkdata.paths().includes("sv_SE.postcode"));
```

```text
sv_SE.address.postal-code 854 01
sv_SE.address.locality Sundsvall
en_US.address.postal-code 29414
en_US.address.region SC
sv_SE.person.first Vinston
false
```

Each call draws afresh, so two calls give parts of two addresses. For a postcode and a city
that match, draw the whole address with `records("{/sv_SE.address}", …)`, as
[Rows as objects](#rows-as-objects) shows.

## Every path

`paths()` returns every path the loaded data holds, as `fejkdata --list` prints them.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
console.log(fejkdata.paths().filter((path) => path.startsWith("sv_SE.address.")).join(" "));
```

```text
sv_SE.address.locality sv_SE.address.municipality sv_SE.address.postal-code sv_SE.address.region sv_SE.address.street sv_SE.address.street-number
```

## Template or path

`isTemplate(text)` tells whether a string is a template or a path. A string holding a brace
is a template, and so is a JSON string or array. Use it to choose between `render` and
`fake` for a string a user gave. It needs no `load`.

```js
import { isTemplate } from "@larvit/fejkdata";

for (const text of ["sv_SE.person", "{/sv_SE.person}", "Hej {/sv_SE.first-name}", '["a", "b"]']) {
	console.log(text, isTemplate(text));
}
```

```text
sv_SE.person false
{/sv_SE.person} true
Hej {/sv_SE.first-name} true
["a", "b"] true
```

## Rows as objects

`records(template, { seed, today, repeat })` draws rows of a record template. The template
is JSON text or an object. A lone reference such as `"{/sv_SE.address}"` is allowed too,
and draws that path's record. Each row is an object with a key per column, in the template's
order. JavaScript puts a key that is a whole number, such as `"2"`, before the others.

A column's `datatype` sets its type in a row:

- An `integer` column is a `bigint`, so it stays exact above 2^53.
- A `number` column is a number, and a `boolean` column a boolean.
- A null is `null`.
- Any other column is a string.

A field that reads one reference, such as `"{/src.age}"`, takes the type of what it reads.
Give it a `datatype` to fix its type.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
const template = {
	format: "{/sv_SE.person as p}",
	id: { format: "{seq()}", datatype: "integer" },
	name: "{p.first} {p.last}",
	member: { format: "{m}", m: ["true", "false"], datatype: "boolean" },
};
for (const row of fejkdata.records(template, { seed: 1, today: "2026-01-15", repeat: 2 })) {
	console.log(typeof row.id, row.id, row.name, typeof row.member, row.member);
}
```

```text
bigint 1n Kenneth Lind boolean true
bigint 2n Ann-Charlott Wiklander boolean true
```

In TypeScript, a row's type follows the template's fields where the template is written
`as const`. A template held in a plain variable types every column `string | number | bigint |
boolean | null`.

A lone reference gives a row with a key per field of the record it reads:

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
const [address] = fejkdata.records("{/sv_SE.address}", { seed: 1, today: "2026-01-15" });
console.log(Object.keys(address).join(" "));
console.log(address["postal-code"], address.locality);
```

```text
street street-number postal-code locality municipality region
475 32 Öckerö
```

`JSON.stringify` and the snapshots of Node's test runner throw on a `bigint`. Turn it into a
string first, such as with `String(row.id)`.

## Your own data

`load({ data, files })` takes two options. Each call gives a new object.

- `data` names the shipped modules to load, such as `["sv_SE"]`. Each module loads the
  modules it reads too. `sv_SE` reads `geo.SE` for its addresses. Without `data`, every shipped
  module loads. A name that is not a shipped module throws, and the error lists the shipped
  modules.
- `files` maps a file name, such as `mydata/team.json`, to its text or bytes. A template
  reads that file as `{/mydata.team}`. The files lie over the shipped data, so a file at
  `sv_SE/color.json` replaces the shipped `sv_SE.color`. A file that doesn't load makes
  `load` throw, with the kind `"load"`.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load({ data: ["sv_SE"], files: { "mydata/team.json": '["Red", "Blue"]' } });
console.log(fejkdata.render("{/sv_SE.first-name} plays for {/mydata.team}", { seed: 1, today: "2026-01-15" }));
console.log(fejkdata.paths().includes("en_US.person"));
```

```text
Vinston plays for Red
false
```

## Shipped paths

The TypeScript type `ShippedPath` is every path of the shipped data, such as
`"sv_SE.first-name"`. An editor completes a string of that type. A template reads the
path as `{/sv_SE.first-name}`.

## Errors

`load` and each function it gives throw a `FejkdataError` when fejkdata refuses the
template, an argument or the data. Its message says what is wrong, in the CLI's words. Its
`kind` says where the error arose:

- `"misuse"`: a mistake in the template, an argument or `load`'s options.
- `"load"`: the data doesn't load, or the template reads a module `data` left out.
- `"limit"`: the output passes 512 MB.
- `"other"`: anything else.

Its `exitCode` is the CLI's exit code: 2 for misuse, 1 for anything else. `render` passes on
the CLI's own, which is 1 for a template mistake that only a format finds, where `records`
gives 2. So a test reads `kind`, not `exitCode`.

A test doesn't check `instanceof FejkdataError` either. A test runner or bundler may load two
copies of the package, and then `instanceof` fails.

A template that is neither a string nor, for `records`, an object throws a `TypeError`. So
does a call without the second argument.

```js
import { load } from "@larvit/fejkdata";

const fejkdata = await load();
try {
	fejkdata.render("{/sv_SE.persn}", { seed: 1, today: "2026-01-15" });
} catch (error) {
	if (error.name !== "FejkdataError") {
		throw error;
	}
	console.log(error.kind, error.exitCode, error.message);
}
```

```text
misuse 2 fejkdata: template: reference {/sv_SE.persn}: no entry "persn"
```

## Globals

The package loads through `wasm_exec.js`, the loader Go ships for its WebAssembly. It sets
`globalThis.Go`. Where the runtime has no `fs`, `process` or `path` global, it sets a stub
for each.
