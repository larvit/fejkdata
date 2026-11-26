# Rules

- Go runs only through `docker compose run --rm <service>`, with the services `compose.yaml` names; the merge gate is what the README's Development section lists.
- Tests first, in their own commit; the implementation follows in the next. A re-pin of seeded output or of `testdata/shipped_shape.txt` is its own commit.
- A change to the shipped data under `data/`, its tests aside, to the shape pin, or to the pin of the listing's keys and kind words (`testdata/listing_shape.txt`), adds its `CHANGELOG.md` entry under `Unreleased` in the same PR, and so does a change to a flag, an exit code, an exported name, a fence, a builtin or the lowest Go; what is major is the README's Versioning table.
- One-line commit messages: no ticket prefix, no repo name, no authorship trailers.
- GitHub is where the project lives, so `gh` is the tool: merges are fast-forward
  only, `gh pr merge --rebase` where the branch is already on top of main.
- Until the repository has 10 GitHub stars, CodeRabbit reviews a pull request only when
  asked: once a PR is marked ready, comment `@coderabbitai review`. Its verdict lands in the
  `CodeRabbit` check and its walkthrough comment, and any finding as a review comment.
- Hard tabs. No comment by default; delete a restatement, a rationale, history, or a file preamble.
- Prose naming a source states what that source states: read the register's own field
  before paraphrasing it, and name the register you read before claiming none
  publishes a spelling.
- A locale page's credit sentence carries each element that `DATA-LICENSES.md`'s attribution for the source lists, and a
  claim about generated text is checked against a draw that could break it.
- Public texts (README, Docker Hub, GitHub description, package descriptions) use plain words and short sentences with few commas, and read like a person wrote them. They say that related values match, such as a postcode and its city, not only a person's. A text too short to name which checkers refuse an id says that ids pass validation "where possible", since a strict checker refuses some test series. Call the `--repeat` cap "a million rows", not "millions"; item 544 may change the cap, and then this rule changes with it.
- A README example is a `json` block that loads and renders as a category; `readme_test.go` runs every one.
- Cyclomatic complexity is gated at 7: split a function that would go over it.

## Project principles

- A company category whose country has no open bulk file of company names weighs its legal forms by the counts the country's statistics office publishes, and draws its names from a language model's recall. Each name shape and word weighs the same. Goal 3.2. Applied by `da_DK.company`, `de_DE.company` and `fi_FI.company`.
- A name table for which no open bulk file of name counts exists under a licence the project takes, not even one covering part of the country, keeps any counts a source publishes, and draws every other name from a language model's recall. The commoner recalled names are ranked from the most to the least common. Each weighs a function of its rank, which the decision applying this principle states, and never more than the smallest published count. The rarer ones follow in no order and share one weight, no higher than the last ranked name's. The README and `DATA-LICENSES.md` call these weights estimates. Goal 3.2. Applied by `da_DK.first-name`, `da_DK.last-name`, `de_DE.last-name`, `lt_LT.first-name`, `lt_LT.last-name`, `lv_LV.last-name`; `et_EE.first-name` and `et_EE.last-name` apply it until item 522 ships.
- A national id draws a series its issuer keeps for tests or never issues, where one is published. The README says which checker, if any, refuses it. Goal 3.3. Applied by `de_DE.steuer-id`, `fi_FI.henkilotunnus`, `nb_NO.fodselsnummer`, `nn_NO.fodselsnummer`, `sv_SE.personnummer`, `sv_SE.personnummer-12`, `sv_SE.samordningsnummer` and `sv_SE.samordningsnummer-12`.
- A national id whose issuer publishes no test series and no unassigned range draws valid ids for people born over the last hundred years, weighted by age band. Such an id may be a living person's, and the locale's page says so. Goal 3. It breaks goal 3.3 until item 23 ships. Applied by `da_DK.cpr`, `et_EE.isikukood`, `fr_FR.nir`, `lt_LT.asmens-kodas` and `lv_LV.personas-kods`.
- A locale's `word` holds an open dictionary's words, one table per part of speech, under `noun`, `verb`, `adj` and `adv`, and draws each word evenly. It drops no word for being rude. Goal 3.2. Applied by `da_DK.word`, `de_DE.word`, `en_US.word`, `et_EE.word`, `fi_FI.word`, `fr_FR.word`, `lt_LT.word`, `lv_LV.word`, `nb_NO.word`, `nn_NO.word` and `sv_SE.word`. Until item 497 ships, each of their imports but `et_EE.word`'s, `fr_FR.word`'s, `lt_LT.word`'s and `lv_LV.word`'s still drops rude words.
- A geo module weighs each locality by a language model's guess of its population where no open source publishes one under a licence the project takes, not even a larger area's to share out among its localities. `docs/locales/geo.md` and `DATA-LICENSES.md` call the guess an estimate. Goal 3.2. Applied by `geo.EE`, for the four localities formed since its census.
- A geo module that takes its localities from an address register leaves out a locality with no street. Goal 3.2. Applied by `geo.EE`, `geo.LT` and `geo.LV`.

