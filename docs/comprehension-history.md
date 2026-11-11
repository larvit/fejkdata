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

## 2026-10-28T12:09:11Z, PR #265 at ece8375, against 1524d0e

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | worse | better | better | better |
| Inherited architect | same | same | better | same | better |

Mid A, decided by:

- Locality: `.github/workflows/packages.yml:7` `on.workflow_call.inputs.tag`: an empty tag turns the release workflow into a dry run, and a dozen distant expressions re-read that switch, so changing one step means holding "no tag publishes nothing" across the file, where 1524d0e had one path that always took a tag
- Shape, Self-sufficiency: `renovate.json:51` `packageRules`: the blanket `!github-actions` rule becomes rules naming exactly what waits, each description saying why, so a reader learns why an update did not merge itself without another document
- Locality, Shape: `release-tooling/goreleaser.yaml:60` `dockers_v2.tags`: one positive `LATEST`, passed in the goreleaser step's env and read by `skip_upload` too, replaces a negated `OLDER_TAG` set through `GITHUB_ENV` and a separate `--skip`; the step's new name says what it does
- Shape: `.github/workflows/test.yml:35` `changed()`: two copied diff blocks become one helper called twice, and the `gate` job says why it runs `always()`
- Locality: `.github/workflows/test.yml:54` `packages-dry-run`: the throwaway-key repository build has one path instead of a copy in the `test` job; partly offsets the dual-mode cost
- Self-sufficiency: `Dockerfile:37` stage `release-tooling`: copies packages.yml, renovate.json and the Dockerfile with nothing saying why; the reason is only in `carried_test.py`

Inherited architect, decided by:

- Shape: `release-tooling/goreleaser.yaml:60` `dockers_v2.tags`: `tags` and `skip_upload` read a positively named `LATEST` and each target says beside itself when it skips, where 1524d0e split the rule over two files as a negated `isEnvSet "OLDER_TAG"` and an outside `--skip`
- Shape: `.github/workflows/packages.yml:72` step `latest`: named for what it does and writes only `latest=true`, where "Channels holding one version" also wrote the skip flag and `OLDER_TAG`
- Shape: `.github/workflows/test.yml:35` `changed()`: two copied if/else blocks become two one-line outputs, and the `gate` job says why it runs on `always()`
- Locality: `.github/workflows/packages.yml:29` the `inputs.tag` branches: two modes and a dozen expressions switching on the tag, some inverted, make a reader hold both at every step; balanced by the package-repos build now having one caller
- Self-sufficiency: `Dockerfile:37` stage `release-tooling`: copies three files with no word why, which only `carried_test.py:15` explains; balanced by the new comments in packages.yml

## 2026-10-28T17:27:59Z, PR #269 at 5582bce, against 70ac017

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | same | better | same | better |
| Inherited architect | better | worse | better | same | same |
| Maintainability senior | better | same | better | better | better |
| Junior A | better | same | better | same | better |

Mid A, decided by:

- Navigation: `README.md:1486` "Development": names the `packaging` line of the `changes` step as the one list of files that trigger a dry run, where 70ac017 restated it in prose that could drift
- Navigation, Shape: `.github/workflows/test.yml:15` `changes`: change detection and the changelog check sit in a job named for them, and `test` only builds, where 70ac017 hid both inside `test`
- Shape: `.github/workflows/packages.yml:133` `dockerhub`: no per-step `if: inputs.tag` and no login step that explains a CI quirk; the dry run's limits come from its own read-only job
- Shape: `vulnerability-filing/file_vulnerabilities.py:115` `release_for`: `needs_minor` and its special case are gone, so it takes two inputs and holds one rule

Inherited architect, decided by:

- Locality: `.github/workflows/packages.yml:163` `repos` and `.github/workflows/packages.yml:204` `pages`: exactly one of `packages` and `packages-dry-run` always skips, so `repos` needs `!cancelled() && !failure()` and `pages` must repeat `needs.repos.result == 'success'`, with no word why; the next job below `repos` without `!cancelled()` skips silently on every release
- Locality: `.github/workflows/packages.yml:42` `packages`: the dry-run job's steps are an anchor back here, so each job carries the other mode's `inputs.tag` branches and a step edit for one mode changes both
- Locality, Shape: `.github/workflows/packages.yml:133` `dockerhub`: every per-step `if: inputs.tag` is gone, so the job reads straight through as release-only
- Navigation, Locality: `.github/workflows/test.yml:15` `changes` and `.github/workflows/test.yml:133` `release`: change detection is a named job, and `release` reads `needs.gate.result`, so whether main passed is decided in one place
- Navigation: `README.md:1486` "Development": points to the one list of packaging files instead of a hand-written copy
- Shape: `vulnerability-filing/file_vulnerabilities.py:115` `release_for`: the `needs_minor` property and the `module` parameter are gone

Maintainability senior, decided by:

- Navigation, Shape: `.github/workflows/test.yml:15` `changes`: change detection and the changelog check sit in a job of their own, `test` only builds, and the gate's step names each job's result, so a red gate leads to the failing job
- Navigation, Self-sufficiency: `README.md:1486` "Development": names the `packaging` line of the `changes` step instead of a hand-kept list that had already left out `go.mod` and `go.sum`
- Shape, Locality: `vulnerability-filing/file_vulnerabilities.py:115` `release_for`: `needs_minor` and the third parameter are gone, so the release choice is one rule, worded as the file's docstring words it
- Self-sufficiency: `renovate.json:74` the cache Actions' packageRule: says at the rule why a cache update waits for the maintainer
- Locality: `.github/workflows/packages.yml:42` `packages`: the steps anchored into the dry run keep `if: inputs.tag` guards that are dead in this job, and `repos` at `:163` needs `!cancelled() && !failure()`, offset by `dockerhub`'s dropped per-step guards

Junior A, decided by:

- Navigation, Shape: `.github/workflows/test.yml:15` `changes`: a job named for detecting changes, which `test` and `packages-dry-run` both read
- Navigation: `README.md:1486` "Development": points to the one list of packaging files
- Shape: `vulnerability-filing/file_vulnerabilities.py:115` `release_for`: one rule, two inputs
- Locality: `.github/workflows/packages.yml:42` `packages`: the anchor shares steps across both modes, which helps and costs about equally

## 2026-10-28T19:03:28Z, PR #270 at bea5b7c, against e192e2b

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | better | better | better | better |
| Inherited architect | better | better | better | better | better |

Mid A, decided by:

