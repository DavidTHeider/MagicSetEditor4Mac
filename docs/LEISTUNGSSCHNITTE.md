# Leistungsschnitte: der Messvertrag

**PERF-01** (#2168):
bevor etwas schneller wird, steht fest, **was** gemessen wird, **unter welchen
Bedingungen**, und woran ein Lauf erkennbar nicht zählt. Der Läufer ist
`scripts/leistungsschnitte.mjs`; sein Beleg liegt in
[`belege/leistungsschnitte.json`](belege/leistungsschnitte.json).

> **Warum es diese Datei gibt.** Die Zahlen davor vermischten Transport,
> Skriptumgebung und Ausfuhr; ein „Zerteilen" maß Treffer im Baumspeicher; und
> die Ruhe des Rechners hielt nicht über die Abschnitte — ein Lauf protokollierte
> fremde Last, hielt aber nicht an. Der Leistungsvergleich mit MSE2
> ([`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md)) misst ganze Aufgaben
> zweier Programme; diese Datei misst **die Teile dieses einen**.

---

## 1 Die Schnitte (AK-01, AK-02)

Jeder Schnitt ist eine Messung, die es im Repositorium gibt, und läuft als
**eigener Prozess** aus einem **einmal gebauten** Testprogramm; dessen SHA-256
steht im Beleg, ebenso die der Eingänge.

| Schnitt | Was er misst | Messung | Eingang |
| --- | --- | --- | --- |
| `ast-frisch` | Zerteilen mit **frischem Neuaufbau** des Baums — jede Quelle einmal | `msescript`, `BenchmarkZerteileKaltUndWarm/kalt` | ein künstlicher Ausdruck, 20 Zeilen |
| `ast-speicher` | Zerteilen, **nur Treffer** im Baumspeicher — ausdrücklich so benannt, keine Zerteilzeit | dito, `/warm` | dito |
| `ebene` | eine Ebene auf ihre Größe abtasten, kalt und aus dem Ebenenspeicher | `render`, `BenchmarkEbeneKaltUndWarm` | künstlich |
| `einfahren` | Kartenlesen: ein Set in den Bestand | `app`, `BenchmarkWegEinfahren` (klein, groß) | `Vanguard`, `Black` |
| `naht-skript` | **öffentliche Naht**: ein Ausdruck je Ruf, wie `lesen.skript` | `app`, `BenchmarkWegKopflos/Skript_je_Ruf` | `Archenemy` über dem Pack |
| `naht-karte` | öffentliche Naht: eine Karte je Ruf, wie `schreiben.karte.zeichnen` | `…/Karte_je_Ruf` | dito |
| `werk-frisch` | **ganze Rechnung** mit frischem Werk: Umgebung samt Abbildung, Plan, Rasterung | `…/Karte,_frisches_Werk` | dito |
| `werk-gehalten` | dieselbe Karte über das gehaltene Werk — der Weg des Fensters | `…/Karte,_Werk_gehalten` | dito |
| `ausfuhr` | **Ausfuhr** über `magic-simple-exporter`, 70 Karten | `app`, `BenchmarkWegAusfuhr` (neu mit PERF-01) | dito |
| `rastern` | Rasterung einer Karte im Fensterweg | `app`, `BenchmarkWegRastern` | `Black` |
| `paare` | dieselbe Karte vor, sofort und nach fünf anderen (PERF-04, §6) | `app`, `BenchmarkKartenpaare` | `Archenemy` |
| `mcp` | **echter Transport**: `mse-mcp` über stdio, drei Ausdrücke je 100 Rufe | `McpSitzung` aus `leistungssitzung.mjs` | `Archenemy` |

**Kalter Zerteiler und Speichertreffer stehen getrennt**, und ein Lauf, der nur
Treffer zählt, heißt so: `ast-speicher` misst rund 7 ns — die Zeit eines
Nachschlagens, nicht die des Zerteilens. **Kleine Schnitte laufen lange genug**:
zwei Sekunden je Wiederholung (`DAUER`), fünf Wiederholungen; die Schnitte über
dem Pack fünf Durchläufe je Wiederholung.

**Die Eingänge** sind das `data` von `Full-Magic-Pack` (34.683 Dateien; der
Abdruck geht über Namen und Größen) und die kanonischen Sets
(`Archenemy.mse-set` mit SHA-256 im Beleg). Sie liegen außerhalb des
Repositoriums (§11) und werden nur gelesen. **Das Pack wird einmal je Lauf**
als APFS-Klon in ein eigenes Zuhause gelegt (`MSE_PAKETZUHAUSE`): ein Klon je
Prozess weckte `fseventsd` mit 250–280 % CPU **während** der gemessenen
Schleife — gemessen im ersten vollen Lauf am 2026-10-07.

## 2 Die Umgebung und die Pause (AK-03)

**Vor und nach jedem Schnitt** stehen im Beleg die Last (`loadavg`) und die
fünf größten **fremden** Prozesse samt ihrer Summe — fremd heißt: außerhalb
des Prozessbaums des Läufers.

| Grenze | Vorgabe | Schalter |
| --- | --- | --- |
| Last (1 Minute) | Kerne / 2 — auf diesem Rechner 5 | `LASTGRENZE` |
| fremde Prozesse zusammen | 250 % CPU | `FREMDGRENZE` |
| längste Wartezeit je Schnitt | 120 s, alle 10 s ein Blick | `PAUSE` |

**Liegt die Umgebung darüber, wartet der Schnitt.** Fällt sie nicht unter die
Grenze, gilt er als **nicht gemessen** und der ganze Lauf als
**unvollständig**; der Grund steht dabei.

**Und während des Schnitts gilt die Grenze für fremde Prozesse weiter** —
Befund 4 des Zweitblicks über Phase I. Alle fünf Sekunden und am Ende nimmt
der Läufer eine Probe (`beobachte`), und liegt eine davon über 250 %, gilt
der Schnitt als **gestört, nicht gemessen** (`gestoert`); im Beleg stehen
Zahl der Proben, verpasste Proben und höchste fremde Last. **Eine verpasste
Probe gilt nicht als ruhig** (Befund 3 des Nachlaufs): scheitert ein Blick,
ist nicht belegt, dass die Grenze hielt, und der Schnitt ist ebenso nicht
gemessen; „nachher" steht dann leer statt mit einer älteren Probe. **Die Last (`loadavg`) zählt nur
vorher**: sie enthält die eigene Arbeit, und ein Schnitt, der die Kerne
belegt, höbe sie selbst über die Grenze — im offiziellen Lauf stand sie
danach bei bis zu 5,44.

**Der offizielle Beleg ist älter als diese Regel.** Er hat keine Proben
während der Schnitte, nur je eine danach; die liegen alle unter der Grenze
(höchstens 243,7 %, nach `ast-speicher`). Dass zwischen Anfang und Ende
keine fremde Last über 250 % lag, belegt er **nicht**. **Fremde Prozesse hält der Lauf nicht
an.** Die Grenze für fremde Prozesse liegt bei 250 %, weil auf diesem Rechner
`StorageManagementService` dauernd ein bis zwei Kerne belegt (gemessen am
2026-10-07: 86–91 %, dazu `ApplicationsStorageExtension` um 20–80 %).

**Beide Fälle sind gefahren** (`leistungsschnittteile.test.mjs`, dazu am
Rechner):

- **Normalfall**: unter den Grenzen misst ein Schnitt sofort; über ihnen
  wartete er im offiziellen Lauf bis zu 40 s, in einem früheren 110 s, bis
  `fseventsd` nach dem Klon ruhig war.
- **Gegenfall** (`ERZWINGE_PAUSE=1 PAUSE=10`): die Grenze liegt unter jeder
  Last, der Schnitt wartet zehn Sekunden, gilt als „pausiert, nicht gemessen",
  und der Lauf endet mit 1 und „unvollständig".

**Und ein unbekannter Schnittname bricht ab**: ein ungeteiltes `$TEIL` in zsh
gab dem Läufer am 2026-10-07 einen einzigen Namen aus vier Wörtern, er maß
nichts und meldete Erfolg.

## 3 Speicher — vier Zahlen, und keine ist die andere (AK-04)

| Im Beleg | Was es ist | Was es nicht ist |
| --- | --- | --- |
| `B/op`, `allocs/op` | Go-Zuteilung **je Durchlauf** (`-benchmem`), über alle Goroutinen des Testprozesses | kein lebender Heap und keine Spitze |
| RSS | höchster residenter Stand des **ganzen** Testprozesses, Aufbau eingeschlossen (`/usr/bin/time -l`) | nicht der Schnitt allein |
| Spitze | `peak memory footprint` desselben Prozesses — was macOS ihm anrechnet, komprimierter Speicher eingeschlossen | dito |
| — | **nicht erhoben**: der lebende Go-Heap, der Adressraum von SQLite als WebAssembly, der Speicher des Fensters | |

## 4 Echter Transport (AK-05)

`mse-mcp` über stdio, gestartet in einem eigenen Zuhause mit eigenem Bestand,
**Guthaben abgeschaltet** — sonst misst der Lauf ab dem hundertsten Ruf die
Drossel (`leistungskontingent.mjs`; so geschehen im ersten vollen Lauf: 8 von
100 Rufen scheiterten, die übrigen Ausdrücke alle). Gemessen werden dieselben
drei Ausdrücke wie im Leistungsvergleich — `1 + 1`, `to_upper("abc")`,
`card.name` an der vierten Karte —, je 100-mal **nach** zehn Runden Aufwärmen;
Start und Kartenliste liegen davor (`bereitMs`). Die Zeit reicht von der
geschriebenen Zeile bis zur gelesenen und zerlegten Antwort
(`McpSitzung.werkzeug`). **Eine Differenz zweier Mediane ist keine Zeit für
Transport oder JSON** — dafür fehlt der gemeinsame Ruf.

## 5 Ergebnis

**Offizieller Lauf am 2026-10-07**: Commit `a13a64b70`, **sauberer Baum**,
Apple M5 (10 Kerne, 32 GB), macOS 27.0.1 (26A434), go1.27.1, **vollständig** —
kein Schnitt pausiert, der längste wartete 40 s vor `ast-frisch`, bis die
Umgebung nach dem Packklon unter der Grenze lag. Median über fünf
Wiederholungen; Zuteilung je Durchlauf; RSS und Spitze des ganzen
Testprozesses. Die Rohzahlen samt Umgebung je Schnitt stehen im Beleg.

| Schnitt | Median | Bereich | Zuteilung | Zuteilungen | RSS · Spitze |
| --- | --- | --- | --- | --- | --- |
| `ast-frisch` | 58 µs | 56–58 µs | 134 KB | 1.200 | 245 · 241 MB |
| `ast-speicher` | 7 ns — **nur Treffer** | | 0 | 0 | 9 · 5 MB |
| `ebene`, kalt | 5,77 ms | 5,75–5,87 ms | 769 KB | 9 | 17 · 14 MB |
| `ebene`, warm | 17 ns — **nur Treffer** | | 0 | 0 | |
| `einfahren`, klein (`Vanguard`) | 46,8 ms | 45,8–49,3 ms | 2,5 MB | 28.066 | 151 · 109 MB |
| `einfahren`, groß (`Black`, 3.788 Karten) | 1,89 s | 1,87–1,91 s | 100 MB | 1,1 Mio. | |
| `naht-skript` | 9,35 ms | 9,25–9,50 ms | 22,9 MB | 72.145 | 479 · 442 MB |
| `naht-karte` | 57,0 ms | 53,4–60,8 ms | 31,4 MB | 79.104 | 480 · 443 MB |
| `werk-frisch` | 671 ms | 647–690 ms | 1,08 GB | 2,19 Mio. | 500 · 462 MB |
| `werk-gehalten` | 56,2 ms | 54,5–58,0 ms | 31,4 MB | 78.970 | 497 · 464 MB |
| `ausfuhr` (70 Karten) | 713 ms | 704–717 ms | 1,66 GB | 6,73 Mio. | 120 · 86 MB |
| `rastern` (`Black`) | 63,8 ms | 58,4–75,6 ms | 57,1 MB | 274.382 | 537 · 497 MB |
| `mcp`, `1 + 1` | 9,76 ms | p95 10,65 ms | | | 482 · 446 MB |
| `mcp`, `to_upper("abc")` | 10,02 ms | p95 10,73 ms | | | |
| `mcp`, `card.name` | 9,96 ms | p95 10,78 ms | | | |

**Was auffällt — und was davon schon eine Aussage ist**:

- **Ein Ausdruck über die Naht kostet 22,9 MB Zuteilung** (`card.name` in
  `naht-skript`), und über `mse-mcp` braucht `1 + 1` dieselbe Zeit wie
  `card.name` — es liegt also nicht am Ausdruck. Vermutet ist, dass die
  Skriptnaht je Ruf die Umgebung der Karte samt ihrer Wertskripte baut;
  **gemessen ist das nicht**, und es ist die nächste Frage an `naht-skript`.
- **`werk-frisch` gegen `werk-gehalten`**: 671 gegen 56 ms — das gehaltene
  Werk ist der Grund, warum das Fenster und seit #2005 die Flächen ohne
  Fenster nicht je Karte neu laden.
- **Die Ausfuhr allokiert je Karte rund 24 MB** — fast so viel wie eine
  ganze Karte über die Naht (31 MB), obwohl sie nichts rastert. Wohin, steht
  in §7.
- `rastern` streut weiter als sein Nachbar `werk-gehalten`, bei gleicher
  Zuteilung; die Last stieg in diesem Schnitt nicht. Eine Ursache ist nicht
  bestimmt.

## 6 PERF-04: dieselbe Karte, gepaart

**PERF-04** (#2171),
im offiziellen Lauf: `Archenemy`, Karte 4 als A, je Runde fünf andere als B,
fünf Runden über das **behaltene** Werk der Flächen ohne Fenster, Güte
„mittel". Je Zeichnung zählt der Beleg, was die drei Zwischenspeicher dabei
getroffen und verfehlt haben (`werkspeicher.go`).

| Schritt | n | Median | Bereich | Paketbilder (Treffer/Fehlgriffe je Zeichnung) | Ebenen (Treffer/Fehlgriffe) |
| --- | --- | --- | --- | --- | --- |
| A kalt — erstes Bild des Werks | 1 | 816 ms | | 0 / 6 | 0 / 6 |
| A vorher | 5 | 55,7 ms | 54,5–59,0 ms | 6 / 0 | 3 / 3 |
| A sofort noch einmal | 5 | 58,0 ms | 56,6–58,5 ms | 6 / 0 | 3 / 3 |
| B, fünf andere Karten | 25 | 62,0 ms | 55,1–111,8 ms | 8 / 0,2 im Mittel | 2,6 / 3,4 im Mittel |
| A nachher | 5 | 57,4 ms | 55,5–57,9 ms | 6 / 0 | 3 / 3 |

**Alle elf Bilder von A tragen denselben SHA-256.** Das Spielbildfach des
Tisches zählt nichts — es gehört nicht zum Kartenweg.

**Was das sagt**: dieselbe Karte sofort noch einmal ist **nicht** schneller
als vorher oder nach fünf anderen — die Mediane liegen innerhalb einer
Spanne von 2,3 ms, und die Zähler sind bei allen dreien **dieselben**. Es gibt
keinen Ergebniszwischenspeicher für das fertige Kartenbild; jede Zeichnung
rechnet Plan und Raster neu, trifft sechs Paketbilder und **verfehlt drei von
sechs Ebenen jedes Mal** — auch unmittelbar hintereinander. Welche drei das
sind, ist nicht bestimmt; der Ebenenspeicher merkt nach seinem Kopf nur
Quellen, die ein Zeiger sind (`render/ebenenspeicher.go`).

**Was es nicht sagt**: dass sich eine Optimierung lohnt. Die erste B-Runde ist
langsamer (76–112 ms), weil sie neue Paketbilder lädt; danach liegen B und A
beieinander. **Eine Rendereroptimierung ist nicht begonnen und nicht
verlangt** (PERF-04: Diagnose ohne Pflichtoptimierung); der nächste Schritt
wäre, die drei verfehlten Ebenen zu benennen — ein eigener Schnitt mit eigener
Messung.

## 7 Zwei Optimierungen danach — und wie sie gemessen sind (AK-06, AK-07)

**Eine Klasse je Änderung, das Budget vor dem Code, die Ausgabe Byte für Byte
gleich.** Für beide gilt derselbe Weg:

1. Der Hotspot steht im Profil des Schnitts, nicht in einer Vermutung.
2. Rote Tests vor dem Produktivcode, dazu eine Gegenprobe, die zeigt, dass der
   Wächter greift.
3. **Gleiche Ausgabe**: `scripts/leistungssemantik.sh` fährt Vorher und Nachher
   über drei kanonische Sets — Kartenliste, vier Ausfuhren samt Befunden, je 25
   Karten als Bild — und vergleicht jede Datei. Für beide: **365 Dateien,
   gleich**.
4. **A/B/A** mit denselben Eingängen und derselben Grenze: der Stand ohne,
   der mit, wieder der ohne.

**Gemessen am 2026-10-07**, Apple M5, macOS 27.0.1, go1.27.1; A ist
`e45599e03`, „PERF-02" `1360e74b3`, „PERF-02 und 03" `a1b542975`. Median über
fünf Wiederholungen, die MCP-Zeile über 100 Rufe:

| Schnitt | A | PERF-02 | PERF-02 und 03 | A |
| --- | --- | --- | --- | --- |
| `naht-skript` | 12,8 ms · 25,6 MB | **9,5 ms** · 25,1 MB | 9,3 ms · 22,9 MB | 12,1 ms · 25,6 MB |
| `mcp`, `1 + 1` | 12,69 ms | **9,79 ms** | 9,40 ms | 12,48 ms |
| `mcp`, `card.name` | 12,66 ms | **9,78 ms** | 9,52 ms | 12,51 ms |
| `naht-karte` | 57,0 ms · 34,1 MB | 57,6 ms · 33,6 MB | **52,9 ms · 31,4 MB** | 55,9 ms · 34,1 MB |
| `ausfuhr` | 721 ms · 1.816 MB | 706 ms · 1.816 MB | 685 ms · **1.658 MB** | 714 ms · 1.816 MB |
| `rastern` | 54,5 ms · 59,1 MB | 54,5 ms · 59,2 MB | 56,1 ms · 57,1 MB | 56,8 ms · 59,2 MB |
| `werk-frisch` | 689 ms · 1.087 MB | 641 ms · 1.087 MB | 656 ms · 1.085 MB | 666 ms · 1.087 MB |

- **PERF-02** (#2169),
  die Paketwurzeln gemerkt (`msepaket.Wurzelsuche`): ein warmer Skriptruf über
  die Naht und über `mse-mcp` wird rund **ein Viertel** schneller, an beiden
  A-Läufen gemessen. Die Zuteilung bewegt sich kaum — gespart sind Systemrufe,
  nicht Speicher. **Gemessen mit der Prüfung über die Änderungszeit allein**;
  seit dem Zweitblick über Phase I prüft sie Änderungs- und Statuszeit,
  Rechte und Knoten (`msepaket/wurzelstempel.go`) — mit **demselben einen
  `lstat`** je gelesenem Ordner, nur über `syscall.Lstat` statt `os.Lstat`.
  Nachgemessen ist das nicht; die Zahl gilt für den Stand `1360e74b3`.
- **PERF-03** (#2170),
  `card` nur neu, wenn sich ein Wert geändert hat: je Ausfuhr **158 MB**
  weniger Zuteilung (8,7 %); das Profil hatte rund 182 MB je Lauf in diesem
  Neubau gezeigt. Die Zeit der Ausfuhr bewegt sich im Bereich der beiden
  A-Läufe — **eine Zeitersparnis ist nicht belegt**. Eine Karte über die Naht
  wird um rund 2 MB kleiner.

**Was als Nächstes käme, und was es nicht ist**: in der Ausfuhr lagen weiter
rund 800 MB je Lauf im Aufruf von Skriptfunktionen (`msescript.ruferufen`,
`verkettet`) und rund 350 MB im erneuten Zerteilen der Paketdateien je
Ausfuhr. **Beides hat E03 aufgenommen** (§7.1).

### 7.1 E03: die Zuteilung halbiert

**#2240.**
Ziel war höchstens die Hälfte der bestätigten Zuteilung von `ausfuhr` und
`werk-frisch`. Der Ausgangswert ist **am selben Tag neu erhoben**, nicht aus
§5 übernommen, und zwar mit demselben Verfahren wie der neue Stand.

Drei Klassen, jede mit einem roten Test vor dem Produktivcode:

| Klasse | Wo | Was sie kostete | Test |
| --- | --- | --- | --- |
| **Ein Glied einer `+`-Kette kopierte seine Argumente** — zweimal je Glied, dazu ließ es die Liste für `input` nachwachsen | `msescript/auswerter_ruf.go` (`verbunden`), `auswerter_verknuepfung.go` (`verkettet`) | rund 1,1 GB der 1,66 GB je Ausfuhr | `kettenzuteilung_test.go`: höchstens zwei Zuteilungen je Glied, und die Liste nicht größer als nötig, auch unter `-race` |
| **Die Vorlagenliste las das Spielpaket für jede Vorlage neu** — im vollen Pack 22 Vorlagen, für ein `magic`-Set jedes Mal `magic.mse-game` samt Modulen | `service/vorlagenliste.go` | rund 350 MB je Ausfuhr | `vorlagenspieleinmal_test.go`: drei Vorlagen tragen **dasselbe** Spielmodell |
| **Der Namensindex der Schriften las jede Datei ganz** — `Apple Color Emoji.ttc` mit 192 MB eingeschlossen | `schriften/namenstabelle.go`, `pakete.go` | rund 860 MB je frischem Werk | `namenteilweise_test.go`: eine um 16 MiB aufgeblähte Schrift kostet keine 8 MiB |

**Gemerkt wird dabei nichts Neues über einen Ruf hinaus.** Das Spielpaket
gilt innerhalb eines `Vorlagen`-Rufs; „nicht gemerkt" über Rufe hinweg
bleibt, wie es war. Der Namensindex war schon vorher je Suche gemerkt; neu
ist nur, **wie viel** er dafür liest. Ein Zwischenspeicher, dessen Größe zu
begrenzen wäre, ist nicht dazugekommen.

**Gemessen am 2026-10-10**, Apple M5 (10 Kerne), macOS 27.0.1, go1.27.2. A ist
`a0569df45`, B `6b5404029`; die Testprogramme hatten die SHA-256
`3e282e8e…` (A) und `980e09d8…` (B). Verfahren: **A, B, A** nacheinander,
Archenemy über dem vollen Pack. **Rohwerte, Befehl und vollständige
Prüfsummen** stehen in [`belege/e03-zuteilung.md`](belege/e03-zuteilung.md)
(nachgetragen am 2026-10-11, J05).

- **Zuteilung und Zuteilungen je Lauf**: Speicherprofil mit
  `memprofilerate=1`, also jede Zuteilung gezählt statt geschätzt; ein Lauf
  über einen Durchlauf und einer über vier, die Differenz geteilt durch drei.
  So fallen Aufbau und erster Durchlauf heraus — dieselbe Abgrenzung wie
  `B/op`.
- **Lebender Heap**: `inuse_space` desselben Profils nach vier Durchläufen.
  `testing` schreibt es nach einer Speicherbereinigung; es ist, was der
  Prozess danach noch hält.
- **RSS und Spitze**: `/usr/bin/time -l` über einen eigenen Lauf mit fünf
  Durchläufen bei gewöhnlicher Profilrate.

| Schnitt | Größe | A | B | A | B gegen A |
| --- | --- | --- | --- | --- | --- |
| `ausfuhr` | Zuteilung je Lauf | 1.581,0 MB | **778,4 MB** | 1.580,4 MB | **49,2 %** |
| | Zuteilungen je Lauf | 6,67 Mio. | **3,15 Mio.** | 6,67 Mio. | 47,2 % |
| | lebender Heap nach vier Läufen | 7,61 MB | 7,64 MB | 7,59 MB | gleich |
| | RSS · Spitze | 122 · 88 MB | **100 · 66 MB** | 127 · 93 MB | |
| `werk-frisch` | Zuteilung je Lauf | 1.032,8 MB | **196,7 MB** | 1.033,6 MB | **19,0 %** |
| | Zuteilungen je Lauf | 1,447 Mio. | 1,440 Mio. | 1,447 Mio. | 99,6 % |
| | lebender Heap nach vier Läufen | 7,58 MB | 7,57 MB | 7,59 MB | gleich |
| | RSS · Spitze | 524 · 489 MB | **103 · 68 MB** | 523 · 489 MB | |

**Die Ausfuhr hält das Ziel knapp** — 49,2 %, und der Rest liegt fast ganz im
Zusammenführen gebundener und benannter Argumente (`verbunden`, rund 220 MB je
Lauf): ein `replace@(…)` in einer Kette braucht beide Listen in einer, weil die
Bibliothek `Benannt` so liest. Sie getrennt zu halten änderte die Gestalt jedes
eingebauten Rufs; das ist ein eigener Schnitt und nicht dieser. **Das frische
Werk** liegt weit darunter, und die Zahl der Zuteilungen bewegt sich dort kaum:
gespart sind wenige, sehr große — die ganzen Schriftdateien.

**Der Ausgangswert liegt unter den 1.658 MB aus §7.** Dort steht `B/op` aus
`-benchmem` über einen anderen Stand (`a1b542975`); hier zählt das Profil am
Stand vor E03. Verglichen wird nur, was mit **demselben** Verfahren gemessen
ist.

**Die Ausgabe ist dieselbe**: `scripts/leistungssemantik.sh` mit `mse` aus A
und aus B, 365 Dateien je Stand, **alle gleich**, kein Schritt gescheitert —
Kartenlisten, Ausfuhren samt Befunden und je 25 Karten als Bild, damit auch die
Wahl der Schrift.

**Was diese Messung nicht ist:**

- **Keine Zeitmessung.** Sie lief neben einer Fuzzkampagne (Last um 13 auf
  zehn Kernen). Zuteilung und Heap hängen daran nicht; die Zeit schon, und
  die Schwellen der Benchmarks rissen dabei — deshalb fuhr nicht der Läufer
  aus §1, der einen gerissenen Lauf zu Recht nicht zählt. **Ob die Ausfuhr
  schneller wird, hat E04 gemessen**
  (#2241,
  [`belege/e04-ausfuhr.md`](belege/e04-ausfuhr.md)): `magic-namecheck`
  über Archenemy 0,57 s gegen MSE2s 0,87 s, bytegleich; über Black bleibt
  es das Dreifache (`LEISTUNGSVERGLEICH.md`).
- **Nicht unter dem Race-Detektor.** Dort verdoppelt sich die Zuteilung
  (§5); die Tests halten die Zusage auch unter `-race`, die Zahlen hier sind
  ohne.
- **Zwei Schnitte**, nicht alle. `naht-karte` und `werk-gehalten` nehmen
  dieselbe Kette; um wie viel sie kleiner werden, ist nicht gemessen.

## 8 Was diese Zahlen nicht sagen

- **Eine Maschine** (`AGENTS.md` §4): ein Apple M5 mit zehn Kernen, auf dem
  `StorageManagementService` dauernd läuft. Die Grenze hält Ausreißer fern,
  macht den Rechner aber nicht ruhig.
- **Kopflos.** Kein Schnitt misst die Latenz des Fensters — weder die Wails-Naht
  noch das Zeichnen der WebView. `werk-gehalten` und `rastern` nehmen den Weg,
  den das Fenster nimmt, ohne das Fenster.
- **Gleiche Ausgabe** heißt: dieselben Bytes für drei Sets, vier Vorlagen und
  je 25 Karten — nicht für jede Karte jedes Sets.
- **„unsauber" im Beleg der A/B/A-Läufe** steht dort, weil die Arbeitsbäume
  den Läufer als Kopie und `.go` als Verweis trugen; am gebauten Programm
  änderte das nichts. Der offizielle Lauf (§5) lief auf sauberem Baum.

## 9 Am Fenster: die Bedienbudgets (B02)

**#2221.**
Die Schnitte oben sind kopflos; dieser Abschnitt misst, was ein Mensch am
**gebauten** Fenster erlebt. Der Messvertrag steht im Kopf von
`scripts/huellenprobe/budgets.mjs`, der Beleg in
[`belege/huellenprobe/budgets.md`](belege/huellenprobe/budgets.md), der
Gastlauf in [`belege/huellenprobe/budgets-gast.md`](belege/huellenprobe/budgets-gast.md).

**Was die ersten vier Läufe falsch maßen**, und warum die Zahlen davor nicht
gelten:

- **Die Startzeit enthielt eine halbe Sekunde Pause** — `Probe.bereit` wartet
  nach der stehenden Oberfläche 500 ms, und die Zeit lief bis dahin. Seit
  dem fünften Lauf endet sie, wenn die Leiste antwortet, alle 20 ms gefragt:
  rund 1.000 ms wurden 480 ms.
- **Die Vorschaumessung lief mit der Taste um die Wette** — Horcher und Taste
  gingen zugleich hinaus, und kam die Taste zuerst, wartete die Messung auf
  ein `keydown`, das vorbei war. Drei bis fünf von 24 „ohne Bild" waren
  keine Lücke der Vorschau. Seither wird die Messung **vor** der Taste
  angelegt, ebenso die des Seitenwechsels vor dem Klick.

| Reihe | Wirt: Median · P95 · max | Gast: Median · max | Vorschlag | Entscheidung |
| --- | --- | --- | --- | --- |
| Start, kalt | 500 · — · 514 ms | 1.532 · **2.002** ms | ≤ 2 s | **bestätigt als Median**; der erste Start im Gast streift die Grenze (unten) |
| Start, warm | 514 · — · 546 ms | 1.529 · 1.622 ms | ≤ 2 s | **bestätigt** |
| Feld → dekodiertes Vorschaubild | 27 · 33 · 97 ms | nicht messbar | Median ≤ 150 ms | **bestätigt** |
| Seitenwechsel über die Palette | 49 · 72 · 73 ms | nicht messbar | — | Ausgangslage, kein Budget |
| Rollbild in `Black` (3.788 Karten) | 22 · **55** · 87 ms | nicht messbar | ≤ 50 ms | **nicht gehalten** — P95 47 bis 55 ms in fünf Läufen; Folgepaket E07 (#2279) |

**Am Endstand der Phase E** (Commit `f1e733bde`, Abnahmebündel, Last vor dem
Lauf 7,6 fallend auf unter 5; Belege
[`budgets.md`](belege/huellenprobe/budgets.md),
[`e01vorschau.md`](belege/huellenprobe/e01vorschau.md),
[`e05start.md`](belege/huellenprobe/e05start.md),
[`e07rollen.md`](belege/huellenprobe/e07rollen.md)) — mit allen Portierungen
aus F02, Textmaske und Maßstabssuche eingeschlossen:

| Reihe | Wirt: Median · P95 · max | Budget | Urteil |
| --- | --- | --- | --- |
| Start, kalt | 494 · — · 508 ms | ≤ 2 s | **gehalten** |
| Start, warm | 499 · — · 511 ms | ≤ 2 s | **gehalten** |
| Feld → dekodiertes Vorschaubild | 28 · 37 · 38 ms | Median ≤ 150 ms | **gehalten** |
| Seitenwechsel über die Palette | 50 · 72 · 77 ms | — | wie zuvor |
| Rollbild in `Black` | 22 · **41** · 74 ms | P95 ≤ 50 ms | **gehalten** — E07 (#2279): P95 36 bis 42 ms in fünf Runden des ersten Rollens |
| Bedienbar nach dem Start (E05) | 492 bis 507 ms je Reihe — Saat, warm, älterer Bestand | Median ≤ 2 s | **gehalten**; die erste Eingabe nach 17 bis 30 ms, ohne Vorwärmen |
| Vorschau nach einem Schub (E01) | in vier Läufen steht das Bild des **letzten** Entwurfs; ein Fehler der Vorschau steht da und vergeht | — | **gehalten** |

**Kopflos (E02)**, am sauberen Stand `37f8a70b4`, Median über fünf
([`leistungsschnitte-e02.json`](belege/leistungsschnitte-e02.json)):
`naht-karte` 54,6 ms (P95 59,5), `werk-gehalten` 53,9 ms, `rastern` 57,8 ms
— **das Ziel „warm ≤ 60 ms" aus E02 hält**, auch mit Textmaske, Kastenfilter
und Maßstabssuche aus F02. Ein Lauf am unsauberen Stand davor gab 56,0 ·
53,0 · 57,3 ms; beide hatten Fremdlast unter der Grenze (`fseventsd` nach dem
Bau). **Die Grundlinie aus B02** bleibt in `leistungsschnitte.json`, mit
allen Schnitten.

**Die Wirtszahlen stammen aus dem Schlusslauf** der Phase B (Commit
`1c0b74221`, Lauf `2026-10-09T19-47-35Z-46912`, Last vor dem Lauf 6,75 /
6,28 / 6,26 auf zehn Kernen), die Gastzahlen aus den Gastläufen davor.

**Wirt**: Apple M5, zehn Kerne, macOS 27.0.1, Last 4,4 / 4,9 / 5,3 —
`StorageManagementService` und seine Erweiterung laufen dauernd mit je einem
Kern (§8). **Gast**: Lume-VM, macOS 15.2, vier Kerne, 4 GB.

**Warum das Rollbild nicht als gehalten gilt**: in fünf Läufen lag der P95
bei 47 bis 55 ms, der Höchstwert bei 59 bis 87 ms — im Schlusslauf 55 und 87.
Bis zum vierten Lauf stand hier „bestätigt als P95"; der fünfte riss es, und
eine Grenze, die nur in vier von fünf Läufen hält, ist keine. **Die langsamen Bilder liegen am
Anfang**: im Lauf vom 2026-10-09 12:52 UTC lagen die ersten 39 Bilder bei 25
bis 51 ms, die folgenden 120 meist bei 13 bis 20 und einzeln bis 59 ms.
**Vermutet**, nicht gemessen: Zeilen, die zum ersten Mal in das Fenster der
Liste kommen (`Kartenliste` zeichnet nur, was sichtbar ist). Ein Höchstwert als Budget hielte jedes
einzelne Bild fest, auch eine fremde Last auf diesem Rechner; der P95 hält
die Regel. **Was das kostet, steht offen**: einzelne Bilder über 50 ms beim
ersten Rollen sind sichtbar, und ob sie sich vermeiden lassen, ist eine Frage
an Phase E (Leistung), nicht an dieses Budget.

**Der erste Start im Gast ist der langsamste**: in den zwei Gastläufen mit
Rohwerten lag er bei 1.744 und 2.002 ms, alle übrigen zwanzig Starts bei
1.430 bis 1.622 ms; in den zwei anderen lag der Höchstwert der Kaltstarts bei
1.975 und 2.022 ms. Es ist der erste Start des eben kopierten Bündels, und
**vermutet** ist, dass das System dabei das neue Programm prüft; gemessen ist
die Ursache nicht. Gehalten wird das Budget als Median; ein erster Start nach
der Installation ist am Wirt nicht gemessen.

**Warum der Gast nur die Starts misst**: er steht nach dem Start hinter seinem
Sperrbildschirm, und dahinter zeichnet WebKit nicht — null
`requestAnimationFrame` in zwei Sekunden, mit und ohne Anzeige, bei laufenden
Zeitgebern. Entsperren hieße ein Kennwort eingeben (`AGENTS.md` §17). **Ein
Gast hat außerdem keine GPU des Wirts**; seine Zahlen stammen aus einer
eigenen Messumgebung — vier Kerne, 4 GB, macOS 15.2, ohne Grafikbeschleunigung
— und sind mit denen des Wirts nur eingeschränkt vergleichbar. Eine Obergrenze
für einen anderen Rechner sind sie nicht.

**Speicher**: der Go-Prozess steht nach dem Start bei rund 160 MB und nach
dem Öffnen des Sets und 24 Vorschauen bei rund 600 MB — die Größenordnung des
gehaltenen Werks aus §5 (`werk-gehalten`, 497 MB RSS im Testprozess). Ob er
**weiter** wächst, beantwortet der Dauerbetrieb (B04), nicht diese Messung.

### 9.1 Im Dauerbetrieb (B04)

**#2223.**
Eine Sitzung, 125 Minuten gemischter Arbeit — Bearbeiten, Navigation,
Sammlung, Tisch, Ausgabe —, je 30 s RSS von Anwendung und WebKit-Prozessen
(`scripts/huellenprobe/dauerbetrieb.mjs`, Beleg
[`belege/huellenprobe/dauerbetrieb.md`](belege/huellenprobe/dauerbetrieb.md)).
Ziel: nach der Aufwärmzeit **≤ 10 % je Stunde und kein stetiger Anstieg über
drei Abschnitte**.

| Lauf | Runden | Speicher je Abschnitt (MB) | je Stunde | Urteil |
| --- | ---: | --- | ---: | --- |
| vor der Behebung | 1.667 | Web 259 → 4.800, Anwendung 159 → 1.800 | +87 % (damalige Rechnung) | rot |
| Schlusslauf `1c0b74221` | 1.712 | 1.324 · 1.054 · 1.160 · 1.228 · 1.318 · (1.559, angebrochen) | −0,3 % erster → letzter ganzer Abschnitt; **+25 % von Minute 40 bis 110** | **rot** — stetig über vier Abschnitte |
| Stand `098bfbc06` (nach #2301) | 1.721 | 1.154 · 1.272 · 1.413 · 1.514 · 1.579 | **+27,6 %** | **rot** — stetig; diesmal wächst auch die Anwendung (unten) |
| Stand `dfaad91b9` (E06, mit Abzügen des Go-Heaps alle 20 min) | 1.717 | 866 · 979 · 1.083 · 1.196 · 1.282 | **+36 %** | **rot** — stetig; der Go-Heap füllt das Spielbildfach, der Webprozess wächst ab Minute 70 (unten) |
| Stand `2913d2853` (E06, nach Verweiszählung und `no-store`, mit Abzügen) | 1.734 | 881 · 938 · 998 · 1.050 · 1.112 | **+19,7 %** | **rot** — stetig; der Go-Heap steht bei 11 MB, gemessen ist RSS (unten: was RSS nicht misst) |
| Stand `40e4ac2ee` (E06, **Fußabdruck** statt RSS, mit Abzügen) | 1.710 | 798 · 847 · 849 · 879 · 972 — nur die Prozesse der Anwendung: 797 · 844 · 843 · 870 · 963 | **+15,6 %** | **ungültig** — zwei fremde WebKit-Prozesse in der Summe, und die Last der Probe wuchs mit (unten) |
| Stand `b9a943091` (E06, Fußabdruck, **gleichbleibende Last**, mit Abzügen) | 1.751 | nur die Prozesse der Anwendung: 746 · 843 · 855 · 875 · 766 | **+2,0 %** | nach dem Wortlaut **rot** (die erste Stunde steigt über drei Abschnitte); nach der Ausweichklausel aus #2278 **gehalten** — der Anstieg ist der Zwischenspeicher von WebKit, und er fällt wieder (unten) |
| Stand `34ef8a1a4` (QA01, Uhr alle 30 s) | 1.500 | 826 · 806 · 832 · 873 · 739 | −7,9 % | nach der Regel **rot**, nach heutiger Bewertung **ungültig**: die Proben fielen in vier verschiedene Teile der Runde (unten) |
| Stand `589bdf8e5` (#2527, **Probe zwischen zwei Runden**) | 1.501 | 507 · 540 · 573 · 588 · 590 | **+12,3 %** | **rot** — stetig; Webinhalt +84 MiB in der ersten Stunde, danach langsamer; Anwendung +0,166 MB/min (#2312) |
| Stand `fe167f897` (**Vermerk 0204**: Aufwärmstunde, Anstieg ab 1 %, Speicherwarnung am Ende) | 1.496 | erste Stunde 519 · 512 · 520; danach **557 · 590 · 601** | **+11,8 %** | **rot** — stetig; der Webinhalt wächst diesmal erst in der zweiten Stunde (0,65 MB/min) und **gibt zurück**: nach Ruhe 311, nach der Warnung 242 MiB (unten) |

**Die Rate hieß hier zuerst +10,6 %**, und das war ein Rechenfehler der
Auswertung: sie zählte den fünf Minuten langen Rest des Laufs wie einen
ganzen Abschnitt und rechnete die Zeit aus der Zahl der Abschnitte
(Zweitblick über Phase B, Befund 7). Nachgerechnet über ganze Abschnitte und
ihre Zeit liegt der letzte (1.318 MB) beinahe beim ersten (1.324) — aber der
erste ist der Stand **vor** der Rückgabe an das System bei Minute 50, und ab
Minute 40 steigt der Speicher stetig. Das Urteil bleibt rot, und zwar
deutlicher als es die erste Zahl sagte.

**Zwei Lecks sind gefunden und behoben**, beide **unter** dem DOM und
**neben** dem Go-Heap: die Texturen des Tisches blieben im globalen Speicher
von PixiJS (`Assets.load` ohne `Assets.unload`), und jede Menüleiste baute ein
neues `NSMenu`, ohne das alte freizugeben — behoben mit zwei Menüs im
Wechsel, denn ein einziges nahm macOS den Vollbildeintrag (Drehbuch
`vollbild`). Nachgemessen: sechs Minuten Navigation und Bearbeiten, und
`heap` zählt durchweg 107 `MenuItem`, 100 `NSMenuItem` und 30 `NSMenu`
(vorher 14.454 `MenuItem` nach fünf Minuten). Eingegrenzt mit
`scripts/huellenprobe/dauerteile.mjs` (je Teil der Runde, mit Ausschluss und
Steigung je Phase).

**Was bleibt**: die Anwendung trägt nichts mehr bei — über die ganze Zeit
−0,34 MB je Minute; sie gab bei Minute 50 Speicher an das System zurück und
wuchs danach auf den Stand davor. **Der Webprozess wächst weiter**, 653 →
995 MB in den letzten 75 Minuten und am Ende schneller (788 → 872 → 995).
Das ist ein drittes Leck — Folgepaket E06
(#2278).

#### E06: was der Lauf vom 2026-10-11 mit Abzügen zeigte

**Gelaufen an `dfaad91b9`** (Leitungsbefehl `speicher`, alle 20 Minuten ein
Stand nach `runtime.GC` und ein Heap-Profil), 125 Minuten, 1.717 Runden,
kein Absturz, nichts verloren. **Rot**: +36 % je Stunde, stetig.

- **Der Go-Heap** (`HeapAlloc`) wuchs von 16 auf 156 MB und steht ab Minute
  100 still. `pprof -base` zwischen den Abzügen: fast alles in
  `vorschauweg.Texturfach.Hinterlege` — das Fach der Spielbilder füllte sich
  bis an sein Budget (128 MB). **Der Grund**: jede Runde ändert eine Karte,
  ihr Spielbild bekommt einen neuen Schlüssel (er folgt aus dem Inhalt), und
  das alte gehört danach keiner Karte mehr, blieb aber liegen, bis das Budget
  es verdrängte. Die Anwendung wuchs dabei um 3,1 MB je Minute — Go gibt den
  freien Teil seines Heaps nur zögernd zurück (`HeapIdle − HeapReleased` am
  Ende 175 MB).
- **Der Webprozess** stand bis Minute 70 bei rund 600 MB und wuchs dann bis
  816 MB, am Ende um 3,5 MB je Minute. Vorschau und Spielbild gingen mit
  `Cache-Control: public, max-age=31536000, immutable` hinaus — und weil jede
  Bearbeitung eine neue Adresse gibt, behielt WebKit jede alte Antwort.
  **Vermutet, nicht gemessen**: ein Speicher von WebKit mit Obergrenze; der
  nächste Lauf entscheidet es.

**Behoben, beide testgetrieben**: das Fach zählt jetzt, wie viele Karten auf
ein Bild zeigen, und wirft es weg, wenn es keine mehr ist
(`service.texturverweise`; sieben gleiche Karten teilen sich weiter eines),
und Vorschau wie Spielbild gehen mit `Cache-Control: no-store` hinaus — das
Bild hält das Fach in Go, ein zweiter Abruf ist eine Kopie.

#### E06: der Lauf nach beiden Korrekturen — und was RSS nicht misst

**Gelaufen an `2913d2853`** am 2026-10-11 ab 23:40, 125 Minuten, 1.734
Runden, kein Absturz, nichts verloren, Abzüge des Go-Heaps alle 20 Minuten
(Beleg [`belege/huellenprobe/dauerbetrieb.md`](belege/huellenprobe/dauerbetrieb.md),
Commit `6ad987746`). **Rot, aber halb so steil**: +19,7 % je Stunde statt
+36 %, stetig.

- **Der Go-Heap hält nichts mehr**: `HeapAlloc` nach Bereinigung 11 MB in
  allen acht Abzügen, rund 71.000 Objekte. Was Go vom System hält, ohne es zu
  nutzen (`HeapIdle − HeapReleased`), sinkt von 13 auf 8 MB. **Die
  Verweiszählung im Fach wirkt.**
- **Trotzdem wächst die Anwendung im RSS** — im Mittel je 25 Minuten 219 →
  243 → 266 → 284 → 307 MB — und der Webprozess 589 → 657 → 714 → 762 →
  820 MB. `no-store` hat ihn nicht flacher gemacht.

**Ein Diagnoselauf über 20 Minuten** (2026-10-12, 01:44, derselbe Stand,
Abzüge mit `footprint`, `vmmap` und `heap` bei Minute 3, 10 und 18) zeigt,
**dass RSS hier das falsche Maß ist** — in beide Richtungen:

| Prozess | RSS | physischer Fußabdruck | warum sie auseinanderliegen |
| --- | ---: | ---: | --- |
| Anwendung | 215 → 228 MB | 86,9 → 88,5 → 89,2 MB | RSS zählt Seiten, die Go mit `MADV_FREE_REUSABLE` schon zurückgegeben hat (`HeapReleased` im Zwei-Stunden-Lauf 10 → 71 MB); der Fußabdruck nicht |
| Webinhalt | 366 → 383 MB | 323 · 610 · 430 · 540 MB | der Grafikspeicher (IOSurface, `Owned physical footprint (graphics)`, 181 bis 452 MB) steht **nicht** im RSS und schwankt mit dem, was gerade gezeichnet ist |

- **Die Anwendung wächst im Fußabdruck kaum**: 2,3 MB in 15 Minuten, am
  Ende langsamer. Ein **kleines echtes Leck** zeigt `heap` trotzdem:
  `_WKUserInitiatedAction` (8.685 Objekte, 0,7 MB) und
  `NSKeyValueDependency` (3.720) wachsen mit jedem Klick des Drehbuchs —
  Objekte von WebKit im Prozess der Anwendung, gehalten von einer Stelle,
  die nicht gefunden ist.
- **Im Webinhalt schrumpft der Speicher der Seite selbst** (`WebKit Malloc`,
  dort liegen JavaScript und DOM): 107 → 65 MB schmutzig, der Rest ist als
  freigebbar markiert. Was wächst und schwankt, ist der Grafikspeicher.
- **Ob er eine Obergrenze hat, ist nicht gemessen.** Eine Speicherwarnung an
  WebKit (`notifyutil -p org.WebKit.lowMemory`) mitten im Lauf senkte den
  Fußabdruck des Webinhalts von 610 auf 430 MB — bei laufendem Zeichnen ist
  das kein Beleg.

**E06 bleibt offen** (#2278),
und der nächste Schritt ist das Maß: der Zwei-Stunden-Lauf misst den
physischen Fußabdruck statt RSS, den Grafikspeicher des Webinhalts getrennt,
und am Ende eine Ruhephase mit Speicherwarnung. **Bis dahin gilt für 0.8.0
die begrenzte Zusage** (README, „Was diese Fassung nicht belegt").

#### E06: der Fußabdruck als Maß — und eine Probe, deren Last wuchs

**Gelaufen an `40e4ac2ee`** am 2026-10-12 ab 06:54 (Lauf
`2026-10-12T04-54-27Z-60044`), 125 Minuten, 1.710 Runden, kein Absturz,
nichts verloren, Abzüge des Go-Heaps alle 20 Minuten, am Ende zwei Minuten
Ruhe und eine Speicherwarnung. **Gemessen ist der physische Fußabdruck**
(`footprint`) je Prozess. Je 20-Minuten-Abschnitt 798 · 847 · 849 · 879 ·
972 MB, +16,4 %/h, stetig — **und das Urteil gilt nicht**, aus zwei Gründen,
die der Zweitblick über die Folgen vor 0.8.0 (Befund 12) und die Durchsicht
des Belegs fanden:

- **Zwei fremde WebKit-Prozesse standen in der Summe.** Die Probe zählte jeden,
  der nach dem Start entstand: einen `WebContent` von Minute 70 bis zum Ende
  und einen `WebContent.EnhancedSecurity` in der letzten Minute — der lebte
  noch, als die Anwendung längst beendet war. **Zurückgerechnet auf die drei
  Prozesse, die mit dem Fenster entstanden**, und den Go-Prozess (genau, weil
  GPU- und Netzprozess eindeutig sind und die Summe jeden enthält): 797 ·
  844 · 843 · 870 · 963 MB, **+15,6 %/h, stetig**.
- **Die Probe ließ ihre eigene Last wachsen.** Jede Runde hängte eine Marke
  an dasselbe Textfeld — nach 1.710 Runden rund 9.000 Zeichen, die jede
  Vorschau und jede Ausgabe neu setzten —, und jeder Münzwurf verlängerte das
  Protokoll derselben Partie. Ein Dauerlauf mit wachsender Last misst kein
  Leck.

**Was der Lauf trotzdem zeigt** (Anteil der Anwendung):

| | Start | während der Arbeit (Median je Abschnitt) | nach 30 s Ruhe | nach der Speicherwarnung |
| --- | ---: | --- | ---: | ---: |
| Go-Prozess | 53 MB | 94 → 124 MB | 134 MB | 127 MB |
| Webinhalt | 190 MB | 588 → 711 MB | 325 MB | 235 MB |
| GPU | 118 MB | 97 → 100 MB | 25 MB | 24 MB |
| zusammen | 367 MB | 797 → 963 MB | **493 MB** | **395 MB** |

- **Was während der Arbeit wächst, ist großteils freigebbar**: in Ruhe fällt
  der Fußabdruck auf die Hälfte, nach der Warnung weiter. **Gesendet ist
  nicht empfangen** — die Probe sieht danach nur den Fußabdruck, nicht, was
  WebKit freigab.
- **Der Go-Heap hält nichts**: `HeapAlloc` nach Bereinigung 11 MB in allen
  acht Abzügen, ungenutzt und nicht zurückgegeben 12 bis 18 MB. `HeapSys`
  wächst von 41 auf 98 MB — als zurückgegebener Adressraum
  (`HeapReleased` 13 → 70 MB): die **Spitzen** einzelner Vorgänge wurden
  größer, und genau das tut eine Last, die mit jeder Runde wächst.
- **Der Go-Prozess wächst trotzdem um rund 0,38 MB/min**, außerhalb des
  Go-Heaps; in Ruhe gibt er nichts zurück. Die Objekte von WebKit in seinem
  Speicher (`_WKUserInitiatedAction`, `NSKeyValueDependency`) erklären davon
  rund 50 KB/min — das übrige ist nicht zugeordnet.

**Behoben in der Probe**: eine Probe ohne vollständige Messwerte und ein
Wechsel der WebKit-Prozesse machen das Urteil ungültig, statt still null zu
zählen; jede Runde **ersetzt** das Textfeld, und die nächste prüft im
Bestand, dass die vorige Marke ankam; alle 50 Runden kommt eine frische
Partie. Der erste Probelauf damit fand einen Fehler der Probe selbst: `⌘A`
markierte nicht immer den Feldinhalt, und eine Marke hing ohne Leerzeichen
am alten Text — markiert im Feld, liefen 70 Runden ohne Verlust.

#### E06: der Lauf mit gleichbleibender Last — und was bleibt

**Gelaufen an `b9a943091`** am 2026-10-12 ab 09:21 (Lauf
`2026-10-12T07-21-07Z-20303`), 125 Minuten, 1.751 Runden, kein Absturz,
keine getippte Änderung verloren — **jede Runde prüfte im Bestand, dass die
vorige ankam**. Das Urteil der Probe war wieder ungültig: das System
startete etwa alle 35 Minuten einen `WebContent.EnhancedSecurity` (8–9 MB),
und der letzte lebte noch, als die Anwendung beendet war. Kein Name
kollidierte mit den drei WebKit-Prozessen der Anwendung; ihr Anteil ist
deshalb genau ablesbar, und die Probe zählt fremde seitdem nicht mehr mit
(`aufteilen`).

| Minute (Abschnitt) | 30 | 50 | 70 | 90 | 110 | nach 30 s Ruhe | nach der Warnung |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| WebKit der Anwendung (Inhalt, GPU, Netz) | 659 | 747 | 756 | 768 | **649** | 289 | 219 |
| Anwendungsprozess (Go) | 86 | 92 | 99 | 102 | 109 | 118 | 108 |
| **zusammen** | **746** | **843** | **855** | **875** | **766** | **407** | **327** |

- **Erster gegen letzten ganzen Abschnitt: +2,0 %/h** — das Ziel von 10 %
  hält.
- **Nach dem Wortlaut verfehlt**: die erste Stunde steigt über drei
  benachbarte Abschnitte (746 → 843 → 855 → 875).
- **Der Anstieg ist der Zwischenspeicher von WebKit, und er hat eine Grenze —
  gemessen**: der Anteil von WebKit steigt in der ersten Stunde um 109 MB,
  steht dann (747 · 756 · 768) und fällt im fünften Abschnitt auf 649 MB;
  in Ruhe gibt er auf 289 MB frei, nach der Warnung auf 219 MB — unter den
  Stand zu Beginn der Arbeit. Das ist die Ausweichklausel aus #2278
  („ein begründet neu gesetztes Ziel, wenn das Wachstum nachweislich ein
  Zwischenspeicher von WebKit mit Obergrenze ist"): **was der Webprozess
  unter gleichbleibender Last belegt, gibt er wieder frei.** Eine Obergrenze
  belegt der Lauf nicht, und ein kleinerer bleibender Anstieg kann unter dem
  Freigebbaren liegen; ganz gleich blieb die Last auch nicht — die abgelösten
  Partien blieben beendet im Bestand, 35 nach zwei Stunden (Nachlauf über die
  Folgen vor 0.8.0, Befund 5; die Probe löscht sie seitdem).
- **Der Go-Heap hält nichts, und seine Spitzen wachsen nicht mehr**:
  `HeapAlloc` 11 MB, `HeapSys` 41,7 MB in allen acht Abzügen — im Lauf davor
  wuchs `HeapSys` mit der Last von 41 auf 98 MB.
- **Was bleibt, liegt im Anwendungsprozess außerhalb des Go-Heaps**: 0,29
  MB/min, rund 17 MB/h, stetig und in Ruhe nicht freigegeben. Die Objekte
  von WebKit in seinem Speicher (`_WKUserInitiatedAction`,
  `NSKeyValueDependency`) erklären davon rund ein Sechstel. **Folgepaket
  #2312**;
  die Zuordnung braucht `MallocStackLogging` an der Abnahmehülle.

**#2312, zugeordnet und zum Teil behoben** (2026-10-12). Ein Diagnoselauf
mit `MallocStackLogging` (`dauerteile`, je Teil sechs Minuten, Abzüge mit
`footprint`, `heap` und `malloc_history`) ordnete das Wachstum zu:

| Quelle | in 24 Minuten | wo |
| --- | ---: | --- |
| Beschriftungen der Menüleiste, je Neubau | 0,93 MB | Wails: `newMenuItem` gibt `label` nie frei (`menuitem_darwin.go`, v3.0.0-beta.27) — **behoben**: eine Leiste derselben Gestalt wird nachgetragen |
| Go-Laufzeit (`Untagged`) | rund 2 MB | Seiten des Go-Heaps; `HeapSys` steht, `HeapSys − HeapReleased` schwankt zwischen 22 und 31 MB |
| IOSurfaces (`Owned physical footprint`) | 1,2 MB | WebKit, Ebenenbaum der Webansicht |
| Malloc von WebKit | 0,4 MB | WebKit |
| Rechtschreibprüfung, Gestenmarken, KVO | 0,9 MB | AppKit (`NSSpellChecker` je Auswahl), WebKit (`recordUserGestureAuthorizationToken` je Taste), AppKit (`NSKeyValueDependency`) |

**Der Nachweislauf danach** (85 Minuten, gleichbleibende Last, Beleg
`docs/belege/huellenprobe/dauerbetrieb.md`): 1.210 Runden, nichts verloren,
Anwendung und Webansicht je Abschnitt 748 · 782 · 750 MB (+0,4 % je Stunde).
**Der Anwendungsprozess allein wächst um 0,167 MB/min statt 0,29**, seine
Abschnitte stehen bei 80 · 83 · 88 MB — noch ein Anstieg über drei
Abschnitte. Was davon in diesem Code liegt, ist behoben; der Rest liegt in
den Bausteinen von macOS 27.0.1 und in der Go-Laufzeit 1.27.2.

**E06 ist damit abgeschlossen**
(#2278):
das Restleck des Webprozesses gibt es unter gleichbleibender Last nicht —
was die Läufe davor als solches zeigten, waren das falsche Maß (RSS),
fremde Prozesse in der Summe und eine Last, die mit jeder Runde wuchs.

#### #2527: die Probe im Gleichtakt — und was bei gleicher Probe bleibt

**Der Lauf der Gesamtabnahme** (`34ef8a1a4`, 2026-10-20) war nach der Regel
rot, bei **fallender** Tendenz: 826 · 806 · 832 · 873 · 739 MiB. Die Proben
nahm eine Uhr alle 30 s; eine Runde dauert rund fünf Sekunden, und die Probe
traf über lange Strecken denselben Teil — zehnmal hintereinander „ausgabe“.
Der Speicher hängt am Teil:

| Teil der Runde beim Abtasten (nach Minute 20) | Proben | Median (MiB) |
| --- | ---: | ---: |
| Tisch | 22 | 893 |
| Ausgabe (samt Pause) | 142 | 830 |
| Navigation | 38 | 740 |
| Bearbeiten | 10 | 676 |

**217 MiB zwischen den Teilen, dreimal so viel wie der Anstieg**, an dem die
Regel brach — die Abschnitte verglichen Teile statt Zeiten. Seitdem nimmt
`fahreDauer` die Probe **zwischen zwei Runden**, und `bewertung` erklärt
Proben aus verschiedenen Teilen für ungültig.

**Der Lauf mit gleichbleibender Probe** (`589bdf8e5`, 2026-10-20, 1.501
Runden, kein Absturz, nichts verloren) ist **rot, und deutlicher**:

| Abschnitt (Minute) | 20–40 | 40–60 | 60–80 | 80–100 | 100–120 |
| --- | ---: | ---: | ---: | ---: | ---: |
| zusammen | 507 | 540 | 573 | 588 | 590 |
| Webinhalt | 333 | 379 | 406 | 417 | 414 |
| Anwendung (Go) | 94 | 95 | 98 | 102 | 106 |
| GPU | 49 | 48 | 51 | 51 | 83 |

- **Der Webinhalt** wächst in der ersten Stunde um rund 84 MiB und danach
  langsamer — Steigung über alle Proben 0,90 MB/min, ab Minute 60 noch
  0,37. Ob er das wieder freigibt, belegt erst ein Lauf mit Speicherwarnung
  am Ende (`MSE_DAUER_WARNUNG=1`), wie bei E06.
- **Die Anwendung** wächst gleichmäßig um **0,166 MB/min** — dieselbe Rate
  wie nach #2312 (0,167), in den Bausteinen von macOS und der Go-Laufzeit.
  **Allein dieses Restwachstum bricht die Regel** („kein stetiger Anstieg
  über drei Abschnitte, auch wenn er klein ist“), sobald der Webinhalt ruht.
- **Die GPU** springt im Sägezahn zwischen 45 und 93 MiB.

**Der erste Lauf erlaubt wegen seiner Abtastung keine belastbare Aussage
über ein Wachstum; der zweite zeigt es bei gleichbleibender Probe.** Ein
kontrollierter Vergleich ist es nicht — zwischen beiden Ständen änderte sich
auch Code der Oberfläche (Nachlauf über das Gesamtreview, Befund 6). Die
Regel bleibt, wie sie ist; die Zusage für D6 ist
begrenzt (README, „Was diese Fassung nicht belegt“), und #2527 blieb offen —
bis zum Lauf unter Vermerk 0204 (nächster Abschnitt).

#### Vermerk 0204: die Regel, vor dem Lauf geändert — und der Lauf

**Entschieden vom Auftraggeber am 2026-10-20** (#2527),
festgelegt vor dem Lauf: eine Stunde Aufwärmen, ein Abschnitt zählt als
Anstieg erst ab +1 %, ≤ 10 % je Stunde, und nach Ruhe und Speicherwarnung
liegt WebKit unter seinem ersten Abschnitt. Der Lauf (`fe167f897`, 125
Minuten, 1.496 Runden) zeigt je Abschnitt (MiB, Median):

| | 0–20 | 20–40 | 40–60 | 60–80 | 80–100 | 100–120 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| zusammen | 519 | 512 | 520 | 557 | 586 | 598 |
| Webinhalt | 430 | 418 | 420 | 454 | 478 | 488 |
| Anwendung (Go) | 94 | 95 | 100 | 103 | 108 | 110 |

(Die Bewertung zählt ihre Abschnitte von der Aufwärmstunde an und kommt auf
557 · 590 · 601; die Tafel hier teilt ab Minute null.)

- **Die Regel hält nicht**: +11,8 % je Stunde, stetig.
- **Der Webinhalt wärmte nicht in der ersten Stunde auf** — anders als in
  den beiden Läufen davor; er wuchs in der zweiten um 0,65 MB/min.
- **Er gibt zurück**: nach zwei Minuten Ruhe 311 MiB, nach der Warnung 242
  MiB, unter dem Median seines ersten Abschnitts nach der Aufwärmstunde (455).
- **Die Anwendung** wächst wie zuvor um 0,15–0,17 MB/min (#2312).

**Was daraus folgt**: was wächst, ist ein Zwischenspeicher, den WebKit unter
Druck freigibt — aber **wann** er wächst, hängt am Lauf und nicht an einer
festen Aufwärmzeit. Die Grenzen bleiben, wie sie vor dem Lauf standen; D6
bleibt R3, und die Zusage ist begrenzt.

#### E06: eingegrenzt — und die Probe maß sich selbst mit

**Die Heap-Probe wuchs mit**: sie vergrößert den Puffer der Ressourcenzeiten
auf eine Million Einträge und liest ihn bei jeder Probe ganz, und jeder
IPC-Ruf der Hülle ist ein `fetch`. Mit ihr wuchs schon der Leerlauf (113 →
173 MB in 27 Minuten, `WebKit Malloc` laut vmmap); **ohne sie ist er flach** —
255 bis 262 MB ab Minute 15, der Fußabdruck laut vmmap 110,3 → 109,6 → 109,6
MB. Der Zwei-Stunden-Lauf hatte sie nie an; `dauerteile` misst seitdem ohne
sie (`MSE_DAUER_HEAP=aus`), und die Steigung zählt erst nach dem Einschwingen
(ab Minute 3 der Phase), der Gesamtfit steht in Klammern daneben — er schloss
den Einbruch beim Start ein (Befund 9 des Zweitblicks über E/F).

**Kein Teil trägt über seine Phase hinaus** (`f6bf769b0`, je zehn Minuten,
ohne Heap-Probe): nach Bearbeiten, Navigation, Tisch und Ausgabe steht der
Webprozess wieder bei 470 bis 486 MB. Am steilsten war die Navigation (4,7
MB/min); **allein über 30 Minuten** wuchs sie um 1,0 MB/min, mit Sägezahn bis
602 MB, die Anwendung um 0,3 (`97e7d5052`, Beleg
[`belege/huellenprobe/dauerteile.md`](belege/huellenprobe/dauerteile.md)).

**Und der Zwei-Stunden-Lauf am Endstand brach zuerst in seiner ersten Runde
ab**: der Klick auf das Textfeld traf »Vorschlagen« unter dem Namen. Das
Formular springt beim ersten Öffnen einer Karte um 118 Punkte, sobald die
Vorschau ihre Flächen kennt (Diagnose `kartenruhe`, Folgepaket
#2301); das
Drehbuch wartet seitdem auf die Vorschau.

**Kein Absturz und nichts verloren**: dieselbe Kennung nach 125 Minuten,
jede Runde fachlich gelungen, jede der 1.712 getippten Marken im Bestand.

**Der Zwei-Stunden-Lauf am Stand `098bfbc06`** (nach #2301, das Drehbuch
wartet auf die Vorschau; ohne Heap-Probe) — 125 Minuten, 1.721 Runden, kein
Absturz, nichts verloren, **aber rot**: 1.154 · 1.272 · 1.413 · 1.514 · 1.579 MB
je Abschnitt, +27,6 % je Stunde, stetig. **Anders als im Schlusslauf davor
wächst diesmal die Anwendung**: im Mittel je 25 Minuten 491 → 591 → 685 →
773 → 809 MB, ab Minute 10 um 3,05 MB/min und in der letzten halben Stunde
nur noch um 0,74 — sie flacht ab. Der Webprozess wächst um 2,05 MB/min, am
Ende um 3,26 (618 → 786 MB im Mittel der Abschnitte).

**Was das heißt, und was nicht**: der Schlusslauf an `1c0b74221` gab bei
Minute 50 Speicher an das System zurück, dieser nicht. Zwischen beiden liegen
die Portierungen aus F02 und das Drehbuch, das jetzt jede Vorschau abwartet,
statt in sie hineinzuklicken — es rastert also mehr fertig. **Ob der Go-Prozess
etwas hält oder nur seinen Heap nicht zurückgibt, ist nicht gemessen**: es gibt
am Fenster nur ein CPU-Profil (`leistungsprofil.go`), kein Heap-Profil. Die
Maskenwege aus #2299 halten nichts fest (kein Speicher in `zeilenmaske.go`,
`kastenfilter.go`). E06 bleibt offen; der nächste Schritt ist ein Heap-Abzug
des Go-Prozesses in der Abnahmehülle zu Beginn, in der Mitte und am Ende
eines Laufs (#2278).

**Kalt heißt hier**: frisches Zuhause und frischer Bestand, aber der
Dateicache des Systems bleibt warm — `purge` braucht ein Kennwort. Deshalb
liegen kalt und warm beieinander.

## 10 Der Messvertrag von 0.9.0 (PERF01)

**#2351,
Befunde A-13, A-14, A-25 und A-26.** Bis hierher hatten nur Start, Vorschau
und Rollen ein Budget, und für Speicher gab es nur ein Wachstumsziel. **Jetzt
hat jeder Kernweg eine Grenze für Zeit und Speicher**, festgelegt vor der
ersten Messung und **durchgesetzt**: eine gerissene Grenze macht das Drehbuch
rot, und der Belegplan (`node scripts/belegplan.mjs`) liest dieselbe Zeile.

**Die Tafel steht in `scripts/messvertrag.mjs`**, und nur dort; aus ihr
folgen die Zeile im Beleg, das Urteil und die Einträge in
`docs/belege/BELEGPLAN.json`. Ein Test hält Tafel und Plan gleich
(`scripts/messvertrag.test.mjs`). **Eine gescheiterte Messung ist kein Wert**:
eine Reihe mit einer −1 heißt „gescheitert", nicht „gehalten".

| Reihe | Grenze | Gemessen (Median · P95 · max) | Beleg |
| --- | --- | --- | --- |
| Start kalt | P95 ≤ 2 s | 510 · 1.154 · 1.154 ms | `huellenprobe/budgets09.md` |
| Start warm | P95 ≤ 2 s | 515 · 550 · 550 ms | dito |
| Set mit 3.788 Karten öffnen | P95 ≤ 1 s | 255 · 708 · 708 ms | dito |
| Feld → Vorschaubild | P95 ≤ 300 ms | 62 · 106 · 147 ms | dito |
| Sammlung mit 25.000 Zeilen durchsuchen | P95 ≤ 500 ms | 120 · 155 · 167 ms | dito |
| 60 Karten als Bilder ausgeben | ≤ 60 s | 5,96 · — · 7,12 s | dito |
| Druckbogen aus 60 Karten | ≤ 60 s | 7,70 · — · 7,91 s | dito |
| Zurückspielen eines Bestands mit 25.000 Zeilen | ≤ 60 s | 0,74 · — · 1,13 s | dito |
| Fußabdruck nach jedem Weg (Anwendung und Webansicht) | ≤ 1,5 GiB | 1.033 · — · **1.363** MiB | dito |
| Bündel, logisch | ≤ 500 MiB | 260,9 MiB | dito |
| Leerlauf, je Prozess und Zustand | ≤ 1 % eines Kerns | höchstens 0,21 % (Anwendung, sichtbar) | `huellenprobe/leerlauf09.md` |
| Fußabdruck nach 120 Minuten | ≤ 1,5 GiB | 815 · 959 · 959 MiB | `huellenprobe/dauerbetrieb.md` |
| Dienst nach Rendern und Ruhe | ≤ 512 MiB | `mse-rest` 31, `mse-mcp` 48 MiB, mit `GOMEMLIMIT` dasselbe | `dienstspeicher.md` |

**Gemessen am 2026-10-15** auf Mac17,3 (Apple M5, zehn Kerne, 32 GiB), macOS
27.0.1, Abnahmebündel `627cab0dae91`; `budgets09` am Commit `73c164bda`,
`leerlauf09` an `43be67a02`, der Dauerbetrieb an `c3e45a2f4`, der
Dienstspeicher an `1316ec35a`. **Der Dauerbetrieb lief am 2026-10-20 erneut**
(`589bdf8e5`, Probe zwischen zwei Runden): nach 120 Minuten 596 · 789 · 789
MiB, gehalten (§9.1). **Die Bestände sind künstlich**
(`scripts/huellenprobe/budgetsorte09.mjs`): 3.788 erfundene Karten mit dem
Stil M15 aus dem lokalen Pack, ein Druckset mit 60 Karten, eine Sammlung mit
25.000 Zeilen, eine Sicherung des Ganzen.

### Was die Zahlen sagen

- **Der Fußabdruck hält, aber knapp**: das Set mit 3.788 Karten zwanzigmal zu
  öffnen hebt Anwendung und Webansicht auf 1.363 MiB, 89 % der Grenze; danach
  fällt er auf 955 MiB. **Wem davon wie viel gehört, trennt die Reihe nicht**
  — sie nennt die Summe; der Dauerbetrieb trennt, und dort trägt die
  Webansicht den größeren Teil.
- **Der erste Kaltstart ist der langsamste** (1.154 ms gegen sonst 427 bis
  519 ms): der erste Start des eben gebauten Bündels, wie in B02 (§9).
- **Der Leerlauf ist still**: sichtbar 0,21 % eines Kerns für die Anwendung und
  0,11 % für den Webinhalt; ausgeblendet und minimiert unter 0,06 %, und die
  Seite meldet sich `hidden` — die Abfrage des Datenstands schweigt dann
  (`useBestandsstand`).
- **Der Dienst gibt zurück, was er unter Rasterlast belegt**: über 10.000
  Bilder in 30 Sekunden, höchstens 44 MiB Fußabdruck, nach zwei Minuten Ruhe
  31 MiB. Die 3,7 GiB aus dem Lastlauf vom 2026-10-02 waren **RSS**, das unter
  macOS zurückgegebene Seiten mitzählt (§9.1); der Fußabdruck sagt, was das
  System wirklich belegt sieht. **`GOMEMLIMIT` ändert hier nichts**, und es
  wird deshalb nicht gesetzt.
- **Rot bleibt der Dauerbetrieb nach B04**: 5 % je Stunde, aber stetig über die
  Abschnitte. Die Anwendung wächst um 0,18 MB/min, die Webansicht um 0,56.
  Das ist #2312 (E2312), und der Belegplan führt es rot, bis es behoben ist.

### Die Mindestausstattung

**Gemessen ist allein dieser Rechner mit 32 GiB.** Der Lauf in einem Gast mit
8 GiB, den das Issue verlangt, braucht ein Wiederherstellungsabbild
(`UniversalMac_27.0.1_26A434_Restore.ipsw`, 26,6 GB), dessen Download nicht
freigegeben ist — **die Mindestausstattung ist deshalb eine benannte
Hardwaregrenze und kein Beleg**: zugesagt wird, was auf Mac17,3 mit 32 GiB
gemessen ist. Abgeleitet, nicht gemessen: Anwendung und Webansicht belegen
im Arbeitsgang höchstens rund 1,4 GiB, ein Dienst unter 64 MiB.

### Was sie nicht sagen

- **Einen anderen Mac**, und einen ersten Start nach der Installation aus
  dem Abbild — das ist Handprüfung 6.
- **Weckungen und Energie** — `powermetrics` braucht Verwalterrechte.
- **Ein Fenster, das ein anderes verdeckt**: gemessen ist ausgeblendet (⌘H),
  in dem WebKit die Seite ebenso verborgen meldet.

## 11 Kennzahlen und die Prüfung eines ganzen Sets (PERF02)

**#2352,
Befunde A-14, B-13, B-14, B-15 und Z-03.** Billiges Zählen und volle Prüfung
sind zwei Dinge: die Startseite zählt über SQL, welche Karten ein Problem
haben, sagt erst das Zeichnen jeder Karte. Die Prüfung läuft deshalb nie je
Aufruf der Startseite, sondern auf Anstoß (`mse probleme`, MCP, `POST
/sets/{id}/probleme`), und ihr Stand sagt, wie weit er gilt.

**Bis zur sichtbaren Kennzahl**, gemessen am 2026-10-15 auf Mac17,3
(`huellenprobe/kennzahlen09.md`, Bündel `ca14154c0440`): je Bestand 30 Starts
kalt (ein frisch geklontes Zuhause), 30 warm und 30 Aufrufe der Startseite.
Gezählt wird bis zur Marke `mse:kennzahlen`, die der Überblick setzt, wenn er
mit seinen Zahlen erscheint — beim Start seit dem Laden der Seite, beim Aufruf
seit dem Augenblick vor ⌘1.

| Weg | 5 Karten | 3.788 Karten | 25.000 Karten in 50 Sets | Grenze |
| --- | --- | --- | --- | --- |
| nach dem Start, kalt (Median · P95) | 157 · 169 ms | 160 · 175 ms | 163 · 174 ms | P95 ≤ 500 ms |
| nach dem Start, warm | 166 · 174 ms | 168 · 176 ms | 170 · 180 ms | dito |
| beim Aufruf der Startseite | 5 · 8 ms | 5 · 7 ms | 5 · 7 ms | dito |
| Öffnen über LaunchServices bis Kennzahl, kalt (ohne Grenze) | 625 · 644 ms | 612 · 636 ms | 625 · 648 ms | — |
| Fußabdruck an der Kennzahl, kalt (Median · max) | 449 · 505 MiB | 445 · 503 MiB | 432 · 496 MiB | — |
| Bindungsrufe bis zur Kennzahl | 25 | 25 | 25 | — |

**Die Zählung wächst mit dem Bestand nicht** — sie ist SQL über Indizes.
**Ein Aufruf der Startseite kostet fast nichts**, weil eine besuchte Seite
verborgen stehen bleibt (`<Activity>`, UX-03.AK-02) und mit ihren Zahlen
wieder erscheint. **Kein Aufruf zeichnet oder prüft**: das Rufbuch nennt
dabei elf Bindungen, keine davon rastert.

**Die Prüfung eines ganzen Sets** hat die Grenze **60 s für 3.788 Karten**:

| Was | Gemessen | Wo |
| --- | ---: | --- |
| `Black`, 3.788 Karten, Schriften des Packs | **51 s** | `BenchmarkWegProblempruefung` |
| 3.788 erfundene Karten mit M15, über `mse probleme` | **27,6 s** | `huellenprobe/kennzahlen09.md` |

**Erst 131 s.** Ein Profil fand drei Ursachen, und jede ist behoben: zu wenige
Arbeiter für einen Plan, der nicht hinter der Pforte des Rasterers steht; das
Schloss der Schriftsuche, auf das die Arbeiter ein Drittel ihrer Zeit warteten;
und ein Bildspeicher, der bei 64 Einträgen räumte, während `Black` 255
verschiedene Bilder braucht. **Die zweite und die dritte treffen jede
Vorschau.** Der Weg der Zahlen steht in
`audits/0.9.0/umsetzung/PERF02-2352.md` (Entwicklungsarchiv).

### Was sie nicht sagen

- **Die Problemzahl am Fenster** — sie steht seit AP-8
  (#2365)
  auf den Kacheln; die Startseite **liest** dafür die Stände aller Sets
  (`Setprobleme`) und prüft nie. Was das Lesen bis zur Kennzahl kostet, steht
  in der Nachmessung darunter.
- **Einen Mac mit weniger Kernen**: die Prüfung nimmt die Hälfte der Kerne.
- **Ob ein Aufruf der Startseite frische Zahlen zeigt, wenn sich dazwischen
  etwas änderte** — der Lauf änderte nichts.

### Nachgemessen mit den Prüfständen am Fenster (AP-8)

**Am 2026-10-17 auf Mac17,3** (`huellenprobe/kennzahlen09.md`, Commit
`3c70b3eb3`), nachdem die Startseite die Prüfstände aller Sets liest
(#2365).
**Der erste Schnitt riss die Grenze**: `Setprobleme` stand im selben
Ladelauf wie die Zahlen, und die Kennzahl kam bei einem Set nach 328 ms, bei
50 Sets kalt nach 526 ms (P95). Seitdem lädt es für sich, und die
Unterzeilen füllen sich danach:

| Weg (Median · P95) | 5 Karten | 3.788 Karten | 25.000 Karten in 50 Sets | Grenze |
| --- | --- | --- | --- | --- |
| nach dem Start, kalt | 167 · 188 ms | 169 · 189 ms | 183 · 200 ms | P95 ≤ 500 ms |
| nach dem Start, warm | 192 · 221 ms | 230 · 248 ms | 216 · 225 ms | dito |
| beim Aufruf der Startseite | 7 · 11 ms | 7 · 30 ms | 7 · 25 ms | dito |

**Warm ist es langsamer als am 2026-10-15** (dort 166 bis 180 ms): das Lesen
der Stände läuft neben den Zahlen und nimmt sich Zeit auf derselben Naht.
Gehalten ist die Grenze überall; **wie lange die Unterzeilen brauchen**, misst
dieser Lauf nicht. Kein Aufruf zeichnet oder prüft — unter den Rufen steht
`Setprobleme`, das liest.

**Und noch einmal nach Teil 2b** (Commit `d0648c232`, derselbe Tag): die
Startseite liest dazu das Portfolio und fragt die Aufmerksamkeit eine Ebene
höher, beides für sich.

| Weg (Median · P95) | 5 Karten | 3.788 Karten | 25.000 Karten in 50 Sets | Grenze |
| --- | --- | --- | --- | --- |
| nach dem Start, kalt | 188 · 201 ms | 198 · 225 ms | 189 · 210 ms | P95 ≤ 500 ms |
| nach dem Start, warm | 247 · 274 ms | 260 · 279 ms | 207 · 229 ms | dito |
| beim Aufruf der Startseite | 8 · 23 ms | 9 · 23 ms | 6 · 16 ms | dito |

**Kalt und warm kosten zwei Rufe mehr auf derselben Naht 10 bis 50 ms**;
gehalten ist die Grenze überall. Die Sammlung der Bestände in diesem Lauf
ist leer — **wie lange das Portfolio über eine große Sammlung rechnet**, misst
er nicht, und die Zahlen warten nicht darauf.
