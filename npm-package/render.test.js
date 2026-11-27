import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync, rmSync, writeFileSync } from "node:fs";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { FejkdataError, isTemplate, load } from "./index.js";

const DAY = "2026-01-15";
const PERSON = '{"format": "{/sv_SE.person as p}", "first": "{p.first}", "last": "{p.last}", "born": "{born(18,90,\'2006-01-02\')}"}';
const fejkdata = await load();
const OPTIONS = { seed: 1, today: DAY };
const LISTED = goCLI("", "--list").split("\n").filter(Boolean);
const MODULES = [...new Set(LISTED.map((path) => path.split(".").slice(0, path.startsWith("geo.") ? 2 : 1).join(".")))];

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
			(error) => error instanceof FejkdataError && error.exitCode === 1 && error.kind === "limit" && /512 MB/.test(error.message),
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

for (const template of ["{/sv_SE.personnummer-12}", "{/sv_SE.color}"]) {
	test(`${template} under a format is misuse, though the CLI exits 1`, () => {
		assert.throws(() => fejkdata.render(template, { seed: 1, today: DAY, format: "json" }), errorOf("misuse", 1));
	});
}

test("a seed is a number or a bigint, up to 2^64 - 1", () => {
	const top = 18446744073709551615n;
	assert.equal(fejkdata.render("{uuid()}", { seed: top, today: DAY }), goCLI("{uuid()}", "--seed", String(top), "--today", DAY));
	assert.equal(fejkdata.render("{uuid()}", { seed: 12n, today: DAY }), fejkdata.render("{uuid()}", { seed: 12, today: DAY }));
});

test("a number too large to be exact is refused, naming bigint", () => {
	assert.throws(
		() => fejkdata.render("x", { seed: 2 ** 53, today: DAY }),
		(error) => error instanceof FejkdataError && error.exitCode === 2 && error.kind === "misuse" && /bigint/.test(error.message),
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
			assert.equal(error.kind, "misuse");
			assert.equal(error.message, 'fejkdata: template: reference {/nope.x}: no entry "nope"');
			return true;
		},
	);
});

