# Data licenses

Every shipped dataset, its source, its licence and the attribution it asks for. A
`data-import/` script rebuilds each sourced table; a curated one is hand-written.
The scripts run through `docker compose run --rm data-import data-import/<name>.py`
and cache their downloads under `data-import/cache/`; `geo-se.py` needs a
Trafikverket API key, free at [data.trafikverket.se](https://data.trafikverket.se/),
in `TRAFIKVERKET_API_KEY` or a `--key-file`; `geo-us.py` fetches two TIGER/Line
files per county it ships, a few hundred megabytes in all.

| Table | Source | Licence | Attribution | Rebuild |
|-------|--------|---------|-------------|---------|
| `geo/SE/locality.tsv`, `postal-code.tsv` | [GeoNames](https://www.geonames.org/) postal codes for SE; populations from SCB tätorter 2023 | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); CC0 1.0 | "Postal codes from GeoNames, www.geonames.org" | `data-import/geo-se.py` |
| `geo/SE/region.tsv`, `municipality.tsv` | [SCB](https://www.scb.se/) län and kommun codes 2026 and population 2024 | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) | none required | `data-import/geo-se.py` |
| `geo/SE/street.tsv` | [Trafikverket NVDB](https://www.trafikverket.se/) Gatunamn, through the open API | CC0 1.0 | none required | `data-import/geo-se.py` |
| `geo/US/region.tsv`, `municipality.tsv`, `locality.tsv` | [Census Bureau](https://www.census.gov/) Gazetteer 2026 and population estimates 2025 | [public domain](https://www.usa.gov/government-works) | none required | `data-import/geo-us.py` |
| `geo/US/postal-code.tsv`, `street.tsv` | Census Bureau ZCTA to place relationships 2020 and TIGER/Line 2025 address ranges and feature names | public domain | none required | `data-import/geo-us.py` |
| `misc/country.tsv` | [datasets/country-codes](https://github.com/datasets/country-codes) | [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/) | none required | `data-import/country.py` |
| `misc/currency.tsv` | [datasets/currency-codes](https://github.com/datasets/currency-codes); symbols from [Unicode CLDR](https://github.com/unicode-org/cldr) `en.xml` and `root.xml` | PDDL 1.0; [Unicode License v3](https://www.unicode.org/license.txt) | CLDR: "Copyright © 1991-2025 Unicode, Inc. Unicode and the Unicode Logo are registered trademarks of Unicode, Inc. in the United States and other countries." | `data-import/currency.py` |
| `misc/httpstatus.tsv` | curated (IANA HTTP status codes are facts) | — | — | — |
| `misc/language.tsv` | curated (ISO 639-1 codes are facts) | — | — | — |
| `misc/mimetype.tsv` | curated (IANA media types are facts) | — | — | — |

Every other category is hand-written JSON under [`data/`](data), MIT like the code.
