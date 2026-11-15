# 0016 — Keine Unicode-Segmentierung als Abhängigkeit

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

#113, zweiter
Teil: `breiteVon` in `internal/render/satz.go` zählt **Runen**. `e` gefolgt von
U+0301 ist ein Buchstabe aus zwei Runen; die Laufweite kam dazwischen, und die
Marke saß neben ihrem Buchstaben statt über ihm.

Das Issue nennt den Weg gleich mit:

> Graphemcluster richtig zu zerlegen heißt, die Unicode-Segmentierung
> mitzubringen. Go hat sie nicht in der Standardbibliothek;
> `golang.org/x/text` hätte sie, und das ist dieselbe
> Abhängigkeitsentscheidung wie in #105.

## Was beim Nachschlagen herauskam

**`golang.org/x/text` führt keine öffentliche Zerlegung nach UAX #29.**
Nachgeschlagen über Context7 (`AGENTS.md` §1), nicht aus dem Gedächtnis: was
dort dokumentiert ist, sind **Wort**grenzen innerhalb von `cases` — als
Näherung für die Titelschreibung — und `unicode/norm`, das an
**Normalisierungs**grenzen trennt. Beides ist nicht dasselbe wie ein
Graphemcluster.

Die verbreitete Umsetzung ist `github.com/rivo/uniseg`. Das ist ein Paket
**außerhalb** der `x`-Familie, und damit ist es nicht dieselbe Entscheidung wie
[0008](0008-schriftrasterung.md): dort ging es um eine Bibliothek vom Go-Team,
BSD-3, mit demselben Freigaberhythmus wie Go selbst.

**Die Prämisse des Issues trägt also nicht.** Das ist der eigentliche Fund, und
er gehört aufgeschrieben, damit ihn niemand ein zweites Mal sucht.

## Entscheidung

**Keine neue Abhängigkeit. Die Laufweite überspringt nicht-abstandhaltende
Marken über `unicode.Mn` und `unicode.Me` aus der Standardbibliothek.**

Das löst den Fehler, der gemessen wurde: 0,2 Punkte je kombinierendem Zeichen
in `magic-m15`, wachsend mit jeder Sprache, die Diakritika stapelt —
Vietnamesisch legt zwei auf einen Vokal.

## Was das nicht abdeckt

Und es gehört genannt, nicht verschwiegen:

| Fall | Wird gezählt als |
| --- | --- |
| Emoji mit Nullbreiten-Verbinder (`👨‍👩‍👧`) | drei Buchstaben statt einem |
| Regionalindikatoren (`🇩🇪`) | zwei statt einem |
| Hangul-Jamo, einzeln geschrieben | so viele, wie dastehen |
| Abstandhaltende Marken (`Mc`, etwa Devanagari) | ein eigener Buchstabe — was hier **richtig** ist |

Der erste Fall ist der einzige, der auf einer Karte vorkommen dürfte, und sein
Preis ist eine Laufweite zu viel in einer Zeile, die ohnehin nur ein
Kartentext-Emoji trägt.

## Warum das der bessere Tausch ist

Eine Abhängigkeit kostet in diesem Projekt: eine Zeile in `DRITTANBIETER.md`,
eine Pinnung, einen Aktualisierungspfad, eine Prüfung ihrer Lizenz und ihres
Freigaberhythmus. `rivo/uniseg` ist gut gepflegt und trotzdem ein Paket einer
Einzelperson.

**Für 0,2 Punkte je kombinierendem Zeichen ist das zu viel** — und für die
Fälle, die es zusätzlich löste, erst recht: sie kommen auf einer Magic-Karte
nicht vor.

## Wann das umzuentscheiden ist

Wenn eine Vorlage auftaucht, die Emoji-Folgen oder Hangul im Kartentext
benutzt, **und** jemand den Unterschied sieht. Dann ist die Frage nicht mehr
„lohnt eine Abhängigkeit für 0,2 Punkte", sondern „welche" — und dann gehört
`rivo/uniseg` gegen eine eigene Umsetzung von UAX #29 abgewogen, nicht gegen
das Nichtstun.

## Wer das ändern will

ändert zuerst diesen Vermerk — und misst vorher, wie groß der Unterschied an
einer echten Karte ist. Der erste Anlauf hier hat mit einer Zahl angefangen
(0,2 Punkte), und das war der Grund, warum die Antwort klein ausfiel.
