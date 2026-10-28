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

## 2026-10-11T09:52:08Z, PR #203 at 4c4bbcc, against bf1becb

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | worse | same | worse |
| Inherited architect | worse | same | worse | same | worse |

Mid A, decided by:

- Shape, Overall: `data/geo/SE/fs.go:2` `SE`: each module embeds a tree repeating its own name, and the package names SE, US, en_US, sv_SE are unidiomatic; five thin packages plus a list replace one embed
- Shape: `reference.go:124` `withDefaultModules`: "default modules" suggests modules that load by default, though nothing loads them; "by default" is spelled out only in the README
- Locality, Shape: `inline.go:40` `NewTemplate`: errors are decorated twice, by hand, at four entry points
- Locality: `fejkdata.go:101` `New`: the hidden shipped default is gone, offsetting the above

Inherited architect, decided by:

- Navigation, Overall: `data/sv_SE/fs.go:8` `FS`: every data file sits under a doubled folder, data/sv_SE/sv_SE/, data/geo/SE/geo/SE/, a stop at each hop of the 3am path
- Navigation: `data/modules.go:15` `go:generate`: the module list is written twice and the manifest writer lives in a root test run from the data package
- Shape: `data/geo/SE/fs.go:2` `SE`: a package named SE reads as Sweden beside sv_SE, and "module" means a Go package here
- Shape: `reference.go:126` `withDefaultModules`: one concept spelled reads, defaultModules and "read by default", the last defined only in the README
- Locality: `fejkdata.go:114` `New`: four call sites apply loadError, withDefaultModules and noDataNote in a fixed order, offset by the hidden shipped source being gone

## 2026-10-11T10:44:13Z, PR #203 at 36b5062, against bf1becb

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | worse | worse | worse | same | worse |
| Inherited architect | worse | worse | same | worse | worse |

Mid A, decided by:

- Locality, Shape, Overall: `index.go:110` `loadReached`: a reads parameter runs through four layers only to add a note to an error, and reads already means a template's reads
- Navigation: `internal/datafiles/datafiles.go:36` `Source.Mounted`: a shipped module's broken file is labelled geo/SE/address.json, a path that no longer exists
- Locality: `internal/datafiles/datafiles.go:77` `Source.Walk`: Walk adds the mount and Load removes it, two forms of a folder path in one type
- Locality: `fejkdata.go:68` `WithDataFS`: the shipped data no longer loads silently, a partial offset
- Shape: `reference.go:127` `withMissingModules`: a typo in manifest-less data still gets a note about modules

Inherited architect, decided by:

- Locality, Overall: `index.go:110` `loadReached`: reads threaded through six signatures for one error note
- Navigation: `internal/datafiles/datafiles.go:36` `Source.Mounted`: errors name sv_SE/address.json, a path that does not exist, where the base named the real data/sv_SE/address.json
- Navigation, Self-sufficiency: `data/svse/.fejkdata.json:26` `namespace`: a namespace no longer mirrors its folder, so the path alone no longer finds the file
- Self-sufficiency: `internal/datafiles/manifest.go:23` `Manifest.Reads`: its comment and withMissingModules' only make sense after the README
- Locality: `internal/datafiles/datafiles.go:77` `Source.Walk`: the mount is translated in opposite directions by Walk and Load

## 2026-10-11T11:30:01Z, PR #203 at 2a4694b, against bf1becb

Ruling: worse

Fell: Navigation. Under the floor of 7.0: Navigation 6.88, so the same ruling counts as worse.

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | worse | better | same | same | same |
| Inherited architect | worse | same | same | same | same |

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7 | 5.5 | 6 | 6 | 6 |
| Mid A | 7 | 5.5 | 6.5 | 6 | 6 |
| Maintainability senior | 7 | 6 | 6.5 | 6.5 | 6.5 |
| Inherited architect | 6.5 | 5.5 | 5.5 | 6.5 | 6 |
| Mean | 6.88 | 5.63 | 6.13 | 6.25 | 6.13 |

Mid A, decided by:

