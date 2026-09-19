# Open data sources and format rules — US/locale-neutral fake-data categories

Researched 2026-09-17 for embedding as data files in an MIT-licensed library. Geography/addresses and Swedish data are out of scope (covered elsewhere). Each item that was not read first-hand from the cited page/file is marked **UNVERIFIED**.

## Licence key (what each licence means inside an MIT repo)

| Licence | Embed? | Obligation |
|---|---|---|
| Public domain (US federal work, 17 USC §105), CC0 1.0, PDDL, Unlicense | Yes | None; cite the source in the data file header as courtesy |
| MIT, BSD-2, ISC | Yes | Keep the copyright + permission notice next to the copied data |
| Unicode License v3 (OSI-approved 2023-11-17, verified) | Yes | Notice in LICENSE/NOTICE or docs |
| WordNet License (verified at wordnet.princeton.edu) | Yes | Notice + disclaimer on all copies; no Princeton name in advertising |
| W3C Document License 2023 | Yes (data tables) | Notice "includes material copied from or derived from <title> <URI>" |
| CC BY 3.0 / 4.0 | Yes | Creator, licence link, indicate changes (verified deed) |
| Apache-2.0 (code), LGPL-2.1/3.0 (data) | Awkward | LGPL data files stay LGPL; keep them out of an MIT repo, use as reference for facts only |
| CC BY-SA, ODbL, CC BY-NC-SA, GFDL | No | Share-alike / non-commercial — incompatible |
| No licence stated, "all rights reserved", proprietary directories (SWIFT BIC, Fed RTN, CUSIP, ICAO 8643, GS1 page, ISO OBP, SAE WMI, whatismybrowser, useragents.me, MySQL manual, Pantone) | No | Generate synthetically from the format rules instead |

## Embed verdict at a glance

| Category | Use (clean) | Use with notice | Avoid |
|---|---|---|---|
| 1 Names/IDs | SSA baby names (CC0), Census 2010 surnames (PD), IRS EIN prefixes / ITIN ranges, SSA SSN rules | — | — (prefix/suffix lists: hand-curate) |
| 2 Finance | ISO 4217 via datasets/currency-codes (PDDL) or SIX list-one.xml; algorithms (ABA, Luhn, ISIN, CUSIP, IBAN mod 97, EIP-55, Bech32) | npm currency-codes (MIT), GLEIF BIC-LEI (custom notice), CLDR currency symbols (Unicode-3.0) | Fed routing directory, SwiftRef BIC directory, real CUSIP lists, php-iban (LGPL) verbatim |
| 3 Company | NAICS 2022 (PD), SEC company_tickers.json (free reuse), GLEIF ELF list (CC0) | — | — |
| 4 Internet | All IANA registries (CC0): media types, HTTP status, ports, TLDs, IPv4/IPv6 special-purpose; RFC 5737/3849/9637 ranges; locally-administered MAC | mime-db (MIT), Linguist languages.yml (MIT), top-user-agents (MIT) / user-agents (BSD-2), IEEE OUI (IEEE "does not assert copyright" — second-hand) | useragents.me, whatismybrowser, MySQL manual tables |
| 5 Vehicle | NHTSA vPIC makes/models/WMIs/value lists (PD); VIN check-digit rule | — | SAE WMI DB, openalpr plate patterns (AGPL), Wikipedia lists verbatim |
| 6 Airline | OurAirports (PD/Unlicense), Wikidata airlines + aircraft types (CC0) | OPTD airlines (CC BY 4.0) | OpenFlights + npm mirrors (ODbL), ICAO Doc 8643 + scraped mirrors, OpenSky aircraft DB |
| 7 Science | PubChem periodic table (PD), NASA planetary fact sheet (PD), Wikidata animals/plants/moons (CC0), Stanford/AABB blood types (facts) | BIPM SI brochure (CC BY 4.0), GBIF backbone (CC BY 4.0 + citation), faker-js lists (MIT) | Bowserinator periodic table (CC BY-SA 3.0) |
| 8 Text | Moby POS (PD), 12dicts PD lists, XKCD colours (CC0), Lorem ipsum/Cicero, Bartlett via Gutenberg (PD) | WordNet 3.x (WordNet License), Open English WordNet (CC BY 4.0), SCOWL (MIT-like), CSS named colours (W3C notice), Google Books ngrams (CC BY 3.0) | wordfreq data (CC BY-SA), hermitdave (CC BY-SA), Norvig count_1w (LDC-derived), Wikiquote |
| 9 Dates | IANA tzdb (PD) | CLDR cldr-dates-full (Unicode-3.0) | — |
| 10 ISO | datasets/country-codes (PDDL), LoC ISO 639-2 (PD, UNVERIFIED layout), Census state.txt (PD) | annexare/Countries, i18n-iso-countries, langs, iso-639-1 (all MIT) | mledoze (ODbL), lukes + stefangabos (CC BY-SA), Debian iso-codes (LGPL), SIL 639-3 table (no-redistribution clause), olahol iso-3166-2 (unclear) |
| 11 Commerce | Check-digit algorithms (EAN/UPC/GTIN, ISBN-10/13, ISSN, IMEI Luhn); GS1 demo prefix 952 | faker-js en commerce word lists (MIT) | GS1 prefix page + ISBN range file verbatim, Wikipedia tables verbatim |
| 12 Books/music/food | Gutenberg pg_catalog.csv (CC0), MusicBrainz genre names (CC0), ID3v1 genres, Wikidata instruments/dishes/cuisines/cocktails (CC0), USDA FoodData Central (CC0) | Open Brewery DB (MIT) | MusicBrainz tag associations (CC BY-NC-SA), IBA site text |

