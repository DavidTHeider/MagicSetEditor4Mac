# Mutationslauf

**12 von 12 gefangen** — Commit `0189cfba7`,
2026-10-09 19:46 UTC, go version go1.27.2 darwin/arm64.
Gefahren mit `node scripts/mutationslauf.mjs`; gefangen ist nur, was der
erwartete Test bei grünem Grundlauf fängt. Was eine Mutation ist und was der
Lauf nicht belegt, steht im Kopf des Skripts.

| | Invariante | Fundstelle | Lauf | Ergebnis | Dauer |
| --- | --- | --- | --- | --- | ---: |
| M1 | eine Rolle reicht nur so weit, wie sie reicht | `internal/domain/akteur.go` | `./internal/domain` | **gefangen** von `TestDieAbsageNenntDieAchse` | 2.1 s |
| M2 | ein nur lesendes Profil schreibt nicht | `internal/domain/akteur.go` | `./internal/domain` | **gefangen** von `TestNurLesenSchlaegtJedeRolle` | 2.1 s |
| M3 | eine Nonce wirkt einmal | `internal/peer/nonceliste.go` | `./internal/peer` | **gefangen** von `TestDasGedaechtnisVergisstNachSeinerFrist` | 0.5 s |
| M4 | eine behauptete Rahmenlänge wird geprüft, bevor sie zugeteilt wird | `internal/leitung/leitung.go` | `./internal/leitung` | **gefangen** von `TestEinBehaupteterRahmenStelltKeinenSpeicher` | 20.9 s |
| M5 | ein Eintrag einer Sicherung führt nicht aus ihr hinaus — auch nicht über einen Umweg | `internal/sicherung/auspacken.go` | `./internal/sicherung` | **gefangen** von `TestEinEintragDerHinausfuehrtWirdAbgewiesen` | 1.4 s |
| M6 | ein falsches Kennwort wird abgewiesen | `internal/kennwort/kennwort.go` | `./internal/kennwort` | **gefangen** von `TestEinFalschesKennwortIstKeinFormfehler` | 0.4 s |
| M7 | eine angewendete Migration wird nicht nachträglich verändert | `internal/storage/migrationsbuchfuehrung.go` | `./internal/storage` | **gefangen** von `TestEineNachtraeglichGeaenderteMigrationWirdAbgewiesen` | 93.9 s |
| M8 | eine doppelt zugestellte Partienachricht wirkt einmal | `internal/service/partieempfang.go` | `./internal/service` | **gefangen** von `TestEineAlteEpocheWirdGemeldetUndFaelltWeg` | 13.6 s |
| M9 | ein Handel wird nicht halb verbucht | `internal/service/handelsverbuchung.go` | `./internal/service` | **gefangen** von `TestEineZuKleineZusageBuchtNicht` | 13.6 s |
| M10 | eine Sitzung staut höchstens so viel, wie ihre Tiefe erlaubt | `internal/service/sitzungsgriff.go` | `./internal/service` | **gefangen** von `TestEineVolleWarteschlangeStautZurueckUndWirftWeg` | 13.7 s |
| M11 | eine Einlösung einer neueren Fassung wird als solche erkannt | `internal/domain/einloesung.go` | `./internal/domain` | **gefangen** von `TestEineNeuereEinloesefassungWirdBenannt` | 2.8 s |
| M12 | eine beendete Partiesitzung wird freigegeben | `internal/app/partieaufnahme.go` | `./internal/app` (`-run Partie|Tisch|Beitritt|Aufstehen|Abriss`) | **gefangen** von `TestNachDemAbrissGibtDerGastDieSitzungFrei` | 13.0 s |
