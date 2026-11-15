# Beleg: die vier Folgen der Ausgangsmessung P0 — #2388 bis #2391

**Gemessen am 2026-10-16**, 07:13 bis 07:23 UTC, auf Mac17,3 (Apple M5) mit
macOS 27.0.1, Zweig `fix/folgen-p0`. Jedes Drehbuch lief am Bündel aus
`scripts/bauen.sh abnahme`, gestartet über LaunchServices, mit eigenem
Bestand, Schlüsselbund und Zuhause (`p.ort()`); die Werkzeuge in einem
Wegwerf-Zuhause mit `MSE_BESTAND`, `MSE_SCHLUESSELBUND` und
`MSE_DIENST_GESPERRT=an`. Keine echten Bestände, keine Konten, kein Netz
außer der Rückschleife. Die Karten sind erfunden
(`testdaten/mtgjson/tisch/TSP.json`), die Drucke `p0-blitz`, `09-blitz` und
`09-schock` gibt es nirgends.

Gefunden hat die vier Fehler die Ausgangsmessung P0
(#2336,
`P0-2336.md` (Entwicklungsarchiv), Befunde 1 bis 4).

## Je Korrektur ein Drehbuch und seine Gegenprobe

**Die Gegenprobe ist dasselbe Drehbuch an einem Bündel, in dem genau die
Korrektur fehlt** — die Datei vom Stand `1fcc423cc` statt der korrigierten,
neu gebaut. Ein Grün allein belegte nicht, dass das Drehbuch den Fehler
sieht (AGENTS.md §10).

| Issue | Drehbuch | Mit der Korrektur | Ohne sie | Was zurückgesetzt war |
| --- | --- | --- | --- | --- |
| #2388 | [`tischfremd09`](huellenprobe/tischfremd09.md) | eine Karte, über `mse` auf die Hand gelegt, steht nach **2,2 s** ohne Klick im Fenster — innerhalb von zwei Abfragezyklen | nach **15 s** noch nicht (Lauf `2026-10-16T07-22-56Z-11844`, Bauabdruck `92be2ba1cd5f`) | `seiten/Spielen.tsx`, das den Bestandsstand an `usePartie` reicht |
| #2389 | [`druckfeld09`](huellenprobe/druckfeld09.md) | `mse handel liste --json` trägt `p0-blitz` und `p0-schock` | **`P0-blitz` und `P0-schock`** (Lauf `2026-10-16T07-13-55Z-6638`, Bauabdruck `aafb8dc0ce38`) | die drei Eigenschaften am Feld »Druck« in `bausteine/Handelspositionsform.tsx` |
| #2390 | [`angreifer09`](huellenprobe/angreifer09.md) | »Angreifer erklären« nennt Moosbär und Glutläufer | **Waldboden**, Moosbär, Glutläufer (Lauf `2026-10-16T07-17-25Z-8592`, Bauabdruck `d13d444d3000`) | `bausteine/Kampffeld.tsx` |
| #2391 | [`tauschneu09`](huellenprobe/tauschneu09.md) | H2 endet in beiden Lagen — **angeboten** und **beendet** — im vorbelegten Formular, der Knopf heißt »Bert aufschlagen« | in beiden Lagen **im laufenden Handel**, der Knopf heißt »Bert aufgeschlagen« (Lauf `2026-10-16T07-20-28Z-10412`, Bauabdruck `6c2750fff64b`) | `seiten/Handeln.tsx` und `seiten/Handelsschublade.tsx` |

**Und die Ausgangsmessung selbst bleibt grün**: `ausgangswerte09spielen`
lief am Commit `b3f82a4ca` (Lauf `2026-10-16T07-24-31Z-12474`, Bauabdruck
`198272f890aa`) **bestanden**, S9 mit vier Handlungen und beobachteter
Wirkung — die Zeilen wählt sie weiter nach dem Namen der Kreatur (Abnahme
von #2390). Ihr Beleg und ihr Datensatz sind **nicht** ersetzt: sie sind
die Ausgangswerte von P0, gegen die Phase 7 und 8 messen.

## Was dabei auffiel

- **macOS schreibt den Anfang erst beim Verlassen des Felds groß.** In der
  Gegenprobe von `druckfeld09` stand nach dem Tippen `p0-blitz` im Feld, und
  im Handel `P0-blitz`. Wer den Wert des Felds liest, misst nichts; das
  Drehbuch liest deshalb die Kommandozeile. Welche der drei Eigenschaften
  WebKit braucht, ist nicht getrennt gemessen — gesetzt sind alle
  (`docs/FELDPRUEFUNG.md`, »Kennungsfelder«).
- **Ein freies Feld zeigte die Großschreibung nicht**, auch nicht am Bündel
  ohne Korrektur: ein erstes Drehbuch, das »Wie du ihn nennst« als
  Gegenprobe nahm, blieb dort klein und belegte darum nichts. Es ist nicht
  eingecheckt.
- **#2391 traf auch die Lage »beendet«**, nicht nur »angeboten« — P0 hatte
  den abgeschlossenen Fall als gutartig gemessen, dort aber nach einem
  Wechsel des aufgeschlagenen Handels. Entschieden ist der Fall in
  `PROTOTYP-0.9.0.md` (Entwicklungsarchiv), Abweichungen: jede Wahl
  beginnt einen neuen Entwurf.

## Was es nicht belegt

- **Eine zweite Maschine und einen Menschen**: getippt hat die Probe über
  `internal/fenster/abnahme_darwin.go`, nicht eine Hand an einer Tastatur.
  Dass ein Mensch dieselbe Großschreibung bekommt, folgt aus P0, wo sie auf
  demselben Weg entstand.
- **Den Fall aus dem Kern** (#2390): die Auswahl der Angreifer bleibt im
  Fenster, bis P3 (#2371)
  die Objektaktionen aus dem Kern liefert; dann ist sie ein Fall von P3.
- **Ein offenes Formular am Tisch** (#2388) bleibt beim Nachlesen stehen —
  dieselbe Partie ist dasselbe Bezugsobjekt (§7, Regel 1). Das ist
  entschieden und in `haken/usePartie.ts` begründet, aber am Fenster nicht
  eigens durchgespielt.
