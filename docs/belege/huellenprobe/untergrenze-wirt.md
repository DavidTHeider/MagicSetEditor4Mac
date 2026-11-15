# Beleg: I02 — der Kernweg am Wirt

**Commit:** `f3846894d` · **Datum:** 2026-10-11 03:20 UTC · **Befehl:** `node scripts/untergrenzegast.mjs --wirt` · **Rückgabe:** 0

**Maschine:** am Wirt (Positivkontrolle) — macOS 27.0.1 (26A434), Mac17,3, 10 Kerne, 32768 MB; Last { 3.51 5.63 10.77 }, beim Messen { 3.51 5.63 10.77 }.

**Bündel:** unverändert, Abnahmebau, Baustand `7a23d86e5b01`, Programm `66ec5913f08262cb…`, gebaut mit `minos 13.0`.

**Ergebnis:** alle 11 Schritte gelungen.

| Fläche | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Kommandozeile | Hilfe | gelungen | 816 | 1 |
| Kommandozeile | Set einfahren | gelungen | 109 | 1 |
| Kommandozeile | Karten lesen | gelungen | 20 | 1 |
| Kommandozeile | Feld setzen | gelungen | 41 | 1 |
| Kommandozeile | Karte zeichnen | gelungen | 143 | 38489 |
| Kommandozeile | Set sichern | gelungen | 20 | 1 |
| Fenster | Start bis stehende Oberfläche | gelungen | 566 | 541.35107421875 |
| Fenster | Set aufschlagen | gelungen | 367 | 1 |
| Fenster | Feld ändern und Vorschau neu | gelungen | 60 | 375×523 |
| Fenster | Wert im Bestand | gelungen | 23 | 1 |
| Fenster | ⌘Q beendet | gelungen | 103 | 1 |

- Die Seite zeichnet: 30 Bilder in 629 ms (`requestAnimationFrame`), Sichtbarkeit »visible«. Der Weg am Fenster geht über die Seite und nicht über die Maus — hinter dem Sperrbildschirm des Gastes ist das Fenster nicht das vordere.
