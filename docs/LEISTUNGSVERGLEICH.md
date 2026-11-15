# Leistungsvergleich mit MSE2

**QA-04** (#1735):
dieselben Aufgaben, dieselben Eingaben, offengelegte Bedingungen — und am Ende
**kein Gesamtfaktor**, sondern je Aufgabe, was besser, was schlechter und was
nur anders ist.

> **Warum es diese Datei gibt.** Die Zeiten, die bis dahin herumstanden — gut
> vierzehn Minuten für 25.148 MSE2-Referenzbilder, gut zwanzig für 4.265
> Karten des Kanonlaufs —, messen verschiedene Dinge: das eine ist ein
> Bildexport, das andere ein Test, der rastert, vergleicht und zerlegt. Ein
> Quotient daraus wäre eine Zahl ohne Frage.

Der Grundsatz ist derselbe wie in [`MESSMETHODIK.md`](MESSMETHODIK.md): die
Reichweite einer Messung ist die Reichweite ihrer Aussage (`AGENTS.md` §10).

---

## 1 Was gepaart wird — und was nicht

**Gepaart ist eine Aufgabe, wenn beide Programme sie ohne Klick ausführen
können, mit derselben Datei hinein und einem prüfbaren Ergebnis heraus.**
MSE2 hat kopflos genau drei Wege (`src/main.cpp`, Zeilen 188–200 und 217–268,
Fassung f9d9356 vom 2021-07-06): `--export-images`, `--export` und `--cli`,
eine Eingabezeile für MSE-Script mit einem **Rohmodus** für Programme. Alles
Gepaarte führt über diese drei; auf der Seite dieses Programms über die
gebauten Flächen `mse` und `mse-mcp` — dieselben Dienste, die das Fenster
ruft, ohne die Wails-Brücke dazwischen.

| Aufgabe (AK-01) | MSE2 | MSE4Mac | Was die Zahl einschließt | Was verschieden bleibt |
| --- | --- | --- | --- | --- |
| **Start** | `--cli --quiet --raw`, Eingabe `:quit` | `mse sets --json` an einem leeren, schon angelegten Bestand | Prozessstart bis zur ersten Antwort und Ende | MSE2 liest Einstellungen und Locale, MSE4Mac öffnet eine SQLite-Datenbank und sucht seine Paketwurzeln. **Das Fenster ist nicht gemessen** (§6) |
| **Öffnen** | `--cli <set> --quiet --raw`, Eingabe `:info`, `:quit` | `mse oeffnen <set> --json` in einen frischen Bestand | Lesen, Deuten, Bereitstellen; bei MSE4Mac **das Einfahren in den Bestand** samt Bildern | **Nicht dieselbe Semantik**: MSE2 hält das Set im Speicher und vergisst es beim Beenden; MSE4Mac schreibt es **einmal** in den Bestand und öffnet es danach aus der Datenbank (Zeile *Wiederöffnen*) |
| **Wiederöffnen** | wie *Öffnen* — MSE2 hat keinen anderen Weg | `mse karten <set> --json` an einem Bestand, der das Set schon trägt | Prozess, Datenbank, Kartenliste | Der Preis des Bestands und sein Ertrag stehen in zwei Zeilen statt in einer |
| **Speichern** | `--cli <set>` und `write_set_file(file:…)` (`src/script/functions/export.cpp`, Zeilen 429–438) | `mse sichern <set> <ziel>` | Bei MSE2 **einschließlich Öffnen** — es gibt kein offenes Set ohne Laden; bei MSE4Mac Prozessstart und Lesen aus dem Bestand | Darum steht bei MSE2 auch der Unterschied zu *Öffnen* da |
| **Navigation** — eine Karte zeigen | `write_image_file(set.cards[i], file:…)` in **einer laufenden** `--cli`-Sitzung (`export.cpp`, Zeilen 400–427) | `schreiben.karte.zeichnen` in **einer laufenden** `mse-mcp`-Sitzung | Je Karte von der geschriebenen Zeile bis zur Antwort: Plan, Rasterung, PNG, Datei | Kopflos gibt es kein „Bild ins Fenster", nur „Bild in eine Datei" — auf **beiden** Seiten. Die erste Karte einer Sitzung zählt als **kalt**, jede weitere als **warm**, dieselbe Karte noch einmal als **Wiederholung** |
| **Skriptauswertung** | ein Ausdruck in der laufenden `--cli`-Sitzung | `lesen.skript` in der laufenden `mse-mcp`-Sitzung | Deuten, Auswerten, Antwort | MSE2 wertet im Kontext des **Sets** aus, MSE4Mac im Kontext einer **Karte** — die Ausdrücke sind so gewählt, dass beides dasselbe ergibt |
| **Batchrendern** | `--export-images <set> <ziel>/{card.name}.png` | eine `mse-mcp`-Sitzung: `lesen.karten`, dann je Karte `schreiben.karte.zeichnen` — **der Weg des Fensters**, das die Karten auch einzeln schreibt (`internal/app/bildausfuhr.go`, `Bildnamen`) | Prozessstart bis zum letzten Bild | MSE2 zeichnet in einem Aufruf, MSE4Mac in N Rufen über eine Leitung |
| **Export** | `--export <vorlage> <set> <ziel>` | `mse vorlagen <set> ausfuehren <vorlage> <ziel>` | Prozessstart bis zur geschriebenen Datei | Gepaart sind `magic-simple-exporter` (wenig Skript) und `magic-namecheck` (viel Skript) — **die zwei, deren Ausgabe beide bytegleich schreiben**. `magic-text` und `magic-mws` schrieben in einer Erkundung am 2026-10-02 die Karten in anderer Reihenfolge als MSE2; **seit #2040 schreiben sie für Archenemy dieselben Bytes und sind gepaart** (`scripts/leistungsteile.mjs`, `ausfuhrvorlagen`) — gemessen ist das gepaarte Ergebnis noch nicht (Freigabelauf, #1672) |
| **Große Sets** | *Öffnen* und *Speichern* an `Black` (3.788 Karten, 295 MB) | dasselbe | dito | Eigenes Budget (§4) |

### Nur auf einer Seite — und deshalb getrennt

| Was | Warum nicht gepaart | Wo es steht |
| --- | --- | --- |
| **Der Start des Fensters**, Navigation und Bearbeitung **am Fenster** | Beide Programme brauchen dafür einen Klick, und ein Klickautomat für wxWidgets und einer für WKWebView messen zwei verschiedene Automaten mit | **offen** (§6) |
| **Bearbeiten eines Feldes** | MSE2 hat kopflos keinen Weg, eine Karte zu ändern: `--cli` wertet aus, schreibt aber kein Feld | **nicht gemessen**, auch für MSE4Mac nicht — ohne Gegenstück bliebe es eine Zahl ohne Vergleich (§5) |
| **Last auf den Dienst** — `mse-rest` mit 1, 4 und 16 gleichzeitigen Anfragen | MSE2 hat keinen Dienst | Ergebnis §4, Lauf `last` |
| **Der Bestand selbst** — Einfahren, Datenbank, Verlauf | MSE2 hat keinen | in der Zeile *Wiederöffnen* als Ertrag |

---

## 2 Bedingungen (AK-02)

**Was hier steht, schreibt der Lauf in seinen Beleg** — die Tabelle ist die
Lesehilfe, [`belege/leistungsvergleich.json`](belege/leistungsvergleich.json)
die Quelle.

| | |
| --- | --- |
| **MSE2** | Quelltext `f9d9356` vom 2021-07-06, gebaut nach [0026](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md) mit `CMAKE_BUILD_TYPE=Release` (`-O3 -DNDEBUG`), Apple Clang, wxWidgets **3.3.3** aus Homebrew (`osx_cocoau`), hunspell 1.7 |
| **Der macOS-Patch** | [`mse2-macos.patch`](mse2-macos.patch): **acht Zeilen**, eine fehlende Ausprägung von `GetDefaultMember::handle` für `unsigned long` — ohne sie linkt MSE2 auf arm64 nicht. **Keine Verhaltensänderung**; sonst keine Abweichung vom Quelltext. Es läuft **nativ**, ohne Wine und ohne Rosetta |
| **MSE4Mac** | der Stand, dessen Commit im Beleg steht: `mse`, `mse-mcp` und `mse-rest` mit schlichtem `go build` und der gepinnten Go-Fassung — ohne Race-Detektor, ohne die Schalter und das `-trimpath -ldflags "-s -w"` der Freigabe (`scripts/bauen.sh`, Zeile 94) |
| **0.5.0** | Tag `v0.5.0`, mit derselben Go-Fassung aus einem eigenen Arbeitsbaum gebaut — **der eigene Regressionsmaßstab** |
| **Hardware** | steht im Beleg: Chip, Kerne, Arbeitsspeicher |
| **Betriebssystem** | steht im Beleg: Fassung und Build |
| **Pakete** | `full-magic-pack/data` — **dieselben Dateien** für beide: MSE2 über den Verweis `build/data` und ein leeres eigenes Zuhause, MSE4Mac über einen **APFS-Klon** (`cp -c`) im Paketverzeichnis seines eigenen Zuhauses — ein Verweis genügt dort nicht, denn `filepath.WalkDir` folgt keinem (`internal/msepaket/paketwurzeln.go`) |
| **Das Zuhause von MSE2** | **`HOME` allein isoliert MSE2 nicht** — gemessen am 2026-10-02: mit leerem `HOME` lud es `magic.mse-game` aus `~/Library/Application Support/magicseteditor/data`, also aus `M15-Magic-Pack`. wxWidgets fragt auf macOS Core Foundation, und erst **`CFFIXED_USER_HOME`** lenkt das um. Darin muss `Library/Application Support/magicseteditor` liegen, sonst hängt MSE2 beim Beenden. **Dasselbe betrifft `scripts/mse2-referenz.sh`** (§6) |
| **Schriften** | dieselben TTF-Dateien: MSE2 über `~/Library/Fonts/mse2-referenz` (die Probe `TestDieReferenzschriftenReichen` hält sie gleich), MSE4Mac über `Magic - Fonts` im Pack. Der Beleg nennt einen Hash über beide |
| **Auflösung** | die **Ausgabe**: MSE2 mit seiner Vorgabe — `card_normal_export` ist wahr (`src/data/settings.cpp`, Zeile 129), also Zoom 1 ohne Drehung (`src/data/format/image.cpp`, Zeilen 27–50) —, MSE4Mac mit der Güte `medium`, „wie MSE2 zeichnet". Gleich ist sie, wenn je Karte Breite und Höhe gleich sind; das prüft der Lauf. Ein Bildschirm spielt kopflos keine Rolle |
| **Cachezustand** | §3 |
| **Wiederholungen** | im Beleg je Aufgabe; Reihenfolge §3 |

---

## 3 Was gemessen wird (AK-03)

| Maß | Woher | Bemerkung |
| --- | --- | --- |
| **Wandzeit** | `performance.now()` im Läufer, um den ganzen Prozess beziehungsweise je Befehl einer Sitzung | Millisekunden; `/usr/bin/time` gibt nur Hundertstel |
| **CPU** | `/usr/bin/time -l`: `user` + `sys` | zeigt, ob ein Programm wartet oder rechnet — und ob es mehrere Kerne nutzt (CPU > Wand) |
| **Speicher** | `/usr/bin/time -l`: `maximum resident set size` und `peak memory footprint`; im Dauerbetrieb der residente Stand alle zwei Sekunden über `ps` | |
| **Durchsatz** | Karten je Sekunde im Batchrendern, Anfragen je Sekunde im Lasttest | |
| **Fehler** | Rückgabewert ≠ 0, ein Rohsatz mit Status 2, ein Werkzeugruf mit `isError` | gezählt, nicht verschluckt |
| **Abstürze** | ein Signal statt eines Rückgabewerts, „Command terminated abnormally" | |
| **Datenintegrität** | Kartenzahl nach dem Öffnen; Kartenzahl und Bilderzahl in der gesicherten Datei; **die gesicherte Datei des einen Programms im anderen geöffnet**; **je Bild vollständig geprüft** (Signatur, Blöcke samt Prüfsumme, entpackte Länge gegen den Kopf — `scripts/pngpruefung.mjs`) und **Breite und Höhe je gepaarter Karte** zwischen MSE2 und MSE4Mac verglichen (`scripts/leistungsmasse.mjs`); der Kartenname je Stelle in beiden Programmen gleich | Die **Pixelnähe** zur Referenz misst der Kanonlauf (`MESSMETHODIK.md` §5b) und nicht dieser Lauf |
| **Dauerbetrieb** | eine Sitzung je Programm, die für eine feste Zeit Karten zeichnet: Zahl, Fehler, Speicherverlauf, Dauer am Anfang gegen Dauer am Ende | eigenes Budget |

**Statistik**: Median, p95, p99 und die Ränder nach der **Nächste-Rang-Regel**
ohne Interpolation — dieselbe wie im Kanonlauf (`MESSMETHODIK.md` §5). **Bei
kleinem n ist p99 das Maximum**, und so steht es da: wer aus fünf
Wiederholungen ein p99 liest, liest das Maximum.

### Kalt und warm

| Zustand | Was er heißt |
| --- | --- |
| **Prozesskalt** | ein frischer Prozess mit frischem Zuhause und — bei *Öffnen* — frischem Bestand. Kein Zwischenspeicher des Programms ist gefüllt |
| **Sitzungswarm** | derselbe Prozess, die zweite und jede weitere Karte: Vorlagen und Rahmenbilder sind geladen |
| **Wiederholung** | dieselbe Karte im selben Prozess noch einmal |
| **Seitenspeicher des Systems** | **nicht kalt zu machen**: `purge` braucht Verwalterrechte, und dieser Lauf hat sie nicht. Deshalb steht die **erste** Wiederholung jeder Aufgabe gesondert im Beleg — sie ist die einzige, die eine Datei womöglich noch von der Platte liest |

### Reihenfolge

**Je Wiederholung um eine Stelle gedreht** — ein lateinisches Quadrat über die
Programme: bei drei Programmen und einer durch drei teilbaren Zahl von
Wiederholungen steht jedes gleich oft an jeder Stelle. Der Beleg nennt je
Aufgabe und Programm den Median der Läufe **vorn** und den der Läufe
**dahinter**; liegen sie weit auseinander, misst die Reihenfolge mit.

---

## 4 Die Kartenauswahl und die Budgets (AK-04)

**Gezogen wird nicht neu**: die Karten kommen aus der Zufallsstichprobe von
QA-02, `docs/belege/kanonstichprobe.json`, und werden je Regel **gleichmäßig**
über die passenden Stellen verteilt — mit der Formel der festen Kohorte
(`scripts/leistungsauswahl.mjs`).

| Schicht | Satz | Vorlage | Karten | Warum |
| --- | --- | --- | ---: | --- |
| **Normal** | Multicolor | `m15-altered` | 12 | die gewöhnliche M15-Karte — 91 % des Kanons |
| **Sonderformat** | Archenemy, Planechase, Vanguard | ihre eigenen | je 6 | Überformat und eigene Stile |
| **Problemfall** | Multicolor | `m15-aftermath`, `m15-mainframe-planeswalker`, `m15-mainframe-dfc` | je bis zu 3 | **die drei Vorlagen, deren Abstand zu MSE2 über dem der gewöhnlichen Karte liegt** (`MESSMETHODIK.md` §5b) |

**Gezogen wurden 38**: 12 Normal, 18 Sonderformat, 8 Problemfall — von
`m15-aftermath` enthält die Stichprobe für Multicolor nur zwei Karten. Welche
es sind, Stelle, Name und Vorlage, steht im Beleg unter `auswahl`.

**Budgets**, jedes für sich:

| Teil | Budget |
| --- | --- |
| alle gepaarten Aufgaben an den kleinen Sätzen | Wiederholungen nach `WIEDERHOLUNGEN` (Vorgabe 6) |
| Navigation und Skript in der Sitzung | `NAVIGATION` Sitzungen je Satz und Programm (Vorgabe 2) |
| Batchrendern | `BILDER` Läufe je Programm (Vorgabe 1) — **eine** Wiederholung, weil MSE4Mac über MCP für 70 Karten gut anderthalb Minuten braucht; die Verteilung je Karte steht bei der Navigation |
| **große Sets** | `Black` mit `GROSS` Wiederholungen (Vorgabe 3) |
| **Dauerbetrieb** | `LANGLAUF` Sekunden je Programm (Vorgabe 120) |
| **Lastgrenze des Dienstes** | `LAST` Sekunden je Stufe, Stufen 1, 4 und 16 (Vorgabe 10) |

---

## 5 Ergebnis vom 2026-10-02

**Beleg**: [`belege/leistungsvergleich-2026-10-02.json`](belege/leistungsvergleich-2026-10-02.json),
Stand `f749bfa4` mit sauberem Baum, Lauf von 02:57 bis 03:13 Uhr — 159 Fälle,
366 Schritte, **null Befunde**. Daneben der erzeugte Bericht mit jeder Zeile:
[`belege/leistungsvergleich-2026-10-02.md`](belege/leistungsvergleich-2026-10-02.md).
**Der Lauf nach #2005 steht in §9** und unter dem gewohnten Namen
[`belege/leistungsvergleich.json`](belege/leistungsvergleich.json); dieser
hier ist umbenannt, damit seine Zahlen bei seinem Text bleiben. **Umbrochen
wurde der Beleg nach dem Lauf**, mit einer Zeile je Fall statt bis ins Letzte
eingerückt (`belegJson`, ab `bfe844af`); die Daten sind dieselben, als JSON
verglichen. Apple M5 mit
zehn Kernen und 32 GB, macOS 27.0 (26A428). **Die Last lag bei jedem Fall
zwischen 2,2 und 3,7**; gewartet werden musste nicht, und neben keinem Fall
lief eine Pipeline.

Die Zahlen sind **Mediane der Wandzeit**, in Klammern p95. Bei sechs
Wiederholungen ist p95 das Maximum, bei drei ebenso (§3). *0.5.0* ist der
Regressionsmaßstab, nicht der Vergleich.

### Ganze Prozesse

| Aufgabe | MSE2 | MSE4Mac | 0.5.0 | n |
| --- | ---: | ---: | ---: | ---: |
| **Start** | 60 ms (68) | **23 ms** (23) | 24 ms (27) | 6 |
| **Öffnen** Archenemy, 70 Karten | 490 ms (498) | **129 ms** (131) | 108 ms (111) | 6 |
| **Wiederöffnen** Archenemy | — wie Öffnen | **24 ms** (42) | 24 ms (26) | 6 |
| **Speichern** Archenemy | 660 ms (687) | **37 ms** (86) | 35 ms (83) | 6 |
| **Export** `magic-simple-exporter` | 483 ms (516) | **303 ms** (309) | 216 ms (249) | 6 |
| **Export** `magic-namecheck` | **881 ms** (887) | 6.717 ms (6.748) | 6.674 ms (6.686) | 6 |
| **Batchrendern** Archenemy, 70 Bilder | **3,1 s** — 22,7 Karten/s | 94,4 s — 0,74 Karten/s | 67,0 s — 1,04 Karten/s | 1 |
| **Öffnen** Black, 3.788 Karten | 5.627 ms (5.655) | **1.943 ms** (1.960) | 1.933 ms (1.941) | 3 |
| **Wiederöffnen** Black | — wie Öffnen | **72 ms** (73) | 72 ms (73) | 3 |
| **Speichern** Black | 11.885 ms (11.896) | **692 ms** (2.152) | 630 ms (2.135) | 3 |

**Nachgemessen am 2026-10-11 (E04, [`belege/e04-ausfuhr.md`](belege/e04-ausfuhr.md))**: `magic-namecheck` braucht über Archenemy **0,57 s** gegen MSE2s 0,87 s, bytegleich. Schnell wurde es schon mit #2040 am 2026-10-03 (`32c7321b8`, per `git bisect` bestimmt) — einen Tag nach dieser Tafel. Über Black bleibt es mit 19,6 gegen 6,3 s das Dreifache; warum, steht im Beleg.

**CPU und Speicher**, je Median: beim Öffnen von Black rechnet MSE2 5,6 s
bei **306 MB**, MSE4Mac 2,0 s bei **108 MB**. `magic-namecheck` kostet MSE4Mac
**12,4 s CPU** in 6,7 s — es rechnet auf mehreren Kernen —, MSE2 0,8 s. Das
Batchrendern kostet MSE4Mac 120 s CPU bei 623 MB, MSE2 3,0 s bei 239 MB.

### In einer laufenden Sitzung

Je Schritt von der geschriebenen Zeile bis zur Antwort: MSE2 über `--cli
--raw`, MSE4Mac über `mse-mcp` — beide **ohne** Prozessstart. Zwei Sitzungen
je Satz und Programm, die Auswahl aus §4.

| Schritt | MSE2 | MSE4Mac | 0.5.0 |
| --- | ---: | ---: | ---: |
| **bereit** — Set geladen, erste Karte zeichenbar | 514 ms (5.453 bei Multicolor) | **29 ms** (257) | 26 ms (246) |
| Normal, **kalt** — erste Karte der Sitzung | **66 ms** | 1.871 ms | 1.403 ms |
| Normal, **warm** — n = 22 | **43 ms** (53) | 1.797 ms (1.912) | 1.288 ms (1.343) |
| Normal, **dieselbe Karte noch einmal** | **41 ms** | 1.896 ms | 1.332 ms |
| Sonderformat, warm — n = 30 | **36 ms** (71) | 1.788 ms (2.105) | 1.217 ms (1.535) |
| Problemfall, warm — n = 16 | **51 ms** (105) | 1.902 ms (2.219) | 1.363 ms (1.614) |
| **Skript**, ein Ausdruck — n = 30 | **< 0,1 ms** | 458 bis 482 ms (≤ 513) | 329 bis 336 ms (≤ 361) |

**Warm und kalt unterscheiden sich bei MSE4Mac nicht**, und dieselbe Karte
ein zweites Mal kostet genauso viel: über `mse-mcp` beginnt **jeder** Ruf von
vorn. Das ist kein Zufall der Messung, sondern die Bauart — `app.Zeichennaht`
baut das Zeichenwerk je Ruf (§6). Ein Abzug mit `sample` während einer
Sitzung (2026-10-02) zeigte die Zeit im Rastern und in
`schriften.(*Suche).paketdateiUnter`, dem Gang durch das Paketverzeichnis;
**dass genau das den Abstand trägt, ist nicht gemessen** — gemessen ist es
seither in §8, mit einem CPU-Profil: drei Viertel der Rechenzeit eines Rufs
gingen in die Schriftsuche.

### Und der Weg des Fensters

**Das Fenster baut sein Zeichenwerk einmal** — und dafür gibt es eine
Messung im Testprozess, dieselbe `App`, die das Fenster trägt, ohne Wails
davor: die Wege `BenchmarkWegRastern`, `BenchmarkWegVorschauzerlegung` und
`BenchmarkWegVorschaukaltstart` (`internal/app/*_bench_intern_test.go`),
**am selben Stand und auf derselben Maschine gefahren**, um 03:17 Uhr bei Last
1,7 ([Beleg](belege/leistungsvergleich-fensterweg.log)):

| Weg im Fenster | MSE4Mac | zum Vergleich: MSE2, warm |
| --- | ---: | ---: |
| eine Karte rastern, Black, Bildschirmauflösung — n = 20 | **61 ms** | Normal 43 ms |
| eine Vorschau ganz, Vanguard — n = 20 | **43 ms** | Sonderformat 36 ms |
| der Kaltstart eines Stils, Vanguard | 442 bis 512 ms | Sitzung bereit 514 ms |

**Damit liegt das Fenster beim Zeichnen in der Nähe von MSE2**, und der
Abstand von vierzig zu eins gehört den kopflosen Flächen. **Zwei Vorbehalte**:
die Messung läuft im Testprozess und nicht im Fenster, und die Schriften des
Packs liegen in ihrem Zuhause **nicht** im Paketverzeichnis — ob die Zeit mit
ihnen anders wäre, ist nicht gemessen. Gepaart ist diese Zeile deshalb nicht;
sie ordnet ein.

### Unversehrtheit

| Prüfung | Ergebnis |
| --- | --- |
| **Kreuzprobe** — die gesicherte Datei des einen im anderen geöffnet | **8 von 8**: Archenemy mit 70 und Black mit 3.788 Karten, in beide Richtungen, für MSE4Mac und 0.5.0 |
| **Export** — SHA-256 der Ausgabe | für beide Vorlagen **bytegleich** in allen drei Programmen |
| **Bilder** — Signatur, Breite, Höhe | 70 von 70 je Programm, alle **435 × 621** |
| **Kartennamen** — je Stelle der Auswahl in beiden Programmen gegen die Setdatei | **38 von 38**, in allen drei Programmen |
| **Fehler** | MSE4Mac: **keiner**. MSE2: **zwei** Schritte der Navigation und **123** im Dauerbetrieb, alle an derselben Karte — Multicolor, Stelle 2779, `m15-mainframe-dfc`: „Variable not set: exists_in_package" in `loyalty_box_image`. Das Bild wird trotzdem geschrieben. Das Pack ist für MSE 2.5.8 geschrieben, der Quelltext ist älter |
| **Abstürze** | **keine**, in keinem der 159 Fälle |

### Dauerbetrieb — 120 Sekunden je Programm

| | MSE2 | MSE4Mac |
| --- | ---: | ---: |
| Schritte | 2.587 | 64 |
| Fehler | 123 (die Karte darüber) | 0 |
| Dauer, erstes Fünftel → letztes | 45,2 → 45,9 ms | 1.880 → 1.847 ms |
| **Resident, Anfang → Höchststand → Ende** | 269 MB → **9.099 MB** → 8.613 MB | 123 MB → **558 MB** → 558 MB |

**MSE2 wächst linear** — rund 84 MB je Sekunde, gut 3 MB je gezeichneter
Karte, ohne Plateau. **MSE4Mac steht nach zwölf Sekunden** bei 534 MB und
danach bei 558 MB. Was MSE2 dabei festhält, ist nicht untersucht; die Liste
der geschriebenen Dateinamen (`ExportInfo::exported_images`) erklärt
höchstens ein paar Byte je Schritt.

### Lastgrenze des Dienstes — nur MSE4Mac

`mse-rest`, `GET /sets/{Archenemy}/karten`, je Stufe zehn Sekunden mit einem
frischen Dienst; wer `429` bekommt, wartet, wie `Retry-After` es sagt.

| Gleichzeitig | angenommen | abgewiesen (429) | Antwortzeit, Median (p99) | erste Abweisung nach | resident |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 138 | 10 | 0,9 ms (5,8) | 119 ms | 52 MB |
| 4 | 138 | 40 | 2,9 ms (7,2) | 87 ms | 54 MB |
| 16 | 138 | 160 | 9,8 ms (25,6) | 79 ms | 54 MB |

**Die Grenze ist das Guthaben und nicht die Rechenleistung**: 120 Einheiten
je Principal, nachgefüllt mit 120 je Minute, ein Lesen kostet eine
(`mse grenzen`). 120 sofort und zwei je Sekunde ergeben die 138 in jeder
Stufe. Eine Abweisung kostet den Dienst um 1 ms; keine Anfrage scheiterte
anders, kein Dienst stürzte ab.

### Kalt, warm und die Reihenfolge

- **Die erste Wiederholung** fällt nur beim Speichern auf: Black braucht bei
  MSE4Mac beim ersten Mal 2.152 ms und danach 692 — das Verzeichnis wird zum
  ersten Mal geschrieben. Bei MSE2 gibt es diesen Unterschied nicht.
- **Vorn oder dahinter** macht keinen Unterschied, der über das Rauschen
  ginge: MSE2 startet vorn wie dahinter in 60 ms, öffnet in 490 gegen 484,
  `magic-namecheck` braucht 880 gegen 881. Die Reihenfolge misst nicht mit.
- **Ausgeglichen ist sie nur bei sechs und drei Wiederholungen.** Navigation
  und Skript (zwei Sitzungen) und das Batchrendern (eine) sind es nicht —
  dort stand 0.5.0 nie vorn. Bei den ganzen Prozessen, wo es ausgeglichen
  ist, zeigt sich kein Effekt; dass er bei den Sitzungen fehlt, ist daraus
  geschlossen und nicht gemessen.

### Bewertung — was besser ist, was schlechter, und was nur anders

**Kein Gesamtfaktor**, denn es gibt keinen: dieselbe Messung zeigt Abstände
von einem Achtzehntel bis zum Zweiundvierzigfachen, in beide Richtungen.

**Wo MSE4Mac sein Ziel erreicht**

- **Start, Öffnen und Speichern** sind um ein Mehrfaches schneller — Öffnen
  3,8-mal bei 70 Karten und 2,9-mal bei 3.788, Speichern 18- und 17-mal —,
  und ein großes Set belegt ein Drittel des Speichers.
- **Ein eingefahrenes Set ist in 24 bis 72 ms wieder offen.** MSE2 bezahlt das
  Laden in jeder Sitzung, bei Multicolor 5,4 s.
- **Dauerbetrieb**: der Speicher steht; MSE2 wächst ohne Grenze.
- **Kein Skriptfehler** an der Vorlage, an der MSE2 einen meldet.

**Was es dafür kostet — die Trade-offs**

- **Speichern schreibt ein Verzeichnis und kein Archiv**
  ([0018](entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)).
  Der größte Teil von MSE2s zwölf Sekunden bei Black ist das Packen; dafür ist
  seine Datei **eine** Datei. MSE2 liest das Verzeichnis (Kreuzprobe).
- **Öffnen heißt bei MSE4Mac Einfahren**: einmal in den Bestand, danach aus
  ihm. Die Zeile „Öffnen" vergleicht also ein einmaliges Einfahren mit einem
  Laden, das MSE2 jedes Mal braucht.

**Wo MSE4Mac deutlich langsamer ist**

- **Eine Karte zeichnen, kopflos**: 1,8 s gegen 43 ms in der laufenden
  Sitzung, 94 s gegen 3,1 s für 70 Karten. Die Ursache ist benannt (§6): die
  kopflosen Flächen bauen ihr Zeichenwerk je Ruf. **Der Weg des Fensters
  braucht 43 bis 61 ms** (oben) — der Renderer ist es nicht.
- **Ein Skriptausdruck** über `mse-mcp`: knapp eine halbe Sekunde gegen unter
  einer Zehntelmillisekunde — dieselbe Ursache.
- **Ein skriptlastiger Export**: `magic-namecheck` 7,6-mal so lang, bei
  gleichem Ergebnis bis aufs Byte. Hier baut niemand etwas neu; das ist der
  Interpreter.

**Gegen 0.5.0 — die Regression**

Zeichnen und Skript sind **langsamer geworden**: je Karte 1.797 gegen
1.288 ms (+40 %), 70 Bilder 94 gegen 67 s (+41 %), ein Ausdruck 480 gegen
330 ms (+45 %), `magic-simple-exporter` 303 gegen 216 ms (+40 %), Öffnen
129 gegen 108 ms (+19 %). **Gleich geblieben** sind Start, Wiederöffnen,
`magic-namecheck` und die großen Sets. **Woher die 40 % kommen, ist nicht
untersucht** — sie liegen auf dem Weg, den jeder kopflose Ruf neu geht.
**Seither untersucht und behoben** (§8): seit #1940 ging die Schriftsuche das
Pack dreimal durch statt zweimal.

### Was die Kriterien sagen

| Kriterium | Stand |
| --- | --- |
| **AK-01** gepaarter Plan | **erfüllt für die kopflosen Wege**: Start, Öffnen, Navigation als Karte in einer laufenden Sitzung, Speichern, Batchrendern, Export und große Sets sind gepaart, der Dienst getrennt; der Zeichenweg des Fensters ist im Testprozess eingeordnet. **Offen**: Start und Navigation **am Fenster** selbst und das **Bearbeiten** — MSE2 kann kopflos nichts ändern, und am Fenster misst diese Datei keines der beiden Programme |
| **AK-02** Bedingungen, 0.5.0 als Maßstab | **erfüllt** (§2, Beleg) |
| **AK-03** Verteilungen, Durchsatz, CPU, Speicher, Fehler, Abstürze, Unversehrtheit, Dauerbetrieb, kalt/warm, Reihenfolge | **erfüllt mit zwei Grenzen**: den Seitenspeicher des Systems kann der Lauf nicht leeren, und der Dauerbetrieb dauert zwei Minuten je Programm |
| **AK-04** Auswahl aus QA-02, Budgets, Bewertung ohne Gesamtfaktor | **erfüllt** |

## 6 Was die Zahlen nicht sagen

- **Nichts über das Fenster.** Start, Navigation und Bearbeitung am Fenster
  sind auf keiner Seite gemessen; kopflos steht für Navigation „eine Karte als
  Datei" — dieselbe Rasterung, ohne Bildschirm. **Und das Fenster zeichnet
  anders als `mse-mcp`**: es baut sein Zeichenwerk einmal
  (`internal/app/verdrahtungzeichnen.go`), die kopflosen Flächen bauten es
  am Stand dieses Laufs **je Ruf** neu (`app.Zeichennaht`). Was eine Karte
  über MCP kostete, war deshalb **nicht** was sie im Fenster kostet — §5.
  **Seit #2005 behalten sie es** je Prozess und Bestand (§8); die Zahlen in
  §5 beschreiben den Stand davor, bis ein neuer Lauf sie ersetzt.
- **Nichts über eine zweite Maschine** ([0007](entscheidungen/0007-kein-entfernter-lauf.md)).
- **Nichts über einen kalten Seitenspeicher** (§3).
- **Nichts über einen leeren Rechner.** Auf diesem Rechner prüfen mehrere
  Arbeitsbäume zugleich; der Läufer wartet vor jedem Fall auf das Ende jeder
  Pipeline und — mit Budget — auf eine Last unter der Kernzahl, und jeder Fall
  trägt die Last, bei der er begann (`last`). Wo sie hoch war, sagt der Beleg
  es; geglättet wird nichts. **Warum das nötig ist**, hat ein erster Lauf
  gezeigt, der in fremde Prüfläufe geriet: bei Last 42 bis 80 brauchte MSE4Mac
  je Karte bis zu 12,7 s statt 1,8 s und MSE2 bis zu 344 ms statt 43 ms — und
  zwar **nicht im selben Verhältnis**. Jener Lauf ist nicht als Beleg
  abgelegt.
- **Nichts über die Pixelnähe** — die misst der Kanonlauf.

### Ein Befund daneben: die Referenz von MSE2 war nicht isoliert

`scripts/mse2-referenz.sh` gab MSE2 ein eigenes `HOME`, „damit es Pakete nur
aus `build/data` lädt". **Das tat es nicht** (§2): MSE2 fragt Core Foundation,
und auf diesem Rechner lag unter `~/Library/Application Support/magicseteditor/data`
ein Verweis auf `M15-Magic-Pack`. Die Messung vom 2026-09-23 — „es änderte
hier nichts" — verglich damit zweimal denselben Zustand.

**Behoben in #2006**:
das Skript setzt `CFFIXED_USER_HOME` und legt darin
`Library/Application Support/magicseteditor` an (`scripts/mse2-referenz.test.sh`
hält beides). **Wiederholt am 2026-10-03, mit echter Isolation**: `Archenemy`,
`Black` und `Multicolor` — **6.866 von 6.866 Bildern bytegleich** mit der
Referenz von davor, bei gleichem Programm-, Schrift- und Set-Hash im Manifest.
Die Referenz steht damit; die Pakete, die sich zwischen den Packs
unterscheiden, unterscheiden sich für diese drei Sätze nicht im Bild.

---

## 7 Wie der Lauf gefahren wird

```bash
node scripts/leistungsvergleich.mjs                       # alle Teile, rund 25 Minuten
node scripts/leistungsvergleich.mjs start oeffnen ausfuhr # nur diese
```

| Angabe | Vorgabe | Wofür |
| --- | --- | --- |
| `MSE2_BINAER` | `~/mse2-quelle/build/magicseteditor` | das gebaute MSE2 |
| `MSE_FREMDPAKETE` | `/Users/david/full-magic-pack/data` | die Pakete beider Programme |
| `MSE_KANON` | `~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets` | die Sets, nur gelesen |
| `MSE2_SCHRIFTEN` | `~/Library/Fonts/mse2-referenz` | die Schriften für MSE2, nur gelesen |
| `ARBEIT` | `$TMPDIR/leistungsvergleich` | Bau, Zuhause, Bestände, Ausgaben |
| `WIEDERHOLUNGEN`, `NAVIGATION`, `BILDER`, `GROSS` | 6, 2, 1, 3 | Wiederholungen je Teil |
| `LANGLAUF`, `LANGLAUF_TEILNEHMER` | 120, `mse2,heute` | Dauerbetrieb |
| `LAST` | 10 | Sekunden je Laststufe |
| `RUHE_HOECHSTENS`, `LAST_HOECHSTENS`, `LAST_BUDGET` | 900, 60, 600 | wie lange auf eine Pipeline und auf Last gewartet wird |
| `BELEG` | `docs/belege/leistungsvergleich.json` | wohin der Beleg geht |

**Ein Vergleich misst nicht die Drosselung.** Gegen heute läuft er mit
abgeschaltetem Guthaben (`grenzen.kontingente aus` im Bestand `vorrat`): einmal
vor dem ersten Teil, zurückgestellt vor `last` — dort ist die Drossel der
Gegenstand — und am Ende, **auf den Stand, den der Lauf vorfand**, auch nach
einem Abbruch oder Signal (`scripts/leistungskontingent.mjs`). Im Lauf vom
2026-10-03 war es nur im Langlauf abgeschaltet; der Batchfall zeichnete 24 von
70 Bildern, und `zusammenfassen` stellte seine Zeit neben die vollständigen.
**Ein Fall mit Fehler oder Befund hat jetzt keine Zeit mehr im Vergleich**; die
Bilanz am Ende des Laufs (und im Beleg unter `bilanz`) zählt **Befunde** — auch
die aus den Fällen — und **Fehler** getrennt, und ein Lauf mit Befund endet mit
Exitcode 1.

**Fehlt eine Voraussetzung, misst er nicht** und nennt sie. Geprüft sind die
Rechnung, die Sitzungen, die Auswahl, der Bericht und die Aufrufe der
Teilnehmer in `scripts/leistungsvergleich.test.sh` — in der Pipeline, ohne
eines der beiden Programme zu fahren.

---

## 8 Folgebefund #2005: das Zeichenwerk je Ruf, und woher die 40 % kamen

#2005 aus
§5 und §6: jede Karte und jeder Ausdruck über `mse-mcp`, `mse-rest` und `mse`
kostete 0,5–1,8 s, und gegen 0.5.0 war es um rund 40 % langsamer geworden.
**Erst gemessen, dann geändert**, an zwei Stellen:

| Messung | Was | Wo |
| --- | --- | --- |
| **`BenchmarkWegKopflos`** | eine Karte und ein Ausdruck über die Naht, wie `schreiben.karte.zeichnen` und `lesen.skript` sie rufen — nach einem ersten Ruf —, die **erste Karte über einem frischen Werk**, derselbe Weg mit **gehaltenem** Werk (der des Fensters) und die Teile eines frischen Werks einzeln; je Durchlauf Wand- **und Rechenzeit**. Er nennt außerdem die Wurzeln der Schriftsuche | `internal/app/kopflosweg_bench_intern_test.go`, hinter `fremdpakete`; Befehl in [`RELEASE.md`](RELEASE.md) |
| **`mse-mcp`-Sitzungen** | dieselbe Gestalt wie die Navigation in §5: gebautes `mse` und `mse-mcp`, eigenes Zuhause, Archenemy eingefahren, acht Karten, dreimal dieselbe, acht Ausdrücke; dazu `mse vorlagen … ausfuehren` und `mse oeffnen` als ganze Prozesse | ein Treiber über `McpSitzung` aus `scripts/leistungssitzung.mjs`, nicht eingecheckt |

**Bedingungen**: Apple M5, zehn Kerne, 32 GB, macOS 27.0 (26A428), Go
1.27.1. `Archenemy.mse-set` aus den kanonischen Sets, gezeichnet die Karte an
Stelle 0; Pakete `full-magic-pack/data` **und** `Magic - Fonts` als APFS-Klon
im Paketverzeichnis eines eigenen Zuhauses — die Aufstellung aus §2.
Gemessene Stände: 0.5.0 (`f498fb4a`), vor und nach #1940 (`3cddf458`,
`bba338eb`), `b6a1d664` (der Stand vor dieser Änderung), `5efe6c59` (nur die
Schriftwurzeln entdoppelt) und `3693a988` (nach #2005 — gemessen am
Arbeitsbaum vor dem Eintragen, im Quelltext gleich).

