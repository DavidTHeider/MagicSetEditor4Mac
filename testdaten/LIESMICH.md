# Testpakete

**Selbst geschrieben, nicht bezogen.** Fremde Vorlagensammlungen dürfen nicht in
dieses Repository ([`docs/entscheidungen/0005`](../docs/entscheidungen/0005-keine-fremden-vorlagen.md)):
sie stehen unter keiner Lizenz und enthalten geschützte Grafiken.

Die Pakete hier tragen dieselbe **Syntax** wie echte MSE-Pakete und eigenen
**Inhalt**. Jede Eigenheit, die sie nachstellen, ist in
[`docs/FORMATE.md`](../docs/FORMATE.md) mit Quelle belegt — was hier steht, ist
also keine Erfindung, sondern ein Nachbau des Belegten.

| Paket | Stellt nach |
| --- | --- |
| `wiederholung.mse-style` | denselben Schlüssel mehrfach (`depends on:`) — der Fall, an dem eine Map scheitert |
| `tief.mse-style` | mehrere Einrückungsebenen und einen leeren Block |
| `zeilenenden.mse-style` | gemischte Zeilenenden in **einer** Datei, dazu die BOM |
| `vierklassen.mse-set` | die vier Fälle aus #88: ein zweites `game:`, ein zweites `set info:`, ein Kommentar auf der Wurzelebene, ein `card:` mit skalarem Rest |
| `probe-symbole.mse-symbol-font` | eine Symbolschrift samt `insert symbol menu:` — alle vier Arten aus `InsertSymbolMenu::Type` nebeneinander, dazu die Kurzform `item: W` |
| `probe-ohnemenue.mse-symbol-font` | eine Symbolschrift **ohne** Einfügemenü — der Fall, in dem es keine Palette gibt |
| `probe-stil.mse-style` | einen `card style:` mit `right` statt `width`, Skripten in Zahlenfeldern und `z index` gegen die Textreihenfolge |
| `skriptblock.mse-style` | einen mehrzeiligen Skriptblock — die Klammer über vier Zeilen, ein Aufruf mit Doppelpunkt darin, ein Feld ohne `card dpi` |

## `eigenwillig.mse-set` stellt nichts nach

Es ist das eine Prüfstück, für das die Regel oben **nicht** gilt, und der Grund
gehört hierher statt in eine Fußnote.

Die Byte-Zusage aus #48
ist keine Aussage über das Format, sondern über **unser Verhalten**: was
hereinkommt, kommt unverändert wieder heraus — gleich, ob MSE2 eine solche Datei
je geschrieben hat. Ein Prüfstück, das nur belegte Eigenheiten trägt, prüft
diese Zusage nicht: `\tname: Feuerball` setzt sich beim Nachbau zufällig genau so
wieder zusammen, und die Gegenprobe — alle Rohzeilen verwerfen — blieb über
`probe.mse-set` und `karg.mse-set` **grün**.

Deshalb trägt `eigenwillig.mse-set` mehrfache Abstände hinter dem Doppelpunkt,
einen Tabulator als Trenner, ein `\r\n` mitten in einer Karte, einen Kommentar
und eine Leerzeile zwischen Feldern und keinen letzten Umbruch. Nichts davon ist
eine Formatbehauptung — es sind Eingaben, und eine Zusage über Eingaben muss die
unbequemen enthalten.

Wer gegen ein **bezogenes** Paket prüfen will, tut das hinter dem Build-Tag
`fremdpakete` (`AGENTS.md` §3). Der läuft nicht in der Pipeline und zählt nicht
zur Abdeckung.

## `absturzspuren/` — was Go bei einem Absturz schreibt

