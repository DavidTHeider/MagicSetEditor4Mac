# Serienarbeit — was die hundertste Karte kostet

**Wofür diese Datei da ist.** Die Abnahme von Phase 4.3
(#1290) lautet:
*„Eine neue Karte mit denselben Grundangaben wie die vorige braucht höchstens X
Nutzerhandlungen und keinen Seitenwechsel — X wird in der Baseline festgelegt
und danach unterschritten oder begründet."* Hier steht die Baseline, hier steht
X, und hier steht, womit gemessen wurde.

**Eine Zahl ohne ihre Erhebung ist keine Zahl** (§10). Jede unten trägt deshalb
ihr Datum, ihr Maß und ihren Gegenstand.

---

## Wie gemessen wird

| | |
| --- | --- |
| **Fenstermaß** | 1280 × 800 — das kleinste, das ein aktuelles MacBook Air im Standardmaßstab zeigt |
| **Gegenstand** | `Vanguard.mse-set`, 107 Karten, Stil `magic-m15-altered`, Kartenlasche mit aufgeschlagener Karte |
| **Weg** | `wails dev` (seit ENG-02 `make dev-browser`), ein Chrome ohne Fenster an `34115`, Tasten über `Input.dispatchKeyEvent` — also **echte** Tastendrücke durch die laufende Brücke, nicht `fireEvent` gegen eine Vorrichtung |
| **Was eine Nutzerhandlung ist** | ein Klick oder ein Tastendruck. Scrollen zählt als eine, gleich wie weit |

**Was diese Zahlen nicht sind.** Sie sind an **einem** Set auf **einem** Rechner
erhoben. Eine Vorlage mit dreißig Feldern statt neunzehn gibt andere; die
Verhältnisse bleiben.

---

## Der Stand nach dem ersten Werkstück, am 2026-08-30

**Diese Tabelle gehört zu #1320**
und nicht zum heutigen Stand — die Abschnitte darunter haben die Zahlen seitdem
zweimal verändert. Wer den **heutigen** Stand sucht, liest die Seitentabelle am
Ende.

| | vorher | jetzt |
| --- | ---: | ---: |
| Schaltflächen | 67 | **79** |
| fokussierbare Elemente | 108 | **129** — und nach #1323/#1328 **64**, siehe unten |
| Seitenhöhe | 2152 px | **2190 px** |
| davon Bildschirme | 2,7 | **2,7** |
| Textfelder im Formular | 19 | 19 |
| davon **über der Falz** | 5 | **4** |
| **Weg zur nächsten Karte** | **keiner** | ⇟, ein Knopf bei 331 px |

Die beiden ersten Zahlen sind **gewachsen**, und das gehört gesagt: die zwei
Blätterknöpfe kommen dazu, und dieser Lauf sieht ein vollständiger geladenes
Formular als der vom 2026-08-28. Was zählt, steht in der letzten Zeile.

---

## Kernweg „anlegen, füllen, weiter"

### Nächste Karte

**Vorher: es gab keinen Weg.** Gemessen am 2026-08-30, vor der Änderung: die
Suche nach „nächste / weiter / next" über **alle** Schaltflächen und Verweise
der Seite gab **null** Treffer, und `frontend/src/rahmen/tastatur.ts` kannte
⌘K ⌘F ⌘Z ⌘S ⌘O ⌘N ⌘W ⌘P — kein Kürzel für Karten.

Wer weiterwollte, tat dies:

| | Handlung |
| ---: | --- |
| 1 | zur Liste scrollen (das Formular reicht bis 2 152 px, die Liste beginnt bei 344) |
| 2 | die nächste Karte anklicken |
| 3 | zurück ins Feld klicken — der Fokus lag danach am `body` |

**Drei Handlungen, und die Bildlage des Formulars war weg.**

### Jetzt: **eine**

| | Handlung |
| ---: | --- |
| 1 | ⇟ |

Gemessen an der laufenden Brücke am 2026-08-30, Fokus im Feld `name`:

```
vorher:  {"fokus":"name","name":"Akroma, Angel of Wrath Avatar","stelle":"1 von 107"}
nach ⇟:  {"fokus":"name","name":"Arcanis, the Omnipotent Avatar","stelle":"2 von 107"}
```

Der Fokus steht danach im **selben** Feld der nächsten Karte, und die Seite
scrollt nicht (`scrollY` bleibt 0).

**Damit ist X = 1** für den Schritt *weiter*.

### Der ganze Weg: **X = 4**

Die Abnahme aus #1290 verlangt eine Zahl fuer den **ganzen** Weg. Gemessen am
2026-08-30 mit `node scripts/kernwegmass.mjs`:

```
Kernweg "anlegen, fuellen, weiter"
  Nutzerhandlungen  4
  Seitenwechsel     keiner
  Karten im Set     2 - Feuerball, Blitzschlag
  aufgeraeumt       das Messset ist weg
```

| | Handlung |
| ---: | --- |
| 1 | *Karte anlegen* |
| 2 | in das Namensfeld |
| 3 | den Namen tippen |
| 4 | ⇞ |

**Und beide Karten tragen ihren Namen** — das ist der Teil, der die Zahl erst
zu einer Aussage macht.

**Der Lauf weist ab, er zeigt nicht nur an.** Stimmt eine der vier Zusagen
nicht — vier Handlungen, kein Seitenwechsel, zwei Karten, beide Namen —, endet
er rot. Angestrichen vom Nachlauf über Phase 4.3, Befund 2: ein Messlauf, der
eine falsche Zahl druckt und mit 0 endet, ist eine Anzeige und keine Abnahme.

**Was er dabei nicht prüft**, und das gehört dazu: „dieselben Grundangaben wie
die vorige" heißt bei ihm **ein** Feld. Ein Weg, der zehn Felder überträgt,
kostet mehr als vier Handlungen, und diese Zahl sagt darüber nichts.

### Was auf dem Weg dorthin gefunden wurde

**⇞ und ⇟ haben das Getippte verworfen.** Ein Feldwert wird beim **Verlassen**
geschrieben (`Feldeingabe.tsx`, `onBlur`), und das Blättern verließ das Feld
nie: der Editor wurde mitsamt dem Eingabefeld abgeräumt, und mit ihm die
Eingabe.

Gemessen an der laufenden Brücke, in einem eigens angelegten Set:

```
A · tippen, dann blur:   ["MitBlur",""]
B · tippen, dann ⇟:      ["MitBlur",""]
```

Im Fall B stand »OhneBlur« im Feld und **nichts** im Bestand. Damit verlor
genau der Weg Arbeit, für den es dieses Kürzel gibt.

**Gefunden hat es der Zweitblick über Phase 4.3 — mittelbar.** Sein Befund 1
lautete: *„Die dokumentierte Abnahme wird nicht gemessen."* Er hatte recht, und
beim Messen des **ganzen** Weges statt eines Schrittes fiel der Datenverlust
auf. Ein Erfolgskriterium, das man verengt, verbirgt genau das, was zwischen
seinen Teilen liegt.

### Drei Fehler im Messgerät, bevor es maß

Sie stehen hier, weil jeder von ihnen eine plausible Zahl ergeben hätte:

| Fehler | Was er gezeigt hätte |
| --- | --- |
| Fokus nach fester Frist statt nach dem Formular | „4 Handlungen" — getippt wurde ins Leere |
| `dispatchKeyEvent` mit `type: "char"` | nichts kam an; ein Chrome ohne Fenster hat keinen Fokus auf der Seite |
| `element.click()` ohne `focus()` | der Wert blieb ungeschrieben, weil kein Mausklick ihn verließ |
| ein einmaliges `focus()` mit Warteschleife | der Lauf war **flatterhaft**: das Formular baute sich zwischen Fokus und Prüfung neu |

**Und eine Grenze, die bleibt**: `bisRuhe` erklärt eine Seite für fertig, wenn
zwei Messungen im Abstand von einer halben Sekunde dieselbe Knotenzahl geben.
**Das ist eine Heuristik und kein Abschlusssignal** — eine Anfrage, die länger
braucht, ohne den Baum zu ändern, käme durch. Angestrichen vom Nachlauf über
Phase 4.3, Befund 4; ein echtes Signal müsste die Anwendung selbst geben.

**Und einer im Bestand**: der erste Anlauf lief gegen den echten Bestand dieses
Rechners und legte zwei leere Karten in `Vanguard.mse-set` an. Sie sind
gelöscht, das Set steht wieder bei 107. `kernwegmass.mjs` legt sich seitdem
sein eigenes Set an und räumt es weg — und sagt es, wenn das misslingt.

---

## Der Weg dorthin: 85 Schritte, dann 20 — und was davon vergleichbar ist

**Wer mit der Tastatur beginnt, brauchte 85 Tabulatorschritte bis zum ersten
Kartenfeld.** Gemessen am 2026-08-30 als Stelle des Feldes `name` in der Liste
aller fokussierbaren Elemente. Aufgeteilt:

| Bereich | Schritte |
| --- | ---: |
| **Kartenliste** | **54** |
| Laschen | 13 |
| Kopfzeile und Werkzeuge | 11 |
| Seitenleiste | 7 |

Die Liste war zwei Drittel des Weges: 27 gezeichnete Zeilen mal zwei
Bedienelemente. Und das sind nur die gezeichneten — bei 3.000 Karten stünden
dort dieselben 54.

**Jetzt ist die Liste eine Tabstelle**, wie jede Liste auf jeder Plattform:
MSE2 zeigt seine Karten in einem `wxListCtrl` (`card_list.cpp`, Fassung
f9d9356 vom 2021-07-06), und ein solches ist eine. Innerhalb bewegen ↑ und ↓,
Pos1 und Ende springen an die Enden, → und ← führen in die Zeile und zurück.

| | vorher | jetzt |
| --- | ---: | ---: |
| **davon Kartenliste** | **54** | **1** |
| bis zum ersten Kartenfeld | ≥ 73 | **20** |
| fokussierbare Elemente insgesamt | 129 | **64** |

**Die erste Zeile ist die belastbare.** Die zweite trägt ein „mindestens", und
warum, steht gleich darunter.

### Und eine Warnung über das Messgerät

**Die 85 und die 20 sind nicht mit demselben Ausdruck erhoben.**

Der erste zählte `button:not([disabled])` und ähnliche — und ein Knopf mit
`tabindex="-1"` erfüllt das, obwohl der Tabulator ihn **nicht** erreicht. So
lange nichts im Fenster ein `tabindex="-1"` trug, war das dasselbe; seit der
Liste ist es das nicht mehr. Der erste Lauf nach der Änderung gab deshalb
weiter **86**, und das sah aus wie ein wirkungsloser Eingriff.

Der Ausdruck ist berichtigt (`:not([tabindex="-1"])` an jeder Stelle, dazu
`[tabindex="0"]`). Was er über die **Liste** sagt, gilt für beide Stände: dort
trug vorher nichts ein `tabindex="-1"`, also zählen beide Ausdrücke gleich —
54 gegen 1.

Was er über die **Laschen** sagte, war schon vorher falsch: sie sind ein
`tablist` mit wanderndem `tabindex` und waren immer **eine** Stelle, nicht 13.

### Und deshalb steht dort „≥ 73" und nicht „85"

**Die 85 und die 20 sind nicht vergleichbar**, und ein Pfeil dazwischen wäre
eine Behauptung. Was sich vergleichen lässt, ist die Aufteilung:

| Bereich | vorher, berichtigt | jetzt |
| --- | ---: | ---: |
| Kartenliste | 54 | **1** |
| Laschen | **1** (nicht 13) | 1 |
| Kopfzeile und Werkzeuge | 11 | 11 |
| Seitenleiste | 7 | 7 |
| **zusammen** | **73** | **20** |

Die 73 ist **gerechnet und nicht gemessen** — der alte Stand steht nicht mehr
da, und ein zweiter Lauf gegen ihn wäre ein Lauf gegen etwas, das es nicht mehr
gibt. Deshalb „≥ 73": mehr als 73 kann es gewesen sein, weniger nicht.

**Angestrichen vom Zweitblick über Phase 4.3, Befund 2.** Der erste Anlauf
schrieb 85→20 in die Tabelle **und** ins Änderungsprotokoll — und erklärte drei
Absätze weiter selbst, warum das nicht geht.

**Das ist der Satz aus §10 an einem Beispiel**: die Reichweite einer Messung
ist die Reichweite ihrer Aussage — und wer zwei Zahlen vergleicht, prüft
zuerst, ob dasselbe Gerät sie erhoben hat.

## Der erste Weg: **sieben Handlungen**, am 2026-09-03

Erhoben mit `node scripts/erstlaufmass.mjs` — von der Anwendung bis zur ersten
Karte, die ihren Namen trägt. Das ist Werkstück D von Phase 4.10
(#1325) und
die Zahl, an der spätere Änderungen an der Oberfläche gemessen werden.

| Station | Handlungen bis hierher |
| --- | ---: |
| Die Anwendung steht | 0 |
| Die Setliste steht | 1 |
| Der Name steht | 3 |
| Das Set ist offen | 4 |
| Die Karte steht | 5 |
| Die Karte heißt „Meine erste Karte" | **7** |

**Was daran auffällt:**

- **Spiel und Stil kosten nichts.** Beide sind vorgewählt — wie in MSE2
  (`src/gui/new_window.cpp`, Zeile 66, Fassung f9d9356 vom 2021-07-06). Ohne
  diese Vorwahl wären es neun.
- **Vier der sieben sind Tippen und Hinzeigen** — zweimal „ins Feld, dann
  schreiben". Das ist die Untergrenze: zwei Namen sind zwei Namen.
- **Drei Klicks**: in die Sets, Set anlegen, Karte anlegen. Ein Weg über die
  Startseite spart keinen — ihre Kachel „Zu den Sets" ist derselbe Klick.

**Was diese Zahl nicht sagt:**

- **Nichts über den ersten Eindruck.** Gezählt ist der Weg, nicht die Frage,
  ob jemand ihn findet. Dafür steht `SEITENKRITIK.md` (Entwicklungsarchiv).
- **Nichts über einen Bestand ohne Saat.** Der Lauf setzt die acht mitgelieferten
  Vorlagenpakete voraus; ohne sie führte der Weg zuerst über die Paketseite und
  einen Dateidialog, den ein Lauf ohne Fenster nicht bedienen kann.
- **Nicht die Zeit.** Eine Dauer wäre eine Aussage über diesen Rechner.

**Und er räumt sein Set weg.** Ein Messlauf, der schreibt, hinterlässt nichts
im Bestand eines Menschen — dieselbe Regel wie beim Kernwegmaß.

---

## Alle zweiundzwanzig Seiten, am 2026-09-03

Erhoben mit `node scripts/serienmass.mjs` bei 1280 × 800, gegen den Bestand
dieses Rechners. **Erstmals über die ganze Anwendung**: neun Seiten, die an
keinem Set hängen, und die dreizehn Laschen eines offenen Sets.

**Der zweite Teil der Tabelle ist der wichtigere**, und er hat bis heute
gefehlt. Hier stand als Begründung, die Laschen hätten „keinen eigenen Weg in
der Leiste" — das stimmt, und es ist ein Grund für einen anderen **Weg**, nicht
für eine fehlende **Messung**. Dort steht die Karte; dort verbringt jemand
seine Zeit.

| Seite | Tabstellen | Knöpfe | Höhe | Bildschirme | Textfelder | über der Falz |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Start | 22 | 20 | 800 px | 1,0 | 0 | 0 |
| Sets | 31 | 27 | 1 054 px | 1,3 | 1 | 1 |
| Pakete | 32 | **411** | 2 115 px | 2,6 | 4 | 0 |
| Vorlagen | 17 | 14 | 800 px | 1,0 | 0 | 0 |
| Sammlung | 39 | 21 | 1 314 px | 1,6 | 11 | 10 |
| Decks | 28 | 19 | 800 px | 1,0 | 6 | 6 |
| Turniere | 23 | 17 | 800 px | 1,0 | 3 | 3 |
| Einstellungen | 40 | 25 | 4 042 px | 5,1 | 0 | 0 |
| Hilfe | 17 | 14 | 5 262 px | 6,6 | 0 | 0 |

| Lasche | Tabstellen | Knöpfe | Höhe | Bildschirme | Textfelder | über der Falz |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Karten | 73 | 87 | 2 190 px | 2,7 | 19 | 4 |
| **Setangaben** | **107** | 27 | **6 274 px** | **7,8** | **55** | 6 |
| Schlüsselwörter | 20 | 29 | 800 px | 1,0 | 0 | 0 |
| Konsole | 19 | 28 | 800 px | 1,0 | 0 | 0 |
| Zufallspack | 22 | 29 | 800 px | 1,0 | 1 | 1 |
| Symbol | 23 | 35 | 978 px | 1,2 | 0 | 0 |
| **Stil** | **103** | 27 | 3 902 px | 4,9 | 12 | 4 |
| Suche | 21 | 27 | 800 px | 1,0 | 0 | 0 |
| Ersetzen | 22 | 28 | 800 px | 1,0 | 1 | 1 |
| Statistik | 20 | 27 | 1 482 px | 1,9 | 0 | 0 |
| Bilder | 26 | 30 | 1 024 px | 1,3 | 1 | 1 |
| Druckbogen | 25 | 28 | 800 px | 1,0 | 0 | 0 |
| Ausgeben | 22 | 31 | 1 293 px | 1,6 | 0 | 0 |

**Erhoben gegen einen bereinigten Bestand.** Bis zum 2026-09-03 trugen die
Zahlen fünf **Testsets**, deren Herkunftsdateien fehlten; sie sind entfernt.
Im Bestand liegen jetzt zwei **maßgebliche** Sets — `Vanguard.mse-set` und
`Archenemy.mse-set` aus `kanonische-sets/` —, ein Turnier mit vier
Teilnehmern und ein Deck. **Zwei Zahlen sind dadurch gesunken und nicht durch
eine Änderung**: Sets von 1 587 auf 1 054 Punkte (zwei Sets statt fünf) und
die Tabstellen jeder Seite um rund vier (die Setliste ist kürzer).

**Und eine ist gestiegen, weil der Umbau sie sichtbar macht**: die
Paketseite trägt **411 Knöpfe** — je installiertem Paket einen zum Entfernen,
und es sind 331 installiert.

**Die Zahl, die aus der Tabelle heraussteht**, steht im zweiten Teil:
**Setangaben mit 7,8 Bildschirmen, 107 Tabstellen und 55 Feldern in einer
Spalte.** Was das bedeutet, steht in `SEITENKRITIK.md` (Entwicklungsarchiv) — je
Seite ein Befund oder das ausdrückliche Gegenteil.

**Drei Zahlen sind seit der ersten Messung gefallen, und zwar durch Arbeit:**

| | vorher | nachher |
| --- | ---: | ---: |
| Kartenlasche — erste Karte | 768 px | **344 px** |
| Setseite — erstes Set | 915 px | **157 px** |
| Stillasche — Höhe | 4 489 px | **3 902 px**, und die Vorschau steht bei 428 statt unter allen Optionen |

**Die Laschen tragen 23 Grundknöpfe, die keiner Lasche gehören.** Jede Zeile
des zweiten Teils zählt die Seitenleiste, die dreizehn Reiter und den Setkopf
mit; „Konsole: 23 Knöpfe" heißt deshalb **null eigene**. Wer zwei Laschen
vergleicht, zieht 23 ab.

---

## Und die ersten neun, am selben Tag

Dieselbe Messung, bevor die Laschen dazukamen — sie steht hier, weil die
Zahlen zwischen zwei Läufen **schwanken** und der Abstand zeigt, wie sehr:

| Seite | Tabstellen | Knöpfe | Höhe | Bildschirme | Textfelder | über der Falz |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Start | 19 | 18 | 800 px | 1,0 | 0 | 0 |
| Sets | 45 | 40 | 1 981 px | 2,5 | 2 | 1 |
| Pakete | 19 | 14 | 1 249 px | 1,6 | 4 | 2 |
| Vorlagen | 12 | 10 | 800 px | 1,0 | 0 | 0 |
| Sammlung | 34 | 17 | 1 314 px | 1,6 | 11 | 10 |
| **Decks** | 20 | 12 | 800 px | 1,0 | 6 | 6 |
| **Turniere** | 16 | 11 | 800 px | 1,0 | 3 | 3 |
| Einstellungen | 23 | 13 | 3 097 px | 3,9 | 0 | 0 |
| Hilfe | 12 | 10 | 5 262 px | 6,6 | 0 | 0 |

**Die Kartenseite fehlt auch hier**, aus demselben Grund wie unten: sie
braucht ein Set und eine aufgeschlagene Karte.

### Was daran auffällt

- **Vier Seiten passen auf einen Bildschirm** — Start, Vorlagen, Decks,
  Turniere. Das ist keine Auszeichnung, sondern eine Frage: sie sind so kurz,
  weil der Bestand dieses Rechners dort wenig hat.
- **Sammlung: zehn von elf Textfeldern über der Falz**, bei fast leerem
  Bestand. Bei Decks sind es sechs von sechs, bei Turnieren drei von drei —
  **das ist gut**: wer etwas anlegen will, sieht das Formular sofort.
- **Einstellungen: 3,9 Bildschirme ohne ein einziges Textfeld.** 23
  Tabstellen für lauter Schalter und Auswahlfelder.
- **Hilfe: 6,6 Bildschirme.** Ein Dokument, kein Formular — eine Auskunft und
  noch kein Befund.
- **Sets: 45 Tabstellen.** Die höchste Zahl der Tabelle, und die Seite ist
  2,5 Bildschirme hoch.

### Die Zahlen schwanken, und das gehört dazu

Zwei Läufe hintereinander gaben für **Pakete** 23 und 19 Tabstellen, für
**Einstellungen** 34 und 23. Beide Seiten hängen am Zustand: wie viele Pakete
installiert sind, welche Abschnitte offen stehen. **Eine Zahl aus dieser
Tabelle ist eine Größenordnung, keine Konstante** — und wer sie als Abnahme
benutzt, misst zweimal.

### Und das Messgerät selbst hatte zwei Mängel

**Es maß sieben von neun Seiten** — Decks und Turniere fehlten in seiner
handgeschriebenen Liste, seit es sie gibt. Ein Wächter hält sie jetzt gegen
`navigation.ts` (siehe unten).

**Und es lud vor jeder Seite neu**, was in der Entwicklungsfassung
unzuverlässig ist: die Wails-Brücke muss danach neu aushandeln, und `wails
dev` meldet dann „Unknown message from front end: runtime:ready" — die Seite
bleibt leer. Der Lauf brach an **wechselnden** Stellen ab, und seine Meldung
nannte dabei eine fehlende Seite („kein Knopf, der mit Turniere beginnt"),
wo eine leere stand.

**Gemessen, nicht vermutet**: ein Probelauf, der einmal lädt und dann durch
alle acht Seiten klickt, hat jede erreicht — mit Überschrift und Knöpfen.
Seither klickt auch das Messgerät, und seine Meldung zählt auf, was dastand.
**Das ist obendrein die ehrlichere Messung**: wer die Anwendung benutzt,
klickt in der Leiste.

---

## Jede Seite, am 2026-08-30

Erhoben mit `node scripts/serienmass.mjs` bei 1280 × 800, gegen den Bestand
dieses Rechners (5 Sets, 331 Vorlagenpakete):

| Seite | Tabstellen | Knöpfe | Höhe | Bildschirme | Textfelder | über der Falz |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Start | 18 | 17 | 800 px | 1,0 | 0 | 0 |
| Sets | 43 | 38 | 1 929 px | 2,4 | 2 | 1 |
| **Pakete** | **25** (vorher 413) | 405 | 2 115 px | 2,6 | 4 | 0 |
| Vorlagen | 10 | 8 | 800 px | 1,0 | 0 | 0 |
| Sammlung | 21 | 9 | 800 px | 1,0 | 8 | 8 |
| Einstellungen | 21 | 16 | 2 471 px | 3,1 | 0 | 0 |
| Hilfe | 10 | 8 | 4 998 px | 6,2 | 0 | 0 |

**Die Kartenseite fehlt in dieser Tabelle**, weil sie ein Set und eine
aufgeschlagene Karte braucht; ihre Zahlen stehen oben.

**Und zwei Seiten fehlten, ohne dass es jemandem auffiel** — gefunden am
2026-09-02 zu Beginn von Phase 4.10
(#1325):
`scripts/serienmass.mjs` führte seine Seiten als **Aufzählung von Hand**, und
die ist beim ersten Lauf geschrieben worden. **Decks und Turniere** kamen in
den Phasen 4.6 und 4.7 dazu, in diese Liste nicht — die Tabelle sah vollständig
aus und war es nie.

Das ist dieselbe Falle, vor der `AGENTS.md` §3 beim Barrierefreiheitstest
ausdrücklich warnt: *„eine zu kurze Aufzählung ist grün, und niemand sieht
es."* Seither hält
[`internal/werkzeuge/serienmassseiten_test.go`](../internal/werkzeuge/serienmassseiten_test.go)
die Liste gegen `anwendungsschluessel` in `navigation.ts`.

**Sein erster Anlauf war selbst hohl**, und das gehört dazu: er verglich
`name:` auf beiden Seiten — nur führt `navigation.ts` gar keine Namen, die
stehen im Sprachkatalog (§9). Die Schleife lief über eine leere Liste, der
Wächter war grün, und erst die Gegenprobe zeigte es: eine entfernte Seite ließ
ihn kalt. Jetzt vergleicht er **Schlüssel**, und die Gegenprobe macht ihn rot.

### Was daran auffällt

- **Pakete hatte 413 Tabstellen** — dieselbe Gestalt wie die Kartenliste vor
  #1323: ein
  Knopf je Paket, und hier sind es 331. Wer mit der Tastatur an das Ende der
  Seite wollte, drückte vierhundertmal Tabulator. **Jetzt sind es 25.**

  `Paketzeilen.tsx` hatte sich ausdrücklich gegen einen wandernden Fokus
  entschieden — „am Ende dieselbe Bedienung, die eine Tabelle mit
  Umschaltknöpfen von selbst hat: `Tab` läuft die Knöpfe ab." Der Satz stimmt,
  und sein **Preis** war nicht bekannt: bei fünf Zeilen ist das eine Bedienung,
  bei 331 eine Sperre.
- **Hilfe: 6,2 Bildschirme.** Das ist ein Dokument, kein Formular; die Zahl ist
  eine Auskunft und noch kein Befund.
- **Sammlung: alle acht Felder über der Falz** — bei leerem Bestand. Mit
  fünftausend Exemplaren sieht es anders aus, und das misst Phase 4.5.
- **Einstellungen: 3,1 Bildschirme** ohne ein einziges Textfeld.

**Was diese Zahlen nicht sagen**: ob eine Seite gut aussieht oder ob jemand
findet, was er sucht. Das ist Phase 4.10
(#1325) — sie
beginnt genau hier, mit diesen Zahlen.

## Was noch offen ist

### Die Falz — die Ordnung stimmt, die Höhe nicht

Das Werkstück in #1290 heißt: *„Was ein Kartenformular zuerst zeigt, entscheidet
die Vorlage, nicht die Reihenfolge im Paket."*

**Die Ordnung war schon richtig**, und zwar seit
#674.
`angeboteneFelder` ordnet die Felder über die **Flächen des Zeichenplans** —
also über die Vorlage — und nicht über das Spielpaket. Der Kommentar dort nennt
genau das Problem, das das Werkstück beschreibt: „`magic.mse-game` erklärt
`border color`, `corners` und `card color` vor `name`. In einer senkrechten
Liste heißt das: die ersten drei Felder, die jemand sieht, sind Rahmenfarbe,
Eckenform und Kartenfarbe — und der **Name** steht an vierter Stelle."

**Nachgemessen am 2026-08-30**, mit und ohne einen Eingriff, der dieselbe
Ordnung auf der Go-Seite noch einmal hergestellt hätte:

```
mit Eingriff:  ["name","image","super type","sub type","rule text", …]
ohne Eingriff: ["name","image","super type","sub type","rule text", …]
```

Identisch. Der Eingriff war wirkungslos und ist zurückgenommen — **eine
Änderung, deren Wirkung sich nicht zeigen lässt, ist keine** (§10).

**Und ein literaler Port wäre ein Rückschritt gewesen.** MSEs
`CompareTabOrder` (`card_editor.cpp`, Zeilen 121–140) mischt `0,1·ax` in die
Höhe, um bei fast gleicher Zeile die Spalte entscheiden zu lassen — eine
billige Näherung für eine Zeilengruppierung. `leseordnung` **hat** die
Gruppierung, samt der Regel „ein Feld gehört zur Zeile, wenn es vor deren Mitte
beginnt", und der Kommentar dort führt aus, warum der erste, naive Anlauf
scheiterte.

**Offen bleibt die andere Hälfte: die Höhe.** Von 19 Textfeldern stehen **4**
über der Falz; das letzte liegt bei 2 146 px. Das ist keine Frage der Ordnung
mehr, sondern eine der Gestaltung — und sie gehört zu Phase 4.10
(#1325), wo
sie an allen Seiten zugleich beantwortet wird.

**Und die übrigen Kernwege sind noch nicht vermessen**: Suchen/Ersetzen,
Massenänderung, Bild setzen, Set ausgeben, Sammlungseintrag, Preis holen.

Sie stehen im Issue und fehlen hier — das ist eine Lücke und keine Aussage
über sie.
