# Swedish (sv_SE) fake-data sources — research 2026-09-17

Scope: names, identifiers, phone, company, plates/cars, words, dates/misc, other. Geography excluded.
"Verified" = fetched and read this session. Local copies of inspected files: `/home/lilleman/.claude/jobs/e377c8d4/tmp/`.

## Licence summary (embed in MIT repo?)

| Source | Licence (exact) | Embed in MIT? | Attribution |
|---|---|---|---|
| SCB Statistikdatabasen + geodata | CC0 1.0 Universal | Yes | None required; SCB asks "Källa: SCB" for statistics taken from scb.se |
| SCB files on scb.se (xlsx: names, SNI, SSYK, SUN) | Not stated per file; scb.se terms: cite "Källa: SCB" | Yes (flag: CC0 statement names only statistikdatabasen/geodata) | "Källa: SCB" |
| Skatteverket testpersonnummer / testsamordningsnummer | CC0 1.0 (in DCAT) | Yes | None |
| Skatteverket namn på nyfödda, efternamn lists | No licence in DCAT/page; "fri att använda, inga avtal eller avgifter" (PSI law 2022:818) | Yes, flag | Cite Skatteverket |
| Bolagsverket statistics CSVs | CC BY 2.5 SE | Yes | Attribution required |
| Bolagsverket/SCB bulk company files | Not stated ("värdefulla datamängder", free, no agreement) | Derive stats only; do not embed raw | Cite |
| Bankinfrastruktur clearing CSV | Not stated; accuracy disclaimer | Facts only; flag | Cite BSAB |
| Bankgirot PDFs | "© Bankgirocentralen BGC AB. All rights reserved." (Informationsklass: Öppen) | No verbatim copy; facts only | — |
| PTS numbering plan PDF, pts.se number pages | Not stated (no licence on pts.se or in PDF) | Facts only; flag | Cite PTS |
| Transportstyrelsen blocked combinations | Not stated on page | Facts (a list of 99 codes); flag | Cite |
| SALDO morphology (Språkbanken) | CC BY 4.0 | Yes | Cite: Språkbanken (2017) SALDO's morphology, DOI 10.23695/agcm-ny22 |
| Språkbanken word statistics | CC BY 4.0 | Yes | Cite Språkbanken |
| Folkets lexikon | CC BY-SA 2.5 Generic | No (share-alike) | — |
| SAOL (svenska.se) | Copyright, written permission required | No | — |
| Hunspell sv_SE (DSSO) | MPL 1.1 / GPL 2 / LGPL 2.1 (tri); yeager/hunspell-sv LGPL-3.0 | Avoid | — |
| Mobility Sweden registrations | "ange källa" (no open licence) | Top-N facts only | Cite Mobility Sweden |
| Wikipedia (sv/en) | CC BY-SA 4.0 | No copying; facts only | — |
| CLDR (via ICU/Intl) | Unicode licence | Yes | Unicode notice |

## 1. Person names

### SCB (historic, frozen)
- SCB stopped producing name statistics from 2024 and refers to Skatteverket (Namnsök page, verified). Statistikdatabasen folder BE0001 now holds only "Äldre tabeller som inte längre uppdateras" (BE0001D newborns, BE0001G whole population).
- Whole-population xlsx (verified, downloaded): https://www.scb.se/contentassets/9fe7dbb460994c72b835163dbc491ef9/namn-med-minst-tva-barare-31-december-2022.xlsx — 13.1 MB, last-modified 2024-11-21, reference date 2022-12-31, names with >= 2 bearers, names in UPPERCASE.

| Sheet | Rows (incl. 4 header rows) | Columns |
|---|---|---|
| Efternamn | 411,802 | Efternamn, Antal bärare |
| Förnamn kvinnor | 91,247 | Förnamn, Antal bärare |
| Förnamn män | 79,128 | Förnamn, Antal bärare |
| Tilltalsnamn kvinnor | 57,785 | Tilltalsnamn, Antal bärare, Medelålder (only where >= 10 bearers, else "-") |
| Tilltalsnamn män | 49,404 | same |

