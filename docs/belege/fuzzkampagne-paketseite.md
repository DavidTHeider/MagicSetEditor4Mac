# Fuzzkampagne

**11 Ziele, 3 mit Fund** — Commit `d387f4d9`,
2026-10-01 07:37 UTC, go version go1.27.1 darwin/arm64.
Budget je Ziel: **30s**, **2** Arbeiter, `GOMEMLIMIT=1GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | Korpus | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `FuzzInstaller` | `internal/msepaket` | 34 s | 11.683 | 20 | 48 | ohne Fund |
| `FuzzPaketpfad` | `internal/msepaket` | 0 s | 1.302 | 9 | 26 | **Fund**: fuzz_paketpfad_test.go:67: Ort "" und Datei " " → "setpaket:! " → "", "" — `testdata/fuzz/FuzzPaketpfad/b746aba7b57f1b45` |
| `FuzzEinbinden` | `internal/msepaket` | 31 s | 1.202.627 | 41 | 371 | ohne Fund |
| `FuzzSpielmodell` | `internal/msepaket` | 31 s | 665.976 | 18 | 168 | ohne Fund |
| `FuzzStilmodell` | `internal/msepaket` | 31 s | 461.561 | 16 | 176 | ohne Fund |
| `FuzzDekodiereBild` | `internal/msepaket` | 18 s | 55.610 | 24 | 46 | **Fund**: fuzz_bild_test.go:102: png, 12336 × 48: 5003344 Byte zugeteilt, zugesagt 3417260 — `testdata/fuzz/FuzzDekodiereBild/eb8144151f142867` |
| `FuzzSymbolzeichnung` | `internal/msepaket` | 31 s | 600.572 | 15 | 188 | ohne Fund |
| `FuzzUmschlag` | `internal/austausch` | 32 s | 493 | 12 | 31 | ohne Fund |
| `FuzzSprachpaketEinlesen` | `internal/service` | 50 s | 182.702 | 152 | 161 | ohne Fund |
| `FuzzZugfolge` | `internal/service` | 5 s | 4.371 | 16 | 31 | **Fund**: fuzz_zugfolge_test.go:103: objekt.mana abgelehnt ohne Klasse: nicht gefunden: das Objekt "" gibt es in dieser Partie nicht — `testdata/fuzz/FuzzZugfolge/2c132501ff4de5b3` |
| `FuzzHandelsfolge` | `internal/domain` | 39 s | 151.324 | 81 | 91 | ohne Fund |
