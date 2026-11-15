# Paritätsmatrix

**Was MSE2 kann — und was davon hier steht.**

Diese Datei beantwortet die eine Frage, die
[`PORTIERUNGSKARTE.md`](PORTIERUNGSKARTE.md) ausdrücklich offen lässt:
**was ist unvermessen?** Die Karte sagt, woher eine gebaute Einheit stammt;
sie ist keine Vollständigkeitsliste und sagt es selbst. Damit ließ sich bis
zum 2026-08-24 nicht unterscheiden, ob etwas **bewusst nicht gebaut** oder
**nie angesehen** wurde — und genau diese Unterscheidung verlangt §15, wenn
sie sagt: `new` darf keine unterlassene Suche verdecken.

**Die Fassung gehört dazu.** Alle Angaben beziehen sich auf `mse2-quelle`,
Fassung **f9d9356 vom 2021-07-06**, geklont **neben** dem Repositorium (§11).
Die Zählungen sind mit `ls` und `grep` über diesen Baum erhoben, am
2026-08-24.

---

## Die sechs Marken

| Marke | Heißt |
| --- | --- |
| **ported** | Verhalten aus dem Altcode übernommen |
| **adapted** | für die neue Plattform angepasst — mit Grund |
| **new** | neue Funktionalität ohne Entsprechung im Altcode. **Sagt, wo gesucht wurde** (§15) |
| **uncertain** | das Verhalten des Originals ist noch nicht geklärt |
| **deferred** | bewusst vertagt — mit der Phase, die es aufnimmt |
| **unsupported** | bewusst nicht vorgesehen — mit Grund |

**`deferred` und `unsupported` sind neu gegenüber §15.** Sie sind der Grund,
aus dem es diese Datei gibt: die vier Marken der Portierungskarte beschreiben,
**wie** etwas portiert wurde, und haben keinen Platz für „gar nicht".

---

## Was diese Matrix nicht sagt

- **Ob die Marke stimmt.** Dass etwas `ported` ist und nicht `uncertain`, ist
  eine Aussage über Verhalten. Das prüft ein Test, kein Wächter über eine
  Tabelle.
- **Ob eine Zeile fachlich vollständig ist.** „Kartenliste — ported" heißt
  nicht, dass jede Spalte, jede Sortierung und jedes Kontextmenü da sind. Wo
  eine Einheit nur zum Teil da ist, steht es in der Spalte *Für Nutzende*.
- **Was MSE2 in Menüs und Tastenkürzeln kann** — jenseits der Zeile *Das
  Setfenster selbst* unten. [`MSE2-OBERFLAECHE.md`](MSE2-OBERFLAECHE.md)
  klammert die Menüs ausdrücklich aus; seit dem 2026-08-27 ist die Leiste
  **vermessen** (dieselbe Zeile nennt, was fehlt), die Tastenkürzel dahinter
  sind es nicht.
- **Was unterhalb der gewählten Körnung liegt.** Vermessen ist je **Datei**
  des Altprojekts in den unten genannten Bereichen; eine einzelne Funktion
  darin kann fehlen, ohne dass diese Tabelle es zeigt.

---

## Bereich 1 — Das Setfenster