- Locality: `fejkdata.go:66` `WithDataFS`: every module is named at the call site, where the base silently started each generator from a package-level embed
- Navigation: `data/geose/fs.go:1` `geose`: a dotfile manifest decides the namespace and the folder follows a lowered, no-separator rule, one translation step from a failing path to its file
- Navigation: `internal/datafiles/datafiles.go:57` `Source.labelled`: a shipped file's error prints sv_SE (address.json), where the base printed the openable data/sv_SE/address.json
- Locality: `data.go:27` `addSource`: where a category lands depends on a field in a hidden file, a partial offset

Inherited architect, decided by:

- Navigation: `data/svse/fs.go:1` `svse`: the step from a dot path to a folder goes through the lowered-namespace rule, stated only in the README and each package doc
- Navigation: `internal/datafiles/datafiles.go:57` `labelled`: errors name sv_SE (address.json), not a path that opens as written
- Locality: `fejkdata.go:99` `New`: New loads exactly its options, cancelled by `data/svse/fs.go:1`'s dependency on geose that nothing in Go expresses

## 2026-10-11T13:08:34Z, PR #203 at fb2c809, against bf1becb

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | same | same | better |
| Inherited architect | better | better | same | better | better |

Mid A, decided by:

- Locality, Overall: `fejkdata.go:66` `WithDataFS`: the caller names every module and New refuses no data naming the fix, where the base prepended a package-level shippedSource behind config{shipped: true}
- Locality: `data/sv_SE/fs.go:19` `Modules`: sv_SE reading geo.SE is a Go import with a doc line, not only a template reference inside address.json
- Shape: `internal/datafiles/mount.go:13` `mounted`: a hand-built virtual fs.FS replaces a base string, balanced by Source losing a field and Walk, Load and Manifest turning path-agnostic

Inherited architect, decided by:

- Locality, Navigation, Overall: `data/sv_SE/fs.go:19` `Modules`: a locale reading its country's places is a Go import, one hop in code where the base needed the README or a trace through JSON references
- Locality, Self-sufficiency: `fejkdata.go:66` `WithDataFS`: every source is named at the call site, and leaving them all out gives an error that spells the fix
- Navigation: `data/modules.go:17` `Modules`: shipped data has a Go home inside data/, each module's manifest beside its files
- Shape: `internal/datafiles/mount.go:13` `mounted`: seven small types fake the folders above the mount, balanced by Source keeping one form and truthful per-module packages
- Self-sufficiency: `data/sv_SE/fs.go:1` `svse`: each module says in code which namespace it supplies and what it reads

## 2026-10-11T13:18:20Z, PR #203 at 3d4131b, against bf1becb

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | worse | same | same |
| Inherited architect | same | worse | worse | same | worse |
| Maintainability senior | same | worse | same | same | worse |
| Junior A | same | worse | worse | better | worse |

Mid A, decided by:

- Shape: `internal/datafiles/mount.go:13` `mounted`: a hand-built virtual fs.FS of five types replaces a base string joined in dirPath
- Locality, Shape: `internal/datafiles/datafiles.go:34` `Source.Mounted`: one field drives file reads, error labels and the manifest read, and a second Mounted call does nothing
- Locality: `data/sv_SE/fs.go:19` `Modules`: a locale reading its country's places is a Go import with a package comment
- Locality: `fejkdata.go:99` `New`: every source is named at the call site, where the base loaded shipped data implicitly
- Shape: `data/modules.go:17` `Modules`: six packages, five near-identical boilerplate

Inherited architect, decided by:

- Locality, Shape, Overall: `internal/datafiles/mount.go:13` `mounted`: a fake fs.FS of five types with made-up folders above the mount, where the base had a base string; where it rots first
- Locality: `internal/datafiles/datafiles.go:34` `Source.Mounted`: the mount point kept in Source.at and mounted.at, and labelled undoes it through underMount
- Locality, Self-sufficiency: `index.go:22` `placeIndex`: index keys carry the full namespaced path, which only the mount maps back, stated nowhere at either site
- Locality: `fejkdata.go:41` `config.given`: every source explicit and errNoData names the fix, a partial offset
- Navigation: `data/modules.go:14` `go:generate`: the manifest generator lives in a test of the parent package

Maintainability senior, decided by:

- Locality, Overall: `internal/datafiles/mount.go:13` `mounted`: a fake fs.FS inventing the folders above the mount, so tracing why a file opens means following three paths in Open, where the base prefixed one base string
- Locality: `internal/datafiles/datafiles.go:56` `labelled`: one field serves the manifest read, file opens and error labels, and labelled undoes the mount through code in mount.go
- Locality: `fejkdata.go:99` `New`: the hidden shipped default is gone, a partial offset

