package domain

// **Ein eingetragener Lauf macht den Anbieter „real geprüft"** — die Probe
// darauf, dass der Eintrag in `anbieterlaeufe` wirkt, bevor es einen gibt.

import "testing"

func TestEinEingetragenerLaufIstRealGeprueft(t *testing.T) {
	vorher := anbieterlaeufe
	anbieterlaeufe = []string{"github"}

	t.Cleanup(func() { anbieterlaeufe = vorher })

	if stand := AnbieterPruefstand("github"); stand != PruefstandReal {
		t.Errorf("github: %q", stand)
	}

	if stand := AnbieterPruefstand("google"); stand != PruefstandSimuliert {
		t.Errorf("google: %q", stand)
	}
}
