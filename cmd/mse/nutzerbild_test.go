package main

// **`--keines` nimmt ein Bild weg — und setzt vorher keines.**
//
// Gefunden beim Messen doppelter Logik für ENG-03
// (#1728):
// `nutzer-bild` und `profil-badge` riefen **erst** das Setzen mit dem Pfad
// `--keines` und **dann** das Entfernen. Lag im Arbeitsverzeichnis eine
// Bilddatei dieses Namens, wurde sie gesetzt, die Fassung zählte hoch, und das
// Entfernen scheiterte an genau dieser Fassung — der Befehl meldete einen
// Konflikt, den er selbst erzeugt hatte, und das falsche Bild blieb stehen.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// imVerzeichnisMitKeines wechselt in ein Verzeichnis, in dem eine echte
// Bilddatei namens `--keines` liegt.
func imVerzeichnisMitKeines(t *testing.T) {
	t.Helper()

	verzeichnis := t.TempDir()

	roh, err := os.ReadFile(probebild(t, 4, 4)) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// #nosec G703 -- ein Pfad aus t.TempDir() und eine Konstante
	if err := os.WriteFile(filepath.Join(verzeichnis, Keinbild), roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Chdir(verzeichnis)
}

func TestKeinesNimmtDasNutzerbildWegOhneEsVorherZuSetzen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "nutzer-bild", "1", probebild(t, 4, 4), "1"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	imVerzeichnisMitKeines(t)

	code, aus, fehler := fahreAn(t, ort, "nutzer-bild", "1", Keinbild, "2")
	if code != CodeGut {
		t.Fatalf("wegnehmen: %d (%s)", code, fehler)
	}

	if feld := strings.Fields(aus); len(feld) != 1 || feld[0] != "1" {
		t.Errorf("nach dem Wegnehmen trägt der Nutzer ein Bild: %q", aus)
	}
}

func TestKeinesNimmtDasAbzeichenWegOhneEsVorherZuSetzen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "profil-badge", "1", probebild(t, 4, 4), "1"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	imVerzeichnisMitKeines(t)

	code, aus, fehler := fahreAn(t, ort, "profil-badge", "1", Keinbild, "2")
	if code != CodeGut {
		t.Fatalf("wegnehmen: %d (%s)", code, fehler)
	}

	if feld := strings.Fields(aus); len(feld) != 1 || feld[0] != "1" {
		t.Errorf("nach dem Wegnehmen trägt das Profil ein Abzeichen: %q", aus)
	}
}
