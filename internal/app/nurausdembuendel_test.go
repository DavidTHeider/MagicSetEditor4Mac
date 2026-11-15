//go:build fremdpakete

package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Eine Karte zeichnen, ohne dass jemand ein Paket bezogen hat** —
// #702.
//
// # Was er prüft, das kein anderer prüft
//
// Jeder andere Lauf dieses Projekts sieht die **bezogenen** Pakete dieses
// Rechners. Die Frage von #702 ist eine andere: reicht, was der Freigabeweg ins
// Bündel legt? Deshalb zeigt hier das Paketverzeichnis ins Leere, und allein
// die mitgelieferten Vorlagen tragen den Lauf.
//
// # Warum hinter `fremdpakete`
//
// Weil er ein gebautes Bündel braucht — `./scripts/mitliefern.sh` muss gelaufen
// sein. In der Pipeline liegt keines (§11).
//
//	MSE_BUENDEL=build/bin/MagicSetEditor4Mac.app go test -tags fremdpakete \
//	  ./internal/app/ -run TestNurAusDemBuendel -v
func TestNurAusDemBuendel(t *testing.T) {
	buendel := os.Getenv("MSE_BUENDEL")
	if buendel == "" {
		t.Skip("MSE_BUENDEL ist nicht gesetzt")
	}

	vorlagen := filepath.Join(buendel, "Contents", "Resources", "vorlagen")
	sets := filepath.Join(buendel, "Contents", "Resources", "beispielsets")

	if _, err := os.Stat(vorlagen); err != nil { //nolint:gosec // MSE_BUENDEL setzt, wer den Lauf startet
		t.Skipf("das Bündel trägt keine Vorlagen: %v", err)
	}

	// **Das Paketverzeichnis zeigt ins Leere.** Genau das ist die Lage auf
	// einem Rechner, der die Anwendung gerade geladen hat.
	leer := t.TempDir()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.paketort = func() (string, error) { return leer, nil }
	anwendung.MitVorlagenort(func() string { return vorlagen })
	anwendung.MitBeispielsetort(func() string { return sets })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	bindungen := NeueBindungen(anwendung)

	liste := bindungen.Beispielsets(t.Context())
	if liste.Fehler != nil || len(liste.Sets) == 0 {
		t.Fatalf("keine Beispielsets: %+v", liste)
	}

	geoeffnet := bindungen.BeispielsetOeffnen(t.Context(), liste.Sets[0].Pfad)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten)
	}

	// **Und die erste Karte wird gezeichnet.** Ohne Rahmenbilder käme eine
	// leere Fläche samt Befund — sichtbar, aber nutzlos.
	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	if len(vorschau.Vorschau.Befunde) > 0 {
		t.Errorf("%d Befunde: %v",
			len(vorschau.Vorschau.Befunde), vorschau.Vorschau.Befunde)
	}

	if vorschau.Vorschau.Breite == 0 || vorschau.Vorschau.Bild == "" {
		t.Errorf("kein Bild: %dx%d", vorschau.Vorschau.Breite, vorschau.Vorschau.Hoehe)
	}

	t.Logf("%d Karten aus dem Bündel, erste gezeichnet: %d×%d",
		len(karten.Karten), vorschau.Vorschau.Breite, vorschau.Vorschau.Hoehe)
}
