import * as everything from "@larvit/fejkdata";
import { FejkdataError, isTemplate, load } from "@larvit/fejkdata";
import type { ColumnValue, FakeOptions, Fejkdata, FejkdataErrorKind, LoadOptions, RecordsOptions, RenderOptions, ShippedPath } from "@larvit/fejkdata";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends <T>() => T extends B ? 1 : 2 ? true : false;

function expectSame<_ extends true>() {}

expectSame<Same<keyof typeof everything, "FejkdataError" | "isTemplate" | "load">>();
expectSame<Same<typeof load, (options?: LoadOptions) => Promise<Fejkdata>>>();
expectSame<Same<typeof isTemplate, (text: string) => boolean>>();
expectSame<Same<LoadOptions, {
	data?: readonly string[];
	files?: Readonly<Record<string, string | Uint8Array>>;
}>>();
expectSame<Same<keyof Fejkdata, "render" | "fake" | "records" | "paths">>();
expectSame<Same<Fejkdata["render"], (template: string, options: RenderOptions) => string>>();
expectSame<Same<Fejkdata["fake"], (path: ShippedPath | (string & {}), options: FakeOptions) => string>>();
expectSame<Same<Fejkdata["paths"], () => string[]>>();
expectSame<Same<RenderOptions, {
	seed: number | bigint;
	today: string | Date;
	repeat?: number;
	format?: "text" | "csv" | "json" | "ndjson" | "pgcopy" | "sql" | "tsv";
}>>();
expectSame<Same<FakeOptions, { seed: number | bigint; today: string | Date }>>();
expectSame<Same<RecordsOptions, { seed: number | bigint; today: string | Date; repeat?: number }>>();
expectSame<Same<ColumnValue, string | number | bigint | boolean | null>>();
expectSame<Same<FejkdataErrorKind, "misuse" | "load" | "limit" | "other">>();
expectSame<Same<ConstructorParameters<typeof FejkdataError>, [message: string, exitCode: number, kind: FejkdataErrorKind]>>();
expectSame<Same<keyof FejkdataError, keyof Error | "exitCode" | "kind">>();
expectSame<Same<FejkdataError["exitCode"], number>>();
expectSame<Same<FejkdataError["kind"], FejkdataErrorKind>>();

const fejkdata = await load();
await load({ data: ["sv_SE", "misc"], files: { "mydata/users.json": "{}", "mydata/teams.json": new Uint8Array() } });
const day = { seed: 1, today: "2026-01-15" };
fejkdata.render("{/sv_SE.first-name}", day);
fejkdata.render("x", { seed: 2n ** 64n - 1n, today: new Date(0), repeat: 3, format: "text" });
fejkdata.fake("sv_SE.address.postal-code", day);
fejkdata.fake("mydata.users", { seed: 1n, today: new Date(0) });

const rows = fejkdata.records({
	format: "{/sv_SE.person as p}",
	hidden: { last: "{p.last}" },
	greeting: "Hej {p.first}",
	first: "{p.first}",
	score: "{/src.score}",
	person: "{p}",
	answer: 42,
	id: { format: "{seq()}", datatype: "integer" },
	share: { format: "{float(0,1,2)}", datatype: "number" },
	paid: { format: "{p}", p: ["true", "false"], datatype: "boolean" },
	color: { format: "{/sv_SE.color}", datatype: "string" },
	read: { format: "{/src.score}" },
	nothing: null,
	maybe: ["x", null],
	maybeId: [{ format: "{seq()}", datatype: "integer" }, null],
}, { ...day, repeat: 2 });
expectSame<Same<typeof rows, {
	greeting: string;
	first: ColumnValue;
	score: ColumnValue;
	person: ColumnValue;
	answer: string;
	id: bigint;
	share: number;
	paid: boolean;
	color: string;
	read: ColumnValue;
	nothing: null;
	maybe: string | null;
	maybeId: bigint | null;
}[]>>();

const template = {
	format: "{/sv_SE.person as p}",
	hidden: { last: "{p.last}" },
	id: { format: "{seq()}", datatype: "integer" },
	name: "{p.first} {last}",
	member: { format: "{m}", m: ["true", "false"], datatype: "boolean" },
};
const heldRows = fejkdata.records(template, { ...day, repeat: 2 });
expectSame<Same<typeof heldRows, { id: ColumnValue; name: ColumnValue; member: ColumnValue }[]>>();

const fixedTemplate = {
	format: "{/sv_SE.person as p}",
	hidden: { last: "{p.last}" },
	id: { format: "{seq()}", datatype: "integer" },
	name: "{p.first} {last}",
	member: { format: "{m}", m: ["true", "false"], datatype: "boolean" },
} as const;
const fixedRows = fejkdata.records(fixedTemplate, day);
expectSame<Same<typeof fixedRows, { id: bigint; name: string; member: boolean }[]>>();

const textRows = fejkdata.records('{"format": "", "id": "{uuid()}"}', day);
expectSame<Same<typeof textRows, Record<string, ColumnValue>[]>>();

const shipped: ShippedPath = "sv_SE.first-name";
void shipped;
// @ts-expect-error: no shipped path is spelled so
const unshipped: ShippedPath = "sv_SE.no-such-path";
void unshipped;

// @ts-expect-error: a path is a string
fejkdata.fake(5, day);
// @ts-expect-error: records takes no format
fejkdata.records('{"format": "", "id": "{uuid()}"}', { ...day, format: "csv" });
// @ts-expect-error: fake takes no repeat
fejkdata.fake("misc.uuid", { ...day, repeat: 2 });
// @ts-expect-error: data is an array of module names
await load({ data: "sv_SE" });
// @ts-expect-error: a file is text or bytes
await load({ files: { "mydata/users.json": 5 } });
// @ts-expect-error: isTemplate reads a string
isTemplate(5);

const error = new FejkdataError("refused", 2, "misuse");
const anError: Error = error;
void anError;
// @ts-expect-error: exitCode is read-only
error.exitCode = 1;
// @ts-expect-error: kind is read-only
error.kind = "load";
// @ts-expect-error: no kind is spelled so
new FejkdataError("refused", 2, "template");
