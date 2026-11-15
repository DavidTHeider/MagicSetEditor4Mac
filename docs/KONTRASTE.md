# Kontraste

**Nachgemessen, nicht geschätzt.** Unter jsdom meldet axe `color-contrast` als
*unentschieden*: ohne Layout gibt es keine gerechnete Farbe. Diese Datei ist
deshalb kein Text über die Prüfung, sondern ihr Ergebnis — und
`internal/werkzeuge/kontraste_test.go` hält sie gegen den Quelltext: eine
Vordergrundfarbe, die in `frontend/src` benutzt wird und hier fehlt, macht die
Pipeline rot.

---

## Und seit dem 2026-08-30 misst es ein Lauf

Die Tabellen unten sind **Paare** — „`text-slate-600` auf `bg-white`: 7,58" —
und ein Wächter hält sie gegen den Quelltext. Was weder Tabelle noch Wächter
sagt: **ob dieses Paar auf der Seite auch wirklich übereinander liegt.** Eine
Klasse steht im Quelltext; was ein Auge sieht, entscheidet der Baum darüber.

```bash
node scripts/kontrastlauf.mjs
```

Er geht sechs Seiten in **beiden** Erscheinungsbildern durch, nimmt jede
Textstelle mit eigenem Text, holt ihre Farbe und den Grund vom nächsten
Vorfahren, der einen hat, und rechnet den Kontrast — über einen Canvas, aus
demselben Grund wie unten.

**Erhoben am 2026-09-03**: **4.768** Textstellen, **null** unter der Schranke
— nach der Umstellung auf die Gestaltungssprache
([0065](entscheidungen/0065-wie-diese-anwendung-aussieht.md)) und die
Anzeigeschrift an den Überschriften. Die Werte der Merkmale `--mse-*` sind
**dieselben**, die die Oberfläche vorher über `slate-*` benutzte; was sich
geändert hat, ist der Ort, an dem sie stehen. Vorher, am 2026-08-30: **4.526**
Textstellen, ebenfalls null.
Die Gegenprobe — Überschriften auf `#a8b0bb` — macht 28 davon rot, und zwar
**nur im hellen** Erscheinungsbild; im dunklen ist dieselbe Farbe hell genug.
Das ist der Beleg, dass beide Seiten wirklich getrennt gemessen werden.

**Was er nicht ersetzt:** die Zeilen unten. Er misst, was dasteht — nicht, ob
die Farbe die richtige ist; und **Ränder ohne Fläche** (WCAG 1.4.11) hat er so
wenig wie axe.

