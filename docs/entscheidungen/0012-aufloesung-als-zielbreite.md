# 0012 — Die Auflösung ist eine Zielbreite, keine dpi-Zahl

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

#49 sagt zu:
„PNG und JPG, **Auflösung wählbar**". Die Frage ist, worin sie gewählt wird.

Der Zeichenplan misst in **Vorlagenpunkten**, und `render.Zeichne` nimmt eine
Vergrößerung: `1` heißt Bildschirm, `4` heißt Druck
(#34). Für eine
Vorschau ist das die richtige Größe — sie *ist* ein Zoom.

## Entscheidung

**Die Ausgabe nimmt eine Zielbreite in Bildpunkten.** Die Vergrößerung entsteht
daraus: `zielbreite / plan.Breite`. `0` heißt „so groß, wie der Plan es sagt".

Die Oberfläche bietet dafür benannte Stufen und zeigt daneben, was dabei
herauskommt — die Zahl bleibt sichtbar und nachrechenbar.

## Verworfen: dpi

Das wäre die Angabe, die jeder kennt, und wir können sie **nicht halten**.

Eine dpi-Zahl ist eine Aussage über Bildpunkte **je Zoll** — sie setzt voraus,
dass irgendwo steht, wie groß die Karte in Millimetern ist. Das steht nirgends:

- Der Plan trägt `Breite` und `Hoehe` in Vorlagenpunkten. Was ein Punkt in der
  Welt ist, sagt er nicht.
- Die Vorlage sagt es auch nicht. `magic-m15.mse-style` misst in Bildpunkten
  einer Bildschirmdarstellung; eine physische Größe kommt darin nicht vor
  (`docs/FORMATE.md`).
- Der Behelfsplan nimmt 375 × 523 — das übliche **Verhältnis** einer Spielkarte
  (63 × 88 mm), nicht ihre Größe.

Wer daraus „300 dpi" anbietet, rechnet mit einer Zahl, die er sich ausgedacht
hat. Das ist genau die Sorte Zusage, die §10 meint: eine Ablehnung braucht
Deckung, und eine Behauptung erst recht.

**Die Rechnung gehört dem, der druckt**, und sie ist eine Zeile: 750
Bildpunkte auf eine 63 mm breite Karte sind 300 dpi. Was wir dafür liefern
müssen, ist die Breite — nicht eine Zahl, die so aussieht, als hätten wir die
Karte gemessen.

## Verworfen: die Vergrößerung durchreichen

Ehrlich und unbrauchbar. „Viermal" sagt nur etwas, wenn man die Größe der
Vorlage kennt, und die steht in einer Datei, die niemand aufmacht. Zwei Sets mit
verschiedenen Stilen ergäben bei derselben Zahl verschieden große Bilder — und
niemand sähe, warum.

## Was das kostet

**Die Vorschau und die Ausgabe zählen verschieden.** Die eine nimmt einen
Faktor, die andere eine Breite. Das ist ein Bruch, und er ist gewollt: ein Zoom
ist ein Faktor, eine Ausgabe ist eine Größe. Wer beides gleich machte, hätte
entweder eine Vorschau, die nach Bildpunkten fragt, oder eine Ausgabe, die nach
einem Faktor fragt.

**Eine Grenze musste dazu.** `GrenzeZielbreite` steht bei 10 000 Bildpunkten,
und sie ist nicht kosmetisch: der rote Lauf des Tests, der 100 000 verlangte,
hat zwei Minuten lang versucht, vierzehn Milliarden Bildpunkte anzufordern. Eine
Zahl aus einem Eingabefeld darf den Rechner nicht anhalten.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn eine Vorlage doch eine physische Größe führt. Dann ist dpi berechenbar,
  und die Oberfläche darf danach fragen — die Zielbreite bleibt darunter.
- Wenn Nutzende regelmäßig danebengreifen, weil sie dpi eintippen wollen. Dann
  fehlt nicht die Einheit, sondern die Umrechnung neben dem Feld.
