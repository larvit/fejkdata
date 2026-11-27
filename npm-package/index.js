import "./wasm_exec.js";

export class FejkdataError extends Error {
	constructor(message, exitCode) {
		super(message);
		this.name = "FejkdataError";
		this.exitCode = exitCode;
	}
}

// The last line of some CLI errors, which helps only at a shell.
const SHELL_ADVICE = /\n(try 'fejkdata --help'|wrap that JSON in single quotes.*)$/;

let loading;

export function load() {
	loading ??= start().catch((error) => {
		loading = undefined;
		throw error;
	});
	return loading;
}

async function start() {
	const go = new globalThis.Go();
	const instance = await WebAssembly.instantiate(await compiled(new URL("./fejkdata.wasm", import.meta.url)), go.importObject);
	const run = await new Promise((resolve, reject) => {
		// The Go program can reach only globals, so it hands its one function over through this one.
		globalThis.__fejkdataRun = resolve;
		go.run(instance).then(() => reject(new Error("fejkdata: the WebAssembly program stopped while loading")), reject);
	}).finally(() => delete globalThis.__fejkdataRun);
	return { render: (template, options) => render(run, template, options) };
}

// Node's fetch refuses a file: URL, so a runtime with Node's fs reads a file on disk. A browser fetches.
async function compiled(url) {
	const fs = url.protocol === "file:" && globalThis.process?.getBuiltinModule?.("node:fs/promises");
	return fs ? WebAssembly.compile(await fs.readFile(url)) : WebAssembly.compileStreaming(fetch(url));
}

function render(run, template, { seed, today, repeat = 1, format = "text" }) {
	if (typeof template !== "string") {
		throw new TypeError("fejkdata: a template is a string; pass an object through JSON.stringify first");
	}
	if (typeof seed === "number" && !Number.isSafeInteger(seed)) {
		throw new FejkdataError(`fejkdata: seed ${seed} is not a whole number from 0 to ${Number.MAX_SAFE_INTEGER}; write a seed above ${Number.MAX_SAFE_INTEGER} as a bigint, such as 9007199254740993n`, 2);
	}
	const day = today instanceof Date && !Number.isNaN(today.getTime()) ? today.toISOString().slice(0, 10) : String(today);
	const done = run(["--seed", String(seed), "--today", day, "--repeat", String(repeat), "--format", String(format)], template);
	if (done.exitCode !== 0) {
		throw new FejkdataError(done.stderr.trim().replace(SHELL_ADVICE, ""), done.exitCode);
	}
	return done.stdout;
}
