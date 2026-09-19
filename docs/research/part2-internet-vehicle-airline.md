# Open data sources: Internet, Vehicle, Airline (researched 2026-09-17)

Verdict key: **EMBED** = safe to vendor into an MIT repo; **EMBED+ATTR** = ok with a notice; **AVOID** = licence incompatible or unclear.

## 4. Internet

### IANA registries (all of them)
- Licence: IANA/IETF statement at https://www.iana.org/help/licensing-terms — "IANA and IETF intend that the Protocol Registries may be freely used by any party for any purpose", data dedicated under **CC0 1.0**. No attribution required. **EMBED**.
- All registries below share this licence; ship one `SOURCES`/NOTICE line pointing at that URL.

| Registry | CSV URL | Rows (2026-09-16) | Notes |
|---|---|---|---|
| Media types | `https://www.iana.org/assignments/media-types/<type>.csv`, type ∈ application, audio, font, haptics, image, message, model, multipart, text, video (also `example`) | application 1800, audio 165, font 6, haptics 3, image 88, message 27, model 42, multipart 17, text 105, video 97 (header excluded) ≈ **2350** | Fields: `Name,Template,Reference`. Template is the full `type/subtype`; some rows are DEPRECATED/OBSOLETED in Name (filter). Format: RFC 6838 §4.2 — `type "/" subtype`, restricted chars, max 127 chars each, case-insensitive. |
| HTTP status codes | https://www.iana.org/assignments/http-status-codes/http-status-codes-1.csv | 75 rows, **64 assigned** (306 and 418 are "(Unused)"; 104 is TEMPORARY) → 62 usable | Fields: `Value,Description,Reference`. Value is 3 digits 1xx–5xx. |
| IPv4 special-purpose | https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry-1.csv | 26 blocks | Fields: `Address Block,Name,RFC,Allocation Date,Termination Date,Source,Destination,Forwardable,Globally Reachable,Reserved-by-Protocol`. |
| IPv6 special-purpose | https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry-1.csv | 27 blocks | Same fields. Contains `2001:db8::/32` and `3fff::/20`. |
| Service names / ports | https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv | 14,534 rows; 11,731 with both name and a single port; **6,102 distinct named ports**; 6,008 tcp-named; 707 tcp names < 1024 | Fields: `Service Name,Port Number,Transport Protocol,Description,Assignee,Contact,Registration Date,Modification Date,Reference,Service Code,Unauthorized Use Reported,Assignment Notes`. Port Number may be empty, single, or a range `a-b`. Well-known = 0–1023, registered 1024–49151, dynamic 49152–65535 (RFC 6335). ~1.1 MB; embed only the named-port subset. |
| TLDs | https://data.iana.org/TLD/tlds-alpha-by-domain.txt | **1,401** TLDs (first line is a `# Version …` comment) | Upper-case ASCII; IDNs as `XN--…` punycode. Same IANA CC0 terms (UNVERIFIED that data.iana.org is explicitly covered by the licensing page; it is the same registry data). Lower-case on use. |