- Gender split: a name appearing in both kvinnor/män sheets gives P(female) = count_k / (count_k + count_m).
- Middle names: "Förnamn" = every given name a person carries; "Tilltalsnamn" = the one used. Förnamn-minus-tilltalsnamn frequency approximates middle-name usage. Legal "mellannamn" (a second surname) is a different concept, no longer grantable under namnlagen 2016:1013 (not verified this session).
- Junk rows to filter: single letters ("A"), initials ("A-C", "K C", "JR"), "A:SON".
- Legacy PxWeb tables (folder BE0001G, verified via web UI): BE0001T06AR, BE0001TNamn10 (tilltalsnamn >= 10 bearers 1999–2020), BE0001T100, BE0001T08AR, BE0001T07AR, BE0001FNamn10 (förnamn >= 10 bearers), BE0001F100, BE0001T03Ar (efternamn top), BE0001ENamn10 (efternamn >= 10 bearers 1999–2020). PxWebApi v1 (`api.scb.se/OV0104/v1/doris/sv/ssd/START/BE/BE0001/BE0001G/...`) returns HTTP 400 and PxWebApi 2 (`/v2beta/api/v2/tables/BE0001T06AR`) returns "Non-existent table" — these tables are web-UI only now (verified). Use the xlsx.
- Licence: SCB terms page (verified): "Statistik och geodata som SCB tillgängliggör som öppna data i statistikdatabasen och i vår geodataplattform har licensen Creative commons 0 1.0 Universal, CC0" and "Om du använder eller sprider statistik från scb.se ska dock som källa alltid SCB anges". The xlsx sits on scb.se, not in statistikdatabasen — flag; treat as CC0 + "Källa: SCB".

### Skatteverket (current)
- Namn på nyfödda (2021+), annual: DCAT https://skatteverket.entryscape.net/store/26/metadata/416; CSV `fbf_namn_nyfodda.csv` https://skatteverket.entryscape.net/store/26/resource/417; JSON API https://skatteverket.entryscape.net/rowstore/dataset/da2556d0-c717-45e8-a1d8-3320161d3a7d/json?_limit=100&_offset=0 (47,529 rows, verified). Columns: rangordning, kön (Kvinna/Man), fodelsear, uppdateringsdatum, namn, gruppering (Kommun/…), grupperingsvärde (e.g. ALE), antal. Names in Title case. No licence field in DCAT — flag.
- Most common surnames (> 2,000 bearers, ~500 names), 2026 text file (verified, 517 lines, format `Andersson 211808`): https://www.skatteverket.se/download/18.70685bee19c85dd5dd03b59/1775049117448/Fria_efternamn_namn_antal_textfil_2026.txt (also two PDFs). Page: https://www.skatteverket.se/privat/folkbokforing/namn/bytaefternamn/sokblanddevanligasteefternamnen.4.515a6be615c637b9aa48e09.html. Licence not stated.
- Skatteverket general open-data terms (verified): "Vår öppna data är fri att använda och kräver varken några avtal eller innebär några avgifter" under lag 2022:818; no CC licence named except on the test-number datasets (CC0).
- Skatteverket publishes no whole-population first-name list as a file; its statistikportalen "Så många har ett visst namn" is an interactive search (page 404 on fetch; unverified).

Recommendation: embed SCB 2022 xlsx-derived lists (top-N per sheet with counts) + Skatteverket 2026 surname counts for freshness.

## 2. Identifiers

### Personnummer (Folkbokföringslagen 1991:481 18 §, verified via lagen.nu)
- 10 digits: YYMMDD + födelsenummer (3 digits) + kontrollsiffra. Födelsenummer odd = male, even = female. Separator "-" ; "+" once the person turns 100. 12-digit form YYYYMMDDNNNC (no separator).
- Check digit: Luhn/mod 10 over the 9 digits YYMMDDNNN (weights 2,1,2,1,… from the left; sum digits of two-digit products; check = (10 − sum mod 10) mod 10).
- Before 1990 the first two födelsenummer digits encoded county; now random (Wikipedia, unverified officially).

### Samordningsnummer
- Same layout; day + 60 (3rd → 63); individual number random 001–999, odd male / even female; same Luhn (Skatteverket rättslig vägledning, via search excerpt; page itself blocked). Now regulated by Lag (2022:1697) om samordningsnummer (18 a § FBL repealed) — flag.
- Example from Skatteverket: man born 1970-10-03, no. 239 → 701063-2391.

