# Schriftuntersuchung

**Warum Text bei MSE2 anders aussieht als hier — Aspekt für Aspekt getrennt.**

QA-03.AK-02 (#1656)
verlangt, dass Schriftdatei, Fassung und Hash, Shaping, Kerning, Zeilenumbruch,
Grundlinie, Rückfall, Hinting und Kantenglättung **getrennt** untersucht werden —
und dass kein Geometriefehler unter einer Glättung verschwindet. Diese Datei hält
fest, was dabei gemessen wurde, womit, und was offen bleibt.

**Stand: 2026-09-23**, auf diesem Rechner (macOS 27, arm64). Die Referenz ist
MSE2 in Fassung f9d9356 vom 2021-07-06, gebaut gegen wxWidgets 3.3.3 aus
Homebrew, neu gezeichnet am selben Tag mit Manifest (`scripts/mse2-referenz.sh`).

## Die zwei Wege

| | MSE2 unter macOS | dieses Programm |
| --- | --- | --- |
| Satz | `TextViewer`, Vorsilbenbreiten von wx, gerastert auf ¼ Punkt | portiert (#598, #546) |
| Zeichnen | `RotatedDC::DrawText` → `draw_resampled_text`: vierfach, weiß auf schwarz, Kastenfilter (`src/util/rotation.cpp` 189–219, `src/gfx/resample_text.cpp` 166–211) | seit QA-03 derselbe Weg (`internal/render/textraster.go`) |
| Glyphen | wxWidgets → **Core Text**: `CTLineCreateWithAttributedString`, `CTLineDraw` | `golang.org/x/image` — `sfnt`, `opentype`, `vector` |
| Kerning | Core Text aus `GPOS` oder `kern` | eigener Leser (`internal/render/kerning.go`) |

**Beim Bildexport läuft der vierfache Weg immer**: `DataViewer::draw` wählt
`QUALITY_AA`, sobald `card_anti_alias` gesetzt ist
(`src/render/card/viewer.cpp` 29–33), und das ist die Vorgabe
(`src/data/settings.cpp` 126); die `mse.config` dieses Rechners überschreibt
sie nicht.

**Und wxWidgets stellt an Core Text nichts um.** Die Zeile trägt die Schrift und
`kCTForegroundColorFromContextAttributeName`, sonst nichts:
`kCTLigatureAttributeName`, `kCTKernAttributeName`,
`kCTFontFeatureSettingsAttribute` und `kCTTrackingAttributeName` kommen im
Quelltext von wxWidgets nicht vor, ebenso wenig `CGContextSetShouldSmoothFonts`
und `CGContextSetAllowsFontSmoothing` — nachgesehen am 2026-09-23 über die
Codesuche von GitHub, mit der Gegenprobe, dass sie
`kCTForegroundColorFromContextAttributeName` in `src/osx/carbon/font.cpp` und
`src/osx/carbon/graphics.cpp` findet. **Was Core Text von sich aus tut, tut es
also auch für MSE2.**

## Übersicht

| Aspekt | MSE2 | hier | Stand |
| --- | --- | --- | --- |
| Datei, Fassung, Hash | acht Familien aus dem Pack | dieselben acht Dateien | **gleich** |
| Shaping | Core Text: Ligaturen `ffi` `ffl`, keine Kontextalternativen | `liga` seit #1805 | **gleich** — 25.102 von 25.102 Namen, Breite und Glyphenzahl |
| Kerning | Core Text, `PairPos` Format 1 und 2 | beide Formate seit QA-03 | **gleich** — 25.102 von 25.102 Namen |
| Zeilenumbruch | Vorsilbenbreiten auf ¼ Punkt | portiert | **gleich** |
| Grundlinie | Stück auf ¼ Punkt, darin die Oberlänge von Core Text | seit QA-03 auf ¼ Punkt gezeichnet | **nahe** — siehe unten |
| Rückfall | Core Text weicht still aus | Ersatzschrift (#515) und ein Befund | **selten** — 2 Zeichen auf 2 Karten |
| Hinting | keines | keines | **gleich** |
| Kantenglättung | vierfach mit Kastenfilter, darin Quartz' Schriftglättung | vierfach mit Kastenfilter, die Glättung seit #1806 als gemessenes Modell | **nahe** — 0,18 statt 1,92 an einer Regeltextzeile |
| Stellung der Glyphen | ¼ Punkt | vorher ganze Punkte, seit QA-03 ¼ Punkt | **behoben** |

## Datei, Fassung, Hash

Die Pläne der festen Kohorte fordern **acht** Familien an, und alle acht nimmt
dieser Rasterer aus dem Pack. MSE2 findet sie unter
`~/Library/Fonts/mse2-referenz` als **dieselbe Datei** —
`TestDieReferenzschriftenReichen` vergleicht den Hash und schreibt Datei,
Fassung und Hash seit QA-03 ins Protokoll:

| Familie | Datei | Fassung laut Namenstabelle |
| --- | --- | --- |
| Beleren Bold | `beleren-bold_P1.01.ttf` | Version P1.01 |
| Beleren Small Caps, … Bold | `belerensmallcaps-bold.ttf` | Version 1.000 |
| MPlantin | `mplantin.ttf` | Converted from C:\1E\PLA.TF1 by ALLTYPE |
| MPlantin-Italic | `mplantinit.ttf` | Converted from C:\1E\PLAI.TF1 by ALLTYPE |
| Matrix | `MATRIX.TTF` | 001.001 |
| ModMatrix | `ModMatrix.ttf` | Version 1.0 |
| Relay-Medium | `relay-medium.ttf` | Version 1.0; 2002; initial release |

**Wie MSE2 einen Namen auflöst, prüft der Lauf nicht** — dort Core Text über
wxWidgets, hier die Namenstabelle. Das Manifest jeder Referenz trägt den Hash
über alle installierten Schriften (`schriften`); ein anderer Satz Schriften ist
damit eine andere Referenz.

## Shaping

**Core Text setzt die Standardligaturen.** Beleren Bold und Beleren Small Caps
führen je zwei — `ffi` und `ffl` (`GSUB`, Merkmal `liga`, Lookup-Art 4) —, und
Core Text nimmt sie: `Officious` wird aus neun Zeichen sieben Glyphen,
`Baffling End` aus zwölf zehn. **Im Bild ist es zu sehen**: MSE2s
`Armored Griffin` (White) trägt über dem zweiten `i` keinen eigenen Punkt.

**Kontextalternativen setzt Core Text dagegen nicht**, obwohl beide Schriften
`calt` führen (13 und 58 Regeln — etwa ein anderes `K` vor `y` oder vor einer
schließenden Klammer): `Ky`, `K)`, `n x` und `A&B` kommen mit den Glyphen aus
der Zeichentabelle zurück. Ebenso wenig `fina`, `init` und `c2sc` der
Kapitälchen und `frac` von ModMatrix. Die übrigen Schriften der Kohorte haben
kein `GSUB` oder keines mit einem Merkmal, das von selbst gilt.

**Bis #1805
ersetzte dieser Rasterer nichts.** Über die Kartennamen des Kanons trägt das
91 von 25.102 Namen eine Ligatur — die Glyphe sah anders aus, und der Name
stand bei 16 Punkt um **1,5 bis 2,4 Punkt breiter** (im Mittel 1,64): die
Ligatur ist schmaler als `f`, `f` und `i` samt ihrem Kerning, denn Beleren
rückt `f` und `i` um 180 Einheiten auseinander.

**Jetzt setzt er `liga`** (`internal/render/ligaturen.go`), Lookup-Art 4, auch
hinter einer Erweiterung — und nur `liga`: `rlig`, `clig` und `ccmp` führen im
Pack nur Schriften für arabische Präsentationsformen und kombinierende Marken,
und keine Schrift des Kanons. Eine Ligatur reist als **Glyphenzeichen** durch
den Satz, eine Rune jenseits von Unicode mit der Glyphennummer darin
(`ligatursatz.go`); Vorsilbe, Zeilenbreite, Textbreite, Zeichner und Symboltext
nehmen den Zug alle aus derselben Funktion. Gezeichnet wird die Glyphe nach
ihrer **Nummer**, auf dem Weg von `opentype.Face.Glyph` — bildpunktgleich mit
der Bibliothek über jede Glyphe von Go Regular bei fünf Größen
(`TestEineGlypheNachNummerGleichtDerNachZeichen`).

Gemessen über **alle 25.102 Kartennamen** gegen Core Text bei 16 Punkt, Breite
auf 0,001 Punkt **und** Glyphenzahl: Beleren Bold 25.102, Beleren Small Caps
25.102, Relay-Medium und ModMatrix je 25.102 — die 91 Namen mit Ligatur
eingeschlossen. `TestDieNamensschriftSetztWieCoreText` hält zwanzig davon hinter
`fremdpakete`, über den Weg, den der Rasterer geht.

## Kerning

**Core Text kernt aus `GPOS`, in beiden Formaten der Paartabelle, und in der
Reihenfolge der Lookups.** Bei Beleren liegen hinter zwei Untertabellen mit
Einzelpaaren (Format 1) Klassenpaare (Format 2); hier stand bis QA-03 nur der
erste Teil, und die Aussage im Leser, die Namensschrift brauche den zweiten
nicht, war falsch.

Gemessen über **alle 25.011 Kartennamen ohne Ligatur** gegen Core Text bei 16
Punkt:

| Leser | Breite gleich (< 0,001 pt) | im Mittel daneben | am weitesten |
| --- | ---: | ---: | ---: |
| nur Format 1 (bis QA-03) | 2.287 | 1,8 pt | 18,5 pt |
| beide Formate, Lookup-Reihenfolge | **25.011** | — | — |

Ebenso Beleren Small Caps (25.011 von 25.011), Relay-Medium und ModMatrix
(25.102 von 25.102). Seit #1805 stimmen auch die 91 Namen mit Ligatur, in
beiden Beleren-Schnitten (oben); `TestDieNamensschriftSetztWieCoreText` hält den
Leser hinter `fremdpakete` gegen zwanzig dieser Breiten.

**GPOS und die alte `kern`-Tabelle zählen nicht doppelt.** `vanilla whale.ttf`
und vierzehn Systemschriften führen in beiden dieselben Paare; Core Text zählt
sie einmal. Welche gilt, wenn sie sich widersprechen, zeigt keine davon —
festgelegt nach HarfBuzz (`hb_ot_shape_planner_t::compile`): die alte Tabelle
nur ohne `GPOS`-Kerning.

**Und ein Paar verschob das falsche Zeichen.** Der Vorschub eines Zeichens trägt
die Korrektur zu seinem Vorgänger; gezeichnet wurde die Glyphe aber vor dem
ganzen Vorschub. Das Paar `AV` rückte damit nicht das `V` an das `A`, sondern
das Zeichen danach. Die Breiten stimmten, deshalb sah es kein Breitentest —
`TestEinKerningpaarVerschiebtSeinZweitesZeichen`.

## Zeilenumbruch

**Portiert und nicht neu vermessen.** MSE2 misst jede Vorsilbe eines Abschnitts
als Ganzes (`FontTextElement::getCharInfo`, `src/render/text/font.cpp` 27–49)
und rastert sie auf ¼ Punkt; `vorsilbenbreite.go` tut dasselbe, und
`TestDieBreiteTrifftWx` hält die Rundung gegen wxWidgets. Seit das Kerning
stimmt, stimmen die Vorsilben mit Kerning mit, und seit #1805 auch die mit
einer Ligatur.

**Eine Grenze bleibt, und sie ist eng**: bricht ein Wort mitten in einer
Ligatur um, misst MSE2 die Vorsilben bis dahin mit der Ligatur und zeichnet
beide Zeilen ohne sie. Hier wird ein Zug erst gesetzt, wenn er feststeht — das
Wort bricht vor der Ligatur. Ein Kartenname bricht nicht um, und im Regeltext
setzt MPlantin keine Ligatur.

## Grundlinie

MSE2 legt jedes Stück an eine Oberkante, die auf ¼ Punkt rastet
(`ysub = int(text_scaling * (pos.y - yi))`), und Core Text setzt die Grundlinie
um seine Oberlänge darunter. Hier rastet die Oberlänge wie MSEs gemessene Höhe
(`gerastertesTextmass`), und gezeichnet wurde bis QA-03 auf `int(grundlinie)` —
**abgeschnitten auf ganze Punkte**.

`TestDieZeilenlageGegenMSE2` misst die Tintenbänder. An `Bow to My Command`
(Archenemy) standen vorher alle sieben Bänder einen Punkt über MSE2s; seit
QA-03 stehen vier davon **genau** auf ihnen, das erste noch einen Punkt höher.
`Dance, Pathetic Marionette` bleibt um zehn Punkte versetzt — das ist die
Zeilenhöhe der Vorlage aus #1029 und keine Frage der Rasterung.

## Rückfall

MSE2 fragt wxWidgets nach einem Namen, und Core Text weicht **still** auf eine
andere Schrift aus, wenn die Datei fehlt oder ihr ein Zeichen fehlt. Hier gibt
eine fehlende Schrift einen Befund der Art Umgebung, und ein fehlendes Zeichen
kommt aus der Ersatzschrift (#515), ebenfalls benannt.

**Im Kanon kommt das kaum vor**: MPlantin führt jedes Zeichen aller Regeltexte
der zehn Sets; MPlantin-Italic fehlen im Flavortext zwei — U+0336 (dreizehnmal
auf einer Karte) und U+2015 (einmal auf einer). Die Namensschrift führt jedes
Zeichen aller 25.102 Namen.

## Hinting

**Keines, auf beiden Seiten.** Dieser Rasterer liest die Umrisse mit
`font.HintingNone` (`glyphen.go`, das Gesicht zu einer Größe). Dass Core Text
hier ebenfalls nicht ans Raster zieht, ist gemessen und nicht angenommen:

- **Die Vorschübe sind linear.** Alle 25.011 Namen treffen die ungehinteten
  Vorschübe mal Größe auf ein Tausendstel Punkt; gehintete Vorschübe rasteten
  auf ganze Bildpunkte.
- **Die Umrisse auch.** Unsere ungehinteten Glyphen, auf MSEs vierfachem Weg
  gezeichnet, treffen Core Text auf demselben Weg **ohne** Schriftglättung auf
  **0,065 von 255** im mittleren Betragsabstand — eine Regeltextzeile in
  MPlantin bei 14 Punkt. Gitteranpassung verschöbe Stämme um ganze Bildpunkte
  der vierfachen Größe.

## Kantenglättung

Zwei Dinge, und nur eines gehört MSE2.

**MSEs eigene Glättung ist der vierfache Weg** — zeichnen, dann zweimal
ganzzahlig mitteln; beim Stauchen rechnet es mit einem Faktor, der fetter macht
(`downsample_to_alpha`, Zeilen 26–140). **Er ist seit QA-03 portiert**
(`textraster.go`), samt dem Puffer von `int(breite) + 3` mal `int(hoehe) + 1`
Punkten, an dessen Rand abgeschnitten wird.

**Die zweite gehört Quartz.** Ein Bitmapkontext, den wxWidgets anlegt, glättet
Schrift: ohne Angabe gezeichnet ist gleich mit `setShouldSmoothFonts(true)`
gezeichnet, und das legt Tinte zu. Gemessen an derselben Regeltextzeile auf
MSEs Weg:

| gezeichnet | Tinte | volle Punkte | Abstand zu MSE2s Weg |
| --- | ---: | ---: | ---: |
| Core Text, Vorgabe (= mit Glättung) | 100 % | 100 % | 0 |
| Core Text, ohne Glättung | 83,9 % | 69,0 % | 1,92 |
| **dieser Rasterer bis QA-03** — ganze Punkte, ungebrochen | 84,5 % | 57,3 % | **4,42** |
| dieser Rasterer, Glyphen an gebrochener Stelle, einfach | 84,4 % | 70,3 % | 2,35 |
| **dieser Rasterer seit QA-03** — MSEs Weg, `vierfachlauf` | 83,9 % | 68,3 % | **1,92** |
| **seit #1806** — dazu die Verdickung aus [0130](entscheidungen/0130-quartz-glaettung-nachgebildet.md) | 98,8 % | 96,6 % | **0,18** |

Alle Zeilen gegen dieselbe Referenz: `raster.swift` mit `vorgabe`, eine Zeile
MPlantin bei 14 Punkt, »Flying, first strike. Whenever this creature
attacks«. Der Abstand ist das Mittel der Beträge über alle 420 × 44
Bildpunkte, in Stufen von 0 bis 255.

**Das beantwortet eine alte Frage**: `glyphendeckung_intern_test.go` misst seit
dem 2026-08-18, dass unsere Glyphen rund 86 % der Tinte MSE2s setzen und in der
obersten Stufe 62 % der Punkte — und die Gammakorrektur, die das angliche, wurde
verworfen, weil der Abstand dabei stieg. Die Lücke ist Quartz' Glättung: ohne
sie liegt Core Text selbst bei 84 % und 69 %.

**Bis #1806 wurde sie nicht nachgebaut** — kein Verhalten von MSE2, sondern
eines Rasterers, dessen Rechnung nicht offenliegt. **Entschieden ist, sie
nachzubilden**, und zwar als Gleichung aus einer Messung
([0130](entscheidungen/0130-quartz-glaettung-nachgebildet.md),
`internal/render/glaettung.go`). Gemessen mit `scripts/coretext/glaettung.swift`
an MPlantin, Beleren Bold und Relay-Medium, jeweils quer durch einen Stamm:

| Geviert (Bildpunkte) | links | rechts | oben | unten |
| ---: | ---: | ---: | ---: | ---: |
| 8 | 0,125 | 0,129 | 0,255 | 0 |
| 16 | 0,251 | 0,251 | 0,506 | 0 |
| 19 | 0,298 | 0,298 | 0,596 | 0 |
| 24 bis 128 | 0,30 | 0,30 | 0,60 | 0 |

(Relay-Medium; die beiden anderen auf ±0,005 gleich, bis auf MPlantin unter 12
Bildpunkten, wo ein dünner Stamm in einem Bildpunkt klemmt.)

**Drei Dinge daran waren nicht zu erraten.** Die Glättung ist eine
**Verdickung**, keine Gammakurve: quer durch einen Stamm rücken beide Kanten,
der Verlauf dazwischen bleibt. Sie wirkt **senkrecht nur nach oben** — die
Unterkante eines waagerechten Stamms bleibt stehen. Und sie ist
**achsparallel**: ein gedrehter Stamm legt bei 30 Grad das 1,38-fache der
Tinte zu, die er ungedreht zulegt, und `|cos| + |sin|` gibt 1,37. Der erste
Anlauf verdickte nach beiden Seiten gleich und kam nur auf 0,90; die Reste
saßen an den Oberkanten und auf der Grundlinie.

An den drei Probezeilen, gegen Core Text mit Glättung:

| Zeile | ohne | mit |
| --- | ---: | ---: |
| MPlantin, 14 Punkt | 1,92 | **0,18** |
| Beleren Bold, 16 Punkt | 2,60 | **0,32** |
| MPlantin, 6 Punkt | 1,74 | **0,54** |

`TestDieGlaettungTrifftQuartz` hält die ersten beiden hinter `fremdpakete`
gegen Core Text selbst. **Was das Modell nicht trägt** — gedrehten Text, die
Einstellung des Systems, sehr kleine Größen — steht im Vermerk.

## Was die Portierung am Kanon bringt

`TestDerAbstandZuMSE2AmKanon`, feste Kohorte, 4.265 Karten aus zehn Sets —
ab dem Beleg 4.266, siehe unten —, gegen die Referenz vom 2026-09-23, je
Schritt ein eigener Lauf:

| Stand | Kanonmittel | Textebenen | Bildebenen | Symbolzüge |
| --- | ---: | ---: | ---: | ---: |
| vor QA-03.AK-02 (`68085725`) | 20,051 | 75,15 | 13,42 | 64,91 |
| Kerning mit Klassenpaaren (`5884843e`) | 19,661 | 73,52 | 13,15 | 64,91 |
| dazu MSEs vierfacher Weg (`126074c7`) | 14,468 | 49,39 | 11,25 | 64,65 |
| dazu die Ligaturen (#1805) | 14,465 | 49,38 | 11,25 | 64,65 |
| **dazu die Glättung** (#1806, Beleg `7972c05`) | **13,792** | **45,46** | 11,02 | 64,53 |

Die Spalten je Ebenenart sind der mittlere Abstand **innerhalb** der
Bildpunkte dieser Art (MESSMETHODIK.md); Text deckt 4 % der Fläche und trug
vor QA-03 56 % des Abstands, nach dem Vierfachweg 43 %, mit der Glättung 42 %. **Auch die Bildebenen rücken
näher**: die Zerlegung ordnet einen Bildpunkt der Ebene zu, die ihn bei uns
zuletzt gemalt hat — wo MSE2 Text zeichnete und wir daneben, zählte der
Unterschied zum Bild.

**Je Satz**, Kanonmittel vor → nach:

| Satz | vorher | Kerning | Vierfachweg | Ligaturen | Glättung |
| --- | ---: | ---: | ---: | ---: | ---: |
| Archenemy | 14,19 | 14,19 | 12,21 | 12,21 | 12,10 |
| Black | 18,87 | 18,59 | 13,38 | 13,38 | 12,68 |
| Blue | 19,78 | 19,41 | 13,99 | 13,99 | 13,29 |
| Colorless | 21,85 | 21,33 | 17,58 | 17,58 | 17,17 |
| Green | 19,49 | 19,07 | 13,64 | 13,64 | 12,93 |
| Multicolor | 21,05 | 20,71 | 15,39 | 15,40 | 14,75 |
| Planechase | 13,12 | 12,79 | 10,48 | 10,48 | 10,31 |
| Red | 19,56 | 19,23 | 13,96 | 13,95 | 13,25 |
| Vanguard | 20,72 | 17,84 | 12,68 | 12,68 | 11,85 |
| White | 20,46 | 20,03 | 14,13 | 14,11 | 13,30 |

Vanguard gewann schon am Kerning am meisten — dort stehen große Namen in
Beleren. **Die Ligaturen bewegen das Mittel kaum** (91 von 25.102 Namen tragen
eine), **die Glättung um 0,67**. **Was bleibt, ist ungleicher verteilt**: der
Median fiel von 19,35 auf 12,44, die Standardabweichung stieg von 4,24 auf 5,42.

**Und die schlechtesten Karten werden mit der Glättung schlechter**, je Satz
um 1 bis 3 — *Garbage Elemental* von 46,5 auf 49,1, *Fungal Behemoth* von
45,7 auf 48,4. Das ist kein Fehler der Glättung, sondern ihre Folge an einem
anderen: **acht der neun** stehen in der Versatzprobe des Texts, ihr Text sitzt
also mindestens einen Punkt neben dem von MSE2, und kräftigerer Text an
falscher Stelle kostet mehr. Die Karten, die jetzt noch weit weg sind, sind es
aus anderen Gründen als der Schrift.

**Die Tinte stimmt jetzt.** `TestWieVielTinteUnsereGlyphenSetzen` zählt seit
dem 2026-08-18 über 671 Karten aus `White`, wie viel Tinte unsere Glyphen im
Regeltextkasten setzen: **86,3 %** von MSE2, in der obersten Deckungsstufe
**61,9 %**. Mit der Glättung sind es **99,0 %** und **96,9 %**.

**Und eine Karte fehlte in allen drei Läufen**: `Phoebe, Head of S.N.E.A.K.`
suchte ihr Referenzbild unter dem falschen Namen — MSE reinigt
`{card.name}.png` als Ganzes, der Schlusspunkt bleibt stehen. Seit `5aaee8d2`
findet der Lauf sie; der Beleg misst deshalb eine Karte mehr.

**Und was es kostet: nichts Messbares.** `BenchmarkWegRastern` — eine
Kartenvorschau aus Black, 30 Durchläufe — braucht vor QA-03 56 ms, nach dem
Vierfachweg 55 ms und mit Ligaturen und Glättung 54 bis 55 ms, bei einer
Schwelle von 120 ms. Text trug schon vorher unter eine
Millisekunde je Karte; vierfach gerastert bleibt er klein gegen die Bilder.

## Wie gemessen wurde

| Frage | Werkzeug |
| --- | --- |
| Breite und Glyphenzahl einer Zeile bei Core Text | `swift scripts/coretext/breiten.swift <schrift> <groesse> < zeilen` |
| Deckung einer Zeile auf MSE2s Weg | `swift scripts/coretext/raster.swift …` (Aufruf im Kopf) |
| dieselbe Schrift, dieselbe Datei | `TestDieReferenzschriftenReichen` hinter `fremdpakete` |
| Kerning und Ligaturen gegen Core Text | `TestDieNamensschriftSetztWieCoreText` hinter `fremdpakete` |
| Tintenbänder gegen MSE2s Bild | `TestDieZeilenlageGegenMSE2` hinter `fremdpakete` |
| Abstand je Ebenenart über den Kanon | `TestDerAbstandZuMSE2AmKanon` hinter `fremdpakete` |

**Die Reichweite der Rastermessung ist eine Zeile** in einer Schrift bei einer
Größe. Was sie über den Kanon sagt, sagt erst der Kanonlauf darüber.
