# 0065 — Wie diese Anwendung aussieht

**Stand:** angenommen
**Datum:** 2026-09-03
**Anlass:** Werkstück E von
#1325 — der
Auftraggeber hat am 2026-09-01 ausdrücklich einen Neuanstrich verlangt: das
Erscheinungsbild soll sich nach einem Fantasy-Sammelkartenspiel anfühlen.

## Lage

**Was heute ist, ist gemessen** (2026-09-01, `frontend/src/`):

| | |
| --- | ---: |
| eigene Schriftarten geladen | **0** — `stil.css` hat 62 Zeilen und importiert nur Tailwind |
| `@font-face`, `--font-*`, Google Fonts | **0** |
| Seiten | **22** |
| Zeilen in `docs/KONTRASTE.md` | **85** |

Das Fenster trägt die Vorgabeschrift von Tailwind und ein Schiefergrau. **Das
ist kein Fehler — es war nie eine Entscheidung.** Genau deshalb gibt es diesen
Vermerk: die nächste Änderung soll sich an etwas messen können.

Das Issue sagt dazu den Satz, um den es geht:

> Der Umbau geht über die **Gestaltungssprache**, nicht über Seiten. Geschmack
> ist weiterhin keine Abnahme.

## Entscheidung

**Sechs Festlegungen, und jede nennt, wogegen sie fiel.**

### 1 Schrift: zwei Stimmen, beide vom System

| | Gewählt | Wogegen |
| --- | --- | --- |
| **Anzeige** — Überschriften, Setnamen, Kartenflächen | `'Hoefler Text', 'Iowan Old Style', Charter, Georgia, serif` | gegen eine mitgelieferte Schriftdatei |
| **Fließtext** — alles andere | die Systemschrift (`-apple-system`, `system-ui`) | gegen eine zweite Anzeigeschrift |

**Warum keine mitgelieferte Datei.** Das Issue verlangt *„mitgeliefert, nicht
bezogen — die Anwendung arbeitet ohne Netz (§16, Regel 5), und ein
Schriftverweis nach draußen bräche das."* Eine Schrift, die **auf jedem
Zielrechner ohnehin liegt**, erfüllt dieselbe Zusage und kostet weder ein
Binärartefakt im Repositorium noch einen Eintrag in `DRITTANBIETER.md`. Das
Ziel ist **darwin/arm64 ab macOS 13** (§12); `Hoefler Text` liegt dort seit
Jahrzehnten, `Iowan Old Style` und `Charter` sind die Rückfälle, `Georgia` und
`serif` der Rest der Kette.

**Was das kostet, gehört ausgesprochen**: auf Windows und Linux — beschrieben,
nicht gebaut (§12) — greift die Kette auf `Georgia` und `serif`. Die
Anwendung sieht dort anders aus. Da sie dort nicht ausgeliefert wird, ist das
kein Mangel, sondern die Folge einer benannten Grenze.

**Warum eine Serifenschrift für die Anzeige.** Jede gedruckte Magic-Karte
trägt ihren Namen in einer Serifenschrift; das Auge kennt das. Eine
Kalligraphie-Schrift (`Trattatello`, `Herculanum`) läge näher am Klischee und
weiter von der Lesbarkeit — sie fiel gegen Grenze 1 unten.

**Warum der Fließtext die Systemschrift bleibt.** §6: *„macOS-Eigenheiten
werden bedient, nicht nachgeahmt."* Eine eigene Textschrift in einem
Mac-Fenster liest sich wie eine Webseite, die sich als Programm ausgibt.

### 2 Farbe: die Sprache steht in Merkmalen, nicht in Klassen

`stil.css` führt **benannte Merkmale** (`--color-grund`, `--color-flaeche`,
`--color-text`, `--color-leise`, `--color-kante`, `--color-marke`). Ein
Bauteil, das eine Farbe unmittelbar schreibt, hat die Sprache umgangen.

**Und seit dem 2026-09-03 sind es Pergament und Tinte.**

Hier stand: *„Die Werte sind zunächst die heutigen … Wogegen: gegen einen
Pergament-Ton als Grund. Er stand zur Wahl und ist verworfen worden, weil er
jede der 85 Zeilen neu erhoben hätte."* Das war eine Abwägung von Aufwand
gegen Wirkung — **und der Auftraggeber hat sie am selben Tag anders
entschieden**: Weiß-Blau ist Tailwinds Vorgabe und war nie eine Wahl; eine
Anwendung für Fantasy-Sammelkarten soll nicht wie ein Formular aussehen.

