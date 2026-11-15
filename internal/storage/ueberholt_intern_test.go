package storage

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein überholter Stand heißt so** — ENG-05.AK-03
// (#1730).
//
// Die sechs Stellen, an denen ein Schreiben gegen eine alte Fassung null
// Zeilen traf, meldeten einen allgemeinen Konflikt — denselben wie ein
// doppelter Name. Nur hier ist „lade neu" der richtige Rat, und nur hier
// soll das Fenster ihn geben.
func TestEinUeberholterStandHeisstSo(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	for _, vorab := range []string{
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (7, 'Deck', 'magic', 1)`,
		`INSERT INTO formate (id, name, spiel, profil_id) VALUES (7, 'Format', 'magic', 1)`,
		`INSERT INTO sammlung (id, printing, menge, zustand, sprache, profil_id)
		 VALUES (7, 'probe', 1, 'nearMint', 'en', 1)`,
		`INSERT INTO zugriffstoken (id, profil_id, bezeichnung, gehuetet, scope, angelegt)
		 VALUES (7, 1, 'Probe', 'x', 'lesen', '2026-09-23T00:00:00.000Z')`,
	} {
		if _, fehler := db.ExecContext(ctx, vorab); fehler != nil {
			t.Fatalf("%s: %v", vorab, fehler)
		}
	}

	for name, fehler := range map[string]error{
		"Deck":     NeueDeckablage(db).warumNicht(ctx, 7, 1),
		"Format":   NeueFormatablage(db).warumNicht(ctx, 7, 1),
		"Exemplar": NeuerSammlungsspeicher(db).warumNicht(ctx, 7, 1),
		"Nutzer":   NeueNutzerablage(db).nutzerNichtGetroffen(ctx, 1),
		"Profil":   NeueNutzerablage(db).profilNichtGetroffen(ctx, 1),
		"Token":    NeueZugriffstokenablage(db).tokenNichtGetroffen(ctx, 7),
	} {
		if !errors.Is(fehler, domain.ErrUeberholt) {
			t.Errorf("%s: %v ist kein überholter Stand", name, fehler)
		}
	}
}
