//go:build fremdpakete

package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Demokarten liegen in einem Set, und das Set hat Angaben.**
//
// # Warum das eine eigene Zusage ist
//
// Bis #384
// setzte der Lauf **keine** Setangaben. Damit war `set.*` im ganzen Lauf leer —
// und das ist kein Randfall: über `M15-Magic-Pack` sind es 420 Zugriffe auf
// `set.symbol`, 345 auf `set.border_color`, 101 auf
// `set.mainframe_rarity_name`.
//
// Sichtbar wurde es am **Setcode**. Er ist kein Kartenwert:
//
//	card field:
//		name: set code
//		editable: false
//		script: set_code_script(value, margin_code:"setcode1")
//
// `magic.mse-game/card_fields`, Zeilen 3080–3091; das Skript dahinter liest
// `set.set_code` (`magic.mse-game/script`, Zeilen 2951–2962). Ohne Set stand im
// Fuß ein Leerzeichen.

// setcode ist der Code des Demosets — **erfunden** (§11).
const setcode = "PRB"

// demoset sind die Angaben des Sets, in dem die Demokarten liegen.
//
// **Selbst geschrieben wie die Karten.** Kein Wert stammt aus einem fremden
// Paket; `PRB` ist kein Setcode, den es gibt.
func demoset() []domain.Value {
	return []domain.Value{
		{Field: "title", Text: "Probeserie"},
		{Field: "set code", Text: setcode},
		{Field: "set language", Text: "DE"},
		{Field: "artist", Text: "Marta Feldhaus"},
		{Field: "copyright", Text: "™ & © 2026 Beispielverlag"},
		{Field: "card number style", Text: "1/99"},
	}
}

// pruefeSetcode verlangt den Setcode im Fuß der Karte.
//
// **Im Plan und nicht im Bild**: dass die Zeile *lesbar* ist, misst
// [helleImFuss]; dass dort der **richtige** Text steht, sagt nur der Plan
// (§10 — die Reichweite einer Messung).
func pruefeSetcode(t *testing.T, plan *render.Plan) {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art != render.ArtText {
			continue
		}

		if strings.Contains(plan.Ebenen[i].Klartext(), setcode) {
			return
		}
	}

	t.Errorf("der Setcode %q steht auf keiner Textebene — das Set kommt nicht an", setcode)
}
