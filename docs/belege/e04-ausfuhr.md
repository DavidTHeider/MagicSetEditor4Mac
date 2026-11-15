# Beleg: E04 — der skriptlastige Export, `magic-namecheck`

**Stand:** vorher Commit `c1f56ed79` (main), nachher derselbe Stand mit den zwei Änderungen dieses Werkstücks · **Datum:** 2026-10-11 · **Maschine:** Apple M5, zehn Kerne, macOS 27.0.1, Last 2,8 bis 6,4 · **Lauf:** je Satz fünfmal abwechselnd MSE2 und MSE4Mac, Zeit von außen genommen (Aufbau unten)

**Aufbau wie im Vergleich** (`docs/LEISTUNGSVERGLEICH.md` §2): MSE2 `--export magic-namecheck.mse-export-template <set> <ziel>` mit eigenem `HOME` und `CFFIXED_USER_HOME`; MSE4Mac `mse vorlagen <set> ausfuehren magic-namecheck.mse-export-template <ziel>` mit eigenem `HOME`, Bestand und Bund, im Paketverzeichnis `full-magic-pack/data` **und** `Magic - Fonts`. Prozessstart bis zur geschriebenen Datei. Der Bestand trägt alle zehn kanonischen Sets.

## Ergebnis

| Satz | Karten | MSE2 (Median aller zehn) | MSE4Mac vorher | MSE4Mac nachher | nachher, je Lauf |
| --- | ---: | ---: | ---: | ---: | --- |
| Archenemy | 70 | 868 ms | 715 ms | **570 ms** | 1190, 569, 570, 570, 583 |
| Black | 3788 | 6251 ms | 28622 ms | **19615 ms** | 19611, 19615, 19668, 19688, 19601 |

**Die Ausgabe ist bytegleich** mit MSE2 — vorher wie nachher: Archenemy 354 Byte, SHA-256 `71650f70e66988a5…`; Black 58.127 Byte, `9fe74db3b3b562d7…`.

**Das Ziel aus E04 — höchstens 2 s gegen die bestätigte Kohorte — hielt schon am Ausgangsstand.** Die 6.717 ms aus der Vergleichstafel vom 2026-10-02 (`f749bfa4`) sind nachgestellt — 6,7 s, 11,7 s CPU, dieselbe Ausgabe — und **ein `git bisect` über 176 Commits** benennt den ersten schnellen: `32c7321b8`, der Squash von #2040 vom 2026-10-03 (gemerkte Muster, Ordnungsspeicher, Setzahl im Interpreter). Davor 6,7 s, ab ihm 0,9 s; die Vergleichstafel stammt vom Tag davor, und der Freigabelauf, der es gezeigt hätte, stand aus.

## Was dieses Werkstück änderte

1. **Ein Stil je Name und Lauf, nicht je Karte.** Das Profil über Black zeigte das Stilpaket je Karte geöffnet und geschlossen (`mitStilen` → `Stilwaehler.Stilpaket`); die Feldberechnung merkte ihre Antworten schon (Befund 3 eines früheren Nachlaufs) — behoben war der Fall, nicht die Klasse. Jetzt merkt `Fuehre` die Stile einmal je Lauf für alle drei Rufer (`Ausfuhrstile.gemerkt`). Black 28,6 → 22,2 s.
2. **Der Abbruch über den `Done`-Kanal.** `ctx.Err()` bei jedem Auswertungsschritt lief die Kette der Kontexte mit Werten hinauf — 8 % der Rechenzeit. Black 22,2 → 19,6 s, Archenemy 0,63 → 0,57 s.

## Was bleibt, und warum nicht hier

**Black braucht das Dreifache von MSE2.** Das Zuteilungsprofil nennt 40 GB für diesen Export: die Verkettung mit `+` (`verkettet`, 31 Millionen Rufe, rund 8.000 je Karte — die Wertskripte des Spiels reihen Dutzende `replace@` aneinander und laufen für jedes Feld jeder Karte), das Zusammenlegen benannter Argumente (`verbunden`) und neue Umgebungen je Aufruf. Der nächste Hebel ist, wie `verkettet` `input` weiterreicht — daran hängt, welches `input` eine eingebaute Funktion sieht, also **Skriptsemantik** und kein Umbau ohne Verhaltensänderung. Der Export bleibt dabei **linear und begrenzt**: rund 5 ms je Karte bei Black, ohne Ausreißer über fünf Läufe.

## Grenzen dieser Messung

- Gemessen sind zwei Sätze und eine Vorlage; `magic-simple-exporter`, `magic-text` und `magic-mws` stehen weiter im Freigabelauf (#1672).
- Der erste Lauf nach einem neuen Bau ist langsamer (Archenemy 1.190 ms), weil das System das neue Programm prüft — er ist im Median enthalten, nicht herausgerechnet.
- Das Profil stammt aus dem Lauf hinter `fremdpakete` (`cmd/mse/ausfuhrprofil_test.go`) — derselbe Weg ohne Prozessstart.
