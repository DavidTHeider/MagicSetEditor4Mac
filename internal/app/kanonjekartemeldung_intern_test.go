//go:build fremdpakete

package app

// **Was der Kanonlauf je Karte festhält und vergleicht** — die Seite von
// `kanonkartenwerte_intern_test.go`, die fremde Pakete braucht (F02,
// #2244).

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"testing"
)

// kartentoleranz ist, um wie viel eine Karte wandern darf, ohne als
// schlechter zu gelten.
//
// **Klein, weil der Lauf sich wiederholt**: dieselbe Rasterung gab am
// 2026-09-20 jede Zahl bis auf die vierte Nachkommastelle wieder
// (MESSMETHODIK.md §5a). Was darüber wandert, hat jemand verändert.
const kartentoleranz = 0.01

// meldeJeKarte schreibt die Werte je Karte nach `MSE_KANON_JEKARTE` und hält
// sie gegen `MSE_KANON_VORHER` — beides nur, wenn genannt.
func meldeJeKarte(t *testing.T, karten []kartenwert) {
	t.Helper()

	if ziel := os.Getenv("MSE_KANON_JEKARTE"); ziel != "" {
		datei, err := os.Create(ziel) //nolint:gosec // ein Ablageort, den der Aufrufer nennt
		if err != nil {
			t.Fatalf("MSE_KANON_JEKARTE: %v", err)
		}

		if err := schreibeKartenwerte(datei, karten); err != nil {
			t.Fatalf("MSE_KANON_JEKARTE: %v", err)
		}

		if err := datei.Close(); err != nil {
			t.Fatalf("MSE_KANON_JEKARTE: %v", err)
		}

		t.Logf("je Karte: %d Werte nach %s", len(karten), ziel)
	}

	quelle := os.Getenv("MSE_KANON_VORHER")
	if quelle == "" {
		return
	}

	datei, err := os.Open(quelle) //nolint:gosec // ein früherer Lauf, den der Aufrufer nennt
	if err != nil {
		t.Fatalf("MSE_KANON_VORHER: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	inhalt, err := io.ReadAll(datei)
	if err != nil {
		t.Fatalf("MSE_KANON_VORHER: %v", err)
	}

	// **Welcher frühere Lauf, steht im Beleg** (Befund 2): ein Name allein
	// sagt nicht, ob die Datei dieselbe ist.
	t.Logf("früherer Lauf: %s, sha256 %x", quelle, sha256.Sum256(inhalt))

	vorher, err := liesKartenwerte(bytes.NewReader(inhalt))
	if err != nil {
		t.Fatalf("MSE_KANON_VORHER: %v", err)
	}

	vergleich := vergleicheKartenwerte(vorher, karten, kartentoleranz)

	t.Logf("gegen %s: %d verglichen — %d besser, %d gleich (±%.2f), %d schlechter; "+
		"%d ohne früheren Wert, %d nicht mehr gemessen",
		quelle, vergleich.verglichen, vergleich.besser, vergleich.gleich, kartentoleranz,
		len(vergleich.schlechter), vergleich.ohneVorher, vergleich.ohneNachher)

	for i, wert := range vergleich.schlechter {
		if i == 20 {
			t.Logf("  … und %d weitere", len(vergleich.schlechter)-i)

			break
		}

		t.Logf("  schlechter: %-12s %4d %-40q %7.3f → %7.3f (%+.3f)",
			wert.Set, wert.Stelle, wert.Name, wert.Vorher, wert.Abstand, wert.Abstand-wert.Vorher)
	}

	for _, meldung := range vergleich.abweichungen(kartentoleranz, quelle) {
		t.Error(meldung)
	}
}
