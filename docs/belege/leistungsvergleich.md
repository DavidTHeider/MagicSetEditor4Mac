# Leistungsvergleich — Bericht

**Stand `c42b1b4ef`**, Lauf ab 2026-10-05T02:09:56.832Z, Apple M5, 10 Kerne, macOS 27.0.1 (26A434). Erzeugt von
`scripts/leistungsvergleich.mjs`; die Zahlen stehen in
`leistungsvergleich.json`, die Bewertung in `docs/LEISTUNGSVERGLEICH.md`.

## Ganze Prozesse

Wandzeit in Millisekunden, CPU (`user` + `sys`) in Sekunden, residenter
Höchststand in MB (10⁶ Byte). *vorn / dahinter*: Median der Läufe, in denen
das Programm zuerst lief, gegen den der übrigen.

| Aufgabe | Programm | n | Median | p95 | p99 | erste | CPU | RSS | Fehler | Abstürze | vorn / dahinter |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| start | mse2 | 6 | 59,5 | 117,2 | 117,2 | 117,2 | 0,02 | 41,1 | 0 | 0 | 62,9 / 59 |
| start | heute | 6 | 23,5 | 23,8 | 23,8 | 23,8 | 0,01 | 40,7 | 0 | 0 | 23,1 / 23,5 |
| start | v050 | 6 | 21,8 | 26,6 | 26,6 | 21,8 | 0,01 | 35,8 | 0 | 0 | 22,2 / 21,7 |
| oeffnen · Archenemy | mse2 | 6 | 489,7 | 498,4 | 498,4 | 498,4 | 0,44 | 66,7 | 0 | 0 | 496,7 / 488,3 |
| oeffnen · Archenemy | heute | 6 | 130,7 | 132,8 | 132,8 | 131,4 | 0,12 | 45,2 | 0 | 0 | 128,7 / 130,7 |
| oeffnen · Archenemy | v050 | 6 | 108,9 | 113 | 113 | 109,9 | 0,1 | 39,5 | 0 | 0 | 108,7 / 108,9 |
| wiederoeffnen · Archenemy | heute | 6 | 25 | 28,1 | 28,1 | 25 | 0,01 | 41,7 | 0 | 0 | 25,1 / 24,9 |
| wiederoeffnen · Archenemy | v050 | 6 | 23,5 | 27,1 | 27,1 | 27,1 | 0,01 | 37,2 | 0 | 0 | 23,5 / 24,5 |
| sichern · Archenemy | mse2 | 6 | 668,1 | 689,4 | 689,4 | 654,4 | 0,62 | 66,8 | 0 | 0 | 654,4 / 674,3 |
| sichern · Archenemy | heute | 6 | 41,3 | 84,2 | 84,2 | 84,2 | 0,04 | 47,2 | 0 | 0 | 38,5 / 43,5 |
| sichern · Archenemy | v050 | 6 | 38,7 | 87,6 | 87,6 | 87,6 | 0,03 | 41,8 | 0 | 0 | 38,7 / 36,2 |
| ausfuhr magic-simple-exporter · Archenemy | mse2 | 6 | 489,8 | 494,2 | 494,2 | 493,2 | 0,45 | 67,0 | 0 | 0 | 489,8 / 488 |
| ausfuhr magic-simple-exporter · Archenemy | heute | 6 | 798 | 832,3 | 832,3 | 832,3 | 1,57 | 113,2 | 0 | 0 | 786,1 / 798 |
| ausfuhr magic-simple-exporter · Archenemy | v050 | 6 | 217,7 | 240,4 | 240,4 | 240,4 | 0,34 | 99,9 | 0 | 0 | 216,5 / 217,7 |
| ausfuhr magic-namecheck · Archenemy | mse2 | 6 | 898,4 | 919,4 | 919,4 | 912,1 | 0,85 | 137,5 | 0 | 0 | 898,4 / 883,5 |
| ausfuhr magic-namecheck · Archenemy | heute | 6 | 934,6 | 1.116,2 | 1.116,2 | 934,6 | 1,78 | 232,9 | 0 | 0 | 919,7 / 942,7 |
| ausfuhr magic-namecheck · Archenemy | v050 | 6 | 6.706,8 | 6.761,7 | 6.761,7 | 6.706,8 | 12,32 | 253,3 | 0 | 0 | 6.675,9 / 6.706,8 |
| ausfuhr magic-text · Archenemy | mse2 | 6 | 491,1 | 522,7 | 522,7 | 491,1 | 0,45 | 66,8 | 0 | 0 | 489,4 / 491,8 |
| ausfuhr magic-text · Archenemy | heute | 6 | 802 | 851,5 | 851,5 | 811,5 | 1,57 | 112,4 | 0 | 0 | 797,4 / 811,5 |
| ausfuhr magic-text · Archenemy | v050 | 6 | 224,5 | 248,4 | 248,4 | 220,9 | 0,34 | 103,3 | 0 | 0 | 219,6 / 225,2 |
| ausfuhr magic-mws · Archenemy | mse2 | 6 | 935,6 | 963,8 | 963,8 | 963,8 | 0,86 | 239,5 | 0 | 0 | 935,6 / 930,6 |
| ausfuhr magic-mws · Archenemy | heute | 6 | 6.307,7 | 6.557,8 | 6.557,8 | 6.557,8 | 7,63 | 628,9 | 0 | 0 | 6.266,7 / 6.323,5 |
| ausfuhr magic-mws · Archenemy | v050 | 6 | 6.001,1 | 6.271,7 | 6.271,7 | 6.271,7 | 6,78 | 622,3 | 0 | 0 | 6.021,4 / 5.990,4 |
| skript-sitzung · Archenemy | mse2 | 2 | 521,4 | 522 | 522 | 522 | 0,44 | 66,7 | 0 | 0 | 522 / 521,4 |
| skript-sitzung · Archenemy | heute | 2 | 541,1 | 1.184,9 | 1.184,9 | 1.184,9 | 0,81 | 503,4 | 0 | 0 | 541,1 / 1.184,9 |
| skript-sitzung · Archenemy | v050 | 2 | 5.223,5 | 5.555,3 | 5.555,3 | 5.555,3 | 8,19 | 512,3 | 0 | 0 | — / 5.223,5 |
| navigation-sitzung · Multicolor | mse2 | 2 | 6.552,1 | 6.643,7 | 6.643,7 | 6.643,7 | 6,44 | 408,0 | 0 | 0 | 6.643,7 / 6.552,1 |
| navigation-sitzung · Multicolor | heute | 2 | 4.013,9 | 4.217,2 | 4.217,2 | 4.217,2 | 4,33 | 525,5 | 0 | 0 | 4.013,9 / 4.217,2 |
| navigation-sitzung · Multicolor | v050 | 2 | 28.585,4 | 28.924 | 28.924 | 28.585,4 | 35,25 | 542,6 | 0 | 0 | — / 28.585,4 |
| navigation-sitzung · Archenemy | mse2 | 2 | 775,5 | 775,6 | 775,6 | 775,6 | 0,71 | 100,5 | 0 | 0 | 775,6 / 775,5 |
| navigation-sitzung · Archenemy | heute | 2 | 1.105,3 | 1.112,9 | 1.112,9 | 1.105,3 | 1,31 | 502,2 | 0 | 0 | 1.112,9 / 1.105,3 |
| navigation-sitzung · Archenemy | v050 | 2 | 6.947,3 | 7.095,1 | 7.095,1 | 6.947,3 | 8,68 | 512,6 | 0 | 0 | — / 6.947,3 |
| navigation-sitzung · Planechase | mse2 | 2 | 1.058 | 1.066,4 | 1.066,4 | 1.066,4 | 0,97 | 127,7 | 0 | 0 | 1.066,4 / 1.058 |
| navigation-sitzung · Planechase | heute | 2 | 1.881,6 | 1.931,4 | 1.931,4 | 1.931,4 | 2,12 | 506,9 | 0 | 0 | 1.881,6 / 1.931,4 |
| navigation-sitzung · Planechase | v050 | 2 | 10.202,1 | 10.521 | 10.521 | 10.521 | 12,3 | 512,5 | 0 | 0 | — / 10.202,1 |
| navigation-sitzung · Vanguard | mse2 | 2 | 720,5 | 740,3 | 740,3 | 740,3 | 0,66 | 97,4 | 0 | 0 | 740,3 / 720,5 |
| navigation-sitzung · Vanguard | heute | 2 | 1.115,5 | 1.121 | 1.121 | 1.115,5 | 1,36 | 501,8 | 0 | 0 | 1.121 / 1.115,5 |
| navigation-sitzung · Vanguard | v050 | 2 | 8.611,5 | 9.025,5 | 9.025,5 | 9.025,5 | 10,69 | 511,0 | 0 | 0 | — / 8.611,5 |
| bilder · Archenemy | mse2 | 1 | 3.137,5 | 3.137,5 | 3.137,5 | 3.137,5 | 3,03 | 238,9 | 0 | 0 | 3.137,5 / — |
| bilder · Archenemy | heute | 1 | 6.137,5 | 6.137,5 | 6.137,5 | 6.137,5 | 6,65 | 618,8 | 0 | 0 | — / 6.137,5 |
| bilder · Archenemy | v050 | 1 | 68.122,5 | 68.122,5 | 68.122,5 | 68.122,5 | 86,82 | 581,1 | 0 | 0 | — / 68.122,5 |
| oeffnen · Black | mse2 | 3 | 5.615,9 | 5.634,3 | 5.634,3 | 5.615,9 | 5,57 | 305,4 | 0 | 0 | 5.615,9 / 5.615,7 |
| oeffnen · Black | heute | 3 | 1.962,6 | 1.972,1 | 1.972,1 | 1.962,6 | 2,02 | 108,6 | 0 | 0 | 1.927,5 / 1.962,6 |
| oeffnen · Black | v050 | 3 | 1.921,2 | 1.936,3 | 1.936,3 | 1.936,3 | 1,99 | 99,5 | 0 | 0 | 1.921,2 / 1.910,9 |
| wiederoeffnen · Black | heute | 3 | 73,8 | 77,1 | 77,1 | 77,1 | 0,09 | 86,3 | 0 | 0 | 71,5 / 73,8 |
| wiederoeffnen · Black | v050 | 3 | 73,5 | 75 | 75 | 73,5 | 0,09 | 86,2 | 0 | 0 | 71,7 / 73,5 |
| sichern · Black | mse2 | 3 | 11.990,1 | 12.034,7 | 12.034,7 | 12.034,7 | 11,92 | 306,6 | 0 | 0 | 12.034,7 / 11.966,4 |
| sichern · Black | heute | 3 | 781,8 | 2.278,4 | 2.278,4 | 2.278,4 | 1,05 | 138,9 | 0 | 0 | 781,8 / 751,8 |
| sichern · Black | v050 | 3 | 715,8 | 2.272,6 | 2.272,6 | 2.272,6 | 0,81 | 134,5 | 0 | 0 | 671,2 / 715,8 |

