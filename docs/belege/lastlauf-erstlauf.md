# Lastlauf

**25 Fälle** — Commit `37d58a05` **mit Änderungen im Baum**, 2026-10-01 23:43:46 UTC, Apple M5, 10 Kerne, 32 GiB, macOS 27.0, go1.27.1 darwin/arm64, Node v24.15.0.
Gefahren mit `scripts/lastlauf.mjs`: je Fall **8 s**, **8** Klienten, gegen einen **frisch gestarteten** Dienst mit eigenem Bestand, eigenem Zuhause und eigenem Schlüsselbund. Was der Lauf misst und was nicht, steht in seinem Kopf und in `docs/LASTMESSUNG.md`.

**Zeit** ist die des Klienten (p50/p95/p99 in ms über alle Rufe, gelungen oder nicht); **Dienst** ist die Obergrenze aus den Eimern von `/metriken` bzw. `lesen.metriken`. **Gedrosselt** ist die erwartete Schutzreaktion und kein Fehler.

## Kontingente an

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 181973 | 22722 | 16.9 | 99.9 % | 0 % | 0.3 | 0.8 | 1.2 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 54 |
| `rest-schreiben` | REST | 35248 | 4404.3 | 8.4 | 99.8 % | 0 % | 1.5 | 3.9 | 8.4 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 55 |
| `rest-rendern` | REST | 11989 | 1498.1 | 8.4 | 99.4 % | 0 % | 2.4 | 6.3 | 18.6 | 1–1 s | ≤ 5 ms / ≤ 25 ms | 2840 |
| `rest-einfuhr` | REST | 22822 | 2851.2 | 2 | 99.9 % | 0 % | 2.3 | 5.4 | 13.8 | 1–4 s | ≤ 5 ms / ≤ 25 ms | 55 |
| `rest-auftraege` | REST-Aufträge | 16807 | 1516.8 | 0.9 | 99.9 % | 0 % | 0.5 | 1.6 | 4.1 | 1–5 s | ≤ 1 ms / ≤ 5 ms | 504 |
| `mcp-stdio` | MCP-stdio | 215223 | 23744.1 | 59.6 | 99.8 % | 0 % | 0.1 | 0.2 | 0.3 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 41 |
| `mcp-http` | MCP-HTTP | 30465 | 3807.2 | 16.9 | 99.6 % | 0 % | 1.9 | 3.7 | 7.3 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 55 |
| `gemischt` | REST lesen | 23314 | 2897.6 | 11.4 | 99.6 % | 0 % | 0.5 | 2 | 9 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 426 |
| `gemischt` | REST schreiben | 5513 | 685.2 | 2.6 | 99.6 % | 0 % | 1.7 | 5.4 | 38.6 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 426 |
| `gemischt` | REST rendern | 2707 | 336.4 | 0.1 | 100 % | 0 % | 1.7 | 5.1 | 39.5 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 426 |
| `gemischt` | MCP-HTTP | 10459 | 1299.9 | 16.8 | 98.7 % | 0 % | 0.8 | 2.7 | 17.5 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 426 |
| `gemischt` | MCP-stdio | 13651 | 1696.6 | 16.8 | 99 % | 0 % | 0.3 | 1.2 | 4.3 | 1–1 s | — | 426 |
| `fairness` | A flutet | 11909 | 1469.4 | 16.7 | 98.9 % | 0 % | 1.1 | 9 | 102.2 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 52 |
| `fairness` | B fragt | 26 | 3.2 | 3.2 | 0 % | 0 % | 4.4 | 28.4 | 32.7 | — | ≤ 5 ms / ≤ 5 ms | 52 |
| `ueberlast` | vorher | 2405 | 1202.5 | 61.5 | 94.9 % | 0 % | 0.2 | 0.7 | 12.8 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 53 |
| `ueberlast` | Überlast | 21129 | 2641.1 | 2.1 | 99.9 % | 0 % | 2.3 | 46.7 | 293.1 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 53 |
| `ueberlast` | nachher (höflich) | 7 | 2.3 | 1.3 | 42.9 % | 0 % | 0.6 | 1.1 | 1.1 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 53 |
| `langsam` | REST hinter 20 halben Köpfen | 11 | 0.9 | 0.9 | 0 % | 0 % | 0.8 | 9799.7 | 9799.7 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 60 | 5 | 5 | 0 % | 0 % | 1.1 | 3.6 | 5.1 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `ruhend` | REST neben 16 ruhenden | 1 | 0.1 | 0 | 0 % | 100 % | 15002.4 | 15002.4 | 15002.4 | — | ≤ 5 ms / ≤ 5 ms | 44 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 0.8 | 2.3 | 2.3 | — | ≤ 5 ms / ≤ 5 ms | 44 |

