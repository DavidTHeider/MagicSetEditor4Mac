# Beleg: Sichtprüfung und Bestand am ausgelieferten Abbild von 0.8.0

**Gefahren am 2026-10-12**, nach der Veröffentlichung, am Abbild
`MagicSetEditor4Mac-0.8.0.dmg` (SHA-256
`99d589bf83751faa5ed96c34b5d55b9a2237b18c5f84da8854e8cda8030127b0`, Commit
`52e1ee3e8`), auf diesem Rechner (Apple M5, macOS 27.0.1). Es sind die
Handprüfungen 3, 4 und 7 des Freigabebelegs. Der Weg ist derselbe wie bei
0.7.0 (`m02-freigabebuendel.md`).

## Wie, ohne den Rechner anzufassen

- Das Abbild ist schreibgeschützt eingehängt
  (`hdiutil attach -readonly -nobrowse`), und die Anwendung wurde **aus ihm**
  gestartet.
- `HOME` und `CFFIXED_USER_HOME` zeigten auf ein Wegwerfzuhause,
  `MSE_SCHLUESSELBUND` auf einen eigenen Bund.
- **Kein Dateidialog** wurde geöffnet. Nach der Berichtigung aus A01 legt ein
  solcher Dialog Vorgaben unter der Produktkennung ab, nicht unter einem
  Zuhause.
- **Vorher und nachher verglichen**: `Application Support`, `Preferences`
  (beide Schreibweisen der Kennung), `WebKit` und `Caches` der echten
  Anwendung. In keinem Ort ist eine Datei neuer als die Zeitmarke vor dem
  Start. Ein Absturzbericht entstand nicht.
- Erscheinungsbild, Sprache und die „Ersten Schritte“ hat das `mse` aus dem
  Abbild im selben Zuhause gesetzt.
- Die Bilder stammen von `screencapture -l`, gemacht vom Fenster des
  Prozesses. Die Fensternummer kommt aus `CGWindowListCopyWindowInfo`, weil
  das Freigabebündel keinen Steuerkanal hat.

## Was zu sehen war (Handprüfungen 3 und 4)

| Bild | Was es zeigt |
| --- | --- |
| Erster Start, hell, Deutsch | Die Startseite mit den „Ersten Schritten“ darüber: „Acht Vorlagenpakete sind schon da“. |
| Nach dem Öffnen des mitgelieferten `Vanguard.mse-set`, hell | 1 Set, 107 Karten; „Was Aufmerksamkeit braucht“: nichts. Lesbar und geordnet, die Leiste mit allen Arbeitsbereichen und Kürzeln. |
| Dasselbe, **dunkel** | Dieselbe Seite im dunklen Erscheinungsbild, lesbar. Der Hauptknopf steht hell auf dunkel. |
| `mse zeichnen` aus dem Abbild, erste Karte | *Akroma, Angel of Wrath Avatar*: Rahmen, Bild, **Titel, Typzeile, Regeltext und Zahlen in den mitgelieferten Schriften**, 375 × 523 in 0,67 s. |

## Der Bestand nach dem Start (Handprüfung 7, nachgestellt)

**Nachgestellt, nicht am echten Bestand.** Den echten Bestand dieses Rechners
zu öffnen, deckt keine Freigabe (`docs/FREIGABEN.md`).

1. Das `mse` aus dem Tag `v0.7.0`, aus dem Quelltext des Tags gebaut, legte
   im Wegwerfzuhause einen Bestand an: das mitgelieferte `Vanguard.mse-set`
   und `saat/probe-werkstatt.mse-set`, also **2 Sets mit 112 Karten** und 97
   verbuchten Wanderungen.
2. Darauf wurde 0.8.0 aus dem Abbild gestartet.

**Ergebnis**:
- Die Startseite zeigt **2 Sets und 112 Karten**.
- Der Bestand trägt danach weiter 97 Wanderungen. Der Aufstieg von 0.7.0
  braucht keinen Schritt, wie Vermerk 0189 zusagt.
- Im Protokoll steht weder eine Warnung noch ein Fehler.

**Nebenbefund, kein Fehler**: „Weiterarbeiten“ bleibt leer, obwohl Sets da
sind. Ein Set, das die Kommandozeile einfährt, gilt nicht als bearbeitet;
vermerkt wird das Öffnen im Fenster (`setOeffnenIm`). So war es auch in 0.7.0:
Der von 0.7.0 angelegte Bestand hatte einen leeren `arbeitsverlauf`, und die
Seite ist seit 0.7.0 nur um Lade- und Fehlerhinweise gewachsen.

## Was dieser Beleg nicht sagt

- **Die übrigen Seiten am ausgelieferten Bündel.** Bedienen ließ es sich nicht:
  Das Freigabebündel hat keinen Steuerkanal. Die übrigen Seiten, hell und
  dunkel, belegt die Hüllenprobe am selben Quelltext.
- **Den Gatekeeper-Weg** (Handprüfung 6). Das Abbild trug hier keine
  Quarantäne, und der Weg über die Systemeinstellungen bleibt einem Menschen
  vorbehalten.
- **Den Druckbogen** (Handprüfung 5).
- **Den Bestand eines echten Rechners**, der mit 0.7.0 gearbeitet hat. Die
  Nachstellung hat nur zwei eingefahrene Sets, keine Sammlung, keine Partien
  und keine Decks.