Junior A, decided by:

- Shape, Locality, Overall: `internal/datafiles/mount.go:13` `mounted`: about 130 lines and five types fake folders so a module's root shows up under its namespace, where prefixing Category.Folders looks like the obvious alternative
- Locality: `internal/datafiles/datafiles.go:34` `Source.Mounted`: at changes both fsys and how errors label files, and a second Mounted call does nothing silently
- Self-sufficiency: `data/sv_SE/fs.go:1` `svse`: each package doc states its namespace and what it reads, and Modules names exactly what it pulls in
- Self-sufficiency: `fejkdata.go:79` `errNoData`: the error names the call that loads the shipped data

## 2026-10-11T18:50:05Z, PR #209 at fbe2fde, against d157e45

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | same | same | better |
| Inherited architect | same | better | better | same | better |

Mid A, decided by:

- Locality, Overall: `cli/cli.go:318` `options`: the modules a run loads come in through Main and run, where the base fetched data.Modules() inside the flag handling
- Locality: `internal/datafiles/datafiles.go:56` `labelled`: reads Source.at, where the base type-asserted the FS to recover the mount point
- Locality, Shape: `internal/datafiles/mount.go:11` `mounted`: two types and one helper, no hidden read flag, where the base held five
- Locality: `internal/datafiles/datafiles.go:34` `Mounted`: the mount point is stored twice, and a second call is silently ignored
- Shape: `cli/cli.go:104` `no-shipped-data`: in a custom CLI the flag drops whatever Main was given, which only Main's doc says
- Shape, Self-sufficiency: `data-import/geo-us.py:219` `main`: the street share is a bare positional float, and places is rebound partway

Inherited architect, decided by:

- Shape, Locality: `internal/datafiles/mount.go:8` `mounted`: Open, ReadFile and ReadDir over one helper, where the base faked a tree in five types
- Locality: `internal/datafiles/manifest.go:53` `ManifestPath`: reads the explicit Source.at, though at is stored twice and only Mounted keeps them in step
- Locality: `cli/cli.go:318` `options`: the shipped set comes in through Main, run and options, where the base fixed it deep in the flag handling
- Shape: `cli/cli.go:104` `no-shipped-data`: once Main takes any modules, the flag's name promises less than it does
- Shape: `data-import/geo-us.py:211` `main`: places is rebound from every incorporated place to the shipped ones
- Locality: `cli/cli.go:651` `moduleVersion`: a pure function over the build info picks fejkdata's version by path

## 2026-10-11T19:02:27Z, PR #209 at 69af4c1, against d157e45

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | better | same | better |
| Inherited architect | same | better | better | same | better |

Mid A, decided by:

- Shape, Locality: `internal/datafiles/mount.go:11` `mounted`: serves Open, ReadFile and ReadDir with one stateless entry type, where the base faked a full tree with a read flag
- Locality: `internal/datafiles/datafiles.go:18` `Source`: the mount point is a field, where labelled and ManifestPath type-asserted the FS; a second Mounted is silently ignored
- Locality: `cli/cli.go:321` `invocation.options`: the modules and the command name arrive as parameters, where the base called data.Modules() in the flag logic
- Shape: `cli/cli.go:678` `versionLine`: the CLI is its own package and main is one line, partly offset by the version line's branches and the COMMAND placeholder

Inherited architect, decided by:

- Shape, Locality: `internal/datafiles/mount.go:11` `mounted`: two types and one pure underMount, where the base faked an fs.File tree in five types with a read cursor
- Locality: `internal/datafiles/manifest.go:53` `ManifestPath`: the manifest reads through s.at whether called before or after Mounted, removing the base's order dependence
- Locality: `internal/datafiles/datafiles.go:34` `Mounted`: a second call is a silent no-op and at lives twice; where it rots first
- Shape: `cmd/fejkdata/main.go:12` `main`: the binary is one line over a library package, where the base bound the CLI to data.Modules() inside the flag handling
- Locality: `cli/cli.go:667` `report`: rewrites the library's prefix inside the joined error text, a string-level contract with another package

