# Rules

- Go runs only through `docker compose run --rm <test|vet|fmt|build>`; the merge gate is `docker build .`.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output is its own commit.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- One spelling per result: reject the other at `New`, and let the error name the spelling to use.
- A standing choice a reader would relitigate goes under Decisions in the README, not in a comment.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
