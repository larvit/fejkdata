# `lv_LV`

`lv_LV`, Latvian, carries `color`, `company`, `company-coined`, `company-word`, `date`, `email`, `first-name`, `last-name`, `month`, `paragraph`, `person`, `personas-kods`, `phone`, `price`, `sentence`, `sex`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. `sex` is `sieviete` and `vīrietis`, `title` is
`Dr.` and `prof.`, drawn whatever the sex. `{/lv_LV.person}` renders e.g. `Anna Bērziņa`.

`first-name` holds the 2,000 commonest first given names of each sex, each weighted by the persons bearing it in the
Natural Persons Register, from the Office of Citizenship and Migration Affairs' "Personu vārdi" of 1 July 2026 (CC0 1.0). A
program showing the first names credits "Avots: Pilsonības un migrācijas lietu pārvalde". The register holds foreigners' names
too, so some names are not Latvian.
The office's open data holds first names only, so `last-name` is a language model's recall of names common in Latvia: the commoner ones
are ranked, each weighing 100000 over the square root of its rank, and the rarer ones share one weight. These weights
are estimates. `last-name` links to `sex` so that a woman bears the feminine form of a last name, such as `Bērziņa` for
`Bērziņš` and `Balode` for `Balodis`; a name such as `Liepa` is the same for both. `lv_LV.sex` carries Latvia's
population on 1 January 2025 from Eurostat's table `demo_pjan`. A program showing the sex counts credits "Source:
Eurostat".

`lv_LV.personas-kods` is a Latvian personal code, the personas kods that the Office of Citizenship and Migration
Affairs gives each person: 11 digits with no hyphen, such as `16117519997`, though forms often write one after the
sixth digit. It passes the check digit but is not
synthetic: the Office publishes no test series and no unassigned range, so a code may be a living person's. A code
issued before 1 July 2017 starts with the date of birth as `DDMMYY`, then `0` for the 1800s, `1` for the 1900s or `2` for
the 2000s, a serial number of three digits and a check digit. From 1 July 2017 the Population Register Act gives each new
person a code that starts with `32`, followed by digits drawn at random, with no date. The Office doesn't state the check
digit. fejkdata computes it as python-stdnum does, which says no independent source confirms it, and the State Revenue
Service's formula as Wikipedia gives it yields the same digit. For both kinds it is 1 plus each of the first ten digits
times its weight, 10, 5, 8, 4, 2, 1, 6, 3, 7 and 9, added up, modulo 11, with 10 written as 0. A checker that reads a date
from the first six digits refuses a code that starts with `32`.

fejkdata draws the year of birth by Eurostat's counts of people under 18, 18 to 64 and 65 to 99 on 1 January 2025
(table `demo_pjan`) and the serial number evenly from 000 to 999, since no source states which the register issues. A
person born from 1 July 2017 gets a code that starts with `32`, eight digits drawn evenly and the check digit: some
9% of the codes drawn in 2026, and a larger share each year until 2035. A person registered from 1 July 2017 but born
earlier, such as an immigrant, also holds a `32` code in the register, which fejkdata doesn't draw. A fixture of a few thousand rows starts to repeat a code, and a `UNIQUE` column needs a category of your own. A
program showing the birth years credits "Source: Eurostat".

`{personaskods(d)}` prints a whole code for the `YYYYMMDD` date that `d` reads, since the date decides the century digit
and whether the code starts with `32`. It prints no code where `d` isn't a date from 1800 to 2099, and `d` is a field, path
or name, never a call. The template below binds the `birth` field of `lv_LV.personas-kods` as `b`, prints it and prints
the code of the same date:

```json
{ "format": "{name}, {code}",
  "name": "{/lv_LV.person}",
  "code": "{/lv_LV.personas-kods.birth as b}born {b}, code {personaskods(b)}" }
```

Renders e.g. `Renārs Kārkliņš, born 20130310, code 10031324059`.

## Latvian companies

`lv_LV.company` draws a legal form as often as the Register of Enterprises' open data counts its merchants. It counts
the merchants of the Commercial Register that had no termination date on 19 November 2026: 140,485 limited liability
companies (`SIA`), 12,146 individual merchants (`IK`), 765 joint-stock companies (`AS`), 535 general partnerships
(`PS`) and 166 limited partnerships (`KS`). The branches, the European companies and the other registers' entries, such
as the associations and the farms, aren't drawn.

