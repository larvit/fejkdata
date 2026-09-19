# Open geographic / postal-address data outside Sweden — for fejkdata (MIT)

Researched 2026-09-16. "Embeddable in MIT repo" below means: the data licence permits redistribution and commercial use with at most an attribution notice, so the data can ship inside an MIT-licensed package with a separate `DATA-LICENSES` file. Anything share-alike (ODbL, CC BY-SA) is flagged as **not** embeddable. Items marked *(unverified)* were not confirmed against a primary source.

## 0. Summary table

| Country | Official open address register | Licence | Addresses | Postcode in register | Postcode data open? | Embed in MIT? |
|---|---|---|---|---|---|---|
| US | DOT National Address Database (NAD) | US public domain | ~80 M (Jun 2026) | ZIP (optional field; fill rate *unverified*) | ZCTA yes (public domain); USPS publishes no open ZIP file | Yes |
| Norway | Kartverket Matrikkelen – Adresse | CC BY 4.0 | 2.58 M (Jun 2024) | Yes (postnummer + poststed) | Yes: Bring postnummerregister, NLOD 2.0 | Yes |
| Denmark | Danmarks Adresseregister (DAR) via DAWA / Datafordeler | CC BY 4.0 | ~3.8 M | Yes | Yes (in DAR; 1,089 postnumre) | Yes |
| Finland | DVV building addresses (final release Feb 2025) + Posti PCF/BAF | CC BY 4.0 / Posti terms (free; redistribute with terms attached) | ~5 M buildings | Yes | Yes (Posti PCF/BAF) | Yes (Posti: attach terms) |
| UK (GB) | No open *address* register. OS Open UPRN (points only), OS Open Names (streets/places/postcodes), Code-Point Open (postcodes) | OGL v3 | 40 M UPRNs (no addresses); 870 k roads; 1.7 M postcodes | n/a | Yes for GB (Code-Point Open, OGL); Northern Ireland restricted | Yes (GB only; no house numbers) |
| Germany | No national register. 15 Länder publish Hauskoordinaten/Gebäudereferenzen (dl-de/by-2-0, dl-de/zero-2-0); Bavaria missing | mixed, mostly permissive | ~19 M points across Länder | Yes in Hauskoordinaten | **No** — Deutsche Post DATAFACTORY is proprietary; open PLZ lists are OSM-derived (ODbL) | Partly (Länder files yes; a national PLZ list is the problem) |
| Netherlands | BAG (Kadaster/PDOK) | CC0 1.0 | ~9.9 M | Yes | Yes (in BAG) | Yes |
| France | Base Adresse Nationale (BAN) | Licence Ouverte / Etalab 2.0 | >25 M | Yes | Yes (BAN; La Poste "base officielle des codes postaux", open licence) | Yes |
| Spain | Catastro INSPIRE Addresses (AD) + INE callejero | Catastro own licence (free, attribution) / INE legal notice | ~15.7 M points (mailwoman count) | Yes (AD:PostCode) | Correos' postcode layer closed since 2017; postcodes still appear per address in Catastro AD | Probably (read Catastro licence PDF first) |
| Australia | G-NAF (Geoscape via data.gov.au) | EULA based on CC BY 4.0 + "no mail-out" restriction | 15.95 M (Aug 2026) | Yes | In G-NAF; Australia Post file is non-commercial only | Yes (with EULA notice) |
| Canada | StatCan National Address Register (NAR) 2024; ODA v1 (2021) | StatCan Open Licence / OGL-Canada | 17.1 M / ~10 M | NAR: MAIL_POSTAL_CODE "for most addresses" | Canada Post's postal code file is **proprietary**; only what NAR/ODA carry | Yes (NAR); postcode completeness unverified |

## 1. Per-country detail

### United States

