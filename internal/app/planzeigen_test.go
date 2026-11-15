//go:build fremdpakete

// Ein Handgriff für die Diagnose: druckt den Zeichenplan einer Referenzkarte.
// **Kein Wächter** — ohne `MSE_PLAN_KARTE` tut er nichts.
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestZeigeDenPlan(t *testing.T) {
	name := os.Getenv("MSE_PLAN_KARTE")
	if name == "" {
		t.Skip("MSE_PLAN_KARTE ist nicht gesetzt")
	}

	referenzen, wurzel := os.Getenv("MSE_REFERENZBILDER"), os.Getenv("MSE_FREMDPAKETE")
	if referenzen == "" || wurzel == "" {
		t.Skip("MSE_REFERENZBILDER oder MSE_FREMDPAKETE ist nicht gesetzt")
	}

	karte, stilname, _ := referenzkarte(t, filepath.Join(referenzen, name+".karte"))

	suche := schriften.Neue(schriften.Orte()).MitPaketen(filepath.Dir(wurzel))
	bilder := msepaket.NeueBildquelle(wurzel)

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(suche).MitBildern(bilder).Stilplan(
		probekontext(t), karte, fremdvorlage(t, fremdstil(t, wurzel, stilname), wurzel))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	t.Logf("Stil %s — %d Ebenen, %d Planbefunde", stilname, len(plan.Ebenen), len(befunde))

	for stelle, ebene := range plan.Ebenen {
		t.Logf("%2d %-10s feld=%-18q misch=%-10q deckung=%.2f maske=%q bild=%q",
			stelle, ebene.Art, ebene.Feld, ebene.Mischart, ebene.Deckung,
			ebene.Maske.Beschreibung(), ebene.Bild.Beschreibung())
	}

	rasterer := render.NeuerRasterer(bilder).MitSchriften(suche).MitErsatzschrift(suche.Ersatz)

	_, rasterbefunde, err := rasterer.Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	for _, befund := range rasterbefunde {
		t.Logf("Rasterbefund: %s", befund)
	}
}
