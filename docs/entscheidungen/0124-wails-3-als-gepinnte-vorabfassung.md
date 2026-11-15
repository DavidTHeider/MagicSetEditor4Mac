# 0124 — Wails 3 als gepinnte Vorabfassung, und was das kostet

**Stand:** angenommen · **Datum:** 2026-09-20
**Berührt:** [0030](0030-die-menueleiste-ist-handgeschrieben.md),
[0007](0007-kein-entfernter-lauf.md),
#1727 (ENG-02),
#255

## Warum es diesen Vermerk gibt

**ENG-02.AK-01 verlangt ihn wörtlich**: *„Wails v3.0.0-beta.23 ist das gepinnte
Planungsziel; ADR benennt Toolchain, Vorabversionsstatus und Einschränkungen."*

Und er ist nicht bloß Formsache: **Wails 3 ist am Tag dieser Entscheidung eine
Beta.** Ein Rahmenwerk ohne Fassungszusage unter einer Anwendung, die
Datenbestände von Menschen hält, ist eine Entscheidung mit Preis — und der
gehört aufgeschrieben, bevor die erste Zeile umgestellt wird.

## Was gemessen wurde

**Am 2026-09-20**, gegen den Modulquelltext von `v3.0.0-beta.23` unter
`.go/mod`, nicht gegen die Erinnerung (§1, §10).

### Die Fassungslage

| | |
| --- | --- |
| Neueste v3 | **`v3.0.0-beta.24`** |
| Gepinnt | **`v3.0.0-beta.23`** — der Stand, den ENG-02.AK-01 nennt |
| Neueste v2 | `v2.16.0` (heute gepinnt: `v2.13.0`) |

**Warum nicht beta.24.** AK-01 sagt: *„Kein latest und keine stillschweigende
Änderung des Zielstands."* beta.24 erschien nach der Planung; sie zu nehmen
wäre genau die stillschweigende Änderung, die der Satz verbietet. Wer sie
will, ändert zuerst diesen Vermerk.

### Die Bestandsmatrix — was v2 hier benutzt und was v3 dafür hat

**Sieben Produktivdateien**, gezählt über `grep -rln "wailsapp/wails/v2"`:
`main.go` und sechs unter `internal/app/`. Dazu neun Testdateien und zwei
Wächter. **Dass es so wenige sind, ist kein Zufall**, sondern §2: `internal/app`
ist die Bindungsfassade, und alles Wails-Eigene steht dort.

| Was v2 hier tut | Wo | v3 beta.23 |
| --- | --- | --- |
| `wails.Run(&options.App{…})` | `main.go` | `application.New(Options)` + `app.Window.NewWithOptions` + `app.Run()` |
| `Bind: []any{…}` | `main.go` | `Services: []Service{application.NewService(…)}` |
| `AssetServer{Assets, Handler}` | `main.go` | `Assets: AssetOptions{Handler, Middleware}` |
| `DragAndDrop{EnableFileDrop}` | `main.go` | Ereignis `common:WindowFilesDropped` |
| `Mac.OnFileOpen` | `main.go` | Ereignis `common:ApplicationOpenedWithFile` |
| `Mac.About` | `main.go` | Menürolle (siehe unten) |
| `runtime.ClipboardGetText/SetText` | `fensternaht.go` | `app.Clipboard.Text()` / `.SetText()` |
| `runtime.EventsEmit` | `fensternaht.go` | `app.Event` |
| `runtime.{Save,Open}FileDialog` | `dateiwahl.go` | `app.Dialog` |
| `pkg/menu` | `menue.go`, `sprache.go` | `application.NewMenu()`, `MenuManager` |
| `MenuSetApplicationMenu` | `menuefenster.go` | `app.Menu` |
| `OnStartup: anwendung.Bereit` | `main.go` | Dienstlebenszyklus |

**Jede Zeile ist am Quelltext nachgesehen**, nicht aus der Dokumentation
übernommen.

### Und es beantwortet eine Frage, die seit einem Jahr offen steht

[0030](0030-die-menueleiste-ist-handgeschrieben.md) und
#255 halten
fest, dass Wails 2 **keine Menürollen** hergibt: `pkg/menu/menuroles.go` führt
sie auskommentiert, und deshalb stehen „Undo", „Cut" und ihresgleichen hier als
**Literale** in `WailsMenu.m` — auf Englisch, gleich welche Sprache eingestellt
ist. Zwei Wächter halten diese Aussage wach.

