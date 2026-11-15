package app

// **Wo die Teilungen liegen** — NET-04.AK-04
// (#1708).
//
// **Unter `Caches`**, neben dem Zwischenspeicher: was dort liegt, darf
// macOS wegräumen, und nichts davon gehört in eine Sicherung. **Nicht neben
// dem Bestand** — eine Kopie, die jemand geteilt hat, ist kein Datum.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// teilaufraeumenErlaubt sagt, ob der Start alte Teilungen wegräumen darf —
// **nie aus einem Testlauf**, erkannt wie beim Umzug des Scryfall-Speichers
// an `os.Args[0]`: ein Lauf ohne umgelenktes Zuhause räumte sonst im echten
// `Caches` dieses Rechners.
var teilaufraeumenErlaubt = !strings.HasSuffix(os.Args[0], ".test")

// Teilverzeichnis ist der Ort der Teilungen.
func Teilverzeichnis() (string, error) {
	zwischen, fehler := zwischenspeicherort()
	if fehler != nil {
		return "", fmt.Errorf("den Ort der Teilungen finden: %w", fehler)
	}

	return filepath.Join(zwischen, "MagicSetEditor4Mac", "teilen"), nil
}

// teilenEinrichten setzt die Teilstelle und räumt weg, was älter ist als
// ein Tag — ein Absturz oder ein vergessenes „Fertig" ließ es liegen.
//
// **Ohne Ort teilt die Anwendung nicht** und sagt es beim ersten Versuch
// (`ohneTeilen`); alles andere geht weiter.
func (a *App) teilenEinrichten(aufraeumen bool) {
	ort, fehler := Teilverzeichnis()
	if fehler != nil {
		return
	}

	stelle := service.NeueTeilstelle(ort, time.Now)

	// **Bestenfalls**: was sich nicht räumen lässt, räumt der nächste Start.
	if aufraeumen {
		_, _ = stelle.Aufraeumen(24 * time.Hour) //nolint:errcheck // der Grund steht darüber
	}

	a.MitTeilablage(stelle)
}
