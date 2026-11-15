# Die Einstellflächen von MSE2 — und was davon hier existiert

**Fassung f9d9356 vom 2021-07-06**, geklont unter `/Users/david/mse2-quelle`
(§11). Jede Zeile ist **an dieser Fassung nachgelesen**, nicht aus dem
Feldnamen erschlossen (§15).

Anlass ist #475:
die Frage „sind die ganzen Optionen bekannt, die via MSE2 im UI zu
konfigurieren sind?" ließ sich am 2026-08-18 nicht beantworten. Sie ist jetzt
beantwortbar — und die Antwort lautet in Teilen „nein, und hier steht warum".

## Was diese Datei **nicht** ist

**Keine Aufgabenliste.** Was fehlt und gebraucht wird, bekommt ein eigenes
Issue; die Spalte *Hier* verweist darauf. Diese Datei sammelt nicht, sie ordnet
zu.

**Keine Aussage über den Rest von MSE2.** Erfasst sind die Einstellungen aus
`src/data/settings.hpp`, die Seiten des Einstellungsfensters, die
Ausgabefenster und die Reiter des Setfensters. Menüeinträge, Tastaturbefehle
und der Symboleditor stehen nicht darin.

---

## 1 Anzeige und Ausgabe je Stil — `StyleSheetSettings`

`src/data/settings.hpp`, Zeilen 93–105. Die Werte sind `Defaultable<T>`: jeder
Stil erbt den Vorgabewert, bis ihn jemand für diesen Stil überschreibt.

