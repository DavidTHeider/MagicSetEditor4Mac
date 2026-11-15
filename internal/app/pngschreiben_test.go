//go:build fremdpakete

// Ein Handgriff für die Sichtprüfung: schreibt die Referenzkarten als PNG nach
// `MSE_PNG_AUSGABE`. **Kein Wächter** — ohne gesetzte Umgebungsvariable tut er
// nichts.
package app_test

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestSchreibeReferenzkartenAlsPNG(t *testing.T) {
	ziel := os.Getenv("MSE_PNG_AUSGABE")
	if ziel == "" {
		t.Skip("MSE_PNG_AUSGABE ist nicht gesetzt")
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

		bild, _, err := render.NeuerRasterer(bilder).
			MitSchriften(suche).MitErsatzschrift(suche.Ersatz).
			Zeichne(probekontext(t), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		var gepackt bytes.Buffer
		if err := png.Encode(&gepackt, bild); err != nil {
			t.Fatalf("packen: %v", err)
		}

		// **Der Pfad kommt aus einer Umgebungsvariablen dieses Rechners** —
		// der Handgriff ist für die Sichtprüfung von Hand da.
		if err := os.WriteFile(filepath.Join(ziel, name+".png"), gepackt.Bytes(), 0o600); err != nil { //#nosec G304,G703 -- ein Ziel aus MSE_PNG_AUSGABE
			t.Fatalf("schreiben: %v", err)
		}
	}
}
