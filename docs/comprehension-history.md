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

## 2026-10-09T17:39:40Z, PR #202 at 0bdb100, against 2652779

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | worse | same | worse |
| Inherited architect | same | worse | worse | same | worse |

Mid A, decided by:

- Locality, Overall: `data.go:13` `loadSources`: whether an indexed category loads in New depends on source order and what other sources hold, through before-and-after snapshots
- Locality: `index.go:119` `affectedBy`: trusts the manifest's `reads` to match the templates, with nothing beside the code checking it
- Locality, Shape: `index.go:174` `loadReached`: a rollback batch to reason about, with names that say less than `unloadedReads` and `linkedTables` did
- Shape: `index.go:16` `ErrLoad`: a public error every entry point depends on, in a file named for the index
- Shape: `internal/datafiles/datafiles.go:103` `decodeManifest`: a hand-written strict decoder doubles the package's responsibilities

Inherited architect, decided by:

- Locality, Overall: `data.go:13` `loadSources`: New's load set follows dependencies forward from templates and backward from the manifest's `reads`, and a stale `reads` leaves a broken reader for its first reach
- Locality: `index.go:201` `batch.load`: the unloaded entry is deleted three calls away, and `putBack` reverses it
- Locality, Shape: `data.go:62` `standing`: a `map[string]any` of nodes and entries compared by identity
- Shape: `index.go:1` `index.go`: fifteen units under generic names in a file named for the index, `ErrLoad` among them
- Shape: `internal/datafiles/datafiles.go:103` `decodeManifest`: a strict JSON decoder shares the tree walker's file with no seam

## 2026-10-10T09:09:41Z, PR #202 at d34cc42, against 2652779

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | better | same | better |
| Inherited architect | same | better | same | same | better |

Mid A, decided by:

- Locality, Overall: `data.go:12` `loadSources`: one path, each source indexed or walked then one `loadReached`, where `config.load` switched between a lazy and an eager mode
- Locality: `index.go:107` `loadReached`: a failed load puts back what it loaded and returns an error, where `loadShipped` panicked partway through
- Locality, Shape: `node.go:31` `folder.put`: two methods keep a name in `children` or `unloaded`, never both
- Shape: `index.go:29` `categoryAt`: `shipped.go` is only the embed, and `categorySite` embeds `categoryAt` in place of a duplicated `unloadedCategory`

Inherited architect, decided by:

- Locality: `data.go:12` `loadSources`: one load path, where `config.load` hid a mode switch between lazy and eager loading
- Locality: `index.go:107` `loadReached`: returns a load error, and names a manifest entry drifted from its table, where `loadShipped` panicked on a hand-regenerated Go file
- Locality: `node.go:23` `folder`: the children-or-unloaded rule stated once and kept by two methods

## 2026-10-10T09:21:22Z, PR #202 at b77ec94, against 2652779

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | same | same | better |
| Inherited architect | same | worse | same | same | worse |
| Maintainability senior | same | better | same | better | better |
| Junior A | same | better | better | same | better |

Mid A, decided by:

- Locality, Overall: `node.go:31` `folder.put`: two methods keep a name in `children` or `unloaded`, where `unloadedCategory.load` kept it by a separate delete
- Locality, Overall: `data.go:22` `addSource`: each source's own manifest decides how it loads, where adding any data path switched the whole shipped set to eager loading

Inherited architect, decided by:

- Locality, Overall: `index.go:109` `loadReached`: one function serves New and first reach, rolls back through `putBack`, and rests on facts held elsewhere
- Locality: `index.go:22` `placeIndex`: how sources merge depends on their order and on which kind each is, and the children-or-unloaded rule holds only while `put` is the one writer

Maintainability senior, decided by:

- Locality, Overall: `data.go:12` `loadSources`: every source goes through one path, where `config.load` branched between two load paths of different eagerness
- Locality: `index.go:15` `indexed`: an unloaded entry carries its own source, where `unloadedCategory.load` was wired to the global `shippedSource`
- Self-sufficiency, Overall: `internal/datafiles/manifest.go:56` `decodeManifest`: the index is data checked strictly, where `shippedEntry` told the reader to empty a generated literal by hand

Junior A, decided by:

- Locality, Shape: `node.go:23` `folder`: `put` and `putUnloaded` keep a name in one map, where `compileInto` and `unloadedCategory.load` kept it as a pair
- Locality, Shape: `data.go:12` `loadSources`: one load path, where `config.load` forked and `loadShipped` was a second, panic-only pipeline
- Shape: `data.go:81` `categorySite`: embeds `categoryAt`, where `unloadedCategory` repeated its fields

## 2026-10-10T11:11:16Z, PR #203 at 421bb49, against bf1becb

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | worse | same | worse |
| Inherited architect | same | worse | worse | same | worse |

Mid A, decided by:

