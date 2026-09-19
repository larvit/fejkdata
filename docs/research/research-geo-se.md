# Open data sources for real Swedish geographic and postal-address data

Researched 2026-09-16. Goal: Sweden → län → kommun → tätort/postort → street, with real postal codes, embeddable in an MIT-licensed library.

## TL;DR

| Need | Best source | Licence | Embed in MIT repo? |
|---|---|---|---|
| Län + kommun names/codes | SCB `kommunlankod-2026.xlsx` | CC0 1.0 | Yes, no attribution needed |
| Population per kommun (weighting) | SCB PxWeb API `BE0101A/BefolkningNy` | CC0 1.0 | Yes |
| Tätorter (2,017) with code, kommun, län, population | SCB open geodata WFS/GeoPackage `stat:Tatorter_2023` + PxWeb `MI0810A/LandarealTatortN` | CC0 1.0 | Yes |
| Streets + house numbers + postnummer + postort + kommun + coordinates (~3.9M points) | Lantmäteriet *Belägenhetsadress Nedladdning, vektor* (STAC, GeoPackage per kommun) | CC BY 4.0 **plus** personal-data terms, purpose review, Basic-auth download | **Not as-is** — raw redistribution is bound by the personal-data terms; a derived aggregate (street ↔ postnummer ↔ postort, no house numbers) is the realistic path but needs a stated purpose approved by Lantmäteriet. Legal review needed. |
| postnummer → postort (list only) | GeoNames `SE.zip` (18,887 rows, 1,780 postorter) | CC BY 4.0 | Yes with attribution — but source is an undated third-party file, likely stale |
| Street names per kommun/postort without Lantmäteriet | Trafikverket NVDB *Gatunamn* (Lastkajen) | CC0 1.0 | Yes, but no postnummer; needs account to download |
| Street names | OpenStreetMap | ODbL 1.0 | Effectively no (share-alike on the extracted database) |
| Municipal address files | OpenAddresses `sources/se/*` (16 kommuner) | Mixed: CC0 (Göteborg, Malmö), CC BY 4.0 (Helsingborg), unspecified (Stockholm 2016 dump, Uppsala) | Per source |
| Postal code register (authoritative) | PostNord → Postnummerservice Norden AB / Geposit AB | Commercial; "resale/sublicensing not permitted" | No |

## 1. Lantmäteriet

