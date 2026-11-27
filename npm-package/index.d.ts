import type { ShippedPath } from "./shipped-paths.js";
export type { ShippedPath };

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

/** The options of fake: a seed and a day, as render takes them. */
export interface FakeOptions {
	/** A whole number from 0 to 2^64 - 1. A number above 2^53 - 1 is written as a bigint. */
	seed: number | bigint;
	/** The day ages and dates count from: a string such as "2026-01-15", or a Date, whose UTC day is read. */
	today: string | Date;
}

/** The options of records: a seed, a day and how many rows to draw. */
export interface RecordsOptions {
	/** A whole number from 0 to 2^64 - 1. A number above 2^53 - 1 is written as a bigint. */
	seed: number | bigint;
	/** The day ages and dates count from: a string such as "2026-01-15", or a Date, whose UTC day is read. */
	today: string | Date;
	/** How many rows to draw. It defaults to 1. */
	repeat?: number;
}

/** A column's value in a row: a bigint for an integer column, a number, a boolean, null, or else a string. */
export type ColumnValue = string | number | bigint | boolean | null;

/** What load reads. */
export interface LoadOptions {
	/** The shipped modules to load, such as ["sv_SE", "misc"], each with the modules it reads. Every module loads without it. */
	data?: readonly string[];
	/** Your own data files, by path, such as "mydata/users.json", laid over the shipped data. */
	files?: Readonly<Record<string, string | Uint8Array>>;
}

type OptionKey = "datatype" | "format" | "hidden" | "let" | "repeat" | "separator" | "weight";

type HiddenKey<T> = T extends { hidden: readonly (infer H)[] } ? H : never;

type LoneReference<S extends string> = S extends `{${infer Inner}}`
	? Inner extends `${string}${"{" | "}" | " " | "|" | "(" | ","}${string}` ? false : true
	: false;

type TextColumn<S> = S extends string ? string extends S ? ColumnValue : LoneReference<S> extends true ? ColumnValue : string : ColumnValue;

type ByDatatype = { integer: bigint; number: number; boolean: boolean; string: string };

type ColumnOf<F> = F extends null ? null
	: F extends readonly (infer Item)[] ? ColumnOf<Item>
	: F extends { datatype: infer D extends keyof ByDatatype } ? ByDatatype[D]
	: F extends { format: infer S } ? TextColumn<S>
	: F extends string ? TextColumn<F>
	: F extends number | boolean ? string
	: ColumnValue;

/** A row of records drawn from the object template T: a key per field that writes a column. */
export type RecordRow<T> = { -readonly [K in Exclude<keyof T, OptionKey | HiddenKey<T>> & string]: ColumnOf<T[K]> };

export interface Fejkdata {
	/**
	 * Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.
	 * Throws a FejkdataError when fejkdata refuses the template or an argument.
	 */
	render(template: string, options: RenderOptions): string;
	/** Draws one value of a path, such as "sv_SE.address.postal-code": a render of `{/path}`. */
	fake(path: ShippedPath | (string & {}), options: FakeOptions): string;
	/**
	 * Draws rows from a record template, given as JSON text or an object. A row is an object with a key per column, in the template's order.
	 * An integer column is a bigint, a number column a number, a boolean column a boolean, and any other a string. A null is null.
	 */
	records(template: string, options: RecordsOptions): Record<string, ColumnValue>[];
	records<const T extends { readonly [key: string]: unknown }>(template: T, options: RecordsOptions): RecordRow<T>[];
	/** Every path the loaded data holds, as `fejkdata --list` prints them. */
	paths(): string[];
}

/** Where a FejkdataError arose: a mistake in a template or an argument, the data failing to load, a limit passed, or anything else. */
export type FejkdataErrorKind = "misuse" | "load" | "limit" | "other";

/** Thrown when fejkdata refuses a template, an argument or the data. The message says what is wrong, in the CLI's words. */
export class FejkdataError extends Error {
	/** The CLI's exit code: 2 for a mistake in the template, the arguments or load's options, 1 for anything else. */
	readonly exitCode: number;
	/** Where the error arose. */
	readonly kind: FejkdataErrorKind;
	constructor(message: string, exitCode: number, kind: FejkdataErrorKind);
}

/** Loads the WebAssembly, compiled once however often it is called. With no options it gives the same object each time. */
export function load(options?: LoadOptions): Promise<Fejkdata>;

/** Whether text is a template rather than a path, by the rule fejkdata's CLI reads an argument by. */
export function isTemplate(text: string): boolean;
