# Data licenses

Every shipped dataset, its source, its licence and the attribution it asks for. A
`data-import/` script rebuilds each sourced table, run as the README's
[Development](README.md#development) section says; a curated one is hand-written.

| Table | Source | Licence | Attribution | Rebuild |
|-------|--------|---------|-------------|---------|
| `geo/SE/locality.tsv`, `postal-code.tsv` | [GeoNames](https://www.geonames.org/) postal codes for SE; populations from SCB tätorter 2023 | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); CC0 1.0 | "Postal codes from GeoNames, www.geonames.org" | `data-import/geo-se.py` |
| `geo/SE/region.tsv`, `municipality.tsv` | [SCB](https://www.scb.se/) län and kommun codes 2026 and population 2024 | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) | none required | `data-import/geo-se.py` |
| `geo/SE/street.tsv` | [Trafikverket NVDB](https://www.trafikverket.se/) Gatunamn, through the open API | CC0 1.0 | none required | `data-import/geo-se.py` |
| `geo/US/region.tsv`, `municipality.tsv`, `locality.tsv` | [Census Bureau](https://www.census.gov/) Gazetteer 2026 and population estimates 2025 | [public domain](https://www.usa.gov/government-works) | none required | `data-import/geo-us.py` |
| `geo/US/postal-code.tsv`, `street.tsv` | Census Bureau ZCTA to place relationships 2020 and TIGER/Line 2025 address ranges and feature names | public domain | none required | `data-import/geo-us.py` |
| `sv_SE/first-name.tsv`, `last-name.tsv` | [SCB](https://www.scb.se/) names with at least two bearers, 31 December 2022 | CC0 1.0 | "Källa: SCB" | `data-import/names-se.py` |
| `en_US/first-name.tsv` | [SSA](https://www.ssa.gov/oact/babynames/) baby names, births 1930 to 2020, through [hackerb9/ssa-baby-names](https://github.com/hackerb9/ssa-baby-names) | public domain | none required | `data-import/names-us.py` |
| `en_US/last-name.tsv` | Census Bureau surnames occurring 100 or more times, 2010 | public domain | none required | `data-import/names-us.py` |
| `sv_SE/sex.tsv`, `sv_SE/birth-number.tsv`, `en_US/sex.tsv` | curated (Skatteverket's test birth numbers are facts) | — | — | — |
| `misc/country.tsv` | [datasets/country-codes](https://github.com/datasets/country-codes) | [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/) | none required | `data-import/country.py` |
| `misc/currency.tsv` | [datasets/currency-codes](https://github.com/datasets/currency-codes); symbols from [Unicode CLDR](https://github.com/unicode-org/cldr) `en.xml` and `root.xml` | PDDL 1.0; [Unicode License v3](https://www.unicode.org/license.txt) | CLDR: "Copyright © 1991-2025 Unicode, Inc. Unicode and the Unicode Logo are registered trademarks of Unicode, Inc. in the United States and other countries." | `data-import/currency.py` |
| `misc/httpstatus.tsv` | curated (IANA HTTP status codes are facts) | — | — | — |
| `misc/language.tsv` | curated (ISO 639-1 codes are facts) | — | — | — |
| `misc/mimetype.tsv` | curated (IANA media types are facts) | — | — | — |

Every other category is hand-written JSON under [`data/`](data), MIT like the code.
