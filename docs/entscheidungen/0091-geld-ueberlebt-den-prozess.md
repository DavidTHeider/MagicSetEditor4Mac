# 0091 — Geld überlebt den Prozess

**Stand:** entschieden · **Datum:** 2026-09-10 · **Werkstück:** O12 von Phase 5.2
(#1439)

## Worum es geht

Das Issue verlangt: „Parallele API-/KI-Aufträge über alle lokalen Prozesse
halten freigegebene Kostenobergrenzen ein" — und sagt dazu ausdrücklich:
„lokale In-Process-Zähler genügen nicht für parallel gestartete
CLI/MCP/REST/Wails-Prozesse."

## Die Entscheidung

**Ein Journal im Bestand, und die Reservierung ist eine Transaktion.**

O5 hat **Guthaben je Prozess** gebaut (`service.Kontingente`): Rufe und
Parallelität, im Arbeitsspeicher. Das ist für Rufe richtig — sie kosten
Rechenzeit, und die gehört diesem Prozess. **Für Geld ist es falsch**: zwei
gleichzeitig gestartete Fenster, ein Befehl und ein Dienst geben aus demselben
Topf aus, und keiner der vier Zähler kennt die anderen drei.

Die Transaktion kann es: sie zählt zusammen, was gebunden ist, vergleicht mit
der Grenze und schreibt den Posten **im selben Zug**. SQLite serialisiert die
Schreiber; wer als Zweiter kommt, sieht den Ersten.

## Vier Entscheidungen darin

### 1. Vier Stände, und `unklar` bindet

| Stand | Heißt |
| --- | --- |
| `reserviert` | vor dem Ruf gebucht, Ausgang offen |
| `abgerechnet` | der Anbieter hat geantwortet, der Betrag steht |
| `freigegeben` | nichts gekostet — der Ruf scheiterte **vor** jeder Annahme |
| `unklar` | die Frist lief **nach** der möglichen Annahme ab |

**Der vierte ist der Grund für das Ganze.** Ein Auftrag, dessen Frist nach der
möglichen Annahme abläuft, ist nicht „nichts passiert". Er bindet den **ganzen
Höchstbetrag**, bis ihn jemand klärt — „unbekannte Kosten nicht als kostenlos
behandeln" verlangt die Abnahme wörtlich.

**Der Preis**: ein Anbieter, der nach einer Frist doch nichts berechnet, hält
Budget fest, bis ein Mensch nachsieht. Das ist die sichere Richtung.

### 2. Die Identität kommt vom Aufrufer

`Reservieren` mit einer Kennung, die schon einen Posten hat, gibt **den
vorhandenen** zurück — auch dann, wenn ein zweiter die Grenze sprengte.

**Das ist der Schutz gegen den zweiten kostenpflichtigen Auftrag**: ein Abbruch
nach dem Ruf, ein Neustart, ein Wiederholungsversuch — alle drei zeigen
dieselbe Identität vor, und alle drei bekommen dieselbe Buchung.

**Verworfen**: eine laufende Nummer aus dem Bestand. Sie wäre nach einem
Absturz nicht wiederzufinden; die Identität muss von dem kommen, der den
Auftrag **wiederholt**.

### 3. Keine Währungsumrechnung

Zwei Beträge verschiedener Währungen lassen sich hier nicht addieren,
vergleichen oder abziehen — jeder Versuch ist ein Fehler.

**Der Grund steht im Issue**: „Keine Wechselkursumrechnung ohne eigene Quelle."
Diese Anwendung hat keine. Ein Kurs von irgendwoher wäre eine Zahl, für die
niemand einsteht — und aus ihr entstünde eine Grenze, die jemand für geprüft
hält.

**Der Preis**: wer ein Konto in Dollar und eines in Euro führt, hat zwei
Budgets und keine Gesamtsumme. Das ist ehrlich.

### 4. Die Grenze ist die Freigabe — und wird nicht über MCP gesetzt

Wer die Grenze selbst höhersetzen könnte, hätte keine. Das ist dasselbe
`bestaetigt`-Feld, das kein Nachweis sein soll
([0078](0078-der-wirksame-akteur.md)).

**Der Stand ist trotzdem lesbar**, auch über MCP: ein Modell, das gegen eine
Wand läuft, die es nicht sieht, kann mit der Absage nichts anfangen — eines,
das vorher weiß, wie viel übrig ist, teilt sich ein.

## Und der Nachweis, den O1 gebaut hat

`Reservieren` verlangt einen [Nachweis](0078-der-wirksame-akteur.md), und er
trägt die **Kosten** als Bindung: `Kostengrenze.deckt` hält den Höchstbetrag
gegen das, was ein Mensch freigegeben hat. Diese Felder standen seit O1 in
`domain.Nachweisbindung` und hatten keinen Benutzer; O12 ist der erste.

**Die Abschlüsse verlangen keinen zweiten.** Er galt dem Auftrag, nicht der
Buchhaltung; ihn dort noch einmal zu verlangen hieße, einen Menschen zu etwas
zustimmen zu lassen, das schon geschehen ist.

## Was das Journal **nicht** kann, und es steht an jeder Fläche

**Fremde Nutzung desselben Schlüssels.** Wer denselben API-Zugang woanders
benutzt, verbraucht Kontingent, von dem dieses Journal nichts weiß. Eine harte
accountweite Zusage verlangt eine **Anbietersperre**; die gibt es hier nicht,
und deshalb heißt die Grenze eine **lokale** Grenze — in der Hilfe, im
Werkzeugtext, auf der Seite.

## Und es gibt heute keinen zahlenden Verbraucher

**Dieselbe Lage wie bei der Epoche aus O7** ([0088](0088-die-epoche-einer-rueckkehr.md)):
Scryfall, GitHub und die drei Kartendienste kosten nichts. Der Weg steht,
damit Phase 5.6 ihn benutzt, statt ihn zu erfinden — und damit er dann
**geprüft** ist und nicht neu.

**Was das kostet, gehört ausgesprochen**: eine Buchführung ohne Buchungen ist
schwerer zu prüfen als eine mit. Was sie hält, hält sie gegen eine echte
Datenbank — nicht gegen einen echten Anbieter.