# Decisions

In [docs/decisions.md](docs/decisions.md):

- GitHub is canonical, and the module path names it
- The Gitea copy stays, as a pull mirror of GitHub
- The vocabulary sits in `internal/engine`'s package doc
- Options and fields share one namespace
- `{a|b}` stays beside nested choices
- Flags follow getopt_long
- A struct tag is a template by its shape
- The CLI renders exactly the template on its stdin, and adds nothing
- The CLI reads only the library's public API
- A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129
- Whitespace around a JSON object, array, string or null is dropped, and a bare number or boolean is a format string
- `FakeTemplate` and `NewTemplate` both stay
- The shipped data is embedded, not discovered
- The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0
- A shipped package's folder spells its namespace, its Go name is that namespace in lower case without separators, and its manifest mounts it with `namespace`
- A source whose manifest carries an index loads each category on first reach; any other loads in `New`, with what it reads
- Every `{…}` draws afresh, and only a name keeps a pick
- Reference sigils follow the filesystem
- A change to what exists is a major; a minor only adds
- An error is a contract by what it names, not its bytes
- A load reports every mistake at once
- Formats are checked after all data is loaded, so their errors name the category
- Raising the lowest supported Go is a major
- The format check runs on the latest Go only
- The changelog heading is the one spelling of a release; CI cuts the tag
- A `--data-path` override rebinds every reference to the category it replaces
- The repeat cap bounds renders, not bytes
- 64-bit targets only
- A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`
- Samples say what they emit, transforms what they do
- A record is a template seen as columns; a Go struct is the one second schema
- A struct's records follow Go's field access, and compile on first use
- A category may reference itself, and a record's fences run at load
- A record's column set is fixed before the first draw
- Null is a `null` item, not a rate
- A typed column holds one value, not composed text
- A column that only reads one reference or name is the column it reads
- A typed column's calc is refused unless proven
- `Column` carries text, not a Go value
- A unit takes the stem of what it is, and a file the stem of the units it holds
- One name, one meaning
- The performance gate asserts allocations, not wall-clock time
- Rows live in a TSV, the shape in JSON
- A selector is bracketed, and a dot inside it is literal
- `parent` names the link column and the table alike
- After a row, a path names a column or a linked table
- A selected row is fixed, not drawn
- `..` steps up to the parent row, and a step down after it draws afresh, or once per name's pick
- A table selects by one key column, a code, never a free-form name
- A parent row with no child row is a load error
- A table is a record of string columns
- The key map is built at load, the rest on first draw
- Two categories may name one TSV
- A cell may read a reference, and each row its own
- A path is walked once without drawing before it is walked for real
- A country's postal codes and streets are siblings under its locality
- A locale package imports the data packages it reads, and its `Modules()` returns them all
- The default embed holds every Swedish postort the import can place, with all its streets, and, of the US places that hold most of a postal code's land, every state capital, every place of 100,000 or more and a population-weighted tenth, each with a quarter of its streets, at least ten
- A locale's `address` binds its country's record once and prints it whole
- A postort's kommun comes from its name, its tätort, the SCB småort holding its centre or its codes, never from distance
- A highway designation is not a street, and a US postal code belongs to the place holding most of its land inside places
- `--list` stays a plain list of paths
- A layout is always quoted
- A title is a table under `sex`
- A table owns the spelling of a selector on it
- A builtin reads the date the caller gives, never the clock, so `born` bounds a date by an age
- `New` requires `WithToday`, and `--today` defaults to the system's date
- `sv_SE.personnummer` and `sv_SE.personnummer-12` draw four age bands, up to age 113; `personnummer` is a table under `sex` with a column per band, and `personnummer-12` has no `ten`
- `misc` is what every locale shares
- A register's canonical spelling loses to the one its domain writes
- `misc.territory` is the spine, and a `misc` table naming a territory links to it
- A table whose register publishes no frequency draws evenly
- `misc.timezone` weighs a zone by the people living in it
- `misc.creditcard-test` and `misc.iban-test` are tables of published test numbers
- `misc.car` is one flat table, not a make linked to its models
- `misc.territory` names its sovereign in a column, and there is no `misc.country` table
- `misc.loglevel` is a table keyed by POSIX's keyword, carrying the code
- `misc.tld` keys carry the leading dot, where other tables key on a bare code
- `misc.tld` is a table of its own, and `misc.territory.tld` stays a column
- `misc.territory` carries a currency code, it does not link to `misc.currency`
- An extension may name two media types
- The Finnish id draws DVV's unissued 900 to 999, and `henkilotunnus-issued` the issued 002 to 899 through `{hetu(x,'issued')}`
- The US given names come from a mirror of the SSA file
- `List` advertises direct descents only
- A path draws through its compiled steps
- A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read
- comprehension floor: every dimension and the overall at 7.0 or above is the target, and blocks no merge
- The node types sit in `internal/engine/tree`, compiling, binding, linking and resolving in `internal/engine/load`, every proof and fence of the finished tree in `internal/engine/fence`, with the name rules in its package `names`, drawing in `internal/engine/render`, and the pipeline and lazy loading in `internal/engine`, with the root the public API and its options; compile builds each node with its syntax, and `LoadFinds.Finish` writes into it what its load found, save a folder's children, which its load puts, and a table's `siblings`, which `load.LinkTable` sets; what reads no engine type, a table's rows included, sits in `internal/`; `internal/shipped` generates the files about the shipped data from a loaded tree
- Goal 4.1's `{…}` covers `{{`, `}}` and a lone `}`, so the goal names no escape
- Goal 6.2 holds while a first template can be written with no escape; a literal brace still takes `{{` or `}}`
- Locales ship by region: the Nordic countries, then the rest of the EU, then the rest of Europe, then the rest of the world
- A country with no open address register covering it whole draws invented streets
- `lv_LV`'s first names are the register's first given names, and its last names a language model's recall
- The project takes open licences, share-alike ones included
- The apt, dnf and apk repositories are static files on GitHub Pages, holding the latest release only
- The image is pushed to GHCR and copied to Docker Hub, whose name the README shows
- A known vulnerability in the Go version or a required module blocks a release, even in code fejkdata never calls, and blocks no merge
- CI releases each update to what a release carries at once, as a fejkdata patch cut from the last release's tag
- A `todo.md` item names goal 1 in its Moves when it makes fejkdata better for the people using it, not only when it wins more of them
- Goal 1.1's "where they look" means LLMs first: the ones developers ask, and the coding agents that pick a tool
- A `slugify` transform makes emails from compound names valid, and the shipped data keeps the compounds
- `ascii` folds the Latin letters of ruby-i18n's default table and the Vietnamese ones, except `ŉ`, so names such as `Wałęsa` and `Nguyễn` keep their letters
- A pull request that changes a workflow or what a release packages runs the Packages workflow as a dry run, and `gate` is `main`'s one required check
- CI keeps Go's build and module caches between `main`'s runs, and a pull request's build starts from them
- `fr_FR.company` draws five forms, names an `EI` by its owner, and takes its words from Sirene
- A Swedish street's postal codes keep delivery digit 9
- A Swedish postort with no published population of its own weighs 200
- A Swedish road segment inside a tätort leans towards the postort the tätort is named after, by distance over the fourth root of its codes
- `en_US.word` holds WordNet's words, under one short path per part of speech, and its notice ships with them
- `de_DE.word.noun` carries `plural`, where the nouns of `nb_NO`, `nn_NO`, `da_DK` and `sv_SE` carry `definite`
- `fr_FR.word` holds the English Wiktionary's French words, and its nouns carry `plural`
- `et_EE.word` holds Wikidata's Estonian words, and its nouns carry `plural`
- `lt_LT.word` holds the English Wiktionary's Lithuanian words, and its nouns carry `plural`
- `lv_LV.word` holds Tēzaurs.lv's Latvian words in use, and its nouns carry `plural`
- `{n.w}` is refused beside a read of `n` that reads `w` through a name of `n`'s category
- No column is declared unique; `{seq()}` makes a column unique
- `de_DE`'s first names are Munich's newborns
- A `hidden` field writes no column and is left out of the listing
- No database or HTTP sink is built in
- A date offset is `adddays(d,n)`, and its `d` reads `date` or `born` calls in one layout that reads back
- A where clause follows its path: `length 5`, `length 4 to 6`, `not 'cat'` and `not containing 'cat'`, joined by `and`
- A distinct draw is `distinct N` after any `where` and before `as`; its members read `[1]` to `[N]`, and `joined 'sep'` prints them as one
- `{reword(text,words)}` replaces each word of `text` with one drawn from the path `words`
- A where clause also filters a call drawing from an alphabet, reads a name unquoted, and may end in `else` and a fallback
- One run writes several linked templates through a repeated `--emit <template>=<count>`, and a template reads another's rows with `{/run.<template>}`
- `geo.FI`'s streets come from Posti's basic address file, which holds none on Åland
- A `geo.FI` `swedish-name` holds the Finnish name where its source gives no Swedish one
- `geo.EE` reads its places and addresses from one file, sums a linnaosa into its linn, and a place with no street ships no place
- `geo.LV` takes a novads or a state city as its region, counts each apartment as an address, and writes the postal code after the locality
- `geo.LT` names an apskritis and a savivaldybė with its kind, counts each of the premises as an address, and ships no point
- `geo.EE`, `geo.LT` and `geo.LV` weigh a place by its people and a postal code or a street by its addresses, and the localities a source doesn't count share, by their addresses, their municipality's people that it counts under no locality
- `geo.EE` takes its EHAK codes from settlement units whose licence was not found
- A table's parent column is read in one pass over its runs of equal cells, and a table larger than one block loads by block from a block index `go generate` writes
- Bulk speed comes from fewer allocations, a higher GC target and rows rendered on every CPU, and a weighted draw stays a binary search over cumulative weights
- The table of every shipped path is a page in `docs/` that the README links
- A first or last name a language model recalled is listed `real`, as a name in use; only its weight is an estimate
- `--null-text` spells a null in `csv` alone, and `csv` quotes a value equal to it
- `pgcopy` escapes a NUL, and `csv`, `sql` and `tsv` write it as it is
- A locale's full numbers are `mobile` and `landline`, and `phone-mobile` and `phone-fixed` stay its prefix tables

# Scoring run

The project values for `~/.claude/skills/comprehension-panel/scoring-run.md`:

- `{language}`: Go
- `{kind}`: a Go library and CLI that renders fake data from JSON templates
- `{domain}`: fake test-data generation and template languages
- `{domain docs}`: the documentation of any fake-data library, such as Faker
- `{3am question}`: a seeded render of `sv_SE.address` prints the postal code of one postort beside a street of another
