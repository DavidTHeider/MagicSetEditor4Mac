package service_test

// **Ein Deck kopieren** — DOM-03.AK-02
// (#1698).

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kopieren gibt dem Doppel die Kopie — Zeile und Plätze, wie der Speicher.
func (d *deckdoppel) Kopieren(_ context.Context, quelle int64, name string) (
	domain.Deck, error,
) {
	d.kopiert = append(d.kopiert, name)

	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	original, da := d.decks[quelle]
	if !da {
		return domain.Deck{}, fmt.Errorf("%w: das Deck gibt es nicht", domain.ErrNotFound)
	}

	d.naechste++

	kopie := original
	kopie.ID, kopie.Name, kopie.Fassung = d.naechste, name, 1
	kopie.Plaetze = append([]domain.Deckplatz(nil), original.Plaetze...)
	d.decks[kopie.ID] = kopie

	return kopie, nil
}

func kopierbau(t *testing.T) (*service.Deckbau, *deckdoppel, domain.Deck) {
	t.Helper()

	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{spiel: "magic"}, &formatdoppel{}).
		MitKartennamen(&kartennamendoppel{namen: map[int64]string{7: "Glutschmied"}})

	vorlage, fehler := bau.Anlegen(t.Context(), domain.Deck{Name: "Aschenrat", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("die Vorlage: %v", fehler)
	}

	if _, fehler := bau.PlatzLegen(t.Context(), vorlage.ID, domain.Deckplatz{
		KarteID: 7, Menge: 4, Bereich: domain.BereichHauptdeck,
	}); fehler != nil {
		t.Fatalf("der Platz: %v", fehler)
	}

	return bau, speicher, vorlage
}

// **Die Kopie trägt den geglätteten Namen und die Namen ihrer Plätze** —
// sie kommt so zurück, wie `Deck` sie gäbe.
func TestEineKopieKommtMitNamenZurueck(t *testing.T) {
	bau, _, vorlage := kopierbau(t)

	kopie, fehler := bau.Kopieren(t.Context(), vorlage.ID, "  Aschenrat II  ")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if kopie.ID == vorlage.ID || kopie.Name != "Aschenrat II" {
		t.Errorf("die Kopie ist %+v", kopie)
	}

	if len(kopie.Plaetze) != 1 || kopie.Plaetze[0].Bezeichnung != "Glutschmied" {
		t.Errorf("die Plätze der Kopie sind %+v — ohne ihre Namen", kopie.Plaetze)
	}
}

// **Ohne Namen keine Kopie** — der Name ist das Einzige, was sie vom Original
// unterscheidet, und im Spiel ist er einmalig. Gefragt wird der Speicher
// dann gar nicht.
func TestEineKopieOhneNamenWirdAbgewiesen(t *testing.T) {
	bau, speicher, vorlage := kopierbau(t)

	_, fehler := bau.Kopieren(t.Context(), vorlage.ID, "   ")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "name" || feld.Key != domain.SchluesselPflichtfeld {
		t.Fatalf("Fehler = %v, erwartet einen Pflichtfeldfehler am Namen", fehler)
	}

	if len(speicher.kopiert) != 0 {
		t.Errorf("der Speicher wurde gefragt: %v", speicher.kopiert)
	}
}

// **Was der Speicher ablehnt, geht unverändert hinaus** — ein fehlendes Deck
// bleibt ErrNotFound.
func TestEineKopieEinesFehlendenDecksMeldetDas(t *testing.T) {
	bau, _, _ := kopierbau(t)

	if _, fehler := bau.Kopieren(t.Context(), 99, "Beute"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}
