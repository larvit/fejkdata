# `et_EE`

`et_EE`, Estonian, carries `color`, `company`, `company-coined`, `company-word`, `date`, `email`, `first-name`, `isikukood`, `last-name`, `month`, `paragraph`, `person`, `phone`, `price`, `sentence`, `sex`, `sex-digit`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. `sex` is `naine` and `mees`, `title` is
`dr` and `prof`, drawn whatever the sex. `{/et_EE.person}` renders e.g. `Liis Tamm`.

Both name tables are a language model's recall of names common in Estonia: the commoner ones of each list are
ranked, each weighing 100000 over the square root of its rank, and the rarer ones share one weight. These weights
are estimates. The lists include the Russian-language names many Estonians bear, and `last-name` links to `sex` so
that a woman bears the feminine form of one: `Ivanova` where a man bears `Ivanov`. `et_EE.sex` carries Estonia's
population on 1 January 2025 from Eurostat's table `demo_pjan`. A program showing the sex counts credits "Source:
Eurostat".

`et_EE.isikukood` is an Estonian personal identification code, the isikukood that the Population Register gives each
person: 11 digits with no spaces, such as `36805280109`. It passes the check digit but is not synthetic: the Interior
Minister's regulation "Isikukoodide moodustamise ja andmise kord" names no test series and no unassigned range, so a
code may be a living person's. Its digits are the sex and the century of birth (`1` for a man and `2` for a woman born
in the 1800s, `3` and `4` in the 1900s, `5` and `6` in the 2000s), the date of birth as `YYMMDD`, a serial number and a
check digit. The regulation leaves the check digit to the standard EVS 585, which is not public, so fejkdata computes
it as python-stdnum does. The ten digits are summed with the weights 1 to 9 and 1, and the check digit is the sum
modulo 11. Where that is 10, the weights are 3 to 9, 1, 2 and 3 instead, and where that is 10 again, the check digit is 0.

fejkdata draws the year of birth by Statistics Estonia's counts of people under 18, 18 to 64 and 65 to 99 on 1 January
2026 (table RV0240) and the serial number evenly from 001 to 999. No source states which serial numbers are issued, so
000, the one number outside the range, is left out. A fixture of some ten thousand rows starts to repeat a code, and a
`UNIQUE` column needs a category of your own. A program showing the birth years credits "Source: Statistics Estonia".

`{isikukood(x,d)}` prints a whole code, since its first digit depends on the date that follows. The template below binds
the sex as `s` and the `birth` field of `et_EE.isikukood`, a day written `YYYYMMDD`, as `b`. The call's operands are the digit of
`et_EE.sex-digit` under `s`, and `b`:

```json
{ "format": "{first} {last}, {code}",
  "first": "{/et_EE.sex as s}{s.first-name}",
  "last": "{s.last-name}",
  "code": "{/et_EE.isikukood.birth as b}{isikukood(s.sex-digit.digit,b)}" }
```

Renders e.g. `Lars Pavlov, 39406231528`.

## Estonian companies

`et_EE.company` draws a legal form as often as the e-Business Register counts the businesses of that form entered in the
register on 19 November 2026: 288,168 private limited companies (`OÜ`), 24,541 sole proprietors (`FIE`), 3,514 limited
partnerships (`UÜ`), 1,901 public limited companies (`AS`) and 1,028 general partnerships (`TÜ`). Those in liquidation
or bankruptcy aren't counted. The non-profits, the apartment associations, the co-operatives and the other forms the
register holds aren't drawn.

An `FIE`'s name is one person's first and last name, such as `Kristo Stepanov`, with no abbreviation, as most of the
register's `FIE` names are. The names of the other forms carry their abbreviation after the name, such as `Saare
Technologies OÜ`, or before it, such as `AS Audio Restoran`, as often as the register's names of that form do. Most `OÜ`
and `UÜ` names end in their abbreviation, and most `AS` and `TÜ` names open with it. `et_EE.company.fie`, `.ou`, `.as`,
`.uu` and `.tu` each draw one form, and `{/et_EE.company as c}{c.legal}` prints the abbreviation of the same pick, empty
for an `FIE`.

The shapes of the other names, the trade words of `company-word` and the words of `company-coined` come from the names
of those businesses, as [`company-ee.py`](../../data-import/company-ee.py) counts them. Many of the words are English, as the
register's names are. A name is invented from those parts, and may also be a real company's. A program showing the
data credits "Source: e-Business Register, Centre of Registers and Information Systems (RIK)", names
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) and says that the data is reduced to five legal forms and
derived from the register's names.

## Estonian dates and emails

`et_EE.date` is a date from 1970 to 2029 as Estonia writes it, a two-digit day and month and a four-digit year with
dots between, such as `02.04.2019`. `et_EE.email` is built like `de_DE.email`: half its local parts are the first and
last name of one person from `et_EE.person`, folded to ASCII, such as `martin.koiv7`. Its domains are `example.com`,
`example.org` and `example.net`, or a subdomain of one, which IANA reserves, so no address reaches a real mailbox.

## Estonian phone numbers

`et_EE.phone` is a mobile number ten times as often as a fixed one. Estonia has no area codes, so every number
is dialled as it stands: `5123 4567` for a mobile and `612 3456` for a fixed number. Its `national` field is that
form, and its `international` field is `+372 5123 4567`, with the country code.

The ranges are those of the Annex to the Estonian numbering plan, regulation no. 25 of 21 May 2018 of the
Minister of Entrepreneurship and Information Technology, in the text of the ministry's draft amending it in
February 2024. A mobile number has eight digits and starts with 5. A fixed number has seven digits and starts
with 32, 33, 35, 38, 39, 43 to 48, 60 to 68, 71 to 79 or 88, each of those 30 starts drawn evenly, since the
Annex publishes no frequency. Not drawn are the seven-digit mobile numbers the Annex allows as exceptions, its
other mobile range 81 to 84, all of 69, where 690 to 696 serve nationwide state communications, and every
number for another service, such as 900 for premium rates.

A number is valid, not real: it lies in a range the Annex allocates, and a real subscriber may hold it. The
Annex doesn't reserve any range for fiction or tests, so don't call or text a drawn number.

## Estonian words

`et_EE.word` draws evenly from about 5,500 words of the English Wiktionary's Estonian entries, as kaikki.org extracts
them: the nouns, verbs, adjectives and adverbs in lower-case letters alone that Wiktionary lists as words of their own,
not as an inflected form or a variant of another entry, so a verb is its `-ma` infinitive and never a conjugated form.
Wiktionary has no frequencies, so a word may be rare, and the tables are small beside the other locales' because no
larger Estonian word list is open without an account. `et_EE.word.noun`, `.verb`, `.adj` and `.adv` each draw one part
of speech, and `et_EE.word.noun.plural` draws a noun and prints its nominative plural, such as `majad`. A noun with no
nominative plural in lower-case letters alone is left out. Bind a noun to read it and its plural from one pick:
`{/et_EE.word.noun as n}{n} {n.plural}` prints `konn konnad`.

Rude words stay in the tables, so read the output before showing it to anyone. `et_EE.sentence` builds on singular
nouns beside `on`, plural nouns and `-ma` infinitives after `hakkama`, so it needs no case ending the tables lack: `Kas see
karjalane on valge?` and `Siis hakkasid lennujaamad niimoodi eelistama.` Its words are drawn at random, so a sentence may
be senseless or even ungrammatical, such as an adjective that doesn't suit its noun. `et_EE.paragraph` joins three to six
sentences.
