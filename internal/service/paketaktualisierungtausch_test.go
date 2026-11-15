package service_test

// **Aktualisieren und zurück** — G04
// (#2251):
// der bisherige Baum bleibt aufgehoben, zurück tauscht beide, und ohne
// Bestätigung geschieht nichts, wo die Vorschau eine verlangt.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAktualisierenHebtDenBisherigenAufUndZurueckTauscht(t *testing.T) {
	dienst := paketserver(t)
	aktualisierung, auskunft, verwaltet := aktualisierungsplatzVon(t, dienst.URL)
	auskunft.vergleich = bezug.Vergleich{Lage: "ahead", Vorsprung: 2}

	ort := filepath.Join(verwaltet, "M15-Magic-Pack")
	if fehler := os.WriteFile(filepath.Join(ort, "alt.txt"), []byte("alt"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ergebnis, fehler := aktualisierung.Aktualisieren(t.Context(), "M15-Magic-Pack", m15Spitze, false)
	if fehler != nil {
		t.Fatalf("Aktualisieren: %v", fehler)
	}

	vorige := filepath.Join(verwaltet+"-vorige-bezuege", "M15-Magic-Pack")
	for pfad, soll := range map[string]bool{
		filepath.Join(ort, "alt.txt"):    false,
		filepath.Join(vorige, "alt.txt"): true,
	} {
		if _, statfehler := os.Stat(pfad); (statfehler == nil) != soll {
			t.Errorf("%s da=%v, erwartet %v", pfad, statfehler == nil, soll)
		}
	}

	if ergebnis.Stand != m15Spitze || ergebnis.Ort != ort {
		t.Errorf("Ergebnis %+v", ergebnis)
	}

	staende, fehler := aktualisierung.Staende(t.Context(), false)
	if fehler != nil || !staende[0].Vorige || staende[0].VorigerStand != m15Stand {
		t.Fatalf("Stände %+v, Fehler %v — der bisherige Stand liegt nicht bereit", staende, fehler)
	}

	rueckweg, fehler := aktualisierung.Zurueck(t.Context(), "M15-Magic-Pack")
	if fehler != nil {
		t.Fatalf("Zurueck: %v", fehler)
	}

	if rueckweg.Stand != m15Stand || rueckweg.Adresse != dienst.URL || rueckweg.Ort != ort {
		t.Errorf("Rückweg %+v", rueckweg)
	}

	// **Zurück tauscht**: der neuere liegt jetzt aufgehoben.
	if _, fehler := os.Stat(filepath.Join(ort, "alt.txt")); fehler != nil {
		t.Errorf("der bisherige Stand ist nicht zurück: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(vorige, "alt.txt")); fehler == nil {
		t.Error("der bisherige Stand liegt noch aufgehoben")
	}
}

func TestOhneBestaetigungGeschiehtNichts(t *testing.T) {
	dienst := paketserver(t)
	aktualisierung, auskunft, verwaltet := aktualisierungsplatzVon(t, dienst.URL)
	auskunft.vergleich = bezug.Vergleich{Lage: "behind", Rueckstand: 4}

	_, fehler := aktualisierung.Aktualisieren(t.Context(), "M15-Magic-Pack", m15Spitze, false)

	var feldfehler *domain.FieldError
	if !errors.As(fehler, &feldfehler) || feldfehler.Field != "bestaetigt" {
		t.Fatalf("Fehler %v, erwartet am Feld bestaetigt", fehler)
	}

	if _, fehler := os.Stat(verwaltet + "-vorige-bezuege"); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("ohne Bestätigung wurde etwas aufgehoben: %v", fehler)
	}

	// **Mit Bestätigung geht es** — dieselbe Lage.
	if _, fehler := aktualisierung.Aktualisieren(t.Context(), "M15-Magic-Pack", m15Spitze, true); fehler != nil {
		t.Errorf("mit Bestätigung: %v", fehler)
	}
}

func TestZurueckOhneAufgehobenenStand(t *testing.T) {
	aktualisierung, _, _ := aktualisierungsplatz(t)

	if _, fehler := aktualisierung.Zurueck(t.Context(), "M15-Magic-Pack"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}

	var feldfehler *domain.FieldError
	if _, fehler := aktualisierung.Zurueck(t.Context(), "../weg"); !errors.As(fehler, &feldfehler) {
		t.Errorf("ein ausbrechender Name: %v", fehler)
	}
}
