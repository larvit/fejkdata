"""Rebuild data/et_EE/first-name.tsv and last-name.tsv from a language model's recall.

    data-import/names-ee.py [--out DIR]

- Every name is a language model's recall of a name common in Estonia.
- Each list holds the commoner names from the most to the least common, then a line `=`, then the rarer ones, sorted. A
  commoner name has the count 100000 over the square root of its rank, and the rarer ones share the count of the last
  rank. These counts are estimates, not published figures.
- A last name has a row under each sex. A Russian one ending in `ov` or `ev` has the same name with an `a` added under
  `f`, as Estonia's Russian-speaking women bear it; the others are the same under both.
"""
import argparse
from pathlib import Path

import recalled_names
import tsv

OUT = Path(__file__).resolve().parent.parent / "data" / "et_EE"
TOP = 100000

GIRLS_RECALLED = """
Olga Tatjana Svetlana Maria Anna Jelena Natalja Irina Ljudmila Marika
Kristiina Liis Mari Tiina Ülle Maire Helle Katrin Piret Kadri
Anneli Merike Külli Liisa Maarja Margit Sirje Aino Eha Reet
Epp Riina Inge Aili Ene Evelin Triin Annika Kaisa Kaja
Krista Heli Helen Mare Ave Anu Aime Malle Lea Ingrid
Karin Margarita Nina Galina Valentina Julia Viktoria Oksana Alina Anastassia
Darja Ksenija Jekaterina Veera Nadežda Elvira Raissa Larissa Marina Aleksandra
Eve Elle Leida Linda Marta Liina Kersti Jaana Pille Aet
Signe Hilja Helga Aita Maia Laine Tiiu Urve Tiia Vilja
Birgit Kristel Kerli Laura Marge Merle Mirjam Monika Ruth Sirle
=
Agnes Aija Aire Aivi Alisa Alla Amanda Andra Andrea Angelika
Anita Anne Annela Anneliis Antonina Arabella Ariana Asta Astrid Auli
Brigitta Britt Carmen Dagmar Diana Dina Ede Edith Eerika Egle
Eleonora Eliisa Elina Elisabeth Elise Ella Elmi Elna Elo Emilia
Emma Erika Erna Esta Etti Eva Evi Gerda Gerli Getter
Grete Gretel Hanna Hedi Heidi Helina Henriette Hiie Ilona Ilse
Imbi Iris Isabel Jaanika Jana Janika Jelizaveta Jenna Johanna Juta
Kai Kaidi Kaire Kairi Karina Katarina Kati Katre Kelly Kerstin
Kirsti Klara Kristi Kristin Kärt Kätlin Laila Leelo Leili Lemme
Lenna Lidia Liia Lilian Lilja Lilli Lisbeth Lola Luule Lydia
Madli Maret Margot Marianne Marit Marju Marleen Marlen Meeli Melanie
Merili Merit Mia Milvi Nele Nora Olivia Pilvi Pirjo Raili
Rebekka Reelika Regina Renate Riin Rita Rosalie Saara Sandra Siiri
Silja Silvi Sirli Sofia Sonja Stella Tamara Tea Teele Tereza
Tuuli Ulvi Vaike Valve Veronika Viire Viivi Vilma Virve Zinaida
Zoja Õie
"""

BOYS_RECALLED = """
Aleksandr Sergei Vladimir Andrei Martin Aleksei Andres Dmitri Igor Toomas
Viktor Margus Roman Kristjan Oleg Jüri Maksim Urmas Marko Aivar
Alexander Mihhail Jaan Peeter Tõnu Mati Priit Rein Ivan Anton
Artur Nikolai Jevgeni Denis Valeri Anatoli Vitali Konstantin Pavel Vadim
Kaido Tarmo Raivo Ando Indrek Mart Taavi Madis Karl Kristo
Hans Erik Siim Meelis Marek Marten Ahto Arvo Heiki Ilmar
Janek Jaak Jaanus Kalev Kalle Lauri Mihkel Oskar Oliver Robin
Rasmus Sander Sten Tanel Tiit Uku Urmo Vello Veiko Voldemar
Leo Lembit Ülo Hugo Mark Aarne Aavo Adrian Agu Ain Aigar
=
Ago Ahti Alar Albert Aleksander Allan Alo Alvar Andre Anti
Ardi Argo Arko Arne Arnold Artjom Arvi Asko August Boris
Dan Daniel David Edgar Eduard Eerik Egon Eino Elmar Emil
Endel Enn Enno Erki Ervin Evald Fjodor Gennadi Georg Gert
Gunnar Harri Harry Heino Helmut Hendrik Henn Henri Henrik Herman
Hillar Ilja Illar Illimar Imre Ivo Jaagup Jakob Jan Jarmo
Jarno Joel Johannes Joonas Kaarel Kaspar Kaupo Kermo Kert Kevin
Kirill Kristofer Kuldar Kurt Kustas Lars Lennart Leonid Lev Madar
Magnus Marcus Mario Markus Marti Mattias Mikk Märt Nikita Nils
Olev Otto Paul Ragnar Raimo Rain Rainer Raul Reigo Remo
Richard Roland Rolf Ruslan Silver Stanislav Sulev Taimo Tauno Teet
Tenno Timo Toivo Tõnis Uno Vahur Valdo Valdur Vallo Valter
Viljar Villu Vjatšeslav
"""

