// **Das Zeugnis des Dienstes** — Werkstück N6 von Phase 5.4
// (#1572),
// seit NET-03 (#1707)
// wahlweise aus einer eigenen PKI.
//
// **Welches es wird, entscheidet `internal/app`** ([app.App.Dienstzeugnis]),
// für `mse-rest` und `mse-mcp` an derselben Stelle
// ([0154](../../docs/entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md)).
// Hier stehen nur die Flaggen und der Weg dorthin.

package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zeugnisFuerDenDienst baut das TLS-Werk jenseits der Schleife.
//
// **Ohne Zeugnis kein Port jenseits der Schleife.** Das ist dieselbe Linie
// wie beim Horchposten: ein offener Port ohne Ausweis sagte jedem, dass es
// dieses Programm gibt, und wiese dann jeden ab.
//
// **Das gezeigte Profil gehört dazu** — die Identität und die Einstellungen
// hängen an ihm, und ein Kontext ohne eines endet mit „dieser Aufruf nennt
// kein Profil" (gemessen am 2026-09-15 beim ersten Start auf `0.0.0.0`).
// **Einmal beim Start**: das Zeugnis gehört dem Lauscher; getauscht werden
// danach nur die Dateien einer eigenen PKI, am Handschlag.
func zeugnisFuerDenDienst(
	ctx context.Context, db *sql.DB, adressen []string, flaggen leitung.Zeugnisdateien,
) (*tls.Config, app.Dienstzeugnisangabe, error) {
	imProfil, fehler := storage.MitGezeigtemProfil(ctx, db)
	if fehler != nil {
		return nil, app.Dienstzeugnisangabe{}, fehler
	}

	return app.DienstflaecheUeber(db).Dienstzeugnis(imProfil, adressen, flaggen,
		func(befund error) {
			slog.WarnContext(ctx, "das neue Zeugnis gilt nicht — das bisherige bleibt",
				"befund", befund)
		})
}
