# Comprehension history

## 2026-10-09T08:59:19Z, PR #190 at 17f732a, against ee4b40b

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `pick.go:9` `pickKey`: the definition sits on the type, beside `levelKeys`, `addressedKeys`, `keptInPick` and `readUnder`, where four files held them
- `render.go:155` `readField`: one named predicate, `keptInPick`, states the whole keep rule
- `datatype.go:217` `columnRead`: beside the code that reads it

Inherited architect, decided by:

- `pick.go:9` `pickKey`: the 3am path for a named pick now ends in one file, with its worked example
- `render.go:155` `readField`: `keptInPick` holds its own nil and reference checks
- `path.go:72` `compiledPath`: the field no longer reuses the vocabulary word level

## 2026-10-09T09:24:31Z, PR #190 at 65a32db, against ee4b40b

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `pick.go:9` `pickKey`: the definitions sit on the type, beside the functions that build and read the keys
- `datatype.go:217` `columnRead`: beside `itemDatatype` and `checkColumns`, which consume it
- `node.go:70` `template.readsColumn`: names the four readers that change together

Inherited architect, decided by:

- `pick.go:9` `pickKey`: a pick-key bug needs one file, where it needed four and the vocabulary
- `render.go:155` `readField`: one named predicate, `keptInPick`, states the keep rule
- `path.go:72` `compiledPath`: `headSpans` is the costliest new read in the diff

## 2026-10-09T09:30:15Z, PR #190 at d9f0fc9, against ee4b40b

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `pick.go:9` `pickKey`: the definitions and the code building and reading the keys share one file
- `render.go:155` `readField`: `keptInPick` holds the whole keep rule and its own nil check
- `path.go:74` `compiledPath`: `headSpans` and the bare 0 its callers pass cost a second read

Inherited architect, decided by:

- `pick.go:9` `pickKey`: the 3am mechanism reads in one file, in reading order
- `datatype.go:217` `columnRead`: beside its consumers, and `node.go:70` names its four readers
- `path.go:119` `compileStep`: six parameters, the one place that reads harder

## 2026-10-09T10:16:38Z, PR #193 at c27e8b7, against d74c0ed

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `data-import/geo.py:75` `fence`: one named step states the locality row and drops childless parents level by level, where `with_streets` pruned localities alone
- `data-import/geo-us.py:174` `main`: the county and region pruning no longer hides inside the `tsv.write` comprehensions
- `data-import/geo-se.py:186` `main`: builds its rows before the fence, so it has the US script's shape

Inherited architect, decided by:

- `data-import/geo.py:75` `fence`: the loader's parent-has-a-child rule is named in the import code and applied in one place at all three levels
- `data-import/geo-us.py:174` `main`: one `fence` call before any write replaces filters buried in the writes
- `data-import/geo-se.py:186` `main`: both scripts read alike, build rows, fence, write

## 2026-10-09T10:29:19Z, PR #193 at 8433689, against d74c0ed

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `data-import/geo.py:75` `fence`: the parent-has-a-child rule lives in one function whose docstring states the rule, each argument's shape and what it returns
- `data-import/geo-us.py:174` `main`: `zctas` is built once and fenced, and every write reads from `kept` with no filter of its own
- `data-import/geo-us.py:128` `zcta_localities`: the name says it returns a ZCTA-to-place map, where `postal_codes` pointed at the wrong output

Inherited architect, decided by:

- `data-import/geo.py:75` `fence`: one unit owns the chain from locality to municipality to region, where the base spread it over `with_streets` and both `main`s
- `data-import/geo-us.py:128` `zcta_localities`: the name and docstring say what it maps to
- `data-import/geo-se.py:181` `main`: `region_names` and `municipality_names` say they map codes to names, and the rows are built once before the fence

## 2026-10-07, scoring run at ad59967

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mean | 7.00 | 5.88 | 6.50 | 6.13 | 6.38 |

## 2026-10-09T11:04:29Z, PR #195 at b9d1af6

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7.5 | 6.5 | 7 | 7 | 7 |
| Mid A | 7.5 | 6 | 6.5 | 7 | 6.5 |
| Maintainability senior | 7.5 | 6 | 6.5 | 6.5 | 6.5 |
| Inherited architect | 7 | 5.5 | 6 | 7 | 6.5 |
| Mean | 7.38 | 6.00 | 6.50 | 6.88 | 6.63 |

Junior A, hardest first:

- `pick.go:70` `keptInPick`: a level, a first-level key, that key moved under `env.pickAt`, and the name's addressed set, held at once
- `namefence.go:76` `refuseTwiceDrawnIn`: mirrors the keep rule in `keptInPick`, and nothing ties the two in code