`src/gui/set/`, **sieben Bereiche und das Fenster**. Gezählt: neun Dateien,
davon `panel.cpp` (die Basisklasse) und `window.cpp` (das Fenster selbst).

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| Kartenbereich — Liste, Editor, Vorschau | `src/gui/set/cards_panel.cpp` | `seiten/Karten.tsx`, `Karteneditor.tsx`, `Kartenvorschau.tsx` | ported | `karten.test.tsx`, `karteneditor.test.tsx` | steht |
| Setangaben | `src/gui/set/set_info_panel.cpp` | `seiten/Setangaben.tsx` | ported | `setangaben.test.tsx`, `setangabengruppen.test.tsx` | steht — **seit UXR-07 in den Gruppen des Spiels** (#1915): die Auskünfte (`type: info`) teilen die Felder, die erste Gruppe steht offen, jede weitere hinter ihrer Überschrift. MSE zeichnet eine Auskunft als Kopfzeile und klappt nichts ein — das Einklappen ist die Gestalt, nicht die Sache |
| Stilbereich | `src/gui/set/style_panel.cpp` | `seiten/Stil.tsx`, `Stiloptionen.tsx` | ported | `stil.test.tsx`, `stiloptionen.test.tsx`, `stiloptionenweitere.test.tsx` | steht — **seit UXR-07** (#1915) wird der Stil aus einer Auswahl mit Suche gewählt, und von den Optionen stehen die ersten vier offen, der Rest in der Reihenfolge des Stils dahinter |
| Auszählung | `src/gui/set/stats_panel.cpp` | `seiten/Statistik.tsx` | ported | `statistik.test.tsx` | steht |
| **Eine Balkengruppe filtert die Kartenliste** | `src/gui/set/stats_panel.cpp` Zeilen 526–547 (`onGraphSelect`, `StatsFilter`, `filterCards`), `src/gui/control/graph.cpp` Zeilen 249–258 (`GraphData::indices`), 1116–1163 (`onMouseDown`, `onChar`) | `seiten/Gruppenbalken.tsx`, `statistikfilter.tsx`, `Statistikfilterhinweis.tsx`, `haken/useGruppenabgleich.ts`, `domain.AuszaehlenKarten`, `service.AuszaehlwerteKarten`, `storage.SetLeser.Verteilung`, `app.WertzahlDTO` | adapted | `gruppenwahl.test.tsx`, `gruppenabgleich.test.tsx`, `kartenfilterstatistik.test.tsx`, `statistikfilter.test.tsx`, `internal/domain/auszaehlkarten_test.go`, `internal/service/auszaehlkarten_test.go`, `internal/flaechen/verteilungskarten_test.go`, Drehbuch `uxrf6` | **Gebaut für UXR-F6** (#2013), ein Kompatibilitätsrest aus UXR-08 (§15). **Nachgelesen am 2026-10-03:** ein Klick wählt **eine** Gruppe je Achse und ersetzt die vorige (`findItem` füllt `current_item` neu; es gibt kein Umschalt-Klicken), nach einer Wahl gibt `GraphData::indices` die Karten der Gruppe zurück — je Karte einmal —, und die Liste zeigt nur sie. **Ohne Wahl ist die Liste leer** (`indices` verlangt einen Vektor so lang wie die Achsen), wechseln die Daten, ist die Wahl weg (`setData` leert `current_item`). Die Pfeiltasten verschieben die Wahl; links über die erste Gruppe hinaus setzt sie auf `-1`, den Platzhalter, und die Liste zeigt dann alle Karten, die im Diagramm stehen. **Adaptiert, jede Abweichung mit Grund:** (1) **die Liste steht auf den Karten und nicht unter dem Diagramm** — die Statistik ist hier eine Schublade (Artboard *Werkzeuge*, „Von hier weiter"), über der Liste steht eine Zeile mit der Wahl, der Zahl und den zwei Wegen (zurück zur Statistik, aufheben); (2) **ohne Wahl zeigt die Liste alle Karten** statt keiner — eine leere Liste ist hier ein Fehler, den niemand erwartet, und der Platzhalter `-1` ist damit der Zustand ohne Filter; (3) **ein Klick auf den gewählten Balken hebt die Wahl auf**, MSE2 kennt außer der Pfeiltaste keinen Weg zurück — ein Klick neben die Balken tut dort nichts —, und die Maus allein käme sonst nie zurück; (4) **Tab, Enter und Leertaste** statt der Pfeiltasten — ein Balken ist ein Knopf, den macOS so bedient; (5) **ein Balken ohne Karten ist angehalten**, MSE2 wählt ihn und zeigt eine leere Liste; (6) **kommt die Statistik nach einer Änderung zurück, folgt die Wahl ihrem Balken**, MSE2 verwirft sie. **Nicht gebaut: Wahl über mehrere Achsen.** MSE2 wählt bei zwei oder drei Dimensionen zugleich eine Zelle (`BarGraph2D::findItem`, `ScatterGraph`); die Statistik hier zählt nach einer — `unsupported`, bis ein Bild mit zwei Dimensionen kommt. Auch **keine Mehrfachwahl mehrerer Gruppen einer Achse**: die gibt es bei MSE2 nicht. **Die Belegung je Feld bleibt zum Lesen** — ihre Balken sind Felder und keine Einteilung, MSE2 kennt sie nicht (`new`; gesucht in `stats_panel.cpp` und `graph.cpp`) |
| Schlüsselwörter | `src/gui/set/keywords_panel.cpp`, `src/gui/control/keyword_list.cpp` | `domain.Keyword`, `service.Schluesselwortpflege`, `Schluesselwoerter.tsx` | adapted | `internal/storage`, `internal/service`, `internal/app`, `internal/flaechen`; Rundlauf byte-genau | **eine eigene Lasche**, an allen vier Flächen: die eigenen anlegen, ändern, löschen, die des Spiels lesen. **Adaptiert ist die Gestalt** — MSE stellt eine Liste neben einen Editor, hier tut es die Seite seit UXR-F3 (#2009) auch: eine Liste der eigenen Wörter (Name, Muster, Verwendung) und **ein** Editor für das gewählte, mit den Wahlregeln von `KeywordList` (beim Öffnen nichts gewählt, *Hinzufügen* wählt das neue, *Entfernen* das Wort an derselben Stelle) und der Tastatur der Kartenliste; bis dahin stand jedes als eigene Karte da. Neben der Liste (ab 48 rem Platz) oder darüber. Was fehlt: die Spalten *Gebrauch* und *Erinnerungstext*, Sortieren, die Suche und Ausschneiden/Kopieren/Einfügen eines Wortes. **MSE2 öffnet das Ergebnis**: am 2026-08-24 gegen `Vanguard.mse-set` gemessen — ein Schlüsselwort angelegt, gefüllt, zurückgeschrieben, und MSEs eigene Kommandozeile gibt für die geschriebene Datei **dieselben 609 Zeilen** wie für das unangetastete Original, ohne eine zusätzliche Warnung. **Und die Rücknahme greift seit dem 2026-08-25** (#1077): anlegen, ändern und löschen schreiben einen Verlaufsschritt, und ein zurückgenommenes Löschen bringt das Wort an **seine** Stelle zurück — MSE hat sie im selben Stapel (`AddKeywordAction`, `src/data/action/keyword.cpp`, Fassung f9d9356 vom 2021-07-06). **Und die Parameterhilfen stehen seit dem 2026-08-31** (#1342): zwei Auswahllisten statt zweier Aufklappmenüs, eingefügt wird **an der Textmarke** wie bei MSE. **Und die des Spiels stehen seit UXR-07 als Text** (#1915): MSE zeigt sie grau in der Liste und eines zur Zeit mit nicht beschreibbaren Feldern; hier stehen alle als benannte Liste, ohne ein Feld. Was fehlt: der `<kw-X>`-Zustand je Stelle |
| Zufallspack | `src/gui/set/random_pack_panel.cpp` | `service.Packbereich`, `Zufallspack.tsx` | adapted | `internal/service`, `internal/app`, `frontend` | **eine eigene Lasche**: Packart wählen, Samen setzen, ziehen. **Adaptiert ist der sichtbare Same** — MSE führt ihn auch, hier steht er neben dem Knopf, weil er die einzige Möglichkeit ist, ein Ergebnis wiederherzustellen. Was fehlt: die gezogenen Karten als Bilder |
| Skriptkonsole | `src/gui/set/console_panel.cpp`, `src/cli/cli_main.cpp` | `service.Skriptkonsole`, `Skriptkonsole.tsx`, `mse skript`, `lesen.skript` | adapted | `internal/service`, `internal/flaechen` | **steht auf allen vier Flächen** — Fenster, Kommandozeile, MCP und Dienst. Das Ergebnis kommt als **Quelltext**, so wie MSEs eigene Konsole es zeigt. Was fehlt: die Beispielkarte, die MSE daneben zeichnet |
| Die Basis aller Bereiche | `src/gui/set/panel.cpp` | `seiten/Seiteninhalt.tsx` | adapted | `navigation.test.ts` | Laschen statt Bereichen — dieselbe Anpassung wie eine Zeile tiefer |
| **Suchen und Ersetzen** | `src/util/find_replace.hpp`, `src/gui/value/text.cpp` (Zeilen 1192–1246), `src/gui/set/cards_panel.cpp` (Zeilen 413–472) | `domain.Ersetzungsfrage`, `service.Ersetzer`, `seiten/Ersetzen.tsx` | adapted | `internal/domain/ersetzung_test.go`, `internal/service/ersetzen_test.go`, `ersetzen.test.tsx`, `internal/flaechen/ersetzlauf_test.go` | steht, und **weiter als MSE2**: dort ist `doReplaceAll` `return false; // TODO`, hier läuft es über ein Set — mit **Vorschau** vorweg und einer Rücknahme, die den ganzen Zug umkehrt. Die Übereinstimmung selbst ist portiert (entmarkter Text, Groß/klein, ganzes Wort, nur Textfelder); **zwei Dinge sind berichtigt** (#1168): MSE2 hält `ab` in `xab` für ein ganzes Wort, und die eigene Fundstelle endete am **Anfang des nächsten** Zeichens statt am Ende des getroffenen — jede Ersetzung unmittelbar vor einer Auszeichnung fraß sie. Der Zweitblick über Phase 3.7 nannte davon nur den Unicode-Fall `İ`; die Klasse ist größer und schon in ASCII sichtbar |
| **Das Setfenster selbst — Laschen, Menü, Titel** | `src/gui/set/window.cpp` (Zeilen 55–110) | `frontend/src/navigation.ts`, `internal/app/menue*.go`, `frontend/src/fenster.ts` | adapted | `navigation.test.ts`, `menueleiste.test.tsx`, `internal/app/menuesetzweig_test.go`, `internal/werkzeuge/menueseiten_test.go` | Laschen statt Bereichen. **Die Leiste steht** — Ablage mit *Ausgeben*, ein Zweig *Set* mit Rückgängig, Wiederherstellen, Suchen, Ersetzen, und *Über* unter Hilfe. **Adaptiert ist der Schnitt**: MSE hängt alles an *Edit*, hier trennt macOS die Textbearbeitung (⌘X/C/V, ⌘, für Einstellungen) von der Setbearbeitung. Was fehlt: *Weitersuchen* (die Suche ist hier eine Seite mit Trefferliste), *Neu laden* (der Bestand ist die Wahrheit, nicht die Datei) und die zuletzt geöffneten Sets (die Startseite führt sie). **Und zwei Lücken, die Umfang waren und keine Anpassung** — angestrichen als Befund 7 des Nachlaufs über Phase 3.7 (#1242): *Alles auswählen* wählt bei MSE2 über `ItemList::doSelectAll` **alle Karten** und nicht nur Text — **seit dem 2026-08-28 gebaut**, als *Alle Karten anhaken* im Zweig *Set*; und der Eintrag *Autoersetzung* **öffnet** dort die Verwaltung — sie hat jetzt ihre **eigene Zeile** in dieser Datei, statt als Nebensatz hier zu stehen |

---

## Bereich 2 — Die Feldarten

`src/data/field/`, **neun Arten**. Gezählt an `src/data/field.cpp:78–86`, wo
MSE2 sie aus dem Namen erzeugt.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| `text` | `src/data/field/text.cpp` | `domain.KindText` | ported | `TestFeldarten` | steht |
| `choice` | `src/data/field/choice.cpp` | `domain.KindChoice` | ported | `stiloptionsarten.test.tsx` | steht |
| `multiple choice` | `src/data/field/multiple_choice.cpp` | `domain.KindMultipleChoice` | ported | `stiloptionsarten.test.tsx` | steht |
| `boolean` | `src/data/field/boolean.cpp` | `domain.KindBoolean` | ported | `stiloptionsarten.test.tsx` | steht |
| `image` | `src/data/field/image.cpp` | `domain.KindImage` | ported | `bildsetzen.test.tsx` | steht |
| `color` | `src/data/field/color.cpp` | `domain.KindColor` | ported | `stiloptionsarten.test.tsx` | steht |
| `info` | `src/data/field/information.cpp` | `domain.KindInfo` | ported | `feldform.test.tsx` | steht |
| `package choice` | `src/data/field/package_choice.cpp` | `domain.KindPackageChoice` | ported | `stiloptionsarten.test.tsx` | steht |
| `symbol` | `src/data/field/symbol.cpp` | `domain.KindSymbol`, `msepaket/spiel.go` | ported | `msepaket/symbolfeld_test.go` | steht seit Phase 4.11 (#1339). **Der Wert ist ein Dateiname im Paket**, keine eingebettete Zeichnung — so steht es im Altcode (`ValueType filename;`, `src/data/field/symbol.hpp`, Fassung f9d9356 vom 2021-07-06), und das Feld trägt dort wörtlich „no extra data" · **Hier stand bis zum 2026-09-02 `deferred`** mit der Begründung, `domain.FieldKind` kenne die Art nicht. Sie kennt sie seit Phase 4.11; die Zeile ist stehen geblieben. Angestrichen von Werkstück D der Phase 4.9 |

---

## Bereich 3 — Die Feldeditoren und ihre Rasterung

`src/gui/value/` und `src/render/value/`, **je neun Dateien** — eine je
Feldart, plus `editor.cpp` und `viewer.cpp` als Basis.

Die Zeilen sind dieselben wie in Bereich 2; hier steht, ob es die
**Bedienstelle** und die **Rasterung** gibt.

| Feldart | C++-Quelle | Hier | Marke | Für Nutzende |
| --- | --- | --- | --- | --- |
| `text` | `src/gui/value/text.cpp`, `src/render/text/viewer.cpp` | `bausteine/Feldeingabe.tsx`, `render.Text` | ported | steht |
| `choice` | `src/gui/value/choice.cpp`, `src/render/value/choice.cpp` | `seiten/Wahlfeld.tsx` | ported | steht |
| `multiple choice` | `src/gui/value/multiple_choice.cpp`, `src/render/value/multiple_choice.cpp` | `seiten/Mehrfachwahl.tsx` | ported | steht |
| `boolean` | `src/gui/value/choice.cpp` (dieselbe Stelle) | `seiten/Wahlfeld.tsx` | ported | steht |
| `image` | `src/gui/value/image.cpp`, `src/render/value/image.cpp` | `seiten/Bildereinstellungen.tsx`, `Zuschnittfenster.tsx` | ported | steht |
| `color` | `src/gui/value/color.cpp`, `src/render/value/color.cpp` | `seiten/Farbwahl.tsx` | ported | steht |
| `info` | `src/gui/value/information.cpp`, `src/render/value/information.cpp` | — (nur Anzeige) | ported | steht |
| `package choice` | `src/gui/value/package_choice.cpp`, `src/render/value/package_choice.cpp` | `seiten/Stilwahl.tsx` | adapted | die Auswahl steht am Stil, nicht am Feld |
| `symbol` | `src/gui/value/symbol.cpp`, `src/render/value/symbol.cpp` | `Feldzeile.tsx` | adapted | `seiten/symbolfeld.test.tsx` | steht seit Phase 4.11. **Die Abweichung**: MSE2 zeichnet den Symbolwert im Feld selbst; hier trägt die Feldzeile einen Knopf, der den **Symboleditor** aufschlägt — dieselbe Fläche, die das Setsymbol bedient. Ein zweiter Editor im Feld wäre eine zweite Antwort auf dieselbe Frage · **Hier stand bis zum 2026-09-02 `deferred`**; siehe Bereich 2 |
| Die Basis aller Feldeditoren | `src/gui/value/editor.cpp`, `src/render/value/viewer.cpp` | `bausteine/Feldzeile.tsx` | ported | steht |

---

## Bereich 4 — Fenster und Dialoge

`src/gui/*.cpp`, **siebzehn Dateien**. `util.cpp` ist Hilfscode und steht
nicht in der Tabelle; `drop_down_list.cpp` und `thumbnail_thread.cpp` sind
Bausteine anderer Fenster und stehen dort, wo sie wirken.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| Über — Fassung, Mitwirkende, Lizenz | `src/gui/about_window.cpp` | `seiten/Abhaengigkeiten.tsx` | adapted | `abhaengigkeiten.test.tsx` | eine Seite statt eines Fensters; sie nennt zusätzlich jede Abhängigkeit |
| Beim Tippen ersetzen | `src/gui/auto_replace_window.cpp` | `AutoErsetzen` | ported | `autoersetzung.test.tsx` | steht; die **Verwaltung** der Regeln steht als eigene Zeile darunter |
| Die Autoersetzung **verwalten** | `src/gui/auto_replace_window.cpp` (Zeilen 158–180, 292–295), `src/gui/set/window.cpp` (797–799) | `domain.Autoersetzungsregel`, `storage.Autoersetzungsablage`, `service.Autoersetzungspflege`, `seiten/Autoersetzungen.tsx` | ported | `autoersetzungsregel_test.go`, `autoersetzungsablage_test.go`, `autoersetzungspflege*_test.go`, `autoersetzungen.test.tsx`, `internal/flaechen/autoersetzung_test.go`, `scripts/nahtvertrag.mjs` (Fall 17) | **Steht seit dem 2026-08-31** (#1341). Anlegen, ändern, löschen, zurücksetzen, Wortgrenze und der globale Schalter — an allen vier Flächen. **Die Regeln des Pakets werden gemischt, nicht kopiert** ([0060](entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)): MSE kopiert sie in `GameSettings::initDefaults` einmalig in seine Einstellungen und notiert selbst `// TODO: when we start saving games, clone here`; eine Kopie hieße hier, dass ein im Paket berichtigtes Muster für immer falsch bliebe. **Eine Regel des Pakets lässt sich abschalten, nicht löschen** — MSEs `remove->Enable(ar && ar->custom)`. **Adaptiert ist allein die Gestalt**: ein Abschnitt auf der Kartenseite statt eines eigenen Fensters (§6). **Der Menüeintrag bleibt offen** — er war der Einstieg in ein Fenster, das es hier nicht gibt |
| Karten für eine Ausgabe wählen | `src/gui/card_select_window.cpp` | `domain.Kartenauswahl`, `bausteine/Ausgabeauswahl.tsx` | adapted | `kartenauswahl_test.go` (domain, app), `bogenauswahl_test.go`, `ausgabeauswahl.test.tsx` | **Steht** seit Werkstück 4 von Phase 3.6 (#1167): Druckbogen, Exportvorlage und Bildausgabe fragen „das ganze Set oder ausgewählte Karten". **Zwei Wahlen statt drei** — MSE2s mittlere (`EXPORT_SEL_SUBSET`) ist eine Liste, die ein Reiter beisteuert, und im Auslieferungsstand tut das genau einer. **Und kein Dialog über einem Dialog** (§6): die Liste steht auf der Seite. MSE2 fragt vor **vier** seiner sieben Ausgabewege — Drucken, Druckvorschau, HTML, mehrere Bilder —, hier sind es die drei, die diesen entsprechen |
| Ausgabe über eine HTML-Vorlage | `src/gui/html_export_window.cpp` | `seiten/SetVorlagenausfuhr.tsx` | ported | `vorlagenausfuhr.test.tsx`, `exportkartenwerte_test.go` | steht. **Die Karten tragen seit #2006 ihre gerechneten Werte** (`export_set` wertet im Kontext des Sets aus, dessen Karten beim Laden gerechnet sind) — `magic-text` und `magic-mws` schreiben für `Archenemy` dieselben Bytes wie MSE2, `magic-text` auch für `Black` (3.788 Karten). **Zwei benannte Abweichungen**: zwei Runden statt eines Abhängigkeitsgraphen (die Ausfuhr rechnet länger — **und eine Kette über zwei Stufen ist nicht gedeckt**: liest ein Feld über `set.cards` ein Feld, das selbst erst mit der Kartenliste gerechnet wird, sieht es in der zweiten Runde dessen Wert aus der ersten; MSE2 rechnet abhängige Skripte nach (`SetScriptManager::updateAllDependend`, `updateRecursive`, `src/script/script_manager.cpp`, Fassung f9d9356 vom 2021-07-06). Über die gemessenen Vorlagen — `magic-text` und `magic-mws` an Archenemy, `magic-text` an Black — trat es nicht auf; für tiefere Ketten ist es ein Kompatibilitätsrest, N5 des Nachlaufs über die QA-Arbeit) und **eine** Abweichung, die behoben ist: **jede Karte rechnet seit #2182 im Stil ihrer Vorlage** — ein Kontext je Stilbogen mit `init script` des Spiels und dann des Stils, wie `SetScriptContext::getContext` (`ausfuhrkontext.go`). Gemessen in VK-01 (#2161, `docs/belege/vk01-kartenstil.md`): an `Planechase` gab MSE2 86 × „plane card", diese Anwendung 86 × „normal" — nach #2182 sind alle 86 Zeilen byte-gleich mit dem Orakel. **Dabei gefunden**: Kommandozeile, REST und MCP setzten keine Stilwahl, `set.stylesheet` war dort `nil`. **Seit #2039 sehen die Vorlagen den Stil**: `set.stylesheet` ist ein Gegenstand (Kurz- und Vollname, Maße), `card.stylesheet` der eigene Stil einer Karte oder `nil` — wie das Orakel antwortet; `card_regions` trägt die Regionen, die der Stil erklärt (`card region:`), **wie in MSE 2.5** ([0165](entscheidungen/0165-card-regions-aus-mse-2-5.md)) — acht Stile des Packs tun es. Gemessen: f9d9356 bricht `magic-lackeyccg` an diesem Glied mit Exit 1 ab und schreibt nichts; hier schneidet die Vorlage bei diesen acht zu und nimmt sonst das Einzelbild |
| Bildausschnitt ziehen | `src/gui/image_slice_window.cpp` | `seiten/Zuschnittfenster.tsx`, `Zuschnittrahmen.tsx` | ported | `zuschnittfenster.test.tsx` | steht |
| Kartenbilder schreiben | `src/gui/images_export_window.cpp` | `seiten/Bilder.tsx` | ported | `bilder.test.tsx` | steht |
| Ein neues Set anlegen | `src/gui/new_window.cpp` | `seiten/SetNeu.tsx` | ported | `setneu.test.tsx`, `setneu_test.go`, `flaechen/setneu_test.go` | **steht seit #1099**, ⌘N. Spiel und Stil wie dort, die Stilliste auf `<spiel>-*` gefiltert, beide vorgewählt und **beide verlangt, an jeder Fläche**: `Set::validate` weist ein Set ohne Stil ab und eines, dessen Stil zu einem anderen Spiel gehört (`src/data/set.cpp:143–149`). Die Fähigkeit gibt es auch ohne Fenster: `mse set-neu`, `POST /sets/neu`, `schreiben.set.neu` — und alle drei verlangen den Stil ebenso, denn die Bedingung steht in `service.Importeur.Neu`. **Nicht portiert ist das Gedächtnis**: MSE merkt sich `default_game` und je Spiel `default_stylesheet`; hier ist das erste Paket vorgewählt — **in MSE2s Ordnung** (`ComparePackagePosHint`, `src/gui/control/package_list.cpp:52–60`: Rang, dann Name) seit dem 2026-09-25, vorher nach dem Namen, und für Magic stand „M15 Leveler D/N" vorn statt „M15 Mainframe". **Ohne Gedächtnis wählt MSE2 gar nichts** (`PackageList::select` setzt `NO_SELECTION`, Zeilen 94–107); hier steht der erste, denn UX-05 S07.NAT-01 verlangt ein Set ohne Kenntnis eines Paketnamens |
| Ein Paket aus einem Gerüst anlegen | **keine** — **gesucht wurde** (§15) am 2026-09-04 in `/Users/david/mse2-quelle` (Fassung f9d9356 vom 2021-07-06) über `src/gui/` nach einem Befehl, der ein Paket anlegt: `src/gui/set/window.cpp` (Dateimenü, Zeilen 58–78) führt `new`, `open`, `save`, `save_as`, `save_as_directory`, `reload_data`, `print`, `exit`; `src/gui/packages_window.cpp` verwaltet **installierte** Pakete und legt keines an; `src/gui/new_window.cpp` legt ein **Set** an und wählt Spiel und Stil aus dem Vorhandenen. Ein neues Paket entsteht dort im Dateimanager | `service.Paketverwalter.Anlegen`, `seiten/Paketanlage.tsx` | new | `paketanlage_test.go`, `paketsaat_intern_test.go`, `flaechen/paketanlage_test.go`, `paketanlage.test.tsx` | **Steht seit dem 2026-09-04** (#1350), an allen vier Flächen. **Neu, weil ein Paket bei MSE2 eine Datei ist und hier eine Fähigkeit** — dieselbe Lage wie beim Löschen eines Sets. Übernommen ist nicht ein Ablauf, sondern **was in der Datei stehen darf**; die Grenze hat das Orakel gezogen und kein Test (§15): `font:` unter einem Bildfeld gibt beim Öffnen „Unexpected key: 'font'" |
| Paketaktualisierungen anzeigen | `src/gui/package_update_list.cpp`, `src/gui/update_checker.cpp` | `service.Paketaktualisierung`, `seiten/Bezugsstaende.tsx` | adapted | `paketaktualisierung_test.go`, `paketaktualisierungtausch_test.go`, `bezugsstaende.test.tsx`, `internal/flaechen/paketbezuege_test.go` | **Seit G04 (#2251, 2026-10-11).** MSE2 holt eine Fassungsliste und vergleicht sie mit den installierten Paketen (`gui/update_checker.cpp`, Zeilen 27–80, Fassung f9d9356 vom 2021-07-06). Hier wird bezogen an einem **Commit**, und gefragt wird die **Quelle selbst** — GitHubs Auskunft über den Hauptzweig, aufgelöst zu einem Commit, **nur auf Knopfdruck**. **Drei Abweichungen mit Grund**: bezogener, geprüfter und neuester Stand stehen getrennt da, und ein Fehler beim Fragen ist ein eigener Zustand; eine Aktualisierung zeigt vorher die berührten Pakete und Sets und verlangt eine Bestätigung, wo etwas wegfallen kann; der bisherige Baum bleibt **aufgehoben** und lässt sich zurücktauschen — MSE2 ersetzt ohne Aufbewahrung (`src/data/installer.cpp`). **Nur GitHub nennt einen neueren Stand**; eine andere Quelle sagt das, statt still „aktuell" zu melden |
| Pakete verwalten | `src/gui/packages_window.cpp` | `seiten/Pakete.tsx`, `Paketliste.tsx`, `Bezugsquellen.tsx` | adapted | `paketverwaltung.test.tsx` | steht, und weiter: Herkunft, Prüfsumme und Abhängigkeiten stehen dabei |
| Einstellungen | `src/gui/preferences_window.cpp` | `seiten/Einstellungen.tsx` | adapted | `einstellungen.test.tsx` | steht; **welche** Einstellungen es gibt, ist nicht Zeile für Zeile verglichen — siehe *Was diese Matrix nicht sagt* |
| Drucken | `src/gui/print_window.cpp` | `seiten/Druckbogen.tsx` | adapted | `druckbogen.test.tsx` | **PDF statt Druckdialog**: der Bogen wird geschrieben, gedruckt wird aus der Vorschau von macOS ([0013](entscheidungen/0013-eigener-pdf-schreiber.md)) |
| Profiler | `src/gui/profiler_window.cpp` | — | unsupported | — | **Keine Laufzeitmessung im Fenster.** Für die Rasterung entsteht sie in Phase Ω als Benchmark, nicht als Fenster |
| Vorschaubilder in der Kartenliste | `src/gui/control/image_card_list.cpp`, `src/gui/thumbnail_thread.cpp` | `internal/service/kartenvorschaubild.go`, `internal/app/kartenbilder.go`, `haken/useListenbilder.ts` | adapted | `kartenvorschaubild_test.go`, `kartenbilder_test.go`, `listenbilder.test.tsx` | **Entschieden in Phase 3.5** (#1166), nach zwei Vertagungen. Nachgelesen ist, **was** MSE2 zeigt: die Datei aus dem **ersten Bildfeld des Spiels**, aus dem Set-Paket geladen und verkleinert — nicht die gezeichnete Karte (`findImageField`, `OnGetItemImage`, `CardThumbnailRequest::generate`, Fassung `5deaeee` vom 2020-05-16). `adapted` an zwei Stellen: die **Maße** (18×14 war ein Symbol von 2007; hier sind es 56×40) und die **Ladeform** — ein Stapel je Sichtfenster statt eines Hintergrundstrangs mit einer Anfrage je Zeile |
| Nach Aktualisierungen sehen | `src/gui/update_checker.cpp` | `internal/aktualisierung` | adapted | `internal/flaechen/aktualisierung_test.go` | **Drei Stufen statt einer**, und die Vorgabe ist **aus**: MSE2 fragt beim Start, hier entscheidet eine Einstellung. Dazu eine Prüfung, die MSE2 nicht hat — Signatur **und** Prüfsumme, bevor eine Datei abgelegt wird ([0051](entscheidungen/0051-die-signatur-wird-geprueft.md)) |
| Willkommensfenster | `src/gui/welcome_window.cpp` | `seiten/Start.tsx` | adapted | `start.test.tsx`, `startkacheln.test.tsx` | eine Startseite statt eines Fensters, mit Kacheln für die nächsten Schritte |
| Kartenliste, Spalten, Filter | `src/gui/control/card_list.cpp`, `card_list_column_select.cpp`, `filter_ctrl.cpp` | `seiten/Kartenliste.tsx`, `Listenspalten.tsx`, `Suche.tsx`, `service.Kartenlistenbau` | adapted | `kartenliste.test.tsx`, `listenspalten.test.tsx`, `suche.test.tsx`, `internal/flaechen/listenlauf_test.go` | steht, samt **Spaltenwahl und Sortierung** — und wie bei MSE2 gehört die Wahl dem **Spiel**, nicht dem Set (#1168). **Die Sortierung war bis zum 2026-08-27 unsichtbar**: der Dienst rechnete sie, die Bindung gab sie zurück, und das Fenster warf sie an der Naht weg — angestrichen als Befund 1 des Zweitblicks über die Phase. **Drei Abweichungen**: die Spalten stehen untereinander statt nebeneinander (die Liste ist hier eine schmale Spalte neben dem Editor), die Breite wird nicht gemerkt (eine Punktzahl von 2007 sagt über einen heutigen Bildschirm nichts), und die Spaltenbreite wird nicht gemerkt. **`sort script:` wird seit #1256 ausgewertet** — die Seltenheit ordnet nach Rang |
| Verteilungsgrafik | `src/gui/control/graph.cpp` | `Verteilung` | ported | `statistik.test.tsx` | steht — **und ihre Balken sind wählbar** (UXR-F6, siehe *Eine Balkengruppe filtert die Kartenliste* in Bereich 1) |
| Symbolliste zur Auswahl | `src/gui/control/gallery_list.cpp` | `Symbolpalette` | ported | `symbolpalette.test.tsx` | steht |
| Schlüsselwortliste | `src/gui/control/keyword_list.cpp` | `service.Parameterverweise`, `Parameterwahl.tsx` | ported | `internal/msepaket`, `internal/service`, `internal/app`, `internal/mcp`, `internal/flaechen`, `scripts/nahtlauf.mjs` | **Gebaut am 2026-08-31** (#1342). Die beiden Parameterhilfen an allen vier Flächen: `insert parameter` schreibt `<atom-param>art</atom-param>` ins **Muster**, `refer parameter` schreibt in den **Erinnerungstext** — `{paramN}`, oder das Ergebnis eines `refer script:` mit `input = "paramN"`. **Adaptiert ist die Gestalt**: zwei Auswahllisten statt zweier Aufklappmenüs an Knöpfen (§6). **Nachgelesen dabei**: ein `refer script:` ist ein `StringScript` (`src/data/keyword.hpp`, Zeile 30) — deshalb steht im Pack `\{{input}\}`, und deshalb hat `msescript` seit demselben Tag auch `\}` als Maskierung |
| Der Symboleditor | `src/gui/symbol/` (13 Dateien) | `msepaket.Symbolzeichnung`, `service.Grundform`, `service.PunktEinfuegen`, `service.GriffZiehen`, `Symbol.tsx` | adapted | `internal/msepaket`, `internal/service`, `internal/app`, `symbol*.test.tsx`; dazu ein Lauf gegen **MSE2 selbst** hinter `fremdpakete` | **Er ist da** (#890): Grundformen, Punkte setzen, ziehen und löschen, Griffe, die ihre drei Sperren **halten**, und alles davon **mit der Tastatur** (§6). `adapted` und nicht `ported`, weil die Zeichenfläche ein SVG ist und kein wxWidgets-Fenster — jeder Punkt ist ein fokussierbares Element, was MSE2 so nicht kennt. **Und die Symmetrie ist bedienbar** seit dem 2026-08-25 (#1092): Art, Kopienzahl, Mitte und Griff, dazu die Kopien blass auf der Fläche und beide Anfasser mit den Pfeiltasten. **Auswahl, Gruppen und Teilreihenfolge stehen** seit dem 2026-08-27 (#1168): ⌘-Klick wählt mehrere, zwei Knöpfe gruppieren und lösen auf, zwei weitere schieben ein Teil durch die Zeichenreihenfolge. **Der Knopf »Gruppe auflösen« war dabei strenger als der Dienst** und schaltete bei einer Gruppe neben einer Form ab — behoben mit Befund 7 des Zweitblicks über die Phase. **Was fehlt**: der Knopf für die **Punktart** am Fenster — MSEs frei, glatt und symmetrisch (`src/gui/symbol/point_editor.cpp`, Zeilen 160–162, Fassung f9d9356 vom 2021-07-06). Die Rechnung steht seit #2436 an Kommandozeile, MCP und Dienst (`lesen.symbol.sperre`, samt dem Mitteln der Griffe); am Fenster wartet sie auf ein Artboard, das sie zeigt (§6). Und die Wahl über zwei Ebenen — die Kinder einer Gruppe stehen eingerückt da, sind aber nicht anklickbar. Das ist auch MSEs Modell (`GroupSymbolPartsAction` läuft über `root.parts`), aber MSE lässt eine Gruppe **öffnen** und darin weiterarbeiten; hier geht das nicht |

---

## Bereich 5 — Das Setmodell

`src/data/*.cpp`, **achtzehn Dateien**.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| Set | `src/data/set.cpp` | `domain.Set`, `storage` | ported | `internal/domain`, `internal/storage` | steht |
| **Karte** | `src/data/card.cpp` | `domain.Card`, `domain.CardRow` | ported | dito, dazu `internal/domain/kartennotizen_test.go` und `internal/flaechen/kartennotizen_test.go` | steht. **`card.notes` hatte bis zum 2026-08-27 zwei Darstellungen** — ein Feld, das niemand füllte, und den Feldwert aus der Datei; über MCP gewann die leere. Jetzt gibt es eine (#1168, Befund Z1-01). **Was fehlt**: `time_created` und `time_modified` — MSE führt sie an jeder Karte, hier gibt es sie nicht |
| Feld | `src/data/field.cpp` | `domain.Field` | ported | dito | steht |
| Spiel | `src/data/game.cpp` | `msepaket`, `service.Vorlage` | ported | `internal/msepaket` | steht |
| Stilbogen | `src/data/stylesheet.cpp` | `msepaket`, `service.Planbauer` | ported | dito | steht |
| Schrift | `src/data/font.cpp` | `render.Schrift`, `internal/schriften` | ported | `internal/render` | steht |
| Symbolschrift | `src/data/symbol_font.cpp` | `msepaket.Symbolschrift`, `service.stilsymbole` | ported | `internal/service` | steht |
| Symbol | `src/data/symbol.cpp` | `internal/msepaket`, `internal/service/symbolablage.go` | ported | `internal/service/symbolablage_test.go` | das Format hinter `.mse-symbol` — **gelesen und geschrieben** seit Phase X |
| Auszählung | `src/data/statistics.hpp` Zeilen 25–45 (`StatsDimension`), `src/data/statistics.cpp` | `app.Statistik`, `app.Verteilung` | adapted | `statistik.test.tsx` | **Hier stand `ported`, und das behauptete zu viel** (#1291, angestrichen vom Audit). Gemessen am 2026-08-30: `magic.mse-game` erklärt **50 `statistics dimension:`** je Sprachfassung, bei **13** Sprachen also 650 — über alle Spiele des Packs **685**. Jede trägt `name`, `icon`, `description`, ein **Skript**, das den Wert rechnet, dazu `group:`-Reihenfolge, `split list:`, `show empty:`, Farben und numerische Klassen. **Die erste Zahl hier war 698 und stammte aus einem rohen `grep`**: es zählte 13 auskommentierte `#statistics dimension:` mit und nannte 18 Sprachen, wo 13 stehen (angestrichen vom Zweitblick über Phase 4.4, Befund 5). Gezählt wird jetzt am Zeilenanfang. **Unser Quelltext kennt den Schlüssel nicht** — kein Treffer in `internal/` · **Abweichung:** gezählt werden Karten, Felder, leere Karten, die Belegung je Feld und die Werteverteilung eines Feldes. Das ist eine **andere, einfachere** Auskunft, keine halbe · **Plan:** die erklärten Dimensionen sind die Zeile darunter, #1338. **Seit #1985 auch die automatischen** je Kartenfeld mit `show statistics:` (`Game::validate`, `src/data/game.cpp` Zeilen 68–77), mit den Farben aus `choice colors:` |
| Erklärte Auszählungsdimensionen | `src/data/statistics.hpp` Zeilen 25–70, `src/gui/set/stats_panel.cpp` | `msepaket.Auszaehldimension`, `domain.Auszaehlen`, `service.Auszaehldienst` | ported | `internal/msepaket`, `internal/domain`, `internal/service`, `internal/flaechen`, `scripts/nahtlauf.mjs`; dazu ein Lauf gegen **MSE2 selbst** hinter `fremdpakete` | **Gebaut am 2026-08-31** (#1338). An allen vier Flächen. Gelesen werden alle elf Angaben, gezählt wird wie `GraphData::GraphData` — erklärte Reihenfolge samt leerer Gruppen, Klassen über `bin size:`, `split list:` an Kommas, `show empty:`. **Die Abnahme ist am Orakel gemessen** (§15): für `White.mse-set` gibt „color category" bei **3.761 von 3.761** Karten denselben Wert wie MSE2. **Alle 50 Dimensionen von `magic.mse-game` rechnen** — 46 waren es am Morgen des 2026-08-31, und die vier anderen sind an diesem Tag gefallen: `keyword_usage` (#1354), `global_value` (#1355) und zweimal `set.cards` (#1356). „Rechnen“ heißt: **keine bricht mit einem Fehler ab** — nicht, dass jede etwas findet. **Nicht gelesen wird `statistics category:`**: der Schlüssel kommt im ganzen Pack null-mal vor (gemessen am 2026-08-31), und MSE2 baut die Kategorien dann selbst aus den Dimensionen |
| Sprache | `src/data/locale.cpp` | `msepaket` liest `.mse-locale` | ported | `internal/msepaket` | steht — **gelesen, aber nicht genutzt**: MSE2 nimmt die Feldbeschriftungen daraus, hier fragt kein Dienst `msepaket.Sprachpaket` (FMT-01, #1690). Die Marke gilt dem Lesen; die Nutzung ist ein Kompatibilitätsrest (`docs/FAEHIGKEITEN.md`, *Operationen je Paketart*) |
| Exportvorlage | `src/data/export_template.cpp` | `service.Exporteur` | ported | `vorlagenausfuhr.test.tsx` | steht |
| Einstellungen | `src/data/settings.cpp` | `storage`, `app.Einstellungen` | adapted | `einstellungen.test.tsx` | im Bestand statt in einer Datei |
| Schnellfilter der Kartenliste | `src/data/filter.cpp`, `Card::contains` (`src/data/card.cpp`, Zeilen 55–61) | `domain.ZerlegeSchnellfilter`, `service.Kartenlistenbau.Filtern` | adapted | `schnellfilter_test.go`, `kartenfiltergerechnet_test.go`, `internal/flaechen/kartenfilter_test.go` | die Sprache aus `parse_quicksearch_query`, an allen vier Flächen. **Gerechnete Werte seit FP-01** (#2155): ein freier Begriff sieht die sichtbaren Spalten, `feld:` das genannte Feld, und gespeichert schlägt gerechnet — MSE2 durchsucht jeden Wert, wie er nach den Skripten steht. Gerechnet wird mit dem `init script` des Spiels, einmal je Anfrage; **ohne den Stil der Karte** |
| Suchen im Set | `CardsPanel::doFind` (`src/gui/set/cards_panel.cpp`, Zeile 445), Fassung f9d9356 vom 2021-07-06 | `service.Suche` (`Suchen`, ⌘F) | adapted | `suche.test.tsx` | **Volltext über SQLite — über den Bestand oder ein Set** — statt Feld für Feld im offenen Set; der Schnellfilter oben ist der Weg, der MSE2s Filter entspricht. Bis WART-01 (#2172) stand hier eine Zeile „Filter → `service.Suche`" für beides |
| Schlüsselwort | `src/data/keyword.cpp` | `msepaket.Schluesselwort`, `service.Wortliste` | ported | `internal/msepaket`, `internal/service`; am gezeichneten Bild nachgesehen | **steht auf der Karte** — samt `condition:`, `default_expand:` und `correct_case`. In allen zehn kanonischen Sets ist `automatic reminder text` leer, und dann zeigt MSE keinen; wer sie setzt, bekommt sie. Was fehlt: der `<kw-X>`-Zustand je Stelle und Platzhalter für fehlende Parameter. **Pflegen lässt sich jedes** — an allen vier Flächen, seit #1062; die Zeile „Schlüsselwörter" oben sagt, wie |
| Pack | `src/data/pack.cpp` | `msepaket.Pack`, `service.Packzieher` | ported | `internal/msepaket`, `internal/service` | **alle zehn Auswahlarten**, samt Tiefen, Gewichten und gleichmäßiger Verteilung; der Zufall ist aussäbar. Der Filter wird **eingesetzt** — wer ihn auswertet, entscheidet der Aufrufer |
| Seteigene Packarten | `src/data/pack.cpp`, Zeilen 373–392; `src/data/set.hpp`, Zeile 57 | `domain.PackArt`, `service.PackartenMitVorrang` | ported | `TestEineGleichnamigeSetpackartErsetzt`, `TestDerPackbereichZiehtAusDerSetart` | **„Additional/replacement pack types"** — MSE sucht zuerst im Set, dann im Spiel; eine neue Art kommt hinzu, eine gleichnamige ersetzt. Bis zum 2026-08-25 fielen sie in den Rohweg und waren unsichtbar (#1081) |
| Wortliste | `src/data/word_list.cpp`, `src/gui/value/text.cpp` (Zeilen 109–300) | `msepaket.Wortliste`, `service.Wortmenue` | ported | `internal/msepaket`, `internal/service`, `internal/app`, `internal/flaechen`; an einer echten Karte nachgesehen | **die Auswahl steht am Feld** — Kästchen für die Präfixe, Liste für die Typen, Untermenüs für die Gruppen, an allen vier Flächen. Zwei Eigenheiten sind mitportiert und benannt: das Leerzeichen hinter einem Komma und der wandernde Strich. **Adaptiert ist die Gestalt**: MSE setzt einen Pfeil in den Text, hier stehen Bedienelemente darunter. Was fehlt: das Menü schließt sich nach jeder Wahl (#1075) |
| Karten aus einem Skript | `src/data/add_cards_script.cpp`, `src/gui/set/cards_panel.cpp` Zeilen 160–176 und 354–357 | `msepaket.Kartenskript`, `service.Kartenskriptdienst`, `storage.KartenAnlegen` | ported | `internal/msepaket`, `internal/service`, `internal/flaechen`, `scripts/nahtlauf.mjs`; an `magic.mse-game` gefahren | **Gebaut am 2026-08-31** (#1337). Alle vier Flächen; die zwölf Blöcke aus `magic.mse-game` laufen. `perform` hängt sie in **einem** Zug ans Set — hier eine Transaktion mit **einer** Verlaufsgruppe, ein ⌘Z nimmt den ganzen Lauf zurück. Die Kopie einer schon im Set liegenden Karte ist mitportiert. **Adaptiert ist die Gestalt**: MSE2 zeigt ein Aufklappmenü an „Karte hinzufügen“, hier steht ein Kasten mit einem Knopf je Skript — ein Menü über einem Knopf wäre ein Menü in einem Menü (§6). Was fehlt: `enabled:` wird gelesen und noch nicht ausgewertet |
| Installer | `src/data/installer.cpp`, `src/util/io/package_manager.cpp` (Zeilen 355–400) | `msepaket.SchreibeInstaller`, `msepaket.OeffneInstaller`, `service.Paketverwalter.Weitergeben`, `service.Paketverwalter.Einlesen`, `internal/bezug` | ported | `TestEinInstallerTraegtDiePaketeWieMSE2`, `TestEinInstallerGehtUeberJedeFlaeche`, `internal/bezug` | **Seit FMT-02 geschlossen** (#1691): ein `.mse-installer` wird geschrieben — Byte für Byte wie MSE2 — und gelesen. **Adapted** ist das Einlesen an einer Stelle: ein vorhandenes Paket wird übergangen statt ersetzt. Der Bezug an einem Stand bleibt daneben |

---

## Bereich 6 — Die Dateiformate

`src/data/format/`, **zehn Dateien**.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| `.mse-set` lesen und schreiben | `src/data/format/mse2.cpp` | `internal/msepaket` | ported | `internal/msepaket` gegen echte Pakete, `TestSetrundlaufMitMSE2` gegen das Orakel | steht. **Am Orakel gemessen am 2026-10-10** (F05, #2247): sechs Sets, zwei Spiele, beide Richtungen ohne Abweichung dieser Anwendung (`docs/FORMATE.md`, „Beide Programme lesen dieselbe Datei gleich“). **Dabei gefunden und behoben**: der setweite Stilblock stand zusätzlich unter `set info:` (MSE2 warnte beim Öffnen), eine hier gesetzte Stiloption kam nicht in die Datei, und MSE2s Rechtschreibmarken wurden als Text gelesen. **Zwei benannte Abweichungen**: MSE2 speichert einen Wert nach seinem Feldskript, diese Anwendung den gelesenen; und diese Anwendung schreibt unbekannte Schlüssel zurück, die MSE2 weglässt |
| MSE **1** lesen | `src/data/format/mse1.cpp` | — | unsupported | — | **Ein Set aus MSE 1 lässt sich nicht öffnen.** **Nachgemessen am 2026-08-30**: `MSE1FileFormat::canImport()` gibt `true`, `canExport()` gibt `false` (`src/data/format/mse1.cpp` Zeilen 27–28) — die Begründung stimmt also, MSE2 liest das Format und schreibt es nicht · **Ablehnung mit Deckung**: das Format ist von 2006, es liegt hier keine einzige Beispieldatei vor, und was nicht prüfbar ist, wird nicht behauptet (§10) |
| Ausgabe über eine HTML-Vorlage | `src/data/format/html.cpp` | `service.Exporteur` | ported | `vorlagenausfuhr.test.tsx` | steht |
| Bild schreiben | `src/data/format/image.cpp` | `service.Bildausfuhr`, `internal/svg` | ported | `bilder.test.tsx` | steht, und weiter: zusätzlich SVG |
| Apprentice | `src/data/format/apprentice.cpp` Zeile 580, `src/gui/set/window.cpp` Zeilen 188 und 683 | — | unsupported | — | **Kein Export nach Apprentice.** Hier stand „kein **Import**", und die Richtung war falsch: MSE2 **schreibt** eine Apprentice-Datenbank (`export_apprentice`, Menü *Export → Apprentice…*) und liest keine. Nachgemessen am 2026-08-30 (#1291, Werkstück C) · **Ablehnung mit Deckung**: das Programm ist seit Jahren eingestellt, und der Weg nach draußen führt heute über MTG JSON und die Exportvorlagen — `magic-cockatrice-v2`, `magic-lackeyccg`, `magic-mws` liegen im Pack |
| Magic Workstation | `src/data/format/mws.cpp` Zeile 56, `src/gui/set/window.cpp` Zeilen 189 und 687 | — | unsupported | — | **Kein Export nach MWS** — dieselbe Richtungskorrektur, dieselbe Deckung. **Und hier gibt es einen Ausweg**: das Pack bringt `magic-mws.mse-export-template` mit, und Exportvorlagen laufen seit dem 2026-08-30 an allen vier Flächen (#1291, Werkstück D) |
| MTG Editor | `src/data/format/mtg_editor.cpp` Zeilen 28–29 | — | unsupported | — | **Kein Import aus MTG Editor** — und diese Richtung stimmt: `canImport()` gibt `true`, `canExport()` gibt `false`. Nachgemessen am 2026-08-30 · **Ablehnung mit Deckung**: das Programm ist eingestellt, und ein Format ohne erreichbare Beispieldateien lässt sich nicht prüfen (§10 — was nicht prüfbar ist, wird nicht behauptet) |
| Zwischenablage | `src/data/format/clipboard.cpp` | `service.Kartenablage` | ported | `internal/service/kartenablage_test.go`, `kartenablage.test.tsx`, `internal/flaechen/ablagelauf_test.go` | steht — Karten gehen von Set A nach Set B, über die Knöpfe der Kartenlasche und über ⌘C/⌘X/⌘V. **Der Inhalt ist MSE2s eigenes Format**, also auch der Weg in ein anderes Programm (#1168). **Was fehlt**: der Transportvertrag. MSE2 legt Karten unter `application/x-mse-cards` ab und prüft beim Einfügen nur dieses Format (`CardsDataObject::format`, `src/data/format/clipboard.cpp`, Zeile 61; `canPaste`, `src/gui/control/card_list.cpp`, Zeile 137, Fassung `f9d9356` vom 2021-07-06) — Text legt es für Karten gar nicht daneben, die Umwandlung steht dort als `// TODO` (`clipboard.cpp`, Zeilen 144–146). Wails gibt nur die Textablage her. **Und eine Datei überbrückt es auch nicht**: der Ablageinhalt ist ein `game_name:` mit `card:`-Blöcken, ein `.mse-set` dagegen ein Paket mit Spiel, Stil und Setangaben (`src/data/format/mse2.cpp`, Zeilen 19–35; bestätigt durch `magic-set-generator`, `src/mse.rs`, Zeilen 124–154). Der Weg dorthin führt über ein **Set**: einfügen und sichern. Angestrichen als Befund 3 des Zweitblicks und berichtigt mit Befund 5 des Nachlaufs über Phase 3.7. **Ein Bildfeld nimmt seine Datei mit — soweit sie hier liegt.** Der Inhalt trägt die Bytes als `image_data`, und beim Einfügen entsteht im Zielset eine eigene Datei mit eigenem Namen (#1257). **Was mit dem Set kam, geht nicht mit**: die Bilder eines geöffneten Sets liegen in seinem Herkunftspaket ([0029](entscheidungen/0029-bilder-liegen-neben-dem-set.md)), und sie durch die Zwischenablage des Systems zu schicken wären bei `Multicolor.mse-set` 229 MB. Dann geht der **Verweis** mit — und das Einfügen sagt es, denn derselbe Name meint im Zielset ein anderes Bild (#1282). **Die drei fensterlosen Flächen tragen ohnehin nur den Verweis**: dort gibt es keine Zwischenablage des Systems, über die ein Mensch etwas hin- und herträgt |
| Bild zu Symbol | `src/data/format/image_to_symbol.cpp` | `service.BildNachzeichnen`, `app.SymbolNachzeichnen` | ported | `bildschwelle_test.go`, `bildumriss_test.go`, `bildnachzeichnen_test.go`, `bildrandfaelle_test.go`, `internal/flaechen/symbolnachzeichnen_test.go` | **Steht seit dem 2026-08-31** (#1340). Schwelle, Rundgang und die sieben Vereinfachungsschritte, samt der MSE-1-Erkennung an 60 × 80. **An allen vier Flächen**, und das ist die Ausnahme unter den Symbolfähigkeiten: die anderen sind Handgriffe auf einer Zeichenfläche, diese ist eine Umrechnung. **Gemessen am 2026-08-31** an einem Kreis von 120 × 120: der Rundgang gibt 340 Punkte, `punkteEntfernen` bringt sie auf 36, und 16 der 36 Strecken werden Geraden — die Zahl folgt aus MSEs Schwelle `0.0002` und ist nicht gewählt. **Adaptiert ist eines**: der unmögliche Stand der Sprungtabelle bleibt stehen und läuft in eine Schranke, statt wie bei MSE zu werfen |
| `include localized file:` — eine Sprachfassung einbinden | **keine verfügbare Quelle kennt ihn** | `msepaket.Einbinder.InSprache` | uncertain | `TestEinSprachverweisWirdAufgeloest` | steht — **und die Bedeutung ist weiter erschlossen, nicht belegt.** Was am 2026-08-30 dazukam, ist der **Grund**, warum keine Quelle sie belegt: `mse2-quelle` f9d9356 ist `project(magicseteditor VERSION 2.1.2)` (CMakeLists Zeile 3), das Pack verlangt `mse version: 2.5.8`; `magic-set-generator` 1b8fb7fa kennt den Schlüssel auch nicht. Im Pack steht er **zweimal** (`statistics`, `keywords`), und neben `keywords` liegen **13** Sprachfassungen. **Die Marke stand einen Tag lang auf `adapted`, und das war falsch**: `adapted` sagt „bekanntes Verhalten, begründet abgewichen" — hier ist das Verhalten gar nicht bekannt. Angestrichen vom Zweitblick über Phase 4.4 (Befund 4) · **Plan:** eine MSE 2.5.x als Gegenprobe, oder die Marke bleibt |
| Die Formatliste selbst | `src/data/format/formats.cpp` | `service.Importeur`, `app.MTGJSONOeffnen` | adapted | `mtgjsonoeffnen.test.tsx` | **MTG JSON kommt dazu** und hat in MSE2 keine Entsprechung — gesucht in `src/data/format/`, alle zehn Dateien angesehen |

---

## Bereich 7 — Die Paketarten

**Acht**: die sieben, die `internal/msepaket.Art` führt, und `.mse-symbol`.

| Einheit | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- |
| `.mse-set` | `msepaket.ArtSet` | ported | `internal/msepaket` | lesen **und** schreiben |
| `.mse-game` | `msepaket.ArtSpiel` | ported | dito | lesen |
| `.mse-style` | `msepaket.ArtStil` | ported | dito | lesen |
| `.mse-symbol-font` | `msepaket.ArtSymbolschrift` | ported | dito | lesen |
| `.mse-include` | `msepaket.ArtEinschub` | ported | dito | lesen |
| `.mse-locale` | `msepaket.ArtSprache` | ported | dito | lesen |
| `.mse-export-template` | `msepaket.ArtExportvorlage` | ported | dito | lesen |
| `.mse-symbol` | `internal/service/symbolablage.go` | ported | `internal/service/symbolablage_test.go` | **gelesen und geschrieben** seit Phase X |

---

## Bereich 8 — Die Sprache

`src/script/`, **neun Dateien**. `internal/msescript` ist vollständig
abgedeckt (100 %); vermessen ist hier, was von der **Umgebung** der Sprache da
ist.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| Die englischen Sprachfunktionen | `src/script/functions/english.cpp` | `msescript.bibliothek_englisch.go`, `bibliothek_englischhinweise.go` | ported | `TestDieEnglischenZahlwoerter`, `TestDieEnglischenHinweiseWerdenAufgeloest` | alle sieben; **206 Aufrufe** im Pack. Ohne `process_english_hints` scheiterte das Zusammensetzen der Erinnerungstexte |
| Die Standardbibliothek — Grundfunktionen | `src/script/functions/basic.cpp` | `msescript.bibliothek*.go` | adapted | `internal/msescript`, `internal/flaechen/skript_test.go` | **59 von 59 angemeldeten Namen** aus `basic.cpp`, nachgezählt am 2026-09-07 im Stand `f9d9356`: 59 `ctx.setVariable`-Zeilen gegen 55 `SCRIPT_FUNCTION`-Definitionen. **Drei sind definiert und nicht angemeldet** — `position_of` (angemeldet als `position`), `warning_if_neq` und `rule`. **Hier stand am 2026-09-07 einen Tag lang „51 von 51“ mit der Behauptung, `rule` und `warning_if_neq` hätten gefehlt.** Sie fehlten nicht: MSE meldet sie nicht an. Sie waren trotzdem eingetragen worden, und der Nachlauf über die Phasen 4.0 bis 4.13 hat es angestrichen (Befund 4). Beide sind wieder fort, und der Namenswächter hält seither **beide** Richtungen. **Was wirklich fehlt, ist `assert`** — in MSE eine feste Form des Zerteilers, die `assert(x == y)` in `warning_if_neq("…", _1: x, _2: y)` übersetzt (`parser.cpp`, Zeilen 643–664). Gemessen über den ganzen `full-magic-pack`: `assert(` steht in **einer** Datei, `magic.mse-game/unit_tests` — MSEs eigener Testdatei; keine Vorlage und kein Stilbogen ruft es. **Hier stand bis zum 2026-09-06 „33 von 51“ mit einer Liste von achtzehn fehlenden Namen** — sechzehn davon waren am 2026-09-02 dazugekommen, und die Zeile hat es nicht mitbekommen. **Die Zeile stand bis zum 2026-08-28 gar nicht in dieser Matrix** — der Wächter überging Unterordner und sah `src/script/functions/` nie (#1170) |
| Die Standardbibliothek — Bildfunktionen | `src/script/functions/image.cpp` | `msescript.bibliothek_bild.go`, `service.stilumgebung.go` | adapted | `internal/msescript/bibliothek_bild_test.go`, `internal/render/bilddrehung_test.go`, `eingebautesbild_test.go`, `internal/werkzeuge/bildverknuepfungen_test.go` | **alle 18, und seit dem 2026-09-03 führt der Rasterer jede aus.** Hier stand „15 von 18", und die Zahl war in beide Richtungen falsch: `symbol_variation` fehlte nie, es wird von `internal/service` gesetzt statt von der Bibliothek (es braucht Paketzugriff) — und `rotate`, sein Alias `rotate_image` und `built_in_image` waren zwar registriert, wurden aber vom Rasterer **nicht ausgeführt**, was die Matrix gar nicht führt. Beides ist erledigt: ein Wächter hält jetzt beide Listen gegeneinander. **Zwei benannte Abweichungen**: ein schräger Drehwinkel wird gemeldet statt neu abgetastet (im vollen Pack steht `rotate(` **null**-mal), und `built_in_image` zeichnet `bool_yes`/`bool_no` selbst, statt MSEs Ressourcen mitzunehmen (§11) — die übrigen 57 Ressourcennamen sind Bedienelemente von MSEs Fenster und kommen im Pack nicht vor |
| Die Standardbibliothek — Muster und Text | `src/script/functions/regex.cpp` | `msescript.bibliothek_muster*.go` | ported | `internal/msescript/bibliothek_muster*_test.go` | **alle fünf**: `replace_text`, `filter_text`, `break_text`, `match`, `split_text` |
| Die Standardbibliothek — die Wahl im Editor | `src/script/functions/editor.cpp` | `msescript.bibliothek_wahl.go` | adapted | `internal/msescript` | **7 von 7**. Hier stand „6 von 7; es fehlt `primary_choice`", und der Name war längst eingetragen (`bibliothek_kartenwelt.go:32`) — angestrichen vom Zweitblick über die Phasen 4.0 bis 4.13 (2026-09-07). **Eine Matrix, die etwas als fehlend führt, das da ist, ist so falsch wie eine, die etwas als da führt** — nur fällt es niemandem auf, weil es nach Vorsicht aussieht |
| Die Standardbibliothek — Ausgabe | `src/script/functions/export.cpp` | `msescript.bibliothek_html.go` | adapted | `internal/msescript/bibliothek_html_test.go` | **2 von 9**: `to_html` und `to_text`. Es fehlen `symbols_to_html`, `to_bbcode`, `copy_file`, `write_text_file`, `write_image_file` und `write_set_file`. **`sanitize` stand bis zum 2026-09-07 in dieser Liste** und war längst eingetragen (`bibliothek_zweitnamen.go:50`). **Die fünf schreibenden sind kein Versehen**: eine Exportvorlage, die aus dem Skript heraus beliebige Dateien schreibt, ist die Klasse, gegen die [0044](entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md) gebaut ist — hier schreibt die Vorlage an **einen** Ort, den ein Mensch genannt hat |
| Die Standardbibliothek — eine Karte aus dem Skript | `src/script/functions/construction.cpp` | `service.KartenAusSkript` | ported | `internal/service`, `internal/flaechen` | **Gebaut am 2026-08-31** (#1337). `new_card` gibt eine **Abbildung**, keine Entität — die Frage „was darf ein Skript anrichten“ ist damit beantwortet, ohne sie zu stellen: das Skript **beschreibt** Karten, angelegt werden sie vom Dienst, in einer Transaktion. Feldnamen kommen in beiden Schreibweisen an (`casting cost` wie `casting_cost`); was kein setzbares Feld benennt, wird als Befund geführt statt verschluckt (§8) |
| Die Standardbibliothek — Rechtschreibung | `src/script/functions/spelling.cpp` | `msescript.bibliothek_rechtschreibung.go` | ported | `internal/msescript` | **beide**: `check_spelling` und `check_spelling_word` |
| Die Rücknahme — ein Set | `src/data/action/set.cpp` | `domain.Verlaufsschritt`, `storage.Verlauf` | adapted | `internal/storage`, `internal/service`, `kartenverlauf.test.tsx` | **Adaptiert ist der Ort**: MSE hält den Stapel im Fenster, hier steht er im **Bestand** — damit überlebt er den Neustart und gilt für jede Fläche. Was fehlt: ein sichtbarer Stapel mit Namen je Schritt |
| Die Rücknahme — ein Feldwert | `src/data/action/value.cpp` | dito | adapted | dito | dito; jede Feldänderung schreibt einen Schritt |
| Die Rücknahme — ein Symbol | `src/data/action/symbol.cpp`, `src/util/action_stack.cpp` (Zeilen 21–51) | `haken/useZeichenverlauf.ts`, `ruecknahme.tsx`, `ruecknahmeweg.ts` | ported | `useZeichenverlauf.test.tsx` (11 Fälle), `symbolruecknahme.test.tsx`, `ruecknahme.test.tsx`, `anwendungruecknahme.test.tsx` | **Steht seit dem 2026-08-31** (#1340). ⌘Z nimmt **einen Handgriff** zurück, nicht die Zeichnung. **Drei Regeln sind portiert und keine erfunden**: ein Zug ist ein Schritt (`ExtendableAction`), ein Tastendruck je einer (`point_editor.cpp` ruft `resetActions()` und legt neu an), und nach einer Rücknahme wird nicht verschmolzen (`last_was_add`). **Adaptiert ist der Ort**: MSE hat ein eigenes Fenster mit eigenem `ActionStack`, hier ist der Editor eine Seite — deshalb entscheidet ein Kontext, ob ⌘Z die offene Zeichnung oder den Bestand meint |
| Die Rücknahme — ein Symbolteil | `src/data/action/symbol_part.cpp` | `haken/useZeichenverlauf.ts`, `haken/usePunktwege.ts` | ported | `useZeichenverlauf.test.tsx`, `symbolruecknahme.test.tsx` | **Dieselbe Zeile von der anderen Seite** (#1340): die Handgriffe an **Punkten** — verschieben, ziehen, einfügen, löschen — tragen ihre Marke, und `zugBeginnt` trennt zwei Züge an demselben Punkt. Das ist MSEs `resetActions()` als Zahl statt als Lebensdauer eines Objekts. **Nicht portiert sind die Namen der Schritte** (`getName`): der Verlauf trägt ein Feld dafür, und die Knöpfe zeigen es noch nicht — vertagt, nicht vergessen |
| Zerteiler | `src/script/parser.cpp` | `msescript.Zerteile` | ported | `internal/msescript` | steht |
| Auswerter und Anweisungen | `src/script/script.cpp` | `msescript.Werte` | ported | dito | steht |
| Werte und Typen | `src/script/value.cpp`, `src/script/scriptable.cpp` | `msescript.Wert` | ported | dito | steht |
| Kontext — Sichtbarkeit, Standardbibliothek | `src/script/context.cpp` | `msescript.Kontext` | ported | dito | steht |
| Abhängigkeiten zwischen Feldern | `src/script/dependency.hpp` Zeilen 15–64, `src/script/dependency.cpp`, `src/script/script_manager.cpp` Zeilen 96–200 | — | adapted | `BenchmarkWegPlanbau` | **Nachgelesen und gemessen am 2026-08-30** (#1291). MSE2 meldet beim Laden je Skriptfeld an, welche anderen Felder es liest (`initDependencies`), unterscheidet dabei acht Arten (Kartenfeld, Kartenfeld über **alle** Karten, Setfeld, Stileigenschaft, stilspezifisches Zusatzfeld, zwei Kopierarten, eine Attrappe) und wertet bei einer Änderung **nur die abhängigen** neu aus (`updateAllDependend`). Hier wird stattdessen der ganze Plan neu gebaut · **Abweichung, und was sie kostet:** **21 ms** je Neubau — gemessen an der ersten Karte von `Black.mse-set` (3.788 Karten) mit dem `M15-Magic-Pack`, `BenchmarkWegPlanbau`, Apple M5, 20 Läufe. Bei Tippgeschwindigkeit liegen rund 200 ms zwischen zwei Anschlägen, und die Vorschau wird ohnehin abgebrochen statt abgewartet (§7). **Was diese Zahl nicht sagt**: was ein Set mit vielen `card list`-Skripten kostet, die über **alle** Karten rechnen — dafür ist `DEP_CARDS_FIELD` da, und das ist hier ungemessen. **`dependency.cpp` stand in dieser Zeile nicht**, obwohl es die Analyse selbst ist — die Attrappe, mit der MSE2 ein Skript einmal trocken durchläuft, um zu sehen, was es liest. Der Wächter, der jede Datei verlangt, lief nie: `make fremdpakete` übergab die MSE2-Quelle nicht (QA-01, #1733) |
| Bilder aus Skripten | `src/script/image.cpp` | `service.Bildskripte`, `render` | ported | `internal/service` | steht |
| Skriptverwaltung je Set | `src/script/script_manager.cpp` | `service.Planbauer` | adapted | dito | kein Zwischenspeicher je Feld |
| Profiler | `src/script/profiler.cpp` | — | unsupported | — | siehe Bereich 4 |

---

## Bereich 9 — Der Rasterer

`src/render/`, **siebzehn Dateien** in vier Unterordnern.

| Einheit | C++-Quelle | Hier | Marke | Beleg | Für Nutzende |
| --- | --- | --- | --- | --- | --- |
| Die Karte als Ganzes | `src/render/card/viewer.cpp` | `service.Planbauer` → `render.Zeichne` | ported | Goldbilder, Kanonlauf | steht |
| Text — Satz, Zeilenumbruch, Auszeichnung | `src/render/text/viewer.cpp`, `src/render/value/text.cpp` | `service.Auszeichnungen`, `render.Text` | ported | `internal/service`, Goldbilder | steht. **Der größte Einzelposten des Abstands** (`ROADMAP.md`). **Gewicht, Lage und Schatten der Schrift seit F02** (#2244) |
| Symbole im Text | `src/render/text/symbol.cpp`, `src/render/text/compound.cpp`, `src/render/text/element.cpp`, `src/render/text/font.cpp` | `service.stilsymbole` | ported | `internal/service` | steht |
| Bildwerte | `src/render/value/image.cpp` | `service.bildebene` | ported | `PORTIERUNGSKARTE.md` | steht — **das Schachbrett eines leeren Bildfelds seit F02** (#2244) |
| Auswahlwerte | `src/render/value/choice.cpp`, `multiple_choice.cpp` | `service` | ported | dito | steht |
| Mehrfachwahl als Bildreihe — `image list` | `src/render/value/multiple_choice.cpp`, Zeilen 23–90 | `service.bildreihe` | ported | `service/bildreihe_test.go`, `service/bildreihe_intern_test.go`, `msepaket/feldstilreihe_test.go` | **Steht seit #2297**: die Bilder der gewählten Wahlen in eigener Größe nebeneinander, ab `align_in_rect(alignment, (0,0), …)`, mit `spacing` und `direction`, gegen MSE2s Ausgabe gehalten mit dem Prüfset `Reihenprobe`. Ein **gerechnetes** Wahlbild in einer Reihe wird benannt und übersprungen. 3 Stile, alle unter `extra card style:` |
| Checkliste und Textliste **auf einer Karte** — `checklist`, `text list`, der Text einer `both list` | `MultipleChoiceValueViewer::draw` und `drawChoice`, `src/render/value/multiple_choice.cpp`, Zeilen 23–84 | — | unsupported | Zählung am 2026-10-11 über `full-magic-pack/data` | **Kein Stil zeichnet sie auf eine Karte.** `checklist` steht 46-mal im Pack, und jedes Mal in einer **Einstellfläche**: 31-mal unter `styling style:`, 14-mal unter `default set style:`, einmal unter `option style:` — dort ist es die Kästchenliste, die die Oberfläche hier für jede Mehrfachwahl zeigt (`STILOPTIONEN.md`: „Kästchen, eines je Wahl"; Setfelder über `seiten/Mehrfachwahl.tsx`). `text list` und `both list` stehen **nirgends**. Auf der Karte bliebe bei MSE2 `item_height` zudem null — es setzt nur der Editor im nativen Aussehen (`src/gui/value/multiple_choice.cpp`, Zeile 105) —, und das Kästchen hätte eine negative Größe. Ein Weg ohne einen einzigen Verbraucher wird nicht gebaut; kommt ein Stil, der ihn braucht, wird er portiert (#2297) |
| Farbe, Information, Paketwahl | `src/render/value/color.cpp`, `information.cpp`, `package_choice.cpp` | `service` | ported | dito | steht |
| Symbolwerte | `src/render/value/symbol.cpp` | siehe Bereich 2 | adapted | siehe Bereich 2 | steht seit Phase 4.11 — dieselbe Zeile wie in Bereich 2, und dieselbe Abweichung · **Hier stand bis zum 2026-09-02 `deferred`** |
| Symbolzeichnung und ihre Filter | `src/render/symbol/viewer.cpp`, `src/render/symbol/filter.cpp` | `internal/render/symbolzug.go`, `internal/render/symbolfilter.go` | ported | `internal/render/symbolfilter_test.go` | **gebaut in Phase X**: drei Bereiche — außen, Rand, innen —, außen durchsichtig |

---

## Die Zählung

**Am 2026-09-04 neu erhoben**, über die Tabellen dieser Datei — **fortgeschrieben am 2026-10-03** um die Zeile *Eine Balkengruppe filtert die Kartenliste* (UXR-F6, `adapted`), **am 2026-10-07** um die Teilung *Filter* in *Schnellfilter der Kartenliste* und *Suchen im Set* (WART-01, beide `adapted`) und **am 2026-10-11** um die Teilung der Mehrfachwahl als Liste in die *Bildreihe* (`ported`) und *Checkliste und Textliste auf einer Karte* (`unsupported`, #2297):

| | |
| --- | --- |
| Zeilen insgesamt | **122** |
| `ported` | **79** |
| `deferred` | **0** |
| `adapted` | **34** |
| `unsupported` | **7** |
| `uncertain` | **1** |
| `new` | **1** |

**`deferred` ist wieder leer.** In F02 stand dort für kurze Zeit die Mehrfachwahl als Liste (#2297); die Bildreihe ist gebaut, und für Checkliste und Textliste **auf einer Karte** ergab die Zählung über das ganze Pack keinen einzigen Verbraucher — `checklist` steht nur in Einstellflächen, die die Oberfläche als Kästchen bedient. Die Zeile ist damit `unsupported` mit ihrer Deckung, nicht vertagt. **Leer war `deferred` schon einmal, als Ertrag von Werkstück D der Phase 4.9**
(#1324). Die
drei Zeilen führten sämtlich das **Symbolfeld einer Karte** — gebaut in Phase
4.11 (#1339),
und die Zeilen sind stehen geblieben. Eine Vertagung, deren Phase vorbei ist
und die niemand fortschreibt, ist keine Vertagung mehr, sondern eine
Verschleppung; genau deshalb verlangt das Werkstück, **jede Zeile einzeln**
durchzugehen.

**Was danebensteht, hat sich nicht bewegt**: die sieben `unsupported` tragen
jede ihre Deckung, und die eine `uncertain` wartet weiter auf eine MSE 2.5.x
als Gegenprobe.

**Die Spalte `uncertain` ist leer, und das ist der Ertrag von Werkstück B**
(#1291). Alle
drei sind am 2026-08-30 im Quelltext **nachgelesen** worden:

- **Karten aus einem Skript** war gelesen und fehlte → `deferred`, mit Issue.
  **Seit dem 2026-08-31 ist es gebaut** und steht auf `ported`, zusammen mit
  `new_card` (#1337)
  und den erklärten Auszählungsdimensionen
  (#1338) —
  die drei Zeilen, um die `ported` gewachsen und `deferred` geschrumpft ist.
- **Abhängigkeiten zwischen Feldern** sind gelesen **und gemessen** → `adapted`;
  was der ganze Neubau kostet, steht als Zahl in der Zeile (21 ms).
- **`include localized file:`** ist gelesen, soweit es überhaupt geht → `adapted`;
  **warum** keine Quelle ihn kennt, ist jetzt belegt (MSE 2.1.2 gegen ein Pack
  für 2.5.8).

**Und eine Zeile hat ihre Marke von `ported` auf `adapted` geändert**: die
Auszählung. Sie behauptete,
`src/data/statistics.cpp` sei portiert — 50 erklärte Dimensionen je Spiel sagen
etwas anderes.

**Und eine zwölfte am selben Tag**: die **Verwaltung** der Autoersetzung stand
als Nebensatz einer Menübegründung, statt eine Zeile zu sein
(#1242). Ein
Nebensatz wird nicht gezählt — und was nicht gezählt wird, taucht in keiner
Übersicht auf.

**Elf Zeilen kamen am 2026-08-28 dazu**, und keine davon durch neue Arbeit:
der Wächter über die Quellen überging **Unterordner**, und damit blieben
`src/script/functions/*` — die halbe Standardbibliothek — und
`src/data/action/*` — die Rücknahme — unbesehen. Der Lauf war grün, und diese
Datei behauptete daneben, sie beantworte „was ist unvermessen?". Angestrichen
vom ersten Zweitblick über die Freigabe 0.3.0
(#1170).

**Vier Zeilen sind am 2026-08-26 von `deferred` weitergerückt**, und keine
davon durch neue Arbeit: die Phasen X und Z hatten sie gebaut, und **niemand
hat es hier nachgetragen**. Das Format hinter `.mse-symbol` wird gelesen und
geschrieben, die Symbolzeichnung samt ihren Filtern steht, und nach neuen
Fassungen sieht die Anwendung seit 0.2.0.

**Neun offene Zeilen nannten eine Phase, die längst in der Chronik steht.** Der
Wächter darüber zählte die Marken nach und sah es nicht — genau die Gestalt aus
§10. Seit dem 2026-08-26 trägt **jede offene Zeile einen `**Plan:**`**, und ein
Plan, der auf eine abgeschlossene Phase zeigt, macht die Pipeline rot
(`internal/werkzeuge/luecken_test.go`).

**Diese Zahlen zählt ein Wächter aus den Tabellen darüber nach**
(`internal/werkzeuge/paritaet_test.go`). Der erste Anlauf schrieb sie von
Hand — 83, 48, 14, 6 —, und **fünf von sechs waren falsch**. Eine Zahl in
Fließtext, die niemand nachzählt, veraltet nicht erst; sie ist von Anfang an
eine Behauptung.

**Und der Wächter über den Quelltext hat drei Zeilen nachgetragen**, die beim
Schreiben fehlten: die Basis der Feldeditoren, die Basis der Setbereiche und
die Rasterung von Textwerten. Genau dafür gibt es ihn — er findet, was **nie
angesehen** wurde, und nicht bloß Tippfehler.

**Kein `new`, und das ist eine Aussage.** Diese Matrix vergleicht gegen MSE2;
was dieses Projekt zusätzlich kann — Scryfall, MTG JSON, MCP, REST, der
Druckbogen als PDF, die Sammlung samt Listeneinfuhr —, steht in [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md) und hat
dort keine Entsprechung zu suchen. Eine `new`-Zeile hier hieße: *im
Altprojekt gesucht und nichts gefunden* — und die drei Fälle, in denen das
zutrifft, stehen als `adapted` mit der Angabe, wo gesucht wurde (MTG JSON in
Bereich 6, der Druckbogen und die Paketverwaltung in Bereich 4).

**`uncertain` war die ehrlichste Spalte dieser Datei**: dort war nicht
nachgelesen worden, und bis zum 2026-08-24 stand nirgends, dass es so ist.

**Am 2026-08-30 sind alle drei gelesen worden**
(#1291,
Werkstück B): eines fehlt und ist vertagt, eines ist `adapted` mit einer
**Zahl** (21 ms je Planneubau).

**Das dritte stand einen Tag lang auf `adapted`, und das war falsch.**
`include localized file:` — die Erkenntnis war, **warum** keine verfügbare
Quelle den Schlüssel kennen kann (MSE 2.1.2 gegen ein Pack für 2.5.8). Das ist
ein Fortschritt über das Nichtwissen, aber kein Beleg für das **Verhalten**;
`adapted` heißt „bekannt und begründet abgewichen", und bekannt ist es nicht.
Angestrichen vom Zweitblick über die Phase (Befund 4), und die Marke steht
wieder auf `uncertain`.

**Ein Beleg über die Ursache des Nichtwissens ist kein Beleg über die Sache** —
das ist der Satz, den dieser Fall gekostet hat.

**Die ehrlichste Spalte ist damit `deferred` geworden**, und sie ist gewachsen
und wieder geschrumpft: eine `deferred` steht heute da — Checkliste und Text
einer Liste aus F02 (#2297) —, am 2026-08-30 waren es acht mehr. Zwei davon sind die von Werkstück B — `add cards script` und
`new_card` —, die dritte die von Werkstück C, die **erklärten
Auszählungsdimensionen**, die vierte die von Werkstück D, die
**Parameterhilfen**, und die fünfte die von Werkstück E, die **Verwaltung der
Autoersetzung**. Keine war unklar; alle fünf waren vertagt, und je ein
Werkstück hat sie geholt.

**Die fünfte ist die, die am längsten als Nebensatz stand.** Sie bekam ihre
eigene Zeile erst am 2026-08-28, weil der Nachlauf über Phase 3.7 sie in einer
Menübegründung fand — und drei Tage später war sie gebaut. Eine Zeile zu haben
ist die Vorbedingung dafür, geholt zu werden.

**Und am selben Tag kamen drei auf einmal**: Werkstück F holte das
Nachzeichnen eines Bildes und die Rücknahme je Handgriff, und die zweite zählt
zweimal — als Symbol und als Symbolteil. Damit ist Phase 4.11 durch, und was
noch `deferred` heißt, ist keine Lücke dieser Phase mehr.

**Das ist der Unterschied zwischen den beiden Marken in einem Satz:** die
vertagte wandert, wenn jemand sie holt, die unklare erst, wenn jemand liest.

**Es waren vier bis zum 2026-08-27.** Die Vorschaubilder der Kartenliste sind
in Phase 3.5 nachgelesen und gebaut worden
(#1166) —
nach **zwei** Vertagungen, und das ist die Lehre daran: eine Marke, die eine
offene Frage benennt, wandert nicht von selbst. Sie braucht ein Werkstück.
