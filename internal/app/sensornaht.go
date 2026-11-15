// **Die Naht zum Sensor dieses Rechners** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Warum diese vier Zeilen eine eigene Datei bekommen
//
// **Weil `.covignore` Dateien ausnimmt und keine Zeilen** (§3). Was hier
// steht, reicht durch: `Bestaetige` **wartet auf einen Finger**, und ein
// Lauf, der es riefe, bliebe stehen, bis jemand den Sensor berührt.
//
// **Alles, was entscheidet, liegt daneben und wird gemessen**:
// `faktorstart.go` legt Geheimnisse ab, `internal/biometrie` unterscheidet
// die Wege über ein Doppel, und `internal/service` entscheidet, wann gefragt
// wird.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sensor fragt die Biometrie dieses Rechners.
//
// **Ein Typ ohne Feld**: `internal/biometrie` hält keinen Zustand, und ein
// Zeiger darauf wäre eine Naht, die niemand austauscht — die liegt eine
// Schicht tiefer, in `biometrie_darwin.go`.
type sensor struct{}

// Vorhanden sagt, ob eine Abfrage möglich wäre — **ohne etwas anzustoßen**.
func (sensor) Vorhanden(ctx context.Context) bool {
	return biometrie.Stand(ctx).Vorhanden
}

// Bestaetige zeigt den Dialog und wartet auf den Sensor.
func (sensor) Bestaetige(ctx context.Context, grund string) error {
	return biometrie.Bestaetige(ctx, grund)
}

// Die Naht, die dieser Verbund erfüllt — geprüft beim Übersetzen.
var _ service.Sensor = sensor{}

// rechnersensor ist der Sensor, den eine Wurzel ohne eingesetzten nimmt.
var rechnersensor service.Sensor = sensor{}
