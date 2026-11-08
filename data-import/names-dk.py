#!/usr/bin/env python3
"""Rebuild data/da_DK/first-name.tsv and last-name.tsv from Statistics Denmark's top 20 and a language model's recall.

    data-import/names-dk.py [--out DIR]

- The top 20 girls' names, boys' names and last names keep the counts Statistics Denmark published for the
  people living in Denmark on 1 January 2026 (CC BY 4.0, "Kilde: Danmarks Statistik";
  dst.dk/en/Statistik/emner/borgere/navne/navne-i-hele-befolkningen).
- Every other name is a language model's recall of a name common in Denmark. The first of them run from
  the most to the least common, and each has the twentieth name's count times 20 over its rank. The rarer
  ones that follow, in no order, share the count of the last rank. These counts are estimates, no published figure.

A name that girls and boys both bear has a row under each sex.
"""
import argparse
import sys
from pathlib import Path

import tsv

OUT = Path(__file__).resolve().parent.parent / "data" / "da_DK"

GIRLS = {
    "Anne": 43021, "Mette": 37932, "Kirsten": 34588, "Hanne": 34482, "Anna": 33441,
    "Helle": 32720, "Maria": 30400, "Susanne": 29668, "Lene": 29290, "Marianne": 25598,
    "Camilla": 24494, "Lone": 23918, "Louise": 23824, "Charlotte": 23455, "Pia": 23359,
    "Tina": 22855, "Emma": 22590, "Ida": 22571, "Gitte": 22382, "Julie": 21912,
}
BOYS = {
    "Peter": 45777, "Michael": 43976, "Lars": 42596, "Thomas": 41800, "Jens": 40558,
    "Henrik": 40416, "Søren": 38087, "Christian": 37349, "Martin": 36799, "Jan": 35210,
    "Morten": 33384, "Jesper": 33139, "Anders": 32980, "Mads": 32239, "Niels": 30836,
    "Rasmus": 30222, "Mikkel": 29280, "Kim": 27770, "Per": 27640, "Ole": 26195,
}
LAST = {
    "Nielsen": 222355, "Jensen": 218886, "Hansen": 186105, "Andersen": 143880, "Pedersen": 140838,
    "Christensen": 106860, "Larsen": 102287, "Sørensen": 97587, "Rasmussen": 84147, "Jørgensen": 78057,
    "Petersen": 68658, "Madsen": 58255, "Kristensen": 55775, "Olsen": 42709, "Thomsen": 37486,
    "Christiansen": 33781, "Poulsen": 29488, "Johansen": 28983, "Møller": 28916, "Mortensen": 27141,
}

GIRLS_RECALLED = """
Karen Inge Birgit Jette Bente Annette Dorte Lisbeth Eva Ulla
Rikke Birthe Lotte Vibeke Ingrid Tove Berit Randi Jytte Lis
Sofie Katrine Line Trine Sara Mia Maja Laura Cecilie Johanne
Josefine Freja Clara Nanna Ellen Signe Sine Stine Mie Kamilla
Sanne Lisa Anette Karin Grethe Dorthe Ruth Bodil Merete Anja
Pernille Lærke Emilie Amalie Frederikke Caroline Isabella Sofia Nicoline Mathilde
Alberte Astrid Agnes Alma Olivia Victoria Marie Christina Kristine Katja
Michelle Malene Heidi Tanja Helene Henriette Winnie Hanna Jeanette Jane
Lena Lillian Elisabeth Else Margit Jonna Gerda Aase Agnete Alice
Birte Conny Doris Gunhild Inger Margrethe Sonja Jenny Linda Susan
Kim Birgitte
=
Abelone Adelina Agathe Agnethe Aina Alette Alexa Alexandra Alicia Alina
Alva Amanda Amy Anastasia Andrea Ane Angelina Anita Annelise Annemarie
Annemette Anni Annie Annika Anny Ashley Aurora Beate Benedikte Betina
Bettina Bibi Bolette Bonnie Brita Britt Camille Carina Carla Carmen
Carolina Cassandra Cathrine Cecilia Celina Cille Connie Dagmar Dagny Diana
Dicte Dina Dorit Dorrit Edda Edith Eira Elena Eleonora Eleonore
Elin Eline Elise Elizabeth Ella Ellinor Elna Elsa Elsebeth Elva
Elvira Emelie Emely Emilia Erika Erna Ester Esther Evelyn Fabiana
Fanny Fatma Felicia Fie Filippa Fiona Frida Gerd Gertrud Grete
Gry Gudny Gudrun Gunna Gunvor Hannah Hanni Hedda Hedvig Helen
Helga Henny Hermine Hilde Hulda Iben Ilsa Ina Ines Ingeborg
Ingelise Irene Iris Isa Isabel Isabelle Isolde Jacoba Jana Janne
Janni Jannie Jasmin Jasmine Jeanne Jessica Jill Joan Joanna Johanna
Jolanda Josephine Julia Juliane Kaja Kamille Kamma Karina Karolina Karoline
Kate Katharina Kathrine Kirstine Kitte Kristin Laila Lajla Laurine Lea
Leonie Leonora Lilian Lilli Lillie Lilly Lina Linnea Lise Liselotte
Liv Lizette Lola Lotus Louisa Lucca Lucia Lucy Lulu Luna
Lykke Madeleine Madelene Magdalene Mai Maia Maiken Maj Majbritt Majken
Malin Malou Mari Mariam Marianna Marlene Marta Martha Maryam Mathea
Matilde Maya Melanie Melissa Meta Michaela Mikala Mille Minna Mira
Molly Mona Monica Mynte Nadia Nadja Naja Naomi Natasja Nathalie
Nelly Nicole Nina Ninna Nora Nynne Olga Oline Patricia Paula
Pauline Petra Petrine Pil Rachel Ragnhild Rakel Rasmine Rebekka Regina
Rie Rigmor Rita Ronja Rosa Rose Rosita Roskva Sabine Sabrina
Saga Sally Salma Sandra Sarah Sasha Selma Sia Sidsel Sigrid
Silje Simone Siri Solvej Sonia Stella Stephanie Susanna Sussie Svea
Tanya Tara Tessa Thea Theodora Therese Tilda Tilde Tonja Tuva
Una Ursula Vanessa Vera Vicky Vilma Vivi Vivian Wilhelmine Yvonne
Zara Zaynab Zoe Zofia
"""

