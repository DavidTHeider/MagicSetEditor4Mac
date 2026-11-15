//go:build darwin

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine gescheiterte Übernahme hinterlässt kein Deck — am Fenster** —
// #1820.
//
// Die übrigen drei Flächen hält `internal/flaechen/uebernahmezug_test.go`;
// dieser Lauf geht durch die Composition Root des Fensters (`leutestart.go`),
// denn sie baut die Übernahme eigens, über der Arbeitsverbindung.
func TestEineGescheiterteUebernahmeHinterlaesstAmFensterKeinDeck(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	bob := anwendung.NutzerAnlegen(ctx, "Bob", "")

	if ich.Fehler != nil || bob.Fehler != nil {
		t.Fatalf("anlegen: %+v %+v", ich.Fehler, bob.Fehler)
	}

	bobsProfil := bob.Nutzer.Profile[0].ID

	// **Bob legt ein Deck mit drei Plätzen an und gibt es mir frei.**
	if zu := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); zu.Fehler != nil {
		t.Fatalf("zu Bob: %+v", zu.Fehler)
	}

	deck := anwendung.DeckAnlegen(ctx, "Nachtlauf", "probe", "")
	if deck.Fehler != nil {
		t.Fatalf("Deck: %+v", deck.Fehler)
	}

	for _, druck := range []string{
		"a1b2c3d4-0000-4000-8000-00000000000a",
		"a1b2c3d4-0000-4000-8000-00000000000b",
		"a1b2c3d4-0000-4000-8000-00000000000c",
	} {
		if gelegt := anwendung.DeckplatzLegen(ctx, deck.Deck.ID, "", druck, 1,
			string(domain.BereichHauptdeck)); gelegt.Fehler != nil {
			t.Fatalf("legen: %+v", gelegt.Fehler)
		}
	}

	if gesetzt := anwendung.FreigabeSetzen(ctx, app.FreigabeangabenDTO{
		ProfilID: bobsProfil, NutzerID: ich.NutzerID, Rolle: string(domain.RolleMitarbeiter),
	}); gesetzt.Fehler != nil {
		t.Fatalf("Freigabe: %+v", gesetzt.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	vorher := len(anwendung.Decks(ctx, "").Decks)

	// **Der dritte Platz scheitert** — ein Auslöser, den nur dieser Lauf setzt.
	db, fehler := storage.Oeffnen(ctx, bestand)
	if fehler != nil {
		t.Fatalf("zweite Verbindung: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	if _, fehler := db.ExecContext(ctx,
		`CREATE TRIGGER dritter BEFORE INSERT ON deckplaetze
		 WHEN (SELECT COUNT(*) FROM deckplaetze WHERE deck_id = NEW.deck_id) >= 2
		 BEGIN SELECT RAISE(ABORT, 'der dritte Platz scheitert'); END`); fehler != nil {
		t.Fatalf("Auslöser: %v", fehler)
	}

	if geholt := anwendung.DeckUebernehmen(ctx, bobsProfil, deck.Deck.ID); geholt.Fehler == nil {
		t.Fatal("die Übernahme scheiterte nicht — der Auslöser griff nicht")
	}

	if nachher := len(anwendung.Decks(ctx, "").Decks); nachher != vorher {
		t.Errorf("%d Decks, vorher %d — ein halbes Deck blieb stehen", nachher, vorher)
	}
}
