# 0048 — Eine Endung wird zugeordnet, sechs nicht

**Datum:** 2026-08-25
**Stand:** angenommen
**Betrifft:** `wails.json`, `build/darwin/Info.plist`, `main.go`,
`internal/app/dateioeffnen.go`

---

## Die Entscheidung

**`.mse-set` wird der Anwendung zugeordnet, in der Rolle `Editor`. Die sechs
anderen MSE-Endungen werden es nicht.**

`.mse-set` ist das **Dokument** dieser Anwendung: eine Datei, die jemand
anlegt, benennt, verschiebt, verschickt und wieder aufschlägt. Ein Doppelklick
darauf ist die selbstverständlichste Handlung einer Dokumentanwendung unter
macOS (#891).

## Die sechs anderen, einzeln

Das Issue verlangt, dass sie **entschieden** sind — nicht, dass sie zugeordnet
werden. Hier stehen sie mit Grund:

| Endung | Was sie ist | Warum nicht zugeordnet |
| --- | --- | --- |
| `.mse-game` | ein Spiel — Kartenfelder, Regeln, Skripte | Kein Dokument, sondern eine **Vorlage**. Sie kommt über den Paketverwalter ins Programm und wird nicht „geöffnet" |
| `.mse-style` | ein Kartenstil | dito — sie gehört zu einem Spiel und ist ohne dieses sinnlos |
| `.mse-include` | ein Baustein für Vorlagen | dito, und sie hat für sich gar keine Darstellung |
| `.mse-locale` | die Übersetzungen eines Spiels | dito |
| `.mse-symbol-font` | eine Symbolschrift | dito |
| `.mse-export-template` | eine Ausfuhrvorlage | dito. Sie **wird** benutzt — über die Ausfuhr eines offenen Sets, nicht über einen Doppelklick |

**Was `.mse-symbol` angeht**: der Symboleditor schreibt sie
(#890), aber
sie gehört zu einem Set und wird von dort geladen. Eine Datei allein
aufzuschlagen ist kein Arbeitsablauf, den es gibt.

## Was die Ablehnung kostet

**Ein Doppelklick auf ein Vorlagenpaket tut nichts.** Wer eines im Finder hat
und einbinden will, geht über die Paketseite.

Das ist eine Zusage, die man später brechen kann und dann begründen muss: eine
Zuordnung, die `PaketInstallieren` ruft, wäre plausibel. Sie ist hier **nicht**
gebaut, weil sie Umfang wäre, den diese Phase nicht hat (§13), und weil eine
Installation ohne Rückfrage etwas anderes ist als ein Dokument, das aufgeht.

## Warum der Weg nicht über `os.Args` allein geht

macOS übergibt eine doppelgeklickte Datei **nicht** in der Kommandozeile,
sondern als Apple-Event `odoc` an die laufende Anwendung. Wails nimmt es im
`NSApplicationDelegate` entgegen und reicht es an `mac.Options.OnFileOpen`
weiter — nachgelesen an
`internal/frontend/desktop/darwin/AppDelegate.m`, `application:openFile:`,
Fassung v2.13.0.

**Beide Wege sind verdrahtet.** `os.Args` gilt für einen Start aus der Shell
oder aus einem Skript; `OnFileOpen` für den Finder. Sie enden in derselben
Methode.

**Und das Ereignis kann vor dem Fenster eintreffen.** Der Finder schickt es
beim Start; ohne Wails-Kontext gibt es niemanden, dem man es melden könnte.
Der Pfad wird deshalb aufgehoben und in `Bereit` nachgeholt.

## Was daran gemessen ist und was nicht

**Gemessen am gebauten Bündel** (§10 — die Reichweite einer Messung ist die
Reichweite ihrer Aussage):

- `plutil -p …/Contents/Info.plist` zeigt `CFBundleDocumentTypes` mit
  `mse-set`, `Editor` und `dokument`.
- Nach `lsregister -f` führt LaunchServices die Bindung `.mse-set` samt
  `iconFiles: Contents/Resources/dokument.icns`.
- Das gebaute Programm mit einem Pfad als Argument gestartet legt das Set im
  Bestand an — geprüft an einem eigenen Zuhause, nicht am Bestand des
  Rechners.

**Nicht gemessen**: der Doppelklick selbst. Er geht über LaunchServices, und
die weckt beim ersten Start Gatekeeper — die Anwendung ist ad hoc signiert und
nicht notarisiert ([0003](0003-keine-notarisierung.md)). Ein Lauf kann dort
nicht klicken; der Schritt steht deshalb in `docs/RELEASE.md` bei der
Handarbeit.

## Verworfen

**Alle sieben Endungen zuordnen.** Der Finder zeigte dann für Vorlagenpakete
ein Symbol dieser Anwendung und öffnete beim Doppelklick eine Anwendung, die
mit der Datei nichts anfängt. Eine Zuordnung ist eine Zusage — und eine, die
nur ein Fenster öffnet, ist eine gebrochene.