LAST_RECALLED = """
Tamm Saar Sepp Mägi Kask Kukk Ilves Rebane Pärn Koppel
Oja Ivanov Petrov Smirnov Kuznetsov Vassiljev Mihhailov Fjodorov Popov Pavlov
Stepanov Volkov Kuusk Raudsepp Lepik Karu Lepp Vaher Kütt Põder
Kallas Kivi Mets Kangur Liiv Lõhmus Laur Kuusik Kaasik Luik
Jõgi Järv Pukk Paas Aas Allik Orav Puusepp Rätsep Sild
Soo Teder Toom Uibo Valge Vares Varik Veski Viik Visnapuu
Õun Tõnisson Mänd Männik Mõttus Mitt Nurk Nurme Org Paju
Palm Peterson Pihlak Pill Põld Raud Reimann Roos Rand Saks
Salu Sarapuu Siim Sikk Silm Tomson Unt Vain Vaik Vilms
Zaitsev Aleksandrov Andrejev Sokolov Morozov Kozlov Sidorov Nikolajev Novikov Romanov
=
Aavik Adamson Aedma Alas Albert Aleksejev Anton Antonov Arro
Aru Asi Baranov Berg Bogdanov Eensalu Eller Erm Faber Gerassimov
Grigorjev Gross Haab Hallik Hansen Hein Heinsalu Helme Hint Hunt
Härm Ignatjev Jaakson Jakobson Jakovlev Jegorov Johanson Juhkam Jürgens Jürgenson
Kalda Kaljurand Kalm Kangro Kannel Karjus Kaur Kikas Kirs Klaus
Koit Kolk Kondratjev Kont Kopli Kovalenko Kreek Kroon Kruus Kuld
Kull Kurg Käär Kõiv Laane Laanemets Laas Laht Lang Lass
Lebedev Leht Lill Link Loorits Lukk Maasik Madisson Makarov Mand
Mark Martinson Medvedev Mikk Mölder Nael Nõmm Oks Olesk Orlov
Ots Paal Padar Pajula Part Pent Pesur Piir Pikk Pärtel
Raag Raid Rammo Ranna Reinsalu Rohtla Rooba Rull Rüütel Saarmann
Sagar Sander Saul Semjonov Sibul Sillaste Sõber Tali Tamberg Tamme
Tammik Tammsaar Tarassov Teär Tiik Titov Tobias Tooming Tuisk Tuul
Uus Vahtra Vaino Valk Vask Veber Viks Vilu Vool Võsu
Zujev Õispuu Õunapuu
"""


def feminine(name: str) -> str:
	return name + "a" if name.endswith(("ov", "ev")) else name


def lasts() -> list[dict[str, object]]:
	"""The rows (name, sex, count) of the last names: each recalled name under m, and its feminine form under f."""
	rows = recalled_names.weighted("last names", LAST_RECALLED, TOP)
	return [{"name": form(name), "sex": sex, "count": n} for name, n in rows for sex, form in (("f", feminine), ("m", str))]


def main() -> None:
	p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	p.add_argument("--out", default=str(OUT))
	out = Path(p.parse_args().out)
	out.mkdir(parents=True, exist_ok=True)
	firsts = [{"name": name, "sex": sex, "count": n} for sex, text in (("f", GIRLS_RECALLED), ("m", BOYS_RECALLED)) for name, n in recalled_names.weighted("girls" if sex == "f" else "boys", text, TOP)]
	tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(firsts, key=lambda r: (str(r["name"]), str(r["sex"]))))
	tsv.write(out / "last-name.tsv", ["name", "sex", "count"], sorted(lasts(), key=lambda r: (str(r["name"]), str(r["sex"]))))


if __name__ == "__main__":
	main()
