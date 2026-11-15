# Setdatei-Rundlauf gegen MSE2 — F05

#2247 · Lauf `TestSetrundlaufMitMSE2` (`internal/flaechen/setrundlauf_fremd_test.go`, Tag `fremdpakete`).

## Lauf

| Angabe | Wert |
| --- | --- |
| Commit | `d7a41e462334f334d0d93ba243877bab3b86a409`, sauberer Baum (eigener Arbeitsbaum) |
| Befehl | `MSE_FREMDPAKETE=/Users/david/full-magic-pack/data go test -tags fremdpakete ./internal/flaechen/ -run TestSetrundlaufMitMSE2 -v -count=1` |
| Zeit | 2026-10-10 21:04:51–21:08:29 UTC, Exitstatus 0, 142 s |
| Maschine | Apple M5, macOS 27.0.1, Last 11,4 (die Fuzzkampagne lief daneben) |
| Orakel | `/Users/david/mse2-quelle/build/magicseteditor`, Fassung f9d9356, sha256 `6c414114a57c6955…` |
| Spiel `magic` | `full-magic-pack/data/magic.mse-game/game`, sha256 `e280c2c45dc405db…` (`mse version: 2.5.8`) |
| Archenemy.mse-set | sha256 `812f3aa3da96a0ca…`, 8824637 Bytes |
| Planechase.mse-set | sha256 `112566f4b051bd5e…`, 17016561 Bytes |
| Vanguard.mse-set | sha256 `8a00d53610141e47…`, 9055706 Bytes |
| White.mse-set | sha256 `ac4cb04c6f05119a…`, 293778578 Bytes |
| eigene Saaten | vom Lauf geschrieben (`setrundlaufsaat_fremd_test.go`): `magic`/`m15` und `saga`/`standard`, je zwei Karten und ein Bild |

## Ergebnis

**Alle sechs Sets bestehen.** Spalten mit „1:" sind Richtung 1 (diese Anwendung schreibt, MSE2 liest), „2:" Richtung 2 (MSE2 schreibt, diese Anwendung liest); die Spalten „MSE2:" zählen, was MSE2 selbst beim Schreiben weglässt oder umschreibt — sie machen den Lauf nicht rot.

| Set | Karten | Werte | davon Unicode | gerechnet (MSE2) | Dateien | Stiloption | 1: Abweichungen | 1: Schlüssel weg | 1: Dateien weg/anders | 1: neue Warnungen | 2: Abweichungen | MSE2: Schlüssel weg | MSE2: Dateien weg | MSE2: Werte geändert |
| --- | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| eigene Saat (magic) | 2 | 11 | 3 | 5 | 1 | `tap symbol` gelesen | 0 | 0 | 0/0 | 0 | 0 | 2 | 0 | 7 |
| eigene Saat (saga) | 2 | 8 | 2 | 30 | 1 | `tap symbol` gelesen | 0 | 0 | 0/0 | 0 | 0 | 3 | 0 | 7 |
| Archenemy | 70 | 1260 | 45 | 0 | 70 | `chop top` gelesen | 0 | 0 | 0/0 | 0 | 0 | 0 | 0 | 330 |
| Planechase | 86 | 1626 | 42 | 8 | 86 | `chop top` gelesen | 0 | 0 | 0/0 | 0 | 0 | 0 | 0 | 503 |
| Vanguard | 107 | 1926 | 51 | 0 | 107 | `chop top` gelesen | 0 | 0 | 0/0 | 0 | 0 | 0 | 0 | 434 |
| White | 3761 | 68591 | 2955 | 130821 | 3801 | `chop top` gelesen | 0 | 0 | 0/0 | 0 | 0 | 0 | 0 | 24293 |

**eigene Saat (magic), was MSE2 nicht mitschreibt**: card#0/eigener_kartenschluessel · set_info#0/eigener_setschluessel · Karte 0, has styling: ab Zeichen 0 "" → "false" · Karte 0, rule text: ab Zeichen 78 "nführungen“ und 𝄞." → "nführungen” und 𝄞." · Karte 0, time created: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, has styling: ab Zeichen 0 "" → "false" · Karte 1, time created: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00"

**eigene Saat (saga), was MSE2 nicht mitschreibt**: card#0/supertype · set_info#0/eigener_setschluessel · card#0/eigener_kartenschluessel · Karte 0, flavor text: ab Zeichen 0 "Ein Satz." → "<i-flavor>Ein Satz.</i-flavor>" · Karte 0, has styling: ab Zeichen 0 "" → "false" · Karte 0, time created: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, has styling: ab Zeichen 0 "" → "false" · Karte 1, time created: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00"

