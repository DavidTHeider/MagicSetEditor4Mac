package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Bezugsquellen erreichen das Fenster** (#263).
//
// Eine Liste, die nur in Go steht, macht aus dem Abtippen einer Adresse keinen
// Klick. Die Naht ist kurz und wird trotzdem geprüft: sie ist die Stelle, an
// der eine Umbenennung eines Feldes still ankommt — im Fenster stünde dann
// eine Zeile ohne Adresse, und der Knopf führte ins Leere.
func TestDieBezugsquellenErreichenDasFenster(t *testing.T) {
	anwendung := app.NeueApp("")

	antwort := anwendung.Bezugsquellen()
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Quellen) == 0 {
		t.Fatal("keine Quelle")
	}

	for _, quelle := range antwort.Quellen {
		if quelle.Name == "" || quelle.Adresse == "" || quelle.Stand == "" ||
			quelle.Schluessel == "" || quelle.Megabyte == 0 {
			t.Errorf("%+v", quelle)
		}
	}
}

// **Und sie brauchen keinen Bestand.** `Bezugsquellen` ist die einzige
// Bindung, die vor jeder Naht antwortet: sie ist genau das, was jemand vor
// seinem ersten Paket sieht. Wer sie an `a.pakete` hinge, bekäme auf einem
// frischen Rechner eine leere Seite.
func TestDieBezugsquellenAntwortenOhneNaht(t *testing.T) {
	if antwort := app.NeueApp("").Bezugsquellen(); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}
}
