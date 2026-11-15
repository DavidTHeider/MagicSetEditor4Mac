// **SVG ist ein Ausgabeformat wie die vier anderen** —
// #852.
//
// Der Zweitblick über Phase P hat angestrichen, dass `svg.AusPlan` außerhalb
// seiner eigenen Tests **nirgends** gerufen wurde: das Paket stand da, und
// niemand konnte eine SVG erzeugen. Hier steht der Weg von der Karte zur
// Datei.
package service_test

import (
	"context"
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// baeumeDouble löst jeden Bildbaum in ein winziges Bild auf.
type baeumeDouble struct{}

func (baeumeDouble) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 2, 2)), ""
}

// vektorausfuhr gibt eine Ausfuhr, die auch SVG kann.
func vektorausfuhr(t *testing.T) *service.Bildausfuhr {
	t.Helper()

	return service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball")},
		planReich{}, render.NeuerRasterer(nil),
	).MitVektor(svg.Quellen{Baeume: baeumeDouble{}})
}

func TestSVGTraegtSeineEndung(t *testing.T) {
	t.Parallel()

	if endung := service.FormatSVG.Endung(); endung != ".svg" {
		t.Errorf("SVG gibt %q, erwartet war %q", endung, ".svg")
	}
}

// **Eine Karte kommt als SVG heraus, und es ist eines.**
func TestEineKarteWirdEineSVGDatei(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	geschrieben, fehler := vektorausfuhr(t).Schreibe(t.Context(), 1, ordner,
		"Feuerball.svg", service.FormatSVG, 0, domain.BildkonfliktNummerierenNeu, render.GueteMittel)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	roh, fehler := os.ReadFile(filepath.Join(ordner, "Feuerball.svg")) //#nosec G304 -- Testverzeichnis
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	inhalt := string(roh)

	if !strings.HasPrefix(inhalt, "<svg") || !strings.HasSuffix(inhalt, "</svg>") {
		t.Errorf("die Datei ist kein SVG:\n%.120s", inhalt)
	}

	// **Und es entsteht aus dem Plan.** Sechs Ebenen, davon fünf Flächen —
	// wäre hier ein gewickeltes Rasterbild, stünde kein einziges `<rect>`
	// darin.
	if !strings.Contains(inhalt, "<rect") {
		t.Error("die Flächen stehen nicht als Geometrie im SVG")
	}

	if geschrieben.Pfad != filepath.Join(ordner, "Feuerball.svg") {
		t.Errorf("der Pfad ist %q", geschrieben.Pfad)
	}
}

// **Die Zielbreite wirkt auch auf eine SVG.**
func TestDieZielbreiteWirktAufDieSVG(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	geschrieben, fehler := vektorausfuhr(t).Schreibe(t.Context(), 1, ordner,
		"breit.svg", service.FormatSVG, 750, domain.BildkonfliktNummerierenNeu, render.GueteMittel)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	roh, fehler := os.ReadFile(filepath.Join(ordner, "breit.svg")) //#nosec G304 -- Testverzeichnis
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if !strings.Contains(string(roh), `width="750"`) {
		t.Errorf("die Zielbreite steht nicht am SVG:\n%.120s", roh)
	}

	// **Gemeldet wird die bestellte Größe**, nicht die des Plans — sonst sagte
	// die Oberfläche etwas anderes, als in der Datei steht.
	if geschrieben.Breite != 750 {
		t.Errorf("gemeldet werden %d Punkte Breite, erwartet waren 750", geschrieben.Breite)
	}
}

// **Die Befunde des Schreibers kommen bis zum Aufrufer.**
//
// Das war der zweite Teil des Befunds: `AusPlan` gab `(string, error)`, ließ
// eine unlesbare Ebene weg und sagte nichts. Eine SVG ohne Kartenrahmen
// entstand erfolgreich.
func TestDieBefundeDerSVGKommenAn(t *testing.T) {
	t.Parallel()

	ohneQuellen := service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball")},
		planMitBild{}, render.NeuerRasterer(nil))

	geschrieben, fehler := ohneQuellen.Schreibe(t.Context(), 1, t.TempDir(),
		"ohne.svg", service.FormatSVG, 0, domain.BildkonfliktNummerierenNeu,
		render.GueteMittel)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(geschrieben.Befunde) != 1 ||
		!strings.Contains(geschrieben.Befunde[0].Meldung, "Bildquelle") {
		t.Errorf("der Befund fehlt: %v", geschrieben.Befunde)
	}

	// **Und er nennt seine Datei**, so wie jeder andere Befund der Ausfuhr.
	if geschrieben.Befunde[0].Datei != "ohne.svg" {
		t.Errorf("der Befund nennt die Datei %q", geschrieben.Befunde[0].Datei)
	}
}

