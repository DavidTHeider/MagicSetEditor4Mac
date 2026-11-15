# Lastlauf

**25 Fälle** — Commit `5a8acaf1`, 2026-10-01 23:56:55 UTC, Apple M5, 10 Kerne, 32 GiB, macOS 27.0, go1.27.1 darwin/arm64, Node v24.15.0.
Gefahren mit `scripts/lastlauf.mjs`: je Fall **8 s**, **8** Klienten, gegen einen **frisch gestarteten** Dienst mit eigenem Bestand, eigenem Zuhause und eigenem Schlüsselbund. Was der Lauf misst und was nicht, steht in seinem Kopf und in `docs/LASTMESSUNG.md`.

**Zeit** ist die des Klienten (p50/p95/p99 in ms über alle Rufe, gelungen oder nicht); **Dienst** ist die Obergrenze aus den Eimern von `/metriken` bzw. `lesen.metriken`. **Gedrosselt** ist die erwartete Schutzreaktion und kein Fehler.

## Kontingente an

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 132106 | 16490.3 | 16.9 | 99.9 % | 0 % | 0.4 | 1 | 1.5 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 53 |
| `rest-schreiben` | REST | 67783 | 8467 | 8.4 | 99.9 % | 0 % | 0.8 | 1.6 | 2.4 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 56 |
| `rest-rendern` | REST | 26957 | 3368.9 | 8.4 | 99.8 % | 0 % | 1.1 | 3.8 | 8.4 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 2812 |
| `rest-einfuhr` | REST | 25303 | 3161.8 | 2 | 99.9 % | 0 % | 2.2 | 4.5 | 9.8 | 1–4 s | ≤ 5 ms / ≤ 5 ms | 55 |
| `rest-auftraege` | REST-Aufträge | 37 | 4.4 | 1.1 | 75.7 % | 0 % | 1.1 | 3350.9 | 7462.5 | 1–5 s | ≤ 5 ms / ≤ 25 ms | 484 |
| `mcp-stdio` | MCP-stdio | 318154 | 35187.7 | 59.7 | 99.8 % | 0 % | 0.1 | 0.2 | 0.2 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 42 |
| `mcp-http` | MCP-HTTP | 58483 | 7303.6 | 16.9 | 99.8 % | 0 % | 1 | 1.9 | 2.6 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 56 |
| `gemischt` | REST lesen | 68211 | 8492.7 | 11.6 | 99.9 % | 0 % | 0.3 | 0.7 | 1 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 426 |
| `gemischt` | REST schreiben | 13797 | 1717.8 | 2.5 | 99.9 % | 0 % | 1.1 | 1.8 | 2.5 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 426 |
| `gemischt` | REST rendern | 6821 | 849.3 | 0.1 | 100 % | 0 % | 1.1 | 1.8 | 2.5 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 426 |
| `gemischt` | MCP-HTTP | 29173 | 3632.2 | 16.8 | 99.5 % | 0 % | 0.5 | 0.8 | 1.1 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 426 |
| `gemischt` | MCP-stdio | 42533 | 5295.6 | 16.8 | 99.7 % | 0 % | 0.2 | 0.3 | 0.4 | 1–1 s | — | 426 |
| `fairness` | A flutet | 106770 | 13296.4 | 16.8 | 99.9 % | 0 % | 0.4 | 0.8 | 1.2 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 54 |
| `fairness` | B fragt | 32 | 4 | 4 | 0 % | 0 % | 1 | 1.7 | 2.3 | — | ≤ 1 ms / ≤ 1 ms | 54 |
| `ueberlast` | vorher (höflich) | 20 | 10 | 10 | 0 % | 0 % | 0.8 | 2.8 | 4.4 | — | ≤ 1 ms / ≤ 1 ms | 54 |
| `ueberlast` | Überlast | 130350 | 16293.8 | 15 | 76 % | 23.9 % | 1.7 | 3.8 | 7.3 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 54 |
| `ueberlast` | nachher (höflich) | 7 | 2.3 | 1.3 | 42.9 % | 0 % | 0.6 | 0.7 | 0.7 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 54 |
| `langsam` | REST hinter 20 halben Köpfen | 11 | 0.9 | 0.9 | 0 % | 0 % | 0.9 | 9801.9 | 9801.9 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 59 | 4.9 | 4.9 | 0 % | 0 % | 1.1 | 3.1 | 17.6 | — | ≤ 1 ms / ≤ 5 ms | 47 |
| `ruhend` | REST neben 16 ruhenden | 59 | 4.9 | 4.9 | 0 % | 0 % | 1 | 4 | 22.5 | — | ≤ 5 ms / ≤ 5 ms | 46 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 0.8 | 1.1 | 1.1 | — | ≤ 5 ms / ≤ 5 ms | 46 |

