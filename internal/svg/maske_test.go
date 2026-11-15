// **Die Maske einer Ebene** —
// #852.
//
// Sie stand im Plan, der Rasterer wandte sie an, der erste SVG-Schreiber las
// sie nicht. **Der rote Kanal wird zur Leuchtdichte**: MSE nimmt Rot als
// Deckung (`src/gfx/mask_image.cpp`, Zeilen 41–44, Fassung f9d9356 vom
// 2021-07-06), SVG kennt für `<mask>` nur Leuchtdichte und Alpha — Rot in alle
// drei Kanäle zu kopieren macht beides gleich.
package svg_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// TestEineMaskeWirdEineMask prüft, dass die Maske ankommt.
//
// **Der rote Kanal wird zur Leuchtdichte.** MSE nimmt Rot als Deckung
// (`src/gfx/mask_image.cpp`, Zeilen 41–44, Fassung f9d9356 vom 2021-07-06);
// SVG kennt für `<mask>` nur
// Leuchtdichte und Alpha. Rot in alle drei Kanäle zu kopieren macht beides
// gleich.
func TestEineMaskeWirdEineMask(t *testing.T) {
	t.Parallel()

	mitMaske := flaeche()
	mitMaske.Maske = render.Blatt("/probe.mse-style/maske.png")

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(mitMaske),
		svg.Quellen{Baeume: baumprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	for _, teil := range []string{"<mask id=", `mask="url(#`, "<defs>"} {
		if !strings.Contains(geschrieben, teil) {
			t.Errorf("%q fehlt im SVG:\n%s", teil, geschrieben)
		}
	}
}

// TestEineUnlesbareMaskeWirdGenannt prüft den Befund.
func TestEineUnlesbareMaskeWirdGenannt(t *testing.T) {
	t.Parallel()

	mitMaske := flaeche()
	mitMaske.Maske = render.Blatt("/probe.mse-style/maske.png")

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(mitMaske),
		svg.Quellen{Baeume: baumprobe{fehlt: true}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "die Maske") {
		t.Errorf("kein Befund über die unlesbare Maske: %v", befunde)
	}

	// **Und die Ebene wird trotzdem gezeichnet** — ohne Maske ist sie zu
	// groß, aber sichtbar falsch schlägt unsichtbar fehlend (§8).
	if !strings.Contains(geschrieben, "<rect") {
		t.Error("die Ebene fehlt, obwohl nur ihre Maske fehlte")
	}
}

// TestOhneBildquelleWirdDieMaskeGenannt prüft den zweiten Weg dorthin.
func TestOhneBildquelleWirdDieMaskeGenannt(t *testing.T) {
	t.Parallel()

	mitMaske := flaeche()
	mitMaske.Maske = render.Blatt("/probe.mse-style/maske.png")

	_, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(mitMaske), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Maske") {
		t.Errorf("kein Befund über die fehlende Bildquelle der Maske: %v", befunde)
	}
}

// TestEineMaskeAnEinerGedrehtenEbeneSitztAmUrsprung prüft, dass sie
// mitgedreht wird und nicht doppelt verschoben.
func TestEineMaskeAnEinerGedrehtenEbeneSitztAmUrsprung(t *testing.T) {
	t.Parallel()

	gedreht := flaeche()
	gedreht.Winkel = 180
	gedreht.Maske = render.Blatt("/probe.mse-style/maske.png")

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gedreht),
		svg.Quellen{Baeume: baumprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	inMaske := regexp.MustCompile(`<mask id="[^"]*"><image[^>]*x="0" y="0"`)
	if !inMaske.MatchString(geschrieben) {
		t.Errorf("die Maske sitzt nicht am Ursprung:\n%s", geschrieben)
	}
}

// TestEinKodiererfehlerInDerMaskeWirdGenannt betritt den letzten Rückweg.
func TestEinKodiererfehlerInDerMaskeWirdGenannt(t *testing.T) {
	t.Parallel()

	mitMaske := flaeche()
	mitMaske.Maske = render.Blatt("/probe.mse-style/maske.png")

	_, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(mitMaske),
		svg.Quellen{Baeume: baumprobe{ohneFlaeche: true}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "einbetten") {
		t.Errorf("der Kodiererfehler der Maske wurde verschluckt: %v", befunde)
	}
}
