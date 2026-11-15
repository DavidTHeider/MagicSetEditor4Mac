# 0104 — Der Stand einer Karte ist eine Antwort, keine Vorbedingung

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A16 von Phase 5.3
(#1515)

---

## Die Frage

Das Issue verlangt für die Ähnlichkeitsfrage einen „autorisierten
Kartensnapshot": gesucht wird zu einer Karte **in einem bestimmten Stand**.
Ändert sie sich zwischen der Frage und der zweiten Seite, gehört die Antwort
zu einer anderen Karte — das Issue nennt das die „still gemischte Seite" und
verbietet sie.

**Wer nennt diesen Stand?**

---

## Warum die naheliegende Antwort nicht geht

Die erste Fassung von `domain.Rechercheanfrage.Pruefe` verlangte ihn:

> **Ohne Revision keine Ähnlichkeit.** Eine Karte, die sich seit der Anfrage
> geändert hat, ist eine andere Frage.

Der Satz ist richtig. Die Regel war trotzdem falsch, und zwar an allen vier
Flächen gleichzeitig: **keine von ihnen kann den Stand nennen.**

Der Stand einer Karte ist kein gespeichertes Feld, sondern eine **Rechnung**
über ihre Werte ([`domain.CardRow.Stand`], eingeführt mit A6). Er entsteht im
Bestand. Das Fenster bekommt `KarteDTO` mit `id`, `nummer` und `werte` — und
keinen Stand; A6 liest ihn serverseitig über `kikarten.Revision` und reicht
ihn nie heraus. Kommandozeile, MCP und Dienst sehen ihn ebenso wenig.

**Damit hätte jede Ähnlichkeitsfrage eine Validierungsabsage bekommen** — an
jeder Fläche, bei jedem Aufruf. Die Fähigkeit wäre vorhanden gewesen, mit
Tests und voller Abdeckung, und **niemand hätte sie benutzen können**. Das ist
die erste Dimension aus §10, und sie steht dort zuerst, weil sie schon einmal
gefehlt hat.

---

## Die Entscheidung

**Ein leerer Stand ist kein fehlender, sondern „der, auf dem die Karte jetzt
steht" — und die Antwort nennt ihn.**

Drei Teile:

1. **`Ausgangsquelle.Ausgangstext` gibt den gerechneten Stand zurück**, auch
   wenn keiner hineinging. Sie liest die Karte ohnehin; sie ist die einzige
   Stelle, die ihn kennt.
2. **Der Dienst schreibt ihn in die Anfrage, bevor der Fortsetzungszeiger
   entsteht.** Nur so bindet der Zeiger einen wirklichen Stand: wer die zweite
   Seite holt, gibt ihn zurück, und eine Karte, die sich dazwischen geändert
   hat, gibt eine **Absage** statt einer Seite aus einer Liste, die es nicht
   mehr gibt.
3. **`Rechercheergebnis.Ausgangsstand` trägt ihn ans Fenster** — und von dort
   in die DTOs aller vier Flächen.

**Wer einen Stand nennt, wird weiterhin geprüft.** Stimmt er nicht mehr, ist
das ein `ErrConflict` mit beiden Ständen im Satz. Das ist die Zusage, um die
es ging; verloren geht nur die Pflicht, sie beim ersten Ruf schon einzulösen.

---

## Was das kostet

**Zwischen dem Lesen der Karte und der Suche liegt ein Zeitraum**, in dem sich
die Karte ändern kann. Das Ergebnis gehört dann zum Stand **beim Lesen** und
nicht zum Stand beim Abschicken. Das ist keine Lücke der Prüfung, sondern die
Reichweite einer Momentaufnahme: welchen Stand die Antwort meint, steht in ihr
— und mehr kann eine Antwort über eine Karte, die jemand anders gerade tippt,
nicht sagen.

**Und die erste Seite ist ungeprüft.** Sie kann es sein, denn sie **definiert**
den Stand; geprüft wird ab der zweiten, und dort liegt der Fall, um den es dem
Issue geht.

---

## Verworfen

**Eine Bindung `Kartenstand(karteID)`.** Sie hätte den Stand herausgegeben,
damit die Flächen ihn nennen können — und damit ein Wettrennen eingebaut: der
Stand vom vorigen Ruf gälte, und zwischen den beiden Rufen läge derselbe
Zeitraum wie oben. Ein zusätzlicher Ruf, der nichts gewinnt.

**Den Stand als Pflichtfeld lassen und die Flächen leer lassen.** Das war der
Zustand, der diesen Vermerk ausgelöst hat: eine Zusage, die niemand einlösen
kann, sieht aus wie Strenge und ist Unbenutzbarkeit.

---

## Gegenprobe

`internal/service/recherchedienst_test.go` führt zwei Läufe, und beide werden
rot, sobald die Zuweisung `anfrage.Ausgangsstand = ausgangsstand` zurückgenommen
wird:

- **`TestEinLeererAusgangsstandWirdZumJetzigen`** — die Antwort nennt `""`.
- **`TestDerZeigerBindetDenGerechnetenAusgangsstand`** — der Zeiger überlebt
  einen Standwechsel, und die zweite Seite gehört zu einer Liste von vorhin.

Der zweite ist der eigentliche: der erste sagt nur, dass eine Zeichenkette
zurückkommt.
