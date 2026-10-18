# todo

## Scoring

`Score = 100 * V / (2**S + 8)`

`Bar = 0`

`Next ID = 337`

## Items

| ID | Release | Blocked by | Exempt | Item | L | E | I | R | S | Moves | Score |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 227 | 0.2.0 |  |  | **Add a README quick start per task, for the bulk fixture author and the Go test author, each rendering one person whose name, sex, id, email and address agree.** | 1 | 2 | 1 | 2 | 2 | 1 by 4, 5 by 3 | 10.8 |
| 230 | 0.2.0 |  |  | **Add `sv_SE.organisationsnummer`, whose first digit is its legal form's group digit and whose last is a Luhn check digit, and `sv_SE.vat`, `SE` + the number + `01`.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 5, 4 by 4 | 7.0 |
| 157 | 0.2.0 |  |  | **Show in the README's CLI section how PowerShell and cmd.exe pipe a template, one line per shell family.** | 1 | 2 | 1 | 2 | 1 | 1 by 2 | 6.2 |
| 231 | 0.2.0 |  |  | **Draw `sv_SE.company`'s legal form weighted by Bolagsverket's 2025 registrations, and its names from patterns derived from Bolagsverket's company register.** | 2 | 2 | 2 | 4 | 3 | 1 by 2, 2.1 by 3, 4.2 by 4 | 6.0 |
| 234 | 0.2.0 |  |  | **Draw `sv_SE.phone` from every area code in PTS's numbering plan, grouped as Språkrådet writes it, and add its international form, `+46 70 123 45 67`.** | 2 | 2 | 2 | 4 | 3 | 1 by 2, 2.1 by 3, 4.2 by 4 | 6.0 |
| 202 | 0.2.0 |  |  | **Add `nb_NO` with its given names by sex, last names and titles, and `nb_NO.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 292 | 0.2.0 | 202 |  | **Add `nb_NO`'s national personal id, the fødselsnummer, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 291 | 0.2.0 | 202 |  | **Add `nb_NO.phone`, drawn from Norway's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 293 | 0.2.0 | 202 |  | **Add `nb_NO.company`, `nb_NO.email` and `nb_NO.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 335 | 0.2.0 |  |  | **Merge every Renovate update to a build, test or release tool once the gate passes, and release nothing for it.** | 3 | 2 | 2 | 5 | 2 | 1.1 by 2 | 2.6 |
| 289 | 0.2.0 | 202 |  | **Add `geo/NO` with Norway's places and postal codes, and `nb_NO.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 336 | 0.2.0 | 335 |  | **Release a patch from CI for every update to what a release carries, once the gate passes with no other change: Go, Go's minor included, any module `go.mod` requires, and GoReleaser.** | 3 | 3 | 3 | 7 | 5 | 1.1 by 6 | 2.3 |
| 290 | 0.2.0 | 289 |  | **Give `geo/NO`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 169 | 0.3.0 |  | question | **Reword the decision "A parent row with no child row is a load error" to say what each geo import does with such a row: drop it, or exit where `geo-se.py` would lose a kommun or a län.** | 1 | 2 | 1 | 2 | 1 | 3.1 by 2 | 1.0 |
| 252 | 0.3.0 |  |  | **Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.** | 1 | 3 | 1 | 3 | 1 | 1.1 by 4 | 6.2 |
| 210 | 0.3.0 |  |  | **Add `fr_FR` with its given names by sex, last names and titles, and `fr_FR.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 273 | 0.3.0 | 210 |  | **Add `fr_FR`'s national personal id, the NIR, the numéro de sécurité sociale, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 206 | 0.3.0 |  |  | **Add `zh_CN` with its given names by sex, last names and titles, and `zh_CN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 277 | 0.3.0 | 206 |  | **Add `zh_CN`'s national personal id, the 18-digit resident identity card number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 208 | 0.3.0 |  |  | **Add `es_MX` with its given names by sex, last names and titles, and `es_MX.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 282 | 0.3.0 | 208 |  | **Add `es_MX`'s national personal id, the CURP, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 212 | 0.3.0 |  |  | **Add `pt_BR` with its given names by sex, last names and titles, and `pt_BR.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 287 | 0.3.0 | 212 |  | **Add `pt_BR`'s national personal id, the CPF, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 203 | 0.3.0 |  |  | **Add `da_DK` with its given names by sex, last names and titles, and `da_DK.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 297 | 0.3.0 | 203 |  | **Add `da_DK`'s national personal id, the CPR number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 204 | 0.3.0 |  |  | **Add `fi_FI` with its given names by sex, last names and titles, and `fi_FI.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 302 | 0.3.0 | 204 |  | **Add `fi_FI`'s national personal id, the henkilötunnus, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 207 | 0.3.0 |  |  | **Add `hi_IN` with its given names by sex, last names and titles, and `hi_IN.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 306 | 0.3.0 | 207 |  | **Add `hi_IN`'s national personal id, the Aadhaar number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 209 | 0.3.0 |  | decision | **Add `ar_EG` with its given names by sex, last names and titles, and `ar_EG.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 310 | 0.3.0 | 209 |  | **Add `ar_EG`'s national personal id, the national ID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 211 | 0.3.0 |  |  | **Add `bn_BD` with its given names by sex, last names and titles, and `bn_BD.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 314 | 0.3.0 | 211 |  | **Add `bn_BD`'s national personal id, the NID number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 213 | 0.3.0 |  |  | **Add `id_ID` with its given names by sex, last names and titles, and `id_ID.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 318 | 0.3.0 | 213 |  | **Add `id_ID`'s national personal id, the NIK, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 214 | 0.3.0 |  | decision | **Add `ur_PK` with its given names by sex, last names and titles, and `ur_PK.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 322 | 0.3.0 | 214 |  | **Add `ur_PK`'s national personal id, the CNIC number, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 205 | 0.3.0 |  |  | **Add `is_IS` with its given names by sex, last names and titles, and `is_IS.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 327 | 0.3.0 | 205 |  | **Add `is_IS`'s national personal id, the kennitala, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 217 | 0.3.0 |  |  | **Add `de_DE` with its given names by sex, last names and titles, and `de_DE.person`.** | 2 | 2 | 1 | 3 | 3 | 1 by 2, 2.1 by 4 | 5.8 |
| 332 | 0.3.0 | 217 |  | **Add `de_DE`'s national personal id, the Steuer-Identifikationsnummer, with its check digits.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 2.1 by 3, 4 by 3 | 5.9 |
| 222 | 0.3.0 |  |  | **Write `--format sql` and `SQLInsert` rows as multi-row INSERTs, so a bulk fixture loads fast.** | 2 | 1 | 2 | 3 | 3 | 1 by 3 | 5.8 |
| 34 | 0.3.0 |  |  | **Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 4 by 4 | 1.8 |
| 154 | 0.3.0 | 34 |  | **Add `en_US.routing`, an ABA routing number over the Fed's prefix ranges with its check digit drawn by `{aba()}`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2.1 by 3, 4 by 3 | 5.3 |
| 243 | 0.3.0 |  |  | **Add `en_US.ein`, drawing only the prefixes the IRS assigns.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 2.1 by 3, 4 by 3 | 5.3 |
| 158 | 0.3.0 |  |  | **Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.** | 1 | 1 | 1 | 1 | 2 | 1 by 2 | 5.1 |
| 163 | 0.3.0 |  |  | **Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.** | 2 | 1 | 2 | 3 | 2 | 1 by 2 | 5.1 |
| 241 | 0.3.0 |  |  | **Give `sv_SE.person` a `birthdate`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2.1 by 4 | 5.1 |
| 155 | 0.3.0 | 34 |  | **Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.** | 1 | 1 | 2 | 2 | 2 | 1 by 1, 2 by 2, 4 by 3 | 5.0 |
| 236 | 0.3.0 |  |  | **Add `sv_SE.licenseplate`, `ABC 123` and `ABC 12A`, never drawing a combination Transportstyrelsen blocks.** | 2 | 2 | 2 | 4 | 2 | 1 by 1, 2.1 by 3, 4 by 2 | 5.0 |
| 22 | 0.3.0 |  |  | **Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.** | 2 | 1 | 2 | 3 | 3 | 1 by 2, 4.3 by 6 | 4.8 |
| 235 | 0.3.0 |  |  | **Add `sv_SE.bankgiro` and `sv_SE.plusgiro` with their Luhn check digit, and `sv_SE.bankaccount` with a clearing number and the check digit its bank's account type demands.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 2.1 by 5, 4 by 5 | 4.8 |
| 256 | 0.3.0 |  |  | **Add `misc.bic`, built by its structure rules.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2, 4 by 2 | 4.8 |
| 245 | 0.3.0 |  |  | **Draw `en_US.phone` from every area code NANPA has in service.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2, 4.2 by 3 | 4.5 |
| 220 | 0.3.0 |  |  | **Suggest the nearest name where a path segment names nothing: `no entry "adress"; did you mean "address"?`.** | 1 | 1 | 1 | 1 | 3 | 1 by 2, 6 by 5 | 4.5 |
| 272 | 0.3.0 | 210 |  | **Add `fr_FR.phone`, drawn from France's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 276 | 0.3.0 | 206 |  | **Add `zh_CN.phone`, drawn from China's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 281 | 0.3.0 | 208 |  | **Add `es_MX.phone`, drawn from Mexico's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 286 | 0.3.0 | 212 |  | **Add `pt_BR.phone`, drawn from Brazil's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 296 | 0.3.0 | 203 |  | **Add `da_DK.phone`, drawn from Denmark's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 301 | 0.3.0 | 204 |  | **Add `fi_FI.phone`, drawn from Finland's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 305 | 0.3.0 | 207 |  | **Add `hi_IN.phone`, drawn from India's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 309 | 0.3.0 | 209 |  | **Add `ar_EG.phone`, drawn from Egypt's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 313 | 0.3.0 | 211 |  | **Add `bn_BD.phone`, drawn from Bangladesh's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 317 | 0.3.0 | 213 |  | **Add `id_ID.phone`, drawn from Indonesia's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 321 | 0.3.0 | 214 |  | **Add `ur_PK.phone`, drawn from Pakistan's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 326 | 0.3.0 | 205 |  | **Add `is_IS.phone`, drawn from Iceland's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 331 | 0.3.0 | 217 |  | **Add `de_DE.phone`, drawn from Germany's numbering plan.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 3 | 4.5 |
| 262 | 0.3.0 |  |  | **Add `misc.bloodtype`, weighted by its global distribution.** | 1 | 2 | 1 | 2 | 1 | 1 by 1, 2 by 1, 4.2 by 1 | 4.4 |
| 27 | 0.3.0 |  |  | **Merge `en_US.ip` and `sv_SE.ip`, today byte-identical, into one `misc.ip`.** | 2 | 1 | 2 | 3 | 2 | 2 by 1 | 0.9 |
| 15 | 0.3.0 | 27 |  | **Add `misc.ip`'s documentation, private and CIDR forms, `misc.ipv6` in `2001:db8::/32`, and a locally administered `misc.mac`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4.3 by 4, 2 by 1 | 4.3 |
| 35 | 0.3.0 |  |  | **Add `{base64url(n)}` for JWT shapes.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 4.3 |
| 265 | 0.3.0 |  |  | **Add `misc.commit` messages, branch names and `sha`s.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 2 | 4.3 |
| 257 | 0.3.0 |  |  | **Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 2 by 3, 4 by 3 | 4.2 |
| 239 | 0.3.0 |  |  | **Add `sv_SE.month` and `sv_SE.weekday` names from CLDR, and `sv_SE.holiday` from lag 1989:253.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 4, 4.2 by 2 | 4.2 |
| 244 | 0.3.0 |  |  | **Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.** | 2 | 2 | 2 | 4 | 4 | 1 by 2, 2.1 by 4, 4.2 by 3 | 4.2 |
| 250 | 0.3.0 |  |  | **Add `en_US.month` and `en_US.weekday` names from CLDR, and `en_US.holiday`, the federal holidays.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 4, 4.2 by 2 | 4.2 |
| 24 | 0.3.0 |  | defect | **Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 4 by 3 | 4.0 |
| 88 | 0.3.0 |  |  | **Add a 12-digit `YYYYMMDDNNNC` form of `sv_SE.personnummer` and `sv_SE.samordningsnummer`, the form Skatteverket's testpersonnummer series and many systems store.** | 1 | 1 | 2 | 2 | 3 | 1 by 1, 2.1 by 3, 4 by 3 | 3.9 |
| 232 | 0.3.0 |  |  | **Add `sv_SE.sni`, a table of SCB's SNI 2025 industry codes with their Swedish labels.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3, 4.2 by 3 | 3.9 |
| 233 | 0.3.0 |  |  | **Add `sv_SE.ssyk`, a table of SCB's SSYK 2012 occupation codes with their Swedish labels.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3, 4.2 by 3 | 3.9 |
| 31 | 0.3.0 |  |  | **Rename `misc.territory.country` to `sovereign`.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 3.3 by 4 | 3.8 |
| 240 | 0.3.0 |  |  | **Add `sv_SE.territory`, the Swedish name of each `misc.territory` row, keyed by its alpha-2 code.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 2 | 3.8 |
| 242 | 0.3.0 |  |  | **Add `sv_SE.bloodtype`, weighted by Sweden's distribution.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 2.1 by 2 | 3.8 |
| 246 | 0.3.0 |  |  | **Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 2.1 by 2 | 3.8 |
| 258 | 0.3.0 |  |  | **Add `misc.product` names, adjectives, materials and departments, and `misc.garmentsize`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 3 | 3.8 |
| 266 | 0.3.0 |  |  | **Add `misc.password`, `hash`, `jwt`, `slug`, `hostname`, `avatar` and `placeholderimage`.** | 1 | 1 | 1 | 1 | 3 | 1 by 1, 2 by 3 | 3.8 |
| 255 | 0.3.0 |  |  | **Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.** | 2 | 1 | 2 | 3 | 3 | 1 by 1, 4 by 3, 2 by 2 | 3.8 |
| 254 | 0.3.0 |  |  | **Say in the docs which of three kinds each category's values are: real, valid or invented.** | 2 | 2 | 1 | 3 | 4 | 1 by 2, 4 by 8 | 3.6 |
| 260 | 0.3.0 |  |  | **Add `misc.element`, `misc.unit` and `misc.planet`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 4.2 by 2 | 3.6 |
| 261 | 0.3.0 |  |  | **Add `misc.animal` by class, and `misc.plant`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 4.2 by 2 | 3.6 |
| 38 | 0.3.0 |  | defect | **Compare `calcParser.binary`'s operator as a rune.** | 1 | 1 | 2 | 2 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 44 | 0.3.0 |  | decision | **Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 45 | 0.3.0 |  |  | **Name the node a selector follows in `childNamed`'s refusal.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 49 | 0.3.0 |  |  | **Report a CRLF rows file holding only its header as having no rows.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 110 | 0.3.0 |  |  | **Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 225 | 0.3.0 |  |  | **Drop "a field cannot be named weight" from the refusal of a `weight` that is no number, and name the column a table's `weight` reads.** | 1 | 1 | 1 | 1 | 1 | 1 by 1, 6 by 2 | 3.5 |
| 269 | 0.3.0 |  |  | **Add `misc.direction`, `misc.continent` and `misc.ulid`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 2 by 1 | 3.4 |
| 263 | 0.3.0 |  |  | **Add `misc.color` names with their hex, rgb, hsl and cmyk forms.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2, 4.2 by 1 | 3.4 |
| 198 | 0.3.0 |  |  | **Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.** | 2 | 2 | 2 | 4 | 3 | 1 by 1, 2.1 by 2, 4.3 by 3 | 3.4 |
| 274 | 0.3.0 | 210 |  | **Add `fr_FR.company`, `fr_FR.email` and `fr_FR.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 278 | 0.3.0 | 206 |  | **Add `zh_CN.company`, `zh_CN.email` and `zh_CN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 283 | 0.3.0 | 208 |  | **Add `es_MX.company`, `es_MX.email` and `es_MX.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 288 | 0.3.0 | 212 |  | **Add `pt_BR.company`, `pt_BR.email` and `pt_BR.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 298 | 0.3.0 | 203 |  | **Add `da_DK.company`, `da_DK.email` and `da_DK.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 303 | 0.3.0 | 204 |  | **Add `fi_FI.company`, `fi_FI.email` and `fi_FI.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 307 | 0.3.0 | 207 |  | **Add `hi_IN.company`, `hi_IN.email` and `hi_IN.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 311 | 0.3.0 | 209 |  | **Add `ar_EG.company`, `ar_EG.email` and `ar_EG.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 315 | 0.3.0 | 211 |  | **Add `bn_BD.company`, `bn_BD.email` and `bn_BD.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 319 | 0.3.0 | 213 |  | **Add `id_ID.company`, `id_ID.email` and `id_ID.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 323 | 0.3.0 | 214 |  | **Add `ur_PK.company`, `ur_PK.email` and `ur_PK.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 328 | 0.3.0 | 205 |  | **Add `is_IS.company`, `is_IS.email` and `is_IS.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 333 | 0.3.0 | 217 |  | **Add `de_DE.company`, `de_DE.email` and `de_DE.date`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 3 | 3.4 |
| 20 | 0.3.0 |  | defect | **Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 4 by 3 | 3.3 |
| 113 | 0.3.0 |  | defect | **Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.** | 2 | 1 | 2 | 3 | 4 | 1 by 2, 4 by 6 | 3.3 |
| 117 | 0.3.0 |  |  | **List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 4 | 3.3 |
| 237 | 0.3.0 |  |  | **Add `sv_SE.car`, a make and model weighted by Mobility Sweden's 2025 registrations.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 2, 4.2 by 2 | 3.2 |
| 248 | 0.3.0 |  |  | **Add `en_US.car`, a make and model from NHTSA vPIC.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2.1 by 2, 4.2 by 2 | 3.2 |
| 32 | 0.3.0 |  |  | **Add `misc.browser` as the parent of `misc.useragent`, so `misc.browser[Chrome].useragent` resolves.** | 2 | 1 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 3.2 |
| 264 | 0.3.0 |  |  | **Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 3.2 |
| 268 | 0.3.0 |  |  | **Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.** | 1 | 2 | 1 | 2 | 3 | 1 by 1, 2 by 2 | 3.2 |
| 164 | 0.3.0 |  |  | **Let a table's format or a cell bind a name, which lives in that format or cell and keeps one pick through the draw.** | 2 | 1 | 2 | 3 | 2 | 1 by 1, 5 by 3 | 3.2 |
| 50 | 0.3.0 |  |  | **Reword `prove`'s refusal of a typed column reading a row, `{/region}`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 52 | 0.3.0 |  |  | **Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 92 | 0.3.0 |  |  | **Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 219 | 0.3.0 |  |  | **Name the path and the Go type where a struct field's type cannot hold what its tag renders: `sv_SE.person renders text, which int cannot hold`.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 221 | 0.3.0 |  |  | **Refuse two references under a record `--format` by saying a run renders one record: today `{/users}{/orders}` reports that an inline record has no fields.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 224 | 0.3.0 |  |  | **Name the module in every error about its manifest or its files: today an `fs.FS` module's errors name only the file inside it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 228 | 0.3.0 |  |  | **Offer `{/path as p}` where a token starts with a name nothing binds and no loaded path starts with, in a record's format too: today `{p.first}` with `p` unbound suggests `{/p.first}`.** | 2 | 1 | 1 | 2 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 229 | 0.3.0 |  |  | **Name the line and column where JSON parsing stopped, when a template that opens like a JSON object then fails as a format string.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 3 | 3.1 |
| 21 | 0.3.0 |  |  | **List `title` in both locales, and `sv_SE`'s `birth-number`, in the README's Data list of what each locale carries.** | 1 | 2 | 1 | 2 | 1 | 1 by 1 | 3.1 |
| 13 | 0.3.0 |  | defect | **Place xlsx cells by their `r` reference in `data-import/xlsx.py`.** | 1 | 2 | 2 | 3 | 2 | 1 by 1, 4.2 by 2 | 3.0 |
| 181 | 0.3.0 |  | decision | **Fail a load where two sources define one category, unless the library's option or the CLI's flag says the later replaces the earlier, and then make every read reach the replacement.** | 4 | 1 | 3 | 6 | 4 | 1 by 2, 6 by 5 | 3.0 |
| 165 | 0.3.0 |  |  | **Let a table carry `name` with neither `key` nor `parent`, and list the rows a name matches by their line where it matches several.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 3.0 |
| 166 | 0.3.0 |  |  | **Let a key or name cell hold a bracket, a brace, a double quote or a pipe, so the table loads, its row draws at random, and only a selector cannot reach it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 2 | 3.0 |
| 226 | 0.3.0 |  |  | **Show in the README how to make a column unique, such as an email's local part carrying `{seq()}`.** | 1 | 2 | 1 | 2 | 2 | 1 by 1, 5 by 2 | 3.0 |
| 39 | 0.3.0 |  |  | **Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 2 | 2.9 |
| 42 | 0.3.0 |  |  | **Refuse a struct tag's `datatype` other than its Go type's before proving its values, as the README's Library section promises.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 2 | 2.9 |
| 46 | 0.3.0 |  |  | **Report the same error every load for a table with two bad options.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 2 | 2.9 |
| 51 | 0.3.0 |  |  | **Spell a table one way across the errors that name it.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 2 | 2.9 |
| 115 | 0.3.0 |  |  | **Spell a rows file's line one way in every error, naming the file.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 6 by 2 | 2.9 |
| 259 | 0.3.0 |  |  | **Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 2 by 4, 4.2 by 3 | 2.8 |
| 161 | 0.3.0 |  |  | **Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.** | 1 | 1 | 1 | 1 | 2 | 1 by 1, 5 by 1 | 2.8 |
| 55 | 0.3.0 |  |  | **Document `ErrNoColumns` and `MaxRepeat` in the README's Library section.** | 1 | 2 | 1 | 2 | 2 | 1 by 1 | 2.6 |
| 28 | 0.3.0 |  | decision | **Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 2.1 by 3, 4 by 2 | 2.5 |
| 247 | 0.3.0 |  |  | **Add `en_US.licenseplate`, drawn per state from that state's pattern.** | 2 | 2 | 2 | 4 | 4 | 1 by 1, 2.1 by 3, 4 by 2 | 2.5 |
| 270 | 0.3.0 | 210 |  | **Add `geo/FR` with France's places and postal codes, and `fr_FR.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 275 | 0.3.0 | 206 |  | **Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 279 | 0.3.0 | 208 |  | **Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 284 | 0.3.0 | 212 |  | **Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 294 | 0.3.0 | 203 |  | **Add `geo/DK` with Denmark's places and postal codes, and `da_DK.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 299 | 0.3.0 | 204 |  | **Add `geo/FI` with Finland's places and postal codes, and `fi_FI.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 304 | 0.3.0 | 207 |  | **Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 308 | 0.3.0 | 209 |  | **Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 312 | 0.3.0 | 211 |  | **Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 316 | 0.3.0 | 213 |  | **Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 320 | 0.3.0 | 214 |  | **Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 324 | 0.3.0 | 205 |  | **Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 329 | 0.3.0 | 217 |  | **Add `geo/DE` with Germany's places and postal codes, and `de_DE.address` printing them.** | 3 | 2 | 2 | 5 | 5 | 1 by 2, 2.1 by 5 | 2.5 |
| 267 | 0.3.0 |  |  | **Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.** | 1 | 2 | 1 | 2 | 4 | 1 by 1, 2 by 2, 4.2 by 2 | 2.4 |
| 238 | 0.3.0 |  |  | **Draw `sv_SE.word` and `sv_SE.sentence` from SALDO's nouns, verbs, adjectives and adverbs, each part of speech a path of its own.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 2.1 by 3, 4.2 by 3 | 2.3 |
| 249 | 0.3.0 |  |  | **Draw `en_US.word` and `en_US.sentence` from Moby or WordNet, each part of speech a path of its own, and add `en_US.paragraph`.** | 2 | 2 | 1 | 3 | 5 | 1 by 2, 2.1 by 3, 4.2 by 3 | 2.3 |
| 19 | 0.3.0 |  |  | **Fill the 398 Swedish localities weighted 200 from SCB småorter.** | 2 | 2 | 2 | 4 | 4 | 1 by 1, 2.1 by 2, 4.2 by 3 | 2.3 |
| 127 | 0.3.0 |  |  | **Give a Swedish street only postal codes whose delivery digit means street or rural delivery, dropping codes whose delivery digit is 9, or record why they stay.** | 2 | 2 | 2 | 4 | 3 | 1 by 1, 4.3 by 2 | 2.2 |
| 23 | 0.3.0 |  |  | **Give every other shipped category that could reach something real a path that never does.** | 2 | 2 | 2 | 4 | 5 | 1 by 2, 4.3 by 6 | 1.9 |
| 81 | 0.3.0 |  | decision | **Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.** | 3 | 1 | 3 | 5 | 4 | 1 by 1, 6 by 4, 4 by 2 | 1.9 |
| 196 | 0.3.0 |  | decision | **Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 6 by 6 | 1.8 |
| 223 | 0.3.0 |  |  | **Let a column draw a key from another record's rows, so an order's `user_id` names a user the same run wrote.** | 3 | 1 | 2 | 4 | 6 | 1 by 4, 5 by 3 | 1.8 |
| 91 | 0.3.0 |  |  | **Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 5 by 4 | 1.7 |
| 271 | 0.3.0 | 270 |  | **Give `geo/FR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 280 | 0.3.0 | 279 |  | **Give `geo/MX`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 285 | 0.3.0 | 284 |  | **Give `geo/BR`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 295 | 0.3.0 | 294 |  | **Give `geo/DK`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 300 | 0.3.0 | 299 |  | **Give `geo/FI`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 325 | 0.3.0 | 324 |  | **Give `geo/IS`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 330 | 0.3.0 | 329 |  | **Give `geo/DE`'s places their streets.** | 2 | 2 | 2 | 4 | 5 | 1 by 1, 2.1 by 4, 4.3 by 2 | 1.7 |
| 53 | 0.3.0 |  |  | **Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.** | 3 | 1 | 2 | 4 | 5 | 1 by 2, 5 by 2 | 1.7 |
| 170 | 0.3.0 |  |  | **Decide in one function whether a read lands in a name's pick, which `refuseTwiceDrawnIn` calls at load and `keptInPick` at render.** | 3 | 1 | 3 | 5 | 3 | 3.2 by 6 | 1.6 |
| 178 | 0.3.0 |  |  | **Plan with the systems architect which `internal/` packages the engine splits into, by area, and how each type stops being written by more than one stage, then file the moves as items.** | 1 | 2 | 1 | 2 | 3 | 3.2 by 6 | 1.6 |
| 194 | 0.3.0 |  | decision | **Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 6 by 4 | 1.6 |
| 18 | 0.3.0 |  |  | **Give the address records one column set across countries: `region` and `municipality` as columns on `geo.SE.address` too.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 5 by 3 | 1.6 |
| 29 | 0.3.0 |  |  | **Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.** | 3 | 1 | 2 | 4 | 4 | 1 by 1, 5 by 3 | 1.6 |
| 56 | 0.3.0 |  | defect | **Build on a manual run of `test.yml` on `main`.** | 1 | 2 | 1 | 2 | 1 | 1.1 by 1 | 1.5 |
| 107 | 0.3.0 | 178 | decision | **Move the template engine into the `internal/` packages item 178 plans, untangled so no type is written by more than one stage, leaving the root the public API and its options.** | 4 | 1 | 3 | 6 | 9 | 3.2 by 9, 3.1 by 4 | 0.1 |
| 177 | 0.3.0 | 107 | decision | **Run the nine-seat comprehension panel and the scoring run again on the moved engine, and file what the panel names.** | 1 | 2 | 1 | 2 | 3 | 3 by 4 | 1.5 |
| 36 | 0.3.0 |  |  | **Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.** | 3 | 1 | 2 | 4 | 6 | 1 by 3, 6 by 7 | 1.5 |
| 171 | 0.3.0 |  |  | **Handle a template that only reads one record column in one place, so `checkColumns`, `itemDatatype`, `proveColumnItem` and `renderLeaf` no longer each special-case `readsColumn`.** | 3 | 1 | 3 | 5 | 3 | 3.2 by 5 | 1.4 |
| 197 | 0.3.0 |  | decision | **Replace the namespace mount with one the loader states in one place, and give `datafiles.Source.at` one job.** | 3 | 1 | 2 | 4 | 3 | 3.2 by 5 | 1.4 |
| 201 | 0.3.0 |  |  | **Place each Swedish street in the postort whose area holds it, so a big postort such as Stockholm carries its own streets.** | 3 | 2 | 2 | 5 | 6 | 1 by 2, 2.1 by 2, 4.3 by 3 | 1.2 |
| 167 | 0.3.0 |  |  | **Lower the reading cost of the files a chunk reads most before its first commit.** | 1 | 2 | 1 | 2 | 3 | 3 by 3 | 1.2 |
| 175 | 0.3.0 |  |  | **Rewrite the comments the 2026-10-09 panel needed several reads of, and delete those it read as restating their code.** | 1 | 2 | 1 | 2 | 2 | 3.4 by 4 | 1.1 |
| 26 | 0.3.0 |  | defect | **Spell `sv_SE.personnummer` with the `+` Skatteverket uses from the year the holder turns 100, judged by a date the caller gives.** | 4 | 2 | 3 | 7 | 6 | 1 by 2, 4 by 6 | 1.1 |
| 172 | 0.3.0 |  |  | **Pass `drawSteps` the kind of read it draws as an argument of its own, so no caller signals a sibling path by a nil `pins` or a fresh read by a nil `memo`.** | 2 | 1 | 2 | 3 | 2 | 3.2 by 3 | 1.1 |
| 216 | 0.3.0 |  |  | **Share between countries what `geo-se.py` and `data_test.go` hard-code for SE, in the first locale chunk that needs each.** | 2 | 2 | 1 | 3 | 2 | 3.2 by 3 | 1.1 |
| 251 | 0.3.0 |  |  | **Export from the library whether a reference body is a lone path, so the CLI's `isPath` stops copying the path grammar.** | 2 | 1 | 2 | 3 | 2 | 3.2 by 3 | 1.1 |
| 108 | 0.3.0 |  | defect | **Name item 26 in the decision "No builtin reads the clock, so a date is bounded by days, never by an age", as the item that ends its conflict with goal 4.** | 1 | 2 | 1 | 2 | 1 | 3.1 by 2 | 1.0 |
| 173 | 0.3.0 |  |  | **Call what a token reads its spelling, never a name, in `format.go`, `arm.go` and `reference.go`, so a name in the engine is only a bound name.** | 1 | 1 | 1 | 1 | 3 | 3.3 by 4 | 1.0 |
| 75 | 0.3.0 |  |  | **Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.** | 2 | 1 | 1 | 2 | 5 | 1 by 1, 5 by 3 | 0.9 |
| 33 | 0.3.0 |  |  | **Group `data/misc` into folders where a group name makes a path easier to guess.** | 3 | 1 | 3 | 5 | 5 | 1 by 1, 5 by 2 | 0.9 |
| 174 | 0.3.0 |  |  | **Spell once the fact that a walk's next step leaves a selected row, which `pathPos.fromRow` and `tableRoute.nextLinked` each spell today.** | 1 | 1 | 1 | 1 | 1 | 3.2 by 2 | 0.9 |
| 193 | 0.3.0 |  | decision | **Move the largest shipped packages into nested Go modules that keep their import paths, once `TestShippedDataStaysUnder100MiB` fails.** | 4 | 2 | 3 | 7 | 6 | 1 by 1, 2 by 3 | 0.9 |
| 195 | 0.3.0 |  | principle | **Parse each builtin argument once, in `checkArgs`, and hand `prep` and `proveNumber` the parsed value.** | 2 | 1 | 2 | 3 | 3 | 3.2 by 3 | 0.8 |
| 17 | 0.3.0 |  |  | **Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.** | 3 | 1 | 3 | 5 | 6 | 1 by 1, 5 by 3, 2 by 2 | 0.8 |
| 72 | 0.3.0 |  | principle | **Fail the merge gate when test coverage falls below the last recorded figure.** | 2 | 2 | 1 | 3 | 3 | 3 by 2 | 0.8 |
| 11 | 0.3.0 |  |  | **Test that every `geo` tree holds the five table names and the `address` columns that port across countries.** | 1 | 1 | 1 | 1 | 2 | 3.2 by 2 | 0.7 |
| 40 | 0.3.0 |  |  | **Parse `defaultTable`'s path with the library's grammar.** | 1 | 1 | 2 | 2 | 2 | 3.2 by 2 | 0.7 |
| 86 | 0.3.0 |  |  | **Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.** | 1 | 1 | 1 | 1 | 2 | 3.2 by 2 | 0.7 |
| 95 | 0.3.0 |  | principle | **Hold a struct's columns and the field each fills in one slice.** | 1 | 1 | 2 | 2 | 2 | 3.2 by 2 | 0.7 |
| 112 | 0.3.0 | 36 |  | **Check a calc's field operands where its template resolves, beside its name operands.** | 2 | 1 | 1 | 2 | 2 | 3.2 by 2 | 0.7 |
| 116 | 0.3.0 |  |  | **Turn a row index into its rows file's line in one function.** | 1 | 1 | 1 | 1 | 2 | 3.2 by 2 | 0.7 |
| 186 | 0.3.0 | 26 | decision | **Load a module carrying functions through an option of its own, its functions reaching the generator's randomness and date through an interface the root declares.** | 4 | 1 | 3 | 6 | 7 | 1 by 3 | 0.7 |
| 168 | 0.3.0 |  | principle | **Test the shared steps of `data-import/` in the merge gate, starting with `geo.fence`.** | 2 | 1 | 1 | 2 | 4 | 3 by 2 | 0.5 |
| 218 | 0.3.0 |  | decision | **Give every country shipped without streets its streets, once the maintainer accepts ODbL data or the country opens a register.** | 3 | 2 | 2 | 5 | 7 | 1 by 1, 2.1 by 4 | 0.5 |
| 129 | 0.3.0 |  | decision | **Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.** | 4 | 1 | 3 | 6 | 8 | 1 by 3, 5 by 7 | 0.4 |
| 30 | 0.3.0 |  |  | **Draw `email.local` from `username`, so the two share one handle list.** | 1 | 1 | 1 | 1 | 2 | 3.2 by 1 | 0.4 |
| 41 | 0.3.0 |  | decision | **Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.** | 4 | 1 | 3 | 6 | 7 | 1 by 1, 5 by 4 | 0.3 |
| 109 | 0.3.0 |  | principle | **Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.** | 1 | 2 | 1 | 2 | 4 | 3 by 1 | 0.3 |
| 14 | 0.3.0 |  |  | **Spell `misc.creditcard`'s digit runs `{digits(n)}`.** | 1 | 1 | 1 | 1 | 2 | 5 by 1 | 0.2 |
| 59 | 0.4.0 |  |  | **Promise in `Generator`'s godoc that its renders run one at a time.** | 1 | 2 | 1 | 2 | 1 | 1 by 1 | 3.1 |
| 62 | 0.4.0 |  |  | **Add `{btc()}` and `{eth()}`.** | 2 | 1 | 2 | 3 | 4 | 1 by 1, 2 by 2, 4 by 2 | 2.4 |
| 65 | 0.4.0 |  |  | **Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.** | 2 | 2 | 1 | 3 | 3 | 1 by 1 | 1.9 |
| 68 | 0.4.0 |  | decision | **Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.** | 1 | 3 | 2 | 4 | 1 | 1.1 by 1 | 1.5 |
| 63 | 0.4.0 |  |  | **Pair a street with its exact postnummer.** | 3 | 2 | 3 | 6 | 6 | 1 by 2, 4 by 6 | 1.1 |
| 66 | 0.4.0 |  |  | **Cut the README's Layout block to the lines that say what a file name cannot.** | 1 | 2 | 1 | 2 | 1 | 3.1 by 2 | 1.0 |
| 67 | 0.4.0 |  |  | **Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.** | 1 | 2 | 1 | 2 | 2 | 3.4 by 2 | 0.6 |
| 64 | 0.4.0 |  |  | **Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.** | 1 | 2 | 1 | 2 | 4 | 3.4 by 2 | 0.3 |
| 334 | 0.5.0 | 202, 289 |  | **Add `nn_NO`, written in Nynorsk, reading `nb_NO`'s names and id and `geo/NO`, with month and weekday names from CLDR and titles of its own.** | 2 | 2 | 1 | 3 | 2 | 1 by 1, 2.1 by 2 | 3.8 |
| 69 | 1.0.0 |  |  | **Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.** | 1 | 2 | 1 | 2 | 3 | 1 by 8 | 15.4 |
| 70 | 1.0.0 |  |  | **Publish a homepage with an in-browser generator, the library compiled to WebAssembly.** | 3 | 2 | 2 | 5 | 6 | 1 by 6 | 2.6 |
| 190 | 1.1.0 |  | question | **Decide whether fejkdata pseudonymises real datasets, and if so add the goal, the audience persona and the items below to the README and the plan.** | 4 | 2 | 3 | 7 | 9 | 1 by 4, 4 by 3 | 0.3 |

## Details

### 227. Add a README quick start per task, for the bulk fixture author and the Go test author, each rendering one person whose name, sex, id, email and address agree.

From the 2026-10-11 audience bench: the bulk fixture author assembled a users record from six sections, Records, Datatype, Functions, the person example, Names and Transforms, and the Go test author took three runs to a struct whose personnummer, name and email agree. The record both reached: `{/sv_SE.sex as s}{/sv_SE.last-name as l}{/sv_SE.address as a}` with `{s.first-name}`, `{date(1940-01-01,2006-12-31,'060102')}-{s.birth-number.number}{luhn()}` and `{lowercase(ascii(s.first-name))}.{lowercase(ascii(l))}`. Bench the quick starts with both personas before merging.

### 230. Add `sv_SE.organisationsnummer`, whose first digit is its legal form's group digit and whose last is a Luhn check digit, and `sv_SE.vat`, `SE` + the number + `01`.

Bolagsverket's group digits and the Luhn rule are facts. [Source research](docs/research/research-sources-se.md).

### 157. Show in the README's CLI section how PowerShell and cmd.exe pipe a template, one line per shell family.

Split from item 129. cmd.exe's `echo` keeps its quotes and a trailing space, so `echo {/users} | fejkdata` sends `{/users} ` and stops naming the record.

### 231. Draw `sv_SE.company`'s legal form weighted by Bolagsverket's 2025 registrations, and its names from patterns derived from Bolagsverket's company register.

Bolagsverket's registration statistics are CC BY 2.5 SE; derive name-token frequencies from its bulk file offline and embed no row of it. [Source research](docs/research/research-sources-se.md).

### 234. Draw `sv_SE.phone` from every area code in PTS's numbering plan, grouped as Språkrådet writes it, and add its international form, `+46 70 123 45 67`.

PTS's 2024 numbering plan is facts, no licence stated; cite PTS. Today the fixed numbers draw from 20 area codes. [Source research](docs/research/research-sources-se.md).

### 202. Add `nb_NO` with its given names by sex, last names and titles, and `nb_NO.person`.

`nb` is Bokmål, the written standard most Norwegians use.

### 292. Add `nb_NO`'s national personal id, the fødselsnummer, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 291. Add `nb_NO.phone`, drawn from Norway's numbering plan.

Find the numbering plan and its licence first.

### 335. Merge every Renovate update to a build, test or release tool once the gate passes, and release nothing for it.

On 2026-10-18 Python 3.15.0 (#219), Fedora 46 (#221) and `golang.org/x/vuln` v1.9.0 (#226) waited four to five days for a hand merge, since `renovate.json` merges only a Go patch itself. None of them reaches the binaries or the image, which is built `FROM scratch`, so no release follows. An update that needs a code or test change still waits for a human.

### 289. Add `geo/NO` with Norway's places and postal codes, and `nb_NO.address` printing them.

Bring's postal register spells every poststed in capitals, so the import recases each poststed before `well_cased` sees it.

### 336. Release a patch from CI for every update to what a release carries, once the gate passes with no other change: Go, Go's minor included, any module `go.mod` requires, and GoReleaser.

The maintainer, 2026-10-18: no patch should wait on a human. A dependency bump that passes the tests unchanged changes no row of the README's Versioning table, so it is a patch; a new Go minor leaves `go.mod`'s `go 1.22`, the lowest supported Go, where it is. `release-tooling/patch_release.py cut` already cuts a Go patch this way from the last release's tag; widen it to the rest, and revise the decision "CI releases each Go patch at once, as a fejkdata patch cut from the last release's tag", which leaves a Go minor to the maintainer. Name in the changelog section each platform a new Go minor stops supporting, such as an older macOS, since a binary user loses it in a patch. Make the comprehension seats' places in `patch_release.py`, `packages.yml` and `goreleaser.yaml` easier to read where the change touches them.

### 290. Give `geo/NO`'s places their streets.

Kartverket's address register (CC BY 4.0) holds every street with its postnummer.

### 169. Reword the decision "A parent row with no child row is a load error" to say what each geo import does with such a row: drop it, or exit where `geo-se.py` would lose a kommun or a län.

Its entry says "The import script drops or fills such rows". There are two geo imports, and `geo-se.py` exits when `geo.fence` drops a kommun or a län, while `geo-us.py` drops counties and states. Rewording a decision is the maintainer's call.

### 252. Publish `fejkdata-bin` to the AUR from the Packages workflow, so Arch installs with `yay -S fejkdata-bin`.

The AUR closed new account registration on 2026-10-14, so 0.1.0's packaging shipped without it. Once an account exists, restore what that chunk dropped: the `aurs` block in `release-tooling/goreleaser.yaml` (`git show 285ed85:release-tooling/goreleaser.yaml`), `aur` in both `--skip` lists, the `AUR_KEY` secret in both workflows, and the README's Install line and secret bullet.

### 210. Add `fr_FR` with its given names by sex, last names and titles, and `fr_FR.person`.

France holds more French speakers than DR Congo: 64M to 49M on Wikipedia's "Geographical distribution of French speakers", which credits a 2022 OIF report but flags its table as not taken from that report.

### 273. Add `fr_FR`'s national personal id, the NIR, the numéro de sécurité sociale, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 206. Add `zh_CN` with its given names by sex, last names and titles, and `zh_CN.person`.

Written in simplified characters.

### 277. Add `zh_CN`'s national personal id, the 18-digit resident identity card number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 208. Add `es_MX` with its given names by sex, last names and titles, and `es_MX.person`.

Mexico holds the most Spanish speakers.

### 282. Add `es_MX`'s national personal id, the CURP, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 212. Add `pt_BR` with its given names by sex, last names and titles, and `pt_BR.person`.

Brazil holds the most Portuguese speakers.

### 287. Add `pt_BR`'s national personal id, the CPF, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 297. Add `da_DK`'s national personal id, the CPR number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 302. Add `fi_FI`'s national personal id, the henkilötunnus, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 207. Add `hi_IN` with its given names by sex, last names and titles, and `hi_IN.person`.

Hindi is written in Devanagari.

### 306. Add `hi_IN`'s national personal id, the Aadhaar number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 209. Add `ar_EG` with its given names by sex, last names and titles, and `ar_EG.person`.

Standard Arabic has no native speakers. Confirm from Ethnologue's country pages that Egypt holds the most second-language speakers, and record the answer in the decision.

### 310. Add `ar_EG`'s national personal id, the national ID number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 211. Add `bn_BD` with its given names by sex, last names and titles, and `bn_BD.person`.

Bangladesh holds the most Bengali speakers.

### 314. Add `bn_BD`'s national personal id, the NID number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 213. Add `id_ID` with its given names by sex, last names and titles, and `id_ID.person`.

Indonesian is spoken almost only in Indonesia.

### 318. Add `id_ID`'s national personal id, the NIK, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 214. Add `ur_PK` with its given names by sex, last names and titles, and `ur_PK.person`.

Urdu is Pakistan's national language, yet India has more native speakers: 50.8M by its 2011 census against Pakistan's 22.3M by its 2023 one. Confirm from Ethnologue's country pages that Pakistan holds the most speakers in all, else the locale is ur_IN, and record the answer in the decision.

### 322. Add `ur_PK`'s national personal id, the CNIC number, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 327. Add `is_IS`'s national personal id, the kennitala, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 332. Add `de_DE`'s national personal id, the Steuer-Identifikationsnummer, with its check digits.

Find the issuer's rules, and a test series if it publishes one, first.

### 34. Add `{isin()}`, `{cusip()}`, `{aba()}` and `{vin()}`.

`isin`: Luhn over letters expanded to digits. `aba`: 3-7-1 weights. `vin`: position 9 over the whole; a sample taking the WMI, since the check sits mid-string.

### 243. Add `en_US.ein`, drawing only the prefixes the IRS assigns.

The IRS's prefix list is facts. [Source research](docs/research/research-sources-world.md).

### 158. Drop a UTF-8 byte order mark that starts stdin, and say in the README how to stop PowerShell 5.1 sending `Göteborg` as `G?teborg`.

A file a Windows editor saves with a BOM prints an invisible U+FEFF first, and a JSON template with one is read as a format string. PowerShell 5.1 encodes what it pipes to a program by `$OutputEncoding`, ASCII by default.

### 163. Lift the CLI's `--repeat` cap, since the CLI streams each render, and give the builtins' count and decimal caps a technical reason in the README.

The README's Audience names a bulk fixture author writing millions of rows, and `-n 5000000` is misuse today, so they loop. The decision "The repeat cap bounds renders, not bytes" covers nested repeats in data, not `-n`. The README's Functions section justifies the builtin caps by a fat-fingered `hex(2000000000)`, a guess at a mistake under goal 5.7; the reason that holds is that one render may not allocate gigabytes. The goals audit of the narrowed goal 5.7 found it.

### 241. Give `sv_SE.person` a `birthdate`.

The decision "No builtin reads the clock, so a date is bounded by days, never by an age" rules out an age; draw the date between two fixed dates.

### 155. Add `misc.isin` and `misc.cusip`, built by their structure rules with their check digits drawn by `{isin()}` and `{cusip()}`.

`{luhn()}` skips letters, so it draws a wrong ISIN check digit, and CUSIP's check digit is not Luhn.

### 236. Add `sv_SE.licenseplate`, `ABC 123` and `ABC 12A`, never drawing a combination Transportstyrelsen blocks.

Transportstyrelsen's 99 blocked combinations are facts, with no licence stated; the `ABC 123` and `ABC 12A` formats come from sv.wikipedia, unconfirmed by Transportstyrelsen. [Source research](docs/research/research-sources-se.md).

### 22. Give `url` and `email` a path that draws only domains nobody can register, keeping the wide set as the default.

17 of the 40 distinct ones shipped today sit on `.se`, `.nu`, `.io` and `.co`, which anyone may register, and only RFC 2606's `example.com`, `.net`, `.org`, `.test`, `.example`, `.invalid` and `.localhost` provably reach nothing.

### 235. Add `sv_SE.bankgiro` and `sv_SE.plusgiro` with their Luhn check digit, and `sv_SE.bankaccount` with a clearing number and the check digit its bank's account type demands.

Bankinfrastruktur's clearing table and Bankgirot's rules are facts; cite Bankinfrastruktur i Sverige AB. [Source research](docs/research/research-sources-se.md).

### 256. Add `misc.bic`, built by its structure rules.

A template; the structure rules are facts.

### 245. Draw `en_US.phone` from every area code NANPA has in service.

Today it draws 21 area codes; the exchange already ships as `en_US.phone-exchange`. No research file covers NANPA yet: find its list of area codes in service and that list's licence first.

### 272. Add `fr_FR.phone`, drawn from France's numbering plan.

Find the numbering plan and its licence first.

### 276. Add `zh_CN.phone`, drawn from China's numbering plan.

Find the numbering plan and its licence first.

### 281. Add `es_MX.phone`, drawn from Mexico's numbering plan.

Find the numbering plan and its licence first.

### 286. Add `pt_BR.phone`, drawn from Brazil's numbering plan.

Find the numbering plan and its licence first.

### 296. Add `da_DK.phone`, drawn from Denmark's numbering plan.

Find the numbering plan and its licence first.

### 301. Add `fi_FI.phone`, drawn from Finland's numbering plan.

Find the numbering plan and its licence first.

### 305. Add `hi_IN.phone`, drawn from India's numbering plan.

Find the numbering plan and its licence first.

### 309. Add `ar_EG.phone`, drawn from Egypt's numbering plan.

Find the numbering plan and its licence first.

### 313. Add `bn_BD.phone`, drawn from Bangladesh's numbering plan.

Find the numbering plan and its licence first.

### 317. Add `id_ID.phone`, drawn from Indonesia's numbering plan.

Find the numbering plan and its licence first.

### 321. Add `ur_PK.phone`, drawn from Pakistan's numbering plan.

Find the numbering plan and its licence first.

### 326. Add `is_IS.phone`, drawn from Iceland's numbering plan.

Find the numbering plan and its licence first.

### 331. Add `de_DE.phone`, drawn from Germany's numbering plan.

Find the numbering plan and its licence first.

### 262. Add `misc.bloodtype`, weighted by its global distribution.

A choice, from AABB's figures, which are facts. Item 242 adds a Swedish one.

### 15. Add `misc.ip`'s documentation, private and CIDR forms, `misc.ipv6` in `2001:db8::/32`, and a locally administered `misc.mac`.

Templates, from RFC 5737, RFC 1918 and RFC 9637, which are facts.

### 265. Add `misc.commit` messages, branch names and `sha`s.

Templates, from lists an LLM writes and `{hex(40)}`.

### 257. Add `misc.ean`, `upc`, `gtin`, `isbn`, `issn`, `imei`, `asin` and `sku`, with their check digits, under GS1's demo prefix 952 where one applies.

Templates; the check-digit rules are facts.

### 239. Add `sv_SE.month` and `sv_SE.weekday` names from CLDR, and `sv_SE.holiday` from lag 1989:253.

CLDR is under the Unicode licence; the law is facts. [Source research](docs/research/research-sources-se.md).

### 244. Draw `en_US.company`'s suffixes and name patterns from SEC company names, and add `en_US.naics`, a table of NAICS 2022 codes.

NAICS 2022 is public domain; SEC's `company_tickers.json` is free to reuse. [Source research](docs/research/research-sources-world.md).

### 250. Add `en_US.month` and `en_US.weekday` names from CLDR, and `en_US.holiday`, the federal holidays.

CLDR is under the Unicode licence. [Source research](docs/research/research-sources-world.md). The federal holidays are listed in 5 U.S.C. 6103, which no research file covers yet.

### 24. Stop `misc.territory[EH].tld` rendering `.eh`, the one shipped TLD `misc.tld` does not hold.

ISO 3166 reserves it for Western Sahara and the root zone has never been delegated it, so no resolver answers for it.

### 88. Add a 12-digit `YYYYMMDDNNNC` form of `sv_SE.personnummer` and `sv_SE.samordningsnummer`, the form Skatteverket's testpersonnummer series and many systems store.

A system that stores the 12-digit form never meets the `+`, so it needs a fixture in that form.

### 232. Add `sv_SE.sni`, a table of SCB's SNI 2025 industry codes with their Swedish labels.

SNI 2025 replaces SNI 2007 from 2025-12-08. SCB's classification pages state no licence; its terms ask for "Källa: SCB". [Source research](docs/research/research-sources-se.md).

### 233. Add `sv_SE.ssyk`, a table of SCB's SSYK 2012 occupation codes with their Swedish labels.

SCB's SSYK pages state no licence; its terms ask for "Källa: SCB". [Source research](docs/research/research-sources-se.md).

### 31. Rename `misc.territory.country` to `sovereign`.

It holds a code, and `country` collides with the `misc.country` path it replaced.

### 240. Add `sv_SE.territory`, the Swedish name of each `misc.territory` row, keyed by its alpha-2 code.

From CLDR's `sv` territory names, under the Unicode licence. [Source research](docs/research/research-sources-se.md).

### 242. Add `sv_SE.bloodtype`, weighted by Sweden's distribution.

geblod.nu's 2007 figures, via sv.wikipedia: A+ 37, O+ 32, B+ 10, A− 7, O− 6, AB+ 5, B− 2, AB− 1 (%). Item 262 plans a global `misc.bloodtype` from AABB. [Source research](docs/research/research-sources-se.md).

### 246. Add `en_US.bankaccount`, an account number of at most 17 characters, the width of a Nacha ACH entry's account field.

No national register defines the account number, and no research file covers Nacha yet. Item 154 adds the routing number beside it.

### 258. Add `misc.product` names, adjectives, materials and departments, and `misc.garmentsize`.

Tables or choices, from lists an LLM writes.

### 266. Add `misc.password`, `hash`, `jwt`, `slug`, `hostname`, `avatar` and `placeholderimage`.

Templates, by each one's shape. A JWT's parts take `{base64url(n)}`, which item 35 adds.

### 255. Give `misc.creditcard` a row per network, with its IIN, length, CVV and expiry, and its Luhn check digit.

A table and its template, from each network's rules, which are facts.

### 254. Say in the docs which of three kinds each category's values are: real, valid or invented.

Goal 4 names the kinds. A real value is a thing that exists, such as the territory `SE`. A valid value is made up but passes the check the real system applies, such as a Swedish personal number. An invented value is made up entirely, such as a town that does not exist. The maintainer, 2026-10-18: all three have their place, and it is good to tell them apart. Before the three words ship, a panel of readers seated as the README's audience picks the spelling.

### 260. Add `misc.element`, `misc.unit` and `misc.planet`.

Tables, from PubChem, BIPM and NASA, under public domain and CC BY 4.0 terms.

### 261. Add `misc.animal` by class, and `misc.plant`.

Tables, from Wikidata (CC0).

### 38. Compare `calcParser.binary`'s operator as a rune.

`byte(p.rs[p.pos])` reads U+012B `ī` as `+`, so `{calc(a ī b)}` compiles as `a + b`.

### 44. Have `misc.tld[se]` select the row keyed `.se`, or have its miss name `[.se]`.

The decision "`misc.tld` keys carry the leading dot, where other tables key on a bare code" leaves `misc.tld[se]` missing today.

### 45. Name the node a selector follows in `childNamed`'s refusal.

`sv_SE.person[1].first` answers `1 is not a table`.

### 49. Report a CRLF rows file holding only its header as having no rows.

`parseRows` answers "line 2 is empty".

### 110. Refuse an empty segment in a path `Fake` and `FakeRecord` take, as a reference does: `sv_SE.person.` reports `path has an empty segment`.

Today `Fake("sv_SE.person.")` reports `no field ""`, while `{/sv_SE.person.}` reports `path has an empty segment`. `grammar.CheckSegments` is the check a reference, a sibling path and a step up `..` run.

### 269. Add `misc.direction`, `misc.continent` and `misc.ulid`.

Choices and a template.

### 263. Add `misc.color` names with their hex, rgb, hsl and cmyk forms.

A table and its templates, from CSS Color 4 (W3C) and the XKCD colour survey (CC0).

### 198. Fetch the TIGER files of every county a drawn US place spans, so a place's streets and postal codes cover all of it.

`streets()` in `geo-us.py` fetches only each drawn place's main county, the one holding most of its people. A ZCTA in a place's other counties then has no address range and is dropped: New York, under Kings County, ships Brooklyn's codes, one Manhattan code and none on Staten Island, and its streets are mostly Brooklyn's. Whether a neighbouring county is fetched depends on which other places the draw picks. The fix reads every county in `localities`' `county_part`.

### 20. Weight `sv_SE.sex` and `en_US.sex` by bearers, as the README's Data section says they are.

Neither names a `weight`, so a bare `sex` draws evenly.

### 113. Draw `{iban(CC)}` with its country's national check digits, and SE's with a real bank's id, so a validator checking more than mod 97 accepts it.

`iban` in `internal/builtinfunc/checksum.go` draws a random numeric BBAN and sets only the mod-97 digits. BE's last two digits are the first ten mod 97, NO carries a mod-11 account check digit, FI a Luhn digit, ES two control digits, and SE's three-digit bank id must be a real bank's (`docs/research/research-sources-se.md`). Goal 4 asks for the check the real system applies; the README's `{iban(CC)}` row promises only length and mod 97, and changes with it.

### 117. List the fields of what a path has reached when its next segment names one that is not there: `{/word as a}{a.zz}` reports `no field "zz"; word has w`.

Today the refusal names only the missing field, so a data author fixing a typo opens the category to find the spelling. Item 92 covers the hyphenated `calc` operand only.

### 237. Add `sv_SE.car`, a make and model weighted by Mobility Sweden's 2025 registrations.

The decision "`misc` is what every locale shares" puts a national register's cars in its locale. Mobility Sweden asks for attribution and grants no open licence, so ship the top models as facts. [Source research](docs/research/research-sources-se.md).

### 248. Add `en_US.car`, a make and model from NHTSA vPIC.

NHTSA vPIC is public domain. The decision "`misc` is what every locale shares" puts a national register's cars in its locale. [Source research](docs/research/part2-internet-vehicle-airline.md).

### 264. Add `misc.programminglanguage`, `misc.os`, `misc.database`'s columns, types and engines, and `misc.fileext`.

Tables or choices, from GitHub Linguist (MIT) and curated lists. Item 32 adds `misc.browser`.

### 268. Add `misc.lorem`, `hacker`, `hipster`, `catchphrase`, `buzzword` and `quote`.

Tables or choices, from lorem ipsum and lists an LLM writes.

### 164. Let a table's format or a cell bind a name, which lives in that format or cell and keeps one pick through the draw.

The maintainer ruled so on 2026-10-09: a name is a lexical closure, as the decision "A name lives in the category binding it, or in the repeat or choice item binding it, and is drawn on its first read" now states. Today `table.go` refuses `{/sv_SE.person as p}{p.first} {p.last}` in a cell with `a table binds no name`. A cell reads no name its table's format binds. Revise that decision's title and the README's Names list with it.

### 50. Reword `prove`'s refusal of a typed column reading a row, `{/region}`.

Say the format is the table's own (`region's format "{name}"`), name the read to write, `{/region.<column>}` with its columns, and stop calling a one-column format, `{code}`, composed text.

