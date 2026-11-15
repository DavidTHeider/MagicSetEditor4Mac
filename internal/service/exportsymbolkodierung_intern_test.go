package service

import (
	"context"
	"errors"
	"image"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine gescheiterte Kodierung wird gemeldet** (§8) —
// #658.
//
// Mit einem `bytes.Buffer` kann `png.Encode` nicht scheitern; deshalb steht
// die Kodierung als Naht am [ausfuhrlauf] und nicht als Aufruf im Code. Ohne
// sie wäre dieser Zweig eine Behauptung darüber, was im Ernstfall geschieht —
// dieselbe Bauart wie bei `App.packen`.
//
// **Der Test steht im Paket**, weil die Naht dorthin gehört: sie ist kein
// Angebot an den Aufrufer, sondern ein Zugang für die Prüfung.
func TestEineGescheiterteSymbolkodierungWirdGemeldet(t *testing.T) {
	t.Parallel()

	lauf := &ausfuhrlauf{
		ausfuehrer: NeuerVorlagenausfuehrer().
			MitSymbolen(einzelschrift{}, ganzflaechigerRasterer{}),
		mittel: Ausfuhrmittel{
			Beigaben:          stummerBeigabenschreiber{},
			Symbolzeichnungen: stummeSymbolquelle{},
			Setpaket:          "/Orte/Probe.mse-set",
		},
		gelegt: map[string][2]int{},
		packen: func(io.Writer, image.Image) error { return errors.New("kaputt") },
	}

	_, err := lauf.wandler(t.Context(), "R", "mana", 12)
	if err == nil || !strings.Contains(err.Error(), "kaputt") {
		t.Fatalf("Fehler = %v", err)
	}

	// **Und derselbe Weg beim Bildwert** (#660): auch `write_image_file`
	// kodiert über diese Naht.
	_, err = lauf.bildwertSchreiben(msescript.Dateibild("b.png"), "b.png", 0, 0)
	if err == nil || !strings.Contains(err.Error(), "kaputt") {
		t.Fatalf("Bildwert: Fehler = %v", err)
	}
}

type stummeSymbolquelle struct{}

func (stummeSymbolquelle) Symbolgestalt(string) msepaket.Symbolzeichnung {
	return msepaket.Standardsymbol()
}

type einzelschrift struct{}

func (einzelschrift) Symbolschrift(string) (msepaket.Symbolschrift, bool) {
	return msepaket.Symbolschrift{
		Name: "mana", Bildgroesse: 100,
		Symbole: []msepaket.Symbol{{Code: "R", Bild: "r.png"}},
	}, true
}

type ganzflaechigerRasterer struct{}

func (ganzflaechigerRasterer) BildAusBaum(
	render.Bildbaum, int, int,
) (bild *image.NRGBA, befund string) {
	return image.NewNRGBA(image.Rect(0, 0, 4, 4)), ""
}

func (ganzflaechigerRasterer) SymbolAlsBild(
	render.Symbol, render.Symbolmasse,
) (bild *image.NRGBA, befund string) {
	return image.NewNRGBA(image.Rect(0, 0, 4, 4)), ""
}

type stummerBeigabenschreiber struct{}

func (stummerBeigabenschreiber) AusVorlage(string) ([]byte, error) { return nil, nil }

func (stummerBeigabenschreiber) Lege(datei string, _ []byte) (string, error) {
	return datei, nil
}

func (stummerBeigabenschreiber) LegeSet(context.Context, string) (string, error) {
	return "", nil
}