**Archenemy, was MSE2 nicht mitschreibt**: Karte 0, flavor text: ab Zeichen 10 "<i-flavor>\"What would you say is my grea" → "<i-flavor>“What would you say is my grea" · Karte 0, stylesheet version: ab Zeichen 2 "2010-05-20" → "2024-05-30" · Karte 0, styling data: ab Zeichen 0 "" → "text_box_mana_symbols: magic-m" · Karte 0, super type: ab Zeichen 0 "Scheme" → "<word-list-type-en>Scheme</wor" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, flavor text: ab Zeichen 10 "<i-flavor>\"The gates open to my new era." → "<i-flavor>“The gates open to my new era." · Karte 1, stylesheet version: ab Zeichen 2 "2010-05-20" → "2024-05-30" · Karte 1, styling data: ab Zeichen 0 "" → "text_box_mana_symbols: magic-m"

**Planechase, was MSE2 nicht mitschreibt**: Karte 0, rule text 2: ab Zeichen 0 "Whenever you roll <sym>chaos</" → "<margin:480:0:0>Whenever you r" · Karte 0, stylesheet version: ab Zeichen 2 "2018-06-01" → "2024-09-23" · Karte 0, styling data: ab Zeichen 0 "" → "text_box_mana_symbols: magic-m" · Karte 0, sub type: ab Zeichen 0 "Dominaria" → "<word-list-plane-en>Dominaria<" · Karte 0, super type: ab Zeichen 0 "Plane" → "<word-list-type-en>Plane</word" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, rule text 2: ab Zeichen 0 "Whenever you roll <sym>chaos</" → "<margin:480:0:0>Whenever you r" · Karte 1, stylesheet version: ab Zeichen 2 "2018-06-01" → "2024-09-23"

**Vanguard, was MSE2 nicht mitschreibt**: Karte 0, stylesheet version: ab Zeichen 3 "2020-03-20" → "2024-09-01" · Karte 0, styling data: ab Zeichen 0 "" → "text_box_mana_symbols: magic-m" · Karte 0, super type: ab Zeichen 0 "Vanguard" → "<word-list-type-en>Vanguard</w" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, stylesheet version: ab Zeichen 3 "2020-03-20" → "2024-09-01" · Karte 1, styling data: ab Zeichen 0 "" → "text_box_mana_symbols: magic-m" · Karte 1, super type: ab Zeichen 0 "Vanguard" → "<word-list-type-en>Vanguard</w" · Karte 1, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00"

**White, was MSE2 nicht mitschreibt**: Karte 0, casting cost: ab Zeichen 0 "[1][W]" → "1W" · Karte 0, stylesheet version: ab Zeichen 3 "2020-09-04" → "2025-10-05" · Karte 0, styling data: ab Zeichen 0 "text_box_mana_symbols: magic-m" → "casting_cost_mana_symbols: mag" · Karte 0, sub type: ab Zeichen 0 "Equipment" → "<word-list-artifact>Equipment<" · Karte 0, super type: ab Zeichen 0 "Artifact" → "<word-list-type-en>Artifact</w" · Karte 0, time modified: ab Zeichen 0 "" → "2026-10-09 00:00:00" · Karte 1, casting cost: ab Zeichen 0 "[1][W]" → "1W" · Karte 1, flavor text: ab Zeichen 10 "<i-flavor>\"What could be funnier than on" → "<i-flavor>“What could be funnier than on"

## Gegenproben

Am selben Abend gegen den Schreiber **vor** der Berichtigung, je nur Archenemy:

- ohne die Stiloptionen im Block `styling:` — rot: „MSE2 liest die setweite Stiloption der Abschrift nicht — styling.chop top ist "", gesetzt war "7"",
- ohne das Entfernen der Zeile unter `set info:` — rot: „MSE2 warnt über die Abschrift, was es über das Original nicht tat — Unexpected key: 'styling'". Der erste Anlauf dieses Wächters las nur den Fehlerstrom und blieb grün; im Rohmodus stehen die Warnungen auf der Standardausgabe.

## Grenzen

- **Ob MSE2 die Abschrift zeichnet wie das Original**, sagt dieser Lauf nicht — das belegt der Kanonlauf.
- **Schlüssel unterhalb der zweiten Ebene** zählt er nicht (`styling data` einer Karte ist ein Block).
- **Ein Spiel außer `magic` und `saga`** ist nicht dabei; alle zehn kanonischen Sets sind `magic`.
- **Bilder** vergleicht er an den Dateien, nicht über MSE2: ein Bildwert ist im Orakel kein Text.
