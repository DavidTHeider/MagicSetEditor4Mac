package service

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Abbruch gibt kein halbes Ergebnis** (#2006). Der Test über [Fuehre]
// sieht nur, dass der Lauf mit einem Fehler endet — das täte er auch, wenn die
// Rechnung der Karten durchliefe, denn die Vorlage hört den Abbruch selbst. Hier
// steht, dass [ausfuhrkarten] **nichts** zurückgibt.
func TestEinAbbruchGibtKeineHalbenKarten(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	vorlage := msepaket.Exportvorlage{
		KurzerName:   "probe",
		Kartenfelder: []domain.Field{{Name: "name", Kind: domain.KindText}},
	}
	karten := []*domain.CardRow{{ID: 1, Values: []domain.Value{{Field: "name", Text: "Anton"}}}}

	gerechnet, befunde := ausfuhrkarten(ctx, vorlage, karten, nil, nil)
	if gerechnet != nil || befunde != nil {
		t.Errorf("nach dem Abbruch: %v / %v", gerechnet, befunde)
	}
}

// **Ohne Karten und ohne Kartenfelder gibt es nichts zu rechnen.**
func TestOhneKartenOderKartenfelderRechnetNichts(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{Name: "name", Kind: domain.KindText}}
	karten := []*domain.CardRow{{ID: 1}}

	for name, fall := range map[string]struct {
		felder []domain.Field
		karten []*domain.CardRow
	}{
		"keine Kartenfelder": {nil, karten},
		"keine Karten":       {felder, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gerechnet, befunde := ausfuhrkarten(t.Context(),
				msepaket.Exportvorlage{Kartenfelder: fall.felder}, fall.karten, nil, nil)
			if gerechnet != nil || befunde != nil {
				t.Errorf("%v / %v", gerechnet, befunde)
			}
		})
	}
}