### 52. Name the file a link error comes from beside its category path, `sub.x (/d/a/sub/x.json): …`.

With several `--data-path` layers the author has to work out which directory won, while a parse error on the same file names it.

### 92. Name the hyphenated field in a `calc`'s refusal, and say to rename it: today `{calc(postal-code * 2)}` reports `no field "postal"`.

A hyphen is always subtraction in a calc, and the shipped data has hyphenated fields (`postal-code` and `street-number` in `geo.SE.address`). The refusal names the first fragment, not the field the author meant.

### 228. Offer `{/path as p}` where a token starts with a name nothing binds and no loaded path starts with, in a record's format too: today `{p.first}` with `p` unbound suggests `{/p.first}`.

From the 2026-10-11 audience bench, a Go test author tagging `{p.first}` before binding `p`. The hint is built in `checkArm` before the root is known, and only where the template has no fields; offering the binding there also told `{sv_SE.person.last}` to bind a name `sv_SE`. Decide between the reference and the binding once the root is loaded.

### 13. Place xlsx cells by their `r` reference in `data-import/xlsx.py`.

A sheet omitting an empty cell shifts every later column left.

### 181. Fail a load where two sources define one category, unless the library's option or the CLI's flag says the later replaces the earlier, and then make every read reach the replacement.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Proposed spellings: `WithReplace()` and a boolean `--replace` covering every `-d`, so it goes anywhere on the line. Bench the option, the flag and the clash error with the data and hand fixture authors before merging; the error names the option or the flag. A clash is found from the categories each source's index names and the files each source without a manifest holds, without parsing them. Revises the decision "A `--data-path` override rebinds every reference to the category it replaces", ending its conflict with goal 9.3. The changelog names the flag, the option and the new refusal. The same `embed.FS` passed twice, as `svse.Modules()` beside `data.Modules()` passes `geose.FS`, loads once and is no clash. Until then, a `-d` file replacing `sv_SE/first-name.json` breaks `{/sv_SE.person}` with `token {n.name}: no variant of this 2-way choice carries "name"`, naming no replacement; the clash error ends that.

