# Zusatzdaten

**Was dieses Programm über ein Set hinaus weiß — Preise, Exemplare, Marken,
Herkunft — und wie es damit umgeht, wenn etwas hinausgeht.** FMT-05
(#1694).

## 1 Wo sie leben

**Im Bestand, und nur dort.** Eingebettet in die `.mse-set`-Datei gingen sie
in MSE2 verloren und machten jedes Öffnen dort laut; eine Datei daneben reiste
nicht mit — beides am echten MSE2 gemessen ([0145](entscheidungen/0145-zusatzdaten-werden-nicht-eingebettet.md),
Beleg [`belege/mse2-zusatzdaten.md`](belege/mse2-zusatzdaten.md)).

## 2 Ein Preis ist eine Beobachtung

Jeder Preisstand (`domain.Preisstand`, [0093](entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md))
trägt **Betrag** (`Wert`, in der kleinsten Einheit der Währung), **Währung**, **Quelle** samt Markt
und Art (Verkauf, Ankauf), **Kartendruck** (`Printing`), **Ausführung**,
**Zustand** und **Sprache**, wo die Quelle sie nennt, und **zwei Zeiten**: wann
die Quelle ihn beobachtet hat und wann er hier ankam. Daraus rechnet die
Oberfläche sein Alter.

**Ein beobachteter Preis ist kein Handelswert.** Die Hilfe sagt es unter
*Sammlung* („Schätzwerte, für die niemand einsteht"), und kein Weg dieses
Programms bietet auf ihm etwas an oder kauft etwas.

## 3 Was unbekannt ist, bleibt

**Ein Feld, das dieses Programm nicht kennt, übersteht Öffnen und Sichern
unverändert** — der Leser führt es als Rohblock mit, und der Rundlauf ist
byte-genau (`TestEinDurchgereichtesSetIstByteFuerByteDasselbe` mit
`testdaten/eigenwillig.mse-set` und seinem `unverstandener block:`). MSE2 tut
das nicht (§1).

**Schemafassungen** stehen mit Aufstieg, Abstieg und Beleg in
[`FASSUNGEN.md`](FASSUNGEN.md): der Bestand mit Prüfsumme je Wanderung, der
Umschlag (`austausch.Fassung`), der Druckverweis der Sammlungs-CSV
(`domain.Verweisfassung`), Partie- und Turnierdatei.

## 4 Was hinausgeht, und was hier bleibt

| Weg | Geht mit | Bleibt hier | Wo es gesagt wird |
| --- | --- | --- | --- |
| **Set sichern** (`.mse-set`) | die Kartendefinitionen, wie MSE2 sie liest | Decks, Exemplare, Preise, Marken, Herkunft einzelner Werte | *Ausgeben → Als Setdatei sichern*, [`DATEIOPERATIONEN.md`](DATEIOPERATIONEN.md) |
| **Deck als Umschlag** | Name, Spiel, jeder Platz auf einen Druck mit Menge und Bereich | **die Notiz, die Herkunft** (sie nennt Quellen und Profilnamen anderer), Plätze auf eigene Karten, Zuordnungen, Format | *Decks → Als Umschlag weitergeben*; ein Lauf prüft den **Inhalt** des Umschlags (`TestEinUmschlagTraegtWederNotizNochHerkunft`) |
| **Sammlungssicht als Umschlag** | nur, was ein Gegenüber sehen darf (Freigabe, T7) | alles andere | die Freigabe selbst |
| **Sammlung als CSV** | jede Angabe jedes Exemplars, **auch Erwerbspreis und Notiz** — es ist die eigene Ausfuhr, keine Weitergabe | Zuordnungen, Wünsche | die Seite der Ausfuhr |
| **Eigene Daten** (Personendaten) | alles, was dieses Programm über dich führt — auf deinen Wunsch | — | *Nutzer und Profile* |

**Ungefragt geht nichts Privates an andere.** Bis FMT-05 trug ein
Deck-Umschlag seine Notiz und seine Herkunft mit; beides bleibt jetzt hier.
Die Sammlungs-CSV trägt Preise und Notizen **mit Absicht**: sie ist die
Sicherung der eigenen Sammlung und kein Weg zu einem Gegenüber.
