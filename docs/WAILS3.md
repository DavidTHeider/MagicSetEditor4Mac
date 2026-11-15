# Wails 3 — was die Hülle trägt und wie es geprüft ist

**Stand: 2026-09-26** — ENG-02
(#1727),
Vermerk [0132](entscheidungen/0132-wails-3-als-huelle.md).

Diese Datei ist die **Bestandsmatrix** aus ENG-02.AK-02: jede Fähigkeit, die
die Hülle unter Wails 2 trug, mit ihrem Weg unter Wails 3, dem Ort im Code und
dem, was sie **belegt** — und was nicht. „Gebaut" heißt hier nie „am Fenster
gesehen"; was nur am nativen Fenster zu sehen ist, steht in der letzten Spalte
mit dem Drehbuch der Hüllenprobe, das es belegt — oder als **offen**.

## Die Matrix

| Fähigkeit | Wails 2 | Wails 3 | Ort | Belegt durch | Am Fenster (Hüllenprobe) |
| --- | --- | --- | --- | --- | --- |
| **Lebenszyklus** — Start, Bereit, Beenden | `wails.Run`, `OnStartup` | `application.New`, `ApplicationStarted` → `Bereit`, `Run` kehrt nach dem Schließen zurück | `internal/fenster/wails.go` | Bau; Start im Servermodus | `start`: Start über LaunchServices, Beenden über ⌘Q |
| **Fenster** — Titel, 1280 × 860, mindestens 1024 × 700, maximiert/normal/zuletzt | `options.App` | `WebviewWindowOptions`, `StartState` | `internal/fenster/ereignisse.go` | `TestDieFensteroptionenFolgenDemStand` | `start`: maximiert und nicht im Vollbild — **erst seit dem 2026-09-26**: `StartState` wirkte unter macOS 27 nicht, jetzt maximiert `WindowRuntimeReady` |
| **Größe beim Schließen merken** | `OnBeforeClose` + `WindowGetSize` | Haken auf `WindowClosing`, `Size()` | `internal/fenster/wails.go` | `TestBeimSchliessenWirdGemerkt` (der Weg dahinter) | `start`: nach ⌘Q und Neustart wieder da — **gemerkt wird auch beim Beenden**, weil Wails beim Aufräumen ohne `WindowClosing` schließt |
| **Bindungen** — die Fassade | `Bind`, `window.go.app.Bindungen` | `Services`, `Call.ByName("<Paket>.Bindungen.<Methode>")`, abgelegt an derselben Stelle | `internal/fenster/wails.go`, `frontend/src/anbindung.ts` | `anbindung.test.ts`; `TestDerVorsatzDerAnbindungIstDerDerFassade`; Nahtlauf im Servermodus | jedes Drehbuch — sie bedienen die Seite, und die Seite ruft über die Brücke |
| **Ereignisse** — Menü, Datei, Fortschritt, Gast | `EventsEmit(ctx, name, daten…)`, `EventsOn` | `Event.Emit(name, liste)`, `Events.On` — die Werte **immer als Liste** | `internal/fenster/ereignisse.go`, `anbindung.ts` | `TestDieNutzlastIstImmerEineListe`, `anbindung.test.ts` | `eintraege`: jeder Menüeintrag erreicht die Seite und wirkt |
| **Menüleiste** — Anwendung, Ablage, Bearbeiten, Darstellung, Set, Fenster, Hilfe (seit UXR-02 die eigenen Menüs zwischen Darstellung und Fenster, wie macOS sie führt); Sprachwechsel; jüngste Sets | `menu.Menu`, `MenuSetApplicationMenu`; drei Menüs als Rollen, englisch | eigenes Modell in `internal/app` (`Menue`), umgesetzt in `application.Menu`, `Menu.Set`; Anwendung, Bearbeiten und Fenster aus **Einträgen mit Rolle**, beschriftet aus dem Menükatalog (#255); **ein Eintrag kann gegraut sein** (`Menueeintrag.Gegraut` → `MenuItem.SetEnabled`, denn Wails schaltet das Selbstaktivieren ab) und *Darstellung* heißt nach dem Stand des Fensters, den die Oberfläche über `DarstellungsstandMelden` meldet (UXR-F14, #2029) | `internal/app/menuebaum.go`, `internal/app/menuemacos.go`, `internal/app/menuedarstellung.go`, `internal/app/menuenaht.go`, `internal/fenster/menue.go` | die Läufe über die Leiste in `internal/app`; `TestDieLeisteErreichtWailsWieSieGebautIst`, `TestJedeRolleHatInWailsEineWirksameRolle` | `leiste`, `eintraege`: Beschriftungen, Kürzel und Handlungen, die Dienste, Ausblenden, Alle einblenden, Alle nach vorne, der Sprachwechsel; `darstellung`: Titel und Aktivierung von Seitenleiste, Navigator und Inspector gegen das Bild |
| **Dateidialoge** — öffnen, sichern, Ordner | `OpenFileDialog`, `SaveFileDialog`, `OpenDirectoryDialog` | `Dialog.OpenFile()`, `Dialog.SaveFile()` mit Filter und Startordner | `internal/fenster/wails.go` | — (braucht ein Fenster; `.covignore`) | `fix-f02` (öffnen, sichern), `eintraege` (Ordner, Bestand sichern, Fensterbild): der Dialog geht auf, ein Abbruch ändert nichts; die **Wahl** kommt als vorbereitete Antwort |
| **Zwischenablage** — Text | `ClipboardGetText/SetText` | `Clipboard.Text/SetText` (ein `bool` statt eines Fehlers) | `internal/fenster/wails.go` | Verdrahtung: `TestJedeNahtErreichtDasFenster` | `kopieren`: Karten in ein anderes Set — über die **Ersatzablage** des Abnahmebaus, nicht über `NSPasteboard` |
| **Finder** — Doppelklick auf ein `.mse-set` | `Mac.OnFileOpen` | `FileAssociations` + `ApplicationOpenedWithFile` | `internal/fenster/wails.go`, `build/darwin/Info.plist` | `open -a … probe.mse-set` gegen das Bündel (Protokoll) | `kernwege`: bei geschlossener und bei laufender Anwendung |
| **Ablegen** — absolute Pfade und Stelle | `DragAndDrop.EnableFileDrop`, `OnFileDrop(x, y, pfade)` | `EnableFileDrop`, `WindowFilesDropped` → Ereignis `datei:abgelegt`; das Feld aus der Stelle (`data-bildfeld`), weil die Webansicht unter Wails 3 kein `dragover` mehr sieht | `internal/fenster/ereignisse.go`, `useAbgelegteDateien.ts`, `Kartenvorschau.tsx` | `TestEinAblegenTraegtStelleUndPfade`, `anbindung.test.ts`, der Lauf über `bildfeldAn` | `kernwege`: ein Bild auf ein Bildfeld, ein Set ins Fenster — ab AppKit, nicht aus dem Finder |
| **Bildweg** — Vorschau, Spielbild, Tischbild | `AssetServer.Handler` (Rückfall hinter den Assets) | `AssetOptions.Middleware` — eine Weiche nach Vorsatz, `vorschauweg.Weg.Vor`; dort, weil die Vorsätze dort stehen und die Hülle so kein `net/http` braucht (#54) | `internal/vorschauweg/vorschauweg.go` | `TestDieWeicheTrenntBildUndOberflaeche` | `fix-f02` (das Bild der Karte), `fix-f04` (der Tisch); Abzüge unter `$TMPDIR/huellenprobe/` |
| **Über** | `Mac.About` | `Options.Description`, Rolle `About` | `internal/fenster/wails.go` | — | `eintraege`: »Über MagicSetEditor4Mac« führt auf den Abschnitt der Hilfe, der Lizenz und Inhalte Dritter nennt — seit #255 ein eigener Eintrag und kein Dialog von macOS |
| **Startfehler** | ein eigenes Modal vor dem Fenster | unverändert (`app.Startfehler`, ohne Wails) | `main.go` | wie vorher | — |
| **Bündel** — Info.plist, Symbole, Saat, Signatur | `wails build` + `postBuildHooks` | `scripts/bauen.sh` (Oberfläche, Go mit `production`, Bündel, `iconutil`, Saat) | `scripts/bauen.sh`, `build/darwin/Info.plist` | `TestDasBauskriptLegtDieSaatInsBuendel`, `TestJederWegZumBuendelGehtUeberDasBauskript`; ein gebautes Bündel (`plutil`, `codesign -dv`) | Gatekeeper-Freigabe wie bisher |
| **Bauschalter** — kein Zugang im Freigabebau | `wails build` ohne `-devtools` | `production` und sonst keiner; `release.sh` liest sie am Programm (`go version -m`) | `scripts/releaseschalter.sh` | `scripts/releaseschalter.test.sh` | — |
| **Entwicklung** | `wails dev` — Fenster **und** Browser unter `34115` | zwei Aufrufe: `make dev` (Fenster) und `make dev-browser` (Bauschalter `server`, `34115`); beide nehmen die Oberfläche von Vite (`FRONTEND_DEVSERVER_URL`) | `scripts/entwickeln.sh`, `Makefile`, `.claude/launch.json` | der Nahtlauf (`scripts/nahtlauf.mjs`) über den Browserweg | — |

## Was sich im Verhalten ändert — und warum es so bleiben darf

| Unterschied | Wails 2 | Wails 3 | Einordnung |
| --- | --- | --- | --- |
| **Bearbeitenmenü** | Rückgängig, Wiederholen, Ausschneiden, Kopieren, Einsetzen, Alles auswählen — englisch | seit #255 selbst gebaut und in der Sprache des Fensters: Widerrufen, Wiederholen, Ausschneiden, Kopieren, Einsetzen, Löschen, Alles auswählen, dazu die Sprachausgabe. **Ohne** „Einsetzen und Stil anpassen": Wails 3 gibt der Rolle keinen Selektor | von macOS erwartet; was Wails in seinen eigenen Rollenmenüs führt, tut an drei Stellen nichts (`UnHide`, `Front`, `PasteAndMatchStyle` ohne Selektor) — deshalb baut die Anwendung sie selbst |
| **Kontextmenü der Webansicht** | im Freigabebau ganz aus | im Freigabebau **nur** in Textfeldern und über einer Auswahl (Ausschneiden, Kopieren, Einsetzen) | eine Verbesserung ohne eigenes Zutun; die eigenen Kontextmenüs (`Kontextmenue.tsx`) bleiben, wie sie sind |
| **Ablegen außerhalb der Wurzel** | jedes Ablegen im Fenster | nur über einem Element mit `data-file-drop-target` — die Wurzel trägt es | gleich, solange die Wurzel die Marke trägt |
| **Vorschau ohne Anwendung** (`npm run vorschau`) | jede Bindung „fehlt" | jeder Ruf scheitert am Netz (kein Go dahinter) | beides zeigt nichts; nur der Wortlaut des Fehlers ist ein anderer |

## Die Kernwege am Fenster — die Hüllenprobe

**ENG-02.AK-04 verlangt die Kernwege vor und nach dem Wechsel.**

**Vorher** ist der Stand, an dem das Audit der Iteration 0.6.0 den Bau mit
Wails 2 gemessen hat (Ausgangsstand `4cf2f38b`, festgehalten mit dem Tag
`basis/wails2`, 0132). Dort standen die Befunde, die #255 und FIX-F02 bis
FIX-F04 beheben: eine halb englische Leiste, ⌘S, das ein neues Set nicht
sicherte, ein CSV-Rückweg über eine Bindung, die die Oberfläche anders rief,
und ein Tischformular, das eine Stelle mitschickte, wo keine hingehört.
**Nachher** fährt die **Hüllenprobe** sie (`make huellenprobe`, seit dem
2026-09-26): das Bündel mit der Abnahmeleitung (`internal/fenster/abnahme.go`),
gestartet über LaunchServices wie ein Doppelklick, je Drehbuch ein frischer
Bestand mit eigenem Schlüsselbund — und **unter einer eigenen Kennung**
(`io.github.davidtheider.magicseteditor4mac.abnahme`, bis 0.8.0
`com.wails.MagicSetEditor4Mac.abnahme`; Vermerk
[0191](entscheidungen/0191-die-kennung-ist-io-github-davidtheider.md)): mit der der Anwendung teilte das
Bündel deren Vorgaben und den Speicher ihrer Webansicht, und ein Dialog der
Probe öffnete dort, wo die Anwendung dieses Rechners zuletzt war. Der Beleg
steht je Drehbuch in [`belege/huellenprobe/`](belege/huellenprobe/).

| Kernweg | Drehbuch | Was es belegt |
| --- | --- | --- |
| 1 Starten, ⌘Q, „wie zuletzt" | `start` | das Fenster geht maximiert auf und nicht im Vollbild; ⌘Q beendet; eine Größe kommt nach dem Neustart wieder |
| 2 Die Leiste, die Sprache | `leiste`, `eintraege` | jeder Eintrag der Anwendung, von Bearbeiten und Fenster mit Kürzel und Handlung, die Dienste beim Öffnen; **jeder** Eintrag von Ablage, Set, Darstellung und Hilfe mit seiner **Wirkung** — Seite, Anhaken, Widerruf im Bestand, Seitenleiste, Zurücksetzen, Sicherung, Schließen, Zuletzt benutzt; die Leiste folgt der Sprache |
| 3 ⌘O, Sichern unter, ein Ordner für MTG JSON | `fix-f02`, `eintraege` | der Dialog geht auf, ein Abbruch ändert nichts, das Gewählte wird geschrieben oder gelesen; der Ordnerdialog wählt Ordner |
| 4 Doppelklick im Finder | `kernwege` | bei geschlossener und bei laufender Anwendung, im selben Fenster |
| 5 Ziehen | `kernwege` | ein Bild auf ein Bildfeld führt zum Zuschnitt, ein Set ins Fenster wird geöffnet |
| 6 Kopieren und Einsetzen | `kopieren` | fünf Karten aus einem Set in ein neues, leeres — dort über »Einfügen« im Leerzustand |
| 7 Der Tisch | `fix-f04` | eine Partie fängt an, WebGL zeichnet, eine Karte geht zonengültig auf die Hand, das Gegenüber sieht sie verdeckt |
| 8 Startfehler und Wiederherstellung | `kernwege` | ein unbrauchbarer Bestand führt zur Wiederherstellung und nicht in einen stillen Abbruch |
| 9 Profilwechsel mit zweitem Faktor | `profilwechsel` | ohne Code wird gefragt, ein falscher abgewiesen, der richtige wechselt |
| Die Inventur (UX-01.AK-02) | `inventur` | jede der 25 Seiten öffnet sich über die Befehlspalette, ohne Fehlermeldung, ohne waagerechtes Rollen; ein Abzug je Seite in `belege/huellenprobe/inventur/` |
| Ausgangswerte (UX-01.AK-03) | `ausgangswerte` | Handlungen und Maschinenzeit für Kartenanlage, Import, Deckbau, Spielstart und Wiederherstellung (`UX-ARCHITEKTUR.md`) |

**Was die Probe nicht belegt**, und warum es bleiben darf:

- **Die Hand vor AppKit.** Tastendrücke und Klicks kommen aus der
  Ereignisschlange der Anwendung und nicht von Tastatur und Maus; ein Ziehen
  beginnt in AppKit und nicht im Finder. Ab AppKit nehmen sie denselben Weg —
  Kürzel der Leiste, Erster Antwortender, Webansicht.
- **Die Wahl im Dateidialog.** Er läuft in einem Dienst von macOS und lässt
  sich von der Anwendung aus nicht bestätigen (`ok:` ist dort nicht
  umgesetzt). Die Probe öffnet den echten Dialog, liest ihn und bricht ab;
  das Bestätigen kommt als **vorbereitete Antwort** über dieselbe Naht.
- **`NSPasteboard`.** Kopiert wird im Abnahmebau in eine Ersatzablage — die
  Zwischenablage dieses Rechners fasst die Probe nicht an.
- **Ob ein Mensch findet, was er sucht**, und was nur ein zweiter Bildschirm
  oder VoiceOver zeigt: [`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md).

**Scheitert einer davon, gilt ENG-02.AK-06**: der Rückweg auf Wails 2 ist der
Stand vor diesem Wechsel — der PR lässt sich als Ganzes zurücknehmen.
