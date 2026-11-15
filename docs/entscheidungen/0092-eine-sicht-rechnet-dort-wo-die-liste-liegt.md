# 0092 — Eine Sicht rechnet dort, wo die Liste liegt

**Datum:** 2026-09-10
**Status:** angenommen
**Anlass:** Werkstück #1500 (Phase 5.2), Befund aus U8

---

## Worum es geht

Die Sammlungsseite holte **alle** Exemplare in einem Ruf und filterte,
sortierte und kappte sie im Fenster. Bei zehntausend Exemplaren sind das rund
**3 MB JSON** über die Wails-Naht — für zweihundert Zeilen, die jemand sieht.

Der Kommentar in `frontend/src/seiten/sammlungssicht.ts` sagte selbst, wann das
zu ändern wäre: „Bei zehntausend Exemplaren würde die ganze Liste weiterhin
geholt und hier gefiltert … wer die Grenze verschiebt, verschiebt sie in den
Dienst — und dann gehört auch die Sortierung dorthin."

## Die Entscheidung

**Die Sicht wandert nach `internal/domain`** — Suche, Ausführung, Ordnung —,
und der Ausschnitt entsteht in `internal/service`. Über die Naht kommt, was das
Fenster zeichnet, samt zwei Zahlen: `gesamt` (die Sicht) und `insgesamt` (die
Sammlung).

**Der Vertrag wird erweitert und nicht gebrochen** (§16): wer nichts angibt,
bekommt weiter alles — an allen vier Flächen. `GET /sammlung` bleibt derselbe
Weg, `lesen.sammlung` dasselbe Werkzeug, `mse sammlung` derselbe Befehl.

## Drei Dinge, die dabei entschieden wurden

### Die Ordnung bleibt in Go und nicht in SQL

**Weil SQLite keine Sprache kennt.** Sein `COLLATE NOCASE` ist ASCII;
„Ätherfluss" stünde damit hinter „Zauber". Das Fenster ordnete mit
`localeCompare`, und wer die Ordnung in den Bestand verschiebt, ohne das zu
prüfen, ändert sie für jeden Namen mit Umlaut.

Geordnet wird deshalb mit `golang.org/x/text/collate` — **ein Ordner je Lauf**:
`collate.Collator` trägt einen Zwischenspeicher und ist nicht nebenläufig
benutzbar. Ein Paketwert davon gab in zwei gleichzeitig laufenden Tests die
Eingabereihenfolge zurück; im Betrieb wären es zwei Fenster, die zugleich
lesen.

**Und ein Schlüssel je Zeile statt einer Zerlegung je Vergleich**:
`CompareString` zerlegt bei jedem Aufruf beide Zeichenketten neu — bei
zehntausend Zeilen rund 120.000 mal. `KeyFromString` zerlegt einmal je Zeile.
Gemessen am 2026-09-10 (Apple M5, `-benchtime=10x`): **62,4 ms → 10,2 ms**.

### Ordnen und Zeigen sind zwei Abfragen

Eine Sicht filtert und ordnet über **alle** Zeilen und zeigt zweihundert. Sie
alle vollständig zu lesen hieße: einundzwanzig Spalten und einen Verbund auf
die Herkunft, zehntausendmal — für Felder, die die meisten Zeilen nie zeigen.

`Ordnungsliste` liest **neun** Spalten ohne Verbund; `NachKennungen` holt die
vollen Zeilen des Ausschnitts, in der Reihenfolge, die die Ordnung ergeben hat.
Gemessen an derselben Datei mit zehntausend Exemplaren: die volle Abfrage samt
Verbund und `ORDER BY` **12,0 ms**, die schmale **5,9 ms**; der ganze Ruf, so
übersetzt wie `wails dev` es tut, **100–110 ms → 80 ms**.

**Die Zeilen der Ordnungsliste sind unvollständig, und das ist der Sinn.** Sie
taugen zum Ordnen und nicht zum Zeigen — wer sie zeigte, zeigte eine Karte ohne
Zustand, ohne Preis und ohne Herkunft.

