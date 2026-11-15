package fenster

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// riegeln führt einen eigenen Rückruf der Hülle aus und fängt eine Panik
// darin — F02 (#2341).
//
// # Warum jeder Rückruf seinen eigenen
//
// **Weil Wails' Schutz nicht trägt**, nachgelesen in v3.0.0-beta.27 und
// wieder in v3.0.0-beta.30: ohne `PanicHandler` endet eine gefangene Panik in
// `os.Exit(1)` (`application.go`, `handleFatalError`), und einen Haken auf
// `WindowClosing` ruft es ohne jeden Schutz (`webview_window.go`,
// `HandleWindowEvent`). Was hier gefangen wird, sieht Wails nie.
//
// **Die Zeile trägt eine Vorgangskennung** — der Schwärzer setzt sie aus dem
// Kontext. Der Rückgabewert sagt, ob der Rückruf zu Ende kam.
func riegeln(was string, rueckruf func()) (fertig bool) {
	defer func() {
		if wert := recover(); wert != nil {
			slog.ErrorContext(app.MitNeuemVorgang(context.Background()),
				"ein Rückruf des Fensters brach mit einer Panik ab", "art", was, "panik", fmt.Sprint(wert))
		}
	}()

	// **Die Naht der Hüllenprobe** — nur im Abnahmebau wirft sie hier.
	probepanik(was)

	rueckruf()

	return true
}

// panikMelden ist der `PanicHandler` für das, was Wails selbst fängt — eine
// Panik in einer gebundenen Methode, in Wails' eigenem Code oder im Rückruf
// eines `InvokeSync`.
//
// **Er beendet nicht**: Wails ruft ihn auch für die Panik einer Bindung, die
// es schon in eine Antwort verwandelt hat (`bindings.go`, `Call`), und ein
// Handler, der dann beendete, nähme die Anwendung für einen Fehler in einer
// Methode mit. Geschrieben wird die erste Zeile des Stapels als `stelle`.
//
// **Auch nicht nach einer Panik in einem `InvokeSync`** (#2398). Bis
// v3.0.0-beta.27 rief Wails `wg.Done()` dort erst nach dem Rückruf, und wer in
// `wg.Wait()` stand, wartete für immer — deshalb endete die Anwendung dann.
// Seit v3.0.0-beta.30 steht es im `defer` (`mainthread.go`, nachgelesen am
// 2026-10-20), und der Wartende kommt frei; belegt am nativen Fenster mit
// `syncpanik09`.
func panikMelden(einzelheiten *application.PanicDetails) {
	stelle, _, _ := strings.Cut(strings.TrimSpace(einzelheiten.StackTrace), "\n")

	slog.ErrorContext(app.MitNeuemVorgang(context.Background()), "Wails fing eine Panik",
		"panik", fmt.Sprint(einzelheiten.Error), "stelle", stelle)
}

// fehlerMelden ist der `ErrorHandler`: was Wails als Fehler meldet, steht
// mit Vorgangskennung im Protokoll statt nur in Wails' eigenem Logger.
func fehlerMelden(fehler error) {
	slog.ErrorContext(app.MitNeuemVorgang(context.Background()),
		"Wails meldete einen Fehler", "grund", fehler.Error())
}

// gesichert gibt einen Rückruf mit Riegel zurück — für die Anmeldung bei
// Wails, die einen Rückruf mit einem Wert erwartet.
func gesichert[T any](was string, rueckruf func(T)) func(T) {
	return func(wert T) {
		riegeln(was, func() { rueckruf(wert) })
	}
}