### 165. Let a table carry `name` with neither `key` nor `parent`, and list the rows a name matches by their line where it matches several.

The maintainer ruled so on 2026-10-09. Today `bindOptions` in `internal/rows` refuses such a table, since `ambiguous`, which reports a name matching several rows, lists those rows by key. Afterwards `ambiguous` lists them by line, the header being line 1, and ends with "make the names unique, or add a key column".

### 166. Let a key or name cell hold a bracket, a brace, a double quote or a pipe, so the table loads, its row draws at random, and only a selector cannot reach it.

The maintainer ruled so on 2026-10-09. Today `checkSelectorCells` in `internal/rows` refuses the whole table. Afterwards a path naming such a key fails with an error naming the character a selector cannot spell, and the README says a row whose name holds such a character is still selected by its key.

### 39. Stop the `columnKinds` suggestion naming a field's own kind or a narrower one.

An `int64` or `float64` field unproven in range is told to become itself, and a `uint64` one to become `int64`.

### 42. Refuse a struct tag's `datatype` other than its Go type's before proving its values, as the README's Library section promises.

`datatype: boolean` over `{int(1,2)}` on an `int` field answers `prints an integer, not a boolean`.

### 46. Report the same error every load for a table with two bad options.

`readTableOptions` in `table.go` returns on the first in Go's map order.