- Shape, Navigation: `Dockerfile:3` `latest`: named for what it is, the build on the latest Go, and for the CI target that runs it, where `portable` hinted at something it doesn't do
- Navigation: `.github/workflows/test.yml:52` `test`: each matrix leg is named for the Dockerfile target it builds, where one `vet + fmt + tests` job didn't say which half broke
- Locality, Self-sufficiency: `Dockerfile:52` `lowest-and-tooling`: each check sits in the stage it belongs to, and the final stage's comment says where a new check goes
- Self-sufficiency, Locality: `.github/workflows/packages.yml:166` `repos`: the comment says why its `if:` checks its needs' results itself, and the one at `:69` says who saves and who restores setup-go's cache
- Locality: `.github/workflows/test.yml:72` `go-caches`: the key names the target, so the two legs can't overwrite each other's cache
- Self-sufficiency: `README.md:1453` "Development": names both targets and what each runs, in the Dockerfile's and the workflow's words

Inherited architect, decided by:

- Navigation, Locality: `Dockerfile:52` `lowest-and-tooling`: a failed job's name leads straight to a stage of that name, and the format check no longer runs at the end of an unnamed final stage
- Locality, Shape: `data_test.go:757` `geoPlaceSet`: the typed `has` replaces a joined `street|code|locality|region` key that two callers rebuilt by hand
- Shape: `Dockerfile:3` `latest`: the latest-Go stage is named as the counterpart of `lowest`, matching the README, the matrix and `release-tooling/fixtures.py:4`
- Self-sufficiency: `.github/workflows/packages.yml:166` `repos`: the comment explains GitHub's skip rule, which a reader of the base had to know from GitHub's docs
- Self-sufficiency: `README.md:1453` "Development": names the two targets and what each runs, the GoReleaser config check included
- Self-sufficiency: `share_draws_race_test.go:5` `shareDraws`: a small cost; nothing says why the race build draws fewer, though `assertShare` scales its bounds with the draws

## 2026-10-28T19:10:18Z, PR #270 at bea5b7c

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7 | 5.5 | 6.5 | 6 | 6 |
| Mid A | 7 | 5.5 | 6.5 | 6 | 6 |
| Maintainability senior | 7 | 5.5 | 6.5 | 6 | 6 |
| Inherited architect | 7.5 | 5.5 | 6 | 7.5 | 6.5 |
| Mean | 7.13 | 5.50 | 6.38 | 6.38 | 6.13 |

Junior A, hardest first:

- `pick.go:9` `pickKey`: levels, addressed keys and moving a key under the template's key took three passes, with a heap-escape rule hidden in readUnder's index arithmetic
- `path.go:190` `drawSteps`: climbed and drawMemo.stepDownPins decide whether a step down draws afresh, rebuilt only through rows.Pins.Above
- `namefence.go:76` `refuseTwiceDrawnIn`: which template shape it refuses stays unclear without an example
- `datatype.go:230` `columnReadOf`: readsColumn couples four functions in three files
- `table.go:253` `table.route`: readsRow, pinnedBelow and fromRow combine into draw

Mid A, hardest first:

- `pick.go:179` `readName`: frames, base, pick, pickAt, the memo fields and the pins held at once, and readUnder changes the levels it is given
- `render.go:146` `readField`: three read paths, the fresh one with its own pins and frame
- `path.go:190` `drawSteps`: climbed and stepDownPins decide when a step down draws again
- `table.go:253` `route`: four booleans into draw and next, proven only with a truth table
- `arm.go:90` `compileArm`: a named read takes levelKeys from 0 and a fresh read from 1, with no word why

Maintainability senior, hardest first:

- `pick.go:28` `pickKey`: levels against addressed keys live in four files: pick.go, arm.go, resolve.go and namefence.go
- `namefence.go:76` `refuseTwiceDrawnIn`: the load-time proof that picks stay consistent, unprovable by reading
- `path.go:190` `drawSteps`: climbed, the step-down memo and route's flags decide together whether a row is drawn
- `datatype.go:230` `columnReadOf`: a four-function change-together set held only by a comment
- `env.go:71` `enter`: which frames a nested render sees, with comments on heap escape rather than visibility
- `path.go:145` `walkEvery`: a slice-capacity trick forces copy-on-append, with nothing saying why

Inherited architect, would restructure first:

- `pick.go:179` `readName`: pick, pin, memo and frame are four views of one fixed draw, held at once
- `node.go:65` `template`: filled in four phases by four files, with readsColumn's change-together set
- `render.go:56` `render`: a switch that renderEdges and contained in graph.go must mirror
- `format.go:1` `compileFormat`: the root format.go checks and compiles, it doesn't format
- `.github/workflows/test.yml:136` `release`: the Tests workflow also releases and plans vulnerability fixes
- `cli/cli.go:508` `isPath`: copies the path grammar

## 2026-10-28T19:15:31Z, PR #270 rerun at bea5b7c

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7 | 5 | 6 | 6 | 6 |
| Mid A | 7 | 5.5 | 6.5 | 6.5 | 6 |
| Maintainability senior | 7 | 5.5 | 6 | 6 | 6 |
| Inherited architect | 7 | 5.5 | 6.5 | 7 | 6.5 |
| Mean | 7.00 | 5.38 | 6.25 | 6.38 | 6.13 |

Junior A, hardest first:

- `pick.go:70` `keptInPick`: pick-key levels, addressed keys, the memo fields and the frame stack held at once to say whether a fresh read is kept
- `name.go:61` `bindNames`: the scope condition in the gather closure carries a rule stated nowhere near it
- `path.go:190` `drawSteps`: climbed and stepDownPins, beside walkEvery's unmarked slice aliasing
- `namefence.go:76` `refuseTwiceDrawnIn`: depends on pick.go's addressed keys, in another file
- `table.go:253` `route`: readsRow, pinnedBelow and draw tabulated by hand

Mid A, hardest first:

- `pick.go:9` `pickKey`: levels and addressed keys explained in one dense comment, with readUnder building them by hand
- `pick.go:179` `readName`: one name, one pick across categories is carried by frames a memo shares, stated at no read site
- `namefence.go:76` `refuseTwiceDrawnIn`: two walks mixed, and which programs it refuses needs the README
- `path.go:190` `drawSteps`: `..`, climbed, stepDownPins and pinnedBelow across three files and rows/pins.go
- `render.go:146` `readField`: three paths, and clearing env.pick only makes sense after pickKey

Maintainability senior, hardest first:

