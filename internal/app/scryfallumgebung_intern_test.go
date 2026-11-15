package app

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die Umgebungsvariable lenkt die Grundadresse um** —
// #1028.
//
// # Warum es sie gibt
//
// Die Flächenläufe starten `mse`, `mse-mcp` und `mse-rest` als **eigene
// Prozesse**. Eine Veränderliche im selben Programm erreichen sie nicht;
// [MitScryfalladresse] hilft dort nicht.
//
// **Was dieser Lauf nicht hält** (§10): dass die drei Programme sie wirklich
// lesen. Das hält `internal/flaechen` — dort wird sie gesetzt, und wenn sie
// nicht wirkte, ginge der Lauf ans echte Netz und die Gegenstelle sähe nichts.
func TestDieUmgebungLenktDieScryfallAdresseUm(t *testing.T) {
	t.Setenv(Scryfallumgebung, "http://127.0.0.1:1/probe")

	if ort := umgelenkteAdresse(); ort != "http://127.0.0.1:1/probe" {
		t.Errorf("Adresse %q", ort)
	}
}

// **Und ohne sie steht die echte da.**
//
// Ohne diesen Fall bestünde den Lauf darüber auch eine Fassung, die **immer**
// umlenkt — und die ginge im Betrieb nirgendwohin.
func TestOhneUmgebungStehtDieEchteAdresseDa(t *testing.T) {
	t.Setenv(Scryfallumgebung, "")

	if ort := umgelenkteAdresse(); ort != scryfall.Adresse {
		t.Errorf("Adresse %q, erwartet %q", ort, scryfall.Adresse)
	}
}