### 51. Spell a table one way across the errors that name it.

In `internal/rows`, `Pins.PinRow` names `t.path` and `Pins.MustRow` panics with `t.segment`, so one table is `territory` and `misc.territory`, and the short spelling names no file where two folders hold that name.

### 115. Spell a rows file's line one way in every error, naming the file.

Today a cell reads `g.t.b, line 3` (`containedNode.labelIn`), a cell check `t.tsv: line 3, b` (`compileCells`) and a link `t.tsv line 3` (`rows.Table.Link`). A reference cycle found from a table's format through a cell names no line at all: `reference cycle: g.t -> format -> b -> b -> /h -> …`, where the column-to-cell step in `renderEdges` repeats the column name.

### 259. Add `misc.airport` with its IATA and ICAO codes, territory and municipality, `misc.airline`, `misc.aircraft`, and `misc.flight`, `seat` and `pnr`.

Tables and templates, from OurAirports (public domain) and Wikidata (CC0).

### 161. Let a field whose name starts or ends with whitespace be bound to a name, or refuse such an identifier, so `{ x as p}` means one thing.

`grammar/identifier.go` allows a field named ` x`, and `{ x}` reads it, but `checkBind` in `format.go` refuses every padded binding, so no spelling binds that field to a name (goal 4.1), and the refusal names one meaning only (goal 5.6). The goals audit of the narrowed goal 5.7 found it.

