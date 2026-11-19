# Data licenses

Every shipped dataset, its source, its licence and the attribution it asks for. A
`data-import/` script rebuilds each sourced table, run as the README's
[Development](README.md#development) section says; a curated one is hand-written.

A heading names a dataset's first file, and a `Files:` line lists them all where they don't fit a
heading. `Rebuild:` names the script, or says the table is hand-written or curated.

Every other category is hand-written under [`data/`](data), MIT like the code.

## `da_DK/company.json`, `company-coined.json`, `company-word.json`

- Source: the weights of the legal forms of `company.json`: [Statistics
  Denmark](https://www.dst.dk/en/)'s table [GF15](https://www.statbank.dk/GF15), "General enterprise
  statistics", the enterprises of 2024: 132,403 sole proprietorships, 218,208 private limited
  companies (`ApS`), 30,750 public limited companies (`A/S`) and 11,679 partnerships and limited
  partnerships, which the table doesn't split, so `I/S` and `K/S` weigh 5,840 and 5,839; the name
  shapes and the two other files are a language model's recall, as no open bulk file of Danish
  company names was found, and weigh evenly
- Licence: the weights: [CC BY
  4.0](https://www.dst.dk/en/Statistik/brug-statistikken/kildeangivelse); the rest: the project's
  own [LICENSE](LICENSE)
- Attribution: "Kilde: Danmarks Statistik"
- Rebuild: hand-written

## `da_DK/first-name.tsv`, `last-name.tsv`

- Source: [Statistics Denmark](https://www.dst.dk/en/)'s ["Names in the entire
  population"](https://www.dst.dk/en/Statistik/emner/borgere/navne/navne-i-hele-befolkningen) for
  the top 20 girls' names, boys' names and last names, counts of 1 January 2026; every other name is
  a language model's recall of a name common in Denmark, as no good online source of Danish name
  counts was found: the commoner ones from the most to the least common, each with an estimated
  count of the twentieth's count times 20 over its rank, then the rarer ones in no order, sharing
  the count of the last rank
- Licence: the top 20 counts: [CC BY
  4.0](https://www.dst.dk/en/Statistik/brug-statistikken/kildeangivelse); the rest: the project's
  own [LICENSE](LICENSE)
- Attribution: "Kilde: Danmarks Statistik"
- Rebuild: `data-import/names-dk.py`

## `da_DK/month.tsv` and 5 more

- Files: `da_DK/month.tsv`, `de_DE/month.tsv`, `en_US/month.tsv`, `fi_FI/month.tsv`,
  `nb_NO/month.tsv`, `nn_NO/month.tsv` and each `weekday.tsv` beside them
- Source: [Unicode CLDR](https://github.com/unicode-org/cldr) 48.0.0, the Danish, German, English,
  Finnish, Norwegian Bokmål and Nynorsk stand-alone full month and weekday names, through
  [cldr-json](https://github.com/unicode-org/cldr-json)
- Licence: [Unicode License v3](https://www.unicode.org/license.txt)
- Attribution: "Copyright © 2015-2024 Unicode, Inc."
- Rebuild: `data-import/cldr.py`

## `da_DK/phone-mobile.tsv`, `phone-fixed.tsv`

- Source: Annex 1 of the Danish [Executive Order No. 1883 of 7 December 2020 on the Overall Danish
  Numbering
  Plan](https://en.digst.dk/media/tztputvv/executive-order-on-the-overall-danish-numbering-plan.pdf),
  in force from 21 December 2020: each three-digit prefix with a third digit other than 0, in a
  range the plan calls mainly mobile or mainly fixed; the prefixes its note lists as mobile inside
  fixed ranges are in the mobile table
- Licence: none stated, facts only; an executive order is a public act, which § 9(1) of the Danish
  Copyright Act (ophavsretsloven) excludes from copyright
- Attribution: none required
- Rebuild: `data-import/phone-dk.py`

## `da_DK/word.json` and 4 more

- Files: `da_DK/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `da_DK/sentence.json` reads
- Source: [COR₁ 1.5.1.0](https://ordregister.dk/files/cor1.5.1.0.tsv), [Dansk
  Sprognævn](https://dsn.dk/)'s central word register of Retskrivningsordbogen, less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  Danish list (CC BY 4.0)
- Licence: CC0
- Attribution: none required
- Rebuild: `data-import/word-dk.py`

## `de_DE/company.json`, `company-coined.json`, `company-word.json`

- Source: the weights of the legal forms of `company.json`: [Destatis](https://www.destatis.de/)'
  [Statistisches Unternehmensregister, Rechtliche Einheiten nach zusammengefassten
  Rechtsformen](https://www.destatis.de/DE/Themen/Branchen-Unternehmen/Unternehmen/Unternehmensregister/Tabellen/unternehmen-rechtsformen-wz08.html),
  the legal units of 2024: 2,062,752 sole proprietors, 440,951 partnerships and 861,720
  corporations; the table doesn't split a group, so `KG` and `OHG` weigh 220,476 and 220,475, and
  `GmbH`, `UG (haftungsbeschränkt)` and `AG` weigh 287,240 each; the 178,442 units of other legal
  forms, such as associations and foundations, aren't drawn; the name shapes and the two other files
  are a language model's recall, as no open bulk file of German company names was found, and weigh
  evenly
- Licence: the weights: [Datenlizenz Deutschland – Namensnennung
  2.0](https://www.govdata.de/dl-de/by-2-0) (`dl-de/by-2-0`); the rest: the project's own
  [LICENSE](LICENSE)
- Attribution: "© Statistisches Bundesamt (Destatis), 2025", with the licence's name and link, and
  that the groups are split and the counts reduced to six legal forms
- Rebuild: hand-written

## `de_DE/first-name.tsv`, `last-name.tsv`

- Source: first names: the city of Munich's open data ["Vornamen von
  Neugeborenen"](https://opendata.muenchen.de/dataset/vornamen-von-neugeborenen), the files for 2013
  to 2025, each name and sex with the sum of the counts they publish for it; a file gives a count
  below 5 (2025: below 6) as a phrase, and a name never published as a number has no row. Last
  names: a language model's recall of names common in Germany, as no open register publishes German
  surname counts: the commoner ones from the most to the least common, each with an estimated count
  of 700000 over the square root of its rank, then the rarer ones in no order, sharing the count of
  the last rank
- Licence: the first names: [Datenlizenz Deutschland – Namensnennung
  2.0](https://www.govdata.de/dl-de/by-2-0); the last names: the project's own [LICENSE](LICENSE)
- Attribution: the first names: "Datenquelle: Landeshauptstadt München, Statistisches Amt", the
  licence's name `dl-de/by-2-0` with a link to its text, a link to the dataset, and that the files
  are summed per name and sex; the last names: none required
- Rebuild: `data-import/names-de.py`

## `de_DE/phone-mobile.tsv`, `phone-fixed.tsv`

- Source: the Bundesnetzagentur's [Zugeteilte Rufnummernblöcke (Mobile
  Dienste)](https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/MobileDienste/RNB_csv.csv),
  one row for each block it lists, and its
  [Vorwahlverzeichnis](https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/ONRufnr/_DL/Vorwahlverzeichnis_ONB.zip.zip)
  of 2022-07-27, one row for each Ortsnetzkennzahl it marks active, with its Ortsnetz; a subscriber
  number's length follows section 2.3 of Verfügung 25/2006 and sections 2.3 and 2.4 of the
  Nummernplan Mobile Dienste
- Licence: [Datenlizenz Deutschland – Namensnennung 2.0](https://www.govdata.de/dl-de/by-2-0), which
  the Bundesnetzagentur's [Impressum](https://www.bundesnetzagentur.de/impressum) gives the data it
  publishes as .xls, .xlsx and .csv
- Attribution: "Datenquelle: Bundesnetzagentur", the licence's name and link, a link to both lists,
  and that the data was changed: the lists' blocks and codes are rewritten as rows
- Rebuild: `data-import/phone-de.py`

## `de_DE/word.json` and 4 more

- Files: `de_DE/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `de_DE/sentence.json` reads
- Source: [Wikidata's
  lexemes](https://dumps.wikimedia.org/wikidatawiki/entities/latest-lexemes.json.gz), those in
  German, as the dump of 11 November 2026 held them, less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  German list (CC BY 4.0)
- Licence: CC0
- Attribution: none required
- Rebuild: `data-import/word-de.py`

## `en_US/first-name.tsv`

- Source: [SSA](https://www.ssa.gov/oact/babynames/) baby names, births 1930 to 2020, through
  [hackerb9/ssa-baby-names](https://github.com/hackerb9/ssa-baby-names)
- Licence: public domain
- Attribution: none required
- Rebuild: `data-import/names-us.py`

## `en_US/last-name.tsv`

- Source: Census Bureau surnames occurring 100 or more times, 2010
- Licence: public domain
- Attribution: none required
- Rebuild: `data-import/names-us.py`

## `en_US/word.json` and 4 more

- Files: `en_US/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `en_US/sentence.json` reads
- Source: [WordNet 3.0](https://wordnet.princeton.edu/)'s index files, less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  English list (CC BY 4.0), and the words whose every sense is in WordNet's usage domains of
  disparagement, ethnic slurs or obscenity
- Licence: [WordNet License](https://wordnet.princeton.edu/license-and-commercial-use)
- Attribution: its notice and disclaimer on every copy: `en_US/WORDNET-LICENSE.txt` holds them, the
  `en_US` module embeds the file, and the release archives and packages carry it
- Rebuild: `data-import/word-us.py`

## `et_EE/first-name.tsv`, `last-name.tsv`

- Source: a language model's recall of names common in Estonia. Statistics Estonia's [name
  statistics](https://www.stat.ee/nimed/) are CC BY-SA 4.0, which the project took no data under when
  this table was built, so none of its counts is used yet. The commoner names of each list are ranked
  from the most to the least common, each with an estimated count of 100000 over the square root of its
  rank, then the rarer ones, sharing the count of the last rank. A
  woman's last name is the masculine one with an `a` added where a Russian name ends in `ov` or `ev`
- Licence: the project's own [LICENSE](LICENSE)
- Attribution: none required
- Rebuild: `data-import/names-ee.py`

## `fi_FI/company.json`, `company-coined.json`, `company-word.json`

- Source: the weights of the legal forms of `company.json`: [Statistics
  Finland](https://stat.fi/en)'s table
  [13w6](https://pxdata.stat.fi/PxWeb/pxweb/fi/StatFin/StatFin__yrti/statfin_yrti_pxt_13w6.px/),
  "Yritykset toimialoittain ja oikeudellisen muodon mukaan (oikeudellinen yksikkö)", the enterprises
  of 2024: 295,203 natural persons, 240,524 limited companies (`Oy`), 13,356 limited partnerships
  (`Ky`), 4,726 general partnerships (`Ay`) and 2,811 co-operatives (`osk`); the name shapes and the
  two other files are a language model's recall, as no open bulk file of Finnish company names was
  found, and weigh evenly
- Licence: the weights: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); the rest: the
  project's own [LICENSE](LICENSE)
- Attribution: "Source: Statistics Finland"
- Rebuild: hand-written

## `fi_FI/first-name.tsv`, `last-name.tsv`

- Source: the Digital and Population Data Services Agency ([DVV](https://dvv.fi/en))'s [name
  statistics](https://avoindata.suomi.fi/data/en_GB/dataset/57282ad6-3ab1-48fb-983a-8aba5ff8d29a) of
  4 August 2026: the sheets `Miehet ens` and `Naiset ens` of the Etunimitilasto and `Nimet` of the
  Sukunimitilasto, the Finnish citizens bearing each first given name and last name, for the 2,000
  commonest first names of each sex and the 5,000 commonest last names
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Lähde: Digi- ja väestötietovirasto"
- Rebuild: `data-import/names-fi.py`

## `fi_FI/henkilotunnus.json`

- Source: the weights of its two age bands: [Statistics Finland](https://stat.fi/en)'s ["Finland in
  Figures
  2026"](https://stat.fi/en/news/Finland-in-Figures-2026-ageing-Finland-is-visible-in-numbers) of 11
  June 2026, a population of 5,652,881 at the end of 2025, 1,347,494 of them aged 65 or over
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Source: Statistics Finland"
- Rebuild: —

## `fi_FI/phone-fixed.tsv`, `phone-mobile.tsv`, `phone-subscriber.json`

- Source: Annex 1 and Table 1 of the Finnish Transport and Communications Agency
  ([Traficom](https://www.traficom.fi/en))'s [Regulation 32 V/2025 M on public telephone network
  numbering](https://www.traficom.fi/sites/default/files/media/regulation/M_32_V_2025_EN_0.pdf), in
  force from 2 June 2025: the 13 area codes of the telecommunications areas, the mobile area codes
  whose row says `No` under Available ranges, and the lengths of a national number, the digits after
  the 0 counting the area code: 7 to 9 for a fixed number and 8 to 10 for a mobile one, and the
  subscriber first digit 2 to 8 of section 11
- Licence: none stated, facts only; a regulation of an authority is an official document, which
  section 9 of the Finnish Copyright Act (tekijänoikeuslaki 404/1961) excludes from copyright
- Attribution: none required
- Rebuild: curated

## `fi_FI/word.json` and 4 more

- Files: `fi_FI/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `fi_FI/sentence.json` reads
- Source: [Nykysuomen sanalista
  2024](https://kotus.fi/sanakirjat/kielitoimiston-sanakirja/nykysuomen-sana-aineistot/nykysuomen-sanalista/),
  from the Institute for the Languages of Finland ([Kotus](https://kotus.fi/en)), less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  Finnish list (CC BY 4.0)
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Kotimaisten kielten keskus: Nykysuomen sanalista 2024, CC BY 4.0"
- Rebuild: `data-import/word-fi.py`

## `fr_FR/first-name.tsv`, `last-name.tsv`

- Source: INSEE's open files. First names: the ["Fichier des
  prénoms"](https://www.insee.fr/fr/statistiques/8205621) (`prenoms-2023-nat`), the births of 1944 to
  2023 summed per first name and sex, for the 2,000 commonest of each sex; INSEE groups the names it
  doesn't publish under `_PRENOMS_RARES`, and the script drops that group and the rows without a
  year. Last names: the ["Fichier des noms"](https://www.insee.fr/fr/statistiques/3536630)
  (`noms2008nat`), the births of 1891 to 2000 summed over their decades, for the 5,000 commonest; INSEE
  writes them in capitals without accents. Both tables drop a name with a one-letter part
- Licence: [Licence Ouverte / Open Licence 2.0](https://www.etalab.gouv.fr/licence-ouverte-open-licence/),
  as INSEE's [conditions of reuse](https://www.insee.fr/fr/information/2008466) state
- Attribution: "Source : Insee, Fichier des prénoms, 08/07/2024" and "Source : Insee, Fichier des noms
  de famille, 22/05/2018": INSEE's conditions ask for the date of the data's last update
- Rebuild: `data-import/names-fr.py`

## `fr_FR/nir.json`

- Source: the weights of its three age bands: [INSEE](https://www.insee.fr/fr/statistiques/8721185)'s
  "Bilan démographique 2025", table "Population au 1er janvier par sexe et âge détaillé", the
  population of France on 1 January 2026: 69,081,996 people, 13,852,666 of them under 18, 39,890,078
  aged 18 to 64 and 15,339,252 aged 65 or over
- Licence: [Licence Ouverte / Open Licence 2.0](https://www.etalab.gouv.fr/licence-ouverte-open-licence/),
  as INSEE's [conditions of reuse](https://www.insee.fr/fr/information/2008466) state
- Attribution: "Source : Insee, Bilan démographique 2025"
- Rebuild: —

## `geo/DE/region.tsv`, `municipality.tsv`, `locality.tsv`, `postal-code.tsv`

- Source: [GeoNames](https://www.geonames.org/)' [Postleitzahlen of
  Germany](https://download.geonames.org/export/zip/DE.zip), read for every row that names a Land
  and has an accuracy, which is GeoNames' measure of how exactly it knows the point: the Land in
  German as ISO 3166-2:DE names it, the Kreis with the name GeoNames spells, written as a German
  writes it, the place name within the Kreis, and its point as the mean of its rows' points; Hamburg
  and Berlin are one place each; a locality weighs the postal codes it holds and a Kreis and a Land
  the sum of theirs, a count of postal codes rather than of people, as GeoNames publishes no
  population
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Postal codes from GeoNames, www.geonames.org"
- Rebuild: `data-import/geo-de.py`

## `geo/DE/street.tsv`, and the house-number weights of `address.json`

- Source: a language model's recall of street names common in Germany, as no open register is known
  to cover the country whole; every name weighs the same and belongs to no place, and the weights of
  the house-number lengths are estimates
- Licence: the project's own [LICENSE](LICENSE)
- Attribution: none required
- Rebuild: `data-import/geo-de.py`, hand-written

## `geo/DK/locality.tsv` and 3 more

- Files: `geo/DK/locality.tsv`, `postal-code.tsv`, `street.tsv`, and the house-number weights of
  `address.json`
- Source: [Danmarks
  Adresseregister](https://datafordeler.dk/dataoversigt/danmarks-adresseregister-dar/) (DAR), read
  through Datafordeleren on 2026-11-08 for every house number in force: each postdistrikt with its
  kommune, its postnumre and streets with their addresses, and how many house numbers have each
  length, with a letter or without; each postdistrikt's point is the mean of
  [GeoNames](https://www.geonames.org/)' points for its postnumre, and its population is its
  kommune's in [Statistics Denmark](https://www.dst.dk/en/)'s table
  [FOLK1A](https://www.statbank.dk/FOLK1A), shared out over the kommune's postdistrikter by their
  addresses
- Licence: DAR: [CC BY
  4.0](https://datafordeler.dk/vejledning/brugervilkaar/danmarks-adresseregister-dar/); GeoNames and
  Statistics Denmark: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: Klimadatastyrelsen credited; "Postal codes from GeoNames, www.geonames.org"; "Kilde:
  Danmarks Statistik"
- Rebuild: `data-import/geo-dk.py`

## `geo/DK/region.tsv`, `municipality.tsv`

- Source: [Statistics Denmark](https://www.dst.dk/en/)'s NUTS classification of regions and
  kommuner, and its table [FOLK1A](https://www.statbank.dk/FOLK1A), population in the latest
  quarter, 2026K3
- Licence: [CC BY 4.0](https://www.dst.dk/en/Statistik/brug-statistikken/kildeangivelse)
- Attribution: "Kilde: Danmarks Statistik"
- Rebuild: `data-import/geo-dk.py`

## `geo/FI/locality.tsv` and 3 more

- Files: `geo/FI/locality.tsv`, `postal-code.tsv`, `street.tsv`, and the house-number weights of
  `address.json`
- Source: [Posti](https://www.posti.fi/)'s postcode file `PCF_20261113`, downloaded 2026-11-13, for
  each normal postinumero's postitoimipaikka, and its basic address file `BAF_20261107`, downloaded
  2026-11-13, for each street under a postinumero with its Swedish name, kunta and house number
  ranges; each postinumero is weighted by its residents in [Statistics
  Finland](https://stat.fi/en)'s Paavo table
  [12ey](https://pxdata.stat.fi/PxWeb/pxweb/fi/Postinumeroalueittainen_avoin_tieto/), on 31 December
  2024, and each postitoimipaikka's point is the mean of [GeoNames](https://www.geonames.org/)'
  points for its postinumerot
- Licence: Posti: [terms of
  use](https://www.posti.fi/mzj3zpe8qb7p/1eKbwM2WAEY5AuGi5TrSZ7/c76a865cf5feb2c527a114b8615e9580/posti-postal-code-services-service-description-and-terms-of-use-20150101.pdf),
  which let the files be disclosed to third parties if the recipient is "aware of the terms of use
  for the service as well as the download date"; Statistics Finland and GeoNames: [CC BY
  4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: Posti's terms, quoted under Licence, and the download days, under Source; "Source:
  Statistics Finland"; "Postal codes from GeoNames, www.geonames.org"
- Rebuild: `data-import/geo-fi.py`

## `geo/FI/region.tsv`, `municipality.tsv`

- Source: [Statistics Finland](https://stat.fi/en)'s classification of kunnat into maakunnat on 1
  January 2026, and its table
  [11ra](https://pxdata.stat.fi/PxWeb/pxweb/fi/StatFin/StatFin__vaerak/statfin_vaerak_pxt_11ra.px/),
  population on 31 December 2025
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Source: Statistics Finland"
- Rebuild: `data-import/geo-fi.py`

## `geo/NO/locality.tsv` and 3 more

- Files: `geo/NO/locality.tsv`, `postal-code.tsv`, `street.tsv`, and the house-number weights of
  `address.json`
- Source: [Kartverket](https://www.kartverket.no/)'s address register, [Matrikkelen –
  Adresse](https://kartkatalog.geonorge.no/metadata/matrikkelen-adresse/f7df7a18-b30f-4745-bd64-d0863812350c),
  its national CSV downloaded 2026-10-22, read for every street address (vegadresse): each poststed
  with its centre point and kommune, its postnumre and streets with their addresses, and how many
  house numbers have each length, with a letter or without; poststed populations shared out from
  [Statistics Norway](https://www.ssb.no/) StatBank table
  [07459](https://www.ssb.no/en/statbank/table/07459) by addresses
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: Kartverket named as the source, with a link to the licence; "Kilde: Statistisk
  sentralbyrå"
- Rebuild: `data-import/geo-no.py`

## `geo/NO/region.tsv`, `municipality.tsv`

- Source: [Statistics Norway](https://www.ssb.no/)'s classifications of
  [fylker](https://dataportal.ssb.no/classifications/104/codes) and
  [kommuner](https://dataportal.ssb.no/classifications/131/codes) at the import's date, and StatBank
  table [07459](https://www.ssb.no/en/statbank/table/07459), population on 1 January 2026
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Kilde: Statistisk sentralbyrå"
- Rebuild: `data-import/geo-no.py`

## `geo/SE/locality.tsv`, `postal-code.tsv`

- Source: [GeoNames](https://www.geonames.org/) postal codes for SE; populations from SCB tätorter
  and småorter 2023
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); CC0 1.0
- Attribution: "Postal codes from GeoNames, www.geonames.org"
- Rebuild: `data-import/geo-se.py`

## `geo/SE/region.tsv`, `municipality.tsv`

- Source: [SCB](https://www.scb.se/) län and kommun codes 2026 and population 2024
- Licence: [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/)
- Attribution: none required
- Rebuild: `data-import/geo-se.py`

## `geo/SE/street.tsv`

- Source: [Trafikverket NVDB](https://www.trafikverket.se/) Gatunamn, through the open API, placed
  with [SCB](https://www.scb.se/)'s tätort areas 2023
- Licence: CC0 1.0
- Attribution: none required
- Rebuild: `data-import/geo-se.py`

## `geo/US/postal-code.tsv`, `street.tsv`

- Source: Census Bureau ZCTA to place relationships 2020 and TIGER/Line 2025 address ranges and
  feature names
- Licence: public domain
- Attribution: none required
- Rebuild: `data-import/geo-us.py`

## `geo/US/region.tsv`, `municipality.tsv`, `locality.tsv`

- Source: [Census Bureau](https://www.census.gov/) Gazetteer 2026 and population estimates 2025;
  which places are state capitals from [Wikidata](https://www.wikidata.org/)
- Licence: [public domain](https://www.usa.gov/government-works); CC0 1.0
- Attribution: none required
- Rebuild: `data-import/geo-us.py`

## `misc/currency.tsv`

- Source: [datasets/currency-codes](https://github.com/datasets/currency-codes); symbols from
  [Unicode CLDR](https://github.com/unicode-org/cldr) `en.xml` and `root.xml`
- Licence: PDDL 1.0; [Unicode License v3](https://www.unicode.org/license.txt)
- Attribution: CLDR: "Copyright © 1991-2025 Unicode, Inc. Unicode and the Unicode Logo are
  registered trademarks of Unicode, Inc. in the United States and other countries."
- Rebuild: `data-import/currency.py`

## `misc/httpmethod.tsv`

- Source: [IANA HTTP Method Registry](https://www.iana.org/assignments/http-methods/)
- Licence: [public domain](https://www.iana.org/help/licensing-terms)
- Attribution: none required
- Rebuild: `data-import/httpmethod.py`

## `misc/httpstatus.tsv`

- Source: [IANA HTTP Status Code Registry](https://www.iana.org/assignments/http-status-codes/)
- Licence: [public domain](https://www.iana.org/help/licensing-terms)
- Attribution: none required
- Rebuild: `data-import/httpstatus.py`

## `misc/language.tsv`

- Source: [datasets/language-codes](https://github.com/datasets/language-codes), the [Library of
  Congress](https://www.loc.gov/standards/iso639-2/) ISO 639-2 register
- Licence: [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/)
- Attribution: none required
- Rebuild: `data-import/language.py`

## `misc/loglevel.tsv`

- Source: [RFC 5424](https://www.rfc-editor.org/rfc/rfc5424.txt) Table 2, the syslog message
  severities; keywords from the
  [POSIX](https://pubs.opengroup.org/onlinepubs/9799919799/basedefs/syslog.h.html) `<syslog.h>`
  severity macros
- Licence: facts
- Attribution: none required
- Rebuild: `data-import/loglevel.py`

## `misc/mimetype.tsv`

- Source: [mime-db](https://github.com/jshttp/mime-db), the IANA media type registry with filename
  extensions
- Licence: [MIT](https://github.com/jshttp/mime-db/blob/master/LICENSE)
- Attribution: "Copyright (c) 2014 Jonathan Ong, Copyright (c) 2015-2022 Douglas Christopher Wilson"
- Rebuild: `data-import/mimetype.py`

## `misc/port.tsv`

- Source: [IANA Service Name and Transport Protocol Port Number
  Registry](https://www.iana.org/assignments/service-names-port-numbers/)
- Licence: [public domain](https://www.iana.org/help/licensing-terms)
- Attribution: none required
- Rebuild: `data-import/port.py`

## `misc/protocol.tsv`

- Source: [IANA Protocol Numbers](https://www.iana.org/assignments/protocol-numbers/)
- Licence: [public domain](https://www.iana.org/help/licensing-terms)
- Attribution: none required
- Rebuild: `data-import/protocol.py`

## `misc/territory.tsv`

- Source: [datasets/country-codes](https://github.com/datasets/country-codes)
- Licence: [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/)
- Attribution: none required
- Rebuild: `data-import/territory.py`

## `misc/timezone.tsv`

- Source: [IANA tzdb](https://www.iana.org/time-zones) 2026d `zone.tab` and the standard offset of
  each zone; weights from [GeoNames](https://www.geonames.org/) `cities15000` populations
- Licence: [public domain](https://data.iana.org/time-zones/tzdb/LICENSE); [CC BY
  4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Populations from GeoNames, www.geonames.org"
- Rebuild: `data-import/timezone.py`

## `misc/tld.tsv`

- Source: [IANA Root Zone Database](https://www.iana.org/domains/root/db), checked against the root
  zone's [TLD list](https://data.iana.org/TLD/tlds-alpha-by-domain.txt)
- Licence: [public domain](https://www.iana.org/help/licensing-terms)
- Attribution: none required
- Rebuild: `data-import/tld.py`

## `misc/useragent.tsv`

- Source: [top-user-agents](https://github.com/microlinkhq/top-user-agents) desktop and mobile lists
- Licence: [MIT](https://github.com/microlinkhq/top-user-agents/blob/master/LICENSE.md)
- Attribution: "Copyright © 2020 Kiko Beats"
- Rebuild: `data-import/useragent.py`

## `nb_NO/birth-day.tsv`, `nb_NO/individ.tsv`

- Source: curated from
  [Skatteetaten](https://www.skatteetaten.no/person/folkeregister/fodsel-og-navnevalg/barn-fodt-i-norge/fodselsnummer/)'s
  rules for the fødselsnummer: the 365 days of a year with 80 added to the month, which marks its
  [synthetic
  series](https://skatteetaten.github.io/folkeregisteret-api-dokumentasjon/ny-versjon-av-syntetisk-folkeregister/),
  and the odd digits for a man's last individ digit and the even for a woman's
- Licence: none stated, facts only
- Attribution: none required
- Rebuild: —

## `nb_NO/company-word.tsv`, `nb_NO/company-coined.json`, `nb_NO/company.json`

- Source: [Enhetsregisteret](https://data.brreg.no/enhetsregisteret/api/docs/index.html) at the
  [Brønnøysund Register Centre](https://www.brreg.no/), its bulk file of the units downloaded
  2026-10-22, read for counts: the active units of each legal form (not the bankrupt, those being
  wound up or those forcibly dissolved) with the trade words that 100 or more of their names carry,
  the shapes of their names, and the words that 60 or more of their names begin with; the endings of
  `company-coined.json`, and the poststeder a name draws, come from `geo/NO`'s place names, under
  `geo/NO`'s sections
- Licence: [NLOD 2.0](https://data.norge.no/nlod/en/2.0); `geo/NO`'s CC BY 4.0
- Attribution: "Contains data under the Norwegian licence for Open Government data (NLOD)
  distributed by Brønnøysundregistrene", with a link to the licence, and that the register is
  reduced to counts and words; for a name holding a poststed, `geo/NO`'s attribution too
- Rebuild: `data-import/company-no.py`

## `nb_NO/first-name.tsv`, `last-name.tsv`

- Source: [Statistics Norway](https://www.ssb.no/) StatBank tables
  [10501](https://www.ssb.no/en/statbank/table/10501) and
  [12891](https://www.ssb.no/en/statbank/table/12891): persons bearing each first name and last name
  at the end of 2025, for the names that 200 or more bear
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Kilde: Statistisk sentralbyrå"
- Rebuild: `data-import/names-no.py`

## `nb_NO/phone-mobile.tsv`, `phone-fixed.tsv`

- Source: [Nkom](https://www.nkom.no/)'s [Samla norsk nummerplan for telefoni mm.
  (E.164)](https://data.norge.no/en/datasets/ecbfe08c-24dc-3eb6-80b5-4a41de967e4b), downloaded
  2026-10-22: the four-digit prefixes whose ten thousand numbers are all assigned to providers as
  mobile or fixed
- Licence: [NLOD 2.0](https://data.norge.no/nlod/en/2.0)
- Attribution: "Contains data under the Norwegian licence for Open Government data (NLOD)
  distributed by Nasjonal kommunikasjonsmyndighet (Nkom)", with a link to the licence, and that the
  register is reduced to four-digit prefixes
- Rebuild: `data-import/phone-no.py`

## `nb_NO/word.json` and 4 more

- Files: `nb_NO/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `nb_NO/sentence.json` reads
- Source: [Norsk ordbank bokmål
  2005](https://www.nb.no/sprakbanken/ressurskatalog/oai-nb-no-sbr-5/), from Språkbanken at the
  National Library of Norway, less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  Norwegian list (CC BY 4.0)
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Språkbanken, Nasjonalbiblioteket: Norsk ordbank bokmål 2005, CC BY 4.0"
- Rebuild: `data-import/word-no.py`

## `nn_NO/word.json` and 4 more

- Files: `nn_NO/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `nn_NO/sentence.json` reads
- Source: [Norsk ordbank nynorsk
  2012](https://www.nb.no/sprakbanken/ressurskatalog/oai-nb-no-sbr-41/), from Språkbanken at the
  National Library of Norway, less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  Norwegian list (CC BY 4.0)
- Licence: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Attribution: "Språkbanken, Nasjonalbiblioteket: Norsk ordbank nynorsk 2012, CC BY 4.0"
- Rebuild: `data-import/word-no.py`

## `sv_SE/bankaccount.tsv`

- Source: the clearing numbers and banks of [Bankinfrastruktur i Sverige
  AB](https://www.bankinfrastruktur.se/framtidens-betalningsinfrastruktur/iban-och-svenskt-nationellt-kontonummer)'s
  list
  [clearingnummertabell-for-nedladdning.csv](https://www.bankinfrastruktur.se/media/sqfjsvkp/clearingnummertabell-for-nedladdning.csv),
  downloaded 2026-11-04, and the check each bank demands, from Bankgirot's "Bankernas kontonummer"
  of 2024-02-22; the 15 clearing numbers whose check the document leaves out are not shipped
- Licence: none stated, facts only; the list disclaims its accuracy
- Attribution: none required
- Rebuild: `data-import/bankaccount-se.py`

## `sv_SE/birth-number.tsv` and 9 more

- Files: `sv_SE/birth-number.tsv`, `sv_SE/title.tsv`, `nb_NO/title.tsv`, `nn_NO/title.tsv`,
  `da_DK/title.tsv`, `fi_FI/title.tsv`, `de_DE/title.tsv`, `et_EE/title.tsv`, `en_US/title.tsv`,
  `misc/car.tsv`
- Source: curated; Skatteverket's test birth numbers are facts, and `car.tsv` waits for an
  international source (goal 3.2)
- Licence: —
- Attribution: —
- Rebuild: —

## `sv_SE/bloodtype.tsv`

- Source: curated: Sweden's blood type shares as sv.wikipedia's
  [Blodgrupp](https://sv.wikipedia.org/wiki/Blodgrupp) gives them, undated; it gives `A-` as 7–8
  percent and the table takes 7, so the shares sum to 100; the spelling `0` is that of
  [1177](https://www.1177.se/liv--halsa/sa-fungerar-kroppen/blodgrupper/)
- Licence: facts
- Attribution: none required
- Rebuild: —

## `sv_SE/car.tsv`

- Source: [Mobility Sweden](https://mobilitysweden.se/)'s [report of new registrations, December
  2025](https://mobilitysweden.se/statistik/Nyregistreringar_per_manad_1/nyregistreringar-2025/definitiva-nyregistreringar-under-2025),
  the sheet "PB - Fabrikat och modell": the passenger car models with 500 or more registrations in
  2025, shipped as facts
- Licence: no open licence; the source asks to be named
- Attribution: "Källa: Mobility Sweden"
- Rebuild: `data-import/car-se.py`

## `sv_SE/company-word.tsv`, `sv_SE/company.json`

- Source: [Bolagsverket](https://www.bolagsverket.se/) company register bulk file, read for counts:
  the trade words that 100 or more active companies' names carry, and the wordings and shapes of
  their names; legal-form weights from the
  [statistics](https://www.bolagsverket.se/statistik/ftgstat_oppna.csv), new registrations 2025
- Licence: register: none stated, facts only; statistics: [CC BY 2.5
  SE](https://creativecommons.org/licenses/by/2.5/se/)
- Attribution: "Källa: Bolagsverket"
- Rebuild: `data-import/company-se.py`

## `sv_SE/first-name.tsv`, `last-name.tsv`

- Source: [SCB](https://www.scb.se/) names with at least two bearers, 31 December 2022
- Licence: CC0 1.0
- Attribution: "Källa: SCB"
- Rebuild: `data-import/names-se.py`

## `sv_SE/holiday.tsv`

- Source: curated: the thirteen public holidays that [lag (1989:253) om allmänna
  helgdagar](https://data.riksdagen.se/dokument/sfs-1989-253.text) 1 § names, less its entry for
  every Sunday; the codes are ours
- Licence: facts
- Attribution: none required
- Rebuild: —

## `sv_SE/licenseplate-letters.tsv`

- Source: every combination of three of the 23 letters that Swedish plates use, from sv.wikipedia's
  [Registreringsskyltar i Sverige](https://sv.wikipedia.org/wiki/Registreringsskyltar_i_Sverige),
  less the 99 that
  [Transportstyrelsen](https://www.transportstyrelsen.se/sv/vagtrafik/fordon/aga-kopa-eller-salja-fordon/registreringsskyltar/byte-av-registreringsnummer/sparrade-bokstavskombinationer/)'s
  page listed as blocked when it was updated on 2026-03-31 and no longer lists since its update of
  2026-08-30
- Licence: none stated, facts only
- Attribution: none required
- Rebuild: `data-import/licenseplate-se.py`

## `sv_SE/month.tsv`, `sv_SE/weekday.tsv`, `sv_SE/territory.tsv`

- Source: [Unicode CLDR](https://github.com/unicode-org/cldr) 48.0.0, the Swedish stand-alone full
  month and weekday names and territory names, through
  [cldr-json](https://github.com/unicode-org/cldr-json); `territory.tsv` holds one row for each
  `misc/territory.tsv` alpha-2 code
- Licence: [Unicode License v3](https://www.unicode.org/license.txt)
- Attribution: "Copyright © 2015-2024 Unicode, Inc."
- Rebuild: `data-import/cldr.py`

## `sv_SE/organisationsnummer.json`

- Source: group weights from [Bolagsverket](https://www.bolagsverket.se/statistik/ftgstat_oppna.csv)
  company statistics, new registrations 2025: 5 is AB and FL, 9 is HB and KB, 7 and 8 share BRF, EK
  and I evenly, which is our choice
- Licence: [CC BY 2.5 SE](https://creativecommons.org/licenses/by/2.5/se/)
- Attribution: "Källa: Bolagsverket"
- Rebuild: —

## `sv_SE/personnummer-date.json`, `personnummer-12.json`

- Source: band weights from [SCB](https://www.scb.se/) Statistikdatabasen, table
  BE0101A/BefolkningNy, at the end of 2024, computed from single-year ages; the bands are ours
- Licence: CC0 1.0
- Attribution: "Källa: SCB"
- Rebuild: —

## `sv_SE/phone-area.tsv`

- Source: the 264 geographic area codes of
  [PTS](https://pts.se/globalassets/globala-block/nummertillstand/the-swedish-numbering-plan-for-telephony-according-to-itu---2024-01-08.pdf)'s
  numbering plan of 2024-01-08, each with its place and the subscriber lengths the plan allows
- Licence: none stated, facts only
- Attribution: "Källa: PTS"
- Rebuild: `data-import/phone-area-se.py`

## `sv_SE/sex.tsv` and 7 more

- Files: `sv_SE/sex.tsv`, `nb_NO/sex.tsv`, `da_DK/sex.tsv`, `fi_FI/sex.tsv`, `de_DE/sex.tsv`,
  `et_EE/sex.tsv`, `fr_FR/sex.tsv`, `en_US/sex.tsv`
- Source: the people of each sex: [SCB](https://www.scb.se/) table BefolkningR1860N at the end of
  2024; [Statistics Norway](https://www.ssb.no/) StatBank table
  [07459](https://www.ssb.no/en/statbank/table/07459) on 1 January 2026; [Statistics
  Denmark](https://www.dst.dk/en/)'s table [FOLK1A](https://www.statbank.dk/FOLK1A) on 1 January
  2026; [Eurostat](https://ec.europa.eu/eurostat)'s table
  [demo_pjan](https://ec.europa.eu/eurostat/databrowser/view/demo_pjan/default/table) for Germany,
  Estonia and France on 1 January 2025, a figure Eurostat marks provisional for France; the
  [Census Bureau](https://www.census.gov/)'s national population estimates by age and sex, vintage 2024,
  for 1 July 2024; [Statistics Finland](https://stat.fi/en)'s table
  [11rb](https://pxdata.stat.fi/PxWeb/pxweb/en/StatFin/StatFin__vaerak/statfin_vaerak_pxt_11rb.px/),
  the population of 31 December 2025; the names are ours
- Licence: CC0 1.0; [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) for Norway, Denmark,
  Germany, Estonia, France and Finland; public domain
- Attribution: "Källa: SCB"; "Kilde: Statistisk sentralbyrå"; "Kilde: Danmarks Statistik"; "Source:
  Eurostat"; "Source: Statistics Finland"; none required
- Rebuild: `data-import/sex-population.py`

## `sv_SE/sni.tsv`

- Source: [SCB](https://www.scb.se/)'s [SNI 2025](https://www.scb.se/globalassets/sni-2025.xlsx),
  the sheet "Detaljgrupp (Femsiffer)"
- Licence: none stated on the classification pages; SCB's [open
  data](https://www.scb.se/vara-tjanster/oppna-data/) is CC0 1.0 and lists other datasets
- Attribution: "Källa: SCB", which SCB recommends
- Rebuild: `data-import/sni-se.py`

## `sv_SE/ssyk.tsv`

- Source: [SCB](https://www.scb.se/)'s [SSYK
  2012](https://www.scb.se/contentassets/0c0089cc085a45d49c1dc83923ad933a/ssyk-2012-koder.xlsx), the
  sheet "4-siffer"
- Licence: none stated on the classification pages; SCB's [open
  data](https://www.scb.se/vara-tjanster/oppna-data/) is CC0 1.0 and lists other datasets
- Attribution: "Källa: SCB", which SCB recommends
- Rebuild: `data-import/ssyk-se.py`

## `sv_SE/word.json` and 4 more

- Files: `sv_SE/word.json`, `word-noun.tsv`, `word-verb.tsv`, `word-adj.tsv`, `word-adv.tsv` and the
  words `sv_SE/sentence.json` reads
- Source: [SALDO's morphology](https://sprakbanken.se/en/resources/saldom), less the words on
  [LDNOOBW](https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words)'s
  Swedish list (CC BY 4.0)
- Licence: CC BY 4.0
- Attribution: "Språkbanken (2017). SALDO's morphology. https://doi.org/10.23695/agcm-ny22"
- Rebuild: `data-import/word-se.py`