- `pick.go:79` `readUnder`: whether a read is kept depends on addressed keys gathered at load from other templates
- `namefence.go:346` `refuseTwiceDrawnIn`: recurses with refuseTwiceDrawn, and no sentence states the invariant they enforce
- `render.go:146` `readField`: why a fresh read forgets the pick lives in another file
- `datatype.go:230` `columnReadOf`: three-index slice appends with no stated reason, feeding four coupled readers
- `node.go:65` `template`: filled in four stages by four files, with no word on which field is valid when

Inherited architect, would restructure first:

- `pick.go:70` `keptInPick`: keeping a name's draw spans six root files plus internal/rows/pins.go
- `render.go:56` `render`: one of six node-kind switches a new kind must touch, kept in step with renderEdges by a comment
- `datatype.go:230` `columnReadOf`: column-read resolution filed under a type name
- `calc.go:51` `checkOperands`: the root calc.go only checks, sharing its stem with two other calc files
- `AGENTS.md:3` "Rules": the compose-service list omits cover, bench, dev, tidy and data-import

## 2026-11-04T20:15:28Z, PR #295 at ba315b6

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7.5 | 6 | 6.5 | 6.5 | 6.5 |
| Mid A | 7 | 5.5 | 6.5 | 6 | 6 |
| Maintainability senior | 7 | 5.5 | 6.5 | 6 | 6 |
| Inherited architect | 7 | 6 | 6 | 7 | 6.5 |
| Mean | 7.13 | 5.75 | 6.38 | 6.38 | 6.25 |

Junior A, hardest first:

- `namefence.go:280` `refuseFreshDraw`: a worklist, a seen set, a selector prepend and an early return on pointer identity held at once, and the result looks order-dependent
- `namefence.go:150` `readTwiceApart`: its last argument only gets its meaning one function down, in drawsAlong
- `namefence.go:422` `readsFieldApart`: the `both` condition stayed opaque, and freshFieldReads reads like a getter but fills a cache
- `pick.go:77` `landsInPick`: fresh and named keys match only through a levelKeys offset whose two call sites don't say why they differ
- `pick.go:196` `readName`: frames, base, pick, pickAt and four memo maps across three files, traced by hand
- `path.go:192` `drawSteps`: two silent modes, nil memo and nil pins, plus a climbed flag
- `columnread.go:24` `columnReadOf`: "this column is really a read" reaches four files, and a synthetic arm mirrors compileArm
- `name.go:61` `bindNames`: which node owns a scope is decided in a closure condition

Mid A, hardest first:

- `namefence.go:280` `refuseFreshDraw`: each continue is a rule whose reason lives in the README, and the selector prepend stayed opaque
- `namefence.go:456` `rendersField`: returns false mid-loop while reads may be true, so the answer looks order-dependent
- `namefence.go:422` `readsFieldApart`: assumes levels and tail line up by an index relation stated nowhere near it
- `pick.go:77` `landsInPick`: load and render share a key scheme that everything in namefence.go needs held in mind
- `pick.go:196` `readName`: frames change on a by-value env over a shared stack, and the lifecycle has no single statement
- `table.go:252` `route`: six fields combine conditionally, and tail[2:] is safe only through a proof elsewhere
- `columnread.go:24` `columnReadOf`: three spellings of one read, with dense slice building
- `cli/cli.go:569` `hasEmptySegment`: the stepUp toggle is a small puzzle its comment doesn't explain

Maintainability senior, hardest first:

- `namefence.go:280` `refuseFreshDraw`: a queue, a seen set and a drawsAlong guard live at once, with examples and no invariant
- `namefence.go:400` `refuseKeptAndFreshField`: the `both` condition can be traced but not justified, and level is shadowed
- `namefence.go:120` `refuseTwiceDrawnIn`: five sibling fences each define "could differ" their own way, with no map to the README's bullets
- `pick.go:77` `landsInPick`: a load fence and a render agree purely through key strings, and no type enforces that
- `path.go:192` `drawSteps`: how a path passes a table is split across routing, compiling and drawing
- `render.go:144` `readField`: a by-value env holds a pointer to a stack mutated out of sight
- `columnread.go:24` `columnReadOf`: three switch arms build start, category and tail from three shapes of read
- `resolve.go:96` `resolveTemplates`: each step fills a partially initialised struct, and the order is held by the pipeline, not by types

Inherited architect, would restructure first:

- `data.go:148` `pipeline.run`: the root's five unnamed strata show only here and in resolveTemplates, after about twelve files
- `namefence.go:23` `checkNameRead`: each refusal predicts what the memo will do, by hand, and each new read adds one
- `path.go:87` `compiledPath.walk`: one of five walkers that each encode how a path moves
- `render.go:54` `render`: one of six node-kind switches a new kind must reach, none failing at compile
- `node.go:65` `template`: fields filled phase by phase, so a reader must know what has run
- `record.go:323` `wholeLevels`: hand-copies what compileArm does
- `cli/cli.go:555` `isPath`: a copy of the grammar, since the CLI reads only the public API

## 2026-11-04T20:15:30Z, PR #295 rerun at ba315b6

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7.5 | 6 | 6.5 | 6.5 | 6.5 |
| Mid A | 7.5 | 6 | 7 | 7 | 6.5 |
| Maintainability senior | 7 | 5.5 | 6 | 6 | 6 |
| Inherited architect | 7 | 5.5 | 6 | 7 | 6.5 |
| Mean | 7.25 | 5.75 | 6.38 | 6.63 | 6.38 |

Junior A, hardest first:

- `namefence.go:280` `refuseFreshDraw`: levels, tail offset, leaf identity and arm identity held at once, and the selector prepend has no comment
- `namefence.go:422` `readsFieldApart`: triple nesting and the `both` expression stayed opaque, and level is shadowed
- `namefence.go:224` `sharedDraw`: a five-way decision in one loop, reached through a bool that stacks several disjuncts
- `pick.go:196` `readName`: load and render must agree on keys, with an allocation-driven quirk in its near-twin readUnder
- `env.go:72` `enter`: whether a frame opens depends on what an earlier read did
- `table.go:252` `route`: six fields with interacting bools, then translated into steps
- `cli/cli.go:555` `isPath`: a second copy of the path grammar, and nothing forces it to follow
- `columnread.go:24` `columnReadOf`: the targets argument exists only because b.target is not set yet

Mid A, hardest first:

- `namefence.go:280` `refuseFreshDraw`: keys, levels, tails and leaves held at once, and readAddressing takes six parameters
- `namefence.go:400` `refuseKeptAndFreshField`: the `both` condition depends on an indexing relation it doesn't state
- `namefence.go:23` `checkNameRead`: five fences in sequence, and nothing says how they divide the problem
- `pick.go:77` `landsInPick`: keys match only through levelKeys' two starts, tied together by prose alone
- `pick.go:196` `readName`: env.pick, pickAt, base, the frames and enteredFrames interact, found by tracing
- `path.go:119` `compileStep`: fromRow, pinnedBelow, up, nextLinked, climbed and at cross two files
- `name.go:61` `bindNames`: the gather closure's scope condition is dense for what the unit does
- `cli/cli.go:555` `isPath`: nothing visible stops it drifting from SplitPath

