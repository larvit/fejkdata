import * as everything from "@larvit/fejkdata";
import { FejkdataError, load } from "@larvit/fejkdata";
import type { Fejkdata, RenderOptions } from "@larvit/fejkdata";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends <T>() => T extends B ? 1 : 2 ? true : false;

function expectSame<_ extends true>() {}

expectSame<Same<keyof typeof everything, "FejkdataError" | "load">>();
expectSame<Same<typeof load, () => Promise<Fejkdata>>>();
expectSame<Same<Fejkdata, { render(template: string, options: RenderOptions): string }>>();
expectSame<Same<RenderOptions, {
	seed: number | bigint;
	today: string | Date;
	repeat?: number;
	format?: "text" | "csv" | "json" | "ndjson" | "pgcopy" | "sql" | "tsv";
}>>();
expectSame<Same<ConstructorParameters<typeof FejkdataError>, [message: string, exitCode: number]>>();
expectSame<Same<keyof FejkdataError, keyof Error | "exitCode">>();
expectSame<Same<FejkdataError["exitCode"], number>>();

const fejkdata = await load();
fejkdata.render("{/sv_SE.first-name}", { seed: 1, today: "2026-01-15" });
fejkdata.render("x", { seed: 2n ** 64n - 1n, today: new Date(0), repeat: 3, format: "text" });

const error = new FejkdataError("refused", 2);
const anError: Error = error;
void anError;
// @ts-expect-error: exitCode is read-only
error.exitCode = 1;