BOYS_RECALLED = """
Poul Jørgen Hans Kurt Flemming Bent Carsten Claus Kenneth Brian
Allan Dennis Jacob Mogens Finn Leif Erik Bo Palle Steen
Benny Frank Jonas Mathias Magnus William Oliver Emil Noah Victor
Alexander Frederik Lucas Nikolaj Simon Daniel Kasper Mark Jonathan Patrick
Tobias Philip Marcus Sebastian Rune Lasse Casper Kristian Johan Andreas
Benjamin Oscar Malthe Villads Anton August Albert Valdemar Elias Alfred
Carl Marius Gustav Johannes Jakob Bjarne Preben Knud Svend Torben
Tommy René Ivan Ib Viggo Kjeld Verner Arne Aage Henning
Jørn Ejvind Egon Gunnar Helge Holger Ulrik Kent Dan Jim
John Jimmy Steffen Stefan Sten Thorkild Tom Thorsten Uffe Lennart
Karsten Klaus Kaj Kåre Joakim Joachim Jasper Jeppe Bastian Jannik
Mikael Tim Hjalte Torsten Rolf Roland Ebbe Eigil Erland Esben
Freddy Georg Karl Kevin Ronni Leo Laust Lauge Nikolas Oluf
Otto Peder Ricky Robert Sigfred Thorvald Tonny Torkil Villum Vilhelm
Mohamed Mohammad Ahmad Ali Omar Yusuf Hussein
=
Adam Adrian Aksel Alan Aleksander Alex Alf Alvin Amir Anker
Arvid Asbjørn Asger Aske Atle Axel Bashir Bendt Bertil Birger
Bjarke Bjørn Bobby Bruno Børge Carlo Charles Christoffer Christopher Claes
David Dion Dominik Edmund Edvard Edward Ejnar Elof Emanuel Emilio
Eric Erling Ernest Ernst Eske Eskild Ewald Fabian Folmer Frede
Frits Frode Gerhard Gert Gorm Gregers Gustaf Hakon Halfdan Harald
Harry Hasse Heine Henri Herman Hermann Hjalmar Ingemar Ingvar Isak
Iver Jack James Jarl Jarne Jeremy Jerry Jesse Jochum Joe
Jon Jonny Josef Joseph Julian Julius Jürgen Karim Ketil Kian
Kjartan Kresten Leon Levi Liam Lorenz Louis Luca Ludvig Lukas
Magne Manfred Marco Mario Markus Marvin Maurice Max Maximilian Melvin
Morgan Nicklas Nicolai Nils Noa Olaf Olav Olof Oskar Ove
Paul Pelle Rafael Ragnar Ravn Reinhard Rickard Robin Ronnie Sam
Samuel Sander Sigurd Sigvald Silas Simeon Skjold Stig Storm Sune
Svenn Sverre Tage Theis Theodor Thor Timm Toke Troels Tue
Tyge Ulf Ulrich Vagn Valentin Vigo Villy Vincent Walter Walther
Werner Wolfgang Xavier Yngve Zacharias
"""