## 2026-10-12T16:29:54Z, PR #216 at 52e0304, against 90ffcff

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | better | better | better |
| Inherited architect | same | better | same | same | better |

Mid A, decided by:

- Locality, Overall: `fejkdata.go:101` `New`: the seed is plain input, where the base drew it from crypto/rand through the swappable package-level randomBytes and drawSeed
- Self-sufficiency, Shape: `fejkdata.go:129` `config.misuse`: one named unit collects every option mistake, and errNoSeed tells the caller what to write
- Shape: `fejkdata.go:40` `config.dataGiven`: the name says what it tracks now that seeded sits beside it
- Self-sufficiency: `cli/cli.go:26` `usage`: --help states the --seed default and stays true for a custom CLI
- Self-sufficiency: `internal/jsonvalue/jsonvalue.go:29` `located`: a JSON syntax error names its line and column
- Shape: `record.go:216` `errChoiceRecord`: an inline errors.New became a named sentinel FakeRecord treats alike
- Locality: `Dockerfile:14` `release-tooling`: a /tested sentinel copy is an indirection, though its comment explains it

Inherited architect, decided by:

- Locality, Overall: `fejkdata.go:101` `New`: the seed is a required input and the package-level randomBytes global is gone; New reads straight through options, misuse, load, build
- Locality: `fejkdata.go:129` `misuse`: the base's two early returns become one unit joining every option mistake before any data loads
- Locality: `cli/cli.go:593` `run`: the CLI's fallback seed is a systemSeed parameter drawn once in Main, so everything that changes the output arrives through run's signature

## 2026-10-12T16:30:06Z, PR #216 at 52e0304

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7 | 5.5 | 6.5 | 6 | 6 |
| Mid A | 7 | 6 | 6.5 | 6.5 | 6.5 |
| Maintainability senior | 7 | 5.5 | 6.5 | 6.5 | 6 |
| Inherited architect | 7 | 6 | 6 | 7 | 6.5 |
| Mean | 7.00 | 5.75 | 6.38 | 6.50 | 6.25 |

Junior A, hardest first:

- `pick.go:179` `readName`: six pieces of state, with keptInPick's decision made at load through addressedKeys
- `path.go:190` `drawSteps`: climbed carries state between steps and pins swap mid-loop
- `table.go:253` `table.route`: six-field tableRoute plus fromRow and pinnedBelow held at once
- `node.go:65` `template`: filled in four phases by different files, with a change-together set
- `AGENTS.md:109` `misc.loglevel`: the decision title says keyed by keyword, the JSON keys by code

Mid A, hardest first:

- `pick.go:9` `pickKey`: keys built at load must match keys built at render, with heap-escape rules mixed in
- `render.go:146` `readField`: four read paths, and the name undersells it
- `path.go:190` `drawSteps`: order-dependent climbed and stepDownPins
- `datatype.go:230` `columnReadOf`: three spellings of a column read, with a four-function change-together contract
- `AGENTS.md:109` `misc.loglevel`: the decision title contradicts the JSON

Maintainability senior, hardest first:

- `render.go:146` `readField`: choosing named, kept or fresh needs env.pick, pickAt, under, levels and addressed at once
- `env.go:71` `enter`: frames pushed here must be found by frameOf later, order dependence across three files
- `path.go:190` `drawSteps`: climbed carries state from stepParent to the next stepChild
- `path.go:145` `compiledPath.walkEvery`: a capacity-clipping alias with no comment on why it is safe
- `index.go:109` `loadReached`: lazy loading mutates the shared tree and rolls back with putBack

Inherited architect, would restructure first:

- `pick.go:371` `keptInPick`: string pickKeys rebuilt at render must match addressedKeys built at load, and a miss silently draws afresh
- `node.go:253` `template`: one struct for six concepts, filled in four stages by four files
- `calc.go:219` `checkCalcFields`: calc lives in five places, and the root calc.go holds only load fences
- `datatype.go:222` `columnRead`: column-read resolution sits in datatype.go
- `cli/cli.go:505` `isPath`: copies the path grammar, the CLI's first rot point

## 2026-10-13T22:09:02Z, PR #220 at 909ddc5, against 891323d

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | worse | same | same | better |
| Inherited architect | better | worse | same | same | same |
| Maintainability senior | better | worse | same | same | same |
| Junior A | better | worse | same | better | better |

