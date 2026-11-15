# 0189 — Nur die Grundlinie, schon in 0.8.0

**Stand:** angenommen am 2026-10-12 ·
#2266 (J02).
**Entschieden vom Auftraggeber**: „es soll nur die Grundlinie geben; auch in
0.8.0" — gegen den Vorschlag dieses Auftrags, die Kette der Wanderungen bis
1.0.0 stehen zu lassen. **Schränkt
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) ein**, Satz 1
„vorwärts, immer".

## Die Frage

J02 verlangt, dass ein frischer Bestand schneller entsteht: jeder Start eines
neuen Bestands und fast jeder Test fuhr alle 97 Wanderungen einzeln, darunter
Umbauten ganzer Tabellen, und unter dem Race-Detektor kostete das 2,28 s statt
0,23 s. Die Grundlinie — das Schema in einem Zug — löst das für den **leeren**
Bestand. Offen war, was mit den 97 Dateien geschieht:

| Weg | Was er kostet |
| --- | --- |
| **Grundlinie und Kette** — ein neuer Bestand nimmt die Grundlinie, ein alter wandert die Kette | zwei Wege zum selben Schema, die ein Test gegeneinander halten muss; 97 Dateien, die niemand mehr ändern darf und jeder weiter mitliest |
| **Nur die Grundlinie** — die Kette fällt | ein Bestand von **vor** 0.7.0 lässt sich hier nicht mehr nachziehen |

Wir sind vor dem Launch-Release 1.0.0; die Zahl der Bestände von vor 0.7.0,
die es außerhalb dieses Rechners gibt, ist klein. Der Auftraggeber hat den
zweiten Weg gewählt, und zwar **schon für 0.8.0**.

## Entscheidung

1. **Es gibt nur die Grundlinie** bis `0097_faktorhaushalt.sql`
   (`internal/storage/grundlinie/`). Die 97 Wanderungen liegen nicht mehr als
   Datei vor; `grundlinie/umfang.txt` trägt jede mit Name und Prüfsumme —
   **dieselben, die 0.7.0 ausgeliefert hat**, verglichen am 2026-10-12 für
   alle 97 gegen den Tag `v0.7.0`.
2. **Ein neuer Bestand entsteht in einem Zug** und verbucht alle 97 mit ihrer
   Prüfsumme — dieselbe Buchführung, die einer hinterlässt, der sie einzeln
   gegangen ist.
3. **Ein Bestand aus 0.7.0 öffnet sich ohne einen Schritt.** Er führt genau
   diese 97.
4. **Ein Bestand von vor 0.7.0 wird mit Rat abgewiesen** — `bestandAeltereFassung`
   am Start, `sicherungAeltereFassung` beim Zurückspielen — und **nicht
   angefasst**: die Abweisung steht vor jeder Schemaänderung, auch vor dem
   Nachtragen einer Prüfsummenspalte in einer alten Buchführung. Der Rat:
   einmal mit 0.7.0 öffnen, dort steigt er auf, danach öffnet ihn diese
   Fassung.
5. **Tabellen ohne Buchführung sind kein Bestand dieser Anwendung** und
   werden abgewiesen. Bis 0.7.0 scheiterte ein solcher an der ersten
   Wanderung, die eine seiner Tabellen noch einmal anlegen wollte; ohne die
   Abweisung bliebe er jetzt liegen.
6. **Was nach der Grundlinie kommt, ist wieder eine Wanderung als Datei**
   unter `internal/storage/migrationen/`, ab `0098`, und läuft wie bisher:
   Schritt für Schritt, in einer Transaktion, mit Sicherung davor.
7. **Das Freigabemanifest nennt die Grundlinie** (`"grundlinie"`), und die
   Prüfung vor einem Update kennt das Hindernis `vorDerGrundlinie`: eine
   künftige Fassung, die sie weiterschiebt, sagt so vorab, welcher Bestand
   sie nicht öffnet.