**Der Rechner war nicht leer, und das bestimmt, wie die Zahlen zu lesen
sind.** Neben diesen Läufen prüften andere Arbeitsbäume; die Last über eine
Minute lag zwischen **9 und 130** und wechselte binnen Minuten. Gewartet wurde
vor jedem Lauf auf das Ende jeder Pipeline und höchstens eine Minute auf eine
Last unter 12. **Absolute Zeiten sind deshalb bis zum Dreifachen gestreckt**;
belastbar sind Läufe, die **abwechselnd zur selben Zeit** gefahren wurden, das
**Minimum** über Wiederholungen, die Rechenzeit neben der Wandzeit — und vor
allem, was gar keine Uhr braucht: welche Wurzeln die Schriftsuche geht.

### Wohin die Zeit ging

**Ein CPU-Profil über `BenchmarkWegKopflos`** am Stand `b6a1d664`
(`go test -cpuprofile`, 48 s Proben über die kalten und warmen Teile):

| Anteil | Wo |
| ---: | --- |
| **72,5 %** | `schriften.(*Suche).Suche` — davon **56,2 %** `paketdatei`, der Gang durch das Pack nach einer Schriftdatei, **12,8 %** `paketnamensindex`, der danach jede Schriftdatei öffnet, um ihren Namen zu lesen, und 3,2 % der Namensindex der Systemorte |
| 78,9 % | in `syscall` — `ReadDir`, `open`, `fdopendir`: der Gang selbst, quer zu den Zeilen darüber |
| 2,8 % · 2,9 % | Planbau und Rastern **ohne** die Schriftsuche darin |
| 1,7 % | der Stempel über die Paketdateien eines Stils (`stilstempel.go`) |
| 1,3 % | das Stilpaket lesen und seine Module einbinden (`Stilspeicher.ausPaket`) |
| unter 1 % | Skripte zerteilen (`msescript`) |

