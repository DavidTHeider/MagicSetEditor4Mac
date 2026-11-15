package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Einstellungen liegen im Bestand** —
// #245,
// letzter Abnahmekasten. Bis hierher lag die Sprache im Browserspeicher.
func TestEineEinstellungUeberlebtDenNeustart(t *testing.T) {
	ablage := storage.NeueEinstellungen(frischeDatenbank(t))

	if err := ablage.Setzen(probekontext(t), "erscheinungsbild", "dunkel"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gelesen, err := ablage.Alle(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen["erscheinungsbild"] != "dunkel" {
		t.Errorf("%v", gelesen)
	}
}

// **Zweimal gesetzt heißt einmal gespeichert.** Ohne `ON CONFLICT` stünde die
// Einstellung zweimal da — und welche gilt, entschiede die Reihenfolge beim
// Lesen.
func TestEineEinstellungWirdUeberschriebenUndNichtVerdoppelt(t *testing.T) {
	ablage := storage.NeueEinstellungen(frischeDatenbank(t))

	for _, wert := range []string{"hell", "dunkel"} {
		if err := ablage.Setzen(probekontext(t), "erscheinungsbild", wert); err != nil {
			t.Fatalf("setzen: %v", err)
		}
	}

	gelesen, err := ablage.Alle(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 || gelesen["erscheinungsbild"] != "dunkel" {
		t.Errorf("%v", gelesen)
	}
}

// **Ohne Wahl ist die Abbildung leer, nicht nil.** Ein `nil` durchquert die
// Naht als `null` und die Oberfläche prüfte darauf; eine leere Abbildung sagt
// dasselbe und braucht keinen zweiten Zweig.
func TestOhneWahlIstDieAbbildungLeer(t *testing.T) {
	gelesen, err := storage.NeueEinstellungen(frischeDatenbank(t)).Alle(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen == nil || len(gelesen) != 0 {
		t.Errorf("%v", gelesen)
	}
}