### Official test numbers (Skatteverket, CC0 1.0)
- Testpersonnummer dataset: https://www.dataportal.se/sv/datasets/6_67959/testpersonnummer (DCAT: https://admin.dataportal.se/store/6/metadata/67959?recursive=dcat). "Testpersonnummer med sekelsiffror för åren 1890-2025. Nya testnummer för kommande år läggs ut i december varje år." Formats: Excel (one file), CSV (split into many period files), JSON API https://skatteverket.entryscape.net/rowstore/dataset/b4de7df7-63c0-4e7e-bb59-1f156a591763/json?_limit=100 — resultCount 43,895 (verified), one column `testpersonnummer`, 12 digits.
- Observed pattern (verified on 1890s CSV file fully and API samples at offsets 0/10000/30000/43880): every day gets two numbers, födelsenummer 238 (female) and 239 (male), through 2026-12-31; 1890s use 980/981 (plus a few 982/983/936). All Luhn-valid. Skatteverket blocks these from ever being assigned.
- Testsamordningsnummer: DCAT https://skatteverket.entryscape.net/store/9/metadata/153; years 1914–2025; CSV per year; sample file 2022 (1,498 rows): days 61–91, individual numbers 238/239, Luhn-valid.
- Should a generator restrict to these? Skatteverket does not mandate it, but only these are guaranteed never to collide with a real person; any other Luhn-valid number may be real. Recommended: default to the official series (date + 238/239, or 980/981 for 1890s) and embed nothing but the rule (~530 kB if the full list were embedded). Flag: the 238/239 rule was sampled, not exhaustively verified for every year.
- Common practice elsewhere (Faker etc.): random date + random 3 digits + Luhn — valid but collision-prone.

### Organisationsnummer
- 10 digits NNNNNN-NNNN, last = Luhn over first 9 (same algorithm as personnummer). Digits 3–4 ("month") are always >= 20, distinguishing from personnummer (Wikipedia; not found on an official page — flag).
- First digit (gruppnummer), Bolagsverket page (verified): 5 = aktiebolag, filialer, banker, försäkringsbolag, europabolag; 9 = handelsbolag, kommanditbolag; 7 or 8 = bostadsrättsföreningar, ekonomiska föreningar, näringsdrivande ideella föreningar etc.; 2 or 8 = trossamfund; 20… = state agencies (Bolagsverket itself 202100-5489); 3 = foreign companies. Wikipedia adds: 1 = dödsbon, 2 = stat/region/kommun, 6 = samfällighetsföreningar, 8 = ideella föreningar/stiftelser. Enskild firma = owner's personnummer.
- VAT number (Skatteverket, verified): "SE" + 12 digits = organisationsnummer (10) + "01" ("De två sista siffrorna är alltid 01"). Written SE556047352101.

### Bankgiro / Plusgiro (Bankgirot "10-modul" PDF 2016-12-01, verified)
- Bankgironummer: 7 or 8 digits, last digit mod-10 (Luhn) check; printed 991-2346 (7) / 5555-5551 (8), i.e. hyphen before the last four digits. 90-konton (charity) 900-000x…904-999x, always 7 digits (Wikipedia).
- Plusgironummer: 2–8 digits, last digit Luhn (Wikipedia/samlogic; not from an official Nordea page — flag). Printed with hyphen before check digit, e.g. "12 34 56-7" (convention, unverified).

### Clearing numbers / bank accounts
- Bankinfrastruktur i Sverige AB CSV (verified, 57 rows, 40 actors): https://www.bankinfrastruktur.se/media/1melztro/tabell-over-clearingnummer-250305.csv (also `/media/sqfjsvkp/clearingnummertabell-for-nedladdning.csv`, 63 rows, newer, includes Zimpler 2130-2139). Columns: `Clearingnummer;Aktör;BIC;IBAN ID;Konto-typ;Metod IBAN konvertering`. Semicolon, UTF-8 BOM. Page: https://www.bankinfrastruktur.se/framtidens-betalningsinfrastruktur/iban-och-svenskt-nationellt-kontonummer. No licence; disclaimer "garanterar inte att publicerade uppgifter är korrekta".
- Key ranges: Nordea 1100–1199, 1400–2099, 3000–3399, 3410–3999, 4000–4999 (3300 and 3782 = personkonto); Danske 1200–1399, 2400–2499; SEB 5000–5999, 9120–9124, 9130–9149; Handelsbanken 6000–6999; Swedbank 7000–7999 (4-digit) and 8000–8999 (5-digit clearing with own check digit); Länsförsäkringar 3400–3409, 9020–9029, 9060–9069; Skandiabanken 9150–9169; ICA 9270–9279; Avanza 9550–9569; Nordnet 9100–9109; Klarna 9780–9789.
- Bankgirot "Bankernas kontonummer" 2024-02-22 (verified text): Typ 1 = 4-digit clearing + 7-digit account (11 digits) with mod-11 check, weights 1,10,9,…,1; comment 1 = weigh clearing minus first digit + 7 digits; comment 2 = whole clearing + 7 digits. Typ 2 = clearing not part of the account: Handelsbanken 9-digit account mod-11 (comment 2); Swedbank 8000–8999 up to 10 digits mod-10 (comment 3); Danske 9180–9189, Nordea personkonto 3300/3782, Sparbanken Syd 9570–9579 10 digits mod-10 (comment 1).
- IBAN SE (Bankinfrastruktur page, verified): 24 chars = "SE" + 2 check digits + 3-digit IBAN ID (from CSV column, e.g. Nordea 300, SEB 500, Handelsbanken 600, Swedbank 800, Danske 120, LF 902) + 17-digit account (zero-padded; clearing included/excluded per "Metod IBAN konvertering" 1–3).

