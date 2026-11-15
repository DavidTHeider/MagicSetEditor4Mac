package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probeschema hat drei Felder, eines davon nicht für die KI.
func probeschema() domain.KISchema {
	return domain.KISchema{Paket: "probe", Fassung: "1", Felder: []domain.KISchemafeld{
		{Name: "name", Typ: domain.KIText, Schreibbar: true},
		{Name: "flavor text", Typ: domain.KIText, Schreibbar: true},
		{Name: "casting cost", Typ: domain.KIText},
	}}
}

// **Gefragt wird nach den genannten Feldern** — AI-05 (#2042): eine kleine
// Aufgabe schickt dem Modell nur Name oder Flavortext, nicht die ganze Karte.
func TestDasSchemaNurFuerGenannteFelder(t *testing.T) {
	t.Parallel()

	eng, fehler := probeschema().Nur([]string{"flavor text"})
	if fehler != nil || len(eng.Felder) != 1 || eng.Felder[0].Name != "flavor text" ||
		eng.Paket != "probe" || eng.Fassung != "1" {
		t.Errorf("%+v %v", eng, fehler)
	}

	if alle, fehler := probeschema().Nur(nil); fehler != nil || len(alle.Felder) != 3 {
		t.Errorf("ohne Namen: %+v %v", alle, fehler)
	}
}

// **Ein Feld, das es nicht gibt oder das die KI nicht setzen darf, ist ein
// Fehler** — und kein leerer Auftrag.
func TestDasSchemaWeistFremdeFelderAb(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"rule text", "casting cost"} {
		if _, fehler := probeschema().Nur([]string{name}); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}
