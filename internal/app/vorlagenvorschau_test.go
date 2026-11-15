// **Die Vorschau einer ungespeicherten Vorlage** — Werkstück 3 von Phase 3.3
// (#1164).
package app_test

import (
	"context"
	"errors"
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zeichnerprobe gibt eine Karte und merkt sich den Ersatz.
type zeichnerprobe struct {
	karte  *domain.CardRow
	fehler error

	gesehen service.Paketersatz
	gerufen int

	// gewuenscht ist die Karte, die die Werkstatt genannt hat — 0 heißt: die
	// Probekarte des Pakets (U7, #1418).
	gewuenscht int64

	// karten sind die Testkarten, die die Werkstatt zur Wahl bekommt.
	karten []*domain.CardRow
}

func (z *zeichnerprobe) Probekarten(context.Context, string) ([]*domain.CardRow, error) {
	return z.karten, z.fehler
}

func (z *zeichnerprobe) Probekarte(_ context.Context, _ string, karteID int64) (*domain.CardRow, error) {
	z.gewuenscht = karteID

	return z.karte, z.fehler
}

func (z *zeichnerprobe) MitErsatz(ersatz service.Paketersatz) (*service.Planbauer, app.Rasterer) {
	z.gesehen = ersatz
	z.gerufen++

	return service.NeuerPlanbauer(), leererRasterer{}
}

// leererRasterer gibt ein Bild von einem Pixel.
type leererRasterer struct{}

func (leererRasterer) Zeichne(
	context.Context, *render.Plan, float64, render.Guete,
) (*image.NRGBA, []render.Befund, error) {
	return image.NewNRGBA(image.Rect(0, 0, 1, 1)), nil, nil
}

func (leererRasterer) SymbolAlsBild(render.Symbol, render.Symbolmasse) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 1, 1)), ""
}

func (leererRasterer) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	return image.NewNRGBA(image.Rect(0, 0, 1, 1)), ""
}

// **Der Ersatz kommt beim Zeichner an — Paket, Datei und Inhalt.**
func TestDerErsatzKommtBeimZeichnerAn(t *testing.T) {
	probe := &zeichnerprobe{karte: &domain.CardRow{ID: 7, SetID: 1}}
	anwendung := mitZeichner(probe)

	antwort := anwendung.Vorlagenvorschau(t.Context(),
		"/pakete/magic-m15.mse-style", "style", "mse version: 2.0.0\n", 1, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.gesehen.Paket != "/pakete/magic-m15.mse-style" ||
		probe.gesehen.Datei != "style" ||
		string(probe.gesehen.Inhalt) != "mse version: 2.0.0\n" {
		t.Errorf("angekommen ist %+v", probe.gesehen)
	}
}

// **Ohne Karte gibt es keine Vorschau — und der Zeichner wird nicht gebaut.**
//
// Die zweite Hälfte ist die Zusicherung: ein Planbauer über einem
// Stilspeicher, den niemand braucht, öffnete das Paket umsonst.
func TestOhneKarteGibtEsKeineVorlagenvorschau(t *testing.T) {
	probe := &zeichnerprobe{fehler: domain.ErrNotFound}
	anwendung := mitZeichner(probe)

	antwort := anwendung.Vorlagenvorschau(t.Context(), "p", "style", "x", 1, "")
	if antwort.Fehler == nil {
		t.Fatal("die fehlende Karte kam durch")
	}

	if probe.gerufen != 0 {
		t.Errorf("der Zeichner wurde %dmal gebaut", probe.gerufen)
	}
}

// **Ohne eingerichteten Zeichner sagt es die Bindung.**
func TestOhneZeichnerSagtDieVorlagenvorschauEs(t *testing.T) {
	if antwort := app.NeueApp("").Vorlagenvorschau(
		probekontext(t), "p", "style", "x", 1, ""); antwort.Fehler == nil {
		t.Error("ohne Zeichner kam eine Vorschau")
	}
}

// **Und die Bindung am Fenster geht denselben Weg.**
func TestDieVorlagenvorschaubindungGehtDenselbenWeg(t *testing.T) {
	anwendung := mitZeichner(&zeichnerprobe{karte: &domain.CardRow{ID: 7}})

	antwort := app.NeueBindungen(anwendung).Vorlagenvorschau(t.Context(), "p", "style", "x", 1, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Vorschau == nil {
		t.Error("es kam kein Bild")
	}
}

// **Ein Fehler beim Zeichnen kommt an.**
func TestEinFehlerBeimVorlagenzeichnenKommtAn(t *testing.T) {
	anwendung := mitZeichner(&zeichnerprobe{fehler: errors.New("keine Platte")})

	if antwort := anwendung.Vorlagenvorschau(
		probekontext(t), "p", "style", "x", 1, ""); antwort.Fehler == nil {
		t.Error("der Fehler kam als Erfolg durch")
	}
}

// mitZeichner baut eine Anwendung mit Vorlagenzeichner **und Kartenleser**.
//
// **Beides, und das ist kein Zufall**: die Vorschau liest am Ende die
// Herkünfte der Karte, damit die Nennung neben dem Bild steht (#1025). Ein
// Aufbau ohne Kartenleser stürzte dort ab — gefunden von genau diesem Lauf,
// bevor er so aussah.
func mitZeichner(zeichner app.Vorlagenzeichner) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitVorlagenzeichner(zeichner)
	anwendung.MitEditor(&schreibprobe{}, karteprobe{herkuenfte: map[string]string{}},
		nil, leererRasterer{})

	return anwendung
}

// **Die Werkstatt wählt ihre Testkarte** — Werkstück U7 von Phase 5.1 (#1418):
// mit einer Kennung zeichnet die Vorschau genau diese Karte, ohne Kennung die
// Probekarte des Pakets wie bisher.
func TestDieVorlagenvorschauNimmtEineGenannteKarte(t *testing.T) {
	probe := &zeichnerprobe{karte: &domain.CardRow{ID: 7, SetID: 1}}
	anwendung := mitZeichner(probe)

	if antwort := anwendung.Vorlagenvorschau(t.Context(), "p", "style", "x", 1, "7"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.gewuenscht != 7 {
		t.Errorf("gewünscht war %d, erwartet 7", probe.gewuenscht)
	}

	if antwort := anwendung.Vorlagenvorschau(t.Context(), "p", "style", "x", 1, ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.gewuenscht != 0 {
		t.Errorf("ohne Kennung war %d gewünscht, erwartet 0", probe.gewuenscht)
	}

	if antwort := anwendung.Vorlagenvorschau(t.Context(), "p", "style", "x", 1, "abc"); antwort.Fehler == nil || antwort.Fehler.Feld != "id" {
		t.Errorf("eine unlesbare Kennung ging durch: %+v", antwort.Fehler)
	}
}
