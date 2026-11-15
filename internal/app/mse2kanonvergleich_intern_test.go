//go:build fremdpakete

// **Der Vergleich Karte für Karte** —
// #1170.
//
// **Eine eigene Datei**, weil `mse2kanon_intern_test.go` mit der Textzusage
// über die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht,
// **was** der Lauf über den Kanon aussagt, hier, **wie** eine einzelne Karte
// gegen ihr Referenzbild gemessen wird.

package app

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vergleicheMitMSE2 zeichnet jede Karte und misst sie gegen MSE2s Bild.
func vergleicheMitMSE2(
	t *testing.T, anwendung *App, satz string, karten []KarteDTO,
	stand *messstand,
) (abstaende []float64, ohneReferenz int) {
	t.Helper()

	abstaende = make([]float64, 0, len(karten))

	var (
		schlechtesteKarte    string
		schlechtesterAbstand float64
	)

	// **Die Gründe getrennt** (#616): „kein Name" ist unsere Sache, „kein
	// Referenzbild" die der Referenz. In einen Topf geworfen sagt die Zahl
	// nicht mehr, wessen Ausfall sie zählt.
	verluste := ausfaelle{}

	for i := range karten {
		name := kartenname(karten[i])
		if name == "" {
			verluste.nimm("die Karte hat keinen Namen")
			ohneReferenz++

			continue
		}

		referenz, err := ladeBild(filepath.Join(referenzort(t, satz),
			referenzdatei(name)))
		if err != nil {
			verluste.nimm("kein Referenzbild", name)
			ohneReferenz++

			continue
		}

		vorschau := anwendung.Vorschau(t.Context(), karten[i].ID, 1)
		if vorschau.Fehler != nil {
			verluste.nimm("die Vorschau scheiterte", name)
			t.Errorf("%s: %+v", name, vorschau.Fehler)

			continue
		}

		unsere := bildAus(t, anwendung, vorschau.Vorschau.Bild)

		abstand, ok := zerlegeUndMisst(t, anwendung, karten[i], name, unsere,
			referenz, stand)
		if !ok {
			verluste.nimm("der Abstand ließ sich nicht rechnen", name)

			continue
		}

		abstaende = append(abstaende, abstand)
		stand.karten = append(stand.karten, kartenwert{Set: satz, Stelle: i, Name: name, Abstand: abstand})

		if abstand > schlechtesterAbstand {
			schlechtesterAbstand, schlechtesteKarte = abstand, name
		}
	}

	verluste.melde(t, len(abstaende))

	if schlechtesteKarte != "" {
		t.Logf("  %-12s am weitesten weg: %q mit %.3f",
			satz, schlechtesteKarte, schlechtesterAbstand)
	}

	return abstaende, ohneReferenz
}

// kartenname liest den Namen einer Karte aus ihren Werten.
func kartenname(karte KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "name" {
			return strings.TrimSpace(wert.Wert)
		}
	}

	return ""
}

// ladeBild liest ein PNG von der Platte.
func ladeBild(pfad string) (image.Image, error) {
	datei, err := os.Open(pfad) //nolint:gosec // ein Pfad aus der MSE2-Ausgabe
	if err != nil {
		return nil, err
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	return png.Decode(datei)
}
