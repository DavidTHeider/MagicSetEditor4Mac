package flaechen_test

// **Ein zu großes Bild ist überall dasselbe Nein** — FMT-06b
// (#1974),
// am Nachzeichnen eines Symbols, das an allen vier Flächen steht.

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zuGrosserSatz ist, was jede Fläche über ein solches Bild sagt.
const zuGrosserSatz = "zu groß, um es auszupacken"

// riesenbild legt ein kleines JPEG ab, dessen Kopf 9.000 × 9.000 Bildpunkte
// behauptet — unter der Bildpunktgrenze, über der Grenze des Speichers.
func riesenbild(t *testing.T) string {
	t.Helper()

	var puffer bytes.Buffer
	if fehler := jpeg.Encode(&puffer, image.NewYCbCr(image.Rect(0, 0, 8, 8), image.YCbCrSubsampleRatio444), nil); fehler != nil {
		t.Fatal(fehler)
	}

	roh := puffer.Bytes()
	stelle := bytes.Index(roh, []byte{0xFF, 0xC0})
	roh[stelle+5], roh[stelle+6], roh[stelle+7], roh[stelle+8] = 0x23, 0x28, 0x23, 0x28

	pfad := filepath.Join(t.TempDir(), "riesig.jpg")
	if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

func TestEinZuGrossesBildIstUeberallDasselbeNein(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		_, klage, code := kommandozeile(t, umgebung, "symbol-nachzeichnen", setID, riesenbild(t))
		if code == 0 || !strings.Contains(klage, zuGrosserSatz) {
			t.Errorf("%d %s", code, klage)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")

		text, istFehler := mcpRuf(t, umgebung, "schreiben.symbolnachzeichnen",
			map[string]any{"setID": zahl(t, setID), "bildpfad": riesenbild(t)})
		if !istFehler || !strings.Contains(text, zuGrosserSatz) {
			t.Errorf("%v %s", istFehler, text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Symbolset")
		an := starteDienst(t, umgebung)

		rumpfHinein, fehler := json.Marshal(map[string]string{"bildpfad": riesenbild(t)})
		if fehler != nil {
			t.Fatal(fehler)
		}

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/symbol-nachzeichnen", string(rumpfHinein))
		if status != http.StatusBadRequest || !strings.Contains(rumpf, zuGrosserSatz) {
			t.Errorf("%d %s", status, rumpf)
		}
	})
}
