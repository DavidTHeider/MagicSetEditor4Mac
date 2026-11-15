# Der Fähigkeitsbestand von MSE2

**Wofür diese Datei da ist.** `docs/ROADMAP.md` sagt: *MagicSetEditor4Mac
enthält die Fähigkeiten von MSE2 als echtes Subset und geht darüber hinaus.*
Ein Subset lässt sich nur behaupten, wenn die **Obermenge** dasteht. Hier steht
sie.

**Sie ist erhoben, nicht geschätzt** — am 2026-08-30 gegen `mse2-quelle`
`f9d9356`, aus `data/en.mse-locale/locale`, Abschnitt `menu:`.

## Was diese Liste ist und was nicht

| | |
| --- | --- |
| **Sie zählt** | was in einem **Menü** steht — 93 Einträge, davon 11 bloße Überschriften |
| **Sie zählt nicht** | Kontextmenüs, Bedienelemente in den Panels, Ziehen und Ablegen, Werkzeugleisten |
| **Damit ist sie** | eine **Untergrenze**. „Vollständig" heißt hier: vollständig für die Menüleiste |

**Zwei weitere Zahlen aus demselben Quelltext**, weil sie die Reichweite dieser
zeigen: `src/util/window_id.hpp` führt **110** Befehlskennungen, und
`add_menu_item_tr` hängt davon **64** in ein Menü. Die Liste unten hat
82 Zeilen — die drei Zahlen messen Verschiedenes und widersprechen
einander nicht.

## Der Stand

| Zeichen | Heißt | Zahl |
| --- | --- | ---: |
| ✓ | es gibt eine Entsprechung | **55** |
| — | es gibt keine, und das ist nachgesehen | **27** |
| ? | **nicht geprüft** | **0** |

**Das `?` war die ehrliche Hälfte dieser Datei** — es hieß nicht „fehlt",
sondern „hier hat noch niemand nachgesehen", und am Tag der Erhebung standen
24 von 82 so da.