### 28. Accept a middle name, and draw a `personnummer` whose sex matches a sex selected through a named pick.

Today `personnummer` draws its own sex, so a name and a number agree only through the README Data section's template, which reads both through one bound `sex`.

### 247. Add `en_US.licenseplate`, drawn per state from that state's pattern.

No open list of state patterns exists: write each state's pattern by hand and check it against that state's DMV. Wikipedia's list (CC BY-SA) and openalpr's (AGPL) may be read, never copied. [Source research](docs/research/research-sources-world.md).

### 270. Add `geo/FR` with France's places and postal codes, and `fr_FR.address` printing them.

The Base Adresse Nationale (Licence Ouverte 2.0) names each address's commune and code postal.

### 275. Add `geo/CN` with China's places and postal codes, and `zh_CN.address` printing them.

`rows.Pins` holds eight pins inline, enough for a five-deep geo tree. If CN's tree runs deeper, raise that number to keep a render off the heap. Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 279. Add `geo/MX` with Mexico's places and postal codes, and `es_MX.address` printing them.

SEPOMEX publishes every código postal with its asentamiento.

### 284. Add `geo/BR` with Brazil's places and postal codes, and `pt_BR.address` printing them.

IBGE's CNEFE lists every address with its CEP.

### 294. Add `geo/DK` with Denmark's places and postal codes, and `da_DK.address` printing them.

