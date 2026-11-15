package app

// **Die Naht zur Datenpflege** — Werkstück O8 von Phase 5.2
// (#1432).
//
// # Warum sie die Verbindung holt und nicht hält
//
// **Weil ein Zurückspielen sie wechselt** — derselbe Grund wie bei
// `rueckkehrnaht`. Und weil die Pflege **vor** ihrem ersten Schreiben sichert:
// die Sicherung geht durch dieselbe Verbindung, und nach einem Tausch ist es
// eine andere.

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Pflegesicherungsname ist die Datei, die vor einer Pflege entsteht.
//
// **Mit dem Zeitpunkt im Namen**, denn es kann mehrere geben — und eine
// Sicherung überschreibt nichts (#799). Dieselbe Entscheidung wie bei
// [Sicherungsname], nur mit einer anderen Kennzeichnung: dort sagt die
// Fassung, wovor gesichert wurde, hier der Augenblick.
func Pflegesicherungsname(jetzt time.Time) string {
	return "bestand-vor-pflege-" + jetzt.UTC().Format("20060102-150405") + ".db"
}

// PflegeUeber baut die Datenpflege über einer Verbindung.
func PflegeUeber(holen func() *sql.DB, bestandsort string) *service.Datenpflege {
	return service.NeueDatenpflege(
		func(ctx context.Context, jetzt time.Time) ([]domain.Pflegebefund, error) {
			return storage.Pflegeanalyse(ctx, holen(), jetzt)
		},
		func(ctx context.Context, jetzt time.Time, revisionen map[string]string) (
			[]domain.Pflegezeile, error,
		) {
			return storage.PflegeAusfuehren(ctx, holen(), jetzt, revisionen)
		},
		func(ctx context.Context) (string, error) {
			ziel := filepath.Join(filepath.Dir(bestandsort), Pflegesicherungsname(time.Now()))

			return ziel, storage.Sichern(ctx, holen(), ziel)
		},
	)
}
