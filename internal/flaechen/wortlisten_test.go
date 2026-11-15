// Die Auswahllisten über jede Fläche
// (#1074).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Warum die Saat ins Paketverzeichnis kopiert wird
//
// Weil Wortlisten **im Spielpaket** stehen. Die Binärdateien dieses Laufs
// liegen in einem Temporärverzeichnis, und `Saatverzeichnis()` sucht neben dem
// Programm — im Bündel läge sie dort, hier nicht. Kopiert wird deshalb dorthin,
// wo ein bezogenes Paket landete.
//
// **Und dafür wird `HOME` umgebogen** (§3): das Paketverzeichnis hängt an
// `os.UserConfigDir()` und nicht an `MSE_BESTAND`. Ohne die zweite Angabe läse
// dieser Lauf die Pakete **dieses Rechners** — und wäre grün, weil es sie
// wirklich gibt.
package flaechen_test

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Dieselbe Auswahl an jeder Fläche, und dieselbe Wirkung.**
func TestDieAuswahlEinesFeldesUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	// **Die Marken kommen im Betrieb aus einem Feldfilter** des Spielpakets;
	// hier werden sie geschrieben, denn geprüft wird die Auswahl.
	roh := "<word-list-probe-art>Große Kreatur</word-list-probe-art>"
	mussKommandozeile(t, umgebung, "feld", setID, karteID, "text", roh)

	t.Run("Kommandozeile", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "wortlisten", karteID, "text")

		for _, erwartet := range []string{"probe-art", "[x] Große", "( ) Spruch"} {
			if !strings.Contains(aus, erwartet) {
				t.Errorf("%q fehlt in\n%s", erwartet, aus)
			}
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.wortlisten",
			map[string]any{"karteID": zahl(t, karteID), "feld": "text"})
		if istFehler {
			t.Fatalf("lesen.wortlisten: %s", text)
		}

		if !strings.Contains(text, "probe-art") || !strings.Contains(text, "Kreatur") {
			t.Errorf("Antwort %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/karten/"+karteID+"/felder/text/wortlisten", "")
		if status != http.StatusOK {
			t.Fatalf("GET wortlisten: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, "probe-art") || !strings.Contains(rumpf, "Kreatur") {
			t.Errorf("Antwort %s", rumpf)
		}
	})

	// **Und die Wahl wirkt an jeder Fläche gleich.** Sie läuft der Reihe nach:
	// jede Fläche wählt etwas anderes, und die nächste sieht, was die vorige
	// hinterlassen hat.
	t.Run("wählen über die Kommandozeile", func(t *testing.T) {
		mussKommandozeile(t, umgebung, "wortwahl", setID, karteID, "text", "0", "2")

		pruefeFeld(t, umgebung, setID, karteID, "Große Spruch")
	})

	t.Run("wählen über MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.wortwahl", map[string]any{
			"setID": zahl(t, setID), "karteID": zahl(t, karteID),
			"feld": "text", "bereich": 0, "weg": []any{1},
		})
		if istFehler {
			t.Fatalf("schreiben.wortwahl: %s", text)
		}

		// **Ein Punkt ersetzt den Punkt und lässt das Präfix stehen** —
		// `Große ` war angekreuzt und bleibt es.
		pruefeFeld(t, umgebung, setID, karteID, "Große Kreatur")
	})

	t.Run("wählen über REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/karten/"+karteID+"/felder/text/wortwahl",
			`{"bereich":0,"weg":[0]}`)
		if status != http.StatusOK {
			t.Fatalf("POST wortwahl: %d (%s)", status, rumpf)
		}

		// **Und ein Präfix wird umgeschaltet**: dieselbe Wahl noch einmal
		// nimmt es wieder weg. Das ist MSEs Verhalten und der Grund, warum
		// `schreiben.wortwahl` keine Idempotenz zusagt.
		pruefeFeld(t, umgebung, setID, karteID, "Kreatur")
	})
}

// pruefeFeld liest das Textfeld der Karte und vergleicht seinen Inhalt.
//
// **Über die JSON-Ausgabe, und deshalb mit geschützten spitzen Klammern.**
// Gos Kodierer schreibt `<` als `\u003c`; wer die rohe Form suchte, fände sie
// nie und hielte den Wert für ungeschrieben.
func pruefeFeld(t *testing.T, umgebung []string, setID, karteID, erwartet string) {
	t.Helper()

	aus := mussKommandozeile(t, umgebung, "karten", setID, "--json")

	voll := `\u003cword-list-probe-art\u003e` + erwartet +
		`\u003c/word-list-probe-art\u003e`
	if !strings.Contains(aus, voll) {
		t.Errorf("%q steht nicht in der Kartenliste:\n%s", voll, aus)
	}

	// **Und das Feld gehört zu dieser Karte** — die Liste führt drei.
	if !strings.Contains(aus, `"id": "`+karteID+`"`) {
		t.Errorf("die Karte %s steht nicht in der Liste", karteID)
	}
}

// mitEigenerSaat gibt eine Umgebung mit eigenem Bestand, eigenem Zuhause und
// der Saat als Paketverzeichnis.
func mitEigenerSaat(t *testing.T) []string {
	t.Helper()

	zuhause := t.TempDir()

	quelle, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	// **Der Bestand am Anwendungsort seines Zuhauses** — seit QA-10.NAT-01
	// (#1741) folgen die Bilder dem Bestand und die Pakete dem Zuhause; so
	// liegen beide an einem Ort, wie ohne Umgebung auch.
	anwendungsort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac")
	ziel := filepath.Join(anwendungsort, "pakete")

	if kopierfehler := os.CopyFS(ziel, os.DirFS(quelle)); kopierfehler != nil {
		t.Fatalf("Saat kopieren: %v", kopierfehler)
	}

	// **Der Stil des Probesets heißt anders als der der Saat.** `probe.mse-set`
	// nennt `probe-stil`; die Saat liefert `probe-schlicht`. Ein zweiter Name
	// auf dasselbe Paket ist billiger als ein zweites Probeset.
	if linkfehler := os.Symlink(
		filepath.Join(ziel, "probe-schlicht.mse-style"),
		filepath.Join(ziel, "probe-stil.mse-style"),
	); linkfehler != nil && !errors.Is(linkfehler, fs.ErrExist) {
		t.Fatalf("Stil verknüpfen: %v", linkfehler)
	}

	// **Der eigene Schlüsselbund steht in `eigenerBestand`** — eine Stelle,
	// an der jeder Kindprozess vorbeikommt, statt einer je Vorrichtung.
	return append(eigenerBestand(t), "HOME="+zuhause,
		"MSE_BESTAND="+filepath.Join(anwendungsort, "bestand.db"))
}