## 3. Phone

- Source: PTS "The Swedish numbering plan for telephony according to ITU-T E.164", 2024-01-08 (verified, text extracted): https://pts.se/globalassets/globala-block/nummertillstand/the-swedish-numbering-plan-for-telephony-according-to-itu---2024-01-08.pdf. Swedish "nrplansammanstallning-2026-02-04.pdf" and https://nummer.pts.se/NbrPlanSearch are behind a Radware bot check (both curl and Playwright blocked). No licence stated by PTS anywhere found; pts.se says contact pts@pts.se for data.
- Reserved for fiction, PTS "Telefonnummer till böcker och filmer" (verified 2026-09-21; curl and Playwright are Radware-blocked, a rendering fetcher gets through): https://www.pts.se/internet-och-telefoni/telefonnummer-och-adressering/telefonnummer-till-bocker-och-filmer/ — mobile 070-1740605–1740699, geographic 08-46500400–46500499 (Stockholm), 031-3900600–3900699 (Göteborg), 040-6280400–6280499 (Malmö), 0980-319200–319299 (Kiruna). 495 numbers PTS holds back so they reach nobody, which is goal 14's never-reaching path for `sv_SE.phone`. Same licence status as the numbering plan: none stated, facts only, cite PTS. The mobile block fits the shipped `{a} {b} {c}` grouping (`174 06 05`–`99`), as do 031 and 040; Stockholm's 8 subscriber digits and Kiruna's 6 need a format of their own, and 0980 is not among the 20 shipped area codes.
- 264 geographic area codes (PDF text yields 263 "Area code for" rows; one likely split across a page break — flag). Area code = trunk "0" + NDC of 1–3 digits (08 Stockholm, 031 Göteborg, 040 Malmö, 0480 Kalmar). Geographic N(S)N (NDC + subscriber) max 9 digits, min 7 (2-digit NDC) or 8 (3-digit NDC); Stockholm min 7. Each row in the PDF gives NDC, max/min length, "Area code for <name>", so a `{ndc, name}` list can be built from it.
- Mobile: NDC 70, 72, 73, 76, 79 — N(S)N exactly 9 digits, i.e. 07X-XXX XX XX (10 digits with trunk 0). Others: 71 mobile broadband/M2M (13 digits), 74 paging, 75 personal numbering, 77 shared cost, 20 freephone, 10 location-independent (10 AXX XX XX, A=1–8), 378 M2M fixed (10 digits).
- Formatting (Språkrådet/Isof frågelådan, via search excerpts): hyphen after area code, subscriber digits grouped 2–3 from the left: `08-668 01 50`, `031-123 45 67`, `0480-123 45`, `070-123 45 67` (also `0701-23 45 67` accepted). International: `+46 70 123 45 67` (drop trunk 0).
- Wikipedia "Lista över riktnummer i Sverige" cites the same PTS document (CC BY-SA — do not copy).

## 4. Company

