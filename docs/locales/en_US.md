# `en_US`

`en_US`'s categories are listed under the README's [Locale categories](../../README.md#locale-categories).

## English words

`en_US.word` draws evenly from about 70,000 words of WordNet 3.0, Princeton's English dictionary: the nouns,
verbs, adjectives and adverbs it spells in lower-case letters alone, so no proper noun or acronym. WordNet counts
the senses of only a quarter of its words in a text sample, so it gives no frequency to draw by, and a word may
be rare. `en_US.word.noun`, `.verb`, `.adj` and `.adv` each draw one part of speech. `en_US.word.verb.present`
draws a verb and prints its third person, such as `carries`. To read a verb and its third person from one pick,
bind it: `{/en_US.word.verb as v}{v} {v.present}` prints `carry carries`. The import drops a word whose every sense is in
WordNet's usage domain of disparagement, ethnic slurs or obscenity, and the words on LDNOOBW's English list of rude
words. A word with a harmless sense, such as `boy`, stays, and other rude words remain, so read the output before showing it to anyone. `en_US.sentence` fills a fixed
pattern with these words and doesn't make them agree, so its sentences aren't grammatical. `en_US.paragraph` joins
three to six sentences. WordNet's licence asks for its notice and disclaimer on every copy of its database:
`data/en_US/WORDNET-LICENSE.txt` holds them, the `en_US` module embeds the file, and the release archives and
packages carry it.

## Social Security numbers and ITINs

`en_US.ssn`, a Social Security number, and `en_US.itin`, an Individual Taxpayer Identification
Number, print `area-group-serial`, and each has those three columns. Both are valid, not
synthetic: a drawn number passes the issuing agency's published rules and nothing marks it
unissued, so it may be a living person's. A serial is `0001` to `9999`. Within the ranges below,
every number is equally likely.

- `en_US.ssn` draws the area from `001` to `665` and `667` to `899`, and the group from `01` to
  `99`. The Social Security Administration (SSA) assigns [no area 000, 666 or 900 to
  999](https://www.ssa.gov/employer/randomization.html), [no group 00 and no serial
  0000](https://www.ssa.gov/employer/randomizationfaqs.html). About 890 million numbers are
  possible, so a fixture of some tens of thousands of rows starts to repeat one.
- `en_US.itin` draws the area from `900` to `999` and the group from `50` to `65`, `70` to `88`,
  `90` to `92` or `94` to `99`, the groups the Internal Revenue Service (IRS) [lists for an
  ITIN](https://www.irs.gov/irm/part3/irm_03-021-263r). About 44 million numbers are possible,
  so a fixture of a few thousand rows starts to repeat one.

A `UNIQUE` column needs a category of your own whose id ends in `.{seq()}`, as
[Records](../../README.md#records) shows.

A checker older than the SSA's 2011 randomization, which refused areas the SSA had not yet
assigned, or one accepting fewer ITIN groups than the IRS lists, refuses some of these numbers:
pin a number for such a checker.

## US phone exchanges

`en_US.phone` draws its exchange from `en_US.phone-exchange`,
which draws evenly from the codes the North American Numbering Plan Administration (NANPA)
may assign: `200` to `999`, except `211`, `311` … `911`, `555`, `950`, `958`, `959` and
`976`.

`en_US.mobile` and `en_US.landline` draw one pool. A US number doesn't show its kind, since mobile
and fixed numbers share area codes and exchanges, so `en_US.landline` reads `en_US.mobile`, and
neither promises a kind. Both print `(202) 555-0123`, which `national` prints too, and
`international` prints `+1 202 555 0123`. Their area codes are `en_US.phone`'s, and their
exchanges `en_US.phone-exchange`'s.
