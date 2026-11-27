import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync, rmSync, writeFileSync } from "node:fs";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { FejkdataError, load } from "./index.js";

const DAY = "2026-01-15";
const PERSON = '{"format": "{/sv_SE.person as p}", "first": "{p.first}", "last": "{p.last}", "born": "{born(18,90,\'2006-01-02\')}"}';
const fejkdata = await load();

// goCLI is what the Go build of the CLI, at $FEJKDATA_CLI, prints for template under flags.
function goCLI(template, ...flags) {
	return execFileSync(process.env.FEJKDATA_CLI, flags, { input: template, encoding: "utf8", maxBuffer: 1 << 28 });
}

for (const format of ["csv", "json", "ndjson", "pgcopy", "sql", "tsv"]) {
	for (const seed of [1, 7, 4242]) {
		test(`renders what the Go build prints, byte for byte, as ${format} under seed ${seed}`, () => {
			const want = goCLI(PERSON, "--seed", String(seed), "--today", DAY, "--repeat", "40", "--format", format);
			assert.equal(fejkdata.render(PERSON, { seed, today: DAY, repeat: 40, format }), want);
		});
	}
}

for (const locale of ["da_DK", "de_DE", "en_US", "et_EE", "fi_FI", "fr_FR", "lt_LT", "lv_LV", "nb_NO", "nn_NO", "sv_SE"]) {
	test(`renders the text the Go build prints for a person and an address of ${locale}`, () => {
		const template = `{/${locale}.person}, {/${locale}.address}\n`;
		const want = goCLI(template, "--seed", "3", "--today", DAY, "--repeat", "25");
		assert.equal(fejkdata.render(template, { seed: 3, today: DAY, repeat: 25 }), want);
	});
}

test("renders a hundred thousand rows as the Go build does", () => {
	const want = goCLI(PERSON, "--seed", "5", "--today", DAY, "--repeat", "100000", "--format", "csv");
	assert.equal(fejkdata.render(PERSON, { seed: 5, today: DAY, repeat: 100000, format: "csv" }), want);
});

test("the same seed and day render the same text, and another seed another", () => {
	const first = fejkdata.render("{/sv_SE.person}", { seed: 7, today: DAY });
	assert.equal(fejkdata.render("{/sv_SE.person}", { seed: 7, today: DAY }), first);
	assert.notEqual(fejkdata.render("{/sv_SE.person}", { seed: 8, today: DAY }), first);
});

test("a render doesn't leave anything behind for the next one", () => {
	const alone = fejkdata.render("{seq()} {/sv_SE.person}", { seed: 7, today: DAY });
	fejkdata.render("{seq()} {/en_US.person}", { seed: 9, today: "2001-02-03", repeat: 5 });
	assert.equal(fejkdata.render("{seq()} {/sv_SE.person}", { seed: 7, today: DAY }), alone);
});

test("load gives the same renderer each time", async () => {
	assert.equal(await load(), fejkdata);
});

test("load leaves no global of its own behind", () => {
	assert.equal("__fejkdataRun" in globalThis, false);
});

test("a load that fails is tried again by the next call", async () => {
	const fresh = await import("./index.js?failing-load");
	const { compile } = WebAssembly;
	WebAssembly.compile = () => Promise.reject(new Error("no file"));
	try {
		await assert.rejects(fresh.load(), /no file/);
	} finally {
		WebAssembly.compile = compile;
	}
	const loaded = await fresh.load();
	assert.equal(loaded.render("{/sv_SE.person}", { seed: 7, today: DAY }), fejkdata.render("{/sv_SE.person}", { seed: 7, today: DAY }));
});

test("a program that stops while loading fails the load, and leaves no global behind", async () => {
	const fresh = await import("./index.js?stopping-load");
	const { Go } = globalThis;
	const { instantiate } = WebAssembly;
	globalThis.Go = class {
		importObject = {};
		run() {
			return Promise.resolve();
		}
	};
	WebAssembly.instantiate = () => Promise.resolve({});
	try {
		await assert.rejects(fresh.load(), /stopped while loading/);
	} finally {
		globalThis.Go = Go;
		WebAssembly.instantiate = instantiate;
	}
	assert.equal("__fejkdataRun" in globalThis, false);
});

test("an output longer than a JavaScript string holds is refused, each time, and the next render works", () => {
	const wide = "x".repeat(70000) + "\n";
	for (let i = 0; i < 2; i++) {
		assert.throws(
			() => fejkdata.render(wide, { seed: 1, today: DAY, repeat: 10000 }),
			(error) => error instanceof FejkdataError && error.exitCode === 1 && /512 MB/.test(error.message),
		);
	}
	assert.equal(fejkdata.render("{/misc.territory[SE].alpha2}", { seed: 1, today: DAY }), "SE");
});

test("an error leaves out the CLI's advice about a shell", () => {
	assert.throws(
		() => fejkdata.render("{/sv_SE.personnummer-12}", { seed: 1, today: DAY, format: "csv" }),
		(error) => error instanceof FejkdataError && error.exitCode === 1 && !/\n|shell/.test(error.message),
	);
});