- Legal forms + suffixes (Bolagsverket, lag 2018:1653 om företagsnamn): aktiebolag must contain "aktiebolag" or "AB" (verified); public companies add "(publ)"; handelsbolag "HB" / "handelsbolag", kommanditbolag "KB" / "kommanditbolag", ekonomisk förening "ekonomisk förening" / "ek. för.", enskild firma no suffix — statutory wording for HB/KB/ek.för. not verified on the fetched page (flag).
- Legal-form codes (SCB "juridisk form", from Skatteverket organisationsnummer page and SCB variabelbeskrivning): 10 fysiska personer, 21 enkla bolag, 22 partrederier, 23 värdepappersfonder, 31 handelsbolag/kommanditbolag, 32 gruvbolag, 41 bankaktiebolag, 42 försäkringsaktiebolag, 43 europabolag, 49 övriga aktiebolag, 51 ekonomiska föreningar, 53 bostadsrättsföreningar, 54 kooperativ hyresrättsförening, 61 ideella föreningar, 62 samfälligheter, 63 registrerade trossamfund, 71 familjestiftelser, 72 övriga stiftelser, 81 statliga enheter, 82 kommuner, 84 regioner, 87 offentliga korporationer, 88 hypoteksföreningar, 91 oskiftade dödsbon, 92 ömsesidiga försäkringsbolag, 93 sparbanker, 94 understödsföreningar, 95 arbetslöshetskassor, 96 övriga svenska juridiska former, 98 utländska juridiska personer — codes beyond 49 quoted from memory of the SCB list, verify against https://www.scb.se/contentassets/8a8eb5c3d45f461ea93482f8e8d4de4f/variabelbeskrivning-api.pdf (flag).
- Distribution (Bolagsverket ftgstat_oppna.csv, CC BY 2.5 SE, 404,779 rows, latin-1, verified): https://www.bolagsverket.se/statistik/ftgstat_oppna.csv. Columns: ar, manad, handelse (1 = nyregistrerade, 2 = alla registrerade, 3 = avslutade), regfam, län/kommun, then counts per form: AB, BAB (bankaktiebolag), BF (bostadsförening), BRF, EK (ekonomisk förening), E (enskild näringsidkare registered at Bolagsverket only), SE (europabolag), FL (filial), FAB (försäkrings-AB), HB, I (näringsdrivande ideell förening), KB, KHF, MB, SF, SB (sparbank), TSF (trossamfund), BFL, OFB, SCE, S, EGTS, FOF, TPAB, OTPB, TPF. New registrations 2025 (computed): AB 50,859; E 8,468; HB 1,557; BRF 398; EK 319; FL 269; KB 176; I 59. Column description xlsx: https://www.bolagsverket.se/download/18.46f4138717c599ee403aac05/1638951733112/beskrivning-av-csv-fil.xlsx.
- Company-name register: Bolagsverket "värdefulla datamängder" bulk file https://vardefulla-datamangder.bolagsverket.se/bolagsverket/bolagsverket_bulkfil.zip (250 MB, updated 2026-09-14; org.nr, names, legal form, business description, address) and SCB bulk https://vardefulla-datamangder.bolagsverket.se/scb/scb_bulkfil.zip (71 MB; adds SNI up to 5 codes, status). Free, no agreement; licence not stated on the page (statistics CSVs are CC BY 2.5 SE) — flag. Use offline to derive name-token frequencies; do not embed rows. API (org.nr lookup only; name search "planned, no date"): https://bolagsverket.se/apierochoppnadata/hamtaforetagsinformation/apiforatthamtaforetagsinformation.3988.html.
- Naming rules (Bolagsverket "Välja företagsnamn", verified): must be distinctive — not only a generic activity ("Bilverkstad AB") or a bare surname; no confusion with existing names/trademarks; no professional titles without credentials; no apparent internet address; "aktiebolag/AB/handelsbolag" ignored when comparing names. Typical patterns (unverified heuristic, derive from bulk file): `<Surname> <Trade> AB`, `<Surname>s <Trade>`, `<Place> <Trade> AB`, `<Fantasy word> AB`, `<Initials> <Trade> HB`, suffix words Bygg, Konsult, Holding, Invest, Fastigheter, Förvaltning, Entreprenad, Teknik, Design.
- SNI 2007 (SCB, verified xlsx, 94 kB): https://www.scb.se/contentassets/d43b798da37140999abf883e206d0545/sni2007.xlsx — sheets Detaljgrupp (5-digit, 821 codes), Undergrupp (615), Grupp (272), Huvudgrupp (88), Avdelning (21 letters). Code format `01.110` + Swedish label. SNI 2025 replaces it from 2025-12-08 (835/651/287/87/22): https://www.scb.se/globalassets/sni-2025.xlsx (93 kB). Skatteverket/Bolagsverket transition status unverified. Licence: no statement on the classification pages; SCB terms apply ("Källa: SCB") — flag.