**Von den drei Verdächtigen aus dem Auftrag — Paketladen, Skriptumgebung,
Schriften — tragen die Schriften fast alles.** Eine Vorlage nennt eine Schrift
unter ihrem eigenen Namen — `Beleren Bold` —, und die Datei im Pack heißt
`beleren-bold_P1.01.ttf`. Die Suche geht dafür das Paketverzeichnis ab (36.380
Einträge unter `data`, 179 unter `Magic - Fonts`) und öffnet danach jede
Schriftdatei. Ein gehaltenes Werk merkt sich das Ergebnis; ein frisches fängt
von vorn an, und ein frisches war jeder Ruf.

**Die Teile eines frischen Werks** (Minimum über neun Durchläufe, ms):

| | 0.5.0 | vor #1940 | mit #1940 | `b6a1d664` | nur Wurzeln | nach #2005 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Paketwurzeln finden | 3 | 6 | 6 | 5 | 4 | 2 |
| erster Plan | 345 | 880 | 1.327 | 1.295 | 396 | 248 |
| erstes Zeichnen | 667 | 1.569 | 2.484 | 2.521 | 512 | 388 |
| dasselbe Zeichnen warm | 29 | 77 | 81 | 83 | 40 | 29 |

**Das Werk zu bauen kostet 3–5 ms**; teuer ist sein erster Gebrauch. Die
Spalten liefen unter verschiedener Last (die Zeile *warm* zeigt, wie sehr) und
sind **untereinander nicht** zu vergleichen — die Paare dafür stehen unten.

