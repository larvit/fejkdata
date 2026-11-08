#!/usr/bin/env python3
"""Rebuild data/da_DK/first-name.tsv and last-name.tsv from Statistics Denmark's top 20 and a language model's recall.

    data-import/names-dk.py [--out DIR]

No open source holds Danish name counts in bulk. Statistics Denmark (CC BY 4.0, "Kilde: Danmarks Statistik")
publishes the top 20 girls' names, boys' names and last names of the people living in Denmark on
1 January 2026 ("Names in the entire population", dst.dk/en/Statistik/emner/borgere/navne/navne-i-hele-befolkningen)
and a count one name at a time, and Wikidata's Danish names miss most of the common ones. So:

- the top 20 of each list keep Statistics Denmark's counts;
- every other name is a language model's recall of a name common in Denmark, listed from the most to
  the least common, with an estimated count that falls as one over its rank below the twentieth name's.

Those counts are estimates of the order of the names, not published figures. A name that girls and
boys both bear has a row under each sex.
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
Anette Karin Grethe Dorthe Ruth Bodil Merete Anja Pernille Sofie
Katrine Line Trine Sara Mia Maja Laura Cecilie Johanne Josefine
Freja Clara Nanna Ellen Signe Sine Stine Mie Kamilla Sanne
Lisa Lærke Emilie Amalie Frederikke Caroline Isabella Sofia Nicoline Mathilde
Alberte Astrid Agnes Alma Olivia Victoria Marie Christina Kristine Katja
Michelle Malene Heidi Tanja Helene Henriette Winnie Hanna Jeanette Jane
Lena Lillian Elisabeth Else Margit Jonna Gerda Aase Agnete Alice
Birte Conny Doris Gunhild Hedvig Iben Irene Janni Jacqueline Joan
Kathrine Kirstine Lea Lilian Liselotte Majbritt Maiken Mary Mona Nina
Rie Rose Sidsel Stephanie Sussie Ursula Vivi Yvonne Ella Luna
Liv Filippa Selma Sally Naja Nora Vilma Esther Silje Thea
Tilde Karoline Cathrine Lykke Elna Edith Hanni Helga Gudrun Inger
Margrethe Hannah Jenny Susan Kim Ditte Lizette Mai Marlene Mikala
Sonja Solvej Sabrina Ronja Rebekka Pauline Patricia Natasja Nadia Mille
Maj Mari Lina Lilly Lise Leonora Laila Kristina Kitte Karla
Karina Julia Josephine Johanna Jessica Jasmin Ina Ilse Hilda Henny
Helena Gry Gunvor Grete Gertrud Gabriella Fie Fanny Evelyn Eline
Elise Elsebeth Emmely Emelie Elin Dorthea Dina Dagmar Cornelia Connie
Christel Cecilia Celina Bettina Betina Bertha Benedikte Beate Ane Anita
Angelina Andrea Anastasia Amanda Alexandra Alva Alicia Agathe Adelina Abelone
Gunna Hedda Ingeborg Ingelise Jannie Joanna Judith Jutta Kaja Kamille
Kate Katharina Klara Kristiane Laerke Leila Lisette Lola Louisa Lucia
Madeleine Magdalene Majken Malou Martha Marta Matilde Melanie Melissa Michaela
Mira Monica Naomi Nicole Olga Oline Paula Petra Rachel Ragnhild
Regina Rigmor Rina Rosa Rosalie Sandra Sidse Sigrid Simone Sophie
Stella Stina Tamara Teresa Thilde Tia Vera Veronica Viola Viktoria
Wilma Zara Alexia Alexine Amelia Anine Annelise Annelie Annemarie Annemette
Anni Annika Annie Aurora Bella Bianca Birgitte Birgitta Bolette Brigitte
Britt Camille Carina Carla Carmen Carolina Cassandra Dorrit Eleonora Elizabeth
Ellinor Elvira Emilia Erica Erika Estrid Felicia Fiona Frida Gerd
Gina Iris Isabel Isabelle Alexa Alina Allis Amy Andrine Angelika
Aya Cleo Dagny Diana Edda Eleonore Elena Elsa Emmeline Erna
Ester Fabiana Hilde Hulda Ilsa Isa Jana Janne Jasmine Jeanne
Jill Jorun Juliane Kamma Karolina Lajla Laurine Leonie Lillie Linda
Lotus Lulu Maia Malin Mariam Marianna Maryam Mathea Maya Meta
Minna Molly Mynte Nadja Nathalie Nikita Ninna Pil Rakel Rasmine
Rita Rosita Roskva Sabine Salma Sania Sarah Sasha Saga Sia
Siri Sonia Susanna Svea Tanya Tara Tessa Theodora Therese Tilda
Tonja Tuva Una Vanessa Vicky Vivian Wilhelmine Wibeke Zaynab Zoe
Zofia Agnethe Aina Alberthe Alette Anny Ashley Bibi Bonnie Brita
Cille Dicte Dorit Eira Elva Emely Fatma Gudny Helen Hermine
Ines Isolde Jacoba Jolanda Kristin Lilli Linnea Lucca Lucy Madelene
Nelly Nynne Petrine
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
Ulf Vagn Villy Vilhelm Axel Aksel Alex Alf Alan Aleksander
Ali Asger Asbjørn Aske Bastian Bendt Birger Bjørn Bobby Bruno
Charles Christoffer David Edvard Edward Emanuel Ernst Esben Felix Gabriel
Georg Harald Hjalte Hugo Jeppe Joe Jon Josef Joseph Julian
Julius Karl Kevin Kristoffer Lauge Laurits Leo Leon Leonard Levi
Liam Louis Luca Ludvig Lukas Marco Mario Markus Marvin Mathis
Matias Max Maximilian Mehmet Mikael Mohamed Mohammad Nick Nicklas Nicolai
Nils Olaf Olav Omar Oskar Otto Ove Pascal Paul Peder
Phillip Rafael Ragnar Ravn Regnar Rikard Robert Robin Roland Rolf
Rudolf Sam Samuel Silas Sigurd Sofus Stig Tage Theis Theodor
Thor Thorvald Tim Torsten Tue Valentin Vincent Walter Yusuf Adam
Adrian Alvin Amir Christopher Dominik Jannik Karsten Klaus Kaj Kåre
Joakim Joachim Jasper Jack Isak Gert Erling Ejnar Fabian Frits
Gorm Gregers Halfdan Heine Henri Hakon Sune Sverre Toke Troels
Tyge Jarl Eric Ernest Frode Frede Hjalmar Iver Karim Anker
Arvid Atle Børge Carlo Claes Edmund Emilio Eskild Ewald Gustaf
Harry James Jerry Jesse Kjartan Kresten Magne Maurice Morgan Noa
Pelle Rickard Ronnie Sander Simeon Storm Svenn Timm Ulrich Vigo
Walther Werner Wolfgang Xavier Yngve Zacharias Bertil Bjarke Jarne Jeremy
Reinhard Manfred Lorenz Melvin Kian Dion Hasse Ingemar Ingvar Jochum
Jonny Jürgen Olof Sigvald Skjold Gerhard Hermann Ahmad Hussein Bashir
Ronni Ebbe Eigil Elof Emmanuel Erland Folmer Freddy Henny Ketil
Laust Nikolas Oluf Orla Ricky Sigfred Tonny Torkil Villum
""".split()