- Locality, Overall: `node.go:28` `folder.reads`: a third per-name map beside children and unloaded with no invariant tying it to them; every placement site must remember a separate setReads call (`data.go:36` addSource, `data.go:76` mergeFolder, `index.go:27` placeIndex)
- Locality, Shape: `resolve.go:96` `withReads`: the manifest's reads travels through folder, categorySite.reads and templateSite.category to feed one error clause, four files to hold for one sentence of error text
- Shape: `internal/datafiles/manifest.go:22` `Manifest.Reads`: reads as a load dependency the loader acts on, though it only feeds error text, and "reads" already means a template reading a name or column
- Shape: `data/geo/SE/fs.go:2` `SE`: package names SE, US, en_US, sv_SE; in Modules SE.FS sits beside sv_SE.FS with nothing saying which is geography
- Locality: `fejkdata.go:94` `New`: loads exactly what the options name, easier than the hidden config{shipped: true} default, though config.given brings back a smaller flag

Inherited architect, decided by:

- Locality, Shape, Overall: `node.go:27` `folder.reads`: a third per-name map that put and putUnloaded leave alone, so addSource, mergeFolder and placeIndex must each call setReads, and batch.load relies on placeIndex having done it; action at a distance for one error suffix
- Locality: `resolve.go:94` `templateSite.withReads`: reads travels folder, categorySite.reads, a *categorySite per template, then withReads, four hops across three files
- Shape: `data/geo/SE/fs.go:2` `SE`: bare package names SE and US do not say geography at WithDataFS(sv_SE.FS, SE.FS); the namespace-doubling rule lives only in README and test code
- Shape, Self-sufficiency: `internal/datafiles/manifest.go:22` `Manifest.Reads`: suggests it drives loading though it only feeds withReads; "by default" needs README goal 9.2
- Locality, Navigation: `data/modules.go:17` `Modules`: the hidden shippedFS default is gone; Modules lists the modules, errNoData names the fix, and the CLI loads data.Modules() in plain sight
- Locality: `fejkdata.go:40` `config.given`: WithDataFS() with no arguments differs from no option, learnt only from its comment

## 2026-10-10T11:24:33Z, PR #203 at 70e552c, against bf1becb

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | worse | same | same |
| Inherited architect | worse | worse | better | same | worse |
| Maintainability senior | worse | same | worse | same | worse |
| Junior A | same | worse | same | same | same |

Mid A, decided by:

- Locality: `fejkdata.go:94` `New`: loads only the modules its options name or fails with errNoData; the base silently added the embedded shippedSource in config.load
- Shape: `internal/datafiles/manifest.go:23` `Manifest.Reads`: Reads and categorySite.defaultModules suggest a loaded dependency, though only a hint reads them, and "by default" is never set against anything
- Locality, Shape: `data.go:12` `loadSources`: default modules reach a category by two routes, the walked map and indexed.defaultModules, and templateSite gains a pointer only to carry them
- Shape: `data/modules.go:18` `Modules`: package names SE, US, sv_SE, en_US are unidiomatic, and every embedded path repeats itself

Inherited architect, decided by:

- Locality, Overall: `data.go:12` `loadSources`: addSource fills a side map keyed by path that loadSources copies onto sites after the merge, so merge order and clashes must be reasoned about on the core load path
- Locality, Shape: `data.go:92` `categorySite.defaultModules`: the manifest's module list passes four carriers, set at two distant places and not by siteIn, so a future caller resolving siteIn sites gets a false "names no module" message
- Navigation: `data/geo/SE/geo/SE/address.json` `geo.SE.address`: every shipped file sits under a stuttered path
- Shape: `fejkdata.go:65` `WithDataFS`: the root no longer embeds data, datafiles.Source drops base, and data sits behind data/modules.go and five small packages
- Navigation, Shape: `internal/datafiles/manifest.go:23` `Manifest.Reads`: import paths that look resolved though nothing resolves them, generated by a test reached only through go:generate

Maintainability senior, decided by:

- Navigation, Overall: `data/geo/SE/geo/SE/address.json` `geo.SE.address`: the path repeats the namespace inside the module, one extra hop from category to file, two for geo
- Shape: `internal/datafiles/manifest.go:21` `Manifest.Reads`: reads sounds like a loaded dependency, and the same list is also defaultModules and walked
- Navigation, Shape: `data/modules.go:15` `go:generate`: regenerating a module's manifest runs a test one package up
- Shape: `data.go:20` `loadSources`: defaultModules reaches categorySite by two routes, and siteIn leaves it empty

Junior A, decided by:

- Locality: `data.go:12` `loadSources`: default modules pass through four files, an out-parameter map patched onto sites after the merge for walked sources and indexed.defaultModules copied in batch.load for indexed ones
- Locality: `fejkdata.go:94` `New`: reads straight through with no hidden default source, which only partly offsets the above