## 5. Licence plates and cars

- Format: `ABC 123` (since 1973) and `ABC 12A` (since 2019-01-16; last character a letter, never O). Letters never used: I, Q, V, Å, Ä, Ö. Digits 001–999 (000 not issued). Personal plates 2–7 characters, may use any letters, must not mimic the standard pattern. Source: sv.wikipedia Registreringsskyltar i Sverige (facts; not confirmed on a Transportstyrelsen page — flag).
- Blocked letter combinations, Transportstyrelsen (verified, page updated 2026-03-31, 99 codes): https://www.transportstyrelsen.se/sv/vagtrafik/fordon/aga-kopa-eller-salja-fordon/registreringsskyltar/byte-av-registreringsnummer/Sparrade-bokstavskombinationer/
  APA ARG ASS BAJ BSS CUC CUK CUM DUM ETA ETT FAG FAN FEG FEL FEM FES FET FNL FUC FUK FUL GAM GAY GEJ GEY GHB GUD GYN HAT HBT HKH HOR HOT KGB KKK KUC KUF KUG KUK KYK LAM LAT LEM LOJ LSD LUS MAD MAO MEN MES MLB MUS NAZ NRP NSF NYP OND OOO ORM PAJ PKK PLO PMS PUB RAP RAS ROM RPS RUS SEG SEX SJU SOS SPY SUG SUP SUR TBC TOA TOK TRE TYP UFO USA WAM WAR WWW XTC XTZ XUK XXL XXX ZEX ZOG ZPY ZUG ZUP ZOO
- Car makes, Mobility Sweden 2025 full-year new passenger cars (272,987 total; press release 2026-01-02 "Svagt fordonsår 2025", via search excerpt; "ange källa" required): Volvo 48,961; Volkswagen 38,677; Toyota 22,189; Kia 19,922; Mercedes-Benz 17,372; Skoda 16,664; BMW 15,001; Audi 14,887; Peugeot 9,059; Polestar 7,601. Top models 2025: Volvo XC60 17,933, Volvo EX40, VW ID.7, Tesla Model Y 5,822 (press excerpts). Monthly xlsx: https://mobilitysweden.se/statistik/Nyregistreringar_per_manad_1. Not an open licence — embed as a curated list with attribution.
- Transportstyrelsen fordonsstatistik (open data, by vehicle type only, not make): https://www.transportstyrelsen.se/sv/om-oss/statistik-och-analys/statistik-inom-vagtrafik/fordonsstatistik/. Trafa (https://www.trafa.se/vagtrafik/fordon/) publishes by owner/technical/region, no make/model list found; licence not stated.

## 6. Words

- SALDO morphology (recommended): https://sprakbanken.se/en/resources/saldom — CC BY 4.0, 128,036 entries, LMF XML `saldom.xml` 254,107,385 bytes (https://svn.spraakbanken.gu.se/sb-arkiv/pub/lmf/saldom/saldom.xml, 2017-09-19). Fields: lemma, POS, inflection paradigm, full word forms with MSD. Derive a small `{lemma, pos}` list (nouns/adjectives/verbs) offline; embed with the citation "Språkbanken (2017). SALDO's morphology. https://doi.org/10.23695/agcm-ny22".
- Frequencies: Språkbanken word statistics https://sprakbanken.se/resurser/ordstatistik — `stats_all.txt.zip` 763.87 MB, CC BY 4.0, 2025-04-22, aggregated over modern corpora (Korp). Column layout not confirmed (FAQ entry not found; per-corpus files are tab-separated, 6 columns incl. word form, POS, lemgram, SALDO sense, compound flag, frequency) — flag; inspect a per-corpus stats file first (smaller).
- Folkets lexikon (KTH): CC BY-SA 2.5 — share-alike, not suitable for an MIT-embedded derivative. SAOL: svenska.se "får inte användas … utan skriftligt tillstånd" — unusable. Hunspell sv_SE/DSSO: MPL/GPL/LGPL — avoid.

## 7. Dates, money, email, domains

