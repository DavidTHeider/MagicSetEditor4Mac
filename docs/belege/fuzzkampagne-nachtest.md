# Fuzzkampagne

**2 Ziele, 0 mit Fund** — Commit `e7f81f35`,
2026-10-01 04:39 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **1m**, **8** Arbeiter, `GOMEMLIMIT=2GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `FuzzZerlegeKontakteinladung` | `internal/domain` | 63 s | 9.295.648 | 116 | 138 | ohne Fund |
| `FuzzZerlegeAdresse` | `internal/domain` | 61 s | 8.140.346 | 103 | 232 | ohne Fund |