Die Zeilen sind also neu erhoben worden. **Das ist der Preis, und er war
richtig veranschlagt** — er ist nur bezahlt worden statt gespart.

| Merkmal | hell | dunkel | wofür |
| --- | --- | --- | --- |
| `grund` | `#f2e8d5` | `#15120e` | Pergament / Nacht |
| `flaeche` | `#fdf9f0` | `#231e17` | die Kästen darauf |
| `text` | `#241c14` | `#f0e6d2` | Tinte |
| `leise` | `#5c4a34` | `#b9a688` | verblasste Tinte |
| `kante` | `#786248` | `#7d6c52` | Ränder und Trennlinien |
| `marke` | `#6b4423` | `#d9a441` | Umbra / Gold — der Akzent |

**Gerechnet, nicht gewählt** (WCAG 2.1, relative Leuchtdichte):

| | hell | dunkel |
| --- | ---: | ---: |
| `text` auf `grund` | 13,81 | 15,07 |
| `leise` auf `grund` | 6,96 | 7,89 |
| `marke` auf `grund` | 6,98 | 8,30 |
| `kante` auf `grund` | **4,75** | **3,68** |

Die Ränder liegen damit über 3 : 1 in **beiden** Erscheinungsbildern — die
Zahl, an der `slate-200/700` bei jedem bisherigen Anlauf gescheitert ist (§3).

**Warum Umbra und Gold für den Akzent, und nicht Bordeaux.** Bordeaux war die
naheliegende Wahl zu Pergament — und sie fiel an **Grenze 3**: zwei nahe
Rottöne, einer „Akzent", einer „Fehler", sind genau die Verwechslung, die
diese Grenze verhindern soll. Umbra und Gold kollidieren weder mit dem Rot
der Fehler noch mit dem Bernstein der Befunde.

**Warum kein Grün.** Es liest sich als „in Ordnung"; ein Knopf, der so
aussieht, sagt etwas über sein Ergebnis, bevor jemand ihn gedrückt hat.

### 3 Kante: eine Rundung, ein Rand

Kanten bleiben, wie sie sind: `rounded`/`rounded-md` und ein Rand in
`--color-kante`. **Seit dem 2026-10-16 drei Stufen** — Nachtrag unten. **Wogegen**: gegen Zierrahmen und Schlagschatten. Ein Rahmen
mit Ornament um jedes Feld erschlüge die Karte, die daneben steht — und die
ist das Hauptobjekt (§6).

### 4 Textur: keine

Kein Pergamentmuster, kein Rauschen, kein Verlauf als Grund. **Wogegen**:
gegen genau das, was ein Fantasy-Thema als Erstes nahelegt. Eine Textur hinter
Text senkt den Kontrast an jeder Stelle verschieden, und der Wächter über
`KONTRASTE.md` misst eine **Fläche**, keine Musterung — sie wäre eine Zusage,
die niemand halten kann.

**Die Ausnahme ist die Karte selbst**: was ein Vorlagenpaket an Rahmen,
Verlauf und Symbolen mitbringt, ist ihr Inhalt und nicht unsere Gestaltung.

### 5 Bewegung: keine, außer sie sagt etwas

Kein Ein- und Ausblenden beim Seitenwechsel. **Wogegen**: gegen Übergänge als
Schmuck. Eine Bewegung darf einen Zustandswechsel zeigen (ein Aufklapp, ein
Ladebalken) — sie darf nicht die Zeit kosten, in der jemand die nächste Karte
bauen wollte, und sie muss `prefers-reduced-motion` folgen.

### 6 Die Karte ist das Hauptobjekt — und die Anzeigeschrift zeigt es

Wo eine Seite eine Hauptsache hat, ist es die Karte. Die Anzeigeschrift steht
deshalb an den Überschriften und am Setkopf — dort, wo die Seite sagt, worum
es geht — und **nicht** im Fließtext, der erklärt.

## Vier Grenzen, die der Umbau nicht überschreitet

Sie stehen im Issue und gelten unverändert:

