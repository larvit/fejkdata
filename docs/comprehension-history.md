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

## 2026-10-09T16:48:45Z, PR #201 at c3ef728, against 2652779

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-09T16:54:01Z, PR #201 at 8b5a972, against 2652779

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-09T17:21:57Z, PR #202 at 65b2de8, against 2652779

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | same | same | worse |
| Inherited architect | same | worse | worse | worse | worse |

Mid A, decided by:

- Locality, Overall: `data.go:13` `loadSources`: one pass mixes lazy and eager loading through before/after snapshots, a `provides` closure and a reverse graph search, and the skip of the first source is unexplained
- Locality: `data.go:107` `mergeFolder`: a category must sit in one of two maps, `children` or `unloaded`, kept by deletes in four functions across two files
- Locality: `internal/datafiles/datafiles.go:83` `embeddedManifests`: a process-wide cache in a package that was otherwise pure

Inherited architect, decided by:

- Locality, Self-sufficiency: `data.go:13` `loadSources`: nothing says why replaced paths drag their readers in, or what the `!after[p]` case covers
- Locality: `index.go:113` `loadReached`: its rollback claims the next call fails the same way, which holds only while no parent table was linked to a rolled-back child
- Shape: `index.go:159` `reaching`: four near-synonyms, `reaching`, `reachedFrom`, `loadReached` and `referenced`, and `referenced` and `categoryUnder` return loaded categories too
- Shape: `data.go:40` `place`: shares its name with `compileInto`'s parameter, and returns a closure answering one question two ways
- Locality: `internal/datafiles/datafiles.go:83` `embeddedManifests`: hidden global state for a small saving

## 2026-10-09T17:32:44Z, PR #202 at fdc6a04, against 2652779

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | same | same | worse |
| Inherited architect | same | worse | worse | worse | worse |

Mid A, decided by:

- Locality, Overall: `index.go:175` `loadReached`: one loop loads, puts back on failure, checks each entry before linking and again after the pipeline, and its rollback rests on an ordering invariant its comment argues
- Locality: `data.go:13` `loadSources`: whether a category loads in New means holding every source and every index entry together, through `standing` snapshots and `dependents`
- Shape: `index.go:36` `categoryAt`: a near-synonym of `categorySite`, beside `dependencies` and `dependents`, which differ by two letters and run in opposite directions

Inherited architect, decided by:

- Locality, Overall: `index.go:175` `loadReached`: five jobs in one queue loop, with a cross-call ordering invariant to hold
- Locality, Shape: `data.go:13` `loadSources`: one source can make another load early through identity-compared snapshots and a reverse graph over index entries
- Shape: `index.go:36` `categoryAt`: two near-synonym types for where a category sits, and three near-identical function names
- Self-sufficiency: `index.go:248` `staleEntry`: its note on linking cannot be read without knowing the pipeline, and `config.load` points at a paragraph-long decision anchor
