package domain_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pngkopf sind die acht Bytes, mit denen jede PNG-Datei anfängt — die
// Signatur aus der Formatbeschreibung.
var pngkopf = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

// alsDatenURL verpackt Bytes so, wie eine Leinwand sie gibt.
func alsDatenURL(rohes []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(rohes)
}

func TestDeuteAbzugNimmtEinPNG(t *testing.T) {
	t.Parallel()

	erwartet := append(append([]byte{}, pngkopf...), 'x', 'y')

	gelesen, fehler := domain.DeuteAbzug(alsDatenURL(erwartet))
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if !bytes.Equal(gelesen, erwartet) {
		t.Errorf("gelesen %q, erwartet %q", gelesen, erwartet)
	}
}

func TestDeuteAbzugBefunde(t *testing.T) {
	t.Parallel()

	gross := append(append([]byte{}, pngkopf...),
		make([]byte, domain.Abzugsgrenze)...)

	faelle := []struct {
		name  string
		wert  string
		teile string
	}{
		{
			name:  "leer",
			wert:  "",
			teile: "leer",
		},
		{
			// **Ein anderer Vorsatz ist kein PNG** — auch dann nicht, wenn
			// er nach einem Bild aussieht: dieser Weg schreibt eine Datei,
			// und was darin steht, muss ihrem Namen entsprechen.
			name:  "falscher Vorsatz",
			wert:  "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pngkopf),
			teile: "PNG",
		},
		{
			name:  "kein Data-URL",
			wert:  "/tmp/bild.png",
			teile: "PNG",
		},
		{
			name:  "kein Base64",
			wert:  "data:image/png;base64,###",
			teile: "lesbar",
		},
		{
			// **Die Signatur wird geprüft und nicht der Vorsatz geglaubt.**
			// Wer den Vorsatz setzt, sagt, was er zu schicken meint; die
			// acht Bytes sagen, was wirklich kommt (§8).
			name:  "ohne Signatur",
			wert:  alsDatenURL([]byte("kein Bild")),
			teile: "PNG",
		},
		{
			name:  "zu groß",
			wert:  alsDatenURL(gross),
			teile: "groß",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := domain.DeuteAbzug(fall.wert)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.teile) {
				t.Errorf("Meldung %q nennt %q nicht", fehler, fall.teile)
			}
		})
	}
}