## Schritte in einer laufenden Sitzung

Je Schritt von der geschriebenen Zeile bis zur Antwort, in Millisekunden.

| Programm · Zustand · Schicht | n | Median | p95 | p99 | max | Fehler |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| mse2 · kalt · Skript: Rechnung | 2 | 0 | 0,1 | 0,1 | 0,1 | 0 |
| mse2 · warm · Skript: Rechnung | 8 | 0 | 0,1 | 0,1 | 0,1 | 0 |
| mse2 · warm · Skript: Zeichenkette | 10 | 0 | 0 | 0 | 0 | 0 |
| mse2 · warm · Skript: Kartenfeld | 10 | 0 | 0,1 | 0,1 | 0,1 | 0 |
| heute · kalt · Skript: Rechnung | 2 | 289 | 297,5 | 297,5 | 297,5 | 0 |
| heute · warm · Skript: Rechnung | 8 | 12,7 | 13,3 | 13,3 | 13,3 | 0 |
| heute · warm · Skript: Zeichenkette | 10 | 12,9 | 13,2 | 13,2 | 13,2 | 0 |
| heute · warm · Skript: Kartenfeld | 10 | 12,5 | 13,2 | 13,2 | 13,2 | 0 |
| v050 · kalt · Skript: Rechnung | 2 | 354,7 | 370,2 | 370,2 | 370,2 | 0 |
| v050 · warm · Skript: Rechnung | 8 | 333,5 | 375,4 | 375,4 | 375,4 | 0 |
| v050 · warm · Skript: Zeichenkette | 10 | 361,9 | 368,8 | 368,8 | 368,8 | 0 |
| v050 · warm · Skript: Kartenfeld | 10 | 355,2 | 374,1 | 374,1 | 374,1 | 0 |
| mse2 · bereit · Sitzung | 8 | 518,8 | 5.496,6 | 5.496,6 | 5.496,6 | 0 |
| mse2 · kalt · Normal | 2 | 62,5 | 69,8 | 69,8 | 69,8 | 0 |
| mse2 · warm · Normal | 22 | 42,9 | 55,3 | 58,1 | 58,1 | 0 |
| mse2 · warm · Problemfall | 16 | 48,4 | 108,1 | 108,1 | 108,1 | 2 |
| mse2 · wiederholung · Normal | 2 | 38,5 | 38,7 | 38,7 | 38,7 | 0 |
| mse2 · kalt · Sonderformat | 6 | 54,2 | 94,5 | 94,5 | 94,5 | 0 |
| mse2 · warm · Sonderformat | 30 | 35,6 | 70,6 | 72 | 72 | 0 |
| mse2 · wiederholung · Sonderformat | 6 | 38,2 | 70,6 | 70,6 | 70,6 | 0 |
| heute · bereit · Sitzung | 8 | 29,4 | 254,8 | 254,8 | 254,8 | 0 |
| heute · kalt · Normal | 2 | 850,2 | 960,2 | 960,2 | 960,2 | 0 |
| heute · warm · Normal | 22 | 101,5 | 136,5 | 148,9 | 148,9 | 0 |
| heute · warm · Problemfall | 16 | 158,3 | 394 | 394 | 394 | 0 |
| heute · wiederholung · Normal | 2 | 165 | 174,4 | 174,4 | 174,4 | 0 |
| heute · kalt · Sonderformat | 6 | 739 | 956,8 | 956,8 | 956,8 | 0 |
| heute · warm · Sonderformat | 30 | 76,6 | 184,4 | 191 | 191 | 0 |
| heute · wiederholung · Sonderformat | 6 | 57,1 | 125,4 | 125,4 | 125,4 | 0 |
| v050 · bereit · Sitzung | 8 | 27,4 | 246,2 | 246,2 | 246,2 | 0 |
| v050 · kalt · Normal | 2 | 1.309,9 | 1.517,4 | 1.517,4 | 1.517,4 | 0 |
| v050 · warm · Normal | 22 | 1.317,1 | 1.406,6 | 1.419,4 | 1.419,4 | 0 |
| v050 · warm · Problemfall | 16 | 1.417,5 | 1.546,3 | 1.546,3 | 1.546,3 | 0 |
| v050 · wiederholung · Normal | 2 | 1.293,1 | 1.342,6 | 1.342,6 | 1.342,6 | 0 |
| v050 · kalt · Sonderformat | 6 | 1.296,9 | 1.543,2 | 1.543,2 | 1.543,2 | 0 |
| v050 · warm · Sonderformat | 30 | 1.238,6 | 1.571,1 | 1.572,4 | 1.572,4 | 0 |
| v050 · wiederholung · Sonderformat | 6 | 1.189,5 | 1.469,8 | 1.469,8 | 1.469,8 | 0 |

## Dauerbetrieb

Eine Sitzung je Programm zeichnet reihum Karten; **Versuche** sind alle Rufe, **gelungen** und
**Fehler** ihre Teile. Die Zeit in Millisekunden gilt für die entstandenen Bilder, die der
Ablehnungen steht daneben. *Anfang → Ende*: Median im ersten und im letzten Fünftel der Bilder.

| Programm | Versuche | gelungen | Fehler | Bild Median | Bild p95 | Ablehnung Median | Anfang → Ende | RSS MB Anfang → Höchststand → Ende |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| mse2 | 2.463 | 2.340 | 123 | 45 | 95,5 | 86,5 | 45,1 → 45,4 | 269 → 8.620 → 7.843 |
| heute | 879 | 879 | 0 | 112 | 305,3 | — | 110,9 → 111,9 | 119 → 535 → 535 |

## Bilanz

195 Fälle, 366 Schritte, 2 Befunde, mit Fehler: 2 Schritte
- magic-text.mse-export-template: v050 schreibt eine andere Ausgabe als MSE2
- magic-mws.mse-export-template: v050 schreibt eine andere Ausgabe als MSE2
