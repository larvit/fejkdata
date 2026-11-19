# `nn_NO`

`nn_NO`, Norwegian in the Nynorsk standard, carries `address`, `color`, `company`, `date`, `email`, `fodselsnummer`,
`month`, `paragraph`, `person`, `phone`, `price`, `sentence`, `time`, `title`, `url`, `username`, `weekday`, `word`, `word-adj`, `word-adv`, `word-noun` and `word-verb`. Both standards spell names, places, streets, ids and phone
numbers alike, so `nn_NO` reads `nb_NO`'s data and `geo.NO`. It has no table of names, `sex`, `birth-day`,
`individ` or phone prefixes, and none of `nb_NO.company`'s forms such as `.as`: read those from `nb_NO`.

## Persons, ids, contact details and companies

`nn_NO.person`, `nn_NO.fodselsnummer`, `nn_NO.phone`, `nn_NO.email`, `nn_NO.company` and `nn_NO.date` draw what the
`nb_NO` category of the same name draws, and are the same kind of value. `nn_NO.person` takes its prefix from
`nn_NO.title`, and `nn_NO.address` prints `geo.NO.address`. Email local parts and company names hold Bokmål
words, read from `nb_NO`. To match a fødselsnummer to a sex, bind `nb_NO.sex` as
[Fødselsnummer matched to a sex](nb_NO.md#fødselsnummer-matched-to-a-sex) shows.

## Nynorsk months, weekdays and titles

`nn_NO.month` and `nn_NO.weekday` are Unicode CLDR's Nynorsk names. `{/nn_NO.month[1]}` is `januar` and `[12]` is
`desember`; `{/nn_NO.weekday[2]}` is `tysdag` and `[7]` is `søndag`, Monday first as ISO 8601 numbers them.
`nn_NO.title` holds `dr.` and `professor`, and has no `parent`, so it is selected as `nn_NO.title[dr.]`.
The month and weekday names are real.

## Nynorsk words

`nn_NO.word` draws evenly from about 106,000 words of Norsk ordbank's Nynorsk 2012, from Språkbanken:
the common nouns, verbs, adjectives and adverbs it spells in lower-case letters alone, in the forms the
current norm holds. The ordbank has no frequencies, so a word may be rare. `nn_NO.word.noun`, `.verb`,
`.adj` and `.adv` each draw one part of speech, and `nn_NO.word.noun.definite`, `nn_NO.word.verb.present`
and `nn_NO.word.adj.definite` print one form, such as `huset`, `kastar` or `store`. Bind a word to read it
and its form from one pick: `{/nn_NO.word.verb as v}{v} {v.present}` prints `kaste kastar`.

The import drops the words on LDNOOBW's Norwegian list of rude words, and other rude words remain, so
read the output before showing it to anyone. `nn_NO.sentence` fills a fixed pattern with these words
and doesn't make them agree, so its sentences aren't grammatical. `nn_NO.paragraph` joins three to six sentences.
A program showing these words credits "Språkbanken, Nasjonalbiblioteket: Norsk ordbank nynorsk 2012, CC BY 4.0".
