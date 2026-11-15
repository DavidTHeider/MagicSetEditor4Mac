// **Was ein Profil vorgeben kann** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// volleQuellen sind fünf Quellen, die je etwas hergeben.
func volleQuellen() app.Vorgabenquellen {
	return app.Vorgabenquellen{
		Formate: func(context.Context, string) ([]domain.Format, error) {
			return []domain.Format{
				{ID: 5, Name: "Standard", Spiel: "magic"},
			}, nil
		},
		Decks: func(context.Context, string) ([]domain.Deck, error) {
			return []domain.Deck{{ID: 7, Name: "Mein Deck", Spiel: "magic"}}, nil
		},
		Pakete: func(context.Context) ([]domain.Paket, error) {
			return []domain.Paket{
				{Art: "mse-game", Name: "magic", Kurzname: "Magic"},
				{Art: "mse-style", Name: "magic-m15"},
				{Art: "mse-include", Name: "magic-mana-small"},
			}, nil
		},
		Lagerorte: func(context.Context) ([]service.Lagerortzeile, error) {
			return []service.Lagerortzeile{
				{Lagerort: "Schrank", Stueck: 4},
				{Lagerort: "", Stueck: 2},
			}, nil
		},
	}
}

// **Jede der fünf Arten bekommt ihre Werte — und ihre Beschriftung.**
func TestJedeVorgabeartHatIhrAngebot(t *testing.T) {
	t.Parallel()

	angebot := app.NeuesVorgabenangebot(volleQuellen())

	faelle := map[domain.Vorgabeart]service.Vorgabewahl{
		domain.VorgabeFormat: {Wert: "5", Beschriftung: "Standard (magic)"},

		// **Der Kurzname, wenn es einen gibt** — denselben Rückfall benutzt
		// MSE.
		domain.VorgabeSpiel: {Wert: "magic", Beschriftung: "Magic"},

		// **Und der Name, wenn nicht.**
		domain.VorgabeVorlage:  {Wert: "magic-m15", Beschriftung: "magic-m15"},
		domain.VorgabeDeck:     {Wert: "7", Beschriftung: "Mein Deck (magic)"},
		domain.VorgabeLagerort: {Wert: "Schrank", Beschriftung: "Schrank"},
	}

	for art, erwartet := range faelle {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			wahlen, fehler := angebot.Angebot(t.Context(), art)
			if fehler != nil {
				t.Fatalf("%v", fehler)
			}

			if len(wahlen) != 1 {
				t.Fatalf("es sind %d Wahlen: %+v", len(wahlen), wahlen)
			}

			if wahlen[0] != erwartet {
				t.Errorf("Wahl = %+v, erwartet %+v", wahlen[0], erwartet)
			}
		})
	}
}

// **Ein Lagerort ohne Namen ist keine Wahl.**
//
// „Ohne Angabe" ist kein Ort, den man zuordnen kann — er steht in der
// Sammlungssicht als Sammelposten. Der Fall ist in [volleQuellen] schon
// enthalten; hier steht, **warum** es nur eine Wahl gibt.
func TestEinLagerortOhneNamenIstKeineWahl(t *testing.T) {
	t.Parallel()

	wahlen, fehler := app.NeuesVorgabenangebot(volleQuellen()).Angebot(
		t.Context(), domain.VorgabeLagerort)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	for _, wahl := range wahlen {
		if wahl.Wert == "" {
			t.Errorf("der leere Lagerort steht zur Wahl: %+v", wahlen)
		}
	}
}

// **Eine Art, die es nicht gibt, ist ein leeres Angebot und kein Fehler.**
func TestEineUnbekannteArtHatEinLeeresAngebot(t *testing.T) {
	t.Parallel()

	wahlen, fehler := app.NeuesVorgabenangebot(volleQuellen()).Angebot(
		t.Context(), "kartenrücken")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	// **Leer und nicht `nil`** (§3): über die JSON-Naht würde daraus `null`.
	if wahlen == nil || len(wahlen) != 0 {
		t.Errorf("Angebot = %+v", wahlen)
	}
}

// **Jede Quelle reicht ihr Nein durch.**
func TestJedeVorgabenquelleReichtIhrNeinDurch(t *testing.T) {
	t.Parallel()

	stolpern := errors.New("die Platte ist voll")

	faelle := map[domain.Vorgabeart]func(app.Vorgabenquellen) app.Vorgabenquellen{
		domain.VorgabeFormat: func(q app.Vorgabenquellen) app.Vorgabenquellen {
			q.Formate = func(context.Context, string) ([]domain.Format, error) {
				return nil, stolpern
			}

			return q
		},
		domain.VorgabeDeck: func(q app.Vorgabenquellen) app.Vorgabenquellen {
			q.Decks = func(context.Context, string) ([]domain.Deck, error) {
				return nil, stolpern
			}

			return q
		},
		domain.VorgabeSpiel: func(q app.Vorgabenquellen) app.Vorgabenquellen {
			q.Pakete = func(context.Context) ([]domain.Paket, error) {
				return nil, stolpern
			}

			return q
		},
		domain.VorgabeLagerort: func(q app.Vorgabenquellen) app.Vorgabenquellen {
			q.Lagerorte = func(context.Context) ([]service.Lagerortzeile, error) {
				return nil, stolpern
			}

			return q
		},
	}

	for art, kaputt := range faelle {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			angebot := app.NeuesVorgabenangebot(kaputt(volleQuellen()))

			if _, fehler := angebot.Angebot(t.Context(), art); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}

	// **Und die Vorlage geht durch dieselbe Quelle wie das Spiel** — der Fall
	// steht eigens da, weil er sonst nur durch die Zeile darüber mitliefe.
	kaputt := volleQuellen()
	kaputt.Pakete = func(context.Context) ([]domain.Paket, error) {
		return nil, stolpern
	}

	if _, fehler := app.NeuesVorgabenangebot(kaputt).Angebot(
		t.Context(), domain.VorgabeVorlage); fehler == nil {
		t.Error("die Vorlage reichte ihr Nein nicht durch")
	}
}
