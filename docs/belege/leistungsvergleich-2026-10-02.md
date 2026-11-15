# Leistungsvergleich — Bericht

**Stand `f749bfa4`**, Lauf ab 2026-10-02T00:57:26.394Z. Erzeugt von
`scripts/leistungsvergleich.mjs`; die Zahlen stehen in
`leistungsvergleich.json`, die Bewertung in `docs/LEISTUNGSVERGLEICH.md`.

## Ganze Prozesse

Wandzeit in Millisekunden, CPU (`user` + `sys`) in Sekunden, residenter
Höchststand in MB (10⁶ Byte). *vorn / dahinter*: Median der Läufe, in denen
das Programm zuerst lief, gegen den der übrigen.

| Aufgabe | Programm | n | Median | p95 | p99 | erste | CPU | RSS | Fehler | Abstürze | vorn / dahinter |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| start | mse2 | 6 | 60,4 | 67,9 | 67,9 | 60 | 0,02 | 41,1 | 0 | 0 | 60 / 60,4 |
| start | heute | 6 | 23,1 | 23,5 | 23,5 | 23,5 | 0,01 | 40,6 | 0 | 0 | 23,1 / 23 |
| start | v050 | 6 | 23,8 | 26,8 | 26,8 | 23,8 | 0,01 | 36,3 | 0 | 0 | 21,8 / 23,8 |
| oeffnen · Archenemy | mse2 | 6 | 489,9 | 497,8 | 497,8 | 497,7 | 0,43 | 66,7 | 0 | 0 | 489,9 / 484 |
| oeffnen · Archenemy | heute | 6 | 129,3 | 130,6 | 130,6 | 128,5 | 0,12 | 45,4 | 0 | 0 | 129,3 / 129,3 |
| oeffnen · Archenemy | v050 | 6 | 107,9 | 111,4 | 111,4 | 107,8 | 0,1 | 39,3 | 0 | 0 | 107,9 / 107,8 |
| wiederoeffnen · Archenemy | heute | 6 | 24,3 | 41,9 | 41,9 | 24,3 | 0,01 | 41,4 | 0 | 0 | 24,3 / 24,2 |
| wiederoeffnen · Archenemy | v050 | 6 | 23,5 | 25,7 | 25,7 | 25,7 | 0,01 | 37,1 | 0 | 0 | 23,1 / 23,9 |
| sichern · Archenemy | mse2 | 6 | 659,6 | 686,9 | 686,9 | 659,6 | 0,62 | 66,8 | 0 | 0 | 659,6 / 658,2 |
| sichern · Archenemy | heute | 6 | 37,3 | 86,2 | 86,2 | 86,2 | 0,03 | 47,2 | 0 | 0 | 37,2 / 37,3 |
| sichern · Archenemy | v050 | 6 | 34,8 | 82,7 | 82,7 | 82,7 | 0,03 | 41,6 | 0 | 0 | 34,4 / 35,6 |
| ausfuhr magic-simple-exporter · Archenemy | mse2 | 6 | 483,2 | 515,6 | 515,6 | 490,4 | 0,44 | 66,9 | 0 | 0 | 488 / 482,7 |
| ausfuhr magic-simple-exporter · Archenemy | heute | 6 | 303,2 | 308,5 | 308,5 | 308,5 | 0,45 | 108,0 | 0 | 0 | 301,6 / 305,2 |
| ausfuhr magic-simple-exporter · Archenemy | v050 | 6 | 215,7 | 248,7 | 248,7 | 248,7 | 0,34 | 99,6 | 0 | 0 | 213,5 / 219,4 |
| ausfuhr magic-namecheck · Archenemy | mse2 | 6 | 880,5 | 887,2 | 887,2 | 880,5 | 0,84 | 137,6 | 0 | 0 | 879,5 / 880,7 |
| ausfuhr magic-namecheck · Archenemy | heute | 6 | 6.717,4 | 6.747,6 | 6.747,6 | 6.673,7 | 12,38 | 278,5 | 0 | 0 | 6.695,7 / 6.717,4 |
| ausfuhr magic-namecheck · Archenemy | v050 | 6 | 6.674,4 | 6.686,4 | 6.686,4 | 6.620 | 12,17 | 263,3 | 0 | 0 | 6.681,3 / 6.627,8 |
| skript-sitzung · Archenemy | mse2 | 2 | 514 | 527,5 | 527,5 | 527,5 | 0,43 | 66,8 | 0 | 0 | 527,5 / 514 |
| skript-sitzung · Archenemy | heute | 2 | 7.234,9 | 7.913,2 | 7.913,2 | 7.913,2 | 10,86 | 522,2 | 0 | 0 | 7.234,9 / 7.913,2 |
| skript-sitzung · Archenemy | v050 | 2 | 5.108,3 | 5.240,9 | 5.240,9 | 5.240,9 | 7,99 | 512,0 | 0 | 0 | — / 5.108,3 |
| navigation-sitzung · Multicolor | mse2 | 2 | 6.535,5 | 6.558,7 | 6.558,7 | 6.535,5 | 6,42 | 408,0 | 0 | 0 | 6.535,5 / 6.558,7 |
| navigation-sitzung · Multicolor | heute | 2 | 39.441,2 | 39.451 | 39.451 | 39.451 | 48,37 | 550,5 | 0 | 0 | 39.441,2 / 39.451 |
| navigation-sitzung · Multicolor | v050 | 2 | 28.251,5 | 28.438 | 28.438 | 28.438 | 35,33 | 542,2 | 0 | 0 | — / 28.251,5 |
| navigation-sitzung · Archenemy | mse2 | 2 | 756 | 761,3 | 761,3 | 761,3 | 0,7 | 100,5 | 0 | 0 | 761,3 / 756 |
| navigation-sitzung · Archenemy | heute | 2 | 9.188,2 | 9.293,1 | 9.293,1 | 9.188,2 | 11,62 | 518,6 | 0 | 0 | 9.293,1 / 9.188,2 |
| navigation-sitzung · Archenemy | v050 | 2 | 6.594,9 | 6.659,9 | 6.659,9 | 6.594,9 | 8,48 | 512,2 | 0 | 0 | — / 6.594,9 |
| navigation-sitzung · Planechase | mse2 | 2 | 1.030,5 | 1.046,1 | 1.046,1 | 1.046,1 | 0,98 | 127,9 | 0 | 0 | 1.046,1 / 1.030,5 |
| navigation-sitzung · Planechase | heute | 2 | 14.182,2 | 14.402,5 | 14.402,5 | 14.402,5 | 17,08 | 521,0 | 0 | 0 | 14.182,2 / 14.402,5 |
| navigation-sitzung · Planechase | v050 | 2 | 10.356,1 | 10.467,2 | 10.467,2 | 10.467,2 | 12,53 | 512,7 | 0 | 0 | — / 10.356,1 |
| navigation-sitzung · Vanguard | mse2 | 2 | 721,1 | 729,2 | 729,2 | 729,2 | 0,67 | 97,2 | 0 | 0 | 729,2 / 721,1 |
| navigation-sitzung · Vanguard | heute | 2 | 12.482,9 | 12.829,3 | 12.829,3 | 12.482,9 | 15,44 | 522,8 | 0 | 0 | 12.829,3 / 12.482,9 |
| navigation-sitzung · Vanguard | v050 | 2 | 8.623,8 | 8.975,7 | 8.975,7 | 8.623,8 | 10,78 | 510,0 | 0 | 0 | — / 8.623,8 |
| bilder · Archenemy | mse2 | 1 | 3.088,8 | 3.088,8 | 3.088,8 | 3.088,8 | 3,01 | 239,1 | 0 | 0 | 3.088,8 / — |
| bilder · Archenemy | heute | 1 | 94.446,1 | 94.446,1 | 94.446,1 | 94.446,1 | 120,14 | 622,7 | 0 | 0 | — / 94.446,1 |
| bilder · Archenemy | v050 | 1 | 67.047,9 | 67.047,9 | 67.047,9 | 67.047,9 | 86,95 | 585,2 | 0 | 0 | — / 67.047,9 |
| oeffnen · Black | mse2 | 3 | 5.627,5 | 5.655,4 | 5.655,4 | 5.627,5 | 5,58 | 305,5 | 0 | 0 | 5.627,5 / 5.558,7 |
| oeffnen · Black | heute | 3 | 1.943,2 | 1.959,7 | 1.959,7 | 1.959,7 | 2,01 | 107,5 | 0 | 0 | 1.939,5 / 1.943,2 |
| oeffnen · Black | v050 | 3 | 1.932,9 | 1.940,6 | 1.940,6 | 1.901,3 | 1,99 | 99,3 | 0 | 0 | 1.932,9 / 1.901,3 |
| wiederoeffnen · Black | heute | 3 | 72,2 | 72,7 | 72,7 | 72,7 | 0,09 | 84,9 | 0 | 0 | 72,2 / 70,6 |
| wiederoeffnen · Black | v050 | 3 | 72,4 | 73,4 | 73,4 | 71,2 | 0,08 | 80,2 | 0 | 0 | 72,4 / 71,2 |
| sichern · Black | mse2 | 3 | 11.885,2 | 11.896 | 11.896 | 11.885,2 | 11,82 | 306,6 | 0 | 0 | 11.885,2 / 11.863,5 |
| sichern · Black | heute | 3 | 691,8 | 2.151,9 | 2.151,9 | 2.151,9 | 0,96 | 140,9 | 0 | 0 | 686,3 / 691,8 |
| sichern · Black | v050 | 3 | 630,5 | 2.135,3 | 2.135,3 | 2.135,3 | 0,76 | 133,3 | 0 | 0 | 630,5 / 598,5 |

