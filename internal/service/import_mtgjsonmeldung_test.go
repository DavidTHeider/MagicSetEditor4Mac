package service_test

// **Was eine MTG-JSON-Datei sagt, die keine ist** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01fan`): eine JSON-Datei ohne Set stand
// als »notizen.json ist kein lesbares MTG JSON — an Byte 22 steht string,
// erwartet war mtgjson.rohSet« da. Der Typ des Dekodierers ist kein Satz für
// einen Menschen, und übersetzen ließ er sich auch nicht. Jetzt trägt der
// Befund einen eigenen Schlüssel samt Datei und Stelle.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineMTGJSONDateiOhneSetSagtEsMitSchluessel(t *testing.T) {
	for _, fall := range []struct {
		name, datei, inhalt string
		schluessel          domain.Schluessel
	}{
		{"kein JSON", "kaputt.json", `{"data": {"name": "Probe", `, domain.SchluesselMTGJSONKeinJSON},
		{"JSON ohne Set", "notizen.json", `{"das": "ist kein Set"}`, domain.SchluesselMTGJSONKeinSet},
		{"ein Wert der falschen Art", "falsch.json", `{"data": {"name": 5, "code": "PRB", "cards": []}}`, domain.SchluesselMTGJSONKeinSet},
	} {
		t.Run(fall.name, func(t *testing.T) {
			pfad := filepath.Join(t.TempDir(), fall.datei)
			lege(t, pfad, fall.inhalt)

			_, err := service.NeuerImporteur(&speicherDouble{}).MTGJSON(t.Context(), pfad, "magic")

			var befund *domain.FieldError
			if !errors.As(err, &befund) || befund.Key != fall.schluessel {
				t.Fatalf("Befund %#v, erwartet Schlüssel %q", err, fall.schluessel)
			}

			if befund.Werte["datei"] != fall.datei || befund.Werte["stelle"] == "" {
				t.Errorf("Werte %v", befund.Werte)
			}

			// **Kein Typ des Dekodierers im Satz** — weder `mtgjson.rohSet`
			// noch ein Go-Typ wie `string`.
			if satz := befund.Message; !strings.HasPrefix(satz, fall.datei+" ist kein lesbares MTG JSON") ||
				strings.Contains(satz, "mtgjson.") || strings.Contains(satz, "erwartet war") {
				t.Errorf("Satz %q", satz)
			}
		})
	}
}