### Die Regression gegen 0.5.0

**Ursache: bba338eb** (#1940, FMT-02). Seither gibt jede Naht das
Paketverzeichnis als erste Paketwurzel mit, auch wenn darin unmittelbar kein
Paket liegt (`paketwurzelnFuer`) — richtig für Stile und Bilder, sonst sähe
niemand ein Paket, das später dort entsteht. **Das Zeichenwerk aber gibt der
Schriftsuche dasselbe Verzeichnis noch einmal als Schriftwurzel davor**, und
`data` liegt darunter. Gelesen aus dem Fehler einer Schrift, die es nicht
gibt — er nennt jede Wurzel, und jede ist ein Gang über ihren Baum:

| Stand | Wurzeln der Schriftsuche hinter den Systemorten | Gänge über das Pack je gesuchter Schrift |
| --- | --- | ---: |
| 0.5.0, vor #1940 | `pakete`, `pakete/full-magic-pack/data` | 2 |
| mit #1940, `b6a1d664` | `pakete`, `pakete`, `pakete/full-magic-pack/data` | 3 |
| nur Wurzeln, nach #2005 | `pakete` | 1 |

**Gemessen** — die erste Karte über einem frischen Werk, Rechenzeit je Ruf in
ms, je Zeile **abwechselnd zur selben Zeit** gefahren:

| Runde, Last | 0.5.0 | vor #1940 | mit #1940 | nur Wurzeln |
| --- | ---: | ---: | ---: | ---: |
| 1, 42–88 | 3.289 | 3.363 | **4.012** | **1.917** |
| 2, 55–88 | 2.922 | 2.924 | **4.079** | **1.968** |
| 3, 39–130 | 3.238 | 2.998 | **4.279** | **2.275** |

**#1940 macht die erste Karte um 19–43 % teurer, und das Entdoppeln drückt
sie um 30–42 % unter 0.5.0** — dort ging die Suche `data` ja auch schon
zweimal. Das ist der Zuwachs aus §5: je Karte über `mse-mcp` +40 %, je
Ausdruck +45 %, denn **jeder** Ruf war eine erste Karte. Die Rechenzeit steht
hier und nicht die Wandzeit: in Runde 3 war die Wandzeit bis zum Dreifachen
der Rechenzeit gestreckt.

**Behoben** in `schriften.(*Suche).MitPaketen` (5efe6c59): eine Wurzel, die
unter einer früheren liegt, wird nicht noch einmal gegangen. Ihr Stand zählt
weiter — ein Paket, das jemand unmittelbar in `data` legt, ändert die Zeit von
`data` und nicht die von `pakete` —, und eine Wurzel hinter einem Verweis
bleibt, weil `filepath.WalkDir` keinem folgt.

**Und zwei Zeilen aus §5 haben eine andere Ursache:**

- **`magic-simple-exporter`, +40 %** — nicht die Schriften: dieser Export misst
  keinen Text. Gemessen als ganzer Prozess in zwei Läufen, je abwechselnd,
  Minimum über acht, bei Last 60–99 — deshalb das Dreifache von §5: 0.5.0
  614 ms, vor FMT-06 627, **mit FMT-06 796**, vor #1940 773; und 0.5.0 610,
  vor #1940 775, mit #1940 785, `b6a1d664` 868, nach #2005 859.
  **Ursache ist 575b4443** (FMT-06, #1813): ein Paketverzeichnis wird
  seither als `os.Root` geöffnet und jede Datei darin begrenzt gelesen — damit
  ein Verweis nicht aus dem Paket hinausführt und eine Datei nicht beliebig
  viel Speicher belegt. Im Profil steht es als `os.(*Root).Open` und `openat`
  in `Vorlagenliste.spielangaben`, das das Spielpaket je Export neu liest.
  **Begründet und nicht geändert**: das ist eine Härtung, und sie gegen
  Zeit zu tauschen ist eine eigene Entscheidung. Die übrigen gut 10 % zwischen
  `bba338eb` und `b6a1d664` sind **nicht weiter zerlegt**.
- **Öffnen, +19 %** — das Anlegen des Bestands: `mse sets --json` an einem
  **frischen** Bestand, also Anlegen und Migrieren allein, kostet 0.5.0 92 ms
  und `b6a1d664` 114 ms (Minimum über sechs, Last 14), Öffnen 124 gegen
  151 ms — derselbe Abstand. Seit 0.5.0 sind 19 Migrationen dazugekommen
  (77 → 96). **Begründet**: das zahlt ein Bestand einmal; *Wiederöffnen* ist
  gleich geblieben.

### Die Änderung

**Die Flächen ohne Fenster behalten ihr Zeichenwerk** — je Prozess und
Bestand, höchstens vier (`internal/app/zeichenvorrat.go`). Neu gebaut wird es,
wenn sich ändert, woraus es gebaut ist: Bilderverzeichnis, Schriftwurzel, die
Paketwurzeln und deren Änderungszeit. Den Rest sieht das Werk selbst:

| Änderung | Wer sie sieht | Wann | Belegt von |
| --- | --- | --- | --- |
| Stil oder Spiel eines Sets — aus derselben Sitzung oder aus einem anderen Prozess | der Stilspeicher, der den Setkopf **je Ruf** liest | beim nächsten Bild | `TestEinStilwechselWirktSofort`, `TestEinStilwechselImBestandGiltOhneVergessen`, `TestEinSpielwechselGiltFuerDenKartenstil`; der Lauf unten |
| eine Datei in einem Paket, auch ein Bild | der Stempel des Stilspeichers; er meldet den Wechsel, und das Werk leert Paketbilder und Ebenen | nach höchstens einer Sekunde | `TestEinGeaenderterRahmenWirktOhneNeustart`, `TestEinGeaendertesPaketMeldetSich` |
| ein Paket kommt in eine Wurzel oder geht, eine Wurzel entsteht | der Vorrat, an der Änderungszeit der Wurzeln | beim nächsten Ruf | `TestEinNeuesPaketGiltBeimNaechstenRuf` |
| ein Ruf aus einem Profil, das das Set nicht führt | niemand — er hinterlässt nichts | — | `TestEinFehlendesSetWirdNichtGemerkt` |
| nebenläufige Rufe | dieselben Speicher, jeder mit Sperre | — | `TestNebenlaeufigeRufeTeilenDasWerkOhneWettlauf`, unter `-race` |

**Die Gegenproben sind gefahren**: ohne die Meldung des Paketwechsels kommt
nach dem neuen Rahmen dasselbe Bild heraus, ohne die Änderungszeit der Wurzeln
bleibt das Werk über ein neues Paket stehen, und ohne die Sperre im Vorrat
meldet der Detektor Wettläufe.

**Was ein behaltenes Werk nicht sieht**, und das gilt im Fenster genauso:
eine Schrift, die jemand in ein **vorhandenes** Paket legt
(`schriften.paketnamensindex`), und ein Bild aus einem Paket, das kein Stil
eines gezeichneten Sets las. Ein neuer Prozess sieht beides.

### Vorher und nachher

**`mse-mcp`-Sitzungen**, je Stand zwei, abwechselnd, Last 17–28; ms, die
zweite Sitzung nach dem Schrägstrich:

| Schritt | 0.5.0 | `b6a1d664` | nach #2005 |
| --- | ---: | ---: | ---: |
| erste Karte der Sitzung | 1.454 / 2.059 | 1.596 / 2.211 | 1.506 / 1.248 |
| weitere Karten, Median aus sieben | 993 / 2.027 | 1.414 / 2.445 | **183 / 152** |
| dieselbe Karte noch einmal | 915–1.037 / 2.671–2.898 | 1.377–1.559 / 3.168–4.212 | **132–137 / 108–142** |
| ein Ausdruck, Median aus acht | 350 / 949 | 509 / 1.542 | **36 / 27** |
| resident nach der Sitzung | 493 / 493 MB | 502 / 553 MB | 482 / 482 MB |

**Und im Testprozess**, `BenchmarkWegKopflos` nach #2005 bei Last 12: eine
Karte über die Naht nach einem ersten Ruf **65 ms** Wandzeit (79 ms
Rechenzeit), ein Ausdruck **16 ms**, dieselbe Karte mit gehaltenem Werk
61 ms — und `b6a1d664` daneben 2.845 und 1.200 ms.

**Damit liegt eine Karte über `mse-mcp` in der Größenordnung des Fensters**
(§5: 43–61 ms) und nicht mehr beim Vierzigfachen. Der Speicher wächst nicht:
was vorher je Ruf neu entstand und wegfiel, bleibt jetzt einmal stehen.
**Die erste Karte einer Sitzung bleibt kalt** — sie baut, was das Werk danach
behält; schneller als vorher ist sie um den Gang über das Pack, den die
Schriftsuche spart.

**Der Lauf zum Stilwechsel**, `Black.mse-set`, eine Karte ohne eigenen Stil,
Bild als SHA-256, ms bei Last 45–52:

| Schritt | `b6a1d664` | nach #2005 | Bild |
| --- | ---: | ---: | --- |
| Stil des Sets `magic-m15-altered` | 2.048 | 130 | `a3ef9015…` |
| dasselbe noch einmal | 3.626 | 114 | `a3ef9015…` |
| `schreiben.stil` → `magic-m15-mainframe-planeswalker`, **in derselben Sitzung** | 4.292 | 389 | `27ec7078…` |
| `mse stile … wechseln` → `magic-m15-altered`, **aus einem anderen Prozess** | 4.303 | 232 | `a3ef9015…` |

**Beide Wechsel wirken beim nächsten Bild**, wie vorher. Dass das Werk
dazwischen dasselbe bleibt, sieht man einem Prozess von außen nicht an — das
belegt `TestEinStilwechselWirktSofort`.

### Was diese Zahlen nicht sagen

- **Nicht, was der volle Leistungsvergleich zeigt.** Sie sind gezielt
  gefahren, auf einem Rechner unter fremder Last, und ersetzen keine Zeile aus
  §5 — das tut der neue Lauf in §9.
- **Nichts über `mse-rest` unter Last** und den Speicher aus QA-08
  (`LASTMESSUNG.md`, bis 3,7 GiB beim Rendern): gemessen ist hier eine
  `mse-mcp`-Sitzung, nicht der Lastlauf.
- **Nichts darüber, ob 1,7 % Stempel je erster Karte nötig sind** — sie
  gehören zum Gedächtnis des Stilspeichers und sind nicht angefasst.

## 9 Lauf vom 2026-10-03 — nach #2005

**Beleg**: [`belege/leistungsvergleich.json`](belege/leistungsvergleich.json)
und der Bericht [`belege/leistungsvergleich.md`](belege/leistungsvergleich.md),
Stand `88426034b` (`main` nach #2034 und die zwei Korrekturen unten) mit
sauberem Baum, Lauf ab 10:26 Uhr — 159 Fälle, 366 Schritte, **null Befunde**.
Dieselbe Maschine wie in §5, frisch gestartet, **Last 2,7 bis 5,6**, kein
Agent und keine Pipeline daneben. **Ein erster Lauf am 2026-10-02 abends
wurde abgebrochen**: parallele Werkstücke trugen die Last auf 45 bei zehn
Kernen, und ein Beleg unter dieser Last belegte nichts.

### In einer laufenden Sitzung

Median, in Klammern p95; *vorher* ist §5.

| Schritt | MSE2 | MSE4Mac | vorher | 0.5.0 |
| --- | ---: | ---: | ---: | ---: |
| **bereit** | 498 ms | 27 ms | 29 ms | 26 ms |
| Normal, **kalt** | 61 ms | 789 ms | 1.871 ms | 1.281 ms |
| Normal, **warm** — n = 22 | **43 ms** (53) | **105 ms** (125) | 1.797 ms | 1.195 ms |
| Normal, dieselbe Karte noch einmal | 38 ms | 163 ms | 1.896 ms | 1.244 ms |
| Sonderformat, warm — n = 30 | 35 ms (70) | **78 ms** (159) | 1.788 ms | 1.157 ms |
| Problemfall, warm — n = 16 | 45 ms (100) | 152 ms (335) | 1.902 ms | 1.246 ms |
| **Skript**, ein Ausdruck warm | < 0,1 ms | **16 ms** (17) | 458–482 ms | 312 ms |

**Damit liegt eine Karte über `mse-mcp` in der Größenordnung des Fensters**
(43–61 ms im Testprozess, §5) — AK-1 von
#2005 —, und
gegen 0.5.0 ist aus 40 % langsamer elfmal schneller geworden. **MSE2 bleibt
bei einer warmen Karte rund zweieinhalbmal schneller**; die erste Karte einer
Sitzung baut, was das Werk danach behält, und braucht 0,8 s.

### Ganze Prozesse — was sich bewegt hat

| Aufgabe | MSE2 | MSE4Mac | vorher | 0.5.0 |
| --- | ---: | ---: | ---: | ---: |
| **Batchrendern** Archenemy, 70 Bilder (MSE4Mac Karte für Karte über `mse-mcp`) | 3,0 s | ~~2,4 s~~ **widerlegt** — der Lauf zeichnete für MSE4Mac 24 der 70 Bilder (Z1 des Zweitblicks über die QA-Arbeit); der Messcode prüft das seit #2040, die Zahl ist noch nicht wiederholt (Freigabelauf, #1672) | 94,4 s | 61,9 s |
| Export `magic-simple-exporter` | 464 ms | 297 ms | 303 ms | 211 ms |
| Export `magic-namecheck` | 860 ms | 6.523 ms — **seit #2040 0,9 s, seit E04 0,57 s** ([`belege/e04-ausfuhr.md`](belege/e04-ausfuhr.md)) | 6.717 ms | 6.498 ms |

**Das Batchrendern war der größte Sprung** — mit einer Einschränkung, die erst
der Zweitblick fand: gemessen waren 24 der 70 Bilder (Z1). Die Ursache des
Sprungs steht (dieselbe Schriftsuche lief je Bild über das ganze Pack, §8), die
Zahl nicht. Der Export bleibt, wo er war — die 40 % aus
FMT-06 sind begründet (§8). Start, Öffnen, Wiederöffnen und Speichern liegen
innerhalb weniger Prozent der Zahlen aus §5.

### Dauerbetrieb und Lastgrenze

| | MSE2 | MSE4Mac |
| --- | ---: | ---: |
| Schritte in 120 s | ~~2.663~~ **widerlegt** — gescheiterte MSE2-Rufe zählten doppelt und ihre Dauer ging in die Verteilung ein (Z6); seit #2040 einmal gezählt, noch nicht wiederholt | **877** (vorher 64) |
| Fehler | 126 (die Karte darüber) | 0 |
| Dauer, erstes Fünftel → letztes | 44,5 → 45,0 ms | 114,0 → 114,1 ms |
| **Resident, Anfang → Höchststand → Ende** | 269 MB → 9.350 MB → 8.630 MB | 125 MB → 540 MB → 540 MB |

**MSE4Mac steht nach zwölf Sekunden bei 532 MB** und wächst danach in 108 s
um 8 MB — rund 9 KB je gezeichneter Karte. Ob das ein Leck ist oder ein
Zwischenspeicher, der sich füllt, ist nicht untersucht; bei 877 Karten fällt
es nicht ins Gewicht, über Stunden würde es. **Die Lastgrenze** gab dieselben
Zahlen wie §5: 138 angenommen je Stufe, 10, 40 und 160 abgewiesen.

### Was am Messwerkzeug geändert wurde

**Der Langlauf misst die Maschine, nicht die Drossel.** Im ersten Lauf nach
#2005 griff nach 71 Karten das Guthaben von `mse-mcp` — Zeichnen kostet seit
#2003 fünf Einheiten und ist seit #2005 schnell —, und die Schleife fragte
ohne Pause weiter: 931.119 Abweisungen in 120 s. Jetzt schaltet sie die
Kontingente **nur für den Langlauf** ab (`grenzen.kontingente aus`, wie
`lastlauf.mjs`) und danach wieder an; die Lastgrenze läuft auf demselben
Bestand und misst die Drossel — ein zweiter Lauf, in dem sie für den ganzen
Bestand aus waren, zeigte dort keine einzige 429. Und **eine Abweisung wird
abgewartet**, so lange, wie sie nennt (`wartezeitAus`,
`scripts/leistungsdauer.test.mjs`). 0.5.0 hat den Schalter nicht; sein
Langlauf läuft per Vorgabe nicht mit.
