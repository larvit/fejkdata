# `nb_NO`

`nb_NO`, Norwegian in the Bokmål standard, carries `address`, `birth-day`, `color`, `company`, `company-coined`,
`company-word`, `date`, `email`, `first-name`, `fodselsnummer`, `individ`, `landline`, `last-name`, `mobile`, `month`, `paragraph`, `person`,
`phone`, `phone-fixed`, `phone-mobile`, `price`, `sentence`, `sex`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`.
`fodselsnummer` reads `birth-day` and `individ`. `title` has no `parent`, like `sv_SE.title`, so it is selected as
`nb_NO.title[dr.]`.

## Fødselsnummer

`nb_NO.fodselsnummer` is Skatteetaten's synthetic series, the numbers of its test register. A
number is the day, the month plus 80 and the year of birth, three individ digits and two check
digits. The individ digits give the century. Skatteetaten assigns 000 to 499 to 1900 to 1999,
500 to 749 to 1854 to 1899, 900 to 999 to 1940 to 1999 and 500 to 999 to 2000 to 2039; fejkdata
draws 000 to 499 for 1940 to 1999 and 500 to 999 for 2000 to 2025, never on 29 February. Its
`birth` draws three bands, 1940 to 1999, 2000 to 2019 and 2020 to 2025, each weighing its count of
years, so every year from 1940 to 2025 is equally likely. The
last individ digit is even for a woman and odd for a man. No real person holds a number from
this series, and a checker that requires a month of 1 to 12 refuses it. A number is one of
about 13 million, so a fixture of a few thousand rows starts to repeat one, and a `UNIQUE`
column needs a category of your own.

`nb_NO.birth-day` holds the 365 days of a year as `ddmm`, with 80 added to the month.

## Fødselsnummer matched to a sex

`nb_NO.individ`, under `sex`, holds the five digits each sex's last individ digit can be. To
make a number agree with the sex the template below binds as `s`, read the `birth` field of `nb_NO.fodselsnummer`, the
date and the first individ digit, then `{fnr(x)}`, which completes the number from the third
individ digit `x` reads:

```json
{ "format": "{first} {last}, {fnr}",
  "first": "{/nb_NO.sex as s}{s.first-name}",
  "last": "{/nb_NO.last-name}",
  "fnr": "{/nb_NO.fodselsnummer.birth}{fnr(s.individ.digit)}" }
```

Renders e.g. `Kjersti Østrem, 07899027420`.

## Norwegian companies

`nb_NO.company` draws a legal form as often as the Brønnøysund Register Centre's Enhetsregisteret held
active units of it on 2026-10-22: sole proprietorship (ENK), limited company (AS), and DA, SA, ANS, ASA and
KS. An ENK's name has no abbreviation after it, and the others end in theirs, such as `Nordicvik Holding AS`.
`nb_NO.company.as` draws only a limited company, and `.enk`, `.da`, `.sa`, `.ans`, `.asa` and `.ks` each draw their
form. Bind a company to read its abbreviation from the same pick: `{/nb_NO.company as c}{c.legal}` prints the
`AS` of `Nordicvik Holding AS`, and nothing for an ENK.

A name has a shape that the register's names of its form have, such as a coined word and a trade word, or a
last name and a trade word, in the proportions the register has them. A last or first name is drawn from
`nb_NO.last-name` or `nb_NO.first-name`. A sole proprietor's name always holds a last name, as
foretaksnavneloven § 2-2 requires. A trade word names a line of business, like `Holding` or `Eiendom`, and
`company-word` holds those that 100 or more of the register's names carry. A coined word is made up:
`company-coined` joins a root that 60 or more names begin with, such as `Nordic`, to an ending that 40 or more
words of `geo/NO`'s place names share, such as `vik`. Where the register's names hold a postal town
(poststed) or a municipality (kommune), the shape draws a poststed from `geo.NO.locality`, as in
`Hamar Nordicdal AS`. A name is invented, and may also be a real
company's.

## Norwegian dates and emails

`nb_NO.date` is day and month with two digits each, then the year with four, and dots between, `02.04.2019`, a spelling the
Language Council of Norway allows. `nb_NO.email` is lower-case ASCII at `example.com`, `example.org`
or `example.net`, or a subdomain of one, and half of its local parts are a first and a last name drawn from
`nb_NO.first-name` and `nb_NO.last-name` and folded to ASCII, `ø` to `o`, `å` to `a` and `æ` to `ae`. IANA reserves those domains, so no address reaches a real
mailbox.

## Norwegian phone numbers

`nb_NO.phone` is a mobile number ten times as often as a fixed one. Norway has no area codes, so
both have eight digits. The Language Council of Norway (Språkrådet) has written both in pairs
since 22 May 2019, on its page «Tall, tid og dato»: `46 12 34 56`.

`nb_NO.phone-mobile` and `nb_NO.phone-fixed` hold the first four digits, such as `46 12`, of the
numbers in Nkom's E.164 register, downloaded on 2026-10-22. A prefix is a row only when Nkom has
assigned all ten thousand numbers behind it to providers, as mobile or as fixed numbers. A row is
selected by its prefix, `phone-fixed[56 12]`.

A number's `block` is one of those prefixes and its `subscriber` the last four digits, such as
`34 56`. A number is valid, not real: it lies in a block Nkom has assigned, and a real subscriber
may hold it. A bound number reads `national`, `46 12 34 56`, and `international`,
`+47 46 12 34 56`. The page names no country-code form, so fejkdata puts `+47` before the same pairs.

`nb_NO.mobile` draws a mobile number and `nb_NO.landline` a fixed one, each in the national form. A bound number reads `national` and `international`.

## Norwegian words

`nb_NO.word` draws evenly from about 142,000 words of Norsk ordbank's Bokmål 2005, from Språkbanken:
the common nouns, verbs, adjectives and adverbs it spells in lower-case letters alone, in the forms the
current norm holds. The ordbank has no frequencies, so a word may be rare. `nb_NO.word.noun`, `.verb`,
`.adj` and `.adv` each draw one part of speech, and `nb_NO.word.noun.definite`, `nb_NO.word.verb.present`
and `nb_NO.word.adj.definite` print one form, such as `huset`, `løper` or `store`. Bind a word to read it
and its form from one pick: `{/nb_NO.word.noun as n}{n} {n.definite}` prints `hus huset`.

The import drops the words on LDNOOBW's Norwegian list of rude words, and other rude words remain, so
read the output before showing it to anyone. `nb_NO.sentence` fills a fixed pattern with these words
and doesn't make them agree, so its sentences aren't grammatical. `nb_NO.paragraph` joins three to six sentences.
A program showing these words credits "Språkbanken, Nasjonalbiblioteket: Norsk ordbank bokmål 2005, CC BY 4.0".
