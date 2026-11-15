# Beleg: I02 — der Kernweg auf macOS 15.2

**Stand:** Commit `2ca9a871c` · **Datum:** 2026-10-11 06:14 UTC · **Befehl:** `node scripts/untergrenzegast.mjs mse080-152` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse080-152` — macOS 15.2 (24C101), VirtualMac2,1, 4 Kerne, 4096 MB; Last { 32.14 6.76 2.39 }, beim Messen { 2.13 2.53 2.83 }.

**Bündel:** Abnahmebau, Baustand `1965cca4f47c`, Programm `1318105eb7adfd33…`, gebaut mit `minos 15.2`.

**Ergebnis:** alle 11 Schritte gelungen.

| Fläche | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Kommandozeile | Hilfe | gelungen | 1264 | 1 |
| Kommandozeile | Set einfahren | gelungen | 159 | 1 |
| Kommandozeile | Karten lesen | gelungen | 49 | 1 |
| Kommandozeile | Feld setzen | gelungen | 97 | 1 |
| Kommandozeile | Karte zeichnen | gelungen | 341 | 38489 |
| Kommandozeile | Set sichern | gelungen | 50 | 1 |
| Fenster | Start bis stehende Oberfläche | gelungen | 1996 | 1995.40307617188 |
| Fenster | Set aufschlagen | gelungen | 388 | 1 |
| Fenster | Feld ändern und Vorschau neu | gelungen | 170 | 375×523 |
| Fenster | Wert im Bestand | gelungen | 76 | 1 |
| Fenster | ⌘Q beendet | gelungen | 111 | 1 |

- Die Seite zeichnet: 0 Bilder in 2200 ms (`requestAnimationFrame`), Sichtbarkeit »hidden«. Der Weg am Fenster geht über die Seite und nicht über die Maus — hinter dem Sperrbildschirm des Gastes ist das Fenster nicht das vordere.
