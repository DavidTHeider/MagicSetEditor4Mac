package app_test

// **Die Punktart setzen** — #2436. MSE2 hat dafür drei Knöpfe am
// Punkteditor, frei, glatt und symmetrisch (`src/gui/symbol/point_editor.cpp`,
// Zeilen 160–162, Fassung f9d9356 vom 2021-07-06), und setzt mit der Sperre
// die beiden Griffe auf ihre Mitte (`ControlPoint::onUpdateLock`,
// `src/data/symbol.cpp`, Zeilen 63–75). Hier geht das über die Rechnung
// `lesen.symbol.sperre`, an Kommandozeile, MCP und Dienst.

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einSchiefesDreieck hat einen ersten Punkt mit zwei ungleichen Griffen.
const einSchiefesDreieck = `[{"art":"Form","name":"a","kombination":"overlap","punkte":[` +
	`{"x":0.5,"y":0.5,"griffDavorX":-0.2,"griffDavorY":0,"griffDanachX":0.1,"griffDanachY":0.1,"sperre":"free"},` +
	`{"x":0.9,"y":0.1,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"sperre":"free"},` +
	`{"x":0.9,"y":0.9,"griffDavorX":0,"griffDavorY":0,"griffDanachX":0,"griffDanachY":0,"sperre":"free"}]}]`

func TestDiePunktartSetztDieSperreUndMitteltDieGriffe(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := app.SymbolRechnen("lesen.symbol.sperre",
		json.RawMessage(`{"teile":`+einSchiefesDreieck+`,"teil":0,"punkt":0,"sperre":"size"}`))
	if fehler != nil {
		t.Fatal(fehler)
	}

	roh, fehler := json.Marshal(ergebnis)
	if fehler != nil {
		t.Fatal(fehler)
	}

	var teile []app.SymbolteilDTO

	if fehler := json.Unmarshal(roh, &teile); fehler != nil || len(teile) != 1 {
		t.Fatalf("die Antwort: %s (%v)", roh, fehler)
	}

	punkt := teile[0].Punkte[0]
	nahe := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// **Symmetrisch heißt: die Mitte und ihr Gegenteil** — (davor − danach) / 2.
	if punkt.Sperre != "size" || !nahe(punkt.GriffDavorX, -0.15) || !nahe(punkt.GriffDavorY, -0.05) ||
		!nahe(punkt.GriffDanachX, 0.15) || !nahe(punkt.GriffDanachY, 0.05) {
		t.Errorf("der Punkt ist %+v", punkt)
	}

	if andere := teile[0].Punkte[1]; andere.Sperre != "free" {
		t.Errorf("ein anderer Punkt änderte sich: %+v", andere)
	}
}

func TestEinePunktartAusserhalbDerDreiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ argumente, kern string }{
		"unbekannte Sperre":   {`{"teile":` + einSchiefesDreieck + `,"teil":0,"punkt":0,"sperre":"klebrig"}`, "erlaubt sind"},
		"Teil gibt es nicht":  {`{"teile":` + einSchiefesDreieck + `,"teil":3,"punkt":0,"sperre":"size"}`, "Stelle"},
		"Punkt gibt es nicht": {`{"teile":` + einSchiefesDreieck + `,"teil":0,"punkt":9,"sperre":"size"}`, "Punkt 9"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := app.SymbolRechnen("lesen.symbol.sperre", json.RawMessage(fall.argumente))
			if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), fall.kern) {
				t.Errorf("Fehler %v, erwartet eine Validierung mit »%s«", fehler, fall.kern)
			}
		})
	}
}
