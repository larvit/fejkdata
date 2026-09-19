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
