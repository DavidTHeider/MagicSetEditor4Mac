// **Ein Kartenbild von der Platte an jeder Fläche** —
// #1291.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probebild schreibt ein winziges PNG und gibt seinen Pfad.
func probebild(t *testing.T) string {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := range 300 {
		for x := range 400 {
			bild.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var roh bytes.Buffer
	if err := png.Encode(&roh, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	pfad := filepath.Join(t.TempDir(), "quelle.png")
	if err := os.WriteFile(pfad, roh.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// kunstset legt ein Set an, dessen Stil ein Bildfeld führt, und eine Karte darin.
//
// **Nicht das Probeset**: `probe-stil` hat keine Fläche für ein Bild, und ein
// Lauf dagegen prüfte den Fehlerweg statt der Fähigkeit. `probe-kunst` aus der
// Saat hat eine — 315 × 230 (§11).
func kunstset(t *testing.T, umgebung []string) (setID, karteID string) {
	t.Helper()

	var angelegt struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung,
		"set-neu", "Kunstset", "probe", "probe-kunst", "--json")), &angelegt); err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	var karte struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung,
		"karte", angelegt.ID, "anlegen", "--json")), &karte); err != nil {
		t.Fatalf("Karte anlegen: %v", err)
	}

	return angelegt.ID, karte.KarteID
}

// **Wer über die Kommandozeile, MCP oder REST ein Bild einsetzt, bekommt
// dieselben Maße.**
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Ein Kartenbild einsetzen | `BildSetzen`, `BildEinsetzen`, `BildAusDatei` | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Dateidialog — und `App.BildAusDatei` nimmt seit
// #246 einen
// Pfad entgegen, weil das Fenster Dateien auch ziehen und ablegen lässt.
func TestEinKartenbildWirdAnJederFlaecheEingesetzt(t *testing.T) {
	quelle := probebild(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)

		aus, fehler, code := kommandozeile(t, umgebung,
			"bild-setzen", setID, karteID, "kunst", quelle, "--json")
		if code != 0 {
			t.Fatalf("setzen: %d (%s)", code, fehler)
		}

		nenntDieMasse(t, "die Kommandozeile", aus)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.karte.bild", map[string]any{
			"setID": zahl(t, setID), "karteID": zahl(t, karteID),
			"feld": "kunst", "pfad": quelle,
		})
		if istFehler {
			t.Fatalf("setzen: %s", text)
		}

		nenntDieMasse(t, "MCP", text)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/karten/"+karteID+"/felder/kunst/bild",
			`{"pfad":`+alsJSONText(t, quelle)+`}`)
		if status != http.StatusOK {
			t.Fatalf("PUT: %d (%s)", status, rumpf)
		}

		nenntDieMasse(t, "REST", rumpf)
	})
}

// **Und dieselbe fachliche Abweisung: eine Datei, die kein Bild ist** (§16).
func TestEineDateiDieKeinBildIstWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	kein := filepath.Join(t.TempDir(), "kein.png")
	if err := os.WriteFile(kein, []byte("hallo"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)

		if _, _, code := kommandozeile(t, umgebung,
			"bild-setzen", setID, karteID, "kunst", kein); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.karte.bild", map[string]any{
			"setID": zahl(t, setID), "karteID": zahl(t, karteID),
			"feld": "kunst", "pfad": kein,
		}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID, karteID := kunstset(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/karten/"+karteID+"/felder/kunst/bild",
			`{"pfad":`+alsJSONText(t, kein)+`}`)
		if status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// nenntDieMasse prüft, dass Quell- und Zielmaße in der Antwort stehen.
//
// **Die Zielmaße kommen aus dem Stil** — 315 × 230 stehen in
// `saat/probe-kunst.mse-style`. Eine Antwort ohne sie hätte nicht
// zugeschnitten, sondern nur eine Datei abgelegt.
func nenntDieMasse(t *testing.T, wer, antwort string) {
	t.Helper()

	for _, erwartet := range []string{"400", "300", "315", "230"} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("%s nennt %s nicht: %s", wer, erwartet, antwort)
		}
	}
}
