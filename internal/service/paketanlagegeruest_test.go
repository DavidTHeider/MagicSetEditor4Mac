// **Include und Exportvorlage aus einem Gerüst** — FMT-02.AK-01
// (#1691):
// die Zelle »Erstellen« der Matrix in `docs/FAEHIGKEITEN.md` stand für beide
// auf „fehlend“.
//
// Was ein Gerüst tragen muss, steht in MSE2: ein Include ist nur der
// Paketkopf (`IMPLEMENT_REFLECTION(IncludePackage)`,
// `src/util/io/package.cpp` Zeilen 656–658), eine Exportvorlage **muss**
// `game:` tragen — sonst bricht `ExportTemplate::validate` ab — und hängt von
// ihrem Spiel ab (`src/data/export_template.cpp` Zeilen 31–38, Fassung
// f9d9356 vom 2021-07-06).
package service_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinIncludeEntstehtAusDemGeruest(t *testing.T) {
	ort, verwalter, _ := anlageplatz(t)

	angelegt, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-include", Name: "bausteine", Vollname: "Bausteine",
	})
	if err != nil || angelegt.Art != "mse-include" {
		t.Fatalf("anlegen: %+v, %v", angelegt, err)
	}

	pfad := filepath.Join(ort, "bausteine.mse-include")
	inhalt := datei(t, filepath.Join(pfad, "include"))

	for _, zeile := range []string{"mse version: 2.0.0", "short name: bausteine", "full name: Bausteine"} {
		if !strings.Contains(inhalt, zeile) {
			t.Errorf("das Include führt %q nicht:\n%s", zeile, inhalt)
		}
	}

	if strings.Contains(inhalt, "game:") {
		t.Errorf("ein Include gehört keinem Spiel:\n%s", inhalt)
	}

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur gelesen
}

func TestEineExportvorlageEntstehtAusDemGeruest(t *testing.T) {
	ort, verwalter, _ := anlageplatz(t)

	if _, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes Spiel",
	}); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	angelegt, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-export-template", Name: "liste", Vollname: "Liste", Spiel: "eigenes",
	})
	if err != nil || angelegt.Art != "mse-export-template" {
		t.Fatalf("anlegen: %+v, %v", angelegt, err)
	}

	pfad := filepath.Join(ort, "eigenes-liste.mse-export-template")
	inhalt := datei(t, filepath.Join(pfad, "export-template"))

	for _, zeile := range []string{
		"game: eigenes", "depends on:", "package: eigenes.mse-game", "file type:", "script:",
	} {
		if !strings.Contains(inhalt, zeile) {
			t.Errorf("die Exportvorlage führt %q nicht:\n%s", zeile, inhalt)
		}
	}

	// **Und ihr Skript zerteilt** — ein Gerüst, das die Werkstatt als kaputt
	// meldet, wäre keines.
	if befund := service.PaketSkriptfehler(inhalt); befund != nil {
		t.Errorf("das Gerüst hat einen Befund: %v", befund)
	}

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur gelesen

	if modell := paket.Exportmodell(); modell.Spiel != "eigenes" || modell.Skript == "" {
		t.Errorf("gelesen: %+v", modell)
	}
}

func TestEineExportvorlageBrauchtIhrSpiel(t *testing.T) {
	for name, spiel := range map[string]string{"ohne Spiel": "", "mit fehlendem Spiel": "gibtesnicht"} {
		_, verwalter, _ := anlageplatz(t)

		_, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
			Art: "mse-export-template", Name: "liste", Vollname: "Liste", Spiel: spiel,
		})

		var feldfehler *domain.FieldError
		if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
			t.Errorf("%s: %v, erwartet ein Befund am Feld »spiel«", name, err)
		}
	}
}
