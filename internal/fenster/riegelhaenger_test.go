package fenster

// **Eine Panik in einem synchronen Ruf auf dem Hauptfaden wird gemeldet, und
// die Anwendung läuft weiter** — #2398, Befund 3 des Zweitblicks über Phase 1
// von 0.9.0 (F02).
//
// Wails v3.0.0-beta.27 rief `wg.Done()` in `InvokeSync` erst nach dem
// Rückruf (`mainthread.go`); warf er, fing `handlePanic` die Panik, und wer in
// `wg.Wait()` stand, wartete für immer — deshalb endete die Anwendung dann.
// **Seit v3.0.0-beta.30 steht `wg.Done()` im `defer`** (nachgelesen am
// 2026-10-20, `pkg/application/mainthread.go`): der Wartende kommt frei, und
// es gilt, was F02 für jede andere gefangene Panik sagt — melden, nicht
// beenden. Dass er am nativen Fenster freikommt, belegt `syncpanik09`.

import (
	"errors"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// stapelAusInvokeSync ist ein Stapel, wie `debug.Stack` ihn in `handlePanic`
// schreibt, wenn der Rückruf eines `InvokeSync` warf — nachgelesen in
// `pkg/application/mainthread.go` und `panic_handler.go`, v3.0.0-beta.30.
const stapelAusInvokeSync = `goroutine 1 [running]:
runtime/debug.Stack()
github.com/wailsapp/wails/v3/pkg/application.newPanicDetails(...)
github.com/wailsapp/wails/v3/pkg/application.handlePanic({0x0, 0x0, 0x0})
panic({0x1, 0x2})
github.com/wailsapp/wails/v3/pkg/application.InvokeSyncWithResult[...].func1()
github.com/wailsapp/wails/v3/pkg/application.(*macosApp).dispatchOnMainThread.func1()`

func TestEinePanikImSynchronenRufWirdNurGemeldet(t *testing.T) {
	gesammelt := protokolliert(t)

	panikMelden(&application.PanicDetails{Error: errors.New("Probepanik im Hauptfaden"), StackTrace: "x", FullStackTrace: stapelAusInvokeSync})

	zeile := gesammelt.String()
	if !strings.Contains(zeile, "Wails fing eine Panik") || !strings.Contains(zeile, "Probepanik im Hauptfaden") ||
		strings.Count(zeile, "kennung=") != 1 {
		t.Errorf("im Protokoll fehlt die Panik mit Kennung:\n%s", zeile)
	}

	// **Und es endet nichts**: kein Satz über ein Ende, und der Lauf lebt.
	if strings.Contains(zeile, "endet") {
		t.Errorf("das Protokoll sagt ein Ende an:\n%s", zeile)
	}
}
