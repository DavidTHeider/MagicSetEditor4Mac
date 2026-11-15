package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jede schwere Arbeit belegt ihren Platz, bevor sie beginnt** — Werkstück O5
// von Phase 5.2 (#1429).
//
// Der Beleg ist derselbe für alle: ein Haushalt, dessen Klasse besetzt ist,
// und ein Ruf, dessen Kontext schon zu Ende ist. Käme die Arbeit vor dem
// Platz, griffe sie in ihre — hier absichtlich fehlenden — Nähte und stürzte.

func besetzterHaushalt(t *testing.T, klasse domain.Ressourcenklasse) *service.Haushalt {
	t.Helper()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel[klasse] = 1
	haushalt := service.NeuerHaushalt(grenzen)

	if _, ok := haushalt.Versuchen(klasse); !ok {
		t.Fatal("Platz nicht belegbar")
	}

	return haushalt
}

func beendeterKontext(t *testing.T) context.Context {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	return ctx
}

func TestJedeSchwereArbeitWartetAufIhrenPlatz(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		klasse domain.Ressourcenklasse
		ruf    func(ctx context.Context, haushalt *service.Haushalt) error
	}{
		{"Set einführen", domain.RessourceEinfuhr, func(ctx context.Context, h *service.Haushalt) error {
			i := service.NeuerImporteur(nil)
			i.MitHaushalt(h)
			_, err := i.Set(ctx, "/nirgends.mse-set")

			return err
		}},
		{"MTG JSON einführen", domain.RessourceEinfuhr, func(ctx context.Context, h *service.Haushalt) error {
			i := service.NeuerImporteur(nil)
			i.MitHaushalt(h)
			_, err := i.MTGJSON(ctx, "/nirgends.json", "magic")

			return err
		}},
		{"MTG JSON ausgeben", domain.RessourceAusfuhr, func(ctx context.Context, h *service.Haushalt) error {
			e := service.NeuerExporteur(nil)
			e.MitHaushalt(h)
			_, err := e.MTGJSON(ctx, 1, "magic", "/nirgends.json")

			return err
		}},
		{"Bogen legen", domain.RessourceAusfuhr, func(ctx context.Context, h *service.Haushalt) error {
			b := service.NeuerBogenleger(nil, nil, nil)
			b.MitHaushalt(h)
			_, err := b.Lege(ctx, 1, "/nirgends.pdf", service.Bogenmasse{}, domain.Kartenauswahl{})

			return err
		}},
		{"Masse ändern", domain.RessourceMasse, func(ctx context.Context, h *service.Haushalt) error {
			m := service.NeueMassenpflege(nil)
			m.MitHaushalt(h)
			_, err := m.Aendern(ctx, service.Massenaenderung{Zeilen: []service.Massenzeile{{}}})

			return err
		}},
		{"Masse löschen", domain.RessourceMasse, func(ctx context.Context, h *service.Haushalt) error {
			m := service.NeueMassenpflege(nil)
			m.MitHaushalt(h)
			_, err := m.Loeschen(ctx, []service.Massenzeile{{}})

			return err
		}},
		{"Preise in Masse", domain.RessourceMasse, func(ctx context.Context, h *service.Haushalt) error {
			p := service.NeuePreispflege(nil, nil, nil, nil)
			p.MitHaushalt(h)
			_, err := p.PreiseInMasse(ctx, time.Hour)

			return err
		}},
		{"Bild holen", domain.RessourceBild, func(ctx context.Context, h *service.Haushalt) error {
			b := service.NeuerBildweg(nil, nil, "")
			b.MitHaushalt(h)
			_, err := b.Holen(ctx, "Lightning Bolt", "png")

			return err
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			err := fall.ruf(beendeterKontext(t), besetzterHaushalt(t, fall.klasse))
			if !errors.Is(err, context.Canceled) {
				t.Errorf("%s: %v", fall.name, err)
			}
		})
	}
}

func TestDerMaschinenhaushaltIstEinerJeProzess(t *testing.T) {
	t.Parallel()

	erster, zweiter := service.Maschinenhaushalt(), service.Maschinenhaushalt()
	if erster != zweiter {
		t.Error("zwei Haushalte")
	}

	// **Ohne Einsatz gilt der der Maschine**: ein Dienst ohne `MitHaushalt`
	// bekommt dort seinen Platz — und scheitert danach erst an der Fachregel.
	_, err := service.NeueMassenpflege(nil).Loeschen(t.Context(), nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ohne eigenen Haushalt: %v", err)
	}

	if service.Maschinenhaushalt().Grenzen().Parallel[domain.RessourceRendern] != domain.Vorgabegrenzen().Parallel[domain.RessourceRendern] {
		t.Error("nicht die Vorgabegrenzen")
	}
}
