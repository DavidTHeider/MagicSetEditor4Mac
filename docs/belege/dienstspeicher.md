# Dienstspeicher nach Rendern und Ruhe

**bestanden** — Commit `1316ec35a`,
2026-10-15 05:03 UTC, Mac17,3, Apple M5, 10 Kerne, 32 GiB, macOS 27.0.1,
`node scripts/dienstspeicher.mjs` — Last 30 s, Ruhe 120 s, `GOMEMLIMIT=384MiB` im zweiten Fall; Last vor dem Lauf 2.07 / 2.13 / 2.66.

| Reihe | n | min | Median | P95 | max | Grenze | Urteil |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Dienst nach Rendern und Ruhe, mse-rest (MiB) | 1 | 31 | 31 | — | 31 | max ≤ 512 | gehalten |
| Dienst nach Rendern und Ruhe, mse-rest, mit GOMEMLIMIT (MiB) | 1 | 31 | 31 | — | 31 | max ≤ 512 | gehalten |
| Dienst nach Rendern und Ruhe, mse-mcp (MiB) | 1 | 48 | 48 | — | 48 | max ≤ 512 | gehalten |
| Dienst nach Rendern und Ruhe, mse-mcp, mit GOMEMLIMIT (MiB) | 1 | 48 | 48 | — | 48 | max ≤ 512 | gehalten |

| Dienst | GOMEMLIMIT | gelungen | abgewiesen | Fußabdruck am Anfang | höchstens unter Last | nach der Ruhe |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| mse-rest | — | 10746 | 0 | 13 MiB | 44 MiB | 31 MiB |
| mse-rest | 384MiB | 11354 | 0 | 13 MiB | 41 MiB | 31 MiB |
| mse-mcp | — | 182 | 0 | 14 MiB | 53 MiB | 48 MiB |
| mse-mcp | 384MiB | 182 | 0 | 14 MiB | 52 MiB | 48 MiB |

**Was es nicht belegt**: eine zweite Maschine und einen Dienst, der Tage läuft. Gemessen ist der Fußabdruck nach `footprint`, nicht RSS.
