package rest_test

// **Die Einstellungen eines Menschen und seine Auskunft am willigen Doppel** —
// Werkstück 8 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `vorrichtungbestand_test.go` sonst über die
// Zeilengrenze aus §2 ginge.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// Nutzereinstellungen gibt die Reihe aus der Domäne mit ihren Vorgaben.
//
// **Aus der Domäne und nicht aus einer Kopie**: ein Doppel, das eigene
// Schlüssel nennte, prüfte etwas, das es im Betrieb nicht gibt (§3).
func (williger) Nutzereinstellungen(context.Context, string) (
	[]rest.Einstellungswahl, error,
) {
	return vorgabewahlen(), nil
}

func (williger) NutzereinstellungSetzen(_ context.Context, _, schluessel, wert string) (
	[]rest.Einstellungswahl, error,
) {
	gesetzt := vorgabewahlen()

	for i := range gesetzt {
		if gesetzt[i].Schluessel == schluessel {
			gesetzt[i].Wert = wert
		}
	}

	return gesetzt, nil
}

func (williger) Personendaten(context.Context, string) (json.RawMessage, error) {
	return json.RawMessage(`{"erhoben":"2026-09-05T00:00:00Z","nutzer":{"name":"Ich"}}`), nil
}

// vorgabewahlen baut die Reihe mit ihren Vorgabewerten.
func vorgabewahlen() []rest.Einstellungswahl {
	reihe := domain.Nutzereinstellungsreihe()
	wahlen := make([]rest.Einstellungswahl, 0, len(reihe))

	for _, schluessel := range reihe {
		wahlen = append(wahlen, rest.Einstellungswahl{
			Schluessel: schluessel,
			Wert:       domain.NutzerVorgabe(schluessel),
			Werte:      domain.ErlaubteNutzerwerte(schluessel),
		})
	}

	return wahlen
}