**v3 hat sie.** `pkg/application/menuitem.go` führt `Role`, `NewRole`,
`SetRole` und `Menu.AddRole`; `menuitem_selectors_darwin.go` bildet jede Rolle
auf einen `NSMenu`-Selektor ab, und damit übersetzt **macOS** sie. Das ist
kein Nebeneffekt der Migration, sondern einer ihrer Gründe.

## Entscheidung

**Die Anwendung wird auf Wails `v3.0.0-beta.23` umgestellt**, gepinnt auf diese
Fassung, und zwar in einem eigenen Pull Request ohne Designbibliothekswechsel
und ohne Schemaänderung (ENG-02.AK-06).

**Der Rückweg bleibt ein Commit weit entfernt.** Die Umstellung ist ein
zusammenhängender Stand; `git revert` auf ihn gibt die Anwendung auf Wails 2
zurück. Das ist die Zusage aus AK-06, und sie kostet nichts, solange die
Umstellung nicht mit anderem vermischt wird.

## Was das kostet, und es steht hier, nicht in einer Fußnote

**Eine Beta hat keine Fassungszusage.** Zwischen beta.23 und beta.24 darf sich
jede Signatur ändern; dieses Repositorium hat für denselben Fall schon einen
Vermerk ([0114](0114-tailcat-gepinnt.md)), und die Lehre dort gilt hier
genauso: **die Pinnung ist die Zusage, nicht die Bibliothek.**

**Der `wails3`-CLI ist ein zweites Werkzeug.** `tools/wails/` hält heute den
v2-CLI in einem eigenen Modul, weil seine Abhängigkeiten mit golangci-lint
kollidieren (§1). Für v3 gilt dieselbe Trennung und dieselbe Begründung.

**Die erzeugten Bindungen wechseln ihre Gestalt.** v2 legt sie unter
`window.go.app.Bindungen` ab; v3 erzeugt ES-Module unter `frontend/bindings/`
und ruft über `@wailsio/runtime`. **Die Fassade muss deshalb nicht umgebaut
werden**: `Call.ByName('paket.struct.methode', …)` steht in beta.23 bereit
(`internal/runtime/desktop/@wailsio/runtime/src/calls.ts`, Zeile 126), und
`fassade.ts` ruft schon heute über den **Namen**. Der Grund, aus dem sie das
tut — `frontend/wailsjs/` ist Generat und in einem frischen Klon nicht da —
gilt für `frontend/bindings/` unverändert.

**Und ein Aufruf über den Namen hat keine Stelligkeitsprüfung.** Dasselbe gilt
heute schon, und genau dafür gibt es seit F03
(#1747)
`internal/werkzeuge/bindungsstelligkeit_test.go`. Nach der Umstellung ist er
nicht weniger wert, sondern mehr.

**Was diese Entscheidung nicht belegt**: dass die umgestellte Anwendung
funktioniert. Das entscheiden die nativen Pflichtprüfungen aus ENG-02.AK-04 am
laufenden Fenster — und für sie gilt
[0007](0007-kein-entfernter-lauf.md): geprüft ist eine Maschine, nicht eine
Plattform.

## Verworfen

**Auf Wails 2 bleiben und nur auf `v2.16.0` heben.** Das wäre billiger und
ließe die Menüfrage aus #255 offen; die englischen Literale blieben. Es ist
außerdem nicht, was die Iteration beschlossen hat.

**Auf die erste stabile v3 warten.** Ein Datum dafür gibt es nicht. Eine
Entscheidung, die auf ein Ereignis ohne Termin wartet, ist keine.

**Die Umstellung mit Mantine zusammen.** ENG-02.AK-06 verbietet es, und der
Grund steht dort: ein einziger Pull Request, in dem Rahmenwerk **und**
Bedienoberfläche wechseln, hat im Fehlerfall keine Ursache, die sich benennen
ließe.

## Woran man merkt, dass die Entscheidung falsch war

- Eine native Pflichtfunktion aus ENG-02.AK-02 ist unter v3 **nicht** zu
  erreichen. Dann verhindert sie den Cutover — so steht es in AK-06.
- Zwei aufeinanderfolgende Betas brechen den Bau, und das Heben kostet mehr
  als die Menürollen einbringen.
- Die erzeugten Bindungen lassen sich nicht über den Namen rufen. Dann müsste
  `fassade.ts` das Generat importieren, und der Grund, aus dem sie es heute
  nicht tut, wäre verloren.
