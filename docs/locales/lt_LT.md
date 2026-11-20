# `lt_LT`

`lt_LT`, Lithuanian, carries `address`, `asmens-kodas`, `color`, `company`, `company-coined`, `company-word`, `date`, `email`, `first-name`, `last-name`, `month`, `paragraph`, `person`, `phone`, `phone-fixed`, `phone-mobile`, `price`, `sentence`, `sex`, `sex-digit`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. `sex` is `moteris` and `vyras`, `title` is
`dr.` and `prof.`, drawn whatever the sex. `{/lt_LT.person}` renders e.g. `Ona Kazlauskienė`.

`first-name` holds, with the counts the State Enterprise Centre of Registers published for the people born since 1900 in the
Population Register (CC BY 4.0), the 50 commonest men's names and women's names, and more names without a published count. A program
showing the first names credits "Valstybės įmonė Registrų centras". The register holds foreigners' names too, so some names, such
as `Ivan`, are not Lithuanian. The Centre publishes no count of a name below its top 50, and no open file of Lithuanian last names
was found, so every other first name and every last name is a language model's recall of a name common in Lithuania: the commoner
ones are ranked, a first name weighing the smallest published count times 50 over its rank, counted on from 51, and a last name
100000 over the square root of its rank, and the rarer ones share one weight. These weights are estimates. `last-name` links to
`sex` so that a woman bears a feminine form of a last name. A Lithuanian name ships two, under `f`: the unmarried, such as
`Kazlauskaitė` for `Kazlauskas`, and the married, `Kazlauskienė`. They share the masculine name's weight. A Russian or Polish name
takes its own language's ending, `Ivanova` for `Ivanov`.
`lt_LT.sex` carries Lithuania's population on 1 January 2025 from Eurostat's table `demo_pjan`. A program showing the sex
counts credits "Source: Eurostat".

`lt_LT.asmens-kodas` is a Lithuanian personal identification code, the asmens kodas that the Population Register gives
each person: 11 digits with no spaces, such as `33309240064`. It passes the check digit but is not synthetic: the
Register publishes no test series and no unassigned range, so a code may be a living person's. Its digits are the sex and the century of birth (`1` for a man and `2` for a woman born in the 1800s, `3`
and `4` in the 1900s, `5` and `6` in the 2000s), the date of birth as `YYMMDD`, a serial number and a check digit. The standard
that sets this structure, LST 1185-91, is not public, so fejkdata computes the check digit as python-stdnum's `lt.asmens` does,
which takes it from python-stdnum's Estonian check: the ten digits are summed with the weights 1 to 9 and 1, and the check digit is the sum modulo 11. Where that is 10, the weights
are 3 to 9, 1, 2 and 3 instead, and where that is 10 again, the check digit is 0.

fejkdata draws the year of birth by Eurostat's counts of people under 18, 18 to 64 and 65 to 99 on 1 January 2025 (table
`demo_pjan`) and the serial number evenly from 001 to 999. No source states which serial numbers are issued, and the Lithuanian Wikipedia calls it the order
number of those born that day, so 000 is left out. A fixture of some ten thousand rows starts to repeat a code, and a `UNIQUE` column needs a
category of your own. fejkdata draws no code with a month or day of 00, which the Lithuanian Wikipedia says a person who
doesn't know the birth date has, and none that starts with 9, which it says has no date. A program showing the birth years credits "Source: Eurostat".

`{asmenskodas(x,d)}` prints a whole code, since its first digit depends on the date that follows. The template below binds
the sex as `s` and the `birth` field of `lt_LT.asmens-kodas`, a day written `YYYYMMDD`, as `b`. The call's operands are the
digit of `lt_LT.sex-digit` under `s`, and `b`:

```json
{ "format": "{first} {last}, {code}",
  "first": "{/lt_LT.sex as s}{s.first-name}",
  "last": "{s.last-name}",
  "code": "{/lt_LT.asmens-kodas.birth as b}{asmenskodas(s.sex-digit.digit,b)}" }
```

Renders e.g. `Andrius Jurevičius, 36707114290`.

## Lithuanian companies

`lt_LT.company` draws a legal form as often as the Register of Legal Entities' open data counts its entities. The
file of 19 November 2026 lists 102,461 private limited companies (`UAB`), 75,713 small partnerships (`MB`), 452
agricultural companies (`ŽŪB`) and 226 public limited companies (`AB`). The other forms aren't drawn. The register
names an individual enterprise, a limited partnership and a general partnership by their form alone. The rest
include branches and representative offices, cooperatives, law partnerships, and non-profit or public bodies, such
as associations.

