# `fi_FI`

`fi_FI`, Finnish, carries `address`, `color`, `company`, `company-coined`, `company-word`, `date`, `email`, `first-name`, `henkilotunnus`, `henkilotunnus-issued`, `individ`, `landline`, `last-name`, `mobile`, `month`, `paragraph`, `person`, `phone`, `phone-fixed`, `phone-mobile`, `phone-subscriber`, `price`, `sentence`, `sex`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`, built like [`nb_NO`'s](nb_NO.md).
`fi_FI.sex` is `nainen` and `mies`, weighted by [Statistics Finland](https://stat.fi/en)'s population of 31 December
2025, and `fi_FI.title` is `tohtori` and `professori`. `fi_FI.first-name` holds each first given name of each sex borne by at least 5 citizens
and `fi_FI.last-name` each last name borne by at least 5, each weighted by the Finnish citizens bearing it in the name statistics of 4 August 2026 of the Digital and
Population Data Services Agency ([DVV](https://dvv.fi/en)). A program showing those counts credits "Lähde: Digi- ja
väestötietovirasto", and one showing the sex counts "Source: Statistics Finland".

## Finnish companies

`fi_FI.company` draws a legal form as often as Statistics Finland counted enterprises of it in 2024 (its table 13w6):
sole trader (a natural person), limited company (`Oy`), limited partnership (`Ky`), general partnership (`Ay`) and
co-operative (`osk`). A sole trader's name has no abbreviation after it, and the others end in theirs, such as
`Pohjanniemi Rakennus Oy`. `fi_FI.company.tmi`, `.oy`, `.ky`, `.ay` and `.osk` each draw one form, and
`{/fi_FI.company as c}{c.legal}` prints the abbreviation of the same pick, empty for a sole trader.
No open bulk file of Finnish company names was found, so the name's shapes and the words of `company-coined` and
`company-word` are a language model's recall. Each name is made of those words and `fi_FI`'s names. A name is
invented, and may also be a real company's. A program showing the counts credits "Source: Statistics Finland".

## Finnish dates and emails

`fi_FI.date` is a date from 1970 to 2029 in the Finnish spelling, day, month and year with dots and no zero before a
one-digit day or month, such as `2.4.2019`; the Institute for the Languages of Finland (Kotus), in "Ajanilmaukset: päiväys ja
vuosiluku" of its Kielitoimiston ohjepankki, says a number under ten needs no leading zero. `fi_FI.email` is built like
`nb_NO.email`: half its local parts are a first and a last name from `fi_FI`'s names, with `ä`, `ö` and `å` folded to ASCII.

## Finnish phone numbers

`fi_FI.phone` is a mobile number ten times as often as a fixed one: `040 234 5678`, or `09 234 5678`
where it is fixed. Its `national` field is that form, and its `international` field is
`+358 40 234 5678`, with the country code and without the 0 dialled before an area code. The
subscriber number is split in two groups near its middle, as Traficom writes its own number,
`+358 29 534 5000`; the regulation fixes only the digits.

`fi_FI.phone-fixed` holds the area codes of the telecommunications areas in Traficom's Regulation
32 V/2025 M, such as `09` for Uusimaa I, each with its `name`, its `ndc` (the code without the 0)
and its `subscriber`. `fi_FI.phone-mobile` holds the mobile codes whose row in Annex 1 says `No`
under Available ranges, such as `040` and `0451`. A row is selected by its code, `phone-fixed[09]`.
A fixed number has 7 to 9 digits after the 0 and a mobile number 8 to 10, as Table 1 of the
regulation gives, each length equally likely since the regulation publishes no frequency. A fixed
subscriber number begins with 2 to 8, the ranges section 11 reserves for it. A number is valid, not
real: a real subscriber may hold it. The tables leave out the mobile codes the plan assigns as five
digits, such as `04300`, among them the machine-to-machine codes `049`, and the `059` codes.

`fi_FI.mobile` draws a mobile number and `fi_FI.landline` a fixed one, each in the national form. A bound number reads `national` and `international`.

## Finnish words

`fi_FI.word` draws evenly from about 101,000 words of Nykysuomen sanalista, Kotus's list of the words
in Kielitoimiston sanakirja: the nouns, verbs, adjectives and adverbs it spells in lower-case letters
alone. A word the list files under two classes, such as a noun that is also an adjective, sits in both
tables. The list has no frequencies, so a word may be rare. `fi_FI.word.noun`, `.verb`, `.adj` and `.adv`
each draw one part of speech. Each word is in its base form, since the list gives no inflected forms.

The import drops the words on LDNOOBW's Finnish list of rude words, and other rude words remain, so
read the output before showing it to anyone. `fi_FI.sentence` fills a fixed pattern that needs only base
forms, such as `Tämä suuri talo on vanha.` A plural noun such as `housut`, or an adjective that only stands before
a noun, such as `viime`, then reads wrong, so some sentences aren't grammatical. `fi_FI.paragraph` joins three to six sentences.
A program showing these words credits "Kotimaisten kielten keskus: Nykysuomen sanalista 2024, CC BY 4.0".

## Henkilötunnus

`fi_FI.henkilotunnus` is a Finnish personal identity code as DVV defines it on its page "Personal identity code",
`ddmmyy`, a century sign, an individual number and a control character, such as `131052-9085`. Its individual
number is from 900 to 999, which DVV issues to no one: its page "Personal identity code" says the numbers issued run
from 002 to 899, and its test table on "Reform of personal identity code" holds codes with 902 to 904. A drawn number
is no living person's, but a checker that treats endings from 900 as synthetic ("keinohenkilötunnus"), as the
[Vero API FAQ on testing](https://www.vero.fi/tietoa-verohallinnosta/kehittaja/vero-api/yleiset-virhetilanteet-ja-ratkaisut/ukk/vero-api-testaus-onko-henkilotunnuksen-muodolle-jotain-tiettyja-vaatimuksia/) does, may apply more checks or refuse it. The sign is `+` for the 1800s, `-` for the 1900s and `A`
for the 2000s; fejkdata doesn't draw the letters DVV has added for the 1900s and 2000s since 2023. The individual
number is odd for a man and even for a woman, and the control character is the remainder of the nine digits of the date
and the individual number divided by 31, read in DVV's table of `0` to `9` and the letters `A` to `Y` without `G`,
`I`, `O`, `Q` and `Z`. The date is of someone aged 0 to 99, drawn evenly within two bands that are weighted by [Statistics Finland](https://stat.fi/en/news/Finland-in-Figures-2026-ageing-Finland-is-visible-in-numbers)'s
1,347,494 of 5,652,881 residents aged 65 or over at the end of 2025: 23.84% of the dates are of the band aged 65 to 99 and
the rest of the band aged 0 to 64. A date of someone 100 or older is not drawn. The builtin `{hetu(x)}` completes a number
after its date: it reads the last six digits emitted, takes the sign from the latest year not after the date of `--today`
or `WithToday` that bears the year digits, then adds an individual number of the parity of the digit `x` reads and the
control character, and renders nothing in the cases the Functions table lists. A fixture of a few
thousand rows starts to repeat a number, and a `UNIQUE` column needs a category of your own.

## Henkilötunnus matched to a sex

`fi_FI.individ`, under `sex`, holds one digit of each sex's parity. To make a number agree with the
sex the template below binds as `s`, read the `birth` field of `fi_FI.henkilotunnus`, the date, then `{hetu(x)}` with the digit
that `individ` gives:

```json
{ "format": "{first} {last}, {hetu}",
  "first": "{/fi_FI.sex as s}{s.first-name}",
  "last": "{/fi_FI.last-name}",
  "hetu": "{/fi_FI.henkilotunnus.birth}{hetu(s.individ.digit)}" }
```

Renders e.g. `Eila Kirjavainen, 260163-9405`.

## Henkilötunnus of the issued range

`fi_FI.henkilotunnus-issued` draws the same dates with an individual number from 002 to 899, the range DVV
issues, such as `160432-581J`. It is for a checker that refuses the numbers from 900. Such a number is valid, and may be a living
person's, so keep it out of anything that reaches a real system. To make one agree with a bound sex, write
`{hetu(s.individ.digit,'issued')}` for `{hetu(s.individ.digit)}` in the template above.
