# Rules

- Go runs only through `docker compose run --rm <test|vet|fmt|build>`; the merge gate is `docker build .`.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output is its own commit.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- One spelling per result: reject the other at `New`, and let the error name the spelling to use.
- A standing choice a reader would relitigate goes under Decisions in the README, not in a comment.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
- Cyclomatic complexity is gated at 14: the table-shaped dispatches (`eachToken`, `calc.factor`, `walkPath`) sit at 13–14 and stay whole; anything else that reaches 14 is decomposed.

# Deferred

- **A record cannot ask for an independent reference draw (2026-09-04).** Within
  one record every tailed reference to a category is one draw, and a bare
  `{/cat}` renders the whole category, so a column wanting its own draw of
  `{/cat.field}` has no spelling for it. Left until a use case names which columns
  should disagree; premise: a record is one coherent row, which is what columns
  are for. Not raised in review before that.

- **Shipped-data de-duplication (2026-09-02).** `email.json`'s `local` is a
  drifted copy of `username.json`, and no shipped file yet uses a held path, an
  operand or a reference. Fixed in the data fill before the first tag, when the
  shipped set is rewritten anyway; premise: nothing depends on the shipped data's
  shape until then. Not raised in review before that.
