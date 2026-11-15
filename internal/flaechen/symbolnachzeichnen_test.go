// **Ein Bild als Symbol nachzeichnen — an jeder Fläche** — Werkstück F von
// Phase 4.11
// (#1340).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Warum diese Fähigkeit überhaupt kopflos steht
//
// Die übrigen Symbolfähigkeiten führt `docs/FAEHIGKEITEN.md` mit Grund allein
// am Fenster: eine Grundform, ein Punktzug, eine Symmetrie sind Handgriffe auf
// einer Zeichenfläche. **Nachzeichnen ist keiner** — Bild hinein, Zeichnung
// heraus (§16).
package flaechen_test

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// kreisbild schreibt ein PNG mit einem gefüllten Kreis und gibt seinen Pfad.
func kreisbild(t *testing.T) string {
	t.Helper()

	const mass = 80

	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))
	mitte := float64(mass) / 2

	for y := range mass {
		for x := range mass {
			ton := uint8(255)
			if math.Hypot(float64(x)-mitte, float64(y)-mitte) <= float64(mass)*0.35 {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	pfad := filepath.Join(t.TempDir(), "kreis.png")

	datei, err := os.Create(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	if err := png.Encode(datei, gemalt); err != nil {
		t.Fatalf("das PNG schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("die Datei schließen: %v", err)
	}

	return pfad
}

// pruefeNachzeichnung hält die Antwort gegen dieselben Zusagen.
func pruefeNachzeichnung(t *testing.T, roh string) {
	t.Helper()

	var antwort struct {
		Name   string `json:"name"`
		Teile  int    `json:"teile"`
		Punkte int    `json:"punkte"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
	}

	if antwort.Name == "" {
		t.Error("die abgelegte Datei hat keinen Namen")
	}

	if antwort.Teile != 1 {
		t.Errorf("%d Formen, erwartet 1 — ein Kreis ist eine", antwort.Teile)
	}

	// **Weiterbearbeitbar heißt: wenige Punkte.** Der Rundgang gibt für dieses
	// Bild rund zweihundert; was ankommt, sind ein paar Dutzend.
	if antwort.Punkte < 4 || antwort.Punkte > 60 {
		t.Errorf("%d Punkte — erwartet zwischen 4 und 60", antwort.Punkte)
	}
}

// **Jede Fläche zeichnet dasselbe Bild zu derselben Zeichnung nach.**
func TestEinBildWirdUeberallGleichNachgezeichnet(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		pruefeNachzeichnung(t, mussKommandozeile(t, umgebung,
			"symbol-nachzeichnen", setID, kreisbild(t), "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.symbolnachzeichnen",
			map[string]any{"setID": zahl(t, setID), "bildpfad": kreisbild(t)})
		if istFehler {
			t.Fatalf("schreiben.symbolnachzeichnen: %s", roh)
		}

		pruefeNachzeichnung(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		an := starteDienst(t, umgebung)

		rumpfHinein, err := json.Marshal(map[string]string{"bildpfad": kreisbild(t)})
		if err != nil {
			t.Fatalf("den Rumpf bauen: %v", err)
		}

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/symbol-nachzeichnen", string(rumpfHinein))
		if status != http.StatusOK {
			t.Fatalf("POST /sets/%s/symbol-nachzeichnen: %d (%s)", setID, status, rumpf)
		}

		pruefeNachzeichnung(t, rumpf)
	})
}

// **Und ein Bild, das es nicht gibt, ist überall dasselbe Nein.**
//
// Der Fall steht hier, weil ein Fehlerweg leicht an einer Fläche anders
// aussieht als an den anderen — ein leeres Ergebnis statt einer Meldung, oder
// ein Rückgabewert 0 für einen Fehlschlag.
func TestEinFehlendesBildIstUeberallEinNein(t *testing.T) {
	t.Parallel()

	fehlt := func(t *testing.T) string {
		t.Helper()

		return filepath.Join(t.TempDir(), "gibtesnicht.png")
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		_, _, code := kommandozeile(t, umgebung,
			"symbol-nachzeichnen", setID, fehlt(t))
		if code == 0 {
			t.Error("Rückgabewert 0 für ein Bild, das es nicht gibt")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.symbolnachzeichnen",
			map[string]any{"setID": zahl(t, setID), "bildpfad": fehlt(t)}); !istFehler {
			t.Error("kein Fehler für ein Bild, das es nicht gibt")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		an := starteDienst(t, umgebung)

		rumpfHinein, err := json.Marshal(map[string]string{"bildpfad": fehlt(t)})
		if err != nil {
			t.Fatalf("den Rumpf bauen: %v", err)
		}

		status, _ := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/symbol-nachzeichnen", string(rumpfHinein))
		if status == http.StatusOK {
			t.Error("Status 200 für ein Bild, das es nicht gibt")
		}
	})
}
