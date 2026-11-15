package main

// **Der eigene Schalter vor dem Start** — Zweitblick über Phase 6.12, Z7
// (#1905).
// Dieselbe Frage wie in `cmd/mse-rest/schalter.go`; zwei Zeilen und kein
// gemeinsamer Ort, weil beide Programme ihr eigenes `main` haben.

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// abgeschaltetVorDemStart sagt, ob der eigene Schalter schon vor dem Start aus
// ist — Zweitblick über Phase 6.12, Z7 —, und sagt es auf der
// Fehlerausgabe. Sonst ginge ein Port auf, und bis zur ersten Anfrage
// stünde eine Marke in der Ansage. **Ohne lesbares Profil gilt alles als
// an**: ein Schalter sichert nichts.
//
// **Und der Aufrufer endet mit Rückgabe 0** (Nachlauf, N4): der LaunchAgent
// startet einen Dienst neu, der mit einem Fehler endet (`KeepAlive` mit
// `SuccessfulExit: false`, `man launchd.plist`). Ein bewusst abgeschalteter
// Dienst liefe sonst immer wieder an — und nach dem Einschalten von selbst.
func abgeschaltetVorDemStart(ctx context.Context, db *sql.DB, flaeche domain.Flaeche, eigener string) bool {
	imProfil, fehler := storage.MitGezeigtemProfil(ctx, db)
	if fehler != nil {
		return false
	}

	abgeschaltet := app.SchalterwacheUeber(db).Tor(flaeche, eigener, func() {})(imProfil, "")
	if abgeschaltet == nil {
		return false
	}

	_, _ = fmt.Fprintln(os.Stderr, abgeschaltet) //nolint:errcheck // die Fehlerausgabe ist der letzte Weg

	return true
}
