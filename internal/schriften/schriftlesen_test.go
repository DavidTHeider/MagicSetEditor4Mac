package schriften

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Eine Schriftdatei hat eine Grenze** — FMT-06.AK-01
// (#1695).
//
// Jede der fünf Lesestellen dieses Pakets las die ganze Datei, gleich wie groß
// sie war. Eine Datei, die jemand als `.ttf` in ein Paket legt und die keine
// ist, bestimmte so den Arbeitsspeicher. Die Datei hier ist dünn belegt: sie
// nennt ihre Größe und kostet auf der Platte nichts.
func TestEineUebergrosseSchriftWirdNichtGelesen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "riesig.ttf")

	datei, err := os.Create(pfad) //nolint:gosec // im Temporärverzeichnis des Tests
	if err != nil {
		t.Fatal(err)
	}

	if err := datei.Truncate(GrenzeSchriftdatei + 1); err != nil {
		t.Fatal(err)
	}

	if err := datei.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := schriftLesen(pfad); !errors.Is(err, ErrSchriftZuGross) {
		t.Errorf("Fehler %v, erwartet ErrSchriftZuGross", err)
	}

	// **Und was an der Grenze liegt, wird gelesen** — die Grenze gehört dazu.
	klein := filepath.Join(t.TempDir(), "klein.ttf")
	if err := os.WriteFile(klein, []byte("roh"), 0o600); err != nil {
		t.Fatal(err)
	}

	if roh, err := schriftLesen(klein); err != nil || string(roh) != "roh" {
		t.Errorf("gelesen %q, %v", roh, err)
	}

	if _, err := schriftLesen(filepath.Join(t.TempDir(), "fehlt.ttf")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("eine fehlende Datei: %v", err)
	}
}

// **Dieselbe Grenze für das teilweise Lesen** (E03,
// #2240):
// wer nur die Namen holt, öffnet die Datei und liest sie nicht ganz —
// gesperrt ist sie trotzdem.
func TestDieGrenzeGiltAuchFuerDasOeffnen(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	riesig, klein, fehlt := filepath.Join(ort, "riesig.ttf"), filepath.Join(ort, "klein.ttf"), filepath.Join(ort, "fehlt.ttf")

	if err := os.WriteFile(klein, []byte("roh"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(riesig, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Truncate(riesig, GrenzeSchriftdatei+1); err != nil {
		t.Fatal(err)
	}

	if datei, err := schriftOeffnen(riesig); !errors.Is(err, ErrSchriftZuGross) || datei != nil {
		t.Errorf("geöffnet %v, Fehler %v, erwartet ErrSchriftZuGross", datei, err)
	}

	if _, err := schriftOeffnen(fehlt); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("eine fehlende Datei: %v", err)
	}

	offen, err := schriftOeffnen(klein)
	if err != nil {
		t.Fatalf("an der Grenze: %v", err)
	}

	if fehler := offen.Close(); fehler != nil {
		t.Error(fehler)
	}

	// **Und was keine Schrift ist, hat keine Namen** — statt eines Fehlers,
	// denn der Index geht weiter zur nächsten Datei.
	for _, ohne := range []string{riesig, klein, fehlt} {
		if namen, _ := namenAusDatei(ohne); namen != nil {
			t.Errorf("%s: Namen %v", filepath.Base(ohne), namen)
		}
	}
}

// **Die Grenze gilt, weil es eine Stelle gibt** — kein Quelltext dieses
// Pakets außer `schriftlesen.go` liest oder öffnet eine Datei selbst. Sonst hielte die
// Grenze an fünf Stellen und fehlte an der sechsten.
func TestEineSchriftWirdNurAnEinerStelleGelesen(t *testing.T) {
	t.Parallel()

	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("keine Quelltexte gefunden: %v", err)
	}

	gesehen := false

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		roh, err := os.ReadFile(datei) //nolint:gosec // Quelltext dieses Pakets
		if err != nil {
			t.Fatal(err)
		}

		// **`os.Open` gehört dazu** (E03): seit die Namen teilweise gelesen
		// werden, ist das Öffnen ein zweiter Weg an eine Schrift.
		if !strings.Contains(string(roh), "os.ReadFile(") && !strings.Contains(string(roh), "os.Open(") {
			continue
		}

		if datei == "schriftlesen.go" {
			gesehen = true

			continue
		}

		t.Errorf("%s liest eine Datei selbst — ohne die Grenze von `schriftLesen`", datei)
	}

	if !gesehen {
		t.Fatal("die Suche findet nicht einmal schriftlesen.go — dann sagt sie nichts")
	}
}
