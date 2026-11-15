// **Die Rücknahme als Werkzeug**
// (#1107).
//
// **Beide Richtungen aus einer Funktion**: sie unterscheiden sich im Namen,
// im Satz und im Aufruf, und drei Unterschiede rechtfertigen keine zweite
// Kopie derselben Wache.

package mcp

import (
	"context"
	"encoding/json"
)

// bewegungswerkzeug gibt `schreiben.zuruecknehmen` oder `…wiederherstellen`.
func bewegungswerkzeug(quelle Bestandsquelle, zurueck bool) Werkzeug {
	name := "schreiben.wiederherstellen"
	satz := "ÄNDERT das Set: führt die jüngste zurückgenommene Änderung wieder aus. " +
		"Mit »vorgang« GENAU diesen Schritt statt des jüngsten. " +
		"»bewegt« ist falsch, wenn es nichts wiederherzustellen gab."

	if zurueck {
		name = "schreiben.zuruecknehmen"
		satz = "ÄNDERT das Set: macht die jüngste noch geltende Änderung rückgängig — " +
			"auch eine Massenänderung, die viele Karten berührt hat. " +
			"MIT »vorgang« aus »schreiben.felder« geht GENAU diese Änderung " +
			"zurück, auch wenn inzwischen jemand anderes in dasselbe Set " +
			"geschrieben hat. OHNE ihn der jüngste Schritt des Sets. " +
			"»bewegt« ist falsch, wenn der Verlauf am Ende ist — oder wenn es " +
			"den genannten Vorgang nicht (mehr) gibt."
	}

	// **Die Methode wird erst im Ruf geholt, nicht hier.** Ein
	// `quelle.Wiederherstellen` beim Bauen greift auf eine Schnittstelle zu,
	// die noch `nil` sein darf — und die Werkzeugliste wird in Tests mit
	// leeren Nähten gebaut, um die *Anmeldung* zu prüfen. Der erste Anlauf
	// stürzte dort ab.
	richtung := func(ctx context.Context, setID int64, vorgang string) (Bewegung, error) {
		if zurueck {
			return quelle.Zuruecknehmen(ctx, setID, vorgang)
		}

		return quelle.Wiederherstellen(ctx, setID, vorgang)
	}

	return Werkzeug{
		Name: name, Art: Schreibend, Erklaerung: satz,
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"vorgang": text("Optional: die Kennung aus »schreiben.felder«. " +
				"Ohne sie der jüngste Schritt des Sets."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				SetID   int64  `json:"setID"`
				Vorgang string `json:"vorgang"`
			}

			if err := lies(argumente, &wahl); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", wahl.SetID); err != nil {
				return "", err
			}

			bewegung, fehler := richtung(ctx, wahl.SetID, wahl.Vorgang)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(bewegung)
		},
	}
}