- CLDR sv-SE (verified with Node 24.18.1 ICU): short date `2026-09-17`; long `17 september 2026`; full `torsdag 17 september 2026 kl. 16:30`; months januari…december (lowercase), short `jan. feb. mars apr. maj juni juli aug. sep. okt. nov. dec.`; weekdays måndag…söndag, short `mån tis ons tors fre lör sön`; currency `1 234,56 kr`; number `1 234 567,891`. ICU uses U+00A0 (NBSP) as group separator; typographic guidance says hard space, decimal comma.
- Språkrådet (search excerpts): ISO 8601 `ÅÅÅÅ-MM-DD` in formal/technical text, `6 juli 2022` in running text; clock `14.30` (period; colon accepted); amounts `1 234,50 kr`.
- Public holidays, Lag (1989:253) om allmänna helgdagar (verified): nyårsdagen 1/1, trettondedag jul 6/1, långfredagen, påskdagen, annandag påsk, Kristi himmelsfärdsdag (6th Thursday after Easter), pingstdagen, nationaldagen 6/6, midsommardagen (Saturday 20–26 June), alla helgons dag (Saturday 31 Oct–6 Nov), juldagen 25/12, annandag jul 26/12, plus Sundays. Semesterlagen 3 a § (verified): "Med söndag jämställs allmän helgdag samt midsommarafton, julafton och nyårsafton". Easter needs computing (Gregorian).
- Email domains: no official statistic found (Internetstiftelsen "Svenskarna och internet" does not report providers). Common providers by general knowledge only (unverified): gmail.com, hotmail.com, outlook.com, icloud.com, yahoo.com/yahoo.se, live.se, telia.com (Telia is phasing out its mail service), bredband.net, comhem.se, spray.se, me.com. Flag as curated.
- TLDs (Internetstiftelsen registrar stats JSON, verified): .se active domains 1,475,495 (2026-07); .nu 203,940 (2026-07). https://internetstiftelsen.se/registrar-stats/se/activedomainsmonthend.json, `/nu/…`. Licence not stated.

## 8. Other

- Occupations, SSYK 2012 (SCB, verified xlsx): https://www.scb.se/contentassets/0c0089cc085a45d49c1dc83923ad933a/ssyk-2012-koder.xlsx (50 kB; sheets 1/2/3/4-siffer; 429 four-digit codes) and job-title index https://www.scb.se/contentassets/0c0089cc085a45d49c1dc83923ad933a/ssyk_index_webb_2026.xlsx (213 kB, 2026-03-18, thousands of yrkesbenämningar → code). Licence: SCB terms ("Källa: SCB") — flag as with SNI.
- Education, SUN 2020 (SCB): https://www.scb.se/contentassets/aeeedec0e28c465aa524429407dcd5ba/sun-2020_niva_inriktning2.xlsx (304 kB; levels + orientations). Population 25–64 (SCB "Befolkningens utbildning 2024", via search excerpt): ~10 % förgymnasial, ~40 % gymnasial, ~31 % eftergymnasial >= 3 years (remainder shorter post-secondary). Approximate — flag.
- Blood groups (geblod.nu 2007 via sv.wikipedia Blodgruppsfördelning; geblod.nu today only says RhD+ ≈ 85 %): A+ 37, O+ 32, B+ 10, AB+ 5, A− 7, O− 6, B− 2, AB− 1 (%).
- Regional/kommun data: out of scope (geography).

## Unverified / flagged items
1. CC0 applies by SCB's statement to statistikdatabasen/geodata; xlsx files on scb.se (names, SNI, SSYK, SUN) carry only "Källa: SCB" terms.
2. Skatteverket name datasets and surname file: no licence field; only "fri att använda" under PSI law.
3. Testpersonnummer 238/239 rule: sampled (1890s file fully, offsets 0/10k/30k/43.9k), not exhaustively checked; per-period CSV list has many files.
4. Organisationsnummer "digits 3–4 >= 20" and group digits 1/6 — Wikipedia only.
5. Plusgiro length/format — Wikipedia/blog only.
6. PTS: no licence; 264 vs 263 area codes counted in the PDF text; Swedish 2026 PDF and nummer.pts.se blocked (Radware).
7. Plate format details (000 excluded, O excluded as last letter, 2019-01-16) — Wikipedia only.
8. Legal-form codes above 49 — from memory; verify against the SCB variabelbeskrivning PDF.
9. Word-statistics column layout; email-provider list; company-name patterns; education shares — heuristic/approximate.
