# Release checklist

What to settle before the first tag, then the work that follows in a later,
data-heavy release.

## Before the first release — settle the record flag and API contract

- `--format` vocabulary — confirm `text`, `json`, `ndjson`, `csv`, `sql`; the
  `json`-as-array vs `ndjson`-as-lines split; `--table` (the SQL INSERT target);
  the `--separator` rejection on record formats; and the exit codes (misuse 2,
  runtime 1).
- Library surface — confirm `Record`, `FakeRecord`, `NewRecordTemplate`,
  `RecordTemplate`, `Column`/`Columns()`, and the `JSON()`, `CSVHeader()`,
  `CSVLine()`, `SQLInsert()` serializers.
- Typed scalars — columns are strings today (`"42"`, quoted SQL). Confirm that
  stays out of scope, or add a per-column `kind` before the tag.
- Struct-filling — `fake:"..."` tags (reflection over an arbitrary struct) stay
  out of scope; `Columns()` hands the caller the values to map themselves.
  Confirm.
- Independent reference draw — within one record every tailed reference to a
  category is one draw, with no spelling for "these columns should disagree".
  Confirm the per-record contract, or add the spelling.

## Later, in a data-heavy release

- Shipped-data de-duplication — `email.json`'s `local` is a drifted copy of
  `username.json`; fold it in when the shipped set grows and we add lots more
  data.