Read them from the register the streets item names, Danmarks Adresseregister, where it holds them; else find an open source first.

### 299. Add `geo/FI` with Finland's places and postal codes, and `fi_FI.address` printing them.

Posti's free postal code file names each postinumeroalue.

### 304. Add `geo/IN` with India's places and postal codes, and `hi_IN.address` printing them.

India's PIN code directory is open data on data.gov.in. It ships without streets (decision: no open address register covering it whole).

### 308. Add `geo/EG` with Egypt's places and postal codes, and `ar_EG.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 312. Add `geo/BD` with Bangladesh's places and postal codes, and `bn_BD.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 316. Add `geo/ID` with Indonesia's places and postal codes, and `id_ID.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 320. Add `geo/PK` with Pakistan's places and postal codes, and `ur_PK.address` printing them.

Find an open source of its places and postal codes, and that source's licence, first. It ships without streets (decision: no open address register covering it whole).

### 324. Add `geo/IS` with Iceland's places and postal codes, and `is_IS.address` printing them.

Read them from the register the streets item names, Staðfangaskrá, where it holds them; else find an open source first.

### 329. Add `geo/DE` with Germany's places and postal codes, and `de_DE.address` printing them.

GeoNames publishes Germany's postal codes.

### 267. Add `misc.book` titles and authors, `misc.genre`, `misc.instrument`, and `misc.dish` and `misc.ingredient`.

Tables, from Project Gutenberg's catalog, MusicBrainz and USDA, each CC0.

### 238. Draw `sv_SE.word` and `sv_SE.sentence` from SALDO's nouns, verbs, adjectives and adverbs, each part of speech a path of its own.

SALDO's morphology is CC BY 4.0; cite Språkbanken (2017), DOI 10.23695/agcm-ny22. [Source research](docs/research/research-sources-se.md).

### 249. Draw `en_US.word` and `en_US.sentence` from Moby or WordNet, each part of speech a path of its own, and add `en_US.paragraph`.

Moby is public domain; WordNet's licence requires its notice and disclaimer on every copy. [Source research](docs/research/part3-science-text-dates.md).

### 127. Give a Swedish street only postal codes whose delivery digit means street or rural delivery, dropping codes whose delivery digit is 9, or record why they stay.

`docs/research/research-geo-se.md` reads delivery digit 0 or 1 as boxes, 2–4 and 6–7 as street, 5 as rural, 8 as reply mail and 9 as competitions or temporary, with exceptions in big cities. `street_delivery` drops 0, 1 and 8 and keeps 9. The panel of 2026-10-05 found it; whether a shipped street carries a 9 code is unchecked.

### 23. Give every other shipped category that could reach something real a path that never does.

`phone` draws live PTS and NANP ranges — PTS's five fiction series are the Swedish inert set, sourced in [research-sources-se.md](docs/research/research-sources-se.md), and the NANP one still wants a source — and `bankgiro`, `plusgiro` and `routing` draw live prefixes.

### 81. Refuse at load every `calc` operand not proven numeric and every divisor not proven nonzero, in a string column too.

The decision "A constant zero divisor is a load error; in a string column a divisor that is not constant prints `Inf`" lets `{calc(a/(b*c))}` with `b` fixed at `"0"` write `Inf` into every row. An operand that is sometimes not a number prints `NaN`. Both work against goals 6.1 and 4. `valueproof.go` already proves a typed column's divisor nonzero. Once every operand is proven, `neverNumeric` and `constantValue` in `calc.go` go, and with them their reading of `NaN` and `Inf` as numbers, which `proven.Literal` refuses: calc reads text as a number one way.

### 196. Name, in the error for a reference nothing loaded provides, the module the reader's manifest reads by default that would provide it, or say none does.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0", and ends the locale-address decision's conflict with goal 6.2. Carved out of item 182 on 2026-10-11, after six rounds of the comprehension panel ruled each attempt worse. Tried and ruled worse: a module list carried per category from each source's manifest (two routes, `walked` and `indexed`); the union of every manifest's list on the `Generator`, noted at four entry points; and the same noted inside `loadReached`, its list threaded through `loadInline` and `structCompile`. Each named every module read, even loaded ones, until the list held namespaces and the note named only those nothing in the tree provides, which the product-owner and goals reviewers accepted. A shipped locale's `Modules()` already carries its places, so this serves a bare `FS` beside a stand-in, a `--data-path` folder and anyone else's module; the error names the module and points to `Modules()`, as the 2026-10-11 bench asked. Bench the manifest key and the note's wording with the data author, since entries changed from import paths to namespaces after the 2026-10-11 bench.

### 223. Let a column draw a key from another record's rows, so an order's `user_id` names a user the same run wrote.

From the 2026-10-11 audience bench: `"user_id": "{/users.id}"` draws a fresh user for each order, so its `{seq()}` id names no user the run wrote. The workaround types the user count twice, in `-n` and in `{int(1,N)}`.

### 91. Accept a path through a name as a `calc` operand, so `{/misc.coordinate as c}{c}: {calc(c.lat * 60)}` computes from the coordinate `{c}` prints.

A calc operand parses as a bare identifier, so `{calc(c.lat * 60)}` fails with `unexpected ".lat * 60"`, and a second binding of `/misc.coordinate.lat` would draw another coordinate. Resolve the path through `compileArm`'s leaves, as `operandNodes` does for a bare name. A data author who bound a row, `{.product as p}`, then bound its column, `{p.price as pr}`, got `no field "p"; a binding names a field, or a reference…` though `p` is a bound name.

### 271. Give `geo/FR`'s places their streets.

The Base Adresse Nationale (Licence Ouverte 2.0) holds every street with its code postal.

### 280. Give `geo/MX`'s places their streets.

INEGI's Marco Geoestadístico names the streets.

### 285. Give `geo/BR`'s places their streets.

IBGE's CNEFE lists every address with its CEP.

### 295. Give `geo/DK`'s places their streets.

Danmarks Adresseregister through Datafordeler (free) holds every street with its postnummer.

### 300. Give `geo/FI`'s places their streets.