- **`rest-auftraege`**: der Auftrag zum Abbrechen kam nicht an (429)
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 1004 ms nach dem Ende der Überlast

## Kontingente aus

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 37564 | 4694.4 | 4694.4 | 0 % | 0 % | 1.4 | 3.5 | 7.7 | — | ≤ 5 ms / ≤ 5 ms | 54 |
| `rest-schreiben` | REST | 3322 | 413.9 | 413.9 | 0 % | 0 % | 18.3 | 33.5 | 45.1 | — | ≤ 100 ms / ≤ 100 ms | 56 |
| `rest-rendern` | REST | 116 | 14 | 14 | 0 % | 0 % | 539.6 | 776.2 | 835 | — | ≤ 2500 ms / ≤ 2500 ms | 2499 |
| `rest-einfuhr` | REST | 15570 | 1945.7 | 1945.7 | 0 % | 0 % | 3.4 | 8.8 | 13.1 | — | ≤ 25 ms / ≤ 25 ms | 56 |
| `rest-auftraege` | REST-Aufträge | 17934 | 1978.6 | 2.8 | 99.9 % | 0 % | 1.5 | 2.8 | 8.3 | 5–5 s | ≤ 5 ms / ≤ 5 ms | 675 |
| `mcp-stdio` | MCP-stdio | 127820 | 15892.2 | 15892.2 | 0 % | 0 % | 0.2 | 0.3 | 0.6 | — | ≤ 1 ms / ≤ 1 ms | 41 |
| `mcp-http` | MCP-HTTP | 19992 | 2498 | 2498 | 0 % | 0 % | 2.8 | 5.8 | 12.7 | — | ≤ 5 ms / ≤ 5 ms | 55 |
| `gemischt` | REST lesen | 4039 | 488.1 | 488.1 | 0 % | 0 % | 4.8 | 13.9 | 25.7 | — | ≤ 25 ms / ≤ 25 ms | 611 |
| `gemischt` | REST schreiben | 1303 | 157.5 | 157.5 | 0 % | 0 % | 10.7 | 23.2 | 38.3 | — | ≤ 25 ms / ≤ 25 ms | 611 |
| `gemischt` | REST rendern | 36 | 4.4 | 4.4 | 0 % | 0 % | 224.6 | 260.3 | 276 | — | ≤ 25 ms / ≤ 25 ms | 611 |
| `gemischt` | MCP-HTTP | 10225 | 1235.7 | 1235.7 | 0 % | 0 % | 1.1 | 3.1 | 7.2 | — | ≤ 1 ms / ≤ 5 ms | 611 |
| `gemischt` | MCP-stdio | 11199 | 1353.4 | 1353.4 | 0 % | 0 % | 0.4 | 1.9 | 4.2 | — | — | 611 |
| `fairness` | A flutet | 17526 | 2155.4 | 2155.4 | 0 % | 0 % | 2.2 | 5.2 | 15.2 | — | ≤ 5 ms / ≤ 25 ms | 54 |
| `fairness` | B fragt | 32 | 3.9 | 3.9 | 0 % | 0 % | 2.5 | 7.1 | 10.2 | — | ≤ 5 ms / ≤ 25 ms | 54 |
| `ueberlast` | vorher | 2618 | 1309 | 1309 | 0 % | 0 % | 0.6 | 1.5 | 3.3 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `ueberlast` | Überlast | 20385 | 2548.1 | 2548.1 | 0 % | 0 % | 10.8 | 22.7 | 51.3 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `ueberlast` | nachher (höflich) | 6548 | 2182.7 | 2182.7 | 0 % | 0 % | 0.3 | 0.8 | 2.1 | — | ≤ 25 ms / ≤ 25 ms | 55 |
| `langsam` | REST hinter 20 halben Köpfen | 11 | 0.9 | 0.9 | 0 % | 0 % | 0.6 | 9801.2 | 9801.2 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 60 | 5 | 5 | 0 % | 0 % | 0.7 | 1 | 1.7 | — | ≤ 1 ms / ≤ 1 ms | 47 |
| `ruhend` | REST neben 16 ruhenden | 1 | 0.1 | 0 | 0 % | 100 % | 15000.7 | 15000.7 | 15000.7 | — | ≤ 5 ms / ≤ 5 ms | 44 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 0.5 | 0.8 | 0.8 | — | ≤ 5 ms / ≤ 5 ms | 44 |
| `rest-lesen (Protokoll ausführlich)` | REST | 52163 | 6504.2 | 6504.2 | 0 % | 0 % | 1 | 2.8 | 5.6 | — | ≤ 5 ms / ≤ 5 ms | 54 |

- **`rest-auftraege`**: ein Auftrag, gleich nach dem Annehmen abgebrochen: `DELETE` → 202 in 1 ms, Stand danach `abgebrochen`
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 1 ms nach dem Ende der Überlast
