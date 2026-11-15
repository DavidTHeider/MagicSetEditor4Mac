package domain_test

// **Eine Regel ohne Muster greift nie** — Werkstück E von Phase 4.11
// (#1341).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
//	void after_reading(AutoReplace& ar, Version) {
//	  if (ar.match.empty()) ar.enabled = false;
//	}
//
// `src/data/word_list.cpp`, Zeilen 51–55, Fassung f9d9356 vom 2021-07-06.
// MSE **schaltet sie ab**; hier wird sie **abgewiesen** — eine Regel, die
// niemand einschalten kann, ist ein Zustand ohne Ausweg, und `msepaket` nimmt
// sie aus demselben Grund gar nicht erst auf.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineAutoersetzungBrauchtEinMuster(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		muster string
		gut    bool
	}{
		"ein Muster":        {"(W)", true},
		"leer":              {"", false},
		"nur Leerzeichen":   {"   ", false},
		"nur ein Umbruch":   {"\n", false},
		"ein Leerzeichen …": {" (W) ", true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := domain.PruefeAutoersetzung(fall.muster)

			if fall.gut && err != nil {
				t.Errorf("%q wurde abgewiesen: %v", fall.muster, err)
			}

			if !fall.gut && !errors.Is(err, domain.ErrValidation) {
				t.Errorf("%q kam durch: %v", fall.muster, err)
			}
		})
	}
}