LAST_RECALLED = """
Frederiksen Jakobsen Mikkelsen Holm Lund Schmidt Eriksen Kristiansen Henriksen Simonsen
Clausen Andreasen Iversen Jeppesen Mogensen Jacobsen Olesen Laursen Frandsen Knudsen
Lauridsen Bertelsen Berg Svendsen Johannesen Thomassen Hermansen Nissen Dam Bruun
Winther Lassen Skov Dahl Bech Kjær Kjeldsen Lauritsen Therkildsen Thorsen
Vestergaard Lindberg Lorentzen Michelsen Nørgaard Nyholm Rosenberg Schou Toft Overgaard
Bøgh Brandt Bak Bruhn Christoffersen Damgaard Ebbesen Eliasen Enevoldsen Fischer
Friis Gregersen Grønbæk Hald Hedegaard Hjorth Holst Jessen Jochumsen Juhl
Kaas Karlsen Kofoed Krogh Kruse Lange Leth Lindholm Mathiesen Nygaard
Nørregaard Paulsen Rask Riis Schultz Skovgaard Steffensen Strøm Sundby Svane
Sønderby Thygesen Villadsen Winter Ravn Bendtsen Bjerregaard Carlsen Dalsgaard Ejlersen
Falk Foged Gram Hartmann Haugaard Hviid Højer Ibsen Jepsen Juul
Klausen Lindgren Lykke Munk Neergaard Nørgård Ohlsen Pallesen Quist Thorup
Valentin Vinther Wagner Wulff
=
Aagaard Abrahamsen Albertsen Asmussen Axelsen Bach Bagge Balle Beck Becker
Bentsen Bille Bjerre Bjørn Bonde Borg Borup Boysen Brink Brix
Brodersen Buus Callesen Christophersen Dahlgaard Dalgaard Daugaard Ditlevsen Dreyer Duus
Eilersen Ellegaard Elmholt Engelbrecht Engholm Esbensen Faber Fabricius Fallesen Fjord
Fogh Frost Fuglsang Gad Gade Gundersen Gustavsen Hammer Hauge Heide
Helmer Hemmingsen Hoffmann Holmberg Holme Husted Hvidt Høeg Høgh Højgaard
Hørup Isaksen Jacobi Jansen Jonassen Juncker Justesen Jæger Kirk Kjeldgaard
Klitgaard Koch Kragh Krarup Kristoffersen Kruuse Kvist Lauersen Lehmann Lerche
Lorenzen Lundgaard Lykkegaard Magnussen Marcussen Markussen Mathisen Meyer Mouritsen Nedergaard
Neumann Nordahl Nordentoft Ottosen Overby Pagh Pape Roed Rosendahl Rosenkilde
Schrøder Schwartz Skaarup Stenbæk Stephansen Stoltenberg Sylvest Thaysen Trolle Tønnesen
Vang Weber Wilhelmsen Wittrup Østergaard
"""


def ranked(top, recalled, what):
    """The rows (name, count) of a list: the top 20 as published, the recalled names in order, then the rarer ones, which share a count."""
    ordered, _, rarer = (part.split() for part in recalled.partition("\n=\n"))
    names = list(top) + ordered + rarer
    dupes = sorted({n for n in names if names.count(n) > 1})
    if dupes:
        sys.exit(f"{what}: listed twice: {', '.join(dupes)}")
    for name in names:
        if not tsv.NAME.match(name):
            sys.exit(f"{what}: {name!r} is not a name in letters")
    twentieth = min(top.values())
    rows = list(top.items())
    for rank, name in enumerate(ordered, start=len(top) + 1):
        rows.append((name, twentieth * len(top) // rank))
    rows += [(name, twentieth * len(top) // len(names)) for name in rarer]
    return rows


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    first = [{"name": n, "sex": "f", "count": c} for n, c in ranked(GIRLS, GIRLS_RECALLED, "girls")]
    first += [{"name": n, "sex": "m", "count": c} for n, c in ranked(BOYS, BOYS_RECALLED, "boys")]
    last = [{"name": n, "count": c} for n, c in ranked(LAST, LAST_RECALLED, "last names")]
    out = Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
    tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
    main()
