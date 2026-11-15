// Die Befunde eines Pakets werden zu Befunden der Domäne.
//
// **Eine eigene Datei**, weil `import.go` sonst über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, wie ein Paket in den Bestand kommt;
// hier, wie seine Befunde die Schicht wechseln.

package service

import (
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// ausMsepaket macht aus den Befunden eines Pakets die der Domäne.
func ausMsepaket(befunde []msepaket.Befund) []domain.Befund {
	umgesetzt := make([]domain.Befund, 0, len(befunde))

	for _, befund := range befunde {
		umgesetzt = append(umgesetzt, domain.Befund{
			Datei:   befund.Stelle.Datei,
			Ort:     strconv.Itoa(befund.Stelle.Zeile),
			Meldung: befund.Meldung,
		})
	}

	return umgesetzt
}
