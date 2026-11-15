# 0188 — Das Ziel für F02, beschlossen vor dem Messen

**Stand:** angenommen am 2026-10-11, das Ergebnis nachgetragen am selben Tag ·
#2244 (F02).
**Entschieden in diesem Auftrag**, nicht von einem Menschen — F02 verlangt
„Ziel vor Optimierung beschließen", und dieser Vermerk ist der Beschluss.

## Die Frage

F02 soll Mittel, Median, P99 und den Textanteil des Abstands zu MSE2 senken,
und **kein einzelner Fall** darf schlechter werden als eine bestätigte
Toleranz. Was „senken" heißt, muss feststehen, **bevor** eine Zahl vorliegt —
sonst wird das Ziel an das Ergebnis angepasst, und eine Zahl, die man sich
nach dem Lauf aussucht, prüft nichts mehr (dieselbe Regel wie in
[`RETRIEVALMESSUNG.md`](../RETRIEVALMESSUNG.md)).

## Was schon geändert ist, bevor gemessen wird

Zwei Fehler, gefunden beim Lesen einer einzelnen Karte und **nicht** an einer
Kanonzahl ausgesucht — „All in Good Time" aus Archenemy, deren Name bei MSE2
106 und hier 101 Bildpunkte breit war:

1. `weight:` und `style:` aus dem `font:`-Block galten nicht
   (`Font::update`, `src/data/font.cpp`, Fassung f9d9356) — 768 Angaben in 207
   Stilen.
2. Die Schriftsuche fand den fetten Schnitt einer Familie nicht, wenn er in
   einer eigenen Datei liegt (`matrixb.ttf`: Familie „Matrix", Schnitt „Bold").

Beide sind **Portierungsfehler**, keine Stellschrauben am Maß. Sie wären auch
ohne F02 behoben worden.

**Und vier weitere, gefunden an einem eigenen Prüfset** — fünf Karten in
`magic-old`, `magic-nouveau`, `magic-nouveau_duels`, `magic-mtgscommunity-1`
und `magic-planeshifted`, von MSE2 und von hier gezeichnet: der Textschatten
fehlte ganz, `to_int` und `to_real` konnten keine Farbe, eine leere
Mehrfachwahl trug ihre erste Wahl, und ein Bildfeld ohne Bild zeigte nicht
MSE2s Schachbrett. Auch sie sind Portierungsfehler und vor der Messung
behoben; den Kanon berühren sie kaum, denn seine Karten haben Bilder und
gewöhnliche Rahmen.

## Was die Referenz selbst falsch zeichnet

**Jedes Setsymbol der MSE2-Referenz ist einfarbig, in der Randfarbe seiner
Fassung** — `uncommon` und `mythic rare` schwarz, `common` weiß. Die zehn
kanonischen Sets tragen kein Setsymbol, jede Karte zeigt also MSEs
Standardquadrat, und bei MSE2 auf dem Mac hat es keinen Innenraum.

**Das ist ein Fehler von wxWidgets unter macOS, nicht von MSE2 und nicht
dieses Programms.** MSE2 legt den Innenraum eines Symbols über einen Puffer
mit `dc.Blit(…, interior, …, wxAND_INVERT)` auf (`combineBuffers`,
`src/render/symbol/viewer.cpp`, Zeilen 85–89, Fassung f9d9356 vom
2021-07-06); unter macOS ist der Speicher-DC ein `wxGCDC`, und dessen
`Blit()` kann nur `wxCOPY`, `wxOR`, `wxNO_OP`, `wxCLEAR` und `wxXOR` — „any
other function will result in an assertion failure and not drawing
anything" (wxWidgets-3.3-Doku zu `wxGCDC`, nachgeschlagen am 2026-10-11).
`filter_symbol` liest den Innenraum am Grünkanal (`src/render/symbol/
filter.cpp`, Zeilen 20–50); ohne die Verknüpfung bleibt er Rand.

**Dieses Programm baut den Fehler nicht nach.** Es zeichnet das Symbol, wie
MSEs Rechnung es meint — Füllung und Rand —, also so, wie ein DC es zeichnet,
der `wxAND_INVERT` kann. Unter Windows ist der Speicher-DC einer von GDI; dass
MSE2 dort richtig zeichnet, ist daraus **geschlossen, nicht gemessen** — ein
Windows lag nicht vor.
Der Preis steht im Kanon: jede Karte behält an ihrem Seltenheitssymbol einen
Abstand, den keine Korrektur hier senken kann, und der Vergleich Karte für
Karte sieht ihn nicht, weil er in beiden Läufen derselbe ist.

## Der Beschluss