Mid A, decided by:

- Navigation: `README.md:32` `table of contents` and `README.md:118` `### What stdin renders`: a link row reaches every top-level section, and stdin's byte rules sit under a heading of their own instead of an unheaded block between the CLI examples and the flag table.
- Navigation: `README.md:1171` `Layout` row for `release-tooling/` and `README.md:1277` `Development`: a packaging failure leads straight to `packages.yml` or `package-repos.Dockerfile`.
- Locality: `release-tooling/goreleaser.yaml:60` `isEnvSet "OLDER_TAG"` and `release-tooling/goreleaser.yaml:61` `dockerfile: ../tooling/release-tooling/image.Dockerfile`: the config assumes the workflow's src/ and tooling/ checkouts, and OLDER_TAG is set by a shell step in `packages.yml:72` with no note where it is read.

Inherited architect, decided by:

- Locality: `release-tooling/goreleaser.yaml:60` `dockers_v2` tags: whether the image gets `latest` hangs on an env var only `.github/workflows/packages.yml:72` `Channels holding one version` sets, and that step's `--skip` keeps the one-version rule in a second file by a second mechanism.
- Locality: `release-tooling/goreleaser.yaml:6` `dist: ../dist`: the config holds only under the two-checkout layout of `packages.yml:42`, and the `snapshot` stage of `package-repos.Dockerfile:5` works only because it skips docker.
- Locality: `.github/workflows/test.yml:35` `Relevant changes`: the packaging path list is restated in README.md:1222 and deferred to by AGENTS.md:3, and the copies already differ ("Go code" against `'*.go' go.mod go.sum`).
- Navigation: `README.md:32` table of contents, `README.md:36` `## Install` and `README.md:118` `### What stdin renders`: a symptom such as a missing trailing newline now leads to a heading.
- Navigation: `README.md:1277` release paragraph: one paragraph traces the changelog heading, `publish_release.py`, the tag output and `packages.yml`, with the secrets and re-packaging.

Maintainability senior, decided by:

- Locality: `release-tooling/goreleaser.yaml:6` `dist` and `release-tooling/goreleaser.yaml:61` `dockers_v2.dockerfile`: `../dist` and `../tooling/...` resolve only under the two sibling checkouts `packages.yml:42` sets up, and the header explains `dist` but not `../tooling/`.
- Locality: `.github/workflows/packages.yml:72` `Channels holding one version` and `release-tooling/goreleaser.yaml:60` `latest` tag template: OLDER_TAG is set in one file and read through `isEnvSet` in another, and the same step handles an older tag a second way through `--skip`.
- Locality: `Dockerfile:18` `packaging`, `release-tooling/package-repos.Dockerfile:5` `snapshot` and `.github/workflows/packages.yml:82` `goreleaser-action`: one GoReleaser version pinned in three files, and `packages.yml:56` reads the Go version by `sed` over the Dockerfile's ARG line.
- Locality: `.github/workflows/test.yml:35` `Relevant changes`: the packaging path list is written here and again as prose in README.md:1222, which AGENTS.md:3 points the gate at.
- Navigation: `README.md:32` table of contents and `README.md:118` `### What stdin renders`: the stdin byte rules have a heading to jump to.
- Navigation: `README.md:1171` `## Layout` and `README.md:1277` `## Development`: a cold reader goes from a wrong apt version to `packages.yml`, `package-repos.Dockerfile` and `check-deb`.

Junior A, decided by:

- Navigation: `README.md:32` table of contents and `README.md:36` `## Install`: a reader with a symptom jumps to a named section, and the stdin byte rules sit under `README.md:118` `### What stdin renders`.
- Navigation, Self-sufficiency: `README.md:1171` `release-tooling/` in Layout and `README.md:1222` `## Development`: a failed install or Packages run leads to `release-tooling/package-repos.Dockerfile`, and each secret is named.
- Locality: `.github/workflows/packages.yml:72` `Channels holding one version` and `release-tooling/goreleaser.yaml:60` `isEnvSet "OLDER_TAG"`: whether the image gets `latest` is decided by a variable set in another file, with no note where it is read.
- Locality: `Dockerfile:18` `packaging`, `release-tooling/package-repos.Dockerfile:5` `snapshot` and `.github/workflows/packages.yml:82` `version`: one GoReleaser version pinned in three places, tied only by renovate.json.
- Self-sufficiency: `release-tooling/goreleaser.yaml:1`, `.github/workflows/packages.yml:48` and `release-tooling/package-repos.Dockerfile:1`: each non-obvious choice is explained where it is made.
- Self-sufficiency: `doc.go:5` package doc: the namespace sentence now has its subject.

