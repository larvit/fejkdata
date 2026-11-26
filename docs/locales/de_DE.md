# `de_DE`

`de_DE`, German, carries `address`, `color`, `company`, `company-coined`, `company-word`, `date`, `email`, `first-name`, `landline`, `last-name`, `mobile`, `month`, `paragraph`, `person`, `phone`, `phone-fixed`, `phone-mobile`, `phone-subscriber`, `price`, `sentence`, `sex`, `steuer-id`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. `sex` is `Frau` and `Mann`, `title` is
`Dr.`, `Prof.` and `Prof. Dr.`, drawn whatever the sex. `{/de_DE.person}` renders e.g. `Emilia Hesse`.

No open register publishes Germany's name counts. The first names are the city of Munich's open data
["Vornamen von Neugeborenen"](https://opendata.muenchen.de/dataset/vornamen-von-neugeborenen): the children
of each sex born to people living in Munich, 2013 to 2025, each name weighing the sum of the counts the files
publish for it. So the names are the newborns' of one city since 2013, not those of Germany's adults. A file
gives a count below 5 (2025: below 6) as a phrase, not a number, so a name never published as a number is
missing. The last names are a language model's recall of names common in Germany: the commoner ones are
ranked, each weighing 700000 over the square root of its rank, and the rarer ones share one weight. These weights
are estimates. `de_DE.sex` carries Germany's population on 1 January 2025 from Eurostat's table `demo_pjan`.

A program showing the first names credits "Datenquelle: Landeshauptstadt München, Statistisches Amt", names the
licence [Datenlizenz Deutschland – Namensnennung 2.0](https://www.govdata.de/dl-de/by-2-0) (`dl-de/by-2-0`), links
the [dataset](https://opendata.muenchen.de/dataset/vornamen-von-neugeborenen), and says the data was changed: the
files are summed per name and sex. One showing the sex counts credits "Source: Eurostat".

## Steuer-ID

`de_DE.steuer-id` is a German tax identification number (Steueridentifikationsnummer), such as
`02476291358`, from ELSTER's test numbers. ELSTER is the tax authorities' online portal. Every number starts
with 0. The Federal Central Tax Office (BZSt) doesn't issue any number with a leading zero, and ELSTER's
"Prüfung der Steuer- und Steueridentifikationsnummer" accepts a leading zero only in a test number. So a
number here is valid and no person holds it, but a checker that refuses a leading zero refuses every one, and
fejkdata draws none without it.

A number has eleven digits. Among the first ten, one digit appears twice in half the numbers and three times
in the other half. No digit appears three times in a row. The eleventh digit is the check digit of ISO 7064
MOD 11,10. The category renders the builtin `{steuerid()}`. There are about 37 million numbers, so a fixture
of a few thousand rows starts to repeat one, and a `UNIQUE` column needs a category of your own.

## German companies

`de_DE.company` draws a sole proprietorship (Einzelunternehmen), a partnership or a corporation as often as Destatis
counted legal units of that kind in 2024 (its Unternehmensregister table of legal units by legal form). The table counts
partnerships and corporations only as groups, so a partnership is a `KG` or an `OHG` and a corporation a `GmbH`, a
`UG (haftungsbeschränkt)` or an `AG`, the forms of a group drawn evenly. A sole proprietorship's name has no
abbreviation after it, and the others end in theirs, such as `… GmbH`. `de_DE.company.einzel`, `.kg`, `.ohg`, `.gmbh`, `.ug` and
`.ag` each draw one form, and `{/de_DE.company as c}{c.legal}` prints the abbreviation of the same pick. No open
bulk file of German company names was found, so the name's shapes and the words of `company-coined` and
`company-word` are a language model's recall. Each name is made of those words, `de_DE`'s names and `geo.DE.locality`.
A name is invented, and may also be a real company's. A program showing the counts credits "© Statistisches
Bundesamt (Destatis), 2025", names the licence [Datenlizenz Deutschland – Namensnennung 2.0](https://www.govdata.de/dl-de/by-2-0)
(`dl-de/by-2-0`), links it, and says that the groups are split and the counts reduced to six legal forms.

## German dates and emails

`de_DE.date` is a date from 1970 to 2029 in the German spelling of DIN 5008, a two-digit day and month and a four-digit year
with dots between, such as `02.04.2019`. `de_DE.email` is built like `nb_NO.email`: half its local parts are a first and a
last name from `de_DE`'s names, folded to ASCII, `ä` to `a`, `ö` to `o`, `ü` to `u` and `ß` to `ss`.

## German phone numbers

`de_DE.phone` is a mobile number ten times as often as a fixed one, such as `01511 2345678` or
`030 12345678`. Its `national` field is that spelling. Its `international` field is `+49 1511 2345678`:
the country code, then the number without the 0 dialled before an area code. One space splits the area
code or mobile block from the subscriber number, which is not grouped.

`de_DE.phone-fixed` holds the 5200 Ortsnetzkennzahlen (area codes) that the Bundesnetzagentur's
Vorwahlverzeichnis of 2022-07-27 marks active. A row has the code as `area`, such as `030`, the Ortsnetz
as `name`, such as Berlin, the code without its 0 as `ndc`, and the `subscriber` number drawn after it.
`de_DE.phone-mobile` holds the 81 blocks, such as `01511` and `015550`, that the Bundesnetzagentur's list
of assigned mobile blocks names, with the columns `area`, `ndc` and `subscriber`.
`de_DE.phone-subscriber` draws a subscriber number of six to eight digits as `digits-6` to `digits-8`. A
row is selected by its area, `phone-fixed[030]`. A code is drawn evenly, since the lists publish no
frequency, so a village is as likely as Berlin.

A number allocated now has ten digits after the 0 in Berlin, Hamburg, Frankfurt and München and eleven
elsewhere, as section 2.3 of Verfügung 25/2006 gives. A mobile number has eleven digits after the 0, or
ten or eleven for a block of three digits such as `170`, as sections 2.3 and 2.4 of the Nummernplan
Mobile Dienste give; both lengths are equally likely. A subscriber number never begins with 0 or 1,
since the Verfügung's checks refuse a number that begins with 110, 112 or 115. A number is valid, and a
real subscriber may hold it.

A program showing the codes or blocks credits "Datenquelle: Bundesnetzagentur", names the licence
[Datenlizenz Deutschland – Namensnennung 2.0](https://www.govdata.de/dl-de/by-2-0) (`dl-de/by-2-0`), links
the [mobile blocks](https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/MobileDienste/RNB_csv.csv)
and the [Vorwahlverzeichnis](https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/ONRufnr/_DL/Vorwahlverzeichnis_ONB.zip.zip),
and says the data was changed: each block and code is a row.

`de_DE.mobile` draws a mobile number and `de_DE.landline` a fixed one, each in the national form. A bound number reads `national` and `international`.

## German words

`de_DE.word` draws evenly from about 205,000 words of Wikidata's German lexemes: the nouns, verbs,
adjectives and adverbs whose lemma is letters alone, a noun's capitalised and the others' lower-case.
Wikidata has no frequencies, so a word may be rare. `de_DE.word.noun`, `.verb`, `.adj` and `.adv` each
draw one part of speech, and `de_DE.word.noun.plural` draws a noun and prints its nominative plural, such as
`Häuser`. A noun is left out where Wikidata gives it no plural of letters alone. Bind a noun to read it and its plural from one
pick: `{/de_DE.word.noun as n}{n} {n.plural}` prints `Haus Häuser`.

The import drops the words on LDNOOBW's German list of rude words, and other rude words remain, so
read the output before showing it to anyone. `de_DE.sentence` builds on plurals, so it needs no
gender: `Die Häuser sind groß.` A verb stands whole in its infinitive, which reads as its third person plural
for most verbs. A verb such as `anrufen` or `sein` then reads wrong, so some sentences aren't grammatical. `de_DE.paragraph`
joins three to six sentences.
