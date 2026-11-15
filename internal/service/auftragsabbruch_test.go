package service_test

// **Ein Auftrag des Fensters zählt und lässt sich anhalten** — FMT-03.AK-02
// (#1692).

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinAbbrechbarerAuftragLaesstSichAnhalten(t *testing.T) {
	t.Parallel()

	buch := service.NeuesAuftragsbuch(time.Now)
	id := buch.Beginnen("bogen", "Mein Set")
	angehalten := 0

	buch.Abbrechbar(id, func() { angehalten++ })
	buch.Zaehlen(id, domain.Fortschritt{Erledigt: 12, Gesamt: 300, Einheit: domain.EinheitKarte})

	if laufend := buch.Auftraege()[0]; !laufend.Abbrechbar || laufend.Gezaehlt.Erledigt != 12 ||
		laufend.Gezaehlt.Gesamt != 300 || laufend.Gezaehlt.Einheit != domain.EinheitKarte {
		t.Fatalf("laufend: %+v", laufend)
	}

	if err := buch.Abbrechen(id); err != nil || angehalten != 1 {
		t.Fatalf("abbrechen: %v, %d-mal angehalten", err, angehalten)
	}

	// **Ein Abbruch ist kein Scheitern** — ohne Grund und ohne Rot.
	buch.Beenden(id, fmt.Errorf("gelegt: %w", context.Canceled))

	if beendet := buch.Auftraege()[0]; beendet.Stand != domain.AuftragAbgebrochen || beendet.Grund != "" ||
		beendet.Abbrechbar {
		t.Errorf("beendet: %+v", beendet)
	}

	if err := buch.Abbrechen(id); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein beendeter Auftrag ließ sich abbrechen: %v", err)
	}
}

// **Nicht jeder Auftrag lässt sich anhalten** — Preise und Sammeldaten laufen
// zu Ende, und das Nein sagt es, statt still nichts zu tun.
func TestEinAuftragOhneAbbruchSagtNein(t *testing.T) {
	t.Parallel()

	buch := service.NeuesAuftragsbuch(time.Now)
	andere := buch.Beginnen("sammeldaten", "")
	id := buch.Beginnen("preise", "")

	buch.Abbrechbar(99, func() { t.Error("eine fremde Kennung wurde abbrechbar") })
	buch.Zaehlen(99, domain.Fortschritt{Erledigt: 1})

	if err := buch.Abbrechen(id); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("abbrechen ohne Abbruch: %v", err)
	}

	if alle := buch.Auftraege(); alle[0].Gezaehlt != (domain.Fortschritt{}) || alle[1].Gezaehlt != (domain.Fortschritt{}) {
		t.Errorf("eine fremde Kennung hat gezählt: %+v", alle)
	}

	buch.Beenden(andere, nil)
	buch.Beenden(id, nil)
}
