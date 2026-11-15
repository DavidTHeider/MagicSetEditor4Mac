package storage_test

// **Der Stil eines Sets** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Er macht einen Abzug selbsttragend**: was auf einem fremden Rechner
// gezeichnet wird, braucht den Stil, mit dem es hier gezeichnet wurde.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerStilEinesSetsKommtZurueck(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser, ctx := storage.NeuerSetLeser(db), probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (9, 1, 'Probe', 'magic', 'm15-mainframe', datetime('now'))`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	stil, fehler := leser.Setstil(ctx, 9)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if stil != "m15-mainframe" {
		t.Fatalf("der Stil ist %q", stil)
	}
}

// **Ein Set, das es nicht gibt, wird gemeldet** — und nicht als leerer Stil.
func TestEinFehlendesSetHatKeinenStil(t *testing.T) {
	t.Parallel()

	_, fehler := storage.NeuerSetLeser(frischeDatenbank(t)).Setstil(
		probekontext(t), 99)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Und ein Set eines fremden Profils ist hier keines.**
func TestEinFremdesSetHatHierKeinenStil(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (9, 1, 'Probe', 'magic', 'm15', datetime('now'))`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	fremd := storage.MitProfil(t.Context(),
		zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier"))

	if _, fehler := leser.Setstil(fremd, 9); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Ohne Profil geht gar nichts, und eine tote Datenbank wird gemeldet.**
func TestDerStilMeldetSeineFehlerwege(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	if _, fehler := leser.Setstil(t.Context(), 9); fehler == nil {
		t.Fatal("ohne Profil kam kein Fehler")
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := leser.Setstil(probekontext(t), 9); fehler == nil {
		t.Fatal("die geschlossene Datenbank kam nicht an")
	}
}
