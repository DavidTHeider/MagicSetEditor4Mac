# Lastlauf

**25 Fälle** — Commit `f381125b`, 2026-10-20 15:59:42 UTC, Apple M5, 10 Kerne, 32 GiB, macOS 27.0.1, go1.27.2 darwin/arm64, Node v24.15.0. Last der Maschine (1/5/15 min) vorher 4.1 / 3.4 / 3.2, nachher 4.1 / 3.9 / 3.5.
Gefahren mit `scripts/lastlauf.mjs`: je Fall **8 s**, **8** Klienten, gegen einen **frisch gestarteten** Dienst mit eigenem Bestand, eigenem Zuhause und eigenem Schlüsselbund. Was der Lauf misst und was nicht, steht in seinem Kopf und in `docs/LASTMESSUNG.md`.

**Zeit** ist die des Klienten (p50/p95/p99 in ms über alle Rufe, gelungen oder nicht); **Dienst** ist die Obergrenze aus den Eimern von `/metriken` bzw. `lesen.metriken`. **Gedrosselt** ist die erwartete Schutzreaktion und kein Fehler. **Gelungen** heißt Status 2xx **und**, wo ein Profil eine Wirkung zusagt, dass eine Stichprobe sie belegt (Bild am Ziel, Feld im Bestand, Set nach der Einfuhr — `scripts/lastlaufwirkung.mjs`); ein Ruf ohne Wirkung zählt als Fehler.

**Vorher die Rohpfade**, je einmal ohne Last und mit belegter Wirkung: Lesen, Rendern, Einfuhr, Auftrag. Eine Gruppe, die Wirkung zusagt und keinen Ruf durchbringt, steht als **ungeprüft** da; mit Kontingenten ist eine nur gedrosselte Gruppe kein Befund, ohne sie macht sie den Lauf rot.

## Kontingente an

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 80078 | 10003.6 | 16.9 | 99.8 % | 0 % | 0.7 | 1.5 | 1.9 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 58 |
| `rest-schreiben` | REST | 50345 | 6288.8 | 0.4 | 99.9 % | 0 % | 1.2 | 1.9 | 2.8 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 69 |
| `rest-rendern` | REST | 50764 | 6341.3 | 3.4 | 100 % | 0 % | 1.2 | 1.9 | 2.5 | 1–3 s | ≤ 5 ms / ≤ 5 ms | 75 |
| `rest-einfuhr` | REST | 50821 | 6346.8 | 2 | 100 % | 0 % | 1.2 | 1.9 | 2.5 | 1–4 s | ≤ 5 ms / ≤ 5 ms | 58 |
| `rest-auftraege` | REST-Aufträge | 30 | 2.7 | 0.5 | 80 % | 0 % | 2.2 | 1635.8 | 2691.5 | 1–5 s | ≤ 5 ms / ≤ 5 ms | 90 |
| `mcp-stdio` | MCP-stdio | 124316 | 13659.8 | 59.3 | 99.6 % | 0 % | 0.2 | 0.4 | 0.6 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 44 |
| `mcp-http` | MCP-HTTP | 44391 | 5544.6 | 16.9 | 99.7 % | 0 % | 1.4 | 2.2 | 2.8 | 1–1 s | ≤ 1 ms / ≤ 5 ms | 58 |
| `gemischt` | REST lesen | 28277 | 3523.6 | 15.7 | 99.6 % | 0 % | 0.8 | 1.4 | 2 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 78 |
| `gemischt` | REST schreiben | 9204 | 1146.9 | 0.2 | 100 % | 0 % | 1.6 | 2.6 | 4.4 | 1–1 s | ≤ 5 ms / ≤ 5 ms | 78 |
| `gemischt` | REST rendern | 4544 | 566.2 | 0.1 | 100 % | 0 % | 1.6 | 2.6 | 3.9 | 2–3 s | ≤ 5 ms / ≤ 5 ms | 78 |
| `gemischt` | MCP-HTTP | 23933 | 2982.3 | 16.8 | 99.4 % | 0 % | 0.6 | 1 | 1.4 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 78 |
| `gemischt` | MCP-stdio | 23779 | 2963.1 | 16.8 | 99.4 % | 0 % | 0.3 | 0.5 | 0.7 | 1–1 s | — | 78 |
| `fairness` | A flutet | 64846 | 8063.3 | 16.8 | 99.8 % | 0 % | 0.7 | 1.2 | 1.5 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 57 |
| `fairness` | B fragt | 32 | 4 | 4 | 0 % | 0 % | 1.3 | 2 | 3 | — | ≤ 1 ms / ≤ 1 ms | 57 |
| `ueberlast` | vorher (höflich) | 20 | 10 | 10 | 0 % | 0 % | 1.8 | 2.2 | 2.3 | — | ≤ 1 ms / ≤ 1 ms | 58 |
| `ueberlast` | Überlast | 78337 | 9792.1 | 15 | 99.9 % | 0 % | 3.1 | 4.9 | 8.1 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 58 |
| `ueberlast` | nachher (höflich) | 7 | 2.3 | 1.3 | 42.9 % | 0 % | 1.6 | 2 | 2 | 1–1 s | ≤ 1 ms / ≤ 1 ms | 58 |
| `langsam` | REST hinter 20 halben Köpfen | 50 | 4.2 | 4.2 | 0 % | 0 % | 1.6 | 3.2 | 1802.8 | — | ≤ 1 ms / ≤ 1 ms | 51 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 50 | 4.2 | 4.2 | 0 % | 0 % | 1.6 | 2.9 | 1802.7 | — | ≤ 1 ms / ≤ 1 ms | 51 |
| `ruhend` | REST neben 16 ruhenden | 59 | 4.9 | 4.9 | 0 % | 0 % | 1.5 | 2.2 | 12.5 | — | ≤ 1 ms / ≤ 5 ms | 49 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 1.5 | 2.1 | 2.1 | — | ≤ 1 ms / ≤ 5 ms | 49 |