## Corrections to the brief's assumptions

- Bowserinator/Periodic-Table-JSON is CC BY-SA 3.0 (not MIT) and has 119 rows; use PubChem CSV (118, PD).
- wordfreq code is Apache-2.0 and its data CC BY-SA 4.0 — avoid the data.
- CSS Color 4 has 148 named colours (rebeccapurple included).
- Gutenberg #27889 Bartlett is the 9th edition (1903); the 1919 10th is not on Gutenberg (UNVERIFIED elsewhere).
- CLDR file path is `cldr-dates-full/main/<locale>/ca-gregorian.json`; current package 48.2.0.
- Diners Club US/Canada is a Mastercard co-brand on 55; replace the loose "Maestro 50/56–58/6xxx" rule with the explicit IINs in part 2.
- Maestro/Discover/Diners/JCB/UnionPay lengths vary 12–19 — generate per the table, not a fixed 16.
- The 987-65-4320..4329 SSN advertising range is repeated by third parties but not found on any ssa.gov/POMS page (UNVERIFIED).
- DUNS dropped its mod-10 check digit in 2006 — no validation rule today.
- IEEE's OUI "does not assert copyright" statement is only recorded second-hand (Debian/FSF); no current IEEE page restates it.
- Sites blocking automation (403/Cloudflare): ssa.gov, www2.census.gov, swift.com, redcrossblood.org, loc.gov, gs1.org (WebFetch), standards-oui.ieee.org (needs browser UA). Download those manually with a browser UA.

## Recommended default picks for an en_US library

- Given names: SSA `names.zip` aggregated to top-N per sex (CC0). Surnames: Census `Names_2010Census.csv` top-N (PD).
- Countries: `datasets/country-codes` CSV (PDDL); flag emoji derived from alpha-2. Currencies: same file's ISO 4217 columns or `datasets/currency-codes`; symbols from CLDR `en` (Unicode-3.0).
- Internet: IANA registries + mime-db; UAs from templates (no data) or `top-user-agents` (MIT).
- Vehicles: vPIC `GetMakesForVehicleType/car` (195) + models per make; VIN generated with check digit.
- Airports: OurAirports rows with `iata_code` and type large/medium (4,569). Airlines: Wikidata (CC0).
- Words: WordNet 3.0 index files (with LICENSE) or Moby POS (PD) for POS-tagged words; XKCD + CSS for colours.
- Timezones: `zone1970.tab` names only; offsets via `Intl` at runtime.
- Check digits: implement from the rules in parts 1, 2 and 4 — no data needed.

