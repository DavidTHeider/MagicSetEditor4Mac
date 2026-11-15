# Beleg: zwei Bauläufe, Eintrag für Eintrag — I06

**Gefahren am 2026-10-12, 02:51–02:55**, an Commit `e48f6d905` (Fassung
0.8.0, Zweig `feat/2264-i06-herkunft`, sauberer Baum), auf diesem Rechner:
Mac17,3, macOS 27.0.1, Go 1.27.2 über `GOTOOLCHAIN`.

```text
./scripts/baugleich.sh zweimal <raum>
```

**Wie unabhängig**: zwei Arbeitsbäume an verschiedenen Pfaden
(`…/baugleich-raum/a` und `…/b`), jeder mit **eigenem
Go-Bauzwischenspeicher** (je 1,7 GB, neu angelegt) und eigenem
`node_modules` aus `npm ci`. Gebaut hat jeder mit `release.sh 0.8.0` unter
`MSE_FREIGABE_NUR_BAU=1`, also auf dem Freigabeweg ohne Pipeline, Stücklisten,
Prüfsummen und Signatur. Geteilt war allein der Modulspeicher, und der ist
nach Prüfsumme adressiert. Beide Läufe endeten mit Status 0.

## Ergebnis

| Vergleich | Einträge | gleich | anders | ungeklärt |
| --- | ---: | ---: | ---: | ---: |
| Bündel (`build/bin/MagicSetEditor4Mac.app`) | 4.473 | 4.473 | 0 | 0 |
| eingehängtes Abbild (`MagicSetEditor4Mac-0.8.0.dmg`) | 4.475 | 4.475 | 0 | 0 |

**Das Programm ist bytegleich**, samt Ad-hoc-Signatur:

```text
5f4b96692eeb020d4afeccc63d9169108c94778d9da30d4bd98ad7b5b19cd6b2  a/…/Contents/MacOS/MagicSetEditor4Mac
5f4b96692eeb020d4afeccc63d9169108c94778d9da30d4bd98ad7b5b19cd6b2  b/…/Contents/MacOS/MagicSetEditor4Mac
```

**Die zwei `.dmg`-Dateien selbst sind es nicht**: `a407761e…` gegen
`471966d4…`. `hdiutil create` schreibt in den Container eine Kennung des
Datenträgers und Zeitstempel; was darin liegt, ist Eintrag für Eintrag gleich.
**Darum gilt die Prüfsumme einer Freigabe dem Abbild, das hochgeladen ist**,
und `herkunft.json` nennt daneben die Prüfsummen jedes Programms. Die bleiben
über zwei Bauläufe gleich.

## Was er nicht zeigt

- **Eine zweite Maschine.** Beide Läufe teilen Rechner, Xcode-Werkzeuge,
  Systemfassung und Schriften. Ob ein anderer Mac dasselbe Programm baut, ist
  nicht gemessen.
- **Einen anderen Tag.** Beide Läufe liegen vier Minuten auseinander. Ein
  Zeitstempel, der nur das Datum trägt, fiele hier nicht auf. Dass keiner im
  Programm steht, sagt der Vergleich des Bündels nur für diesen Abstand.
