# Fuzzkampagne

**3 Ziele, 0 mit Fund** — Commit `7af56d7e`,
2026-10-01 07:43 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **30s**, **2** Arbeiter, `GOMEMLIMIT=1GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `FuzzPaketpfad` | `internal/msepaket` | 37 s | 191.488 | 140 | 167 | ohne Fund |
| `FuzzDekodiereBild` | `internal/msepaket` | 32 s | 369.891 | 67 | 114 | ohne Fund |
| `FuzzZugfolge` | `internal/service` | 39 s | 59.071 | 109 | 151 | ohne Fund |
