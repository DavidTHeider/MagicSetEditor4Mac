package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die Stiloptionen kommen aus dem Paket** — die zweite Zusage aus
// #41.
//
// Gegen die Saat des Repositoriums und nicht gegen eine nachgebaute
// Zeichenkette (§3): `probe-schlicht.mse-style` führt eine Auswahl und einen
// Schalter, beide mit Beschreibung und Vorgabewert.
func TestDerStilBringtSeineOptionenMit(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "saat", "probe-schlicht.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Stiloptionen) != 3 {
		t.Fatalf("%d Stiloptionen, erwartet drei: %+v", len(stil.Stiloptionen), stil.Stiloptionen)
	}

	wahl := stil.Stiloptionen[0]
	if wahl.Name != "hintergrund" || wahl.Kind != domain.KindChoice {
		t.Errorf("Auswahl: %+v", wahl)
	}

	if wahl.Initial != "hell" || len(wahl.Choices) != 2 {
		t.Errorf("Vorgabe %q, Möglichkeiten %v", wahl.Initial, wahl.Choices)
	}

	// **Die Beschreibung kommt mit und bleibt deutsch, wie sie dasteht** (§9):
	// was aus einem Paket kommt, wird nicht übersetzt.
	if wahl.Description == "" {
		t.Error("die Beschreibung fehlt")
	}

	schalter := stil.Stiloptionen[1]
	if schalter.Kind != domain.KindBoolean || schalter.Initial != "no" {
		t.Errorf("Schalter: %+v", schalter)
	}
}

// **Eine Option ohne Namen ist ein Befund, kein Abbruch** (§8).
//
// Fremde Daten sind der Normalfall: ein halb geschriebener `styling field:`
// -Block soll das Stilpaket nicht unlesbar machen.
func TestEineStiloptionOhneNamenIstEinBefund(t *testing.T) {
	stil, befunde := stilAusText(t, "mse version: 2.0.0\ngame: probe\n"+
		"styling field:\n\ttype: boolean\n\tdescription: ohne Namen\n"+
		"styling field:\n\ttype: boolean\n\tname: geht\n")

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}

	if len(stil.Stiloptionen) != 1 || stil.Stiloptionen[0].Name != "geht" {
		t.Errorf("Stiloptionen: %+v", stil.Stiloptionen)
	}
}

// stilAusText schreibt ein Stilpaket und liest sein Modell.
func stilAusText(t *testing.T, inhalt string) (msepaket.Stil, []msepaket.Befund) {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return paket.Stilmodell()
}
