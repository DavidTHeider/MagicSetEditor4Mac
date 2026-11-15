package app

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Anklickbar ist, was auf der Karte liegt** — und sonst nichts.
//
// # Wofür
//
// MSE fragt vor jedem Feld dieselbe Frage:
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356 vom
// 2021-07-06. `isVisible` ist die geometrische Hälfte davon
// (`src/render/value/viewer.cpp`, Zeilen 36–42) — eine Ebene ohne Fläche oder
// an der Sentinelstelle 1.000.000 liegt nicht auf der Karte.
//
// # Was das ohne die Prüfung kostet
//
// Ein Feldstil ohne `left:` trägt MSEs Vorgabe 1.000.000
// ([massvorbelegung.go]). Daraus wurde eine Schaltfläche bei 1.000.000
// Bildpunkten: außerhalb jedes Fensters, in der Tastaturreihenfolge aber
// mittendrin. Wer über die Vorschau tabbte, landete auf einem Ziel, das
// niemand sieht.
//
// # Und `editable` steht nicht hier
//
// Über alle `*.mse-game/game` in `M15-Magic-Pack` — weiter reicht die Zahl
// nicht — kommt `editable:` **kein einziges Mal** vor. Ein Zweig, den kein
// vorhandenes Paket betritt, wäre eine Zusage ohne Gegenprobe (§10).
func TestNurGezeichneteFelderWerdenZuFlaechen(t *testing.T) {
	plan := &render.Plan{Feldflaechen: []render.Feldflaeche{
		{Feld: "name", Rahmen: render.Rechteck{Links: 10, Oben: 20, Breite: 100, Hoehe: 30}},
		{Feld: "unplatziert", Rahmen: render.Rechteck{
			Links: 1000000, Oben: 1000000, Breite: 100, Hoehe: 30,
		}},
		{Feld: "ohne breite", Rahmen: render.Rechteck{Links: 10, Oben: 20, Breite: 0, Hoehe: 30}},
		{Feld: "ohne hoehe", Rahmen: render.Rechteck{Links: 10, Oben: 20, Breite: 100, Hoehe: 0}},
	}}

	flaechen := feldflaechen(plan, 1)

	if len(flaechen) != 1 || flaechen[0].Feld != "name" {
		t.Fatalf("erwartet nur `name`, bekommen: %+v", flaechen)
	}
}
