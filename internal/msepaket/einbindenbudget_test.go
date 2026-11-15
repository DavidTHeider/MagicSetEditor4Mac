package msepaket_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Einbinden hat ein Budget** — FMT-06.AK-01
// (#1695).
//
// Kreise erkennt der Einbinder seit jeher. Eine **Raute** aber ist kein Kreis:
// eine Datei, die dieselbe andere zehnmal einbindet, sechs Ebenen tief, gibt
// eine Million Auflösungen, acht Ebenen hundert Millionen — das »Billion
// Laughs« des MSE-Formats. Und eine lange Kette verschiedener Dateien wächst
// den Stapel, bis er reißt.
func TestEineRauteImEinbindenBleibtImBudget(t *testing.T) {
	t.Parallel()

	quelle := dateien{}

	for tiefe := 1; tiefe <= 6; tiefe++ {
		quelle[fmt.Sprintf("r.mse-include/ebene%d", tiefe)] = strings.Repeat(fmt.Sprintf("include file: /r.mse-include/ebene%d\n", tiefe+1), 10)
	}

	quelle["r.mse-include/ebene7"] = "blatt: 1\n"

	baum, befunde := einbinden(t, quelle, "include file: /r.mse-include/ebene1\n")

	if blaetter := len(baum.Alle("blatt")); blaetter > msepaket.GrenzeEinbindungen {
		t.Errorf("%d Blätter eingebunden — mehr als die Grenze von %d Einbindungen", blaetter,
			msepaket.GrenzeEinbindungen)
	}

	if !befundNennt(befunde, "Einbindungen") {
		t.Errorf("kein Befund über die Grenze: %v", befunde)
	}
}

// **Und eine lange Kette endet an der Tiefengrenze**, mit Befund.
func TestEineLangeKetteEndetAnDerTiefengrenze(t *testing.T) {
	t.Parallel()

	quelle := dateien{}
	laenge := msepaket.GrenzeEinbindetiefe + 5

	for glied := 1; glied < laenge; glied++ {
		quelle[fmt.Sprintf("k.mse-include/glied%d", glied)] = fmt.Sprintf("tiefe: %d\ninclude file: /k.mse-include/glied%d\n", glied, glied+1)
	}

	quelle[fmt.Sprintf("k.mse-include/glied%d", laenge)] = "ende: ja\n"

	baum, befunde := einbinden(t, quelle, "include file: /k.mse-include/glied1\n")

	if len(baum.Alle("ende")) != 0 {
		t.Error("das Ende der Kette wurde erreicht — über die Tiefengrenze hinaus")
	}

	if !befundNennt(befunde, "tief") {
		t.Errorf("kein Befund über die Tiefe: %v", befunde)
	}
}

// befundNennt sagt, ob ein Befund das Wort trägt.
func befundNennt(befunde []msepaket.Befund, wort string) bool {
	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, wort) {
			return true
		}
	}

	return false
}