Maintainability senior, hardest first:

- `namefence.go:280` `refuseFreshDraw`: why each continue is safe stayed largely opaque, and the selector prepend has no explanation
- `namefence.go:400` `refuseKeptAndFreshField`: the `both` condition and the tail slice mix two coordinate systems
- `namefence.go:23` `checkNameRead`: five refusals with no statement of what they jointly protect, so the list's completeness can't be judged
- `pick.go:77` `landsInPick`: resolved only after several rereads with both levelKeys call sites open
- `path.go:192` `drawSteps`: several switch sites over step kinds, with flags and nils changing behaviour
- `render.go:144` `readField`: three read paths treat the pick and frames differently, and an escape-analysis rule can't be seen at call sites
- `columnread.go:24` `columnReadOf`: a four-way case where an empty category means "own record"
- `name.go:61` `bindNames`: real semantics in one scope condition, and a kind classification that depends on case order

Inherited architect, would restructure first:

- `namefence.go:23` `checkNameRead`: the fences statically re-model what render's memo does, each walk mirroring render its own way
- `env.go:33` `renderEnv`: who opens which frame takes three places held at once
- `path.go:192` `drawSteps`: its mode is an unnamed concept recovered from which arguments are nil
- `path.go:87` `compiledPath.walk`: a second walker in callerPathSteps is easy to leave unpatched
- `node.go:65` `template`: filled by stages whose order sits in pipeline.run
- `render.go:54` `render`: a child rendered here must also be listed in renderEdges, held by a comment and a panic
- `data/geo/SE/address.json:3` `street`: binds l inside the field its siblings read, correct but reading as a bug on the file the 3am question opens
- `internal/rows/rows.go:343` `Descendant`: ranges a map, sound only because linked tables are siblings with unique names, and nothing says so

## 2026-11-04T21:57:22Z, PR #296 at 3473360, against 6d1bafa

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | better | better | better | better |
| Inherited architect | better | better | better | better | better |

Mid A, decided by:

- Overall, Locality, Shape: `namefence.go:120` `walkRoutes`: one walk lists every read with its kind and place in the pick, where the older version had three walks with slightly different following rules; `route`'s fields valid only for some kinds and the `inPick` and `kept` state remain a cost
- Navigation, Self-sufficiency: `namefence.go:36` `checkNameRead`: its comment lists the two ways a name breaks its promise and maps each to the refuser that catches it
- Locality: `namefence.go:467` `freshFieldReads`: a plain function called at `resolve.go:115`, where a check used to fill `names.fresh` lazily
- Locality, Shape: `cli/cli.go:252` `recordFormat`: each format carries its own `write` and `empty`, where `writeRecords` branched on the format and `writeNoRecord` inferred its output
- Shape: `pick.go:40` `namedLevels`: this and `freshLevels` replace `levelKeys(path, from)` and its bare 0 or 1 with two truthful names
- Shape: `namefence.go:534` `readAddressing`: a method taking a `refReach` replaces a 7-argument function, and `refuseFreshDraw` shrinks into `refsReached` and `underSelector`
- Shape: `columnread.go:25` `columnReadOf`: splits into `readsOneArm` and `columnSource`, replacing one long function with a long guard
- Self-sufficiency: `pick.go:125` `namedPick`: its comment gives the pick's lifecycle, which the older version left the reader to reconstruct
- Self-sufficiency, Locality: `data.go:152` `pipeline.run`: its comment states the stage order and why it holds

Inherited architect, decided by:

- Navigation, Self-sufficiency, Overall: `namefence.go:36` `checkNameRead`: its comment lists the two ways a name breaks its promise and which refuser handles each, so a load error leads straight to the right fence
- Shape, Locality: `namefence.go:120` `walkRoutes`: one walk and one `routes` list feed every fence, where separate walkers each had their own rule; `route`'s partial fields, the `inPick` flag and the silent reset at `:141` are smaller costs than what it removes
- Locality: `resolve.go:115` `resolveTemplates`: `fresh` is computed up front by a free function, removing a hidden write that depended on call order
- Shape, Self-sufficiency: `pick.go:40` `namedLevels`: this and `freshLevels` replace `levelKeys(path, from)` with a bare 0 or 1, and each comment says why its levels start where they do
- Self-sufficiency: `data.go:152` `pipeline.run`: the new comment says why each stage must follow the one before
- Locality, Shape: `cli/cli.go:252` `recordFormat`: each format owns its writer and its zero-record output, so adding a format touches one table entry; `write` takes a `table` only the SQL writer uses
- Navigation, Self-sufficiency: `release-tooling/propose_plan.py:43` `propose`: about 30 lines of shell in the workflow are now a named script with a test file
- Shape: `columnread.go:37` `readsOneArm`: this and `columnSource` split the older `columnReadOf` into a test and a lookup whose names say which is which
- Shape, Locality: `namefence.go:41` `checkNameRead`: `names.pick(b)` and `refReach` replace a repeated struct literal, the 7-argument `readAddressing` and the nested loop in `refuseFreshDraw`

## 2026-11-04T22:20:35Z, PR #298 at 65f8f06

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 8 | 6 | 7 | 6.5 | 7 |
| Mid A | 7.5 | 6 | 7 | 7 | 6.5 |
| Maintainability senior | 7.5 | 5.5 | 6.5 | 7 | 6.5 |
| Inherited architect | 7 | 5.5 | 6 | 7 | 6.5 |
| Mean | 7.50 | 5.75 | 6.63 | 6.88 | 6.63 |

Junior A, hardest first:

- `namefence.go:349` `sharedDraw`: two parallel draw routes, keys and levels and the addressed set across five orderings; needs the README's Names section and the changelog to see what a fence protects, and a wrong change shows only in seeded output
- `namefence.go:120` `walkRoutes`: recursive walk with kept, fresh and apart state in a value struct; understood only in outline
- `pick.go:85` `landsInPick`: a named and a fresh read must make one key though their levels start at different nodes; the comment resolves it on a careful reread
- `render.go:144` `readField`: behaviour depends on frame-stack order, and `enter` and `renderFrame` act at a distance
- `path.go:192` `drawSteps`: pins swapped mid-walk after `..`, pins and memo nil by caller
- `table.go:244` `route`: `fromRow`, `readsRow`, `pinnedBelow` and six `tableRoute` fields, and two translations of a path
- `name.go:61` `bindNames`: closure `gather`'s scope condition decides where a name lives with no local explanation
- `columnread.go:45` `columnSource`: one fact has two sources by phase, `targets[b]` during load and `b.target` after

