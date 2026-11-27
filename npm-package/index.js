import "./wasm_exec.js";

export class FejkdataError extends Error {
	constructor(message, exitCode, kind) {
		super(message);
		this.name = "FejkdataError";
		this.exitCode = exitCode;
		this.kind = kind;
	}
}

// The last line of some CLI errors, which helps only at a shell.
const SHELL_ADVICE = /\n(try 'fejkdata --help'|wrap that JSON in single quotes.*)$/;

let starting;
let shipped;

export function load(options) {
	if (options === undefined) {
		shipped ??= open({}).catch((error) => {
			shipped = undefined;
			throw error;
		});
		return shipped;
	}
	return open(options);
}

export function isTemplate(text) {
	if (typeof text !== "string") {
		throw new TypeError("fejkdata: isTemplate reads a string");
	}
	if (text.includes("{")) {
		return true;
	}
	if (!/^\s*["[]/.test(text)) {
		return false;
	}
	try {
		JSON.parse(text);
		return true;
	} catch {
		return false;
	}
}

async function open(options) {
	if (typeof options !== "object" || options === null) {
		throw new TypeError('fejkdata: load takes an object, such as { data: ["sv_SE"] }');
	}
	const { data, files = {} } = options;
	if (data !== undefined && !(Array.isArray(data) && data.every((name) => typeof name === "string"))) {
		throw new TypeError('fejkdata: data is an array of module names, such as ["sv_SE"]');
	}
	if (typeof files !== "object" || files === null) {
		throw new TypeError('fejkdata: files is an object of file names and their text, such as { "mydata/team.json": "[]" }');
	}
	const names = data === undefined ? null : [...data];
	const pairs = Object.entries(files).map(([path, body]) => [path, bytes(path, body)]);
	const bridge = await started();
	const opened = (call) => (...args) => bridge[call](names, pairs, ...args);
	answer(opened("check")());
	return {
		render: (template, options) => render(opened("render"), template, options),
		fake: (path, options) => fake(opened("fake"), path, options),
		records: (template, options) => records(opened("records"), template, options),
		paths: () => answer(opened("paths")()).paths,
	};
}

function bytes(path, body) {
	if (typeof body === "string") {
		return new TextEncoder().encode(body);
	}
	if (body instanceof Uint8Array) {
		return new Uint8Array(body);
	}
	throw new TypeError(`fejkdata: files["${path}"] is a string or a Uint8Array`);
}

function started() {
	starting ??= start().catch((error) => {
		starting = undefined;
		throw error;
	});
	return starting;
}

async function start() {
	const go = new globalThis.Go();
	const instance = await WebAssembly.instantiate(await compiled(new URL("./fejkdata.wasm", import.meta.url)), go.importObject);
	return new Promise((resolve, reject) => {
		// The Go program can reach only globals, so it hands its functions over through this one.
		globalThis.__fejkdataRun = resolve;
		go.run(instance).then(() => reject(new Error("fejkdata: the WebAssembly program stopped while loading")), reject);
	}).finally(() => delete globalThis.__fejkdataRun);
}

// Node's fetch refuses a file: URL, so a runtime with Node's fs reads a file on disk. A browser fetches.
async function compiled(url) {
	const fs = url.protocol === "file:" && globalThis.process?.getBuiltinModule?.("node:fs/promises");
	return fs ? WebAssembly.compile(await fs.readFile(url)) : WebAssembly.compileStreaming(fetch(url));
}

function drawArguments(seed, today) {
	if (typeof seed === "number" && !Number.isSafeInteger(seed)) {
		throw new FejkdataError(`fejkdata: seed ${seed} is not a whole number from 0 to ${Number.MAX_SAFE_INTEGER}; write a seed above ${Number.MAX_SAFE_INTEGER} as a bigint, such as 9007199254740993n`, 2, "misuse");
	}
	const day = today instanceof Date && !Number.isNaN(today.getTime()) ? today.toISOString().slice(0, 10) : String(today);
	return [String(seed), day];
}

function render(run, template, { seed, today, repeat = 1, format = "text" }) {
	if (typeof template !== "string") {
		throw new TypeError("fejkdata: a template is a string; pass an object through JSON.stringify first");
	}
	const done = run(...drawArguments(seed, today), String(repeat), String(format), template);
	if (done.exitCode !== 0) {
		throw new FejkdataError(done.stderr.trim().replace(SHELL_ADVICE, "") + done.note, done.exitCode, done.kind);
	}
	return done.stdout;
}

function fake(run, path, { seed, today }) {
	if (typeof path !== "string") {
		throw new TypeError("fejkdata: a path is a string, such as \"sv_SE.address.postal-code\"");
	}
	return answer(run(...drawArguments(seed, today), path)).value;
}

const READ = { integer: BigInt, number: Number, boolean: (text) => text === "true", string: String };

function records(run, template, { seed, today, repeat = 1 }) {
	const text = typeof template === "object" && template !== null ? JSON.stringify(template) : template;
	if (typeof text !== "string") {
		throw new TypeError("fejkdata: a record template is JSON text or an object");
	}
	const { names, types, values } = answer(run(...drawArguments(seed, today), String(repeat), text));
	const reads = types.map((type) => READ[type]);
	const rows = [];
	for (let i = 0; i < values.length; i += names.length) {
		rows.push(Object.fromEntries(names.map((name, j) => [name, values[i + j] === null ? null : reads[j](values[i + j])])));
	}
	return rows;
}

function answer(done) {
	if (done.error) {
		throw new FejkdataError(done.error, done.exitCode, done.kind);
	}
	return done;
}
