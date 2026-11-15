package service_test

// **Die Runde entlässt ihre Gäste** — Z3 des Zweitblicks über Phase 6.12
// (#1905).
//
// Die Freigabe zu beenden schickt niemanden vom Tisch, und so soll es
// bleiben. Ein abgeschaltetes Spielen oder Netz muss es aber: sonst stellt
// ein verbundener Gast weiter Befehle an einen Tisch, den der Wirt nicht
// mehr sieht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieRundeEntlaesstJedenGast(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	bert, bertsLeitung := einPlatz("sitz-2")
	carla, carlasLeitung := einPlatz("sitz-3")

	runde.Setze(bert)
	runde.Setze(carla)

	runde.Entlasse()

	if bertsLeitung.geschlossen != 1 || carlasLeitung.geschlossen != 1 {
		t.Errorf("geschlossen: Bert %d, Carla %d", bertsLeitung.geschlossen, carlasLeitung.geschlossen)
	}
}