Mid A, hardest first:

- `namefence.go:406` `refuseFreshDraw`: several conditions at once, and the code doesn't show why the exclusion for a template not rendering `reached.ref` is safe
- `namefence.go:349` `sharedDraw`: predicates over key levels in an ordered case list; comments say what each case detects, not why the others are safe; also the unit least wanted to modify
- `namefence.go:490` `refuseKeptAndFreshField`: index arithmetic over `u.levels` and a long error hide a small predicate
- `namefence.go:120` `walkRoutes`: a load-time model of `readField`, kept in step by hand beyond `landsInPick`
- `pick.go:85` `landsInPick`: named and fresh keys match by where each level list starts; resolved after several rereads
- `path.go:192` `drawSteps`: `pins` swapped mid-loop, `climbed`, and nil `memo` and `levels` serve two modes
- `table.go:244` `route`: `selected`, `fromRow` and `pinnedBelow` combine, and `routeSteps`' `next != t` case means "left to a render's own draw" only by `route`'s comment
- `name.go:61` `bindNames`: a closure with an `item` flag and a terse scope condition, explained only in the README

Maintainability senior, hardest first:

- `namefence.go:406` `refuseFreshDraw`: a read is excluded for rendering `reached.ref`, a claim about render that takes simulating `readUnder` to check; also the unit least wanted to modify
- `namefence.go:490` `refuseKeptAndFreshField`: three nested loops around a nine-argument error, with `x` and `u` naming neither route
- `namefence.go:120` `walkRoutes`: a recursive closure deduped on three parts, with `readTwiceApart`'s offset stated only at `readsFieldApart`
- `pick.go:85` `landsInPick`: load and render agree through an off-by-one convention between string keys
- `env.go:72` `enter`: who opens a frame is spread over six units, and `base` silently hides frames
- `render.go:144` `readField`: three outcomes, and clearing `env.pick` takes a trace to believe
- `path.go:192` `drawSteps`: nil pins and nil memo are modes, and `climbed` changes the next step
- `record.go:313` `renderRecord`: a forged arm with no steps, safe only by the `boundWhole` condition

Inherited architect, would restructure first:

- `pick.go:85` `landsInPick`: a field read inside a pick is kept only where a read of the name addresses that key; the costliest correction to the map, with rereads across four files
- `namefence.go:120` `walkRoutes`: a hand-kept second model of `readField`, and with `scopeAround` and `neverNumeric` a node-kind switch that falls through silently
- `namefence.go:534` `readAddressing`: with `refuseFreshDraw` and `rendersField`, among the hardest places to read and first to rot
- `format.go:42` `checkTokens`: the file is named for formatting and mostly holds token checks
- `namefence.go:96` `route`: one name with `tableRoute`, and `pick`, `name` and `fresh` each carry more than one concept too
- `path.go:74` `compiledPath`: this and `callerPathSteps` walk one route twice
- `resolve.go:87` `targetOf`: `b.target` and `names.targets` hold one fact
- `cli/cli.go:559` `isPath`: copies `SplitPath`, marked and justified, and can still drift

## 2026-11-04T22:20:38Z, PR #298 rerun at 65f8f06

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7.5 | 5.5 | 6.5 | 7 | 6.5 |
| Mid A | 7.5 | 6 | 7 | 7 | 6.5 |
| Maintainability senior | 7.5 | 6 | 7 | 7.5 | 6.5 |
| Inherited architect | 7.5 | 6 | 6 | 7 | 6.5 |
| Mean | 7.50 | 5.88 | 6.63 | 7.13 | 6.50 |

Junior A, hardest first:

- `namefence.go:349` `sharedDraw`: five ordered cases with neither why each holds nor why they come in that order, and a failed `childNamed` counts as true in `drawsAlong`; also the unit least wanted to modify
- `namefence.go:534` `readAddressing`: the exclusion rests on one sentence and double negatives across `rendersField`
- `namefence.go:490` `refuseKeptAndFreshField`: turns on the index arithmetic of `u.levels[j]`, opaque without a worked example
- `namefence.go:120` `walkRoutes`: threads `at` and `inPick`, and the default `routeApart` kind has no visible assignment
- `pick.go:85` `landsInPick`: load and render agree through a correspondence nothing enforces at the point of use
- `render.go:144` `readField`: frames are pushed from several places, and the combined stack discipline is written down nowhere
- `path.go:192` `drawSteps`: `climbed`, swapped `pins`, and nil receivers the call site doesn't show
- `table.go:244` `route`: flags and `at++` shift later steps, and `route` means something else in `namefence.go`

Mid A, hardest first:

- `namefence.go:349` `sharedDraw`: pick keys, levels, addressed keys and the tail offset all at once, and the order of five exits is part of the meaning
- `pick.go:85` `landsInPick`: two key spaces line up only by convention, and `readUnder` mixes an allocation concern into the logic
- `env.go:72` `enter`: frames taken from a memo, a base hiding older frames and a pop mark passed back, across two files
- `path.go:192` `drawSteps`: what it does depends on which of `pins`, `memo` and `levels` are nil, and the combinations are listed nowhere
- `table.go:244` `route`: `selected`, `readsRow`, `pinnedBelow` and `fromRow` combine, and `routeSteps`' `at` must match `followStep`
- `namefence.go:120` `walkRoutes`: why a table clears the pick isn't stated at that case
- `record.go:290` `recordColumns`: a column read is manufactured to reuse the pick machinery
- `name.go:61` `bindNames`: the closure's captured state is longer to hold than its line count suggests
- `namefence.go:36` `checkNameRead`: the unit least wanted to modify, since a missed case produces no error, only quietly mismatched output

Maintainability senior, hardest first:

- `namefence.go:534` `readAddressing`: what resolved it was neither code nor comment but the changelog's Breaking entry
- `namefence.go:349` `sharedDraw`: five rules whose order matters, and `levels` sit one index off from `tail`
- `pick.go:85` `landsInPick`: two key origins, the steepest concept in the project; the worked example made it land
- `env.go:72` `enter`: whether `renderFrame` opens a frame depends on an earlier `enter`, and the frame stack's rules have no owner
- `namefence.go:120` `walkRoutes`: a table silently resets the walk, and `at`, `levels` and `by` hold only for some kinds
- `path.go:192` `drawSteps`: nil encodes modes, and `climbed` changes what the next step does
- `table.go:244` `route`: several flags at once, plus an `at` index rule to keep in mind
- `resolve.go:97` `resolveTemplates`: eight ordered stages, and `target` assigned last makes two ways to get one fact
- `namefence.go:406` `refuseFreshDraw`: the unit least wanted to modify, grown case by case with no single stated invariant

