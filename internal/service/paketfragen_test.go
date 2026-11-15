package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`exists_in_package` fragt, ob eine Datei in einem Paket liegt.**
//
//	if exists_in_package(path) then path
//	else "/magic-modules.mse-include/loyalty/" + name
//
// `magic.mse-game/script`, Zeile 3614. Die Vorlagen wählen damit zwischen
// einem besonderen Bild und einem allgemeinen — **126 Befunde** über die 190
// Stile von `M15-Magic-Pack` lauteten auf diesen Namen
// (#270).
//
// **Sie steht in `service` und nicht in `msescript`**: welche Pakete es gibt,
// weiß die Bildquelle, und `msescript` kennt dieses Projekt nicht (§2).
//
// **Ohne Bildquelle ist die Antwort nein.** Ein Planbauer ohne Pakete hat keine
// Dateien; eine erfundene Zusage schickte die Vorlage auf ein Bild, das es
// nicht gibt.
func TestExistsInPackageFragtNachEinerDatei(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "name", Kind: domain.KindText,
			Skript: `if exists_in_package("/probe.mse-style/style") then "da" else "weg"`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Ohne Bildquelle: nein.**
	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "weg" {
		t.Errorf("ohne Bildquelle steht %q da, erwartet \"weg\"",
			plan.Ebenen[len(plan.Ebenen)-1].Klartext())
	}
}

// **Mit einer Bildquelle antwortet sie über die Pakete.**
func TestExistsInPackageAntwortetUeberDiePakete(t *testing.T) {
	wurzel := stilwurzel(t)

	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "name", Kind: domain.KindText,
			Skript: `if exists_in_package("/probe.mse-style/style") then "da" else "weg"`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitBildern(msepaket.NeueBildquelle(wurzel)).
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "da" {
		t.Errorf("mit Bildquelle steht %q da, erwartet \"da\"",
			plan.Ebenen[len(plan.Ebenen)-1].Klartext())
	}

	// **Und eine Datei, die es nicht gibt, ist ein Nein.**
	vorlage.Kartenfelder[0].Skript = `if exists_in_package("/probe.mse-style/gibtsnicht") then "da" else "weg"`

	plan, _, err = service.NeuerPlanbauer().
		MitBildern(msepaket.NeueBildquelle(wurzel)).
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if plan.Ebenen[1].Klartext() != "weg" {
		t.Errorf("für eine fehlende Datei steht %q da", plan.Ebenen[1].Klartext())
	}
}

// stilwurzel legt ein Paketverzeichnis mit einer Datei an.
func stilwurzel(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-style")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(paket, "style"), []byte("mse version: 2.0.0"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return wurzel
}
