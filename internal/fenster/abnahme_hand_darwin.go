//go:build abnahme && darwin

package fenster

// **Maus, Tasten und Ablegen** — was ein Mensch mit der Hand ins Fenster
// gibt, über die Ereignisschlange von AppKit.

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ebitengine/purego"
	"github.com/tmc/apple/objc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// zug ist, was ein gerade gespieltes Ablegen trägt — die Tafel mit den
// Pfaden und die Stelle im Fenster. Es gibt immer nur eines.
var zug struct {
	tafel  objc.ID
	stelle punkt
	ziel   objc.ID
	info   objc.ID
}

// ziehklasse ist die Klasse, die AppKit als `NSDraggingInfo` sieht.
var ziehklasse = sync.OnceValues(func() (objc.Class, error) {
	return objc.RegisterClass("MSEAbnahmeZiehen", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
		{Cmd: objc.Sel("draggingPasteboard"), Fn: func(objc.ID, objc.SEL) objc.ID { return zug.tafel }},
		{Cmd: objc.Sel("draggingLocation"), Fn: func(objc.ID, objc.SEL) punkt { return zug.stelle }},
		{Cmd: objc.Sel("draggingSourceOperationMask"), Fn: func(objc.ID, objc.SEL) uint64 { return 1 }},
		{Cmd: objc.Sel("draggingSequenceNumber"), Fn: func(objc.ID, objc.SEL) int { return 1 }},
		{Cmd: objc.Sel("draggingSource"), Fn: func(objc.ID, objc.SEL) objc.ID { return 0 }},
	})
})

// ablegen spielt ein Ablegen aus dem Finder nach: Dateien an einer Stelle der
// Seite. **Nicht der Fensterserver zieht**, sondern die Probe ruft die vier
// Schritte, die AppKit bei einem Ziehen ruft, an der Sicht, die Wails dafür
// anlegt (`WebviewDrag`, `webview_window_darwin_drag.m`, v3.0.0-beta.23) —
// mit einer Tafel unter `NSFilenamesPboardType`, wie der Finder sie füllt.
// Belegt ist damit der Weg ab AppKit: Wails' Ablegen, das Ereignis, die
// Anwendung.
func ablegen(fenster *application.WebviewWindow, pfade []string, x, y float64) error {
	ziel, info, fehler := ueberziehen(fenster, pfade, x, y)
	if fehler != nil {
		return fehler
	}

	if !objc.Send[bool](ziel, objc.Sel("prepareForDragOperation:"), info) ||
		!objc.Send[bool](ziel, objc.Sel("performDragOperation:"), info) {
		return errors.New("die Sicht nahm das Ablegen nicht an")
	}

	return nil
}

// ueberziehen spielt die ersten beiden Schritte: die Dateien treten ein und
// stehen über der Stelle — **ohne Loslassen**. AP-4 (#2361): was das Fenster
// vor dem Loslassen sagt, lässt sich nur so am Fenster prüfen. Das Ziel und
// die Auskunft bleiben im [zug], damit [wegziehen] den Zug beenden kann.
func ueberziehen(fenster *application.WebviewWindow, pfade []string, x, y float64) (ziel, info objc.ID, fehler error) {
	klasse, fehler := ziehklasse()
	if fehler != nil {
		return 0, 0, fmt.Errorf("die Klasse zum Ziehen: %w", fehler)
	}

	ziel = sichtDerArt(objc.Send[objc.ID](fensterID(fenster), objc.Sel("contentView")), "WebviewDrag")
	sicht := webansicht(fenster)

	if ziel == 0 || sicht == 0 {
		return 0, 0, errors.New("das Fenster trägt keine Sicht zum Ablegen")
	}

	namen := make([]objc.ID, 0, len(pfade))
	for _, pfad := range pfade {
		namen = append(namen, objc.String(pfad))
	}

	liste := objc.Send[objc.ID](objc.ID(objc.GetClass("NSMutableArray")), objc.Sel("array"))
	for _, name := range namen {
		objc.Send[struct{}](liste, objc.Sel("addObject:"), name)
	}

	art := objc.String("NSFilenamesPboardType")
	arten := objc.Send[objc.ID](objc.ID(objc.GetClass("NSArray")), objc.Sel("arrayWithObject:"), art)
	zug.tafel = objc.Send[objc.ID](objc.ID(objc.GetClass("NSPasteboard")), objc.Sel("pasteboardWithUniqueName"))
	objc.Send[int](zug.tafel, objc.Sel("declareTypes:owner:"), arten, objc.ID(0))
	objc.Send[bool](zug.tafel, objc.Sel("setPropertyList:forType:"), liste, art)
	zug.stelle = objc.Send[punkt](sicht, objc.Sel("convertPoint:toView:"), punkt{X: x, Y: y}, objc.ID(0))

	info = objc.Send[objc.ID](objc.Send[objc.ID](objc.ID(klasse), objc.Sel("alloc")), objc.Sel("init"))
	objc.Send[uint64](ziel, objc.Sel("draggingEntered:"), info)
	objc.Send[uint64](ziel, objc.Sel("draggingUpdated:"), info)
	zug.ziel, zug.info = ziel, info

	return ziel, info, nil
}