**Von Go 1.27.2 geschrieben, nicht nachgebaut** — die Ausgabe auf die
Fehlerausgabe, dieselbe, die `debug.SetCrashOutput` in die Spur einer Sitzung
gibt. Je Datei ein Fall aus `erzeuger.go.txt`, gebaut mit `-trimpath` und
gelaufen am 2026-10-14. In jedem freien Wert steht `hunter2` oder `4711`;
`internal/app/absturzspurecht_intern_test.go` sagt zu, dass davon nichts ins
Diagnosepaket geht und der Stapel bleibt (Nachlauf über Phase 1 von 0.9.0,
Befund 5). **Die Marken am Kopf einer Goroutine** (`tracebacklabels`) zeigte
Go hier nicht; ihre Gestalt steht aus `runtime/traceback.go` im Test daneben.

## `fehlerdateien/` — der Regressionsbestand für kaputte Dateien

**Selbst geschrieben, je eine Fehlerklasse**, damit ein Fund aus einem fremden
Paket hier nachgebaut werden kann, ohne das Paket einzuchecken (FMT-06.AK-04,
#1695).
`internal/app/fehlerdateien_test.go` liest jede ein und sagt zu: ein
**benannter** Fehler oder ein Set samt dem, was nicht gelesen wurde — nie
„kein Fehler, keine Karte, kein Befund". Eine Datei ohne Eintrag in seiner
Tabelle macht ihn rot.

| Datei | Stellt nach | Ergibt |
| --- | --- | --- |
| `kein-utf8.mse-set` | ein ungültiges Byte in Zeile 5 | Ablehnung mit Datei, Zeile und Spalte |
| `utf16.mse-set` | eine Datei in UTF-16 | Ablehnung, die UTF-16 nennt |
| `abgeschnitten.mse-set` | ein Archiv, das mittendrin endet | Ablehnung: kein lesbares Archiv |
| `leer.mse-set`, `nur-bom.mse-set` | eine leere Hauptdatei, eine mit nur der BOM | Ablehnung: das Set nennt kein Spiel |
| `kaputte-karte.mse-set` | eine Zeile ohne Doppelpunkt in einer Karte | Set mit zwei Karten; die Zeile reist als Vorbehalt mit ([0014](../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)) |

**Wer eine neue Klasse findet**, legt sie hier als Minimaldatei ab und trägt
ihre Erwartung in die Tabelle des Tests ein.

## `auspackkosten/` — die teuren Fälle beim Auspacken

Selbst erzeugte Bilder von 512 × 512 Bildpunkten in einer Farbe, für die
Messung in `internal/msepaket/auspackkosten_test.go`
([0160](../docs/entscheidungen/0160-was-ein-bild-ausgepackt-kostet.md)). Hier
liegen nur die Fälle, die die Standardbibliothek nicht schreibt:

| Datei | Stellt nach |
| --- | --- |
| `grau-progressiv.jpg`, `ycbcr444-progressiv.jpg`, `cmyk-progressiv.jpg` | progressive JPEG — je Farbauszug vier Byte Koeffizienten, geschrieben mit PIL |
| `cmyk.jpg` | ein Grundlinien-JPEG in CMYK |
| `rgba16-adam7.png` | ein PNG mit 16 Bit je Kanal und Zeilensprung — von Hand aus zlib und den sieben Durchgängen gebaut |

## `fassung-0.5.0/` und `fassung-0.7.0/` stellen auch nichts nach

**Es sind die Originale**: Sicherung, Einstellungsdatei und Umschlag hat die
Fassung 0.5.0 selbst geschrieben, damit diese Fassung beweist, dass sie sie
liest (ENG-06.AK-03,
#1731) —
**die Sicherung von dort wird seit Vermerk 0189 mit Rat abgewiesen**, denn sie
steht vor der Grundlinie. Die Sicherung in `fassung-0.7.0/` hat 0.7.0
geschrieben, und sie geht ohne einen Schritt auf (J02,
#2266). Wie
sie entstanden sind, steht im `LIESMICH.md` daneben; wer sie neu schreibt,
schreibt sie mit derselben alten Fassung und nicht mit der heutigen.
