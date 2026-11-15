# Fuzzkampagne

**1 Ziele, 0 mit Fund** — Commit `190716a8`,
2026-10-01 18:59 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **30s**, **2** Arbeiter, `GOMEMLIMIT=2GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf. **Anteil** ist der
Anteil der Anweisungen des Pakets, den Saat und erarbeitetes Korpus betreten —
keine Kantenabdeckung, und der Nenner ist das ganze Paket.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Anteil | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `FuzzTextskript` | `internal/msescript` | 32 s | 906.967 | 128 | 139 | 21.1 % | ohne Fund |