// planMitBild gibt einen Plan mit einer Bildebene.
type planMitBild struct{}

func (planMitBild) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{
		Breite: 100, Hoehe: 140,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Feld:    "rahmen",
			Rahmen:  render.Rechteck{Breite: 100, Hoehe: 140},
			Bild:    render.Blatt("/probe.mse-style/rahmen.png"),
			Deckung: 1,
		}},
	}, nil, nil
}

// **Ein Namenskonflikt gilt für SVG wie für jedes andere Format.**
func TestEinNamenskonfliktGiltAuchFuerSVG(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ausfuhr := vektorausfuhr(t)

	if _, fehler := ausfuhr.Schreibe(t.Context(), 1, ordner, "gleich.svg",
		service.FormatSVG, 0, domain.BildkonfliktNummerierenNeu, render.GueteMittel); fehler != nil {
		t.Fatalf("erstes Schreiben: %v", fehler)
	}

	geschrieben, fehler := ausfuhr.Schreibe(t.Context(), 1, ordner, "gleich.svg",
		service.FormatSVG, 0, domain.BildkonfliktAltbehalten, render.GueteMittel)
	if fehler != nil {
		t.Fatalf("zweites Schreiben: %v", fehler)
	}

	if !geschrieben.Uebersprungen {
		t.Error("die vorhandene Datei wurde überschrieben")
	}
}

// **`PackeBild` weist SVG mit dem Grund ab, nicht mit einem Achselzucken.**
//
// SVG entsteht aus dem Plan und nicht aus Bildpunkten; ein Kodierer dafür wäre
// der Name ohne die Sache. Der Zweig sagt es, statt „ist kein Bildformat" zu
// melden — was für ein Format, das die Oberfläche anbietet, ratlos macht.
func TestPackeBildWeistSVGMitGrundAb(t *testing.T) {
	t.Parallel()

	_, fehler := service.PackeBild(durchsichtigeEcke(), service.FormatSVG)
	if fehler == nil || !strings.Contains(fehler.Error(), "Zeichenplan") {
		t.Errorf("Fehler %v — erwartet war der Grund", fehler)
	}
}

// **Die beiden Fehlerwege der SVG-Ausgabe.**
//
// Sie sind dieselben wie beim Rasterweg — ein Abbruch und ein Ziel, das sich
// nicht schreiben lässt —, und sie stehen hier, weil `alsSVG` sie ein zweites
// Mal betritt. Ein Fehlerweg, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht (§3).
func TestDieFehlerwegeDerSVGAusgabe(t *testing.T) {
	t.Parallel()

	t.Run("ein abgebrochener Lauf schreibt nichts", func(t *testing.T) {
		t.Parallel()

		ctx, abbrechen := context.WithCancel(t.Context())
		abbrechen()

		ordner := t.TempDir()

		if _, fehler := vektorausfuhr(t).Schreibe(ctx, 1, ordner, "ab.svg",
			service.FormatSVG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); fehler == nil {
			t.Fatal("kein Fehler")
		}

		if _, fehler := os.Stat(filepath.Join(ordner, "ab.svg")); !errors.Is(fehler, os.ErrNotExist) {
			t.Error("es wurde trotzdem geschrieben")
		}
	})

	t.Run("das Ziel lässt sich nicht schreiben", func(t *testing.T) {
		t.Parallel()

		// Ein Verzeichnis da, wo die Datei hin müsste.
		ordner := t.TempDir()
		if fehler := os.MkdirAll(filepath.Join(ordner, "imweg.svg"), 0o750); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		if _, fehler := vektorausfuhr(t).Schreibe(t.Context(), 1, ordner, "imweg.svg",
			service.FormatSVG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); fehler == nil {
			t.Error("kein Fehler")
		}
	})
}