## 2026-10-14T09:22:16Z, PR #222 at e492511, against c3cc251

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-14T11:34:34Z, PR #225 at f734c9e, against f90b4d6

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | same | same | same |
| Inherited architect | same | worse | same | same | worse |
| Maintainability senior | same | worse | same | worse | worse |
| Junior A | same | worse | same | worse | worse |

Mid A, decided by:

- Locality: `Dockerfile:26` `vulnerability-filing`: the filer's test reads the real todo.md inside docker build ., so a plan edit can turn the Go gate red
- Locality: `vulnerability-filing/file_vulnerabilities.py:24` `ROW`: the filer splices todo.md as text, assuming its row shape, the Next ID marker and the Details heading
- Locality: `.github/workflows/test.yml:106` `release`: a release now also waits on vuln.go.dev's answer, visible only in a needs list

Inherited architect, decided by:

- Locality, Overall: `Dockerfile:26` `vulnerability-filing`: the plan is copied into the merge-gate build, so a planning edit that changes the table's shape fails docker build .
- Locality: `vulnerability-filing/file_vulnerabilities.py:101` `file_items`: edits todo.md as text; a missing Next ID crashes with exit 1, which reads like a vulnerability already planned
- Locality: `.github/workflows/test.yml:69` `vulnerabilities`: whether a release happens depends on vuln.go.dev that day, and every PR shows a red job while one is known

Maintainability senior, decided by:

- Locality: `vulnerability-filing/file_vulnerabilities.py:101` `file_items`: finds the release by an unnamed regex group, never checks the Next ID match, and splices Details by a replace that does nothing when the heading is missing
- Locality, Self-sufficiency: `.github/workflows/test.yml:83` `File each one in todo.md`: one exit code holds the release and triggers the proposal, which must tell filed, planned and crashed apart by a git diff
- Locality: `Dockerfile:26` `vulnerability-filing`: copies todo.md into the gate build, and the path filter lists todo.md
- Self-sufficiency: `README.md:1213` "Development": says the build runs the release tooling's tests and leaves out the filer's
- Self-sufficiency: `vulnerability-filing/file_vulnerabilities.py:2` module docstring: builds on rank_todo.py, which the repository does not hold

Junior A, decided by:

- Locality, Overall: `Dockerfile:26` `vulnerability-filing`: the Go merge gate copies todo.md and files an item into it, so editing the plan can turn the build red
- Locality: `.github/workflows/test.yml:106` `release`: whether a push to main releases also depends on vuln.go.dev's answer that day
- Self-sufficiency: `vulnerability-filing/file_vulnerabilities.py:24` `ROW`: the regex and the Details splice encode todo.md's layout, which the script never shows
- Self-sufficiency: `README.md:1213` "Development": the gate paragraph leaves out the filer's tests while AGENTS.md now defers to it

## 2026-10-15T13:25:45Z, PR #227 at ee4ce67, against 5aca8a1

Ruling: worse

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | worse | worse | worse | same | worse |
| Inherited architect | worse | worse | worse | worse | worse |

Mid A, decided by:

- Navigation, Locality: `release-tooling/patch_release.py:36` `FILER`: the release script calls the filer by a relative path, and the release-tooling test stage copies that folder and moved to the full image with no word why
- Navigation: `README.md:1176` "Layout": says vulnerability-filing moves fixed items into CHANGELOG.md, while its docstring says patch_release.py owns it
- Navigation, Shape: `.github/workflows/packages.yml:69` "Channels holding one version": the latest release now comes from a script named for cutting patches
- Locality: `release-tooling/patch_release.py:187` `cut`: assumes main's checkout as working directory while building another tree
- Locality: `.github/workflows/test.yml:129` `plan`: release and planning are four jobs tied by outputs, an artifact and compound conditions
- Shape: `release-tooling/patch_release.py:177` `fixes`: None means still vulnerable, an empty string fixed nothing
- Shape, Locality: `vulnerability-filing/file_vulnerabilities.py:119` `release_for`: a three-way release rule, and resolve runs only on a clean report

