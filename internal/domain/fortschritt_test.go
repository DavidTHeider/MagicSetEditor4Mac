package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinFortschrittLiestSichAlsSatz(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		stand domain.Fortschritt
		satz  string
	}{
		{"mit Gesamtzahl", domain.Fortschritt{Erledigt: 12, Gesamt: 300, Einheit: domain.EinheitKarte}, "12 von 300 Karten"},
		{"ohne Gesamtzahl", domain.Fortschritt{Erledigt: 2, Einheit: domain.EinheitSet}, "2 Sets"},
		{"ein Paket", domain.Fortschritt{Erledigt: 1, Gesamt: 1, Einheit: domain.EinheitPaket}, "1 von 1 Paketen"},
		{"ein Bild", domain.Fortschritt{Erledigt: 3, Gesamt: 9, Einheit: domain.EinheitBild}, "3 von 9 Bildern"},
		{"unbekannte Einheit", domain.Fortschritt{Erledigt: 3, Gesamt: 9, Einheit: "x"}, "3 von 9"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if satz := fall.stand.Satz(); satz != fall.satz {
				t.Errorf("Satz: %q, erwartet %q", satz, fall.satz)
			}
		})
	}
}