1. **Jede Vordergrundfarbe wird gemessen und steht in
   [`KONTRASTE.md`](../KONTRASTE.md)** — beide Erscheinungsbilder, auch Ränder
   ohne Fläche (§3). Atmosphäre kostet keine Lesbarkeit.
2. **axe-core bleibt sauber** auf allen Seiten, gegen WCAG 2.2 AA.
3. **Kein Zustand hängt allein an einer Farbe, einer Textur oder einem Bild.**
4. **Die Zahlen aus Werkstück A werden nicht schlechter** — Bedienelemente,
   Höhe in Bildschirmen, was über der Falz steht.

## Was damit noch nicht getan ist

**Die Bauteile schreiben ihre Farben weiter unmittelbar.** Rund zweihundert
Dateien tragen `text-slate-600`, `border-slate-500` und ihresgleichen; die
Merkmale stehen bereit, benutzt werden sie erst von den Überschriften.

Das ist kein vergessener Rest, sondern eine Reihenfolge: **erst die Sprache,
dann die Übersetzung.** Ein Umbau beider zugleich hätte 85 Kontrastzeilen und
zweihundert Dateien in einem Zug bewegt, und ein Fehler darin wäre nicht mehr
zuzuordnen gewesen.

Die Übersetzung bekommt ein eigenes Werkstück; sie ist mechanisch und
prüfbar — ein Wächter kann zählen, wie viele Bauteile noch an Tailwinds
Palette hängen.

## Nachtrag vom 2026-09-03: eine Faser und eine wählbare Schrift

Zwei Dinge sind an diesem Tag dazugekommen, beide auf Wunsch des
Auftraggebers, und beide ändern etwas an den Festlegungen oben.

### Festlegung 7 — Grund und Fläche tragen eine Textur

Eine Farbe allein macht kein Pergament. Über `bg-grund` und `bg-flaeche` liegt
seit heute eine Kachel aus zwei Rauschfeldern — feine Faser und grobe
Alterung —, als SVG in `frontend/src/pergament-hell.svg` und
`pergament-dunkel.svg`.

**Sie ist einseitig.** Im hellen Bild dunkelt sie nur ab, im dunklen hellt sie
nur auf. Damit ist der Extremfall bekannt und messbar, und `KONTRASTE.md` kann
eine Zahl nennen, die wirklich gilt — nämlich die des ungünstigsten von 810.000
Bildpunkten.

**Und ihre Kurve ist oben abgeschnitten.** Das ist der Kern und der Ertrag von
drei Fehlversuchen: eine Textur, die kräftig genug für den richtigen Eindruck
ist, riss acht von sechzehn Kontrastpaaren; eine, die alle hielt, war fast
unsichtbar. Erst die Klemmung im `feComponentTransfer` trennt die zwei Fragen —
wie stark die Struktur wirkt, und wie weit sie im Extremfall geht.

**Der Preis steht in den Zahlen**: zwischen 0,45 und 1,58 Kontrastpunkten. Der
knappste Wert ist `border-kante` mit 3,99 gegen eine Schranke von 3. Wer die
Textur verstärken will, hat 0,99 Punkte, und danach ist Schluss.

**Keine Vignette.** Sie stand im Entwurf und ist an dieser Messung gescheitert:
sie hätte genau den Rest des Budgets genommen. Ein Fensterrand ist ohnehin kein
Blattrand — links liegt die Seitenleiste.

**`prefers-contrast: more` schaltet sie ab.** Wer im System mehr Kontrast
verlangt, hat einen Grund; eine Textur, die abdunkelt, arbeitet dagegen.

### Festlegung 8 — die Anzeigeschrift ist wählbar

Festlegung 6 sagte: *eine* Anzeigeschrift, `Hoefler Text`, und keine
Kalligraphie. Der zweite Teil steht weiter — für die **Vorgabe**. Wählbar sind
seit heute vier, unter *Einstellungen → Anzeigeschrift*:

| Wahl | Schrift | Anmutung |
| --- | --- | --- |
| `pergament` | `Hoefler Text` | ruhig und lesbar — die Vorgabe |
| `fantasy` | `Luminari` | verspielt, unziale Formen |
| `antike` | `Herculanum` | römische Kapitalis |
| `kalligrafie` | `Trattatello` | von Hand geschrieben |