**Die Grundlinie** ist der Kanonlauf am Stand **vor** F02 — Commit
`f388884d0` —, mit derselben MSE2-Ausgabe, denselben Schriften und demselben
Stichprobenplan wie der Lauf danach. **Nicht** die Zahlen von 0.6.0
(Mittel 13,792): zwischen 0.6.0 und heute liegen zwei Fassungen, und ein
Vergleich mit ihnen schriebe F02 zu, was dazwischen geschah.

**Vor der Grundlinie läuft MSE2 selbst noch einmal** (`scripts/mse2-referenz.sh`
in ein frisches Verzeichnis), und jedes Bild wird gegen die liegende Referenz
gehalten. Weicht eines ab, ist die Referenz veraltet und wird ersetzt, bevor
irgendetwas gemessen wird.

**Das Ziel**, alles gegen die Grundlinie:

| Größe | Ziel |
| --- | --- |
| Kanonmittel (geschichtet) | **kleiner** |
| Median | **nicht größer** |
| P99 | **nicht größer** |
| Anteil des Texts am Abstand | **kleiner** |
| jede gezogene Karte | **höchstens 0,01 schlechter** (`MSE_KANON_VORHER`) |

**Die Toleranz von 0,01** ist keine Großzügigkeit: der Lauf ist ohne Zufall
gezogen und wiederholt sich bis auf die vierte Nachkommastelle
([`MESSMETHODIK.md`](../MESSMETHODIK.md) §5a). Was darüber wandert, hat eine
Änderung bewegt.

**Median und P99 nur „nicht größer"**: die beiden Änderungen treffen Felder
mit `weight: bold` — vor allem die kleinen Schichten Archenemy, Planechase und
Vanguard, zusammen 263 von 25.148 Karten. Dass sie den Median verschieben, ist
nicht zu erwarten, und ein Ziel, das man nur durch Glück erreicht, wäre keines.

## Was es kostet

**Ein kleines Ziel.** Der große Hebel — die Glättung der Glyphen, die Quartz
der Plattform gibt (`docs/SCHRIFTUNTERSUCHUNG.md`) — ist damit nicht
angegangen; er gehört nicht MSE, und ihn nachzubauen ist eine eigene Frage.

**Eine Karte, die schlechter wird, macht das Ziel verfehlt** — auch wenn das
Mittel sinkt. Wird eine schlechter, weil MSE2 dort etwas tut, das dieses
Programm jetzt richtig nachbildet, und die Referenz dennoch anders aussieht,
wird das **benannt** und hier nachgetragen, nicht still in die Toleranz
gelegt.

## Das Ergebnis, gegen dieses Ziel

**Gemessen am 2026-10-11**, Grundlinie `f388884d0`, Stand `098bfbc06`, dieselbe
MSE2-Ausgabe (vor der Grundlinie neu gezogen: 25.148 Bilder, keines anders),
dieselben Schriften, derselbe Plan, 4.266 gezogene Karten, je Karte gegen
`vorher.tsv` (sha256 `e2cb07e5…f9a5`):

| Größe | Grundlinie | Stand | Ziel |
| --- | ---: | ---: | --- |
| Kanonmittel (geschichtet) | 13,792 | **11,475** | kleiner — **erreicht** |
| Median | 12,439 | **10,520** | nicht größer — **erreicht** |
| P99 | 39,429 | **35,760** | nicht größer — **erreicht** |
| Anteil des Texts am Abstand | 41,98 % | **33,41 %** | kleiner — **erreicht** |
| jede Karte höchstens 0,01 schlechter | — | 4.195 besser, 52 gleich, **19 schlechter** | **verfehlt, benannt (unten)** |

