package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Nur ein Spiel und ein Stil lassen sich anlegen** — und dieser Lauf ist der
// ausführbare Test der übrigen fünf Zellen »Erstellen« in der Matrix aus
// FMT-01.AK-02 (#1690).
// Ein Set legt `SetAnlegen` an; eine Symbolschrift und ein Sprachpaket könnten
// leer nichts (nicht anwendbar); ein Include und eine Exportvorlage fehlen noch
// (FMT-02, #1691).
// **Wer eine davon anlegbar macht, sieht ihn rot** — und trägt die Matrix nach.
// **Include und Exportvorlage gehen seit FMT-02 (#1691)** — sie stehen in
// `paketanlagegeruest_test.go`; hier bleibt, was sich nicht anlegen lässt.
func TestNurSpielUndStilLassenSichAnlegen(t *testing.T) {
	for _, art := range []string{
		"mse-set", "mse-symbol-font", "mse-locale",
	} {
		t.Run(art, func(t *testing.T) {
			ort, verwalter, katalog := anlageplatz(t)

			_, err := verwalter.Anlegen(t.Context(),
				service.Neuanlage{Art: art, Name: "eigenes", Vollname: "X"})

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != "art" {
				t.Fatalf("Fehler ist %v", err)
			}

			// **Und abgewiesen heißt: nichts geschrieben** — weder im Katalog
			// noch im verwalteten Verzeichnis.
			if len(katalog.pakete) != 0 {
				t.Errorf("im Katalog: %+v", katalog.pakete)
			}

			if eintraege, _ := os.ReadDir(ort); len(eintraege) != 0 { //nolint:errcheck // ein fehlendes Verzeichnis ist leer
				t.Errorf("im Verzeichnis: %v", eintraege)
			}
		})
	}
}

// **Speichern gibt es für eine Vorlage nicht, weil jede Änderung sofort
// schreibt** — der ausführbare Test der Zelle »Speichern: nicht anwendbar«
// (FMT-01.AK-02). Gelesen wird **an der Werkstatt vorbei von der Platte**: ein
// Zwischenstand, der erst ein Sichern bräuchte, fiele hier auf.
func TestEineAenderungLiegtSofortAufDerPlatte(t *testing.T) {
	t.Parallel()

	wurzel := werkwurzel(t)
	werk := service.NeuePaketwerkstatt(wurzel)

	_, stand, fehler := werk.Lesen("probe.mse-style", "style")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	neu := "mse version: 2.0.0\ngame: probe\nfull name: Sofort\n"
	if _, schreibfehler := werk.Schreiben(t.Context(), "probe.mse-style", "style", neu, stand); schreibfehler != nil {
		t.Fatalf("Schreiben: %v", schreibfehler)
	}

	roh, fehler := os.ReadFile(filepath.Join(wurzel, "probe.mse-style", "style")) //nolint:gosec // im eigenen Wegwerfverzeichnis
	if fehler != nil || string(roh) != neu {
		t.Errorf("auf der Platte %q (%v), geschrieben %q", roh, fehler, neu)
	}
}
