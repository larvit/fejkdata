# Part 3 — Science/nature, Text, Dates: open data sources for an MIT fake-data library

Researched 2026-09-17. "OK to embed" = redistributable inside an MIT repo with the stated notice. Anything not fetched first-hand is marked **UNVERIFIED**.

## 7. Science / nature

### Periodic table

| Source | Licence | Embed in MIT repo | Fields | Rows | Format |
|---|---|---|---|---|---|
| PubChem periodic table — CSV https://pubchem.ncbi.nlm.nih.gov/rest/pug/periodictable/CSV, JSON https://pubchem.ncbi.nlm.nih.gov/rest/pug/periodictable/JSON (page: https://pubchem.ncbi.nlm.nih.gov/periodic-table/) | US-government work, public domain. NLM policy (https://www.ncbi.nlm.nih.gov/home/about/policies/): "Information that is created by or for the US government on this site is within the public domain … may be freely distributed and copied. However, it is requested that in any subsequent use of this work, NLM be given appropriate acknowledgment." | Yes; add an acknowledgment line ("Element data: PubChem/NLM"). | `AtomicNumber, Symbol, Name, AtomicMass, CPKHexColor, ElectronConfiguration, Electronegativity, AtomicRadius, IonizationEnergy, ElectronAffinity, OxidationStates, StandardState, MeltingPoint, BoilingPoint, Density, GroupBlock, YearDiscovered` (17) | 118 (Z=1..118, verified) | CSV; JSON is `{Table:{Columns:{Column:[...]},Row:[{Cell:[...]}]}}` |
| NIST periodic table https://www.nist.gov/pml/periodic-table-elements | Public domain (17 USC 105); NIST asks for citation (https://www.nist.gov/open/copyright-fair-use-and-licensing-statements-srd-data-software-and-technical-series-publications) | Yes, but | PDF only (no CSV/JSON) — not usable as a data source | 118 | PDF |
| Bowserinator/Periodic-Table-JSON https://github.com/Bowserinator/Periodic-Table-JSON | **CC BY-SA 3.0** (LICENSE.md, verified; data derived from Wikipedia) | **No** — ShareAlike is incompatible with MIT redistribution | 33 keys incl. `name, symbol, number, atomic_mass, category, phase, period, group, block, shells, electron_configuration, summary, cpk-hex, …` | **119** (includes hypothetical Ununennium) | JSON, CSV |
| faker-js `src/locales/en/science/chemical_element.ts` https://github.com/faker-js/faker (branch `next`) | MIT (LICENSE verified) | Yes, keep MIT notice | `{symbol, name, atomicNumber}` | 118 | TS |

Validation rules: `AtomicNumber` 1..118 unique; `Symbol` 1–2 letters, first uppercase, unique; `Name` unique. PubChem leaves numeric fields empty for Z≥104 where unknown; `ElectronConfiguration` for Og is `"[Rn]7s2 7p6 5f14 6d10 (predicted)"` and `StandardState` is `"Expected to be a Gas"` — strip "(predicted)"/"Expected to be a" if you want an enum. `AtomicMass` for unstable elements is the mass number of the most stable isotope as a bare decimal (e.g. `295.216`), no brackets. `CPKHexColor` is 6 hex digits without `#`, empty for some.

### SI units

