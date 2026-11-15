package service

// **Grenzfälle, die erst die Zufallsmutation zeigte** — B05
// (#2224).
// Paketintern, weil die drei Stellen kleine Helfer sind, die kein öffentlicher
// Weg einzeln betritt.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die Namen reisen mit** — und ein fehlender wird zur leeren Liste, nicht
// zu `null` in der Antwort.
func TestEineEinstellungsgruppeTraegtIhreNamen(t *testing.T) {
	t.Parallel()

	if gruppe := einstellungsgruppe(2, []string{"sprache", "erscheinungsbild"}); !reflect.DeepEqual(gruppe[0].Namen, []string{"sprache", "erscheinungsbild"}) {
		t.Errorf("die Gruppe trägt %v", gruppe[0].Namen)
	}

	if gruppe := einstellungsgruppe(0, nil); gruppe[0].Namen == nil {
		t.Error("ohne Namen steht nil da, nicht die leere Liste")
	}
}

// **Auch eine Neun ist eine Ziffer** — `image9` hat den Stamm `image`.
func TestDerNamensstammLaesstAuchDieNeunWeg(t *testing.T) {
	t.Parallel()

	for name, stamm := range map[string]string{"image9": "image", "image19": "image", "bild0": "bild"} {
		if got := ohneZiffern(name); got != stamm {
			t.Errorf("ohneZiffern(%q) = %q, erwartet %q", name, got, stamm)
		}
	}
}

// **Ohne Ablage wird nichts aus der Ablage gelesen** — auch nicht relativ zum
// Arbeitsverzeichnis, wo zufällig eine gleichnamige Datei liegt.
func TestOhneAblageWirdNichtsAusDerAblageGelesen(t *testing.T) {
	ort := t.TempDir()
	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Raute", Kombination: msepaket.KombinationVereinen,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.05}},
			{Stelle: msepaket.Vektor{X: 0.95, Y: 0.5}},
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.95}},
		},
	}}}

	if err := os.WriteFile(filepath.Join(ort, "symbol1.mse-symbol"), msepaket.SchreibeSymbolzeichnung(zeichnung), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Chdir(ort)

	if _, gefunden := (ablagensymbolquelle{}).ausDerAblage("symbol1.mse-symbol"); gefunden {
		t.Error("ohne Ablage las die Quelle eine Datei des Arbeitsverzeichnisses")
	}
}
