# Fuzzkampagne

**30 Ziele, 1 mit Fund, 0 unvollständig** — Commit `61df27a1e`,
2026-10-10 12:39 UTC, go version go1.27.2 darwin/arm64.
Budget je Ziel: **10m**, **5** Arbeiter, `GOMEMLIMIT=2GiB`,
Minimierung höchstens 5 s. Gefahren mit `scripts/fuzzkampagne.mjs`; was
der Bericht misst und was nicht, steht in seinem Kopf. **Anteil** ist der
Anteil der Anweisungen des Pakets, den Saat und erarbeitetes Korpus betreten —
keine Kantenabdeckung, und der Nenner ist das ganze Paket. **Unvollständig**
heißt: ein Teil der Kampagne (Fuzzlauf, Korpusübernahme, Abdeckungslauf) ist
gescheitert; der Strich beim Anteil ist dann keine Messung. **Zuletzt neu**
ist die Sekunde, in der die letzte neue Eingabe kam; **gesättigt** heißt, im
letzten Viertel des Budgets kam keine mehr — ein festgelegtes Maß, kein
Beweis, dass nichts mehr zu finden wäre. **Nicht bewertbar** heißt, der Lauf
endete vor dem Budget (ein Fund) oder seine letzte Meldung kam früher.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | zuletzt neu | Sättigung | Korpus | Anteil | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- | ---: | ---: | --- |
| `FuzzZerteile` | `internal/msepaket` | 601 s | 47.670.640 | 224 | 597 s | wächst noch | 237 | 5.5 % | ohne Fund |
| `FuzzAuswerten` | `internal/msescript` | 601 s | 49.925.871 | 1.240 | 594 s | wächst noch | 1.257 | 34.7 % | ohne Fund |
| `FuzzZerteileSprache` | `internal/msescript` | 600 s | 66.596.589 | 778 | 600 s | wächst noch | 795 | 19.1 % | ohne Fund |
| `FuzzAuspacken` | `internal/bezug` | 607 s | 46.774 | 459 | 600 s | wächst noch | 472 | 39.4 % | ohne Fund |
| `FuzzLiesKarte` | `internal/scryfall` | 601 s | 27.638.871 | 641 | 594 s | wächst noch | 653 | 1.5 % | ohne Fund |
| `FuzzLiesInhalt` | `internal/mtgjson` | 198 s | 19.747.891 | 536 | 198 s | nicht bewertbar | 550 | 50.8 % | **Fund**: fuzz_lesen_test.go:93: ein Befund ohne Set im Ort: "" — "\"\" ist kein Set: json: cannot unmarshal string into Go value of type map[string]jsontext.Value" (§8) — `testdata/fuzz/FuzzLiesInhalt/39bc009b47c09643` |
| `FuzzDokument` | `internal/pdf` | 601 s | 6.399.365 | 6 | 3 s | gesättigt | 12 | 93.8 % | ohne Fund |
| `FuzzBeschaedigtesPaket` | `internal/msepaket` | 602 s | 344.283 | 403 | 600 s | wächst noch | 413 | 12.9 % | ohne Fund |
| `FuzzZerlegeAustausch` | `internal/domain` | 605 s | 41.164.385 | 780 | 558 s | wächst noch | 801 | 3.6 % | ohne Fund |
| `FuzzManifestAus` | `internal/aktualisierung` | 602 s | 26.124.731 | 589 | 597 s | wächst noch | 605 | 3.6 % | ohne Fund |
| `FuzzPruefeRueckruf` | `internal/anmeldung` | 601 s | 30.125.242 | 64 | 345 s | gesättigt | 80 | 16.4 % | ohne Fund |
| `FuzzLiesListe` | `internal/sammelliste` | 601 s | 42.546.371 | 358 | 567 s | wächst noch | 373 | 39.3 % | ohne Fund |
| `FuzzLiesCSV` | `internal/sammelliste` | 601 s | 46.966.286 | 598 | 582 s | wächst noch | 616 | 44.3 % | ohne Fund |
| `FuzzDienstruf` | `internal/rest` | 605 s | 1.868.264 | 305 | 597 s | wächst noch | 319 | 13.6 % | ohne Fund |
| `FuzzTokenPruefen` | `internal/service` | 609 s | 13.717.721 | 83 | 552 s | wächst noch | 99 | 0.2 % | ohne Fund |
| `FuzzEntpacken` | `internal/sicherung` | 601 s | 409.601 | 297 | 600 s | wächst noch | 306 | 34.2 % | ohne Fund |
| `FuzzLoeseAuf` | `internal/domain` | 602 s | 29.434.698 | 680 | 597 s | wächst noch | 695 | 2.2 % | ohne Fund |
| `FuzzLiesKarten` | `internal/mtgapi` | 602 s | 22.143.511 | 548 | 594 s | wächst noch | 563 | 44.9 % | ohne Fund |
| `FuzzLies` | `internal/mtggraphql` | 601 s | 27.462.883 | 573 | 597 s | wächst noch | 590 | 53.3 % | ohne Fund |
| `FuzzLiesKarten` | `internal/scrydex` | 601 s | 28.713.341 | 520 | 594 s | wächst noch | 564 | 19.8 % | ohne Fund |
| `FuzzLies` | `internal/mtgjsonpreise` | 601 s | 26.973.226 | 809 | 597 s | wächst noch | 824 | 74.3 % | ohne Fund |
| `FuzzLiesGepackt` | `internal/mtgjsonpreise` | 601 s | 26.788.027 | 930 | 600 s | wächst noch | 947 | 78.7 % | ohne Fund |
| `FuzzModelle` | `internal/anthropic` | 602 s | 6.496.881 | 626 | 600 s | wächst noch | 642 | 46.9 % | ohne Fund |
| `FuzzModelle` | `internal/openai` | 601 s | 7.114.630 | 628 | 600 s | wächst noch | 641 | 34.1 % | ohne Fund |
| `FuzzModelle` | `internal/xai` | 601 s | 6.788.707 | 600 | 591 s | wächst noch | 613 | 31.2 % | ohne Fund |
| `FuzzModelle` | `internal/ollama` | 602 s | 6.517.826 | 565 | 594 s | wächst noch | 579 | 31.9 % | ohne Fund |
| `FuzzAngebote` | `internal/cardtrader` | 601 s | 5.199.709 | 560 | 597 s | wächst noch | 576 | 32.6 % | ohne Fund |
| `FuzzRahmen` | `internal/leitung` | 601 s | 31.559.861 | 12 | 96 s | gesättigt | 19 | 6.4 % | ohne Fund |
| `FuzzAuspacke` | `internal/peer` | 602 s | 59.788.335 | 640 | 576 s | wächst noch | 646 | 31.5 % | ohne Fund |
| `FuzzFingerabdruck` | `internal/peer` | 601 s | 61.694.350 | 295 | 540 s | wächst noch | 299 | 16.3 % | ohne Fund |
