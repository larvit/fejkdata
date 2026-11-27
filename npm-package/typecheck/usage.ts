import { FejkdataError, load } from "@larvit/fejkdata";
import type { Fejkdata, RenderOptions } from "@larvit/fejkdata";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends <T>() => T extends B ? 1 : 2 ? true : false;

function expectType<Expected>() {
	return <Actual>(_value: Actual, _same: Same<Actual, Expected>) => {};
}

const loading = load();
expectType<Promise<Fejkdata>>()(loading, true);

const fejkdata = await loading;
expectType<string>()(fejkdata.render("{/sv_SE.first-name}", { seed: 1, today: "2026-01-15" }), true);
fejkdata.render("x", { seed: 2n ** 64n - 1n, today: new Date(0), repeat: 3, format: "text" });

const everyFormat: Same<NonNullable<RenderOptions["format"]>, "text" | "csv" | "json" | "ndjson" | "pgcopy" | "sql" | "tsv"> = true;
const everyOption: Same<keyof RenderOptions, "seed" | "today" | "repeat" | "format"> = true;
const everyMethod: Same<keyof Fejkdata, "render"> = true;
void [everyFormat, everyOption, everyMethod];

// @ts-expect-error: render requires its options
fejkdata.render("x");
// @ts-expect-error: seed is required
fejkdata.render("x", { today: "2026-01-15" });
// @ts-expect-error: today is required
fejkdata.render("x", { seed: 1 });
// @ts-expect-error: a seed is a number or a bigint
fejkdata.render("x", { seed: "1", today: "2026-01-15" });
// @ts-expect-error: today is a string or a Date
fejkdata.render("x", { seed: 1, today: 0 });
// @ts-expect-error: repeat is a number
fejkdata.render("x", { seed: 1, today: "2026-01-15", repeat: "3" });
// @ts-expect-error: xml is no format
fejkdata.render("x", { seed: 1, today: "2026-01-15", format: "xml" });
// @ts-expect-error: an unknown option is refused
fejkdata.render("x", { seed: 1, today: "2026-01-15", locale: "sv_SE" });

const error = new FejkdataError("refused", 2);
expectType<FejkdataError>()(error, true);
expectType<number>()(error.exitCode, true);
expectType<string>()(error.message, true);
const anError: Error = error;
void anError;
// @ts-expect-error: exitCode is read-only
error.exitCode = 1;
// @ts-expect-error: a FejkdataError carries an exit code
new FejkdataError("refused");