**Die sechs Linien der Kartenvorschau (U5, #1416) sind aus der Palette
gerechnet, nicht vom Bildschirm gelesen**: `theme.css` von Tailwind 4 nennt
jede Farbe als `oklch(…)`, und die OKLab-Matrizen führen sie in dasselbe sRGB,
das ein Canvas zurückgibt — `#e7000b`, `#0084d1`, `#bb4d00`, `#ffd230`,
`#ff6467`, `#74d4ff`. Gemessen gegen den Seitengrund, weil das Kartenbild
darunter jede Farbe haben kann; dafür sind die Linien 2 px breit und
gestrichelt oder gepunktet, damit sie sich auch von einem gleichfarbigen Bild
abheben.

**Seit AP-2 (2026-10-16) tragen zwei der drei Linien ein Merkmal mit anderem
Ton**: der Befund `fehler` (`#8c2f16` hell, `#f0a07a` dunkel — 6,82 und
8,90 gegen den Seitengrund) statt Rot 600/400, die Bildgrenzen `auswahlkante`
(`#0069a8` hell — 4,82; dunkel `#74d4ff` wie bisher, 11,21) statt Himmel 600.
Die Beschnittlinie bleibt Bernstein 700 und 300 (`markierungskante`, 4,14 und
12,90). Die Zahlen gegen den Seitengrund Pergament; gemessen am Fenster mit
`farbsprache09`, die Tafel steht unten im Abschnitt „Ein Fehlersystem“.

## Wie gemessen wurde

Im laufenden Vorschauserver (`npm --prefix frontend run vorschau`), im Browser,
nicht im Test:

1. Tailwind 4 gibt seine Palette als CSS-Variablen aus. Der Wert einer Farbe
   kommt aus `getComputedStyle(document.documentElement)` — also aus dem
   Stylesheet, das die Anwendung wirklich lädt.
2. Die Farbe wird in einem 1×1-Zeichenbereich ausgefüllt und als RGB
   zurückgelesen. **Das ist der Punkt an der Sache:** Tailwind 4 schreibt
   `oklch()`, und ein Verhältnis aus den Ziffern dieser Angabe zu rechnen, gibt
   Unsinn — ein früherer Anlauf las Schwarz auf Weiß als 2,10 : 1.
3. Verhältnis nach WCAG 2.1, relative Leuchtdichte über die sRGB-Formel.

Ein Alphakanal von 0 gilt als **Fehler**, nicht als Schwarz: eine Klasse, die
Tailwind gar nicht ausgegeben hat, käme sonst als perfekter Kontrast durch.
Auch das ist ein bezahlter Befund.

**Erschlossen am:** 2026-08-08, ergänzt am 2026-08-15, zweimal am 2026-08-19, am 2026-08-20, am 2026-08-25, am 2026-08-26 und am 2026-08-28 · **Tailwind:** 4.3.3

**Die Messung vom 2026-08-28** kam mit den Wahlbildern der Stiloptionen
(#1225) und hat die gewählte Kachel gerechnet — Rand und Grund in beiden
Erscheinungsbildern. **Auch sie hat sich selbst geprüft**: sie hat
`border-sky-700` auf Weiß und `outline-sky-300` auf `slate-900` noch einmal
gerechnet und wieder 5,86 und 10,70 bekommen, die Zahlen, die seit dem
2026-08-08 beziehungsweise dem 2026-08-19 hier stehen.

**Was sie nebenbei mitgemessen hat und was hier nicht als Zeile steht:** die
Schrift **auf** der gewählten Kachel — `slate-900` auf `sky-50` gibt 16,73,
`slate-100` auf `sky-950` gibt 12,68. Sie steht nicht in der Tabelle, weil die
Grundfarbe der Kachel keine eigene Vordergrundfarbe setzt: gezeigt wird die
Schriftfarbe des Körpers, und die steht schon oben.

**Die zweite Messung vom 2026-08-26** kam mit dem Quelltextfeld (#1164) und
hat vier Farben dazugelegt. Sie hat dabei `text-sky-800`/`bg-white` und
`text-red-800`/`bg-white` mitgerechnet — 7,51 und 8,36, wie sie hier stehen.

**Sie kam nur zustande, weil der Wächter erweitert wurde**: er las bis dahin
nur `.tsx`, und die Farbtabelle des Feldes steht in `bausteine/farben.ts`.
Sechs Vordergrundfarben lagen damit außerhalb seiner Reichweite — er meldete
nichts, und das hieß nicht, dass nichts fehlte (§10). Mit `.ts` fand er
sofort zwei ungemessene.

**Die Messung vom 2026-08-26 hat es genauso gehalten**: sie hat die beiden
neuen Bernsteintöne **am gerenderten Element** gerechnet — im laufenden Fenster
unter `wails dev`, an der Warnung über einer aufgeschlagenen Vorlage — und
dabei `text-red-800`/`bg-white`, `text-red-300`/`bg-slate-900` und
`text-sky-800`/`bg-white` mitgemessen: 8,36, 9,29 und 7,51, wie sie hier
stehen. Der Wert am Element (6,47) und der aus der Palette gerechnete (`amber-800`
auf `slate-100`) sind derselbe — das ist die Probe darauf, dass die Tabelle
misst, was das Fenster zeigt.

**Die Messung vom 2026-08-25 hat sich selbst geprüft**: sie hat neben den
beiden neuen Rändern auch `text-sky-800` auf Weiß und `text-sky-300` auf
`slate-900` noch einmal gerechnet und dieselben Zahlen bekommen, die seit dem
2026-08-08 hier stehen — 7,51 und 10,70. Eine Messung, die die vorhandene
Tabelle nicht reproduziert, misst etwas anderes; das ist billig zu prüfen und
war es einmal wert.

---

## Zwei Erscheinungsbilder, drei Einstellungen

**Seit #245
ist das Erscheinungsbild wählbar** — hell, dunkel oder wie das System. Das sind
drei Einstellungen und weiterhin **zwei** Paletten: `system` wählt eine der
beiden aus, es bringt keine dritte mit. Die Tabellen unten gelten damit
unverändert, und der Abnahmekasten „die Kontraste sind in jedem
Erscheinungsbild gemessen" ist mit ihnen eingelöst.

**Was die Wahl ändert, ist wann welche gilt.** Bis #245 entschied das
`prefers-color-scheme` des Rechners; jetzt ein Attribut am Wurzelelement
(`data-erscheinungsbild`), und ohne Attribut wieder das System. Nachgesehen im
Browser am 2026-08-19: bei dunkel eingestelltem System und Wahl „hell" steht
die Anwendung hell, und umgekehrt.

**Die Schriftgröße ändert an den Kontrasten nichts** — sie skaliert die
Wurzelschrift in vier Stufen (87,5 % bis 125 %) und lässt jede Farbe stehen.
Für die Schwelle „großer Text ab 24 px" arbeitet sie zugunsten der Lesbarkeit:
was bei 100 % darunter liegt, liegt bei 125 % näher daran.

---

## Ein Platzhalter ist Text

**Tailwinds Vorgabe reicht nicht.** `::placeholder` bekommt dort
`currentColor` zu 50 %; gemessen am 2026-08-19 ergibt das

| Vordergrund | Grund | Verhältnis |
| --- | --- | --- |
| `slate-900` zu 50 % | `bg-slate-100` | **3,34** |
| `slate-900` zu 50 % | `bg-white` | **3,41** |
| `slate-100` zu 50 % | `bg-slate-950` | 4,97 |
| `slate-100` zu 50 % | `bg-slate-900` | 4,88 |

Die beiden hellen liegen **unter** der Schwelle 4,5 aus WCAG 1.4.3, und ein
Platzhalter ist Fließtext: seit
#563 steht in
ihm der **gerechnete Wert** eines Feldes, also eine Auskunft und kein Hinweis
darauf, was man tippen könnte.

**Gefunden hat es die Sichtprüfung.** axe meldet `color-contrast` unter jsdom
als *unentschieden*, und kein DOM-Test misst eine gemischte Farbe.

Wo ein Platzhalter steht, trägt er deshalb `placeholder:text-slate-600` und
`dark:placeholder:text-slate-400` — die Nebentextfarben dieses Projekts, beide
oben geführt.

---

## Schwellen

| Was | Nach | Mindestens |
| --- | --- | --- |
| Fließtext | WCAG 1.4.3 (AA) | 4,5 : 1 |
| Großer Text ab 24 px | WCAG 1.4.3 (AA) | 3 : 1 |
| Ränder, Fokusringe, Bedienelemente | WCAG 1.4.11 (AA) | 3 : 1 |

---

## Text

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `text-slate-900` | `bg-white` | 17,83 | Fließtext hell |
| `text-slate-100` | `bg-slate-900` | 16,28 | Fließtext dunkel |
| `text-slate-600` | `bg-white` | 7,58 | Nebentext hell |
| `text-slate-400` | `bg-slate-900` | 6,78 | Nebentext dunkel |
| `text-slate-900` | `bg-slate-100` | 16,28 | gewählter Navigationseintrag hell |
| `text-slate-100` | `bg-slate-800` | 13,34 | gewählter Navigationseintrag dunkel |
| `text-slate-600` | `bg-slate-100` | 6,92 | Nebentext auf gewähltem Eintrag hell |
| `text-slate-400` | `bg-slate-800` | 5,56 | Nebentext auf gewähltem Eintrag dunkel |
| `text-red-800` | `bg-white` | 8,36 | Fehlermeldung hell |
| `text-red-300` | `bg-slate-900` | 9,29 | Fehlermeldung dunkel |
| `text-red-200` | `bg-slate-900` | 12,27 | Meldungsband, Fehler dunkel (#180) |
| `text-emerald-800` | `bg-white` | 7,61 | Meldungsband, Erfolg hell (#180) |
| `text-emerald-200` | `bg-slate-900` | 13,94 | Meldungsband, Erfolg dunkel (#180) |
| `text-sky-800` | `bg-white` | 7,51 | Schaltfläche als Verweis, hell |
| `text-sky-300` | `bg-slate-900` | 10,70 | Schaltfläche als Verweis, dunkel |
| `text-slate-100` | `bg-slate-950` | 18,40 | Fließtext auf dem Anwendungsgrund, dunkel (#288) |
| `text-slate-400` | `bg-slate-950` | 7,66 | Nebentext auf dem Anwendungsgrund, dunkel (#288) |
| `text-white` | `bg-sky-800` | 7,51 | Hauptknopf hell (#288) |
| `text-slate-950` | `bg-sky-300` | 12,10 | Hauptknopf dunkel (#288) |
| `text-slate-600` | `bg-slate-100` | 6,92 | Platzhalter einer Herleitung, hell (#563) |
| `text-slate-400` | `bg-slate-950` | 7,66 | Platzhalter einer Herleitung, dunkel (#563) |
| `text-slate-600` | `bg-slate-200` | 6,15 | Code in einem Hilfesatz, hell (#699) |
| `text-slate-400` | `bg-slate-800` | 5,56 | Code in einem Hilfesatz, dunkel (#699) |
| `text-red-800` | `bg-slate-100` | 7,63 | Vorbehalt an der Paketzeile, hell (#244) |
| `text-red-300` | `bg-slate-950` | 10,50 | Vorbehalt an der Paketzeile, dunkel (#244) |
| `text-fehler` | `bg-grund` | 6,82 | Was schiefging, hell (#1643; bis AP-2 `text-warnung`) |
| `text-fehler` | `bg-flaeche` | 7,89 | Was schiefging auf einer Fläche, hell (#1643; bis AP-2 `text-warnung`) |
| `text-fehler` | `bg-betont` | 6,10 | Was schiefging auf einer betonten Zeile, hell (#1643; bis AP-2 `text-warnung`) |
| `text-fehler` | `bg-grund` | 8,90 | Was schiefging, dunkel (#1643; bis AP-2 `text-warnung`) |
| `text-fehler` | `bg-flaeche` | 7,89 | Was schiefging auf einer Fläche, dunkel (#1643; bis AP-2 `text-warnung`) |
| `text-fehler` | `bg-betont` | 6,95 | Was schiefging auf einer betonten Zeile, dunkel (#1643; bis AP-2 `text-warnung`) |
| `text-amber-800` | `bg-slate-100` | 6,47 | Warnung über der Vorlage, hell (#1164) — **bis N13**: seit Phase 6.10 (#1671) steht dort `text-warnung` auf `bg-betont`, am Fenster nachgemessen mit `n13` (6,10 hell, 6,95 dunkel — dieselben Werte wie die Zeilen von #1643). Die Zeile bleibt, weil das Paar an anderen Stellen gilt |
| `text-amber-300` | `bg-slate-950` | 13,93 | Warnung über der Vorlage, dunkel (#1164) — bis N13, wie darüber |
| `text-violet-800` | `bg-slate-100` | 8,37 | Zahl im Quelltextfeld, hell (#1164) |
| `text-violet-300` | `bg-slate-950` | 10,87 | Zahl im Quelltextfeld, dunkel (#1164) |
| `text-emerald-800` | `bg-slate-100` | 6,95 | Zeichenkette im Quelltextfeld, hell (#1164) |
| `text-emerald-200` | `bg-slate-950` | 15,76 | Zeichenkette im Quelltextfeld, dunkel (#1164) |

**Der Anwendungsgrund ist seit #288
nicht mehr `bg-white`/`bg-slate-900`.** Die Seitenleiste und jede Fläche stehen
als Blatt darauf; der Grund selbst ist `bg-slate-100` hell und `bg-slate-950`
dunkel. Für hell waren beide Textfarben schon gemessen (16,28 und 6,92 — die
Zeilen zum gewählten Navigationseintrag), für dunkel stehen sie jetzt hier.

**`text-white` trifft der Wächter sehr wohl** — seine Regel nennt `white` und
`black` seit dem Nachlauf zu #576 ausdrücklich. Hier stand das Gegenteil, und
der Satz war älter als die Regel; angestrichen hat es der Zweitblick über
Phase O.

**`text-slate-600` auf `bg-slate-900` wäre 2,35** — deshalb steht neben jedem
`text-slate-600` ein `dark:text-slate-400`, ohne Ausnahme. Die Zeile steht hier,
damit sichtbar ist, was das Paar verhindert.

## Ränder und Fokus (WCAG 1.4.11)

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `border-slate-500` | `bg-white` | 4,76 | Eingabefelder, Auswahlfelder, Trennlinie der Navigation |
| `border-slate-500` | `bg-slate-900` | 3,74 | dieselben, dunkel |
| `divide-slate-500` | `bg-white` | 4,76 | Trennlinien zwischen Listeneinträgen |
| `divide-slate-500` | `bg-slate-900` | 3,74 | dieselben, dunkel |
| `border-red-700` | `bg-white` | 6,42 | Rand des Meldungsbands, Fehler hell (#180) |
| `border-red-400` | `bg-slate-900` | 6,17 | derselbe, dunkel |
| `border-emerald-700` | `bg-white` | 5,36 | Rand des Meldungsbands, Erfolg hell (#180) |
| `border-emerald-400` | `bg-slate-900` | 9,20 | derselbe, dunkel |
| `border-red-600` | `bg-white` | 4,76 | Diagnoselinie über der Kartenvorschau, hell (U5, #1416) — **über dem Bild**, dessen Farben niemand kennt; gemessen gegen den Seitengrund, deshalb 2 px breit |
| `border-red-400` | `bg-slate-900` | 6,17 | dieselbe, dunkel |
| `border-sky-600` | `bg-white` | 4,02 | Bildgrenzen über der Kartenvorschau, hell (U5) — gestrichelt, 2 px |
| `border-sky-300` | `bg-slate-900` | 10,71 | dieselben, dunkel |
| `border-amber-700` | `bg-white` | 5,05 | Beschnittlinie über der Kartenvorschau, hell (U5) — gepunktet |
| `border-amber-300` | `bg-slate-900` | 12,34 | dieselbe, dunkel |
| `border-sky-700` | `bg-white` | 5,86 | Rand des Fassungsbands, hell (#1100) |
| `border-sky-400` | `bg-slate-900` | 8,18 | derselbe, dunkel (#1100) |

**`border-red-700` auf `bg-slate-900` wäre 2,78** — unter der Schwelle von 3.
Deshalb steht am Meldungsband `dark:border-red-400`, und die Zeile hier sagt,
was das Paar verhindert. Dasselbe für `border-emerald-700`, das dunkel bei 3,32
gerade eben trüge und mit `emerald-400` bei 9,20 steht.
| `border-red-700` | `bg-white` | 6,42 | Feld mit Befund, hell |
| `border-red-400` | `bg-slate-900` | 6,17 | Feld mit Befund, dunkel |
| `outline-sky-700` | `bg-white` | 5,86 | Fokusring hell |
| `outline-sky-300` | `bg-slate-900` | 10,70 | Fokusring dunkel |
| `border-slate-500` | `bg-slate-100` | 4,35 | Rand einer Fläche auf dem Anwendungsgrund, hell (#288) |
| `border-slate-500` | `bg-slate-950` | 4,23 | derselbe, dunkel (#288) |
| `border-sky-700` | `bg-sky-50` | 5,49 | gewählte Wahlkachel einer Stiloption, hell (#1225) |
| `border-sky-300` | `bg-sky-950` | 8,34 | dieselbe, dunkel (#1225) |

**`border-slate-400` lag hell bei 2,63 und `border-slate-300` bei 1,49.** Beides
war eingebaut — die Trennlinien der Set- und Kartenliste und der Rand der
Schaltfläche zur Befehlspalette. Beide sind auf `slate-500` gehoben, das als
einziger Ton in **beiden** Ansichten über 3 : 1 liegt.

**Und dieser Satz stimmte ein halbes Jahr lang nicht ganz.**
`border-slate-400` stand am 2026-08-22 noch an **vier** Stellen — dem Rand der
Symbolschaltfläche, dem Rahmen um die Palette, dem um die Zuschnittfläche und
dem um die Ausfuhrangaben. Alle vier trugen `dark:border-slate-500`: dunkel in
Ordnung, hell bei 2,63 : 1.

**Gesehen hat es niemand, weil der Wächter eine Erwähnung genügen ließ.** Er
suchte die Farbe irgendwo in dieser Datei — und fand sie in genau dem Absatz,
der ihre Abschaffung behauptete. Seit
#816 verlangt
er eine **Tabellenzeile mit einem Wert**, und der Wert muss über 3 liegen. Das
kostet keine Ausnahme: von den 24 Farben mit einer Zeile lag keine darunter.

## Flächen ohne Textfunktion

| Fläche | Nachbar | Verhältnis | Warum das reicht |
| --- | --- | --- | --- |
| `bg-sky-800` | `bg-slate-300` | 5,05 | Balkenfüllung gegen ihre Bahn, hell |
| `bg-sky-300` | `bg-slate-700` | 6,22 | dieselbe, dunkel |
| `bg-slate-300` | `bg-white` | 1,49 | die **Bahn** des Balkens |
| `bg-slate-700` | `bg-slate-900` | 1,36 | dieselbe, dunkel |
| `bg-marke` | `bg-betont` | 6,24 | Füllung gegen Bahn im Zugkasten, hell (Artboard *Tisch*) |
| `bg-marke` | `bg-betont` | 6,48 | dieselbe, dunkel |
| `bg-betont` | `bg-flaeche` | 1,29 | die **Bahn** im Zugkasten, hell |
| `bg-betont` | `bg-flaeche` | 1,13 | dieselbe, dunkel |

Die Bahn eines Balkens trägt keine Auskunft: die Zahl und der Anteil stehen als
Text daneben, und die Fläche ist `aria-hidden`. Was gelesen werden muss, ist das
Verhältnis von Füllung zu Bahn — und das sind 5,05 beziehungsweise 6,22.

**Der Balken im Zugkasten** (Artboard *Tisch*, seit dem 2026-09-24) folgt
derselben Regel: sein Name ist der Abschnitt, der als Text darüber steht, und
die Füllung hält gegen die Bahn 6,24 hell und 6,48 dunkel. Gerechnet aus den
Farben, die WebKit am Element meldet (`play03-quer.probe`); es sind die
Tokenwerte ohne Deckkraft.

---

## Ein Rahmen auf einem fremden Bild

**Der Zuschnittrahmen liegt auf einem Bild, das jede Farbe haben kann**
(#576). Ein
Verhältnis gegen „den Grund" gibt es hier nicht — es gibt keinen.

**MSE löst das mit zwei Linien**, und dieselbe Antwort steht hier:

	dc.SetPen(*wxBLACK_PEN);
	dc.DrawRectangle(left - 1, top - 1, width + 2, height + 2);
	dc.SetPen(Color(64,64,64));
	dc.DrawRectangle(left - 2, top - 2, width + 4, height + 4);

`src/gui/image_slice_window.cpp`, Zeilen 504–507, Fassung f9d9356 vom
2021-07-06.

**Die Zusage ist deshalb eine andere Zahl**: nicht „diese Farbe gegen jenen
Grund", sondern **die bessere der beiden Linien, im schlechtesten aller
Gründe**. Gemessen wie oben, über jede mögliche Leuchtdichte von 0 bis 1 in
Schritten von 1/10.000:

| Paarung | Schlechtester Grund | Was dort bleibt |
| --- | --- | --- |
| `outline-white` gegen `ring-slate-900` | Leuchtdichte 0,199 | **4,22** |
| `border-white` gegen `bg-sky-900` (Griff) | — | **3,08** |

Beide liegen über der Schwelle 3 : 1 aus WCAG 1.4.11, und zwar **für jedes
denkbare Bild**. Zum Vergleich stehen die beiden Linien zueinander bei
**17,83**, der Griffrand zu seiner Füllung bei **9,48**.

**Der Versatz ist der Punkt an der Sache, und er kam aus der Sichtprüfung.**
Ohne `outline-offset-2` liegen beide Linien deckungsgleich — die weiße über der
dunklen —, und auf einem hellen Bild ist der Rahmen **unsichtbar**. Kein Test
hat das gefunden; nachgemessen im laufenden Fenster stand `outline-offset: 0px`
gegen einen `box-shadow` mit 2 px Ausdehnung. MSE zeichnet die beiden
Rechtecke bei −1 und −2, also konzentrisch, und genau das tut der Versatz.

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `outline-white` | `ring-slate-900` | 17,83 | die beiden Rahmenlinien zueinander (#576) |
| `ring-slate-900` | `outline-white` | 17,83 | dieselbe Paarung, andersherum gelesen (#576) |
| `border-white` | `bg-sky-900` | 9,48 | Rand eines Griffs gegen seine Füllung (#576) |
| `stroke-amber-700` | weiß (Zeichenfläche) | 5,03 | Mitte und Griff einer Symmetrie, hell (#1092) |
| `stroke-amber-300` | `bg-slate-950` (Zeichenfläche) | 13,93 | dieselben, dunkel (#1092) |

---

## Die Gestaltungssprache auf Pergament — 2026-09-03

**Seit Werkstück J von Phase 4.10**
(#1325) gibt
es neben Tailwinds Paletten eine eigene: neun Merkmale mit sprechenden Namen,
je Erscheinungsbild ein Wert, festgelegt in
[0065](entscheidungen/0065-wie-diese-anwendung-aussieht.md). Die Tabellen
darüber bleiben — die `slate-*`-Farben stehen weiter im Quelltext, wo eine
Fläche eine bestimmte Bedeutung trägt.

**Und über Grund und Fläche liegt eine Textur.** Damit hat der Hintergrund
keine *eine* Farbe mehr, und jede Zahl hier ist die des **ungünstigsten**
Bildpunkts: `node scripts/pergamentmass.mjs` rastert die Kachel aus
`frontend/src/pergament-hell.svg` beziehungsweise `-dunkel.svg`, legt sie über
die Grundfarbe und rechnet den Kontrast über alle 810.000 Punkte.

**`bg-betont` und `bg-marke` tragen die Faser nicht** — die Regel in
`stil.css` hängt sie an `.bg-grund` und `.bg-flaeche`, sonst nirgends. Für
diese beiden ist die Zahl deshalb die des reinen Farbpaars.

**Was der Verlust durch die Textur kostet, steht im Lauf**: zwischen 0,45 und
1,58 Punkten. Der knappste Wert ist `border-kante` mit 3,99 gegen eine Schranke
von 3 — die Textur ist so kräftig eingestellt, wie dieses Budget es zulässt,
und die Klemmung in den beiden SVG ist genau dafür da.

**Und weil der knappste Wert 3,99 ist, gibt es keine verdünnte Kante.**
`border-kante/60` ist nicht `border-kante`: bei 3,99 im hellen Erscheinungsbild
bleiben davon **2,08** übrig — gemessen am 2026-09-11 an der Fensterliste bei
der Sichtprüfung zu
#1500. Der
Wächter über dieser Datei sah es nicht: sein Muster endet an der Farbe, und
`border-kante/60` trifft `border-kante`. Seit demselben Tag steht neben ihm
einer, der jede verdünnte Vordergrundfarbe abweist
(`internal/werkzeuge/verduennung_test.go`) — eine Zahl hinter dem Schrägstrich
macht die Zeile in dieser Tabelle zu einer Aussage über eine andere Farbe.

| Vordergrund | Grund | Verhältnis | Erscheinungsbild |
| --- | --- | --- | --- |
| `text-text` | `bg-grund` | 11,62 | hell |
| `text-leise` | `bg-grund` | 5,85 | hell |
| `border-kante` | `bg-grund` | 3,99 | hell |
| `divide-kante` | `bg-grund` | 3,99 | hell |
| `text-marke` | `bg-grund` | 5,87 | hell |
| `border-marke` | `bg-grund` | 5,87 | hell |
| `outline-ring` | `bg-grund` | 5,87 | hell |
| `text-text` | `bg-flaeche` | 13,20 | hell |
| `text-leise` | `bg-flaeche` | 6,65 | hell |
| `text-emerald-800` | `bg-flaeche` | 6,26 | Dienstkatalog, Status »bereit«, hell (#1440) — gemessen am 2026-09-09 am gerenderten Element über den Canvas |
| `text-emerald-200` | `bg-flaeche` | 14,60 | dieselbe, dunkel |
| `text-amber-800` | `bg-flaeche` | 5,83 | Dienstkatalog, »nie gegen den echten Dienst gelaufen« und »nicht mit dieser Fassung«, hell (#1440) |
| `text-amber-300` | `bg-flaeche` | 12,90 | dieselbe, dunkel |
| `border-kante` | `bg-flaeche` | 4,54 | hell |
| `divide-kante` | `bg-flaeche` | 4,54 | hell |
| `text-marke` | `bg-flaeche` | 6,67 | hell |
| `border-marke` | `bg-flaeche` | 6,67 | hell |
| `outline-ring` | `bg-flaeche` | 6,67 | hell |
| `text-text` | `bg-betont` | 12,35 | hell |
| `text-leise` | `bg-betont` | 6,22 | hell |
| `border-kante` | `bg-betont` | 4,24 | hell |
| `outline-ring` | `bg-betont` | 6,24 | hell |
| `text-marketext` | `bg-marke` | 8,07 | hell |
| `text-text` | `bg-grund` | 12,92 | dunkel |
| `text-leise` | `bg-grund` | 6,76 | dunkel |
| `border-kante` | `bg-grund` | 4,62 | dunkel |
| `divide-kante` | `bg-grund` | 4,62 | dunkel |
| `text-marke` | `bg-grund` | 7,12 | dunkel |
| `border-marke` | `bg-grund` | 7,12 | dunkel |
| `outline-ring` | `bg-grund` | 7,12 | dunkel |
| `text-text` | `bg-flaeche` | 11,15 | dunkel |
| `text-leise` | `bg-flaeche` | 5,84 | dunkel |
| `border-kante` | `bg-flaeche` | 3,99 | dunkel |
| `divide-kante` | `bg-flaeche` | 3,99 | dunkel |
| `text-marke` | `bg-flaeche` | 6,14 | dunkel |
| `border-marke` | `bg-flaeche` | 6,14 | dunkel |
| `outline-ring` | `bg-flaeche` | 6,14 | dunkel |
| `text-text` | `bg-betont` | 11,77 | dunkel |
| `text-leise` | `bg-betont` | 6,16 | dunkel |
| `border-kante` | `bg-betont` | 4,21 | dunkel |
| `outline-ring` | `bg-betont` | 6,48 | dunkel |
| `text-marketext` | `bg-marke` | 8,06 | dunkel |

**Der Fokusring auf dem Hauptknopf fehlt hier, und das ist Absicht.**
`outline-ring` auf `bg-marke` wäre 1,00 — beide sind dieselbe Farbe. Der Ring
liegt dort nie: der Hauptknopf ist die einzige Fläche in Markenfarbe, und er
trägt `outline-offset-2`; sein Ring liegt also auf dem Grund dahinter. Die
Stellen **ohne** Offset sind Eingabefelder auf `bg-flaeche` — nachgezählt am
2026-09-03.

**Die Anzeigeschrift ändert keine dieser Zahlen.** Sie wechselt die Schriftart
der Überschriften, nicht ihre Farbe. Was sie ändert, ist die *Strichstärke* —
`Trattatello` setzt dünner als `Hoefler Text` —, und WCAG kennt dafür keine
Schwelle. Wer eine sehr dünne Schrift ergänzt, sieht sie sich an.

## Der Platzhalter in einem Eingabefeld — 2026-09-15

**Werkstück N8 von Phase 5.4**
(#1569).

### Was vorher galt

**Keine Datei nannte diese Farbe.** Ein `::placeholder` ohne eigene Regel
bekommt vom Browser eine **halbdurchsichtige Ableitung** der Textfarbe —
gemessen am 2026-09-14 in der laufenden Anwendung, helles Erscheinungsbild:

| Feld | Untergrund | Kontrast |
| --- | --- | ---: |
| Einstellungssuche | `rgb(253, 249, 240)` | **3,25 : 1** |
| Adresse eines Kontakts | `rgb(242, 232, 213)` | **3,12 : 1** |

**WCAG 2.2 AA, 1.4.3 verlangt 4,5 : 1.** Ein Platzhalter ist Text, den jemand
liest, um zu wissen, was in das Feld gehört.

### Was jetzt gilt

`frontend/src/stil.css` setzt `::placeholder` auf **`--color-leise`** — also
genau die Farbe, die `text-leise` trägt, und deren Zeilen oben stehen: **5,85
bis 6,76 : 1** auf jedem der drei Gründe, in beiden Erscheinungsbildern. Der
schlechteste Wert liegt damit **1,3 Punkte über** der Schwelle.

**Und `opacity: 1` steht daneben**, weil Firefox einem Platzhalter von sich
aus eine Deckkraft unter eins gibt. Ohne diese Zeile wäre die gesetzte Farbe
nicht die wirksame — und die Messung ginge an der Wirklichkeit vorbei.

### Nachgemessen am laufenden Programm

**Am 2026-09-15**, über einen Canvas an der Einstellungssuche
(`wails dev`, beide Erscheinungsbilder):

| | Platzhalter | Untergrund | Kontrast |
| --- | --- | --- | ---: |
| hell | `rgb(92, 74, 52)` | `rgb(253, 249, 240)` | **8,04 : 1** |
| dunkel | `rgb(185, 166, 136)` | `rgb(35, 30, 23)` | **6,99 : 1** |

**Die Zahl für hell ist höher als die 6,65 der Tabelle oben**, und das ist
kein Widerspruch: jene misst über dem **gefaserten** Grund, diese über der
reinen Farbe. Beide liegen über der Schwelle, und die strengere ist die, die
gilt.

**Die Deckkraft ist dabei mitgemessen worden** (`1`) — sonst sagte die Farbe
nichts über das, was zu sehen ist.

### Was ein Wächter davon nicht sieht

**`internal/werkzeuge/kontraste_test.go` liest Vordergrundfarben aus dem
Quelltext der Bauteile** — `text-leise` und ihresgleichen in den `className`.
Eine Regel in `stil.css` steht dort nicht, und die **alte** Farbe stand
nirgends: sie war eine Browservorgabe. Genau deshalb ist der Befund erst in
der Sichtprüfung aufgefallen und nicht in der Pipeline.

**Was daran hängt**: eine künftige Regel in `stil.css`, die eine Farbe setzt,
fällt wieder durch. Wer eine setzt, misst sie und trägt sie hier ein.

### Was ausdrücklich kein Befund ist

**Die beiden Knöpfe bei 1,16 : 1.** Sie sind `disabled`, und WCAG 1.4.3 nimmt
inaktive Bedienelemente ausdrücklich aus. Gemessen wurden sie mit, und sie
stehen hier, damit niemand sie für übersehen hält.

## Hoher Kontrast — 2026-09-25

**UX-10.AK-03** (#1682)
verlangt, dass hoher Kontrast dokumentiert ist. **Bis hierher war er nur die
fehlende Faser**: `[data-kontrast='hoch']` nahm die Textur des Grundes weg und
ließ jede Farbe stehen.

### Was vorher galt

Jede Farbe hält 4,5 : 1 (AA). **Vier Paare lagen knapp unter 7 : 1** (AAA,
WCAG 1.4.6) — gemessen im WKWebView mit dem Drehbuch `ux10-darstellung`,
bis auf `warnung`, die auf den gemessenen Seiten nicht zu sehen war und
**gerechnet** ist:

| Paar | hell | dunkel |
| --- | --- | --- |
| `leise` auf `grund` | 6,96 | über 7 |
| `leise` auf `flaeche` | 8,04 | 6,99 |
| `marke` auf `grund` | 6,98 | 8,30 |
| `warnung` auf `grund` | 6,84 (gerechnet) | 8,9 (gerechnet) |

### Was jetzt gilt

| Merkmal | hell | dunkel |
| --- | --- | --- |
| `leise`, `kante` | Tinte — `var(--mse-text)` | Tinte — `var(--mse-text)` |
| `marke`, `ring` | `#4e2f16` | unverändert `#d9a441` |
| `warnung` | `#6e2410` | unverändert `#f0a07a` |

**Über `var(--mse-text)` und nicht über einen zweiten Wert**: die Tinte steht
je Erscheinungsbild schon da. Die Regeln stehen in
`frontend/src/barrierefreiheit.css` unter „Hoher Kontrast hebt die Farben".

### Nachgemessen am laufenden Programm

Mit dem Drehbuch `ux10-darstellung`: die Einstellung „Kontrast: Immer hoch"
über *Nutzer und Profile* gesetzt, dann **jedes sichtbare Textelement gegen
seinen wirksamen Grund** — der erste Vorfahr mit deckendem Hintergrund,
Halbdurchsichtiges verrechnet. Das kleinste Verhältnis je Seite:

| Seite | hell | dunkel |
| --- | --- | --- |
| Nutzer und Profile | 12,35 | — |
| Start | 11,47 | — |
| Decks | 9,92 | 8,06 |
| Sammlung | 9,92 | 8,30 |

**Keines unter 7 : 1** — der Lauf prüft es und wird sonst rot.

### Was diese Messung nicht sieht

- **Die Systemeinstellung „Kontrast erhöhen"** (`prefers-contrast: more`). Sie
  bekommt dieselben Werte über dieselben Regeln, aber die Probe kann die
  Medienabfrage nicht setzen.
- **Inaktive Bedienelemente** — WCAG 1.4.3 nimmt sie aus, und die Messung
  zählt sie mit; hätte eines unter 7 gelegen, stünde es hier.
- **Seiten, die der Lauf nicht öffnet.** Die vier stehen für die Bausteine,
  die überall wiederkehren; eine Seite mit einer eigenen Farbe wäre eine
  neue Zeile.

## Die Farbwelten — 2026-09-25

**UX-09.AK-01** (#1681):
neben Pergament drei weitere Farbwelten, je hell und dunkel, in
`frontend/src/themen.css`. **Gerechnet und nicht geschätzt**:
`frontend/src/themen.test.ts` liest die Datei und prüft jedes Paar aus
`kontrast.ts` (`lesbarePaare` — dieselben Paare wie in der Tafel zu Pergament)
gegen seine Schwelle. **Keine Faser**: die Zahlen sind die der reinen
Farbpaare.

| Farbwelt | Bild | größtes Textpaar | kleinstes Textpaar | kleinstes Randpaar | Text unter dem Schleier, schlechtester Bildpunkt | Meldungen unter dem Schleier |
| --- | --- | --- | --- | --- | --- | --- |
| Klar | hell | 17,72 (`text-text` auf `bg-flaeche`) | 6,09 (`text-leise` auf `bg-betont`) | 3,81 (`border-kante` auf `bg-betont`) | 4,89 | 5,17 |
| Klar | dunkel | 19,06 (`text-text` auf `bg-grund`) | 5,81 (`text-leise` auf `bg-betont`) | 4,41 (`border-kante` auf `bg-betont`) | 6,20 | 10,99 |
| Schiefer | hell | 15,56 (`text-text` auf `bg-flaeche`) | 5,86 (`text-leise` auf `bg-betont`) | 3,78 (`border-kante` auf `bg-betont`) | 5,23 | 4,75 |
| Schiefer | dunkel | 15,67 (`text-text` auf `bg-grund`) | 6,45 (`text-leise` auf `bg-betont`) | 4,17 (`border-kante` auf `bg-betont`) | 6,50 | 9,85 |
| Moos | hell | 15,46 (`text-text` auf `bg-flaeche`) | 5,78 (`text-leise` auf `bg-betont`) | 4,04 (`border-kante` auf `bg-betont`) | 5,08 | 4,78 |
| Moos | dunkel | 15,58 (`text-text` auf `bg-grund`) | 6,36 (`text-leise` auf `bg-betont`) | 4,09 (`border-kante` auf `bg-betont`) | 6,53 | 9,86 |

**Die letzten zwei Spalten sind das Fensterbild.** Über jedem Bildpunkt liegt
der Grund zu 90 % (`--mse-schleier`); der schlechteste Punkt ist Schwarz oder
Weiß, und gerechnet ist `text`, `leise`, `marke` und `warnung` gegen beide —
gemischt wie `color-mix(in srgb, …)` — und **die Farben der Meldungen**
(Bernstein und Smaragd des Erscheinungsbilds, `hinweis` und `erfolg`; das
Rot der Fehler ist seit AP-2 `warnung` selbst und steht in der Spalte davor).
**Bis AP-2 stand in der letzten Spalte dunkel 8,28, 7,42 und 7,43** — das
schwächste Paar war Tailwinds `red-300`, das es seither nicht mehr gibt;
nachgerechnet am 2026-10-16 mit derselben Rechnung, die die alten Werte
wieder ergibt. **Zwei Farben
wurden dafür nachgezogen**: `warnung` in Klar hell (vorher 4,50 über
Schwarz) und `kante` in Moos hell (vorher 2,97 über Schwarz).

**Pergament**, die Vorgabe, rechnet `themen.test.ts` seit dem Nachlauf über
Phase 6.3 mit: Text 5,47 (hell) und 6,05 (dunkel), Meldungen 4,68 und 9,90
(bis AP-2 7,46, aus demselben Grund wie oben).
**Bei 88 % stand die Warnung auf Pergament hell bei 4,45 : 1** — deshalb
90 % (N4); die Meldungen maß vorher niemand unter dem Schleier.

**Nachgemessen am gerenderten Element** in WKWebView (`ux09-gestaltung`):
jeder sichtbare Text gegen seinen wirksamen Grund, je Farbwelt hell und
dunkel; kein Wert unter 4,5 : 1.

**Eigene Farben stehen hier nicht** — sie stehen unter keiner Tafel, weil sie
jemand anderes wählt. Für sie gilt die **Sperre**
([0134](entscheidungen/0134-eigene-gestaltung-sind-farben.md)): dieselben Paare,
gemessen am Wurzelelement, bevor gezeichnet wird; was darunter liegt, gilt
nicht.

**Hoher Kontrast sticht jede Farbwelt**: dann gilt die Tinte aus dem
Abschnitt darüber, und ihre Zahlen stehen dort.

## Ein Fehlersystem und die Merkmale der Meldungen — 2026-10-16

**AP-2** (#2359),
Befunde B-07 und B-09. Bis dahin trugen Fehlermeldungen **zwei** Farben:
`text-warnung` aus der Gestaltungssprache, je Farbwelt gesetzt und von der
Sperre gemessen, und über 200-mal Tailwinds `text-red-800 dark:text-red-300`,
an dem weder eine Farbwelt noch eine eigene Farbe etwas änderte. **Seither
heißt die Farbe `fehler` und kommt aus `--mse-warnung`** — die Zeilen oben
unter „Die Gestaltungssprache auf Pergament“ gelten für sie unverändert, und
die Farbwelten rechnet `themen.test.ts`.

**Die Tafeln „Text“ und „Ränder und Fokus“ oben nennen Tailwinds Palette**
(`text-red-800`, `border-sky-600`, …). **Seit AP-2 steht keine dieser
Klassen mehr in einem Bauteil** — `frontend/src/farbsprache.test.ts` weist sie
ab. Die Zeilen bleiben als Geschichte der Messung; was heute gilt, steht in
diesem Abschnitt und in „Die Gestaltungssprache auf Pergament“.

**Die anderen Merkmale tragen die Werte, die vorher als Palettenklasse
dastanden** — in `stil.css` als Hexzahl, damit die Rechnung hier dieselbe ist
wie im Fenster. Sie folgen dem Erscheinungsbild und nicht der Farbwelt; die
Sperre misst `hinweis` und `erfolg` gegen eigene Flächen mit (`kontrast.ts`,
`statusfarben`).

**Gerechnet aus den Hexwerten und nachgemessen am gerenderten Element** in
WKWebView (`farbsprache09`, Beleg unter `belege/huellenprobe/`): jede Klasse
an einem Probeelement auf Grund, Fläche und betonter Fläche, hell und dunkel.

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `text-hinweis` | `bg-grund` | 5,83 | Warnung, hell |
| `text-hinweis` | `bg-flaeche` | 6,75 | Warnung auf einer Fläche, hell |
| `text-hinweis` | `bg-betont` | 5,21 | Warnung auf einer betonten Zeile, hell |
| `text-hinweis` | `bg-grund` | 12,90 | Warnung, dunkel |
| `text-hinweis` | `bg-betont` | 10,07 | Warnung auf einer betonten Zeile, dunkel |
| `text-erfolg` | `bg-grund` | 6,26 | Erfolg, hell |
| `text-erfolg` | `bg-betont` | 5,60 | Erfolg auf einer betonten Zeile, hell |
| `text-erfolg` | `bg-grund` | 14,60 | Erfolg, dunkel |
| `text-erfolg` | `bg-betont` | 11,40 | Erfolg auf einer betonten Zeile, dunkel |
| `text-zeichenkette` | `bg-flaeche` | 7,24 | Zeichenkette im Quelltextfeld, hell |
| `text-zeichenkette` | `bg-flaeche` | 12,93 | Zeichenkette im Quelltextfeld, dunkel |
| `text-zahl` | `bg-flaeche` | 8,73 | Zahl im Quelltextfeld, hell |
| `text-zahl` | `bg-flaeche` | 8,92 | Zahl im Quelltextfeld, dunkel |
| `border-fehler` | `bg-betont` | 6,10 | Rand eines ungültigen Felds und der Löschfrage, hell — die Farbe von `text-fehler` |
| `border-fehler` | `bg-betont` | 6,95 | dasselbe, dunkel |
| `border-erfolg` | `bg-betont` | 5,60 | Rand des Meldungsbands, Erfolg, hell |
| `border-erfolg` | `bg-betont` | 11,40 | dasselbe, dunkel |
| `border-auswahlkante` | `bg-betont` | 4,31 | Bildgrenzen in der Kartenvorschau, hell — gestrichelt, 2 px |
| `border-auswahlkante` | `bg-betont` | 8,75 | dasselbe, dunkel |
| `border-markierungskante` | `bg-betont` | 3,70 | Beschnittlinie in der Kartenvorschau, hell — gepunktet |
| `border-markierungskante` | `bg-betont` | 10,07 | dasselbe, dunkel |
| `outline-bildhell` | `ring-bilddunkel` | 17,83 | Zuschnittrahmen über einem Bild — gleich in jedem Erscheinungsbild, Weiß über Schiefer wie bisher |
| `ring-bilddunkel` | `outline-bildhell` | 17,83 | dasselbe, von innen |
| `border-bildhell` | `bg-bilddunkel` | 17,83 | Griff des Zuschnitts |

**Die Ränder stehen gegen die betonte Fläche**, weil sie dort am knappsten
sind; gegen Grund und Fläche liegen sie höher. **Die beiden Linien der
Kartenvorschau liegen über einem Kartenbild**, das jede Farbe haben kann —
gemessen ist wie bei U5 gegen den Seitengrund, und sie heben sich durch
Strichart und Breite ab, nicht durch ihre Farbe allein.

**Was hier nicht steht**: `fill-` und `stroke-` des Symboleditors
(`auswahl`, `markierung`, `leise`, `betont/60`) — Füllungen ohne Text, gegen
den Grund der Zeichenfläche gemessen in `farbsprache09`, aber der Wächter
über diese Datei verlangt sie nicht, weil er nur Vordergrund liest.

## Gewählt, Lasche und Auswahlpfeil — 2026-10-16

**AP-3** (#2360),
Befunde B-08 und B-10, und Festlegung 10 aus
[0065](entscheidungen/0065-wie-diese-anwendung-aussieht.md): **jede neue
Zustandsfarbe** bekommt hier ihre Zeile, in beiden Erscheinungsbildern. Drei
sind neu — der gedrückte Umschalter trägt jetzt einen Rand in der Marke statt
der Füllung, die gewählte Lasche eine Unterlinie, und das Auswahlfeld einen
Pfeil in `leise`, gezeichnet im Stilblatt.

**Gemessen am gerenderten Element** in WKWebView (`ap3gestalt`, Pergament,
Beleg unter `belege/huellenprobe/`). Ein Rand zählt gegen den Grund
**außerhalb**, denn er trennt das Element von seiner Umgebung; der Pfeil ist
ein Verlauf und wird über eine Probe in seiner Farbe auf dem Grund des Felds
gemessen.

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `border-marke` | `bg-flaeche` | 6,98 | Rand eines gedrückten Umschalters (»Kacheln«) gegen seine Umgebung, hell |
| `border-marke` | `bg-flaeche` | 8,30 | dasselbe, dunkel |
| `border-marke` | `bg-grund` | 6,98 | Unterlinie der gewählten Lasche, hell |
| `border-marke` | `bg-grund` | 8,30 | dasselbe, dunkel |
| `text-text` | `bg-grund` | 13,81 | Schrift der gewählten Lasche, hell |
| `text-text` | `bg-grund` | 15,07 | dasselbe, dunkel |
| `text-leise` | `bg-grund` | 6,96 | Schrift einer ungewählten Lasche, hell |
| `text-leise` | `bg-grund` | 7,89 | dasselbe, dunkel |
| `--color-leise` (Pfeil) | `bg-flaeche` | 8,04 | Pfeil des nativen Auswahlfelds und der Auswahl mit Suche, hell — ein Bedienelement ohne Text, 3 : 1 |
| `--color-leise` (Pfeil) | `bg-flaeche` | 6,99 | dasselbe, dunkel |

**Der Fokusring** liegt 2 px außerhalb und ist 2 px breit, der Rand des
gewählten Segments 1 px: in Pergament dunkel tragen beide die Farbe der
Marke, und sie unterscheiden sich an Lage und Breite, nicht an der Farbe —
gemessen in demselben Lauf.

## Hover, Druck und Ablageziel — 2026-10-16

**AP-4** (#2361),
Befunde B-11 und B-24, Festlegung 10 aus
[0065](entscheidungen/0065-wie-diese-anwendung-aussieht.md): der Grund eines
Knopfs unter dem Zeiger und die Kante eines Ablageziels sind neue
Zustandsfarben. **Gemessen am gerenderten Element** in WKWebView, mit der
echten Maus über dem Knopf (`ap4zustaende`, Pergament).

| Vordergrund | Grund | Verhältnis | Wo |
| --- | --- | --- | --- |
| `text-text` | `bg-betont` | 12,35 | Nebenknopf unter dem Zeiger (»Set öffnen …«), hell |
| `text-text` | `bg-betont` | 11,77 | dasselbe, dunkel |
| `text-marketext` | `bg-marke/90` | 6,43 | Hauptknopf unter dem Zeiger (»Weiter«), hell — die Marke zu neun Zehnteln über dem Grund |
| `text-marketext` | `bg-marke/90` | 6,72 | dasselbe, dunkel |
| `text-marke` | `bg-betont` | 6,24 | leiser Knopf unter dem Zeiger (»Überspringen«), hell |
| `text-marke` | `bg-betont` | 6,48 | dasselbe, dunkel |
| `outline-marke` | `bg-grund` | 6,98 | gestrichelte Kante eines Bildfelds, über dem eine Datei steht, hell — sie liegt über dem Kartenbild, das jede Farbe haben kann; gemessen wie bei U5 gegen den Seitengrund |
| `outline-marke` | `bg-grund` | 8,30 | dasselbe, dunkel |

**Gedrückt** ändert keine Farbe, sondern die Lage (ein Punkt tiefer); ein
gesperrter Knopf antwortet weder dem Zeiger noch dem Druck.

## Was hier nicht steht

- **Zustände, die nur ein Zeigegerät auslöst.** `hover:bg-slate-100` ist nicht
  gemessen; ein Zustand, der ohne Maus nicht eintritt, trägt keine Auskunft, die
  nicht auch ohne ihn dasteht.
- **Bilder und Kartenausgabe.** Was `internal/render` zeichnet, kommt aus fremden
  Vorlagen; sein Kontrast ist die Sache der Vorlage, nicht unsere.
- **Farben, die in einem Stylesheet stehen statt in einer `className`.** Der
  Wächter liest die Bauteile; eine Regel in `stil.css` oder
  `barrierefreiheit.css` sieht er nicht. Heute sind es drei — `::placeholder`,
  die Werte des hohen Kontrasts und seit AP-3 der Auswahlpfeil in
  `erscheinungsmerkmale.css`, alle oben gemessen —, und wer eine vierte
  setzt, misst sie und trägt sie ein.
- **Ob eine Farbe an der richtigen Stelle steht.** Der Wächter prüft, dass jede
  benutzte Vordergrundfarbe hier eine Zeile hat — nicht, dass die Zeile den
  richtigen Grund nennt. Ein `text-slate-600` auf dunklem Grund wäre hier
  gedeckt und trotzdem falsch. Dagegen hilft nur die Sichtprüfung, und die
  bleibt eine Handarbeit.

## Die Farben des Paketformats — 2026-08-28

**Keine neue Zahl, und das ist die Aussage.**
#1181 hat dem
Vorlagenbereich eine eigene Einfärbung gegeben; ihre beiden neuen Arten teilen
sich die Farben mit den vorhandenen:

| Art | Klasse | schon gemessen als |
| --- | --- | --- |
| `Schlüssel` | `text-sky-800` / `dark:text-sky-300` | `Schlüsselwort` der Sprache |
| `Wert` | `text-slate-900` / `dark:text-slate-100` | `Bezeichner` der Sprache |

**Warum dieselben und nicht eigene**: ein Schlüssel des Paketformats hat
dieselbe Rolle wie ein Schlüsselwort der Sprache — er sagt, worum es in der
Zeile geht; ein Wert ist der Inhalt, wie ein Bezeichner. Zwei weitere Farben
für dieselben zwei Rollen machten die Anzeige bunter und nicht klarer.

**Und `Kommentar` gilt für beide.** In beiden Sprachen ist es `#` bis zum
Zeilenende; zwei Farben dafür wären zwei Antworten auf dieselbe Frage.

## Der Spieltisch — 2026-09-16

**Zehn Paare, und alle tragen dieselbe Linie.** Werkstück G3 von Phase 5.5
(#1460)
zeichnet einen Tisch auf eine Leinwand; was dort liegt, ist kein DOM und wird
vom Lauf `kontrastlauf.mjs` nicht gesehen. Gemessen wurde deshalb von Hand,
und zwar **jede Farbe gegen jeden Untergrund, auf dem sie wirklich liegt**
(§10: die Reichweite einer Messung ist die Reichweite ihrer Aussage).

**Zwei Holztöne nebeneinander tragen keine 3:1** — das war der erste Anlauf,
und er scheiterte an jeder Kombination, die noch nach Holz aussah. Die Lösung
steht in WCAG 1.4.11 und in §3 dieser Leitlinien: ein **Rand ohne Fläche**
zählt. Deshalb umrandet `--mse-tisch-kante` jede Sitzfläche, jedes Zonenfeld
und jede Karte, und nur sie muss die Schranke halten. Die Füllungen sind dann
frei — sie tragen keine Bedeutung, die ein Kontrast sichern müsste.

| Erscheinungsbild | Kante auf … | Werte | Kontrast |
| --- | --- | --- | --- |
| hell | Tischplatte | `#786248` / `#d4bb94` | 3,11 : 1 |
| hell | fremde Sitzfläche | `#786248` / `#dcc4a0` | 3,42 : 1 |
| hell | eigene Sitzfläche | `#786248` / `#e4cfae` | 3,80 : 1 |
| hell | verdeckte Karte | `#786248` / `#fdf9f0` | 5,49 : 1 |
| hell | wartende Karte | `#786248` / `#e0d0b0` | 3,80 : 1 |
| dunkel | Tischplatte | `#9b8768` / `#221b13` | 4,91 : 1 |
| dunkel | fremde Sitzfläche | `#9b8768` / `#2a2218` | 4,52 : 1 |
| dunkel | eigene Sitzfläche | `#9b8768` / `#332a1e` | 4,06 : 1 |
| dunkel | verdeckte Karte | `#9b8768` / `#231e17` | 4,77 : 1 |
| dunkel | wartende Karte | `#9b8768` / `#3d3226` | 3,60 : 1 |

**Und der Auswahlrahmen** — er ist die einzige Farbe am Tisch, die etwas
*bedeutet* statt etwas darzustellen, und deshalb misst er gegen jeden Grund,
auf dem er liegen kann:

| Erscheinungsbild | Auswahl auf … | Werte | Kontrast |
| --- | --- | --- | ---: |
| hell | Tischplatte | `#2d1206` / `#d4bb94` | 9,44 : 1 |
| hell | fremde Sitzfläche | `#2d1206` / `#dcc4a0` | 10,37 : 1 |
| hell | eigene Sitzfläche | `#2d1206` / `#e4cfae` | 11,52 : 1 |
| hell | Karte | `#2d1206` / `#fdf9f0` | 16,64 : 1 |
| dunkel | Tischplatte | `#fbf0dc` / `#221b13` | 15,07 : 1 |
| dunkel | fremde Sitzfläche | `#fbf0dc` / `#2a2218` | 13,88 : 1 |
| dunkel | eigene Sitzfläche | `#fbf0dc` / `#332a1e` | 12,48 : 1 |
| dunkel | Karte | `#fbf0dc` / `#231e17` | 14,65 : 1 |

> **Hier standen bis zum 2026-09-20 `#8a3d1e` und `#e8a05a`** — die Farben des
> **ersten** Anlaufs, mit ihren Kontrasten. Der Absatz unter dieser Tabelle
> nannte schon damals die richtigen Zahlen (3,03 und 3,07 gegen die Kante),
> und das sind die Werte der **neuen** Farben; die alten ergeben dort 1,32 und
> 1,59 — also genau die Zahl, die derselbe Absatz als den gescheiterten
> Anlauf benennt.
>
> Die Farbe wurde gewechselt, die Schlussfolgerung nachgetragen, die Tabelle
> blieb stehen. **Gesehen hat es niemand, weil es für die sieben
> `--mse-tisch-*` keinen Wächter gab**: `kontraste_test.go` hält
> Tailwind-Klassennamen, und Merkmale sind keine Klassen — er sagt das in
> seinem Kopf auch. Seit dem 2026-09-20 hält
> [`internal/werkzeuge/tischfarben_test.go`](../internal/werkzeuge/tischfarben_test.go)
> jeden Wert gegen diese Datei.

**Und die Schrift auf dem Tisch** — Zonenname und Kartenzahl, seit
PLAY-03.AK-06
(#1711).

**Sie ist nicht die Kante**, obwohl sie dieselbe Aufgabe zu haben scheint: die
Kante steht im hellen Bild bei 3,11 : 1 gegen die Tischplatte. Für eine Linie
reicht das — WCAG 1.4.11 verlangt 3:1 —, für **Text** nicht: 1.4.3 verlangt
4,5:1. Ein Zonenname in Kantenfarbe wäre eine Beschriftung, die man erraten
muss.

| Erscheinungsbild | Schrift auf … | Werte | Kontrast |
| --- | --- | --- | ---: |
| hell | Tischplatte | `#2d1206` / `#d4bb94` | 9,44 : 1 |
| hell | fremde Sitzfläche | `#2d1206` / `#dcc4a0` | 10,37 : 1 |
| hell | eigene Sitzfläche | `#2d1206` / `#e4cfae` | 11,52 : 1 |
| dunkel | Tischplatte | `#fbf0dc` / `#221b13` | 15,07 : 1 |
| dunkel | fremde Sitzfläche | `#fbf0dc` / `#2a2218` | 13,88 : 1 |
| dunkel | eigene Sitzfläche | `#fbf0dc` / `#332a1e` | 12,48 : 1 |

**Der schlechteste Wert ist 9,44 : 1**, und das ist kein Zufall: gebraucht wird
das Dunkelste beziehungsweise Hellste dieser Palette, und genau das trägt auch
der Auswahlrahmen. **Ein eigenes Merkmal bleibt es trotzdem** — die beiden
sagen Verschiedenes, und ein gemeinsames zwänge sie, sich gemeinsam zu ändern.

**Was hier nicht gemessen ist**: die Schrift auf einer **Karte**. Sie liegt
dort nie — ein Zonenname steht am Feld, eine Zahl unter ihm, und beide sind
von Karten überdeckt, nicht auf ihnen.

**Ein Rahmen und keine Färbung**: eine getönte Karte hieße, ihr Bild zu
verändern — und wer zwei Karten derselben Vorlage nebeneinander legte, sähe
zwei verschiedene Farben.

**Und die Zahl, auf die es ankommt, ist eine andere**: der Auswahlrahmen muss
sich von der **Kante** abheben, sonst sähe er aus wie ein gewöhnlicher
Kartenrand. Der erste Anlauf stand dort bei **1,32:1** — geprüft worden war
nur gegen die Flächen, und zwar mit 4,5:1 statt der für grafische Objekte
geltenden 3:1. Eine strengere Zahl an der falschen Stelle ersetzt keine
Prüfung an der richtigen (§10). Heute sind es 3,03:1 im hellen und 3,07:1 im
dunklen Erscheinungsbild.

## Die fünf weiteren Tischbilder — gerechnet statt gemessen

**Sie stehen nicht in dieser Datei, und das ist Absicht.** `stil.css` führt
einen Satz, den ein Auge prüfen muss; die fünf weiteren sind Zahlen im
Quelltext (`frontend/src/tisch/tischbilder.ts`), und Zahlen im Quelltext prüft
ein Lauf: `tischbilder.test.ts` rechnet für **jeden** Satz die Kante und den
Auswahlrahmen gegen jede seiner drei Flächen nach — sechs Sätze, zwei
Erscheinungsbilder, und die Gegenprobe (einen Satz absichtlich zu dunkel)
macht ihn rot.

**Sechs Sätze von Hand zu messen veraltet beim siebten.**

**Wie gemessen wurde:** relative Leuchtdichte nach WCAG 2.2, sRGB-Linearisierung
mit der Schwelle 0,04045, Gewichte 0,2126 / 0,7152 / 0,0722, Kontrast als
(L₁ + 0,05) / (L₂ + 0,05). Die Werte sind Hexwerte aus `frontend/src/stil.css`
— **keine `oklch()`**, und deshalb braucht es hier den Canvas nicht, der oben
beschrieben ist.

**Was diese Tabelle nicht sagt:**

- **Ob ein echtes Kartenbild den Kontrast hält.** Sobald eine Textur geladen
  ist, zeigt die Karte das Bild aus dem Paket; was darauf steht, entscheidet
  dessen Urheber. Der Rand bleibt, und er ist der Teil, für den dieses Programm
  einstehen kann.
- **Ob der Wächter es hält.** `internal/werkzeuge/kontraste_test.go` sucht
  Tailwind-Klassennamen; `--mse-tisch-*` sind Merkmale und keine Klassen. Diese
  zehn Zeilen sind deshalb Handarbeit — wer eine der sechs Farben ändert, misst
  neu.