Inherited architect, decided by:

- Locality, Self-sufficiency: `vulnerability-filing/file_vulnerabilities.py:228` `plan`: resolve runs only on a clean report, correct only because test.yml runs plan after release
- Locality, Shape: `release-tooling/patch_release.py:187` `cut`: parameters shadow module functions, a local plan shadows the module's, and it leans on main's checkout and leaves worktrees
- Navigation: `README.md:1176` "Layout": misattributes CHANGELOG.md writing; CHANGELOG.md now has three writers
- Navigation, Shape: `.github/workflows/packages.yml:69` "Channels holding one version": a latest-release query in a script named for patches
- Locality: `Dockerfile:22` `release-tooling`: the test stage copies vulnerability-filing for a relative-path subprocess
- Shape, Self-sufficiency: `vulnerability-filing/file_vulnerabilities.py:119` `release_for`: the release rule lives in one long module docstring sentence
- Locality: `.github/workflows/test.yml:93` `release`: job conditions combine five states
- Locality: `release-tooling/publish_release.py:44` `publish`: the tag check now precedes the published check, unexplained

## 2026-10-21T08:20:58Z, PR #238 at f743f3d, against e134ac5

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | same | better | better | better |
| Inherited architect | better | better | better | better | better |

Mid A, decided by:

- Shape: `release-tooling/patch_release.py:204` `cut`: its injected effects are `scanner`, `shipper` and `publisher`, so they no longer shadow the module's `scan`, `ship` and `publish`; the local `patch` no longer shadows `plan`.
- Shape, Self-sufficiency: `release-tooling/patch_release.py:190` `fixed_by_new_go`: returns the named sentinel `STILL_VULNERABLE` where e134ac5 returned a bare `None` that only the docstring explained.
- Navigation: `README.md:1305` "Layout" (`vulnerability-filing/`): says the filer names the fixed items and `release-tooling/` writes `CHANGELOG.md`, where e134ac5 said the filer moved them into it.
- Self-sufficiency: `release-tooling/patch_release.py:159` `carry_workflows`: its docstring says why `.github/` is copied from main, which the code alone did not show.

Inherited architect, decided by:

- Locality, Shape: `release-tooling/patch_release.py:204` `cut`: the parameters and locals that shadowed module functions now have their own names, so each name means one thing.
- Shape, Self-sufficiency: `release-tooling/patch_release.py:190` `fixed_by_new_go`: the sentinel `STILL_VULNERABLE` names what a bare `None` used to mean.
- Navigation, Shape: `README.md:1305` "Layout" (`vulnerability-filing/`): the line points at `release-tooling/` for `CHANGELOG.md`, where e134ac5 sent a reader to the wrong file.
- Self-sufficiency: `release-tooling/patch_release.py:159` `carry_workflows`: its docstring says why it copies main's workflows, which e134ac5 did not show.
- Self-sufficiency: `vulnerability-filing/file_vulnerabilities.py:33` `ROW`: reads harder, since the regex skips two unnamed cells that point at `todo.md`'s layout; outweighed by the places above.

## 2026-10-21T09:18:03Z, PR #239 at 5b98b19, against 798acf8

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-21T09:37:39Z, PR #240 at 1fe0c57, against b1537e7

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-21T15:03:54Z, PR #241 at 8758383, against 5154ed3

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-22T00:31:47Z, PR #247 at 4d4554b, against e6a787e

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | same | same | same |
| Inherited architect | same | worse | same | same | same |

Mid A, decided by:

- Locality: `data/sv_SE/organisationsnummer.json:5` `vat`: reads `{h}` and `{t}`, bound only in the category's `format`, and `data/sv_SE/vat.json:1` reaches it from outside, so a rename in `format` silently changes what `vat` renders
- Locality: `data/sv_SE/phone-area.tsv:2` `subscriber`: each row holds a reference into `phone-subscriber`, so that category's field names are load-bearing strings inside a 264-row table
- Locality: `data/sv_SE/company.json:29` `ab.written`: reads its form's name by the full path `{.company.ab.name}`, and only the changelog says that this draws a fresh name

