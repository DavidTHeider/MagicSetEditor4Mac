package service_test

// **Der Plan nennt das Paket, mit dem er gebaut ist** — #2314 (E2314).
//
// Eine Karte, die den Stil des Sets erbt, schreibt `''`; welches Paket das
// ist, entscheidet das Set — und das kann ein anderer Prozess ändern, während
// das Fenster die Karte zeigt. Das Formular ordnet seine Felder je Paket,
// und das Paket kennt nur, wer den Stil auflöst: der Planbauer. **Ohne
// zweite Abfrage**: er hat den Stil ohnehin in der Hand (Budget aus E02).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// setstil ist eine Stilquelle, deren Setstil ein anderer Prozess ändern kann.
type setstil struct{ paket string }

func (s *setstil) StilFuerSet(context.Context, int64) (service.Vorlage, []domain.Befund, bool) {
	return service.Vorlage{Stil: msepaket.Stil{Breite: 375, Hoehe: 523, Paket: s.paket}}, nil, true
}

func (s *setstil) StilFuerKarte(ctx context.Context, karte *domain.CardRow) (service.Vorlage, []domain.Befund, bool) {
	return s.StilFuerSet(ctx, karte.SetID)
}

func TestDerPlanFolgtEinemFremdenSetstilwechsel(t *testing.T) {
	t.Parallel()

	quelle := &setstil{paket: "magic-alt.mse-style"}
	planbauer := service.NeuerPlanbauer().MitStilen(quelle)
	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Erbt"}}}

	vorher, _, err := planbauer.Kartenplan(t.Context(), karte)
	if err != nil || vorher.Stil != "magic-alt.mse-style" {
		t.Fatalf("vorher: %q, %v", vorher.Stil, err)
	}

	quelle.paket = "magic-neu.mse-style"

	nachher, _, err := planbauer.Kartenplan(t.Context(), karte)
	if err != nil || nachher.Stil != "magic-neu.mse-style" {
		t.Errorf("nach dem fremden Wechsel: %q, %v", nachher.Stil, err)
	}
}

func TestDerBehelfsplanNenntKeinPaket(t *testing.T) {
	t.Parallel()

	plan, _, err := service.NeuerPlanbauer().Kartenplan(t.Context(), &domain.CardRow{ID: 1})
	if err != nil || plan.Stil != "" {
		t.Errorf("der Behelfsplan nennt %q, %v", plan.Stil, err)
	}
}
