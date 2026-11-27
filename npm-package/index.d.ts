export interface RenderOptions {
	/** A whole number from 0 to 2^64 - 1. A number above 2^53 - 1 is written as a bigint. The same seed and day give the same text. */
	seed: number | bigint;
	/** The day ages and dates count from: a string such as "2026-01-15", or a Date, whose UTC day is read. */
	today: string | Date;
	/** How many times to render the template. It defaults to 1. */
	repeat?: number;
	/** "text" unless you ask for records: one per render of a JSON template, with a column for each field. */
	format?: "text" | "csv" | "json" | "ndjson" | "pgcopy" | "sql" | "tsv";
}

export interface Fejkdata {
	/**
	 * Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.
	 * Throws a FejkdataError when fejkdata refuses the template or an argument.
	 */
	render(template: string, options: RenderOptions): string;
}

/** Thrown when fejkdata refuses a template or an argument. The message says what is wrong, in the CLI's words. */
export class FejkdataError extends Error {
	/** The CLI's exit code: 2 for a mistake in the template or the arguments, 1 for anything else. */
	readonly exitCode: number;
	constructor(message: string, exitCode: number);
}

/** Loads the WebAssembly, once however often it is called. */
export function load(): Promise<Fejkdata>;
