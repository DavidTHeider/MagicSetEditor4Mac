# Fuzzkampagne

**1 Ziele, 0 mit Fund** — Commit `7af6b5ca`,
2026-10-01 09:24 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **60s**, **2** Arbeiter, `GOMEMLIMIT=1GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `FuzzZugfolge` | `internal/service` | 80 s | 51.004 | 83 | 249 | ohne Fund |
