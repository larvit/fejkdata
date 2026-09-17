# Data licenses

Every shipped dataset, its source, its licence and the attribution it asks for. A
`data-import/` script rebuilds each sourced table; a curated one is hand-written.

| Table | Source | Licence | Attribution | Rebuild |
|-------|--------|---------|-------------|---------|
| `misc/country.tsv` | [datasets/country-codes](https://github.com/datasets/country-codes) | [PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/) | none required | `data-import/country.py` |
| `misc/currency.tsv` | [datasets/currency-codes](https://github.com/datasets/currency-codes); symbols from [Unicode CLDR](https://github.com/unicode-org/cldr) `en.xml` and `root.xml` | PDDL 1.0; [Unicode License v3](https://www.unicode.org/license.txt) | CLDR: "Copyright © 1991-2025 Unicode, Inc. Unicode and the Unicode Logo are registered trademarks of Unicode, Inc. in the United States and other countries." | `data-import/currency.py` |
| `misc/httpstatus.tsv` | curated (IANA HTTP status codes are facts) | — | — | — |
| `misc/language.tsv` | curated (ISO 639-1 codes are facts) | — | — | — |
| `misc/mimetype.tsv` | curated (IANA media types are facts) | — | — | — |

Every other category is hand-written JSON under [`data/`](data), MIT like the code.
