//go:build darwin

package biometrie

import (
	"context"
	"fmt"

	"github.com/tmc/apple/localauthentication"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bindung ist alles, was diese Datei vom Framework braucht.
//
// # Warum eine Naht und nicht der unmittelbare Ruf
//
// **Weil sonst jede Verzweigung hier den Sensor bräuchte.** Ein Rechner ohne
// Touch ID, ein Face ID, ein Fehlschlag — das sind drei Wege, und keiner
// davon lässt sich auf **diesem** Rechner herbeiführen. Mit der Naht sind sie
// prüfbar, und ungemessen bleibt nur, was wirklich ungemessen bleiben muss:
// die drei Rufe in `bindung_darwin.go` (§3, `.covignore`).
type bindung struct {
	kann      func() (bool, error)
	merkmal   func() string
	pruefe    func(ctx context.Context, grund string) (bool, error)
	schliesse func()
}

// oeffne ist die Naht — im Betrieb [neueBindung], im Lauf ein Doppel.
var oeffne = neueBindung

// stand fragt `canEvaluatePolicy` — **ohne Dialog und ohne Sensor**.
func stand(ctx context.Context) Auskunft {
	var auskunft Auskunft

	// **Auch die Auskunft steht hinter dem Riegel.** Sie ruft dieselbe
	// Bindung, und ein Absturz beim bloßen Nachfragen wäre der schlimmste:
	// er käme, bevor jemand etwas verlangt hat.
	if fehler := Gefangen(func() error {
		auskunft = fragen(ctx)

		return nil
	}); fehler != nil {
		return Auskunft{Grund: fehler.Error()}
	}

	return auskunft
}

// fragen holt die Auskunft aus dem Framework.
func fragen(_ context.Context) Auskunft {
	gebunden := oeffne()
	defer gebunden.schliesse()

	geht, fehler := gebunden.kann()
	if !geht {
		// **Der Grund des Frameworks schlägt den eigenen Satz.** „Kein Touch
		// ID eingerichtet" und „Touch ID gesperrt nach zu vielen Versuchen"
		// sind zwei Lagen mit zwei Antworten, und nur das Framework kennt
		// den Unterschied.
		grund := "Dieser Rechner hat kein Touch ID oder es ist nicht eingerichtet."
		if fehler != nil {
			grund = fehler.Error()
		}

		return Auskunft{Grund: grund}
	}

	return Auskunft{Vorhanden: true, Art: gebunden.merkmal()}
}

// bestaetige zeigt den Dialog und wartet auf den Sensor.
func bestaetige(ctx context.Context, grund string) error {
	gebunden := oeffne()
	defer gebunden.schliesse()

	geglueckt, fehler := gebunden.pruefe(ctx, grund)
	if fehler != nil {
		return fmt.Errorf("%w: %s", domain.ErrValidation, fehler.Error())
	}

	if !geglueckt {
		return fmt.Errorf(
			"%w: die Abfrage wurde nicht bestätigt", domain.ErrValidation)
	}

	return nil
}

// artNach gibt den Namen, unter dem Apple das Merkmal führt.
//
// **Die Namen sind Eigennamen und werden nicht übersetzt** (§9) — „Touch ID"
// heißt in jeder Sprache so.
func artNach(art localauthentication.LABiometryType) string {
	switch art {
	case localauthentication.LABiometryTypeTouchID:
		return "Touch ID"
	case localauthentication.LABiometryTypeFaceID:
		return "Face ID"
	}

	return ""
}
