//go:build fremdpakete

// **Eine Schicht der Zufallsstichprobe messen** — abgetrennt von
// `zufallsstichprobe_intern_test.go`, damit dort steht, was zugesagt wird, und
// hier, wie gemessen wird (QA-02,
// #1734).
package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// kartenDesSatzes öffnet einen kanonischen Satz und gibt seine Karten — in
// der Ordnung, auf die sich die Stellen des Plans beziehen.
func kartenDesSatzes(t *testing.T, anwendung *App, satz string) []KarteDTO {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(), filepath.Join(kanonverzeichnis(), satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("%s öffnen: %+v", satz, geoeffnet.Fehler)
	}

	liste := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if liste.Fehler != nil {
		t.Fatalf("%s: Karten: %+v", satz, liste.Fehler)
	}

	return liste.Karten
}

// namenDer gibt die Namen einer Kartenliste, in ihrer Ordnung.
func namenDer(karten []KarteDTO) []string {
	namen := make([]string, 0, len(karten))
	for _, karte := range karten {
		namen = append(namen, kartenname(karte))
	}

	return namen
}

// kleingeschrieben gibt den Namen einer Befundart für eine Kennzahl.
func kleingeschrieben(art render.Befundart) string {
	return strings.ToLower(string(art))
}

// messeSchicht misst die gezogenen Karten eines Satzes — Befunde und Abstand
// in **einem** Durchgang, denn jede Karte zu zeichnen kostet eine
// Viertelsekunde.
//
// **Der Plan nennt Stellen, und die gelten nur für diese Liste.** Passt sie
// nicht mehr — eine Karte mehr, eine andere Ordnung —, bricht der Lauf ab:
// ein neuer Plan ist eine neue Kohorte, und das entscheidet niemand still.
func messeSchicht(t *testing.T, anwendung *App, schicht planschicht, gruende kanongruende,
	stand *messstand,
) []float64 {
	t.Helper()

	karten := kartenDesSatzes(t, anwendung, schicht.Satz)
	if len(karten) != schicht.Karten || kartenhash(namenDer(karten)) != schicht.Kartenhash {
		t.Fatalf("%s: der Plan passt nicht mehr zu diesem Satz — %d Karten statt %d oder "+
			"eine andere Liste; ein neuer Plan ist eine neue Kohorte", schicht.Satz,
			len(karten), schicht.Karten)
	}

	abstaende := make([]float64, 0, len(schicht.Stichprobe))
	portierung, ohneReferenz := 0, 0

	for _, stelle := range schicht.Stichprobe {
		karte := karten[stelle]
		name := kartenname(karte)

		vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
		if vorschau.Fehler != nil {
			t.Errorf("%s, %q: %+v", schicht.Satz, name, vorschau.Fehler)

			continue
		}

		for _, meldung := range vorschau.Vorschau.Befunde {
			klasse := kanonklasse(meldung)
			gruende.zaehle(klasse, schicht.Satz, name)

			if render.ArtAus(klasse) == render.ArtPortierung {
				portierung++
			}
		}

		referenz, fehler := ladeBild(filepath.Join(referenzort(t, schicht.Satz), referenzdatei(name)))
		if fehler != nil {
			ohneReferenz++

			continue
		}

		if abstand, gemessen := zerlegeUndMisst(t, anwendung, karte, name,
			bildAus(t, anwendung, vorschau.Vorschau.Bild), referenz, stand); gemessen {
			abstaende = append(abstaende, abstand)
		}
	}

	meldeSchicht(t, schicht, len(abstaende), portierung, ohneReferenz)

	return abstaende
}

// meldeSchicht sagt, was eine Schicht belegt — **und was nicht**.
//
// **Eine Zufallsstichprobe, der Karten fehlen, trägt ihre Zusage nicht**: die
// fehlenden sind nicht zufällig weggefallen, sondern weil ihnen etwas fehlt.
func meldeSchicht(t *testing.T, schicht planschicht, gemessen, portierung, ohneReferenz int) {
	t.Helper()

	if ohneReferenz > 0 {
		t.Errorf("%s: %d gezogene Karten haben kein Referenzbild — die Stichprobe ist "+
			"nicht mehr die gezogene", schicht.Satz, ohneReferenz)
	}

	if portierung > 0 {
		t.Logf("  %-12s %4d von %5d gezogen, %4d gemessen — %d Befunde der Art "+
			"Portierung: die Schranke gilt nicht", schicht.Satz, len(schicht.Stichprobe),
			schicht.Karten, gemessen, portierung)

		return
	}

	t.Logf("  %-12s %4d von %5d gezogen, %4d gemessen — mit %.2f %% Sicherheit liefern "+
		"weniger als 1 %% der Karten einen Befund der Art Portierung", schicht.Satz,
		len(schicht.Stichprobe), schicht.Karten, gemessen, 100*schicht.Findewahrscheinlichkeit)
}
