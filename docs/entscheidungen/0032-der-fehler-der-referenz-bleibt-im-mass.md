# 0032 — Der Fehler der Referenz bleibt im Maß, und er wird beziffert

**Stand:** angenommen
**Datum:** 2026-08-20
**Ergänzt:** [0026](0026-mse2-rendert-die-referenz-selbst.md) — MSE2 rastert die
Referenz selbst; hier steht, was zu tun ist, wenn MSE2 dabei irrt

## Lage

**Die Referenz ist an einer Stelle nachweislich falsch.** MSE2 schneidet das
Innere eines Setsymbols über `dc.Blit(…, wxAND_INVERT)` aus. Unter
**wxOSX-Cocoa 3.3.3 gibt dieser Blit `false` zurück und tut nichts** — gemessen,
nicht erschlossen. Das ganze Symbol bekommt damit die **Randfarbe** der Fassung
statt der Füllfarbe.

Über die 3.761 Karten von `White`, ein Punkt mitten im Symbol:

| | |
| --- | --- |
| Randfarbe | **100 %** |
| Füllfarbe | **0 %** |
| `common` reinweiß | **1.526-mal**, obwohl seine Füllfarbe schwarz ist |

Es gibt keine Lesart des Quelltexts, unter der ein richtig arbeitendes MSE ein
Common-Symbol weiß malt
(#458).

**Und unsere Rechnung ist an derselben Stelle belegt richtig**: gemessen ist
(172,172,172), gerechnet ergibt `lerp(224, 84, 0,41)` den Wert **167** —
dieselbe Farbe innerhalb der Abtastung, also Silber, wie ein Uncommon sein
soll. `symbolfilter.go` trägt seine `filter.cpp`-Zeilen.

**Die beiden Zahlen sind nicht dieselbe**, und hier stand bis zum 2026-08-20,
die eine „ergebe sich" aus der anderen. Sie belegen zusammen die **Farbklasse**
und nicht den Bildpunkt; der Zweitblick über Phase K hat den Widerspruch zu
Recht angestrichen.

Das Issue nannte zwei Wege, beide mit Preis, und keiner war entschieden.

## Entscheidung

**Der Fehler bleibt im Maß, und er wird beziffert.**

`internal/app/referenzfehler_intern_test.go` rechnet den Kanonlauf ein zweites
Mal ohne das Feld `rarity` und meldet beide Zahlen. Am 2026-08-20 über 573
Karten aus vier Sätzen:

| | |
| --- | ---: |
| Abstand ganz | **20,692** |
| eigener Abstand des Feldes `rarity` | 87,031 |
| Anteil an der Fläche | 0,452 % |
| **Abstand ohne dieses Feld** | **19,774** |
| **was der Fehler höchstens kostet** | **0,918** — 4,4 % |

**Obergrenze und keine Zuordnung.** Der Lauf nimmt das ganze Feld heraus, nicht
nur den Anteil, den `wxAND_INVERT` verdirbt. Dass der Löwenanteil der Referenz
gehört, steht an der Auszählung oben und nicht in dieser Zahl.

**„Ohne dieses Feld" heißt: als träfe es die Referenz genau.** Die Bildpunkte
des Feldes zählen weiter zur Karte, sie tragen nur nichts mehr bei — der Nenner
bleibt die ganze Karte. Die andere Lesart, der Abstand der **übrigen** Fläche,
ergibt **19,820** statt 19,774; der Schluss trägt beide Male. Angestrichen vom
Zweitblick über Phase K, und beide stehen seither im Lauf.

**Nicht umgefärbt.** Wer `symbolfilter.go` an die Referenz anpasste, portierte
eine Einschränkung von wxOSX und machte die Ausgabe schlechter, während die
Zahl besser würde.

## Verworfen

**Den macOS-Bau patchen, sodass `combineBuffers` das `AND_INVERT` auf den
Bilddaten rechnet, und die Referenz neu erzeugen.** Der Reiz ist echt: die
Referenz wäre dann näher an dem, was MSE2 unter Windows tut — und *das* ist
das Ziel dieses Projekts.

Dagegen steht, was [0007](0007-kein-entfernter-lauf.md) über diese Werkstatt
sagt: **geprüft ist eine Maschine, nicht eine Plattform.** Ein Patch, der
Verhalten *wiederherstellt*, ist eine Behauptung über Windows, die hier
niemand nachprüfen kann — im Unterschied zu `docs/mse2-macos.patch`, der nur
den Bau auf arm64 möglich macht und am Verhalten nichts ändert. Die Referenz
verlöre damit die Eigenschaft, die sie in
[0026](0026-mse2-rendert-die-referenz-selbst.md) überhaupt erst brauchbar
macht: **Eingabe und Ausgabe stammen von MSE2 selbst**, auf diesem Rechner,
nachvollziehbar.

**Die Symbolfläche vom Vergleich ausnehmen.** Billig und ehrlich, und trotzdem
falsch: eine ausgenommene Fläche ist eine Fläche, in der eine echte
Verschlechterung unsichtbar wird. Der Ausschluss müsste je Vorlage gepflegt
werden, und die erste Vorlage, die ihr Symbol woanders hinlegt, hätte ihn
still verfehlt. Die Zahl daneben leistet dasselbe, ohne die Messung zu
verkleinern.

**Nichts tun und den Satz aus der Roadmap stehen lassen.** Das war der Stand
bis heute — „er trägt einen Fehler der Referenz". Ein Satz ohne Zahl ist
genau das, was §10 anstreicht: „dort ist nichts zu holen" sagt nicht, wie viel
„dort" ist, und wer den Abstand plant, kann damit nicht rechnen.

## Folgen

- **Der Maßstab in `docs/ROADMAP.md` nennt beide Zahlen** — 20,692 und, ohne
  das Feld, 19,774.
- **Eine Verbesserung, die aus dem Symbolfeld käme, ist ab jetzt verdächtig.**
  Wer dort etwas gewinnt, hat wahrscheinlich die Referenz nachgebaut statt
  MSE2s Absicht.
- **Der Lauf gehört wiederholt, wenn sich der Maßstab ändert.** Er ist eine
  Messung, kein Deckel: er meldet und scheitert nicht.
- **Fällt die Einschränkung von wxOSX einmal weg** — eine neuere wx-Fassung,
  ein anderer Rückweg —, ist diese Entscheidung zu prüfen und nicht
  fortzuschreiben.
