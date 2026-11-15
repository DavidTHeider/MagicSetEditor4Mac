package fenster

// **Der Griff ins laufende Fenster** — allein in dieser Datei, damit die
// Ausnahme in `.covignore` nur ihn deckt (§3). Jede Zeile hier ruft Wails und
// braucht eine laufende Anwendung; was darüber entscheidet, steht daneben und
// wird gemessen.

import (
	"context"
	"io/fs"
	"log/slog"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// Optionen ist, was `main.go` dem Fenster mitgibt.
type Optionen struct {
	// Oberflaeche ist das gebaute Frontend — `frontend/dist`, eingebettet.
	Oberflaeche fs.FS

	// Bildweg liefert Vorschau, Spielbild und Tischbild (#1270) — als
	// `*vorschauweg.Weg` und nicht als `http.Handler`, damit die Hülle kein
	// `net/http` braucht (#54).
	Bildweg *vorschauweg.Weg

	// Protokoll ist das Protokoll der Anwendung (#848).
	Protokoll *slog.Logger

	// Stand sagt, wie das Fenster aufgeht (UX-14.AK-01).
	Stand app.Fensterstand
}

// wailshuelle ist die [huelle] über einer laufenden Wails-Anwendung.
type wailshuelle struct{ wails *application.App }

// melden gibt **eine** Liste weiter — Wails reicht einen einzelnen Wert als
// ihn selbst durch ([nutzlast]).
func (h wailshuelle) melden(name string, daten []any) { h.wails.Event.Emit(name, daten) }

func (h wailshuelle) menueSetzen(leiste *application.Menu) { h.wails.Menu.Set(leiste) }

// teilen zeigt den Dialog am vorderen Fenster — oder am ersten, das es gibt.
func (h wailshuelle) teilen(pfad, text string) error {
	fenster := h.wails.Window.Current()
	if fenster == nil {
		if alle := h.wails.Window.GetAll(); len(alle) > 0 {
			fenster = alle[0]
		}
	}

	return systemTeilen(fenster, pfad, text)
}

// dateiWaehlen öffnet den nativen Dialog — **nativ und nicht selbst gebaut**
// (§6), und zwei Arten, weil macOS zwei kennt: eine Datei wählt man anders als
// einen Ordner.
func (h wailshuelle) dateiWaehlen(auswahl app.Dateiauswahl) (string, error) {
	// **Im Abnahmebau kann die Antwort schon bereitliegen** (`abnahme.go`) —
	// sonst nie: außerhalb gibt `abnahmeDialog` nichts.
	if pfad, bereit := abnahmeDialog(); bereit {
		return pfad, nil
	}

	if auswahl.Speichern {
		dialog := h.wails.Dialog.SaveFile().
			SetMessage(auswahl.Titel).
			SetDirectory(auswahl.Startverzeichnis).
			SetFilename(auswahl.Vorschlag)
		if auswahl.Muster != "" {
			dialog.AddFilter(auswahl.Filter, auswahl.Muster)
		}

		return dialog.PromptForSingleSelection()
	}

	dialog := h.wails.Dialog.OpenFile().
		SetTitle(auswahl.Titel).
		SetDirectory(auswahl.Startverzeichnis).
		CanChooseFiles(!auswahl.Verzeichnis).
		CanChooseDirectories(auswahl.Verzeichnis)
	if !auswahl.Verzeichnis && auswahl.Muster != "" {
		dialog.AddFilter(auswahl.Filter, auswahl.Muster)
	}

	return dialog.PromptForSingleSelection()
}

func (h wailshuelle) ablageLesen() (string, error) {
	if ersatz := abnahmeAblage(); ersatz != nil {
		return ersatz.lesen(), nil
	}

	text, gelesen := h.wails.Clipboard.Text()
	if !gelesen {
		return "", errAblageStumm
	}

	return text, nil
}

func (h wailshuelle) ablageSchreiben(text string) error {
	if ersatz := abnahmeAblage(); ersatz != nil {
		ersatz.schreiben(text)

		return nil
	}

	if !h.wails.Clipboard.SetText(text) {
		return errAblageStumm
	}

	return nil
}

// plattform ist der [umsetzer] des Betriebs — beide Griffe brauchen eine
// laufende Anwendung und stehen deshalb hier.
var plattform = umsetzer{rollen: plattformrollen, binden: anKlick}

// anKlick hängt eine Handlung an einen Eintrag.
func anKlick(befehl *application.MenuItem, handlung func()) {
	befehl.OnClick(gesichert("ein Befehl aus dem Menü", func(*application.Context) { handlung() }))
}

// plattformrollen hängt einen Eintrag mit Rolle an, beschriftet von der
// Anwendung (#255). **Hier und nicht in `menue.go`**: Wails baut manche
// Einträge aus dem Namen der laufenden Anwendung (`NewHideMenuItem` liest
// `globalApplication`), und ohne eine gibt es nur einen Absturz.
func plattformrollen(ziel *application.Menu, rolle app.Menuerolle, beschriftung string) {
	eintrag := application.NewRole(wailsrollen[rolle])
	if eintrag == nil {
		return
	}

	eintrag.SetLabel(beschriftung)
	if ohneKuerzel[rolle] {
		eintrag.RemoveAccelerator()
	}

	ziel.Append(application.NewMenuFromItems(eintrag))
}

// Lauf öffnet das Fenster und kehrt zurück, wenn es geschlossen ist.
//
// # Was hier geschieht, und in welcher Reihenfolge
//
//  1. Die Wails-Anwendung entsteht mit der Fassade als Dienst, der
//     eingebetteten Oberfläche und der Weiche zum Bildweg.
//  2. Die Nähte der Anwendung werden auf das Fenster gesetzt ([verdrahten]).
//  3. Die Menüleiste beim Start ist die deutsche — welche Sprache gilt, weiß
//     erst die Oberfläche, und sie meldet es über `SpracheSetzen`.
//  4. Das Fenster geht auf; beim Schließen merkt es sich seine Größe.
//  5. Steht die Anwendung, bekommt sie ihren Kontext (`Bereit`); eine Datei
//     aus dem Finder geht an `DateiOeffnen`.
func Lauf(anwendung *app.App, optionen Optionen) error {
	wails := application.New(application.Options{
		Name: "MagicSetEditor4Mac",

		// **Der Satz des Über-Fensters** — unter Wails 2 `Mac.About.Message`.
		Description: "GPL-2.0-or-later, ausgenommen Inhalte Dritter.\nSiehe DRITTANBIETER.md.",

		// **Gebunden wird die Fassade, nicht die Anwendung.** Wails bindet
		// jede exportierte Methode; mit `anwendung` stünden `Starten`,
		// `Beenden` und die `Mit…`-Nähte im Fenster.
		Services: []application.Service{
			application.NewService(app.NeueBindungen(anwendung)),
		},

		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(optionen.Oberflaeche),
			Middleware: riegelweiche(rufbuchHuelle(schutzbau(assetweiche(optionen.Bildweg), optionen.Oberflaeche, optionen.Protokoll))),

			// **Jede Anfrage im Protokoll wäre jedes Bild** — das Protokoll
			// schneidet sich bei 4 MiB selbst ab und verlöre die Zeilen, auf
			// die es ankommt.
			DisableLogging: true,
		},

		Logger: optionen.Protokoll,

		// **Was Wails selbst fängt, steht mit Vorgangskennung im Protokoll**
		// — F02 (#2341). Der Handler beendet nicht (`riegel.go`); die eigenen
		// Rückrufe darunter fangen ihre Panik selbst.
		PanicHandler: panikMelden,
		ErrorHandler: fehlerMelden,

		// **Das Fenster zu schließen beendet die Anwendung** — wie unter
		// Wails 2. Diese Anwendung hat ein Fenster; ⌘W schließt das Set.
		Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},

		// **Der Weg vom Finder zur Anwendung** (#891): macOS übergibt eine
		// doppelgeklickte Datei als Apple-Event, nicht in `os.Args`.
		FileAssociations: []string{".mse-set"},
	})

	verdrahten(anwendung, wailshuelle{wails: wails}, plattform)

	fenster := wails.Window.NewWithOptions(fensteroptionen(optionen.Stand))

	// **Die Abnahmeleitung** — nur im Abnahmebau und nur mit `MSE_ABNAHME`
	// (`abnahme.go`); sonst tut der Ruf nichts.
	abnahmeAnschliessen(wails, fenster)

	// **Der Vollbild-Eintrag benennt sich um** (#1994, `vollbild_darwin.go`).
	vollbildAnschliessen(fenster)

	// **Ein Ende des Webprozesses wird gezählt** (OPS02, `webprozess_darwin.go`).
	webprozessAnschliessen(wails.Context(), fenster, anwendung.WebprozessNeuGestartet)

	// **Die Größe wird gemerkt, wenn sie sich ändert** (#2439, `letztegroesse.go`):
	// der Schluss läuft auf dem Hauptfaden und kann sie dort nicht erfragen.
	// Wails ruft Hörer in einer eigenen Goroutine, die Frage geht hier an
	// einen freien Hauptfaden.
	groesse := &letzteGroesse{}

	fenster.OnWindowEvent(events.Common.WindowDidResize, gesichert("das Merken der Größe", func(*application.WindowEvent) {
		groesse.melden(fenster.Size)
	}))

	fenster.OnWindowEvent(events.Common.WindowFilesDropped, gesichert("das Ablegen", func(ereignis *application.WindowEvent) {
		wails.Event.Emit(EreignisAbgelegt, ablageAus(
			ereignis.Context().DroppedFiles(), ereignis.Context().DropTargetDetails()))
	}))

	// **Beim Schließen wird die Größe gemerkt** — für `zuletzt` (UX-14.AK-01).
	// Ein Merken, das das Schließen aufhielte, wäre eine Bequemlichkeit, die
	// im Weg steht: der Haken bricht nichts ab.
	//
	// **Und in der Frist des Schlusses** (BK-02, #2110): das Merken geht über
	// die Profilnaht und den Pool mit einer Verbindung; hält jemand beides,
	// hielt es das Fenster offen. Nach der Frist geht die Größe verloren —
	// ein Bestenfalls-Verlust, und `app.Bestenfalls` sagt es im Protokoll.
	fenster.RegisterHook(events.Common.WindowClosing, gesichert("das Schließen des Fensters", func(*application.WindowEvent) {
		frist, aufhoeren := context.WithTimeout(wails.Context(), app.Schlussfrist)
		defer aufhoeren()

		knapp, knappAufhoeren := context.WithTimeout(frist, app.Bestenfallsfrist)
		breite, hoehe, da := fenstergroesse(knapp, groesse.abschliessen)

		knappAufhoeren()

		if da {
			app.Bestenfalls(frist, "die Fenstergröße", func() {
				anwendung.FenstergroesseMerken(frist, breite, hoehe)
			})
		}
	}))

	// **Und beim Beenden**: ⌘Q geht nicht über `WindowClosing` — Wails
	// schließt das Fenster beim Aufräumen selbst, ohne den Haken, und die
	// Größe ging verloren (Hüllenprobe, 2026-09-26: nach ⌘Q standen
	// `fensterbreite` und `fensterhoehe` leer). Der Kontext der Anwendung ist
	// dann schon abgebrochen; gemerkt wird ohne den Abbruch, mit seinen
	// Werten.
	//
	// **Und hier endet die Anwendung geordnet** (BK-07a, BK-02): Tor zu,
	// Größe merken, eigene Läufe beenden, auf sie warten, Bestand zu — in
	// einer Gesamtfrist. Bis dahin merkte der Haken nur die Größe, ohne
	// Frist, und der Bestand blieb offen, bis der Prozess endete.
	//
	// **Zwei Riegel und nicht einer**: bricht der Schluss mit einer Panik ab,
	// endet `fm serve` trotzdem.
	//
	// **Und jeder Schritt in seinem eigenen Riegel** (Zweitblick über Phase 1
	// von 0.9.0, Befund 3) — die Größe, der Schluss, ein zweiter Schluss nach
	// einer Panik, `fm serve`; siehe [schlussFahren].
	//
	// **`fm serve` endet hier** (AI-05.AK-07, #2042): ⌘Q kehrt nicht aus
	// `Run` zurück, und `App.Beenden` im `defer` von `main` läuft dann nicht
	// — gemessen an der Hüllenprobe `ai05` am 2026-10-04: nach ⌘Q stand ein
	// verwaistes `fm serve`.
	wails.OnShutdown(func() {
		riegeln("der Rahmen des Schlusses", func() {
			frist, aufhoeren := context.WithTimeout(context.WithoutCancel(wails.Context()), app.Schlussfrist)
			defer aufhoeren()

			schlussFahren(frist, groesse.abschliessen, anwendung.Schliessen, anwendung.FenstergroesseMerken, app.GeraetBeenden)
		})
	})

	wails.Event.OnApplicationEvent(events.Common.ApplicationStarted, gesichert("der Start", func(*application.ApplicationEvent) {
		anwendung.Bereit(wails.Context())
		probeabsturz()
		probeSyncpanik()
	}))

	// **Eingepasst wird, wenn das Fenster steht** — beim ersten
	// `WindowRuntimeReady`, nicht beim Start der Anwendung: dann gibt es das
	// NSWindow noch nicht, und `Maximise` merkte sich nur den Startzustand,
	// den Wails schon einmal vergeblich gesetzt hat (Hüllenprobe, 2026-09-26).
	// **Einmal**: ein Neuladen der Seite meldet sich wieder, und wer das
	// Fenster inzwischen verkleinert hat, will es so behalten.
	var eingepasst sync.Once

	fenster.OnWindowEvent(events.Common.WindowRuntimeReady, gesichert("das Einpassen", func(*application.WindowEvent) {
		eingepasst.Do(func() { einpassen(fenster, optionen.Stand) })

		// **Und die Größe, mit der es dasteht** — wer es nie zieht, hat sonst
		// keine Meldung, und ⌘Q merkte nichts.
		groesse.melden(fenster.Size)
	}))

	wails.Event.OnApplicationEvent(events.Common.ApplicationOpenedWithFile,
		gesichert("das Öffnen aus dem Finder", func(ereignis *application.ApplicationEvent) {
			anwendung.DateiOeffnen(ereignis.Context().Filename())
		}))

	return wails.Run()
}