Most of the register's names put the form before the name and quote the name, such as `SIA "Alfa"`. The others put the
form before the name unquoted, or after it, quoted or not. `lv_LV.company` places the abbreviation and the quotes as
often as the names of the form do. A name that spells the form out, such as `Sabiedrība ar ierobežotu atbildību
"Alfa"`, counts as a name with the abbreviation, and `lv_LV.company` never spells a form out. `lv_LV.company.sia`,
`.ik`, `.as`, `.ps` and `.ks` each draw one form, and `{/lv_LV.company as c}{c.legal}` prints the abbreviation of the
same pick.

The shapes of the names, the trade words of `company-word` and the words of `company-coined` come from the names of
those merchants, as [`company-lv.py`](../../data-import/company-lv.py) counts them. Many of the words are English, since many
of the register's names are. A name is invented from those parts, and may also be a real company's. The register's open data
is CC0 1.0, so a program showing it needs no credit; `DATA-LICENSES.md` names the source.

## Latvian dates and emails

`lv_LV.date` is a date from 1970 to 2029 as Latvia writes it, a two-digit day and month and a four-digit year with
dots between, such as `02.04.2019`. `lv_LV.email` is built like `de_DE.email`: half its local parts are the first and
last name of one person from `lv_LV.person`, folded to ASCII, such as `janis.berzins7`. Its domains are `example.com`,
`example.org` and `example.net`, or a subdomain of one, which IANA reserves, so no address reaches a real mailbox; its
`reserved` path draws the same.

## Latvian phone numbers

`lv_LV.phone` is a mobile number ten times as often as a fixed one. Latvia has no area codes, so every number is
dialled as it stands, in eight digits grouped two, three and three: `29 123 456` for a mobile and `67 123 456` for a
fixed number. The `national` field of `lv_LV.phone` is that form, and its `international` field is
`+371 29 123 456`, with the country code.

A mobile number starts with 2 and a fixed number with 6, and each of the other seven digits is drawn evenly. These are
the ranges of Table 1 of the National Numbering Plan (Nacionālais numerācijas plāns), decision no. 1/33 of
27 September 2022 of the Public Utilities Commission (Sabiedrisko pakalpojumu regulēšanas komisija), in the text
amended to 1 September 2025.

A number is valid, not real: it lies in a range the plan allocates, and a real subscriber may hold it. The plan doesn't
reserve any range for fiction or tests, so don't call or text a drawn number. The plan doesn't split either range
further, but Google's libphonenumber refuses some blocks of the mobile range 23: it rejects about one mobile number
in a hundred that fejkdata draws.

## Latvian words

`lv_LV.word` draws evenly from about 88,000 words of Tēzaurs.lv: the nouns, verbs, adjectives and adverbs whose lemma is
lower-case letters alone and that carry no label calling them regional, a dialectism, obsolete, historical, outdated or an old
rarely used word. A
verb is an infinitive, ending in `-t` or `-ties`. Tēzaurs.lv has no frequencies, so a word may be rare.
`lv_LV.word.noun`, `.verb`, `.adj` and `.adv` each draw one part of speech, and `lv_LV.word.noun.plural` draws a noun and
prints its nominative plural, such as `mājas`. A noun with no plural in lower-case letters alone is left out. Bind a noun
to read it and its plural from one pick: `{/lv_LV.word.noun as n}{n} {n.plural}` prints `iedaliena iedalienas`.

Rude words stay in the tables, so read the output before showing it to anyone. `lv_LV.sentence` builds on plural nouns,
infinitives and adverbs, and its adjectives follow `tas ir`, so it needs no gender and no agreement:
`Tad vingrojumi sāka glāsaini nobraukāt.` and `Vai skalbes var ieteicami atdalīt?` Its words are drawn at random, so a
sentence may be senseless. `lv_LV.paragraph` joins three to six sentences.

A program showing these words credits "© Artificial Intelligence Laboratory, IMCS, University of Latvia, 2009–2026",
names the licence, [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), links to [tezaurs.lv](https://tezaurs.lv),
says that the words were changed from Tēzaurs.lv's entries, and shares any change to them under the same licence.