Most of the register's names put the form before the name, and most `UAB` quote the name, such as `UAB "Alfa"`. The
others put the form before the name unquoted, or after it, quoted or not. `lt_LT.company` places the abbreviation and
the quotes as often as the names of the form do. A name that spells the form out, such as `Uždaroji akcinė bendrovė
"Alfa"`, counts as a name with the abbreviation, and `lt_LT.company` never spells a form out. `lt_LT.company.uab`,
`.mb`, `.zub` and `.ab` each draw one form, and `{/lt_LT.company as c}{c.legal}` prints the abbreviation of the same
pick.

The shapes of the names, the trade words of `company-word` and the words of `company-coined` come from the names of
those entities, as [`company-lt.py`](../../data-import/company-lt.py) counts them. Many of the words are English, since
many of the register's names are. A name is invented from those parts, and may also be a real company's. The register's
open data is CC BY 4.0, so a program showing the data credits "Valstybės įmonė Registrų centras"; `DATA-LICENSES.md` has the
rest.

## Lithuanian dates and emails

`lt_LT.date` is a date from 1970 to 2029 in CLDR's short Lithuanian format, `y-MM-dd`: a four-digit year, then a
two-digit month and day, with hyphens between, such as `2019-04-02`. `lt_LT.email` is built like `de_DE.email`:
half its local parts are the first and last name of one person from `lt_LT.person`, folded to ASCII, such as
`jonas.kazlauskas7`. Its domains are `example.com`, `example.org` and `example.net`, or a subdomain of one, which IANA
reserves, so no address reaches a real mailbox; its `reserved` path draws the same.

## Lithuanian phone numbers

`lt_LT.phone` is a mobile number ten times as often as a fixed one: `0 612 34567`, or `0 5 234 5678`
where it is fixed in Vilnius. Its `national` field is that form, and its `international` field is
`+370 612 34567`, with the country code and without the 0 that opens the national form, which replaced
the old 8 on 1 March 2024. A number is valid, not real: a real subscriber may hold it, since Lithuania's
numbering plan keeps no range for fiction.

The plan is the one the Communications Regulatory Authority (RRT) adopted on 1 September 2023, which the
ITU published on 6 March 2025. A number has 8 digits after the 0, and a mobile number opens with 6.
`lt_LT.phone-fixed` holds the 47 area codes that the plan's table of fixed numbers allots to a city or
district and doesn't mark "Reserved", and `lt_LT.phone-mobile` the one mobile code, 6. Each row's `area` is
its code with the trunk 0, such as `037` for Kaunas, and its `ndc` is the code without the 0. Its
`subscriber` is the digits after the code, and a fixed row also has a `name`, the city or district. A row is selected by its `area`,
`phone-fixed[037]`, and each is drawn evenly, since the plan publishes no frequency. Where the plan fixes
the first digits of the subscriber numbers, `subscriber` draws only those: Vilnius's `5` is followed by
20 to 27, Panevėžys's `45` by 4 or 5, and Klaipėda's `46` by 2 to 4.

The digits are grouped by the length of the area code, `0 5 234 5678` for Vilnius's 1 digit, `0 37 12 3456`
for a 2-digit code such as Kaunas's, and `0 310 12345` for a district's 3; a mobile number groups as a
district. The plan fixes only the digits, and RRT's recommendation, as the Klaipėda city municipality's
page "Telefonų numerių rašymas" gives it, lets them be grouped as the writer likes, so the groups are this
project's choice.

## Lithuanian words

`lt_LT.word` draws evenly from about 5,100 words of the English Wiktionary's Lithuanian entries: the nouns, verbs, adjectives
and adverbs that are headwords in lower-case letters alone, not forms or variants of another entry. A verb is almost always an
infinitive, such as `bėgti`; an adjective may be a participle used as one, such as `sergantis`. Wiktionary has no
frequencies, so a word may be rare.
`lt_LT.word.noun`, `.verb`, `.adj` and `.adv` each draw one part of speech, and `lt_LT.word.noun.plural` draws a noun and
prints its plural, such as `mamos`. A noun whose entry gives no plural is left out. Bind a noun to read it and its plural
from one pick: `{/lt_LT.word.noun as n}{n} {n.plural}` prints `vyras vyrai`.

Rude words stay in the tables, so read the output before showing it to anyone. `lt_LT.sentence` builds on nouns in the
nominative, infinitives and adverbs, and quotes each adjective as a word, `Žodis „didelis“ yra būdvardis`, so no
word has to agree with another: `Čia yra barakai ir mamos.` and `Ar tai yra vardiklis?` Its words are drawn at random,
so a sentence may be senseless. `lt_LT.paragraph` joins three to six sentences.

A program showing these words credits "Wiktionary contributors", [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)
and [en.wiktionary.org](https://en.wiktionary.org/), says that the words were changed from Wiktionary's entries, and
shares any change to them under the same licence.
