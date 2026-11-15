//go:build darwin

package fenster

// **Der Vollbild-Eintrag sagt, was er tut** —
// #1994.
//
// macOS hängt *Vollbild aktivieren* selbst an *Darstellung* und benennt ihn im
// Vollbild in *Vollbild beenden* um — über `validateMenuItem:` des Fensters,
// das AppKit beim Öffnen des Menüs fragt. **Wails schaltet diese Frage ab**
// (`[menu setAutoenablesItems:NO]`, `menu_darwin.go` Zeile 30,
// v3.0.0-beta.23), und der Eintrag hieß im Vollbild weiter „Enter Full
// Screen" — gemessen am 2026-10-01 in der Abnahme-VM; Safari daneben sagte
// „Exit Full Screen".
//
// **Gefragt wird deshalb hier**, nach jedem Wechsel: das Fenster benennt den
// Eintrag selbst, in der Sprache des Systems, wie bei jeder anderen
// Anwendung. Kein eigener Text, also auch keine Übersetzung.
//
// **Ohne Lauf, und das steht in `.covignore`**: es braucht ein echtes
// Fenster. Belegt ist es mit der Hüllenprobe (Drehbuch `vollbild`).

import (
	"github.com/tmc/apple/objc"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// vollbildAnschliessen hängt die Nachführung an beide Wechsel.
func vollbildAnschliessen(fenster *application.WebviewWindow) {
	nachfuehren := func(*application.WindowEvent) {
		application.InvokeAsync(func() { vollbildtitelNachfuehren(fenster) })
	}

	fenster.OnWindowEvent(events.Mac.WindowDidEnterFullScreen, gesichert("der Vollbildeintrag", nachfuehren))
	fenster.OnWindowEvent(events.Mac.WindowDidExitFullScreen, gesichert("der Vollbildeintrag", nachfuehren))
}

// vollbildtitelNachfuehren lässt das Fenster den Eintrag benennen.
func vollbildtitelNachfuehren(fenster *application.WebviewWindow) {
	anwendung := objc.Send[objc.ID](objc.ID(objc.GetClass("NSApplication")), objc.Sel("sharedApplication"))
	eintrag := eintragMitAktion(objc.Send[objc.ID](anwendung, objc.Sel("mainMenu")), objc.Sel("toggleFullScreen:"))

	if eintrag == 0 {
		return
	}

	objc.Send[bool](objc.ID(uintptr(fenster.NativeWindow())), objc.Sel("validateMenuItem:"), eintrag)
}

// eintragMitAktion sucht den Eintrag mit dieser Handlung, auch in Untermenüs.
func eintragMitAktion(menue objc.ID, aktion objc.SEL) objc.ID {
	if menue == 0 {
		return 0
	}

	anzahl := objc.Send[int](menue, objc.Sel("numberOfItems"))
	for stelle := range anzahl {
		eintrag := objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle)

		if objc.Send[objc.SEL](eintrag, objc.Sel("action")) == aktion {
			return eintrag
		}

		if gefunden := eintragMitAktion(objc.Send[objc.ID](eintrag, objc.Sel("submenu")), aktion); gefunden != 0 {
			return gefunden
		}
	}

	return 0
}