**Der große Teil kam nicht aus dem, was oben vor dem Messen beschlossen
war**, sondern aus vier Portierungsfehlern, die der Vergleich Karte für Karte
ans Licht brachte — jeder am C++ nachgelesen und, wo er eine Zahl braucht,
über einen Orakelstil gefragt: die **Maske eines Textfelds** bestimmt die
Breite jeder Zeile
(#2299); ein
**Symbol im Text** misst in ganzen Punkten, hat ohne Angabe einen Punkt
Abstand und schrumpft mit dem Text
(#2300) —
die Manakosten fielen damit von 64,53 auf 25,15; ein zu langer Text wird
**gesucht wie bei MSE2** und steht einen Schritt unter dem größten passenden
Maßstab; und die **Messung setzt, was gezeichnet wird** (Symbole, Stilbild,
`scale down to:`). **Das Ziel selbst ist nicht nachträglich geändert** —
die Zeilen stehen, wie sie vor dem Messen standen.

**Die Zwischenstände** gehören dazu, weil sie zeigen, was welche Änderung
tat: mit den Korrekturen von oben und den ersten aus F02 (Zusatzwerte, leere
Felder, Bildreihe) 13,766, mit
der Messung samt Symbolen 13,752 (16 Karten schlechter, die Planeswalker
voran), mit der Maske 13,742, mit Symbolmaßen und Suche 12,832 (519
schlechter — der fehlende senkrechte Abstand), mit dem Abstand 11,475.

**Am Endstand nachgemessen** (2026-10-11, `make fremdpakete` am Stand
`1a45973`, Beleg `docs/belege/fremdpakete.json`, PASS): Kanonmittel
**11,470**, Median **10,514**, P99 **35,760**, Textanteil **33,40 %**; die
Zufallsstichprobe (4.582 Karten) **11,394**, Band [11,279 · 11,509]. Gegen
den Stand oben änderten die Befunde des Nachlaufs — rechter Absatzrand,
Zwangsumbruch, oberer Rand des ersten Absatzes, Maß einer Wahl, freie
`symbol_variation` — **12 Karten, alle besser**, und keine der 19 unten. Der
Deckel des Volllaufs (`deckelMittel` 20,4, für `MSE_STICHPROBE=alle`) ist
damit weit unterschritten; nachgezogen wird er mit dem nächsten Volllauf über
alle 25.148 Karten.

**Die 19 schlechteren Karten**: Clockwork Beast +8,02, Illusionary Mask
+5,33, Rowan Kenrith +4,65, Will Kenrith +2,76, Cruel Revival +2,75, Liliana,
Death's Majesty +2,24, Starke +2,02, Illusionary Presence +1,88, Bleed Dry
+1,62, Command Performance +1,38, Noxious Ghoul +1,34, Trivia Contest +1,25,
Calim, Djinn Emperor +0,84, Minotaur de Force +0,69, Aerialephant +0,64,
Carnival Carnivore +0,51, Juju Bubble +0,40, Elephant Grass +0,06, Colossus
Hammer +0,04. **Bei den beiden untersuchten kippt ein Zeilenumbruch**:
Clockwork Beast bricht in voller Größe in elf statt zehn Zeilen und wird
deshalb verkleinert, bei Rowan Kenrith passt eine Zeile nicht mehr in die
Maskenzeile. **Die Textbreite ist es nicht** — über Orakelstile gefragt
stimmen Zeilenhöhe, die Breite von fünf Zeilen dieser Karten und die Maße
eines Manasymbols auf 0,25 Punkte. Was die verfügbare Breite einer Zeile bei
MSE2 größer macht, ist **nicht gefunden**; es steht als Folgepaket
#2302 da,
mit Orakelpflicht.

**Und die Grundlinie war an einer Stelle zufällig gut**: vor F02 maß die
Messung ohne Symbole, und die Treuekästen der Planeswalker standen dadurch
dort, wo MSE2 sie zeichnet — zwei Fehler, die sich aufhoben. Einige
„Verschlechterungen" der Zwischenstände waren das Ende dieses Zufalls.

## Nachtrag vom 2026-10-14 — die sechs Karten aus #2315

**Sie verkleinern nicht anders.** Am Orakel gefragt — ein instrumentiertes
MSE2, je Feld und Satz Start- und Endmaßstab, Zeilen, Größe und Kasten —
zeichnet MSE2 alle sechs in Maßstab 1, mit denselben Zeilen wie hier. Sie
zeichneten **falsche Symbole**: `in_context` wirkte nicht, und auf der
Rückfalllinie `regexp2` galten keine POSIX-Klassen
(E2315 (Entwicklungsarchiv)). Nach dem Port stehen sie
bei 7,2 bis 9,0 statt 15,5 bis 20,1.

**Karte für Karte gegen die Grundlinie** (am Stand `b1eef91cb`, 4.266 Karten,
Kanonmittel **9,328**): zwei Karten sind um mehr als 0,01 schlechter — Warlock Class (16,48 → 19,41) und Cleric Class (12,08 → 16,17). Beide sind eine **begründete Abweichung**: MSE2 setzt beim Laden die Stufentexte aus ihren gespeicherten Werten zusammen, diese Anwendung aus den gerechneten, wie MSE2 nach einer Bearbeitung ([`PORTIERUNGSKARTE.md`](../PORTIERUNGSKARTE.md), E2315; der Weg zu beidem in #2396). Alle übrigen halten die Grundlinie oder liegen darunter, die 19 von oben eingeschlossen.

**Nachtrag vom 2026-10-20**: seit #2396 zeichnet diese Anwendung wie MSE2 beim Laden ([0203](0203-gespeichert-wird-was-gerechnet-ist.md)). Warlock Class steht bei 10,59, Cleric Class bei 10,99 — beide unter der Grundlinie. Karte für Karte gegen den Stand davor: zwölf besser, keine schlechter, Kanonmittel 9,319.

## Verworfen

- **Gegen die Zahlen von 0.6.0 messen** — oben.
- **Ein festes Ziel wie „Mittel unter 13"** — es hinge an einer Wirkung, die
  vor dem Lauf niemand kennt, und es verführte dazu, so lange zu drehen, bis
  es erreicht ist.