Inherited architect, decided by:

- Locality: `data/sv_SE/company.json:28` `ab.written`: reads the absolute path `{.company.ab.name}`, so a bound company's form name is a fresh draw, and only the changelog's Breaking entry explains it
- Locality: `data/sv_SE/organisationsnummer.json:5` `vat`: reads names bound only in the top-level `format`, reached from `data/sv_SE/vat.json:1` from outside

## 2026-10-22T10:59:32Z, PR #249 at e21e6d6, against 67a8a92

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | worse | same | same |

Inherited architect, decided by:

- Shape: `internal/builtinfunc/builtins.go:27` `builtins`: `fnr` is filed as a derivation but also draws its second individ digit and reads an operand, so the registry comment's two shapes, derivation and sample, no longer cover every entry
- Shape: `data/nb_NO/birth-day.json:1` `nb_NO.birth-day`: a public category named for a birthday that renders `0181`, the month plus 80, which only the README says

## 2026-10-23T02:54:37Z, PR #256 at 11c6e9f, against cdaae5b

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | same | same | same | same |
| Inherited architect | same | same | same | same | same |

## 2026-10-28T10:30:06Z, PR #263 at ff9c22d, against dcdb6c1

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | better | same | same | better |
| Inherited architect | better | better | same | same | better |

Mid A, decided by:

- Navigation, Overall: `release-tooling/changelog.py:1` module docstrings of `carried.py`, `changelog.py` and `go_release_notes.py`: at dcdb6c1 one `patch_release.py` held version parsing, changelog editing, git and docker; now each concern has a file whose first line says what it owns, so a wrong changelog section or a missing Breaking line leads to one file
- Locality, Overall: `release-tooling/patch_release.py:203` `cut`, with `Effects` at :53: takes the checkout as `root` and its effects as one named tuple, where at dcdb6c1 it took four unnamed callables and leaned on the working directory in `scan` and `carry_workflows`; the `worktree` context manager also removes the temporary checkout
- Locality: `Dockerfile:15` stage `lowest`: the lowest Go is an ARG checked against go.mod's `go` line, where at dcdb6c1 `1.22.12` sat in test.yml with nothing tying it to go.mod
- Locality: `release-tooling/patch_release.py:240` `unreleased_breaking`: `plan` still reads `Dockerfile` and goreleaser.yaml relative to the working directory while `cut` takes `root`, and `go_release_notes.systems` reads only the one-line `goos: [...]` form; partly offsets the gains above

Inherited architect, decided by:

- Locality: `release-tooling/patch_release.py:203` `cut`, with `Effects` at :53: the checkout is passed in as `root`, the effects tests replace are named fields, and the `worktree` context manager removes the release worktree; at dcdb6c1 `cut` took unnamed positional callables, leaned on the current directory and left its worktree behind
- Locality: `release-tooling/patch_release.py:152` `patch_tree`: raises a `ValueError` naming what it found, where dcdb6c1's `STILL_VULNERABLE` sentinel equal to `None` had every caller test it with `is`
- Navigation: `release-tooling/changelog.py:64` `reconcile`, beside `carried.py` and `go_release_notes.py`: a wrong changelog section, no patch cut, or a missing Breaking line each points to one file, where at dcdb6c1 all sat in one flat `patch_release.py`
- Locality: `Dockerfile:15` stage `lowest`: the lowest Go is an ARG checked against go.mod, where at dcdb6c1 it was a literal in test.yml nothing checked
- Locality: `release-tooling/patch_release.py:231` `plan` and :240 `unreleased_breaking`: the first place to rot; `cut` takes `root`, but `plan` reads files relative to the current directory and runs `git show` with no `cwd`, two conventions in one module
- Locality: `release-tooling/changelog.py:95` `record` and :74 in `reconcile`, with `go_release_notes.py:65` `breaking_lines`: deduplication and reconciliation match the exact text `breaking_lines` produces, so rewording it in one file quietly breaks the other; a new coupling
- Navigation: `release-tooling/carried.py:8` `GORELEASER`: the regex reading packages.yml's GoReleaser version copies `renovate.json`'s `matchStrings`, so a change to that step has to know both
