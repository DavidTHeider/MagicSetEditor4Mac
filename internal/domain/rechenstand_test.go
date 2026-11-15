package domain_test

// **Der Stand einer Rechnung trägt Karte und Setangaben** — #2533 (Nachlauf
// über das Gesamtreview über 0.9.0, Befund 4): ein Wertskript liest auch
// `set.*`, und der Kartenstand allein ließ eine Änderung daran durch.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerRechenstandTraegtKarteUndSetangaben(t *testing.T) {
	t.Parallel()

	karte := domain.CardRow{ID: 7, Values: []domain.Value{{Field: "name", Text: "Glut"}}}
	setwerte := []domain.Value{{Field: "copyright", Text: "2026"}}

	stand := domain.Rechenstand(karte.Stand(), setwerte)

	if k, s := domain.StandTeile(stand); k != karte.Stand() || s != domain.Setstand(setwerte) {
		t.Errorf("geteilt in %q / %q", k, s)
	}

	if domain.Setstand([]domain.Value{{Field: "copyright", Text: "2027"}}) == domain.Setstand(setwerte) {
		t.Error("ein anderer Setwert gibt denselben Setstand")
	}

	// **Auch ohne Setangaben steht der Setteil da** — sonst fiele eine erste
	// Angabe, die dazwischen entsteht, nicht auf.
	if _, s := domain.StandTeile(domain.Rechenstand(karte.Stand(), nil)); s == "" {
		t.Error("ohne Setangaben fehlt der Setteil")
	}
}

// **Ein reiner Kartenstand bleibt einer** — so nennt ihn ein KI-Auftrag.
func TestEinKartenstandHatKeinenSetteil(t *testing.T) {
	t.Parallel()

	karte := domain.CardRow{ID: 7}

	if k, s := domain.StandTeile(karte.Stand()); k != karte.Stand() || s != "" {
		t.Errorf("geteilt in %q / %q", k, s)
	}
}
