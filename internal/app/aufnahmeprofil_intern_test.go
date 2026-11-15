package app

// **Eine Aufnahme wartet nicht auf den Profilwechsel** — Befund Z2 des
// Zweitblicks über Phase 6.7 (NET-Teil).
//
// # Die Verklemmung, die es hier gab
//
// Der Wechsel hält die Umschaltsperre und beendet den Horchposten; der
// wartet auf seine laufenden Aufnahmen; und eine Aufnahme, die sich ihr
// Profil noch einmal über [App.imProfil] holte, wartete auf dieselbe Sperre.
// **Die Aufnahme braucht die Sperre nicht**: ihr Kontext stammt aus dem
// Horchposten, und der trägt das Profil, das ihn geöffnet hat.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineAufnahmeWartetNichtAufDenProfilwechsel(t *testing.T) {
	t.Parallel()

	for _, zweck := range []domain.Zweck{domain.ZweckPaarung, domain.ZweckHandel, domain.ZweckPartie} {
		t.Run(string(zweck), func(t *testing.T) {
			t.Parallel()

			anna := eineEinladungsseite(t)
			ctx := anna.imProfil(riegelkontext(t))

			sitzung := eineAngenommene(t, zweck)
			sitzung.lesefehler = errors.New("aufgelegt")

			// **Der Wechsel steht mitten im Umbau.**
			anna.umschaltung.Lock()
			t.Cleanup(anna.umschaltung.Unlock)

			fertig := make(chan struct{})

			go func() {
				anna.sitzungAufnehmen(ctx, sitzung)
				close(fertig)
			}()

			select {
			case <-fertig:
			case <-time.After(3 * time.Second):
				t.Fatal("die Aufnahme wartet auf die Umschaltsperre")
			}
		})
	}
}