Digiroad (CC BY 4.0) names the streets. A street carries its Finnish name, plus a Swedish-name column where the municipality is bilingual.

### 325. Give `geo/IS`'s places their streets.

Staðfangaskrá, Iceland's address register, holds every street with its póstnúmer.

### 330. Give `geo/DE`'s places their streets.

Each Land publishes its own address data. Where open registers cover every Land, they are its street source; else it ships without streets (decision: no open address register covering it whole), and this item goes.

### 53. Let a table column carry a `datatype`, so `--format json` writes `"safe": true` and `--format sql` a boolean.

Today only a JSON field takes one, so `misc.httpmethod`'s booleans are typed in Go and text everywhere else. Typing a shipped column changes what `json` and `sql` write, so the capability is a minor and applying it to `misc.httpmethod` is a major.

### 170. Decide in one function whether a read lands in a name's pick, which `refuseTwiceDrawnIn` calls at load and `keptInPick` at render.

All 13 seats of the 2026-10-09 panel on b9d1af6 named `keptInPick` and `readUnder` hardest, and 12 named the rule both apply as the code they would least want to change. A key `addressedKeys` builds at load must equal the one `pickKey.under` builds at render, and a mismatch splits a correlated draw without an error. `refuseTwiceDrawnIn` applies the same rule at load in code of its own. In the new function's comment, say why a fresh read's levels start at its head, `levelKeys(…, 1)` in `pathArm`, and a named read's at its target, `levelKeys(full, 0)` in `compileArm`, which three seats rebuilt by hand.

### 178. Plan with the systems architect which `internal/` packages the engine splits into, by area, and how each type stops being written by more than one stage, then file the moves as items.

The maintainer chose on 2026-10-09 to regroup the engine into packages, untangled first, over naming the root files by area. What the panels of 2026-10-09 on b9d1af6 found:

- The 25 flat root files are the worst level of the tree; both architect seats grouped them by hand into public API, load, names, render, paths and proofs.
- The named-pick machinery spans `name.go`, `namefence.go`, `pick.go`, `env.go`, `render.go`, `path.go` and `record.go`, and every seat found it hardest.
- `template` is filled in four phases by four functions, `readsColumn` has four readers in three files, and `renderEnv` shares one `frameStack` by pointer behind a by-value env.
- Calc lives in five places, and paths are walked by `compiledPath`, `callerPathSteps` and `drawSteps` over one `compileStep`.

The decision "The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`" says why the engine stayed whole: `template`, `table`, `arm`, `op` and `nameBinding` sit in one cycle, since several load stages each write part of the same structs. The plan answers four questions. How does load finish each value before render reads it? Which packages does that give, and what is each one's fan-out? How does the allocation gate hold? Which of items 170 to 175 does it answer or close? It lands as a `docs/decisions.md` entry replacing that decision, and as items under item 107, each shipping with a paired ruling no worse.

### 194. Let a data author write and check their own `.fejkdata.json` index with the library or the CLI, and fail a load whose index leaves out a category file.

Goal 6.3 has an author prove their module, and goal 9.1 has fejkdata's data load as anyone's; today only fejkdata's own test, `TestShippedManifestsAreCurrent` over `indexEntry`, writes and checks an index. A category file the index leaves out never loads, silently.

### 29. Let a path go from a selected row into the template beside its family: `geo.US.locality[1714000].address`.

Require the path step to reach a sibling category.

### 56. Build on a manual run of `test.yml` on `main`.

A manual run diffs against `origin/main`, so on `main` itself it compares `HEAD` with itself and skips `docker build`.

### 107. Move the template engine into the `internal/` packages item 178 plans, untangled so no type is written by more than one stage, leaving the root the public API and its options.

Item 178 splits it into items before any code moves. The engine returns column values and the root builds `Record`. Revises the decision "The template engine stays the root package until item 107; what reads no engine type, a table's rows included, sits in `internal/`", and, if the vocabulary moves, "The vocabulary sits below `doc.go`'s package clause, not in the package doc".

### 177. Run the nine-seat comprehension panel and the scoring run again on the moved engine, and file what the panel names.

The decision "comprehension floor: every dimension and the overall at 7.0 or above is the target, and blocks no merge" holds the scores this run is measured against.

### 36. Report every mistake a load finds in one error, as the decision "A load reports every mistake at once" states.

`New` and `NewTemplate` stop at the first, so data holding two mistakes takes two runs to fix. Split it into items before starting, one of them the shape a Go caller iterates and the CLI prints, and the order of an inline template's mistakes, which name no path; reword the README's "`New` refuses a mistake in the data" to every mistake.

### 171. Handle a template that only reads one record column in one place, so `checkColumns`, `itemDatatype`, `proveColumnItem` and `renderLeaf` no longer each special-case `readsColumn`.

Eleven seats of the 2026-10-09 panel named the "change them together" in `readsColumn`'s comment (`node.go`) as action at a distance across `datatype.go`, `valueproof.go` and `render.go`.

### 197. Replace the namespace mount with one the loader states in one place, and give `datafiles.Source.at` one job.

Item 182 shipped on 2026-10-11 over a paired ruling with Shape fallen, by the maintainer's decision recorded in the comprehension-floor entry. Three seats named the same two costs: `mounted` in `internal/datafiles/mount.go`, a partial `fs.FS` that fakes the folders above a module's namespace and serves only `ReadDir` and `ReadFile`; and `Source.at`, which is at once the mounted flag, the prefix `labelled` strips from error labels and the folder `Manifest` reads from. Tried and ruled down within item 182: doubled folders (`data/sv_SE/sv_SE/`), the mount as a prefix on `Category.Folders` that `Load` strips (two coordinate systems), and a complete `fs.FS` wrapper. Whatever replaces it keeps errors naming files where they are on disk for a data path, and keeps `index.go` unaware of namespaces.

### 201. Place each Swedish street in the postort whose area holds it, so a big postort such as Stockholm carries its own streets.

GeoNames gives Stockholm's 1,139 postal codes 2 points and Göteborg's 758 one, and `streets()` in `geo-se.py` counts each road segment toward the nearest postort's point. Every street then ships at most 363 names under Stockholm, the rest going to Bromma, Hägersten and other neighbours. Find an open source of areas first, such as SCB's tätort polygons (CC0), and how a tätort maps to a postort; record the source in `DATA-LICENSES.md` and the method in the decision on the default embed. The README's warning that a street may sit under a neighbouring postort is revised to match.

### 167. Lower the reading cost of the files a chunk reads most before its first commit.

`reading-cost` on 2026-10-11: the last 5 chunks' median is 91452 new input tokens before their first commit, the 5 before them 69756. Read most over the last 5 (opens and searches): `README.md` 37, `todo.md` 29, `docs/decisions.md` 15, `CHANGELOG.md` 12, `data_test.go` 11, `shipped_test.go` 8, `cmd/fejkdata/main.go` 7, `data-import/geo-us.py` 7, `DATA-LICENSES.md` 5, `bench_test.go` 5. The chunk shipping this item runs `reading-cost --reset`.

### 175. Rewrite the comments the 2026-10-09 panel needed several reads of, and delete those it read as restating their code.

Reread: `readField`'s note on clearing the pick (`render.go`), `pickKey.under`, which three seats read backwards, `scopeAround`'s "short of the frames n renders itself", `enter` returning `env.base` as the pop mark, and add a line at `walkEvery`'s three-index slice, which three seats found unmarked. Restating: `compileCells`, `columnOrLinked`, `render`'s first line, `categorySites`, `madeFolder`, `categoryOf`, `rendersInside`, `templateSite`, `templateOptions`, `compileFrom`, `flagArg`, `drawSeed`, the `calc` parser's precedence line in `internal/grammar`, and `compileChoice`'s note that a uniform choice draws in O(1), which `choice`'s doc already says. A doc comment `go doc` shows on an exported name stays.

### 26. Spell `sv_SE.personnummer` with the `+` Skatteverket uses from the year the holder turns 100, judged by a date the caller gives.

The format hard-codes `-`, and its earliest birthdate, 1930-01-01, makes the oldest draws invalid from 2030. A system holding Swedish personnummer often breaks on the `+`, so a fixture carrying one is a valuable test. Whether a number takes `+` depends on the date it is read, and the decision "No builtin reads the clock, so a date is bounded by days, never by an age" forbids a builtin from reading that date off the clock. This item revises it: once the caller gives the date, an age is a bound, and the CLI's default date makes output change with the day it runs unless the flag is given. Do the first step before the rest:
- Explore how a caller passes input data, such as the date a number is read on, to a render: in the template, as a plain option, or both. Make the library require the date, as it requires the seed, and let the CLI default it to the system's when its flag is absent. Bench the spelling; `--now` reads like the clock. Shape the input so it can later narrow the draws to adults, children, 65 and over, or a mix, and so let a test ask for a centenarian; narrowing itself waits for its own item.
- Draw birthdates back past 100 years before the given date, so a `+` can appear. Read in 2026, the oldest shipped birthdate is 96 years old, so no draw carries a `+` until 2030. Skatteverket's test series reaches back to 1890, with birth numbers 980/981 before 1900. Draw no birthdate after the given date.
- Weight the ages by the territory's population, so the share of `+` numbers follows its demography: [research-age-bands.md](docs/research/research-age-bands.md).
- Let a data author's own category spell the separator too, through a builtin or a function a data module provides (item 186), and update the README's personnummer recipe and its "a day in 1930–2025" count.
- Rewrite the README's `--seed` and `WithSeed` rows to goal 10's wording, and name the new required option in the changelog: `New` breaks again here, the first time after 0.1.0, as the maintainer accepted.
- Research: `sv_SE.samordningsnummer` also hard-codes `-` and draws from 1930. Check whether Lag (2022:1697) om samordningsnummer gives it the same `+`; folkbokföringslagen 18 § states the rule for personnummer only.

### 172. Pass `drawSteps` the kind of read it draws as an argument of its own, so no caller signals a sibling path by a nil `pins` or a fresh read by a nil `memo`.

Ten seats of the 2026-10-09 panel named `drawSteps` (`path.go`); three named the nil arguments as the mode they had to infer.

### 216. Share between countries what `geo-se.py` and `data_test.go` hard-code for SE, in the first locale chunk that needs each.

A second country reading GeoNames' postal codes moves `geonames()`, and the `streets()` wrapper around `geo.Nearest`, from `geo-se.py` into `geo.py`. `data_test.go`'s `geoPlaces` reads every country but US by SE's column positions, and a locality is keyed by name in SE and by code in US: a new country states its own.

### 251. Export from the library whether a reference body is a lone path, so the CLI's `isPath` stops copying the path grammar.

The inherited architect of the 0.1.0 scoring run on 52e0304 named `isPath` (`cli/cli.go`) the CLI's first rot point when the path grammar changes: it copies the grammar because the CLI reads only the library's public API.

### 108. Name item 26 in the decision "No builtin reads the clock, so a date is bounded by days, never by an age", as the item that ends its conflict with goal 4.

Item 26 revises the decision, but the decision does not point back to it.

### 173. Call what a token reads its spelling, never a name, in `format.go`, `arm.go` and `reference.go`, so a name in the engine is only a bound name.

`checkReads` "proves each name t reads", `splitArm` "splits one name into head and tail" and asks `t.isName` of the result in `compileArm`, and `refTokens` returns "reference names". Both architect seats of the 2026-10-09 panel and the scoring run's architect named the word as one name over several concepts, against the decision "One name, one meaning". Item 136 did this for `internal/grammar`.

### 75. Run every README `sh`, `go` and `text` example as a test, and compare each "Renders" line to a seeded render.

`readme_test.go` loads and renders each `json` block, but runs no `sh` example (CLI, Records, Table, Linked tables), no `go` example (Library) and no `text` block of error text, and compares no "Renders e.g." output, so goal 12.2 is not met.