- Source: BIPM SI Brochure, 9th ed. (2019; current revision v4.01, 2026) https://www.bipm.org/en/publications/si-brochure ; PDF https://www.bipm.org/documents/20126/41483022/SI-Brochure-9-EN.pdf
- Licence: **CC BY 4.0** — stated on the brochure page and on BIPM's copyright page (https://www.bipm.org/en/copyright): content may be adapted, copied, used in commercial products with "appropriate acknowledgement of the BIPM and its source"; BIPM name/logo must not imply endorsement. Verified via the HTML pages; the PDF's own front-matter licence line is **UNVERIFIED** (could not text-extract the PDF here).
- Embed: Yes — the unit *names/symbols* are facts (not copyrightable); attribute "SI Brochure, BIPM, CC BY 4.0" anyway.
- Data (cross-checked against https://en.wikipedia.org/wiki/International_System_of_Units):
  - 7 base units: second s (time), metre m (length), kilogram kg (mass), ampere A (electric current), kelvin K (thermodynamic temperature), mole mol (amount of substance), candela cd (luminous intensity).
  - 22 coherent derived units with special names: radian rad, steradian sr, hertz Hz, newton N, pascal Pa, joule J, watt W, coulomb C, volt V, farad F, ohm Ω, siemens S, weber Wb, tesla T, henry H, degree Celsius °C, lumen lm, lux lx, becquerel Bq, gray Gy, sievert Sv, katal kat.
  - 24 prefixes: quetta Q 10^30, ronna R 10^27, yotta Y 10^24, zetta Z 10^21, exa E 10^18, peta P 10^15, tera T 10^12, giga G 10^9, mega M 10^6, kilo k 10^3, hecto h 10^2, deca da 10^1, deci d 10^-1, centi c 10^-2, milli m 10^-3, micro µ 10^-6, nano n 10^-9, pico p 10^-12, femto f 10^-15, atto a 10^-18, zepto z 10^-21, yocto y 10^-24, ronto r 10^-27, quecto q 10^-30.
- Format rules for valid output: unit symbols are case-sensitive (`s` vs `S`, `m` vs `M`); prefix and unit symbol are joined without a space (`kJ`, `µs`); a space separates number and unit (`12 kg`), except the degree/minute/second of plane angle; `kg` already carries a prefix — never `µkg`; micro is Greek mu U+03BC (U+00B5 micro sign is the legacy compatibility character); ohm is U+03A9 (U+2126 OHM SIGN is deprecated); `°C` not `° C`.
- MIT alternative: faker-js `en/science/unit.ts` — 29 `{name, symbol}` objects, MIT.

### Animals by class

- **Wikidata** (https://www.wikidata.org/wiki/Wikidata:Licensing): all structured data **CC0** — no attribution required. Embed: Yes.
  - Feasibility, live SPARQL (https://query.wikidata.org/sparql), pattern `?t wdt:P31 wd:Q16521; wdt:P105 wd:Q7432; wdt:P171* wd:<class>; wdt:P1843 ?cn FILTER(LANG(?cn)='en')` (taxon, rank species, parent-taxon chain, English common name):
    - Mammalia Q7377: **5,759** species with ≥1 English common name.
    - Aves Q5113: **12,432**.
    - Reptilia Q10811: **16,369** — suspiciously large; Wikidata's parent chain likely pulls birds/Sauropsida in. Use Squamata Q122422 + Testudines Q223044 + Crocodilia Q1387 (or check the chain) instead.
    - Insecta Q1390 and Plantae Q756: **timed out (502/504) every attempt** — trees too big for the 60 s public endpoint. Options: paginate by order (`P171` one level at a time), use the weekly JSON dump, or take GBIF instead.
  - Sample rows (mammals): `Notoryctes typhlops` → "Central Desert Marsupial Mole", "Itjaritjari", "Marsupial Mole", "Southern marsupial mole", "Southern Marsupial Mole" — i.e. **several P1843 values per taxon, inconsistent casing, and indigenous-language names tagged `en`**. Validation: pick one name per taxon (prefer the one matching the item's English `rdfs:label`), normalise case, drop names that equal the scientific name.
- **GBIF Backbone Taxonomy** https://www.gbif.org/dataset/d7dddbf4-2cf0-4f39-9b2a-bb099caae36c — licence **CC BY 4.0** (verified via https://api.gbif.org/v1/dataset/d7dddbf4-2cf0-4f39-9b2a-bb099caae36c, `license: http://creativecommons.org/licenses/by/4.0/legalcode`); required citation text: "GBIF Secretariat (2023). GBIF Backbone Taxonomy. Checklist dataset https://doi.org/10.15468/39omei". Embed: Yes with that citation. Vernacular names come via `/v1/species/{key}/vernacularNames`, each row from a *different* source checklist with its own licence (some CC BY-NC) — record and check per row. GBIF's site terms page (https://www.gbif.org/terms) was unreachable (timeout) — **UNVERIFIED** beyond the API field.
- **faker-js** `src/locales/en/animal/*` (MIT; counts = quoted string lines on branch `next`): bird 821, snake 533, dog 493, cow 467, horse 342, rodent 161, insect 129, fish 95, cat 55, cetacean 52, rabbit 49, type 44 (generic: "bat, bear, bee, …, zebra"), pet_name 42, crocodilia 24, bear 8, lion 7. These are breeds/species mixed with common names, no class field; fine as MIT drop-ins. (A WebFetch summariser claimed 1,036 dogs; the regex count is 493 — treat the exact count as UNVERIFIED.)

### Plants

- Same Wikidata approach (CC0) — but the Plantae query timed out; paginate by order or family (`?t wdt:P171 ?family . ?family wdt:P105 wd:Q35409`) or use the dump.
- GBIF Plantae (kingdomKey 6) vernacular names, CC BY 4.0, same per-row-licence caveat.
- faker-js has **no** plant list (en locale dirs: airline, animal, app, book, color, commerce, company, database, date, finance, food, hacker, internet, location, lorem, medical, music, person, phone_number, science, team, vehicle, word).
- USDA PLANTS database (public domain, has common names) — **UNVERIFIED**, not fetched.

### Planets / moons

- NASA NSSDCA Planetary Fact Sheet https://nssdc.gsfc.nasa.gov/planetary/factsheet/ — 10 columns (Mercury, Venus, Earth, Moon, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto) × 20 rows (mass 10^24 kg, diameter km, density, gravity, escape velocity, rotation period, day length, distance from Sun, perihelion, aphelion, orbital period, orbital velocity, inclination, eccentricity, obliquity, mean temp °C, surface pressure, number of moons, rings Y/N, magnetic field Y/N). HTML table only. NASA content "generally not subject to copyright in the United States" (https://www.nasa.gov/nasa-brand-center/images-and-media/); NASA asks to be acknowledged; the NASA insignia must not be used. Embed: Yes with "Source: NASA NSSDCA".
- Wikidata (CC0): planets = `?p wdt:P31/wdt:P279* wd:Q634; wdt:P397 wd:Q525` → Mercury Q308, Venus Q313, Earth Q2, Mars Q111, Jupiter Q319, Saturn Q193, Uranus Q324, Neptune Q332 (plus hypothetical Theia Q1053432 — filter out). Moons = `?m wdt:P397 <planet>; wdt:P31/wdt:P279* wd:Q2537` (natural satellite; note **Q2199 is "dwarf planet"**, not moon): Saturn 165, Jupiter 85, Uranus 29, Neptune 16, Earth 5, Mars 2. Caveats: Wikidata lags IAU counts (Saturn has 274 recognised moons as of 2025 — UNVERIFIED figure); Earth's 5 includes quasi-satellites; provisional designations (`S/2004 S 3`) are labels too — filter `^S/\d{4}` if you want proper names only.

### Blood type distribution

- **US** (Stanford Blood Center https://stanfordbloodcenter.org/donate-blood/blood-donation-facts/blood-types/, verified; cites AABB Technical Manual 18th ed.): O+ 37.4 %, A+ 35.7 %, B+ 8.5 %, AB+ 3.4 %, O− 6.6 %, A− 6.3 %, B− 1.5 %, AB− 0.6 % (sums to 100.0). Facts — freely embeddable; cite Stanford/AABB.
- American Red Cross https://www.redcrossblood.org/donate-blood/blood-types.html — **UNVERIFIED**: site returns HTTP 403 (Akamai) to both WebFetch and the headless browser. Search snippets say the Red Cross gives type O by ethnicity (Latino 57 %, African American 51 %, Caucasian 45 %) rather than an 8-way national table.
- **Global**: no authoritative figure found. Wikipedia "Blood type distribution by country" gives a population-weighted row (O+ 38.4, A+ 27.3, B+ 8.1, AB+ 2.0, O− 13.1, A− 8.1, B− 2.0, AB− 0.01) that its own editors flag "unreliable source"; WorldAtlas gives O+ 42, A+ 31, B+ 15, AB+ 5, O− 3, A− 2.5, B− 1, AB− 0.5 with no source. Recommend shipping only the US table, or per-country tables from Wikipedia (facts; Wikipedia prose is CC BY-SA but the numbers are cited to national blood services — cite those).
- Validation: ABO ∈ {A, B, AB, O}; Rh ∈ {+, −}; weights must sum to 100 ± rounding.

## 8. Text

### English word lists with part of speech

| Source | Licence | Embed | Counts | Format |
|---|---|---|---|---|
| Princeton WordNet 3.0 / 3.1 https://wordnet.princeton.edu/license-and-commercial-use ; download https://wordnetcode.princeton.edu/wn3.1.dict.tar.gz | "WordNet License" (SPDX `WordNet`), BSD-style: permission to use/copy/modify/distribute "for any purpose and without fee or royalty … provided that … the following copyright notice and statements, including the disclaimer … appear on ALL copies … including modifications". Notice text: "WordNet 3.0 Copyright 2006 by Princeton University. All rights reserved." + AS-IS disclaimer; no use of Princeton's name in advertising. | Yes — ship the LICENSE text alongside the data | WordNet 3.0 (wnstats(7WN), verified): unique strings noun 117,798 / verb 11,529 / adj 21,479 / adv 4,481 (total 155,287); synsets 82,115 / 13,767 / 18,156 / 3,621 | WNDB: `index.noun|verb|adj|adv` (one lemma per line: `lemma pos synset_cnt p_cnt [ptr…] sense_cnt tagsense_cnt synset_offset…`), `data.<pos>` (synsets with glosses). Multi-word lemmas use `_` for spaces; lower-case; entries can contain digits/apostrophes/hyphens. |
| Open English WordNet 2025 https://en-word.net/ ; https://github.com/globalwordnet/english-wordnet | **CC BY 4.0**; cite "Open English WordNet … derived from Princeton WordNet" | Yes, with attribution | 135,969 words, 107,519 synsets (README, verified) | `english-wordnet-2025.zip` (WNDB, 9.2 MB), `english-wordnet-2025.xml.gz` (WN-LMF, 10.8 MB), `english-wordnet-2025-json.zip` (9.5 MB), `.ttl.gz` (16.9 MB) |
| SCOWL v2 / ESDB https://github.com/en-wl/wordlist (branch v2) ; classic http://wordlist.aspell.net/ | Custom MIT-like (Copyright 2000-2026 Kevin Atkinson: use/copy/modify/distribute/sell "provided that the above copyright notice appears in all copies and that both the above copyright notice and this notice appear in supporting documentation"). Sources 12dicts + ENABLE2K are public domain; lists above size 80 add the UKACD notice; **POS assignments partly from WordNet, so the WordNet notice "MIGHT apply"** (their words). Australian-English parts carry an extra notice. | Yes, ship the Copyright file | Sizes 35 (small), 50 (medium), 60 (spell-check default), 70 (large), 80 (incl. game words), 85 (archaic). No per-size counts published in the v2 README. Classic SCOWL sizes 10–95 — counts **UNVERIFIED** (README not reachable). | v2: `scowl.db` (SQLite), `scowl.txt`, Python `libscowl`. ESDB rows carry POS, spelling (A/B/Z/C/D) and region codes. |
| 12dicts http://wordlist.aspell.net/12dicts-readme/ | Public domain ("I explicitly release them to the public domain, but request acknowledgment"), **except** `2of12inf` and the `2+2+3*` lists, which depend on AGID and inherit its terms | Yes (PD lists); acknowledge Alan Beale | 3esl ≈22k, 6of12 ≈32k, 2of12 ≈41k, 2of12inf ≈82k, 3of6game ≈65k, 5d+2a ≈68k, 3of6all ≈83k, 2+2+3lem ≈84k, 2of5core ≈4.7k, 6phrase ≈22k, neol2016 ≈600 | Plain text, one word per line; marker suffixes `+ ! ^ &` (not POS). **No POS tags** in any current 12dicts list. |
| Moby Part-of-Speech II https://www.gutenberg.org/ebooks/3203 ; file https://www.gutenberg.org/files/3203/files/mobypos.txt | Public domain ("Public Domain material by grant from the author, January, 2001", in the package README) | Yes, no notice needed | **233,356** lines (verified) | One entry per line: `word\CODES`, CRLF line ends (e.g. `A-line\NA`, `a tempo\Avh`). Codes in priority order: `N` noun, `p` plural, `h` noun phrase, `V` verb (usu. participle), `t` transitive verb, `i` intransitive verb, `A` adjective, `v` adverb, `C` conjunction, `P` preposition, `!` interjection, `r` pronoun, `D` definite article, `I` indefinite article, `o` nominative. Many entries are phrases, proper nouns, or 1990s-era; non-ASCII entries use a legacy 8-bit encoding (e.g. `a bon march\v` lost its `é`) — restrict to ASCII `[a-z]+` for a clean list. |
| faker-js `en/word/*` (MIT) | MIT | Yes | adjective 1000, noun 1000, verb 1000, adverb 325, preposition 109, conjunction 51, interjection 46 | TS string arrays |

### Word frequency lists

| Source | Licence | Verdict |
|---|---|---|
| Google Books Ngram data v2/v3 https://storage.googleapis.com/books/ngrams/books/datasetsv3.html | **CC BY 3.0** ("This compilation is licensed under a Creative Commons Attribution 3.0 Unported License") | Embeddable with attribution. Format `ngram TAB year TAB match_count TAB volume_count`; 1-gram files split by leading letter, GBs each — aggregate offline to a top-N list. |
| wordfreq https://github.com/rspeer/wordfreq | Code **Apache 2.0** (not MIT). Data: "may be redistributed under a Creative Commons Attribution-ShareAlike 4.0 license" (mix of Google Books, Wikipedia, OpenSubtitles, SUBTLEX, Leeds, ParaCrawl, Twitter); maintainer notes the CSV export "does not follow the CC-By-SA license". Project in sunset mode (data snapshot ≈2021). | **Avoid** embedding the data (ShareAlike). |
| Peter Norvig count_1w.txt https://norvig.com/ngrams/ | Page says "Code … under the MIT license"; the **data** comes from the Google Web 1T corpus distributed by LDC under LDC terms — no data licence stated. | **UNVERIFIED / avoid**. 333,333 words, `word TAB count`, lowercase. |
| hermitdave/FrequencyWords https://github.com/hermitdave/FrequencyWords | "MIT License for code. CC-by-sa-4.0 for content." | **Avoid** (ShareAlike). |

### Lorem ipsum

- Canonical paragraph (Letraset, 1966; public domain — garbled Cicero): "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
- Source: Cicero, *De finibus bonorum et malorum* 1.10.32–33 (45 BC), public domain; the "dolorem ipsum" fragment begins "Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit…". Reference: https://en.wikipedia.org/wiki/Lorem_ipsum
- MIT word pool alternative: faker-js `en/lorem/word.ts` — 999 words.

### Quote collections

- Project Gutenberg #27889 https://www.gutenberg.org/ebooks/27889 — Bartlett, *Familiar Quotations*, **9th edition** (title page "NINTH EDITION", copyright 1875/1882/1891/1903) — not the 1919 10th. "Public domain in the USA." Plain text 3.1 MB, HTML zip 50 MB, EPUB.
- Project Gutenberg #16732 https://www.gutenberg.org/ebooks/16732 — an early (Hurst & Co. reprint) edition, no edition number given. PD in USA.
- 10th edition (1914/1919, ed. Nathan Haskell Dole) exists on archive.org (https://archive.org/details/bartlettsfamilia0000john_o4b4) — not found on Gutenberg; **UNVERIFIED** availability as clean text.
- Either Bartlett text needs parsing (quote / author / source headings); no structured file. Wikiquote is CC BY-SA — avoid.
- Validation: strip Gutenberg header/footer (`*** START OF THE PROJECT GUTENBERG EBOOK` … `*** END …`) before extraction; Gutenberg's trademark terms only bind if you keep the "Project Gutenberg" name on the output — strip it.

### Colour names

| Source | Licence | Embed | Count / format |
|---|---|---|---|
| CSS Color Module Level 4 §named colors https://www.w3.org/TR/css-color-4/#named-colors | W3C Document License (2023) https://www.w3.org/copyright/document-license-2023/ — copy/distribute allowed for any purpose; derivatives allowed only "to facilitate implementation of the technical specifications"; notice required: "Copyright © 2023 W3C®. This software or document includes material copied from or derived from [title and URI]." | Yes — the name→hex table is factual data and is copied into every browser; include the W3C notice line with the spec title + URI. | **148** names (verified: aliceblue … yellowgreen, incl. `rebeccapurple #663399`); all names lowercase ASCII, hex 6 lowercase digits. |
| XKCD colour survey https://xkcd.com/color/rgb.txt | **CC0** (file header: `# License: https://creativecommons.org/publicdomain/zero/1.0/`) | Yes, no notice needed | **949** lines (verified). Format `name TAB #rrggbb TAB` (note the trailing tab). Max name length 26; all hex valid lowercase; no duplicate names. Contains crude names ("poo", "baby poop green", "diarrhea", "vomit") — apply a blocklist. |
| faker-js `en/color/human.ts` | MIT | Yes | 31 names |
| Pantone | Proprietary | **Avoid** | — |

## 9. Dates

### IANA tz database

- Source https://www.iana.org/time-zones ; files https://data.iana.org/time-zones/tzdb/ (current version **2026d**, verified). LICENSE: "Unless specified below, all files in the tz code and data (including this LICENSE file) are in the public domain" (only `date.c`, `newstrftime.3`, `strftime.c` are BSD-3). Embed: Yes, no notice needed.
- `zone1970.tab` — **312** data rows; tab-separated; UTF-8; `#` comments. Columns: (1) comma-separated ISO 3166-1 alpha-2 codes of countries overlapping the zone, most-populous first; (2) coordinates in ISO 6709 `±DDMM±DDDMM` or `±DDMMSS±DDDMMSS` (lat then lon, no separator); (3) TZ name (`Europe/Zurich`); (4) comment, present iff the country has several zones. One row per timezone where civil time has agreed since 1970.
- `zone.tab` — **418** rows, older one-country-per-row table (column 1 is a single country code); kept for backward compatibility. Prefer `zone1970.tab`.
- `backward` — **252** `Link TARGET LINK-NAME` lines mapping old/merged names (`US/Eastern`, `Asia/Calcutta`, `Europe/Kiev`) to current ones; also a few `Zone` entries. Use it to (a) accept aliases as input and (b) never emit them as output.
- UTC offsets: do **not** embed — offsets depend on DST rules that change several times a year. Derive at runtime: `new Intl.DateTimeFormat('en-US', { timeZone, timeZoneName: 'longOffset' }).formatToParts(date)` → part `timeZoneName` = `GMT-05:00` (or `GMT` for zero) — strip `GMT`, treat bare `GMT` as `+00:00`. Validation of a generated zone name: `Intl.supportedValuesOf('timeZone')` (canonical names only; whether links like `US/Eastern` appear depends on the ICU build) or `try { new Intl.DateTimeFormat(undefined, { timeZone }) } catch { invalid }`, which accepts links too.

### ISO 8601 layouts (https://en.wikipedia.org/wiki/ISO_8601)

| Layout | Extended | Basic | Example |
|---|---|---|---|
| Calendar date | `YYYY-MM-DD` | `YYYYMMDD` | 2009-01-06 |
| Ordinal date | `YYYY-DDD` | `YYYYDDD` | 1981-095 |
| Week date | `YYYY-Www-D` | `YYYYWwwD` | 2009-W01-1 (weeks start Monday, week 1 contains 4 Jan) |
| Time | `hh:mm:ss[.fff]` | `hhmmss` | 13:47:30 (24 h; `T` prefix optional standalone) |
| UTC / offset | `Z` or `±hh:mm` | `±hhmm`, `±hh` | 2007-04-05T14:30Z, 2024-06-01T09:00-05:00 |
| Date-time | date `T` time offset | | 2007-04-05T14:30:00+02:00 |
| Duration | `PnYnMnDTnHnMnS` / `PnW` | | P3Y6M4DT12H30M5S |
| Interval | start`/`end, start`/`duration, duration`/`end | | 2007-03-01T13:00:00Z/2008-05-11T15:30:00Z |
| Recurring | `Rn/`interval | | R5/2008-03-01T13:00:00Z/P1Y2M10DT2H30M |

RFC 3339 profile (what JSON/APIs usually mean): extended forms only, `T`/`Z` may be lowercase, space allowed instead of `T`, `-00:00` = unknown offset, no durations/intervals/week/ordinal dates. Generate RFC 3339 by default.

### Unicode CLDR (locale month/weekday names, date patterns)

- Licence: **Unicode License v3** (SPDX `Unicode-3.0`), OSI-approved 2023-11-17 (https://opensource.org/license/unicode-3-0). Text https://www.unicode.org/license.txt — MIT-style: use/copy/modify/sell "provided that either (a) this copyright and permission notice appear with all copies of the Data Files or Software, or (b) this copyright and permission notice appear in associated Documentation"; no use of the Unicode name in advertising. Embed: Yes; put the notice in LICENSE/NOTICE.
- npm: `cldr-core`, `cldr-dates-full`, `cldr-numbers-full` (peer of dates), etc.; current **48.2.0** (CLDR 48, verified via npm registry, `license: Unicode-3.0`); repo https://github.com/unicode-org/cldr-json .
- Path is **`cldr-dates-full/main/<locale>/ca-gregorian.json`** (not `dates/gregorian.json`), rooted at `main.<locale>.dates.calendars.gregorian`. Keys (verified for `en`): `months`, `days`, `quarters`, `dayPeriods`, `eras`, `dateFormats`, `dateSkeletons`, `timeFormats`, `timeSkeletons`, `dateTimeFormats`, `dateTimeFormats-atTime`, `dateTimeFormats-relative`.
  - `months.{format|stand-alone}.{abbreviated|narrow|wide}` keyed `"1".."12"`; `days.{format|stand-alone}.{abbreviated|narrow|short|wide}` keyed `sun..sat`.
  - `dateFormats`: en = full `EEEE, MMMM d, y`, long `MMMM d, y`, medium `MMM d, y`, short `M/d/yy`.
  - `timeFormats`: en = full `h:mm:ss a zzzz`, long `h:mm:ss a z`, medium `h:mm:ss a`, short `h:mm a` — **the `a` is preceded by U+202F NARROW NO-BREAK SPACE**; `*-alt-ascii` variants use a plain space. Pick one spelling (ASCII) and ignore the other.
  - `dateTimeFormats.{full|long|medium|short}` = glue like `{1}, {0}` ({1} date, {0} time); `availableFormats` = skeleton→pattern map (`yMMMd` → `MMM d, y`), plus `intervalFormats`, `appendItems`.
  - `dayPeriods.format.abbreviated`: `am: AM`, `pm: PM`, plus `midnight`, `noon`, `morning1`… variants.
- Pattern letters follow LDML (https://unicode.org/reports/tr35/tr35-dates.html#Date_Field_Symbol_Table): `y` year, `M`/`L` month (format/stand-alone), `d` day, `E`/`c` weekday, `h` 1–12, `H` 0–23, `a` AM/PM, `z`/`zzzz` zone name, `x`/`X` ISO offset; literal text in single quotes. Same patterns drive `Intl.DateTimeFormat` internally, so runtime `Intl` can replace embedding for month/weekday names if bundle size matters.

## Summary of licence verdicts

- Embed freely (PD/CC0): PubChem periodic table, NIST, NASA fact sheet, Wikidata, Moby POS, 12dicts (PD lists), Bartlett via Gutenberg, XKCD colours, IANA tz, Lorem ipsum/Cicero.
- Embed with notice: WordNet 3.x (WordNet License), Open English WordNet (CC BY 4.0), SCOWL/ESDB (custom MIT-like), GBIF backbone (CC BY 4.0 + citation), BIPM SI (CC BY 4.0), CSS named colours (W3C notice), CLDR (Unicode-3.0), Google Books ngrams (CC BY 3.0), faker-js lists (MIT).
- Avoid: Bowserinator periodic table (CC BY-SA 3.0), wordfreq data (CC BY-SA 4.0), hermitdave (CC BY-SA 4.0), Norvig count_1w (LDC-derived, unstated), Wikiquote (CC BY-SA), Pantone.
- UNVERIFIED: American Red Cross figures (site blocks fetches), any global blood-type table, BIPM PDF front-matter licence line, classic SCOWL per-size counts, USDA PLANTS, IAU moon counts, exact faker dog count, Bartlett 10th-edition text availability.