test("a seed is a number or a bigint, up to 2^64 - 1", () => {
	const top = 18446744073709551615n;
	assert.equal(fejkdata.render("{uuid()}", { seed: top, today: DAY }), goCLI("{uuid()}", "--seed", String(top), "--today", DAY));
	assert.equal(fejkdata.render("{uuid()}", { seed: 12n, today: DAY }), fejkdata.render("{uuid()}", { seed: 12, today: DAY }));
});

test("a number too large to be exact is refused, naming bigint", () => {
	assert.throws(
		() => fejkdata.render("x", { seed: 2 ** 53, today: DAY }),
		(error) => error instanceof FejkdataError && error.exitCode === 2 && /bigint/.test(error.message),
	);
});

test("today is the day builtins read, as a string or the UTC day of a Date", () => {
	const template = "{date(today,today,'2006-01-02')}";
	assert.equal(fejkdata.render(template, { seed: 1, today: DAY }), DAY);
	assert.equal(fejkdata.render(template, { seed: 1, today: new Date("2025-03-09T23:30:00Z") }), "2025-03-09");
	assert.equal(fejkdata.render(template, { seed: 1, today: new Date("2025-03-09T23:30:00-05:00") }), "2025-03-10");
});

test("a Date that holds no day is refused", () => {
	assert.throws(
		() => fejkdata.render("x", { seed: 1, today: new Date("nonsense") }),
		(error) => error instanceof FejkdataError && error.exitCode === 2 && /--today/.test(error.message),
	);
});

test("repeat defaults to one render and format to text", () => {
	assert.equal(fejkdata.render("{/misc.territory[SE].alpha2} ", { seed: 1, today: DAY }), "SE ");
	assert.equal(fejkdata.render("{/misc.territory[SE].alpha2} ", { seed: 1, today: DAY, repeat: 3 }), "SE SE SE ");
});

test("text outside ASCII comes back whole", () => {
	assert.equal(fejkdata.render("Malmö 🙂 {/misc.territory[SE].alpha2}", { seed: 1, today: DAY }), "Malmö 🙂 SE");
});

test("a template that does not compile throws the Go error text and exit code 2", () => {
	assert.throws(
		() => fejkdata.render("{/nope.x}", { seed: 1, today: DAY }),
		(error) => {
			assert.ok(error instanceof FejkdataError);
			assert.equal(error.name, "FejkdataError");
			assert.equal(error.exitCode, 2);
			assert.equal(error.message, 'fejkdata: template: reference {/nope.x}: no entry "nope"');
			return true;
		},
	);
});

test("an option the Go build refuses throws its error text", () => {
	assert.throws(
		() => fejkdata.render("x", { seed: -1, today: DAY }),
		(error) => error instanceof FejkdataError && error.exitCode === 2 && error.message === 'fejkdata: --seed needs an unsigned integer, got "-1"',
	);
});

test("a missing seed or day is refused, never filled in", () => {
	assert.throws(() => fejkdata.render("x", { today: DAY }), FejkdataError);
	assert.throws(() => fejkdata.render("x", { seed: 1 }), FejkdataError);
	assert.throws(() => fejkdata.render("x"), TypeError);
});

test("a template that is not a string is refused, naming JSON.stringify", () => {
	assert.throws(() => fejkdata.render({ format: "x" }, { seed: 1, today: DAY }), /JSON\.stringify/);
});

test("the typed formats are the ones the Go build names", () => {
	const typed = readFileSync(new URL("./index.d.ts", import.meta.url), "utf8").match(/format\?: (.*);/)[1];
	const named = goCLI("", "--help").match(/output form: (.*)\n/)[1];
	assert.deepEqual(typed.match(/\w+/g).sort(), named.replace("(default)", "").replace(" or ", " ").match(/\w+/g).sort());
});

// runBlocks runs each `js` block of markdown as a file of this folder, where the package's own name
// resolves. It fails unless the block prints the `text` block after it, and returns how many it ran.
function runBlocks(markdown) {
	const blocks = [...markdown.matchAll(/^```(\w+)\n(.*?)^```$/gms)].map(([, language, code]) => ({ language, code }));
	let ran = 0;
	blocks.forEach(({ language, code }, i) => {
		if (language !== "js") {
			return;
		}
		assert.equal(blocks[i + 1]?.language, "text", code);
		const file = new URL("./readme-block.mjs", import.meta.url);
		writeFileSync(file, code);
		const done = spawnSync(process.execPath, [fileURLToPath(file)], { encoding: "utf8" });
		rmSync(file);
		assert.equal(done.stderr, "", code);
		assert.equal(done.stdout, blocks[i + 1].code, code);
		ran++;
	});
	return ran;
}

test("the package's README prints what its text blocks show", () => {
	assert.equal(runBlocks(readFileSync(new URL("./README.md", import.meta.url), "utf8")), 3);
});

test("a js block with no text block after it fails", () => {
	assert.throws(() => runBlocks("```js\nconsole.log(1);\n```\n\nIt prints 1.\n"), assert.AssertionError);
});

test("a js block that prints something else fails", () => {
	assert.throws(() => runBlocks("```js\nconsole.log(1);\n```\n\n```text\n2\n```\n"), assert.AssertionError);
});
