// **Der Schnellfilter an der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416).
package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func filterkarten() []*domain.CardRow {
	return []*domain.CardRow{
		{ID: 1, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Grizzly Bears"}, {Field: "type", Text: "Creature"}}},
		{ID: 2, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Goblin Guide"}, {Field: "type", Text: "Creature"}}},
		{ID: 3, SetID: 1, Values: []domain.Value{{Field: "name", Text: "Island"}, {Field: "type", Text: "Basic Land"}}},
	}
}

func TestDerFilterGibtDieKennungenInDerReihenfolgeDesBestands(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		anfrage string
		soll    []int64
	}{
		"leer heißt alle":    {"", []int64{1, 2, 3}},
		"ein Wort":           {"creature", []int64{1, 2}},
		"mit Feld und Minus": {"type:creature -goblin", []int64{1}},
		"nichts trifft":      {"dragon", []int64{}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := &listendoppel{spiel: "probe", karten: filterkarten()}
			bau := service.NeuerKartenlistenbau(doppel, doppel, doppel, doppel, doppel)

			ist, fehler := bau.Filtern(context.Background(), 1, fall.anfrage)
			if fehler != nil {
				t.Fatalf("filtern: %v", fehler)
			}

			if !slices.Equal(ist, fall.soll) {
				t.Errorf("Filtern(%q) = %v, erwartet %v", fall.anfrage, ist, fall.soll)
			}
		})
	}
}

func TestDerFilterReichtDenLesefehlerDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte")
	doppel := &listendoppel{spiel: "probe", kartenfehler: kaputt}
	bau := service.NeuerKartenlistenbau(doppel, doppel, doppel, doppel, doppel)

	if _, fehler := bau.Filtern(context.Background(), 1, "x"); !errors.Is(fehler, kaputt) {
		t.Errorf("Fehler ist %v, erwartet %v", fehler, kaputt)
	}

	koepfeKaputt := &listendoppel{spiel: "probe", lesefehler: kaputt}
	bau = service.NeuerKartenlistenbau(koepfeKaputt, koepfeKaputt, koepfeKaputt, koepfeKaputt, koepfeKaputt)

	if _, fehler := bau.Filtern(context.Background(), 1, "x"); !errors.Is(fehler, kaputt) {
		t.Errorf("Fehler der Setköpfe ist %v, erwartet %v", fehler, kaputt)
	}
}

func TestDerFilterWeistEinUnbekanntesSetAb(t *testing.T) {
	t.Parallel()

	doppel := &listendoppel{spiel: "probe", karten: filterkarten()}
	bau := service.NeuerKartenlistenbau(doppel, doppel, doppel, doppel, doppel)

	if _, fehler := bau.Filtern(context.Background(), 2, "x"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", fehler)
	}
}