// wegziehen beendet einen Zug, der nicht abgelegt wurde — die Dateien
// verlassen das Fenster.
func wegziehen() error {
	if zug.ziel == 0 {
		return errors.New("es wird nichts gezogen")
	}

	objc.Send[struct{}](zug.ziel, objc.Sel("draggingExited:"), zug.info)
	zug.ziel, zug.info = 0, 0

	return nil
}

// sichtDerArt sucht eine Sicht einer Klasse unter einer anderen.
func sichtDerArt(sicht objc.ID, klasse string) objc.ID {
	if sicht == 0 {
		return 0
	}

	if istArt(sicht, klasse) {
		return sicht
	}

	for _, kind := range objc.NSArrayToSlice(objc.Send[objc.ID](sicht, objc.Sel("subviews"))) {
		if gefunden := sichtDerArt(kind, klasse); gefunden != 0 {
			return gefunden
		}
	}

	return 0
}

// druecken legt einen Tastendruck in die Ereignisschlange der Anwendung —
// an das Fenster, oder an sein Blatt, wenn eines offen ist. Von
// dort nimmt AppKit denselben Weg wie bei einer Taste: Kürzel der Leiste,
// dann die Webansicht.
func druecken(fenster *application.WebviewWindow, druck Tastendruck) {
	// **An das eigene Fenster, nicht an das Schlüsselfenster der Anwendung**:
	// ist sie vorn, kann das ein anderes sein (gemessen am 2026-09-26 — dann
	// kam keine Taste an), und ist sie es nicht, gibt es keines.
	ziel := blattAn(fenster)
	if ziel == 0 {
		ziel = fensterID(fenster)
	}

	nummer := objc.Send[int](ziel, objc.Sel("windowNumber"))
	zeit := objc.Send[float64](objc.Send[objc.ID](objc.ID(objc.GetClass("NSProcessInfo")),
		objc.Sel("processInfo")), objc.Sel("systemUptime"))

	for _, art := range []uint64{artTasteRunter, artTasteHoch} {
		ereignis := objc.Send[objc.ID](objc.ID(objc.GetClass("NSEvent")),
			objc.Sel("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"),
			art, punkt{}, druck.Schalter, zeit, nummer, objc.ID(0),
			objc.String(druck.Zeichen), objc.String(druck.Ohne), stapelwort(false, druck.Code))
		objc.Send[struct{}](nsApp(), objc.Sel("postEvent:atStart:"), ereignis, false)
	}
}

// webansicht sucht die `WKWebView` im Inhalt des Fensters.
func webansicht(fenster *application.WebviewWindow) objc.ID {
	var suche func(sicht objc.ID) objc.ID
	suche = func(sicht objc.ID) objc.ID {
		if istArt(sicht, "WKWebView") {
			return sicht
		}

		for _, kind := range objc.NSArrayToSlice(objc.Send[objc.ID](sicht, objc.Sel("subviews"))) {
			if gefunden := suche(kind); gefunden != 0 {
				return gefunden
			}
		}

		return 0
	}

	return suche(objc.Send[objc.ID](fensterID(fenster), objc.Sel("contentView")))
}

// klickeAn legt Maus runter und hoch an eine Stelle der Seite — **in die
// Ereignisschlange**, also mit Trefferprüfung, Fokuswechsel und allem, was
// WebKit bei einem Klick tut. Die Stelle rechnet die Webansicht selbst ins
// Fenster um (`convertPoint:toView:`), denn sie ist gespiegelt und das
// Fenster nicht. **Alle Schritte in einem Zug auf dem Hauptfaden**: die
// Klicks eines Doppelklicks liegen so dicht beieinander wie bei einer Hand,
// und welcher der wievielte ist, sagt die Klickzahl ([klickschritte]).
func klickeAn(fenster *application.WebviewWindow, schritte []mausschritt) error {
	for _, schritt := range schritte {
		if fehler := mausereignis(fenster, schritt); fehler != nil {
			return fehler
		}
	}

	return nil
}

// Zugpause ist die Zeit zwischen zwei Schritten eines Zugs — A04 (#2218).
//
// **Eine Hand lässt sie, und die Seite braucht sie**: das Fassen setzt einen
// Zustand, den React erst mit dem nächsten Bild annimmt; ein Schieben im
// selben Augenblick fände nichts gefasst. 40 ms sind etwas mehr als zwei
// Bilder bei 60 Hz.
const Zugpause = 40 * time.Millisecond

// mausSchicken legt die Schritte eines Zugs nacheinander in die
// Ereignisschlange — **jeden auf dem Hauptfaden, die Pause dazwischen nicht**:
// schliefe der Hauptfaden, verarbeitete AppKit in der Zeit nichts.
func mausSchicken(fenster *application.WebviewWindow, schritte []mausschritt) error {
	for stelle, schritt := range schritte {
		if stelle > 0 {
			time.Sleep(Zugpause)
		}

		if fehler := application.InvokeSyncWithError(func() error {
			return mausereignis(fenster, schritt)
		}); fehler != nil {
			return fehler
		}
	}

	return nil
}

