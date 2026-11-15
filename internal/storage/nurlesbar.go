package storage

import (
	"errors"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// nurLesbar macht aus einem verweigerten Zugriff „nur lesbar", wenn sich die
// Datei lesen lässt — F05
// (#2344).
//
// **SQLite unterscheidet es hier nicht**: eine Datei ohne Schreibrecht
// meldet beim Öffnen `SQLITE_CANTOPEN` mit „permission denied", genau wie
// eine ohne jedes Recht — gemessen am 2026-10-14
// (`startlagen_intern_test.go`). Der Rat ist aber ein anderer: dort fehlt
// der Zugang, hier nur das Schreibrecht.
func nurLesbar(quelle string, fehler error) error {
	if !errors.Is(fehler, domain.ErrGesperrt) {
		return fehler
	}

	//nolint:gosec // G304: der Pfad des Bestands, den der Aufrufer öffnet
	// #nosec G304 -- dieselbe Begründung für den eigenständigen gosec
	datei, err := os.Open(quelle)
	if err != nil {
		return fehler
	}

	_ = datei.Close() //nolint:errcheck // nur zum Prüfen geöffnet

	return domain.MitGrund(domain.ErrNurLesend, fehler,
		"die Datei lässt sich lesen, aber nicht schreiben — gib ihr Schreibrecht")
}
