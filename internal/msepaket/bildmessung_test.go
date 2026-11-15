//go:build fremdpakete

// Was ein Druckbogen die Bildquelle kostet — gemessen, bevor etwas gebaut wird.
//
// #187 sagt es
// als erstes Häkchen: **erst messen, was ein Bogen kostet.** Ein
// Zwischenspeicher, der eine Zahl verbessert, die niemand kennt, ist eine
// Behauptung mit Aufwand.
//
// # Was gemessen wurde
//
// Neunzig Läufe über `/magic-m15-planeswalker.mse-style/mcard.jpg` (169.858
// Byte), auf einem MacBook:
//
//	ohne Zwischenspeicher   1,188 s    13,10 ms je Lauf
//	mit Zwischenspeicher    0,011 ms    121 ns je Lauf
//
// **Und das für ein einziges Bild.** Eine Karte trägt ein Dutzend Ebenen; ein
// Bogen über neunzig Karten sparte damit gut zwei Minuten Entschlüsselung.
//
// Die Zahl steht hier, damit der nächste sieht, ob sie sich bewegt hat — und
// damit niemand den Speicher wieder ausbaut, ohne sie zu kennen.
package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestWasEinBogenDieBildquelleKostet(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	// Ein echter Kartenrahmen aus dem Pack — ein Planeswalker-Grund.
	verweis := "/magic-m15-planeswalker.mse-style/mcard.jpg"

	quelle := msepaket.NeueBildquelle(wurzel)

	if _, err := quelle.Bild(verweis); err != nil {
		t.Skipf("%s liegt hier nicht: %v", verweis, err)
	}

	// **Neunzig Karten** — ein voller Bogen aus neun Karten je Seite über zehn
	// Seiten, und jede Karte liest ihren Rahmen einmal.
	const karten = 90

	begonnen := time.Now()

	for range karten {
		if _, err := quelle.Bild(verweis); err != nil {
			t.Fatalf("lesen: %v", err)
		}
	}

	gebraucht := time.Since(begonnen)

	angaben, err := os.Stat(filepath.Join(wurzel, //nolint:gosec // ein Pfad aus MSE_FREMDPAKETE (§3)
		"magic-m15-planeswalker.mse-style", "mcard.jpg"))
	if err != nil {
		t.Fatalf("Größe: %v", err)
	}

	t.Logf("%d Läufe über %s (%d Byte): %v — %v je Lauf",
		karten, verweis, angaben.Size(), gebraucht, gebraucht/karten)
}
