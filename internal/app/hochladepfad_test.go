package app_test

// **Der Ordner, in dem ein Dateidialog aufgeht** — Werkstück 9 von
// Phase 4.12
// (#1351),
// Punkt 8.
//
// # Warum das hier geprüft wird und nicht am Dialog
//
// **Weil der Dialog ein Fenster braucht** (`dateiwahl.go` steht in
// `.covignore`). Was sich prüfen lässt, ist die **Übergabe**: bekommt die
// Naht das Verzeichnis, das eingestellt ist? Was der native Dialog damit tut,
// ist Apples Sache.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der eingestellte Ordner erreicht die Dateiwahl.**
func TestDerHochladepfadErreichtDieDateiwahl(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	// **Ein Ordner, den es gibt** — seit OPS-02 (#1719) fragt das Setzen
	// die Platte.
	ordner := t.TempDir()

	if antwort := anwendung.EinstellungSetzen(
		t.Context(), domain.VorgabeHochladepfad, ordner,
	); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	var gesehen app.Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gesehen = auswahl

		return "", nil
	})

	// **Über einen echten Weg und nicht über die Naht selbst**: dass der
	// Mantel sitzt, sagt erst ein Ruf, der durch ihn geht.
	anwendung.NutzerbildWaehlen(t.Context(), "1", "1")

	if gesehen.Startverzeichnis != ordner {
		t.Errorf("der Dialog beginnt bei %q", gesehen.Startverzeichnis)
	}
}

// **Ein Sicherndialog bekommt ihn nicht.**
//
// **Das ist die halbe Zusage, und ohne sie wäre die andere zu weit**: die
// Kommentierung nennt Ordner, aus denen etwas **hochgeladen** wird. Wohin
// etwas hinausgeht, entscheidet, wer es hinausschreibt.
func TestEinSicherndialogBekommtDenHochladepfadNicht(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	// **Ein Ordner, den es gibt** — seit OPS-02 (#1719) fragt das Setzen
	// die Platte.
	ordner := t.TempDir()

	if antwort := anwendung.EinstellungSetzen(
		t.Context(), domain.VorgabeHochladepfad, ordner,
	); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	var gesehen app.Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gesehen = auswahl

		return "", nil
	})

	// **`PersonendatenSichern` öffnet einen Sicherndialog** — und ist der
	// kürzeste Weg zu einem.
	anwendung.PersonendatenSichern(t.Context(), "1")

	if !gesehen.Speichern {
		t.Fatalf("das war kein Sicherndialog: %+v", gesehen)
	}

	if gesehen.Startverzeichnis != "" {
		t.Errorf("der Sicherndialog beginnt bei %q", gesehen.Startverzeichnis)
	}
}

// **Ohne Einstellung bleibt es, wie es war.**
func TestOhneHochladepfadBleibtDerDialogWoErWar(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	var gesehen app.Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gesehen = auswahl

		return "", nil
	})

	anwendung.NutzerbildWaehlen(t.Context(), "1", "1")

	if gesehen.Startverzeichnis != "" {
		t.Errorf("der Dialog beginnt bei %q", gesehen.Startverzeichnis)
	}
}

// **Ein Lesefehler hält den Dialog nicht an.**
//
// **Er geht dann dort auf, wo das Betriebssystem ihn öffnet** — genau wie
// ohne Einstellung. Der Ausfall einer Bequemlichkeit ist kein Grund, die
// Arbeit anzuhalten (§8); eine Meldung „der Ordner ließ sich nicht lesen"
// stünde vor einem Dialog, den jemand gerade öffnen wollte.
func TestEinLesefehlerHaeltDenDialogNichtAn(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitEinstellungen(&einstellungsprobe{
		fehler: errors.New("die Platte schweigt"),
	})

	var gesehen app.Dateiauswahl
	gerufen := false

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gesehen = auswahl
		gerufen = true

		return "", nil
	})

	anwendung.NutzerbildWaehlen(t.Context(), "1", "1")

	if !gerufen {
		t.Fatal("der Dialog kam gar nicht")
	}

	if gesehen.Startverzeichnis != "" {
		t.Errorf("der Dialog beginnt bei %q", gesehen.Startverzeichnis)
	}
}