**Official address register — National Address Database (NAD), US DOT**
- Licence: US public domain (`http://www.usa.gov/publicdomain/label/1.0/`), stated on data.gov: https://catalog.data.gov/dataset/national-address-database-nad-text-file. No attribution required. Embeddable.
- Size: ~80 M records; current release compiled 2026-06-30, dataset updated 2026-09-03. Sources: https://data.transportation.gov/dataset/National-Address-Database-NAD-Text-File/fc2s-wawr/about_data, https://www.placekey.io/datasets/national-address-database (80 M; "24 fields" in the text export).
- Coverage: aggregated from state/local providers; a mix of fully, partially and non-participating states (placekey page; the DOT page https://www.transportation.gov/gis/national-address-database blocks scrapers, so the coverage map was not read).
- Formats: zipped flat text and File Geodatabase. The text file is ordered by spatial cluster, not by OID.
- Fields (NAD schema https://www.transportation.gov/sites/dot.gov/files/2023-07/NAD_Schema_202304.pdf, blocked for bots; list taken from the FGDC content-requirements doc https://www.fgdc.gov/organization/working-groups-subcommittees/address-sc/220810-nad-content-requirements-approved.pdf): address number parts (`AddNum_Pre`, `Add_Number`, `AddNum_Suf`), street name parts (`St_PreDir`, `St_PreTyp`, `St_Name`, `St_PosTyp`, `St_PosDir`, `StNam_Full`), sub-address (`Building`, `Floor`, `Unit`, `Room`), place names (`Inc_Muni`, `Post_City`, `Census_Plc`, `Uninc_Comm`), `County`, `State`, `Zip_Code`, `Plus_4`, `Longitude`, `Latitude`, `NatGrid`, `Addr_Type`, `Placement`, `NAD_Source`, `DataSet_ID`. FGDC: "A Zip Code is recommended but not mandatory"; a complete place name and state are mandatory. Expect some records without ZIP.
- Census Address Count Listing files give housing-unit counts per block, not addresses: https://census.gov/geographies/reference-files/2025/geo/addcountlisting.html. The Census MAF itself is not public.

**Administrative hierarchy — Census Bureau (public domain)**
- States: 50 + DC + PR + 4 island areas = 56 FIPS state codes ("States & Equivalents: 56" in the 2020 tallies). ANSI list: https://www.census.gov/library/reference/code-lists/ansi/ansi-codes-for-states.html (also carries FM, MH, PW, UM).
- Counties: 3,144 in the 50 states + DC; 3,234 including PR and island areas (2020 tallies https://www.census.gov/geographies/reference-files/time-series/geo/tallies.html; https://en.wikipedia.org/wiki/List_of_United_States_counties_and_county_equivalents). Connecticut switched to 9 planning regions as county equivalents in 2022 — use a current vintage.
- Places: 19,734 incorporated places + 12,454 CDPs (2020 tallies, incl. PR/island areas); "approximately 19,500" incorporated places in the 50 states (https://www.census.gov/library/stories/2020/05/america-a-nation-of-small-towns.html).
- TIGER/Line 2025 shapefiles and the pipe-delimited **Gazetteer files** (2026 vintage: states, counties, county subdivisions, places, ZCTAs; GEOID, name, area, centroid lat/lon; **no population**): https://www.census.gov/geographies/reference-files/time-series/geo/gazetteer-files.html. Gazetteer excludes island areas, includes PR.
- Population per place: **SUB-EST2025** CSVs (incorporated places + MCDs, April 2020 base through July 2025, released May 2026): https://www.census.gov/data/tables/time-series/demo/popest/2020s-total-cities-and-towns.html. Public domain. The right file for population-weighted draws.

**Postal codes**
- USPS publishes no open ZIP list file; counts quoted 41,541–41,695 active ZIPs (https://facts.usps.com/42000-zip-codes/ says 41,554). Commercial "USPS-licensed" databases exist.
- Census **ZCTA**: 33,791 ZCTAs (2020), public domain, in TIGER + Gazetteer + relationship files (ZCTA↔place, ZCTA↔county): https://www.census.gov/programs-surveys/geography/technical-documentation/records-layout/2020-zcta-record-layout.html. PO-box-only ZIPs have no ZCTA (https://en.wikipedia.org/wiki/ZIP_Code_Tabulation_Area).
- GeoNames `US.txt` (CC BY 4.0): ZIP → place name, state, county, lat/lon. Count not stated on the site *(≈41 k, unverified)*.

**Street names**
- TIGER/Line `featnames` relationship files (per county) hold every feature name with prefix/suffix parts; joined to `edges` and `addrfeat` (address ranges). Public domain. No published count of unique street names *(unverified; must be computed)*.
- NAD `St_Name`/`StNam_Full` distinct per `Post_City` is the cheaper source for "real streets in Denver".

**Sizes for the embed-vs-download decision (US)**

| Level | Count | Source |
|---|---|---|
| States + DC + PR + island areas | 56 | Census tallies 2020 |
| Counties / equivalents | 3,144 (50+DC) / 3,234 (with PR + islands) | Census tallies, Wikipedia |
| Incorporated places | 19,734 (all) / ~19,500 (50+DC) | Census tallies |
| CDPs | 12,454 | Census tallies |
| ZCTAs / USPS ZIPs | 33,791 / ~41,500 | Census tallies / USPS facts |
| Address points (NAD) | ~80 M | DOT |
| OpenAddresses US collections | NE 2.35 GB + South 7.11 GB + West 8.54 GB + Midwest 4.86 GB (zipped CSV); 2,381 US sources | https://batch.openaddresses.io/ (2026-09) |
| Unique street names | unknown; compute from NAD/TIGER | — |

### Norway
- **Matrikkelen – Adresse** (Kartverket). Licence CC BY 4.0 (https://data.norge.no/en/datasets/8f1151bf-ad63-3e47-adbb-143f876776ee/matrikkelen-adresse; Geonorge register https://register.geonorge.no/inspire-statusregister/matrikkelen-adresse/73bc0329-faac-419b-80ff-d113c0ffe6a0). Formats: CSV, SOSI, GML, SQL/PostGIS, FGDB, Atom feed, REST API `https://api.kartverket.no/adresser/v1/`. Files per kommune (daily) and fylke/country (weekly).
- Counts (Kartverket "Antall vegadresser og matrikkeladresser", extract 2024-06-01, https://www.kartverket.no/datakvalitet/rapporter/adresse/Antall%20vegadresser%20og%20matrikkeladresser.xlsx): **2,577,825 addresses** — 2,536,663 vegadresser (street addresses) + 41,162 matrikkeladresser (cadastral, no street name).
- Fields (CSV): adressenavn (street), nummer, bokstav, postnummer, poststed, kommunenummer, kommunenavn, grunnkrets, tettsted, coordinates (EPSG:25833 / 4258); bruksenhetsnummer only in non-CSV formats. Fylke = first two digits of kommunenummer.
- Hierarchy: 15 fylker + Svalbard, 357 kommuner (2024). Kartverket/SSB code lists are open *(SSB Klass API licence NLOD, unverified)*.
- Postal codes: Bring/Posten **Postnummerregister** — postnummer, poststed, kommunenummer, kommunenavn, kategori; TAB text + xlsx; **NLOD 2.0** per data.norge.no (https://data.norge.no/en/datasets/5e6847ba-156d-4e14-85d3-8d7f8b727523/postnummer-i-norge). Updated annually (Sept/Oct). Embeddable.
- OpenAddresses: 17 `no/` sources.

### Denmark
- **Danmarks Adresseregister (DAR)**, served by **DAWA** (`https://api.dataforsyningen.dk/`) and Datafordeler. Licence **CC BY 4.0**, credit "Klimadatastyrelsen" (https://datafordeler.dk/vejledning/brugervilkaar/danmarks-adresseregister-dar/). Embeddable.
- Counts: ~3.8 M adresser (Klimadatastyrelsen DAWA report); DAWA today: **1,089 postnumre, 53,976 vejnavne, 99 kommuner, 5 regioner** (queried 2026-09-16 via `?format=csv`).
- Fields (DAWA `adresser`): vejnavn, husnr, etage, dør, postnr, postnrnavn, supplerende bynavn, kommunekode/-navn, regionskode/-navn, sogn, koordinater (ETRS89/WGS84), DAR ids. Downloads: CSV, JSON, GeoJSON, per kommune or whole country; `adgangsadresser` = access address (building entrance, no floor/door). Docs: https://dawadocs.dataforsyningen.dk/dok/adresser.
- Hierarchy: 5 regioner → 98 kommuner (99 rows incl. Christiansø) → postnumre/byer. All in DAWA.
- Postcodes: in DAR (source PostNord, updated on change), same licence.
- OpenAddresses: `dk/countrywide` (1 source).

### Finland
- No single address-register download today. Sources:
  - **DVV (Digital and Population Data Services Agency) building address open data** — CC BY 4.0; final official release 2025-02-14, distribution ended 2025-03-14; archived copies (CSV/GeoPackage, ~5 M buildings: street fi/sv, house number, postal code, municipality, WGS84) at https://markuskainu.fi/posts/2025-02-03-dvv-rakennusten-osoitetiedot/. Successor: **SYKE/Ryhti building addresses** via OGC API Features, CC BY 4.0 (https://avoindata.suomi.fi/data/fi/dataset/syke-rakennusten-osoitteet-ogc-api-features) — page blocked the fetch; verify fields and bulk download before relying on it.
  - **NLS (Maanmittauslaitos)** topographic database address points — CC BY 4.0 (https://www.maanmittauslaitos.fi/en/opendata-licence-cc40); address points computed from road-name data.
  - **Digiroad** (Fintraffic since 2026-01-01) — CC BY 4.0; road names + address number ranges (https://vayla.fi/en/transport-network/data/digiroad).
- **Posti postal code files** (https://www.posti.fi/en/for-businesses/customer-support/postal-code-services; directory https://www.posti.fi/webpcode/): **PCF** (postal code, names fi/sv + abbreviations, type code, region code/name, municipality code/name fi/sv, language code; daily), **BAF** (every street per postal code with odd/even building-number ranges + municipality; weekly; excludes Åland), **POM** (monthly changes). Fixed-width `.dat`, Unix LF. Terms (service description 2015-01-01): "The files are freely downloadable … The data can be disclosed to third parties but it must be ensured that the recipient of the data is aware of the terms of use for the service as well as the download date." No licence name, no fee, no commercial restriction stated. Embeddable if the terms PDF + download date ship with the data.
- Hierarchy: Statistics Finland classifications, CC BY 4.0: Municipalities 2026 https://stat.fi/en/luokitukset/kunta/kunta_1_20260101, Regions 2026 https://stat.fi/en/luokitukset/maakunta/maakunta_1_20260101. Posti PCF also carries municipality + region codes.
- OpenAddresses: `fi/countrywide-fi`, `fi/countrywide-sv` + 38 municipal sources.

### United Kingdom (Great Britain; NI separate)
- No open full-address register (AddressBase is paid). Open pieces, all **OGL v3**, attribution "Contains OS data © Crown copyright and database right [year]" (https://wiki.openstreetmap.org/wiki/Ordnance_Survey_OpenData_Licence; OS moved to OGL v3 in 2015):
  - **OS Open Names** — "over 870 000 named and numbered roads, nearly 44 000 settlements and over 1.6 million postcodes"; 34 attributes incl. `name1`, `name2` (Welsh/Gaelic), `type`, `local_type` (populatedPlace: City/Town/Village/Hamlet/Suburban Area/Other Settlement; transportNetwork: Named Road, Numbered Road, Section Of …; other: Postcode), `postcode_district`, `populated_place`, `district_borough`, `county_unitary`, `region`, `country`, `GEOMETRY_X/Y`. CSV/GML/GeoPackage, quarterly. https://docs.os.uk/os-downloads/products/addresses-and-names-portfolio/os-open-names. Gives street → settlement → district → county → region → country, but **no house numbers** and only the postcode district per road.
  - **Code-Point Open** — ~1.7 M GB postcode units; fields Postcode, Positional_quality_indicator, Eastings, Northings, Country_code, NHS codes, Admin_county_code, Admin_district_code, Admin_ward_code. No street or locality names. GB only; NI excluded. https://docs.os.uk/os-downloads/products/areas-and-zones-portfolio/code-point-open.
  - **OS Open UPRN** — ~40 M UPRN + coordinates, no addresses. https://www.ordnancesurvey.co.uk/products/os-open-uprn.
  - **ONS Postcode Directory (ONSPD)** — postcode → all admin geographies; OGL, but requires Royal Mail attribution ("Contains Royal Mail data © Royal Mail copyright and database right [year]") and NI (BT) postcodes are end-user-licence only (no commercial redistribution): https://www.ons.gov.uk/methodology/geography/licences. Drop NI rows.
- Hierarchy codes: ONS GSS codes (countries, regions, counties/UAs, districts, wards) via the ONS Open Geography Portal, OGL. Population: ONS mid-year estimates (OGL) at local-authority level; settlement population must come from GeoNames/Wikidata.
- OpenAddresses: 0 `gb/` sources.
- Practical: a deliverable GB address needs street + number + postcode unit; open data gives street + postcode *district* only. "Real street in a real town with a plausible district" is achievable; a real full postcode for that street is not.

### Germany
- **No national open address register.** BKG sells "Georeferenzierte Adressdaten" commercially. Open data is per Land: Hauskoordinaten / Gebäudereferenzen from 15 Länder under dl-de/by-2-0, dl-de/zero-2-0 or CC BY 4.0; **Bavaria missing** (~13 M people) — https://github.com/sister-software/mailwoman/issues/2300 (19.3 M permissive points vs 20.3 M in an ODbL build). Example: NRW "Gebäudereferenzen NW", dl-de/zero-2-0, semi-annual, https://open.nrw/dataset/172a0ba8-d470-47c1-ac89-b85f8190ac7e. Thuringia is dl-de/by-2-0. OpenAddresses lists 14 `de/*/statewide` sources (bb, bw, hb, he, hh, mv, ni, nw, rp, sh, sl, sn, st, th) — no `by`; Berlin appears as a Geoportal source *(check)*.
- Hierarchy: **BKG VG250** (Land → Regierungsbezirk → Kreis → Verwaltungsgemeinschaft → Gemeinde with AGS keys; VG250-EW adds population), dl-de/by-2-0: https://gdz.bkg.bund.de/index.php/default/open-data/verwaltungsgebiete-1-250-000-stand-01-01-vg250-01-01.html. **Destatis Gemeindeverzeichnis GV100/GV-ISys** (all Gemeinden with AGS, population, area, PLZ of the Verwaltungssitz), quarterly ASCII/Excel: https://www.destatis.de/DE/Themen/Laender-Regionen/Regionales/Gemeindeverzeichnis/_inhalt.html — licence dl-de/by-2-0 *(not stated on that page; Destatis default, unverified)*. 16 Länder, ~400 Kreise, ~10,800 Gemeinden *(approx.)*.
- **Postal codes are the blocker**: Deutsche Post **DATAFACTORY** is proprietary and licensed for internal use only (https://www.deutschepost.de/de/d/deutsche-post-direkt/datafactory.html; a FragDenStaat request for the data was refused). Open PLZ lists (OpenPLZ API, yetzt/postleitzahlen, Geofabrik postcode polygons) are **OSM-derived → ODbL** (OpenPLZ sources page https://www.openplzapi.org/en/sources/: streets and PLZ from OSM, municipalities from Destatis GV100). GeoNames `DE.txt` provenance not stated *(unverified)*. The Länder Hauskoordinaten carry PLZ per address, so a PLZ↔Ort↔Straße table can be derived legitimately for the 15 covered Länder; GV100 gives one PLZ per Gemeinde seat.
- OpenAddresses: 32 `de/` sources.

### Netherlands
- **BAG** (Basisregistratie Adressen en Gebouwen), Kadaster. Licence **CC0 1.0** (https://data.overheid.nl/en/dataset/basisregistratie-adressen-en-gebouwen--bag-; OSM wiki: "released under a Public Domain license"). Embeddable, no attribution.
- ~9.9 M addresses (mailwoman count). Object types: panden, verblijfsobjecten, **nummeraanduidingen** (huisnummer + huisletter + toevoeging + postcode), **openbare ruimten** (streets), **woonplaatsen**. Fields yield street, number, postcode (4 digits + 2 letters), woonplaats, gemeente (CBS code), RD/WGS84 point.
- Download: BAG Extract XML (monthly full + mutations) and BAG GeoPackage via PDOK Atom `https://service.pdok.nl/lv/bag/atom/bag.xml`; REST API `https://api.pdok.nl/lv/bag/`. data.overheid.nl mentions distribution fees for some Kadaster products; the PDOK extract/GeoPackage is free.
- Hierarchy: 12 provincies → 342 gemeenten (2025) → ~2,500 woonplaatsen. CBS gebiedsindelingen (CC BY 4.0); CBS "Kerncijfers wijken en buurten" for population.
- Postcodes: in BAG since 2012 (CC0). PostNL's product is proprietary but unnecessary.
- OpenAddresses: `nl/countrywide` (1 source).

### France
- **Base Adresse Nationale (BAN)** — **Licence Ouverte / Etalab 2.0** since 2020-01-01 (https://fr.wikipedia.org/wiki/Base_adresse_nationale; https://adresse.data.gouv.fr/outils/telechargements). Embeddable with attribution.
- >25 M addresses. Downloads: CSV national / per département / per commune, "CSV with BAN ids", addok format; daily updates with weekly/monthly archives; also MVT/WFS/WMS. CSV fields `id, id_fantoir, numero, rep, nom_voie, code_postal, code_insee, nom_commune, code_insee_ancienne_commune, nom_ancienne_commune, x, y, lon, lat, type_position, alias, nom_ld, libelle_acheminement, nom_afnor, source_position, source_nom_voie, certification_commune, cad_parcelles` *(from BAN docs, not re-verified this session)*.
- Hierarchy: INSEE **Code Officiel Géographique** (18 régions, 101 départements, ~34,900 communes), Licence Ouverte; INSEE populations légales per commune (Licence Ouverte).
- Postal codes: in BAN; also La Poste **"Base officielle des codes postaux"** (commune → code postal → libellé d'acheminement, INSEE code), open licence, https://datanova.laposte.fr/datasets/laposte-hexasmal and https://www.data.gouv.fr/datasets/base-officielle-des-codes-postaux.
- OpenAddresses: 107 `fr/` sources (one per département, fed weekly from BAN).

### Spain
- **Catastro INSPIRE Addresses (AD)** — per-municipality GML via Atom `http://www.catastro.minhap.es/INSPIRE/Addresses/ES.SDGC.AD.atom.xml`, refreshed ~6-monthly; WFS `http://ovc.catastro.meh.es/INSPIRE/wfsAD.aspx`. Fields: ThoroughfareName, address number, `AD:PostCode` (5 digits), AdminUnitName (municipio, provincia), point at building entrance or parcel centroid. Coverage: 95% of Catastro territory, **excludes País Vasco and Navarra** (own cadastres; Navarra has OpenAddresses source `es/nc/statewide`). Licence: "licencia de cesión de derechos que se obtendrá de manera automática", free, attribution "© Dirección General del Catastro" (https://www.catastro.hacienda.gob.es/webinspire/documentos/Conjuntos%20de%20datos.pdf; https://www.catastro.hacienda.gob.es/webinspire/index.html). mailwoman labels it CC BY 4.0 (15.66 M points); the Catastro PDF does not say "CC BY" — **read the "Descripción de la licencia" PDF before embedding**.
- **INE Callejero del Censo Electoral** — VIAS, TRAMOS, PSEUDOVIAS, UNIDADES POBLACIONALES; ASCII ZIP, national + per province, Jan/Jul each year, `https://www.ine.es/prodyser/callejero/caj_esp/caj_esp_MMYYYY.zip` (https://datos.gob.es/en/catalogo/ea0010587-callejero-de-censo-electoral). Licence = INE aviso legal (free reuse with source citation *(unverified; page blocked)*). Whether TRAMOS carry código postal: *(unverified)*.
- **CartoCiudad (IGN/CNIG)** — addresses, postcodes, toponyms, admin units; CC BY 4.0 (https://datos.gob.es/es/aplicaciones/direcciones-postales-de-cartociudad-espana). Its postcode polygon layer came from Correos and was **withdrawn in 2017** (Correos now sells it, ~€6k/yr) — https://www.nosolosig.com/articulos/asi-hice-el-mapa-de-los-codigos-postales-de-espana-con-sig-y-datos-abiertos. Per-address postcodes remain in Catastro AD and CartoCiudad addresses.
- Hierarchy: 17 CCAA + 2 ciudades autónomas → 50 provincias → 8,132 municipios; INE padrón gives population per municipio; INE codes are open.
- OpenAddresses: `es/countrywide` (Catastro) + `es/nc/statewide` + 4 others.

### Australia
- **G-NAF** (Geoscape via data.gov.au, https://data.gov.au/data/dataset/geocoded-national-address-file-g-naf). August 2026: **15,949,543 addresses** (15,108,510 principal). Quarterly. PSV tables (~5 GB unpacked, many tables to join) + **G-NAF Core** single simplified table; GDA94 or GDA2020.
- Licence: **G-NAF End User Licence Agreement, "based on CC BY 4.0"**, plus: must not be used to compile addresses for sending mail unless each is verified against a secondary source; must follow the Australian Privacy Principles. Preferred attribution: "Incorporates or developed using G-NAF © Geoscape Australia licensed by the Commonwealth of Australia under the Open Geo-coded National Address File (G-NAF) End User Licence Agreement." Embeddable with the EULA text and that notice; the mail restriction is irrelevant to fake data but must be passed on.
- Fields: number first/last with prefix/suffix, flat/level, street name + type + suffix, locality, state, postcode, lat/lon, mesh block; LGA via ABS codes.
- Hierarchy: 8 states/territories → LGAs (~560) → localities (~15,000 in the G-NAF locality table). **ABS ASGS Edition 3** boundaries/allocation files, CC BY 4.0: https://www.abs.gov.au/statistics/standards/australian-statistical-geography-standard-asgs/edition-3-july-2021-june-2026/access-and-downloads/digital-boundary-files. Population: ABS ERP by SA2/LGA (CC BY 4.0).
- Postcodes: in G-NAF. Australia Post's postcode file is free only as a **non-commercial PDF**; CSV products are paid (https://auspost.com.au/business/services/data-services/address-data/postcode-data). ABS "Postal Areas" (2,644 POAs, CC BY 4.0) are mesh-block approximations.
- OpenAddresses: `au/countrywide` + 7 statewide + councils (58 sources).

### Canada
- **National Address Register (NAR)**, Statistics Canada — December 2024 release, **17.1 M** records, all 13 provinces/territories; two CSVs per province (addresses, locations). Licence **Statistics Canada Open Licence** (use, reproduce, distribute, sell, sublicense; attribution "Adapted from Statistics Canada, National Address Register, 2024. This does not constitute an endorsement by Statistics Canada of this product."): https://www.statcan.gc.ca/en/reference/licence; user guide https://www150.statcan.gc.ca/n1/pub/46-26-0002/462600022024002-eng.htm. Fields: `CIVIC_NO, CIVIC_NO_SUFFIX, OFFICIAL_STREET_NAME/TYPE/DIR, APT_NO_LABEL, MAIL_STREET_NAME/TYPE/DIR, MAIL_MUN_NAME, MAIL_PROV_ABVN, MAIL_POSTAL_CODE, CSD_CODE, CSD_ENG_NAME, CSD_FRE_NAME, PROV_CODE, REPPOINT_LATITUDE/LONGITUDE, BG_X/Y, BU_USE`. Postal code present "for most addresses" — completeness per province unverified. Embeddable.
- **Open Database of Addresses (ODA) v1** (2021), ~10 M records from 99 datasets, OGL-Canada, CSV per province, postal code where the provider had it: https://www.statcan.gc.ca/en/lode/databases/oda. Older; prefer NAR.
- Hierarchy: StatCan Standard Geographical Classification 2021 — 13 provinces/territories → 293 census divisions → ~5,160 census subdivisions (municipalities); population per CSD from Census 2021. StatCan Open Licence.
- **Postal codes are proprietary**: Canada Post asserts copyright, sells the file, and sued Geocoder.ca in 2012 (https://opennorth.ca/resources/open-postal-code-data/). StatCan's PCCF is licensed from Canada Post and not redistributable. GeoNames `CA.txt` = **FSA (first 3 chars) only**. Only NAR/ODA `MAIL_POSTAL_CODE` gives open per-address postal codes.
- OpenAddresses: `ca/countrywide` + 221 sources; Canada collection 1.25 GB.

## 2. Cross-country hierarchy sources

| Source | Licence | Content | Notes |
|---|---|---|---|
| Debian **iso-codes** (`iso_3166-1.json`, `iso_3166-2.json`; XML deprecated) | LGPL-2.1 (https://salsa.debian.org/iso-codes-team/iso-codes) | All ISO 3166-1 countries and 3166-2 subdivisions with codes, types, parents; 90+ translations | LGPL on a data file is awkward inside MIT; ship it as a separately-licensed data file or consume it only at build time. ISO's own OBP is not open. |
| **GeoNames** `admin1CodesASCII.txt`, `admin2Codes.txt` | CC BY 4.0 | admin1/admin2 names + geonameids per country | admin1 codes follow ISO 3166-2 for some countries and FIPS/numeric codes for others. |
| `amckenna41/iso3166-2` (PyPI) | MIT (code); data compiled from Wikipedia/RestCountries | 250 countries, >5,000 subdivisions with names, lat/lon | Convenient; provenance mixed. |
| Per-country official lists (section 1) | see country | authoritative codes + population | Preferred for the 11 target countries. |

## 3. Localities with population

| Source | Licence | Size | Fit |
|---|---|---|---|
| **GeoNames** `cities500/1000/5000/15000` + `allCountries` (https://download.geonames.org/export/dump/) | CC BY 4.0 (attribution to geonames.org) | cities500 ≈185 k, cities1000 ≈130 k, cities5000 ≈50 k, cities15000 ≈25 k; fields geonameid, name, asciiname, alternatenames, lat/lon, feature class/code (PPL, PPLA…), country, admin1–4 codes, population, timezone | Best single cross-country locality+population source; population uneven (many zeros). Embeddable with attribution. |
| **Census SUB-EST2025** (US) | public domain | ~19.5 k places | Authoritative US population. |
| **Wikidata** (P1082 population, P131 located-in, P281 postal code) | CC0 | all | No attribution; needs a SPARQL/dump pipeline; quality varies. |
| **Who's On First** (https://www.whosonfirst.org/docs/licenses/) | CC0 for WOF's own work; 312 sources ranging CC0…CC BY 3.0/4.0, OGL | localities, regions, counties with hierarchy + population where sourced | Attribution list is per source; no share-alike found. Heavy (one GeoJSON per record). |
| **dr5hn/countries-states-cities-database** (https://github.com/dr5hn/countries-states-cities-database) | **ODbL 1.0** | 250 countries, 5,299 states, 153,765 cities, 844,248 postcodes (125 countries) | **Share-alike → not embeddable in MIT.** Community-maintained; README: data "may contain errors or lag behind geopolitical changes"; issue tracker shows large re-parenting fixes (8,727 French cities). Not a source of truth. |

## 4. Postal codes — GeoNames and per-country openness

- **GeoNames postal** (https://download.geonames.org/export/zip/, CC BY 4.0): TSV fields `country code, postal code, place name, admin name1, admin code1, admin name2, admin code2, admin name3, admin code3, latitude, longitude, accuracy (1–6)`. ~100 countries. Readme caveats: Canada and Netherlands partial codes only (full NL/CA files separate); UK from Royal Mail (© Royal Mail 2022 — attribution needed); Ireland/Malta first letters only; Argentina 5 chars; Brazil only `-000` codes; Chile/China partial. Lat/lon is algorithmic (matched to a GeoNames toponym or averaged from neighbours) — a locality centroid, not the postcode's. Per-country provenance is not documented, so DE's origin is unknown.

| Country | Open? | Best open source |
|---|---|---|
| US | Yes (ZCTA, public domain); USPS list not a file | Census ZCTA + GeoNames US.txt |
| Norway | Yes (NLOD 2.0) | Bring postnummerregister; also in Matrikkelen |
| Denmark | Yes (CC BY 4.0) | DAWA `postnumre` (1,089) |
| Finland | Yes (Posti terms) | Posti PCF (codes) + BAF (streets per code) |
| UK | GB yes (OGL, Code-Point Open); NI restricted | Code-Point Open + OS Open Names |
| Germany | **No** (Deutsche Post proprietary); OSM lists are ODbL | Derive PLZ↔Ort↔Straße from Länder Hauskoordinaten (15 Länder); GV100 for seat PLZ |
| Netherlands | Yes (CC0, in BAG) | BAG |
| France | Yes (Licence Ouverte) | La Poste hexasmal + BAN |
| Spain | Polygon layer closed since 2017; per-address codes open via Catastro AD (own licence) | Catastro AD `AD:PostCode`; INE callejero |
| Australia | In G-NAF (EULA/CC BY 4.0); Australia Post file paid/non-commercial | G-NAF locality table |
| Canada | **No** (Canada Post proprietary; GeoNames FSA only) | NAR `MAIL_POSTAL_CODE` where present |

## 5. OpenAddresses and OpenStreetMap

**OpenAddresses** (https://openaddresses.io/, https://batch.openaddresses.io/)
- Global collection 48.39 GB (2026-09-07); US in 4 regional collections (table in section 1); Canada 1.25 GB. 3,282 sources (1,586 flagged "need attention"). Sources per country in the batch list: US 2,381, CA 222, FR 107, AU 58, FI 40, DE 32, NO 17, ES 6, DK 1, NL 1, GB 0.
- Licensing: the repo LICENSE (BSD-3) covers source definitions only. **Each source keeps its own licence**, recorded in the source JSON and `state.txt`; "most sources only require attribution", but several are share-alike (ODbL, CC BY-SA) — e.g. OSM-derived and some municipal sources (https://geocode.earth/docs/reference/data_sources/). For MIT: filter by the per-source licence field, keep PD/CC0/CC BY/OGL/dl-de sources, generate the attribution file from `state.txt`. The old results.openaddresses.io coverage page is frozen at mid-2021.
- The official registers above are themselves the OA sources for NO/DK/FI/NL/FR/ES/AU (and NAD/state sources for the US); going to the register directly gives cleaner licensing and fresher data. OA's value is one uniform CSV schema (`LON,LAT,NUMBER,STREET,UNIT,CITY,DISTRICT,REGION,POSTCODE,ID,HASH`).

**OpenStreetMap** — ODbL 1.0
- Any *substantial* extract (systematic, >100 features, or an area with >1,000 inhabitants) is a Derivative Database; distributing it — including embedded in a library — requires offering it under ODbL (share-alike) plus attribution. OSMF guideline: https://osmfoundation.org/wiki/Licence/Community_Guidelines/Substantial_-_Guideline ("village map OK, town map not OK"; repeated small extractions count as one).
- Consequence: an embedded street/postcode list derived from OSM (OpenPLZ, Geofabrik postcode polygons, yetzt/postleitzahlen, OSM-derived dr5hn rows) cannot ship inside an MIT package without that data file being ODbL and tracked as a separate database; downstream users mixing it into their databases inherit share-alike. Keep OSM out of the default; at most an *optional* download pack clearly labelled ODbL.

## 6. Embed vs. optional download — sizes and a recommendation

**US (ships by default with en_US)**
- Embed (small, public domain): 56 states (+ USPS abbreviations, FIPS), 3,234 counties (GEOID, name, state), ~19.5 k incorporated places with population + county + lat/lon (SUB-EST2025 joined to the Gazetteer), 33,791 ZCTAs with the ZCTA→place/county relationship, and a compact **street-name pool per place** derived from NAD (distinct `St_Name`+`St_PosTyp` per `Post_City`/`State`, top-N by frequency). Places+ZCTAs ≈ 2–4 MB gzipped; a street pool for all ~19.5 k places would run to tens of MB — cap to the ~1,000 largest places or ~20 streets per place for the default pack.
- Download pack: full NAD (~80 M rows, multi-GB), TIGER `featnames`/`addrfeat` per county, OA US collections (~23 GB zipped). Real house numbers + ZIP+4 come only from here.
- Deliverability caveat: NAD ZIP is optional, and place naming mixes `Post_City` (USPS city) with `Inc_Muni`; use `Post_City` + `Zip_Code` for mailing realism.

**Other countries** — same shape: embed hierarchy + localities-with-population + postcode↔locality table (a few MB per country: NO's 2.6 M addresses collapse to ~4.5 k postcodes and ~90 k street names; DK 1,089 postcodes / 54 k street names; NL/FR/AU similar), with street pools for the largest localities; keep full address points as download packs fetched from the official Atom/CSV endpoints at runtime, never vendored (G-NAF ~5 GB, BAN CSV ~2 GB, BAG ~2 GB, NAR ~1 GB).

**Licence file needed in the repo** (`DATA-LICENSES.md`): CC BY 4.0 (GeoNames, Kartverket, Klimadatastyrelsen, Statistics Finland, DVV/SYKE, NLS, ABS, CartoCiudad), NLOD 2.0 (Bring), OGL v3 (OS, ONS), Licence Ouverte 2.0 (BAN, La Poste, INSEE), G-NAF EULA notice, StatCan Open Licence notice, dl-de/by-2-0 and dl-de/zero-2-0 (BKG, Länder), Posti terms + download date, Catastro "© Dirección General del Catastro". CC0/public domain (US Census, DOT NAD, BAG, Wikidata) need nothing but are worth listing.

## 7. Open questions / unverified
- NAD ZIP fill rate and the list of non-participating states (DOT site blocks fetches; read the NAD coverage map manually).
- Unique US street-name count — compute from NAD once downloaded.
- Catastro licence text ("Descripción de la licencia" PDF) — confirm redistribution inside a software package.
- INE aviso legal wording; whether callejero TRAMOS include código postal.
- Destatis GV100 licence statement; Berlin/Bayern Hauskoordinaten status.
- SYKE/Ryhti address API fields and bulk-download availability (avoindata.fi blocked the fetch).
- NAR `MAIL_POSTAL_CODE` completeness per province.
- GeoNames DE/US postal-code provenance and counts.
- BAN CSV field list re-verification.