- **`rest-einfuhr`**: 50821 Rufe: 1 legte ein Set an, 15 erkannten die Quelle als schon offen, 0 abgewiesen, 50805 gedrosselt
- **`rest-auftraege`**: ein Auftrag, gleich nach dem Annehmen abgebrochen: `DELETE` → 202, Stand `abgebrochen` nach 22 ms
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 1003 ms nach dem Ende der Überlast

## Kontingente aus

| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |
| `rest-lesen` | REST | 49143 | 6138.8 | 6138.8 | 0 % | 0 % | 1.2 | 2.1 | 2.6 | — | ≤ 5 ms / ≤ 5 ms | 57 |
| `rest-schreiben` | REST | 8399 | 1048.8 | 0.4 | 0 % | 0 % | 7.5 | 9 | 11.5 | — | ≤ 25 ms / ≤ 25 ms | 70 |
| `rest-rendern` | REST | 2718 | 338.8 | 338.8 | 0 % | 0 % | 23.3 | 24.9 | 28.5 | — | ≤ 25 ms / ≤ 100 ms | 77 |
| `rest-einfuhr` | REST | 28980 | 3615.4 | 3615.4 | 0 % | 0 % | 2.1 | 3 | 3.9 | — | ≤ 5 ms / ≤ 5 ms | 59 |
| `rest-auftraege` | REST-Aufträge | 54 | 5.4 | 4.6 | 14.8 % | 0 % | 736.1 | 793.3 | 849.2 | 5–5 s | ≤ 1 ms / ≤ 5 ms | 94 |
| `mcp-stdio` | MCP-stdio | 97886 | 12094.4 | 12094.4 | 0 % | 0 % | 0.3 | 0.5 | 0.7 | — | ≤ 1 ms / ≤ 1 ms | 44 |
| `mcp-http` | MCP-HTTP | 35314 | 4410.7 | 4410.7 | 0 % | 0 % | 1.7 | 2.6 | 3.4 | — | ≤ 5 ms / ≤ 5 ms | 59 |
| `gemischt` | REST lesen | 11056 | 1377.6 | 1377.6 | 0 % | 0 % | 2.1 | 3.7 | 5.7 | — | ≤ 25 ms / ≤ 25 ms | 83 |
| `gemischt` | REST schreiben | 2591 | 322.8 | 0.4 | 0 % | 0 % | 5.8 | 8.1 | 11.4 | — | ≤ 25 ms / ≤ 25 ms | 83 |
| `gemischt` | REST rendern | 502 | 62.5 | 62.5 | 0 % | 0 % | 15.4 | 18.8 | 22.3 | — | ≤ 25 ms / ≤ 25 ms | 83 |
| `gemischt` | MCP-HTTP | 18864 | 2350.4 | 2350.4 | 0 % | 0 % | 0.8 | 1.3 | 1.6 | — | ≤ 1 ms / ≤ 1 ms | 83 |
| `gemischt` | MCP-stdio | 21232 | 2645.5 | 2645.5 | 0 % | 0 % | 0.3 | 0.6 | 0.8 | — | — | 83 |
| `fairness` | A flutet | 44498 | 5535.4 | 5535.4 | 0 % | 0 % | 1 | 1.6 | 2 | — | ≤ 1 ms / ≤ 5 ms | 58 |
| `fairness` | B fragt | 32 | 4 | 4 | 0 % | 0 % | 1.3 | 2.2 | 2.6 | — | ≤ 1 ms / ≤ 5 ms | 58 |
| `ueberlast` | vorher (höflich) | 20 | 10 | 10 | 0 % | 0 % | 1.5 | 2.3 | 13.3 | — | ≤ 5 ms / ≤ 5 ms | 58 |
| `ueberlast` | Überlast | 47864 | 5983 | 5983 | 0 % | 0 % | 5.1 | 7.5 | 10.8 | — | ≤ 5 ms / ≤ 5 ms | 58 |
| `ueberlast` | nachher (höflich) | 30 | 10 | 10 | 0 % | 0 % | 1.5 | 1.9 | 1.9 | — | ≤ 5 ms / ≤ 5 ms | 58 |
| `langsam` | REST hinter 20 halben Köpfen | 51 | 4.3 | 4.3 | 0 % | 0 % | 1.6 | 2.5 | 1803.5 | — | ≤ 1 ms / ≤ 1 ms | 51 |
| `langsam` | MCP-HTTP hinter 20 halben Köpfen | 50 | 4.2 | 4.2 | 0 % | 0 % | 1.6 | 10.6 | 1803.5 | — | ≤ 1 ms / ≤ 1 ms | 51 |
| `ruhend` | REST neben 16 ruhenden | 59 | 4.9 | 4.9 | 0 % | 0 % | 1.8 | 2.4 | 12.6 | — | ≤ 1 ms / ≤ 5 ms | 50 |
| `ruhend` | REST, nachdem sie gingen | 10 | 5 | 5 | 0 % | 0 % | 1.8 | 2 | 2 | — | ≤ 1 ms / ≤ 5 ms | 50 |
| `rest-lesen (Protokoll ausführlich)` | REST | 49550 | 6188.7 | 6188.7 | 0 % | 0 % | 1.2 | 2 | 2.6 | — | ≤ 5 ms / ≤ 5 ms | 57 |

- **`rest-einfuhr`**: 28980 Rufe: 1 legte ein Set an, 28979 erkannten die Quelle als schon offen, 0 abgewiesen, 0 gedrosselt
- **`rest-auftraege`**: ein Auftrag, gleich nach dem Annehmen abgebrochen: `DELETE` → 202, Stand `abgebrochen` nach 24 ms
- **`mcp-stdio`**: 4 Prozesse, je einer ein eigener Principal und ein eigenes Guthaben
- **`ueberlast`**: erster gelungener Ruf 0 ms nach dem Ende der Überlast
