package fenster

// **Die zuletzt gemeldete Größe des Fensters** — #2439, eine Regression aus
// F02 (#2341).
//
// # Warum der Schluss die Größe nicht mehr erfragt
//
// `WebviewWindow.Size()` geht über `InvokeSync`, und Wails ruft dort
// **direkt**, wenn es schon auf dem Hauptfaden ist, sonst reiht es ein
// (`App.dispatchOnMainThread`, `pkg/application/application.go:981`,
// v3.0.0-beta.27). Der Rückruf von `OnShutdown` läuft auf dem Hauptfaden;
// seit F02 fragte [fenstergroesse] in einer eigenen Goroutine und wartete auf
// sie. Die eingereihte Frage kam nie dran, die Frist lief ab, und nach ⌘Q
// standen `fensterbreite` und `fensterhoehe` leer — still, ohne Zeile im
// Protokoll (Hüllenprobe `start`, 2026-10-16).
//
// Direkt zu fragen, wie vor F02, brächte den Hänger zurück, gegen den F02
// steht: `InvokeSync` kehrt nach einer inneren Panik nie zurück, und auf dem
// Hauptfaden hinge dann das Beenden.
//
// **Deshalb merkt sich die Hülle die Größe, wenn sie sich ändert** —
// `WindowDidResize` kommt in einer eigenen Goroutine (Wails ruft die Hörer
// mit `go`, `webview_window.go`, `HandleWindowEvent`), die Frage dort geht
// an einen freien Hauptfaden. Der Schluss liest nur noch, was dasteht.

// **Und die Meldungen sind geordnet** — Nachlauf über Phase 5 von 0.9.0,
// Befund 2. Jede Meldung zieht ihre Nummer, **bevor** sie fragt; gemerkt wird
// nur, was jünger ist als das Gemerkte. Zwei Hörer, deren Antworten sich
// überholen, hinterlassen so die Größe, die zuletzt **kam**. Und der Schluss
// liest mit [letzteGroesse.abschliessen]: danach wird nichts mehr gemerkt und
// nicht mehr gefragt.
//
// **Was das nicht abfängt**: eine Frage, die ihre Nummer vor dem Schluss zog
// und noch beim Hauptfaden wartet. Wails prüft in `Size()` den Abbau des
// Fensters **vor** dem Einreihen und im eingereihten Ruf nicht noch einmal
// (`webview_window.go`, v3.0.0-beta.27); ihre Antwort wird verworfen, die
// Frage selbst lässt sich von hier nicht zurückholen. Ein Absturz daraus ist
// nicht beobachtet.

import "sync"

// letzteGroesse hält die Größe, die das Fenster zuletzt meldete.
type letzteGroesse struct {
	mu            sync.Mutex
	breite, hoehe int

	// vergeben ist die Nummer der jüngsten Meldung, uebernommen die der
	// gemerkten.
	vergeben, uebernommen uint64

	abgeschlossen bool
}

// melden fragt die Größe und merkt sie — **nach dem Schluss fragt es nicht
// mehr**.
func (g *letzteGroesse) melden(frage func() (int, int)) {
	nummer, offen := g.nummer()
	if !offen {
		return
	}

	breite, hoehe := frage()
	g.setzen(nummer, breite, hoehe)
}

// nummer vergibt die Nummer einer Meldung — und sagt, ob noch gefragt werden
// darf.
func (g *letzteGroesse) nummer() (uint64, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.abgeschlossen {
		return 0, false
	}

	g.vergeben++

	return g.vergeben, true
}

// setzen merkt die Größe der Meldung mit dieser Nummer — eine leere nicht:
// ein Fenster, das schon fort ist, meldet 0 × 0, und das wäre keine Größe,
// die jemand wiederhaben will. **Und keine, die älter ist als die gemerkte,
// und keine nach dem Schluss.**
func (g *letzteGroesse) setzen(nummer uint64, breite, hoehe int) {
	if breite <= 0 || hoehe <= 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.abgeschlossen || nummer <= g.uebernommen {
		return
	}

	g.uebernommen = nummer
	g.breite, g.hoehe = breite, hoehe
}

// abschliessen gibt die zuletzt gemerkte Größe — 0 × 0, wenn noch keine kam —
// und nimmt danach keine mehr an. Ein zweiter Ruf gibt dieselbe.
func (g *letzteGroesse) abschliessen() (breite, hoehe int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.abgeschlossen = true

	return g.breite, g.hoehe
}
