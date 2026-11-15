package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// gerechneterWert liest einen Feldwert aus den **gerechneten** Feldern.
//
// Der Name ist der des Kartenfelds mit Leerzeichen; die Umgebung führt ihn mit
// Unterstrichen.
func gerechneterWert(felder []msescript.Eintrag, feld string) (string, bool) {
	gesucht := strings.ReplaceAll(feld, " ", "_")

	for _, eintrag := range felder {
		if eintrag.Name == gesucht {
			return eintrag.Wert.AlsText(), true
		}
	}

	return "", false
}
