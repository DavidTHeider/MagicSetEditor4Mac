package storage_test

// **Das Profil im Kontext** — Werkstück 2 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was hineingeht, kommt heraus.**
func TestDasProfilReistImKontext(t *testing.T) {
	t.Parallel()

	gelesen, err := storage.ProfilAus(storage.MitProfil(t.Context(), 7))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen != 7 {
		t.Errorf("das Profil ist %d", gelesen)
	}
}

// **Ein Kontext ohne Profil ist ein Fehler, keine Null.**
//
// Eine Abfrage mit `profil_id = 0` fände nichts, und das sähe aus wie ein
// leerer Bestand statt wie ein Fehler im Aufruf. Genau deshalb ist die
// Alternative — stillschweigend null — die falsche.
func TestEinKontextOhneProfilIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"gar keines", "die Null"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// **Hier ausdrücklich `t.Context()`** und nicht `probekontext` —
			// dieser Lauf prüft gerade den Kontext **ohne** Profil.
			ctx := t.Context()
			if name == "die Null" {
				ctx = storage.MitProfil(ctx, 0)
			}

			if _, err := storage.ProfilAus(ctx); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler ist %v, erwartet ErrValidation", err)
			}
		})
	}
}
