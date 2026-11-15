package app

// **Die Übergabe an macOS** — UX-13.AK-04
// (#1685).
//
// Eine eigene Datei, weil ihr Ruf einen echten Finder aufmacht: ein Lauf, der
// ihn riefe, risse dem Prüfenden ein Fenster auf, und geprüft wäre, dass
// macOS `open` kann (`.covignore`). **Was entscheidet, steht in
// `dateiansicht.go`** und wird gemessen — dieselbe Gestalt wie
// `browserOeffnen` in `anmeldestart.go`.
//
// # Nur am Fenster
//
// Ein Server, der eine Datei öffnete, öffnete sie auf dem Rechner, auf dem
// **er** läuft — nicht bei dem, der fragt (`docs/FAEHIGKEITEN.md`).

import (
	"context"
	"fmt"
	"os/exec"
)

// dateiAufmachen ruft `open` mit geprüften Argumenten.
func dateiAufmachen(ctx context.Context, argumente ...string) error {
	// #nosec G204 -- das Programm steht fest, die Argumente prüft `service.DateiUebergabePruefen`
	if fehler := exec.CommandContext(ctx, "/usr/bin/open", argumente...).Run(); fehler != nil {
		return fmt.Errorf("open: %w", fehler)
	}

	return nil
}

// DateiImFinderZeigen zeigt eine Datei im Finder.
func (a *App) DateiImFinderZeigen(ctx context.Context, pfad string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.dateiZeigen(ctx, pfad, dateiAufmachen)
}

// DateiMitProgrammOeffnen gibt eine Datei dem Standardprogramm — nicht
// [App.DateiOeffnen], das eine Datei aus dem Finder als Set aufschlägt.
func (a *App) DateiMitProgrammOeffnen(ctx context.Context, pfad string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.dateiOeffnen(ctx, pfad, dateiAufmachen)
}

// DateiImFinderZeigen ist die Bindung.
func (b *Bindungen) DateiImFinderZeigen(ctx context.Context, pfad string) SchreibAntwort {
	return b.app(ctx).DateiImFinderZeigen(b.ruf(ctx), pfad)
}

// DateiMitProgrammOeffnen ist die Bindung.
func (b *Bindungen) DateiMitProgrammOeffnen(ctx context.Context, pfad string) SchreibAntwort {
	return b.app(ctx).DateiMitProgrammOeffnen(b.ruf(ctx), pfad)
}
