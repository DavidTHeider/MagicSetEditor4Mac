package app_test

import (
	"context"
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Wahlbilder einer Stiloption erreichen das Fenster** —
// #1225.
//
// MSE zeigt in der Auswahlliste je Wahl ein Vorschaubild
// (`src/gui/value/choice.cpp`, Zeilen 145–185, Fassung f9d9356 vom
// 2021-07-06); dieses Fenster zeigte drei Wörter, wo MSE drei Tapsymbole
// zeigt.
func TestDieWahlbilderEinerStiloptionKommenAlsBilder(t *testing.T) {
	t.Parallel()

	antwort := wahlbildanwendung(wahlbildrasterer{}).
		Stilwahlbilder(probekontext(t), "7", "tap symbol")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 1 {
		t.Fatalf("%d Bilder: %+v", len(antwort.Bilder), antwort.Bilder)
	}

	if antwort.Bilder[0].Wahl != "modern" {
		t.Errorf("Wahl: %q", antwort.Bilder[0].Wahl)
	}

	if !strings.HasPrefix(antwort.Bilder[0].Bild, "data:image/png;base64,") {
		t.Errorf("Bild ist keine Datenadresse: %.40q", antwort.Bilder[0].Bild)
	}
}

// **Eine Kennung, die keine ist, wird als Befund abgewiesen** — und nicht als
// leere Liste. Eine leere Liste sähe aus wie „diese Option hat keine Bilder".
func TestStilwahlbilderWeisenEineUngueltigeKennungAb(t *testing.T) {
	t.Parallel()

	antwort := wahlbildanwendung(wahlbildrasterer{}).
		Stilwahlbilder(probekontext(t), "keine", "tap symbol")
	if antwort.Fehler == nil {
		t.Fatalf("kein Befund: %+v", antwort)
	}
}

// **Ein Baum, den der Rasterer nicht auflösen kann, fällt weg und wird
// gemeldet.**
//
// Ein leeres `<img>` in einer Auswahlreihe sähe aus wie ein Fehler der
// Anwendung; die Wahl bleibt dann bei ihrem Namen (§8).
func TestEinUnaufloesbaresWahlbildFaelltWegUndWirdGemeldet(t *testing.T) {
	t.Parallel()

	antwort := wahlbildanwendung(leererWahlbildrasterer{}).
		Stilwahlbilder(probekontext(t), "7", "tap symbol")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 0 {
		t.Fatalf("%d Bilder: %+v", len(antwort.Bilder), antwort.Bilder)
	}

	if len(antwort.Befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(antwort.Befunde), antwort.Befunde)
	}
}

// **Ein Befund des Planbaus reist mit dem Bild.**
func TestDieBefundeDesWahlbildbausKommenMit(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEditor(nil, nil, wahlbildbauer{meldung: "die Angabe ließ sich nicht ausrechnen"},
		wahlbildrasterer{})

	antwort := anwendung.Stilwahlbilder(t.Context(), "7", "tap symbol")
	if len(antwort.Befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(antwort.Befunde), antwort.Befunde)
	}
}

// wahlbildanwendung baut eine Anwendung mit dem gegebenen Rasterer.
func wahlbildanwendung(rasterer app.Rasterer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(nil, nil, wahlbildbauer{}, rasterer)

	return anwendung
}

// wahlbildbauer gibt einen Baum je Wahl.
type wahlbildbauer struct {
	behelfsplanbauer

	meldung string
	fehler  error
}

func (w wahlbildbauer) Optionsbilder(_ context.Context, setID int64, option string) (
	[]service.Optionsbild, []domain.Befund, error,
) {
	var befunde []domain.Befund
	if w.meldung != "" {
		befunde = []domain.Befund{{Ort: option, Meldung: w.meldung}}
	}

	if w.fehler != nil {
		return nil, nil, w.fehler
	}

	if setID != 7 || option != "tap symbol" {
		return nil, befunde, nil
	}

	return []service.Optionsbild{{Wahl: "modern", Baum: render.Blatt("/paket/t.png")}},
		befunde, nil
}

// wahlbildrasterer malt ein Bild, das sich packen lässt.
type wahlbildrasterer struct{ leererRasterer }

func (wahlbildrasterer) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 4, 4)), ""
}

// leererWahlbildrasterer löst keinen Baum auf und sagt, warum.
type leererWahlbildrasterer struct{ leererRasterer }

func (leererWahlbildrasterer) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	return nil, "die Datei fehlt"
}

// **Ein Fehler des Planbaus wird übersetzt und nicht verschluckt.**
func TestEinFehlerDesWahlbildbausErreichtDieOberflaeche(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEditor(nil, nil,
		wahlbildbauer{fehler: errors.New("planmäßig gescheitert")}, wahlbildrasterer{})

	if antwort := anwendung.Stilwahlbilder(t.Context(), "7", "tap symbol"); antwort.Fehler == nil {
		t.Fatalf("kein Befund: %+v", antwort)
	}
}