- **`rest-auftraege`**: der Auftrag zum Abbrechen kam nicht an (429)
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 1003 ms nach dem Ende der Überlast

## Kontingente aus

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 19127 | 2389.8 | 2389.8 | 0 % | 0 % | 2.5 | 6.4 | 19.6 | — | ≤ 5 ms / ≤ 25 ms | 53 |
| `rest-schreiben` | REST | 2620 | 326.9 | 326.9 | 0 % | 0 % | 22.8 | 38.6 | 53.3 | — | ≤ 100 ms / ≤ 100 ms | 57 |
| `rest-rendern` | REST | 81 | 9.5 | 9.5 | 0 % | 0 % | 800.4 | 1059.8 | 1268.9 | — | ≤ 2500 ms / ≤ 2500 ms | 2578 |
| `rest-einfuhr` | REST | 8556 | 1068.8 | 1068.8 | 0 % | 0 % | 6.7 | 13.1 | 22.5 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `rest-auftraege` | REST-Aufträge | 27 | 2.7 | 1.9 | 29.6 % | 0 % | 2031.7 | 2190.2 | 2218.6 | 5–5 s | ≤ 5 ms / ≤ 5 ms | 490 |
| `mcp-stdio` | MCP-stdio | 68064 | 8434.8 | 8434.8 | 0 % | 0 % | 0.3 | 1 | 2.3 | — | ≤ 1 ms / ≤ 5 ms | 42 |
| `mcp-http` | MCP-HTTP | 13454 | 1681.3 | 1681.3 | 0 % | 0 % | 4.1 | 8 | 21.7 | — | ≤ 5 ms / ≤ 25 ms | 55 |
| `gemischt` | REST lesen | 2732 | 338.6 | 338.6 | 0 % | 0 % | 5.5 | 23.2 | 83 | — | ≤ 25 ms / ≤ 100 ms | 612 |
| `gemischt` | REST schreiben | 935 | 115.9 | 115.9 | 0 % | 0 % | 11.6 | 51.9 | 109.5 | — | ≤ 25 ms / ≤ 100 ms | 612 |
| `gemischt` | REST rendern | 29 | 3.6 | 3.6 | 0 % | 0 % | 252.8 | 405.9 | 474 | — | ≤ 25 ms / ≤ 100 ms | 612 |
| `gemischt` | MCP-HTTP | 7800 | 966.7 | 966.7 | 0 % | 0 % | 1.2 | 4 | 17.4 | — | ≤ 5 ms / ≤ 5 ms | 612 |
| `gemischt` | MCP-stdio | 8920 | 1105.5 | 1105.5 | 0 % | 0 % | 0.4 | 2 | 6.8 | — | — | 612 |
| `fairness` | A flutet | 18570 | 2286.8 | 2286.8 | 0 % | 0 % | 2.1 | 4.7 | 11.7 | — | ≤ 5 ms / ≤ 5 ms | 54 |
| `fairness` | B fragt | 32 | 3.9 | 3.9 | 0 % | 0 % | 2.7 | 4 | 5.3 | — | ≤ 5 ms / ≤ 5 ms | 54 |
| `ueberlast` | vorher (höflich) | 20 | 10 | 10 | 0 % | 0 % | 0.9 | 1.5 | 1.5 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `ueberlast` | Überlast | 26059 | 3257.4 | 2002.9 | 0 % | 38.5 % | 7.1 | 32.7 | 63.7 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `ueberlast` | nachher (höflich) | 29 | 9.7 | 9.7 | 0 % | 0 % | 0.9 | 10.2 | 12.2 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `langsam` | REST hinter 20 halben Köpfen | 11 | 0.9 | 0.9 | 0 % | 0 % | 0.7 | 9800.2 | 9800.2 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 60 | 5 | 5 | 0 % | 0 % | 1.1 | 6 | 10.5 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `ruhend` | REST neben 16 ruhenden | 60 | 5 | 5 | 0 % | 0 % | 1 | 1.3 | 1.7 | — | ≤ 5 ms / ≤ 5 ms | 46 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 0.8 | 1 | 1 | — | ≤ 5 ms / ≤ 5 ms | 46 |
| `rest-lesen (Protokoll ausführlich)` | REST | 17506 | 2187.9 | 2187.9 | 0 % | 0 % | 2.6 | 7 | 30.3 | — | ≤ 5 ms / ≤ 25 ms | 54 |

- **`rest-auftraege`**: ein Auftrag, gleich nach dem Annehmen abgebrochen: `DELETE` → 202 in 4 ms, Stand danach `laeuft`
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 7 ms nach dem Ende der Überlast