| Feld | Was es tut (nachgelesen) | Hier |
| --- | --- | --- |
| `card_zoom` | Zoomfaktor der Kartenanzeige. Das Einstellungsfenster zeigt ihn in Prozent mit den Stufen 50/66/75/100/120/150/200 (`preferences_window.cpp:210–216`) | **fehlt** — die Vorschau zeichnet bei 1 |
| `card_angle` | Drehung der **ganzen** Karte; `DataViewer::getRotation` (`src/render/card/viewer.cpp:113`) — nicht zu verwechseln mit `angle:` am Feldstil, das gebaut ist (#486) | **fehlt** |
| `card_anti_alias` | Im Fenster „high quality" (`preferences_window.cpp:196`). Steuert `RenderQuality` beim Zeichnen | **fehlt** — hier wird immer geglättet |
| `card_borders` | Im Fenster „show lines": zeichnet die **Feldkästen** als Hilfslinien im Editor | **fehlt** |
| `card_draw_editing` | Im Fenster „show editing hints": Editierhinweise, etwa der Cursor im Textfeld | **fehlt** |
| `card_normal_export` | Im Fenster **negiert** als „zoom export" (`preferences_window.cpp:205`, `SetValue(!…)`): ob die Bildausgabe den Zoom mitnimmt oder in Normalgröße schreibt | **fehlt** — die Bildausgabe nimmt eine Breite entgegen |
| `card_spellcheck_enabled` | Im Fenster „spellcheck enabled" | **teilweise** — `check_spelling` gibt es, aber ohne Wörterbuch (#458) |

**Die ersten fünf sind Editorhilfen und Anzeigeoptionen, keine Eigenschaften
der Karte.** Sie ändern nichts an der Datei und nichts an der Ausgabe — außer
`card_normal_export`. Wer sie baut, baut Bedienkomfort; wer sie wegen der
Parität baut, baut das Falsche.

---

## 2 Einstellungen je Spiel — `GameSettings`

`src/data/settings.hpp`, Zeilen 71–89.

| Feld | Was es tut (nachgelesen) | Hier |
| --- | --- | --- |
| `default_stylesheet` | Welcher Stil bei einem neuen Set vorausgewählt ist | **fehlt** — die Stilwahl zeigt alle |
| `default_export` | Welche Exportvorlage vorausgewählt ist | **fehlt** |
| `cardlist_columns` | Welche Felder die Kartenliste als Spalten zeigt, in welcher Breite und Reihenfolge (`ColumnSettings`, Zeilen 55–65) | **fehlt** — die Kartenliste hat feste Spalten |
| `sort_cards_by`, `sort_cards_ascending` | Sortierung der Kartenliste | **fehlt** |
| `images_export_filename` | Muster für die Dateinamen der Bildausgabe; das Fenster zeigt es als freien Text (`images_export_window.cpp:27`) | **da** — `Bildereinstellungen.tsx` führt `vorlage` |
| `images_export_conflicts` | Was bei einem vorhandenen Dateinamen geschieht: `keep old`, `overwrite`, `number`, `number overwrite` (`images_export_window.cpp:29–32`) | **da** — `domain.EinstellungBildkonflikt` führt dieselben vier Werte, `Bildereinstellungen.tsx` zeigt sie |
| `use_auto_replace`, `auto_replaces` | Textersetzungen beim Tippen (`AutoReplace`, `auto_replace_window.cpp`) | **teilweise** — `AutoErsetzen` wendet die `auto replace:`-Regeln **des Spielpakets** an. Was fehlt, ist die **eigene** Liste und der Schalter dafür |
| `pack_amounts`, `pack_seed_random`, `pack_seed` | Der Boosterzusammensteller (`random_pack_panel.cpp`) | **da** — der Reiter *Zufallspack* führt Packart und Samen; derselbe Same gibt dieselbe Ziehung |

---

## 3 Programmweite Einstellungen — `Settings`

`src/data/settings.hpp`, Zeilen 133–210.

| Feld | Was es tut | Hier |
| --- | --- | --- |
| `locale` | Programmsprache, gewählt aus den vorhandenen `.mse-locale`-Paketen (`preferences_window.cpp:150–163`) | **da, anders** — zwei fest eingebaute Kataloge (§9), keine `.mse-locale` |
| `recent_sets` (max. 9) | Zuletzt geöffnete Sets | **da** seit #1101 — neun wie dort, im Menü **und** auf der Startseite. **Anders im Ort**: sie stehen im Bestand (`sets.geoeffnet`), nicht in den Einstellungen; ein Eintrag, dessen Datei weg ist, sagt das und lässt sich trotzdem aufschlagen |
| `default_set_dir`, `default_image_dir`, `default_symbol_dir`, `default_export_dir` | Zuletzt benutzte Verzeichnisse je Zweck | **teilweise** — die Bildausgabe merkt sich ihren Ordner |
| `set_window_maximized`, `set_window_width`, `set_window_height`, `card_notes_height` | Fenstergeometrie | **fehlt** — macOS stellt Fenster selbst wieder her |
| `open_sets_in_new_window` | Ob ein Set ein neues Fenster bekommt (`preferences_window.cpp:151`) | **fehlt** — ein Fenster |
| `symbol_grid_size`, `symbol_grid`, `symbol_grid_snap` | Raster im Symboleditor | **fehlt** — der Symboleditor steht (Phase X), aber ohne Raster und ohne Einrasten |
| `default_game` | Welches Spiel bei „Neu" vorausgewählt ist | **fehlt** |
| `print_layout` | `LAYOUT_NO_SPACE` oder `LAYOUT_EQUAL_SPACE` (`settings.hpp:115–118`); das Druckfenster bietet es als „Put space between cards" (`print_window.cpp:169`) | **da, anders** — der Druckbogen führt Papier, Anschnitt, Auflösung, Schnittmarken |
| `apprentice_location` | Pfad zu einem Fremdprogramm (`preferences_window.cpp:299`) | **fehlt, und nicht gewollt** — MSE exportiert dorthin; das ist ein Format aus den 1990ern |
| `package_versions_url`, `installer_list_url`, `check_updates`, `check_updates_all`, `website_url` | Aktualisierungsprüfung für Programm und Pakete | **fehlt** — der Bezug holt an einem festen Stand ([0025](entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md)) |
| `install_type` | Ob als Programm oder portabel installiert | **entfällt** — ein `.app`-Bündel |
| `export_options` | Je Exportvorlage die Werte ihrer eigenen Optionsfelder (`exportOptionsFor`, Zeile 185) | **da** — `ExportoptionSetzen`, `SetVorlagenausfuhr.tsx` |

---

## 4 Die Reiter des Setfensters

`src/gui/set/`.

| Reiter | Was er tut | Hier |
| --- | --- | --- |
| `cards_panel` | Kartenliste und Karteneditor | **da** — `Karten.tsx`, `Karteneditor.tsx` |
| `set_info_panel` | Die `set field:`-Felder aus dem `.mse-game` | **teilweise** — `Setangaben.tsx` führt einen Teil. Hier stand bis UXR-04 `SetSichern.tsx` — eine Datei, die nie Setangaben trug und mit UXR-04 im Menü der Setkarte aufging |
| `style_panel` | Die `styling field:`-Blöcke des Stils | **da** — `Stiloptionen.tsx`; 374 der 376 Stilpakete führen solche Felder |
| `keywords_panel` | Schlüsselwörter des Sets bearbeiten | **da** — der Reiter *Schlüsselwörter*, an allen vier Flächen. Was fehlt, sind die Parameterhilfen (`insert parameter`, `refer parameter`) |
| `stats_panel` | Auswertungen über das Set | **da** — `Statistik.tsx` |
| `random_pack_panel` | Booster zusammenstellen | **da** — der Reiter *Zufallspack* |
| `console_panel` | Skriptkonsole samt Fehlerliste | **da** — der Reiter *Konsole* wertet einen Ausdruck im Kontext der Karte aus, zeigt Fehlerstelle und Satz des Zerteilers, führt die verfügbaren Namen, schlägt sie im Feld vor und bietet neun Beispiele zum Anklicken (Phase 3.3). **Der Verlauf kam mit U13** (#1424) und geht über MSE2 hinaus: dort ist er je Sitzung (`HistoryTextCtrl::history`, ↑/↓), hier bleibt er, und ⌘↩ führt aus. Hier stand bis zum 2026-09-10 „was fehlt: der Verlauf" — der Zweitblick über Phase 5.1 hat den Widerspruch zur Portierungskarte angestrichen |

---

## 5 Was **nicht** erfasst ist

- **Der Symboleditor** (`src/gui/symbol/`) — dieses Programm hat keinen.
- **Menü und Tastatur** (`src/gui/set/window.cpp`) — §6 sagt, dass macOS-Eigenheiten bedient und nicht nachgeahmt werden; eine Zeile-für-Zeile-Zuordnung wäre irreführend.
- **Das HTML-Ausgabefenster** über die Vorlagenwahl hinaus: seine Optionen kommen aus der Exportvorlage selbst und stehen deshalb unter `export_options`.

## 6 Warum es keinen Wächter gibt

Die Abnahme von #475
verlangte einen Wächter, der diese Datei gegen `src/data/settings.hpp` hält —
**oder die benannte Begründung, warum es keinen gibt.** Sie lautet:

**Der Quelltext von MSE2 liegt nicht im Repositorium** (§11) und ist auf keiner
anderen Maschine vorhanden; ein Wächter darüber wäre auf genau diesem Rechner
grün und sonst übersprungen. Das ist dieselbe Lage wie bei den Läufen hinter
`fremdpakete` — nur ohne deren Ertrag: dort misst der Lauf etwas, hier zählte
er Zeilen einer Kopfdatei.

**Was stattdessen gilt:** die Fassung steht oben. Wer MSE2 aktualisiert, liest
diese Datei neu — und sieht an der Fassungsangabe, dass sie zu einer anderen
gehört.
