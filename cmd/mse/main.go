// `mse` — die Kommandozeile zu demselben Bestand
// (#886).
//
// # Warum es sie gibt
//
// Das Fenster kann 37 Fähigkeiten, der MCP-Server neun, die Kommandozeile bis
// heute keine (`docs/FAEHIGKEITEN.md`). Wer diese Anwendung in einem Skript
// benutzen will, hat keinen Weg — und die vierzehn Bindungen, die einen
// nativen Dateidialog öffnen, hätte er auch dann nicht, wenn jemand alle
// Bindungen freilegte.
//
// # Was `main` tut und was nicht
//
// **Nur greifen und beenden.** Alles Weitere steht in `lauf.go` und wird
// gemessen; `main` selbst greift nach den Strömen und setzt den Rückgabewert,
// und beides lässt sich nicht in einem Test fahren — dieselbe Trennung wie in
// `cmd/mse-mcp`.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func main() {
	// **⌃C hält an, statt zu warten**
	// (#886).
	//
	// Ein Rastern über ein ganzes Set dauert; wer abbricht, will, dass es
	// aufhört — und nicht, dass die Kommandozeile den Lauf zu Ende bringt und
	// **danach** merkt, dass jemand gedrückt hat. Der Kontext reicht bis in
	// den Rasterer, der ihn je Ebene fragt
	// (`internal/render/rasterer.go`).
	//
	// **Und beim zweiten Signal endet der Prozess.** `NotifyContext` gibt die
	// Behandlung nach dem ersten zurück ans Betriebssystem; wer zweimal
	// drückt, weil das Aufräumen hängt, wird nicht festgehalten.
	ctx, aufraeumen := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)

	// **Geschwärzt und mit Kennung, wie Dienst und Werkzeugserver** —
	// Nachlauf über Phase 6.4, N3: hier schrieb der Standardschreiber von Go
	// volle Pfade mit dem Benutzerverzeichnis.
	slog.SetDefault(app.Serverprotokoll(os.Stderr))
	// **Der Ort des Bestands ist derselbe wie im Fenster** — es sei denn,
	// MSE_BESTAND sagt etwas anderes. Der Grund steht an
	// [app.BestandsortFuerFlaechen].
	pfad, fehler := app.BestandsortFuerFlaechen()
	if fehler != nil {
		code := ohneBestandsort(os.Args[1:], os.Stdout, os.Stderr, fehler)
		aufraeumen()
		os.Exit(code)
	}

	code := lauf(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, pfad)

	// **Das Gerätemodell endet mit dem Prozess** — AI-05.AK-07 (#2042):
	// `fm serve` ist ein Kind, das sonst weiterliefe.
	app.GeraetBeenden()

	// **Ausdrücklich aufräumen und nicht mit `defer`** — jenes liefe vor
	// `os.Exit` nicht mehr, und die Signalbehandlung bliebe auf uns gerichtet.
	aufraeumen()

	os.Exit(code)
}
