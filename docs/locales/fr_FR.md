# `fr_FR`

`fr_FR`, French, carries `address`, `color`, `company`, `company-coined`, `company-word`, `date`, `department`, `email`, `first-name`, `landline`, `last-name`, `mobile`, `month`, `nir`, `nonzero-number`, `paragraph`, `person`, `phone`, `phone-fixed`, `phone-mobile`, `price`, `sentence`, `sex`, `sex-digit`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. `sex` is
`femme` and `homme`, and `title` is `Dr` and `Pr`, drawn whatever the sex. `{/fr_FR.person}` renders e.g. `Danielle Faucher`.
`department` prints the code INSEE gives a department in a NIR, such as `75`, `2A` or `974`, and `nonzero-number` a number
from 1 up, written with leading zeros: `two` from `01` to `99` and `three` from `001` to `999`. Both exist for `nir`.

`fr_FR.first-name` holds each first name given to at least 5 children of a sex in those births and `fr_FR.last-name` each last
name borne by at least 5 people, from INSEE's open files. A first name weighs the children who got it as their first one in the births of 1944
to 2023, the last 80 years, in the "Fichier des prénoms"; a name gone out of use weighs less than in the whole file,
and the children of a year in which fewer than 3 got the name are left out, so a count runs slightly low. A last name
weighs the people born in 1891 to 2000 who bear it, in the "Fichier des noms". INSEE writes last names in capitals
without accents, so `Lefevre` also counts the bearers of `Lefèvre`. `fr_FR.sex` carries France's population on 1 January 2025
from Eurostat's table `demo_pjan`; Eurostat marks that figure provisional.

A program showing the name counts credits "Source : Insee, Fichier des prénoms, 08/07/2024" and "Source : Insee,
Fichier des noms de famille, 22/05/2018", one showing the sex counts "Source: Eurostat", and one showing the birth
years of a NIR "Source : Insee, Bilan démographique 2025".

## NIR

`fr_FR.nir` is a French NIR, the numéro d'inscription au répertoire that INSEE gives each person and that is printed as
the numéro de sécurité sociale: 15 characters with no spaces, such as `271121647338041`. It passes the key check but is
not synthetic: INSEE publishes no test series and no unassigned range, so a number may be a living person's. The 13
characters before the key are the sex (`1` for a man, `2` for a woman), the year and month of birth, the department and
commune of birth (5 characters), and the order number, `001` to `999`. The key is 97 less the remainder of those 13
divided by 97, with Corsica's `2A` and `2B` counted as 19 and 18.

fejkdata draws the year of birth by INSEE's counts of people under 18, 18 to 64 and 65 to 99 on 1 January 2026
("Population au 1er janvier par sexe et âge détaillé", Bilan démographique 2025). The department is drawn evenly from
INSEE's 101: 01 to 19, 2A, 2B, 21 to 95, and the overseas 971 to 974 and 976. The commune and the order number are drawn
evenly, from 001 to 999 (01 to 99 for an overseas commune). The commune is a code in its department's range, not one
that exists. Corsica is always `2A` or `2B`, though a birth there before 1976 is under `20`. A person born abroad, whose
department is `99`, and the first digits 3, 4, 7 and 8 are not drawn. A fixture of a few hundred thousand rows starts
to repeat a number, so a `UNIQUE` column needs a category of your own.

`{nir()}` completes a NIR: it reads the last 13 letters and digits emitted and prints the key. To make a number agree
with a sex the template binds as `s`, write `fr_FR.sex-digit`'s digit under `s`, then the `body` field of `fr_FR.nir`
(everything between the sex and the key), then `{nir()}`:

```json
{ "format": "{first} {last}, {nir}",
  "first": "{/fr_FR.sex as s}{s.first-name}",
  "last": "{/fr_FR.last-name}",
  "nir": "{s.sex-digit.digit}{/fr_FR.nir.body}{nir()}" }
```

Renders e.g. `Cassie Duflot, 253059724702338`.

## French companies

`fr_FR.company` draws a legal form as often as INSEE's Sirene register counts active legal units of it on 1 November
2026: 7,845,810 sole proprietors (`EI`), 2,226,407 `SARL`, 1,839,229 `SAS`, 99,237 `SNC` and 68,436 `SA`, the last being
Sirene's two categories of a société anonyme with no other indication, added together. Sirene files an `EURL` under
its `SARL` and a `SASU` under its `SAS`, so the data doesn't tell them apart. The `SCI`, the associations and the
other forms the register holds aren't drawn. Sirene counts a unit active until someone reports it ceased, so the
counts run higher than the businesses trading.