LAST_RECALLED = """
Frederiksen Jakobsen Mikkelsen Holm Lund Schmidt Eriksen Kristiansen Henriksen Simonsen
Clausen Andreasen Iversen Jeppesen Mogensen Jacobsen Olesen Laursen Frandsen Knudsen
Lauridsen Bertelsen Berg Svendsen Johannesen Thomassen Hermansen Nissen Hjort Dam
Bruun Winther Lassen Skov Dahl Bech Kjær Kjeldsen Lauritsen Therkildsen
Thorsen Vestergaard Lindberg Lorentzen Michelsen Nørgaard Nyholm Rosenberg Schou Toft
Overgaard Bøgh Brandt Bak Bruhn Christoffersen Damgaard Ebbesen Eliasen Enevoldsen
Fischer Friis Gregersen Grønbæk Hald Hansson Hedegaard Hjorth Holst Jessen
Jochumsen Juhl Kaas Karlsen Kofoed Krogh Kruse Lange Leth Lindholm
Mathiesen Nygaard Nørregaard Paulsen Rask Riis Rohde Sand Schultz Skovgaard
Steffensen Strøm Sundby Svane Sønderby Thygesen Ulrich Vang Villadsen Winter
Ravn Bendtsen Bjerregaard Bundgaard Carlsen Dalsgaard Ejlersen Falk Foged Gram
Guldberg Hartmann Haugaard Hviid Højer Ibsen Jepsen Juul Klausen Lindgren
Lykke Munk Neergaard Nørgård Ohlsen Pallesen Quist Rømer Storgaard Thorup
Valentin Vinther Wagner Wulff Aagaard Abrahamsen Albertsen Arildsen Asmussen Axelsen
Bach Bagge Balle Beck Becker Bengtsson Bentsen Bille Bjerre Bjørn
Bonde Borg Borup Boysen Brink Brix Brodersen Buus Callesen Christophersen
Dahlgaard Dalgaard Daugaard Ditlevsen Dreyer Duus Eilersen Ellegaard Elmholt Engelbrecht
Engholm Esbensen Faber Fabricius Fallesen Fjord Fogh Frost Fuglsang Gad
Gade Gundersen Gustavsen Hammer Hauge Heide Helmer Hemmingsen Hoffmann Holmberg
Holme Hvidt Høeg Høgh Højgaard Hørup Isaksen Jansen Jarl Jonassen
Justesen Jæger Kjeldgaard Klitgaard Koch Kragh Krarup Kristoffersen Kruuse Lorenzen
Lundgaard Lykkegaard Magnussen Mathisen Meyer Neumann Nedergaard Ottosen Schrøder Tang
Vestergård Weber Wilhelmsen Østergaard Ørsted
""".split()


def ranked(top, recalled, what):
    """The rows (name, count) of a list: the top 20 as published, then the recalled names, each one below the one before."""
    names = list(top) + recalled
    seen = {}
    for name in names:
        if name in seen:
            seen[name] += 1
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
