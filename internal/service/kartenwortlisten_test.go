package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Wortlisten einer Karte gehen über ihren Stil** —
// #1074.
//
// **Dieselbe Auflösung wie [service.Planbauer.Kartenpalette]**, und aus
// demselben Grund: die Vorschau, die Palette und die Wortliste sprechen über
// dieselbe Karte, und drei Wege zum Stil liefen beim ersten Sonderfall
// auseinander.
func TestDieKartenwortlistenGehenUeberDenStilDerKarte(t *testing.T) {
	karte := karteMitTyp("<word-list-type-en>Creature</word-list-type-en>")

	t.Run("ohne Stilverzeichnis gibt es nichts", func(t *testing.T) {
		bereiche, befunde, err := service.NeuerPlanbauer().
			Kartenwortlisten(t.Context(), karte, "type")
		if err != nil || bereiche != nil || len(befunde) != 0 {
			t.Errorf("%+v / %+v / %v", bereiche, befunde, err)
		}
	})

	t.Run("ohne passenden Stil kommen nur seine Befunde", func(t *testing.T) {
		bauer := service.NeuerPlanbauer().MitStilen(festerStil{
			befunde: []domain.Befund{{Meldung: "kein Stilpaket"}},
		})

		bereiche, befunde, err := bauer.Kartenwortlisten(t.Context(), karte, "type")
		if err != nil || bereiche != nil || len(befunde) != 1 {
			t.Errorf("%+v / %+v / %v", bereiche, befunde, err)
		}
	})

	t.Run("mit Stil kommen die Bereiche", func(t *testing.T) {
		bauer := service.NeuerPlanbauer().MitStilen(festerStil{
			vorlage: vorlageMitListen(
				msepaket.Wortliste{Name: "type-en", Woerter: typenliste()}),
			da: true,
		})

		bereiche, _, err := bauer.Kartenwortlisten(t.Context(), karte, "type")
		if err != nil {
			t.Fatalf("Kartenwortlisten: %v", err)
		}

		if len(bereiche) != 1 || bereiche[0].Menue.Gewaehlt != 2 {
			t.Errorf("die Bereiche sind %+v", bereiche)
		}
	})
}

// **Und eine Wahl geht denselben Weg.**
func TestEineKartenwortwahlGehtUeberDenselbenStil(t *testing.T) {
	karte := karteMitTyp("<word-list-type-en>Creature</word-list-type-en>")

	t.Run("ohne Stil ist der Bereich nicht da", func(t *testing.T) {
		_, _, err := service.NeuerPlanbauer().
			Kartenwortwahl(t.Context(), karte, "type", 0, []int{0})

		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("der Fehler ist %v", err)
		}
	})

	t.Run("mit Stil kommt der neue Wert", func(t *testing.T) {
		bauer := service.NeuerPlanbauer().MitStilen(festerStil{
			vorlage: vorlageMitListen(
				msepaket.Wortliste{Name: "type-en", Woerter: typenliste()}),
			da: true,
		})

		neu, _, err := bauer.Kartenwortwahl(t.Context(), karte, "type", 0, []int{1})
		if err != nil {
			t.Fatalf("Kartenwortwahl: %v", err)
		}

		if neu != "<word-list-type-en>Legendary Creature</word-list-type-en>" {
			t.Errorf("daraus wird %q", neu)
		}
	})
}
