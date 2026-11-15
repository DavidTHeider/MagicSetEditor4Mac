package service_test

// **Das Druckregister im Arbeitsspeicher** — die Vorrichtung zu
// `druckverweiseinfuhr_test.go` (A2, #1441): nicht großzügiger als der
// Betrieb, deshalb prüft `Sichern` den Druck wie die Ablage.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// druckregisterdoppel ist die Druckablage im Arbeitsspeicher.
type druckregisterdoppel struct {
	drucke   []domain.Druck
	naechste int64

	// fehlerBei nennt den Weg, der scheitern soll — "" für keinen.
	fehlerBei string
}

var errRegister = errors.New("das Register klemmt")

func (r *druckregisterdoppel) NachKennung(_ context.Context, kennung domain.Druckkennung) ([]domain.Druck, error) {
	if r.fehlerBei == "NachKennung" {
		return nil, errRegister
	}

	var treffer []domain.Druck

	for _, druck := range r.drucke {
		for _, eigene := range druck.Kennungen {
			if eigene.System == kennung.System && eigene.Semantik == kennung.Semantik && eigene.Wert == kennung.Wert {
				treffer = append(treffer, druck)

				break
			}
		}
	}

	return treffer, nil
}

func (r *druckregisterdoppel) Kandidaten(_ context.Context, name, setcode, nummer string) ([]domain.Druck, error) {
	if r.fehlerBei == "Kandidaten" {
		return nil, errRegister
	}

	var treffer []domain.Druck

	for _, druck := range r.drucke {
		if strings.EqualFold(druck.Name, name) && strings.EqualFold(druck.Setcode, setcode) &&
			strings.EqualFold(druck.Nummer, nummer) {
			treffer = append(treffer, druck)
		}
	}

	return treffer, nil
}

func (r *druckregisterdoppel) Sichern(_ context.Context, druck *domain.Druck) error {
	if r.fehlerBei == "Sichern" {
		return errRegister
	}

	// **Nicht großzügiger als der Betrieb** (§3): die Ablage prüft den Druck,
	// bevor sie ihn schreibt — und weist eine Scryfall-Kennung ab, die keine
	// UUID ist.
	druck.Normalisieren()

	if befunde := druck.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	r.naechste++
	druck.ID = r.naechste
	r.drucke = append(r.drucke, *druck)

	return nil
}

func (r *druckregisterdoppel) Fuer(_ context.Context, id int64) (*domain.Druck, error) {
	if r.fehlerBei == "Fuer" {
		return nil, errRegister
	}

	for _, druck := range r.drucke {
		if druck.ID == id {
			kopie := druck

			return &kopie, nil
		}
	}

	return nil, domain.ErrNotFound
}

// scryfallDruck legt einen Druck mit Scryfall-Kennung ins Register.
func scryfallDruck(t *testing.T, register *druckregisterdoppel, wert, name, setcode, nummer string) int64 {
	t.Helper()

	druck := &domain.Druck{
		Name: name, Setcode: setcode, Nummer: nummer, Sprache: "en",
		Kennungen: []domain.Druckkennung{{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: wert}},
	}
	if err := register.Sichern(t.Context(), druck); err != nil {
		t.Fatalf("Sichern: %v", err)
	}

	return druck.ID
}
