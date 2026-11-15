package domain_test

// **Die Ordnung in einer Zone und die Zonenstelle, an die gelegt wird** —
// Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum es das überhaupt gibt
//
// Weil „ins Deck zurück" ohne eine Zonenstelle **unbestimmt** ist: oben, unten
// oder eingemischt sind drei verschiedene Handlungen, und eine Karte ohne
// Angabe landete an einer Zonenstelle, die niemand gewählt hat.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestZoneGeordnet(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		zone     domain.Zone
		geordnet bool
		warum    string
	}{
		{domain.ZoneBibliothek, true, "die Reihenfolge ist ihr Geheimnis"},
		{domain.ZoneStapel, true, "zuletzt gelegt wird zuerst aufgelöst"},
		{domain.ZoneHand, false, "eine Hand ist keine Reihe"},
		{domain.ZoneTisch, false, "der Tisch ist eine Fläche"},
		{domain.ZoneFriedhof, true, "CR 400.5 nennt ihn unter den geordneten — und CR 404.1 legt oben darauf"},
		{domain.ZoneExil, false, "dito"},
		{domain.ZoneKommando, false, "dito"},
	} {
		t.Run(string(fall.zone), func(t *testing.T) {
			t.Parallel()

			if fall.zone.Geordnet() != fall.geordnet {
				t.Errorf("%s geordnet=%v, erwartet %v — %s",
					fall.zone, fall.zone.Geordnet(), fall.geordnet, fall.warum)
			}
		})
	}

	// **Eine unbekannte Zone ist keine geordnete.** Wer sie für eine hielte,
	// legte an eine Zonenstelle in etwas, das es nicht gibt.
	if domain.Zone("kiste").Geordnet() {
		t.Error("eine unbekannte Zone gilt als geordnet")
	}
}

func TestIstStelle(t *testing.T) {
	t.Parallel()

	for _, stelle := range domain.Zonenstellen() {
		if !domain.IstZonenstelle(stelle) {
			t.Errorf("%q steht in Zonenstellen() und gilt nicht als Zonenstelle", stelle)
		}
	}

	if len(domain.Zonenstellen()) != 3 {
		t.Errorf("%d Zonenstellen, erwartet drei", len(domain.Zonenstellen()))
	}

	for _, keine := range []domain.Zonenstelle{"", "mitte", "OBEN"} {
		if domain.IstZonenstelle(keine) {
			t.Errorf("%q gilt als Zonenstelle", keine)
		}
	}
}
