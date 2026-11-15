package app

// **Ein übergroßes Deck kommt nicht an den Tisch** — #2065, am Weg durch die
// Schichten: was die Domäne abweist, kommt an der Fläche als fachlicher
// Fehler an, und es entsteht keine Partie.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinUebergrossesDeckKommtNichtAnDenTisch(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	ablage := storage.NeueDeckablage(db)

	deck, fehler := ablage.Anlegen(ctx, domain.Deck{Name: "Riese", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("Deck anlegen: %v", fehler)
	}

	if _, fehler := ablage.PlatzSetzen(ctx, deck.ID, domain.Deckplatz{
		Printing: "wald", Menge: 20000, Bereich: domain.BereichHauptdeck,
	}); fehler != nil {
		t.Fatalf("Platz setzen: %v", fehler)
	}

	riese := strconv.FormatInt(deck.ID, 10)
	flaeche := PartieflaecheUeber(db)

	antwort := flaeche.PartieAusWuenschen(ctx, "magic", "anna", []Sitzwunsch{
		{Kennung: "anna", Name: "Anna", DeckID: riese},
		{Kennung: "bert", Name: "Bert", DeckID: riese},
	})
	if antwort.Fehler == nil || antwort.Fehler.Code != CodeValidierung {
		t.Fatalf("die Partie fing an: %+v", antwort.Fehler)
	}

	if liste := flaeche.Partien(ctx); len(liste.Partien) != 0 {
		t.Errorf("es blieb eine Partie zurück: %+v", liste.Partien)
	}

	// **Dass jemand zuhört** (§3): ein gewöhnliches Deck kommt an.
	gewoehnlich := einDeckImBestand(t, db)
	if antwort := flaeche.PartieAusWuenschen(ctx, "magic", "anna", []Sitzwunsch{
		{Kennung: "anna", Name: "Anna", DeckID: gewoehnlich},
		{Kennung: "bert", Name: "Bert", DeckID: gewoehnlich},
	}); antwort.Fehler != nil {
		t.Errorf("ein gewöhnliches Deck kam nicht an: %+v", antwort.Fehler)
	}
}