Inherited architect, would restructure first:

- `render.go:144` `readField`: one named read runs through `readName`, `enter`, `drawSteps` and `landsInPick`, the most rereading of the session
- `namefence.go:349` `sharedDraw`: with `refuseFreshDraw`, `readAddressing` and `rendersField`, fences reason about what render would do, so the semantics live twice
- `node.go:66` `template`: stages fill its fields in the order `resolveTemplates` sets, which the type doesn't show
- `render.go:53` `render`: a child rendered must also be listed in `renderEdges`, a mirror kept by comment
- `namefence.go:96` `route`: one name with `tableRoute`, the most misleading of several, with `format`, `arm` and three `calc.go` files
- `record.go:271` `settleRecords`: a load stage in a file that is also serializer, public API and render
- `cli/cli.go:559` `isPath`: copies the grammar because the CLI reads only the public API

## 2026-11-09T17:09:39Z, PR #327 at 7d2b3a5

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7.5 | 5.5 | 6.5 | 6.5 | 6.5 |
| Mid A | 7 | 5.5 | 6.5 | 7 | 6.5 |
| Maintainability senior | 7.5 | 5.5 | 6.5 | 6.5 | 6 |
| Inherited architect | 7.5 | 6 | 6 | 7.5 | 6.5 |
| Mean | 7.38 | 5.63 | 6.38 | 6.88 | 6.38 |

Junior A, hardest first:

- `namefence.go:349` `sharedDraw`: five ordered rules over pick-key levels, route kinds and the `resolvedNames` cache at once; the comments say what each refuses, not why that order
- `pick.go:87` `landsInPick`: fresh and named reads share one key space through `freshLevels` and `namedLevels`; clear only on a second read of its comment
- `namefence.go:490` `refuseKeptAndFreshField`: an error built from seven spellings, whose slots take tracing both callers
- `path.go:192` `drawSteps`: `climbed`, `memo.stepDownPins` and `table.route`'s six flags; resolved in part by the README's `..` sections
- `render.go:156` `readField`: why `env.pick` is cleared takes knowing how `enter` and the base marks interact, and `renderEnv` mixes a value with shared state
- `where.go:580` `drawIndex`: a weighted draw without replacement over sorted indexes, with a rounding correction, hard to check by eye
- `columnread.go:93` `columnSource`: three cases concatenate paths from three sources; the inline examples save it
- `format.go:80` `checkReads`: `own := len(names) - len(clause)` splits a list with no word on what lies on each side

Mid A, hardest first:

- `namefence.go:349` `sharedDraw`: pick-key levels, addressed keys and an ordered five-rule walk held at once, with the level reasoning defined in pick.go; also the unit least wanted to modify
- `pick.go:87` `landsInPick`: rests on how `namedLevels` and `freshLevels` align; the `pickKey` comment resolves it after rereads
- `where.go:580` `drawIndex`: a weighted draw without replacement over cumulative sums, with a rounding fallback, dense
- `render.go:156` `readField`: the comment on clearing `env.pick` doesn't say what breaks without it
- `path.go:145` `walkEvery`: a slice-capacity trick shares prefixes, commented but easy to break by tidying
- `table.go:250` `route`: `pinnedBelow` and `fromRow` decide a fresh draw, followed only with the README's `..` semantics
- `node.go:132` `position`: six values whose differences show only where they are used
- `columnread.go:77` `readsOneArm`: one boolean of about eight clauses

Maintainability senior, hardest first:

- `namefence.go:406` `refuseFreshDraw`: a breadth-first walk across categories that takes the whole pick-key model; the list it implements, the README's "Refused at `New`", sits far away and neither links the other
- `namefence.go:349` `sharedDraw`: an ordered walk of five cases where the order is the logic, and nothing shows why
- `namefence.go:490` `refuseKeptAndFreshField`: three nested loops, level-index arithmetic and a nine-argument error; not safe to change
- `pick.go:87` `landsInPick`: keys shift through `at.under(levels[0])`, and named and fresh levels start from different origins
- `where.go:635` `proveEveryDrawPasses`: the low end takes a name's upper bound with no word that it is deliberate, and `lengthEnd`'s `read == -1` means no name while its zero value reads name 0
- `path.go:192` `drawSteps`: `climbed` and `memo.stepDownPins` say what, never why
- `table.go:250` `route`: six flags, with `draw` derived from `readsRow` and `pinnedBelow`
- `format.go:80` `checkReads`: right only because `tokenReads` appends the clause's names last, an order no type enforces

Inherited architect, would restructure first:

- `render.go:53` `render`: a node kind's children are listed again in `renderEdges`, `paths`, `valueProof.prove`, `compileItem` and `drawSteps`, held in step by a comment and a render-time panic; rots first
- `data-import/geo-se.py:239` `streets`: the 3am answer, a road segment's postort from the nearest code centroid and a tätort override, is found only after ruling out the engine
- `internal/rows/rows.go:418` `Table.DrawIn`: `internal/rows` holds pinning, the parent-first draw and the link proof, not just a TSV, and its generic owner `O` breaks an import cycle nothing names
- `namefence.go:96` `route`: shares its name with `tableRoute`, as `template` and three `calc.go` files each cover two or three concepts
- `sampledraw.go:14` `sampleDraw`: "sample", the README's "alphabet calls" and "a builtin that draws each character from an alphabet" name one concept
- `data-import/geo-se.py:254` `main`: SE overrides `geo.COLUMNS`, so a street's weight counts segments where the schema says addresses, and US `municipality` is a county
- `env.go:33` `renderEnv`: a harmless-looking refactor moves every render's env to the heap, said only in comments and the allocation gate
- `README.md:2000` "Layout": the root's 29 source files fall in seven clusters that exist only as this section's line order, with `where.go` and `cli/cli.go` near 800 lines

## 2026-11-09T17:13:48Z, PR #327 rerun at 7d2b3a5

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Junior A | 7 | 5 | 5.5 | 6 | 5.5 |
| Mid A | 7 | 5.5 | 6.5 | 6 | 6 |
| Maintainability senior | 7 | 5.5 | 6 | 6 | 6 |
| Inherited architect | 7.5 | 6 | 6 | 7.5 | 6.5 |
| Mean | 7.13 | 5.50 | 6.00 | 6.38 | 6.00 |

