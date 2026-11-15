package app

import (
	"context"
	"image"
	"io"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Wahlbild, das sich nicht packen lässt, wird übergangen und gemeldet**
// (#1225).
//
// **Nicht abgebrochen** — dieselbe Unterscheidung wie an der Symbolpalette:
// eine Vorschau **ist** ihr Bild, eine Wahlreihe sind drei Kacheln. Eine
// Kachel ohne Bild steht mit ihrem Namen da, und daneben steht, warum (§8).
//
// **Ein interner Test**, weil die Naht [App.packen] intern ist: mit einem
// `bytes.Buffer` kann `png.Encode` nicht scheitern, und im Betrieb schreibt
// irgendwann jemand auf etwas, das voll ist.
func TestEineGescheiterteKodierungLaesstDieWahlreiheStehen(t *testing.T) {
	anwendung := NeueApp("")
	anwendung.MitEditor(nil, nil, einWahlbild{}, rastererProbe{})
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }

	antwort := anwendung.Stilwahlbilder(t.Context(), "7", "tap symbol")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 0 {
		t.Fatalf("%d Bilder: %+v", len(antwort.Bilder), antwort.Bilder)
	}

	if len(antwort.Befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(antwort.Befunde), antwort.Befunde)
	}
}

// einWahlbild gibt genau einen Baum.
type einWahlbild struct{ planbauerProbe }

func (einWahlbild) Optionsbilder(context.Context, int64, string) (
	[]service.Optionsbild, []domain.Befund, error,
) {
	return []service.Optionsbild{{Wahl: "modern", Baum: render.Blatt("/paket/t.png")}}, nil, nil
}
