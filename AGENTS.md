# Rules

- Go runs only through `docker compose run --rm <test|vet|fmt|build|cyclo>`; the merge gate is `docker build .` plus CI's changelog check.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output or of `testdata/shipped_shape.txt` is its own commit.
- A change under `data/` or to the shape pin adds its `CHANGELOG.md` entry under `Unreleased` in the same PR, and so does a change to a flag, an exit code, an exported name, a fence, a builtin or the lowest Go; what is major is the README's Versioning table.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- One spelling per result: reject the other at `New`, and let the error name the spelling to use.
- A standing choice a reader would relitigate goes under Decisions in the README, not in a comment.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
- Cyclomatic complexity is gated at 14: the table-shaped dispatches (`linkParent`, `compileTemplate`, `renderEdges`) sit at it and stay whole; a function that would pass it is decomposed.
