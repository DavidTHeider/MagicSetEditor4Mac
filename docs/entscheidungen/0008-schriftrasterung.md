# 0008 — Glyphen über `golang.org/x/image`

**Stand:** angenommen
**Datum:** 2026-08-09

## Lage

`internal/render` **setzt** Text seit Phase 4: Umbruch, Ausrichtung, Laufweite,
Zeilenhöhe. Es **zeichnet** ihn nicht. Eine Textebene mit gefundener Schrift gibt
bis heute einen Befund zurück:

> Text ist gesetzt (drei Zeilen), aber noch nicht gezeichnet — es fehlt der
> Rasterer für Schriftdateien

Der Grund ist keine Nachlässigkeit: **Go bringt keinen Rasterer für
Schriftdateien mit.** Weder `image/font` noch sonst etwas in der
Standardbibliothek setzt eine TrueType- oder OpenType-Datei in Glyphen um. Es
gibt in der Standardbibliothek überhaupt keinen Schrifttyp jenseits der
eingebauten Bitmapschrift von `golang.org/x/image/font/basicfont` — und die ist
schon nicht Standardbibliothek.

Damit steht die Phase 6 an zwei Stellen still: der Bildexport
(#49) schriebe
Karten ohne Text, der Druckbogen
(#50) legte sie
leer auf den Bogen. Auch der Goldbildvergleich aus
#35 prüft
bislang Bilder ohne Schrift.

## Entscheidung

**Glyphen kommen aus `golang.org/x/image`** — `font/sfnt` liest die Schriftdatei,
`font/opentype` gibt die Metrik und die Umrisse, `vector` rastert sie.

Die Abhängigkeit wird wie jede andere gepinnt (`go.mod`), steht in
`DRITTANBIETER.md` und in der Tabelle aus `AGENTS.md` §1.

## Was das kostet, und warum es trotzdem der Weg ist

**Es ist die erste Abhängigkeit in `internal/render`.** Das Paket kennt bis
heute nur die Standardbibliothek — nicht einmal `internal/domain` —, und diese
Eigenschaft geht verloren. Sie war kein Selbstzweck: ein Renderer ohne fremden
Code ist einer, dessen Ausgabe nur von uns abhängt. Ab jetzt hängt sie auch
davon ab, wie `x/image` Kurven rastert.

**Für den Goldbildvergleich heißt das:** ein Aktualisieren der Abhängigkeit kann
Goldbilder verschieben, ohne dass jemand etwas an unserem Code geändert hat. Das
ist kein Grund gegen die Abhängigkeit, aber ein Grund, sie zu pinnen und beim
Anheben die Goldbilder anzusehen statt sie neu zu schreiben.

**Was es nicht kostet:** die Kreuzkompilierung. `x/image` ist reines Go; die
Bauten für `GOOS=windows` und `GOOS=linux` aus §12 laufen weiter.

## Verworfen

**Core Text über cgo.** Wäre der beste Rasterer auf dieser Plattform — macOS
zeichnet Schrift besser als jede Bibliothek, die wir einbinden könnten. Und er
**beendet die Kreuzkompilierung**: mit cgo braucht jedes Ziel eine Werkzeugkette
für dieses Ziel, und §12 verlangt ausdrücklich, dass für Windows und Linux
übersetzt wird, damit plattformabhängiger Code nicht unbemerkt bricht. Ein
Renderer, der nur auf macOS baut, nähme diesem Wächter die Grundlage.

Der Vollständigkeit halber: derselbe Einwand gälte für jede C-Bibliothek —
FreeType, HarfBuzz, Pango.

**Selbst schreiben.** Ein Zerteiler für TrueType-Tabellen und ein Rasterer für
quadratische Bézierkurven mit Anti-Aliasing. Das ist ein eigenes Projekt, und es
wäre schlechter als `x/image`, das dieselbe Aufgabe seit Jahren löst.

**Text als Bild vorberechnen.** Wäre kein Renderer, sondern eine Ausrede: die
Kartentexte entstehen aus Skripten und ändern sich beim Tippen.

**Weiter ohne Text.** Verworfen, weil es die Zusage aus
[0002](0002-sqlite-als-arbeitsformat.md) und den Zweck des Programms trifft: ein
Karteneditor, der keine Kartentexte zeichnet, ist ein Bildbetrachter für Rahmen.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn ein Anheben der Abhängigkeit Goldbilder so verschiebt, dass die Toleranz
  aus #35 nicht mehr trägt — dann ist entweder die Toleranz zu eng oder die
  Bibliothek zu unstet.
- Wenn die Textausgabe sichtbar schlechter aussieht als in MSE2. Dann ist der
  Vergleich zu führen und **nicht** stillschweigend hinzunehmen: der Preis
  dieser Entscheidung war die Kreuzkompilierung, und der ist zu hoch, wenn das
  Ergebnis nicht trägt.

## Was daraus folgt

Der **Suchweg** für Schriften bleibt davon unberührt und ist eine eigene Sache:
`x/image` liest eine Schriftdatei, es findet keine. Wo `MPlantin` auf macOS
liegt und wie das anderswo aussähe, steht in `docs/PORTIERUNG.md`.
