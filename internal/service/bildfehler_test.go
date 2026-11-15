// Die Fehlerwege der Bildausgabe.
//
// **Eine eigene Datei** (AGENTS.md §2), und der Schnitt folgt der Frage: dort
// steht, was ein geschriebenes Bild ausmacht, hier, was geschieht, wenn es
// nicht dazu kommt.
package service_test

import (
	"context"
	"errors"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Fehlerwege, jeder mit seiner eigenen Ursache.
func TestSchreibfehler(t *testing.T) {
	faelle := []struct {
		name    string
		ausfuhr func(t *testing.T) *service.Bildausfuhr
		ort     func(t *testing.T) string
		format  service.Bildformat
	}{
		{
			name: "die Karte gibt es nicht",
			ausfuhr: func(t *testing.T) *service.Bildausfuhr {
				t.Helper()

				return service.NeueBildausfuhr(
					kartenDouble{fehlt: true}, planDouble{}, render.NeuerRasterer(nil))
			},
		},
		{
			name: "der Plan lässt sich nicht bauen",
			ausfuhr: func(t *testing.T) *service.Bildausfuhr {
				t.Helper()

				return bildausfuhr(t, planDouble{fehler: errors.New("kein Stil")})
			},
		},
		{
			name: "der Plan ist nicht zeichenbar",
			ausfuhr: func(t *testing.T) *service.Bildausfuhr {
				t.Helper()

				return bildausfuhr(t, planDouble{breite: -1, hoehe: -1})
			},
		},
		{
			name:   "das Format gibt es nicht",
			format: service.Bildformat("webp"),
		},
		{
			name: "das Ziel lässt sich nicht anlegen",
			ort: func(t *testing.T) string {
				t.Helper()

				// Eine Datei da, wo ein Verzeichnis hin müsste.
				ort := filepath.Join(t.TempDir(), "imweg")
				if err := os.WriteFile(ort, []byte("x"), 0o600); err != nil {
					t.Fatalf("anlegen: %v", err)
				}

				return filepath.Join(ort, "tiefer")
			},
		},
		{
			name: "das Ziel lässt sich nicht schreiben",
			ort: func(t *testing.T) string {
				t.Helper()

				// Ein Verzeichnis da, wo die Datei hin müsste.
				ort := t.TempDir()
				if err := os.MkdirAll(filepath.Join(ort, "001.png"), 0o750); err != nil {
					t.Fatalf("anlegen: %v", err)
				}

				return ort
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ausfuhr := bildausfuhr(t, planDouble{})
			if fall.ausfuhr != nil {
				ausfuhr = fall.ausfuhr(t)
			}

			ort := t.TempDir()
			if fall.ort != nil {
				ort = fall.ort(t)
			}

			format := service.FormatPNG
			if fall.format != "" {
				format = fall.format
			}

			if _, err := ausfuhr.Schreibe(t.Context(), 1, ort, "001.png", format, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// Ein abgebrochener Lauf schreibt nichts mehr.
func TestEinAbgebrochenerLaufSchreibtNicht(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	ort := t.TempDir()
	ziel := filepath.Join(ort, "001.png")

	if _, err := bildausfuhr(t, planDouble{}).
		Schreibe(ctx, 1, ort, "001.png", service.FormatPNG, 0, domain.BildkonfliktUeberschreiben,
			render.GueteMittel); err == nil {
		t.Fatal("kein Fehler")
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// rasterDouble gibt ein vorgegebenes Bild zurück.
//
// **Damit die Packfehler erreichbar sind.** `png.Encode` weist ein Bild ohne
// Fläche ab, `jpeg.Encode` eines ab 65 536 Bildpunkten Breite — beides über den
// echten Rasterer nur mit einer Größe zu erreichen, die den Arbeitsspeicher
// sprengt.
type rasterDouble struct{ bild *image.NRGBA }

func (r rasterDouble) Zeichne(context.Context, *render.Plan, float64, render.Guete) (
	*image.NRGBA, []render.Befund, error,
) {
	return r.bild, nil, nil
}

// Ein Bild, das sich nicht packen lässt, wird gemeldet statt geschrieben.
func TestEinBildDasSichNichtPackenLaesst(t *testing.T) {
	faelle := []struct {
		name   string
		bild   *image.NRGBA
		format service.Bildformat
	}{
		{
			name:   "ohne Fläche als PNG",
			bild:   image.NewNRGBA(image.Rect(0, 0, 0, 0)),
			format: service.FormatPNG,
		},
		{
			// 65 536 Bildpunkte breit und einen hoch: ein Viertelmegabyte, und
			// mehr, als das JPG-Format in seinem Kopf unterbringt.
			name:   "zu breit als JPG",
			bild:   image.NewNRGBA(image.Rect(0, 0, 65536, 1)),
			format: service.FormatJPG,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ausfuhr := service.NeueBildausfuhr(
				kartenDouble{leserDouble: karten("Feuerball")},
				planDouble{}, rasterDouble{bild: fall.bild})

			ort := t.TempDir()
			ziel := filepath.Join(ort, "001.png")

			if _, err := ausfuhr.Schreibe(t.Context(), 1, ort, "001.png", fall.format, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); err == nil {
				t.Fatal("kein Fehler")
			}

			if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
				t.Error("es wurde trotzdem eine Datei angelegt")
			}
		})
	}
}

// **Eine Zielbreite hat eine Obergrenze, und sie steht vor dem Rastern.**
//
// Wer 100 000 tippt — versehentlich oder aus Neugier —, bekäme bei einer Karte
// von 375 × 523 Punkten ein Bild mit 14 Milliarden Bildpunkten: 55 Gigabyte,
// angefordert in einem Stück. Der Rechner ist danach nicht langsam, sondern
// weg.
func TestEineZuGrosseZielbreiteWirdAbgewiesen(t *testing.T) {
	_, err := bildausfuhr(t, planDouble{}).
		Schreibe(t.Context(), 1, t.TempDir(), "riesig.png", service.FormatPNG, 100_000, domain.BildkonfliktUeberschreiben, render.GueteMittel)
	if err == nil {
		t.Fatal("kein Fehler — der Arbeitsspeicher wäre alle")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler: %v, erwartet ErrValidation", err)
	}
}

// **Ein Dateiname mit Pfadanteil wird abgewiesen.**
//
// Der Name kommt aus `Namen` und ist dort gesäubert — aber er reist durch die
// Oberfläche, und was von dort zurückkommt, ist eine Eingabe. Wer `verzeichnis`
// und `datei` erst im Aufrufer zusammensetzt, verlässt sich darauf, dass jeder
// Aufrufer das prüft; hier steht es an **einer** Stelle.
func TestEinDateinameMitPfadanteilWirdAbgewiesen(t *testing.T) {
	for _, datei := range []string{"../oben.png", "tiefer/001.png", "", "."} {
		t.Run(datei, func(t *testing.T) {
			_, err := bildausfuhr(t, planDouble{}).
				Schreibe(t.Context(), 1, t.TempDir(), datei, service.FormatPNG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel)
			if err == nil {
				t.Fatal("kein Fehler")
			}

			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler: %v, erwartet ErrValidation", err)
			}
		})
	}
}