Mid A, hardest first:

- `pick.go:79` `readUnder`: a key built at load in `addressedKeys` must equal one built at render through `pickKey.under`, and a mismatch is silent
- `node.go:70` `readsColumn`: four functions in three files must change together

Maintainability senior, hardest first:

- `render.go:146` `readField`: whether a read is kept in a pick rests on string keys from three places and on `env.pick` cleared at the right moment
- `path.go:190` `drawSteps`: a nil `pins` or `memo` names the mode, and `climbed` carries state across iterations

Inherited architect, would restructure first:

- `pick.go:179` `readName`: keeping one named pick needs `name.go`, `pick.go`, `env.go`, `render.go`, `path.go` and `record.go` held at once
- `node.go:52` `template`: filled in four phases by four functions

## 2026-10-09T11:09:34Z, PR #195 at 344023d, against 1c342bc

Ruling: better

| Seat | Vote |
|---|---|
| Mid A | better |
| Inherited architect | better |

Mid A, decided by:

- `shipped.go:61` `loadCallerPath`: `Fake` and `FakeRecord` get their segments from the one call that loads them, where each repeated the split and the load
- `node.go:30` `choice`: the `shared` comment says what the field counts, where it claimed `List` reads it
- `data-import/tsv.py:7` `write`: its check and message match the loader's one-row rule

Inherited architect, decided by:

- `shipped.go:61` `loadCallerPath`: the load-before-walk order is held by data flow, where a comment on `loadShippedAt` stated it
- `node.go:30` `choice`: the `shared` comment no longer sends a `--list` bug to the wrong field
- `README.md:163` `## Data`: the opening names the `geo` trees the section documents

## 2026-10-09T13:47:17Z, PR #199 at aea1f9d, against 2652779

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | same | worse | worse |
| Inherited architect | same | worse | same | worse | worse |

Mid A, decided by:

- Locality: `version.go:3` `Version`: one version string must agree in the changelog heading, the constant, the CLI's require and the tags, and a `_test.go` file rewrites source elsewhere under REGENERATE
- Locality, Self-sufficiency: `compose.yaml:14` `test`: every command builds its package list with `go list -m`, and why the CLI needs its own module sits only in docs/decisions.md
- Self-sufficiency: `README.md:71` "Install main's CLI from a checkout": the sentence names `@main`, which is no checkout install, and contradicts itself

Inherited architect, decided by:

- Locality: `version_test.go:19` `TestVersionIsTheNewestChangelogHeading`: one release spreads over seven places, and the test rewrites version.go and every go.mod under REGENERATE
- Self-sufficiency: `cmd/fejkdata/go.mod:5` `require`: read cold it pins a version that does not exist, and why it works sits in other files
- Self-sufficiency: `README.md:71` CLI section: the label reads as an instruction for the command the sentence warns against
- Locality: `compose.yaml:47` `tidy`: runs on the root module only while the other services cover every module, with nothing at the service saying so

## 2026-10-09T14:48:51Z, PR #199 at b3411ec, against 2652779

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | worse | worse | worse |
| Inherited architect | worse | worse | worse | worse | worse |

Mid A, decided by:

- Locality, Shape: `release-tooling/publish_release.py:111` `publish`: deciding between rerun, burnt and proceed needs the whole release-commit scheme in mind, and its names say little
- Locality: `release-tooling/publish_release.py:58` `release_go_mods`: which requires ship is spread over the compose service, the script, git ls-files and a gate test
- Self-sufficiency: `cmd/fejkdata/go.mod:1` module: imports the library with no require, which only go.work and the release docs explain
- Self-sufficiency: `compose.yaml:14` test: the go list -m pattern is explained only through a decisions anchor
- Shape: `compose.yaml:55` tidy: covers the root only while its siblings cover every module

Inherited architect, decided by:

- Locality, Navigation: `release-tooling/publish_release.py:58` `release_go_mods`: the go.mod a user installs exists only on a release branch, written from a listing made in another container
- Shape: `release-tooling/publish_release.py:122` `publish`: two compound conditions over competing names decide the tag states, where the base had one linear check
- Self-sufficiency: `Dockerfile:7` comment: the go list -m pattern is explained only by a decisions anchor
- Locality: `compose.yaml:17` test: the module-spanning list is copied into six services while generate, cyclo and tidy stay on the root
- Navigation: `.github/workflows/test.yml:77` release job: tracing a bad release input crosses three files