An `EI`'s name is a first and a last name, such as `Nadia Cordier`, with no abbreviation after it, as the register
holds it, although a sole proprietor must add `EI` to their name on their documents since 2022. The others end in their
abbreviation, such as `… SARL`. `fr_FR.company.ei`, `.sarl`, `.sas`, `.snc` and `.sa` each draw one form, and
`{/fr_FR.company as c}{c.legal}` prints the abbreviation of the same pick, empty for an `EI`.

The shapes of the other names, the 120 commonest trade words of `company-word` and the roots of `company-coined`
come from the names of those units, as [`company-fr.py`](../../data-import/company-fr.py) counts them. Sirene writes a name
in capitals without accents, so the trade words and the commoner roots have the accents of French spelling, written by hand, and any other root
keeps Sirene's spelling, such as `Acces`. A name is invented from those parts, and may also be a real company's.
A program showing the data credits "Source : Insee, Sirene, StockUniteLegale, 1er novembre 2026", names the
[Licence Ouverte / Open Licence 2.0](https://www.etalab.gouv.fr/licence-ouverte-open-licence/) and says that the
data is reduced to five legal forms and derived from the register's names.

## French dates and emails

`fr_FR.date` is a date from 1970 to 2029 as France writes it, a two-digit day and month and a four-digit year with
slashes between, such as `02/04/2019`. `fr_FR.email` is built like `de_DE.email`: half its local parts are a first and
a last name from `fr_FR`'s names, folded to ASCII, with `-` for a space or an apostrophe, such as `anne-marie.le-gall`.
Its domains are `example.com`, `example.org` and `example.net`, or a subdomain of one, which IANA reserves, so no
address reaches a real mailbox.

## French phone numbers

`fr_FR.phone` is a mobile number ten times as often as a fixed one, such as `06 39 98 12 34`, or
`01 99 00 12 34` where it is fixed. Its `national` field is that form, ten digits in pairs as France writes
them. Its `international` field is `+33 6 39 98 12 34`, with the country code and without the 0 that opens
the national form.

Every number is invented and reaches no one. The Arcep's plan national de numérotation, in the version of
1 January 2026, lists six roots, the first six digits of a number, under "Numéros pour œuvres
audiovisuelles": it doesn't allot them to anyone, and says they may be used as telephone numbers in fiction.
`fr_FR.phone-fixed` holds the five fixed roots, `01 99 00`, `02 61 91`, `03 53 01`, `04 65 71` and
`05 36 49`, and `fr_FR.phone-mobile` the mobile one, `06 39 98`. Each is a row's `area`, with its `ndc`, the
root without the 0, and its `subscriber`, the last four digits. A row is selected by its `area`,
`phone-fixed[01 99 00]`, and `{fr_FR.phone-fixed.area}` prints one.
The tables hold 60,000 numbers, and the 10,000 mobile ones are ten draws in eleven, so a fixture of a few
hundred rows repeats a number and a `UNIQUE` column needs a category of your own. No
checker is promised to accept a number: the roots are no operator's.

`fr_FR.mobile` draws a mobile number and `fr_FR.landline` a fixed one, each in the national form. A bound number reads `national` and `international`.

## French words

`fr_FR.word` draws evenly from about 64,000 words of the English Wiktionary's French entries, as kaikki.org
extracts them: the nouns, verbs, adjectives and adverbs in lower-case letters alone that Wiktionary lists as words of
their own, not as an inflected form or a variant of another entry, so a verb is an infinitive, never a conjugated form.
Wiktionary has no frequencies, so a word may be rare. `fr_FR.word.noun`, `.verb`, `.adj` and `.adv` each draw one part
of speech, and `fr_FR.word.noun.plural` draws a noun and prints its plural, such as `chevaux`: the first plural
Wiktionary lists, or the noun itself where it is invariable, as `pays`. A noun with no plural in lower-case letters
alone is left out. Bind a noun to read it and its plural from one pick: `{/fr_FR.word.noun as n}{n} {n.plural}` prints
`cheval chevaux`.

Rude words stay in the tables, so read the output before showing it to anyone. `fr_FR.sentence` builds on plural
nouns, infinitives and adverbs, so it needs no gender and no agreement: `Il y a des données et des garennes.` and
`Voit-on souvent des dirigeables ?` Its words are drawn at random, so a sentence may be senseless or even
ungrammatical, such as an intransitive verb given an object. `fr_FR.paragraph` joins three to six sentences.
A program showing these words credits "Wiktionary contributors, CC BY-SA 4.0" with links to
[the licence](https://creativecommons.org/licenses/by-sa/4.0/) and to [en.wiktionary.org](https://en.wiktionary.org/),
says that the words were changed from Wiktionary's entries, and shares any change to them under the same licence.
