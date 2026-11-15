package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Auskunft antwortet, wie der Wechsel fragen würde — ohne zu fragen**
// (ID-03, #1702). Dieselbe Reihenfolge wie `PruefeBeimWechsel`: die
// Einstellung zuerst, dann die Faktoren. **Kein Sensor wird berührt**: er ist
// in diesem Lauf vorhanden, und ein Faktor ist Biometrie — eine Auskunft, die
// über `Pruefe` ginge, öffnete hier einen Dialog, und das Doppel merkte es.
func TestDieAuskunftAntwortetWieDerWechselFragte(t *testing.T) {
	t.Parallel()

	totp := []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}}
	finger := []domain.Faktor{{NutzerID: 1, Art: domain.FaktorBiometrie}}

	faelle := []struct {
		name          string
		faktoren      []domain.Faktor
		einstellungen *einstellungsdoppel
		verlangt      bool
	}{
		{"ohne Faktor", nil, &einstellungsdoppel{}, false},
		{"mit Faktor und der Vorgabe", totp, &einstellungsdoppel{}, true},
		{"mit Faktor, beim Wechsel abgeschaltet", totp, &einstellungsdoppel{
			werte: map[string]string{domain.NutzerFaktorbeimWechsel: domain.FaktorNurAnmelden},
		}, false},
		{"mit Faktor und ohne Einstellungen", totp, nil, true},
		{"mit Biometrie und einem Sensor", finger, &einstellungsdoppel{}, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := &faktordoppel{faktoren: fall.faktoren}
			sensor := &sensordoppel{vorhanden: true}

			dienst := service.NeuerZweiterfaktor(ablage, ablage, sensor)
			if fall.einstellungen != nil {
				dienst = dienst.MitEinstellungen(service.NeueNutzereinstellungen(fall.einstellungen))
			}

			verlangt, fehler := dienst.VerlangtBeimWechsel(t.Context(), 1)
			if fehler != nil || verlangt != fall.verlangt {
				t.Fatalf("verlangt %v (%v), erwartet %v", verlangt, fehler, fall.verlangt)
			}

			if sensor.gefragt != "" {
				t.Errorf("die Auskunft fragte den Sensor: %q", sensor.gefragt)
			}
		})
	}
}

// **Ein unlesbarer Bestand ist keine Antwort** — weder an der Einstellung
// noch an den Faktoren.
func TestEineUnlesbareAuskunftIstKeinNein(t *testing.T) {
	t.Parallel()

	einstellungsfehler := service.NeuerZweiterfaktor(&faktordoppel{}, &faktordoppel{}, &sensordoppel{}).
		MitEinstellungen(service.NeueNutzereinstellungen(&einstellungsdoppel{fehler: errStummerBestand}))

	if _, fehler := einstellungsfehler.VerlangtBeimWechsel(t.Context(), 1); !errors.Is(fehler, errStummerBestand) {
		t.Errorf("die Einstellung unlesbar: %v", fehler)
	}

	faktorfehler := &faktordoppel{kaputt: errStummerBestand}

	if _, fehler := service.NeuerZweiterfaktor(faktorfehler, faktorfehler, &sensordoppel{}).
		VerlangtBeimWechsel(t.Context(), 1); !errors.Is(fehler, errStummerBestand) {
		t.Errorf("die Faktoren unlesbar: %v", fehler)
	}
}