// einpassen klemmt das Fenster auf die Arbeitsfläche seines Bildschirms —
// UX-14.AK-04 (#1686). **Nicht, wenn es maximiert aufgeht**: dann füllt es
// den Bildschirm. Die Entscheidung steht in [passend].
//
// **Maximiert wird hier noch einmal**, wenn das Fenster steht. Wails
// setzt den Startzustand (`zoom:`), bevor das Fenster auf dem Bildschirm ist,
// und macOS ließ es dann bei der eingebauten Größe — gemessen am 2026-09-26
// mit der Hüllenprobe (ENG-02.AK-04): 1.279 × 838 Punkte, zentriert, statt
// der Arbeitsfläche von 1.470 × 841. `Maximise` fragt vorher `isZoomed` und
// kippt ein Fenster, das es schon ist, nicht zurück.
func einpassen(fenster *application.WebviewWindow, stand app.Fensterstand) {
	if stand.Maximiert {
		fenster.Maximise()

		return
	}

	// **Die gemerkte Größe ist ein Rahmen**: `Size` liest den Rahmen samt
	// Titelleiste, das Fenster entsteht aber aus einer Inhaltsgröße. Ohne
	// diese Zeile wuchs es bei jedem Start um die Titelleiste — gemessen mit
	// der Hüllenprobe: 1.100 × 760 geschlossen, 1.099 × 791 wieder offen.
	if stand.Breite > 0 && stand.Hoehe > 0 {
		fenster.SetSize(stand.Breite, stand.Hoehe)
	}

	bildschirm, err := fenster.GetScreen()
	if err != nil || bildschirm == nil {
		return
	}

	breite, hoehe := fenster.Size()
	flaecheBreite, flaecheHoehe := inPunkten(bildschirm.WorkArea.Width,
		bildschirm.WorkArea.Height, bildschirm.ScaleFactor)

	neueBreite, neueHoehe, zuGross := passend(breite, hoehe, flaecheBreite, flaecheHoehe)
	if zuGross {
		fenster.SetSize(neueBreite, neueHoehe)
		fenster.Center()
	}
}
