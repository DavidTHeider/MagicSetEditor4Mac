# Beleg: Zufällige Mutation der Kernregeln — B05

**Commit `80ccb6c49`**, 2026-10-10 00:48 UTC, Mac17,3, macOS 27.0.1,
Befehl `node scripts/zufallsmutation.mjs --samen 2224 --je 40` — Samen 2224, je Paket 40 aus 17638 Stellen.

**Quote gefangen ÷ geeignet: 99,1 %** (111 von 112), 95-%-Intervall nach Wilson 95,1 % bis 99,8 %. Geeignet ist, was übersetzt, in der Frist lief und nicht äquivalent ist; **„nicht übersetzt" schließt ein, was `go vet` im Testlauf abweist** (etwa ein `a != x || a != y`). **Die Lücken sind seitdem geschlossen**: je Überlebendem steht der Test, der ihn jetzt fängt — gemessen ist der Stand davor. **Keine Hochrechnung aufs Gesamtprojekt**: die Zahl gilt für `domain`, `service` und `storage` und die fünf Arten aus `scripts/mutationsorte.go`.

| Paket | gefangen | Lücke | äquivalent | nicht übersetzt | Frist | Quote |
| --- | --- | --- | --- | --- | --- | --- |
| internal/domain | 37 | 1 | 0 | 2 | 0 | 97,4 % |
| internal/service | 37 | 0 | 1 | 2 | 0 | 100,0 % |
| internal/storage | 37 | 0 | 2 | 0 | 1 | 100,0 % |

| Art | gefangen | Lücke | äquivalent | nicht übersetzt | Frist | Quote |
| --- | --- | --- | --- | --- | --- | --- |
| bedingung | 49 | 0 | 0 | 0 | 0 | 100,0 % |
| rechnung | 5 | 1 | 1 | 3 | 0 | 83,3 % |
| vergleich | 45 | 0 | 0 | 0 | 1 | 100,0 % |
| verknuepfung | 7 | 0 | 0 | 1 | 0 | 100,0 % |
| wahrheitswert | 5 | 0 | 2 | 0 | 0 | 100,0 % |

## Überlebende (4)

| # | Stelle | Mutation | Durchsicht |
| --- | --- | --- | --- |
| 31 | `internal/domain/ersetzungseffekt.go:115` | `+` → `-` | lücke — nicht äquivalent, wie zuerst eingetragen: ein eingelesener Stand mit den Stempeln 1 und 2 bekäme mit len−1 wieder die 1. Dahinter stand ein Fehler des Originals selbst — nach dem Aufräumschritt bekam ein neuer Effekt den Stempel eines lebenden, und einer verhinderte nichts. Behoben mit naechsterZeitstempel (Ersetzungen und Dauereffekte), seitdem TestEinNeuerEffektNachDemZugendeTraegtEinenEigenenStempel und TestEinNeuerDauereffektNachDemZugendeTraegtEinenEigenenStempel (Nachlauf über Phase B) |
| 62 | `internal/service/setangaben.go:57` | `+` → `-` | äquivalent — nur die Größenangabe einer Map (`make(map…, len(felder)-len(werte))`): eine negative behandelt die Laufzeit wie null — gemessen mit Go 1.27.2 —, und gelesen wird die Map unabhängig davon |
| 82 | `internal/storage/setwertaenderung.go:88` | `false` → `true` | äquivalent — der zweite Rückgabewert (`hatte`) neben einem Fehler: jeder Aufrufer prüft zuerst den Fehler (`internal/mcp/setangaben.go:121`, `internal/app/setangaben.go:148`, `internal/service/symbolbibliothekwege.go:108`) und liest `hatte` dann nicht |
| 93 | `internal/storage/mindestform.go:138` | `true` → `false` | äquivalent — `angelegt` wird nur über seine Schlüssel gelesen (`for name := range angelegt`), sein Wert nie; die Gegenprobe am Schaden (`TestEineUmbenannteTabelleWirdUnterIhremNeuenNamenVerlangt`) blieb mit dem Mutanten grün |

Rohdaten: `docs/belege/zufallsmutation.json`; Durchsicht: `docs/belege/zufallsmutation-durchsicht.json`.