**Warum das kein Widerspruch zu Festlegung 6 ist.** Der Einwand dort lautete,
eine Kalligraphie sei näher am Klischee und weiter von Grenze 1 — *keine
Karnevalsversion einer Fantasy-Oberfläche*. Das gilt für eine Vorgabe, die
jeder bekommt. Es gilt nicht für eine Wahl, die jemand ausdrücklich trifft: wer
`Trattatello` einstellt, will genau das, und die Grenze schützt ihn nicht,
sondern bevormundet ihn.

**Was nicht wählbar ist: der Fließtext.** Er bleibt die Systemschrift, in jeder
Wahl. Eine Zierschrift über einem Absatz ist keine Gestaltung, sondern eine
Zumutung — und §6 verlangt Lesbarkeit, nicht Stimmung.

**Alle vier liegen auf jedem Mac.** Nachgesehen am 2026-09-03 auf macOS 26. Das
ist dieselbe Bedingung wie in Festlegung 6 und im Verworfen darunter: die
Anwendung arbeitet ohne Netz (§16, Regel 5), und eine geladene Schrift wäre
entweder ein Binärartefakt im Baum oder eine Abhängigkeit vom Netz.

**Und jede Kette endet bei `serif`.** macOS ist das Ziel, aber die Pipeline
übersetzt auch für Windows und Linux (§12); dort entscheidet der Rückfall, ob
eine Überschrift überhaupt eine Serifenschrift ist.

## Nachtrag vom 2026-09-20: Tiefe, und wo sie aufhört

Auf Wunsch des Auftraggebers soll der Spieltisch plastisch wirken. Das
berührt Festlegung 3, und deshalb steht es hier und nicht im Quelltext.

### Festlegung 9 — der Tisch ist ein Gegenstand, keine Bedienfläche

Festlegung 3 verwirft Zierrahmen und Schlagschatten, und ihr Grund steht
daneben: *„Ein Rahmen mit Ornament um jedes Feld erschlüge die Karte, die
daneben steht."* Das ist ein Satz über **Bedienflächen**. Der Spieltisch
ist keine — er stellt einen Tisch dar, auf dem Karten liegen.

**Tiefe gibt es an genau zwei Stellen:**

1. **Tischfläche und Sitze** tragen einen Lichtverlauf.
2. **Eine Karte** wirft einen Schatten, wo sie als körperlicher Gegenstand
   liegt. Sie ist das Hauptobjekt (§6), und ein Blatt auf einem Tisch
   wirft einen.

**Überall sonst gilt Festlegung 3 unverändert**: Kästen, Felder, Drawer,
Befehlspalette und Popover tragen ihre Kante und keinen Schatten.

**Der Verlauf ist einseitig, und das ist kein Vorbehalt, sondern das
Messergebnis.** Gemessen am 2026-09-20 nach WCAG 2.1 über sRGB: die
Tischkante steht im hellen Bild gegen den Tischgrund bei **3,11** — elf
Hundertstel über der Schranke von 3. Ein symmetrischer Verlauf von nur
±4 % fällt auf **2,88** und ist damit erledigt, bevor er sichtbar wird.

Deshalb läuft der Verlauf von der Tokenfarbe aus **nur in die sichere
Richtung**: im hellen Bild hellt er auf, im dunklen dunkelt er ab, je um
12 %. Damit bleibt die Tokenfarbe der **ungünstigste** Punkt der Fläche —
die Zahlen in [`KONTRASTE.md`](../KONTRASTE.md) bleiben die Untergrenze,
und der Verlauf verbessert jede Stelle:

| Fläche | Ton | Ende des Verlaufs | Kante vorher | Kante am Ende |
| --- | --- | --- | ---: | ---: |
| hell, Tischgrund | `#d4bb94` | `#d9c3a1` | 3,11 | **3,37** |
| hell, fremder Sitz | `#dcc4a0` | `#e0cbab` | 3,42 | **3,65** |
| hell, eigener Sitz | `#e4cfae` | `#e7d5b8` | 3,80 | **4,01** |
| dunkel, Tischgrund | `#221b13` | `#1e1811` | 4,91 | **5,07** |
| dunkel, fremder Sitz | `#2a2218` | `#251e15` | 4,52 | **4,75** |
| dunkel, eigener Sitz | `#332a1e` | `#2d251a` | 4,06 | **4,36** |

