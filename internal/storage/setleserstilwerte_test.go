package storage_test

// **Der Setschreiber liest die setweiten Stiloptionen** — F05
// (#2247):
// was über die Stilablage gesetzt wurde, gibt der Leser des Schreibers
// zurück, in seiner Reihenfolge.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerSetleserGibtDieSetweitenStiloptionen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (9, 1, 'Probe', 'magic', 'm15', datetime('now'))`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	ablage := storage.NeueStilablage(db)
	for _, option := range [][2]string{{"chop top", "7"}, {"shrink name text", "3"}} {
		if fehler := ablage.StiloptionSetzen(ctx, 9, option[0], option[1]); fehler != nil {
			t.Fatalf("setzen: %v", fehler)
		}
	}

	werte, fehler := storage.NeuerSetLeser(db).Stilwerte(ctx, 9)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(werte) != 2 || werte[0].Field != "chop top" || werte[0].Text != "7" || werte[1].Field != "shrink name text" {
		t.Errorf("Stiloptionen %+v", werte)
	}
}