Junior A, hardest first:

- `namefence.go:349` `sharedDraw`: a five-way ordered rule list whose order carries meaning no comment states, with `readsFieldApart`'s index arithmetic beside it; resolved in part by doc.go's vocabulary
- `pick.go:87` `landsInPick`: levels, keys moved under `at`, addressed keys and `namedLevels` against `freshLevels` held at once; the comment takes more than one read
- `pick.go:213` `readName`: the memo's entered frames and `renderEnv.base` decide which frame a read sees, and popping by mark is order kept by hand
- `render.go:156` `readField`: three branches, the fresh one split again on a reference head; why `env.pick` is cleared makes sense only after `landsInPick`
- `path.go:192` `drawSteps`: `climbed` and `memo.stepDownPins` carry state across steps, and nil arguments at its callers choose a mode no name states
- `table.go:250` `route`: four booleans combine into a routing decision whose combinations a reader enumerates alone
- `where.go:580` `drawIndex`: a weighted draw skipping drawn indexes with a rounding correction, and `-1` meaning none in `lengthEnd` and `fallback`
- `node.go:66` `template`: fields filled in stages by four functions, so which hold depends on how far loading has got; the stage comments mostly resolve it

Mid A, hardest first:

- `namefence.go:349` `sharedDraw`: a five-way switch whose cases depend on their order, under `checkNameRead`'s five refusals; why a template is refused takes the README's Names list; also the unit least wanted to modify
- `where.go:580` `drawIndex`: skips drawn indexes through cumulative odds, and `undrawnNear`'s rounding correction never says what goes wrong without it
- `pick.go:105` `readUnder`: whether a read reuses a pick spreads over `drawMemo`, `enteredFrame`, `stepDownPins` and the frame stack, with `renderEnv`'s `base` a hidden coupling
- `path.go:192` `drawSteps`: `climbed` and `stepDownPins` are hard to follow, and `render.go:44` passes two bare nils
- `table.go:250` `route`: `readsRow`, `pinnedBelow` and `fromRow` combine into routes worked out on paper
- `format.go:84` `checkReads`: `own := len(names) - len(clause)` relies on the layout `tokenReads` builds in another function
- `index.go:136` `load`: a queue that grows while consumed, rolled back by `putBack` on error; its comment resolves it on a second read

Maintainability senior, hardest first:

- `namefence.go:406` `refuseFreshDraw`: four helpers decide together whether a fresh read shares a draw, and no comment says how they combine; rebuilt from the README's Names section
- `namefence.go:349` `sharedDraw`: a five-case switch whose order is the meaning, with nothing saying which order is required or why
- `pick.go:87` `landsInPick`: a named read keys its levels from `""` and a fresh one from its head moved under `at`; the most expensive comment to absorb
- `where.go:635` `proveEveryDrawPasses`: `lo` takes the name's upper bound and `hi` its lower with no word why, and `-1` is a sentinel in `lengthEnd` and `fallback`
- `resolve.go:108` `resolveTemplates`: about twelve passes in a fixed order, with the facts on which needs which only in callee comments
- `path.go:192` `drawSteps`: two modes chosen by nil pointers, and `stepColumn` and `stepChild` mean nothing until table.go explains them
- `table.go:250` `route`: six fields and four flags combine, and the `pinnedBelow` rule takes two reads
- `format.go:80` `checkReads`: aligns the clause's names after the token's own reads, right only while `tokenReads` appends them in that order

Inherited architect, would restructure first:

- `README.md:2000` "Layout": the root's 29 engine files and about 45 tests sit in one flat directory, grouped only by this section's lines; the worst level
- `namefence.go:36` `checkNameRead`: five refusal routes over one walk, each new read kind adding one that must agree with `landsInPick`; rots first
- `data-import/geo.py:157` `address_json`: generates DK's and NO's address templates while SE's and US's are hand-written copies, recorded nowhere and found only by grep
- `data-import/geo-se.py:239` `streets`: the 3am bug's second suspect, in a second language, beside the engine's bound locality
- `namefence.go:96` `route`: shares its name with `table.route`, as `key`, `name`, `table`, `column` and `paths` each cover two concepts
- `sampledraw.go:14` `sampleDraw`: its name doesn't say it is the where clause's draw for alphabet builtins, beside `valueDraw` in where.go
- `internal/builtinfunc/builtins.go:32` `builtins`: confusable with `builtinfunc.go`, and the root, grammar and builtinfunc share stems such as `calc` with no stated rule for the split

## 2026-11-09T18:36:41Z, PR #328 at 35cf848, against 4cd2f6d

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | better | better | better | same | better |
| Inherited architect | better | better | better | better | better |

Mid A, decided by:

- Locality, Shape: `arm.go:67` `op`: the where clause's names have their own `clause` field instead of overwriting `operands`, and `reads()` lists every arm in one place
- Locality: `where.go:32` `lengthEnd`: an `*arm` that is nil for none replaces an index into a separate `reads` slice with -1 for none
- Locality, Shape: `resolve.go:124` `resolveTemplates`: draws pass between named stages as `bindingDraws`, and only the commit writes the bindings
- Navigation, Shape: `arm.go:93` `tokenDraw`: one draw behind one interface replaces three nullable fields that `expand` checked in turn
- Shape: `adddays.go:13` `prepareAddDays`: building the call and proving the day count are two functions, so `proveAddDays` no longer hides a write
- Navigation, Shape: `pathvalues.go:32` `valueWalk`: the path walk and length index left a 790-line `where.go`, and `alphabetDraw` names what it draws from
- Locality: `valueproof.go:16` `valueProof.draws`: a proof must be handed the draws still being resolved, a small cost the comment names

Inherited architect, decided by:

- Locality, Shape: `arm.go:67` `op`: `clause` is its own field, one `draw tokenDraw` replaces three per-kind fields, and `reads()` lists every arm a render reads
- Locality, Self-sufficiency: `where.go:32` `lengthEnd`: each end holds an `*arm`, nil where nothing is read, so no `reads` slice is threaded through every call
- Locality, Navigation: `resolve.go:124` `resolveTemplates`: three named stages pass draws explicitly, and only the commit writes the bindings; `bindingDraws.of` still falls back to `b.values`
- Shape: `adddays.go:13` `prepareAddDays`: preparing the call and proving the day count are separate units, and the panic stub is gone
- Navigation, Shape: `pathvalues.go:32` `valueWalk`: the walk and `lengthIndex` live in their own file, leaving the clause in `where.go`
- Shape: `alphabetdraw.go:14` `alphabetDraw`: the name says it draws from an alphabet, and `mayFallBack` sits on the type
- Self-sufficiency: `format.go:79` `checkReads`: `tokenReads` returns own and clause reads apart, so no index arithmetic splits one list
- Locality: `arm.go:103` `drawToken`: a type switch mirrors the `tokenDraw` interface, so a new draw type is added in two places

