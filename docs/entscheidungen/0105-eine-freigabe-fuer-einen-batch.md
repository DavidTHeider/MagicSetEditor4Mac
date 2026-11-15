# 0105 — Eine Freigabe gilt einem Batch, nicht jedem Auftrag darin

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Anlass:** Werkstück A13.1 von Phase 5.3
(#1452)

## Der Widerspruch

Seit #1539
verlangt jeder kostenpflichtige KI-Auftrag den Nachweis eines Menschen, und
`service.Nachweise.Einloesen` **verbraucht** ihn. Der Satz aus
[0078](0078-der-wirksame-akteur.md) lautet: **ein Nachweis wird einmal
eingelöst**, und ein zweiter Auftrag braucht einen zweiten.

Ein Batchvorschlag über zwölf Karten schickt zwölf Anfragen hinaus. Nach dem
Buchstaben wären das zwölf Freigaben.

## Was verworfen wurde

**Zwölf Freigaben.** Sie halten den Buchstaben und zerstören den Zweck: eine
Frage, die man zwölfmal hintereinander beantwortet, ist keine Frage mehr. Wer
sie stellt, erzieht zum Wegklicken — und dann gilt sie auch dort nichts mehr,
wo sie zählt. Das ist genau der Zustand, gegen den es Nachweise gibt.

**Ein Batch ohne Freigabe**, weil „die einzelnen Aufträge ja schon geprüft
sind". Sie sind es nicht: geprüft wird beim Einlösen, und ohne Einlösen ist
nichts geprüft.

## Was gilt

**Der Nachweis lautet auf den Batch.** Sein `Ziel` ist die Kennung des
Batches, sein Betrag der, den ein Mensch für **ihn** gelesen hat. Der Batch
reserviert **einmal**; die Aufträge darin reservieren nicht noch einmal.

**Die Zusage ändert sich dadurch nicht.** Sie lautet weiterhin: ein Mensch hat
einen Betrag für eine abgegrenzte Menge Arbeit gelesen und freigegeben, und
der Nachweis ist danach verbraucht. Was sich ändert, ist der **Zuschnitt** der
Arbeit — und der Zuschnitt ist das, was am Fenster steht.

## Was daraus folgt

1. **Ein Batch hat eine eigene Kennung.** Ohne sie gäbe es kein Ziel, auf das
   der Nachweis lauten könnte.
2. **`service.KIBudget` bekommt eine Gestalt „schon gedeckt".** Die Aufträge
   im Batch buchen nicht selbst.

   **Nicht über `KostenpostenOhneKosten`**: diese Marke heißt „lokal, kostet
   nichts", und das ist hier das Gegenteil dessen, was zutrifft. Zwei
   verschiedene Sachverhalte unter einer Marke sind zwei Auskünfte, von denen
   eine falsch ist — dieselbe Falle wie bei den fünf Einrichtungsständen
   ([0097](0097-was-ueber-ein-ki-modell-feststeht.md)).
3. **Die Kostengrenze des Batches ist die des Nachweises.** Mehr als den
   freigegebenen Betrag kann er nicht kosten, gleich wie viele Ziele darin
   stehen.
4. **Und die Zahl der Ziele bleibt zusätzlich begrenzt.** Eine Freigabe über
   einen Betrag sagt nichts über die Zahl der Anfragen: ein Anbieter drosselt
   nach Anfragen und nicht nach Geld, und ein Batch über zweihundert Karten
   wäre auch dann unhöflich, wenn er billig bliebe (§16, Regel 4).

## Was das kostet

**Wer einen Batch freigibt, hat die einzelnen Aufträge nicht gelesen.** Er hat
gelesen, **wie viele** es sind, **an wen** sie gehen und **was sie höchstens
kosten** — und genau das muss am Fenster stehen, bevor gefragt wird. Steht es
nicht da, ist diese Entscheidung die Vollmacht, die sie nicht sein will.

## Was hier nicht entschieden ist

**Ob eine Gruppe aus einem Batch atomar übernommen wird.** Das ist eine Frage
über das Schreiben und nicht über das Bezahlen; sie steht in
#1452 als
eigenes Werkstück.
