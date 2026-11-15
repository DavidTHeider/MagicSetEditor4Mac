package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorgabenplan gibt einen Plan mit den gegebenen Vorgaben.
//
// **Über den echten Planbauer gelegt**, statt die Schnittstelle noch einmal
// abzubilden: `Kartenpalette` gehört dazu und hat mit Vorgabewerten nichts zu
// tun; sie hier zu wiederholen wäre eine zweite Fassung, die beim nächsten
// Zusatz vergessen wird.
type vorgabenplan struct {
	*service.Planbauer

	vorgaben []render.Vorgabe
}

func (v vorgabenplan) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{Breite: 10, Hoehe: 14, Vorgaben: v.vorgaben}, nil, nil
}

// **Die gerechneten Vorgabewerte erreichen das Fenster** (#563).
//
// Die Naht ist kurz und wird trotzdem geprüft: sie ist die Stelle, an der eine
// umbenannte Angabe still ankommt — im Editor stünde dann an jedem Feld
// „Hergeleitet:" ohne Wert.
func TestDieVorgabenErreichenDieVorschau(t *testing.T) {
	stand := bindungsprobe(t)
	stand.Anwendung.MitEditor(
		stand.Schreiber,
		karteprobe{karte: &domain.CardRow{ID: 1}},
		vorgabenplan{
			Planbauer: service.NeuerPlanbauer(),
			vorgaben:  []render.Vorgabe{{Feld: "card color", Wert: "blau"}},
		},
		render.NeuerRasterer(ohneBilder{}),
	)

	antwort := stand.Bindungen.Vorschau(t.Context(), "1", 1)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Vorschau == nil {
		t.Fatal("keine Vorschau")
	}

	if len(antwort.Vorschau.Vorgaben) != 1 {
		t.Fatalf("%+v", antwort.Vorschau.Vorgaben)
	}

	if antwort.Vorschau.Vorgaben[0].Feld != "card color" ||
		antwort.Vorschau.Vorgaben[0].Wert != "blau" {
		t.Errorf("%+v", antwort.Vorschau.Vorgaben[0])
	}
}

// **Ohne Vorgaben eine leere Liste.** Der Grund steht am DTO: ein `null`
// zwingt die Oberfläche zu zwei Prüfungen statt einer.
func TestOhneVorgabenKommtEineLeereListe(t *testing.T) {
	stand := bindungsprobe(t)
	stand.Anwendung.MitEditor(
		stand.Schreiber,
		karteprobe{karte: &domain.CardRow{ID: 1}},
		vorgabenplan{Planbauer: service.NeuerPlanbauer()},
		render.NeuerRasterer(ohneBilder{}),
	)

	antwort := stand.Bindungen.Vorschau(t.Context(), "1", 1)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Vorschau == nil || antwort.Vorschau.Vorgaben == nil {
		t.Fatalf("%+v", antwort.Vorschau)
	}

	if len(antwort.Vorschau.Vorgaben) != 0 {
		t.Errorf("%+v", antwort.Vorschau.Vorgaben)
	}
}
