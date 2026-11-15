//go:build fremdpakete

// **Wo die Referenzbilder liegen, und über welche Sätze gemessen wird** —
// #1029.
//
// **Eine eigene Datei**, weil `mse2kanon_intern_test.go` sonst über die
// Zeilengrenze liefe (§2). Der Schnitt folgt der Frage: dort steht, **wie**
// gemessen wird, hier **worüber**.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mse2ausgabe ist der Ort, an den `scripts/mse2-referenz.sh` schreibt.
//
// **`MSE_MSE2AUSGABE` geht vor `HOME`** — aus demselben Grund wie bei
// [kanonverzeichnis], und mit demselben Preis: ohne ihn sah dieser Weg das
// eigene Zuhause des Testlaufs und fand nichts.
func mse2ausgabe(satz string) string {
	ort := os.Getenv("MSE_MSE2AUSGABE")
	if ort == "" {
		ort = filepath.Join(os.Getenv("HOME"), "Library", "Application Support",
			"MagicSetEditor4Mac", "mse2-ausgabe")
	}

	return filepath.Join(ort, satz)
}

// kanonsaetze sind die zehn Sätze, die MSE2 selbst mitbringt.
//
// **Eine Aufzählung und kein Verzeichnisscan**: sie sagt, was gemessen werden
// **soll**. Wer sie aus `mse2-ausgabe/` läse, bekäme eine Messung, die still
// schrumpft, sobald jemand ein Verzeichnis löscht — und eine Zahl über acht
// Sätze sähe aus wie eine über zehn.
var kanonsaetze = []string{
	"Archenemy", "Black", "Blue", "Colorless", "Green",
	"Multicolor", "Planechase", "Red", "Vanguard", "White",
}

// saetzeMitReferenz gibt die Sätze, für die MSE2 gerastert hat.
//
// **Und sagt, welche fehlen** — das ist der Punkt (§10). Ein Lauf über vier von
// zehn Sätzen ist eine Aussage über vier; wer sie für eine über den Kanon hält,
// hat den Rest nie gesehen. Genau so blieben `Black`, `Blue`, `Colorless`,
// `Planechase` und `Vanguard` monatelang unverglichen
// (#880).
func saetzeMitReferenz(t *testing.T) []string {
	t.Helper()

	var da, fehlend []string

	for _, satz := range kanonsaetze {
		if _, err := os.Stat(mse2ausgabe(satz)); err == nil {
			// **Und nur mit Manifest** (QA-03, #1656): ein Verzeichnis voller
			// Bilder sagt nicht, zu welchem Stand des Sets sie gehören.
			pruefeReferenzmanifest(t, satz)

			da = append(da, satz)

			continue
		}

		fehlend = append(fehlend, satz)
	}

	if len(fehlend) > 0 {
		t.Logf("ohne Referenzbilder und damit ungemessen: %s "+
			"(erzeugen mit `scripts/mse2-referenz.sh %s`)",
			strings.Join(fehlend, ", "), strings.Join(fehlend, " "))
	}

	return da
}
