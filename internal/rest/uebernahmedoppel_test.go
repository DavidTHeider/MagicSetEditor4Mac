package rest_test

// **Die Doppel der Deckübernahme** — Werkstück 12 von Phase 4.12
// (#1351).

import (
	"context"
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) DeckUebernehmen(_ context.Context, _, quellDeck string) (
	rest.Uebernommenes, error,
) {
	return rest.Uebernommenes{
		DeckID: "99", Name: "Nachtlauf", Spiel: "probe", Plaetze: 2,
		Befunde: []string{"Das Format des Decks " + quellDeck +
			" bleibt im anderen Profil."},
	}, nil
}

func (u unwilliger) DeckUebernehmen(context.Context, string, string) (
	rest.Uebernommenes, error,
) {
	return rest.Uebernommenes{}, u.fehler
}

func (z *zaehlender) DeckUebernehmen(
	ctx context.Context, quellProfil, quellDeck string,
) (rest.Uebernommenes, error) {
	z.zaehle()

	return z.williger.DeckUebernehmen(ctx, quellProfil, quellDeck)
}

func (williger) DeckeZurUebernahme(context.Context, string) (any, error) {
	return map[string]any{"decks": []map[string]string{
		{"id": "7", "name": "Nachtlauf"},
	}}, nil
}

func (u unwilliger) DeckeZurUebernahme(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) DeckeZurUebernahme(
	ctx context.Context, quellProfil string,
) (any, error) {
	z.zaehle()

	return z.williger.DeckeZurUebernahme(ctx, quellProfil)
}

// **Ein krummer Rumpf wird abgewiesen.**
//
// **Der Zweig, den der Endpunktlauf nicht betritt**: er schickt gültige
// Rümpfe. Ohne diesen Lauf bliebe `gelesen` unbetreten — und ein Weg, der
// jeden Unsinn annimmt, wäre dort grün (§10).
func TestEinKrummerUebernahmerumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/decks/uebernehmen", rumpf: `{`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400 (%s)", antwort.Code, antwort.Body)
	}
}
