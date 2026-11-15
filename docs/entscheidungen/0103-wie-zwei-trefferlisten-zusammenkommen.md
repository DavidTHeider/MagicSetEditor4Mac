# 0103 — Wie zwei Trefferlisten zusammenkommen

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A15 von Phase 5.3
(#1514)

---

## Die Frage

Eine hybride Suche hat zwei Hälften. Die lexikalische ordnet nach **BM25**, die
semantische nach **Cosinus**. Beide geben eine Liste; gebraucht wird eine.

**Wie kommen sie zusammen?**

---

## Warum die naheliegende Antwort falsch ist

**Punkte addieren geht nicht.** BM25 ist nicht nach oben begrenzt und wächst
mit der Länge des Korpus und der Seltenheit der Wörter; ein Cosinus liegt
zwischen −1 und 1. `0,7 · bm25 + 0,3 · cosinus` sieht nach einer Gewichtung
aus und ist eine Zahl ohne Einheit — bei einem doppelt so großen Korpus
verschiebt sich das Verhältnis, ohne dass jemand etwas geändert hat.

**Das Issue verbietet es ausdrücklich**: „BM25, Cosinus und Rerankwerte nicht
roh addieren und nicht als Wahrscheinlichkeit ausgeben." Und es sagt, was
stattdessen gilt: „Standard zunächst dokumentierte Rangfusion, z.B. Reciprocal
Rank Fusion; alternative gewichtete Scores nur nach expliziter
Normalisierung/Kalibrierung."

**Eine Kalibrierung wäre möglich** — und sie braucht eine Urteilsmenge,
Messungen je Korpusgröße und eine Nachmessung bei jedem Modellwechsel. Das ist
Werkstück 5 dieses Issues und nicht seine Voraussetzung.

---

## Die Entscheidung

**Reciprocal Rank Fusion mit k = 60.** Je Liste zählt ein Dokument

    1 / (k + Rang)

und die Summe über die Listen ordnet. Ein Dokument, das in **beiden** Listen
steht, sammelt zweimal und steht deshalb vor einem, das nur eine Liste kennt —
auch wenn es dort weiter hinten lag. Genau das ist der Gewinn einer
Hybridsuche.

**k = 60 ist übernommen und nicht gewählt.** Die Zahl steht so in der Arbeit,
die das Verfahren beschreibt (Cormack, Clarke, Büttcher, SIGIR 2009). Sie
dämpft die vorderen Plätze: ohne sie schlüge ein erster Platz jede
Übereinstimmung zwischen den Listen.

**Gerechnet wird über die Reihenfolge der Eingabe, nicht über das Feld
`Rang`.** Eine Ablage, die vertauschte Ränge lieferte, verschöbe sonst still
die Gewichte.

**Der Gleichstand bricht an der Kennung.** `map` ordnet in Go absichtlich
nicht; ohne diesen zweiten Teil gäbe dieselbe Frage zweimal zwei Listen.

**Das Verfahren steht im Manifest** (`rrf-60`). Ein Index mit anderer
Zusammenführung ist ein anderer Index — er sucht im selben Vektorraum, aber
eine Messung gegen ihn vergleicht zwei Verfahren statt zweier Korpora.

---

## Was daraus folgt

**Es gibt keine Ähnlichkeitszahl in der Antwort.** Ein Treffer trägt seinen
**Rang** und seine Herkunft, keinen Wert zwischen null und eins. Wer eine
solche Zahl zeigte, behauptete eine Genauigkeit, die das Verfahren nicht hat —
und jemand würde sie vergleichen.

**Und die Filter greifen danach, aber vor der Grenze.** Gefragt wird weiter,
als geantwortet wird (`service.Kandidatenfenster`, Faktor zehn, mindestens 50,
höchstens 500). Der Grund steht im Issue: „Filterfall darf nicht allein durch
ungefiltertes Top-k seine Treffer verlieren."

**Was das Fenster nicht leistet, steht neben ihm**: ein Filter, dessen Treffer
sämtlich jenseits des Fensters liegen, verliert sie trotzdem. Das Fenster
verschiebt die Grenze; es hebt sie nicht auf. Wer die Zusage braucht, filtert
im Bestand — und das ist eine andere Abfrage als diese.

---

## Verworfen

**Gewichtete Punktsummen.** Siehe oben — ohne Kalibrierung eine Zahl ohne
Einheit, mit Kalibrierung eine Messung, die bei jedem Modellwechsel neu
fällig ist.

**Nur semantisch suchen.** Eine exakte Suche nach einem Kartennamen oder einer
Regelnummer ist lexikalisch stark und semantisch schwach; das Issue verlangt
außerdem, dass „exakte Suche ohne installierten Encoder funktioniert".

**Nur lexikalisch suchen.** Dann gäbe es A15 nicht. Der lexikalische Rückfall
ist eine **Betriebsart mit sichtbarem Grund** und kein Ersatz — das Issue
schließt das ausdrücklich aus: „lexikalischer Fallback wird nicht als
abgeschlossene semantische Suche verkauft."

---

## Was dieser Vermerk nicht entscheidet

**Ob die Hybridsuche in diesem Korpus besser ist als die lexikalische allein.**
Das ist eine Messung, keine Überlegung — Werkstück 5 des Issues, und sie steht
noch aus. Bis sie vorliegt, ist die Rangfusion die **dokumentierte Vorgabe**
und keine belegte Verbesserung.