### Das Fenster behält seine volle Länge

Was noch nicht geholt ist, steht als **Platzhalter** in der Liste. Ohne sie
bräche der Ausschnitt die drei Zusagen aus U8: `aria-rowcount` nennt weiter
alle Zeilen, die Tastatur erreicht jede, und ein Sprung ans Ende holt nach.

**Die gewählte Zeile wird gemerkt**, nicht nur nachgeschlagen: sie kann außer-
halb des geholten Fensters liegen, und ein leerer Inspector wäre die falsche
Antwort auf „ich hatte doch diese Karte offen".

## Was das gekostet hat

**Die Suche kostet jetzt einen Weg über die Naht.** Vorher lagen alle Zeilen
im Fenster, und ein Tastendruck filterte sie in **1 ms**; jetzt sind es rund
**110 ms** je Suchlauf (gemessen mit `arbeitswegmass.mjs sammlung` gegen
`wails dev`). Das ist der Preis dafür, dass die Seite nicht mehr 3 MB holt, um
sie danach wegzuwerfen — und er fällt bei jeder Sammlungsgröße gleich an,
während der alte Weg mit ihr wuchs.

## Was gemessen wurde und was die Zahl nicht sagt

`arbeitswegmass.mjs sammlung` gegen `wails dev`, 10.001 Exemplare, 1280 × 800,
je Lauf ein frisch gestarteter Server (ein Messlauf hinterlässt eine tote
Websocket-Verbindung, und der nächste misst sie mit):

| | Läufe |
| --- | --- |
| vorher (2026-09-10, derselbe Rechner) | 265 · 270 · 332 ms |
| nachher | 175 · 175 · 266 · 267 · 318 ms |

**Die Zahlen zerfallen in zwei Häufungen**, und der Abstand ist rund ein
Ruf über die Naht. Eine Spur mit einem Beobachter im Fenster (statt einer
Abfrage im Takt, die selbst kostet) zeigt warum: unter `StrictMode` läuft jeder
Ladeeffekt **zweimal**, und `useLaden` verwirft nach §7 Regel 2 die erste
Antwort. Bei ⌘6 gehen sechs Rufe hinaus, wo die Anwendung drei macht — und die
Zählzeile wartet mal auf die erste Antwort (175 ms), mal auf die zweite
(266 ms). **`StrictMode` wirkt nur im Entwicklungsmodus**; die ausgelieferte
Anwendung ruft einmal.

**Und `wails dev` übersetzt mit `all=-N -l`** — ohne Optimierung und ohne
Inlining (Wails v2.13.0, `pkg/commands/build/base.go:199–202`). Derselbe Ruf
gegen dieselbe Datei kostet als Prozess **80 ms** so übersetzt und **30–50 ms**
optimiert, jeweils samt Prozessstart, Öffnen des Bestands und JSON.

**Die Reichweite dieser Zahlen ist ein Rechner** (0007) und ein
Entwicklungsmodus. Was sie tragen, ist der Vergleich zwischen vorher und
nachher — nicht die Zusage einer Dauer an Nutzende.

## Verworfen

**Ordnen in SQL über eine eigene Kollation.** `modernc.org/sqlite` ließe eine
Vergleichsfunktion registrieren, und dann könnte `ORDER BY … LIMIT 200` die
ganze Arbeit tun. Dagegen steht dreierlei: der Ruf ginge je Vergleich von C
nach Go, ein Index über eine eigene Kollation müsste mit ihr angelegt sein, und
die Ordnung einer Sammlung stünde danach an zwei Stellen — in der Abfrage und
im Fachkern. Der Gewinn wäre ein Bruchteil dessen, was die zwei Abfragen schon
bringen.

**Eine Spalte mit einem gefalteten Sortierschlüssel.** Sie ließe SQLite ordnen
und indizieren, kostet aber eine Wanderung, einen Schreibweg und die Zusage,
dass die gefaltete Ordnung dieselbe ist wie die von `x/text`. Das ist ein
eigenes Werkstück, kein Nebenzug.