## 2026-11-10T19:49:53Z, PR #339 at 527e3ae, against 7a33428

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | better | better | better |
| Inherited architect | same | better | better | better | better |

Mid A, decided by:

- Locality, Shape: `arm.go:83` `newOp`: one function builds each op's call and draw, where three passes wrote the same op in place before
- Locality: `resolve.go:169` `buildOps`: stores each op whole and records a binding's draw in one `if`, not inside a switch arm
- Shape: `arm.go:37` `splitArm`: a read's text is called a spelling, not a name
- Self-sufficiency: `doc.go:27` "spelling": defines the term the renames use

Inherited architect, decided by:

- Locality, Shape: `arm.go:83` `newOp`: each op is built whole in one place; the one ordering need sits in its comment
- Locality, Shape: `resolve.go:139` `compileTemplates`: loses a stage, `prepareAddDays` and `buildDraws` becoming one `buildOps`
- Shape, Self-sufficiency: `node.go:93` `templateRefs.bySpelling`: the raw string is a spelling, and `doc.go:27` defines it

## 2026-11-10T20:42:56Z, PR #341 at 800aba4, against 33e465b

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | better | same | better |
| Inherited architect | same | better | better | better | better |

Mid A, decided by:

- Locality, Shape: `path.go:205` `drawSteps`: the caller names the kind of read through drawMode, where nil pins and nil memo meant it before
- Shape, Locality: `arm.go:166` `compileArm`: it builds each arm whole in one literal, where splitArm returned an arm with no steps
- Shape: `table.go:237` `tableRoute`: one then pathPos replaces next, rest and nextLinked, so compileStep assigns it whole

Inherited architect, decided by:

- Locality, Shape: `arm.go:166` `compileArm`: splitRead returns only text, head and tail, and an arm exists only once complete
- Locality, Self-sufficiency: `path.go:205` `drawSteps`: three named modes replace the nil-ness of pins and memo, checked in three places
- Shape, Locality: `table.go:237` `tableRoute`: one shared vocabulary with the walk's pathPos
- Self-sufficiency: `render.go:147` `readField`: the comment names the action at a distance, a reference clearing the pick in enter

## 2026-11-11T00:16:30Z, PR #343 at 8613dd6, against 0b9b2ce

Ruling: better

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | same | same | better |
| Inherited architect | same | better | same | same | better |

Mid A, decided by:

- Locality: `loadfinds.go:9` `loadFinds`: what one load finds sits in one value behind one read rule, and only finish writes it into the nodes, where the base kept it in three homes, each with its own fallback to an earlier load
- Locality: `namefence.go:343` `refuseDrawnApart`: reads inner.target and outer.target off the nodes, where the base read one binding through targetOf and the other through targets, falling back unevenly
- Locality: `data.go:158` `pipeline.run`: the order of stages reads in one function, and its comment says which read p.finds and which the nodes, where the base hid the fences inside resolveTemplates
- Locality: `graph.go:139` `renderCount`: a nil loadFinds is a mode meaning only what was finished, and loadInline pairs the l it compiled into with the one it runs, a smaller cost than the split it replaces
- Locality: `record.go:279` `recordOf`: the guard t.templateFinds == nil says the load hasn't finished, where the base made the reader derive it from counts

Inherited architect, decided by:

- Locality: `loadfinds.go:9` `loadFinds`: one side table, written into the nodes only by finish, answers whether a field is written yet, where the base wrote templates in place and kept bindings and draws in two maps with their own lookups
- Locality, Shape: `resolve.go:151` `fenceTemplates`: the fences read b.addressed, b.target and b.values off the nodes, so each reads on its own, and resolvedNames drops from six fields to three
- Locality: `internal/rows/rows.go:288` `Link`: writes only its own table, since child tables come from the folder, where the base also wrote the parent's children, which putBack could leave stale
- Shape: `namefence.go:292` `readTwiceApart`: x.in.startOf(x.a.head, x.in.refs) hands a method its receiver's own field, which with the next place cancels the slimmer fences' gain
- Shape: `datatype.go:63` `checkColumns`: calls after finish pass a bare nil for l, a rule only of's comment states, and three spellings reach an entry with nothing saying why each call picks its own
- Locality: `columnread.go:133` `itemDatatype`: reads l.of(t).readsColumn while the next function reads t.readsColumn, and only pipeline.run's comment says which phase makes each right

## 2026-11-11T01:35:08Z, PR #345 at 4fc9122, against ab57ff0

Ruling: same

| Seat | Navigation | Locality | Shape | Self-sufficiency | Overall |
|---|---|---|---|---|---|
| Mid A | same | better | worse | same | same |
| Inherited architect | same | better | worse | worse | same |

Mid A, decided by:

- Locality: `internal/engine/tree/node.go:69` `Template`: its syntax and what its load found are two named parts, and only `LoadFinds.Finish` writes the found part
- Locality: `table.go:18` `compileTable`: the compile steps return values that `Table.SetRows` and `Table.SetSyntax` store once, where they wrote the table's fields as side effects
- Shape: `where.go:246` `newValueDraw`: `tree.NewValueDraw` takes nine positional arguments, a bare bool and same-typed pairs that swap silently, where fields were set by name
- Shape: `internal/engine/tree/node.go:134` the getter block: dozens of one-line getters with receiver `r`, and call chains heavier than the field reads they replaced

Inherited architect, decided by:

- Locality: `internal/engine/tree/loadfinds.go:72` `LoadFinds.SetNameScope`: every write of a load's results goes through one file of named setters, and `Finish` alone gives a node its finds
- Locality: `internal/engine/tree/node.go:69` `Template`: the syntax and finds split shows the two lifetimes in the type, and `Finished()` names the check
- Shape: `where.go:246` `newValueDraw`: nine positional arguments, a bare bool and swappable `LengthEnd`s; `NewAlphabetDraw` and `NewPathStep` lose names the same way
- Shape: `internal/engine/tree/node.go:134` the getter block: about a hundred getters at the bottom of each file, away from their type, returning live maps and slices
- Self-sufficiency: `internal/engine/tree/pick.go:3` `PickKey`: its doc explains it through root functions a reader of `tree` can't reach without opening another package