**Dieselbe Einseitigkeit wie bei der Faser**, und aus demselben Grund: ein
Extremfall, den man kennt, lässt sich messen; einer, der in beide
Richtungen läuft, nicht.

**Die vier Grenzen oben gelten weiter**, und die dritte ist hier die
scharfe: **kein Zustand hängt an der Schattierung.** Getappt ist eine
Drehung, angreifend eine Lage, „wartet" ein Wort — nicht eine hellere
oder dunklere Fläche.

**Was das nicht erlaubt.** Keine Vignette — Festlegung 7 hat sie an genau
dieser Messung schon einmal verworfen. Keine Textur auf dem Tisch: die
sechs Tischbilder bleiben Farbsätze. Und keinen Schatten unter etwas, das
man bedient.

**Und es nimmt etwas zurück.** Der Prototyp der Zieloberfläche trug am
Tag seiner Veröffentlichung `box-shadow` an Drawer, Befehlspalette und
Popover. Das war ein Verstoß gegen Festlegung 3, keine Auslegung davon,
und es ist behoben.

**Die Anwendung trug am selben Tag sechs, und niemand hatte es gesehen.**
Seitenleiste als Schublade (`shadow-xl`), Befehlspalette (`shadow-2xl`),
Menüliste als Popover (`shadow-lg`), dazu Tooltip, Einführung und
Kommandozentrale — also genau die drei Flächen, die der Satz oben
benennt, und drei weitere.

**Warum es niemand gesehen hat, ist der eigentliche Fund.** Der Entwurf
belegte die Einhaltung mit dem Satz *„`frontend/src/stil.css` enthält
null `box-shadow`"*, und der stimmt. Die Schatten standen nicht dort: sie
standen als **Tailwind-Klassen** in den Bauteilen. Eine Messung über die
Merkmalsdatei reicht so weit wie die Merkmalsdatei (§10).

Sie tragen seit dem 2026-09-20 ihre Kante — `border-2 border-kante`, wie
der Entwurf sie zeichnet —, und
[`internal/werkzeuge/schlagschatten_test.go`](../../internal/werkzeuge/schlagschatten_test.go)
hält es. **Was er nicht hält**, steht in seinem Kopf; vor allem gilt: ein
Ring ist eine Kante und kein Schatten, und `Zuschnittrahmen.tsx` braucht
ihn für die zweite Linie aus `image_slice_window.cpp`.

## Nachtrag vom 2026-10-13: Bewegung mit Grenze und eine Ebene für Überlagerungen

