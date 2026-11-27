export interface RenderOptions {
	/** A whole number from 0 to 2^64 - 1. A number above 2^53 - 1 is written as a bigint. The same seed and day give the same text. */
	seed: number | bigint;
	/** The day ages and dates count from: a string such as "2026-01-15", or a Date, whose UTC day is read. */
	today: string | Date;
	/** How many times to render the template. It defaults to 1. */
	repeat?: number;
	/** "text" unless you ask for records. */
	format?: "text" | "json" | "ndjson" | "csv" | "tsv" | "pgcopy" | "sql";
}

export interface Fejkdata {
	/**
	 * Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.
	 * A render fejkdata refuses throws a FejkdataError.
	 */
	render(template: string, options: RenderOptions): string;
}

/** A render fejkdata refused. The message is the CLI's error text. */
export class FejkdataError extends Error {
	/** The CLI's exit code: 2 for a mistake in the template or the options, 1 for a runtime error. */
	readonly exitCode: number;
	constructor(message: string, exitCode: number);
}

/** Loads the WebAssembly, once however often it is called. */
export function load(): Promise<Fejkdata>;
