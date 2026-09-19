# Part 4 — ISO lists, commerce, books/music/food

Researched 2026-09-17. "Embed?" = can the data ship inside an MIT-licensed repo. Anything marked UNVERIFIED was not confirmed against the primary source.

## 10. ISO lists

### ISO 3166-1 countries

| Source | Licence | Embed in MIT repo? | Rows | Fields | Format |
|---|---|---|---|---|---|
| [datasets/country-codes](https://github.com/datasets/country-codes) | PDDL (README: "Public Domain Dedication and License"; GitHub API detects no licence file) | Yes, no attribution required. README caveat: ISO itself says its list is "for internal use and non-commercial purposes free of charge" — the repo argues no rights subsist in a list of facts | 249 | ISO3166-1-Alpha-2/-Alpha-3/-numeric, official_name_en, UNTERM names (ar/zh/en/fr/ru/es short+formal), Dial, TLD, Capital, Continent, Languages, ISO4217 currency code/name/numeric/minor_unit, Region/Sub-region/Intermediate names+codes (M49), FIPS, IOC, FIFA, MARC, ITU, WMO, GAUL, EDGAR, Geoname ID, wikidata_id, CLDR display name, is_independent, LDC/LLDC/SIDS flags | CSV `data/country-codes.csv` |
| [annexare/Countries](https://github.com/annexare/Countries) (npm `countries-list` 3.4.1, 2026-07-14) | MIT | Yes; keep MIT notice | 252 countries, 185 languages (incl. non-ISO entries like AC, XK) | per country: name, native, phone[] (calling codes), continent, capital, currency[], languages[], optional alias[], partOf; separate ISO 4217 currency table (name, native, symbol, numeric, decimals); languages: name, native | TS source, exported JSON/CSV/SQL |
| [Debian iso-codes](https://salsa.debian.org/iso-codes-team/iso-codes) | LGPL-2.1-or-later (Debian copyright file) | Not safely as embedded data in an MIT repo — LGPL applies to the data files; you would have to ship it as a separately licensed component with LGPL notice. Avoid unless accepting that | 3166-1: 249 | alpha_2, alpha_3, numeric, name, official_name, common_name, flag | JSON `data/iso_3166-1.json` (raw URL needs `?inline=false`) |
| [mledoze/countries](https://github.com/mledoze/countries) | ODbL-1.0 | No — share-alike + attribution; derived databases must be ODbL. Avoid | ~250 | common/official names, cca2/cca3/ccn3/cioc, tld, currencies, idd (calling codes), capital, region/subregion, languages, translations, latlng, borders, area, demonyms, flag emoji, UN member | JSON/CSV/XML/YAML |
| [restcountries](https://gitlab.com/restcountries/restcountries) | MPL-2.0 (source repo; hosted API v5 is proprietary/API-key) | File-level copyleft: a copied data file stays MPL and must carry the notice; OK to ship next to MIT code but not to relicense. Prefer PDDL/MIT sources | 250+ | name, tld, cca2, ccn3, cca3, cioc, fifa, independent, status, unMember, currencies, idd, capital, capitalInfo, altSpellings, region, subregion, continents, languages, translations, latlng, landlocked, borders, area, flag (emoji), demonyms, flags, coatOfArms, population, maps, gini, car, postalCode, startOfWeek, timezones | JSON `src/main/resources/countriesV3.1.json` |
| [lukes/ISO-3166-Countries-with-Regional-Codes](https://github.com/lukes/ISO-3166-Countries-with-Regional-Codes) | CC BY-SA 4.0 (LICENSE.md) | No — share-alike. Avoid | 249 | name, alpha-2, alpha-3, country-code, iso_3166-2, region, sub-region, intermediate-region + codes | CSV/JSON/XML (all, slim-2, slim-3) |
| [stefangabos/world_countries](https://github.com/stefangabos/world_countries) | Data: CC BY-SA 4.0 (README, GitHub licence detection); npm `world_countries_lists` package.json says LGPL-3.0-or-later — contradictory | No — share-alike either way. Avoid | 249 world / 193 UN; subdivisions 5046 | id (numeric), alpha2, alpha3, name (37 languages); subdivisions: country, code, name, name_en, type, parent | JSON/CSV/PHP/SQL/XML |
| [i18n-iso-countries](https://www.npmjs.com/package/i18n-iso-countries) 7.14.0 | MIT | Yes | 78 language files | alpha2, alpha3, numeric, localized names | JSON per language |
| Wikipedia ISO 3166-1 | CC BY-SA 4.0 | No (share-alike) | — | — | — |
| [ISO OBP](https://www.iso.org/obp) | Proprietary; ISO grants free use only for "internal use and non-commercial purposes" (quoted in datasets/country-codes README) | No | — | — | — |

- Flag emoji: derive, no data needed — for alpha-2 `XY`, emit U+1F1E6 + (X − 'A') and U+1F1E6 + (Y − 'A') (regional indicator symbols).
- Recommendation: `datasets/country-codes` (PDDL) covers every requested field (name, alpha2, alpha3, numeric, calling code, TLD, capital, currency, languages) in one CSV; flag emoji derived. `annexare` (MIT) is the fallback for native names/calling codes if PDDL's ISO caveat worries you; it lacks TLD.
- Validation rules: alpha-2 `^[A-Z]{2}$`, alpha-3 `^[A-Z]{3}$`, numeric `^\d{3}$` (zero-padded, keep as string); calling code 1–3 digits, NANP countries share `1`; TLD `^\.[a-z]{2}$` (ccTLD = lowercase alpha-2, exceptions: GB uses `.uk`).

### ISO 639 languages

| Source | Licence | Embed? | Rows | Fields | Format |
|---|---|---|---|---|---|
| [langs](https://www.npmjs.com/package/langs) 2.0.0 (2017, unmaintained) | MIT | Yes | 184 | name, local, 1 (639-1), 2, 2T, 2B, 3 | `data.js` array |
| [iso-639-1](https://www.npmjs.com/package/iso-639-1) 3.1.6 (2026-07-02) | MIT | Yes | 183 | code, name, nativeName | `src/data.js` |
| annexare/Countries languages | MIT | Yes | 185 | code (639-1), name, native | TS/JSON |
| Debian iso-codes 639-2 / 639-3 / 639-5 | LGPL-2.1+ | See above (LGPL) | 639-2: 487; 639-3: 7923 | 639-2: alpha_2, alpha_3, bibliographic, name, common_name; 639-3: + inverted_name, scope, type | JSON |
| [SIL ISO 639-3 tables](https://iso639-3.sil.org/code_tables/download_tables) | SIL terms: may incorporate into software (commercial or not) with attribution to iso639-3.sil.org, must not modify identifiers, and product "does not provide a means to redistribute the code set" | Risky — a git repo with the table IS a means to redistribute; avoid embedding the full table. Codes themselves are facts | 7927 | Id, Part2b, Part2t, Part1, Scope, Language_Type, Ref_Name, Comment | tab-delimited `iso-639-3.tab` |
| [Library of Congress ISO 639-2](https://www.loc.gov/standards/iso639-2/ascii_8bits.html) | US federal government work → public domain in the US | Yes | ~487 (matches Debian 639-2 count) | pipe-delimited: alpha3-B, alpha3-T, alpha2, English name, French name — UNVERIFIED: loc.gov is behind a Cloudflare bot check (WebFetch, curl and Playwright all blocked); field layout from memory | `ISO-639-2_utf-8.txt` |

- Recommendation: `langs` or `iso-639-1` (MIT) for 639-1 names; both are small and derived from public code tables. Anything 639-3-sized is LGPL or SIL-restricted.
- Rules: 639-1 `^[a-z]{2}$`, 639-2/3 `^[a-z]{3}$`; 639-2 has B/T pairs (e.g. `fre`/`fra`, `ger`/`deu`) — pick one consistently (T codes = 639-3).

### ISO 3166-2 subdivisions

| Source | Licence | Embed? | Rows | Fields |
|---|---|---|---|---|
| Debian iso-codes 3166-2 | LGPL-2.1+ | LGPL caveat | 5046 | code, name, type, parent |
| [olahol/iso-3166-2.json](https://github.com/olahol/iso-3166-2.json) | package.json says ISC; no LICENSE file; GitHub detects none; data scraped from eQuest xls (provenance unclear); last push 2021 | No — unclear licence and provenance | 237 countries / 3807 divisions | `{alpha2: {name, divisions: {code: name}}}` |
| [esosedi/3166](https://github.com/esosedi/3166) (npm `iso3166-2-db` 2.3.11) | MIT | Yes; data mixes GeoNames (CC BY 4.0), Wikipedia (CC BY-SA), OSM (ODbL) — the MIT label may not survive scrutiny for the Wikipedia/OSM-derived parts. UNVERIFIED how much is derived from each | "all countries and regions" (count not stated) | ISO 3166-2 + FIPS codes, names in 12 languages, admin level, GeoNames/OSM/Wikipedia/WOF ids |
| stefangabos subdivisions | CC BY-SA 4.0 | No | 5046 | country, code, name, name_en, type, parent |
| [US Census state.txt](https://www2.census.gov/geo/docs/reference/state.txt) | US government work → public domain | Yes | 57 (50 states + DC + 6 territories) | STATE (FIPS), STUSAB, STATE_NAME, STATENS |

- Recommendation: for an en_US library, ship only the US list (Census/USPS, public domain: 50 states + DC, optionally territories) — ISO 3166-2 code = `US-` + USPS abbreviation. Skip the ~5000-row world list; every open compilation is LGPL, CC BY-SA or unclear.
- Rule: `^[A-Z]{2}-[A-Z0-9]{1,3}$`.

## 11. Commerce

### Word lists (product name / adjective / material)

- [@faker-js/faker](https://github.com/faker-js/faker) 10.6.0 — MIT (LICENSE covers code and locale data; copyright Faker contributors 2022-2025 and Marak Squires 2011-2020; acknowledges Ruby Faker and Perl Data::Faker, both MIT). No separate data licence.
- Provenance policy in [CONTRIBUTING.md](https://github.com/faker-js/faker/blob/next/CONTRIBUTING.md): "Faker must not contain copyrighted materials"; facts (finite known lists) are OK; compilations must not be copied from a single source (Wikipedia, "most popular" articles, government sites) because "a compilation of facts can be copyrighted" — contributors are told to draw on multiple sources and their own judgement.
- Verdict: reusing faker's `en` commerce lists (`product_name.adjective/material/product`, `department`) is fine under MIT — keep the MIT notice with the copied data. The lists are short curated words (no external dataset). Nothing in the repo or docs restricts locale-data reuse beyond MIT.

### Garment and shoe sizes

- Letter sizes XS–XXL: EN 13402-3 letter codes with chest/bust ranges (Wikipedia [EN 13402](https://en.wikipedia.org/wiki/EN_13402), CC BY-SA — but the table is a handful of facts; the standard itself is paywalled). Men chest / women bust cm: XXS 70–78 / 66–74, XS 78–86 / 74–82, S 86–94 / 82–90, M 94–102 / 90–98, L 102–110 / 98–107, XL 110–118 / 107–119, XXL 118–129 / 119–131, 3XL 129–141 / 131–143.
- US numeric women's sizes (0–20, even) come from ASTM D5585 (paywalled); no open dataset. Just generate the even-number series.
- Shoe sizes: no open dataset found (search hits are ML demos and retailer charts). Use the formulas from Wikipedia [Shoe size](https://en.wikipedia.org/wiki/Shoe_size) (facts, not copyrightable): EU (Paris point) ≈ 1.5 × foot cm + 2 (i.e. 3/20 × mm + 2); UK adult ≈ 3 × foot in − 23; US men ≈ 3 × foot in − 22 = UK + 1; US women (common) = US men + 1.5 (FIA scale: 3 × foot in − 21); Mondopoint = foot length mm. ISO/TS 19407:2023 has the official conversion tables (paywalled). Generate from a foot length (men 24–31 cm, women 21–27 cm) in half-size steps so all systems agree.

### Check-digit rules (all verified against Wikipedia articles; algorithms are facts)

| Code | Rule |
|---|---|
| EAN-13 / UPC-A / GTIN-8/12/13/14 | Weights 3,1,3,1… starting from the rightmost data digit (the one next to the check digit); check = (10 − sum mod 10) mod 10. UPC-A = EAN-13 with a leading 0 (GS1 prefix of a 12-digit GTIN is `0` + first two digits). GTIN-14 first digit is an indicator 1–8 (packaging level) or 9 (variable measure); 0 is not valid there. |
| UPC-A number system (first digit) | 0,1,6,7,8,9 regular; 2 variable-weight; 3 drugs (NDC); 4 in-store/local; 5 coupons. |
| ISBN-10 | Σ(weights 10..2 × first 9 digits) + check ≡ 0 mod 11; check = (11 − sum mod 11) mod 11, 10 → `X`. |
| ISBN-13 | EAN-13 with prefix 978 or 979. 979 groups in use: 979-8 USA, 979-10 France, 979-11 Korea, 979-12 Italy. 978 groups 0 and 1 = English-language. For an en_US generator: `978-0…`, `978-1…`, `979-8…`. The official range file (`https://www.isbn-international.org/export_rangemessage.xml`) is XML only and the site says "You must not republish material from our site … without our written permission" — do not embed it; group/registrant length only matters for hyphenation, not validity. |
| ISSN | Format `NNNN-NNNC`; weights 8..2 over first 7 digits, sum mod 11; check = 0 if remainder 0 else 11 − remainder; 10 → `X`. EAN-13 form: `977` + 7 ISSN digits (no check) + 2 variant digits + EAN check. |
| IMEI | 15 digits: TAC 8 (RBI 2 + 6) + serial 6 + Luhn check (ISO/IEC 7812; from the right, double every second digit, sum digits, total ≡ 0 mod 10). SVN never enters the check. IMEISV = 14 digits + 2-digit SVN, no check. RBI `00` = test IMEI (Wikipedia [Reporting Body Identifier](https://en.wikipedia.org/wiki/Reporting_Body_Identifier)); live RBIs: 01 CTIA/PTCRB (US), 35 TÜV SÜD/BABT (UK), 86 TAF (China), 99 GHA. For fake data use TAC `00xxxxxx` so it can never match a real handset — UNVERIFIED: third-party pages claim the GSMA TS.06 test-TAC shape is `00 44` + 4 digits; the TS.06 PDF could not be text-extracted here. |
| ASIN | 10 chars `[A-Z0-9]{10}`; non-books start `B0` + 8 alphanumerics (`^B0[A-Z0-9]{8}$`); books reuse their ISBN-10 verbatim. No check digit (Wikipedia). |
| SKU | Not standardised (Wikipedia). Convention only: 8–12 uppercase alphanumerics, e.g. `[A-Z]{3}-\d{4}-[A-Z]{2}`; avoid leading 0 and O/I if you want scanner-safe. |

### GS1 prefixes

- Official table: [gs1.org/standards/id-keys/company-prefix](https://www.gs1.org/standards/id-keys/company-prefix) (rendered via Playwright; WebFetch gets 403). Copyright notice ([gs1.org/terms-use](https://www.gs1.org/terms-use)): reproduction allowed only "in unaltered form … for your personal, non-commercial use or use within your organisation"; "You are not permitted to re-transmit, distribute or commercialise the information or material without seeking prior written approval from GS1." → Do not copy the GS1 page. The prefix-to-country mapping is factual and also on Wikipedia ([List of GS1 country codes](https://en.wikipedia.org/wiki/List_of_GS1_country_codes), CC BY-SA 4.0 — avoid verbatim copy). Safe path: embed only the US/special ranges you need, written from the facts below.
- Ranges relevant to an en_US generator (from the GS1 page): 001–019, 030–039, 060–139 GS1 US (UPC-A compatible); 020–029 restricted circulation within a region; 040–049 restricted within a company; 050–059 GS1 US reserved; 200–299 restricted circulation (region); 952 "used for demonstrations and examples of the GS1 system"; 977 ISSN; 978–979 ISBN; 980 refund receipts; 981–983 coupons (common currency); 990–999 coupons. Prefixes do not identify country of origin. Selected others: 300–379 France, 400–440 Germany, 450–459 & 490–499 Japan, 500–509 UK, 690–699 China, 730–739 Sweden, 750 Mexico, 754–755 Canada, 760–769 Switzerland, 800–839 Italy, 840–849 Spain, 870–879 Netherlands, 880–881 South Korea, 890 India, 930–939 Australia, 940–949 New Zealand.
- Generator rule: for fake UPC/EAN use prefix `952` (GS1 demo) or a `2xx` restricted-circulation prefix — never a real company prefix; for US-looking UPC-A use number system 0–1/6–8 with a random 5-digit manufacturer code and accept that it may collide with real products.

## 12. Books, music, food

### Project Gutenberg catalog

- URL: `https://www.gutenberg.org/cache/epub/feeds/pg_catalog.csv` (20 MB) / `pg_catalog.csv.gz` (5.3 MB), regenerated weekly; also RDF (`rdf-files.tar.bz2`, 121 MB) and MARC.
- Fields: `Text#, Type, Issued, Title, Language, Authors, Subjects, LoCC, Bookshelves` (comma CSV, quoted; Authors as `Last, First, birth-death`, `;`-separated).
- Count (2026-09-17): 79 381 rows; Type: Text 78 130, Sound 1 114, Dataset 89, Image 33, other 12; Language: en 62 860, fr 4 190, fi 3 684, de 2 425, it 1 110.
- Licence: each RDF record carries `<cc:Work><cc:license rdf:resource="https://creativecommons.org/publicdomain/zero/1.0/"/>` → catalog metadata is CC0 1.0 (verified on `cache/epub/1/pg1.rdf`). The CSV itself has no header notice; treat as CC0 by the same source. Terms of use: do not hammer their servers (bulk feeds are the sanctioned path), "Project Gutenberg" is a trademark — royalties for commercial use of the *name*; do not brand the data file with it beyond a source citation.
- Embed: yes (CC0). For a fake-data library ship a filtered subset (Type=Text, Language=en, title + first author, maybe 5–10k rows) rather than 20 MB.

### Music genres and instruments

| Source | Licence | Embed? | Count | Notes |
|---|---|---|---|---|
| [MusicBrainz genre list](https://musicbrainz.org/genres) | Genre is a core entity (schema doc lists 13 core entities incl. Genre) → core data dump `mbdump.tar.bz2` is CC0. The genre→entity *associations* come via user tags (CC BY-NC-SA 3.0) — do not use those | Yes, names only | 2 202 (`/ws/2/genre/all?fmt=json`, `genre-count`) | WS API needs a descriptive User-Agent; list is a JSON array of `{id, name, disambiguation}` |
| [Discogs data dumps](https://data.discogs.com/) | CC0 | Yes | genre/style vocab is embedded in release XML (monthly dumps, GBs) — no standalone style list; UNVERIFIED count (~15 genres, ~600 styles from memory) | Impractical to extract; prefer MusicBrainz |
| ID3v1 genres | Names 0–79 from the 1999 ID3v1 spec, 80–125 Winamp, 126–191 Winamp 5.6 (2010) — a de-facto spec table, treated as public domain facts | Yes | 192 (0–191) | Wikipedia [List of ID3v1 genres](https://en.wikipedia.org/wiki/List_of_ID3v1_genres) (CC BY-SA page; the list itself is a spec enumeration). id3.org returned HTTP 500 during research — UNVERIFIED against the original spec page |
| Wikidata instruments | CC0 | Yes | 9 194 items `P31/P279* Q34379` with English labels; 2 372 with a Hornbostel–Sachs number (P1762) | Filter by P1762 for a clean orchestral/folk instrument list |
| Wikidata music genres | CC0 | Yes | 6 619 (`Q188451`) | Noisier than MusicBrainz |

### USDA FoodData Central

- Downloads: [fdc.nal.usda.gov/download-datasets](https://fdc.nal.usda.gov/download-datasets/) (April 2026 release):
  - Foundation Foods: JSON 459 KB zip / 6.5 MB; CSV 3.7 MB zip / 32 MB — 394 foods (API `totalHits`, dataType=Foundation).
  - SR Legacy (final, April 2018): JSON 12.3 MB zip / 205 MB; CSV 6.7 MB zip / 54 MB — 7 793 foods.
  - FNDDS 2021-2023: CSV 200 MB zip / 1.6 GB — 5 432 foods.
  - Branded: CSV 428 MB zip / 2.9 GB — 433 403 foods.
  - Full: CSV 460 MB zip / 3.1 GB.
- Licence ([API guide](https://fdc.nal.usda.gov/api-guide/)): "USDA FoodData Central data are in the public domain and they are not copyrighted. They are published under CC0 1.0 Universal (CC0 1.0)". No permission needed; USDA *requests* listing FoodData Central as source and notifying them. Suggested citation: "U.S. Department of Agriculture, Agricultural Research Service. FoodData Central, 2019. fdc.nal.usda.gov."
- Embed: yes (CC0). Ship a derived list (SR Legacy `description` + `food_category`) not the raw dump.
- CSV structure (UNVERIFIED — the field-description PDF could not be text-extracted; from memory): `food.csv` (fdc_id, data_type, description, food_category_id, publication_date), `food_category.csv` (id, code, description), `nutrient.csv` (id, name, unit_name, nutrient_nbr, rank), `food_nutrient.csv` (id, fdc_id, nutrient_id, amount, …), `food_portion.csv`, `measure_unit.csv`, `sr_legacy_food.csv` (fdc_id, NDB_number), `foundation_food.csv`.

### Dishes, cuisines, drinks

| Source | Licence | Embed? | Count | Notes |
|---|---|---|---|---|
| Wikidata dishes (`Q746549`) | CC0 | Yes | 7 480 with English label | SPARQL `?i wdt:P31/wdt:P279* wd:Q746549` |
| Wikidata cuisines (`Q1968435`) | CC0 | Yes | 219 | |
| Wikidata cocktails (`Q134768`) | CC0 | Yes | 305 | |
| Wikidata beer (`Q44` subclass tree) / wine (`Q282`) | CC0 | Yes | 420 / 3 025 | Wine tree is mostly appellations/brands; filter by P279 depth |
| IBA official cocktails | IBA site "© IBA 2026 – All rights reserved"; Wikipedia list CC BY-SA 4.0 | Names only are facts (102 cocktails: 34 Unforgettables, 34 Contemporary Classics, 34 New Era, 2024 revision); do not copy recipes/descriptions | 102 | Safest: take the names from Wikidata (`P31 Q134768`) rather than the IBA page |
| [Open Brewery DB](https://github.com/openbrewerydb/openbrewerydb) | MIT (LICENSE, © 2025 Open Brewery DB) | Yes | 11 931 breweries (US 8 308, DE 1 445, AU 514, BE 478, CA 283, NZ 243) | `breweries.csv`: id (UUID), name, brewery_type (micro 5 906, brewpub 3 947, closed 644, planning 635, regional 240, contract 208, large 137, proprietor, taproom, bar, nano, cidery, beergarden), address_1..3, city, state_province, postal_code, country, phone, website_url, longitude, latitude. Real businesses — fine for names, but generating "fake" data with real addresses/phones may be undesirable |
| Wikipedia lists (cuisines, dishes, IBA) | CC BY-SA 4.0 | No verbatim copies | | |

## Summary of safe picks

- Countries: `datasets/country-codes` (PDDL) or `annexare/Countries` (MIT); flag emoji derived from alpha-2.
- Languages: `langs` / `iso-639-1` (MIT, 639-1 only).
- Subdivisions: US-only list from Census (public domain); skip world list.
- Commerce words: faker `en` lists (MIT). Check digits: implement from the rules above; GS1 prefix table: hand-write the few ranges needed, use `952`/`2xx` for generated barcodes; ISBN range file: do not embed.
- Books: Gutenberg `pg_catalog.csv` (CC0), filtered.
- Music: MusicBrainz genre names (CC0, 2 202), ID3v1 list (192), Wikidata instruments (CC0).
- Food: USDA FDC SR Legacy/Foundation descriptions (CC0, cite USDA), Wikidata dishes/cuisines (CC0), Open Brewery DB (MIT).
- Avoid: mledoze (ODbL), lukes (CC BY-SA), stefangabos (CC BY-SA/LGPL), Debian iso-codes (LGPL), SIL 639-3 table (no-redistribution clause), GS1/ISBN-International pages (all rights reserved), Wikipedia tables verbatim (CC BY-SA), MusicBrainz tag associations (CC BY-NC-SA).
