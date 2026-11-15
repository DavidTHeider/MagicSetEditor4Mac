# Beleg: Sichtprüfung und Bestand am ausgelieferten Abbild von 0.9.0

**Gefahren am 2026-10-21**, vor der Veröffentlichung, am Abbild
`MagicSetEditor4Mac-0.9.0.dmg` (SHA-256
`c7185efebb96fc3f986ee6f4048d1eddaac24d7c4a0064ad77d1c7cb41f8cdb2`, Commit
`3af5055ee`), auf diesem Rechner (Apple M5, macOS 27.0.1). Es sind die
Handprüfungen 3, 4 und 7 des Freigabebelegs. Der Weg ist derselbe wie bei
0.8.0 (`freigabebuendel-0.8.0.md`).

## Wie, ohne den Rechner anzufassen

- Das Abbild ist schreibgeschützt eingehängt
  (`hdiutil attach -readonly -nobrowse`), und die Anwendung wurde **aus ihm**
  gestartet.
- `HOME` und `CFFIXED_USER_HOME` zeigten auf ein Wegwerfzuhause,
  `MSE_SCHLUESSELBUND` auf einen eigenen Bund.
- **Kein Dateidialog** wurde geöffnet.
- Erscheinungsbild, Sprache, die „Ersten Schritte“ und das Öffnen des Sets hat
  das `mse` aus dem Abbild im selben Zuhause gesetzt.
- Die Bilder stammen von `screencapture -l`, gemacht vom Fenster des
  Prozesses; die Fensternummer kommt aus `CGWindowListCopyWindowInfo`, weil
  das Freigabebündel keinen Steuerkanal hat.
- **Vorher und nachher verglichen**: `Application Support`, `Preferences`,
  `WebKit`, `Caches`, `HTTPStorages` und die Absturzberichte der echten
  Anwendung, gegen eine Zeitmarke vor dem ersten Start.

## Was zu sehen war (Handprüfungen 3 und 4)

| Bild | Was es zeigt |
| --- | --- |
| Erster Start, hell, Deutsch | Die Startseite mit den „Ersten Schritten“ darüber: „Acht Vorlagenpakete sind schon da“. |
| Nach dem Öffnen des mitgelieferten `Vanguard.mse-set`, hell | 1 Set, 107 Karten; „Probleme: 1 von 1 Sets nicht geprüft“ mit „Jetzt prüfen“; „Was Aufmerksamkeit braucht“: nichts. Lesbar und geordnet, die Leiste mit allen Arbeitsbereichen und Kürzeln. |
| Dasselbe, **dunkel** | Dieselbe Seite im dunklen Erscheinungsbild, lesbar. Der Hauptknopf steht hell auf dunkel. |
| `mse zeichnen` aus dem Abbild, erste Karte | *Akroma, Angel of Wrath Avatar*: Rahmen, Bild, **Titel, Typzeile, Regeltext und Zahlen in den mitgelieferten Schriften**, 375 × 523 in 0,67 s. |

## Der Bestand nach dem Start (Handprüfung 7, nachgestellt)

**Nachgestellt, nicht am echten Bestand.** Den echten Bestand dieses Rechners
zu öffnen, deckt keine Freigabe (`docs/FREIGABEN.md`).

1. Das `mse` aus dem Tag `v0.8.0`, aus dem Quelltext des Tags gebaut, legte
   im Wegwerfzuhause einen Bestand an: das mitgelieferte `Vanguard.mse-set`
   und `saat/probe-werkstatt.mse-set`, also **2 Sets mit 112 Karten** und 97
   verbuchten Wanderungen.
2. Darauf wurde 0.9.0 aus dem Abbild gestartet.

**Ergebnis**:
- 0.9.0 **sicherte den Bestand vor dem Wandern** (`bestand-vor-0.9.0.db`,
  darin 97 Wanderungen, 2 Sets, 112 Karten) und hob ihn auf **101
  Wanderungen** (`0098` bis `0101`).
- Die Startseite zeigt **2 Sets und 112 Karten**.
- Im Protokoll steht weder eine Warnung noch ein Fehler.

## Was sich an den echten Orten änderte

**Nichts — bis auf die Vorgaben unter der neuen Kennung.**
`~/Library/Preferences/io.github.davidtheider.magicseteditor4mac.plist` wurde
beim ersten Start geschrieben. Der Grund ist der Umzug der Kennung (R02,
Vermerk 0191, `internal/app/kennungsumzug.go`): er zieht die Vorgaben der
alten Kennung über `defaults` in die neue Domäne, und `defaults` spricht mit
dem `cfprefsd` des angemeldeten Nutzers — **das umgelenkte Zuhause hält ihn
nicht auf**.

Die Vorgaben danach sind **gleich den alten** (`com.wails.MagicSetEditor4Mac`,
beide Schreibweisen: vier Schlüssel, je gleich), dazu die Marke
`Kennungsumzug`. Das ist dasselbe, was der erste echte Start täte; der alte
Ordner bleibt, wie zugesagt, liegen. Der **Webspeicher** zieht über eine
Markendatei im neuen Ordner um, unter dem Zuhause, das der Start nennt — er
lag hier im Wegwerfzuhause, und der erste echte Start zieht den echten um.

**Für einen nächsten Lauf**: wer eine Freigabe am Abbild prüft, prüft die
Vorgaben damit am echten Nutzer; eine Isolation gäbe erst ein eigener Nutzer
oder ein Gast.

## Was dieser Beleg nicht sagt

- **Die übrigen Seiten am ausgelieferten Bündel.** Bedienen ließ es sich nicht:
  Das Freigabebündel hat keinen Steuerkanal. Die übrigen Seiten, hell und
  dunkel, belegt die Hüllenprobe am selben Quelltext.
- **Den Gatekeeper-Weg** (Handprüfung 6). Das Abbild trug hier keine
  Quarantäne, und der Weg über die Systemeinstellungen bleibt einem Menschen
  vorbehalten.
- **Den Druckbogen** (Handprüfung 5).
- **Den Bestand eines echten Rechners**, der mit 0.8.0 gearbeitet hat. Die
  Nachstellung hat nur zwei eingefahrene Sets, keine Sammlung, keine Partien
  und keine Decks.
