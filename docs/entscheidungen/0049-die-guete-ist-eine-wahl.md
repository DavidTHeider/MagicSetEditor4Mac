# 0049 — Die Güte ist eine Wahl der Nutzenden, nicht ein Schluss aus der Vergrößerung

**Stand:** angenommen
**Datum:** 2026-08-25
**Gehört zu:** Werkstück Ω.4 aus
#1029
**Berührt:** die Filterwahl aus
#532 — die
Vorgabe bleibt MSEs nachgemessene Filterpolitik; die drei anderen weichen ausdrücklich
ab und sagen es an der Ausgabe

## Lage

Der Rasterer bekommt vier Güten. Die naheliegende Verdrahtung wäre, sie aus der
**Vergrößerung** abzuleiten: klein zeichnen heißt Vorschau heißt schnell, groß
zeichnen heißt Druck heißt sorgfältig. Ein Parameter statt zwei, und niemand
muss etwas wählen.

Das Werkstück verlangt ausdrücklich das Gegenteil, und dieser Vermerk sagt,
warum.

## Was gemessen ist

Über zwanzig Karten aus `Black.mse-set` am 2026-08-25:

| Abschnitt | Mitte |
| --- | --- |
| die Karte lesen | 72 µs |
| den Plan bauen | 35,8 ms |
| rastern | 28,8 ms |

Der Plan ist für **jede** Güte derselbe — Layout und Zeilenumbruch bleiben
gleich, das ist die Zusage des Werkstücks. Eine Güte kann also nur die 45
Prozent anfassen, die das Rastern trägt. Davon tragen Bildebenen 78 Prozent,
und dort liegt der Hebel: die Abtastung.

Gemessen an denselben zwanzig Karten, je Güte:

| Güte | Mitte | gegen `medium` |
| --- | --- | --- |
| `fast` | 17,5 ms | −27 % |
| `medium` | 23,9 ms | — |
| `high` | 32,4 ms | +35 % |
| `ultra` | 45,1 ms | +89 % |

## Entscheidung

**Güte und Vergrößerung gehen getrennt in den Rasterer.** Die Güte ist eine
Wahl; die Vergrößerung sagt, wie groß das Bild wird.

## Warum nicht abgeleitet

**Weil die beiden Fragen nicht dieselbe sind.** „Wie groß?" und „wie
sorgfältig?" fallen nur in dem einen Fall zusammen, den man sich beim Entwurf
vorstellt.

Drei Fälle, in denen die Ableitung falsch läge:

- **Eine große Ansicht beim Tippen.** Wer die Karte formatfüllend offen hat und
  am Regeltext schreibt, bekäme bei jeder Taste die sorgfältigste Rechnung —
  und wartete auf jede.
- **Eine kleine Vorschau vor dem Druck.** Wer prüfen will, ob eine feine Linie
  flimmert, braucht die Druckgüte in der Ansicht, sonst prüft er die Vorschau
  statt den Druck.
- **Ein Vergleich mit MSE.** Der Goldbildvergleich braucht die **Parität**, und
  zwar in jeder Größe. Eine Ableitung machte die Vergleichbarkeit von der
  Vergrößerung abhängig — und der Vergleich prüfte dann sich selbst mit.

## Was verworfen wurde

**Güte aus der Vergrößerung erschließen** — verworfen aus den drei Gründen
oben.

**Nur zwei Güten, schnell und genau** — verworfen, weil die Messung vier
unterscheidbare Stufen hergibt und die beiden mittleren verschiedene Fragen
beantworten: `medium` folgt MSEs Filterpolitik, `high` ist das ruhigere Bild.
Wer sie zusammenlegt, muss eine der beiden Zusagen aufgeben.

**Die Güte am Rasterer festhalten statt am Aufruf** — verworfen, weil ein
Rasterer von fünf Wegen gleichzeitig erreichbar ist
(#748). Eine
Vorschau und eine Ausfuhr laufen nebeneinander und wollen verschiedene Güten.

## Preis

**Zwei Parameter statt einem**, an jeder Naht und an jeder Fläche. Wer die Güte
vergisst, bekommt den Nullwert — und der ist mit Absicht `GueteMittel`, die
Parität: nicht die schnellste und nicht die schönste, sondern die, die MSE
zeichnet. Eine Güte, deren Vergessen das Bild verschöbe, wäre eine Falle.

**Und eine Wahl, die niemand trifft, ist eine Einstellung mehr.** Deshalb hat
jede Fläche eine Vorgabe, und nur wer will, weicht ab.