test("an option the Go build refuses throws its error text", () => {
	assert.throws(
		() => fejkdata.render("x", { seed: -1, today: DAY }),
		(error) => error instanceof FejkdataError && error.exitCode === 2 && error.kind === "misuse" && error.message === 'fejkdata: --seed needs an unsigned integer, got "-1"',
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

function errorOf(kind, exitCode) {
	return (error) => {
		assert.ok(error instanceof FejkdataError, error);
		assert.equal(error.kind, kind, error.message);
		assert.equal(error.exitCode, exitCode, error.message);
		return true;
	};
}

test("paths lists what the Go build lists", () => {
	assert.deepEqual(fejkdata.paths(), LISTED);
});

test("load with no argument gives the same object each time, and load with options a new one", async () => {
	assert.equal(await load(), fejkdata);
	assert.notEqual(await load({}), fejkdata);
});

test("every load shares one compile of the WebAssembly", async (t) => {
	const fresh = await import("./index.js?one-compile");
	const compile = t.mock.method(WebAssembly, "compile");
	const compileStreaming = t.mock.method(WebAssembly, "compileStreaming");
	await Promise.all([fresh.load(), fresh.load({ data: ["misc"] }), fresh.load({ files: {} })]);
	await fresh.load({ data: ["sv_SE"] });
	assert.equal(compile.mock.callCount() + compileStreaming.mock.callCount(), 1);
});

test("naming every shipped module loads what load() does", async () => {
	assert.deepEqual((await load({ data: MODULES })).paths(), LISTED);
});

test("naming a module loads the modules it reads and leaves out the rest", async () => {
	const swedish = await load({ data: ["sv_SE"] });
	const paths = swedish.paths();
	assert.ok(paths.includes("sv_SE.address.postal-code"));
	assert.ok(paths.includes("geo.SE.address"));
	assert.ok(!paths.some((path) => path.startsWith("en_US.") || path.startsWith("geo.US.")));
	assert.equal(swedish.render("{/sv_SE.address}", OPTIONS), fejkdata.render("{/sv_SE.address}", OPTIONS));
});

test("a module name the package doesn't ship is refused, naming every shipped one", async () => {
	await assert.rejects(load({ data: ["sv_SE", "xx_XX"] }), (error) => {
		errorOf("misuse", 2)(error);
		assert.match(error.message, /"xx_XX"/);
		for (const name of MODULES) {
			assert.ok(error.message.includes(name), `${error.message} names ${name}`);
		}
		return true;
	});
	await assert.rejects(load({ data: "sv_SE" }), TypeError);
	await assert.rejects(load({ data: [5] }), TypeError);
});

const swedish = await load({ data: ["sv_SE"] });
const leftOut = {
	render: () => swedish.render("{/en_US.person}", OPTIONS),
	fake: () => swedish.fake("en_US.person", OPTIONS),
	records: () => swedish.records({ format: "", person: "{/en_US.person}" }, OPTIONS),
};
for (const [call, run] of Object.entries(leftOut)) {
	test(`${call} reading a module that data leaves out is a load error, naming load's data`, () => {
		assert.throws(run, (error) => errorOf("load", 2)(error) && /load's data/.test(error.message));
	});
}

test("a reference to no module is misuse, though data leaves modules out", () => {
	assert.throws(() => swedish.render("{/xx_XX.person}", OPTIONS), errorOf("misuse", 2));
});

test("a bad argument is misuse, though the template reads a module data leaves out", () => {
	assert.throws(() => swedish.render("{/en_US.person}", { seed: -1, today: DAY }), errorOf("misuse", 2));
});

test("files layer over the shipped data, as text or bytes", async () => {
	const mine = await load({
		data: ["sv_SE"],
		files: {
			"mydata/team.json": '["Red"]',
			"mydata/user.json": new TextEncoder().encode('{"format": "{/sv_SE.first-name} of {/mydata.team}"}'),
			"sv_SE/color.json": '["Ultraviolett"]',
		},
	});
	assert.ok(mine.paths().includes("mydata.user"));
	assert.match(mine.fake("mydata.user", OPTIONS), /^\S+ of Red$/);
	assert.equal(mine.fake("sv_SE.color", OPTIONS), "Ultraviolett");
	assert.equal(fejkdata.paths().includes("mydata.user"), false);
});

test("a file that is no text or bytes, or whose path is none, is refused at load", async () => {
	await assert.rejects(load({ files: { "mydata/team.json": 5 } }), TypeError);
	await assert.rejects(load({ files: { "../team.json": "[]" } }), errorOf("misuse", 2));
});

test("a file that fails to load is a load error in each call", async () => {
	const broken = await load({ files: { "mydata/team.json": "[" } });
	assert.throws(() => broken.render("x", OPTIONS), errorOf("load", 1));
	assert.throws(() => broken.fake("misc.uuid", OPTIONS), errorOf("load", 1));
	assert.throws(() => broken.records({ format: "", id: "{uuid()}" }, OPTIONS), errorOf("load", 1));
	assert.throws(() => broken.paths(), errorOf("load", 1));
});

const FAKED = {
	"sv_SE.address.postal-code": /^\d{3} \d{2}$/,
	"en_US.person": /^\S.* \S+$/,
	"misc.territory[SE].alpha2": /^SE$/,
};
for (const [path, shape] of Object.entries(FAKED)) {
	test(`fake draws a value of ${path}, the same for the same seed and day`, () => {
		assert.match(fejkdata.fake(path, OPTIONS), shape);
		assert.equal(fejkdata.fake(path, OPTIONS), fejkdata.fake(path, OPTIONS));
	});
}

test("fake reads a path with a leading slash", () => {
	assert.equal(fejkdata.fake("/sv_SE.person", OPTIONS), fejkdata.fake("sv_SE.person", OPTIONS));
});

for (const text of ["{/sv_SE.person}", "sv_SE.person} {/en_US.person", "sv_SE.nope", "sv_SE.first-name|/en_US.first-name", "sv_SE.person as p", "sv_SE.first-name where length 3"]) {
	test(`fake refuses ${JSON.stringify(text)}, which is no path to a value`, () => {
		assert.throws(() => fejkdata.fake(text, OPTIONS), errorOf("misuse", 2));
	});
}

test("fake refuses a path that is no string", () => {
	assert.throws(() => fejkdata.fake(5, OPTIONS), TypeError);
});

const TYPED = {
	format: "",
	zeta: "{/sv_SE.first-name}",
	id: { format: "{seq()}", datatype: "integer" },
	share: { format: "{float(0,1,2)}", datatype: "number" },
	paid: { format: "{p}", p: ["true", "false"], datatype: "boolean" },
	nothing: null,
	maybe: ["x", null],
	color: { format: "{/sv_SE.color}", datatype: "string" },
};

test("records are the rows render writes as ndjson, each column read by its datatype", () => {
	const options = { seed: 4, today: DAY, repeat: 30 };
	const rows = fejkdata.records(TYPED, options);
	const written = fejkdata.render(JSON.stringify(TYPED), { ...options, format: "ndjson" }).trimEnd().split("\n").map((line) => JSON.parse(line));
	assert.equal(rows.length, 30);
	assert.deepEqual(rows.map((row) => ({ ...row, id: Number(row.id) })), written);
	for (const row of rows) {
		assert.deepEqual(Object.keys(row), ["zeta", "id", "share", "paid", "nothing", "maybe", "color"]);
		assert.equal(typeof row.id, "bigint");
		assert.equal(typeof row.share, "number");
		assert.equal(typeof row.paid, "boolean");
		assert.equal(row.nothing, null);
	}
	assert.ok(rows.some((row) => row.maybe === null) && rows.some((row) => row.maybe === "x"));
});

test("an integer column is a bigint, exact beyond 2^53", () => {
	const [row] = fejkdata.records({ format: "", big: { format: "9007199254740993", datatype: "integer" } }, OPTIONS);
	assert.equal(row.big, 9007199254740993n);
});

test("records takes JSON text as it takes an object, and a lone reference as its record", () => {
	const options = { seed: 2, today: DAY, repeat: 3 };
	assert.deepEqual(fejkdata.records(JSON.stringify(TYPED), options), fejkdata.records(TYPED, options));
	const written = fejkdata.render("{/sv_SE.address}", { ...options, format: "ndjson" }).trimEnd().split("\n").map((line) => JSON.parse(line));
	assert.deepEqual(fejkdata.records("{/sv_SE.address}", options), written);
	assert.deepEqual(fejkdata.records('"{/sv_SE.address}"\n', options), written);
});

test("records gives one row unless repeat says otherwise", () => {
	assert.equal(fejkdata.records(TYPED, OPTIONS).length, 1);
	assert.deepEqual(fejkdata.records(TYPED, { ...OPTIONS, repeat: 0 }), []);
});

for (const template of [{ format: "x" }, '{"format": "{/nope.x}", "a": "b"}', "{/sv_SE.color}"]) {
	test(`records refuses ${JSON.stringify(template)} as misuse`, () => {
		assert.throws(() => fejkdata.records(template, OPTIONS), errorOf("misuse", 2));
	});
}

test("records refuses a template that fails on its first draw, though repeat is 0", () => {
	assert.throws(() => fejkdata.records("{/sv_SE.color}", { ...OPTIONS, repeat: 0 }), errorOf("misuse", 2));
});

for (const options of [{ seed: -1, today: DAY }, { seed: 1, today: "15/01/2026" }, { seed: 1, today: DAY, repeat: 1e7 }, { seed: 1, today: DAY, repeat: -1 }, { seed: 2 ** 53, today: DAY }, { today: DAY }]) {
	test(`records refuses the options ${JSON.stringify(options)} as misuse`, () => {
		assert.throws(() => fejkdata.records(TYPED, options), errorOf("misuse", 2));
	});
}

test("records refuses a template that is no text or object", () => {
	assert.throws(() => fejkdata.records(5, OPTIONS), TypeError);
});

test("records tells whitespace around a lone reference, as render does", () => {
	const written = () => fejkdata.render(" {/sv_SE.address} ", { ...OPTIONS, format: "ndjson" });
	const drawn = () => fejkdata.records(" {/sv_SE.address} ", OPTIONS);
	assert.throws(written, /the whitespace around \{\/sv_SE\.address\} makes it text/);
	assert.throws(drawn, (error) => errorOf("misuse", 2)(error) && /the whitespace around \{\/sv_SE\.address\} makes it text/.test(error.message));
});

test("records keeps a column named __proto__", () => {
	const [row] = fejkdata.records('{"format": "", "__proto__": "x", "a": "y"}', OPTIONS);
	assert.deepEqual(Object.entries(row), [["__proto__", "x"], ["a", "y"]]);
});

test("records past 512 MB are refused as a limit, and the next call works", () => {
	assert.throws(() => fejkdata.records({ format: "", wide: "x".repeat(70000) }, { ...OPTIONS, repeat: 10000 }), errorOf("limit", 1));
	assert.equal(fejkdata.records(TYPED, OPTIONS).length, 1);
});

const IS_TEMPLATE = JSON.parse(readFileSync(new URL("../testdata/is_template.json", import.meta.url), "utf8"));

for (const arg of IS_TEMPLATE.template) {
	test(`isTemplate takes ${JSON.stringify(arg)} as a template, as Go's IsTemplate does`, () => {
		assert.equal(isTemplate(arg), true);
	});
}

for (const arg of [...IS_TEMPLATE.path, ...Object.keys(IS_TEMPLATE.malformed)]) {
	test(`isTemplate takes ${JSON.stringify(arg)} as no template, as Go's IsTemplate does`, () => {
		assert.equal(isTemplate(arg), false);
	});
}

test("isTemplate refuses what is no string", () => {
	assert.throws(() => isTemplate(5), TypeError);
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
	assert.equal(runBlocks(readFileSync(new URL("./README.md", import.meta.url), "utf8")), 6);
});

test("the repository's README prints what its text blocks show", () => {
	assert.equal(runBlocks(readFileSync(new URL("../README.md", import.meta.url), "utf8")), 1);
});

test("a js block with no text block after it fails", () => {
	assert.throws(() => runBlocks("```js\nconsole.log(1);\n```\n\nIt prints 1.\n"), assert.AssertionError);
});

test("a js block that prints something else fails", () => {
	assert.throws(() => runBlocks("```js\nconsole.log(1);\n```\n\n```text\n2\n```\n"), assert.AssertionError);
});
