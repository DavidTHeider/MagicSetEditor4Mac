//go:build abnahme && darwin

package fenster

// **Die Dateidialoge** — was an einem Fenster hängt, und wie die Leitung
// ihn liest, abbricht oder vorher beantwortet. Warum sie ihn nicht
// bestätigen kann, steht bei `vorbereitet`.

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tmc/apple/objc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Blattstand ist, was am Fenster hängt — ein Dateidialog oder etwas anderes.
type Blattstand struct {
	Art        string `json:"art"`
	Nachricht  string `json:"nachricht,omitempty"`
	Ordner     string `json:"ordner,omitempty"`
	Name       string `json:"name,omitempty"`
	Ordnerwahl bool   `json:"ordnerwahl,omitempty"`
}

var errKeinBlatt = errors.New("am Fenster hängt kein Blatt")

// vorbereitet ist die Antwort für den **nächsten** Dateidialog.
//
// # Warum es das braucht
//
// **Bestätigen lässt sich ein Dateidialog von hier aus nicht.** Er läuft in
// einem Dienst von macOS; der Stellvertreter in der Anwendung kennt `cancel:`,
// aber nicht `ok:` (`-[NSSavePanel ok:]: not implemented`), und Tasten aus der
// eigenen Ereignisschlange erreichen den Dienst nicht — beides gemessen am
// 2026-09-26. Die Probe öffnet den echten Dialog deshalb einmal, liest ihn
// und bricht ihn ab; für das Bestätigen legt sie die Antwort vorher hierher,
// und `dateiWaehlen` gibt sie zurück, ohne den Dialog zu zeigen. **Belegt ist
// damit alles vor und nach dem Dialog, nicht die Wahl in ihm.**
var vorbereitet struct {
	sync.Mutex

	bereit bool
	datei  string
}

// ablage ist die Zwischenablage des Abnahmebaus, und leitungOffen sagt, ob
// sie gilt: **nur mit Leitung** — ein Abnahmebau ohne `MSE_ABNAHME` kopiert
// wie jeder andere.
var (
	ablage       ersatzablage
	leitungOffen atomic.Bool
)

// abnahmeAblage gibt die Ersatzablage, solange die Leitung offen ist.
func abnahmeAblage() *ersatzablage {
	if !leitungOffen.Load() {
		return nil
	}

	return &ablage
}

// abnahmeDialog gibt die vorbereitete Antwort genau einmal.
func abnahmeDialog() (string, bool) {
	vorbereitet.Lock()
	defer vorbereitet.Unlock()

	if !vorbereitet.bereit {
		return "", false
	}

	vorbereitet.bereit = false

	return vorbereitet.datei, true
}

// blattAn gibt, was am Fenster hängt — oder einen Dateidialog, der frei
// steht: Wails zeigt ihn als Blatt nur, wenn man ihm ein Fenster gibt, und
// `dateiWaehlen` gibt keines (`wails.go`).
func blattAn(fenster *application.WebviewWindow) objc.ID {
	if blatt := objc.Send[objc.ID](fensterID(fenster), objc.Sel("attachedSheet")); blatt != 0 {
		return blatt
	}

	for _, frei := range objc.NSArrayToSlice(objc.Send[objc.ID](nsApp(), objc.Sel("windows"))) {
		if istArt(frei, "NSSavePanel") && objc.Send[bool](frei, objc.Sel("isVisible")) {
			return frei
		}
	}

	return 0
}

func blattLesen(fenster *application.WebviewWindow) Blattstand {
	blatt := blattAn(fenster)
	if blatt == 0 {
		return Blattstand{}
	}

	stand := Blattstand{Art: "anderes"}

	switch {
	case istArt(blatt, "NSOpenPanel"):
		stand.Art = "oeffnen"
		stand.Ordnerwahl = objc.Send[bool](blatt, objc.Sel("canChooseDirectories")) &&
			!objc.Send[bool](blatt, objc.Sel("canChooseFiles"))
	case istArt(blatt, "NSSavePanel"):
		stand.Art = "sichern"
		stand.Name = text(blatt, "nameFieldStringValue")
	default:
		return stand
	}

	stand.Nachricht = text(blatt, "message")
	stand.Ordner = text(objc.Send[objc.ID](blatt, objc.Sel("directoryURL")), "path")

	return stand
}

func dateiadresse(pfad string) objc.ID {
	return objc.Send[objc.ID](objc.ID(objc.GetClass("NSURL")), objc.Sel("fileURLWithPath:"), objc.String(pfad))
}

// blattBeantworten füllt einen Dateidialog und bestätigt ihn — oder bricht
// ihn ab. **Zwischen Füllen und Bestätigen liegt Bedenkzeit**: der Dialog
// liest ein Verzeichnis nebenläufig, und ein sofortiges `ok:` bestätigte den
// Ordner von vorhin.
func blattBeantworten(fenster *application.WebviewWindow, befehl Abnahmebefehl) error {
	fehler := application.InvokeSyncWithError(func() error {
		blatt := blattAn(fenster)
		switch {
		case blatt == 0:
			return errKeinBlatt
		case befehl.Abbrechen:
			objc.Send[struct{}](blatt, objc.Sel("cancel:"), objc.ID(0))

			return nil
		case befehl.Datei == "":
			return errors.New("weder eine Datei noch Abbrechen")
		case istArt(blatt, "NSOpenPanel"):
			objc.Send[struct{}](blatt, objc.Sel("setDirectoryURL:"), dateiadresse(befehl.Datei))
		case istArt(blatt, "NSSavePanel"):
			objc.Send[struct{}](blatt, objc.Sel("setDirectoryURL:"), dateiadresse(filepath.Dir(befehl.Datei)))
			objc.Send[struct{}](blatt, objc.Sel("setNameFieldStringValue:"), objc.String(filepath.Base(befehl.Datei)))
		default:
			return errors.New("das Blatt ist kein Dateidialog")
		}

		return nil
	})
	if fehler != nil || befehl.Abbrechen {
		return fehler
	}

	time.Sleep(blattBedenkzeit)

	// **Bestätigt wird mit dem Zeilenschalter**, wie ein Mensch es täte: der
	// Dialog läuft in einem Dienst von macOS, und `ok:` kennt der Stellvertreter
	// hier nicht (`-[NSSavePanel ok:]: not implemented`, gemessen am
	// 2026-09-26).
	druck, _ := tasteAus("Enter", nil) //nolint:errcheck // Enter steht in der Tafel

	return application.InvokeSyncWithError(func() error {
		if blattAn(fenster) == 0 {
			return errKeinBlatt
		}

		druecken(fenster, druck)

		return nil
	})
}
