package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func probekarte(werte ...domain.Value) *domain.CardRow {
	return &domain.CardRow{ID: 1, Number: 0, Values: werte}
}

// Der Behelfsplan ist ein **zeichenbarer** Plan — das ist seine einzige Zusage.
func TestDerBehelfsplanIstZeichenbar(t *testing.T) {
	plan, err := service.NeuerPlanbauer().Behelfsplan(t.Context(),
		probekarte(domain.Value{Field: "name", Text: "Feuerball"}))
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	befunde, err := render.Pruefe(plan)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Jeder Feldwert bekommt eine Zeile — auch ein unbekannter.**
//
// Ein fremdes Spiel führt Felder, die wir nicht kennen; sie wegzulassen hieße,
// die Vorschau ausgerechnet dort zu beschneiden, wo sie gebraucht wird.
func TestJederFeldwertBekommtEineZeile(t *testing.T) {
	plan, err := service.NeuerPlanbauer().Behelfsplan(t.Context(), probekarte(
		domain.Value{Field: "name", Text: "Feuerball"},
		domain.Value{Field: "gibtsInKeinemSpiel", Text: "trotzdem da"},
	))
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	var texte []string

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText {
			texte = append(texte, ebene.Klartext())
		}
	}

	if len(texte) != 2 {
		t.Fatalf("%d Textebenen: %v", len(texte), texte)
	}

	if texte[1] != "gibtsInKeinemSpiel: trotzdem da" {
		t.Errorf("%q", texte[1])
	}
}

// Eine Karte ohne Feldwerte ergibt trotzdem einen Plan.
func TestEineKarteOhneWerteErgibtEinenPlan(t *testing.T) {
	plan, err := service.NeuerPlanbauer().Behelfsplan(t.Context(), probekarte())
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(plan.Ebenen) == 0 || plan.Breite != service.Kartenbreite {
		t.Errorf("%+v", plan)
	}
}

// Der Bau lässt sich abbrechen (§7).
func TestDerPlanbauLaesstSichAbbrechen(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, err := service.NeuerPlanbauer().Behelfsplan(ctx, probekarte()); !errors.Is(err, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war der Abbruch", err)
	}
}
