package app_test

// **Die Bindung mit ihrem Dateidialog** —
// #1340.
//
// **Der Dialog ist eine Naht** — `MitDateiwahl` stopft sie. Was hier geprüft
// wird, ist genau die Trennung, die es der Kommandozeile erlaubt, dieselbe
// Fähigkeit anzubieten: die Bindung fragt, die Anwendung rechnet.

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kreisPNG schreibt ein PNG mit einem gefüllten Kreis und gibt seinen Pfad.
func kreisPNG(t *testing.T) string {
	t.Helper()

	const mass = 80

	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))
	mitte := float64(mass) / 2

	for y := range mass {
		for x := range mass {
			ton := uint8(255)
			if math.Hypot(float64(x)-mitte, float64(y)-mitte) <= float64(mass)*0.35 {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	pfad := filepath.Join(t.TempDir(), "kreis.png")

	datei, err := os.Create(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	if err := png.Encode(datei, gemalt); err != nil {
		t.Fatalf("das PNG schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("die Datei schließen: %v", err)
	}

	return pfad
}

// mitWahl baut Bindungen, deren Dateidialog `gewaehlt` antwortet.
func mitWahl(t *testing.T, gewaehlt string, klage error) *app.Bindungen {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return gewaehlt, klage
	})

	return app.NeueBindungen(anwendung)
}

// TestDieBindungZeichnetDasGewaehlteBildNach ist der Weg im Fenster.
func TestDieBindungZeichnetDasGewaehlteBildNach(t *testing.T) {
	t.Parallel()

	antwort := mitWahl(t, kreisPNG(t), nil).SymbolNachzeichnen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 1 {
		t.Fatalf("%d Teile, erwartet 1", len(antwort.Teile))
	}
}

// **Ein Abbruch ist kein Fehler** — dieselbe Übereinkunft wie bei `SetOeffnen`
// und `BildSetzen`: „nichts gewählt" ist eine Aussage.
//
// **Und die Liste ist `[]` und nicht `null`.**
//
// **Diese Zusage steht hier und nicht im Nahtvertrag**, obwohl sie genau die
// Form ist, die er sonst prüft. Am 2026-08-31 stand sie dort — und der Lauf
// hing fünfundzwanzig Minuten, weil `Bindungen.SymbolNachzeichnen` einen
// **nativen Dateidialog** öffnet und auf einen Menschen wartete. Der Kopf des
// Vertrags sagt es ausdrücklich: solche Bindungen „kämen nie zurück", und was
// sie tun, prüft `internal/app` mit gestopfter Naht.
//
// **Was dieser Lauf damit nicht hält**: dass die leere Liste die Brücke als
// `[]` erreicht. Struct-Marken sieht die Go-Seite nicht (§3). Die Antwortform
// selbst ist dieselbe wie bei `Symbol` und `Grundform`, und die stehen im
// Vertrag — mehr ist über den Dialog hinweg nicht zu haben.
func TestEinAbbruchBeimNachzeichnenIstKeinFehler(t *testing.T) {
	t.Parallel()

	antwort := mitWahl(t, "", nil).SymbolNachzeichnen(t.Context())
	if antwort.Fehler != nil {
		t.Errorf("ein Abbruch gab einen Fehler: %+v", antwort.Fehler)
	}

	if antwort.Teile == nil {
		t.Error("Teile ist null statt []")
	}

	if len(antwort.Teile) != 0 {
		t.Errorf("%d Teile nach einem Abbruch", len(antwort.Teile))
	}
}

// **Ein Fehler im Dialog ist einer.**
func TestEinFehlerImDialogBeimNachzeichnenWirdGemeldet(t *testing.T) {
	t.Parallel()

	antwort := mitWahl(t, "", errors.New("der Dialog ging nicht auf")).SymbolNachzeichnen(t.Context())
	if antwort.Fehler == nil {
		t.Error("kein Fehler, obwohl der Dialog scheiterte")
	}

	if antwort.Teile == nil {
		t.Error("Teile ist null statt []")
	}
}
