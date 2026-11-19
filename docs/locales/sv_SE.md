# `sv_SE`

The categories it carries are listed under [Locale categories](../../README.md#locale-categories).

## Bankgiro, plusgiro and bank accounts

`sv_SE.bankgiro` draws a bankgiro number of seven or eight digits at even odds, `991-2346` or `5555-5551`, the last
digit a Luhn check digit. `sv_SE.plusgiro` draws a plusgiro number of two to eight digits, the last a Luhn check
digit, with a hyphen before it and the rest in pairs from the right, `12 34 56-6`. Wikipedia and convention give that
length and spelling; no official page does.

`sv_SE.bankaccount` draws a clearing number and an account number that passes the check its bank demands,
`1522-8923497`. `{/sv_SE.bankaccount[5000]}` selects a clearing number, and a row's fields are `clearing`, `bank` and
`number`. The clearing numbers and banks are those of Bankinfrastruktur i Sverige AB's list, and the checks are those
of Bankgirot's "Bankernas kontonummer" of 2024-02-22. A Swedbank clearing number from `8000` to `8999` has a fifth
digit that this category leaves out. The 15 clearing numbers `9565`–`9569` and `9600`–`9609` are not shipped, since
that document doesn't publish their check.

`bankgiro`, `plusgiro` and `bankaccount` draw valid numbers, not real ones: a number drawn may belong to a real
account.

## Blood types

`sv_SE.bloodtype` draws `A+`, `0-` and the six other types, spelled with the digit `0` as Swedish does, as
often as sv.wikipedia's Blodgrupp article gives them for Sweden: `A+` is the commonest, at 37 percent. The
shares are real.

## Cars

`sv_SE.car` is a make and a model, `Volvo XC60`, weighted by Mobility Sweden's registrations of new
passenger cars in 2025. It holds the 119 models with 500 or more registrations, as Mobility Sweden groups
them: `EX/XC40` is two models in one row. A model is real, spelled as its maker writes it, and the shares
are those of 2025's new cars, not of the cars on the road.

## Industries and occupations

`sv_SE.sni` is SCB's SNI 2025 industry classification at its finest level, 835 five-digit codes, and
`sv_SE.ssyk` its SSYK 2012 occupation classification at four digits, 429 codes. Each is selected by its
code: `{/sv_SE.sni[62.100].name}` is `Dataprogrammering` and `{/sv_SE.ssyk[2512].name}` is
`Mjukvaru- och systemutvecklare m.fl.`. A draw picks a code evenly and prints it with its label,
`62.100 Dataprogrammering`. Both are real.

## Licence plates

`sv_SE.licenseplate` draws a registration number, `ABC 123` or `ABC 12A` at even odds, in its fields `old` and
`current`. It is valid, not real: a drawn plate has the shape of a Swedish one and may be on a real car. Its three
letters come from `licenseplate-letters`, every combination of the 23 letters plates use (no I, Q or V) but
the 99 that Transportstyrelsen's page listed as blocked in March 2026 and no longer lists, such as `APA` and `XXX`.
The last character of `ABC 12A` is any of those letters but O. The numbers run from `001` to `999`, and from `00` to `99`
in `ABC 12A`. The forms, letters and numbers are those of sv.wikipedia's article on Swedish plates, which
Transportstyrelsen's pages do not confirm.

## Months, weekdays, holidays and territories

`sv_SE.month` and `sv_SE.weekday` are Unicode CLDR's Swedish names. `{/sv_SE.month[1]}` is `januari` and
`[12]` is `december`; `{/sv_SE.weekday[1]}` is `måndag` and `[7]` is `söndag`, Monday first as ISO 8601
numbers them. `sv_SE.holiday` holds the thirteen public holidays that lag (1989:253) names, each selected by a
code in lower case ASCII with hyphens, `{/sv_SE.holiday[nationaldagen]}` or `[forsta-maj]`. It leaves out the
law's entry for every Sunday, and holds no dates, since seven of the holidays move from year to year.
`sv_SE.territory` is the Swedish name of each `misc.territory` row, selected by its alpha-2 code:
`{/sv_SE.territory[DE]}` is `Tyskland`. All four are real names.

## Organisationsnummer and VAT numbers

A bound `organisationsnummer`
has its own VAT number in `vat`, and `sv_SE.vat` draws one with no organisationsnummer.

## Person

`sv_SE.person` draws one invented birth date from the four age bands, and its `birthdate`, `personnummer` and
`samordningsnummer` read it. Both numbers spell the person's sex from the same test series, with the limits
above, and the samordningsnummer adds 60 to the day, so `{/sv_SE.person as p}{p.birthdate} {p.personnummer} {p.samordningsnummer}` renders e.g.
`1932-04-16 320416-2394 320476-2391`.

## Personnummer and samordningsnummer

A `personnummer`'s birth number, `sv_SE.birth-number` under `sex`, is Skatteverket's test series, 238 for a woman and 239 for a man, which no real person is ever given.

`sv_SE.personnummer` and `sv_SE.personnummer-12` draw four age bands as often as SCB counted them in Sweden
at the end of 2024, and a birthdate evenly inside each: 20.36% aged 0–17, 58.80% aged 18–64, 20.81% aged
65–99 and 0.028% aged 100 to 113, the oldest age validated in Sweden. Ages count as [`born`](../../README.md#functions)
counts them, from the date given to `WithToday` or `--today`, so what a seed draws changes with that date. The short form spells `-`, and `+` from
the year the holder turns 100, as folkbokföringslagen 18 § states: with today at 2026-11-04, a birth in 1926
or earlier carries the `+`. A date before 2013 can draw a birth before 1900, where the test series
holds other birth numbers, 980 and 981 among them, so such a number is outside the series. A birth in a
year Skatteverket hasn't published yet is outside it too; it adds the coming year each December.
`sv_SE.samordningsnummer` draws a date from 1930 to 2025 and always spells `-`, since lag (2022:1697), the
law that governs it, states no `+`.

Skatteverket's test series is small: a personnummer is one of about 83,000 values, a day in the 114 years up to the given date against the two birth numbers, so a fixture past a few hundred rows repeats one and a `UNIQUE` column needs a category of your own.

## Personnummer and samordningsnummer, twelve digits

`sv_SE.personnummer-12` and `sv_SE.samordningsnummer-12` draw a personnummer and a samordningsnummer as twelve
digits, the century first and no hyphen, `199001012385`: the form many systems store. As in the short form, the Luhn
check digit covers the last ten digits, which `sv_SE.samordningsnummer-12`'s field `ten` holds.

## Personnummer of a sex or an age band

`sv_SE.personnummer` is a table under `sex`, so `sv_SE.sex[f].personnummer` draws a woman's number, and a
bound sex gives a name and a number that agree:

```json
{ "format": "{/sv_SE.sex as s}{s.first-name} {/sv_SE.last-name}, {s.personnummer}" }
```

Renders e.g. `Love Fogelström, 670711-2394`.

Each age band is a column: `child` for those who turn 0 to 17 in the year of `--today` or `WithToday`, `adult` for 18 to 64,
`senior` for 65 to 99 and `centenarian` for 100 to 113, which always carries the `+`. `sv_SE.personnummer.centenarian` draws one of
either sex, `sv_SE.sex[f].personnummer.centenarian` a woman's and `{s.personnummer.centenarian}` one of the bound sex. For a mix of your own, weigh
the columns in a choice:

```json
[
  "{/sv_SE.personnummer.child}",
  { "format": "{/sv_SE.personnummer.centenarian}", "weight": 0.1 }
]
```

Renders e.g. `251006-2389`. `sv_SE.personnummer-date` is the date and separator that `sv_SE.personnummer`
starts with: one of the four bands, drawn by its share of the population.

## Swedish companies

`sv_SE.company` draws a legal form by Bolagsverket's new registrations of 2025: aktiebolag,
enskild näringsidkare (a name with no legal-form word), handelsbolag, bostadsrättsförening, ekonomisk
förening and kommanditbolag. It spells the form as the register's active companies do, so a handelsbolag
reads `Handelsbolag`, `Handelsbolaget` or `handelsbolag`, never `HB`. Bind a company to read its legal-form word from the same pick:
`{/sv_SE.company as c}{c} ({c.legal})` prints `Solvik Bygg AB (AB)` or
`Handelsbolaget Solvik (Handelsbolaget)`, and `()` for an enskild näringsidkare.
The name has a shape those companies' names have, such as a made-up
word and a trade word like `Bygg` or `Holding`, or a surname and a trade word. `company-word` holds the
trade words that 100 or more of the register's names carry, and `company-coined` the made-up words.
A name is invented, and may also be a real company's.

## Swedish phone numbers

`sv_SE.phone` is a mobile number ten times as often as a fixed one. A fixed number starts with an
area code drawn evenly from `sv_SE.phone-area`, which holds the 264 geographic area codes in the
Swedish Post and Telecom Authority's (PTS) numbering plan of 2024-01-08. A row is selected by the
code as Swedes write it, `phone-area[08]`, and carries its `ndc`, the code without the trunk `0`
that international dialling uses (`8`), and the `name` of its place.

The rest of a fixed number comes from `sv_SE.phone-subscriber`. It has any length the plan allows
that area code, five to eight digits, and is grouped as Språkrådet writes it, such as `08-16 01 01`
and `031-668 01 50`. No subscriber number starts with the trunk `0`. A number is valid, not real: it
follows the plan, and a real subscriber may hold it. A bound number reads `national`,
`070-123 45 67`, and `international`, `+46 70 123 45 67`.

## Swedish words

`sv_SE.word` draws evenly from about 108,000 words of SALDO, Språkbanken's Swedish dictionary.
SALDO has no frequencies, so a word may be rare. `sv_SE.word.noun`, `.verb`, `.adj` and `.adv` each
draw one part of speech. `sv_SE.word.noun.definite`, `sv_SE.word.verb.present` and
`sv_SE.word.adj.definite` each draw a word and print one form of it, such as `huset`, `springer` or
`stora`. To read a word and its form from one pick, bind it: `{/sv_SE.word.noun as n}{n} {n.definite}`
prints `hus huset`.

The import drops the words on LDNOOBW's Swedish list of rude words, and other rude words remain, so
read the output before showing it to anyone. `sv_SE.sentence` fills a fixed pattern with these words
and doesn't make them agree, so its sentences aren't grammatical. `sv_SE.paragraph` joins three to six sentences.