// mausereignis legt ein Mausereignis an eine Stelle der Seite.
func mausereignis(fenster *application.WebviewWindow, schritt mausschritt) error {
	sicht := webansicht(fenster)
	if sicht == 0 {
		return errors.New("das Fenster trägt keine Webansicht")
	}

	ersterKlickZaehlt(sicht)

	stelle := objc.Send[punkt](sicht, objc.Sel("convertPoint:toView:"),
		punkt{X: schritt.ort.X, Y: schritt.ort.Y}, objc.ID(0))
	nummer := objc.Send[int](fensterID(fenster), objc.Sel("windowNumber"))
	zeit := objc.Send[float64](objc.Send[objc.ID](objc.ID(objc.GetClass("NSProcessInfo")),
		objc.Sel("processInfo")), objc.Sel("systemUptime"))

	ereignis := objc.Send[objc.ID](objc.ID(objc.GetClass("NSEvent")),
		objc.Sel("mouseEventWithType:location:modifierFlags:timestamp:windowNumber:context:eventNumber:clickCount:pressure:"),
		schritt.art, stelle, schritt.schalter, zeit, nummer, objc.ID(0), 0, max(schritt.klicks, 1), float32(1))
	objc.Send[struct{}](nsApp(), objc.Sel("postEvent:atStart:"), ereignis, false)

	return nil
}

// ersterKlick sorgt dafür, dass die Webansicht nur einmal umgestellt wird.
var ersterKlick sync.Once

// ersterKlickZaehlt lässt die Webansicht einen Klick annehmen, auch wenn die
// Anwendung nicht vorn ist (`acceptsFirstMouse:`).
//
// **Warum**: ein Klick in das Fenster einer Anwendung im Hintergrund ist für
// AppKit der Klick, der sie nach vorn holt — und seit macOS 14 holt er sie
// nicht, weil sie aus einem Hintergrundprozess startet. Der Klick ging dann
// verloren: gemessen am 2026-09-26, dieselben Knöpfe nahmen ihn an, wenn die
// Anwendung vorn war, und nicht, wenn nicht. Dieselbe Wendung wie in der
// WebKit-Probe (`scripts/webkitprobe/Motoren.swift`).
//
// **Die Methode der Klasse wird ersetzt, nicht die Klasse des Objekts.** Bis
// zum 2026-10-07 tauschte `object_setClass` die Klasse dieser einen
// Webansicht — nachdem AppKit sie schon beobachtete (KVO auf `frame`). Die
// nächste Größenänderung räumte den Beobachter ab und fand ihn unter der
// neuen Klasse nicht: `NSRangeException` aus `updateTitlebarAdjacencyState`,
// und die Anwendung war fort (gemessen mit der Hüllenprobe `p01groesse`; vor
// dem ersten Klick ging dieselbe Größenänderung). `class_replaceMethod` an
// der Klasse, die die Webansicht nennt, lässt die Beobachtung, wo sie ist —
// und trifft im Abnahmebau jede Webansicht, was hier gewollt ist.
func ersterKlickZaehlt(sicht objc.ID) {
	ersterKlick.Do(func() {
		ersetzen, fehler := methodenersatz()
		if fehler != nil {
			slog.Error("die Webansicht der Abnahme ließ sich nicht umstellen", "grund", fehler.Error())

			return
		}

		ersetzen(objc.Send[objc.Class](sicht, objc.Sel("class")), objc.Sel("acceptsFirstMouse:"),
			objc.NewIMP(func(objc.ID, objc.SEL, objc.ID) bool { return true }), "B@:@")
	})
}

// methodenersatz bindet `class_replaceMethod` aus der Laufzeit, die schon
// geladen ist — `tmc/apple` reicht sie nur mit einem Funktionstyp ohne
// Argumente weiter, und der taugt nicht als Umsetzung einer Methode.
func methodenersatz() (func(objc.Class, objc.SEL, objc.IMP, string) objc.IMP, error) {
	stelle, fehler := purego.Dlsym(purego.RTLD_DEFAULT, "class_replaceMethod")
	if fehler != nil {
		return nil, fmt.Errorf("class_replaceMethod: %w", fehler)
	}

	var ersetzen func(objc.Class, objc.SEL, objc.IMP, string) objc.IMP
	purego.RegisterFunc(&ersetzen, stelle)

	return ersetzen, nil
}

// aktivieren holt die Anwendung nach vorn — ein Mensch täte es mit einem
// Klick ins Fenster.
func aktivieren(fenster *application.WebviewWindow) {
	objc.Send[struct{}](nsApp(), objc.Sel("activateIgnoringOtherApps:"), true)
	objc.Send[struct{}](fensterID(fenster), objc.Sel("makeKeyAndOrderFront:"), objc.ID(0))
}
