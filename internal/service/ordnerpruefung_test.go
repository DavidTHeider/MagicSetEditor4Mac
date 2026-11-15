package service_test

// **Ein Ordner wird geprüft, bevor er gilt** — OPS-02.AK-01 und AK-03
// (#1719):
// Existenz, Schreibbarkeit, Platz. **Die Prüfung selbst steht in `app`**,
// weil sie die Platte fragt; hier steht, dass jeder Weg, der eine
// Einstellung setzt, an ihr vorbeikommt — und dass ohne sie nichts gilt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinOrdnerGehtDurchDiePruefung(t *testing.T) {
	t.Parallel()

	gesehen := []string{}
	pruefung := func(_ context.Context, angabe string) error {
		gesehen = append(gesehen, angabe)
		if angabe == "/Volumes/Weg" {
			return domain.NewFieldError("wert", domain.SchluesselValidierung, "nicht erreichbar")
		}

		return nil
	}

	ablage := &einstellungsablage{}
	verwalter := service.NeuerEinstellungsverwalter(ablage).MitOrdnerpruefung(pruefung)
	ctx := t.Context()

	if fehler := verwalter.Setzen(ctx, domain.VorgabeSicherungspfad, "/Volumes/Weg"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein fehlender Ordner wurde angenommen: %v", fehler)
	}

	if _, da := ablage.werte[domain.VorgabeSicherungspfad]; da {
		t.Error("geschrieben wurde trotzdem")
	}

	if fehler := verwalter.Setzen(ctx, domain.VorgabeSicherungspfad, "~/Sicherungen"); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	// **Leer heißt: keine Vorgabe** — dafür wird die Platte nicht gefragt,
	// und eine andere Einstellung auch nicht.
	for schluessel, wert := range map[string]string{
		domain.VorgabeSicherungspfad:     "",
		domain.EinstellungSchriftgroesse: "gross",
	} {
		if fehler := verwalter.Setzen(ctx, schluessel, wert); fehler != nil {
			t.Fatalf("%s: %v", schluessel, fehler)
		}
	}

	if len(gesehen) != 2 || gesehen[1] != "~/Sicherungen" {
		t.Errorf("geprüft wurden %v", gesehen)
	}
}

// **Eine Angabe, die die Domäne abweist, erreicht die Platte gar nicht.**
func TestEineUngueltigeAngabeFragtDiePlatteNicht(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{}).
		MitOrdnerpruefung(func(context.Context, string) error {
			t.Error("die Platte wurde gefragt")

			return nil
		})

	if fehler := verwalter.Setzen(t.Context(), domain.VorgabeAusfuhrpfad, "relativ"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v", fehler)
	}
}
