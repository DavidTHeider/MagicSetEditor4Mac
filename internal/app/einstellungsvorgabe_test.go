package app_test

// **Jede Einstellung nennt ihre Vorgabe** — Werkstück U11 von Phase 5.1
// (#1422): das Fenster setzt darauf zurück, statt die Vorgabe ein zweites
// Mal zu führen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeEinstellungNenntIhreVorgabe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischeAppdatenbank(t)

	alle, err := app.EinstellungsnahtUeber(db).Alle(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(alle) == 0 {
		t.Fatal("keine Einstellungen")
	}

	for _, eintrag := range alle {
		if eintrag.Vorgabe != domain.EinstellungsVorgabe(eintrag.Schluessel) {
			t.Errorf("%s nennt die Vorgabe %q, domain sagt %q",
				eintrag.Schluessel, eintrag.Vorgabe, domain.EinstellungsVorgabe(eintrag.Schluessel))
		}
	}
}
