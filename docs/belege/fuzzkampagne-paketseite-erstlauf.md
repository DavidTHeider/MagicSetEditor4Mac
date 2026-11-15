# Fuzzkampagne — Paketseite, erster Lauf

**7 Ziele, 2 mit Fund** — Commit `ac80aea9` **mit den Zielen aus QA-06b im
Arbeitsbaum**, noch nicht eingetragen (#1967),
2026-10-01 06:58 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **30s**, **2** Arbeiter, `GOMEMLIMIT=1GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `FuzzInstaller` | `internal/msepaket` | 33 s | 20.868 | 20 | 28 | ohne Fund |
| `FuzzPaketpfad` | `internal/msepaket` | 0 s | 0 | 0 | 0 | **Fund**: fuzz_paketpfad_test.go:66: "\xff" in "0" gibt die Datei "\xff" — kein Weg im Paket — `testdata/fuzz/FuzzPaketpfad/ce4c7b7271bf0acb` |
| `FuzzEinbinden` | `internal/msepaket` | 31 s | 1.418.029 | 54 | 330 | ohne Fund |
| `FuzzSpielmodell` | `internal/msepaket` | 30 s | 1.656.060 | 146 | 150 | ohne Fund |
| `FuzzStilmodell` | `internal/msepaket` | 30 s | 1.214.668 | 157 | 160 | ohne Fund |
| `FuzzDekodiereBild` | `internal/msepaket` | 4 s | 36.001 | 8 | 21 | **Fund**: fuzz_bild_test.go:102: , 0 × 0: 10499552 Byte zugeteilt, zugesagt 1048608 — `testdata/fuzz/FuzzDekodiereBild/20794022d0a16eb2` |
| `FuzzSymbolzeichnung` | `internal/msepaket` | 30 s | 1.560.399 | 168 | 173 | ohne Fund |
