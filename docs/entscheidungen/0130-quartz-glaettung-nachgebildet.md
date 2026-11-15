# 0130 — Quartz' Schriftglättung wird nachgebildet, als gemessenes Modell

**Stand:** entschieden am 2026-09-23
(#1806) —
die Entscheidung, **ob**, hat der Mensch getroffen (»Ja, nachbilden.«); wie,
steht hier.

## Was galt

`docs/SCHRIFTUNTERSUCHUNG.md` hielt fest, dass ein Bitmapkontext, wie
wxWidgets ihn für MSE2 anlegt, Schrift glättet und damit Tinte zulegt — und
baute es **nicht** nach: die Glättung sei kein Verhalten von MSE2, sondern
eines Rasterers, dessen Rechnung nicht offenliegt, und ein Nachbau wäre
gemessen statt nachgelesen. Nach QA-03 trug Text noch 43 % des Abstands zu
MSE2 am Kanon.

## Was gilt

**Die Glättung wird nachgebildet, und zwar als Gleichung aus einer Messung**
(`internal/render/glaettung.go`):

- **Sie ist eine Verdickung des Umrisses, keine Gammakurve.** Quer durch
  einen Stamm rücken die Kanten; der Verlauf dazwischen bleibt.
- **Waagerecht je Seite `d = min(Geviert/64; 0,3)` Bildpunkte, senkrecht nur
  nach oben, um `2d`.** Die Unterkante bleibt stehen.
- **Achsparallel**: eine Kante mit Normale `n` rückt um `d·(|nₓ| + |n_y|)`,
  wie eine Minkowski-Summe mit einem Quadrat — nicht um `d` entlang der
  Normalen.
- Die Punkte des Umrisses wandern dorthin, wo sich ihre zwei verschobenen
  Kanten schneiden; das Gerüst samt Grenzen ist FreeTypes
  `FT_Outline_EmboldenXY`.

**Gemessen am 2026-09-23** mit `scripts/coretext/glaettung.swift` an MPlantin,
Beleren Bold und Relay-Medium von 8 bis 128 Bildpunkten je Geviert:
links = rechts = `d` auf ±0,005, oben `2d`, unten 0. Die Winkelabhängigkeit an
einem gedrehten Stamm: bei 10, 20 und 30 Grad legt er das 1,17-, 1,29- und
1,38-fache der Tinte zu, die er ungedreht zulegt; `|cos| + |sin|` gibt 1,16,
1,28 und 1,37.

**Und am Ergebnis**: eine Zeile auf MSE2s vierfachem Weg, gegen Core Text mit
Glättung, mittlerer Betragsabstand in Stufen von 0 bis 255:

| Zeile | ohne Nachbildung | mit |
| --- | ---: | ---: |
| MPlantin, 14 Punkt, Regeltext | 1,92 | **0,18** |
| Beleren Bold, 16 Punkt, Name | 2,60 | **0,32** |
| MPlantin, 6 Punkt | 1,74 | **0,54** |

`TestDieGlaettungTrifftQuartz` hält die ersten beiden gegen Core Text selbst —
er lässt `raster.swift` zeichnen, statt eine gespeicherte Zahl zu vergleichen,
und er sieht auch, wenn eine künftige Fassung von macOS anders glättet.

**Für die Ausgabe, nicht nur für die Messung.** Dieser Rasterer hat einen Weg,
und die Messung gegen MSE2 misst ihn. Eine Glättung nur für die Referenz wäre
eine Messung an einem Bild, das niemand sieht — und ein zweiter Rasterer, der
dem ersten davonläuft.

## Was das Modell nicht trägt

Die Grenzen stehen hier und nicht in einer Fußnote (§10):

- **Gedrehter Text.** Quartz verdickt im **Gerätebild**: ein um 90° gedrehter
  Stamm wächst im Bild nach oben, nicht in Richtung seiner Glyphe. Nahe 45
  Grad glättet Quartz schwächer und bei genau 45 Grad gar nicht. Dieser
  Rasterer verdickt im Glyphenbild; bei gedrehtem Text weicht er ab.
- **Die Einstellung des Systems.** Ob macOS' Einstellung zur Schriftglättung
  den Wert ändert, ist nicht gemessen — gemessen ist dieser Rechner mit seiner
  Vorgabe, macOS 27.0 (26A428).
- **Kleine Größen.** Unter rund 12 Bildpunkten je Geviert streut die Messung
  an dünnen Stämmen, die in einem Bildpunkt klemmen; MSE2 zeichnet dort nur
  Text unter 3 Punkt.
- **Was Quartz sonst noch tut.** Die Gleichung ist die, die die Messung trägt,
  nicht die Rechnung von Quartz. Der verbleibende Abstand von 0,18 bis 0,54
  ist nicht aufgeteilt.

## Verworfen

- **Nicht nachbilden und den Anteil im Kanonbericht benennen.** Das war die
  andere Möglichkeit aus #1806; entschieden ist anders.
- **Eine Gammakurve auf die Deckung.** Am 2026-08-20 probiert: sie brachte die
  Tinte auf 99 % und den Abstand **nach oben** (20,692 → 20,774). Die fehlende
  Tinte sitzt an den Kanten, nicht in der Mitte der Stämme.
- **Eine Verdickung entlang der Normalen**, wie FreeType sie rechnet. Sie ist
  drehungsgleich, und die Messung am gedrehten Stamm ist es nicht. Sie traf
  die Tinte zu 97 % statt 99 %.
- **Nach beiden Seiten gleich, auch senkrecht.** So stand es im ersten Anlauf,
  und der Abstand fiel nur auf 0,90 — die Reste saßen Zeile für Zeile an den
  Oberkanten und auf der Grundlinie. Die Messung an beiden Kanten eines
  waagerechten Stamms zeigte warum.
