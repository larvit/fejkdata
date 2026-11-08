#!/usr/bin/env python3
"""Rebuild data/da_DK/first-name.tsv and last-name.tsv from Statistics Denmark's top 20 and a language model's recall.

    data-import/names-dk.py [--out DIR]

- The top 20 girls' names, boys' names and last names keep the counts Statistics Denmark published for the
  people living in Denmark on 1 January 2026 (CC BY 4.0, "Kilde: Danmarks Statistik";
  dst.dk/en/Statistik/emner/borgere/navne/navne-i-hele-befolkningen).
- Every other name is a language model's recall of a name common in Denmark, listed from the most to the
  least common. Its count is the twentieth name's count times 20 over its rank, an estimate that keeps the
  order and is no published figure.

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
Kim Birgitte Ella Liv Luna Selma Filippa Sally Naja Nora
Vilma Esther Silje Thea Tilde Karoline Cathrine Lykke Iben Irene
Janni Joan Kathrine Kirstine Lea Lilian Liselotte Majbritt Maiken Mona
Nina Rie Rose Sidsel Stephanie Sussie Ursula Vivi Yvonne Mille
Maj Mari Lina Lise Karina Kitte Laila Lilly Leonora Hedvig
Edith Elna Hanni Helga Gudrun Hannah Lizette Mai Marlene Mikala
Solvej Sabrina Ronja Rebekka Pauline Patricia Natasja Nadia Julia Johanna
Jessica Jasmin Ina Britt Anita Andrea Amanda Alexandra Annie Anni
Annika Annelise Annemarie Annemette Nicole Melanie Melissa Monica Paula Petra
Rachel Rosa Sandra Sigrid Simone Stella Vera Zara Regina Rigmor
Ragnhild Gry Gunvor Grete Gertrud Fie Fanny Evelyn Eline Elise
Elsebeth Emelie Elin Dina Dagmar Connie Cecilia Celina Bettina Betina
Benedikte Beate Ane Angelina Anastasia Alva Alicia Agathe Adelina Abelone
Gunna Hedda Ingeborg Ingelise Jannie Kaja Kamille Katharina Kate Lola
Louisa Lucia Madeleine Magdalene Majken Malou Martha Marta Matilde Michaela
Mira Naomi Olga Oline Aurora Bolette Camille Carina Carla Carmen
Carolina Cassandra Dorrit Eleonora Elizabeth Ellinor Elvira Emilia Erika Felicia
Fiona Frida Gerd Iris Isabel Isabelle Alexa Alina Amy Dagny
Diana Edda Eleonore Elena Elsa Erna Ester Hilde Isa Jana
Janne Jasmine Jeanne Jill Josephine Juliane Kamma Karolina Lajla Laurine
Leonie Lillie Lulu Maia Malin Maryam Mathea Maya Meta Minna
Molly Mynte Nadja Nathalie Ninna Pil Rita Rasmine Sabine Salma
Sarah Sasha Saga Sia Siri Sonia Susanna Svea Tara Tessa
Theodora Therese Tilda Una Vanessa Vivian Wilhelmine Zoe Aina Alette
Anny Bibi Bonnie Brita Cille Dicte Dorit Eira Elva Emely
Gudny Helen Hermine Ines Isolde Jacoba Kristin Lilli Linnea Lucca
Lucy Madelene Nelly Nynne Petrine Henny Agnethe Ashley Fatma Jolanda
Joanna Zofia Zaynab Vicky Tuva Tonja Tanya Rakel Rosita Roskva
Mariam Marianna Lotus Ilsa Hulda Fabiana
""".split()

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
Mohamed Mohammad Ahmad Ali Omar Yusuf Hussein Amir Bashir Vagn
Villy Ulf Axel Aksel Alex Alf Alan Aleksander Asger Asbjørn
Aske Bendt Birger Bjørn Bobby Bruno Charles Christoffer David Edvard
Edward Emanuel Erling Ernst Gert Gorm Harald Herman Jack Joe
Jon Josef Joseph Julian Julius Lukas Louis Ludvig Luca Liam
Leon Levi Marco Mario Markus Marvin Max Maximilian Nicklas Nicolai
Nils Olaf Olav Oskar Ove Paul Rafael Ragnar Ravn Robin
Sam Samuel Silas Sigurd Stig Tage Theis Theodor Thor Tue
Valentin Vincent Walter Adam Adrian Alvin Christopher Dominik Eskild Frode
Frede Hjalmar Iver Karim Anker Arvid Atle Børge Carlo Claes
Edmund Emilio Ewald Gustaf Harry James Jerry Jesse Kjartan Kresten
Magne Maurice Morgan Noa Pelle Rickard Ronnie Sander Simeon Storm
Svenn Timm Ulrich Vigo Walther Werner Wolfgang Xavier Yngve Zacharias
Bertil Bjarke Jarne Jeremy Reinhard Manfred Lorenz Melvin Kian Dion
Hasse Ingemar Ingvar Jochum Jonny Jürgen Olof Sigvald Skjold Gerhard
Hermann Eric Ernest Sune Sverre Toke Troels Tyge Jarl Fabian
Frits Gregers Halfdan Heine Henri Hakon Ejnar Elof Folmer Ketil
Isak Eske
""".split()

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
Valentin Vinther Wagner Wulff Aagaard Abrahamsen Albertsen Asmussen Axelsen Bach
Bagge Balle Beck Becker Bentsen Bille Bjerre Bjørn Bonde
Borg Borup Boysen Brink Brix Brodersen Buus Callesen Christophersen Dahlgaard
Dalgaard Daugaard Ditlevsen Dreyer Duus Eilersen Ellegaard Elmholt Engelbrecht Engholm
Esbensen Faber Fabricius Fallesen Fjord Fogh Frost Fuglsang Gad Gade
Gundersen Gustavsen Hammer Hauge Heide Helmer Hemmingsen Hoffmann Holmberg Holme
Husted Hvidt Høeg Høgh Højgaard Hørup Isaksen Jansen Jonassen Justesen
Jæger Kjeldgaard Klitgaard Koch Kragh Krarup Kristoffersen Kruuse Lauersen Lorenzen
Lundgaard Lykkegaard Magnussen Mathisen Meyer Neumann Nedergaard Nordahl Ottosen Rosendahl
Rosenkilde Roed Skaarup Stenbæk Stephansen Thaysen Tønnesen Trolle Vang
Lehmann Markussen Mouritsen Nordentoft Overby Weber Wilhelmsen Wittrup Østergaard Pagh
Pape Schwartz Schrøder Stoltenberg Sylvest Jacobi Juncker Kirk
Kvist Lerche Marcussen
""".split()


def ranked(top, recalled, what):
    """The rows (name, count) of a list: the top 20 as published, then the recalled names, each one below the one before."""
    names = list(top) + recalled
    dupes = sorted({n for n in names if names.count(n) > 1})
    if dupes:
        sys.exit(f"{what}: listed twice: {', '.join(dupes)}")
    for name in names:
        if not tsv.NAME.match(name):
            sys.exit(f"{what}: {name!r} is not a name in letters")
    twentieth = min(top.values())
    rows = list(top.items())
    for rank, name in enumerate(recalled, start=len(top) + 1):
        rows.append((name, twentieth * len(top) // rank))
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
