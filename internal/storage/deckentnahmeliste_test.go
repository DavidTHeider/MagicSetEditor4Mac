package storage_test

// **Die Decks eines anderen Profils — zur Wahl bei der Übernahme** —
// UX-08.AK-04 (#1680).
//
// **Dieselbe Grenze wie beim einzelnen Deck** (`deckherkunft_test.go`): die
// Entnahme liest über sie, die gewöhnliche Liste nicht — und nur die Decks
// **des genannten** Profils, keines von einem dritten.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieEntnahmeNenntDieDecksEinesProfils(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	for _, anlage := range []struct {
		profil int64
		name   string
	}{
		{2, "Drueben"},
		{2, "Auch drueben"},
		{3, "Ganz woanders"},
	} {
		if _, fehler := ablage.Anlegen(imProfil(t, anlage.profil),
			domain.Deck{Name: anlage.name, Spiel: "probe"}); fehler != nil {
			t.Fatalf("anlegen %q: %v", anlage.name, fehler)
		}
	}

	// **Aus Profil 1 heraus gelesen** — die gewöhnliche Liste sieht keines.
	eigene, fehler := ablage.Decks(imProfil(t, 1), "")
	if fehler != nil || len(eigene) != 0 {
		t.Fatalf("die gewöhnliche Liste sah %d Decks (%v)", len(eigene), fehler)
	}

	gelesen, fehler := storage.NeueDeckentnahme(db).DeckeAusProfil(
		imProfil(t, 1), 2)
	if fehler != nil {
		t.Fatalf("DeckeAusProfil: %v", fehler)
	}

	namen := make([]string, 0, len(gelesen))
	for _, deck := range gelesen {
		namen = append(namen, deck.Name)
	}

	if len(namen) != 2 {
		t.Fatalf("gelesen wurden %v", namen)
	}

	for _, name := range namen {
		if name == "Ganz woanders" {
			t.Errorf("ein Deck aus Profil 3 kam mit: %v", namen)
		}
	}
}