8. **Die Grundlinie ist eingefroren.** Eine Änderung am Schema ist eine neue
   Wanderung, nie eine Änderung an ihr — `TestDieGrundlinieIstUnveraendert`
   hält einen Abdruck über alle ihre Dateien.

## Was das gegenüber 0033 heißt

0033 sagt in Satz 1: ein Bestand einer älteren Fassung wird beim Start
migriert — **vorwärts, immer**. Das gilt ab jetzt für jeden Bestand **ab
0.7.0**. Für einen älteren gilt es über eine Zwischenstation: 0.7.0 bleibt
beziehbar (`v0.7.0` auf GitHub) und kennt jede Wanderung bis zur Grundlinie.
Die übrigen Sätze von 0033 — keine Rückwärtswanderung, Sicherung davor,
Abweisung eines neueren Bestands — bleiben, wie sie sind.

## Belegt durch

- **Gleiches Schema**: ein frischer Bestand aus dem heutigen Stand und einer,
  den 0.7.0 selbst angelegt hat, tragen dasselbe `sqlite_master` (1.910
  Zeilen, verglichen am 2026-10-12) und dieselbe Buchführung, Name und
  Prüfsumme. Bis zum Wegfall der Kette hielt
  `TestDieGrundlinieGleichtDenWanderungen` dasselbe gegen die Kette
  (Commit `ebd92fa37`).
- **0.7.0 → hier**: `TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet`
  an einer Sicherung, die die Kommandozeile aus `v0.7.0` geschrieben hat
  (`testdaten/fassung-0.7.0/`).
- **Davor → Abweisung**: `TestEineSicherungAusNullFuenfWirdMitRatAbgewiesen`
  an der Sicherung aus 0.5.0, `TestEinBestandVorDerGrundlinieWirdMitRatAbgewiesen`
  und `TestEinBestandOhnePruefsummenspalteBleibtUnberuehrt`.
- **Danach**: `TestEineKuenftigeWanderungLaeuftAufDerGrundlinie` mit einer
  Probewanderung `0098`, über einem neuen Bestand wie über einem aus 0.7.0,
  und `TestDieSicherungSiehtDenStandVorDerWanderung`.

## Was wegfiel

**Die Tests einzelner Datenwanderungen an alten Beständen** — `0004`, `0022`,
`0033`, `0050`, `0051`, `0070`, `0074`, `0078`, `0079`, `0080`, `0086`,
`0097`, der Weg von den ersten drei Schritten bis heute und der Aufstieg aus
0.4.0 und 0.5.0. Sie belegten, dass ein Schritt Daten aus einem älteren
Schema mitnimmt; einen solchen Schritt gibt es hier nicht mehr. **Die Tests
der Maschine** — Abbruch mitten in der Folge, Sicherung davor, Abweisung
eines neueren Bestands, Prüfsumme — laufen weiter, gegen erfundene Quellen
oder die Probewanderung `0098`.

## Der Preis

- **Eine Fassung vor 0.8.0 kennt die Grundlinie nicht.** Ihre Prüfung vor
  einem Update hält die Freigabe 0.8.0 für einen Bestand aus 0.6.0 für
  verträglich; erst der Start von 0.8.0 weist ihn ab — ohne Schema oder Inhalt
  anzufassen.

- **Wer einen Bestand von vor 0.7.0 hat, braucht einen Zwischenschritt.**
  Die Anwendung sagt ihn, aber sie nimmt ihn nicht ab.
- **Die Wanderungen bis 0097 sind als Quelltext nur noch in der Geschichte**
  des Repositoriums — `git show v0.7.0:internal/storage/migrationen/…`. Was
  ein Schritt mit den Daten tat, steht dort und in den Vermerken, die ihn
  begründen.

## Verworfen

**Die Kette bis 1.0.0 stehen lassen** — der Vorschlag dieses Auftrags. Er
hätte den Zwischenschritt erspart und zwei Wege zum selben Schema erhalten;
der Auftraggeber hat dagegen entschieden, weil vor 1.0.0 die Zahl der
betroffenen Bestände am kleinsten ist.