### MIME extension mapping: jshttp/mime-db
- https://github.com/jshttp/mime-db — **MIT** (GitHub API spdx MIT; npm `license: MIT`), version 1.54.0 on npm.
- `db.json` (CDN: https://cdn.jsdelivr.net/npm/mime-db/db.json): object keyed by lower-case type → `{source, extensions[], compressible, charset}`.
- Counts (1.54.0): **2,522 types**, **1,015 with extensions**; source = iana 2,136, apache 275, nginx 13, custom 98.
- Sources: IANA registry (CC0), Apache httpd `mime.types` (Apache-2.0), nginx `mime.types` (BSD-2). **EMBED+ATTR** (keep the MIT copyright notice). Data updates are not semver-breaking per README.

### IPv4 / IPv6 generation rules (RFCs, no data file needed)
- RFC 5737 documentation: `192.0.2.0/24` (TEST-NET-1), `198.51.100.0/24` (TEST-NET-2), `203.0.113.0/24` (TEST-NET-3). https://www.rfc-editor.org/rfc/rfc5737.html
- RFC 1918 private: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`. Also useful: `100.64.0.0/10` CGNAT (RFC 6598), `198.18.0.0/15` benchmarking (RFC 2544), `169.254.0.0/16` link-local.
- IPv6 documentation: `2001:db8::/32` (RFC 3849) and `3fff::/20` (RFC 9637, "expands on the existing 2001:db8::/32 … with the reservation of an additional, larger prefix"). ULA `fc00::/7` (use `fd00::/8`), link-local `fe80::/10`.
- Validation: generate inside those blocks; avoid network/broadcast (`.0`/`.255`) if the consumer expects host addresses. Format IPv6 per RFC 5952 (lower-case hex, `::` once, longest zero run).

### MAC OUI (IEEE MA-L)
- CSV: https://standards-oui.ieee.org/oui/oui.csv (3.8 MB, **40,161 rows**, fields `Registry,Assignment,Organization Name,Organization Address`; Assignment is 6 hex chars). Server returns HTTP 418 to non-browser user agents; fetch with a browser UA. Also MA-M `oui28.csv`, MA-S `oui36.csv`, CID `cid.csv`.
- Licence: no licence on the IEEE pages themselves (site footer "© IEEE – All rights reserved" is generic). Debian's `ieee-data` copyright file (https://metadata.ftp-master.debian.org/changelogs/main/i/ieee-data/unstable_copyright) records IEEE's 2014 statement: *"IEEE does not assert any copyright in the OUI Public Listing or attempt to restrict distribution of the listing in any way. The IEEE Registration Authority does, however, strongly encourage those who use the list to obtain it directly from IEEE…"* — same text in the FSF directory. Debian ships it in `main` on that basis. **EMBED+ATTR** (cite the statement); the org-address column is not needed — keep `Assignment,Organization Name` only (~1 MB → few hundred KB). UNVERIFIED: no current IEEE page restates it (regauth FAQ and IPR pages do not mention it).
- Alternative needing no data: locally administered unicast MAC — first octet's bit 1 (U/L) = 1, bit 0 (I/G) = 0 → second hex digit ∈ {2, 6, A, E}, e.g. `x2:xx:xx:xx:xx:xx`. Guaranteed never to collide with a vendor OUI. Notations: `01:23:45:67:89:ab`, `01-23-45-67-89-AB`, Cisco `0123.4567.89ab`.

### User-agent strings
| Source | Licence | Verdict |
|---|---|---|
| https://www.useragents.me/ | No licence stated anywhere on the page ("© 2022–2025 Useragents.me"); `/api` returns 404 now; JSON blobs embedded in page, weekly updates | **AVOID** (unlicensed = all rights reserved) |
| https://explore.whatismybrowser.com/useragents/explore/ + API legal https://developers.whatismybrowser.com/api/about/legal/ | Proprietary: "You may NOT share the unique database download URL, or the downloaded file itself … EXCEPT for making it publicly available on the internet" | **AVOID** |
| npm `user-agents` (intoli) https://github.com/intoli/user-agents | **BSD-2-Clause**, v2.1.185, daily-rebuilt dataset from Intoli traffic, weighted by frequency; fields `userAgent, platform, vendor, appName, deviceCategory, screenWidth/Height, viewportWidth/Height, connection{…}, oscpu, cpuClass, pluginsLength` | **EMBED+ATTR** (keep BSD notice). Count not stated on the page (UNVERIFIED; dataset is a few thousand rows). |
| npm `top-user-agents` (microlinkhq) https://github.com/microlinkhq/top-user-agents | **MIT** (GitHub API). Top 100 UA strings from microlink.io traffic, weekly; JSON: https://cdn.jsdelivr.net/gh/microlinkhq/top-user-agents@master/src/index.json (+ desktop.json, mobile.json) | **EMBED+ATTR** — smallest clean option |
- None of the above is CC0. Alternative: generate UAs from templates (Chrome/Firefox/Safari/Edge strings are fixed-form; only version numbers vary), no data needed.

### Programming languages: GitHub Linguist
- https://raw.githubusercontent.com/github-linguist/linguist/main/lib/linguist/languages.yml — **MIT** (repo LICENSE, "Copyright (c) 2017 GitHub, Inc."). **EMBED+ATTR**.
- **835** top-level entries; **562** with `type: programming` (others: data, markup, prose). Per-entry fields: `type, color, extensions, aliases, tm_scope, ace_mode, language_id, group, interpreters, …`.

### Operating systems
- No open dataset found beyond awesome-lists. OS names/versions are facts — hand-author a small table (Windows 10/11, Windows Server, macOS versions+names, Ubuntu/Debian/Fedora/Arch/Alpine, iOS, Android, FreeBSD…). `user-agents` (BSD-2) carries `platform`/`oscpu` values if a data-driven list is wanted.

### Database engines / collations
- Engine names: hand-author (facts).
- MySQL manual (https://dev.mysql.com/doc/refman/8.4/en/preface.html): Oracle terms — "you may not … copy, reproduce … distribute … any part" except distributing the docs with the software. **AVOID copying tables from MySQL docs.** Instead dump `SHOW COLLATION` from a `mysql:8.4.x` container (server output is factual, not the manual) — ~286 collations / 41 charsets in 8.x (UNVERIFIED count). PostgreSQL docs are under the PostgreSQL licence (permissive, keep notice) — `pg_collation` dump or docs both fine. MariaDB KB is CC BY-SA — avoid.

## 5. Vehicle

### NHTSA vPIC
- API base `https://vpic.nhtsa.dot.gov/api/vehicles/`, `?format=json|xml|csv`. Endpoints verified: `GetAllMakes` (**12,363 makes**, fields `Make_ID,Make_Name`; 612 KB JSON, mostly small trailer/custom builders — filter with `GetMakesForVehicleType/car` → 195 makes, `/truck` → 207), `GetModelsForMake/honda` (362 models, `Make_ID,Make_Name,Model_ID,Model_Name`), `DecodeWMI/1HG`, `GetWMIsForManufacturer/honda` (45 WMIs, fields `WMI,Name,Country,VehicleType,Id,…`), `GetVehicleVariableList`, `GetVehicleVariableValuesList/<display name>` (must be the display name, URL-encoded, e.g. `Fuel%20Type%20-%20Primary`; the camel-case name returns 0 rows).
- Value lists (counts): **Fuel Type - Primary 14**, **Body Class 71**, **Transmission Style 12**, **Drive Type 23**, **Vehicle Type 9**. Fields `Id,Name`.
- Bulk: https://vpic.nhtsa.dot.gov/downloads/ — `vPICList_lite_2026_08.bak.zip` (MS SQL Server 2019+ backup, 190 MB, updated 2026-08-14), also `.plain.zip` / `.custom.zip`. Page says standalone DB "limited to VIN decoding"; makes/models/variables still via API. Contains `Wmi`, `Make`, `Model`, `WMIYearValidChars` tables (UNVERIFIED table names).
- Licence: FAQ https://vpic.nhtsa.dot.gov/api/home/index/faq — "No, NHTSA is a government agency and the services provided on the API are free for use by the public as an offering as a part of our Open Data initiatives"; no rate-limit number stated, batch jobs asked to run at night. Work of the US federal government → public domain under 17 USC §105; data.gov entry lists licence as "unknown". **EMBED** (note source; no attribution legally required).

### VIN (ISO 3779 / 49 CFR 565, North America)
- 17 chars from `0-9 A-H J-N P R-Z` (**I, O, Q never**). Positions: 1–3 WMI, 4–8 VDS, 9 check digit, 10 model year, 11 plant, 12–17 serial (last 4 must be digits for US, positions 12–14 alphanumeric — 49 CFR 565.15).
- Check digit: transliterate `A=1 B=2 C=3 D=4 E=5 F=6 G=7 H=8 J=1 K=2 L=3 M=4 N=5 P=7 R=9 S=2 T=3 U=4 V=5 W=6 X=7 Y=8 Z=9`, digits as-is; weights `8 7 6 5 4 3 2 10 0 9 8 7 6 5 4 3 2`; sum mod 11 → digit, remainder 10 → `X`.
- Model year (pos 10): excludes `I O Q U Z 0`; `A=2010 … H=2017 J=2018 K=2019 L=2020 M=2021 N=2022 P=2023 R=2024 S=2025 T=2026 V=2027 W=2028 X=2029 Y=2030`, `1=2031…9=2039` (30-year cycle; `1–9` also = 2001–2009). Source: https://en.wikipedia.org/wiki/Vehicle_identification_number (rules are facts; do not copy prose).

### WMI list
- SAE J1044 / SAE WMI database is paid (https://www.sae.org/standards/j1044_202501-world-manufacturer-identifier) — **AVOID**.
- Wikipedia "List of WMIs" — CC BY-SA — **AVOID**.
- Use vPIC: `GetWMIsForManufacturer/<name or id>` per manufacturer, or the `Wmi` table in the standalone `.bak`. Public domain. Country prefix rule (first char): `1,4,5` USA, `2` Canada, `3` Mexico, `J` Japan, `K` Korea, `S` UK, `W` Germany, `Y` Sweden/Finland, `Z` Italy, `L` China (ISO 3780 facts).

### US licence plate formats
| Source | Licence | Verdict |
|---|---|---|
| openalpr `runtime_data/postprocess/us.patterns` (https://github.com/openalpr/openalpr) | **AGPL-3.0** | **AVOID** (291 lines, `@`=letter `#`=digit — useful only as a reference to check hand-authored patterns) |
| Wikipedia "United States license plate designs and serial formats" | CC BY-SA | **AVOID copying**; formats themselves are facts |
| jonnii/platekit | MIT, but SVG rendering, no serial patterns | not useful |
- No MIT/CC0 per-state pattern list found. Recommendation: hand-author a per-state table (facts); current standard passenger formats as listed by Wikipedia (verify against DMV pages before use): AL `0AXXXXX`/`00AXXXX`, AK `ABC 123`, AZ `XXX 1XX`, AR `ABC 12D`, CA `1ABC123` (digit-3 letters-3 digits), CO `ABC-D12`, CT `AB·12345`, DE `123456`, DC `AB-1234`, FL `ABC D12`, GA `ABC1234`, HI `ABC 123`, ID `A 1234U`, IL `AB 12345`, IN `123ABC`, IA `ABC 123`, KS `1234ABC`, KY `ABC123`, LA `123 ABC`, ME `123·ABC`, MD `1AB2345`, MA `1ABC 23`, MI `ABC 1234`, MN `ABC-123`, MS `ABC 123`, MO `AB1 C2D`, MT `0-AB1234`, NE `ABC 123`, NV `123·A45`, NH `123 4567`, NJ `D12-ABC`, NM `123-ABC`, NY `ABC-1234`, NC `ABC-1234`, ND `123 ABC`, OH `ABC 1234`, OK `ABC-123`, OR `123 ABC`, PA `ABC1234`, RI `1AB 234`, SC `123ABC`, SD `0A1 234`, TN `ABC 1234`, TX `ABC-1234`, UT `A12 3BC`, VT `ABC 123`, VA `ABC-1234`, WA `ABC1234`, WV `X1A 2345`, WI `ABC-1234`, WY `1A-123A`. Most states omit I/O/Q from serials (state-specific, UNVERIFIED per state).

## 6. Airline

### OurAirports
- https://ourairports.com/data/ — "All data is released to the Public Domain"; mirror repo https://github.com/davidmegginson/ourairports-data is **Unlicense** (GitHub API). Nightly updates. **EMBED**.
- Raw URLs: `https://davidmegginson.github.io/ourairports-data/{airports,airport-frequencies,runways,navaids,countries,regions}.csv`.

| File | Rows | Fields |
|---|---|---|
| airports.csv (12.7 MB) | **86,089**; by type: small_airport 42,734, heliport 23,216, closed 13,524, medium 4,106, seaplane_base 1,273, large 1,174, balloonport 62 | `id,ident,type,name,latitude_deg,longitude_deg,elevation_ft,continent,iso_country,iso_region,municipality,scheduled_service,icao_code,iata_code,gps_code,local_code,home_link,wikipedia_link,keywords` |
| — with `iata_code` | **9,055** (large 1,171, medium 3,398, small 4,233, seaplane 153, heliport 100); `scheduled_service=yes` 4,335; US with IATA 2,036 (873 large/medium) | Suggested subset: `iata_code != '' AND type IN (large_airport, medium_airport)` → 4,569 rows |
| countries.csv | **249** | `id,code,name,continent,wikipedia_link,keywords` |
| regions.csv | **3,987** | `id,code,local_code,name,continent,iso_country,wikipedia_link,keywords` (code = ISO 3166-2, e.g. `US-KS`) |
- No airline file. Derived: datasets/airport-codes on datahub (PDDL) — no need, use upstream.

### OpenFlights
- https://openflights.org/data.php — airports/airlines/routes/planes under **ODbL 1.0** (+ DbCL); airline/plane data partly from Wikipedia (GFDL/CC BY-SA). 5,888 airlines (2012). **AVOID** (share-alike). npm `airline-codes` (npow, ISC) is a straight OpenFlights sync → inherits ODbL → **AVOID**.

### Airline IATA/ICAO codes — open options
| Source | Licence | Count | Verdict |
|---|---|---|---|
| Wikidata SPARQL (https://query.wikidata.org/sparql) | **CC0** (https://www.wikidata.org/wiki/Wikidata:Licensing) | **2,928** items with IATA code (P229), **3,644** with ICAO code (P230) | **EMBED**. Query: `SELECT ?item ?itemLabel ?iata ?icao ?callsign WHERE { ?item wdt:P229 ?iata . OPTIONAL { ?item wdt:P230 ?icao } OPTIONAL { ?item wdt:P432 ?callsign } SERVICE wikibase:label { bd:serviceParam wikibase:language "en" } }`. Includes defunct airlines — filter `MINUS { ?item wdt:P576 ?dissolved }` and/or P31 = Q46970 (airline). |
| OpenTravelData `optd_airlines.csv` (https://github.com/opentraveldata/opentraveldata) | **CC BY 4.0** | 1,620 rows; `^`-separated; fields `pk^env_id^validity_from^validity_to^3char_code^2char_code^num_code^name^name2^alliance_code^alliance_status^type^wiki_link^flt_freq^alt_names^bases^key^version^parent_pk_list^successor_pk_list` | **EMBED+ATTR** (attribution notice required) |
- Format rules: IATA designator = 2 alphanumeric chars (`[A-Z0-9]{2}`, at least one letter in practice; a third optional char exists in the standard but is unused); "controlled duplicates" share a code between non-overlapping regional carriers. ICAO designator = 3 letters `[A-Z]{3}`, unique. Accounting/prefix code = 3 digits (ticket number prefix, e.g. `016` United).

### Aircraft types
| Source | Licence | Verdict |
|---|---|---|
| ICAO Doc 8643 (https://www.icao.int/operational-safety/doc-8643-aircraft-type-designators) | ICAO copyright notice: "None of the materials … may be used, reproduced or transmitted … without permission in writing from ICAO"; API Data Service is paid (25 free trial calls) | **AVOID** — including GitHub mirrors such as ColtJD45/icao-aircraft-designator-list (MIT-labelled, 7,388 rows, fields `manufacturer,model,type_designator,description,engine_type,engine_count,wtc`) because the underlying data is scraped ICAO content (UNVERIFIED provenance) |
| OpenSky aircraft database (https://opensky-network.org/data/aircraft; CSV https://s3.opensky-network.org/data-samples/metadata/aircraftDatabase.csv) | "The aircraft database is unlicensed and does not fall under our terms of use … offered as is"; built from registries, openflights.org and ICAO Doc 8643; citation requested; page says it is no longer up to date | **AVOID** ("unlicensed" ≠ open; upstream includes ODbL/ICAO). Fields for reference: `icao24,registration,manufacturericao,manufacturername,model,typecode,serialnumber,linenumber,icaoaircrafttype,operator,operatorcallsign,operatoricao,operatoriata,owner,…,engines,…,categoryDescription` |
| Wikidata | **CC0** | **550** items with ICAO aircraft type designator (P8305); also IATA aircraft code P9040 (UNVERIFIED property id), manufacturer P176 | **EMBED**. Query: `SELECT ?item ?itemLabel ?icao ?mfrLabel WHERE { ?item wdt:P8305 ?icao . OPTIONAL { ?item wdt:P176 ?mfr } SERVICE wikibase:label { bd:serviceParam wikibase:language "en" } }` |
- Format: ICAO type designator 2–4 alphanumerics (`[A-Z0-9]{2,4}`, e.g. `B738`, `A20N`, `E190`); IATA aircraft code 3 alphanumerics (`738`, `32N`).

### Airport / flight / seat format rules
- IATA airport code `[A-Z]{3}`; ICAO location indicator `[A-Z0-9]{4}` (US `K…`, Canada `C…`, UK `EG…`, Sweden `ES…`). Use OurAirports rows so codes are real.
- Flight number: 2-char IATA designator + 1–4 digits, no leading zeros in display (`QF9`, `AA1234`); systems cap at 0001–9999. Conventions (not rules): 1–999 mainline, 3000–5999 regional affiliates, ≥6000 codeshares, 8xxx charters, ≥9000 ferry/positioning. ICAO form: 3-letter designator + 1–4 alphanumerics (`AFR1`). Source: https://en.wikipedia.org/wiki/Flight_number.
- Seat: `row + letter`. Rows typically 1–~60 (A380 up to ~90); many airlines skip 13 (and 14/17 on some carriers). Letters `A–K` **skipping I** (looks like 1/l); narrow-body 3-3 = `ABC DEF` (A/F windows, C/D aisles); wide-body 3-4-3 = `ABC DEFG HJK`, 2-4-2 = `AB DEFG JK`. Regex: `^[1-9][0-9]?[A-HJK]$`.
