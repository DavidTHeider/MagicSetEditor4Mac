# 0151 — Der Handel in vier Phasen, und nur die Schritte, die gehen

**Stand:** entschieden am 2026-09-30 · TRADE-02.AK-01 bis AK-04
(#1716)

## Die Frage

Der Handelsraum reihte sechs Blöcke untereinander — Abgleich, Post, Wahl,
Vertrag mit allen neun Schrittknöpfen, Zusagen, Ausgang und Buchung —, gut
1.700 Punkte hoch. TRADE-02 verlangt ungefähr vier verständliche Phasen, eine
Lage mit der nächsten zulässigen Handlung, einen Entwurf, den Zurück nicht
kostet, und einen Beleg über denselben Szenariosatz vorher und nachher.

## Was MSE2 tut

**Es handelt nicht** (siehe [0149](0149-was-eine-handelsnachricht-behaupten-darf.md)).
Marke **new**.

## Was gilt

1. **Vier Phasen, aus `domain`** (`Handelszustand.Phase`): *Mit wem*
   (keine Zustände), *Was tauscht ihr* (Entwurf, angeboten, Verhandlung),
   *Zusagen* (angenommen, vorbereitet), *Verbuchen* (Übergabe bis
   beidseitig verbucht, dazu die Klärung). Abgelehnt, abgebrochen und
   abgelaufen sind **kein** fünfter Abschnitt: sie zeigen die Tafel des
   Angebots, und die Lage sagt „Beendet". Das Zustandsmodell bleibt, wie es
   ist.
2. **Lage, nächster und mögliche Schritte kommen mit dem Vorgang** an allen
   vier Flächen (`lage`, `naechster`, `moegliche`, `phase` — und der
   `beschluss` aus TRADE-03, der bis hierher nur das Fenster erreichte). Das
   Fenster zeigt **nur**, was `moegliche` nennt, den üblichen Schritt als
   Hauptknopf. Der alte Einwand gegen ausgegraute Knöpfe — ein Fenster, das
   die Matrix nachrechnete, hätte eine zweite Fassung von ihr — gilt weiter,
   und deshalb rechnet es nicht.
3. **Die Schritte des Abschlusses sind keine Knöpfe der Leiste**: verbuchen,
   abgleichen und bestätigen bewegen Bestand und Beleg und stehen mit ihrer
   Freigabe in der Tafel *Verbuchen*.
4. **Was am gewählten Handel hängt, steht bei ihm**: kopieren, archivieren,
   verwerfen und seine Marken über allen Tafeln — nicht hinter *Mit wem*.
5. **Die Tafeln bleiben eingehängt** und werden über `hidden` verborgen —
   Zurück kostet den Entwurf nicht (AK-03). **Jede Phase lässt sich
   ansehen**, auch eine kommende: ihr Journal, was unterwegs ist, was
   festzuhalten wäre. Die Tafel sagt selbst, was dort noch nicht geht.
6. **Die Ansicht folgt dem Vorgang**, wenn er weiterrückt oder ein anderer
   gewählt wird, und bleibt, wo sie ist, solange er steht.
7. **Abgleich und Post sind Kästen**: *Was ihr tauschen könntet* in *Was
   tauscht ihr*, *Was unterwegs ist* in *Verbuchen*.

## Gemessen (AK-04)

Derselbe Szenariosatz vor und nach dem Umbau — drei Tausche aus Sicht der
Initiatorin (Entwurf, angeboten, angenommen), am **nativen Fenster** mit der
Hüllenprobe `trade02`, gezählt im rechten Bereich der Seite:

| | vorher (`main` @ 4bdd9e0c) | nachher |
| --- | ---: | ---: |
| sichtbare Knöpfe | 91 | 35 |
| sichtbare Eingaben | 27 | 12 |
| Überschriften | 15 | 3 |
| **Schrittknöpfe, die hier nicht gehen** | **20** | **0** |
| Höhe des Raums je Zustand | rund 1.700 pt | 838 pt |

**Unverändert sind die Zusagen des Vertrags**: dieselben Bindungen,
dieselbe Matrix, dieselben Läufe gegen Handelsvertrag, Draht und Abnahme.
**Was die Probe nicht misst**: Fehlversuche von Menschen — gezählt sind
Knöpfe, auf die ein Fehlversuch möglich war.

## Verworfen

- **Knöpfe ausgrauen statt ausblenden**: ein ausgegrauter Knopf, der nie
  geht, ist ein Satz ohne Inhalt, und er zählte zu den zwanzig.
- **Kommende Phasen sperren**: wer vor dem Abschluss sein Journal ansehen
  will, fände es nicht mehr — und eine gesperrte Phase erklärt nicht, warum.

## Was offen bleibt

- ~~**Die zwei Spalten des Artboards**~~ — gebaut mit TRADE-02b
  ([0153](0153-preise-als-anhalt.md)).
- **„Neuer Tausch" in der Liste links** als eigener Knopf.