**Anlass:** Befund B-30 aus dem Audit zur Produktionsreife 0.9.0 und AP-1
(#2337).
Der Auftrag für 0.9.0 verlangt eine lebendigere Oberfläche — Rückmeldung bei
Hover, Druck und Fokus, sichtbares Auf- und Zuklappen, Überlagerungen, die
sich vom Grund abheben. Festlegung 3 (keine Schlagschatten) und Festlegung 5
(Bewegung nur, wenn sie etwas sagt) stehen dem in ihrer heutigen Fassung im
Weg. **Still abzuweichen wäre ein Befund** (§6); deshalb steht die Grenze
hier, bevor ein Effekt gebaut wird.

### Festlegung 10 — Zustandsbewegung, höchstens 200 ms

**Erlaubt ist Bewegung, die einen Zustandswechsel zeigt** — Hover, Druck,
Fokus, Auswahl, Auf- und Zuklappen, das Erscheinen einer Überlagerung, ein
Ablageziel beim Ziehen. Festlegung 5 bleibt: **kein Übergang beim
Seitenwechsel**, kein Ein- und Ausblenden ganzer Seiten.

| Grenze | Festlegung |
| --- | --- |
| **Dauer** | höchstens **200 ms**; die zwei Merkmale `--dauer-kurz` (**120 ms**: Farbe und Deckkraft bei Hover und Druck) und `--dauer-mittel` (**160 ms**: Erscheinen und Lage einer Überlagerung) |
| **Eigenschaften** | nur **Farbe** (Text, Grund, Kante, Umriss), **Deckkraft** und **Lage** (`transform`) — nie Breite, Höhe, Abstand oder Position im Fluss: eine Bewegung, die das Layout verschiebt, ist ein Layoutsprung mit Ansage |
| **Wenig Bewegung** | unter `prefers-reduced-motion: reduce` sind beide Merkmale **0 ms**; die Regel in `barrierefreiheit.css`, die jede Dauer auf 0,01 ms setzt, bleibt darüber stehen |
| **Kein Zustand hängt an der Bewegung** | Grenze 3 oben: was sich bewegt, zeigt einen Zustand, der auch ohne Bewegung dasteht |

**Die Kartenbewegung am Tisch ist nicht gemeint.** Sie zeichnet PixiJS
(`frontend/src/tisch/bewegung.ts`, 220 ms), sie zeigt einen Gegenstand, der
die Zone wechselt, und sie folgt Festlegung 9 — samt
`prefers-reduced-motion` (`bewegungErwuenscht`). Ob sie kürzer wird,
entscheidet P5 (#2373)
an seiner Messung, nicht dieser Nachtrag.

### Festlegung 11 — eine Ebene für Überlagerungen

**Eine einzige Schattenstufe**, das Merkmal `--shadow-ueberlage`, und nur an
dem, was **über** der Seite liegt und sie nicht ersetzt:

- die **Befehlspalette**,
- ein **Popover** samt Menüliste,
- ein **Kontextmenü**.

**Die Kante bleibt** (`border-2 border-kante`): der Schatten kommt hinzu und
ersetzt nichts — eine Überlagerung ist auch ohne ihn abgegrenzt, und
`prefers-contrast: more` nimmt ihn weg (`contrast-more:shadow-none` an
derselben Stelle, wie die Faser in Festlegung 7). **Nie unter etwas, das man bedient**:
kein Knopf, kein Feld, kein Kasten der Seite trägt ihn. Schublade, Dialog und
Einführung bleiben ohne — sie verdecken die Seite mit einem Schleier, und ein
Schatten sagte dort nichts, was der Schleier nicht schon sagt.

**Nimmt das Festlegung 9 zurück?** Teilweise, und ausdrücklich: dort fielen
am 2026-09-20 die Schatten an Palette und Popover, weil sie **ohne
Entscheidung** dastanden. Jetzt steht eine da, mit **einer** Stufe statt
sechs, und mit Kante darunter.

### Der Vertrag mit den Kontrasten

- **Der Schatten liegt außerhalb der Fläche** und ändert keinen gemessenen
  Kontrast in ihr; die Kante der Überlagerung gegen den Grund bleibt die Zahl
  aus [`KONTRASTE.md`](../KONTRASTE.md).
- **Jede neue Zustandsfarbe** — ein Hover-Grund, ein gedrückter Knopf, ein
  Ablageziel — bekommt ihre Zeile in `KONTRASTE.md`, gemessen in **beiden**
  Erscheinungsbildern, bevor sie gebaut ist: Text mindestens 4,5 : 1,
  Bedienelemente und Trennlinien mindestens 3 : 1. Der Wächter über die Datei
  hält es wie bisher.

### Was es nicht erlaubt

Keine Seitenübergänge, keine Bewegung über 200 ms an der Oberfläche, keine
zweite Schattenstufe, keinen Schatten an Bedienelementen, keine Textur und
keine Vignette (Festlegungen 4 und 7 bleiben).

**Zwei Wächter halten es**:
[`internal/werkzeuge/bewegung_test.go`](../../internal/werkzeuge/bewegung_test.go)
für Dauer und Eigenschaften,
[`internal/werkzeuge/schlagschatten_test.go`](../../internal/werkzeuge/schlagschatten_test.go)
für die eine Stufe und ihre drei Dateien.

## Nachtrag vom 2026-10-16: Festlegungen 2 und 3 werden eingelöst

**AP-2** (#2359),
Befunde B-07 und B-09 aus dem Audit zu 0.9.0. Festlegung 2 sagt seit Phase
4.10, dass ein Bauteil, das eine Farbe unmittelbar schreibt, die Sprache
umgangen hat. **Gezählt am 2026-10-16 standen trotzdem 568
Palettenklassen** in 196 Dateien, und Fehlermeldungen hatten **zwei**
Farben: 77-mal `text-warnung` und über 200-mal `text-red-800
dark:text-red-300`. Festlegung 3 nannte eine Rundung und ließ sechs zu.

### Was jetzt gilt

- **Ein Fehlersystem.** `text-fehler`, `border-fehler` und `bg-fehler/…`
  kommen aus `--mse-warnung` — dem Merkmal, das die Farbwelten setzen und die
  Sperre misst. Der Name der Variablen bleibt, weil er ein Schlüssel in
  `domain.Gestaltungsmerkmale` ist und in gespeicherten eigenen Farben steht
  ([0134](0134-eigene-gestaltung-sind-farben.md)); umbenannt ist nur die
  Farbe, unter der Tailwind sie führt. **Das Rot und das Umbra der Grenze 3
  bleiben auseinander** — dieser Satz oben gilt für `fehler` wie bisher für
  `warnung`.
- **Neue Merkmale für das, was vorher Palette war** — mit den Werten von
  vorher: `hinweis` (Bernstein, Warnungen), `erfolg` (Smaragd),
  `zeichenkette` und `zahl` (Quelltext), `auswahl`/`auswahlkante` und
  `markierung`/`markierungskante` (Zeichenfläche und Linien der Vorschau),
  `bildhell`/`bilddunkel` (was über einem Bild liegt und deshalb keine
  Farbwelt hat). Sie folgen dem Erscheinungsbild, nicht der Farbwelt; die
  Sperre misst `hinweis` und `erfolg` gegen eigene Flächen mit.
- **Drei Rundungen und rund**, in `@theme`: `rounded-sm` (0,25 rem — Feld,
  Knopf, das alte `rounded`), `rounded-md` (0,375 rem — Kasten),
  `rounded-lg` (0,5 rem — Überlagerung), dazu `rounded-full` und
  `rounded-none`. `rounded-xs` und `rounded-xl` gibt es nicht mehr; drei
  Stellen mit `xl` tragen seither `lg`.

**Ein Wächter hält beides**: `frontend/src/farbsprache.test.ts` weist jede
Palettenklasse und jede andere Rundung außerhalb der Tests ab. Kommentare
liest er nicht — sie dürfen erzählen, was früher dastand. **Was er nicht
hält**: eine Farbe, die als Wert geschrieben ist (`style={{ color: … }}`); am
selben Tag gezählt stand keine in einem Bauteil.

**Gemessen** am gerenderten Element in WKWebView (`farbsprache09`) und
gerechnet in `docs/KONTRASTE.md`, Abschnitt „Ein Fehlersystem und die
Merkmale der Meldungen“.

## Verworfen

- **Eine mitgelieferte Schriftdatei.** Grund oben: sie kostet ein
  Binärartefakt, einen Lizenzvermerk und eine Ladezeit, und sie löst kein
  Problem, das die Systemschriften offen ließen.
- ~~Ein Pergament-Grund.~~ **Am 2026-09-03 angenommen statt verworfen** — der
  Auftraggeber hat es verlangt, und der veranschlagte Preis ist bezahlt
  worden: die Kontraste sind neu erhoben. Der Grund für das ursprüngliche
  Nein bleibt richtig; er war ein Grund gegen den **Aufwand**, nicht gegen die
  Sache.
- **Bordeaux als Akzent.** Zu nah am Rot der Fehler (Grenze 3).
- **Grün als Akzent.** Es liest sich als „in Ordnung", und ein Knopf soll
  nichts über sein Ergebnis sagen, bevor jemand ihn drückt.
- ~~Eine Kalligraphie-Schrift für Überschriften.~~ **Am 2026-09-03 als
  *Wahl* angenommen, als Vorgabe weiter verworfen** — Festlegung 8. Der Grund
  gilt für das, was jeder bekommt, nicht für das, was jemand ausdrücklich
  einstellt.
- **Übergänge beim Seitenwechsel.** Sie kosten Zeit und sagen nichts —
  Festlegung 10 erlaubt Bewegung an Zuständen, nicht an Seiten.
- **Mehr als eine Schattenstufe** (Festlegung 11). Zwei Stufen verlangten
  eine Rangfolge der Überlagerungen, und jede zweite Stufe ist ein Schatten
  mehr, der etwas zu sagen scheint.
