//go:build fremdpakete

// **`magic-crystal` zeichnet seine Maske** — #1934, AK-03.
//
// Der Stil schreibt `mask: border-mask{pt()}.png`, und `pt` ist
// `{ if card.pt != "" then "-pt" else "" }`; `card.pt` rechnet das Spiel aus
// `power` und `toughness` (`combined_editor`, `save value: false`). Dasselbe
// steht an `foil layer`, und das zeichnet nur mit einem Überzug — geprüft wird
// deshalb `border color`.
//
// MSE2 liest die Angabe im Zeichenkettenmodus (`src/script/image.cpp`
// Zeile 73, Fassung f9d9356 vom 2021-07-06); bis #1934 suchte dieses Programm
// die Datei `border-mask{pt()}.png`, und die Maske fehlte.
//
// Läuft wie `fremdpakete_test.go` nur von Hand, gegen den Pack:
//
//	MSE_FREMDPAKETE=/Users/david/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/service -run MagicCrystal
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestMagicCrystalZeichnetSeineMaske(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	stil, offen := stilAus(t, filepath.Join(wurzel, "magic-crystal.mse-style"), wurzel)
	if !offen {
		t.Fatal("magic-crystal.mse-style liegt nicht in MSE_FREMDPAKETE")
	}

	vorlage := vorlageFuer(t, stil, wurzel)

	for name, fall := range map[string]struct {
		werte  []domain.Value
		masken map[string]string
	}{
		"mit Stärke": {
			werte: []domain.Value{
				{Field: "name", Text: "Probe"}, {Field: "power", Text: "2"}, {Field: "toughness", Text: "2"},
			},
			masken: map[string]string{"border color": "/magic-crystal.mse-style/border-mask-pt.png"},
		},
		"ohne Stärke": {
			werte:  []domain.Value{{Field: "name", Text: "Probe"}},
			masken: map[string]string{"border color": "/magic-crystal.mse-style/border-mask.png"},
		},
	} {
		karte := &domain.CardRow{ID: 1, Number: 1, Values: fall.werte}

		plan, befunde, err := service.NeuerPlanbauer().
			MitBildern(msepaket.NeueBildquelle(wurzel)).
			Stilplan(t.Context(), karte, vorlage)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		for _, befund := range befunde {
			if strings.Contains(befund.Meldung, "mask") {
				t.Errorf("%s: Befund über eine Maske: %s", name, befund)
			}
		}

		gefunden := map[string]string{}

		for _, ebene := range plan.Ebenen {
			if _, gefragt := fall.masken[ebene.Feld]; gefragt && ebene.Maske.Pfad != "" {
				gefunden[ebene.Feld] = ebene.Maske.Pfad
			}
		}

		for feld, pfad := range fall.masken {
			if gefunden[feld] != pfad {
				t.Errorf("%s: Maske von %q ist %q, erwartet %q", name, feld, gefunden[feld], pfad)
			}

			//nolint:gosec // Pfad aus dem Plan eines Pakets unter MSE_FREMDPAKETE, nur gelesen
			if _, fehlt := os.Stat(filepath.Join(wurzel, strings.TrimPrefix(pfad, "/"))); fehlt != nil {
				t.Errorf("%s: %s liegt nicht im Paket: %v", name, pfad, fehlt)
			}
		}
	}
}
