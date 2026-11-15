//go:build fremdpakete

// Ein Handgriff für die Sichtprüfung (AGENTS.md §3): schreibt die
// Referenzkarten als SVG nach `MSE_SVG_AUSGABE`.
//
// **Kein Wächter.** Er sichert nichts zu; er legt hin, was ein Mensch ansehen
// soll. Ohne gesetzte Umgebungsvariable tut er nichts.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

func TestSchreibeReferenzkartenAlsSVG(t *testing.T) {
	ziel := os.Getenv("MSE_SVG_AUSGABE")
	if ziel == "" {
		t.Skip("MSE_SVG_AUSGABE ist nicht gesetzt")
	}

	referenzen, wurzel := os.Getenv("MSE_REFERENZBILDER"), os.Getenv("MSE_FREMDPAKETE")
	if referenzen == "" || wurzel == "" {
		t.Skip("MSE_REFERENZBILDER oder MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(referenzen)
	if err != nil {
		t.Fatalf("Referenzen lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		name, istKarte := strings.CutSuffix(eintrag.Name(), ".karte")
		if !istKarte {
			continue
		}

		karte, stilname, _ := referenzkarte(t, filepath.Join(referenzen, eintrag.Name()))

		suche := schriften.Neue(schriften.Orte()).MitPaketen(filepath.Dir(wurzel))
		bilder := msepaket.NeueBildquelle(wurzel)

		plan, _, err := service.NeuerPlanbauer().
			MitSchriften(suche).MitBildern(bilder).Stilplan(
			probekontext(t), karte, fremdvorlage(t, fremdstil(t, wurzel, stilname), wurzel))
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		rasterer := render.NeuerRasterer(bilder).
			MitSchriften(suche).MitErsatzschrift(suche.Ersatz)

		geschrieben, _, err := svg.AusPlan(probekontext(t), plan, svg.Quellen{
			Baeume: rasterer, Raster: rasterer, Schriften: suche, Symbole: bilder,
		})
		if err != nil {
			t.Fatalf("SVG: %v", err)
		}

		pfad := filepath.Join(ziel, name+".svg")
		// **Der Pfad kommt aus einer Umgebungsvariablen dieses Rechners** —
		// der Handgriff ist für die Sichtprüfung von Hand da.
		if err := os.WriteFile(pfad, []byte(geschrieben), 0o600); err != nil { //#nosec G703 -- ein Ziel aus MSE_SVG_AUSGABE
			t.Fatalf("schreiben: %v", err)
		}

		t.Logf("%s (%d Bytes)", pfad, len(geschrieben))
	}
}
