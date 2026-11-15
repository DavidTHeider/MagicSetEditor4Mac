package service_test

// **Auch die Einstellung richtet keinen Plan über einem verschlüsselten
// Bestand ein** — Nachlauf über Phase 4, Befund 4: der Ordnerdialog prüfte,
// `sicherung.planordner` ließ sich über Fensterbindung, Kommandozeile, MCP
// und REST trotzdem setzen. Der Verwalter ist der Weg, an dem alle vier
// vorbeikommen.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerPlanordnerVerlangtEinenKlartextbestand(t *testing.T) {
	t.Parallel()

	verschluesselt := true
	ablage := &einstellungsablage{}
	verwalter := service.NeuerEinstellungsverwalter(ablage).
		MitVerschluesselungsstand(func(context.Context) (bool, error) { return verschluesselt, nil })

	var feld *domain.FieldError
	if fehler := verwalter.Setzen(t.Context(), domain.SicherungsplanOrdner, "/Volumes/Stick"); !errors.As(fehler, &feld) ||
		feld.Key != domain.SchluesselPlanVerschluesselt {
		t.Fatalf("über einem verschlüsselten Bestand: %v", fehler)
	}

	if _, da := ablage.werte[domain.SicherungsplanOrdner]; da {
		t.Error("geschrieben wurde trotzdem")
	}

	// **Beenden geht immer** — leer heißt: kein Plan.
	if fehler := verwalter.Setzen(t.Context(), domain.SicherungsplanOrdner, ""); fehler != nil {
		t.Errorf("den Plan beenden: %v", fehler)
	}

	verschluesselt = false

	if fehler := verwalter.Setzen(t.Context(), domain.SicherungsplanOrdner, "/Volumes/Stick"); fehler != nil {
		t.Errorf("über einem Klartextbestand: %v", fehler)
	}
}
