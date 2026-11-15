package domain_test

// **Ein Kartenbild über die Leitung** — P9
// (#2377),
// Vermerk 0198. Wer ein Bild nicht selbst zeichnen kann, fragt den, dem das
// Deck gehört — nach einem **Objekt** des Tisches, nie nach einer Zeile.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestBildwunschUndBildSindKontrollarten(t *testing.T) {
	t.Parallel()

	for _, art := range []domain.Kontrollart{domain.KontrolleBildwunsch, domain.KontrolleBild} {
		if fehler := art.Pruefe(); fehler != nil {
			t.Errorf("%q: %v", art, fehler)
		}
	}
}

func TestEinBildwunschNenntObjektSeiteUndGuete(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wunsch domain.Bildwunsch
		grund  string
	}{
		"vollständig":   {domain.Bildwunsch{Objekt: "o1", Seite: "vorne", Guete: "tisch"}, ""},
		"ohne Objekt":   {domain.Bildwunsch{Seite: "vorne", Guete: "tisch"}, "Objekt"},
		"ohne Seite":    {domain.Bildwunsch{Objekt: "o1", Guete: "tisch"}, "Seite"},
		"ohne Güte":     {domain.Bildwunsch{Objekt: "o1", Seite: "vorne"}, "Güte"},
		"überlange Art": {domain.Bildwunsch{Objekt: strings.Repeat("o", 200), Seite: "vorne", Guete: "tisch"}, "Objekt"},
		// **Die Bildkennung ist freiwillig** (R2) — ein Gast vor dem Nachlauf
		// schickt sie nicht —, aber nicht beliebig lang.
		"mit Bildkennung":       {domain.Bildwunsch{Objekt: "o1", Bild: "b1", Seite: "vorne", Guete: "tisch"}, ""},
		"überlange Bildkennung": {domain.Bildwunsch{Objekt: "o1", Bild: strings.Repeat("b", 200), Seite: "vorne", Guete: "tisch"}, "Bildkennung"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.wunsch.Pruefe()
			if fall.grund == "" {
				if fehler != nil {
					t.Fatalf("abgewiesen: %v", fehler)
				}

				return
			}

			if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), fall.grund) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Eine Antwort trägt ein Bild oder einen Grund** — nie beides, nie keins.
func TestEineBildantwortTraegtBildOderGrund(t *testing.T) {
	t.Parallel()

	wunsch := domain.Bildwunsch{Objekt: "o1", Seite: "vorne", Guete: "tisch"}

	faelle := map[string]struct {
		antwort domain.Bildantwort
		gilt    bool
	}{
		"ein Bild":        {domain.Bildantwort{Wunsch: wunsch, Png: []byte{1}, Breite: 2, Hoehe: 3}, true},
		"ein Grund":       {domain.Bildantwort{Wunsch: wunsch, Grund: "nicht zu sehen"}, true},
		"beides":          {domain.Bildantwort{Wunsch: wunsch, Png: []byte{1}, Breite: 2, Hoehe: 3, Grund: "x"}, false},
		"keins":           {domain.Bildantwort{Wunsch: wunsch}, false},
		"ohne Maße":       {domain.Bildantwort{Wunsch: wunsch, Png: []byte{1}}, false},
		"falscher Wunsch": {domain.Bildantwort{Wunsch: domain.Bildwunsch{}, Grund: "x"}, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := fall.antwort.Pruefe(); (fehler == nil) != fall.gilt {
				t.Errorf("gilt %v, Prüfung sagt %v", fall.gilt, fehler)
			}
		})
	}
}
