# 0094 — Der Race-Detektor misst nicht wazero

**Stand:** angenommen · **Datum:** 2026-09-11 · **Anlass:** die Pipeline hielt auf

## Das Problem

`make ci` brauchte am 2026-09-11 rund **fünfzig Minuten**, und davon **42 die
Teststufe**. Ein Werkstück in einer Nacht kostete damit vier Stunden Warten —
und in dieser Nacht waren es fünf Läufe für **einen** Pull Request.

**Keiner der Funde dieser fünf Läufe brauchte die volle Pipeline**: eslint, die
Zeilengrenze, zwei Abdeckungslücken und eine falsche Zahl in der Dokumentation.
Alle vier findet eine Prüfung von fünf Minuten.

## Was gemessen wurde

Aus einem vollen Lauf abgelesen:

| Paket | mit `-race` | ohne | Faktor | Goroutinen im Produktivcode |
| --- | --- | --- | --- | --- |
| `internal/storage` | **1357 s** | 28,5 s | **48×** | **0** |
| `internal/app` | 1194 s | — | — | 7 |
| `cmd/mse` | 512 s | — | — | — |
| `internal/werkzeuge` | **372 s** | 30,4 s | **12×** | **0** |

`storage` und `werkzeuge` zusammen: **29 Minuten** für eine Aussage über
**null** Goroutinen.

## Warum gerade diese beiden so teuer sind

**SQLite läuft in diesem Projekt als WebAssembly.** Der Treiber ist
`ncruces/go-sqlite3`, und er führt SQLite über `wazero` aus — einen
WebAssembly-Interpreter in reinem Go. Der Race-Detektor instrumentiert **jeden**
Speicherzugriff, den er sieht; in einem Interpreter sind das die Zugriffe auf
den simulierten Speicher der Gastmaschine, und davon gibt es je SQL-Anweisung
sehr viele.

**Er misst damit wazero und nicht diesen Code.**

## Die Entscheidung

`scripts/ci.sh` fährt zwei Testläufe statt eines:

* **mit `-race`** — alle Pakete außer denen in `OHNE_RENNEN`;
* **ohne** — die aus `OHNE_RENNEN`.

Beide mit leerem Zuhause, beide mit Abdeckungsprofil; die zwei Profile werden
vereint, und die Schwellen gelten unverändert über das Ganze.

`OHNE_RENNEN` führt heute `./internal/storage/...` und
`./internal/werkzeuge/...`.

## Was das kostet

**Ein Wettlauf im Produktivcode dieser beiden Pakete bliebe unentdeckt.**

Das ist der Preis, und er ist nur deshalb klein, weil dort **nichts** nebenläufig
ist: kein `go func`, kein `sync.`. Damit das so bleibt, hält
`internal/werkzeuge/rennliste_test.go` die Liste gegen den Quelltext — ein Paket,
das eine Goroutine oder ein `sync.`-Mittel bekommt, macht die Pipeline rot und
muss zurück in den Race-Lauf.

**Die Gegenprobe ist gemessen**: ein eingeschmuggeltes `sync.Mutex` in
`internal/storage` macht den Wächter rot.

**Was er nicht hält**, steht neben ihm: einen Wettlauf im **Testcode**. Die
Tests dieser Pakete laufen nebenläufig (`t.Parallel`), und ein Doppel mit
gemeinsamem Zustand ist weiterhin möglich. Geprüft wird die Aussage über den
Produktivcode, auf der die Ausnahme beruht.

## Der zweite Teil: ein Riegel davor

`make vorpruefung` (`scripts/vorpruefung.sh`) fährt in rund fünf Minuten:

* `go vet` und `golangci-lint` über alles,
* die Zeilengrenze aus §2,
* `go test` über die acht Pakete mit harter Schwelle, **mit Zählung der
  unbetretenen Anweisungen** — dieselbe Rechnung wie in der Pipeline, nur früher,
* `tsc`, `eslint`, `prettier` und `vitest` im Frontend.

**Er ersetzt nichts.** Was dort fehlt, zählt sein eigener Kopf auf: die Läufe
mit leerem Zuhause, die Übersetzung für Windows und Linux, `fremdpakete` und
`netzlauf`, `gosec`, `govulncheck`, `npm audit`, die Stückliste, die
Frontend-Abdeckung und die Skriptprüfungen.

## Der dritte Teil: der Rhythmus

**Ein `make ci` je Push, nicht je Werkstück.** §4 sagt, dass nichts gepusht
wird, was nicht grün war — nicht, wie oft gepusht wird. Mehrere Werkstücke einer
Phase gehen auf einen Zweig, dazwischen läuft der Riegel, und die volle Pipeline
läuft einmal davor.

**Was das kostet**: wird sie doch rot, ist die Fläche, in der zu suchen ist,
größer. Dagegen steht die Messung dieser Nacht — von fünf roten Läufen fand
**keiner** etwas, das der Riegel nicht auch gefunden hätte.

## Was verworfen wurde

**`-race` ganz abschalten.** `internal/service` hält Guthaben, Auftragsbücher
und Haushalte über Goroutinen; `internal/app` sieben Stellen. Dort hat der
Detektor am 2026-08-31 einen echten Wettlauf gefunden (ein `int`-Zähler in einem
Testdoppel). Er bleibt, wo er etwas sieht.

**Die Testläufe parallel fahren.** Sie laufen schon parallel — `go test` über
mehrere Pakete verteilt von sich aus. Was bremst, ist die Instrumentierung
innerhalb eines Pakets, und die teilt sich nicht auf.
