# Die Pipeline

**Verlegt aus `AGENTS.md` am 2026-09-22** — ENG-01.AK-04
(#1726):
„Detailkonzepte wandern in passende Dokumente statt beide Regeldateien
weiter mit ganzen Projektplänen zu überladen." **Die Regel steht weiter
dort** — hier steht, warum sie so lautet, samt jeder Messung.

---

Ein Einstiegspunkt, lokal wie in GitHub Actions:

```bash
make ci                 # alle Stufen
./scripts/ci.sh lint    # eine Stufe
```

Der GitHub-Workflow ruft **dasselbe Skript** auf. Wer eine Prüfung ergänzt,
ergänzt sie in `scripts/ci.sh` — nicht in der Workflow-Datei.

Stufen: `format` → `types` → `lint` → `test` → `security`.

**Die Teststufe fährt zwei Läufe**, nicht einen: mit Race-Detektor über alles,
was Goroutinen hat, und ohne über `internal/storage` und `internal/werkzeuge`.
Dort kostete er das **48-fache** bei **null** Goroutinen im Produktivcode —
SQLite läuft hier als WebAssembly, und der Detektor instrumentiert den
Interpreter statt diesen Code
([0094](entscheidungen/0094-der-detektor-misst-nicht-wazero.md)). Ein
Wächter hält die Liste ehrlich: ein Paket, das eine Goroutine bekommt, muss
zurück.

**Und der Lauf mit Detektor geht für `internal/app` in sechs Teile**
(ENG-03, [`WARTBARKEIT.md`](WARTBARKEIT.md)): dasselbe Testprogramm, einmal
übersetzt, in sechs Prozessen mit je einem Sechstel der Tests
(`scripts/teilrennen.sh`). **`internal/app` war die Teststufe** — 1647, 1566
und 1489 Sekunden in drei Läufen vom 2026-09-24, bei einer Frist von 1800,
während kein anderes Paket mehr als 517 brauchte und neun von zehn Kernen
stillstanden. **Geteilt: 511 Sekunden** allein gemessen — mit vier Teilen
waren es 610, weil die Teile ungleich schwer sind — und **829** in der
Pipeline, neben allen anderen Paketen. Der ganze Lauf brauchte davor 37 und
29 Minuten (ENG-02, #1824), mit dem geteilten **20** — gemessen am
2026-09-24 am Anlegen und letzten Schreiben des Protokolls.

**Seitdem ist die Maschine in der Teststufe voll**, und die drei Wächter,
die Wanduhrzeiten ins Verhältnis setzten, messen deshalb Rechenzeit,
abwechselnd und ohne Speicherbereinigung ([`WARTBARKEIT.md`](WARTBARKEIT.md))
— der dritte erst seit dem Zweitblick über Phase 6.2, der ihn anstrich. **Was die Teilung kostet**, steht im
Kopf des Skripts: zwei nebenläufige Tests in verschiedenen Teilen sieht der
Detektor nicht mehr zusammen.

**Ein Teil, der in seine Frist läuft, nennt, was noch lief** — seit dem
2026-10-04 (Zweitblick über Phase 6.11). Ein Teil von `internal/app` hing
dreißig Minuten, und der Bericht nannte nur den zuletzt begonnenen Test,
der allein in vier Sekunden grün ist; Go schreibt die laufenden Tests selbst
unter „running tests:", und genau dieser Block fiel weg. Der Teil allein
nachgefahren war grün (658 Tests, der langsamste 14 s) — **die Ursache ist
nicht gefunden**, der nächste Hänger nennt sie.

**Die Pipeline läuft vor dem Push, nicht darin.**

```bash
make ci && git push
```

**Und davor der Riegel**, der in **sechs** Minuten findet, was die Pipeline in
**einundzwanzig** findet:

```bash
make vorpruefung
```

**Beide Zahlen sind gemessen und nicht geschätzt** — am 2026-09-14 mit
`/usr/bin/time` auf diesem Rechner, **allein**: `make vorpruefung` 369 s,
`make ci` 1.278 s. Hier standen vorher „fünf" und „fünfzig"; die erste war zu
niedrig, die zweite mehr als doppelt zu hoch, und **keine von beiden war je
gemessen worden**. Wer sie ändert, misst wieder — und misst dabei nichts
anderes nebenher laufen, sonst misst er die Maschine und nicht den Lauf.

Er fährt `go vet`, `golangci-lint`, **`gosec`**, die Zeilengrenze, die acht
Pakete mit harter Schwelle **samt Zählung der unbetretenen Anweisungen**, dazu
`tsc`, `eslint`, `prettier` und `vitest` **mit Abdeckung**.

**Die Abdeckung des Fensters kam am 2026-09-16 dazu**, aus demselben Grund wie
`gosec`: sie kostet **44 Sekunden** und spart einundzwanzig Minuten — gemessen
am selben Tag, `vitest run` 78 s gegen `vitest run --coverage` 122 s. Hier
stand vorher „ohne Abdeckung", und die Schwelle von 100 % stand nur in
`make ci`; zweimal an einem Vormittag fiel dort ein Zweig um, den dieser Lauf
nicht gemessen hatte. **Die Go-Seite war schon drin** — die Hälfte einer
Zusage ist schlimmer als keine (§10).

**`gosec` kam am 2026-09-14 dazu**, weil er **3,2 Sekunden** kostet und
einundzwanzig Minuten spart — dreimal gemessen, mit dem Umfang der Pipeline
(`./internal/... ./cmd/... .`) und nicht mit `./...`: jenes läuft vom
Wurzelverzeichnis durch `frontend/node_modules` und kommt nicht wieder.

**Der Riegel selbst maß am selben Tag 369 s und 409 s** — der Unterschied ist
**Streuung zwischen zwei Läufen und nicht diese Stufe**. Wer die vierzig
Sekunden ihr zuschriebe, schriebe einer Ursache zu, was keine ist (§10). Er stand vorher unter „was hier fehlt", und genau
dort ist ein Fund hindurchgefallen: `//nolint:gosec` ist eine Anweisung an
**golangci-lint**; `gosec` läuft eigenständig und liest sie nicht. Der Riegel
war grün, die Stufe `security` rot — zwanzig Minuten später und ohne
Zusammenhang zu dem, woran gerade jemand saß
(#1455). **Was er nicht fährt, zählt sein eigener
Kopf auf** — er ersetzt `make ci` nicht, er verkürzt den Weg dorthin
([0094](entscheidungen/0094-der-detektor-misst-nicht-wazero.md)).

**Seit dem 2026-09-13 lintet er auch über die Build-Tags.** Er tat es nicht,
und das hat eine Nacht gekostet: zwei Befunde in den beiden `netzlauf`-Läufen
von MTGJSON waren hier grün und haben `make ci` rot gemacht — zehn Minuten
später und ohne Zusammenhang zu dem, woran gerade jemand saß. Genau das soll
dieser Riegel verhindern; derselbe Fund gab es 2026-08-19 schon einmal, damals
in `make ci` selbst (#526).

**Über die Pakete, die den Tag tragen, und nicht über alle.** Gemessen am
2026-09-13 bei warmem Zwischenspeicher kostet das **zwei Sekunden** statt elf.
Tragen darf die Einengung nur, weil **jede** Datei hinter den Tags eine
Testdatei ist — ein Tag ändert damit die Übersetzung seines **eigenen** Pakets
und keines anderen. Wer eine Produktivdatei hinter einen Tag legt, hat diese
Überlegung aufgehoben; dann gilt wieder der volle Lauf aus `make ci`.

**Die Paketliste wird berechnet und nicht eingetragen** — eine Aufzählung
veraltete beim nächsten Paket, das einen Tag bekommt, und wäre still zu kurz.
`scripts/vorpruefung.test.sh` hält es, samt der Gegenprobe: eine neue markierte
Datei muss in der Liste auftauchen.

**Ein `make ci` je Push, nicht je Werkstück.** Der Satz oben sagt, dass nichts
gepusht wird, was nicht grün war — nicht, wie oft gepusht wird. Mehrere
Werkstücke einer Phase gehen auf einen Zweig, dazwischen läuft der Riegel, und
die volle Pipeline läuft **einmal** davor. In der Nacht zum 2026-09-11 waren es
fünf Läufe für einen Pull Request, und **kein** Fund davon hätte den Riegel
überlebt.

Der `pre-push`-Hook (`make hooks`) **prüft** dann nur noch, ob sie für genau
diesen Stand grün war — er rechnet nicht mehr. `--no-verify` ist für Notfälle
und muss im PR begründet werden.

**Der Grund ist gemessen und nicht überlegt**
(#1389).
Hier stand bis zum 2026-09-02 „vor jedem Push läuft die Pipeline über den
Hook", und **das hat nicht funktioniert**: `git` öffnet die SSH-Verbindung zu
`git-receive-pack`, bevor es den Hook ruft — mit `GIT_TRACE=1` abgelesen liegen
1,8 Sekunden dazwischen. Zehn Minuten Pipeline im Hook halten diese Verbindung
offen, GitHub schließt sie, und der Paketversand endet an SIGPIPE:

```
HOOK: warte 600 Sekunden ab 21:33:45
Connection to github.com closed by remote host.
HOOK: fertig 21:43:45, gebe frei
RÜCKGABE=141
```

**Ohne Meldung von git**, ohne `To github.com`. Bei 300 Sekunden ging derselbe
Weg noch durch; die Grenze liegt dazwischen. Der Preis war `--no-verify` als
Gewohnheit — und damit fiel auch die Sperre gegen den direkten Push auf `main`.

**Die Zusage ändert sich dabei nicht**: nichts wird gepusht, was nicht grün
war. Nur der Ort der Arbeit ändert sich. `scripts/ci.sh all` schreibt nach
einem grünen Lauf `.git/mse-pipeline` mit der Kennung des Commits und der
Angabe, ob der Baum dabei **sauber** war; der Hook weist ab, wenn eines von
beidem nicht stimmt.

### Wohin die Zeit geht — gemessen am 2026-10-01 (QA-07)

**Drei Stufen und ihr Anlass** (#1738, QA-07.AK-02):

| Stufe | Aufruf | Dauer | Wann | Pflicht |
| --- | --- | ---: | --- | --- |
| **Riegel** | `make vorpruefung` | rund 6 min | nach jedem Werkstück, vor `make ci` | Lint, Zeilengrenze, Kernpakete samt unbetretener Anweisungen, Fenster mit Abdeckung — **nicht** die Racetests, die Flächen und `security` |
| **Vollprüfung** | `make ci` | siehe unten | vor jedem Push; der Hook verlangt ihren Stempel | alles, was eine Fassung tragen muss |
| **Kampagnen** | `make fuzzkampagne`, `make mutationslauf`, `make fremdpakete`, `go test -tags netzlauf`, `make huellenprobe`, `make webkitabnahme` | Minuten bis Stunden | vor einer Freigabe und wenn ihr Gegenstand sich ändert | nicht im Push; ihr Beleg steht in `docs/belege/` und altert sichtbar (`make belege`) |

**Gemessen, was ein Lauf kostet** (QA-07.AK-01). Zwei grüne Läufe derselben
Nacht brauchten **22 und 32 Minuten**, und das Protokoll sagte nicht, wo die
zehn Minuten blieben. **Seitdem nennt jeder Lauf seine Stufendauern**
(`Dauer je Stufe: …` vor der Schlusszeile, `scripts/ci-beweis.sh`).

**Der lange Pol der Teststufe war `cmd/mse`** — und zwar auf **einem** Kern:

| Paket | Lauf A | Lauf B | Lauf C |
| --- | ---: | ---: | ---: |
| `cmd/mse` | 1.168 s | 1.120 s | 1.139 s |
| `internal/app`, geteilt in sechs | 1.226 s | 1.190 s | 1.293 s |
| `internal/flaechen` | 290 s | — | 311 s |
| `internal/service` | 86 s | — | 121 s |

**Der Grund**: 259 Tests, von denen 7 `t.Parallel` tragen. Die übrigen setzen
`HOME` mit `t.Setenv` (`mitKarte` in `karten_test.go`), und das geht nur je
Prozess. In der Prozessliste stand am Ende der Teststufe allein `mse.test`
mit 98,6 % — ein Kern von zehn —, und alles andere wartete.

**Die Maßnahme ist dieselbe wie bei `internal/app` (ENG-03)**: `cmd/mse` läuft
jetzt in sechs Prozessen (`GETEILT` in `scripts/ci.sh`). Dasselbe
Testprogramm, einmal übersetzt, jeder Test in genau einem Teil — `t.Setenv`
gilt je Prozess und stört dort nicht. **Was es kostet**, steht in
`scripts/teilrennen.sh`: einen Wettlauf zwischen zwei Tests verschiedener
Teile sieht der Detektor nicht; zwischen zwei Tests ohne `t.Parallel` sah er
ohnehin keinen.

**Und was es gebracht hat — gemessen, nicht erhofft.** Der erste Lauf mit
der Teilung brauchte **1.919 s** (Format 6 · Typisierung 55 · Linting 143 ·
Tests 1.696 · Sicherheit 19), der letzte davor **1.917 s**. **Kein Gewinn an
Zeit**, und der Grund steht im selben Protokoll: die Maschine ist in der
Teststufe ausgelastet. `cmd/mse` war nur der **letzte** Läufer — am Ende
allein auf einem Kern —, nicht die Ursache; geteilt rechnen jetzt zwölf
Prozesse (sechs von `app`, sechs von `cmd/mse`) auf zehn Kernen, und die
übrigen Pakete wurden langsamer (`render` 364 s, `flaechen` 348 s,
`service` 152 s). Die geteilten zusammen: 1.263 s, so lang wie `app` allein.

**Gewonnen ist Abstand zur Frist**: `cmd/mse` lag ungeteilt bei 1.120 bis
1.168 s von 1.800 (`-test.timeout 30m`) und wuchs mit jedem Befehl; je Teil
sind es jetzt rund 200 Sekunden. Deshalb bleibt die Teilung.

**Das Ziel „40 % schneller" ist verfehlt** (QA-07.AK-04 verlangt, das samt
Ursache aufzuschreiben, und es ist kein Freigabekriterium). **Der Hebel liegt
in der Rechenzeit, nicht in der Verteilung**: jeder der 259 Tests in
`cmd/mse` legt einen frischen Bestand an und migriert ihn — unter `-race` und
mit Abdeckung. Ein einmal migrierter Bestand, je Test kopiert statt neu
migriert, wäre der nächste Schritt; er ist gemessen erst, wenn er gebaut ist.

**Gebaut und gemessen am 2026-10-01** (QA-07b,
#1968):
- Ein **Musterbestand** wird je Testlauf einmal migriert und je Test kopiert (`cmd/mse/musterbestand_test.go`).
- Die Wanderungen laufen beim ersten Befehl weiterhin. Über einem fertigen Bestand sind sie ein Blick in ihre Tafel.
- Gemessen ist `go test -race -covermode=atomic ./cmd/mse`, allein auf der Maschine, abwechselnd vorher und nachher. Beide Testprogramme waren vorher übersetzt.

| | vorher (main, `6aaf768c`) | nachher |
| --- | ---: | ---: |
| Wanduhr, Lauf 1 / Lauf 2 | 553,7 s / 562,6 s | 76,9 s / 76,0 s |
| Rechenzeit (user), Lauf 1 / Lauf 2 | 528,8 s / 538,6 s | 54,1 s / 52,9 s |

**Rund 86 % weniger Wanduhr und 90 % weniger Rechenzeit** — für dieses
Paket. Was die Vollprüfung davon hat, sagt ihr eigenes Protokoll
(`Dauer je Stufe`); die Zeile steht im PR, der es gebaut hat.

**Ein Test legt seinen Bestand weiter selbst an**, wenn er den Ort selbst
braucht. `TestEinNichtOeffenbarerBestandEndetFachlich` stellt ein
Verzeichnis an die Stelle der Datei; mit dem Musterbestand stand dort schon
eine, und der Test wurde rot. Gefunden hat das die Messung.

### Ein leerer Bestand in einem Zug (J02)

**#2266.**
Fast jeder Test legt einen frischen Bestand an, und bis 0.7.0 hieß das: 97
Wanderungen einzeln, darunter Umbauten ganzer Tafeln — unter dem
Race-Detektor **2,28 s** je Bestand. Seit Vermerk
[0189](entscheidungen/0189-nur-die-grundlinie.md) gibt es nur noch die
Grundlinie, und ein frischer Bestand entsteht in einem Zug: **0,23 s**
(`internal/storage/frischerbestand_bench_test.go`). Dieselben Stufen,
dieselben Grenzen des Race-Detektors, kein Lauf ausgelassen; entfallen sind
nur die Tests einzelner alter Wanderungen.

**Gemessen, nicht gerechnet** — `Dauer je Stufe` aus dem Protokoll jedes
Laufs, auf dieser Maschine (Mac17,3, macOS 27.0.1), mit warmem
Übersetzungszwischenspeicher, jeder Lauf allein:

| | Lauf | Format | Typisierung | Linting | **Tests** | Sicherheit | zusammen |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| vorher | 2026-10-12, `9d9e489b5` | 7 s | 54 s | 198 s | **1.586 s** | 14 s | 1.860 s |
| vorher | 2026-10-12, `f9c2b2668` | 8 s | 25 s | 88 s | **1.496 s** | 9 s | 1.626 s |
| vorher | 2026-10-12, `7e2cc1d71` | 8 s | 25 s | 72 s | **1.552 s** | 9 s | 1.666 s |
| nachher | 2026-10-12, `743b16369` | 8 s | 42 s | 231 s | **901 s** | 9 s | 1.191 s |
| nachher | 2026-10-12, `a34f88fc9` | 23 s | 43 s | 178 s | **883 s** | 10 s | 1.137 s |
| nachher | 2026-10-12, `52e1ee3e8` (Freigabe) | 8 s | 30 s | 86 s | **888 s** | 10 s | 1.022 s |

**Die Tests brauchen 41 bis 44 Prozent weniger**, 883 bis 901 statt 1.496
bis 1.586 Sekunden, und der ganze Lauf 1.022 bis 1.191 statt 1.626 bis
1.860 Sekunden. Dass das Linting im ersten Lauf danach länger brauchte, ist
der Zwischenspeicher von golangci-lint: der Lauf kam nach Änderungen in
vielen Paketen; im dritten, der Pipeline der Freigabe, lag es bei 86 s. **Was der Vergleich nicht isoliert**: mit der Grundlinie entfielen
die Tests einzelner alter Wanderungen (Vermerk 0189), und Treffer im
Übersetzungszwischenspeicher gehen in jede Zeile ein. Gemessen ist die Dauer,
die ein Lauf heute braucht, nicht der Anteil der Grundlinie daran.

### Das Ziel: zwanzig Minuten (PERF03)

**#2353,
Befund A-28.** Bis dahin gab es kein Ziel für die Dauer. **Seitdem sind es
zwanzig Minuten auf dieser Maschine** (Mac17,3, zehn Kerne, macOS 27.0.1),
und ein Lauf darüber **warnt** in seiner Schlusszeile (`MSE_CI_ZIEL`, Vorgabe
1200 s, `scripts/ci-beweis.sh`) — eine Warnung und kein Rot: ein langsamer
Lauf ist kein falscher.

**Und das Ziel wird heute nicht gehalten — ein ausdrücklicher Budgetbefund,
keine nachträglich passend gemachte Grenze.** Gemessen an den beiden Läufen
von Phase 2 von 0.9.0, je allein auf der Maschine:

| Lauf | Format | Typisierung | Linting | **Tests** | Sicherheit | zusammen |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 2026-10-14, `7b545aa3e` (M02) | 8 s | 64 s | 255 s | **1.010 s** | 16 s | **1.353 s** |
| 2026-10-14, `c5c4daa05` (Zweitblick) | 9 s | 67 s | 271 s | **1.012 s** | 17 s | **1.376 s** |

**Gegenüber der Freigabe von 0.8.0** (1.022 s, oben) kosten die Tests rund
120 s mehr — die neuen Läufe von Phase 1 und 2, darunter Flächenläufe mit
echten Prozessen —, und das Linting lag nach Änderungen in vielen Paketen bei
255 bis 271 statt 86 s; sein Zwischenspeicher trägt dort den größten Teil.
**Was das nicht sagt**: wie lange ein Lauf mit warmem Zwischenspeicher am
Ende von 0.9.0 braucht. Die Gesamtabnahme (Phase 9) misst es noch einmal und
entscheidet dann, ob die Teststufe geschnitten wird oder das Ziel begründet
fällt.

### Ein überholter Lauf hört auf (QA-07.AK-03)

**Ändert sich der Stand während eines vollen Laufs, gibt er keinen Beleg
mehr** (`stempel.sh`), und jede weitere Stufe kostete Minuten für nichts.
`noch_aktuell` fragt deshalb zwischen jeder Stufe, ob der Stand noch der vom
Anfang ist, und hört sonst mit `✗ überholt` auf. **Was das nicht ist**: ein
Abbruch **mitten** in einer Stufe — die laufende Stufe läuft zu Ende.
`scripts/hook.test.sh` hält, dass `--aktuell` einen geänderten Stand erkennt
und einen zurückgesetzten nicht; dass `ci.sh` zwischen den Stufen fragt, hält
es nur an der Gestalt.

**Was dabei nicht gestrichen wurde** (QA-07.AK-04): keine Stufe, kein Racelauf,
keine Abdeckungsschwelle. **Ein Cachetreffer ist kein übersprungener Test**:
Go gibt ein Ergebnis nur wieder, wenn Testprogramm, Umgebungsvariablen und
jede gelesene Datei gleich sind, und das Protokoll schreibt es mit dem Wort
„(cached)" — am 2026-10-01 traf das `internal/storage` im Lauf ohne Detektor; die geteilten Pakete laufen als übersetztes Testprogramm und damit ohne Cache.

### Ein Lauf hinterlässt einen Beweis, und er läuft allein

**Beides steht im Skript und nicht in der Umsicht des Aufrufers.**

**Das Protokoll schreibt der Lauf selbst**, nach `.ci-protokoll.log`, bevor
seine Ausgabe irgendwohin geht. Am 2026-09-14 ist ein voller Lauf rot geworden
und die rote Stufe war **weg**: der Aufrufer hatte die Ausgabe durch `tail -45`
geschickt. Zwanzig Minuten für die Zeile `✗ Pipeline rot`, und die einzige
Auskunft war ein zweiter voller Lauf.

**Der Aufrufer war dabei nur die Hälfte.** Das Skript wusste, welche Stufe
gefallen war — `fehl` lief in ihr — und hat es weggeworfen. Die Schlusszeile
nennt sie jetzt: `✗ Pipeline rot — rot war: test`, dazu den Pfad des
Protokolls.

**Und derselbe Fehler saß im Riegel selbst.** `scripts/vorpruefung.sh` fuhr
`npx vitest run | tail -5` — die Bilanz blieb stehen, der Name des roten
Laufs fiel hinten heraus. Am 2026-09-15 war einer von **3.533** rot, die
Stufe sagte „1 failed | 3532 passed", und drei Wiederholungen von Hand waren
grün, weil er nur unter Last auftrat. Seitdem geht die ganze Ausgabe in ein
Protokoll unter `mktemp`, und der rote Fall zeigt die Zeilen, die den Lauf
benennen.

**Unter `mktemp` und nicht unter einem festen Namen** — auch das ist bezahlt:
am 2026-09-13 räumten zwei gleichzeitige Läufe einander eine fest benannte
Protokolldatei weg und meldeten **beide grün**. Der Hinweis steht seitdem in
`.gitignore`.

> **Die Regel dahinter gilt für jede Stufe, die ein Werkzeug fährt**: was
> sie im roten Fall zeigt, muss den Fund benennen. Eine Bilanz ohne Namen
> kostet einen zweiten vollen Lauf — und wenn der Fund von der Last abhängt,
> auch den noch.

**Und zwei Läufe nebeneinander sind nicht zweimal derselbe Lauf.** Eine Sperre
weist den zweiten ab: der Race-Detektor teilt sich sonst die Maschine, und
`internal/app` läuft in seine Frist statt durch. Der zweite Lauf kostet nicht
nur seine eigenen zwanzig Minuten, er verdirbt die des ersten.
`MSE_CI_TROTZDEM=1` hebt die Sperre auf — wer sie aufhebt, weiß warum, und
sein Lauf schreibt in eine eigene Datei.

**Die Sperre nimmt der äußere Lauf, bevor er das Protokoll öffnet** — T04
(#2098).
Bis zum 2026-10-06 leerte `tee` die Datei des ersten Laufs, ehe der zweite
abgewiesen wurde; vom ersten blieb die Ablehnung des zweiten. Die Ablehnung
steht seitdem in `.ci-protokoll-abgewiesen.log`. **Sie gilt je Arbeitsbaum,
nicht je Maschine** — zwei Arbeitsbäume fahren zwei Läufe zugleich.

Gehalten wird das von `scripts/ciprotokoll.test.sh`, und zwar **ausgeführt
statt behauptet**: er fährt `ci.sh` mit verworfener Ausgabe und sieht nach, ob
das Protokoll trotzdem dasteht; er schneidet `kopf` und `fehl` aus den
Skripten heraus und fährt sie, samt Gegenprobe — eine **grüne** Stufe darf
nicht in der Liste stehen; und er legt eine belegte Sperre an und verlangt die
Abweisung. Was er nicht hält, steht in seinem Kopf: **dass eine Stufe das
Richtige prüft.** Er prüft den Beleg, nicht die Prüfung.

### Eine Lücke ist kein Beleg

**QA-01.AK-02** (#1733):
„Fehlende Pflichtfixtures, null Messfälle, abgebrochene Prozesse oder
unvollständige Artefakte können keinen gültigen grünen Pflichtbeleg erzeugen."
**Der Stempel ist der Pflichtbeleg dieses Laufs** — nach ihm lässt der Hook
pushen.

**Bis zum 2026-09-22 stempelte auch ein Lauf mit Lücke.** Fehlte
`frontend/node_modules`, übersprang `ci.sh` vier Stufen; antwortete die
Registry nicht, übersprang es `npm audit`. Die Schlusszeile sagte in beiden
Fällen „✓ Pipeline grün — N Prüfungen übersprungen", und der Stempel stand da.

**Jetzt endet ein solcher Lauf mit »◐ Pipeline unvollständig«, Rückgabe 1 und
ohne Stempel.** Wer bewusst so pushen will — die Registry ist gerade nicht zu
erreichen, und der Push kann nicht warten —, sagt es:
`MSE_CI_UEBERSPRUNGEN_ERLAUBT=1 make ci`. Dann stempelt der Lauf, und die
Schlusszeile nennt es „ausdrücklich erlaubt". Gehalten wird das von
`scripts/ciprotokoll.test.sh`, mit Gegenprobe: ein vollständiger Lauf stempelt
weiter.

### Eine Skriptprüfung läuft nur, wenn eine Hülle sie fährt

**C06** (#2231):
`skriptpruefungen` fährt die `scripts/*.test.sh`; eine `scripts/*.test.mjs`
läuft nur, wenn eine davon sie `node --test` übergibt. **Am 2026-10-10 tat das
für sieben keine** — `belegplan`, `dauerbetrieb`, `dauerteile`,
`huellenbeleg`, `antwortform` und beide der Zufallsmutation: geschrieben, rot
gesehen, grün gemacht und danach nie wieder gelaufen. Grün waren sie
trotzdem noch, gemessen beim Einhängen.

`scripts/skripttests.test.sh` hält jetzt jede `*.test.mjs` gegen die Hüllen
und nennt die, die keine fährt — mit Gegenprobe: ein gestrichener Eintrag
macht ihn rot. **Was er nicht hält**: dass die nennende Hülle die Datei
auch wirklich ausführt; er sucht den Pfad in Anführungszeichen, wie ihn jede
Hülle übergibt.

### `fm serve` läuft mit — nur, wo es da ist

**AI-02** (#1743),
Vorgabe des Auftraggebers vom 2026-10-04: im Rahmen von `make ci` wird **nur**
Apples Gerätemodell echt geprüft, und nur, falls `/usr/bin/fm` vorhanden ist.
Die Teststufe fährt dann `TestDasEchteGeraetemodell` — fünf Fragen über
`fm serve`, die Antwort in Form, eine erwartete Ablehnung, kein TCP-Port —
**ohne einen Beleg in den Baum zu schreiben** (`MSE_FMLAUF_OHNE_BELEG`).

**Was der Schritt nicht verlangt: Text in jedem Vorschlag.** Am 2026-10-04
gab das echte Modell für „Name, englisch" ein leeres `{}`, und die Pipeline
war rot, obwohl `fm serve` tat, was es zusagt. Ein leerer Wert ist ein
Ausgang des Modells, den die Anwendung kennt (#1671, Vermerk 0166); er
steht im Beleg als „(leer)". Rot bleibt, was den Dienst betrifft: keine
feste Form, ein falscher Ausgang, ein Fehler, ein TCP-Port. Die Güte misst
`fmguete_test.go`.

**Fehlt `fm`, ist das kein `weg`.** Ein Rechner vor macOS 27 hat es nicht;
das ist eine Eigenschaft der Maschine und nicht des Stands, und ein `weg`
hielte dort jeden Stempel an. Die Stufe meldet es gelb, und die Reichweite
nach dem Lauf sagt es noch einmal. **Ist `fm` da und das Modell nicht
bereit, ist der Lauf rot** (BLOCKED). Gehalten wird beides von
`scripts/geraetemodell.test.sh`, mit Gegenproben.

**Die kleinen Ollama-Modelle laufen nicht hier.** Sie sind in AI-02 gemessen
(`docs/belege/modellmessung.md`, `docs/belege/modellvergleich.md`) und danach
wieder gelöscht worden; ein Lauf, der sie bräuchte, hinge an 2,0 GB
Gewichten, die auf keinem frischen Rechner liegen.

### Ein Messlauf hinterlässt einen Beleg

**Die Läufe gegen fremde Pakete laufen nicht in `make ci`** (§3): sie brauchen
Pakete, Sets und Referenzbilder, die nicht im Repositorium liegen. Damit waren
sie die Prüfebene, deren Ausbleiben niemand sah. Vom 2026-09-14 bis zum
2026-09-20 übersprangen drei Kanonläufe, und `make fremdpakete` meldete
`ok` (#1656).
**Und `TestDerAbstandZuMSE2AmKanon` — der Lauf hinter den Kanonzahlen im
README — sprang in jedem Lauf über dieses Ziel**: das Makefile setzte zwei der
vier Eingaben nicht, die `MESSMETHODIK.md` für ihn verlangt.

**Seit QA-01 fährt `make fremdpakete` die Läufe über `internal/laufbeleg/lauf`**
und schreibt `docs/belege/fremdpakete.json`:

| Der Beleg nennt | Woher |
| --- | --- |
| Stand und Sauberkeit des Baums, **vor und nach** dem Lauf — der eigene Beleg zählt dabei nicht | `git rev-parse HEAD`, `git status --porcelain` |
| den **Inhalt** dessen, was er misst, als Hash | `git ls-files -s` über die Pfade des Plans und den Plan selbst |
| Werkzeuge und Plattform | `go env GOVERSION`, `sw_vers` |
| Anfang, Ende, Befehl, Rückgabewert | der Lauf selbst |
| jede Pflichteingabe mit **Inhaltshash** | `docs/belege/fremdpakete.plan.json` nennt sie; ohne sie wird **nicht** gemessen — BLOCKED |
| jeden erwarteten Fall mit **Ausführung** und **Ergebnis** getrennt | die Tests hinter dem Tag, **berechnet** aus den Dateien und nicht eingetragen |
| die Kennzahlen | jede Zeile `kennzahl name=wert`, die ein Lauf loggt |
| das Rohprotokoll als Artefakt mit Hash | `.laufbeleg/fremdpakete.jsonl`, nicht eingecheckt |

**Fünf Stände, und keiner steht für einen anderen**: PASS, FAIL, BLOCKED,
SKIPPED, NOT_RUN. Ein Sprung ist kein Bestehen, ein Treffer aus dem
Zwischenspeicher keine Messung (`-count=1`), und ein Absturz kein roter Fall,
sondern ein gescheiterter Lauf. **Ein bewusst roter Trendlauf** —
`TestDerAbstandZuMSE2` — steht im Plan als Trend: sein Ergebnis darf rot sein,
seine Ausführung nicht (AK-03). Bis dahin stand im Makefile ein `|| true`, und
das schluckte auch einen Übersetzungsfehler.

**`make belege` hält jeden Beleg gegen den heutigen Stand** (AK-04):
*aktuell*, *anderer Stand* — was der Lauf misst, hat heute einen anderen
Inhalt, der Arbeitsbaum weicht dort ab, eine Pflichteingabe hat einen anderen
Inhalt oder lässt sich nicht erheben, oder er lief an einem Baum mit
Änderungen — oder *überfällig*, wenn er älter als die Frist im Plan ist.
**Die Eingaben werden dafür neu gerechnet**, über Namen und Inhalt wie beim
Messen; das kostet so lange, wie sie groß sind. **Und womit gemessen wurde,
zählt mit**: die Go-Fassung, die Plattform und die **Schalter** des Plans —
Umgebungswerte wie `MSE_STICHPROBE`, die ändern, *was* gemessen wird. Jede
Umgebungsvariable, die ein Messlauf liest, ordnet der Plan ein — Eingabe,
Schalter oder ohne Wirkung, je mit Grund —, und
`internal/werkzeuge/belegschalter_test.go` hält die Liste vollständig. Bis zum Zweitblick über Phase
6.0 sah die Prüfung nur den Index: ein Beleg hieß »aktuell«, während eine
Datei im Baum oder das Verzeichnis der Referenzbilder schon anders war — und
der gezogene Stichprobenplan stand nicht einmal unter `misst`. Seitdem hält
`internal/werkzeuge/belegmisst_test.go`, dass jeder Pfad im Baum, den ein
Messlauf als Literal liest, dort steht — **und jede Datei, die `go:embed` in
den Bau nimmt**: eine geänderte Wanderung hätte sonst einen Beleg »aktuell«
gelassen, der am alten Schema gemessen war. **Verglichen wird der Inhalt und nicht der
Commit**: jeder PR wird per Squash gemergt, und danach liegt der gemessene
Commit in keiner Geschichte mehr — nach der Frist von `git gc` gäbe es ihn
nicht einmal. Hier stand ein Vergleich über `git diff <Commit> HEAD`, und er
wäre zwei Wochen nach dem ersten Merge an einem fehlenden Objekt gescheitert. Rückgabe 0 nur für einen grünen und
aktuellen Beleg. **Und jeder grüne `make ci` nennt denselben Stand** am Ende,
bei dem, was er nicht belegt — als Bericht und nicht als Sperre: eine
Prüfung, die mit der Uhr rot würde, hielte jeden Push an, ohne dass sich im
Baum etwas geändert hätte. **Er rechnet die Eingaben nicht**
(`-ohne-eingaben`) und sagt es: ein passender Beleg heißt dort höchstens
*ohne Eingaben* und nie *aktuell*.

**Was das nicht prüft**: ob eine Messung das Richtige misst. Der Beleg sagt,
dass sie lief, woran und womit — ihr Maß steht in `MESSMETHODIK.md`.

### Die Toolchain ist gepinnt — der Anlass

**Verlegt aus `AGENTS.md` §1 am 2026-10-06**, damit die Regeldatei unter ihrer
Obergrenze bleibt. Die Regel steht weiter dort.

**Der Anlass ist behoben, die Regel bleibt.** Seit dem 2026-09-15 steht die
Pinnung auf **Go 1.27.1**, golangci-lint damals auf **v2.13.2** — dessen v2.13.0
brachte „go1.27 support", nachgelesen im Änderungsprotokoll. Gehoben wurde
auf Zuruf und für `tailcat` (N3, #1456),
das ab v0.4.0 Go ≥ 1.27 verlangt.

**Hier stand, das sei nicht möglich** — in einem Vermerksentwurf, der
daraus einen Ablehnungsgrund für tailcat machte. Es war keiner: es gibt
eine Linterfassung für 1.27, und sie war einen Blick ins
Änderungsprotokoll entfernt. Wer eine Ablehnung auf eine Unmöglichkeit
stützt, prüft zuerst, ob sie eine ist (§10).

### Eine Pinnung gilt erst, wenn das laufende Werkzeug sie trägt

**QA-CI01** (#2106):
der Audit für 0.7.0 meldete zehn nicht gofumpt-konforme Dateien in
`internal/mcp`. Es waren keine — im Hauptarbeitsbaum lag gofumpt **v0.11.0**
vom 2026-09-14, gepinnt war seit #1784 **v0.12.0**, und mit dem steht der Baum
sauber da. Gemessen am 2026-10-06: mit v0.11.0 nennt die Stufe genau die zehn
Dateien, mit v0.12.0 keine.

Seitdem hält die Stufe `format` jedes Werkzeug auf dem Pfad mit
`go version -m` gegen seine `require`-Zeile in `tools/go.mod` und nennt beide
Zahlen (`scripts/ci-pinnung.sh`, Wächter `scripts/werkzeugpinnung.test.sh`).
Die Abhilfe ist `make tools`.

### Ein grüner Lauf hinterlässt seine Zeiten

**T08** (#2101):
bis zum 2026-10-06 löschte die Teststufe die Protokolle des geteilten
Racelaufs nach einem grünen Lauf — und mit ihnen Übersetzungszeit, Wandzeit
je Teil und die Dauer jedes Tests. Jetzt legt sie davor einen Dauerbeleg
unter `.ci-dauer/<zeit>-<commit>.tsv` ab (`teile_aufbewahren` in
`scripts/teilrennen.sh`): je Zeile `bau`, `teil` oder `test`, Paket, Teil,
Test und Sekunden. Die jüngsten zwanzig bleiben, eingecheckt wird keiner.
**Was er nicht hält**: die Pakete außerhalb der Teilung — für sie steht die
Gesamtzeit im Protokoll. **Und er ist keine Messung unter Ruhe**: er zählt,
was dieser Lauf brauchte, mit allem, was daneben lief.
