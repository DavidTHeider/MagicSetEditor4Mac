# 0160 — Was ein Bild ausgepackt kostet

**Stand:** angenommen am 2026-10-01 · FMT-06b
(#1974),
ein Folgebefund aus QA-06b (#1967).

## Die Frage

Die Grenze für Bilder zählte nur Bildpunkte: höchstens 32.768 je Seite und
hundert Millionen zusammen. Ein Bildpunkt kostet beim Auspacken aber je nach
Kodierer zwischen einem und vierundzwanzig Byte. Unter der Grenze waren
deshalb bis zu 2,4 GB je Bild erlaubt. `FuzzDekodiereBild` hat das
gefunden. Zu entscheiden war, ob und wie die Grenze auch in Byte zählt.

## Die Entscheidung

1. **Eine zweite Grenze, `GrenzeAuspacken`: ein Gibibyte.** Geprüft wird
   sie nach der Bildpunktgrenze und vor dem Auspacken.
   - Das größte Bild im vollen Pack ist eine Karte bei 1200 dpi mit 12,4
     Millionen Bildpunkten. Selbst als progressives JPEG in CMYK kostet sie
     unter 320 MB.
   - Ein JPEG in YCbCr darf damit rund 67 Millionen Bildpunkte haben.
   - Ein PNG mit RGBA hält schon die Bildpunktgrenze darunter.
2. **Die Schätzung ist gemessen, nicht erschlossen.** Gemessen wurde am
   2026-10-01 mit Go 1.27.1, je Format und Farbmodell ein Bild von 512 × 512
   Bildpunkten, die zugeteilten Byte über `runtime.MemStats`:

   | Format | Farbmodell | gemessen (Byte je Bildpunkt) | geschätzt |
   | --- | --- | --- | --- |
   | PNG | Grau, Palette | 1,2 · 1,3 — mit Zeilensprung 2,2 | 3 |
   | PNG | Grau16 | 2,2 | 5 |
   | PNG | RGBA | 4,2 — mit Zeilensprung 8,3 | 9 |
   | PNG | RGBA64 | 8,2 — mit Zeilensprung 16,3 | 17 |
   | GIF | Palette | 2,1 | 3 |
   | JPEG | Grau | 1,1 — progressiv 5,1 | 6 |
   | JPEG | YCbCr | 1,6 (4:2:0) — progressiv 15,1 (4:4:4) | 16 |
   | JPEG | CMYK | 8,1 — progressiv 24,1 | 25 |
   | unbekannt | | — | 25 |

   - **Der Kopf sagt nicht alles.** Ob ein PNG Zeilensprung trägt und ob
     ein JPEG progressiv ist, steht nicht in `image.Config`. Geschätzt wird
     deshalb der teurere Fall, ein Byte darüber.
   - Ein JPEG rechnet über die Fläche, aufgerundet auf Blöcke zu sechzehn
     Bildpunkten.
   - `auspackkosten_test.go` hält die Messung fest: je Fall ein Bild, und
     was die Kodierer zuteilen, bleibt unter der Schätzung plus einem
     Sockel von 1 MiB.
   - Bilder, die die Standardbibliothek nicht schreibt — progressives JPEG
     und PNG mit Zeilensprung —, liegen selbst erzeugt unter
     `testdaten/auspackkosten/`.
3. **Eine Abweisung mit eigenem Schlüssel: `bildZuGross`**, in allen acht
   Katalogen und an jeder Stelle, die Bilder einliest.
   - Bis hierher hieß ein zu großes Bild überall „kein lesbares Bild“. Das
     schickte jemanden nach einem Fehler in einer Datei suchen, die in
     Ordnung ist.
   - `msepaket.ZuGrossOder` reicht die Abweisung durch, statt sie
     umzuschreiben.
4. **`FuzzDekodiereBild` sagt dieselbe Zahl zu.** Statt pauschal 24 Byte
   gilt dort die Schätzung je Format plus Sockel. Die Grenze, die der
   Fuzzer prüft, ist damit die, die gilt.

## Verworfen

- **Den Kopf weiter lesen**, um Zeilensprung und Progressivität zu
  erkennen. Das spart bei Grundlinien-JPEG den Faktor zehn, kostet aber
  einen eigenen Leser für zwei Formate neben dem der Standardbibliothek. Bei
  12,4 Millionen Bildpunkten im größten echten Fall lohnt das nicht.
- **Die Bildpunktgrenze senken**: sie träfe auch Bilder, die billig
  auszupacken sind, und ließe die teuren Fälle trotzdem zu teuer.

## Was es kostet

- Ein Grundlinien-JPEG in YCbCr wird zehnmal so teuer geschätzt, wie es
  ist. Über 67 Millionen Bildpunkten wird es abgewiesen, obwohl es
  ausgepackt nur rund 110 MB kostete.
- Die Faktoren gelten für Go 1.27.1. Ein neuer Kodierer kann sie
  verschieben. Der Messlauf merkt es, weil er gegen die Schätzung prüft und
  nicht gegen eine Zahl in einem Kommentar.
