package service_test

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Teilbefund beim Rastern erreicht die Ausfuhr** — Befund des
// Zweitblicks durch Codex vom 2026-08-21.
//
// `render.Rasterer.SymbolAlsBild` und `BildAusBaum` geben **Bild und Befund**:
// eine noch nicht gebaute Verknüpfung zeichnet das Ausgangsbild und meldet es
// dazu. Beide Wege haben die Meldung verworfen, sobald ein Bild kam — die
// Beigabe wurde geschrieben, und niemand erfuhr, dass sie halb falsch aussieht.
//
// **Ein Befund und kein Fehler** (§8): eine Datei, die zu 90 % stimmt, ist
// brauchbar; eine fehlende ist es nicht. Dieselbe Abwägung wie bei einer
// Kartenebene.
type meldenderRasterer struct{ meldung string }

func (m meldenderRasterer) SymbolAlsBild(
	render.Symbol, render.Symbolmasse,
) (bild *image.NRGBA, befund string) {
	return image.NewNRGBA(image.Rect(0, 0, 5, 5)), m.meldung
}

func (m meldenderRasterer) BildAusBaum(
	render.Bildbaum, int, int,
) (bild *image.NRGBA, befund string) {
	return image.NewNRGBA(image.Rect(0, 0, 5, 5)), m.meldung
}

func TestEinTeilbefundBeimRasternErreichtDieAusfuhr(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, skript string }{
		{
			name:   "über symbols_to_html",
			skript: `symbols_to_html("W", symbol_font: "mana")`,
		},
		{
			name:   "über write_image_file mit einem Bildwert",
			skript: `write_image_file("b.png", file: "x.png")`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ausfuehrer := service.NeuerVorlagenausfuehrer()
			ausfuehrer.MitSymbolen(manaschrift(),
				meldenderRasterer{meldung: "die Verknüpfung rotate ist noch nicht gebaut"})

			_, befunde, err := ausfuhrprobe{
				ausfuehrer: ausfuehrer,
				mittel: service.Ausfuhrmittel{
					Beigaben:          &beigabenmitschrift{},
					Symbolzeichnungen: &symbolgestaltprobe{},
					Setpaket:          "/Orte/Probe.mse-set",
					Verzeichnis:       "ausfuhr-files",
				},
			}.Fuehre(t.Context(),
				vorlageMitFassung(fall.skript), eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			gefunden := false

			for _, befund := range befunde {
				if strings.Contains(befund.Meldung, "rotate") {
					gefunden = true
				}
			}

			if !gefunden {
				t.Errorf("die Meldung fehlt in den Befunden: %v", befunde)
			}
		})
	}
}
