# 0013 — Der PDF-Schreiber ist selbst geschrieben

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

#50 verlangt
einen **Druckbogen als PDF**: Karten maßhaltig auf einer Seite, mit Anschnitt
und Schnittmarken.

PDF ist ein großes Format. Was ein Druckbogen davon braucht, ist es nicht.

## Entscheidung

**`internal/pdf` schreibt PDF selbst** — Seiten mit einer Größe, Bilder an
genauen Stellen, gerade Linien. Dreihundert Zeilen, keine Abhängigkeit.

Kein Text, keine Schriften, keine Formulare, keine Verschlüsselung, keine
Ebenen, keine Farbprofile. Bilder gehen als `FlateDecode` in `DeviceRGB` hinein,
ohne Alphakanal.

## Verworfen: eine PDF-Bibliothek

`go-pdf/fpdf` und Verwandte können all das und viel mehr. Genau das ist der
Einwand: der Teil, den wir brauchen, ist der kleine, und der Rest kommt mit —
mit seiner Pflege, seinen Fassungen und seiner Angriffsfläche für einen Nutzen
von null.

Dazu kommt, was dieses Projekt ohnehin tut: den MSE-Zerteiler, den
Skriptauswerter, den Rasterer und den MCP-Server hat es selbst geschrieben, weil
die Zusage jeweils genauer war als das, was eine Bibliothek gibt. Hier ist es
dasselbe: **maßhaltig** heißt, dass wir die Zahlen im Inhaltsstrom selbst
setzen.

## Verworfen: ein Bild je Seite, gebaut vom Rasterer

Der Bogen wäre ein großes Bild, und das PDF trüge es. Das ist einfacher und
falsch:

- Ein A4-Bogen bei 300 dpi ist 2480 × 3508 Bildpunkte. Als eine Fläche gerastert
  wären das 26 MB je Seite im Arbeitsspeicher — und die Karten würden **zweimal**
  skaliert, einmal in ihre Größe und einmal in die Fläche.
- Schnittmarken wären Bildpunkte statt Linien. Auf Papier sieht man den
  Unterschied: eine Vektorlinie ist bei jeder Auflösung eine Haarlinie, eine
  gerasterte ist ein grauer Streifen.

## Was das kostet

**Dieses Paket kann genau eine Sorte PDF.** Wer Text im PDF will — eine
Seitenzahl, den Setnamen am Rand —, baut es hier ein: eine Standardschrift
(`/Helvetica`) und ein `BT … ET` sind je zehn Zeilen, aber sie fehlen heute.

**Es gibt keine Querverweisströme und keine Objektströme** (PDF 1.5). Die Datei
ist dadurch etwas größer und von jedem Leser zu lesen, auch von einem alten.

**Und es prüft sich nicht selbst.** Ein PDF, das nur unser eigener Leser liest,
ist kein PDF. Deshalb steht in jedem Test ein **fremder** Leser: `sips` gehört
zu macOS, ruft Quartz und gibt die Seitengröße zurück. Die Gegenprobe ist
gefahren — mit einer um sieben Bytes verschobenen Querverweistabelle antwortet
Quartz `0.00 × 0.00`, liest also nicht.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn der Bogen Text braucht, der sich als Text markieren lassen soll. Dann
  wächst dieses Paket über seinen Zuschnitt hinaus, und eine Bibliothek ist die
  ehrlichere Antwort.
- Wenn ein Drucker die Datei zurückweist, den Quartz annimmt. Dann ist `sips`
  als einziger äußerer Maßstab zu weich, und es gehört ein zweiter daneben —
  Ghostscript liegt auf dieser Maschine und könnte ihn geben.
- Wenn die Dateigröße stört. Dann ist der Weg `DCTDecode` für Bilder, die
  ohnehin aus Fotos bestehen — nicht eine andere Bibliothek.
