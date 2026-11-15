# Fuzzkampagne

**60 Ziele, 3 mit Fund, 0 unvollständig** — Commit `a397d98dc` (sauber gegen den Commit, Saat `2467374e423c`),
2026-10-11 02:37 UTC, go version go1.27.2 darwin/arm64.
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
| `FuzzZerteile` | `internal/msepaket` | 607 s | 45.934.332 | 210 | 468 s | wächst noch | 223 | 5.5 % | ohne Fund |
| `FuzzAuswerten` | `internal/msescript` | 603 s | 48.093.950 | 1.280 | 600 s | wächst noch | 1.297 | 35.0 % | ohne Fund |
| `FuzzZerteileSprache` | `internal/msescript` | 600 s | 63.171.777 | 803 | 588 s | wächst noch | 820 | 19.1 % | ohne Fund |
| `FuzzAuspacken` | `internal/bezug` | 608 s | 49.094 | 441 | 600 s | wächst noch | 454 | 40.2 % | ohne Fund |
| `FuzzLiesKarte` | `internal/scryfall` | 601 s | 29.908.113 | 630 | 591 s | wächst noch | 642 | 1.5 % | ohne Fund |
| `FuzzLiesInhalt` | `internal/mtgjson` | 601 s | 49.718.193 | 639 | 600 s | wächst noch | 654 | 50.3 % | ohne Fund |
| `FuzzDokument` | `internal/pdf` | 601 s | 6.745.502 | 8 | 198 s | gesättigt | 14 | 93.8 % | ohne Fund |
| `FuzzBeschaedigtesPaket` | `internal/msepaket` | 602 s | 341.561 | 414 | 600 s | wächst noch | 424 | 12.2 % | ohne Fund |
| `FuzzZerlegeAustausch` | `internal/domain` | 604 s | 39.371.639 | 800 | 588 s | wächst noch | 821 | 3.6 % | ohne Fund |
| `FuzzManifestAus` | `internal/aktualisierung` | 604 s | 28.923.333 | 639 | 600 s | wächst noch | 655 | 3.6 % | ohne Fund |
| `FuzzPruefeRueckruf` | `internal/anmeldung` | 601 s | 30.130.717 | 66 | 276 s | gesättigt | 82 | 16.4 % | ohne Fund |
| `FuzzLiesListe` | `internal/sammelliste` | 601 s | 42.470.719 | 364 | 573 s | wächst noch | 379 | 39.3 % | ohne Fund |
| `FuzzLiesCSV` | `internal/sammelliste` | 601 s | 49.775.442 | 604 | 591 s | wächst noch | 622 | 44.3 % | ohne Fund |
| `FuzzDienstruf` | `internal/rest` | 605 s | 1.589.818 | 369 | 600 s | wächst noch | 383 | 13.4 % | ohne Fund |
| `FuzzTokenPruefen` | `internal/service` | 612 s | 15.585.415 | 78 | 210 s | gesättigt | 94 | 0.2 % | ohne Fund |
| `FuzzEntpacken` | `internal/sicherung` | 602 s | 428.470 | 284 | 597 s | wächst noch | 293 | 33.8 % | ohne Fund |
| `FuzzLoeseAuf` | `internal/domain` | 601 s | 47.359.086 | 649 | 591 s | wächst noch | 664 | 2.2 % | ohne Fund |
| `FuzzLiesKarten` | `internal/mtgapi` | 601 s | 28.900.329 | 573 | 591 s | wächst noch | 588 | 44.9 % | ohne Fund |
| `FuzzLies` | `internal/mtggraphql` | 601 s | 27.353.928 | 555 | 600 s | wächst noch | 572 | 53.3 % | ohne Fund |
| `FuzzLiesKarten` | `internal/scrydex` | 601 s | 28.963.323 | 532 | 537 s | wächst noch | 576 | 19.8 % | ohne Fund |
| `FuzzLies` | `internal/mtgjsonpreise` | 601 s | 23.815.563 | 856 | 564 s | wächst noch | 871 | 75.7 % | ohne Fund |
| `FuzzLiesGepackt` | `internal/mtgjsonpreise` | 601 s | 27.029.388 | 986 | 600 s | wächst noch | 1.003 | 78.7 % | ohne Fund |
| `FuzzModelle` | `internal/anthropic` | 601 s | 6.164.555 | 677 | 594 s | wächst noch | 693 | 47.6 % | ohne Fund |
| `FuzzModelle` | `internal/openai` | 603 s | 5.160.633 | 626 | 600 s | wächst noch | 639 | 34.7 % | ohne Fund |
| `FuzzModelle` | `internal/xai` | 601 s | 6.326.155 | 647 | 585 s | wächst noch | 660 | 31.8 % | ohne Fund |
| `FuzzModelle` | `internal/ollama` | 602 s | 6.791.628 | 624 | 600 s | wächst noch | 638 | 31.9 % | ohne Fund |
| `FuzzAngebote` | `internal/cardtrader` | 602 s | 5.757.416 | 587 | 600 s | wächst noch | 603 | 32.6 % | ohne Fund |
| `FuzzRahmen` | `internal/leitung` | 601 s | 35.954.541 | 10 | 54 s | gesättigt | 17 | 7.5 % | ohne Fund |
| `FuzzAuspacke` | `internal/peer` | 602 s | 62.105.077 | 639 | 597 s | wächst noch | 645 | 32.6 % | ohne Fund |
| `FuzzFingerabdruck` | `internal/peer` | 601 s | 62.863.353 | 260 | 507 s | wächst noch | 264 | 16.3 % | ohne Fund |
| `FuzzZerlegeKontakteinladung` | `internal/domain` | 601 s | 38.298.689 | 462 | 600 s | wächst noch | 471 | 2.1 % | ohne Fund |
| `FuzzZerlegeEinloesung` | `internal/domain` | 601 s | 40.526.633 | 680 | 588 s | wächst noch | 690 | 2.1 % | ohne Fund |
| `FuzzZerlegeEinloeseantwort` | `internal/domain` | 601 s | 45.220.106 | 690 | 600 s | wächst noch | 698 | 2.1 % | ohne Fund |
| `FuzzZerlegeEinladung` | `internal/domain` | 601 s | 47.195.513 | 8 | 54 s | gesättigt | 18 | 1.9 % | ohne Fund |
| `FuzzZerlegeAdresse` | `internal/domain` | 601 s | 39.354.165 | 242 | 522 s | wächst noch | 256 | 2.6 % | ohne Fund |
| `FuzzDeute` | `internal/entdeckung` | 603 s | 37.815.669 | 429 | 582 s | wächst noch | 434 | 8.5 % | ohne Fund |
| `FuzzLiesKontrolle` | `internal/domain` | 601 s | 32.845.516 | 681 | 594 s | wächst noch | 689 | 2.1 % | ohne Fund |
| `FuzzPartienachricht` | `internal/domain` | 601 s | 42.763.063 | 646 | 591 s | wächst noch | 653 | 1.9 % | ohne Fund |
| `FuzzEmpfaengerstand` | `internal/domain` | 601 s | 50.656.635 | 1 | 3 s | gesättigt | 4 | 2.0 % | ohne Fund |
| `FuzzChat` | `internal/domain` | 601 s | 44.232.544 | 93 | 93 s | gesättigt | 100 | 1.9 % | ohne Fund |
| `FuzzWiederaufnahme` | `internal/domain` | 601 s | 42.953.851 | 6 | 3 s | gesättigt | 11 | 2.0 % | ohne Fund |
| `FuzzHandelsumschlag` | `internal/domain` | 601 s | 35.975.086 | 238 | 552 s | wächst noch | 241 | 2.1 % | ohne Fund |
| `FuzzInstaller` | `internal/msepaket` | 71 s | 26.369 | 101 | 69 s | nicht bewertbar | 109 | 12.3 % | **Fund**: fuzz_installer_test.go:86: in den Puffer schreiben: zip: write to directory — `testdata/fuzz/FuzzInstaller/9c6f04f43a8d5ff1` |
| `FuzzPaketpfad` | `internal/msepaket` | 600 s | 35.281.562 | 371 | 474 s | wächst noch | 387 | 2.1 % | ohne Fund |
| `FuzzEinbinden` | `internal/msepaket` | 601 s | 41.665.438 | 604 | 594 s | wächst noch | 612 | 10.1 % | ohne Fund |
| `FuzzSpielmodell` | `internal/msepaket` | 601 s | 47.424.783 | 283 | 594 s | wächst noch | 287 | 10.8 % | ohne Fund |
| `FuzzStilmodell` | `internal/msepaket` | 602 s | 43.646.956 | 328 | 600 s | wächst noch | 331 | 14.2 % | ohne Fund |
| `FuzzDekodiereBild` | `internal/msepaket` | 601 s | 7.508.918 | 690 | 600 s | wächst noch | 705 | 1.9 % | ohne Fund |
| `FuzzSymbolzeichnung` | `internal/msepaket` | 285 s | 17.909.675 | 233 | 282 s | nicht bewertbar | 238 | 13.2 % | **Fund**: fuzz_symbol_test.go:87: Schreiben ist kein Fixpunkt: — `testdata/fuzz/FuzzSymbolzeichnung/a2901ac8792fc369` |
| `FuzzUmschlag` | `internal/austausch` | 602 s | 21.757 | 556 | 600 s | wächst noch | 564 | 66.7 % | ohne Fund |
| `FuzzSprachpaketEinlesen` | `internal/service` | 40 s | 1.031.637 | 264 | 39 s | nicht bewertbar | 273 | 0.4 % | **Fund**: fuzz_sprachpaket_test.go:87: abgewiesen ohne Feld oder Schlüssel: Eine Übersetzung hat zwischen einem und 4000 Zeichen. — `testdata/fuzz/FuzzSprachpaketEinlesen/b257223c79af9260` |
| `FuzzZugfolge` | `internal/service` | 601 s | 958.486 | 397 | 597 s | wächst noch | 428 | 1.5 % | ohne Fund |
| `FuzzHandelsfolge` | `internal/domain` | 602 s | 19.970.894 | 181 | 600 s | wächst noch | 191 | 3.7 % | ohne Fund |
| `FuzzBildangabe` | `internal/service` | 604 s | 6.471.916 | 585 | 600 s | wächst noch | 593 | 0.3 % | ohne Fund |
| `FuzzTextskript` | `internal/msescript` | 601 s | 22.500.473 | 549 | 600 s | wächst noch | 560 | 22.5 % | ohne Fund |
| `FuzzKartenskript` | `internal/service` | 602 s | 10.711.428 | 992 | 594 s | wächst noch | 1.002 | 0.5 % | ohne Fund |
| `FuzzEntpackenVerschluesselt` | `internal/sicherung` | 601 s | 537.053 | 252 | 600 s | wächst noch | 261 | 57.4 % | ohne Fund |
| `FuzzUmschlagdeck` | `internal/app` | 695 s | 1.331.774 | 557 | 600 s | wächst noch | 563 | 5.8 % | ohne Fund |
| `FuzzHandschlag` | `internal/peer` | 601 s | 9.382.388 | 555 | 600 s | wächst noch | 565 | 71.9 % | ohne Fund |
| `FuzzSpielbefehl` | `internal/service` | 602 s | 4.833.375 | 303 | 594 s | wächst noch | 555 | 3.0 % | ohne Fund |
