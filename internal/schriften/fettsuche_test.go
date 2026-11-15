package schriften_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// legeSchrift legt eine **echte** Schriftdatei unter dem gegebenen Namen ab.
//
// **Echt und nicht erfunden**, weil die Suche die Datei am Ende öffnet; eine
// Attrappe fiele erst dort auf, und dann prüfte der Test den Fehlerweg.
func legeSchrift(t *testing.T, ort, dateiname string) {
	t.Helper()

	quelle, _ := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if err := schreibe(t, ort, dateiname, roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Fett ist ein Gewicht — gesucht wird trotzdem ein Name.**
//
// MSE reicht Familienname und Gewicht an wxWidgets weiter; unsere Suche kennt
// nur Namen. Aufgelöst wird deshalb über MSEs eigene Regel: das abschließende
// ` Bold` gehört zum Schnitt, nicht zur Familie
// (`src/data/font.cpp`, Zeilen 82–108, Fassung f9d9356 vom 2021-07-06;
// [0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
func TestFettSuchtDenFettenSchnittDerFamilie(t *testing.T) {
	ort := t.TempDir()

	// **Gefunden wird über den Dateinamen.** Welchen Namen eine Schrift über
	// sich selbst führt, ist hier gleichgültig — geprüft wird, *wonach* die
	// Suche fragt, und nicht, wie sie eine Datei aufschlägt.
	legeSchrift(t, ort, "Probe.ttf")
	legeSchrift(t, ort, "Probe Bold.ttf")

	suche := schriften.Neue([]string{ort})

	faelle := []struct {
		name    string
		gesucht string
		fett    bool
		datei   string
	}{
		{"aufrecht", "Probe", false, "Probe.ttf"},
		{"fett", "Probe", true, "Probe Bold.ttf"},

		// **Ein Name, der schon auf ` Bold` endet, ist fett.** MSE schneidet
		// den Zusatz ab und setzt das Gewicht; `<b>` ändert dort nichts mehr.
		{"schon im Namen fett", "Probe Bold", false, "Probe Bold.ttf"},
		{"fett und schon im Namen fett", "Probe Bold", true, "Probe Bold.ttf"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			fundort, err := suche.Fundort(fall.gesucht, fall.fett)
			if err != nil {
				t.Fatalf("suchen: %v", err)
			}

			if filepath.Base(fundort) != fall.datei {
				t.Errorf("gefunden wurde %s, erwartet %s", fundort, fall.datei)
			}
		})
	}
}

// **Ohne fetten Schnitt bleibt es beim aufrechten.**
//
// Dieselbe Bestenfalls-Arbeit wie bei der Kursiven (§8): ein aufrechter Text
// ist eine Karte, ein fehlender ist keine. `MPlantin` liegt in
// `M15-Magic-Pack` ohne fetten Schnitt.
func TestOhneFettenSchnittBleibtEsBeimAufrechten(t *testing.T) {
	ort := t.TempDir()

	legeSchrift(t, ort, "Einzeln.ttf")

	suche := schriften.Neue([]string{ort})

	fundort, err := suche.Fundort("Einzeln", true)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if filepath.Base(fundort) != "Einzeln.ttf" {
		t.Errorf("gefunden wurde %s, erwartet Einzeln.ttf", fundort)
	}
}

// **Und eine Schrift, die es gar nicht gibt, bleibt ein Fehler.**
//
// Auch fett: der Rückfall gilt dem Schnitt, nicht der Familie.
func TestEineFehlendeSchriftBleibtAuchFettEinFehler(t *testing.T) {
	suche := schriften.Neue([]string{t.TempDir()})

	if _, err := suche.Fundort("Gibtsnicht", true); !errors.Is(err, schriften.ErrNichtGefunden) {
		t.Errorf("der Fehler ist %v, erwartet ErrNichtGefunden", err)
	}
}
