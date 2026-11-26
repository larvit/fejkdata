# `da_DK`

`da_DK`, Danish, carries `address`, `color`, `company`, `company-coined`, `company-word`, `cpr`, `date`, `email`, `first-name`, `individ`, `last-name`, `month`, `paragraph`, `person`, `phone`, `phone-fixed`, `phone-fixed-prefix`, `phone-mobile`, `phone-mobile-prefix`, `price`, `sentence`, `sex`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`, built like [`nb_NO`'s](nb_NO.md):
`da_DK.sex` is `kvinde` and `mand`, and `da_DK.title` is `dr.` and `professor`. No good online source of
Danish name counts was found. The top 20 girls' names, boys' names and last names keep [Statistics
Denmark](https://www.dst.dk/en/Statistik/emner/borgere/navne/navne-i-hele-befolkningen)'s counts of
1 January 2026. Every other name is a language model's recall of a name common in Denmark: the commoner ones come first,
from the most to the least common, and the rarer ones share one weight. These weights are estimates. A program showing the
top 20 counts or the sex counts credits "Kilde: Danmarks Statistik".

## CPR number

`da_DK.cpr` is a Danish personal number as the CPR office defines it in "Personnummeret i
CPR-systemet", `ddmmyy-ssss`. It passes the office's checks but is not synthetic: the office
publishes no test series and no unassigned range, and never reuses a number, so one may be a
living person's. The ten digits, weighted 4 3 2 7 6 5 4 3 2 1, add up to a multiple of 11, and the
last digit is even for a woman and odd for a man. Since 2007 the office also issues numbers that
fail the modulus 11 check; fejkdata doesn't draw them, so a checker that requires the check accepts
every number here.

The seventh digit names the century together with the year digits. fejkdata draws a date of someone
under 100 by Statistics Denmark's shares of under 18, 18 to 64 and 65 and over on 1 January 2026,
and a seventh digit evenly from those the office's table of century digits allows for the year: 0 to
3 for 1900 to 1936, 0 to 4 and 9 for 1937 to 1999, and 4 to 9 for 2000 to 2036. The builtin
`{cpr(x)}` completes a number after its date: it takes the century from the latest year not after
the date of `--today` or `WithToday` that bears the year digits, then adds the seventh digit, an
eighth and a ninth digit it picks, and the check digit, which it gives the parity of the digit `x`
reads. The date comes from `born(0,99,…)`, so a date of someone 100 or older, 0.02% of the
population, is not drawn. A fixture of a few thousand rows starts to repeat a number, and a
`UNIQUE` column needs a category of your own.

## CPR number matched to a sex

`da_DK.individ`, under `sex`, holds one digit of each sex's parity. To make a number agree with the
sex the template below binds as `s`, read the `birth` field of `da_DK.cpr`, the date, then `{cpr(x)}` with the digit that
`individ` gives:

```json
{ "format": "{first} {last}, {cpr}",
  "first": "{/da_DK.sex as s}{s.first-name}",
  "last": "{/da_DK.last-name}",
  "cpr": "{/da_DK.cpr.birth}-{cpr(s.individ.digit)}" }
```

Renders e.g. `Bodil Holm, 141062-2730`.

## Danish companies

`da_DK.company` draws a legal form as often as Statistics Denmark counted enterprises of it in 2024 (its table GF15):
sole proprietorship (ENK), private limited company (`ApS`), public limited company (`A/S`), and `I/S` or `K/S`, which
the table counts together as partnerships and limited partnerships, so they are drawn evenly. An ENK's name has no
abbreviation after it, and the others end in theirs, such as `Nordholm Holding ApS`. `da_DK.company.enk`, `.aps`, `.as`,
`.is` and `.ks` each draw one form, and `{/da_DK.company as c}{c.legal}` prints the abbreviation of the same pick.
No open bulk file of Danish company names was found, so the name's shapes and the words of `company-coined` and
`company-word` are a language model's recall. Each name is made of those words, `da_DK`'s names and
`geo.DK.locality`. A name is invented, and may also be a real company's.

## Danish dates and emails

`da_DK.date` is a date from 1970 to 2029 in the traditional Danish spelling, day, month and year with dots and no zero
before a one-digit day or month, such as `2.4.2019`; the Danish Language Council (Dansk Sprognævn) advises against the
zero. `da_DK.email` is built like `nb_NO.email`: half its local parts are a first and a last name from `da_DK`'s names,
with `æ`, `ø` and `å` folded to ASCII.

## Danish phone numbers

`da_DK.phone` is a mobile number ten times as often as a fixed one. Denmark has no area codes, so
both have eight digits, written in pairs: `41 23 45 67`, and `+45 41 23 45 67` with the country code.
That follows the Danish Language Council (Dansk Sprognævn), whose page "telefonnumre" says it is most
reader-friendly to write a number as it is said, "som regel i grupper af tocifrede tal", and that
the plus and the country code are written together.

`da_DK.phone-mobile-prefix` and `da_DK.phone-fixed-prefix` hold the three-digit prefixes, such as `34 2`, that
Annex 1 of Executive Order No. 1883 of 7 December 2020 on the Overall Danish Numbering Plan
allocates to mainly mobile or mainly fixed numbers. A prefix is mobile where its first two digits
are in a range the plan calls mainly mobile, or where the plan's note lists it as mobile inside a
fixed range. A prefix with third digit 0 is left out, as the plan keeps it for expansion. So are the
ranges the plan reserves, 67, 68, 83 to 85, 94 and 95, and those it gives to other services: 80 to
freephone, 90 to premium rate services and 37 to 12-digit M2M numbers. A row is selected by its
prefix, `phone-mobile-prefix[34 2]`.

A number's `block` is one of those prefixes and its `subscriber` the last five digits, such as
`5 12 34`. A number is valid, not real: a real subscriber may hold it. It lies in a range the plan
allocates, and the plan says "mainly" because numbers move between fixed and mobile networks.
Neither the order nor the authority's guide to it lists the numbers assigned to providers. A bound
number reads `national`, `41 23 45 67`, and `international`, `+45 41 23 45 67`.

`da_DK.phone-mobile` draws a mobile number and `da_DK.phone-fixed` a fixed one, each in the national form.

## Danish words

`da_DK.word` draws evenly from about 60,000 words of COR, Dansk Sprognævn's central word register of
Retskrivningsordbogen: the nouns, verbs, adjectives and adverbs it spells in lower-case letters alone,
in the forms it marks as the norm. COR has no frequencies, so a word may be rare. `da_DK.word.noun`,
`.verb`, `.adj` and `.adv` each draw one part of speech, and `da_DK.word.noun.definite`,
`da_DK.word.verb.present` and `da_DK.word.adj.definite` print one form, such as `huset`, `løber` or
`store`. Bind a word to read it and its form from one pick: `{/da_DK.word.noun as n}{n} {n.definite}`
prints `hus huset`.

The import drops the words on LDNOOBW's Danish list of rude words, and other rude words remain, so
read the output before showing it to anyone. `da_DK.sentence` fills a fixed pattern with these words
and doesn't make them agree, so its sentences aren't grammatical. `da_DK.paragraph` joins three to six sentences.
