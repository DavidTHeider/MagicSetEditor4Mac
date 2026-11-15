# Was der Spieltisch kostet

**Werkstück G3 von Phase 5.5**
(#1460).

#1460 verlangt für die Leistungsszene: „4 Spielern, 150 Karten, 100 Countern,
50 Tokens, 30 Beziehungsgrafiken, Hintergrund, Animation und aktivem Drag
messen: 60 FPS als Ziel mit benanntem Mac/macOS, Auflösung, Dauer, Warm-/
Kaltcache und Framezeitverteilung; keine reine FPS-Mittelwertbehauptung."

**Diese Datei löst davon die eine Hälfte ein und sagt für die andere, dass sie
offen ist.** Welche Hälfte welche ist, steht unten — und zwar zuerst, damit
niemand die Zahlen für mehr nimmt, als sie sind.

---

## Was gemessen ist und was nicht

**Und wo**, denn das ist nicht dasselbe — PLAY-06.AK-01
(#1714):
„Native Wails-WebView, Browservorschau und synthetische Planberechnung werden
getrennt ausgewiesen; ein schneller CPU-Tischplan belegt keine
GPU-Framezeit."

|                                                       | Zustand                                                                  | Abschnitt |
| ----------------------------------------------------- | ------------------------------------------------------------------------ | --- |
| **Die Rechnung** — Partie → Plan → Sprites → Abgleich | **gemessen am 2026-09-16**, synthetisch in Node, ohne Grafikkarte        | Die Rechnung |
| **Die Rasterung** — was PixiJS aus den Sprites macht  | **gemessen am 2026-09-16**, in **headless Chrome** (ANGLE über Metal)    | Die Rasterung |
| **Warm- und Kaltcache der Texturen**                  | **gemessen am 2026-09-16**, dieselbe Messung — das erste Bild zählt gesondert | Die Rasterung |
| **Die Rasterung in WebKit** — dieselbe Messseite im WKWebView wie in Wails | **gemessen am 2026-09-24**, in der WebKit-Probe (#1828) | Die Rasterung in WebKit |
| **Framezeitverteilung im Wails-Fenster**              | **offen** — gemessen ist WebKit, nicht das Fenster; was dazwischen liegt, steht unten | — |
| **Aktiver Drag während der Messung**                  | **gemessen am 2026-09-25**, in WebKit — `TISCHART=zug`, derselbe Planweg wie beim Ziehen am Tisch | Ziehen, Animation, Speicher und Dauer |
| **Animation während der Messung**                     | **gemessen am 2026-09-25**, in WebKit — `TISCHART=animation`, Flüge von 220 ms | Ziehen, Animation, Speicher und Dauer |
| **Speicher der WebKit-Prozesse**                      | **gemessen am 2026-09-25**, jede Sekunde über `ps`, nur die Prozesse der Messung | Ziehen, Animation, Speicher und Dauer |
| **Eine lange Sitzung**                                | **gemessen am 2026-09-25**, drei Minuten Animation, in Dritteln ausgewertet | Ziehen, Animation, Speicher und Dauer |
| **Kontextverlust, Wiederherstellung, Laden, Größe**   | **gemessen am 2026-09-25**, am echten Tisch in der WebKit-Probe | Kontextverlust, Laden und Größe |
| **Externes Display, Vollbild**                        | **offen** — sie brauchen einen zweiten Bildschirm ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)) | — |

> **Hier stand bis zum 2026-09-22 ein Widerspruch**, und er ist der Anlass von
> PLAY-06.AK-05. Die Tabelle führte **die Rasterung**, **den Warmcache** und
> **die Framezeitverteilung** als offen — während weiter unten der Abschnitt
> „Die Rasterung — erhoben am 2026-09-16" mit Zahlen für alle drei stand. Wer
> nur die Tabelle las, hielt eine erhobene Zahl für fehlend; wer nur den
> Abschnitt las, hielt eine offene Frage für beantwortet.
>
> **Und beim Drag stand „es gibt ihn noch nicht".** Es gibt ihn seit dem
> 2026-09-21.
>
> Seitdem hält
> [`internal/werkzeuge/tischleistung_test.go`](../internal/werkzeuge/tischleistung_test.go)
> die beiden Stellen zusammen: ein Zustand ist **offen** oder **gemessen am
> JJJJ-MM-TT**, und im zweiten Fall steht **der genannte Abschnitt** unterhalb
> mit demselben Datum. Bis zum 2026-09-25 genügte das Datum irgendwo darunter
> — fünf Zeilen tragen dasselbe, und der Abschnitt einer davon hätte fehlen
> können (Befund 5 des Zweitblicks über Phase 6.9). Was er **nicht** hält,
> steht in seinem Kopf.

**Warum die Framezeit am Fenster offen bleibt:** sie braucht ein echtes
Wails-Fenster mit stabilem WebGL-Kontext. Die **Browservorschau** unter
`wails dev` hat ihn nicht — sie verliert ihn (gemessen am 2026-09-16: „Could
not retrieve shader source", und PixiJS scheiterte, bis der Ticker
abgeschaltet war). Deshalb misst `scripts/tischframezeit.mjs` in **headless
Chrome**: dieselbe Rendermaschinerie, eine andere Laufzeit als die WebView.

**Was das kostet, gehört ausgesprochen**: Chrome und WKWebView teilen sich
weder Fassung noch Treiberweg. Die Zahlen aus Chrome sagen, was **diese Szene
auf dieser Maschine** kostet — sie sagen nicht, was das Fenster kostet.

**Seit dem 2026-09-24 misst dasselbe Skript auch in WebKit**
(`TISCHBROWSER=webkit`, #1828):
dieselbe Messseite in der WebKit-Probe — ein WKWebView mit den Einstellungen,
die Wails 3 setzt, und dem WebKit dieses Rechners. **Das ist die Laufzeit des
Fensters, nicht das Fenster**: die Probe liegt außerhalb des Bildschirms, ihr
Fenster gehört keinem Go-Prozess, und eine Brücke gibt es dort nicht. Für die
Zeichenzeit ist das gleichgültig — sie ist die Arbeit von PixiJS in WebKit.
Für den Bildabstand nicht ganz: ob ein Fenster auf einem Display mit höherer
Wiederholrate mehr Bilder bekommt, zeigt diese Messung nicht.

**Eine Zahl, die man nicht erheben kann, wird nicht geschätzt** (§10). Sie
steht hier als offen, und das Issue bleibt es mit ihr.

---

## Die Rechnung — erhoben am 2026-09-16

```bash
node scripts/tischmass.mjs
```

**Maschine:** Mac17,3, 10 Kerne, macOS 26.6.2, Node v24.15.0.
**Szene:** `extreme-board-state` — 4 Sitze, 330 Objekte, 262 Sprites.
**Läufe:** 200 je Messung, 20 vorab zum Warmlaufen.

| Was                |    Median |       p95 |       p99 |       max |
| ------------------ | --------: | --------: | --------: | --------: |
| `tischplan()`      |     0,154 |     0,191 |     0,405 |     0,595 |
| `kartenbilder()`   |     0,013 |     0,040 |     0,107 |     0,660 |
| `abgeglichen()`    |     0,010 |     0,014 |     0,023 |     0,030 |
| **alles zusammen** | **0,171** | **0,201** | **0,246** | **0,323** |

Alle Zeiten in Millisekunden. **Ein Bild bei 60 Hz hat 16,667 ms.**

**Was daraus folgt:** die Rechnung braucht im Median **1,0 %** eines Bildes und
im schlechtesten von 200 Läufen **1,9 %**. Sie ist nicht der Engpass, und wenn
der Tisch ruckelt, liegt es nicht an ihr.

**Verteilung statt Mittelwert**, weil das Issue es verlangt und weil ein
Mittelwert genau den Ausreißer verbirgt, den ein Auge als Ruckeln sieht.

### Nachgemessen zur Freigabe 0.6.0 — 2026-10-05

Dieselbe Szene, dieselben Läufe, macOS 27.0.1, Node v24.15.0 — **bei einer
Last von 3,6** aus Systemdiensten und der Claude-App, die nicht anzuhalten
waren:

| Was                |    Median |       p95 |       p99 |       max |
| ------------------ | --------: | --------: | --------: | --------: |
| `tischplan()`      |     0,186 |     0,269 |     0,354 |     0,574 |
| `kartenbilder()`   |     0,003 |     0,009 |     0,050 |     0,054 |
| `abgeglichen()`    |     0,008 |     0,010 |     0,040 |     0,046 |
| **alles zusammen** | **0,185** | **0,257** | **0,466** | **0,611** |

**1,1 % eines Bildes im Median, 3,7 % im schlechtesten Lauf.** Gegen den
2026-09-16 ein Hauch mehr — innerhalb dessen, was die Last erklärt; der
Schluss oben gilt unverändert.

---

## Was die Zahlen nicht sagen

- **Nichts über die Grafikkarte.** Siehe oben.
- **Nichts über eine andere Maschine.** Eine Messung, und eine Maschine
  (Vermerk [0007](entscheidungen/0007-kein-entfernter-lauf.md)).
- **Wenig über den Dauerbetrieb.** Gemessen ist dieselbe Szene zweihundert
  Mal und — seit dem 2026-09-25 — drei Minuten Animation am Stück (unten);
  eine Partie über eine Stunde mit wachsendem Friedhof ist etwas anderes.
- **Den Speicher nur als Prozessspeicher.** Wie viele Texturen ein Tisch hält
  und wann das Fach in Go sie räumt, steht in G2
  (#1459);
  gemessen ist hier, was Web- und GPU-Prozess von WebKit belegen.

---

## Die Szene selbst

`frontend/src/test/tischszenen.ts` führt sie, und die Zahlen werden
**gezählt** statt geschätzt: `frontend/src/tisch/szenen.test.ts` hält 150
Karten, 100 Marken, 50 Spielsteine und 30 Anhängsel nach. Eine Leistungsszene,
die um ein Drittel danebenläge, machte jede Messung darauf unvergleichbar.

Daneben stehen neun weitere Szenen — `empty-board`, `two-player-simple`,
`creature-combat`, `large-land-piles`, `stack-heavy`, `tokens-and-counters`,
`four-player-commander`, `double-faced-cards` und `many-attachments`. Sie sind
die Regressionsprobe des Tischplans, nicht seine Leistungsprobe.

**Sie haben beim ersten Lauf einen Fehler gefunden**: `heimatlos` zählte jedes
Objekt ohne Lage, und eine Bibliothek mit 53 Karten meldete damit 45 Befunde —
ein Stapel zeichnet nur seine obersten acht. Eine Diagnose, die bei jeder
gewöhnlichen Partie dreistellig ausfällt, liest niemand mehr.

---

## Die Rasterung — erhoben am 2026-09-16

```bash
node scripts/tischframezeit.mjs
```

**Maschine:** Mac17,3, macOS 26.6.2.
**Renderer:** ANGLE (Apple, ANGLE Metal Renderer: **Apple M5**).
**Fläche:** 1440 × 813 Punkte.
**Szene:** `extreme-board-state`, 240 Bilder.

### Bildabstand — hält es 60 Hz?

|                           |                   ms |
| ------------------------- | -------------------: |
| erstes Bild, kalt         |                17,80 |
| Median                    |                16,60 |
| p95                       |                18,30 |
| p99                       |                18,70 |
| Maximum                   |                18,90 |
| **über 33,33 ms (30 Hz)** | **0,0 % der Bilder** |

**Ja.** Der Median liegt auf dem VSync-Takt (16,667 ms), und kein einziges von
240 Bildern fiel unter 30 Hz.

### Zeichenzeit — wie viel Luft bleibt?

|                   |    ms |
| ----------------- | ----: |
| erstes Bild, kalt | 17,70 |
| Median            |  0,40 |
| p95               |  0,60 |
| p99               |  1,30 |
| Maximum           |  3,40 |

**Viel.** Das Zeichnen braucht **2,4 %** eines Bildes im Median und **20,4 %**
im schlechtesten von 240 Läufen.

**Warum zwei Zahlen und nicht eine:** bei VSync liegt der _Bildabstand_ immer
nahe 16,67 ms, gleich wie wenig Arbeit anfällt — er sagt, ob die Bildrate
gehalten wird, und sonst nichts. Die _Zeichenzeit_ sagt, ob das mit Mühe oder
mit Leichtigkeit geschieht. Wer nur die erste angäbe, hätte eine Zahl ohne
Aussage über die Reserve.

**Das erste Bild zählt gesondert** und ist die Antwort auf „Warm-/Kaltcache":
es bezahlt den Aufbau jedes Sprites und jedes Shaders — 17,70 ms Zeichenzeit
gegen 0,40 ms danach, also das **vierundvierzigfache**. Ein Tisch, der neu
aufgeht, lässt genau ein Bild aus; danach ist er warm.

## Die Rasterung in WebKit — erhoben am 2026-09-24

```bash
TISCHBROWSER=webkit node scripts/tischframezeit.mjs
```

**Maschine:** Mac17,3 (Apple M5), macOS 27.0, WebKit 22625.1.29.11.27.
**Renderer:** „Apple GPU" — mehr nennt WebKit nicht; der Chip steht deshalb
eine Zeile darüber und nicht aus der Messseite.
**Fläche:** 1440 × 900 Punkte bei **doppelter** Dichte, also 2880 × 1800
Pixel — gut viermal so viele wie in Chrome, das mit einfacher Dichte misst.
**Szene:** `extreme-board-state`, 240 Bilder.

**Alle Werte sind ganze Millisekunden.** WebKit gibt `performance.now()` hier
nur auf eine Millisekunde genau heraus — beobachtet, nicht angenommen: jeder
der 480 Werte dieses Laufs ist ganzzahlig. Eine Zeichenzeit von „2" heißt
deshalb *zwischen eins und drei*, und der Anteil an einem Bild darunter ist
auf sechs Prozentpunkte genau und nicht auf einen.

### Bildabstand in WebKit

|                           |                   ms |
| ------------------------- | -------------------: |
| erstes Bild, kalt         |                   71 |
| Median                    |                   17 |
| p95                       |                   18 |
| p99                       |                   19 |
| Maximum                   |                   21 |
| **über 33,33 ms (30 Hz)** | **0,0 % der Bilder** |

**Ja, auch hier.** 195 von 239 Abständen sind 16 oder 17 ms — der VSync-Takt
von 16,667 ms, auf ganze Millisekunden gerundet.

### Zeichenzeit in WebKit

|                   |   ms |
| ----------------- | ---: |
| erstes Bild, kalt |   71 |
| Median            |    2 |
| p95               |    3 |
| p99               |    3 |
| Maximum           |    7 |

**Genug.** Das Zeichnen braucht **12 %** eines Bildes im Median und **42 %**
im schlechtesten von 239 Läufen — bei gut viermal so vielen Pixeln wie in
Chrome.
Das **erste** Bild kostet 71 ms, gut das Vierfache von Chrome — woran es
liegt, zeigt diese Messung nicht. Ein Tisch, der neu aufgeht, lässt in WebKit
**vier** Bilder aus statt eines.

> **Die erste Messung am selben Tag hatte eine Fläche von 720 × 450 Punkten**
> — ein Viertel der Leinwand. Das war kein Messfehler, sondern ein Fehler des
> Tischs: `pixinaht.ts` teilte die Fläche durch die Dichte, die PixiJS 8 schon
> herausgerechnet hat, und auf jedem Retina-Bildschirm lag der Tisch im linken
> oberen Viertel seiner Leinwand. In Chrome mit einfacher Dichte teilt die
> Rechnung durch eins; gesehen hat es erst WebKit mit doppelter. Seitdem steht
> sie in `zeichenflaeche.ts` und wird gemessen (Zeile **Retina** unten).

---

## Ziehen, Animation, Speicher und Dauer — erhoben am 2026-09-25

PLAY-06.AK-02 und AK-03 (#1714).

```bash
TISCHBROWSER=webkit TISCHART=ruhe|zug|animation node scripts/tischframezeit.mjs
TISCHBROWSER=webkit TISCHART=animation TISCHBILDER=10800 node scripts/tischframezeit.mjs
```

**Maschine:** Mac17,3, macOS 27.0, Apple GPU. **Laufzeit:** die WebKit-Probe
— ein WKWebView, eingerichtet wie in Wails, 1.440 × 900 Punkte bei doppelter
Dichte. **Szene:** `extreme-board-state`, 240 Bilder je Art.

| Art | erstes Bild, kalt | Bildabstand Median | p99 | Maximum | unter 30 Hz | Zeichenzeit Median | p99 | Webprozess | GPU-Prozess |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| ruhend | 87 ms | 17 ms | 19 ms | 19 ms | 0 % | 1 ms | 2 ms | 136–141 MB | 37 MB |
| **mit Ziehen** | 87 ms | 17 ms | 19 ms | 19 ms | 0 % | 1 ms | 2 ms | 135–145 MB | 37 MB |
| **mit Animation** | 95 ms | 17 ms | 19 ms | 22 ms | 0 % | 2 ms je Sprung | 5 ms | 107–123 MB | 36–37 MB |

**Ziehen** heißt: jedes Bild bekommt einen neuen Plan aus `mitZug`, derselbe
Weg wie am Tisch (`Tischflaeche.tsx`); nachgestellt ist nur der Zeiger.
**Animation** heißt: eine Karte springt alle zwanzig Bilder, und der Maler
lässt sie fliegen — **gezeichnet wird nur beim Sprung**, den Flug von 220 ms
rendert der Ticker bis zur Landung, so wie am Tisch. Die Zeichenzeit dieser
Zeile zählt deshalb die Sprünge (elf in 240 Bildern), der Flug steht im
Bildabstand.

> **Die Zeile stand hier mit anderen Zahlen, und sie maßen etwas anderes.**
> Bis zum 2026-09-25 rief die Messseite `zeichne` in **jedem** Bild, und jeder
> Ruf setzte den laufenden Flug neu an — kein Flug lief je zu Ende. Gemessen
> war eine **Zeichenlast**, und kein Flug, der landet (Befund 4 des
> Zweitblicks über Phase 6.9). Wer sie wiederholen will: `TISCHALT=1`. **Kalt- und Warmcache** trennt das erste Bild: es bezahlt
Sprites und Shader, jedes weitere nicht. WebKit rundet `performance.now()` auf
ganze Millisekunden; die Zahlen sind deshalb ganze.

**Speicher**: jede Sekunde über `ps`, und **nur die Prozesse, die während der
Messung entstanden** — der erste Anlauf zählte alle WebKit-Prozesse des
Rechners mit, auch fremder Programme, und kam auf 452 MB.

### Die lange Sitzung

**Drei Minuten Animation am Stück**, 10.799 Bilder nach dem ersten
(`TISCHBILDER=10800`), in Dritteln ausgewertet — wächst etwas mit der Zeit,
wird das letzte Drittel langsamer als das erste. **Neu erhoben am
2026-09-25** mit Flügen, die zu Ende laufen:

| Drittel | Bildabstand Median | p99 |
| --- | ---: | ---: |
| erstes | 17 ms | 19 ms |
| zweites | 17 ms | 19 ms |
| letztes | 17 ms | 19 ms |

**Über alles**: Maximum 36 ms — **mindestens ein Bild** über 33,33 ms, also
unter 30 Hz; die Auswertung zeigte den Anteil damals gerundet als „0,0 %" und
damit weniger als 0,05 % der 10.799 Bilder. Zeichenzeit je Sprung im Median
1 ms, im schlechtesten 5 ms, über 539 Sprünge. *Hier stand zweimal „kein Bild
unter 30 Hz": für den ersten Lauf bei 49 ms (Befund 6 des Zweitblicks über
Phase 6.9) und für diesen bei 36 ms (Befund 4 des Nachlaufs). Beide Male lag
das Maximum darüber, und beide Male stand die gerundete Zahl als „keines" da.
Das Skript nennt seitdem die **Zahl** der Bilder, nicht nur ihren Anteil.*

**Der Webprozess wuchs von 108 auf 403 MB**, fast gerade; der GPU-Prozess
blieb bei 36–37 MB. **Das ist neu**: der erste Lauf, der in jedem Bild
zeichnete, kam auf 132 → 158 MB. Die Gegenproben desselben Abends grenzen es
ein (#1852):

| Lauf | Dauer | Webprozess |
| --- | ---: | --- |
| Animation, drei Sprünge je Sekunde | 90 s | 108 → 263 MB, stetig |
| dieselbe ohne Flug (`TISCHBEWEGT=0`) | 45 s | 106 → 160 MB |
| **derselbe Plan**, drei Mal je Sekunde gezeichnet (`TISCHGLEICH=1`) | 70 s | 109 → 179 MB, noch steigend |
| derselbe Plan, **einmal je Sekunde** (`TISCHSPRUNG=60`) | 70 s | 105 → 135 MB, **eingependelt** |
| in **jedem** Bild gezeichnet (`TISCHALT=1`) | 70 s | 134 → 149 MB, **eingependelt** |
| ruhender Tisch, jedes Bild | 90 s | 134 → 150 MB, **eingependelt** |

**Die Gegenproben sprachen gegen den Flug und gegen den neuen Plan und für
den Takt des Zeichnens** — derselbe Plan wuchs bei drei Zeichnungen je
Sekunde und pendelte sich bei einer und bei sechzig ein. Ob das ein Leck war
oder Speicher, den JavaScriptCore später zurückgibt, sagten diese Läufe
nicht; sie dauerten höchstens drei Minuten.

**Zehn Minuten sagen es: es pendelt sich ein**
(#1852,
AK-03). Gemessen am 2026-09-25 und 2026-09-26 auf derselben Maschine, in der
WebKit-Probe, je 36.000 Bilder nach dem ersten; der Webprozess in gleichen
Abständen über die ganze Dauer:

| Lauf | Aufruf | Webprozess (RSS) | GPU-Prozess |
| --- | --- | --- | ---: |
| **derselbe Plan**, drei Mal je Sekunde | `TISCHGLEICH=1 TISCHBILDER=36000` | 132 → 135 → 136 → 139 → 140 → 141 → 141 MB, **höchstens 141** | 38 MB |
| **Animation**, wie oben | `TISCHBILDER=36000` | 138 → 185 → 186 → 237 → 237 → 238 → 238 MB, **höchstens 238** | 38 MB |

**Beide pendeln sich ein.** Derselbe Plan blieb über zehn Minuten und 1.799
Sprünge unter dem, was der 70-Sekunden-Lauf mit demselben Aufruf erreicht
hatte (179 MB). Die Animation stieg in zwei Stufen und stand die letzten
vier Minuten bei 237–238 MB — der Drei-Minuten-Lauf vom Vortag war in
derselben Zeit bei 403 MB. **Ein Leck je Zeichnung sähe anders aus**: es
wüchse mit jedem der 1.799 Sprünge, auch in den letzten vier Minuten. Warum
derselbe Aufruf an zwei Tagen verschieden weit stieg, sagt keiner der Läufe;
wann JavaScriptCore freigibt, hängt an mehr als an diesem Programm. Einen
Heap-Abzug gibt es weiterhin nicht — die Abnahme verlangt ihn nur für den
Fall, dass es eines ist. Der Bildabstand blieb im Median bei 17 ms (p99
22–23 ms); in der Animation lag kein Bild über 33,33 ms, beim selben Plan
eines.

**Der Tisch zeichnet im Spiel auf Zuruf**, einen Flug je Zug; drei Sprünge
je Sekunde über Minuten sind eine Prüflast. **60 Bilder je Sekunde bleiben
ein Ziel auf dieser Maschine**, keine Zusage für jede (AK-03).

## Kontextverlust, Laden und Größe — erhoben am 2026-09-25

Am **echten Tisch** in der WebKit-Probe, Belege in `docs/belege/webkitprobe/`:

| Was | Wie | Ergebnis |
| --- | --- | --- |
| **Laden** | `play06-tisch`: von der Seite selbst gemessen, im Servermodus — ohne den Start des Go-Prozesses | Oberfläche bereit nach 160 ms, Tisch mit seinen Knöpfen 207 ms nach „Weiterspielen" |
| **Größe** | `play06-tisch`: das Fenster von 1.440 auf 1.024 Punkte | die Leinwand folgt nach 1 ms, kein Knopf liegt außerhalb |
| **Kontextverlust** | `play06-kontextverlust`: `WEBGL_lose_context` am laufenden Tisch | die Meldung steht da, die Liste darunter bleibt bedienbar, **null** Bildanforderungen in zwei Sekunden |
| **Wiederherstellung** | dieselbe Probe: die Seite neu laden, wie die Meldung sagt | der Tisch zeichnet wieder, mit frischem Kontext |

**Dabei gefunden und behoben** — PLAY-06.AK-04 verlangt, dass ein verlorener
Kontext keine dauerhafte CPU-Schleife erzeugt. **Es liefen zwei, und zwar
schon vorher**: 246 Bildanforderungen in zwei Sekunden am ruhenden Tisch, 256
nach dem Verlust. Die eine war PixiJS' `Ticker.system`, an dem der Scheduler
jedes Renderers und das Ereignissystem hängen; die andere die Messung des
Bildabstands in der Diagnose. **Jetzt: null**, ruhend wie nach dem Verlust —
PixiJS räumt einmal je Sekunde auf, und der Bildabstand wird gemessen, wenn
jemand die Leitung prüft.

**Die Wiederherstellung geht seit dem 2026-10-06 auch über einen
Seitenwechsel** — UI-N01 (#2089).
Bis dahin blieb mit der Seite (UX-03) auch die Leinwand samt verlorenem
Kontext stehen, und **schlimmer**: schon das gewöhnliche Verlassen des Tischs
verlor ihn — PixiJS nahm die Leinwand, die React hält, beim Abräumen aus dem
Dokument und verlor ihren Kontext, und der **Wiedereinstieg** bekam dieselbe
Leinwand zurück. `checkMaxIfStatementsInShader` (8.20.1) halbierte dann bis
null und versuchte es ohne Ende weiter; die Webansicht stand. **Erstbesuch
und Wiederbesuch** sind seitdem derselbe Weg: jeder Anlauf legt eine eigene
Leinwand in den Behälter, den React hält (`tischmalerei`), und ein Kontext,
der den Shader ohne `if` nicht übersetzt, kommt gar nicht erst zu PixiJS.

Gemessen am nativen Fenster (`tischwiedereinstieg`, Beleg in
`docs/belege/huellenprobe/`): je Erscheinungsbild zehn Wiedereinstiege bei
1280 × 800, jeder mit **neuer** Leinwand, denselben 16 Karten und wirkendem
»Näher heran«, eine Leinwand im Dokument vom ersten bis zum letzten,
120–400 ms je Zyklus; ein erzwungener Kontextverlust wird gemeldet, und der
nächste Einstieg zeichnet wieder. **Was das nicht belegt**: hundert Wechsel
derselben Sitzung samt Prozesslast — das ist die Abnahme der Iteration.

---

## Die Probe an den Rändern

#1460 verlangt: „Retina, externes Display, Resize, Vollbild, hell/dunkel,
Kontraste, reduzierte Bewegung, WebGL-Ausfall und Context-Loss geprüft; UI
bleibt bei fehlendem Renderer mit verständlichem Fehler bedienbar."

Neun Punkte, und sie sind **nicht** alle auf demselben Weg zu prüfen
(`docs/HAERTUNG.md` trennt die Ebenen).

| Punkt                                 | Wie geprüft                                                                                                                                                                                              | Stand                                          |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| **hell/dunkel**                       | Die Farben kommen aus `--mse-tisch-*` in `stil.css` und werden bei jedem Anlauf frisch gelesen (`tischfarben`, `ausDemDokument`); Läufe für beide Sätze                                                  | **belegt**                                     |
| **Kontraste**                         | Zehn Paare, von Hand gemessen, `docs/KONTRASTE.md`                                                                                                                                                       | **belegt**                                     |
| **Resize**                            | `useTisch` beobachtet den **Behälter** der Leinwand und lässt den Maler anpassen, **bevor** es misst (`passeAn`); Lauf in `tischbehaelter.test.ts` mit einem Maler, der wie PixiJS erst beim Anpassen die neue Größe kennt, und in WebKit bei 1.440 → 1.024 Punkten (`play06-tisch.probe`) | **belegt**                                     |
| **WebGL-Ausfall**                     | `webglDa` gibt unter jsdom nein; die ganze Suite läuft im Rückfall, und `tischflaeche.test.tsx` prüft ihn ausdrücklich                                                                                   | **belegt**                                     |
| **Context-Loss**                      | `beiVerlust` meldet `verloren`; die Sichtprüfung am 2026-09-16 hat ihn **wirklich ausgelöst** und zwei Fehler freigelegt, die WebKit-Probe am 2026-09-25 (`play06-kontextverlust.probe`) zwei Dauerschleifen — siehe oben                                                                                 | **belegt**                                     |
| **Fehler verständlich, UI bedienbar** | Drei Gründe, drei Sätze; die Kartenknöpfe werden sichtbar statt durchsichtig                                                                                                                             | **belegt**                                     |
| **Reduzierte Bewegung**               | `bewegungErwuenscht` fragt `prefers-reduced-motion` und gibt der Systemeinstellung die **stärkere Stimme**: wer sie setzt, bekommt keinen Flug, auch wenn die Anwendung ihn anböte. Läufe in `bewegung.test.ts`; der Weg bis zur Leinwand in `tischflaeche`                                                     | **belegt**                                     |
| **Retina**                            | `autoDensity` und `resolution` in `pixinaht.ts`; **die Fläche** misst `zeichenflaeche.test.ts` gegen das echte `ViewSystem` von PixiJS bei ein-, zwei- und dreifacher Dichte, und `play06-tisch.probe` prüft in WebKit bei doppelter, dass der Tisch sie füllt                                                                  | **belegt**                                     |
| **Externes Display, Vollbild**        | —                                                                                                                                                                                                        | **offen**                                      |

> **Hier stand bis zum 2026-09-20 das Gegenteil**: *„Es gibt **keine**
> Animation … wer eine Animation nachrüstet, muss `prefers-reduced-motion`
> mitbringen."* Es hat jemand nachgerüstet — `bewegung.ts` führt seit
> Werkstück G4 eine `Bewegungsdauer` von 220 ms —, **und er hat mitgebracht,
> was der Satz verlangt**. Der Code war also richtig; falsch war diese Zeile.
>
> **Und das ist die gefährlichere Hälfte**: wer sie las, hielt den Punkt für
> „trivial erfüllt" und prüfte ihn nicht. Ein Dokument, das sich selbst
> widerspricht, wird an der falschen Stelle geglaubt (PLAY-06.AK-05,
> #1714).
> Seitdem hält
> [`internal/werkzeuge/tischbewegung_test.go`](../internal/werkzeuge/tischbewegung_test.go)
> die beiden auseinander.

> **Resize stand hier bis zum 2026-09-24 als belegt** — mit „`useTisch` misst
> die Fläche am `resize` des Fensters nach". Die Zusage hielt nur die
> Vorrichtung: ihr Maler kannte die neue Größe sofort. PixiJS kennt sie erst
> im nächsten Bild, der Tisch wurde in die **alte** Fläche eingepasst und
> stand abgeschnitten da — und wenn die Seitenleiste auf- oder zuging,
> meldete das Fenster gar nichts. Gesehen hat es die WebKit-Probe.
>
> **Retina stand hier bis zum 2026-09-24 als „gesetzt, nicht gemessen"** — und
> genau dort lag der Fehler, von dem der Abschnitt über WebKit erzählt. Gesetzt
> war die Dichte richtig; falsch war, was danach mit ihr gerechnet wurde. Eine
> Zeile, die eine Einstellung nennt, sagt nichts über ihre Wirkung (§10).

**Die letzte Zeile braucht ein echtes Fenster an einem zweiten Bildschirm.**
Das ist dieselbe Grenze wie oben bei der Rasterung und dieselbe wie in Vermerk
[0007](entscheidungen/0007-kein-entfernter-lauf.md): geprüft ist eine Maschine,
keine Plattform. Sie stehen hier als offen, nicht als erledigt.