**Seit dem 2026-08-30 steht dort null.** Alle vierundzwanzig sind im
MSE2-Quelltext nachgesehen worden (Werkstück A von
#1291); elf
hatten eine Entsprechung, dreizehn nicht. Was dabei auffiel:

- **Zweimal fehlt hier das Gegenteil von dem, was die Zeile fragt.** „Save As
  Directory" ist unsere **einzige** Form — ein Set als `.mse-set`-**Archiv** zu
  schreiben können wir nicht.
- **Die Ablage im Textfeld kommt vom Browser** und musste nie gebaut werden;
  die Ablage der **Schlüsselwortliste** gibt es dagegen nicht.
- **Der Symboleditor hängt am Set, nicht an einer Datei.** Vier der fünf
  Dateibefehle des MSE2-Symboleditors haben deshalb keine Entsprechung — und
  brauchen auch keine, solange kein fremdes `.mse-symbol` geöffnet werden soll.
- **Ein Malwerkzeug gibt es nicht**: ein Symbol ist hier ausschließlich eine
  Folge von Formen mit Bézierpunkten.

**Ein `—` ist keine Absage.** Es sagt, dass nachgesehen wurde und nichts da
ist; ob es gebaut werden soll, entscheidet `docs/PARITAET.md` und die Roadmap.

**Ein ✓ sagt nichts über die Flächen.** Ob eine Fähigkeit auch über
Kommandozeile, MCP und REST erreichbar ist, sagt
[`FAEHIGKEITEN.md`](FAEHIGKEITEN.md) — dort fehlen heute 23, 21 und 22.

## Die Befehle

| MSE2 | Beschriftung | Kürzel | | Bei uns |
| --- | --- | --- | --- | --- |
| `new set` | New... | Ctrl+N | ✓ | Ein leeres Set anlegen — Menü *Neu*, ⌘N |
| `open set` | Open... | Ctrl+O | ✓ | Ein Set öffnen — ⌘O |
| `save set` | Save | Ctrl+S | ✓ | Ein Set sichern — ⌘S |
| `save set as` | Save As... | F12 | ✓ | Menü *Sichern unter*, ⇧⌘S — **F12 gibt es auf einer Mac-Tastatur nicht** |
| `save set as directory` | Save As Directory... | — | ✓ | **Und mehr als das: es ist unsere einzige Form.** `Setschreiber.Schreibe` legt ein **Verzeichnis** an (`os.MkdirAll`) und schreibt die Hauptdatei hinein; MSE2 hat beides und macht das Archiv zur Vorgabe (`Set::saveAs(…, true, true)`, `window.cpp` 605–614). **Was uns fehlt, ist also das Gegenteil**: ein Set als `.mse-set`-**Archiv** zu schreiben. `msepaket` kann ein vorhandenes Archiv umschreiben (`insArchiv`), aber keines neu anlegen |
| `export html` | HTML... | — | ✓ | Über eine Exportvorlage ausgeben |
| `export image` | Card Image... | — | ✓ | Kartenbilder schreiben |
| `export images` | All Card Images... | — | ✓ | Kartenbilder schreiben |
| `export apprentice` | Apprentice... | — | — | kein Treffer im Quelltext |
| `export mws` | Magic Workstation... | — | — | kein Treffer im Quelltext |
| `check updates` | Check Updates... | — | ✓ | `internal/aktualisierung` |
| `print preview` | Print Preview... | — | — | **Nachgesehen am 2026-08-30**: MSE2 öffnet ein `wxPreviewFrame` über denselben `CardsPrintout`, der auch druckt (`print_window.cpp` 196–207) — dieselben Seiten, nur am Bildschirm. **Hier gibt es keine**: `Druckbogen.tsx` stellt Papier, Anschnitt und Marken ein und schreibt eine PDF-Datei; angesehen wird sie im Betrachter des Systems. Das ist eine Abweichung mit Grund — der Systembetrachter blättert, zoomt und druckt besser als ein nachgebautes Fenster — aber es ist keine Entsprechung |
| `print` | Print... | Ctrl+P | ✓ | Einen Druckbogen legen — ⌘P |
| `reload data` | Reload Data | Ctrl+F5 | — | **Nachgesehen**: MSE2 wirft **alle Pakete** weg, liest die Einstellungen neu und öffnet das Set erneut — der Weg für jemanden, der eine Vorlage von Hand geändert hat (`window.cpp` 715–732). Hier gibt es `Neu laden` nur in der **Paketwerkstatt** (`seitenvorlagen.ts`), also für eine Datei, nicht für den ganzen Paketstand. **Und der Katalog frischt bei jedem `pakete`-Aufruf auf**, was die halbe Wirkung schon hat |
| `show profiler` | Show Profiler | Ctrl+F6 | — | Entwicklerwerkzeug; kein Treffer |
| `exit` | Exit | Alt+F4 | ✓ | ⌘Q |
| `undo` | Undo | Ctrl+Z | ✓ | Eine Änderung zurücknehmen — ⌘Z |
| `redo` | Redo | Ctrl+Y | ✓ | ⇧⌘Z |
| `cut` | Cut | Ctrl+X | ✓ | **Vom Browser**, nicht von uns: jedes Kartenfeld ist ein `input` oder `textarea`, und ⌘X/⌘C/⌘V arbeiten dort so, wie das System es vorsieht — samt Rücknahme im Feld. MSE2 muss es selbst bauen, weil sein Texteditor eigen ist |
| `cut card` | Cut | Ctrl+X | ✓ | Karten über die Zwischenablage bewegen |
| `cut keyword` | Cut | Ctrl+X | — | **Die Schlüsselwortliste hat keine eigene Ablage.** Gemessen am 2026-08-30: in `Schluesselwoerter.tsx` kommt kein Kopieren, Ausschneiden oder Einfügen vor. Ein Schlüsselwort lässt sich anlegen, ändern und löschen — zwischen zwei Sets bewegen nicht |
| `copy` | Copy | Ctrl+C | ✓ | **Vom Browser** — siehe `cut` |
| `copy card` | Copy | Ctrl+C | ✓ | Karten über die Zwischenablage bewegen |
| `copy keyword` | Copy | Ctrl+C | — | siehe `cut keyword` |
| `paste` | Paste | Ctrl+V | ✓ | **Vom Browser** — siehe `cut` |
| `paste card` | Paste | Ctrl+V | ✓ | Karten über die Zwischenablage bewegen |
| `paste keyword` | Paste | Ctrl+V | — | siehe `cut keyword` |
| `select all` | Select All | Ctrl+A | ✓ | Alle Karten anhaken |
| `find` | Find | Ctrl+F | ✓ | Im Volltext suchen — ⌘F |
| `find next` | Find Next | F3 | ✓ | steht im Menü als *Weitersuchen*. **Was nicht geprüft ist**, bleibt gesagt: ob es an derselben Stelle weiterspringt wie MSE2 — dort geht die Suche über **Felder und Karten hinweg** weiter, hier über die Trefferliste. Das ist eine Abweichung der Gestalt, keine fehlende Fähigkeit |
| `replace` | Replace | Ctrl+H | ✓ | Im Text ersetzen |
| `auto replace` | Auto Replace... | — | ✓ | Beim Tippen ersetzen |
| `preferences` | Preferences... | — | ✓ | Einstellungen — ⌘, |
| `previous card` | Select Previous Card | PgUp | ✓ | ⇞ — seit #1320 |
| `next card` | Select Next Card | PgDn | ✓ | ⇟ — seit #1320 |
| `search cards` | Search Cards | Ctrl+K | ✓ | Im Volltext suchen |
| `add card` | Add Card | Ctrl+Enter | ✓ | Eine Karte anlegen |
| `add cards` | Add Multiple Cards... | — | — | **Nachgelesen am 2026-08-30** und als #1337 eröffnet. `magic.mse-game` bringt zwölf solcher Skripte mit; hier ist weder der Paketschlüssel noch `new_card()` bekannt |
| `remove card` | Delete Selected Card | — | ✓ | Eine Karte löschen |
| `orientation` | Orientation | — | — | **Nachgesehen**: ein Untermenü mit vier Ankreuzeinträgen — 0°, 90°, 180°, 270° je Karte (`cards_panel.cpp` 79–84). Es dreht die **Karte** in der Anzeige, nicht ihr Bild. Hier gibt es das nicht; ein Treffer auf „drehen" oder „Ausrichtung" fehlt in den Sprachkatalogen |
| `card list columns` | Card List Columns... | — | ✓ | Die Gestalt der Kartenliste |
| `previous keyword` | Select Previous Keyword | PgUp | — | kein Treffer |
| `next keyword` | Select Next Keyword | PgDn | — | kein Treffer |
| `search keywords` | Search Keywords | Ctrl+K | — | **Nachgesehen**: es setzt den Fokus in ein Filterfeld über der Schlüsselwortliste (`keywords_panel.cpp` 107 und 135). Hier hat die Schlüsselwortseite kein Filterfeld — bei den zwölf des Spiels fällt es nicht auf, bei einem Spielpaket mit Hunderten schon |
| `add keyword` | Add Keyword | Ctrl+Enter | ✓ | Ein Schlüsselwort anlegen |
| `remove keyword` | Remove Select Keyword | Del | ✓ | Ein Schlüsselwort löschen |
| `bold` | Bold | Ctrl+B | — | kein Treffer |
| `italic` | Italic | Ctrl+I | — | kein Treffer |
| `reminder text` | Reminder Text | Ctrl+R | — | **Nachgesehen**: ein Ankreuzbefehl **im Texteditor**, der nur greift, wenn die Einfügemarke in einem `<kw…>`-Abschnitt steht (`text.cpp` 559 und 805) — er blendet den Erinnerungstext **dieses einen Vorkommens** ein oder aus. Der Begriff steht hier nur als **Erklärung** im Sprachkatalog; ein Schalter je Vorkommen gibt es nicht |
| `insert symbol` | Insert Symbol | — | ✓ | Symbole zur Auswahl stellen |
| `no spelling suggestions` | (no suggestions) | — | — | keine Rechtschreibprüfung |
| `pie` | Pie | 1 | — | keine Diagramme |
| `bar` | Bar | 2 | — | keine Diagramme |
| `stack` | Stacked Bars | 3 | — | keine Diagramme |
| `scatter` | Scatter Plot | 4 | — | keine Diagramme |
| `scatter pie` | Scatter-Pie | 5 | — | keine Diagramme |
| `new window` | New Window | — | — | ein Fenster |
| `cards tab` | Cards | Alt+1 | ✓ | Lasche *Karten* |
| `style tab` | Style | Alt+2 | ✓ | Schublade *Eigenschaften*, Lasche *Stil* |
| `set info tab` | Set Information | Alt+3 | ✓ | Schublade *Eigenschaften*, Lasche *Setangaben* |
| `keywords tab` | Keywords | Alt+4 | ✓ | Schublade *Eigenschaften*, Lasche *Schlüsselwörter* |
| `stats tab` | Statistics | Alt+5 | ✓ | Ein Set auszählen |
| `random pack tab` | Random Packs | Alt+6 | ✓ | Die Packarten eines Spiels zeigen |
| `console tab` | Console | Alt+7 | ✓ | Skriptkonsole |
| `index` | Index... | F1 | ✓ | Hilfe |
| `website` | Website... | — | — | öffnet die MSE-Webseite im Browser (`window.cpp` 106). Hier gibt es keinen solchen Menüeintrag — und es ist die eine Zeile dieser Liste, deren Fehlen niemanden bei der Arbeit aufhält |
| `about` | About Magic Set Editor... | — | ✓ | Über |
| `new symbol` | New... | Ctrl+N | ✓ | **Der Symboleditor ist da** — `Symbol.tsx` samt Fläche, Griffen, Anfassern, Teilliste und Palette. Er hängt am **Set**, nicht an einer eigenen Datei: ein Symbol entsteht, wenn man die Seite aufschlägt, und wird mit `SymbolSetzen` ins Set gelegt. **Das ist die Abweichung** und sie gilt für die vier Dateibefehle darunter |
| `open symbol` | Open... | Ctrl+O | — | **Ein Symbol aus einer `.mse-symbol`-Datei öffnen gibt es nicht.** Gelesen wird das Symbol **des Sets** (`Symbol`); eine fremde Datei lässt sich nicht laden |
| `save symbol` | Save | Ctrl+S | ✓ | als `SymbolSetzen` — es legt die Zeichnung ins Set. **Nicht als eigene Datei** |
| `save symbol as` | Save As... | F12 | — | dieselbe Abweichung: es gibt keinen Ort außerhalb des Sets, an den geschrieben würde |
| `store symbol` | Store | Ctrl+Enter | ✓ | Ein gezeichnetes Symbol ins Set legen |
| `close symbol editor` | Close | Alt+F4 | ✓ | die Symbolseite wird verlassen wie jede andere — es gibt kein eigenes Fenster, das zu schließen wäre |
| `duplicate` | Duplicate | Ctrl+D | ✓ | Teile gruppieren, auflösen und umordnen |
| `group` | Group | Ctrl+G | ✓ | Teile gruppieren, auflösen und umordnen |
| `ungroup` | Ungroup | Ctrl+U | ✓ | Teile gruppieren, auflösen und umordnen |
| `select` | Select | F5 | ✓ | Teile anklicken und ziehen — `Symbolteilliste.tsx`, `Symbolanfasser.tsx`. **Ohne Tastenkürzel**: MSE2 schaltet mit F5–F10 zwischen Werkzeugmodi, hier gibt es keine Modi |
| `rotate` | Rotate | F6 | ✓ | über die Anfasser der Fläche — dieselbe Abweichung wie bei `select` |
| `points` | Points | F7 | ✓ | Einen Punkt einer Form einfügen oder löschen |
| `basic shapes` | Basic Shapes | F8 | ✓ | `Grundform` legt Kreis, Rechteck, Dreieck und Vieleck an — `docs/FAEHIGKEITEN.md` führt sie als eigene Fähigkeit |
| `symmetry` | Symmetry | F9 | ✓ | Eine Symmetrie bedienen |
| `paint` | Paint | F10 | — | **Nachgesehen**: MSE2 hat ein Malwerkzeug, das Bildpunkte in die Symbolmaske setzt. Hier ist ein Symbol **ausschließlich** eine Folge von Formen mit Bézierpunkten; einen Pinsel gibt es nicht |
