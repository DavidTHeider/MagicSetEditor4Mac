# Fuzzkampagne

**7 Ziele, 1 mit Fund** — Commit `190716a8`,
2026-10-01 18:57 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **30s**, **2** Arbeiter, `GOMEMLIMIT=2GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf. **Anteil** ist der
Anteil der Anweisungen des Pakets, den Saat und erarbeitetes Korpus betreten —
keine Kantenabdeckung, und der Nenner ist das ganze Paket.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Anteil | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `FuzzBildangabe` | `internal/service` | 39 s | 510.696 | 135 | 143 | 0.3 % | ohne Fund |
| `FuzzTextskript` | `internal/msescript` | 1 s | 0 | 0 | 0 | 19.3 % | **Fund**: fuzz_textskript_test.go:75: ein Text ohne Ausdruck kam verändert zurück: "\xfa" → "�" — `testdata/fuzz/FuzzTextskript/3940e35d8f932097` |
| `FuzzKartenskript` | `internal/service` | 31 s | 436.544 | 318 | 328 | 0.5 % | ohne Fund |
| `FuzzEntpackenVerschluesselt` | `internal/sicherung` | 31 s | 24.511 | 18 | 27 | 65.3 % | ohne Fund |
| `FuzzUmschlagdeck` | `internal/app` | 68 s | 14.175 | 40 | 46 | 15.5 % | ohne Fund |
| `FuzzHandschlag` | `internal/peer` | 32 s | 88.677 | 15 | 24 | 74.3 % | ohne Fund |
| `FuzzSpielbefehl` | `internal/service` | 31 s | 397.225 | 37 | 289 | 3.1 % | ohne Fund |