---

# Part 1 — Names, finance, company: sources and rules

Researched 2026-09-17. "Embed" = ship as a data file in an MIT repo. Everything not marked UNVERIFIED was read from the cited page or file this session.

## 1. US person names and identifiers

### SSA baby names (given names by year and sex)
- Page: https://www.ssa.gov/oact/babynames/limits.html — national zip https://www.ssa.gov/oact/babynames/names.zip (7 MB), state zip https://www.ssa.gov/oact/babynames/state/namesbystate.zip (23 MB), territories 228 kB.
- Format (per SSA page + readme as mirrored by https://github.com/hackerb9/ssa-baby-names): one `yobYYYY.txt` per year 1880–2025, CSV `name,sex,count`, name 2–15 chars, sex `M`/`F`, sorted by sex then count desc. Names with < 5 occurrences in a year are excluded.
- Counts: ~100,364 distinct names, ~2.02 M rows across all years (hackerb9 snapshot; year of snapshot UNVERIFIED). Top 1000 names cover 71.5 % of 2025 births (SSA page).
- Licence: data.gov catalog entry https://catalog.data.gov/dataset/baby-names-from-social-security-card-applications-national-data lists licence **CC0 1.0** (`https://creativecommons.org/publicdomain/zero/1.0/`), publisher SSA. US federal work, no attribution required; embeddable.
- Note: ssa.gov returns 403 to non-browser user agents (curl/WebFetch); download with a browser UA or via the data.gov link.

### Census 2010 surnames
- Page: https://www.census.gov/topics/population/genealogy/data/2010_surnames.html
- Files: full set https://www2.census.gov/topics/genealogy/2010surnames/names.zip (CSV + XLSX, `Names_2010Census.csv`), top 1000 `Names_2010Census_Top1000.xlsx`, docs https://www2.census.gov/topics/genealogy/2010surnames/surnames.pdf.
- Fields (from surnames.pdf): `name, rank, count, prop100k, cum_prop100k, pctwhite, pctblack, pctapi, pctaian, pct2prace, pcthispanic`. `(S)` = suppressed percentage. Names are UPPERCASE.
- Count: 162,253 surnames occurring ≥ 100 times (covers 90.1 % of people with a recorded surname); the CSV also has a trailing `ALL OTHER NAMES` row (UNVERIFIED — www2.census.gov returned 403 to curl this session; download with a browser).
- Licence: US Census Bureau work → US federal government work, public domain in the US (17 U.S.C. §105). Census only asks for citation (https://www.census.gov/about/policies/citation.html: "U.S. Census Bureau, [Table], [Product], [Vintage], [URL], accessed on [date]"). Embeddable; cite source in the data file header.

### Name prefixes / suffixes
- No authoritative enumerated list exists. USPS Publication 28 §33 (https://pe.usps.com/text/pub28/28c3_011.htm) defines the fields: Name Prefix, First Name, Middle Name or Initial, Surname, **Suffix Title** = "maturity (e.g., JR, SR) and professional (e.g., PHD, DDS) suffixes"; example "MR. WALTER W. WITHERSPOON JR.".
- Recommendation: hand-curate. Prefixes `Mr., Mrs., Ms., Mx., Dr., Rev.`; suffixes `Jr., Sr., II, III, IV, PhD, MD, DDS, Esq.`. Rule of thumb: generational suffix (Jr/Sr/II–IV) only with a male name in traditional US usage; never combine Jr. with Sr.; at most one generational suffix.

### SSN
- Structure: `AAA-GG-SSSS` (area 3, group 2, serial 4). No check digit (Wikipedia).
- Randomization since **2011-06-25** (https://www.ssa.gov/employer/randomization.html): "Previously unassigned area numbers were introduced for assignment excluding area numbers 000, 666 and 900-999." Area no longer geographic.
- Invalid per SSA POMS RM 10201.035 (https://secure.ssa.gov/apps10/poms.nsf/lnx/0110201035): area `000`, `666`, `900–999`; group `00`; serial `0000`.
- Generator rule: area ∈ [001,899] \ {666}; group ∈ [01,99]; serial ∈ [0001,9999].
- Advertising SSNs (https://www.ssa.gov/history/ssn/misused.html): `078-05-1120` (Woolworth wallet card, 1938; > 40,000 people claimed it) and `219-09-9999` (1940 Social Security Board pamphlet). Both are structurally valid — exclude from output.
- `987-65-4320`–`987-65-4329` "reserved for advertising": repeated by searchbug.com, ssn-check.org, a Missouri state standard (https://oa.mo.gov/sites/default/files/CC-SocialSecurityNumberNamingStandardV050305.pdf, which even misprints it as 987-54-4329). **UNVERIFIED** — not found on any ssa.gov page or POMS this session; Wikipedia no longer carries it. Safe to use these ten as the library's "obviously fake" examples, but do not cite SSA for it.

### ITIN (IRS)
- IRS Publication 4757 (Rev. 6-2026), p. 5 (https://www.irs.gov/pub/irs-pdf/p4757.pdf): "All valid ITINs are nine-digit numbers in the same format as the SSN (9XX-8X-XXXX), beginning with a "9" and the 4th and 5th digits ranging from 50 to 65, 70 to 88, 90 to 92, and 94 to 99."
- Group `93` = ATIN (adoption TIN); `89` reserved (Intuit/TaxSlayer support pages, not IRS — UNVERIFIED).
- Generator rule: `9` + 2 digits + group ∈ {50–65, 70–88, 90–92, 94–99} + 4 digits (serial `0000` not documented as invalid for ITIN; avoid it anyway).

### EIN (IRS)
- Format `NN-NNNNNNN`. Page: https://www.irs.gov/businesses/small-businesses-self-employed/how-eins-are-assigned-and-valid-ein-prefixes ("The first two digits of your EIN show the IRS campus that assigned it or if you applied online").
- Valid prefixes (union of the table, 2026-09-17):
  `01 02 03 04 05 06 10 11 12 13 14 15 16 20 21 22 23 24 25 26 27 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 50 51 52 53 54 55 56 57 58 59 60 61 62 63 64 65 66 67 68 71 72 73 74 75 76 77 80 81 82 83 84 85 86 87 88 90 91 92 93 94 95 98 99`
  By campus: Andover 10,12 · Atlanta 60,67 · Austin 50,53 · Brookhaven 01–06,11,13,14,16,21–23,25,34,51,52,54–59,65 · Cincinnati 30,32,35–38,61 · Fresno 15,24 · Kansas City 40,44 · Memphis 94,95 · Ogden 80,90 · Philadelphia 33,39,41–43,46,48,62–64,66,68,71–77,85–88,91–93,98,99 · Internet 20,26,27,33,39,41,42,45–47,81–88,92,93,99 · SBA 31.
- Not valid: 00, 07, 08, 09, 17, 18, 19, 28, 29, 49, 69, 70, 78, 79, 89, 96, 97. No check digit.

## 2. Finance

### ABA routing number (RTN)
- Wikipedia https://en.wikipedia.org/wiki/ABA_routing_transit_number: 9 digits; check: `(3(d1+d4+d7) + 7(d2+d5+d8) + (d3+d6+d9)) mod 10 = 0`.
- First two digits: `00` US Government; `01–12` Federal Reserve districts (normal banks); `21–32` thrifts (historic, still valid); `61–72` non-bank payment processors/clearinghouses; `80` traveler's checks. Other prefixes unused (Wikipedia does not say so explicitly — inferred).
- Generator rule: prefix ∈ {01–12, 21–32, 61–72}, 6 random digits, compute 9th as `(10 − (3d1+7d2+d3+3d4+7d5+d6+3d7+7d8) mod 10) mod 10`.
- Open list of real RTNs: the Federal Reserve E-Payments Routing Directory (https://www.frbservices.org/resources/routing-number-directory/index.html) says "may not be sold, re-licensed, or otherwise used for commercial gain" → **not embeddable**. Generate synthetically.

### Payment card numbers (Wikipedia "Payment card number", https://en.wikipedia.org/wiki/Payment_card_number; all Luhn-checked)
| Network | IIN ranges | Length | CVV |
|---|---|---|---|
| Visa | 4 | 13, 16, 19 | 3 |
| Mastercard | 51–55, 2221–2720 | 16 | 3 |
| American Express | 34, 37 | 15 | 4 (front) |
| Discover | 6011, 644–649, 65, 622126–622925 (UnionPay co-brand) | 16–19 | 3 |
| JCB | 3528–3589 | 16–19 | 3 |
| Diners Club International | 30 (covers 300–305, 3095), 36, 38, 39 | 14–19 | 3 |
| Diners Club US & Canada | 55 (Mastercard co-brand) | 16 | 3 |
| UnionPay | 62 | 16–19 | 3 |
| Maestro | 5018, 5020, 5038, 5893, 6304, 6759, 6761, 6762, 6763 | 12–19 | 3 |
| Maestro UK | 6759, 676770, 676774 | 12–19 | 3 |
| Dankort 5019 · Mir 2200–2204 · RuPay 60,65,81,82,508 · Troy 65,9792 · UATP 1 (15) · Verve 506099–506198, 507865–507964, 650002–650027 (16/18/19) | | | |
- CVV lengths: https://en.wikipedia.org/wiki/Card_security_code ("three-digit ... Visa, Mastercard, and Discover"; "American Express is a four-digit code on the front").
- The requested "Maestro 50/56–58/6xxx" is a looser legacy rule; use Wikipedia's explicit IINs.
- Wikipedia text is CC BY-SA 4.0 — copy the *facts* (not prose) into your table; facts are not copyrightable.

### Published test cards
- Stripe: https://docs.stripe.com/testing — Visa 4242424242424242, Mastercard 5555555555554444, Amex 378282246310005, Discover 6011111111111117, Diners 3056930009020004 (14), JCB 3566002020360505, UnionPay 6200000000000005; any future expiry, any CVC (4 for Amex).
- PayPal: https://developer.paypal.com/tools/sandbox/card-testing/ — Visa 4005519200000004, 4012000033330026, 4012000077777777, 4012888888881881, 4217651111111119, 4500600000000061, 4772129056533503, 4915805038587737; Mastercard 2223000048400011; Amex 371449635398431, 376680816376961; Diners 36461510000039, 36461510000013; Maestro 6304000000000000, 5063516945005047; JCB 3636500000000260, 3636500000000989; CUP 6200680000000004, 6200680000000038.
- Reuse: neither page carries an open licence (Stripe docs are under Stripe's site terms). The numbers themselves are Luhn-valid facts, widely reproduced; embedding the numbers with a "from Stripe/PayPal test docs" note is low-risk but **UNVERIFIED** as an explicit grant. Alternative: generate Luhn-valid numbers from the IIN table above — no dependency on either vendor.

### BIC / SWIFT (ISO 9362)
- https://en.wikipedia.org/wiki/ISO_9362: 4 letters bank code + 2 letters ISO 3166-1 country + 2 alphanumeric location + optional 3 alphanumeric branch (8 or 11 chars; `XXX` = primary office). Location second char `0` = test BIC, `1` = passive participant, `2` = reverse billing.
- Generator rule: `[A-Z]{4}[A-Z]{2}[A-Z2-9][A-NP-Z0-9]([A-Z0-9]{3})?` — avoid `0`/`1` as the 8th char for a "live" BIC.
- SWIFT's BIC Directory is proprietary (SwiftRef licence, redistribution needs a separate "SwiftRef Redistribution License": https://www.swift.com/myswift/ordering/order-products-services/swiftref-redistribution-license) → **not embeddable**.
- Open alternative: GLEIF BIC-to-LEI mapping, https://www.gleif.org/en/lei-data/lei-mapping/download-bic-to-lei-relationship-files, files at https://mapping.gleif.org/api/v2/bic-lei/ (monthly zip; `LEI-BIC-20260828.zip` → `lei-bic-20260828T000000.csv`, columns `LEI,BIC`, 39,347 rows). Licence: "BIC/LEI Mapping Table License Agreement" (https://www.gleif.org/lei-data/lei-mapping/download-bic-to-lei-relationship-files/2017-12-21_annex-2_bic-to-lei-mapping-table-license-agreement_final.pdf) — a CC0-style grant "for any purpose whatsoever, including ... commercial" but **conditional on reproducing the notice** "SWIFT © and database rights [month year]. All rights reserved. This Mapping Table has been developed by SWIFT. Any use of the Mapping Table ... is subject to the BIC/LEI Mapping Table License Agreement ... For the latest BIC information and updates, always refer to www.swift.com/bic." Embeddable in an MIT repo with that notice in the data file; it is a list of real live BICs (bank names not included — join to GLEIF LEI data if names are wanted).

### ISIN (https://en.wikipedia.org/wiki/International_Securities_Identification_Number)
- 12 chars: 2-letter country + 9 alphanumeric NSIN + 1 check digit.
- Check: convert letters A=10…Z=35 (ASCII − 55) to produce a digit string, then Luhn over that string (double every second digit from the right, sum digits, check makes total ≡ 0 mod 10). Pitfall: work on the *expanded* digit string; a transposed letter pair can pass.

### CUSIP (https://en.wikipedia.org/wiki/CUSIP)
- 9 chars: 6 issuer + 2 issue + 1 check. Values: digits as-is, A=10…Z=35, `*`=36, `@`=37, `#`=38.
- Check: for i in 1..8, v = value; if i even, v ×= 2; sum += v div 10 + v mod 10; check = (10 − sum mod 10) mod 10.
- Real CUSIPs are proprietary (CGS/FactSet); generate synthetic ones only.

### IBAN (ISO 13616)
- Validation (https://en.wikipedia.org/wiki/International_Bank_Account_Number): move first 4 chars to end, letters → 10–35, integer mod 97 must equal 1; max 34 chars; check digits 02–98.
- SWIFT registry: https://www.swift.com/standards/data-standards/iban-international-bank-account-number, PDF release 101 https://www.swift.com/sites/default/files/files/iban-registry-v101.pdf (release 100 was Oct 2025); TXT also published (URL **UNVERIFIED** — swift.com returns 403/errors to automation this session). Terms: free of charge; no explicit licence text found (**UNVERIFIED**). Country count: 89 countries per Wikipedia (Dec 2024).
- Embeddable per-country length tables already extracted from the registry:
  - php-iban `registry.txt` (https://github.com/globalcitizen/php-iban, **LGPL-3.0**): pipe-separated, 121 rows (116 official + unofficial), columns `country_code|country_name|domestic_example|bban_example|bban_format_swift|bban_format_regex|bban_length|iban_example|iban_format_swift|iban_format_regex|iban_length|bban_bankid_start_offset|...|country_sepa|swift_official|...|currency_iso4217|central_bank_url|central_bank_name|membership`. LGPL data in an MIT repo is awkward; use it as a *reference* to build your own table (formats are facts) rather than copying the file.
  - ibankit-js (https://github.com/koblas/ibankit-js, **Apache-2.0**, registry v95) — same caveat.
  - Best route: own table `{country, length, bban_format}` derived from the SWIFT PDF (facts), cite SWIFT.

### Currencies (ISO 4217)
- ISO says use is free: https://www.iso.org/iso-4217-currency-codes.html — "ISO allows free-of-charge use of its country, currency and language codes from ISO 3166, ISO 4217 and ISO 639". Official list from SIX: https://www.six-group.com/dam/download/financial-information/data-center/iso-currrency/lists/list-one.xml (+ `.xls`, `list-three` historic). Fields `CtryNm, CcyNm, Ccy, CcyNbr, CcyMnrUnts`; published 2026-01-01; 280 entity rows, 178 distinct codes. No symbols. No licence text inside the XML.
- Alternatives:
  - datasets/currency-codes (https://github.com/datasets/currency-codes): **PDDL** (public domain); `data/codes-all.csv` fields `Entity, Currency, AlphabeticCode, NumericCode, MinorUnit, WithdrawalDate`; built from SIX list one + three; no symbols. Embeddable.
  - npm `currency-codes` v2.2.0 (https://github.com/freeall/currency-codes): **MIT**; fields `code, number, digits, currency, countries[]`; generated from the SIX XML; no symbols. Embeddable.
  - umpirsky/currency-list (https://github.com/umpirsky/currency-list): **MIT**; code → localised name only (311 entries in `data/en_US/currency.json`, includes historic); no symbols, no numeric codes, no decimals.
  - Debian iso-codes (https://salsa.debian.org/iso-codes-team/iso-codes): **LGPL-2.1+**; `data/iso_4217.json` has `alpha_3, name, numeric` only (179 entries); no symbols/minor units. LGPL → avoid embedding.
  - Symbols: none of the above carry them. Unicode CLDR (`common/main/en.xml` currency symbols, Unicode License, MIT-compatible with notice) is the usual source — **UNVERIFIED this session**.

### Crypto addresses
- Bitcoin (https://en.bitcoin.it/wiki/List_of_address_prefixes): P2PKH version 0x00 → leading `1`; P2SH 0x05 → leading `3`; Base58Check 25–34 chars (alphabet excludes `0OIl`; last 4 bytes = first 4 of double-SHA256 of version+payload). Testnet `m`/`n`, `2`, `tb1`.
- Bech32 (BIP-173, https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki, BSD-2-Clause): HRP `bc`, separator `1`, charset `qpzry9x8gf2tvdw0s3jn54khce6mua7l`, all-lowercase (or all-uppercase), max 90 chars; P2WPKH `bc1q…` = 42 chars, P2WSH `bc1q…` = 62 chars.
- Bech32m (BIP-350, https://github.com/bitcoin/bips/blob/master/bip-0350.mediawiki): witness v1+ (P2TR `bc1p…`, 62 chars), checksum constant `0x2bc830a3` instead of 1.
- Ethereum EIP-55 (https://eips.ethereum.org/EIPS/eip-55, CC0): `0x` + 40 hex; keccak256 of the lowercase hex (no 0x, as ASCII); for each hex letter, uppercase iff the corresponding hash nibble ≥ 8 (bit 4·i set). Digits unchanged.

## 3. Company

### NAICS 2022 (US Census)
- Page https://www.census.gov/naics/ ; files: 6-digit list https://www.census.gov/naics/2022NAICS/6-digit_2022_Codes.xlsx (82 kB; **1,012** six-digit codes, 111110…928120, verified by parsing), 2–6 digit https://www.census.gov/naics/2022NAICS/2-6%20digit_2022_Codes.xlsx, structure https://www.census.gov/naics/2022NAICS/2022_NAICS_Structure.xlsx, descriptions https://www.census.gov/naics/2022NAICS/2022_NAICS_Descriptions.xlsx, manual PDF https://www.census.gov/naics/reference_files_tools/2022_NAICS_Manual.pdf.
- Format: xlsx, column A code (numeric), column B title (trailing spaces in some titles — trim).
- Licence: no statement on the page or in the manual front matter; US federal government work → public domain in the US (17 U.S.C. §105). Embeddable; cite Census/OMB.

### SEC EDGAR company names
- `https://www.sec.gov/files/company_tickers.json`: object keyed `"0".."N"` of `{cik_str, ticker, title}`; **10,422** entries (2026-09-17). `company_tickers_exchange.json`: `{"fields":["cik","name","ticker","exchange"],"data":[...]}`.
- `https://www.sec.gov/Archives/edgar/cik-lookup-data.txt`: 40 MB, **1,059,372** lines, `NAME:CIK:` (`COMPANY NAME:0001234567:`), all filers incl. individuals — filter to the ticker list for company names.
- Access: declare `User-Agent: Company Name email@domain`, ≤ 10 req/s (https://www.sec.gov/os/webmaster-faq). Licence: "All Government-created content on sec.gov and EDGAR public filing content are free to access and reuse" (same FAQ). Embeddable; the ticker file (~800 kB) is the practical one.

### DUNS
- https://en.wikipedia.org/wiki/Data_Universal_Numbering_System: 9 digits, no significance; shown `NN-NNN-NNNN` or plain. Had a mod-10 check digit until ~Dec 2006; **dropped** since (expanded the pool by 800 M). So: any 9 digits are structurally valid today; no validation rule to enforce. Optional DUNS+4 suffix (4 alphanumerics) is user-assigned.

### Legal entity suffixes per country
- GLEIF ISO 20275 Entity Legal Forms code list v1.6 (2026-02-19): page https://www.gleif.org/en/lei-data/code-lists/iso-20275-entity-legal-forms-code-list ; CSV https://www.gleif.org/lei-data/code-lists/iso-20275-entity-legal-forms-code-list/2026-02-19-elf-code-list-v1.6.csv (758 kB; xlsx alongside).
- Columns: `ELF Code, Country of formation, Country Code (ISO 3166-1), Jurisdiction of formation, Country sub-division code (ISO 3166-2), Entity Legal Form name Local name, Language, Language Code (ISO 639-1), Entity Legal Form name Transliterated name, Abbreviations Local language, Abbreviations transliterated, Date created, ELF Status ACTV/INAC, Modification, Modification date, Reason`.
- Counts: 4,002 rows; 130 countries; 3,790 ACTV / 212 INAC; 1,518 rows have a local abbreviation (e.g. US `N.A.`, `FSA`; SE `AB`… multiple abbreviations separated by `;`). US has 737 rows (per-state forms). Abbreviations are the field you want for `Inc`, `LLC`, `GmbH`, `AB`, `SAS`, `Pty Ltd`; many rows have none, so filter on non-empty abbreviation and status ACTV.
- Licence: GLEIF Open Data page (https://www.gleif.org/en/about/open-data): "The data on GLEIF's website is provided under a Creative Commons (CC0) license"; LEI Data Terms of Use: "provided under the CC0 licence, see CC0 1.0 Universal". The ELF page itself names no licence — that CC0 explicitly covers this code list is **UNVERIFIED**, but GLEIF treats all its published data this way. Embeddable.

## Quick embed verdict
| Item | Embed? | Licence / attribution |
|---|---|---|
| SSA baby names | yes | CC0 (data.gov) |
| Census 2010 surnames | yes | US gov PD; cite Census |
| EIN prefixes, ITIN/SSN rules | yes (rules) | facts from IRS/SSA |
| NAICS 2022 | yes | US gov PD |
| SEC company_tickers.json | yes | free to reuse (SEC) |
| GLEIF ELF code list | yes | CC0 (GLEIF) |
| GLEIF BIC-LEI (real BICs) | yes, with mandatory SWIFT notice | custom permissive licence |
| ISO 4217 (SIX XML / datasets PDDL / npm MIT) | yes | free use (ISO), PDDL, MIT |
| Fed routing directory, SwiftRef BIC directory, CUSIP lists | no | restricted/proprietary — generate synthetically |
| php-iban / ibankit registries | copy facts only | LGPL-3.0 / Apache-2.0 |
| Stripe / PayPal test cards | numbers only, note source | no explicit grant (UNVERIFIED) |
# Part 2 — Internet, vehicle, airline

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
