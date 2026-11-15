# Beleg: Sichtprüfung und Bestand am ausgelieferten Abbild von 1.0.0

**Gefahren am 2026-11-15**, vor der Veröffentlichung, am Abbild
`MagicSetEditor4Mac-1.0.0.dmg` (SHA-256
`e530d575236824b22d8fc537cce8a13f35589d84ed9daf0575f00784c456b81f`, Commit
`11e52bd20`, gebaut mit go1.27.4), auf diesem Rechner (Apple M5, macOS 27.2).
Es sind die Handprüfungen 3, 4 und 7 des Freigabebelegs; der Weg ist derselbe
wie bei 0.9.0 (`freigabebuendel-0.9.0.md`).

## Wie, ohne den Rechner anzufassen

- Das Abbild ist schreibgeschützt eingehängt
  (`hdiutil attach -readonly -nobrowse`), und die Anwendung wurde **aus ihm**
  gestartet.
- `HOME` und `CFFIXED_USER_HOME` zeigten auf ein Wegwerfzuhause,
  `MSE_SCHLUESSELBUND` auf einen Bund, den es nicht gibt — so landet kein Bund
  in der Suchliste des echten Nutzers.
- **Kein Dateidialog** wurde geöffnet.
- Sprache, Erscheinungsbild, die „Ersten Schritte“ und das Öffnen des Sets hat
  das `mse` aus dem Abbild im selben Zuhause gesetzt.
- Die Bilder stammen von `screencapture -l` am Fenster des Prozesses; die
  Fensternummer kommt aus `CGWindowListCopyWindowInfo`, weil das
  Freigabebündel keinen Steuerkanal hat.
- **Vorher und nachher verglichen**: die Vorgaben unter beiden Kennungen
  (`plutil -p`), `Application Support`, `WebKit`, `Caches`, `HTTPStorages` und
  die Absturzberichte, gegen eine Zeitmarke vor dem ersten Start.

## Was zu sehen war (Handprüfungen 3 und 4)

| Bild | Was es zeigt |
| --- | --- |
| Erster Start, hell, Deutsch | Die Startseite mit den „Ersten Schritten“ darüber: „Acht Vorlagenpakete sind schon da“. |
| Nach dem Öffnen des mitgelieferten `Vanguard.mse-set`, hell | 1 Set, 107 Karten; „Probleme: 1 von 1 Sets nicht geprüft“ mit „Jetzt prüfen“; „Was Aufmerksamkeit braucht“: nichts. Lesbar und geordnet. |
| Dasselbe, **dunkel** | Dieselbe Seite im dunklen Erscheinungsbild, lesbar; der Hauptknopf steht hell auf dunkel. |
| `mse zeichnen` aus dem Abbild, erste Karte | *Akroma, Angel of Wrath Avatar*: Rahmen, Bild, **Titel, Typzeile, Regeltext und Zahlen in den mitgelieferten Schriften**, 375 × 523 in 0,57 s. |

## Der Bestand nach dem Start (Handprüfung 7, nachgestellt)

**Nachgestellt, nicht am echten Bestand.** Statt eines Bestands aus einem
alten Abbild dient die Sicherung, die 0.7.0 **selbst** geschrieben hat
(`testdaten/fassung-0.7.0/sicherung.zip`: 2 Sets, 4 Karten, 1 Deck, Sprache
`en`, dunkel, 97 Wanderungen), ausgepackt an den Ort, an dem eine installierte
Fassung ihren Bestand hat.

**Ergebnis**:
- 1.0.0 **sicherte den Bestand vor dem Wandern** (`bestand-vor-1.0.0.db`,
  darin 97 Wanderungen und 2 Sets) und hob ihn auf **101 Wanderungen**.
- Die Startseite zeigt **2 Sets, 4 Karten und 1 Deck**, englisch und dunkel,
  wie der alte Bestand es eingestellt hatte.
- Im Protokoll stehen drei Zeilen, keine Warnung und kein Fehler.

## Was sich an den echten Orten änderte

**Nichts.** Die Vorgaben unter `io.github.davidtheider.magicseteditor4mac` und
unter `com.wails.MagicSetEditor4Mac` sind vorher und nachher gleich; in
`Application Support`, `WebKit`, `Caches` und `HTTPStorages` liegt nichts, das
jünger ist als die Zeitmarke, und es gibt keinen Absturzbericht. Anders als bei
0.9.0 schrieb der erste Start keine Vorgaben: der Umzug der Kennung (Vermerk
0191) war auf diesem Rechner schon geschehen.

**Ein Aufruf ging am Wegwerfzuhause vorbei**: bei der Prüfung der Fassung im
Bündel lief `mse` aus dem Abbild einmal ohne umgelenktes Zuhause. Nachgesehen
danach: `bestand.db`, `mse.log` und die Vorgaben des echten Nutzers trugen
keine Änderung seit dem 31.10. — der Befehl hatte den Bestand nicht geöffnet.

## Die Handprüfungen 1 und 2

**Nicht neu gefahren.** `make fremdpakete` am Stand `11e52bd20` meldete
BLOCKED, 0 von 92 Fällen: zwei seiner zehn Eingaben lagen nicht mehr vor —
das gebaute MSE2 (`MSE2_BINAER`) und seine eigene Rasterung der kanonischen
Sets (`MSE_MSE2AUSGABE`). **Übernommen ist der Lauf zur Freigabe 0.9.0** am
Stand `3b89bed` (PASS, 83 von 92, Kanonmittel 9,319): zwischen ihm und
`11e52bd20` unterscheiden sich `internal/render`, `internal/service`,
`internal/msescript`, `internal/msepaket`, `internal/schriften`,
`internal/svg`, `internal/pdf`, `internal/mtgjson` und `cmd/mse` außerhalb der
Tests **nur in Kommentaren** (831 Dateien, keine Codezeile). Neu zu fahren,
sobald MSE2 nach Vermerk 0026 wieder gebaut ist.

## Was dieser Beleg nicht sagt

- **Die übrigen Seiten am ausgelieferten Bündel.** Bedienen ließ es sich
  nicht: das Freigabebündel hat keinen Steuerkanal.
- **Den Gatekeeper-Weg** (Handprüfung 6) und **den Druckbogen**
  (Handprüfung 5) — beide bleiben einem Menschen vorbehalten.
- **Den Bestand eines echten Rechners** mit Sammlung, Partien und Handel.