## Schritte in einer laufenden Sitzung

Je Schritt von der geschriebenen Zeile bis zur Antwort, in Millisekunden.

| Programm · Zustand · Schicht | n | Median | p95 | p99 | max | Fehler |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| mse2 · kalt · Skript: Rechnung | 2 | 0,1 | 0,1 | 0,1 | 0,1 | 0 |
| mse2 · warm · Skript: Rechnung | 8 | 0 | 0 | 0 | 0 | 0 |
| mse2 · warm · Skript: Zeichenkette | 10 | 0 | 0,1 | 0,1 | 0,1 | 0 |
| mse2 · warm · Skript: Kartenfeld | 10 | 0 | 0 | 0 | 0 | 0 |
| heute · kalt · Skript: Rechnung | 2 | 475,3 | 487,5 | 487,5 | 487,5 | 0 |
| heute · warm · Skript: Rechnung | 8 | 481,9 | 513,1 | 513,1 | 513,1 | 0 |
| heute · warm · Skript: Zeichenkette | 10 | 479,1 | 511,4 | 511,4 | 511,4 | 0 |
| heute · warm · Skript: Kartenfeld | 10 | 457,9 | 501,6 | 501,6 | 501,6 | 0 |
| v050 · kalt · Skript: Rechnung | 2 | 352,6 | 416,7 | 416,7 | 416,7 | 0 |
| v050 · warm · Skript: Rechnung | 8 | 329,1 | 337,6 | 337,6 | 337,6 | 0 |
| v050 · warm · Skript: Zeichenkette | 10 | 331,8 | 360,2 | 360,2 | 360,2 | 0 |
| v050 · warm · Skript: Kartenfeld | 10 | 335,6 | 361,4 | 361,4 | 361,4 | 0 |
| mse2 · bereit · Sitzung | 8 | 513,9 | 5.453,4 | 5.453,4 | 5.453,4 | 0 |
| mse2 · kalt · Normal | 2 | 65,8 | 66,5 | 66,5 | 66,5 | 0 |
| mse2 · warm · Normal | 22 | 43 | 52,9 | 54,1 | 54,1 | 0 |
| mse2 · warm · Problemfall | 16 | 51 | 104,5 | 104,5 | 104,5 | 2 |
| mse2 · wiederholung · Normal | 2 | 40,5 | 41,5 | 41,5 | 41,5 | 0 |
| mse2 · kalt · Sonderformat | 6 | 57,9 | 98,1 | 98,1 | 98,1 | 0 |
| mse2 · warm · Sonderformat | 30 | 36,4 | 70,5 | 70,9 | 70,9 | 0 |
| mse2 · wiederholung · Sonderformat | 6 | 38,4 | 66,2 | 66,2 | 66,2 | 0 |
| heute · bereit · Sitzung | 8 | 28,6 | 257 | 257 | 257 | 0 |
| heute · kalt · Normal | 2 | 1.870,9 | 1.954,9 | 1.954,9 | 1.954,9 | 0 |
| heute · warm · Normal | 22 | 1.797,2 | 1.912,2 | 1.912,7 | 1.912,7 | 0 |
| heute · warm · Problemfall | 16 | 1.902,3 | 2.219,3 | 2.219,3 | 2.219,3 | 0 |
| heute · wiederholung · Normal | 2 | 1.896 | 1.928,7 | 1.928,7 | 1.928,7 | 0 |
| heute · kalt · Sonderformat | 6 | 1.808,3 | 2.132,3 | 2.132,3 | 2.132,3 | 0 |
| heute · warm · Sonderformat | 30 | 1.788,1 | 2.105,3 | 2.123,6 | 2.123,6 | 0 |
| heute · wiederholung · Sonderformat | 6 | 1.693 | 2.021 | 2.021 | 2.021 | 0 |
| v050 · bereit · Sitzung | 8 | 26,1 | 245,5 | 245,5 | 245,5 | 0 |
| v050 · kalt · Normal | 2 | 1.403 | 1.452,7 | 1.452,7 | 1.452,7 | 0 |
| v050 · warm · Normal | 22 | 1.288,4 | 1.342,9 | 1.359 | 1.359 | 0 |
| v050 · warm · Problemfall | 16 | 1.363 | 1.614,4 | 1.614,4 | 1.614,4 | 0 |
| v050 · wiederholung · Normal | 2 | 1.331,6 | 1.387,3 | 1.387,3 | 1.387,3 | 0 |
| v050 · kalt · Sonderformat | 6 | 1.235 | 1.529,6 | 1.529,6 | 1.529,6 | 0 |
| v050 · warm · Sonderformat | 30 | 1.217,3 | 1.534,7 | 1.542,9 | 1.542,9 | 0 |
| v050 · wiederholung · Sonderformat | 6 | 1.212,4 | 1.487,1 | 1.487,1 | 1.487,1 | 0 |
