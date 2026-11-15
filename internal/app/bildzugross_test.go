package app

// **Ein zu großes Bild heißt an der Naht „zu groß"** — FMT-06b
// (#1974).
// Bis hierher hieß es „kein lesbares Bild", und das schickte jemanden nach
// einem Fehler in einer Datei suchen, die in Ordnung ist.

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// riesendatei legt ein kleines JPEG ab, dessen Kopf 9.000 × 9.000 Bildpunkte
// behauptet — unter der Bildpunktgrenze, über der Grenze des Speichers.
func riesendatei(t *testing.T) string {
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

func TestEinZuGrossesBildHeisstZuGross(t *testing.T) {
	t.Parallel()

	if antwort := (&App{}).SymbolNachzeichnen(t.Context(), riesendatei(t)); antwort.Fehler == nil ||
		antwort.Fehler.Schluessel != domain.SchluesselBildZuGross {
		t.Errorf("Nachzeichnen: %+v", antwort.Fehler)
	}

	if _, fehler := bilddateiLesen(riesendatei(t)); Uebersetze(t.Context(), fehler).Schluessel != domain.SchluesselBildZuGross {
		t.Errorf("Bild setzen: %v", fehler)
	}
}