Since 2025-02-03 the EU High Value Datasets (HVD) products are fee-free but Lantmäteriet explicitly says they are *not* "öppna data" because conditions attach ([Avgiftsfria produkter](https://www.lantmateriet.se/sv/geodata/vara-produkter/avgiftsfria-produkter/)). Two tiers exist:

- **Öppna data** — licence "Creative Commons, CC0" ([Öppna data](https://www.lantmateriet.se/sv/geodata/vara-produkter/avgiftsfria-produkter/oppna-data/)). Verified on Geotorget product pages for *Topografi 50/100/250/1M Nedladdning, vektor* ("Villkor: Creative Commons, CC0", "Juridisk prövning: Nej", GeoPackage, download in Geotorget or via the *Geotorget Nedladdning* API).
- **Värdefulla datamängder (HVD)** — licence **CC BY 4.0** per the terms document *Användningsvillkor för värdefulla datamängder* (DNR LM2025/009266 v1.0, 2025-02-01, [PDF](https://www.lantmateriet.se/globalassets/geodata/geodataprodukter/anvandningsvillkor_for_vardefulla_datamangder.pdf)). Products with personal data use *Användningsvillkor för värdefulla datamängder som innehåller personuppgifter* (DNR LM2025/009269 v1.1, 2025-03-04, [PDF](https://www.lantmateriet.se/globalassets/geodata/geodataprodukter/anvandningsvillkor_for_vardefulla_datamangder_pu.pdf)).

Per-product status (Geotorget product pages, rendered 2026-09-16):

| Product | Fee | Terms | Legal review | Access | Format |
|---|---|---|---|---|---|
| [Belägenhetsadress Nedladdning, vektor](https://geotorget.lantmateriet.se/geodataprodukter/belagenhetsadress-nedladdning-vektor-api) | 0 | HVD terms **with personal data** | **Yes** | STAC API `https://api.lantmateriet.se/stac-vektor/v1`, collection `belagenhetsadresser` | GeoPackage, one zip per kommun |
| [Belägenhetsadress Direkt](https://geotorget.lantmateriet.se/geodataprodukter/belagenhetsadress-direkt-api) | 0 | HVD terms with personal data | Yes | REST `https://api.lantmateriet.se/distribution/produkter/belagenhetsadress/v4.2` (OAuth2 via API portal) | JSON / XML |
| [Belägenhetsadress Nedladdning, Inspire](https://www.lantmateriet.se/sv/geodata/vara-produkter/produktlista/belagenhetsadress-nedladdning-inspire/) | 0 | HVD terms with personal data | Yes (3–5 working days extra for security review) | Atom feed, login | GML; SWEREF 99 TM / ETRS89; semi-annual |
| [Ortnamn Nedladdning, vektor](https://geotorget.lantmateriet.se/geodataprodukter/ortnamn-nedladdning-vektor-api) | 0 | HVD terms (no personal data) | No | STAC collection `ortnamn`, one national zip (58.5 MB) | GeoPackage |
| [Ortnamn Direkt](https://geotorget.lantmateriet.se/geodataprodukter/ortnamn-direkt-api) | 0 | HVD terms | No | REST `…/distribution/produkter/ortnamn/v2.2` | JSON / XML |
| Kommun, län och rike (administrativ indelning) | 0 | STAC says `CC-BY-4.0`; a Lantmäteriet HVD presentation (GISS, 2025-03-04) tables it as **CC0** ("CC0 används som licens för tjänster som är gemensamma för HVD och NGP … Kommun, Län och Rike") — **conflicting, unverified which prevails** | No | STAC collection `kommun-lan-rike` (zip ~6 MB, yearly + "aktuell"); OGC API Features `https://api.lantmateriet.se/ogc-features/v1/administrativ-indelning` (collections `kommuner`, `lan`, `rike`, plus `-2025`/`-2026`; items endpoint returns 401) | GeoPackage / GeoJSON |
| [Topografi 10 Nedladdning, vektor](https://geotorget.lantmateriet.se/geodataprodukter/topografi-10-nedladdning-vektor) | 0 | HVD terms with personal data | **Yes** | Geotorget download / API | GeoPackage |
| Topografi 50 / 100 / 250 / 1M Nedladdning, vektor | 0 | **CC0** | No | Geotorget download / *Geotorget Nedladdning* API | GeoPackage |

Notes:
- *Administrativ indelning* is not a standalone Geotorget product; it is the STAC collection `kommun-lan-rike`, the OGC API above, an [Inspire download](https://www.lantmateriet.se/sv/geodata/vara-produkter/produktlista/administrativ-indelning-nedladdning-inspire/) (Atom/WCS), or a layer in Topografi 100/250/1M (CC0). Topografi 1M also carries "tätortsgränser och tätortssymboler".
- The STAC catalog root and collection/item listings are public (HTTP 200, no auth); it self-describes: "avgiftsfria och får användas enligt creative commons licens CC BY 4.0. För vissa informationsmängder kommer din användning att prövas juridiskt … och du behöver då godkänna särskilda användningsvillkor." Asset downloads on `dl1.lantmateriet.se` return **401 Basic** — credentials are the Geotorget account (private person) or a system account (organisation). Private accounts are "endast för privat bruk" ([konto-privatperson](https://geotorget.lantmateriet.se/konto-privatperson)); an MIT library is not private use, so apply as an organisation.
- **Record count.** 290 STAC items (one per kommun), total 324 MB zipped. Smallest Bjurholm 143 KB, largest Stockholm 9.1 MB, Göteborg 8.9 MB, Malmö 4.9 MB. Fastighetsregistret held 3,882,361 valid addresses in 2022 (2,886,796 street addresses + 964,737 rural) per [SCB/LM presentation](https://kartografiska.se/wp-content/uploads/5C_Foretagsadresser_SCB_LM.pdf); the OSM import thread cites 3.7 M. So ~3.9 M address points, not 4–5 M.
- **Fields** (JSON schema `belagenhetsadress-4.2.2.json`, matches *Nationell specifikation Adress* v1.0.1): `kommunkod`, `kommunnamn`, `kommundel` (geografisk kommundel), `adressomrade` + `adressomradestyp` (gatuadressområde / metertalsadressområde / byadressområde), `gardsadressomrade`, `adressplatsnummer`, `bokstavstillagg`, `lagestillagg`/`lagestillaggsnummer`, `adressplatstyp`, `adressplatsbeteckning`, `postnummer`, `postort` (both "beslutas av PostNord AB", only on current addresses), `popularnamn`, `distriktskod`/`distriktsnamn`, `adressattAnlaggning`, `anmarkningstyp`/`anmarkningstext`, `registerenhetsreferens` (property link), `objektidentitet`, `objektstatus`, `versionGiltigFran`, `geometri` (SWEREF 99 TM point). No län field — derive from kommunkod (first two digits).
- Lantmäteriet treats addresses as personal data via Fastighetsregisterlagen § 6: usage only for the purpose approved in the application, EU/EEA-only storage, spot checks, revocation (terms PU §§ 3.3–4.2). The OSM community reports Lantmäteriet refused ODbL compatibility and OpenAddresses ([#7657](https://github.com/openaddresses/openaddresses/issues/7657), [#7608](https://github.com/openaddresses/openaddresses/issues/7608)) has not been able to include it; the OSM thread ([127427](https://community.openstreetmap.org/t/lantmateriet-belagenhetsadress/127427), 91 posts, last 2026-07-25) says one mapper got approval for a full 3.7 M import but no formal import has happened.
- **Attribution text required (CC BY 4.0 + terms § 3.1):** product name as data source, "©Lantmäteriet", whether the data was processed, and that CC BY 4.0 applies — may be placed in accompanying documentation/metadata.
- Ortnamn: name types include `Tätort` ("minst 200 invånare"), Bebyggelse, Trakt, Kyrka, Anläggning, nature names; Ortnamnsregistret holds ~980,000 names (Lantmäteriet, figure not re-verified). Not delivered as CSV.

## 2. SCB (Statistics Sweden)

- **Licence.** Statistics and geodata published as open data: "Creative Commons 0 1.0 Universal, CC0" ([användningsvillkor](https://www.scb.se/om-scb/om-scb.se-och-anvandningsvillkor)); since 2021-07-01. Optional credit "Källa: SCB" / "Source: Statistics Sweden"; do not cite SCB as source for data you have processed. API limit 10 requests / 10 s per IP.
- **Län and kommuner with codes.** 21 län, 290 kommuner ([lan-och-kommuner](https://www.scb.se/hitta-statistik/regional-statistik-och-kartor/regionala-indelningar/lan-och-kommuner/)). Download: `https://www.scb.se/contentassets/7a89e48960f741e08918e489ea36354a/kommunlankod-2026.xlsx` (18 KB, HTTP 200). Codes: län 2 digits (01 Stockholm … 25 Norrbotten), kommun 4 digits = län + 2 (0180 Stockholm).
- **Population per kommun.** PxWeb API `https://api.scb.se/OV0104/v1/doris/sv/ssd/START/BE/BE0101/BE0101A/BefolkningNy` — region dimension has 312 values (riket, 21 län, 290 kommuner), years 1968–2024 (verified 2026-09-16). Also `BE0101C/BefArealTathetKon` (density, 1991–2025).
- **Tätorter.** 2,017 tätorter in 2023 (WFS `resultType=hits` on `stat:Tatorter_2023` → `numberOfFeatures="2017"`; SCB also states 2,017). Open geodata: `https://geodata.scb.se/geoserver/stat/wfs` (WFS/WMS) and GeoPackage download ([statistiska tätorter](https://www.scb.se/vara-tjanster/oppna-data/oppna-geodata/statistiska-tatorter/)). Attributes: `tatortskod`, `tatort`, `kommun`, `kommunnamn`, `lan`, `lannamn`, `area_ha`, `bef` (population), `ar`. Population per tätort also in PxWeb `MI0810/MI0810A/LandarealTatortN` (2,827 region codes, every 5 years to 2023). Also `Smaorter_2023`, `DeSO_2025`, `RegSO_2025`.

## 3. Postal codes

- **Owner.** PostNord Sverige owns and operates the postnummer system; PTS is supervisory authority; a Postnummerråd approves changes 3–4 times a year (sv.wikipedia). Register management is delegated to **Postnummerservice Norden AB**, machine access via **Geposit AB**. Postnummerservice's [terms](https://postnummerservice.se/en/information/kopvillkor): "free use within your own organization", "resale of data is not permitted", "sublicensing to third parties is not permitted" — unusable for an open library.
- **No open list from PostNord, Lantmäteriet or Digg.** dataportal.se search for "postnummer" (2026-09-16) yields only Postnummerservice's own commercial catalogue entry ("Svenska postnummer och postorter", published 2012), Helsingborg's municipal postnummer polygons, and Lantmäteriet's Inspire address entry. The old dataportal community called the postnummer system "slarvats bort till ett privat bolag". GitHub CSVs (`zegl/sweden-zipcode`, `lapplandi/sveriges-postnummer`, `beshrkayali/sverige_postnummer`) state no licence or provenance — do not use.
- **Lantmäteriet addresses carry `postnummer` + `postort`** on every current address point (set by PostNord), so the address product is the only lawful open-ish route to a *complete, current* postnummer ↔ postort ↔ street mapping.
- **GeoNames `SE.zip`** ([readme](https://download.geonames.org/export/zip/readme.txt)): CC BY 4.0 ("a link on your website to www.geonames.org is ok"). Downloaded and counted: **18,887 rows, 18,887 distinct postnummer, 1,780 postorter, 22 admin1 values (21 län + blank), 285 admin2 (kommun code) values; 3,055 rows lack kommun; accuracy: 16,115 rows = 4 (centroid of postal code area), 389 = 1, 2,383 blank**. Fields: country, postal code (`111 64` with space), place name, admin1 name/code (län), admin2 name/code (kommun 4-digit), admin3, lat, lon, accuracy. The GeoNames sources page lists Sweden's source as `www.pellesoft.se/upload/program/prg00745.zip` with **no date** (URL now dead) — a hobbyist upload, so freshness is unknown; the row count (18.9 k) exceeds today's ~17 k total codes, indicating retired codes are still present. The zip is rebuilt daily (Last-Modified 2026-09-16) but that says nothing about the underlying data. 96 KB zipped.
- **Counts.** Postnummerservice: "~17 000 postal codes", "10 500 deliverable postal codes", "1 743 postal cities"; Kartanalys: 10,839 geographic postnummer, 1,740 postorter; Wikipedia: >16,100 areas (2008). Changes quarterly.
- **Structure** (sv/en Wikipedia): 5 digits written `NNN NN`. Lower numbers further south, except 1xx xx = Stockholm. First digit/tens: 10–19 Stockholm, 20–29 Skåne (Malmö 20–21), 30–39 southern Sweden, 40–49 Göteborg area (Göteborg 40–41), 50–59, 60–69, 70–79, 80–89, 90–99 progressively north to Norrbotten (98x). Postorter come in 2-, 3- or 5-position sizes; in two-position places the 3rd digit (in three-position the 4th) marks delivery type: 0 = boxes/postal, 1 = boxes/business, 2–4 and 6–7 = street delivery, 5 = rural (lantbrevbäring), 8 = reply mail, 9 = competitions/temporary — with exceptions in big cities. Box codes are therefore not street-deliverable and should be excluded when generating street addresses.

## 4. Streets

- **Lantmäteriet** — see § 1; `adressomrade` with `adressomradestyp = gatuadressområde` is the street name.
- **Trafikverket NVDB** — CC0 1.0 Universal since 2017-04-03 ([trafikverket.se](https://www.trafikverket.se/e-tjanster/hamta-data-fran-trafikverket/) links `creativecommons.org/publicdomain/zero/1.0/deed.sv`; [OSM wiki](https://wiki.openstreetmap.org/wiki/Sweden/trafikverket)). Data product **Gatunamn** = "det officiellt adressbildande namnet på gatan" (municipal decision); Trafikverket is cleaning it in 2026 so it holds only names used in street addresses ([Vägnamn i NVDB](https://www.nvdb.se/sv/aktuellt/nyhetsarkiv/2026/vagnamn-i-nvdb/)). Download via Lastkajen (registration with e-mail + accept licence; Shapefile/GeoPackage), or the Datautbytesportal API. Gives street ↔ kommun (and geometry) but **no postnummer/postort**. Record count not published on the pages checked.
- **OpenStreetMap** — ODbL 1.0. A list of street names extracted from OSM is a Derivative Database: "Where you make our data or any Derivative Database available to others, it must continue to be licensed under the ODbL" ([OSMF FAQ](https://osmfoundation.org/wiki/Licence/Licence_and_Legal_FAQ)); only *Produced Works* (maps etc.) escape share-alike. Embedding an OSM-derived JSON in an MIT repo therefore forces that data file to ODbL with attribution "© OpenStreetMap contributors" + link to openstreetmap.org/copyright, and downstream consumers inherit share-alike on the data. Not recommended.
- **OpenAddresses** — 16 Swedish sources, all municipal (`sources/se/municipality_of_*.json`: Alingsås, Gislaved, Göteborg, Helsingborg, Höganäs, Kalmar, Kristinehamn, Malmö, Nacka, Sävsjö, Stockholm, Uppsala, Västerås, Vaxholm, Växjö, Österåker). OA does not relicense; licence is per source. Checked: Göteborg **CC0 1.0** (attribution "Göteborgs stad", shapefile, `gatunamn`), Malmö **CC0 1.0** (CSV with `ADRESSOMR`, `ADRESSPLAT`, `POSTNR`, `POSTORT`, cached 2024-11), Helsingborg **CC BY 4.0** ("Helsingborgs stad", GeoJSON with Gatunamn/Postnummer/Stad), Uppsala attribution required, licence name not stated (ArcGIS REST), Stockholm **no licence field**, data is a 2016 cache of a defunct WS. Usable for a few big cities only; no national coverage.
- **Unique street-name counts.** SCB (Lägenhetsregistret 2019, [artikel](https://www.scb.se/hitta-statistik/artiklar/2021/pa-ringvagen-bor-det-flest/)): "Antalet gator eller vägar med en eller flera adresser är 399 175 i hela landet" — this is streets counted per place, not unique names. Most frequent: Ringvägen 205, Skogsvägen 204, Björkvägen 200, Skolgatan 196, Storgatan 180 (in 180 kommuner, 55,360 residents). A national unique-name count was not found; expect well under 399 k (probably 100–200 k) — **unverified**.

## 5. Sizes

| Level | Count | Source |
|---|---|---|
| Län | 21 | SCB |
| Kommuner | 290 | SCB; 290 STAC items |
| Tätorter (2023) | 2,017 | SCB WFS hits |
| Postorter | ~1,740–1,780 | Postnummerservice 1,743; Kartanalys 1,740; GeoNames 1,780 |
| Postnummer | ~17,000 total, ~10,500–10,840 deliverable/geographic | Postnummerservice, Kartanalys |
| Streets with addresses (per place) | 399,175 (2019) | SCB |
| Unique street names | not published; likely 100–200 k | unverified |
| Address points | ~3.9 M (3,882,361 in 2022) | Lantmäteriet/SCB |
| Lantmäteriet address GeoPackages | 324 MB zipped, 290 files | STAC |
| GeoNames SE.zip | 96 KB (18,887 rows) | download |
| Ortnamn national GeoPackage | 58.5 MB zipped; ~980 k names | STAC; Lantmäteriet |

Embedding suggestion (size-driven, before licence): län + kommun + population (~20 KB) and tätorter with population (~100 KB) are trivially embeddable under CC0. A postnummer→postort table is ~300 KB raw. A national (street, postort, postnummer-range) table at ~400 k rows is roughly 15–25 MB JSON, ~3–5 MB gzipped — optional download pack territory; a default pack could carry the top N streets per kommun (e.g. 20 × 290 ≈ 6 k rows, <300 KB).

## 6. Licence compatibility with an MIT repo

| Source | Licence | Embed? | Required notice |
|---|---|---|---|
| SCB codes, population, tätorter | CC0 1.0 | Yes | None; optional "Källa: SCB" |
| Lantmäteriet Topografi 50/100/250/1M (admin boundaries, tätort symbols) | CC0 | Yes | None (courtesy credit suggested) |
| Lantmäteriet Ortnamn | CC BY 4.0 + HVD terms (no personal data, no review) | Yes, data file stays CC BY 4.0 inside the MIT repo | "Källa: Ortnamn Nedladdning, vektor, ©Lantmäteriet, bearbetad, CC BY 4.0" (product name, ©Lantmäteriet, processed-flag, CC BY 4.0) |
| Lantmäteriet kommun-län-rike | CC BY 4.0 per STAC (CC0 per LM slide — unresolved) | Yes | Same attribution form as above unless CC0 is confirmed |
| Lantmäteriet Belägenhetsadress | CC BY 4.0 + personal-data terms, purpose approval, EU-only storage, revocable | **Only with an approved application** whose stated purpose covers deriving and publishing an aggregated street/postnummer/postort list under CC BY 4.0. Raw points (house numbers, property links) should not be redistributed. Get written confirmation from geodatasupport@lm.se; treat as unresolved until then. | Product name, ©Lantmäteriet, "bearbetad", CC BY 4.0 |
| GeoNames SE | CC BY 4.0 | Yes | Link/credit to www.geonames.org; note staleness |
| Trafikverket NVDB Gatunamn | CC0 1.0 | Yes | None |
| OpenStreetMap | ODbL 1.0 | Not practical (share-alike on the data) | "© OpenStreetMap contributors" + ODbL if ever used |
| OpenAddresses Göteborg, Malmö | CC0 1.0 | Yes | None (Göteborg asks credit "Göteborgs stad") |
| OpenAddresses Helsingborg | CC BY 4.0 | Yes | "Helsingborgs stad" |
| OpenAddresses Stockholm/Uppsala | unspecified | No | — |
| PostNord/Postnummerservice files | commercial, no redistribution | No | — |

CC BY 4.0 and CC0 data files can sit in an MIT repo; the MIT licence covers the code, and a `DATA-LICENSES.md`/NOTICE lists each dataset with its licence and the attribution above. CC BY 4.0 § 2(a)(5)(B) (no technical protection measures) is irrelevant for a plain JSON file.

## Could not verify
- Whether the kommun-län-rike collection is CC0 or CC BY 4.0 (sources disagree).
- Whether Lantmäteriet will approve an application whose purpose is publishing a derived open dataset; OSM/OpenAddresses experience suggests refusal is likely.
- GeoNames Sweden source date; assume stale.
- National count of unique street names; Gatunamn (NVDB) record count.
- Exact number of active postnummer today (three sources: ~17 k total / 10.5–10.8 k deliverable).