### 174. Spell once the fact that a walk's next step leaves a selected row, which `pathPos.fromRow` and `tableRoute.nextLinked` each spell today.

`compileStep` copies one into the other (`path.go`). Two seats of the 2026-10-09 panel named it.

### 193. Move the largest shipped packages into nested Go modules that keep their import paths, once `TestShippedDataStaysUnder100MiB` fails.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Each nested module requires the core release that drops its package, so no build sees the package twice. It brings per-module tags and a workspace, so it moves into the release of the chunk whose data trips the gate.

### 195. Parse each builtin argument once, in `checkArgs`, and hand `prep` and `proveNumber` the parsed value.

Applies "Validate data and build DTOs as early as possible" in `~/.claude/principles/technical-principles.md`. Today `checkArgs` (`internal/builtinfunc/builtins.go`) checks an argument, then `prep` and `proveNumber` parse it again through `atoi`/`atof`, which panic on bad input; `layout.go`, `transform.go` and `calc.go` call them too; two parsers can drift, and a drift panics at compile instead of failing the load.

### 17. Give every shipped category a column for each part its format composes, and one record shape across the national ids; re-pin the shape.

`misc.uuid` gains `variant`, for one.

### 72. Fail the merge gate when test coverage falls below the last recorded figure.

`technical-principles.md` holds that test coverage should not decline. `docker compose run --rm cover` prints the figure, and neither the `Dockerfile` gate nor `.github/workflows/test.yml` checks it.

### 11. Test that every `geo` tree holds the five table names and the `address` columns that port across countries.

Nothing checks it, and a new country breaks it silently.

### 40. Parse `defaultTable`'s path with the library's grammar.

The CLI counts brackets of its own, so a change to the selector grammar desyncs the default `--table` silently.

### 86. Read both `en_US.phone` formats' area codes from one category, as their exchange reads `en_US.phone-exchange`.

Both phone formats carry the same 21-code `area` list, so an edit to one copy can miss the other.

### 95. Hold a struct's columns and the field each fills in one slice.

`structShape.fieldIndexes` pairs with its record template's `columns` by index alone, against the technical principle "one owner per value".

### 112. Check a calc's field operands where its template resolves, beside its name operands.

`checkCalcFields` runs at compile and `checkCalcNames` in `resolveTemplates`, two passes over one rule. Moving the field check there before item 36 would change which of two mistakes in two category files a load reports, since every category compiles before any resolves.

### 116. Turn a row index into its rows file's line in one function.

Today `containedNode`'s cell line in `graph.go`, `table.compileCells` and each site in `internal/rows` add the header line and the one-based count.

### 186. Load a module carrying functions through an option of its own, its functions reaching the generator's randomness and date through an interface the root declares.

Carries out the decision "The shipped data is Go packages a library imports by choice, and the CLI carries every one, from 0.1.0". Its functions get the date item 26 has the caller give. `cli.Main`'s signature grows here, which a minor before 1.0 allows.

### 168. Test the shared steps of `data-import/` in the merge gate, starting with `geo.fence`.

Against "Most, if not all, code should have automated tests" in `technical-principles.md`: no Python under `data-import/` has a test, and the gate, `docker build .`, runs Go alone. Today `geo.fence` is checked only by regenerating both countries' TSVs from the full downloads and diffing them. Each locale's places item adds a geo script for its country, each calling `fence`.

### 218. Give every country shipped without streets its streets, once the maintainer accepts ODbL data or the country opens a register.

Ends the decision "A country with no open address register covering it whole ships its places and postal codes without streets".

### 129. Take a template wherever the library renders, so `fake:"{/sv_SE.person.first}"` and `Fake("{/sv_SE.person}")` render, and text with no `{…}` prints as written, as the CLI already does.

Goal 5.9. Today a path is a second spelling of the template that reads it; afterwards only the template remains.

- A lone reference, a template of one reference token and nothing else, is what it references, as a record column reading one reference already is: `FakeRecord("{/users}")` returns the columns of `users`. The CLI's `loneReference` and the `isPath` check it copies from the token grammar then go, and `--format sql` takes its default `--table` from the library's parse of the reference.
- `Fake` and `FakeRecord` take a template. `FakeTemplate` folds into `Fake`, `FakeRecordTemplate` into `FakeRecord`, and `IsTemplate` goes, with `grammar.IsTemplate`.
- `Fake` of a lone reference costs what `Fake` of its path costs today, and the allocation gate holds a case for it.
- `List` and `--list` give each entry as the template that renders it, `{/sv_SE.person}`.
- Text that is not a JSON value and holds no `{…}` prints as written (goal 5.1), in a string field's struct tag and `Fake` alike, as stdin's text does in the CLI: `Fake("sv_SE.person")` returns `sv_SE.person`. `CHANGELOG.md` says that a path with no `{…}` prints as its text. A struct tag's `-` still skips its field.
- Where text with no `{…}` is refused, as a record or in a typed struct field, and it spells a path the data offers, the error names `{/that.path}`.
- A lone reference to nothing exits 2 under `--format`, where it exits 1 today.
- A lone reference written as an object holding only a format, `{"format":"{/users}"}`, or as a one-item choice, `["{/users}"]`, names the record as `{/users}` does.
- A struct tag of a bare JSON number or boolean, `fake:"42"`, renders its text, as stdin and `NewTemplate` already do.
- `grammar.CallerPath`, which strips a path's leading `/`, goes with the path spelling, and so does the README's Library line allowing it.
- Decide whether a lone reference under `--format` draws as its path draws today or as a template draws: under one seed, drawing `users` as a path and as the template `{/users}` picks different rows.
- This revises the decisions "A struct tag is a template by its shape", "`FakeTemplate` and `NewTemplate` both stay", "`--list` stays a plain list of paths", "The CLI renders exactly the template on its stdin, and adds nothing", and the exit-1 rule for a lone reference to nothing in "A template that does not compile is misuse (exit 2), including a reference that resolves to nothing, save a lone reference under the CLI's `--format` until item 129". The README's examples and exit codes, the godoc examples in `doc.go` and `CHANGELOG.md`'s entries move to the template spelling. Delete items 40 and 110, which this closes.

Split it into items for the CLI, struct tags and the library before starting.

### 41. Key every table by one key column, as the decision "A table selects by one key column, a code, never a free-form name" states.

`geo.SE.municipality[Lund]` and `misc.territory[Sweden]` select a row today. Steps:
- Have a selector that misses name the key of every row whose cell spells it, matched without case, else say the table selects by its key: `misc.territory[Sweden]` and `misc.territory[se]` name `[SE]`, `misc.protocol[tcp]` its key, and `misc.loglevel[Error]` `[err]`.
- Delete the `name` option and the load checks it carries, and have `New` refuse a table still carrying `name`, naming `key`.
- Key `misc.loglevel` on `keyword`, and `geo.SE.locality` on the integer `data-import/geo-se.py` assigns as the decision states, linking `street` and `postal-code` by it.
- Drop `key` from `last-name` and `sv_SE.title`, and re-pin `testdata/shipped_shape.txt`.
- Leave no selection by name in README.md, CHANGELOG.md, docs/ or `--help`.

### 109. Indent every script under `data-import/` with hard tabs, as `AGENTS.md` requires and `release-tooling/` does.

Technical principle: indent with hard tabs. All 19 scripts indent with four spaces, so an edit there either keeps the wrong style or rewrites the whole file.

### 14. Spell `misc.creditcard`'s digit runs `{digits(n)}`.

A `"repeat": 14` over `{digits(1)}` draws every 14-digit run with the odds `{digits(14)}` does. Both load; make the shipped data write the shorter one. Re-pin seeded output in its own commit.

### 59. Promise in `Generator`'s godoc that its renders run one at a time.

A caller wanting parallel throughput then makes one generator per goroutine.

### 62. Add `{btc()}` and `{eth()}`.

They need sha256 and keccak for Base58Check and EIP-55.

### 65. Hold every change `AGENTS.md` says owes a `CHANGELOG.md` entry to one in CI.

The check covers `data` and `testdata/shipped_shape.txt` alone, where `AGENTS.md` adds a flag, an exit code, an exported name, a fence, a builtin and the lowest Go.

### 68. Make `gitea.larvit.se/larvit/fejkdata` a pull mirror of GitHub.

Gitea converts a repository to a mirror only by re-creating it, so until that runs the copy there is a second owner of one history.

### 63. Pair a street with its exact postnummer.

Needs an application to Lantmäteriet; today a street goes to the nearest postal code centroid. It rewrites shipped rows, so it is a major once 1.0 is cut and a minor before that.

### 66. Cut the README's Layout block to the lines that say what a file name cannot.

`doc.go` restates its name.

### 67. Cut `AGENTS.md` to the rules only it states, and explain every term its comprehension rule uses.

The `gh` rule opens by restating the decision "GitHub is canonical, and the module path names it", a second copy that drifts. The comprehension rule packs the gate, its exceptions, the round and the at-or-above-7.0 regime into one line, and uses "depth 1", "nine-seat", "four-seat" and "answers it in one run" without saying what they mean.

### 64. Give every entry in `docs/decisions.md` the goal it serves, its date, who made it and the premise it rests on.

Propose a goal for an entry that serves none.

### 334. Add `nn_NO`, written in Nynorsk, reading `nb_NO`'s names and id and `geo/NO`, with month and weekday names from CLDR and titles of its own.

Nynorsk is Norway's second written standard, the official form in some municipalities. Names, places, streets, ids and phone numbers read the same in both standards, so only the words that differ need a source of their own.

### 69. Announce v1.0.0 where a developer choosing a fake-data tool already reads, with a README first screen for someone deciding in a minute.

The human cuts v1.0.0 once the shipped data is in its record shape and one full minor has shipped with no breaking change, per the README's Versioning table; pairing a street with its exact postnummer rewrites shipped rows, so a minor with no breaking change follows it first.

### 190. Decide whether fejkdata pseudonymises real datasets, and if so add the goal, the audience persona and the items below to the README and the plan.

Pseudonymising a copy of production is work a hosted model cannot do well: sending the data out is the leak, a model cannot keep one person's fake identical across tables and runs, and millions of rows cost real money. A model can still name which columns hold personal data and of what kind; fejkdata then replaces them offline. Greenmask, PostgreSQL Anonymizer and Presidio cover parts of this; none draws Swedish values a real validator accepts.

Proposed goal, its place among the others for the maintainer to choose:

> **It pseudonymises a real dataset on your own machine: each personal value becomes a valid fake one.**
> 1. Columns you do not name pass through byte for byte.
> 2. The same value under the same key becomes the same fake, in every table and every run; without the key, nobody can tell which real value a fake came from.
> 3. Values that belong to one person come from one pick, so a row's name, email and personnummer still agree.
> 4. Two values never become one fake in a column that must stay unique.

Proposed audience persona: a **pseudonymising author**, who replaces the personal data in a copy of production before it reaches a test or staging environment, and relies on one real value giving one fake across tables and runs.

Items to file once the goal is in:

1. Read CSV and NDJSON records on stdin, replace the columns a mapping names with a template's render, and stream every other column through unchanged. The flag or file spelling goes to a bench with the persona first.
2. Seed each replaced value's draw from HMAC-SHA256 under a key read from a file or the environment, never a flag. An unkeyed hash fails sub-goal 2: every personnummer can be hashed and the mapping reversed.
3. Key a whole row's pick by one identity column, such as a customer id, so goal 4.1's named pick holds for a person across rows and tables.
4. Redraw on a collision in a column marked unique, deterministically from the key, within a memory bound the README states.
5. Keep, on request, a personnummer's birth date or sex and a postal code's region, the docs naming what each kept field gives away.
6. Read a PostgreSQL `COPY` or MySQL `INSERT` dump.
7. Name the output pseudonymised in the README, never anonymised: under GDPR, data is personal while the key exists, and a rare mix of kept fields can still identify someone.
