package app_test

// **Dienst und Werkzeugserver schreiben geschwärzt und mit Kennung** —
// Befund 7 des Zweitblicks über Phase 6.4
// (#1665):
// sie schrieben über den Standardschreiber von Go, ohne Schwärzung und ohne
// Vorgangskennung — und die Dokumentation sagte das Gegenteil.

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDasServerprotokollSchwaerztUndNenntDenVorgang(t *testing.T) {
	t.Parallel()

	heim, fehler := os.UserHomeDir()
	if fehler != nil {
		t.Skip("kein Zuhause")
	}

	puffer := &bytes.Buffer{}
	protokoll := app.Serverprotokoll(puffer)

	ctx := app.MitVorgang(t.Context(), "0a1b2c3d")
	protokoll.InfoContext(ctx, "ein Ruf", "ort", heim+"/Karten/probe.mse-set")

	zeile := puffer.String()
	if !strings.Contains(zeile, "kennung=0a1b2c3d") || strings.Contains(zeile, heim) ||
		!strings.Contains(zeile, "~/Karten") {
		t.Errorf("%q", zeile)
	}
}
