package service

// **Die `auto replace:`-Regeln eines Spielpakets** —
// #897.
//
// Eine eigene Datei, weil `stilwahl.go` mit ihr über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, **welcher Stil** gilt,
// hier, **was beim Tippen ersetzt wird**.

import (
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// spielregeln gibt die `auto replace:`-Regeln eines Spielpakets.
//
// **Ohne Fehler**: ein Spielpaket, das nicht daliegt, heißt „keine Regeln" —
// derselbe Umgang wie bei [Stilwaehler.spielfelder], und aus demselben Grund
// (#897).
func (w *Stilwaehler) spielregeln(spiel string) []msepaket.AutoErsetzung {
	name := spiel + "." + string(msepaket.ArtSpiel)

	for _, wurzel := range w.wurzeln {
		geoeffnet, err := msepaket.Oeffne(filepath.Join(wurzel, name))
		if err != nil {
			continue
		}

		regeln := geoeffnet.AutoErsetzungen()
		_ = geoeffnet.Schliessen() //nolint:errcheck // die Regeln sind kopiert

		return regeln
	}

	return nil
}
