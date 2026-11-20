# `geo`

The `geo` namespace holds one tree per country under its alpha-2 code: five [linked tables](../../README.md#linked-tables) named alike, and an `address` record binding one locality and drawing its postal code inside it, and its street too where the country's streets are real. A locale's `address` prints that record, and each of its columns describes the address printed.

| Table | `geo.DE` | `geo.DK` | `geo.EE` | `geo.FI` | `geo.FR` | `geo.LV` | `geo.NO` | `geo.SE` | `geo.US` | Weight |
|-------|----------|----------|----------|----------|----------|----------|----------|----------|----------|--------|
| `region` | Land, by code (`BY`) or name | region, by code or name | maakond, by EHAK code (`0037`) or name | maakunta, by code or name | région, by INSEE code (`84`) or name | novads or state city, by ATVK code (`0051000`) or name | fylke, by code or name | län, by code or name | state, by USPS abbreviation or name; `code` is the FIPS code | DE: postal codes; EE and LV: addresses; elsewhere population |
| `municipality` | Kreis, by code or name | kommune, by code or name | vald or linn, by EHAK code (`0198`) or name | kunta, by code or name, with its `swedish-name` | département, by code (`2A`) or name | pagasts or pilsēta, by ATVK code (`0048460`) or name | kommune, by code or name | kommun, by code or name | county, by FIPS code or name | DE: postal codes; EE and LV: addresses; elsewhere population |
| `locality` | place name within a Kreis, by code: every one with an accuracy, GeoNames' measure of how exactly it knows the point | postdistrikt, by name: every one with a street address in Danmarks Adresseregister | asustusüksus, or a linn whole, by EHAK code (`0784`): every one with a street in the address data | postitoimipaikka, by name, with its `swedish-name`: every one with a street in Posti's basic address file under a normal postinumero with residents | commune, by INSEE code (`75056`): every one with people and an address in the Base Adresse Nationale | pilsēta or ciems, by its code in the State Address Register (`100003003`): every one with an address on a street | poststed, by name: every one with a street address in Kartverket's register | postort, by name; [`geo-se.py`](../../data-import/geo-se.py)'s `localities` picks the candidates, and one with no street is dropped | incorporated place, by GEOID or name; Hawaii has none | DE: its postal codes; DK and NO: the kommune's population, shared out over its postdistrikter or poststeder by their addresses; EE and LV: its addresses; FI: its postinumerot's residents; FR: the commune's population; SE: tätort population, the kommun's where the postort names it, the småort's where one holds the postort's centre, else 200; US: place population |
| `postal-code` | Postleitzahl, one row per place it holds: every one with an accuracy | postnummer with a street address, by code | sihtnumber, one row per place it holds | normal postinumero with a street and residents, by code | code postal, one row per commune it holds | pasta indekss, written `LV-1050`, one row per place it holds | postnummer with a street address, by code | postnummer with street delivery, by code | ZCTA, by code | DE: one; DK, EE, FR, LV and NO: addresses; FI: residents in Statistics Finland's Paavo; SE: one; US: address ranges |
| `street` | invented street name, belonging to no place | street name, every one in Danmarks Adresseregister | street name, every one that ends in a kind of street, such as `tn` or `tee` | street name, every one [`geo-fi.py`](../../data-import/geo-fi.py)'s `street_name` accepts, with its `swedish-name` | invented street name, belonging to no place | street name, every one with a numbered address | street name, every one [`geo-no.py`](../../data-import/geo-no.py)'s `street_name` accepts | gatunamn, every one [`geo-se.py`](../../data-import/geo-se.py)'s `street_name` accepts | street name, per place the quarter with most address ranges, at least ten | DE and FR: one; DK, EE, LV and NO: addresses; FI: ranges in Posti's basic address file; SE: segments; US: address ranges |

## Across countries

What ports across countries is the five table names, their `name` column, selection by name, and the `address` record's columns: `street`, `street-number`, `postal-code`, `locality`, `municipality` and `region`. The record's `region` is a US state's abbreviation, and elsewhere the region's name. A locality in `geo.DE`, `geo.EE`, `geo.FR` or `geo.LV` is selected by its code alone, since place names repeat. Every other column of a table is the country's own, `code` on a Swedish region but `abbr` on a US one.

## Danish places

`da_DK.address` prints `geo.DK.address`, such as `Vestergade 12\n8000 Aarhus C`. A postdistrikt such
as `København K` holds many postnumre. A kommune's `code` is four digits, `0101`, as the address register
writes it, and a region's is `1084`, as DAGI, the register of Denmark's divisions, writes it. Christiansø ships no place: its addresses lie in the
postdistrikt Gudhjem, on Bornholm. A Danish street sits under each postdistrikt where it has an address.

## Estonian places

`et_EE.address` prints `geo.EE.address`, such as `Sarapuu tn 70\n76901 Tabasalu alevik`. A delivery letter
is written in lower case, as in `2b`. Every table weighs a row by its addresses in the Land and Spatial Development
Board's address data, and each code is the four-digit EHAK code, the board's classification of Estonia's divisions.
A locality is an asustusüksus: a küla, an alevik, an alev or a linn inside a vald. A linn that governs itself is one
locality under its own code, and its linnaosad are no places, so Tallinn is `0784` whole. A street's name ends in its
kind, written as the address data writes it: `tn`, `mnt` and `pst` for tänav, maantee and puiestee, and `tee`, `allee`,
`põik`, `väljak` and `plats` in full. Most külad have no street: an address there is a farm's name, such as `Antsu`. A
locality with no street is left out, which leaves 1,131 localities holding 78% of the addresses, and leaves out four
island municipalities whole: Kihnu, Muhu, Ruhnu and Vormsi. A letter to a küla names its vald and its maakond, which
the record's `municipality` and `region` columns hold. An Estonian street sits under each locality where it has an
address.

## Finnish places

`fi_FI.address` prints `geo.FI.address`, such as `Huvilakatu 34\n00700 Helsinki`. A delivery letter
is written in lower case, as in `12a`. Posti's files hold no street on Åland, so Åland ships no place. Posti
writes a postitoimipaikka in capitals, `HEINOLA KK`, and `geo.FI` writes it as a place name, `Heinola kk`.
A kunta's `code` is three digits, `091`, and a maakunta's two, `01`, as Statistics Finland writes them.
A postitoimipaikka, a kunta and a street each have a `swedish-name`. A postitoimipaikka's is the Swedish name
Posti gives its postinumerot most often, written as a place name, such as `Helsingfors` for `Helsinki`. A kunta's is
Statistics Finland's, such as `Åbo` for `Turku`. A street's is the one Posti gives it most often in its
postitoimipaikka, such as `Nylandsvägen` for `Uudenmaantie`. Where the source gives none, `swedish-name` holds the
Finnish name. A Finnish street sits under each postitoimipaikka where Posti lists it.

## French places

`fr_FR.address` prints `geo.FR.address`, such as `94 Rue des Hirondelles\n14490 Balleroy-sur-Drôme`. A region is a
région, a municipality a département, and a locality a commune, selected by its INSEE code: `{/geo.FR.locality[2A004]}`
prints `Ajaccio`. The names, the divisions and each commune's population are those of
[INSEE](https://www.insee.fr/fr/statistiques/8680726)'s legal populations of 2023, counting the people whose usual home
is in the commune. The postal codes are the ones the [Base Adresse Nationale](https://adresse.data.gouv.fr/) files a
commune's addresses under, each weighted by its addresses there. Both sources file Paris, Lyon and Marseille by
arrondissement. `geo.FR` holds each as one commune, `75056`, `69123` and `13055`, with all its arrondissements' people
and postal codes, as a letter addresses `75011 Paris`. A commune with no people, or with no address under a postal code,
ships no place. INSEE's file leaves out Mayotte and the overseas collectivities, such as Saint-Pierre-et-Miquelon, so
they ship none either. A région overseas keeps its own time zone, in its `timezone` column.

The streets are invented, until `geo.FR` takes its streets from the Base Adresse Nationale. `geo.FR.street` holds street
names that a language model recalls as common in France. Each is drawn as often as any other, beside any commune. A
street number runs from 1 to 9999. One, two, three and four digits are each drawn in the share of the Base Adresse
Nationale's addresses that have as many. About one in twenty-five ends in ` bis`, as theirs do. A `ter`, a `quater` or a letter is never drawn.

## German places

`de_DE.address` prints `geo.DE.address`, such as `Hauptstraße 12\n01945 Tettau`. The places and postal codes are
[GeoNames](https://www.geonames.org/)' Postleitzahlen. A region is a Land, a municipality a Kreis, and a locality a place
name within a Kreis. A locality's `code` is its Kreis's `code` and its name in ASCII, with umlauts and ß written out, as
`12066-tettau`, so a refresh of GeoNames' file leaves it where it was. Two villages of one name in a Kreis are one locality, at the mean of their points. A Kreis is named as a German writes it, `Stadt Salzgitter`
for GeoNames' `Salzgitter, Stadt`. GeoNames writes a Landkreis with or without the word, `Landkreis Traunstein` but
`Bamberg`, and `geo.DE` keeps that. All of Hamburg and of Berlin is one place each, as a letter addresses it. GeoNames gives each place an
accuracy, its measure of how exactly it knows the point. A row without one is mostly a firm's or an office's own
Postleitzahl and ships no place. Jungholz, which is Austrian, ships none either. GeoNames publishes no population, so a
locality weighs the postal codes it holds, and a Kreis and a Land the sum of theirs. That counts postal codes, not people:
small places and the Länder with many of them are drawn more often than their people warrant.

The streets are invented. The Länder each publish their own address data, and no open register is known to cover all
sixteen. So `geo.DE.street` holds street names that a language model recalls as common in Germany. Each is drawn as often as
another and belongs to no place, so a street sits beside any postal code. The weights of the house-number lengths are
estimates too.

## Latvian places

`lv_LV.address` prints `geo.LV.address`, such as `Strautu ceļš 32\nJelgava, LV-3001`: the postal code follows the
locality after a comma, as the Universal Postal Union's guide to Latvian addresses writes it. Every table weighs a row
by its addresses in the State Land Service's State Address Register: a building is one address, and each apartment in
it one more. The register holds no level above the novads, so a region is a novads or one of the seven state cities,
such as Rīga. A municipality is a pagasts or a pilsēta, and a locality a pilsēta or a ciems: a pilsēta is both, and a
state city all three. A name is written as the register writes it, `Talsu nov.` and `Abavas pag.`. A region and a
municipality keep their seven-digit codes in ATVK, the classification of Latvia's territories. A ciems has none, so a
locality keeps its nine-digit code in the register. An address on a street is a house number, and one straight under a
ciems or a pagasts is a house's name, such as `Riņņi`. A locality with no numbered address on a street is left out,
which leaves 819 localities holding 83% of the addresses, and leaves out 141 pagasti whole. A house number may end in a
capital letter, as in `4A`; a korpuss, as in `195 k-1`, is never drawn. A Latvian street sits under each locality where
it has a numbered address.

## Norwegian places

`nb_NO.address` prints `geo.NO.address`, such as `Thorstein Stenbeks vei 65\n3613 Kongsberg`.
Svalbard has no kommune in Statistics Norway's classification, so it ships no place.
Kartverket's register spells a poststed in capitals, `MO I RANA`, and `geo.NO` writes it as a
place name, `Mo i Rana`. A kommune or fylke that has an official name in more than one language,
such as `Trondheim - Tråante`, ships the Norwegian one, `Trondheim`, which Kartverket writes.
Kommuner that share a name keep Statistics Norway's fylke in brackets, `Herøy (Nordland)` and
`Herøy (Møre og Romsdal)`, so each name selects one. A Norwegian street sits under each poststed
where it has an address.

## Swedish places

`geo.SE.region[Skåne län].municipality` draws a kommun in Skåne, `geo.SE.locality[Lund].street` a street in Lund.

Do not assert that a Swedish street lies in its postort: GeoNames gives a postal code a point and a postort no area, so the import places each road segment by distance, leaning towards the postort a tätort is named after, and a street may sit under a neighbouring one.

## US places

`geo.US.region[IL].locality[Springfield]` settles which Springfield. The US places come
from the incorporated places that hold most of some postal code's land: a tenth of them,
drawn with odds by population, every state capital but Honolulu, and every place of
100,000 or more. A smaller city may be missing, so look one up in
[`locality.tsv`](../../data/geo/US/locality.tsv) before selecting it. A region row
carries its `timezone`, the state's predominant zone, and a locality its `lat` and
`lon`.
